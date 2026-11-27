package abci

// Line-by-line port of ABC's functional support computation
// (src/opt/sim/simSupp.c: Sim_ComputeFunSupp and its helpers), the
// Sim_Man_t part of simMan.c and the simUtils.c routines it uses. Like C it
// works on the strashed Abc_Ntk_t: SAT checks prove
// Abc_NtkMiterForCofactors miters with the old FRAIG package, and random
// patterns come from the process-wide rand() stream (util.C_Srand,
// util.Sim_RandomUnsigned), which Sim_ComputeTwoVarSymms continues.

import (
	"fmt"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/fraig"
)

// Sim_Man_t is sim.h's support simulation manager.
type Sim_Man_t struct {
	// info about the network
	pNtk         *Abc_Ntk_t
	fLightweight int32
	nInputs      int32
	nOutputs     int32
	iInput       int32
	// simulation information
	nSimBits  int32
	nSimWords int32
	vSim0     [][]uint32
	vSim1     [][]uint32
	// support information
	nSuppBits  int32
	nSuppWords int32
	vSuppStr   [][]uint32
	vSuppFun   [][]uint32
	// patterns to be analyzed
	vSuppTargs [][]int32 // support targets: unresolved outputs for each input
	vFifo      []*Sim_Pat_t
	vDiffs     []int32
	// runtime statistics
	nSatRuns      int32
	nSatRunsSat   int32
	nSatRunsUnsat int32
	timeSim       int64
	timeTrav      int64
	timeFraig     int64
	timeSat       int64
	timeTotal     int64
}

// Sim_Pat_t is one counter-example pattern.
type Sim_Pat_t struct {
	Input  int32
	Output int32
	pData  []uint32
}

/*
*Function*************************************************************

	Synopsis    [Starts the simulation manager.]

**********************************************************************
*/
func Sim_ManStart(pNtk *Abc_Ntk_t, fLightweight int32) *Sim_Man_t {
	p := &Sim_Man_t{}
	p.pNtk = pNtk
	p.nInputs = baseabc.Abc_NtkCiNum(p.pNtk)
	p.nOutputs = baseabc.Abc_NtkCoNum(p.pNtk)
	// internal simulation information
	p.nSimBits = 2048
	p.nSimWords = symSimNumWords(p.nSimBits)
	p.vSim0 = symInfoAlloc(baseabc.Abc_NtkObjNumMax(pNtk), p.nSimWords)
	p.fLightweight = fLightweight
	if p.fLightweight == 0 {
		p.vSim1 = symInfoAlloc(baseabc.Abc_NtkObjNumMax(pNtk), p.nSimWords)
		// support information
		p.nSuppBits = baseabc.Abc_NtkCiNum(pNtk)
		p.nSuppWords = symSimNumWords(p.nSuppBits)
		p.vSuppStr = Sim_ComputeStrSupp(pNtk)
		p.vSuppFun = symInfoAlloc(baseabc.Abc_NtkCoNum(p.pNtk), p.nSuppWords)
		// other data
		p.vFifo = make([]*Sim_Pat_t, 0, 100)
		p.vDiffs = make([]int32, 0, 100)
		// allocate support targets (array of unresolved outputs for each input)
		p.vSuppTargs = make([][]int32, p.nInputs)
	}
	return p
}

/*
*Function*************************************************************

	Synopsis    [Stops the simulation manager.]

**********************************************************************
*/
func Sim_ManStop(p *Sim_Man_t) {
	Sim_ManPrintStats(p)
	*p = Sim_Man_t{}
}

/*
*Function*************************************************************

	Synopsis    [Prints the manager statisticis.]

**********************************************************************
*/
func Sim_ManPrintStats(p *Sim_Man_t) {
	fmt.Printf("Total func supps   = %8d.\n", Sim_UtilCountSuppSizes(p, 0))
	fmt.Printf("Total struct supps = %8d.\n", Sim_UtilCountSuppSizes(p, 1))
	fmt.Printf("Sat runs SAT       = %8d.\n", p.nSatRunsSat)
	fmt.Printf("Sat runs UNSAT     = %8d.\n", p.nSatRunsUnsat)
	util.Abc_PrintTime(1, "Simulation  ", p.timeSim)
	util.Abc_PrintTime(1, "Traversal   ", p.timeTrav)
	util.Abc_PrintTime(1, "Fraiging    ", p.timeFraig)
	util.Abc_PrintTime(1, "SAT         ", p.timeSat)
	util.Abc_PrintTime(1, "TOTAL       ", p.timeTotal)
}

/*
*Function*************************************************************

	Synopsis    [Returns one simulation pattern.]

**********************************************************************
*/
func Sim_ManPatAlloc(p *Sim_Man_t) *Sim_Pat_t {
	return &Sim_Pat_t{Output: -1, pData: make([]uint32, p.nSuppWords)}
}

/*
*Function*************************************************************

	Synopsis    [Computes functional supports of the network.]

	Description [Supports are returned as arrays of bit strings, one for each CO.]

**********************************************************************
*/
func Sim_ComputeFunSupp(pNtk *Abc_Ntk_t, fVerbose int32) *Vec_Ptr_t {
	var nSolved, i int32
	clk := util.Abc_Clock()

	util.C_Srand(0xABC)

	// start the simulation manager
	p := Sim_ManStart(pNtk, 0)

	// compute functional support using one round of random simulation
	Sim_UtilAssignRandom(p)
	Sim_ComputeSuppRound(p, 0)

	// set the support targets
	Sim_ComputeSuppSetTargets(p)
	if fVerbose != 0 {
		fmt.Printf("Number of support targets after simulation = %5d.\n", simVecSizeSize(p.vSuppTargs))
	}
	if simVecSizeSize(p.vSuppTargs) == 0 {
		goto exit
	}

	for i = 0; i < 1; i++ {
		// compute patterns using one round of random simulation
		Sim_UtilAssignRandom(p)
		nSolved = Sim_ComputeSuppRound(p, 1)
		if simVecSizeSize(p.vSuppTargs) == 0 {
			goto exit
		}
		if fVerbose != 0 {
			fmt.Printf("Targets = %5d.   Solved = %5d.  Fifo = %5d.\n",
				simVecSizeSize(p.vSuppTargs), nSolved, len(p.vFifo))
		}
	}

	// try to solve the support targets
	for simVecSizeSize(p.vSuppTargs) > 0 {
		// solve targets until the first disproved one (which gives counter-example)
		Sim_SolveTargetsUsingSat(p, p.nSimWords/p.nSuppWords)
		// compute additional functional support
		Sim_UtilAssignFromFifo(p)
		nSolved = Sim_ComputeSuppRound(p, 1)
		if fVerbose != 0 {
			fmt.Printf("Targets = %5d.   Solved = %5d.  Fifo = %5d.  SAT runs = %3d.\n",
				simVecSizeSize(p.vSuppTargs), nSolved, len(p.vFifo), p.nSatRuns)
		}
	}

exit:
	p.timeTotal = util.Abc_Clock() - clk
	vSuppFun := p.vSuppFun
	Sim_ManStop(p)
	vResult := vec.Vec_PtrAlloc(len(vSuppFun))
	for _, pSupp := range vSuppFun {
		vec.Vec_PtrPush(vResult, pSupp)
	}
	return vResult
}

// simVecSizeSize is Vec_VecSizeSize.
func simVecSizeSize(p [][]int32) int32 {
	var Counter int32
	for _, vVec := range p {
		Counter += int32(len(vVec))
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Computes functional support using one round of simulation.]

**********************************************************************
*/
func Sim_ComputeSuppRound(p *Sim_Man_t, fUseTargets int32) int32 {
	var i, Counter int32
	// perform one round of random simulation
	clk := util.Abc_Clock()
	Sim_UtilSimulate(p, 0)
	p.timeSim += util.Abc_Clock() - clk
	// iterate through the CIs and detect COs that depend on them
	for i = p.iInput; i < p.nInputs; i++ {
		vTargets := p.vSuppTargs[i]
		if fUseTargets != 0 && len(vTargets) == 0 {
			continue
		}
		Counter += Sim_ComputeSuppRoundNode(p, i, fUseTargets)
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Computes functional support for one node.]

**********************************************************************
*/
func Sim_ComputeSuppRoundNode(p *Sim_Man_t, iNumCi int32, fUseTargets int32) int32 {
	var Counter int32
	fFirst := true
	var clk int64

	// collect nodes by level in the TFO of the CI
	// this procedure does not collect the CIs and COs
	// but it increments TravId of the collected nodes and CIs/COs
	clk = util.Abc_Clock()
	pNodeCi := baseabc.Abc_NtkCi(p.pNtk, iNumCi)
	vNodesByLevel := baseabc.Abc_DfsLevelized(pNodeCi, 0)
	p.timeTrav += util.Abc_Clock() - clk
	// complement the simulation info of the selected CI
	Sim_UtilInfoFlip(p, pNodeCi)
	// simulate the levelized structure of nodes
	for _, vLevel := range vNodesByLevel {
		for _, pNode := range vLevel {
			fType0 := baseabc.Abc_NodeIsTravIdCurrent(baseabc.Abc_ObjFanin0(pNode))
			fType1 := baseabc.Abc_NodeIsTravIdCurrent(baseabc.Abc_ObjFanin1(pNode))
			clk = util.Abc_Clock()
			Sim_UtilSimulateNode(p, pNode, 1, fType0, fType1)
			p.timeSim += util.Abc_Clock() - clk
		}
	}
	// set the simulation info of the affected COs
	if fUseTargets != 0 {
		for i := len(p.vSuppTargs[iNumCi]) - 1; i >= 0; i-- {
			vTargets := p.vSuppTargs[iNumCi]
			// get the target output
			Output := vTargets[i]
			// get the target node
			pNode := baseabc.Abc_ObjFanin0(baseabc.Abc_NtkCo(p.pNtk, Output))
			// the output should be in the cone
			assert(baseabc.Abc_NodeIsTravIdCurrent(pNode) != 0, "Abc_NodeIsTravIdCurrent(pNode)")

			// skip if the simulation info is equal
			if Sim_UtilInfoCompare(p, pNode) != 0 {
				continue
			}

			// otherwise, we solved a new target
			p.vSuppTargs[iNumCi] = simIntRemove(vTargets, Output)
			Counter++
			// make sure this variable is not yet detected
			assert(!symHasBit(p.vSuppFun[Output], iNumCi), "!Sim_SuppFunHasVar(p->vSuppFun, Output, iNumCi)")
			// set this variable
			symSetBit(p.vSuppFun[Output], iNumCi)

			// detect the differences in the simulation info
			p.vDiffs = Sim_UtilInfoDetectDiffs(p.vSim0[pNode.ID], p.vSim1[pNode.ID], p.nSimWords, p.vDiffs)
			// create new patterns
			if !fFirst && len(p.vFifo) > 1000 {
				continue
			}

			for _, LuckyPat := range p.vDiffs {
				// set the new pattern
				pPat := Sim_ManPatAlloc(p)
				pPat.Input = iNumCi
				pPat.Output = Output
				for v := int32(0); v < p.nInputs; v++ {
					if symHasBit(p.vSim0[baseabc.Abc_NtkCi(p.pNtk, v).ID], LuckyPat) {
						symSetBit(pPat.pData, v)
					}
				}
				p.vFifo = append(p.vFifo, pPat)
				fFirst = false
				break
			}
		}
	} else {
		for Output := int32(0); Output < p.nOutputs; Output++ {
			pNode := baseabc.Abc_NtkCo(p.pNtk, Output)
			if baseabc.Abc_NodeIsTravIdCurrent(pNode) == 0 {
				continue
			}
			if Sim_UtilInfoCompare(p, baseabc.Abc_ObjFanin0(pNode)) == 0 {
				if !symHasBit(p.vSuppFun[Output], iNumCi) {
					Counter++
					symSetBit(p.vSuppFun[Output], iNumCi)
				}
			}
		}
	}
	return Counter
}

// simIntRemove is Vec_PtrRemove: deletes the last occurrence of Entry.
func simIntRemove(p []int32, Entry int32) []int32 {
	var i int
	for i = len(p) - 1; i >= 0; i-- {
		if p[i] == Entry {
			break
		}
	}
	assert(i >= 0, "i >= 0")
	copy(p[i:], p[i+1:])
	return p[:len(p)-1]
}

/*
*Function*************************************************************

	Synopsis    [Sets the simulation targets.]

**********************************************************************
*/
func Sim_ComputeSuppSetTargets(p *Sim_Man_t) {
	for i := int32(0); i < p.nOutputs; i++ {
		pNode := baseabc.Abc_NtkCo(p.pNtk, i)
		pSuppStr := p.vSuppStr[pNode.ID]
		pSuppFun := p.vSuppFun[i]
		// find vars in the structural support that are not in the functional support
		p.vDiffs = Sim_UtilInfoDetectNews(pSuppFun, pSuppStr, p.nSuppWords, p.vDiffs)
		for _, Num := range p.vDiffs {
			p.vSuppTargs[Num] = append(p.vSuppTargs[Num], i)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Assigns random simulation info to the PIs.]

**********************************************************************
*/
func Sim_UtilAssignRandom(p *Sim_Man_t) {
	// assign the random/systematic simulation info to the PIs
	for i := int32(0); i < p.nInputs; i++ {
		pSimInfo := p.vSim0[baseabc.Abc_NtkCi(p.pNtk, i).ID]
		for k := int32(0); k < p.nSimWords; k++ {
			pSimInfo[k] = util.Sim_RandomUnsigned() // SIM_RANDOM_UNSIGNED
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Sets the new patterns from fifo.]

**********************************************************************
*/
func Sim_UtilAssignFromFifo(p *Sim_Man_t) {
	var nWordsNew, iWord, iWordLim, i, w int32
	// go through the patterns and fill in the dist-1 minterms for each
	for iWord = 0; len(p.vFifo) > 0; iWord = iWordLim {
		// get the pattern
		pPat := p.vFifo[len(p.vFifo)-1]
		p.vFifo = p.vFifo[:len(p.vFifo)-1]
		// get the number of words for the remaining inputs
		nWordsNew = p.nSuppWords
		// get the first word of the next series
		if iWord+nWordsNew < p.nSimWords {
			iWordLim = iWord + nWordsNew
		} else {
			iWordLim = p.nSimWords
		}
		// set the pattern for all CIs from iWord to iWord + nWordsNew
		for i = 0; i < p.nInputs; i++ {
			pSimInfo := p.vSim0[baseabc.Abc_NtkCi(p.pNtk, i).ID]
			if symHasBit(pPat.pData, i) {
				for w = iWord; w < iWordLim; w++ {
					pSimInfo[w] = ^uint32(0) // SIM_MASK_FULL
				}
			} else {
				for w = iWord; w < iWordLim; w++ {
					pSimInfo[w] = 0
				}
			}
			// the last series may be shorter than nSuppWords
			if i < 32*(iWordLim-iWord) {
				symXorBit(pSimInfo[iWord:], i)
			}
		}
		// stop if we ran out of room for patterns
		if iWordLim == p.nSimWords {
			break
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Get the given number of counter-examples using SAT.]

**********************************************************************
*/
func Sim_SolveTargetsUsingSat(p *Sim_Man_t, Limit int32) {
	var Params Fraig_Params_t
	var Counter int32
	var clk int64

	p.nSatRuns = 0
	// put targets into one array
	// Vec_VecForEachEntryReverse: levels forward, entries backward
	for Input := int32(0); Input < int32(len(p.vSuppTargs)); Input++ {
		for k := len(p.vSuppTargs[Input]) - 1; k >= 0; k-- {
			pEntry := p.vSuppTargs[Input][k]
			p.nSatRuns++
			Output := pEntry

			// set up the miter for the two cofactors of this output w.r.t. this input
			pMiter := Abc_NtkMiterForCofactors(p.pNtk, Output, Input, -1)

			// transform the miter into a fraig
			fraig.Fraig_ParamsSetDefault(&Params)
			Params.NSeconds = util.ABC_INFINITY
			Params.FInternal = 1
			clk = util.Abc_Clock()
			pMan := Abc_NtkToFraig(pMiter, &Params, 0, 0).(*Fraig_Man_t)
			p.timeFraig += util.Abc_Clock() - clk
			clk = util.Abc_Clock()
			fraig.Fraig_ManProveMiter(pMan)
			p.timeSat += util.Abc_Clock() - clk

			// analyze the result
			RetValue := fraig.Fraig_ManCheckMiter(pMan)
			assert(RetValue >= 0, "RetValue >= 0")
			if RetValue == 1 { // unsat
				p.nSatRunsUnsat++
				p.vSuppTargs[Input] = simIntRemove(p.vSuppTargs[Input], pEntry)
			} else { // sat
				p.nSatRunsSat++
				pModel := fraig.Fraig_ManReadModel(pMan)
				assert(pModel != nil, "pModel != NULL")
				assert(Sim_SolveSuppModelVerify(p.pNtk, pModel, Input, Output) != 0, "Sim_SolveSuppModelVerify( p->pNtk, pModel, Input, Output )")

				// set the new pattern
				pPat := Sim_ManPatAlloc(p)
				pPat.Input = Input
				pPat.Output = Output
				for v := int32(0); v < p.nInputs; v++ {
					if pModel[v] != 0 {
						symSetBit(pPat.pData, v)
					}
				}
				p.vFifo = append(p.vFifo, pPat)
				Counter++
			}
			// delete the fraig manager
			fraig.Fraig_ManFree(pMan)
			// delete the miter
			baseabc.Abc_NtkDelete(pMiter)

			// makr the input, which we are processing
			p.iInput = Input

			// stop when we found enough patterns
			if Counter == 1 {
				return
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Verifies that pModel distinguishes the two cofactors.]

	Description [Sim_NtkSimTwoPats_rec keeps two-bit values in pCopy; here
	they live in a local slice indexed by object ID.]

**********************************************************************
*/
func Sim_SolveSuppModelVerify(pNtk *Abc_Ntk_t, pModel []int32, Input, Output int32) int32 {
	values := make([]int32, baseabc.Abc_NtkObjNumMax(pNtk))
	visited := make([]bool, len(values))
	// set the PI values
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		visited[pNode.ID] = true
		if pNode == baseabc.Abc_NtkCi(pNtk, Input) {
			values[pNode.ID] = 1
		} else if pModel[i] == 1 {
			values[pNode.ID] = 3
		} else {
			values[pNode.ID] = 0
		}
	}
	// perform the traversal
	var rec func(pNode *Abc_Obj_t) int32
	rec = func(pNode *Abc_Obj_t) int32 {
		if visited[pNode.ID] {
			return values[pNode.ID]
		}
		visited[pNode.ID] = true
		Value0 := rec(baseabc.Abc_ObjFanin0(pNode))
		Value1 := rec(baseabc.Abc_ObjFanin1(pNode))
		if baseabc.Abc_ObjFaninC0(pNode) != 0 {
			Value0 = ^Value0
		}
		if baseabc.Abc_ObjFaninC1(pNode) != 0 {
			Value1 = ^Value1
		}
		values[pNode.ID] = Value0 & Value1
		return Value0 & Value1
	}
	RetValue := 3 & rec(baseabc.Abc_ObjFanin0(baseabc.Abc_NtkCo(pNtk, Output)))
	if RetValue == 1 || RetValue == 2 {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------- simUtils.c

/*
*Function*************************************************************

	Synopsis    [Simulates the internal nodes.]

**********************************************************************
*/
func Sim_UtilSimulate(p *Sim_Man_t, fType int32) {
	// simulate the internal nodes
	for _, pNode := range p.pNtk.Objects() {
		if baseabc.Abc_ObjIsNode(pNode) != 0 {
			Sim_UtilSimulateNode(p, pNode, fType, fType, fType)
		}
	}
	// assign simulation info of the CO nodes
	for i := int32(0); i < p.nOutputs; i++ {
		Sim_UtilSimulateNode(p, baseabc.Abc_NtkCo(p.pNtk, i), fType, fType, fType)
	}
}

/*
*Function*************************************************************

	Synopsis    [Simulates one node.]

**********************************************************************
*/
func Sim_UtilSimulateNode(p *Sim_Man_t, pNode *Abc_Obj_t, fType, fType1, fType2 int32) {
	var pSimmNode, pSimmNode1, pSimmNode2 []uint32
	var k int32
	pick := func(f int32, id int) []uint32 {
		if f != 0 {
			return p.vSim1[id]
		}
		return p.vSim0[id]
	}
	// simulate the internal nodes
	if baseabc.Abc_ObjIsNode(pNode) != 0 {
		pSimmNode = pick(fType, pNode.ID)
		pSimmNode1 = pick(fType1, baseabc.Abc_ObjFanin0(pNode).ID)
		pSimmNode2 = pick(fType2, baseabc.Abc_ObjFanin1(pNode).ID)
		fComp1 := baseabc.Abc_ObjFaninC0(pNode) != 0
		fComp2 := baseabc.Abc_ObjFaninC1(pNode) != 0
		if fComp1 && fComp2 {
			for k = 0; k < p.nSimWords; k++ {
				pSimmNode[k] = ^pSimmNode1[k] & ^pSimmNode2[k]
			}
		} else if fComp1 && !fComp2 {
			for k = 0; k < p.nSimWords; k++ {
				pSimmNode[k] = ^pSimmNode1[k] & pSimmNode2[k]
			}
		} else if !fComp1 && fComp2 {
			for k = 0; k < p.nSimWords; k++ {
				pSimmNode[k] = pSimmNode1[k] &^ pSimmNode2[k]
			}
		} else {
			for k = 0; k < p.nSimWords; k++ {
				pSimmNode[k] = pSimmNode1[k] & pSimmNode2[k]
			}
		}
	} else {
		assert(baseabc.Abc_ObjFaninNum(pNode) == 1, "Abc_ObjFaninNum(pNode) == 1")
		pSimmNode = pick(fType, pNode.ID)
		pSimmNode1 = pick(fType1, baseabc.Abc_ObjFanin0(pNode).ID)
		if baseabc.Abc_ObjFaninC0(pNode) != 0 {
			for k = 0; k < p.nSimWords; k++ {
				pSimmNode[k] = ^pSimmNode1[k]
			}
		} else {
			for k = 0; k < p.nSimWords; k++ {
				pSimmNode[k] = pSimmNode1[k]
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Flips the simulation info of the node.]

**********************************************************************
*/
func Sim_UtilInfoFlip(p *Sim_Man_t, pNode *Abc_Obj_t) {
	pSimInfo1 := p.vSim0[pNode.ID]
	pSimInfo2 := p.vSim1[pNode.ID]
	for k := int32(0); k < p.nSimWords; k++ {
		pSimInfo2[k] = ^pSimInfo1[k]
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the simulation infos are equal.]

**********************************************************************
*/
func Sim_UtilInfoCompare(p *Sim_Man_t, pNode *Abc_Obj_t) int32 {
	pSimInfo1 := p.vSim0[pNode.ID]
	pSimInfo2 := p.vSim1[pNode.ID]
	for k := int32(0); k < p.nSimWords; k++ {
		if pSimInfo2[k] != pSimInfo1[k] {
			return 0
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns the positions where pInfo2 is 1 while pInfo1 is 0.]

**********************************************************************
*/
func Sim_UtilInfoDetectNews(pInfo1, pInfo2 []uint32, nWords int32, vDiffs []int32) []int32 {
	vDiffs = vDiffs[:0]
	for w := int32(0); w < nWords; w++ {
		if uMask := pInfo2[w] &^ pInfo1[w]; uMask != 0 {
			for b := int32(0); b < 32; b++ {
				if uMask&(uint32(1)<<uint(b)) != 0 {
					vDiffs = append(vDiffs, 32*w+b)
				}
			}
		}
	}
	return vDiffs
}

/*
*Function*************************************************************

	Synopsis    [Returns the positions where the simulation info differs.]

**********************************************************************
*/
func Sim_UtilInfoDetectDiffs(pInfo1, pInfo2 []uint32, nWords int32, vDiffs []int32) []int32 {
	vDiffs = vDiffs[:0]
	for w := int32(0); w < nWords; w++ {
		if uMask := pInfo2[w] ^ pInfo1[w]; uMask != 0 {
			for b := int32(0); b < 32; b++ {
				if uMask&(uint32(1)<<uint(b)) != 0 {
					vDiffs = append(vDiffs, 32*w+b)
				}
			}
		}
	}
	return vDiffs
}

/*
*Function*************************************************************

	Synopsis    [Counts the total size of the supports.]

**********************************************************************
*/
func Sim_UtilCountSuppSizes(p *Sim_Man_t, fStruct int32) int32 {
	var Counter int32
	if fStruct != 0 {
		for i := int32(0); i < p.nOutputs; i++ {
			pNode := baseabc.Abc_NtkCo(p.pNtk, i)
			for v := int32(0); v < p.nInputs; v++ {
				if symHasBit(p.vSuppStr[pNode.ID], v) {
					Counter++
				}
			}
		}
	} else {
		for i := int32(0); i < p.nOutputs; i++ {
			for v := int32(0); v < p.nInputs; v++ {
				if symHasBit(p.vSuppFun[i], v) {
					Counter++
				}
			}
		}
	}
	return Counter
}
