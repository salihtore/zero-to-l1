package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type MockScannerRunner struct {
	TargetScanned string
	Err           error
}

func (m *MockScannerRunner) Scan(_ context.Context, targetPath string) error {
	m.TargetScanned = targetPath
	return m.Err
}

func TestDefaultRunner_TargetNotExists(t *testing.T) {
	runner := &DefaultRunner{}
	err := runner.Scan(context.Background(), "non_existent_contract.sol")
	if err == nil {
		t.Fatal("expected error when target file does not exist, got nil")
	}
}

func TestDefaultRunner_EmptyTarget(t *testing.T) {
	runner := &DefaultRunner{}
	err := runner.Scan(context.Background(), "")
	if err == nil {
		t.Fatal("expected error when target is empty, got nil")
	}
}

func TestMockScannerRunner(t *testing.T) {
	mock := &MockScannerRunner{}
	err := mock.Scan(context.Background(), "test.sol")
	if err != nil {
		t.Fatalf("unexpected mock error: %v", err)
	}
	if mock.TargetScanned != "test.sol" {
		t.Errorf("expected scanned target 'test.sol', got '%s'", mock.TargetScanned)
	}
}

func TestDefaultRunner_FindScannerScript(t *testing.T) {
	tempDir := t.TempDir()
	dummyScript := filepath.Join(tempDir, "scanner.py")
	if err := os.WriteFile(dummyScript, []byte("#!/usr/bin/env python3\nprint('mock')\n"), 0755); err != nil {
		t.Fatalf("failed to write dummy script: %v", err)
	}

	runner := &DefaultRunner{
		ScannerPath: dummyScript,
	}

	scriptPath, err := runner.findScannerScript()
	if err != nil {
		t.Fatalf("expected to find scanner script, got error: %v", err)
	}
	if scriptPath != dummyScript {
		t.Errorf("expected script path '%s', got '%s'", dummyScript, scriptPath)
	}
}
