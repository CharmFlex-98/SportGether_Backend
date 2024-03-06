package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) routes() http.Handler {
	httpRouter := httprouter.New()

	httpRouter.NotFound = http.HandlerFunc(app.notFound)
	httpRouter.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	websiteHandlerFunc(app, httpRouter)
	userHandlerFunc(app, httpRouter)
	eventHandlerFunc(app, httpRouter)
	profileHandlerFunc(app, httpRouter)
	messageCentreHandlerFunc(app, httpRouter)

	return app.recoverPanic(app.requiredMinAppVersion(app.authenticationHandler(httpRouter)))
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
	httpRouter.HandlerFunc(http.MethodPut, "/v1/user/activate", app.activateUser)
	httpRouter.HandlerFunc(http.MethodPut, "/v1/user/deregister-request", app.deregisterUserRequest)
	httpRouter.HandlerFunc(http.MethodDelete, "/v1/user/deregister", app.deactivateUser)
}

func profileHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.HandlerFunc(http.MethodGet, "/v1/user/profile/onboard-status", app.requiredActivatedUser(app.checkIfUserOnboarded))
	httpRouter.HandlerFunc(http.MethodPost, "/v1/user/profile/setup", app.requiredActivatedUser(app.onboardUser))
	httpRouter.HandlerFunc(http.MethodGet, "/v1/user/profile", app.requiredActivatedUser(app.getUserProfileDetail))
	httpRouter.HandlerFunc(http.MethodGet, "/v1/user/profile/other/:userId", app.requiredActivatedUser(app.getOtherUserProfileDetail))
	httpRouter.HandlerFunc(http.MethodPatch, "/v1/user/profile/update", app.requiredActivatedUser(app.updateUserProfile))
	httpRouter.HandlerFunc(http.MethodGet, "/v1/user/profile/mutual-info/:userId", app.requiredActivatedUser(app.getMutualEventInfo))
}

func eventHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.HandlerFunc(http.MethodPost, "/v1/event/all", app.requiredActivatedUser(app.getAllEvents))
	httpRouter.HandlerFunc(http.MethodPost, "/v1/user/event", app.requiredActivatedUser(app.getUserEvents))
	httpRouter.HandlerFunc(http.MethodGet, "/v1/event/:eventId", app.requiredActivatedUser(app.getEventById))
	httpRouter.HandlerFunc(http.MethodPost, "/v1/event/create", app.requiredActivatedUser(app.createEvent))
	httpRouter.HandlerFunc(http.MethodPatch, "/v1/event/update/:eventId", app.requiredActivatedUser(app.updateEvent))
	httpRouter.HandlerFunc(http.MethodPost, "/v1/event/join", app.requiredActivatedUser(app.joinEvent))
	httpRouter.HandlerFunc(http.MethodGet, "/v1/event-history/all", app.requiredActivatedUser(app.getEventHistory))
	httpRouter.HandlerFunc(http.MethodDelete, "/v1/event/quit/:eventId", app.requiredActivatedUser(app.quitEvent))
	httpRouter.HandlerFunc(http.MethodDelete, "/v1/event/delete/:eventId", app.requiredActivatedUser(app.deleteEvent))
	httpRouter.HandlerFunc(http.MethodPatch, "/v1/event/host/config/update", app.requiredActivatedUser(app.updateUserHostingConfigInfo))
	httpRouter.HandlerFunc(http.MethodPost, "/v1/event/host/config-init/", app.requiredActivatedUser(app.initHostingConfig))
}

func messageCentreHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.HandlerFunc(http.MethodGet, "/v1/message-centre/sports/all", app.requiredActivatedUser(app.getSportDetails))
	httpRouter.HandlerFunc(http.MethodGet, "/v1/message-centre/main", app.requiredActivatedUser(app.getMainMessage))
	httpRouter.HandlerFunc(http.MethodPost, "/v1/message-centre/register", app.requiredActivatedUser(app.registerFirebaseToken))
}

func websiteHandlerFunc(app *Application, httpRouter *httprouter.Router) {
	httpRouter.ServeFiles("/Sport-Gether/*filepath", http.Dir("static"))
}
