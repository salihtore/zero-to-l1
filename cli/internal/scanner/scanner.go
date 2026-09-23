package scanner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	ScannerRequirementsDoc = "https://github.com/salihtore/zero-to-l1/blob/main/security-scanner/README.md"
)

// Runner abstracts running the security scanner for testability.
type Runner interface {
	Scan(ctx context.Context, targetPath string) error
}

// DefaultRunner executes the python-based Slither security scanner.
type DefaultRunner struct {
	PythonPath  string // optional custom python executable
	ScannerPath string // optional custom path to scanner.py
}

// Scan executes security-scanner/scanner.py on the given target solidity contract/directory.
func (r *DefaultRunner) Scan(ctx context.Context, targetPath string) error {
	if targetPath == "" {
		return fmt.Errorf("target contract or directory path cannot be empty")
	}

	// Verify target exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("target path '%s' does not exist", targetPath)
	}

	pythonBin, err := r.findPython(ctx)
	if err != nil {
		fmt.Printf("\n❌ ERROR: Python 3 with Slither is required to run the security scanner.\n")
		fmt.Printf("👉 Setup Guide: %s\n", ScannerRequirementsDoc)
		fmt.Printf("👉 Installation: pip install -r security-scanner/requirements.txt\n\n")
		return fmt.Errorf("python environment check failed: %w", err)
	}

	scannerScript, err := r.findScannerScript()
	if err != nil {
		return fmt.Errorf("scanner script check failed: %w", err)
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}

	fmt.Printf("🔍 Scanning target: %s\n", absTarget)

	cmd := exec.CommandContext(ctx, pythonBin, scannerScript, "--target", absTarget)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if runErr := cmd.Run(); runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			return fmt.Errorf("security scanner finished with exit code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("security scanner execution failed: %w", runErr)
	}

	return nil
}

// findPython locates python3 or python in system PATH.
func (r *DefaultRunner) findPython(ctx context.Context) (string, error) {
	if r.PythonPath != "" {
		return r.PythonPath, nil
	}

	candidates := []string{"python3", "python", "py"}
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			// Verify python runs
			cmd := exec.CommandContext(ctx, path, "--version")
			if err := cmd.Run(); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("python3/python binary not found on PATH")
}

// findScannerScript locates security-scanner/scanner.py in repo layout.
func (r *DefaultRunner) findScannerScript() (string, error) {
	if r.ScannerPath != "" {
		if _, err := os.Stat(r.ScannerPath); err == nil {
			return r.ScannerPath, nil
		}
	}

	candidates := []string{
		"security-scanner/scanner.py",
		"../security-scanner/scanner.py",
		"../../security-scanner/scanner.py",
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if _, statErr := os.Stat(abs); statErr == nil {
				return abs, nil
			}
		}
	}

	return "", fmt.Errorf("could not locate security-scanner/scanner.py")
}
