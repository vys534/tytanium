package routes

import (
	"bytes"
	_ "embed"
	"github.com/valyala/fasthttp"
	"io"
)

//go:embed favicon.ico
var Favicon []byte

const (
	FaviconContentType = "image/x-icon"
)

// ServeFavicon returns the favicon's image.
func ServeFavicon(ctx *fasthttp.RequestCtx) {
	b := bytes.NewBuffer(Favicon)
	ctx.Response.Header.Set("Content-Type", FaviconContentType)
	_, e := io.Copy(ctx.Response.BodyWriter(), b)
	if e != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}
