package agent

import (
	"fmt"
	"os"
	"sync"
)

// VirtualFileSystem implements an in-memory change buffer.
type VirtualFileSystem struct {
	mu     sync.RWMutex
	files  map[string][]byte
}

func NewVirtualFileSystem() *VirtualFileSystem {
	return &VirtualFileSystem{
		files: make(map[string][]byte),
	}
}

// WriteFile writes a file to the VFS.
func (vfs *VirtualFileSystem) WriteFile(path string, content []byte) {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()
	vfs.files[path] = content
}

// ReadFile reads a file, falling back to the physical disk if not in VFS.
func (vfs *VirtualFileSystem) ReadFile(path string) ([]byte, error) {
	vfs.mu.RLock()
	content, ok := vfs.files[path]
	vfs.mu.RUnlock()

	if ok {
		return content, nil
	}

	return os.ReadFile(path)
}

// Flush writes all modified files to the physical disk.
func (vfs *VirtualFileSystem) Flush() error {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	for path, content := range vfs.files {
		if err := os.WriteFile(path, content, 0644); err != nil {
			return fmt.Errorf("failed to flush %s: %w", path, err)
		}
		fmt.Printf("[hermes:vfs] Flushed: %s\n", path)
	}

	// Clear the buffer after flushing
	vfs.files = make(map[string][]byte)

	return nil
}
