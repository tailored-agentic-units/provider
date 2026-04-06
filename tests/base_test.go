package provider_test

import (
	"testing"

	"github.com/tailored-agentic-units/provider"
)

func TestBaseProvider_Name(t *testing.T) {
	bp := provider.NewBaseProvider("test-provider", "https://example.com")

	if bp.Name() != "test-provider" {
		t.Errorf("got name %q, want %q", bp.Name(), "test-provider")
	}
}

func TestBaseProvider_BaseURL(t *testing.T) {
	bp := provider.NewBaseProvider("test-provider", "https://example.com")

	if bp.BaseURL() != "https://example.com" {
		t.Errorf("got base URL %q, want %q", bp.BaseURL(), "https://example.com")
	}
}

func TestBaseProvider_EmptyValues(t *testing.T) {
	bp := provider.NewBaseProvider("", "")

	if bp.Name() != "" {
		t.Errorf("got name %q, want empty", bp.Name())
	}

	if bp.BaseURL() != "" {
		t.Errorf("got base URL %q, want empty", bp.BaseURL())
	}
}
