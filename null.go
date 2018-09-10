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

func (c nullCodec) Decoder(io.ReadCloser) (sound.Source, sample.Codec, error) {
	return nil, AnySampleCodec, ErrUnsupportedFunction
}

func (c nullCodec) SeekingDecoder(IoReadSeekCloser) (sound.SourceSeeker, sample.Codec, error) {
	return nil, AnySampleCodec, ErrUnsupportedFunction
}

func (_ nullCodec) Encoder(w io.WriteCloser, c sample.Codec) (sound.Sink, error) {
	return nil, ErrUnsupportedFunction
}

func (_ nullCodec) RandomAccess(ws IoReadWriteSeekCloser, c sample.Codec) (sound.RandomAccess, error) {
	return nil, ErrUnsupportedFunction
}
