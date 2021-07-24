package main

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/utils"
	"io"
	"os"
	"path"
	"sync"
)

const fileHandler = "file"

// ServeUpload handles all incoming POST requests to /upload. It will take a multipart form, parse the file, then write it to disk.
// The file's information will also be inserted into the database.
func (b *BaseHandler) ServeUpload(ctx *fasthttp.RequestCtx) {
	auth := b.IsAuthorized(ctx)
	if !auth {
		return
	}
	mp, e := ctx.Request.MultipartForm()
	if e != nil {
		if e == fasthttp.ErrNoMultipartForm {
			SendTextResponse(ctx, "You didn't send a multipart form.", fasthttp.StatusBadRequest)
			return
		}
		SendTextResponse(ctx, fmt.Sprintf("The form couldn't be parsed. %v", e), fasthttp.StatusBadRequest)
		return
	}
	defer ctx.Request.RemoveMultipartFormFiles()
	if len(mp.File[fileHandler]) == 0 {
		SendTextResponse(ctx, "No files were sent.", fasthttp.StatusBadRequest)
		return
	}
	f := mp.File[fileHandler][0]

	if b.Config.RateLimit.Bandwidth.Upload > 0 && b.Config.RateLimit.Bandwidth.ResetAfter > 0 {
		isUploadBandwidthLimitNotReached, err := Try(ctx, b.RedisClient, fmt.Sprintf("BW_UP_%s", utils.GetIP(ctx)), b.Config.RateLimit.Bandwidth.Upload, b.Config.RateLimit.Bandwidth.ResetAfter, f.Size)
		if err != nil {
			SendTextResponse(ctx, fmt.Sprintf("Bandwidth limit couldn't be checked. %v", err), fasthttp.StatusInternalServerError)
			return
		}
		if !isUploadBandwidthLimitNotReached {
			SendTextResponse(ctx, "Upload bandwidth limit reached; try again later.", fasthttp.StatusTooManyRequests)
			return
		}
	}

	openedFile, e := f.Open()
	if e != nil {
		SendTextResponse(ctx, fmt.Sprintf("File failed to open. %v", e), fasthttp.StatusInternalServerError)
		return
	}
	defer func() {
		_ = openedFile.Close()
	}()

	mimeType, e := mimetype.DetectReader(openedFile)
	if e != nil {
		SendTextResponse(ctx, "Cannot detect the mime type of this file.", fasthttp.StatusBadRequest)
		return
	}

	status := b.FilterCheck(ctx, mimeType.String())
	if status == FilterFail {
		return
	}
	_, e = openedFile.Seek(0, io.SeekStart)
	if e != nil {
		SendTextResponse(ctx, fmt.Sprintf("Reader could not be reset to its initial position. %v", e), fasthttp.StatusInternalServerError)
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
			utils.RandBytes(int(b.Config.Storage.IDLen), randomStringChan, func() { wg.Done() })
		}()
		wg.Wait()
		fileId := <-randomStringChan
		fileName = fileId + path.Ext(f.Filename)

		i, e := os.Stat(path.Join(b.Config.Storage.Directory, fileName))
		if e != nil {
			if os.IsNotExist(e) || e == os.ErrNotExist {
				break
			}
		}
		if i == nil {
			break
		}
		attempts++
		if attempts >= int(b.Config.Storage.CollisionCheckAttempts) {
			SendTextResponse(ctx, "Tried too many times to find a valid file ID to use. Consider increasing the ID length.", fasthttp.StatusInternalServerError)
			return
		}
	}

	fsFile, err := os.Create(path.Join(b.Config.Storage.Directory, fileName))
	defer func() {
		_ = fsFile.Close()
	}()

	if err != nil {
		if err == os.ErrPermission {
			SendTextResponse(ctx, fmt.Sprintf("Permission to create a file was denied. %v", err), fasthttp.StatusInternalServerError)
			return
		}
		SendTextResponse(ctx, fmt.Sprintf("Could not create the file. %v", err), fasthttp.StatusInternalServerError)
		return
	}

	_, writeErr := io.Copy(fsFile, openedFile)
	if writeErr != nil {
		SendTextResponse(ctx, fmt.Sprintf("The file failed to write to disk. %v", e), fasthttp.StatusInternalServerError)
		return
	}

	if string(ctx.QueryArgs().Peek("zerowidth")) == "1" {
		fileName = StringToZWS(fileName)
	}

	var u string
	if string(ctx.QueryArgs().Peek("omitdomain")) == "1" {
		u = fileName
	} else {
		u = fmt.Sprintf("%s/%s", utils.GetRoot(ctx), fileName)
	}

	SendTextResponse(ctx, u, fasthttp.StatusOK)
}
