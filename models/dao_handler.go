package models

import "database/sql"

type Daos struct {
	UserDao
	EventDao
}

func NewDaoHandler(database *sql.DB) Daos {
	return Daos{
		UserDao{
			database,
		},
		EventDao{
			db: database,
		},
	}
}
