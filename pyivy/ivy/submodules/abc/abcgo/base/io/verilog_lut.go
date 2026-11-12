package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
)

func Io_WriteVerilogObjectsLut(w io.Writer, n *abc.Network, size int32) {
	if size < 2 || size > 6 {
		panic("io: LUT size must be between 2 and 6")
	}
	writeVerilogBoxes(w, n)
	nodes := n.ObjectsOfKind(abc.ObjectNode)
	length := 0
	for _, node := range nodes {
		length = max(length, len(Io_WriteVerilogGetName(abc.Abc_ObjName(abc.Abc_ObjFanout0(node)))))
		for _, edge := range node.Fanins() {
			length = max(length, len(Io_WriteVerilogGetName(abc.Abc_ObjName(edge.Object))))
		}
	}
	digits := len(strconv.Itoa(len(nodes)))
	for i, node := range nodes {
		if abc.Abc_ObjFaninNum(node) > size {
			panic("io: node exceeds LUT size")
		}
		sop, ok := node.Data.(string)
		if !ok {
			panic("io: LUT node has no SOP")
		}
		truth := abc.Abc_SopToTruth(sop, abc.Abc_ObjFaninNum(node))
		bits := uint(1) << uint(size)
		if bits < 64 {
			truth &= (uint64(1) << bits) - 1
		}
		fmt.Fprintf(w, "  lut%d #(%d'h%0*x) lut_%0*d ( {", size, bits, int(bits/4), truth, digits, i)
		pins := make([]string, 0, size)
		for k := size; k > abc.Abc_ObjFaninNum(node); k-- {
			pins = append(pins, "1'b0")
		}
		for k := abc.Abc_ObjFaninNum(node) - 1; k >= 0; k-- {
			pins = append(pins, Io_WriteVerilogGetName(abc.Abc_ObjName(abc.Abc_ObjFanin(node, k))))
		}
		for j, pin := range pins {
			if j > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprintf(w, "%*s", length, pin)
		}
		fmt.Fprintf(w, "}, %*s );\n", length, Io_WriteVerilogGetName(abc.Abc_ObjName(abc.Abc_ObjFanout0(node))))
	}
}

func Io_WriteVerilogLutInt(w io.Writer, n *abc.Network, size int32) {
	writeVerilogModule(w, n, func() { Io_WriteVerilogObjectsLut(w, n, size) })
}

func WriteVerilogLut(w io.Writer, n *abc.Network, size int32) error {
	if n == nil || size < 2 || size > 6 {
		return fmt.Errorf("io: invalid network or LUT size")
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if abc.Abc_ObjFaninNum(node) > size {
			return fmt.Errorf("io: node %q exceeds LUT size %d", abc.Abc_ObjName(node), size)
		}
	}
	net := abc.Abc_NtkToNetlist(n)
	if net == nil {
		return fmt.Errorf("io: cannot convert network to netlist")
	}
	defer abc.Abc_NtkDelete(net)
	if abc.Abc_NtkToSop(net, -1, util.ABC_INFINITY) == 0 {
		return fmt.Errorf("io: cannot convert network to SOP")
	}
	buffer := bufio.NewWriter(w)
	fmt.Fprintf(buffer, "// Benchmark %q written by ABC on %s\n\n", n.Name, extra.Extra_TimeStamp())
	Io_WriteLutModule(buffer, size)
	Io_WriteVerilogLutInt(buffer, net, size)
	fmt.Fprintln(buffer)
	return buffer.Flush()
}

func Io_WriteVerilogLut(n *abc.Network, filename string, size int32) {
	var b strings.Builder
	if err := WriteVerilogLut(&b, n, size); err != nil {
		fmt.Println(err)
		return
	}
	if err := os.WriteFile(filename, []byte(b.String()), 0666); err != nil {
		fmt.Printf("Io_WriteVerilog(): Cannot write output file %q: %v\n", filename, err)
	}
}
