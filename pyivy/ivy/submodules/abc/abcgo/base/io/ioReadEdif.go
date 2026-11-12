// EDIF s-expression reader. Connectivity comes from joined port references,
// independent of whitespace, line layout, declaration order, or net names.
package io

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

type edifPort struct{ name, direction string }
type edifCell struct {
	key, name, library string
	ports              []edifPort
	contents           *edifExpr
	model              *blifModel
}
type edifInstance struct {
	name string
	cell *edifCell
	pins map[string]string
}

func edifPorts(cell *edifExpr) ([]edifPort, *edifExpr, error) {
	var view *edifExpr
	for _, candidate := range cell.children("view") {
		if candidate.child("interface") != nil {
			view = candidate
			break
		}
	}
	if view == nil {
		return nil, nil, fmt.Errorf("cell has no interface view")
	}
	var ports []edifPort
	seen := map[string]bool{}
	for _, port := range view.child("interface").children("port") {
		if len(port.list) < 2 {
			return nil, nil, fmt.Errorf("port missing name")
		}
		id, err := edifIdentifier(port.list[1])
		if err != nil {
			return nil, nil, err
		}
		direction := port.child("direction")
		if direction == nil || len(direction.list) != 2 {
			return nil, nil, fmt.Errorf("port %s missing direction", id)
		}
		kind := strings.ToUpper(direction.list[1].atom)
		if kind != "INPUT" && kind != "OUTPUT" {
			return nil, nil, fmt.Errorf("unsupported EDIF port direction %s", kind)
		}
		count := 1
		if port.list[1].head() == "array" {
			array := port.list[1]
			if len(array.list) != 3 {
				return nil, nil, fmt.Errorf("invalid EDIF array")
			}
			count, err = strconv.Atoi(array.list[2].atom)
			if err != nil || count <= 0 || count > 1<<20 {
				return nil, nil, fmt.Errorf("invalid EDIF array size")
			}
		}
		for i := 0; i < count; i++ {
			name := id
			if port.list[1].head() == "array" {
				name = fmt.Sprintf("%s[%d]", id, i)
			}
			if seen[name] {
				return nil, nil, fmt.Errorf("duplicate EDIF port %s", name)
			}
			seen[name] = true
			ports = append(ports, edifPort{name, kind})
		}
	}
	return ports, view.child("contents"), nil
}
func edifCellRef(expression *edifExpr, defaultLibrary string, cells map[string]*edifCell) (*edifCell, error) {
	ref := expression.child("cellref")
	if ref == nil {
		if view := expression.child("viewref"); view != nil {
			ref = view.child("cellref")
		}
	}
	if ref == nil || len(ref.list) < 2 {
		return nil, fmt.Errorf("missing EDIF cellRef")
	}
	name, err := edifIdentifier(ref.list[1])
	if err != nil {
		return nil, err
	}
	library := defaultLibrary
	if lib := ref.child("libraryref"); lib != nil && len(lib.list) >= 2 {
		library, err = edifIdentifier(lib.list[1])
		if err != nil {
			return nil, err
		}
	}
	cell := cells[library+"/"+name]
	if cell == nil {
		return nil, fmt.Errorf("unknown EDIF cell %s/%s", library, name)
	}
	return cell, nil
}
func edifGate(name string) string {
	upper := strings.ToUpper(name)
	for _, pair := range [][2]string{{"EXNOR", "XNOR"}, {"EXOR", "XOR"}, {"XNOR", "XNOR"}, {"XOR", "XOR"}, {"NAND", "NAND"}, {"NOR", "NOR"}, {"AND", "AND"}, {"OR", "OR"}, {"INV", "NOT"}, {"NOT", "NOT"}, {"BUF", "BUF"}, {"FLIP", "DFF"}, {"DFF", "DFF"}} {
		if strings.HasPrefix(upper, pair[0]) {
			return pair[1]
		}
	}
	return ""
}

type edifUnion map[string]string

func (u edifUnion) find(name string) string {
	parent, ok := u[name]
	if !ok {
		u[name] = name
		return name
	}
	if parent != name {
		u[name] = u.find(parent)
	}
	return u[name]
}
func (u edifUnion) join(a, b string) { u[u.find(a)] = u.find(b) }
func buildEdifContents(cell *edifCell, cells map[string]*edifCell) error {
	m := cell.model
	if cell.contents == nil {
		m.network.Function = uint8(abc.ABC_FUNC_BLACKBOX)
		return nil
	}
	instances := map[string]*edifInstance{}
	var ordered []*edifInstance
	for _, form := range cell.contents.children("instance") {
		if len(form.list) < 2 {
			return fmt.Errorf("instance missing name")
		}
		name, err := edifIdentifier(form.list[1])
		if err != nil {
			return err
		}
		if instances[name] != nil {
			return fmt.Errorf("duplicate instance %s", name)
		}
		model, err := edifCellRef(form, cell.library, cells)
		if err != nil {
			return err
		}
		instance := &edifInstance{name: name, cell: model, pins: map[string]string{}}
		instances[name] = instance
		ordered = append(ordered, instance)
	}
	union := edifUnion{}
	for _, port := range cell.ports {
		union.find("p:" + port.name)
	}
	for _, instance := range ordered {
		for _, port := range instance.cell.ports {
			union.find("i:" + instance.name + ":" + port.name)
		}
	}
	var netNames []string
	for _, form := range cell.contents.children("net") {
		if len(form.list) < 2 {
			return fmt.Errorf("net missing name")
		}
		net, err := edifIdentifier(form.list[1])
		if err != nil {
			return err
		}
		netNames = append(netNames, net)
		key := "n:" + net
		joined := form.child("joined")
		if joined == nil {
			return fmt.Errorf("net %s missing joined", net)
		}
		for _, reference := range joined.children("portref") {
			if len(reference.list) < 2 {
				return fmt.Errorf("portRef missing port")
			}
			port, err := edifIdentifier(reference.list[1])
			if err != nil {
				return err
			}
			pin := "p:" + port
			if owner := reference.child("instanceref"); owner != nil {
				if len(owner.list) != 2 {
					return fmt.Errorf("invalid instanceRef")
				}
				id, err := edifIdentifier(owner.list[1])
				if err != nil {
					return err
				}
				instance := instances[id]
				if instance == nil {
					return fmt.Errorf("unknown instance %s", id)
				}
				known := false
				for _, candidate := range instance.cell.ports {
					if candidate.name == port {
						known = true
						break
					}
				}
				if !known {
					return fmt.Errorf("unknown instance port %s.%s", id, port)
				}
				pin = "i:" + id + ":" + port
			} else {
				known := false
				for _, candidate := range cell.ports {
					if candidate.name == port {
						known = true
						break
					}
				}
				if !known {
					return fmt.Errorf("unknown top-level port %s", port)
				}
			}
			union.join(key, pin)
		}
	}
	names := map[string]string{}
	for _, port := range cell.ports {
		if port.direction == "INPUT" {
			root := union.find("p:" + port.name)
			if names[root] != "" && names[root] != port.name {
				return fmt.Errorf("multiple input drivers on one net")
			}
			names[root] = port.name
		}
	}
	for _, port := range cell.ports {
		root := union.find("p:" + port.name)
		if names[root] == "" {
			names[root] = port.name
		}
	}
	for _, name := range netNames {
		root := union.find("n:" + name)
		if names[root] == "" {
			names[root] = name
		}
	}
	for _, instance := range ordered {
		for _, port := range instance.cell.ports {
			root := union.find("i:" + instance.name + ":" + port.name)
			signal := names[root]
			if signal == "" {
				signal = m.uniqueNetName("__edif_unconnected")
				names[root] = signal
			}
			instance.pins[port.name] = signal
		}
	}
	for _, instance := range ordered {
		gate := edifGate(instance.cell.name)
		if instance.cell.contents != nil || gate == "" {
			bindings := []string{instance.cell.key}
			for _, port := range instance.cell.ports {
				bindings = append(bindings, port.name+"="+instance.pins[port.name])
			}
			byName := map[string]*blifModel{instance.cell.key: instance.cell.model}
			if err := blifInstance(m, bindings, byName, false); err != nil {
				return err
			}
			continue
		}
		var inputs, outputs []string
		for _, port := range instance.cell.ports {
			if port.direction == "INPUT" {
				if gate == "DFF" && (strings.EqualFold(port.name, "CK") || strings.EqualFold(port.name, "CLK") || strings.EqualFold(port.name, "RESET")) {
					continue
				}
				inputs = append(inputs, instance.pins[port.name])
			} else {
				outputs = append(outputs, instance.pins[port.name])
			}
		}
		if len(outputs) != 1 {
			return fmt.Errorf("primitive %s must have one output", instance.cell.name)
		}
		if gate == "DFF" {
			if len(inputs) != 1 {
				return fmt.Errorf("flip-flop must have one data input")
			}
			if err := blifLatch(m, []string{inputs[0], outputs[0], "0"}); err != nil {
				return err
			}
			continue
		}
		cover, err := benchGateCover(gate, len(inputs))
		if err != nil {
			return err
		}
		if _, err = m.node(append(inputs, outputs[0]), cover); err != nil {
			return err
		}
	}
	for _, port := range cell.ports {
		if port.direction != "OUTPUT" {
			continue
		}
		signal := names[union.find("p:"+port.name)]
		if signal != port.name {
			if _, err := m.node([]string{signal, port.name}, "1 1\n"); err != nil {
				return err
			}
		}
	}
	return nil
}
func ParseEDIF(filename, text string, check bool) (*abc.Network, error) {
	root, err := parseEdifSyntax(filename, text)
	if err != nil {
		return nil, err
	}
	cells := map[string]*edifCell{}
	var ordered []*edifCell
	for _, library := range root.list {
		if library.head() != "library" && library.head() != "external" {
			continue
		}
		if len(library.list) < 2 {
			return nil, fmt.Errorf("%s: library missing name", filename)
		}
		libraryName, err := edifIdentifier(library.list[1])
		if err != nil {
			return nil, err
		}
		for _, form := range library.children("cell") {
			if len(form.list) < 2 {
				return nil, fmt.Errorf("cell missing name")
			}
			name, err := edifIdentifier(form.list[1])
			if err != nil {
				return nil, err
			}
			key := libraryName + "/" + name
			if cells[key] != nil {
				return nil, fmt.Errorf("duplicate EDIF cell %s", key)
			}
			ports, contents, err := edifPorts(form)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: cell %s: %w", filename, form.line, key, err)
			}
			cell := &edifCell{key: key, name: name, library: libraryName, ports: ports, contents: contents}
			cells[key] = cell
			ordered = append(ordered, cell)
		}
	}
	designForm := root.child("design")
	if designForm == nil {
		return nil, fmt.Errorf("%s: EDIF design declaration missing", filename)
	}
	top, err := edifCellRef(designForm, "", cells)
	if err != nil {
		return nil, err
	}
	design := abc.NewDesign(filename)
	for _, cell := range ordered {
		networkName := strings.ReplaceAll(cell.key, "/", "__")
		if cell == top {
			networkName = top.name
		}
		n := abc.NewNetwork(networkName)
		n.Spec = filename
		n.Type = uint8(abc.ABC_NTK_NETLIST)
		n.Function = uint8(abc.ABC_FUNC_SOP)
		if cell.contents == nil {
			n.Function = uint8(abc.ABC_FUNC_BLACKBOX)
		}
		cell.model = &blifModel{network: n, nets: map[string]*abc.Object{}, domains: map[string]*abc.MVVariable{}}
		if err = design.AddModule(n); err != nil {
			return nil, err
		}
	}
	// Interfaces are established for every model before any instance is linked.
	for _, cell := range ordered {
		for _, port := range cell.ports {
			if port.direction == "INPUT" {
				err = cell.model.input(port.name)
			} else {
				err = cell.model.output(port.name)
			}
			if err != nil {
				return nil, err
			}
		}
	}
	for _, cell := range ordered { // buildEdifCell's declarations already exist here.
		if err = buildEdifContents(cell, cells); err != nil {
			return nil, fmt.Errorf("%s: cell %s: %w", filename, cell.key, err)
		}
	}
	design.Tops = []*abc.Network{top.model.network}
	if check {
		for _, cell := range ordered {
			if err = cell.model.network.Check(); err != nil {
				return nil, err
			}
			if cell.model.network.Function != uint8(abc.ABC_FUNC_BLACKBOX) {
				for name, net := range cell.model.nets {
					if len(net.Fanouts()) > 0 && len(net.Fanins()) == 0 {
						return nil, fmt.Errorf("%s: cell %s: undriven net %s", filename, cell.key, name)
					}
				}
				roots := append(cell.model.network.ObjectsOfKind(abc.ObjectPO), cell.model.network.ObjectsOfKind(abc.ObjectBI)...)
				if _, err = cell.model.network.Topological(roots); err != nil {
					return nil, err
				}
			}

		}
	}
	return top.model.network, nil
}
func ReadEDIF(filename string, check bool) (*abc.Network, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseEDIF(filename, string(data), check)
}
func Io_ReadEdif(filename string, check int32) *abc.Network {
	n, err := ReadEDIF(filename, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func init() { RegisterFormat(IO_FILE_EDIF, ReadEDIF, nil) }
