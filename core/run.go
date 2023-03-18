package core

import (
	"context"
	"robot/app"

	"github.com/redis/go-redis/v9"
)

type Robotgo struct{}

// Метод запускает работу робота
// Вызывает метод app.Queue.RunQueue()
func (rg *Robotgo) Run(ctx context.Context, redis *redis.Client) {
	appqueue := &app.Queue{
		Redis: redis,
		Ctx:   ctx,
	}

	appqueue.ContinueWorkers()
	appqueue.RunQueue()
	go appqueue.SitesQueue()
}
