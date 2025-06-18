package conversation

import (
	"fmt"
	"sync"
	"time"

	"github.com/tienpdinh/gpt-home/pkg/models"

	"github.com/google/uuid"
)

type Manager struct {
	conversations map[uuid.UUID]*models.Conversation
	mutex         sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		conversations: make(map[uuid.UUID]*models.Conversation),
	}
}

func (m *Manager) CreateConversation() *models.Conversation {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	conv := &models.Conversation{
		ID:        uuid.New(),
		Messages:  []models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Context: models.Context{
			ReferencedDevices: []string{},
			UserPreferences:   make(map[string]string),
			SessionData:       make(map[string]any),
		},
	}

	m.conversations[conv.ID] = conv
	return conv
}

func (m *Manager) GetConversation(id uuid.UUID) (*models.Conversation, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conv, exists := m.conversations[id]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", id)
	}

	return conv, nil
}

func (m *Manager) UpdateConversation(conv *models.Conversation) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.conversations[conv.ID]; !exists {
		return fmt.Errorf("conversation not found: %s", conv.ID)
	}

	conv.UpdatedAt = time.Now()
	m.conversations[conv.ID] = conv
	return nil
}

func (m *Manager) DeleteConversation(id uuid.UUID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.conversations[id]; !exists {
		return fmt.Errorf("conversation not found: %s", id)
	}

	delete(m.conversations, id)
	return nil
}

func (m *Manager) GetAllConversations() []*models.Conversation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conversations := make([]*models.Conversation, 0, len(m.conversations))
	for _, conv := range m.conversations {
		conversations = append(conversations, conv)
	}

	return conversations
}

func (m *Manager) AddMessage(conversationID uuid.UUID, message models.Message) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	conv, exists := m.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	conv.Messages = append(conv.Messages, message)
	conv.UpdatedAt = time.Now()
	return nil
}

func (m *Manager) UpdateContext(conversationID uuid.UUID, context models.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	conv, exists := m.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	conv.Context = context
	conv.UpdatedAt = time.Now()
	return nil
}

func (m *Manager) GetRecentMessages(conversationID uuid.UUID, limit int) ([]models.Message, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conv, exists := m.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	messages := conv.Messages
	if len(messages) <= limit {
		return messages, nil
	}

	return messages[len(messages)-limit:], nil
}

func (m *Manager) CleanupOldConversations(maxAge time.Duration) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	deleted := 0

	for id, conv := range m.conversations {
		if conv.UpdatedAt.Before(cutoff) {
			delete(m.conversations, id)
			deleted++
		}
	}

	return deleted
}

func (m *Manager) GetConversationStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	totalMessages := 0
	for _, conv := range m.conversations {
		totalMessages += len(conv.Messages)
	}

	return map[string]interface{}{
		"total_conversations": len(m.conversations),
		"total_messages":      totalMessages,
	}
}
