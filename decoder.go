// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package codec

import (
	"io"

	"zikichombo.org/sound"
	"zikichombo.org/sound/sample"
)

type Decoder interface {
	Extensions() []string
	Manufacturer() string
	Decoder(r io.Reader) (sound.Source, sample.Codec, error)
}
