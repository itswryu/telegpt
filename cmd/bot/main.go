package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/itswryu/telegpt/pkg/config"
	"github.com/itswryu/telegpt/pkg/logger"
	"github.com/itswryu/telegpt/pkg/openai"
	"github.com/itswryu/telegpt/pkg/telegram"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	if err := logger.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Log startup
	logger.Info("TeleGPT starting up...")
	logger.Info("Configuration loaded successfully")

	// Create OpenAI client
	openaiClient := openai.NewClient(cfg)
	logger.Info("OpenAI client initialized")

	// Create Telegram bot
	bot, err := telegram.NewBot(cfg, openaiClient)
	if err != nil {
		logger.Fatal("Failed to create Telegram bot: %v", err)
	}
	logger.Info("Telegram bot created successfully")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start bot in a goroutine
	go func() {
		logger.Info("Starting Telegram bot...")
		if err := bot.Start(); err != nil {
			logger.Fatal("Bot error: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigChan
	logger.Info("Received signal: %v, shutting down...", sig)

	// Stop the bot gracefully
	bot.Stop()
	logger.Info("Shutdown complete")
}
