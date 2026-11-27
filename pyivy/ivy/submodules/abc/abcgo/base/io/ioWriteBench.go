package io

import (
	"bytes"
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
)

// WriteBench writes BENCH text as ABC would: an AIG goes through
// Abc_NtkToNetlistBench and Io_WriteBench's gate form (Io_Write's BENCH
// branch); any other network through the AIG netlist and the LUT form
// (write_bench's default). Names with parentheses are refused.
func WriteBench(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	net := n
	switch {
	case abc.Abc_NtkIsStrash(n) != 0:
		net = abc.Abc_NtkToNetlistBench(n)
	case abc.Abc_NtkIsNetlist(n) == 0:
		net = abc.Abc_NtkToNetlist(n)
	}
	if net == nil {
		return fmt.Errorf("Converting to netlist has failed.")
	}
	if net != n {
		defer abc.Abc_NtkDelete(net)
	}
	if abc.Abc_NtkIsStrash(n) == 0 && abc.Abc_NtkHasAig(net) == 0 && abc.Abc_NtkToAig(net) == 0 {
		return fmt.Errorf("io: cannot convert network to AIG functions")
	}
	if Io_WriteBenchCheckNames(net) == 0 {
		return fmt.Errorf("io: signal names in this benchmark contain parentheses")
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "# Benchmark \"%s\" written by ABC on %s\n", net.Name, extra.Extra_TimeStamp())
	if abc.Abc_NtkIsStrash(n) != 0 {
		Io_WriteBenchOne(&out, net)
	} else if err := WriteBenchLut(&out, net); err != nil {
		return err
	}
	_, err := w.Write(out.Bytes())
	return err
}

// WriteBenchFile is Io_Write's BENCH branch (write_bench -l).
func WriteBenchFile(n *abc.Network, filename string) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if abc.Abc_NtkIsStrash(n) == 0 {
		fmt.Printf("Writing traditional BENCH is available for AIGs only (use \"write_bench\").\n")
		return nil
	}
	net := abc.Abc_NtkToNetlistBench(n)
	if net == nil {
		fmt.Printf("Converting to netlist has failed.\n")
		return nil
	}
	defer abc.Abc_NtkDelete(net)
	Io_WriteBench(net, filename)
	return nil
}

// Io_WriteBench writes an SOP netlist of gates (Abc_NtkToNetlistBench's).
func Io_WriteBench(n *abc.Network, filename string) int32 {
	if Io_WriteBenchCheckNames(n) == 0 {
		fmt.Printf("Io_WriteBench(): Signal names in this benchmark contain parentheses making them impossible to reproduce in the BENCH format. Use \"short_names\".\n")
		return 0
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "# Benchmark \"%s\" written by ABC on %s\n", n.Name, extra.Extra_TimeStamp())
	Io_WriteBenchOne(&out, n)
	if os.WriteFile(filename, out.Bytes(), 0644) != nil {
		fmt.Printf("Io_WriteBench(): Cannot open the output file.\n")
		return 0
	}
	if n.Exdc != nil {
		fmt.Printf("Io_WriteBench: EXDC is not written (warning).\n")
	}
	return 1
}

// Io_WriteBenchOne writes the ports, the latches as DFFs and one gate per
// node (constant, buffer, inverter or two-input AND).
func Io_WriteBenchOne(w stdio.Writer, n *abc.Network) int32 {
	ioWriteBenchPorts(w, n, "%-11s = DFF(%s)\n")
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		if o := abc.Abc_NtkObj(n, i); o != nil && abc.Abc_ObjIsNode(o) != 0 {
			Io_WriteBenchOneNode(w, o)
		}
	}
	return 1
}

func ioWriteBenchPorts(w stdio.Writer, n *abc.Network, latchFormat string) {
	for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
		fmt.Fprintf(w, "INPUT(%s)\n", abc.Abc_ObjName(abc.Abc_ObjFanout0(abc.Abc_NtkPi(n, i))))
	}
	for i := int32(0); i < abc.Abc_NtkPoNum(n); i++ {
		fmt.Fprintf(w, "OUTPUT(%s)\n", abc.Abc_ObjName(abc.Abc_ObjFanin0(abc.Abc_NtkPo(n, i))))
	}
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		fmt.Fprintf(w, latchFormat, abc.Abc_ObjName(abc.Abc_ObjFanout0(abc.Abc_ObjFanout0(latch))), abc.Abc_ObjName(abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(latch))))
	}
}

func Io_WriteBenchOneNode(w stdio.Writer, node *abc.Object) int32 {
	out := abc.Abc_ObjName(abc.Abc_ObjFanout0(node))
	switch abc.Abc_ObjFaninNum(node) {
	case 0:
		fmt.Fprintf(w, "%-11s", out)
		fmt.Fprintf(w, " = vdd\n")
	case 1:
		if abc.Abc_NodeIsBuf(node) != 0 {
			fmt.Fprintf(w, "%-11s = BUFF(", out)
		} else {
			fmt.Fprintf(w, "%-11s = NOT(", out)
		}
		fmt.Fprintf(w, "%s)\n", abc.Abc_ObjName(abc.Abc_ObjFanin0(node)))
	default:
		fmt.Fprintf(w, "%-11s", out)
		fmt.Fprintf(w, " = AND(%s, ", abc.Abc_ObjName(abc.Abc_ObjFanin0(node)))
		fmt.Fprintf(w, "%s)\n", abc.Abc_ObjName(abc.Abc_ObjFanin1(node)))
	}
	return 1
}

// Io_WriteBenchCheckNames rejects networks with parentheses in any name.
func Io_WriteBenchCheckNames(n *abc.Network) int32 {
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		if o := abc.Abc_NtkObj(n, i); o != nil && strings.ContainsAny(o.Name(), "()") {
			return 0
		}
	}
	return 1
}
func init() { RegisterFormat(IO_FILE_BENCH, nil, WriteBenchFile) }
