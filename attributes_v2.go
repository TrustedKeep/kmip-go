package kmip

import (
	"github.com/ovh/kmip-go/ttlv"
)

// This file implements the KMIP 2.0 attribute model (Phase B prototype).
//
// KMIP 2.0 removes Template-Attribute and the name-on-the-wire Attribute model.
// Instead, attributes are carried inside an "Attributes" structure where each
// attribute is encoded under its own dedicated TTLV tag (e.g. Cryptographic
// Algorithm is tag 0x420028), in ascending tag order. Individual attributes are
// also carried alone inside the "Current Attribute" and "New Attribute"
// structures by the attribute-mutation operations.
//
// The logical Go representation reuses the existing [Attribute] value (an
// AttributeName paired with a typed value), so higher layers can share a single
// attribute abstraction across protocol versions; only the wire codec differs.

// attrNameToTag maps each standard attribute name to its dedicated KMIP 2.0 TTLV
// tag. It is keyed off the compile-time Tag* constants (rather than the ttlv name
// registry) to avoid any dependency on package init ordering. attrTagToName is the
// reverse mapping, built in init.
var (
	attrNameToTag = map[AttributeName]int{
		AttributeNameUniqueIdentifier:               TagUniqueIdentifier,
		AttributeNameName:                           TagName,
		AttributeNameObjectType:                     TagObjectType,
		AttributeNameOperationPolicyName:            TagOperationPolicyName,
		AttributeNameObjectGroup:                    TagObjectGroup,
		AttributeNameContactInformation:             TagContactInformation,
		AttributeNameInitialDate:                    TagInitialDate,
		AttributeNameActivationDate:                 TagActivationDate,
		AttributeNameProcessStartDate:               TagProcessStartDate,
		AttributeNameProtectStopDate:                TagProtectStopDate,
		AttributeNameDeactivationDate:               TagDeactivationDate,
		AttributeNameDestroyDate:                    TagDestroyDate,
		AttributeNameCompromiseOccurrenceDate:       TagCompromiseOccurrenceDate,
		AttributeNameCompromiseDate:                 TagCompromiseDate,
		AttributeNameArchiveDate:                    TagArchiveDate,
		AttributeNameLastChangeDate:                 TagLastChangeDate,
		AttributeNameCryptographicLength:            TagCryptographicLength,
		AttributeNameLeaseTime:                      TagLeaseTime,
		AttributeNameCryptographicAlgorithm:         TagCryptographicAlgorithm,
		AttributeNameCryptographicParameters:        TagCryptographicParameters,
		AttributeNameCryptographicDomainParameters:  TagCryptographicDomainParameters,
		AttributeNameCertificateType:                TagCertificateType,
		AttributeNameDigest:                         TagDigest,
		AttributeNameCryptographicUsageMask:         TagCryptographicUsageMask,
		AttributeNameState:                          TagState,
		AttributeNameRevocationReason:               TagRevocationReason,
		AttributeNameLink:                           TagLink,
		AttributeNameCertificateIdentifier:          TagCertificateIdentifier,
		AttributeNameCertificateSubject:             TagCertificateSubject,
		AttributeNameCertificateIssuer:              TagCertificateIssuer,
		AttributeNameUsageLimits:                    TagUsageLimits,
		AttributeNameApplicationSpecificInformation: TagApplicationSpecificInformation,
		AttributeNameCertificateLength:              TagCertificateLength,
		AttributeNameFresh:                          TagFresh,
		AttributeNameX509CertificateIdentifier:      TagX_509CertificateIdentifier,
		AttributeNameX509CertificateSubject:         TagX_509CertificateSubject,
		AttributeNameX509CertificateIssuer:          TagX_509CertificateIssuer,
		AttributeNameDigitalSignatureAlgorithm:      TagDigitalSignatureAlgorithm,
		AttributeNameAlternativeName:                TagAlternativeName,
		AttributeNameKeyValuePresent:                TagKeyValuePresent,
		AttributeNameKeyValueLocation:               TagKeyValueLocation,
		AttributeNameOriginalCreationDate:           TagOriginalCreationDate,
		AttributeNameRandomNumberGenerator:          TagRandomNumberGenerator,
		AttributeNamePKCS_12FriendlyName:            TagPKCS_12FriendlyName,
		AttributeNameDescription:                    TagDescription,
		AttributeNameComment:                        TagComment,
		AttributeNameSensitive:                      TagSensitive,
		AttributeNameAlwaysSensitive:                TagAlwaysSensitive,
		AttributeNameExtractable:                    TagExtractable,
		AttributeNameNeverExtractable:               TagNeverExtractable,
		AttributeNameProtectionStorageMask:          TagProtectionStorageMask,
	}
	attrTagToName = map[int]AttributeName{}
)

func init() {
	for name, tag := range attrNameToTag {
		attrTagToName[tag] = name
	}
	// The Enumeration form of AttributeReference uses an attribute's dedicated tag as
	// its value and the canonical tag name (e.g. "CryptographicAlgorithm") as its
	// text form. Any attribute tag may be referenced — including 2.0-only attributes
	// not in the v1.x attribute model (e.g. ShortUniqueIdentifier, CertificateIssuerC)
	// — so register the whole kmip tag table. tagNames is a package-level literal, so
	// it is fully populated before any init runs, independent of init order. A private
	// enum type carries the registration; AttributeReference itself is a struct (it may
	// also take a vendor structure form) and encodes/decodes via its own codec.
	refNames := make(map[attributeReferenceEnum]string, len(tagNames))
	for tag, name := range tagNames {
		refNames[attributeReferenceEnum(tag)] = name
	}
	ttlv.RegisterEnum(TagAttributeReference, refNames)
}

type attributeReferenceEnum uint32

// AttributeReference identifies an attribute in KMIP 2.0. It is used by
// GetAttributes to request specific attributes and by DeleteAttribute to name the
// attribute to remove. On the wire it takes one of two forms:
//   - an Enumeration whose value is a standard attribute's dedicated tag, or
//   - a Structure carrying a vendor attribute's Vendor Identification and Attribute Name.
type AttributeReference struct {
	// Tag is the dedicated tag of the referenced standard attribute (the
	// Enumeration form). It is zero for a vendor attribute reference.
	Tag int
	// VendorIdentification and AttributeName identify a vendor attribute (the
	// Structure form). They are empty for a standard attribute reference.
	VendorIdentification string
	AttributeName        string
}

// AttributeReferenceFor returns an Enumeration-form reference to the named standard
// attribute. It returns the zero AttributeReference if the name is unknown.
func AttributeReferenceFor(name AttributeName) AttributeReference {
	return AttributeReference{Tag: attrNameToTag[name]}
}

// TagEncodeTTLV implements the polymorphic AttributeReference encoding.
func (r AttributeReference) TagEncodeTTLV(e *ttlv.Encoder, tag int) {
	if r.VendorIdentification != "" || r.AttributeName != "" {
		e.Struct(tag, func(e *ttlv.Encoder) {
			e.TextString(TagVendorIdentification, r.VendorIdentification)
			e.TextString(TagAttributeName, r.AttributeName)
		})
		return
	}
	//nolint:gosec // attribute tags are small positive constants.
	e.Enum(TagAttributeReference, tag, uint32(r.Tag))
}

// TagDecodeTTLV implements the polymorphic AttributeReference decoding.
func (r *AttributeReference) TagDecodeTTLV(d *ttlv.Decoder, tag int) error {
	if d.Type() == ttlv.TypeStructure {
		return d.Struct(tag, func(d *ttlv.Decoder) error {
			if d.Tag() == TagVendorIdentification {
				if err := d.TagAny(TagVendorIdentification, &r.VendorIdentification); err != nil {
					return err
				}
			}
			return d.TagAny(TagAttributeName, &r.AttributeName)
		})
	}
	v, err := d.Enum(TagAttributeReference, tag)
	r.Tag = int(v)
	return err
}

// Attributes is the KMIP 2.0 attribute collection. On the wire it is a Structure
// (tag Attributes, 0x420125) in which each attribute is encoded under its own
// dedicated tag, in ascending tag order. It replaces the v1.x TemplateAttribute.
type Attributes []Attribute

// Get returns the first attribute value for the given name, or nil if absent.
func (a Attributes) Get(name AttributeName) any {
	for _, attr := range a {
		if attr.AttributeName == name {
			return attr.AttributeValue
		}
	}
	return nil
}

// Add appends an attribute to the collection and returns the collection, allowing
// chained construction: attrs.Add(name, v).Add(name2, v2).
func (a Attributes) Add(name AttributeName, value any) Attributes {
	return append(a, Attribute{AttributeName: name, AttributeValue: value})
}

// TagEncodeTTLV encodes the collection as an Attributes structure. Insertion order
// is preserved: KMIP 2.0 does not require ascending tag order (the OASIS conformance
// vectors themselves are unsorted), and decoders must accept any order.
func (a Attributes) TagEncodeTTLV(e *ttlv.Encoder, tag int) {
	e.Struct(tag, func(e *ttlv.Encoder) {
		for _, attr := range a {
			encodeAttribute(e, attr)
		}
	})
}

// TagDecodeTTLV decodes an Attributes structure, reading each attribute by its
// dedicated tag regardless of order.
func (a *Attributes) TagDecodeTTLV(d *ttlv.Decoder, tag int) error {
	return d.Struct(tag, func(d *ttlv.Decoder) error {
		for d.Tag() != 0 {
			attr, err := decodeAttribute(d)
			if err != nil {
				return err
			}
			*a = append(*a, attr)
		}
		return nil
	})
}

// encodeAttribute writes a single attribute under its dedicated tag. Standard
// attributes use their mapped tag; vendor attributes (with a VendorIdentification)
// are written as an Attribute sub-structure; unknown attributes previously decoded
// into a raw ttlv.Value are re-emitted under their preserved tag so they round-trip.
func encodeAttribute(e *ttlv.Encoder, attr Attribute) {
	if attr.VendorIdentification != "" {
		e.Struct(TagAttribute, func(e *ttlv.Encoder) {
			e.TextString(TagVendorIdentification, attr.VendorIdentification)
			e.TextString(TagAttributeName, string(attr.AttributeName))
			e.TagAny(TagAttributeValue, attr.AttributeValue)
		})
		return
	}
	if t, ok := attrNameToTag[attr.AttributeName]; ok {
		e.TagAny(t, attr.AttributeValue)
		return
	}
	// Unknown/new standard attribute: it was decoded generically, so re-emit it
	// under the tag preserved in the ttlv.Value.
	if v, ok := attr.AttributeValue.(ttlv.Value); ok && v.Tag != 0 {
		v.EncodeTTLV(e)
	}
}

// decodeAttribute reads a single attribute from the current position. Vendor
// attributes (Attribute sub-structures) preserve their VendorIdentification and
// AttributeName; standard attributes are decoded into their native Go type;
// unknown tags are preserved as raw ttlv.Value with their tag intact.
func decodeAttribute(d *ttlv.Decoder) (Attribute, error) {
	t := d.Tag()
	if t == TagAttribute {
		return decodeVendorAttribute(d)
	}
	name, known := attrTagToName[t]
	holder := newAttributeByTag(t)
	if err := d.TagAny(t, holder.Interface()); err != nil {
		return Attribute{}, err
	}
	attr := Attribute{AttributeValue: holder.Elem().Interface()}
	if known {
		attr.AttributeName = name
	} else {
		attr.AttributeName = AttributeName(ttlv.TagString(t))
	}
	return attr, nil
}

// decodeVendorAttribute decodes a KMIP 2.0 vendor attribute (an Attribute
// sub-structure: Vendor Identification, Attribute Name, Attribute Value). The
// value is kept as a raw ttlv.Value to preserve arbitrary vendor content.
func decodeVendorAttribute(d *ttlv.Decoder) (Attribute, error) {
	var attr Attribute
	err := d.Struct(TagAttribute, func(d *ttlv.Decoder) error {
		if d.Tag() == TagVendorIdentification {
			if err := d.TagAny(TagVendorIdentification, &attr.VendorIdentification); err != nil {
				return err
			}
		}
		var name string
		if err := d.TagAny(TagAttributeName, &name); err != nil {
			return err
		}
		attr.AttributeName = AttributeName(name)
		val := &ttlv.Value{}
		if err := d.TagAny(TagAttributeValue, val); err != nil {
			return err
		}
		attr.AttributeValue = *val
		return nil
	})
	return attr, err
}

// CurrentAttribute wraps a single existing attribute value. It is used by the
// KMIP 2.0 attribute-mutation operations (e.g. ModifyAttribute, DeleteAttribute)
// to identify the current value of an attribute.
type CurrentAttribute struct {
	Attribute
}

func (c CurrentAttribute) TagEncodeTTLV(e *ttlv.Encoder, tag int) {
	e.Struct(tag, func(e *ttlv.Encoder) { encodeAttribute(e, c.Attribute) })
}

func (c *CurrentAttribute) TagDecodeTTLV(d *ttlv.Decoder, tag int) error {
	return d.Struct(tag, func(d *ttlv.Decoder) error {
		attr, err := decodeAttribute(d)
		c.Attribute = attr
		return err
	})
}

// NewAttribute wraps a single attribute value to add or set. It is used by the
// KMIP 2.0 attribute-mutation operations (AddAttribute, ModifyAttribute,
// SetAttribute) to carry the attribute value being written.
type NewAttribute struct {
	Attribute
}

func (n NewAttribute) TagEncodeTTLV(e *ttlv.Encoder, tag int) {
	e.Struct(tag, func(e *ttlv.Encoder) { encodeAttribute(e, n.Attribute) })
}

func (n *NewAttribute) TagDecodeTTLV(d *ttlv.Decoder, tag int) error {
	return d.Struct(tag, func(d *ttlv.Decoder) error {
		attr, err := decodeAttribute(d)
		n.Attribute = attr
		return err
	})
}
