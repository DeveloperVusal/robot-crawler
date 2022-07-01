package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"robot/core"
	dbpkg "robot/database"

	"github.com/tidwall/evio"
)

var dbn *sql.DB
var ctx context.Context

func init() {
	var err error

	db := dbpkg.Database{}

	// Подключаемся к БД
	ctx, dbn, err = db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	rb := &core.Robotgo{}

	var events evio.Events

	events.Tick = func() (delay time.Duration, action evio.Action) {
		rb.Run(ctx, dbn)

		delay = time.Second * 1

		return
	}

	if err := evio.Serve(events, "tcp://localhost:5000"); err != nil {
		panic(err.Error())
	}
}
