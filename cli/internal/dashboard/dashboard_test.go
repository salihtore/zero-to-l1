package dashboard

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type MockDashboardRunner struct {
	BackendPort  int
	FrontendPort int
	Err          error
}

func (m *MockDashboardRunner) Start(_ context.Context, backendPort int, frontendPort int) error {
	m.BackendPort = backendPort
	m.FrontendPort = frontendPort
	return m.Err
}

func TestMockDashboardRunner(t *testing.T) {
	mock := &MockDashboardRunner{}
	err := mock.Start(context.Background(), 8080, 5173)
	if err != nil {
		t.Fatalf("unexpected mock error: %v", err)
	}
	if mock.BackendPort != 8080 {
		t.Errorf("expected backend port 8080, got %d", mock.BackendPort)
	}
	if mock.FrontendPort != 5173 {
		t.Errorf("expected frontend port 5173, got %d", mock.FrontendPort)
	}
}

func TestDefaultRunner_FindDirectories(t *testing.T) {
	tempDir := t.TempDir()

	backendDir := filepath.Join(tempDir, "backend")
	_ = os.MkdirAll(backendDir, 0755)
	_ = os.WriteFile(filepath.Join(backendDir, "main.go"), []byte("package main\nfunc main(){}"), 0644)

	frontendDir := filepath.Join(tempDir, "frontend")
	_ = os.MkdirAll(frontendDir, 0755)
	_ = os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte("{}"), 0644)

	runner := &DefaultRunner{
		BackendDir:  backendDir,
		FrontendDir: frontendDir,
	}

	foundBackend, err := runner.findBackendDir()
	if err != nil {
		t.Fatalf("expected backend dir to be found: %v", err)
	}
	if foundBackend != backendDir {
		t.Errorf("expected backend dir %s, got %s", backendDir, foundBackend)
	}

	foundFrontend, err := runner.findFrontendDir()
	if err != nil {
		t.Fatalf("expected frontend dir to be found: %v", err)
	}
	if foundFrontend != frontendDir {
		t.Errorf("expected frontend dir %s, got %s", frontendDir, foundFrontend)
	}
}

