package main

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"sportgether/constants"
	"sportgether/models"
	"sportgether/tools"
	"strings"
)

func (app *Application) authenticationHandler(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")

		// If user is unauthenticated
		if authorizationHeader == "" {
			r = app.SetUserContext(r, models.UnauthenticatedUser)
			nextHandler.ServeHTTP(w, r)
			return
		}

		// Obtain jwt token from header
		headers := strings.Split(authorizationHeader, " ")
		if len(headers) != 2 || headers[0] != "Bearer" {
			app.writeInvalidAuthenticationErrorResponse(w, r)
			return
		}

		token, err := tools.ParseJwtToken(headers[1])
		if err != nil {
			switch {
			case errors.Is(err, tools.BadTokenSignatureError):
				app.writeInvalidAuthenticationErrorResponse(w, r)
			default:
				app.logError(err, r)
				app.writeInvalidAuthenticationErrorResponse(w, r)
			}
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			app.writeInvalidAuthenticationErrorResponse(w, r)
			return
		}

		userId, isValid := tools.IsValidClaims(claims)
		if !isValid {
			app.writeInvalidAuthenticationErrorResponse(w, r)
			return
		}

		// user lookup
		user, err := app.daos.UserDao.GetById(*userId)
		if err != nil {
			switch {
			case errors.Is(err, constants.UserNotFoundError):
				app.writeInvalidAuthenticationErrorResponse(w, r)
			default:
				app.writeInternalServerErrorResponse(w, r)
			}

			return
		}

		// If user is authenticated
		r = app.SetUserContext(r, user)

		// Serve next handler
		nextHandler.ServeHTTP(w, r)
	})
}
