package main

import (
	"errors"
	"net/http"
	"sportgether/constants"
	"sportgether/models"
	"sportgether/tools"
)

// Register User
func (app *Application) registerUser(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := app.readRequest(r, &input)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	reqValidator := tools.NewRequestValidator()

	tools.ValidateUsername(reqValidator, input.Username)
	tools.ValidateEmail(reqValidator, input.Email)
	tools.ValidatePassword(reqValidator, input.Password)

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
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}

	err = app.daos.UserDao.Insert(&user)
	if err != nil {
		switch {
		case models.UniqueConstrainError(err, "email"):
			app.writeError(w, r, http.StatusUnprocessableEntity, constants.RegisteredEmailError.Code, constants.RegisteredEmailError.Error())
		case models.UniqueConstrainError(err, "username"):
			app.writeError(w, r, http.StatusUnprocessableEntity, constants.RegisteredUsernameError.Code, constants.RegisteredUsernameError.Error())
		default:
			app.logError(err, r)
			app.writeInternalServerErrorResponse(w, r)
		}
		return
	}

	err = app.writeResponse(w, nil, http.StatusAccepted, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

// Login user
func (app *Application) loginUser(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	err := app.readRequest(r, &input)
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	validator := tools.NewRequestValidator()
	tools.ValidateUsername(validator, input.Username)
	tools.ValidatePassword(validator, input.Password)

	// todo If input is not valid, then return error
	if !validator.Valid() {
		app.writeError(w, r, http.StatusBadRequest, http.StatusBadRequest, validator.Errors)
		return
	}

	// todo Check if username exists in database. If not return error.
	user, err := app.daos.UserDao.GetByUsername(input.Username)
	var errorCode constants.ErrorCode
	if err != nil {
		switch {
		case errors.As(err, &errorCode):
			app.writeError(w, r, http.StatusUnprocessableEntity, errorCode.Code, errorCode.Error())
		default:
			app.logError(err, r)
			app.writeInternalServerErrorResponse(w, r)
		}
		return
	}

	// todo Check if password matches the one stored in database
	matches, err := user.Password.Matches(input.Password)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
	if !matches {
		app.writeError(w, r, http.StatusForbidden, constants.WrongPasswordError.Code, constants.WrongPasswordError.Error())
		return
	}

	// todo If everything ok, generate a token to user.
	tokenString, err := tools.GenerateJwtToken(user.ID, 720, tools.AUTHENTICATION_SCOPE)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"token": tokenString}, http.StatusCreated, nil)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) updateProfileIcon(w http.ResponseWriter, r *http.Request) {
	input := struct {
		ProfileIconUrl string `json:"profileIconUrl"`
	}{}
	err := app.readRequest(r, &input)
	if err != nil {
		app.writeBadRequestResponse(w, r)
		return
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	err = app.daos.UserDao.UpdateProfileIconUrl(user.ID, input.ProfileIconUrl)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}

}
