package payloads

import (
	"time"

	"github.com/ovh/kmip-go"
)

func init() {
	kmip.RegisterOperationPayload[LoginRequestPayload, LoginResponsePayload](kmip.OperationLogin)
	kmip.RegisterOperationPayload[LogoutRequestPayload, LogoutResponsePayload](kmip.OperationLogout)
}

// Login is a KMIP 2.0 operation that authenticates a client and establishes a
// session, returning a Ticket the client presents on subsequent requests.
type LoginRequestPayload struct {
	// The requested lifetime of the session ticket.
	LeaseTime *time.Duration
	// The number of requests the ticket may be used for.
	RequestCount int32 `ttlv:",omitempty"`
}

func (pl *LoginRequestPayload) Operation() kmip.Operation {
	return kmip.OperationLogin
}

// Response for the Login operation.
type LoginResponsePayload struct {
	// The ticket issued for the authenticated session.
	Ticket kmip.Ticket
}

func (pl *LoginResponsePayload) Operation() kmip.Operation {
	return kmip.OperationLogin
}

// Logout is a KMIP 2.0 operation that ends the session identified by the ticket.
type LogoutRequestPayload struct {
	// The ticket identifying the session to end.
	Ticket kmip.Ticket
}

func (pl *LogoutRequestPayload) Operation() kmip.Operation {
	return kmip.OperationLogout
}

// Response for the Logout operation. It carries no data.
type LogoutResponsePayload struct{}

func (pl *LogoutResponsePayload) Operation() kmip.Operation {
	return kmip.OperationLogout
}
