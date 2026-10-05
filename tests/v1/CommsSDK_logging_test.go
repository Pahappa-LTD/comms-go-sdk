package v1_test

import (
	"fmt"
	"testing"

	"github.com/Pahappa-LTD/comms-go-sdk/src/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingLogger has the same method set as *slog.Logger, so it stands in for one here.
type recordingLogger struct{ records []string }

func (l *recordingLogger) record(level, msg string, kv []interface{}) {
	l.records = append(l.records, fmt.Sprint(level, " ", msg, kv))
}
func (l *recordingLogger) Debug(msg string, kv ...interface{}) { l.record("DEBUG", msg, kv) }
func (l *recordingLogger) Info(msg string, kv ...interface{})  { l.record("INFO", msg, kv) }
func (l *recordingLogger) Warn(msg string, kv ...interface{})  { l.record("WARN", msg, kv) }
func (l *recordingLogger) Error(msg string, kv ...interface{}) { l.record("ERROR", msg, kv) }

func TestLogging_MessagesGoToTheConfiguredLogger(t *testing.T) {
	logger := &recordingLogger{}
	v1.SetLogger(logger)
	t.Cleanup(func() { v1.SetLogger(nil) })
	stubTransport(t, reply(200, `{"Status":"Failed","Message":"Invalid credentials"}`))

	sdk, err := v1.Sandbox("user", "key")
	require.NoError(t, err)
	assert.False(t, sdk.IsAuthenticated())

	assert.Contains(t, logger.records, "ERROR Error validating credentials: Invalid credentials[message Invalid credentials]")
	assert.Contains(t, logger.records, "ERROR Authentication failed[]")
}
