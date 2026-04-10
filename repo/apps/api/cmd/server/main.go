package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"travel-platform/apps/api/internal/app"
	"travel-platform/apps/api/internal/config"
	"travel-platform/apps/api/internal/db"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "Run database migrations and exit without starting the server")
	flag.Parse()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := db.RunMigrations(context.Background(), pool, logger); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	if *migrateOnly {
		logger.Info("migrations complete, exiting (--migrate-only)")
		return
	}

	if err := db.RunSeed(context.Background(), pool, logger); err != nil {
		logger.Fatal("failed to run seed", zap.Error(err))
	}

	application := app.New(pool, logger, cfg)
	e := application.SetupRouter()

	application.StartWorker()

	addr := fmt.Sprintf(":%s", cfg.Port)
	go func() {
		logger.Info("starting server", zap.String("addr", addr))
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	application.StopWorker()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited gracefully")
}

func newLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return cfg.Build()
}
