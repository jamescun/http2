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
		Header   *Header
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
			&Header{Length: 18, Type: TypeSettings},
			[]byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x64, 0x00, 0x04, 0x40, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			hdr := new(Header)
			bytes, err := test.Settings.MarshalFrame(hdr)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Header, hdr)
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
		Header   *Header
		Bytes    []byte
		Settings *Settings
		Error    error
	}{
		{
			"nghttp2.org",
			&Header{Length: 18, Type: TypeSettings},
			[]byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x64, 0x00, 0x04, 0x40, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00},
			&Settings{
				Header: Header{Length: 18, Type: TypeSettings},
				Settings: []settings.Setting{
					settings.MaxConcurrentStreams{Streams: 100},
					settings.InitialWindowSize{Size: 1073741824},
					settings.EnablePush{Enabled: false},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			frame := new(Settings)
			err := frame.UnmarshalFrame(test.Header, test.Bytes)

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

func TestHeadersMarshalFrame(t *testing.T) {
	tests := []struct {
		Name    string
		Headers *Headers
		Header  *Header
		Bytes   []byte
		Error   error
	}{
		{
			"nghttp2.org",
			&Headers{Block: []byte{
				0x3f, 0xe1, 0x1f, 0x82, 0x04, 0x88, 0x62, 0x7b, 0x69, 0x1d,
				0x48, 0x5d, 0x3e, 0x53, 0x86, 0x41, 0x88, 0xaa, 0x69, 0xd2,
				0x9a, 0xc4, 0xb9, 0xec, 0x9b, 0x7a, 0x88, 0x25, 0xb6, 0x50,
				0xc3, 0xab, 0xb8, 0x15, 0xc1, 0x53, 0x03, 0x2a, 0x2f, 0x2a,
			}},
			&Header{Length: 40, Type: TypeHeaders},
			[]byte{
				0x3f, 0xe1, 0x1f, 0x82, 0x04, 0x88, 0x62, 0x7b, 0x69, 0x1d,
				0x48, 0x5d, 0x3e, 0x53, 0x86, 0x41, 0x88, 0xaa, 0x69, 0xd2,
				0x9a, 0xc4, 0xb9, 0xec, 0x9b, 0x7a, 0x88, 0x25, 0xb6, 0x50,
				0xc3, 0xab, 0xb8, 0x15, 0xc1, 0x53, 0x03, 0x2a, 0x2f, 0x2a,
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			hdr := new(Header)
			bytes, err := test.Headers.MarshalFrame(hdr)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Header, hdr)
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

func TestHeadersUnmarshalFrame(t *testing.T) {
	tests := []struct {
		Name    string
		Header  *Header
		Bytes   []byte
		Headers *Headers
		Error   error
	}{
		{
			"nghttp2.org",
			&Header{Length: 40, Type: TypeHeaders},
			[]byte{
				0x3f, 0xe1, 0x1f, 0x82, 0x04, 0x88, 0x62, 0x7b, 0x69, 0x1d,
				0x48, 0x5d, 0x3e, 0x53, 0x86, 0x41, 0x88, 0xaa, 0x69, 0xd2,
				0x9a, 0xc4, 0xb9, 0xec, 0x9b, 0x7a, 0x88, 0x25, 0xb6, 0x50,
				0xc3, 0xab, 0xb8, 0x15, 0xc1, 0x53, 0x03, 0x2a, 0x2f, 0x2a,
			},
			&Headers{
				Header: Header{Length: 40, Type: TypeHeaders},
				Block: []byte{
					0x3f, 0xe1, 0x1f, 0x82, 0x04, 0x88, 0x62, 0x7b, 0x69, 0x1d,
					0x48, 0x5d, 0x3e, 0x53, 0x86, 0x41, 0x88, 0xaa, 0x69, 0xd2,
					0x9a, 0xc4, 0xb9, 0xec, 0x9b, 0x7a, 0x88, 0x25, 0xb6, 0x50,
					0xc3, 0xab, 0xb8, 0x15, 0xc1, 0x53, 0x03, 0x2a, 0x2f, 0x2a,
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			frame := new(Headers)
			err := frame.UnmarshalFrame(test.Header, test.Bytes)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Headers, frame)
				}
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, test.Error, err)
				}
			}
		})
	}
}

func TestDatarMarshalFrame(t *testing.T) {
	tests := []struct {
		Name   string
		Data   *Data
		Header *Header
		Bytes  []byte
		Error  error
	}{
		{
			"nghttp2.org",
			&Data{Data: []byte("User-agent: *\nDisallow: \n\nSitemap: //nghttp2.org/sitemap.xml \n")},
			&Header{Length: 62, Type: TypeData},
			[]byte("User-agent: *\nDisallow: \n\nSitemap: //nghttp2.org/sitemap.xml \n"),
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			hdr := new(Header)
			bytes, err := test.Data.MarshalFrame(hdr)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Header, hdr)
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

func TestDataUnmarshalFrame(t *testing.T) {
	tests := []struct {
		Name   string
		Header *Header
		Bytes  []byte
		Data   *Data
		Error  error
	}{
		{
			"nghttp2.org",
			&Header{Length: 62, Type: TypeData},
			[]byte("User-agent: *\nDisallow: \n\nSitemap: //nghttp2.org/sitemap.xml \n"),
			&Data{
				Header: Header{Length: 62, Type: TypeData},
				Data:   []byte("User-agent: *\nDisallow: \n\nSitemap: //nghttp2.org/sitemap.xml \n"),
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			frame := new(Data)
			err := frame.UnmarshalFrame(test.Header, test.Bytes)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Data, frame)
				}
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, test.Error, err)
				}
			}
		})
	}
}
