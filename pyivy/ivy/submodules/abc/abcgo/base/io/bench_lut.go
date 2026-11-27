package io

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
)

// WriteBenchLut is Io_WriteBenchLutOne: ports, latches as DFFRSE and one
// LUT per node of an AIG netlist.
func WriteBenchLut(w io.Writer, n *abc.Network) error {
	if n == nil || abc.Abc_NtkIsAigNetlist(n) == 0 {
		return fmt.Errorf("io: LUT BENCH requires AIG netlist")
	}
	manager, ok := n.FunctionManager.(*hop.Hop_Man_t)
	if !ok || manager == nil {
		return fmt.Errorf("io: missing AIG manager")
	}
	buffer := bufio.NewWriter(w)
	ioWriteBenchPorts(buffer, n, "%-11s = DFFRSE( %s, gnd, gnd, gnd, gnd )\n")
	scratch := vec.Vec_IntAlloc(10000)
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		node := abc.Abc_NtkObj(n, i)
		if node == nil || abc.Abc_ObjIsNode(node) == 0 {
			continue
		}
		if err := Io_WriteBenchLutOneNode(buffer, manager, node, scratch); err != nil {
			return err
		}
	}
	return buffer.Flush()
}

// Io_WriteBenchLutOneNode writes gnd, vdd, a one-input LUT 0x1/0x2, or the
// truth table in hexadecimal with the fanins.
func Io_WriteBenchLutOneNode(w io.Writer, manager *hop.Hop_Man_t, node *abc.Object, scratch *vec.Vec_Int_t) error {
	count := abc.Abc_ObjFaninNum(node)
	if count > 8 {
		return fmt.Errorf("io: node %q exceeds eight LUT inputs", abc.Abc_ObjName(node))
	}
	function, ok := node.Data.(*hop.Hop_Obj_t)
	if !ok || function == nil {
		return fmt.Errorf("io: node lacks AIG function")
	}
	truth := hop.Hop_ManConvertAigToTruth(manager, hop.Hop_Regular(function), count, scratch, 0)
	if hop.Hop_IsComplement(function) != 0 {
		extra.Extra_TruthNot(truth, truth, count)
	}
	out := abc.Abc_ObjName(abc.Abc_ObjFanout0(node))
	if extra.Extra_TruthIsConst0(truth, count) != 0 {
		fmt.Fprintf(w, "%-11s = gnd\n", out)
		return nil
	}
	if extra.Extra_TruthIsConst1(truth, count) != 0 {
		fmt.Fprintf(w, "%-11s = vdd\n", out)
		return nil
	}
	if count == 1 {
		code := 1
		if abc.Abc_NodeIsBuf(node) != 0 {
			code = 2
		}
		fmt.Fprintf(w, "%-11s = LUT 0x%d ( %s )\n", out, code, abc.Abc_ObjName(abc.Abc_ObjFanin0(node)))
		return nil
	}
	fmt.Fprintf(w, "%-11s = LUT 0x", out)
	extra.Extra_PrintHexadecimal(w, truth, count)
	fmt.Fprintf(w, " (")
	for i, edge := range node.Fanins() {
		sep := ","
		if int32(i) == count-1 {
			sep = ""
		}
		fmt.Fprintf(w, " %s%s", abc.Abc_ObjName(edge.Object), sep)
	}
	fmt.Fprintf(w, " )\n")
	return nil
}

// Io_WriteBenchLut writes an AIG netlist in the LUT form of BENCH.
func Io_WriteBenchLut(n *abc.Network, filename string) int32 {
	if Io_WriteBenchCheckNames(n) == 0 {
		fmt.Printf("Io_WriteBenchLut(): Signal names in this benchmark contain parentheses making them impossible to reproduce in the BENCH format. Use \"short_names\".\n")
		return 0
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "# Benchmark \"%s\" written by ABC on %s\n", n.Name, extra.Extra_TimeStamp())
	if err := WriteBenchLut(&out, n); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 0
	}
	if os.WriteFile(filename, out.Bytes(), 0644) != nil {
		fmt.Printf("Io_WriteBench(): Cannot open the output file.\n")
		return 0
	}
	if n.Exdc != nil {
		fmt.Printf("Io_WriteBench: EXDC is not written (warning).\n")
	}
	return 1
}
