package app

import (
	"context"
	"regexp"
	"robot/database"

	"github.com/redis/go-redis/v9"
)

type SearchDB struct {
	Redis *redis.Client
	Ctx   context.Context
}

// Метод проверяет имется ли страница в базе для поиска
func (srdb *SearchDB) IsWebPageBase(url *string) (uint64, bool) {
	db := database.PgSQL{}
	ctx, dbn, err := db.ConnPgSQL("rw_pgsql_search")

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	defer dbn.Close(ctx)

	var row_id uint64

	dbn.QueryRow(ctx, `SELECT id FROM web_pages WHERE page_url=$1`, *url).Scan(&row_id)

	if row_id > 0 {
		return row_id, true
	} else {
		return 0, false
	}
}

// Метод добавляет страницу в базу для поиска
func (srdb *SearchDB) AddWebPageBase(domain_id *uint64, resp *PageReqData) (uint64, string) {
	if resp.StatusCode == 200 {
		matched, _ := regexp.MatchString(`^(text\/html|text\/plain)`, resp.Header.Get("Content-Type"))

		if matched {
			rbtxt := &Robotstxt{
				Redis:       srdb.Redis,
				Ctx:         srdb.Ctx,
				Domain_id:   *domain_id,
				IndexPgFind: []string{"*", "/", "?"},
				IndexpgRepl: []string{".*", "\\/", "\\?"},
			}
			isValid, newUrl := rbtxt.UrlHandle(&resp.Url)

			if len(newUrl) > 4 && newUrl != resp.Url {
				resp.Url = newUrl
			}

			if isValid {
				db := database.PgSQL{}
				ctx, dbn, err := db.ConnPgSQL("rw_pgsql_search")

				if err != nil {
					log := &Logs{}
					log.LogWrite(err)
				}

				defer dbn.Close(ctx)

				var insertId uint64
				var sql string = `INSERT INTO web_pages (
									domain_id,
									page_url,
									http_code,
									created_at
								) 
								VALUES(
									$1,
									$2,
									$3,
									NOW()::timestamp
								) RETURNING id
								`

				err = dbn.QueryRow(ctx, sql, *domain_id, resp.Url, resp.StatusCode).Scan(&insertId)

				if err != nil {
					log := &Logs{}
					log.LogWrite(err)
				}

				return insertId, resp.Url
			}
		}
	}

	return 0, ""
}
