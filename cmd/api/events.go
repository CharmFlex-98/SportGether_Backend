package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sportgether/constants"
	"sportgether/models"
	"sportgether/tools"
)

func (app *Application) getAllEvents(w http.ResponseWriter, r *http.Request) {
	filter := tools.Filter{}
	err := app.readRequest(r, &filter)
	if err != nil {
		app.writeBadRequestResponse(w, r)
	}
	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	events, err := app.daos.EventDao.GetEvents(filter, user)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"events": events.Events, "nextCursorId": events.NextCursorId}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
}

func (app *Application) getUserEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	events, err := app.daos.GetUserEvents(user.ID)
	if err != nil {
		app.logError(err, r)

		switch {
		case errors.Is(err, constants.SportConfigNotFoundError):
			app.writeBadRequestResponse(w, r)
		default:
			app.writeInternalServerErrorResponse(w, r)
		}
		return
	}

	err = app.writeResponse(w, responseData{"userEvents": events.UserEvents}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
}

func (app *Application) createEvent(w http.ResponseWriter, r *http.Request) {
	input := struct {
		EventName           string         `json:"eventName"`
		StartTime           string         `json:"startTime"`
		EndTime             string         `json:"endTime"`
		Destination         string         `json:"destination"`
		LongLat             models.GeoType `json:"longLat"`
		EventType           string         `json:"eventType"`
		MaxParticipantCount int            `json:"maxParticipantCount"`
		Description         string         `json:"description"`
	}{}

	err := app.readRequest(r, &input)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	host, ok := app.GetUserContext(r)
	if !ok {
		app.logError(errors.New("cannot get user object from request context"), r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
	event := &models.Event{
		EventName:           input.EventName,
		HostId:              host.ID,
		StartTime:           input.StartTime,
		EndTime:             input.EndTime,
		Destination:         input.Destination,
		LongLat:             input.LongLat,
		EventType:           input.EventType,
		MaxParticipantCount: input.MaxParticipantCount,
		Description:         input.Description,
	}

	// Create transaction
	err = app.daos.WithTransaction(func(tx *sql.Tx) error {
		err = app.daos.CreateEvent(event, tx)
		if err != nil {
			return err
		}

		err = app.daos.JoinEvent(event.ID, host.ID, tx)
		if err != nil {
			return err
		}

		_, err = app.daos.UpdateUserHostingConfig(host.ID, true, tx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) getEventById(w http.ResponseWriter, r *http.Request) {
	eventId, err := app.readParam("eventId", r)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	event, err := app.daos.EventDao.GetEventById(*eventId, user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, event, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) joinEvent(w http.ResponseWriter, r *http.Request) {
	input := struct {
		EventId int64 `json:"eventId"`
	}{}
	err := app.readRequest(r, &input)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
	}

	// Get event detail
	eventDetail, err := app.daos.GetEventById(input.EventId, user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	// Create transaction
	err = app.daos.WithTransaction(func(tx *sql.Tx) error {

		// Try join event
		err = app.daos.JoinEvent(input.EventId, user.ID, tx)
		if err != nil {
			return err
		}

		// Get event participant count
		currentParticipantCount, err := app.daos.CheckEventParticipantCount(input.EventId)
		if err != nil {
			return err
		}

		// If exceed, revert join event
		if currentParticipantCount > eventDetail.MaxParticipantCount {
			return constants.StaleInfoError
		}

		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, constants.StaleInfoError):
			app.logError(errors.New("stale info error, race condition happenned, done reverting"), r)
			app.writeError(w, r, http.StatusConflict, constants.StaleInfoError.Code, "The event is staled. Please refresh")

		default:
			app.logError(err, r)
			app.writeInternalServerErrorResponse(w, r)
		}
	}

	detail, err := app.daos.GetProfileDetail(user.ID)
	if err != nil {
		app.logError(err, r)
		// Fail silently, so we don't want to affect the client.
		return
	}
	err = app.broadCastEventJoinedMessage(input.EventId, *detail.PreferredName)
	if err != nil {
		app.logError(err, r)
	}
}

func (app *Application) quitEvent(w http.ResponseWriter, r *http.Request) {
	value, err := app.readParam("eventId", r)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	err = app.daos.EventDao.QuitEvent(*value, user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) deleteEvent(w http.ResponseWriter, r *http.Request) {
	value, err := app.readParam("eventId", r)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	detail, err := app.daos.GetEventById(*value, user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	if !detail.IsHost {
		app.logError(errors.New(fmt.Sprintf("this user = %d cannot delete this event as no authority right", user.ID)), r)
		app.writeBadRequestResponse(w, r)
		return
	}

	err = app.daos.EventDao.DeleteEvent(*value)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.broadcastEventDeletedMessage(*value, user.ID)
	if err != nil {
		app.logError(err, r)
		// Fail silently
	}
}

func (app *Application) getEventHistory(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageNumber, err := app.readInt(query, "pageNumber", 0)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	pageSize, err := app.readInt(query, "pageSize", 0)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	res, err := app.daos.EventDao.GetHistory(user.ID, pageNumber, pageSize)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"history": res}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) getMutualEventInfo(w http.ResponseWriter, r *http.Request) {
	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	otherUserId, err := app.readParam("userId", r)
	if err != nil {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	joinedEventCount, err := app.daos.GetUserJoinedEventCount(*otherUserId)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	mutualEventCount, err := app.daos.GetMutualJoinedEventCount(user.ID, *otherUserId)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	output := struct {
		JoinedEventCount int `json:"joinedEventCount"`
		MutualEventCount int `json:"mutualEventCount"`
	}{
		JoinedEventCount: joinedEventCount,
		MutualEventCount: mutualEventCount,
	}

	err = app.writeResponse(w, responseData{"mutualEventInfo": output}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) initHostingConfig(w http.ResponseWriter, r *http.Request) {
	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	err := app.daos.InitialiseUserHostingConfig(user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) updateUserHostingConfigInfo(w http.ResponseWriter, r *http.Request) {
	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	config, err := app.daos.UpdateUserHostingConfig(user.ID, false, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"hostingConfigInfo": config}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}
