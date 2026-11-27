package abc

import "fmt"

// Abc_ObjAdjustFanoutCount applies the temporary count used by cone
// traversal. Keep it on the node so changing edges still affects the count.
func Abc_ObjAdjustFanoutCount(o *Object, delta int32) {
	if o == nil {
		panic("cannot adjust fanouts of a nil object")
	}
	offset := o.fanoutAdjustment + int(delta)
	if len(o.fanouts)+offset < 0 {
		panic("fanout adjustment made a count negative")
	}
	o.fanoutAdjustment = offset
}

// Abc_NtkFinishFanoutCountAdjustments is retained for callers that delimit
// temporary MFFC adjustments. Each balanced pair now restores its node directly.
func Abc_NtkFinishFanoutCountAdjustments(n *Network) {}
func Abc_NtkIncrementTravId(n *Network) {
	n.travID++
	if n.travID == 0 {
		for _, o := range n.Objects() {
			o.travID = 0
		}
		n.travID = 1
	}
}
func Abc_NodeIsTravIdCurrent(o *Object) int32 { return abcBool(o.travID == o.network.travID) }
func Abc_NodeSetTravIdCurrent(o *Object)      { o.travID = o.network.travID }
func refDeref(o *Object, reference, label, stop bool) int32 {
	if label {
		Abc_NodeSetTravIdCurrent(o)
	}
	if o.IsCI() || o.Kind == ObjectConst1 {
		return 0
	}
	count := int32(1)
	for _, e := range o.fanins {
		if stop && e.Complement {
			continue
		}
		child := e.Object
		if reference {
			before := Abc_ObjFanoutNum(child)
			Abc_ObjAdjustFanoutCount(child, 1)
			if before == 0 {
				count += refDeref(child, true, label, stop)
			}
		} else {
			Abc_ObjAdjustFanoutCount(child, -1)
			if Abc_ObjFanoutNum(child) == 0 {
				count += refDeref(child, false, label, stop)
			}
		}
	}
	return count
}
func Abc_NodeRefDeref(o *Object, reference, label int32) int32 {
	return refDeref(o, reference != 0, label != 0, false)
}
func Abc_NodeRefDerefStop(o *Object, reference int32) int32 {
	return refDeref(o, reference != 0, false, true)
}
func Abc_NodeDeref_rec(o *Object) int32 { return Abc_NodeRefDeref(o, 0, 0) }
func Abc_NodeRef_rec(o *Object) int32   { return Abc_NodeRefDeref(o, 1, 0) }
func mffcSize(o *Object, stop, label bool) int32 {
	if o.Kind != ObjectNode {
		panic("MFFC root must be a node")
	}
	if len(o.fanins) == 0 {
		return 0
	}
	before := refDeref(o, false, label, stop)
	after := refDeref(o, true, false, stop)
	if before != after {
		panic("unbalanced MFFC references")
	}
	return before
}
func Abc_NodeMffcSize(o *Object) int32     { return mffcSize(o, false, false) }
func Abc_NodeMffcSizeStop(o *Object) int32 { return mffcSize(o, true, false) }
func Abc_NodeMffcLabelAig(o *Object) int32 { return mffcSize(o, false, true) }
func mffcCollect(o *Object, cone, support *[]*Object, topmost, labelBoundary bool) {
	refs := Abc_ObjFanoutNum(o)
	boundary := !topmost && (o.IsCI() || o.Kind == ObjectConst1 || refs > 0)
	if boundary && !labelBoundary {
		return
	}
	if Abc_NodeIsTravIdCurrent(o) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(o)
	if boundary {
		if support != nil {
			*support = append(*support, o)
		}
		return
	}
	for _, e := range o.fanins {
		mffcCollect(e.Object, cone, support, false, labelBoundary)
	}
	if cone != nil {
		*cone = append(*cone, o)
	}
}

// Output vectors are pointers to slices so nil retains the C optional-output
// convention without requiring a container package.
func Abc_NodeMffcConeSupp_rec(o *Object, cone, support *[]*Object, topmost int32) {
	mffcCollect(o, cone, support, topmost != 0, true)
}
func Abc_NodeMffcConeSupp(o *Object, cone, support *[]*Object) {
	if o.Kind != ObjectNode {
		panic("MFFC root must be node")
	}
	if cone != nil {
		*cone = (*cone)[:0]
	}
	if support != nil {
		*support = (*support)[:0]
	}
	Abc_NtkIncrementTravId(o.network)
	mffcCollect(o, cone, support, true, true)
}
func Abc_NodeMffcConeSuppPrint(o *Object) {
	var cone, support []*Object
	Abc_NodeDeref_rec(o)
	Abc_NodeMffcConeSupp(o, &cone, &support)
	Abc_NodeRef_rec(o)
	fmt.Printf("Node = %6s : Supp = %3d  Cone = %3d  (", Abc_ObjName(o), len(support), len(cone))
	for _, inside := range cone {
		fmt.Printf(" %s", Abc_ObjName(inside))
	}
	fmt.Println(" )")
}
func Abc_NodeMffcInside(o *Object, leaves []*Object, inside *[]*Object) int32 {
	n := o.network
	for _, leaf := range leaves {
		if !n.contains(leaf) {
			panic("foreign MFFC leaf")
		}
	}
	for _, leaf := range leaves {
		Abc_ObjAdjustFanoutCount(leaf, 1)
	}
	before := Abc_NodeDeref_rec(o)
	Abc_NodeMffcConeSupp(o, inside, nil)
	after := Abc_NodeRef_rec(o)
	for _, leaf := range leaves {
		Abc_ObjAdjustFanoutCount(leaf, -1)
	}
	if before != after {
		panic("unbalanced MFFC references")
	}
	return before
}
func Abc_NodeMffcInsideCollect(o *Object) []*Object {
	var inside []*Object
	Abc_NodeMffcInside(o, nil, &inside)
	return inside
}
func Abc_NodeMffcLabel_rec(o *Object, topmost int32, nodes *[]*Object) {
	mffcCollect(o, nodes, nil, topmost != 0, false)
}
func Abc_NodeMffcLabel(o *Object, nodes *[]*Object) int32 {
	before := Abc_NodeDeref_rec(o)
	Abc_NtkIncrementTravId(o.network)
	Abc_NodeMffcLabel_rec(o, 1, nodes)
	after := Abc_NodeRef_rec(o)
	if before != after {
		panic("unbalanced MFFC references")
	}
	return before
}

// Abc_NodeMffcLabelWithLeaves labels the MFFC with each cut leaf held by one
// extra reference, matching ABC's temporary fanout-count increments.
func Abc_NodeMffcLabelWithLeaves(o *Object, leaves []*Object, nodes *[]*Object) int32 {
	n := o.network
	for _, leaf := range leaves {
		if !n.contains(leaf) {
			panic("foreign MFFC leaf")
		}
	}
	for _, leaf := range leaves {
		Abc_ObjAdjustFanoutCount(leaf, 1)
	}
	result := Abc_NodeMffcLabel(o, nodes)
	for _, leaf := range leaves {
		Abc_ObjAdjustFanoutCount(leaf, -1)
	}
	return result
}

func Abc_NodeIsTravIdPrevious(o *Object) int32 { return abcBool(o.travID == o.network.travID-1) }
func Abc_NodeSetTravIdPrevious(o *Object)      { o.travID = o.network.travID - 1 }
