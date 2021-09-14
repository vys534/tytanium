package routes

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/constants"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/response"
	"github.com/vysiondev/tytanium/security"
	"github.com/vysiondev/tytanium/utils"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	rawParam                    = "raw"
	ZeroWidthCharacterFirstByte = 243
)

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
		response.SendTextResponse(ctx, "Path is too long.", fasthttp.StatusBadRequest)
		return
	}

	if len(pBytes) <= 1 {
		response.SendTextResponse(ctx, "Path is too short.", fasthttp.StatusBadRequest)
		return
	}

	p := string(pBytes[1:])
	// Convert entire path to normal string if a zero-width character is detected at the beginning.
	if pBytes[1] == ZeroWidthCharacterFirstByte {
		p = utils.StringToZeroWidthCharacters(p)
		if len(p) == 0 {
			response.SendTextResponse(ctx, "Malformed zero-width URL path.", fasthttp.StatusBadRequest)
			return
		}
	}

	filePath := path.Join(global.Configuration.Storage.Directory, p)

	// we only need to know if it exists or not
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			ServeNotFound(ctx)
			return
		}
		response.SendTextResponse(ctx, fmt.Sprintf("os.Stat() could not be called on the file. %v", err), fasthttp.StatusInternalServerError)
		return
	}

	if fileInfo.IsDir() {
		response.SendTextResponse(ctx, "This is a directory, not a file", fasthttp.StatusBadRequest)
		return
	}

	// We don't need a limited reader because mimetype.DetectReader automatically caps it
	fileReader, e := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if e != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("The file could not be opened. %v", err), fasthttp.StatusInternalServerError)
		return
	}
	defer func() {
		_ = fileReader.Close()
	}()

	if global.Configuration.RateLimit.Bandwidth.Download > 0 && global.Configuration.RateLimit.Bandwidth.ResetAfter > 0 {
		isBandwidthLimitNotReached, err := security.Try(ctx, global.RedisClient, fmt.Sprintf("BW_DN_%s", utils.GetIP(ctx)), int64(global.Configuration.RateLimit.Bandwidth.Download), int64(global.Configuration.RateLimit.Bandwidth.ResetAfter), fileInfo.Size())
		if err != nil {
			response.SendTextResponse(ctx, fmt.Sprintf("Bandwidth limit couldn't be checked. %v", err), fasthttp.StatusInternalServerError)
			return
		}
		if !isBandwidthLimitNotReached {
			response.SendTextResponse(ctx, "Download bandwidth limit reached; try again later.", fasthttp.StatusTooManyRequests)
			return
		}
	}

	mimeType, e := mimetype.DetectReader(fileReader)
	if e != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("Cannot detect the mime type of this file retrieved from server. It might be corrupted. %v", e), fasthttp.StatusBadRequest)
		return
	}

	if discordBotRegex.Match(ctx.Request.Header.UserAgent()) && !ctx.QueryArgs().Has(rawParam) {
		if mimetype.EqualsAny(mimeType.String(), "image/png", "image/jpeg", "image/gif") {
			ctx.Response.Header.SetContentType("text/html; charset=utf8")
			ctx.Response.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
			ctx.Response.Header.Add("Pragma", "no-cache")
			ctx.Response.Header.Add("Expires", "0")

			u := fmt.Sprintf("%s/%s?%s=true", utils.GetServerRoot(ctx), p, rawParam)
			_, _ = fmt.Fprint(ctx.Response.BodyWriter(), strings.Replace(discordHTML, "{{.}}", u, 1))
			return
		}
	}

	filterStatus := security.FilterCheck(ctx, mimeType.String())
	if filterStatus == security.FilterFail {
		// already sent a response if filter check failed
		return
	} else if filterStatus == security.FilterSanitize {
		ctx.Response.Header.Set("Content-Type", "text/plain")
	} else {
		ctx.Response.Header.Set("Content-Type", mimeType.String())
	}
	ctx.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", p))
	ctx.Response.Header.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	_, e = fileReader.Seek(0, io.SeekStart)
	if e != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("Reader could not be reset to its initial position. %v", e), fasthttp.StatusInternalServerError)
		return
	}

	_, copyErr := io.Copy(ctx.Response.BodyWriter(), fileReader)
	if copyErr != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("File wasn't written to the client successfully. %v", copyErr), fasthttp.StatusInternalServerError)
		return
	}
}
