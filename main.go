package main

import (
	"context"
	"database/sql"
	"time"

	"robot/app"
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
	ctx, dbn, err = db.ConnMySQL("rw_mysql_dbqueue")

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
		rb.Run(ctx, dbn)

		delay = time.Second * 1

		return
	}

	if err := evio.Serve(events, "tcp://localhost:5000"); err != nil {
		log := &app.Logs{}
		log.LogWrite(err)

		panic(err.Error())
	}
}
