package memory

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

// Node represents a graph node with an ID and a JSON payload.
type Node struct {
    ID      string         `json:"id"`
    Payload map[string]any `json:"payload"`
}

// GraphStore provides CRUD operations on a simple directed graph stored in SQLite.
// It reuses the same SQLite file as the memory store (memory.db).
type GraphStore struct {
    db *sql.DB
}

// NewGraphStore creates (or opens) the SQLite database and ensures the
// required tables exist. It returns nil on failure.
func NewGraphStore() *GraphStore {
    // Reuse initSQLite from memory package to get the SQLite DB handle.
    // initSQLite creates the DB file under $HOME/.hermes/memory.db.
    sqlite := initSQLite()
    if sqlite == nil {
        return nil
    }
    g := &GraphStore{db: sqlite.db}
    if err := g.ensureSchema(); err != nil {
        fmt.Fprintf(os.Stderr, "[hermes:memory] graph schema error: %v\n", err)
        return nil
    }
    return g
}

// ensureSchema creates the nodes and edges tables if they don't exist.
func (g *GraphStore) ensureSchema() error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS nodes (
            id TEXT PRIMARY KEY,
            payload TEXT NOT NULL
        );`,
        `CREATE TABLE IF NOT EXISTS edges (
            src TEXT NOT NULL,
            dst TEXT NOT NULL,
            rel TEXT NOT NULL,
            PRIMARY KEY (src, dst, rel)
        );`,
        // Indexes for faster lookups.
        `CREATE INDEX IF NOT EXISTS idx_edges_src ON edges(src);`,
        `CREATE INDEX IF NOT EXISTS idx_edges_dst ON edges(dst);`,
    }
    for _, q := range stmts {
        if _, err := g.db.Exec(q); err != nil {
            return fmt.Errorf("failed to execute schema query: %w", err)
        }
    }
    return nil
}

// AddNode inserts or replaces a node with the given ID and payload.
func (g *GraphStore) AddNode(id string, payload any) error {
    if strings.TrimSpace(id) == "" {
        return fmt.Errorf("node id cannot be empty")
    }
    // Encode payload as JSON.
    data, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }
    _, err = g.db.Exec(`INSERT OR REPLACE INTO nodes (id, payload) VALUES (?,?)`, id, string(data))
    if err != nil {
        return fmt.Errorf("failed to insert node: %w", err)
    }
    return nil
}

// AddEdge creates a directed edge from src to dst with a relationship label.
func (g *GraphStore) AddEdge(src, dst, rel string) error {
    if strings.TrimSpace(src) == "" || strings.TrimSpace(dst) == "" || strings.TrimSpace(rel) == "" {
        return fmt.Errorf("src, dst and rel must be non-empty")
    }
    _, err := g.db.Exec(`INSERT OR REPLACE INTO edges (src, dst, rel) VALUES (?,?,?)`, src, dst, rel)
    if err != nil {
        return fmt.Errorf("failed to insert edge: %w", err)
    }
    return nil
}

// FindRelated returns nodes reachable from the given node ID up to the
// specified depth (number of hops). Depth 0 returns the starting node only.
func (g *GraphStore) FindRelated(id string, depth int) ([]Node, error) {
    if depth < 0 {
        return nil, fmt.Errorf("depth cannot be negative")
    }
    visited := map[string]bool{id: true}
    frontier := []string{id}
    result := []Node{}

    // Helper to load a node by ID.
    loadNode := func(nodeID string) (*Node, error) {
        var payloadStr string
        err := g.db.QueryRow(`SELECT payload FROM nodes WHERE id = ?`, nodeID).Scan(&payloadStr)
        if err != nil {
            if err == sql.ErrNoRows {
                return nil, nil
            }
            return nil, err
        }
        var payload map[string]any
        if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
            return nil, err
        }
        return &Node{ID: nodeID, Payload: payload}, nil
    }

    for d := 0; d <= depth && len(frontier) > 0; d++ {
        nextFrontier := []string{}
        for _, cur := range frontier {
            // Load node data.
            n, err := loadNode(cur)
            if err != nil {
                return nil, err
            }
            if n != nil {
                result = append(result, *n)
            }
            // Find outgoing edges.
            rows, err := g.db.Query(`SELECT dst FROM edges WHERE src = ?`, cur)
            if err != nil {
                return nil, err
            }
            for rows.Next() {
                var dst string
                if err := rows.Scan(&dst); err != nil {
                    rows.Close()
                    return nil, err
                }
                if !visited[dst] {
                    visited[dst] = true
                    nextFrontier = append(nextFrontier, dst)
                }
            }
            rows.Close()
        }
        frontier = nextFrontier
    }
    return result, nil
}

// Query executes a very small DSL supporting two patterns:
//   1. "MATCH (n) RETURN n" – returns all nodes.
//   2. "MATCH (n)-[:REL]->(m) WHERE n.id = \"ID\" RETURN m" – returns direct neighbors.
func (g *GraphStore) Query(q string) ([]Node, error) {
    q = strings.TrimSpace(q)
    // Pattern 1
    if strings.EqualFold(q, "MATCH (n) RETURN n") {
        rows, err := g.db.Query(`SELECT id, payload FROM nodes`)
        if err != nil {
            return nil, err
        }
        defer rows.Close()
        var nodes []Node
        for rows.Next() {
            var id, payloadStr string
            if err := rows.Scan(&id, &payloadStr); err != nil {
                return nil, err
            }
            var payload map[string]any
            if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
                return nil, err
            }
            nodes = append(nodes, Node{ID: id, Payload: payload})
        }
        return nodes, nil
    }
    // Pattern 2 – simple parsing for "MATCH (n)-[:REL]->(m) WHERE n.id = \"ID\" RETURN m"
    // This parser extracts the relationship label and source node ID.
    if strings.HasPrefix(strings.ToUpper(q), "MATCH") && strings.Contains(strings.ToUpper(q), "WHERE") && strings.Contains(strings.ToUpper(q), "RETURN") {
        // Extract relationship label between "[:" and "]"
        relStart := strings.Index(q, "[:")
        relEnd := strings.Index(q, "]")
        if relStart == -1 || relEnd == -1 || relEnd <= relStart+2 {
            return nil, fmt.Errorf("invalid relationship syntax")
        }
        rel := q[relStart+2 : relEnd]

        // Extract source ID after "n.id = \""
        idPrefix := "n.id = \""
        idIdx := strings.Index(q, idPrefix)
        if idIdx == -1 {
            return nil, fmt.Errorf("missing source ID in query")
        }
        // Find closing quote.
        start := idIdx + len(idPrefix)
        end := strings.Index(q[start:], "\"")
        if end == -1 {
            return nil, fmt.Errorf("unterminated source ID quote")
        }
        idVal := q[start : start+end]

        // Query edges for this src ID and relationship.
        rows, err := g.db.Query(`SELECT dst FROM edges WHERE src = ? AND rel = ?`, idVal, rel)
        if err != nil {
            return nil, err
        }
        defer rows.Close()
        var nodes []Node
        for rows.Next() {
            var dst string
            if err := rows.Scan(&dst); err != nil {
                return nil, err
            }
            // Load node data for each dst.
            var payloadStr string
            err = g.db.QueryRow(`SELECT payload FROM nodes WHERE id = ?`, dst).Scan(&payloadStr)
            if err != nil {
                if err == sql.ErrNoRows {
                    continue
                }
                return nil, err
            }
            var payload map[string]any
            if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
                return nil, err
            }
            nodes = append(nodes, Node{ID: dst, Payload: payload})
        }
        return nodes, nil
    }
    return nil, fmt.Errorf("unsupported query: %s", q)
}

// Helper to ensure HOME directory exists for tests – not used in production.