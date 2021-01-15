package main

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"google.golang.org/api/option"
	"log"
	"time"
)

const (
	Version   = "1.13.0"
	GCSKeyLoc = "./conf/key.json"
)

func main() {
	log.Print("Tytanium " + Version + "\n\n")

	viper.SetConfigName("config")
	viper.AddConfigPath("./conf/")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	var configuration Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Set undefined variables
	viper.SetDefault("server.port", "3030")
	viper.SetDefault("net.redis.db", 0)
	viper.SetDefault("server.idlen", 5)
	viper.SetDefault("server.concurrency", 128*4)
	viper.SetDefault("server.maxconnsperip", 16)
	viper.SetDefault("security.maxsizebytes", 52428800)
	viper.SetDefault("security.publicmode", false)
	// Make 20 requests globally per minute. Overrides all path-specific rate limits.
	viper.SetDefault("security.ratelimit.global", 20)
	// Upload 10 times per minute.
	viper.SetDefault("security.ratelimit.upload", 10)
	viper.SetDefault("security.ratelimit.resetafter", time.Minute)
	// Download 50 MB per 5 minutes.
	viper.SetDefault("security.bandwidthlimit.download", 52428800)
	// Upload 250 MB per 5 minutes.
	viper.SetDefault("security.bandwidthlimit.upload", 262144000)
	viper.SetDefault("security.bandwidthlimit.resetafter", time.Minute*5)
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(configuration.Security.MasterKey) == 0 {
		log.Fatal("A master key MUST be set.")
	}
	if configuration.Security.PublicMode {
		log.Println("WARNING: Public mode is ENABLED. Authentication will not be required to upload!")
	}

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

	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(GCSKeyLoc))
	if err != nil {
		log.Fatal("Could not instantiate storage client: " + err.Error())
	}
	log.Println("Google Cloud Storage connection established")
	b := NewBaseHandler(c, redisClient, configuration)

	s := &fasthttp.Server{
		ErrorHandler:                  nil,
		Handler:                       b.limitPath(handleCORS(b.handleHTTPRequest)),
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Name:                          "Tytanium " + Version,
		Concurrency:                   configuration.Server.Concurrency,
		DisableKeepalive:              false,
		ReadTimeout:                   30 * time.Minute,
		WriteTimeout:                  30 * time.Minute,
		MaxConnsPerIP:                 configuration.Server.MaxConnsPerIP,
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

	log.Println("-> Listening for new requests on port " + b.Config.Server.Port)
	if err = s.ListenAndServe(":" + b.Config.Server.Port); err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}

}
