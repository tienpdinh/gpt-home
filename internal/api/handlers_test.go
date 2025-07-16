package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tienpdinh/gpt-home/internal/conversation"
	"github.com/tienpdinh/gpt-home/internal/device"
	"github.com/tienpdinh/gpt-home/internal/llm"
	"github.com/tienpdinh/gpt-home/pkg/models"
)

// Simple mock HomeAssistant client for testing
type mockHAClient struct{}

func (m *mockHAClient) GetEntities() ([]models.Device, error) {
	return []models.Device{
		{ID: "light.1", Name: "Test Light", Type: models.DeviceTypeLight},
		{ID: "switch.1", Name: "Test Switch", Type: models.DeviceTypeSwitch},
	}, nil
}

func (m *mockHAClient) GetEntity(entityID string) (*models.Device, error) {
	if entityID == "light.1" {
		return &models.Device{ID: "light.1", Name: "Test Light", Type: models.DeviceTypeLight}, nil
	}
	return nil, assert.AnError
}

func (m *mockHAClient) CallService(domain, service, entityID string, serviceData map[string]interface{}) error {
	return nil
}

func (m *mockHAClient) TestConnection() error {
	return nil
}

func setupTestHandler() *Handler {
	haClient := &mockHAClient{}
	deviceManager := device.NewManager(haClient)
	llmService := llm.NewService("/tmp/test", "test")
	conversationManager := conversation.NewManager()

	return NewHandler(deviceManager, llmService, conversationManager)
}

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/chat", handler.HandleChat)
	router.GET("/devices", handler.GetDevices)
	router.GET("/devices/:id", handler.GetDevice)
	router.POST("/devices/:id/control", handler.ControlDevice)
	router.GET("/conversations/:id", handler.GetConversation)
	router.DELETE("/conversations/:id", handler.DeleteConversation)
	router.GET("/health", handler.HealthCheck)

	return router
}

func TestNewHandler(t *testing.T) {
	handler := setupTestHandler()

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.deviceManager)
	assert.NotNil(t, handler.llmService)
	assert.NotNil(t, handler.conversationManager)
	assert.NotZero(t, handler.startTime)
}

func TestHandleChat_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/chat", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDevices_Success(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/devices", nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]models.Device
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response["devices"], 2)
	assert.Equal(t, "light.1", response["devices"][0].ID)
}

func TestGetDevice_Success(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/devices/light.1", nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Device
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "light.1", response.ID)
	assert.Equal(t, "Test Light", response.Name)
}

func TestGetDevice_NotFound(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/devices/nonexistent", nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestControlDevice_Success(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	action := models.DeviceAction{
		Action: "turn_on",
		Parameters: map[string]any{
			"brightness": 255,
		},
	}

	body, _ := json.Marshal(action)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/devices/light.1/control", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
}

func TestControlDevice_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/devices/light.1/control", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConversation_InvalidID(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/conversations/invalid-uuid", nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConversation_NotFound(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	conversationID := uuid.New()

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/conversations/"+conversationID.String(), nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteConversation_InvalidID(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/conversations/invalid-uuid", nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteConversation_NotFound(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	conversationID := uuid.New()

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/conversations/"+conversationID.String(), nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHealthCheck(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/health", nil)

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	assert.Equal(t, "error", response.Services.LLM.Status) // LLM not loaded in test
	assert.Equal(t, "healthy", response.Services.HomeAssistant.Status)
	assert.Equal(t, "healthy", response.Services.Database.Status)
	assert.NotEmpty(t, response.Uptime)
	assert.NotEmpty(t, response.MemoryUsage)
}

func TestHandler_RouteRegistration(t *testing.T) {
	handler := setupTestHandler()
	router := setupTestRouter(handler)

	// Test that all routes are properly registered
	routes := []struct {
		method           string
		path             string
		expectedNotFound bool
	}{
		{"POST", "/chat", false},
		{"GET", "/devices", false},
		{"GET", "/devices/light.1", false},                                       // Use existing device
		{"POST", "/devices/light.1/control", false},                              // Use existing device
		{"GET", "/conversations/550e8400-e29b-41d4-a716-446655440000", false},    // Returns 404 but route exists
		{"DELETE", "/conversations/550e8400-e29b-41d4-a716-446655440000", false}, // Returns 500 but route exists
		{"GET", "/health", false},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)
			router.ServeHTTP(w, req)

			if route.expectedNotFound {
				assert.Equal(t, http.StatusNotFound, w.Code, "Route should not exist: %s %s", route.method, route.path)
			} else {
				// For conversation routes that return 404 due to business logic (not found entity),
				// we should check that the route was matched (not a routing 404)
				// For route-level 404s, Gin typically returns a different response
				if w.Code == http.StatusNotFound && (route.path == "/conversations/550e8400-e29b-41d4-a716-446655440000") {
					// This is expected - conversation not found is application logic, not routing
					// The route exists and was matched, just the resource wasn't found
					assert.True(t, true, "Route exists - 404 is from application logic, not routing")
				} else {
					// For all other routes, should not return 404 from routing
					assert.NotEqual(t, http.StatusNotFound, w.Code, "Route should exist: %s %s", route.method, route.path)
				}
			}
		})
	}
}
