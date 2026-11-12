package io

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func IoCommandReadTruth(frame *abcmain.Frame, args []string) int32 {
	flags, source, ok := ioReadOptions(frame, args, "read_truth", "xf")
	if !ok {
		return 1
	}
	truth := source
	if flags['f'] {
		data, err := os.ReadFile(source)
		if err != nil {
			fmt.Fprintln(frame.Err, err)
			return 1
		}
		truth = string(data)
	}
	truth = strings.TrimRight(truth, "\r\n")
	var sop string
	if flags['x'] {
		sop = abc.Abc_SopFromTruthBin(truth)
	} else {
		sop = abc.Abc_SopFromTruthHex(truth)
	}
	if sop == "" {
		fmt.Fprintln(frame.Err, "Reading truth table has failed.")
		return 1
	}
	n := abc.Abc_NtkCreateWithNode(sop)
	if n == nil {
		fmt.Fprintln(frame.Err, "Deriving the network has failed.")
		return 1
	}
	return ioAcceptNetwork(frame, n)
}
