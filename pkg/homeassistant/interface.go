package homeassistant

import "github.com/tienpdinh/gpt-home/pkg/models"

// ClientInterface defines the interface for HomeAssistant clients
type ClientInterface interface {
	GetEntities() ([]models.Device, error)
	GetEntity(entityID string) (*models.Device, error)
	CallService(domain, service, entityID string, serviceData map[string]interface{}) error
	TestConnection() error
}
