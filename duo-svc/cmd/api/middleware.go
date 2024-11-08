package main

import (
	"fmt"
	"net/http"
)

// Sets headers for incoming requests, we can set these as environment variables for the server
// config if needed.
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'")
		// Block pages from loading when they detect reflected cross-site scripting (XSS) attacks
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// Logs the details of each incoming request. It captures and logs the client's IP address,
// the protocol used, the HTTP method, and the requested URI.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.logger.Info("Request received", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

// This function handles unexpected errors during request processing. If a panic occurs, the function
// intercepts it, recovers normal execution flow, closes the connection, and sends an error response
// to the client, ensuring that the server can continue to handle other requests gracefully.
// If you don't close the connection after a panic, the client that sent the request might hang or
// wait indefinitely for a response. Basically, this closes the connection and sends the error message.
func (app *application) gracefulRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Defer ensures the following function is executed after the surrounding function.
		defer func() {
			// The recover built-in function allows a program to manage behavior of a
			// panicking goroutine. Executing a call to recover inside a deferred
			// function (but not any function called by it) stops the panicking sequence
			// by restoring normal execution and retrieves the error value passed to the
			// call of panic.o panic().
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
