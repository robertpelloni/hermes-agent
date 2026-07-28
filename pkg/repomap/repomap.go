package repomap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Options controls repo map generation.
type Options struct {
	BaseDir        string
	MaxFiles       int
	IncludeTests   bool
	MentionedFiles []string
	MentionedIdents []string
}

// Result contains the generated repo map and metadata.
type Result struct {
	Map      string            `json:"map"`
	Files    []RankedFile      `json:"files"`
	Stats    Stats             `json:"stats"`
	RankedIdents map[string]int `json:"rankedIdents,omitempty"`
}

// RankedFile represents a file with its relevance rank.
type RankedFile struct {
	Path   string `json:"path"`
	Rank   int    `json:"rank"`
	Lines  int    `json:"lines"`
	Reason string `json:"reason"`
}

// Stats provides statistics about the repo map generation.
type Stats struct {
	TotalFiles   int `json:"totalFiles"`
	Included     int `json:"included"`
	Skipped      int `json:"skipped"`
	TotalLines   int `json:"totalLines"`
	MentionMatch int `json:"mentionMatch"`
}

// DefaultIgnoredPatterns lists files/dirs commonly excluded from repo maps.
var DefaultIgnoredPatterns = map[string]bool{
	".git": true, "node_modules": true, ".venv": true, "venv": true,
	"__pycache__": true, ".next": true, "dist": true, "build": true,
	".pnpm": true, ".npm": true, "target": true, "vendor": true,
	".hermes": true, ".pi": true, ".jules": true, ".tormentnexus": true,
	"uv.lock": true, "pnpm-lock.yaml": true, "package-lock.json": true,
}

// Generate creates a ranked repo map for context condensation.
func Generate(opts Options) (*Result, error) {
	if opts.BaseDir == "" {
		opts.BaseDir = "."
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 40
	}

	baseDir, err := filepath.Abs(opts.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("resolving base dir: %w", err)
	}

	var files []RankedFile
	mentionSet := make(map[string]bool)
	for _, f := range opts.MentionedFiles {
		mentionSet[strings.ToLower(filepath.Base(f))] = true
	}
	identSet := make(map[string]bool)
	for _, id := range opts.MentionedIdents {
		identSet[strings.ToLower(id)] = true
	}

	stats := Stats{}
	rankedIdents := make(map[string]int)

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if DefaultIgnoredPatterns[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		stats.TotalFiles++

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			relPath = path
		}

		// Skip certain extensions by default
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".exe" || ext == ".dll" || ext == ".so" || ext == ".dylib" ||
			ext == ".jpg" || ext == ".png" || ext == ".gif" || ext == ".ico" ||
			ext == ".woff2" || ext == ".ttf" || ext == ".eot" {
			stats.Skipped++
			return nil
		}

		// Count lines
		content, err := os.ReadFile(path)
		if err != nil {
			stats.Skipped++
			return nil
		}
		lineCount := 0
		for _, b := range content {
			if b == '\n' {
				lineCount++
			}
		}
		if lineCount == 0 {
			lineCount = 1
		}

		// Calculate rank based on mentions and path depth
		rank := calculateRank(relPath, lineCount, mentionSet, identSet, content)

		files = append(files, RankedFile{
			Path:  relPath,
			Rank:  rank,
			Lines: lineCount,
			Reason: rankReason(rank, mentionSet, relPath),
		})
		return nil
	}

	filepath.Walk(baseDir, walkFn)

	// Sort by rank descending
	sort.Slice(files, func(i, j int) bool {
		if files[i].Rank != files[j].Rank {
			return files[i].Rank > files[j].Rank
		}
		return files[i].Path < files[j].Path
	})

	// Apply max files limit
	if len(files) > opts.MaxFiles {
		files = files[:opts.MaxFiles]
	}

	stats.Included = len(files)
	for _, f := range files {
		stats.TotalLines += f.Lines
		if mentionSet[strings.ToLower(filepath.Base(f.Path))] {
			stats.MentionMatch++
		}
	}

	// Build the text map
	var mapBuilder strings.Builder
	mapBuilder.WriteString("repo map:\n")
	for _, f := range files {
		reasonStr := ""
		if f.Reason != "" {
			reasonStr = fmt.Sprintf(" // %s", f.Reason)
		}
		mapBuilder.WriteString(fmt.Sprintf("  %s (%d lines)%s\n", f.Path, f.Lines, reasonStr))
	}

	return &Result{
		Map:          mapBuilder.String(),
		Files:        files,
		Stats:        stats,
		RankedIdents: rankedIdents,
	}, nil
}

func calculateRank(relPath string, lineCount int, mentionSet, identSet map[string]bool, content []byte) int {
	rank := 0

	// Prioritize mentioned files
	base := strings.ToLower(filepath.Base(relPath))
	if mentionSet[base] {
		rank += 1000
	}

	// Shallow paths get higher rank (more likely to be entry points)
	depth := len(strings.Split(relPath, string(filepath.Separator)))
	rank += max(50-depth*5, 0)

	// Root-level files (README, config, etc.)
	if depth <= 2 {
		rank += 30
	}

	// Important config files
	lower := strings.ToLower(relPath)
	if strings.Contains(lower, "readme") || strings.Contains(lower, "license") {
		rank += 20
	}
	if strings.Contains(lower, "package.json") || strings.Contains(lower, "go.mod") ||
		strings.Contains(lower, "pyproject.toml") || strings.Contains(lower, "cargo.toml") {
		rank += 25
	}

	// Main entry points
	if strings.Contains(lower, "main.") || strings.Contains(lower, "cli.") || strings.Contains(lower, "index.") {
		rank += 15
	}

	// Ident search (simple, matches exact bytes)
	if len(identSet) > 0 {
		for ident := range identSet {
			if strings.Contains(strings.ToLower(string(content)), ident) {
				rank += 50
				break
			}
		}
	}

	return rank
}

func rankReason(rank int, mentionSet map[string]bool, relPath string) string {
	switch {
	case rank >= 1000:
		return "mentioned file"
	case rank >= 50:
		return "high relevance"
	case rank >= 20:
		return "project root or config"
	case rank > 0:
		return "supporting file"
	default:
		return ""
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
