package services

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// VcdSessionRequest represents the request body for obtaining a zerto session token
type VcdSessionRequest struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// VcdSessionResponse represents the response body for obtaining a zerto session token
type VcdSessionResponse struct {
	SessionId string `json:"sessionId"`
}

func vcdSession(r VcdSessionRequest) (w VcdSessionResponse, e error) {
	// Create POST request
	req, err := http.NewRequest("POST", r.URL, nil)
	if err != nil {
		return VcdSessionResponse{}, err
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
		return VcdSessionResponse{}, err
	}
	defer res.Body.Close()

	// Check response status code
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		// Read the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return VcdSessionResponse{}, fmt.Errorf("vcd session request failed with status: %s", res.Status)
		} else {
			return VcdSessionResponse{}, fmt.Errorf("vcd session request failed with body: %s", body)
		}

		// Log the response headers
		var headers strings.Builder
		for name, values := range res.Header {
			headers.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")))
		}

		return VcdSessionResponse{}, fmt.Errorf("vcd session request failed with status: %s", res.Status)
	}

	// Extract the session ID from the response headers
	sessionId := res.Header.Get("x-vcd-session")

	if sessionId == "" {
		return VcdSessionResponse{}, fmt.Errorf("failed to extract session Id from response headers")
	}

	// Create the response struct
	responseStruct := VcdSessionResponse{
		SessionId: sessionId,
	}

	return responseStruct, nil
}
