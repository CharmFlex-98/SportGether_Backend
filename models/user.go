package models

import (
	"context"
	"database/sql"
	"time"
)

type UserDao struct {
	db *sql.DB
}

type User struct {
	ID        int64     `json:"id"`
	UserName  string    `json:"usernam"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
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
		user.Password,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := dao.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		return err
	}

	return nil
}
