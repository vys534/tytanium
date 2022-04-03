package routes

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/sio"
	"github.com/valyala/fasthttp"
	"io"
	"os"
	"path"
	"tytanium/constants"
	"tytanium/encryption"
	"tytanium/global"
	"tytanium/logger"
	"tytanium/response"
	"tytanium/security"
	"tytanium/utils"
)

const fileHandler = "file"

// ServeUpload handles all incoming POST requests to /upload. It will take a multipart form, parse the file,
// then write it to disk.
func ServeUpload(ctx *fasthttp.RequestCtx) {
	auth := security.IsAuthorized(ctx)
	if !auth {
		return
	}
	mp, e := ctx.Request.MultipartForm()
	if e != nil {
		if e == fasthttp.ErrNoMultipartForm {
			response.SendJSONResponse(ctx, response.JSONResponse{
				Status:  response.RequestStatusError,
				Data:    nil,
				Message: "No multipart form was present in the request.",
			}, fasthttp.StatusOK)
			return
		}
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: fmt.Sprintf("The multipart form couldn't be parsed. %v", e),
		}, fasthttp.StatusOK)
		return
	}
	defer ctx.Request.RemoveMultipartFormFiles()
	if mp.File == nil || len(mp.File[fileHandler]) == 0 {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "No files were sent.",
		}, fasthttp.StatusOK)
		return
	}
	f := mp.File[fileHandler][0]

	if global.Configuration.RateLimit.Bandwidth.Upload > 0 && global.Configuration.RateLimit.Bandwidth.ResetAfter > 0 {
		isUploadBandwidthLimitNotReached, err := security.Try(ctx, global.RedisClient, fmt.Sprintf("%s_%s", constants.RateLimitBandwidthUpload, utils.GetIP(ctx)), int64(global.Configuration.RateLimit.Bandwidth.Upload), int64(global.Configuration.RateLimit.Bandwidth.ResetAfter), f.Size)
		if err != nil {
			response.SendJSONResponse(ctx, response.JSONResponse{
				Status:  response.RequestStatusInternalError,
				Data:    nil,
				Message: fmt.Sprintf("Bandwidth limit could not be checked. %v", err),
			}, fasthttp.StatusOK)
			return
		}
		if !isUploadBandwidthLimitNotReached {
			response.SendJSONResponse(ctx, response.JSONResponse{
				Status:  response.RequestStatusError,
				Data:    nil,
				Message: "Upload bandwidth limit reached; try again later.",
			}, fasthttp.StatusTooManyRequests)
			return
		}
	}

	ext := path.Ext(f.Filename)

	if len(ext) > constants.ExtensionLengthLimit {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "File extension is too long.",
		}, fasthttp.StatusOK)
		return
	}

	openedFile, e := f.Open()
	if e != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("File could not be opened. %v", e),
		}, fasthttp.StatusOK)
		return
	}
	defer func() {
		_ = openedFile.Close()
	}()

	mimeType, e := mimetype.DetectReader(openedFile)
	if e != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: "Cannot detect the mime type of this file.",
		}, fasthttp.StatusOK)
		return
	}

	status := security.FilterCheck(ctx, mimeType.String())
	if status == security.FilterFail {
		// response already sent if filter check failed, so no need to send anything here
		return
	}

	_, e = openedFile.Seek(0, io.SeekStart)
	if e != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Reader could not be reset to its initial position. %v", e),
		}, fasthttp.StatusOK)
		return
	}

	var fileName string
	attempts := 0

	// loop until an unoccupied id is found
	for {

		fileId := utils.RandString(global.Configuration.Storage.IDLength)
		fileName = fileId + ext

		i, e := os.Stat(path.Join(global.Configuration.Storage.Directory, fileName))
		if e != nil {
			if os.IsNotExist(e) || e == os.ErrNotExist {
				break
			}
		}
		if i == nil {
			break
		}
		attempts++
		if attempts >= global.Configuration.Storage.CollisionCheckAttempts {
			response.SendJSONResponse(ctx, response.JSONResponse{
				Status:  response.RequestStatusError,
				Data:    nil,
				Message: "Tried too many times to find a valid file ID to use. Consider increasing the ID length.",
			}, fasthttp.StatusOK)
			return
		}
	}

	destFile, err := os.Create(path.Join(global.Configuration.Storage.Directory, fileName))
	defer func() {
		_ = destFile.Close()
	}()

	if err != nil {
		if err == os.ErrPermission {
			response.SendJSONResponse(ctx, response.JSONResponse{
				Status:  response.RequestStatusInternalError,
				Data:    nil,
				Message: fmt.Sprintf("Permission to create the file was denied. %v", err),
			}, fasthttp.StatusOK)
			return
		}
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Failed to create the file. %v", err),
		}, fasthttp.StatusOK)
		return
	}

	masterKey := utils.RandString(global.Configuration.Encryption.EncryptionKeyLength)

	key, err := encryption.DeriveKey([]byte(masterKey), []byte(global.Configuration.Encryption.Nonce))
	if err != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Failed to generate encryption key. %v", err),
		}, fasthttp.StatusOK)
		return
	}

	if _, err = sio.Encrypt(destFile, openedFile, sio.Config{Key: key[:]}); err != nil {
		if _, ok := err.(sio.Error); ok {
			response.SendInvalidEncryptionKeyResponse(ctx)
			return
		}
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Failed to write encrypted file to disk. %v", err),
		}, fasthttp.StatusOK)
		return
	}

	if global.Configuration.Logging.Enabled {
		logger.InfoLogger.Printf("File %s was created, size: %d", fileName, f.Size)
	}

	targetPath := fmt.Sprintf("%s?enc_key=%s", fileName, masterKey)

	if global.Configuration.ForceZeroWidth || string(ctx.QueryArgs().Peek("zerowidth")) == "1" {
		targetPath = utils.StringToZeroWidth(targetPath)
	}

	response.SendJSONResponse(ctx, response.JSONResponse{
		Status: response.RequestStatusOK,
		Data: struct {
			URI           string `json:"uri"`
			Path          string `json:"path"`
			FileName      string `json:"file_name"`
			EncryptionKey string `json:"encryption_key"`
		}{
			URI:           global.Configuration.Domain + "/" + targetPath,
			Path:          targetPath,
			FileName:      fileName,
			EncryptionKey: masterKey,
		},
		Message: "",
	}, fasthttp.StatusOK)
}
