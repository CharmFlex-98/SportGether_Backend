package main

import (
	"errors"
	"fmt"
	"net/http"
	"sportgether/constants"
	"sportgether/internal/models"
	"sportgether/tools"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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
				app.logError(err, r)
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

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic // as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or // not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been // sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type any, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using // our custom Logger type at the ERROR level and send the client a 500 // Internal Server Error response.
				app.logError(fmt.Errorf("There is panic triggered by error --> %s", err), r)
				app.writeInternalServerErrorResponse(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
