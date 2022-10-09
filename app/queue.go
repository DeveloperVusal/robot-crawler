package app

import (
	"context"
	"database/sql"
	"log"
	neturl "net/url"
	"time"
)

type Queue struct {
	DBLink *sql.DB
	Ctx    context.Context
}

// Метод проверяет и запускает индексацию страниц в очереди
func (q *Queue) RunQueue() {
	dbn := q.DBLink
	ctx, cancelfunc := context.WithTimeout(q.Ctx, 180*time.Second)

	defer cancelfunc()

	// Проверяет пуста ли очередь
	rows, err := dbn.QueryContext(ctx, "CALL queue_handle();")

	if err != nil {
		log.Fatalln(err)
	}

	var id uint64
	var domain_id uint64
	var url string
	var domain_full string

	for rows.Next() {
		rows.Scan(&id, &url, &domain_id, &domain_full)

		if id > 0 {
			q.HandleQueue(&id, &url, &domain_id, &domain_full)
		}
	}
}

// Обработка страницы из очереди
func (q *Queue) HandleQueue(id *uint64, url *string, domain_id *uint64, domain_full *string) {
	isContinue := true
	_, errp := neturl.Parse(*url)

	// Если в очередь попала не корректный url
	if errp != nil {
		q.SetQueue(*id, 503, 0)
		isContinue = false
	}

	if isContinue {
		req := &Request{
			DBLink: q.DBLink,
			Ctx:    q.Ctx,
		}

		// Если есть, что запустить в обработку
		if *id > 0 && req.IsRequestLimit(url) {
			q.SetQueue(*id, 0, 1) // Включаем обработку url

			resp := req.GetPageData(url) // Делаем запрос и получаем данные url
			indx := &Indexing{
				DBLink:      q.DBLink,
				Ctx:         q.Ctx,
				QueueId:     *id,
				Domain_id:   *domain_id,
				Domain_full: *domain_full,
				Resp:        &resp,
			}

			srchdb := &SearchDB{
				DBLink: q.DBLink,
				Ctx:    q.Ctx,
			}
			idPage, isPage := srchdb.IsWebPageBase(&resp.Url)

			// Если такой url есть в базе
			if isPage {
				// Запускаем индексацию
				go indx.Run(idPage, resp.Url)
			} else { // Иначе
				if len(resp.Url) > 4 {
					// Добавляем url в базу
					lastInsertId, origUrl := srchdb.AddWebPageBase(domain_id, &resp)

					// Если url добавлен
					if lastInsertId > 0 {
						// Запускаем индексацию
						go indx.Run(lastInsertId, origUrl)
					} else {
						q.SetQueue(*id, 500, 2) // Пропускаем url
					}
				} else {
					q.SetQueue(*id, 501, 2) // Пропускаем url
				}
			}
		}
	}
}

// Метод устанвливает статус страницы в очереди
func (q *Queue) SetQueue(id uint64, _status int, _handler int) {
	dbn := q.DBLink
	ctx, cancelfunc := context.WithTimeout(q.Ctx, 180*time.Second)

	defer cancelfunc()

	// Обновляем статус в очереди
	_, err2 := dbn.ExecContext(ctx, "UPDATE `queue_pages` SET status = ?, handler = ?, thread_time = NOW() WHERE id = ?", _status, _handler, id)

	if err2 != nil {
		log.Fatalln(err2)
	}
}

// Пропустить зависшую страницу в очереди
func (q *Queue) ContinueQueue() {
	dbn := q.DBLink
	ctx, cancelfunc := context.WithTimeout(q.Ctx, 180*time.Second)

	defer cancelfunc()

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

	rows, err2 := dbn.QueryContext(ctx, sql)

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

// Очищаем старые обработанные url из очереди
func (q *Queue) ClearQueue() {
	dbn := q.DBLink
	ctx, cancelfunc := context.WithTimeout(q.Ctx, 180*time.Second)

	defer cancelfunc()

	// Очищаем обработанные страницы в очереди более 3-х дней
	_, err2 := dbn.ExecContext(ctx, `
		DELETE
		FROM queue_pages
		WHERE
			status != 0 AND
			handler != 0 AND
			(UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(thread_time)) >= 259200
	`)

	if err2 != nil {
		log.Fatalln(err2)
	}
}

// Метод добавляет страницу в очередь на обработку
func (q *Queue) AddUrlQueue(url string, domain_id uint64, domain_full string) {
	dbn := q.DBLink
	ctx, cancelfunc := context.WithTimeout(q.Ctx, 180*time.Second)

	defer cancelfunc()

	// Проверяем не добавлена ли страница в очередь
	var selID int
	rows, err := dbn.QueryContext(ctx, "SELECT `id` FROM `queue_pages` WHERE `url` = ?", url)

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
		_, err := dbn.ExecContext(ctx, "INSERT INTO queue_pages (domain_id, domain_full, url) VALUES (?, ?, ?)", domain_id, domain_full, url)

		if err != nil {
			log.Fatalln(err)
		}
	}
}
