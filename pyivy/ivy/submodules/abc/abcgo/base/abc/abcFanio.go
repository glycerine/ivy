package abc

// Fanin APIs accept an explicit Edge wherever C accepted a complemented pointer.
func Abc_ObjAddFanin(o *Object, fanin any) {
	if err := o.network.AddFanin(o, objectEdge(fanin)); err != nil {
		panic(err)
	}
}
func Abc_ObjDeleteFanin(o, fanin *Object) {
	for i, e := range o.fanins {
		if e.Object == fanin {
			if err := o.network.DeleteFanin(o, i); err != nil {
				panic(err)
			}
			return
		}
	}
}
func Abc_ObjRemoveFanins(o *Object) {
	for len(o.fanins) > 0 {
		if err := o.network.DeleteFanin(o, len(o.fanins)-1); err != nil {
			panic(err)
		}
	}
}
func Abc_ObjPatchFanin(o, old *Object, replacement any) {
	e := objectEdge(replacement)
	for i, fanin := range o.fanins {
		if fanin.Object == old {
			e.Complement = e.Complement != fanin.Complement
			if err := o.network.PatchFanin(o, i, e); err != nil {
				panic(err)
			}
			return
		}
	}
}
func Abc_ObjPatchFanoutFanin(o *Object, newID int32) {
	replacement := o.network.Object(int(newID))
	if replacement == nil {
		panic("unknown replacement object")
	}
	for _, fanout := range o.Fanouts() {
		Abc_ObjPatchFanin(fanout, o, replacement)
	}
}
func Abc_ObjFanoutFaninNum(fanout, fanin *Object) int32 {
	for i, e := range fanout.fanins {
		if e.Object == fanin {
			return int32(i)
		}
	}
	return -1
}
func Abc_ObjInsertBetween(in, out *Object, kind ObjectKind) *Object {
	index := Abc_ObjFanoutFaninNum(out, in)
	if index < 0 {
		return nil
	}
	if in.network != out.network {
		panic("foreign insertion")
	}
	if kind == ObjectPI || kind == ObjectConst1 || kind <= ObjectNone || kind > ObjectBlackbox {
		panic("invalid insertion kind")
	}
	n := in.network
	middle := n.NewObject(kind)
	if err := n.AddFanin(middle, in.Signal()); err != nil {
		panic(err)
	}
	old := out.fanins[index]
	old.Object = middle
	if err := n.PatchFanin(out, int(index), old); err != nil {
		panic(err)
	}
	return middle
}
func Abc_ObjTransferFanout(from, to *Object) {
	if from == to || from.network != to.network {
		panic("invalid fanout transfer")
	}
	cone, err := from.network.Topological([]*Object{to})
	if err != nil {
		panic(err)
	}
	for _, o := range cone {
		if o == from {
			panic("fanout transfer would introduce cycle")
		}
	}
	if to.Kind == ObjectPO {
		panic("PO cannot drive fanouts")
	}
	for _, fanout := range from.Fanouts() {
		Abc_ObjPatchFanin(fanout, from, to)
	}
}
func deleteDangling(o *Object, nodesOnly bool) {
	if len(o.fanouts) != 0 {
		panic("deletion requires no fanouts")
	}
	n := o.network
	fanins := o.Fanins()
	if err := n.DeleteObject(o); err != nil {
		panic(err)
	}
	for _, e := range fanins {
		source := e.Object
		if !n.contains(source) || len(source.fanouts) != 0 || source == n.constant {
			continue
		}
		if (nodesOnly && source.Kind == ObjectNode) || (!nodesOnly && source.Kind != ObjectPI) {
			deleteDangling(source, nodesOnly)
		}
	}
}
func Abc_ObjReplace(old, replacement *Object) {
	Abc_ObjTransferFanout(old, replacement)
	deleteDangling(old, true)
}
func Abc_ObjReplaceByConstant(o *Object, one int32) {
	e := o.network.One()
	if one == 0 {
		e = e.Not()
	}
	for _, fanout := range o.Fanouts() {
		Abc_ObjPatchFanin(fanout, o, e)
	}
	if o.Kind == ObjectNode {
		deleteDangling(o, true)
	}
}
