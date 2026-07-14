package payloads

import (
	"github.com/ovh/kmip-go"
	"github.com/ovh/kmip-go/ttlv"
)

func init() {
	kmip.RegisterOperationPayload[RegisterRequestPayload, RegisterResponsePayload](kmip.OperationRegister)
}

// This operation requests the server to register a Managed Object that was created by the client or obtained by the client through some other means,
// allowing the server to manage the object. The arguments in the request are similar to those in the Create operation, but contain the object itself for storage by the server.
//
// The request contains information about the type of object being registered and attributes to be assigned to the object
// (e.g., Cryptographic Algorithm, Cryptographic Length, etc.). This information SHALL be specified by the use of a Template-Attribute object.
//
// The response contains the Unique Identifier assigned by the server to the registered object. The server SHALL copy the Unique Identifier returned by this
// operations into the ID Placeholder variable. The Initial Date attribute of the object SHALL be set to the current time.
//
// If a Managed Cryptographic Object is registered, then the following attributes SHALL be included in the Register request, either explicitly,
// or via specification of a template that contains the attribute:
//   - Cryptographic Algorithm: MAY be omitted only if this information is encapsulated in the Key Block. Does not apply to Secret Data. If present, then Cryptographic Length below SHALL also be present.
//   - Cryptographic Length: MAY be omitted only if this information is encapsulated in the Key Block. Does not apply to Secret Data. If present, then Cryptographic Algorithm above SHALL also be present.
//   - Certificate Length: Only applies to Certificates.
//   - Cryptographic Usage Mask
//   - Digital Signature Algorithm: MAY be omitted only if this information is encapsulated in the Certificate object. Only applies to Certificates.
type RegisterRequestPayload struct {
	// Determines the type of object being registered.
	ObjectType kmip.ObjectType
	// Specifies desired object attributes to be associated with the new object using templates and/or individual attributes.
	//
	// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
	// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
	//
	// Removed in KMIP 2.0 in favor of Attributes.
	TemplateAttribute kmip.TemplateAttribute `ttlv:",omitempty,version=..v1.4"`
	// Attributes carries the object's attributes in KMIP 2.0+, replacing TemplateAttribute.
	Attributes kmip.Attributes `ttlv:"Attributes,omitempty,version=v2.0.."`
	// The object being registered. The object and attributes MAY be wrapped.
	Object kmip.Object
}

func (pl *RegisterRequestPayload) Operation() kmip.Operation {
	return kmip.OperationRegister
}

func (pl *RegisterRequestPayload) TagDecodeTTLV(d *ttlv.Decoder, tag int) error {
	return d.Struct(tag, func(d *ttlv.Decoder) error {
		if err := d.Any(&pl.ObjectType); err != nil {
			return err
		}

		// Attributes are carried either as a TemplateAttribute (v1.x) or as an
		// Attributes structure (v2.0+). Branch on the tag actually present so we
		// decode correctly regardless of the negotiated version.
		switch d.Tag() {
		case kmip.TagAttributes:
			if err := d.TagAny(kmip.TagAttributes, &pl.Attributes); err != nil {
				return err
			}
		case kmip.TagTemplateAttribute:
			if err := d.Any(&pl.TemplateAttribute); err != nil {
				return err
			}
		}

		var err error
		if pl.Object, err = kmip.NewObjectForType(pl.ObjectType); err != nil {
			return err
		}
		return d.Any(&pl.Object)
	})
}

// Response for the Register operation.
type RegisterResponsePayload struct {
	// The Unique Identifier of the newly registered object.
	UniqueIdentifier string
	// An OPTIONAL list of object attributes with values that were not specified in the request, but have been implicitly set by the key management server.
	//
	// The Template Managed Object is deprecated as of version 1.3 of this specification and MAY be removed from subsequent versions of the specification.
	// Individual Attributes SHOULD be used in operations which currently support use of a Name within a Template-Attribute to reference a Template.
	//
	// Removed in KMIP 2.0: the Register response returns only the Unique Identifier.
	TemplateAttribute *kmip.TemplateAttribute `ttlv:",version=..v1.4"`
}

func (pl *RegisterResponsePayload) Operation() kmip.Operation {
	return kmip.OperationRegister
}
