package io

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
)

// WriteVerilog writes an AIG or mapped netlist and its hierarchical modules.
func WriteVerilog(w io.Writer, n *abc.Network, onlyAnds int32) error {
	if n == nil || abc.Abc_NtkIsAigNetlist(n) == 0 && abc.Abc_NtkIsMappedNetlist(n) == 0 {
		return fmt.Errorf("io: Verilog requires an AIG or mapped netlist")
	}
	buffer := bufio.NewWriter(w)
	w = buffer
	fmt.Fprintf(w, "// Benchmark %q written by ABC on %s\n\n", n.Name, extra.Extra_TimeStamp())
	Io_WriteVerilogInt(w, n, onlyAnds)
	if n.Design != nil {
		for _, module := range n.Design.Modules {
			if module == n {
				continue
			}
			if abc.Abc_NtkIsNetlist(module) == 0 {
				return fmt.Errorf("io: module %q is not a netlist", module.Name)
			}
			fmt.Fprintln(w)
			Io_WriteVerilogInt(w, module, onlyAnds)
		}
	}
	fmt.Fprintln(w)
	return buffer.Flush()
}

func Io_WriteVerilog(n *abc.Network, filename string, onlyAnds int32) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Io_WriteVerilog(): Cannot open the output file %q.\n", filename)
		return
	}
	if err = WriteVerilog(f, n, onlyAnds); err != nil {
		fmt.Println(err)
	}
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}
}

func Io_WriteVerilogInt(w io.Writer, n *abc.Network, onlyAnds int32) {
	inputs, outputs := n.ObjectsOfKind(abc.ObjectPI), n.ObjectsOfKind(abc.ObjectPO)
	latches := n.ObjectsOfKind(abc.ObjectLatch)
	clock := len(latches) > 0
	for _, pi := range inputs {
		if abc.Abc_ObjName(pi) == "clock" {
			clock = false
			break
		}
	}
	fmt.Fprintf(w, "module %s ( ", Io_WriteVerilogGetName(n.Name))
	if clock {
		fmt.Fprint(w, "clock, ")
	}
	fmt.Fprint(w, "\n   ")
	if len(inputs) > 0 {
		Io_WriteVerilogPis(w, n, 3)
		fmt.Fprint(w, ",\n   ")
	}
	if len(outputs) > 0 {
		Io_WriteVerilogPos(w, n, 3)
	}
	fmt.Fprint(w, "  );\n")
	if clock {
		fmt.Fprintln(w, "  input  clock;")
	}
	if len(inputs) > 0 {
		fmt.Fprint(w, "  input ")
		Io_WriteVerilogPis(w, n, 10)
		fmt.Fprintln(w, ";")
	}
	if len(outputs) > 0 {
		fmt.Fprint(w, "  output")
		Io_WriteVerilogPos(w, n, 5)
		fmt.Fprintln(w, ";")
	}
	if abc.Abc_NtkHasBlackbox(n) == 0 {
		if len(latches) > 0 {
			fmt.Fprint(w, "  reg")
			Io_WriteVerilogRegs(w, n, 4)
			fmt.Fprintln(w, ";")
		}
		if Io_WriteVerilogWiresCount(n) > 0 {
			fmt.Fprint(w, "  wire")
			Io_WriteVerilogWires(w, n, 4)
			fmt.Fprintln(w, ";")
		}
		Io_WriteVerilogObjects(w, n, onlyAnds)
		Io_WriteVerilogLatches(w, n)
	}
	fmt.Fprint(w, "endmodule\n\n")
}
