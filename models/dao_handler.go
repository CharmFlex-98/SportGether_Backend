package models

import (
	"database/sql"
)

type Daos struct {
	database *sql.DB
	UserDao
	EventDao
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
	}
}

type Transaction func() error

func (daos Daos) WithTransaction(transaction Transaction) error {
	tx, err := daos.database.Begin()
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	err = transaction()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
