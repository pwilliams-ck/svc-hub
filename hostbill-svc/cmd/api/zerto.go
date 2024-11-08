package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/CloudKey-io/hostbill-svc/internal/services"
)

// TODO: Move API logic to internal/services/zerto.go
// TODO: Comments
// TODO: Cleanup extra logs

func (app *application) zertoSession(r services.ZertoSessionRequest) (w services.ZertoSessionResponse, e error) {
	// Create POST request
	req, err := http.NewRequest("POST", r.URL, nil)
	if err != nil {
		app.logger.Error("Failed to create Zerto session request", "error", err)
		return services.ZertoSessionResponse{}, err
	}

	// Set Basic authentication header
	req.SetBasicAuth(r.Username, r.Password)

	// Disable TLS
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !app.config.useTLS,
		},
	}
	client := &http.Client{
		Transport: transport,
	}

	// Send request
	res, err := client.Do(req)
	if err != nil {
		app.logger.Error("failed to send Zerto session request", "error", err)
		return services.ZertoSessionResponse{}, err
	}
	defer res.Body.Close()

	// Check response status code
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		// Read the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			app.logger.Error("failed to read Zerto session response body", "error", err)
		} else {
			app.logger.Error("zerto session request failed", "status", res.Status, "body", string(body))
		}

		// Log the response headers
		var headers strings.Builder
		for name, values := range res.Header {
			headers.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")))
		}
		app.logger.Error("zerto session response headers", "headers", headers.String())

		return services.ZertoSessionResponse{}, fmt.Errorf("zerto session request failed with status: %s", res.Status)
	}

	// Extract the session ID from the response headers
	sessionId := res.Header.Get("x-zerto-session")

	if sessionId == "" {
		app.logger.Error("failed to extract session Id from response headers")
		return services.ZertoSessionResponse{}, fmt.Errorf("failed to extract session Id from response headers")
	}

	// Create the response struct
	responseStruct := services.ZertoSessionResponse{
		SessionId: sessionId,
	}
	app.logger.Info("zerto session response", "response", responseStruct)

	return responseStruct, nil
}

func (app *application) createZertoHandler(w http.ResponseWriter, r *http.Request) {
	var zertoData services.ZertoCreateOrgRequest

	// Read and parse the request body
	err := app.readJSON(w, r, &zertoData)
	if err != nil {
		app.logger.Error("Failed to read request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.logger.Info("Received request", "body", zertoData)

	// Obtain a zerto session token
	sessionConfig := services.ZertoSessionRequest{
		URL:      os.Getenv("ZERTO_URL") + os.Getenv("ZERTO_SESSION_START"),
		Username: os.Getenv("ZERTO_USER"),
		Password: os.Getenv("ZERTO_PW"),
	}
	sessionToken, err := app.zertoSession(sessionConfig)

	app.logger.Info("Received request", "sessionToken", sessionToken)

	if err != nil {
		app.logger.Error("Failed to obtain zerto session token", "URL", sessionConfig.URL)
		app.logger.Error("Failed to obtain zerto session token", "username", sessionConfig.Username)
		app.logger.Error("Failed to obtain zerto session token", "password", sessionConfig.Password)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create the request body for the Zerto API
	zertoBody := map[string]interface{}{
		"Name":          zertoData.Name,
		"CrmIdentifier": zertoData.CrmIdentifier,
		"TenantInfo": map[string]interface{}{
			"CompanyName":             zertoData.Name,
			"DomainName":              zertoData.Name + ".cloudkey.io",
			"Country":                 zertoData.TenantInfo.Country,
			"State":                   zertoData.TenantInfo.State,
			"PostalCode":              zertoData.TenantInfo.PostalCode,
			"IsMultiCloudProductType": zertoData.TenantInfo.IsMultiCloudProductType,
		},
	}

	// Log the JSON services being sent
	jsonData, _ := json.Marshal(zertoBody)
	app.logger.Info("Request JSON services", "services", string(jsonData))

	// Create the POST request to the Zerto API
	zertoOrgURL := os.Getenv("ZERTO_URL") + os.Getenv("ZERTO_CREATE_ZORG")

	req, err := http.NewRequest("POST", zertoOrgURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.logger.Error("Failed to create zerto API request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-zerto-session", sessionToken.SessionId)

	// Disable TLS
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !app.config.useTLS,
		},
	}
	client := &http.Client{
		Transport: transport,
	}

	// Send the request to the Zerto API
	res, err := client.Do(req)
	if err != nil {
		app.logger.Error("Failed to send Zerto API request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	// Log response body
	responseBody, _ := io.ReadAll(res.Body)
	app.logger.Info("Response body", "services", string(responseBody))

	// Check the response status code
	if res.StatusCode != http.StatusOK {
		app.logger.Error("Zerto API request failed", "status", res.StatusCode)
		http.Error(w, "Failed to create zerto organization", res.StatusCode)
		return
	}

	// Return a success response
	response := map[string]string{"message": "Zerto organization created successfully"}
	if err := app.writeJSON(w, http.StatusOK, response, nil); err != nil {
		app.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (app *application) updateZertoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Add Zerto storage...")
}

func (app *application) deleteZertoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.logger.Error(err.Error())
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Delete Zerto storage with ID %d...\n", id)
}
