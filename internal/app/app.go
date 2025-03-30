package app

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
	"log"
	"tele/internal/api/media"
	"tele/internal/config"
	"tele/internal/mistral"
	"tele/internal/tg"
)

type App struct {
	cfg *config.Config
	bot *tg.Bot
	mc  *mistral.Client

	mediaHandler *media.Handler
}

func New(cfg *config.Config) (*App, error) {
	app := &App{
		cfg: cfg,
	}
	bot, err := tg.New(cfg.Bot)
	if err != nil {
		return nil, fmt.Errorf("tg.New: %w", err)
	}
	app.bot = bot

	mc := mistral.New(cfg.Mistral)
	app.mc = &mc

	apiMediaHandler := media.New(*app.mc)
	app.mediaHandler = apiMediaHandler

	return app, nil
}

func (app *App) start() {
	app.bot.Handle(telebot.OnMedia, app.mediaHandler.Handle)
	app.bot.Start()
}

func Start() error {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	app, err := New(cfg)
	if err != nil {
		return err
	}

	log.Println("starting bot")

	app.start()

	return nil
}
