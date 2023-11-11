package models

import "database/sql"

type Daos struct {
	UserDao
}

func NewDaoHandler(database *sql.DB) Daos {
	return Daos{
		UserDao{
			database,
		},
	}
}
