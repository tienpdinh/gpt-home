package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tienpdinh/gpt-home/pkg/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8123", "test-token")

	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8123", client.baseURL)
	assert.Equal(t, "test-token", client.token)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestGetEntities_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/states", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		entities := []HAEntity{
			{
				EntityID:    "light.living_room",
				State:       "on",
				Attributes:  map[string]interface{}{"friendly_name": "Living Room Light", "brightness": 255},
				LastChanged: "2023-01-01T12:00:00Z",
				LastUpdated: "2023-01-01T12:00:00Z",
			},
			{
				EntityID:    "switch.kitchen",
				State:       "off",
				Attributes:  map[string]interface{}{"friendly_name": "Kitchen Switch"},
				LastChanged: "2023-01-01T11:00:00Z",
				LastUpdated: "2023-01-01T11:00:00Z",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entities)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	devices, err := client.GetEntities()

	require.NoError(t, err)
	assert.Len(t, devices, 2)

	// Check first device (light)
	assert.Equal(t, "light.living_room", devices[0].ID)
	assert.Equal(t, "Living Room Light", devices[0].Name)
	assert.Equal(t, models.DeviceTypeLight, devices[0].Type)
	assert.Equal(t, "on", devices[0].State)
	assert.Equal(t, "light", devices[0].Domain)
	assert.Equal(t, float64(255), devices[0].Attributes["brightness"])

	// Check second device (switch)
	assert.Equal(t, "switch.kitchen", devices[1].ID)
	assert.Equal(t, "Kitchen Switch", devices[1].Name)
	assert.Equal(t, models.DeviceTypeSwitch, devices[1].Type)
	assert.Equal(t, "off", devices[1].State)
	assert.Equal(t, "switch", devices[1].Domain)
}

func TestGetEntities_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	devices, err := client.GetEntities()

	assert.Error(t, err)
	assert.Nil(t, devices)
	assert.Contains(t, err.Error(), "API request failed with status: 500")
}

func TestGetEntities_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	devices, err := client.GetEntities()

	assert.Error(t, err)
	assert.Nil(t, devices)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestGetEntity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/states/light.living_room", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		entity := HAEntity{
			EntityID:    "light.living_room",
			State:       "on",
			Attributes:  map[string]interface{}{"friendly_name": "Living Room Light", "brightness": 128},
			LastChanged: "2023-01-01T12:00:00Z",
			LastUpdated: "2023-01-01T12:00:00Z",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entity)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	device, err := client.GetEntity("light.living_room")

	require.NoError(t, err)
	require.NotNil(t, device)

	assert.Equal(t, "light.living_room", device.ID)
	assert.Equal(t, "Living Room Light", device.Name)
	assert.Equal(t, models.DeviceTypeLight, device.Type)
	assert.Equal(t, "on", device.State)
	assert.Equal(t, "light", device.Domain)
	assert.Equal(t, float64(128), device.Attributes["brightness"])
}

func TestGetEntity_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	device, err := client.GetEntity("nonexistent.entity")

	assert.Error(t, err)
	assert.Nil(t, device)
	assert.Contains(t, err.Error(), "entity not found: nonexistent.entity")
}

func TestGetEntity_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	device, err := client.GetEntity("light.living_room")

	assert.Error(t, err)
	assert.Nil(t, device)
	assert.Contains(t, err.Error(), "API request failed with status: 500")
}

func TestCallService_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/services/light/turn_on", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "POST", r.Method)

		var serviceCall HAServiceCall
		err := json.NewDecoder(r.Body).Decode(&serviceCall)
		require.NoError(t, err)

		assert.Equal(t, "light", serviceCall.Domain)
		assert.Equal(t, "turn_on", serviceCall.Service)
		assert.Equal(t, []string{"light.living_room"}, serviceCall.Target.EntityID)
		assert.Equal(t, float64(255), serviceCall.ServiceData["brightness"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	serviceData := map[string]interface{}{
		"brightness": 255,
	}

	err := client.CallService("light", "turn_on", "light.living_room", serviceData)
	assert.NoError(t, err)
}

func TestCallService_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	serviceData := map[string]interface{}{"brightness": 255}

	err := client.CallService("light", "turn_on", "light.living_room", serviceData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service call failed with status 400")
	assert.Contains(t, err.Error(), "Bad request")
}

func TestTestConnection_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "API running."}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.TestConnection()

	assert.NoError(t, err)
}

func TestTestConnection_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.TestConnection()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HomeAssistant API returned status: 401")
}

func TestTestConnection_NetworkError(t *testing.T) {
	client := NewClient("http://invalid:9999", "test-token")
	err := client.TestConnection()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to HomeAssistant")
}

func TestConvertEntityToDevice(t *testing.T) {
	client := NewClient("http://localhost", "token")

	testCases := []struct {
		name           string
		entity         HAEntity
		expectedDevice models.Device
	}{
		{
			name: "Light entity with friendly name",
			entity: HAEntity{
				EntityID:    "light.living_room",
				State:       "on",
				Attributes:  map[string]interface{}{"friendly_name": "Living Room Light", "brightness": 255},
				LastUpdated: "2023-01-01T12:00:00Z",
			},
			expectedDevice: models.Device{
				ID:         "light.living_room",
				Name:       "Living Room Light",
				Type:       models.DeviceTypeLight,
				State:      "on",
				Domain:     "light",
				EntityID:   "light.living_room",
				Attributes: map[string]interface{}{"friendly_name": "Living Room Light", "brightness": 255},
			},
		},
		{
			name: "Switch entity without friendly name",
			entity: HAEntity{
				EntityID:    "switch.kitchen",
				State:       "off",
				Attributes:  map[string]interface{}{},
				LastUpdated: "2023-01-01T11:00:00Z",
			},
			expectedDevice: models.Device{
				ID:         "switch.kitchen",
				Name:       "switch.kitchen",
				Type:       models.DeviceTypeSwitch,
				State:      "off",
				Domain:     "switch",
				EntityID:   "switch.kitchen",
				Attributes: map[string]interface{}{},
			},
		},
		{
			name: "Sensor entity",
			entity: HAEntity{
				EntityID:    "sensor.temperature",
				State:       "23.5",
				Attributes:  map[string]interface{}{"friendly_name": "Temperature Sensor", "unit_of_measurement": "°C"},
				LastUpdated: "2023-01-01T10:00:00Z",
			},
			expectedDevice: models.Device{
				ID:         "sensor.temperature",
				Name:       "Temperature Sensor",
				Type:       models.DeviceTypeSensor,
				State:      "23.5",
				Domain:     "sensor",
				EntityID:   "sensor.temperature",
				Attributes: map[string]interface{}{"friendly_name": "Temperature Sensor", "unit_of_measurement": "°C"},
			},
		},
		{
			name: "Unknown domain defaults to sensor",
			entity: HAEntity{
				EntityID:    "unknown.entity",
				State:       "state",
				Attributes:  map[string]interface{}{},
				LastUpdated: "2023-01-01T09:00:00Z",
			},
			expectedDevice: models.Device{
				ID:         "unknown.entity",
				Name:       "unknown.entity",
				Type:       models.DeviceTypeSensor,
				State:      "state",
				Domain:     "unknown",
				EntityID:   "unknown.entity",
				Attributes: map[string]interface{}{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			device := client.convertEntityToDevice(tc.entity)

			assert.Equal(t, tc.expectedDevice.ID, device.ID)
			assert.Equal(t, tc.expectedDevice.Name, device.Name)
			assert.Equal(t, tc.expectedDevice.Type, device.Type)
			assert.Equal(t, tc.expectedDevice.State, device.State)
			assert.Equal(t, tc.expectedDevice.Domain, device.Domain)
			assert.Equal(t, tc.expectedDevice.EntityID, device.EntityID)
			assert.Equal(t, tc.expectedDevice.Attributes, device.Attributes)
			assert.NotZero(t, device.LastUpdated)
		})
	}
}

func TestDomainToDeviceType(t *testing.T) {
	client := NewClient("http://localhost", "token")

	testCases := []struct {
		domain   string
		expected models.DeviceType
	}{
		{"light", models.DeviceTypeLight},
		{"switch", models.DeviceTypeSwitch},
		{"sensor", models.DeviceTypeSensor},
		{"binary_sensor", models.DeviceTypeSensor},
		{"climate", models.DeviceTypeClimate},
		{"cover", models.DeviceTypeCover},
		{"fan", models.DeviceTypeFan},
		{"media_player", models.DeviceTypeMedia},
		{"unknown", models.DeviceTypeSensor},
		{"", models.DeviceTypeSensor},
	}

	for _, tc := range testCases {
		t.Run(tc.domain, func(t *testing.T) {
			result := client.domainToDeviceType(tc.domain)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestClient_RequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all requests have proper headers
		assert.Equal(t, "Bearer test-token-123", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)

		// Return appropriate response based on endpoint
		if r.URL.Path == "/api/states" {
			// Return array for GetEntities
			w.Write([]byte(`[]`))
		} else {
			// Return single entity for GetEntity
			entity := HAEntity{
				EntityID:    "test.entity",
				State:       "on",
				Attributes:  map[string]interface{}{},
				LastUpdated: "2023-01-01T12:00:00Z",
			}
			json.NewEncoder(w).Encode(entity)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token-123")

	// Test GetEntities headers
	_, err := client.GetEntities()
	assert.NoError(t, err)

	// Test GetEntity headers
	_, err = client.GetEntity("test.entity")
	assert.NoError(t, err)
}

func TestClient_Timeout(t *testing.T) {
	// Create server that sleeps longer than client timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.httpClient.Timeout = 50 * time.Millisecond // Set very short timeout

	_, err := client.GetEntities()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
}

func TestClient_MalformedURL(t *testing.T) {
	client := NewClient("not-a-url", "test-token")

	_, err := client.GetEntities()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
}
