package payloads

import "github.com/ovh/kmip-go"

func init() {
	kmip.RegisterOperationPayload[DeriveKeyRequestPayload, DeriveKeyResponsePayload](kmip.OperationDeriveKey)
}

// This operation requests the server to derive a Symmetric Key or Secret Data object
// from a key or secret data that is already known to the key management system.
//
// The request contains the Unique Identifier(s) of the base object(s), the derivation
// method, the derivation parameters, and the attributes for the derived object.
type DeriveKeyRequestPayload struct {
	// Determines the type of object to be created (Symmetric Key or Secret Data).
	ObjectType kmip.ObjectType
	// Determines the object or objects to be used to derive a new key. At least one
	// base object SHALL be provided.
	UniqueIdentifier []string
	// An Enumeration object specifying the method to be used to derive the new key.
	DerivationMethod kmip.DerivationMethod
	// A structure containing the parameters needed by the specified derivation method.
	DerivationParameters kmip.DerivationParameters
	// Specifies desired attributes to be associated with the new object.
	//
	// Replaced by Attributes in KMIP 2.0.
	TemplateAttribute kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
	// Attributes carries the derived object's attributes in KMIP 2.0+.
	Attributes kmip.Attributes `ttlv:"Attributes,omitempty,version=v2.0.."`
}

func (pl *DeriveKeyRequestPayload) Operation() kmip.Operation {
	return kmip.OperationDeriveKey
}

// Response for the DeriveKey operation.
type DeriveKeyResponsePayload struct {
	// The Unique Identifier of the newly derived object.
	UniqueIdentifier string
	// An OPTIONAL list of object attributes set implicitly by the server.
	//
	// Removed in KMIP 2.0: the response returns only the Unique Identifier.
	TemplateAttribute *kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
}

func (pl *DeriveKeyResponsePayload) Operation() kmip.Operation {
	return kmip.OperationDeriveKey
}
