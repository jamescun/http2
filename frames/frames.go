// Package frames implements HTTP/2 frames exchanged by peers as defined in
// RFC 7540 Section 6.
package frames

import (
	"errors"
	"fmt"

	"github.com/jamescun/http2/settings"
)

var (
	// ErrFrameTooBig is returned when attempting to marshal a Frame but its
	// configured length exceeds a uint24.
	ErrFrameTooBig = errors.New("frames: too big")

	// ErrShortHeader is returned when attempting to unmarshal a Header but
	// not enough bytes are available.
	ErrShortHeader = errors.New("frames: short header")

	// ErrShortFrame is returned when attempting to unmarshal a Frame but not
	// enough bytes are available.
	ErrShortFrame = errors.New("frames: too short")
)

// Frame is implemented by all HTTP/2 Frame definitions, as defined in RFC 7540
// Section 6.
type Frame interface {
	// MarshalFrame converts a Frame into it's wire format and set the Frame
	// specific fields on the given Header.
	MarshalFrame(*Header) ([]byte, error)

	// UnmarshalFrame converts a Frame from it's wire format. Header is the
	// Frame header that preceded the Frame being unmarshalled.
	UnmarshalFrame(*Header, []byte) error
}

// Type is the unique identifier given to each Frame. FrameTypes greater than
// 0x09 are considered extensions and MUST be ignored if not understood.
// RFC 7540 Section 4.1
type Type uint8

const (
	// TypeData (0x0) is defined by RFC 7540 Section 6.1.
	TypeData = Type(0x0)

	// TypeHeaders (0x1) is defined by RFC 7540 Section 6.2
	TypeHeaders = Type(0x1)

	// TypePriority (0x2) is defined by RFC 7540 Section 6.3.
	TypePriority = Type(0x2)

	// TypeResetStream (0x3) is defined by RFC 7540 Section 6.4.
	TypeResetStream = Type(0x3)

	// TypeSettings (0x4) is defined by RFC 7540 Section 6.5.
	TypeSettings = Type(0x4)

	// TypePushPromise (0x5) is defined by RFC 7540 Section 6.6.
	TypePushPromise = Type(0x5)

	// TypePing (0x6) is defined by RFC 7540 Section 6.7.
	TypePing = Type(0x6)

	// TypeGoAway (0x7) is defined by RFC 7540 Section 6.8.
	TypeGoAway = Type(0x7)

	// TypeWindowUpdate (0x8) is defined by RFC 7540 Section 6.9.
	TypeWindowUpdate = Type(0x8)

	// TypeContinuation (0x9) is defined by RFC 7540 Section 6.10.
	TypeContinuation = Type(0x9)
)

// Flags are Frame specific options set on the FrameHeader.
// RFC 7540 Section 4.1
type Flags uint8

// Set sets Flags v on Flags f.
func (f Flags) Set(v Flags) {
	f = f | v
}

// Has returns true if Flags f contains Flags v.
func (f Flags) Has(v Flags) bool {
	return f&v != 0
}

// HeaderLength is the fixed length of a Header Frame in bytes.
// RFC 7540 Section 4.1
const HeaderLength = 9

// Header prefixes all HTTP/2 payloads identifying Frame type, length,
// optional flags and its associated Stream.
// RFC 7540 Section 4.1
type Header struct {
	Length   uint32
	Type     Type
	Flags    Flags
	StreamID uint32
}

// MarshalFrameHeader marshals Header to the wire format.
func (h *Header) MarshalFrameHeader() ([]byte, error) {
	// NOTE(jc): Header contains a uint32 but the protocol demands a uint24,
	// unavailable in Go, throw ErrFrameTooBig if given >uint24.
	if h.Length >= (1 << 24) {
		return nil, ErrFrameTooBig
	}

	b := make([]byte, HeaderLength)

	putUint24(b, h.Length)
	b[3] = byte(h.Type)
	b[4] = byte(h.Flags)
	putUint31(b[5:], h.StreamID)

	return b, nil
}

// UnmarshalFrameHeader unmarshals a Header from the wire format.
func (h *Header) UnmarshalFrameHeader(b []byte) error {
	if len(b) < HeaderLength {
		return ErrShortHeader
	}

	h.Length = uint24(b)
	h.Type = Type(b[3])
	h.Flags = Flags(b[4])
	h.StreamID = uint31(b[5:])

	return nil
}

const (
	// FlagSettingsAck indicates a Settings frame is an acknowledgement of a
	// previously sent Settings frame.
	// RFC 7540 Section 6.5
	FlagSettingsAck = Flags(0x1)
)

// Settings conveys and acknowledges configration values between peers, it
// is it not used for negotiation.
// RFC 7540 Section 6.5
type Settings struct {
	Header

	// Ack acknowledges a previously sent Settings frame and MUST NOT contain
	// any Settings itself.
	Ack bool

	Settings []settings.Setting
}

// MarshalFrame marshals Settings into the wire format.
func (s *Settings) MarshalFrame(hdr *Header) ([]byte, error) {
	hdr.Length = uint32(6 * len(s.Settings))
	hdr.Type = TypeSettings
	hdr.StreamID = 0

	if s.Ack {
		hdr.Flags = FlagSettingsAck
	}

	b := make([]byte, 0, 6*len(s.Settings))

	for _, setting := range s.Settings {
		b = settings.AppendSetting(b, setting)
	}

	return b, nil
}

// UnmarshalFrame unmarshals Settings from the wire format.
func (s *Settings) UnmarshalFrame(hdr *Header, b []byte) error {
	if len(b) == 0 {
		return nil
	}

	// NOTE(jc): settings identifiers and values are always a multiple of six.
	if len(b)%6 != 0 {
		return ErrShortFrame
	}

	s.Header = *hdr
	s.Settings = make([]settings.Setting, 0, len(b)/6)

	for len(b) > 0 {
		setting, err := settings.ParseSetting(b)

		b = b[6:]

		if err == settings.ErrUnknown {
			continue
		} else if err != nil {
			return err
		}

		s.Settings = append(s.Settings, setting)
	}

	return nil
}

var (
	// FlagHeadersEndStream indicates a Headers frame is to also terminate its
	// Stream (excluding any Continuation frames).
	// RFC 7540 Section 6.2
	FlagHeadersEndStream = Flags(0x01)

	// FlagHeadersEndHeaders indicates a Headers frame is the last of the
	// Headers sent by a peer.
	// RFC 7540 Section 6.2
	FlagHeadersEndHeaders = Flags(0x04)

	// FlagHeadersPadded indicates a Headers frame contains trailing padding.
	// RFC 7540 Section 6.2
	FlagHeadersPadded = Flags(0x08)

	// FlagHeadersPriority indicates a Headers frame contains priority
	// information similar to a Priority frame.
	// RFC 7540 Section 6.2
	FlagHeadersPriority = Flags(0x20)
)

// Headers is used to initialize a Stream and contains zero or more HPACK
// header block fragments.
//
// NOTE(jc): Padding and Priority are not currently implemented.
//
// RFC 7540 Section 6.2
type Headers struct {
	Header

	// EndStream indicates this Header frame (and possible Continuation frames)
	// are the last in this Stream.
	EndStream bool

	// EndHeaders indicates this Headers frame is the last of this set and no
	// other Headers frame or Continuation frame will be sent.
	EndHeaders bool

	// Block contains an HPACK header block fragment, described in RFC 7541.
	Block []byte
}

// MarshalFrame marshals Headers into the wire format.
func (h *Headers) MarshalFrame(hdr *Header) ([]byte, error) {
	// TODO(jc): implement security padding and stream prioritization.
	if h.Header.Flags.Has(FlagHeadersPadded) {
		return nil, fmt.Errorf("headers: padding not implemented")
	} else if h.Header.Flags.Has(FlagHeadersPriority) {
		return nil, fmt.Errorf("headers: priority not implemented")
	}

	if h.EndStream {
		hdr.Flags.Set(FlagHeadersEndStream)
	}
	if h.EndHeaders {
		hdr.Flags.Set(FlagHeadersEndHeaders)
	}

	hdr.Type = TypeHeaders
	hdr.Length = uint32(len(h.Block))
	hdr.StreamID = h.StreamID

	b := make([]byte, len(h.Block))
	copy(b, h.Block)

	return b, nil
}

// UnmarshalFrame unmarshals Headers from the wire format.
func (h *Headers) UnmarshalFrame(hdr *Header, b []byte) error {
	// TODO(jc): implement security padding and stream prioritization from
	// initial Headers frame.
	if hdr.Flags.Has(FlagHeadersPadded) {
		return fmt.Errorf("headers: padding not implemented")
	} else if hdr.Flags.Has(FlagHeadersPriority) {
		return fmt.Errorf("headers: priority not implemented")
	}

	if hdr.Flags.Has(FlagHeadersEndStream) {
		h.EndStream = true
	}
	if hdr.Flags.Has(FlagHeadersEndHeaders) {
		h.EndHeaders = true
	}

	h.Header = *hdr
	h.Block = make([]byte, len(b))
	copy(h.Block, b)

	return nil
}

const (
	// FlagDataEndStream indicates a Data frame also terminates its Stream.
	// RFC 7540 Section 6.1
	FlagDataEndStream = Flags(0x01)

	// FlagDataPadded indicates a Data frame contains trailing padding.
	// RFC 7540 Section 6.1
	FlagDataPadded = Flags(0x08)
)

// Data is used to carry request or response data between peers.
//
// NOTE(jc): Padding is not currently implemented.
//
// RFC 7540 Section 6.1
type Data struct {
	Header

	// EndStream indicates this Data frame terminates the Stream.
	EndStream bool

	// Application data from peer.
	Data []byte
}

// MarshalFrame marshals Data into the wire format.
func (d *Data) MarshalFrame(hdr *Header) ([]byte, error) {
	// TODO(jc): implement security padding.
	if d.Header.Flags.Has(FlagDataPadded) {
		return nil, fmt.Errorf("data: padding not implemented")
	}

	if d.EndStream {
		hdr.Flags.Set(FlagDataEndStream)
	}

	hdr.Type = TypeData
	hdr.Length = uint32(len(d.Data))
	hdr.StreamID = d.Header.StreamID

	b := make([]byte, len(d.Data))
	copy(b, d.Data)

	return b, nil
}

// UnmarshalFrame unmarshals Data from the wire format.
func (d *Data) UnmarshalFrame(hdr *Header, b []byte) error {
	// TODO(jc): implement security padding.
	if hdr.Flags.Has(FlagDataPadded) {
		return fmt.Errorf("data: padding not implemented")
	}

	if hdr.Flags.Has(FlagDataEndStream) {
		d.EndStream = true
	}

	d.Header = *hdr
	d.Data = make([]byte, len(b))
	copy(d.Data, b)

	return nil
}

const (
	// FlagPingAck indicates a Ping Frame is an acknowledgement of received
	// Ping Frame.
	// RFC 7540 Section 6.7
	FlagPingAck = Flags(0x1)
)

func uint24(b []byte) uint32 {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

func putUint24(b []byte, v uint32) {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func uint31(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return (uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24) & (1<<31 - 1)
}

func putUint31(b []byte, v uint32) {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}
