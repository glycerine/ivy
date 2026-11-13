package io

import (
	"fmt"
	"strings"
	"sync"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

var specialBLIFWriter struct {
	sync.RWMutex
	fn func(*abc.Network, string, string, bool) error
}

// RegisterSpecialBLIFWriter supplies ABC's LUT-decomposing BLIF writer.
func RegisterSpecialBLIFWriter(fn func(*abc.Network, string, string, bool) error) {
	specialBLIFWriter.Lock()
	specialBLIFWriter.fn = fn
	specialBLIFWriter.Unlock()
}

func IoCommandWriteBlif(frame *abcmain.Frame, args []string) int32 {
	var structure string
	special, hierarchical := false, false
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, option := range args[i][1:] {
			switch option {
			case 'S':
				i++
				if i >= len(args) {
					fmt.Fprintln(frame.Err, "usage: write_blif [-S str] [-jah] <file>")
					return 1
				}
				structure = args[i]
				if len(structure) != 2 && len(structure) != 3 {
					fmt.Fprintln(frame.Err, "LUT structure must have two or three characters")
					return 1
				}
			case 'j':
				special = !special
			case 'a':
				hierarchical = !hierarchical
			default:
				fmt.Fprintln(frame.Err, "usage: write_blif [-S str] [-jah] <file>")
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
		fmt.Fprintln(frame.Err, "usage: write_blif [-S str] [-jah] <file>")
		return 1
	}
	if special || structure != "" {
		specialBLIFWriter.RLock()
		writer := specialBLIFWriter.fn
		specialBLIFWriter.RUnlock()
		if writer == nil {
			fmt.Fprintln(frame.Err, "io: specialized BLIF writer is not registered")
			return 1
		}
		if err := writer(n, args[i], structure, hierarchical); err != nil {
			fmt.Fprintln(frame.Err, err)
		}
	} else if err := WriteNetwork(n, args[i], IO_FILE_BLIF); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
