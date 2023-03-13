package app

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Request struct {
	Redis *redis.Client
	Ctx   context.Context
}

type PageReqData struct {
	Url        string
	Body       io.Reader
	Header     http.Header
	StatusCode int
	Status     string
}

// Метод получает данные запрашиваемого url адреса
func (rq *Request) GetPageData(url *string) (PageReqData, bool) {
	var err = godotenv.Load()

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	nextUrl := *url

	var resp *http.Response
	var i int

	for i < 100 {
		req, err := http.NewRequest("GET", nextUrl, nil)
		req.Header.Set("User-Agent", os.Getenv("BOT_USERAGENT"))
		// req.Header.Set("Accept-Encoding", "deflate, gzip;q=1.0, *;q=0.5")

		if err != nil {
			log := &Logs{}
			log.LogWrite(err)
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err = client.Do(req)

		if err != nil {
			log := &Logs{}
			log.LogWrite(err)
		}

		if resp.StatusCode == 200 {
			// fmt.Println("Done!")
			break
		} else {
			location := resp.Header.Get("Location")

			if len(location) > 0 {
				parseLocation, _ := neturl.Parse(location)

				if len(parseLocation.Hostname()) <= 0 {
					ifFirst := nextUrl[(len(nextUrl) - 1):]
					ifSecond := location[0:1]

					if ifFirst == "/" && ifSecond == "/" {
						nextUrl = nextUrl[0:len(nextUrl)-1] + location
					} else if ifFirst != "/" && ifSecond != "/" {
						nextUrl = nextUrl + "/" + location
					} else {
						nextUrl = nextUrl + location
					}
				} else {
					nextUrl = resp.Header.Get("Location")
				}
			}

			i += 1
		}
	}

	return PageReqData{
		Url:        fmt.Sprintf("%v", resp.Request.URL),
		Body:       resp.Body,
		Header:     resp.Header,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}, true
}

// Метод фиксирует лимиты запросов в n секунд
func (rq *Request) IsRequestLimit(domain_full *string) bool {
	domain, _ := neturl.Parse(*domain_full)
	startTime := time.Now().Unix()

	var limSeconds int64 = 1
	var limQty int = 5

	requests, _ := rq.Redis.SMembers(rq.Ctx, domain.Host).Result()

	if len(requests) <= 0 {
		rq.Redis.SAdd(rq.Ctx, domain.Host, time.Now().Unix())

		return true
	} else {
		for {
			requests, _ = rq.Redis.SMembers(rq.Ctx, domain.Host).Result()

			if len(requests) < limQty {
				rq.Redis.SAdd(rq.Ctx, domain.Host, time.Now().Unix())

				for _, tm := range requests {
					_tm, _ := strconv.ParseInt(tm, 10, 64)

					if (time.Now().Unix() - _tm) > limSeconds {
						rq.Redis.SPop(rq.Ctx, domain.Host)
					}
				}

				return true
			}

			if time.Now().Unix()-startTime >= 300 {
				return false
			}
		}
	}
}

// Метод скачивает и получает содержимое файла (txt, csv и др.)
func (rq *Request) GetReadFile(url *string) []string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Get(*url)

	if err != nil {
		log := &Logs{}
		log.LogWrite(err)
	}

	defer resp.Body.Close()

	scan := bufio.NewScanner(resp.Body)

	var txtlines []string

	for scan.Scan() {
		txtlines = append(txtlines, scan.Text())
	}

	return txtlines
}
