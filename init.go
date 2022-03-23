package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/vysiondev/tytanium/constants"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/logger"
	"log"
	"os"
	"time"
)

const (
	mebibyte = 1 << 20
	minute   = 60000
)

func init() {
	log.Print("⬢ Tytanium secure file host server v" + constants.Version + "\n\n")
	initConfiguration()
	initLogger()
	checkStorage()
	initRedis()
	log.Println("[init] Initial checks completed")
}

func initConfiguration() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./conf/")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("No config file set, %v", err)
		} else {
			log.Fatalf("Error reading config file, %v", err)
		}
	}

	viper.SetDefault("Storage.Directory", "files")
	viper.SetDefault("Storage.MaxSize", 50*mebibyte)
	viper.SetDefault("Storage.IDLength", 5)

	viper.SetDefault("RateLimit.ResetAfter", minute)
	viper.SetDefault("RateLimit.Path.Upload", 10)
	viper.SetDefault("RateLimit.Path.Global", 60)
	viper.SetDefault("RateLimit.Bandwidth.ResetAfter", 5*minute)
	viper.SetDefault("RateLimit.Bandwidth.Download", 500*mebibyte)
	viper.SetDefault("RateLimit.Bandwidth.Upload", 1000*mebibyte)

	viper.SetDefault("Server.Port", 3030)
	viper.SetDefault("Server.Concurrency", 128*4)
	viper.SetDefault("Server.ReadTimeout", 5*minute)
	viper.SetDefault("Server.WriteTimeout", 5*minute)

	viper.SetDefault("StatsCollectionInterval", 30000)

	viper.SetDefault("Logging.Enabled", true)
	viper.SetDefault("Logging.LogFile", "log.txt")

	err := viper.Unmarshal(&global.Configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(global.Configuration.Security.MasterKey) == 0 {
		log.Println("Warning: Master key has not set in your configuration. Anyone on the Internet has permission to upload!")
		if !global.Configuration.Security.DisableEmptyMasterKeyWarning {
			log.Println("Continuing in 5 seconds... (you can set Security.DisableEmptyMasterKeyWarning to true to disable this in the configuration)")
			time.Sleep(time.Second * 5)
		}
	}

	// - ID length * 4 bytes,
	// - extension length limit * 4 bytes,
	// - 1 byte for the / character,
	// - 4 bytes for the . character
	constants.PathLengthLimitBytes = (global.Configuration.Storage.IDLength * 4) + (constants.ExtensionLengthLimit * 4) + 5

	log.Println("[init] Loaded configuration")
}

func initLogger() {
	if !global.Configuration.Logging.Enabled {
		return
	}
	file, err := os.OpenFile(global.Configuration.Logging.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file! %v", err)
	}

	logger.InfoLogger = log.New(file, "info:", log.Ldate|log.Ltime|log.Lshortfile)
	logger.ErrorLogger = log.New(file, "error:", log.Ldate|log.Ltime|log.Lshortfile)

	log.Println("[init] Loggers initialized, output file: " + global.Configuration.Logging.LogFile)
}

func checkStorage() {
	i, err := os.Stat(global.Configuration.Storage.Directory)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("The storage directory %s doesn't exist. Did you forget to create it?", global.Configuration.Storage.Directory)
		} else {
			log.Fatalf("Can't stat the files directory, %v", err)
		}
	}
	if i != nil && !i.IsDir() {
		log.Fatalf("Specified storage path (%s) is not a directory or not usable.", global.Configuration.Storage.Directory)
	}
	log.Println("[init] Storage directory is OK")
}

func initRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	global.RedisClient = redis.NewClient(&redis.Options{
		Addr:     global.Configuration.Redis.URI,
		Password: global.Configuration.Redis.Password,
		DB:       global.Configuration.Redis.DB,
	})

	status := global.RedisClient.Ping(ctx).Err()
	if status != nil {
		cancel()
		log.Fatalf("Could not ping Redis database, %v", status.Error())
	}
	cancel()

	log.Println("[init] Redis database connection established")
}
