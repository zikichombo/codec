// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package wav

import (
	"errors"
	"io"
	"time"

	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

// Decoder encapsulates state for decoding and seeking
// a (pcm) wav file.
type Decoder struct {
	fmt    *Format
	dChunk *chunk

	r    ReadSeekerCloser
	buf  []byte
	p    int // buffer index
	e    int // end of buffer (can be less than last bit of data)
	n    int // number of buffers preceding current one
	frms int // number of decoded frames
	nFrm int // number of frames
	//dFunc     func([]byte) float64
	byteDepth int // byte depth
}

type ReadSeekerCloser interface {
	io.ReadSeeker
	io.Closer
}

// NewDecoder creates a decoder from a wav file (seekable, readable).
func NewDecoder(r ReadSeekerCloser) (*Decoder, error) {
	riff, fcc, e := readRiff(r)
	if e != nil {
		return nil, e
	}
	if fcc != _wave4Cc {
		return nil, errors.New("not a wave file")
	}
	fc, err := riff.findChunk(r, _fmt4Cc)
	if err != nil {
		return nil, err
	}

	f, e := ParseFormat(r, fc.length)
	if e != nil {
		return nil, e
	}
	dc, e := riff.findChunk(r, _dat4Cc)
	if e != nil {
		return nil, e
	}
	nFrm := dc.length / (f.Bytes() * f.Channels())

	//df := f.Decoder()
	bd := int(f.Bytes())
	frameSize := f.Bytes() * f.Channels()
	bufSize := frameSize * 1024
	buf := make([]byte, bufSize)
	res := &Decoder{
		fmt:    f,
		dChunk: dc,
		r:      r,
		buf:    buf,
		p:      0,
		e:      0,
		n:      0,
		nFrm:   nFrm,
		//dFunc:     df,
		byteDepth: bd}
	return res, nil
}

// sound.Source methods

var _ sound.Source = (*Decoder)(nil)

func (d *Decoder) Codec() sample.Codec {
	return d.fmt.Codec
}

func (d *Decoder) SampleRate() freq.T {
	return d.fmt.SampleRate()
}

func (d *Decoder) Channels() int {
	return d.fmt.Channels()
}

func (d *Decoder) Receive(dst []float64) (int, error) {
	nC := d.Channels()
	if len(dst)%nC != 0 {
		return 0, sound.ErrChannelAlignment
	}
	var err error
	var c, f int
	nF := len(dst) / nC
	for i := range dst {
		dst[c*nF+f], err = d.sample()
		if err == io.EOF {
			if i == 0 {
				return 0, io.EOF
			}
			return f, nil
		}
		if err != nil {
			return 0, err
		}
		c++
		if c == nC {
			c = 0
			f++
		}
	}
	return f, nil
}

func (d *Decoder) sample() (float64, error) {
	p, rd := d.p>>1, d.p&1 == 1
	if !rd {
		n, e := d.r.Read(d.buf)
		if n > 0 {
			d.e = n
		} else if e == nil {
			panic("read with 0 bytes no error")
		} else {
			return 0, e
		}
	}
	// here we have successfully read sometime in the past.
	q := p + d.byteDepth
	var vs [1]float64
	d.Codec().Decode(vs[:], d.buf[p:q])
	s := vs[0]
	//s := d.dFunc(d.buf[p:q])
	//fmt.Printf("decoded %d from buffer[%d:%d]\n", s, p, q)
	if q == d.e {
		d.p = 0
		d.n++ // more precise: d.n += d.e and reinterpret d.n elsewhere?
	} else {
		d.p = (q << 1) | 1
	}
	d.frms++

	return s, nil
}

// sound.Seeker methods

func (d *Decoder) Pos() int64 {
	q := int64((d.p >> 1)) / d.bpf()
	o := int64(d.n*len(d.buf)) / d.bpf()
	return int64(q + o)
}

func (d *Decoder) When() time.Duration {
	p := d.Pos()
	return time.Duration(p) * d.fdur()
}

func (d *Decoder) Len() int64 {
	return int64(d.nFrm)
}

func (d *Decoder) Duration() time.Duration {
	return time.Duration(d.Len()) * d.fdur()
}

func (d *Decoder) Seek(f int64) error {
	bpf := d.bpf()
	fpb := int64(len(d.buf)) / bpf
	nBuf := f / fpb
	m := f % fpb
	rd := d.p&1 == 1

	if e := d.dChunk.Seek(d.r, nBuf*int64(len(d.buf))); e != nil {
		return e
	}
	if d.n == int(nBuf) && rd {
		d.p = int(((m * bpf) << 1) | 1)
		return nil
	}
	d.n = int(nBuf)
	d.p = int((m * bpf) << 1)
	return nil
}

func (d *Decoder) Close() error {
	return d.r.Close()
}

func (d *Decoder) SeekDur(dur time.Duration) error {
	return d.Seek(int64(dur / d.fdur()))
}

func (d *Decoder) bpf() int64 {
	return int64(d.fmt.Bytes()) * int64(d.fmt.Channels())
}

func (d *Decoder) fdur() time.Duration {
	return d.SampleRate().Period()
}
