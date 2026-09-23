package launcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// RPCChainVM protocol versions (as of AvalancheGo v1.15.0 / Helicon upgrade)
const (
	// AvalancheGoV1150Protocol is the RPCChainVM protocol version required by AvalancheGo v1.15.0.
	AvalancheGoV1150Protocol = 46

	// SubnetEVMV080Protocol is the RPCChainVM protocol version implemented by Subnet-EVM v0.8.0.
	// This version is incompatible with AvalancheGo v1.15.0.
	SubnetEVMV080Protocol = 44

	// CliLatestReleaseAPI is the GitHub API endpoint for avalanche-cli releases.
	CliLatestReleaseAPI = "https://api.github.com/repos/ava-labs/avalanche-cli/releases/latest"

	// SubnetEVMBuildGuide explains how to rebuild the compatible Subnet-EVM plugin.
	SubnetEVMBuildGuide = `
┌─────────────────────────────────────────────────────────────────────────┐
│  🔧  RPCChainVM Protocol Uyumsuzluğu — Manuel Çözüm                    │
├─────────────────────────────────────────────────────────────────────────┤
│  AvalancheGo v1.15.0 → RPCChainVM protokol v46 gerektiriyor            │
│  Subnet-EVM v0.8.0   → RPCChainVM protokol v44 implemente ediyor       │
│                                                                         │
│  ✅ Çözüm: Subnet-EVM'i AvalancheGo v1.15.0 kaynak kodundan derle:    │
│                                                                         │
│  # WSL/Linux terminalinde çalıştır:                                     │
│  git clone https://github.com/ava-labs/avalanchego                      │
│  cd avalanchego                                                         │
│  git checkout v1.15.0                                                   │
│  cd graft/subnet-evm                                                    │
│  go build -o ~/.avalanche-cli/local/zrgchain-local-node-fuji/           │
│           plugins/vvk1HuzGxFqyuLauWNFBjg9JBQ2GpUwd4pT3QMGdF8KUA1psW  │
│                                                                         │
│  📌 Plugin binary adı VM-ID ile aynı olmalı.                           │
│  📌 AvalancheGo v1.14.x'e geçmek KESİNLİKLE çözüm değildir —         │
│     Fuji ağı v1.15.0 (Helicon) gerektiriyor.                           │
└─────────────────────────────────────────────────────────────────────────┘`
)

// GitHubRelease holds the relevant fields from the GitHub releases API response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"` // release notes
}

// CLIVersionChecker abstracts version/protocol checks for testability.
type CLIVersionChecker interface {
	// FetchLatestRelease retrieves the latest avalanche-cli release from GitHub.
	FetchLatestRelease(ctx context.Context) (*GitHubRelease, error)
	// GetLocalCLIVersion returns the installed avalanche-cli version string (e.g. "v1.9.6").
	GetLocalCLIVersion(ctx context.Context) (string, error)
	// GetLocalAvalancheGoVersion returns the local avalanchego version string if available via WSL.
	GetLocalAvalancheGoVersion(ctx context.Context) (string, error)
}

// DefaultCLIVersionChecker performs real network & subprocess calls.
type DefaultCLIVersionChecker struct {
	// WSLAvalanchePath is the full path to the avalanche binary inside WSL (e.g. /home/salih/bin/avalanche).
	WSLAvalanchePath string
}

// FetchLatestRelease calls the GitHub releases API and returns the latest release info.
func (c *DefaultCLIVersionChecker) FetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, CliLatestReleaseAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build GitHub API request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	return &release, nil
}

// GetLocalCLIVersion runs `avalanche --version` (via WSL if needed) and parses the version string.
func (c *DefaultCLIVersionChecker) GetLocalCLIVersion(ctx context.Context) (string, error) {
	var out bytes.Buffer
	var cmd *exec.Cmd

	avalanchePath := c.WSLAvalanchePath
	if avalanchePath == "" {
		avalanchePath = "avalanche"
	}

	// Try WSL first
	if _, err := exec.LookPath("wsl"); err == nil {
		cmd = exec.CommandContext(ctx, "wsl", avalanchePath, "--version")
	} else {
		cmd = exec.CommandContext(ctx, avalanchePath, "--version")
	}

	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run avalanche --version: %w", err)
	}

	return parseVersionFromOutput(out.String()), nil
}

// GetLocalAvalancheGoVersion attempts to find the avalanchego binary in WSL and return its version.
func (c *DefaultCLIVersionChecker) GetLocalAvalancheGoVersion(ctx context.Context) (string, error) {
	if _, err := exec.LookPath("wsl"); err != nil {
		return "", fmt.Errorf("wsl not available")
	}

	var out bytes.Buffer
	// Look for avalanchego in common locations or via the avalanche-cli plugins dir
	cmd := exec.CommandContext(ctx, "wsl", "bash", "-lc",
		`avalanchego --version 2>/dev/null || `+
			`find ~/.avalanche-cli/local -name 'avalanchego' -type f 2>/dev/null | head -1 | xargs -I{} {} --version 2>/dev/null || `+
			`echo "not found"`)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to detect avalanchego version: %w", err)
	}

	output := strings.TrimSpace(out.String())
	if output == "" || output == "not found" {
		return "", fmt.Errorf("avalanchego binary not found in WSL")
	}

	return parseVersionFromOutput(output), nil
}

// parseVersionFromOutput extracts a vX.Y.Z semver token from arbitrary command output.
func parseVersionFromOutput(output string) string {
	for _, field := range strings.Fields(output) {
		if strings.HasPrefix(field, "v") && strings.Count(field, ".") >= 1 {
			return strings.TrimRight(field, ",;:")
		}
	}
	return strings.TrimSpace(output)
}

// CLIUpdateResult summarises the outcome of CheckAvalancheCLIUpdate.
type CLIUpdateResult struct {
	LocalVersion  string
	LatestVersion string
	IsUpToDate    bool
	// BuiltInVMNoted is true when the release notes mention "built-in VM",
	// which would mean the plugin situation has changed upstream.
	BuiltInVMNoted bool
	// ProtocolMismatchDetected is set when we know the locally installed
	// Subnet-EVM plugin uses an older RPCChainVM protocol than AvalancheGo requires.
	ProtocolMismatchDetected bool
}

// CheckAvalancheCLIUpdate queries GitHub for the latest avalanche-cli release,
// compares it against the locally installed version, and additionally checks for
// the known RPCChainVM protocol mismatch between AvalancheGo v1.15.0 and
// Subnet-EVM v0.8.0 (protocol v46 vs v44).
//
// It prints a human-readable status report and returns the full result for
// programmatic use.
func CheckAvalancheCLIUpdate(ctx context.Context, checker CLIVersionChecker) (*CLIUpdateResult, error) {
	fmt.Println("🔍 Checking avalanche-cli version & Subnet-EVM protocol compatibility...")

	result := &CLIUpdateResult{}

	// ── 1. Local CLI version ──────────────────────────────────────────────────
	localVer, err := checker.GetLocalCLIVersion(ctx)
	if err != nil {
		fmt.Printf("⚠️  Could not determine local avalanche-cli version: %v\n", err)
		localVer = "unknown"
	}
	result.LocalVersion = localVer

	// ── 2. Latest release from GitHub ─────────────────────────────────────────
	release, err := checker.FetchLatestRelease(ctx)
	if err != nil {
		fmt.Printf("⚠️  Could not fetch latest release from GitHub: %v\n", err)
		fmt.Println("   (Skipping update check — please verify manually)")
	} else {
		result.LatestVersion = release.TagName
		result.IsUpToDate = normalizeVersion(localVer) == normalizeVersion(release.TagName)
		result.BuiltInVMNoted = strings.Contains(strings.ToLower(release.Body), "built-in vm")

		if result.IsUpToDate {
			fmt.Printf("✅ avalanche-cli is up-to-date (%s)\n", localVer)
		} else {
			fmt.Printf("🆕 Update available: %s → %s\n", localVer, release.TagName)
			fmt.Printf("   Run: curl -sSfL https://raw.githubusercontent.com/ava-labs/avalanche-cli/main/scripts/install.sh | sh\n")
		}

		if result.BuiltInVMNoted {
			fmt.Printf("📢 Release notes for %s mention 'built-in VM' — the plugin situation may have changed upstream. Please review: https://github.com/ava-labs/avalanche-cli/releases/tag/%s\n",
				release.TagName, release.TagName)
		}
	}

	// ── 3. Context note: maintenance mode ─────────────────────────────────────
	// avalanche-cli entered maintenance mode in December 2025; new feature
	// development has stopped. Updating the CLI is NOT the fix for the
	// RPCChainVM protocol mismatch.
	if result.IsUpToDate || result.LatestVersion == "" {
		fmt.Println("ℹ️  avalanche-cli entered maintenance mode (Dec 2025). Updating is NOT the fix for the RPCChainVM mismatch.")
	}

	// ── 4. RPCChainVM protocol mismatch check ─────────────────────────────────
	mismatch, err := CheckRPCProtocolMismatch(ctx, checker)
	if err != nil {
		fmt.Printf("⚠️  Protocol mismatch check could not complete: %v\n", err)
	}
	result.ProtocolMismatchDetected = mismatch

	if mismatch {
		fmt.Println(SubnetEVMBuildGuide)
	}

	return result, nil
}

// CheckRPCProtocolMismatch detects whether the local AvalancheGo node uses a
// different RPCChainVM protocol version than the Subnet-EVM plugin provided by
// avalanche-cli v1.9.6.
//
// Currently this is a heuristic check:
//   - If the local AvalancheGo version is ≥ v1.15.0 AND avalanche-cli still
//     ships Subnet-EVM v0.8.0 (protocol v44), the mismatch is certain.
//
// Returns (true, nil) when a mismatch is detected, (false, nil) when safe,
// and (false, err) when detection was not possible.
func CheckRPCProtocolMismatch(ctx context.Context, checker CLIVersionChecker) (bool, error) {
	agVersion, err := checker.GetLocalAvalancheGoVersion(ctx)
	if err != nil {
		// Cannot determine → assume mismatch as a conservative warning because
		// Fuji has been running Helicon (v1.15.0) since 22 September 2026.
		fmt.Printf("ℹ️  Could not detect local AvalancheGo version (%v). Fuji runs AvalancheGo v1.15.0 (Helicon, RPCChainVM v%d).\n",
			err, AvalancheGoV1150Protocol)
		fmt.Printf("⚠️  If you are deploying to Fuji: avalanche-cli v1.9.6 installs Subnet-EVM v0.8.0 (RPCChainVM v%d), which is INCOMPATIBLE.\n",
			SubnetEVMV080Protocol)
		return true, nil
	}

	if isVersionAtLeast(agVersion, "v1.15.0") {
		fmt.Printf("⚠️  Detected AvalancheGo %s (RPCChainVM v%d required).\n", agVersion, AvalancheGoV1150Protocol)
		fmt.Printf("   avalanche-cli v1.9.6 ships Subnet-EVM v0.8.0 (RPCChainVM v%d) — INCOMPATIBLE!\n", SubnetEVMV080Protocol)
		return true, nil
	}

	fmt.Printf("✅ AvalancheGo %s detected — no RPCChainVM protocol mismatch.\n", agVersion)
	return false, nil
}

// normalizeVersion strips a leading "v" for comparison purposes.
func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// isVersionAtLeast returns true when `version` is semantically ≥ `minVersion`.
// Both arguments should be in "vX.Y.Z" format. This is a lightweight
// implementation that compares major, minor, and patch integers without
// importing semver libraries to keep the dependency footprint minimal.
func isVersionAtLeast(version, minVersion string) bool {
	parts := func(v string) [3]int {
		v = strings.TrimPrefix(strings.TrimSpace(v), "v")
		var major, minor, patch int
		fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch) //nolint:errcheck
		return [3]int{major, minor, patch}
	}

	vp := parts(version)
	mp := parts(minVersion)

	for i := range vp {
		if vp[i] > mp[i] {
			return true
		}
		if vp[i] < mp[i] {
			return false
		}
	}
	return true // equal
}
