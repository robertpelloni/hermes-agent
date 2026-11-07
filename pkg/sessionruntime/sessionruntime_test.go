package sessionruntime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
)

func TestSessionRuntime_Lifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hermes-session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rt := NewRuntime(tempDir, true)

	// Create a new session
	sess, err := rt.Create("sess-1")
	if err != nil {
		t.Fatalf("unexpected error creating session: %v", err)
	}
	if sess.ID != "sess-1" {
		t.Errorf("expected session id 'sess-1', got %s", sess.ID)
	}

	// Cannot create duplicate
	if _, err := rt.Create("sess-1"); err == nil {
		t.Errorf("expected error creating duplicate session, got nil")
	}

	// Append messages
	msg1 := ai.NewTextMessage(ai.RoleUser, "Hello world")
	if err := rt.AppendMessage("sess-1", msg1); err != nil {
		t.Fatalf("failed to append message: %v", err)
	}

	msg2 := ai.NewTextMessage(ai.RoleAssistant, "Hello! How can I help you today?")
	if err := rt.AppendMessage("sess-1", msg2); err != nil {
		t.Fatalf("failed to append message: %v", err)
	}

	got, ok := rt.Get("sess-1")
	if !ok {
		t.Fatalf("failed to get session 'sess-1'")
	}
	if len(got.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(got.Messages))
	}

	// Verify persistence file exists
	sessionFile := filepath.Join(tempDir, "sess-1.json")
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", sessionFile)
	}

	// Test Search
	searchRes := rt.Search("help you", 10)
	if len(searchRes) != 1 {
		t.Errorf("expected 1 search result, got %d", len(searchRes))
	} else if searchRes[0].SessionID != "sess-1" {
		t.Errorf("expected search result session 'sess-1', got %s", searchRes[0].SessionID)
	}

	// Test Fork
	forked, err := rt.Fork("sess-1", "sess-1-fork", 1)
	if err != nil {
		t.Fatalf("failed to fork session: %v", err)
	}
	if len(forked.Messages) != 1 {
		t.Errorf("expected forked session to have 1 message, got %d", len(forked.Messages))
	}

	// Test List
	all := rt.List()
	if len(all) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(all))
	}

	// Test Delete
	if err := rt.Delete("sess-1-fork"); err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}
	if _, ok := rt.Get("sess-1-fork"); ok {
		t.Errorf("expected sess-1-fork to be deleted")
	}
}

func TestSessionRuntime_Truncate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hermes-session-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rt := NewRuntime(tempDir, false)
	_, _ = rt.Create("sess-truncate")

	_ = rt.AppendMessage("sess-truncate", ai.NewTextMessage("system", "sys prompt"))
	_ = rt.AppendMessage("sess-truncate", ai.NewTextMessage(ai.RoleUser, "msg 1"))
	_ = rt.AppendMessage("sess-truncate", ai.NewTextMessage(ai.RoleAssistant, "msg 2"))
	_ = rt.AppendMessage("sess-truncate", ai.NewTextMessage(ai.RoleUser, "msg 3"))
	_ = rt.AppendMessage("sess-truncate", ai.NewTextMessage(ai.RoleAssistant, "msg 4"))

	// Truncate to keep 2 messages (+ system)
	if err := rt.Truncate("sess-truncate", 2); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	sess, _ := rt.Get("sess-truncate")
	if len(sess.Messages) != 3 {
		t.Errorf("expected 3 messages after truncation, got %d", len(sess.Messages))
	}
	if sess.Messages[0].Content[0].Text != "sys prompt" {
		t.Errorf("expected first message to remain system prompt, got %s", sess.Messages[0].Content[0].Text)
	}
}
