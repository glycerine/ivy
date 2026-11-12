package io

import (
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
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
	fmt.Fprintln(&out, ".model", n.Name)
	for _, kind := range []abc.ObjectKind{abc.ObjectPI, abc.ObjectPO} {
		directive := ".inputs"
		if kind == abc.ObjectPO {
			directive = ".outputs"
		}
		out.WriteString(directive)
		for _, port := range n.ObjectsOfKind(kind) {
			fmt.Fprint(&out, " ", blifPortName(port, kind == abc.ObjectPI))
		}
		out.WriteByte('\n')
	}
	if n.Function == uint8(abc.ABC_FUNC_BLACKBOX) {
		out.WriteString(".blackbox\n.end\n")
		return writeText(w, out.String())
	}
	if mv {
		for _, net := range n.ObjectsOfKind(abc.ObjectNet) {
			if variable := abc.Abc_ObjMvVar(net); variable != nil {
				fmt.Fprintf(&out, ".mv %s %d", blifObjectName(net), variable.Values)
				for _, label := range variable.Names {
					fmt.Fprint(&out, " ", label)
				}
				out.WriteByte('\n')
			}
		}
	}
	if writeLatches {
		for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
			fans := latch.Fanins()
			outs := latch.Fanouts()
			if len(fans) != 1 || len(outs) != 1 {
				return fmt.Errorf("malformed latch %d", latch.ID)
			}
			input, output := blifPortNet(fans[0].Object, false), blifPortNet(outs[0], true)
			if input == nil || output == nil {
				return fmt.Errorf("latch is not connected through nets")
			}
			init := 2
			if abc.Abc_LatchIsInit0(latch) != 0 {
				init = 0
			} else if abc.Abc_LatchIsInit1(latch) != 0 {
				init = 1
			}
			fmt.Fprintf(&out, ".latch %s %s %d\n", blifObjectName(input), blifObjectName(output), init)
		}
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		outputs := node.Fanouts()
		if len(outputs) != 1 {
			return fmt.Errorf("node %d must drive one net", node.ID)
		}
		out.WriteString(".names")
		for _, e := range node.Fanins() {
			if e.Complement {
				return fmt.Errorf("complemented netlist edge requires explicit inverter")
			}
			fmt.Fprint(&out, " ", blifObjectName(e.Object))
		}
		fmt.Fprintln(&out, " "+blifObjectName(outputs[0]))
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
				fmt.Fprintln(&out, ".default", term)
			}
			for _, cube := range data.Cubes {
				if len(cube.Inputs) != len(node.Fanins()) {
					return fmt.Errorf("MV table arity mismatch")
				}
				for i, input := range cube.Inputs {
					term, err := blifMVTerm(input, node.Fanins()[i].Object, node.Fanins())
					if err != nil {
						return err
					}
					fmt.Fprint(&out, term, " ")
				}
				term, err := blifMVTerm(cube.Output, outputs[0], node.Fanins())
				if err != nil {
					return err
				}
				fmt.Fprintln(&out, term)
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
		fmt.Fprint(&out, ".subckt ", model.Name)
		if mv {
			fmt.Fprint(&out, " ", blifObjectName(box))
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
			fmt.Fprintf(&out, " %s=%s", blifPortName(port, true), blifObjectName(net))
		}
		for i, port := range outputs {
			net := blifPortNet(outs[i], true)
			if net == nil {
				return fmt.Errorf("box output has no net")
			}
			fmt.Fprintf(&out, " %s=%s", blifPortName(port, false), blifObjectName(net))
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
	err = WriteBLIF(f, n, multiValued, writeLatches)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Io_WriteBlif(n *abc.Network, filename string, writeLatches, bb2wb, seq int32) error {
	return WriteBLIFFile(n, filename, false, writeLatches != 0)
}
func Io_WriteBlifLogic(n *abc.Network, filename string, writeLatches int32) error {
	return WriteBLIFFile(n, filename, false, writeLatches != 0)
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
