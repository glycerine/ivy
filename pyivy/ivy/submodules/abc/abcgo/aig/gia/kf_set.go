package gia

import (
	"bytes"
	"encoding/binary"
	"sort"
	"time"
)

type Kf_Man_t struct {
	pGia         *Gia_Man_t
	pPars        *Jf_Par_t
	pMem         jfCutMemory
	vCuts, vTime Vec_Int_t
	vArea, vRefs Vec_Flt_t
	vTemp        *Vec_Int_t
	clkStart     time.Time
	pSett        [KF_PROC_MAX]Kf_Set_t
}
type Kf_Set_t struct {
	pMan                                        *Kf_Man_t
	nLutSize, nCutNum                           uint16
	nCuts0, nCuts1, nCuts, nTEntries, TableMask int32
	pTable, pValue                              [1 << KF_LOG_TABLE]int32
	pPlace                                      [KF_LEAF_MAX]int32
	pList                                       [KF_LEAF_MAX + 1]int32
	pCuts0, pCuts1                              [KF_CUT_MAX]Kf_Cut_t
	pCutsR                                      [KF_CUT_MAX * KF_CUT_MAX]Kf_Cut_t
	ppCuts                                      [KF_CUT_MAX]*Kf_Cut_t
	pCutBest                                    *Kf_Cut_t
	CutCount                                    [4]word
}

func Kf_SetCutId(p *Kf_Set_t, c *Kf_Cut_t) int32 {
	for i := range p.pCutsR {
		if &p.pCutsR[i] == c {
			return int32(i)
		}
	}
	panic("KF cut outside set")
}
func Kf_SetCut(p *Kf_Set_t, i int32) *Kf_Cut_t {
	if i < 0 {
		return nil
	}
	return &p.pCutsR[i]
}
func Kf_SetPrepare(p *Kf_Set_t, a, b []int32) {
	for i := 0; i <= int(p.nLutSize); i++ {
		p.pList[i] = -1
	}
	p.nCuts0 = Kf_SetLoadCuts(p.pCuts0[:], a)
	p.nCuts1 = Kf_SetLoadCuts(p.pCuts1[:], b)
	p.nCuts = 0
}

// The original comparator uses memcmp on native little-endian int arrays.
func Kf_SetCompareCuts(a, b *Kf_Cut_t) int32 {
	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}
	if a.nLeaves != b.nLeaves {
		return a.nLeaves - b.nLeaves
	}
	for i := int32(0); i < a.nLeaves; i++ {
		var x, y [4]byte
		binary.LittleEndian.PutUint32(x[:], uint32(a.pLeaves[i]))
		binary.LittleEndian.PutUint32(y[:], uint32(b.pLeaves[i]))
		if n := bytes.Compare(x[:], y[:]); n != 0 {
			return int32(n)
		}
	}
	return 0
}
func Kf_SetAddToList(p *Kf_Set_t, c *Kf_Cut_t, ordered int32) {
	id := Kf_SetCutId(p, c)
	place := &p.pList[c.nLeaves]
	if ordered != 0 {
		sort.Slice(c.pLeaves[:c.nLeaves], func(i, j int) bool { return c.pLeaves[i] < c.pLeaves[j] })
		for *place >= 0 {
			other := Kf_SetCut(p, *place)
			cmp := Kf_SetCompareCuts(other, c)
			if cmp > 0 {
				break
			}
			if cmp == 0 {
				panic("KF duplicate cut")
			}
			place = &other.iNext
		}
	}
	c.iNext = *place
	*place = id
}
func Kf_SetStoreAddOne(p *Kf_Set_t, n, limit int32, c *Kf_Cut_t, area int32) int32 {
	p.ppCuts[n] = c
	if n == 0 {
		return 1
	}
	for i := n; i > 0; i-- {
		if Kf_CutCompare(p.ppCuts[i-1], p.ppCuts[i], area) <= 0 {
			break
		}
		p.ppCuts[i-1], p.ppCuts[i] = p.ppCuts[i], p.ppCuts[i-1]
	}
	return min(n+1, limit)
}
func Kf_SetSelectBest(p *Kf_Set_t, area, ordered int32) {
	var n int32
	for i := 0; i <= int(p.nLutSize); i++ {
		for c := Kf_SetCut(p, p.pList[i]); c != nil; c = Kf_SetCut(p, c.iNext) {
			n = Kf_SetStoreAddOne(p, n, int32(p.nCutNum)-1, c, area)
		}
	}
	if n <= 0 || n >= int32(p.nCutNum) {
		panic("KF invalid selected cut count")
	}
	p.nCuts = n
	p.pCutBest = p.ppCuts[0]
	if ordered == 0 {
		return
	}
	for i := 0; i <= int(p.nLutSize); i++ {
		p.pList[i] = -1
	}
	for i := int32(0); i < n; i++ {
		Kf_SetAddToList(p, p.ppCuts[i], 0)
	}
	p.nCuts = 0
	for i := int(p.nLutSize); i >= 0; i-- {
		for c := Kf_SetCut(p, p.pList[i]); c != nil; c = Kf_SetCut(p, c.iNext) {
			p.ppCuts[p.nCuts] = c
			p.nCuts++
		}
	}
}
func Kf_CheckCut(base, cut *Kf_Cut_t) bool { return Kf_SetCutIsContainedSimple(base, cut) }
func Kf_CheckCuts(p *Kf_Set_t) bool {
	if p.nCuts <= 0 {
		panic("KF empty cut set")
	}
	var all []*Kf_Cut_t
	for i := 0; i <= int(p.nLutSize); i++ {
		for c := Kf_SetCut(p, p.pList[i]); c != nil; c = Kf_SetCut(p, c.iNext) {
			all = append(all, c)
			for a := int32(0); a < c.nLeaves; a++ {
				for b := a + 1; b < c.nLeaves; b++ {
					if c.pLeaves[a] == c.pLeaves[b] {
						panic("KF duplicate leaf")
					}
				}
			}
		}
	}
	for _, a := range all {
		for _, b := range all {
			if a != b && Kf_CheckCut(a, b) {
				panic("KF dominated cut")
			}
		}
	}
	return true
}
func Kf_HashLookup(p *Kf_Set_t, id int32) int32 {
	if id <= 0 {
		panic("KF invalid hash leaf")
	}
	for k := id & p.TableMask; ; k = (k + 1) & p.TableMask {
		if p.pTable[k] == 0 {
			return k
		}
		if p.pTable[k] == id {
			return -1
		}
	}
}
func Kf_HashFindOrAdd(p *Kf_Set_t, id int32) int32 {
	k := Kf_HashLookup(p, id)
	if k < 0 {
		return 0
	}
	if p.nTEntries == int32(p.nLutSize) {
		return 1
	}
	p.pTable[k] = id
	p.pPlace[p.nTEntries] = k
	p.pValue[k] = p.nTEntries
	p.nTEntries++
	return 0
}
func Kf_HashPopulate(p *Kf_Set_t, cut *Kf_Cut_t) {
	if p.nTEntries != 0 {
		panic("KF hash not empty")
	}
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		Kf_HashFindOrAdd(p, id)
	}
	if p.nTEntries != cut.nLeaves {
		panic("KF invalid hash population")
	}
}
func Kf_HashCleanup(p *Kf_Set_t, start int32) {
	for i := start; i < p.nTEntries; i++ {
		p.pTable[p.pPlace[i]] = 0
	}
	p.nTEntries = start
}

func kfRemoveDuplicates(p *Kf_Set_t, candidate *Kf_Cut_t, ordered bool) bool {
	for c := Kf_SetCut(p, p.pList[candidate.nLeaves]); c != nil; c = Kf_SetCut(p, c.iNext) {
		if c.Sign != candidate.Sign {
			continue
		}
		if ordered && Kf_SetCutIsContainedOrder(c, candidate) || !ordered && Kf_SetCutIsContainedSimple(c, candidate) {
			return true
		}
	}
	return false
}
func Kf_SetRemoveDuplicatesSimple(p *Kf_Set_t, c *Kf_Cut_t) int32 {
	if kfRemoveDuplicates(p, c, false) {
		return 1
	}
	return 0
}
func Kf_SetRemoveDuplicatesOrder(p *Kf_Set_t, c *Kf_Cut_t) int32 {
	if kfRemoveDuplicates(p, c, true) {
		return 1
	}
	return 0
}
func kfFilter(p *Kf_Set_t, ordered bool) {
	if p.nCuts <= 0 {
		panic("KF empty cut set")
	}
	for i := 0; i <= int(p.nLutSize); i++ {
		place := &p.pList[i]
		for *place >= 0 {
			cut := Kf_SetCut(p, *place)
			dominated := false
			for k := int32(0); k < cut.nLeaves && !dominated; k++ {
				for other := Kf_SetCut(p, p.pList[k]); other != nil; other = Kf_SetCut(p, other.iNext) {
					if cut.Sign&other.Sign != other.Sign {
						continue
					}
					if ordered && Kf_SetCutIsContainedOrder(cut, other) || !ordered && Kf_SetCutIsContainedSimple(cut, other) {
						dominated = true
						break
					}
				}
			}
			if dominated {
				*place = cut.iNext
				p.nCuts--
			} else {
				place = &cut.iNext
			}
		}
	}
	if p.nCuts <= 0 {
		panic("KF all cuts filtered")
	}
}
func Kf_SetFilterSimple(p *Kf_Set_t) { kfFilter(p, false) }
func Kf_SetFilterOrder(p *Kf_Set_t)  { kfFilter(p, true) }
func kfSetMerge(p *Kf_Set_t, a, b []int32, area int32, ordered bool) {
	Kf_SetPrepare(p, a, b)
	p.CutCount[0] += word(p.nCuts0 * p.nCuts1)
	for i := int32(0); i < p.nCuts0; i++ {
		for j := int32(0); j < p.nCuts1; j++ {
			x, y := &p.pCuts0[i], &p.pCuts1[j]
			sign := x.Sign | y.Sign
			if x.nLeaves+y.nLeaves > int32(p.nLutSize) && Kf_SetCountBits(sign) > int32(p.nLutSize) {
				continue
			}
			p.CutCount[1]++
			candidate := &p.pCutsR[p.nCuts]
			var ok int32
			if ordered {
				ok = Kf_SetMergeOrderOne(x, y, candidate, int32(p.nLutSize))
			} else {
				ok = Kf_SetMergeSimpleOne(x, y, candidate, int32(p.nLutSize))
			}
			if ok == 0 {
				continue
			}
			p.CutCount[2]++
			candidate.Sign = sign
			if kfRemoveDuplicates(p, candidate, ordered) {
				continue
			}
			p.nCuts++
			if candidate.nLeaves <= 1 {
				panic("KF cut must have more than one leaf")
			}
			candidate.Delay = max(x.Delay, y.Delay)
			candidate.Area = x.Area + y.Area
			Kf_SetAddToList(p, candidate, 0)
		}
	}
	kfFilter(p, ordered)
	p.CutCount[3] += word(min(p.nCuts, int32(p.nCutNum)-1))
	Kf_SetSelectBest(p, area, 1)
}
func Kf_SetMergeSimple(p *Kf_Set_t, a, b []int32, area, cutMin int32) {
	kfSetMerge(p, a, b, area, false)
}
func Kf_SetMergeOrder(p *Kf_Set_t, a, b []int32, area, cutMin int32) { kfSetMerge(p, a, b, area, true) }

func Kf_SetCutDominatedByThis(p *Kf_Set_t, cut *Kf_Cut_t) int32 {
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		if Kf_HashLookup(p, id) >= 0 {
			return 0
		}
	}
	return 1
}
func Kf_SetRemoveDuplicates(p *Kf_Set_t, n int32, sign word) int32 {
	for c := Kf_SetCut(p, p.pList[n]); c != nil; c = Kf_SetCut(p, c.iNext) {
		if c.Sign == sign && Kf_SetCutDominatedByThis(p, c) != 0 {
			return 1
		}
	}
	return 0
}
func Kf_SetFilter(p *Kf_Set_t) {
	if p.nCuts <= 0 {
		panic("KF empty cut set")
	}
	for i := 0; i <= int(p.nLutSize); i++ {
		place := &p.pList[i]
		for *place >= 0 {
			cut := Kf_SetCut(p, *place)
			Kf_HashPopulate(p, cut)
			dominated := false
			for k := int32(0); k < cut.nLeaves && !dominated; k++ {
				for other := Kf_SetCut(p, p.pList[k]); other != nil; other = Kf_SetCut(p, other.iNext) {
					if cut.Sign&other.Sign == other.Sign && Kf_SetCutDominatedByThis(p, other) != 0 {
						dominated = true
						break
					}
				}
			}
			if dominated {
				*place = cut.iNext
				p.nCuts--
			} else {
				place = &cut.iNext
			}
			Kf_HashCleanup(p, 0)
		}
	}
	if p.nCuts <= 0 {
		panic("KF all cuts filtered")
	}
}
func Kf_SetMergePairs(p *Kf_Set_t, first *Kf_Cut_t, others []Kf_Cut_t, n, area int32) {
	Kf_HashPopulate(p, first)
	for i := int32(0); i < n; i++ {
		other := &others[i]
		sign := first.Sign | other.Sign
		if first.nLeaves+other.nLeaves > int32(p.nLutSize) && Kf_SetCountBits(sign) > int32(p.nLutSize) {
			continue
		}
		Kf_HashCleanup(p, first.nLeaves)
		k := int32(0)
		for ; k < other.nLeaves; k++ {
			if Kf_HashFindOrAdd(p, other.pLeaves[k]) != 0 {
				break
			}
		}
		if k < other.nLeaves {
			continue
		}
		p.CutCount[1]++
		if Kf_SetRemoveDuplicates(p, p.nTEntries, sign) != 0 {
			continue
		}
		cut := &p.pCutsR[p.nCuts]
		p.nCuts++
		cut.nLeaves = p.nTEntries
		for k := int32(0); k < p.nTEntries; k++ {
			cut.pLeaves[k] = p.pTable[p.pPlace[k]]
		}
		cut.Sign = sign
		cut.Delay = max(first.Delay, other.Delay)
		cut.Area = first.Area + other.Area
		Kf_SetAddToList(p, cut, 0)
	}
	Kf_HashCleanup(p, 0)
}
func Kf_SetMerge(p *Kf_Set_t, a, b []int32, area, cutMin int32) {
	Kf_SetPrepare(p, a, b)
	p.CutCount[0] += word(p.nCuts0 * p.nCuts1)
	for i, j := int32(0), int32(0); i < p.nCuts0 && j < p.nCuts1; {
		if p.pCuts0[i].nLeaves >= p.pCuts1[j].nLeaves {
			Kf_SetMergePairs(p, &p.pCuts0[i], p.pCuts1[j:], p.nCuts1-j, area)
			i++
		} else {
			Kf_SetMergePairs(p, &p.pCuts1[j], p.pCuts0[i:], p.nCuts0-i, area)
			j++
		}
	}
	p.CutCount[2] += word(p.nCuts)
	Kf_SetFilter(p)
	p.CutCount[3] += word(min(p.nCuts, int32(p.nCutNum)-1))
	Kf_SetSelectBest(p, area, 1)
}
