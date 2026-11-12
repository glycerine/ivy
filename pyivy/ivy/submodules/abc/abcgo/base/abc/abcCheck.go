package abc

import (
	"github.com/glycerine/abcgo/misc/nm"
	"sort"
	"strings"
)

func Abc_ObjNameNet(o *Object) string {
	if o.Kind == ObjectNode && o.network.Type == uint8(ABC_NTK_NETLIST) && len(o.fanouts) > 0 {
		return o.fanouts[0].Name()
	}
	return o.Name()
}

// Go's library check entry points always validate. They do not depend on a
// process-global CLI flag or overwrite user Copy fields as scratch memory.
func Abc_NtkCheck(n *Network) int32     { return Abc_NtkDoCheck(n) }
func Abc_NtkCheckRead(n *Network) int32 { return Abc_NtkDoCheck(n) }
func Abc_NtkDoCheck(n *Network) int32 {
	if n == nil || n.Type < uint8(ABC_NTK_NETLIST) || n.Type > uint8(ABC_NTK_STRASH) || n.Function < uint8(ABC_FUNC_SOP) || n.Function > uint8(ABC_FUNC_BLACKBOX) {
		return 0
	}
	if n.Check() != nil || Abc_NtkCheckPis(n) == 0 || Abc_NtkCheckPos(n) == 0 {
		return 0
	}
	if !checkCombinedTerminals(n, n.cis, true) || !checkCombinedTerminals(n, n.cos, false) {
		return 0
	}
	if Abc_NtkCheckNames(n) == 0 {
		return 0
	}
	onlyLatches := true
	for _, box := range n.Boxes() {
		if box.Kind != ObjectLatch {
			onlyLatches = false
		}
	}
	if onlyLatches && (len(n.cis) != len(n.ObjectsOfKind(ObjectPI))+len(n.Boxes()) || len(n.cos) != len(n.ObjectsOfKind(ObjectPO))+len(n.Boxes())) {
		return 0
	}
	if n.Function == uint8(ABC_FUNC_BLACKBOX) {
		return 1
	}
	for _, o := range n.Objects() {
		if Abc_NtkCheckObj(n, o) == 0 {
			return 0
		}
	}
	nets := n.ObjectsOfKind(ObjectNet)
	if n.Type != uint8(ABC_NTK_NETLIST) && len(nets) > 0 {
		return 0
	}
	for _, net := range nets {
		if Abc_NtkCheckNet(n, net) == 0 {
			return 0
		}
	}
	if n.Type == uint8(ABC_NTK_STRASH) {
		if Abc_AigCheck(Abc_AigAlloc(n)) == 0 {
			return 0
		}
	} else {
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			if Abc_NtkCheckNode(n, o) == 0 {
				return 0
			}
		}
	}
	for _, latch := range n.ObjectsOfKind(ObjectLatch) {
		if Abc_NtkCheckLatch(n, latch) == 0 {
			return 0
		}
	}
	return 1
}
func checkCombinedTerminals(n *Network, objects []*Object, input bool) bool {
	seen := map[*Object]bool{}
	for _, o := range objects {
		if !n.contains(o) || seen[o] || (input && !o.IsCI()) || (!input && !o.IsCO()) {
			return false
		}
		seen[o] = true
	}
	for _, o := range n.Objects() {
		if ((input && o.IsCI()) || (!input && o.IsCO())) && !seen[o] {
			return false
		}
	}
	return true
}
func checkKindOrder(n *Network, kind ObjectKind) bool {
	seen := map[*Object]bool{}
	for _, o := range n.ObjectsOfKind(kind) {
		if !n.contains(o) || seen[o] || o.Kind != kind {
			return false
		}
		seen[o] = true
	}
	for _, o := range n.Objects() {
		if o.Kind == kind && !seen[o] {
			return false
		}
	}
	return true
}
func Abc_NtkCheckNames(n *Network) int32 {
	if n.Type == uint8(ABC_NTK_NETLIST) {
		return 1
	}
	for _, o := range append(n.Inputs(), n.Outputs()...) {
		if !n.contains(o) {
			return 0
		}
		if _, exists := n.names.Lookup(int32(o.ID)); !exists {
			return 0
		}
	}
	for _, id := range nm.Nm_ManReturnNameIds(n.names) {
		if n.Object(int(id)) == nil {
			return 0
		}
	}
	return abcBool(Abc_NtkCheckUniqueCiNames(n) != 0 && Abc_NtkCheckUniqueCoNames(n) != 0 && Abc_NtkCheckUniqueCioNames(n) != 0)
}
func Abc_NtkCheckPis(n *Network) int32 {
	if !checkKindOrder(n, ObjectPI) {
		return 0
	}
	for _, o := range n.ObjectsOfKind(ObjectPI) {
		if o.Data != nil || len(o.fanins) != 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkCheckPos(n *Network) int32 {
	if !checkKindOrder(n, ObjectPO) {
		return 0
	}
	for _, o := range n.ObjectsOfKind(ObjectPO) {
		if o.Data != nil || len(o.fanins) != 1 || len(o.fanouts) != 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkCheckObj(n *Network, o *Object) int32 {
	if !n.contains(o) {
		return 0
	}
	incoming, outgoing := map[*Object]int{}, map[*Object]int{}
	for _, e := range o.fanins {
		if !n.contains(e.Object) {
			return 0
		}
		incoming[e.Object]++
	}
	for source, count := range incoming {
		found := 0
		for _, fanout := range source.fanouts {
			if fanout == o {
				found++
			}
		}
		if count != found {
			return 0
		}
	}
	for _, fanout := range o.fanouts {
		if !n.contains(fanout) {
			return 0
		}
		outgoing[fanout]++
	}
	for fanout, count := range outgoing {
		found := 0
		for _, e := range fanout.fanins {
			if e.Object == o {
				found++
			}
		}
		if count != found {
			return 0
		}
	}
	return 1
}
func Abc_NtkCheckNet(n *Network, o *Object) int32 {
	return abcBool(n.contains(o) && o.Kind == ObjectNet && len(o.fanins) == 1)
}

// NodeFunctionValidator lets functionality managers validate their own handles
// without introducing dependencies from base/abc back to every manager package.
type NodeFunctionValidator interface{ ValidateFanins(int) error }

func Abc_NtkCheckNode(n *Network, o *Object) int32 {
	if !n.contains(o) || o.Kind != ObjectNode {
		return 0
	}
	if n.Type == uint8(ABC_NTK_NETLIST) {
		if len(o.fanouts) == 0 {
			return 0
		}
		for _, e := range o.fanins {
			if e.Object.Kind != ObjectNet {
				return 0
			}
		}
		for _, fanout := range o.fanouts {
			if fanout.Kind != ObjectNet {
				return 0
			}
		}
	}
	if isBarrierBuffer(o) {
		return 1
	}
	if o.Data == nil {
		return 0
	}
	if checker, ok := o.Data.(NodeFunctionValidator); ok {
		return abcBool(checker.ValidateFanins(len(o.fanins)) == nil)
	}
	if n.Function == uint8(ABC_FUNC_BLIFMV) {
		_, err := nodeMVTable(o)
		return abcBool(err == nil)
	}
	if n.Function == uint8(ABC_FUNC_SOP) {
		if text, ok := o.Data.(string); ok {
			if text == aigANDMarker {
				return abcBool(len(o.fanins) == 2)
			}
			return abcBool(checkSOPSyntax(text, len(o.fanins)))
		}
		_, ok := o.Data.(func([]bool) bool)
		return abcBool(ok)
	}
	return 1
}
func latchInitValue(o *Object) int {
	switch value := o.Data.(type) {
	case Abc_InitType_t:
		return int(value)
	case int:
		return value
	case int32:
		return int(value)
	case uint32:
		return int(value)
	default:
		return -1
	}
}
func Abc_NtkCheckLatch(n *Network, o *Object) int32 {
	init := latchInitValue(o)
	if !n.contains(o) || o.Kind != ObjectLatch || init < int(ABC_INIT_ZERO) || init > int(ABC_INIT_DC) || len(o.fanins) != 1 || len(o.fanouts) != 1 {
		return 0
	}
	input, output := o.fanins[0].Object, o.fanouts[0]
	return abcBool(input.Kind == ObjectBI && output.Kind == ObjectBO && len(input.fanins) == 1 && len(input.fanouts) == 1 && len(output.fanins) == 1)
}
func compareNamedObjects(a, b []*Object) int32 {
	if len(a) != len(b) {
		return 0
	}
	for i, o := range a {
		if o.Name() != b[i].Name() {
			return 0
		}
	}
	return 1
}
func Abc_NtkComparePis(a, b *Network, _ int32) int32 {
	return compareNamedObjects(a.ObjectsOfKind(ObjectPI), b.ObjectsOfKind(ObjectPI))
}
func Abc_NtkComparePos(a, b *Network, _ int32) int32 {
	return compareNamedObjects(a.ObjectsOfKind(ObjectPO), b.ObjectsOfKind(ObjectPO))
}
func Abc_NtkCompareBoxes(a, b *Network, combinational int32) int32 {
	if combinational == 0 {
		return 1
	}
	as, bs := a.Boxes(), b.Boxes()
	if len(as) != len(bs) {
		return 0
	}
	for i, o := range as {
		if o.Kind != ObjectLatch || bs[i].Kind != ObjectLatch || len(o.fanouts) != 1 || len(bs[i].fanouts) != 1 || o.fanouts[0].Name() != bs[i].fanouts[0].Name() {
			return 0
		}
	}
	return 1
}
func orderInterfacesForComparison(n *Network, combinational bool) {
	for _, kind := range []ObjectKind{ObjectPI, ObjectPO} {
		objects := n.ObjectsOfKind(kind)
		sort.SliceStable(objects, func(i, j int) bool { return objects[i].Name() < objects[j].Name() })
		ids := make([]int, len(objects))
		for i, o := range objects {
			ids[i] = o.ID
		}
		if err := n.SetTerminalOrder(kind, ids); err != nil {
			panic(err)
		}
	}
	if combinational {
		sort.SliceStable(n.boxes, func(i, j int) bool {
			a, b := n.boxes[i], n.boxes[j]
			an, bn := a.Name(), b.Name()
			if len(a.fanouts) > 0 {
				an = a.fanouts[0].Name()
			}
			if len(b.fanouts) > 0 {
				bn = b.fanouts[0].Name()
			}
			if an == bn {
				return a.ID < b.ID
			}
			return an < bn
		})
	}
	for _, kind := range []ObjectKind{ObjectLatch, ObjectWhitebox, ObjectBlackbox} {
		n.kindOrder[kind] = nil
	}
	for _, box := range n.boxes {
		n.kindOrder[box.Kind] = append(n.kindOrder[box.Kind], box)
	}
	n.cis = n.ObjectsOfKind(ObjectPI)
	n.cos = n.ObjectsOfKind(ObjectPO)
	for _, box := range n.boxes {
		for _, input := range box.fanins {
			n.cos = append(n.cos, input.Object)
		}
		n.cis = append(n.cis, box.fanouts...)
	}
}
func Abc_NtkCompareSignals(a, b *Network, onlyInputs, combinational int32) int32 {
	orderInterfacesForComparison(a, combinational != 0)
	orderInterfacesForComparison(b, combinational != 0)
	if Abc_NtkComparePis(a, b, combinational) == 0 {
		return 0
	}
	if onlyInputs != 0 {
		return 1
	}
	return abcBool(Abc_NtkCompareBoxes(a, b, combinational) != 0 && Abc_NtkComparePos(a, b, combinational) != 0)
}
func Abc_NtkIsAcyclicHierarchy_rec(n *Network) int32 {
	state := map[*Network]uint8{}
	var visit func(*Network) bool
	visit = func(model *Network) bool {
		if state[model] == 1 {
			return false
		}
		if state[model] == 2 {
			return true
		}
		state[model] = 1
		if model.Function != uint8(ABC_FUNC_BLACKBOX) {
			for _, box := range model.Boxes() {
				if box.Kind == ObjectLatch {
					continue
				}
				child, ok := box.Data.(*Network)
				if !ok || child == nil || !visit(child) {
					return false
				}
			}
		}
		state[model] = 2
		return true
	}
	return abcBool(visit(n))
}
func Abc_NtkIsAcyclicHierarchy(n *Network) int32 { return Abc_NtkIsAcyclicHierarchy_rec(n) }
func Abc_NtkNamesCompare(a, b *string) int32     { return int32(strings.Compare(*a, *b)) }
func uniqueNames(objects []*Object) int32 {
	seen := map[string]bool{}
	for _, o := range objects {
		name := o.Name()
		if seen[name] {
			return 0
		}
		seen[name] = true
	}
	return 1
}
func Abc_NtkCheckUniqueCiNames(n *Network) int32 { return uniqueNames(n.Inputs()) }
func Abc_NtkCheckUniqueCoNames(n *Network) int32 { return uniqueNames(n.Outputs()) }
func Abc_NtkCheckUniqueCioNames(n *Network) int32 {
	names := map[string]*Object{}
	for _, o := range n.Inputs() {
		names[o.Name()] = o
	}
	for _, o := range n.Outputs() {
		if input := names[o.Name()]; input != nil {
			if len(o.fanins) != 1 || o.fanins[0].Object != input {
				return 0
			}
		}
	}
	return 1
}

func checkSOPSyntax(text string, nVars int) bool {
	text = strings.TrimRight(text, "\x00")
	zero, one := false, false
	for text != "" {
		end := strings.IndexByte(text, '\n')
		if end < 0 {
			return false
		}
		line := text[:end]
		text = text[end+1:]
		if len(line) != nVars+2 || line[nVars] != ' ' {
			return false
		}
		for _, c := range line[:nVars] {
			if c != '0' && c != '1' && c != '-' {
				return false
			}
		}
		switch line[nVars+1] {
		case '0':
			zero = true
		case '1':
			one = true
		case 'x', 'n':
		default:
			return false
		}
	}
	return !zero || !one
}
