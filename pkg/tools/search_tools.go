package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func init() {
	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "search_content",
		Description: "Search file contents for a pattern. Returns matching file paths with line numbers and context.",
		Category:    "file",
		Parameters: map[string]any{
			"pattern":    map[string]string{"type": "string", "description": "The regex or literal pattern to search for"},
			"path":       map[string]string{"type": "string", "description": "Directory or file to search (default: current dir)"},
			"glob":       map[string]string{"type": "string", "description": "File glob filter (e.g., '*.go', '**/*.ts')"},
			"ignoreCase": map[string]string{"type": "string", "description": "Case-insensitive search ('true' or 'false')"},
			"literal":    map[string]string{"type": "string", "description": "Treat pattern as literal string ('true' or 'false')"},
			"context":    map[string]string{"type": "string", "description": "Number of context lines to show before and after each match"},
			"limit":      map[string]string{"type": "string", "description": "Maximum number of matches to return"},
		},
		Handler: searchContentHandler,
		Native:  true,
	})

	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "find_files",
		Description: "Find files by glob pattern. Supports wildcards: '*.go', '**/*.ts', 'src/**/*.py'.",
		Category:    "file",
		Parameters: map[string]any{
			"pattern": map[string]string{"type": "string", "description": "Glob pattern to search for"},
			"path":    map[string]string{"type": "string", "description": "Base directory to search (default: current dir)"},
			"limit":   map[string]string{"type": "string", "description": "Maximum number of results"},
		},
		Handler: findFilesHandler,
		Native:  true,
	})

	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "list_directory",
		Description: "List contents of a directory. Returns entries sorted with size and modification time.",
		Category:    "file",
		Parameters: map[string]any{
			"path":  map[string]string{"type": "string", "description": "Directory path to list (default: current dir)"},
			"limit": map[string]string{"type": "string", "description": "Maximum number of entries"},
		},
		Handler: listDirectoryHandler,
		Native:  true,
	})

	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "edit_file",
		Description: "Apply a targeted search-replace edit to a file. The 'old' string must match a unique region in the file.",
		Category:    "file",
		Parameters: map[string]any{
			"path":    map[string]string{"type": "string", "description": "Path to the file to edit"},
			"old":     map[string]string{"type": "string", "description": "Exact text to replace (must be unique)"},
			"new":     map[string]string{"type": "string", "description": "Replacement text"},
		},
		Handler: editFileHandler,
		Native:  true,
	})
}

// searchContentHandler searches file contents for a pattern.
func searchContentHandler(args map[string]any, _ map[string]any) (any, error) {
	pattern := getArgStr(args, "pattern")
	if pattern == "" {
		return nil, fmt.Errorf("missing 'pattern' argument")
	}

	searchDir := getArgStr(args, "path")
	if searchDir == "" {
		searchDir = "."
	}

	globFilter := getArgStr(args, "glob")
	ignoreCase := getArgStr(args, "ignoreCase") == "true"
	isLiteral := getArgStr(args, "literal") == "true"
	contextLines := getArgInt(args, "context", 0)
	limit := getArgInt(args, "limit", 100)

	var re *regexp.Regexp
	if isLiteral {
		p := regexp.QuoteMeta(pattern)
		if ignoreCase {
			p = "(?i)" + p
		}
		re = regexp.MustCompile(p)
	} else if ignoreCase {
		re = regexp.MustCompile("(?i)" + pattern)
	} else {
		re = regexp.MustCompile(pattern)
	}

	type match struct {
		Path       string `json:"path"`
		Line       int    `json:"line"`
		LineContent string `json:"lineContent"`
		ContextBefore []string `json:"contextBefore,omitempty"`
		ContextAfter  []string `json:"contextAfter,omitempty"`
	}

	results := make([]match, 0)
	matchCount := 0

	err := filepath.Walk(searchDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if globFilter != "" {
			matchGlob, _ := filepath.Match(globFilter, filepath.Base(filePath))
			if !matchGlob {
				return nil
			}
		}

		// Skip binary files
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == ".exe" || ext == ".dll" || ext == ".so" || ext == ".dylib" || ext == ".bin" || ext == ".png" || ext == ".jpg" || ext == ".ico" {
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		lines := make([]string, 0)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil
		}

		for lineNum, line := range lines {
			if re.MatchString(line) {
				relPath, _ := filepath.Rel(searchDir, filePath)

				before := make([]string, 0, contextLines)
				for i := max(0, lineNum-contextLines); i < lineNum; i++ {
					before = append(before, lines[i])
				}
				after := make([]string, 0, contextLines)
				for i := lineNum + 1; i <= min(len(lines)-1, lineNum+contextLines); i++ {
					after = append(after, lines[i])
				}

				results = append(results, match{
					Path:        relPath,
					Line:        lineNum + 1,
					LineContent: strings.TrimSpace(line),
					ContextBefore: before,
					ContextAfter:  after,
				})
				matchCount++
				if matchCount >= limit {
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}

	return map[string]any{
		"pattern": pattern,
		"matches": results,
		"count":   len(results),
		"limitReached": matchCount >= limit,
	}, nil
}

// findFilesHandler finds files by glob pattern.
func findFilesHandler(args map[string]any, _ map[string]any) (any, error) {
	pattern := getArgStr(args, "pattern")
	if pattern == "" {
		return nil, fmt.Errorf("missing 'pattern' argument")
	}

	searchDir := getArgStr(args, "path")
	if searchDir == "" {
		searchDir = "."
	}

	limit := getArgInt(args, "limit", 1000)

	files := make([]string, 0)
	count := 0

	err := filepath.Walk(searchDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err != nil {
			return nil
		}
		if !matched {
			// Also try matching the full relative path
			relPath, _ := filepath.Rel(searchDir, filePath)
			matched, _ = filepath.Match(pattern, relPath)
		}
		if matched {
			relPath, _ := filepath.Rel(searchDir, filePath)
			files = append(files, relPath)
			count++
			if count >= limit {
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("find error: %w", err)
	}

	return map[string]any{
		"pattern": pattern,
		"files":   files,
		"count":   len(files),
		"limitReached": count >= limit,
	}, nil
}

// listDirectoryHandler lists directory contents.
func listDirectoryHandler(args map[string]any, _ map[string]any) (any, error) {
	dirPath := getArgStr(args, "path")
	if dirPath == "" {
		dirPath = "."
	}

	limit := getArgInt(args, "limit", 200)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	type entry struct {
		Name    string `json:"name"`
		IsDir   bool   `json:"isDir"`
		Size    int64  `json:"size,omitempty"`
		ModTime string `json:"modTime,omitempty"`
	}

	result := make([]entry, 0, min(len(entries), limit))
	for i, e := range entries {
		if i >= limit {
			break
		}
		info, err := e.Info()
		ent := entry{Name: e.Name(), IsDir: e.IsDir()}
		if err == nil {
			ent.Size = info.Size()
			ent.ModTime = info.ModTime().Format("Jan _2 15:04")
		}
		result = append(result, ent)
	}

	return map[string]any{
		"directory": dirPath,
		"entries":   result,
		"count":     len(result),
		"total":     len(entries),
		"truncated": len(entries) > limit,
	}, nil
}

// editFileHandler applies a search-replace edit to a file.
func editFileHandler(args map[string]any, _ map[string]any) (any, error) {
	filePath := getArgStr(args, "path")
	if filePath == "" {
		return nil, fmt.Errorf("missing 'path' argument")
	}

	oldText := getArgStr(args, "old")
	if oldText == "" {
		return nil, fmt.Errorf("missing 'old' argument")
	}

	newText := getArgStr(args, "new")

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	content := string(data)

	// Count occurrences
	count := strings.Count(content, oldText)
	if count == 0 {
		return nil, fmt.Errorf("text not found in %s", filePath)
	}
	if count > 1 {
		return nil, fmt.Errorf("text found %d times in %s (must be unique)", count, filePath)
	}

	// Apply replacement
	newContent := strings.Replace(content, oldText, newText, 1)

	// Write back
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return map[string]any{
		"path":    filePath,
		"oldSize": len(content),
		"newSize": len(newContent),
		"diff":    fmt.Sprintf("%d bytes replaced", len(oldText)),
	}, nil
}

// Helper functions

func getArgStr(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getArgInt(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case string:
			var n int
			if _, err := fmt.Sscanf(val, "%d", &n); err == nil {
				return n
			}
		}
	}
	return def
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
