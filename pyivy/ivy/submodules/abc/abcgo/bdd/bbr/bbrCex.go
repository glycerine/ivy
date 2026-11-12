// Backward onion-ring counterexample extraction from bbrCex.c.
package bbr

import (
	"github.com/glycerine/abcgo/aig/saig"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
)

// pickCube chooses the else branch first and fills don't-care variables with
// zero, giving a complete assignment suitable for replay and next-state cubes.
func pickCube(dd *DdManager, f *DdNode) []bool {
	if dd == nil || f == nil || f == dd.LogicZero || !cudd.Cudd_bddLeq(dd, f, f) {
		return nil
	}
	values := make([]bool, len(dd.Vars))
	for !cudd.Cudd_IsConstant(f) {
		n := cudd.Cudd_Regular(f)
		lo := cudd.Cudd_NotCond(n.E, cudd.Cudd_IsComplement(f))
		hi := cudd.Cudd_NotCond(n.T, cudd.Cudd_IsComplement(f))
		if lo != dd.LogicZero {
			f = lo
		} else {
			values[n.Index] = true
			f = hi
		}
	}
	if f != dd.One {
		return nil
	}
	return values
}
func bbrStateCube(dd *DdManager, values []bool, start, source, n int) *DdNode {
	f := dd.One
	for i := 0; i < n; i++ {
		if source+i >= len(values) || start+i >= len(dd.Vars) {
			return nil
		}
		v := cudd.Cudd_NotCond(dd.Vars[start+i], !values[source+i])
		f = cudd.Cudd_bddAnd(dd, f, v)
	}
	return f
}
func Aig_ManVerifyUsingBddsCountExample(p *Aig_Man_t, dd *DdManager, parts []*DdNode, rings []*DdNode, first *DdNode, output int, verbose, silent bool) *Abc_Cex_t {
	pis, regs, pos, ok := dimensions(p)
	if !ok || dd == nil || len(parts) != regs || len(dd.Vars) < pis+2*regs || output < 0 || output >= pos {
		return nil
	}
	values := pickCube(dd, first)
	if values == nil {
		return nil
	}
	cex := util.Abc_CexAlloc(int32(regs), int32(pis), int32(len(rings)+1))
	if cex == nil {
		return nil
	}
	cex.IPo = int32(output)
	cex.IFrame = int32(len(rings))
	writeInputs := func(frame int, values []bool) {
		for i := 0; i < pis; i++ {
			if values[i] {
				bit := regs + frame*pis + i
				cex.PData[bit>>5] |= 1 << uint(bit&31)
			}
		}
	}
	writeInputs(len(rings), values)
	if len(rings) == 0 {
		for i := 0; i < regs; i++ {
			if values[pis+i] {
				return nil
			}
		}
		if saig.Saig_ManVerifyCex(p, cex) != 1 {
			return nil
		}
		return cex
	}
	nsCube := Bbr_bddComputeRangeCube(dd, pis+regs, pis+2*regs)
	tree := Bbr_bddImageStart(dd, nsCube, regs, parts, pis+regs, dd.Vars, 100000000, verbose)
	if tree == nil {
		return nil
	}
	defer Bbr_bddImageTreeDelete(tree)
	state := bbrStateCube(dd, values, pis+regs, pis, regs)
	cudd.Cudd_Ref(state)
	defer func() { cudd.Cudd_RecursiveDeref(dd, state) }()
	for frame := len(rings) - 1; frame >= 0; frame-- {
		image := Bbr_bddImageCompute(tree, state)
		if image == nil {
			return nil
		}
		intersection := cudd.Cudd_bddAnd(dd, image, rings[frame])
		values = pickCube(dd, intersection)
		if values == nil {
			return nil
		}
		writeInputs(frame, values)
		if frame == 0 {
			for i := 0; i < regs; i++ {
				if values[pis+i] {
					return nil
				}
			}
			break
		}
		next := bbrStateCube(dd, values, pis+regs, pis, regs)
		cudd.Cudd_Ref(next)
		cudd.Cudd_RecursiveDeref(dd, state)
		state = next
	}
	if saig.Saig_ManVerifyCex(p, cex) != 1 {
		return nil
	}
	return cex
}
