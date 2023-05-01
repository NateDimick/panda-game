package redisconn

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConn interface {
	Get(context.Context, string) *redis.StringCmd
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
	Subscribe(context.Context, ...string) *redis.PubSub
}
