package utils

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
)

func GetRoot(ctx *fasthttp.RequestCtx) string {
	protocol := "https"
	if strings.Contains(string(ctx.Request.Host()), "localhost:") {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s", protocol, ctx.Request.Host())
}
