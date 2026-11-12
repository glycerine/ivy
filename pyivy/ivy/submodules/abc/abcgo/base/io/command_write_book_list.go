package io

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func IoCommandWriteBook(frame *abcmain.Frame, args []string) int32 {
	if len(args) != 2 || strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: write_book [-h] <file> [-options]")
		return 1
	}
	n, _ := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if err := WriteNetwork(n, args[1], IO_FILE_BOOK); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}

// IoCommandWriteList is disabled by an unconditional return in ABC.
func IoCommandWriteList(frame *abcmain.Frame, args []string) int32 {
	fmt.Fprintln(frame.Out, "This command currently does not work.")
	return 0
}

// ioCommandWriteList retains the implementation following ABC's early return.
func ioCommandWriteList(frame *abcmain.Frame, args []string) int32 {
	useHost := true
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, option := range args[i][1:] {
			if option == 'n' {
				useHost = !useHost
			} else {
				fmt.Fprintln(frame.Err, "usage: write_list [-nh] <file>")
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
	if i+1 != len(args) {
		fmt.Fprintln(frame.Err, "usage: write_list [-nh] <file>")
		return 1
	}
	host := int32(0)
	if useHost {
		host = 1
	}
	if err := Io_WriteList(n, args[i], host); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
