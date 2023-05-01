package redisconn

import "github.com/redis/go-redis/v9"

func NewRedisConn() RedisConn {
	conn := redis.NewClient(&redis.Options{
		Addr: "todo",
		DB:   0,
	})

	return conn
}
