// TODO: add seek support; only enabled in the dev branch of mewkiz/flac.

package flac

import (
	"io"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"zikichombo.org/sound"
	"zikichombo.org/sound/freq"
)

// Decoder encapsulates state for decoding and seeking a FLAC audio stream.
type Decoder struct {
	stream *flac.Stream
	frame  *frame.Frame // current frame.
	i      int          // index of current sample in subframe(s).
}

// NewDecoder creates a decoder from a FLAC audio stream (seekable, readable).
func NewDecoder(r io.Reader) (*Decoder, error) {
	stream, err := flac.New(r)
	if err != nil {
		return nil, err
	}
	d := &Decoder{
		stream: stream,
	}
	return d, nil
}

func (d *Decoder) Receive(dst []float64) (int, error) {
	if len(dst)%d.Channels() != 0 {
		return 0, sound.ChannelAlignmentError
	}
	n := 0                       // number of frames (samples per channel) buffered.
	m := len(dst) / d.Channels() // number of frames (samples per channel) to buffer.
	bps := int(d.stream.Info.BitsPerSample)
	for n < m {
		if d.frame == nil {
			frame, err := d.stream.ParseNext()
			if err != nil {
				return n, err
			}
			d.frame = frame
			d.i = 0
		}
		samplesLeft := len(d.frame.Subframes[0].Samples[d.i:])
		j := m
		if j > samplesLeft {
			j = samplesLeft
		}
		if n+j > m {
			j = m - n
		}
		if j == 0 {
			panic("cannot retrieve zero samples")
		}
		for c := 0; c < d.Channels(); c++ {
			toFloats(dst[c*m:c*m+j], d.frame.Subframes[c].Samples[d.i:d.i+j], bps)
		}
		d.i += j
		n += j
		if len(d.frame.Subframes[0].Samples[d.i:]) == 0 {
			d.frame = nil
			d.i = 0
		}
	}
	if n < m {
		// TODO: handle n < m. Move frames in dst to their correct position
		//
		//    dst[c*n:(c+1)*n]
		panic("flac.Decoder.Receive: reached end of stream with n < m; support for this case is not yet implemented;\n\nThe implementation should move frames (in ziki terminology) in dst to their correct position, i.e.\n\n   dst[c*n:(c+1)*n]")
	}
	return n, nil
}

func (d *Decoder) Channels() int {
	return int(d.stream.Info.NChannels)
}

func (d *Decoder) SampleRate() freq.T {
	return freq.T(d.stream.Info.SampleRate) * freq.Hertz
}

func (d *Decoder) Close() error {
	return d.stream.Close()
}

// ### [ Helper functions ] ####################################################

// Copied from sound/sample/fix.go and adjusted from int64 to int32.

func toFloat(d int32, nBits int) float64 {
	s := float64(int32(1 << uint(nBits-1)))
	return float64(d) / s
}

func toFloats(dst []float64, src []int32, nBits int) []float64 {
	if cap(dst) < len(src) {
		panic("capacity of dst buffer too small")
	}
	dst = dst[:len(src)]
	for i, v := range src {
		dst[i] = toFloat(v, nBits)
	}
	return dst
}
