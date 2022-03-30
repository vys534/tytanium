package routes

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/sio"
	"github.com/valyala/fasthttp"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"tytanium/constants"
	"tytanium/encryption"
	"tytanium/global"
	"tytanium/response"
	"tytanium/security"
	"tytanium/utils"
)

const encryptionKeyParam = "enc_key"
const rawParam = "raw"

// discordHTML represents what is sent back to any client which User-Agent contains the regex contained in
// discordBotRegex.
// Derived from https://github.com/whats-this/cdn-origin/blob/8b05fa8425db01cce519ca8945203f9d3050c33b/main.go#L439.
const discordHTML = `<html>
	<head>
		<meta property="twitter:card" content="summary_large_image" />
		<meta property="twitter:image" content="{{.}}" />
		<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
		<meta http-equiv="Pragma" content="no-cache" />
		<meta http-equiv="Expires" content="0" />
	</head>
</html>`

var (
	// discordBotRegex checks if the User-Agent contains a string comparable to "discordbot".
	discordBotRegex = regexp.MustCompile("(?i)discordbot")
)

// ServeFile will serve the / endpoint. It gets the "id" variable from mux and tries to find the file's information in the database.
// If an ID is either not provided or not found, the function hands the request off to ServeNotFound.
func ServeFile(ctx *fasthttp.RequestCtx) {
	pBytes := ctx.Request.URI().Path()

	if len(pBytes) > constants.PathLengthLimitBytes {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "Path is too long.",
		}, fasthttp.StatusOK)
		return
	}

	if len(pBytes) <= 1 {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "Path is too short.",
		}, fasthttp.StatusOK)
		return
	}

	if len(ctx.QueryArgs().Peek(encryptionKeyParam)) == 0 {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "No encryption key was provided. (enc_key)",
		}, fasthttp.StatusOK)
		return
	}

	p := string(pBytes[1:])

	filePath := path.Join(global.Configuration.Storage.Directory, p)

	// we only need to know if it exists or not
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			ServeNotFound(ctx)
			return
		}
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("os.Stat() could not be called on the file. %v", err),
		}, fasthttp.StatusOK)
		return
	}

	if fileInfo.IsDir() {
		ServeNotFound(ctx)
		return
	}

	if global.Configuration.RateLimit.Bandwidth.Download > 0 && global.Configuration.RateLimit.Bandwidth.ResetAfter > 0 {
		isBandwidthLimitNotReached, err := security.Try(ctx, global.RedisClient, fmt.Sprintf("%s_%s", constants.RateLimitBandwidthDownload, utils.GetIP(ctx)), int64(global.Configuration.RateLimit.Bandwidth.Download), int64(global.Configuration.RateLimit.Bandwidth.ResetAfter), fileInfo.Size())
		if err != nil {
			response.SendJSONResponse(ctx, response.JSONResponse{
				Status:  response.RequestStatusInternalError,
				Data:    nil,
				Message: fmt.Sprintf("Bandwidth limit couldn't be checked. %v", err),
			}, fasthttp.StatusOK)
			return
		}
		if !isBandwidthLimitNotReached {
			response.SendJSONResponse(ctx, response.JSONResponse{
				Status:  response.RequestStatusError,
				Data:    nil,
				Message: "Download bandwidth limit reached; try again later.",
			}, fasthttp.StatusTooManyRequests)
			return
		}
	}

	// We don't need a limited reader because mimetype.DetectReader automatically caps it
	fileReader, e := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if e != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("The file could not be opened. %v", err),
		}, fasthttp.StatusOK)
		return
	}
	defer func() {
		_ = fileReader.Close()
	}()

	key, err := encryption.DeriveKey(ctx.QueryArgs().Peek(encryptionKeyParam), []byte(global.Configuration.Encryption.Nonce))
	if err != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Failed to generate encryption key. %v", err),
		}, fasthttp.StatusOK)
		return
	}

	decryptedReader, err := sio.DecryptReader(fileReader, sio.Config{Key: key[:]})
	if err != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Failed to create a decrypted reader for mime type inspection. %v", e),
		}, fasthttp.StatusOK)
		return
	}

	mimeType, e := mimetype.DetectReader(decryptedReader)
	if e != nil {
		response.SendInvalidEncryptionKeyResponse(ctx)
		return
	}

	filterStatus := security.FilterCheck(ctx, mimeType.String())
	if filterStatus == security.FilterFail {
		// already sent a response if filter check failed
		return
	} else if filterStatus == security.FilterSanitize {
		ctx.Response.Header.Set("Content-Type", "text/plain; charset=utf8")
	} else {
		ctx.Response.Header.Set("Content-Type", mimeType.String())
	}

	ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", p))
	ctx.Response.Header.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	if discordBotRegex.Match(ctx.Request.Header.UserAgent()) && !ctx.QueryArgs().Has(rawParam) {
		if mimetype.EqualsAny(mimeType.String(), "image/png", "image/jpeg", "image/gif") {
			ctx.Response.Header.SetContentType("text/html; charset=utf8")
			ctx.Response.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
			ctx.Response.Header.Add("Pragma", "no-cache")
			ctx.Response.Header.Add("Expires", "0")

			u := fmt.Sprintf("%s/%s?%s=true&enc_key=%s", global.Configuration.Domain, p, rawParam, string(ctx.QueryArgs().Peek(encryptionKeyParam)))
			_, _ = fmt.Fprint(ctx.Response.BodyWriter(), strings.Replace(discordHTML, "{{.}}", u, 1))
			return
		}
	}

	_, err = fileReader.Seek(0, io.SeekStart)
	if err != nil {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Failed to reset file reader to 0. %v", err),
		}, fasthttp.StatusOK)
		return
	}

	if _, err = sio.Decrypt(ctx.Response.BodyWriter(), fileReader, sio.Config{Key: key[:]}); err != nil {
		if _, ok := err.(sio.Error); ok {
			response.SendInvalidEncryptionKeyResponse(ctx)
			return
		}
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusInternalError,
			Data:    nil,
			Message: fmt.Sprintf("Failed to write decrypted file to the response body. %v", err),
		}, fasthttp.StatusOK)
	}
}
