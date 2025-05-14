package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Telegram TelegramConfig `yaml:"telegram"`
	OpenAI   OpenAIConfig   `yaml:"openai"`
	Auth     AuthConfig     `yaml:"auth"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// TelegramConfig holds Telegram-specific configuration
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
}

// OpenAIConfig holds OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	AllowedChatIDs    []int64 `yaml:"allowed_chat_ids,omitempty"`
	AllowedChatIDsStr string  `yaml:"allowed_chat_ids_str,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface to handle both string and array formats
func (a *AuthConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// 1. 기본 방식 (배열)으로 시도
	var arrayConfig struct {
		AllowedChatIDs    []int64 `yaml:"allowed_chat_ids"`
		AllowedChatIDsStr string  `yaml:"allowed_chat_ids_str"`
	}

	// 배열 방식으로 언마샬링 시도
	if err := unmarshal(&arrayConfig); err == nil {
		a.AllowedChatIDs = arrayConfig.AllowedChatIDs
		a.AllowedChatIDsStr = arrayConfig.AllowedChatIDsStr
	}

	// 2. 문자열 방식 시도 (우선순위 높음)
	var stringConfig struct {
		AllowedChatIDs    string `yaml:"allowed_chat_ids"`
		AllowedChatIDsStr string `yaml:"allowed_chat_ids_str"`
	}

	// 문자열 방식으로 언마샬링 시도
	if err := unmarshal(&stringConfig); err == nil && stringConfig.AllowedChatIDs != "" {
		a.AllowedChatIDsStr = stringConfig.AllowedChatIDs
		// 문자열이 있으면 AllowedChatIDs를 초기화 (새로 파싱)
		a.AllowedChatIDs = nil
	} else if stringConfig.AllowedChatIDsStr != "" {
		a.AllowedChatIDsStr = stringConfig.AllowedChatIDsStr
	}

	// 3. 문자열로부터 배열 파싱 (필요한 경우)
	if a.AllowedChatIDsStr != "" && len(a.AllowedChatIDs) == 0 {
		if err := a.ParseAllowedChatIDs(); err != nil {
			return err
		}
	}

	return nil
}

// ParseAllowedChatIDs parses the AllowedChatIDsStr into AllowedChatIDs
func (a *AuthConfig) ParseAllowedChatIDs() error {
	if a.AllowedChatIDsStr == "" {
		return fmt.Errorf("allowed_chat_ids_str is required")
	}

	ids := strings.Split(a.AllowedChatIDsStr, ",")
	a.AllowedChatIDs = make([]int64, 0, len(ids))

	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}

		chatID, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chat ID %q: %w", trimmed, err)
		}
		a.AllowedChatIDs = append(a.AllowedChatIDs, chatID)
	}

	if len(a.AllowedChatIDs) == 0 {
		return fmt.Errorf("no valid chat IDs found in %q", a.AllowedChatIDsStr)
	}

	return nil
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level   string `yaml:"level"`
	File    string `yaml:"file"`
	Console bool   `yaml:"console"`
}

// LoadConfig loads configuration from config file and/or environment variables
func LoadConfig() (*Config, error) {
	// Try to load .env file if it exists
	_ = godotenv.Load()

	// Try to load from config.yaml if it exists
	cfg := &Config{}
	if configFileExists() {
		if err := loadFromFile(cfg); err != nil {
			return nil, fmt.Errorf("error loading config from file: %w", err)
		}
	}

	// Override with environment variables if they exist
	if err := loadFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("error loading config from env: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func configFileExists() bool {
	_, err := os.Stat("config.yaml")
	return err == nil
}

func loadFromFile(cfg *Config) error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	// 설정 파일 내의 환경 변수 플레이스홀더(${ENV_NAME})를 실제 환경 변수 값으로 치환
	content := string(data)
	envVarRegex := regexp.MustCompile(`\${([^}]+)}`)
	
	content = envVarRegex.ReplaceAllStringFunc(content, func(match string) string {
		// ${...} 에서 변수명만 추출
		envName := match[2 : len(match)-1]
		envValue := os.Getenv(envName)
		
		if envValue != "" {
			return envValue
		}
		// 환경 변수가 없을 경우 원래 플레이스홀더 유지
		return match
	})
	
	// 치환된 내용으로 설정 로드
	return yaml.Unmarshal([]byte(content), cfg)
}

func loadFromEnv(cfg *Config) error {
	// Telegram Bot Token
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		cfg.Telegram.BotToken = token
	}

	// OpenAI API Key
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		cfg.OpenAI.APIKey = apiKey
	}

	// OpenAI Model
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		cfg.OpenAI.Model = model
	}

	// Allowed Chat IDs
	if chatIDs := os.Getenv("ALLOWED_CHAT_IDS"); chatIDs != "" {
		cfg.Auth.AllowedChatIDsStr = chatIDs
		if err := cfg.Auth.ParseAllowedChatIDs(); err != nil {
			return fmt.Errorf("failed to parse ALLOWED_CHAT_IDS: %w", err)
		}
	}

	// Logging configuration
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		cfg.Logging.File = logFile
	}

	if logConsole := os.Getenv("LOG_CONSOLE"); logConsole != "" {
		cfg.Logging.Console = logConsole == "true" || logConsole == "1" || logConsole == "yes"
	}

	return nil
}

// AuthConfig 메서드를 사용하므로 별도의 함수는 필요 없음

func validateConfig(cfg *Config) error {
	if cfg.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is required")
	}

	if cfg.OpenAI.APIKey == "" {
		return fmt.Errorf("OpenAI API key is required")
	}

	if cfg.OpenAI.Model == "" {
		// Set default model if not specified
		cfg.OpenAI.Model = "gpt-4.1-nano"
	}

	// Parse allowed chat IDs from string if present
	if cfg.Auth.AllowedChatIDsStr != "" {
		if err := cfg.Auth.ParseAllowedChatIDs(); err != nil {
			return fmt.Errorf("failed to parse allowed chat IDs: %w", err)
		}
	}

	// After parsing, check if we have any allowed chat IDs
	if len(cfg.Auth.AllowedChatIDs) == 0 {
		return fmt.Errorf("at least one allowed chat ID is required")
	}

	// Default logging configuration
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}

	if !cfg.Logging.Console && cfg.Logging.File == "" {
		// Default to console logging if nothing is specified
		cfg.Logging.Console = true
	}

	return nil
}
