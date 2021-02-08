// Package settings implements HTTP/2 Settings values, to be included in a
// Settings Frame, informing each Peer of the Sender's configuration.
package settings

const (
	// HeaderTableSizeID (0x1) is the identifier for the
	// SETTINGS_HEADER_TABLE_SIZE setting.
	// RFC 7540 Section 6.5.2
	HeaderTableSizeID = uint16(0x1)

	// EnablePushID (0x2) is the identifier for the SETTINGS_ENABLE_PUSH
	// setting.
	// RFC 7540 Section 6.5.2
	EnablePushID = uint16(0x2)

	// MaxConcurrentStreamsID (0x3) is the identifier for the
	// SETTINGS_MAX_CONCURRENT_STREAMS setting.
	// RFC 7540 Section 6.5.2
	MaxConcurrentStreamsID = uint16(0x3)

	// InitialWindowSizeID (0x4) is the identifier for the
	// SETTINGS_INITIAL_WINDOW_SIZE setting.
	// RFC 7540 Section 6.5.2
	InitialWindowSizeID = uint16(0x4)

	// MaxFrameSizeID (0x5) is the identifier for the SETTINGS_MAX_FRAME_SIZE
	// setting.
	// RFC 7540 Section 6.5.2
	MaxFrameSizeID = uint16(0x5)

	// MaxHeaderListSizeID (0x6) is the identifier for the
	// SETTINGS_MAX_HEADER_LIST_SIZE setting.
	// RFC 7540 Section 6.5.2
	MaxHeaderListSizeID = uint16(0x6)
)

// Setting is implemented by types that contain connection-level configuration
// values to be shared between peers.
// RFC 7540 Section 6.5.1
type Setting interface {
	// ID is the unique identifier given to each Setting. Values greater than
	// 0x6 are extensions and MUST be ignored if not understood.
	ID() uint16

	// Value is the 4-byte (32-bit) value of a Setting.
	Value() uint32
}

// HeaderTableSize informs a peer of the maximum size of the header compression
// table used to decode header blocks, in bytes.
// RFC 7540 Section 6.5.2
type HeaderTableSize struct {
	Size uint32
}

// ID implements Setting.
func (h HeaderTableSize) ID() uint16 {
	return HeaderTableSizeID
}

// Value implements Setting.
func (h HeaderTableSize) Value() uint32 {
	return h.Size
}

// EnablePush informs a peer if the sender can receive PushPromise Frames.
// RFC 7540 Section 6.5.2
type EnablePush struct {
	Enabled bool
}

// ID implements Setting.
func (e EnablePush) ID() uint16 {
	return EnablePushID
}

// Value implements Setting.
func (e EnablePush) Value() uint32 {
	if e.Enabled {
		return 1
	}

	return 0
}

// MaxConcurrentStreams limits the maximum number of concurrent active streams.
// A value of Zero (0) is NOT special, and should only be used to reject new
// streams.
// RFC 7540 Section 6.5.2
type MaxConcurrentStreams struct {
	Streams uint32
}

// ID implements Setting.
func (m MaxConcurrentStreams) ID() uint16 {
	return MaxConcurrentStreamsID
}

// Value implements Setting.
func (m MaxConcurrentStreams) Value() uint32 {
	return m.Streams
}

// InitialWindowSize indicates to a peer the initial window size of the sender,
// which may be updated later with a WindowUpdate Frame, in bytes.
// RFC 7540 Section 6.5.2
type InitialWindowSize struct {
	Size uint32
}

// ID implements Setting.
func (i InitialWindowSize) ID() uint16 {
	return InitialWindowSizeID
}

// Value implements Setting.
func (i InitialWindowSize) Value() uint32 {
	return i.Size
}

// MaxFrameSize indicates to a peer the maximum size of a payload to accept.
// RFC 7540 Section 6.5.2
type MaxFrameSize struct {
	Size uint32
}

// ID implements Setting.
func (m MaxFrameSize) ID() uint16 {
	return MaxFrameSizeID
}

// Value implements Setting.
func (m MaxFrameSize) Value() uint32 {
	return m.Size
}

// MaxHeaderListSize advises the maximum size of the uncompressed header list
// the sender will accept, in bytes.
// RFC 7540 Section 6.5.2
type MaxHeaderListSize struct {
	Size uint32
}

// ID implements Setting.
func (m MaxHeaderListSize) ID() uint16 {
	return MaxHeaderListSizeID
}

// Value implements Setting.
func (m MaxHeaderListSize) Value() uint32 {
	return m.Size
}
