package app

import (
	"net/http"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/middleware"
	contractsmod "travel-platform/apps/api/internal/modules/contracts"
	filesmod "travel-platform/apps/api/internal/modules/files"
	reviewsmod "travel-platform/apps/api/internal/modules/reviews"

	"github.com/labstack/echo/v4"
)

func (a *App) SetupRouter() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recovery(a.Logger))
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(a.Logger))

	publicPaths := []string{
		"/api/v1/auth/login",
		"/health",
		"/ready",
		"/api/v1/files/download/*",
	}

	e.Use(middleware.JWTAuth(a.AuthService, publicPaths))

	e.Use(middleware.AuditLogger(a.DB, a.Logger))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/ready", func(c echo.Context) error {
		if err := a.DB.Ping(c.Request().Context()); err != nil {
			return common.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "database not ready")
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	v1 := e.Group("/api/v1")

	authGroup := v1.Group("/auth")
	a.AuthHandler.RegisterRoutes(authGroup)

	userGroup := v1.Group("/users")
	a.UserHandler.RegisterRoutes(userGroup)

	itinGroup := v1.Group("/itineraries")
	a.ItineraryHandler.RegisterRoutes(itinGroup)

	couponGroup := v1.Group("/coupons")
	a.PricingHandler.RegisterRoutes(couponGroup)

	bookingGroup := v1.Group("/bookings")
	a.BookingHandler.RegisterRoutes(bookingGroup)

	notifGroup := v1.Group("")
	a.NotificationHandler.RegisterRoutes(notifGroup, userGroup)

	fileGroup := v1.Group("/files")
	filesmod.RegisterRoutes(fileGroup, a.FileHandler)

	finGroup := v1.Group("")
	a.FinanceHandler.RegisterRoutes(finGroup)

	reviewGroup := v1.Group("")
	reviewsmod.RegisterRoutes(reviewGroup, a.ReviewHandler)

	contractGroup := v1.Group("")
	contractsmod.RegisterRoutes(contractGroup, a.ContractHandler)

	procGroup := v1.Group("")
	a.ProcurementHandler.RegisterRoutes(procGroup)

	riskGroup := v1.Group("")
	adminGroup := v1.Group("/admin")
	a.RiskHandler.RegisterRoutes(riskGroup, adminGroup)

	a.AdminHandler.RegisterRoutes(adminGroup)

	a.UserHandler.RegisterAdminRoutes(adminGroup)

	return e
}
