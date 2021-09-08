package main

import (
	"context"
	_ "embed"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/middleware"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"
)

func main() {
	log.Print("* Tytanium " + global.Version + "\n\n")

	viper.SetConfigName("config")
	viper.AddConfigPath("./conf/")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	s := &fasthttp.Server{
		ErrorHandler:                  nil,
		Handler:                       middleware.HandleCORS(middleware.LimitPath(middleware.HandleHTTPRequest)),
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Concurrency:                   int(global.Configuration.Server.Concurrency),
		DisableKeepalive:              false,
		ReadTimeout:                   3 * time.Second,
		TCPKeepalive:                  false,
		TCPKeepalivePeriod:            0,
		MaxRequestBodySize:            int(global.Configuration.Storage.MaxSize) + 2048,
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

	portAsString := strconv.Itoa(int(global.Configuration.Server.Port))
	log.Println("Will listen for new requests on port " + portAsString)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := s.ListenAndServe(":" + portAsString); err != nil {
			log.Fatalf("Listen error: %v\n", err)
		}
	}()

	if global.Configuration.MoreStats {
		// collect stats every 30 seconds
		go func() {
			for {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				e := global.RedisClient.Set(ctx, "ty_mem_usage", m.Alloc, 0).Err()
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
