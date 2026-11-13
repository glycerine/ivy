package io

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
)

func IoCommandWriteTruth(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "write_truth", "xr")
	if !ok {
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Current network is not available.")
		return 0
	}
	if abc.Abc_NtkIsLogic(n) == 0 {
		fmt.Fprintln(frame.Out, "Current network should not be an AIG. Run \"logic\".")
		return 0
	}
	nodes := n.ObjectsOfKind(abc.ObjectNode)
	if abc.Abc_NtkPoNum(n) != 1 || len(nodes) != 1 {
		fmt.Fprintln(frame.Out, "Current network should have exactly one primary output and one node.")
		return 0
	}
	node := abc.Abc_ObjFanin0(abc.Abc_NtkPo(n, 0))
	count := abc.Abc_ObjFaninNum(node)
	if count < 1 || count > 16 {
		fmt.Fprintln(frame.Out, "Can only write logic functions with 1 to 16 inputs.")
		return 0
	}
	if abc.Abc_NtkToAig(n) == 0 {
		fmt.Fprintln(frame.Err, "Cannot convert network to AIG functions.")
		return 0
	}
	manager, ok := n.FunctionManager.(*hop.Hop_Man_t)
	if !ok {
		fmt.Fprintln(frame.Err, "Missing AIG manager.")
		return 0
	}
	function, ok := node.Data.(*hop.Hop_Obj_t)
	if !ok {
		fmt.Fprintln(frame.Err, "Missing AIG function.")
		return 0
	}
	reverse := int32(0)
	if flags['r'] {
		reverse = 1
	}
	truth := hop.Hop_ManConvertAigToTruth(manager, function, count, vec.Vec_IntAlloc(0), reverse)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	if flags['x'] {
		extra.Extra_PrintBinary(f, truth, 1<<uint(count))
	} else {
		extra.Extra_PrintHex2(f, truth, count)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}

func IoCommandWriteTruths(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "&write_truths", "rb")
	if !ok {
		return 1
	}
	p, ok := abcmain.Abc_FrameReadGia(frame).(*gia.Gia_Man_t)
	if !ok || p == nil {
		fmt.Fprintln(frame.Err, "IoCommandWriteTruths(): There is no AIG.")
		return 1
	}
	vars := gia.Gia_ManPiNum(p)
	if vars < 3 || vars > 16 {
		fmt.Fprintln(frame.Err, "IoCommandWriteTruths(): Requires 3 to 16 inputs.")
		return 0
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	defer f.Close()
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		truth := gia.Gia_ObjComputeTruthTable(p, gia.Gia_ManCo(p, i))
		if flags['b'] {
			for _, word := range truth {
				if err = binary.Write(f, binary.LittleEndian, word); err != nil {
					fmt.Fprintln(frame.Err, err)
					return 0
				}
			}
		} else {
			words := make([]uint32, len(truth)*2)
			for j, word := range truth {
				words[2*j], words[2*j+1] = uint32(word), uint32(word>>32)
			}
			extra.Extra_PrintHex(f, words, vars)
			fmt.Fprintln(f)
		}
	}
	return 0
}
