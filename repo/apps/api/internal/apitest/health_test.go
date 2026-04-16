//go:build integration

package apitest

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHealth_Returns200(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/health", "", "")
	assertStatus(t, rec, http.StatusOK)

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestReady_Returns200(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/ready", "", "")
	assertStatus(t, rec, http.StatusOK)

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse ready response: %v", err)
	}
	if body["status"] != "ready" {
		t.Errorf("expected status=ready, got %q", body["status"])
	}
}
