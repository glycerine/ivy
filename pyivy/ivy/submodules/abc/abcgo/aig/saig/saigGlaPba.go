package saig

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Aig_Gla2Man_t owns the bounded unrolling used by gate-level localization
// abstraction. Clause IDs are the proof solver's stable, one-based root IDs.
type Aig_Gla2Man_t struct {
	pAig                         *Aig_Man_t
	nStart, nFramesMax, fVerbose int32
	objVars                      []int32
	claObj, claFrame             []int32
	pSat                         *bsat.Solver2
	clauseConflict               bool
}

func newGlaManager(g *Aig_Man_t, start, frames, verbose int32) *Aig_Gla2Man_t {
	return &Aig_Gla2Man_t{pAig: g, nStart: start, nFramesMax: frames, fVerbose: verbose}
}

func (p *Aig_Gla2Man_t) variable(edge aig.Edge, frame int32) int32 {
	id := int(aig.Aig_ObjId(aig.Aig_Regular(edge)))
	if id < 0 || id >= len(p.objVars) || frame < 0 || frame >= p.nFramesMax {
		panic("saig GLA: invalid object/frame")
	}
	v := p.objVars[id*int(p.nFramesMax)+int(frame)]
	if v < 0 {
		panic("saig GLA: variable was not assigned during unrolling")
	}
	return v
}

func (p *Aig_Gla2Man_t) addClause(object, frame int32, literals ...bsat.Literal) bool {
	if p.clauseConflict {
		return false
	}
	ok := p.pSat.AddClause(literals...)
	// IDs advance for every root clause, including an immediately conflicting
	// one; the solver wrapper assigns IDs before detecting that conflict.
	p.claObj = append(p.claObj, object)
	p.claFrame = append(p.claFrame, frame)
	p.clauseConflict = !ok
	return ok
}

func glaLit(variable int32, complement bool) bsat.Literal {
	lit := bsat.Positive(variable)
	if complement {
		lit = bsat.Negative(variable)
	}
	return lit
}

func (p *Aig_Gla2Man_t) assign(edge aig.Edge, frame int32) {
	g := p.pAig
	e := aig.Aig_Regular(edge)
	id := int(aig.Aig_ObjId(e))
	index := id*int(p.nFramesMax) + int(frame)
	if p.objVars[index] >= 0 {
		return
	}
	p.objVars[index] = p.pSat.NewVariable()
	if aig.Aig_ObjIsConst1(g, e) {
		return
	}
	if Saig_ObjIsPo(g, e) {
		p.assign(aig.Aig_ObjFanin0(g, e), frame)
		return
	}
	if aig.Aig_ObjIsCi(g, e) {
		if Saig_ObjIsLo(g, e) && frame > 0 {
			p.assign(aig.Aig_ObjFanin0(g, Saig_ObjLoToLi(g, e)), frame-1)
		}
		return
	}
	if aig.Aig_ObjIsNode(g, e) {
		p.assign(aig.Aig_ObjFanin0(g, e), frame)
		p.assign(aig.Aig_ObjFanin1(g, e), frame)
	}
}

func (p *Aig_Gla2Man_t) build() bool {
	if p == nil || p.pAig == nil || p.nFramesMax <= 0 || p.nStart < 0 || p.nStart >= p.nFramesMax || Saig_ManPoNum(p.pAig) != 1 {
		return false
	}
	maxObjects := int(aig.Aig_ManObjNumMax(p.pAig))
	if maxObjects > int(^uint(0)>>1)/int(p.nFramesMax) {
		return false
	}
	p.objVars = make([]int32, maxObjects*int(p.nFramesMax))
	for i := range p.objVars {
		p.objVars[i] = -1
	}
	p.pSat = bsat.NewSolver2()
	for f := p.nFramesMax - 1; f >= 0; f-- {
		p.assign(aig.Aig_ManCo(p.pAig, 0), f)
	}
	add := func(obj aig.Edge, frame int32, lits ...bsat.Literal) bool {
		return p.addClause(aig.Aig_ObjId(obj), frame, lits...)
	}
	ok := true
	for id := 0; id < maxObjects; id++ {
		for frame := 0; frame < int(p.nFramesMax); frame++ {
			variable := p.objVars[id*int(p.nFramesMax)+frame]
			if variable < 0 {
				continue
			}
			obj := aig.Aig_ManObj(p.pAig, int32(id))
			f := int32(frame)
			switch {
			case aig.Aig_ObjIsNode(p.pAig, obj):
				a := p.variable(aig.Aig_ObjFanin0(p.pAig, obj), f)
				b := p.variable(aig.Aig_ObjFanin1(p.pAig, obj), f)
				c0 := aig.Aig_ObjFaninC0(p.pAig, obj) != 0
				c1 := aig.Aig_ObjFaninC1(p.pAig, obj) != 0
				ok = add(obj, f, glaLit(variable, true), glaLit(a, c0)) && ok
				ok = add(obj, f, glaLit(variable, true), glaLit(b, c1)) && ok
				ok = add(obj, f, glaLit(variable, false), glaLit(a, !c0), glaLit(b, !c1)) && ok
			case Saig_ObjIsLo(p.pAig, obj):
				if f == 0 {
					ok = add(obj, f, glaLit(variable, true)) && ok
				} else {
					li := Saig_ObjLoToLi(p.pAig, obj)
					driver := p.variable(aig.Aig_ObjFanin0(p.pAig, li), f-1)
					c := aig.Aig_ObjFaninC0(p.pAig, li) != 0
					ok = add(obj, f, glaLit(variable, false), glaLit(driver, !c)) && ok
					ok = add(obj, f, glaLit(variable, true), glaLit(driver, c)) && ok
				}
			case Saig_ObjIsPo(p.pAig, obj):
				driver := p.variable(aig.Aig_ObjFanin0(p.pAig, obj), f)
				c := aig.Aig_ObjFaninC0(p.pAig, obj) != 0
				ok = add(obj, f, glaLit(variable, false), glaLit(driver, !c)) && ok
				ok = add(obj, f, glaLit(variable, true), glaLit(driver, c)) && ok
			case aig.Aig_ObjIsConst1(p.pAig, obj):
				ok = add(obj, f, glaLit(variable, false)) && ok
			}
		}
	}
	property := make([]bsat.Literal, 0, p.nFramesMax-p.nStart)
	for f := p.nStart; f < p.nFramesMax; f++ {
		property = append(property, glaLit(p.variable(aig.Aig_ManCo(p.pAig, 0), f), false))
	}
	ok = p.addClause(0, 0, property...) && ok
	return ok
}

// Aig_Gla2ManStart allocates an abstraction manager; no SAT work is done yet.
func Aig_Gla2ManStart(g *Aig_Man_t, start, frames, verbose int32) *Aig_Gla2Man_t {
	return newGlaManager(g, start, frames, verbose)
}

func Aig_Gla2FetchVar(p *Aig_Gla2Man_t, obj *Aig_Obj_t, frame int32) int32 {
	if p == nil || obj == nil {
		panic("saig GLA: nil manager or object")
	}
	return p.variable(*obj, frame)
}

func Aig_Gla2AssignVars_rec(p *Aig_Gla2Man_t, obj *Aig_Obj_t, frame int32) {
	if p == nil || obj == nil {
		panic("saig GLA: nil manager or object")
	}
	p.assign(*obj, frame)
}

func Aig_Gla2CreateSatSolver(p *Aig_Gla2Man_t) int32 {
	if p == nil {
		return 0
	}
	return boolToInt(p.build())
}

func Aig_Gla2ManStop(p *Aig_Gla2Man_t) {
	if p != nil {
		*p = Aig_Gla2Man_t{}
	}
}

// Saig_AbsSolverUnsatCore solves the bounded problem and returns the proof's
// root clause IDs. The return status uses 1 for UNSAT, 0 for SAT, and -1 for
// resource exhaustion, matching the C helper's convention.
func Saig_AbsSolverUnsatCore(s *bsat.Solver2, conflicts, verbose int32, status *int32) *Vec_Int_t {
	if status != nil {
		*status = -1
	}
	if s == nil {
		return nil
	}
	if conflicts > 0 {
		s.SetLimits(int64(conflicts), 0)
	}
	result := s.Solve()
	if result == bsat.Satisfiable {
		if status != nil {
			*status = 0
		}
		return nil
	}
	if result != bsat.Unsatisfiable {
		return nil
	}
	if status != nil {
		*status = 1
	}
	ids := s.ProofCore()
	core := vec.Vec_IntAlloc(len(ids))
	vec.Vec_IntPushArray(core, ids, len(ids))
	return core
}

// Aig_Gla2ManCollect maps proof-root clause IDs back to the graph objects they
// constrain. The result is a dense object bitmap and always includes const1.
func Aig_Gla2ManCollect(p *Aig_Gla2Man_t, core *Vec_Int_t) *Vec_Int_t {
	if p == nil || p.pAig == nil {
		return nil
	}
	result := vec.Vec_IntStart(int(aig.Aig_ManObjNumMax(p.pAig)))
	vec.Vec_IntWriteEntry(result, 0, 1)
	for _, raw := range vec.Vec_IntArray(core) {
		id := int(raw) - 1
		if id < 0 || id >= len(p.claObj) {
			continue
		}
		obj := aig.Aig_ManObj(p.pAig, p.claObj[id])
		if Saig_ObjIsPi(p.pAig, obj) || Saig_ObjIsPo(p.pAig, obj) || aig.Aig_ObjIsConst1(p.pAig, obj) {
			continue
		}
		vec.Vec_IntWriteEntry(result, int(p.claObj[id]), 1)
	}
	return result
}

// Aig_Gla2ManPerform unrolls the property, solves it, and returns the bitmap
// of objects participating in the proof core. nil means SAT, unknown, or bad input.
func Aig_Gla2ManPerform(g *Aig_Man_t, start, frames, conflicts, seconds, skipRandom, verbose int32) *Vec_Int_t {
	p := Aig_Gla2ManStart(g, start, frames, verbose)
	defer Aig_Gla2ManStop(p)
	if !p.build() {
		return nil
	}
	if conflicts > 0 {
		p.pSat.SetLimits(int64(conflicts), 0)
	}
	if seconds > 0 {
		p.pSat.SetDeadline(time.Now().Add(time.Duration(seconds) * time.Second))
	}
	status := p.pSat.Solve()
	if status != bsat.Unsatisfiable {
		return nil
	}
	ids := p.pSat.ProofCore()
	core := vec.Vec_IntAlloc(len(ids))
	vec.Vec_IntPushArray(core, ids, len(ids))
	return Aig_Gla2ManCollect(p, core)
}

// Gla2AddConst, Gla2AddBuffer, and Gla2AddNode expose the CNF encodings used
// by the unroller for callers that build compatible proof solvers themselves.
func Aig_Gla2AddConst(s *bsat.Solver2, v int32, complement int32) int32 {
	if complement != 0 {
		return boolToInt(s.AddClause(glaLit(v, true)))
	}
	return boolToInt(s.AddClause(glaLit(v, false)))
}
func Aig_Gla2AddBuffer(s *bsat.Solver2, a, b int32, c int32) int32 {
	ok := s.AddClause(glaLit(a, false), glaLit(b, c == 0))
	ok = s.AddClause(glaLit(a, true), glaLit(b, c != 0)) && ok
	return boolToInt(ok)
}
func Aig_Gla2AddNode(s *bsat.Solver2, z, a, b int32, ca, cb int32) int32 {
	ok := s.AddClause(glaLit(z, true), glaLit(a, ca != 0))
	ok = s.AddClause(glaLit(z, true), glaLit(b, cb != 0)) && ok
	ok = s.AddClause(glaLit(z, false), glaLit(a, ca == 0), glaLit(b, cb == 0)) && ok
	return boolToInt(ok)
}
func boolToInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
