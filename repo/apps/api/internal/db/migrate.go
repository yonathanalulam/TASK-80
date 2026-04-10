package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func migrationsDir() string {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	return "infra/sql/migrations"
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir())
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("no migrations directory found, skipping migrations", zap.String("dir", migrationsDir()))
			return nil
		}
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir(), name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		logger.Info("applying migration", zap.String("version", name))

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for migration %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}

		logger.Info("migration applied", zap.String("version", name))
	}

	return nil
}
