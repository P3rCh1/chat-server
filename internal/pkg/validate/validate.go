package validate

import (
	"regexp"
	"unicode"
)

var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func Email(email string) bool {
	return emailRegexp.MatchString(email)
}

func Username(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}
	return true
}

func Password(password string) bool {
	if len(password) < 8 || len(password) > 128 {
		return false
	}
	return true
}
