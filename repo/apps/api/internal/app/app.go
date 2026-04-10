package app

import (
	"travel-platform/apps/api/internal/auth"
	"travel-platform/apps/api/internal/config"
	"travel-platform/apps/api/internal/modules/admin"
	"travel-platform/apps/api/internal/modules/bookings"
	"travel-platform/apps/api/internal/modules/contracts"
	"travel-platform/apps/api/internal/modules/files"
	"travel-platform/apps/api/internal/modules/finance"
	"travel-platform/apps/api/internal/modules/itineraries"
	"travel-platform/apps/api/internal/modules/notifications"
	"travel-platform/apps/api/internal/modules/pricing"
	"travel-platform/apps/api/internal/modules/procurement"
	"travel-platform/apps/api/internal/modules/reviews"
	"travel-platform/apps/api/internal/modules/risk"
	"travel-platform/apps/api/internal/modules/users"
	"travel-platform/apps/api/internal/worker"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	DB     *pgxpool.Pool
	Logger *zap.Logger
	Config *config.Config

	AuthService *auth.Service

	AuthHandler          *auth.Handler
	ItineraryHandler     *itineraries.Handler
	BookingHandler       *bookings.Handler
	PricingHandler       *pricing.Handler
	NotificationHandler  *notifications.Handler
	FileHandler          *files.Handler
	FinanceHandler       *finance.Handler
	ReviewHandler        *reviews.Handler
	ContractHandler      *contracts.Handler
	ProcurementHandler   *procurement.Handler
	RiskHandler          *risk.Handler
	AdminHandler         *admin.Handler
	UserHandler          *users.Handler

	Worker *worker.Worker
}

func New(db *pgxpool.Pool, logger *zap.Logger, cfg *config.Config) *App {
	a := &App{
		DB:     db,
		Logger: logger,
		Config: cfg,
	}

	a.AuthService = auth.NewService(db, logger, cfg.JWTSecret)
	a.AuthHandler = auth.NewHandler(a.AuthService)

	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo, logger)
	a.UserHandler = users.NewHandler(userService, logger)

	itinRepo := itineraries.NewRepository(db)
	itinService := itineraries.NewService(itinRepo, logger)
	a.ItineraryHandler = itineraries.NewHandler(itinService, logger)

	pricingRepo := pricing.NewRepository(db)
	a.PricingHandler = pricing.NewHandler(pricingRepo)

	riskRepo := risk.NewRepository(db)
	riskService := risk.NewService(riskRepo, logger)
	a.RiskHandler = risk.NewHandler(riskService, logger)

	bookingRepo := bookings.NewRepository(db)
	bookingService := bookings.NewService(db, bookingRepo, pricingRepo, riskService, logger)
	a.BookingHandler = bookings.NewHandler(bookingService)

	notifRepo := notifications.NewRepository(db, logger)
	notifService := notifications.NewService(notifRepo, logger)
	a.NotificationHandler = notifications.NewHandler(notifService, logger)

	fileRepo := files.NewRepository(db)
	fileService := files.NewFileVaultService(fileRepo, cfg, logger)
	a.FileHandler = files.NewHandler(fileService, logger)

	financeRepo := finance.NewRepository(db)
	financeService := finance.NewService(financeRepo, riskService, logger)
	a.FinanceHandler = finance.NewHandler(financeService, logger)

	reviewRepo := reviews.NewRepository(db)
	reviewService := reviews.NewReviewService(reviewRepo, logger)
	a.ReviewHandler = reviews.NewHandler(reviewService, logger)

	contractRepo := contracts.NewRepository(db)
	contractService := contracts.NewContractService(contractRepo, fileService, logger)
	a.ContractHandler = contracts.NewHandler(contractService, logger)

	procRepo := procurement.NewRepository(db)
	procService := procurement.NewService(procRepo, riskService, logger)
	a.ProcurementHandler = procurement.NewHandler(procService, logger)

	adminRepo := admin.NewRepository(db)
	a.AdminHandler = admin.NewHandler(adminRepo, logger)

	a.Worker = worker.New(db, logger)

	return a
}

func (a *App) StartWorker() {
	a.Worker.Start()
}

func (a *App) StopWorker() {
	a.Worker.Stop()
}
