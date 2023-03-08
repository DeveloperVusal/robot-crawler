package database

import (
	"context"
	"database/sql"
	"net/url"

	"robot/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgconn/stmtcache"
	"github.com/jackc/pgx/v4"
)

type Database struct{}

// Метод подключения к БД MySQL
func (db *Database) ConnMySQL(serverName string) (context.Context, *sql.DB, error) {
	cfg := config.ConfigDatabaseLoad()

	var driverName string = ""
	var dataSourceName string = ""

	for key := range cfg {
		if key == "mysql" {
			if cfg[key]["servname"] == serverName {
				driverName = cfg[key]["driver"]
				dataSourceName = cfg[key]["username"] + ":" + cfg[key]["password"] + "@tcp(" + cfg[key]["host"] + ":" + cfg[key]["port"] + ")/" + cfg[key]["dbname"]
			}
		}
	}

	_db, err := db.DriverMySQL(&driverName, &dataSourceName)

	_db.SetMaxOpenConns(30)

	return context.Background(), _db, err
}

// Метод подключения к БД PgSQL
func (db *Database) ConnPgSQL(serverName string) (context.Context, *pgx.Conn, error) {
	cfg := config.ConfigDatabaseLoad()

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

// Вызов драйвера mysql
func (db *Database) DriverMySQL(driverName *string, dataSourceame *string) (*sql.DB, error) {
	return sql.Open(*driverName, *dataSourceame)
}

// Вызов драйвера pgsql
func (db *Database) DriverPgSQL(config *pgx.ConnConfig) (*pgx.Conn, error) {
	return pgx.ConnectConfig(context.Background(), config)
}
