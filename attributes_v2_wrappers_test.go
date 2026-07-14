package kmip_test

import (
	"testing"

	"github.com/ovh/kmip-go"
	"github.com/ovh/kmip-go/ttlv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCurrentNewAttributeRoundTrip verifies the KMIP 2.0 Current Attribute and
// New Attribute wrapper structures round-trip through binary TTLV, each carrying
// exactly one attribute under its dedicated tag.
func TestCurrentNewAttributeRoundTrip(t *testing.T) {
	cur := kmip.CurrentAttribute{Attribute: kmip.Attribute{
		AttributeName:  kmip.AttributeNameState,
		AttributeValue: kmip.StateActive,
	}}
	newAttr := kmip.NewAttribute{Attribute: kmip.Attribute{
		AttributeName:  kmip.AttributeNameSensitive,
		AttributeValue: true,
	}}

	t.Run("CurrentAttribute", func(t *testing.T) {
		enc := ttlv.NewTTLVEncoder()
		enc.TagAny(kmip.TagCurrentAttribute, cur)
		dec, err := ttlv.NewTTLVDecoder(enc.Bytes())
		require.NoError(t, err)
		var got kmip.CurrentAttribute
		require.NoError(t, dec.TagAny(kmip.TagCurrentAttribute, &got))
		assert.Equal(t, cur, got)
	})

	t.Run("NewAttribute", func(t *testing.T) {
		enc := ttlv.NewTTLVEncoder()
		enc.TagAny(kmip.TagNewAttribute, newAttr)
		dec, err := ttlv.NewTTLVDecoder(enc.Bytes())
		require.NoError(t, err)
		var got kmip.NewAttribute
		require.NoError(t, dec.TagAny(kmip.TagNewAttribute, &got))
		assert.Equal(t, newAttr, got)
	})
}

// TestAttributeReference verifies AttributeReference round-trips in both its
// Enumeration form (standard attribute by tag) and its Structure form (vendor
// attribute by identification + name).
func TestAttributeReference(t *testing.T) {
	roundTrip := func(t *testing.T, ref kmip.AttributeReference) {
		t.Helper()
		enc := ttlv.NewTTLVEncoder()
		enc.TagAny(kmip.TagAttributeReference, ref)
		dec, err := ttlv.NewTTLVDecoder(enc.Bytes())
		require.NoError(t, err)
		var got kmip.AttributeReference
		require.NoError(t, dec.TagAny(kmip.TagAttributeReference, &got))
		assert.Equal(t, ref, got)
	}

	t.Run("enumeration form", func(t *testing.T) {
		ref := kmip.AttributeReferenceFor(kmip.AttributeNameSensitive)
		assert.Equal(t, kmip.TagSensitive, ref.Tag)
		roundTrip(t, ref)

		// The Enumeration renders with the canonical tag name in XML.
		enc := ttlv.NewXMLEncoder()
		enc.TagAny(kmip.TagAttributeReference, ref)
		assert.Contains(t, string(enc.Bytes()), `value="Sensitive"`)
	})

	t.Run("vendor structure form", func(t *testing.T) {
		roundTrip(t, kmip.AttributeReference{VendorIdentification: "x", AttributeName: "ID"})
	})
}
