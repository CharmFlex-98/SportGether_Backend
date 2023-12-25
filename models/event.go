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
	ID                  int64  `json:"id"`
	EventName           string `json:"eventName"`
	HostId              int64  `json:"-"`
	StartTime           string `json:"startTime"`
	EndTime             string `json:"endTime"`
	Destination         string `json:"destination"`
	EventType           string `json:"eventType"`
	MaxParticipantCount int    `json:"maxParticipantCount"`
	Description         string `json:"description"`
}

type EventParticipantDetail struct {
	ParticipantId       int64  `json:"userId"`
	ParticipantUsername string `json:"username"`
	ProfileIconName     string `json:"profileIconName"`
}

type EventHostDetail EventParticipantDetail

type EventDetail struct {
	Event
	EventHostDetail `json:"host"`
	IsHost          bool                     `json:"isHost"`
	Participants    []EventParticipantDetail `json:"participants"`
}

type UserScheduledEventsResponse struct {
	UserEvents []*UserScheduledEventDetail `json:"userEvents"`
}

type UserScheduledEventDetail struct {
	EventId     int64  `json:"eventId"`
	EventName   string `json:"eventName"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Destination string `json:"destination"`
	EventType   string `json:"eventType"`
}

type EventDetailResponse struct {
	Events       []*EventDetail `json:"events"`
	NextCursorId string         `json:"nextCursorId"`
}

type output struct {
	eventId                    int64
	eventName                  string
	hostId                     int64
	hostName                   string
	hostProfileIconName        string
	destination                string
	startTime                  string
	endTime                    string
	eventType                  string
	maxParticipantCount        int
	description                string
	participantId              *int64
	participantName            *string
	participantProfileIconName *string
}

func (eventDao EventDao) CreateEvent(event *Event) error {
	query := `
	INSERT INTO sportgether_schema.events (event_name, host_id, destination, start_time, end_time, event_type, max_participant_count, description)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id
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

	err := eventDao.db.QueryRowContext(ctx, query, args...).Scan(&event.ID)
	if err != nil {
		return err
	}

	return nil
}

func (eventDao EventDao) GetEvents(filter tools.Filter, user *User) (*EventDetailResponse, error) {
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

	if cursor.IsNext && cursor.ID != nil {
		pagination += fmt.Sprintf(" WHERE event.id > $%d ORDER BY event.id ASC LIMIT $%d", len(values)+1, len(values)+2)
		values = append(values, cursor.ID, filter.PageSize)
	} else if !cursor.IsNext && cursor.ID != nil {
		pagination += fmt.Sprintf(" WHERE event.id < $%d ORDER BY event.id DESC LIMIT $%d", len(values)+1, len(values)+2)
		values = append(values, cursor.ID, filter.PageSize)
	} else {
		pagination += fmt.Sprintf(" ORDER BY event.id ASC LIMIT $%d", len(values)+1)
		values = append(values, filter.PageSize)
	}

	query := fmt.Sprintf(`
	with event as (select * from sportgether_schema.events event %s) SELECT 
	    event.id, 
	    event_name, 
	    host_id,
	    u.username as host_name, 
	    u.profile_icon_name as host_profile_icon_name, 
	    destination, 
	    start_time, 
	    end_time, 
	    event_type, 
	    max_participant_count, 
	    description, 
	    ep.participantid as participant_id, 
	    u1.username as participant_name,
	    u1.profile_icon_name as participant_profile_icon_name from event
	    INNER JOIN sportgether_schema.users u ON host_id = u.id
	    LEFT JOIN sportgether_schema.event_participant ep on ep.eventid = event.id
	    LEFT join sportgether_schema.users u1 on ep.participantid = u1.id 
		ORDER BY event.id ASC
`, pagination)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := eventDao.db.QueryContext(ctx, query, values...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	eventsMap := make(map[int64]*EventDetail)
	lastRowId := cursor.ID
	for rows.Next() {
		output := output{}

		err = rows.Scan(
			&output.eventId,
			&output.eventName,
			&output.hostId,
			&output.hostName,
			&output.hostProfileIconName,
			&output.destination,
			&output.startTime,
			&output.endTime,
			&output.eventType,
			&output.maxParticipantCount,
			&output.description,
			&output.participantId,
			&output.participantName,
			&output.participantProfileIconName,
		)
		if err != nil {
			return nil, err
		}

		if _, ok := eventsMap[output.eventId]; !ok {
			eventsMap[output.eventId] = &EventDetail{
				Event: Event{
					ID:                  output.eventId,
					EventName:           output.eventName,
					HostId:              output.hostId,
					StartTime:           output.startTime,
					EndTime:             output.endTime,
					Destination:         output.destination,
					EventType:           output.eventType,
					MaxParticipantCount: output.maxParticipantCount,
					Description:         output.description,
				},
				IsHost: output.hostId == user.ID,
				EventHostDetail: EventHostDetail{
					ParticipantId:       output.hostId,
					ParticipantUsername: output.hostName,
					ProfileIconName:     output.hostProfileIconName,
				},
				Participants: []EventParticipantDetail{},
			}
		}

		if output.participantId != nil && output.participantName != nil && output.participantProfileIconName != nil {
			eventsMap[output.eventId].Participants = append(eventsMap[output.eventId].Participants, EventParticipantDetail{
				ParticipantId:       *output.participantId,
				ParticipantUsername: *output.participantName,
				ProfileIconName:     *output.participantProfileIconName,
			})
		}

		lastRowId = &output.eventId
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	lastCursor := tools.Cursor{ID: lastRowId, IsNext: true}
	nextCursorId, e := tools.EncodeToBase32(lastCursor)
	if e != nil {
		return nil, e
	}

	res := &EventDetailResponse{
		Events: []*EventDetail{},
	}

	res.NextCursorId = nextCursorId
	for _, event := range eventsMap {
		res.Events = append(res.Events, event)
	}

	return res, nil
}

func (EventDao EventDao) GetUserEvents(userId int64) (*UserScheduledEventsResponse, error) {
	query := `
  		select 
  		    e.id, 
  			e.event_name,
  			e.start_time, 
  			e.end_time, 
  			e.destination, 
  			e.event_type
  		from sportgether_schema.event_participant ep 
  		inner join sportgether_schema.events e on ep.eventId = e.id
  		WHERE ep.participantId = $1
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := EventDao.db.QueryContext(ctx, query, userId)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	userEvents := []*UserScheduledEventDetail{}
	for rows.Next() {
		event := &UserScheduledEventDetail{}
		err = rows.Scan(
			&event.EventId,
			&event.EventName,
			&event.StartTime,
			&event.EndTime,
			&event.Destination,
			&event.EventType,
		)
		if err != nil {
			return nil, err
		}
		userEvents = append(userEvents, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &UserScheduledEventsResponse{
		UserEvents: userEvents,
	}, nil
}

func (eventDao EventDao) JoinEvent(eventId int64, participantId int64) error {
	query := `
	INSERT INTO sportgether_schema.event_participant (eventid, participantid)
	VALUES ($1, $2)
`
	args := []any{
		eventId,
		participantId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := eventDao.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
