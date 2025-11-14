package device

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tienpdinh/gpt-home/pkg/homeassistant"
	"github.com/tienpdinh/gpt-home/pkg/models"

	"github.com/sirupsen/logrus"
)

type Manager struct {
	haClient     homeassistant.ClientInterface
	devices      map[string]models.Device
	devicesMutex sync.RWMutex
	lastUpdate   time.Time
	validator    *Validator
}

func NewManager(haClient homeassistant.ClientInterface) *Manager {
	return &Manager{
		haClient:  haClient,
		devices:   make(map[string]models.Device),
		validator: NewValidator(),
	}
}

func (m *Manager) GetAllDevices() ([]models.Device, error) {
	m.devicesMutex.RLock()

	// Refresh devices if cache is stale (older than 30 seconds)
	if time.Since(m.lastUpdate) > 30*time.Second {
		m.devicesMutex.RUnlock()
		if err := m.RefreshDevices(); err != nil {
			// If refresh fails and we have no cached data, return error
			m.devicesMutex.RLock()
			if len(m.devices) == 0 {
				m.devicesMutex.RUnlock()
				return nil, err
			}
			logrus.WithError(err).Warn("Failed to refresh devices, using cached data")
		}
		m.devicesMutex.RLock()
	}

	defer m.devicesMutex.RUnlock()

	devices := make([]models.Device, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

func (m *Manager) GetDevice(deviceID string) (*models.Device, error) {
	m.devicesMutex.RLock()
	device, exists := m.devices[deviceID]
	m.devicesMutex.RUnlock()

	if !exists {
		// Try to get fresh data from HomeAssistant
		freshDevice, err := m.haClient.GetEntity(deviceID)
		if err != nil {
			return nil, fmt.Errorf("device not found: %s", deviceID)
		}

		// Update cache
		m.devicesMutex.Lock()
		m.devices[deviceID] = *freshDevice
		m.devicesMutex.Unlock()

		return freshDevice, nil
	}

	return &device, nil
}

func (m *Manager) RefreshDevices() error {
	devices, err := m.haClient.GetEntities()
	if err != nil {
		return fmt.Errorf("failed to fetch devices from HomeAssistant: %w", err)
	}

	m.devicesMutex.Lock()
	defer m.devicesMutex.Unlock()

	// Clear existing devices
	m.devices = make(map[string]models.Device)

	// Add new devices
	for _, device := range devices {
		m.devices[device.ID] = device
	}

	m.lastUpdate = time.Now()
	logrus.Infof("Refreshed %d devices from HomeAssistant", len(devices))

	return nil
}

func (m *Manager) ExecuteAction(action models.DeviceAction) error {
	// This would need device context - for now, return error
	return fmt.Errorf("action execution requires device context")
}

func (m *Manager) ExecuteActionOnDevice(deviceID string, action models.DeviceAction) error {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	// Validate action before execution
	validationResult := m.validator.ValidateAction(&action)
	if !validationResult.Valid {
		return fmt.Errorf("action validation failed: %s", validationResult.Error)
	}

	if validationResult.Warning != "" {
		logrus.Warnf("Action warning for device %s: %s", deviceID, validationResult.Warning)
	}

	// Use the safe action from validation
	safeAction := validationResult.SafeAction

	// Map action to HomeAssistant service call
	domain, service, serviceData := m.mapActionToService(device, *safeAction)
	if domain == "" || service == "" {
		return fmt.Errorf("unsupported action %s for device type %s", safeAction.Action, device.Type)
	}

	// Execute the service call
	if err := m.haClient.CallService(domain, service, deviceID, serviceData); err != nil {
		return fmt.Errorf("failed to execute action: %w", err)
	}

	logrus.Infof("Executed action %s on device %s", safeAction.Action, deviceID)
	return nil
}

func (m *Manager) FindDevicesByName(name string) []models.Device {
	m.devicesMutex.RLock()
	defer m.devicesMutex.RUnlock()

	var matches []models.Device
	lowerName := strings.ToLower(name)

	for _, device := range m.devices {
		if strings.Contains(strings.ToLower(device.Name), lowerName) {
			matches = append(matches, device)
		}
	}

	return matches
}

func (m *Manager) FindDevicesByType(deviceType models.DeviceType) []models.Device {
	m.devicesMutex.RLock()
	defer m.devicesMutex.RUnlock()

	var matches []models.Device
	for _, device := range m.devices {
		if device.Type == deviceType {
			matches = append(matches, device)
		}
	}

	return matches
}

func (m *Manager) IsConnected() bool {
	return m.haClient.TestConnection() == nil
}

func (m *Manager) mapActionToService(device *models.Device, action models.DeviceAction) (domain, service string, serviceData map[string]interface{}) {
	serviceData = make(map[string]interface{})

	// Copy action parameters to service data
	for key, value := range action.Parameters {
		serviceData[key] = value
	}

	switch device.Type {
	case models.DeviceTypeLight:
		domain = "light"
		switch action.Action {
		case "turn_on":
			service = "turn_on"
		case "turn_off":
			service = "turn_off"
		case "toggle":
			service = "toggle"
		case "set_brightness":
			service = "turn_on"
			if brightness, ok := action.Parameters["brightness"]; ok {
				serviceData["brightness"] = brightness
			}
		case "set_color":
			service = "turn_on"
			// Handle RGB color setting
			if rgb, ok := action.Parameters["rgb_color"]; ok {
				serviceData["rgb_color"] = rgb
			}
		}

	case models.DeviceTypeSwitch:
		domain = "switch"
		switch action.Action {
		case "turn_on":
			service = "turn_on"
		case "turn_off":
			service = "turn_off"
		case "toggle":
			service = "toggle"
		}

	case models.DeviceTypeClimate:
		domain = "climate"
		switch action.Action {
		case "set_temperature":
			service = "set_temperature"
			if temp, ok := action.Parameters["temperature"]; ok {
				serviceData["temperature"] = temp
			}
		case "set_hvac_mode":
			service = "set_hvac_mode"
			if mode, ok := action.Parameters["hvac_mode"]; ok {
				serviceData["hvac_mode"] = mode
			}
		}

	case models.DeviceTypeCover:
		domain = "cover"
		switch action.Action {
		case "open":
			service = "open_cover"
		case "close":
			service = "close_cover"
		case "stop":
			service = "stop_cover"
		case "set_position":
			service = "set_cover_position"
			if position, ok := action.Parameters["position"]; ok {
				serviceData["position"] = position
			}
		}

	case models.DeviceTypeFan:
		domain = "fan"
		switch action.Action {
		case "turn_on":
			service = "turn_on"
		case "turn_off":
			service = "turn_off"
		case "toggle":
			service = "toggle"
		case "set_speed":
			service = "set_percentage"
			if speed, ok := action.Parameters["percentage"]; ok {
				serviceData["percentage"] = speed
			}
		}

	case models.DeviceTypeMedia:
		domain = "media_player"
		switch action.Action {
		case "play":
			service = "media_play"
		case "pause":
			service = "media_pause"
		case "stop":
			service = "media_stop"
		case "volume_set":
			service = "volume_set"
			if volume, ok := action.Parameters["volume_level"]; ok {
				serviceData["volume_level"] = volume
			}
		}
	}

	return domain, service, serviceData
}
