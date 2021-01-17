package main

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/httputil/net"
	"github.com/vysiondev/httputil/rand"
	"io"
	"path"
	"sync"
)

const fileHandler = "file"

// ServeUpload handles all incoming POST requests to /upload. It will take a multipart form, parse the file, then write it to GCS.
// The file's information will also be inserted into the database.
func (b *BaseHandler) ServeUpload(ctx *fasthttp.RequestCtx) {
	auth := b.IsAuthorized(ctx)
	if !auth && !b.Config.Security.PublicMode {
		SendTextResponse(ctx, "Not authorized to upload.", fasthttp.StatusUnauthorized)
		return
	}
	mp, e := ctx.Request.MultipartForm()
	if e != nil {
		if e == fasthttp.ErrNoMultipartForm {
			SendTextResponse(ctx, "Multipart form not sent.", fasthttp.StatusBadRequest)
			return
		}
		SendTextResponse(ctx, "There was a problem parsing the form. "+e.Error(), fasthttp.StatusBadRequest)
		return
	}
	defer ctx.Request.RemoveMultipartFormFiles()
	if len(mp.File[fileHandler]) == 0 {
		SendTextResponse(ctx, "No files were uploaded.", fasthttp.StatusBadRequest)
		return
	}
	f := mp.File[fileHandler][0]

	if b.Config.Security.BandwidthLimit.Upload > 0 && b.Config.Security.BandwidthLimit.ResetAfter > 0 {
		isUploadBandwidthLimitNotReached, err := Try(ctx, b.RedisClient, fmt.Sprintf("BW_UP_%s", net.GetIP(ctx)), b.Config.Security.BandwidthLimit.Upload, b.Config.Security.RateLimit.ResetAfter, f.Size)
		if err != nil {
			SendTextResponse(ctx, "There was a problem checking bandwidth limits for uploading. "+err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		if !isUploadBandwidthLimitNotReached {
			SendTextResponse(ctx, "Upload bandwidth limit reached; try again later.", fasthttp.StatusTooManyRequests)
			return
		}
	}

	openedFile, e := f.Open()
	if e != nil {
		SendTextResponse(ctx, "Failed to open file from request: "+e.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer openedFile.Close()

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
		SendTextResponse(ctx, "Failed to reset the reader.", fasthttp.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	randomStringChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		rand.RandBytes(b.Config.Server.IDLen, randomStringChan, func() { wg.Done() })
	}()
	wg.Wait()
	fileId := <-randomStringChan
	fileName := fileId + path.Ext(f.Filename)

	wc := b.GCSClient.Bucket(b.Config.Net.GCS.BucketName).Object(fileName).Key(b.Key).NewWriter(ctx)
	defer wc.Close()

	_, writeErr := io.Copy(wc, openedFile)
	if writeErr != nil {
		SendTextResponse(ctx, "There was a problem writing the file to GCS. "+writeErr.Error(), fasthttp.StatusInternalServerError)
		return
	}

	u := fmt.Sprintf("%s/%s", net.GetRoot(ctx), fileName)
	SendTextResponse(ctx, u, fasthttp.StatusOK)
}
