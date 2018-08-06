// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type fourCc [4]byte

var (
	_riff4Cc = [4]byte{'R', 'I', 'F', 'F'}
	_wave4Cc = [4]byte{'W', 'A', 'V', 'E'}
	_fmt4Cc  = [4]byte{'f', 'm', 't', ' '}
	_dat4Cc  = [4]byte{'d', 'a', 't', 'a'}
	_list4Cc = [4]byte{'L', 'I', 'S', 'T'}
)

const chunkHdrSize = 8

func (f fourCc) isList() bool {
	return f == _riff4Cc || f == _list4Cc
}

type chunk struct {
	fourCc   fourCc
	start    int64
	length   int
	parent   *chunk
	children []*chunk
}

func readChunk(r io.Reader, off int64) (*chunk, error) {
	var buf [8]byte
	ttl, n := 0, 0
	var err error
	for ttl < 8 {
		n, err = r.Read(buf[ttl:])
		if err != nil {
			return nil, err
		}
		ttl += n
	}
	c := &chunk{}
	copy(c.fourCc[:], buf[:4])
	c.start = off
	c.length = int(binary.LittleEndian.Uint32(buf[4:]))
	return c, nil
}

func (c *chunk) readChunk(r io.Reader) (*chunk, error) {
	start := c.start + 8
	if len(c.children) != 0 {
		p := c.children[len(c.children)-1]
		start = p.start + int64(p.length) + 8
	}
	child, err := readChunk(r, start)
	if err != nil {
		return nil, err
	}
	child.parent = c
	child.parent.children = append(child.parent.children, child)
	return child, nil
}

func (c *chunk) findChunk(r io.Reader, fcc fourCc) (*chunk, error) {
	for {
		nxt, err := c.readChunk(r)
		if err != nil {
			return nil, err
		}
		if string(nxt.fourCc[:]) == string(fcc[:]) {
			return nxt, nil
		}
		skip(r, int(nxt.length))
	}
}

func skip(r io.Reader, n int) error {
	if s, ok := r.(io.ReadSeeker); ok {
		_, err := s.Seek(int64(n), os.SEEK_CUR)
		return err
	}
	buf := make([]byte, n)
	ttl, m := 0, 0
	var err error
	for ttl < n {
		m, err = r.Read(buf[ttl:])
		if err != nil {
			return err
		}
		ttl += m
	}
	return nil
}

func (c *chunk) Seek(s io.Seeker, off int64) error {
	_, err := s.Seek(c.start+off+8, os.SEEK_SET)
	return err
}

func (c *chunk) writeHdr(w io.Writer) error {
	var buf [8]byte
	copy(buf[:4], c.fourCc[:])
	binary.LittleEndian.PutUint32(buf[4:], uint32(c.length))
	_, err := w.Write(buf[:])
	return err
}

func readRiff(r io.Reader) (*chunk, fourCc, error) {
	var riff *chunk
	var err error
	var fcc fourCc
	riff, err = readChunk(r, 4)
	if err != nil {
		return nil, fcc, err
	}
	if string(riff.fourCc[:]) != "RIFF" {
		return nil, fcc, fmt.Errorf("not a riff file\n")
	}
	n, e := r.Read(fcc[:])
	if e != nil {
		return nil, fcc, e
	}
	if n != 4 {
		return nil, fcc, fmt.Errorf("couldn't read 4\n")
	}
	return riff, fcc, nil
}
