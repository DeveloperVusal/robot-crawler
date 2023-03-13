package app

import (
	"context"
	neturl "net/url"
	"strconv"
	"time"

	dbpkg "robot/database"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	Redis *redis.Client
	Ctx   context.Context
}

// Метод проверяет и запускает индексацию страниц в очереди
func (q *Queue) RunQueue() {
	// Получаем наличие очереди
	count, _ := q.Redis.LLen(q.Ctx, "queue").Result()

	// Если очередь не пуста
	if count > 0 {
		// Забираем url из конца списка/очереди
		_url, err := q.Redis.RPop(q.Ctx, "queue").Result()

		if err != nil {
			log := &Logs{}
			log.LogWrite(err)
		} else {
			domain_id, _ := q.Redis.HGet(q.Ctx, _url, "domain_id").Result()
			domain_full, _ := q.Redis.HGet(q.Ctx, _url, "domain_full").Result()

			d_id, _ := strconv.ParseUint(domain_id, 10, 64)

			q.HandleQueue(_url, d_id, domain_full)
		}
	}
}

// Обработка страницы из очереди
func (q *Queue) HandleQueue(url string, domain_id uint64, domain_full string) {
	isContinue := true
	_, errp := neturl.Parse(url)

	// Если в очередь попал не корректный url
	if errp != nil {
		q.SetHash(url, 503, 0, false)
		isContinue = false

		log := &Logs{}
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

				// Если такой url есть в базе
				if isPage {
					// Запускаем индексацию в потоке
					go indx.Run(idPage, resp.Url)
				} else { // Иначе
					if len(resp.Url) > 4 {
						// Добавляем url в базу
						lastInsertId, origUrl := srchdb.AddWebPageBase(&domain_id, &resp)

						// Если url добавлен
						if lastInsertId > 0 {
							// Запускаем индексацию в потоке
							go indx.Run(lastInsertId, origUrl)
						} else {
							q.SetHash(url, 500, 2, false) // Пропускаем url
						}
					} else {
						q.SetHash(url, 501, 2, false) // Пропускаем url
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

// Пропустить зависшую страницу в очереди
func (q *Queue) ContinueQueue() {

}

// Очищаем старые обработанные url из очереди
func (q *Queue) ClearQueue() {

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
	db := dbpkg.Database{}
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
