package middleware

import (
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var sensitiveHeaders = map[string]bool{
	"authorization": true,
	"cookie":        true,
	"set-cookie":    true,
	"x-api-key":     true,
}

var sensitiveQueryParams = map[string]bool{
	"token":    true,
	"password": true,
	"secret":   true,
	"api_key":  true,
}

func Logger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			reqID, _ := c.Get(ContextKeyRequestID).(string)

			fields := []zap.Field{
				zap.String("request_id", reqID),
				zap.String("method", req.Method),
				zap.String("path", sanitizePath(req.URL.Path)),
				zap.String("query", maskQueryParams(req.URL.RawQuery)),
				zap.Int("status", res.Status),
				zap.Duration("latency", latency),
				zap.String("remote_ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
				zap.Int64("bytes_out", res.Size),
			}

			status := res.Status
			switch {
			case status >= 500:
				logger.Error("request completed", fields...)
			case status >= 400:
				logger.Warn("request completed", fields...)
			default:
				logger.Info("request completed", fields...)
			}

			return nil
		}
	}
}

func sanitizePath(path string) string {
	const prefix = "/files/download/"
	if idx := strings.Index(path, prefix); idx >= 0 {
		tokenStart := idx + len(prefix)
		if tokenStart < len(path) {
			return path[:tokenStart] + "[REDACTED]"
		}
	}
	return path
}

func maskQueryParams(raw string) string {
	if raw == "" {
		return ""
	}

	pairs := strings.Split(raw, "&")
	for i, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 && sensitiveQueryParams[strings.ToLower(parts[0])] {
			pairs[i] = parts[0] + "=***"
		}
	}
	return strings.Join(pairs, "&")
}
