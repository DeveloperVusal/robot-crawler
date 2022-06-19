package app

import (
	"fmt"
	"log"
	"regexp"

	// dbpkg "robot/database"

	"github.com/PuerkitoBio/goquery"
	// _ "github.com/go-sql-driver/mysql"
	// "github.com/microcosm-cc/bluemonday"
)

type Buildpages struct {
	Domain_id   uint64
	Domain_full string
	Resp        *PageReqData
}

func (bp *Buildpages) Run(id uint64, url string, domain_id uint64, domain_full string) {
	if bp.Resp.StatusCode == 200 {
		matched, _ := regexp.MatchString(`^(text\/html|text\/plain)`, bp.Resp.Header.Get("Content-Type"))

		if matched {
			doc, err := goquery.NewDocumentFromReader(bp.Resp.Body)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(doc.Find("body").Html())

			// doc.Find("body").Each(func(i int, s *goquery.Selection) {
			// 	fmt.Println(s.Html())
			// })
		}

	}
}
