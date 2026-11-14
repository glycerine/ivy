package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Abc_NtkCleanCopy(n *Network) {
	for _, o := range n.Objects() {
		o.Copy, o.CopyEdge, o.CopyData = nil, Edge{}, nil
	}
}

func hierarchyWalk(n *Network, visit func(*Network)) {
	seen := make(map[*Network]bool)
	var walk func(*Network)
	walk = func(current *Network) {
		if current == nil || seen[current] {
			return
		}
		seen[current] = true
		visit(current)
		for _, box := range current.boxes {
			if model, ok := box.Data.(*Network); ok {
				walk(model)
			}
		}
	}
	walk(n)
}

func Abc_NtkCleanCopy_rec(n *Network) { hierarchyWalk(n, Abc_NtkCleanCopy) }
func Abc_NtkCleanData(n *Network) {
	for _, o := range n.Objects() {
		o.Data = nil
	}
}
func Abc_NtkFillTemp(n *Network) {
	for _, o := range n.Objects() {
		o.Temp = -1
	}
}
func Abc_NtkCountCopy(n *Network) int32 {
	var count int32
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if getCopy(o).Object != nil {
			count++
		}
	}
	return count
}
func Abc_NtkSaveCopy(n *Network) *vec.Vec_Ptr_t {
	copies := vec.Vec_PtrStart(len(n.objects))
	for _, o := range n.Objects() {
		vec.Vec_PtrWriteEntry(copies, o.ID, getCopy(o))
	}
	return copies
}
func Abc_NtkLoadCopy(n *Network, copies *vec.Vec_Ptr_t) {
	if copies == nil || vec.Vec_PtrSize(copies) < len(n.objects) {
		panic("abc: incomplete saved copies")
	}
	for _, o := range n.Objects() {
		switch value := vec.Vec_PtrEntry(copies, o.ID).(type) {
		case Edge:
			setCopy(o, value)
		case *Object:
			if value == nil {
				setCopy(o, Edge{})
			} else {
				setCopy(o, value.Signal())
			}
		case nil:
			setCopy(o, Edge{})
		default:
			panic("abc: invalid saved copy")
		}
	}
}
func Abc_NtkCleanNext(n *Network) {
	for _, o := range n.Objects() {
		o.Next = nil
	}
}
func Abc_NtkCleanNext_rec(n *Network) { hierarchyWalk(n, Abc_NtkCleanNext) }
func Abc_NtkCleanMarkA(n *Network) {
	for _, o := range n.Objects() {
		o.MarkA = false
	}
}
func Abc_NtkCleanMarkB(n *Network) {
	for _, o := range n.Objects() {
		o.MarkB = false
	}
}
func Abc_NtkCleanMarkC(n *Network) {
	for _, o := range n.Objects() {
		o.MarkC = false
	}
}
func Abc_NtkCleanMarkAB(n *Network) {
	for _, o := range n.Objects() {
		o.MarkA, o.MarkB = false, false
	}
}
func Abc_NtkCleanMarkABC(n *Network) {
	for _, o := range n.Objects() {
		o.MarkA, o.MarkB, o.MarkC = false, false, false
	}
}
func Abc_NodeFindFanin(node, fanin *Object) int32 {
	for i, edge := range node.fanins {
		if edge.Object == fanin {
			return int32(i)
		}
	}
	return -1
}
func Abc_NodeFindCoFanout(node *Object) *Object {
	for _, fanout := range node.fanouts {
		if fanout.IsCO() {
			return fanout
		}
	}
	return nil
}
func Abc_NodeFindNonCoFanout(node *Object) *Object {
	for _, fanout := range node.fanouts {
		if !fanout.IsCO() {
			return fanout
		}
	}
	return nil
}
func Abc_NodeHasUniqueCoFanout(node *Object) *Object {
	var found *Object
	for _, output := range node.fanouts {
		if !output.IsCO() || output.fanins[0].Complement {
			continue
		}
		if found == nil {
			found = output
		} else if Abc_ObjName(found) != Abc_ObjName(output) {
			return nil
		}
	}
	return found
}

// Abc_NtkTransferCopy composes two successive network-copy mappings.
func Abc_NtkTransferCopy(n *Network) {
	for _, node := range n.Objects() {
		if node.Kind == ObjectNet {
			continue
		}
		first := getCopy(node)
		if first.Object == nil {
			setCopy(node, Edge{})
			continue
		}
		second := getCopy(first.Object)
		if second.Object != nil {
			second.Complement = second.Complement != first.Complement
		}
		setCopy(node, second)
	}
}

func Abc_ObjCrossCutInc(node *Object) int32 {
	node.crossCutVisits++
	return abcBool(node.crossCutVisits == len(node.fanouts))
}

func Abc_NtkCrossCut_rec(node *Object, size, maximum *int32) int32 {
	if node.IsCI() {
		return 0
	}
	if Abc_NodeIsTravIdCurrent(node) != 0 {
		return Abc_ObjCrossCutInc(node)
	}
	Abc_NodeSetTravIdCurrent(node)
	var decrement int32
	for _, edge := range node.fanins {
		decrement += Abc_NtkCrossCut_rec(edge.Object, size, maximum)
	}
	*size++
	if *maximum < *size {
		*maximum = *size
	}
	*size -= decrement
	return Abc_ObjCrossCutInc(node)
}

func Abc_NtkCrossCut(n *Network) int32 {
	for _, node := range n.Objects() {
		node.crossCutVisits = 0
	}
	Abc_NtkIncrementTravId(n)
	var size, maximum int32
	for _, output := range n.cos {
		Abc_NtkCrossCut_rec(output, &size, &maximum)
		size--
	}
	if size != 0 {
		panic("abc: unbalanced cross-cut traversal")
	}
	ratio := 0.0
	if len(n.Objects()) != 0 {
		ratio = 100 * float64(maximum) / float64(len(n.Objects()))
	}
	fmt.Printf("Max cross cut size = %6d.  Ratio = %6.2f %%\n", maximum, ratio)
	return maximum
}
