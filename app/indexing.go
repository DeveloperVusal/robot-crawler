package app

import (
	"context"
	"database/sql"

	// "fmt"
	neturl "net/url"
	"regexp"
	"strings"

	dbpkg "robot/database"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/microcosm-cc/bluemonday"
)

type Indexing struct {
	DBLink          *sql.DB
	Ctx             context.Context
	QueueId         uint64
	Domain_id       uint64
	Domain_full     string
	Resp            *PageReqData
	PageCurrentData map[string]string
}

// Метод Запускает индексирование страницы
func (indx *Indexing) Run(id uint64, url string) {
	pageHead := map[string]string{}
	pageBody := map[string][]string{}
	pageContent := []string{}
	pageLinks := []string{}

	appqueue := &Queue{
		DBLink: indx.DBLink,
		Ctx:    indx.Ctx,
	}

	if indx.Resp.StatusCode == 200 {
		matched, _ := regexp.MatchString(`^(text\/html|text\/plain)`, indx.Resp.Header.Get("Content-Type"))

		if matched {
			rbtxt := &Robotstxt{
				DBLink:      indx.DBLink,
				Ctx:         indx.Ctx,
				Domain_id:   indx.Domain_id,
				IndexPgFind: []string{"*", "/", "?"},
				IndexpgRepl: []string{".*", "\\/", "\\?"},
			}
			isValid, newUrl := rbtxt.UrlHandle(&indx.Resp.Url)

			if len(newUrl) > 4 && newUrl != indx.Resp.Url {
				indx.Resp.Url = newUrl
			}

			// fmt.Println("link", indx.Resp.Url)

			if isValid {
				// Получаем Dom документ страницы
				doc, err := goquery.NewDocumentFromReader(indx.Resp.Body)

				if err != nil {
					log := &Logs{}
					log.LogWrite(err)
				}

				// Парсинг полнного домена
				uParseDom, _ := neturl.Parse(indx.Domain_full)

				// Функции для фильтра
				filterFunc := Filter{}

				meta_robots, _ := doc.Find("meta[name=robots]").Attr("content")
				isNoindex := true
				isNofollow := true

				if len(meta_robots) > 3 {
					if strings.Contains(meta_robots, "noindex") || strings.Contains(meta_robots, "nofollow") {
						isNoindex, _ = regexp.MatchString(`noindex`, meta_robots)
						isNofollow, _ = regexp.MatchString(`nofollow`, meta_robots)
					}
				}

				// Если индексация доступна
				if isNoindex {
					// Мета теги и Title
					pageHead["title"] = doc.Find("title").Text()

					meta_descr, _ := doc.Find("meta[name=description]").Attr("content")
					pageHead["description"] = meta_descr

					meta_keywords, _ := doc.Find("meta[name=keywords]").Attr("content")
					pageHead["keywords"] = meta_keywords

					// Заголовок h1
					doc.Find("body").Each(func(i int, s *goquery.Selection) {
						pageBody["h1"] = append(pageBody["h1"], strings.Trim(s.Find("h1").Text(), " \t\r\n"))

						// Если переход по ссылкам доступен
						if isNofollow {
							s.Find("a").Each(func(ix int, sx *goquery.Selection) {
								attrHref, _ := sx.Attr("href")
								attrRel, _ := sx.Attr("rel")
								uParseDom2, _ := neturl.Parse(attrHref)

								if len(uParseDom2.Host) <= 0 {
									if len(uParseDom2.Path) > 0 {
										if uParseDom2.Path[0:1] != "/" {
											uParseDom2.Path = "/" + uParseDom2.Path
										}
									}

									attrHref = uParseDom.Scheme + `://` + uParseDom.Host + uParseDom2.Path
								}

								if filterFunc.IsValidLink(attrHref, indx.Domain_full) && attrRel != "nofollow" {
									pageLinks = append(pageLinks, attrHref)
								}
							})
						}
					})

					if len(pageHead["title"]) > 2710 {
						pageHead["title"] = filterFunc.Substr(pageHead["title"], 0, 2710)
					}

					if len(pageBody["h1"][0]) > 2710 {
						pageBody["h1"][0] = filterFunc.Substr(pageBody["h1"][0], 0, 2710)
					}

					// Если переход по ссылкам доступен
					if isNofollow {
						// Получаем уникальные ссылки
						pageLinks = filterFunc.SliceStrUnique(pageLinks)
					}

					// Содержание страницы
					indx.GetContent(doc, &pageContent)

					pageContent = filterFunc.SliceStrUnique(pageContent)
					pageBody["content"] = append(pageBody["content"], strings.Trim(strings.Join(pageContent[:], " "), " \t\r\n"))

					// Помечаем страницу посещенной
					isPageCrawl := indx.PageCrawl(&id)

					// Если есть что изменить,
					// то обновляем данные страницы
					if pageHead["title"] != indx.PageCurrentData["title"] ||
						pageHead["description"] != indx.PageCurrentData["meta_description"] ||
						pageHead["keywords"] != indx.PageCurrentData["meta_keywords"] ||
						pageBody["h1"][0] != indx.PageCurrentData["page_h1"] ||
						pageBody["content"][0] != indx.PageCurrentData["page_text"] {
						indx.PageUpdate(&id, map[string]string{
							"url":              url,
							"meta_title":       pageHead["title"],
							"meta_description": pageHead["description"],
							"meta_keywords":    pageHead["keywords"],
							"page_h1":          pageBody["h1"][0],
							"page_text":        pageBody["content"][0],
						})
					}

					// Если страница была посещена
					if isPageCrawl {
						// Добавляем ссылки в очередь
						for k := range pageLinks {
							appqueue.AddUrlQueue(pageLinks[k], indx.Domain_id, indx.Domain_full)
						}

						// Указываем заврешении индексации
						appqueue.SetQueue(indx.QueueId, 1, 2)
					} else {
						// Указываем в очереди о недоступности индексирования
						// страница не была обновлена в базе
						appqueue.SetQueue(indx.QueueId, 600, 500)
					}
				} else {
					// Указываем заврешении индексации
					appqueue.SetQueue(indx.QueueId, 1, 2)
				}
			}
		} else {
			// Удаляем страницу из индексации
			indx.PageDeleteIndex(&id, &indx.Resp.StatusCode)

			// Указываем в очереди о недоступности индексирования
			// Страница не является TEXT или HTML
			appqueue.SetQueue(indx.QueueId, 500, 500)
		}
	} else {
		// Отключаем страницу из индексации
		indx.PageDisableIndex(&id)

		// Указываем в очереди о недоступности индексирования
		// Страница не доступна
		appqueue.SetQueue(indx.QueueId, indx.Resp.StatusCode, 500)
	}

	// fmt.Println("")
}

// Метод получает содержимое/контент на странице
func (indx *Indexing) GetContent(doc *goquery.Document, output *[]string) {
	filterRemove := []string{
		"script",
		"style",
		"iframe",
	}
	filterSel := []string{
		"body [class*=\"content\"]",
		"body [id*=\"content\"]",
		"body p",
		"body h2", "body h3", "body h4",
		"body h5", "body h6",
	}
	filterFunc := Filter{}
	striptags := bluemonday.StripTagsPolicy()

	striptags.AddSpaceWhenStrippingTag(true)

	// Удаляем ненужный контент
	for key := range filterRemove {
		doc.Find(filterRemove[key]).Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})
	}

	// Извлекаем необхоимый контент
	for key := range filterSel {
		doc.Find(filterSel[key]).Each(func(i int, s *goquery.Selection) {
			html, _ := s.Html()
			htmlText := striptags.Sanitize(html)
			htmlText = filterFunc.ClearBreak(htmlText)

			if len(htmlText) > 0 {
				*output = append(*output, htmlText)
			}
		})
	}
}

// Метод ищет содержание дочерних элементов
func (indx *Indexing) ChildrenNodes(s *goquery.Selection, output *[]string) {
	striptags := bluemonday.StripTagsPolicy()

	striptags.AddSpaceWhenStrippingTag(true)

	filterFunc := Filter{}
	html, _ := s.Html()
	htmlText := filterFunc.ClearBreak(html)
	htmlText = striptags.Sanitize(htmlText)

	if len(htmlText) > 0 {
		*output = append(*output, htmlText)
	}

	if s.Length() > 0 {
		indx.ChildrenNodes(s.Children(), output)
	}
}

// Метод помечает страницу посещенной
func (indx *Indexing) PageCrawl(id *uint64) bool {
	db := dbpkg.Database{}
	ctx, dbn, err := db.ConnPgSQL("rw_pgsql_search")

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)

		return false
	}

	defer dbn.Close(ctx)

	var sql string = `
		UPDATE 
			web_pages 
		SET
			crawled_at = NOW()::timestamp
		WHERE
			id = $1
	`

	res, err := dbn.Exec(ctx, sql, *id)

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)

		return false
	}

	_ = res.RowsAffected()

	return true
}

// Метод обновляет данные индексируемой страницы
func (indx *Indexing) PageUpdate(id *uint64, details map[string]string) bool {
	db := dbpkg.Database{}
	ctx, dbn, err := db.ConnPgSQL("rw_pgsql_search")

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	defer dbn.Close(ctx)

	var sql string = `
		UPDATE 
			web_pages 
		SET
			page_url = $2,
			meta_title = $3,
			meta_description = $4,
			meta_keywords = $5,
			page_h1 = $6,
			page_text = $7,
			index_status = true,
			updated_at = NOW()::timestamp
		WHERE
			id = $1
	`

	res, err := dbn.Exec(ctx, sql, *id, details["url"], details["meta_title"], details["meta_description"], details["meta_keywords"], details["page_h1"], details["page_text"])

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	_ = res.RowsAffected()

	var sql2 string = `
		INSERT INTO vector_model_search
		(page_id, hint_id, created_at)
		(
			SELECT
				WP.id,
				SH.id,
				NOW()::timestamp
			FROM
				web_pages AS WP,
				search_hints AS SH
			WHERE
				WP.id = $1 AND
				(
					WP.meta_title LIKE '' || SH.query || '%' OR
					WP.page_h1 LIKE '' || SH.query || '%' OR
					WP.page_text LIKE '' || SH.query || '%'
				)
		)
		ON CONFLICT (page_id, hint_id) DO
		UPDATE SET updated_at = NOW()::timestamp
	`
	res2, err2 := dbn.Exec(ctx, sql2, *id)

	if err2 != nil {
		log := &Logs{}
		log.LogWrite(err2)
	}

	_ = res2.RowsAffected()

	return true
}

// Метод отключает страницу из индекса
func (indx *Indexing) PageDisableIndex(id *uint64) {
	db := dbpkg.Database{}
	ctx, dbn, err := db.ConnPgSQL("rw_pgsql_search")

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	defer dbn.Close(ctx)

	var sql string = `
		UPDATE 
			web_pages 
		SET
			index_status = false,
			updated_at = NOW()::timestamp
		WHERE
			id = $1
	`

	res, err := dbn.Exec(ctx, sql, *id)

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	_ = res.RowsAffected()
}

// Метод удаляет страницу из индекса
func (indx *Indexing) PageDeleteIndex(id *uint64, status_code *int) {
	db := dbpkg.Database{}
	ctx, dbn, err := db.ConnPgSQL("rw_pgsql_search")

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	defer dbn.Close(ctx)

	var sql string = `
		UPDATE 
			web_pages 
		SET
			index_status = false,
			is_delete = true,
			delete_time = NOW()::timestamp,
			http_code = $2
		WHERE
			id = $1
	`

	res, err := dbn.Exec(ctx, sql, *id, *status_code)

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	_ = res.RowsAffected()
}
