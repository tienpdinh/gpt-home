package homeassistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tienpdinh/gpt-home/pkg/models"

	"github.com/sirupsen/logrus"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type HAEntity struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
	Context     HAContext              `json:"context"`
}

type HAContext struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	UserID   string `json:"user_id"`
}

type HAServiceCall struct {
	Domain      string                 `json:"domain"`
	Service     string                 `json:"service"`
	Target      *HAServiceTarget       `json:"target,omitempty"`
	ServiceData map[string]interface{} `json:"service_data,omitempty"`
}

type HAServiceTarget struct {
	EntityID []string `json:"entity_id,omitempty"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetEntities() ([]models.Device, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/states", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var entities []HAEntity
	if err := json.NewDecoder(resp.Body).Decode(&entities); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	devices := make([]models.Device, 0, len(entities))
	for _, entity := range entities {
		device := c.convertEntityToDevice(entity)
		devices = append(devices, device)
	}

	return devices, nil
}

func (c *Client) GetEntity(entityID string) (*models.Device, error) {
	url := fmt.Sprintf("%s/api/states/%s", c.baseURL, entityID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("entity not found: %s", entityID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var entity HAEntity
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	device := c.convertEntityToDevice(entity)
	return &device, nil
}

func (c *Client) CallService(domain, service string, entityID string, serviceData map[string]interface{}) error {
	serviceCall := HAServiceCall{
		Domain:  domain,
		Service: service,
		Target: &HAServiceTarget{
			EntityID: []string{entityID},
		},
		ServiceData: serviceData,
	}

	jsonData, err := json.Marshal(serviceCall)
	if err != nil {
		return fmt.Errorf("failed to marshal service call: %w", err)
	}

	url := fmt.Sprintf("%s/api/services/%s/%s", c.baseURL, domain, service)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("service call failed with status %d: %s", resp.StatusCode, string(body))
	}

	logrus.Debugf("Successfully called service %s.%s for entity %s", domain, service, entityID)
	return nil
}

func (c *Client) TestConnection() error {
	req, err := http.NewRequest("GET", c.baseURL+"/api/", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to HomeAssistant: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HomeAssistant API returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) convertEntityToDevice(entity HAEntity) models.Device {
	// Parse domain from entity_id
	domain := ""
	if len(entity.EntityID) > 0 {
		for i, char := range entity.EntityID {
			if char == '.' {
				domain = entity.EntityID[:i]
				break
			}
		}
	}

	// Get friendly name from attributes
	name := entity.EntityID
	if friendlyName, ok := entity.Attributes["friendly_name"].(string); ok {
		name = friendlyName
	}

	// Convert domain to device type
	deviceType := c.domainToDeviceType(domain)

	// Parse last updated time
	lastUpdated := time.Now()
	if entity.LastUpdated != "" {
		if t, err := time.Parse(time.RFC3339, entity.LastUpdated); err == nil {
			lastUpdated = t
		}
	}

	return models.Device{
		ID:          entity.EntityID,
		Name:        name,
		Type:        deviceType,
		State:       entity.State,
		Attributes:  entity.Attributes,
		LastUpdated: lastUpdated,
		Domain:      domain,
		EntityID:    entity.EntityID,
	}
}

func (c *Client) domainToDeviceType(domain string) models.DeviceType {
	switch domain {
	case "light":
		return models.DeviceTypeLight
	case "switch":
		return models.DeviceTypeSwitch
	case "sensor", "binary_sensor":
		return models.DeviceTypeSensor
	case "climate":
		return models.DeviceTypeClimate
	case "cover":
		return models.DeviceTypeCover
	case "fan":
		return models.DeviceTypeFan
	case "media_player":
		return models.DeviceTypeMedia
	default:
		return models.DeviceTypeSensor
	}
}
