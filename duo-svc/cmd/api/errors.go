package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Comment for testing.

// The logError() method is a generic helper for logging an error message along
// with the current request method and URL as attributes in the log entry.
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// The logTrace() method is a generic helper for logging a stack trace message along
// with the current request method and URL as attributes in the log entry.
func (app *application) logTrace(r *http.Request, err error) {
	trace := string(debug.Stack())

	app.logger.Error(err.Error(), "trace", trace)
}

// The errorResponse() method is a generic helper for sending JSON-formatted error
// messages to the client with a given status code. Note that we're using the any
// type for the message parameter, rather than just a string type, as this gives us
// more flexibility over the values that we can include in the response.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	// Write the response using the writeJSON() helper. If this happens to return an
	// error then log it, and fall back to sending the client an empty response with a
	// 500 Internal Server Error status code.
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// The serverErrorResponse() method a function that handles internal server errors. It logs the error, method,
// URI, and stack trace, then responds to the client with a 500 status code. This helps prevent server
// crashes and supports easier server-side debugging.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// The notFoundResponse() method will be used to send a 404 Not Found status code and
// JSON response to the client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// The methodNotAllowedResponse() method will be used to send a 405 Method Not Allowed
// status code and JSON response to the client.
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// Used to send a bad request response when reading JSON.
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// Used to send a failed validation response when reading JSON.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
