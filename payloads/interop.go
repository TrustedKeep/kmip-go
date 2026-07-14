package payloads

import "github.com/ovh/kmip-go"

func init() {
	kmip.RegisterOperationPayload[InteropRequestPayload, InteropResponsePayload](kmip.OperationInterop)
}

// Interop is a KMIP 2.0 operation used by test harnesses to delimit and reset
// interoperability test runs. It has no effect on managed objects.
type InteropRequestPayload struct {
	// The interop function to perform (Begin, End or Reset).
	InteropFunction kmip.InteropFunction
	// An identifier for the interop test run.
	InteropIdentifier string `ttlv:",omitempty"`
}

func (pl *InteropRequestPayload) Operation() kmip.Operation {
	return kmip.OperationInterop
}

// Response for the Interop operation. It carries no data.
type InteropResponsePayload struct{}

func (pl *InteropResponsePayload) Operation() kmip.Operation {
	return kmip.OperationInterop
}
