package device

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tienpdinh/gpt-home/pkg/models"
	"github.com/tienpdinh/gpt-home/test/mocks"
)

func TestNewManager(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.haClient)
	assert.NotNil(t, manager.devices)
	assert.Empty(t, manager.devices)
}

func TestGetAllDevices(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	// Test initial call (should fetch from HA)
	devices, err := manager.GetAllDevices()
	require.NoError(t, err)
	assert.Greater(t, len(devices), 0)

	// Verify devices are cached
	assert.NotEmpty(t, manager.devices)

	// Test cached call (within 30 seconds)
	devices2, err := manager.GetAllDevices()
	require.NoError(t, err)
	assert.Equal(t, len(devices), len(devices2))
}

func TestGetAllDevicesWithConnectionError(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	mockClient.SetConnectionError(true)
	manager := NewManager(mockClient)

	devices, err := manager.GetAllDevices()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection error")
	assert.Empty(t, devices)
}

func TestGetDevice(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	// Test getting existing device
	device, err := manager.GetDevice("light.living_room")
	require.NoError(t, err)
	assert.Equal(t, "light.living_room", device.ID)
	assert.Equal(t, "Living Room Light", device.Name)
	assert.Equal(t, models.DeviceTypeLight, device.Type)

	// Test getting non-existent device
	_, err = manager.GetDevice("nonexistent.device")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device not found")
}

func TestRefreshDevices(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	// Add a custom device to mock client
	customDevice := models.Device{
		ID:       "light.custom",
		Name:     "Custom Light",
		Type:     models.DeviceTypeLight,
		State:    "on",
		Domain:   "light",
		EntityID: "light.custom",
	}
	mockClient.AddMockEntity(customDevice)

	err := manager.RefreshDevices()
	require.NoError(t, err)

	// Verify the custom device is now available
	device, err := manager.GetDevice("light.custom")
	require.NoError(t, err)
	assert.Equal(t, "Custom Light", device.Name)
}

func TestRefreshDevicesWithError(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	mockClient.SetConnectionError(true)
	manager := NewManager(mockClient)

	err := manager.RefreshDevices()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch devices")
}

func TestExecuteActionOnDevice(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	tests := []struct {
		name     string
		deviceID string
		action   models.DeviceAction
		wantErr  bool
	}{
		{
			name:     "turn on light",
			deviceID: "light.living_room",
			action: models.DeviceAction{
				Action:     "turn_on",
				Parameters: map[string]any{},
			},
			wantErr: false,
		},
		{
			name:     "turn off light",
			deviceID: "light.bedroom",
			action: models.DeviceAction{
				Action:     "turn_off",
				Parameters: map[string]any{},
			},
			wantErr: false,
		},
		{
			name:     "set brightness",
			deviceID: "light.living_room",
			action: models.DeviceAction{
				Action: "set_brightness",
				Parameters: map[string]any{
					"brightness": 128,
				},
			},
			wantErr: false,
		},
		{
			name:     "set temperature",
			deviceID: "climate.main",
			action: models.DeviceAction{
				Action: "set_temperature",
				Parameters: map[string]any{
					"temperature": 24.0,
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid device",
			deviceID: "nonexistent.device",
			action: models.DeviceAction{
				Action:     "turn_on",
				Parameters: map[string]any{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ExecuteActionOnDevice(tt.deviceID, tt.action)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteActionOnDeviceWithServiceError(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	mockClient.SetServiceError(true)
	manager := NewManager(mockClient)

	action := models.DeviceAction{
		Action:     "turn_on",
		Parameters: map[string]any{},
	}

	err := manager.ExecuteActionOnDevice("light.living_room", action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute action")
}

func TestFindDevicesByName(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	// Populate cache
	_, err := manager.GetAllDevices()
	require.NoError(t, err)

	tests := []struct {
		name             string
		searchName       string
		expectedMinCount int
	}{
		{
			name:             "find light devices",
			searchName:       "light",
			expectedMinCount: 2, // living room and bedroom lights
		},
		{
			name:             "find living room",
			searchName:       "living room",
			expectedMinCount: 1,
		},
		{
			name:             "case insensitive search",
			searchName:       "BEDROOM",
			expectedMinCount: 1,
		},
		{
			name:             "no matches",
			searchName:       "nonexistent",
			expectedMinCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices := manager.FindDevicesByName(tt.searchName)
			assert.GreaterOrEqual(t, len(devices), tt.expectedMinCount)
		})
	}
}

func TestFindDevicesByType(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	// Populate cache
	_, err := manager.GetAllDevices()
	require.NoError(t, err)

	tests := []struct {
		name             string
		deviceType       models.DeviceType
		expectedMinCount int
	}{
		{
			name:             "find lights",
			deviceType:       models.DeviceTypeLight,
			expectedMinCount: 2,
		},
		{
			name:             "find switches",
			deviceType:       models.DeviceTypeSwitch,
			expectedMinCount: 1,
		},
		{
			name:             "find climate devices",
			deviceType:       models.DeviceTypeClimate,
			expectedMinCount: 1,
		},
		{
			name:             "find sensors",
			deviceType:       models.DeviceTypeSensor,
			expectedMinCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices := manager.FindDevicesByType(tt.deviceType)
			assert.GreaterOrEqual(t, len(devices), tt.expectedMinCount)

			// Verify all returned devices are of the correct type
			for _, device := range devices {
				assert.Equal(t, tt.deviceType, device.Type)
			}
		})
	}
}

func TestIsConnected(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	// Test successful connection
	assert.True(t, manager.IsConnected())

	// Test connection failure
	mockClient.SetConnectionError(true)
	assert.False(t, manager.IsConnected())
}

func TestMapActionToService(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	tests := []struct {
		name            string
		device          *models.Device
		action          models.DeviceAction
		expectedDomain  string
		expectedService string
		expectedData    map[string]interface{}
	}{
		{
			name:            "light turn on",
			device:          &models.Device{Type: models.DeviceTypeLight},
			action:          models.DeviceAction{Action: "turn_on"},
			expectedDomain:  "light",
			expectedService: "turn_on",
			expectedData:    map[string]interface{}{},
		},
		{
			name:   "light set brightness",
			device: &models.Device{Type: models.DeviceTypeLight},
			action: models.DeviceAction{
				Action:     "set_brightness",
				Parameters: map[string]any{"brightness": 255},
			},
			expectedDomain:  "light",
			expectedService: "turn_on",
			expectedData:    map[string]interface{}{"brightness": 255},
		},
		{
			name:            "switch toggle",
			device:          &models.Device{Type: models.DeviceTypeSwitch},
			action:          models.DeviceAction{Action: "toggle"},
			expectedDomain:  "switch",
			expectedService: "toggle",
			expectedData:    map[string]interface{}{},
		},
		{
			name:   "climate set temperature",
			device: &models.Device{Type: models.DeviceTypeClimate},
			action: models.DeviceAction{
				Action:     "set_temperature",
				Parameters: map[string]any{"temperature": 22.5},
			},
			expectedDomain:  "climate",
			expectedService: "set_temperature",
			expectedData:    map[string]interface{}{"temperature": 22.5},
		},
		{
			name:            "cover open",
			device:          &models.Device{Type: models.DeviceTypeCover},
			action:          models.DeviceAction{Action: "open"},
			expectedDomain:  "cover",
			expectedService: "open_cover",
			expectedData:    map[string]interface{}{},
		},
		{
			name:   "fan set speed",
			device: &models.Device{Type: models.DeviceTypeFan},
			action: models.DeviceAction{
				Action:     "set_speed",
				Parameters: map[string]any{"percentage": 75},
			},
			expectedDomain:  "fan",
			expectedService: "set_percentage",
			expectedData:    map[string]interface{}{"percentage": 75},
		},
		{
			name:   "media player volume",
			device: &models.Device{Type: models.DeviceTypeMedia},
			action: models.DeviceAction{
				Action:     "volume_set",
				Parameters: map[string]any{"volume_level": 0.8},
			},
			expectedDomain:  "media_player",
			expectedService: "volume_set",
			expectedData:    map[string]interface{}{"volume_level": 0.8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, service, serviceData := manager.mapActionToService(tt.device, tt.action)

			assert.Equal(t, tt.expectedDomain, domain)
			assert.Equal(t, tt.expectedService, service)
			assert.Equal(t, tt.expectedData, serviceData)
		})
	}
}

func TestMapActionToServiceUnsupported(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	device := &models.Device{Type: models.DeviceTypeSensor}
	action := models.DeviceAction{Action: "turn_on"}

	domain, service, _ := manager.mapActionToService(device, action)
	assert.Empty(t, domain)
	assert.Empty(t, service)
}

func TestCacheExpiration(t *testing.T) {
	mockClient := mocks.NewMockHomeAssistantClient()
	manager := NewManager(mockClient)

	// First call populates cache
	devices1, err := manager.GetAllDevices()
	require.NoError(t, err)
	assert.NotEmpty(t, devices1)

	// Manually set last update time to simulate cache expiration
	manager.lastUpdate = time.Now().Add(-31 * time.Second)

	// Add a new device to mock client
	newDevice := models.Device{
		ID:       "light.new",
		Name:     "New Light",
		Type:     models.DeviceTypeLight,
		State:    "off",
		Domain:   "light",
		EntityID: "light.new",
	}
	mockClient.AddMockEntity(newDevice)

	// Second call should refresh cache
	devices2, err := manager.GetAllDevices()
	require.NoError(t, err)
	assert.Greater(t, len(devices2), len(devices1))

	// Verify new device is available
	device, err := manager.GetDevice("light.new")
	require.NoError(t, err)
	assert.Equal(t, "New Light", device.Name)
}
