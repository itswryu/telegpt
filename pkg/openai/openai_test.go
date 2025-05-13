package openai

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/swryu/telegpt/pkg/config"
)

func TestClientCreation(t *testing.T) {
	cfg := &config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4.1-nano",
		},
	}

	client := NewClient(cfg)

	if client.apiKey != "test-key" {
		t.Errorf("NewClient() apiKey = %v, expected %v", client.apiKey, "test-key")
	}

	if client.model != "gpt-4.1-nano" {
		t.Errorf("NewClient() model = %v, expected %v", client.model, "gpt-4.1-nano")
	}

	if client.conversations == nil {
		t.Error("NewClient() conversations map was not initialized")
	}
}

func TestResetConversation(t *testing.T) {
	cfg := &config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4.1-nano",
		},
	}

	client := NewClient(cfg)
	userID := int64(123456789)

	// Add a message to the conversation
	client.mutex.Lock()
	client.conversations[userID] = &Conversation{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
		LastUpdate: time.Now(),
	}
	client.mutex.Unlock()

	// Check that conversation exists
	client.mutex.RLock()
	conv, exists := client.conversations[userID]
	client.mutex.RUnlock()
	if !exists {
		t.Fatal("Conversation should exist before reset")
	}
	if len(conv.Messages) != 2 {
		t.Errorf("Conversation messages count = %v, expected %v", len(conv.Messages), 2)
	}

	// Reset conversation
	client.ResetConversation(userID)

	// Check that conversation was reset
	client.mutex.RLock()
	conv, exists = client.conversations[userID]
	client.mutex.RUnlock()
	if !exists {
		t.Fatal("Conversation should still exist after reset")
	}
	if len(conv.Messages) != 0 {
		t.Errorf("Conversation messages count after reset = %v, expected %v", len(conv.Messages), 0)
	}
}

func TestGenerateResponseWithMockAPI(t *testing.T) {
	// Define constants for expected values
	const (
		testAPIKey     = "test-key"
		testModel      = "gpt-4.1-nano"
		testResponse   = "This is a test response"
		testUserID     = int64(123456)
		testUserPrompt = "Hello, how are you?"
	)

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request headers and method
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected 'application/json' Content-Type, got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer "+testAPIKey {
			t.Errorf("Expected 'Bearer %s' Authorization, got '%s'", testAPIKey, r.Header.Get("Authorization"))
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "mock-id",
			"object": "chat.completion",
			"created": 1234567890,
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "` + testResponse + `"
					},
					"finish_reason": "stop"
				}
			]
		}`))
	}))
	defer server.Close()

	// Create a client that uses the mock server
	cfg := &config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey: testAPIKey,
			Model:  testModel,
		},
	}

	client := NewClient(cfg)
	// Override the base URL to use the mock server
	client.SetBaseURL(server.URL)

	// Test the GenerateResponse method
	response, err := client.GenerateResponse(testUserID, testUserPrompt)
	
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}
	
	if response != testResponse {
		t.Errorf("GenerateResponse() = %v, expected %v", response, testResponse)
	}
}
