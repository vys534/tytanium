package main

import (
	"github.com/valyala/fasthttp"
)

func (b *BaseHandler) ServeNotFound(ctx *fasthttp.RequestCtx) {
	SendTextResponse(ctx, "Not found", fasthttp.StatusNotFound)
}
