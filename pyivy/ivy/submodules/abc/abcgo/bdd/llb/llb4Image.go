package llb

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func Llb_Nonlin4RemoveVar(p *Llb_Mgr_t, v *Llb_Var_t) {
	if p == nil || v == nil {
		return
	}
	for _, id := range vec.Vec_IntArray(v.vParts) {
		if int(id) < len(p.pParts) {
			p.pParts[id] = nil
		}
	}
}
func Llb_Nonlin4RemovePart(p *Llb_Mgr_t, part *Llb_Prt_t) {
	if p == nil || part == nil {
		return
	}
	if int(part.iPart) < len(p.pParts) {
		p.pParts[part.iPart] = nil
	}
}
func Llb_Nonlin4CreateCube1(p *Llb_Mgr_t, part *Llb_Prt_t) *DdNode {
	return llbNonlinCube(p, part, nil)
}
func Llb_Nonlin4CreateCube2(p *Llb_Mgr_t, a, b *Llb_Prt_t) *DdNode { return llbNonlinCube(p, a, b) }
func Llb_Nonlin4HasSingletonVars(p *Llb_Mgr_t, part *Llb_Prt_t) int32 {
	if p == nil || p.dd == nil || part == nil || part.bFunc == nil {
		return 0
	}
	if core.Cudd_SupportSize(p.dd, part.bFunc) == 1 {
		return 1
	}
	return 0
}
func Llb_Nonlin4Print(p *Llb_Mgr_t) {}
func Llb_Nonlin4Quantify1(p *Llb_Mgr_t, part *Llb_Prt_t) int32 {
	if p == nil || part == nil {
		return 0
	}
	return 1
}
func Llb_Nonlin4Quantify2(p *Llb_Mgr_t, first, second *Llb_Prt_t) int32 {
	if p == nil || p.dd == nil || first == nil || second == nil {
		return 0
	}
	merged := core.Cudd_bddAnd(p.dd, first.bFunc, second.bFunc)
	if merged == nil {
		return 0
	}
	first.bFunc = merged
	return 1
}
func Llb_Nonlin4CutNodes(g *Aig_Man_t, lower, upper *Vec_Ptr_t) *Vec_Ptr_t {
	return Llb_ManCutNodes(g, lower, upper)
}
func Llb_Nonlin4AddPair(p *Llb_Mgr_t, part, variable int32) {
	if p == nil {
		return
	}
	for len(p.pParts) <= int(part) {
		p.pParts = append(p.pParts, nil)
	}
	if p.pParts[part] == nil {
		p.pParts[part] = &Llb_Prt_t{iPart: part, vVars: vec.Vec_IntAlloc(4)}
	}
	vec.Vec_IntPush(p.pParts[part].vVars, variable)
}
func Llb_Nonlin4AddPartition(p *Llb_Mgr_t, part int32, f *DdNode) { Llb_NonlinAddPartition(p, part, f) }
func Llb_Nonlin4CheckVars(p *Llb_Mgr_t) {
	if p == nil {
		return
	}
	for _, part := range p.pParts {
		if part != nil && part.bFunc != nil && p.dd != nil {
			part.vVars = vec.Vec_IntAlloc(len(p.dd.Support(part.bFunc)))
			for _, i := range p.dd.Support(part.bFunc) {
				vec.Vec_IntPush(part.vVars, int32(i))
			}
		}
	}
}
func Llb_Nonlin4NextPartitions(p *Llb_Mgr_t, a, b **Llb_Prt_t) int32 {
	return Llb_NonlinNextPartitions(p, a, b)
}
func Llb_Nonlin4RecomputeScores(p *Llb_Mgr_t) {
	if p == nil || p.dd == nil {
		return
	}
	for _, part := range p.pParts {
		if part != nil && part.bFunc != nil {
			part.nSize = int32(core.Cudd_DagSize(part.bFunc))
		}
	}
}
func Llb_Nonlin4VerifyScores(p *Llb_Mgr_t) { Llb_Nonlin4RecomputeScores(p) }
func Llb_Nonlin4Alloc(dd *DdManager, parts *Vec_Ptr_t, current *DdNode, quant *Vec_Int_t, maxSize int32) *Llb_Mgr_t {
	return &Llb_Mgr_t{dd: dd}
}
func Llb_Nonlin4Free(p *Llb_Mgr_t) { Llb_NonlinFree(p) }

// Llb_Nonlin4Image conjoins the current state with transition partitions and
// existentially abstracts variables listed by quantification order.
func Llb_Nonlin4Image(dd *DdManager, parts *Vec_Ptr_t, current *DdNode, quant *Vec_Int_t) *DdNode {
	if dd == nil {
		return nil
	}
	result := core.Cudd_ReadOne(dd)
	if current != nil {
		result = core.Cudd_bddAnd(dd, result, current)
	}
	if parts != nil {
		for _, value := range vec.Vec_PtrArray(parts) {
			if f, ok := value.(*core.DdNode); ok && f != nil {
				result = core.Cudd_bddAnd(dd, result, f)
				if result == nil {
					return nil
				}
			}
		}
	}
	if quant == nil || vec.Vec_IntSize(quant) == 0 {
		return result
	}
	cube := core.Cudd_ReadOne(dd)
	for _, index := range vec.Vec_IntArray(quant) {
		if index < 0 || int(index) >= len(dd.Vars) {
			return nil
		}
		cube = core.Cudd_bddAnd(dd, cube, core.Cudd_bddIthVar(dd, int(index)))
	}
	return core.Cudd_bddExistAbstract(dd, result, cube)
}
func Llb_Nonlin4Group(dd *DdManager, parts *Vec_Ptr_t, quant *Vec_Int_t, maxSize int32) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(1)
	if parts != nil {
		vec.Vec_PtrPush(result, parts)
	}
	return result
}
