// Package bedrock implements the Provider interface for AWS Bedrock
// using the Converse API.
//
// It includes AWS credential management with SigV4 request signing and
// an EventStreamReader for Bedrock's binary event stream format.
//
// Register explicitly when needed:
//
//	import "github.com/tailored-agentic-units/provider/bedrock"
//	bedrock.Register()
package bedrock
