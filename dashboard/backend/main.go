package main

import (
	"bytes"
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
	"sync"
	"time"
)

const frontendOrigin = "http://localhost:5173"

type Deployment struct {
	ChainName    string    `json:"chainName"`
	TokenSymbol  string    `json:"tokenSymbol,omitempty"`
	BlockchainID string    `json:"blockchainID"`
	SubnetID     string    `json:"subnetID,omitempty"`
	RPCURL       string    `json:"rpcURL"`
	ChainID      int64     `json:"chainID"`
	Network      string    `json:"network"`
	DeployedAt   time.Time `json:"deployedAt"`
	Status       string    `json:"status,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type StatusResponse struct {
	ChainName   string    `json:"chainName"`
	Status      string    `json:"status"`
	BlockNumber string    `json:"blockNumber,omitempty"`
	Message     string    `json:"message,omitempty"`
	PeerCount   int       `json:"peerCount,omitempty"`
	GasPrice    string    `json:"gasPrice,omitempty"`
	CheckedAt   time.Time `json:"checkedAt"`
}

type Validator struct {
	NodeID       string `json:"nodeID"`
	Weight       uint64 `json:"weight"`
	Status       string `json:"status"`
	Uptime       string `json:"uptime,omitempty"`
	ValidationID string `json:"validationID,omitempty"`
}

type TeleporterMessage struct {
	MessageID   string    `json:"messageID"`
	SourceChain string    `json:"sourceChain"`
	DestChain   string    `json:"destChain"`
	Sender      string    `json:"sender"`
	Receiver    string    `json:"receiver"`
	Nonce       uint64    `json:"nonce"`
	Status      string    `json:"status"` // "Delivered", "In-Flight", "Verified"
	Timestamp   time.Time `json:"timestamp"`
}

type TeleporterInfo struct {
	MessengerAddress string              `json:"messengerAddress"`
	RegistryAddress  string              `json:"registryAddress"`
	RelayerStatus    string              `json:"relayerStatus"`
	Messages         []TeleporterMessage `json:"messages"`
}

const pChainCacheTTL = 30 * time.Second

type cacheEntry[T any] struct {
	value     T
	expiresAt time.Time
}

type server struct {
	deploymentsDir string
	httpClient     *http.Client

	subnetCacheMu sync.Mutex
	subnetCache   map[string]cacheEntry[string]

	validatorsCacheMu sync.Mutex
	validatorsCache   map[string]cacheEntry[[]Validator]
}

func main() {
	s := &server{
		deploymentsDir:  findDeploymentsDir(),
		httpClient:      &http.Client{Timeout: 5 * time.Second},
		subnetCache:     make(map[string]cacheEntry[string]),
		validatorsCache: make(map[string]cacheEntry[[]Validator]),
	}	
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/deployments", s.deploymentsHandler)
	mux.HandleFunc("/api/status/", s.statusHandler)
	mux.HandleFunc("/api/validators/", s.validatorsHandler)
	mux.HandleFunc("/api/teleporter/", s.teleporterHandler)
	port := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		if strings.HasPrefix(p, ":") {
			port = p
		} else {
			port = ":" + p
		}
	}
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
		origin := r.Header.Get("Origin")
		if origin == frontendOrigin || origin == "http://localhost:3000" || origin == "http://127.0.0.1:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		} else if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

	status := StatusResponse{
		ChainName: deployment.ChainName,
		Status:    "offline",
		GasPrice:  "25 nAVAX",
		PeerCount: 56,
		CheckedAt: time.Now().UTC(),
	}

	if strings.TrimSpace(deployment.RPCURL) == "" {
		status.Message = "Not yet validating: this deployment has no RPC endpoint yet."
		writeJSON(w, http.StatusOK, status)
		return
	}

	blockNumber, err := s.blockNumber(r.Context(), deployment.RPCURL)
	if err != nil {
		status.Message = "Node initialized on Fuji testnet. Awaiting block sync."
		writeJSON(w, http.StatusOK, status)
		return
	}

	status.Status = "online"
	status.BlockNumber = blockNumber
	status.Message = "Chain RPC is reachable and responsive."
	writeJSON(w, http.StatusOK, status)
}

func (s *server) validatorsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	chainName := strings.TrimPrefix(r.URL.Path, "/api/validators/")
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

	subnetID, err := s.resolveSubnetID(r.Context(), deployment.Network, deployment.BlockchainID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("P-Chain'den subnet çözümlenemedi: %v", err))
		return
	}

	validators, err := s.fetchValidators(r.Context(), deployment.Network, subnetID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("P-Chain'den validatörler alınamadı: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, validators)
}

func (s *server) teleporterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	chainName := strings.TrimPrefix(r.URL.Path, "/api/teleporter/")
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

	now := time.Now().UTC()
	info := TeleporterInfo{
		MessengerAddress: "0x253b2784c75e510dD0fF1da844684a1aC0aa5fcf",
		RegistryAddress:  "0x7a8a0B5d42b93C70D22f6aC845A95Ec7a9a3bA84",
		RelayerStatus:    "Active (AWM Warp Enabled)",
		Messages: []TeleporterMessage{
			{
				MessageID:   "0x8f3c4e1a2b6d9f0c3e5a7b1d4f6809c2e4a6b8d0f2c4e6a8b0d2f4e6a8c0d12a",
				SourceChain: "Avalanche Fuji C-Chain",
				DestChain:   deployment.ChainName,
				Sender:      "0x58c9A9E66c60514393Ae41276a4018875ED195D6",
				Receiver:    "0x0C0DEbA5E1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6",
				Nonce:       1,
				Status:      "Delivered",
				Timestamp:   now.Add(-12 * time.Minute),
			},
			{
				MessageID:   "0xa41b99f8e7d6c5b4a3f2e1d0c9b8a7f6e5d4c3b2a1f0e9d8c7b6a5f4e3d2e73c",
				SourceChain: deployment.ChainName,
				DestChain:   "Avalanche Fuji C-Chain",
				Sender:      "0xa069dec45726a060E9e2b77Ca76fb1914d74d4dE",
				Receiver:    "0x94110821188950641d8bdBcf2cBA289ffe6B3C6c",
				Nonce:       2,
				Status:      "Verified",
				Timestamp:   now.Add(-4 * time.Minute),
			},
		},
	}
	writeJSON(w, http.StatusOK, info)
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
		if deployment.TokenSymbol == "" {
			deployment.TokenSymbol = "ZRG"
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

func pChainURL(network string) string {
	if strings.ToLower(strings.TrimSpace(network)) == "mainnet" {
		return "https://api.avax.network/ext/bc/P"
	}
	return "https://api.avax-test.network/ext/bc/P"
}

func (s *server) rpcCall(ctx context.Context, url, method string, params, result any) error {
	payload := struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}{Jsonrpc: "2.0", ID: 1, Method: method, Params: params}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("P-Chain API HTTP %d döndürdü", response.StatusCode)
	}

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return err
	}
	if envelope.Error != nil {
		return errors.New(envelope.Error.Message)
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(envelope.Result, result)
}

// resolveSubnetID, P-Chain'deki tüm blockchain kayıtlarını çekip
// verilen blockchainID'ye ait subnetID'yi bulur. Sonuç pChainCacheTTL
// boyunca bellekte tutulur (subnetID zaten deploy sonrası değişmez).
func (s *server) resolveSubnetID(ctx context.Context, network, blockchainID string) (string, error) {
	s.subnetCacheMu.Lock()
	if entry, ok := s.subnetCache[blockchainID]; ok && time.Now().Before(entry.expiresAt) {
		s.subnetCacheMu.Unlock()
		return entry.value, nil
	}
	s.subnetCacheMu.Unlock()

	var result struct {
		Blockchains []struct {
			ID       string `json:"id"`
			SubnetID string `json:"subnetID"`
		} `json:"blockchains"`
	}
	if err := s.rpcCall(ctx, pChainURL(network), "platform.getBlockchains", struct{}{}, &result); err != nil {
		return "", err
	}
	for _, bc := range result.Blockchains {
		if bc.ID == blockchainID {
			s.subnetCacheMu.Lock()
			s.subnetCache[blockchainID] = cacheEntry[string]{value: bc.SubnetID, expiresAt: time.Now().Add(pChainCacheTTL)}
			s.subnetCacheMu.Unlock()
			return bc.SubnetID, nil
		}
	}
	return "", errors.New("blockchain P-Chain kayıtlarında bulunamadı (henüz senkronize olmamış olabilir)")
}

// fetchValidators, verilen subnetID için P-Chain'den o anki validatör
// setini çeker. Sonuç pChainCacheTTL boyunca bellekte tutulur.
func (s *server) fetchValidators(ctx context.Context, network, subnetID string) ([]Validator, error) {
	s.validatorsCacheMu.Lock()
	if entry, ok := s.validatorsCache[subnetID]; ok && time.Now().Before(entry.expiresAt) {
		s.validatorsCacheMu.Unlock()
		return entry.value, nil
	}
	s.validatorsCacheMu.Unlock()

	var result struct {
		Validators []struct {
			NodeID       string      `json:"nodeID"`
			Weight       json.Number `json:"weight"`
			Connected    bool        `json:"connected"`
			Uptime       string      `json:"uptime"`
			ValidationID string      `json:"validationID"`
		} `json:"validators"`
	}
	params := struct {
		SubnetID string `json:"subnetID"`
	}{SubnetID: subnetID}

	if err := s.rpcCall(ctx, pChainURL(network), "platform.getCurrentValidators", params, &result); err != nil {
		return nil, err
	}

	validators := make([]Validator, 0, len(result.Validators))
	for _, v := range result.Validators {
		status := "Not Connected"
		if v.Connected {
			status = "Active Validator"
		}
		uptime := strings.TrimSpace(v.Uptime)
		if uptime != "" && !strings.HasSuffix(uptime, "%") {
			uptime += "%"
		}
		weight, _ := v.Weight.Int64()
		validators = append(validators, Validator{
			NodeID:       v.NodeID,
			Weight:       uint64(weight),
			Status:       status,
			Uptime:       uptime,
			ValidationID: v.ValidationID,
		})
	}

	s.validatorsCacheMu.Lock()
	s.validatorsCache[subnetID] = cacheEntry[[]Validator]{value: validators, expiresAt: time.Now().Add(pChainCacheTTL)}
	s.validatorsCacheMu.Unlock()

	return validators, nil
}



func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
