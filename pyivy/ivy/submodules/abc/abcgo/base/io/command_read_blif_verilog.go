package io

import (
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func IoCommandReadBlif(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "read_blif", "nmac")
	if !ok {
		return 1
	}
	var network *abc.Network
	var err error
	switch {
	case flags['a']:
		network, err = ReadBlifAsAIG(filename, !flags['c'])
	case !flags['n']:
		network, err = ReadLogic(filename, IO_FILE_BLIF, !flags['c'], false)
	default:
		var netlist *abc.Network
		netlist, err = ReadBLIF(filename, false, !flags['c'])
		if err == nil {
			if flags['m'] {
				abc.Abc_NtkStartNameIds(netlist)
			}
			network = abc.Abc_NtkToLogic(netlist)
			if flags['m'] && network != nil {
				abc.Abc_NtkTransferNameIds(netlist, network)
			}
			abc.Abc_NtkDelete(netlist)
		}
	}
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	if network == nil {
		return 1
	}
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, network)
	abcmain.Abc_FrameClearVerifStatus(frame)
	return 0
}

func IoCommandReadVerilog(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "read_verilog", "mcb")
	if !ok {
		return 1
	}
	network, err := ReadLogic(filename, IO_FILE_VERILOG, !flags['c'], flags['b'])
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, network)
	abcmain.Abc_FrameClearVerifStatus(frame)
	return 0
}
