package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type MainMessage struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

func (app *Application) getMainMessage(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("./data/main_message_config.json")
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	mainMessage := &MainMessage{}
	err = json.Unmarshal(content, &mainMessage)
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
