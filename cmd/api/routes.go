package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *Application) routes() http.Handler {
	httpRouter := httprouter.New()

	httpRouter.NotFound = http.HandlerFunc(app.notFound)
	httpRouter.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	userHandlerFunc(app, httpRouter)
	eventHandlerFunc(app, httpRouter)
	profileHandlerFunc(app, httpRouter)
	messageCentreHandlerFunc(app, httpRouter)

	return app.authenticationHandler(httpRouter)
	//return httpRouter
}

// Custom method not allow handler
func (app *Application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The method '%s' is not supported!", r.Method)
	app.writeError(w, r, http.StatusMethodNotAllowed, http.StatusMethodNotAllowed, message)
}

// Custom not found handler
func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The URL '%s' is not found", r.URL.Path)
	app.writeError(w, r, http.StatusNotFound, http.StatusNotFound, message)
}

func userHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.HandlerFunc(http.MethodPost, "/v1/user/register", app.registerUser)
	httpRouter.HandlerFunc(http.MethodPost, "/v1/user/login", app.loginUser)
	//httpRouter.HandlerFunc(http.MethodPatch, "/v1/user/profile-icon", app.updateProfileIcon)
}

func profileHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.HandlerFunc(http.MethodGet, "/v1/user/profile/onboard-status", app.checkIfUserOnboarded)
	httpRouter.HandlerFunc(http.MethodPost, "/v1/user/profile/setup", app.onboardUser)
	httpRouter.HandlerFunc(http.MethodGet, "/v1/user/profile", app.getUserProfileDetail)
	httpRouter.HandlerFunc(http.MethodPatch, "/v1/user/profile/update", app.updateUserProfile)
}

func eventHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.HandlerFunc(http.MethodPost, "/v1/event/all", app.getAllEvents)
	httpRouter.HandlerFunc(http.MethodPost, "/v1/user/event", app.getUserEvents)
	httpRouter.HandlerFunc(http.MethodGet, "/v1/event/:eventId", app.getEventById)
	httpRouter.HandlerFunc(http.MethodPost, "/v1/event/create", app.createEvent)
	httpRouter.HandlerFunc(http.MethodPost, "/v1/event/join", app.joinEvent)
	httpRouter.HandlerFunc(http.MethodGet, "/v1/event-history/all", app.getEventHistory)
	httpRouter.HandlerFunc(http.MethodDelete, "/v1/event/quit/:eventId", app.quitEvent)
}

func messageCentreHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.HandlerFunc(http.MethodGet, "/v1/message-centre/main", app.getAllEvents)
}
