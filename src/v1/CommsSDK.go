package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/utils"
)

var API_URL = "https://comms.egosms.co/api/v1/json/"

type CommsSDK struct {
	apiKey          string
	userName        string
	senderId        string
	isAuthenticated bool
}

func (sdk *CommsSDK) GetApiKey() string {
	return sdk.apiKey
}

func (sdk *CommsSDK) GetUserName() string {
	return sdk.userName
}

func (sdk *CommsSDK) GetSenderId() string {
	return sdk.senderId
}

func (sdk *CommsSDK) IsAuthenticated() bool {
	return sdk.isAuthenticated
}

func (sdk *CommsSDK) SetAuthenticated() {
	sdk.isAuthenticated = true
}

func (sdk *CommsSDK) GetApiURL() string {
	return API_URL
}

func (sdk *CommsSDK) SendSMS(numbers interface{}, message string) (bool, error) {
	return sdk.SendSMSFull(numbers, message, sdk.senderId, models.HIGH)
}

func (sdk *CommsSDK) SendSMSWithSenderId(numbers interface{}, message string, senderId string) (bool, error) {
	return sdk.SendSMSFull(numbers, message, senderId, models.HIGH)
}

func (sdk *CommsSDK) SendSMSWithPriority(numbers interface{}, message string, priority models.MessagePriority) (bool, error) {
	return sdk.SendSMSFull(numbers, message, sdk.senderId, priority)
}

func UseSandBox() {
	API_URL = "https://comms-test.pahappa.net/api/v1/json"
}

func UseLiveServer() {
	API_URL = "https://comms.egosms.co/api/v1/json"
}

func Authenticate(userName string, apiKey string) (*CommsSDK, error) {
	sdk := &CommsSDK{
		userName: userName,
		apiKey:   apiKey,
		senderId: "EgoSMS",
	}

	isValid, err := utils.ValidateCredentials(sdk)
	if err != nil {
		return nil, err
	}

	sdk.isAuthenticated = isValid
	return sdk, nil
}

func (sdk *CommsSDK) WithSenderId(senderId string) *CommsSDK {
	sdk.senderId = senderId
	return sdk
}

func (sdk *CommsSDK) SendSMSFull(numbers interface{}, message string, senderId string, priority models.MessagePriority) (bool, error) {
	apiResponse, err := sdk.QuerySendSMSFull(numbers, message, senderId, priority)
	if err != nil {
		return false, err
	}

	if apiResponse == nil {
		fmt.Println("Failed to get a response from the server.")
		return false, fmt.Errorf("failed to get a response from the server")
	}

	switch apiResponse.Status {
	case models.OK:
		fmt.Println("SMS sent successfully.")
		if apiResponse.MessageFollowUpCode != "" {
			fmt.Printf("MessageFollowUpUniqueCode: %s\n", apiResponse.MessageFollowUpCode)
		}
		return true, nil
	case models.Failed:
		fmt.Printf("Failed: %s\n", apiResponse.Message)
		return false, fmt.Errorf("failed: %s ", apiResponse.Message)
	default:
		return false, fmt.Errorf("unexpected response status: %v", apiResponse.Status)
	}
}

// QuerySendSMS is the same as SendSMS but returns the full ApiResponse object
func (sdk *CommsSDK) QuerySendSMS(numbers interface{}, message string) (*models.ApiResponse, error) {
	return sdk.QuerySendSMSFull(numbers, message, sdk.senderId, models.HIGH)
}

// QuerySendSMSWithSenderId is the same as SendSMSWithSenderId but returns the full ApiResponse object
func (sdk *CommsSDK) QuerySendSMSWithSenderId(numbers interface{}, message string, senderId string) (*models.ApiResponse, error) {
	return sdk.QuerySendSMSFull(numbers, message, senderId, models.HIGH)
}

// QuerySendSMSWithPriority is the same as SendSMSWithPriority but returns the full ApiResponse object
func (sdk *CommsSDK) QuerySendSMSWithPriority(numbers interface{}, message string, priority models.MessagePriority) (*models.ApiResponse, error) {
	return sdk.QuerySendSMSFull(numbers, message, sdk.senderId, priority)
}

// QuerySendSMSFull is the same as SendSMSFull but returns the full ApiResponse object
func (sdk *CommsSDK) QuerySendSMSFull(numbers interface{}, message string, senderId string, priority models.MessagePriority) (*models.ApiResponse, error) {
	if !sdk.isAuthenticated {
		fmt.Fprintf(os.Stderr, "SDK is not authenticated. Please authenticate before performing actions.\n")
		fmt.Fprintf(os.Stderr, "Attempting to re-authenticate with provided credentials...\n")
		isValid, err := utils.ValidateCredentials(sdk)
		if err != nil || !isValid {
			return nil, nil
		}
		sdk.isAuthenticated = true
	}

	var numberSlice []string
	switch v := numbers.(type) {
	case string:
		numberSlice = []string{v}
	case []string:
		numberSlice = v
	default:
		return nil, errors.New("numbers must be a string or a slice of strings")
	}

	if len(numberSlice) == 0 {
		return nil, errors.New("numbers list cannot be empty")
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	if len(message) == 1 {
		return nil, errors.New("message cannot be a single character")
	}

	senderId = strings.TrimSpace(senderId)
	if senderId == "" {
		senderId = sdk.senderId
	}

	if len(senderId) > 11 {
		fmt.Println("Warning: Sender ID length exceeds 11 characters. Some networks may truncate or reject messages.")
	}

	validatedNumbers := utils.ValidateNumbers(numberSlice)

	if len(validatedNumbers) == 0 {
		fmt.Fprintf(os.Stderr, "No valid phone numbers provided. Please check inputs.\n")
		return nil, fmt.Errorf("no valid phone numbers provided. Please check inputs")
	}

	var messageModels []models.MessageModel
	for _, number := range validatedNumbers {
		messageModels = append(messageModels, models.MessageModel{
			Number:   number,
			Message:  message,
			SenderId: senderId,
			Priority: priority,
		})
	}

	apiRequest := models.ApiRequest{
		Method:      "SendSms",
		Userdata:    models.UserData{UserName: sdk.userName, ApiKey: sdk.apiKey},
		MessageData: messageModels,
		WalletType:  models.Local,
	}

	jsonBody, err := json.Marshal(apiRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send SMS: %v\n", err)
		return nil, nil
	}

	resp, err := http.Post(API_URL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send SMS: %v\n", err)
		return nil, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send SMS: %v\n", err)
		return nil, nil
	}

	var apiResponse models.ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send SMS: %v\n", err)
		return nil, nil
	}

	return &apiResponse, nil
}

// QueryBalance is the same as GetBalance but returns the full ApiResponse object.
// Queries the Local wallet; use QueryBalanceFull to query a specific wallet.
func (sdk *CommsSDK) QueryBalance() (*models.ApiResponse, error) {
	return sdk.QueryBalanceFull(models.Local)
}

// QueryBalanceFull is the same as QueryBalance but lets you specify which wallet to query.
func (sdk *CommsSDK) QueryBalanceFull(walletType models.WalletType) (*models.ApiResponse, error) {
	if walletType == "" {
		walletType = models.Local
	}

	if !sdk.isAuthenticated {
		fmt.Fprintf(os.Stderr, "SDK is not authenticated. Please authenticate before performing actions.\n")
		fmt.Fprintf(os.Stderr, "Attempting to re-authenticate with provided credentials...\n")
		isValid, err := utils.ValidateCredentials(sdk)
		if err != nil || !isValid {
			return nil, nil
		}
		sdk.isAuthenticated = true
	}

	apiRequest := models.ApiRequest{
		Method:     "Balance",
		Userdata:   models.UserData{UserName: sdk.userName, ApiKey: sdk.apiKey},
		WalletType: walletType,
	}

	jsonBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	resp, err := http.Post(API_URL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	var apiResponse models.ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	return &apiResponse, nil
}

// GetBalance returns the Local wallet balance; use GetBalanceFull to query a specific wallet.
func (sdk *CommsSDK) GetBalance() (*float64, error) {
	return sdk.GetBalanceFull(models.Local)
}

// GetBalanceFull is the same as GetBalance but lets you specify which wallet to query.
func (sdk *CommsSDK) GetBalanceFull(walletType models.WalletType) (*float64, error) {
	response, err := sdk.QueryBalanceFull(walletType)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	if response.Balance == nil {
		return nil, nil
	}

	return response.Balance, nil
}

func (sdk *CommsSDK) String() string {
	return fmt.Sprintf("SDK(%s => %s)", sdk.userName, sdk.apiKey)
}
