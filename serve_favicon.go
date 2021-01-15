package main

import (
	"github.com/valyala/fasthttp"
	"io"
	"os"
)

const FaviconName = "./conf/favicon.ico"

func ServeFavicon(ctx *fasthttp.RequestCtx) {
	f, e := os.OpenFile(FaviconName, os.O_RDONLY, 0666)
	if e != nil {
		if e == os.ErrNotExist {
			ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
		} else {
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		}
		return
	}
	defer f.Close()
	ctx.Response.Header.Set("Content-Type", "image/x-icon")
	_, e = io.Copy(ctx.Response.BodyWriter(), f)
	if e != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}
