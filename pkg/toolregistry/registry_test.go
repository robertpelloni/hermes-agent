package toolregistry

import (
	"errors"
	"testing"
)

func TestToolRegistry_Lifecycle(t *testing.T) {
	reg := New()
	if reg.Count() != 0 {
		t.Fatalf("expected empty registry, got %d tools", reg.Count())
	}

	// Register invalid tools
	if err := reg.Register(nil); err == nil {
		t.Errorf("expected error registering nil tool")
	}
	if err := reg.Register(&Tool{Name: ""}); err == nil {
		t.Errorf("expected error registering tool with empty name")
	}
	if err := reg.Register(&Tool{Name: "no-handler"}); err == nil {
		t.Errorf("expected error registering tool without handler")
	}

	// Register valid tool
	toolA := &Tool{
		Name:        "tool_a",
		Description: "Tool A",
		Handler: func(args, ctx map[string]any) (any, error) {
			if _, ok := args["fail"]; ok {
				return nil, errors.New("forced failure")
			}
			return "result_a", nil
		},
	}
	if err := reg.Register(toolA); err != nil {
		t.Fatalf("unexpected error registering tool: %v", err)
	}

	if reg.Count() != 1 {
		t.Errorf("expected count 1, got %d", reg.Count())
	}

	// Get tool
	got, ok := reg.Get("tool_a")
	if !ok || got.Description != "Tool A" {
		t.Errorf("expected to get tool_a")
	}

	// Call tool success
	res, err := reg.Call("tool_a", map[string]any{}, nil)
	if err != nil || res != "result_a" {
		t.Errorf("unexpected call result: %v (err: %v)", res, err)
	}

	// Call tool failure
	_, errFail := reg.Call("tool_a", map[string]any{"fail": true}, nil)
	if errFail == nil {
		t.Errorf("expected call error when fail arg is set")
	}

	// Call non-existent tool
	_, errMissing := reg.Call("non_existent", nil, nil)
	if errMissing == nil {
		t.Errorf("expected error calling missing tool")
	}

	// List tools
	tools := reg.List()
	if len(tools) != 1 || tools[0].Name != "tool_a" {
		t.Errorf("unexpected list output: %v", tools)
	}

	// Remove tool
	reg.Remove("tool_a")
	if reg.Count() != 0 {
		t.Errorf("expected count 0 after removal, got %d", reg.Count())
	}
}

func TestGlobalRegistrySingleton(t *testing.T) {
	g1 := Global()
	g2 := Global()
	if g1 != g2 {
		t.Errorf("expected Global() to return singleton instance")
	}
}
