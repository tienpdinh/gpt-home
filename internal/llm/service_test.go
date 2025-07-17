package llm

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tienpdinh/gpt-home/pkg/models"
)

func TestNewService(t *testing.T) {
	service := NewService("http://localhost:11434", "llama3.2")

	assert.NotNil(t, service)
	assert.Equal(t, "http://localhost:11434", service.ollamaURL)
	assert.Equal(t, "llama3.2", service.modelName)
	assert.False(t, service.isConnected)
	assert.Equal(t, "llama3.2-chat", service.modelInfo.Name)
	assert.Equal(t, "llama3.2", service.modelInfo.Type)
	assert.Equal(t, "ollama", service.modelInfo.Version)
	assert.False(t, service.modelInfo.Loaded)
}

func TestIsLoaded(t *testing.T) {
	service := NewService("http://localhost:11434", "llama3.2")

	// Test initial state
	assert.False(t, service.IsLoaded())

	// Simulate connection
	service.isConnected = true
	service.modelInfo.Loaded = true

	// Verify loaded state
	assert.True(t, service.IsLoaded())
}

func TestGetModelInfo(t *testing.T) {
	service := NewService("http://localhost:11434", "qwen2.5")

	modelInfo := service.GetModelInfo()
	assert.Equal(t, "qwen2.5-chat", modelInfo.Name)
	assert.Equal(t, "qwen2.5", modelInfo.Type)
	assert.Equal(t, "ollama", modelInfo.Version)
	assert.False(t, modelInfo.Loaded)

	// Simulate connection and check again
	service.isConnected = true
	service.modelInfo.Loaded = true
	modelInfo = service.GetModelInfo()
	assert.True(t, modelInfo.Loaded)
}

func TestLoadModel_Success(t *testing.T) {
	// Create mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		case "/api/generate":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"response":"Hello","done":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, "llama3.2")

	err := service.LoadModel()
	require.NoError(t, err)
	assert.True(t, service.IsLoaded())
	assert.True(t, service.GetModelInfo().Loaded)
}

func TestLoadModel_ConnectionFailure(t *testing.T) {
	service := NewService("http://nonexistent:11434", "llama3.2")

	err := service.LoadModel()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to Ollama")
	assert.False(t, service.IsLoaded())
}

func TestLoadModel_ModelNotAvailable(t *testing.T) {
	// Create mock server that responds to /api/tags but fails on model test
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		case "/api/generate":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"model not found"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, "nonexistent-model")

	err := service.LoadModel()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model nonexistent-model not available")
	assert.False(t, service.IsLoaded())
}

func TestProcessMessage_NotConnected(t *testing.T) {
	service := NewService("http://localhost:11434", "llama3.2")
	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	_, _, err := service.ProcessMessage("turn on the lights", context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected to Ollama")
}

func TestProcessMessage_Success(t *testing.T) {
	// Create mock Ollama server that returns a smart response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		case "/api/generate":
			w.WriteHeader(http.StatusOK)
			// Simulate LLM response that mentions turning on lights
			w.Write([]byte(`{"response":"I'll turn on the lights for you.","done":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, "llama3.2")
	err := service.LoadModel()
	require.NoError(t, err)

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	response, actions, err := service.ProcessMessage("turn on the lights", context)

	require.NoError(t, err)
	assert.Equal(t, "I'll turn on the lights for you.", response)
	assert.Len(t, actions, 1)
	assert.Equal(t, "turn_on", actions[0].Action)
}

func TestProcessMessage_FallbackToRuleBased(t *testing.T) {
	// Create mock server that fails generation but allows connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		case "/api/generate":
			if strings.Contains(r.URL.Path, "generate") && r.Method == "POST" {
				// Fail generation to test fallback
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"generation failed"}`))
			} else {
				// Allow model check to pass
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"response":"test","done":true}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, "llama3.2")
	
	// First, simulate successful connection for LoadModel
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/api/generate" {
			w.Write([]byte(`{"response":"test","done":true}`))
		} else {
			w.Write([]byte(`{"models":[]}`))
		}
	}))
	defer testServer.Close()
	
	service.ollamaURL = testServer.URL
	err := service.LoadModel()
	require.NoError(t, err)

	// Now switch to failing server for ProcessMessage
	service.ollamaURL = server.URL

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	response, actions, err := service.ProcessMessage("turn on the lights", context)

	// Should fall back to rule-based parsing
	require.NoError(t, err)
	assert.Contains(t, response, "turn on the lights")
	assert.Len(t, actions, 1)
	assert.Equal(t, "turn_on", actions[0].Action)
}

func TestExtractActionsFromResponse(t *testing.T) {
	service := NewService("http://localhost:11434", "llama3.2")

	tests := []struct {
		name           string
		response       string
		expectedAction string
		expectAction   bool
	}{
		{
			name:           "turn on response",
			response:       "I'll turn on the lights for you.",
			expectedAction: "turn_on",
			expectAction:   true,
		},
		{
			name:           "turn off response",
			response:       "I'm turning off the lights now.",
			expectedAction: "turn_off",
			expectAction:   true,
		},
		{
			name:           "dim response", 
			response:       "I'll dim the lights to a comfortable level.",
			expectedAction: "set_brightness",
			expectAction:   true,
		},
		{
			name:         "no action response",
			response:     "The weather is nice today.",
			expectAction: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := service.extractActionsFromResponse(tt.response)

			if tt.expectAction {
				assert.Len(t, actions, 1)
				assert.Equal(t, tt.expectedAction, actions[0].Action)
			} else {
				assert.Empty(t, actions)
			}
		})
	}
}

func TestCreateSmartHomePrompt(t *testing.T) {
	service := NewService("http://localhost:11434", "llama3.2")

	context := models.Context{
		ReferencedDevices: []string{"living_room_light", "bedroom_light"},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	prompt := service.createSmartHomePrompt("turn on the lights", context)

	assert.Contains(t, prompt, "smart home assistant")
	assert.Contains(t, prompt, "turn on the lights")
	assert.Contains(t, prompt, "living_room_light, bedroom_light")
	assert.Contains(t, prompt, "turn_on/turn_off")
	assert.Contains(t, prompt, "set_brightness")
	assert.Contains(t, prompt, "set_temperature")
}

func TestCreateSmartHomePrompt_NoDevices(t *testing.T) {
	service := NewService("http://localhost:11434", "llama3.2")

	context := models.Context{
		ReferencedDevices: []string{},
		UserPreferences:   make(map[string]string),
		SessionData:       make(map[string]any),
	}

	prompt := service.createSmartHomePrompt("what can you do?", context)

	assert.Contains(t, prompt, "smart home assistant")
	assert.Contains(t, prompt, "what can you do?")
	assert.NotContains(t, prompt, "Previously referenced devices")
}

func TestUnloadModel(t *testing.T) {
	service := NewService("http://localhost:11434", "llama3.2")

	// Simulate connected state
	service.isConnected = true
	service.modelInfo.Loaded = true
	assert.True(t, service.IsLoaded())

	// Unload model
	err := service.UnloadModel()
	require.NoError(t, err)
	assert.False(t, service.IsLoaded())

	modelInfo := service.GetModelInfo()
	assert.False(t, modelInfo.Loaded)
}

func TestConcurrentProcessing(t *testing.T) {
	// Create mock server for concurrent testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		case "/api/generate":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"response":"I'll help you with that.","done":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, "llama3.2")
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
		ollamaURL string
		modelName string
	}{
		{
			name:      "llama config",
			ollamaURL: "http://localhost:11434",
			modelName: "llama3.2",
		},
		{
			name:      "qwen config",
			ollamaURL: "http://192.168.1.100:11434",
			modelName: "qwen2.5",
		},
		{
			name:      "custom config",
			ollamaURL: "http://my-server:8080",
			modelName: "custom-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.ollamaURL, tt.modelName)

			assert.Equal(t, tt.ollamaURL, service.ollamaURL)
			assert.Equal(t, tt.modelName, service.modelName)
			assert.Equal(t, tt.modelName+"-chat", service.modelInfo.Name)
			assert.Equal(t, tt.modelName, service.modelInfo.Type)
			assert.Equal(t, "ollama", service.modelInfo.Version)
		})
	}
}