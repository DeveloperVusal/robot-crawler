package main

import (
	"context"
	"time"

	"robot/app"
	"robot/core"
	dbpkg "robot/database"

	"github.com/redis/go-redis/v9"
	"github.com/tidwall/evio"
)

var rdb *redis.Client

func init() {
	var err error

	redis := dbpkg.Redis{}

	// Подключаемся к БД
	rdb = redis.Init()

	// Если есть ошибки выводим в лог
	if err != nil {
		log := &app.Logs{}
		log.LogWrite(err)
	}
}

func main() {
	rb := &core.Robotgo{}

	var events evio.Events

	events.Tick = func() (delay time.Duration, action evio.Action) {
		rb.Run(context.Background(), rdb)

		delay = time.Second * 1

		return
	}

	if err := evio.Serve(events, "tcp://localhost:5000"); err != nil {
		log := &app.Logs{}
		log.LogWrite(err)

		panic(err.Error())
	}
}
