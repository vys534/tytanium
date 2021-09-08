package security

import (
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/response"
)

// IsAuthorized compares the Authorization header to the master key. If they don't match,
// HTTP status code 401 is returned.
func IsAuthorized(ctx *fasthttp.RequestCtx) bool {
	if string(ctx.Request.Header.Peek("authorization")) != global.Configuration.Security.MasterKey {
		response.SendTextResponse(ctx, "Not authorized.", fasthttp.StatusUnauthorized)
		return false
	}
	return true
}
