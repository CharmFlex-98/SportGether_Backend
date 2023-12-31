package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

// Logging
func (app *Application) logInfo(message string, args ...any) {
	app.logger.Info(message, args)
}

func (app *Application) logError(error error, r *http.Request) {
	app.logger.Error(error.Error(), "METHOD", r.Method)
}

func (app *Application) logWarning(message string, args ...any) {
	app.logger.Error(message, args)
}

type responseData map[string]any
type responseHeader map[string]string

// Json response
func (app *Application) writeResponse(w http.ResponseWriter, content any, code int, headers responseHeader) error {
	if content == nil {
		w.WriteHeader(code)
		return nil
	}

	res, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header().Set(value, headers[key])
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(res)

	return nil
}

func (app *Application) writeError(w http.ResponseWriter, r *http.Request, code int, errorCode int, message any) {
	errContent := map[string]any{
		"errorCode": errorCode,
		"message":   message,
	}
	data := responseData{
		"error": errContent,
	}

	err := app.writeResponse(w, data, code, nil)
	if err != nil {
		app.logError(err, r)
		w.WriteHeader(500)
	}
}

func (app *Application) readRequest(r *http.Request, input any) error {
	reqBody := r.Body

	err := json.NewDecoder(reqBody).Decode(&input)
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) readParam(paramName string, r *http.Request) (*int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	value, err := strconv.ParseInt(params.ByName(paramName), 10, 64)
	if err != nil {
		return nil, err
	}

	return &value, nil
}
