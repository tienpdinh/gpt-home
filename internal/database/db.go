package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"

	"github.com/tienpdinh/gpt-home/pkg/models"
)

type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes the schema
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

// initSchema creates the necessary tables
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		created_at DATETIME,
		updated_at DATETIME,
		context_data TEXT
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp DATETIME,
		metadata_data TEXT,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// SaveConversation saves a conversation and all its messages to the database
func (db *DB) SaveConversation(conv *models.Conversation) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Save conversation metadata
	contextJSON, err := json.Marshal(conv.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO conversations (id, created_at, updated_at, context_data)
		VALUES (?, ?, ?, ?)
	`, conv.ID.String(), conv.CreatedAt, conv.UpdatedAt, string(contextJSON))
	if err != nil {
		return fmt.Errorf("failed to save conversation: %w", err)
	}

	// Save messages
	for _, msg := range conv.Messages {
		metadataJSON, err := json.Marshal(msg.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		_, err = tx.Exec(`
			INSERT OR REPLACE INTO messages (id, conversation_id, role, content, timestamp, metadata_data)
			VALUES (?, ?, ?, ?, ?, ?)
		`, msg.ID.String(), conv.ID.String(), msg.Role, msg.Content, msg.Timestamp, string(metadataJSON))
		if err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetConversation retrieves a conversation by ID from the database
func (db *DB) GetConversation(id uuid.UUID) (*models.Conversation, error) {
	var contextJSON string
	var createdAt, updatedAt time.Time

	err := db.conn.QueryRow(`
		SELECT created_at, updated_at, context_data FROM conversations WHERE id = ?
	`, id.String()).Scan(&createdAt, &updatedAt, &contextJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	var context models.Context
	if err := json.Unmarshal([]byte(contextJSON), &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	// Get messages
	rows, err := db.conn.Query(`
		SELECT id, role, content, timestamp, metadata_data FROM messages
		WHERE conversation_id = ? ORDER BY timestamp ASC
	`, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	messages := []models.Message{}
	for rows.Next() {
		var msgID, role, content, metadataJSON string
		var timestamp time.Time

		if err := rows.Scan(&msgID, &role, &content, &timestamp, &metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msgUUID, err := uuid.Parse(msgID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse message ID: %w", err)
		}

		var metadata models.Metadata
		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
				logrus.Warnf("failed to unmarshal metadata: %v", err)
			}
		}

		messages = append(messages, models.Message{
			ID:        msgUUID,
			Role:      models.MessageRole(role),
			Content:   content,
			Timestamp: timestamp,
			Metadata:  metadata,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	convID, err := uuid.Parse(id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
	}

	return &models.Conversation{
		ID:        convID,
		Messages:  messages,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Context:   context,
	}, nil
}

// DeleteConversation deletes a conversation from the database
func (db *DB) DeleteConversation(id uuid.UUID) error {
	result, err := db.conn.Exec(`DELETE FROM conversations WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("conversation not found: %s", id)
	}

	return nil
}

// GetAllConversations retrieves all conversations from the database
func (db *DB) GetAllConversations() ([]*models.Conversation, error) {
	rows, err := db.conn.Query(`
		SELECT id, created_at, updated_at, context_data FROM conversations
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	conversations := []*models.Conversation{}
	for rows.Next() {
		var id, contextJSON string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &createdAt, &updatedAt, &contextJSON); err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		convID, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation ID: %w", err)
		}

		var context models.Context
		if err := json.Unmarshal([]byte(contextJSON), &context); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context: %w", err)
		}

		// Get messages for this conversation
		msgRows, err := db.conn.Query(`
			SELECT id, role, content, timestamp, metadata_data FROM messages
			WHERE conversation_id = ? ORDER BY timestamp ASC
		`, id)
		if err != nil {
			return nil, fmt.Errorf("failed to query messages: %w", err)
		}

		messages := []models.Message{}
		for msgRows.Next() {
			var msgID, role, content, metadataJSON string
			var timestamp time.Time

			if err := msgRows.Scan(&msgID, &role, &content, &timestamp, &metadataJSON); err != nil {
				msgRows.Close()
				return nil, fmt.Errorf("failed to scan message: %w", err)
			}

			msgUUID, err := uuid.Parse(msgID)
			if err != nil {
				msgRows.Close()
				return nil, fmt.Errorf("failed to parse message ID: %w", err)
			}

			var metadata models.Metadata
			if metadataJSON != "" {
				if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
					logrus.Warnf("failed to unmarshal metadata: %v", err)
				}
			}

			messages = append(messages, models.Message{
				ID:        msgUUID,
				Role:      models.MessageRole(role),
				Content:   content,
				Timestamp: timestamp,
				Metadata:  metadata,
			})
		}
		msgRows.Close()

		if err = msgRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating messages: %w", err)
		}

		conversations = append(conversations, &models.Conversation{
			ID:        convID,
			Messages:  messages,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Context:   context,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversations: %w", err)
	}

	return conversations, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
