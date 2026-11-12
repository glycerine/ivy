package io

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func IoCommandWriteBench(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "write_bench", "l")
	if !ok {
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	if flags['l'] {
		if err := WriteNetwork(n, filename, IO_FILE_BENCH); err != nil {
			fmt.Fprintln(frame.Err, err)
		}
		return 0
	}
	net := abc.Abc_NtkToNetlist(n)
	if net == nil {
		fmt.Fprintln(frame.Err, "io: netlist conversion failed")
		return 0
	}
	defer abc.Abc_NtkDelete(net)
	if abc.Abc_NtkToAig(net) == 0 {
		fmt.Fprintln(frame.Err, "io: AIG conversion failed")
		return 0
	}
	Io_WriteBenchLut(net, filename)
	return 0
}
