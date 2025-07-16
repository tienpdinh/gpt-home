package llm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tienpdinh/gpt-home/pkg/models"
)

// createTempModel creates a temporary model file for testing
func createTempModel(t *testing.T, name string) (string, func()) {
	tempFile := "/tmp/" + name + ".gguf"
	f, err := os.Create(tempFile)
	require.NoError(t, err)
	f.Close()
	return tempFile, func() { os.Remove(tempFile) }
}

func TestNewService(t *testing.T) {
	service := NewService("/path/to/model.bin", "tinyllama")

	assert.NotNil(t, service)
	assert.NotNil(t, service.backend)
	assert.False(t, service.IsLoaded())

	modelInfo := service.GetModelInfo()
	assert.Equal(t, "tinyllama-local", modelInfo.Name)
	assert.Equal(t, "tinyllama", modelInfo.Type)
	assert.Equal(t, "1.0.0", modelInfo.Version)
	assert.False(t, modelInfo.Loaded)
}

func TestLoadModel(t *testing.T) {
	// Create a temporary model file for testing
	tempFile := "/tmp/test-model.gguf"
	f, err := os.Create(tempFile)
	require.NoError(t, err)
	f.Close()
	defer os.Remove(tempFile)

	service := NewService(tempFile, "tinyllama")

	// Test initial state
	assert.False(t, service.IsLoaded())

	// Load model
	err = service.LoadModel()
	require.NoError(t, err)

	// Verify loaded state
	assert.True(t, service.IsLoaded())

	modelInfo := service.GetModelInfo()
	assert.True(t, modelInfo.Loaded)
	assert.Equal(t, "tinyllama-local", modelInfo.Name)
	assert.Equal(t, "tinyllama", modelInfo.Type)
}

func TestGetModelInfo(t *testing.T) {
	// Create a temporary model file for testing
	tempFile := "/tmp/test-phi2.gguf"
	f, err := os.Create(tempFile)
	require.NoError(t, err)
	f.Close()
	defer os.Remove(tempFile)

	service := NewService(tempFile, "phi2")

	modelInfo := service.GetModelInfo()
	assert.Equal(t, "phi2-local", modelInfo.Name)
	assert.Equal(t, "phi2", modelInfo.Type)
	assert.Equal(t, "1.0.0", modelInfo.Version)
	assert.False(t, modelInfo.Loaded)

	// Load model and check again
	service.LoadModel()
	modelInfo = service.GetModelInfo()
	assert.True(t, modelInfo.Loaded)
}

func TestProcessMessageWithoutLoadedModel(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-unloaded")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")
	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	_, _, err := service.ProcessMessage("turn on the lights", context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model not loaded")
}

func TestProcessMessageLightCommands(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-lights")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")
	err := service.LoadModel()
	require.NoError(t, err)

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	tests := []struct {
		name           string
		message        string
		expectedAction string
		expectedInResp string
	}{
		{
			name:           "turn on lights",
			message:        "turn on the lights",
			expectedAction: "turn_on",
			expectedInResp: "turn on the lights",
		},
		{
			name:           "turn off lights",
			message:        "turn off the lights",
			expectedAction: "turn_off",
			expectedInResp: "turn off the lights",
		},
		{
			name:           "dim lights",
			message:        "dim the lights",
			expectedAction: "set_brightness",
			expectedInResp: "dim the lights",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, actions, err := service.ProcessMessage(tt.message, context)

			require.NoError(t, err)
			assert.Contains(t, response, tt.expectedInResp)
			assert.Len(t, actions, 1)
			assert.Equal(t, tt.expectedAction, actions[0].Action)

			if tt.expectedAction == "set_brightness" {
				assert.Equal(t, float64(128), actions[0].Parameters["brightness"])
			}
		})
	}
}

func TestProcessMessageTemperatureCommands(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-temp")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")
	err := service.LoadModel()
	require.NoError(t, err)

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	tests := []struct {
		name           string
		message        string
		expectAction   bool
		expectedAction string
	}{
		{
			name:           "set temperature",
			message:        "set the temperature to 24 degrees",
			expectAction:   true,
			expectedAction: "set_temperature",
		},
		{
			name:           "adjust thermostat",
			message:        "set the thermostat",
			expectAction:   true,
			expectedAction: "set_temperature",
		},
		{
			name:         "query temperature",
			message:      "what's the temperature?",
			expectAction: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, actions, err := service.ProcessMessage(tt.message, context)

			require.NoError(t, err)
			assert.NotEmpty(t, response)

			if tt.expectAction {
				assert.Len(t, actions, 1)
				assert.Equal(t, tt.expectedAction, actions[0].Action)
				if tt.expectedAction == "set_temperature" {
					assert.Equal(t, float64(22), actions[0].Parameters["temperature"])
				}
			} else {
				assert.Empty(t, actions)
			}
		})
	}
}

func TestProcessMessageStatusQueries(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-status")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")
	err := service.LoadModel()
	require.NoError(t, err)

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	tests := []string{
		"what's the status?",
		"show me device status",
		"what devices are available?",
	}

	for _, message := range tests {
		t.Run(message, func(t *testing.T) {
			response, actions, err := service.ProcessMessage(message, context)

			require.NoError(t, err)
			assert.NotEmpty(t, response)
			assert.Empty(t, actions) // Status queries shouldn't generate actions
			assert.Contains(t, response, "smart home")
		})
	}
}

func TestProcessMessageDefaultResponse(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-default")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")
	err := service.LoadModel()
	require.NoError(t, err)

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	// Test with unrecognized command
	response, actions, err := service.ProcessMessage("make me a sandwich", context)

	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Empty(t, actions)
	assert.Contains(t, response, "smart home")
	assert.Contains(t, response, "not sure")
}

func TestParseCommandVariations(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-variations")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")
	err := service.LoadModel()
	require.NoError(t, err)

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	// Test case sensitivity and whitespace handling
	tests := []struct {
		message        string
		expectedAction string
	}{
		{"TURN ON THE LIGHT", "turn_on"},
		{"  turn on the lights  ", "turn_on"},
		{"Turn Off The Lights", "turn_off"},
		{"DIM THE LIGHT", "set_brightness"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			response, actions, err := service.ProcessMessage(tt.message, context)

			require.NoError(t, err)
			assert.NotEmpty(t, response)
			assert.Len(t, actions, 1)
			assert.Equal(t, tt.expectedAction, actions[0].Action)
		})
	}
}

func TestUnloadModel(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-unload")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")

	// Load model first
	err := service.LoadModel()
	require.NoError(t, err)
	assert.True(t, service.IsLoaded())

	// Unload model
	err = service.UnloadModel()
	require.NoError(t, err)
	assert.False(t, service.IsLoaded())

	modelInfo := service.GetModelInfo()
	assert.False(t, modelInfo.Loaded)
}

func TestConcurrentProcessing(t *testing.T) {
	tempFile, cleanup := createTempModel(t, "test-concurrent")
	defer cleanup()

	service := NewService(tempFile, "tinyllama")
	err := service.LoadModel()
	require.NoError(t, err)

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	// Test concurrent message processing
	done := make(chan bool, 5)
	messages := []string{
		"turn on the lights",
		"turn off the lights",
		"dim the lights",
		"set temperature to 22",
		"what's the status?",
	}

	for _, msg := range messages {
		go func(message string) {
			_, _, err := service.ProcessMessage(message, context)
			assert.NoError(t, err)
			done <- true
		}(msg)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestServiceConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		modelPath string
		modelType string
	}{
		{
			name:      "tinyllama config",
			modelPath: "/models/tinyllama.bin",
			modelType: "tinyllama",
		},
		{
			name:      "phi2 config",
			modelPath: "/models/phi2.bin",
			modelType: "phi2",
		},
		{
			name:      "custom config",
			modelPath: "/custom/path/model.bin",
			modelType: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.modelPath, tt.modelType)
			modelInfo := service.GetModelInfo()

			assert.NotNil(t, service.backend)
			assert.Equal(t, tt.modelType+"-local", modelInfo.Name)
			assert.Equal(t, tt.modelType, modelInfo.Type)
		})
	}
}
