package assimilation

import (
	"strings"
	"testing"
)

func TestAssimilationInventory(t *testing.T) {
	items := Inventory()
	if len(items) == 0 {
		t.Fatalf("expected non-empty assimilation inventory")
	}

	foundHyperCode := false
	foundPiMono := false
	for _, item := range items {
		if item.ID == "hypercode" {
			foundHyperCode = true
		}
		if item.ID == "pi-mono" {
			foundPiMono = true
		}
	}

	if !foundHyperCode {
		t.Errorf("expected hypercode in assimilation inventory")
	}
	if !foundPiMono {
		t.Errorf("expected pi-mono in assimilation inventory")
	}

	formatted := FormatInventory()
	if !strings.Contains(formatted, "Assimilation Inventory") || !strings.Contains(formatted, "HyperCode") {
		t.Errorf("unexpected formatted inventory: %s", formatted)
	}
}
