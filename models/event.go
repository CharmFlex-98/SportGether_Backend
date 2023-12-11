package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sportgether/tools"
	"time"
)

type EventDao struct {
	db *sql.DB
}

type Event struct {
	ID                  string `json:"id"`
	EventName           string `json:"eventName"`
	HostId              int64  `json:"-"`
	StartTime           string `json:"startTime"`
	EndTime             string `json:"endTime"`
	Destination         string `json:"destination"`
	EventType           string `json:"eventType"`
	MaxParticipantCount int    `json:"maxParticipantCount"`
	Description         string `json:"description"`
}

type EventHostDetail struct {
	HostId       int64  `json:"userId"`
	HostUsername string `json:"username"`
}

type EventParticipantDetail struct {
	ParticipantId       int64  `json:"userId"`
	ParticipantUsername string `json:"username"`
}

type EventDetail struct {
	Event
	EventHostDetail `json:"host"`
	Participants    []EventParticipantDetail `json:"participants"`
}

func (eventDao EventDao) CreateEvent(event *Event) error {
	query := `
	INSERT INTO sportgether_schema.events (event_name, host_id, destination, start_time, end_time, event_type, max_participant_count, description)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`
	args := []any{
		event.EventName,
		event.HostId,
		event.Destination,
		event.StartTime,
		event.EndTime,
		event.EventType,
		event.MaxParticipantCount,
		event.Description,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := eventDao.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (eventDao EventDao) GetEvents(filter tools.Filter) ([]*EventDetail, error) {
	validator := tools.NewRequestValidator()
	filter.Validate(validator)

	if !validator.Valid() {
		data, e := json.Marshal(validator.Errors)
		if e != nil {
			return nil, e
		}
		return nil, errors.New(string(data))
	}

	err, cursor := filter.DecodeCursor()
	if err != nil {
		return nil, err
	}

	pagination := ""
	values := []any{}

	if filter.HasNextCursor() && cursor.ID != nil {
		pagination += fmt.Sprintf(" WHERE id > $%d ORDER BY id ASC LIMIT $%d", len(values)+1, len(values)+2)
		values = append(values, cursor.ID, filter.PageSize)
	} else if filter.HasPrevCursor() && cursor.ID != nil {
		pagination += fmt.Sprintf(" WHERE id < $%d ORDER BY id DESC LIMIT $%d", len(values)+1, len(values)+2)
		values = append(values, cursor.ID, filter.PageSize)
	}

	query := fmt.Sprintf(`
WITH event AS (
    SELECT * from sportgether_schema.events %s
)
	SELECT 
	    event.id, 
	    event_name, 
	    host_id, 
	    destination, 
	    start_time, 
	    end_time, 
	    event_type, 
	    max_participant_count, 
	    description, 
	    username

	FROM event
	    INNER JOIN sportgether_schema.users u
	    ON host_id = u.id
`, pagination)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := eventDao.db.QueryContext(ctx, query, values...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	events := []*EventDetail{}
	for rows.Next() {
		event := EventDetail{
			Participants: []EventParticipantDetail{},
		}

		err := rows.Scan(
			&event.ID,
			&event.EventName,
			&event.EventHostDetail.HostId,
			&event.Destination,
			&event.StartTime,
			&event.EndTime,
			&event.EventType,
			&event.MaxParticipantCount,
			&event.Description,
			&event.EventHostDetail.HostUsername,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
