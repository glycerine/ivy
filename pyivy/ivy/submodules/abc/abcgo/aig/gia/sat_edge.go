package gia

// Line-by-line port of src/aig/gia/giaSatEdge.c (SAT-based edge assignment,
// the default algorithm of '&edge').

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Seg_Man_t struct {
	pSat     *bsat.Solver // SAT solver
	nVars    int32        // max vars (edge num)
	LogN     int32        // base-2 log of max vars
	Power2   int32        // power-2 of LogN
	FirstVar int32        // first variable to be used
	// parameters
	nBTLimit int32 // conflicts
	DelayMax int32 // external delay
	nEdges   int32 // the number of edges
	fDelay   int32 // delay mode
	fReverse int32 // reverse windowing
	fVerbose int32 // verbose
	// window
	pGia    *Gia_Man_t
	vPolars *Vec_Int_t // polarity
	vToSkip *Vec_Int_t // edges to skip
	vEdges  *Vec_Int_t // edges as fanin/fanout pairs
	vFirsts *Vec_Int_t // first SAT variable
	vNvars  *Vec_Int_t // number of SAT variables
	vLits   *Vec_Int_t // literals
	pLevels []int32    // levels
	// statistics
	timeStart int64
}

func segLit(v, c int32) int32 { return v + v + c }

// Seg_ManCountIntEdges counts the number of edges between internal nodes.
func Seg_ManCountIntEdges(p *Gia_Man_t, vPolars, vToSkip *Vec_Int_t, nFanouts int32) *Vec_Int_t {
	vEdges := vec.Vec_IntAlloc(1000)
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: SAT edge assignment requires mapping")
	}
	vec.Vec_IntClear(vPolars)
	vec.Vec_IntClear(vToSkip)
	if nFanouts != 0 {
		Gia_ManSetLutRefs(p)
	}
	for iLut := int32(1); iLut < p.nObjs; iLut++ {
		if Gia_ObjIsLut(p, iLut) == 0 {
			continue
		}
		for _, iFanin := range Gia_ObjLutFanins(p, iLut) {
			if Gia_ObjIsAnd(Gia_ManObj(p, iFanin)) != 0 {
				if p.VEdge1 != nil && Gia_ObjCheckEdge(p, iFanin, iLut) != 0 {
					vec.Vec_IntPush(vPolars, int32(vec.Vec_IntSize(vEdges)/2))
				}
				if nFanouts != 0 && Gia_ObjLutRefNumId(p, iFanin) >= nFanouts {
					vec.Vec_IntPush(vToSkip, int32(vec.Vec_IntSize(vEdges)/2))
				}
				vec.Vec_IntPushTwo(vEdges, iFanin, iLut)
			}
		}
	}
	if nFanouts != 0 {
		p.pLutRefs = nil
	}
	return vEdges
}
func Seg_ManCollectObjEdges(vEdges *Vec_Int_t, nObjs int32) *Vec_Wec_t {
	vRes := vec.Vec_WecStart(int(nObjs))
	data := vec.Vec_IntArray(vEdges)
	for i := 0; i+1 < len(data); i += 2 {
		vec.Vec_WecPush(vRes, int(data[i]), int32(i/2))
		vec.Vec_WecPush(vRes, int(data[i+1]), int32(i/2))
	}
	return vRes
}
func Seg_ManCountIntLevels(p *Seg_Man_t, iStartVar int32) int32 {
	vec.Vec_IntFill(p.vFirsts, int(Gia_ManObjNum(p.pGia)), -1)
	vec.Vec_IntFill(p.vNvars, int(Gia_ManObjNum(p.pGia)), 0)
	p.pLevels = nil
	if p.pGia.PManTime != nil {
		p.DelayMax = Gia_ManLutLevelWithBoxes(p.pGia)
		p.pLevels = vec.Vec_IntArray(p.pGia.vLevels)
		p.pGia.vLevels = nil
	} else {
		p.DelayMax = Gia_ManLutLevel(p.pGia, &p.pLevels)
	}
	for iLut := int32(1); iLut < p.pGia.nObjs; iLut++ {
		pObj := Gia_ManObj(p.pGia, iLut)
		if Gia_ObjIsCo(pObj) != 0 {
			continue
		}
		if Gia_ObjIsAnd(pObj) != 0 && Gia_ObjIsLut(p.pGia, iLut) == 0 {
			continue
		}
		vec.Vec_IntWriteEntry(p.vFirsts, int(iLut), iStartVar)
		nVars := p.pLevels[iLut]
		if nVars < 2 {
			nVars = 0
		}
		vec.Vec_IntWriteEntry(p.vNvars, int(iLut), nVars)
		iStartVar += nVars
	}
	return iStartVar
}

func Seg_ManAlloc(pGia *Gia_Man_t, nFanouts int32) *Seg_Man_t {
	p := &Seg_Man_t{}
	p.vPolars = vec.Vec_IntAlloc(1000)
	p.vToSkip = vec.Vec_IntAlloc(1000)
	p.vEdges = Seg_ManCountIntEdges(pGia, p.vPolars, p.vToSkip, nFanouts)
	p.nVars = int32(vec.Vec_IntSize(p.vEdges) / 2)
	p.LogN = util.Abc_Base2Log(uint32(p.nVars))
	p.Power2 = 1 << p.LogN
	p.pSat = bsat.NewSolver()
	p.pSat.SetNumVars(p.nVars)
	p.FirstVar = p.pSat.NumVars()
	p.pSat.BookmarkC()
	p.pGia = pGia
	// internal
	p.vFirsts = vec.Vec_IntAlloc(0)
	p.vNvars = vec.Vec_IntAlloc(0)
	p.vLits = vec.Vec_IntAlloc(0)
	nVarsAll := Seg_ManCountIntLevels(p, p.FirstVar)
	p.pSat.SetNumVars(nVarsAll)
	// other
	Gia_ManFillValue(pGia)
	return p
}
func Seg_ManClean(p *Seg_Man_t) {
	p.timeStart = util.Abc_Clock()
	p.pSat.RollbackC()
	p.pSat.BookmarkC()
	// internal
	vec.Vec_IntClear(p.vEdges)
	vec.Vec_IntClear(p.vFirsts)
	vec.Vec_IntClear(p.vNvars)
	vec.Vec_IntClear(p.vLits)
	vec.Vec_IntClear(p.vPolars)
	vec.Vec_IntClear(p.vToSkip)
	// other
	Gia_ManFillValue(p.pGia)
}
func Seg_ManStop(p *Seg_Man_t) {
	p.pSat.Delete()
	p.pSat = nil
	p.pLevels = nil
}

func Seg_ManCreateCnf(p *Seg_Man_t, fTwo, fVerbose int32) {
	var pTim *tim.Tim_Man_t
	if p.pGia.PManTime != nil {
		pTim = p.pGia.PManTime.(*tim.Tim_Man_t)
	}
	var pLits [3]int32
	Count := int32(0)
	clk := util.Abc_Clock()
	// C asserts that every clause is accepted; release builds go on.
	addClause := func(lits []int32) { p.pSat.AddClauseC(lits) }
	// edge delay constraints
	nConstr := p.pSat.NClauses()
	for iLut := int32(0); iLut < p.pGia.nObjs; iLut++ {
		pObj := Gia_ManObj(p.pGia, iLut)
		iFirstLut := vec.Vec_IntEntry(p.vFirsts, int(iLut))
		nVarsLut := vec.Vec_IntEntry(p.vNvars, int(iLut))
		if pTim != nil && Gia_ObjIsCi(pObj) != 0 {
			iBox := tim.Tim_ManBoxForCi(pTim, int(Gia_ObjCioId(pObj)))
			if nVarsLut > 0 && iBox >= 0 {
				iCiId := int32(tim.Tim_ManBoxOutputFirst(pTim, iBox))
				if iCiId == Gia_ObjCioId(pObj) { // first input
					nIns := int32(tim.Tim_ManBoxInputNum(pTim, iBox))
					iIn0 := int32(tim.Tim_ManBoxInputFirst(pTim, iBox))
					for i := int32(0); i < nIns-1; i++ { // skip carry-in pin
						pOut := Gia_ManCo(p.pGia, iIn0+i)
						iDriverId := Gia_ObjFaninId0p(p.pGia, pOut)
						iFirst := vec.Vec_IntEntry(p.vFirsts, int(iDriverId))
						nVars := vec.Vec_IntEntry(p.vNvars, int(iDriverId))
						AddOn := int32(0)
						if nVars < nVarsLut {
							AddOn = 1
						}
						for k := int32(0); k < nVars; k++ {
							pLits[0] = segLit(iFirst+k, 1)
							pLits[1] = segLit(iFirstLut+k+AddOn, 0)
							addClause(pLits[:2])
						}
					}
				} else { // intermediate input
					pIn := Gia_ManCi(p.pGia, iCiId)
					iObjId := Gia_ObjId(p.pGia, pIn)
					iFirst := vec.Vec_IntEntry(p.vFirsts, int(iObjId))
					nVars := vec.Vec_IntEntry(p.vNvars, int(iObjId))
					if nVars > 0 {
						for k := int32(0); k < nVars; k++ {
							pLits[0] = segLit(iFirst+k, 1)
							pLits[1] = segLit(iFirstLut+k, 0)
							addClause(pLits[:2])
						}
					}
				}
			}
			continue
		}
		if Gia_ObjIsLut(p.pGia, iLut) == 0 {
			continue
		}
		for _, iFanin := range Gia_ObjLutFanins(p.pGia, iLut) {
			if pTim != nil && Gia_ObjIsCi(Gia_ManObj(p.pGia, iFanin)) != 0 {
				iFirst := vec.Vec_IntEntry(p.vFirsts, int(iFanin))
				nVars := vec.Vec_IntEntry(p.vNvars, int(iFanin))
				if nVars > 0 {
					for k := int32(0); k < nVars; k++ {
						pLits[0] = segLit(iFirst+k, 1)
						pLits[1] = segLit(iFirstLut+k, 0)
						addClause(pLits[:2])
					}
				}
			} else if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iFanin)) != 0 {
				iFirst := vec.Vec_IntEntry(p.vFirsts, int(iFanin))
				nVars := vec.Vec_IntEntry(p.vNvars, int(iFanin))
				// add initial
				if nVars == 0 {
					pLits[0] = segLit(Count, 1)
					pLits[1] = segLit(iFirstLut, 0)
					addClause(pLits[:2])

					pLits[0] = segLit(Count, 0)
					pLits[1] = segLit(iFirstLut+1, 0)
					addClause(pLits[:2])
				}
				// add others
				for k := int32(0); k < nVars; k++ {
					pLits[0] = segLit(iFirst+k, 1)
					pLits[1] = segLit(Count, 1)
					pLits[2] = segLit(iFirstLut+k, 0)
					addClause(pLits[:3])

					pLits[0] = segLit(iFirst+k, 1)
					pLits[1] = segLit(Count, 0)
					pLits[2] = segLit(iFirstLut+k+1, 0)
					addClause(pLits[:3])
				}
				Count++
			}
		}
	}
	if fVerbose != 0 {
		fmt.Printf("Delay constraints = %d. ", p.pSat.NClauses()-nConstr)
	}
	nConstr = p.pSat.NClauses()
	// edge compatibility constraint
	vObjEdges := Seg_ManCollectObjEdges(p.vEdges, Gia_ManObjNum(p.pGia))
	for i := 0; i < vec.Vec_WecSize(vObjEdges); i++ {
		vLevel := vec.Vec_IntArray(vec.Vec_WecEntry(vObjEdges, i))
		if (fTwo == 0 && len(vLevel) >= 2) || (fTwo != 0 && len(vLevel) > 10) {
			for v1 := 0; v1 < len(vLevel); v1++ {
				for v2 := v1 + 1; v2 < len(vLevel); v2++ {
					pLits[0] = segLit(vLevel[v1], 1)
					pLits[1] = segLit(vLevel[v2], 1)
					addClause(pLits[:2])
				}
			}
		} else if fTwo != 0 && len(vLevel) >= 3 {
			for v1 := 0; v1 < len(vLevel); v1++ {
				for v2 := v1 + 1; v2 < len(vLevel); v2++ {
					for v3 := v2 + 1; v3 < len(vLevel); v3++ {
						pLits[0] = segLit(vLevel[v1], 1)
						pLits[1] = segLit(vLevel[v2], 1)
						pLits[2] = segLit(vLevel[v3], 1)
						addClause(pLits[:3])
					}
				}
			}
		}
	}
	// block forbidden edges
	for _, Edge := range vec.Vec_IntArray(p.vToSkip) {
		pLits[0] = segLit(Edge, 1)
		addClause(pLits[:1])
	}
	if fVerbose != 0 {
		fmt.Printf("Edge constraints = %d. ", p.pSat.NClauses()-nConstr)
	}
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	}
}

func Seg_ManConvertResult(p *Seg_Man_t) *Vec_Int_t {
	vEdges2 := vec.Vec_IntAlloc(1000)
	data := vec.Vec_IntArray(p.vEdges)
	for i := 0; i+1 < len(data); i += 2 {
		if p.pSat.VarValue(int32(i / 2)) {
			vec.Vec_IntPushTwo(vEdges2, data[i], data[i+1])
		}
	}
	return vEdges2
}

// Seg_ManComputeDelay searches for an edge assignment with minimum delay.
func Seg_ManComputeDelay(pGia *Gia_Man_t, DelayInit, nFanouts, fTwo, fVerbose int32) {
	nBTLimit := int64(0)
	clk := util.Abc_Clock()
	var vEdges2 *Vec_Int_t
	p := Seg_ManAlloc(pGia, nFanouts)
	DelayStart := DelayInit
	if DelayStart == 0 {
		DelayStart = p.DelayMax
	}
	if fVerbose != 0 {
		fmt.Printf("Running SatEdge with starting delay %d and edge %d (edge vars %d, total vars %d)\n",
			DelayStart, fTwo+1, p.FirstVar, p.pSat.NumVars())
	}
	Seg_ManCreateCnf(p, fTwo, fVerbose)
	// set resource limits
	p.pSat.SetResourceLimitsC(nBTLimit, 0, 0, 0)
	p.pSat.SetRandomC(1)
	p.pSat.SetPolarityC(vec.Vec_IntArray(p.vPolars))
	// increment delay gradually
	for Delay := p.DelayMax; Delay >= 0; Delay-- {
		// we constrain COs, although it would be fine to constrain only POs
		i := int32(0)
		for ; i < Gia_ManCoNum(p.pGia); i++ {
			iLut := Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))
			iFirst := vec.Vec_IntEntry(p.vFirsts, int(iLut))
			nVars := vec.Vec_IntEntry(p.vNvars, int(iLut))
			if Delay < nVars && p.pSat.PushC(segLit(iFirst+Delay, 1)) == 0 {
				break
			}
		}
		if i < Gia_ManCoNum(p.pGia) {
			fmt.Printf("Proved UNSAT for delay %d.  ", Delay)
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			break
		}
		if Delay > DelayStart {
			continue
		}
		// solve with assumptions
		nConfs := p.pSat.NConflicts()
		status := p.pSat.SolveInternalC()
		nConfs = p.pSat.NConflicts() - nConfs
		if status == 1 { // l_True
			if fVerbose != 0 {
				Count := int32(0)
				for i := int32(0); i < p.nVars; i++ {
					if p.pSat.VarValue(i) {
						Count++
					}
				}
				fmt.Printf("Solution with delay %2d and %5d edges exists. Conf = %8d.  ", Delay, Count, nConfs)
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			// save the result
			vEdges2 = Seg_ManConvertResult(p)
		} else {
			if fVerbose != 0 {
				if status == -1 { // l_False
					fmt.Printf("Proved UNSAT for delay %d.  ", Delay)
				} else {
					fmt.Printf("Resource limit reached for delay %d.  ", Delay)
				}
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			break
		}
	}
	if vEdges2 != nil { // no solution when the delay limit (-D) is below the feasible delay (CBUGS_FIXED.md #31)
		Gia_ManEdgeFromArray(p.pGia, vEdges2)
	}
	Seg_ManStop(p)
}
