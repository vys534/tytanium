package main

import (
	"context"
	_ "embed"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"
)

//go:embed conf/favicon.ico
var Favicon []byte

const (
	Version = "1.2.0"
)

func main() {
	log.Print(">> Tytanium " + Version + "\n\n")

	viper.SetConfigName("config")
	viper.AddConfigPath("./conf/")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	var configuration Configuration

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("No config file set: %v", err)
		} else {
			log.Fatalf("Error reading config file: %v", err)
		}
	}

	// Set undefined variables
	viper.SetDefault("storage.directory", "files")
	viper.SetDefault("storage.idlen", 5)
	viper.SetDefault("storage.collisioncheckattempts", 3)
	viper.SetDefault("server.port", 3030)
	viper.SetDefault("server.concurrency", 128*4)
	viper.SetDefault("ratelimit.resetafter", 60000)
	viper.SetDefault("ratelimit.bandwidth.resetafter", 60000*5)
	viper.SetDefault("security.maxsize", 52428800)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("morestats", false)

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(configuration.Security.MasterKey) == 0 {
		log.Fatal("A master key MUST be set.")
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

	log.Println("Saving all incoming files to directory", configuration.Storage.Directory)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     configuration.Redis.URI,
		Password: configuration.Redis.Password,
		DB:       int(configuration.Redis.Db),
	})

	status := redisClient.Ping(ctx).Err()
	if status != nil {
		cancel()
		log.Fatalf("Could not ping Redis database, %v", status.Error())
	}
	cancel()
	log.Println("Redis connection established")

	b := NewBaseHandler(redisClient, configuration)

	s := &fasthttp.Server{
		ErrorHandler:                  nil,
		Handler:                       handleCORS(b.limitPath(b.handleHTTPRequest)),
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Concurrency:                   int(configuration.Server.Concurrency),
		DisableKeepalive:              false,
		ReadTimeout:                   3 * time.Second,
		TCPKeepalive:                  false,
		TCPKeepalivePeriod:            0,
		MaxRequestBodySize:            int(configuration.Storage.MaxSize) + 2048,
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

	portAsString := strconv.Itoa(int(b.Config.Server.Port))
	log.Println("Will listen for new requests on port " + portAsString)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err = s.ListenAndServe(":" + portAsString); err != nil {
			log.Fatalf("Listen error: %v\n", err)
		}
	}()

	if b.Config.MoreStats {
		// collect stats every 30 seconds
		go func() {
			for {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				e := redisClient.Set(ctx, "ty_mem_usage", m.Alloc, 0).Err()
				if e != nil {
					log.Printf("Failed to write metrics! %v", e)
				}
				cancel()

				time.Sleep(time.Second * 30)
			}
		}()
	}

	<-stop
	log.Println("Shutting down")

	if err := s.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v\n", err)
	}

	log.Println("Shut down")
	os.Exit(0)
}
