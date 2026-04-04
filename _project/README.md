# provider

LLM transport, authentication, and streaming implementations. This is where cloud SDK weight lives — consumers who only use local inference don't need to import cloud provider dependencies.

## Vision

A transport layer that handles the mechanics of reaching LLM APIs: endpoint construction, authentication, request preparation, and streaming. Providers know nothing about wire format serialization (that's tau/format) — they handle HTTP, auth, and streaming transport.

## Core Premise

The three-layer separation: Provider handles transport, Format handles serialization, Response holds the result. Provider receives already-marshaled bytes and returns raw response bytes. This independence from tau/format means adding a new provider never requires format changes and vice versa.

## Phases

| Phase | Focus Area | Version Target |
|-------|-----------|----------------|
| Phase 1 - Foundation | Provider interface, streaming implementations, Ollama/Azure/Bedrock providers, identity management | v0.1.0 |

## Architecture

```
provider (root)    — Provider interface, BaseProvider, registry, Request type,
                     Ollama, Azure, Bedrock implementations
  streaming/       — SSE reader, EventStream reader
  identities/      — AWS SigV4, Azure managed identity
```

## Source References

**Primary source for Provider interface + Bedrock**: `~/code/go-agents/pkg/providers/`
**Primary source for streaming implementations**: `~/code/go-agents/pkg/streaming/`
**Primary source for identity management**: `~/code/go-agents/pkg/identities/`
**Merge source for Ollama + Azure**: `~/tau/kernel/agent/providers/` (kernel implementations to be reconciled with go-agents)

### File-by-File Source Mapping

#### Root Package (Provider Interface + Implementations)

| Source | Destination | Lines | Action |
|--------|-------------|-------|--------|
| `~/code/go-agents/pkg/providers/provider.go` (56 lines) | `provider.go` | Provider interface, Request type | Port with import changes |
| `~/code/go-agents/pkg/providers/base.go` (31 lines) | `base.go` | BaseProvider | Port intact |
| `~/code/go-agents/pkg/providers/registry.go` (68 lines) | `registry.go` | Provider registry | Port with import changes |
| `~/code/go-agents/pkg/providers/ollama.go` (122 lines) | `ollama.go` | Ollama provider | Merge kernel + go-agents |
| `~/code/go-agents/pkg/providers/azure.go` (178 lines) | `azure.go` | Azure provider | Merge kernel + go-agents |
| `~/code/go-agents/pkg/providers/bedrock.go` (152 lines) | `bedrock.go` | Bedrock provider | Port from go-agents (new to TAU) |

#### streaming/ Package

| Source | Destination | Lines | Action |
|--------|-------------|-------|--------|
| `~/code/go-agents/pkg/streaming/sse.go` (75 lines) | `streaming/sse.go` | SSE reader impl | Port with import changes |
| `~/code/go-agents/pkg/streaming/eventstream.go` (87 lines) | `streaming/eventstream.go` | EventStream reader impl | Port with import changes |

Note: `~/code/go-agents/pkg/streaming/streaming.go` (28 lines) — the interfaces — goes to tau/protocol, NOT here.

#### identities/ Package

| Source | Destination | Lines | Action |
|--------|-------------|-------|--------|
| `~/code/go-agents/pkg/identities/aws.go` (132 lines) | `identities/aws.go` | AWS SigV4 + credential chains | Port intact |
| `~/code/go-agents/pkg/identities/azure.go` (63 lines) | `identities/azure.go` | Azure managed identity | Port intact |

### Deviations from go-agents Source

**Import path changes**:
- `github.com/JaimeStill/go-agents/pkg/protocol` → `github.com/tailored-agentic-units/protocol`
- `github.com/JaimeStill/go-agents/pkg/config` → `github.com/tailored-agentic-units/protocol/config`
- `github.com/JaimeStill/go-agents/pkg/streaming` (interfaces) → `github.com/tailored-agentic-units/protocol/streaming`
- Internal streaming implementations use local `streaming/` package

**Streaming package split**:
- go-agents has one `streaming` package with interfaces + implementations together
- tau splits: interfaces in tau/protocol/streaming, implementations in tau/provider/streaming
- Provider implementations import from tau/protocol/streaming for the `StreamReader` interface, then return their local SSE/EventStream reader implementations

### Deviations from Kernel Source

The kernel's provider architecture is significantly different. Key changes:

**1. Marshal/ProcessResponse methods REMOVED**

Kernel's `BaseProvider` has `Marshal()` that builds OpenAI-shaped JSON:
```go
// ~/tau/kernel/agent/providers/base.go
func (bp *BaseProvider) Marshal(p protocol.Protocol, model string, data any) ([]byte, error)
```
This is **eliminated** — marshaling moves entirely to tau/format. Providers only handle transport.

**2. Provider interface gains new methods**

Kernel's Provider interface (`~/tau/kernel/agent/providers/provider.go`, 69 lines):
```go
type Provider interface {
    Name() string
    BaseURL() string
    Endpoint(p protocol.Protocol) (string, error)
    SetHeaders(req *http.Request)                              // no context, no error
    Marshal(p protocol.Protocol, model string, data any) ([]byte, error)  // moves to format
    ProcessResponse(p protocol.Protocol, body []byte) (any, error)        // moves to format
    ProcessStreamResponse(p protocol.Protocol, body []byte) (any, error)  // moves to format
}
```

go-agents Provider interface (adopted by tau/provider):
```go
type Provider interface {
    Name() string
    BaseURL() string
    Endpoint(p protocol.Protocol) (string, error)
    Stream() streaming.StreamReader                                         // NEW
    SetHeaders(ctx context.Context, req *http.Request) error                // ctx + error added
    PrepareRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)       // NEW
    PrepareStreamRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error) // NEW
}
```

Changes:
- `SetHeaders(req)` → `SetHeaders(ctx, req) error` — context for credential refresh (SigV4, Azure tokens)
- `Marshal()` → **removed** (goes to tau/format)
- `ProcessResponse()` → **removed** (goes to tau/format)
- `ProcessStreamResponse()` → **removed** (goes to tau/format)
- `Stream()` → **added** — returns a StreamReader for this provider's streaming protocol
- `PrepareRequest()` → **added** — constructs URL + headers from protocol and marshaled body
- `PrepareStreamRequest()` → **added** — same for streaming with appropriate Accept headers

**3. Data types REMOVED from providers**

`~/tau/kernel/agent/providers/data.go` (43 lines) defines `ChatData`, `VisionData`, `ToolsData`, `EmbeddingsData`, `AudioData`. These move to tau/format as format data types. tau/provider has no data types.

**4. Bedrock provider is NEW**

`~/code/go-agents/pkg/providers/bedrock.go` (152 lines) — not present in kernel. Uses Converse API, SigV4 signing, EventStream transport.

**5. Streaming is NEW**

Kernel has no streaming abstraction. Kernel's client returns `<-chan any` and providers parse streaming responses inline. go-agents' `StreamReader` interface with SSE/EventStream implementations is entirely new to the TAU ecosystem.

**6. Identity management is NEW**

Kernel providers handle auth inline (`SetHeaders` with hardcoded logic). go-agents extracts credential management into dedicated `identities/` package with `AWSCredentialSource` and `AzureTokenSource`. This enables:
- Credential refresh without provider reconstruction
- Multiple auth strategies per provider (default chain, static, profile for AWS)
- Azure managed identity support

### Kernel Files NOT Ported (Replaced by go-agents Architecture)

| Kernel File | Why Not Ported |
|-------------|---------------|
| `~/tau/kernel/agent/providers/base.go` `Marshal()` method | Marshaling moves to tau/format |
| `~/tau/kernel/agent/providers/data.go` (43 lines) | Data types move to tau/format |
| `~/tau/kernel/agent/providers/doc.go` (252 lines) | Rewritten for new architecture |

### Provider Merge Strategy (Ollama + Azure)

Both kernel and go-agents have Ollama and Azure providers. The merge strategy:

**Ollama**: Start from go-agents implementation (has `PrepareRequest`, `Stream()`, `SetHeaders(ctx)`). Verify feature parity with kernel's Ollama (auth options, URL handling).

**Azure**: Start from go-agents implementation (has managed identity via `identities.AzureTokenSource`). Kernel's Azure has deployment-based routing which should be preserved if not already in go-agents.

## Dependencies

- `github.com/tailored-agentic-units/protocol` — Protocol constants, config types, streaming interfaces
- `github.com/aws/aws-sdk-go-v2/*` — Bedrock SigV4, credential chains
- `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream` — EventStream binary framing
- `github.com/Azure/azure-sdk-for-go/sdk/*` — Azure managed identity

## Integration Points

- **tau/agent** client/ uses Provider for request preparation and streaming
- **tau/agent** constructs providers via registry during agent wiring
- **tau/kernel** transitively depends via tau/agent
