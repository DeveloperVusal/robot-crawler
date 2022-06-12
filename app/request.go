package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Request struct{}

type PageReqData struct {
	Url        string
	Body       string
	Header     http.Header
	StatusCode int
	Status     string
}

func (rq *Request) GetPageData(url string) PageReqData {
	var err = godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", os.Getenv("BOT_USERAGENT"))

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	return PageReqData{
		Url:        fmt.Sprintf("%v", resp.Request.URL),
		Body:       string(body),
		Header:     resp.Header,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}
}
