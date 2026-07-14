package payloads

import (
	"time"

	"github.com/ovh/kmip-go"
)

func init() {
	kmip.RegisterOperationPayload[DelegatedLoginRequestPayload, DelegatedLoginResponsePayload](kmip.OperationDelegatedLogin)
}

// DelegatedLogin is a KMIP 2.0 operation that establishes a session restricted to a
// specific set of rights, returning a Ticket scoped to those rights.
type DelegatedLoginRequestPayload struct {
	// The requested lifetime of the session ticket.
	LeaseTime *time.Duration
	// The number of requests the ticket may be used for.
	RequestCount int32 `ttlv:",omitempty"`
	// The rights the delegated session is granted.
	Rights kmip.Rights
}

func (pl *DelegatedLoginRequestPayload) Operation() kmip.Operation {
	return kmip.OperationDelegatedLogin
}

// Response for the DelegatedLogin operation.
type DelegatedLoginResponsePayload struct {
	// The ticket issued for the delegated session.
	Ticket kmip.Ticket
}

func (pl *DelegatedLoginResponsePayload) Operation() kmip.Operation {
	return kmip.OperationDelegatedLogin
}
