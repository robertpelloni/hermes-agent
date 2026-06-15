package orchestration

import (
	"fmt"
	"os"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/adapters"
	"github.com/robertpelloni/hermes-agent/pkg/repomap"
)

// PlanRequest holds parameters for building an execution plan.
type PlanRequest struct {
	Prompt       string `json:"prompt"`
	WorkingDir   string `json:"workingDir,omitempty"`
	IncludeRepo  bool   `json:"includeRepo"`
	MaxRepoFiles int    `json:"maxRepoFiles,omitempty"`
	TaskType     string `json:"taskType,omitempty"`
	Cost         string `json:"cost,omitempty"`
	RequireLocal bool   `json:"requireLocal,omitempty"`
}

// PlanResult represents a generated execution plan.
type PlanResult struct {
	Prompt            string      `json:"prompt"`
	TaskType          string      `json:"taskType"`
	ProviderRoute     adapters.ProviderRoute `json:"providerRoute"`
	RepoMap           string      `json:"repoMap,omitempty"`
	RepoMapIncluded   bool        `json:"repoMapIncluded"`
	Steps             []string    `json:"steps"`
	SystemContextHint string      `json:"systemContextHint,omitempty"`
}

// BuildPlan creates an execution plan for a given prompt.
func BuildPlan(req PlanRequest) (PlanResult, error) {
	cwd := strings.TrimSpace(req.WorkingDir)
	if cwd == "" {
		resolved, err := os.Getwd()
		if err != nil {
			return PlanResult{}, err
		}
		cwd = resolved
	}
	if req.MaxRepoFiles <= 0 {
		req.MaxRepoFiles = 8
	}

	// Route to appropriate provider
	route, err := adapters.SelectProvider(adapters.ProviderRouteRequest{
		TaskType:       req.TaskType,
		CostPreference: req.Cost,
		RequireLocal:   req.RequireLocal,
	})
	if err != nil {
		return PlanResult{}, err
	}

	// Derive steps based on task type
	steps := deriveSteps(route.Reason, req.Prompt)

	result := PlanResult{
		Prompt:            req.Prompt,
		TaskType:          route.Reason,
		ProviderRoute:     route,
		Steps:             steps,
		SystemContextHint: fmt.Sprintf("Route: %s/%s", route.Provider, route.Model),
	}

	// Include repo map if requested or implied
	if req.IncludeRepo || shouldIncludeRepoMap(req.Prompt) {
		mapResult, err := repomap.Generate(repomap.Options{
			BaseDir:  cwd,
			MaxFiles: req.MaxRepoFiles,
		})
		if err == nil {
			result.RepoMap = mapResult.Map
			result.RepoMapIncluded = true
		} else {
			result.Steps = append(result.Steps, fmt.Sprintf("Repo map unavailable: %v", err))
		}
	}

	return result, nil
}

func shouldIncludeRepoMap(prompt string) bool {
	lower := strings.ToLower(prompt)
	needsRepo := []string{"repo", "repository", "codebase", "file", "files", "refactor", "architecture", "search"}
	for _, needle := range needsRepo {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func deriveSteps(taskType, prompt string) []string {
	steps := []string{"Interpret the request and confirm the primary objective."}

	switch taskType {
	case "coding", "fast and cost-effective for code", "strong coding capabilities":
		steps = append(steps,
			"Inspect relevant files and gather repository context.",
			"Prepare an execution route using the provider adapter.",
			"Apply or propose the required code changes with exact-name tools.",
			"Verify the outcome and summarize next actions.",
		)
	case "analysis", "strong reasoning", "fast analysis via gemini":
		steps = append(steps,
			"Collect repository context and relevant files.",
			"Route the analysis through the selected provider profile.",
			"Synthesize findings and identify concrete next actions.",
		)
	case "chat", "fast and cheap for conversation", "cheap routing":
		steps = append(steps,
			"Gather enough context to answer accurately.",
			"Select an execution route appropriate for the task.",
			"Deliver the result with follow-up recommendations.",
		)
	default:
		steps = append(steps,
			"Gather enough context to answer accurately.",
			"Select an execution route appropriate for the task.",
			"Deliver the result with follow-up recommendations.",
		)
	}

	if strings.TrimSpace(prompt) != "" {
		steps = append(steps, fmt.Sprintf("Original request: %s", strings.TrimSpace(prompt)))
	}

	return steps
}

// FormatPlan returns a human-readable plan summary.
func (p *PlanResult) FormatPlan() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("Task Type: %s\n", p.TaskType))
	out.WriteString(fmt.Sprintf("Provider: %s/%s (%s)\n", p.ProviderRoute.Provider, p.ProviderRoute.Model, p.ProviderRoute.Cost))
	out.WriteString(fmt.Sprintf("Reason: %s\n", p.ProviderRoute.Reason))
	out.WriteString("\nSteps:\n")
	for i, step := range p.Steps {
		out.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step))
	}
	if p.RepoMapIncluded {
		out.WriteString("\nRepo Map: included\n")
		out.WriteString(p.RepoMap)
	}
	return out.String()
}