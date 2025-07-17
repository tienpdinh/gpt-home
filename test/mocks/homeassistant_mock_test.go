package mocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tienpdinh/gpt-home/pkg/models"
)

func TestNewMockHomeAssistantClient(t *testing.T) {
	client := NewMockHomeAssistantClient()

	assert.NotNil(t, client)
	assert.False(t, client.connectionError)
	assert.False(t, client.serviceError)
	assert.NotEmpty(t, client.entities)
}

func TestSetConnectionError(t *testing.T) {
	client := NewMockHomeAssistantClient()

	// Test enabling connection error
	client.SetConnectionError(true)
	assert.True(t, client.connectionError)

	// Test disabling connection error
	client.SetConnectionError(false)
	assert.False(t, client.connectionError)
}

func TestSetServiceError(t *testing.T) {
	client := NewMockHomeAssistantClient()

	// Test enabling service error
	client.SetServiceError(true)
	assert.True(t, client.serviceError)

	// Test disabling service error
	client.SetServiceError(false)
	assert.False(t, client.serviceError)
}

func TestGetEntities(t *testing.T) {
	client := NewMockHomeAssistantClient()

	// Test normal operation
	entities, err := client.GetEntities()
	assert.NoError(t, err)
	assert.NotEmpty(t, entities)

	// Test with connection error
	client.SetConnectionError(true)
	entities, err = client.GetEntities()
	assert.Error(t, err)
	assert.Nil(t, entities)
}

func TestTestConnection(t *testing.T) {
	client := NewMockHomeAssistantClient()

	// Test normal operation
	err := client.TestConnection()
	assert.NoError(t, err)

	// Test with connection error
	client.SetConnectionError(true)
	err = client.TestConnection()
	assert.Error(t, err)
}

func TestGetEntity(t *testing.T) {
	client := NewMockHomeAssistantClient()

	// Test getting existing entity
	entity, err := client.GetEntity("light.living_room")
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "light.living_room", entity.ID)

	// Test getting non-existent entity
	entity, err = client.GetEntity("nonexistent.entity")
	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "entity not found")

	// Test with connection error
	client.SetConnectionError(true)
	entity, err = client.GetEntity("light.living_room")
	assert.Error(t, err)
	assert.Nil(t, entity)
}

func TestCallService(t *testing.T) {
	client := NewMockHomeAssistantClient()

	// Test normal service call
	err := client.CallService("light", "turn_on", "light.living_room", nil)
	assert.NoError(t, err)

	// Test with connection error
	client.SetConnectionError(true)
	err = client.CallService("light", "turn_on", "light.living_room", nil)
	assert.Error(t, err)

	// Reset and test with service error
	client.SetConnectionError(false)
	client.SetServiceError(true)
	err = client.CallService("light", "turn_on", "light.living_room", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error")
}

func TestAddMockEntity(t *testing.T) {
	client := NewMockHomeAssistantClient()
	initialCount := len(client.entities)

	newEntity := models.Device{
		ID:       "switch.new_switch",
		Name:     "New Switch",
		Type:     models.DeviceTypeSwitch,
		EntityID: "switch.new_switch",
	}

	client.AddMockEntity(newEntity)
	assert.Len(t, client.entities, initialCount+1)

	// Verify entity was added
	entity, err := client.GetEntity("switch.new_switch")
	assert.NoError(t, err)
	assert.Equal(t, "New Switch", entity.Name)
}
