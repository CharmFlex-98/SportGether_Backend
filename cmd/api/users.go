package main

import (
	"net/http"
	"sportgether/models"
	"sportgether/tools"
	"unicode/utf8"
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

	validateUsername(reqValidator, input.Username)
	validateEmail(reqValidator, input.Email)
	validatePassword(reqValidator, input.Password)

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

	err = app.writeResponse(w, nil, http.StatusAccepted, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func validateEmail(requestValidator *validator.RequestValidator, email string) {
	requestValidator.Check(email != "", "email", "email must not leave blank")
	requestValidator.Check(requestValidator.Matches(validator.EmailRX, email), "email", "Malformed email format.")
}

func validateUsername(requestValidator *validator.RequestValidator, username string) {
	requestValidator.Check(username != "", "username", "username must not leave blank")
	requestValidator.Check(utf8.RuneCountInString(username) >= 3, "username_min_length", "username too short")
	requestValidator.Check(utf8.RuneCountInString(username) <= 20, "username_max_length", "username too long")
}

func validatePassword(requestValidator *validator.RequestValidator, password string) {
	requestValidator.Check(password != "", "email", "password must not leave blank")
	requestValidator.Check(utf8.RuneCountInString(password) >= 6, "password_min_length", "password must contain at least 6 chars")
}

func (app *Application) loginUser(w http.ResponseWriter, r *http.Request) {

}

// Validate
func validateRegisterUserResponse() {

}
