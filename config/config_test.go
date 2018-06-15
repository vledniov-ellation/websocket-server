package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const configPath = "test/config.yaml"

func TestListen(t *testing.T) {
	assert.Equal(t, "0.0.0.0:8080", Listen())
}

func TestServerReadTimeout(t *testing.T) {
	assert.Equal(t, 420*time.Second, ServerReadTimeout())
}

func TestServerWriteTimeout(t *testing.T) {
	assert.Equal(t, 420*time.Second, ServerWriteTimeout())
}

func TestLogLevel(t *testing.T) {
	assert.Equal(t, "debug", LogLevel())
}

func TestLogOutput(t *testing.T) {
	assert.Equal(t, []string{"stdout", "app.log"}, LogOutput())
}

func TestShouldLogCaller(t *testing.T) {
	assert.True(t, ShouldLogCaller())
}

func TestShouldLogStacktrace(t *testing.T) {
	assert.False(t, ShouldLogStacktrace())
}

func TestSetDefaults(t *testing.T) {
	// reset global config and set only defaults.
	cnfg = viper.New()
	setDefaults()
	// init global config back
	defer InitConfig(configPath)

	for _, testCase := range []struct {
		name     string
		config   interface{}
		expected interface{}
	}{
		{"ServerReadTimeout", ServerReadTimeout(), 10 * time.Second},
		{"ServerWriteTimeout", ServerWriteTimeout(), 10 * time.Second},
		{"LogLevel", LogLevel(), "info"},
		{"LogOutput", LogOutput(), []string{"app.log"}},
		{"ShouldLogCaller", ShouldLogCaller(), false},
		{"ShouldLogStacktrace", ShouldLogStacktrace(), true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, testCase.config, "Invalid "+testCase.name)
		})
	}
}

func TestMain(m *testing.M) {
	InitConfig(configPath)
	os.Exit(m.Run())
}
