package nwk

import "github.com/glycerine/abcgo/misc/vec"

func Nwk_ObjCollectFanins(o *Nwk_Obj_t, v *Vec_Ptr_t) {
	vec.Vec_PtrClear(v)
	for _, n := range o.Fanins {
		vec.Vec_PtrPush(v, n)
	}
}
func Nwk_ObjCollectFanouts(o *Nwk_Obj_t, v *Vec_Ptr_t) {
	vec.Vec_PtrClear(v)
	for _, n := range o.Fanouts {
		vec.Vec_PtrPush(v, n)
	}
}
func findNode(a []*Nwk_Obj_t, n *Nwk_Obj_t) int {
	for i, v := range a {
		if v == n {
			return i
		}
	}
	return -1
}
func Nwk_ObjFindFanin(o, n *Nwk_Obj_t) int32  { return int32(findNode(o.Fanins, n)) }
func Nwk_ObjFindFanout(o, n *Nwk_Obj_t) int32 { return int32(findNode(o.Fanouts, n)) }
func Nwk_ObjReallocIsNeeded(o *Nwk_Obj_t) int32 {
	return bit(len(o.Fanins) == cap(o.Fanins) || len(o.Fanouts) == cap(o.Fanouts))
}
func Nwk_ManReallocNode(o *Nwk_Obj_t) *Nwk_Obj_t {
	a := make([]*Nwk_Obj_t, len(o.Fanins), 2*cap(o.Fanins)+2)
	copy(a, o.Fanins)
	o.Fanins = a
	a = make([]*Nwk_Obj_t, len(o.Fanouts), 2*cap(o.Fanouts)+2)
	copy(a, o.Fanouts)
	o.Fanouts = a
	return nil
}
func Nwk_ObjAddFanin(o, n *Nwk_Obj_t) {
	if o.Manager != n.Manager || o.ID < 0 || n.ID < 0 {
		panic("nwk: foreign/deleted fanin")
	}
	o.Fanins = append(o.Fanins, n)
	n.Fanouts = append(n.Fanouts, o)
	o.Level = max(o.Level, n.Level+Nwk_ObjIsNode(o))
}
func removeNode(a []*Nwk_Obj_t, i int) []*Nwk_Obj_t {
	if i < 0 {
		panic("nwk: absent adjacency")
	}
	copy(a[i:], a[i+1:])
	a[len(a)-1] = nil
	return a[:len(a)-1]
}
func Nwk_ObjDeleteFanin(o, n *Nwk_Obj_t) {
	i, j := findNode(o.Fanins, n), findNode(n.Fanouts, o)
	if i < 0 || j < 0 {
		panic("nwk: asymmetric adjacency")
	}
	o.Fanins = removeNode(o.Fanins, i)
	n.Fanouts = removeNode(n.Fanouts, j)
}
func Nwk_ObjPatchFanin(o, old, new *Nwk_Obj_t) {
	if o == new || old == new || o.Manager != new.Manager || old.Manager != o.Manager {
		panic("nwk: invalid fanin patch")
	}
	i, j := findNode(o.Fanins, old), findNode(old.Fanouts, o)
	if i < 0 || j < 0 {
		panic("nwk: absent old fanin")
	}
	o.Fanins[i] = new
	old.Fanouts = removeNode(old.Fanouts, j)
	new.Fanouts = append(new.Fanouts, o)
}
func Nwk_ObjTransferFanout(from, to *Nwk_Obj_t) {
	if from == to || from.Manager != to.Manager {
		panic("nwk: invalid fanout transfer")
	}
	fanouts := append([]*Nwk_Obj_t(nil), from.Fanouts...)
	for _, o := range fanouts {
		Nwk_ObjPatchFanin(o, from, to)
	}
}
func Nwk_ObjReplace(old, new *Nwk_Obj_t) { Nwk_ObjTransferFanout(old, new); Nwk_ManDeleteNode_rec(old) }
