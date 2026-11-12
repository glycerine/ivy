// Package ret implements logic-network retiming with explicit latch state.
package ret

import "github.com/glycerine/abcgo/base/abc"

type Abc_Ntk_t = abc.Network
type Abc_Obj_t = abc.Object

func retBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func retLatches(n *abc.Network) []*abc.Object { return n.ObjectsOfKind(abc.ObjectLatch) }
func retNodes(n *abc.Network) []*abc.Object   { return n.ObjectsOfKind(abc.ObjectNode) }
func retFan0(o *abc.Object) *abc.Object       { return abc.Abc_ObjFanin0(o) }
func retAdd(o *abc.Object, e abc.Edge) {
	if err := o.Network().AddFanin(o, e); err != nil {
		panic(err)
	}
}
func retPatch(o *abc.Object, index int, e abc.Edge) {
	if err := o.Network().PatchFanin(o, index, e); err != nil {
		panic(err)
	}
}
func retDelete(o *abc.Object) { abc.Abc_NtkDeleteObj(o) }

// Prepared retiming graphs connect latches directly, so a sequential cycle can
// appear combinational to the general network transfer validator. Patch only
// these incident edges; latch boundaries are restored before network checking.
func retTransfer(from, to *abc.Object) {
	if from == to {
		return
	}
	if from.Network() != to.Network() {
		panic("ret: cross-network transfer")
	}
	for _, fanout := range from.Fanouts() {
		for i, e := range fanout.Fanins() {
			if e.Object == from {
				replacement := to.Signal()
				replacement.Complement = e.Complement
				retPatch(fanout, i, replacement)
			}
		}
	}
}
func retSopNode(n *abc.Network, source *abc.Object) *abc.Object {
	return abc.Abc_NtkDupObj(n, source, 0)
}
func retLatchValue(o *abc.Object) int32 { return abc.Abc_LatchIsInit1(o) }
