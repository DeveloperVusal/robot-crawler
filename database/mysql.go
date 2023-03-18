package database

import (
	"context"
	"database/sql"
	"robot/config"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct{}

// Метод подключения к БД MySQL
func (db *MySQL) ConnMySQL(serverName string) (context.Context, *sql.DB, error) {
	loadCfg := config.Database{}
	cfg := loadCfg.Load()

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

// Вызов драйвера mysql
func (db *MySQL) DriverMySQL(driverName *string, dataSourceame *string) (*sql.DB, error) {
	return sql.Open(*driverName, *dataSourceame)
}
