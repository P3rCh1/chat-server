package utils

import "regexp"

var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func CheckEmail(email string) bool {
	return emailRegexp.MatchString(email)
}
