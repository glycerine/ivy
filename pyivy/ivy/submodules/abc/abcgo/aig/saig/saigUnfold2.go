package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ssw_ManFindDirectImplications2 discovers and filters type-I/II candidates.
func Ssw_ManFindDirectImplications2(g *aig.Graph, frames, conflicts, propagations, verbose int32) *vec.Vec_Vec_t {
	candidates, _, err := discoverUnfold2Candidates(g, frames, conflicts, propagations, verbose)
	if err != nil {
		panic(err)
	}
	return candidates
}

func unfoldConstraints2(g *aig.Graph, classes unfold2Classes) *aig.Graph {
	if g == nil {
		return nil
	}
	if len(classes.typeI)+len(classes.typeII) == 0 {
		return aig.Aig_ManDupSimple(g)
	}
	q := aig.NewGraph()
	q.Name, q.Spec = g.Name, g.Spec
	copies := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
	for _, e := range g.Inputs() {
		copies[e] = q.NewInput()
	}
	newRegs := make([]aig.Edge, len(classes.typeII))
	for i := range newRegs {
		newRegs[i] = q.NewInput()
	}
	for i := int32(0); i < Saig_ManPoNum(g); i++ {
		q.AddOutput(temporCopy(q, g, copies, aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, i))))
	}
	for _, e := range classes.typeI {
		q.AddOutput(temporCopy(q, g, copies, e))
	}
	for i, e := range classes.typeII {
		q.AddOutput(q.AND(newRegs[i], temporCopy(q, g, copies, e)))
	}
	for i := int32(0); i < Saig_ManRegNum(g); i++ {
		q.AddOutput(temporCopy(q, g, copies, aig.Aig_ObjChild0(g, Saig_ManLi(g, i))))
	}
	for _, e := range classes.typeII {
		q.AddOutput(temporCopy(q, g, copies, e))
	}
	aig.Aig_ManSetRegNum(q, Saig_ManRegNum(g)+int32(len(newRegs)))
	aig.Aig_ManSetConstrNum(q, aig.Aig_ManConstrNum(g)+int32(len(classes.typeI)+len(classes.typeII)))
	return q
}

func Saig_ManDupUnfoldConstrsFunc2(g *aig.Graph, frames, conflicts, propagations, oldAlgorithm, verbose int32, typeIICount *int32) *aig.Graph {
	if typeIICount != nil {
		*typeIICount = 0
	}
	if g == nil {
		return nil
	}
	var classes unfold2Classes
	if oldAlgorithm != 0 {
		groups := Saig_ManDetectConstrFunc(g, frames, conflicts, propagations, verbose)
		candidates := vec.Vec_VecAlloc(len(groups))
		for level, group := range groups {
			for _, e := range group {
				vec.Vec_VecPush(candidates, level, e)
			}
		}
		var err error
		classes, err = filterUnfold2Candidates(g, candidates, conflicts, propagations, verbose)
		if err != nil {
			panic(err)
		}
	} else {
		var err error
		_, classes, err = discoverUnfold2Candidates(g, frames, conflicts, propagations, verbose)
		if err != nil {
			panic(err)
		}
	}
	if typeIICount != nil {
		*typeIICount = int32(len(classes.typeII))
	}
	return unfoldConstraints2(g, classes)
}

// Saig_ManDupFoldConstrsFunc2 folds violations into a sticky latch and freezes
// the existing registers after a violation, matching the original routine.
func Saig_ManDupFoldConstrsFunc2(g *aig.Graph, complement, verbose, typeIICount int32) *aig.Graph {
	if g == nil {
		return nil
	}
	constraints := aig.Aig_ManConstrNum(g)
	if constraints == 0 {
		return aig.Aig_ManDupSimple(g)
	}
	pos, regs := Saig_ManPoNum(g), Saig_ManRegNum(g)
	if constraints >= pos {
		panic("saig: no property outputs remain")
	}
	q := aig.NewGraph()
	q.Name, q.Spec = g.Name, g.Spec
	copies := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
	for _, e := range g.Inputs() {
		copies[e] = q.NewInput()
	}
	violation := q.Const0()
	for i := pos - constraints; i < pos; i++ {
		e := temporCopy(q, g, copies, aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, i)))
		if complement != 0 {
			e = e.Not()
		}
		violation = q.OR(violation, e)
	}
	if regs > 0 {
		violation = q.OR(violation, q.NewInput())
	}
	for i := int32(0); i < pos-constraints; i++ {
		e := temporCopy(q, g, copies, aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, i)))
		q.AddOutput(q.AND(e, violation.Not()))
	}
	for i := int32(0); i < regs; i++ {
		li, lo := Saig_ManLi(g, i), Saig_ManLo(g, i)
		next := temporCopy(q, g, copies, aig.Aig_ObjChild0(g, li))
		q.AddOutput(q.Mux(violation.Not(), next, copies[lo]))
	}
	if regs > 0 {
		q.AddOutput(violation)
		regs++
	}
	aig.Aig_ManSetRegNum(q, regs)
	aig.Aig_ManCleanup(q)
	aig.Aig_ManSeqCleanup(q)
	_ = verbose
	_ = typeIICount
	return q
}
