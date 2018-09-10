package codec

import (
	"bufio"
	"errors"
	"io"
	"reflect"

	"zikichombo.org/sound"
	"zikichombo.org/sound/ops"
	"zikichombo.org/sound/sample"
)

// ErrUnknownCodec is an error representing that a codec is unknown.
var ErrUnknownCodec = errors.New("unknown codec")

// ErrUnsupportedFunction is an error which is returned
// when a request is made of a Codec to perform some
// function it cannot (amongst decoding/encoding/seeking/random access).
var ErrUnsupportedFunction = errors.New("unsupported function")

// ErrUnsupportedSampleCodec can be used by codec implementations
// which receive a request for encoding or decoding with a sample codec
// that is unsupported or doesn't make sense.
var ErrUnsupportedSampleCodec = errors.New("unsupported sample codec")

// AnySampleCodec is a sample.Codec which is used as a wildcard sample.Codec
// in this interface.
var AnySampleCodec = sample.Codec(-1)

// IoReadSeekCloser just wraps io.ReadSeeker and io.Closer, as a convenience
// for specifying a decoding function which can also seek (codec functions
// always have a Close())
type IoReadSeekCloser interface {
	io.ReadSeeker
	io.Closer
}

// IoReadWriteSeekCloser wraps io.Reader, io.Writer, io.Seeker, and
// io.Closer as a convenience for specifying a codec function for
// sound.RandomAccess.
type IoReadWriteSeekCloser interface {
	IoReadSeekCloser
	io.Writer
}

// Codec represents a way of encoding and decoding sound.
type Codec interface {
	// Extensions lists the filename extensions which this codec claims to support.
	// Examples are .wav, .ogg, .caf.  The extension string includes the leading '.'.
	//
	// The returned slice should be read-only.
	Extensions() []string

	// Sniff is a function which when provided with a *bufio.Reader r, may
	// call r.Peek(), and only r.Peek().  Sniff should return true only if
	// this codec has a Decoder function for the data based on the data
	// aquired via r.Peek().
	Sniff(*bufio.Reader) bool

	// DefaultSampleCodec gives a default or preferred sample codec for the codec.
	// Some codecs, such as perception based compressed codecs, don't really have
	// a defined sample.Codec and should use AnySampleCodec for this field. Codecs
	// which only have decoding capabilities should also have AnySampleCodec in this
	// field.
	DefaultSampleCodec() sample.Codec

	// Decoder tries to turn an io.ReadCloser into a sound.Source.  In the event
	// the decoder does not use a single defined sample.Codec during the entire
	// decoding process for the resulting sound.Source, then the second return
	// value should be AnySampleCodec (if the error return value is nil).
	Decoder() func(io.ReadCloser) (sound.Source, sample.Codec, error)

	// SeekingDecoder is exactly like Decoder but returns a sound.SourceSeeker
	// given an io.ReadSeekClose.
	SeekingDecoder() func(IoReadSeekCloser) (sound.SourceSeeker, sample.Codec, error)

	// Encoder tries to turn an io.WriteCloser into a sound.Sink.
	// The sample.Codec argument can specify the desired sample Codec.
	// For encodings which don't use a defined sample.Codec, the function
	// should return (nil, ErrUnsupportedSampleCodec) in the event c
	// is not AnySampleCodec.
	Encoder() func(w io.WriteCloser, c sample.Codec) (sound.Sink, error)

	// RandomAccess tries to turn an io.ReadWriteSeeker into sound.RandomAccess.
	// If the codec does not make use of a defined sample.Codec and c is
	// not AnySampleCodec, then the function should return (nil, ErrUnsupporteSampleCodec).
	RandomAccess() func(ws IoReadWriteSeekCloser, c sample.Codec) (sound.RandomAccess, error)
}

type codec struct {
	Codec

	// pkgPath is the package path of the codec functions above.  It is populated
	// by RegisterCodec().  RegisterCodec() will only succeed if all non-nil codec
	// functions have the same package path.  CodecFor() allows callers to select
	// Codecs by pkgPath in the case of conflicts when there are multiple codecs
	// available.
	pkgPath string
}

var codecs []codec

func getPkgPath(v interface{}) string {
	typ := reflect.ValueOf(v).Type()
	return typ.PkgPath()
}

// RegisterCodec registers a codec so that consumers of this package
// can treat sound I/O generically and switch between codecs.
//
// A package "p" implementing a Codec can register a codec in its init()
// function.
func RegisterCodec(c Codec) {
	codecs = append(codecs, codec{
		Codec:   c,
		pkgPath: getPkgPath(c)})
}

// CodecFor tries to find a codec based on a filename extension.
//
// The returned codec, although a pointer to a struct with fields, should be
// treated as read-only.  Not doing so may result in race conditions or worse.
//
// The function pkgSel may be used to filter or select packages implementing
// codecs.  If the supplied value is nil, then by default the behavior
// is as if the function body were "return true".  As multiple codec implementations
// may exist, the first codec whose package path p is such that pkgSel(p) is true
// will be returned.
func CodecFor(ext string, pkgSel func(string) bool) (Codec, error) {
	for i := range codecs {
		c := &codecs[i]
		for _, codExt := range c.Extensions() {
			if ext == codExt {
				if pkgSel == nil || pkgSel(c.pkgPath) {
					return c.Codec, nil
				}
			}
		}
	}
	return nil, ErrUnknownCodec
}

// Decoder tries to turn an io.ReadCloser into a sound.Source.  If it fails, it
// returns a non-nil error.
//
// The function pkgSel may be used to filter or select packages implementing
// codecs.  If the supplied value is nil, then by default the behavior is as if
// the function body were "return true".  As multiple codec implementations may
// exist, the first codec whose package path p is such that pkgSel(p) is true
// will be returned.
//
// If it succeeds, it also returns a sample.Codec which may either be:
//
// 1. AnySampleCodec, indicating there is no fixed sample codec for the decoder
// behind sound.Source; or
//
// 2. A sample.Codec which defined the data received in sound.Source.
func Decoder(r io.ReadCloser, pkgSel func(string) bool) (sound.Source, sample.Codec, error) {
	theCodec, rr := sniff(r, pkgSel)
	if theCodec == nil {
		return nil, AnySampleCodec, ErrUnknownCodec
	}
	dec := theCodec.Decoder()
	if dec == nil {
		return nil, AnySampleCodec, ErrUnsupportedFunction
	}
	return dec(rr)
}

// SeekingDecoder is exactly like Decoder with respect to all arguments and
// functionality with the following exceptions
//
// 1. The io.ReadCloser must be seekable.
//
// 2. It returns a sound.SourceSeeker rather than a sound.Source.
func SeekingDecoder(r IoReadSeekCloser, pkgSel func(string) bool) (sound.SourceSeeker, sample.Codec, error) {
	theCodec, _ := sniff(r, pkgSel)
	if theCodec == nil {
		return nil, AnySampleCodec, ErrUnknownCodec
	}
	dec := theCodec.SeekingDecoder()
	if dec == nil {
		return nil, AnySampleCodec, ErrUnsupportedFunction
	}
	r.Seek(0, io.SeekStart)
	return dec(r)
}

type brCloser struct {
	*bufio.Reader
	io.Closer
}

func sniff(r io.ReadCloser, pkgSel func(string) bool) (Codec, *brCloser) {
	br := bufio.NewReader(r)
	var theCodec Codec
	for i := range codecs {
		c := &codecs[i]
		if c.Sniff != nil && c.Sniff(br) && (pkgSel == nil || pkgSel(c.pkgPath)) {
			theCodec = c.Codec
		}
	}
	return theCodec, &brCloser{Reader: br, Closer: r}
}

// Encoder tries to turn an io.WriteCloser into a sound.Sink
// given a filename extension.
func Encoder(dst io.WriteCloser, ext string) (sound.Sink, error) {
	co, err := CodecFor(ext, nil)
	if err != nil {
		return nil, err
	}
	enc := co.Encoder()
	if enc == nil {
		return nil, ErrUnsupportedFunction
	}
	return enc(dst, co.DefaultSampleCodec())
}

// EncoderWith tries to turn an io.WriteCloser into a sound.Sink
// given a filename extension and a sample.Codec.
//
// The sample codec may be AnySampleCodec, which should be used when
// the caller is not sure of the desired sample codec c.
func EncoderWith(dst io.WriteCloser, ext string, c sample.Codec) (sound.Sink, error) {
	co, err := CodecFor(ext, nil)
	if err != nil {
		return nil, err
	}
	enc := co.Encoder()
	if enc == nil {
		return nil, ErrUnsupportedFunction
	}
	return enc(dst, c)
}

// Encode encodes a sound.Source to an io.WriteCloser, selecting
// the codec based on a filename extension ext. It returns any
// error that may have been encountered in that process.
func Encode(dst io.WriteCloser, src sound.Source, ext string) error {
	return EncodeWith(dst, src, ext, AnySampleCodec)
}

// EncodeWith encodes a sound.Source to an io.WriteCloser, selecting
// the codec based on a filename extension ext and desired sample codec co.
//
// EncodeWith returns any error that may have been encountered in that process.
func EncodeWith(dst io.WriteCloser, src sound.Source, ext string, co sample.Codec) error {
	snk, err := EncoderWith(dst, ext, co)
	if err != nil {
		return err
	}
	return ops.Copy(snk, src)
}
