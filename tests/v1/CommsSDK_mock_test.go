package v1_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturedRequest mirrors models.ApiRequest so tests can assert on exactly what was
// marshalled onto the wire, independent of the SDK's own (de)serialization code.
type capturedRequest struct {
	Method      string                `json:"method"`
	Userdata    models.UserData       `json:"userdata"`
	MessageData []models.MessageModel `json:"msgdata"`
	WalletType  models.WalletType     `json:"walletType"`
}

// mockServer starts an httptest server, points the package-level v1.API_URL at it for
// the duration of the test (restored via t.Cleanup), and records every request body.
// "Balance" requests always get a Status:OK reply so Authenticate() succeeds; every
// other method gets sendResponse, so tests can control what SendSms sees.
func mockServer(t *testing.T, sendResponse models.ApiResponse) *[]capturedRequest {
	t.Helper()
	var captured []capturedRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req capturedRequest
		require.NoError(t, json.Unmarshal(body, &req))
		captured = append(captured, req)

		w.Header().Set("Content-Type", "application/json")
		if req.Method == "Balance" {
			_ = json.NewEncoder(w).Encode(models.ApiResponse{Status: models.OK, Message: "Credentials validated successfully."})
			return
		}
		_ = json.NewEncoder(w).Encode(sendResponse)
	}))

	original := v1.API_URL
	v1.API_URL = srv.URL
	t.Cleanup(func() {
		v1.API_URL = original
		srv.Close()
	})

	return &captured
}

// This is the specific regression guard: the live API rejects an explicit
// "walletType": null on requests, but the SendSms builder used to omit/zero it.
// Every outgoing SendSms request must always carry an explicit Local value.
func TestQuerySendSMSFull_AlwaysSendsExplicitLocalWalletType(t *testing.T) {
	captured := mockServer(t, models.ApiResponse{Status: models.OK, Message: "SMS sent", MessageFollowUpCode: "ABC123"})

	sdk, err := v1.Authenticate("user", "key")
	require.NoError(t, err)

	resp, err := sdk.QuerySendSMS([]string{"256700000000"}, "Test message")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, models.OK, resp.Status)
	assert.Equal(t, "ABC123", resp.MessageFollowUpCode)

	require.Len(t, *captured, 2, "expected one Balance call for auth and one SendSms call")
	sendReq := (*captured)[1]
	assert.Equal(t, "SendSms", sendReq.Method)
	assert.Equal(t, models.Local, sendReq.WalletType, "SendSms must always send an explicit Local walletType")
	require.Len(t, sendReq.MessageData, 1)
	assert.Equal(t, "256700000000", sendReq.MessageData[0].Number)
	assert.Equal(t, models.HIGH, sendReq.MessageData[0].Priority, "priority should default to HIGH when unspecified")
}

func TestValidateCredentials_AlwaysSendsExplicitLocalWalletType(t *testing.T) {
	captured := mockServer(t, models.ApiResponse{Status: models.OK})

	sdk, err := v1.Authenticate("user", "key")
	require.NoError(t, err)
	assert.True(t, sdk.IsAuthenticated())

	require.Len(t, *captured, 1)
	assert.Equal(t, "Balance", (*captured)[0].Method)
	assert.Equal(t, models.Local, (*captured)[0].WalletType, "credential validation must always send an explicit Local walletType")
}

func TestQuerySendSMSFull_StatusOK_ParsesResponse(t *testing.T) {
	mockServer(t, models.ApiResponse{Status: models.OK, Message: "SMS sent", MessageFollowUpCode: "XYZ789"})

	sdk, err := v1.Authenticate("user", "key")
	require.NoError(t, err)

	success, err := sdk.SendSMS([]string{"256700000000"}, "Test message")
	require.NoError(t, err)
	assert.True(t, success)
}

func TestQuerySendSMSFull_StatusFailed_HandledGracefullyNotPanic(t *testing.T) {
	mockServer(t, models.ApiResponse{Status: models.Failed, Message: "insufficient balance"})

	sdk, err := v1.Authenticate("user", "key")
	require.NoError(t, err)

	success, err := sdk.SendSMS([]string{"256700000000"}, "Test message")
	assert.Error(t, err)
	assert.False(t, success)

	resp, err := sdk.QuerySendSMS([]string{"256700000000"}, "Test message")
	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, models.Failed, resp.Status)
}

func TestQueryBalance_DefaultsToLocalWalletType(t *testing.T) {
	captured := mockServer(t, models.ApiResponse{})

	sdk, err := v1.Authenticate("user", "key")
	require.NoError(t, err)

	_, err = sdk.QueryBalance()
	require.NoError(t, err)

	require.Len(t, *captured, 2, "expected one Balance call for auth and one for QueryBalance")
	assert.Equal(t, models.Local, (*captured)[1].WalletType)

	_, err = sdk.GetBalance()
	require.NoError(t, err)
	require.Len(t, *captured, 3)
	assert.Equal(t, models.Local, (*captured)[2].WalletType)
}

func TestQueryBalanceFull_RespectsExplicitInternational(t *testing.T) {
	captured := mockServer(t, models.ApiResponse{})

	sdk, err := v1.Authenticate("user", "key")
	require.NoError(t, err)

	_, err = sdk.QueryBalanceFull(models.International)
	require.NoError(t, err)

	require.Len(t, *captured, 2)
	assert.Equal(t, models.International, (*captured)[1].WalletType)

	_, err = sdk.GetBalanceFull(models.International)
	require.NoError(t, err)
	require.Len(t, *captured, 3)
	assert.Equal(t, models.International, (*captured)[2].WalletType)
}
