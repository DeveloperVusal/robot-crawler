package main

import (
	"time"

	"robot/core"

	"github.com/tidwall/evio"
)

func main() {
	rb := &core.Robotgo{}

	var events evio.Events

	events.Tick = func() (delay time.Duration, action evio.Action) {
		rb.Run()

		delay = time.Second * 1

		return
	}

	if err := evio.Serve(events, "tcp://localhost:5000"); err != nil {
		panic(err.Error())
	}
}
