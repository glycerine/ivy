package io

import (
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
)

func blifPortNet(port *abc.Object, input bool) *abc.Object {
	if input {
		for _, o := range port.Fanouts() {
			if o.Kind == abc.ObjectNet {
				return o
			}
		}
	} else {
		for _, e := range port.Fanins() {
			if e.Object.Kind == abc.ObjectNet {
				return e.Object
			}
		}
	}
	return nil
}

// blifCName prints a network without a name (pName == NULL, e.g. from
// Abc_NtkFromAigPhase) as glibc prints a NULL %s.
func blifCName(name string) string {
	if name == "" {
		return "(null)"
	}
	return name
}
func blifObjectName(o *abc.Object) string {
	if name := o.Name(); name != "" {
		return name
	}
	return fmt.Sprintf("n%d", o.ID)
}
func blifPortName(port *abc.Object, input bool) string {
	if net := blifPortNet(port, input); net != nil {
		return blifObjectName(net)
	}
	return blifObjectName(port)
}
func writeBLIFModel(w stdio.Writer, n *abc.Network, mv bool, writeLatches bool) error {
	var out strings.Builder
	if done, err := writeBLIFHead(&out, n, mv, writeLatches); err != nil || done {
		if err != nil {
			return err
		}
		return writeText(w, out.String())
	}
	return writeBLIFBody(w, &out, n, mv)
}

// writeBLIFHead writes the model, I/O and latch lines; done reports a
// black box, which has nothing more.
func writeBLIFHead(out *strings.Builder, n *abc.Network, mv bool, writeLatches bool) (done bool, err error) {
	fmt.Fprintln(out, ".model", blifCName(n.Name))
	for _, kind := range []abc.ObjectKind{abc.ObjectPI, abc.ObjectPO} {
		directive := ".inputs"
		if kind == abc.ObjectPO {
			directive = ".outputs"
		}
		out.WriteString(directive)
		var names []string
		for _, port := range n.ObjectsOfKind(kind) {
			names = append(names, blifPortName(port, kind == abc.ObjectPI))
		}
		writeBlifNames(out, names, len(directive), 0)
		out.WriteByte('\n')
	}
	if n.Function == uint8(abc.ABC_FUNC_BLACKBOX) {
		out.WriteString(".blackbox\n.end\n")
		return true, nil
	}
	if mv {
		for _, net := range n.ObjectsOfKind(abc.ObjectNet) {
			if variable := abc.Abc_ObjMvVar(net); variable != nil {
				fmt.Fprintf(out, ".mv %s %d", blifObjectName(net), variable.Values)
				for _, label := range variable.Names {
					fmt.Fprint(out, " ", label)
				}
				out.WriteByte('\n')
			}
		}
	}
	latches := n.ObjectsOfKind(abc.ObjectLatch)
	if writeLatches && len(latches) > 0 {
		out.WriteString("\n")
	}
	if writeLatches {
		for _, latch := range latches {
			fans := latch.Fanins()
			outs := latch.Fanouts()
			if len(fans) != 1 || len(outs) != 1 {
				return false, fmt.Errorf("malformed latch %d", latch.ID)
			}
			input, output := blifPortNet(fans[0].Object, false), blifPortNet(outs[0], true)
			if input == nil || output == nil {
				return false, fmt.Errorf("latch is not connected through nets")
			}
			// Io_NtkWriteLatch writes the reset value minus one.
			init := -1
			switch {
			case abc.Abc_LatchIsInit0(latch) != 0:
				init = 0
			case abc.Abc_LatchIsInit1(latch) != 0:
				init = 1
			case abc.Abc_LatchIsInitDc(latch) != 0:
				init = 2
			}
			fmt.Fprintf(out, ".latch %10s %10s  %d\n", blifObjectName(input), blifObjectName(output), init)
		}
		if len(latches) > 0 {
			out.WriteString("\n")
		}
	}
	return false, nil
}

// blifWriteNodeGate is Io_NtkWriteNodeGate: the rest of a mapped node's
// .gate line. It reports whether the next node (the second output of a
// two-output gate) was written too.
var blifTwinReported bool

func blifWriteNodeGate(out *strings.Builder, node *abc.Object, length int) bool {
	gate := node.Data.(*mio.Mio_Gate_t)
	fmt.Fprintf(out, " %-*s ", length, mio.Mio_GateReadName(gate))
	i := 0
	for pin := mio.Mio_GateReadPins(gate); pin != nil; pin = mio.Mio_PinReadNext(pin) {
		fmt.Fprintf(out, "%s=%s ", mio.Mio_PinReadName(pin), blifObjectName(node.FaninAt(int32(i)).Object))
		i++
	}
	fmt.Fprintf(out, "%s=%s", mio.Mio_GateReadOutName(gate), blifObjectName(node.FanoutAt(0)))
	if mio.Mio_GateReadTwin(gate) == nil {
		return false
	}
	node2 := abc.Abc_NtkFetchTwinNode(node)
	if node2 == nil {
		if !blifTwinReported {
			blifTwinReported = true
			fmt.Printf("Warning: Missing second output of gate(s) \"%s\".\n", mio.Mio_GateReadName(gate))
		}
		return false
	}
	fmt.Fprintf(out, " %s=%s", mio.Mio_GateReadOutName(node2.Data.(*mio.Mio_Gate_t)), blifObjectName(node2.FanoutAt(0)))
	return true
}

func writeBLIFBody(w stdio.Writer, out *strings.Builder, n *abc.Network, mv bool) error {
	if abc.Abc_NtkHasMapping(n) != 0 {
		// Io_NtkWriteNode on a mapped netlist: .gate (or .barbuf) lines
		length := int(mio.Mio_LibraryReadGateNameMax(n.FunctionManager.(*mio.Mio_Library_t)))
		nodes := n.ObjectsOfKind(abc.ObjectNode)
		for i := 0; i < len(nodes); i++ {
			node := nodes[i]
			if abc.Abc_ObjIsBarBuf(node) != 0 {
				fmt.Fprintf(out, ".barbuf %s %s\n", blifObjectName(node.FaninAt(0).Object), blifObjectName(node.FanoutAt(0)))
				continue
			}
			out.WriteString(".gate")
			skip := blifWriteNodeGate(out, node, length)
			out.WriteString("\n")
			if skip { // skip the next node
				i++
			}
		}
	}
	namesNodes := n.ObjectsOfKind(abc.ObjectNode)
	if abc.Abc_NtkHasMapping(n) != 0 {
		namesNodes = nil
	}
	for _, node := range namesNodes {
		outputs := node.Fanouts()
		if len(outputs) != 1 {
			return fmt.Errorf("node %d must drive one net", node.ID)
		}
		out.WriteString(".names")
		var names []string
		for _, e := range node.Fanins() {
			if e.Complement {
				return fmt.Errorf("complemented netlist edge requires explicit inverter")
			}
			names = append(names, blifObjectName(e.Object))
		}
		length, count := writeBlifNames(out, names, 6, 0)
		// The output name uses a 75-column limit (Io_NtkWriteNodeFanins).
		if name := blifObjectName(outputs[0]); count > 0 && length+len(name)+1 > 75 {
			out.WriteString(" \\\n " + name + "\n")
		} else {
			out.WriteString(" " + name + "\n")
		}
		switch data := node.Data.(type) {
		case string:
			out.WriteString(strings.TrimRight(data, "\x00"))
			if !strings.HasSuffix(data, "\n") {
				out.WriteByte('\n')
			}
		case *abc.MVTable:
			if !mv {
				return fmt.Errorf("MV node requires BLIF-MV output")
			}
			if data.Default != nil {
				term, err := blifMVTerm(*data.Default, outputs[0], node.Fanins())
				if err != nil {
					return err
				}
				fmt.Fprintln(out, ".default", term)
			}
			for _, cube := range data.Cubes {
				if len(cube.Inputs) != node.FaninNum() {
					return fmt.Errorf("MV table arity mismatch")
				}
				for i, input := range cube.Inputs {
					term, err := blifMVTerm(input, node.FaninAt(int32(i)).Object, node.Fanins())
					if err != nil {
						return err
					}
					fmt.Fprint(out, term, " ")
				}
				term, err := blifMVTerm(cube.Output, outputs[0], node.Fanins())
				if err != nil {
					return err
				}
				fmt.Fprintln(out, term)
			}

		default:
			return fmt.Errorf("node %d has no supported SOP/MV function", node.ID)
		}
	}
	for _, box := range n.Boxes() {
		if box.Kind == abc.ObjectLatch {
			continue
		}
		model, ok := box.Data.(*abc.Network)
		if !ok || model == nil {
			return fmt.Errorf("box %d has no model", box.ID)
		}
		fmt.Fprint(out, ".subckt ", model.Name)
		if mv {
			fmt.Fprint(out, " ", blifObjectName(box))
		}
		inputs, outputs := model.ObjectsOfKind(abc.ObjectPI), model.ObjectsOfKind(abc.ObjectPO)
		fans, outs := box.Fanins(), box.Fanouts()
		if len(inputs) != len(fans) || len(outputs) != len(outs) {
			return fmt.Errorf("box interface mismatch")
		}
		for i, port := range inputs {
			net := blifPortNet(fans[i].Object, false)
			if net == nil {
				return fmt.Errorf("box input has no net")
			}
			fmt.Fprintf(out, " %s=%s", blifPortName(port, true), blifObjectName(net))
		}
		for i, port := range outputs {
			net := blifPortNet(outs[i], true)
			if net == nil {
				return fmt.Errorf("box output has no net")
			}
			fmt.Fprintf(out, " %s=%s", blifPortName(port, false), blifObjectName(net))
		}
		out.WriteByte('\n')
	}
	out.WriteString(".end\n")
	return writeText(w, out.String())
}
func WriteBLIF(w stdio.Writer, n *abc.Network, multiValued, writeLatches bool) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	models := []*abc.Network{n}
	if n.Design != nil {
		for _, model := range n.Design.Modules {
			if model != n {
				models = append(models, model)
			}
		}
	}
	for _, model := range models {
		if err := writeBLIFModel(w, model, multiValued, writeLatches); err != nil {
			return err
		}
	}
	return nil
}
func WriteBLIFFile(n *abc.Network, filename string, multiValued, writeLatches bool) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	// Io_WriteBlif starts every file with this comment.
	err = writeText(f, fmt.Sprintf("# Benchmark \"%s\" written by ABC on %s\n", blifCName(n.Name), extra.Extra_TimeStamp()))
	if err == nil {
		err = WriteBLIF(f, n, multiValued, writeLatches)
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Io_WriteBlif(n *abc.Network, filename string, writeLatches, bb2wb, seq int32) error {
	return WriteBLIFFile(n, filename, false, writeLatches != 0)
}

func writeNetworkBLIF(n *abc.Network, filename string) error {
	if abc.Abc_NtkIsNetlist(n) == 0 {
		n = abc.Abc_NtkToNetlist(n)
		if n == nil {
			return fmt.Errorf("Writing BLIF has failed.") // Io_WriteBlifLogic
		}
		defer abc.Abc_NtkDelete(n)
		if abc.Abc_NtkHasSop(n) == 0 && abc.Abc_NtkHasMapping(n) == 0 && abc.Abc_NtkToSop(n, -1, 1_000_000_000) == 0 {
			return fmt.Errorf("io: cannot convert network to SOP")
		}
	}
	return Io_WriteBlif(n, filename, 1, 0, 0)
}

// Io_WriteBlifLogic derives the netlist first, as ABC does.
func Io_WriteBlifLogic(n *abc.Network, filename string, writeLatches int32) error {
	netlist := abc.Abc_NtkToNetlist(n)
	if netlist == nil {
		return fmt.Errorf("Writing BLIF has failed.")
	}
	defer abc.Abc_NtkDelete(netlist)
	return Io_WriteBlif(netlist, filename, writeLatches, 0, 0)
}

// blifLineLength is IO_WRITE_LINE_LENGTH.
const blifLineLength = 78

// writeBlifNames writes " name" for each name, continuing with " \" before
// a name that would pass ABC's line length (Io_NtkWritePis/Pos/NodeFanins).
// length is the width already used on the line; the updated width and
// count of names on the current line are returned.
func writeBlifNames(out *strings.Builder, names []string, length, count int) (int, int) {
	for _, name := range names {
		added := len(name) + 1
		if count > 0 && length+added+3 > blifLineLength {
			out.WriteString(" \\\n")
			length, count = 0, 0
		}
		out.WriteString(" " + name)
		length += added
		count++
	}
	return length, count
}

func blifMVTerm(term abc.MVTerm, net *abc.Object, inputs []abc.Edge) (string, error) {
	if term.Any {
		return "-", nil
	}
	if term.EqualInput >= 0 {
		if term.EqualInput >= len(inputs) {
			return "", fmt.Errorf("MV equality index outside inputs")
		}
		return "=" + blifObjectName(inputs[term.EqualInput].Object), nil
	}
	variable := abc.Abc_ObjMvVar(net)
	values := make([]string, len(term.Values))
	for i, value := range term.Values {
		if value < 0 || value >= int(abc.Abc_ObjMvVarNum(net)) {
			return "", fmt.Errorf("MV value outside domain")
		}
		values[i] = fmt.Sprint(value)
		if variable != nil && len(variable.Names) > 0 {
			if value >= len(variable.Names) {
				return "", fmt.Errorf("MV label table incomplete")
			}
			values[i] = variable.Names[value]
		}
	}
	if len(values) == 0 {
		return "", fmt.Errorf("empty MV value set")
	}
	if len(values) == 1 {
		return values[0], nil
	}
	return "(" + strings.Join(values, ",") + ")", nil
}

// WriteBLIFSpecialFile writes Io_WriteBlifInt's file: the comment, model,
// I/O and latch lines of netlist n, then the node section that body writes
// (name is Abc_ObjName of a net).
func WriteBLIFSpecialFile(n *abc.Network, filename string, body func(out *strings.Builder, name func(*abc.Object) string)) error {
	var out strings.Builder
	fmt.Fprintf(&out, "# Benchmark \"%s\" written by ABC on %s\n", n.Name, extra.Extra_TimeStamp())
	if _, err := writeBLIFHead(&out, n, false, true); err != nil {
		return err
	}
	body(&out, blifObjectName)
	return os.WriteFile(filename, []byte(out.String()), 0o644)
}
