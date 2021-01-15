package main

import (
	"github.com/valyala/fasthttp"
)

func (b *BaseHandler) IsAuthorized(ctx *fasthttp.RequestCtx) bool {
	if string(ctx.Request.Header.Peek("authorization")) != b.Config.Security.MasterKey {
		SendTextResponse(ctx, "Not authorized.", fasthttp.StatusUnauthorized)
		return false
	}
	return true
}
