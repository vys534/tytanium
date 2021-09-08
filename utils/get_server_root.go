package utils

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
)

// GetServerRoot gets the base domain that the server is using.
// If localhost is being used, it replaces https with http.
func GetServerRoot(ctx *fasthttp.RequestCtx) string {
	protocol := "https"
	if strings.Contains(string(ctx.Request.Host()), "localhost:") {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s", protocol, ctx.Request.Host())
}
