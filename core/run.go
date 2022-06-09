package core

import (
	"fmt"
	"log"
	"robot/app"
)

type Robotgo struct{}

func (rg *Robotgo) IsQueue() {
	// Подключаемся к БД
	db := Database{}
	dbn, err := db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err.Error())
	}

	defer dbn.Close()

	// Проверяет пуста ли очередь
	rows, err := dbn.Query("CALL queue_handle();")

	if err != nil {
		log.Fatalln(err.Error())
	}

	var id uint64
	var domain_id uint64
	var url string
	var domain_full string

	for rows.Next() {
		rows.Scan(&id, &url, &domain_id, &domain_full)
	}

	// Если есть, что запустить в обработку
	if id > 0 {
		rg.AddQueue(id)

		if rg.IsWebPageBase(&url) {
			indx := &app.Indexing{}
			go indx.Run(id, url, domain_id, domain_full)
		}
	}
}

func (rg *Robotgo) AddQueue(id uint64) {
	// Подключаемся к БД
	db := Database{}
	dbn, err := db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err.Error())
	}

	defer dbn.Close()

	// Обновляем статус в очереди
	res, err := dbn.Exec("UPDATE `queue_pages` SET status = 0, handler = 1 WHERE id = ?", id)

	if err != nil {
		log.Fatalln(err.Error())
	}

	if res != nil {
		fmt.Println("id = ", id, " queue added")
	}
}

func (rg *Robotgo) IsWebPageBase(url *string) bool {
	db := Database{}
	ctx, dbn, err := db.ConnPgSQL("pgsql")

	if err != nil {
		log.Fatalln(err.Error())
	}

	defer dbn.Close(ctx)

	var row_id uint64
	var row_domain_id uint64

	fmt.Println("url", *url)

	dbn.QueryRow(ctx, "SELECT id, domain_id FROM web_pages WHERE url=$1", *url).Scan(&row_id, &row_domain_id)

	if row_id > 0 {
		fmt.Println("web_pages", row_id, row_domain_id)

		return true
	} else {
		fmt.Println("empty is web_pages")

		return false
	}
}
