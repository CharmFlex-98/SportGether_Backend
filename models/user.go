package models

import (
	"context"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserDao struct {
	db *sql.DB
}

type password struct {
	plainTextPassword string
	passwordHashed    []byte
}

type User struct {
	ID        int64     `json:"id"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"password"`
	CreatedAt time.Time `json:"-"`
	IsBlocked bool      `json:"is_blocked"`
	Version   int32     `json:"-"`
}

func (dao UserDao) Insert(user *User) error {
	query := `
		INSERT INTO sportgether_schema.users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version	
`

	args := []any{
		user.UserName,
		user.Email,
		user.Password.passwordHashed,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := dao.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		return err
	}

	return nil
}

func (password *password) Set(plainTextPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}

	password.plainTextPassword = plainTextPassword
	password.passwordHashed = hashed

	return nil
}
