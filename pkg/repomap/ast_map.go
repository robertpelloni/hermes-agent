package repomap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// GenerateASTMap parses files to create an AST-based context map
func GenerateASTMap(baseDir string, files []string) (string, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	var builder strings.Builder
	builder.WriteString("AST-Based Repository Map:\n")

	for _, relPath := range files {
		fullPath := filepath.Join(baseDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue // Skip unreadable
		}

		// Parse the file
		tree := parser.Parse(nil, content)
		root := tree.RootNode()

		builder.WriteString(fmt.Sprintf("\n## %s\n", relPath))

		// Extract structural elements based on the tree (simplified)
		extractElements(root, content, &builder, 0)
	}

	return builder.String(), nil
}

func extractElements(node *sitter.Node, content []byte, builder *strings.Builder, depth int) {
	if node == nil {
		return
	}

	// Simple heuristic: extract names of functions, structs, interfaces, etc.
	typ := node.Type()
	if typ == "function_declaration" || typ == "method_declaration" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			name := nameNode.Content(content)
			indent := strings.Repeat("  ", depth)
			builder.WriteString(fmt.Sprintf("%s- func %s\n", indent, name))
		}
	} else if typ == "type_spec" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			name := nameNode.Content(content)
			indent := strings.Repeat("  ", depth)
			builder.WriteString(fmt.Sprintf("%s- type %s\n", indent, name))
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		extractElements(node.Child(i), content, builder, depth+1)
	}
}
