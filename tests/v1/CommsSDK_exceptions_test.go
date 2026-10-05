package v1_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/exceptions"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/models"
	"github.com/Pahappa-LTD/comms-go-sdk/src/v1/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const okBalance = `{"Status":"OK","Message":"Success","Balance":100}`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// stubTransport routes every http.Post made by the SDK through respond (no network) and records
// the requested URLs. The original transport and API_URL are restored when the test ends.
func stubTransport(t *testing.T, respond func(*http.Request) (*http.Response, error)) *[]string {
	t.Helper()
	var urls []string
	originalTransport := http.DefaultTransport
	originalApiUrl := v1.API_URL
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		urls = append(urls, r.URL.String())
		return respond(r)
	})
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
		v1.API_URL = originalApiUrl
	})
	return &urls
}

func reply(status int, body string) func(*http.Request) (*http.Response, error) {
	return func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	}
}

func balanceRequest() models.ApiRequest {
	return models.ApiRequest{Method: "Balance", Userdata: models.UserData{UserName: "user", ApiKey: "key"}, WalletType: models.Local}
}

func TestLegacy_EveryPreRefactorPublicMemberStillWorks(t *testing.T) {
	stubTransport(t, reply(200, okBalance))
	v1.UseSandBox()
	assert.Equal(t, v1.SandboxApiURL, v1.API_URL)
	v1.UseLiveServer()
	assert.Equal(t, v1.LiveApiURL, v1.API_URL)

	sdk, err := v1.Authenticate("user", "secret-key")
	require.NoError(t, err)
	assert.True(t, sdk.IsAuthenticated())
	assert.Equal(t, "user", sdk.GetUserName())
	assert.Equal(t, "secret-key", sdk.GetApiKey())
	assert.Equal(t, "EgoSMS", sdk.WithSenderId("EgoSMS").GetSenderId())
	assert.NotContains(t, sdk.String(), "secret-key")

	resp, err := sdk.QueryBalance()
	require.NoError(t, err)
	assert.NotNil(t, resp)
	balance, err := sdk.GetBalanceFull(models.Local)
	require.NoError(t, err)
	assert.Equal(t, 100.0, *balance)
}

func TestLegacy_InstanceFollowsTheGlobalUrlEvenAfterCreation(t *testing.T) {
	urls := stubTransport(t, reply(200, okBalance))
	v1.API_URL = "http://first.test/"
	sdk, err := v1.Authenticate("user", "key")
	require.NoError(t, err)

	v1.API_URL = "http://second.test/"
	_, err = sdk.QueryBalance()
	require.NoError(t, err)

	assert.Equal(t, []string{"http://first.test/", "http://second.test/"}, *urls)
	assert.Equal(t, "http://second.test/", sdk.GetApiURL())
}

func TestFactories_LiveAndSandboxIgnoreTheGlobalUrl(t *testing.T) {
	urls := stubTransport(t, reply(200, okBalance))
	v1.API_URL = "http://never-called.test/"

	live, err := v1.Live("user", "key")
	require.NoError(t, err)
	sandbox, err := v1.Sandbox("user", "key")
	require.NoError(t, err)
	_, _ = live.QueryBalance()
	_, _ = sandbox.QueryBalance()

	assert.Equal(t, []string{v1.LiveApiURL, v1.SandboxApiURL, v1.LiveApiURL, v1.SandboxApiURL}, *urls)
}

func TestFactories_RejectedCredentialsReturnAnUnauthenticatedInstance(t *testing.T) {
	stubTransport(t, reply(400, `{"Status":"Failed","Message":"Invalid credentials"}`))

	sdk, err := v1.Live("baduser", "badkey")

	require.NoError(t, err)
	assert.False(t, sdk.IsAuthenticated())
}

func TestAuthenticate_RejectedCredentialsReturnAnAuthenticationError(t *testing.T) {
	stubTransport(t, reply(200, `{"Status":"Failed","Message":"Invalid credentials"}`))

	_, err := v1.Authenticate("baduser", "badkey")

	var authErr *exceptions.CommsAuthenticationError
	require.True(t, errors.As(err, &authErr))
	assert.Equal(t, "credentials validation failed", err.Error())
	var commsErr exceptions.CommsError
	assert.True(t, errors.As(err, &commsErr))
}

func TestQueryBalance_WrapsAnUnreachableServerWithTheCause(t *testing.T) {
	calls := 0
	stubTransport(t, func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return reply(200, okBalance)(r)
		}
		return nil, errors.New("connection refused")
	})
	sdk, err := v1.Live("user", "key")
	require.NoError(t, err)

	_, err = sdk.QueryBalance()

	var apiErr *exceptions.CommsApiError
	require.True(t, errors.As(err, &apiErr))
	assert.True(t, strings.HasPrefix(err.Error(), "failed to get balance"))
	assert.NotNil(t, errors.Unwrap(err))
}

func TestBadInput_IsAValidationError(t *testing.T) {
	stubTransport(t, reply(200, okBalance))
	sdk, err := v1.Live("user", "key")
	require.NoError(t, err)

	for _, tc := range []struct {
		numbers interface{}
		message string
	}{{"256700000000", ""}, {"256700000000", "x"}, {[]string{}, "Hello"}, {42, "Hello"}} {
		_, err := sdk.QuerySendSMS(tc.numbers, tc.message)
		var validationErr *exceptions.CommsValidationError
		assert.True(t, errors.As(err, &validationErr), "numbers=%v message=%q", tc.numbers, tc.message)
	}

	_, err = v1.Live("", "key")
	var validationErr *exceptions.CommsValidationError
	assert.True(t, errors.As(err, &validationErr))
}

func TestNetworkPost_FallsBackToTheStatusAndBodyForANonJsonReply(t *testing.T) {
	stubTransport(t, reply(502, "Bad Gateway"))

	_, err := utils.Post(balanceRequest(), "http://mock.test/")

	var apiErr *exceptions.CommsApiError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, "HTTP 502: Bad Gateway", err.Error())
}

func TestNetworkPost_NamesTheUrlAndKeepsTheCauseWhenUnreachable(t *testing.T) {
	cause := errors.New("connection refused")
	stubTransport(t, func(*http.Request) (*http.Response, error) { return nil, cause })

	_, err := utils.Post(balanceRequest(), "http://127.0.0.1:1/")

	var apiErr *exceptions.CommsApiError
	require.True(t, errors.As(err, &apiErr))
	assert.Contains(t, err.Error(), "http://127.0.0.1:1/")
	assert.True(t, errors.Is(err, cause))
}

func TestNetworkPost_ReturnsAFailedJsonReplyInsteadOfAnError(t *testing.T) {
	stubTransport(t, reply(400, `{"Status":"Failed","Message":"Invalid credentials"}`))

	resp, err := utils.Post(balanceRequest(), "http://mock.test/")

	require.NoError(t, err)
	assert.Equal(t, models.Failed, resp.Status)
	assert.Equal(t, "Invalid credentials", resp.Message)
}
