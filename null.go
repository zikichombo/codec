package codec

import (
	"bufio"
	"io"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

type nullCodec struct {
}

// NullCodec implements a codec that supports nothing.  It is useful
// for embedding in codec implementations that only support some of
// encoding/decoding functions.
var NullCodec Codec = nullCodec{}

func (c nullCodec) Extensions() []string {
	return nil
}

func (c nullCodec) Sniff(*bufio.Reader) bool {
	return false
}

func (c nullCodec) DefaultSampleCodec() sample.Codec {
	return AnySampleCodec
}

func (c nullCodec) Decoder() func(io.ReadCloser) (sound.Source, sample.Codec, error) {
	return nil
}

func (c nullCodec) SeekingDecoder() func(IoReadSeekCloser) (sound.SourceSeeker, sample.Codec, error) {
	return nil
}

func (c nullCodec) Encoder() func(w io.WriteCloser, c sample.Codec) (sound.Sink, error) {
	return nil
}

func (c nullCodec) RandomAccess() func(ws IoReadWriteSeekCloser, c sample.Codec) (sound.RandomAccess, error) {
	return nil
}
