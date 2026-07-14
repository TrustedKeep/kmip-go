package payloads

import "github.com/ovh/kmip-go"

func init() {
	kmip.RegisterOperationPayload[CreateKeyPairRequestPayload, CreateKeyPairResponsePayload](kmip.OperationCreateKeyPair)
}

// This operation requests the server to generate a new public/private key pair and register the two corresponding new Managed Cryptographic Objects.
//
// The request contains attributes to be assigned to the objects (e.g., Cryptographic Algorithm, Cryptographic Length, etc.).
// Attributes and Template Names MAY be specified for both keys at the same time by specifying a Common Template-Attribute object in the request.
// Attributes not common to both keys (e.g., Name, Cryptographic Usage Mask) MAY be specified using the Private Key Template-Attribute and
// Public Key Template-Attribute objects in the request, which take precedence over the Common Template-Attribute object.
//
// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
//
// For the Private Key, the server SHALL create a Link attribute of Link Type Public Key pointing to the Public Key.
// For the Public Key, the server SHALL create a Link attribute of Link Type Private Key pointing to the Private Key.
// The response contains the Unique Identifiers of both created objects. The ID Placeholder value SHALL be set to the Unique Identifier of the Private Key.
//
// For multi-instance attributes, the union of the values found in the templates and attributes of the Common, Private, and Public Key Template-Attribute SHALL be used. For single-instance attributes, the order of precedence is as follows:
//   - attributes specified explicitly in the Private and Public Key Template-Attribute, then
//   - attributes specified via templates in the Private and Public Key Template-Attribute, then
//   - attributes specified explicitly in the Common Template-Attribute, then
//   - attributes specified via templates in the Common Template-Attribute.
//
// If there are multiple templates in the Common, Private, or Public Key Template-Attribute, then the last value of the single-instance attribute that conflicts takes precedence.
type CreateKeyPairRequestPayload struct {
	// Specifies desired attributes in templates and/or as individual attributes to be associated with the new object that apply to both the Private and Public Key Objects.
	//
	// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
	// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
	//
	// Replaced by CommonAttributes in KMIP 2.0.
	CommonTemplateAttribute *kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
	// Specifies templates and/or attributes to be associated with the new object that apply to the Private Key Object. Order of precedence applies.
	//
	// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
	// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
	//
	// Replaced by PrivateKeyAttributes in KMIP 2.0.
	PrivateKeyTemplateAttribute *kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
	// Specifies templates and/or attributes to be associated with the new object that apply to the Public Key Object. Order of precedence applies.
	//
	// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
	// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
	//
	// Replaced by PublicKeyAttributes in KMIP 2.0.
	PublicKeyTemplateAttribute *kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
	// KMIP 2.0+ attributes applying to both keys.
	CommonAttributes kmip.Attributes `ttlv:"CommonAttributes,omitempty,version=v2.0.."`
	// KMIP 2.0+ attributes applying to the Private Key object.
	PrivateKeyAttributes kmip.Attributes `ttlv:"PrivateKeyAttributes,omitempty,version=v2.0.."`
	// KMIP 2.0+ attributes applying to the Public Key object.
	PublicKeyAttributes kmip.Attributes `ttlv:"PublicKeyAttributes,omitempty,version=v2.0.."`
}

func (a *CreateKeyPairRequestPayload) Operation() kmip.Operation {
	return kmip.OperationCreateKeyPair
}

// Response for the create key-pair operation.
type CreateKeyPairResponsePayload struct {
	// The Unique Identifier of the newly created Private Key object.
	PrivateKeyUniqueIdentifier string
	// The Unique Identifier of the newly created Public Key object.
	PublicKeyUniqueIdentifier string
	// An OPTIONAL list of attributes, for the Private Key Object, with values that were not specified in the request, but have been implicitly set by the key management server.
	//
	// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
	// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
	//
	// Removed in KMIP 2.0: the response returns only the two Unique Identifiers.
	PrivateKeyTemplateAttribute *kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
	// An OPTIONAL list of attributes, for the Public Key Object, with values that were not specified in the request, but have been implicitly set by the key management server.
	//
	// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
	// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
	//
	// Removed in KMIP 2.0: the response returns only the two Unique Identifiers.
	PublicKeyTemplateAttribute *kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
}

func (a *CreateKeyPairResponsePayload) Operation() kmip.Operation {
	return kmip.OperationCreateKeyPair
}
