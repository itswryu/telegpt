package config

import (
	"os"
	"reflect"
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
		name        string
		input       string
		expectedIDs []int64
		shouldError bool
	}{
		{
			name:        "Valid single ID",
			input:       "123456",
			expectedIDs: []int64{123456},
			shouldError: false,
		},
		{
			name:        "Valid multiple IDs",
			input:       "123456,789012",
			expectedIDs: []int64{123456, 789012},
			shouldError: false,
		},
		{
			name:        "Empty string",
			input:       "",
			expectedIDs: nil,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAllowedChatIDs(tt.input)
			if (err != nil) != tt.shouldError {
				t.Errorf("parseAllowedChatIDs() error = %v, shouldError %v", err, tt.shouldError)
				return
			}
			if !tt.shouldError && !reflect.DeepEqual(got, tt.expectedIDs) {
				t.Errorf("parseAllowedChatIDs() = %v, want %v", got, tt.expectedIDs)
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

func TestAuthConfig_ParseAllowedChatIDs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int64
		wantErr bool
	}{
		{
			name:    "Valid single ID",
			input:   "123456789",
			want:    []int64{123456789},
			wantErr: false,
		},
		{
			name:    "Valid multiple IDs",
			input:   "123456789,987654321",
			want:    []int64{123456789, 987654321},
			wantErr: false,
		},
		{
			name:    "Valid multiple IDs with spaces",
			input:   "123456789, 987654321",
			want:    []int64{123456789, 987654321},
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid number",
			input:   "123456789,invalid",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AuthConfig{AllowedChatIDsStr: tt.input}
			err := cfg.ParseAllowedChatIDs()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAllowedChatIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(cfg.AllowedChatIDs) != len(tt.want) {
					t.Errorf("ParseAllowedChatIDs() got = %v, want %v", cfg.AllowedChatIDs, tt.want)
					return
				}
				for i, id := range cfg.AllowedChatIDs {
					if id != tt.want[i] {
						t.Errorf("ParseAllowedChatIDs()[%d] = %v, want %v", i, id, tt.want[i])
					}
				}
			}
		})
	}
}
