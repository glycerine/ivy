package abc

// Line-by-line port of src/opt/sfm/sfmDec.c, the SAT-based decomposition
// engine of mfs3 (Abc_NtkPerformMfs3). It lives here because, as in C, it
// works directly on the mapped Abc_Ntk_t.

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dau"
	"github.com/glycerine/abcgo/opt/sfm"
	"github.com/glycerine/abcgo/sat/bsat"
)

const (
	sfmSuppMax  = sfm.SFM_SUPP_MAX
	sfmWordMax  = sfm.SFM_WORD_MAX
	sfmWinMax   = sfm.SFM_WIN_MAX
	sfmDecMax   = sfm.SFM_DEC_MAX
	sfmSimWords = sfm.SFM_SIM_WORDS
)

type Sfm_Dec_t struct {
	// external
	pPars *sfm.Sfm_Par_t // parameters
	pLib  *sfm.Sfm_Lib_t // library
	pTim  *Sfm_Tim_t     // timing
	pMit  *Sfm_Mit_t     // timing
	pNtk  *Network       // network
	// library
	vGateSizes []int32           // fanin counts
	vGateFuncs []uint64          // gate truth tables
	vGateCnfs  [][]int8          // gate CNFs
	vGateHands []*mio.Mio_Gate_t // gate handles
	GateConst0 int32             // special gates
	GateConst1 int32             // special gates
	GateBuffer int32             // special gates
	GateInvert int32             // special gates
	GateAnd    [4]int32          // special gates
	GateOr     [4]int32          // special gates
	// objects
	nDivs       int32             // the number of divisors
	nMffc       int32             // the number of divisors
	AreaMffc    int32             // the area of gates in MFFC
	DelayMin    int32             // temporary min delay
	iTarget     int32             // target node
	iUseThis    int32             // next cofactoring var to try
	DeltaCrit   int32             // critical delta
	AreaInv     int32             // inverter area
	DelayInv    int32             // inverter delay
	pGateInv    *mio.Mio_Gate_t   // inverter
	uCareSet    uint64            // computed careset
	vObjRoots   []int32           // roots of the window
	vObjGates   []int32           // functionality
	vObjFanins  [][]int32         // fanin IDs
	vObjMap     []int32           // object map
	vObjDec     []int32           // decomposition
	vObjMffc    []int32           // MFFC nodes
	vObjInMffc  []int32           // inputs of MFFC nodes
	vObjSims    []uint64          // simulation patterns
	vObjSims2   []uint64          // simulation patterns
	vMatchGates []*mio.Mio_Gate_t // matched gates
	vMatchFans  [][]byte          // matched fanins
	// solver
	pSat           *bsat.Solver // reusable solver
	vClauses       [][]int32    // CNF clauses for the node
	vImpls         [2][]int32   // onset/offset implications
	vSets          [2][]uint64  // onset/offset patterns
	nPats          [2]int32     // CEX count
	nPatWords      [2]int32     // CEX words
	nDivWords      int32        // div words
	nDivWordsAlloc int32        // div words
	TtElems        [sfmSuppMax][sfmWordMax]uint64
	pTtElems       [sfmSuppMax][]uint64
	pDivWords      [sfmSuppMax][]uint64
	// temporary
	vNewNodes []int32
	vGateTfi  []int32
	vGateTfo  []int32
	vGateCut  []int32
	vGateTemp []int32
	vGateMffc []int32
	vCands    []int32
	Copy      [4]uint64
	nSuppVars int32
	// statistics
	timeLib        int64
	timeWin        int64
	timeCnf        int64
	timeSat        int64
	timeSatSat     int64
	timeSatUnsat   int64
	timeEval       int64
	timeTime       int64
	timeOther      int64
	timeStart      int64
	timeTotal      int64
	nTotalNodesBeg int32
	nTotalEdgesBeg int32
	nTotalNodesEnd int32
	nTotalEdgesEnd int32
	nNodesTried    int32
	nNodesChanged  int32
	nNodesConst0   int32
	nNodesConst1   int32
	nNodesBuf      int32
	nNodesInv      int32
	nNodesAndOr    int32
	nNodesResyn    int32
	nSatCalls      int32
	nSatCallsSat   int32
	nSatCallsUnsat int32
	nSatCallsOver  int32
	nTimeOuts      int32
	nNoDecs        int32
	nEfforts       int32
	nMaxDivs       int32
	nMaxWin        int32
	nAllDivs       uint64
	nAllWin        uint64
	nLuckySizes    [sfmSuppMax + 1]int32
	nLuckyGates    [sfmSuppMax + 1]int32
}

const (
	SFM_MASK_PI    = 1  // supp(node) is contained in supp(TFI(pivot))
	SFM_MASK_INPUT = 2  // supp(node) does not overlap with supp(TFI(pivot))
	SFM_MASK_FANIN = 4  // the same as above (pointed to by node with SFM_MASK_PI | SFM_MASK_INPUT)
	SFM_MASK_MFFC  = 8  // MFFC nodes, including the target node
	SFM_MASK_PIVOT = 16 // the target node
)

func Sfm_DecMan(p *Object) *Sfm_Dec_t { return p.network.Data.(*Sfm_Dec_t) }
func Sfm_DecObjSim(p *Sfm_Dec_t, pObj *Object) uint64 {
	return p.vObjSims[Abc_ObjId(pObj)]
}
func Sfm_DecObjSim2(p *Sfm_Dec_t, pObj *Object) uint64 {
	return p.vObjSims2[Abc_ObjId(pObj)]
}
func Sfm_DecDivPats(p *Sfm_Dec_t, d, c int32) []uint64 {
	return p.vSets[c][d*sfmSimWords : (d+1)*sfmSimWords]
}

func Sfm_ManReadObjDelay(p *Sfm_Dec_t, Id int32) int32 {
	if p.pMit != nil {
		return Sfm_MitReadObjDelay(p.pMit, Id)
	}
	return Sfm_TimReadObjDelay(p.pTim, Id)
}
func Sfm_ManReadNtkDelay(p *Sfm_Dec_t) int32 {
	if p.pMit != nil {
		return Sfm_MitReadNtkDelay(p.pMit)
	}
	return Sfm_TimReadNtkDelay(p.pTim)
}
func Sfm_ManReadNtkMinSlack(p *Sfm_Dec_t) int32 {
	if p.pMit != nil {
		return Sfm_MitReadNtkMinSlack(p.pMit)
	}
	return 0
}

func sfmLibGen() any { return abcmain.Abc_FrameReadLibGen() }

func sfmB2I(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func sfmMioGate(pObj *Object) *mio.Mio_Gate_t { return pObj.Data.(*mio.Mio_Gate_t) }

func sfmIntFind(v []int32, Entry int32) int32 {
	for i, e := range v {
		if e == Entry {
			return int32(i)
		}
	}
	return -1
}
func sfmIntPushUnique(v *[]int32, Entry int32) {
	if sfmIntFind(*v, Entry) >= 0 {
		return
	}
	*v = append(*v, Entry)
}

func Sfm_DecStart(pPars *sfm.Sfm_Par_t, pLib *mio.Mio_Library_t, pNtk *Network) *Sfm_Dec_t {
	p := new(Sfm_Dec_t)
	p.timeStart = util.Abc_Clock()
	p.pPars = pPars
	p.pNtk = pNtk
	p.pSat = bsat.NewSolver()
	p.pGateInv = mio.Mio_LibraryReadInv(pLib)
	p.AreaInv = scl.Scl_Flt2Int(float32(mio.Mio_GateReadArea(p.pGateInv)))
	p.DelayInv = scl.Scl_Flt2Int(float32(mio.Mio_GateReadDelayMax(p.pGateInv)))
	if pPars.DeltaCrit != 0 {
		p.DeltaCrit = scl.Scl_Flt2Int(float32(pPars.DeltaCrit))
	} else {
		p.DeltaCrit = 5 * scl.Scl_Flt2Int(mio.Mio_LibraryReadDelayInvMax(pLib)) / 2
	}
	p.timeLib = util.Abc_Clock()
	p.pLib = sfm.Sfm_LibPrepare(int32(pPars.VarMax), 1, sfmB2I(!pPars.Area), sfmB2I(pPars.Verbose), sfmB2I(pPars.LibVerbose))
	p.timeLib = util.Abc_Clock() - p.timeLib
	if !pPars.Area {
		if abcmain.Abc_FrameReadLibScl() != nil {
			p.pMit = Sfm_MitStart(pLib, abcmain.Abc_FrameReadLibScl(), scl.Scl_ConReadMan(), pNtk, p.DeltaCrit)
		}
		if p.pMit == nil {
			p.pTim = Sfm_TimStart(pLib, scl.Scl_ConReadMan(), pNtk, p.DeltaCrit)
		}
	}
	if pPars.VeryVerbose {
		//    if ( pPars->fVerbose )
		sfm.Sfm_LibPrint(p.pLib)
	}
	pNtk.Data = p
	// enter library
	assert(Abc_NtkIsMappedLogic(pNtk) != 0, "Abc_NtkIsMappedLogic(pNtk)")
	sfm.Sfm_LibPreprocess(pLib, &p.vGateSizes, &p.vGateFuncs, &p.vGateCnfs, &p.vGateHands)
	p.GateConst0 = mio.Mio_GateReadValue(mio.Mio_LibraryReadConst0(pLib))
	p.GateConst1 = mio.Mio_GateReadValue(mio.Mio_LibraryReadConst1(pLib))
	p.GateBuffer = mio.Mio_GateReadValue(mio.Mio_LibraryReadBuf(pLib))
	p.GateInvert = mio.Mio_GateReadValue(mio.Mio_LibraryReadInv(pLib))
	// elementary truth tables
	for i := 0; i < sfmSuppMax; i++ {
		p.pTtElems[i] = p.TtElems[i][:]
	}
	util.Abc_TtElemInit(p.pTtElems[:], sfmSuppMax)
	p.iUseThis = -1
	return p
}
func Sfm_DecStop(p *Sfm_Dec_t) {
	pNtk := p.pNtk
	for i := int32(0); i < Abc_NtkObjNumMax(pNtk); i++ {
		pObj := Abc_NtkObj(pNtk, i)
		if pObj == nil || Abc_ObjIsNode(pObj) == 0 {
			continue
		}
		if int32(pObj.Level) != Abc_ObjLevelNew(pObj) {
			fmt.Printf("Level count mismatch at node %d.\n", i)
		}
	}
	sfm.Sfm_LibStop(p.pLib)
	if p.pTim != nil {
		Sfm_TimStop(p.pTim)
	}
	if p.pMit != nil {
		Sfm_MitStop(p.pMit)
	}
	p.pSat.Delete()
	*p = Sfm_Dec_t{}
	pNtk.Data = nil
}

func Sfm_ObjSimulate(pObj *Object) uint64 {
	p := Sfm_DecMan(pObj)
	vExpr := mio.Mio_GateReadExpr(sfmMioGate(pObj))
	var uFanins [6]uint64
	assert(Abc_ObjFaninNum(pObj) <= 6, "Abc_ObjFaninNum(pObj) <= 6")
	for i, e := range pObj.fanins {
		uFanins[i] = Sfm_DecObjSim(p, e.Object)
	}
	return mio.Exp_Truth6(Abc_ObjFaninNum(pObj), vExpr, uFanins[:])
}
func Sfm_ObjSimulate2(pObj *Object) uint64 {
	p := Sfm_DecMan(pObj)
	vExpr := mio.Mio_GateReadExpr(sfmMioGate(pObj))
	var uFanins [6]uint64
	for i, e := range pObj.fanins {
		pFanin := e.Object
		if pFanin.Temp&SFM_MASK_PIVOT != 0 {
			uFanins[i] = Sfm_DecObjSim2(p, pFanin)
		} else {
			uFanins[i] = Sfm_DecObjSim(p, pFanin)
		}
	}
	return mio.Exp_Truth6(Abc_ObjFaninNum(pObj), vExpr, uFanins[:])
}
func sfmWrdFill(v []uint64, nSize int, Fill uint64) []uint64 {
	if cap(v) < nSize {
		v = make([]uint64, nSize)
	}
	v = v[:nSize]
	for i := range v {
		v[i] = Fill
	}
	return v
}
func Sfm_NtkSimulate(pNtk *Network) {
	p := Sfm_DecMan(Abc_NtkPi(pNtk, 0))
	p.vObjSims = sfmWrdFill(p.vObjSims, 2*int(Abc_NtkObjNumMax(pNtk)), 0)
	p.vObjSims2 = sfmWrdFill(p.vObjSims2, 2*int(Abc_NtkObjNumMax(pNtk)), 0)
	gia.Gia_ManRandomW(1)
	assert(p.pPars.UseSim, "p->pPars->fUseSim")
	for i := int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pObj := Abc_NtkCi(pNtk, i)
		p.vObjSims[Abc_ObjId(pObj)] = gia.Gia_ManRandomW(0)
	}
	vNodes := Abc_NtkDfs(pNtk, 1)
	for _, pObj := range vNodes {
		p.vObjSims[Abc_ObjId(pObj)] = Sfm_ObjSimulate(pObj)
	}
}
func Sfm_ObjSimulateNode(pObj *Object) {
	p := Sfm_DecMan(pObj)
	if !p.pPars.UseSim {
		return
	}
	p.vObjSims[Abc_ObjId(pObj)] = Sfm_ObjSimulate(pObj)
	if pObj.Temp&SFM_MASK_PIVOT != 0 {
		p.vObjSims2[Abc_ObjId(pObj)] = Sfm_ObjSimulate2(pObj)
	}
}
func Sfm_ObjFlipNode(pObj *Object) {
	p := Sfm_DecMan(pObj)
	if !p.pPars.UseSim {
		return
	}
	p.vObjSims2[Abc_ObjId(pObj)] = ^Sfm_DecObjSim(p, pObj)
}
func Sfm_ObjFindCareSet(pNtk *Network, vRoots []int32) uint64 {
	p := Sfm_DecMan(Abc_NtkPi(pNtk, 0))
	Res := uint64(0)
	if !p.pPars.UseSim {
		return 0
	}
	for _, id := range vRoots {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		Res |= Sfm_DecObjSim(p, pObj) ^ Sfm_DecObjSim2(p, pObj)
	}
	return Res
}
func Sfm_ObjSetupSimInfo(pObj *Object) {
	p := Sfm_DecMan(pObj)
	p.nPats[0], p.nPats[1] = 0, 0
	p.nPatWords[0], p.nPatWords[1] = 0, 0
	p.vSets[0] = sfmWrdFill(p.vSets[0], int(p.nDivs)*sfmSimWords, 0)
	p.vSets[1] = sfmWrdFill(p.vSets[1], int(p.nDivs)*sfmSimWords, 0)
	// alloc divwords
	p.nDivWords = util.Abc_Bit6WordNum(4 * p.nDivs)
	if p.nDivWordsAlloc < p.nDivWords {
		p.nDivWordsAlloc = max(16, p.nDivWords)
		for i := 0; i < sfmSuppMax; i++ {
			grown := make([]uint64, p.nDivWordsAlloc)
			copy(grown, p.pDivWords[i])
			p.pDivWords[i] = grown
		}
	}
	clear(p.pDivWords[0][:p.nDivWords])
	// collect simulation info
	if p.pPars.UseSim && p.uCareSet != 0 {
		uCareSet := p.uCareSet
		uValues := Sfm_DecObjSim(p, pObj)
		var Indexes [2][64]int32
		assert(p.iTarget == int32(pObj.Temp), "p->iTarget == pObj->iTemp")
		assert(p.pPars.UseSim, "p->pPars->fUseSim")
		// find what patterns go to on-set/off-set
		for i := 0; i < 64; i++ {
			if (uCareSet>>uint(i))&1 != 0 {
				c := 1 - int32((uValues>>uint(i))&1)
				Indexes[c][p.nPats[c]] = int32(i)
				p.nPats[c]++
			}
		}
		for c := 0; c < 2; c++ {
			p.nPatWords[c] = 1 + (p.nPats[c] >> 6)
		}
		// write patterns
		for d := int32(0); d < p.nDivs; d++ {
			uSim := p.vObjSims[p.vObjMap[d]]
			for c := int32(0); c < 2; c++ {
				for i := int32(0); i < p.nPats[c]; i++ {
					if (uSim>>uint(Indexes[c][i]))&1 != 0 {
						util.Abc_TtSetBit(Sfm_DecDivPats(p, d, c), i)
					}
				}
			}
		}
		//printf( "Node %d : Onset = %d. Offset = %d.\n", pObj->Id, p->nPats[0], p->nPats[1] );
	}
}
func Sfm_ObjSetdownSimInfo(pObj *Object) {
	nPatKeep := int32(32)
	p := Sfm_DecMan(pObj)
	var uSims [2]uint64
	if !p.pPars.UseSim {
		return
	}
	for d := int32(0); d < p.nDivs; d++ {
		uSim := p.vObjSims[p.vObjMap[d]]
		for c := int32(0); c < 2; c++ {
			uMask := util.Abc_Tt6Mask(min(p.nPats[c], nPatKeep))
			uSims[c] = (Sfm_DecDivPats(p, d, c)[0] & uMask) | (uSim & ^uMask)
			uSim >>= 32
		}
		uSim = (uSims[0] & 0xFFFFFFFF) | (uSims[1] << 32)
		p.vObjSims[p.vObjMap[d]] = uSim
	}
}

func Sfm_DecPrepareSolver(p *Sfm_Dec_t) int32 {
	vRoots := p.vObjRoots
	nTfiSize := p.iTarget + 1 // including node
	nWinSize := int32(len(p.vObjGates))
	nSatVars := 2*nWinSize - nTfiSize
	assert(nWinSize == int32(len(p.vObjGates)), "nWinSize == Vec_IntSize(&p->vObjGates)")
	assert(p.iTarget < nWinSize, "p->iTarget < nWinSize")
	// create SAT solver
	p.pSat.RestartSearch()
	p.pSat.SetnvarsC(nSatVars + int32(len(vRoots)))
	// add CNF clauses for the TFI
	for i, Gate := range p.vObjGates {
		if Gate == -1 {
			continue
		}
		// generate CNF
		vLevel := append(p.vGateTemp[:0], p.vObjFanins[i]...)
		vLevel = append(vLevel, int32(i))
		p.vGateTemp = vLevel
		sfm.Sfm_TranslateCnf(&p.vClauses, p.vGateCnfs[Gate], vLevel, -1)
		// add clauses
		for _, vClause := range p.vClauses {
			if len(vClause) == 0 {
				break
			}
			if p.pSat.AddClauseC(vClause) == 0 {
				return 0
			}
		}
	}
	// add CNF clauses for the TFO
	for i := nTfiSize; i < int32(len(p.vObjGates)); i++ {
		Gate := p.vObjGates[i]
		assert(Gate != -1, "Gate != -1")
		vFaninVars := p.vGateTemp[:0]
		for _, iObj := range p.vObjFanins[i] {
			if iObj <= p.iTarget {
				vFaninVars = append(vFaninVars, iObj)
			} else {
				vFaninVars = append(vFaninVars, iObj+nWinSize-nTfiSize)
			}
		}
		vFaninVars = append(vFaninVars, i+nWinSize-nTfiSize)
		p.vGateTemp = vFaninVars
		// generate CNF
		sfm.Sfm_TranslateCnf(&p.vClauses, p.vGateCnfs[Gate], vFaninVars, p.iTarget)
		// add clauses
		for _, vClause := range p.vClauses {
			if len(vClause) == 0 {
				break
			}
			if p.pSat.AddClauseC(vClause) == 0 {
				return 0
			}
		}
	}
	if nTfiSize < nWinSize {
		// create XOR clauses for the roots
		vFaninVars := p.vGateTemp[:0]
		for _, iObj := range vRoots {
			vFaninVars = append(vFaninVars, nSatVars*2) // Abc_Var2Lit(nSatVars, 0)
			p.pSat.AddXor(iObj, iObj+nWinSize-nTfiSize, nSatVars, false)
			nSatVars++
		}
		p.vGateTemp = vFaninVars
		// make OR clause for the last nRoots variables
		if p.pSat.AddClauseC(vFaninVars) == 0 {
			return 0
		}
		assert(nSatVars == p.pSat.NVarsC(), "nSatVars == sat_solver_nvars(p->pSat)")
	} else {
		assert(len(vRoots) == 1, "Vec_IntSize(vRoots) == 1")
	}
	// finalize
	p.pSat.SimplifyC()
	return 1
}

func Sfm_DecFindCost(p *Sfm_Dec_t, c int32, iLit int32, pMask []uint64) int32 {
	pPats := Sfm_DecDivPats(p, iLit>>1, 1-c)
	return util.Abc_TtCountOnesVecMask(pPats, pMask, p.nPatWords[1-c], iLit&1)
}
func Sfm_DecPrint(p *Sfm_Dec_t, Masks *[2][sfmSimWords]uint64) {
	for c := int32(0); c < 2; c++ {
		vLevel := p.vObjFanins[p.iTarget]
		set := "ON"
		if c != 0 {
			set = "OFF"
		}
		fmt.Printf("%s-SET of object %d (divs = %d) with gate \"%s\" and fanins: ",
			set, p.iTarget, p.nDivs,
			mio.Mio_GateReadName(p.vGateHands[p.vObjGates[p.iTarget]]))
		for _, Entry := range vLevel {
			fmt.Printf("%d ", Entry)
		}
		fmt.Printf("\n")

		fmt.Printf("Implications: ")
		for _, Entry := range p.vImpls[c] {
			neg := ""
			if Entry&1 != 0 {
				neg = "!"
			}
			fmt.Printf("%s%d(%d) ", neg, Entry>>1, Sfm_DecFindCost(p, c, Entry, Masks[1-c][:]))
		}
		fmt.Printf("\n")
		fmt.Printf("     ")
		for i := int32(0); i < p.nDivs; i++ {
			fmt.Printf("%d", (i/10)%10)
		}
		fmt.Printf("\n")
		fmt.Printf("     ")
		for i := int32(0); i < p.nDivs; i++ {
			fmt.Printf("%d", i%10)
		}
		fmt.Printf("\n")
		for k := int32(0); k < p.nPats[c]; k++ {
			fmt.Printf("%2d : ", k)
			for i := int32(0); i < p.nDivs; i++ {
				fmt.Printf("%d", util.Abc_TtGetBit(Sfm_DecDivPats(p, i, c), k))
			}
			fmt.Printf("\n")
		}
		//printf( "\n" );
	}
}

func Sfm_DecVarCost(p *Sfm_Dec_t, Masks *[2][sfmSimWords]uint64, d int32, Counts *[2][2]int32) {
	for c := int32(0); c < 2; c++ {
		pPats := Sfm_DecDivPats(p, d, c)
		Num := util.Abc_TtCountOnesVec(Masks[c][:], p.nPatWords[c])
		Counts[c][1] = util.Abc_TtCountOnesVecMask(pPats, Masks[c][:], p.nPatWords[c], 0)
		Counts[c][0] = Num - Counts[c][1]
		assert(Counts[c][0] >= 0 && Counts[c][1] >= 0, "Counts")
	}
}
func Sfm_DecFindBestVar2(p *Sfm_Dec_t, Masks *[2][sfmSimWords]uint64) int32 {
	var Counts [2][2]int32
	VarBest, CostBest := int32(-1), int32(sfmAbcInfinity)
	for d := int32(0); d < p.nDivs; d++ {
		Sfm_DecVarCost(p, Masks, d, &Counts)
		if (Counts[0][0] < Counts[0][1]) == (Counts[1][0] < Counts[1][1]) {
			continue
		}
		Cost := min(Counts[0][0], Counts[0][1]) + min(Counts[1][0], Counts[1][1])
		if CostBest > Cost {
			CostBest = Cost
			VarBest = d
		}
	}
	return VarBest
}
func Sfm_DecFindBestVar(p *Sfm_Dec_t, Masks *[2][sfmSimWords]uint64) int32 {
	Var, CostMin := int32(-1), int32(sfmAbcInfinity)
	for c := int32(0); c < 2; c++ {
		for _, iLit := range p.vImpls[c] {
			if len(p.vImpls[c]) > 1 && sfmIntFind(p.vObjDec, iLit>>1) >= 0 {
				continue
			}
			Cost := Sfm_DecFindCost(p, c, iLit, Masks[1-c][:])
			if CostMin > Cost {
				CostMin = Cost
				Var = iLit >> 1
			}
		}
	}
	return Var
}

func sfmGateArea(pObj *Object) int32 {
	return scl.Scl_Flt2Int(float32(mio.Mio_GateReadArea(sfmMioGate(pObj))))
}

func Sfm_DecMffcArea(pNtk *Network, vMffc []int32) int32 {
	nAreaMffc := int32(0)
	for _, id := range vMffc {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		nAreaMffc += sfmGateArea(pObj)
	}
	return nAreaMffc
}
func Sfm_MffcDeref_rec(pObj *Object) int32 {
	Area := sfmGateArea(pObj)
	for _, e := range pObj.fanins {
		pFanin := e.Object
		assert(Abc_ObjFanoutNum(pFanin) > 0, "pFanin->vFanouts.nSize > 0")
		Abc_ObjAdjustFanoutCount(pFanin, -1)
		if Abc_ObjFanoutNum(pFanin) == 0 && Abc_ObjIsCi(pFanin) == 0 {
			Area += Sfm_MffcDeref_rec(pFanin)
		}
	}
	return Area
}
func Sfm_MffcRef_rec(pObj *Object, vMffc *[]int32) int32 {
	Area := sfmGateArea(pObj)
	for _, e := range pObj.fanins {
		pFanin := e.Object
		wasZero := Abc_ObjFanoutNum(pFanin) == 0
		Abc_ObjAdjustFanoutCount(pFanin, 1)
		if wasZero && Abc_ObjIsCi(pFanin) == 0 {
			Area += Sfm_MffcRef_rec(pFanin, vMffc)
		}
	}
	if vMffc != nil {
		*vMffc = append(*vMffc, Abc_ObjId(pObj))
	}
	return Area
}
func Sfm_DecMffcAreaReal(pPivot *Object, vCut []int32, vMffc *[]int32) int32 {
	pNtk := pPivot.network
	assert(Abc_ObjIsNode(pPivot) != 0, "Abc_ObjIsNode(pPivot)")
	if vMffc != nil {
		*vMffc = (*vMffc)[:0]
	}
	for _, id := range vCut {
		if pObj := Abc_NtkObj(pNtk, id); pObj != nil {
			Abc_ObjAdjustFanoutCount(pObj, 1)
		}
	}
	Area1 := Sfm_MffcDeref_rec(pPivot)
	Area2 := Sfm_MffcRef_rec(pPivot, vMffc)
	for _, id := range vCut {
		if pObj := Abc_NtkObj(pNtk, id); pObj != nil {
			Abc_ObjAdjustFanoutCount(pObj, -1)
		}
	}
	assert(Area1 == Area2, "Area1 == Area2")
	return Area1
}
func Sfm_DecPrepareVec(vMap []int32, pNodes []int32, nNodes int32, vCut *[]int32) {
	*vCut = (*vCut)[:0]
	for i := int32(0); i < nNodes; i++ {
		*vCut = append(*vCut, vMap[pNodes[i]])
	}
}
func Sfm_DecComputeFlipInvGain(p *Sfm_Dec_t, pPivot *Object, pfNeedInv *int32) int32 {
	fNeedInv, Gain := int32(0), int32(0)
	for _, pFanout := range pPivot.fanouts {
		if Abc_ObjIsNode(pFanout) == 0 {
			fNeedInv = 1
			continue
		}
		pGate := sfmMioGate(pFanout)
		if Abc_ObjFaninNum(pFanout) == 1 && mio.Mio_GateIsInv(pGate) != 0 {
			Gain += p.AreaInv
			continue
		}
		Handle := sfm.Sfm_LibFindComplInputGate(p.vGateFuncs, mio.Mio_GateReadValue(pGate), Abc_ObjFaninNum(pFanout), Abc_NodeFindFanin(pFanout, pPivot), nil)
		if Handle == -1 {
			fNeedInv = 1
			continue
		}
		pGateNew := p.vGateHands[Handle]
		Gain += scl.Scl_Flt2Int(float32(mio.Mio_GateReadArea(pGate))) - scl.Scl_Flt2Int(float32(mio.Mio_GateReadArea(pGateNew)))
	}
	if fNeedInv != 0 {
		Gain -= p.AreaInv
	}
	if pfNeedInv != nil {
		*pfNeedInv = fNeedInv
	}
	return Gain
}

// sfmIntTwoMerge2Int is Vec_IntTwoMerge2Int on arrays.
func sfmIntTwoMerge2Int(a1, a2 []int32, out []int32) int32 {
	n, i1, i2 := 0, 0, 0
	for i1 < len(a1) && i2 < len(a2) {
		if a1[i1] == a2[i2] {
			out[n] = a1[i1]
			n++
			i1++
			i2++
		} else if a1[i1] < a2[i2] {
			out[n] = a1[i1]
			n++
			i1++
		} else {
			out[n] = a2[i2]
			n++
			i2++
		}
	}
	for i1 < len(a1) {
		out[n] = a1[i1]
		n++
		i1++
	}
	for i2 < len(a2) {
		out[n] = a2[i2]
		n++
		i2++
	}
	assert(n >= len(a1) && n >= len(a2), "merge size")
	return int32(n)
}

// sfmIntPushOrder is Vec_IntPushOrder on an array of size n (capacity
// permitting); it returns the new size.
func sfmIntPushOrder(arr []int32, n int32, Entry int32) int32 {
	n++
	i := n - 2
	for ; i >= 0; i-- {
		if arr[i] > Entry {
			arr[i+1] = arr[i]
		} else {
			break
		}
	}
	arr[i+1] = Entry
	return n
}

func Sfm_DecCombineDec(p *Sfm_Dec_t, pTruth0, pTruth1 []uint64, pSupp0, pSupp1 []int32, nSupp0, nSupp1 int32, pTruth []uint64, pSupp []int32, Var int32) int32 {
	nWords0 := util.Abc_TtWordNum(nSupp0)
	// check the case of equal cofactors
	if nSupp0 == nSupp1 && sfmIntsEqual(pSupp0[:nSupp0], pSupp1[:nSupp0]) && sfmWordsEqual(pTruth0[:nWords0], pTruth1[:nWords0]) {
		copy(pSupp[:nSupp0], pSupp0[:nSupp0])
		copy(pTruth[:nWords0], pTruth0[:nWords0])
		util.Abc_TtStretch6(pTruth, nSupp0, int32(p.pPars.VarMax))
		return nSupp0
	}
	// merge support variables
	nSupp := sfmIntTwoMerge2Int(pSupp0[:nSupp0], pSupp1[:nSupp1], pSupp)
	nSupp = sfmIntPushOrder(pSupp, nSupp, Var)
	if nSupp > int32(p.pPars.VarMax) {
		return -2
	}
	// expand truth tables
	util.Abc_TtStretch6(pTruth0, nSupp0, nSupp)
	util.Abc_TtStretch6(pTruth1, nSupp1, nSupp)
	sfmTtExpand(pTruth0, nSupp, pSupp0, nSupp0, pSupp, nSupp)
	sfmTtExpand(pTruth1, nSupp, pSupp1, nSupp1, pSupp, nSupp)
	// perform operation
	iSuppVar := sfmIntFind(pSupp[:nSupp], Var)
	util.Abc_TtMux(pTruth, p.pTtElems[iSuppVar], pTruth1, pTruth0, util.Abc_TtWordNum(nSupp))
	util.Abc_TtStretch6(pTruth, nSupp, int32(p.pPars.VarMax))
	return nSupp
}

// sfmTtExpand is Abc_TtExpand as release builds of ABC run it: without
// its asserts. With -m (AND/OR detection) the supports are not sorted, so
// pCut is not always a superset of pCut0.
func sfmTtExpand(pTruth0 []uint64, nVars int32, pCut0 []int32, nCutSize0 int32, pCut []int32, nCutSize int32) {
	k := nCutSize0 - 1
	for i := nCutSize - 1; i >= 0 && k >= 0; i-- {
		if pCut[i] > pCut0[k] {
			continue
		}
		//assert( pCut[i] == pCut0[k] );
		if k < i {
			util.Abc_TtSwapVars(pTruth0, nVars, k, i)
		}
		k--
	}
	//assert( k == -1 );
}

func sfmIntsEqual(a, b []int32) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func sfmWordsEqual(a, b []uint64) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (p *Sfm_Dec_t) solve(pAssump []int32) int32 {
	return p.pSat.SolveC(pAssump, p.pPars.BTLimit, 0, 0, 0)
}

const (
	sfm_l_True  = 1
	sfm_l_False = -1
	sfm_l_Undef = 0
)

func Sfm_DecPeformDec_rec(p *Sfm_Dec_t, pTruth []uint64, pSupp []int32, pAssump []int32, nAssump int32, Masks *[2][sfmSimWords]uint64, fCofactor int32, nSuppAdd int32) int32 {
	var status, Var int32 = 0, -1
	pDivWords := p.pDivWords[nAssump]
	var clk int64
	assert(nAssump <= sfmSuppMax, "nAssump <= SFM_SUPP_MAX")
	if p.pPars.VeryVerbose {
		fmt.Printf("\nObject %d\n", p.iTarget)
		fmt.Printf("Divs = %d.  Nodes = %d.  Mffc = %d.  Mffc area = %.2f.    ", p.nDivs, len(p.vObjGates), p.nMffc, scl.Scl_Int2Flt(p.AreaMffc))
		fmt.Printf("Pat0 = %d.  Pat1 = %d.    ", p.nPats[0], p.nPats[1])
		fmt.Printf("\n")
		if nAssump != 0 {
			fmt.Printf("Cofactor: ")
			for i := int32(0); i < nAssump; i++ {
				neg := ""
				if pAssump[i]&1 != 0 {
					neg = "!"
				}
				fmt.Printf(" %s%d", neg, pAssump[i]>>1)
			}
			fmt.Printf("\n")
		}
	}
	// check constant
	for c := int32(0); c < 2; c++ {
		if util.Abc_TtIsConst0(Masks[c][:], p.nPatWords[c]) == 0 { // there are some patterns
			continue
		}
		p.nSatCalls++
		pAssump[nAssump] = 2*p.iTarget + c
		clk = util.Abc_Clock()
		status = p.solve(pAssump[:nAssump+1])
		if status == sfm_l_Undef {
			p.nTimeOuts++
			return -2
		}
		if status == sfm_l_False {
			p.nSatCallsUnsat++
			p.timeSatUnsat += util.Abc_Clock() - clk
			util.Abc_TtConst(pTruth, util.Abc_TtWordNum(int32(p.pPars.VarMax)), c)
			if p.pPars.VeryVerbose {
				fmt.Printf("Found constant %d.\n", c)
			}
			return 0
		}
		assert(status == sfm_l_True, "status == l_True")
		p.nSatCallsSat++
		p.timeSatSat += util.Abc_Clock() - clk
		if p.nPats[c] == 64*sfmSimWords {
			p.nSatCallsOver++
			continue //return -2;//continue;
		}
		for i := int32(0); i < p.nDivs; i++ {
			if p.pSat.VarValueC(i) != 0 {
				util.Abc_TtSetBit(Sfm_DecDivPats(p, i, c), p.nPats[c])
			}
		}
		p.nPatWords[c] = 1 + (p.nPats[c] >> 6)
		util.Abc_TtSetBit(Masks[c][:], p.nPats[c])
		p.nPats[c]++
	}

	if p.iUseThis != -1 {
		Var = p.iUseThis
		p.iUseThis = -1
		goto cofactor
	}

	// check implications
	p.vImpls[0] = p.vImpls[0][:0]
	p.vImpls[1] = p.vImpls[1][:0]
	for d := int32(0); d < p.nDivs; d++ {
		Impls := [2]int32{-1, -1}
		for c := int32(0); c < 2; c++ {
			pPats := Sfm_DecDivPats(p, d, c)
			fHas0s := util.Abc_TtIntersect(pPats, Masks[c][:], p.nPatWords[c], 1)
			fHas1s := util.Abc_TtIntersect(pPats, Masks[c][:], p.nPatWords[c], 0)
			if fHas0s != 0 && fHas1s != 0 {
				continue
			}
			fHas1 := sfmB2I(fHas1s != 0)
			pAssump[nAssump] = 2*p.iTarget + c
			pAssump[nAssump+1] = 2*d + fHas1 // if there are 1s, check if 0 is SAT
			clk = util.Abc_Clock()
			if util.Abc_TtGetBit(pDivWords, 4*d+2*c+fHas1) != 0 {
				p.nSatCallsUnsat--
				status = sfm_l_False
			} else {
				p.nSatCalls++
				status = p.solve(pAssump[:nAssump+2])
			}
			if status == sfm_l_Undef {
				p.nTimeOuts++
				return -2
			}
			if status == sfm_l_False {
				p.nSatCallsUnsat++
				p.timeSatUnsat += util.Abc_Clock() - clk
				Impls[c] = pAssump[nAssump+1] ^ 1
				p.vImpls[c] = append(p.vImpls[c], pAssump[nAssump+1]^1)
				util.Abc_TtSetBit(pDivWords, 4*d+2*c+fHas1)
				continue
			}
			assert(status == sfm_l_True, "status == l_True")
			p.nSatCallsSat++
			p.timeSatSat += util.Abc_Clock() - clk
			if p.nPats[c] == 64*sfmSimWords {
				p.nSatCallsOver++
				continue //return -2;//continue;
			}
			// record this status
			for i := int32(0); i < p.nDivs; i++ {
				if p.pSat.VarValueC(i) != 0 {
					util.Abc_TtSetBit(Sfm_DecDivPats(p, i, c), p.nPats[c])
				}
			}
			p.nPatWords[c] = 1 + (p.nPats[c] >> 6)
			util.Abc_TtSetBit(Masks[c][:], p.nPats[c])
			p.nPats[c]++
		}
		if Impls[0] == -1 || Impls[1] == -1 {
			continue
		}
		if Impls[0] == Impls[1] {
			p.vImpls[0] = p.vImpls[0][:len(p.vImpls[0])-1]
			p.vImpls[1] = p.vImpls[1][:len(p.vImpls[1])-1]
			continue
		}
		assert(Impls[0]>>1 == Impls[1]>>1, "Abc_Lit2Var(Impls[0]) == Abc_Lit2Var(Impls[1])")
		// found buffer/inverter
		util.Abc_TtUnit(pTruth, util.Abc_TtWordNum(int32(p.pPars.VarMax)), Impls[0]&1)
		pSupp[0] = Impls[0] >> 1
		if p.pPars.VeryVerbose {
			neg := ""
			if Impls[0]&1 != 0 {
				neg = "!"
			}
			fmt.Printf("Found variable %s%d.\n", neg, pSupp[0])
		}
		return 1
	}
	if nSuppAdd > int32(p.pPars.VarMax)-2 {
		if p.pPars.VeryVerbose {
			fmt.Printf("The number of assumption is more than MFFC size.\n")
		}
		return -2
	}
	// try using all implications at once
	if p.pPars.UseAndOr {
		for c := int32(0); c < 2; c++ {
			if len(p.vImpls[1-c]) < 2 {
				continue
			}
			p.nSatCalls++
			pAssump[nAssump] = 2*p.iTarget + c
			assert(len(p.vImpls[1-c]) < sfmWinMax-10, "Vec_IntSize(&p->vImpls[!c]) < SFM_WIN_MAX-10")
			i := int32(0)
			for _, iLit := range p.vImpls[1-c] {
				pAssump[nAssump+1+i] = iLit
				i++
			}
			clk = util.Abc_Clock()
			status = p.solve(pAssump[:nAssump+1+i])
			if status == sfm_l_Undef {
				p.nTimeOuts++
				return -2
			}
			if status == sfm_l_False {
				pFinal := p.pSat.FinalC()
				nFinal := int32(len(pFinal))
				p.nSatCallsUnsat++
				p.timeSatUnsat += util.Abc_Clock() - clk
				if nFinal+nSuppAdd > 6 {
					continue
				}
				// collect only relevant literals
				d := int32(0)
				for i = 0; i < nFinal; i++ {
					if sfmIntFind(p.vImpls[1-c], pFinal[i]^1) >= 0 {
						pSupp[d] = pFinal[i] ^ 1
						d++
					}
				}
				nFinal = d
				// create AND/OR gate
				assert(nFinal <= 6, "nFinal <= 6")
				if c != 0 {
					pTruth[0] = ^uint64(0)
					for i = 0; i < nFinal; i++ {
						if pSupp[i]&1 != 0 {
							pTruth[0] &= ^sfmTruths6[i]
						} else {
							pTruth[0] &= sfmTruths6[i]
						}
						pSupp[i] = pSupp[i] >> 1
					}
				} else {
					pTruth[0] = 0
					for i = 0; i < nFinal; i++ {
						if pSupp[i]&1 != 0 {
							pTruth[0] |= sfmTruths6[i]
						} else {
							pTruth[0] |= ^sfmTruths6[i]
						}
						pSupp[i] = pSupp[i] >> 1
					}
				}
				util.Abc_TtStretch6(pTruth, nFinal, int32(p.pPars.VarMax))
				p.nNodesAndOr++
				if p.pPars.VeryVerbose {
					fmt.Printf("Found %d-input AND/OR gate.\n", nFinal)
				}
				return nFinal
			}
			assert(status == sfm_l_True, "status == l_True")
			p.nSatCallsSat++
			p.timeSatSat += util.Abc_Clock() - clk
			if p.nPats[c] == 64*sfmSimWords {
				p.nSatCallsOver++
				continue //return -2;//continue;
			}
			for i = 0; i < p.nDivs; i++ {
				if p.pSat.VarValueC(i) != 0 {
					util.Abc_TtSetBit(Sfm_DecDivPats(p, i, c), p.nPats[c])
				}
			}
			p.nPatWords[c] = 1 + (p.nPats[c] >> 6)
			util.Abc_TtSetBit(Masks[c][:], p.nPats[c])
			p.nPats[c]++
		}
	}

	// find the best cofactoring variable
	//    if ( !fCofactor || Vec_IntSize(&p->vImpls[0]) + Vec_IntSize(&p->vImpls[1]) > 2 )
	Var = Sfm_DecFindBestVar(p, Masks)
	//    if ( Var == -1 )
	//        Var = Sfm_DecFindBestVar2( p, Masks );

	if Var == -1 && fCofactor != 0 {
		//for ( Var = p->nDivs - 1; Var >= 0; Var-- )
		i := int32(len(p.vObjInMffc)) - 1
		for ; i >= 0; i-- {
			Var = p.vObjInMffc[i]
			if sfmIntFind(p.vObjDec, Var) == -1 {
				break
			}
		}
		//        if ( i == Vec_IntSize(&p->vObjInMffc) )
		if i == -1 {
			Var = -1
		}
		fCofactor = 0
	}

	if p.pPars.VeryVerbose {
		Sfm_DecPrint(p, Masks)
		fmt.Printf("Best var %d\n", Var)
		fmt.Printf("\n")
	}
cofactor:
	// cofactor the problem
	if Var >= 0 {
		var uTruth [2][sfmWordMax]uint64
		var MasksNext [2][sfmSimWords]uint64
		var Supp [2][2 * sfmSuppMax]int32
		var nSupp [2]int32
		p.vObjDec = append(p.vObjDec, Var)
		for i := int32(0); i < 2; i++ {
			for c := int32(0); c < 2; c++ {
				util.Abc_TtAndSharp(MasksNext[c][:], Masks[c][:], Sfm_DecDivPats(p, Var, c), p.nPatWords[c], sfmB2I(i == 0))
				for w := p.nPatWords[c]; w < sfmSimWords; w++ {
					MasksNext[c][w] = 0
				}
			}
			pAssump[nAssump] = 2*Var + sfmB2I(i == 0)
			copy(p.pDivWords[nAssump+1][:p.nDivWords], p.pDivWords[nAssump][:p.nDivWords])
			add := int32(0)
			if i != 0 {
				add = nSupp[0]
			}
			nSupp[i] = Sfm_DecPeformDec_rec(p, uTruth[i][:], Supp[i][:], pAssump, nAssump+1, &MasksNext, fCofactor, add+nSuppAdd+1)
			if nSupp[i] == -2 {
				return -2
			}
		}
		// combine solutions
		return Sfm_DecCombineDec(p, uTruth[0][:], uTruth[1][:], Supp[0][:], Supp[1][:], nSupp[0], nSupp[1], pTruth, pSupp, Var)
	}
	return -2
}

var sfmTruths6 = [6]uint64{
	0xAAAAAAAAAAAAAAAA,
	0xCCCCCCCCCCCCCCCC,
	0xF0F0F0F0F0F0F0F0,
	0xFF00FF00FF00FF00,
	0xFFFF0000FFFF0000,
	0xFFFFFFFF00000000,
}

func Sfm_DecPeformDec2(p *Sfm_Dec_t, pObj *Object) int32 {
	var uTruth [sfmDecMax][sfmWordMax]uint64
	var Masks [2][sfmSimWords]uint64
	var pSupp [sfmDecMax][2 * sfmSuppMax]int32
	var nSupp [sfmDecMax]int32
	var pAssump [sfmWinMax]int32
	fVeryVerbose := p.pPars.PrintDecs || p.pPars.VeryVerbose
	nDecs := max(int32(p.pPars.DecMax), 1)
	//int fNeedInv, AreaGainInv = Sfm_DecComputeFlipInvGain(p, pObj, &fNeedInv);
	var RetValue, AreaThis, AreaNew int32
	Prev, iBest := int32(0), int32(-1)
	GainBest, iLibObj, iLibObjBest := int32(-1), int32(0), int32(-1)
	assert(p.pPars.Area, "p->pPars->fArea == 1")
	if fVeryVerbose {
		fmt.Printf("\nNode %4d : MFFC %2d\n", p.iTarget, p.nMffc)
	}
	assert(p.pPars.DecMax <= sfmDecMax, "p->pPars->nDecMax <= SFM_DEC_MAX")
	Sfm_ObjSetupSimInfo(pObj)
	p.vObjDec = p.vObjDec[:0]
	for i := int32(0); i < nDecs; i++ {
		// reduce the variable array
		if int32(len(p.vObjDec)) > Prev {
			p.vObjDec = p.vObjDec[:Prev]
		}
		Prev = int32(len(p.vObjDec)) + 1
		// perform decomposition
		util.Abc_TtMask(Masks[0][:], sfmSimWords, p.nPats[0])
		util.Abc_TtMask(Masks[1][:], sfmSimWords, p.nPats[1])
		nSupp[i] = Sfm_DecPeformDec_rec(p, uTruth[i][:], pSupp[i][:], pAssump[:], 0, &Masks, 1, 0)
		if nSupp[i] == -2 {
			if fVeryVerbose {
				fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1])
			}
			continue
		}
		if fVeryVerbose {
			fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  Supp = %d  ", i, p.nPats[0], p.nPats[1], nSupp[i])
		}
		if fVeryVerbose {
			dau.Dau_DsdPrintFromTruth(uTruth[i][:], nSupp[i])
		}
		if nSupp[i] < 2 {
			p.nSuppVars = nSupp[i]
			util.Abc_TtCopy(p.Copy[:], uTruth[i][:], sfmWordMax, 0)
			RetValue = sfm.Sfm_LibImplementSimple(p.pLib, uTruth[i][:], pSupp[i][:], nSupp[i], &p.vObjGates, &p.vObjFanins)
			assert(nSupp[i] <= int32(p.pPars.VarMax), "nSupp[i] <= p->pPars->nVarMax")
			p.nLuckySizes[nSupp[i]]++
			assert(RetValue <= 2, "RetValue <= 2")
			p.nLuckyGates[RetValue]++
			//printf( "\n" );
			return RetValue
		}

		p.nSuppVars = nSupp[i]
		util.Abc_TtCopy(p.Copy[:], uTruth[i][:], sfmWordMax, 0)
		AreaNew = sfm.Sfm_LibFindAreaMatch(p.pLib, uTruth[i][:], nSupp[i], &iLibObj)
		if AreaNew == -1 {
			continue
		}

		// compute area savings
		Sfm_DecPrepareVec(p.vObjMap, pSupp[i][:], nSupp[i], &p.vGateCut)
		AreaThis = Sfm_DecMffcAreaReal(pObj, p.vGateCut, nil)
		assert(p.AreaMffc <= AreaThis, "p->AreaMffc <= AreaThis")
		if p.pPars.ZeroCost {
			if AreaNew > AreaThis {
				continue
			}
		} else if AreaNew >= AreaThis {
			continue
		}
		// find the best gain
		GainThis := AreaThis - AreaNew
		assert(GainThis >= 0, "GainThis >= 0")
		if GainBest < GainThis {
			GainBest = GainThis
			iLibObjBest = iLibObj
			iBest = i
		}
	}
	Sfm_ObjSetdownSimInfo(pObj)
	if iBest == -1 {
		if fVeryVerbose {
			fmt.Printf("Best  : NO DEC.\n")
		}
		p.nNoDecs++
		//printf( "\n" );
		return -2
	}
	if fVeryVerbose {
		fmt.Printf("Best %d: %d  ", iBest, nSupp[iBest])
	}
	if fVeryVerbose {
		dau.Dau_DsdPrintFromTruth(uTruth[iBest][:], nSupp[iBest])
	}
	// implement
	assert(iLibObjBest >= 0, "iLibObjBest >= 0")
	RetValue = sfm.Sfm_LibImplementGatesArea(p.pLib, pSupp[iBest][:], nSupp[iBest], iLibObjBest, &p.vObjGates, &p.vObjFanins)
	assert(nSupp[iBest] <= int32(p.pPars.VarMax), "nSupp[iBest] <= p->pPars->nVarMax")
	p.nLuckySizes[nSupp[iBest]]++
	assert(RetValue <= 2, "RetValue <= 2")
	p.nLuckyGates[RetValue]++
	return 1
}
func Sfm_DecPeformDec3(p *Sfm_Dec_t, pObj *Object) int32 {
	var uTruth [sfmDecMax][sfmWordMax]uint64
	var Masks [2][sfmSimWords]uint64
	var pSupp [sfmDecMax][2 * sfmSuppMax]int32
	var nSupp [sfmDecMax]int32
	var pAssump [sfmWinMax]int32
	fVeryVerbose := p.pPars.PrintDecs || p.pPars.VeryVerbose
	nDecs := max(int32(p.pPars.DecMax), 1)
	var DelayOrig, DelayMin, GainMax, AreaMffc, nMatches, RetValue int32
	iBest, Prev := int32(-1), int32(0)
	var pGate1Best, pGate2Best *mio.Mio_Gate_t
	var pFans1Best, pFans2Best []byte
	assert(!p.pPars.Area, "p->pPars->fArea == 0")
	p.DelayMin = 0
	//Sfm_DecPrint( p, NULL );
	if fVeryVerbose {
		fmt.Printf("\nNode %4d : MFFC %2d\n", p.iTarget, p.nMffc)
	}
	// set limit on search for decompositions in delay-model
	assert(p.pPars.DecMax <= sfmDecMax, "p->pPars->nDecMax <= SFM_DEC_MAX")
	Sfm_ObjSetupSimInfo(pObj)
	p.vObjDec = p.vObjDec[:0]
	for i := int32(0); i < nDecs; i++ {
		GainMax = 0
		DelayOrig = Sfm_ManReadObjDelay(p, Abc_ObjId(pObj))
		DelayMin = DelayOrig
		// reduce the variable array
		if int32(len(p.vObjDec)) > Prev {
			p.vObjDec = p.vObjDec[:Prev]
		}
		Prev = int32(len(p.vObjDec)) + 1
		// perform decomposition
		util.Abc_TtMask(Masks[0][:], sfmSimWords, p.nPats[0])
		util.Abc_TtMask(Masks[1][:], sfmSimWords, p.nPats[1])
		nSupp[i] = Sfm_DecPeformDec_rec(p, uTruth[i][:], pSupp[i][:], pAssump[:], 0, &Masks, 1, 0)
		if nSupp[i] == -2 {
			if fVeryVerbose {
				fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1])
			}
			continue
		}
		if fVeryVerbose {
			fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  Supp = %d  ", i, p.nPats[0], p.nPats[1], nSupp[i])
		}
		if fVeryVerbose {
			dau.Dau_DsdPrintFromTruth(uTruth[i][:], nSupp[i])
		}
		if p.pTim != nil && nSupp[i] == 1 && uTruth[i][0] == 0x5555555555555555 && DelayMin <= p.DelayInv+Sfm_ManReadObjDelay(p, p.vObjMap[pSupp[i][0]]) {
			if fVeryVerbose {
				fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1])
			}
			continue
		}
		if p.pMit != nil && nSupp[i] == 1 && uTruth[i][0] == 0x5555555555555555 {
			if fVeryVerbose {
				fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1])
			}
			continue
		}
		if nSupp[i] < 2 {
			p.nSuppVars = nSupp[i]
			util.Abc_TtCopy(p.Copy[:], uTruth[i][:], sfmWordMax, 0)
			RetValue = sfm.Sfm_LibImplementSimple(p.pLib, uTruth[i][:], pSupp[i][:], nSupp[i], &p.vObjGates, &p.vObjFanins)
			assert(nSupp[i] <= int32(p.pPars.VarMax), "nSupp[i] <= p->pPars->nVarMax")
			p.nLuckySizes[nSupp[i]]++
			assert(RetValue <= 2, "RetValue <= 2")
			p.nLuckyGates[RetValue]++
			return RetValue
		}

		// get MFFC
		Sfm_DecPrepareVec(p.vObjMap, pSupp[i][:], nSupp[i], &p.vGateCut) // returns cut in p->vGateCut
		AreaMffc = Sfm_DecMffcAreaReal(pObj, p.vGateCut, &p.vGateMffc)   // returns MFFC in p->vGateMffc

		// try the delay
		p.nSuppVars = nSupp[i]
		util.Abc_TtCopy(p.Copy[:], uTruth[i][:], sfmWordMax, 0)
		nMatches = sfm.Sfm_LibFindDelayMatches(p.pLib, uTruth[i][:], pSupp[i][:], nSupp[i], &p.vMatchGates, &p.vMatchFans)
		for k := int32(0); k < nMatches; k++ {
			clk := util.Abc_Clock()
			pGate1 := p.vMatchGates[2*k+0]
			pGate2 := p.vMatchGates[2*k+1]
			area2 := 0.0
			if pGate2 != nil {
				area2 = mio.Mio_GateReadArea(pGate2)
			}
			AreaNew := scl.Scl_Flt2Int(float32(mio.Mio_GateReadArea(pGate1) + area2))
			pFans1 := p.vMatchFans[2*k+0]
			pFans2 := p.vMatchFans[2*k+1]
			vFanins := pSupp[i][:nSupp[i]]
			// skip identical gate
			//if ( pGate2 == NULL && pGate1 == (Mio_Gate_t *)pObj->pData )
			//    continue;
			if p.pMit != nil {
				Gain := Sfm_MitEvalRemapping(p.pMit, p.vGateMffc, pObj, vFanins, p.vObjMap, pGate1, pFans1, pGate2, pFans2)
				if p.pPars.DelAreaRatio != 0 && AreaNew > AreaMffc && (Gain/(AreaNew-AreaMffc)) < int32(p.pPars.DelAreaRatio) {
					continue
				}
				if GainMax < Gain {
					GainMax = Gain
					pGate1Best = pGate1
					pGate2Best = pGate2
					pFans1Best = pFans1
					pFans2Best = pFans2
					iBest = i
				}
			} else {
				Delay := Sfm_TimEvalRemapping(p.pTim, vFanins, p.vObjMap, pGate1, pFans1, pGate2, pFans2)
				if p.pPars.DelAreaRatio != 0 && AreaNew > AreaMffc && (Delay/(AreaNew-AreaMffc)) < int32(p.pPars.DelAreaRatio) {
					continue
				}
				if DelayMin > Delay {
					DelayMin = Delay
					pGate1Best = pGate1
					pGate2Best = pGate2
					pFans1Best = pFans1
					pFans2Best = pFans2
					iBest = i
				}
			}
			p.timeEval += util.Abc_Clock() - clk
		}
	}
	//printf( "Gain max = %d.\n", GainMax );
	Sfm_ObjSetdownSimInfo(pObj)
	if iBest == -1 {
		if fVeryVerbose {
			fmt.Printf("Best  : NO DEC.\n")
		}
		p.nNoDecs++
		return -2
	}
	if fVeryVerbose {
		fmt.Printf("Best %d: %d  ", iBest, nSupp[iBest])
	}
	//    if ( fVeryVerbose )
	//        Dau_DsdPrintFromTruth( uTruth[iBest], nSupp[iBest] );
	RetValue = sfm.Sfm_LibImplementGatesDelay(p.pLib, pSupp[iBest][:], pGate1Best, pGate2Best, pFans1Best, pFans2Best, &p.vObjGates, &p.vObjFanins)
	assert(nSupp[iBest] <= int32(p.pPars.VarMax), "nSupp[iBest] <= p->pPars->nVarMax")
	p.nLuckySizes[nSupp[iBest]]++
	assert(RetValue <= 2, "RetValue <= 2")
	p.nLuckyGates[RetValue]++
	p.DelayMin = DelayMin
	_ = DelayOrig
	return 1
}

// Abc_NtkUpdateIncLevel_rec is the incremental level update.
func Abc_NtkUpdateIncLevel_rec(pObj *Object) {
	LevelNew := Abc_ObjLevelNew(pObj)
	if LevelNew == Abc_ObjLevel(pObj) && Abc_ObjIsNode(pObj) != 0 && Abc_ObjFaninNum(pObj) > 0 {
		return
	}
	pObj.Level = int(LevelNew)
	if Abc_ObjIsCo(pObj) == 0 {
		for i := 0; i < len(pObj.fanouts); i++ {
			Abc_NtkUpdateIncLevel_rec(pObj.fanouts[i])
		}
	}
}

func Abc_NtkDfsCheck_rec(pObj, pPivot *Object) int32 {
	if pObj == pPivot {
		return 0
	}
	if Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return 1
	}
	Abc_NodeSetTravIdCurrent(pObj)
	if Abc_ObjIsCi(pObj) != 0 {
		return 1
	}
	assert(Abc_ObjIsNode(pObj) != 0, "Abc_ObjIsNode(pObj)")
	for _, e := range pObj.fanins {
		if Abc_NtkDfsCheck_rec(e.Object, pPivot) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkDfsReverseOne_rec(pObj *Object, vTfo *[]int32, nLevelMax, nFanoutMax int32) {
	if Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(pObj)
	if Abc_ObjIsCo(pObj) != 0 || Abc_ObjLevel(pObj) > nLevelMax {
		return
	}
	assert(Abc_ObjIsNode(pObj) != 0, "Abc_ObjIsNode( pObj )")
	if Abc_ObjFanoutNum(pObj) <= nFanoutMax {
		i := int32(0)
		for ; i < Abc_ObjFanoutNum(pObj); i++ {
			pFanout := Abc_ObjFanout(pObj, i)
			if Abc_ObjIsCo(pFanout) != 0 || Abc_ObjLevel(pFanout) > nLevelMax {
				break
			}
		}
		if i == Abc_ObjFanoutNum(pObj) {
			for i = 0; i < Abc_ObjFanoutNum(pObj); i++ {
				Abc_NtkDfsReverseOne_rec(Abc_ObjFanout(pObj, i), vTfo, nLevelMax, nFanoutMax)
			}
		}
	}
	*vTfo = append(*vTfo, Abc_ObjId(pObj))
	pObj.Temp = 0
}
func Abc_NtkDfsOne_rec(pObj *Object, vTfi *[]int32, nLevelMin int32, CiLabel int) int {
	if Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return pObj.Temp
	}
	Abc_NodeSetTravIdCurrent(pObj)
	if Abc_ObjIsCi(pObj) != 0 || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0) {
		*vTfi = append(*vTfi, Abc_ObjId(pObj))
		pObj.Temp = CiLabel
		return pObj.Temp
	}
	assert(Abc_ObjIsNode(pObj) != 0, "Abc_ObjIsNode(pObj)")
	if Abc_ObjFaninNum(pObj) != 0 {
		pObj.Temp = 0
	} else {
		pObj.Temp = CiLabel
	}
	for _, e := range pObj.fanins {
		pObj.Temp |= Abc_NtkDfsOne_rec(e.Object, vTfi, nLevelMin, CiLabel)
	}
	*vTfi = append(*vTfi, Abc_ObjId(pObj))
	Sfm_ObjSimulateNode(pObj)
	return pObj.Temp
}
func Sfm_DecAddNode(pObj *Object, vMap *[]int32, vGates *[]int32, fSkip bool, fVeryVerbose bool) {
	if fVeryVerbose {
		fmt.Printf("%d:%d(%d) ", len(*vMap), Abc_ObjId(pObj), pObj.Temp)
	}
	if fVeryVerbose {
		Abc_ObjPrint(os.Stdout, pObj)
	}
	*vMap = append(*vMap, Abc_ObjId(pObj))
	if fSkip {
		*vGates = append(*vGates, -1)
	} else {
		*vGates = append(*vGates, mio.Mio_GateReadValue(sfmMioGate(pObj)))
	}
}
func Sfm_DecNodeIsMffc(p *Object, nLevelMin int32) bool {
	return Abc_ObjIsNode(p) != 0 && Abc_ObjFanoutNum(p) == 1 && Abc_NodeIsTravIdCurrent(p) != 0 && (Abc_ObjLevel(p) >= nLevelMin || Abc_ObjFaninNum(p) == 0)
}
func Sfm_DecNodeIsMffcInput(p *Object, nLevelMin int32, pTim *Sfm_Tim_t, pPivot *Object) bool {
	return Abc_NodeIsTravIdCurrent(p) != 0 && Sfm_TimNodeIsNonCritical(pTim, pPivot, p)
}
func Sfm_DecNodeIsMffcInput2(p *Object, nLevelMin int32, pMit *Sfm_Mit_t, pPivot *Object) bool {
	return Abc_NodeIsTravIdCurrent(p) != 0 && Sfm_MitNodeIsNonCritical(pMit, pPivot, p) != 0
}
func Sfm_DecMarkMffc(pPivot *Object, nLevelMin, nMffcMax int32, fVeryVerbose bool, vMffc *[]int32, vInMffc *[]int32, pTim *Sfm_Tim_t, pMit *Sfm_Mit_t) {
	assert(nMffcMax > 0, "nMffcMax > 0")
	*vMffc = append((*vMffc)[:0], Abc_ObjId(pPivot))
	if pMit != nil {
		pPivot.Temp |= SFM_MASK_MFFC
		pPivot.Temp |= SFM_MASK_PIVOT
		// collect MFFC inputs (these are low-delay nodes close to the pivot)
		*vInMffc = (*vInMffc)[:0]
		for _, e := range pPivot.fanins {
			if Sfm_DecNodeIsMffcInput2(e.Object, nLevelMin, pMit, pPivot) {
				sfmIntPushUnique(vInMffc, Abc_ObjId(e.Object))
			}
		}
		for _, e := range pPivot.fanins {
			for _, e2 := range e.Object.fanins {
				if Sfm_DecNodeIsMffcInput2(e2.Object, nLevelMin, pMit, pPivot) {
					sfmIntPushUnique(vInMffc, Abc_ObjId(e2.Object))
				}
			}
		}
		for _, e := range pPivot.fanins {
			for _, e2 := range e.Object.fanins {
				for _, e3 := range e2.Object.fanins {
					if Sfm_DecNodeIsMffcInput2(e3.Object, nLevelMin, pMit, pPivot) {
						sfmIntPushUnique(vInMffc, Abc_ObjId(e3.Object))
					}
				}
			}
		}
	} else if pTim != nil {
		pPivot.Temp |= SFM_MASK_MFFC
		pPivot.Temp |= SFM_MASK_PIVOT
		// collect MFFC inputs (these are low-delay nodes close to the pivot)
		*vInMffc = (*vInMffc)[:0]
		for _, e := range pPivot.fanins {
			if Sfm_DecNodeIsMffcInput(e.Object, nLevelMin, pTim, pPivot) {
				sfmIntPushUnique(vInMffc, Abc_ObjId(e.Object))
			}
		}
		for _, e := range pPivot.fanins {
			for _, e2 := range e.Object.fanins {
				if Sfm_DecNodeIsMffcInput(e2.Object, nLevelMin, pTim, pPivot) {
					sfmIntPushUnique(vInMffc, Abc_ObjId(e2.Object))
				}
			}
		}
		for _, e := range pPivot.fanins {
			for _, e2 := range e.Object.fanins {
				for _, e3 := range e2.Object.fanins {
					if Sfm_DecNodeIsMffcInput(e3.Object, nLevelMin, pTim, pPivot) {
						sfmIntPushUnique(vInMffc, Abc_ObjId(e3.Object))
					}
				}
			}
		}
	} else {
		// collect MFFC
		for _, e := range pPivot.fanins {
			if Sfm_DecNodeIsMffc(e.Object, nLevelMin) && int32(len(*vMffc)) < nMffcMax {
				sfmIntPushUnique(vMffc, Abc_ObjId(e.Object))
			}
		}
		for _, e := range pPivot.fanins {
			if Sfm_DecNodeIsMffc(e.Object, nLevelMin) && int32(len(*vMffc)) < nMffcMax {
				for _, e2 := range e.Object.fanins {
					if Sfm_DecNodeIsMffc(e2.Object, nLevelMin) && int32(len(*vMffc)) < nMffcMax {
						sfmIntPushUnique(vMffc, Abc_ObjId(e2.Object))
					}
				}
			}
		}
		for _, e := range pPivot.fanins {
			if Sfm_DecNodeIsMffc(e.Object, nLevelMin) && int32(len(*vMffc)) < nMffcMax {
				for _, e2 := range e.Object.fanins {
					if Sfm_DecNodeIsMffc(e2.Object, nLevelMin) && int32(len(*vMffc)) < nMffcMax {
						for _, e3 := range e2.Object.fanins {
							if Sfm_DecNodeIsMffc(e3.Object, nLevelMin) && int32(len(*vMffc)) < nMffcMax {
								sfmIntPushUnique(vMffc, Abc_ObjId(e3.Object))
							}
						}
					}
				}
			}
		}
		// mark MFFC
		assert(int32(len(*vMffc)) <= nMffcMax, "Vec_IntSize(vMffc) <= nMffcMax")
		for _, id := range *vMffc {
			if pObj := Abc_NtkObj(pPivot.network, id); pObj != nil {
				pObj.Temp |= SFM_MASK_MFFC
			}
		}
		pPivot.Temp |= SFM_MASK_PIVOT
		// collect MFFC inputs
		*vInMffc = (*vInMffc)[:0]
		for _, id := range *vMffc {
			pObj := Abc_NtkObj(pPivot.network, id)
			if pObj == nil {
				continue
			}
			for _, e := range pObj.fanins {
				pFanin := e.Object
				if Abc_NodeIsTravIdCurrent(pFanin) != 0 && pFanin.Temp == SFM_MASK_PI {
					sfmIntPushUnique(vInMffc, Abc_ObjId(pFanin))
				}
			}
		}
	}
}

func Sfm_DecExtract(pNtk *Network, pPars *sfm.Sfm_Par_t, pPivot *Object, vRoots *[]int32, vGates *[]int32, vFanins *[][]int32, vMap *[]int32, vTfi *[]int32, vTfo *[]int32, vMffc *[]int32, vInMffc *[]int32, pTim *Sfm_Tim_t, pMit *Sfm_Mit_t) int32 {
	fVeryVerbose := false //pPars->fVeryVerbose;
	nLevelMax := int32(pPivot.Level + pPars.TfoLevMax)
	nLevelMin := int32(pPivot.Level - pPars.TfiLevMax)
	var nTfiSize int
	nDivs := int32(-1)
	assert(Abc_ObjIsNode(pPivot) != 0, "Abc_ObjIsNode(pPivot)")
	if fVeryVerbose {
		fmt.Printf("\n\nTarget %d\n", Abc_ObjId(pPivot))
	}
	// collect TFO nodes
	*vTfo = (*vTfo)[:0]
	Abc_NtkIncrementTravId(pNtk)
	Abc_NtkDfsReverseOne_rec(pPivot, vTfo, nLevelMax, int32(pPars.FanoutMax))
	// count internal fanouts
	for _, id := range *vTfo {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		for _, e := range pObj.fanins {
			e.Object.Temp++
		}
	}
	// compute roots
	*vRoots = (*vRoots)[:0]
	for _, id := range *vTfo {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		if int32(pObj.Temp) != Abc_ObjFanoutNum(pObj) {
			*vRoots = append(*vRoots, Abc_ObjId(pObj))
		}
	}
	assert(len(*vRoots) > 0, "Vec_IntSize(vRoots) > 0")
	// collect TFI and mark nodes
	*vTfi = (*vTfi)[:0]
	Abc_NtkIncrementTravId(pNtk)
	Abc_NtkDfsOne_rec(pPivot, vTfi, nLevelMin, SFM_MASK_PI)
	nTfiSize = len(*vTfi)
	Sfm_ObjFlipNode(pPivot)
	// additinally mark MFFC
	Sfm_DecMarkMffc(pPivot, nLevelMin, int32(pPars.MffcMax), fVeryVerbose, vMffc, vInMffc, pTim, pMit)
	assert(int32(len(*vMffc)) <= int32(pPars.MffcMax), "Vec_IntSize(vMffc) <= pPars->nMffcMax")
	if fVeryVerbose {
		fmt.Printf("Mffc size = %d. Mffc area = %.2f. InMffc size = %d.\n", len(*vMffc), scl.Scl_Int2Flt(Sfm_DecMffcArea(pNtk, *vMffc)), len(*vInMffc))
	}
	// collect TFI(TFO)
	for _, id := range *vRoots {
		if pObj := Abc_NtkObj(pNtk, id); pObj != nil {
			Abc_NtkDfsOne_rec(pObj, vTfi, nLevelMin, SFM_MASK_INPUT)
		}
	}
	// mark input-only nodes pointed to by mixed nodes
	for _, id := range (*vTfi)[nTfiSize:] {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		if pObj.Temp != SFM_MASK_INPUT {
			for _, e := range pObj.fanins {
				if e.Object.Temp == SFM_MASK_INPUT {
					e.Object.Temp = SFM_MASK_FANIN
				}
			}
		}
	}
	// collect nodes supported only on TFI fanins and not MFFC
	if fVeryVerbose {
		fmt.Printf("\nDivs:\n")
	}
	*vMap = (*vMap)[:0]
	*vGates = (*vGates)[:0]
	for _, id := range *vTfi {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		if pObj.Temp == SFM_MASK_PI {
			Sfm_DecAddNode(pObj, vMap, vGates, Abc_ObjIsCi(pObj) != 0 || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0), fVeryVerbose)
		}
	}
	nDivs = int32(len(*vMap))
	// add other nodes that are not in TFO and not in MFFC
	if fVeryVerbose {
		fmt.Printf("\nSides:\n")
	}
	for _, id := range *vTfi {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		if pObj.Temp == (SFM_MASK_PI|SFM_MASK_INPUT) || pObj.Temp == SFM_MASK_FANIN {
			Sfm_DecAddNode(pObj, vMap, vGates, pObj.Temp == SFM_MASK_FANIN, fVeryVerbose)
		}
	}
	// reorder nodes acording to delay
	if pMit != nil {
		nOldSize := len(*vMap)
		*vTfo = append((*vTfo)[:0], *vMap...)
		nDivsNew := Sfm_MitSortArrayByArrival(pMit, vTfo, Abc_ObjId(pPivot))
		// collect again
		*vMap = (*vMap)[:0]
		*vGates = (*vGates)[:0]
		for _, id := range *vTfo {
			pObj := Abc_NtkObj(pNtk, id)
			if pObj == nil {
				continue
			}
			Sfm_DecAddNode(pObj, vMap, vGates, Abc_ObjIsCi(pObj) != 0 || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0) || pObj.Temp == SFM_MASK_FANIN, false)
		}
		assert(nOldSize == len(*vMap), "nOldSize == Vec_IntSize(vMap)")
		// update divisor count
		nDivs = nDivsNew
	} else if pTim != nil {
		nOldSize := len(*vMap)
		*vTfo = append((*vTfo)[:0], *vMap...)
		nDivsNew := Sfm_TimSortArrayByArrival(pTim, vTfo, Abc_ObjId(pPivot))
		// collect again
		*vMap = (*vMap)[:0]
		*vGates = (*vGates)[:0]
		for _, id := range *vTfo {
			pObj := Abc_NtkObj(pNtk, id)
			if pObj == nil {
				continue
			}
			Sfm_DecAddNode(pObj, vMap, vGates, Abc_ObjIsCi(pObj) != 0 || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0) || pObj.Temp == SFM_MASK_FANIN, false)
		}
		assert(nOldSize == len(*vMap), "nOldSize == Vec_IntSize(vMap)")
		// update divisor count
		nDivs = nDivsNew
	}
	// add the TFO nodes
	if fVeryVerbose {
		fmt.Printf("\nTFO:\n")
	}
	for _, id := range *vTfi {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		if pObj.Temp >= SFM_MASK_MFFC {
			Sfm_DecAddNode(pObj, vMap, vGates, false, fVeryVerbose)
		}
	}
	assert(len(*vMap) == len(*vGates), "Vec_IntSize(vMap) == Vec_IntSize(vGates)")
	if fVeryVerbose {
		fmt.Printf("\n")
	}
	// create node IDs
	*vFanins = (*vFanins)[:0]
	for i, id := range *vMap {
		pObj := Abc_NtkObj(pNtk, id)
		if pObj == nil {
			continue
		}
		pObj.Temp = i
		var vLevel []int32
		if (*vGates)[i] >= 0 {
			for _, e := range pObj.fanins {
				vLevel = append(vLevel, int32(e.Object.Temp))
			}
		}
		*vFanins = append(*vFanins, vLevel)
	}
	// compute care set
	Sfm_DecMan(pPivot).uCareSet = Sfm_ObjFindCareSet(pPivot.network, *vRoots)

	// remap roots
	for i, id := range *vRoots {
		if pObj := Abc_NtkObj(pNtk, id); pObj != nil {
			(*vRoots)[i] = int32(pObj.Temp)
		}
	}
	// remap inputs to MFFC
	for i, id := range *vInMffc {
		if pObj := Abc_NtkObj(pNtk, id); pObj != nil {
			(*vInMffc)[i] = int32(pObj.Temp)
		}
	}
	return nDivs
}
func Sfm_DecInsert(pNtk *Network, pPivot *Object, Limit int32, vGates []int32, vFanins [][]int32, vMap *[]int32, vGateHandles []*mio.Mio_Gate_t, GateBuf, GateInv int32, vFuncs []uint64, vNewNodes *[]int32, pMit *Sfm_Mit_t) *Object {
	var pObjNew *Object
	if vNewNodes != nil {
		*vNewNodes = (*vNewNodes)[:0]
	}
	// assuming that new gates are appended at the end
	assert(Limit < int32(len(vGates)), "Limit < Vec_IntSize(vGates)")
	assert(Limit == int32(len(*vMap)), "Limit == Vec_IntSize(vMap)")
	if Limit+1 == int32(len(vGates)) {
		Gate := vGates[len(vGates)-1]
		if Gate == GateBuf {
			iObj := vFanins[Limit][0]
			pObjNew = Abc_NtkObj(pNtk, (*vMap)[iObj])
			// transfer load
			if pMit != nil {
				Sfm_MitTransferLoad(pMit, pObjNew, pPivot)
			}
			// replace logic cone
			Abc_ObjReplace(pPivot, pObjNew)
			// update level
			pObjNew.Level = 0
			Abc_NtkUpdateIncLevel_rec(pObjNew)
			if vNewNodes != nil {
				*vNewNodes = append(*vNewNodes, Abc_ObjId(pObjNew))
			}
			return pObjNew
		} else if vNewNodes == nil && Gate == GateInv {
			// check if fanouts can be updated
			i := 0
			for ; i < len(pPivot.fanouts); i++ {
				pFanout := pPivot.fanouts[i]
				if Abc_ObjIsNode(pFanout) == 0 || sfm.Sfm_LibFindComplInputGate(vFuncs, mio.Mio_GateReadValue(sfmMioGate(pFanout)), Abc_ObjFaninNum(pFanout), Abc_NodeFindFanin(pFanout, pPivot), nil) == -1 {
					break
				}
			}
			// update fanouts
			if int32(i) == Abc_ObjFanoutNum(pPivot) {
				for _, pFanout := range pPivot.fanouts {
					iFanin, iFaninNew := Abc_NodeFindFanin(pFanout, pPivot), int32(-1)
					iGate := mio.Mio_GateReadValue(sfmMioGate(pFanout))
					iGateNew := sfm.Sfm_LibFindComplInputGate(vFuncs, iGate, Abc_ObjFaninNum(pFanout), iFanin, &iFaninNew)
					assert(iGateNew >= 0 && iGateNew != iGate && iFaninNew >= 0, "iGateNew >= 0 && iGateNew != iGate && iFaninNew >= 0")
					pFanout.Data = vGateHandles[iGateNew]
					//assert( iFanin == iFaninNew );
					// swap fanins
					if iFanin != iFaninNew {
						pFanout.fanins[iFanin], pFanout.fanins[iFaninNew] = pFanout.fanins[iFaninNew], pFanout.fanins[iFanin]
					}
				}
				iObj := vFanins[Limit][0]
				pObjNew = Abc_NtkObj(pNtk, (*vMap)[iObj])
				Abc_ObjReplace(pPivot, pObjNew)
				// update level
				pObjNew.Level = 0
				Abc_NtkUpdateIncLevel_rec(pObjNew)
				return pObjNew
			}
		}
	}
	// introduce new gates
	for i := Limit; i < int32(len(vGates)); i++ {
		Gate := vGates[i]
		vLevel := vFanins[i]
		pObjNew = Abc_NtkCreateNode(pNtk)
		for _, iObj := range vLevel {
			Abc_ObjAddFanin(pObjNew, Abc_NtkObj(pNtk, (*vMap)[iObj]))
		}
		pObjNew.Data = vGateHandles[Gate]
		assert(Abc_ObjFaninNum(pObjNew) == mio.Mio_GateReadPinNum(sfmMioGate(pObjNew)), "Abc_ObjFaninNum(pObjNew) == Mio_GateReadPinNum(pObjNew->pData)")
		*vMap = append(*vMap, Abc_ObjId(pObjNew))
		if vNewNodes != nil {
			*vNewNodes = append(*vNewNodes, Abc_ObjId(pObjNew))
		}
	}
	// transfer load
	if pMit != nil {
		Sfm_MitTimingGrow(pMit)
		Sfm_MitTransferLoad(pMit, pObjNew, pPivot)
	}
	// replace logic cone
	Abc_ObjReplace(pPivot, pObjNew)
	// update level
	for _, id := range (*vMap)[Limit:] {
		if pObjNew = Abc_NtkObj(pNtk, id); pObjNew != nil {
			Abc_NtkUpdateIncLevel_rec(pObjNew)
		}
	}
	return pObjNew
}

func sfmPRTP(a string, t, T int64) {
	util.Abc_PrintTimeP(1, a, t, T)
}

func Sfm_DecPrintStats(p *Sfm_Dec_t) {
	fmt.Printf("Node = %d. Try = %d. Change = %d.   Const0 = %d. Const1 = %d. Buf = %d. Inv = %d. Gate = %d. AndOr = %d. Effort = %d.  NoDec = %d.\n",
		p.nTotalNodesBeg, p.nNodesTried, p.nNodesChanged, p.nNodesConst0, p.nNodesConst1, p.nNodesBuf, p.nNodesInv, p.nNodesResyn, p.nNodesAndOr, p.nEfforts, p.nNoDecs)
	fmt.Printf("MaxDiv = %d. MaxWin = %d.   AveDiv = %d. AveWin = %d.   Calls = %d. (Sat = %d. Unsat = %d.)  Over = %d.  T/O = %d.\n",
		p.nMaxDivs, p.nMaxWin, int32(p.nAllDivs/uint64(max(1, p.nNodesTried))), int32(p.nAllWin/uint64(max(1, p.nNodesTried))), p.nSatCalls, p.nSatCallsSat, p.nSatCallsUnsat, p.nSatCallsOver, p.nTimeOuts)

	p.timeTotal = util.Abc_Clock() - p.timeStart
	p.timeOther = p.timeTotal - p.timeLib - p.timeWin - p.timeCnf - p.timeSat - p.timeTime

	sfmPRTP("Lib   ", p.timeLib, p.timeTotal)
	sfmPRTP("Win   ", p.timeWin, p.timeTotal)
	sfmPRTP("Cnf   ", p.timeCnf, p.timeTotal)
	sfmPRTP("Sat   ", p.timeSat-p.timeEval, p.timeTotal)
	sfmPRTP(" Sat  ", p.timeSatSat, p.timeTotal)
	sfmPRTP(" Unsat", p.timeSatUnsat, p.timeTotal)
	sfmPRTP("Eval  ", p.timeEval, p.timeTotal)
	sfmPRTP("Timing", p.timeTime, p.timeTotal)
	sfmPRTP("Other ", p.timeOther, p.timeTotal)
	sfmPRTP("ALL   ", p.timeTotal, p.timeTotal)

	fmt.Printf("Cone sizes:  ")
	for i := 0; i <= sfmSuppMax; i++ {
		if p.nLuckySizes[i] != 0 {
			fmt.Printf("%d=%d  ", i, p.nLuckySizes[i])
		}
	}
	fmt.Printf("  ")

	fmt.Printf("Gate sizes:  ")
	for i := 0; i <= sfmSuppMax; i++ {
		if p.nLuckyGates[i] != 0 {
			fmt.Printf("%d=%d  ", i, p.nLuckyGates[i])
		}
	}
	fmt.Printf("\n")

	fmt.Printf("Reduction:   ")
	fmt.Printf("Nodes  %6d out of %6d (%6.2f %%)   ", p.nTotalNodesBeg-p.nTotalNodesEnd, p.nTotalNodesBeg, 100.0*float64(p.nTotalNodesBeg-p.nTotalNodesEnd)/float64(max(1, p.nTotalNodesBeg)))
	fmt.Printf("Edges  %6d out of %6d (%6.2f %%)   ", p.nTotalEdgesBeg-p.nTotalEdgesEnd, p.nTotalEdgesBeg, 100.0*float64(p.nTotalEdgesBeg-p.nTotalEdgesEnd)/float64(max(1, p.nTotalEdgesBeg)))
	fmt.Printf("\n")
}
func Abc_NtkCountStats(p *Sfm_Dec_t, Limit int32) {
	nGates := int32(len(p.vObjGates))
	if nGates == Limit {
		return
	}
	Gate := p.vObjGates[nGates-1]
	if nGates > Limit+1 {
		p.nNodesResyn++
	} else if Gate == p.GateConst0 {
		p.nNodesConst0++
	} else if Gate == p.GateConst1 {
		p.nNodesConst1++
	} else if Gate == p.GateBuffer {
		p.nNodesBuf++
	} else if Gate == p.GateInvert {
		p.nNodesInv++
	} else {
		p.nNodesResyn++
	}
}

func Abc_NtkAreaOptOne(p *Sfm_Dec_t, i int32) *Object {
	var clk int64
	pNtk := p.pNtk
	pPars := p.pPars
	pObj := Abc_NtkObj(p.pNtk, i)
	var Limit, RetValue int32
	if pPars.MffcMin > 1 && int(Abc_NodeMffcLabel(pObj, nil)) < pPars.MffcMin {
		return nil
	}
	if pPars.NodeOne != 0 && int(i) != pPars.NodeOne {
		return nil
	}
	if pPars.NodeOne != 0 {
		pPars.VeryVerbose = int(i) == pPars.NodeOne
	}
	p.nNodesTried++
	clk = util.Abc_Clock()
	p.nDivs = Sfm_DecExtract(pNtk, pPars, pObj, &p.vObjRoots, &p.vObjGates, &p.vObjFanins, &p.vObjMap, &p.vGateTfi, &p.vGateTfo, &p.vObjMffc, &p.vObjInMffc, nil, nil)
	p.timeWin += util.Abc_Clock() - clk
	if pPars.WinSizeMax != 0 && pPars.WinSizeMax < len(p.vObjGates) {
		return nil
	}
	p.nMffc = int32(len(p.vObjMffc))
	p.AreaMffc = Sfm_DecMffcArea(pNtk, p.vObjMffc)
	p.nMaxDivs = max(p.nMaxDivs, p.nDivs)
	p.nAllDivs += uint64(p.nDivs)
	p.iTarget = int32(pObj.Temp)
	Limit = int32(len(p.vObjGates))
	p.nMaxWin = max(p.nMaxWin, Limit)
	p.nAllWin += uint64(Limit)
	clk = util.Abc_Clock()
	RetValue = Sfm_DecPrepareSolver(p)
	p.timeCnf += util.Abc_Clock() - clk
	if RetValue == 0 {
		return nil
	}
	clk = util.Abc_Clock()
	RetValue = Sfm_DecPeformDec2(p, pObj)
	if pPars.MoreEffort && RetValue < 0 {
		for i := len(p.vObjInMffc) - 1; i >= 0; i-- {
			Var := p.vObjInMffc[i]
			p.iUseThis = Var
			RetValue = Sfm_DecPeformDec2(p, pObj)
			p.iUseThis = -1
			if RetValue < 0 {
				//printf( "Node %d: Not found among %d.\n", Abc_ObjId(pObj), Vec_IntSize(&p->vObjInMffc) );
			} else {
				p.nEfforts++
				if p.pPars.Verbose {
					//printf( "Node %5d: (%2d out of %2d)  Gate=%s ", Abc_ObjId(pObj), i, Vec_IntSize(&p->vObjInMffc), Mio_GateReadName((Mio_Gate_t*)pObj->pData) );
					//Dau_DsdPrintFromTruth( p->Copy, p->nSuppVars );
				}
				break
			}
		}
	}
	if p.pPars.VeryVerbose {
		fmt.Printf("\n\n")
	}
	p.timeSat += util.Abc_Clock() - clk
	if RetValue < 0 {
		return nil
	}
	p.nNodesChanged++
	Abc_NtkCountStats(p, Limit)
	return Sfm_DecInsert(pNtk, pObj, Limit, p.vObjGates, p.vObjFanins, &p.vObjMap, p.vGateHands, p.GateBuffer, p.GateInvert, p.vGateFuncs, nil, p.pMit)
}
func Abc_NtkAreaOpt(p *Sfm_Dec_t) {
	nStop := Abc_NtkObjNumMax(p.pNtk)
	for i := int32(0); i < Abc_NtkObjNumMax(p.pNtk); i++ {
		pObj := Abc_NtkObj(p.pNtk, i)
		if pObj == nil || Abc_ObjIsNode(pObj) == 0 {
			continue
		}
		if i >= nStop || (p.pPars.NodesMax != 0 && int(i) > p.pPars.NodesMax) {
			break
		}
		Abc_NtkAreaOptOne(p, i)
	}
}
func Abc_NtkAreaOpt2(p *Sfm_Dec_t) {
	nStop := Abc_NtkObjNumMax(p.pNtk)
	var vFront []*Object
	for _, pObj := range p.pNtk.objects {
		if pObj != nil {
			assert(!pObj.MarkB, "pObj->fMarkB == 0")
		}
	}
	// start the queue of nodes to be tried
	for i := int32(0); i < Abc_NtkCoNum(p.pNtk); i++ {
		pObj := Abc_NtkCo(p.pNtk, i)
		if Abc_ObjIsNode(Abc_ObjFanin0(pObj)) != 0 && !Abc_ObjFanin0(pObj).MarkB {
			Abc_ObjFanin0(pObj).MarkB = true
			vFront = append(vFront, Abc_ObjFanin0(pObj))
		}
	}
	// process nodes in this order
	for i := 0; i < len(vFront); i++ {
		pObj := vFront[i]
		if pObj.Kind == ObjectNone || pObj.network == nil || Abc_NtkObj(p.pNtk, Abc_ObjId(pObj)) != pObj { // Abc_ObjIsNone
			continue
		}
		pObjNew := Abc_NtkAreaOptOne(p, Abc_ObjId(pObj))
		if pObjNew != nil {
			if Abc_ObjIsNode(pObjNew) == 0 || Abc_ObjFaninNum(pObjNew) == 0 || pObjNew.MarkB {
				continue
			}
			if Abc_ObjId(pObjNew) < nStop {
				pObjNew.MarkB = true
				vFront = append(vFront, pObjNew)
				continue
			}
		} else {
			pObjNew = pObj
		}
		for _, e := range pObjNew.fanins {
			pFanin := e.Object
			if Abc_ObjIsNode(pFanin) != 0 && Abc_ObjFaninNum(pObjNew) > 0 && !pFanin.MarkB {
				pFanin.MarkB = true
				vFront = append(vFront, pFanin)
			}
		}
	}
	for _, pObj := range p.pNtk.objects {
		if pObj != nil {
			pObj.MarkB = false
		}
	}
}

func Abc_NtkDelayOpt(p *Sfm_Dec_t) {
	pNtk := p.pNtk
	pPars := p.pPars
	Abc_NtkCleanMarkABC(pNtk)
	for n := 0; pPars.NodesMax == 0 || n < pPars.NodesMax; n++ {
		var clk int64
		var Limit, RetValue int32
		// collect nodes
		if pPars.NodeOne != 0 {
			p.vCands = append(p.vCands[:0], int32(pPars.NodeOne))
		} else if p.pTim != nil && !Sfm_TimPriorityNodes(p.pTim, &p.vCands, int32(p.pPars.TimeWin)) {
			break
		} else if p.pMit != nil && Sfm_MitPriorityNodes(p.pMit, &p.vCands, int32(p.pPars.TimeWin)) == 0 {
			break
		}
		// try improving delay for the nodes according to the priority
		for i := 0; i < len(p.vCands); i++ {
			pObj := Abc_NtkObj(p.pNtk, p.vCands[i])
			if pObj == nil {
				continue
			}
			OldId := Abc_ObjId(pObj)
			DelayOld := Sfm_ManReadObjDelay(p, OldId)
			assert(!pObj.MarkA, "pObj->fMarkA == 0")
			p.nNodesTried++
			clk = util.Abc_Clock()
			p.nDivs = Sfm_DecExtract(pNtk, pPars, pObj, &p.vObjRoots, &p.vObjGates, &p.vObjFanins, &p.vObjMap, &p.vGateTfi, &p.vGateTfo, &p.vObjMffc, &p.vObjInMffc, p.pTim, p.pMit)
			p.timeWin += util.Abc_Clock() - clk
			if p.nDivs < 2 || (pPars.WinSizeMax != 0 && pPars.WinSizeMax < len(p.vObjGates)) {
				pObj.MarkA = true
				continue
			}
			p.nMffc = int32(len(p.vObjMffc))
			p.AreaMffc = Sfm_DecMffcArea(pNtk, p.vObjMffc)
			p.nMaxDivs = max(p.nMaxDivs, p.nDivs)
			p.nAllDivs += uint64(p.nDivs)
			p.iTarget = int32(pObj.Temp)
			Limit = int32(len(p.vObjGates))
			p.nMaxWin = max(p.nMaxWin, Limit)
			p.nAllWin += uint64(Limit)
			clk = util.Abc_Clock()
			RetValue = Sfm_DecPrepareSolver(p)
			p.timeCnf += util.Abc_Clock() - clk
			if RetValue == 0 {
				pObj.MarkA = true
				continue
			}
			clk = util.Abc_Clock()
			RetValue = Sfm_DecPeformDec3(p, pObj)
			if pPars.MoreEffort && RetValue < 0 {
				for k := len(p.vObjInMffc) - 1; k >= 0; k-- {
					Var := p.vObjInMffc[k]
					p.iUseThis = Var
					RetValue = Sfm_DecPeformDec3(p, pObj)
					p.iUseThis = -1
					if RetValue < 0 {
						//printf( "Node %d: Not found among %d.\n", Abc_ObjId(pObj), Vec_IntSize(&p->vObjInMffc) );
					} else {
						p.nEfforts++
						if p.pPars.Verbose {
							//printf( "Node %5d: (%2d out of %2d)  Gate=%s ", Abc_ObjId(pObj), i, Vec_IntSize(&p->vObjInMffc), Mio_GateReadName((Mio_Gate_t*)pObj->pData) );
							//Dau_DsdPrintFromTruth( p->Copy, p->nSuppVars );
						}
						break
					}
				}
			}
			if p.pPars.VeryVerbose {
				fmt.Printf("\n\n")
			}
			p.timeSat += util.Abc_Clock() - clk
			if RetValue < 0 {
				pObj.MarkA = true
				continue
			}
			assert(int32(len(p.vObjGates))-Limit > 0, "Vec_IntSize(&p->vObjGates) - Limit > 0")
			assert(int32(len(p.vObjGates))-Limit <= 2, "Vec_IntSize(&p->vObjGates) - Limit <= 2")
			p.nNodesChanged++
			Abc_NtkCountStats(p, Limit)
			// reduce load due to removed MFFC
			if p.pMit != nil {
				Sfm_MitUpdateLoad(p.pMit, p.vGateMffc, 0) // assuming &p->vGateMffc contains MFFC
			}
			Sfm_DecInsert(pNtk, pObj, Limit, p.vObjGates, p.vObjFanins, &p.vObjMap, p.vGateHands, p.GateBuffer, p.GateInvert, p.vGateFuncs, &p.vNewNodes, p.pMit)
			// increase load due to added new nodes
			if p.pMit != nil {
				Sfm_MitUpdateLoad(p.pMit, p.vNewNodes, 1) // assuming &p->vNewNodes contains new nodes
			}
			clk = util.Abc_Clock()
			if p.pMit != nil {
				Sfm_MitUpdateTiming(p.pMit, p.vNewNodes)
			} else {
				Sfm_TimUpdateTiming(p.pTim, p.vNewNodes)
			}
			p.timeTime += util.Abc_Clock() - clk
			pObjNew := Abc_NtkObj(pNtk, Abc_NtkObjNumMax(pNtk)-1)
			// C: assert( p->pMit || p->DelayMin == 0 || p->DelayMin == Sfm_ManReadObjDelay(p, Abc_ObjId(pObjNew)) );
			// It fails with -R > 1 (Sfm_DecPeformDec3 keeps the DelayMin of
			// the last decomposition, not of the best one); release builds
			// of ABC do not check it, so neither does the port.
			// report
			if pPars.DelayVerbose {
				fmt.Printf("Node %5d  %5d :  I =%3d.  Cand = %5d (%6.2f %%)   Old =%8.2f.  New =%8.2f.  Final =%8.2f.  WNS =%8.2f.\n",
					OldId, Abc_NtkObjNumMax(p.pNtk), i, len(p.vCands), 100.0*float64(len(p.vCands))/float64(Abc_NtkNodeNum(p.pNtk)),
					scl.Scl_Int2Flt(DelayOld), scl.Scl_Int2Flt(Sfm_ManReadObjDelay(p, Abc_ObjId(pObjNew))),
					scl.Scl_Int2Flt(Sfm_ManReadNtkDelay(p)), scl.Scl_Int2Flt(Sfm_ManReadNtkMinSlack(p)))
			}
			break
		}
		if pPars.NodeOne != 0 {
			break
		}
	}
	Abc_NtkCleanMarkABC(pNtk)
}
func Abc_NtkPerformMfs3(pNtk *Network, pPars *sfm.Sfm_Par_t) {
	pLib, _ := pNtk.FunctionManager.(*mio.Mio_Library_t)
	// Sfm_DecStart reads these gates (same messages as Mio_CollectRootsNew2)
	if mio.Mio_LibraryReadConst0(pLib) == nil {
		fmt.Printf("Error: Cannot find constant 0 gate in the library.\n")
		return
	}
	if mio.Mio_LibraryReadConst1(pLib) == nil {
		fmt.Printf("Error: Cannot find constant 1 gate in the library.\n")
		return
	}
	if mio.Mio_LibraryReadBuf(pLib) == nil {
		fmt.Printf("Error: Cannot find buffer gate in the library.\n")
		return
	}
	if mio.Mio_LibraryReadInv(pLib) == nil {
		fmt.Printf("Error: Cannot find inverter gate in the library.\n")
		return
	}
	p := Sfm_DecStart(pPars, pLib, pNtk)
	if pPars.Verbose {
		fmt.Printf("Remapping parameters: ")
		if pPars.TfoLevMax != 0 {
			fmt.Printf("TFO = %d. ", pPars.TfoLevMax)
		}
		if pPars.TfiLevMax != 0 {
			fmt.Printf("TFI = %d. ", pPars.TfiLevMax)
		}
		if pPars.FanoutMax != 0 {
			fmt.Printf("FanMax = %d. ", pPars.FanoutMax)
		}
		if pPars.WinSizeMax != 0 {
			fmt.Printf("WinMax = %d. ", pPars.WinSizeMax)
		}
		if pPars.BTLimit != 0 {
			fmt.Printf("Confl = %d. ", pPars.BTLimit)
		}
		if pPars.MffcMin != 0 && pPars.Area {
			fmt.Printf("MffcMin = %d. ", pPars.MffcMin)
		}
		if pPars.MffcMax != 0 && pPars.Area {
			fmt.Printf("MffcMax = %d. ", pPars.MffcMax)
		}
		if pPars.DecMax != 0 {
			fmt.Printf("DecMax = %d. ", pPars.DecMax)
		}
		if pPars.NodeOne != 0 {
			fmt.Printf("Pivot = %d. ", pPars.NodeOne)
		}
		if !pPars.Area {
			fmt.Printf("Win = %d. ", pPars.TimeWin)
		}
		if !pPars.Area {
			fmt.Printf("Delta = %.2f ps. ", scl.Scl_Int2Flt(p.DeltaCrit))
		}
		if pPars.Area {
			yes := "no"
			if pPars.ZeroCost {
				yes = "yes"
			}
			fmt.Printf("0-cost = %s. ", yes)
		}
		yes := "no"
		if pPars.MoreEffort {
			yes = "yes"
		}
		fmt.Printf("Effort = %s. ", yes)
		yes = "no"
		if pPars.UseSim {
			yes = "yes"
		}
		fmt.Printf("Sim = %s. ", yes)
		fmt.Printf("\n")
	}
	// preparation steps
	Abc_NtkLevel(pNtk)
	if p.pPars.UseSim {
		Sfm_NtkSimulate(pNtk)
	}
	// record statistics
	if pPars.Verbose {
		p.nTotalNodesBeg = Abc_NtkNodeNum(pNtk)
	}
	if pPars.Verbose {
		p.nTotalEdgesBeg = Abc_NtkGetTotalFanins(pNtk)
	}
	// perform optimization
	if pPars.Area {
		if pPars.AreaRev {
			Abc_NtkAreaOpt2(p)
		} else {
			Abc_NtkAreaOpt(p)
		}
	} else {
		Abc_NtkDelayOpt(p)
	}
	// record statistics
	if pPars.Verbose {
		p.nTotalNodesEnd = Abc_NtkNodeNum(pNtk)
	}
	if pPars.Verbose {
		p.nTotalEdgesEnd = Abc_NtkGetTotalFanins(pNtk)
	}
	// print stats and quit
	if pPars.Verbose {
		Sfm_DecPrintStats(p)
	}
	if pPars.LibVerbose {
		sfm.Sfm_LibPrint(p.pLib)
	}
	Sfm_DecStop(p)
	if pPars.Area {
		Abc_NtkChangePerform(pNtk, sfmB2I(pPars.Verbose))
	}
}
