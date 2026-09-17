package launcher

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockRunner implements the Runner interface for unit testing.
type MockRunner struct {
	Mode         string
	ExecPath     string
	CheckErr     error
	CreateErr    error
	DeployErr    error
	DeployOutput string
}

func (m *MockRunner) CheckInstalled(ctx context.Context) (string, string, error) {
	if m.CheckErr != nil {
		return "", "", m.CheckErr
	}
	if m.Mode == "" {
		m.Mode = "native"
	}
	if m.ExecPath == "" {
		m.ExecPath = "/usr/local/bin/avalanche"
	}
	return m.Mode, m.ExecPath, nil
}

func (m *MockRunner) CreateConfig(ctx context.Context, config LaunchConfig) error {
	return m.CreateErr
}

func (m *MockRunner) Deploy(ctx context.Context, chainName string, network string) (string, error) {
	if m.DeployErr != nil {
		return m.DeployOutput, m.DeployErr
	}
	if m.DeployOutput == "" {
		m.DeployOutput = `
Deployment complete!
Blockchain ID: 2qWpG9xZ5hN5q4rP
RPC URL: http://127.0.0.1:9650/ext/bc/2qWpG9xZ5hN5q4rP/rpc
Chain ID: 43113
`
	}
	return m.DeployOutput, nil
}

func TestParseDeployOutput(t *testing.T) {
	sampleOutput := `
Successfully deployed blockchain!
Blockchain ID: 2qWpG9xZ5hN5q4rP
RPC URL: http://127.0.0.1:9650/ext/bc/2qWpG9xZ5hN5q4rP/rpc
Chain ID: 43113
Network: Fuji
`

	bcID, rpcURL, chainID, err := ParseDeployOutput(sampleOutput)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bcID != "2qWpG9xZ5hN5q4rP" {
		t.Errorf("expected Blockchain ID '2qWpG9xZ5hN5q4rP', got '%s'", bcID)
	}

	if rpcURL != "http://127.0.0.1:9650/ext/bc/2qWpG9xZ5hN5q4rP/rpc" {
		t.Errorf("expected RPC URL 'http://127.0.0.1:9650/ext/bc/2qWpG9xZ5hN5q4rP/rpc', got '%s'", rpcURL)
	}

	if chainID != 43113 {
		t.Errorf("expected Chain ID 43113, got %d", chainID)
	}
}

func TestParseDeployOutputInvalid(t *testing.T) {
	invalidOutput := "random error output with no details"
	_, _, _, err := ParseDeployOutput(invalidOutput)
	if err == nil {
		t.Fatal("expected error parsing invalid output, got nil")
	}
}

func TestSaveDeploymentRecord(t *testing.T) {
	tempDir := t.TempDir()

	record := DeploymentRecord{
		ChainName:    "testchain",
		TokenSymbol:  "TST",
		BlockchainID: "11111111111111111111111111111111LpoYY",
		RPCURL:       "http://127.0.0.1:9650/ext/bc/1111/rpc",
		ChainID:      99999,
		Network:      "fuji",
		DeployedAt:   time.Now().UTC(),
		Status:       "deployed",
	}

	err := SaveDeploymentRecord(record, tempDir)
	if err != nil {
		t.Fatalf("failed to save record: %v", err)
	}

	filePath := filepath.Join(tempDir, "testchain.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read saved record: %v", err)
	}

	var readRecord DeploymentRecord
	if err := json.Unmarshal(data, &readRecord); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if readRecord.ChainName != record.ChainName {
		t.Errorf("expected ChainName %s, got %s", record.ChainName, readRecord.ChainName)
	}

	if readRecord.Status != "deployed" {
		t.Errorf("expected Status 'deployed', got '%s'", readRecord.Status)
	}
}

func TestMockLauncherWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	mockRunner := &MockRunner{}

	config := LaunchConfig{
		ChainName:   "mockchain",
		TokenSymbol: "MCK",
		VMType:      "Subnet-EVM",
	}

	ctx := context.Background()

	// 1. Test CreateConfig
	if err := mockRunner.CreateConfig(ctx, config); err != nil {
		t.Fatalf("mock CreateConfig failed: %v", err)
	}

	// 2. Test Deploy
	output, err := mockRunner.Deploy(ctx, config.ChainName, "fuji")
	if err != nil {
		t.Fatalf("mock Deploy failed: %v", err)
	}

	bcID, rpcURL, chainID, err := ParseDeployOutput(output)
	if err != nil {
		t.Fatalf("failed parsing mock deploy output: %v", err)
	}

	record := DeploymentRecord{
		ChainName:    config.ChainName,
		TokenSymbol:  config.TokenSymbol,
		BlockchainID: bcID,
		RPCURL:       rpcURL,
		ChainID:      chainID,
		Network:      "fuji",
		DeployedAt:   time.Now().UTC(),
		Status:       "deployed",
	}

	if err := SaveDeploymentRecord(record, tempDir); err != nil {
		t.Fatalf("failed saving record: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(filepath.Join(tempDir, "mockchain.json")); os.IsNotExist(err) {
		t.Fatalf("expected deployment JSON file does not exist")
	}
}
