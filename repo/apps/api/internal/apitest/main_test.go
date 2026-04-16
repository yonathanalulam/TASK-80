//go:build integration

package apitest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"travel-platform/apps/api/internal/app"
	"travel-platform/apps/api/internal/auth"
	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/config"
	"travel-platform/apps/api/internal/db"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Shared test state — initialised once by TestMain.
var (
	testEcho    *echo.Echo
	testApp     *app.App
	authService *auth.Service
)

// Well-known seed user IDs (from infra/sql/seeds/seed.sql).
const (
	AdminUserID      = "c0000000-0000-0000-0000-000000000001"
	OrganizerUserID  = "c0000000-0000-0000-0000-000000000002"
	TravelerUserID   = "c0000000-0000-0000-0000-000000000004"
	SupplierUserID   = "c0000000-0000-0000-0000-000000000008"
	CourierUserID    = "c0000000-0000-0000-0000-000000000011"
	AccountantUserID = "c0000000-0000-0000-0000-000000000013"

	AdminEmail      = "admin@travel.local"
	OrganizerEmail  = "organizer1@travel.local"
	TravelerEmail   = "traveler1@travel.local"
	SupplierEmail   = "supplier1@travel.local"
	CourierEmail    = "courier1@travel.local"
	AccountantEmail = "accountant@travel.local"

	SeedPassword = "password123"

	// Seed itinerary / booking IDs.
	SeedItineraryID = "d0000000-0000-0000-0000-000000000001"
	SeedBookingID   = "d3000000-0000-0000-0000-000000000001"
)

const testJWTSecret = "integration-test-jwt-secret-32!!"

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://travel:travel_test_pass@localhost:5433/travel_platform_test?sslmode=disable"
	}

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		panic("connect to test db: " + err.Error())
	}
	defer pool.Close()

	logger, _ := zap.NewDevelopment()

	if err := db.RunMigrations(ctx, pool, logger); err != nil {
		panic("run migrations: " + err.Error())
	}
	if err := db.RunSeed(ctx, pool, logger); err != nil {
		panic("run seed: " + err.Error())
	}

	_ = os.MkdirAll("/tmp/test-vault", 0o755)

	cfg := &config.Config{
		Port:                "0",
		DatabaseURL:         dbURL,
		JWTSecret:           testJWTSecret,
		MasterEncryptionKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		FileVaultPath:       "/tmp/test-vault",
		DownloadTokenTTL:    15 * time.Minute,
		LogLevel:            "error",
	}

	testApp = app.New(pool, logger, cfg)
	testEcho = testApp.SetupRouter()
	authService = testApp.AuthService

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// tokenFor generates a JWT for the given seed user (via the real auth service).
func tokenFor(t *testing.T, userID, email string, roles []string) string {
	t.Helper()
	tok, err := authService.GenerateTestToken(userID, email, roles)
	if err != nil {
		t.Fatalf("generate test token: %v", err)
	}
	return tok
}

// doRequest dispatches an HTTP request through the real Echo router and
// returns the recorder.
func doRequest(t *testing.T, method, path string, body string, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	testEcho.ServeHTTP(rec, req)
	return rec
}

// parseJSON decodes the standard API envelope.
func parseJSON(t *testing.T, rec *httptest.ResponseRecorder) common.APIResponse {
	t.Helper()
	var resp common.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v (body=%s)", err, rec.Body.String())
	}
	return resp
}

// dataMap extracts Data as map[string]interface{}.
func dataMap(t *testing.T, resp common.APIResponse) map[string]interface{} {
	t.Helper()
	m, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data map, got %T", resp.Data)
	}
	return m
}

// assertStatus checks the HTTP status code.
func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, want, rec.Body.String())
	}
}
