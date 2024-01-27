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

	app.logInfo(fmt.Sprintf("Starting server in env=%s", app.config.env))
	err := server.ListenAndServeTLS("cert.crt", "cert.key")
	if err != nil {
		app.logInfo("error: %s, stopping server...", err)
		os.Exit(1)
	}
}
