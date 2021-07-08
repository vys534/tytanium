package main

import (
	"github.com/valyala/fasthttp"
)

type PingResponse struct {
	Public        bool   `json:"public"`
	ServerVersion string `json:"version"`
	MaxSize       int    `json:"max_size"`
}

func (b *BaseHandler) ServePing(ctx *fasthttp.RequestCtx) {
	SendJSONResponse(ctx, &PingResponse{
		ServerVersion: Version,
		MaxSize:       b.Config.Security.MaxSizeBytes,
	})
}
