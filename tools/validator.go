package validator

import "regexp"

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type RequestValidator struct {
	Errors map[string]string
}

func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		Errors: make(map[string]string),
	}
}

func (v *RequestValidator) Valid() bool {
	return len(v.Errors) <= 0
}

func (v *RequestValidator) Check(valid bool, errorType string, detail string) {
	if !valid {
		v.appendError(errorType, detail)
	}
}

func (v *RequestValidator) appendError(errorType string, detail string) {
	if _, exist := v.Errors[errorType]; !exist {
		v.Errors[errorType] = detail
	}
}

func (v *RequestValidator) Matches(regex *regexp.Regexp, email string) bool {
	return regex.MatchString(email)
}
