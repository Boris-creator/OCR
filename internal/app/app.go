package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"tele/internal/api/about"
	"tele/internal/api/media"
	"tele/internal/api/middleware"
	"tele/internal/config"
	"tele/internal/db/repository"
	"tele/internal/mistral"
	"tele/internal/s3"
	"tele/internal/tg"
	"tele/internal/usecase/metadata"
	"tele/internal/usecase/ocr"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
)

type App struct {
	cfg *config.Config
	bot *tg.Bot
	mc  *mistral.Client

	s3 *s3.Storage
	db *pgxpool.Pool

	documentRepository *repository.DocumentRepository
	chatRepository     *repository.ChatRepository

	mediaService    *ocr.ImageTextRecognizer[*mistral.OCRResponse]
	metadataService *metadata.About

	mediaHandler *media.Handler
	aboutHandler *about.Handler

	mediaValidatorMw *middleware.ImageValidator
	activityMw       *middleware.Activity

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

	err := app.setup()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (app *App) setup() error {
	if err := app.setupDB(); err != nil {
		return fmt.Errorf("app.setupDb: %w", err)
	}

	if err := app.setupBot(); err != nil {
		return fmt.Errorf("app.setupBot: %w", err)
	}

	if err := app.setupMinio(); err != nil {
		return fmt.Errorf("app.setupMinio: %w", err)
	}

	app.setupMistralClient().
		setupRepositories().
		setupServices().
		setupHandlers().
		setupMiddlewares()

	return nil
}

func (app *App) setupDB() error {
	dbCfg := app.cfg.DB
	pool, err := pgxpool.New(context.TODO(), fmt.Sprintf("postgresql://%s:%s@%s/%s",
		dbCfg.User, dbCfg.Password,
		net.JoinHostPort(dbCfg.Host, dbCfg.Port),
		dbCfg.Name,
	))

	if err != nil {
		return fmt.Errorf("pgxpool.New: %w", err)
	}

	app.db = pool

	return nil
}

func (app *App) setupBot() error {
	bot, err := tg.New(app.cfg.Bot)
	if err != nil {
		return fmt.Errorf("tg.New: %w", err)
	}

	app.bot = bot

	return nil
}

func (app *App) setupMinio() error {
	minioClient, err := s3.NewClient(app.cfg.S3, false)
	if err != nil {
		return fmt.Errorf("minio.NewClient: %w", err)
	}

	app.s3 = s3.New(*minioClient, app.cfg.S3)

	return nil
}

func (app *App) setupMistralClient() *App {
	mistralClient := mistral.New(app.cfg.Mistral)
	app.mc = &mistralClient

	return app
}

func (app *App) setupRepositories() *App {
	app.documentRepository = repository.NewDocumentRepository(app.db)
	app.chatRepository = repository.NewChatRepository(app.db)

	return app
}

func (app *App) setupServices() *App {
	app.mediaService = ocr.New(*app.mc, app.s3, app.documentRepository, *app.logger)
	app.metadataService = metadata.New()

	return app
}

func (app *App) setupHandlers() *App {
	app.mediaHandler = media.New(app.bot.Bot, app.logger, app.mediaService)
	app.aboutHandler = about.New(app.bot.Bot, app.metadataService, app.logger)

	return app
}

func (app *App) setupMiddlewares() *App {
	app.mediaValidatorMw = middleware.NewImageValidator()
	app.activityMw = middleware.NewActivityMiddleware(app.chatRepository, app.logger)

	return app
}

func (app *App) start() {
	app.bindHandlers()
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

func (app *App) bindHandlers() {
	app.bot.Use(app.activityMw.RegisterOrRecordRequest)

	app.bot.Handle(telebot.OnMedia, app.mediaHandler.Handle, app.mediaValidatorMw.Validate)
	app.bot.Handle("/about", app.aboutHandler.Handle)
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
