package main

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/valyala/fasthttp"
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
func (b *BaseHandler) FilterCheck(ctx *fasthttp.RequestCtx, mimeType string) FilterStatus {
	if len(b.Config.Security.Filter.Blacklist) > 0 && mimetype.EqualsAny(mimeType, b.Config.Security.Filter.Blacklist...) {
		SendTextResponse(ctx, "File type is blacklisted.", fasthttp.StatusBadRequest)
		return FilterFail
	}
	if len(b.Config.Security.Filter.Whitelist) > 0 && !mimetype.EqualsAny(mimeType, b.Config.Security.Filter.Whitelist...) {
		SendTextResponse(ctx, "File type is not whitelisted.", fasthttp.StatusBadRequest)
		return FilterFail
	}
	if len(b.Config.Security.Filter.Sanitize) > 0 && mimetype.EqualsAny(mimeType, b.Config.Security.Filter.Sanitize...) {
		return FilterSanitize
	}
	return FilterPass
}
