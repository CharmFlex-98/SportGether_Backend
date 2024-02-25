package main

import (
	"context"
	"errors"
	"net/http"
	"sportgether/models"
	"sportgether/tools"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

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
	if err != nil {
		app.logError(err, r)
		app.writeInternalServerErrorResponse(w, r)
		return
	}

	err = app.writeResponse(w, responseData{"onboarded": onboarded}, http.StatusCreated, nil)
	if err != nil {
		app.logError(err, r)
	}
}

func (app *Application) updateUserProfile(w http.ResponseWriter, r *http.Request) {
	input := &struct {
		PreferredName       *string `json:"preferredName"`
		BirthDate           *string `json:"dateOfBirth"`
		Signature           *string `json:"signature"`
		Memo                *string `json:"memo"`
		ProfileIconUrl      *string `json:"profileIconUrl"`
		ProfileIconPublicId *string `json:"profileIconPublicId"`
		Gender              *string `json:"gender"`
	}{}
	err := app.readRequest(r, input)
	if err != nil {
		app.writeBadRequestResponse(w, r)
		return
	}

	userProfileDetail := models.UserProfileDetail{
		PreferredName:       input.PreferredName,
		BirthDate:           input.BirthDate,
		Signature:           input.Signature,
		Memo:                input.Memo,
		ProfileIconUrl:      input.ProfileIconUrl,
		ProfileIconPublicId: input.ProfileIconPublicId,
		Gender:              input.Gender,
	}

	user, ok := app.GetUserContext(r)
	if !ok {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	// If nothing needed to update, return
	if input.PreferredName == nil && input.BirthDate == nil && input.Signature == nil && input.Memo == nil && input.ProfileIconPublicId == nil && input.ProfileIconUrl == nil && input.Gender == nil {
		app.logWarning("User userId = %d no need to update profile since nothing is changed", user.ID)
		return
	}

	// This is to delete old profile icon url after performing and update.
	// This is fail safe. If fails, no need to do anything.
	if userProfileDetail.ProfileIconUrl != nil && userProfileDetail.ProfileIconPublicId != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userProfileIconUrl, publicId, err := app.daos.GetUserProfileIconUrl(user.ID)
		if err != nil {
			app.logError(err, r)
			app.writeInternalServerErrorResponse(w, r)
			return
		}

		// Delete only if existed.
		if userProfileIconUrl != nil && publicId != nil {
			res, err := app.cloudinaryApp.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: *publicId})
			if err != nil {
				app.logError(err, r)
				app.writeInternalServerErrorResponse(w, r)
				return
			}
			if res.Result != "ok" {
				app.logError(errors.New(res.Error.Message), r)
				app.writeInternalServerErrorResponse(w, r)
				return
			}
		}
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

func (app *Application) getOtherUserProfileDetail(w http.ResponseWriter, r *http.Request) {
	userId, err := app.readParam("userId", r)
	if err != nil {
		app.writeInvalidAuthenticationErrorResponse(w, r)
		return
	}

	userProfileDetail, err := app.daos.GetProfileDetail(*userId)
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
