package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sportgether/constants"
	"sportgether/remote_config"
	"sportgether/tools"
	"sportgether/utils"
	"strings"
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
	ParticipantId            int64   `json:"userId"`
	ParticipantUsername      string  `json:"username"`
	ParticipantPreferredName string  `json:"userPreferredName"`
	ProfileIconUrl           *string `json:"profileIconUrl"`
}

type EventHostDetail EventParticipantDetail

type EventDetail struct {
	Event
	EventHostDetail `json:"host"`
	IsHost          bool                     `json:"isHost"`
	IsJoined        bool                     `json:"isJoined"`
	Status          EventStatus              `json:"status"`
	Participants    []EventParticipantDetail `json:"participants"`
	Version         int                      `json:"version"`
}

type EventStatus string

var (
	full           = EventStatus("FULL")
	available      = EventStatus("AVAILABLE")
	eventCancelled = EventStatus("CANCEL")
)

type UserScheduledEventsResponse struct {
	UserEvents []*UserScheduledEventDetail `json:"userEvents"`
}

type UserScheduledEventDetail struct {
	EventId       int64  `json:"eventId"`
	EventName     string `json:"eventName"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Destination   string `json:"destination"`
	EventType     string `json:"eventType"`
	Deleted       bool   `json:"isDeleted"`
	SportImageUrl string `json:"sportImageUrl"`
}

type EventDetailResponse struct {
	Events       []*EventDetail `json:"events"`
	NextCursorId string         `json:"nextCursorId"`
}

type EventHistoryResponse struct {
	EventName      string `json:"eventName"`
	EventStartTime string `json:"eventStartTime"`
	EventType      string `json:"eventType"`
}

func (eventDao EventDao) UpdateEvent(event *Event) error {
	query := `
	UPDATE sportgether_schema.events
	SET start_time = $1, end_time = $2, description = $3
	WHERE id = $4 AND host_id = $5
	`

	args := []any{
		event.StartTime,
		event.EndTime,
		event.Description,
		event.ID,
		event.HostId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := eventDao.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil

}

func (eventDao EventDao) CreateEvent(event *Event, tx *sql.Tx) error {
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

	err := tx.QueryRowContext(ctx, query, args...).Scan(&event.ID)
	if err != nil {
		return err
	}

	return nil
}

func (eventDao EventDao) GetEvents(filter tools.Filter, user *User) (*EventDetailResponse, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
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

	values := []any{}

	eventTypeQueryPlaceHolder := make([]string, 0, len(filter.EventTypes))
	for _, value := range filter.EventTypes {
		eventTypeQueryPlaceHolder = append(eventTypeQueryPlaceHolder, fmt.Sprintf("$%d", len(values)+1))
		values = append(values, value)
	}
	var eventTypeQuery = ""
	if len(filter.EventTypes) > 0 {
		eventTypeQuery = fmt.Sprintf(" event.event_type IN (%s)", strings.Join(eventTypeQueryPlaceHolder, ","))
	} else {
		eventTypeQuery = fmt.Sprintf(" event.event_type IN ($%d)", len(values)+1)
		values = append(values, "-1")
	}

	whereClause := fmt.Sprintf("WHERE event.start_time > $%d AND %s AND event.deleted IS FALSE", len(values)+1, eventTypeQuery)
	values = append(values, time.Now())

	distanceQuery := fmt.Sprintf("ST_DistanceSphere(ST_SetSRID(ST_MakePoint($%d, $%d), 4326), event.long_lat)", len(values)+1, len(values)+2)
	values = append(values, filter.FromLocation.Longitude, filter.FromLocation.Latitude)

	visitedIndex := make([]string, 0, len(cursor.VisitedEventIndex))
	for _, value := range cursor.VisitedEventIndex {
		visitedIndex = append(visitedIndex, fmt.Sprintf("%d", value))
	}
	visitedQuery := fmt.Sprintf("event.id NOT IN (%s)", strings.Join(visitedIndex, ","))

	orderClause := fmt.Sprintf("ORDER BY %s ASC LIMIT $%d", distanceQuery, len(values)+1)
	values = append(values, filter.PageSize)

	if cursor.IsNext && cursor.LastDistance != nil {
		whereClause += fmt.Sprintf(" AND %s >= $%d AND %s", distanceQuery, len(values)+1, visitedQuery)
		values = append(values, *cursor.LastDistance)
	}

	query := fmt.Sprintf(`
	with event as (select * from sportgether_schema.events event %s %s) SELECT 
	    event.id, 
	    event_name, 
	    host_id,
	    u.username as host_name,
		up.preferred_name as host_preferred_name,
	    up.profile_icon_url as host_profile_icon_url, 
	    destination, 
		%s as distance, 
	    start_time, 
	    end_time, 
	    event_type, 
	    max_participant_count, 
	    description, 
	    ep.participantid as participant_id, 
	    u1.username as participant_name,
		pup.preferred_name as participant_preferred_name,
	    pup.profile_icon_url as participant_profile_icon_url from event
	    INNER JOIN sportgether_schema.users u ON host_id = u.id
		LEFT JOIN sportgether_schema.user_profile up ON host_id = up.user_id
	    LEFT JOIN sportgether_schema.event_participant ep on ep.eventid = event.id
	    LEFT join sportgether_schema.users u1 on ep.participantid = u1.id
		LEFT join sportgether_schema.user_profile pup on u1.id = pup.user_id
		ORDER by distance
`, whereClause, orderClause, distanceQuery)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := eventDao.db.QueryContext(ctx, query, values...)
	defer rows.Close()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, err
		default:
			logger.Error(err.Error())
			return nil, err
		}
	}

	eventsMap := make(map[int64]*EventDetail)
	newCursor := &tools.Cursor{IsNext: true, LastDistance: cursor.LastDistance, VisitedEventIndex: cursor.VisitedEventIndex}
	for rows.Next() {
		eventDetail := &EventDetail{
			Participants: []EventParticipantDetail{},
		}
		participant := struct {
			id             *int64
			name           *string
			preferredName  *string
			profileIconUrl *string
		}{}

		err = rows.Scan(
			&eventDetail.Event.ID,
			&eventDetail.Event.EventName,
			&eventDetail.EventHostDetail.ParticipantId,
			&eventDetail.EventHostDetail.ParticipantUsername,
			&eventDetail.EventHostDetail.ParticipantPreferredName,
			&eventDetail.EventHostDetail.ProfileIconUrl,
			&eventDetail.Destination,
			&eventDetail.Distance,
			&eventDetail.StartTime,
			&eventDetail.EndTime,
			&eventDetail.EventType,
			&eventDetail.MaxParticipantCount,
			&eventDetail.Description,
			&participant.id,
			&participant.name,
			&participant.preferredName,
			&participant.profileIconUrl,
		)
		if err != nil {
			return nil, err
		}

		eventDetail.HostId = eventDetail.EventHostDetail.ParticipantId
		eventDetail.IsHost = eventDetail.HostId == user.ID

		if _, ok := eventsMap[eventDetail.Event.ID]; !ok {
			eventsMap[eventDetail.Event.ID] = eventDetail
		}

		if participant.id != nil && participant.name != nil {
			eventsMap[eventDetail.Event.ID].Participants = append(eventsMap[eventDetail.Event.ID].Participants, EventParticipantDetail{
				ParticipantId:            *participant.id,
				ParticipantUsername:      *participant.name,
				ParticipantPreferredName: *participant.preferredName,
				ProfileIconUrl:           participant.profileIconUrl,
			})
			appendEventStatus(eventsMap[eventDetail.Event.ID])
		}
		if participant.id != nil && *participant.id == user.ID {
			eventsMap[eventDetail.Event.ID].IsJoined = true
		}

		if newCursor.LastDistance == nil || *newCursor.LastDistance == eventDetail.Distance {
			newCursor.VisitedEventIndex = append(newCursor.VisitedEventIndex, eventDetail.ID)
		} else {
			newCursor.VisitedEventIndex = []int64{eventDetail.ID}
		}

		newCursor.LastDistance = &eventDetail.Distance
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	nextCursorId, e := tools.EncodeToBase32(newCursor)
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
	slices.SortFunc(res.Events, func(a, b *EventDetail) int {
		if a.Distance < b.Distance {
			return -1
		}
		if a.Distance > b.Distance {
			return 1
		}
		return 0
	})

	return res, nil
}

func appendEventStatus(detail *EventDetail) {
	switch {
	case len(detail.Participants) >= detail.MaxParticipantCount:
		detail.Status = full
	default:
		detail.Status = available
	}
}

func (EventDao EventDao) GetUserEvents(userId int64) (*UserScheduledEventsResponse, error) {
	query := `
  		select 
  		    e.id, 
  			e.event_name,
  			e.start_time, 
  			e.end_time, 
  			e.destination, 
  			e.event_type, 
  			e.deleted
  		from sportgether_schema.event_participant ep 
  		inner join sportgether_schema.events e on ep.eventId = e.id
  		WHERE e.end_time > $1 AND ep.participantId = $2
		ORDER BY e.start_time
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := EventDao.db.QueryContext(ctx, query, time.Now(), userId)
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
			&event.Deleted,
		)
		if err != nil {
			return nil, err
		}
		event.SportImageUrl, err = remote_config.FromSportToImageUrl(event.EventType)
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
		    up.profile_icon_url as host_profile_icon_url, 
		    event.destination, 
		    event.start_time, 
		    event.end_time, 
		    event.event_type, 
		    event.max_participant_count, 
		    event.description, 
			event.deleted
		
		FROM event
		INNER JOIN sportgether_schema.users u on event.host_id = u.id
		LEFT JOIN sportgether_schema.user_profile up on u.id = up.user_id
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var cancelled bool
	err := eventDao.db.QueryRowContext(ctx, query, eventId).Scan(
		&eventDetail.ID,
		&eventDetail.EventName,
		&eventDetail.EventHostDetail.ParticipantId,
		&eventDetail.EventHostDetail.ParticipantUsername,
		&eventDetail.EventHostDetail.ProfileIconUrl,
		&eventDetail.Destination,
		&eventDetail.StartTime,
		&eventDetail.EndTime,
		&eventDetail.EventType,
		&eventDetail.MaxParticipantCount,
		&eventDetail.Description,
		&cancelled,
	)
	if err != nil {
		return nil, err
	}
	eventDetail.HostId = eventDetail.EventHostDetail.ParticipantId
	eventDetail.IsHost = eventDetail.HostId == userId
	if cancelled {
		eventDetail.Status = eventCancelled
	}

	query = `
	SELECT 
	    u.id, 
	    u.username, 
		up.preferred_name, 
	    up.profile_icon_url
	FROM sportgether_schema.event_participant ep 
	inner join sportgether_schema.users u on u.id = ep.participantid
	left join sportgether_schema.user_profile up on u.id = up.user_id
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
			&participant.ParticipantPreferredName,
			&participant.ProfileIconUrl,
		)
		if err != nil {
			return nil, err
		}
		eventDetail.Participants = append(eventDetail.Participants, participant)
		if participant.ParticipantId == userId {
			eventDetail.IsJoined = true
		}
	}

	return &eventDetail, nil
}

func (eventDao EventDao) JoinEventByOwner(eventId int64, ownerId int64, tx *sql.Tx) error {
	query := `
	INSERT INTO sportgether_schema.event_participant (eventid, participantid)
	VALUES ($1, $2)
`
	args := []any{
		eventId,
		ownerId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

// This will execute in Transaction, always
func (eventDao EventDao) JoinEventByParticipant(eventId int64, maxParticipantCount int, participantId int64) error {
	query := `
	INSERT INTO sportgether_schema.event_participant (eventid, participantid)
	SELECT $1, $2 
	FROM sportgether_schema.event_participant 
	WHERE eventid = $1 GROUP BY eventid HAVING count(participantid) < $3
`
	args := []any{
		eventId,
		participantId,
		maxParticipantCount,
	}

	fmt.Printf("eventId: %d, maxParticipantCount: %d, participantId: %d", eventId, maxParticipantCount, participantId)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// var eventIdOutput int64
	// var participantIdOutput int64
	res, err := eventDao.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if row_count, err := res.RowsAffected(); err != nil {
		return err
	} else if row_count == 0 {
		return constants.StaleInfoError
	}

	return nil

	// if eventIdOutput != eventId || participantId != participantIdOutput {
	// 	return constants.StaleInfoError
	// } else {
	// 	return nil

	// }
}

func (eventDao EventDao) CheckEventParticipantCount(eventId int64, tx *sql.Tx) (int, error) {
	query := `
		SELECT COUNT(*) FROM sportgether_schema.event_participant ep where ep.eventId = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := tx.QueryRowContext(ctx, query, eventId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil

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

func (eventDao EventDao) DeleteEvent(eventId int64) error {
	query := `
		UPDATE sportgether_schema.events
		    SET deleted = $1
		WHERE id = $2
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := eventDao.db.ExecContext(ctx, query, true, eventId)
	if err != nil {
		return err
	}

	return nil
}

func (eventDao EventDao) GetHistory(userId int64, pageNumber int64, pageSize int64) (*[]EventHistoryResponse, error) {
	logger := slog.Logger{}
	query := `
	SELECT 
	    event.event_name, 
	    event.event_type, 
	    event.start_time
	FROM sportgether_schema.events event
	INNER JOIN sportgether_schema.event_participant ep on ep.eventid = event.id
	WHERE event.end_time < $1 AND ep.participantid = $2 AND event.deleted IS FALSE
	ORDER BY event.start_time DESC LIMIT $3 OFFSET $4
`
	res := []EventHistoryResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := eventDao.db.QueryContext(ctx, query, time.Now(), userId, pageSize, (pageNumber-1)*pageSize)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, err
		default:
			logger.Error(err.Error())
			return nil, err
		}

	}

	for rows.Next() {
		event := EventHistoryResponse{}
		err := rows.Scan(
			&event.EventName,
			&event.EventType,
			&event.EventStartTime,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, event)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &res, nil
}

func (eventDao EventDao) GetUserJoinedEventCount(userId int64) (int, error) {
	query := `
	SELECT count(*) from sportgether_schema.users u
	         INNER JOIN sportgether_schema.event_participant ep on u.id = ep.participantid
			 INNER JOIN sportgether_schema.events e on ep.eventid = e.id
	WHERE u.id = $1 AND e.end_time < $2 AND e.deleted IS FALSE
`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := eventDao.db.QueryRowContext(ctx, query, userId, time.Now()).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, err
}

func (eventDao EventDao) GetMutualJoinedEventCount(userId int64, participantId int64) (int, error) {
	query := `
	SELECT  count(*)
	    from sportgether_schema.event_participant ep
		INNER JOIN sportgether_schema.event_participant ep2 on ep.eventid = ep2.eventid
		INNER JOIN sportgether_schema.events e on ep.eventid = e.id
	WHERE ep.participantid = $1 AND ep2.participantid = $2 AND e.deleted IS FALSE
`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := eventDao.db.QueryRowContext(ctx, query, userId, participantId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, err
}

func (eventDao EventDao) InitialiseUserHostingConfig(userId int64) error {
	query := `
		INSERT INTO sportgether_schema.user_hosting_config (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := eventDao.db.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

type UserHostingConfigInfo struct {
	HostCount    int    `json:"hostCount"`
	MaxHostCount int    `json:"maxHostCount"`
	RefreshInMin int    `json:"refreshInMin"`
	Status       string `json:"status"`
}

type HostingConfigurator struct {
	MaxCount      int `json:"maxCount"`
	RefreshPeriod int `json:"refreshPeriod"`
}

type UpdateHostingConfigInput struct {
	userId                int64
	needRefreshTime       bool
	needUpdateHostCount   bool
	currentUserConfigInfo UserHostingConfigInfo
	configurator          HostingConfigurator
}

// Update and return result if needed, else just return result, as though calling get
func (eventDao EventDao) UpdateUserHostingConfig(userId int64, updateAfterUserHosted bool, tx *sql.Tx) (*UserHostingConfigInfo, error) {
	var configurator HostingConfigurator = HostingConfigurator{}
	err := utils.ReadJsonFromFile("./data/hosting_config.json", &configurator)
	if err != nil {
		return nil, err
	}

	queryResult := struct {
		hostCount         int
		last_refresh_time time.Time
	}{}
	currentConfigQuery := `SELECT hc.host_count, hc.last_refresh_time FROM sportgether_schema.user_hosting_config hc WHERE hc.user_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = eventDao.db.QueryRowContext(ctx, currentConfigQuery, userId).Scan(&queryResult.hostCount, &queryResult.last_refresh_time)
	if err != nil {
		return nil, err
	}

	needRefresh := time.Since(queryResult.last_refresh_time) >= time.Duration(configurator.RefreshPeriod*int(time.Minute))

	currentUserConfigInfo := UserHostingConfigInfo{
		MaxHostCount: configurator.MaxCount,
		HostCount:    queryResult.hostCount,
	}
	if needRefresh || updateAfterUserHosted {
		input := UpdateHostingConfigInput{
			userId:                userId,
			needRefreshTime:       needRefresh,
			needUpdateHostCount:   updateAfterUserHosted,
			currentUserConfigInfo: currentUserConfigInfo,
			configurator:          configurator,
		}
		userHostingConfig, err := eventDao.updateHostingConfig(input, tx)
		if err != nil {
			return nil, err
		}
		return userHostingConfig, nil
	}

	currentUserConfigInfo.RefreshInMin = int(queryResult.last_refresh_time.Add(time.Duration(configurator.RefreshPeriod * int(time.Minute))).Sub(time.Now()).Minutes())
	currentUserConfigInfo.appendStatus(configurator)
	return &currentUserConfigInfo, nil
}

func (eventDao EventDao) updateHostingConfig(input UpdateHostingConfigInput, tx *sql.Tx) (*UserHostingConfigInfo, error) {
	values := []any{}
	setQuery := "SET"

	if input.needRefreshTime {
		setQuery += fmt.Sprintf(" last_refresh_time=$%d, host_count=$%d", len(values)+1, len(values)+2)
		values = append(values, time.Now(), 0)
	}
	if input.needUpdateHostCount {
		if input.needRefreshTime {
			values[len(values)-1] = 1
		} else {
			setQuery += fmt.Sprintf(" host_count=$%d", len(values)+1)
			values = append(values, input.currentUserConfigInfo.HostCount+1)
		}
	}

	updateQuery := fmt.Sprintf(`
	UPDATE sportgether_schema.user_hosting_config
	%s
	WHERE user_id = $%d
	RETURNING host_count, last_refresh_time
`, setQuery, len(values)+1)
	values = append(values, input.userId)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res = UserHostingConfigInfo{
		MaxHostCount: input.configurator.MaxCount,
	}
	var lastRefreshTime time.Time

	if tx != nil {
		err := tx.QueryRowContext(ctx, updateQuery, values...).Scan(&res.HostCount, &lastRefreshTime)
		if err != nil {
			return nil, err
		}
	} else {
		err := eventDao.db.QueryRowContext(ctx, updateQuery, values...).Scan(&res.HostCount, &lastRefreshTime)
		if err != nil {
			return nil, err
		}
	}

	res.RefreshInMin = int(lastRefreshTime.Add(time.Duration(input.configurator.RefreshPeriod * int(time.Minute))).Sub(time.Now()).Minutes())
	res.appendStatus(input.configurator)

	return &res, nil
}

func (config *UserHostingConfigInfo) appendStatus(configurator HostingConfigurator) {
	status := ""
	if config.HostCount >= configurator.MaxCount {
		status = "INVALID"
	} else {
		status = "VALID"
	}
	config.Status = status
}
