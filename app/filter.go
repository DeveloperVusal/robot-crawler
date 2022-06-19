package app

import (
	"net/url"
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
	result = strings.Replace(result, "\t", "", -1)

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
