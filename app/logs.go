package app

import (
	"log"
	"os"
	"path/filepath"

	cfg "robot/config"
)

type Logs struct{}

func (l *Logs) LogWrite(err2 error) {
	appConfig := &cfg.AppConfig{}
	projectPath := appConfig.Get("projectPath")

	filename := filepath.Join(projectPath, "/logs/app.log")
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	log.SetOutput(f)
	log.Println(err2)
}
