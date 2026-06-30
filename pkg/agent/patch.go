package agent

import (
	"fmt"
	"strings"
)

// ApplyDiffBlock applies a SEARCH/REPLACE diff block to a file's content
// It handles slightly inexact search boundaries by normalizing whitespace and indentation.
func ApplyDiffBlock(content string, searchBlock string, replaceBlock string) (string, error) {
	// 1. Exact match attempt
	if strings.Contains(content, searchBlock) {
		return strings.Replace(content, searchBlock, replaceBlock, 1), nil
	}

	// 2. Fallback: normalize whitespace to handle minor formatting discrepancies
	normalizedContent := normalizeWhitespace(content)
	normalizedSearch := normalizeWhitespace(searchBlock)

	if !strings.Contains(normalizedContent, normalizedSearch) {
		return "", fmt.Errorf("search block not found in file content (even after whitespace normalization)")
	}

	// If the normalized search works but exact doesn't, we need to do a line-by-line fuzzy match
	// to find the exact character boundaries in the original text, to apply the replacement.
	// For this simplified example, we'll return an error indicating an inexact match was found
	// but fuzzy replacement is not yet fully implemented.
	return "", fmt.Errorf("search block found via fuzzy match, but fuzzy replacement is not yet implemented")
}

// normalizeWhitespace compresses all sequences of whitespace into a single space
func normalizeWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inSpace := false
	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(ch)
			inSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}
