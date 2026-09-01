package v1_test

import (
	"os"
	"testing"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLiveSandbox_SendSMSGoldenPath is the one test in this suite that talks to the
// real sandbox API. It only runs when COMMS_SANDBOX_USERNAME/COMMS_SANDBOX_API_KEY are
// set in the environment (never hardcoded), and skips cleanly otherwise so the rest of
// the suite stays runnable without credentials.
func TestLiveSandbox_SendSMSGoldenPath(t *testing.T) {
	username := os.Getenv("COMMS_SANDBOX_USERNAME")
	apiKey := os.Getenv("COMMS_SANDBOX_API_KEY")
	if username == "" || apiKey == "" {
		t.Skip("COMMS_SANDBOX_USERNAME/COMMS_SANDBOX_API_KEY not set; skipping live sandbox smoke test")
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
