package routes

import (
	"github.com/valyala/fasthttp"
	"tytanium/response"
)

// ServeNotFound will always return an HTTP status code of 404 + error message text.
func ServeNotFound(ctx *fasthttp.RequestCtx) {
	response.SendJSONResponse(ctx, response.JSONResponse{
		Status:  response.RequestStatusError,
		Data:    nil,
		Message: "Not found.",
	}, fasthttp.StatusNotFound)
}
