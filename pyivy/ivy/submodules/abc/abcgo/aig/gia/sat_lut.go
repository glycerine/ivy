package gia

// Line-by-line port of src/aig/gia/giaSatLut.c (SAT-based LUT remapping,
// '&satlut').

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Sbl_Man_t struct {
	pSat      *bsat.Solver // SAT solver
	vCardVars *Vec_Int_t   // candinality variables
	nVars     int32        // max vars
	LogN      int32        // base-2 log of max vars
	Power2    int32        // power-2 of LogN
	FirstVar  int32        // first variable to be used
	// statistics
	nTried     int32 // nodes tried
	nImproved  int32 // nodes improved
	nRuns      int32 // the number of runs
	nHashWins  int32 // the number of hashed windows
	nSmallWins int32 // the number of small windows
	nLargeWins int32 // the number of large windows
	nIterOuts  int32 // the number of iters exceeded
	// parameters
	LutSize          int32 // LUT size
	nBTLimit         int32 // conflicts
	DelayMax         int32 // external delay
	nEdges           int32 // the number of edges
	fDelay           int32 // delay mode
	fReverse         int32 // reverse windowing
	fVerbose         int32 // verbose
	fVeryVerbose     int32 // verbose
	fVeryVeryVerbose int32 // verbose
	// window
	pGia      *Gia_Man_t
	vLeaves   *Vec_Int_t // leaf nodes
	vAnds     *Vec_Int_t // AND-gates
	vNodes    *Vec_Int_t // internal LUTs
	vRoots    *Vec_Int_t // driver nodes (a subset of vAnds)
	vRootVars *Vec_Int_t // driver nodes (as SAT variables)
	pHash     *vec.Hsh_VecMan_t
	// timing
	vArrs   *Vec_Int_t // arrival times
	vReqs   *Vec_Int_t // required times
	vWindow *Vec_Wec_t // fanins of each node in the window
	vPath   *Vec_Int_t // critical path (as SAT variables)
	vEdges  *Vec_Int_t // fanin edges
	// cuts
	vCutsI1    *Vec_Wrd_t // input bit patterns
	vCutsI2    *Vec_Wrd_t // input bit patterns
	vCutsN1    *Vec_Wrd_t // node bit patterns
	vCutsN2    *Vec_Wrd_t // node bit patterns
	vCutsNum   *Vec_Int_t // cut counts for each obj
	vCutsStart *Vec_Int_t // starting cuts for each obj
	vCutsObj   *Vec_Int_t // object to which this cut belongs
	vTempI1    *Vec_Wrd_t // temporary cuts
	vTempI2    *Vec_Wrd_t // temporary cuts
	vTempN1    *Vec_Wrd_t // temporary cuts
	vTempN2    *Vec_Wrd_t // temporary cuts
	vSolInit   *Vec_Int_t // initial solution
	vSolCur    *Vec_Int_t // current solution
	vSolBest   *Vec_Int_t // best solution
	// temporary
	vLits   *Vec_Int_t // literals
	vAssump *Vec_Int_t // literals
	vPolar  *Vec_Int_t // variables polarity
	// statistics
	timeWin    int64 // windowing
	timeCut    int64 // cut computation
	timeSat    int64 // SAT runtime
	timeSatSat int64 // satisfiable time
	timeSatUns int64 // unsatisfiable time
	timeSatUnd int64 // undecided time
	timeTime   int64 // timing time
	timeStart  int64 // starting time
	timeTotal  int64 // all runtime
	timeOther  int64 // other time
}

func sblLit(v, c int32) int32 { return v + v + c }

func Sbl_ManAlloc(pGia *Gia_Man_t, Number int32) *Sbl_Man_t {
	p := &Sbl_Man_t{}
	p.nVars = Number
	p.LogN = util.Abc_Base2Log(uint32(Number))
	p.Power2 = 1 << p.LogN
	p.pSat = Sbm_AddCardinSolver(p.LogN, &p.vCardVars)
	p.FirstVar = p.pSat.NumVars()
	p.pSat.BookmarkC()
	// window
	p.pGia = pGia
	p.vLeaves = vec.Vec_IntAlloc(int(p.nVars))
	p.vAnds = vec.Vec_IntAlloc(int(p.nVars))
	p.vNodes = vec.Vec_IntAlloc(int(p.nVars))
	p.vRoots = vec.Vec_IntAlloc(int(p.nVars))
	p.vRootVars = vec.Vec_IntAlloc(int(p.nVars))
	p.pHash = vec.Hsh_VecManStart(1000)
	// timing
	p.vArrs = vec.Vec_IntAlloc(0)
	p.vReqs = vec.Vec_IntAlloc(0)
	p.vWindow = vec.Vec_WecAlloc(128)
	p.vPath = vec.Vec_IntAlloc(32)
	p.vEdges = vec.Vec_IntAlloc(32)
	// cuts
	p.vCutsI1 = vec.Vec_WrdAlloc(1000)
	p.vCutsI2 = vec.Vec_WrdAlloc(1000)
	p.vCutsN1 = vec.Vec_WrdAlloc(1000)
	p.vCutsN2 = vec.Vec_WrdAlloc(1000)
	p.vCutsNum = vec.Vec_IntAlloc(64)
	p.vCutsStart = vec.Vec_IntAlloc(64)
	p.vCutsObj = vec.Vec_IntAlloc(1000)
	p.vSolInit = vec.Vec_IntAlloc(64)
	p.vSolCur = vec.Vec_IntAlloc(64)
	p.vSolBest = vec.Vec_IntAlloc(64)
	p.vTempI1 = vec.Vec_WrdAlloc(32)
	p.vTempI2 = vec.Vec_WrdAlloc(32)
	p.vTempN1 = vec.Vec_WrdAlloc(32)
	p.vTempN2 = vec.Vec_WrdAlloc(32)
	// internal
	p.vLits = vec.Vec_IntAlloc(64)
	p.vAssump = vec.Vec_IntAlloc(64)
	p.vPolar = vec.Vec_IntAlloc(1000)
	// other
	Gia_ManFillValue(pGia)
	return p
}
func Sbl_ManClean(p *Sbl_Man_t) {
	p.timeStart = util.Abc_Clock()
	p.pSat.RollbackC()
	p.pSat.BookmarkC()
	// internal
	vec.Vec_IntClear(p.vLeaves)
	vec.Vec_IntClear(p.vAnds)
	vec.Vec_IntClear(p.vNodes)
	vec.Vec_IntClear(p.vRoots)
	vec.Vec_IntClear(p.vRootVars)
	// timing
	vec.Vec_IntClear(p.vArrs)
	vec.Vec_IntClear(p.vReqs)
	vec.Vec_WecClear(p.vWindow)
	vec.Vec_IntClear(p.vPath)
	vec.Vec_IntClear(p.vEdges)
	// cuts
	vec.Vec_WrdClear(p.vCutsI1)
	vec.Vec_WrdClear(p.vCutsI2)
	vec.Vec_WrdClear(p.vCutsN1)
	vec.Vec_WrdClear(p.vCutsN2)
	vec.Vec_IntClear(p.vCutsNum)
	vec.Vec_IntClear(p.vCutsStart)
	vec.Vec_IntClear(p.vCutsObj)
	vec.Vec_IntClear(p.vSolInit)
	vec.Vec_IntClear(p.vSolCur)
	vec.Vec_IntClear(p.vSolBest)
	vec.Vec_WrdClear(p.vTempI1)
	vec.Vec_WrdClear(p.vTempI2)
	vec.Vec_WrdClear(p.vTempN1)
	vec.Vec_WrdClear(p.vTempN2)
	// temporary
	vec.Vec_IntClear(p.vLits)
	vec.Vec_IntClear(p.vAssump)
	vec.Vec_IntClear(p.vPolar)
	// other
	Gia_ManFillValue(p.pGia)
}
func Sbl_ManStop(p *Sbl_Man_t) {
	p.pSat.Delete()
	p.pSat = nil
}

// sblCutFanins pushes the window objects of a cut, in C's bit order.
func sblCutFanins(p *Sbl_Man_t, c int32, vObj *Vec_Int_t) {
	CutI1 := vec.Vec_WrdEntry(p.vCutsI1, int(c))
	CutI2 := vec.Vec_WrdEntry(p.vCutsI2, int(c))
	CutN1 := vec.Vec_WrdEntry(p.vCutsN1, int(c))
	CutN2 := vec.Vec_WrdEntry(p.vCutsN2, int(c))
	for b := 0; b < 64; b++ {
		if (CutI1>>uint(b))&1 != 0 {
			vec.Vec_IntPush(vObj, vec.Vec_IntEntry(p.vLeaves, b))
		}
	}
	for b := 0; b < 64; b++ {
		if (CutI2>>uint(b))&1 != 0 {
			vec.Vec_IntPush(vObj, vec.Vec_IntEntry(p.vLeaves, 64+b))
		}
	}
	for b := 0; b < 64; b++ {
		if (CutN1>>uint(b))&1 != 0 {
			vec.Vec_IntPush(vObj, vec.Vec_IntEntry(p.vAnds, b))
		}
	}
	for b := 0; b < 64; b++ {
		if (CutN2>>uint(b))&1 != 0 {
			vec.Vec_IntPush(vObj, vec.Vec_IntEntry(p.vAnds, 64+b))
		}
	}
}

// Sbl_ManGetCurrentMapping creates fanins for each node in the window.
func Sbl_ManGetCurrentMapping(p *Sbl_Man_t) {
	vec.Vec_WecClear(p.vWindow)
	vec.Vec_WecInit(p.vWindow, vec.Vec_IntSize(p.vAnds))
	for _, c := range vec.Vec_IntArray(p.vSolCur) {
		iObj := vec.Vec_IntEntry(p.vCutsObj, int(c))
		vObj := vec.Vec_WecEntry(p.vWindow, int(iObj))
		vec.Vec_IntClear(vObj)
		sblCutFanins(p, c, vObj)
	}
}

// Sbl_ManComputeDelay is the timing computation.
func Sbl_ManComputeDelay(p *Sbl_Man_t, iLut int32, vFanins *Vec_Int_t) int32 {
	Delay := int32(0)
	for _, iFan := range vec.Vec_IntArray(vFanins) {
		Delay = max(Delay, vec.Vec_IntEntry(p.vArrs, int(iFan))+1)
	}
	return Delay
}
func sblDowndate(v *Vec_Int_t, i, value int32) {
	if vec.Vec_IntEntry(v, int(i)) > value {
		vec.Vec_IntWriteEntry(v, int(i), value)
	}
}
func Sbl_ManCreateTiming(p *Sbl_Man_t, DelayStart int32) int32 {
	DelayMax := DelayStart
	var pTim *tim.Tim_Man_t
	if p.pGia.PManTime != nil {
		pTim = p.pGia.PManTime.(*tim.Tim_Man_t)
	}
	// compute arrival times
	vec.Vec_IntFill(p.vArrs, int(Gia_ManObjNum(p.pGia)), 0)
	if pTim != nil && tim.Tim_ManBoxNum(pTim) != 0 {
		vNodes := Gia_ManOrderWithBoxes(p.pGia)
		tim.Tim_ManIncrementTravId(pTim)
		for _, iLut := range vec.Vec_IntArray(vNodes) {
			pObj := Gia_ManObj(p.pGia, iLut)
			if Gia_ObjIsAnd(pObj) != 0 {
				if Gia_ObjIsLut2(p.pGia, iLut) != 0 {
					Delay := Sbl_ManComputeDelay(p, iLut, Gia_ObjLutFanins2(p.pGia, iLut))
					vec.Vec_IntWriteEntry(p.vArrs, int(iLut), Delay)
					DelayMax = max(DelayMax, Delay)
				}
			} else if Gia_ObjIsCi(pObj) != 0 {
				arrTime := tim.Tim_ManGetCiArrival(pTim, int(Gia_ObjCioId(pObj)))
				vec.Vec_IntWriteEntry(p.vArrs, int(iLut), int32(arrTime))
			} else if Gia_ObjIsCo(pObj) != 0 {
				arrTime := vec.Vec_IntEntry(p.vArrs, int(Gia_ObjFaninId0(pObj, iLut)))
				tim.Tim_ManSetCoArrival(pTim, int(Gia_ObjCioId(pObj)), float32(arrTime))
			}
		}
	} else {
		for iLut := int32(1); iLut < Gia_ManObjNum(p.pGia); iLut++ {
			if Gia_ObjIsLut2(p.pGia, iLut) == 0 {
				continue
			}
			Delay := Sbl_ManComputeDelay(p, iLut, Gia_ObjLutFanins2(p.pGia, iLut))
			vec.Vec_IntWriteEntry(p.vArrs, int(iLut), Delay)
			DelayMax = max(DelayMax, Delay)
		}
	}
	// compute required times
	vec.Vec_IntFill(p.vReqs, int(Gia_ManObjNum(p.pGia)), util.ABC_INFINITY)
	for k := int32(0); k < Gia_ManCoNum(p.pGia); k++ {
		sblDowndate(p.vReqs, Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, k)), DelayMax)
	}
	if pTim != nil && tim.Tim_ManBoxNum(pTim) != 0 {
		vNodes := vec.Vec_IntArray(Gia_ManOrderWithBoxes(p.pGia))
		tim.Tim_ManIncrementTravId(pTim)
		tim.Tim_ManInitPoRequiredAll(pTim, float32(DelayMax))
		// C reuses k for the fanin loop of the reverse traversal
		// (Gia_ManForEachObjVecReverse( vNodes, p->pGia, pObj, k ) with an
		// inner Vec_IntForEachEntry( vFanins, iFan, k )).
		for k := len(vNodes) - 1; k >= 0; k-- {
			iLut := vNodes[k]
			pObj := Gia_ManObj(p.pGia, iLut)
			if Gia_ObjIsAnd(pObj) != 0 {
				if Gia_ObjIsLut2(p.pGia, iLut) != 0 {
					Delay := vec.Vec_IntEntry(p.vReqs, int(iLut)) - 1
					vFanins := vec.Vec_IntArray(Gia_ObjLutFanins2(p.pGia, iLut))
					for k = 0; k < len(vFanins); k++ {
						sblDowndate(p.vReqs, vFanins[k], Delay)
					}
				}
			} else if Gia_ObjIsCi(pObj) != 0 {
				reqTime := vec.Vec_IntEntry(p.vReqs, int(iLut))
				tim.Tim_ManSetCiRequired(pTim, int(Gia_ObjCioId(pObj)), float32(reqTime))
			} else if Gia_ObjIsCo(pObj) != 0 {
				reqTime := tim.Tim_ManGetCoRequired(pTim, int(Gia_ObjCioId(pObj)))
				vec.Vec_IntWriteEntry(p.vReqs, int(Gia_ObjFaninId0(pObj, iLut)), int32(reqTime))
			}
		}
	} else {
		for iLut := Gia_ManObjNum(p.pGia) - 1; iLut > 0; iLut-- {
			if Gia_ObjIsLut2(p.pGia, iLut) == 0 {
				continue
			}
			Delay := vec.Vec_IntEntry(p.vReqs, int(iLut)) - 1
			for _, iFan := range vec.Vec_IntArray(Gia_ObjLutFanins2(p.pGia, iLut)) {
				sblDowndate(p.vReqs, iFan, Delay)
			}
		}
	}
	return DelayMax
}

// Sbl_ManEvaluateMappingEdge checks if the mapping in p->vSolCur meets delay.
func Sbl_ManEvaluateMappingEdge(p *Sbl_Man_t, DelayGlo int32) int32 {
	clk := util.Abc_Clock()
	vec.Vec_IntClear(p.vPath)
	// update new timing
	Sbl_ManGetCurrentMapping(p)
	// derive new timing
	DelayMax := Gia_ManEvalWindow(p.pGia, p.vLeaves, p.vAnds, p.vWindow, p.vPolar, 1)
	p.timeTime += util.Abc_Clock() - clk
	if DelayMax <= DelayGlo {
		return 1
	}
	// create critical path composed of all nodes
	for i := 0; i < vec.Vec_WecSize(p.vWindow); i++ {
		if vec.Vec_IntSize(vec.Vec_WecEntry(p.vWindow, i)) > 0 {
			vec.Vec_IntPush(p.vPath, sblLit(int32(i), 1))
		}
	}
	return 0
}

// Sbl_ManCriticalFanin returns the fanin on the critical path.
func Sbl_ManCriticalFanin(p *Sbl_Man_t, iLut int32, vFanins *Vec_Int_t) int32 {
	Delay := vec.Vec_IntEntry(p.vArrs, int(iLut))
	for _, iFan := range vec.Vec_IntArray(vFanins) {
		if vec.Vec_IntEntry(p.vArrs, int(iFan))+1 == Delay {
			return iFan
		}
	}
	return -1
}

// Sbl_ManEvaluateMapping returns 1 if the mapping satisfies the timing,
// and 0 if a critical path is detected.
func Sbl_ManEvaluateMapping(p *Sbl_Man_t, DelayGlo int32) int32 {
	clk := util.Abc_Clock()
	iLut := int32(-1)
	if p.pGia.VEdge1 != nil {
		return Sbl_ManEvaluateMappingEdge(p, DelayGlo)
	}
	vec.Vec_IntClear(p.vPath)
	// derive timing
	Sbl_ManCreateTiming(p, DelayGlo)
	// update new timing
	Sbl_ManGetCurrentMapping(p)
	for i, id := range vec.Vec_IntArray(p.vAnds) {
		iLut = id
		vFanins := vec.Vec_WecEntry(p.vWindow, i)
		Delay := Sbl_ManComputeDelay(p, iLut, vFanins)
		vec.Vec_IntWriteEntry(p.vArrs, int(iLut), Delay)
	}
	// compare timing at the root nodes
	i := 0
	for ; i < vec.Vec_IntSize(p.vRoots); i++ {
		iLut = vec.Vec_IntEntry(p.vRoots, i)
		Delay := vec.Vec_IntEntry(p.vArrs, int(iLut))
		Required := vec.Vec_IntEntry(p.vReqs, int(iLut))
		if Delay > Required { // updated timing exceeded original timing
			break
		}
	}
	p.timeTime += util.Abc_Clock() - clk
	if i == vec.Vec_IntSize(p.vRoots) {
		return 1
	}
	// derive the critical path
	// find SAT variable of the node whose GIA ID is "iLut"
	iAnd := int32(vec.Vec_IntFind(p.vAnds, iLut))
	for {
		vec.Vec_IntPush(p.vPath, sblLit(iAnd, 1))
		// find fanins of this node
		vFanins := vec.Vec_WecEntry(p.vWindow, int(iAnd))
		// find critical fanin
		iLut = Sbl_ManCriticalFanin(p, iLut, vFanins)
		// find SAT variable of the node whose GIA ID is "iLut"
		iAnd = int32(vec.Vec_IntFind(p.vAnds, iLut))
		if iAnd == -1 {
			break
		}
	}
	return 0
}

func Sbl_ManUpdateMapping(p *Sbl_Man_t) {
	for _, iObj := range vec.Vec_IntArray(p.vAnds) {
		vObj := vec.Vec_WecEntry(p.pGia.vMapping2, int(iObj))
		for _, iTemp := range vec.Vec_IntArray(vObj) {
			Gia_ObjLutRefDecId(p.pGia, iTemp)
		}
		vec.Vec_IntClear(vObj)
	}
	for _, c := range vec.Vec_IntArray(p.vSolBest) {
		iObj := vec.Vec_IntEntry(p.vCutsObj, int(c))
		iObj = vec.Vec_IntEntry(p.vAnds, int(iObj))
		vObj := vec.Vec_WecEntry(p.pGia.vMapping2, int(iObj))
		vec.Vec_IntClear(vObj)
		sblCutFanins(p, c, vObj)
		for _, iTemp := range vec.Vec_IntArray(vObj) {
			Gia_ObjLutRefIncId(p.pGia, iTemp)
		}
	}
}

func Sbl_ManPrintCut(CutI1, CutI2, CutN1, CutN2 word) int32 {
	Count := int32(0)
	fmt.Printf("{ ")
	for b := 0; b < 64; b++ {
		if (CutI1>>uint(b))&1 != 0 {
			fmt.Printf("i%d ", b)
			Count++
		}
	}
	for b := 0; b < 64; b++ {
		if (CutI2>>uint(b))&1 != 0 {
			fmt.Printf("i%d ", 64+b)
			Count++
		}
	}
	fmt.Printf(" ")
	for b := 0; b < 64; b++ {
		if (CutN1>>uint(b))&1 != 0 {
			fmt.Printf("n%d ", b)
			Count++
		}
	}
	for b := 0; b < 64; b++ {
		if (CutN2>>uint(b))&1 != 0 {
			fmt.Printf("n%d ", 64+b)
			Count++
		}
	}
	fmt.Printf("};\n")
	return Count
}
func Sbl_ManFindAndPrintCut(p *Sbl_Man_t, c int32) int32 {
	return Sbl_ManPrintCut(vec.Vec_WrdEntry(p.vCutsI1, int(c)), vec.Vec_WrdEntry(p.vCutsI2, int(c)), vec.Vec_WrdEntry(p.vCutsN1, int(c)), vec.Vec_WrdEntry(p.vCutsN2, int(c)))
}
func Sbl_CutIsFeasible(CutI1, CutI2, CutN1, CutN2 word, LutSize int32) int32 {
	nz := func() int32 {
		return giaBool(CutI1 != 0) + giaBool(CutI2 != 0) + giaBool(CutN1 != 0) + giaBool(CutN2 != 0)
	}
	strip := func() {
		CutI1 &= CutI1 - 1
		CutI2 &= CutI2 - 1
		CutN1 &= CutN1 - 1
		CutN2 &= CutN2 - 1
	}
	Count := nz()
	for i := 0; i < 4; i++ {
		strip()
		Count += nz()
	}
	if LutSize <= 4 {
		return giaBool(Count <= 4)
	}
	for i := 0; i < 2; i++ {
		strip()
		Count += nz()
	}
	return giaBool(Count <= 6)
}
func Sbl_CutPushUncontained(vCutsI1, vCutsI2, vCutsN1, vCutsN2 *Vec_Wrd_t, CutI1, CutI2, CutN1, CutN2 word) int32 {
	k := 0
	a1, a2, b1, b2 := vec.Vec_WrdArray(vCutsI1), vec.Vec_WrdArray(vCutsI2), vec.Vec_WrdArray(vCutsN1), vec.Vec_WrdArray(vCutsN2)
	n := len(a1)
	for i := 0; i < n; i++ {
		if a1[i]&CutI1 == a1[i] && a2[i]&CutI2 == a2[i] && b1[i]&CutN1 == b1[i] && b2[i]&CutN2 == b2[i] {
			return 1
		}
	}
	for i := 0; i < n; i++ {
		if a1[i]&CutI1 != CutI1 || a2[i]&CutI2 != CutI2 || b1[i]&CutN1 != CutN1 || b2[i]&CutN2 != CutN2 {
			vec.Vec_WrdWriteEntry(vCutsI1, k, a1[i])
			vec.Vec_WrdWriteEntry(vCutsI2, k, a2[i])
			vec.Vec_WrdWriteEntry(vCutsN1, k, b1[i])
			vec.Vec_WrdWriteEntry(vCutsN2, k, b2[i])
			k++
		}
	}
	vec.Vec_WrdShrink(vCutsI1, k)
	vec.Vec_WrdShrink(vCutsI2, k)
	vec.Vec_WrdShrink(vCutsN1, k)
	vec.Vec_WrdShrink(vCutsN2, k)
	vec.Vec_WrdPush(vCutsI1, CutI1)
	vec.Vec_WrdPush(vCutsI2, CutI2)
	vec.Vec_WrdPush(vCutsN1, CutN1)
	vec.Vec_WrdPush(vCutsN2, CutN2)
	return 0
}
func Sbl_ManComputeCutsOne(p *Sbl_Man_t, Fan0, Fan1, Obj int32) {
	pCutsI1 := vec.Vec_WrdArray(p.vCutsI1)
	pCutsI2 := vec.Vec_WrdArray(p.vCutsI2)
	pCutsN1 := vec.Vec_WrdArray(p.vCutsN1)
	pCutsN2 := vec.Vec_WrdArray(p.vCutsN2)
	Start0 := vec.Vec_IntEntry(p.vCutsStart, int(Fan0))
	Start1 := vec.Vec_IntEntry(p.vCutsStart, int(Fan1))
	Limit0 := Start0 + vec.Vec_IntEntry(p.vCutsNum, int(Fan0))
	Limit1 := Start1 + vec.Vec_IntEntry(p.vCutsNum, int(Fan1))
	vec.Vec_WrdClear(p.vTempI1)
	vec.Vec_WrdClear(p.vTempI2)
	vec.Vec_WrdClear(p.vTempN1)
	vec.Vec_WrdClear(p.vTempN2)
	for i := Start0; i < Limit0; i++ {
		for k := Start1; k < Limit1; k++ {
			if Sbl_CutIsFeasible(pCutsI1[i]|pCutsI1[k], pCutsI2[i]|pCutsI2[k], pCutsN1[i]|pCutsN1[k], pCutsN2[i]|pCutsN2[k], p.LutSize) != 0 {
				Sbl_CutPushUncontained(p.vTempI1, p.vTempI2, p.vTempN1, p.vTempN2, pCutsI1[i]|pCutsI1[k], pCutsI2[i]|pCutsI2[k], pCutsN1[i]|pCutsN1[k], pCutsN2[i]|pCutsN2[k])
			}
		}
	}
	vec.Vec_IntPush(p.vCutsStart, int32(vec.Vec_WrdSize(p.vCutsI1)))
	vec.Vec_IntPush(p.vCutsNum, int32(vec.Vec_WrdSize(p.vTempI1)+1))
	vec.Vec_WrdAppend(p.vCutsI1, p.vTempI1)
	vec.Vec_WrdAppend(p.vCutsI2, p.vTempI2)
	vec.Vec_WrdAppend(p.vCutsN1, p.vTempN1)
	vec.Vec_WrdAppend(p.vCutsN2, p.vTempN2)
	vec.Vec_WrdPush(p.vCutsI1, 0)
	vec.Vec_WrdPush(p.vCutsI2, 0)
	if Obj < 64 {
		vec.Vec_WrdPush(p.vCutsN1, word(1)<<uint(Obj))
		vec.Vec_WrdPush(p.vCutsN2, 0)
	} else {
		vec.Vec_WrdPush(p.vCutsN1, 0)
		vec.Vec_WrdPush(p.vCutsN2, word(1)<<uint(Obj-64))
	}
	for i := 0; i <= vec.Vec_WrdSize(p.vTempI1); i++ {
		vec.Vec_IntPush(p.vCutsObj, Obj)
	}
}
func Sbl_ManFindCut(p *Sbl_Man_t, Obj int32, CutI1, CutI2, CutN1, CutN2 word) int32 {
	pCutsI1 := vec.Vec_WrdArray(p.vCutsI1)
	pCutsI2 := vec.Vec_WrdArray(p.vCutsI2)
	pCutsN1 := vec.Vec_WrdArray(p.vCutsN1)
	pCutsN2 := vec.Vec_WrdArray(p.vCutsN2)
	Start0 := vec.Vec_IntEntry(p.vCutsStart, int(Obj))
	Limit0 := Start0 + vec.Vec_IntEntry(p.vCutsNum, int(Obj))
	for i := Start0; i < Limit0; i++ {
		if pCutsI1[i] == CutI1 && pCutsI2[i] == CutI2 && pCutsN1[i] == CutN1 && pCutsN2[i] == CutN2 {
			return i
		}
	}
	return -1
}
func Sbl_ManComputeCuts(p *Sbl_Man_t) int32 {
	clk := util.Abc_Clock()
	nLeaves := uint32(vec.Vec_IntSize(p.vLeaves))
	// assign leaf cuts
	vec.Vec_IntClear(p.vCutsStart)
	vec.Vec_IntClear(p.vCutsObj)
	vec.Vec_IntClear(p.vCutsNum)
	vec.Vec_WrdClear(p.vCutsI1)
	vec.Vec_WrdClear(p.vCutsI2)
	vec.Vec_WrdClear(p.vCutsN1)
	vec.Vec_WrdClear(p.vCutsN2)
	for i, id := range vec.Vec_IntArray(p.vLeaves) {
		pObj := Gia_ManObj(p.pGia, id)
		vec.Vec_IntPush(p.vCutsStart, int32(vec.Vec_WrdSize(p.vCutsI1)))
		vec.Vec_IntPush(p.vCutsObj, -1)
		vec.Vec_IntPush(p.vCutsNum, 1)
		if i < 64 {
			vec.Vec_WrdPush(p.vCutsI1, word(1)<<uint(i))
			vec.Vec_WrdPush(p.vCutsI2, 0)
		} else {
			vec.Vec_WrdPush(p.vCutsI1, 0)
			vec.Vec_WrdPush(p.vCutsI2, word(1)<<uint(i-64))
		}
		vec.Vec_WrdPush(p.vCutsN1, 0)
		vec.Vec_WrdPush(p.vCutsN2, 0)
		pObj.Value = uint32(i)
	}
	// assign internal cuts
	for i, id := range vec.Vec_IntArray(p.vAnds) {
		pObj := Gia_ManObj(p.pGia, id)
		Sbl_ManComputeCutsOne(p, int32(Gia_ObjFanin0(pObj).Value), int32(Gia_ObjFanin1(pObj).Value), int32(i))
		pObj.Value = nLeaves + uint32(i)
	}
	// check that roots are mapped nodes
	vec.Vec_IntClear(p.vRootVars)
	for _, id := range vec.Vec_IntArray(p.vRoots) {
		pObj := Gia_ManObj(p.pGia, id)
		if Gia_ObjIsCi(pObj) != 0 {
			continue
		}
		vec.Vec_IntPush(p.vRootVars, int32(pObj.Value-nLeaves))
	}
	// create current solution
	vec.Vec_IntClear(p.vPolar)
	vec.Vec_IntClear(p.vSolInit)
	for i, Obj := range vec.Vec_IntArray(p.vAnds) {
		var CutI1, CutI2, CutN1, CutN2 word
		if Gia_ObjIsLut2(p.pGia, Obj) == 0 {
			continue
		}
		// add node
		vec.Vec_IntPush(p.vPolar, int32(i))
		vec.Vec_IntPush(p.vSolInit, int32(i))
		// add its cut
		for _, Fanin := range vec.Vec_IntArray(Gia_ObjLutFanins2(p.pGia, Obj)) {
			pFanin := Gia_ManObj(p.pGia, Fanin)
			if ^pFanin.Value == 0 {
				continue
			}
			if pFanin.Value < nLeaves {
				if pFanin.Value < 64 {
					CutI1 |= word(1) << pFanin.Value
				} else {
					CutI2 |= word(1) << (pFanin.Value - 64)
				}
			} else {
				if pFanin.Value-nLeaves < 64 {
					CutN1 |= word(1) << (pFanin.Value - nLeaves)
				} else {
					CutN2 |= word(1) << (pFanin.Value - nLeaves - 64)
				}
			}
		}
		// find the new cut
		Index := Sbl_ManFindCut(p, int32(nLeaves)+int32(i), CutI1, CutI2, CutN1, CutN2)
		if Index < 0 {
			continue
		}
		vec.Vec_IntPush(p.vPolar, p.FirstVar+Index)
	}
	// clean value
	for _, id := range vec.Vec_IntArray(p.vLeaves) {
		Gia_ManObj(p.pGia, id).Value = ^uint32(0)
	}
	for _, id := range vec.Vec_IntArray(p.vAnds) {
		Gia_ManObj(p.pGia, id).Value = ^uint32(0)
	}
	p.timeCut += util.Abc_Clock() - clk
	return int32(vec.Vec_WrdSize(p.vCutsI1))
}
func Sbl_ManCreateCnf(p *Sbl_Man_t) int32 {
	var pLits [2]int32
	pCutsN1 := vec.Vec_WrdArray(p.vCutsN1)
	pCutsN2 := vec.Vec_WrdArray(p.vCutsN2)
	p.pSat.SetNumVars(p.pSat.NumVars() + int32(vec.Vec_WrdSize(p.vCutsI1)))
	for i := int32(0); i < int32(vec.Vec_IntSize(p.vAnds)); i++ {
		Start0 := vec.Vec_IntEntry(p.vCutsStart, vec.Vec_IntSize(p.vLeaves)+int(i))
		Limit0 := Start0 + vec.Vec_IntEntry(p.vCutsNum, vec.Vec_IntSize(p.vLeaves)+int(i)) - 1
		// add main clause
		vec.Vec_IntClear(p.vLits)
		vec.Vec_IntPush(p.vLits, sblLit(i, 1))
		for k := Start0; k < Limit0; k++ {
			vec.Vec_IntPush(p.vLits, sblLit(p.FirstVar+k, 0))
		}
		p.pSat.AddClauseC(vec.Vec_IntArray(p.vLits))
		// binary clauses
		for k := Start0; k < Limit0; k++ {
			Cut1 := pCutsN1[k]
			Cut2 := pCutsN2[k]
			// root clause
			pLits[0] = sblLit(p.FirstVar+k, 1)
			pLits[1] = sblLit(i, 0)
			p.pSat.AddClauseC(pLits[:])
			// fanin clauses
			for c := int32(0); c < 64 && Cut1 != 0; c, Cut1 = c+1, Cut1>>1 {
				if Cut1&1 == 0 {
					continue
				}
				pLits[1] = sblLit(c, 0)
				p.pSat.AddClauseC(pLits[:])
			}
			for c := int32(0); c < 64 && Cut2 != 0; c, Cut2 = c+1, Cut2>>1 {
				if Cut2&1 == 0 {
					continue
				}
				pLits[1] = sblLit(c+64, 0)
				p.pSat.AddClauseC(pLits[:])
			}
		}
	}
	p.pSat.SetPolarityC(vec.Vec_IntArray(p.vPolar))
	return 1
}

// Sbl_ManWindow uses the entire graph as the SAT window.
func Sbl_ManWindow(p *Sbl_Man_t) {
	// collect leaves
	vec.Vec_IntClear(p.vLeaves)
	for i := int32(0); i < Gia_ManCiNum(p.pGia); i++ {
		vec.Vec_IntPush(p.vLeaves, Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i)))
	}
	// collect internal
	vec.Vec_IntClear(p.vAnds)
	for id := int32(1); id < p.pGia.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, id)) != 0 {
			vec.Vec_IntPush(p.vAnds, id)
		}
	}
	// collect roots
	vec.Vec_IntClear(p.vRoots)
	for i := int32(0); i < Gia_ManCoNum(p.pGia); i++ {
		vec.Vec_IntPush(p.vRoots, Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i)))
	}
}
func Sbl_ManWindow2(p *Sbl_Man_t, iPivot int32) int32 {
	clk := util.Abc_Clock()
	var vRoots, vNodes, vLeaves, vAnds *Vec_Int_t
	Count := Gia_ManComputeOneWin(p.pGia, iPivot, &vRoots, &vNodes, &vLeaves, &vAnds)
	p.timeWin += util.Abc_Clock() - clk
	if Count == 0 {
		return 0
	}
	vec.Vec_IntClear(p.vRoots)
	vec.Vec_IntAppend(p.vRoots, vRoots)
	vec.Vec_IntClear(p.vNodes)
	vec.Vec_IntAppend(p.vNodes, vNodes)
	vec.Vec_IntClear(p.vLeaves)
	vec.Vec_IntAppend(p.vLeaves, vLeaves)
	vec.Vec_IntClear(p.vAnds)
	vec.Vec_IntAppend(p.vAnds, vAnds)
	return Count
}

func Sbl_ManTestSat(p *Sbl_Man_t, iPivot int32) int32 {
	fKeepTrying := int32(1)
	clk := util.Abc_Clock()
	var clk2 int64
	nConfTotal, nIters := int32(0), int32(0)
	nEntries := vec.Hsh_VecSize(p.pHash)
	p.nTried++

	Sbl_ManClean(p)

	// compute one window
	Count := Sbl_ManWindow2(p, iPivot)
	if Count == 0 {
		if p.fVeryVerbose != 0 {
			fmt.Printf("Obj %d: Window with less than %d nodes does not exist.\n", iPivot, p.nVars)
		}
		p.nSmallWins++
		return 0
	}
	vec.Hsh_VecManAdd(p.pHash, p.vAnds)
	if nEntries == vec.Hsh_VecSize(p.pHash) {
		if p.fVeryVerbose != 0 {
			fmt.Printf("Obj %d: This window was already tried.\n", iPivot)
		}
		p.nHashWins++
		return 0
	}
	if p.fVeryVerbose != 0 {
		fmt.Printf("\nObj = %6d : Leaf = %2d.  AND = %2d.  Root = %2d.    LUT = %2d.\n",
			iPivot, vec.Vec_IntSize(p.vLeaves), vec.Vec_IntSize(p.vAnds), vec.Vec_IntSize(p.vRoots), vec.Vec_IntSize(p.vNodes))
	}

	if vec.Vec_IntSize(p.vLeaves) > 128 || vec.Vec_IntSize(p.vAnds) > int(p.nVars) {
		if p.fVeryVerbose != 0 {
			fmt.Printf("Obj %d: Encountered window with %d inputs and %d internal nodes.\n", iPivot, vec.Vec_IntSize(p.vLeaves), vec.Vec_IntSize(p.vAnds))
		}
		p.nLargeWins++
		return 0
	}
	if vec.Vec_IntSize(p.vAnds) < 10 {
		if p.fVeryVerbose != 0 {
			fmt.Printf("Skipping.\n")
		}
		return 0
	}

	// derive cuts
	Sbl_ManComputeCuts(p)
	// derive SAT instance
	Sbl_ManCreateCnf(p)

	if p.fVeryVeryVerbose != 0 {
		fmt.Printf("All clauses = %d.  Multi clauses = %d.  Binary clauses = %d.  Other clauses = %d.\n\n",
			p.pSat.NClauses(), vec.Vec_IntSize(p.vAnds), vec.Vec_WrdSize(p.vCutsI1)-vec.Vec_IntSize(p.vAnds),
			int(p.pSat.NClauses())-vec.Vec_WrdSize(p.vCutsI1))
	}

	// create assumptions
	// cardinality
	vec.Vec_IntClear(p.vAssump)
	vec.Vec_IntPush(p.vAssump, -1)
	// unused variables
	for i := int32(vec.Vec_IntSize(p.vAnds)); i < p.Power2; i++ {
		vec.Vec_IntPush(p.vAssump, sblLit(i, 1))
	}
	// root variables
	for _, Root := range vec.Vec_IntArray(p.vRootVars) {
		vec.Vec_IntPush(p.vAssump, sblLit(Root, 0))
	}

	StartSol := int32(vec.Vec_IntSize(p.vSolInit)) + 1
	for fKeepTrying != 0 && StartSol-fKeepTrying > 0 {
		Count, LitCount := int32(0), int32(0)
		if p.fVeryVerbose != 0 {
			fmt.Printf("Trying to find mapping with %d LUTs.\n", StartSol-fKeepTrying)
		}
		vec.Vec_IntWriteEntry(p.vAssump, 0, sblLit(vec.Vec_IntEntry(p.vCardVars, int(StartSol-fKeepTrying)), 1))
		// solve the problem
		clk2 = util.Abc_Clock()
		nConfBef := p.pSat.NConflicts()
		status := p.pSat.SolveWithResourceLimits(vec.Vec_IntArray(p.vAssump), int64(p.nBTLimit), 0, 0, 0)
		p.timeSat += util.Abc_Clock() - clk2
		nConfAft := p.pSat.NConflicts()
		nConfTotal += nConfAft - nConfBef
		nIters++
		p.nRuns++
		if status == bsat.Satisfiable {
			p.timeSatSat += util.Abc_Clock() - clk2
		} else if status == bsat.Unsatisfiable {
			p.timeSatUns += util.Abc_Clock() - clk2
		} else {
			p.timeSatUnd += util.Abc_Clock() - clk2
		}
		if status == bsat.Unknown {
			break
		}
		if status == bsat.Satisfiable {
			if p.fVeryVeryVerbose != 0 {
				for i := 0; i < vec.Vec_IntSize(p.vAnds); i++ {
					fmt.Printf("%d", giaBool(p.pSat.VarValue(int32(i))))
				}
				fmt.Printf("\n")
				for i := 0; i < vec.Vec_IntSize(p.vAnds); i++ {
					if p.pSat.VarValue(int32(i)) {
						fmt.Printf("%d=%d ", i, 1)
						Count++
					}
				}
				fmt.Printf("Count = %d\n", Count)
			}
			Count = 1
			vec.Vec_IntClear(p.vSolCur)
			for i := p.FirstVar; i < p.pSat.NumVars(); i++ {
				if p.pSat.VarValue(i) {
					if p.fVeryVeryVerbose != 0 {
						fmt.Printf("Cut %3d : Node = %3d %6d  ", Count, vec.Vec_IntEntry(p.vCutsObj, int(i-p.FirstVar)), vec.Vec_IntEntry(p.vAnds, int(vec.Vec_IntEntry(p.vCutsObj, int(i-p.FirstVar)))))
						Count++
						LitCount += Sbl_ManFindAndPrintCut(p, i-p.FirstVar)
					}
					vec.Vec_IntPush(p.vSolCur, i-p.FirstVar)
				}
			}
		}

		// check the timing
		if status == bsat.Satisfiable {
			if p.fDelay != 0 && Sbl_ManEvaluateMapping(p, p.DelayMax) == 0 {
				if p.fVeryVerbose != 0 {
					fmt.Printf("Critical path of length (%d) is detected:   ", vec.Vec_IntSize(p.vPath))
					for i, iLit := range vec.Vec_IntArray(p.vPath) {
						fmt.Printf("%d=%d ", i, vec.Vec_IntEntry(p.vAnds, int(iLit>>1)))
					}
					fmt.Printf("\n")
				}
				// add the clause
				p.pSat.AddClauseC(vec.Vec_IntArray(p.vPath))
			} else {
				vec.Vec_IntClear(p.vSolBest)
				vec.Vec_IntAppend(p.vSolBest, p.vSolCur)
				fKeepTrying++
			}
		} else {
			fKeepTrying = 0
		}
		if p.fVeryVerbose != 0 {
			if status == bsat.Unsatisfiable {
				fmt.Printf("UNSAT ")
			} else if status == bsat.Satisfiable {
				fmt.Printf("SAT   ")
			} else {
				fmt.Printf("UNDEC ")
			}
			fmt.Printf("confl =%8d.    ", nConfAft-nConfBef)
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk2)

			fmt.Printf("Total ")
			fmt.Printf("confl =%8d.    ", nConfTotal)
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			if p.fVeryVeryVerbose != 0 && status == bsat.Satisfiable {
				fmt.Printf("LitCount = %d.\n", LitCount)
			}
			fmt.Printf("\n")
		}
		if nIters == 10 {
			p.nIterOuts++
			break
		}
	}

	// update solution
	if vec.Vec_IntSize(p.vSolBest) > 0 && vec.Vec_IntSize(p.vSolBest) < vec.Vec_IntSize(p.vSolInit) {
		var nDelayCur, nEdgesCur int32
		Sbl_ManUpdateMapping(p)
		if p.pGia.VEdge1 != nil {
			nDelayCur = Gia_ManEvalEdgeDelay(p.pGia)
			nEdgesCur = Gia_ManEvalEdgeCount(p.pGia)
		} else {
			nDelayCur = Sbl_ManCreateTiming(p, p.DelayMax)
		}
		if p.fVerbose != 0 {
			fmt.Printf("Object %5d : Saved %2d nodes  (Conf =%8d)  Iter =%3d  Delay = %d  Edges = %4d\n",
				iPivot, vec.Vec_IntSize(p.vSolInit)-vec.Vec_IntSize(p.vSolBest), nConfTotal, nIters, nDelayCur, nEdgesCur)
		}
		p.timeTotal += util.Abc_Clock() - p.timeStart
		p.nImproved++
		return 2
	}
	p.timeTotal += util.Abc_Clock() - p.timeStart
	return 1
}
func Sbl_ManPrintRuntime(p *Sbl_Man_t) {
	fmt.Printf("Runtime breakdown:\n")
	p.timeOther = p.timeTotal - p.timeWin - p.timeCut - p.timeSat - p.timeTime
	util.Abc_PrintTimeP(1, "Win   ", p.timeWin, p.timeTotal)
	util.Abc_PrintTimeP(1, "Cut   ", p.timeCut, p.timeTotal)
	util.Abc_PrintTimeP(1, "Sat   ", p.timeSat, p.timeTotal)
	util.Abc_PrintTimeP(1, " Sat  ", p.timeSatSat, p.timeTotal)
	util.Abc_PrintTimeP(1, " Unsat", p.timeSatUns, p.timeTotal)
	util.Abc_PrintTimeP(1, " Undec", p.timeSatUnd, p.timeTotal)
	util.Abc_PrintTimeP(1, "Timing", p.timeTime, p.timeTotal)
	util.Abc_PrintTimeP(1, "Other ", p.timeOther, p.timeTotal)
	util.Abc_PrintTimeP(1, "ALL   ", p.timeTotal, p.timeTotal)
}
func Gia_ManLutSat(pGia *Gia_Man_t, LutSize, nNumber, nImproves, nBTLimit, DelayMax, nEdges, fDelay, fReverse, fVerbose, fVeryVerbose int32) {
	nImproveCount := int32(0)
	p := Sbl_ManAlloc(pGia, nNumber)
	p.LutSize = LutSize   // LUT size
	p.nBTLimit = nBTLimit // conflicts
	p.DelayMax = DelayMax // external delay
	p.nEdges = nEdges     // the number of edges
	p.fDelay = fDelay     // delay mode
	p.fReverse = fReverse // reverse windowing
	p.fVerbose = fVerbose | fVeryVerbose
	p.fVeryVerbose = fVeryVerbose
	if p.fVerbose != 0 {
		fmt.Printf("Parameters: WinSize = %d AIG nodes.  Conf = %d.  DelayMax = %d.\n", p.nVars, p.nBTLimit, p.DelayMax)
	}
	// determine delay limit
	if fDelay != 0 && pGia.VEdge1 != nil && p.DelayMax == 0 {
		p.DelayMax = Gia_ManEvalEdgeDelay(pGia)
	}
	// iterate through the internal nodes
	Gia_ManComputeOneWinStart(pGia, nNumber, fReverse)
	for iLut := int32(1); iLut < Gia_ManObjNum(pGia); iLut++ {
		if Gia_ObjIsLut2(pGia, iLut) == 0 {
			continue
		}
		if Sbl_ManTestSat(p, iLut) != 2 {
			continue
		}
		nImproveCount++
		if nImproveCount == nImproves {
			break
		}
	}
	Gia_ManComputeOneWin(pGia, -1, nil, nil, nil, nil)
	if p.fVerbose != 0 {
		fmt.Printf("Tried = %d. Used = %d. HashWin = %d. SmallWin = %d. LargeWin = %d. IterOut = %d.  SAT runs = %d.\n",
			p.nTried, p.nImproved, p.nHashWins, p.nSmallWins, p.nLargeWins, p.nIterOuts, p.nRuns)
	}
	if p.fVerbose != 0 {
		Sbl_ManPrintRuntime(p)
	}
	Sbl_ManStop(p)
	pGia.VPacking = nil
}
