package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	IsBlocked bool      `json:"is_blocked"`
	Version   int32     `json:"-"`
}
