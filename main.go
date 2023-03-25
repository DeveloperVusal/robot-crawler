package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"robot/app"
	"robot/core"
	"robot/database"
	"robot/helpers"

	"github.com/redis/go-redis/v9"
	"github.com/tidwall/evio"
)

var eventloop uint64 = 1
var rdb *redis.Client
var max_threads uint64

func init() {
	var err error

	redis := database.Redis{}

	// Подключаемся к БД
	rdb = redis.Init()

	// Если есть ошибки выводим в лог
	if err != nil {
		log := &app.Logs{}
		log.LogWrite(err)
	}

	env := helpers.Env{}
	env.LoadEnv()

	env_threads := env.Env("MAX_THREADS")
	max_threads, _ = strconv.ParseUint(env_threads, 10, 64)
}

func main() {
	rb := &core.Robotgo{}

	var events evio.Events

	events.Tick = func() (delay time.Duration, action evio.Action) {
		fmt.Println("EventLoop", eventloop)
		rb.Run(context.Background(), rdb, max_threads)
		fmt.Println("")

		delay = time.Second * 1

		eventloop++

		return
	}

	if err := evio.Serve(events, "tcp://localhost:5000"); err != nil {
		log := &app.Logs{}
		log.LogWrite(err)

		panic(err.Error())
	}
}
