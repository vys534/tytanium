package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/utils"
	"strings"
	"time"
)

type LimitPath int

const (
	LimitUploadPath LimitPath = iota
	LimitGeneralPath
)

// limit rate limits based on what LimitPath is given, so that a global rate limit isn't necessary.
// IPs will be stored like 0_192.168.1.1 for path 0, 1_192.168.1.1 for path 1, and so on.
// Bandwidth checking for uploading is set as BW_UP_192.168.1.1, for another example.
func (b *BaseHandler) limitPath(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if b.Config.Security.RateLimit.ResetAfter <= 0 {
			h(ctx)
		} else {
			p := string(ctx.Request.URI().Path())
			pathType := LimitGeneralPath
			reqLimit := b.Config.Security.RateLimit.Global

			switch strings.ToLower(p) {
			case "/upload":
				pathType = LimitUploadPath
				reqLimit = b.Config.Security.RateLimit.Upload
			}
			if reqLimit <= 0 {
				h(ctx)
			} else {
				rlString := ""
				// Check the global rate limit
				isGlobalRateLimitOk, err := Try(ctx, b.RedisClient, fmt.Sprintf("G_%s", utils.GetIP(ctx)), b.Config.Security.RateLimit.Global, b.Config.Security.RateLimit.ResetAfter, 1)
				if err != nil {
					SendTextResponse(ctx, "Failed to call Try() to get information on global rate limit. "+err.Error(), fasthttp.StatusInternalServerError)
					return
				}
				if !isGlobalRateLimitOk {
					rlString = "Global"
				}

				if pathType != LimitGeneralPath {
					// Check the route exclusive rate limit
					isPathOk, err := Try(ctx, b.RedisClient, fmt.Sprintf("%d_%s", pathType, utils.GetIP(ctx)), reqLimit, b.Config.Security.RateLimit.ResetAfter, 1)
					if err != nil {
						SendTextResponse(ctx, "Failed to call Try() to get information on path-specific rate limit. "+err.Error(), fasthttp.StatusInternalServerError)
						return
					}

					if !isPathOk {
						rlString = fmt.Sprintf("(path: %d)", pathType)
					}
				}
				if len(rlString) > 0 {
					SendTextResponse(ctx, fmt.Sprintf("You are being rate limited. (path: %s)", rlString), fasthttp.StatusTooManyRequests)
					return
				}
				h(ctx)
			}
		}
	}
}

func handleCORS(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Authorization")
		if ctx.Request.Header.IsOptions() {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		} else {
			h(ctx)
		}
	}
}

func (b *BaseHandler) handleHTTPRequest(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/upload":
		fasthttp.TimeoutHandler(b.ServeUpload, time.Minute*30, "Upload timed out")(ctx)
		break
	case "/favicon.ico":
		ServeFavicon(ctx)
		break
	case "/checkauth":
		b.ServeCheckAuth(ctx)
		break
	case "/ping":
		b.ServePing(ctx)
		break
	default:
		if !ctx.IsGet() {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
		fasthttp.TimeoutHandler(b.ServeFile, time.Minute*30, "Fetching file timed out")(ctx)
	}

}
