package main

import (
	"cloud.google.com/go/storage"
	"encoding/base64"
	"github.com/go-redis/redis/v8"
	"log"
)

type BaseHandler struct {
	GCSClient   *storage.Client
	RedisClient *redis.Client
	Key         []byte
	Config      Configuration
}

func NewBaseHandler(gcsClient *storage.Client, redisClient *redis.Client, c Configuration) *BaseHandler {
	k, e := base64.StdEncoding.DecodeString(c.Net.GCS.SecretKey)
	if e != nil {
		log.Fatal("Key not properly formatted to Base64.")
	}

	return &BaseHandler{
		GCSClient:   gcsClient,
		Key:         k,
		RedisClient: redisClient,
		Config:      c,
	}
}
