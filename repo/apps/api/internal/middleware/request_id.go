package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	HeaderXRequestID = "X-Request-ID"
	ContextKeyRequestID = "request_id"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(HeaderXRequestID)
			if reqID == "" {
				reqID = uuid.New().String()
			}

			c.Set(ContextKeyRequestID, reqID)
			c.Response().Header().Set(HeaderXRequestID, reqID)

			return next(c)
		}
	}
}
