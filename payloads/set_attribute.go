package payloads

import "github.com/ovh/kmip-go"

func init() {
	kmip.RegisterOperationPayload[SetAttributeRequestPayload, SetAttributeResponsePayload](kmip.OperationSetAttribute)
}

// SetAttribute is a KMIP 2.0 operation that requests the server to set the value of
// an attribute of a Managed Object. Unlike AddAttribute, which fails if the attribute
// already exists, SetAttribute creates the attribute if absent and replaces it if
// present (for single-instance attributes). It applies only to attributes that are
// not multi-instance.
type SetAttributeRequestPayload struct {
	// The Unique Identifier of the object. If omitted, then the ID Placeholder value is used by the server as the Unique Identifier.
	UniqueIdentifier string `ttlv:",omitempty"`
	// Specifies the attribute value to set.
	NewAttribute kmip.NewAttribute `ttlv:"NewAttribute"`
}

func (pl *SetAttributeRequestPayload) Operation() kmip.Operation {
	return kmip.OperationSetAttribute
}

// Response for the SetAttribute operation.
type SetAttributeResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

func (pl *SetAttributeResponsePayload) Operation() kmip.Operation {
	return kmip.OperationSetAttribute
}
