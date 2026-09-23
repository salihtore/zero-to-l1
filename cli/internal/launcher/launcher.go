package launcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

const (
	InstallDocURL = "https://docs.avax.network/tooling/cli-guides/install-avalanche-cli"
	FaucetURL     = "https://core.app/tools/testnet-faucet/"
)

// LaunchConfig contains the user inputs for blockchain creation.
type LaunchConfig struct {
	ChainName   string `json:"chainName"`
	TokenSymbol string `json:"tokenSymbol"`
	ChainID     int64  `json:"chainID"`
	VMType      string `json:"vmType"`
}

// DeploymentRecord holds the deployment details saved to deployments/<chainName>.json.
type DeploymentRecord struct {
	ChainName    string    `json:"chainName"`
	TokenSymbol  string    `json:"tokenSymbol"`
	BlockchainID string    `json:"blockchainID"`
	RPCURL       string    `json:"rpcURL"`
	ChainID      int64     `json:"chainID"`
	Network      string    `json:"network"`
	DeployedAt   time.Time `json:"deployedAt"`
	Status       string    `json:"status"` // "deployed" or "partial"
}

// Runner abstracts avalanche-cli execution for testability and cross-environment (Native/WSL) support.
type Runner interface {
	CheckInstalled(ctx context.Context) (mode string, execPath string, err error)
	CreateConfig(ctx context.Context, config LaunchConfig) error
	Deploy(ctx context.Context, chainName string, network string) (output string, err error)
}

// DefaultRunner executes real avalanche-cli subprocess commands.
type DefaultRunner struct {
	Mode     string
	ExecPath string
}

func (r *DefaultRunner) CheckInstalled(ctx context.Context) (string, string, error) {
	// 1. Check Native Windows PATH
	if p, err := exec.LookPath("avalanche"); err == nil {
		r.Mode = "native"
		r.ExecPath = p
		return r.Mode, r.ExecPath, nil
	}

	// 2. Check WSL Ubuntu environment
	if wslPath, err := exec.LookPath("wsl"); err == nil {
		// Test running avalanche in WSL (checking PATH or common binary location)
		cmd := exec.CommandContext(ctx, wslPath, "bash", "-l", "-c", "which avalanche || find ~/bin /usr/local/bin ~/.avalanche-cli/bin -name avalanche 2>/dev/null | head -n 1")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			foundPath := strings.TrimSpace(out.String())
			if foundPath != "" && !strings.Contains(foundPath, "not found") {
				r.Mode = "wsl"
				r.ExecPath = foundPath
				return r.Mode, r.ExecPath, nil
			}
		}
	}

	return "", "", fmt.Errorf("avalanche-cli binary not found in system PATH or WSL environment")
}

func (r *DefaultRunner) CreateConfig(ctx context.Context, config LaunchConfig) error {
	var cmd *exec.Cmd
	// Do not pass --vm-version. CLI v1.9.6 resolves it to Subnet-EVM v0.8.0,
	// incompatible with AvalancheGo v1.15.0-fuji (RPCChainVM v46).
	// TODO: automate the matching-plugin flow in scripts/build-subnet-evm-plugin.sh.
	args := []string{"blockchain", "create", config.ChainName, "--force"}

	if r.Mode == "wsl" {
		wslArgs := append([]string{r.ExecPath}, args...)
		cmd = exec.CommandContext(ctx, "wsl", wslArgs...)
	} else {
		cmd = exec.CommandContext(ctx, r.ExecPath, args...)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n🚀 Launching avalanche-cli creation wizard for blockchain '%s'...\n", config.ChainName)
	return cmd.Run()
}

func (r *DefaultRunner) Deploy(ctx context.Context, chainName string, network string) (string, error) {
	// 60 minutes timeout for Fuji archive download & node bootstrapping
	deployCtx, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	args := []string{"blockchain", "deploy", chainName, "--" + network}

	if r.Mode == "wsl" {
		wslArgs := append([]string{r.ExecPath}, args...)
		cmd = exec.CommandContext(deployCtx, "wsl", wslArgs...)
	} else {
		cmd = exec.CommandContext(deployCtx, r.ExecPath, args...)
	}

	cmd.Stdin = os.Stdin
	var buf bytes.Buffer
	multiOut := io.MultiWriter(os.Stdout, &buf)
	multiErr := io.MultiWriter(os.Stderr, &buf)
	cmd.Stdout = multiOut
	cmd.Stderr = multiErr

	fmt.Printf("\n🌐 Deploying blockchain '%s' to Avalanche %s Testnet...\n", chainName, network)
	fmt.Println("ℹ️  İlk deploy denemesinde Fuji ağının arşivi indirilecek, bu işlem internet hızına bağlı olarak 10-30 dakika sürebilir. Lütfen bekleyin, terminali kapatmayın.")

	err := cmd.Run()
	return buf.String(), err
}

// ParseDeployOutput extracts Blockchain ID, RPC URL, and Chain ID from avalanche-cli deploy logs.
func ParseDeployOutput(output string) (blockchainID string, rpcURL string, chainID int64, err error) {
	bcIDRegex := regexp.MustCompile(`(?i)\bBlockchain\s*ID:\s*([A-Za-z0-9]+)`)
	rpcURLRegex := regexp.MustCompile(`(?i)\bRPC\s*URL:\s*(https?://[^\s]+)`)
	chainIDRegex := regexp.MustCompile(`(?i)\bChain\s*ID:\s*([0-9]+)`)

	bcMatches := bcIDRegex.FindStringSubmatch(output)
	if len(bcMatches) > 1 {
		blockchainID = bcMatches[1]
	}

	rpcMatches := rpcURLRegex.FindStringSubmatch(output)
	if len(rpcMatches) > 1 {
		rpcURL = rpcMatches[1]
	}

	chainIDMatches := chainIDRegex.FindStringSubmatch(output)
	if len(chainIDMatches) > 1 {
		if val, pErr := strconv.ParseInt(chainIDMatches[1], 10, 64); pErr == nil {
			chainID = val
		}
	}

	if blockchainID == "" && rpcURL == "" && chainID == 0 {
		return "", "", 0, fmt.Errorf("could not parse deployment details from output")
	}

	return blockchainID, rpcURL, chainID, nil
}

// SaveDeploymentRecord writes deployment metadata to deployments/<chainName>.json.
func SaveDeploymentRecord(record DeploymentRecord, deploymentsDir string) error {
	if err := os.MkdirAll(deploymentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create deployments directory: %w", err)
	}

	filePath := filepath.Join(deploymentsDir, fmt.Sprintf("%s.json", record.ChainName))
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deployment record: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// PromptUserInput interactively collects chain configuration from the user.
func PromptUserInput() (LaunchConfig, error) {
	nameRegex := regexp.MustCompile(`^[a-zA-Z]+$`)

	promptChainName := promptui.Prompt{
		Label:   "Chain Name",
		Default: "mychain",
		Validate: func(input string) error {
			trimmed := strings.TrimSpace(input)
			if trimmed == "" {
				return fmt.Errorf("chain name cannot be empty")
			}
			if !nameRegex.MatchString(trimmed) {
				return fmt.Errorf("chain name must contain only letters (a-z, A-Z)")
			}
			return nil
		},
	}
	chainName, err := promptChainName.Run()
	if err != nil {
		return LaunchConfig{}, fmt.Errorf("prompt cancelled: %w", err)
	}

	promptTokenSymbol := promptui.Prompt{
		Label:   "Token Symbol",
		Default: "MYT",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("token symbol cannot be empty")
			}
			return nil
		},
	}
	tokenSymbol, err := promptTokenSymbol.Run()
	if err != nil {
		return LaunchConfig{}, fmt.Errorf("prompt cancelled: %w", err)
	}

	// Propose random EVM Chain ID between 90001 and 99999
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomChainID := rnd.Intn(99999-90001+1) + 90001

	promptChainID := promptui.Prompt{
		Label:   "EVM Chain ID",
		Default: strconv.Itoa(randomChainID),
		Validate: func(input string) error {
			trimmed := strings.TrimSpace(input)
			if trimmed == "" {
				return nil
			}
			val, err := strconv.ParseInt(trimmed, 10, 64)
			if err != nil || val <= 0 {
				return fmt.Errorf("EVM Chain ID must be a positive integer")
			}
			return nil
		},
	}
	chainIDStr, err := promptChainID.Run()
	if err != nil {
		return LaunchConfig{}, fmt.Errorf("prompt cancelled: %w", err)
	}

	chainIDVal, err := strconv.ParseInt(strings.TrimSpace(chainIDStr), 10, 64)
	if err != nil {
		chainIDVal = int64(randomChainID)
	}

	promptVM := promptui.Select{
		Label: "VM Type",
		Items: []string{"Subnet-EVM"},
	}
	_, vmType, err := promptVM.Run()
	if err != nil {
		return LaunchConfig{}, fmt.Errorf("prompt cancelled: %w", err)
	}

	fmt.Println("\n💡 Not: Token Symbol ve Chain ID'yi az sonra avalanche-cli'nin kendi sorularında da gireceksin, tutarlı olması için aynı değerleri kullan.")

	return LaunchConfig{
		ChainName:   strings.TrimSpace(chainName),
		TokenSymbol: strings.TrimSpace(tokenSymbol),
		ChainID:     chainIDVal,
		VMType:      strings.TrimSpace(vmType),
	}, nil
}

// ExecuteLaunch orchestrates the entire launch workflow.
func ExecuteLaunch(ctx context.Context, runner Runner, deploymentsDir string) error {
	fmt.Println("🔍 Checking avalanche-cli installation...")
	mode, execPath, err := runner.CheckInstalled(ctx)
	if err != nil {
		fmt.Printf("\n❌ ERROR: avalanche-cli is not installed on your system or WSL environment.\n")
		fmt.Printf("👉 Official Installation Guide: %s\n", InstallDocURL)
		fmt.Printf("👉 WSL Installation Command: curl -sSfL https://raw.githubusercontent.com/ava-labs/avalanche-cli/main/scripts/install.sh | sh\n\n")
		return fmt.Errorf("precheck failed: avalanche-cli missing")
	}

	fmt.Printf("✅ Found avalanche-cli (%s mode: %s)\n\n", mode, execPath)

	// ── Pre-check: version & RPCChainVM protocol compatibility ───────────────
	// avalanche-cli v1.9.6 ships Subnet-EVM v0.8.0 (RPCChainVM protocol v44)
	// but Fuji requires AvalancheGo v1.15.0 (RPCChainVM protocol v46).
	// Warn the user and ask for confirmation before proceeding with deployment.
	checker := &DefaultCLIVersionChecker{
		WSLAvalanchePath: execPath,
	}
	precheckCtx, precheckCancel := context.WithTimeout(ctx, 20*time.Second)
	result, _ := CheckAvalancheCLIUpdate(precheckCtx, checker)
	precheckCancel()

	if result != nil && result.ProtocolMismatchDetected {
		fmt.Println("\n⚠️  WARNING: A RPCChainVM protocol mismatch has been detected.")
		fmt.Println("   Proceeding will likely result in a failed deploy or a node that cannot track your blockchain.")
		fmt.Print("\n   Continue anyway? (y/N): ")
		var answer string
		if _, scanErr := fmt.Scanln(&answer); scanErr != nil || !strings.EqualFold(strings.TrimSpace(answer), "y") {
			return fmt.Errorf("aborted by user due to RPCChainVM protocol mismatch — fix the plugin first")
		}
		fmt.Println()
	}
	// ─────────────────────────────────────────────────────────────────────────

	config, err := PromptUserInput()
	if err != nil {
		return fmt.Errorf("user input failed: %w", err)
	}

	// Step 1: Create Config
	if err := runner.CreateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to create blockchain config: %w", err)
	}

	// Create partial deployment record in case deploy fails later
	partialRecord := DeploymentRecord{
		ChainName:   config.ChainName,
		TokenSymbol: config.TokenSymbol,
		ChainID:     config.ChainID,
		Network:     "fuji",
		DeployedAt:  time.Now().UTC(),
		Status:      "partial",
	}
	_ = SaveDeploymentRecord(partialRecord, deploymentsDir)

	// Step 2: Deploy to Fuji
	deployOutput, err := runner.Deploy(ctx, config.ChainName, "fuji")
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "deadline exceeded") {
			fmt.Printf("\n⚠️  DEPLOYMENT TIMEOUT: Fuji ağının arşivinin indirilmesi 60 dakikalık zaman aşımı süresini aştı.\n")
			fmt.Printf("💡 Likely Cause: Yavaş internet bağlantısı veya Fuji düğümünün bootstrap işleminin uzun sürmesi.\n")
			fmt.Printf("ℹ️  Kaldığın yerden 'avalanche blockchain deploy %s --fuji' komutu ile devam edebilirsin.\n\n", config.ChainName)
		} else {
			fmt.Printf("\n⚠️  DEPLOYMENT ERROR: avalanche-cli deploy failed: %v\n", err)
			fmt.Printf("💡 Likely Cause: Testnet cüzdanınızda Fuji gaz ücretleri için yeterli test AVAX olmayabilir.\n")
			fmt.Printf("👉 Get free testnet AVAX from the official faucet: %s\n", FaucetURL)
		}
		fmt.Printf("ℹ️  A partial deployment record has been saved to '%s/%s.json'.\n\n", deploymentsDir, config.ChainName)
		return fmt.Errorf("deploy failed: %w", err)
	}

	// Step 3: Parse Output & Update Record
	bcID, rpcURL, parsedChainID, parseErr := ParseDeployOutput(deployOutput)
	if parseErr != nil {
		fmt.Printf("⚠️  Warning: Deployment succeeded but output details could not be parsed: %v\n", parseErr)
	}

	finalChainID := config.ChainID
	if parsedChainID != 0 {
		finalChainID = parsedChainID
	}

	finalRecord := DeploymentRecord{
		ChainName:    config.ChainName,
		TokenSymbol:  config.TokenSymbol,
		BlockchainID: bcID,
		RPCURL:       rpcURL,
		ChainID:      finalChainID,
		Network:      "fuji",
		DeployedAt:   time.Now().UTC(),
		Status:       "deployed",
	}

	if err := SaveDeploymentRecord(finalRecord, deploymentsDir); err != nil {
		return fmt.Errorf("failed to save final deployment record: %w", err)
	}

	fmt.Printf("\n🎉 SUCCESS! Avalanche L1/Subnet '%s' successfully deployed to Fuji!\n", config.ChainName)
	fmt.Printf("📄 Deployment record saved to: %s/%s.json\n", deploymentsDir, config.ChainName)
	if bcID != "" {
		fmt.Printf("🔗 Blockchain ID: %s\n", bcID)
	}
	if rpcURL != "" {
		fmt.Printf("🌐 RPC URL:       %s\n", rpcURL)
	}
	if finalChainID != 0 {
		fmt.Printf("🔢 Chain ID:      %d\n", finalChainID)
	}

	return nil
}
