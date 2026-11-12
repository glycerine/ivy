package io

import (
	"fmt"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func ioCommandReadFormat(frame *abcmain.Frame, args []string, name string, kind Io_FileType_t) int32 {
	check := true
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, option := range args[i][1:] {
			if option == 'c' {
				check = !check
			} else {
				fmt.Fprintf(frame.Err, "usage: %s [-ch] <file>\n", name)
				return 1
			}
		}
		i++
	}
	if i+1 != len(args) {
		fmt.Fprintf(frame.Err, "usage: %s [-ch] <file>\n", name)
		return 1
	}
	n, err := ReadLogic(args[i], kind, check, false)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, n)
	abcmain.Abc_FrameClearVerifStatus(frame)
	return 0
}

func IoCommandReadAiger(f *abcmain.Frame, a []string) int32 {
	return ioCommandReadFormat(f, a, "read_aiger", IO_FILE_AIGER)
}
func IoCommandReadBaf(f *abcmain.Frame, a []string) int32 {
	return ioCommandReadFormat(f, a, "read_baf", IO_FILE_BAF)
}
func IoCommandReadBblif(f *abcmain.Frame, a []string) int32 {
	return ioCommandReadFormat(f, a, "read_bblif", IO_FILE_BBLIF)
}
func IoCommandReadBlifMv(f *abcmain.Frame, a []string) int32 {
	return ioCommandReadFormat(f, a, "read_blif_mv", IO_FILE_BLIFMV)
}
func IoCommandReadBench(f *abcmain.Frame, a []string) int32 {
	return ioCommandReadFormat(f, a, "read_bench", IO_FILE_BENCH)
}
func IoCommandReadEdif(f *abcmain.Frame, a []string) int32 {
	return ioCommandReadFormat(f, a, "read_edif", IO_FILE_EDIF)
}
func IoCommandReadEqn(f *abcmain.Frame, a []string) int32 {
	return ioCommandReadFormat(f, a, "read_eqn", IO_FILE_EQN)
}
