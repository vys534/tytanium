package main

import (
	"github.com/valyala/fasthttp"
)

// ServeCheckAuth validates either the standard or master key.
func (b *BaseHandler) ServeCheckAuth(ctx *fasthttp.RequestCtx) {
	if !b.IsAuthorized(ctx) {
		return
	}
	SendNothing(ctx)
}
