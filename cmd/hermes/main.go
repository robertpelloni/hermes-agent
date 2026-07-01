package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/robertpelloni/hermes-agent/pkg/agent"
	"github.com/robertpelloni/hermes-agent/pkg/adapters"
	"github.com/robertpelloni/hermes-agent/pkg/assimilation"
	"github.com/robertpelloni/hermes-agent/pkg/compat"
	"github.com/robertpelloni/hermes-agent/pkg/gateway"
	"github.com/robertpelloni/hermes-agent/pkg/mcp"
	"github.com/robertpelloni/hermes-agent/pkg/memory"
	"github.com/robertpelloni/hermes-agent/pkg/modelregistry"
	"github.com/robertpelloni/hermes-agent/pkg/orchestration"
	"github.com/robertpelloni/hermes-agent/pkg/plugin"
	"github.com/robertpelloni/hermes-agent/pkg/repomap"
	"github.com/robertpelloni/hermes-agent/pkg/skill"
)

func findPython() (string, error) {
	root := findProjectRoot()
	candidates := []string{
		filepath.Join(root, ".venv", "Scripts", "python.exe"),
		filepath.Join(root, "venv", "Scripts", "python.exe"),
		filepath.Join(root, ".venv", "bin", "python"),
		filepath.Join(root, "venv", "bin", "python"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return exec.LookPath("python")
}

func findProjectRoot() string {
	d, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(d, "hermes_cli")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		d = parent
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func main() {
	fmt.Println("hermes desktop -- the self-improving ai agent")
	fmt.Println()

	// ----- mode handling ---------------------------------------------------
	// default mode launches the full dashboard (web UI with embedded TUI).
	// optional flags allow launching only specific harnesses.
	mode := "agent"
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--dashboard", "--tui":
			mode = "dashboard"
		case "--gateway":
			mode = "gateway"
		case "--cli":
			mode = "cli"
		case "--foundation":
			mode = "foundation"
		case "--help", "-h":
			fmt.Println("Usage: hermes-desktop [mode]")
			fmt.Println()
			fmt.Println("Modes (default: agent):")
			fmt.Println("  (no flags)     Agent mode – run the Go agent interactively")
			fmt.Println("  --dashboard    Start the Python web dashboard with embedded TUI")
			fmt.Println("  --tui          Start the Python TUI (same as --dashboard)")
			fmt.Println("  --gateway      Gateway-only mode (no dashboard)")
			fmt.Println("  --cli          Launch the Python CLI")
			fmt.Println("  --foundation   Show integrated feature inventory")
			fmt.Println("  --help         Show this help")
			return
		default:
			mode = "agent"
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := findProjectRoot()
	if root == "" {
		log.Fatal("could not find project root (hermes_cli/ not found)")
	}

	py, err := findPython()
	if err != nil {
		log.Fatalf("could not locate python: %v", err)
	}
	fmt.Printf("  project: %s\n", root)
	fmt.Printf("  python:  %s\n", py)
	fmt.Println()

	port := os.Getenv("HERMES_PORT")
	if port == "" {
		port = "9120"
	}

	// --------- harness specific branches ---------------------------------
	switch mode {
	case "dashboard", "tui":
		// Dashboard (or TUI) mode – same launch path, the Python web server
		// always serves the TUI when `--tui` flag is passed.
		dashRunning := false
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s", port))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			dashRunning = true
			fmt.Printf("  dashboard already running on http://127.0.0.1:%s\n", port)
		} else if err == nil {
			resp.Body.Close()
		}

		var dashboardCmd *exec.Cmd
		if !dashRunning {
			args := []string{"-m", "hermes_cli.main", "web", "--port", port, "--tui"}
			if mode == "tui" {
				// TUI flag already present in args above; keep it.
			}
			fmt.Printf("  starting dashboard on http://127.0.0.1:%s ...\n", port)

			dashboardCmd = exec.CommandContext(ctx, py, args...)
			dashboardCmd.Dir = root
			dashboardCmd.Stdout = os.Stdout
			dashboardCmd.Stderr = os.Stderr
			dashboardCmd.Env = os.Environ()

			if err := dashboardCmd.Start(); err != nil {
				log.Fatalf("failed to start dashboard: %v", err)
			}

			fmt.Print("  waiting for server")
			for i := 0; i < 60; i++ {
				dresp, derr := http.Get(fmt.Sprintf("http://127.0.0.1:%s", port))
				if derr == nil && dresp.StatusCode == http.StatusOK {
					dresp.Body.Close()
					fmt.Println()
					break
				}
				if derr == nil {
					dresp.Body.Close()
				}
				fmt.Print(".")
				time.Sleep(1 * time.Second)
			}
		}

		url := fmt.Sprintf("http://127.0.0.1:%s", port)
		fmt.Printf("\n  dashboard ready at %s\n", url)
		fmt.Println()

		openBrowser(url)

	case "gateway":
		// Gateway‑only mode – do **not** start the Python dashboard.
		fmt.Println("starting in gateway‑only mode (no Python dashboard)")

	case "cli":
		// Direct Python CLI mode – runs `hermes_cli.main cli`.
		fmt.Println("launching hermes CLI (python) …")
		cliCmd := exec.CommandContext(ctx, py, "-m", "hermes_cli.main", "cli")
		cliCmd.Dir = root
		cliCmd.Stdout = os.Stdout
		cliCmd.Stderr = os.Stderr
		cliCmd.Env = os.Environ()
		if err := cliCmd.Start(); err != nil {
			log.Fatalf("failed to start hermes CLI: %v", err)
		}
		defer cliCmd.Process.Kill()

	case "agent":
		// Agent mode – run the Go agent interactively via stdin/stdout.
		fmt.Println("running Go agent interactively. type '/quit' or '/exit' to stop.")
		memStore := memory.NewStore()
		ag := agent.New(agent.DefaultConfig(), memStore)
		if err := ag.Run(ctx); err != nil {
			log.Printf("agent run error: %v", err)
		}
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("\n> ")
			if !scanner.Scan() {
				break
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if line == "/quit" || line == "/exit" {
				fmt.Println("exiting agent")
				break
			}
			resp, err := ag.HandleMessage(ctx, "cli", "local", line)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(resp)
			}
		}
		return
		return

	case "foundation":
		// Foundation mode – exposes integrated features from pi-mono and hyperharness.
		fmt.Println("foundation mode – integrated features from pi-mono and hyperharness")
		fmt.Println()
		
		// Show tool parity catalog
		catalog := compat.DefaultCatalog()
		fmt.Println(catalog.Summary())
		
		// Show provider routing info
		fmt.Println("\nProvider Routing:")
		fmt.Println(adapters.FormatProviderInfo())
		
		// Show repo map if possible
		fmt.Println("\nRepo Map (top 10 files):")
		result, err := repomap.Generate(repomap.Options{
			BaseDir:  root,
			MaxFiles: 10,
		})
		if err != nil {
			fmt.Printf("  (error generating: %v)\n", err)
		} else {
			for _, f := range result.Files {
				fmt.Printf("  %s (%d lines) – %s\n", f.Path, f.Lines, f.Reason)
			}
		}
		
		// Show model registry status
		fmt.Println("\nModel Registry:")
		mr := modelregistry.NewModelRegistry()
		fmt.Println(mr.FormatProviderStatus())
		
		// Show assimilation inventory
		fmt.Println("\nAssimilation Inventory:")
		fmt.Println(assimilation.FormatInventory())
		
		// Demo plan generation
		fmt.Println("\nSample Execution Plan:")
		planReq := orchestration.PlanRequest{
			Prompt:       "List all Go files in this repository and summarize their purpose.",
			IncludeRepo:  true,
			MaxRepoFiles: 8,
		}
		plan, err := orchestration.BuildPlan(planReq)
		if err != nil {
			fmt.Printf("  (error generating plan: %v)\n", err)
		} else {
			fmt.Println(plan.FormatPlan())
		}
		
	}

	// ----- common subsystem initialization -----------------------------------
	memStore := memory.NewStore()
	mcpServer := mcp.NewServer("hermes-agent-go", "0.1.0")
	pluginMgr := plugin.NewManager()
	if discovered, err := pluginMgr.Discover(); err == nil && len(discovered) > 0 {
		fmt.Printf("  plugins discovered: %v\n", discovered)
	}

	skillRepo := skill.Global()
	if err := skillRepo.DiscoverAndLoad(filepath.Join(root, "skills")); err == nil {
		fmt.Printf("  skills loaded: %d\n", len(skillRepo.List()))
	}

	ag := agent.New(agent.DefaultConfig(), memStore)

	// Start MCP server in background
	go func() {
		mcpAddr := os.Getenv("HERMES_MCP_PORT")
		if mcpAddr == "" {
			mcpAddr = ":9090"
		}
		fmt.Printf("  MCP server starting on %s...\n", mcpAddr)
		if err := mcpServer.Start(mcpAddr); err != nil {
			log.Printf("MCP server error: %v", err)
		}
	}()

	gw := gateway.New(ag)
	if err := gw.Start(ctx); err != nil {
		log.Printf("gateway warning: %v", err)
	}

	fmt.Println("press ctrl+c to stop")
	fmt.Println()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println()
	fmt.Println("shutting down...")
	cancel()
	memStore.Close()
	gw.Stop()
	fmt.Println("done.")
}
