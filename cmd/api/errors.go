package main

import (
	"net/http"
)

func (app *Application) writeBadRequestResponse(w http.ResponseWriter, r *http.Request) {
	message := "Malformed request. Please try again"
	app.writeError(w, r, http.StatusBadRequest, http.StatusBadRequest, message)
}

func (app *Application) writeInternalServerErrorResponse(w http.ResponseWriter, r *http.Request) {
	message := "Server encounters uan unknown error. Please try again later"
	app.writeError(w, r, http.StatusInternalServerError, http.StatusInternalServerError, message)
}
