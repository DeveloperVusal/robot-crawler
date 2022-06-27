package core

import (
	"robot/app"
)

type Robotgo struct{}

// Метод запускает работу робота
// Вызывает метод app.Queue.IsQueue()
func (rg *Robotgo) Run() {
	appqueue := &app.Queue{}

	appqueue.IsQueue()
}
