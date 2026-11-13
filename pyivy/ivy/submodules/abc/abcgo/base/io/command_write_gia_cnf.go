package io

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
)

func IoCommandWriteCnf2(frame *abcmain.Frame, args []string) int32 {
	flags, size, filename, ok := ioWriteOptions(frame, args, "&write_cnf", "aiov", 'K')
	if !ok {
		return 1
	}
	if !flags['K'] {
		size = 8
	}
	p, ok := abcmain.Abc_FrameReadGia(frame).(*gia.Gia_Man_t)
	if !ok || p == nil {
		fmt.Fprintln(frame.Err, "IoCommandWriteCnf2(): There is no AIG.")
		return 1
	}
	if gia.Gia_ManRegNum(p) > 0 {
		fmt.Fprintln(frame.Err, "IoCommandWriteCnf2(): Works only for combinational miters.")
		return 0
	}
	if size < 3 || size > 8 {
		fmt.Fprintf(frame.Err, "IoCommandWriteCnf2(): Invalid LUT size (%d).\n", size)
		return 0
	}
	if flags['a'] && extra.Sdm_ManCanRead() == 0 {
		fmt.Fprintln(frame.Err, "IoCommandWriteCnf2(): Cannot input precomputed DSD information.")
		return 0
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	if err = f.Close(); err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	verbose, ids, output := int32(0), int32(0), int32(1)
	if flags['v'] {
		verbose = 1
	}
	if flags['i'] {
		ids = 1
	}
	if flags['o'] {
		output = 0
	}
	if flags['a'] {
		gia.Jf_ManDumpCnf(p, filename, verbose)
	} else {
		gia.Mf_ManDumpCnf(p, filename, size, ids, output, verbose)
	}
	return 0
}
