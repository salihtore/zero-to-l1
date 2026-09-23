package launcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// ── Precheck unit tests ──────────────────────────────────────────────────────

// MockCLIVersionChecker implements CLIVersionChecker for tests without network or subprocesses.
type MockCLIVersionChecker struct {
	LocalCLIVersion  string
	LocalCLIErr      error
	LocalAGVersion   string
	LocalAGErr       error
	LatestRelease    *GitHubRelease
	LatestReleaseErr error
}

func (m *MockCLIVersionChecker) FetchLatestRelease(_ context.Context) (*GitHubRelease, error) {
	return m.LatestRelease, m.LatestReleaseErr
}

func (m *MockCLIVersionChecker) GetLocalCLIVersion(_ context.Context) (string, error) {
	return m.LocalCLIVersion, m.LocalCLIErr
}

func (m *MockCLIVersionChecker) GetLocalAvalancheGoVersion(_ context.Context) (string, error) {
	return m.LocalAGVersion, m.LocalAGErr
}

// TestCheckAvalancheCLIUpdate_UpToDate verifies that when local == latest, IsUpToDate is true
// and ProtocolMismatchDetected is true when AG version is ≥ v1.15.0.
func TestCheckAvalancheCLIUpdate_UpToDate(t *testing.T) {
	mock := &MockCLIVersionChecker{
		LocalCLIVersion: "v1.9.6",
		LocalAGVersion:  "v1.15.0",
		LatestRelease: &GitHubRelease{
			TagName: "v1.9.6",
			Body:    "minor security fixes",
		},
	}

	result, err := CheckAvalancheCLIUpdate(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsUpToDate {
		t.Errorf("expected IsUpToDate=true, got false")
	}
	if !result.ProtocolMismatchDetected {
		t.Errorf("expected ProtocolMismatchDetected=true when AG=v1.15.0")
	}
}

// TestCheckAvalancheCLIUpdate_UpdateAvailable verifies update detection when local < latest.
func TestCheckAvalancheCLIUpdate_UpdateAvailable(t *testing.T) {
	mock := &MockCLIVersionChecker{
		LocalCLIVersion: "v1.9.5",
		LocalAGVersion:  "v1.14.0",
		LatestRelease: &GitHubRelease{
			TagName: "v1.9.6",
			Body:    "performance improvements",
		},
	}

	result, err := CheckAvalancheCLIUpdate(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsUpToDate {
		t.Errorf("expected IsUpToDate=false, got true")
	}
	if result.LatestVersion != "v1.9.6" {
		t.Errorf("expected LatestVersion='v1.9.6', got '%s'", result.LatestVersion)
	}
	if result.ProtocolMismatchDetected {
		t.Errorf("expected ProtocolMismatchDetected=false when AG=v1.14.0")
	}
}

// TestCheckAvalancheCLIUpdate_BuiltInVMNote verifies BuiltInVMNoted flag when release notes mention it.
func TestCheckAvalancheCLIUpdate_BuiltInVMNote(t *testing.T) {
	mock := &MockCLIVersionChecker{
		LocalCLIVersion: "v1.9.6",
		LocalAGVersion:  "v1.14.0",
		LatestRelease: &GitHubRelease{
			TagName: "v1.10.0",
			Body:    "Subnet-EVM is now a Built-In VM, no separate plugin required.",
		},
	}

	result, err := CheckAvalancheCLIUpdate(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.BuiltInVMNoted {
		t.Errorf("expected BuiltInVMNoted=true when release notes contain 'built-in vm'")
	}
}

// TestCheckAvalancheCLIUpdate_GitHubError verifies graceful handling of GitHub API failure.
func TestCheckAvalancheCLIUpdate_GitHubError(t *testing.T) {
	mock := &MockCLIVersionChecker{
		LocalCLIVersion:  "v1.9.6",
		LocalAGVersion:   "v1.15.0",
		LatestReleaseErr: errors.New("network timeout"),
	}

	result, err := CheckAvalancheCLIUpdate(context.Background(), mock)
	// Should not return an error — GitHub failure is handled gracefully.
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result even on GitHub error")
	}
	// Mismatch should still be detected from local AG version.
	if !result.ProtocolMismatchDetected {
		t.Errorf("expected ProtocolMismatchDetected=true when AG=v1.15.0 even if GitHub is unreachable")
	}
}

// TestCheckRPCProtocolMismatch_Detected verifies mismatch detection for AG ≥ v1.15.0.
func TestCheckRPCProtocolMismatch_Detected(t *testing.T) {
	cases := []struct {
		agVersion string
		want      bool
	}{
		{"v1.15.0", true},
		{"v1.16.0", true},
		{"v2.0.0", true},
		{"v1.14.9", false},
		{"v1.14.0", false},
		{"v1.13.0", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("AG=%s", tc.agVersion), func(t *testing.T) {
			mock := &MockCLIVersionChecker{LocalAGVersion: tc.agVersion}
			got, err := CheckRPCProtocolMismatch(context.Background(), mock)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("AG=%s: expected mismatch=%v, got %v", tc.agVersion, tc.want, got)
			}
		})
	}
}

// TestCheckRPCProtocolMismatch_AGNotFound verifies conservative (mismatch=true) behaviour
// when the local avalanchego binary cannot be found.
func TestCheckRPCProtocolMismatch_AGNotFound(t *testing.T) {
	mock := &MockCLIVersionChecker{
		LocalAGErr: errors.New("avalanchego binary not found in WSL"),
	}

	got, err := CheckRPCProtocolMismatch(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if !got {
		t.Error("expected mismatch=true (conservative) when avalanchego cannot be found")
	}
}

// ── isVersionAtLeast unit tests ──────────────────────────────────────────────

func TestIsVersionAtLeast(t *testing.T) {
	cases := []struct {
		version    string
		minVersion string
		expected   bool
	}{
		{"v1.15.0", "v1.15.0", true},
		{"v1.16.0", "v1.15.0", true},
		{"v2.0.0", "v1.15.0", true},
		{"v1.14.9", "v1.15.0", false},
		{"v1.15.1", "v1.15.0", true},
		{"v1.14.0", "v1.15.0", false},
		{"v0.9.0", "v1.0.0", false},
		{"v1.0.0", "v0.9.0", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%s>=%s", tc.version, tc.minVersion), func(t *testing.T) {
			got := isVersionAtLeast(tc.version, tc.minVersion)
			if got != tc.expected {
				t.Errorf("isVersionAtLeast(%q, %q) = %v, want %v",
					tc.version, tc.minVersion, got, tc.expected)
			}
		})
	}
}

// ── parseVersionFromOutput unit tests ───────────────────────────────────────

func TestParseVersionFromOutput(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"avalanche version v1.9.6", "v1.9.6"},
		{"v1.15.0, commit abc123", "v1.15.0"},
		{"AvalancheGo/1.15.0", "AvalancheGo/1.15.0"}, // no leading v — falls through
		{"  v2.0.0  ", "v2.0.0"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got := parseVersionFromOutput(tc.input)
			if got != tc.expected {
				t.Errorf("parseVersionFromOutput(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
