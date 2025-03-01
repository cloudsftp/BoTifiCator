package notificator

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
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

func (n *Notificator) SendMessage(ctx context.Context, message string) error {
	_, err := n.telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: n.chatID,
		Text:   message,
	})

	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}

func (n *Notificator) SendMessageDeployed(ctx context.Context) error {
	_, err := n.telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: n.chatID,
		Text:   "Deployed!",
	})

	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}
