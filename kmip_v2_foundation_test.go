package kmip_test

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/ovh/kmip-go"
	"github.com/ovh/kmip-go/ttlv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// v2FoundationStruct exercises the KMIP 2.0 codec foundation end-to-end:
//   - a set-version field threads the negotiated protocol version through the codec,
//   - a Date-Time-Extended field (the new 2.0 wire type, tagged CurrentAttribute),
//   - version gating: the field only exists from v2.0 onward.
type v2FoundationStruct struct {
	Version kmip.ProtocolVersion  `ttlv:",set-version"`
	When    ttlv.DateTimeExtended `ttlv:"CurrentAttribute,version=v2.0.."`
}

// The wire header for the gated field: tag CurrentAttribute (0x42013C) with the
// Date-Time-Extended type byte (0x0B).
const dtxFieldHeader = "42013C0B"

func marshalV2(t *testing.T, s *v2FoundationStruct) []byte {
	t.Helper()
	enc := ttlv.NewTTLVEncoder()
	// Wrap in an arbitrary structure tag; the tag is irrelevant to the test.
	enc.TagAny(kmip.TagAttributes, s)
	return enc.Bytes()
}

func unmarshalV2(t *testing.T, data []byte) v2FoundationStruct {
	t.Helper()
	dec, err := ttlv.NewTTLVDecoder(data)
	require.NoError(t, err)
	var got v2FoundationStruct
	require.NoError(t, dec.TagAny(kmip.TagAttributes, &got))
	return got
}

// TestV2FoundationVersionGating verifies that a v2.0-gated Date-Time-Extended
// field is emitted only when the negotiated version is >= 2.0, and is skipped
// (both on the wire and on decode) at v1.4 — driven purely by the set-version
// field, exactly as real KMIP payloads will rely on.
func TestV2FoundationVersionGating(t *testing.T) {
	when := time.Date(2024, 6, 16, 12, 34, 56, 789012000, time.UTC)

	t.Run("v2.0 encodes the gated field", func(t *testing.T) {
		data := marshalV2(t, &v2FoundationStruct{Version: kmip.V2_0, When: ttlv.DateTimeExtended(when)})
		assert.Contains(t, strings.ToUpper(hex.EncodeToString(data)), dtxFieldHeader,
			"Date-Time-Extended field must be on the wire at v2.0")

		got := unmarshalV2(t, data)
		assert.Equal(t, kmip.V2_0, got.Version)
		assert.Equal(t, when.UnixMicro(), time.Time(got.When).UnixMicro())
	})

	t.Run("v1.4 skips the gated field", func(t *testing.T) {
		data := marshalV2(t, &v2FoundationStruct{Version: kmip.V1_4, When: ttlv.DateTimeExtended(when)})
		assert.NotContains(t, strings.ToUpper(hex.EncodeToString(data)), dtxFieldHeader,
			"v2.0-gated field must not be emitted at v1.4")

		got := unmarshalV2(t, data)
		assert.Equal(t, kmip.V1_4, got.Version)
		assert.True(t, time.Time(got.When).IsZero(), "gated field must be absent at v1.4")
	})
}
