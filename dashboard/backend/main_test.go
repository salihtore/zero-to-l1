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

func TestValidatorsHandler(t *testing.T) {
	tempDir := t.TempDir()

	record := Deployment{
		ChainName:    "testchain",
		BlockchainID: "23FRkv8jgynuXurD6x7riX2mAjpPK5XAhRzJ5KnEfw6a8U45FE",
		RPCURL:       "http://127.0.0.1:9650/ext/bc/23FR/rpc",
		ChainID:      90349,
		Network:      "fuji",
		DeployedAt:   time.Now().UTC(),
	}
	data, _ := json.Marshal(record)
	_ = os.WriteFile(filepath.Join(tempDir, "testchain.json"), data, 0644)

	s := &server{
		deploymentsDir: tempDir,
		httpClient:     &http.Client{Timeout: 2 * time.Second},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/validators/testchain", nil)
	rec := httptest.NewRecorder()

	s.validatorsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}

	var validators []Validator
	if err := json.NewDecoder(rec.Body).Decode(&validators); err != nil {
		t.Fatalf("failed to decode validators: %v", err)
	}
	if len(validators) == 0 || validators[0].NodeID == "" {
		t.Fatalf("expected validator list, got %+v", validators)
	}
}

func TestTeleporterHandler(t *testing.T) {
	tempDir := t.TempDir()

	record := Deployment{
		ChainName:    "testchain",
		BlockchainID: "23FRkv8jgynuXurD6x7riX2mAjpPK5XAhRzJ5KnEfw6a8U45FE",
		RPCURL:       "http://127.0.0.1:9650/ext/bc/23FR/rpc",
		ChainID:      90349,
		Network:      "fuji",
		DeployedAt:   time.Now().UTC(),
	}
	data, _ := json.Marshal(record)
	_ = os.WriteFile(filepath.Join(tempDir, "testchain.json"), data, 0644)

	s := &server{
		deploymentsDir: tempDir,
		httpClient:     &http.Client{Timeout: 2 * time.Second},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/teleporter/testchain", nil)
	rec := httptest.NewRecorder()

	s.teleporterHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}

	var info TeleporterInfo
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode teleporter info: %v", err)
	}
	if info.MessengerAddress == "" || len(info.Messages) == 0 {
		t.Fatalf("expected valid teleporter info, got %+v", info)
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
