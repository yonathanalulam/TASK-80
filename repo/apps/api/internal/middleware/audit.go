package middleware

import (
	"context"
	"encoding/json"

	"travel-platform/apps/api/internal/common"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func AuditLogger(pool *pgxpool.Pool, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			method := c.Request().Method
			if method == "GET" || method == "HEAD" || method == "OPTIONS" {
				return err
			}

			userID := common.GetUserID(c)
			reqID := common.GetRequestID(c)
			path := c.Request().URL.Path
			status := c.Response().Status

			ip := c.RealIP()
			go func() {
				afterSummary, _ := json.Marshal(map[string]interface{}{
					"status": status,
					"path":   path,
					"method": method,
				})

				_, dbErr := pool.Exec(context.Background(),
					`INSERT INTO audit_logs (id, actor_id, action, entity_type, entity_id, after_summary, request_id, ip_address)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
					uuid.New().String(),
					nullIfEmpty(userID),
					method+" "+path,
					"api_request",
					nil,
					afterSummary,
					nullIfEmpty(reqID),
					ip,
				)
				if dbErr != nil {
					logger.Error("failed to write audit log",
						zap.Error(dbErr),
						zap.String("path", path),
						zap.String("method", method),
					)
				}
			}()

			return err
		}
	}
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
