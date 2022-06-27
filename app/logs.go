package app

import (
	"log"
	"os"
)

type Logs struct{}

func (l *Logs) LogWrite(err2 error) {
	f, err := os.OpenFile("./../logs/app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	log.SetOutput(f)
	log.Println(err2)
	log.Println("===========================================")
}
