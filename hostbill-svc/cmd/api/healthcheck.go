package main

import (
	"encoding/json"
	"net/http"
)

// `healthcheckHandler` is an HTTP handler providing a health check endpoint for the application.
// It returns a JSON response with the current application status, environment, and version.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	// TODO: JSON Marshalling
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Error(err.Error())
		// Errors during response preparation are logged and result in a 500 status response to the client.
		http.Error(w, "JSON marshalling error.", http.StatusInternalServerError)
		return
	}

	// Append new line to JSON, for readability in terminals.
	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")

	w.Write(js)
}
