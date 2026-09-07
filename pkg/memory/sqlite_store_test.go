package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteStoreInitialization(t *testing.T) {
	// Override the home directory for the test so we don't mess with the real one
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	store := initSQLite()
	if store == nil {
		t.Fatalf("Expected sqliteDB to be initialized, got nil")
	}
	defer store.close()

	// Verify that the file was created
	expectedPath := filepath.Join(tmpDir, ".hermes", "sessions.db")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected database file to exist at %s", expectedPath)
	}
}
