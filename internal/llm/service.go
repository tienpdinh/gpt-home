package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/tienpdinh/gpt-home/pkg/models"

	"github.com/sirupsen/logrus"
)

type Service struct {
	backend   LLMBackend
	config    GenerationConfig
	template  PromptTemplate
	mutex     sync.RWMutex
}

type ModelInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Loaded  bool   `json:"loaded"`
}

func NewService(modelPath, modelType string) *Service {
	// Create appropriate backend based on configuration
	var backend LLMBackend
	backend = NewLocalBackend(modelPath, modelType)

	return &Service{
		backend: backend,
		config: GenerationConfig{
			MaxTokens:   512,
			Temperature: 0.7,
			TopP:        0.9,
			TopK:        40,
			StopTokens:  []string{"\n\n", "User:", "Assistant:"},
		},
		template: SmartHomePromptTemplate,
	}
}

func (s *Service) LoadModel() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.backend.LoadModel()
}

func (s *Service) IsLoaded() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.backend.IsLoaded()
}

func (s *Service) GetModelInfo() ModelInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.backend.GetModelInfo()
}

func (s *Service) ProcessMessage(message string, context models.Context) (string, []models.DeviceAction, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.backend.IsLoaded() {
		return "", nil, fmt.Errorf("model not loaded")
	}

	// Format prompt using template
	prompt, err := s.formatPrompt(message, context)
	if err != nil {
		return "", nil, fmt.Errorf("failed to format prompt: %w", err)
	}

	// Generate response using LLM backend
	response, err := s.backend.GenerateResponse(prompt, s.config)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Parse the response to extract actions
	returnMessage, actions, err := s.parseResponse(response)
	if err != nil {
		logrus.Warnf("Failed to parse LLM response, using raw response: %v", err)
		returnMessage = response
		actions = []models.DeviceAction{}
	}

	logrus.Debugf("Processed message: %s -> %s (actions: %d)", message, returnMessage, len(actions))
	return returnMessage, actions, nil
}

// formatPrompt creates a formatted prompt using the template
func (s *Service) formatPrompt(message string, context models.Context) (string, error) {
	// Create template data
	templateData := struct {
		Message string
		Context models.Context
	}{
		Message: message,
		Context: context,
	}

	// Parse and execute system prompt
	systemTmpl, err := template.New("system").Parse(s.template.SystemPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse system template: %w", err)
	}

	var systemBuf bytes.Buffer
	if err := systemTmpl.Execute(&systemBuf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute system template: %w", err)
	}

	// Parse and execute user prompt
	userTmpl, err := template.New("user").Parse(s.template.UserTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse user template: %w", err)
	}

	var userBuf bytes.Buffer
	if err := userTmpl.Execute(&userBuf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute user template: %w", err)
	}

	// Combine system and user prompts
	fullPrompt := systemBuf.String() + "\n\n" + userBuf.String()
	return fullPrompt, nil
}

// parseResponse parses the LLM response to extract message and actions
func (s *Service) parseResponse(response string) (string, []models.DeviceAction, error) {
	// Try to parse as JSON first
	var jsonResponse struct {
		Response string `json:"response"`
		Actions  []struct {
			Action     string         `json:"action"`
			DeviceType string         `json:"device_type"`
			Parameters map[string]any `json:"parameters"`
		} `json:"actions"`
	}

	// Clean response - sometimes LLM adds extra text
	response = strings.TrimSpace(response)
	if idx := strings.Index(response, "{"); idx >= 0 {
		response = response[idx:]
	}
	if idx := strings.LastIndex(response, "}"); idx >= 0 {
		response = response[:idx+1]
	}

	err := json.Unmarshal([]byte(response), &jsonResponse)
	if err != nil {
		return response, []models.DeviceAction{}, nil // Return raw response if JSON parsing fails
	}

	// Convert to models.DeviceAction
	actions := make([]models.DeviceAction, len(jsonResponse.Actions))
	for i, action := range jsonResponse.Actions {
		actions[i] = models.DeviceAction{
			Action:     action.Action,
			Parameters: action.Parameters,
		}
	}

	return jsonResponse.Response, actions, nil
}

func (s *Service) UnloadModel() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.backend.UnloadModel()
}

// UpdateConfig updates the generation configuration
func (s *Service) UpdateConfig(config GenerationConfig) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.config = config
}

// GetConfig returns the current generation configuration
func (s *Service) GetConfig() GenerationConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.config
}
