// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// Copyright 2017 The IriFrance Audio Authors. All rights reserved.  Use of
// this source code is governed by a license that can be found in the License
// file.

package wav

import (
	"bytes"
	"testing"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

func TestFormatIO(t *testing.T) {
	testFormatIo(t, NewStereoFmt())
	testFormatIo(t, NewFormatForm(sound.MonoCd(), sample.SFloat32L))
}

func testFormatIo(t *testing.T, f *Format) {
	buf := bytes.NewBuffer(nil)
	f.Write(buf)
	n := buf.Len() - 8
	buf = bytes.NewBuffer(buf.Bytes()[8:])
	g, e := ParseFormat(buf, n)
	if e != nil {
		t.Error(e)
		return
	}
	if f.Channels() != g.Channels() {
		t.Errorf("channels mismatch got %d not %d\n", f.Channels(), g.Channels())
	}
	if f.SampleRate() != g.SampleRate() {
		t.Errorf("freq mismatch got %s not %s\n", f.SampleRate(), g.SampleRate())
	}
	if f.Codec != g.Codec {
		t.Errorf("codec mismatch got %s not %s\n", g.Codec, f.Codec)
	}
}
