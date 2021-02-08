package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendSetting(t *testing.T) {
	tests := []struct {
		Name    string
		Source  []byte
		Setting Setting
		Result  []byte
	}{
		{"Nil", nil, nil, nil},
		{"HeaderTableSize", nil, HeaderTableSize{Size: 4096}, []byte{0x00, 0x01, 0x00, 0x00, 0x10, 0x00}},
		{"EnablePush", nil, EnablePush{Enabled: true}, []byte{0x00, 0x02, 0x00, 0x00, 0x00, 0x01}},
		{"MaxConcurrentStreams", nil, MaxConcurrentStreams{Streams: 100}, []byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x64}},
		{"InitialWindowSize", nil, InitialWindowSize{Size: 65535}, []byte{0x00, 0x04, 0x00, 0x00, 0xFF, 0xFF}},
		{"MaxFrameSize", nil, MaxFrameSize{Size: 16384}, []byte{0x00, 0x05, 0x00, 0x00, 0x40, 0x00}},
		{"MaxHeaderListSize", nil, MaxHeaderListSize{Size: 65535}, []byte{0x00, 0x06, 0x00, 0x00, 0xFF, 0xFF}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := AppendSetting(test.Source, test.Setting)

			assert.Equal(t, test.Result, result)
		})
	}
}

func TestParseSetting(t *testing.T) {
	tests := []struct {
		Name    string
		Source  []byte
		Setting Setting
		Error   error
	}{
		{"Nil", nil, nil, ErrInvalid},
		{"Empty", []byte{}, nil, ErrInvalid},
		{"HeaderTableSize", []byte{0x00, 0x01, 0x00, 0x00, 0x10, 0x00}, HeaderTableSize{Size: 4096}, nil},
		{"EnablePush", []byte{0x00, 0x02, 0x00, 0x00, 0x00, 0x01}, EnablePush{Enabled: true}, nil},
		{"MaxConcurrentStreams", []byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x64}, MaxConcurrentStreams{Streams: 100}, nil},
		{"InitialWindowSize", []byte{0x00, 0x04, 0x00, 0x00, 0xFF, 0xFF}, InitialWindowSize{Size: 65535}, nil},
		{"MaxFrameSize", []byte{0x00, 0x05, 0x00, 0x00, 0x40, 0x00}, MaxFrameSize{Size: 16384}, nil},
		{"MaxHeaderListSize", []byte{0x00, 0x06, 0x00, 0x00, 0xFF, 0xFF}, MaxHeaderListSize{Size: 65535}, nil},
		{"Unknown", []byte{0xFF, 0xFF, 0x00, 0x00, 0x00, 0x01}, nil, ErrUnknown},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			setting, err := ParseSetting(test.Source)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Setting, setting)
				}
			} else {
				if assert.Error(t, err) {
					assert.Nil(t, setting)
					assert.Equal(t, test.Error, err)
				}
			}
		})
	}
}
