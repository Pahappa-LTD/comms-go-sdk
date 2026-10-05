package utils

import (
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/exceptions"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
)

type CommsSdkInterface interface {
	GetApiKey() string
	GetUserName() string
	SetAuthenticated()
	GetApiURL() string
}

// ValidateCredentials verifies the SDK's credentials against the server it is bound to.
//
// It returns a *exceptions.CommsValidationError if the SDK, its user name or its API key is missing,
// and a *exceptions.CommsAuthenticationError if the server rejected the credentials or they could not
// be verified.
func ValidateCredentials(sdk CommsSdkInterface) (bool, error) {
	if sdk == nil {
		return false, &exceptions.CommsValidationError{Message: "CommsSDK instance cannot be null"}
	}

	if sdk.GetApiKey() == "" || sdk.GetUserName() == "" {
		return false, &exceptions.CommsValidationError{Message: "either API Key or Username must be provided"}
	}

	if !isValidCredential(sdk) {
		Log().Error("Authentication failed")
		return false, &exceptions.CommsAuthenticationError{Message: "credentials validation failed"}
	}

	Log().Info("Credentials validated successfully.")
	Log().Info("Validated using basic auth")
	sdk.SetAuthenticated()
	return true, nil
}

func isValidCredential(sdk CommsSdkInterface) bool {
	apiRequest := models.ApiRequest{}
	apiRequest.Method = "Balance"
	apiRequest.Userdata = models.UserData{UserName: sdk.GetUserName(), ApiKey: sdk.GetApiKey()}
	apiRequest.WalletType = models.Local

	apiResponse, err := Post(apiRequest, sdk.GetApiURL())
	if err != nil {
		Log().Error("Error validating credentials.", "error", err)
		return false
	}

	if apiResponse.Status == models.OK {
		Log().Info("Credentials validated successfully.")
		return true
	} else {
		Log().Error("Error validating credentials: "+apiResponse.Message, "message", apiResponse.Message)
		return false
	}
}
