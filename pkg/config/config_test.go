package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// 테스트용 상수 정의
const (
	testToken  = "test-token"
	testKey    = "test-key"
	testModel  = "gpt-4.1-nano"
	testChatID = "123456789"
	envToken   = "env-token"
	envKey     = "env-api-key"
	envModel   = "env-model"
	envChatIDs = "111,222,333"
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
					BotToken: testToken,
				},
				OpenAI: OpenAIConfig{
					APIKey: testKey,
					Model:  testModel,
				},
				Auth: AuthConfig{
					AllowedChatIDsStr: testChatID,
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
					APIKey: testKey,
					Model:  testModel,
				},
				Auth: AuthConfig{
					AllowedChatIDsStr: testChatID,
				},
			},
			expectError: true,
		},
		{
			name: "Missing OpenAI API key",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: testToken,
				},
				OpenAI: OpenAIConfig{
					APIKey: "",
					Model:  testModel,
				},
				Auth: AuthConfig{
					AllowedChatIDsStr: testChatID,
				},
			},
			expectError: true,
		},
		{
			name: "No allowed chat IDs",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: testToken,
				},
				OpenAI: OpenAIConfig{
					APIKey: testKey,
					Model:  testModel,
				},
				Auth: AuthConfig{
					AllowedChatIDsStr: "",
				},
			},
			expectError: true,
		},
		{
			name: "Default model provided",
			cfg: &Config{
				Telegram: TelegramConfig{
					BotToken: testToken,
				},
				OpenAI: OpenAIConfig{
					APIKey: testKey,
					Model:  "",
				},
				Auth: AuthConfig{
					AllowedChatIDsStr: testChatID,
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

// 테스트는 TestAuthConfig_ParseAllowedChatIDs로 대체

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
	os.Setenv("TELEGRAM_BOT_TOKEN", envToken)
	os.Setenv("OPENAI_API_KEY", envKey)
	os.Setenv("OPENAI_MODEL", envModel)
	os.Setenv("ALLOWED_CHAT_IDS", envChatIDs)

	// Test loading from env
	cfg := &Config{}
	err := loadFromEnv(cfg)
	if err != nil {
		t.Errorf("loadFromEnv() error = %v", err)
	}

	// Check if values were loaded correctly
	if cfg.Telegram.BotToken != envToken {
		t.Errorf("loadFromEnv() botToken = %v, expected %v", cfg.Telegram.BotToken, envToken)
	}
	if cfg.OpenAI.APIKey != envKey {
		t.Errorf("loadFromEnv() apiKey = %v, expected %v", cfg.OpenAI.APIKey, envKey)
	}
	if cfg.OpenAI.Model != envModel {
		t.Errorf("loadFromEnv() model = %v, expected %v", cfg.OpenAI.Model, envModel)
	}
	if cfg.Auth.AllowedChatIDsStr != envChatIDs {
		t.Errorf("loadFromEnv() allowedChatIDs = %v, expected %v", cfg.Auth.AllowedChatIDsStr, envChatIDs)
	}
	if len(cfg.Auth.AllowedChatIDs) != 3 {
		t.Errorf("loadFromEnv() allowedChatIDs size = %v, expected %v", len(cfg.Auth.AllowedChatIDs), 3)
	}
}

func TestAuthConfigParse(t *testing.T) {
	// 유효한 단일 ID 테스트
	t.Run("Valid single ID", func(t *testing.T) {
		cfg := &AuthConfig{AllowedChatIDsStr: "123456789"}
		err := cfg.ParseAllowedChatIDs()

		if err != nil {
			t.Errorf("ParseAllowedChatIDs() unexpected error: %v", err)
		}

		if len(cfg.AllowedChatIDs) != 1 || cfg.AllowedChatIDs[0] != 123456789 {
			t.Errorf("ParseAllowedChatIDs() got = %v, want [123456789]", cfg.AllowedChatIDs)
		}
	})

	// 유효한 다중 ID 테스트
	t.Run("Valid multiple IDs", func(t *testing.T) {
		cfg := &AuthConfig{AllowedChatIDsStr: "123456789,987654321"}
		err := cfg.ParseAllowedChatIDs()

		if err != nil {
			t.Errorf("ParseAllowedChatIDs() unexpected error: %v", err)
		}

		if len(cfg.AllowedChatIDs) != 2 ||
			cfg.AllowedChatIDs[0] != 123456789 ||
			cfg.AllowedChatIDs[1] != 987654321 {
			t.Errorf("ParseAllowedChatIDs() got = %v, want [123456789, 987654321]", cfg.AllowedChatIDs)
		}
	})

	// 빈 문자열 테스트
	t.Run("Empty string", func(t *testing.T) {
		cfg := &AuthConfig{AllowedChatIDsStr: ""}
		err := cfg.ParseAllowedChatIDs()

		if err == nil {
			t.Error("ParseAllowedChatIDs() expected error for empty string")
		}
	})
}

func TestAuthConfigUnmarshalYAML(t *testing.T) {
	// 문자열 형태의 allowed_chat_ids 테스트
	t.Run("String allowed_chat_ids", func(t *testing.T) {
		yamlContent := []byte(`allowed_chat_ids: "123456789,987654321"`)
		var auth AuthConfig

		err := yaml.Unmarshal(yamlContent, &auth)
		if err != nil {
			t.Errorf("UnmarshalYAML() unexpected error: %v", err)
		}

		// 값이 올바르게 파싱되었는지 확인
		if len(auth.AllowedChatIDs) != 2 ||
			auth.AllowedChatIDs[0] != 123456789 ||
			auth.AllowedChatIDs[1] != 987654321 {
			t.Errorf("UnmarshalYAML() got = %v, want [123456789, 987654321]", auth.AllowedChatIDs)
		}
	})

	// 배열 형태의 allowed_chat_ids 테스트 (기존 방식)
	t.Run("Array allowed_chat_ids", func(t *testing.T) {
		yamlContent := []byte(`allowed_chat_ids:
  - 123456789
  - 987654321`)
		var auth AuthConfig

		err := yaml.Unmarshal(yamlContent, &auth)
		if err != nil {
			t.Errorf("UnmarshalYAML() unexpected error: %v", err)
		}

		// 값이 올바르게 파싱되었는지 확인
		if len(auth.AllowedChatIDs) != 2 ||
			auth.AllowedChatIDs[0] != 123456789 ||
			auth.AllowedChatIDs[1] != 987654321 {
			t.Errorf("UnmarshalYAML() got = %v, want [123456789, 987654321]", auth.AllowedChatIDs)
		}
	})
}

func TestLoadConfigWithStringAllowedChatIDs(t *testing.T) {
	// Save original config file
	origFile := "config.yaml"
	backupFile := "config.yaml.bak"
	configExists := false

	if _, err := os.Stat(origFile); err == nil {
		configExists = true
		if err := os.Rename(origFile, backupFile); err != nil {
			t.Fatalf("Failed to backup config: %v", err)
		}
		defer func() {
			os.Remove(origFile)
			os.Rename(backupFile, origFile)
		}()
	}

	// Create test config
	testConfig := []byte(`
telegram:
  bot_token: "test-token"
openai:
  api_key: "test-key"
  model: "test-model"
auth:
  allowed_chat_ids: "123456789,987654321"
logging:
  level: "debug"
`)

	if err := os.WriteFile(origFile, testConfig, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer func() {
		if !configExists {
			os.Remove(origFile)
		}
	}()

	// Load config and test
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// Check string chat IDs were parsed correctly
	if len(cfg.Auth.AllowedChatIDs) != 2 {
		t.Errorf("Expected 2 chat IDs, got %d", len(cfg.Auth.AllowedChatIDs))
	}

	if cfg.Auth.AllowedChatIDs[0] != 123456789 || cfg.Auth.AllowedChatIDs[1] != 987654321 {
		t.Errorf("Wrong chat IDs: got %v", cfg.Auth.AllowedChatIDs)
	}
}
