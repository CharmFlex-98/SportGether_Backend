package constants

import "errors"

type ErrorCode struct {
	Code int
	error
}

var (
	UsernameNotFoundError = ErrorCode{Code: 1001, error: errors.New("username not found")}
	WrongPasswordError    = ErrorCode{Code: 1002, error: errors.New("wrong password")}
)
