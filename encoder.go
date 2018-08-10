// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package codec

import (
	"io"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

type Encoder interface {
	Manufacturer() string
	Extensions() []string
	SampleCodecs() []sample.Codec
	Encoder(v sound.Form, w io.WriteCloser) (sound.Sink, error)
	EncoderWith(v sound.Form, c sample.Codec, w io.WriteCloser) (sound.Sink, error)
}
