package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iosifbrudnyi/url-shortner/internal/config"
	"github.com/iosifbrudnyi/url-shortner/internal/http-server/handlers/redirect"
	"github.com/iosifbrudnyi/url-shortner/internal/http-server/handlers/save"
	mvLogger "github.com/iosifbrudnyi/url-shortner/internal/http-server/middleware/logger"
	"github.com/iosifbrudnyi/url-shortner/internal/lib/logger/sl"
	"github.com/iosifbrudnyi/url-shortner/internal/storage/postgres"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))
	log.Info("initializing server", slog.String("address", cfg.HttpServer.Address))
	log.Debug("logger debug mode enabled")

	storage, err := postgres.New(cfg.Db)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mvLogger.New(log))

	router.Post("/", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return logger
}
