package validators

import "regexp"

const (
	passwordValRegexStr = "^.{6,24}$"
	emailValRegexStr    = "^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$"
)

func Password(password string) bool {
	var passwordRegex = regexp.MustCompile(passwordValRegexStr)
	return passwordRegex.MatchString(password)
}

func Email(email string) bool {
	var emailRegex = regexp.MustCompile(emailValRegexStr)
	return emailRegex.MatchString(email)
}
