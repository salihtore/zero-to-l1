package dashboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Runner abstracts launching the dashboard services for testability.
type Runner interface {
	Start(ctx context.Context, backendPort int, frontendPort int) error
}

// DefaultRunner executes both the Go backend and Vite React frontend processes.
type DefaultRunner struct {
	BackendDir  string
	FrontendDir string
}

// Start launches the backend HTTP server and frontend dev server.
func (r *DefaultRunner) Start(ctx context.Context, backendPort int, frontendPort int) error {
	if backendPort <= 0 {
		backendPort = 8080
	}
	if frontendPort <= 0 {
		frontendPort = 5173
	}

	backendDir, err := r.findBackendDir()
	if err != nil {
		return fmt.Errorf("dashboard backend directory not found: %w", err)
	}

	frontendDir, err := r.findFrontendDir()
	if err != nil {
		return fmt.Errorf("dashboard frontend directory not found: %w", err)
	}

	fmt.Printf("🚀 Starting Zero to Secure L1 Dashboard...\n")
	fmt.Printf("   Backend : http://localhost:%d\n", backendPort)
	fmt.Printf("   Frontend: http://localhost:%d\n\n", frontendPort)
	fmt.Printf("Press Ctrl+C to stop all services.\n\n")

	// 1. Launch Backend
	backendCmd := exec.CommandContext(ctx, "go", "run", "main.go")
	backendCmd.Dir = backendDir
	backendCmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=:%d", backendPort),
	)
	backendCmd.Stdout = os.Stdout
	backendCmd.Stderr = os.Stderr

	if err := backendCmd.Start(); err != nil {
		return fmt.Errorf("failed to start dashboard backend: %w", err)
	}
	defer func() {
		if backendCmd.Process != nil {
			_ = backendCmd.Process.Kill()
		}
	}()

	// Small pause to let backend bind port
	time.Sleep(500 * time.Millisecond)

	// 2. Launch Frontend (npm run dev)
	npmBin := "npm"
	if _, err := exec.LookPath("npm.cmd"); err == nil {
		npmBin = "npm.cmd"
	}

	frontendCmd := exec.CommandContext(ctx, npmBin, "run", "dev", "--", "--port", fmt.Sprintf("%d", frontendPort))
	frontendCmd.Dir = frontendDir
	frontendCmd.Stdout = os.Stdout
	frontendCmd.Stderr = os.Stderr

	if err := frontendCmd.Start(); err != nil {
		return fmt.Errorf("failed to start dashboard frontend: %w", err)
	}
	defer func() {
		if frontendCmd.Process != nil {
			_ = frontendCmd.Process.Kill()
		}
	}()

	// Wait for context cancellation or process termination
	<-ctx.Done()
	fmt.Println("\n🛑 Shutting down dashboard services...")
	return nil
}

func (r *DefaultRunner) findBackendDir() (string, error) {
	if r.BackendDir != "" {
		if _, err := os.Stat(r.BackendDir); err == nil {
			return r.BackendDir, nil
		}
	}

	candidates := []string{
		"dashboard/backend",
		"../dashboard/backend",
		"../../dashboard/backend",
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if _, statErr := os.Stat(filepath.Join(abs, "main.go")); statErr == nil {
				return abs, nil
			}
		}
	}

	return "", fmt.Errorf("could not locate dashboard/backend/main.go")
}

func (r *DefaultRunner) findFrontendDir() (string, error) {
	if r.FrontendDir != "" {
		if _, err := os.Stat(r.FrontendDir); err == nil {
			return r.FrontendDir, nil
		}
	}

	candidates := []string{
		"dashboard/frontend",
		"../dashboard/frontend",
		"../../dashboard/frontend",
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if _, statErr := os.Stat(filepath.Join(abs, "package.json")); statErr == nil {
				return abs, nil
			}
		}
	}

	return "", fmt.Errorf("could not locate dashboard/frontend/package.json")
}

