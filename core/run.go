package core

import (
	"context"
	"robot/app"

	"github.com/redis/go-redis/v9"
)

type Robotgo struct{}

// Метод запускает работу робота
// Вызывает метод app.Queue.RunQueue()
func (rg *Robotgo) Run(ctx context.Context, redis *redis.Client, max_threads uint64) {
	appqueue := &app.Queue{
		Redis:      redis,
		Ctx:        ctx,
		MaxThreads: max_threads,
	}

	appqueue.ContinueWorkers()
	appqueue.RunQueue()
	go appqueue.SitesQueue()
	go appqueue.ClearQueue()
}
