package ttlv

import "time"

// DateTimeExtended represents a KMIP Date-Time-Extended value: a POSIX timestamp
// with microsecond precision. On the binary wire it is encoded as an 8-byte value
// of type [TypeDateTimeExtended] (0x0B), holding the number of microseconds since
// the Unix epoch. It was introduced in KMIP 2.0.
//
// It is defined as a distinct Go type (rather than reusing [time.Time], which maps
// to the second-precision [TypeDateTime]) so the codec can select the correct wire
// type from a struct field's Go type alone.
type DateTimeExtended time.Time

// NewDateTimeExtended returns a DateTimeExtended for the given time.
func NewDateTimeExtended(t time.Time) DateTimeExtended {
	return DateTimeExtended(t)
}

// Time returns the value as a standard [time.Time].
func (d DateTimeExtended) Time() time.Time {
	return time.Time(d)
}

// String returns an RFC3339 representation with microsecond precision.
func (d DateTimeExtended) String() string {
	return time.Time(d).Format(time.RFC3339Nano)
}
