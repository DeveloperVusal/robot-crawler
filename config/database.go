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
			"servname": os.Getenv("PGSQL_SERVNAME"),
			"host":     os.Getenv("PGSQL_HOST"),
			"port":     os.Getenv("PGSQL_PORT"),
			"username": os.Getenv("PGSQL_USERNAME"),
			"password": os.Getenv("PGSQL_PASSWORD"),
			"dbname":   os.Getenv("PGSQL_DATABASE"),
		},
		"mysql": {
			"driver":   "mysql",
			"servname": os.Getenv("MYSQL_SERVNAME"),
			"host":     os.Getenv("MYSQL_HOST"),
			"port":     os.Getenv("MYSQL_PORT"),
			"username": os.Getenv("MYSQL_USERNAME"),
			"password": os.Getenv("MYSQL_PASSWORD"),
			"dbname":   os.Getenv("MYSQL_DATABASE"),
		},
		"solr": {
			"scheme": os.Getenv("SOLR_SCHEME"),
			"host":   os.Getenv("SOLR_HOST"),
			"port":   os.Getenv("SOLR_PORT"),
			"core":   os.Getenv("SOLR_CORE"),
		},
		"redis": {
			"host":     os.Getenv("REDIS_HOST"),
			"port":     os.Getenv("REDIS_PORT"),
			"password": os.Getenv("REDIS_PASSWORD"),
		},
	}
}
