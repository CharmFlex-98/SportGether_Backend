package validator

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
