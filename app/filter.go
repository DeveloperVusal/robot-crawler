package app

import (
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

type Filter struct{}

// Метод проверяет на валидность ссылки
func (f *Filter) IsValidLink(s string, domaincheck string) bool {
	matched, _ := regexp.MatchString(`((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w_-]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)`, s)

	if !matched {
		return false
	} else {
		matched, _ := regexp.MatchString(`^(mailto:)|(tel:)|(callto:)|(callto:)|(javascript:)|(whatsapp:)|(viber:)|(telegram:)`, s)

		if matched {
			return false
		} else {
			uParseDom, _ := url.Parse(domaincheck)
			pattern := `^(` + uParseDom.Scheme + `\:)?\/\/` + strings.ReplaceAll(uParseDom.Host, ".", "\\.")

			matched, _ := regexp.MatchString(pattern, s)

			if matched {
				return true
			} else {
				return false
			}
		}
	}
}

// Метод удаляет переносы, табуляции и лишние пробелы
func (f *Filter) ClearBreak(s string) string {
	result := strings.Replace(s, "\r", "", -1)
	result = strings.Replace(result, "\n", "", -1)
	result = strings.Replace(result, "\r\n", "", -1)

	m1 := regexp.MustCompile(`\s+`)
	result = m1.ReplaceAllString(result, " ")

	return strings.Trim(result, " ")
}

// Метод возвращает массив из уникальных значений
func (f *Filter) SliceStrUnique(sl []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range sl {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

func (f *Filter) Substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

func (f *Filter) RemoveSliceStr(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}

func (f *Filter) ItemExists(arrayType interface{}, item interface{}) bool {
	arr := reflect.ValueOf(arrayType)

	if arr.Kind() != reflect.Array {
		log.Fatalln("Invalid data-type")
	}

	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

func (f *Filter) FilterSliceStr(slice []string, condition string) []string {
	pos := []string{}

	for i := range slice {
		if slice[i] == condition {
			pos = append(pos, slice[i])
		}
	}

	return pos
}

func (f *Filter) ReplaceArrayStr(slice string, find *[]string, repl *[]string) string {
	result := slice

	for key := range *find {
		result = strings.ReplaceAll(result, (*find)[key], (*repl)[key])
	}

	return result
}

func (f *Filter) Unlist(x []string) (string, string) {
	return x[0], x[1]
}
