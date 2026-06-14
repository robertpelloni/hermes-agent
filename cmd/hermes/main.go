package main

import (
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
	"github.com/robertpelloni/hermes-agent/pkg/compat"
	"github.com/robertpelloni/hermes-agent/pkg/gateway"
	"github.com/robertpelloni/hermes-agent/pkg/mcp"
	"github.com/robertpelloni/hermes-agent/pkg/memory"
	"github.com/robertpelloni/hermes-agent/pkg/repomap"
	"github.com/robertpelloni/hermes-agent/pkg/scheduler"
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
	mode := "dashboard"
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--tui":
			mode = "tui"
		case "--gateway":
			mode = "gateway"
		case "--cli":
			mode = "cli"
		case "--foundation":
			mode = "foundation"
		default:
			mode = "dashboard"
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
		
	}

	// ----- common subsystem initialization -----------------------------------
	memStore := memory.NewStore()
	skillRepo := skill.NewRepository()
	mcpServer := mcp.NewServer("hermes-agent-go", "0.1.0")
	sched := scheduler.New()

	ag := agent.New(agent.Config{
		Model:     os.Getenv("HERMES_MODEL"),
		Provider:  os.Getenv("HERMES_PROVIDER"),
		Memory:    memStore,
		Skills:    skillRepo,
		MCPServer: mcpServer,
		Scheduler: sched,
	})

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
	gw.Stop()
	fmt.Println("done.")
}
