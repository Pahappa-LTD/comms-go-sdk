package v1_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sandboxCredentials reads the two env vars every gated test in this file needs.
// Returns ok=false if either is unset, so callers can skip cleanly.
func sandboxCredentials(t *testing.T) (username, apiKey string, ok bool) {
	username = os.Getenv("COMMS_SANDBOX_USERNAME")
	apiKey = os.Getenv("COMMS_SANDBOX_API_KEY")
	if username == "" || apiKey == "" {
		t.Skip("COMMS_SANDBOX_USERNAME/COMMS_SANDBOX_API_KEY not set; skipping live sandbox smoke test")
		return "", "", false
	}
	return username, apiKey, true
}

// TestLiveSandbox_SendSMSGoldenPath is the one test in this suite that talks to the
// real sandbox API. It only runs when COMMS_SANDBOX_USERNAME/COMMS_SANDBOX_API_KEY are
// set in the environment (never hardcoded), and skips cleanly otherwise so the rest of
// the suite stays runnable without credentials.
func TestLiveSandbox_SendSMSGoldenPath(t *testing.T) {
	username, apiKey, ok := sandboxCredentials(t)
	if !ok {
		return
	}

	v1.UseSandBox()
	t.Cleanup(v1.UseLiveServer)

	sdk, err := v1.Authenticate(username, apiKey)
	require.NoError(t, err)
	require.True(t, sdk.IsAuthenticated())

	success, err := sdk.SendSMS("256700000000", "Test message from Go SDK live smoke test")
	require.NoError(t, err)
	assert.True(t, success)
}

// TestLiveSandbox_WrongCredentials always runs (never gated) since it never reaches
// the point of sending anything - it costs nothing and gives baseline coverage even
// without sandbox access. It targets the sandbox explicitly, never live production.
func TestLiveSandbox_WrongCredentials(t *testing.T) {
	v1.UseSandBox()
	t.Cleanup(v1.UseLiveServer)

	sdk, err := v1.Authenticate("invalid-user", "invalid-key-00000000000000000000000000000000")
	assert.Error(t, err)
	assert.Nil(t, sdk)
}

func TestLiveSandbox_SendSMSMultipleNumbers(t *testing.T) {
	username, apiKey, ok := sandboxCredentials(t)
	if !ok {
		return
	}

	v1.UseSandBox()
	t.Cleanup(v1.UseLiveServer)

	sdk, err := v1.Authenticate(username, apiKey)
	require.NoError(t, err)
	require.True(t, sdk.IsAuthenticated())

	numbers := []string{"256700000000", "256700000001", "256700000002"}
	success, err := sdk.SendSMS(numbers, "Test message from Go SDK live smoke test (multiple)")
	require.NoError(t, err)
	assert.True(t, success)
}

// TestLiveSandbox_RejectsMoreThan1000Numbers confirms the SDK surfaces the real API's
// server-side rejection of oversized recipient lists as a clean error, rather than
// panicking or silently batch-sending 1001 real messages.
func TestLiveSandbox_RejectsMoreThan1000Numbers(t *testing.T) {
	username, apiKey, ok := sandboxCredentials(t)
	if !ok {
		return
	}

	v1.UseSandBox()
	t.Cleanup(v1.UseLiveServer)

	sdk, err := v1.Authenticate(username, apiKey)
	require.NoError(t, err)
	require.True(t, sdk.IsAuthenticated())

	numbers := make([]string, 0, 1001)
	for i := 0; i < 1001; i++ {
		numbers = append(numbers, fmt.Sprintf("256700%06d", i))
	}

	success, err := sdk.SendSMS(numbers, "This should be rejected for exceeding 1000 recipients")
	assert.Error(t, err)
	assert.False(t, success)
}

func TestLiveSandbox_BalanceMethods(t *testing.T) {
	username, apiKey, ok := sandboxCredentials(t)
	if !ok {
		return
	}

	v1.UseSandBox()
	t.Cleanup(v1.UseLiveServer)

	sdk, err := v1.Authenticate(username, apiKey)
	require.NoError(t, err)
	require.True(t, sdk.IsAuthenticated())

	balance, err := sdk.GetBalance()
	require.NoError(t, err)
	require.NotNil(t, balance)
	assert.GreaterOrEqual(t, *balance, 0.0)

	response, err := sdk.QueryBalance()
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, models.OK, response.Status)
}
