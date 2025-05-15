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

// 설정 파일을 임시로 생성하는 헬퍼 함수
func createTempConfigFile(t *testing.T, content []byte) func() {
	origFile := "config.yaml"
	backupFile := "config.yaml.bak"
	configExists := false

	// 기존 설정 파일 백업
	if _, err := os.Stat(origFile); err == nil {
		configExists = true
		if err := os.Rename(origFile, backupFile); err != nil {
			t.Fatalf("Failed to backup config: %v", err)
		}
	}

	// 테스트용 설정 파일 생성
	if err := os.WriteFile(origFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// 설정 파일 정리를 위한 클린업 함수 반환
	return func() {
		os.Remove(origFile)
		if configExists {
			os.Rename(backupFile, origFile)
		}
	}
}

// 환경 변수 치환 테스트
func TestLoadConfigWithEnvVarSubstitution(t *testing.T) {
	// 테스트 상수 정의
	const (
		testEnvToken   = "env-test-token"
		testEnvKey     = "env-test-key"
		testEnvChatIDs = "111222,333444"
	)

	// 원본 환경 변수 저장 및 복원
	origBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	origApiKey := os.Getenv("OPENAI_API_KEY")
	origChatIDs := os.Getenv("ALLOWED_CHAT_IDS")

	defer func() {
		os.Setenv("TELEGRAM_BOT_TOKEN", origBotToken)
		os.Setenv("OPENAI_API_KEY", origApiKey)
		os.Setenv("ALLOWED_CHAT_IDS", origChatIDs)
	}()

	// 테스트용 환경 변수 설정
	os.Setenv("TELEGRAM_BOT_TOKEN", testEnvToken)
	os.Setenv("OPENAI_API_KEY", testEnvKey)
	os.Setenv("ALLOWED_CHAT_IDS", testEnvChatIDs)

	// 환경변수 플레이스홀더를 사용한 설정 파일 생성
	testConfig := []byte(`
telegram:
  bot_token: "${TELEGRAM_BOT_TOKEN}"
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4.1-nano"
auth:
  allowed_chat_ids: "${ALLOWED_CHAT_IDS}"
logging:
  level: "debug"
`)

	cleanup := createTempConfigFile(t, testConfig)
	defer cleanup()

	// 설정 로드 테스트
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
		return
	}

	// 환경변수 값이 올바르게 치환되었는지 확인
	if cfg.Telegram.BotToken != testEnvToken {
		t.Errorf("Expected BotToken %q, got %q", testEnvToken, cfg.Telegram.BotToken)
	}

	if cfg.OpenAI.APIKey != testEnvKey {
		t.Errorf("Expected APIKey %q, got %q", testEnvKey, cfg.OpenAI.APIKey)
	}

	// 채팅 ID 배열 확인 간소화
	expectedIDs := []int64{111222, 333444}
	validateChatIDs(t, expectedIDs, cfg.Auth.AllowedChatIDs)
}

// 채팅 ID 슬라이스 비교 헬퍼 함수
func validateChatIDs(t *testing.T, expected, actual []int64) {
	if len(actual) != len(expected) {
		t.Errorf("Expected %d chat IDs, got %d", len(expected), len(actual))
		return
	}

	for i, id := range expected {
		if actual[i] != id {
			t.Errorf("Expected chat ID %d at position %d, got %d", id, i, actual[i])
		}
	}
}

func TestLoadFewShotConfig(t *testing.T) {
	// 원본 config 파일 백업
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
			if configExists {
				os.Rename(backupFile, origFile)
			}
		}()
	}

	// 퓨샷 설정이 있는 설정 파일 생성
	testConfig := []byte(`
openai:
  api_key: "test-key"
  model: "test-model"
  system_prompt: "당신은 테스트 봇입니다."
  few_shot_enabled: true
  few_shot_examples:
    - user_question: "테스트 질문 1"
      bot_response: "테스트 응답 1"
    - user_question: "테스트 질문 2"
      bot_response: "테스트 응답 2"
telegram:
  bot_token: "test-token"
auth:
  allowed_chat_ids: "123456789"
`)

	if err := os.WriteFile(origFile, testConfig, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// 설정 로드 테스트
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// 퓨샷 설정 확인
	if cfg.OpenAI.SystemPrompt != "당신은 테스트 봇입니다." {
		t.Errorf("Expected system prompt %q, got %q", "당신은 테스트 봇입니다.", cfg.OpenAI.SystemPrompt)
	}

	if !cfg.OpenAI.FewShotEnabled {
		t.Errorf("Expected few_shot_enabled to be true")
	}

	if len(cfg.OpenAI.FewShotExamples) != 2 {
		t.Errorf("Expected 2 few-shot examples, got %d", len(cfg.OpenAI.FewShotExamples))
	} else {
		// 첫 번째 예시 검증
		if cfg.OpenAI.FewShotExamples[0].UserQuestion != "테스트 질문 1" {
			t.Errorf("Expected first example question %q, got %q",
				"테스트 질문 1", cfg.OpenAI.FewShotExamples[0].UserQuestion)
		}
		if cfg.OpenAI.FewShotExamples[0].BotResponse != "테스트 응답 1" {
			t.Errorf("Expected first example response %q, got %q",
				"테스트 응답 1", cfg.OpenAI.FewShotExamples[0].BotResponse)
		}

		// 두 번째 예시 검증
		if cfg.OpenAI.FewShotExamples[1].UserQuestion != "테스트 질문 2" {
			t.Errorf("Expected second example question %q, got %q",
				"테스트 질문 2", cfg.OpenAI.FewShotExamples[1].UserQuestion)
		}
		if cfg.OpenAI.FewShotExamples[1].BotResponse != "테스트 응답 2" {
			t.Errorf("Expected second example response %q, got %q",
				"테스트 응답 2", cfg.OpenAI.FewShotExamples[1].BotResponse)
		}
	}
}

func TestFewShotEnvironmentVariables(t *testing.T) {
	// 원본 config 파일 백업
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
			if configExists {
				os.Rename(backupFile, origFile)
			}
		}()
	}

	// 최소한의 설정 파일 생성
	minimalConfig := []byte(`
openai:
  api_key: "test-key"
  model: "test-model"
telegram:
  bot_token: "test-token"
auth:
  allowed_chat_ids: "123456789"
`)

	if err := os.WriteFile(origFile, minimalConfig, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// 환경 변수 설정
	os.Setenv("OPENAI_SYSTEM_PROMPT", "환경 변수 테스트 프롬프트")
	os.Setenv("OPENAI_FEW_SHOT_ENABLED", "true")
	defer func() {
		os.Unsetenv("OPENAI_SYSTEM_PROMPT")
		os.Unsetenv("OPENAI_FEW_SHOT_ENABLED")
	}()

	// 설정 로드
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}

	// 환경 변수가 우선되는지 검사
	if cfg.OpenAI.SystemPrompt != "환경 변수 테스트 프롬프트" {
		t.Errorf("환경 변수에서 설정된 시스템 프롬프트가 적용되지 않았습니다. 받은 값: %s",
			cfg.OpenAI.SystemPrompt)
	}

	if !cfg.OpenAI.FewShotEnabled {
		t.Error("환경 변수에서 설정된 퓨샷 활성화 설정이 적용되지 않았습니다")
	}
}
