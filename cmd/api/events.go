package main

import (
	"errors"
	"net/http"
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
		app.writeInternalServerErrorResponse(w, r)
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
		EventName           string `json:"eventName"`
		StartTime           string `json:"startTime"`
		EndTime             string `json:"endTime"`
		Destination         string `json:"destination"`
		EventType           string `json:"eventType"`
		MaxParticipantCount int    `json:"maxParticipantCount"`
		Description         string `json:"description"`
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
		EventType:           input.EventType,
		MaxParticipantCount: input.MaxParticipantCount,
		Description:         input.Description,
	}

	// Create transaction
	err = app.daos.WithTransaction(func() error {
		err = app.daos.CreateEvent(event)
		if err != nil {
			return err
		}

		err = app.daos.JoinEvent(event.ID, host.ID)
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
	err = app.daos.JoinEvent(input.EventId, user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}
