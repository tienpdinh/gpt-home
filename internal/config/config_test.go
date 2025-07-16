package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	config, err := Load()
	require.NoError(t, err)

	// Test default values
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, "debug", config.Server.Mode)
	assert.Equal(t, 10*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, config.Server.WriteTimeout)

	assert.Equal(t, "http://homeassistant.local:8123", config.HomeAssistant.URL)
	assert.Equal(t, "", config.HomeAssistant.Token)
	assert.Equal(t, 30, config.HomeAssistant.Timeout)

	assert.Equal(t, "./models/tinyllama-1.1b-chat-q4_0.bin", config.LLM.ModelPath)
	assert.Equal(t, "tinyllama", config.LLM.ModelType)
	assert.Equal(t, 512, config.LLM.MaxTokens)
	assert.Equal(t, float32(0.7), config.LLM.Temperature)
	assert.Equal(t, float32(0.9), config.LLM.TopP)
	assert.Equal(t, 40, config.LLM.TopK)
	assert.Equal(t, 2048, config.LLM.ContextLength)

	assert.Equal(t, "memory", config.Storage.Type)
	assert.Equal(t, "./data", config.Storage.Path)
	assert.True(t, config.Storage.InMemory)

	assert.Equal(t, "info", config.LogLevel)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set test environment variables
	envVars := map[string]string{
		"SERVER_PORT":          "9090",
		"SERVER_HOST":          "127.0.0.1",
		"SERVER_MODE":          "release",
		"SERVER_READ_TIMEOUT":  "15",
		"SERVER_WRITE_TIMEOUT": "20",
		"HA_URL":               "http://test-ha:8123",
		"HA_TOKEN":             "test-token-123",
		"HA_TIMEOUT":           "45",
		"LLM_MODEL_PATH":       "/custom/model/path.bin",
		"LLM_MODEL_TYPE":       "phi2",
		"LLM_MAX_TOKENS":       "1024",
		"LLM_TEMPERATURE":      "0.5",
		"LLM_TOP_P":            "0.8",
		"LLM_TOP_K":            "50",
		"LLM_CONTEXT_LENGTH":   "4096",
		"STORAGE_TYPE":         "file",
		"STORAGE_PATH":         "/custom/data",
		"STORAGE_IN_MEMORY":    "false",
		"LOG_LEVEL":            "debug",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Clean up after test
	defer func() {
		for key := range envVars {
			os.Unsetenv(key)
		}
	}()

	config, err := Load()
	require.NoError(t, err)

	// Test environment variable values
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, "release", config.Server.Mode)
	assert.Equal(t, 15*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 20*time.Second, config.Server.WriteTimeout)

	assert.Equal(t, "http://test-ha:8123", config.HomeAssistant.URL)
	assert.Equal(t, "test-token-123", config.HomeAssistant.Token)
	assert.Equal(t, 45, config.HomeAssistant.Timeout)

	assert.Equal(t, "/custom/model/path.bin", config.LLM.ModelPath)
	assert.Equal(t, "phi2", config.LLM.ModelType)
	assert.Equal(t, 1024, config.LLM.MaxTokens)
	assert.Equal(t, float32(0.5), config.LLM.Temperature)
	assert.Equal(t, float32(0.8), config.LLM.TopP)
	assert.Equal(t, 50, config.LLM.TopK)
	assert.Equal(t, 4096, config.LLM.ContextLength)

	assert.Equal(t, "file", config.Storage.Type)
	assert.Equal(t, "/custom/data", config.Storage.Path)
	assert.False(t, config.Storage.InMemory)

	assert.Equal(t, "debug", config.LogLevel)
}

func TestGetEnvHelpers(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected interface{}
		helper   string
	}{
		{"string value", "TEST_STRING", "hello", "hello", "string"},
		{"string default", "MISSING_STRING", "", "default", "string"},
		{"int value", "TEST_INT", "42", 42, "int"},
		{"int default", "MISSING_INT", "", 100, "int"},
		{"float32 value", "TEST_FLOAT", "3.14", float32(3.14), "float32"},
		{"float32 default", "MISSING_FLOAT", "", float32(2.5), "float32"},
		{"bool true", "TEST_BOOL_TRUE", "true", true, "bool"},
		{"bool false", "TEST_BOOL_FALSE", "false", false, "bool"},
		{"bool default", "MISSING_BOOL", "", true, "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if value is provided
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			switch tt.helper {
			case "string":
				result := getEnv(tt.key, tt.expected.(string))
				assert.Equal(t, tt.expected, result)
			case "int":
				result := getEnvAsInt(tt.key, tt.expected.(int))
				assert.Equal(t, tt.expected, result)
			case "float32":
				result := getEnvAsFloat32(tt.key, tt.expected.(float32))
				assert.Equal(t, tt.expected, result)
			case "bool":
				result := getEnvAsBool(tt.key, tt.expected.(bool))
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInvalidEnvValues(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		helper   string
		expected interface{}
	}{
		{"invalid int", "INVALID_INT", "not-a-number", "int", 100},
		{"invalid float", "INVALID_FLOAT", "not-a-float", "float32", float32(2.5)},
		{"invalid bool", "INVALID_BOOL", "not-a-bool", "bool", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.key, tt.value)
			defer os.Unsetenv(tt.key)

			switch tt.helper {
			case "int":
				result := getEnvAsInt(tt.key, tt.expected.(int))
				assert.Equal(t, tt.expected, result)
			case "float32":
				result := getEnvAsFloat32(tt.key, tt.expected.(float32))
				assert.Equal(t, tt.expected, result)
			case "bool":
				result := getEnvAsBool(tt.key, tt.expected.(bool))
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
