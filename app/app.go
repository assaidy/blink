package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/assaidy/blink/app/cache"
	"github.com/assaidy/blink/app/config"
	"github.com/assaidy/blink/app/db"
	"github.com/assaidy/blink/app/handlers"
	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils/mailer"
	"github.com/assaidy/blink/app/utils/pubsub"
	valkey_pubsub "github.com/assaidy/blink/app/utils/pubsub/valkey"
	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/valkey-io/valkey-go"
)

type App struct {
	config          *config.Config
	logger          *slog.Logger
	db              *sql.DB
	cache           valkey.Client
	pubsub          pubsub.Pubsub
	authService     *services.AuthService
	profileService  *services.ProfileService
	presenceService *services.PresenceService
	chatService     *services.ChatService
	router          *fiber.App
}

func NewApp() *App {
	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: true}))

	conf := config.Load()

	db := db.GetPostgresConnectionPool(conf.PostgresUrl)
	valkeyClient := cache.GetValkeyClient(conf.ValkeyAddr)
	pubsub := valkey_pubsub.New(valkeyClient)
	mailer := email.NewPapercutMailer(conf.PapercutHost, conf.EmailFrom)

	authService := services.NewAuthService(db, mailer, conf.Secret)
	profileService := services.NewProfileService(db, pubsub)
	presenceService := services.NewPresenceService(db, valkeyClient, logger, pubsub)
	chatService := services.NewChatService(db, presenceService, pubsub)

	return &App{
		config:          conf,
		logger:          logger,
		db:              db,
		cache:           valkeyClient,
		pubsub:          pubsub,
		authService:     authService,
		profileService:  profileService,
		presenceService: presenceService,
		chatService:     chatService,
		router: fiber.New(fiber.Config{
			AppName:      "blink",
			ErrorHandler: handlers.ErrorHandler(logger),
			Prefork:      conf.Environment == config.EnvDevelopment,
		}),
	}
}

func (me *App) Run() {
	quitCtx, quitCtxCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		me.registerRoutes()
		if err := me.router.Listen(fmt.Sprintf(":%s", me.config.Port)); err != nil {
			me.logger.Error("failed to start server", "error", err)
		}
	}()

	<-quitCtx.Done()
	quitCtxCancel()
	me.logger.Info("gracefully shutting down server. press Ctrl-c to force shutdown.")

	if err := me.router.ShutdownWithTimeout(5 * time.Second); err != nil {
		me.logger.Error("failed to shutdown server", "error", err)
	}
}
