package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	protostreaming "github.com/tailored-agentic-units/protocol/streaming"
	"github.com/tailored-agentic-units/provider"
)

const bedrockService = "bedrock"

// Register registers the Bedrock provider with the global provider registry.
func Register() {
	provider.Register("bedrock", NewBedrock)
}

// NewBedrock creates a new BedrockProvider from configuration.
// Requires "region" in options. Optionally accepts "auth_type" (default, static, profile)
// and corresponding credential options.
func NewBedrock(c *config.ProviderConfig) (provider.Provider, error) {
	region, ok := c.Options["region"].(string)
	if !ok || region == "" {
		return nil, fmt.Errorf("region is required for Bedrock provider")
	}

	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	}

	authTypeStr, _ := c.Options["auth_type"].(string)
	authType := AWSAuthType(authTypeStr)

	credSource, err := NewAWSCredentialSource(
		context.Background(),
		region,
		authType,
		c.Options,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize AWS credentials: %w", err)
	}

	return &BedrockProvider{
		BaseProvider: provider.NewBaseProvider(c.Name, baseURL),
		region:       region,
		credSource:   credSource,
		stream:       NewEventStreamReader(),
	}, nil
}

// BedrockProvider implements Provider for AWS Bedrock using the Converse API.
type BedrockProvider struct {
	*provider.BaseProvider
	region     string
	credSource *AWSCredentialSource
	stream     protostreaming.StreamReader
}

// Endpoint returns the Bedrock endpoint path template for a protocol.
func (p *BedrockProvider) Endpoint(proto protocol.Protocol) (string, error) {
	switch proto {
	case protocol.Chat, protocol.Vision, protocol.Tools:
		return "/model/%s/converse", nil
	case protocol.Embeddings:
		return "", fmt.Errorf("embeddings protocol not supported by Bedrock converse API")
	default:
		return "", fmt.Errorf("protocol %s not supported by Bedrock", proto)
	}
}

// Stream returns the event stream reader for Bedrock streaming responses.
func (p *BedrockProvider) Stream() protostreaming.StreamReader {
	return p.stream
}

// PrepareRequest prepares a standard Bedrock request with model ID in the URL path.
func (p *BedrockProvider) PrepareRequest(
	ctx context.Context,
	proto protocol.Protocol,
	body []byte,
	headers map[string]string,
) (*provider.Request, error) {
	pathTemplate, err := p.Endpoint(proto)
	if err != nil {
		return nil, err
	}

	modelID, err := extractModelID(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s"+pathTemplate, p.BaseURL(), modelID)

	return &provider.Request{
		URL:     url,
		Headers: headers,
		Body:    body,
	}, nil
}

// PrepareStreamRequest prepares a streaming Bedrock request using the converse-stream endpoint.
func (p *BedrockProvider) PrepareStreamRequest(
	ctx context.Context,
	proto protocol.Protocol,
	body []byte,
	headers map[string]string,
) (*provider.Request, error) {
	modelID, err := extractModelID(body)
	if err != nil {
		return nil, err
	}

	switch proto {
	case protocol.Chat, protocol.Vision, protocol.Tools:
		// Use converse-stream endpoint
	default:
		return nil, fmt.Errorf("protocol %s does not support streaming on Bedrock", proto)
	}

	url := fmt.Sprintf("%s/model/%s/converse-stream", p.BaseURL(), modelID)

	streamHeaders := make(map[string]string)
	maps.Copy(streamHeaders, headers)
	streamHeaders["Accept"] = EventStreamMedia

	return &provider.Request{
		URL:     url,
		Headers: streamHeaders,
		Body:    body,
	}, nil
}

// SetHeaders signs the HTTP request using AWS SigV4 credentials.
func (p *BedrockProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	return p.credSource.SignRequest(ctx, req, bedrockService)
}

func extractModelID(body []byte) (string, error) {
	var envelope struct {
		ModelID string `json:"modelId"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", fmt.Errorf("failed to extract modelId from request body: %w", err)
	}
	if envelope.ModelID == "" {
		return "", fmt.Errorf("modelId is empty in request body")
	}

	return envelope.ModelID, nil
}
