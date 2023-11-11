package tools

import "unicode/utf8"

func ValidateEmail(requestValidator *RequestValidator, email string) {
	requestValidator.Check(email != "", "email", "email must not leave blank")
	requestValidator.Check(requestValidator.Matches(EmailRX, email), "email", "Malformed email format.")
}

func ValidateUsername(requestValidator *RequestValidator, username string) {
	requestValidator.Check(username != "", "username", "username must not leave blank")
	requestValidator.Check(utf8.RuneCountInString(username) >= 3, "username_min_length", "username too short")
	requestValidator.Check(utf8.RuneCountInString(username) <= 20, "username_max_length", "username too long")
}

func ValidatePassword(requestValidator *RequestValidator, password string) {
	requestValidator.Check(password != "", "email", "password must not leave blank")
	requestValidator.Check(utf8.RuneCountInString(password) >= 6, "password_min_length", "password must contain at least 6 chars")
}
