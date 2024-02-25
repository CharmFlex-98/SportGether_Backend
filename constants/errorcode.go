package constants

import (
	"errors"
)

type ErrorCode struct {
	Code int
	error
}

var (
	UserNotFoundError        = ErrorCode{Code: 10004, error: errors.New("username not found")}
	WrongPasswordError       = ErrorCode{Code: 10003, error: errors.New("wrong password")}
	RegisteredEmailError     = ErrorCode{Code: 10002, error: errors.New("email is registered")}
	RegisteredUsernameError  = ErrorCode{Code: 10001, error: errors.New("username is registered")}
	SportConfigNotFoundError = ErrorCode{Code: 20000, error: errors.New("Sport config is not found")}
	StaleInfoError           = ErrorCode{Code: 22222, error: errors.New("Stale info")}
)
