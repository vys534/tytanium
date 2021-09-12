package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/vysiondev/tytanium/global"
	"log"
	"os"
	"time"
)

const (
	mebibyte = 1 << 20
	minute   = 60000
)

func init() {
	log.Print("* Tytanium " + global.Version + "\n\n")
	initConfiguration()
	checkStorage()
	initRedis()
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

	err := viper.Unmarshal(&global.Configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(global.Configuration.Security.MasterKey) == 0 {
		log.Println("Warning: Security.MasterKey is not set. Anyone on the Internet has permission to upload!")
		if !global.Configuration.Security.DisableEmptyMasterKeyWarning {
			log.Println("Continuing in 5 seconds... (you can set Security.DisableEmptyMasterKeyWarning to true to disable this)")
			time.Sleep(time.Second * 5)
		}
	}

	log.Println("Loaded configuration")
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
		log.Fatalf("Specified storage path is not a directory.")
	}
	log.Println("Storage directory is ok")
}

func initRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	global.RedisClient = redis.NewClient(&redis.Options{
		Addr:     global.Configuration.Redis.URI,
		Password: global.Configuration.Redis.Password,
		DB:       int(global.Configuration.Redis.DB),
	})

	status := global.RedisClient.Ping(ctx).Err()
	if status != nil {
		cancel()
		log.Fatalf("Could not ping Redis database, %v", status.Error())
	}
	cancel()

	log.Println("Redis connection established")
}
