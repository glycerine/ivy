package io

import (
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func ioFlag(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func ioAcceptNetwork(frame *abcmain.Frame, network *abc.Network) int32 {
	if network == nil {
		return 1
	}
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, network)
	abcmain.Abc_FrameClearVerifStatus(frame)
	return 0
}

func IoCommandReadDsd(frame *abcmain.Frame, args []string) int32 {
	_, expression, ok := ioReadOptions(frame, args, "read_dsd", "c")
	if !ok {
		return 1
	}
	return ioAcceptNetwork(frame, Io_ReadDsd(expression))
}

func IoCommandReadPla(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "read_pla", "zbdxc")
	if !ok {
		return 1
	}
	var network *abc.Network
	var err error
	if flags['z'] || flags['b'] || flags['d'] || flags['x'] {
		var netlist *abc.Network
		netlist, err = readPLAFile(filename, ioFlag(flags['z']), ioFlag(flags['b']), ioFlag(flags['d']), flags['x'], !flags['c'])
		if err == nil {
			network = abc.Abc_NtkToLogic(netlist)
			abc.Abc_NtkDelete(netlist)
		}
	} else {
		network, err = ReadLogic(filename, IO_FILE_PLA, !flags['c'], false)
	}
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	return ioAcceptNetwork(frame, network)
}

func IoCommandReadPlaMo(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "read_plamo", "mv")
	if !ok {
		return 1
	}
	return ioAcceptNetwork(frame, Mop_ManTest(filename, ioFlag(flags['m']), ioFlag(flags['v'])))
}
