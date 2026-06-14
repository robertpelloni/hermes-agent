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
	"github.com/robertpelloni/hermes-agent/pkg/gateway"
	"github.com/robertpelloni/hermes-agent/pkg/mcp"
	"github.com/robertpelloni/hermes-agent/pkg/memory"
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

	// check if dashboard is already running
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

	// initialize subsystems (for future go-native expansion)
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
