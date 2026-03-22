package notificator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Notificator struct {
	telegramBot *bot.Bot
	chatIDs     []string
}

func filter[T any](list []T, pred func(T) bool) []T {
	var result []T

	for _, element := range list {
		if pred(element) {
			result = append(result, element)
		}
	}

	return result
}

func New() (*Notificator, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("no environment variable TELEGRAM_TOKEN")
	}

	chatIDs := filter(
		strings.Split(os.Getenv("TELEGRAM_CHANNEL_ID"), ";"),
		func(id string) bool { return len(id) != 0 },
	)
	if len(chatIDs) == 0 {
		return nil, fmt.Errorf("no environment variable TELEGRAM_CHANNEL_ID")
	}

	opts := []bot.Option{
		bot.WithDebug(),
	}

	bot, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to bot: %w", err)
	}

	return &Notificator{bot, chatIDs}, nil
}

func (n *Notificator) SendMessage(ctx context.Context, message string) error {
	for _, chatID := range n.chatIDs {
		_, err := n.telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      message,
			ParseMode: models.ParseModeMarkdown,
		})

		if err != nil {
			return fmt.Errorf("could not send message: %w", err)
		}
	}

	return nil
}

func (n *Notificator) SendMessageDeployed(ctx context.Context) error {
	for _, chatID := range n.chatIDs {
		_, err := n.telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:              chatID,
			Text:                "Deployed!",
			DisableNotification: true,
		})

		if err != nil {
			return fmt.Errorf("could not send message: %w", err)
		}
	}

	return nil
}

func (n *Notificator) SendDailyReports(
	ctx context.Context,
	report *analyzer.DailyReport,
) error {
	message, err := report.Markdown("Yesterday")
	if err != nil {
		return fmt.Errorf("could not get markdown message: %w", err)
	}

	err = n.SendMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("could not send markdown of daily report: %w", err)
	}

	return nil
}
