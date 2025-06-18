package llm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/tienpdinh/gpt-home/pkg/models"

	"github.com/sirupsen/logrus"
)

type Service struct {
	modelPath string
	modelType string
	isLoaded  bool
	mutex     sync.RWMutex
	modelInfo ModelInfo
}

type ModelInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Loaded  bool   `json:"loaded"`
}

func NewService(modelPath, modelType string) *Service {
	return &Service{
		modelPath: modelPath,
		modelType: modelType,
		isLoaded:  false,
		modelInfo: ModelInfo{
			Name:    fmt.Sprintf("%s-chat", modelType),
			Type:    modelType,
			Version: "1.0.0",
			Loaded:  false,
		},
	}
}

func (s *Service) LoadModel() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// For now, simulate model loading
	// In a real implementation, this would load the actual model
	logrus.Infof("Loading model from: %s", s.modelPath)

	// Simulate loading time
	// time.Sleep(2 * time.Second)

	s.isLoaded = true
	s.modelInfo.Loaded = true

	logrus.Infof("Model %s loaded successfully", s.modelType)
	return nil
}

func (s *Service) IsLoaded() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isLoaded
}

func (s *Service) GetModelInfo() ModelInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.modelInfo
}

func (s *Service) ProcessMessage(message string, context models.Context) (string, []models.DeviceAction, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.isLoaded {
		return "", nil, fmt.Errorf("model not loaded")
	}

	// For now, implement a simple rule-based system
	// In a real implementation, this would use the actual LLM
	response, actions := s.parseCommand(message, context)

	logrus.Debugf("Processed message: %s -> %s", message, response)
	return response, actions, nil
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

func (s *Service) UnloadModel() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.isLoaded = false
	s.modelInfo.Loaded = false

	logrus.Info("Model unloaded")
	return nil
}
