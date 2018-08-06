// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// Copyright 2017 The IriFrance Audio Authors. All rights reserved.  Use of
// this source code is governed by a license that can be found in the License
// file.

package wav

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"zikichombo.org/sound/freq"
	"zikichombo.org/sound/sample"
)

func TestEncodeDecode(t *testing.T) {
	fmt := NewMonoFmt()
	encodeDecode(fmt, 2148, t)
}

func encodeDecode(ft *Format, N int, t *testing.T) {
	file, err := ioutil.TempFile(".", "wavtest")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()
	M := N * ft.Channels()
	d := make([]float64, M)
	fM := float64(M)
	for i := 0; i < M; i++ {
		d[i] = float64(i) / fM
	}
	//fmt.Printf("encoding...\n")
	if err := encode(d, ft, file); err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("done...\n")
	file, err = os.Open(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	dcd, err := NewDecoder(file)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("got decoder...\n")
	buf := make([]float64, M)
	n, e := dcd.Receive(buf)
	if e != nil {
		t.Fatal(e)
	}
	if n != N {
		t.Fatalf("only decoded %d/%d frames\n", n, N)
	}
	for i, v := range buf[:n*ft.Channels()] {
		inv := float64(i) / fM
		if math.Abs(v-inv) > 0.01 {
			t.Fatalf("%d: decoded %f expected %f\n", i, v, inv)
		}
	}
}

func TestStereo(t *testing.T) {
	fmt := NewStereoFmt()
	encodeDecode(fmt, 128, t)
}

func TestBitDepth(t *testing.T) {
	for _, sc := range []sample.Codec{sample.SByte, sample.SInt16L, sample.SInt24L, sample.SInt32L, sample.SFloat32L} {
		fmt := NewFormat(1, 44100, sc)
		encodeDecode(fmt, 128, t)
	}
}

func TestSeeker(t *testing.T) {
	f, err := ioutil.TempFile(".", "wavtest")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	format := NewMonoFmt()
	d := make([]float64, format.SampleRate()/freq.Hertz)
	if e := encode(d, format, f); e != nil {
		t.Fatal(e)
	}
	f, err = os.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	dec, err := NewDecoder(f)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Duration() < time.Second-time.Millisecond || dec.Duration() > time.Second+time.Millisecond {
		t.Errorf("duration %s != %s", dec.Duration(), time.Second)
	}
	if e := dec.Seek(22050); e != nil {
		t.Fatal(err)
	}
	//fmt.Printf("seek 22050, pos %d when: %s\n", dec.Pos(), dec.When())
	dec.SeekDur(dec.When())
	//fmt.Printf("after seekdur: pos %d when: %s\n", dec.Pos(), dec.When())
}

func TestSeekRead(t *testing.T) {
	N := 1024*2 + 760
	format := NewMonoFmt()
	f, e := ioutil.TempFile(".", "wavtest")
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	d := make([]float64, format.SampleRate()/freq.Hertz)
	for i := 0; i < N; i++ {
		d[i] = float64(i) / float64(N)
	}
	if e := encode(d, format, f); e != nil {
		t.Fatal(e)
	}
	f, e = os.Open(f.Name())
	if e != nil {
		t.Fatal(e)
	}
	dec, err := NewDecoder(f)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]float64, 1)
	for i := 0; i < 4096; i++ {
		frm := rand.Intn(N)
		if e := dec.Seek(int64(frm)); e != nil {
			t.Fatal(e)
		}
		if frm%2 == 0 {
			continue
		}
		s, e := dec.Receive(buf)
		if e != nil {
			t.Fatal(e)
		}
		if s != 1 {
			t.Fatalf("asked for 1 frame, got %d\n", s)
		}
		if math.Abs(buf[0]-float64(frm)/float64(N)) > 0.0001 {
			s := fmt.Sprintf("fmt %s error decode after seek: %f != %f\n", format, buf[0], float64(frm)/float64(N))
			t.Error(s)
		}
	}
}

func encode(d []float64, format *Format, f *os.File) error {
	enc, err := NewEncoder(format, f)
	if err != nil {
		return err
	}
	if format.Channels() < 1 {
		return fmt.Errorf("invalid number of channels: 0")
	}
	if err := enc.Send(d); err != nil {
		return err
	}
	return enc.Close()
}
