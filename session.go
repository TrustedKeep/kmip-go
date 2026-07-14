package kmip

// This file defines the KMIP 2.0 session / authorization structures used by the
// Login, Logout and Delegated Login operations.

// Ticket is a credential issued by the server (via Login or Delegated Login) that a
// client presents on subsequent requests to prove an authenticated session.
type Ticket struct {
	TicketType  TicketType `ttlv:",omitempty"`
	TicketValue []byte     `ttlv:",omitempty"`
}

// Rights is the set of rights granted to a delegated login ticket.
type Rights struct {
	Right []Right
}

// Right describes a single granted right: the operations permitted and, optionally,
// the objects they apply to.
type Right struct {
	Operations *Operations `ttlv:",omitempty"`
	Objects    *Objects    `ttlv:",omitempty"`
}

// Operations is a list of KMIP operations, used within a Right.
type Operations struct {
	Operation []Operation
}

// Objects is a list of object types, used within a Right.
type Objects struct {
	ObjectType []ObjectType
}
