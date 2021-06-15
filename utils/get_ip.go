package utils

import "github.com/valyala/fasthttp"

func GetIP(ctx *fasthttp.RequestCtx) string {
	forwardedIP := ctx.Request.Header.Peek("CF-Connecting-IP")
	if len(forwardedIP) != 0 {
		return string(forwardedIP)
	}
	return ctx.RemoteIP().String()
}
