package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	dbpkg "robot/database"

	"github.com/joho/godotenv"
)

type Robotstxt struct {
	Domain_id   uint64
	IndexPgFind []string
	IndexpgRepl []string
}

// Метод проверяет ссылку на правила в robots.txt
func (r *Robotstxt) UrlHandle(link *string) (bool, string) {
	uParseDom, _ := url.Parse(*link)
	filename := uParseDom.Scheme + "://" + uParseDom.Host + "/robots.txt"

	fmt.Println("filename", filename)

	robotsData := r.get(&filename)
	userAgent := os.Getenv("BOT_USERAGENT")
	var isValid bool
	var handleUrl string

	for key, value := range robotsData {
		if key == userAgent {
			for _, val := range value {
				for k, v := range val {
					isValid, handleUrl = r.handleDirective(uParseDom, &k, &v)
				}
			}
		}
	}

	if len(handleUrl) > 4 {
		return isValid, handleUrl
	} else {
		return isValid, ""
	}
}

// Получаем содержимое robots.txt валидное для ButaGoBot
func (r *Robotstxt) get(filename *string) map[string][]map[string][]string {
	db := dbpkg.Database{}
	ctx, dbn, err := db.ConnPgSQL("pgsql")

	if err != nil {
		log.Fatalln(err)
	}

	defer dbn.Close(ctx)

	var sql = ` SELECT 
					RB.data
				FROM robotstxt RB
				WHERE
					(extract(epoch from localtimestamp) - extract(epoch from RB.updated_at) < 86400) AND
					RB.domain_id = $1`
	var rbData string

	rows, err := dbn.Query(ctx, sql, r.Domain_id)

	if err != nil {
		log.Fatalln(err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&rbData)

		if err != nil {
			log.Fatalln(err)
		}
	}

	if err != nil {
		log.Fatalln(err)
	}

	var rulesRobots map[string][]map[string][]string

	if rbData == "" {
		req := Request{}
		rbtxtxData := req.GetReadFile(filename)

		rulesRobots = r.parse(&rbtxtxData)
		jsonStr, err := json.Marshal(rulesRobots)

		if err != nil {
			fmt.Println(err.Error())
		} else {
			sql := `
				INSERT INTO robotstxt
				(domain_id, data, updated_at, created_at)
				VALUES
				($1, $2, NOW()::timestamp, NOW()::timestamp)
				ON CONFLICT (domain_id) DO
				UPDATE SET updated_at = NOW()::timestamp
			`
			res, err := dbn.Exec(ctx, sql, int(r.Domain_id), string(jsonStr))

			if err != nil {
				log.Fatalln(err)
			}

			_ = res.RowsAffected()
		}
	} else {
		json.Unmarshal([]byte(rbData), &rulesRobots)
	}

	return rulesRobots
}

// Парсим содержимое robots.txt
func (r *Robotstxt) parse(data *[]string) map[string][]map[string][]string {
	var err = godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	userAgent := os.Getenv("BOT_USERAGENT")
	rules := map[string][]map[string][]string{
		userAgent: {},
	}
	isAllowBot := false
	filterFunc := &Filter{}

	for _, line := range *data {
		re, _ := regexp.Compile(`^ *?#`)
		matched := re.MatchString(line)

		if !matched {
			expl := strings.Split(line, ":")

			if len(expl) > 1 {
				directive := strings.Trim(strings.ToLower(expl[0]), " \t\r\n")
				expl = filterFunc.RemoveSliceStr(expl, 0)
				value := strings.Trim(strings.Join(expl, ":"), " \t\r\n")

				r.handleRules(&directive, &value, &rules, &isAllowBot)
			} else {
				continue
			}
		} else {
			continue
		}
	}

	return rules
}

// Обработка правил robots.txt
func (r *Robotstxt) handleRules(dir *string, val *string, rules *map[string][]map[string][]string, isAllowBot *bool) {
	userAgent := os.Getenv("BOT_USERAGENT")
	filterFunc := &Filter{}

	if *dir == "user-agent" {
		if *val == "*" || *val == userAgent {
			*isAllowBot = true
		} else {
			*isAllowBot = false
		}
	}

	if *isAllowBot && *dir != "user-agent" {
		if len((*rules)[userAgent]) <= 0 {
			(*rules)[userAgent] = append((*rules)[userAgent], map[string][]string{*dir: {*val}})
		} else {
			isDirFind := false

			for key, vl := range (*rules)[userAgent] {
				if _, ok := vl[*dir]; ok {
					(*rules)[userAgent][key][*dir] = append((*rules)[userAgent][key][*dir], *val)

					(*rules)[userAgent][key][*dir] = filterFunc.SliceStrUnique((*rules)[userAgent][key][*dir])

					isDirFind = true
				}
			}

			if !isDirFind {
				(*rules)[userAgent] = append((*rules)[userAgent], map[string][]string{*dir: {*val}})
			}
		}
	}
}

// ОБработка директив robots.txt
func (r *Robotstxt) handleDirective(_url *url.URL, dir *string, data *[]string) (bool, string) {
	filterFunc := &Filter{}

	switch *dir {
	case "clean-param":
		body := _url.Query()

		for i := range *data {
			if len((*data)[i]) > 500 {
				return false, fmt.Sprintf("%v", _url)
			}

			var params []string
			var section, strParams string

			if strings.Contains((*data)[i], " ") {
				strParams, section = filterFunc.Unlist(strings.Split(strings.Trim((*data)[i], " \t"), " "))
				params = strings.Split(strParams, "&")
			} else {
				params = strings.Split(strings.Trim((*data)[i], " \t"), "&")
				section = ""
			}

			for key := range body {
				for j := range params {
					condition := true

					if len(section) > 0 {
						cond, _ := regexp.MatchString(filterFunc.ReplaceArrayStr(section, &r.IndexPgFind, &r.IndexpgRepl), _url.Path)

						if cond && key == params[j] {
							condition = false
						}
					} else {
						if key == params[j] {
							condition = false
						}
					}

					if !condition {
						body.Del(key)
					}
				}
			}
		}

		_url.RawQuery = body.Encode()

		return true, fmt.Sprintf("%v", _url)

	case "disallow":
		positive := []string{}

		for i := range *data {
			if len((*data)[i]) > 0 {
				condition, _ := regexp.MatchString(filterFunc.ReplaceArrayStr((*data)[i], &r.IndexPgFind, &r.IndexpgRepl), _url.Path)

				if condition {
					positive = append(positive, (*data)[i])
				}
			}
		}

		if len(positive) > 0 {
			return false, ""
		} else {
			return true, ""
		}
	default:
		return true, ""
	}
}