package main

import (
	"net/http"
	"sportgether/models"
	"sportgether/tools"
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

	reqValidator := validator.NewRequestValidator()

	reqValidator.Check(input.Username != "", "username", "username must not leave blank")
	reqValidator.Check(input.Email != "", "email", "email must not leave blank")
	reqValidator.Check(input.Password != "", "email", "password must not leave blank")

	if !reqValidator.Valid() {
		app.writeError(w, r, http.StatusBadRequest, http.StatusBadRequest, reqValidator.Errors)
		return
	}

	user := models.User{
		UserName: input.Username,
		Email:    input.Email,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.writeInternalServerErrorResponse(w, r)
	}

	err = app.daos.UserDao.Insert(&user)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"user": user}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}

}

func (app *Application) loginUser(w http.ResponseWriter, r *http.Request) {

}

// Validate
func validateRegisterUserResponse() {

}
