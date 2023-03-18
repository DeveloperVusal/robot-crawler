package app

import (
	"log"
	"os"
	"path/filepath"
)

type Logs struct{}

func (l *Logs) LogWrite(_err error) {
	ex, err := os.Executable()

	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	filename := filepath.Join(exPath, "/logs/app.log")

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	log.SetOutput(f)
	log.Println(_err)
}
