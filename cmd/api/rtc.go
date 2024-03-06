package main

import (
	"context"
	"fmt"
	"net/http"

	"firebase.google.com/go/v4/messaging"
)

func (app *Application) registerFirebaseToken(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Token string `json:"token"`
	}{}

	err := app.readRequest(r, &input)
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

	err = app.daos.UpdateFCMToken(user.ID, input.Token)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) broadCastEventUpdatedMessage(r *http.Request, eventId int64) error {
	tokens, err := app.daos.GetEventParticipantTokens(eventId)
	if err != nil {
		return err
	}

	// Create a list containing up to 500 registration tokens.
	// This registration tokens come from the client FCM SDKs.
	message := &messaging.MulticastMessage{
		Data: map[string]string{
			"type":     "event",
			"eventId":  fmt.Sprintf("%d", eventId),
			"title":    "The event you participated had been updated",
			"subtitle": "Click here to view the updated details",
		},
		Tokens: *tokens,
	}

	app.fcmSend(r, context.Background(), message)

	return nil
}

func (app *Application) broadCastEventJoinedMessage(r *http.Request, eventId int64, userPreferredName string) error {
	tokens, err := app.daos.GetEventParticipantTokens(eventId)
	if err != nil {
		return err
	}

	// Create a list containing up to 500 registration tokens.
	// This registration tokens come from the client FCM SDKs.
	message := &messaging.MulticastMessage{
		Data: map[string]string{
			"type":     "event",
			"eventId":  fmt.Sprintf("%d", eventId),
			"title":    "Welcome your partner!",
			"subtitle": fmt.Sprintf("%s has joined the event!", userPreferredName),
		},
		Tokens: *tokens,
	}

	app.fcmSend(r, context.Background(), message)

	return nil
}

func (app *Application) broadcastEventDeletedMessage(r *http.Request, eventId int64, userId int64) error {
	// Get event detail
	event, err := app.daos.GetEventById(eventId, userId)
	if err != nil {
		return err
	}

	// Obtain a messaging.Client from the App.
	tokens, err := app.daos.GetEventParticipantTokens(eventId)
	if err != nil {
		return err
	}

	// Create a list containing up to 500 registration tokens.
	// This registration tokens come from the client FCM SDKs.
	message := &messaging.MulticastMessage{
		Data: map[string]string{
			"type":     "event",
			"eventId":  fmt.Sprintf("%d", eventId),
			"title":    "Event cancelled",
			"subtitle": fmt.Sprintf("The host had cancelled the event: %s", event.EventName),
		},
		Tokens: *tokens,
	}

	app.fcmSend(r, context.Background(), message)

	return nil
}

func (app *Application) fcmSend(r *http.Request, context context.Context, message *messaging.MulticastMessage) {
	app.background(func() {
		client, err := app.firebaseApp.Messaging(context)
		if err != nil {
			app.logError(err, r)
			return
		}

		num, err := client.SendEachForMulticast(context, message)
		if err != nil {
			app.logError(err, r)
		}
		app.logInfo("success count: %d, failure count: %d, response: %s", num.SuccessCount, num.FailureCount, num.Responses)
	}, r)
}
