package mocks

import (
	"fmt"
	"time"

	"github.com/tienpdinh/gpt-home/pkg/models"
)

// MockHomeAssistantClient is a mock implementation of the HomeAssistant client
type MockHomeAssistantClient struct {
	entities        []models.Device
	connectionError bool
	serviceError    bool
}

// NewMockHomeAssistantClient creates a new mock HomeAssistant client
func NewMockHomeAssistantClient() *MockHomeAssistantClient {
	return &MockHomeAssistantClient{
		entities:        createMockEntities(),
		connectionError: false,
		serviceError:    false,
	}
}

// SetConnectionError simulates connection failures
func (m *MockHomeAssistantClient) SetConnectionError(enabled bool) {
	m.connectionError = enabled
}

// SetServiceError simulates service call failures
func (m *MockHomeAssistantClient) SetServiceError(enabled bool) {
	m.serviceError = enabled
}

// GetEntities returns mock device entities
func (m *MockHomeAssistantClient) GetEntities() ([]models.Device, error) {
	if m.connectionError {
		return nil, fmt.Errorf("connection error: unable to connect to HomeAssistant")
	}
	return m.entities, nil
}

// GetEntity returns a specific mock entity
func (m *MockHomeAssistantClient) GetEntity(entityID string) (*models.Device, error) {
	if m.connectionError {
		return nil, fmt.Errorf("connection error: unable to connect to HomeAssistant")
	}

	for _, entity := range m.entities {
		if entity.ID == entityID {
			return &entity, nil
		}
	}
	return nil, fmt.Errorf("entity not found: %s", entityID)
}

// CallService simulates service calls
func (m *MockHomeAssistantClient) CallService(domain, service, entityID string, serviceData map[string]interface{}) error {
	if m.connectionError {
		return fmt.Errorf("connection error: unable to connect to HomeAssistant")
	}
	if m.serviceError {
		return fmt.Errorf("service error: failed to call %s.%s", domain, service)
	}

	// Update entity state based on service call
	for i, entity := range m.entities {
		if entity.ID == entityID {
			switch service {
			case "turn_on":
				m.entities[i].State = "on"
			case "turn_off":
				m.entities[i].State = "off"
			case "toggle":
				if entity.State == "on" {
					m.entities[i].State = "off"
				} else {
					m.entities[i].State = "on"
				}
			case "set_temperature":
				if temp, ok := serviceData["temperature"]; ok {
					m.entities[i].Attributes["temperature"] = temp
				}
			case "set_brightness":
				if brightness, ok := serviceData["brightness"]; ok {
					m.entities[i].Attributes["brightness"] = brightness
				}
			}
			m.entities[i].LastUpdated = time.Now()
			break
		}
	}

	return nil
}

// TestConnection simulates connection testing
func (m *MockHomeAssistantClient) TestConnection() error {
	if m.connectionError {
		return fmt.Errorf("connection test failed")
	}
	return nil
}

// AddMockEntity adds a new mock entity for testing
func (m *MockHomeAssistantClient) AddMockEntity(device models.Device) {
	m.entities = append(m.entities, device)
}

// UpdateMockEntity updates an existing mock entity
func (m *MockHomeAssistantClient) UpdateMockEntity(entityID string, updates map[string]interface{}) {
	for i, entity := range m.entities {
		if entity.ID == entityID {
			if state, ok := updates["state"].(string); ok {
				m.entities[i].State = state
			}
			if attributes, ok := updates["attributes"].(map[string]interface{}); ok {
				for key, value := range attributes {
					m.entities[i].Attributes[key] = value
				}
			}
			m.entities[i].LastUpdated = time.Now()
			break
		}
	}
}

// createMockEntities creates a set of mock entities for testing
func createMockEntities() []models.Device {
	return []models.Device{
		{
			ID:       "light.living_room",
			Name:     "Living Room Light",
			Type:     models.DeviceTypeLight,
			State:    "off",
			Domain:   "light",
			EntityID: "light.living_room",
			Attributes: map[string]any{
				"friendly_name": "Living Room Light",
				"brightness":    0,
				"color_mode":    "brightness",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:       "light.bedroom",
			Name:     "Bedroom Light",
			Type:     models.DeviceTypeLight,
			State:    "on",
			Domain:   "light",
			EntityID: "light.bedroom",
			Attributes: map[string]any{
				"friendly_name": "Bedroom Light",
				"brightness":    255,
				"color_mode":    "rgb",
				"rgb_color":     []int{255, 255, 255},
			},
			LastUpdated: time.Now(),
		},
		{
			ID:       "switch.porch",
			Name:     "Porch Switch",
			Type:     models.DeviceTypeSwitch,
			State:    "off",
			Domain:   "switch",
			EntityID: "switch.porch",
			Attributes: map[string]any{
				"friendly_name": "Porch Switch",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:       "climate.main",
			Name:     "Main Thermostat",
			Type:     models.DeviceTypeClimate,
			State:    "heat",
			Domain:   "climate",
			EntityID: "climate.main",
			Attributes: map[string]any{
				"friendly_name":       "Main Thermostat",
				"temperature":         22.0,
				"target_temp_high":    24.0,
				"target_temp_low":     20.0,
				"current_temperature": 21.5,
				"hvac_mode":           "heat",
				"hvac_modes":          []string{"off", "heat", "cool", "auto"},
			},
			LastUpdated: time.Now(),
		},
		{
			ID:       "sensor.temperature",
			Name:     "Temperature Sensor",
			Type:     models.DeviceTypeSensor,
			State:    "21.5",
			Domain:   "sensor",
			EntityID: "sensor.temperature",
			Attributes: map[string]any{
				"friendly_name":       "Temperature Sensor",
				"unit_of_measurement": "°C",
				"device_class":        "temperature",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:       "cover.garage_door",
			Name:     "Garage Door",
			Type:     models.DeviceTypeCover,
			State:    "closed",
			Domain:   "cover",
			EntityID: "cover.garage_door",
			Attributes: map[string]any{
				"friendly_name":    "Garage Door",
				"current_position": 0,
				"device_class":     "garage",
			},
			LastUpdated: time.Now(),
		},
		{
			ID:       "fan.ceiling",
			Name:     "Ceiling Fan",
			Type:     models.DeviceTypeFan,
			State:    "off",
			Domain:   "fan",
			EntityID: "fan.ceiling",
			Attributes: map[string]any{
				"friendly_name": "Ceiling Fan",
				"percentage":    0,
				"preset_modes":  []string{"low", "medium", "high"},
			},
			LastUpdated: time.Now(),
		},
		{
			ID:       "media_player.living_room",
			Name:     "Living Room Speaker",
			Type:     models.DeviceTypeMedia,
			State:    "idle",
			Domain:   "media_player",
			EntityID: "media_player.living_room",
			Attributes: map[string]any{
				"friendly_name":   "Living Room Speaker",
				"volume_level":    0.5,
				"is_volume_muted": false,
				"media_title":     "",
				"media_artist":    "",
			},
			LastUpdated: time.Now(),
		},
	}
}
