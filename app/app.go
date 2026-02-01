package app

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
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
	"github.com/gofiber/fiber/v2/middleware/filesystem"
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
	me.router.Use(recovermw.New(recovermw.Config{EnableStackTrace: true}))
	me.router.Use(handlers.WithLogging(me.logger))
	me.router.Use(handlers.WithErrorResolver(me.logger))
	// TODO: use rate limiting for different purposes

	me.registerApiRoutes()
	me.registerHtmlRoutes()
}

func (me *App) registerApiRoutes() {
	authHandler := handlers.NewAuthHandler(me.logger, me.authService)
	profileHandler := handlers.NewProfileHandler(me.profileService)
	chatHandler := handlers.NewChatHandler(me.chatService)
	wsHandler := handlers.NewWebsocketHandler(me.logger, me.presenceService)

	v1 := me.router.Group("/api/v1")
	{
		v1.Post("auth/register", authHandler.HandleApiRegister)
		v1.Post("auth/otp/request", authHandler.HandleApiRequestOtp)
		v1.Post("auth/otp/verify", authHandler.HandleApiVerifyOtp)
		v1.Post("auth/logout", authHandler.WithSessionAndCSRFTokens, authHandler.HandleApiLogout)
		v1.Get("auth/sessions", authHandler.WithSessionAndCSRFTokens, authHandler.HandleApiGetActiveSessions)

		v1.Get("/profiles", authHandler.WithSessionAndCSRFTokens, profileHandler.HandleApiSearchProfiles)
		v1.Get("/profiles/others/:user_id", authHandler.WithSessionAndCSRFTokens, profileHandler.HandleApiGetProfile)
		v1.Get("/profiles/me", authHandler.WithSessionAndCSRFTokens, profileHandler.HandleApiGetMyProfile)
		v1.Put("/profiles", authHandler.WithSessionAndCSRFTokens, profileHandler.HandleApiUpdateProfile)
		v1.Delete("/profiles", authHandler.WithSessionAndCSRFTokens, profileHandler.HandleApiDeleteProfile)

		v1.Get("/chats", authHandler.WithSessionAndCSRFTokens, chatHandler.HandleApiGetChatPartners)
		v1.Delete("/chats/:partner_id", authHandler.WithSessionAndCSRFTokens, chatHandler.HandleApiDeleteChat)
		v1.Get("/chats/:partner_id", authHandler.WithSessionAndCSRFTokens, chatHandler.HandleApiGetChatMessages)
		v1.Post("/chats/:partner_id/mark_as_read", authHandler.WithSessionAndCSRFTokens, chatHandler.HandleApiMarkMessagesAsRead)

		v1.Get("/ws", authHandler.WithSessionAndCSRFTokens, wsHandler.WithWebsocket, websocket.New(
			wsHandler.HandleApiWebsocket,
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
}

//go:embed web/public/*
var staticFS embed.FS

func (me *App) registerHtmlRoutes() {
	authHandler := handlers.NewAuthHandler(me.logger, me.authService)
	chatHandler := handlers.NewChatHandler(me.chatService)

	me.router.Use("/public", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "web/public",
	}))
	me.router.Use(handlers.WithRedirectToLogin)

	me.router.Get("/register", authHandler.HandleRegisterPage)
	me.router.Post("/register", authHandler.HandleRegister)
	me.router.Get("/login", authHandler.HandleLoginPage)
	me.router.Post("/login", authHandler.HandleLogin)
	me.router.Post("/verify_otp", authHandler.HandleVerifyOtp)
	me.router.Get("/", authHandler.WithSessionToken, chatHandler.HandleChatPage)
}
