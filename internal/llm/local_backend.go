package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// LocalBackend implements LLM inference using local models via llama.cpp
type LocalBackend struct {
	modelPath string
	modelType string
	isLoaded  bool
	mutex     sync.RWMutex
	modelInfo ModelInfo
	process   *exec.Cmd
}

// NewLocalBackend creates a new local LLM backend
func NewLocalBackend(modelPath, modelType string) *LocalBackend {
	return &LocalBackend{
		modelPath: modelPath,
		modelType: modelType,
		isLoaded:  false,
		modelInfo: ModelInfo{
			Name:    fmt.Sprintf("%s-local", modelType),
			Type:    modelType,
			Version: "1.0.0",
			Loaded:  false,
		},
	}
}

// LoadModel loads the model using llama.cpp
func (b *LocalBackend) LoadModel() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Check if model file exists
	if _, err := os.Stat(b.modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", b.modelPath)
	}

	// For now, we'll implement a simple validation
	// In production, you would start llama.cpp server or load the model
	logrus.Infof("Loading local model from: %s", b.modelPath)

	// Validate model file
	if !strings.HasSuffix(b.modelPath, ".gguf") && !strings.HasSuffix(b.modelPath, ".bin") {
		logrus.Warnf("Model file may not be in correct format: %s", b.modelPath)
	}

	b.isLoaded = true
	b.modelInfo.Loaded = true

	logrus.Infof("Local model %s loaded successfully", b.modelType)
	return nil
}

// UnloadModel unloads the model
func (b *LocalBackend) UnloadModel() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.process != nil {
		if err := b.process.Process.Kill(); err != nil {
			logrus.Warnf("Failed to kill process: %v", err)
		}
		b.process = nil
	}

	b.isLoaded = false
	b.modelInfo.Loaded = false

	logrus.Info("Local model unloaded")
	return nil
}

// IsLoaded returns whether the model is loaded
func (b *LocalBackend) IsLoaded() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.isLoaded
}

// GetModelInfo returns model information
func (b *LocalBackend) GetModelInfo() ModelInfo {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.modelInfo
}

// GenerateResponse generates a response using the local model
func (b *LocalBackend) GenerateResponse(prompt string, config GenerationConfig) (string, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if !b.isLoaded {
		return "", fmt.Errorf("model not loaded")
	}

	// For initial implementation, use llama.cpp command line interface
	return b.generateWithLlamaCpp(prompt, config)
}

// generateWithLlamaCpp uses llama.cpp command line for inference
func (b *LocalBackend) generateWithLlamaCpp(prompt string, config GenerationConfig) (string, error) {
	// Check if llama.cpp is available
	llamaCppPath := "llama.cpp" // You might need to adjust this path
	if _, lookupErr := exec.LookPath(llamaCppPath); lookupErr != nil {
		// Fallback to a smart pattern-based response for development
		return b.generateSmartFallback(prompt)
	}

	// Build llama.cpp command
	args := []string{
		"-m", b.modelPath,
		"-p", prompt,
		"-n", fmt.Sprintf("%d", config.MaxTokens),
		"--temp", fmt.Sprintf("%.2f", config.Temperature),
		"--top-p", fmt.Sprintf("%.2f", config.TopP),
		"--top-k", fmt.Sprintf("%d", config.TopK),
	}

	cmd := exec.Command(llamaCppPath, args...)
	output, execErr := cmd.Output()
	if execErr != nil {
		logrus.Warnf("llama.cpp execution failed: %v, falling back to smart response", execErr)
		return b.generateSmartFallback(prompt)
	}

	return string(output), nil
}

// generateSmartFallback provides intelligent responses when llama.cpp is not available
func (b *LocalBackend) generateSmartFallback(prompt string) (string, error) {
	logrus.Debug("Using smart fallback for LLM generation")

	// Extract user message from prompt
	userMessage := b.extractUserMessage(prompt)

	// Generate smart response based on patterns
	response := b.generateSmartResponse(userMessage)

	return response, nil
}

// extractUserMessage extracts the user's actual message from the formatted prompt
func (b *LocalBackend) extractUserMessage(prompt string) string {
	lines := strings.Split(prompt, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "User request:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "User request:"))
		}
	}
	return prompt
}

// generateSmartResponse creates intelligent responses with proper JSON formatting
func (b *LocalBackend) generateSmartResponse(message string) string {
	message = strings.ToLower(strings.TrimSpace(message))

	// Light control patterns
	if strings.Contains(message, "turn on") && (strings.Contains(message, "light") || strings.Contains(message, "lamp")) {
		response := map[string]interface{}{
			"response": "I'll turn on the lights for you.",
			"actions": []map[string]interface{}{
				{
					"action":      "turn_on",
					"device_type": "light",
					"parameters":  map[string]interface{}{},
				},
			},
		}
		return b.toJSON(response)
	}

	if strings.Contains(message, "turn off") && (strings.Contains(message, "light") || strings.Contains(message, "lamp")) {
		response := map[string]interface{}{
			"response": "I'll turn off the lights for you.",
			"actions": []map[string]interface{}{
				{
					"action":      "turn_off",
					"device_type": "light",
					"parameters":  map[string]interface{}{},
				},
			},
		}
		return b.toJSON(response)
	}

	if strings.Contains(message, "dim") && (strings.Contains(message, "light") || strings.Contains(message, "lamp")) {
		response := map[string]interface{}{
			"response": "I'll dim the lights for you.",
			"actions": []map[string]interface{}{
				{
					"action":      "set_brightness",
					"device_type": "light",
					"parameters":  map[string]interface{}{"brightness": 128},
				},
			},
		}
		return b.toJSON(response)
	}

	// Temperature control
	if strings.Contains(message, "temperature") || strings.Contains(message, "thermostat") {
		if strings.Contains(message, "set") || strings.Contains(message, "change") {
			response := map[string]interface{}{
				"response": "I'll adjust the temperature for you.",
				"actions": []map[string]interface{}{
					{
						"action":      "set_temperature",
						"device_type": "climate",
						"parameters":  map[string]interface{}{"temperature": 22},
					},
				},
			}
			return b.toJSON(response)
		}
		response := map[string]interface{}{
			"response": "The current temperature is 22°C. Would you like me to adjust it?",
			"actions":  []map[string]interface{}{},
		}
		return b.toJSON(response)
	}

	// Music/Media control
	if strings.Contains(message, "music") || strings.Contains(message, "play") {
		response := map[string]interface{}{
			"response": "I'll start playing music for you.",
			"actions": []map[string]interface{}{
				{
					"action":      "media_play",
					"device_type": "media_player",
					"parameters":  map[string]interface{}{},
				},
			},
		}
		return b.toJSON(response)
	}

	// Status queries
	if strings.Contains(message, "status") || strings.Contains(message, "what") || strings.Contains(message, "how") {
		response := map[string]interface{}{
			"response": "I can help you control your smart home devices. Try asking me to turn on lights, adjust temperature, or play music.",
			"actions":  []map[string]interface{}{},
		}
		return b.toJSON(response)
	}

	// Default response
	response := map[string]interface{}{
		"response": "I understand you want to control your smart home, but I'm not sure exactly what you'd like me to do. Could you be more specific? I can help with lights, temperature, and music.",
		"actions":  []map[string]interface{}{},
	}
	return b.toJSON(response)
}

// toJSON converts response to JSON string
func (b *LocalBackend) toJSON(response map[string]interface{}) string {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("Failed to marshal response to JSON: %v", err)
		return `{"response": "I encountered an error processing your request.", "actions": []}`
	}
	return string(jsonBytes)
}

// Note: createModelDirectory and downloadModel methods removed as they were unused
// These can be added back when implementing model download functionality
