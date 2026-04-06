package provider_test

import (
	"context"
	"strings"
	"testing"

	"github.com/tailored-agentic-units/provider/streaming"
)

func TestSSEReader_BasicData(t *testing.T) {
	input := "data: {\"content\":\"hello\"}\n\n"
	reader := streaming.NewSSEReader()

	ctx := context.Background()
	ch := reader.ReadStream(ctx, strings.NewReader(input))

	var lines []string
	for line := range ch {
		if line.Err != nil {
			t.Fatalf("unexpected error: %v", line.Err)
		}
		if line.Done {
			break
		}
		lines = append(lines, string(line.Data))
	}

	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}

	if lines[0] != "{\"content\":\"hello\"}" {
		t.Errorf("got %q, want %q", lines[0], "{\"content\":\"hello\"}")
	}
}

func TestSSEReader_MultipleLines(t *testing.T) {
	input := "data: first\n\ndata: second\n\ndata: third\n\ndata: [DONE]\n\n"
	reader := streaming.NewSSEReader()

	ctx := context.Background()
	ch := reader.ReadStream(ctx, strings.NewReader(input))

	var lines []string
	for line := range ch {
		if line.Err != nil {
			t.Fatalf("unexpected error: %v", line.Err)
		}
		if line.Done {
			break
		}
		lines = append(lines, string(line.Data))
	}

	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}

	expected := []string{"first", "second", "third"}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("line %d: got %q, want %q", i, lines[i], want)
		}
	}
}

func TestSSEReader_DoneSignal(t *testing.T) {
	input := "data: hello\n\ndata: [DONE]\n\n"
	reader := streaming.NewSSEReader()

	ctx := context.Background()
	ch := reader.ReadStream(ctx, strings.NewReader(input))

	var gotDone bool
	for line := range ch {
		if line.Err != nil {
			t.Fatalf("unexpected error: %v", line.Err)
		}
		if line.Done {
			gotDone = true
		}
	}

	if !gotDone {
		t.Error("expected done signal, got none")
	}
}

func TestSSEReader_SkipsNonDataLines(t *testing.T) {
	input := "event: message\ndata: content\nid: 123\n\n"
	reader := streaming.NewSSEReader()

	ctx := context.Background()
	ch := reader.ReadStream(ctx, strings.NewReader(input))

	var lines []string
	for line := range ch {
		if line.Err != nil {
			t.Fatalf("unexpected error: %v", line.Err)
		}
		if line.Done {
			break
		}
		lines = append(lines, string(line.Data))
	}

	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}

	if lines[0] != "content" {
		t.Errorf("got %q, want %q", lines[0], "content")
	}
}

func TestSSEReader_EmptyStream(t *testing.T) {
	input := ""
	reader := streaming.NewSSEReader()

	ctx := context.Background()
	ch := reader.ReadStream(ctx, strings.NewReader(input))

	var count int
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("got %d lines from empty stream, want 0", count)
	}
}

func TestSSEReader_ContextCancellation(t *testing.T) {
	input := "data: first\n\ndata: second\n\ndata: third\n\n"
	reader := streaming.NewSSEReader()

	ctx, cancel := context.WithCancel(context.Background())

	ch := reader.ReadStream(ctx, strings.NewReader(input))

	// Read first line then cancel
	<-ch
	cancel()

	// Channel should eventually close
	for range ch {
		// drain remaining
	}
}
