package io

import (
	"fmt"
	"io"

	"github.com/glycerine/abcgo/base/abc"
)

func Io_WriteVerilogLatches(w io.Writer, n *abc.Network) {
	latches := n.ObjectsOfKind(abc.ObjectLatch)
	if len(latches) == 0 {
		return
	}
	fmt.Fprintln(w, "  always @ (posedge clock) begin")
	for _, latch := range latches {
		output := abc.Abc_ObjFanout0(abc.Abc_ObjFanout0(latch))
		input := abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(latch))
		fmt.Fprintf(w, "    %s <= %s;\n", Io_WriteVerilogGetName(abc.Abc_ObjName(output)), Io_WriteVerilogGetName(abc.Abc_ObjName(input)))
	}
	fmt.Fprintln(w, "  end")
	hasInitial := false
	for _, latch := range latches {
		if abc.Abc_LatchIsInit0(latch) != 0 || abc.Abc_LatchIsInit1(latch) != 0 {
			hasInitial = true
			break
		}
	}
	if !hasInitial {
		return
	}
	fmt.Fprintln(w, "  initial begin")
	for _, latch := range latches {
		value := -1
		if abc.Abc_LatchIsInit0(latch) != 0 {
			value = 0
		} else if abc.Abc_LatchIsInit1(latch) != 0 {
			value = 1
		}
		if value >= 0 {
			output := abc.Abc_ObjFanout0(abc.Abc_ObjFanout0(latch))
			fmt.Fprintf(w, "    %s <= 1'b%d;\n", Io_WriteVerilogGetName(abc.Abc_ObjName(output)), value)
		}
	}
	fmt.Fprintln(w, "  end")
}
