package response

import (
	json2 "encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/logger"
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
		if global.Configuration.Logging.Enabled {
			logger.ErrorLogger.Printf("500 response sent; error message: %s", msg)
		}
	}

	ctx.SetStatusCode(code)
	_, e := fmt.Fprint(ctx.Response.BodyWriter(), msg)
	if e != nil {
		log.Printf(fmt.Sprintf("Request failed to send! %v, status code %d", e, code))
		if global.Configuration.Logging.Enabled {
			logger.ErrorLogger.Printf("Failed to send response; error message: %s, status code: %d", e, code)
		}
	}
}

// SendJSONResponse sends a JSON encoded response to the client along with an HTTP status code of 200 OK.
func SendJSONResponse(ctx *fasthttp.RequestCtx, json interface{}) {
	ctx.SetContentType(jsonContentType)
	e := json2.NewEncoder(ctx.Response.BodyWriter()).Encode(json)
	if e != nil {
		if global.Configuration.Logging.Enabled {
			logger.ErrorLogger.Printf("Failed to send JSON response; error message: %v", e)
		}
		log.Printf(fmt.Sprintf("JSON failed to send! %v", e))
	}

}

// SendNothing sends 204 No Content.
// TODO: remove in the future if not needed.
func SendNothing(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}
