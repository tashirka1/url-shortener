package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url_shortener"

	auth_handler "url_shortener/internal/auth/handler"
	"url_shortener/internal/core/db"
	"url_shortener/internal/core/health"
	link_handler "url_shortener/internal/link/handler"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	slog.Error("HTTP error", "error", err.Error(), "path", c.Request().URL.Path)

	if he, ok := err.(*echo.HTTPError); ok {
		c.JSON(he.Code, map[string]any{
			"error":   http.StatusText(he.Code),
			"message": he.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, map[string]any{
		"error":   "Internal Server Error",
		"message": err.Error(),
	})
}

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using environment variables")
	}

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		slog.Error("SESSION_KEY is not set")
		os.Exit(1)
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		slog.Error("DB_NAME is not set")
		os.Exit(1)
	}

	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	database, err := db.NewDB(dbName)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// session store
	sessionStore := sessions.NewCookieStore([]byte(sessionKey))

	// echo
	e := echo.New()
	e.HideBanner = true            // Hides the ASCII art banner
	e.HidePort = true              // Hides the "HTTP server started on" message
	e.Logger.SetOutput(io.Discard) // Discards all default engine logs
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	// middleware
	protection := http.NewCrossOriginProtection()
	csrfMiddleware := echo.WrapMiddleware(func(next http.Handler) http.Handler {
		return protection.Handler(next)
	})
	e.Use(csrfMiddleware)
	e.Use(middleware.RequestLogger())
	e.Use(session.Middleware(sessionStore))
	e.Use(middleware.ContextTimeout(10 * time.Second))
	e.StaticFS("/static", echo.MustSubFS(url_shortener.EmbeddedStatic, "static"))
	e.GET("/health", health.Handler(database))

	// routes
	auth_handler.SetupHandlers(e, database, sessionStore)
	link_handler.SetupHandlers(e, database, sessionStore)

	// Start Server
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting on :8000")
		if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
