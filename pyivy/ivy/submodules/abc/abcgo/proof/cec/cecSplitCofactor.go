package cec

// Cofactor selection from cecSplit.c by Alan Mishchenko (UC Berkeley, 2005).
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_PermuteSpecialOrder lists primary inputs in decreasing fanout order.
func Gia_PermuteSpecialOrder(p *gia.Gia_Man_t) []int32 {
	if p.Refs == nil {
		gia.Gia_ManCreateRefs(p)
	}
	costs := make([]int32, gia.Gia_ManPiNum(p))
	for i := range costs {
		costs[i] = gia.Gia_ObjRefNum(p, gia.Gia_ManPi(p, int32(i)))
	}
	return util.Abc_QuickSortCost(costs, int32(len(costs)), 1)
}

func Gia_PermuteSpecial(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	order := Gia_PermuteSpecialOrder(p)
	return gia.Gia_ManDupPerm(p, vec.Vec_IntAllocArrayCopy(order, len(order)))
}

func Gia_SplitCofVar2(p *gia.Gia_Man_t, fanouts, cost *int32) int32 {
	if p.Refs == nil {
		gia.Gia_ManCreateRefs(p)
	}
	best, bestCost := int32(-1), int32(-1)
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		refs := gia.Gia_ObjRefNum(p, gia.Gia_ManPi(p, i))
		if refs > bestCost {
			best, bestCost = i, refs
		}
	}
	if best < 0 {
		panic("cec: cofactor selection requires a primary input")
	}
	*fanouts, *cost = bestCost, -1
	return best
}

func Gia_SplitCofVar(p *gia.Gia_Man_t, lookAhead int32, fanouts, cost *int32) int32 {
	if lookAhead == 1 {
		return Gia_SplitCofVar2(p, fanouts, cost)
	}
	order := Gia_PermuteSpecialOrder(p)
	if lookAhead > int32(len(order)) {
		lookAhead = int32(len(order))
	}
	best, bestCost := int32(-1), int32(^uint32(0)>>1)
	for _, variable := range order[:lookAhead] {
		zero := gia.Gia_ManDupCofactorVar(p, variable, 0)
		one := gia.Gia_ManDupCofactorVar(p, variable, 1)
		current := gia.Gia_ManAndNum(zero) + gia.Gia_ManAndNum(one)
		gia.Gia_ManStop(zero)
		gia.Gia_ManStop(one)
		if current < bestCost {
			best, bestCost = variable, current
		}
	}
	if best < 0 {
		panic("cec: cofactor selection requires a positive lookahead")
	}
	*fanouts, *cost = gia.Gia_ObjRefNum(p, gia.Gia_ManPi(p, best)), bestCost
	return best
}
