package database

import (
	"context"
	"database/sql"

	config "robot/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v4"
)

type Database struct{}

// Метод подключения к БД MySQL
func (db *Database) ConnMySQL(serverName string) (*sql.DB, error) {
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

	return db.DriverMySQL(&driverName, &dataSourceName)
}

// Метод подключения к БД PgSQL
func (db *Database) ConnPgSQL(serverName string) (context.Context, *pgx.Conn, error) {
	cfg := config.ConfigDatabaseLoad()

	var dataSourceName string = ""

	for key := range cfg {
		if key == "pgsql" {
			if cfg[key]["servname"] == serverName {
				dataSourceName = "postgres://" + cfg[key]["username"] + ":" + cfg[key]["password"] + "@" + cfg[key]["host"] + ":" + cfg[key]["port"] + "/" + cfg[key]["dbname"]
			}
		}
	}

	_db, _err := db.DriverPgSQL(&dataSourceName)

	return context.Background(), _db, _err

}

// Вызов драйвера mysql
func (db *Database) DriverMySQL(driverName *string, dataSourceame *string) (*sql.DB, error) {
	return sql.Open(*driverName, *dataSourceame)
}

// Вызов драйвера pgsql
func (db *Database) DriverPgSQL(dataSourceame *string) (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), *dataSourceame)
}
