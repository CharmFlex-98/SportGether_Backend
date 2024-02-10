package main

import (
	"net/http"
	remoteConfig "sportgether/remote_config"
)

type MainMessage struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	ButtonText string `json:"buttonText"`
}

func (app *Application) getMainMessage(w http.ResponseWriter, r *http.Request) {
	mainMessage := MainMessage{}
	err := readJsonFromFile("./data/main_message_config.json", &mainMessage)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"message": mainMessage}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
}

func (app *Application) getSportDetails(w http.ResponseWriter, r *http.Request) {
	sportDetails := remoteConfig.SportDetails{}
	err := readJsonFromFile("./data/available_sports_detail.json", &sportDetails)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, sportDetails, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
}
