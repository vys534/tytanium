package response

import (
	json2 "encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
	"tytanium/global"
	"tytanium/logger"
)

const (
	plainTextContentType = "text/plain; charset=utf8"
	jsonContentType      = "application/json"
)

type RequestStatus int

const (
	RequestStatusOK = iota
	RequestStatusError
	RequestStatusInternalError
)

type JSONResponse struct {
	Status  RequestStatus `json:"status"`
	Data    interface{}   `json:"data,omitempty"`
	Message string        `json:"message,omitempty"`
}

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
func SendJSONResponse(ctx *fasthttp.RequestCtx, json interface{}, statusCode int) {
	ctx.SetContentType(jsonContentType)
	ctx.SetStatusCode(statusCode)
	e := json2.NewEncoder(ctx.Response.BodyWriter()).Encode(json)
	if e != nil {
		if global.Configuration.Logging.Enabled {
			logger.ErrorLogger.Printf("Failed to send JSON response; error message: %v", e)
		}
		log.Printf(fmt.Sprintf("JSON failed to send! %v", e))
	}
}

func SendInvalidEncryptionKeyResponse(ctx *fasthttp.RequestCtx) {
	SendJSONResponse(ctx, JSONResponse{
		Status:  RequestStatusError,
		Data:    nil,
		Message: "Encryption key is invalid, or file contents have been modified.",
	}, fasthttp.StatusOK)
}
