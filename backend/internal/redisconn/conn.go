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

func GetRaw(key string, conn RedisConn) (string, error) {
	resp := conn.Get(context.Background(), key)
	if resp.Err() != nil {
		return "", resp.Err()
	}
	return resp.Val(), nil
}

func SetRaw(key, value string, conn RedisConn) error {
	return conn.Set(context.Background(), key, value, time.Hour).Err()
}

func GetThing[T any](key string, conn RedisConn) (*T, error) {
	val, err := GetRaw(key, conn)
	if err != nil {
		return nil, err
	}

	return util.FromJSONString[T](val)
}

func SetThing[T any](key string, thing *T, conn RedisConn) error {
	s, err := util.ToJSONString(thing)
	if err != nil {
		return err
	}

	return SetRaw(key, s, conn)
}

func DelThing(key string, conn RedisConn) error {
	result := conn.Del(context.Background(), key)
	return result.Err()
}
