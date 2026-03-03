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
	"github.com/assaidy/blink/app/utils/mailer"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/charmbracelet/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/valkey-io/valkey-go"
)

// TODO: Implement delete/pin chats.
// TODO: Use rate limiting for different purposes

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

	me.registerApiRoutes()
	me.registerHTMLRoutes()
}

func (me *App) registerApiRoutes() {
	jsonHandler := handlers.NewJsonHandler(
		me.logger,
		me.authService,
		me.chatService,
		me.profileService,
		me.presenceService,
		me.pubsub,
	)
	withSessionTokenCookie := handlers.WithSessionTokenCookie(me.authService)
	withCsrfTokenHeader := handlers.WithCsrfTokenHeader(me.authService)
	withCsrfTokenQuery := handlers.WithCsrfTokenQuery(me.authService)

	v1 := me.router.Group("/api/v1")
	{
		v1.Post("/auth/register",
			jsonHandler.HandleRegister,
		)
		v1.Post("/auth/otp/request",
			jsonHandler.HandleRequestOtp,
		)
		v1.Post("/auth/otp/verify",
			jsonHandler.HandleVerifyOtp,
		)
		v1.Post("/auth/logout",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleLogout,
		)
		v1.Post("/auth/sessions/:session_id",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleDeleteSession,
		)
		v1.Get("/auth/sessions",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleGetActiveSessions,
		)

		v1.Get("/profiles",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleSearchProfiles,
		)
		v1.Get("/profiles/others/:user_id",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleGetProfile,
		)
		v1.Get("/profiles/me",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleGetMyProfile,
		)
		v1.Put("/profiles",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleUpdateProfile,
		)
		v1.Delete("/profiles",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleDeleteProfile,
		)

		v1.Get("/chats",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleGetChatPartners,
		)
		v1.Delete("/chats/:partner_id",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleDeleteChat,
		)
		v1.Get("/chats/:partner_id/messages",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleGetChatMessages,
		)
		v1.Post("/chats/:partner_id/messages/mark_as_read",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleMarkMessagesAsRead,
		)
		v1.Delete("/chats/messages/:message_id",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleDeleteChatMessage,
		)
		v1.Put("/chats/messages/:message_id",
			withSessionTokenCookie,
			withCsrfTokenHeader,
			jsonHandler.HandleUpdateChatMessage,
		)

		v1.Get("/ws",
			withSessionTokenCookie,
			withCsrfTokenQuery,
			handlers.WithWebsocket,
			websocket.New(jsonHandler.HandleWebsocket, websocket.Config{
				RecoverHandler: websocketResolver,
			}),
		)
	}
}

//go:embed web/public/*
var staticFS embed.FS

func (me *App) registerHTMLRoutes() {
	htmlHandler := handlers.NewHtmlHandler(
		me.logger,
		me.pubsub,
		me.authService,
		me.chatService,
		me.profileService,
		me.presenceService,
	)
	withSessionTokenCookie := handlers.WithSessionTokenCookie(me.authService)
	withCsrfTokenHeader := handlers.WithCsrfTokenHeader(me.authService)
	withCsrfTokenQuery := handlers.WithCsrfTokenQuery(me.authService)

	me.router.Use(handlers.WithRedirectUnauthorizedToLogin)
	me.router.Use("/public",
		handlers.WithForbiddenAsInvalidEndpoint,
		compress.New(compress.Config{Level: compress.LevelBestSpeed}),
		filesystem.New(filesystem.Config{
			Root:       http.FS(staticFS),
			PathPrefix: "web/public",
		}),
	)

	me.router.Get("/",
		withSessionTokenCookie,
		htmlHandler.HandleChatPage,
	)
	me.router.Get("/register",
		htmlHandler.HandleRegisterPage,
	)
	me.router.Get("/login",
		htmlHandler.HandleLoginPage,
	)

	me.router.Post("/register",
		htmlHandler.HandleRegister,
	)
	me.router.Post("/login",
		htmlHandler.HandleLogin,
	)
	me.router.Post("/verify_otp",
		htmlHandler.HandleVerifyOtp,
	)
	me.router.Get("/profile_modal",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleProfileModal,
	)
	me.router.Put("/profile",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleUpdateProfile,
	)
	me.router.Delete("/profile",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleDeleteProfile,
	)
	me.router.Post("/logout",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleLogout,
	)
	me.router.Delete("/sessions/:session_id",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleDeleteSession,
	)
	me.router.Get("/search_modal",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleSearchModal,
	)
	me.router.Get("/search/users",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleSearchUsers,
	)
	me.router.Get("/search/users/select/:partner_id",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleSelectPartnerFromSearch,
	)
	me.router.Get("/partners",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleGetChatPartners,
	)
	me.router.Get("/chat/:partner_id",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleSelectPartnerFromPartnersList,
	)
	me.router.Get("/chat/:partner_id/messages",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleChatMessages,
	)
	me.router.Put("/chat/:partner_id/messages/:message_id",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleUpdateChatMessage,
	)
	me.router.Get("chat/:partner_id/message_input_form",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleGetChatMessageInputForm,
	)
	me.router.Get("/chat/:partner_id/edit_message_input_form/:message_id",
		withSessionTokenCookie,
		withCsrfTokenHeader,
		htmlHandler.HandleGetEditChatMessageInputForm,
	)

	me.router.Get("/ws",
		withSessionTokenCookie,
		withCsrfTokenQuery,
		handlers.WithWebsocket,
		websocket.New(htmlHandler.HandleWebsocket, websocket.Config{
			RecoverHandler: websocketResolver,
		}),
	)
}

func websocketResolver(c *websocket.Conn) {
	// https://docs.gofiber.io/contrib/websocket/#note-with-recover-middleware
	if err := recover(); err != nil {
		fmt.Fprintf(os.Stderr, "panic: %v\n\n%s\n", err, debug.Stack())
	}
}
