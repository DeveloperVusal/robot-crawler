package main

import (
	"time"

	"robot/core"

	"github.com/tidwall/evio"
)

func main() {
	bgs := &core.Robotgo{}

	var events evio.Events

	events.Tick = func() (delay time.Duration, action evio.Action) {
		bgs.IsQueue()

		delay = time.Second * 5
		return
	}

	if err := evio.Serve(events, "tcp://localhost:5000"); err != nil {
		panic(err.Error())
	}
}
