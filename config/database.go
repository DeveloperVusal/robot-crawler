package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Метод получает данные для подключения к БД
func ConfigDatabaseLoad() map[string]map[string]string {
	appConfig := &AppConfig{}
	projectPath := appConfig.Get("projectPath")

	if projectPath == "" {
		fl, _ := os.Getwd()
		projectPath = filepath.Join(filepath.Dir(fl), filepath.Base(fl))
	}

	filename := filepath.Join(projectPath, "/.env")

	var err = godotenv.Load(filename)

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return map[string]map[string]string{
		"pgsql": {
			"driver":   "pgsql",
			"servname": os.Getenv("DB_SERVNAME"),
			"host":     os.Getenv("DB_HOST"),
			"port":     os.Getenv("DB_PORT"),
			"username": os.Getenv("DB_USERNAME"),
			"password": os.Getenv("DB_PASSWORD"),
			"dbname":   os.Getenv("DB_DATABASE"),
		},
		"mysql": {
			"driver":   "mysql",
			"servname": os.Getenv("DB_SERVNAME_2"),
			"host":     os.Getenv("DB_HOST_2"),
			"port":     os.Getenv("DB_PORT_2"),
			"username": os.Getenv("DB_USERNAME_2"),
			"password": os.Getenv("DB_PASSWORD_2"),
			"dbname":   os.Getenv("DB_DATABASE_2"),
		},
		"solr": {
			"scheme": os.Getenv("SOLR_SCHEME"),
			"host":   os.Getenv("SOLR_HOST"),
			"port":   os.Getenv("SOLR_PORT"),
			"core":   os.Getenv("SOLR_CORE"),
		},
	}
}
