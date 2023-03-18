package config

import (
	"robot/helpers"
)

type Database struct{}

func (cfg *Database) Load() map[string]map[string]string {
	env := helpers.Env{}
	env.LoadEnv()

	return map[string]map[string]string{
		"pgsql": {
			"driver":   "pgsql",
			"servname": env.Env("PGSQL_SERVNAME"),
			"host":     env.Env("PGSQL_HOST"),
			"port":     env.Env("PGSQL_PORT"),
			"username": env.Env("PGSQL_USERNAME"),
			"password": env.Env("PGSQL_PASSWORD"),
			"dbname":   env.Env("PGSQL_DATABASE"),
		},
		"mysql": {
			"driver":   "mysql",
			"servname": env.Env("MYSQL_SERVNAME"),
			"host":     env.Env("MYSQL_HOST"),
			"port":     env.Env("MYSQL_PORT"),
			"username": env.Env("MYSQL_USERNAME"),
			"password": env.Env("MYSQL_PASSWORD"),
			"dbname":   env.Env("MYSQL_DATABASE"),
		},
		"solr": {
			"scheme": env.Env("SOLR_SCHEME"),
			"host":   env.Env("SOLR_HOST"),
			"port":   env.Env("SOLR_PORT"),
			"core":   env.Env("SOLR_CORE"),
		},
		"redis": {
			"host":     env.Env("REDIS_HOST"),
			"port":     env.Env("REDIS_PORT"),
			"password": env.Env("REDIS_PASSWORD"),
		},
	}
}
