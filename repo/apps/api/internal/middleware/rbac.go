package middleware

import (
	"net/http"

	"travel-platform/apps/api/internal/common"

	"github.com/labstack/echo/v4"
)

func RequireRole(roles ...string) echo.MiddlewareFunc {
	required := make(map[string]bool, len(roles))
	for _, r := range roles {
		required[r] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoles := common.GetRoles(c)
			if len(userRoles) == 0 {
				return common.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
			}

			for _, r := range userRoles {
				if required[r] {
					return next(c)
				}
			}

			return common.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		}
	}
}
