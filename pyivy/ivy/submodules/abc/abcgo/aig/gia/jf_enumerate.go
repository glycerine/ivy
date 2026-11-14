package gia

import (
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Jf_TtComputeForCut(p *Jf_Man_t, f0, f1 int32, c0, c1, out []int32) int32 {
	lut := p.pPars.NLutSize
	words := 1
	if lut > 6 {
		words = 1 << uint(lut-6)
	}
	t0, t1, result := make([]uint64, words), make([]uint64, words), make([]uint64, words)
	util.Abc_TtCopy(t0, vec.Vec_MemReadEntry(p.vTtMem, int(f0>>1)), int32(words), f0&1)
	util.Abc_TtCopy(t1, vec.Vec_MemReadEntry(p.vTtMem, int(f1>>1)), int32(words), f1&1)
	util.Abc_TtExpand(t0, lut, c0[1:], Jf_CutSize(c0), out[1:], Jf_CutSize(out))
	util.Abc_TtExpand(t1, lut, c1[1:], Jf_CutSize(c1), out[1:], Jf_CutSize(out))
	phase := int32(t0[0] & t1[0] & 1)
	util.Abc_TtAnd(result, t0, t1, int32(words), phase)
	out[0] = util.Abc_TtMinBase(result, out[1:], out[0], lut)
	if result[0]&1 != 0 {
		panic("JF truth table phase")
	}
	return int32(vec.Vec_MemHashInsert(p.vTtMem, result))<<1 | phase
}
func jfCutList(list []int32) [][]int32 {
	cuts := make([][]int32, 0, list[0])
	offset := 1
	for i := int32(0); i < list[0]; i++ {
		if offset >= len(list) {
			panic("JF malformed cut list")
		}
		cut := list[offset:]
		size := int(Jf_CutSize(cut)) + 1
		if size > len(cut) {
			panic("JF malformed cut")
		}
		cuts = append(cuts, cut[:size])
		offset += size
	}
	return cuts
}
func Jf_ObjComputeCuts(p *Jf_Man_t, obj *Gia_Obj_t, edge int32) {
	lut, limit, id := p.pPars.NLutSize, p.pPars.NCutNum, Gia_ObjId(p.pGia, obj)
	var storage [JF_CUT_MAX + 2]Jf_Cut_t
	var store [JF_CUT_MAX + 2]*Jf_Cut_t
	for i := range store {
		store[i] = &storage[i]
		store[i].iFunc = -1
	}
	a := jfCutList(Jf_ObjCuts(p, Gia_ObjFaninId0(obj, id)))
	b := jfCutList(Jf_ObjCuts(p, Gia_ObjFaninId1(obj, id)))
	p.CutCount[0] += word(len(a) * len(b))
	var count int32
	for _, c0 := range a {
		sign0 := Jf_CutGetSign(c0)
		for _, c1 := range b {
			sign := sign0 | Jf_CutGetSign(c1)
			if Jf_CountBits(sign) > lut {
				continue
			}
			p.CutCount[1]++
			candidate := store[count]
			cut := candidate.pCut[:]
			if p.pPars.FCutMin == 0 {
				if Jf_CutMergeOrder(c0, c1, cut, lut) == 0 {
					continue
				}
			} else if p.pPars.FFuncDsd != 0 {
				config := Jf_CutMerge2(c0, c1, cut, lut)
				if config == 0 {
					continue
				}
				old := cut[0]
				candidate.iFunc = extra.Sdm_ManComputeFunc(p.pDsd, Jf_ObjFunc0(obj, c0), Jf_ObjFunc1(obj, c1), cut, config, 0)
				if candidate.iFunc < 0 {
					continue
				}
				if p.pPars.FGenCnf != 0 && Jf_CutCnfSizeF(p, candidate.iFunc>>1) >= 12 {
					continue
				}
				if cut[0] > old {
					panic("JF DSD increased support")
				}
				if cut[0] < old {
					sign = Jf_CutGetSign(cut)
				}
			} else {
				if Jf_CutMergeOrder(c0, c1, cut, lut) == 0 {
					continue
				}
				old := cut[0]
				candidate.iFunc = Jf_TtComputeForCut(p, Jf_ObjFunc0(obj, c0), Jf_ObjFunc1(obj, c1), c0, c1, cut)
				if cut[0] > old {
					panic("JF truth increased support")
				}
				if cut[0] < old {
					sign = Jf_CutGetSign(cut)
				}
				if candidate.iFunc >= 1<<24 {
					panic("JF truth table limit reached")
				}
			}
			candidate.Sign = sign
			p.CutCount[2]++
			candidate.Time = 0
			if p.pPars.FAreaOnly == 0 {
				candidate.Time = Jf_CutArr(p, cut)
			}
			candidate.Flow = Jf_CutFlow(p, cut)
			count = Jf_ObjAddCutToStore(p, store[:], count, limit)
		}
	}
	if !Jf_ObjIsUnit(obj) && Jf_ObjHasCutWithSize(store[:], count, 2) == 0 {
		if !Jf_ObjIsUnit(Gia_ObjFanin0(obj)) || !Jf_ObjIsUnit(Gia_ObjFanin1(obj)) {
			panic("JF coarse fanins not unit")
		}
		cut := store[count]
		if p.pPars.FCutMin != 0 {
			cut.iFunc = 4
		}
		cut.pCut[0] = 2
		cut.pCut[1] = Jf_ObjLit(Gia_ObjFaninId0(obj, id), Gia_ObjFaninC0(obj))
		cut.pCut[2] = Jf_ObjLit(Gia_ObjFaninId1(obj, id), Gia_ObjFaninC1(obj))
		count++
	}
	if Jf_ObjIsUnit(obj) && !(p.pPars.FCutMin != 0 && Jf_ObjHasCutWithSize(store[:], count, 1) != 0) {
		cut := store[count]
		if p.pPars.FCutMin != 0 {
			cut.iFunc = 2
		}
		cut.pCut[0] = 1
		cut.pCut[1] = Jf_ObjLit(id, 0)
		count++
	}
	if count == 0 {
		panic("JF no cuts")
	}
	cost := func(c *Jf_Cut_t) int32 {
		if p.pPars.FGenCnf != 0 {
			return Jf_CutCnfSizeF(p, c.iFunc>>1)
		}
		return 1
	}
	best := store[0]
	best.Cost = cost(best)
	if best.Cost < 0 || best.Flow < 0 {
		panic("JF invalid best cut")
	}
	vec.Vec_IntWriteEntry(&p.vArr, int(id), best.Time)
	extraCost := best.Cost
	if edge != 0 {
		extraCost = best.pCut[0]
	}
	vec.Vec_FltWriteEntry(&p.vFlow, int(id), (best.Flow+float32(extraCost))/Jf_ObjRefs(p, id))
	vec.Vec_IntClear(p.vTemp)
	vec.Vec_IntPush(p.vTemp, count)
	for i := int32(0); i < count; i++ {
		c := store[i]
		c.Cost = cost(c)
		vec.Vec_IntPush(p.vTemp, Jf_CutSetAll(c.iFunc, c.Cost, c.pCut[0]))
		for k := int32(1); k <= c.pCut[0]; k++ {
			vec.Vec_IntPush(p.vTemp, c.pCut[k])
		}
	}
	vec.Vec_IntWriteEntry(&p.vCuts, int(id), p.pMem.append(vec.Vec_IntArray(p.vTemp), vec.Vec_IntSize(p.vTemp)))
	p.CutCount[3] += word(count)
}
