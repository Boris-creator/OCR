package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
	"log"
	"log/slog"
	"os"
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
	db *pgxpool.Pool

	mediaHandler *media.Handler

	logger *slog.Logger
}

func New(cfg *config.Config) (*App, error) {
	app := &App{
		cfg: cfg,
	}

	app.logger = slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		),
	)

	dbCfg := app.cfg.DB
	pool, err := pgxpool.New(context.TODO(), fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		dbCfg.User, dbCfg.Password,
		dbCfg.Host, dbCfg.Port,
		dbCfg.Name,
	))
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	app.db = pool

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

	apiMediaHandler := media.New(app.bot.Bot, app.mc, app.s3, app.db, app.logger)
	app.mediaHandler = apiMediaHandler

	return app, nil
}

func (app *App) start() {
	app.bot.Handle(telebot.OnMedia, app.mediaHandler.Handle)
	app.bot.Start()
}

func (app *App) stop() {
	if app.db != nil {
		app.db.Close()
	}
	if app.bot != nil {
		app.bot.Stop()
	}
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
	defer app.stop()

	app.start()

	return nil
}
