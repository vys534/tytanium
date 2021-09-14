package response

import (
	json2 "encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
)

const (
	plainTextContentType = "text/plain; charset=utf8"
	jsonContentType      = "application/json"
)

// SendTextResponse sends a plaintext response to the client along with an HTTP status code.
func SendTextResponse(ctx *fasthttp.RequestCtx, msg string, code int) {
	ctx.Response.Header.SetContentType(plainTextContentType)
	if code == fasthttp.StatusInternalServerError {
		log.Printf(fmt.Sprintf("Unhandled error!, %s", msg))
	}

	ctx.SetStatusCode(code)
	_, e := fmt.Fprint(ctx.Response.BodyWriter(), msg)
	if e != nil {
		log.Printf(fmt.Sprintf("Request failed to send! %v, status code %d", e, code))
	}
}

// SendJSONResponse sends a JSON encoded response to the client along with an HTTP status code of 200 OK.
func SendJSONResponse(ctx *fasthttp.RequestCtx, json interface{}) {
	ctx.SetContentType(jsonContentType)
	e := json2.NewEncoder(ctx.Response.BodyWriter()).Encode(json)
	if e != nil {
		log.Printf(fmt.Sprintf("JSON failed to send! %v", e))
	}

}

// SendNothing sends 204 No Content.
func SendNothing(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}
