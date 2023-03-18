package database

import (
	"context"
	"net/url"

	"robot/config"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgconn/stmtcache"
	"github.com/jackc/pgx/v4"
)

type PgSQL struct{}

// Метод подключения к БД PgSQL
func (db *PgSQL) ConnPgSQL(serverName string) (context.Context, *pgx.Conn, error) {
	loadCfg := &config.Database{}
	cfg := loadCfg.Load()

	var dataSourceName string = ""

	for key := range cfg {
		if key == "pgsql" {
			if cfg[key]["servname"] == serverName {
				cfg[key]["password"] = url.QueryEscape(cfg[key]["password"])

				dataSourceName = "postgres://" + cfg[key]["username"] + ":" + cfg[key]["password"] + "@" + cfg[key]["host"] + ":" + cfg[key]["port"] + "/" + cfg[key]["dbname"]
			}
		}
	}

	config, _ := pgx.ParseConfig(dataSourceName)
	config.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
		return stmtcache.New(conn, stmtcache.ModeDescribe, 1024)
	}

	_db, _err := db.DriverPgSQL(config)

	return context.Background(), _db, _err
}

// Вызов драйвера pgsql
func (db *PgSQL) DriverPgSQL(config *pgx.ConnConfig) (*pgx.Conn, error) {
	return pgx.ConnectConfig(context.Background(), config)
}
