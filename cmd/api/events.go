package main

import "net/http"

func (app *Application) getAllEvents(w http.ResponseWriter, r *http.Request) {
	res := "success"
	err := app.writeResponse(w, responseData{"testing": res}, http.StatusOK, nil)
	if err != nil {
		app.writeError(w, r, http.StatusInternalServerError, http.StatusInternalServerError, "Hmm")
	}
}
