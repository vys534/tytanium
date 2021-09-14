package routes

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/tytanium/constants"
	"github.com/vysiondev/tytanium/global"
	"github.com/vysiondev/tytanium/response"
	"runtime"
	"strconv"
)

// GeneralStats represent all stats returned when making a GET request to /stats.
type GeneralStats struct {
	ServerVersion  string               `json:"server_version"`
	RuntimeVersion string               `json:"runtime_version,omitempty"`
	MemoryUsage    int64                `json:"memory_usage,omitempty"`
	SizeStats      StatsFromSizeChecker `json:"size_stats"`
}

// StatsFromSizeChecker represent all stats returned by an external size checker program.
type StatsFromSizeChecker struct {
	TotalSize      int64 `json:"total_size"`
	FileCount      int64 `json:"file_count"`
	TimeToComplete int64 `json:"time_to_complete"`
	LastUpdated    int64 `json:"last_updated"`
}

// ServeStats serves stats. StatsFromSizeChecker are populated into redis by https://github.com/vysiondev/size-checker.
func ServeStats(ctx *fasthttp.RequestCtx) {
	var stats GeneralStats
	stats.ServerVersion = constants.Version

	totalSize, err := getStatValueFromRedis(ctx, global.RedisClient, "sc_total_size")
	if err != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("An error occurred while trying to get sc_total_size from Redis: %v", err), fasthttp.StatusInternalServerError)
		return
	}
	stats.SizeStats.TotalSize = totalSize

	fileCount, err := getStatValueFromRedis(ctx, global.RedisClient, "sc_file_count")
	if err != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("An error occurred while trying to get sc_file_count from Redis: %v", err), fasthttp.StatusInternalServerError)
		return
	}
	stats.SizeStats.FileCount = fileCount

	timeToComplete, err := getStatValueFromRedis(ctx, global.RedisClient, "sc_time_to_complete")
	if err != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("An error occurred while trying to get sc_time_to_complete from Redis: %v", err), fasthttp.StatusInternalServerError)
		return
	}
	stats.SizeStats.TimeToComplete = timeToComplete

	lastUpdated, err := getStatValueFromRedis(ctx, global.RedisClient, "sc_last_updated")
	if err != nil {
		response.SendTextResponse(ctx, fmt.Sprintf("An error occurred while trying to get sc_last_updated from Redis: %v", err), fasthttp.StatusInternalServerError)
		return
	}
	stats.SizeStats.LastUpdated = lastUpdated

	if global.Configuration.MoreStats {
		memUsage, err := getStatValueFromRedis(ctx, global.RedisClient, "ty_mem_usage")
		if err != nil {
			response.SendTextResponse(ctx, fmt.Sprintf("An error occurred while trying to get ty_mem_usage from Redis: %v", err), fasthttp.StatusInternalServerError)
			return
		}
		stats.MemoryUsage = memUsage
		stats.RuntimeVersion = runtime.Version()
	}

	response.SendJSONResponse(ctx, &stats)
}

func getStatValueFromRedis(ctx *fasthttp.RequestCtx, c *redis.Client, key string) (int64, error) {
	v, e := c.Get(ctx, key).Result()
	if e != nil {
		if e == redis.Nil {
			return 0, nil
		}
		return 0, e
	}
	i, e := strconv.ParseInt(v, 10, 64)
	if e != nil {
		return 0, e
	}
	return i, nil
}
