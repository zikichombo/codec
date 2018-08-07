// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package wav

import (
	"encoding/binary"
	"os"

	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

// Encoder encapsulates state for encoding a (pcm) wav file.
type Encoder struct {
	w   *os.File
	h   *hdr
	f   *Format
	buf []byte
	p   int
	n   int64
	//eFunc func([]byte, float64)
}

// NewEncoder creates a new encoder with the specified
// format to the writer/seeker w.
func NewEncoder(f *Format, w *os.File) (*Encoder, error) {
	enc := &Encoder{w: w, f: f}
	h := &hdr{}
	enc.h = h
	if e := h.Write(w); e != nil {
		return nil, e
	}
	if e := f.Write(w); e != nil {
		return nil, e
	}
	d := &chunk{fourCc: _dat4Cc}
	if e := d.writeHdr(w); e != nil {
		return nil, e
	}
	enc.p = 0
	enc.buf = make([]byte, f.Bytes()*f.Channels()*1024)
	//ef := f.Encoder()
	//enc.eFunc = ef
	enc.w = w
	return enc, nil
}

var _e *Encoder
var _f sound.Sink = _e

// Codec Implements sound.Sink.
func (e *Encoder) Codec() sample.Codec {
	return e.f.Codec
}

// Channels implements sound.Sink.
func (e *Encoder) Channels() int {
	return e.f.Channels()
}

func (e *Encoder) SampleRate() freq.T {
	return e.f.freq
}

// Put adds the sound sample s to the encoded stream, returning
// an error if there is a problem.
func (e *Encoder) put(s float64) error {
	bd := int(e.f.Codec.Bytes())
	//e.eFunc(e.buf[e.p:e.p+bd], s)
	e.f.Codec.Encode(e.buf[e.p:e.p+bd], []float64{s})
	e.n++
	e.p += bd
	if e.p == len(e.buf) {
		_, err := e.w.Write(e.buf)
		//fmt.Printf("encoder writing buffer\n")
		if err != nil {
			return err
		}
		e.p = 0
	}
	return nil
}

func (e *Encoder) Send(src []float64) error {
	nC := e.Channels()
	if len(src)%nC != 0 {
		return sound.ChannelAlignmentError
	}
	nF := len(src) / nC
	var err error
	var c, f int
	var v float64
	for i := range src {
		_ = i
		v = src[c*nF+f]
		if err = e.put(v); err != nil {
			return err
		}
		c++
		if c == nC {
			c = 0
			f++
		}
	}
	return nil
}

// Close closes the encoder and underlying file,
// returning an error if there is a problem.
//
// For wav files, the end of an encoding stream requires seeking and
// writing meta data in the headers.
func (e *Encoder) Close() error {
	if e.p != 0 {
		_, err := e.w.Write(e.buf[:e.p])
		if err != nil {
			return err
		}
	}
	audioBytes := e.n * int64(e.f.Bytes())
	_, err := e.w.Seek(4, os.SEEK_SET)
	if err != nil {
		return err
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(audioBytes)+uint32(e.f.chunkSize())+uint32(chunkHdrSize))
	_, err = e.w.Write(buf)
	if err != nil {
		return err
	}
	_, err = e.w.Seek(hdrChunkSize+int64(e.f.chunkSize())+chunkHdrSize-4, os.SEEK_SET)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(buf, uint32(audioBytes))
	_, err = e.w.Write(buf)
	if err != nil {
		return err
	}
	return e.w.Close()
}
