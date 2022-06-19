package app

import (
	"log"
	"regexp"
	dbpkg "robot/database"
)

type SearchDB struct{}

// Метод проверяет имеется ли страница в базе для поиска
func (srdb *SearchDB) IsWebPageBase(url *string) (uint64, bool) {
	db := dbpkg.Database{}
	ctx, dbn, err := db.ConnPgSQL("pgsql")

	if err != nil {
		log.Fatalln(err)
	}

	defer dbn.Close(ctx)

	var row_id uint64

	dbn.QueryRow(ctx, "SELECT id FROM web_pages WHERE page_url=$1", *url).Scan(&row_id)

	if row_id > 0 {
		return row_id, true
	} else {
		return 0, false
	}
}

// Метод добавляет страницу в базу для поиска
func (srdb *SearchDB) AddWebPageBase(domain_id *uint64, url *string, resp PageReqData) (uint64, string) {
	if resp.StatusCode == 200 {
		matched, _ := regexp.MatchString(`^(text\/html|text\/plain)`, resp.Header.Get("Content-Type"))

		if matched {
			db := dbpkg.Database{}
			ctx, dbn, err := db.ConnPgSQL("pgsql")

			if err != nil {
				log.Fatalln(err)
			}

			defer dbn.Close(ctx)

			var insertId uint64
			var sql string = `INSERT INTO web_pages (
								domain_id,
								page_url,
								meta_title,
								meta_description,
								meta_keywords,
								page_text,
								http_code,
								created_at
							) 
							VALUES(
								$1,
								$2,
								'',
								'',
								'',
								'',
								$3,
								NOW()::timestamp
							) RETURNING id
							`

			err = dbn.QueryRow(ctx, sql, *domain_id, resp.Url, resp.StatusCode).Scan(&insertId)

			if err != nil {
				log.Fatalln(err)
			}

			return insertId, resp.Url
		}
	}

	return 0, ""
}
