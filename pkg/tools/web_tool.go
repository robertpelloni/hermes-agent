package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func init() {
	// Register web_search
	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "web_search",
		Description: "Search the web for information. Returns a list of results with URLs and snippets.",
		Category:    "web",
		Parameters: map[string]any{
			"query": map[string]string{"type": "string", "description": "The search query"},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results",
				"default":     5,
			},
		},
		Handler: webSearchHandler,
		Native:  true,
	})

	// Register web_extract
	toolregistry.Global().Register(&toolregistry.Tool{
		Name:        "web_extract",
		Description: "Extract text content from a web page.",
		Category:    "web",
		Parameters: map[string]any{
			"url": map[string]string{"type": "string", "description": "The URL to extract content from"},
		},
		Handler: webExtractHandler,
		Native:  true,
	})
}

type WebSearchResult struct {
	Query   string              `json:"query"`
	Results []WebSearchItem     `json:"results"`
}

type WebSearchItem struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

func webSearchHandler(args map[string]any, _ map[string]any) (any, error) {
	qRaw, ok := args["query"]
	if !ok {
		return nil, fmt.Errorf("missing 'query' argument")
	}
	query, ok := qRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'query' must be a string")
	}

	maxResults := 5
	if mRaw, ok := args["max_results"]; ok {
		if mFloat, ok := mRaw.(float64); ok {
			maxResults = int(mFloat)
		}
	}

	// Use DuckDuckGo's instant answer API (no API key required)
	// Fallback: JSON endpoint
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("[web_search] HTTP error: %v", err)
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	result := WebSearchResult{
		Query:   query,
		Results: make([]WebSearchItem, 0),
	}

	// Try to parse as JSON
	var ddgResponse struct {
		AbstractURL   string `json:"AbstractURL"`
		AbstractText  string `json:"AbstractText"`
		AbstractSource string `json:"AbstractSource"`
		Heading       string `json:"Heading"`
		RelatedTopics []struct {
			Result string `json:"Result"`
			Text   string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &ddgResponse); err == nil {
		// Abstract result is the main answer
		if ddgResponse.AbstractURL != "" && ddgResponse.AbstractText != "" {
			result.Results = append(result.Results, WebSearchItem{
				URL:     ddgResponse.AbstractURL,
				Title:   ddgResponse.Heading,
				Snippet: ddgResponse.AbstractText,
			})
		}

		// Related topics provide search results
		for _, topic := range ddgResponse.RelatedTopics {
			if len(result.Results) >= maxResults {
				break
			}
			if topic.Text != "" {
				result.Results = append(result.Results, WebSearchItem{
					URL:     topic.FirstURL,
					Title:   topic.Text,
					Snippet: topic.Text,
				})
			}
		}
	}

	return result, nil
}

func webExtractHandler(args map[string]any, _ map[string]any) (any, error) {
	urlRaw, ok := args["url"]
	if !ok {
		return nil, fmt.Errorf("missing 'url' argument")
	}
	pageURL, ok := urlRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'url' must be a string")
	}

	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return map[string]any{
		"url":   pageURL,
		"body":  string(body),
		"bytes": len(body),
	}, nil
}
