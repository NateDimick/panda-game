package redisconn

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConn interface {
	Get(context.Context, string) *redis.StringCmd
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
	Del(context.Context, ...string) *redis.IntCmd
	Subscribe(context.Context, ...string) *redis.PubSub
}
