package utils

import "github.com/valyala/fasthttp"

const (
	// CloudflareForwardedIP is the client's original IP given by CloudFlare in the request header.
	cloudflareForwardedIP = "CF-Connecting-IP"
)

// GetIP gets the forwarded IP from Cloudflare if it's available,
// or gets the remote IP as given by fasthttp.RequestCtx as a fallback.
func GetIP(ctx *fasthttp.RequestCtx) string {
	forwardedIP := ctx.Request.Header.Peek(cloudflareForwardedIP)
	if len(forwardedIP) != 0 {
		return string(forwardedIP)
	}
	return ctx.RemoteIP().String()
}
