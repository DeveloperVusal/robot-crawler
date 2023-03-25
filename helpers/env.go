package helpers

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Env struct{}

func (h *Env) LoadEnv() {
	ex, err := os.Executable()

	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	filename := filepath.Join(exPath, "/.env")

	var err2 = godotenv.Load(filename)

	if err2 != nil {
		dir, _ := os.Getwd()
		projectPath := filepath.Join(filepath.Dir(dir), filepath.Base(dir))
		filename := filepath.Join(projectPath, "/.env")

		var err = godotenv.Load(filename)

		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func (h *Env) Env(env string) string {
	return os.Getenv(env)
}
