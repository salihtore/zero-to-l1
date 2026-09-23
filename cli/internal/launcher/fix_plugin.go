package launcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PluginFixer abstracts running the build-subnet-evm-plugin script.
type PluginFixer interface {
	FixPlugin(ctx context.Context, chainName string, tag string) error
}

// DefaultPluginFixer executes the build script inside WSL.
type DefaultPluginFixer struct {
	ScriptPath string // relative or absolute path to build-subnet-evm-plugin.sh
}

// FixPlugin runs the build-subnet-evm-plugin.sh script inside WSL for the specified chain.
func (f *DefaultPluginFixer) FixPlugin(ctx context.Context, chainName string, tag string) error {
	if chainName == "" {
		chainName = "zrgchain"
	}
	if tag == "" {
		tag = "v1.15.0"
	}

	wslPath, err := exec.LookPath("wsl")
	if err != nil {
		return fmt.Errorf("WSL is required to build the Subnet-EVM plugin on Windows: %w", err)
	}

	script := f.ScriptPath
	if script == "" {
		// Resolve relative to working directory or standard repo layout
		for _, candidate := range []string{
			"cli/scripts/build-subnet-evm-plugin.sh",
			"scripts/build-subnet-evm-plugin.sh",
			"../scripts/build-subnet-evm-plugin.sh",
		} {
			if _, statErr := os.Stat(candidate); statErr == nil {
				script = candidate
				break
			}
		}
	}

	if script == "" {
		return fmt.Errorf("could not find build-subnet-evm-plugin.sh script")
	}

	absPath, err := filepath.Abs(script)
	if err != nil {
		return fmt.Errorf("failed to resolve script absolute path: %w", err)
	}

	// Convert Windows path (e.g. C:\Users\... to /mnt/c/Users/...) for WSL
	wslScriptPath := windowsToWSLPath(absPath)

	fmt.Printf("🔨 Building protocol-compatible Subnet-EVM plugin for '%s' (AvalancheGo %s)...\n", chainName, tag)
	fmt.Printf("   Running script: %s\n\n", wslScriptPath)

	cmd := exec.CommandContext(ctx, wslPath, "bash", wslScriptPath, chainName, tag)
	cmd.Stdin = os.Stdin

	var buf bytes.Buffer
	multiOut := io.MultiWriter(os.Stdout, &buf)
	multiErr := io.MultiWriter(os.Stderr, &buf)
	cmd.Stdout = multiOut
	cmd.Stderr = multiErr

	if runErr := cmd.Run(); runErr != nil {
		return fmt.Errorf("failed to build Subnet-EVM plugin: %w\nOutput:\n%s", runErr, buf.String())
	}

	return nil
}

// windowsToWSLPath converts a standard Windows path to its /mnt/<drive>/ format in WSL.
func windowsToWSLPath(winPath string) string {
	cleaned := filepath.Clean(winPath)
	vol := filepath.VolumeName(cleaned)
	if len(vol) >= 2 && vol[1] == ':' {
		driveLetter := strings.ToLower(string(vol[0]))
		rest := cleaned[len(vol):]
		rest = filepath.ToSlash(rest)
		return fmt.Sprintf("/mnt/%s%s", driveLetter, rest)
	}
	return filepath.ToSlash(cleaned)
}
