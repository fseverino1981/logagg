package reader

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadLines_MultipleLines(t *testing.T) {
	// Create a temporary file with multiple lines
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "line 1\nline 2\nline 3\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := ReadLines(ctx, tmpFile.Name(), false)

	var lines []string
	for line := range ch {
		lines = append(lines, line)
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	expectedLines := []string{"line 1", "line 2", "line 3"}
	for i, expected := range expectedLines {
		if i >= len(lines) {
			t.Errorf("missing line %d", i)
			continue
		}
		if !strings.Contains(lines[i], expected) {
			t.Errorf("line %d: expected to contain %q, got %q", i, expected, lines[i])
		}
	}
}

func TestReadLines_CorrectPrefix(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "myapp-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "test line\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := ReadLines(ctx, tmpFile.Name(), false)

	line := <-ch
	expectedPrefix := "[" + filepath.Base(tmpFile.Name()) + "]"

	if !strings.HasPrefix(line, expectedPrefix) {
		t.Errorf("expected line to start with %q, got %q", expectedPrefix, line)
	}

	if !strings.Contains(line, "test line") {
		t.Errorf("expected line to contain 'test line', got %q", line)
	}
}

func TestReadLines_ContextCancellation(t *testing.T) {
	// Create a temporary file with content
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	ctx, cancel := context.WithCancel(context.Background())

	ch := ReadLines(ctx, tmpFile.Name(), false)

	// Read first line
	<-ch

	// Cancel context
	cancel()

	// Channel should close soon after cancellation
	timeout := time.After(2 * time.Second)
	select {
	case _, ok := <-ch:
		if ok {
			// Still receiving, keep draining
			for range ch {
			}
		}
	case <-timeout:
		t.Error("channel did not close after context cancellation")
	}
}

func TestReadLines_EmptyFile(t *testing.T) {
	// Create an empty temporary file
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := ReadLines(ctx, tmpFile.Name(), false)

	var lines []string
	for line := range ch {
		lines = append(lines, line)
	}

	if len(lines) != 0 {
		t.Errorf("expected 0 lines from empty file, got %d", len(lines))
	}
}

func TestReadLines_TailMode(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write initial content
	if _, err := tmpFile.WriteString("initial line\n"); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := ReadLines(ctx, tmpFile.Name(), true)

	// Read initial line
	line := <-ch
	if !strings.Contains(line, "initial line") {
		t.Errorf("expected 'initial line', got %q", line)
	}

	// In tail mode, the goroutine should still be running
	// We'll cancel to stop it
	cancel()

	// Drain the channel
	timeout := time.After(2 * time.Second)
	done := make(chan bool)
	go func() {
		for range ch {
		}
		done <- true
	}()

	select {
	case <-done:
		// Channel closed as expected
	case <-timeout:
		t.Error("channel did not close in tail mode after context cancellation")
	}
}

func TestReadLines_SingleLine(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "single line without newline"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := ReadLines(ctx, tmpFile.Name(), false)

	var lines []string
	for line := range ch {
		lines = append(lines, line)
	}

	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}

	if len(lines) > 0 && !strings.Contains(lines[0], content) {
		t.Errorf("expected line to contain %q, got %q", content, lines[0])
	}
}
