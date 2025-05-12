package openai

import (
	"sync"
	"time"
)

// Conversation represents a chat session with its history
type Conversation struct {
	Messages   []Message
	LastUpdate time.Time
}

// ConversationManager manages user conversations
type ConversationManager struct {
	conversations map[int64]*Conversation
	mutex         sync.RWMutex
	maxHistory    int
	ttl           time.Duration
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(maxHistory int, ttl time.Duration) *ConversationManager {
	manager := &ConversationManager{
		conversations: make(map[int64]*Conversation),
		maxHistory:    maxHistory,
		ttl:           ttl,
	}

	// Start a cleanup goroutine
	go manager.cleanup()

	return manager
}

// GetConversation retrieves a user's conversation
func (m *ConversationManager) GetConversation(userID int64) *Conversation {
	m.mutex.RLock()
	conv, exists := m.conversations[userID]
	m.mutex.RUnlock()

	if !exists || time.Since(conv.LastUpdate) > m.ttl {
		// Create a new conversation if none exists or if it's expired
		m.mutex.Lock()
		m.conversations[userID] = &Conversation{
			Messages:   []Message{},
			LastUpdate: time.Now(),
		}
		conv = m.conversations[userID]
		m.mutex.Unlock()
	}

	return conv
}

// AddMessage adds a message to the conversation
func (m *ConversationManager) AddMessage(userID int64, message Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	conv, exists := m.conversations[userID]
	if !exists {
		conv = &Conversation{
			Messages:   []Message{},
			LastUpdate: time.Now(),
		}
		m.conversations[userID] = conv
	}

	// Add the new message
	conv.Messages = append(conv.Messages, message)

	// Trim history if it exceeds the maximum
	if len(conv.Messages) > m.maxHistory {
		conv.Messages = conv.Messages[len(conv.Messages)-m.maxHistory:]
	}

	conv.LastUpdate = time.Now()
}

// ResetConversation clears the conversation history for a user
func (m *ConversationManager) ResetConversation(userID int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.conversations[userID] = &Conversation{
		Messages:   []Message{},
		LastUpdate: time.Now(),
	}
}

// cleanup periodically removes old conversations
func (m *ConversationManager) cleanup() {
	ticker := time.NewTicker(m.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		for userID, conv := range m.conversations {
			if time.Since(conv.LastUpdate) > m.ttl {
				delete(m.conversations, userID)
			}
		}
		m.mutex.Unlock()
	}
}
