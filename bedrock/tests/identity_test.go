package bedrock_test

import (
	"context"
	"testing"

	"github.com/tailored-agentic-units/provider/bedrock"
)

func TestAWSAuthType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		authType bedrock.AWSAuthType
		expected string
	}{
		{"default", bedrock.AWSAuthDefault, "default"},
		{"static", bedrock.AWSAuthStatic, "static"},
		{"profile", bedrock.AWSAuthProfile, "profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.authType) != tt.expected {
				t.Errorf("got %q, want %q", string(tt.authType), tt.expected)
			}
		})
	}
}

func TestNewAWSCredentialSource_UnsupportedAuthType(t *testing.T) {
	_, err := bedrock.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		bedrock.AWSAuthType("unsupported"),
		map[string]any{},
	)

	if err == nil {
		t.Error("expected error for unsupported auth type, got nil")
	}
}

func TestNewAWSCredentialSource_Static_MissingAccessKey(t *testing.T) {
	_, err := bedrock.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		bedrock.AWSAuthStatic,
		map[string]any{
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	)

	if err == nil {
		t.Error("expected error for missing access_key_id, got nil")
	}
}

func TestNewAWSCredentialSource_Static_MissingSecretKey(t *testing.T) {
	_, err := bedrock.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		bedrock.AWSAuthStatic,
		map[string]any{
			"access_key_id": "AKIAIOSFODNN7EXAMPLE",
		},
	)

	if err == nil {
		t.Error("expected error for missing secret_access_key, got nil")
	}
}

func TestNewAWSCredentialSource_Profile_MissingProfile(t *testing.T) {
	_, err := bedrock.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		bedrock.AWSAuthProfile,
		map[string]any{},
	)

	if err == nil {
		t.Error("expected error for missing profile, got nil")
	}
}

func TestNewAWSCredentialSource_Static_Valid(t *testing.T) {
	source, err := bedrock.NewAWSCredentialSource(
		context.Background(),
		"us-east-1",
		bedrock.AWSAuthStatic,
		map[string]any{
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	)

	if err != nil {
		t.Fatalf("NewAWSCredentialSource failed: %v", err)
	}

	if source == nil {
		t.Fatal("NewAWSCredentialSource returned nil")
	}
}

func TestNewAWSCredentialSource_Static_WithSessionToken(t *testing.T) {
	source, err := bedrock.NewAWSCredentialSource(
		context.Background(),
		"us-west-2",
		bedrock.AWSAuthStatic,
		map[string]any{
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"session_token":     "FwoGZXIvYXdzEBYaDH...",
		},
	)

	if err != nil {
		t.Fatalf("NewAWSCredentialSource failed: %v", err)
	}

	if source == nil {
		t.Fatal("NewAWSCredentialSource returned nil")
	}
}
