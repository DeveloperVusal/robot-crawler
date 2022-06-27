package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	dbpkg "robot/database"

	"github.com/joho/godotenv"
)

type Request struct{}

type PageReqData struct {
	Url        string
	Body       io.Reader
	Header     http.Header
	StatusCode int
	Status     string
}

// Метод получает данные запрашиваемого url адреса
func (rq *Request) GetPageData(url *string) PageReqData {
	var err = godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", *url, nil)

	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", os.Getenv("BOT_USERAGENT"))
	// req.Header.Set("Accept-Encoding", "deflate, gzip;q=1.0, *;q=0.5")

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	return PageReqData{
		Url:        fmt.Sprintf("%v", resp.Request.URL),
		Body:       resp.Body,
		Header:     resp.Header,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}
}

// Метод фиксирует лимиты запросов в n секунд
func (rq *Request) IsRequestLimit(url *string) bool {
	startTime := time.Now().Unix()

	// Подключаемся к БД
	db := dbpkg.Database{}
	dbn, err := db.ConnMySQL("mysql")

	// Если есть ошибки выводим в лог
	if err != nil {
		log.Fatalln(err)
	}

	defer dbn.Close()

	var limSeconds int = 1
	var limQty int = 3

	var limResCount int

	sql := `SELECT 
				COUNT(id) AS COUNT
			FROM requests_limit
			WHERE
				(UNIX_TIMESTAMP(CURTIME(3)) - UNIX_TIMESTAMP(created_at)) < ?`

	for {
		err = dbn.QueryRow(sql, limSeconds).Scan(&limResCount)

		if err != nil {
			log.Fatalln(err)
		}

		if limResCount < limQty {
			_, err := dbn.Exec("INSERT INTO requests_limit (url) VALUES (?)", url)

			if err != nil {
				log.Fatalln(err)
			}

			_, err2 := dbn.Exec("DELETE FROM `requests_limit` WHERE (UNIX_TIMESTAMP(CURTIME(3)) - UNIX_TIMESTAMP(created_at)) > ?", limSeconds)

			if err2 != nil {
				log.Fatalln(err)
			}

			return true
		}

		if time.Now().Unix()-startTime >= 300 {
			return false
		}
	}
}

// Метод скачивает и получает содержимое файла (txt, csv и др.)
func (rq *Request) GetReadFile(url *string) []string {
	client := http.Client{}
	resp, err := client.Get(*url)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	scan := bufio.NewScanner(resp.Body)

	var txtlines []string

	for scan.Scan() {
		txtlines = append(txtlines, scan.Text())
	}

	return txtlines
}
