package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"sportgether/constants"
	"strings"
	"time"
)

type UserDao struct {
	db *sql.DB
}

type password struct {
	plainTextPassword string
	passwordHashed    []byte
}

var (
	UnauthenticatedUser = &User{}
)

func unauthenticatedUser(user *User) bool {
	return user == UnauthenticatedUser
}

type User struct {
	ID              int64     `json:"id"`
	UserName        string    `json:"username"`
	Email           string    `json:"email"`
	Gender          string    `json:"gender"`
	ProfileIconName string    `json:"profileIconName"`
	Password        password  `json:"password"`
	CreatedAt       time.Time `json:"-"`
	IsBlocked       bool      `json:"is_blocked"`
	Version         int32     `json:"-"`
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

func (password *password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(password.passwordHashed, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (dao UserDao) GetByUsername(username string) (*User, error) {
	query := `SELECT * from sportgether_schema.users WHERE username = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}

	err := dao.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Gender,
		&user.ProfileIconName,
		&user.Password.passwordHashed,
		&user.CreatedAt,
		&user.IsBlocked,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, constants.UserNotFoundError
		default:
			return nil, err
		}
	}

	return user, nil
}

func (dao UserDao) GetById(userId int64) (*User, error) {
	query := `SELECT * from sportgether_schema.users WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}

	err := dao.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Gender,
		&user.ProfileIconName,
		&user.Password.passwordHashed,
		&user.CreatedAt,
		&user.IsBlocked,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, constants.UserNotFoundError
		default:
			return nil, err
		}
	}

	return user, nil
}

// UniqueConstrainError Constant
func UniqueConstrainError(err error, columnName string) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("duplicate key value violates unique constraint %q", "users_"+columnName+"_key"))
}
