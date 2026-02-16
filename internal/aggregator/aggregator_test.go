package aggregator

import (
	"context"
	"sort"
	"testing"
	"time"
)

func TestAggregate_MultipleChannels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create multiple input channels
	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)

	// Send data to channels
	go func() {
		ch1 <- "msg1-ch1"
		ch1 <- "msg2-ch1"
		close(ch1)
	}()

	go func() {
		ch2 <- "msg1-ch2"
		ch2 <- "msg2-ch2"
		close(ch2)
	}()

	go func() {
		ch3 <- "msg1-ch3"
		close(ch3)
	}()

	// Aggregate channels
	result := Aggregate(ctx, ch1, ch2, ch3)

	// Collect all messages
	var messages []string
	for msg := range result {
		messages = append(messages, msg)
	}

	// Verify we received all messages
	if len(messages) != 5 {
		t.Errorf("expected 5 messages, got %d: %v", len(messages), messages)
	}

	// Sort to verify all expected messages are present
	sort.Strings(messages)
	expected := []string{"msg1-ch1", "msg1-ch2", "msg1-ch3", "msg2-ch1", "msg2-ch2"}
	sort.Strings(expected)

	for i, exp := range expected {
		if i >= len(messages) {
			t.Errorf("missing message: %s", exp)
			continue
		}
		if messages[i] != exp {
			t.Errorf("expected message %s, got %s", exp, messages[i])
		}
	}
}

func TestAggregate_EmptyChannels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create empty channels
	ch1 := make(chan string)
	ch2 := make(chan string)

	close(ch1)
	close(ch2)

	// Aggregate channels
	result := Aggregate(ctx, ch1, ch2)

	// Collect all messages
	var messages []string
	for msg := range result {
		messages = append(messages, msg)
	}

	// Verify no messages received
	if len(messages) != 0 {
		t.Errorf("expected 0 messages from empty channels, got %d", len(messages))
	}
}

func TestAggregate_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create channels that will send many messages
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		for i := 0; i < 100; i++ {
			select {
			case ch1 <- "msg-ch1":
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
		close(ch1)
	}()

	go func() {
		for i := 0; i < 100; i++ {
			select {
			case ch2 <- "msg-ch2":
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
		close(ch2)
	}()

	// Aggregate channels
	result := Aggregate(ctx, ch1, ch2)

	// Read a few messages
	<-result
	<-result

	// Cancel context
	cancel()

	// Channel should close or stop sending
	timeout := time.After(2 * time.Second)
	done := make(chan bool)
	go func() {
		for range result {
			// Drain remaining messages
		}
		done <- true
	}()

	select {
	case <-done:
		// Successfully completed
	case <-timeout:
		t.Error("aggregate did not stop after context cancellation")
	}
}

func TestAggregate_SingleChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := make(chan string)

	go func() {
		ch <- "msg1"
		ch <- "msg2"
		ch <- "msg3"
		close(ch)
	}()

	result := Aggregate(ctx, ch)

	var messages []string
	for msg := range result {
		messages = append(messages, msg)
	}

	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}
}

func TestAggregate_NoChannels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Call Aggregate with no channels
	result := Aggregate(ctx)

	// Should close immediately
	timeout := time.After(1 * time.Second)
	select {
	case _, ok := <-result:
		if ok {
			t.Error("expected channel to be closed, but it's still open")
		}
	case <-timeout:
		t.Error("channel did not close when no input channels provided")
	}
}

func TestAggregate_ChannelOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create channels with specific timing
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		ch1 <- "first"
		time.Sleep(10 * time.Millisecond)
		ch1 <- "third"
		close(ch1)
	}()

	go func() {
		time.Sleep(5 * time.Millisecond)
		ch2 <- "second"
		close(ch2)
	}()

	result := Aggregate(ctx, ch1, ch2)

	var messages []string
	for msg := range result {
		messages = append(messages, msg)
	}

	// We should receive all messages (order may vary due to concurrency)
	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}

	// Verify all expected messages are present
	expectedMap := map[string]bool{
		"first":  false,
		"second": false,
		"third":  false,
	}

	for _, msg := range messages {
		if _, exists := expectedMap[msg]; exists {
			expectedMap[msg] = true
		}
	}

	for msg, found := range expectedMap {
		if !found {
			t.Errorf("expected to receive message %q", msg)
		}
	}
}

func TestAggregate_LargeVolume(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	numChannels := 10
	messagesPerChannel := 100

	channels := make([]<-chan string, numChannels)
	for i := 0; i < numChannels; i++ {
		ch := make(chan string, messagesPerChannel)
		channels[i] = ch

		go func(ch chan string, id int) {
			for j := 0; j < messagesPerChannel; j++ {
				ch <- "msg"
			}
			close(ch)
		}(ch, i)
	}

	result := Aggregate(ctx, channels...)

	count := 0
	for range result {
		count++
	}

	expectedCount := numChannels * messagesPerChannel
	if count != expectedCount {
		t.Errorf("expected %d messages, got %d", expectedCount, count)
	}
}
