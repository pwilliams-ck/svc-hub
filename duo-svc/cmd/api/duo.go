package main

import (
	"net/http"
)

// For now we simply return a plain-text placeholder response.
func (app *application) createSsoHandler(w http.ResponseWriter, r *http.Request) {
}

// For now we simply return a plain-text placeholder response.
func (app *application) updateSsoHandler(w http.ResponseWriter, r *http.Request) {
}

// "GET /v1/snips/:id" endpoint. For now, we retrieve the "id" parameter from the
// current URL and include it in a placeholder response.
func (app *application) showSsoHandler(w http.ResponseWriter, r *http.Request) {
}
