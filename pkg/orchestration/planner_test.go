package orchestration

import (
	"strings"
	"testing"
)

func TestBuildPlan(t *testing.T) {
	req := PlanRequest{
		Prompt:       "Refactor the Go code in this repository",
		TaskType:     "coding",
		Cost:         "budget",
		IncludeRepo:  false,
		MaxRepoFiles: 5,
	}

	plan, err := BuildPlan(req)
	if err != nil {
		t.Fatalf("unexpected error building plan: %v", err)
	}

	if plan.Prompt != req.Prompt {
		t.Errorf("expected prompt %q, got %q", req.Prompt, plan.Prompt)
	}

	if len(plan.Steps) == 0 {
		t.Errorf("expected non-empty steps")
	}

	// shouldIncludeRepoMap should trigger true because "repository" is in prompt
	if !plan.RepoMapIncluded {
		t.Errorf("expected RepoMapIncluded to be true given repository keyword")
	}

	formatted := plan.FormatPlan()
	if !strings.Contains(formatted, "Task Type:") || !strings.Contains(formatted, "Steps:") {
		t.Errorf("unexpected formatted plan: %s", formatted)
	}
}
