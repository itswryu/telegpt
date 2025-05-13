package config

import (
	"fmt"
	"os"
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

	return yaml.Unmarshal(data, cfg)
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
		ids, err := parseAllowedChatIDs(chatIDs)
		if err != nil {
			return fmt.Errorf("failed to parse ALLOWED_CHAT_IDS: %w", err)
		}
		cfg.Auth.AllowedChatIDs = ids
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

// parseAllowedChatIDs converts a comma-separated string of chat IDs to []int64
func parseAllowedChatIDs(chatIDsStr string) ([]int64, error) {
	if chatIDsStr == "" {
		return nil, fmt.Errorf("allowed_chat_ids is required but was empty")
	}

	strIDs := strings.Split(chatIDsStr, ",")
	ids := make([]int64, 0, len(strIDs))

	for _, str := range strIDs {
		id, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat ID format: %s", str)
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid chat IDs found in: %s", chatIDsStr)
	}

	return ids, nil
}

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

	if len(cfg.Auth.AllowedChatIDs) == 0 {
		return fmt.Errorf("at least one allowed chat ID is required")
	}

	if err := cfg.Auth.ParseAllowedChatIDs(); err != nil {
		return fmt.Errorf("failed to parse allowed chat IDs: %w", err)
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
