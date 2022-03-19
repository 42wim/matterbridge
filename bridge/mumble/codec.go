package bmumble

import (
	"fmt"

	"layeh.com/gumble/gumble"
)

// This is a dummy implementation of a Gumble audio codec which claims
// to implement Opus, but does not actually do anything.  This serves
// as a workaround until https://github.com/layeh/gumble/pull/61 is
// merged.
// See https://github.com/42wim/matterbridge/issues/1750 for details.

const (
	audioCodecIDOpus = 4
)

func registerNullCodecAsOpus() {
	codec := &NullCodec{
		encoder: &NullAudioEncoder{},
		decoder: &NullAudioDecoder{},
	}
	gumble.RegisterAudioCodec(audioCodecIDOpus, codec)
}

type NullCodec struct {
	encoder *NullAudioEncoder
	decoder *NullAudioDecoder
}

func (c *NullCodec) ID() int {
	return audioCodecIDOpus
}

func (c *NullCodec) NewEncoder() gumble.AudioEncoder {
	e := &NullAudioEncoder{}
	return e
}

func (c *NullCodec) NewDecoder() gumble.AudioDecoder {
	d := &NullAudioDecoder{}
	return d
}

type NullAudioEncoder struct{}

func (e *NullAudioEncoder) ID() int {
	return audioCodecIDOpus
}

func (e *NullAudioEncoder) Encode(pcm []int16, mframeSize, maxDataBytes int) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *NullAudioEncoder) Reset() {
}

type NullAudioDecoder struct{}

func (d *NullAudioDecoder) ID() int {
	return audioCodecIDOpus
}

func (d *NullAudioDecoder) Decode(data []byte, frameSize int) ([]int16, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *NullAudioDecoder) Reset() {
}
