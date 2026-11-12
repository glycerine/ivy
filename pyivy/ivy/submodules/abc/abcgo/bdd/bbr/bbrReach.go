// Partitioned breadth-first BDD reachability from bbrReach.c.
package bbr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"time"
)

func Bbr_ManSetDefaultParams(p *Saig_ParBbr_t) {
	if p != nil {
		*p = Saig_ParBbr_t{NBddMax: 50000, NIterMax: 1000, FPartition: 1, FReorder: 1, FReorderImage: 1, IFrame: -1}
	}
}
func Bbr_bddComputeRangeCube(dd *DdManager, start, stop int) *DdNode {
	if dd == nil || start < 0 || stop < start || stop > len(dd.Vars) {
		return nil
	}
	f := dd.One
	for _, v := range dd.Vars[start:stop] {
		f = cudd.Cudd_bddAnd(dd, f, v)
	}
	return f
}
func Bbr_StopManager(dd *DdManager) { cudd.Cudd_Quit(dd) }
func Aig_ManInitStateVarMap(dd *DdManager, p *Aig_Man_t, verbose bool) *DdNode {
	pis, regs, _, ok := dimensions(p)
	if !ok || dd == nil || len(dd.Vars) < pis+2*regs {
		return nil
	}
	f := dd.One
	for _, v := range dd.Vars[pis : pis+regs] {
		f = cudd.Cudd_bddAnd(dd, f, cudd.Cudd_Not(v))
	}
	return f
}
func bbrVarMap(dd *DdManager, p *Aig_Man_t, f *DdNode) *DdNode {
	pis, regs, _, ok := dimensions(p)
	if !ok {
		return nil
	}
	perm := make([]int, len(dd.Vars))
	for i := range perm {
		perm[i] = i
	}
	for i := 0; i < regs; i++ {
		perm[pis+i], perm[pis+regs+i] = pis+regs+i, pis+i
	}
	return cudd.Cudd_bddPermute(dd, f, perm)
}
func Aig_ManCreateOutputs(dd *DdManager, p *Aig_Man_t) []*DdNode {
	_, _, pos, ok := dimensions(p)
	if !ok {
		return nil
	}
	out := make([]*DdNode, pos)
	for i, e := range p.Outputs()[:pos] {
		out[i] = Aig_ObjGlobalBdd(p, e)
		if out[i] == nil {
			for _, f := range out {
				cudd.Cudd_RecursiveDeref(dd, f)
			}
			return nil
		}
		cudd.Cudd_Ref(out[i])
	}
	return out
}
func Aig_ManCreatePartitions(dd *DdManager, p *Aig_Man_t, reorder, verbose bool) []*DdNode {
	pis, regs, pos, ok := dimensions(p)
	if !ok || dd == nil || len(dd.Vars) != pis+regs {
		return nil
	}
	if regs > 0 {
		cudd.Cudd_bddIthVar(dd, pis+2*regs-1)
	}
	parts := make([]*DdNode, regs)
	for i, e := range p.Outputs()[pos:] {
		f := Aig_ObjGlobalBdd(p, e)
		if f == nil {
			for _, part := range parts {
				cudd.Cudd_RecursiveDeref(dd, part)
			}
			return nil
		}
		parts[i] = cudd.Cudd_bddXnor(dd, dd.Vars[pis+regs+i], f)
		cudd.Cudd_Ref(parts[i])
	}
	Aig_ManFreeGlobalBdds(p, dd)
	if reorder {
		cudd.Cudd_ReduceHeap(dd, cudd.CUDD_REORDER_SYMM_SIFT, 100)
	}
	return parts
}

// Aig_ManComputeReachable returns 0 for a replayable failing output, 1 for a
// fixed point, and -1 for exhausted iteration/node/time budget or invalid input.
func Aig_ManComputeReachable(dd *DdManager, p *Aig_Man_t, parts []*DdNode, initial *DdNode, outputs []*DdNode, params *Saig_ParBbr_t, check bool) int32 {
	pis, regs, pos, ok := dimensions(p)
	if !ok || dd == nil || params == nil || len(parts) != regs || len(outputs) != pos || len(dd.Vars) < pis+2*regs || params.NIterMax < 0 || params.NBddMax < 0 || !cudd.Cudd_bddLeq(dd, initial, initial) {
		return -1
	}
	start := time.Now()
	care := Bbr_bddComputeRangeCube(dd, pis, pis+regs)
	var tree *Bbr_ImageTree_t
	var mono *Bbr_ImageTree2_t
	if params.FPartition != 0 {
		tree = Bbr_bddImageStart(dd, care, regs, parts, regs, dd.Vars[pis+regs:], int(params.NBddMax), params.FVerbose != 0)
		if tree == nil {
			return -1
		}
		defer Bbr_bddImageTreeDelete(tree)
	} else {
		mono = Bbr_bddImageStart2(dd, care, regs, parts, regs, dd.Vars[pis+regs:], params.FVerbose != 0)
		if mono == nil {
			return -1
		}
		defer Bbr_bddImageTreeDelete2(mono)
	}
	current, reached := initial, initial
	cudd.Cudd_Ref(current)
	cudd.Cudd_Ref(reached)
	rings := []*DdNode{initial}
	cudd.Cudd_Ref(initial)
	defer func() {
		cudd.Cudd_RecursiveDeref(dd, current)
		cudd.Cudd_RecursiveDeref(dd, reached)
		for _, ring := range rings {
			cudd.Cudd_RecursiveDeref(dd, ring)
		}
	}()
	for iter := 0; iter < int(params.NIterMax); iter++ {
		if params.TimeLimit > 0 && time.Since(start) >= time.Duration(params.TimeLimit)*time.Second {
			params.IFrame = int32(iter - 1)
			return -1
		}
		var next *DdNode
		if tree != nil {
			next = Bbr_bddImageCompute(tree, current)
		} else {
			next = Bbr_bddImageCompute2(mono, current)
		}
		if next == nil {
			params.IFrame = int32(iter - 1)
			return -1
		}
		next = bbrVarMap(dd, p, next)
		if next == nil {
			return -1
		}
		cudd.Cudd_Ref(next)
		if cudd.Cudd_bddLeq(dd, next, reached) {
			cudd.Cudd_RecursiveDeref(dd, next)
			params.IFrame = int32(iter - 1)
			return 1
		}
		if cudd.Cudd_DagSize(next) > int(params.NBddMax) {
			cudd.Cudd_RecursiveDeref(dd, next)
			params.IFrame = int32(iter - 1)
			return -1
		}
		if check {
			for i, output := range outputs {
				if !cudd.Cudd_bddLeq(dd, next, cudd.Cudd_Not(output)) {
					intersection := cudd.Cudd_bddAnd(dd, next, output)
					cex := Aig_ManVerifyUsingBddsCountExample(p, dd, parts, rings, intersection, i, params.FVerbose != 0, params.FSilent != 0)
					cudd.Cudd_RecursiveDeref(dd, next)
					params.IFrame = int32(iter)
					if cex == nil || p.SetSequentialCounterexample(cex) != nil {
						return -1
					}
					return 0
				}
			}
		}
		frontier := cudd.Cudd_bddAnd(dd, next, cudd.Cudd_Not(reached))
		cudd.Cudd_Ref(frontier)
		cudd.Cudd_RecursiveDeref(dd, current)
		current = frontier
		rings = append(rings, frontier)
		cudd.Cudd_Ref(frontier)
		union := cudd.Cudd_bddOr(dd, reached, next)
		cudd.Cudd_Ref(union)
		cudd.Cudd_RecursiveDeref(dd, reached)
		reached = union
		cudd.Cudd_RecursiveDeref(dd, next)
		if params.FVerbose != 0 {
			fmt.Printf("Frame = %3d. BDD = %5d.\n", iter, cudd.Cudd_DagSize(frontier))
		}
	}
	params.IFrame = params.NIterMax - 1
	return -1
}
func Aig_ManVerifyUsingBdds_int(p *Aig_Man_t, params *Saig_ParBbr_t) int32 {
	_, regs, _, ok := dimensions(p)
	if !ok || params == nil || regs <= 0 || params.NBddMax < 0 || params.NIterMax < 0 {
		return -1
	}
	start := time.Now()
	dd := Aig_ManComputeGlobalBdds(p, int(params.NBddMax), true, params.FReorder != 0, params.FVerbose != 0)
	if dd == nil {
		return -1
	}
	defer cudd.Cudd_Quit(dd)
	defer Aig_ManFreeGlobalBdds(p, dd)
	if params.TimeLimit > 0 && time.Since(start) >= time.Duration(params.TimeLimit)*time.Second {
		return -1
	}
	outputs := Aig_ManCreateOutputs(dd, p)
	if outputs == nil {
		return -1
	}
	defer func() {
		for _, f := range outputs {
			cudd.Cudd_RecursiveDeref(dd, f)
		}
	}()
	parts := Aig_ManCreatePartitions(dd, p, params.FReorder != 0, params.FVerbose != 0)
	if parts == nil {
		return -1
	}
	defer func() {
		for _, f := range parts {
			cudd.Cudd_RecursiveDeref(dd, f)
		}
	}()
	initial := Aig_ManInitStateVarMap(dd, p, params.FVerbose != 0)
	if initial == nil {
		return -1
	}
	cudd.Cudd_Ref(initial)
	defer cudd.Cudd_RecursiveDeref(dd, initial)
	if params.FSkipOutCheck == 0 {
		for i, output := range outputs {
			if !cudd.Cudd_bddLeq(dd, initial, cudd.Cudd_Not(output)) {
				cex := Aig_ManVerifyUsingBddsCountExample(p, dd, parts, nil, cudd.Cudd_bddAnd(dd, initial, output), i, params.FVerbose != 0, params.FSilent != 0)
				if cex == nil || p.SetSequentialCounterexample(cex) != nil {
					return -1
				}
				return 0
			}
		}
	}
	return Aig_ManComputeReachable(dd, p, parts, initial, outputs, params, params.FSkipOutCheck == 0)
}
func Aig_ManVerifyUsingBdds(p *Aig_Man_t, params *Saig_ParBbr_t) int32 {
	pis, _, _, ok := dimensions(p)
	if !ok || params == nil {
		return -1
	}
	needsTrim := false
	for _, e := range p.Inputs()[:pis] {
		if aig.Aig_ObjRefs(p, e) == 0 {
			needsTrim = true
			break
		}
	}
	if !needsTrim {
		return Aig_ManVerifyUsingBdds_int(p, params)
	}
	trimmed := aig.Aig_ManDupTrim(p)
	defer aig.Aig_ManStop(trimmed)
	mapping := make([]int, pis)
	for i, e := range p.Inputs()[:pis] {
		mapping[i] = -1
		copy := aig.Aig_ObjCopy(p, e)
		if copy != aig.InvalidEdge {
			mapping[i] = int(aig.Aig_ObjCioId(trimmed, copy))
		}
	}
	status := Aig_ManVerifyUsingBdds_int(trimmed, params)
	if status != 0 {
		return status
	}
	old := trimmed.SequentialCounterexample()
	if old == nil {
		return -1
	}
	cex := util.Abc_CexAlloc(old.NRegs, int32(pis), old.IFrame+1)
	if cex == nil {
		return -1
	}
	cex.IFrame, cex.IPo = old.IFrame, old.IPo
	copyBit := func(src, dst int) {
		if old.PData[src>>5]&(1<<uint(src&31)) != 0 {
			cex.PData[dst>>5] |= 1 << uint(dst&31)
		}
	}
	for i := 0; i < int(old.NRegs); i++ {
		copyBit(i, i)
	}
	for frame := 0; frame <= int(old.IFrame); frame++ {
		for i, j := range mapping {
			if j >= 0 {
				copyBit(int(old.NRegs)+frame*int(old.NPis)+j, int(cex.NRegs)+frame*pis+i)
			}
		}
	}
	if saig.Saig_ManVerifyCex(p, cex) != 1 || p.SetSequentialCounterexample(cex) != nil {
		return -1
	}
	return 0
}
