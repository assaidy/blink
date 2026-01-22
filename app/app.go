package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/assaidy/blink/app/cache"
	"github.com/assaidy/blink/app/db"
	"github.com/assaidy/blink/app/env"
	"github.com/assaidy/blink/app/handlers"
	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/charmbracelet/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/valkey-io/valkey-go"
)

type App struct {
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
	db := db.GetPool()
	cache := cache.GetClient()
	pubsub := pubsub.NewValkeyPubsub(cache, logger)
	authService := services.NewAuthService(db)
	profileService := services.NewProfileService(db, pubsub)
	presenceService := services.NewPresenceService(db, cache, logger, pubsub)
	chatService := services.NewChatService(db, presenceService, pubsub)

	return &App{
		logger:          logger,
		db:              db,
		cache:           cache,
		pubsub:          pubsub,
		authService:     authService,
		profileService:  profileService,
		presenceService: presenceService,
		chatService:     chatService,
		router: fiber.New(fiber.Config{
			AppName:      "blink",
			ErrorHandler: handlers.ErrorHandler(logger),
			Prefork:      env.Environment == env.EnvDevelopment,
		}),
	}
}

func (me *App) Run() {
	quitCtx, quitCtxCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		me.registerRoutes()
		if err := me.router.Listen(fmt.Sprintf(":%s", env.Port)); err != nil {
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

func (me *App) registerRoutes() {
	authHandler := handlers.NewAuthHandler(me.authService)
	profileHandler := handlers.NewProfileHandler(me.profileService)
	chatHandler := handlers.NewChatHandler(me.chatService)
	wsHandler := handlers.NewWebsocketHandler(me.logger, me.presenceService)

	me.router.Use(recovermw.New(recovermw.Config{EnableStackTrace: true}))
	me.router.Use(handlers.WithLogging(me.logger))
	me.router.Use(handlers.WithErrorResolver(me.logger))
	// TODO: use rate limiting for different purposes

	// when the client app opens, it checks if the client id is stored locally
	// if not exists, it will send the client info, and will recieve the created client id
	me.router.Post("auth/clients", authHandler.HandleCreateClient)
	// will be sent after updating the client app to new version
	me.router.Put("auth/clients/:client_id", authHandler.HandleUpdateClient)
	me.router.Post("auth/register", authHandler.HandleRegister)
	me.router.Post("auth/otp/request", authHandler.HandleRequestOtp)
	me.router.Post("auth/otp/verify", authHandler.HandleVerifyOtp)
	me.router.Post("auth/logout", authHandler.WithSession, authHandler.HandleLogout)
	me.router.Get("auth/sessions", authHandler.WithSession, authHandler.HandleGetActiveSessions)

	me.router.Get("/profiles", authHandler.WithSession, profileHandler.HandleSearchProfiles)
	me.router.Get("/profiles/others/:user_id", authHandler.WithSession, profileHandler.HandleGetProfile)
	me.router.Get("/profiles/me", authHandler.WithSession, profileHandler.HandleGetMyProfile)
	me.router.Put("/profiles", authHandler.WithSession, profileHandler.HandleUpdateProfile)
	me.router.Delete("/profiles", authHandler.WithSession, profileHandler.HandleDeleteProfile)

	me.router.Get("/chats", authHandler.WithSession, chatHandler.HandleGetChatPartners)
	me.router.Delete("/chats/:partner_id", authHandler.WithSession, chatHandler.HandleDeleteChat)
	me.router.Get("/chats/:partner_id", authHandler.WithSession, chatHandler.HandleGetChatMessages)
	me.router.Post("/chats/:partner_id/mark_as_read", authHandler.WithSession, chatHandler.HandleMarkMessagesAsRead)

	me.router.Get("/ws", authHandler.WithSession, wsHandler.WithWebsocket, websocket.New(
		wsHandler.HandleWebsocket,
		websocket.Config{
			// https://docs.gofiber.io/contrib/websocket/#note-with-recover-middleware
			RecoverHandler: func(c *websocket.Conn) {
				if err := recover(); err != nil {
					fmt.Fprintf(os.Stderr, "panic: %v\n\n%s\n", err, debug.Stack())
				}
			},
		},
	))
}
