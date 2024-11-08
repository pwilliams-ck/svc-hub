package main

import "net/http"

func (app *application) routes() http.Handler {
	// The ServeMux is a type provided by the net/http package in Go and acts
	// as an HTTP request multiplexer or router.
	mux := http.NewServeMux()

	// Healthcheck endpoint
	mux.HandleFunc("GET /api/v1/healthcheck", app.healthcheckHandler)

	// SSO endpoints
	// We should only need POST, PUT, and DELETE endpoints, this is for updating SAML
	// on both VCD and Duo.
	mux.HandleFunc("PUT /api/v1/sso", app.updateSsoHandler)
	// Need to add in "archiving" functionality, then we would eventually
	// delete resources.
	mux.HandleFunc("DELETE /api/v1/sso", app.deleteSsoHandler)

	// Zerto endpoints
	// We should only need POST, PUT, and  DELETE endpoints, users can still
	// manage and "get" data from their self-serve portal. Unless we want
	// GET data showing up in Hostbill.
	mux.HandleFunc("POST /api/v1/veeam", app.createVeeamHandler)
	mux.HandleFunc("PUT /api/v1/veeam", app.updateVeeamHandler)
	// Need to add in "archiving" functionality, then we would eventually
	// delete resources.
	mux.HandleFunc("DELETE /api/v1/veeam", app.deleteVeeamHandler)

	// Veeam endpoints
	// We should only need POST, PUT, and  DELETE endpoints, users can still
	// manage and "get" data from their self-serve portal. Unless we want
	// GET data showing up in Hostbill.
	mux.HandleFunc("POST /api/v1/zerto", app.createZertoHandler)
	mux.HandleFunc("PUT /api/v1/zerto", app.updateZertoHandler)
	// Need to add in "archiving" functionality, then we would eventually
	// delete resources.
	mux.HandleFunc("DELETE /api/v1/zerto", app.deleteZertoHandler)

	return app.gracefulRecovery(app.logRequest((commonHeaders(mux))))
}
