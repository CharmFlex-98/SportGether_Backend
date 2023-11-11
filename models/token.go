package models

import "time"

type TokenDao struct {
	Hash   []byte
	token  string
	Expiry time.Time
	Scope  string
	userId int64
}
