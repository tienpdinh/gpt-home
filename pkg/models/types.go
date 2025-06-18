package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a smart home device
type Device struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        DeviceType     `json:"type"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastUpdated time.Time      `json:"last_updated"`
	Domain      string         `json:"domain"`
	EntityID    string         `json:"entity_id"`
}

// DeviceType represents the type of device
type DeviceType string

const (
	DeviceTypeLight   DeviceType = "light"
	DeviceTypeSwitch  DeviceType = "switch"
	DeviceTypeSensor  DeviceType = "sensor"
	DeviceTypeClimate DeviceType = "climate"
	DeviceTypeCover   DeviceType = "cover"
	DeviceTypeFan     DeviceType = "fan"
	DeviceTypeMedia   DeviceType = "media_player"
)

// DeviceAction represents an action to perform on a device
type DeviceAction struct {
	Action     string         `json:"action"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID        uuid.UUID `json:"id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Context   Context   `json:"context"`
}

// Message represents a single message in a conversation
type Message struct {
	ID        uuid.UUID   `json:"id"`
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  Metadata    `json:"metadata,omitempty"`
}

// MessageRole represents who sent the message
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// Context represents conversation context
type Context struct {
	ReferencedDevices []string          `json:"referenced_devices"`
	LastAction        *DeviceAction     `json:"last_action,omitempty"`
	UserPreferences   map[string]string `json:"user_preferences"`
	SessionData       map[string]any    `json:"session_data"`
}

// Metadata represents additional message metadata
type Metadata struct {
	DevicesReferenced []string `json:"devices_referenced,omitempty"`
	ActionsPerformed  []string `json:"actions_performed,omitempty"`
	ProcessingTime    float64  `json:"processing_time,omitempty"`
	ModelUsed         string   `json:"model_used,omitempty"`
	Confidence        float64  `json:"confidence,omitempty"`
}

// ChatRequest represents an incoming chat request
type ChatRequest struct {
	Message        string    `json:"message" binding:"required"`
	ConversationID uuid.UUID `json:"conversation_id,omitempty"`
	Context        *Context  `json:"context,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Response         string         `json:"response"`
	ConversationID   uuid.UUID      `json:"conversation_id"`
	MessageID        uuid.UUID      `json:"message_id"`
	Context          Context        `json:"context"`
	ActionsPerformed []DeviceAction `json:"actions_performed,omitempty"`
	Metadata         Metadata       `json:"metadata"`
}

// HealthStatus represents system health
type HealthStatus struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
	MemoryUsage string    `json:"memory_usage"`
	Services    Services  `json:"services"`
}

// Services represents status of different services
type Services struct {
	LLM           ServiceStatus `json:"llm"`
	HomeAssistant ServiceStatus `json:"home_assistant"`
	Database      ServiceStatus `json:"database"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Message     string    `json:"message,omitempty"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	ModelPath     string  `json:"model_path"`
	ModelType     string  `json:"model_type"`
	MaxTokens     int     `json:"max_tokens"`
	Temperature   float32 `json:"temperature"`
	TopP          float32 `json:"top_p"`
	TopK          int     `json:"top_k"`
	ContextLength int     `json:"context_length"`
}
