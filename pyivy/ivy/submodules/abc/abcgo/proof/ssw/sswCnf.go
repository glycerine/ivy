package ssw

// Incremental SAT encoding for SSW, ported from sswCnf.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Ssw_SatStart(fPolarFlip int32) *Ssw_Sat_t {
	p := &Ssw_Sat_t{fPolarFlip: fPolarFlip, vSatVars: vec.Vec_IntStart(10000), vFanins: vec.Vec_PtrAlloc(100), vUsedPis: vec.Vec_PtrAlloc(100), pSat: bsat.NewSolver(), nSatVars: 2}
	lit := int32(2)
	if fPolarFlip != 0 {
		lit ^= 1
	}
	if !p.pSat.AddClause(lit) {
		panic("ssw: cannot initialize constant SAT variable")
	}
	vec.Vec_IntWriteEntry(p.vSatVars, 0, 1)
	return p
}
func Ssw_SatStop(p *Ssw_Sat_t) {
	if p != nil {
		*p = Ssw_Sat_t{}
	}
}

// sswLit translates an edge's Boolean value to the solver's possibly phase-flipped polarity.
func sswLit(p *Ssw_Sat_t, e aig.Edge, neg bool) int32 {
	regular := aig.Aig_Regular(e)
	lit := Ssw_ObjSatNum(p, regular) << 1
	if e.Complemented() != neg {
		lit ^= 1
	}
	if p.fPolarFlip != 0 && aig.Aig_ObjPhase(p.pAig, regular) {
		lit ^= 1
	}
	return lit
}
func sswAddClause(p *Ssw_Sat_t, lits ...int32) {
	if !p.pSat.AddClause(lits...) {
		panic("ssw: inconsistent CNF encoding")
	}
}
func Ssw_AddClausesMux(p *Ssw_Sat_t, node aig.Edge) {
	control, yes, no := aig.Aig_ObjRecognizeMux(p.pAig, node)
	sswAddClause(p, sswLit(p, control, true), sswLit(p, yes, true), sswLit(p, node, false))
	sswAddClause(p, sswLit(p, control, true), sswLit(p, yes, false), sswLit(p, node, true))
	sswAddClause(p, sswLit(p, control, false), sswLit(p, no, true), sswLit(p, node, false))
	sswAddClause(p, sswLit(p, control, false), sswLit(p, no, false), sswLit(p, node, true))
	if yes.NodeID() != no.NodeID() {
		sswAddClause(p, sswLit(p, yes, false), sswLit(p, no, false), sswLit(p, node, true))
		sswAddClause(p, sswLit(p, yes, true), sswLit(p, no, true), sswLit(p, node, false))
	}
}
func Ssw_AddClausesSuper(p *Ssw_Sat_t, node aig.Edge, super *vec.Vec_Ptr_t) {
	var large []int32
	for _, item := range vec.Vec_PtrArray(super) {
		fanin := item.(aig.Edge)
		sswAddClause(p, sswLit(p, fanin, false), sswLit(p, node, true))
		large = append(large, sswLit(p, fanin, true))
	}
	large = append(large, sswLit(p, node, false))
	sswAddClause(p, large...)
}
func Ssw_CollectSuper_rec(g *aig.Graph, obj aig.Edge, super *vec.Vec_Ptr_t, first, useMuxes int32) {
	if obj.Complemented() || aig.Aig_ObjIsCi(g, obj) || (first == 0 && aig.Aig_ObjRefs(g, obj) > 1) || (useMuxes != 0 && aig.Aig_ObjIsMuxType(g, obj)) || g.Node(obj).Kind != aig.And {
		vec.Vec_PtrPushUnique(super, obj)
		return
	}
	Ssw_CollectSuper_rec(g, aig.Aig_ObjChild0(g, obj), super, 0, useMuxes)
	Ssw_CollectSuper_rec(g, aig.Aig_ObjChild1(g, obj), super, 0, useMuxes)
}
func Ssw_CollectSuper(g *aig.Graph, obj aig.Edge, useMuxes int32, super *vec.Vec_Ptr_t) {
	vec.Vec_PtrClear(super)
	Ssw_CollectSuper_rec(g, obj, super, 1, useMuxes)
}
func Ssw_ObjAddToFrontier(p *Ssw_Sat_t, obj aig.Edge, frontier *vec.Vec_Ptr_t) {
	if obj.Complemented() {
		panic("ssw: complemented frontier object")
	}
	if Ssw_ObjSatNum(p, obj) != 0 {
		return
	}
	if aig.Aig_ObjIsConst1(p.pAig, obj) {
		return
	}
	if aig.Aig_ObjIsCi(p.pAig, obj) {
		vec.Vec_PtrPush(p.vUsedPis, obj)
	}
	Ssw_ObjSetSatNum(p, obj, p.nSatVars)
	p.nSatVars++
	for p.pSat.NumVars() < p.nSatVars {
		p.pSat.NewVariable()
	}
	if aig.Aig_ObjIsNode(p.pAig, obj) {
		vec.Vec_PtrPush(frontier, obj)
	}
}
func Ssw_CnfNodeAddToSolver(p *Ssw_Sat_t, obj aig.Edge) {
	if p.pAig == nil || obj.Complemented() {
		panic("ssw: invalid CNF graph or object")
	}
	if Ssw_ObjSatNum(p, obj) != 0 {
		return
	}
	frontier := vec.Vec_PtrAlloc(100)
	Ssw_ObjAddToFrontier(p, obj, frontier)
	for i := 0; i < vec.Vec_PtrSize(frontier); i++ {
		node := vec.Vec_PtrEntry(frontier, i).(aig.Edge)
		if p.pAig.Node(node).Kind == aig.ExclusiveOr {
			a, b := aig.Aig_ObjChild0(p.pAig, node), aig.Aig_ObjChild1(p.pAig, node)
			Ssw_ObjAddToFrontier(p, aig.Aig_Regular(a), frontier)
			Ssw_ObjAddToFrontier(p, aig.Aig_Regular(b), frontier)
			sswAddClause(p, sswLit(p, a, false), sswLit(p, b, false), sswLit(p, node, true))
			sswAddClause(p, sswLit(p, a, true), sswLit(p, b, true), sswLit(p, node, true))
			sswAddClause(p, sswLit(p, a, false), sswLit(p, b, true), sswLit(p, node, false))
			sswAddClause(p, sswLit(p, a, true), sswLit(p, b, false), sswLit(p, node, false))
			continue
		}
		if aig.Aig_ObjIsMuxType(p.pAig, node) {
			c, y, n := aig.Aig_ObjRecognizeMux(p.pAig, node)
			vec.Vec_PtrClear(p.vFanins)
			for _, e := range []aig.Edge{c, y, n} {
				vec.Vec_PtrPushUnique(p.vFanins, aig.Aig_Regular(e))
			}
			for _, item := range vec.Vec_PtrArray(p.vFanins) {
				Ssw_ObjAddToFrontier(p, item.(aig.Edge), frontier)
			}
			Ssw_AddClausesMux(p, node)
		} else {
			Ssw_CollectSuper(p.pAig, node, 1, p.vFanins)
			for _, item := range vec.Vec_PtrArray(p.vFanins) {
				Ssw_ObjAddToFrontier(p, aig.Aig_Regular(item.(aig.Edge)), frontier)
			}
			Ssw_AddClausesSuper(p, node, p.vFanins)
		}
	}
}
func Ssw_CnfGetNodeValue(p *Ssw_Sat_t, obj aig.Edge) int32 {
	if obj.Complemented() {
		panic("ssw: complemented model object")
	}
	if v := Ssw_ObjSatNum(p, obj); v > 0 {
		value, ok := p.pSat.Value(v)
		if !ok {
			panic("ssw: SAT model unavailable")
		}
		if value {
			return 1
		}
		return 0
	}
	if aig.Aig_ObjIsCi(p.pAig, obj) {
		return 0
	}
	if !aig.Aig_ObjIsNode(p.pAig, obj) {
		panic("ssw: unsupported model object")
	}
	a, b := aig.Aig_ObjChild0(p.pAig, obj), aig.Aig_ObjChild1(p.pAig, obj)
	x, y := Ssw_CnfGetNodeValue(p, aig.Aig_Regular(a)), Ssw_CnfGetNodeValue(p, aig.Aig_Regular(b))
	if a.Complemented() {
		x ^= 1
	}
	if b.Complemented() {
		y ^= 1
	}
	if p.pAig.Node(obj).Kind == aig.ExclusiveOr {
		return x ^ y
	}
	return x & y
}
