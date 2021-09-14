package middleware

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/constants"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/response"
	"github.com/vysiondev/tytanium/routes"
	"github.com/vysiondev/tytanium/security"
	"github.com/vysiondev/tytanium/utils"
	"strings"
	"time"
)

// LimitPath generally handles all paths.
// IPs will be stored like 0_192.168.1.1 for path 0, 1_192.168.1.1 for path 1, and so on.
// Bandwidth checking for uploading is set as BW_UP_192.168.1.1, for another example.
func LimitPath(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ip := utils.GetIP(ctx)
		if global.Configuration.RateLimit.ResetAfter <= 0 {
			h(ctx)
		} else {
			p := string(ctx.Request.URI().Path())
			pathType := constants.LimitGeneralPath
			reqLimit := global.Configuration.RateLimit.Path.Global

			switch strings.ToLower(p) {
			case "/upload":
				pathType = constants.LimitUploadPath
				reqLimit = global.Configuration.RateLimit.Path.Upload
			}
			if reqLimit <= 0 {
				h(ctx)
			} else {
				rlString := ""
				// Check the global rate limit
				isGlobalRateLimitOk, err := security.Try(ctx, global.RedisClient, fmt.Sprintf("G_%s", ip), int64(global.Configuration.RateLimit.Path.Global), int64(global.Configuration.RateLimit.ResetAfter), 1)
				if err != nil {
					response.SendTextResponse(ctx, "Failed to call Try() to get information on global rate limit. "+err.Error(), fasthttp.StatusInternalServerError)
					return
				}
				if !isGlobalRateLimitOk {
					rlString = "Global"
				}

				if pathType != constants.LimitGeneralPath {
					// Check the route exclusive rate limit
					isPathOk, err := security.Try(ctx, global.RedisClient, fmt.Sprintf("%d_%s", pathType, ip), int64(reqLimit), int64(global.Configuration.RateLimit.ResetAfter), 1)
					if err != nil {
						response.SendTextResponse(ctx, "Failed to call Try() to get information on path-specific rate limit. "+err.Error(), fasthttp.StatusInternalServerError)
						return
					}

					if !isPathOk {
						rlString = fmt.Sprintf("(path: %d)", pathType)
					}
				}
				if len(rlString) > 0 {
					response.SendTextResponse(ctx, fmt.Sprintf("You are being rate limited. (path: %s)", rlString), fasthttp.StatusTooManyRequests)
					return
				}
				h(ctx)
			}
		}
	}
}

// HandleCORS returns headers if the request is an OPTIONS request.
func HandleCORS(h fasthttp.RequestHandler) fasthttp.RequestHandler {
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

// HandleHTTPRequest routes HTTP requests to their respective handlers based on the path.
func HandleHTTPRequest(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/upload":
		fasthttp.TimeoutHandler(routes.ServeUpload, time.Minute*30, "Upload timed out")(ctx)
		break
	case "/favicon.ico":
		routes.ServeFavicon(ctx)
		break
	case "/check_auth":
		routes.ServeAuthCheck(ctx)
		break
	case "/stats":
		routes.ServeStats(ctx)
		break
	default:
		if !ctx.IsGet() {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
		fasthttp.TimeoutHandler(routes.ServeFile, time.Minute*30, "Fetching file timed out")(ctx)
	}

}
