package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/tienpdinh/gpt-home/internal/conversation"
	"github.com/tienpdinh/gpt-home/internal/device"
	"github.com/tienpdinh/gpt-home/internal/llm"
	"github.com/tienpdinh/gpt-home/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	deviceManager       *device.Manager
	llmService          *llm.Service
	conversationManager *conversation.Manager
	startTime           time.Time
}

func NewHandler(deviceManager *device.Manager, llmService *llm.Service, conversationManager *conversation.Manager) *Handler {
	return &Handler{
		deviceManager:       deviceManager,
		llmService:          llmService,
		conversationManager: conversationManager,
		startTime:           time.Now(),
	}
}

// HandleChat processes chat messages and returns AI responses
func (h *Handler) HandleChat(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Get or create conversation
	var conv *models.Conversation
	var err error

	if req.ConversationID != uuid.Nil {
		conv, err = h.conversationManager.GetConversation(req.ConversationID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get conversation")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversation"})
			return
		}
	} else {
		conv = h.conversationManager.CreateConversation()
	}

	// Add user message to conversation
	userMessage := models.Message{
		ID:        uuid.New(),
		Role:      models.MessageRoleUser,
		Content:   req.Message,
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, userMessage)

	// Process message with LLM
	response, actions, err := h.llmService.ProcessMessage(req.Message, conv.Context)
	if err != nil {
		logrus.WithError(err).Error("Failed to process message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	// Execute device actions if any
	for _, action := range actions {
		if err := h.deviceManager.ExecuteAction(action); err != nil {
			logrus.WithError(err).Errorf("Failed to execute action: %s", action.Action)
		}
	}

	// Add assistant response to conversation
	assistantMessage := models.Message{
		ID:        uuid.New(),
		Role:      models.MessageRoleAssistant,
		Content:   response,
		Timestamp: time.Now(),
		Metadata: models.Metadata{
			ProcessingTime: time.Since(startTime).Seconds(),
			ModelUsed:      h.llmService.GetModelInfo().Name,
		},
	}
	conv.Messages = append(conv.Messages, assistantMessage)

	// Update conversation
	if err := h.conversationManager.UpdateConversation(conv); err != nil {
		logrus.WithError(err).Warn("Failed to update conversation")
	}

	// Return response
	chatResponse := models.ChatResponse{
		Response:         response,
		ConversationID:   conv.ID,
		MessageID:        assistantMessage.ID,
		Context:          conv.Context,
		ActionsPerformed: actions,
		Metadata:         assistantMessage.Metadata,
	}

	c.JSON(http.StatusOK, chatResponse)
}

// GetDevices returns all available devices
func (h *Handler) GetDevices(c *gin.Context) {
	devices, err := h.deviceManager.GetAllDevices()
	if err != nil {
		logrus.WithError(err).Error("Failed to get devices")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

// GetDevice returns a specific device by ID
func (h *Handler) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")

	device, err := h.deviceManager.GetDevice(deviceID)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to get device: %s", deviceID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// ControlDevice executes an action on a specific device
func (h *Handler) ControlDevice(c *gin.Context) {
	deviceID := c.Param("id")

	var action models.DeviceAction
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.deviceManager.ExecuteActionOnDevice(deviceID, action); err != nil {
		logrus.WithError(err).Errorf("Failed to control device: %s", deviceID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to control device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetConversation returns a specific conversation
func (h *Handler) GetConversation(c *gin.Context) {
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	conv, err := h.conversationManager.GetConversation(conversationID)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to get conversation: %s", conversationID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// DeleteConversation deletes a specific conversation
func (h *Handler) DeleteConversation(c *gin.Context) {
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	if err := h.conversationManager.DeleteConversation(conversationID); err != nil {
		logrus.WithError(err).Errorf("Failed to delete conversation: %s", conversationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// HealthCheck returns system health status
func (h *Handler) HealthCheck(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	health := models.HealthStatus{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     "1.0.0",
		Uptime:      time.Since(h.startTime).String(),
		MemoryUsage: formatBytes(memStats.Alloc),
		Services: models.Services{
			LLM: models.ServiceStatus{
				Status:      h.getLLMStatus(),
				LastChecked: time.Now(),
			},
			HomeAssistant: models.ServiceStatus{
				Status:      h.getHAStatus(),
				LastChecked: time.Now(),
			},
			Database: models.ServiceStatus{
				Status:      "healthy",
				LastChecked: time.Now(),
			},
		},
	}

	c.JSON(http.StatusOK, health)
}

func (h *Handler) getLLMStatus() string {
	if h.llmService.IsLoaded() {
		return "healthy"
	}
	return "error"
}

func (h *Handler) getHAStatus() string {
	if h.deviceManager.IsConnected() {
		return "healthy"
	}
	return "error"
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return "0 B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return "%.1f %cB"
}
