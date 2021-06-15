// Parts of this file were derived from https://github.com/whats-this/cdn-origin/blob/8b05fa8425db01cce519ca8945203f9d3050c33b/main.go#L439.
// The implementation reason was a workaround found by this repository to prevent discord from hiding image URLs.

package main

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/utils"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const rawParam = "raw"
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
	discordBotRegex = regexp.MustCompile("(?i)discordbot")
)

// ServeFile will serve the / endpoint. It gets the "id" variable from mux and tries to find the file's information in the database.
// If an ID is either not provided or not found, the function hands the request off to ServeNotFound.
func (b *BaseHandler) ServeFile(ctx *fasthttp.RequestCtx) {
	id := string(ctx.Request.URI().LastPathSegment())
	if len(id) == 0 {
		b.ServeNotFound(ctx)
		return
	}
	decoded := url.QueryEscape(id)

	// Most likely a zero-with URL but we can check for that
	if strings.HasPrefix(decoded, "%") {
		id = ZWSToString(id)
		if len(id) == 0 {
			SendTextResponse(ctx, "There was a problem converting the path segment to a string.", fasthttp.StatusBadRequest)
			return
		}
	}

	// we only need to know if it exists or not
	fileInfo, err := os.Stat(path.Join(b.Config.Storage.Directory, id))
	if err != nil {
		if os.IsNotExist(err) {
			b.ServeNotFound(ctx)
			return
		}
		SendTextResponse(ctx, "There was a problem calling stat on the file. "+err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	// We don't need a limited reader because mimetype.DetectReader automatically caps it
	fileReader, e := os.OpenFile(path.Join(b.Config.Storage.Directory, id), os.O_RDONLY, 0644)
	if e != nil {
		SendTextResponse(ctx, "There was a problem reading the file. "+e.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer func() {
		_ = fileReader.Close()
	}()

	if b.Config.Security.BandwidthLimit.Download > 0 && b.Config.Security.BandwidthLimit.ResetAfter > 0 {
		isBandwidthLimitNotReached, err := Try(ctx, b.RedisClient, fmt.Sprintf("BW_DN_%s", utils.GetIP(ctx)), b.Config.Security.BandwidthLimit.Download, b.Config.Security.RateLimit.ResetAfter, fileInfo.Size())
		if err != nil {
			SendTextResponse(ctx, "There was a problem checking bandwidth limits. "+err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		if !isBandwidthLimitNotReached {
			SendTextResponse(ctx, "Download bandwidth limit reached; try again later.", fasthttp.StatusTooManyRequests)
			return
		}
	}

	mimeType, e := mimetype.DetectReader(fileReader)
	if e != nil {
		SendTextResponse(ctx, "Cannot detect the mime type of this file retrieved from server. It might be corrupted.", fasthttp.StatusBadRequest)
		return
	}

	if discordBotRegex.Match(ctx.Request.Header.UserAgent()) && !ctx.QueryArgs().Has(rawParam) {
		if mimetype.EqualsAny(mimeType.String(), "image/png", "image/jpeg", "image/gif") {
			ctx.Response.Header.SetContentType("text/html; charset=utf8")
			ctx.Response.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
			ctx.Response.Header.Add("Pragma", "no-cache")
			ctx.Response.Header.Add("Expires", "0")

			u := fmt.Sprintf("%s/%s?%s=true", utils.GetRoot(ctx), id, rawParam)
			_, _ = fmt.Fprint(ctx.Response.BodyWriter(), strings.Replace(discordHTML, "{{.}}", u, 1))
			return
		}
	}

	filterStatus := b.FilterCheck(ctx, mimeType.String())
	if filterStatus == FilterFail {
		return
	} else if filterStatus == FilterSanitize {
		ctx.Response.Header.Set("Content-Type", "text/plain")
	} else {
		ctx.Response.Header.Set("Content-Type", mimeType.String())
	}
	ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", id))
	ctx.Response.Header.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	_, e = fileReader.Seek(0, io.SeekStart)
	if e != nil {
		SendTextResponse(ctx, "Failed to reset the reader to 0.", fasthttp.StatusInternalServerError)
		return
	}

	_, copyErr := io.Copy(ctx.Response.BodyWriter(), fileReader)
	if copyErr != nil {
		SendTextResponse(ctx, "Could not write file to client. "+copyErr.Error(), fasthttp.StatusInternalServerError)
		return
	}
}
