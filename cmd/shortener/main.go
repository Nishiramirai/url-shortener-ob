package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener-ob/internal/config"
	"url-shortener-ob/internal/handler"
	"url-shortener-ob/internal/handler/middleware"
	"url-shortener-ob/internal/repository/memory"
	"url-shortener-ob/internal/repository/postgres"
	"url-shortener-ob/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)
	slog.SetDefault(logger)

	logger.Info("starting url-shortener", slog.String("env", cfg.Env))
	logger.Debug("config loaded", slog.Any("config", cfg))

	var repo service.Repository
	dsn := cfg.Postgres.ConnectionURL()

	if cfg.StorageType == "postgres" {
		db, err := postgres.NewDB(dsn)
		if err != nil {
			logger.Error("failed to connect to db", slog.Any("err", err))
			os.Exit(1)
		}
		repo = postgres.New(db)
		logger.Info("postgres storage started")
	} else {
		repo = memory.New(cfg.MemoryStorageLimit)
		logger.Info("memory storage started")
	}

	urlService := service.New(repo)

	if cfg.Env == envProd {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(middleware.SlogLogger(logger), gin.Recovery())

	h := handler.New(urlService, logger)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		logger.Info("http server is running", slog.String("address", cfg.HTTPServer.Address))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen and serve error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server gracefully...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("Server exited properly")
}

func setupLogger(env string) *slog.Logger {
	var logHandler slog.Handler

	switch env {
	case envLocal:
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envDev:
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envProd:
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}

	return slog.New(logHandler)
}
