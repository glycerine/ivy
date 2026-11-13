package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Gla_ManStart2 initializes the legacy manager without duplicating its graph.
func Gla_ManStart2(g *gia.Gia_Man_t, pars *Abs_Par_t) *Gla_Man_t {
	if g == nil || pars == nil || gia.Gia_ManGateClasses(g) == nil {
		panic("old GLA requires abstraction classes")
	}
	data, err := glaOldCnfForGia(g)
	if err != nil {
		panic(err)
	}
	p := &Gla_Man_t{pGia: g, pPars: pars, vAbs: vec.Vec_IntAlloc(100), vTemp: vec.Vec_IntAlloc(100), vAddedNew: vec.Vec_IntAlloc(100)}
	p.nObjs = 1
	for i := int32(0); i < gia.Gia_ManObjNum(g); i++ {
		obj := gia.Gia_ManObj(g, i)
		if int(i) >= len(data.Obj2Count) {
			panic("AIG/GIA object counts differ")
		}
		if data.Obj2Count[i] >= 0 {
			obj.Value = uint32(p.nObjs)
			p.nObjs++
		} else {
			obj.Value = ^uint32(0)
		}
	}
	clauses := make([][]int32, len(data.Clauses))
	for i, clause := range data.Clauses {
		clauses[i] = make([]int32, len(clause))
		for j, lit := range clause {
			obj := gia.Gia_ManObj(g, lit>>1)
			if obj.Value == ^uint32(0) {
				panic("unmapped old GLA clause variable")
			}
			clauses[i][j] = int32(obj.Value)*2 + (lit & 1)
		}
	}
	p.pCnf = cnf.NewMappedCNF(int(p.nObjs), clauses, nil, data.Obj2Clause, data.Obj2Count)
	p.pObjs = make([]Gla_Obj_t, p.nObjs)
	p.pObj2Obj = make([]uint32, gia.Gia_ManObjNum(g))
	p.pvRefis = make([][]Rfn_Obj_t, gia.Gia_ManObjNum(g))
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	for i := int32(0); i < gia.Gia_ManObjNum(g); i++ {
		obj := gia.Gia_ManObj(g, i)
		p.pObj2Obj[i] = obj.Value
		if obj.Value == ^uint32(0) {
			continue
		}
		v := Gla_ManObj(p, int32(obj.Value))
		v.iGiaObj = i
		v.fConst = uint32(gia.Gia_ObjIsConst0(obj))
		v.fPi = uint32(gia.Gia_ObjIsPi(g, obj))
		v.fPo = uint32(gia.Gia_ObjIsPo(g, obj))
		v.fRi = uint32(gia.Gia_ObjIsRi(g, obj))
		v.fRo = uint32(gia.Gia_ObjIsRo(g, obj))
		v.fAnd = uint32(gia.Gia_ObjIsAnd(obj))
		if v.fConst != 0 || v.fPi != 0 {
			continue
		}
		if v.fAnd != 0 || gia.Gia_ObjIsCo(obj) != 0 {
			if gia.Gia_ObjIsCo(obj) != 0 {
				v.fCompl0 = uint32(gia.Gia_ObjFaninC0(obj))
			}
			Gla_ManCollectFanins(p, v, int32(obj.Value), p.vTemp)
			v.nFanins = uint32(vec.Vec_IntSize(p.vTemp))
			copy(v.Fanins[:], vec.Vec_IntArray(p.vTemp))
			continue
		}
		if v.fRo == 0 {
			panic("invalid old GLA object")
		}
		ri := gia.Gia_ObjRoToRi(g, obj)
		v.nFanins = 1
		v.Fanins[0] = int32(gia.Gia_ObjFanin0(ri).Value)
		v.fCompl0 = uint32(gia.Gia_ObjFaninC0(ri))
	}
	p.pObjRoot = Gla_ManObj(p, int32(gia.Gia_ManPo(g, 0).Value))
	for i := int32(1); i < p.nObjs; i++ {
		obj := Gla_ManObj(p, i)
		if vec.Vec_IntEntry(gia.Gia_ManGateClasses(g), int(obj.iGiaObj)) != 0 {
			obj.fAbs = 1
			vec.Vec_IntPush(p.vAbs, i)
		}
	}
	p.pSat = bsat.NewLabeledSolver2ProofMode(false)
	p.nSatVars = 1
	return p
}
