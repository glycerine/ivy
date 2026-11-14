package gia

import "github.com/glycerine/abcgo/misc/util"

func Lf_ObjMergeOrder(p *Lf_Man_t, id int32) {
	var storage [LF_CUT_MAX]Lf_Cut_t
	var result [LF_CUT_MAX]*Lf_Cut_t
	for i := range result {
		result[i] = &storage[i]
	}
	obj := Gia_ManObj(p.pGia, id)
	if Gia_ObjIsBuf(obj) != 0 {
		panic("LF cannot merge a buffer")
	}
	best := Lf_ObjReadBest(p, id)
	refs := Lf_ObjFlowRefs(p, id)
	required := Lf_ObjRequired(p, id)
	lut := p.pPars.NLutSize
	if p.pPars.FCutGroup != 0 {
		lut /= 2
	}
	limit := p.pPars.NCutNum
	var set0, set1, set2 []Lf_Cut_t
	n0 := Lf_ManPrepareSet(p, Gia_ObjFaninId0(obj, id), 0, &set0)
	n1 := Lf_ManPrepareSet(p, Gia_ObjFaninId1(obj, id), 1, &set1)
	comp0, comp1 := Gia_ObjFaninC0(obj), Gia_ObjFaninC1(obj)
	var count int32
	if p.Iter != 0 {
		Lf_MemLoadCut(&p.vStoreOld, int32(best.Cut[0].Handle), id, result[0], p.pPars.FCutMin, 1)
		if Lf_BestDiffCuts(best) != 0 {
			Lf_MemLoadCut(&p.vStoreOld, int32(best.Cut[1].Handle), id, result[1], p.pPars.FCutMin, 1)
		}
		if p.fUseEla != 0 && Lf_ObjMapRefNum(p, id) > 0 {
			Lf_CutDeref_rec(p, result[Lf_BestIndex(best)])
		}
		if required == util.ABC_INFINITY {
			required = Lf_CutRequired(p, result[0])
		}
		Lf_CutParams(p, result[0], required, refs, obj)
		count++
		if Lf_BestDiffCuts(best) != 0 {
			Lf_CutParams(p, result[count], required, refs, obj)
			count = Lf_SetAddCut(result[:], count, limit)
		}
		if result[0].fLate != 0 {
			p.nTimeFails++
		}
	}
	if sibling := Gia_ObjSibl(p.pGia, id); sibling != 0 {
		phase := Gia_ObjPhase(obj) ^ Gia_ObjPhase(Gia_ObjSiblObj(p.pGia, id))
		n := Lf_ManPrepareSet(p, sibling, 2, &set2)
		for i := int32(0); i < n; i++ {
			cut := &set2[i]
			if cut.pLeaves[0] == sibling {
				continue
			}
			Lf_CutCopy(result[count], cut, p.nCutWords)
			if result[count].iFunc >= 0 {
				result[count].iFunc ^= phase
			}
			Lf_CutParams(p, result[count], required, refs, obj)
			count = Lf_SetAddCut(result[:], count, limit)
		}
	}
	words := int32(1)
	if lut > 6 {
		words = 1 << uint(lut-6)
	}
	if Gia_ObjIsMuxId(p.pGia, id) != 0 {
		control := Gia_ObjFaninId2(p.pGia, id)
		comp2 := Gia_ObjFaninC2(p.pGia, obj)
		n2 := Lf_ManPrepareSet(p, control, 2, &set2)
		p.CutCount[0] += float64(n0 * n1 * n2)
		var saved *Lf_Cut_t
		combine := func(a, b, c *Lf_Cut_t, size int32, check bool) {
			p.CutCount[1]++
			candidate := result[count]
			if Lf_CutMergeOrderMux(a, b, c, candidate, size) == 0 || Lf_SetLastCutIsContained(result[:], count) != 0 {
				return
			}
			p.CutCount[2]++
			if p.pPars.FCutMin != 0 && Lf_CutComputeTruthMux(p, a, b, c, comp0, comp1, comp2, candidate) != 0 {
				candidate.Sign = Lf_CutGetSign(candidate)
			}
			if check && p.pPars.NLutSizeMux != 0 && p.pPars.NLutSizeMux == int32(candidate.nLeaves) && Lf_ManFindCofVar(Lf_CutTruth(p, candidate), words, int32(candidate.nLeaves)) == -1 {
				return
			}
			Lf_CutParams(p, candidate, required, refs, obj)
			count = Lf_SetAddCut(result[:], count, limit)
		}
		for i := int32(0); i < n0; i++ {
			a := &set0[i]
			if int32(a.nLeaves) > lut {
				continue
			}
			for j := int32(0); j < n1; j++ {
				b := &set1[j]
				if int32(b.nLeaves) > lut {
					continue
				}
				for k := int32(0); k < n2; k++ {
					c := &set2[k]
					if int32(c.nLeaves) > lut {
						continue
					}
					saved = c
					if Lf_CutCountBits(a.Sign|b.Sign|c.Sign) > lut {
						continue
					}
					combine(a, b, c, lut, true)
				}
			}
		}
		if p.pPars.FCutGroup != 0 {
			if saved == nil || saved.nLeaves != 1 || saved.pLeaves[0] != control {
				panic("LF missing unit control cut")
			}
			for i := int32(0); i < n0; i++ {
				a := &set0[i]
				if int32(a.nLeaves) > lut {
					continue
				}
				for j := int32(0); j < n1; j++ {
					b := &set1[j]
					if int32(b.nLeaves) > lut {
						continue
					}
					if int32(a.nLeaves+b.nLeaves)+1 > p.pPars.NLutSize {
						panic("LF grouped cut too large")
					}
					combine(a, b, saved, p.pPars.NLutSize, false)
				}
			}
		}
	} else {
		p.CutCount[0] += float64(n0 * n1)
		xor := Gia_ObjIsXor(obj)
		for i := int32(0); i < n0; i++ {
			a := &set0[i]
			if int32(a.nLeaves) > lut {
				continue
			}
			for j := int32(0); j < n1; j++ {
				b := &set1[j]
				if int32(b.nLeaves) > lut {
					continue
				}
				if int32(a.nLeaves+b.nLeaves) > lut && Lf_CutCountBits(a.Sign|b.Sign) > lut {
					continue
				}
				p.CutCount[1]++
				candidate := result[count]
				if Lf_CutMergeOrder(a, b, candidate, lut) == 0 || Lf_SetLastCutIsContained(result[:], count) != 0 {
					continue
				}
				p.CutCount[2]++
				if p.pPars.FCutMin != 0 && Lf_CutComputeTruth(p, a, b, comp0, comp1, candidate, xor) != 0 {
					candidate.Sign = Lf_CutGetSign(candidate)
				}
				if p.pPars.NLutSizeMux != 0 && p.pPars.NLutSizeMux == int32(candidate.nLeaves) && Lf_ManFindCofVar(Lf_CutTruth(p, candidate), words, int32(candidate.nLeaves)) == -1 {
					continue
				}
				Lf_CutParams(p, candidate, required, refs, obj)
				count = Lf_SetAddCut(result[:], count, limit)
			}
		}
	}
	if count <= 0 || count >= limit || (count > 1 && result[0].Delay > result[1].Delay) {
		panic("LF invalid result cuts")
	}
	best.Cut[0].fUsed = 0
	best.Cut[1].fUsed = 0
	handle := uint32(Lf_MemSaveCut(&p.vStoreNew, result[0], id))
	best.Cut[0].Handle = handle
	best.Cut[1].Handle = handle
	best.Delay[0], best.Delay[1] = result[0].Delay, result[0].Delay
	best.Flow[0], best.Flow[1] = result[0].Flow, result[0].Flow
	p.nCutCounts[result[0].nLeaves]++
	p.CutCount[3] += float64(count)
	p.nCutEqual++
	used := int32(0)
	if count > 1 && result[0].Flow > result[1].Flow+LF_EPSILON {
		best.Cut[1].Handle = uint32(Lf_MemSaveCut(&p.vStoreNew, result[1], id))
		best.Delay[1] = result[1].Delay
		best.Flow[1] = result[1].Flow
		p.nCutCounts[result[1].nLeaves]++
		p.nCutEqual--
		if result[1].fLate == 0 {
			used = 1
		}
	}
	if p.pPars.FUseMux7 != 0 && Gia_ObjIsMuxId(p.pGia, id) != 0 {
		c := Lf_ObjCutMux(p, id)
		Lf_CutParams(p, c, required, refs, obj)
		best.Delay[2] = c.Delay
		best.Flow[2] = c.Flow
	}
	if p.fUseEla != 0 {
		best.Cut[used].fUsed = 1
		if Lf_ObjMapRefNum(p, id) > 0 {
			Lf_CutRef_rec(p, result[used])
		}
	}
	if obj.Value == 0 {
		return
	}
	target := Lf_ManFetchSet(p, id)
	for i := int32(0); i < limit; i++ {
		if i < count {
			Lf_CutCopy(&target[i], result[i], p.nCutWords)
		} else if i == count && result[0].nLeaves > 1 && (count == 1 || result[1].nLeaves > 1) {
			Lf_CutCreateUnit(&target[i], id)
		} else {
			target[i].nLeaves = LF_NO_LEAF
		}
	}
}
