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
)

// VeeamSessionRequest represents the request body for obtaining a Veeam session token
type VeeamSessionRequest struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// VeeamSessionResponse represents the response body for obtaining a Veeam session token
type VeeamSessionResponse struct {
	SessionId string `json:"sessionId"`
}

func (app *application) veeamSession(v VeeamSessionRequest) (VeeamSessionResponse, error) {
	// Create POST request
	req, err := http.NewRequest("POST", v.URL, nil)
	if err != nil {
		app.logger.Error("Failed to create Veeam session request", "error", err)
		return VeeamSessionResponse{}, err
	}

	// Set Basic authentication header
	req.SetBasicAuth(v.Username, v.Password)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// Send request
	client := &http.Client{
		Transport: transport,
	}

	app.logger.Info("stuf", "things", v)

	res, err := client.Do(req)
	if err != nil {
		app.logger.Error("Failed to send Veeam session request", "error", err)
		return VeeamSessionResponse{}, err
	}
	defer res.Body.Close()

	// Check response status code
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		// Read the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			app.logger.Error("Failed to read Veeam session response body", "error", err)
		} else {
			app.logger.Error("Veeam session request failed", "status", res.Status, "body", string(body))
		}

		// Log the response headers
		var headers strings.Builder
		for name, values := range res.Header {
			headers.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")))
		}
		app.logger.Error("Veeam session response headers", "headers", headers.String())

		return VeeamSessionResponse{}, fmt.Errorf("veeam session request failed with status: %s", res.Status)
	}

	// Extract the session ID from the response headers
	sessionId := res.Header.Get("X-RestSvcSessionId")

	if sessionId == "" {
		app.logger.Error("Failed to extract session ID from response headers")
		return VeeamSessionResponse{}, fmt.Errorf("failed to extract session ID from response headers")
	}

	// Create the response struct
	responseStruct := VeeamSessionResponse{
		SessionId: sessionId,
	}
	app.logger.Info("Veeam session response", "response", responseStruct)

	return responseStruct, nil
}

// VeeamCreateOrgRequest represents the request body for creating a Veeam organization
type VeeamCreateOrgRequest struct {
	OrganizationName       string      `json:"OrganizationName"`
	BackupServerUid        string      `json:"BackupServerUid"`
	RepositoryUid          string      `json:"RepositoryUid"`
	QuotaGb                json.Number `json:"QuotaGb"`
	RepositoryFriendlyName string      `json:"RepositoryFriendlyName"`
	JobSchedulerType       string      `json:"JobSchedulerType"`
	HighPriorityJob        bool        `json:"HighPriorityJob"`
	HostUid                string      `json:"HostUid"`
}

// VeeamCreateOrgResponse represents the response body for creating a Veeam organization
type VeeamCreateOrgResponse struct {
	Message string `json:"message"`
}

func (app *application) createVeeamHandler(w http.ResponseWriter, r *http.Request) {
	var veeamData VeeamCreateOrgRequest

	// Read and parse the request body
	err := app.readJSON(w, r, &veeamData)
	if err != nil {
		app.logger.Error("Failed to read request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.logger.Info("Received request", "body", veeamData)

	// Obtain a Veeam session token
	// Need to use Github Secrets here
	sessionConfig := VeeamSessionRequest{
		URL:      os.Getenv("VEEAM_URL") + os.Getenv("VEEAM_SESSION"),
		Username: os.Getenv("VEEAM_USERNAME"),
		Password: os.Getenv("VEEAM_PASSWORD"),
	}
	sessionToken, err := app.veeamSession(sessionConfig)

	app.logger.Info("Received request", "sessionToken", sessionToken)

	if err != nil {
		app.logger.Error("Failed to obtain Veeam session token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Strip whitespace from the OrganizationName field
	veeamData.OrganizationName = strings.ReplaceAll(veeamData.OrganizationName, " ", "")

	// Create the request body for the Veeam API
	veeamBody := map[string]interface{}{
		"OrganizationName":       veeamData.OrganizationName,
		"BackupServerUid":        os.Getenv("VEEAM_BACKUP_SERVER_UID"),
		"RepositoryUid":          os.Getenv("VEEAM_REPOSITORY_UID"),
		"QuotaGb":                veeamData.QuotaGb,
		"RepositoryFriendlyName": "Testing Boii",
		"JobSchedulerType":       "Full",
		"HighPriorityJob":        false,
		"HostUid":                os.Getenv("VEEAM_HOST_UID"),
	}

	// Log the JSON data being sent
	jsonData, _ := json.Marshal(veeamBody)
	app.logger.Info("Request JSON data", "data", string(jsonData))

	// Create the POST request to the Veeam API
	veeamOrgURL := os.Getenv("VEEAM_URL") + os.Getenv("VEEAM_CREATE_ORG")

	req, err := http.NewRequest("POST", veeamOrgURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.logger.Error("Failed to create Veeam API request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RestSvcSessionId", sessionToken.SessionId)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// Send the request to the Veeam API
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		app.logger.Error("Failed to send Veeam API request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Log the response body
	responseBody, _ := io.ReadAll(resp.Body)
	app.logger.Info("Response body", "data", string(responseBody))

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		app.logger.Error("Veeam API request failed", "status", resp.StatusCode)
		http.Error(w, "Failed to create Veeam organization", resp.StatusCode)
		return
	}

	// Return a success response
	response := map[string]string{"message": "Veeam organization created successfully"}
	if err := app.writeJSON(w, http.StatusOK, response, nil); err != nil {
		app.logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (app *application) updateVeeamHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Add Veeam storage...")
}

func (app *application) deleteVeeamHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.logger.Error(err.Error())
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Delete Veeam storage with ID %d...\n", id)
}
