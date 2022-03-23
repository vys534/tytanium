package routes

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/constants"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/logger"
	"github.com/vysiondev/tytanium/response"
	"github.com/vysiondev/tytanium/security"
	"github.com/vysiondev/tytanium/utils"
	"io"
	"os"
	"path"
	"sync"
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
			response.SendTextResponse(ctx, "No multipart form was in the request.", fasthttp.StatusBadRequest)
			return
		}
		response.SendTextResponse(ctx, fmt.Sprintf("The multipart form couldn't be parsed. %v", e), fasthttp.StatusBadRequest)
		return
	}
	defer ctx.Request.RemoveMultipartFormFiles()
	if mp.File == nil || len(mp.File[fileHandler]) == 0 {
		response.SendTextResponse(ctx, "No files were sent.", fasthttp.StatusBadRequest)
		return
	}
	f := mp.File[fileHandler][0]

	if global.Configuration.RateLimit.Bandwidth.Upload > 0 && global.Configuration.RateLimit.Bandwidth.ResetAfter > 0 {
		isUploadBandwidthLimitNotReached, err := security.Try(ctx, global.RedisClient, fmt.Sprintf("%s_%s", constants.RateLimitBandwidthUpload, utils.GetIP(ctx)), int64(global.Configuration.RateLimit.Bandwidth.Upload), int64(global.Configuration.RateLimit.Bandwidth.ResetAfter), f.Size)
		if err != nil {
			response.SendTextResponse(ctx, fmt.Sprintf("Bandwidth limit couldn't be checked. %v", err), fasthttp.StatusInternalServerError)
			return
		}
		if !isUploadBandwidthLimitNotReached {
			response.SendTextResponse(ctx, "Upload bandwidth limit reached; try again later.", fasthttp.StatusTooManyRequests)
			return
		}
	}

	ext := path.Ext(f.Filename)

	if len(ext) > constants.ExtensionLengthLimit {
		response.SendTextResponse(ctx, "The file extension is too long.", fasthttp.StatusBadRequest)
		return
	}

	openedFile, e := f.Open()
	if e != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("File failed to open. %v", e), fasthttp.StatusInternalServerError)
		return
	}
	defer func() {
		_ = openedFile.Close()
	}()

	mimeType, e := mimetype.DetectReader(openedFile)
	if e != nil {
		response.SendTextResponse(ctx, "Cannot detect the mime type of this file.", fasthttp.StatusBadRequest)
		return
	}

	status := security.FilterCheck(ctx, mimeType.String())
	if status == security.FilterFail {
		// response already sent if filter check failed, so no need to send anything here
		return
	}

	_, e = openedFile.Seek(0, io.SeekStart)
	if e != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("Reader could not be reset to its initial position. %v", e), fasthttp.StatusInternalServerError)
		return
	}

	var fileName string
	attempts := 0

	// loop until an unoccupied id is found
	for {
		var wg sync.WaitGroup
		randomStringChan := make(chan string, 1)
		go func() {
			wg.Add(1)
			utils.RandBytes(global.Configuration.Storage.IDLength, randomStringChan, func() { wg.Done() })
		}()
		wg.Wait()
		fileId := <-randomStringChan

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
			response.SendTextResponse(ctx, "Tried too many times to find a valid file ID to use. Consider increasing the ID length.", fasthttp.StatusInternalServerError)
			return
		}
	}

	fsFile, err := os.Create(path.Join(global.Configuration.Storage.Directory, fileName))
	defer func() {
		_ = fsFile.Close()
	}()

	if err != nil {
		if err == os.ErrPermission {
			response.SendTextResponse(ctx, fmt.Sprintf("Permission to create a file was denied. %v", err), fasthttp.StatusInternalServerError)
			return
		}
		response.SendTextResponse(ctx, fmt.Sprintf("Could not create the file. %v", err), fasthttp.StatusInternalServerError)
		return
	}

	_, writeErr := io.Copy(fsFile, openedFile)
	if writeErr != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("The file failed to write to disk. %v", e), fasthttp.StatusInternalServerError)
		return
	}

	if global.Configuration.Logging.Enabled {
		logger.InfoLogger.Printf("File %s was created, size: %d", fileName, f.Size)
	}

	if global.Configuration.ForceZeroWidth || string(ctx.QueryArgs().Peek("zerowidth")) == "1" {
		fileName = utils.ZeroWidthCharactersToString(fileName)
	}

	var u string
	if string(ctx.QueryArgs().Peek("omitdomain")) == "1" {
		u = fileName
	} else {
		u = fmt.Sprintf("%s/%s", utils.GetServerRoot(ctx), fileName)
	}

	response.SendTextResponse(ctx, u, fasthttp.StatusOK)
}
