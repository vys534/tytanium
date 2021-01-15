package main

import (
	json2 "encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
)

// SendTextResponse sends a plaintext response to the client along with an HTTP status code.
func SendTextResponse(ctx *fasthttp.RequestCtx, msg string, code int) {
	ctx.Response.Header.SetContentType("text/plain; charset=utf8")
	if code == fasthttp.StatusInternalServerError {
		log.Printf(fmt.Sprintf("Unhandled error!, %s", msg))
	}

	ctx.SetStatusCode(code)
	_, _ = fmt.Fprint(ctx.Response.BodyWriter(), msg)
	return
}

// SendJSONResponse sends a JSON encoded response to the client along with an HTTP status code of 200 OK.
func SendJSONResponse(ctx *fasthttp.RequestCtx, json interface{}) {
	ctx.SetContentType("application/json")
	_ = json2.NewEncoder(ctx.Response.BodyWriter()).Encode(json)
}

// SendNothing sends 204 No Content.
func SendNothing(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNoContent)
	return
}
