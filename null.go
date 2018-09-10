package codec

import (
	"bufio"
	"io"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

// Type NullCodec is a type whose values implement Codec and
// support nothing.  It is useful for embedding codec implementations
// that only support some of the encoding/decoding functions.
//
type NullCodec struct {
}

func (c NullCodec) Extensions() []string {
	return nil
}

func (c NullCodec) Sniff(*bufio.Reader) bool {
	return false
}

func (c NullCodec) DefaultSampleCodec() sample.Codec {
	return AnySampleCodec
}

func (c NullCodec) Decoder(io.ReadCloser) (sound.Source, sample.Codec, error) {
	return nil, AnySampleCodec, ErrUnsupportedFunction
}

func (c NullCodec) SeekingDecoder(IoReadSeekCloser) (sound.SourceSeeker, sample.Codec, error) {
	return nil, AnySampleCodec, ErrUnsupportedFunction
}

func (_ NullCodec) Encoder(w io.WriteCloser, c sample.Codec) (sound.Sink, error) {
	return nil, ErrUnsupportedFunction
}

func (_ NullCodec) RandomAccess(ws IoReadWriteSeekCloser, c sample.Codec) (sound.RandomAccess, error) {
	return nil, ErrUnsupportedFunction
}
