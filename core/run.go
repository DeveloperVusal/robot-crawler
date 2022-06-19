package core

import "robot/app"

type Robotgo struct{}

func (rg *Robotgo) Run() {
	appqueue := &app.Queue{}

	appqueue.IsQueue()
}
