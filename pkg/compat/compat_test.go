package compat

import (
	"strings"
	"testing"
)

func TestDefaultCatalog(t *testing.T) {
	cat := DefaultCatalog()
	if cat.Count() == 0 {
		t.Fatalf("expected non-empty tool catalog")
	}

	sources := cat.Sources()
	if len(sources) == 0 {
		t.Fatalf("expected non-empty sources list")
	}

	// Lookup tools for hermes
	hermesTools := cat.ContractsBySource("hermes")
	if len(hermesTools) == 0 {
		t.Errorf("expected hermes tools to be registered")
	}

	// Lookup by name
	readTools := cat.Lookup("read_file")
	if len(readTools) == 0 {
		t.Errorf("expected read_file to be present in catalog")
	}

	// Lookup equivalent across other sources
	equivalents := cat.LookupEquivalent("hermes", "read_file")
	if len(equivalents) == 0 {
		t.Errorf("expected cross-source equivalents for read_file")
	}

	summary := cat.Summary()
	if !strings.Contains(summary, "Tool Parity Catalog") || !strings.Contains(summary, "hermes") {
		t.Errorf("unexpected catalog summary: %s", summary)
	}
}
