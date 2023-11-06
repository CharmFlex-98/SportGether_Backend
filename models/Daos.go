package models

import "database/sql"

type Daos struct {
	UserDao
}

func NewDaos(database *sql.DB) Daos {
	return Daos{
		UserDao{
			database,
		},
	}
}
