package fraig

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/msat"
)

func Fraig_MarkTfi_rec(p *Fraig_Man_t, n *Fraig_Node_t) int32 {
	if n.TravId == p.nTravIds {
		return 0
	}
	n.TravId = p.nTravIds
	if Fraig_NodeIsConst(n) != 0 {
		return 0
	}
	if n.NumPi >= 0 {
		return 1
	}
	return Fraig_MarkTfi_rec(p, Fraig_Regular(n.p1)) + Fraig_MarkTfi_rec(p, Fraig_Regular(n.p2))
}
func Fraig_MarkTfi2_rec(p *Fraig_Man_t, n *Fraig_Node_t) int32 {
	if n.TravId == p.nTravIds {
		return 0
	}
	if n.TravId == p.nTravIds-1 {
		n.TravId = p.nTravIds
		return 1
	}
	n.TravId = p.nTravIds
	if Fraig_NodeIsConst(n) != 0 {
		return 0
	}
	if n.NumPi >= 0 {
		return 1
	}
	return Fraig_MarkTfi2_rec(p, Fraig_Regular(n.p1)) + Fraig_MarkTfi2_rec(p, Fraig_Regular(n.p2))
}
func Fraig_MarkTfi3_rec(p *Fraig_Man_t, n *Fraig_Node_t) int32 {
	if n.TravId == p.nTravIds {
		return 1
	}
	if n.TravId == p.nTravIds-1 {
		n.TravId = p.nTravIds
		return 1
	}
	n.TravId = p.nTravIds
	if Fraig_NodeIsConst(n) != 0 {
		return 1
	}
	if n.NumPi >= 0 {
		return 0
	}
	return Fraig_MarkTfi3_rec(p, Fraig_Regular(n.p1)) * Fraig_MarkTfi3_rec(p, Fraig_Regular(n.p2))
}
func Fraig_VarsStudy(p *Fraig_Man_t, old, node *Fraig_Node_t) {
	p.nTravIds++
	pis := Fraig_MarkTfi_rec(p, node)
	fmt.Printf("(%d)(%d,%d):", pis, old.Level, node.Level)
	if old.TravId == p.nTravIds {
		fmt.Print("* ")
		return
	}
	p.nTravIds++
	cut := Fraig_MarkTfi2_rec(p, old)
	fmt.Print(cut)
	p.nTravIds++
	contained := Fraig_MarkTfi3_rec(p, node)
	if contained != 0 {
		fmt.Print("+ ")
	} else {
		fmt.Print("- ")
	}
}
func Fraig_PrepareCones(p *Fraig_Man_t, old, node *Fraig_Node_t) {
	assert(old != node && !Fraig_IsComplement(old) && !Fraig_IsComplement(node), "distinct regular SAT roots")
	msat.Msat_IntVecFill(p.vVarsUsed, msat.Msat_IntVecReadSize(p.vVarsUsed), 0)
	msat.Msat_IntVecClear(p.vVarsInt)
	p.nTravIds++
	Fraig_PrepareCones_rec(p, node)
	Fraig_PrepareCones_rec(p, old)
}
func Fraig_PrepareCones_rec(p *Fraig_Man_t, n *Fraig_Node_t) {
	assert(!Fraig_IsComplement(n), "regular SAT cone node")
	if n.TravId == p.nTravIds {
		return
	}
	n.TravId = p.nTravIds
	msat.Msat_IntVecPush(p.vVarsInt, int(n.Num))
	msat.Msat_IntVecWriteEntry(p.vVarsUsed, int(n.Num), 1)
	if Fraig_NodeIsAnd(n) == 0 {
		return
	}
	fresh := n.vFanins == nil
	if fresh {
		assert(n.fClauses == 0, "new supergate has no clauses")
		if Fraig_NodeIsMuxType(n) != 0 {
			n.vFanins = Fraig_NodeVecAlloc(4)
			a, b := Fraig_Regular(n.p1), Fraig_Regular(n.p2)
			for _, edge := range []*Fraig_Node_t{a.p1, a.p2, b.p1, b.p2} {
				Fraig_NodeVecPushUnique(n.vFanins, Fraig_Regular(edge))
			}
			Fraig_SupergateAddClausesMux(p, n)
		} else {
			n.vFanins = Fraig_CollectSupergate(n, 1)
			Fraig_SupergateAddClauses(p, n, n.vFanins)
		}
		assert(Fraig_NodeVecReadSize(n.vFanins) > 1, "multiple fanins")
		n.fClauses = 1
		p.nVarsClauses++
		adj := p.vAdjacents[n.Num]
		assert(msat.Msat_IntVecReadSize(adj) == 0, "empty adjacency")
		for _, edge := range n.vFanins.pArray {
			msat.Msat_IntVecPush(adj, int(Fraig_Regular(edge).Num))
		}
	}
	for _, edge := range n.vFanins.pArray {
		Fraig_PrepareCones_rec(p, Fraig_Regular(edge))
	}
	if fresh {
		for _, edge := range n.vFanins.pArray {
			msat.Msat_IntVecPush(p.vAdjacents[Fraig_Regular(edge).Num], int(n.Num))
		}
	}
}
func Fraig_SetupAdjacent(p *Fraig_Man_t, vars *msat.Msat_IntVec_t) {
	for _, index := range msat.Msat_IntVecReadArray(vars) {
		adj := p.vAdjacents[index]
		msat.Msat_IntVecClear(adj)
		n := p.vNodes.pArray[index]
		if Fraig_NodeIsAnd(n) != 0 {
			for _, edge := range n.vFanins.pArray {
				msat.Msat_IntVecPush(adj, int(Fraig_Regular(edge).Num))
			}
		}
	}
	for _, index := range msat.Msat_IntVecReadArray(vars) {
		n := p.vNodes.pArray[index]
		if Fraig_NodeIsAnd(n) != 0 {
			for _, edge := range n.vFanins.pArray {
				msat.Msat_IntVecPush(p.vAdjacents[Fraig_Regular(edge).Num], index)
			}
		}
	}
}
func Fraig_DetectFanoutFreeCone_rec(n *Fraig_Node_t, super, inside *Fraig_NodeVec_t, first int32) {
	n = Fraig_Regular(n)
	if first == 0 && n.nRefs > 1 || Fraig_NodeIsVar(n) != 0 {
		Fraig_NodeVecPushUnique(super, n)
		return
	}
	Fraig_DetectFanoutFreeCone_rec(n.p1, super, inside, 0)
	Fraig_DetectFanoutFreeCone_rec(n.p2, super, inside, 0)
	Fraig_NodeVecPushUnique(inside, n)
}
func Fraig_DetectFanoutFreeConeMux_rec(n *Fraig_Node_t, super, inside *Fraig_NodeVec_t, first int32) {
	n = Fraig_Regular(n)
	if first == 0 && n.nRefs > 1 || Fraig_NodeIsVar(n) != 0 || Fraig_NodeIsMuxType(n) == 0 {
		Fraig_NodeVecPushUnique(super, n)
		return
	}
	a, b := Fraig_Regular(n.p1), Fraig_Regular(n.p2)
	for _, edge := range []*Fraig_Node_t{a.p1, a.p2, b.p1, b.p2} {
		Fraig_DetectFanoutFreeConeMux_rec(edge, super, inside, 0)
	}
	Fraig_NodeVecPushUnique(inside, n)
}
func Fraig_DetectFanoutFreeConeMux(p *Fraig_Man_t, n *Fraig_Node_t) {
	fanins, inside := Fraig_NodeVecAlloc(8), Fraig_NodeVecAlloc(8)
	Fraig_DetectFanoutFreeConeMux_rec(n, fanins, inside, 1)
	assert(inside.pArray[len(inside.pArray)-1] == n, "cone root is last")
	fmt.Printf("%d(%d)", len(fanins.pArray), 0)
}
