package fraig

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/msat"
	"io"
	"math/bits"
	"os"
)

func Fraig_Dfs(p *Fraig_Man_t, equiv int32) *Fraig_NodeVec_t {
	p.nTravIds++
	v := Fraig_NodeVecAlloc(100)
	for _, n := range p.vOutputs.pArray {
		Fraig_Dfs_rec(p, Fraig_Regular(n), v, equiv)
	}
	return v
}
func Fraig_DfsOne(p *Fraig_Man_t, n *Fraig_Node_t, equiv int32) *Fraig_NodeVec_t {
	p.nTravIds++
	v := Fraig_NodeVecAlloc(100)
	Fraig_Dfs_rec(p, Fraig_Regular(n), v, equiv)
	return v
}
func Fraig_DfsNodes(p *Fraig_Man_t, nodes []*Fraig_Node_t, count, equiv int32) *Fraig_NodeVec_t {
	p.nTravIds++
	v := Fraig_NodeVecAlloc(100)
	for _, n := range nodes[:count] {
		Fraig_Dfs_rec(p, Fraig_Regular(n), v, equiv)
	}
	return v
}
func Fraig_Dfs_rec(p *Fraig_Man_t, n *Fraig_Node_t, v *Fraig_NodeVec_t, equiv int32) {
	assert(!Fraig_IsComplement(n), "DFS node is regular")
	if n.TravId == p.nTravIds {
		return
	}
	n.TravId = p.nTravIds
	if Fraig_NodeIsAnd(n) != 0 {
		Fraig_Dfs_rec(p, Fraig_Regular(n.p1), v, equiv)
		Fraig_Dfs_rec(p, Fraig_Regular(n.p2), v, equiv)
	}
	if equiv != 0 && n.pNextE != nil {
		Fraig_Dfs_rec(p, n.pNextE, v, equiv)
	}
	Fraig_NodeVecPush(v, n)
}
func Fraig_CountNodes(p *Fraig_Man_t, equiv int32) int32 {
	return Fraig_NodeVecReadSize(Fraig_Dfs(p, equiv))
}
func Fraig_CheckTfi(p *Fraig_Man_t, old, node *Fraig_Node_t) int32 {
	assert(!Fraig_IsComplement(old) && !Fraig_IsComplement(node), "TFI nodes are regular")
	p.nTravIds++
	return Fraig_CheckTfi_rec(p, node, old)
}
func Fraig_CheckTfi_rec(p *Fraig_Man_t, n, old *Fraig_Node_t) int32 {
	if n == nil || n.Num < old.Num && p.fChoicing == 0 {
		return 0
	}
	if n == old {
		return 1
	}
	if n.TravId == p.nTravIds {
		return 0
	}
	n.TravId = p.nTravIds
	if Fraig_CheckTfi_rec(p, Fraig_Regular(n.p1), old) != 0 || Fraig_CheckTfi_rec(p, Fraig_Regular(n.p2), old) != 0 {
		return 1
	}
	return Fraig_CheckTfi_rec(p, n.pNextE, old)
}
func Fraig_CheckTfi2(p *Fraig_Man_t, old, node *Fraig_Node_t) int32 {
	Fraig_DfsOne(p, node, 1)
	return fraigBool(old.TravId == p.nTravIds)
}
func Fraig_ManMarkRealFanouts(p *Fraig_Man_t) {
	nodes := Fraig_Dfs(p, 0)
	for _, n := range nodes.pArray {
		n.nFanouts = 0
		n.pData0 = nil
	}
	for _, n := range nodes.pArray {
		for _, edge := range []*Fraig_Node_t{n.p1, n.p2} {
			child := Fraig_Regular(edge)
			if child != nil && child.nFanouts < 2 {
				child.nFanouts++
			}
		}
	}
}
func Fraig_BitStringCountOnes(words []uint32, count int32) int32 {
	var n int32
	for _, w := range words[:count] {
		n += int32(bits.OnesCount32(w))
	}
	return n
}
func Fraig_ManCheckConsistency(p *Fraig_Man_t) int32 {
	for _, n := range Fraig_Dfs(p, 0).pArray {
		if n.pRepr != nil {
			fmt.Printf("Node %d is a secondary node.\n", n.Num)
		}
		if Fraig_NodeIsAnd(n) != 0 {
			for i, child := range []*Fraig_Node_t{n.p1, n.p2} {
				if Fraig_Regular(child).pRepr != nil {
					fmt.Printf("Internal node %d has fanin %d (%d) that is a secondary node.\n", n.Num, i+1, Fraig_Regular(child).Num)
				}
			}
		}
	}
	return 1
}
func Fraig_PrintBinary(w io.Writer, words []uint32, nBits int32) {
	for i := nBits - 1; i >= 0; i-- {
		fmt.Fprintf(w, "%d", (words[i/32]>>uint(i%32))&1)
	}
}
func Fraig_PrintNode(p *Fraig_Man_t, node *Fraig_Node_t) {
	for _, n := range Fraig_DfsOne(p, node, 0).pArray {
		if Fraig_NodeIsVar(n) != 0 {
			fmt.Printf("%3d : PI          ", n.Num)
		} else if Fraig_NodeIsConst(n) != 0 {
			fmt.Printf("%3d : CONST       ", n.Num)
		} else {
			fmt.Printf("%3d : %t%3d %t%3d   ", n.Num, Fraig_IsComplement(n.p1), Fraig_Regular(n.p1).Num, Fraig_IsComplement(n.p2), Fraig_Regular(n.p2).Num)
		}
		Fraig_PrintBinary(os.Stdout, n.puSimR, 20)
		fmt.Print("   ")
		Fraig_PrintBinary(os.Stdout, n.puSimD, 20)
		fmt.Printf("  %d\n", n.fInv)
	}
}
func Fraig_GetMaxLevel(p *Fraig_Man_t) int32 {
	var level int32
	for _, edge := range p.vOutputs.pArray {
		if n := Fraig_Regular(edge).Level; n > level {
			level = n
		}
	}
	return level
}
func Fraig_MappingUpdateLevel_rec(p *Fraig_Man_t, n *Fraig_Node_t, maximum int32) int32 {
	assert(!Fraig_IsComplement(n), "level node is regular")
	if Fraig_NodeIsAnd(n) == 0 || n.TravId == p.nTravIds {
		return n.Level
	}
	n.TravId = p.nTravIds
	a, b := Fraig_MappingUpdateLevel_rec(p, Fraig_Regular(n.p1), maximum), Fraig_MappingUpdateLevel_rec(p, Fraig_Regular(n.p2), maximum)
	n.Level = a + 1
	if b > a {
		n.Level = b + 1
	}
	if n.pNextE != nil {
		level := Fraig_MappingUpdateLevel_rec(p, n.pNextE, maximum)
		if maximum != 0 && n.Level < level || maximum == 0 && n.Level > level {
			n.Level = level
		}
		if n.pRepr == nil {
			for other := n.pNextE; other != nil; other = other.pNextE {
				other.Level = n.Level
			}
		}
	}
	return n.Level
}
func Fraig_MappingSetChoiceLevels(p *Fraig_Man_t, maximum int32) {
	p.nTravIds++
	for _, edge := range p.vOutputs.pArray {
		Fraig_MappingUpdateLevel_rec(p, Fraig_Regular(edge), maximum)
	}
}
func Fraig_ManReportChoices(p *Fraig_Man_t) {
	before := Fraig_GetMaxLevel(p)
	Fraig_MappingSetChoiceLevels(p, 0)
	after := Fraig_GetMaxLevel(p)
	nodes, choices := 0, 0
	for _, n := range p.vNodes.pArray {
		if n.pRepr == nil && n.pNextE != nil {
			nodes++
			for e := n; e != nil; e = e.pNextE {
				choices++
			}
		}
	}
	fmt.Printf("Maximum level: Original = %d. Reduced due to choices = %d.\nChoice stats:  Choice nodes = %d. Total choices = %d.\n", before, after, nodes, choices)
}
func fraigMuxChildren(node *Fraig_Node_t) (*Fraig_Node_t, *Fraig_Node_t) {
	n := Fraig_Regular(node)
	if Fraig_NodeIsAnd(n) == 0 || Fraig_NodeIsAnd(n.p1) == 0 || !Fraig_IsComplement(n.p1) || Fraig_NodeIsAnd(n.p2) == 0 || !Fraig_IsComplement(n.p2) {
		return nil, nil
	}
	a, b := Fraig_Regular(n.p1), Fraig_Regular(n.p2)
	assert(a.Num < b.Num, "ordered fanins")
	return a, b
}
func Fraig_NodeIsExorType(n *Fraig_Node_t) int32 {
	a, b := fraigMuxChildren(n)
	if a == nil {
		return 0
	}
	return fraigBool(a.p1 == Fraig_Not(b.p1) && a.p2 == Fraig_Not(b.p2))
}
func Fraig_NodeIsMuxType(n *Fraig_Node_t) int32 {
	a, b := fraigMuxChildren(n)
	if a == nil {
		return 0
	}
	return fraigBool(a.p1 == Fraig_Not(b.p1) || a.p1 == Fraig_Not(b.p2) || a.p2 == Fraig_Not(b.p1) || a.p2 == Fraig_Not(b.p2))
}
func Fraig_NodeIsExor(n *Fraig_Node_t) int32 {
	assert(!Fraig_IsComplement(n) && Fraig_NodeIsExorType(n) != 0, "regular XOR node")
	a := Fraig_Regular(n.p1)
	return fraigBool(Fraig_IsComplement(a.p1) == Fraig_IsComplement(a.p2))
}
func Fraig_NodeRecognizeMux(n *Fraig_Node_t, t, e **Fraig_Node_t) *Fraig_Node_t {
	assert(!Fraig_IsComplement(n) && Fraig_NodeIsMuxType(n) != 0, "regular MUX node")
	a, b := Fraig_Regular(n.p1), Fraig_Regular(n.p2)
	for _, pair := range []struct{ controlA, dataA, controlB, dataB *Fraig_Node_t }{{a.p1, a.p2, b.p1, b.p2}, {a.p1, a.p2, b.p2, b.p1}, {a.p2, a.p1, b.p1, b.p2}, {a.p2, a.p1, b.p2, b.p1}} {
		if pair.controlA == Fraig_Not(pair.controlB) {
			if Fraig_IsComplement(pair.controlA) {
				*t = Fraig_Not(pair.dataB)
				*e = Fraig_Not(pair.dataA)
				return pair.controlB
			}
			*t = Fraig_Not(pair.dataA)
			*e = Fraig_Not(pair.dataB)
			return pair.controlA
		}
	}
	panic("invalid MUX")
}
func Fraig_ManCountExors(p *Fraig_Man_t) int32 {
	var count int32
	for _, n := range p.vNodes.pArray {
		count += Fraig_NodeIsExorType(n)
	}
	return count
}
func Fraig_ManCountMuxes(p *Fraig_Man_t) int32 {
	var count int32
	for _, n := range p.vNodes.pArray {
		count += Fraig_NodeIsMuxType(n)
	}
	return count
}
func Fraig_NodeSimsContained(p *Fraig_Man_t, a, b *Fraig_Node_t) int32 {
	for _, v := range []struct {
		x, y []uint32
		n    int32
	}{{a.puSimR, b.puSimR, p.nWordsRand}, {a.puSimD, b.puSimD, p.iWordStart}} {
		for i := int32(0); i < v.n; i++ {
			if v.x[i]&^v.y[i] != 0 {
				return 0
			}
		}
	}
	return 1
}
func Fraig_CountPis(p *Fraig_Man_t, vars *msat.Msat_IntVec_t) int32 {
	var count int32
	for _, i := range msat.Msat_IntVecReadArray(vars) {
		count += Fraig_NodeIsVar(p.vNodes.pArray[i])
	}
	return count
}
func Fraig_ManPrintRefs(p *Fraig_Man_t) int32 {
	var pivots []*Fraig_Node_t
	for _, n := range p.vNodes.pArray {
		if n.nOnes != 0 && n.nOnes != uint32(p.nWordsRand)*32 && n.nRefs > 5 {
			pivots = append(pivots, n)
		}
	}
	fmt.Printf("Total nodes = %d. Referenced nodes = %d.\n", len(p.vNodes.pArray), len(pivots))
	count, proved := 0, 0
	for i, a := range pivots {
		for _, b := range pivots[i+1:] {
			if Fraig_NodeSimsContained(p, a, b) != 0 {
				count++
				if Fraig_NodeIsImplication(p, a, b, -1) != 0 {
					proved++
				}
			} else if Fraig_NodeSimsContained(p, b, a) != 0 {
				count++
				if Fraig_NodeIsImplication(p, b, a, -1) != 0 {
					proved++
				}
			}
		}
	}
	fmt.Printf("Number of candidate pairs = %d.  Proved = %d.\n", count, proved)
	return 0
}
func Fraig_NodeIsInSupergate(old, node *Fraig_Node_t) int32 {
	if Fraig_Regular(old) == Fraig_Regular(node) {
		if old == node {
			return 1
		}
		return -1
	}
	if Fraig_IsComplement(old) || Fraig_NodeIsVar(old) != 0 {
		return 0
	}
	a, b := Fraig_NodeIsInSupergate(old.p1, node), Fraig_NodeIsInSupergate(old.p2, node)
	if a == -1 || b == -1 {
		return -1
	}
	if a == 1 || b == 1 {
		return 1
	}
	return 0
}
func Fraig_CollectSupergate_rec(n *Fraig_Node_t, v *Fraig_NodeVec_t, first, stopMux int32) {
	if first == 0 && Fraig_Regular(n).nRefs > 1 || Fraig_IsComplement(n) || Fraig_NodeIsVar(n) != 0 || stopMux != 0 && Fraig_NodeIsMuxType(n) != 0 {
		Fraig_NodeVecPushUnique(v, n)
		return
	}
	Fraig_CollectSupergate_rec(n.p1, v, 0, stopMux)
	Fraig_CollectSupergate_rec(n.p2, v, 0, stopMux)
}
func Fraig_CollectSupergate(n *Fraig_Node_t, stopMux int32) *Fraig_NodeVec_t {
	v := Fraig_NodeVecAlloc(8)
	Fraig_CollectSupergate_rec(n, v, 1, stopMux)
	return v
}
func Fraig_ManIncrementTravId(p *Fraig_Man_t)                    { p.nTravIds2++ }
func Fraig_NodeSetTravIdCurrent(p *Fraig_Man_t, n *Fraig_Node_t) { n.TravId2 = p.nTravIds2 }
func Fraig_NodeIsTravIdCurrent(p *Fraig_Man_t, n *Fraig_Node_t) int32 {
	return fraigBool(n.TravId2 == p.nTravIds2)
}
func Fraig_NodeIsTravIdPrevious(p *Fraig_Man_t, n *Fraig_Node_t) int32 {
	return fraigBool(n.TravId2 == p.nTravIds2-1)
}
