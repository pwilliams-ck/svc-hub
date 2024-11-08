package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/CloudKey-io/hostbill-svc/internal/services"
)

// TODO: Move API logic to internal/services/sso.go
// TODO: Comments
// TODO: Cleanup extra logs

func (app *application) vcdSession(r services.VcdSessionRequest) (w services.VcdSessionResponse, e error) {
	// Create POST request
	req, err := http.NewRequest("POST", r.URL, nil)
	if err != nil {
		app.logger.Error("failed to create Vcd session request", "error", err)
		return services.VcdSessionResponse{}, err
	}

	// Set Basic authentication header
	req.SetBasicAuth(r.Username, r.Password)

	// Disable TLS
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: transport,
	}

	// Send request
	res, err := client.Do(req)
	if err != nil {
		app.logger.Error("failed to send Vcd session request", "error", err)
		return services.VcdSessionResponse{}, err
	}
	defer res.Body.Close()

	// Check response status code
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		// Read the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			app.logger.Error("failed to read Vcd session response body", "error", err)
		} else {
			app.logger.Error("vcd session request failed", "status", res.Status, "body", string(body))
		}

		// Log the response headers
		var headers strings.Builder
		for name, values := range res.Header {
			headers.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")))
		}
		app.logger.Error("vcd session response headers", "headers", headers.String())

		return services.VcdSessionResponse{}, fmt.Errorf("vcd session request failed with status: %s", res.Status)
	}

	// Extract the session ID from the response headers
	sessionId := res.Header.Get("x-vcd-session")

	if sessionId == "" {
		app.logger.Error("failed to extract session Id from response headers")
		return services.VcdSessionResponse{}, fmt.Errorf("failed to extract session Id from response headers")
	}

	// Create the response struct
	responseStruct := services.VcdSessionResponse{
		SessionId: sessionId,
	}
	app.logger.Info("vcd session response", "response", responseStruct)

	return responseStruct, nil
}

func (app *application) createSsoHandler(w http.ResponseWriter, r *http.Request) {
}

func (app *application) updateSsoHandler(w http.ResponseWriter, r *http.Request) {
}

func (app *application) deleteSsoHandler(w http.ResponseWriter, r *http.Request) {
}
