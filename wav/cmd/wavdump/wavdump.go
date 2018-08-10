// Copyright 2018 The ZikiChombo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"zikichombo.org/codec/wav"
)

var cpuprof = flag.String("cpuprof", "", "cpu profile")

func main() {
	flag.Parse()
	if *cpuprof != "" {
		f, e := os.Create(*cpuprof)
		if e != nil {
			log.Fatal(e)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	out := bufio.NewWriter(os.Stdout)
	for _, fn := range flag.Args() {
		dec, err := wav.Load(fn)
		if err != nil {
			log.Printf("%s: %s\n", fn, err.Error())
			continue
		}
		d := make([]float64, 1024)
		for {
			n, err := dec.Receive(d)
			for _, v := range d[:n] {
				fmt.Fprintf(out, "%d\n", int32(v))
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error: %s\n", err.Error())
				break
			}
		}
	}
	out.Flush()
}
