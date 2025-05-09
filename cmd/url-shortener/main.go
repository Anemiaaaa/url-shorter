package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"os"
	"path/filepath"
	"url-shortener/internal/config"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	endDev   = "dec"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Starting URL shortener service", "env", cfg.Env)
	log.Debug("debug message are enabled")

	// Извлекаем директорию из пути
	storageDir := filepath.Dir(cfg.StoragePath)
	fmt.Println(storageDir)

	// Проверяем и создаем директорию, если она отсутствует
	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		log.Error("Failed to create storage directory", sl.Err(err))
		os.Exit(1)
	}

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("Failed to create storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	_ = storage
	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case "local":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
