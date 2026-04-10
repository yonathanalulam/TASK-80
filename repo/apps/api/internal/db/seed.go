package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func RunSeed(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) error {
	seedFile := os.Getenv("SEED_FILE")
	if seedFile == "" {
		return nil
	}

	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 'seed')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("check seed status: %w", err)
	}
	if exists {
		logger.Info("seed already applied, skipping")
		return nil
	}

	content, err := os.ReadFile(seedFile)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("seed file not found, skipping", zap.String("file", seedFile))
			return nil
		}
		return fmt.Errorf("read seed file: %w", err)
	}

	logger.Info("applying seed data", zap.String("file", seedFile))

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("execute seed: %w", err)
	}

	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ('seed')"); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("record seed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed: %w", err)
	}

	logger.Info("seed data applied successfully")
	return nil
}
