package app

import (
	"context"
	"errors"
	neturl "net/url"
	"strconv"
	"time"

	"robot/database"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	Redis      *redis.Client
	Ctx        context.Context
	MaxThreads uint64
}

// Метод проверяет и запускает индексацию страниц в очереди
func (q *Queue) RunQueue() {
	// Получаем наличие очереди
	count, _ := q.Redis.LLen(q.Ctx, "queue").Result()

	// Если очередь не пуста
	if count > 0 {
		// Проверяем доступность места на обработку
		wcount, _ := q.Redis.LLen(q.Ctx, "worker").Result()

		if uint64(wcount) < q.MaxThreads {
			// Забираем url из конца списка/очереди
			_url, err := q.Redis.RPop(q.Ctx, "queue").Result()

			if err != nil {
				log := &Logs{}
				log.LogWrite(err)
			} else {
				is_add := q.AddWorker(_url)

				// Если url добавлена в обработку
				if is_add {
					domain_id, _ := q.Redis.HGet(q.Ctx, _url, "domain_id").Result()
					domain_full, _ := q.Redis.HGet(q.Ctx, _url, "domain_full").Result()

					d_id, _ := strconv.ParseUint(domain_id, 10, 64)

					defer q.HandleQueue(_url, d_id, domain_full)
				}
			}
		}
	}
}

// Обработка страницы из очереди
func (q *Queue) HandleQueue(url string, domain_id uint64, domain_full string) {
	log := &Logs{}

	isContinue := true
	_, errp := neturl.Parse(url)

	// Если в очередь попал не корректный url
	if errp != nil {
		q.SetHash(url, 500, 0, false)
		q.RemoveWorker(url)
		isContinue = false

		log.LogWrite(errors.New(`err step: 500, 0 -- ` + url))
		log.LogWrite(errp)
	}

	// Если url корректный продолжаем
	if isContinue {
		req := &Request{
			Redis: q.Redis,
			Ctx:   q.Ctx,
		}

		// Если есть доступный лимит для обработки
		if req.IsRequestLimit(&domain_full) {
			q.SetHash(url, 0, 1, true)

			resp, isDisable := req.GetPageData(&url) // Делаем запрос и получаем данные url

			if isDisable {
				indx := &Indexing{
					Redis:       q.Redis,
					Ctx:         q.Ctx,
					QueueKey:    url,
					Domain_id:   domain_id,
					Domain_full: domain_full,
					Resp:        &resp,
				}

				// Получаем данные об url
				srchdb := &SearchDB{
					Redis: q.Redis,
					Ctx:   q.Ctx,
				}
				idPage, isPage := srchdb.IsWebPageBase(&resp.Url)

				// log.LogWrite(errors.New(`idPage:` + strconv.FormatUint(idPage, 10) + `,  isPage:` + strconv.FormatBool(isPage)))

				// Если такой url есть в базе
				if isPage {
					// Запускаем индексацию в потоке
					go indx.Run(idPage, resp.Url)

					return
				} else { // Иначе
					if len(resp.Url) > 4 {
						// Добавляем url в базу
						lastInsertId, origUrl := srchdb.AddWebPageBase(&domain_id, &resp)

						// Если url добавлен
						if lastInsertId > 0 {
							// Запускаем индексацию в потоке
							go indx.Run(lastInsertId, origUrl)

							return
						} else {
							q.SetHash(url, 501, 0, false) // Пропускаем url
							log.LogWrite(errors.New(`err step: 501, 0 -- ` + resp.Url))
							q.RemoveWorker(url)
						}
					} else {
						q.SetHash(url, 502, 0, false) // Пропускаем url
						log.LogWrite(errors.New(`err step: 501, 0 -- ` + resp.Url))
						q.RemoveWorker(url)
					}
				}
			}
		}
	}
}

// Метод устанвливает статус обработки страницы
func (q *Queue) SetHash(url string, _status int, _handler int, isRun bool) {
	q.Redis.HSet(q.Ctx, url, "status", _status)
	q.Redis.HSet(q.Ctx, url, "handler", _handler)

	if isRun {
		q.Redis.HSet(q.Ctx, url, "launched_at", time.Now().Unix())
	}
}

// Метод добавляет страницу в обработку
func (q *Queue) AddWorker(_url string) bool {
	// Проверяем есть ли урл в обработке
	indxPos, _ := q.Redis.LPos(q.Ctx, "worker", _url, redis.LPosArgs{Rank: 1}).Result()

	// Если нет, то добавляем
	if indxPos <= 0 {
		status, _ := q.Redis.HGet(q.Ctx, _url, "status").Result()
		handler, _ := q.Redis.HGet(q.Ctx, _url, "handler").Result()

		d_status, _ := strconv.ParseUint(status, 10, 8)
		d_handler, _ := strconv.ParseUint(handler, 10, 8)

		// Если url доступен для обработки
		if d_status == 0 && d_handler == 0 {
			_int, _ := q.Redis.LPush(q.Ctx, "worker", _url).Result()

			if _int > 0 {
				return true
			}
		}
	}

	return false
}

// Метод удаляет страницу из обработки
func (q *Queue) RemoveWorker(_url string) {
	q.Redis.LRem(q.Ctx, "worker", 0, _url)
}

// Пропустить зависшие воркеры в обработке
func (q *Queue) ContinueWorkers() {
	keys, _ := q.Redis.LRange(q.Ctx, "worker", 0, -1).Result()

	for _, _url := range keys {
		_urlParse, _ := neturl.Parse(_url)

		if len(_urlParse.Host) > 0 {
			if len(_urlParse.Host) > 0 {
				_launched_at, _ := q.Redis.HGet(q.Ctx, _url, "launched_at").Result()

				if len(_launched_at) > 0 {
					launched_at, _ := strconv.ParseInt(_launched_at, 10, 64)

					if (time.Now().Unix() - launched_at) > 180 {
						// q.Redis.Del(q.Ctx, _url)
						q.RemoveWorker(_url)
					}
				}
			}
		}
	}
}

// Очищаем обработанные url более чем 1 сутки
func (q *Queue) ClearQueue() {
	var checkCursor int64 = 0
	var cursor uint64 = 0

	for checkCursor > -1 {
		keys, _cursor, _ := q.Redis.Scan(q.Ctx, cursor, "", 20).Result()

		for _, _url := range keys {
			_urlParse, _ := neturl.Parse(_url)

			if len(_urlParse.Host) > 0 {
				_launched_at, _ := q.Redis.HGet(q.Ctx, _url, "launched_at").Result()

				if len(_launched_at) > 0 {
					launched_at, _ := strconv.ParseInt(_launched_at, 10, 64)

					if (time.Now().Unix() - launched_at) > 86400 {
						q.Redis.Del(q.Ctx, _url)
					}
				}
			}
		}

		if _cursor > 0 {
			cursor = _cursor
		} else {
			checkCursor = -1
		}
	}
}

// Метод создает хэш ключ
func (q *Queue) CreateHash(url string, domain_id uint64, domain_full string) {
	q.Redis.HSetNX(q.Ctx, url, "domain_id", domain_id)
	q.Redis.HSetNX(q.Ctx, url, "domain_full", domain_full)
	q.Redis.HSetNX(q.Ctx, url, "status", 0)
	q.Redis.HSetNX(q.Ctx, url, "handler", 0)
	q.Redis.HSetNX(q.Ctx, url, "launched_at", "")
	q.Redis.HSetNX(q.Ctx, url, "created_at", time.Now().Unix())
}

// Метод добавляет страницу в хэш и в очередь на обработку
func (q *Queue) AddUrlQueue(url string, domain_id uint64, domain_full string) {
	// Создаем хэш, если его нет
	q.CreateHash(url, domain_id, domain_full)

	// Проверяем есть ли урл в очереди
	indxPos, _ := q.Redis.LPos(q.Ctx, "queue", url, redis.LPosArgs{Rank: 1}).Result()

	// Если нет, то добавляем урл в очередь
	if indxPos <= 0 {
		q.Redis.LPush(q.Ctx, "queue", url).Result()

		// if st > 0 {
		// 	// fmt.Println("add queue =>", url)
		// }
	}
}

// Метод проверяет и добавляет сайты в очередь
func (q *Queue) SitesQueue() {
	db := database.PgSQL{}
	ctx, dbn, err := db.ConnPgSQL("rw_pgsql_search")

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	defer dbn.Close(ctx)

	// Получаем все сайты из базы Postgres
	rows, err := dbn.Query(ctx, `SELECT id, domain_full FROM web_sites`)

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	for rows.Next() {
		var id uint64
		var domain_full string

		rows.Scan(&id, &domain_full)

		// Пробуем добавить сайт в очередь
		q.AddUrlQueue(domain_full, id, domain_full)
	}
}
