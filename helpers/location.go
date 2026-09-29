package helpers

// [Yasser] Location formatting helper.

import (
	"strings"
	"unicode"
)

// FormatLocation converts raw API locations into readable text.
//
//	"london-uk"          -> "London, UK"
//	"north_carolina-usa" -> "North Carolina, USA"
//	"new_zealand"        -> "New Zealand"
func FormatLocation(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	idx := strings.LastIndex(raw, "-")
	if idx == -1 {
		return titleCase(raw)
	}

	city := titleCase(raw[:idx])
	country := strings.ReplaceAll(raw[idx+1:], "_", " ")
	if len([]rune(country)) <= 3 { // uk, usa -> UK, USA
		country = strings.ToUpper(country)
	} else {
		country = titleCase(country)
	}

	if city == "" {
		return country
	}
	if country == "" {
		return city
	}
	return city + ", " + country
}

// titleCase replaces "_" and "-" with spaces and capitalises each word.
func titleCase(s string) string {
	s = strings.NewReplacer("_", " ", "-", " ").Replace(s)
	words := strings.Fields(s)
	for i, w := range words {
		r := []rune(strings.ToLower(w))
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}
