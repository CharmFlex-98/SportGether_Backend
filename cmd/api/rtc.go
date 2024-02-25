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

func (app *Application) broadCastEventJoinedMessage(eventId int64, userPreferredName string) error {
	tokens, err := app.daos.GetEventParticipantTokens(eventId)
	if err != nil {
		return err
	}

	ctx := context.Background()
	client, err := app.firebaseApp.Messaging(ctx)
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

	num, err := client.SendEachForMulticast(context.Background(), message)
	if err != nil {
		return err
	}
	app.logInfo("success count: %d, failure count: %d, response: %s", num.SuccessCount, num.FailureCount, num.Responses)

	return nil
}

func (app *Application) broadcastEventDeletedMessage(eventId int64, userId int64) error {
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

	ctx := context.Background()
	client, err := app.firebaseApp.Messaging(ctx)
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

	num, err := client.SendEachForMulticast(context.Background(), message)
	if err != nil {
		return err
	}
	app.logInfo("success count: %d, failure count: %d, response: %s", num.SuccessCount, num.FailureCount, num.Responses)

	return nil
}
