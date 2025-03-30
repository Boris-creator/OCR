package app

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
	"log"
	"tele/internal/api/media"
	"tele/internal/config"
	"tele/internal/mistral"
	"tele/internal/s3"
	"tele/internal/tg"
)

type App struct {
	cfg *config.Config
	bot *tg.Bot
	mc  *mistral.Client

	s3 *s3.Storage

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

	mistralClient := mistral.New(cfg.Mistral)
	app.mc = &mistralClient

	minioClient, err := s3.NewClient(app.cfg.S3, false)
	if err != nil {
		return nil, fmt.Errorf("minio.NewClient: %w", err)
	}
	app.s3 = s3.New(*minioClient, app.cfg.S3)

	apiMediaHandler := media.New(*app.mc, *app.s3)
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
