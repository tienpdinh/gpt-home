package conversation

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tienpdinh/gpt-home/pkg/models"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.conversations)
	assert.Empty(t, manager.conversations)
}

func TestCreateConversation(t *testing.T) {
	manager := NewManager()

	conv := manager.CreateConversation()

	assert.NotNil(t, conv)
	assert.NotEqual(t, uuid.Nil, conv.ID)
	assert.Empty(t, conv.Messages)
	assert.NotZero(t, conv.CreatedAt)
	assert.NotZero(t, conv.UpdatedAt)
	assert.NotNil(t, conv.Context.ReferencedDevices)
	assert.NotNil(t, conv.Context.UserPreferences)
	assert.NotNil(t, conv.Context.SessionData)

	// Verify conversation is stored in manager
	storedConv, err := manager.GetConversation(conv.ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, storedConv.ID)
}

func TestGetConversation(t *testing.T) {
	manager := NewManager()

	// Test getting non-existent conversation
	nonExistentID := uuid.New()
	_, err := manager.GetConversation(nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversation not found")

	// Test getting existing conversation
	conv := manager.CreateConversation()
	retrievedConv, err := manager.GetConversation(conv.ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, retrievedConv.ID)
	assert.Equal(t, conv.CreatedAt, retrievedConv.CreatedAt)
}

func TestUpdateConversation(t *testing.T) {
	manager := NewManager()
	conv := manager.CreateConversation()

	// Update conversation
	originalUpdateTime := conv.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure time difference
	conv.Messages = append(conv.Messages, models.Message{
		ID:        uuid.New(),
		Role:      models.MessageRoleUser,
		Content:   "test message",
		Timestamp: time.Now(),
	})

	err := manager.UpdateConversation(conv)
	require.NoError(t, err)

	// Verify update
	updatedConv, err := manager.GetConversation(conv.ID)
	require.NoError(t, err)
	assert.Len(t, updatedConv.Messages, 1)
	assert.Equal(t, "test message", updatedConv.Messages[0].Content)
	assert.True(t, updatedConv.UpdatedAt.After(originalUpdateTime))

	// Test updating non-existent conversation
	nonExistentConv := &models.Conversation{ID: uuid.New()}
	err = manager.UpdateConversation(nonExistentConv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversation not found")
}

func TestDeleteConversation(t *testing.T) {
	manager := NewManager()
	conv := manager.CreateConversation()

	// Delete existing conversation
	err := manager.DeleteConversation(conv.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = manager.GetConversation(conv.ID)
	assert.Error(t, err)

	// Test deleting non-existent conversation
	err = manager.DeleteConversation(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversation not found")
}

func TestGetAllConversations(t *testing.T) {
	manager := NewManager()

	// Test empty manager
	conversations := manager.GetAllConversations()
	assert.Empty(t, conversations)

	// Create multiple conversations
	conv1 := manager.CreateConversation()
	conv2 := manager.CreateConversation()
	conv3 := manager.CreateConversation()

	conversations = manager.GetAllConversations()
	assert.Len(t, conversations, 3)

	// Verify all conversations are present
	ids := make(map[uuid.UUID]bool)
	for _, conv := range conversations {
		ids[conv.ID] = true
	}
	assert.True(t, ids[conv1.ID])
	assert.True(t, ids[conv2.ID])
	assert.True(t, ids[conv3.ID])
}

func TestAddMessage(t *testing.T) {
	manager := NewManager()
	conv := manager.CreateConversation()

	message := models.Message{
		ID:        uuid.New(),
		Role:      models.MessageRoleUser,
		Content:   "Hello, GPT-Home!",
		Timestamp: time.Now(),
	}

	err := manager.AddMessage(conv.ID, message)
	require.NoError(t, err)

	// Verify message was added
	updatedConv, err := manager.GetConversation(conv.ID)
	require.NoError(t, err)
	assert.Len(t, updatedConv.Messages, 1)
	assert.Equal(t, message.Content, updatedConv.Messages[0].Content)
	assert.Equal(t, message.Role, updatedConv.Messages[0].Role)

	// Test adding message to non-existent conversation
	err = manager.AddMessage(uuid.New(), message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversation not found")
}

func TestUpdateContext(t *testing.T) {
	manager := NewManager()
	conv := manager.CreateConversation()

	newContext := models.Context{
		ReferencedDevices: []string{"light.living_room", "switch.kitchen"},
		LastAction: &models.DeviceAction{
			Action: "turn_on",
			Parameters: map[string]any{
				"brightness": 255,
			},
		},
		UserPreferences: map[string]string{
			"preferred_brightness": "80",
		},
		SessionData: map[string]any{
			"session_start": time.Now(),
		},
	}

	err := manager.UpdateContext(conv.ID, newContext)
	require.NoError(t, err)

	// Verify context was updated
	updatedConv, err := manager.GetConversation(conv.ID)
	require.NoError(t, err)
	assert.Equal(t, newContext.ReferencedDevices, updatedConv.Context.ReferencedDevices)
	assert.Equal(t, newContext.LastAction.Action, updatedConv.Context.LastAction.Action)
	assert.Equal(t, newContext.UserPreferences["preferred_brightness"], updatedConv.Context.UserPreferences["preferred_brightness"])

	// Test updating context for non-existent conversation
	err = manager.UpdateContext(uuid.New(), newContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversation not found")
}

func TestGetRecentMessages(t *testing.T) {
	manager := NewManager()
	conv := manager.CreateConversation()

	// Add multiple messages
	for i := 0; i < 5; i++ {
		message := models.Message{
			ID:        uuid.New(),
			Role:      models.MessageRoleUser,
			Content:   "Message " + string(rune('A'+i)),
			Timestamp: time.Now(),
		}
		err := manager.AddMessage(conv.ID, message)
		require.NoError(t, err)
	}

	// Test getting recent messages within limit
	recentMessages, err := manager.GetRecentMessages(conv.ID, 3)
	require.NoError(t, err)
	assert.Len(t, recentMessages, 3)
	assert.Equal(t, "Message C", recentMessages[0].Content) // Should be the 3rd message
	assert.Equal(t, "Message E", recentMessages[2].Content) // Should be the last message

	// Test getting all messages when limit is larger
	allMessages, err := manager.GetRecentMessages(conv.ID, 10)
	require.NoError(t, err)
	assert.Len(t, allMessages, 5)

	// Test with non-existent conversation
	_, err = manager.GetRecentMessages(uuid.New(), 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversation not found")
}

func TestCleanupOldConversations(t *testing.T) {
	manager := NewManager()

	// Create conversations with different update times
	conv1 := manager.CreateConversation()
	conv2 := manager.CreateConversation()
	conv3 := manager.CreateConversation()

	// Manually set update times to simulate old conversations
	manager.conversations[conv1.ID].UpdatedAt = time.Now().Add(-2 * time.Hour)
	manager.conversations[conv2.ID].UpdatedAt = time.Now().Add(-30 * time.Minute)
	manager.conversations[conv3.ID].UpdatedAt = time.Now().Add(-5 * time.Minute)

	// Cleanup conversations older than 1 hour
	deleted := manager.CleanupOldConversations(1 * time.Hour)
	assert.Equal(t, 1, deleted) // Only conv1 should be deleted

	// Verify correct conversations remain
	_, err := manager.GetConversation(conv1.ID)
	assert.Error(t, err) // Should be deleted

	_, err = manager.GetConversation(conv2.ID)
	assert.NoError(t, err) // Should remain

	_, err = manager.GetConversation(conv3.ID)
	assert.NoError(t, err) // Should remain
}

func TestGetConversationStats(t *testing.T) {
	manager := NewManager()

	// Test empty stats
	stats := manager.GetConversationStats()
	assert.Equal(t, 0, stats["total_conversations"])
	assert.Equal(t, 0, stats["total_messages"])

	// Create conversations with messages
	conv1 := manager.CreateConversation()
	conv2 := manager.CreateConversation()

	// Add messages to conv1
	for i := 0; i < 3; i++ {
		message := models.Message{
			ID:        uuid.New(),
			Role:      models.MessageRoleUser,
			Content:   "Message",
			Timestamp: time.Now(),
		}
		manager.AddMessage(conv1.ID, message)
	}

	// Add messages to conv2
	for i := 0; i < 2; i++ {
		message := models.Message{
			ID:        uuid.New(),
			Role:      models.MessageRoleAssistant,
			Content:   "Response",
			Timestamp: time.Now(),
		}
		manager.AddMessage(conv2.ID, message)
	}

	stats = manager.GetConversationStats()
	assert.Equal(t, 2, stats["total_conversations"])
	assert.Equal(t, 5, stats["total_messages"])
}

func TestConcurrentAccess(t *testing.T) {
	manager := NewManager()
	conv := manager.CreateConversation()

	// Test concurrent read/write operations
	done := make(chan bool, 2)

	// Goroutine 1: Add messages
	go func() {
		for i := 0; i < 10; i++ {
			message := models.Message{
				ID:        uuid.New(),
				Role:      models.MessageRoleUser,
				Content:   "Concurrent message",
				Timestamp: time.Now(),
			}
			manager.AddMessage(conv.ID, message)
		}
		done <- true
	}()

	// Goroutine 2: Read conversation
	go func() {
		for i := 0; i < 10; i++ {
			manager.GetConversation(conv.ID)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	finalConv, err := manager.GetConversation(conv.ID)
	require.NoError(t, err)
	assert.Len(t, finalConv.Messages, 10)
}
