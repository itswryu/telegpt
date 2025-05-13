package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/itswryu/telegpt/pkg/config"
	"github.com/itswryu/telegpt/pkg/logger"
	"github.com/itswryu/telegpt/pkg/openai"
)

// Bot represents a Telegram bot
type Bot struct {
	api            *tgbotapi.BotAPI
	openaiClient   *openai.Client
	allowedChatIDs map[int64]bool
}

// NewBot creates a new Telegram bot
func NewBot(cfg *config.Config, openaiClient *openai.Client) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Telegram bot: %w", err)
	}

	// Create a map for faster lookup
	allowedChatIDs := make(map[int64]bool)
	for _, id := range cfg.Auth.AllowedChatIDs {
		allowedChatIDs[id] = true
	}

	return &Bot{
		api:            bot,
		openaiClient:   openaiClient,
		allowedChatIDs: allowedChatIDs,
	}, nil
}

// Start starts the bot and listens for messages
func (b *Bot) Start() error {
	logger.Info("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		// Check if the user is allowed
		if !b.isAllowedUser(chatID) {
			logger.Warn("Unauthorized access attempt from Chat ID: %d", chatID)
			msg := tgbotapi.NewMessage(chatID, "Unauthorized access. You are not allowed to use this bot.")
			_, _ = b.api.Send(msg)
			continue
		}

		// Process the message
		if update.Message.Text != "" {
			switch update.Message.Text {
			case "/start":
				b.handleStartCommand(chatID)
			case "🆕 New Chat":
				b.handleNewChat(chatID)
			case "🔄 Reset Chat":
				b.openaiClient.ResetConversation(chatID)
				msg := tgbotapi.NewMessage(chatID, "Conversation history has been reset.")
				_, _ = b.api.Send(msg)
			default:
				// Handle normal message
				go b.handleMessage(update.Message)
			}
		}
	}

	return nil
}

// Stop gracefully stops the bot
func (b *Bot) Stop() {
	// Stop getting updates
	b.api.StopReceivingUpdates()
	logger.Info("Bot stopped receiving updates")
}

// isAllowedUser checks if a user is allowed to use the bot
func (b *Bot) isAllowedUser(chatID int64) bool {
	return b.allowedChatIDs[chatID]
}

// handleMessage processes a message and generates a response
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userMessage := message.Text

	// Send "typing" action
	typingMsg := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	_, _ = b.api.Send(typingMsg)

	logger.Info("Received message from %d: %s", chatID, userMessage)

	// Generate response using OpenAI
	response, err := b.openaiClient.GenerateResponse(chatID, userMessage)
	if err != nil {
		logger.Error("Error generating response: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Sorry, I encountered an error generating a response. Please try again later.")
		_, _ = b.api.Send(msg)
		return
	}

	// Send response back to user
	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = b.createMainMenu()
	_, err = b.api.Send(msg)

	if err != nil {
		// If markdown parsing fails, try sending without markdown
		logger.Warn("Error sending markdown message: %v. Trying without markdown.", err)
		msg.ParseMode = ""
		_, _ = b.api.Send(msg)
	}
}

// createMainMenu creates the main keyboard menu
func (b *Bot) createMainMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🆕 New Chat"),
			tgbotapi.NewKeyboardButton("🔄 Reset Chat"),
		),
	)
}

// handleStartCommand handles the /start command
func (b *Bot) handleStartCommand(chatID int64) {
	welcomeText := "Welcome to TeleGPT! 🤖\n\n" +
		"I'm here to help you with your questions and tasks.\n\n" +
		"You can:\n" +
		"• Start a new chat with '🆕 New Chat'\n" +
		"• Reset the current chat with '🔄 Reset Chat'\n" +
		"• Just type your message to continue the current conversation"

	msg := tgbotapi.NewMessage(chatID, welcomeText)
	msg.ReplyMarkup = b.createMainMenu()
	_, _ = b.api.Send(msg)
}

// handleNewChat handles starting a new chat
func (b *Bot) handleNewChat(chatID int64) {
	b.openaiClient.ResetConversation(chatID)
	msg := tgbotapi.NewMessage(chatID, "Starting a new chat! 🆕\nWhat would you like to discuss?")
	_, _ = b.api.Send(msg)
}
