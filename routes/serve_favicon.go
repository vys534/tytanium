package routes

import (
	"bytes"
	_ "embed"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/logger"
	"io"
)

//go:embed favicon.ico
var Favicon []byte

const (
	FaviconContentType = "image/x-icon"
)

// ServeFavicon returns the favicon image.
func ServeFavicon(ctx *fasthttp.RequestCtx) {
	b := bytes.NewBuffer(Favicon)
	ctx.Response.Header.Set("Content-Type", FaviconContentType)
	_, e := io.Copy(ctx.Response.BodyWriter(), b)
	if e != nil {
		if global.Configuration.Logging.Enabled {
			logger.ErrorLogger.Printf("Failed to send favicon: %v", e)
		}
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}
