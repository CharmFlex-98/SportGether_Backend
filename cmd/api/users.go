package main

import (
	"net/http"
	"sportgether/models"
	"time"
)

func (app *Application) registerUser(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := app.readRequest(r, &input)
	if err != nil {
		app.writeBadRequestResponse(w, r)
		return
	}

	query := `INSERT INTO ()`

	user := models.User{
		ID:        1,
		UserName:  input.Username,
		Email:     input.Email,
		CreatedAt: time.Now(),
		IsBlocked: false,
		Version:   1,
	}

	err = app.writeResponse(w, responseData{"user": user}, http.StatusOK, nil)
	if err != nil {
		app.writeInternalServerErrorResponse(w, r)
	}

}

func (app *Application) loginUser(w http.ResponseWriter, r *http.Request) {

}
