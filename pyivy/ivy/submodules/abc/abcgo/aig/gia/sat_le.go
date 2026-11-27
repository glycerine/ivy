package gia

// Line-by-line port of the Sle_Man_t part of src/aig/gia/giaSatLE.c
// ('&edge -m': SAT-based LUT mapping combined with edge assignment).

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Sle_Man_t struct {
	// user's data
	pGia      *Gia_Man_t // user's manager (with mapping and edges)
	nLevels   int32      // total number of levels
	fVerbose  int32      // verbose flag
	nSatCalls int32      // the number of SAT calls
	// SAT variables
	nNodeVars  int32 // node variables (Gia_ManAndNum(pGia))
	nCutVars   int32 // cut variables (total number of non-trivial cuts)
	nEdgeVars  int32 // edge variables (total number of internal edges)
	nDelayVars int32 // delay variables (nNodeVars * nLevelsMax)
	nVarsTotal int32 // total number of variables
	// SAT clauses
	nCutClas   int32 // cut clauses
	nEdgeClas  int32 // edge clauses
	nEdgeClas2 int32 // edge clauses exclusivity
	nDelayClas int32 // delay clauses
	// internal data
	pSat         *bsat.Solver // SAT solver
	vMask        *Vec_Bit_t   // internal node mask
	vCuts        *Vec_Int_t   // cuts for each node
	vCutFanins   *Vec_Wec_t   // internal cut fanins of each node
	vFanoutEdges *Vec_Wec_t   // internal cut fanins of each node
	vEdgeCuts    *Vec_Wec_t   // cuts of each edge for one node
	vObjMap      *Vec_Int_t   // temporary object map
	vCutFirst    *Vec_Int_t   // first cut of each node
	vEdgeFirst   *Vec_Int_t   // first edge of each node
	vDelayFirst  *Vec_Int_t   // first edge of each node
	vPolars      *Vec_Int_t   // initial
	vLits        *Vec_Int_t   // literals
	// statistics
	timeStart int64
}

func Sle_ManList(p *Sle_Man_t, i int32) []int32 {
	data := vec.Vec_IntArray(p.vCuts)
	return data[data[i]:]
}

func Sle_ManAlloc(pGia *Gia_Man_t, nLevels, fVerbose int32) *Sle_Man_t {
	p := &Sle_Man_t{}
	p.pGia = pGia
	p.nLevels = nLevels
	p.fVerbose = fVerbose
	p.vMask = Sle_ManInternalNodeMask(pGia)
	p.vCuts = Sle_ManComputeCuts(pGia, 4, fVerbose)
	p.vCutFanins = Sle_ManCollectCutFanins(pGia, p.vCuts, p.vMask)
	p.vFanoutEdges = vec.Vec_WecStart(int(Gia_ManObjNum(pGia)))
	p.vEdgeCuts = vec.Vec_WecAlloc(100)
	p.vObjMap = vec.Vec_IntStartFull(int(Gia_ManObjNum(pGia)))
	p.vCutFirst = vec.Vec_IntStartFull(int(Gia_ManObjNum(pGia)))
	p.vEdgeFirst = vec.Vec_IntStartFull(int(Gia_ManObjNum(pGia)))
	p.vDelayFirst = vec.Vec_IntStartFull(int(Gia_ManObjNum(pGia)))
	p.vPolars = vec.Vec_IntAlloc(100)
	p.vLits = vec.Vec_IntAlloc(100)
	p.nLevels = Sle_ManComputeDelay(pGia, p.vCuts)
	return p
}
func Sle_ManStop(p *Sle_Man_t) {
	if p.pSat != nil {
		p.pSat.Delete()
	}
	p.pSat = nil
}

func Sle_ManMarkupVariables(p *Sle_Man_t) {
	Counter := Gia_ManObjNum(p.pGia)
	// node variables
	p.nNodeVars = Counter
	// cut variables
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) != 0 {
			vec.Vec_IntWriteEntry(p.vCutFirst, int(iObj), Counter)
			Counter += Sle_ListCutNum(Sle_ManList(p, iObj))
		}
	}
	p.nCutVars = Counter - p.nNodeVars
	// edge variables
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) != 0 {
			vec.Vec_IntWriteEntry(p.vEdgeFirst, int(iObj), Counter)
			Counter += int32(vec.Vec_IntSize(vec.Vec_WecEntry(p.vCutFanins, int(iObj))))
		}
	}
	p.nEdgeVars = Counter - p.nCutVars - p.nNodeVars
	// delay variables
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) != 0 {
			vec.Vec_IntWriteEntry(p.vDelayFirst, int(iObj), Counter)
			Counter += p.nLevels
		}
	}
	p.nDelayVars = Counter - p.nEdgeVars - p.nCutVars - p.nNodeVars
	p.nVarsTotal = Counter
}

// Sle_ManCheckContained returns 1 if Cut can represent LUT (Cut is equal or
// is contained in LUT).
func Sle_ManCheckContained(pCutLeaves []int32, nCutLeaves int32, pLutFanins []int32, nLutFanins int32) int32 {
	if nCutLeaves > nLutFanins {
		return 0
	}
	for i := int32(0); i < nCutLeaves; i++ {
		k := int32(0)
		for ; k < nLutFanins; k++ {
			if pCutLeaves[i] == pLutFanins[k] {
				break
			}
		}
		if k == nLutFanins { // not found
			return 0
		}
	}
	return 1
}

// Sle_ManDeriveInit derives initial variable assignment.
func Sle_ManDeriveInit(p *Sle_Man_t) {
	if Gia_ManHasMapping(p.pGia) == 0 {
		return
	}
	// derive initial state
	vec.Vec_IntClear(p.vPolars)
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) == 0 || Gia_ObjIsLut(p.pGia, iObj) == 0 {
			continue
		}
		vec.Vec_IntPush(p.vPolars, iObj) // node var
		pFanins := Gia_ObjLutFanins(p.pGia, iObj)
		nFanins := int32(len(pFanins))
		// find cut
		// When no cut is found (a LUT of more than four inputs), C's loop
		// leaves pCut at the unit cut that follows the node's cuts.
		iFound := -1
		cuts := sleList(p.vCuts, iObj, true)
		pCut := cuts[len(cuts)-1]
		for i, cut := range cuts[:len(cuts)-1] {
			if Sle_ManCheckContained(Sle_CutLeaves(cut), Sle_CutSize(cut), pFanins, nFanins) != 0 {
				iFound = i
				pCut = cut
				break
			}
		}
		if iFound == -1 {
			fmt.Printf("Cannot find the following cut at node %d: {", iObj)
			for _, f := range pFanins {
				fmt.Printf(" %d", f)
			}
			fmt.Printf(" }\n")
			Sle_ManPrintCuts(p.pGia, p.vCuts, iObj)
			// assert( iFound >= 0 ) is compiled out in release builds
		}
		vec.Vec_IntPush(p.vPolars, vec.Vec_IntEntry(p.vCutFirst, int(iObj))+int32(iFound)) // cut var
		// check if the cut contains only primary inputs - if so, its delay is equal to 1
		if Sle_ManCutHasPisOnly(pCut, p.vMask) != 0 {
			vec.Vec_IntPush(p.vPolars, vec.Vec_IntEntry(p.vDelayFirst, int(iObj))) // delay var
		}
	}
	vec.Vec_IntSort(p.vPolars, 0)
	// find zero-delay edges
	if p.pGia.VEdge1 == nil {
		return
	}
	vEdges := vec.Vec_IntArray(Gia_ManEdgeToArray(p.pGia))
	for i := 0; i+1 < len(vEdges); i += 2 {
		iFanin, iObj := vEdges[i], vEdges[i+1]
		// find edge
		iEdge := vec.Vec_IntFind(vec.Vec_WecEntry(p.vCutFanins, int(iObj)), iFanin)
		if iEdge < 0 {
			continue
		}
		vec.Vec_IntPush(p.vPolars, vec.Vec_IntEntry(p.vEdgeFirst, int(iObj))+int32(iEdge)) // edge
	}
}

func sleLit(v, c int32) int32 { return v + v + c }

func (p *Sle_Man_t) addLits() {
	// C asserts that every clause is accepted; release builds go on.
	p.pSat.AddClauseC(vec.Vec_IntArray(p.vLits))
}

// Sle_ManDeriveCnf derives CNF.
func Sle_ManDeriveCnf(p *Sle_Man_t, nBTLimit, fDynamic int32) {
	// start the SAT solver
	p.pSat = bsat.NewSolver()
	p.pSat.SetNumVars(p.nVarsTotal)
	p.pSat.SetResourceLimitsC(int64(nBTLimit), 0, 0, 0)
	p.pSat.SetRandomC(1)
	p.pSat.SetPolarityC(vec.Vec_IntArray(p.vPolars))

	// set drivers to be mapped
	for i := int32(0); i < Gia_ManCoNum(p.pGia); i++ {
		iObj := Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))
		if vec.Vec_BitEntry(p.vMask, int(iObj)) != 0 { // internal node
			vec.Vec_IntFill(p.vLits, 1, sleLit(iObj, 0)) // pos lit
			p.addLits()
		}
	}

	// add cover clauses and edge-to-cut clauses
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) == 0 {
			continue
		}
		nEdges := int32(0)
		iCutVar0 := vec.Vec_IntEntry(p.vCutFirst, int(iObj))
		iEdgeVar0 := vec.Vec_IntEntry(p.vEdgeFirst, int(iObj))
		pList := Sle_ManList(p, iObj)
		vCutFans := vec.Vec_WecEntry(p.vCutFanins, int(iObj))
		// node requires one of the cuts
		vec.Vec_IntFill(p.vLits, 1, sleLit(iObj, 1)) // neg lit
		for i := int32(0); i < Sle_ListCutNum(pList); i++ {
			vec.Vec_IntPush(p.vLits, sleLit(iCutVar0+i, 0))
		}
		p.addLits()
		// cuts are mutually exclusive
		for i := int32(0); i < Sle_ListCutNum(pList); i++ {
			for e := i + 1; e < Sle_ListCutNum(pList); e++ {
				vec.Vec_IntFillTwo(p.vLits, 2, sleLit(iCutVar0+i, 1), sleLit(iCutVar0+e, 1))
				p.addLits()
			}
		}
		// cut requires fanin nodes
		vec.Vec_WecInit(p.vEdgeCuts, vec.Vec_IntSize(vCutFans))
		for i, pCut := range sleList(p.vCuts, iObj, false) {
			for _, leaf := range Sle_CutLeaves(pCut) {
				if vec.Vec_BitEntry(p.vMask, int(leaf)) == 0 {
					continue
				}
				vec.Vec_IntFillTwo(p.vLits, 2, sleLit(iCutVar0+int32(i), 1), sleLit(leaf, 0))
				p.addLits()
				// find the edge ID between pC[k] and iObj
				iEdge := vec.Vec_IntEntry(p.vObjMap, int(leaf))
				if iEdge == -1 {
					iEdge = nEdges
					nEdges++
					vec.Vec_IntWriteEntry(p.vObjMap, int(leaf), iEdge)
					vec.Vec_WecPush(p.vFanoutEdges, int(leaf), iEdgeVar0+iEdge)
				}
				vec.Vec_WecPush(p.vEdgeCuts, int(iEdge), iCutVar0+int32(i))
				p.nCutClas++
			}
			// cut requires the node
			vec.Vec_IntFillTwo(p.vLits, 2, sleLit(iCutVar0+int32(i), 1), sleLit(iObj, 0))
			p.addLits()
		}

		// edge requires one of the fanout cuts
		for e := 0; e < vec.Vec_WecSize(p.vEdgeCuts); e++ {
			vec.Vec_IntFill(p.vLits, 1, sleLit(iEdgeVar0+int32(e), 1)) // neg lit (edge)
			for _, Entry := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vEdgeCuts, e)) {
				vec.Vec_IntPush(p.vLits, sleLit(Entry, 0)) // pos lit (cut)
			}
			p.addLits()
			p.nEdgeClas++
		}

		// clean object map
		for _, Entry := range vec.Vec_IntArray(vCutFans) {
			vec.Vec_IntWriteEntry(p.vObjMap, int(Entry), -1)
		}
	}

	// mutual exclusivity of edges
	for iObj := 0; iObj < vec.Vec_WecSize(p.vFanoutEdges); iObj++ {
		vArray := vec.Vec_WecEntry(p.vFanoutEdges, iObj)
		iEdgeVar0 := vec.Vec_IntEntry(p.vEdgeFirst, iObj)
		vCutFans := vec.Vec_WecEntry(p.vCutFanins, iObj)
		// add fanin edges
		for i := 0; i < vec.Vec_IntSize(vCutFans); i++ {
			vec.Vec_IntPush(vArray, iEdgeVar0+int32(i))
		}
		// generate pairs
		if fDynamic != 0 {
			continue
		}
		edges := vec.Vec_IntArray(vArray)
		for j := range edges {
			for k := j + 1; k < len(edges); k++ {
				vec.Vec_IntFillTwo(p.vLits, 2, sleLit(edges[j], 1), sleLit(edges[k], 1))
				p.addLits()
			}
		}
		p.nEdgeClas2 += int32(len(edges) * (len(edges) - 1) / 2)
	}

	// add delay clauses
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) == 0 {
			continue
		}
		iEdgeVar0 := vec.Vec_IntEntry(p.vEdgeFirst, int(iObj))
		iDelayVar0 := vec.Vec_IntEntry(p.vDelayFirst, int(iObj))
		vCutFans := vec.Vec_WecEntry(p.vCutFanins, int(iObj))
		// check if the node has cuts containing only primary inputs
		for _, pCut := range sleList(p.vCuts, iObj, false) {
			if Sle_ManCutHasPisOnly(pCut, p.vMask) != 0 {
				vec.Vec_IntFill(p.vLits, 1, sleLit(iDelayVar0, 0)) // pos lit
				p.addLits()
				break
			}
		}
		// create delay requirements for each cut fanin of this node
		for e, iFanin := range vec.Vec_IntArray(vCutFans) {
			iDelayVarIn := vec.Vec_IntEntry(p.vDelayFirst, int(iFanin))
			for d := int32(0); d < p.nLevels; d++ {
				vec.Vec_IntClear(p.vLits)
				vec.Vec_IntPush(p.vLits, sleLit(iObj, 1))
				vec.Vec_IntPush(p.vLits, sleLit(iFanin, 1))
				vec.Vec_IntPush(p.vLits, sleLit(iDelayVarIn+d, 1))
				vec.Vec_IntPush(p.vLits, sleLit(iEdgeVar0+int32(e), 0))
				if d < p.nLevels-1 {
					vec.Vec_IntPush(p.vLits, sleLit(iDelayVar0+d+1, 0))
				}
				p.addLits()

				vec.Vec_IntClear(p.vLits)
				vec.Vec_IntPush(p.vLits, sleLit(iObj, 1))
				vec.Vec_IntPush(p.vLits, sleLit(iFanin, 1))
				vec.Vec_IntPush(p.vLits, sleLit(iDelayVarIn+d, 1))
				if d < p.nLevels-1 {
					vec.Vec_IntPush(p.vLits, sleLit(iEdgeVar0+int32(e), 1))
				}
				vec.Vec_IntPush(p.vLits, sleLit(iDelayVar0+d, 0))
				p.addLits()
			}
			p.nDelayClas += 2 * p.nLevels
		}
	}
}

// Sle_ManAddEdgeConstraints adds edge compatibility constraints; returns
// the number of nodes that received them.
func Sle_ManAddEdgeConstraints(p *Sle_Man_t, nEdges int32) int32 {
	vTemp := vec.Vec_IntAlloc(100)
	nAdded := int32(0)
	for iObj := 0; iObj < vec.Vec_WecSize(p.vFanoutEdges); iObj++ {
		// check if they are incompatible
		vec.Vec_IntClear(vTemp)
		for _, EdgeJ := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vFanoutEdges, iObj)) {
			if p.pSat.VarValue(EdgeJ) {
				vec.Vec_IntPush(vTemp, EdgeJ)
			}
		}
		if int32(vec.Vec_IntSize(vTemp)) <= nEdges {
			continue
		}
		nAdded++
		t := vec.Vec_IntArray(vTemp)
		n := int32(len(t))
		if nEdges == 1 {
			// generate pairs
			for j := range t {
				for k := j + 1; k < len(t); k++ {
					vec.Vec_IntFillTwo(p.vLits, 2, sleLit(t[j], 1), sleLit(t[k], 1))
					p.addLits()
				}
			}
			p.nEdgeClas2 += n * (n - 1) / 2
		} else if nEdges == 2 {
			// generate triples
			for j := range t {
				for k := j + 1; k < len(t); k++ {
					for m := k + 1; m < len(t); m++ {
						vec.Vec_IntFillTwo(p.vLits, 2, sleLit(t[j], 1), sleLit(t[k], 1))
						vec.Vec_IntPush(p.vLits, sleLit(t[m], 1))
						p.addLits()
					}
				}
			}
			p.nEdgeClas2 += n * (n - 1) * (n - 2) / 6
		}
	}
	return nAdded
}

func Sle_ManDeriveResult(p *Sle_Man_t, vEdge2, vMapping *Vec_Int_t) {
	// create mapping
	vec.Vec_IntFill(vMapping, int(Gia_ManObjNum(p.pGia)), 0)
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) == 0 {
			continue
		}
		iCutVar0 := vec.Vec_IntEntry(p.vCutFirst, int(iObj))
		var pCutThis []int32
		if !p.pSat.VarValue(iObj) {
			continue
		}
		for iCut, pCut := range sleList(p.vCuts, iObj, false) {
			if p.pSat.VarValue(iCutVar0 + int32(iCut)) {
				pCutThis = pCut
			}
		}
		vec.Vec_IntWriteEntry(vMapping, int(iObj), int32(vec.Vec_IntSize(vMapping)))
		vec.Vec_IntPush(vMapping, Sle_CutSize(pCutThis))
		for _, leaf := range Sle_CutLeaves(pCutThis) {
			vec.Vec_IntPush(vMapping, leaf)
		}
		vec.Vec_IntPush(vMapping, iObj)
	}
	vMapTemp := p.pGia.VMapping
	p.pGia.VMapping = vMapping
	// collect edges
	vec.Vec_IntClear(vEdge2)
	for iObj := int32(1); iObj < p.pGia.nObjs; iObj++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)) == 0 {
			continue
		}
		iEdgeVar0 := vec.Vec_IntEntry(p.vEdgeFirst, int(iObj))
		if !p.pSat.VarValue(iObj) {
			continue
		}
		for i, iFanin := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vCutFanins, int(iObj))) {
			if p.pSat.VarValue(iFanin) && p.pSat.VarValue(iEdgeVar0+int32(i)) {
				// verify edge
				found := false
				for _, f := range Gia_ObjLutFanins(p.pGia, iObj) {
					if f == iFanin {
						found = true
						break
					}
				}
				if !found {
					continue
				}
				vec.Vec_IntPushTwo(vEdge2, iFanin, iObj)
			}
		}
	}
	p.pGia.VMapping = vMapTemp
}

// Sle_ManExplore searches for a low-delay LUT mapping and compatible edges.
func Sle_ManExplore(pGia *Gia_Man_t, nBTLimit, DelayInit, fDynamic, fTwoEdges, fVerbose int32) {
	clk := util.Abc_Clock()
	vEdges2 := vec.Vec_IntAlloc(1000)
	vMapping := vec.Vec_IntAlloc(1000)
	var status, nConfs int32
	DelayStart := DelayInit
	if !(DelayInit != 0 || Gia_ManHasMapping(pGia) == 0) {
		DelayStart = Gia_ManLutLevel(pGia, nil)
	}
	p := Sle_ManAlloc(pGia, DelayStart, fVerbose)
	if fVerbose != 0 {
		fDyn := "no"
		if fDynamic != 0 {
			fDyn = "yes"
		}
		fmt.Printf("Running solver with %d conflicts, %d initial delay, and %d edges. Dynamic constraints = %s.\n", nBTLimit, DelayInit, 1+fTwoEdges, fDyn)
	}
	Sle_ManMarkupVariables(p)
	if fVerbose != 0 {
		fmt.Printf("Vars:  Total = %d.  Node = %d. Cut = %d. Edge = %d. Delay = %d.\n",
			p.nVarsTotal, p.nNodeVars, p.nCutVars, p.nEdgeVars, p.nDelayVars)
	}
	Sle_ManDeriveInit(p)
	Sle_ManDeriveCnf(p, nBTLimit, giaBool(fDynamic != 0 || fTwoEdges != 0))
	if fVerbose != 0 {
		fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.\n",
			p.pSat.NClauses(), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas)
	}
	for Delay := p.nLevels; Delay >= 0; Delay-- {
		// we constrain COs, although it would be fine to constrain only POs
		if Delay < p.nLevels {
			i := int32(0)
			for ; i < Gia_ManCoNum(p.pGia); i++ {
				iLut := Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))
				if vec.Vec_BitEntry(p.vMask, int(iLut)) != 0 { // internal node
					iFirstVar := vec.Vec_IntEntry(p.vDelayFirst, int(iLut))
					if p.pSat.PushC(sleLit(iFirstVar+Delay, 1)) == 0 {
						break
					}
				}
			}
			if i < Gia_ManCoNum(p.pGia) {
				if fVerbose != 0 {
					fmt.Printf("Proved UNSAT for delay %d.  ", Delay)
					util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
				}
				break
			}
		}
		// solve with assumptions
		nConfs = p.pSat.NConflicts()
		for {
			p.nSatCalls++
			status = p.pSat.SolveInternalC()
			if status != 1 { // l_True
				break
			}
			if Sle_ManAddEdgeConstraints(p, 1+fTwoEdges) == 0 {
				break
			}
		}
		nConfs = p.pSat.NConflicts() - nConfs
		if status == 1 {
			if fVerbose != 0 {
				nNodes, nEdges := 0, 0
				for i := int32(0); i < p.nNodeVars; i++ {
					if p.pSat.VarValue(i) {
						nNodes++
					}
				}
				for i := int32(0); i < p.nEdgeVars; i++ {
					if p.pSat.VarValue(p.nNodeVars + p.nCutVars + i) {
						nEdges++
					}
				}
				fmt.Printf("Solution with delay %2d, node count %5d, and edge count %5d exists. Conf = %8d.  ", Delay, nNodes, nEdges, nConfs)
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			// save the result
			Sle_ManDeriveResult(p, vEdges2, vMapping)
		} else {
			if fVerbose != 0 {
				if status == -1 { // l_False
					fmt.Printf("Proved UNSAT for delay %d. Conf = %8d.  ", Delay, nConfs)
				} else {
					fmt.Printf("Resource limit reached for delay %d. Conf = %8d.  ", Delay, nConfs)
				}
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			break
		}
	}
	if fVerbose != 0 {
		fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.  Calls = %d.\n",
			p.pSat.NClauses(), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas, p.nSatCalls)
	}
	if vec.Vec_IntSize(vMapping) > 0 {
		Gia_ManEdgeFromArray(p.pGia, vEdges2)
		p.pGia.VMapping = vMapping
	}
	p.pGia.VPacking = nil
	Sle_ManStop(p)
}
