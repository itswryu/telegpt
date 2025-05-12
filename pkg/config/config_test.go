package config

import (
	"os"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
	}{
		{
			name: "Valid config",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "test-token",
				},
				OpenAI: OpenAIConfig{
					APIKey: "test-key",
					Model:  "gpt-4.1-nano",
				},
				Auth: AuthConfig{
					AllowedChatIDs: []int64{123456789},
				},
			},
			expectError: false,
		},
		{
			name: "Missing telegram token",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "",
				},
				OpenAI: OpenAIConfig{
					APIKey: "test-key",
					Model:  "gpt-4.1-nano",
				},
				Auth: AuthConfig{
					AllowedChatIDs: []int64{123456789},
				},
			},
			expectError: true,
		},
		{
			name: "Missing OpenAI API key",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "test-token",
				},
				OpenAI: OpenAIConfig{
					APIKey: "",
					Model:  "gpt-4.1-nano",
				},
				Auth: AuthConfig{
					AllowedChatIDs: []int64{123456789},
				},
			},
			expectError: true,
		},
		{
			name: "No allowed chat IDs",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "test-token",
				},
				OpenAI: OpenAIConfig{
					APIKey: "test-key",
					Model:  "gpt-4.1-nano",
				},
				Auth: AuthConfig{
					AllowedChatIDs: []int64{},
				},
			},
			expectError: true,
		},
		{
			name: "Default model provided",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: "test-token",
				},
				OpenAI: OpenAIConfig{
					APIKey: "test-key",
					Model:  "",
				},
				Auth: AuthConfig{
					AllowedChatIDs: []int64{123456789},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if (err != nil) != tt.expectError {
				t.Errorf("validateConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestParseAllowedChatIDs(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedIDs  []int64
		expectedSize int
	}{
		{
			name:         "Valid IDs",
			input:        "123,456,789",
			expectedIDs:  []int64{123, 456, 789},
			expectedSize: 3,
		},
		{
			name:         "IDs with spaces",
			input:        "123, 456, 789",
			expectedIDs:  []int64{123, 456, 789},
			expectedSize: 3,
		},
		{
			name:         "Empty string",
			input:        "",
			expectedIDs:  []int64{},
			expectedSize: 0,
		},
		{
			name:         "Invalid ID",
			input:        "123,abc,789",
			expectedIDs:  []int64{123, 789},
			expectedSize: 2,
		},
		{
			name:         "Empty elements",
			input:        "123,,789",
			expectedIDs:  []int64{123, 789},
			expectedSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAllowedChatIDs(tt.input)
			if len(result) != tt.expectedSize {
				t.Errorf("parseAllowedChatIDs() size = %v, expected %v", len(result), tt.expectedSize)
			}
			for i, expected := range tt.expectedIDs {
				if i >= len(result) || result[i] != expected {
					t.Errorf("parseAllowedChatIDs()[%d] = %v, expected %v", i, result[i], expected)
				}
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original environment and restore after test
	originalTelegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	originalOpenAIKey := os.Getenv("OPENAI_API_KEY")
	originalOpenAIModel := os.Getenv("OPENAI_MODEL")
	originalAllowedChatIDs := os.Getenv("ALLOWED_CHAT_IDS")
	defer func() {
		os.Setenv("TELEGRAM_BOT_TOKEN", originalTelegramToken)
		os.Setenv("OPENAI_API_KEY", originalOpenAIKey)
		os.Setenv("OPENAI_MODEL", originalOpenAIModel)
		os.Setenv("ALLOWED_CHAT_IDS", originalAllowedChatIDs)
	}()

	// Set test environment values
	os.Setenv("TELEGRAM_BOT_TOKEN", "env-token")
	os.Setenv("OPENAI_API_KEY", "env-api-key")
	os.Setenv("OPENAI_MODEL", "env-model")
	os.Setenv("ALLOWED_CHAT_IDS", "111,222,333")

	// Test loading from env
	cfg := &Config{}
	err := loadFromEnv(cfg)
	if err != nil {
		t.Errorf("loadFromEnv() error = %v", err)
	}

	// Check if values were loaded correctly
	if cfg.Telegram.BotToken != "env-token" {
		t.Errorf("loadFromEnv() botToken = %v, expected %v", cfg.Telegram.BotToken, "env-token")
	}
	if cfg.OpenAI.APIKey != "env-api-key" {
		t.Errorf("loadFromEnv() apiKey = %v, expected %v", cfg.OpenAI.APIKey, "env-api-key")
	}
	if cfg.OpenAI.Model != "env-model" {
		t.Errorf("loadFromEnv() model = %v, expected %v", cfg.OpenAI.Model, "env-model")
	}
	if len(cfg.Auth.AllowedChatIDs) != 3 {
		t.Errorf("loadFromEnv() allowedChatIDs size = %v, expected %v", len(cfg.Auth.AllowedChatIDs), 3)
	}
}
