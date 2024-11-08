package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) Authenticate(w http.ResponseWriter, r *http.Request) {
	// Define a struct to hold the request payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Read the JSON payload from the request body and decode it into the requestPayload struct
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		// If there's an error reading the JSON payload, send an error response
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate the user against the database using the provided email
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		// If the user is not found or there's an error, send an "invalid credentials" error response
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// Check if the provided password matches the user's password
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		// If there's an error or the password doesn't match, send an "invalid credentials" error response
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// Log authentcation
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// Create a success response payload
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	// Send the success response with the user data
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *application) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-svc/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
