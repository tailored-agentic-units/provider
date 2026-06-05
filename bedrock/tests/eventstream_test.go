package bedrock_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"

	"github.com/tailored-agentic-units/provider/bedrock"
)

// encodeEvent encodes a single event stream message into binary format.
func encodeEvent(t *testing.T, messageType, eventType string, payload []byte) []byte {
	t.Helper()

	msg := eventstream.Message{
		Headers: eventstream.Headers{
			{
				Name:  ":message-type",
				Value: eventstream.StringValue(messageType),
			},
			{
				Name:  ":event-type",
				Value: eventstream.StringValue(eventType),
			},
		},
		Payload: payload,
	}

	var buf bytes.Buffer
	encoder := eventstream.NewEncoder()
	if err := encoder.Encode(&buf, msg); err != nil {
		t.Fatalf("failed to encode event: %v", err)
	}
	return buf.Bytes()
}

func TestEventStreamMedia(t *testing.T) {
	expected := "application/vnd.amazon.eventstream"
	if bedrock.EventStreamMedia != expected {
		t.Errorf("got EventStreamMedia %q, want %q", bedrock.EventStreamMedia, expected)
	}
}

func TestEventStreamReader_ContentBlockDelta(t *testing.T) {
	payload := []byte(`{"contentBlockIndex":0,"delta":{"text":"Hello"}}`)
	data := encodeEvent(t, "event", "contentBlockDelta", payload)

	reader := bedrock.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(data))

	line := <-lines
	if line.Err != nil {
		t.Fatalf("unexpected error: %v", line.Err)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(line.Data, &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	raw, ok := envelope["contentBlockDelta"]
	if !ok {
		t.Fatal("expected contentBlockDelta key in envelope")
	}

	var delta struct {
		ContentBlockIndex int `json:"contentBlockIndex"`
		Delta             struct {
			Text string `json:"text"`
		} `json:"delta"`
	}
	if err := json.Unmarshal(raw, &delta); err != nil {
		t.Fatalf("failed to unmarshal delta: %v", err)
	}

	if delta.Delta.Text != "Hello" {
		t.Errorf("got text %q, want %q", delta.Delta.Text, "Hello")
	}
}

func TestEventStreamReader_MessageStop(t *testing.T) {
	payload := []byte(`{"stopReason":"end_turn"}`)
	data := encodeEvent(t, "event", "messageStop", payload)

	reader := bedrock.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(data))

	line := <-lines
	if line.Err != nil {
		t.Fatalf("unexpected error: %v", line.Err)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(line.Data, &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	raw, ok := envelope["messageStop"]
	if !ok {
		t.Fatal("expected messageStop key in envelope")
	}

	var stop struct {
		StopReason string `json:"stopReason"`
	}
	if err := json.Unmarshal(raw, &stop); err != nil {
		t.Fatalf("failed to unmarshal stop: %v", err)
	}

	if stop.StopReason != "end_turn" {
		t.Errorf("got stop reason %q, want %q", stop.StopReason, "end_turn")
	}
}

func TestEventStreamReader_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(encodeEvent(t, "event", "contentBlockDelta", []byte(`{"delta":{"text":"Hello"}}`)))
	buf.Write(encodeEvent(t, "event", "contentBlockDelta", []byte(`{"delta":{"text":" world"}}`)))
	buf.Write(encodeEvent(t, "event", "messageStop", []byte(`{"stopReason":"end_turn"}`)))

	reader := bedrock.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), &buf)

	count := 0
	for line := range lines {
		if line.Err != nil {
			t.Fatalf("unexpected error on event %d: %v", count, line.Err)
		}
		count++
	}

	if count != 3 {
		t.Errorf("got %d events, want 3", count)
	}
}

func TestEventStreamReader_Exception(t *testing.T) {
	data := encodeEvent(t, "exception", "validationException", []byte(`{"message":"Invalid model ID"}`))

	reader := bedrock.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(data))

	line := <-lines
	if line.Err == nil {
		t.Fatal("expected error for exception event, got nil")
	}

	// Channel should close after exception
	_, open := <-lines
	if open {
		t.Error("expected channel to be closed after exception")
	}
}

func TestEventStreamReader_EOF(t *testing.T) {
	reader := bedrock.NewEventStreamReader()
	lines := reader.ReadStream(context.Background(), bytes.NewReader(nil))

	_, open := <-lines
	if open {
		t.Error("expected channel to be closed on EOF")
	}
}
