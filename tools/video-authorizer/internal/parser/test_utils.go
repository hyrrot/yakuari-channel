package parser

import (
	"os"
	"testing"
)

// createTempFile creates a temporary file for testing
func createTempFile(t *testing.T, name, content string) *os.File {
	t.Helper()
	
	tmpfile, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("failed to write to temp file: %v", err)
	}

	if err := tmpfile.Sync(); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("failed to sync temp file: %v", err)
	}

	// Reset file pointer to beginning
	if _, err := tmpfile.Seek(0, 0); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("failed to seek temp file: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		os.Remove(tmpfile.Name())
	})

	return tmpfile
}