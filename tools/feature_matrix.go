// tools/feature_matrix.go - Generates markdown feature matrices for each CLI in submodules/
// Usage: go run tools/feature_matrix.go <submodules_dir> <output_dir>
package main

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run tools/feature_matrix.go <submodules_dir> <output_dir>")
		os.Exit(1)
	}
	submodsDir := os.Args[1]
	outputDir := os.Args[2]

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	// Iterate over submodules directory
	err := fs.WalkDir(os.DirFS(submodsDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == "." {
			return nil
		}
		// Expect one executable per top-level dir
		binaryName := strings.TrimSuffix(d.Name(), ".exe")
		binPath := filepath.Join(submodsDir, d.Name(), binaryName)
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			// Binary not found, skip
			return nil
		}

		// Write markdown file
		outPath := filepath.Join(outputDir, d.Name()+".md")
		template := fmt.Sprintf(`# %s CLI - Feature Matrix

## Overview
- Placeholder: replace with actual description
- Command: %s

## Commands
| Command | Description |
|---------|-------------|
| (fill in) | 

## Authentication
- (fill in)

## Model Support
- (fill in)

## External Dependencies
- (fill in)

## Re‑implementation Targets
- (fill in)
`, capitalize(binaryName), binaryName)

		if err := os.WriteFile(outPath, []byte(template), 0o644); err != nil {
			return err
		}
		fmt.Printf("Generated %s\n", outPath)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk error: %v\n", err)
		os.Exit(1)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}