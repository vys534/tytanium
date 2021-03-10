package main

import (
	"github.com/go-redis/redis/v8"
)

type BaseHandler struct {
	RedisClient *redis.Client
	Config      Configuration
}

func NewBaseHandler(redisClient *redis.Client, c Configuration) *BaseHandler {
	return &BaseHandler{
		RedisClient: redisClient,
		Config:      c,
	}
}
