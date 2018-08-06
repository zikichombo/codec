// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package wav

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Top level header of a WAV file
type hdr struct {
	SGroupId string
	Length   uint32
	Type     string
}

// Read reads the header returning an error if the format is unexpected.
func (h *hdr) Read(r io.Reader) error {
	buf := make([]byte, 12)
	n, e := r.Read(buf)
	if e != nil && e != io.EOF {
		return e
	}
	if n != 12 {
		return fmt.Errorf("didn't read all of hdr")
	}
	if buf[0] != 'R' || buf[1] != 'I' || buf[2] != 'F' || buf[3] != 'F' {
		return fmt.Errorf("doesn't start with 'RIFF'")
	}
	if buf[8] != 'W' || buf[9] != 'A' || buf[10] != 'V' || buf[11] != 'E' {
		return fmt.Errorf("not wave header")
	}
	h.Length = binary.LittleEndian.Uint32(buf[4:8])
	return nil
}

const hdrChunkSize = 12

// Write writes the header (12 bytes), returning an error if the format is not correctly written.
func (h *hdr) Write(w io.Writer) error {
	buf := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'A', 'V', 'E'}
	binary.LittleEndian.PutUint32(buf[4:8], h.Length)
	n, e := w.Write(buf)
	if e != nil {
		return e
	}
	if n != 12 {
		return fmt.Errorf("unable to write all of header (%d/12 bytes)", n)
	}
	return nil
}
