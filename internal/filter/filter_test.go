package filter

import (
	"testing"
	"time"
)

func TestFilter_WithMatch(t *testing.T) {
	// Create input channel
	input := make(chan string)

	// Start filter
	output := Filter(input, "ERROR")

	// Send messages
	go func() {
		input <- "[app.log] - This is an ERROR message"
		input <- "[app.log] - This is an INFO message"
		input <- "[app.log] - Another ERROR occurred"
		input <- "[app.log] - DEBUG information"
		close(input)
	}()

	// Collect filtered messages
	var messages []string
	for msg := range output {
		messages = append(messages, msg)
	}

	// Should only receive ERROR messages
	if len(messages) != 2 {
		t.Errorf("expected 2 messages with ERROR, got %d", len(messages))
	}

	for _, msg := range messages {
		if !contains(msg, "ERROR") {
			t.Errorf("expected message to contain 'ERROR', got %q", msg)
		}
	}
}

func TestFilter_NoMatch(t *testing.T) {
	input := make(chan string)

	output := Filter(input, "CRITICAL")

	go func() {
		input <- "[app.log] - This is an ERROR message"
		input <- "[app.log] - This is an INFO message"
		input <- "[app.log] - DEBUG information"
		close(input)
	}()

	var messages []string
	timeout := time.After(2 * time.Second)
	done := make(chan bool)

	go func() {
		for msg := range output {
			messages = append(messages, msg)
		}
		done <- true
	}()

	select {
	case <-done:
		if len(messages) != 0 {
			t.Errorf("expected 0 messages with CRITICAL, got %d", len(messages))
		}
	case <-timeout:
		t.Error("filter did not close output channel")
	}
}

func TestFilter_EmptyFilter(t *testing.T) {
	input := make(chan string)

	// Empty filter should match all strings (as all strings contain empty string)
	output := Filter(input, "")

	go func() {
		input <- "[app.log] - Message 1"
		input <- "[app.log] - Message 2"
		input <- "[app.log] - Message 3"
		close(input)
	}()

	var messages []string
	for msg := range output {
		messages = append(messages, msg)
	}

	// Empty string is contained in all strings, so all should pass
	if len(messages) != 3 {
		t.Errorf("expected 3 messages with empty filter, got %d", len(messages))
	}
}

func TestFilter_EmptyInput(t *testing.T) {
	input := make(chan string)
	close(input)

	output := Filter(input, "ERROR")

	var messages []string
	for msg := range output {
		messages = append(messages, msg)
	}

	if len(messages) != 0 {
		t.Errorf("expected 0 messages from empty input, got %d", len(messages))
	}
}

func TestFilter_CaseSensitive(t *testing.T) {
	input := make(chan string)

	output := Filter(input, "error")

	go func() {
		input <- "[app.log] - This is an ERROR message"
		input <- "[app.log] - This is an error message"
		input <- "[app.log] - Error occurred"
		close(input)
	}()

	var messages []string
	for msg := range output {
		messages = append(messages, msg)
	}

	// Should only match lowercase "error"
	if len(messages) != 1 {
		t.Errorf("expected 1 message with lowercase 'error', got %d", len(messages))
	}

	if len(messages) > 0 && !contains(messages[0], "error") {
		t.Errorf("expected message to contain lowercase 'error', got %q", messages[0])
	}
}

func TestFilter_PartialMatch(t *testing.T) {
	input := make(chan string)

	output := Filter(input, "ERR")

	go func() {
		input <- "[app.log] - ERROR message"
		input <- "[app.log] - WARNING message"
		input <- "[app.log] - ERRNO 404"
		close(input)
	}()

	var messages []string
	for msg := range output {
		messages = append(messages, msg)
	}

	// Should match both ERROR and ERRNO (both contain "ERR")
	if len(messages) != 2 {
		t.Errorf("expected 2 messages containing 'ERR', got %d", len(messages))
	}
}

func TestFilter_SpecialCharacters(t *testing.T) {
	input := make(chan string)

	output := Filter(input, "[ERROR]")

	go func() {
		input <- "[app.log] - [ERROR] Something went wrong"
		input <- "[app.log] - ERROR: issue detected"
		input <- "[app.log] - [ERROR] Another problem"
		close(input)
	}()

	var messages []string
	for msg := range output {
		messages = append(messages, msg)
	}

	// Should only match exact pattern "[ERROR]"
	if len(messages) != 2 {
		t.Errorf("expected 2 messages with '[ERROR]', got %d", len(messages))
	}

	for _, msg := range messages {
		if !contains(msg, "[ERROR]") {
			t.Errorf("expected message to contain '[ERROR]', got %q", msg)
		}
	}
}

func TestFilter_MultipleOccurrences(t *testing.T) {
	input := make(chan string)

	output := Filter(input, "log")

	go func() {
		input <- "[app.log] - Logging information to log file"
		input <- "[system.log] - System message"
		input <- "[debug.txt] - Debug info"
		close(input)
	}()

	var messages []string
	for msg := range output {
		messages = append(messages, msg)
	}

	// First two messages contain "log"
	if len(messages) != 2 {
		t.Errorf("expected 2 messages containing 'log', got %d", len(messages))
	}
}

func TestFilter_LargeVolume(t *testing.T) {
	input := make(chan string, 1000)

	output := Filter(input, "match")

	// Send 1000 messages, 100 of which contain "match"
	go func() {
		for i := 0; i < 1000; i++ {
			if i%10 == 0 {
				input <- "message with match"
			} else {
				input <- "message without"
			}
		}
		close(input)
	}()

	count := 0
	timeout := time.After(5 * time.Second)
	done := make(chan bool)

	go func() {
		for range output {
			count++
		}
		done <- true
	}()

	select {
	case <-done:
		if count != 100 {
			t.Errorf("expected 100 matching messages, got %d", count)
		}
	case <-timeout:
		t.Error("filter test timed out")
	}
}

func TestFilter_ChannelClosure(t *testing.T) {
	input := make(chan string)

	output := Filter(input, "test")

	go func() {
		input <- "test message 1"
		input <- "test message 2"
		close(input)
	}()

	// Ensure output channel closes after input closes
	timeout := time.After(2 * time.Second)
	done := make(chan bool)

	go func() {
		count := 0
		for range output {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 messages, got %d", count)
		}
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-timeout:
		t.Error("output channel did not close after input closed")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
