package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Cec_ManFraSpecReduction constructs the speculative XOR outputs for the
// current candidate classes, recording each representative/node pair.
func Cec_ManFraSpecReduction(p *Cec_ManFra_t) *Gia_Man_t {
	g := p.pAig
	gia.Gia_ManSetPhase(g)
	vec.Vec_IntClear(p.vXorNodes)
	if p.pPars.NLevelMax != 0 {
		gia.Gia_ManLevelNum(g)
	}
	out := gia.Gia_ManStart(gia.Gia_ManObjNum(g))
	name, _ := g.Identity()
	out.SetIdentity(name, name)
	gia.Gia_ManHashAlloc(out)
	copies := make([]int32, g.ObjectCount())
	depths := make([]int32, g.ObjectCount())
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for id := int32(1); int(id) < g.ObjectCount(); id++ {
		a, b, _, kind := g.Gate(id)
		if kind == 1 {
			copies[id] = gia.Gia_ManAppendCi(out)
			continue
		}
		if kind == 6 {
			continue
		}
		if kind != 2 {
			panic("cec: speculative reduction requires an AND-inverter graph")
		}
		if copies[a>>1] < 0 || copies[b>>1] < 0 {
			continue
		}
		node := gia.Gia_ManHashAnd(out, copies[a>>1]^(a&1), copies[b>>1]^(b&1))
		copies[id] = node
		depths[id] = max(depths[a>>1], depths[b>>1])
		repr := gia.Gia_ObjRepr(g, id)
		if repr == gia.GIA_VOID || gia.Gia_ObjFailed(g, id) != 0 {
			continue
		}
		reprCopy := copies[repr]
		if reprCopy < 0 || node>>1 == reprCopy>>1 {
			continue
		}
		if p.pPars.NLevelMax != 0 && (gia.Gia_ObjLevelId(g, id) > p.pPars.NLevelMax || gia.Gia_ObjLevelId(g, reprCopy>>1) > p.pPars.NLevelMax) {
			continue
		}
		if p.pPars.FDualOut != 0 {
			var differs int32
			if p.pPars.fColorDiff != 0 {
				differs = gia.Gia_ObjDiffColors(g, repr, id)
			} else {
				differs = gia.Gia_ObjDiffColors2(g, repr, id)
			}
			if differs == 0 {
				continue
			}
		}
		phase := gia.Gia_ObjPhaseReal(gia.Gia_ManObj(g, id)) ^ gia.Gia_ObjPhaseReal(gia.Gia_ManObj(g, repr))
		copies[id] = reprCopy ^ phase
		if gia.Gia_ObjProved(g, id) != 0 {
			continue
		}
		miter := gia.Gia_ManHashXor(out, node, copies[id])
		gia.Gia_ManAppendCo(out, miter)
		vec.Vec_IntPushTwo(p.vXorNodes, repr, id)
		depths[id] = 1 + max(depths[id], depths[repr])
		if p.pPars.NDepthMax != 0 && depths[id] >= p.pPars.NDepthMax {
			copies[id] = -1
		}
	}
	gia.Gia_ManHashStop(out)
	gia.Gia_ManSetRegNum(out, 0)
	clean := gia.Gia_ManCleanup(out)
	gia.Gia_ManStop(out)
	return clean
}

func Cec_ManFraClassesUpdate_rec(obj *Gia_Obj_t) int32 {
	if gia.Gia_ObjMark0(obj) != 0 {
		return 1
	}
	if gia.Gia_ObjIsCi(obj) != 0 || gia.Gia_ObjIsConst0(obj) != 0 {
		return 0
	}
	value := Cec_ManFraClassesUpdate_rec(gia.Gia_ObjFanin0(obj)) | Cec_ManFraClassesUpdate_rec(gia.Gia_ObjFanin1(obj))
	gia.Gia_ObjSetMark0(obj, value)
	return value
}

func Cec_ManFraCreateInfo(p *Cec_ManSim_t, ciInfo, info *Vec_Ptr_t, series int32) {
	for i := int32(0); i < gia.Gia_ManCiNum(p.pAig); i++ {
		dst := vec.Vec_PtrEntry(ciInfo, int(i)).([]uint32)
		src := vec.Vec_PtrEntry(info, int(i)).([]uint32)
		start := int(p.nWords * series)
		copy(dst, src[start:start+int(p.nWords)])
	}
}

func Cec_ManFraClassesUpdate(p *Cec_ManFra_t, sim *Cec_ManSim_t, pat *Cec_ManPat_t, queries *Gia_Man_t) int32 {
	start := util.Abc_Clock()
	info := Cec_ManPatCollectPatterns(pat, gia.Gia_ManCiNum(p.pAig), sim.nWords)
	p.timePat += util.Abc_Clock() - start
	start = util.Abc_Clock()
	if info != nil {
		gia.Gia_ManCreateValueRefs(p.pAig)
		for i := int32(0); i < pat.nSeries; i++ {
			Cec_ManFraCreateInfo(sim, sim.vCiSimInfo, info, i)
			if Cec_ManSimSimulateRound(sim, sim.vCiSimInfo, sim.vCoSimInfo) != 0 {
				return 1
			}
		}
	}
	p.timeSim += util.Abc_Clock() - start
	if vec.Vec_IntSize(p.vXorNodes) != 2*int(gia.Gia_ManCoNum(queries)) {
		panic("cec: speculative output count mismatch")
	}
	g := p.pAig
	if p.pPars.NDepthMax != 1 {
		gia.Gia_ManCleanMark0(g)
		gia.Gia_ManCleanMark1(g)
		for k := int32(0); k < gia.Gia_ManCoNum(queries); k++ {
			query := gia.Gia_ManCo(queries, k)
			if gia.Gia_ObjMark0(query) == 0 && gia.Gia_ObjMark1(query) == 1 {
				continue
			}
			id := vec.Vec_IntEntry(p.vXorNodes, int(2*k+1))
			gia.Gia_ObjSetMark0(gia.Gia_ManObj(g, id), 1)
		}
		for id := int32(1); int(id) < g.ObjectCount(); id++ {
			obj := gia.Gia_ManObj(g, id)
			if gia.Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			mark := gia.Gia_ObjMark0(obj) | gia.Gia_ObjMark0(gia.Gia_ObjFanin0(obj)) | gia.Gia_ObjMark0(gia.Gia_ObjFanin1(obj))
			gia.Gia_ObjSetMark0(obj, mark)
		}
		for k := int32(0); k < gia.Gia_ManCoNum(queries); k++ {
			query := gia.Gia_ManCo(queries, k)
			if gia.Gia_ObjMark0(query) == 0 && gia.Gia_ObjMark1(query) == 1 {
				continue
			}
			obj := gia.Gia_ManObj(g, vec.Vec_IntEntry(p.vXorNodes, int(2*k+1)))
			if gia.Gia_ObjMark0(gia.Gia_ObjFanin0(obj)) == 0 && gia.Gia_ObjMark0(gia.Gia_ObjFanin1(obj)) == 0 {
				gia.Gia_ObjSetMark1(obj, 1)
			}
		}
		for id := int32(1); int(id) < g.ObjectCount(); id++ {
			obj := gia.Gia_ManObj(g, id)
			if gia.Gia_ObjIsAnd(obj) != 0 && gia.Gia_ObjMark1(obj) != 0 {
				gia.Gia_ObjSetMark0(obj, 0)
				gia.Gia_ObjSetMark1(obj, 0)
			}
		}
	}
	p.nAllProved, p.nAllDisproved, p.nAllFailed = 0, 0, 0
	for k := int32(0); k < gia.Gia_ManCoNum(queries); k++ {
		query := gia.Gia_ManCo(queries, k)
		repr := vec.Vec_IntEntry(p.vXorNodes, int(2*k))
		id := vec.Vec_IntEntry(p.vXorNodes, int(2*k+1))
		r, o := gia.Gia_ManObj(g, repr), gia.Gia_ManObj(g, id)
		if gia.Gia_ObjMark1(query) != 0 {
			if gia.Gia_ObjMark0(r) == 0 && gia.Gia_ObjMark0(o) == 0 {
				gia.Gia_ObjSetProved(g, id)
				p.nAllProved++
			}
		} else if gia.Gia_ObjMark0(query) != 0 {
			if gia.Gia_ObjMark0(r) == 0 && gia.Gia_ObjMark0(o) == 0 {
				if repr == gia.Gia_ObjRepr(g, id) {
					fmt.Println("Cec_ManFraClassesUpdate(): Error! Node is not refined!")
				}
				p.nAllDisproved++
			}
		} else {
			gia.Gia_ObjSetFailed(g, id)
			p.nAllFailed++
		}
	}
	p.nAllProvedS += p.nAllProved
	p.nAllDisprovedS += p.nAllDisproved
	p.nAllFailedS += p.nAllFailed
	return 0
}
