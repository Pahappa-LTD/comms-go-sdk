package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
)

type CommsSdkInterface interface {
	GetApiKey() string
	GetUserName() string
	SetAuthenticated()
	GetApiURL() string
}

func ValidateCredentials(sdk CommsSdkInterface) (bool, error) {
	if sdk == nil {
		return false, errors.New("CommsSDK instance cannot be null")
	}

	if sdk.GetApiKey() == "" || sdk.GetUserName() == "" {
		return false, errors.New("either API Key or Username must be provided")
	}

	if !isValidCredential(sdk) {
		fmt.Println("                                                      _                    ")
		fmt.Println("  /\\     _|_ |_   _  ._ _|_ o  _  _. _|_ o  _  ._    |_ _. o |  _   _| | |")
		fmt.Println(" /--\\ |_| |_ | | (/_ | | |_ | (_ (_|  |_ | (_) | |   | (_| | | (/_ (_| o o ")
		fmt.Println("                                                                           ")
		fmt.Println()
		return false, errors.New("credentials validation failed")
	}

	fmt.Println("Credentials validated successfully.")
	fmt.Println("Validated using basic auth")
	sdk.SetAuthenticated()
	return true, nil
}

func isValidCredential(sdk CommsSdkInterface) bool {
	apiRequest := models.ApiRequest{}
	apiRequest.Method = "Balance"
	apiRequest.Userdata = models.UserData{UserName: sdk.GetUserName(), ApiKey: sdk.GetApiKey()}

	jsonBody, err := json.Marshal(apiRequest)
	if err != nil {
		fmt.Printf("Error marshalling request: %v\n", err)
		return false
	}

	resp, err := http.Post(sdk.GetApiURL(), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return false
	}

	var apiResponse models.ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
		return false
	}

	if apiResponse.Status == models.OK {
		fmt.Println("Credentials validated successfully.")
		return true
	} else {
		fmt.Printf("Error validating credentials: %s\n", apiResponse.Message)
		return false
	}
}
