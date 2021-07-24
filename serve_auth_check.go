package main

import (
	"github.com/valyala/fasthttp"
)

// ServeAuthCheck validates either the master key.
func (b *BaseHandler) ServeAuthCheck(ctx *fasthttp.RequestCtx) {
	if !b.IsAuthorized(ctx) {
		return
	}
	SendNothing(ctx)
}
