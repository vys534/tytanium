package security

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/valyala/fasthttp"
	"tytanium/global"
	"tytanium/response"
)

type FilterStatus int

const (
	FilterPass FilterStatus = iota
	FilterFail
	FilterSanitize
)

// FilterCheck runs the file's mime type through a series of checks.
// FilterPass means the mime type is fine and no additional action should be taken.
// FilterFail means a response was already returned, and the caller should terminate its function.
// FilterSanitize means the file's Content-Type header returned to the client should be changed to text/plain.
func FilterCheck(ctx *fasthttp.RequestCtx, mimeType string) FilterStatus {
	if len(global.Configuration.Filter.Blacklist) > 0 && mimetype.EqualsAny(mimeType, global.Configuration.Filter.Blacklist...) {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "This file type is blacklisted.",
		}, fasthttp.StatusOK)
		return FilterFail
	}
	if len(global.Configuration.Filter.Whitelist) > 0 && !mimetype.EqualsAny(mimeType, global.Configuration.Filter.Whitelist...) {
		response.SendJSONResponse(ctx, response.JSONResponse{
			Status:  response.RequestStatusError,
			Data:    nil,
			Message: "This file type is not whitelisted.",
		}, fasthttp.StatusOK)
		return FilterFail
	}
	if len(global.Configuration.Filter.Sanitize) > 0 && mimetype.EqualsAny(mimeType, global.Configuration.Filter.Sanitize...) {
		return FilterSanitize
	}
	return FilterPass
}
