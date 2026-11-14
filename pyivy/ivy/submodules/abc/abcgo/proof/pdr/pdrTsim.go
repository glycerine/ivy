// Ternary simulation and cube generalization, ported from ABC pdrTsim.c.
package pdr

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
)

const (
	PDR_ZER int32 = 1
	PDR_ONE int32 = 2
	PDR_UND int32 = 3
)

func Pdr_ManSimInfoNot(v int32) int32 {
	if v == PDR_ZER {
		return PDR_ONE
	}
	if v == PDR_ONE {
		return PDR_ZER
	}
	return PDR_UND
}
func Pdr_ManSimInfoAnd(a, b int32) int32 {
	if a == PDR_ZER || b == PDR_ZER {
		return PDR_ZER
	}
	if a == PDR_ONE && b == PDR_ONE {
		return PDR_ONE
	}
	return PDR_UND
}

// Like ABC's pTerSimData, Xsim is temporary per-object simulation storage.
// Other ternary simulations must not be interleaved on the same graph.
func Pdr_ManSimInfoGet(g *aig.Graph, e aig.Edge) int32    { return aig.Aig_ObjGetXsim(g, e) }
func Pdr_ManSimInfoSet(g *aig.Graph, e aig.Edge, v int32) { aig.Aig_ObjSetXsim(g, e, v) }

func Pdr_ManCollectCone_rec(g *aig.Graph, e aig.Edge, cis, nodes *vec.Vec_Int_t) {
	if aig.Aig_ObjIsTravIdCurrent(g, e) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(g, e)
	if aig.Aig_ObjIsCi(g, e) {
		vec.Vec_IntPush(cis, int32(e.NodeID()))
		return
	}
	Pdr_ManCollectCone_rec(g, aig.Aig_ObjFanin0(g, e), cis, nodes)
	if aig.Aig_ObjIsCo(g, e) {
		return
	}
	if g.Node(e).Kind == aig.Buffer {
		vec.Vec_IntPush(nodes, int32(e.NodeID()))
		return
	}
	Pdr_ManCollectCone_rec(g, aig.Aig_ObjFanin1(g, e), cis, nodes)
	vec.Vec_IntPush(nodes, int32(e.NodeID()))
}
func Pdr_ManCollectCone(g *aig.Graph, cos, cis, nodes *vec.Vec_Int_t) {
	vec.Vec_IntClear(cis)
	vec.Vec_IntClear(nodes)
	aig.Aig_ManIncrementTravId(g)
	aig.Aig_ObjSetTravIdCurrent(g, g.Const1())
	for _, id := range vec.Vec_IntArray(cos) {
		Pdr_ManCollectCone_rec(g, aig.Aig_ManObj(g, id), cis, nodes)
	}
}
func Pdr_ManExtendOneEval(g *aig.Graph, e aig.Edge) int32 {
	a := Pdr_ManSimInfoGet(g, aig.Aig_ObjFanin0(g, e))
	if aig.Aig_ObjFaninC0(g, e) != 0 {
		a = Pdr_ManSimInfoNot(a)
	}
	if aig.Aig_ObjIsCo(g, e) || g.Node(e).Kind == aig.Buffer {
		Pdr_ManSimInfoSet(g, e, a)
		return a
	}
	b := Pdr_ManSimInfoGet(g, aig.Aig_ObjFanin1(g, e))
	if aig.Aig_ObjFaninC1(g, e) != 0 {
		b = Pdr_ManSimInfoNot(b)
	}
	v := Pdr_ManSimInfoAnd(a, b)
	if g.Node(e).Kind == aig.ExclusiveOr {
		if a == PDR_UND || b == PDR_UND {
			v = PDR_UND
		} else if a == b {
			v = PDR_ZER
		} else {
			v = PDR_ONE
		}
	}
	Pdr_ManSimInfoSet(g, e, v)
	return v
}
func Pdr_ManSimDataInit(g *aig.Graph, cis, vals, nodes, cos, outvals, removed *vec.Vec_Int_t) int32 {
	Pdr_ManSimInfoSet(g, g.Const1(), PDR_ONE)
	for i, id := range vec.Vec_IntArray(cis) {
		v := PDR_ZER
		if vec.Vec_IntEntry(vals, i) != 0 {
			v = PDR_ONE
		}
		Pdr_ManSimInfoSet(g, aig.Aig_ManObj(g, id), v)
	}
	if removed != nil {
		for _, id := range vec.Vec_IntArray(removed) {
			Pdr_ManSimInfoSet(g, aig.Aig_ManObj(g, id), PDR_UND)
		}
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		Pdr_ManExtendOneEval(g, aig.Aig_ManObj(g, id))
	}
	for _, id := range vec.Vec_IntArray(cos) {
		Pdr_ManExtendOneEval(g, aig.Aig_ManObj(g, id))
	}
	for i, id := range vec.Vec_IntArray(cos) {
		v := Pdr_ManSimInfoGet(g, aig.Aig_ManObj(g, id))
		expected := PDR_ZER
		if vec.Vec_IntEntry(outvals, i) != 0 {
			expected = PDR_ONE
		}
		if v != expected {
			return 0
		}
	}
	return 1
}
func Pdr_ManExtendOne(g *aig.Graph, e aig.Edge, undo, vis *vec.Vec_Int_t) int32 {
	if !saig.Saig_ObjIsLo(g, e) || !aig.Aig_ObjIsTravIdCurrent(g, e) {
		panic("pdr: expected a marked latch output")
	}
	old := Pdr_ManSimInfoGet(g, e)
	if old != PDR_ZER && old != PDR_ONE {
		panic("pdr: expected known latch value")
	}
	vec.Vec_IntPush(undo, int32(e.NodeID()))
	vec.Vec_IntPush(undo, old)
	Pdr_ManSimInfoSet(g, e, PDR_UND)
	vec.Vec_IntClear(vis)
	vec.Vec_IntPush(vis, int32(e.NodeID()))
	for i := 0; i < vec.Vec_IntSize(vis); i++ {
		id := vec.Vec_IntEntry(vis, i)
		for k, handle := int32(0), aig.Aig_ObjFanout0Int(g, id); k < aig.Aig_ObjRefs(g, aig.Aig_ManObj(g, id)); k, handle = k+1, aig.Aig_ObjFanoutNext(g, handle) {
			f := aig.Aig_ManObj(g, handle/2)
			if !aig.Aig_ObjIsTravIdCurrent(g, f) {
				continue
			}
			old = Pdr_ManSimInfoGet(g, f)
			if old == PDR_UND {
				continue
			}
			now := Pdr_ManExtendOneEval(g, f)
			if now == old {
				continue
			}
			if now != PDR_UND {
				panic("pdr: non-monotone ternary propagation")
			}
			vec.Vec_IntPush(undo, int32(f.NodeID()))
			vec.Vec_IntPush(undo, old)
			if aig.Aig_ObjIsCo(g, f) {
				return 0
			}
			vec.Vec_IntPushOrder(vis, int32(f.NodeID()))
		}
	}
	return 1
}
func Pdr_ManExtendUndo(g *aig.Graph, undo *vec.Vec_Int_t) {
	a := vec.Vec_IntArray(undo)
	for i := 0; i < len(a); i += 2 {
		e := aig.Aig_ManObj(g, a[i])
		if Pdr_ManSimInfoGet(g, e) != PDR_UND {
			panic("pdr: invalid undo")
		}
		Pdr_ManSimInfoSet(g, e, a[i+1])
	}
}
func Pdr_ManDeriveResult(g *aig.Graph, cis, vals, removed, res, pis *vec.Vec_Int_t) {
	aig.Aig_ManIncrementTravId(g)
	for _, id := range vec.Vec_IntArray(removed) {
		aig.Aig_ObjSetTravIdCurrent(g, aig.Aig_ManObj(g, id))
	}
	vec.Vec_IntClear(res)
	vec.Vec_IntClear(pis)
	for i, id := range vec.Vec_IntArray(cis) {
		e := aig.Aig_ManObj(g, id)
		lit := 2 * aig.Aig_ObjCioId(g, e)
		if vec.Vec_IntEntry(vals, i) == 0 {
			lit++
		}
		if saig.Saig_ObjIsPi(g, e) {
			vec.Vec_IntPush(pis, lit)
		} else if !aig.Aig_ObjIsTravIdCurrent(g, e) {
			vec.Vec_IntPush(res, lit-2*saig.Saig_ManPiNum(g))
		}
	}
	if vec.Vec_IntSize(res) == 0 {
		vec.Vec_IntPush(res, 0)
	}
}
func Pdr_ManPrintCex(g *aig.Graph, cis, vals, removed *vec.Vec_Int_t) {
	b := make([]byte, len(g.Inputs()))
	for i := range b {
		b[i] = '-'
	}
	for i, id := range vec.Vec_IntArray(cis) {
		c := byte('0')
		if vec.Vec_IntEntry(vals, i) != 0 {
			c = '1'
		}
		b[aig.Aig_ObjCioId(g, aig.Aig_ManObj(g, id))] = c
	}
	if removed != nil {
		for _, id := range vec.Vec_IntArray(removed) {
			b[aig.Aig_ObjCioId(g, aig.Aig_ManObj(g, id))] = 'x'
		}
	}
	fmt.Println(string(b))
}

// Pdr_ManTernarySim generalizes the SAT assignment while retaining the
// selected output values. Removing a latch is rolled back if it changes a CO.
func Pdr_ManTernarySim(p *Pdr_Man_t, k int32, cube *Pdr_Set_t) *Pdr_Set_t {
	g := p.pAig
	cos, cis, nodes := p.vCoObjs, p.vCiObjs, p.vNodes
	vals, outvals, removed := p.vCiVals, p.vCoVals, p.vCi2Rem
	vec.Vec_IntClear(cos)
	if cube == nil {
		vec.Vec_IntPush(cos, int32(aig.Aig_ManCo(g, p.iOutCur).NodeID()))
	} else {
		for _, lit := range cube.Lits[:cube.nLits] {
			if lit != -1 {
				vec.Vec_IntPush(cos, int32(saig.Saig_ManLi(g, lit>>1).NodeID()))
			}
		}
	}
	if p.pPars.FVeryVerbose != 0 {
		fmt.Print("Trying to justify cube ")
		if cube != nil {
			Pdr_SetPrint(os.Stdout, cube, aig.Aig_ManRegNum(g), nil)
		} else {
			fmt.Print("<prop=fail>")
		}
		fmt.Printf(" in frame %d.\n", k)
	}
	Pdr_ManCollectCone(g, cos, cis, nodes)
	Pdr_ManCollectValues(p, k, cis, vals)
	Pdr_ManCollectValues(p, k, cos, outvals)
	if p.pPars.FVeryVerbose != 0 {
		Pdr_ManPrintCex(g, cis, vals, nil)
	}
	if Pdr_ManSimDataInit(g, cis, vals, nodes, cos, outvals, nil) == 0 {
		panic("pdr: SAT assignment fails ternary simulation")
	}
	vec.Vec_IntClear(removed)
	try := func(e aig.Edge) {
		vec.Vec_IntClear(p.vUndo)
		if Pdr_ManExtendOne(g, e, p.vUndo, p.vVisits) != 0 {
			vec.Vec_IntPush(removed, int32(e.NodeID()))
		} else {
			Pdr_ManExtendUndo(g, p.vUndo)
		}
	}
	if p.pPars.FFlopPrio != 0 {
		vec.Vec_IntClear(p.vRes)
		for _, id := range vec.Vec_IntArray(cis) {
			e := aig.Aig_ManObj(g, id)
			if saig.Saig_ObjIsLo(g, e) {
				vec.Vec_IntPush(p.vRes, aig.Aig_ObjCioId(g, e)-saig.Saig_ManPiNum(g))
			}
		}
		vec.Vec_IntSelectSortCost(vec.Vec_IntArray(p.vRes), vec.Vec_IntSize(p.vRes), p.vPrio)
		for _, index := range vec.Vec_IntArray(p.vRes) {
			try(aig.Aig_ManCi(g, saig.Saig_ManPiNum(g)+index))
		}
	} else {
		for pass := 0; pass < 2; pass++ {
			for _, id := range vec.Vec_IntArray(cis) {
				e := aig.Aig_ManObj(g, id)
				if !saig.Saig_ObjIsLo(g, e) {
					continue
				}
				index := aig.Aig_ObjCioId(g, e) - saig.Saig_ManPiNum(g)
				high := vec.Vec_IntEntry(p.vPrio, int(index)) != 0
				if high != (pass == 1) {
					continue
				}
				try(e)
			}
		}
	}
	if p.pPars.FVeryVerbose != 0 {
		Pdr_ManPrintCex(g, cis, vals, removed)
	}
	if Pdr_ManSimDataInit(g, cis, vals, nodes, cos, outvals, removed) == 0 {
		panic("pdr: generalized cube fails ternary simulation")
	}
	Pdr_ManDeriveResult(g, cis, vals, removed, p.vRes, p.vLits)
	if p.pPars.FUseAbs != 0 && p.vAbsFlops != nil {
		n := 0
		for _, lit := range vec.Vec_IntArray(p.vRes) {
			if vec.Vec_IntEntry(p.vAbsFlops, int(lit>>1)) != 0 {
				vec.Vec_IntWriteEntry(p.vRes, n, lit)
				n++
			} else {
				vec.Vec_IntPush(p.vLits, 2*saig.Saig_ManPiNum(g)+lit)
			}
		}
		vec.Vec_IntShrink(p.vRes, n)
	}
	return Pdr_SetCreate(p.vRes, p.vLits)
}
