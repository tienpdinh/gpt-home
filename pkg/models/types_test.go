package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCreation(t *testing.T) {
	tests := []struct {
		name     string
		device   Device
		expected Device
	}{
		{
			name: "create light device",
			device: Device{
				ID:       "light.living_room",
				Name:     "Living Room Light",
				Type:     DeviceTypeLight,
				State:    "on",
				Domain:   "light",
				EntityID: "light.living_room",
			},
			expected: Device{
				ID:       "light.living_room",
				Name:     "Living Room Light",
				Type:     DeviceTypeLight,
				State:    "on",
				Domain:   "light",
				EntityID: "light.living_room",
			},
		},
		{
			name: "create switch device",
			device: Device{
				ID:       "switch.porch",
				Name:     "Porch Switch",
				Type:     DeviceTypeSwitch,
				State:    "off",
				Domain:   "switch",
				EntityID: "switch.porch",
			},
			expected: Device{
				ID:       "switch.porch",
				Name:     "Porch Switch",
				Type:     DeviceTypeSwitch,
				State:    "off",
				Domain:   "switch",
				EntityID: "switch.porch",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.ID, tt.device.ID)
			assert.Equal(t, tt.expected.Name, tt.device.Name)
			assert.Equal(t, tt.expected.Type, tt.device.Type)
			assert.Equal(t, tt.expected.State, tt.device.State)
			assert.Equal(t, tt.expected.Domain, tt.device.Domain)
			assert.Equal(t, tt.expected.EntityID, tt.device.EntityID)
		})
	}
}

func TestDeviceAction(t *testing.T) {
	action := DeviceAction{
		Action: "turn_on",
		Parameters: map[string]any{
			"brightness": 255,
			"color":      "red",
		},
	}

	assert.Equal(t, "turn_on", action.Action)
	assert.Equal(t, 255, action.Parameters["brightness"])
	assert.Equal(t, "red", action.Parameters["color"])
}

func TestConversationCreation(t *testing.T) {
	conv := Conversation{
		ID:        uuid.New(),
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Context: Context{
			ReferencedDevices: []string{},
			UserPreferences:   make(map[string]string),
			SessionData:       make(map[string]any),
		},
	}

	assert.NotEqual(t, uuid.Nil, conv.ID)
	assert.Empty(t, conv.Messages)
	assert.NotZero(t, conv.CreatedAt)
	assert.NotZero(t, conv.UpdatedAt)
	assert.Empty(t, conv.Context.ReferencedDevices)
	assert.NotNil(t, conv.Context.UserPreferences)
	assert.NotNil(t, conv.Context.SessionData)
}

func TestMessageRoles(t *testing.T) {
	tests := []struct {
		name string
		role MessageRole
	}{
		{"user role", MessageRoleUser},
		{"assistant role", MessageRoleAssistant},
		{"system role", MessageRoleSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := Message{
				ID:        uuid.New(),
				Role:      tt.role,
				Content:   "test message",
				Timestamp: time.Now(),
			}

			assert.Equal(t, tt.role, message.Role)
			assert.Equal(t, "test message", message.Content)
			assert.NotEqual(t, uuid.Nil, message.ID)
		})
	}
}

func TestDeviceTypes(t *testing.T) {
	tests := []struct {
		name       string
		deviceType DeviceType
		expected   string
	}{
		{"light type", DeviceTypeLight, "light"},
		{"switch type", DeviceTypeSwitch, "switch"},
		{"sensor type", DeviceTypeSensor, "sensor"},
		{"climate type", DeviceTypeClimate, "climate"},
		{"cover type", DeviceTypeCover, "cover"},
		{"fan type", DeviceTypeFan, "fan"},
		{"media type", DeviceTypeMedia, "media_player"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.deviceType))
		})
	}
}

func TestChatRequestValidation(t *testing.T) {
	req := ChatRequest{
		Message:        "turn on the lights",
		ConversationID: uuid.New(),
		Context: &Context{
			ReferencedDevices: []string{"light.living_room"},
			UserPreferences:   make(map[string]string),
			SessionData:       make(map[string]any),
		},
	}

	assert.Equal(t, "turn on the lights", req.Message)
	assert.NotEqual(t, uuid.Nil, req.ConversationID)
	assert.NotNil(t, req.Context)
	assert.Contains(t, req.Context.ReferencedDevices, "light.living_room")
}

func TestHealthStatus(t *testing.T) {
	health := HealthStatus{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     "1.0.0",
		Uptime:      "5m30s",
		MemoryUsage: "256MB",
		Services: Services{
			LLM: ServiceStatus{
				Status:      "healthy",
				LastChecked: time.Now(),
				Message:     "Model loaded successfully",
			},
			HomeAssistant: ServiceStatus{
				Status:      "healthy",
				LastChecked: time.Now(),
				Message:     "Connected to HA",
			},
			Database: ServiceStatus{
				Status:      "healthy",
				LastChecked: time.Now(),
				Message:     "In-memory storage active",
			},
		},
	}

	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.Equal(t, "5m30s", health.Uptime)
	assert.Equal(t, "256MB", health.MemoryUsage)
	assert.Equal(t, "healthy", health.Services.LLM.Status)
	assert.Equal(t, "healthy", health.Services.HomeAssistant.Status)
	assert.Equal(t, "healthy", health.Services.Database.Status)
}
