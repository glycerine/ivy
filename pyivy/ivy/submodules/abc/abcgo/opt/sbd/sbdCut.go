package sbd

// Port of src/opt/sbd/sbdCut.c: cut computation.

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const SBD_MAX_CUTSIZE = 10
const SBD_MAX_CUTNUM = 501
const SBD_MAX_TT_WORDS = 16 // ((SBD_MAX_CUTSIZE > 6) ? 1 << (SBD_MAX_CUTSIZE-6) : 1)

const SBD_CUT_NO_LEAF = 0xF

// Sbd_Cut_t keeps the C bitfields (nTreeLeaves:9, nSlowLeaves:9,
// nTopLeaves:10, nLeaves:4) as plain integers; every stored value fits.
type Sbd_Cut_t struct {
	Sign        word  // signature
	iFunc       int32 // functionality
	Cost        int32 // cut cost
	CostLev     int32 // cut cost
	nTreeLeaves int32 // tree leaves
	nSlowLeaves int32 // slow leaves
	nTopLeaves  int32 // top leaves
	nLeaves     int32 // leaf count
	pLeaves     [SBD_MAX_CUTSIZE]int32
}

type Sbd_Sto_t struct {
	nLutSize  int32
	nCutSize  int32
	nCutNum   int32
	fCutMin   int32
	fVerbose  int32
	pGia      *Gia_Man_t                   // user's AIG manager (will be modified by adding nodes)
	vMirrors  *Vec_Int_t                   // mirrors for each node
	vDelays   *Vec_Int_t                   // delays for each node
	vLevels   *Vec_Int_t                   // levels for each node
	vRefs     *Vec_Int_t                   // refs for each node
	vCuts     [][]int32                    // cuts for each node
	vTtMem    *vec.Vec_Mem_t               // truth tables
	pCuts     [3][SBD_MAX_CUTNUM]Sbd_Cut_t // temporary cuts
	ppCuts    [SBD_MAX_CUTNUM]*Sbd_Cut_t   // temporary cut pointers
	nCutsR    int32                        // the number of cuts
	Pivot     int32                        // current object
	iCutBest  int32                        // best-delay cut
	nCutsSpec int32                        // special cuts
	nCutsOver int32                        // overflow cuts
	DelayMin  int32                        // minimum delay
	CutCount  [4]float64                   // cut counters
	clkStart  int64                        // starting time
}

func Sbd_CutTruth(p *Sbd_Sto_t, pCut *Sbd_Cut_t) []word {
	return vec.Vec_MemReadEntry(p.vTtMem, int(Abc_Lit2Var(pCut.iFunc)))
}

func Sbd_CutGetSign(pCut *Sbd_Cut_t) word {
	var Sign word
	for i := int32(0); i < pCut.nLeaves; i++ {
		Sign |= word(1) << uint(pCut.pLeaves[i]&0x3F)
	}
	return Sign
}

// check if pCut is contained in pBase
func Sbd_CutCheck(pBase, pCut *Sbd_Cut_t) int32 {
	nSizeB := pBase.nLeaves
	nSizeC := pCut.nLeaves
	for i := int32(0); i < nSizeC; i++ {
		var k int32
		for k = 0; k < nSizeB; k++ {
			if pCut.pLeaves[i] == pBase.pLeaves[k] {
				break
			}
		}
		if k == nSizeB {
			return 0
		}
	}
	return 1
}
func Sbd_CutSetCheckArray(ppCuts []*Sbd_Cut_t, nCuts int32) int32 {
	assert(nCuts > 0, "nCuts > 0")
	for i := int32(0); i < nCuts; i++ {
		pCut0 := ppCuts[i]
		assert(pCut0.nLeaves <= SBD_MAX_CUTSIZE, "pCut0->nLeaves <= SBD_MAX_CUTSIZE")
		assert(pCut0.Sign == Sbd_CutGetSign(pCut0), "pCut0->Sign == Sbd_CutGetSign(pCut0)")
		for m := int32(0); m < pCut0.nLeaves; m++ {
			for n := m + 1; n < pCut0.nLeaves; n++ {
				assert(pCut0.pLeaves[m] < pCut0.pLeaves[n], "pCut0->pLeaves[m] < pCut0->pLeaves[n]")
			}
		}
		for k := int32(0); k < nCuts; k++ {
			pCut1 := ppCuts[k]
			if pCut0 == pCut1 {
				continue
			}
			assert(Sbd_CutCheck(pCut0, pCut1) == 0, "Value == 0")
		}
	}
	return 1
}

func Sbd_CutMergeOrder(pCut0, pCut1, pCut *Sbd_Cut_t, nCutSize int32) int32 {
	nSize0 := pCut0.nLeaves
	nSize1 := pCut1.nLeaves
	pC0 := &pCut0.pLeaves
	pC1 := &pCut1.pLeaves
	pC := &pCut.pLeaves
	// the case of the largest cut sizes
	if nSize0 == nCutSize && nSize1 == nCutSize {
		for i := int32(0); i < nSize0; i++ {
			if pC0[i] != pC1[i] {
				return 0
			}
			pC[i] = pC0[i]
		}
		pCut.nLeaves = nCutSize
		pCut.iFunc = -1
		pCut.Sign = pCut0.Sign | pCut1.Sign
		return 1
	}
	// compare two cuts with different numbers
	i, k, c := int32(0), int32(0), int32(0)
	if nSize0 == 0 {
		goto FlushCut1
	}
	if nSize1 == 0 {
		goto FlushCut0
	}
	for {
		if c == nCutSize {
			return 0
		}
		if pC0[i] < pC1[k] {
			pC[c] = pC0[i]
			c++
			i++
			if i >= nSize0 {
				goto FlushCut1
			}
		} else if pC0[i] > pC1[k] {
			pC[c] = pC1[k]
			c++
			k++
			if k >= nSize1 {
				goto FlushCut0
			}
		} else {
			pC[c] = pC0[i]
			c++
			i++
			k++
			if i >= nSize0 {
				goto FlushCut1
			}
			if k >= nSize1 {
				goto FlushCut0
			}
		}
	}

FlushCut0:
	if c+nSize0 > nCutSize+i {
		return 0
	}
	for i < nSize0 {
		pC[c] = pC0[i]
		c++
		i++
	}
	pCut.nLeaves = c
	pCut.iFunc = -1
	pCut.Sign = pCut0.Sign | pCut1.Sign
	return 1

FlushCut1:
	if c+nSize1 > nCutSize+k {
		return 0
	}
	for k < nSize1 {
		pC[c] = pC1[k]
		c++
		k++
	}
	pCut.nLeaves = c
	pCut.iFunc = -1
	pCut.Sign = pCut0.Sign | pCut1.Sign
	return 1
}

// check if pCut is contained in pBase
func Sbd_CutSetCutIsContainedOrder(pBase, pCut *Sbd_Cut_t) int32 {
	nSizeB := pBase.nLeaves
	nSizeC := pCut.nLeaves
	if nSizeB == nSizeC {
		for i := int32(0); i < nSizeB; i++ {
			if pBase.pLeaves[i] != pCut.pLeaves[i] {
				return 0
			}
		}
		return 1
	}
	assert(nSizeB > nSizeC, "nSizeB > nSizeC")
	if nSizeC == 0 {
		return 1
	}
	k := int32(0)
	for i := int32(0); i < nSizeB; i++ {
		if pBase.pLeaves[i] > pCut.pLeaves[k] {
			return 0
		}
		if pBase.pLeaves[i] == pCut.pLeaves[k] {
			k++
			if k == nSizeC {
				return 1
			}
		}
	}
	return 0
}
func Sbd_CutSetLastCutIsContained(pCuts []*Sbd_Cut_t, nCuts int32) int32 {
	for i := int32(0); i < nCuts; i++ {
		if pCuts[i].nLeaves <= pCuts[nCuts].nLeaves && (pCuts[i].Sign&pCuts[nCuts].Sign) == pCuts[i].Sign && Sbd_CutSetCutIsContainedOrder(pCuts[nCuts], pCuts[i]) != 0 {
			return 1
		}
	}
	return 0
}

func Sbd_CutCompare(pCut0, pCut1 *Sbd_Cut_t) int32 {
	if pCut0.nLeaves <= 4 && pCut1.nLeaves <= 4 {
		if pCut0.nLeaves < pCut1.nLeaves {
			return -1
		}
		if pCut0.nLeaves > pCut1.nLeaves {
			return 1
		}
		if pCut0.Cost < pCut1.Cost {
			return -1
		}
		if pCut0.Cost > pCut1.Cost {
			return 1
		}
		if pCut0.CostLev < pCut1.CostLev {
			return -1
		}
		if pCut0.CostLev > pCut1.CostLev {
			return 1
		}
	} else if pCut0.nLeaves <= 4 {
		return -1
	} else if pCut1.nLeaves <= 4 {
		return 1
	} else {
		if pCut0.nTreeLeaves < pCut1.nTreeLeaves {
			return -1
		}
		if pCut0.nTreeLeaves > pCut1.nTreeLeaves {
			return 1
		}
		if pCut0.Cost < pCut1.Cost {
			return -1
		}
		if pCut0.Cost > pCut1.Cost {
			return 1
		}
		if pCut0.CostLev < pCut1.CostLev {
			return -1
		}
		if pCut0.CostLev > pCut1.CostLev {
			return 1
		}
		if pCut0.nLeaves < pCut1.nLeaves {
			return -1
		}
		if pCut0.nLeaves > pCut1.nLeaves {
			return 1
		}
	}
	return 0
}
func Sbd_CutCompare2(pCut0, pCut1 *Sbd_Cut_t) int32 {
	// C asserts nLeaves > 4, which holds only for nLutSize >= 4 (callers pass
	// cuts larger than the LUT size); release builds compile it out.
	if sbdDebugAsserts {
		assert(pCut0.nLeaves > 4 && pCut1.nLeaves > 4, "pCut0->nLeaves > 4 && pCut1->nLeaves > 4")
	}
	if pCut0.nSlowLeaves < pCut1.nSlowLeaves {
		return -1
	}
	if pCut0.nSlowLeaves > pCut1.nSlowLeaves {
		return 1
	}
	if pCut0.nTreeLeaves < pCut1.nTreeLeaves {
		return -1
	}
	if pCut0.nTreeLeaves > pCut1.nTreeLeaves {
		return 1
	}
	if pCut0.Cost < pCut1.Cost {
		return -1
	}
	if pCut0.Cost > pCut1.Cost {
		return 1
	}
	if pCut0.CostLev < pCut1.CostLev {
		return -1
	}
	if pCut0.CostLev > pCut1.CostLev {
		return 1
	}
	if pCut0.nLeaves < pCut1.nLeaves {
		return -1
	}
	if pCut0.nLeaves > pCut1.nLeaves {
		return 1
	}
	return 0
}

func Sbd_CutSetLastCutContains(pCuts []*Sbd_Cut_t, nCuts int32) int32 {
	fChanges := false
	for i := int32(0); i < nCuts; i++ {
		if pCuts[nCuts].nLeaves < pCuts[i].nLeaves && (pCuts[nCuts].Sign&pCuts[i].Sign) == pCuts[nCuts].Sign && Sbd_CutSetCutIsContainedOrder(pCuts[i], pCuts[nCuts]) != 0 {
			pCuts[i].nLeaves = SBD_CUT_NO_LEAF
			fChanges = true
		}
	}
	if !fChanges {
		return nCuts
	}
	k := int32(0)
	for i := int32(0); i <= nCuts; i++ {
		if pCuts[i].nLeaves == SBD_CUT_NO_LEAF {
			continue
		}
		if k < i {
			pCuts[k], pCuts[i] = pCuts[i], pCuts[k]
		}
		k++
	}
	return k - 1
}
func Sbd_CutSetSortByCost(pCuts []*Sbd_Cut_t, nCuts int32) {
	for i := nCuts; i > 0; i-- {
		if Sbd_CutCompare(pCuts[i-1], pCuts[i]) < 0 {
			return
		}
		pCuts[i-1], pCuts[i] = pCuts[i], pCuts[i-1]
	}
}
func Sbd_CutSetAddCut(pCuts []*Sbd_Cut_t, nCuts, nCutNum int32) int32 {
	if nCuts == 0 {
		return 1
	}
	nCuts = Sbd_CutSetLastCutContains(pCuts, nCuts)
	assert(nCuts >= 0, "nCuts >= 0")
	Sbd_CutSetSortByCost(pCuts, nCuts)
	// add new cut if there is room
	return min(nCuts+1, nCutNum-1)
}

func Sbd_CutComputeTruth6(p *Sbd_Sto_t, pCut0, pCut1 *Sbd_Cut_t, fCompl0, fCompl1 int32, pCutR *Sbd_Cut_t, fIsXor int32) int32 {
	nOldSupp := pCutR.nLeaves
	t0 := Sbd_CutTruth(p, pCut0)[0]
	t1 := Sbd_CutTruth(p, pCut1)[0]
	if Abc_LitIsCompl(pCut0.iFunc)^fCompl0 != 0 {
		t0 = ^t0
	}
	if Abc_LitIsCompl(pCut1.iFunc)^fCompl1 != 0 {
		t1 = ^t1
	}
	t0 = util.Abc_Tt6Expand(t0, pCut0.pLeaves[:], pCut0.nLeaves, pCutR.pLeaves[:], pCutR.nLeaves)
	t1 = util.Abc_Tt6Expand(t1, pCut1.pLeaves[:], pCut1.nLeaves, pCutR.pLeaves[:], pCutR.nLeaves)
	var t word
	if fIsXor != 0 {
		t = t0 ^ t1
	} else {
		t = t0 & t1
	}
	fCompl := int32(t & 1)
	if fCompl != 0 {
		t = ^t
	}
	pCutR.nLeaves = util.Abc_Tt6MinBase(&t, pCutR.pLeaves[:], pCutR.nLeaves)
	assert(int32(t&1) == 0, "(int)(t & 1) == 0")
	truthId := int32(vec.Vec_MemHashInsert(p.vTtMem, []word{t}))
	pCutR.iFunc = Abc_Var2Lit(truthId, fCompl)
	assert(pCutR.nLeaves <= nOldSupp, "(int)pCutR->nLeaves <= nOldSupp")
	return b2i(pCutR.nLeaves < nOldSupp)
}
func Sbd_CutComputeTruth(p *Sbd_Sto_t, pCut0, pCut1 *Sbd_Cut_t, fCompl0, fCompl1 int32, pCutR *Sbd_Cut_t, fIsXor int32) int32 {
	if p.nCutSize <= 6 {
		return Sbd_CutComputeTruth6(p, pCut0, pCut1, fCompl0, fCompl1, pCutR, fIsXor)
	}
	var uTruth, uTruth0, uTruth1 [SBD_MAX_TT_WORDS]word
	nOldSupp := pCutR.nLeaves
	nCutSize := p.nCutSize
	nWords := util.Abc_Truth6WordNum(nCutSize)
	pTruth0 := Sbd_CutTruth(p, pCut0)
	pTruth1 := Sbd_CutTruth(p, pCut1)
	util.Abc_TtCopy(uTruth0[:], pTruth0, nWords, Abc_LitIsCompl(pCut0.iFunc)^fCompl0)
	util.Abc_TtCopy(uTruth1[:], pTruth1, nWords, Abc_LitIsCompl(pCut1.iFunc)^fCompl1)
	util.Abc_TtExpand(uTruth0[:], nCutSize, pCut0.pLeaves[:], pCut0.nLeaves, pCutR.pLeaves[:], pCutR.nLeaves)
	util.Abc_TtExpand(uTruth1[:], nCutSize, pCut1.pLeaves[:], pCut1.nLeaves, pCutR.pLeaves[:], pCutR.nLeaves)
	var fCompl int32
	if fIsXor != 0 {
		fCompl = int32((uTruth0[0] ^ uTruth1[0]) & 1)
		util.Abc_TtXor(uTruth[:], uTruth0[:], uTruth1[:], nWords, fCompl)
	} else {
		fCompl = int32((uTruth0[0] & uTruth1[0]) & 1)
		util.Abc_TtAnd(uTruth[:], uTruth0[:], uTruth1[:], nWords, fCompl)
	}
	pCutR.nLeaves = util.Abc_TtMinBase(uTruth[:], pCutR.pLeaves[:], pCutR.nLeaves, nCutSize)
	assert((uTruth[0]&1) == 0, "(uTruth[0] & 1) == 0")
	truthId := int32(vec.Vec_MemHashInsert(p.vTtMem, uTruth[:nWords]))
	pCutR.iFunc = Abc_Var2Lit(truthId, fCompl)
	assert(pCutR.nLeaves <= nOldSupp, "(int)pCutR->nLeaves <= nOldSupp")
	return b2i(pCutR.nLeaves < nOldSupp)
}

func Sbd_CutCost(p *Sbd_Sto_t, pCut *Sbd_Cut_t) int32 {
	Cost := int32(0)
	for i := int32(0); i < pCut.nLeaves; i++ {
		Cost += vec.Vec_IntEntry(p.vDelays, int(pCut.pLeaves[i]))
	}
	return Cost
}
func Sbd_CutCostLev(p *Sbd_Sto_t, pCut *Sbd_Cut_t) int32 {
	Cost := int32(0)
	for i := int32(0); i < pCut.nLeaves; i++ {
		Cost += vec.Vec_IntEntry(p.vLevels, int(pCut.pLeaves[i]))
	}
	return Cost
}
func Sbd_CutTreeLeaves(p *Sbd_Sto_t, pCut *Sbd_Cut_t) int32 {
	Cost := int32(0)
	for i := int32(0); i < pCut.nLeaves; i++ {
		Cost += b2i(vec.Vec_IntEntry(p.vRefs, int(pCut.pLeaves[i])) == 1)
	}
	return Cost
}
func Sbd_CutSlowLeaves(p *Sbd_Sto_t, iObj int32, pCut *Sbd_Cut_t) int32 {
	Count, Delay := int32(0), vec.Vec_IntEntry(p.vDelays, int(iObj))
	for i := int32(0); i < pCut.nLeaves; i++ {
		Count += b2i(vec.Vec_IntEntry(p.vDelays, int(pCut.pLeaves[i]))-Delay >= -1)
	}
	return Count
}
func Sbd_CutTopLeaves(p *Sbd_Sto_t, iObj int32, pCut *Sbd_Cut_t) int32 {
	Count, Delay := int32(0), vec.Vec_IntEntry(p.vDelays, int(iObj))
	for i := int32(0); i < pCut.nLeaves; i++ {
		Count += b2i(vec.Vec_IntEntry(p.vDelays, int(pCut.pLeaves[i]))-Delay == -2)
	}
	return Count
}
func Sbd_CutAddUnit(p *Sbd_Sto_t, iObj int32) {
	vThis := p.vCuts[iObj]
	if len(vThis) == 0 {
		vThis = append(vThis, 1)
	} else {
		vThis[0]++
	}
	vThis = append(vThis, 1, iObj, 2)
	p.vCuts[iObj] = vThis
}
func Sbd_CutAddZero(p *Sbd_Sto_t, iObj int32) {
	assert(len(p.vCuts[iObj]) == 0, "Vec_IntSize(vThis) == 0")
	p.vCuts[iObj] = append(p.vCuts[iObj], 1, 0, 0)
}
func Sbd_StoPrepareSet(p *Sbd_Sto_t, iObj int32, Index int) int32 {
	pList := p.vCuts[iObj]
	// Sbd_ForEachCut( pList, pCut, i )
	pos := int32(1)
	for i := int32(0); i < pList[0]; i++ {
		pCut := pList[pos:]
		pCutTemp := &p.pCuts[Index][i]
		pCutTemp.nLeaves = pCut[0]
		for v := int32(1); v <= pCut[0]; v++ {
			pCutTemp.pLeaves[v-1] = pCut[v]
		}
		pCutTemp.iFunc = pCut[pCut[0]+1]
		pCutTemp.Sign = Sbd_CutGetSign(pCutTemp)
		pCutTemp.Cost = Sbd_CutCost(p, pCutTemp)
		pCutTemp.CostLev = Sbd_CutCostLev(p, pCutTemp)
		pCutTemp.nTreeLeaves = Sbd_CutTreeLeaves(p, pCutTemp)
		pCutTemp.nSlowLeaves = Sbd_CutSlowLeaves(p, iObj, pCutTemp)
		pCutTemp.nTopLeaves = Sbd_CutTopLeaves(p, iObj, pCutTemp)
		pos += pCut[0] + 2
	}
	return pList[0]
}
func Sbd_StoInitResult(p *Sbd_Sto_t) {
	for i := 0; i < SBD_MAX_CUTNUM; i++ {
		p.ppCuts[i] = &p.pCuts[2][i]
	}
}
func Sbd_StoStoreResult(p *Sbd_Sto_t, iObj int32, pCuts []*Sbd_Cut_t, nCuts int32) {
	vList := p.vCuts[iObj]
	vList = append(vList, nCuts)
	for i := int32(0); i < nCuts; i++ {
		vList = append(vList, pCuts[i].nLeaves)
		vList = append(vList, pCuts[i].pLeaves[:pCuts[i].nLeaves]...)
		vList = append(vList, pCuts[i].iFunc)
	}
	p.vCuts[iObj] = vList
}
func Sbd_StoComputeDelay(p *Sbd_Sto_t, iObj int32, pCuts []*Sbd_Cut_t, nCuts int32) {
	DelayMin := int32(ABC_INFINITY)
	assert(nCuts > 0, "nCuts > 0")
	p.iCutBest = -1
	for i := int32(0); i < nCuts; i++ {
		if pCuts[i].nLeaves > p.nLutSize {
			continue
		}
		Delay := int32(0)
		for v := int32(0); v < pCuts[i].nLeaves; v++ {
			Delay = max(Delay, vec.Vec_IntEntry(p.vDelays, int(pCuts[i].pLeaves[v])))
		}
		if DelayMin > Delay {
			DelayMin = Delay
			p.iCutBest = i
		} else if DelayMin == Delay && p.iCutBest >= 0 && pCuts[p.iCutBest].nLeaves > pCuts[i].nLeaves {
			p.iCutBest = i
		}
	}
	assert(p.iCutBest >= 0, "p->iCutBest >= 0")
	assert(DelayMin < ABC_INFINITY, "DelayMin < ABC_INFINITY")
	if nCuts > 1 || pCuts[0].nLeaves > 1 {
		DelayMin++
	}
	vec.Vec_IntWriteEntry(p.vDelays, int(iObj), DelayMin)
	p.DelayMin = max(p.DelayMin, DelayMin)
}
func Sbd_StoComputeSpec(p *Sbd_Sto_t, iObj int32, pCuts []*Sbd_Cut_t, nCuts int32) {
	for i := int32(0); i < nCuts; i++ {
		pCuts[i].nTopLeaves = Sbd_CutTopLeaves(p, iObj, pCuts[i])
		pCuts[i].nSlowLeaves = Sbd_CutSlowLeaves(p, iObj, pCuts[i])
		p.nCutsSpec += b2i(pCuts[i].nSlowLeaves == 0)
	}
}
func Sbd_CutPrint(p *Sbd_Sto_t, iObj int32, pCut *Sbd_Cut_t) {
	nDigits := util.Abc_Base10Log(uint32(Gia_ManObjNum(p.pGia)))
	Delay := vec.Vec_IntEntry(p.vDelays, int(iObj))
	if pCut == nil {
		fmt.Printf("No cut.\n")
		return
	}
	fmt.Printf("%d  {", pCut.nLeaves)
	i := int32(0)
	for ; i < pCut.nLeaves; i++ {
		fmt.Printf(" %*d", nDigits, pCut.pLeaves[i])
	}
	for ; i < p.nCutSize; i++ {
		fmt.Printf(" %*s", nDigits, " ")
	}
	fmt.Printf("  }  Cost = %3d  CostL = %3d  Tree = %d  Slow = %d  Top = %d  ",
		pCut.Cost, pCut.CostLev, pCut.nTreeLeaves, pCut.nSlowLeaves, pCut.nTopLeaves)
	if pCut.nSlowLeaves == 0 {
		fmt.Printf("%c ", '*')
	} else {
		fmt.Printf("%c ", ' ')
	}
	for i = 0; i < pCut.nLeaves; i++ {
		fmt.Printf("%3d ", vec.Vec_IntEntry(p.vDelays, int(pCut.pLeaves[i]))-Delay)
	}
	fmt.Printf("\n")
}
func Sbd_StoMergeCuts(p *Sbd_Sto_t, iObj int32) {
	pObj := Gia_ManObj(p.pGia, iObj)
	fIsXor := b2i(Gia_ObjIsXor(pObj))
	nCutSize := p.nCutSize
	nCutNum := p.nCutNum
	Lit0m, Lit1m := int32(-1), int32(-1)
	if p.vMirrors != nil {
		Lit0m = vec.Vec_IntEntry(p.vMirrors, int(Gia_ObjFaninId0(pObj, iObj)))
		Lit1m = vec.Vec_IntEntry(p.vMirrors, int(Gia_ObjFaninId1(pObj, iObj)))
	}
	fComp0 := Gia_ObjFaninC0(pObj) ^ b2i(Lit0m >= 0 && Abc_LitIsCompl(Lit0m) != 0)
	fComp1 := Gia_ObjFaninC1(pObj) ^ b2i(Lit1m >= 0 && Abc_LitIsCompl(Lit1m) != 0)
	Fan0 := Gia_ObjFaninId0(pObj, iObj)
	if Lit0m >= 0 {
		Fan0 = Abc_Lit2Var(Lit0m)
	}
	Fan1 := Gia_ObjFaninId1(pObj, iObj)
	if Lit1m >= 0 {
		Fan1 = Abc_Lit2Var(Lit1m)
	}
	nCuts0 := Sbd_StoPrepareSet(p, Fan0, 0)
	nCuts1 := Sbd_StoPrepareSet(p, Fan1, 1)
	nCutsR := int32(0)
	pCutsR := p.ppCuts[:]
	assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)")
	assert(!Gia_ObjIsMux(p.pGia, pObj), "!Gia_ObjIsMux(p->pGia, pObj)")
	Sbd_StoInitResult(p)
	p.CutCount[0] += float64(nCuts0 * nCuts1)
	for i := int32(0); i < nCuts0; i++ {
		pCut0 := &p.pCuts[0][i]
		for k := int32(0); k < nCuts1; k++ {
			pCut1 := &p.pCuts[1][k]
			if pCut0.nLeaves+pCut1.nLeaves > nCutSize && util.Abc_TtCountOnes(pCut0.Sign|pCut1.Sign) > nCutSize {
				continue
			}
			p.CutCount[1]++
			if Sbd_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nCutSize) == 0 {
				continue
			}
			if Sbd_CutSetLastCutIsContained(pCutsR, nCutsR) != 0 {
				continue
			}
			p.CutCount[2]++
			if p.fCutMin != 0 && Sbd_CutComputeTruth(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0 {
				pCutsR[nCutsR].Sign = Sbd_CutGetSign(pCutsR[nCutsR])
			}
			pCutsR[nCutsR].Cost = Sbd_CutCost(p, pCutsR[nCutsR])
			pCutsR[nCutsR].CostLev = Sbd_CutCostLev(p, pCutsR[nCutsR])
			pCutsR[nCutsR].nTreeLeaves = Sbd_CutTreeLeaves(p, pCutsR[nCutsR])
			nCutsR = Sbd_CutSetAddCut(pCutsR, nCutsR, nCutNum)
		}
	}
	Sbd_StoComputeDelay(p, iObj, pCutsR, nCutsR)
	Sbd_StoComputeSpec(p, iObj, pCutsR, nCutsR)
	p.CutCount[3] += float64(nCutsR)
	p.nCutsOver += b2i(nCutsR == nCutNum-1)
	p.nCutsR = nCutsR
	p.Pivot = iObj
	// verify
	assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum")
	if sbdCheckCuts {
		assert(Sbd_CutSetCheckArray(pCutsR, nCutsR) != 0, "Sbd_CutSetCheckArray(pCutsR, nCutsR)")
	}
	// store the cutset
	Sbd_StoStoreResult(p, iObj, pCutsR, nCutsR)
	if nCutsR > 1 || pCutsR[0].nLeaves > 1 {
		Sbd_CutAddUnit(p, iObj)
	}
}

// sbdDebugAsserts enables C assertions that native release builds (NDEBUG)
// compile out and that fail for legal options other than the defaults
// (-K 2/3 with -c, -K 5/6 with -c, -K 3 with -d) without affecting results.
var sbdDebugAsserts = false

// sbdCheckCuts enables the quadratic cut-set verification that C performs
// only in debug builds (assert( Sbd_CutSetCheckArray(...) )); tests set it.
var sbdCheckCuts = false

// Incremental cut computation.
func Sbd_StoAlloc(pGia *Gia_Man_t, vMirrors *Vec_Int_t, nLutSize, nCutSize, nCutNum, fCutMin, fVerbose int32) *Sbd_Sto_t {
	assert(nLutSize <= nCutSize, "nLutSize <= nCutSize")
	assert(nCutSize < SBD_CUT_NO_LEAF, "nCutSize < SBD_CUT_NO_LEAF")
	assert(nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE, "nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE")
	assert(nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM, "nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM")
	p := &Sbd_Sto_t{}
	p.clkStart = util.Abc_Clock()
	p.nLutSize = nLutSize
	p.nCutSize = nCutSize
	p.nCutNum = nCutNum
	p.fCutMin = fCutMin
	p.fVerbose = fVerbose
	p.pGia = pGia
	p.vMirrors = vMirrors
	p.vDelays = vec.Vec_IntStart(int(Gia_ManObjNum(pGia)))
	p.vLevels = vec.Vec_IntStart(int(Gia_ManObjNum(pGia)))
	p.vRefs = vec.Vec_IntAlloc(int(Gia_ManObjNum(pGia)))
	p.vCuts = make([][]int32, Gia_ManObjNum(pGia))
	if fCutMin != 0 {
		p.vTtMem = vec.Vec_MemAllocForTT(int(nCutSize), 0)
	}
	return p
}
func Sbd_StoFree(p *Sbd_Sto_t) {}
func Sbd_StoComputeCutsObj(p *Sbd_Sto_t, iObj, Delay, Level int32) {
	if int(iObj) < vec.Vec_IntSize(p.vDelays) {
		vec.Vec_IntWriteEntry(p.vDelays, int(iObj), Delay)
		vec.Vec_IntWriteEntry(p.vLevels, int(iObj), Level)
	} else {
		assert(int(iObj) == vec.Vec_IntSize(p.vDelays), "iObj == Vec_IntSize(p->vDelays)")
		assert(int(iObj) == vec.Vec_IntSize(p.vLevels), "iObj == Vec_IntSize(p->vLevels)")
		assert(int(iObj) == len(p.vCuts), "iObj == Vec_WecSize(p->vCuts)")
		vec.Vec_IntPush(p.vDelays, Delay)
		vec.Vec_IntPush(p.vLevels, Level)
		p.vCuts = append(p.vCuts, nil)
	}
}
func Sbd_StoComputeCutsConst0(p *Sbd_Sto_t, iObj int32) {
	Sbd_StoComputeCutsObj(p, iObj, 0, 0)
	Sbd_CutAddZero(p, iObj)
}
func Sbd_StoComputeCutsCi(p *Sbd_Sto_t, iObj, Delay, Level int32) {
	Sbd_StoComputeCutsObj(p, iObj, Delay, Level)
	Sbd_CutAddUnit(p, iObj)
}
func Sbd_StoComputeCutsNode(p *Sbd_Sto_t, iObj int32) int32 {
	pObj := Gia_ManObj(p.pGia, iObj)
	Lev0 := vec.Vec_IntEntry(p.vLevels, int(Gia_ObjFaninId0(pObj, iObj)))
	Lev1 := vec.Vec_IntEntry(p.vLevels, int(Gia_ObjFaninId1(pObj, iObj)))
	Sbd_StoComputeCutsObj(p, iObj, -1, 1+max(Lev0, Lev1))
	Sbd_StoMergeCuts(p, iObj)
	return vec.Vec_IntEntry(p.vDelays, int(iObj))
}
func Sbd_StoSaveBestDelayCut(p *Sbd_Sto_t, iObj int32, pCut []int32) {
	pCutBest := p.ppCuts[p.iCutBest]
	assert(iObj == p.Pivot, "iObj == p->Pivot")
	pCut[0] = pCutBest.nLeaves
	for i := int32(0); i < pCutBest.nLeaves; i++ {
		pCut[i+1] = pCutBest.pLeaves[i]
	}
}
func Sbd_StoObjRefs(p *Sbd_Sto_t, iObj int32) int32 {
	return vec.Vec_IntEntry(p.vRefs, int(iObj))
}
func Sbd_StoRefObj(p *Sbd_Sto_t, iObj, iMirror int32) {
	pObj := Gia_ManObj(p.pGia, iObj)
	assert(int(iObj) == vec.Vec_IntSize(p.vRefs), "iObj == Vec_IntSize(p->vRefs)")
	assert(iMirror < iObj, "iMirror < iObj")
	vec.Vec_IntPush(p.vRefs, 0)
	if iMirror > 0 {
		vec.Vec_IntWriteEntry(p.vRefs, int(iObj), vec.Vec_IntEntry(p.vRefs, int(iMirror)))
		vec.Vec_IntWriteEntry(p.vRefs, int(iMirror), 1)
	}
	if Gia_ObjIsAnd(pObj) {
		Lit0m := vec.Vec_IntEntry(p.vMirrors, int(Gia_ObjFaninId0(pObj, iObj)))
		Lit1m := vec.Vec_IntEntry(p.vMirrors, int(Gia_ObjFaninId1(pObj, iObj)))
		Fan0 := Gia_ObjFaninId0(pObj, iObj)
		if Lit0m >= 0 {
			Fan0 = Abc_Lit2Var(Lit0m)
		}
		Fan1 := Gia_ObjFaninId1(pObj, iObj)
		if Lit1m >= 0 {
			Fan1 = Abc_Lit2Var(Lit1m)
		}
		vec.Vec_IntAddToEntry(p.vRefs, int(Fan0), 1)
		vec.Vec_IntAddToEntry(p.vRefs, int(Fan1), 1)
	} else if Gia_ObjIsCo(pObj) {
		Lit0m := vec.Vec_IntEntry(p.vMirrors, int(Gia_ObjFaninId0(pObj, iObj)))
		assert(Lit0m == -1, "Lit0m == -1")
		vec.Vec_IntAddToEntry(p.vRefs, int(Gia_ObjFaninId0(pObj, iObj)), 1)
	}
}

// Sbd_StoDerefObj returns unconditionally in the original before its code.
func Sbd_StoDerefObj(p *Sbd_Sto_t, iObj int32) {}

func Sbd_StoObjBestCut(p *Sbd_Sto_t, iObj, nSize int32, pLeaves []int32) int32 {
	var pCutBest *Sbd_Cut_t
	assert(p.Pivot == iObj, "p->Pivot == iObj")
	for i := int32(0); i < p.nCutsR; i++ {
		if nSize != 0 && p.ppCuts[i].nLeaves != nSize {
			continue
		}
		if p.ppCuts[i].nLeaves > p.nLutSize &&
			p.ppCuts[i].nSlowLeaves <= 1 &&
			p.ppCuts[i].nTopLeaves <= p.nLutSize-1 &&
			(pCutBest == nil || Sbd_CutCompare2(pCutBest, p.ppCuts[i]) == 1) {
			pCutBest = p.ppCuts[i]
		}
	}
	if pCutBest == nil {
		return -1
	}
	assert(pCutBest.nLeaves <= SBD_DIV_MAX, "pCutBest->nLeaves <= SBD_DIV_MAX")
	for i := int32(0); i < pCutBest.nLeaves; i++ {
		pLeaves[i] = pCutBest.pLeaves[i]
	}
	return pCutBest.nLeaves
}
func Sbd_StoComputeCutsTest(pGia *Gia_Man_t) {
	p := Sbd_StoAlloc(pGia, nil, 4, 8, 100, 1, 1)
	// prepare references
	for iObj := int32(0); iObj < Gia_ManObjNum(p.pGia); iObj++ {
		Sbd_StoRefObj(p, iObj, -1)
	}
	// compute cuts
	Sbd_StoComputeCutsConst0(p, 0)
	for i := int32(0); i < Gia_ManCiNum(p.pGia); i++ {
		Sbd_StoComputeCutsCi(p, Gia_ManCiIdToId(p.pGia, i), 0, 0)
	}
	for iObj := int32(0); iObj < Gia_ManObjNum(p.pGia); iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) {
			Sbd_StoComputeCutsNode(p, iObj)
		}
	}
	if p.fVerbose != 0 {
		fmt.Printf("Running cut computation with LutSize = %d  CutSize = %d  CutNum = %d:\n", p.nLutSize, p.nCutSize, p.nCutNum)
		fmt.Printf("CutPair = %.0f  ", p.CutCount[0])
		fmt.Printf("Merge = %.0f (%.2f %%)  ", p.CutCount[1], 100.0*p.CutCount[1]/p.CutCount[0])
		fmt.Printf("Eval = %.0f (%.2f %%)  ", p.CutCount[2], 100.0*p.CutCount[2]/p.CutCount[0])
		fmt.Printf("Cut = %.0f (%.2f %%)  ", p.CutCount[3], 100.0*p.CutCount[3]/p.CutCount[0])
		fmt.Printf("Cut/Node = %.2f  ", p.CutCount[3]/float64(gia_ManAndNum(p.pGia)))
		fmt.Printf("\n")
		fmt.Printf("Spec = %4d  ", p.nCutsSpec)
		fmt.Printf("Over = %4d  ", p.nCutsOver)
		fmt.Printf("Lev = %4d  ", p.DelayMin)
		util.Abc_PrintTime(0, "Time", util.Abc_Clock()-p.clkStart)
	}
	Sbd_StoFree(p)
}
