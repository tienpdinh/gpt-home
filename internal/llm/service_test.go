package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tienpdinh/gpt-home/pkg/models"
)

func TestNewService(t *testing.T) {
	service := NewService("/path/to/model.bin", "tinyllama")

	assert.NotNil(t, service)
	assert.Equal(t, "/path/to/model.bin", service.modelPath)
	assert.Equal(t, "tinyllama", service.modelType)
	assert.False(t, service.isLoaded)
	assert.Equal(t, "tinyllama-chat", service.modelInfo.Name)
	assert.Equal(t, "tinyllama", service.modelInfo.Type)
	assert.Equal(t, "1.0.0", service.modelInfo.Version)
	assert.False(t, service.modelInfo.Loaded)
}

func TestLoadModel(t *testing.T) {
	service := NewService("/path/to/model.bin", "tinyllama")

	// Test initial state
	assert.False(t, service.IsLoaded())

	// Load model
	err := service.LoadModel()
	require.NoError(t, err)

	// Verify loaded state
	assert.True(t, service.IsLoaded())

	modelInfo := service.GetModelInfo()
	assert.True(t, modelInfo.Loaded)
	assert.Equal(t, "tinyllama-chat", modelInfo.Name)
	assert.Equal(t, "tinyllama", modelInfo.Type)
}

func TestGetModelInfo(t *testing.T) {
	service := NewService("/custom/model/phi2.bin", "phi2")

	modelInfo := service.GetModelInfo()
	assert.Equal(t, "phi2-chat", modelInfo.Name)
	assert.Equal(t, "phi2", modelInfo.Type)
	assert.Equal(t, "1.0.0", modelInfo.Version)
	assert.False(t, modelInfo.Loaded)

	// Load model and check again
	service.LoadModel()
	modelInfo = service.GetModelInfo()
	assert.True(t, modelInfo.Loaded)
}

func TestProcessMessageWithoutLoadedModel(t *testing.T) {
	service := NewService("/path/to/model.bin", "tinyllama")
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
	service := NewService("/path/to/model.bin", "tinyllama")
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
				assert.Equal(t, 128, actions[0].Parameters["brightness"])
			}
		})
	}
}

func TestProcessMessageTemperatureCommands(t *testing.T) {
	service := NewService("/path/to/model.bin", "tinyllama")
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
					assert.Equal(t, 22, actions[0].Parameters["temperature"])
				}
			} else {
				assert.Empty(t, actions)
			}
		})
	}
}

func TestProcessMessageStatusQueries(t *testing.T) {
	service := NewService("/path/to/model.bin", "tinyllama")
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
	service := NewService("/path/to/model.bin", "tinyllama")
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
	service := NewService("/path/to/model.bin", "tinyllama")
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
	service := NewService("/path/to/model.bin", "tinyllama")

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
	service := NewService("/path/to/model.bin", "tinyllama")
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

			assert.Equal(t, tt.modelPath, service.modelPath)
			assert.Equal(t, tt.modelType, service.modelType)
			assert.Equal(t, tt.modelType+"-chat", service.modelInfo.Name)
			assert.Equal(t, tt.modelType, service.modelInfo.Type)
		})
	}
}
