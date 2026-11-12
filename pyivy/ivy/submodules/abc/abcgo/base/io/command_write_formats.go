package io

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func ioCommandWriteFormat(frame *abcmain.Frame, args []string, name string, kind Io_FileType_t) int32 {
	if len(args) > 1 && strings.HasPrefix(args[1], "-") {
		fmt.Fprintf(frame.Err, "usage: %s [-h] <file>\n", name)
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	if len(args) != 2 {
		fmt.Fprintf(frame.Err, "usage: %s [-h] <file>\n", name)
		return 1
	}
	if err := WriteNetwork(n, args[1], kind); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}

func IoCommandWriteBaf(f *abcmain.Frame, a []string) int32 {
	return ioCommandWriteFormat(f, a, "write_baf", IO_FILE_BAF)
}
func IoCommandWriteBblif(f *abcmain.Frame, a []string) int32 {
	return ioCommandWriteFormat(f, a, "write_bblif", IO_FILE_BBLIF)
}
func IoCommandWriteBlifMv(f *abcmain.Frame, a []string) int32 {
	return ioCommandWriteFormat(f, a, "write_blif_mv", IO_FILE_BLIFMV)
}
func IoCommandWriteDot(f *abcmain.Frame, a []string) int32 {
	return ioCommandWriteFormat(f, a, "write_dot", IO_FILE_DOT)
}
func IoCommandWriteEqn(f *abcmain.Frame, a []string) int32 {
	return ioCommandWriteFormat(f, a, "write_eqn", IO_FILE_EQN)
}
func IoCommandWriteGml(f *abcmain.Frame, a []string) int32 {
	return ioCommandWriteFormat(f, a, "write_gml", IO_FILE_GML)
}
func IoCommandWriteSmv(f *abcmain.Frame, a []string) int32 {
	return ioCommandWriteFormat(f, a, "write_smv", IO_FILE_SMV)
}
