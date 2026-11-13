package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Gla_ManCollectFanins collects the distinct mapped variables used by a cut.
func Gla_ManCollectFanins(p *Gla_Man_t, obj *Gla_Obj_t, id int32, fanins *vec.Vec_Int_t) {
	vec.Vec_IntClear(fanins)
	start, count := p.pCnf.Obj2Clause[obj.iGiaObj], p.pCnf.Obj2Count[obj.iGiaObj]
	if start < 0 || count < 0 || int64(start)+int64(count) > int64(len(p.pCnf.Clauses)) {
		panic("invalid old GLA cut clause range")
	}
	for _, clause := range p.pCnf.Clauses[start : start+count] {
		for _, lit := range clause {
			if lit>>1 != id {
				vec.Vec_IntPushUnique(fanins, lit>>1)
			}
		}
	}
	if vec.Vec_IntSize(fanins) > 4 {
		panic("old GLA cut has more than four fanins")
	}
	vec.Vec_IntSort(fanins, 0)
}

// Gla_ManStart builds the mapped working graph and its labeled SAT manager.
func Gla_ManStart(original *gia.Gia_Man_t, pars *Abs_Par_t) *Gla_Man_t {
	if original == nil || pars == nil || gia.Gia_ManGateClasses(original) == nil {
		panic("old GLA requires abstraction classes")
	}
	graph := gia.Gia_ManToAigSimple(original)
	data, err := cnf.Cnf_DeriveOtherMapped(graph, 1)
	if err != nil {
		panic(err)
	}
	mapping := vec.Vec_IntAllocArrayCopy(data.Mapping, len(data.Mapping))
	work := Gia_ManDupMapped(original, mapping)
	if pars.fPropFanout != 0 {
		gia.Gia_ManStaticFanoutStart(work)
	}
	p := &Gla_Man_t{pGia0: original, pGia: work, pPars: pars, vAbs: vec.Vec_IntAlloc(100), vTemp: vec.Vec_IntAlloc(100), vAddedNew: vec.Vec_IntAlloc(100), vObjCounts: vec.Vec_IntAlloc(100), vCoreCounts: vec.Vec_IntStart(int(gia.Gia_ManObjNum(work))), vProofIds: vec.Vec_IntAlloc(0)}
	n := int(gia.Gia_ManObjNum(work))
	workClasses := vec.Vec_IntStart(n)
	gia.Gia_ManSetGateClasses(work, workClasses)
	counts, offsets := make([]int32, n), make([]int32, n)
	for i := range counts {
		counts[i], offsets[i] = -1, -1
	}
	mapped := vec.Vec_IntStart(n)
	for old := int32(0); old < gia.Gia_ManObjNum(original); old++ {
		obj := gia.Gia_ManObj(original, old)
		if obj.Value == ^uint32(0) {
			continue
		}
		if obj.Value&1 != 0 {
			panic("old GLA object mapping is complemented")
		}
		id := int(obj.Value >> 1)
		obj.Value = uint32(id)
		if id >= n || int(old) >= len(data.Obj2Count) {
			panic("invalid old GLA object mapping")
		}
		counts[id], offsets[id] = data.Obj2Count[old], data.Obj2Clause[old]
		if vec.Vec_IntEntry(gia.Gia_ManGateClasses(original), int(old)) != 0 {
			vec.Vec_IntWriteEntry(workClasses, id, 1)
		}
		off := data.Mapping[old]
		if off == 0 {
			continue
		}
		vec.Vec_IntWriteEntry(mapped, id, int32(vec.Vec_IntSize(mapped)))
		vec.Vec_IntPush(mapped, data.Mapping[off])
		for k := int32(1); k <= 4; k++ {
			leaf := data.Mapping[off+k]
			if leaf < 0 {
				vec.Vec_IntPush(mapped, -1)
			} else {
				value := gia.Gia_ManObj(original, leaf).Value
				if value == ^uint32(0) {
					panic("unmapped old GLA cut leaf")
				}
				vec.Vec_IntPush(mapped, int32(value))
			}
		}
	}
	p.nObjs = 1
	for i := int32(0); i < gia.Gia_ManObjNum(work); i++ {
		obj := gia.Gia_ManObj(work, i)
		if counts[i] >= 0 {
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
			origID := lit >> 1
			orig := gia.Gia_ManObj(original, origID)
			if orig.Value == ^uint32(0) {
				panic("old GLA clause references unmapped original node")
			}
			obj := gia.Gia_ManObj(work, int32(orig.Value))
			if obj.Value == ^uint32(0) {
				panic("old GLA clause references unmapped cut node")
			}
			clauses[i][j] = int32(obj.Value)*2 + (lit & 1)
		}
	}
	p.pCnf = cnf.NewMappedCNF(int(p.nObjs), clauses, nil, offsets, counts)
	p.pObjs = make([]Gla_Obj_t, p.nObjs)
	p.pObj2Obj = make([]uint32, n)
	p.pvRefis = make([][]Rfn_Obj_t, n)
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	for i := int32(0); i < gia.Gia_ManObjNum(work); i++ {
		obj := gia.Gia_ManObj(work, i)
		p.pObj2Obj[i] = obj.Value
		if obj.Value == ^uint32(0) {
			continue
		}
		v := Gla_ManObj(p, int32(obj.Value))
		v.iGiaObj = i
		v.fConst = uint32(gia.Gia_ObjIsConst0(obj))
		v.fPi = uint32(gia.Gia_ObjIsPi(work, obj))
		v.fPo = uint32(gia.Gia_ObjIsPo(work, obj))
		v.fRi = uint32(gia.Gia_ObjIsRi(work, obj))
		v.fRo = uint32(gia.Gia_ObjIsRo(work, obj))
		v.fAnd = uint32(gia.Gia_ObjIsAnd(obj))
		if v.fConst != 0 || v.fPi != 0 {
			continue
		}
		if gia.Gia_ObjIsCo(obj) != 0 {
			v.fCompl0 = uint32(gia.Gia_ObjFaninC0(obj))
			v.nFanins = 1
			v.Fanins[0] = int32(gia.Gia_ObjFanin0(obj).Value)
			continue
		}
		if v.fAnd != 0 {
			off := vec.Vec_IntEntry(mapped, int(i))
			if off == 0 {
				panic("missing old GLA LUT cut")
			}
			for k := 1; k <= 4; k++ {
				leaf := vec.Vec_IntEntry(mapped, int(off)+k)
				if leaf >= 0 {
					v.Fanins[v.nFanins] = int32(gia.Gia_ManObj(work, leaf).Value)
					v.nFanins++
				}
			}
			continue
		}
		if v.fRo == 0 {
			panic("invalid old GLA object type")
		}
		ri := gia.Gia_ObjRoToRi(work, obj)
		v.nFanins = 1
		v.Fanins[0] = int32(gia.Gia_ObjFanin0(ri).Value)
		v.fCompl0 = uint32(gia.Gia_ObjFaninC0(ri))
	}
	p.pObjRoot = Gla_ManObj(p, int32(gia.Gia_ManPo(work, 0).Value))
	for i := int32(1); i < p.nObjs; i++ {
		obj := Gla_ManObj(p, i)
		if vec.Vec_IntEntry(workClasses, int(obj.iGiaObj)) != 0 {
			obj.fAbs = 1
			vec.Vec_IntPush(p.vAbs, i)
		}
	}
	p.pSat = bsat.NewLabeledSolver2()
	p.pSat.SetLearnedPolicy(pars.nLearnedStart, pars.nLearnedDelta, pars.nLearnedPerce)
	p.nSatVars = 1
	p.pRnm = Rnm_ManStart(work)
	return p
}
