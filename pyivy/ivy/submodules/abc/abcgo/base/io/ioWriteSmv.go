// SMV writer, ported from ABC ioWriteSmv.c.
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

// cleanUNSAFE replaces '|' (introduced by hierarchy flattening, and read as
// OR by NuSMV) with '_'.
func cleanUNSAFE(text string) string { return strings.ReplaceAll(text, "|", "_") }

func hasPrefix(needle, haystack string) bool { return strings.HasPrefix(haystack, needle) }

// WriteSMV writes an SMV model as Io_WriteSmv does. A strashed network goes
// through Abc_NtkToNetlistBench first, as in Io_Write; any other network
// must already be an SOP netlist of constants, buffers, inverters and
// two-input ANDs.
func WriteSMV(w stdio.Writer, n *abc.Network) error {
	data, err := ioWriteSmvData(n)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func ioWriteSmvData(n *abc.Network) ([]byte, error) {
	if n == nil {
		return nil, fmt.Errorf("nil network")
	}
	if abc.Abc_NtkIsStrash(n) != 0 {
		net := abc.Abc_NtkToNetlistBench(n)
		if net == nil {
			return nil, fmt.Errorf("Converting to netlist has failed.")
		}
		defer abc.Abc_NtkDelete(net)
		n = net
	}
	if abc.Abc_NtkIsSopNetlist(n) == 0 {
		return nil, fmt.Errorf("SMV requires an SOP netlist")
	}
	if Io_WriteSmvCheckNames(n) == 0 {
		fmt.Printf("Io_WriteSmv(): Signal names in this benchmark contain parentheses making them impossible to reproduce in the SMV format. Use \"short_names\".\n")
		return nil, errSmvNames
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "-- benchmark \"%s\" written by ABC on %s\n", n.Name, extra.Extra_TimeStamp())
	Io_WriteSmvOne(&out, n)
	if n.Exdc != nil {
		fmt.Printf("Io_WriteSmv: EXDC is not written (warning).\n")
	}
	return out.Bytes(), nil
}

var errSmvNames = fmt.Errorf("io: names cannot be written in SMV")

// WriteSMVFile is Io_Write's SMV branch.
func WriteSMVFile(n *abc.Network, filename string) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if abc.Abc_NtkIsStrash(n) == 0 {
		fmt.Printf("Writing traditional SMV is available for AIGs only.\n")
		return nil
	}
	data, err := ioWriteSmvData(n)
	if err == errSmvNames {
		return nil
	}
	if err != nil {
		return err
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("Io_WriteSmv(): Cannot open the output file.\n")
	}
	return nil
}

func Io_WriteSmv(n *abc.Network, filename string) int32 {
	data, err := ioWriteSmvData(n)
	if err != nil {
		return 0
	}
	if os.WriteFile(filename, data, 0644) != nil {
		fmt.Printf("Io_WriteSmv(): Cannot open the output file.\n")
		return 0
	}
	return 1
}

// Io_WriteSmvOne writes the module: inputs, state variables, one DEFINE per
// node, the latch assignments and the specs recognized by PO name.
func Io_WriteSmvOne(w stdio.Writer, n *abc.Network) int32 {
	name := func(o *abc.Object) string { return cleanUNSAFE(abc.Abc_ObjName(o)) }
	latches := n.ObjectsOfKind(abc.ObjectLatch)
	fmt.Fprintf(w, "MODULE main\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "VAR  -- inputs\n")
	for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
		fmt.Fprintf(w, "    %s : boolean;\n", name(abc.Abc_ObjFanout0(abc.Abc_NtkPi(n, i))))
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "VAR  -- state variables\n")
	for _, latch := range latches {
		fmt.Fprintf(w, "    %s : boolean;\n", name(abc.Abc_ObjFanout0(abc.Abc_ObjFanout0(latch))))
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "DEFINE\n")
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		if o := abc.Abc_NtkObj(n, i); o != nil && abc.Abc_ObjIsNode(o) != 0 {
			Io_WriteSmvOneNode(w, o)
		}
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "ASSIGN\n")
	for _, latch := range latches {
		output := name(abc.Abc_ObjFanout0(abc.Abc_ObjFanout0(latch)))
		switch {
		case abc.Abc_LatchIsInit0(latch) != 0:
			fmt.Fprintf(w, "    init(%s) := %d;\n", output, 0)
		case abc.Abc_LatchIsInit1(latch) != 0:
			fmt.Fprintf(w, "    init(%s) := %d;\n", output, 1)
		}
		fmt.Fprintf(w, "    next(%s) := ", output)
		fmt.Fprintf(w, "%s;\n", name(abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(latch))))
	}
	fmt.Fprintf(w, "\n")
	for i := int32(0); i < abc.Abc_NtkPoNum(n); i++ {
		net := name(abc.Abc_ObjFanin0(abc.Abc_NtkPo(n, i)))
		switch {
		case hasPrefix("assume_fair_", net):
			fmt.Fprintf(w, "FAIRNESS %s;\n", net)
		case hasPrefix("Assert_", net) || hasPrefix("assert_safety_", net):
			fmt.Fprintf(w, "INVARSPEC %s;\n", net)
		case hasPrefix("assert_fair_", net):
			fmt.Fprintf(w, "LTLSPEC G F %s;\n", net)
		}
	}
	return 1
}

// Io_WriteSmvOneNode writes a constant, buffer, inverter or two-input AND.
func Io_WriteSmvOneNode(w stdio.Writer, node *abc.Object) int32 {
	name := func(o *abc.Object) string { return cleanUNSAFE(abc.Abc_ObjName(o)) }
	switch abc.Abc_ObjFaninNum(node) {
	case 0:
		fmt.Fprintf(w, "    %s", name(abc.Abc_ObjFanout0(node)))
		fmt.Fprintf(w, " := 1;\n")
	case 1:
		if abc.Abc_NodeIsBuf(node) != 0 {
			fmt.Fprintf(w, "    %s := ", name(abc.Abc_ObjFanout0(node)))
		} else {
			fmt.Fprintf(w, "    %s := !", name(abc.Abc_ObjFanout0(node)))
		}
		fmt.Fprintf(w, "%s;\n", name(abc.Abc_ObjFanin0(node)))
	default:
		fmt.Fprintf(w, "    %s", name(abc.Abc_ObjFanout0(node)))
		fmt.Fprintf(w, " := %s & ", name(abc.Abc_ObjFanin0(node)))
		fmt.Fprintf(w, "%s;\n", name(abc.Abc_ObjFanin1(node)))
	}
	return 1
}

// Io_WriteSmvCheckNames rejects networks with parentheses in any name.
func Io_WriteSmvCheckNames(n *abc.Network) int32 {
	if n == nil {
		return 0
	}
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		if o := abc.Abc_NtkObj(n, i); o != nil && strings.ContainsAny(o.Name(), "()") {
			return 0
		}
	}
	return 1
}

func init() { RegisterFormat(IO_FILE_SMV, nil, WriteSMVFile) }
