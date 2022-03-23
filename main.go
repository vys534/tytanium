package main

import (
	"context"
	_ "embed"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/constants"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/logger"
	"github.com/vysiondev/tytanium/middleware"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"
)

func main() {
	s := &fasthttp.Server{
		ErrorHandler: nil,
		// yo what da fuck
		Handler:                       middleware.HandleCORS(middleware.LimitPath(middleware.HandleHTTPRequest)),
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Concurrency:                   global.Configuration.Server.Concurrency,
		DisableKeepalive:              false,
		ReadTimeout:                   time.Millisecond * time.Duration(global.Configuration.Server.ReadTimeout),
		WriteTimeout:                  time.Millisecond * time.Duration(global.Configuration.Server.WriteTimeout),
		TCPKeepalive:                  false,
		TCPKeepalivePeriod:            0,
		MaxRequestBodySize:            int(global.Configuration.Storage.MaxSize) + constants.RequestMaxBodySizePadding,
		ReduceMemoryUsage:             false,
		GetOnly:                       false,
		DisablePreParseMultipartForm:  true,
		LogAllErrors:                  false,
		DisableHeaderNamesNormalizing: false,
		NoDefaultServerHeader:         true,
		NoDefaultDate:                 true,
		NoDefaultContentType:          true,
		KeepHijackedConns:             false,
	}

	portAsString := strconv.Itoa(global.Configuration.Server.Port)
	log.Println("Server is listening for new requests on port " + portAsString)

	if global.Configuration.Logging.Enabled {
		logger.InfoLogger.Println("Server online, port " + portAsString)
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := s.ListenAndServe(":" + portAsString); err != nil {
			log.Fatalf("Listen error: %v\n", err)
		}
	}()

	if global.Configuration.MoreStats {
		// collect stats every n seconds
		go func() {
			for {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				e := global.RedisClient.Set(ctx, "ty_mem_usage", m.Alloc, 0).Err()
				if e != nil {
					log.Printf("Failed to write metrics! %v", e)
					if global.Configuration.Logging.Enabled {
						logger.ErrorLogger.Printf("Metrics failed to update: %v", e)
					}
				}
				cancel()

				time.Sleep(time.Millisecond * time.Duration(global.Configuration.StatsCollectionInterval))
			}
		}()
	}

	<-stop
	log.Println("Shutting down, please wait")
	logger.InfoLogger.Println("Server started graceful shutdown")

	if err := s.Shutdown(); err != nil {
		logger.ErrorLogger.Printf("Server failed to shut down gracefully: %v", err)
		log.Fatalf("Failed to shutdown gracefully: %v\n", err)
	}

	log.Println("Shut down")
	logger.InfoLogger.Println("Server shut down successfully")
	os.Exit(0)
}
