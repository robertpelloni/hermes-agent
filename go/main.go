package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/robertpelloni/hermes-agent/internal/shadowpilot"
	"github.com/robertpelloni/hermes-agent/internal/agent"
)

func streamChatResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	streamReader := strings.NewReader("Hello\nfrom\nthe\nzero-allocation\npipeline!\n")
	scanner := bufio.NewScanner(streamReader)

	for scanner.Scan() {
		line := scanner.Bytes()
		w.Write([]byte("data: "))
		w.Write(line)
		w.Write([]byte("\n\n"))
		flusher.Flush()
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "chat" {
		// Launch the native Go Agent Loop
		agent.StartRepl()
		return
	}

	http.HandleFunc("/system/status", func(w http.ResponseWriter, r *http.Request) {
		status, err := shadowpilot.GetSubmoduleStatus(".")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, "Submodule Status:\n")
		io.WriteString(w, status)
	})

	http.HandleFunc("/api/chat", streamChatResponse)

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
