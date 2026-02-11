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
	"github.com/assaidy/blink/app/utils/email"
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
	mailer := email.NewPapercutMailer()

	authService := services.NewAuthService(db, mailer)
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
	// TODO: Use rate limiting for different purposes

	me.registerApiRoutes()
	me.registerHTMLRoutes()
}

func (me *App) registerApiRoutes() {
	apiHandler := handlers.NewApiHandler(
		me.logger,
		me.authService,
		me.chatService,
		me.profileService,
		me.presenceService,
		me.pubsub,
	)
	withSessionAndCsrfTokens := handlers.WithSessionAndCsrfTokens(me.authService)

	v1 := me.router.Group("/api/v1")
	{
		v1.Post("/auth/register", apiHandler.HandleRegister)
		v1.Post("/auth/otp/request", apiHandler.HandleRequestOtp)
		v1.Post("/auth/otp/verify", apiHandler.HandleVerifyOtp)
		v1.Post("/auth/logout", withSessionAndCsrfTokens, apiHandler.HandleLogout)
		v1.Post("/auth/sessions/:session_id", withSessionAndCsrfTokens, apiHandler.HandleDeleteSession)
		v1.Get("/auth/sessions", withSessionAndCsrfTokens, apiHandler.HandleGetActiveSessions)

		v1.Get("/profiles", withSessionAndCsrfTokens, apiHandler.HandleSearchProfiles)
		v1.Get("/profiles/others/:user_id", withSessionAndCsrfTokens, apiHandler.HandleGetProfile)
		v1.Get("/profiles/me", withSessionAndCsrfTokens, apiHandler.HandleGetMyProfile)
		v1.Put("/profiles", withSessionAndCsrfTokens, apiHandler.HandleUpdateProfile)
		v1.Delete("/profiles", withSessionAndCsrfTokens, apiHandler.HandleDeleteProfile)

		v1.Get("/chats", withSessionAndCsrfTokens, apiHandler.HandleGetChatPartners)
		v1.Delete("/chats/:partner_id", withSessionAndCsrfTokens, apiHandler.HandleDeleteChat)
		v1.Get("/chats/:partner_id", withSessionAndCsrfTokens, apiHandler.HandleGetChatMessages)
		v1.Post("/chats/:partner_id/mark_as_read", withSessionAndCsrfTokens, apiHandler.HandleMarkMessagesAsRead)

		v1.Get("/ws", withSessionAndCsrfTokens, apiHandler.WithWebsocket, websocket.New(
			apiHandler.HandleWebsocket,
			websocket.Config{
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

func (me *App) registerHTMLRoutes() {
	htmlHandler := handlers.NewHtmlHandler(
		me.logger,
		me.authService,
		me.chatService,
		me.profileService,
	)
	withSessionToken := handlers.WithSessionToken(me.authService)
	withSessionAndCsrfTokens := handlers.WithSessionAndCsrfTokens(me.authService)

	me.router.Use(handlers.WithRedirectUnauthorizedToLogin)
	me.router.Use("/public", handlers.WithForbiddenAsInvalidEndpoint, filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "web/public",
	}))

	me.router.Get("/register", htmlHandler.HandleRegisterPage)
	me.router.Post("/register", htmlHandler.HandleRegister)
	me.router.Get("/login", htmlHandler.HandleLoginPage)
	me.router.Post("/login", htmlHandler.HandleLogin)
	me.router.Post("/verify_otp", htmlHandler.HandleVerifyOtp)
	me.router.Get("/", withSessionToken, htmlHandler.HandleChatPage)
	me.router.Get("/profile_modal", withSessionAndCsrfTokens, htmlHandler.HandleProfileModal)
	me.router.Put("/profile", withSessionAndCsrfTokens, htmlHandler.HandleUpdateProfile)
	me.router.Post("/logout", withSessionAndCsrfTokens, htmlHandler.HandleLogout)
	me.router.Delete("/sessions/:session_id", withSessionAndCsrfTokens, htmlHandler.HandleDeleteSession)
	me.router.Get("/search_modal", withSessionAndCsrfTokens, htmlHandler.HandleSearchModal)
	me.router.Get("/search/users", withSessionAndCsrfTokens, htmlHandler.HandleSearchUsers)
	me.router.Get("/partners", withSessionAndCsrfTokens, htmlHandler.HandleGetChatPartners)
	me.router.Get("/chat/:partner_id", withSessionAndCsrfTokens, htmlHandler.HandleChatContainer)
	me.router.Get("/chat/:partner_id/messages", withSessionAndCsrfTokens, htmlHandler.HandleChatMessages)
}
