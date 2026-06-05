package bedrock_test

import (
	"testing"

	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	"github.com/tailored-agentic-units/provider/bedrock"
)

func validBedrockConfig() *config.ProviderConfig {
	return &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":            "us-east-1",
			"auth_type":         "static",
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}
}

func TestNewBedrock(t *testing.T) {
	p, err := bedrock.NewBedrock(validBedrockConfig())
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	if p == nil {
		t.Fatal("NewBedrock returned nil provider")
	}

	if p.Name() != "bedrock" {
		t.Errorf("got name %q, want %q", p.Name(), "bedrock")
	}
}

func TestNewBedrock_AutoBaseURL(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":            "us-west-2",
			"auth_type":         "static",
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	p, err := bedrock.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	expectedURL := "https://bedrock-runtime.us-west-2.amazonaws.com"
	if p.BaseURL() != expectedURL {
		t.Errorf("got base URL %q, want %q", p.BaseURL(), expectedURL)
	}
}

func TestNewBedrock_CustomBaseURL(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "bedrock",
		BaseURL: "https://custom-endpoint.example.com",
		Options: map[string]any{
			"region":            "us-east-1",
			"auth_type":         "static",
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	p, err := bedrock.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	if p.BaseURL() != "https://custom-endpoint.example.com" {
		t.Errorf("got base URL %q, want %q", p.BaseURL(), "https://custom-endpoint.example.com")
	}
}

func TestNewBedrock_MissingRegion(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "bedrock",
		Options: map[string]any{},
	}

	_, err := bedrock.NewBedrock(cfg)
	if err == nil {
		t.Error("expected error for missing region, got nil")
	}
}

func TestBedrock_Endpoint(t *testing.T) {
	p, err := bedrock.NewBedrock(validBedrockConfig())
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	tests := []struct {
		name        string
		proto       protocol.Protocol
		expectError bool
	}{
		{name: "chat", proto: protocol.Chat, expectError: false},
		{name: "vision", proto: protocol.Vision, expectError: false},
		{name: "tools", proto: protocol.Tools, expectError: false},
		{name: "embeddings", proto: protocol.Embeddings, expectError: true},
		{name: "unsupported", proto: protocol.Protocol("unsupported"), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Endpoint(tt.proto)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBedrock_Name(t *testing.T) {
	tests := []struct {
		name     string
		cfgName  string
		expected string
	}{
		{"standard name", "bedrock", "bedrock"},
		{"custom name", "my-bedrock", "my-bedrock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validBedrockConfig()
			cfg.Name = tt.cfgName

			p, err := bedrock.NewBedrock(cfg)
			if err != nil {
				t.Fatalf("NewBedrock failed: %v", err)
			}

			if p.Name() != tt.expected {
				t.Errorf("got name %q, want %q", p.Name(), tt.expected)
			}
		})
	}
}
