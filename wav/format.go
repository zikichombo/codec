// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package wav

import (
	"encoding/binary"
	"fmt"
	"io"

	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

const (
	_TAG_PCM     = 1
	_TAG_FLOAT32 = 3
)

// Format describes a format wav chunk (simplified for only the PCM case).
type Format struct {
	sample.Codec
	channels int
	freq     freq.T
}

func (f *Format) String() string {
	return fmt.Sprintf(`Samples: %s
Channels: %d
SampleRate: %s
`, f.Codec, f.channels, f.freq)
}

// Channels returns the number of channels in the data.
func (f *Format) Channels() int {
	return f.channels
}

// SampleRate returns the sample frequency in Hertz.
func (f *Format) SampleRate() freq.T {
	return f.freq
}

// NewFormat creates a new Format with chans channels at frequency freq
// using sample codec sc.
func NewFormat(chans int, freq freq.T, sc sample.Codec) *Format {
	return &Format{
		channels: chans,
		freq:     freq,
		Codec:    sc}
}

func NewFormatForm(v sound.Form, sc sample.Codec) *Format {
	return NewFormat(v.Channels(), v.SampleRate(), sc)
}

// NewMonoFmt returns a common mono format of
// 16 bit depth at 44.1 KHz.
func NewMonoFmt() *Format {
	return &Format{
		channels: 1,
		freq:     44100 * freq.Hertz,
		Codec:    sample.SInt16L}
}

// NewStereoFmt returns a common stereo format
// of 16 bit depth at 44.1 KHz.
func NewStereoFmt() *Format {
	return &Format{
		channels: 2,
		freq:     44100 * freq.Hertz,
		Codec:    sample.SInt16L}
}

// The total size of a Format Chunk (for PCM case)
// other cases can add more
const fmtStartChunkSize = 2 + 2 + 4 + 4 + 2 + 2

func (f *Format) chunkSize() int {
	res := 4 + 4 + 2 + 2 + 4 + 4 + 2 + 2
	if f.Codec.IsFloat() {
		res += 2
	}
	return res
}

// ParseFormat parses a format from a reader, returning a non-nil error
// if there is a problem.
func ParseFormat(r io.Reader, N int) (*Format, error) {
	if N < fmtStartChunkSize {
		return nil, fmt.Errorf("format chunk too small: %d", N)
	}
	buf := make([]byte, N)
	n, e := r.Read(buf)
	if e != nil && e != io.EOF {
		return nil, e
	}
	if n != N {
		return nil, fmt.Errorf("only read %d/%d bytes of format", n, N)
	}
	tag := binary.LittleEndian.Uint16(buf[:2])
	if tag != _TAG_PCM && tag != _TAG_FLOAT32 {
		return nil, fmt.Errorf("tag isn't for PCM wav data: %d", tag)
	}
	channels := int(binary.LittleEndian.Uint16(buf[2:4]))
	frq := int(binary.LittleEndian.Uint32(buf[4:8]))
	bps := binary.LittleEndian.Uint32(buf[8:12])
	block := int(binary.LittleEndian.Uint16(buf[12:14]))
	bitDepth := binary.LittleEndian.Uint16(buf[14:16])
	if bps != uint32(frq)*uint32(block) {
		return nil, fmt.Errorf("bytes per sec is %d not %d", bps, frq*block)
	}
	if block*8 != int(bitDepth)*channels {
		return nil, fmt.Errorf("block align %d != %d", block, int(bitDepth)*channels/8)
	}
	aFreq := freq.T(frq) * freq.Hertz
	if tag == _TAG_FLOAT32 {
		if N != fmtStartChunkSize+2 {
			//return nil, fmt.Errorf("warning, wav format chunk too short but has full Float32 spec\n")
		}
		return &Format{channels: channels, freq: aFreq, Codec: sample.SFloat32L}, nil
	}
	if tag != _TAG_PCM {
		return nil, fmt.Errorf("unsupported format tag: %d", tag)
	}
	if N != fmtStartChunkSize {
		return nil, fmt.Errorf("bad format chunk size: %d", N)
	}
	var f *Format
	switch bitDepth {
	case 8:
		f = &Format{channels: channels, freq: aFreq, Codec: sample.SByte}
	case 16:
		f = &Format{channels: channels, freq: aFreq, Codec: sample.SInt16L}
	case 24:
		f = &Format{channels: channels, freq: aFreq, Codec: sample.SInt24L}
	case 32:
		f = &Format{channels: channels, freq: aFreq, Codec: sample.SInt32L}
	default:
		return nil, fmt.Errorf("unsupported bit depth: %d", bitDepth)
	}
	return f, nil
}

// Write writes a wav format chunk to a writer, returning an error if there is an
// IO error.
func (f *Format) Write(w io.Writer) error {
	buf := make([]byte, fmtStartChunkSize+8, f.chunkSize())
	buf[0] = 'f'
	buf[1] = 'm'
	buf[2] = 't'
	buf[3] = ' '
	freq := uint32(f.freq / freq.Hertz)
	tag := _TAG_PCM
	if f.Codec.IsFloat() {
		buf = buf[:f.chunkSize()]
		tag = _TAG_FLOAT32
	}
	binary.LittleEndian.PutUint32(buf[4:8], uint32(f.chunkSize()-8))
	binary.LittleEndian.PutUint16(buf[8:10], uint16(tag))
	binary.LittleEndian.PutUint16(buf[10:12], uint16(f.channels))
	binary.LittleEndian.PutUint32(buf[12:16], freq)
	bpspc := f.Bytes()
	binary.LittleEndian.PutUint32(buf[16:20], freq*uint32(f.channels)*uint32(bpspc))
	binary.LittleEndian.PutUint16(buf[20:22], uint16(f.channels)*uint16(bpspc))
	binary.LittleEndian.PutUint16(buf[22:24], uint16(f.Bits()))
	if tag == _TAG_FLOAT32 {
		binary.LittleEndian.PutUint16(buf[24:26], uint16(0))
	}
	n, e := w.Write(buf)
	if e != nil {
		return e
	}
	if n != len(buf) {
		return fmt.Errorf("couldn't write all of hdr %d/%d bytes", n, len(buf))
	}
	return nil
}
