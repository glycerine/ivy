package saig

import (
	"time"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Aig_Gla3Man_t uses the proof-logging SAT backend for the same gate-level
// abstraction encoding as Gla2. The engine is shared to keep frame semantics
// and clause-to-object mapping identical across both entry points.
type Aig_Gla3Man_t struct{ manager *Aig_Gla2Man_t }

func Aig_Gla3ManStart(g *Aig_Man_t, start, frames, verbose int32) *Aig_Gla3Man_t {
	return &Aig_Gla3Man_t{manager: newGlaManager(g, start, frames, verbose)}
}
func Aig_Gla3ManStop(p *Aig_Gla3Man_t) {
	if p != nil {
		Aig_Gla2ManStop(p.manager)
		p.manager = nil
	}
}
func Aig_Gla3FetchVar(p *Aig_Gla3Man_t, obj *Aig_Obj_t, frame int32) int32 {
	return p.manager.variable(*obj, frame)
}
func Aig_Gla3AssignVars_rec(p *Aig_Gla3Man_t, obj *Aig_Obj_t, frame int32) {
	p.manager.assign(*obj, frame)
}
func Aig_Gla3CreateSatSolver(p *Aig_Gla3Man_t) int32 {
	if p == nil || p.manager == nil {
		return 0
	}
	// The original Gla3 builder leaves an immediately inconsistent formula in
	// the proof solver and lets Solve produce its core. Gla2 instead reports the
	// root conflict while constructing the clauses.
	p.manager.build()
	return 1
}
func Aig_Gla3AddConst(s *bsat.Solver2, v, compl int32) int32 { return Aig_Gla2AddConst(s, v, compl) }
func Aig_Gla3AddBuffer(s *bsat.Solver2, a, b, compl int32) int32 {
	return Aig_Gla2AddBuffer(s, a, b, compl)
}
func Aig_Gla3AddNode(s *bsat.Solver2, z, a, b, ca, cb int32) int32 {
	return Aig_Gla2AddNode(s, z, a, b, ca, cb)
}

// Aig_Gla3ManUnsatCore returns proof root IDs. A nil result means SAT or
// resource exhaustion; the optional status is 1 for UNSAT, 0 for SAT, -1 unknown.
func Aig_Gla3ManUnsatCore(s *bsat.Solver2, conflicts, verbose int32, status *int32) *Vec_Int_t {
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
	v := vec.Vec_IntAlloc(len(ids))
	vec.Vec_IntPushArray(v, ids, len(ids))
	return v
}
func Aig_Gla3ManCollect(p *Aig_Gla3Man_t, core *Vec_Int_t) *Vec_Int_t {
	if p == nil {
		return nil
	}
	return Aig_Gla2ManCollect(p.manager, core)
}
func Aig_Gla3ManPerform(g *Aig_Man_t, start, frames, conflicts, seconds, skipRandom, verbose int32) *Vec_Int_t {
	p := Aig_Gla3ManStart(g, start, frames, verbose)
	defer Aig_Gla3ManStop(p)
	if Aig_Gla3CreateSatSolver(p) == 0 {
		return nil
	}
	if seconds > 0 {
		p.manager.pSat.SetDeadline(time.Now().Add(time.Duration(seconds) * time.Second))
	}
	core := Aig_Gla3ManUnsatCore(p.manager.pSat, conflicts, verbose, nil)
	if core == nil {
		return nil
	}
	return Aig_Gla3ManCollect(p, core)
}
