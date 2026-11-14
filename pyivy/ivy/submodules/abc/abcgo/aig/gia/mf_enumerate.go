package gia

func Mf_ObjMergeOrder(p *Mf_Man_t, id int32) {
	obj := Gia_ManObj(p.pGia, id)
	best := Mf_ManObj(p, id)
	limit, num := p.pPars.NLutSize, p.pPars.NCutNum
	var left, right, third, storage [MF_CUT_MAX]Mf_Cut_t
	var cuts [MF_CUT_MAX]*Mf_Cut_t
	for i := range cuts {
		cuts[i] = &storage[i]
	}
	n0 := Mf_ManPrepareCuts(left[:], p, Gia_ObjFaninId0(obj, id), 1)
	n1 := Mf_ManPrepareCuts(right[:], p, Gia_ObjFaninId1(obj, id), 1)
	c0, c1 := Gia_ObjFaninC0(obj), Gia_ObjFaninC1(obj)
	var n int32
	if sibling := Gia_ObjSibl(p.pGia, id); sibling != 0 {
		count := Mf_ManPrepareCuts(third[:], p, sibling, 0)
		complement := Gia_ObjPhase(obj) ^ Gia_ObjPhase(Gia_ObjSiblObj(p.pGia, id))
		for i := int32(0); i < count; i++ {
			*cuts[n] = third[i]
			cuts[n].iFunc ^= uint32(complement)
			Mf_CutParams(p, cuts[n], best.nFlowRefs)
			n = Mf_SetAddCut(cuts[:], n, num)
		}
	}
	if Gia_ObjIsMuxId(p.pGia, id) != 0 {
		n2 := Mf_ManPrepareCuts(third[:], p, Gia_ObjFaninId2(p.pGia, id), 1)
		c2 := Gia_ObjFaninC2(p.pGia, obj)
		p.CutCount[0] += float64(n0 * n1 * n2)
		for i := int32(0); i < n0; i++ {
			for j := int32(0); j < n1; j++ {
				for k := int32(0); k < n2; k++ {
					a, b, c := &left[i], &right[j], &third[k]
					if Mf_CutCountBits(a.Sign|b.Sign|c.Sign) > limit {
						continue
					}
					p.CutCount[1]++
					if Mf_CutMergeOrderMux(a, b, c, cuts[n], limit) == 0 || Mf_SetLastCutIsContained(cuts[:], n) != 0 {
						continue
					}
					p.CutCount[2]++
					if p.pPars.FCutMin != 0 && Mf_CutComputeTruthMux(p, a, b, c, c0, c1, c2, cuts[n]) != 0 {
						cuts[n].Sign = Mf_CutGetSign(cuts[n].pLeaves[:], int32(cuts[n].nLeaves))
					}
					Mf_CutParams(p, cuts[n], best.nFlowRefs)
					n = Mf_SetAddCut(cuts[:], n, num)
				}
			}
		}
	} else {
		p.CutCount[0] += float64(n0 * n1)
		for i := int32(0); i < n0; i++ {
			for j := int32(0); j < n1; j++ {
				a, b := &left[i], &right[j]
				if int32(a.nLeaves+b.nLeaves) > limit && Mf_CutCountBits(a.Sign|b.Sign) > limit {
					continue
				}
				p.CutCount[1]++
				if Mf_CutMergeOrder(a, b, cuts[n], limit) == 0 || Mf_SetLastCutIsContained(cuts[:], n) != 0 {
					continue
				}
				p.CutCount[2]++
				if p.pPars.FCutMin != 0 && Mf_CutComputeTruth(p, a, b, c0, c1, cuts[n], Gia_ObjIsXor(obj)) != 0 {
					cuts[n].Sign = Mf_CutGetSign(cuts[n].pLeaves[:], int32(cuts[n].nLeaves))
				}
				Mf_CutParams(p, cuts[n], best.nFlowRefs)
				n = Mf_SetAddCut(cuts[:], n, num)
			}
		}
	}
	if n <= 0 || n >= num {
		panic("gia: invalid number of mapping cuts")
	}
	best.Flow, best.Delay = cuts[0].Flow, uint32(cuts[0].Delay)&0xffff
	best.iCutSet = Mf_ManSaveCuts(p, cuts[:], n)
	p.nCutCounts[cuts[0].nLeaves]++
	p.CutCount[3] += float64(n)
}
