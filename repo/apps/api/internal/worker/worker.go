package worker

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Worker struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	cancel context.CancelFunc
}

func New(pool *pgxpool.Pool, logger *zap.Logger) *Worker {
	return &Worker{
		pool:   pool,
		logger: logger,
	}
}

func (w *Worker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go w.runLoop(ctx, "deferred_notifications", 1*time.Minute, w.processDeferredNotifications)
	go w.runLoop(ctx, "stale_download_tokens", 5*time.Minute, w.cleanupStaleDownloadTokens)
	go w.runLoop(ctx, "risk_recomputation", 10*time.Minute, w.recomputeRiskScores)
	go w.runLoop(ctx, "idempotency_cleanup", 15*time.Minute, w.cleanupExpiredIdempotencyKeys)

	w.logger.Info("background worker started")
}

func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.logger.Info("background worker stopped")
}

func (w *Worker) runLoop(ctx context.Context, name string, interval time.Duration, fn func(context.Context) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := fn(ctx); err != nil {
				w.logger.Error("worker job failed",
					zap.String("job", name),
					zap.Error(err),
				)
			}
		}
	}
}

func (w *Worker) processDeferredNotifications(ctx context.Context) error {
	tag, err := w.pool.Exec(ctx,
		`UPDATE notification_recipients
		 SET status = 'delivered', delivered_at = NOW()
		 WHERE status = 'deferred'
		   AND deferred_until IS NOT NULL
		   AND deferred_until <= NOW()`)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		w.logger.Info("delivered deferred notifications",
			zap.Int64("count", tag.RowsAffected()),
		)
	}
	return nil
}

func (w *Worker) cleanupStaleDownloadTokens(ctx context.Context) error {
	tag, err := w.pool.Exec(ctx,
		`DELETE FROM download_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		w.logger.Info("cleaned up expired download tokens",
			zap.Int64("count", tag.RowsAffected()),
		)
	}
	return nil
}

func (w *Worker) recomputeRiskScores(ctx context.Context) error {
	rows, err := w.pool.Query(ctx,
		`SELECT DISTINCT user_id FROM risk_events
		 WHERE created_at > NOW() - INTERVAL '1 hour'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}

		var score float64
		err := w.pool.QueryRow(ctx,
			`SELECT COALESCE(SUM(
				CASE severity
					WHEN 'critical' THEN 10
					WHEN 'high' THEN 5
					WHEN 'medium' THEN 2
					WHEN 'low' THEN 1
					ELSE 1
				END
			), 0) FROM risk_events WHERE user_id = $1 AND created_at > NOW() - INTERVAL '30 days'`,
			userID,
		).Scan(&score)
		if err != nil {
			w.logger.Error("risk score computation failed", zap.String("user_id", userID), zap.Error(err))
			continue
		}

		_, err = w.pool.Exec(ctx,
			`INSERT INTO risk_scores (id, user_id, score, factors_json, computed_at)
			 VALUES (gen_random_uuid(), $1, $2, '{}', NOW())`,
			userID, score,
		)
		if err != nil {
			w.logger.Error("risk score insert failed", zap.String("user_id", userID), zap.Error(err))
			continue
		}
		count++
	}

	if count > 0 {
		w.logger.Info("recomputed risk scores", zap.Int("count", count))
	}
	return rows.Err()
}

func (w *Worker) cleanupExpiredIdempotencyKeys(ctx context.Context) error {
	tag, err := w.pool.Exec(ctx,
		`DELETE FROM idempotency_keys WHERE expires_at IS NOT NULL AND expires_at < NOW()`)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		w.logger.Info("cleaned up expired idempotency keys",
			zap.Int64("count", tag.RowsAffected()),
		)
	}
	return nil
}
