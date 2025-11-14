package database

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tienpdinh/gpt-home/pkg/models"
)

func TestDBSaveAndGetConversation(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	// Create a test conversation
	convID := uuid.New()
	conv := &models.Conversation{
		ID:        convID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []models.Message{
			{
				ID:        uuid.New(),
				Role:      models.MessageRoleUser,
				Content:   "Turn on the lights",
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				Role:      models.MessageRoleAssistant,
				Content:   "Lights turned on",
				Timestamp: time.Now().Add(time.Second),
			},
		},
		Context: models.Context{
			ReferencedDevices: []string{"light_bedroom"},
			UserPreferences:   map[string]string{"brightness": "50%"},
			SessionData:       map[string]any{"last_device": "light_bedroom"},
		},
	}

	// Save conversation
	err = db.SaveConversation(conv)
	require.NoError(t, err)

	// Get conversation
	retrieved, err := db.GetConversation(convID)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, conv.ID, retrieved.ID)
	assert.Equal(t, len(conv.Messages), len(retrieved.Messages))
	assert.Equal(t, conv.Messages[0].Content, retrieved.Messages[0].Content)
	assert.Equal(t, conv.Context.ReferencedDevices, retrieved.Context.ReferencedDevices)
}

func TestDBGetNonexistentConversation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	_, err = db.GetConversation(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDBDeleteConversation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	// Save a conversation
	convID := uuid.New()
	conv := &models.Conversation{
		ID:        convID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []models.Message{},
		Context: models.Context{
			UserPreferences: make(map[string]string),
			SessionData:     make(map[string]any),
		},
	}

	err = db.SaveConversation(conv)
	require.NoError(t, err)

	// Verify it exists
	_, err = db.GetConversation(convID)
	require.NoError(t, err)

	// Delete it
	err = db.DeleteConversation(convID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = db.GetConversation(convID)
	assert.Error(t, err)
}

func TestDBGetAllConversations(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	// Save multiple conversations
	for i := 0; i < 3; i++ {
		conv := &models.Conversation{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []models.Message{},
			Context: models.Context{
				UserPreferences: make(map[string]string),
				SessionData:     make(map[string]any),
			},
		}
		err = db.SaveConversation(conv)
		require.NoError(t, err)
	}

	// Get all conversations
	conversations, err := db.GetAllConversations()
	require.NoError(t, err)

	assert.Equal(t, 3, len(conversations))
}

func TestDBUpdateConversation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	// Create and save initial conversation
	convID := uuid.New()
	conv := &models.Conversation{
		ID:        convID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []models.Message{
			{
				ID:        uuid.New(),
				Role:      models.MessageRoleUser,
				Content:   "Hello",
				Timestamp: time.Now(),
			},
		},
		Context: models.Context{
			UserPreferences: make(map[string]string),
			SessionData:     make(map[string]any),
		},
	}

	err = db.SaveConversation(conv)
	require.NoError(t, err)

	// Add a new message and update
	newMsg := models.Message{
		ID:        uuid.New(),
		Role:      models.MessageRoleAssistant,
		Content:   "Hi there!",
		Timestamp: time.Now().Add(time.Second),
	}
	conv.Messages = append(conv.Messages, newMsg)

	err = db.SaveConversation(conv)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := db.GetConversation(convID)
	require.NoError(t, err)

	assert.Equal(t, 2, len(retrieved.Messages))
	assert.Equal(t, "Hi there!", retrieved.Messages[1].Content)
}
