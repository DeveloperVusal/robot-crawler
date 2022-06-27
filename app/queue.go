package app

import (
	"log"
	dbpkg "robot/database"
)

type Queue struct{}

// Метод проверяет и запускает индексацию страниц в очереди
func (q *Queue) IsQueue() {
	// Подключаемся к БД
	db := dbpkg.Database{}
	dbn, err := db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err)
	}

	defer dbn.Close()

	// Проверяет пуста ли очередь
	rows, err := dbn.Query("CALL queue_handle();")

	if err != nil {
		log.Fatalln(err)
	}

	var id uint64
	var domain_id uint64
	var url string
	var domain_full string

	for rows.Next() {
		rows.Scan(&id, &url, &domain_id, &domain_full)
	}

	req := &Request{}

	// Если есть, что запустить в обработку
	if id > 0 && req.IsRequestLimit(&url) {
		q.SetQueue(id, 0, 1) // Включаем обработку url

		req := &Request{}
		resp := req.GetPageData(&url) // Делаем запрос и получаем данные url
		indx := &Indexing{
			QueueId:     id,
			Domain_id:   domain_id,
			Domain_full: domain_full,
			Resp:        &resp,
		}

		srchdb := &SearchDB{}
		idPage, isPage := srchdb.IsWebPageBase(&resp.Url)

		// Если такой url есть в базе
		if isPage {
			// Запускаем индексацию
			go indx.Run(idPage, resp.Url)
		} else { // Иначе
			// Добавляем url в базу
			lastInsertId, origUrl := srchdb.AddWebPageBase(&domain_id, resp)

			// Если url добавлен
			if lastInsertId > 0 {
				// Запускаем индексацию
				go indx.Run(lastInsertId, origUrl)
			} else {
				q.SetQueue(id, 500, 2) // Пропускаем url
			}
		}
	}
}

// Метод статус страницы в очереди
func (q *Queue) SetQueue(id uint64, _status int, _handler int) {
	// Подключаемся к БД
	db := dbpkg.Database{}
	dbn, err := db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err)
	}

	defer dbn.Close()

	// Обновляем статус в очереди
	_, err2 := dbn.Exec("UPDATE `queue_pages` SET status = ?, handler = ?, thread_time = NOW() WHERE id = ?", _status, _handler, id)

	if err2 != nil {
		log.Fatalln(err2)
	}
}

// Пропустить зависшую страницу в очереди
func (q *Queue) ContinueQueue() {
	// Подключаемся к БД
	db := dbpkg.Database{}
	dbn, err := db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err)
	}

	defer dbn.Close()

	// Проверяем имеются ли в очереди страницы зависщие на более 5 минут
	var id uint64
	var sql string = `
		SELECT
			id
		FROM queue_pages
		WHERE
			status = 0 AND
			handler = 1 AND
			UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(thread_time) >= 300`

	rows, err2 := dbn.Query(sql)

	if err2 != nil {
		log.Fatalln(err2)
	}

	for rows.Next() {
		rows.Scan(&id)
	}

	if id > 0 {
		q.SetQueue(id, 700, 0)
	}
}

// Метод добавляет страницу в очередь на обработку
func (q *Queue) AddUrlQueue(url string, domain_id uint64, domain_full string) {
	// Подключаемся к БД
	db := dbpkg.Database{}
	dbn, err := db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err)
	}

	defer dbn.Close()

	// Проверяем не добавлена ли в очередь
	var selID int
	rows, err := dbn.Query("SELECT `id` FROM `queue_pages` WHERE `url` = ?", url)

	if err != nil {
		log.Fatalln(err)
	}

	for rows.Next() {
		err := rows.Scan(&selID)

		if err != nil {
			log.Fatalln(err)
		}
	}

	// Если такого url нет в очереди
	if selID <= 0 {
		// Добавляем url в очередь
		_, err := dbn.Exec("INSERT INTO queue_pages (domain_id, domain_full, url) VALUES (?, ?, ?)", domain_id, domain_full, url)

		if err != nil {
			log.Fatalln(err)
		}
	}
}
