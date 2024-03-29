package redisconn

import (
	"context"
	"os"
	"pandagame/internal/util"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisConn() RedisConn {
	conn := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})

	return conn
}

func GetThing[T any](key string, conn RedisConn) (*T, error) {
	resp := conn.Get(context.Background(), key)
	if resp.Err() != nil {
		return nil, resp.Err()
	}

	return util.FromJSONString[T](resp.Val())
}

func SetThing[T any](key string, thing *T, conn RedisConn) error {
	s, err := util.ToJSONString(thing)
	if err != nil {
		return err
	}

	if err := conn.Set(context.Background(), key, s, time.Hour).Err(); err != nil {
		return err
	}
	return nil
}

func DelThing(key string, conn RedisConn) error {
	result := conn.Del(context.Background(), key)
	return result.Err()
}
