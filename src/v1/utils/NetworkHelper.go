package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/exceptions"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
)

// Post sends the request as JSON and decodes the response.
//
// It returns a *exceptions.CommsApiError if the server cannot be reached, or replies with a body that
// is not JSON. A JSON reply with Status "Failed" is returned as a response, not an error.
func Post(apiRequest models.ApiRequest, apiUrl string) (*models.ApiResponse, error) {
	jsonBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, &exceptions.CommsApiError{Message: fmt.Sprintf("could not encode request: %v", err), Cause: err}
	}

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, &exceptions.CommsApiError{Message: fmt.Sprintf("could not complete request to %s: %v", apiUrl, err), Cause: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &exceptions.CommsApiError{Message: fmt.Sprintf("could not complete request to %s: %v", apiUrl, err), Cause: err}
	}

	var apiResponse models.ApiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		message := fmt.Sprintf("HTTP %d", resp.StatusCode)
		if len(body) > 0 {
			message += ": " + string(body)
		}
		return nil, &exceptions.CommsApiError{Message: message, Cause: err}
	}
	return &apiResponse, nil
}
