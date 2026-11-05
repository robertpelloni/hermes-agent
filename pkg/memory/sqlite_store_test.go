package memory

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteDBOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hermes_mem_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test_memory.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	s := &sqliteDB{db: db}
	if err := s.ensureSchema(); err != nil {
		t.Fatalf("ensureSchema failed: %v", err)
	}

	// 1. Save messages
	if err := s.saveMessage("sess-1", "user", "hello world"); err != nil {
		t.Fatalf("saveMessage failed: %v", err)
	}
	if err := s.saveMessage("sess-1", "assistant", "hi there"); err != nil {
		t.Fatalf("saveMessage failed: %v", err)
	}
	if err := s.saveMessage("sess-2", "user", "another session"); err != nil {
		t.Fatalf("saveMessage failed: %v", err)
	}

	// 2. Load messages
	msgs, err := s.loadMessages("sess-1")
	if err != nil {
		t.Fatalf("loadMessages failed: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "hello world" {
		t.Errorf("expected first message to be 'hello world', got %s", msgs[0].Content)
	}

	// 3. Count messages
	cnt, err := s.countMessages("sess-1")
	if err != nil {
		t.Fatalf("countMessages failed: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("expected count 2, got %d", cnt)
	}

	// 4. List sessions
	sessions, err := s.listSessions()
	if err != nil {
		t.Fatalf("listSessions failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	// 5. Delete session
	if err := s.deleteSession("sess-1"); err != nil {
		t.Fatalf("deleteSession failed: %v", err)
	}
	cntAfter, err := s.countMessages("sess-1")
	if err != nil {
		t.Fatalf("countMessages failed: %v", err)
	}
	if cntAfter != 0 {
		t.Fatalf("expected count 0 after deletion, got %d", cntAfter)
	}
}
