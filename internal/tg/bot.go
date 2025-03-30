package tg

import (
	"fmt"
	"gopkg.in/telebot.v4"
	"tele/internal/config"
	"time"
)

type Bot struct {
	*telebot.Bot
	cfg *config.BotConfig
}

func New(cfg config.BotConfig) (*Bot, error) {
	pref := telebot.Settings{
		Token:  cfg.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("telebot.NewBot: %w", err)
	}
	b := &Bot{
		bot,
		&cfg,
	}

	return b, nil
}
