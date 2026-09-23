package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const frontendOrigin = "http://localhost:5173"

type Deployment struct {
	ChainName    string    `json:"chainName"`
	BlockchainID string    `json:"blockchainID"`
	RPCURL       string    `json:"rpcURL"`
	ChainID      int64     `json:"chainID"`
	Network      string    `json:"network"`
	DeployedAt   time.Time `json:"deployedAt"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type StatusResponse struct {
	ChainName   string    `json:"chainName"`
	Status      string    `json:"status"`
	BlockNumber string    `json:"blockNumber,omitempty"`
	Message     string    `json:"message,omitempty"`
	CheckedAt   time.Time `json:"checkedAt"`
}

type server struct {
	deploymentsDir string
	httpClient     *http.Client
}

func main() {
	s := &server{deploymentsDir: findDeploymentsDir(), httpClient: &http.Client{Timeout: 5 * time.Second}}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/deployments", s.deploymentsHandler)
	mux.HandleFunc("/api/status/", s.statusHandler)
	port := ":8080"
	fmt.Printf("Dashboard backend listening on port %s (deployments: %s)...\n", port, s.deploymentsDir)
	if err := http.ListenAndServe(port, cors(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func findDeploymentsDir() string {
	if configured := os.Getenv("DEPLOYMENTS_DIR"); configured != "" {
		return configured
	}
	for _, candidate := range []string{"../../cli/deployments", "cli/deployments", "deployments"} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return "../../cli/deployments"
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == frontendOrigin {
			w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (s *server) deploymentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	deployments, err := loadDeployments(s.deploymentsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read deployment records")
		return
	}
	writeJSON(w, http.StatusOK, deployments)
}

func (s *server) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	chainName := strings.TrimPrefix(r.URL.Path, "/api/status/")
	if chainName == "" || strings.Contains(chainName, "/") {
		writeError(w, http.StatusBadRequest, "a chain name is required")
		return
	}
	deployment, err := findDeployment(s.deploymentsDir, chainName)
	if errors.Is(err, os.ErrNotExist) {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read deployment record")
		return
	}
	status := StatusResponse{ChainName: deployment.ChainName, Status: "offline", CheckedAt: time.Now().UTC()}
	if strings.TrimSpace(deployment.RPCURL) == "" {
		status.Message = "Not yet validating: this deployment has no RPC endpoint yet."
		writeJSON(w, http.StatusOK, status)
		return
	}
	blockNumber, err := s.blockNumber(r.Context(), deployment.RPCURL)
	if err != nil {
		status.Message = "Not yet validating or RPC endpoint is unreachable."
		writeJSON(w, http.StatusOK, status)
		return
	}
	status.Status, status.BlockNumber, status.Message = "online", blockNumber, "Chain RPC is reachable."
	writeJSON(w, http.StatusOK, status)
}

func loadDeployments(dir string) ([]Deployment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	deployments := make([]Deployment, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var deployment Deployment
		if err := json.Unmarshal(data, &deployment); err != nil {
			return nil, fmt.Errorf("invalid deployment file %s: %w", entry.Name(), err)
		}
		if deployment.ChainName == "" {
			deployment.ChainName = strings.TrimSuffix(entry.Name(), ".json")
		}
		deployments = append(deployments, deployment)
	}
	sort.Slice(deployments, func(i, j int) bool { return deployments[i].ChainName < deployments[j].ChainName })
	return deployments, nil
}

func findDeployment(dir, chainName string) (Deployment, error) {
	deployments, err := loadDeployments(dir)
	if err != nil {
		return Deployment{}, err
	}
	for _, deployment := range deployments {
		if deployment.ChainName == chainName {
			return deployment, nil
		}
	}
	return Deployment{}, os.ErrNotExist
}

func (s *server) blockNumber(ctx context.Context, rpcURL string) (string, error) {
	body := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber","params":[]}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("RPC returned HTTP %d", response.StatusCode)
	}
	var payload struct {
		Result string `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Error != nil || payload.Result == "" {
		return "", errors.New("RPC did not return a block number")
	}
	block, ok := new(big.Int).SetString(strings.TrimPrefix(payload.Result, "0x"), 16)
	if !ok {
		return "", errors.New("invalid block number returned by RPC")
	}
	return block.String(), nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
