package frames

import (
	"testing"

	"github.com/jamescun/http2/settings"

	"github.com/stretchr/testify/assert"
)

func TestHeaderMarshalFrameHeader(t *testing.T) {
	tests := []struct {
		Name   string
		Header Header
		Bytes  []byte
		Error  error
	}{
		{"Empty", Header{}, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, nil},
		{"NoFlags", Header{Length: 8, Type: TypePing}, []byte{0x00, 0x00, 0x08, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00}, nil},
		{"Flags", Header{Length: 8, Type: TypePing, Flags: FlagPingAck}, []byte{0x00, 0x00, 0x08, 0x06, 0x01, 0x00, 0x00, 0x00, 0x00}, nil},
		{"StreamID", Header{Length: 4, Type: TypeResetStream, StreamID: 1}, []byte{0x00, 0x00, 0x04, 0x03, 0x00, 0x00, 0x00, 0x00, 0x01}, nil},
		{"LengthUint24", Header{Length: (1 << 25), Type: TypePing}, []byte{}, ErrFrameTooBig},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			bytes, err := test.Header.MarshalFrameHeader()

			if test.Error == nil {
				if assert.NoError(t, test.Error, err) {
					assert.Equal(t, test.Bytes, bytes)
				}
			} else {
				if assert.Error(t, err) {
					assert.Nil(t, bytes)
					assert.Equal(t, test.Error, err)
				}
			}
		})
	}
}

func TestHeaderUnmarshalFrameHeader(t *testing.T) {
	tests := []struct {
		Name   string
		Bytes  []byte
		Header *Header
		Error  error
	}{
		{"Nil", nil, nil, ErrShortHeader},
		{"Empty", []byte{}, nil, ErrShortHeader},
		{"NoFlags", []byte{0x00, 0x00, 0x08, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00}, &Header{Length: 8, Type: TypePing}, nil},
		{"Flags", []byte{0x00, 0x00, 0x08, 0x06, 0x01, 0x00, 0x00, 0x00, 0x00}, &Header{Length: 8, Type: TypePing, Flags: FlagPingAck}, nil},
		{"StreamID", []byte{0x00, 0x00, 0x04, 0x03, 0x00, 0x00, 0x00, 0x00, 0x01}, &Header{Length: 4, Type: TypeResetStream, StreamID: 1}, nil},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			header := new(Header)
			err := header.UnmarshalFrameHeader(test.Bytes)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Header, header)
				}
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, test.Error, err)
				}
			}
		})
	}
}

func TestSettingsMarshalFrame(t *testing.T) {
	tests := []struct {
		Name     string
		Settings *Settings
		Bytes    []byte
		Error    error
	}{
		{
			"nghttp2.org",
			&Settings{Settings: []settings.Setting{
				settings.MaxConcurrentStreams{Streams: 100},
				settings.InitialWindowSize{Size: 1073741824},
				settings.EnablePush{Enabled: false},
			}},
			[]byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x64, 0x00, 0x04, 0x40, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			bytes, err := test.Settings.MarshalFrame()

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Bytes, bytes)
				}
			} else {
				if assert.Error(t, err) {
					assert.Nil(t, bytes)
					assert.Equal(t, test.Error, err)
				}
			}
		})
	}
}

func TestSettingsUnmarshalFrame(t *testing.T) {
	tests := []struct {
		Name     string
		Bytes    []byte
		Settings *Settings
		Error    error
	}{
		{
			"nghttp2.org",
			[]byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x64, 0x00, 0x04, 0x40, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			&Settings{Settings: []settings.Setting{
				settings.MaxConcurrentStreams{Streams: 100},
				settings.InitialWindowSize{Size: 1073741824},
				settings.EnablePush{Enabled: false},
			}},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			frame := new(Settings)
			err := frame.UnmarshalFrame(test.Bytes)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Settings, frame)
				}
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, test.Error, err)
				}
			}
		})
	}
}
