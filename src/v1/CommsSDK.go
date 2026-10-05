package v1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/exceptions"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/utils"
)

const (
	LiveApiURL    = "https://comms.egosms.co/api/v1/json"
	SandboxApiURL = "https://comms-test.pahappa.net/api/v1/json"
)

// API_URL is the global API endpoint, followed by instances created through Authenticate.
//
// Deprecated: each instance now carries its own endpoint. Use Live or Sandbox instead.
var API_URL = LiveApiURL

type CommsSDK struct {
	apiKey          string
	userName        string
	senderId        string
	isAuthenticated bool
	// apiUrl is this instance's own endpoint, or empty when it follows API_URL.
	apiUrl string
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

// GetApiURL returns the endpoint this instance talks to: its own, or API_URL if it follows it.
func (sdk *CommsSDK) GetApiURL() string {
	if sdk.apiUrl != "" {
		return sdk.apiUrl
	}
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

// UseSandBox switches the global endpoint to the sandbox. This also affects existing instances
// created through Authenticate.
//
// Deprecated: use Sandbox, which binds one instance without changing global state.
func UseSandBox() {
	API_URL = SandboxApiURL
}

// UseLiveServer switches the global endpoint to the live server. This also affects existing
// instances created through Authenticate.
//
// Deprecated: use Live, which binds one instance without changing global state.
func UseLiveServer() {
	API_URL = LiveApiURL
}

// Logger is what the SDK writes its log messages to. *slog.Logger (Go 1.21+) and hclog.Logger
// satisfy it as they are.
type Logger = utils.Logger

// SetLogger sets the logger the SDK writes to, for example slog.Default(). The SDK is silent until a
// logger is set; passing nil makes it silent again.
func SetLogger(l Logger) {
	utils.SetLogger(l)
}

// Live creates an instance bound to the live server and verifies the credentials.
// Rejected credentials still return an instance; check IsAuthenticated for the result.
//
// It returns a *exceptions.CommsValidationError if the user name or API key is empty.
func Live(userName string, apiKey string) (*CommsSDK, error) {
	return create(userName, apiKey, LiveApiURL)
}

// Sandbox creates an instance bound to the sandbox server (for testing) and verifies the credentials.
// Rejected credentials still return an instance; check IsAuthenticated for the result.
//
// It returns a *exceptions.CommsValidationError if the user name or API key is empty.
func Sandbox(userName string, apiKey string) (*CommsSDK, error) {
	return create(userName, apiKey, SandboxApiURL)
}

func create(userName string, apiKey string, apiUrl string) (*CommsSDK, error) {
	sdk := &CommsSDK{
		userName: userName,
		apiKey:   apiKey,
		senderId: "EgoSMS",
		apiUrl:   apiUrl,
	}

	isValid, err := utils.ValidateCredentials(sdk)
	var authErr *exceptions.CommsAuthenticationError
	if err != nil && !errors.As(err, &authErr) {
		return nil, err
	}

	sdk.isAuthenticated = isValid
	return sdk, nil
}

// Authenticate creates a new instance that follows the global API_URL and verifies the credentials.
//
// It returns a *exceptions.CommsValidationError if the user name or API key is empty, and a
// *exceptions.CommsAuthenticationError if the server rejected the credentials.
//
// Deprecated: use Live or Sandbox, which bind the instance to one endpoint.
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

// SendSMSFull sends an SMS to one or more numbers.
//
// It returns a *exceptions.CommsValidationError for bad input, and a *exceptions.CommsApiError when
// no response is received, the server reports Failed, or the status is not recognised.
func (sdk *CommsSDK) SendSMSFull(numbers interface{}, message string, senderId string, priority models.MessagePriority) (bool, error) {
	apiResponse, err := sdk.QuerySendSMSFull(numbers, message, senderId, priority)
	if err != nil {
		return false, err
	}

	if apiResponse == nil {
		utils.Log().Error("Failed to get a response from the server.")
		return false, &exceptions.CommsApiError{Message: "failed to get a response from the server"}
	}

	switch apiResponse.Status {
	case models.OK:
		utils.Log().Info("SMS sent successfully.", "messageFollowUpUniqueCode", apiResponse.MessageFollowUpCode)
		return true, nil
	case models.Failed:
		utils.Log().Error("Failed: "+apiResponse.Message, "message", apiResponse.Message)
		return false, &exceptions.CommsApiError{Message: fmt.Sprintf("failed: %s ", apiResponse.Message)}
	default:
		return false, &exceptions.CommsApiError{Message: fmt.Sprintf("unexpected response status: %v", apiResponse.Status)}
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

// QuerySendSMSFull is the same as SendSMSFull but returns the full ApiResponse object, or nil on error.
//
// It returns a *exceptions.CommsValidationError for bad input.
func (sdk *CommsSDK) QuerySendSMSFull(numbers interface{}, message string, senderId string, priority models.MessagePriority) (*models.ApiResponse, error) {
	if sdk.notAuthenticated() {
		return nil, nil
	}

	var numberSlice []string
	switch v := numbers.(type) {
	case string:
		numberSlice = []string{v}
	case []string:
		numberSlice = v
	default:
		return nil, &exceptions.CommsValidationError{Message: "numbers must be a string or a slice of strings"}
	}

	if len(numberSlice) == 0 {
		return nil, &exceptions.CommsValidationError{Message: "numbers list cannot be empty"}
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return nil, &exceptions.CommsValidationError{Message: "message cannot be empty"}
	}

	if len(message) == 1 {
		return nil, &exceptions.CommsValidationError{Message: "message cannot be a single character"}
	}

	senderId = strings.TrimSpace(senderId)
	if senderId == "" {
		senderId = sdk.senderId
	}

	if len(senderId) > 11 {
		utils.Log().Warn("Warning: Sender ID length exceeds 11 characters. Some networks may truncate or reject messages.", "senderId", senderId)
	}

	validatedNumbers := utils.ValidateNumbers(numberSlice)

	if len(validatedNumbers) == 0 {
		utils.Log().Error("No valid phone numbers provided. Please check inputs.")
		return nil, &exceptions.CommsValidationError{Message: "no valid phone numbers provided. Please check inputs"}
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

	return sdk.SendCustomSMS(messageModels)
}

// SendCustomSMS sends a custom-built list of MessageModel values. It returns nil (and no error) when
// the request could not be completed; the reason is logged.
func (sdk *CommsSDK) SendCustomSMS(messages []models.MessageModel) (*models.ApiResponse, error) {
	if sdk.notAuthenticated() {
		return nil, nil
	}

	apiRequest := models.ApiRequest{
		Method:      "SendSms",
		Userdata:    models.UserData{UserName: sdk.userName, ApiKey: sdk.apiKey},
		MessageData: messages,
		WalletType:  models.Local,
	}

	apiResponse, err := utils.Post(apiRequest, sdk.GetApiURL())
	if err != nil {
		utils.Log().Error("Failed to send SMS.", "error", err)
		return nil, nil
	}
	return apiResponse, nil
}

// notAuthenticated re-verifies the credentials if needed and reports whether the instance is still
// unauthenticated.
func (sdk *CommsSDK) notAuthenticated() bool {
	if sdk.isAuthenticated {
		return false
	}
	utils.Log().Warn("SDK is not authenticated. Please authenticate before performing actions.")
	utils.Log().Warn("Attempting to re-authenticate with provided credentials...")
	isValid, err := utils.ValidateCredentials(sdk)
	if err != nil || !isValid {
		return true
	}
	sdk.isAuthenticated = true
	return false
}

// QueryBalance is the same as GetBalance but returns the full ApiResponse object.
// Queries the Local wallet; use QueryBalanceFull to query a specific wallet.
func (sdk *CommsSDK) QueryBalance() (*models.ApiResponse, error) {
	return sdk.QueryBalanceFull(models.Local)
}

// QueryBalanceFull is the same as QueryBalance but lets you specify which wallet to query.
// It returns nil (and no error) if the credentials could not be verified.
//
// It returns a *exceptions.CommsApiError if the balance request fails.
func (sdk *CommsSDK) QueryBalanceFull(walletType models.WalletType) (*models.ApiResponse, error) {
	if walletType == "" {
		walletType = models.Local
	}

	if sdk.notAuthenticated() {
		return nil, nil
	}

	apiRequest := models.ApiRequest{
		Method:     "Balance",
		Userdata:   models.UserData{UserName: sdk.userName, ApiKey: sdk.apiKey},
		WalletType: walletType,
	}

	apiResponse, err := utils.Post(apiRequest, sdk.GetApiURL())
	if err != nil {
		return nil, &exceptions.CommsApiError{Message: fmt.Sprintf("failed to get balance: %v", err), Cause: err}
	}
	return apiResponse, nil
}

// GetBalance returns the Local wallet balance; use GetBalanceFull to query a specific wallet.
func (sdk *CommsSDK) GetBalance() (*float64, error) {
	return sdk.GetBalanceFull(models.Local)
}

// GetBalanceFull is the same as GetBalance but lets you specify which wallet to query.
//
// It returns a *exceptions.CommsApiError if the balance request fails.
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
	return fmt.Sprintf("SDK(%s, %s, %s)", sdk.userName, sdk.senderId, sdk.GetApiURL())
}
