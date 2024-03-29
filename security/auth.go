package security

import (
	"github.com/valyala/fasthttp"
	"tytanium/global"
	"tytanium/response"
)

// IsAuthorized compares the Authorization header to the master key. If they don't match,
// HTTP status code 401 is returned.
func IsAuthorized(ctx *fasthttp.RequestCtx) bool {
	if string(ctx.Request.Header.Peek("authorization")) != global.Configuration.Security.MasterKey {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "Not authorized to access that.",
		}, fasthttp.StatusUnauthorized)
		return false
	}
	return true
}
