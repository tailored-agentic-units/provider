# Changelog

## [v0.1.0] - 2026-04-06

Initial release. AWS Bedrock provider for the TAU ecosystem.

**Added**:
- `BedrockProvider` implementing `provider.Provider` for AWS Bedrock Converse API
- Model ID extraction from request body for endpoint routing
- `AWSCredentialSource` with default chain, static, and profile authentication
- AWS SigV4 request signing
- `EventStreamReader` implementing `protocol/streaming.StreamReader`
- `EventStreamMedia` constant for AWS event stream MIME type
- `Register()` for explicit provider registration
- `NewBedrock` factory constructor
