// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package wav

import (
	"io"
	"os"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

func FormFormat(v sound.Form, codec sample.Codec) *Format {
	return NewFormat(v.Channels(), v.SampleRate(), codec)
}

func Save(src sound.Source, path string) error {
	format := FormFormat(src, sample.SFloat32L)
	f, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		return e
	}
	defer f.Close()
	enc, e := NewEncoder(format, f)
	if e != nil {
		return e
	}
	buf := make([]float64, 1024*src.Channels())
	var err error
	n := 0
	for {
		n, err = src.Receive(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = enc.Send(buf[:n*src.Channels()])
		if err != nil {
			return err
		}
	}
	return enc.Close()
}

func Load(path string) (*Decoder, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	dec, e := NewDecoder(f)
	if e != nil {
		return nil, e
	}
	return dec, e
}
