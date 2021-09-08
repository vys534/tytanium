package routes

import (
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/response"
	"github.com/vysiondev/tytanium/security"
)

// ServeAuthCheck validates the master key by calling IsAuthorized.
func ServeAuthCheck(ctx *fasthttp.RequestCtx) {
	if !security.IsAuthorized(ctx) {
		return
	}
	response.SendNothing(ctx)
}
