package memory

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestGraphStoreBasicOperations(t *testing.T) {
    // Use a temporary HOME directory to isolate SQLite DB.
    tmpHome := t.TempDir()
    // Set HOME env var for initSQLite to use our temp directory.
    t.Setenv("HOME", tmpHome)

    // Ensure the .hermes directory path is created.
    hermesDir := filepath.Join(tmpHome, ".hermes")
    if err := os.MkdirAll(hermesDir, 0o755); err != nil {
        t.Fatalf("failed to create hermes dir: %v", err)
    }

    // Create a new graph store.
    gs := NewGraphStore()
    if gs == nil {
        t.Fatalf("NewGraphStore returned nil")
    }

    // Add two nodes.
    payloadA := map[string]any{"type": "module", "name": "A"}
    if err := gs.AddNode("A", payloadA); err != nil {
        t.Fatalf("AddNode A error: %v", err)
    }
    payloadB := map[string]any{"type": "module", "name": "B"}
    if err := gs.AddNode("B", payloadB); err != nil {
        t.Fatalf("AddNode B error: %v", err)
    }

    // Add an edge A -> B with relationship "depends".
    if err := gs.AddEdge("A", "B", "depends"); err != nil {
        t.Fatalf("AddEdge error: %v", err)
    }

    // Query all nodes.
    nodes, err := gs.Query("MATCH (n) RETURN n")
    if err != nil {
        t.Fatalf("Query all nodes error: %v", err)
    }
    if len(nodes) != 2 {
        t.Fatalf("expected 2 nodes, got %d", len(nodes))
    }
    // Verify node IDs present.
    foundA, foundB := false, false
    for _, n := range nodes {
        if n.ID == "A" {
            foundA = true
        }
        if n.ID == "B" {
            foundB = true
        }
    }
    if !foundA || !foundB {
        t.Fatalf("expected both A and B in node list, got %+v", nodes)
    }

    // Query outgoing edges from A using the mini DSL.
    q := "MATCH (n)-[:depends]->(m) WHERE n.id = \"A\" RETURN m"
    related, err := gs.Query(q)
    if err != nil {
        t.Fatalf("Query edge error: %v", err)
    }
    if len(related) != 1 || related[0].ID != "B" {
        t.Fatalf("expected B as related node, got %+v", related)
    }

    // Test FindRelated depth 1 returns B.
    relNodes, err := gs.FindRelated("A", 1)
    if err != nil {
        t.Fatalf("FindRelated error: %v", err)
    }
    // Should contain A and B (depth includes start node).
    if len(relNodes) != 2 {
        t.Fatalf("FindRelated depth 1 expected 2 nodes (A and B), got %d", len(relNodes))
    }
    // Ensure B is included.
    bFound := false
    for _, n := range relNodes {
        if n.ID == "B" {
            bFound = true
            break
        }
    }
    if !bFound {
        t.Fatalf("FindRelated did not include B: %+v", relNodes)
    }

    // Verify JSON payload round‑trip.
    var payloadCheck map[string]any
    raw, err := json.Marshal(payloadA)
    if err != nil {
        t.Fatalf("marshal payloadA failed: %v", err)
    }
    if err := json.Unmarshal(raw, &payloadCheck); err != nil {
        t.Fatalf("unmarshal payloadA failed: %v", err)
    }
    if payloadCheck["name"] != "A" {
        t.Fatalf("payload round‑trip mismatch: %+v", payloadCheck)
    }
}
