package models

import (
	"context"
	"database/sql"
	"time"
)

type MessagingDao struct {
	db *sql.DB
}

func (dao MessagingDao) UpdateFCMToken(userId int64, token string) error {
	query := `
	INSERT INTO sportgether_schema.firebase_messaging_token_table (user_id, token)
	VALUES($1, $2) 
	ON CONFLICT (user_id) 
	DO
	UPDATE
	SET user_id=$1, token=$2;
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := dao.db.ExecContext(ctx, query, userId, token)
	if err != nil {
		return err
	}

	return nil
}

func (dao MessagingDao) GetEventParticipantTokens(eventId int64) (*[]string, error) {
	tokens := []string{}

	query := `SELECT fcm.token FROM sportgether_schema.users u
    	INNER JOIN sportgether_schema.event_participant ep on u.id = ep.participantid
         INNER JOIN sportgether_schema.firebase_messaging_token_table fcm on ep.participantid = fcm.user_id
		WHERE ep.eventid = $1
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := dao.db.QueryContext(ctx, query, eventId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var token string
		err = rows.Scan(&token)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &tokens, nil
}
