package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func WriteBenchLut(w io.Writer, n *abc.Network) error {
	if n == nil || abc.Abc_NtkIsAigNetlist(n) == 0 {
		return fmt.Errorf("io: LUT BENCH requires AIG netlist")
	}
	if Io_WriteBenchCheckNames(n) == 0 {
		return fmt.Errorf("io: signal names cannot be represented in BENCH")
	}
	manager, ok := n.FunctionManager.(*hop.Hop_Man_t)
	if !ok || manager == nil {
		return fmt.Errorf("io: missing AIG manager")
	}
	buffer := bufio.NewWriter(w)
	for _, port := range n.ObjectsOfKind(abc.ObjectPI) {
		fmt.Fprintf(buffer, "INPUT(%s)\n", abc.Abc_ObjName(abc.Abc_ObjFanout0(port)))
	}
	for _, port := range n.ObjectsOfKind(abc.ObjectPO) {
		fmt.Fprintf(buffer, "OUTPUT(%s)\n", abc.Abc_ObjName(abc.Abc_ObjFanin0(port)))
	}
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		fmt.Fprintf(buffer, "%-11s = DFFRSE( %s, gnd, gnd, gnd, gnd )\n", abc.Abc_ObjName(abc.Abc_ObjFanout0(abc.Abc_ObjFanout0(latch))), abc.Abc_ObjName(abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(latch))))
	}
	scratch := vec.Vec_IntAlloc(10000)
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		count := abc.Abc_ObjFaninNum(node)
		if count > 8 {
			return fmt.Errorf("io: node %q exceeds eight LUT inputs", abc.Abc_ObjName(node))
		}
		function, ok := node.Data.(*hop.Hop_Obj_t)
		if !ok || function == nil {
			return fmt.Errorf("io: node lacks AIG function")
		}
		truth := hop.Hop_ManConvertAigToTruth(manager, function, count, scratch, 0)
		bits := 1 << uint(count)
		allZero, allOne := true, true
		for bit := 0; bit < bits; bit++ {
			value := truth[bit/32] >> uint(bit%32) & 1
			allZero = allZero && value == 0
			allOne = allOne && value == 1
		}
		out := abc.Abc_ObjName(abc.Abc_ObjFanout0(node))
		if allZero {
			fmt.Fprintf(buffer, "%-11s = gnd\n", out)
			continue
		}
		if allOne {
			fmt.Fprintf(buffer, "%-11s = vdd\n", out)
			continue
		}
		var hex strings.Builder
		for digit := (bits+3)/4 - 1; digit >= 0; digit-- {
			bit := digit * 4
			nibble := truth[bit/32] >> uint(bit%32) & 15
			if bits < 4 {
				nibble &= (1 << uint(bits)) - 1
			}
			fmt.Fprintf(&hex, "%x", nibble)
		}
		fmt.Fprintf(buffer, "%-11s = LUT 0x%s (", out, hex.String())
		for i, edge := range node.Fanins() {
			if i > 0 {
				fmt.Fprint(buffer, ",")
			}
			fmt.Fprintf(buffer, " %s", abc.Abc_ObjName(edge.Object))
		}
		fmt.Fprintln(buffer, " )")
	}
	return buffer.Flush()
}

func Io_WriteBenchLut(n *abc.Network, filename string) int32 {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 0
	}
	err = WriteBenchLut(f, n)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 0
	}
	return 1
}
