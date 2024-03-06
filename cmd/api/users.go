package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sportgether/constants"
	"sportgether/internal/models"
	"sportgether/tools"
	"time"
)

var (
	RequestActivationHeaderName = "x-sg-auth-req"
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

	if !user.ActivatedUser() {
		err = app.sendActivationRequest(&user, w, r)
		if err != nil {
			app.logError(err, r)
			app.writeInternalServerErrorResponse(w, r)
			return
		}
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

	// Check if user is activated, if not, send back client requesting activation
	if !user.ActivatedUser() {
		err = app.sendActivationRequest(user, w, r)
		if err != nil {
			app.logError(err, r)
			app.writeInternalServerErrorResponse(w, r)
			return
		}
	}

	// todo If everything ok, generate a token to user.
	tokenString, err := tools.GenerateJwtToken(user.ID, user.UserName, 876000, tools.AUTHENTICATION_SCOPE)
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

func (app *Application) sendActivationRequest(user *models.User, w http.ResponseWriter, r *http.Request) error {
	token, err := app.daos.TokenDao.New(user.ID, 3*time.Minute, models.AccountActivationScope)
	fmt.Printf("activate code: %s, hash: %s", token.PlainText, token.Hash)
	if err != nil {
		return err
	}

	app.background(func() {
		data := map[string]any{
			"userId":         user.UserName,
			"activationCode": token.PlainText,
		}
		err = app.mailer.Send(user.Email, "welcome_user.tmpl", data)
		if err != nil {
			// Just log error. We don't want to return error to client
			app.logError(err, r)
		}
	}, r)

	// Send back unauthorised error so that client can trigger activation code sending
	err = app.writeUserActivationRequiredResponse(w, r)
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) deregisterUserRequest(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	app.readRequest(r, &input)

	// todo Check if username exists in database. If not return error.
	user, err := app.daos.UserDao.GetByUsername(input.Username)
	if err != nil {
		return
	}

	// todo Check if password matches the one stored in database
	matches, err := user.Password.Matches(input.Password)
	if err != nil {
		return
	}
	if !matches {
		return
	}

	// If everything ok, send email to get deactivation token
	token, err := app.daos.TokenDao.New(user.ID, 3*time.Minute, models.AcccountDeactivationScope)
	if err != nil {
		return
	}

	app.background(func() {
		data := map[string]any{
			"userId":           user.UserName,
			"deactivationCode": token.PlainText,
		}
		err = app.mailer.Send(user.Email, "deactivate_user.tmpl", data)
		if err != nil {
			// Just log error. We don't want to return error to client
			app.logError(err, r)
		}
	}, r)
}

func (app *Application) activateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"code"`
	}
	err := app.readRequest(r, &input)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	validator := tools.NewRequestValidator()

	validator.Check(input.TokenPlainText != "", "token", "must be provided")
	validator.Check(len(input.TokenPlainText) == 26, "token", "must be 26 bytes long")

	if !validator.Valid() {
		app.logError(errors.New("Validation error"), r)
		app.writeResponse(w, validator.Errors, http.StatusBadRequest, nil)
		return
	}

	user, err := app.daos.GetUserByToken(models.AccountActivationScope, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.logError(errors.New("Token expired"), r)
			app.writeError(w, r, http.StatusBadRequest, http.StatusBadRequest, "Token expired")
			return
		default:
			app.logError(err, r)
			app.writeInternalServerErrorResponse(w, r)
			return
		}
	}

	user.Status = "ACTIVATED"

	err = app.daos.UpdateUser(*user)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.daos.DeleteAllForUser(models.AccountActivationScope, user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}
}

func (app *Application) deactivateUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	deactivationCode, err := app.readString(query, "code", "")
	if err != nil {
		app.logError(err, r)
		app.writeBadRequestResponse(w, r)
		return
	}

	validator := tools.NewRequestValidator()

	validator.Check(deactivationCode != "", "token", "must be provided")
	validator.Check(len(deactivationCode) == 26, "token", "must be 26 bytes long")

	if !validator.Valid() {
		app.logError(fmt.Errorf("deactivate account bad request: code: %s", deactivationCode), r)
		app.writeBadRequestResponse(w, r)
		return
	}

	user, err := app.daos.GetUserByToken(models.AcccountDeactivationScope, deactivationCode)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.logError(fmt.Errorf("either oken expired or no token found for deactivation"), r)
		default:
			app.logError(err, r)
		}
		app.writeBadRequestResponse(w, r)
		return
	}

	err = app.daos.DeleteUser(user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
	app.logInfo("User deleted", "username", user.UserName)
}
