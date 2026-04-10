package middleware

import (
	"net/http"
	"strings"

	"travel-platform/apps/api/internal/auth"
	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

func JWTAuth(authService *auth.Service, publicPaths []string) echo.MiddlewareFunc {
	publicSet := make(map[string]bool, len(publicPaths))
	for _, p := range publicPaths {
		publicSet[p] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			if publicSet[path] {
				return next(c)
			}

			for p := range publicSet {
				if strings.HasSuffix(p, "*") && strings.HasPrefix(path, strings.TrimSuffix(p, "*")) {
					return next(c)
				}
			}

			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
			}

			token := parts[1]
			claims, err := authService.ValidateToken(token)
			if err != nil {
				return common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			}

			c.Set(common.ContextKeyUserID, claims.UserID)
			c.Set(common.ContextKeyEmail, claims.Email)
			c.Set(common.ContextKeyRoles, claims.Roles)

			return next(c)
		}
	}
}
