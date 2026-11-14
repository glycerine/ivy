package abc

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/nm"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_VecObjPushUniqueOrderByLevel maintains descending level order.
func Abc_VecObjPushUniqueOrderByLevel(v *vec.Vec_Ptr_t, node *Object) {
	if node == nil {
		panic("abc: nil ordered node")
	}
	if vec.Vec_PtrPushUnique(v, node) != 0 {
		return
	}
	entries := vec.Vec_PtrArray(v)
	for i := len(entries) - 1; i > 0; i-- {
		previous, ok := entries[i-1].(*Object)
		if !ok || previous == nil {
			panic("abc: invalid ordered node")
		}
		if node.Level <= previous.Level {
			break
		}
		entries[i], entries[i-1] = entries[i-1], entries[i]
	}
}

func Abc_NodeCompareLevelsIncrease(a, b **Object) int32 {
	if (*a).Level < (*b).Level || (*a).Level == (*b).Level && (*a).ID < (*b).ID {
		return -1
	}
	if (*a).Level > (*b).Level || (*a).Level == (*b).Level && (*a).ID > (*b).ID {
		return 1
	}
	return 0
}

func Abc_NodeCompareLevelsDecrease(a, b **Object) int32 {
	return -Abc_NodeCompareLevelsIncrease(a, b)
}

func Abc_NtkFanoutCounts(n *Network) *vec.Vec_Int_t {
	counts := vec.Vec_IntAlloc(len(n.objects))
	vec.Vec_IntFill(counts, len(n.objects), -1)
	refs := n.refs
	for _, o := range n.Objects() {
		if o.IsCI() || o.Kind == ObjectNode {
			count := len(o.fanouts)
			if refs != nil {
				count = refs[o]
			}
			vec.Vec_IntWriteEntry(counts, o.ID, int32(count))
		}
	}
	return counts
}

func Abc_NtkCollectObjects(n *Network) *vec.Vec_Ptr_t {
	objects := vec.Vec_PtrAlloc(len(n.objects))
	for _, o := range n.Objects() {
		vec.Vec_PtrPush(objects, o)
	}
	return objects
}

func Abc_NtkGetCiIds(n *Network) *vec.Vec_Int_t {
	ids := vec.Vec_IntAlloc(len(n.cis))
	for _, input := range n.cis {
		vec.Vec_IntPush(ids, int32(input.ID))
	}
	return ids
}

// Abc_NtkReassignIds compacts an AIG into terminal and DFS order.
func Abc_NtkReassignIds(n *Network) {
	if n == nil || Abc_NtkIsStrash(n) == 0 {
		panic("abc: ID reassignment requires an AIG")
	}
	oldObjects := n.objects
	var ordered []*Object
	seen := make(map[*Object]bool)
	add := func(o *Object) {
		if o == nil || !n.contains(o) || seen[o] {
			panic("abc: invalid or repeated object during ID reassignment")
		}
		seen[o] = true
		ordered = append(ordered, o)
	}
	add(n.constant)
	for _, kind := range []ObjectKind{ObjectPI, ObjectPO} {
		for _, o := range n.ObjectsOfKind(kind) {
			add(o)
		}
	}
	for _, box := range n.boxes {
		add(box)
		for _, edge := range box.fanins {
			add(edge.Object)
		}
		for _, terminal := range box.fanouts {
			add(terminal)
		}
	}
	for _, node := range Abc_AigDfs(n, 1, 0) {
		if node != n.constant {
			add(node)
		}
	}
	if len(ordered) != len(n.Objects()) {
		panic("abc: ID reassignment did not cover all objects")
	}
	for _, attr := range n.Attributes {
		if attr == nil {
			continue
		}
		for id, value := range vec.Vec_AttArray(attr) {
			if value != nil && (id >= len(oldObjects) || oldObjects[id] == nil) {
				panic("abc: attribute exists for a missing object")
			}
		}
	}
	var mapping *scl.NetworkMapping
	if n.timingMapping != nil {
		var ok bool
		mapping, ok = n.timingMapping.(*scl.NetworkMapping)
		if !ok {
			panic("abc: unknown ID-indexed timing metadata")
		}
	}
	names := nm.Nm_ManCreate(int32(len(ordered)))
	var nameIDs []int32
	if n.NameIDs != nil {
		nameIDs = make([]int32, len(ordered))
	}
	var gateIDs []int32
	var mioGates []*mio.Mio_Gate_t
	if mapping != nil && mapping.GateIDs != nil {
		gateIDs = make([]int32, len(ordered))
		for i := range gateIDs {
			gateIDs[i] = -1
		}
	}
	if mapping != nil && mapping.MioGates != nil {
		mioGates = make([]*mio.Mio_Gate_t, len(ordered))
	}
	attrValues := make([][]any, len(n.Attributes))
	for i, attr := range n.Attributes {
		if attr != nil {
			attrValues[i] = append([]any(nil), vec.Vec_AttArray(attr)...)
			clear(vec.Vec_AttArray(attr))
		}
	}
	kindOrder := make(map[ObjectKind][]*Object)
	for id, o := range ordered {
		oldID := o.ID
		if name, ok := n.names.Lookup(int32(oldID)); ok {
			nm.Nm_ManStoreIdName(names, int32(id), int32(o.Kind), name, "")
		}
		if nameIDs != nil && oldID < len(n.NameIDs) {
			nameIDs[id] = n.NameIDs[oldID]
		}
		if gateIDs != nil && oldID < len(mapping.GateIDs) {
			gateIDs[id] = mapping.GateIDs[oldID]
		}
		if mioGates != nil && oldID < len(mapping.MioGates) {
			mioGates[id] = mapping.MioGates[oldID]
		}
		for i, attr := range n.Attributes {
			if attr != nil && oldID < len(attrValues[i]) {
				vec.Vec_AttWriteEntry(attr, id, attrValues[i][oldID])
			}
		}
		o.ID = id
		kindOrder[o.Kind] = append(kindOrder[o.Kind], o)
	}
	for _, attr := range n.Attributes {
		if attr != nil {
			values := vec.Vec_AttArray(attr)
			if len(values) > len(ordered) {
				clear(values[len(ordered):])
			}
		}
	}
	n.objects, n.names, n.NameIDs, n.kindOrder = ordered, names, nameIDs, kindOrder
	n.Topo = nil
	if mapping != nil {
		mapping.GateIDs, mapping.MioGates = gateIDs, mioGates
	}
	manager, ok := n.FunctionManager.(*Abc_Aig_t)
	if !ok || manager == nil {
		manager = Abc_AigAlloc(n)
	}
	Abc_AigRehash(manager)
}
