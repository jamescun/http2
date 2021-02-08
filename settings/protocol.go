package settings

import (
	"errors"
)

var (
	// ErrInvalid is returned when parsing a Setting but not enough bytes have
	// been supplied.
	ErrInvalid = errors.New("settings: short")

	// ErrUnknown is returned when parsing a Setting but its identifier is
	// not supported by this package. Receivers MUST ignore unknown settings.
	ErrUnknown = errors.New("settings: unknown identifier")
)

// AppendSetting marshals a Setting to the wire format, appends it to b and
// returns the extended buffer.
// RFC 7540 Section 6.5.1
func AppendSetting(b []byte, setting Setting) []byte {
	if setting == nil {
		return b
	}

	k, v := setting.ID(), setting.Value()

	return append(b,
		byte(k>>8),
		byte(k),
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v),
	)
}

// ParseSetting unmarshals a Setting from the wire format. ErrUnknown is
// returned if a Setting with an unknown identifier is encountered, receivers
// MUST ignore unknown settings.
// RFC 7540 Section 6.5.1
func ParseSetting(b []byte) (Setting, error) {
	if len(b) < 6 {
		return nil, ErrInvalid
	}

	k := uint16(b[1]) | uint16(b[0])<<8
	v := uint32(b[5]) | uint32(b[4])<<8 | uint32(b[3])<<16 | uint32(b[2])<<24

	switch k {
	case HeaderTableSizeID:
		return HeaderTableSize{Size: v}, nil

	case EnablePushID:
		return EnablePush{Enabled: v > 0}, nil

	case MaxConcurrentStreamsID:
		return MaxConcurrentStreams{Streams: v}, nil

	case InitialWindowSizeID:
		return InitialWindowSize{Size: v}, nil

	case MaxFrameSizeID:
		return MaxFrameSize{Size: v}, nil

	case MaxHeaderListSizeID:
		return MaxHeaderListSize{Size: v}, nil

	default:
		return nil, ErrUnknown
	}
}
