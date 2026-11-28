package sfm

// Line-by-line port of src/opt/sfm/sfmLib.c (the library of one- and
// two-gate implementations used by the mfs3 decomposition engine).

import (
	"fmt"
	"time"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

// SFM_WORD_MAX is ((SFM_SUPP_MAX>6) ? (1<<(SFM_SUPP_MAX-6)) : 1).
const SFM_WORD_MAX = 4

type Sfm_Fun_t struct {
	Next   int32                  // next function in the list
	Area   int32                  // area of this function
	pFansT [SFM_SUPP_MAX + 1]byte // top gate ID, followed by fanin perm
	pFansB [SFM_SUPP_MAX + 1]byte // bottom gate ID, followed by fanin perm
}
type Sfm_Lib_t struct {
	nVars       int32             // variable count
	nWords      int32             // truth table words
	fVerbose    int32             // verbose statistics
	pCells      []mio.Mio_Cell2_t // library gates
	nCells      int32             // library gate count
	fDelay      int32             // uses delay profile
	nObjs       int32             // object count
	nObjsAlloc  int32             // object count
	pObjs       []Sfm_Fun_t       // objects
	vTtMem      *vec.Vec_Mem_t    // truth tables
	vLists      []int32           // lists of funcs for each truth table
	vCounts     []int32           // counters of functions for each truth table
	vHits       []int32           // the number of times this function was used
	vProfs      []int32           // area/delay profiles
	vStore      []int32           // storage for area/delay profiles
	vTemp       []int32           // temporary storage for candidates
	nObjSkipped int32
	nObjRemoved int32
}

func Sfm_LibFun(p *Sfm_Lib_t, i int32) *Sfm_Fun_t {
	if i == -1 {
		return nil
	}
	return &p.pObjs[i]
}
func Sfm_LibFunId(p *Sfm_Lib_t, pFun *Sfm_Fun_t) int32 {
	for i := range p.pObjs {
		if &p.pObjs[i] == pFun {
			return int32(i)
		}
	}
	panic("sfm: function is not in the library")
}

var s_Truth8 = [8][4]uint64{
	{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
	{0xCCCCCCCCCCCCCCCC, 0xCCCCCCCCCCCCCCCC, 0xCCCCCCCCCCCCCCCC, 0xCCCCCCCCCCCCCCCC},
	{0xF0F0F0F0F0F0F0F0, 0xF0F0F0F0F0F0F0F0, 0xF0F0F0F0F0F0F0F0, 0xF0F0F0F0F0F0F0F0},
	{0xFF00FF00FF00FF00, 0xFF00FF00FF00FF00, 0xFF00FF00FF00FF00, 0xFF00FF00FF00FF00},
	{0xFFFF0000FFFF0000, 0xFFFF0000FFFF0000, 0xFFFF0000FFFF0000, 0xFFFF0000FFFF0000},
	{0xFFFFFFFF00000000, 0xFFFFFFFF00000000, 0xFFFFFFFF00000000, 0xFFFFFFFF00000000},
	{0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF},
	{0x0000000000000000, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
}

var sfmTruths6 = [6]uint64{
	0xAAAAAAAAAAAAAAAA,
	0xCCCCCCCCCCCCCCCC,
	0xF0F0F0F0F0F0F0F0,
	0xFF00FF00FF00FF00,
	0xFFFF0000FFFF0000,
	0xFFFFFFFF00000000,
}

// Sfm_DecCreateCnf derives the CNF of each library gate.
func Sfm_DecCreateCnf(vGateSizes []int32, vGateFuncs []uint64, vGateCnfs [][]int8) {
	var vCnf []int8
	var vCover []uint32
	for i, uTruth := range vGateFuncs {
		Sfm_TruthToCnf(uTruth, nil, int(vGateSizes[i]), &vCover, &vCnf)
		vGateCnfs[i] = append(vGateCnfs[i][:0], vCnf...)
	}
}

// Sfm_LibPreprocess preprocesses the library: it numbers the gates (their
// Mio value) and derives their sizes, functions and CNFs.
func Sfm_LibPreprocess(pLib *mio.Mio_Library_t, vGateSizes *[]int32, vGateFuncs *[]uint64, vGateCnfs *[][]int8, vGateHands *[]*mio.Mio_Gate_t) {
	nGates := mio.Mio_LibraryReadGateNum(pLib)
	*vGateCnfs = make([][]int8, nGates)
	for pGate := mio.Mio_LibraryReadGates(pLib); pGate != nil; pGate = mio.Mio_GateReadNext(pGate) {
		*vGateSizes = append(*vGateSizes, mio.Mio_GateReadPinNum(pGate))
		*vGateFuncs = append(*vGateFuncs, mio.Mio_GateReadTruth(pGate))
		mio.Mio_GateSetValue(pGate, int32(len(*vGateHands)))
		*vGateHands = append(*vGateHands, pGate)
	}
	Sfm_DecCreateCnf(*vGateSizes, *vGateFuncs, *vGateCnfs)
}

func Sfm_LibFindComplInputGate(vFuncs []uint64, iGate, nFanins, iFanin int32, piFaninNew *int32) int32 {
	uTruthGate := vFuncs[iGate]
	uTruthFlip := util.Abc_Tt6Flip(uTruthGate, iFanin)
	assert(iFanin >= 0 && iFanin < nFanins, "iFanin >= 0 && iFanin < nFanins")
	if piFaninNew != nil {
		*piFaninNew = iFanin
	}
	for i, uTruth := range vFuncs {
		if uTruth == uTruthFlip {
			return int32(i)
		}
	}
	if iFanin-1 >= 0 {
		if piFaninNew != nil {
			*piFaninNew = iFanin - 1
		}
		uTruthSwap := util.Abc_Tt6SwapAdjacent(uTruthFlip, iFanin-1)
		for i, uTruth := range vFuncs {
			if uTruth == uTruthSwap {
				return int32(i)
			}
		}
	}
	if iFanin+1 < nFanins {
		if piFaninNew != nil {
			*piFaninNew = iFanin + 1
		}
		uTruthSwap := util.Abc_Tt6SwapAdjacent(uTruthFlip, iFanin)
		for i, uTruth := range vFuncs {
			if uTruth == uTruthSwap {
				return int32(i)
			}
		}
	}
	// add checking for complemeting control input of a MUX
	if piFaninNew != nil {
		*piFaninNew = -1
	}
	return -1
}

func Sfm_LibStart(nVars, fDelay, fVerbose int32) *Sfm_Lib_t {
	p := new(Sfm_Lib_t)
	assert(nVars <= SFM_SUPP_MAX, "nVars <= SFM_SUPP_MAX")
	p.vTtMem = vec.Vec_MemAllocForTT(int(nVars), 0)
	p.vLists = make([]int32, 2, 1<<16)
	p.vCounts = make([]int32, 2, 1<<16)
	p.vHits = make([]int32, 2, 1<<16)
	p.vLists[0], p.vLists[1] = -1, -1
	p.vCounts[0], p.vCounts[1] = -1, -1
	p.vHits[0], p.vHits[1] = -1, -1
	p.nObjsAlloc = 1 << 16
	p.pObjs = make([]Sfm_Fun_t, p.nObjsAlloc)
	p.fDelay = fDelay
	p.nVars = nVars
	p.nWords = util.Abc_TtWordNum(nVars)
	p.fVerbose = fVerbose
	return p
}
func Sfm_LibStop(p *Sfm_Lib_t) {
	vec.Vec_MemHashFree(p.vTtMem)
	vec.Vec_MemFree(p.vTtMem)
	*p = Sfm_Lib_t{}
}

func Sfm_LibTruth6Two(pCellBot, pCellTop *mio.Mio_Cell2_t, InTop int32) uint64 {
	var uFanins [SFM_SUPP_MAX]uint64
	uTruthBot := mio.Exp_Truth6(pCellBot.Fanins(), pCellBot.Expression(), nil)
	assert(InTop >= 0 && InTop < pCellTop.Fanins(), "InTop >= 0 && InTop < pCellTop->nFanins")
	k := pCellBot.Fanins()
	for i := int32(0); i < pCellTop.Fanins(); i++ {
		if i == InTop {
			uFanins[i] = uTruthBot
		} else {
			uFanins[i] = sfmTruths6[k]
			k++
		}
	}
	assert(pCellBot.Fanins()+pCellTop.Fanins() == k+1, "fanin count")
	uTruthBot = mio.Exp_Truth6(pCellTop.Fanins(), pCellTop.Expression(), uFanins[:])
	return uTruthBot
}
func Sfm_LibTruth8Two(pCellBot, pCellTop *mio.Mio_Cell2_t, InTop int32, pRes []uint64) {
	var uTruthBot [4]uint64
	var puFanins [SFM_SUPP_MAX][]uint64
	mio.Exp_Truth8(pCellBot.Fanins(), pCellBot.Expression(), nil, uTruthBot[:])
	assert(InTop >= 0 && InTop < pCellTop.Fanins(), "InTop >= 0 && InTop < pCellTop->nFanins")
	k := pCellBot.Fanins()
	for i := int32(0); i < pCellTop.Fanins(); i++ {
		if i == InTop {
			puFanins[i] = uTruthBot[:]
		} else {
			puFanins[i] = s_Truth8[k][:]
			k++
		}
	}
	assert(pCellBot.Fanins()+pCellTop.Fanins() == k+1, "fanin count")
	mio.Exp_Truth8(pCellTop.Fanins(), pCellTop.Expression(), puFanins[:pCellTop.Fanins()], pRes)
}

func Sfm_LibCellProfile(pCellBot, pCellTop *mio.Mio_Cell2_t, InTop, nFanins int32, Perm []int32, pProf []int32) {
	DelayAdd := int32(0)
	if pCellTop != nil {
		DelayAdd = 1
	}
	for i := int32(0); i < nFanins; i++ {
		if Perm[i] < pCellBot.Fanins() {
			pProf[i] = 1 + DelayAdd
		} else if Perm[i] < pCellBot.Fanins()+InTop {
			pProf[i] = 1
		} else { // if ( Perm[i] >= (int)pCellBot->nFanins + InTop )
			pProf[i] = 1
		}
	}
}
func Sfm_LibNewIsContained(pObj *Sfm_Fun_t, pProf []int32, Area int32, pProfNew []int32, nFanins int32) bool {
	if Area < pObj.Area {
		return false
	}
	for k := int32(0); k < nFanins; k++ {
		if pProfNew[k] < pProf[k] {
			return false
		}
	}
	return true
}
func Sfm_LibNewContains(pObj *Sfm_Fun_t, pProf []int32, Area int32, pProfNew []int32, nFanins int32) bool {
	if Area > pObj.Area {
		return false
	}
	for k := int32(0); k < nFanins; k++ {
		if pProfNew[k] > pProf[k] {
			return false
		}
	}
	return true
}
func Sfm_LibPrepareAdd(p *Sfm_Lib_t, pTruth []uint64, Perm []int32, nFanins int32, pCellBot, pCellTop *mio.Mio_Cell2_t, InTop int32) {
	var InvPerm, Profile [SFM_SUPP_MAX]int32
	Area := int32(pCellBot.AreaW)
	if pCellTop != nil {
		Area += int32(pCellTop.AreaW)
	}
	iFunc := int32(vec.Vec_MemHashInsert(p.vTtMem, pTruth))
	if int(iFunc) == len(p.vLists) {
		p.vLists = append(p.vLists, -1)
		p.vCounts = append(p.vCounts, 0)
		p.vHits = append(p.vHits, 0)
	}
	assert(pCellBot != nil, "pCellBot != NULL")
	// iterate through the supergates of this truth table
	if p.fDelay != 0 {
		assert(int32(len(p.vProfs)) == p.nObjs, "Vec_IntSize(&p->vProfs) == p->nObjs")
		Sfm_LibCellProfile(pCellBot, pCellTop, InTop, nFanins, Perm, Profile[:])
		// check if new one is contained in old ones
		p.vTemp = p.vTemp[:0]
		for pObj := Sfm_LibFun(p, p.vLists[iFunc]); pObj != nil; pObj = Sfm_LibFun(p, pObj.Next) {
			id := Sfm_LibFunId(p, pObj)
			p.vTemp = append(p.vTemp, id)
			Offset := p.vProfs[id]
			pProf := p.vStore[Offset:]
			if Sfm_LibNewIsContained(pObj, pProf, Area, Profile[:], nFanins) {
				p.nObjSkipped++
				return
			}
		}
		// check if old ones are contained in new one
		k := 0
		i := 0
		for ; i < len(p.vTemp); i++ {
			Id := p.vTemp[i]
			Offset := p.vProfs[Id]
			pProf := p.vStore[Offset:]
			if !Sfm_LibNewContains(Sfm_LibFun(p, Id), pProf, Area, Profile[:], nFanins) {
				p.vTemp[k] = Id
				k++
			} else {
				p.nObjRemoved++
			}
		}
		if k < i { // change
			if k == 0 {
				p.vLists[iFunc] = -1
			} else {
				p.vTemp = p.vTemp[:k]
				Prev := p.vTemp[0]
				p.vLists[iFunc] = Prev
				for _, Id := range p.vTemp[1:] {
					Sfm_LibFun(p, Prev).Next = Id
					Prev = Id
				}
				Sfm_LibFun(p, Prev).Next = -1
			}
		}
	} else {
		for pObj := Sfm_LibFun(p, p.vLists[iFunc]); pObj != nil; pObj = Sfm_LibFun(p, pObj.Next) {
			if Area >= pObj.Area {
				return
			}
		}
	}
	for k := int32(0); k < nFanins; k++ {
		InvPerm[Perm[k]] = k
	}
	// create delay profile
	if p.fDelay != 0 {
		p.vProfs = append(p.vProfs, int32(len(p.vStore)))
		for k := int32(0); k < nFanins; k++ {
			p.vStore = append(p.vStore, Profile[k])
		}
	}
	// create new object
	if p.nObjs == p.nObjsAlloc {
		nObjsAlloc := 2 * p.nObjsAlloc
		p.pObjs = append(p.pObjs, make([]Sfm_Fun_t, p.nObjsAlloc)...)
		p.nObjsAlloc = nObjsAlloc
	}
	pObj := &p.pObjs[p.nObjs]
	pObj.Area = Area
	pObj.Next = p.vLists[iFunc]
	p.vLists[iFunc] = p.nObjs
	p.nObjs++
	p.vCounts[iFunc]++
	// create gate
	assert(pCellBot.Id < 128, "pCellBot->Id < 128")
	pObj.pFansB[0] = byte(pCellBot.Id)
	var k int32
	for k = 0; k < pCellBot.Fanins(); k++ {
		pObj.pFansB[k+1] = byte(InvPerm[k])
	}
	if pCellTop == nil {
		return
	}
	assert(pCellTop.Id < 128, "pCellTop->Id < 128")
	pObj.pFansT[0] = byte(pCellTop.Id)
	for i := int32(0); i < pCellTop.Fanins(); i++ {
		if i == InTop {
			pObj.pFansT[i+1] = 16
		} else {
			pObj.pFansT[i+1] = byte(InvPerm[k])
			k++
		}
	}
	assert(k == nFanins, "k == nFanins")
}
func Sfm_LibPrepare(nVars, fTwo, fDelay, fVerbose, fLibVerbose int32) *Sfm_Lib_t {
	clk := time.Now()
	p := Sfm_LibStart(nVars, fDelay, fLibVerbose)
	var pPerm [SFM_SUPP_MAX + 1][]int32
	var nPerms [SFM_SUPP_MAX + 1]int32
	var Perm [SFM_SUPP_MAX]int32
	var tTemp1, tCur [4]uint64
	assert(nVars <= SFM_SUPP_MAX, "nVars <= SFM_SUPP_MAX")
	// precompute gates
	p.pCells = mio.Mio_CollectRootsNewDefault2(min(6, nVars), &p.nCells, 0)
	// find useful ones
	for c := int32(4); c < p.nCells; c++ {
		pCell1 := &p.pCells[c]
		uTruth := pCell1.Truth()
		pCell1.Type = 0
		if util.Abc_Tt6IsAndType(uTruth, pCell1.Fanins()) != 0 || util.Abc_Tt6IsOrType(uTruth, pCell1.Fanins()) != 0 {
			pCell1.Type = 1
		} else {
			var pRes []byte
			if dau.Dau_DsdDecompose([]uint64{uTruth}, pCell1.Fanins(), 0, 0, &pRes) <= 3 {
				pCell1.Type = 2
			} else if fLibVerbose != 0 {
				fmt.Printf("Skipping gate \"%s\" with non-DSD function %s\n", pCell1.Name(), string(pRes))
			}
		}
	}
	// generate permutations
	for i := int32(2); i <= nVars; i++ {
		pPerm[i] = extra.Extra_PermSchedule(i)
	}
	for i := int32(2); i <= nVars; i++ {
		nPerms[i] = extra.Extra_Factorial(i)
	}
	// add single cells
	for c := int32(4); c < p.nCells; c++ {
		pCell1 := &p.pCells[c]
		nFanins := pCell1.Fanins()
		assert(nFanins >= 2 && nFanins <= nVars, "nFanins >= 2 && nFanins <= nVars")
		for i := int32(0); i < nFanins; i++ {
			Perm[i] = i
		}
		// permute truth table
		tCur[0] = pCell1.Truth()
		tTemp1[0] = tCur[0]
		if p.nVars > 6 {
			tTemp1[1], tTemp1[2], tTemp1[3] = tCur[0], tCur[0], tCur[0]
			tCur[1], tCur[2], tCur[3] = tCur[0], tCur[0], tCur[0]
		}
		for n := int32(0); n < nPerms[nFanins]; n++ {
			Sfm_LibPrepareAdd(p, tCur[:p.nWords], Perm[:], nFanins, pCell1, nil, -1)
			// update
			util.Abc_TtSwapAdjacent(tCur[:], p.nWords, pPerm[nFanins][n])
			s := pPerm[nFanins][n]
			Perm[s], Perm[s+1] = Perm[s+1], Perm[s]
		}
		assert(util.Abc_TtEqual(tTemp1[:], tCur[:], p.nWords) != 0, "Abc_TtEqual(tTemp1, tCur, p->nWords)")
	}
	// add double cells
	if fTwo != 0 {
		for c1 := int32(4); c1 < p.nCells; c1++ { // Bot
			pCell1 := &p.pCells[c1]
			if pCell1.Type == 0 {
				continue
			}
			for c2 := int32(4); c2 < p.nCells; c2++ { // Top
				pCell2 := &p.pCells[c2]
				if pCell2.Type == 0 {
					continue
				}
				if pCell1.Fanins()+pCell2.Fanins() > nVars+1 {
					continue
				}
				for f := int32(0); f < pCell2.Fanins(); f++ {
					nFanins := pCell1.Fanins() + pCell2.Fanins() - 1
					assert(nFanins >= 2 && nFanins <= nVars, "nFanins >= 2 && nFanins <= nVars")
					for i := int32(0); i < nFanins; i++ {
						Perm[i] = i
					}
					// permute truth table
					if p.nVars > 6 {
						Sfm_LibTruth8Two(pCell1, pCell2, f, tCur[:])
						util.Abc_TtCopy(tTemp1[:], tCur[:], p.nWords, 0)
					} else {
						tCur[0] = Sfm_LibTruth6Two(pCell1, pCell2, f)
						tTemp1[0] = tCur[0]
					}
					for n := int32(0); n < nPerms[nFanins]; n++ {
						Sfm_LibPrepareAdd(p, tCur[:p.nWords], Perm[:], nFanins, pCell1, pCell2, f)
						if nFanins > 5 {
							break
						}
						// update
						util.Abc_TtSwapAdjacent(tCur[:], p.nWords, pPerm[nFanins][n])
						s := pPerm[nFanins][n]
						Perm[s], Perm[s+1] = Perm[s+1], Perm[s]
					}
					assert(util.Abc_TtEqual(tTemp1[:], tCur[:], p.nWords) != 0, "Abc_TtEqual(tTemp1, tCur, p->nWords)")
				}
			}
		}
	}
	if fVerbose != 0 {
		nFun := int32(vec.Vec_MemEntryNum(p.vTtMem) - 2)
		fmt.Printf("Library processing: Var = %d. Cell = %d.  Fun = %d. Obj = %d. Ave = %.2f.  Skip = %d. Rem = %d.  ",
			nVars, p.nCells, nFun,
			p.nObjs-p.nObjRemoved, 1.0*float64(p.nObjs-p.nObjRemoved)/float64(nFun),
			p.nObjSkipped, p.nObjRemoved)
		fmt.Printf("%s =%9.2f sec\n", "Time", time.Since(clk).Seconds())
	}
	return p
}
func Sfm_LibPrintGate(pCell *mio.Mio_Cell2_t, pFanins []byte, pCell2 *mio.Mio_Cell2_t, pFanins2 []byte) {
	fmt.Printf(" %-20s(", pCell.Name())
	for k := int32(0); k < pCell.Fanins(); k++ {
		if pFanins[k] == 16 {
			Sfm_LibPrintGate(pCell2, pFanins2, nil, nil)
		} else {
			fmt.Printf(" %c", 'a'+pFanins[k])
		}
	}
	fmt.Printf(" )")
}
func Sfm_LibPrintObj(p *Sfm_Lib_t, pObj *Sfm_Fun_t) {
	pCellB := &p.pCells[pObj.pFansB[0]]
	pCellT := &p.pCells[pObj.pFansT[0]]
	nFanins := pCellB.Fanins()
	if pCellT != &p.pCells[0] {
		nFanins += pCellT.Fanins() - 1
	}
	fmt.Printf("F = %d  A =%6.2f  ", nFanins, scl.Scl_Int2Flt(pObj.Area))
	if pCellT == &p.pCells[0] {
		Sfm_LibPrintGate(pCellB, pObj.pFansB[1:], nil, nil)
	} else {
		Sfm_LibPrintGate(pCellT, pObj.pFansT[1:], pCellB, pObj.pFansB[1:])
	}
	// get hold of delay info
	if p.fDelay != 0 {
		Offset := p.vProfs[Sfm_LibFunId(p, pObj)]
		pProf := p.vStore[Offset:]
		for i := int32(0); i < nFanins; i++ {
			fmt.Printf("%6.2f ", scl.Scl_Int2Flt(pProf[i]))
		}
	}
}
func Sfm_LibPrint(p *Sfm_Lib_t) {
	for i := 0; i < vec.Vec_MemEntryNum(p.vTtMem); i++ {
		pTruth := vec.Vec_MemReadEntry(p.vTtMem, i)
		if i < 2 || p.vHits[i] == 0 {
			continue
		}
		nFanins := util.Abc_TtSupportSize(pTruth, p.nVars)
		fmt.Printf("%8d : ", i)
		fmt.Printf("Num =%5d  ", p.vCounts[i])
		fmt.Printf("Hit =%4d  ", p.vHits[i])
		for pObj := Sfm_LibFun(p, p.vLists[i]); pObj != nil; pObj = Sfm_LibFun(p, pObj.Next) {
			Sfm_LibPrintObj(p, pObj)
			break
		}
		fmt.Printf("    ")
		dau.Dau_DsdPrintFromTruth(pTruth, nFanins)
	}
}
func Sfm_LibTest() {
	fVerbose := int32(1)
	if abcmain.Abc_FrameReadLibGen() == nil {
		fmt.Printf("There is no current library.\n")
		return
	}
	p := Sfm_LibPrepare(7, 1, 1, 1, fVerbose)
	if fVerbose != 0 {
		Sfm_LibPrint(p)
	}
	Sfm_LibStop(p)
}

func Sfm_LibFindAreaMatch(p *Sfm_Lib_t, pTruth []uint64, nFanins int32, piObj *int32) int32 {
	var pObj *Sfm_Fun_t
	iFunc := *vec.Vec_MemHashLookup(p.vTtMem, pTruth[:p.nWords])
	if iFunc == -1 {
		return -1
	}
	pObj = Sfm_LibFun(p, p.vLists[iFunc])
	if piObj != nil {
		*piObj = Sfm_LibFunId(p, pObj)
	}
	return pObj.Area
}

// Sfm_LibFindDelayMatches collects pairs (bottom, top) of gates and their
// fanin permutations; a missing top gate is nil.
func Sfm_LibFindDelayMatches(p *Sfm_Lib_t, pTruth []uint64, pFanins []int32, nFanins int32, vGates *[]*mio.Mio_Gate_t, vFans *[][]byte) int32 {
	if nFanins > 6 {
		var pCopy [4]uint64
		util.Abc_TtCopy(pCopy[:], pTruth, 4, 0)
		dau.Dau_DsdPrintFromTruth(pCopy[:], p.nVars)
	}
	*vGates = (*vGates)[:0]
	*vFans = (*vFans)[:0]
	// look for gate
	assert(util.Abc_TtIsConst0(pTruth, p.nWords) == 0 &&
		util.Abc_TtIsConst1(pTruth, p.nWords) == 0 &&
		util.Abc_TtEqual(pTruth, s_Truth8[0][:], p.nWords) == 0 &&
		util.Abc_TtOpposite(pTruth, s_Truth8[0][:], p.nWords) == 0, "non-trivial function")
	iFunc := *vec.Vec_MemHashLookup(p.vTtMem, pTruth[:p.nWords])
	if iFunc == -1 {
		// print functions not found in the library
		if p.fVerbose != 0 || nFanins > 6 {
			fmt.Printf("Not found in the precomputed library: ")
			dau.Dau_DsdPrintFromTruth(pTruth, nFanins)
		}
		return 0
	}
	p.vHits[iFunc]++
	// collect matches
	for pObj := Sfm_LibFun(p, p.vLists[iFunc]); pObj != nil; pObj = Sfm_LibFun(p, pObj.Next) {
		pCellB := &p.pCells[pObj.pFansB[0]]
		pCellT := &p.pCells[pObj.pFansT[0]]
		*vGates = append(*vGates, pCellB.Gate())
		if pCellT == &p.pCells[0] {
			*vGates = append(*vGates, nil)
		} else {
			*vGates = append(*vGates, pCellT.Gate())
		}
		*vFans = append(*vFans, pObj.pFansB[1:])
		if pCellT == &p.pCells[0] {
			*vFans = append(*vFans, nil)
		} else {
			*vFans = append(*vFans, pObj.pFansT[1:])
		}
	}
	return int32(len(*vGates) / 2)
}

func sfmLibReadLibGen() *mio.Mio_Library_t {
	pLib, _ := abcmain.Abc_FrameReadLibGen().(*mio.Mio_Library_t)
	return pLib
}

func Sfm_LibImplementSimple(p *Sfm_Lib_t, pTruth []uint64, pFanins []int32, nFanins int32, vGates *[]int32, vFanins *[][]int32) int32 {
	pLib := sfmLibReadLibGen()
	var pGate *mio.Mio_Gate_t
	if util.Abc_TtIsConst0(pTruth, p.nWords) != 0 || util.Abc_TtIsConst1(pTruth, p.nWords) != 0 {
		assert(nFanins == 0, "nFanins == 0")
		if util.Abc_TtIsConst1(pTruth, p.nWords) != 0 {
			pGate = mio.Mio_LibraryReadConst1(pLib)
		} else {
			pGate = mio.Mio_LibraryReadConst0(pLib)
		}
		*vGates = append(*vGates, mio.Mio_GateReadValue(pGate))
		*vFanins = append(*vFanins, nil)
		return 1
	}
	if util.Abc_TtEqual(pTruth, s_Truth8[0][:], p.nWords) != 0 || util.Abc_TtOpposite(pTruth, s_Truth8[0][:], p.nWords) != 0 {
		assert(nFanins == 1, "nFanins == 1")
		if util.Abc_TtEqual(pTruth, s_Truth8[0][:], p.nWords) != 0 {
			pGate = mio.Mio_LibraryReadBuf(pLib)
		} else {
			pGate = mio.Mio_LibraryReadInv(pLib)
		}
		*vGates = append(*vGates, mio.Mio_GateReadValue(pGate))
		*vFanins = append(*vFanins, []int32{pFanins[0]})
		return 1
	}
	assert(false, "Sfm_LibImplementSimple")
	return -1
}
func Sfm_LibImplementGatesArea(p *Sfm_Lib_t, pFanins []int32, nFanins, iObj int32, vGates *[]int32, vFanins *[][]int32) int32 {
	pLib := sfmLibReadLibGen()
	pObjMin := &p.pObjs[iObj]
	// get the gates
	pCellB := &p.pCells[pObjMin.pFansB[0]]
	pCellT := &p.pCells[pObjMin.pFansT[0]]
	// create bottom gate
	pGate := mio.Mio_LibraryReadGateByName(pLib, pCellB.Name(), "")
	assert(pGate == pCellB.Gate(), "pGate == pCellB->pMioGate")
	*vGates = append(*vGates, mio.Mio_GateReadValue(pGate))
	var vLevel []int32
	for i := int32(0); i < pCellB.Fanins(); i++ {
		vLevel = append(vLevel, pFanins[pObjMin.pFansB[i+1]])
	}
	*vFanins = append(*vFanins, vLevel)
	if pCellT == &p.pCells[0] {
		return 1
	}
	// create top gate
	pGate = mio.Mio_LibraryReadGateByName(pLib, pCellT.Name(), "")
	assert(pGate == pCellT.Gate(), "pGate == pCellT->pMioGate")
	*vGates = append(*vGates, mio.Mio_GateReadValue(pGate))
	vLevel = nil
	*vFanins = append(*vFanins, nil)
	for i := int32(0); i < pCellT.Fanins(); i++ {
		if pObjMin.pFansT[i+1] == 16 {
			vLevel = append(vLevel, int32(len(*vFanins))-2)
		} else {
			vLevel = append(vLevel, pFanins[pObjMin.pFansT[i+1]])
		}
	}
	(*vFanins)[len(*vFanins)-1] = vLevel
	return 2
}
func Sfm_LibImplementGatesDelay(p *Sfm_Lib_t, pFanins []int32, pGateB, pGateT *mio.Mio_Gate_t, pFansB, pFansT []byte, vGates *[]int32, vFanins *[][]int32) int32 {
	// create bottom gate
	*vGates = append(*vGates, mio.Mio_GateReadValue(pGateB))
	var vLevel []int32
	nFanins := mio.Mio_GateReadPinNum(pGateB)
	for i := int32(0); i < nFanins; i++ {
		vLevel = append(vLevel, pFanins[pFansB[i]])
	}
	*vFanins = append(*vFanins, vLevel)
	if pGateT == nil {
		return 1
	}
	// create top gate
	*vGates = append(*vGates, mio.Mio_GateReadValue(pGateT))
	vLevel = nil
	*vFanins = append(*vFanins, nil)
	nFanins = mio.Mio_GateReadPinNum(pGateT)
	for i := int32(0); i < nFanins; i++ {
		if pFansT[i] == 16 {
			vLevel = append(vLevel, int32(len(*vFanins))-2)
		} else {
			vLevel = append(vLevel, pFanins[pFansT[i]])
		}
	}
	(*vFanins)[len(*vFanins)-1] = vLevel
	return 2
}
