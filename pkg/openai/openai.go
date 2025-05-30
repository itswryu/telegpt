package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/itswryu/telegpt/pkg/config"
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
	apiKey          string
	model           string
	baseURL         string
	client          *http.Client
	convManager     *ConversationManager
	systemPrompt    string
	fewShotEnabled  bool
	fewShotExamples []FewShotExample
}

// FewShotExample defines a single example for few-shot prompting
type FewShotExample struct {
	UserQuestion string
	BotResponse  string
}

// NewClient creates a new OpenAI client
func NewClient(cfg *config.Config) *Client {
	client := &Client{
		apiKey:         cfg.OpenAI.APIKey,
		model:          cfg.OpenAI.Model,
		baseURL:        defaultOpenAIBaseURL,
		client:         &http.Client{Timeout: timeout},
		convManager:    NewConversationManager(maxHistory, historyTTL),
		systemPrompt:   cfg.OpenAI.SystemPrompt,
		fewShotEnabled: cfg.OpenAI.FewShotEnabled,
	}

	// 퓨샷 예시 설정
	if len(cfg.OpenAI.FewShotExamples) > 0 {
		for _, example := range cfg.OpenAI.FewShotExamples {
			client.fewShotExamples = append(client.fewShotExamples, FewShotExample{
				UserQuestion: example.UserQuestion,
				BotResponse:  example.BotResponse,
			})
		}
	}

	return client
}

// SetBaseURL allows overriding the API base URL (useful for testing)
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// cleanupOldConversations is no longer needed as ConversationManager handles cleanup

// GenerateResponse generates a response using the OpenAI API
func (c *Client) GenerateResponse(userID int64, userMessage string) (string, error) {
	// Get the user's conversation
	conv := c.convManager.GetConversation(userID)

	// Add the user's message to the conversation history
	userMsg := Message{
		Role:    "user",
		Content: userMessage,
	}
	c.convManager.AddMessage(userID, userMsg)

	// Create a copy of the conversation messages
	c.convManager.mutex.RLock()
	messages := make([]Message, len(conv.Messages))
	copy(messages, conv.Messages)
	c.convManager.mutex.RUnlock()

	// 시스템 메시지와 퓨샷 예시를 추가
	messages = c.prepareMessages(messages)

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
	c.convManager.AddMessage(userID, result.Choices[0].Message)

	return result.Choices[0].Message.Content, nil
}

// ResetConversation clears the conversation history for a user
func (c *Client) ResetConversation(userID int64) {
	c.convManager.ResetConversation(userID)
}

// addMessageToHistory adds a message to the conversation history
// This is a helper method used for testing
func (c *Client) addMessageToHistory(userID int64, role, content string) {
	// Create a message
	msg := Message{
		Role:    role,
		Content: content,
	}

	// Add the message using the conversation manager
	c.convManager.AddMessage(userID, msg)
}

// prepareMessages prepares the messages with system prompt and few-shot examples if configured
func (c *Client) prepareMessages(messages []Message) []Message {
	var preparedMessages []Message

	// 시스템 프롬프트 설정
	systemContent := "You are a helpful assistant."
	if c.systemPrompt != "" {
		systemContent = c.systemPrompt
	}

	// 시스템 메시지 추가
	systemMsg := Message{
		Role:    "system",
		Content: systemContent,
	}
	preparedMessages = append(preparedMessages, systemMsg)

	// 퓨샷 예시 추가 (설정되어 있고 활성화된 경우에만)
	if c.fewShotEnabled && len(c.fewShotExamples) > 0 {
		for _, example := range c.fewShotExamples {
			if example.UserQuestion != "" && example.BotResponse != "" {
				preparedMessages = append(preparedMessages, Message{
					Role:    "user",
					Content: example.UserQuestion,
				})
				preparedMessages = append(preparedMessages, Message{
					Role:    "assistant",
					Content: example.BotResponse,
				})
			}
		}
	}

	// 실제 대화 메시지 추가 (시스템 메시지 제외)
	for _, msg := range messages {
		if msg.Role != "system" {
			preparedMessages = append(preparedMessages, msg)
		}
	}

	return preparedMessages
}
