package repomap

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	nonGoFuncRe = regexp.MustCompile(`^\s*(?:pub\s+)?(?:async\s+)?(?:fn|function|def)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	nonGoTypeRe = regexp.MustCompile(`^\s*(?:pub\s+)?(?:export\s+)?(?:struct|class|interface|enum|trait|type)\s+([A-Za-z_][A-Za-z0-9_]*)`)
)

// GenerateASTMap parses files to create an AST-based context map without requiring CGO.
// Go source files use standard library go/parser + go/ast; non-Go source files use structural declaration extraction.
func GenerateASTMap(baseDir string, files []string) (string, error) {
	fset := token.NewFileSet()
	var builder strings.Builder
	builder.WriteString("AST-Based Repository Map:\n")

	for _, relPath := range files {
		fullPath := filepath.Join(baseDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue // Skip unreadable
		}

		builder.WriteString(fmt.Sprintf("\n## %s\n", relPath))

		if strings.HasSuffix(strings.ToLower(relPath), ".go") {
			f, err := parser.ParseFile(fset, fullPath, content, 0)
			if err == nil && f != nil {
				extractGoDecls(f, &builder)
				continue
			}
		}

		extractPolyglotDecls(string(content), &builder)
	}

	return builder.String(), nil
}

func extractGoDecls(f *ast.File, builder *strings.Builder) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name != nil {
				builder.WriteString(fmt.Sprintf("  - func %s\n", d.Name.Name))
			}
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name != nil {
						builder.WriteString(fmt.Sprintf("  - type %s\n", ts.Name.Name))
					}
				}
			}
		}
	}
}

func extractPolyglotDecls(src string, builder *strings.Builder) {
	for _, line := range strings.Split(src, "\n") {
		if m := nonGoFuncRe.FindStringSubmatch(line); len(m) > 1 {
			builder.WriteString(fmt.Sprintf("  - func %s\n", m[1]))
		} else if m := nonGoTypeRe.FindStringSubmatch(line); len(m) > 1 {
			builder.WriteString(fmt.Sprintf("  - type %s\n", m[1]))
		}
	}
}
