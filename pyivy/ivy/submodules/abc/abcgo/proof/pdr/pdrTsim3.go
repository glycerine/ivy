package pdr

// Improved justification-based ternary simulation, ported from pdrTsim3.c
// by Alan Mishchenko (UC Berkeley, 2010).
import (
	"fmt"
	"slices"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Txs3_Man_t_ struct {
	pGia                               *gia.Gia_Man_t
	vPrio                              *vec.Vec_Int_t
	vCiObjs, vFosPre, vFosAbs, vCoObjs *vec.Vec_Int_t
	vCiVals, vCoVals, vNodes, vTemp    *vec.Vec_Int_t
	vPiLits, vFfLits                   *vec.Vec_Int_t
	pMan                               *Pdr_Man_t
	nPiLits                            int32
	giaToAig                           []int32
	justificationCNF                   *cnf.Cnf_Dat_t
}

func Txs3_ManStart(manager *Pdr_Man_t, graph *aig.Graph, priorities *vec.Vec_Int_t) *Txs3_Man_t {
	if vec.Vec_IntSize(priorities) != int(aig.Aig_ManRegNum(graph)) {
		panic("pdr: ternary simulator priority count differs from register count")
	}
	converted := gia.Gia_ManFromAigSimple(graph)
	p := &Txs3_Man_t_{
		pGia: converted, vPrio: priorities,
		vCiObjs: vec.Vec_IntAlloc(100), vFosPre: vec.Vec_IntAlloc(100),
		vFosAbs: vec.Vec_IntAlloc(100), vCoObjs: vec.Vec_IntAlloc(100),
		vCiVals: vec.Vec_IntAlloc(100), vCoVals: vec.Vec_IntAlloc(100),
		vNodes: vec.Vec_IntAlloc(100), vTemp: vec.Vec_IntAlloc(100),
		vPiLits: vec.Vec_IntAlloc(100), vFfLits: vec.Vec_IntAlloc(100), pMan: manager,
	}
	// GIA conversion may reorder nodes, but preserves input/output order.
	p.giaToAig = make([]int32, gia.Gia_ManObjNum(converted))
	for i := range p.giaToAig {
		p.giaToAig[i] = -1
	}
	for i, edge := range graph.Inputs() {
		p.giaToAig[gia.Gia_ObjId(converted, gia.Gia_ManCi(converted, int32(i)))] = int32(edge.NodeID())
	}
	for i, edge := range graph.Outputs() {
		p.giaToAig[gia.Gia_ObjId(converted, gia.Gia_ManCo(converted, int32(i)))] = int32(edge.NodeID())
	}
	return p
}

func Txs3_ManStop(p *Txs3_Man_t) {
	if p == nil {
		return
	}
	gia.Gia_ManStop(p.pGia)
	*p = Txs3_Man_t_{}
}

// Txs3_ManCollectCone_rec visits each fanin once, separating primary inputs,
// present flops, and abstract flops as in the C implementation.
func Txs3_ManCollectCone_rec(p *Txs3_Man_t, obj *gia.Gia_Obj_t) {
	if obj.Value == ^uint32(0) {
		return
	}
	obj.Value = ^uint32(0)
	if gia.Gia_ObjIsCi(obj) != 0 {
		id := gia.Gia_ObjId(p.pGia, obj)
		if gia.Gia_ObjIsPi(p.pGia, obj) != 0 {
			vec.Vec_IntPush(p.vCiObjs, id)
			return
		}
		index := gia.Gia_ObjCioId(obj) - gia.Gia_ManPiNum(p.pGia)
		if vec.Vec_IntEntry(p.vPrio, int(index)) != 0 {
			vec.Vec_IntPush(p.vFosPre, id)
		} else {
			vec.Vec_IntPush(p.vFosAbs, id)
		}
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("pdr: unexpected object in ternary simulation cone")
	}
	Txs3_ManCollectCone_rec(p, gia.Gia_ObjFanin0(obj))
	Txs3_ManCollectCone_rec(p, gia.Gia_ObjFanin1(obj))
	vec.Vec_IntPush(p.vNodes, gia.Gia_ObjId(p.pGia, obj))
}

func Txs3_ManCollectCone(p *Txs3_Man_t, verbose int32) {
	vec.Vec_IntClear(p.vCiObjs)
	vec.Vec_IntClear(p.vFosPre)
	vec.Vec_IntClear(p.vFosAbs)
	vec.Vec_IntClear(p.vNodes)
	gia.Gia_ManConst0(p.pGia).Value = ^uint32(0)
	for _, id := range vec.Vec_IntArray(p.vCoObjs) {
		Txs3_ManCollectCone_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManObj(p.pGia, id)))
	}
	if verbose != 0 {
		fmt.Printf("%d %d %d\n", vec.Vec_IntSize(p.vCiObjs), vec.Vec_IntSize(p.vFosPre), vec.Vec_IntSize(p.vFosAbs))
	}
	p.nPiLits = int32(vec.Vec_IntSize(p.vCiObjs))
	slices.Sort(vec.Vec_IntArray(p.vCiObjs))
	slices.Sort(vec.Vec_IntArray(p.vFosPre))
	slices.Reverse(vec.Vec_IntArray(p.vFosPre))
	slices.Sort(vec.Vec_IntArray(p.vFosAbs))
	slices.Reverse(vec.Vec_IntArray(p.vFosAbs))
	for _, id := range vec.Vec_IntArray(p.vFosPre) {
		vec.Vec_IntPush(p.vCiObjs, id)
	}
	for _, id := range vec.Vec_IntArray(p.vFosAbs) {
		vec.Vec_IntPush(p.vCiObjs, id)
	}
	for _, id := range vec.Vec_IntArray(p.vCiObjs) {
		gia.Gia_ManObj(p.pGia, id).Value = 0
	}
	for _, id := range vec.Vec_IntArray(p.vNodes) {
		gia.Gia_ManObj(p.pGia, id).Value = 0
	}
}

func Txs3_ManTernarySim(p *Txs3_Man_t, k int32, cube *Pdr_Set_t) *Pdr_Set_t {
	vec.Vec_IntClear(p.vCoObjs)
	if cube == nil {
		obj := gia.Gia_ManCo(p.pGia, p.pMan.iOutCur)
		vec.Vec_IntPush(p.vCoObjs, gia.Gia_ObjId(p.pGia, obj))
	} else {
		for _, lit := range cube.Lits[:cube.nLits] {
			if lit < 0 {
				continue
			}
			obj := gia.Gia_ManCo(p.pGia, gia.Gia_ManPoNum(p.pGia)+(lit>>1))
			vec.Vec_IntPush(p.vCoObjs, gia.Gia_ObjId(p.pGia, obj))
		}
	}
	Txs3_ManCollectCone(p, 0)
	mapObjects := func(objects *vec.Vec_Int_t) *vec.Vec_Int_t {
		mapped := vec.Vec_IntAlloc(vec.Vec_IntSize(objects))
		for _, id := range vec.Vec_IntArray(objects) {
			if id < 0 || int(id) >= len(p.giaToAig) || p.giaToAig[id] < 0 {
				panic("pdr: GIA interface object lacks an AIG counterpart")
			}
			vec.Vec_IntPush(mapped, p.giaToAig[id])
		}
		return mapped
	}
	Pdr_ManCollectValues(p.pMan, k, mapObjects(p.vCiObjs), p.vCiVals)
	Pdr_ManCollectValues(p.pMan, k, mapObjects(p.vCoObjs), p.vCoVals)
	deadline := Pdr_ManTimeLimit(p.pMan)
	if deadline != 0 && util.Abc_Clock() >= deadline || p.pMan.pPars.pFuncStop != nil && p.pMan.pPars.pFuncStop(p.pMan.pPars.RunId) != 0 {
		return nil
	}
	// Frame solvers contain learned reachability clauses. A justification
	// derived under those clauses need not describe the actual transition
	// relation, so minimize against an independent, unconstrained CNF.
	if p.justificationCNF == nil {
		p.justificationCNF = cnf.Cnf_DeriveWithMan(p.pMan.pCnfMan, p.pMan.pAig, aig.Aig_ManCoNum(p.pMan.pAig))
	}
	solver := cnf.Cnf_DataWriteIntoSolver(p.justificationCNF, 1, 0)
	if solver == nil {
		panic("pdr: cannot build justification solver")
	}
	pdrConfigureSolver(solver, p.pMan)
	aux := solver.NewVariable() << 1
	lits := make([]int32, 0, vec.Vec_IntSize(p.vCoObjs)+1)
	for i, id := range vec.Vec_IntArray(p.vCoObjs) {
		variable := p.justificationCNF.VarNums[p.giaToAig[id]]
		if variable < 0 {
			panic("pdr: output missing from justification CNF")
		}
		lit := variable << 1
		if vec.Vec_IntEntry(p.vCoVals, i) != 0 {
			lit |= 1
		}
		lits = append(lits, lit)
	}
	lits = append(lits, aux)
	if !solver.AddClause(lits...) {
		panic("pdr: activation clause rejected")
	}
	solver.Compress()
	vec.Vec_IntClear(p.vTemp)
	vec.Vec_IntPush(p.vTemp, aux^1)
	ids, values := vec.Vec_IntArray(p.vCiObjs), vec.Vec_IntArray(p.vCiVals)
	varToID := make(map[int32]int32, len(ids))
	if len(ids) != len(values) {
		panic("pdr: inconsistent cone values")
	}
	for i, id := range ids {
		obj := aig.Aig_ManObj(p.pMan.pAig, p.giaToAig[id])
		variable := p.justificationCNF.VarNums[obj.NodeID()]
		if variable < 0 || !aig.Aig_ObjIsCi(p.pMan.pAig, obj) {
			panic("pdr: invalid cone input")
		}
		varToID[variable] = int32(obj.NodeID())
		sign := int32(0)
		if values[i] == 0 {
			sign = 1
		}
		vec.Vec_IntPush(p.vTemp, variable<<1|sign)
	}
	status := pdrSolve(p.pMan, solver, vec.Vec_IntArray(p.vTemp), 0)
	if status == bsat.Unknown {
		return nil
	}
	if status != bsat.Unsatisfiable {
		panic("pdr: justification is not unsatisfiable")
	}
	core := solver.UnsatCore()
	vec.Vec_IntClear(p.vTemp)
	for _, lit := range core {
		vec.Vec_IntPush(p.vTemp, lit^1)
	}
	slices.Sort(vec.Vec_IntArray(p.vTemp))
	vec.Vec_IntClear(p.vPiLits)
	vec.Vec_IntClear(p.vFfLits)
	for _, lit := range vec.Vec_IntArray(p.vTemp) {
		if lit == aux^1 {
			continue
		}
		id, ok := varToID[lit>>1]
		if !ok {
			panic("pdr: core literal is not a cone input")
		}
		obj := aig.Aig_ManObj(p.pMan.pAig, id)
		if !aig.Aig_ObjIsCi(p.pMan.pAig, obj) {
			panic("pdr: core literal is not a CI")
		}
		cio := aig.Aig_ObjCioId(p.pMan.pAig, obj)
		if saig.Saig_ObjIsPi(p.pMan.pAig, obj) {
			vec.Vec_IntPush(p.vPiLits, cio<<1|lit&1)
		} else {
			vec.Vec_IntPush(p.vFfLits, (cio-saig.Saig_ManPiNum(p.pMan.pAig))<<1|lit&1)
		}
	}
	if vec.Vec_IntSize(p.vTemp) != vec.Vec_IntSize(p.vPiLits)+vec.Vec_IntSize(p.vFfLits)+1 {
		panic("pdr: missing activation assumption in core")
	}
	if p.pMan.pPars.fUseAbs != 0 && p.pMan.vAbsFlops != nil {
		kept := 0
		for _, lit := range vec.Vec_IntArray(p.vFfLits) {
			if vec.Vec_IntEntry(p.pMan.vAbsFlops, int(lit>>1)) != 0 {
				vec.Vec_IntWriteEntry(p.vFfLits, kept, lit)
				kept++
			} else {
				vec.Vec_IntPush(p.vPiLits, 2*saig.Saig_ManPiNum(p.pMan.pAig)+lit)
			}
		}
		vec.Vec_IntShrink(p.vFfLits, kept)
	}
	result := Pdr_SetCreate(p.vFfLits, p.vPiLits)
	// Abstracted flops become pseudo-inputs, so their removal can leave a
	// state cube that intersects the abstract initial state. Refinement then
	// decides whether the resulting counterexample is concrete.
	if p.pMan.pPars.fUseAbs == 0 && k != 0 && Pdr_SetIsInit(result, -1) != 0 {
		panic("pdr: justification intersects initial state")
	}
	return result
}
