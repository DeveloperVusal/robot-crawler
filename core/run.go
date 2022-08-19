package core

import (
	"context"
	"database/sql"
	"robot/app"
)

type Robotgo struct{}

// Метод запускает работу робота
// Вызывает метод app.Queue.RunQueue()
func (rg *Robotgo) Run(ctx context.Context, mysql *sql.DB) {
	appqueue := &app.Queue{
		DBLink: mysql,
		Ctx:    ctx,
	}

	// appqueue.ClearQueue()
	appqueue.ContinueQueue()
	appqueue.RunQueue()
}
