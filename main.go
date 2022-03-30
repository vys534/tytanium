package main

import (
	_ "embed"
	"github.com/valyala/fasthttp"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
	"tytanium/constants"
	"tytanium/global"
	"tytanium/logger"
	"tytanium/middleware"
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
		logger.InfoLogger.Printf("Server online, port %s, version %s", portAsString, constants.Version)
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := s.ListenAndServe(":" + portAsString); err != nil {
			log.Fatalf("Listen error: %v\n", err)
		}
	}()

	<-stop
	log.Println("Server is shutting down, please wait")
	logger.InfoLogger.Println("Server started graceful shutdown")

	if err := s.Shutdown(); err != nil {
		logger.ErrorLogger.Printf("Server failed to shut down gracefully: %v", err)
		log.Fatalf("Failed to shutdown gracefully: %v\n", err)
	}

	log.Println("Shut down. See you next time!")
	logger.InfoLogger.Println("Server shut down successfully")
	os.Exit(0)
}
