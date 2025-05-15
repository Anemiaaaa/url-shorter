package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/deleteURL"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
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

	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found or failed to load")
	}

	cfg := config.MustLoad()
	var log *slog.Logger = setupLogger(cfg.Env)

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

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HttpServer.User: cfg.HttpServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", deleteURL.New(log, storage))
	})

	router.Get("/url/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("Failed to start server", sl.Err(err))
	}

	log.Error("Server stopped", sl.Err(err))
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
