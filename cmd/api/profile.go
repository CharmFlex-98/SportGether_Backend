package main

import (
	"net/http"
	"sportgether/models"
	"sportgether/tools"
)

// User Profile

//func (app *Application) updateProfileIcon(w http.ResponseWriter, r *http.Request) {
//	input := struct {
//		ProfileIconUrl string `json:"profileIconUrl"`
//	}{}
//	err := app.readRequest(r, &input)
//	if err != nil {
//		app.writeBadRequestResponse(w, r)
//		return
//	}
//
//	user, ok := app.GetUserContext(r)
//	if !ok {
//		app.writeInvalidAuthenticationErrorResponse(w, r)
//		return
//	}
//
//	err = app.daos.UserDao.UpdateProfileIconUrl(user.ID, input.ProfileIconUrl)
//	if err != nil {
//		app.logError(err, r)
//		app.writeInternalServerErrorResponse(w, r)
//	}
//}

func (app *Application) checkIfUserOnboarded(w http.ResponseWriter, r *http.Request) {
	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	userOnboarded, err := app.daos.UserIsOnboarded(user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"onboarded": userOnboarded}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
	}
}

func (app *Application) onboardUser(w http.ResponseWriter, r *http.Request) {
	input := &struct {
		PreferredName *string `json:"preferredName"`
		BirthDate     *string `json:"birthDate"`
		Gender        *string `json:"gender"`
	}{}
	err := app.readRequest(r, &input)
	if err != nil {
		app.writeBadRequestResponse(w, r)
		return
	}

	validator := tools.RequestValidator{}
	validator.Check(input.PreferredName != nil, "preferredName", "Cannot be null")
	validator.Check(input.BirthDate != nil, "birthDate", "Cannot be null")
	validator.Check(input.Gender != nil, "gender", "Cannot be null")

	if !validator.Valid() {
		app.writeError(w, r, http.StatusBadRequest, http.StatusBadRequest, validator.Errors)
		return
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	onboarded, err := app.daos.OnboardUser(user.ID, *input.PreferredName, *input.BirthDate, *input.Gender)
	err = app.writeResponse(w, responseData{"onboarded": onboarded}, http.StatusCreated, nil)
	if err != nil {
		app.logError(err, r)
	}
}

func (app *Application) updateUserProfile(w http.ResponseWriter, r *http.Request) {
	input := &struct {
		PreferredName  *string `json:"preferredName"`
		BirthDate      *string `json:"birthDate"`
		Signature      *string `json:"signature"`
		Memo           *string `json:"memo"`
		ProfileIconUrl *string `json:"profileIconUrl"`
		Gender         *string `json:"gender"`
	}{}
	err := app.readRequest(r, input)
	if err != nil {
		app.writeBadRequestResponse(w, r)
		return
	}

	userProfileDetail := models.UserProfileDetail{
		PreferredName:  input.PreferredName,
		BirthDate:      input.BirthDate,
		Signature:      input.Signature,
		Memo:           input.Memo,
		ProfileIconUrl: input.ProfileIconUrl,
		Gender:         input.Gender,
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}
	err = app.daos.UpdateUserProfile(user.ID, userProfileDetail)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
	}
}

func (app *Application) getUserProfileDetail(w http.ResponseWriter, r *http.Request) {
	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	userProfileDetail, err := app.daos.GetProfileDetail(user.ID)
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"profileDetail": userProfileDetail}, http.StatusOK, nil)
	if err != nil {
		app.logError(err, r)
	}
}
