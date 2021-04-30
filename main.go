package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"log"
	"os"
	"time"
)

const (
	Version = "1.13.4"
)

func main() {
	log.Print(">> Tytanium " + Version + "\n\n")

	viper.SetConfigName("config")
	viper.AddConfigPath("./conf/")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	var configuration Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Set undefined variables
	viper.SetDefault("storage.directory", "files")
	viper.SetDefault("server.port", "3030")
	viper.SetDefault("net.redis.db", 0)
	viper.SetDefault("server.idlen", 5)
	viper.SetDefault("server.concurrency", 128*4)
	viper.SetDefault("server.collisioncheckattempts", 3)
	viper.SetDefault("security.maxsizebytes", 52428800)
	viper.SetDefault("security.publicmode", false)
	viper.SetDefault("security.ratelimit.resetafter", 60000)
	viper.SetDefault("security.bandwidthlimit.resetafter", 60000*5)
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(configuration.Security.MasterKey) == 0 {
		log.Fatal("A master key MUST be set.")
	}
	if configuration.Security.PublicMode {
		log.Println("** WARNING: Public mode is ENABLED. Authentication will not be required to upload! **")
	}

	i, err := os.Stat(configuration.Storage.Directory)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Creating new directory at", configuration.Storage.Directory)
			err = os.MkdirAll(configuration.Storage.Directory, 0777)
			if err != nil {
				log.Fatalf("Unable to create directory for files, %v", err)
			}
		} else {
			log.Fatalf("Can't stat the files directory, %v", err)
		}
	}
	if i != nil && !i.IsDir() {
		log.Fatalf("Specified storage path is not a directory.")
	}

	log.Println("Saving files to directory", configuration.Storage.Directory)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     configuration.Net.Redis.URI,
		Password: configuration.Net.Redis.Password,
		DB:       configuration.Net.Redis.Db,
	})

	status := redisClient.Ping(ctx).Err()
	if status != nil {
		log.Fatal("Could not ping Redis database: " + status.Error())
	}
	log.Println("Redis connection established")

	b := NewBaseHandler(redisClient, configuration)

	s := &fasthttp.Server{
		ErrorHandler:                  nil,
		Handler:                       handleCORS(b.limitPath(b.handleHTTPRequest)),
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Concurrency:                   configuration.Server.Concurrency,
		DisableKeepalive:              false,
		ReadTimeout:                   30 * time.Minute,
		WriteTimeout:                  30 * time.Minute,
		TCPKeepalive:                  false,
		TCPKeepalivePeriod:            0,
		MaxRequestBodySize:            configuration.Security.MaxSizeBytes + 2048,
		ReduceMemoryUsage:             false,
		GetOnly:                       false,
		DisablePreParseMultipartForm:  false,
		LogAllErrors:                  false,
		DisableHeaderNamesNormalizing: false,
		NoDefaultServerHeader:         false,
		NoDefaultDate:                 false,
		NoDefaultContentType:          false,
		KeepHijackedConns:             false,
	}

	log.Println(">> Listening for new requests on port " + b.Config.Server.Port)
	if err = s.ListenAndServe(":" + b.Config.Server.Port); err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}

}
