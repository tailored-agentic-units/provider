package ollama_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	protostreaming "github.com/tailored-agentic-units/protocol/streaming"
	"github.com/tailored-agentic-units/provider/ollama"
)

func TestNewOllama(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	p, err := ollama.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	if p == nil {
		t.Fatal("NewOllama returned nil provider")
	}

	if p.Name() != "ollama" {
		t.Errorf("got name %q, want %q", p.Name(), "ollama")
	}
}

func TestNewOllama_URLSuffixHandling(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "URL without /v1 suffix",
			baseURL:     "http://localhost:11434",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
		{
			name:        "URL with /v1 suffix",
			baseURL:     "http://localhost:11434/v1",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
		{
			name:        "URL with trailing slash",
			baseURL:     "http://localhost:11434/",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: tt.baseURL,
			}

			p, err := ollama.NewOllama(cfg)
			if err != nil {
				t.Fatalf("NewOllama failed: %v", err)
			}

			endpoint, err := p.Endpoint(protocol.Chat)
			if err != nil {
				t.Fatalf("Endpoint failed: %v", err)
			}

			if endpoint != tt.expectedURL {
				t.Errorf("got endpoint %q, want %q", endpoint, tt.expectedURL)
			}
		})
	}
}

func TestOllama_Endpoint(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	p, err := ollama.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	tests := []struct {
		proto    protocol.Protocol
		expected string
	}{
		{protocol.Chat, "http://localhost:11434/v1/chat/completions"},
		{protocol.Vision, "http://localhost:11434/v1/chat/completions"},
		{protocol.Tools, "http://localhost:11434/v1/chat/completions"},
		{protocol.Embeddings, "http://localhost:11434/v1/embeddings"},
	}

	for _, tt := range tests {
		t.Run(string(tt.proto), func(t *testing.T) {
			endpoint, err := p.Endpoint(tt.proto)
			if err != nil {
				t.Fatalf("Endpoint failed: %v", err)
			}

			if endpoint != tt.expected {
				t.Errorf("got endpoint %q, want %q", endpoint, tt.expected)
			}
		})
	}
}

func TestOllama_PrepareRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	p, err := ollama.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	body := []byte(`{"model":"llama2","messages":[{"role":"user","content":"Hello"}]}`)
	headers := map[string]string{"Content-Type": "application/json"}

	request, err := p.PrepareRequest(context.Background(), protocol.Chat, body, headers)
	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareRequest returned nil request")
	}

	expectedURL := "http://localhost:11434/v1/chat/completions"
	if request.URL != expectedURL {
		t.Errorf("got URL %q, want %q", request.URL, expectedURL)
	}

	if len(request.Body) == 0 {
		t.Error("request body is empty")
	}

	if request.Headers == nil {
		t.Error("request headers is nil")
	}
}

func TestOllama_PrepareStreamRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	p, err := ollama.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	body := []byte(`{"model":"llama2","messages":[{"role":"user","content":"Hello"}],"stream":true}`)
	headers := map[string]string{"Content-Type": "application/json"}

	request, err := p.PrepareStreamRequest(context.Background(), protocol.Chat, body, headers)
	if err != nil {
		t.Fatalf("PrepareStreamRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareStreamRequest returned nil request")
	}

	if request.Headers["Accept"] != protostreaming.SSEMedia {
		t.Errorf("got Accept header %q, want %q", request.Headers["Accept"], protostreaming.SSEMedia)
	}

	if request.Headers["Cache-Control"] != "no-cache" {
		t.Errorf("got Cache-Control header %q, want %q", request.Headers["Cache-Control"], "no-cache")
	}
}

func TestOllama_SetHeaders_Bearer(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Options: map[string]any{
			"auth_type": "bearer",
			"token":     "test-bearer-token",
		},
	}

	p, err := ollama.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	req, _ := http.NewRequest("POST", "http://localhost:11434/v1/chat/completions", nil)
	if err := p.SetHeaders(context.Background(), req); err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	expected := "Bearer test-bearer-token"
	if req.Header.Get("Authorization") != expected {
		t.Errorf("got Authorization header %q, want %q", req.Header.Get("Authorization"), expected)
	}
}

func TestOllama_SetHeaders_ApiKey(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Options: map[string]any{
			"auth_type": "api_key",
			"token":     "test-api-key",
		},
	}

	p, err := ollama.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	req, _ := http.NewRequest("POST", "http://localhost:11434/v1/chat/completions", nil)
	if err := p.SetHeaders(context.Background(), req); err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	if req.Header.Get("X-API-Key") != "test-api-key" {
		t.Errorf("got X-API-Key header %q, want %q", req.Header.Get("X-API-Key"), "test-api-key")
	}
}

func TestOllama_SetHeaders_NoAuth(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
	}

	p, err := ollama.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	req, _ := http.NewRequest("POST", "http://localhost:11434/v1/chat/completions", nil)
	if err := p.SetHeaders(context.Background(), req); err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	if req.Header.Get("Authorization") != "" {
		t.Errorf("expected no Authorization header, got %q", req.Header.Get("Authorization"))
	}

	if req.Header.Get("X-API-Key") != "" {
		t.Errorf("expected no X-API-Key header, got %q", req.Header.Get("X-API-Key"))
	}
}

func TestOllama_Name(t *testing.T) {
	tests := []struct {
		name     string
		cfgName  string
		expected string
	}{
		{"standard name", "ollama", "ollama"},
		{"custom name", "my-ollama", "my-ollama"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ProviderConfig{
				Name:    tt.cfgName,
				BaseURL: "http://localhost:11434",
			}

			p, err := ollama.NewOllama(cfg)
			if err != nil {
				t.Fatalf("NewOllama failed: %v", err)
			}

			if p.Name() != tt.expected {
				t.Errorf("got name %q, want %q", p.Name(), tt.expected)
			}
		})
	}
}
