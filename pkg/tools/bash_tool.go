package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func init() {
	// Register bash
	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "bash",
		Description: "Execute a bash command. Returns stdout, stderr, and exit code.",
		Category:    "terminal",
		Parameters: map[string]any{
			"command":   map[string]string{"type": "string", "description": "The bash command to execute"},
			"timeout":   map[string]string{"type": "string", "description": "Timeout (e.g. '30s', '5m'), optional"},
			"workdir":   map[string]string{"type": "string", "description": "Working directory, optional"},
			"background": map[string]string{"type": "string", "description": "Run in background ('true'), optional"},
		},
		Handler: bashHandler,
		Native:  true,
	})
}

type BashResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

func bashHandler(args map[string]any, ctxParams map[string]any) (any, error) {
	cmdRaw, ok := args["command"]
	if !ok {
		return nil, fmt.Errorf("missing 'command' argument")
	}
	command, ok := cmdRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'command' must be a string")
	}

	timeoutStr := "30s"
	if tRaw, ok := args["timeout"]; ok {
		if t, ok := tRaw.(string); ok {
			timeoutStr = t
		}
	}

	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		duration = 30 * time.Second
	}

	workdir := ""
	if wRaw, ok := args["workdir"]; ok {
		if w, ok := wRaw.(string); ok {
			workdir = w
		}
	}

	// Background mode
	if bgRaw, ok := args["background"]; ok {
		if bg, ok := bgRaw.(string); ok && strings.ToLower(bg) == "true" {
			go func() {
				cmd := exec.Command("cmd", "/c", command)
				if workdir != "" {
					cmd.Dir = workdir
				}
				_ = cmd.Start()
			}()
			return map[string]string{"status": "started", "message": "Command started in background"}, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "cmd", "/c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if workdir != "" {
		cmd.Dir = workdir
	}
	cmd.Env = os.Environ()

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return BashResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
