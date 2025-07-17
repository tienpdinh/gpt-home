package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tienpdinh/gpt-home/internal/config"
	"github.com/tienpdinh/gpt-home/pkg/models"

	"github.com/sirupsen/logrus"
)

type Service struct {
	ollamaURL   string
	modelName   string
	isConnected bool
	mutex       sync.RWMutex
	modelInfo   ModelInfo
	httpClient  *http.Client
	config      OllamaConfig
}

type OllamaConfig struct {
	URL         string
	Model       string
	MaxTokens   int
	Temperature float32
	TopP        float32
	TopK        int
	Timeout     time.Duration
}

// Ollama API request/response structures
type OllamaGenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type OllamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

type ModelInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Loaded  bool   `json:"loaded"`
}

func NewService(ollamaURL, modelName string) *Service {
	return &Service{
		ollamaURL:   ollamaURL,
		modelName:   modelName,
		isConnected: false,
		modelInfo: ModelInfo{
			Name:    fmt.Sprintf("%s-chat", modelName),
			Type:    modelName,
			Version: "ollama",
			Loaded:  false,
		},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: OllamaConfig{
			URL:         ollamaURL,
			Model:       modelName,
			MaxTokens:   512,
			Temperature: 0.7,
			TopP:        0.9,
			TopK:        40,
			Timeout:     30 * time.Second,
		},
	}
}

func NewServiceWithConfig(ollamaURL, modelName string, cfg config.LLMConfig) *Service {
	return &Service{
		ollamaURL:   ollamaURL,
		modelName:   modelName,
		isConnected: false,
		modelInfo: ModelInfo{
			Name:    fmt.Sprintf("%s-chat", modelName),
			Type:    modelName,
			Version: "ollama",
			Loaded:  false,
		},
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		config: OllamaConfig{
			URL:         ollamaURL,
			Model:       modelName,
			MaxTokens:   cfg.MaxTokens,
			Temperature: cfg.Temperature,
			TopP:        cfg.TopP,
			TopK:        cfg.TopK,
			Timeout:     time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

func (s *Service) LoadModel() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logrus.Infof("Connecting to Ollama at: %s", s.ollamaURL)

	// Test connection to Ollama
	if err := s.testConnection(); err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}

	// Check if model is available
	if err := s.checkModel(); err != nil {
		return fmt.Errorf("model %s not available: %w", s.modelName, err)
	}

	s.isConnected = true
	s.modelInfo.Loaded = true

	logrus.Infof("Connected to Ollama with model %s", s.modelName)
	return nil
}

func (s *Service) testConnection() error {
	resp, err := s.httpClient.Get(s.ollamaURL + "/api/tags")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logrus.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama server returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *Service) checkModel() error {
	// Try to generate a simple test prompt to verify model availability
	testReq := OllamaGenerateRequest{
		Model:   s.modelName,
		Prompt:  "Hello",
		Stream:  false,
		Options: map[string]interface{}{"num_predict": 1},
	}

	reqBody, err := json.Marshal(testReq)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Post(s.ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logrus.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("model test failed: %s", string(body))
	}

	return nil
}

func (s *Service) IsLoaded() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isConnected
}

func (s *Service) GetModelInfo() ModelInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.modelInfo
}

func (s *Service) ProcessMessage(message string, context models.Context) (string, []models.DeviceAction, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.isConnected {
		return "", nil, fmt.Errorf("not connected to Ollama")
	}

	// Create a smart home assistant prompt
	prompt := s.createSmartHomePrompt(message, context)

	// Generate response using Ollama
	llmResponse, err := s.generateResponse(prompt)
	if err != nil {
		logrus.Errorf("Failed to generate response: %v", err)
		// Fallback to rule-based parsing
		fallbackResponse, actions := s.parseCommand(message, context)
		return fallbackResponse, actions, nil
	}

	// Parse actions from the LLM response
	actions := s.extractActionsFromResponse(llmResponse)

	logrus.Debugf("Processed message: %s -> %s", message, llmResponse)
	return llmResponse, actions, nil
}

func (s *Service) parseCommand(message string, context models.Context) (string, []models.DeviceAction) {
	message = strings.ToLower(strings.TrimSpace(message))

	// Simple command parsing - this would be replaced by actual LLM processing
	actions := []models.DeviceAction{}

	// Light commands
	if strings.Contains(message, "turn on") && strings.Contains(message, "light") {
		actions = append(actions, models.DeviceAction{
			Action:     "turn_on",
			Parameters: map[string]any{},
		})
		return "I'll turn on the lights for you.", actions
	}

	if strings.Contains(message, "turn off") && strings.Contains(message, "light") {
		actions = append(actions, models.DeviceAction{
			Action:     "turn_off",
			Parameters: map[string]any{},
		})
		return "I'll turn off the lights for you.", actions
	}

	if strings.Contains(message, "dim") && strings.Contains(message, "light") {
		actions = append(actions, models.DeviceAction{
			Action: "set_brightness",
			Parameters: map[string]any{
				"brightness": 128, // 50% brightness
			},
		})
		return "I'll dim the lights for you.", actions
	}

	// Temperature commands
	if strings.Contains(message, "temperature") || strings.Contains(message, "thermostat") {
		if strings.Contains(message, "set") {
			actions = append(actions, models.DeviceAction{
				Action: "set_temperature",
				Parameters: map[string]any{
					"temperature": 22, // Default to 22°C
				},
			})
			return "I'll adjust the temperature for you.", actions
		}
		return "The current temperature is 22°C. Would you like me to adjust it?", actions
	}

	// Status queries
	if strings.Contains(message, "status") || strings.Contains(message, "what") {
		return "I can help you control your smart home devices. Try asking me to turn on lights, adjust temperature, or check device status.", actions
	}

	// Default response
	return "I understand you want to control your smart home, but I'm not sure exactly what you'd like me to do. Could you be more specific?", actions
}

func (s *Service) generateResponse(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()

	// Prepare Ollama request
	req := OllamaGenerateRequest{
		Model:  s.config.Model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"num_predict": s.config.MaxTokens,
			"temperature": s.config.Temperature,
			"top_p":       s.config.TopP,
			"top_k":       float64(s.config.TopK),
			"stop":        []string{"</response>", "Human:", "User:"},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Ollama
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.ollamaURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logrus.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("Ollama error: %s", ollamaResp.Error)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

func (s *Service) createSmartHomePrompt(message string, context models.Context) string {
	deviceContext := ""
	if len(context.ReferencedDevices) > 0 {
		deviceContext = fmt.Sprintf("\nPreviously referenced devices: %s", strings.Join(context.ReferencedDevices, ", "))
	}

	return fmt.Sprintf(`You are a helpful smart home assistant. You can control lights, switches, climate, and other devices.

Available actions:
- turn_on/turn_off: For lights and switches
- set_brightness: For lights (0-255)
- set_temperature: For climate (degrees)

Respond naturally and briefly. If you perform an action, mention it.%s

Human: %s
Assistant:`, deviceContext, message)
}

func (s *Service) extractActionsFromResponse(response string) []models.DeviceAction {
	// Simple extraction - in a production system, you'd use more sophisticated parsing
	actions := []models.DeviceAction{}
	response = strings.ToLower(response)

	if strings.Contains(response, "turn on") || strings.Contains(response, "turning on") {
		actions = append(actions, models.DeviceAction{
			Action:     "turn_on",
			Parameters: map[string]any{},
		})
	}

	if strings.Contains(response, "turn off") || strings.Contains(response, "turning off") {
		actions = append(actions, models.DeviceAction{
			Action:     "turn_off",
			Parameters: map[string]any{},
		})
	}

	if strings.Contains(response, "dim") || strings.Contains(response, "dimming") {
		actions = append(actions, models.DeviceAction{
			Action: "set_brightness",
			Parameters: map[string]any{
				"brightness": 128,
			},
		})
	}

	return actions
}

func (s *Service) UnloadModel() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.isConnected = false
	s.modelInfo.Loaded = false

	logrus.Info("Disconnected from Ollama")
	return nil
}
