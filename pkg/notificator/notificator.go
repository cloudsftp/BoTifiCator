package notificator

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Notificator struct {
	telegramBot *bot.Bot
	chatID      string
}

func New() (*Notificator, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("no environment variable TELEGRAM_TOKEN")
	}

	chatID := os.Getenv("TELEGRAM_CHANNEL_ID")
	if chatID == "" {
		return nil, fmt.Errorf("no environment variable TELEGRAM_CHANNEL_ID")
	}

	opts := []bot.Option{
		bot.WithDebug(),
	}

	bot, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to bot: %w", err)
	}

	return &Notificator{bot, chatID}, nil
}

func (n *Notificator) Message(ctx context.Context) error {
	msg, err := n.telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:              n.chatID,
		Text:                "Hello",
		ParseMode:           "",
		Entities:            []models.MessageEntity{},
		LinkPreviewOptions:  &models.LinkPreviewOptions{},
		DisableNotification: false,
		ProtectContent:      false,
		AllowPaidBroadcast:  false,
		MessageEffectID:     "",
		ReplyParameters:     &models.ReplyParameters{},
		ReplyMarkup:         nil,
	})
	_ = msg

	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}
