package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func (app *Application) serve() {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	app.logInfo("new commit, test here 2")
	app.logInfo("Starting server...")
	err := server.ListenAndServe()
	if err != nil {
		app.logInfo("stopping server...")
		os.Exit(1)
	}
}
