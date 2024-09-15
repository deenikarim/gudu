package main

import (
	"regexp"
	"strings"
	"unicode"
)

// capitalizeFirst Convert the first letter of each word to uppercase
func capitalizeFirst(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// toSnakeCase Convert string to Snake Case
func toSnakeCase(s string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	s = re.ReplaceAllString(s, "${1}_${2}")
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	return s

}

// toCamelCase Convert string to Camel Case
func toCamelCase(s string) string {
	s = toSnakeCase(s) // Start with snake_case
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = capitalizeFirst(parts[i])
	}

	return strings.Join(parts, "")
}

// ToLowerCamelCase Convert string to Lower Camel Case
func ToLowerCamelCase(s string) string {
	s = toCamelCase(s)
	return strings.ToLower(string(s[0])) + s[1:]
}
