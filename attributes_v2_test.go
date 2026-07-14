package kmip_test

import (
	"testing"

	"github.com/ovh/kmip-go"
	"github.com/ovh/kmip-go/ttlv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oasisMDO1Attributes is the <Attributes> structure exactly as it appears in the
// OASIS KMIP 2.0 conformance test case TC-MDO-1-20 (Register a SymmetricKey).
// Note the wire order is NOT ascending by tag (KeyValueLocation 0x4200B8 precedes
// Name 0x420053), which is why the decoder must be order-independent.
const oasisMDO1Attributes = `<Attributes>
  <CryptographicAlgorithm type="Enumeration" value="AES"/>
  <CryptographicLength type="Integer" value="128"/>
  <CryptographicUsageMask type="Integer" value="Encrypt Verify Sign"/>
  <KeyValueLocation>
    <KeyValueLocationValue type="TextString" value="HSM-13345"/>
    <KeyValueLocationType type="Enumeration" value="UninterpretedTextString"/>
  </KeyValueLocation>
  <Name>
    <NameValue type="TextString" value="TC-MDO-1-20"/>
    <NameType type="Enumeration" value="UninterpretedTextString"/>
  </Name>
</Attributes>`

// TestAttributesDecodeOASISVector decodes the real OASIS 2.0 Attributes structure
// and asserts every attribute is read with the correct dedicated tag and typed value.
func TestAttributesDecodeOASISVector(t *testing.T) {
	dec, err := ttlv.NewXMLDecoder([]byte(oasisMDO1Attributes))
	require.NoError(t, err)

	var attrs kmip.Attributes
	require.NoError(t, dec.TagAny(kmip.TagAttributes, &attrs))

	require.Len(t, attrs, 5)

	// Values are read into their native Go types via the shared attrTypes registry.
	assert.Equal(t, kmip.CryptographicAlgorithmAES, attrs.Get(kmip.AttributeNameCryptographicAlgorithm))
	assert.Equal(t, int32(128), attrs.Get(kmip.AttributeNameCryptographicLength))
	assert.Equal(t,
		kmip.CryptographicUsageEncrypt|kmip.CryptographicUsageVerify|kmip.CryptographicUsageSign,
		attrs.Get(kmip.AttributeNameCryptographicUsageMask))

	name, ok := attrs.Get(kmip.AttributeNameName).(kmip.Name)
	require.True(t, ok)
	assert.Equal(t, "TC-MDO-1-20", name.NameValue)

	loc, ok := attrs.Get(kmip.AttributeNameKeyValueLocation).(kmip.KeyValueLocation)
	require.True(t, ok)
	assert.Equal(t, "HSM-13345", loc.KeyValueLocationValue)
}

// TestAttributesRoundTrip verifies that a decoded Attributes structure re-encodes
// and decodes identically across binary TTLV and XML, preserving order and values.
func TestAttributesRoundTrip(t *testing.T) {
	dec, err := ttlv.NewXMLDecoder([]byte(oasisMDO1Attributes))
	require.NoError(t, err)
	var orig kmip.Attributes
	require.NoError(t, dec.TagAny(kmip.TagAttributes, &orig))

	for _, tc := range []struct {
		name string
		enc  func() ttlv.Encoder
		dec  func([]byte) (ttlv.Decoder, error)
	}{
		{"binary", ttlv.NewTTLVEncoder, ttlv.NewTTLVDecoder},
		{"xml", ttlv.NewXMLEncoder, ttlv.NewXMLDecoder},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enc := tc.enc()
			enc.TagAny(kmip.TagAttributes, orig)
			data := enc.Bytes()

			d, err := tc.dec(data)
			require.NoError(t, err)
			var got kmip.Attributes
			require.NoError(t, d.TagAny(kmip.TagAttributes, &got))

			require.Len(t, got, len(orig))
			for i := range orig {
				assert.Equal(t, orig[i].AttributeName, got[i].AttributeName, "attr %d name", i)
				assert.Equal(t, orig[i].AttributeValue, got[i].AttributeValue, "attr %d value", i)
			}
		})
	}
}
