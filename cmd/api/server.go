package main

import (
	"fmt"
	"log"
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
	f, err := os.Create("where_are_you.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Println(f.Name())
	certConfig := app.config.getCertConfig()
	app.logInfo("What is the cert? %s %s", certConfig.certPath, certConfig.certKey)
	err = server.ListenAndServeTLS(certConfig.certPath, certConfig.certKey)
	if err != nil {
		app.logInfo("error: %s, stopping server...", err)
		os.Exit(1)
	}
}
