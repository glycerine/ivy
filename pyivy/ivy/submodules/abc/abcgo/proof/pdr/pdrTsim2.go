package pdr

// Justification-based ternary simulation, ported from pdrTsim2.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

type Txs_Man_t_ struct {
	pGia                                                                       *gia.Gia_Man_t
	vPrio, vCiObjs, vCoObjs, vCiVals, vCoVals, vNodes, vTemp, vPiLits, vFfLits *vec.Vec_Int_t
	pMan                                                                       *Pdr_Man_t
	giaToAig                                                                   []int32
}

func Txs_ManStart(man *Pdr_Man_t, graph *aig.Graph, prio *vec.Vec_Int_t) *Txs_Man_t {
	if vec.Vec_IntSize(prio) != int(aig.Aig_ManRegNum(graph)) {
		panic("pdr: priority count differs from register count")
	}
	g := gia.Gia_ManFromAigSimple(graph)
	p := &Txs_Man_t_{pGia: g, vPrio: prio, pMan: man, vCiObjs: vec.Vec_IntAlloc(100), vCoObjs: vec.Vec_IntAlloc(100), vCiVals: vec.Vec_IntAlloc(100), vCoVals: vec.Vec_IntAlloc(100), vNodes: vec.Vec_IntAlloc(100), vTemp: vec.Vec_IntAlloc(100), vPiLits: vec.Vec_IntAlloc(100), vFfLits: vec.Vec_IntAlloc(100), giaToAig: make([]int32, gia.Gia_ManObjNum(g))}
	for i := range p.giaToAig {
		p.giaToAig[i] = -1
	}
	for i, e := range graph.Inputs() {
		p.giaToAig[gia.Gia_ObjId(g, gia.Gia_ManCi(g, int32(i)))] = int32(e.NodeID())
	}
	for i, e := range graph.Outputs() {
		p.giaToAig[gia.Gia_ObjId(g, gia.Gia_ManCo(g, int32(i)))] = int32(e.NodeID())
	}
	return p
}
func Txs_ManStop(p *Txs_Man_t) {
	if p != nil {
		gia.Gia_ManStop(p.pGia)
		*p = Txs_Man_t_{}
	}
}

// The two original GIA mark bits carry a Boolean value and a justification mark.
func txsVal(o *gia.Gia_Obj_t) int32       { return gia.Sat_ObjXValue(o) & 1 }
func txsMark(o *gia.Gia_Obj_t) bool       { return gia.Sat_ObjXValue(o)&2 != 0 }
func txsSetVal(o *gia.Gia_Obj_t, v int32) { gia.Sat_ObjSetXValue(o, (gia.Sat_ObjXValue(o)&2)|(v&1)) }
func txsSetMark(o *gia.Gia_Obj_t, v bool) {
	x := gia.Sat_ObjXValue(o) & 1
	if v {
		x |= 2
	}
	gia.Sat_ObjSetXValue(o, x)
}
func Txs_ManCollectCone_rec(g *gia.Gia_Man_t, o *gia.Gia_Obj_t, cis, nodes *vec.Vec_Int_t) {
	if o.Value == ^uint32(0) {
		return
	}
	o.Value = ^uint32(0)
	if gia.Gia_ObjIsCi(o) != 0 {
		vec.Vec_IntPush(cis, gia.Gia_ObjId(g, o))
		return
	}
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("pdr: invalid object in simulation cone")
	}
	Txs_ManCollectCone_rec(g, gia.Gia_ObjFanin0(o), cis, nodes)
	Txs_ManCollectCone_rec(g, gia.Gia_ObjFanin1(o), cis, nodes)
	vec.Vec_IntPush(nodes, gia.Gia_ObjId(g, o))
}
func Txs_ManCollectCone(g *gia.Gia_Man_t, cos, cis, nodes *vec.Vec_Int_t) {
	vec.Vec_IntClear(cis)
	vec.Vec_IntClear(nodes)
	gia.Gia_ManConst0(g).Value = ^uint32(0)
	for _, id := range vec.Vec_IntArray(cos) {
		Txs_ManCollectCone_rec(g, gia.Gia_ObjFanin0(gia.Gia_ManObj(g, id)), cis, nodes)
	}
}
func Txs_ManForwardPass(g *gia.Gia_Man_t, prio, cis, ciVals, nodes, cos, coVals *vec.Vec_Int_t) {
	c := gia.Gia_ManConst0(g)
	gia.Sat_ObjSetXValue(c, 0)
	c.Value = math.MaxInt32
	for i, id := range vec.Vec_IntArray(cis) {
		o := gia.Gia_ManObj(g, id)
		gia.Sat_ObjSetXValue(o, vec.Vec_IntEntry(ciVals, i))
		o.Value = math.MaxInt32
		if gia.Gia_ObjIsPi(g, o) == 0 {
			o.Value = uint32(vec.Vec_IntEntry(prio, int(gia.Gia_ObjCioId(o)-gia.Gia_ManPiNum(g))))
		}
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		o := gia.Gia_ManObj(g, id)
		a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
		x, y := txsVal(a)^gia.Gia_ObjFaninC0(o), txsVal(b)^gia.Gia_ObjFaninC1(o)
		gia.Sat_ObjSetXValue(o, x&y)
		switch {
		case x != 0 && y != 0:
			o.Value = min(a.Value, b.Value)
		case x != 0:
			o.Value = b.Value
		case y != 0:
			o.Value = a.Value
		default:
			o.Value = max(a.Value, b.Value)
		}
	}
	for i, id := range vec.Vec_IntArray(cos) {
		o := gia.Gia_ManObj(g, id)
		a := gia.Gia_ObjFanin0(o)
		txsSetVal(o, txsVal(a)^gia.Gia_ObjFaninC0(o))
		txsSetMark(a, true)
		if txsVal(o) != vec.Vec_IntEntry(coVals, i) {
			panic("pdr: simulation disagrees with SAT model")
		}
	}
}
func Txs_ObjIsJust(g *gia.Gia_Man_t, o *gia.Gia_Obj_t) bool {
	a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
	x, y := txsVal(a)^gia.Gia_ObjFaninC0(o), txsVal(b)^gia.Gia_ObjFaninC1(o)
	if txsVal(o) != 0 {
		return txsMark(a) && txsMark(b)
	}
	if x != 0 {
		return txsMark(b) || gia.Gia_ObjIsPi(g, b) != 0
	}
	if y != 0 {
		return txsMark(a) || gia.Gia_ObjIsPi(g, a) != 0
	}
	return txsMark(a) || txsMark(b) || gia.Gia_ObjIsPi(g, a) != 0 || gia.Gia_ObjIsPi(g, b) != 0
}
func Txs_ManBackwardPass(g *gia.Gia_Man_t, cis, nodes, pis, ffs *vec.Vec_Int_t) {
	ns := vec.Vec_IntArray(nodes)
	for i := len(ns) - 1; i >= 0; i-- {
		o := gia.Gia_ManObj(g, ns[i])
		if !txsMark(o) {
			continue
		}
		txsSetMark(o, false)
		a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
		if txsVal(o) != 0 {
			txsSetMark(a, true)
			txsSetMark(b, true)
			continue
		}
		x, y := txsVal(a)^gia.Gia_ObjFaninC0(o), txsVal(b)^gia.Gia_ObjFaninC1(o)
		switch {
		case x != 0:
			txsSetMark(b, true)
		case y != 0:
			txsSetMark(a, true)
		case txsMark(a) || txsMark(b):
		case gia.Gia_ObjIsPi(g, a) != 0:
			txsSetMark(a, true)
		case gia.Gia_ObjIsPi(g, b) != 0:
			txsSetMark(b, true)
		case gia.Gia_ObjIsAnd(a) != 0 && Txs_ObjIsJust(g, a):
			txsSetMark(a, true)
		case gia.Gia_ObjIsAnd(b) != 0 && Txs_ObjIsJust(g, b):
			txsSetMark(b, true)
		case a.Value >= b.Value:
			txsSetMark(a, true)
		default:
			txsSetMark(b, true)
		}
	}
	vec.Vec_IntClear(pis)
	vec.Vec_IntClear(ffs)
	for _, id := range vec.Vec_IntArray(cis) {
		o := gia.Gia_ManObj(g, id)
		if !txsMark(o) {
			continue
		}
		idx := gia.Gia_ObjCioId(o)
		lit := (idx << 1) | (txsVal(o) ^ 1)
		if gia.Gia_ObjIsPi(g, o) != 0 {
			vec.Vec_IntPush(pis, lit)
		} else {
			vec.Vec_IntPush(ffs, lit-(gia.Gia_ManPiNum(g)<<1))
		}
	}
	if vec.Vec_IntSize(ffs) == 0 {
		panic("pdr: backward simulation found no flops")
	}
}
func Txs_ManSelectJustPath(g *gia.Gia_Man_t, nodes, cos, res *vec.Vec_Int_t) {
	for _, id := range vec.Vec_IntArray(cos) {
		txsSetMark(gia.Gia_ObjFanin0(gia.Gia_ManObj(g, id)), true)
	}
	vec.Vec_IntClear(res)
	ns := vec.Vec_IntArray(nodes)
	for i := len(ns) - 1; i >= 0; i-- {
		o := gia.Gia_ManObj(g, ns[i])
		if !txsMark(o) {
			continue
		}
		txsSetMark(o, false)
		vec.Vec_IntPush(res, ns[i])
		a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
		if txsVal(o) != 0 {
			txsSetMark(a, true)
			txsSetMark(b, true)
			continue
		}
		x, y := txsVal(a)^gia.Gia_ObjFaninC0(o), txsVal(b)^gia.Gia_ObjFaninC1(o)
		if x != 0 {
			txsSetMark(b, true)
		} else if y != 0 {
			txsSetMark(a, true)
		} else {
			txsSetMark(a, true)
			txsSetMark(b, true)
		}
	}
	vec.Vec_IntReverseOrder(res)
}
func Txs_ManCollectJustPis(g *gia.Gia_Man_t, cis, pis *vec.Vec_Int_t) {
	vec.Vec_IntClear(pis)
	for _, id := range vec.Vec_IntArray(cis) {
		o := gia.Gia_ManObj(g, id)
		if txsMark(o) && gia.Gia_ObjIsPi(g, o) != 0 {
			vec.Vec_IntPush(pis, gia.Gia_ObjCioId(o)<<1|txsVal(o)^1)
		}
	}
}
func Txs_ManInitPrio(g *gia.Gia_Man_t, cis *vec.Vec_Int_t) {
	gia.Gia_ManConst0(g).Value = math.MaxInt32
	for _, id := range vec.Vec_IntArray(cis) {
		o := gia.Gia_ManObj(g, id)
		o.Value = math.MaxInt32
		if gia.Gia_ObjIsPi(g, o) == 0 {
			o.Value = uint32(gia.Gia_ObjCioId(o) - gia.Gia_ManPiNum(g))
		}
	}
}
func Txs_ManPropagatePrio(g *gia.Gia_Man_t, nodes, prio *vec.Vec_Int_t) {
	for _, id := range vec.Vec_IntArray(nodes) {
		o := gia.Gia_ManObj(g, id)
		a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
		x, y := txsVal(a)^gia.Gia_ObjFaninC0(o), txsVal(b)^gia.Gia_ObjFaninC1(o)
		if txsVal(o) != 0 {
			switch {
			case a.Value == math.MaxInt32:
				o.Value = b.Value
			case b.Value == math.MaxInt32:
				o.Value = a.Value
			case vec.Vec_IntEntry(prio, int(a.Value)) < vec.Vec_IntEntry(prio, int(b.Value)):
				o.Value = a.Value
			default:
				o.Value = b.Value
			}
			continue
		}
		switch {
		case x != 0:
			o.Value = b.Value
		case y != 0:
			o.Value = a.Value
		case a.Value == math.MaxInt32 || b.Value == math.MaxInt32:
			o.Value = math.MaxInt32
		case vec.Vec_IntEntry(prio, int(a.Value)) >= vec.Vec_IntEntry(prio, int(b.Value)):
			o.Value = a.Value
		default:
			o.Value = b.Value
		}
	}
}
func Txs_ManFindMinId(g *gia.Gia_Man_t, cos, prio *vec.Vec_Int_t) int32 {
	best := int32(-1)
	for _, id := range vec.Vec_IntArray(cos) {
		v := gia.Gia_ObjFanin0(gia.Gia_ManObj(g, id)).Value
		if v != math.MaxInt32 && (best < 0 || vec.Vec_IntEntry(prio, int(best)) > vec.Vec_IntEntry(prio, int(v))) {
			best = int32(v)
		}
	}
	return best
}
func Txs_ManFindCiReduction(g *gia.Gia_Man_t, prio, cis, nodes, cos, pis, ffs, temp *vec.Vec_Int_t) {
	Txs_ManSelectJustPath(g, nodes, cos, temp)
	Txs_ManCollectJustPis(g, cis, pis)
	vec.Vec_IntClear(ffs)
	Txs_ManInitPrio(g, cis)
	for {
		Txs_ManPropagatePrio(g, temp, prio)
		id := Txs_ManFindMinId(g, cos, prio)
		if id < 0 {
			break
		}
		o := gia.Gia_ManCi(g, gia.Gia_ManPiNum(g)+id)
		vec.Vec_IntPush(ffs, id<<1|txsVal(o)^1)
		o.Value = math.MaxInt32
	}
}
func Txs_ManPrintFlopLits(ffs, prio *vec.Vec_Int_t) {
	fmt.Printf("%3d : ", vec.Vec_IntSize(ffs))
	for _, lit := range vec.Vec_IntArray(ffs) {
		sign := "-"
		if lit&1 != 0 {
			sign = "+"
		}
		fmt.Printf("%s%d(%d) ", sign, lit>>1, vec.Vec_IntEntry(prio, int(lit>>1)))
	}
	fmt.Println()
}
func Txs_ManVerify(g *gia.Gia_Man_t, cis, nodes, pis, ffs, cos, vals *vec.Vec_Int_t) {
	gia.Gia_ObjTerSimSet0(gia.Gia_ManConst0(g))
	for _, id := range vec.Vec_IntArray(cis) {
		gia.Gia_ObjTerSimSetX(gia.Gia_ManObj(g, id))
	}
	set := func(o *gia.Gia_Obj_t, lit int32) {
		if gia.Gia_ObjTerSimGetX(o) == 0 {
			panic("pdr: duplicate simulation literal")
		}
		if lit&1 != 0 {
			gia.Gia_ObjTerSimSet0(o)
		} else {
			gia.Gia_ObjTerSimSet1(o)
		}
	}
	for _, lit := range vec.Vec_IntArray(pis) {
		set(gia.Gia_ManPi(g, lit>>1), lit)
	}
	for _, lit := range vec.Vec_IntArray(ffs) {
		set(gia.Gia_ManCi(g, gia.Gia_ManPiNum(g)+(lit>>1)), lit)
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		gia.Gia_ObjTerSimAnd(gia.Gia_ManObj(g, id))
	}
	for i, id := range vec.Vec_IntArray(cos) {
		o := gia.Gia_ManObj(g, id)
		gia.Gia_ObjTerSimCo(o)
		if (vec.Vec_IntEntry(vals, i) != 0 && gia.Gia_ObjTerSimGet1(o) == 0) || (vec.Vec_IntEntry(vals, i) == 0 && gia.Gia_ObjTerSimGet0(o) == 0) {
			panic("pdr: ternary reduction does not justify output")
		}
	}
}
func Txs_ManTernarySim(p *Txs_Man_t, k int32, cube *Pdr_Set_t) *Pdr_Set_t {
	vec.Vec_IntClear(p.vCoObjs)
	if cube == nil {
		vec.Vec_IntPush(p.vCoObjs, gia.Gia_ObjId(p.pGia, gia.Gia_ManCo(p.pGia, p.pMan.iOutCur)))
	} else {
		for _, lit := range cube.Lits[:cube.nLits] {
			if lit >= 0 {
				vec.Vec_IntPush(p.vCoObjs, gia.Gia_ObjId(p.pGia, gia.Gia_ManCo(p.pGia, gia.Gia_ManPoNum(p.pGia)+(lit>>1))))
			}
		}
	}
	Txs_ManCollectCone(p.pGia, p.vCoObjs, p.vCiObjs, p.vNodes)
	mapIDs := func(ids *vec.Vec_Int_t) *vec.Vec_Int_t {
		out := vec.Vec_IntAlloc(vec.Vec_IntSize(ids))
		for _, id := range vec.Vec_IntArray(ids) {
			if id < 0 || int(id) >= len(p.giaToAig) || p.giaToAig[id] < 0 {
				panic("pdr: GIA interface object lacks AIG counterpart")
			}
			vec.Vec_IntPush(out, p.giaToAig[id])
		}
		return out
	}
	Pdr_ManCollectValues(p.pMan, k, mapIDs(p.vCiObjs), p.vCiVals)
	Pdr_ManCollectValues(p.pMan, k, mapIDs(p.vCoObjs), p.vCoVals)
	Txs_ManForwardPass(p.pGia, p.vPrio, p.vCiObjs, p.vCiVals, p.vNodes, p.vCoObjs, p.vCoVals)
	Txs_ManFindCiReduction(p.pGia, p.vPrio, p.vCiObjs, p.vNodes, p.vCoObjs, p.vPiLits, p.vFfLits, p.vTemp)
	Txs_ManVerify(p.pGia, p.vCiObjs, p.vNodes, p.vPiLits, p.vFfLits, p.vCoObjs, p.vCoVals)
	return Pdr_SetCreate(p.vFfLits, p.vPiLits)
}
