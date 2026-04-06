# provider/bedrock

AWS Bedrock provider for the [TAU](https://github.com/tailored-agentic-units) ecosystem.

```
go get github.com/tailored-agentic-units/provider/bedrock
```

## Supported Protocols

| Protocol | Endpoint |
|----------|----------|
| Chat | `/model/{modelId}/converse` |
| Vision | `/model/{modelId}/converse` |
| Tools | `/model/{modelId}/converse` |

Embeddings are not supported by the Bedrock Converse API.

## Authentication

| Type | Description |
|------|-------------|
| `default` | Default AWS credential chain |
| `static` | Explicit access key, secret key, and optional session token |
| `profile` | Named profile from shared AWS configuration |

All requests are signed with AWS SigV4.

## Usage

```go
import (
    "github.com/tailored-agentic-units/provider"
    "github.com/tailored-agentic-units/provider/bedrock"
    "github.com/tailored-agentic-units/protocol/config"
)

bedrock.Register()

p, err := provider.Create(&config.ProviderConfig{
    Name: "bedrock",
    Options: map[string]any{
        "region":    "us-east-1",
        "auth_type": "default",
    },
})
```

## Dependencies

- `github.com/tailored-agentic-units/provider` — Provider interface and BaseProvider
- `github.com/tailored-agentic-units/protocol` — Protocol constants and streaming types
- `github.com/aws/aws-sdk-go-v2` — AWS SDK core, SigV4 signing
- `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream` — Event stream decoding
- `github.com/aws/aws-sdk-go-v2/config` — AWS configuration loading
- `github.com/aws/aws-sdk-go-v2/credentials` — Static credential provider
