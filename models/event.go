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
type GeoType struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
type Event struct {
	ID                  int64   `json:"id"`
	EventName           string  `json:"eventName"`
	HostId              int64   `json:"-"`
	StartTime           string  `json:"startTime"`
	EndTime             string  `json:"endTime"`
	Destination         string  `json:"destination"`
	Distance            float64 `json:"distance"`
	LongLat             GeoType `json:"longLat"`
	EventType           string  `json:"eventType"`
	MaxParticipantCount int     `json:"maxParticipantCount"`
	Description         string  `json:"description"`
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
	IsJoined        bool                     `json:"isJoined"`
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

func (eventDao EventDao) CreateEvent(event *Event) error {
	query := `
	INSERT INTO sportgether_schema.events (event_name, host_id, destination, long_lat, start_time, end_time, event_type, max_participant_count, description)
	VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7, $8, $9, $10)
	RETURNING id
`
	args := []any{
		event.EventName,
		event.HostId,
		event.Destination,
		event.LongLat.Longitude,
		event.LongLat.Latitude,
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

	fromLongitudeArgIndex := fmt.Sprintf("$%d", len(values)+1)
	fromLatitudeArgIndex := fmt.Sprintf("$%d", len(values)+2)
	values = append(values, filter.FromLocation.Longitude, filter.FromLocation.Latitude)

	query := fmt.Sprintf(`
	with event as (select * from sportgether_schema.events event %s) SELECT 
	    event.id, 
	    event_name, 
	    host_id,
	    u.username as host_name, 
	    u.profile_icon_name as host_profile_icon_name, 
	    destination, 
		ST_DistanceSphere(ST_SetSRID(ST_MakePoint(%s, %s), 4326), event.long_lat) as distance, 
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
		ORDER BY distance ASC
`, pagination, fromLongitudeArgIndex, fromLatitudeArgIndex)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := eventDao.db.QueryContext(ctx, query, values...)
	defer rows.Close()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, err
		default:
			return nil, err
		}
	}

	eventsMap := make(map[int64]*EventDetail)
	lastRowId := cursor.ID
	for rows.Next() {
		eventDetail := &EventDetail{
			Participants: []EventParticipantDetail{},
		}
		participant := struct {
			id              *int64
			name            *string
			profileIconName *string
		}{}

		err = rows.Scan(
			&eventDetail.Event.ID,
			&eventDetail.Event.EventName,
			&eventDetail.EventHostDetail.ParticipantId,
			&eventDetail.EventHostDetail.ParticipantUsername,
			&eventDetail.EventHostDetail.ProfileIconName,
			&eventDetail.Destination,
			&eventDetail.Distance,
			&eventDetail.StartTime,
			&eventDetail.EndTime,
			&eventDetail.EventType,
			&eventDetail.MaxParticipantCount,
			&eventDetail.Description,
			&participant.id,
			&participant.name,
			&participant.profileIconName,
		)
		if err != nil {
			return nil, err
		}

		eventDetail.IsHost = eventDetail.HostId == user.ID

		if _, ok := eventsMap[eventDetail.Event.ID]; !ok {
			eventsMap[eventDetail.Event.ID] = eventDetail
		}

		if participant.id != nil && participant.name != nil && participant.profileIconName != nil {
			eventsMap[eventDetail.Event.ID].Participants = append(eventsMap[eventDetail.Event.ID].Participants, EventParticipantDetail{
				ParticipantId:       *participant.id,
				ParticipantUsername: *participant.name,
				ProfileIconName:     *participant.profileIconName,
			})
		}
		if participant.id != nil && *participant.id == user.ID {
			eventsMap[eventDetail.Event.ID].IsJoined = true
		}

		lastRowId = &eventDetail.Event.ID
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

func (eventDao EventDao) GetEventById(eventId int64, userId int64) (*EventDetail, error) {
	eventDetail := EventDetail{}
	query := `
		WITH event AS (SELECT * FROM sportgether_schema.events e where e.id = $1)
		SELECT 
		    event.id, 
		    event.event_name, 
		    u.id as host_id, 
		    u.username as host_username, 
		    u.profile_icon_name as host_profile_icon_name, 
		    event.destination, 
		    event.start_time, 
		    event.end_time, 
		    event.event_type, 
		    event.max_participant_count, 
		    event.description
		
		FROM event
		INNER JOIN sportgether_schema.users u on event.host_id = u.id
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := eventDao.db.QueryRowContext(ctx, query, eventId).Scan(
		&eventDetail.ID,
		&eventDetail.EventName,
		&eventDetail.EventHostDetail.ParticipantId,
		&eventDetail.EventHostDetail.ParticipantUsername,
		&eventDetail.EventHostDetail.ProfileIconName,
		&eventDetail.Destination,
		&eventDetail.StartTime,
		&eventDetail.EndTime,
		&eventDetail.EventType,
		&eventDetail.MaxParticipantCount,
		&eventDetail.Description,
	)
	if err != nil {
		return nil, err
	}
	eventDetail.IsHost = eventDetail.HostId == userId

	query = `
	SELECT 
	    u.id, 
	    u.username, 
	    u.profile_icon_name
	FROM sportgether_schema.event_participant ep 
	inner join sportgether_schema.users u on u.id = ep.participantid
	WHERE ep.eventid = $1
`
	ctx, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()

	row, err := eventDao.db.QueryContext(ctx, query, eventId)
	if err != nil {
		return nil, err
	}

	for row.Next() {
		participant := EventParticipantDetail{}
		err := row.Scan(
			&participant.ParticipantId,
			&participant.ParticipantUsername,
			&participant.ProfileIconName,
		)
		if err != nil {
			return nil, err
		}
		eventDetail.Participants = append(eventDetail.Participants, participant)
	}
	eventDetail.IsJoined = true

	return &eventDetail, nil
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

func (eventDao EventDao) QuitEvent(eventId int64, userId int64) error {
	query := `
		DELETE FROM sportgether_schema.event_participant ep where ep.participantid = $1 and ep.eventid = $2
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := eventDao.db.ExecContext(ctx, query, userId, eventId)
	if err != nil {
		return err
	}

	return nil
}
