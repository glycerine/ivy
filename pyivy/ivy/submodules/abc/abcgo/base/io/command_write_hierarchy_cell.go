package io

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func IoCommandWriteHie(frame *abcmain.Frame, args []string) int32 {
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		for _, option := range args[i][1:] {
			if option != 'm' {
				fmt.Fprintln(frame.Err, "usage: write_hie [-mh] <orig> <file>")
				return 1
			}
		}
		i++
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	if i+2 != len(args) {
		fmt.Fprintln(frame.Err, "usage: write_hie [-mh] <orig> <file>")
		return 1
	}
	if err := Io_WriteHie(n, args[i], args[i+1]); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}

func IoCommandWriteCellNet(frame *abcmain.Frame, args []string) int32 {
	if len(args) > 1 && strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: write_cellnet [-h] <file>")
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	if len(args) != 2 {
		fmt.Fprintln(frame.Err, "usage: write_cellnet [-h] <file>")
		return 1
	}
	if abc.Abc_NtkIsLogic(n) == 0 {
		fmt.Fprintln(frame.Out, "The network should be a logic network (if it an AIG, use command \"logic\")")
		return 0
	}
	if err := Io_WriteCellNet(n, args[1]); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
