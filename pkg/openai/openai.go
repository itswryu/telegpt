package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/swryu/telegpt/pkg/config"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1/chat/completions"
	timeout              = 60 * time.Second
	maxHistory           = 10
	historyTTL           = 30 * time.Minute
)

// Message represents a message in a chat conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Conversation represents a chat session with its history
type Conversation struct {
	Messages   []Message
	LastUpdate time.Time
}

// ChatCompletionRequest represents a request to create a chat completion
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatCompletionResponse represents a response from the OpenAI API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

// Client represents an OpenAI API client
type Client struct {
	apiKey        string
	model         string
	baseURL       string
	client        *http.Client
	conversations map[int64]*Conversation
	mutex         sync.RWMutex
}

// NewClient creates a new OpenAI client
func NewClient(cfg *config.Config) *Client {
	client := &Client{
		apiKey:        cfg.OpenAI.APIKey,
		model:         cfg.OpenAI.Model,
		baseURL:       defaultOpenAIBaseURL,
		client:        &http.Client{Timeout: timeout},
		conversations: make(map[int64]*Conversation),
	}

	// Start a cleanup goroutine
	go client.cleanupOldConversations()

	return client
}

// SetBaseURL allows overriding the API base URL (useful for testing)
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// cleanupOldConversations periodically removes old conversations
func (c *Client) cleanupOldConversations() {
	ticker := time.NewTicker(historyTTL / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		for userID, conv := range c.conversations {
			if time.Since(conv.LastUpdate) > historyTTL {
				delete(c.conversations, userID)
			}
		}
		c.mutex.Unlock()
	}
}

// GenerateResponse generates a response using the OpenAI API
func (c *Client) GenerateResponse(userID int64, userMessage string) (string, error) {
	// Get or create the user's conversation
	c.mutex.Lock()
	conv, exists := c.conversations[userID]
	if !exists || time.Since(conv.LastUpdate) > historyTTL {
		conv = &Conversation{
			Messages:   []Message{},
			LastUpdate: time.Now(),
		}
		c.conversations[userID] = conv
	}

	// Add the user's message to the conversation history
	userMsg := Message{
		Role:    "user",
		Content: userMessage,
	}
	conv.Messages = append(conv.Messages, userMsg)

	// Trim history if it exceeds the maximum
	if len(conv.Messages) > maxHistory {
		conv.Messages = conv.Messages[len(conv.Messages)-maxHistory:]
	}

	// Create a copy of the conversation messages
	messages := make([]Message, len(conv.Messages))
	copy(messages, conv.Messages)
	c.mutex.Unlock()

	// If conversation is empty (fresh start), add a system message
	if len(messages) <= 1 {
		systemMsg := Message{
			Role:    "system",
			Content: "You are a helpful assistant.",
		}
		messages = append([]Message{systemMsg}, messages...)
	}

	reqBody := ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	// Save the assistant's response to the conversation history
	c.mutex.Lock()
	conv = c.conversations[userID] // Refresh in case it was deleted
	if conv != nil {
		conv.Messages = append(conv.Messages, result.Choices[0].Message)
		conv.LastUpdate = time.Now()
	}
	c.mutex.Unlock()

	return result.Choices[0].Message.Content, nil
}

// ResetConversation clears the conversation history for a user
func (c *Client) ResetConversation(userID int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.conversations[userID] = &Conversation{
		Messages:   []Message{},
		LastUpdate: time.Now(),
	}
}

// addMessageToHistory adds a message to the conversation history
// This is a helper method used for testing and internally by GenerateResponse
func (c *Client) addMessageToHistory(userID int64, role, content string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get or create the user's conversation
	conv, exists := c.conversations[userID]
	if !exists {
		conv = &Conversation{
			Messages:   []Message{},
			LastUpdate: time.Now(),
		}
		c.conversations[userID] = conv
	} else {
		// Update the conversation's last update time
		conv.LastUpdate = time.Now()
	}

	// Add the message to the conversation history
	msg := Message{
		Role:    role,
		Content: content,
	}
	conv.Messages = append(conv.Messages, msg)

	// Trim history if it exceeds the maximum
	if len(conv.Messages) > maxHistory {
		conv.Messages = conv.Messages[len(conv.Messages)-maxHistory:]
	}
}
