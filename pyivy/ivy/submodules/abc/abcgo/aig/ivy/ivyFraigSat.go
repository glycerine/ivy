// Incremental SAT cone encoding and equivalence checking from ivyFraig.c.
package ivy

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"math"
	"slices"
	"time"
)

func ivyFraigSolver(p *Ivy_FraigMan_t) {
	if p.pSat != nil {
		return
	}
	p.pSat = bsat.NewSolver()
	p.pSat.NewVariable()
	p.pSat.AddClause(bsat.Positive(0))
	p.nSatVars = 1
}
func ivyFraigLiteral(n *Ivy_Obj_t) bsat.Literal {
	variable := Ivy_ObjSatNum(Ivy_Regular(n))
	if Ivy_IsComplement(n) != 0 {
		return bsat.Negative(variable)
	}
	return bsat.Positive(variable)
}
func ivyFraigNeg(literal bsat.Literal) bsat.Literal { return literal ^ 1 }
func Ivy_FraigCollectSuper_rec(edge *Ivy_Obj_t, super *[]*Ivy_Obj_t, first, useMuxes int32) {
	n := Ivy_Regular(edge)
	if Ivy_IsComplement(edge) != 0 || Ivy_ObjIsPi(n) != 0 || Ivy_ObjIsConst1(n) != 0 || n.Type != IVY_AND || (first == 0 && n.nRefs > 1) || (useMuxes != 0 && Ivy_ObjIsMuxType(n) != 0) {
		ivyCutPushUnique(super, edge)
		return
	}
	Ivy_FraigCollectSuper_rec(n.pFanin0, super, 0, useMuxes)
	Ivy_FraigCollectSuper_rec(n.pFanin1, super, 0, useMuxes)
}
func Ivy_FraigCollectSuper(n *Ivy_Obj_t, useMuxes int32) []*Ivy_Obj_t {
	if n == nil || Ivy_IsComplement(n) != 0 || Ivy_ObjIsNode(n) == 0 {
		panic("ivy: supergate root required")
	}
	var super []*Ivy_Obj_t
	Ivy_FraigCollectSuper_rec(n, &super, 1, useMuxes)
	return super
}
func Ivy_FraigObjAddToFrontier(p *Ivy_FraigMan_t, n *Ivy_Obj_t, front *[]*Ivy_Obj_t) {
	ivyFraigSolver(p)
	if n == nil || Ivy_IsComplement(n) != 0 || n.owner != p.pManFraig {
		panic("ivy: foreign SAT cone node")
	}
	if n == p.pManFraig.pConst1 || Ivy_ObjSatNum(n) != 0 {
		return
	}
	variable := p.pSat.NewVariable()
	Ivy_ObjSetSatNum(n, variable)
	p.nSatVars = p.pSat.NumVars()
	if Ivy_ObjIsNode(n) != 0 {
		*front = append(*front, n)
	} else if Ivy_ObjIsPi(n) != 0 {
		p.vPiVars = append(p.vPiVars, n)
	} else {
		panic("ivy: unsupported SAT cone node")
	}
}
func Ivy_FraigAddClausesSuper(p *Ivy_FraigMan_t, n *Ivy_Obj_t, super []*Ivy_Obj_t) {
	output := ivyFraigLiteral(n)
	large := make([]bsat.Literal, 0, len(super)+1)
	for _, edge := range super {
		input := ivyFraigLiteral(edge)
		p.pSat.AddClause(input, ivyFraigNeg(output))
		large = append(large, ivyFraigNeg(input))
	}
	large = append(large, output)
	p.pSat.AddClause(large...)
}
func Ivy_FraigAddClausesMux(p *Ivy_FraigMan_t, n *Ivy_Obj_t) {
	var then, otherwise *Ivy_Obj_t
	control := Ivy_ObjRecognizeMux(n, &then, &otherwise)
	c, t, e, f := ivyFraigLiteral(control), ivyFraigLiteral(then), ivyFraigLiteral(otherwise), ivyFraigLiteral(n)
	p.pSat.AddClause(ivyFraigNeg(c), ivyFraigNeg(t), f)
	p.pSat.AddClause(ivyFraigNeg(c), t, ivyFraigNeg(f))
	p.pSat.AddClause(c, ivyFraigNeg(e), f)
	p.pSat.AddClause(c, e, ivyFraigNeg(f))
	if Ivy_ObjSatNum(then) != Ivy_ObjSatNum(otherwise) {
		p.pSat.AddClause(t, e, ivyFraigNeg(f))
		p.pSat.AddClause(ivyFraigNeg(t), ivyFraigNeg(e), f)
	}
}
func Ivy_FraigNodeAddToSolver(p *Ivy_FraigMan_t, old, next *Ivy_Obj_t) {
	ivyFraigSolver(p)
	var front []*Ivy_Obj_t
	if old != nil {
		Ivy_FraigObjAddToFrontier(p, old, &front)
	}
	if next != nil {
		Ivy_FraigObjAddToFrontier(p, next, &front)
	}
	for i := 0; i < len(front); i++ {
		n := front[i]
		var fanins []*Ivy_Obj_t
		switch {
		case n.Type == IVY_EXOR:
			fanins = []*Ivy_Obj_t{n.pFanin0, n.pFanin1}
		case Ivy_ObjIsMuxType(n) != 0:
			var then, otherwise *Ivy_Obj_t
			control := Ivy_ObjRecognizeMux(n, &then, &otherwise)
			for _, edge := range []*Ivy_Obj_t{control, then, otherwise} {
				ivyCutPushUnique(&fanins, Ivy_Regular(edge))
			}
		default:
			fanins = Ivy_FraigCollectSuper(n, 1)
		}
		for _, edge := range fanins {
			Ivy_FraigObjAddToFrontier(p, Ivy_Regular(edge), &front)
		}
		switch {
		case n.Type == IVY_EXOR:
			a, b, f := ivyFraigLiteral(n.pFanin0), ivyFraigLiteral(n.pFanin1), ivyFraigLiteral(n)
			p.pSat.AddClause(a, b, ivyFraigNeg(f))
			p.pSat.AddClause(ivyFraigNeg(a), ivyFraigNeg(b), ivyFraigNeg(f))
			p.pSat.AddClause(a, ivyFraigNeg(b), f)
			p.pSat.AddClause(ivyFraigNeg(a), b, f)
		case Ivy_ObjIsMuxType(n) != 0:
			Ivy_FraigAddClausesMux(p, n)
		default:
			Ivy_FraigAddClausesSuper(p, n, fanins)
		}
		Ivy_ObjSetFaninVec(n, fanins)
	}
}
func Ivy_FraigSetActivityFactors_rec(p *Ivy_FraigMan_t, n *Ivy_Obj_t, minimum, maximum int32) int32 {
	if Ivy_ObjIsTravIdCurrent(p.pManFraig, n) != 0 {
		return 0
	}
	Ivy_ObjSetTravIdCurrent(p.pManFraig, n)
	if n.Level <= minimum || Ivy_ObjIsPi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
		return 0
	}
	p.pSat.SetActivityFactor(Ivy_ObjSatNum(n), p.pParams.dActConeBumpMax*float64(n.Level-minimum)/float64(maximum-minimum))
	count := int32(1)
	for _, edge := range Ivy_ObjFaninVec(n) {
		count += Ivy_FraigSetActivityFactors_rec(p, Ivy_Regular(edge), minimum, maximum)
	}
	return count
}
func Ivy_FraigSetActivityFactors(p *Ivy_FraigMan_t, old, next *Ivy_Obj_t) int32 {
	start := time.Now()
	p.pSat.ClearActivityFactors()
	Ivy_ManIncrementTravId(p.pManFraig)
	maximum := int32(0)
	if old != nil {
		maximum = old.Level
	}
	if next != nil {
		maximum = max(maximum, next.Level)
	}
	minimum := int32(float64(maximum) * (1 - p.pParams.dActConeRatio))
	if minimum < 0 {
		minimum = 0
	}
	if minimum >= maximum {
		minimum = maximum
	}
	for _, n := range []*Ivy_Obj_t{old, next} {
		if n != nil && n != p.pManFraig.pConst1 {
			Ivy_FraigSetActivityFactors_rec(p, n, minimum, maximum)
		}
	}
	p.timeTrav += time.Since(start)
	return 1
}
func Ivy_FraigPrintActivity(p *Ivy_FraigMan_t) {
	for variable := int32(0); variable < p.nSatVars; variable++ {
		fmt.Printf("%d %d  ", variable, int64(p.pSat.VariableActivity(variable)))
	}
	fmt.Println()
}

func Ivy_FraigCreateModel(p *Ivy_FraigMan_t) []int32 {
	model := make([]int32, len(p.pManFraig.vPis))
	for i, n := range p.pManFraig.vPis {
		if Ivy_ObjSatNum(n) == 0 {
			continue
		}
		value, known := p.pSat.Value(Ivy_ObjSatNum(n))
		if known && value {
			model[i] = 1
		}
	}
	return model
}
func Ivy_FraigSavePattern(p *Ivy_FraigMan_t) {
	clear(p.pPatWords)
	for i, value := range Ivy_FraigCreateModel(p) {
		if value != 0 {
			p.pPatWords[i/32] |= 1 << uint(i%32)
		}
	}
}
func Ivy_FraigSavePattern2(p *Ivy_FraigMan_t) { Ivy_FraigSavePattern(p) }
func Ivy_FraigSavePattern3(p *Ivy_FraigMan_t) {
	for i := range p.pPatWords {
		p.pPatWords[i] = p.rng.Uint32()
	}
	for _, n := range p.vPiVars {
		value, known := p.pSat.Value(Ivy_ObjSatNum(n))
		if known {
			index := int(n.Id - 1)
			mask := uint32(1) << uint(index%32)
			if value {
				p.pPatWords[index/32] |= mask
			} else {
				p.pPatWords[index/32] &^= mask
			}
		}
	}
}
func ivyFraigSolve(p *Ivy_FraigMan_t, assumptions []bsat.Literal, limit int32) bsat.Status {
	p.pSat.SetLimits(int64(max(limit, 0)), 0)
	start := time.Now()
	result := p.pSat.SolveWithGlobalLimits(assumptions, p.nBTLimitGlobal, p.nInsLimitGlobal)
	elapsed := time.Since(start)
	p.timeSat += elapsed
	switch result {
	case bsat.Satisfiable:
		p.timeSatSat += elapsed
		p.nSatCallsSat++
		if p.pPatWords != nil {
			Ivy_FraigSavePattern(p)
		}
	case bsat.Unsatisfiable:
		p.timeSatUnsat += elapsed
		p.nSatCallsUnsat++
		clause := slices.Clone(assumptions)
		for i := range clause {
			clause[i] = ivyFraigNeg(clause[i])
		}
		p.pSat.AddClause(clause...)
	default:
		p.timeSatFail += elapsed
		p.nSatFailsReal++
	}
	return result
}
func Ivy_FraigNodesAreEquiv(p *Ivy_FraigMan_t, old, next *Ivy_Obj_t) int32 {
	if old == nil || next == nil || Ivy_IsComplement(old) != 0 || Ivy_IsComplement(next) != 0 {
		panic("ivy: regular equivalence candidates required")
	}
	if old == next {
		return 1
	}
	limit := p.pParams.NBTLimitNode
	if limit > 0 && (old.fFailTfo != 0 || next.fFailTfo != 0) {
		p.nSatFails++
		if limit <= 10 {
			return -1
		}
		limit = int32(math.Pow(float64(limit), 0.7))
	}
	p.nSatCalls++
	Ivy_FraigNodeAddToSolver(p, old, next)
	Ivy_FraigSetActivityFactors(p, old, next)
	a, b := ivyFraigLiteral(old), ivyFraigLiteral(next)
	firstB := b
	if old.fPhase == next.fPhase {
		firstB = ivyFraigNeg(b)
	}
	for _, assumptions := range [][]bsat.Literal{{a, firstB}, {ivyFraigNeg(a), ivyFraigNeg(firstB)}} {
		result := ivyFraigSolve(p, assumptions, limit)
		if result == bsat.Satisfiable {
			return 0
		}
		if result != bsat.Unsatisfiable {
			if old != p.pManFraig.pConst1 {
				old.fFailTfo = 1
			}
			next.fFailTfo = 1
			return -1
		}
		if old == p.pManFraig.pConst1 {
			break
		}
	}
	p.nSatProof++
	return 1
}
func Ivy_FraigNodeIsConst(p *Ivy_FraigMan_t, n *Ivy_Obj_t) int32 {
	p.nSatCalls++
	Ivy_FraigNodeAddToSolver(p, nil, n)
	Ivy_FraigSetActivityFactors(p, nil, n)
	literal := ivyFraigLiteral(n)
	if n.fPhase != 0 {
		literal = ivyFraigNeg(literal)
	}
	result := ivyFraigSolve(p, []bsat.Literal{literal}, p.pParams.NBTLimitMiter)
	if result == bsat.Unsatisfiable {
		p.nSatProof++
		return 1
	}
	if result == bsat.Satisfiable {
		return 0
	}
	n.fFailTfo = 1
	return -1
}
