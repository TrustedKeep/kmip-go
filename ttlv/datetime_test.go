package ttlv_test

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ovh/kmip-go/ttlv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dtxHolder struct {
	When ttlv.DateTimeExtended `ttlv:"DTX"`
}

func init() {
	ttlv.RegisterTag("DTX", 0x420125)
	ttlv.RegisterTag("DtxHolder", 0x420126, reflect.TypeFor[dtxHolder]())
}

// TestDateTimeExtendedWireFormat asserts the on-the-wire binary encoding of a
// Date-Time-Extended value: type byte 0x0B, an 8-byte length, and the value as
// a big-endian microsecond POSIX timestamp.
func TestDateTimeExtendedWireFormat(t *testing.T) {
	tm := time.Date(2021, 1, 1, 0, 0, 0, 123*int(time.Microsecond), time.UTC)
	micros := tm.UnixMicro()

	enc := ttlv.NewTTLVEncoder()
	enc.DateTimeExtended(0x420125, tm)
	got := strings.ToUpper(hex.EncodeToString(enc.Bytes()))

	// tag(3)=420125 | type(1)=0B | length(4)=00000008 | value(8)=micros
	want := "4201250B00000008" + strings.ToUpper(hex.EncodeToString(uint64be(uint64(micros))))
	assert.Equal(t, want, got)
}

func uint64be(v uint64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(v)
		v >>= 8
	}
	return b
}

// TestDateTimeExtendedRoundTrip exercises encode/decode across all binary/json/xml
// encodings and verifies microsecond precision survives the trip (a precision a
// plain second-granularity DateTime would truncate).
func TestDateTimeExtendedRoundTrip(t *testing.T) {
	// 789012000 ns = 789012 µs of sub-second precision.
	orig := dtxHolder{When: ttlv.DateTimeExtended(
		time.Date(2024, 6, 16, 12, 34, 56, 789012000, time.UTC),
	)}
	wantMicros := time.Time(orig.When).UnixMicro()

	for _, tc := range []struct {
		name      string
		marshal   func(any) []byte
		unmarshal func([]byte, any) error
	}{
		{"binary", ttlv.MarshalTTLV, ttlv.UnmarshalTTLV},
		{"json", ttlv.MarshalJSON, ttlv.UnmarshalJSON},
		{"xml", ttlv.MarshalXML, ttlv.UnmarshalXML},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.marshal(&orig)
			var got dtxHolder
			require.NoError(t, tc.unmarshal(data, &got))
			assert.Equal(t, wantMicros, time.Time(got.When).UnixMicro(),
				"microsecond precision must round-trip via %s", tc.name)
		})
	}
}
