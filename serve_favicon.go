package main

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"io"
)

func ServeFavicon(ctx *fasthttp.RequestCtx) {
	f, e := Favicon.ReadFile("./conf/favicon.ico")
	if e != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	b := bytes.NewBuffer(f)
	ctx.Response.Header.Set("Content-Type", "image/x-icon")
	_, e = io.Copy(ctx.Response.BodyWriter(), b)
	if e != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}
