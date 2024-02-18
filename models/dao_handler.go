package models

import (
	"database/sql"
)

type Daos struct {
	database *sql.DB
	UserDao
	EventDao
	UserProfileDao
	MessagingDao
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
	}
}

type Transaction func(tx *sql.Tx) error

func (daos Daos) WithTransaction(transaction Transaction) error {
	tx, err := daos.database.Begin()
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
