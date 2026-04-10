package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					stack := string(buf[:n])

					reqID, _ := c.Get(ContextKeyRequestID).(string)

					logger.Error("panic recovered",
						zap.String("request_id", reqID),
						zap.String("error", fmt.Sprintf("%v", r)),
						zap.String("stack", stack),
						zap.String("method", c.Request().Method),
						zap.String("path", c.Request().URL.Path),
					)

					_ = common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
				}
			}()
			return next(c)
		}
	}
}
