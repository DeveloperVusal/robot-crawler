package database

import (
	"github.com/redis/go-redis/v9"

	"robot/config"
)

type Redis struct{}

func (s *Redis) Init() *redis.Client {
	cfg := config.ConfigDatabaseLoad()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg["redis"]["host"] + ":" + cfg["redis"]["port"],
		Password: cfg["redis"]["password"], // no password set
	})

	return rdb
}
