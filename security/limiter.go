package security

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

// Try checks to see if the ID exceeds the max count specified. Additionally, a time to reset the limit and the increment amount is also specified.
// Based on https://github.com/etcinit/speedbump/blob/6ba8259841f7fb6dad77d2163bbf1c3794dc1853/speedbump.go#L99.
func Try(ctx context.Context, redisClient *redis.Client, id string, max int64, resetAfter int64, incrBy int64) (bool, error) {
	exists := true
	val, err := redisClient.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			exists = false
		} else {
			return false, err
		}
	}

	if exists {
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return false, err
		}

		if intVal >= max {
			return false, nil
		}
	}

	err = redisClient.Watch(ctx, func(rx *redis.Tx) error {
		_, err := rx.Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
			if err := pipeliner.IncrBy(ctx, id, incrBy).Err(); err != nil {
				return err
			}
			return pipeliner.Expire(ctx, id, time.Duration(resetAfter)*time.Millisecond).Err()
		})

		return err
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
