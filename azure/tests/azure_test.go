package azure_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	protostreaming "github.com/tailored-agentic-units/protocol/streaming"
	"github.com/tailored-agentic-units/provider/azure"
)

func validAzureConfig() *config.ProviderConfig {
	return &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}
}

func TestNewAzure(t *testing.T) {
	p, err := azure.NewAzure(validAzureConfig())
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	if p == nil {
		t.Fatal("NewAzure returned nil provider")
	}

	if p.Name() != "azure" {
		t.Errorf("got name %q, want %q", p.Name(), "azure")
	}
}

func TestNewAzure_MissingDeployment(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	_, err := azure.NewAzure(cfg)
	if err == nil {
		t.Error("expected error for missing deployment, got nil")
	}
}

func TestNewAzure_MissingAuthType(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	_, err := azure.NewAzure(cfg)
	if err == nil {
		t.Error("expected error for missing auth_type, got nil")
	}
}

func TestNewAzure_MissingAPIVersion(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment": "gpt-4-deployment",
			"auth_type":  "api_key",
			"token":      "test-key",
		},
	}

	_, err := azure.NewAzure(cfg)
	if err == nil {
		t.Error("expected error for missing api_version, got nil")
	}
}

func TestNewAzure_MissingToken_ApiKey(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"api_version": "2024-02-01",
		},
	}

	_, err := azure.NewAzure(cfg)
	if err == nil {
		t.Error("expected error for missing token, got nil")
	}
}

func TestNewAzure_MissingToken_Bearer(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "bearer",
			"api_version": "2024-02-01",
		},
	}

	_, err := azure.NewAzure(cfg)
	if err == nil {
		t.Error("expected error for missing token with bearer auth, got nil")
	}

	if !strings.Contains(err.Error(), `auth_type "bearer"`) {
		t.Errorf("error should mention auth_type, got: %v", err)
	}
}

func TestNewAzure_UnsupportedAuthType(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "unknown",
			"api_version": "2024-02-01",
		},
	}

	_, err := azure.NewAzure(cfg)
	if err == nil {
		t.Error("expected error for unsupported auth_type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported auth_type") {
		t.Errorf("error should mention unsupported auth_type, got: %v", err)
	}
}

func TestAzure_Endpoint(t *testing.T) {
	p, err := azure.NewAzure(validAzureConfig())
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	tests := []struct {
		proto    protocol.Protocol
		expected string
	}{
		{
			protocol.Chat,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Vision,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Tools,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Embeddings,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/embeddings?api-version=2024-02-01",
		},
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

func TestAzure_PrepareRequest(t *testing.T) {
	p, err := azure.NewAzure(validAzureConfig())
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	body := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}`)
	headers := map[string]string{"Content-Type": "application/json"}

	request, err := p.PrepareRequest(context.Background(), protocol.Chat, body, headers)
	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareRequest returned nil request")
	}

	expectedURL := "https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01"
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

func TestAzure_PrepareStreamRequest(t *testing.T) {
	p, err := azure.NewAzure(validAzureConfig())
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	body := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}],"stream":true}`)
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

func TestAzure_SetHeaders_ApiKey(t *testing.T) {
	p, err := azure.NewAzure(validAzureConfig())
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	req, _ := http.NewRequest("POST", "https://example.com", nil)
	if err := p.SetHeaders(context.Background(), req); err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	if req.Header.Get("api-key") != "test-key" {
		t.Errorf("got api-key header %q, want %q", req.Header.Get("api-key"), "test-key")
	}
}

func TestAzure_SetHeaders_Bearer(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "bearer",
			"token":       "test-bearer-token",
			"api_version": "2024-02-01",
		},
	}

	p, err := azure.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	req, _ := http.NewRequest("POST", "https://example.com", nil)
	if err := p.SetHeaders(context.Background(), req); err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	expected := "Bearer test-bearer-token"
	if req.Header.Get("Authorization") != expected {
		t.Errorf("got Authorization header %q, want %q", req.Header.Get("Authorization"), expected)
	}
}

func TestAzure_Name(t *testing.T) {
	tests := []struct {
		name     string
		cfgName  string
		expected string
	}{
		{"standard name", "azure", "azure"},
		{"custom name", "my-azure", "my-azure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validAzureConfig()
			cfg.Name = tt.cfgName

			p, err := azure.NewAzure(cfg)
			if err != nil {
				t.Fatalf("NewAzure failed: %v", err)
			}

			if p.Name() != tt.expected {
				t.Errorf("got name %q, want %q", p.Name(), tt.expected)
			}
		})
	}
}
