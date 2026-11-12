package io

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
)

func frameCounterexample(source *abcmain.Counterexample) (*util.Abc_Cex_t, error) {
	if source == nil || source.Frame < 0 || source.Frame == 1<<31-1 || source.Registers < 0 || source.PIs < 0 {
		return nil, fmt.Errorf("invalid counterexample dimensions")
	}
	bits := int64(source.Registers) + int64(source.PIs)*(int64(source.Frame)+1)
	if bits > 1<<31-1 || bits > int64(len(source.Bits))*64 {
		return nil, fmt.Errorf("counterexample data is truncated")
	}
	result := util.Abc_CexAlloc(source.Registers, source.PIs, source.Frame+1)
	result.IPo, result.IFrame = source.Output, source.Frame
	for i := range result.PData {
		word := source.Bits[i/2]
		if i%2 != 0 {
			word >>= 32
		}
		result.PData[i] = uint32(word)
	}
	return result, nil
}

func IoCommandWriteAigerCex(frame *abcmain.Frame, args []string) int32 {
	if len(args) > 1 && strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: write_aiger_cex [-h] <file>")
		return 1
	}
	_, _, snapshot := frame.VerificationResult()
	if snapshot == nil {
		fmt.Fprintln(frame.Out, "There is no current CEX.")
		return 0
	}
	if len(args) != 2 {
		fmt.Fprintln(frame.Err, "usage: write_aiger_cex [-h] <file>")
		return 1
	}
	cex, err := frameCounterexample(snapshot)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	network, _ := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if err := WriteAigerCounterexample(cex, network, abcmain.Abc_FrameReadGia(frame), args[1]); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
