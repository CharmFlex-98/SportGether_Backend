package models

import (
	"context"
	"database/sql"
	"time"
)

type Daos struct {
	database *sql.DB
	UserDao
	EventDao
	UserProfileDao
	MessagingDao
	TokenDao
}

func NewDaoHandler(database *sql.DB) Daos {
	return Daos{
		database,
		UserDao{
			database,
		},
		EventDao{
			db: database,
		},
		UserProfileDao{
			db: database,
		},
		MessagingDao{
			db: database,
		},
		TokenDao{
			db: database,
		},
	}
}

type Transaction func(tx *sql.Tx) error

func (daos Daos) WithTransaction(transaction Transaction) error {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := daos.database.BeginTx(context, nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	err = transaction(tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
