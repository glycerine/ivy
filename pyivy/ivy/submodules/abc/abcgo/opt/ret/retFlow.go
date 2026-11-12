package ret

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"time"
)

// A real Go sentinel replaces C's tagged pointer (Abc_Obj_t*)1. It is never
// inserted into a graph or dereferenced as a graph object.
var retPathEnd = &abc.Object{}

func Abc_ObjSetPath(o, next *abc.Object) int32 { o.Copy = next; return 1 }
func Abc_ObjGetPath(o *abc.Object) *abc.Object { return o.Copy }
func retNeighbors(o *abc.Object, forward bool) []*abc.Object {
	if forward {
		return o.Fanouts()
	}
	edges := o.Fanins()
	out := make([]*abc.Object, len(edges))
	for i, e := range edges {
		out[i] = e.Object
	}
	return out
}
func retPathPredecessor(o *abc.Object, forward bool, both bool) *abc.Object {
	for _, next := range retNeighbors(o, !forward) {
		if next.Copy == o {
			return next
		}
	}
	if both {
		for _, next := range retNeighbors(o, forward) {
			if next.Copy == o {
				return next
			}
		}
	}
	return nil
}
func Abc_ObjGetFanoutPath(o *abc.Object) *abc.Object     { return retPathPredecessor(o, false, false) }
func Abc_ObjGetFaninPath(o *abc.Object) *abc.Object      { return retPathPredecessor(o, true, false) }
func Abc_ObjGetPredecessorBwd(o *abc.Object) *abc.Object { return retPathPredecessor(o, false, true) }
func Abc_ObjGetPredecessorFwd(o *abc.Object) *abc.Object { return retPathPredecessor(o, true, true) }

// Directed augmenting paths reroute the predecessor's occupied vertex through
// another successor, or cancel its old outgoing flow when it can be rerouted.
func retFlowPath(o *abc.Object, forward, directed bool) int32 {
	if abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return 0
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	var predecessor *abc.Object
	if !directed {
		predecessor = retPathPredecessor(o, forward, true)
	}
	neighbors := func(node *abc.Object) []*abc.Object {
		v := retNeighbors(node, forward)
		if !directed {
			v = append(v, retNeighbors(node, !forward)...)
		}
		return v
	}
	if o.Copy == nil {
		if o.MarkA {
			return Abc_ObjSetPath(o, retPathEnd)
		}
		for _, next := range neighbors(o) {
			if directed || next != predecessor && next.Kind != abc.ObjectLatch {
				if retFlowPath(next, forward, directed) != 0 {
					return Abc_ObjSetPath(o, next)
				}
			}
		}
		return 0
	}
	if directed {
		predecessor = retPathPredecessor(o, forward, false)
	}
	if predecessor == nil {
		return 0
	}
	for _, next := range neighbors(predecessor) {
		if directed || next.Kind != abc.ObjectLatch {
			if retFlowPath(next, forward, directed) != 0 {
				return Abc_ObjSetPath(predecessor, next)
			}
		}
	}
	if retFlowPath(predecessor, forward, directed) != 0 {
		return Abc_ObjSetPath(predecessor, nil)
	}
	return 0
}
func Abc_NtkMaxFlowBwdPath_rec(o *abc.Object) int32  { return retFlowPath(o, false, false) }
func Abc_NtkMaxFlowFwdPath_rec(o *abc.Object) int32  { return retFlowPath(o, true, false) }
func Abc_NtkMaxFlowBwdPath2_rec(o *abc.Object) int32 { return retFlowPath(o, false, true) }
func Abc_NtkMaxFlowFwdPath2_rec(o *abc.Object) int32 { return retFlowPath(o, true, true) }
func Abc_NtkMaxFlowFwdPath3_rec(o, previous *abc.Object, fanin int32) int32 {
	if abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return 0
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	if fanin != 0 && previous.Copy != nil {
		return 0
	}
	if o.Copy == nil {
		if o.MarkA {
			return Abc_ObjSetPath(o, retPathEnd)
		}
		for _, next := range o.Fanouts() {
			if Abc_NtkMaxFlowFwdPath3_rec(next, o, 1) != 0 {
				if fanin != 0 {
					return Abc_ObjSetPath(previous, o)
				}
				return 1
			}
		}
	}
	for _, e := range o.Fanins() {
		if e.Object.Kind != abc.ObjectLatch && Abc_NtkMaxFlowFwdPath3_rec(e.Object, o, 0) != 0 {
			return Abc_ObjSetPath(e.Object, nil)
		}
	}
	return 0
}
func Abc_NtkMaxFlow(n *abc.Network, forward, verbose int32) []*abc.Object {
	start := time.Now()
	for _, o := range n.Objects() {
		o.Copy = nil
	}
	flow := int32(0)
	abc.Abc_NtkIncrementTravId(n)
	augment := func(latch *abc.Object) int32 {
		if forward != 0 {
			return Abc_NtkMaxFlowFwdPath2_rec(latch.Fanouts()[0])
		}
		return Abc_NtkMaxFlowBwdPath2_rec(retFan0(latch))
	}
	for _, latch := range retLatches(n) {
		if augment(latch) != 0 {
			flow++
			abc.Abc_NtkIncrementTravId(n)
		}
	}
	abc.Abc_NtkIncrementTravId(n)
	for _, latch := range retLatches(n) {
		if augment(latch) != 0 {
			panic("ret: residual augmenting path after max-flow")
		}
	}
	cut := Abc_NtkMaxFlowMinCut(n, forward)
	if Abc_NtkMaxFlowVerifyCut(n, cut, forward) == 0 {
		panic("ret: computed minimum cut is not a cut")
	}
	cut = Abc_NtkMaxFlowMinCutUpdate(n, cut, forward)
	if verbose != 0 {
		fmt.Printf("L = %6d. Forward = %t max-flow = %6d. Min-cut = %6d. Time = %.2f sec\n", len(retLatches(n)), forward != 0, flow, len(cut), time.Since(start).Seconds())
	}
	return cut
}
func Abc_NtkMaxFlowMinCut(n *abc.Network, forward int32) []*abc.Object {
	var cut []*abc.Object
	for _, o := range n.Objects() {
		if o.Copy != nil && abc.Abc_NodeIsTravIdCurrent(o) != 0 && (o.MarkA || o.Copy == retPathEnd || abc.Abc_NodeIsTravIdCurrent(o.Copy) == 0) {
			cut = append(cut, o)
		}
	}
	return cut
}
func Abc_NtkMaxFlowMarkCut_rec(o *abc.Object) {
	if o.MarkA {
		return
	}
	o.MarkA = true
	for _, e := range o.Fanins() {
		Abc_NtkMaxFlowMarkCut_rec(e.Object)
	}
}
func Abc_NtkMaxFlowCollectCut_rec(o *abc.Object, cut *[]*abc.Object) {
	if abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	if o.MarkA {
		*cut = append(*cut, o)
		return
	}
	for _, e := range o.Fanins() {
		Abc_NtkMaxFlowCollectCut_rec(e.Object, cut)
	}
}
func Abc_NtkMaxFlowMinCutUpdate(n *abc.Network, cut []*abc.Object, forward int32) []*abc.Object {
	for _, o := range n.Objects() {
		o.MarkA = false
	}
	for _, latch := range retLatches(n) {
		latch.Fanouts()[0].MarkA = true
	}
	for _, o := range cut {
		Abc_NtkMaxFlowMarkCut_rec(o)
	}
	cut = cut[:0]
	if forward != 0 {
		for _, o := range n.Objects() {
			if !o.MarkA {
				continue
			}
			for _, next := range o.Fanouts() {
				if !next.MarkA {
					cut = append(cut, o)
					break
				}
			}
		}
	} else {
		abc.Abc_NtkIncrementTravId(n)
		for _, latch := range retLatches(n) {
			Abc_NtkMaxFlowCollectCut_rec(retFan0(latch), &cut)
		}
		for _, o := range n.Objects() {
			o.MarkA = abc.Abc_NodeIsTravIdCurrent(o) != 0
		}
		for _, o := range cut {
			o.MarkA = false
		}
	}
	return cut
}
func Abc_NtkMaxFlowVerifyCut_rec(o *abc.Object, forward int32) int32 {
	if abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return 1
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	if forward != 0 && o.IsCO() || forward == 0 && o.IsCI() {
		return 0
	}
	for _, next := range retNeighbors(o, forward != 0) {
		if Abc_NtkMaxFlowVerifyCut_rec(next, forward) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkMaxFlowVerifyCut(n *abc.Network, cut []*abc.Object, forward int32) int32 {
	abc.Abc_NtkIncrementTravId(n)
	for _, o := range cut {
		abc.Abc_NodeSetTravIdCurrent(o)
	}
	for _, latch := range retLatches(n) {
		start := retFan0(latch)
		if forward != 0 {
			start = latch.Fanouts()[0]
		}
		if Abc_NtkMaxFlowVerifyCut_rec(start, forward) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkMaxFlowTest(n *abc.Network) {
	for _, o := range n.ObjectsOfKind(abc.ObjectPO) {
		o.MarkA = true
	}
	for _, o := range retLatches(n) {
		o.MarkA = true
		retFan0(o).MarkA = true
	}
	Abc_NtkMaxFlow(n, 1, 1)
	for _, o := range n.Objects() {
		o.MarkA = false
	}
	for _, o := range n.ObjectsOfKind(abc.ObjectPI) {
		o.MarkA = true
	}
	for _, o := range retLatches(n) {
		o.MarkA = true
		o.Fanouts()[0].MarkA = true
	}
	Abc_NtkMaxFlow(n, 0, 1)
	for _, o := range n.Objects() {
		o.MarkA = false
	}
}
func Abc_NtkMaxFlowPrintFlow(n *abc.Network, forward int32) {
	for _, latch := range retLatches(n) {
		o := retFan0(latch)
		if forward != 0 {
			o = latch.Fanouts()[0]
		}
		if o.Copy == nil {
			continue
		}
		fmt.Print("Path = ")
		for o != retPathEnd && o != nil {
			fmt.Printf("%s(%d) ", o.Name(), o.ID)
			o = o.Copy
		}
		fmt.Println()
	}
}
func Abc_NtkMaxFlowPrintCut(n *abc.Network, cut []*abc.Object) {
	fmt.Printf("Cut = %d: ", len(cut))
	for _, o := range cut {
		fmt.Printf("%s(%d) ", o.Name(), o.ID)
	}
	fmt.Println()
}
