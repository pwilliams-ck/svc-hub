package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance.
	mux := http.NewServeMux()

	// Register the relevant methods, URL patterns and handler functions for our
	// endpoints using the HandleFunc() method.
	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("POST /v1/sso", app.createSsoHandler)
	mux.HandleFunc("GET /v1/sso/{id}", app.showSsoHandler)

	// Return mux router with middleware.
	return app.gracefulRecovery(app.logRequest((commonHeaders(mux))))
}
