package main

import (
	"context"
	"net/http"
	"sportgether/models"
)

type contextKey string

var userContextKey = contextKey("user")

func (app *Application) SetUserContext(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *Application) GetUserContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	return user, ok
}
