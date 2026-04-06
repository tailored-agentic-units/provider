package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	protostreaming "github.com/tailored-agentic-units/protocol/streaming"
	"github.com/tailored-agentic-units/provider"
)

// mockProvider is a minimal Provider implementation for registry tests.
type mockProvider struct {
	*provider.BaseProvider
}

func (m *mockProvider) Endpoint(_ protocol.Protocol) (string, error) {
	return "", nil
}

func (m *mockProvider) Stream() protostreaming.StreamReader {
	return nil
}

func (m *mockProvider) SetHeaders(_ context.Context, _ *http.Request) error {
	return nil
}

func (m *mockProvider) PrepareRequest(_ context.Context, _ protocol.Protocol, _ []byte, _ map[string]string) (*provider.Request, error) {
	return nil, nil
}

func (m *mockProvider) PrepareStreamRequest(_ context.Context, _ protocol.Protocol, _ []byte, _ map[string]string) (*provider.Request, error) {
	return nil, nil
}

func mockFactory(name string) provider.Factory {
	return func(c *config.ProviderConfig) (provider.Provider, error) {
		return &mockProvider{
			BaseProvider: provider.NewBaseProvider(name, "https://example.com"),
		}, nil
	}
}

func errorFactory() provider.Factory {
	return func(c *config.ProviderConfig) (provider.Provider, error) {
		return nil, fmt.Errorf("factory error")
	}
}

func TestCreate(t *testing.T) {
	provider.Register("test-alpha", mockFactory("test-alpha"))

	p, err := provider.Create(&config.ProviderConfig{Name: "test-alpha"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if p == nil {
		t.Fatal("Create returned nil provider")
	}

	if p.Name() != "test-alpha" {
		t.Errorf("got name %q, want %q", p.Name(), "test-alpha")
	}
}

func TestCreate_UnknownProvider(t *testing.T) {
	_, err := provider.Create(&config.ProviderConfig{Name: "unknown-provider"})

	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}

func TestCreate_FactoryError(t *testing.T) {
	provider.Register("test-error", errorFactory())

	_, err := provider.Create(&config.ProviderConfig{Name: "test-error"})

	if err == nil {
		t.Error("expected error from factory, got nil")
	}
}

func TestCreate_Multiple(t *testing.T) {
	provider.Register("test-beta", mockFactory("test-beta"))
	provider.Register("test-gamma", mockFactory("test-gamma"))

	for _, name := range []string{"test-beta", "test-gamma"} {
		p, err := provider.Create(&config.ProviderConfig{Name: name})
		if err != nil {
			t.Fatalf("Create(%q) failed: %v", name, err)
		}

		if p.Name() != name {
			t.Errorf("got name %q, want %q", p.Name(), name)
		}
	}
}

func TestListProviders(t *testing.T) {
	provider.Register("test-list-a", mockFactory("test-list-a"))
	provider.Register("test-list-b", mockFactory("test-list-b"))

	names := provider.ListProviders()

	if len(names) == 0 {
		t.Error("ListProviders returned empty list")
	}

	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	if !found["test-list-a"] {
		t.Error("test-list-a not found in ListProviders")
	}

	if !found["test-list-b"] {
		t.Error("test-list-b not found in ListProviders")
	}
}
