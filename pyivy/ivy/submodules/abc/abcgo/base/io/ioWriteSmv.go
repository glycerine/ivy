package io

import (
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

// cleanUNSAFE retains the source helper name but owns its result; there is no
// static buffer and no truncation. The full exporter also resolves collisions.
func cleanUNSAFE(text string) string {
	var out strings.Builder
	for i, c := range text {
		valid := c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || i > 0 && c >= '0' && c <= '9'
		if valid {
			out.WriteRune(c)
		} else {
			out.WriteByte('_')
		}
	}
	if out.Len() == 0 {
		return "signal"
	}
	name := out.String()
	switch strings.ToUpper(name) {
	case "MODULE", "TRUE", "FALSE", "BOOLEAN", "CASE", "ESAC", "NEXT", "INIT", "VAR", "IVAR", "FROZENVAR", "DEFINE", "ASSIGN", "INVAR", "INVARSPEC", "LTLSPEC", "FAIRNESS", "JUSTICE", "COMPASSION", "TRANS", "SPEC", "CTLSPEC", "CONSTANTS":
		name = "signal_" + name
	}
	return name
}
func hasPrefix(needle, haystack string) bool { return strings.HasPrefix(haystack, needle) }
func smvNames(n *abc.Network) func(*abc.Object) string {
	names := map[*abc.Object]string{}
	used := map[string]bool{}
	return func(o *abc.Object) string {
		if name := names[o]; name != "" {
			return name
		}
		base := cleanUNSAFE(blifObjectName(o))
		name := base
		for suffix := 2; used[name]; suffix++ {
			name = fmt.Sprintf("%s_%d", base, suffix)
		}
		used[name] = true
		names[o] = name
		return name
	}
}
func WriteSMV(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if n.Type != uint8(abc.ABC_NTK_NETLIST) || n.Function != uint8(abc.ABC_FUNC_SOP) {
		return fmt.Errorf("SMV requires an SOP netlist")
	}
	for _, box := range n.Boxes() {
		if box.Kind != abc.ObjectLatch {
			return fmt.Errorf("SMV requires flattened hierarchy")
		}
	}
	name := smvNames(n)
	var out strings.Builder
	out.WriteString("MODULE main\n")
	if len(n.ObjectsOfKind(abc.ObjectPI)) > 0 {
		out.WriteString("\nVAR -- inputs\n")
	}
	for _, pi := range n.ObjectsOfKind(abc.ObjectPI) {
		net := blifPortNet(pi, true)
		if net == nil {
			return fmt.Errorf("PI lacks net")
		}
		fmt.Fprintf(&out, "    %s : boolean;\n", name(net))
	}
	if len(n.ObjectsOfKind(abc.ObjectLatch)) > 0 {
		out.WriteString("\nVAR -- state variables\n")
	}
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		outputs := latch.Fanouts()
		if len(outputs) != 1 {
			return fmt.Errorf("malformed latch")
		}
		net := blifPortNet(outputs[0], true)
		if net == nil {
			return fmt.Errorf("latch lacks output net")
		}
		fmt.Fprintf(&out, "    %s : boolean;\n", name(net))
	}
	if len(n.ObjectsOfKind(abc.ObjectNode)) > 0 {
		out.WriteString("\nDEFINE\n")
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		outputs := node.Fanouts()
		if len(outputs) != 1 {
			return fmt.Errorf("node must drive one net")
		}
		expression, err := sopExpression(node, name, " & ", " | ", "TRUE", "FALSE")
		if err != nil {
			return err
		}
		fmt.Fprintf(&out, "    %s := %s;\n", name(outputs[0]), expression)
	}
	if len(n.ObjectsOfKind(abc.ObjectLatch)) > 0 {
		out.WriteString("\nASSIGN\n")
	}
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		inputs, outputs := latch.Fanins(), latch.Fanouts()
		if len(inputs) != 1 || len(outputs) != 1 {
			return fmt.Errorf("malformed latch")
		}
		input, output := blifPortNet(inputs[0].Object, false), blifPortNet(outputs[0], true)
		if input == nil || output == nil {
			return fmt.Errorf("latch lacks nets")
		}
		if abc.Abc_LatchIsInit0(latch) != 0 {
			fmt.Fprintf(&out, "    init(%s) := FALSE;\n", name(output))
		} else if abc.Abc_LatchIsInit1(latch) != 0 {
			fmt.Fprintf(&out, "    init(%s) := TRUE;\n", name(output))
		}
		fmt.Fprintf(&out, "    next(%s) := %s;\n", name(output), name(input))
	}
	out.WriteByte('\n')
	for _, po := range n.ObjectsOfKind(abc.ObjectPO) {
		net := blifPortNet(po, false)
		if net == nil {
			return fmt.Errorf("PO lacks net")
		}
		original := net.Name()
		if p := po.Name(); strings.HasPrefix(p, "assume_fair_") || strings.HasPrefix(p, "Assert_") || strings.HasPrefix(p, "assert_") {
			original = p
		}
		switch {
		case strings.HasPrefix(original, "assume_fair_"):
			fmt.Fprintf(&out, "FAIRNESS %s;\n", name(net))
		case strings.HasPrefix(original, "Assert_"), strings.HasPrefix(original, "assert_safety_"):
			fmt.Fprintf(&out, "INVARSPEC %s;\n", name(net))
		case strings.HasPrefix(original, "assert_fair_"):
			fmt.Fprintf(&out, "LTLSPEC G F %s;\n", name(net))
		}
	}
	return writeText(w, out.String())
}
func WriteSMVFile(n *abc.Network, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = WriteSMV(file, n)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Io_WriteSmv(n *abc.Network, filename string) int32 {
	if err := WriteSMVFile(n, filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 0
	}
	return 1
}
func Io_WriteSmvOne(w stdio.Writer, n *abc.Network) int32 {
	if err := WriteSMV(w, n); err != nil {
		return 0
	}
	return 1
}
func Io_WriteSmvCheckNames(n *abc.Network) int32 {
	if n == nil {
		return 0
	}
	return 1
}
func init() { RegisterFormat(IO_FILE_SMV, nil, WriteSMVFile) }
