package utils

import (
	"encoding/base64"
	"os"
	"regexp"
	"strings"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func Base64DecodeStripped(s string) (string, error) {
	if i := len(s) % 4; i != 0 {
		s += strings.Repeat("=", 4-i)
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	return string(decoded), err
}

func IsValidCharacter(input string) bool {
	regex := `^[\s\w\d_.,-;()/]*$`
	match, err := regexp.MatchString(regex, input)
	if err != nil {
		return false
	}
	return match
}

func IsValidProductName(input string) bool {
	regex := `^[a-zA-Z0-9 _.,'-]+$`
	match, err := regexp.MatchString(regex, input)
	if err != nil {
		return false
	}
	return match
}

func IsEmptyString(value string) bool {
	return strings.TrimSpace(value) == ""
}
