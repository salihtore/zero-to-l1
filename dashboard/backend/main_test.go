package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeploymentsHandler(t *testing.T) {
	tempDir := t.TempDir()

	record := Deployment{
		ChainName:    "testchain",
		BlockchainID: "23FRkv8jgynuXurD6x7riX2mAjpPK5XAhRzJ5KnEfw6a8U45FE",
		RPCURL:       "http://127.0.0.1:9650/ext/bc/23FR/rpc",
		ChainID:      90349,
		Network:      "fuji",
		DeployedAt:   time.Now().UTC(),
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "testchain.json"), data, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	s := &server{
		deploymentsDir: tempDir,
		httpClient:     &http.Client{Timeout: 2 * time.Second},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/deployments", nil)
	rec := httptest.NewRecorder()

	s.deploymentsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}

	var items []Deployment
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(items) != 1 || items[0].ChainName != "testchain" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}

	var res HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", res.Status)
	}
}
