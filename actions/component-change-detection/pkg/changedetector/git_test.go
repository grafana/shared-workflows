package changedetector

import (
	"testing"
)

// Note: These are basic tests for git.go. Full integration tests with real git
// operations are in detector_integration_test.go

func TestNewGitOps(t *testing.T) {
	ops := NewGitOps()
	if ops == nil {
		t.Error("NewGitOps() returned nil")
	}

	// Verify it returns DefaultGitOps
	if _, ok := ops.(*DefaultGitOps); !ok {
		t.Error("NewGitOps() did not return *DefaultGitOps")
	}
}

func TestGitOperations_Interface(t *testing.T) {
	// Verify DefaultGitOps implements GitOperations interface
	var _ GitOperations = (*DefaultGitOps)(nil)
}

// TestGetChangedFiles_ErrorHandling tests error paths
// Note: Success paths are tested in integration tests
func TestGetChangedFiles_ErrorHandling(t *testing.T) {
	ops := &DefaultGitOps{}

	// Test with invalid refs (this should fail in any git repo)
	_, err := ops.GetChangedFiles("invalid-ref-that-does-not-exist-12345", "HEAD")
	if err == nil {
		t.Error("GetChangedFiles() with invalid ref should return error")
	}
}

// TestRefExists_ErrorHandling tests error paths
func TestRefExists_ErrorHandling(t *testing.T) {
	ops := &DefaultGitOps{}

	// Test with clearly invalid ref
	exists, err := ops.RefExists("this-is-definitely-not-a-valid-ref-12345678901234567890")
	if err != nil {
		// Some git versions might return an error instead of false
		// Both are acceptable
		return
	}

	// If no error, should return false
	if exists {
		t.Error("RefExists() should return false for invalid ref")
	}
}
