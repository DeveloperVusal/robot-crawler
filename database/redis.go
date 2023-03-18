package database

import (
	"robot/config"

	"github.com/redis/go-redis/v9"
)

type Redis struct{}

func (s *Redis) Init() *redis.Client {
	loadCfg := &config.Database{}
	cfg := loadCfg.Load()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg["redis"]["host"] + ":" + cfg["redis"]["port"],
		Password: cfg["redis"]["password"], // no password set
	})

	return rdb
}
