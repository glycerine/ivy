package abci

// Line-by-line port of ABC's simulation/SAT-based two-variable symmetry
// computation: src/opt/sim/simSym.c, simSymStr.c, simSymSim.c,
// simSymSat.c, the Sym_Man_t part of simMan.c and the helpers of
// simUtils.c/simSupp.c it uses. It works on the strashed Abc_Ntk_t (as C)
// because its SAT checks build Abc_NtkMiterForCofactors miters and prove
// them with the old FRAIG package, whose models steer later simulation.
// The random patterns come from the process-wide rand() stream
// (util.C_Srand/util.Sim_RandomUnsigned), continuing after
// Sim_ComputeFunSupp exactly as in C.

import (
	"fmt"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sim"
	"github.com/glycerine/abcgo/proof/fraig"
)

// Sym_Man_t is sim.h's symmetry manager.
type Sym_Man_t struct {
	// info about the network
	pNtk     *Abc_Ntk_t   // the network
	vNodes   []*Abc_Obj_t // internal nodes in topological order
	nInputs  int32
	nOutputs int32
	// internal simulation information
	nSimWords int32      // the number of bits in simulation info
	vSim      [][]uint32 // simulation info
	// support information
	vSuppFun  [][]uint32 // bit representation
	vSupports [][]int32  // integer representation
	// symmetry info for each output
	vMatrSymms    []*extra.Extra_BitMat_t // symmetric pairs
	vMatrNonSymms []*extra.Extra_BitMat_t // non-symmetric pairs
	vPairsTotal   []int32                 // total pairs
	vPairsSym     []int32                 // symmetric pairs
	vPairsNonSym  []int32                 // non-symmetric pairs
	// temporary simulation info
	uPatRand []uint32
	uPatCol  []uint32
	uPatRow  []uint32
	// temporary
	vVarsU   []int32
	vVarsV   []int32
	iOutput  int32
	iVar1    int32
	iVar2    int32
	iVar1Old int32
	iVar2Old int32
	// internal data structures
	nSatRuns      int32
	nSatRunsSat   int32
	nSatRunsUnsat int32
	// pairs
	nPairsSymm    int32
	nPairsSymmStr int32
	nPairsNonSymm int32
	nPairsRem     int32
	nPairsTotal   int32
	// runtime statistics
	timeStruct int64
	timeCount  int64
	timeMatr   int64
	timeSim    int64
	timeFraig  int64
	timeSat    int64
	timeTotal  int64
}

func symSimNumWords(n int32) int32 { return int32(sim.SIM_NUM_WORDS(int(n))) }

func symInfoAlloc(nSize, nWords int32) [][]uint32 {
	return sim.Sim_UtilInfoAlloc(int(nSize), int(nWords), true)
}

/*
*Function*************************************************************

	Synopsis    [Computes two variable symmetries.]

**********************************************************************
*/
func Sim_ComputeTwoVarSymms(pNtk *Abc_Ntk_t, fVerbose int32) int32 {
	var p *Sym_Man_t
	var Result int32
	var i int32
	var clk int64
	clkTotal := util.Abc_Clock()

	util.C_Srand(0xABC)

	// start the simulation manager
	p = Sym_ManStart(pNtk, fVerbose)
	p.nPairsTotal = Sim_UtilCountAllPairs(p.vSuppFun, p.nSimWords, p.vPairsTotal)
	p.nPairsRem = p.nPairsTotal
	if fVerbose != 0 {
		fmt.Printf("Total = %8d.  Sym = %8d.  NonSym = %8d.  Remaining = %8d.\n",
			p.nPairsTotal, p.nPairsSymm, p.nPairsNonSymm, p.nPairsRem)
	}

	// detect symmetries using circuit structure
	clk = util.Abc_Clock()
	Sim_SymmsStructCompute(pNtk, p.vMatrSymms, p.vSuppFun)
	p.timeStruct = util.Abc_Clock() - clk

	Sim_UtilCountPairsAll(p)
	p.nPairsSymmStr = p.nPairsSymm
	if fVerbose != 0 {
		fmt.Printf("Total = %8d.  Sym = %8d.  NonSym = %8d.  Remaining = %8d.\n",
			p.nPairsTotal, p.nPairsSymm, p.nPairsNonSymm, p.nPairsRem)
	}

	// detect symmetries using simulation
	for i = 1; i <= 1000; i++ {
		// simulate this pattern
		Sim_UtilSetRandom(p.uPatRand, p.nSimWords)
		Sim_SymmsSimulate(p, p.uPatRand, p.vMatrNonSymms)
		if i%50 != 0 {
			continue
		}
		// check disjointness
		assert(Sim_UtilMatrsAreDisjoint(p) != 0, "Sim_UtilMatrsAreDisjoint( p )")
		// count the number of pairs
		Sim_UtilCountPairsAll(p)
		if i%500 != 0 {
			continue
		}
		if fVerbose != 0 {
			fmt.Printf("Total = %8d.  Sym = %8d.  NonSym = %8d.  Remaining = %8d.\n",
				p.nPairsTotal, p.nPairsSymm, p.nPairsNonSymm, p.nPairsRem)
		}
	}

	// detect symmetries using SAT
	for i = 1; Sim_SymmsGetPatternUsingSat(p, p.uPatRand) != 0; i++ {
		// simulate this pattern in four polarities
		Sim_SymmsSimulate(p, p.uPatRand, p.vMatrNonSymms)
		symXorBit(p.uPatRand, p.iVar1)
		Sim_SymmsSimulate(p, p.uPatRand, p.vMatrNonSymms)
		symXorBit(p.uPatRand, p.iVar2)
		Sim_SymmsSimulate(p, p.uPatRand, p.vMatrNonSymms)
		symXorBit(p.uPatRand, p.iVar1)
		Sim_SymmsSimulate(p, p.uPatRand, p.vMatrNonSymms)
		symXorBit(p.uPatRand, p.iVar2)

		if i%10 != 0 {
			continue
		}
		// check disjointness
		assert(Sim_UtilMatrsAreDisjoint(p) != 0, "Sim_UtilMatrsAreDisjoint( p )")
		// count the number of pairs
		Sim_UtilCountPairsAll(p)
		if i%50 != 0 {
			continue
		}
		if fVerbose != 0 {
			fmt.Printf("Total = %8d.  Sym = %8d.  NonSym = %8d.  Remaining = %8d.\n",
				p.nPairsTotal, p.nPairsSymm, p.nPairsNonSymm, p.nPairsRem)
		}
	}

	// count the number of pairs
	Sim_UtilCountPairsAll(p)
	if fVerbose != 0 {
		fmt.Printf("Total = %8d.  Sym = %8d.  NonSym = %8d.  Remaining = %8d.\n",
			p.nPairsTotal, p.nPairsSymm, p.nPairsNonSymm, p.nPairsRem)
	}
	Result = p.nPairsSymm
	p.timeTotal = util.Abc_Clock() - clkTotal
	Sym_ManStop(p)
	return Result
}

func symSetBit(p []uint32, i int32)      { p[i>>5] |= uint32(1) << uint(i&31) }
func symXorBit(p []uint32, i int32)      { p[i>>5] ^= uint32(1) << uint(i&31) }
func symHasBit(p []uint32, i int32) bool { return p[i>>5]&(uint32(1)<<uint(i&31)) != 0 }

/*
*Function*************************************************************

	Synopsis    [Starts the simulation manager.] (simMan.c)

**********************************************************************
*/
func Sym_ManStart(pNtk *Abc_Ntk_t, fVerbose int32) *Sym_Man_t {
	var i, v int32
	// start the manager
	p := &Sym_Man_t{}
	p.pNtk = pNtk
	p.vNodes = baseabc.Abc_NtkDfs(pNtk, 0)
	p.nInputs = baseabc.Abc_NtkCiNum(p.pNtk)
	p.nOutputs = baseabc.Abc_NtkCoNum(p.pNtk)
	// internal simulation information
	p.nSimWords = symSimNumWords(p.nInputs)
	p.vSim = symInfoAlloc(baseabc.Abc_NtkObjNumMax(pNtk), p.nSimWords)
	// symmetry info for each output
	p.vMatrSymms = make([]*extra.Extra_BitMat_t, p.nOutputs)
	p.vMatrNonSymms = make([]*extra.Extra_BitMat_t, p.nOutputs)
	p.vPairsTotal = make([]int32, p.nOutputs)
	p.vPairsSym = make([]int32, p.nOutputs)
	p.vPairsNonSym = make([]int32, p.nOutputs)
	for i = 0; i < p.nOutputs; i++ {
		p.vMatrSymms[i] = extra.Extra_BitMatrixStart(p.nInputs)
		p.vMatrNonSymms[i] = extra.Extra_BitMatrixStart(p.nInputs)
	}
	// temporary patterns
	p.uPatRand = make([]uint32, p.nSimWords)
	p.uPatCol = make([]uint32, p.nSimWords)
	p.uPatRow = make([]uint32, p.nSimWords)
	p.vVarsU = make([]int32, 0, 100)
	p.vVarsV = make([]int32, 0, 100)
	// compute supports
	vSuppFun := Sim_ComputeFunSupp(pNtk, fVerbose)
	p.vSuppFun = make([][]uint32, vec.Vec_PtrSize(vSuppFun))
	for i = 0; i < int32(len(p.vSuppFun)); i++ {
		p.vSuppFun[i] = vec.Vec_PtrEntry(vSuppFun, int(i)).([]uint32)
	}
	p.vSupports = make([][]int32, p.nOutputs)
	for i = 0; i < p.nOutputs; i++ {
		for v = 0; v < p.nInputs; v++ {
			if symHasBit(p.vSuppFun[i], v) {
				p.vSupports[i] = append(p.vSupports[i], v)
			}
		}
	}
	return p
}

/*
*Function*************************************************************

	Synopsis    [Stops the simulation manager.]

**********************************************************************
*/
func Sym_ManStop(p *Sym_Man_t) {
	Sym_ManPrintStats(p)
	for i := int32(0); i < p.nOutputs; i++ {
		extra.Extra_BitMatrixStop(p.vMatrSymms[i])
		extra.Extra_BitMatrixStop(p.vMatrNonSymms[i])
	}
	*p = Sym_Man_t{}
}

/*
*Function*************************************************************

	Synopsis    [Prints the manager statisticis.]

**********************************************************************
*/
func Sym_ManPrintStats(p *Sym_Man_t) {
	fmt.Printf("Total symm         = %8d.\n", p.nPairsSymm)
	fmt.Printf("Structural symm    = %8d.\n", p.nPairsSymmStr)
	fmt.Printf("Total non-sym      = %8d.\n", p.nPairsNonSymm)
	fmt.Printf("Total var pairs    = %8d.\n", p.nPairsTotal)
	fmt.Printf("Sat runs SAT       = %8d.\n", p.nSatRunsSat)
	fmt.Printf("Sat runs UNSAT     = %8d.\n", p.nSatRunsUnsat)
	util.Abc_PrintTime(1, "Structural  ", p.timeStruct)
	util.Abc_PrintTime(1, "Simulation  ", p.timeSim)
	util.Abc_PrintTime(1, "Matrix      ", p.timeMatr)
	util.Abc_PrintTime(1, "Counting    ", p.timeCount)
	util.Abc_PrintTime(1, "Fraiging    ", p.timeFraig)
	util.Abc_PrintTime(1, "SAT         ", p.timeSat)
	util.Abc_PrintTime(1, "TOTAL       ", p.timeTotal)
}

// ---------------------------------------------------------------- simSymStr.c

/*
*Function*************************************************************

	Synopsis    [Computes symmetries for a single output function.]

**********************************************************************
*/
func Sim_SymmsStructCompute(pNtk *Abc_Ntk_t, vMatrs []*extra.Extra_BitMat_t, vSuppFun [][]uint32) {
	var pTemp *Abc_Obj_t
	var i int32
	assert(baseabc.Abc_NtkCiNum(pNtk)+10 < (1<<16), "Abc_NtkCiNum(pNtk) + 10 < (1<<16)")
	s := &simSymStr{
		// get the structural support
		vSupps: Sim_ComputeStrSupp(pNtk),
		// SIM_READ_SYMMS/SIM_SET_SYMMS keep the per-node vectors in pCopy
		symms: make([][]int32, baseabc.Abc_NtkObjNumMax(pNtk)),
	}
	// set elementary info for the CIs
	for i = 0; i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		s.symms[baseabc.Abc_NtkCi(pNtk, i).ID] = []int32{}
	}
	// create the map of CI ids into their numbers
	s.pMap = Sim_SymmsCreateMap(pNtk)
	// collect the nodes in the TFI cone of this output
	vNodes := baseabc.Abc_NtkDfs(pNtk, 0)
	for _, pTemp = range vNodes {
		s.computeOne(pTemp)
	}
	// collect the results for the COs;
	for i = 0; i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		pTemp = baseabc.Abc_ObjFanin0(baseabc.Abc_NtkCo(pNtk, i))
		if baseabc.Abc_ObjIsCi(pTemp) != 0 || baseabc.Abc_AigNodeIsConst(pTemp) != 0 {
			continue
		}
		Sim_SymmsTransferToMatrix(vMatrs[i], s.symms[pTemp.ID], vSuppFun[i])
	}
}

type simSymStr struct {
	vSupps [][]uint32
	symms  [][]int32
	pMap   []int32
}

/*
*Function*************************************************************

	Synopsis    [Recursively computes symmetries.]

**********************************************************************
*/
func (s *simSymStr) computeOne(pNode *Abc_Obj_t) {
	var vNodes, vNodesPi0, vNodesPi1, vNodesOther []baseabc.Edge
	// collect the fanins of the implication supergate
	Sim_SymmsBalanceCollect_rec(baseabc.Edge{Object: pNode}, &vNodes)
	// sort the nodes in the implication supergate
	Sim_SymmsPartitionNodes(vNodes, &vNodesPi0, &vNodesPi1, &vNodesOther)
	// start the resulting set
	vSymms := make([]int32, 0, 10)
	// generate symmetries from the groups
	s.appendFromGroup(vNodesPi0, vNodesOther, &vSymms)
	s.appendFromGroup(vNodesPi1, vNodesOther, &vSymms)
	// add symmetries from other inputs
	for i := range vNodesOther {
		pTemp := vNodesOther[i].Object
		s.appendFromNode(s.symms[pTemp.ID], vNodesOther, vNodesPi0, vNodesPi1, &vSymms)
	}
	// set the symmetry at the node
	s.symms[pNode.ID] = vSymms
}

func symPushUniqueEdge(v *[]baseabc.Edge, e baseabc.Edge) {
	for _, x := range *v {
		if x == e {
			return
		}
	}
	*v = append(*v, e)
}

func symPushUniqueInt(v *[]int32, e int32) {
	for _, x := range *v {
		if x == e {
			return
		}
	}
	*v = append(*v, e)
}

/*
*Function*************************************************************

	Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

**********************************************************************
*/
func Sim_SymmsBalanceCollect_rec(pNode baseabc.Edge, vNodes *[]baseabc.Edge) {
	// if the new node is complemented, another gate begins
	if pNode.Complement {
		symPushUniqueEdge(vNodes, pNode)
		return
	}
	// if pNew is the PI node, return
	if baseabc.Abc_ObjIsCi(pNode.Object) != 0 {
		symPushUniqueEdge(vNodes, pNode)
		return
	}
	// go through the branches
	Sim_SymmsBalanceCollect_rec(baseabc.Abc_ObjChild0(pNode.Object), vNodes)
	Sim_SymmsBalanceCollect_rec(baseabc.Abc_ObjChild1(pNode.Object), vNodes)
}

/*
*Function*************************************************************

	Synopsis    [Divides PI variables into groups.]

**********************************************************************
*/
func Sim_SymmsPartitionNodes(vNodes []baseabc.Edge, vNodesPis0, vNodesPis1, vNodesOther *[]baseabc.Edge) {
	for _, pNode := range vNodes {
		if baseabc.Abc_ObjIsCi(pNode.Object) == 0 {
			*vNodesOther = append(*vNodesOther, pNode)
		} else if pNode.Complement {
			*vNodesPis0 = append(*vNodesPis0, pNode)
		} else {
			*vNodesPis1 = append(*vNodesPis1, pNode)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Makes the product of two partitions.]

**********************************************************************
*/
func (s *simSymStr) appendFromGroup(vNodesPi, vNodesOther []baseabc.Edge, vSymms *[]int32) {
	var uSymm uint32
	if len(vNodesPi) == 0 {
		return
	}
	pMap := s.pMap
	// go through the pairs
	for i := 0; i < len(vNodesPi); i++ {
		for k := i + 1; k < len(vNodesPi); k++ {
			// get the two PI nodes
			pNode1 := vNodesPi[i].Object
			pNode2 := vNodesPi[k].Object
			assert(pMap[pNode1.ID] != pMap[pNode2.ID], "pMap[pNode1->Id] != pMap[pNode2->Id]")
			assert(pMap[pNode1.ID] >= 0, "pMap[pNode1->Id] >= 0")
			assert(pMap[pNode2.ID] >= 0, "pMap[pNode2->Id] >= 0")
			// generate symmetry
			if pMap[pNode1.ID] < pMap[pNode2.ID] {
				uSymm = uint32(pMap[pNode1.ID])<<16 | uint32(pMap[pNode2.ID])
			} else {
				uSymm = uint32(pMap[pNode2.ID])<<16 | uint32(pMap[pNode1.ID])
			}
			// check if symmetry belongs
			if s.isCompatibleWithNodes(uSymm, vNodesOther) {
				symPushUniqueInt(vSymms, int32(uSymm))
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Add the filters symmetries from the nodes.]

**********************************************************************
*/
func (s *simSymStr) appendFromNode(vSymms0 []int32, vNodesOther, vNodesPi0, vNodesPi1 []baseabc.Edge, vSymms *[]int32) {
	if len(vSymms0) == 0 {
		return
	}
	// go through the pairs
	for i := range vSymms0 {
		uSymm := uint32(vSymms0[i])
		// check if symmetry belongs
		if s.isCompatibleWithNodes(uSymm, vNodesOther) &&
			s.isCompatibleWithGroup(uSymm, vNodesPi0) &&
			s.isCompatibleWithGroup(uSymm, vNodesPi1) {
			symPushUniqueInt(vSymms, int32(uSymm))
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if symmetry is compatible with the group of nodes.]

**********************************************************************
*/
func (s *simSymStr) isCompatibleWithNodes(uSymm uint32, vNodesOther []baseabc.Edge) bool {
	if len(vNodesOther) == 0 {
		return true
	}
	// get the indices of the PI variables
	Ind1 := int32(uSymm & 0xffff)
	Ind2 := int32(uSymm >> 16)
	// go through the nodes
	// if they do not belong to a support, it is okay
	// if one belongs, the other does not belong, quit
	// if they belong, but are not part of symmetry, quit
	for i := range vNodesOther {
		pNode := vNodesOther[i].Object
		fIsVar1 := symHasBit(s.vSupps[pNode.ID], Ind1)
		fIsVar2 := symHasBit(s.vSupps[pNode.ID], Ind2)
		if !fIsVar1 && !fIsVar2 {
			continue
		}
		if fIsVar1 != fIsVar2 {
			return false
		}
		// both belong
		// check if there is a symmetry
		vSymmsNode := s.symms[pNode.ID]
		found := false
		for _, x := range vSymmsNode {
			if uSymm == uint32(x) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if symmetry is compatible with the group of PIs.]

**********************************************************************
*/
func (s *simSymStr) isCompatibleWithGroup(uSymm uint32, vNodesPi []baseabc.Edge) bool {
	if len(vNodesPi) == 0 {
		return true
	}
	// get the indices of the PI variables
	Ind1 := int32(uSymm & 0xffff)
	Ind2 := int32(uSymm >> 16)
	// go through the PI nodes
	fHasVar1, fHasVar2 := false, false
	for i := range vNodesPi {
		pNode := vNodesPi[i].Object
		if s.pMap[pNode.ID] == Ind1 {
			fHasVar1 = true
		} else if s.pMap[pNode.ID] == Ind2 {
			fHasVar2 = true
		}
	}
	return fHasVar1 == fHasVar2
}

/*
*Function*************************************************************

	Synopsis    [Transfers from the symmetry vector to the matrix.]

**********************************************************************
*/
func Sim_SymmsTransferToMatrix(pMatSymm *extra.Extra_BitMat_t, vSymms []int32, pSupport []uint32) {
	// add diagonal elements
	nInputs := extra.Extra_BitMatrixReadSize(pMatSymm)
	for i := int32(0); i < nInputs; i++ {
		extra.Extra_BitMatrixInsert1(pMatSymm, i, i)
	}
	// add non-diagonal elements
	for i := range vSymms {
		uSymm := uint32(vSymms[i])
		Ind1 := int32(uSymm & 0xffff)
		Ind2 := int32(uSymm >> 16)
		// skip variables that are not in the true support
		assert(symHasBit(pSupport, Ind1) == symHasBit(pSupport, Ind2), "Sim_HasBit(pSupport, Ind1) == Sim_HasBit(pSupport, Ind2)")
		if !symHasBit(pSupport, Ind1) || !symHasBit(pSupport, Ind2) {
			continue
		}
		extra.Extra_BitMatrixInsert1(pMatSymm, Ind1, Ind2)
		extra.Extra_BitMatrixInsert2(pMatSymm, Ind1, Ind2)
	}
}

/*
*Function*************************************************************

	Synopsis    [Mapping of indices into numbers.]

**********************************************************************
*/
func Sim_SymmsCreateMap(pNtk *Abc_Ntk_t) []int32 {
	pMap := make([]int32, baseabc.Abc_NtkObjNumMax(pNtk))
	for i := range pMap {
		pMap[i] = -1
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		pMap[baseabc.Abc_NtkCi(pNtk, i).ID] = i
	}
	return pMap
}

/*
*Function*************************************************************

	Synopsis    [Computes structural supports.] (simSupp.c)

	Description [Supports are returned as arrays of bit strings, one for each CO.]

**********************************************************************
*/
func Sim_ComputeStrSupp(pNtk *Abc_Ntk_t) [][]uint32 {
	var i, k int32
	// allocate room for structural supports
	nSuppWords := symSimNumWords(baseabc.Abc_NtkCiNum(pNtk))
	vSuppStr := symInfoAlloc(baseabc.Abc_NtkObjNumMax(pNtk), nSuppWords)
	// assign the structural support to the PIs
	for i = 0; i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		symSetBit(vSuppStr[baseabc.Abc_NtkCi(pNtk, i).ID], i)
	}
	// derive the structural supports of the internal nodes
	for _, pNode := range pNtk.Objects() {
		if baseabc.Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		pSimmNode := vSuppStr[pNode.ID]
		pSimmNode1 := vSuppStr[baseabc.Abc_ObjFanin0(pNode).ID]
		pSimmNode2 := vSuppStr[baseabc.Abc_ObjFanin1(pNode).ID]
		for k = 0; k < nSuppWords; k++ {
			pSimmNode[k] = pSimmNode1[k] | pSimmNode2[k]
		}
	}
	// set the structural supports of the PO nodes
	for i = 0; i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCo(pNtk, i)
		pSimmNode := vSuppStr[pNode.ID]
		pSimmNode1 := vSuppStr[baseabc.Abc_ObjFanin0(pNode).ID]
		for k = 0; k < nSuppWords; k++ {
			pSimmNode[k] = pSimmNode1[k]
		}
	}
	return vSuppStr
}

// ---------------------------------------------------------------- simSymSim.c

/*
*Function*************************************************************

	Synopsis    [Detects non-symmetric pairs using one pattern.]

**********************************************************************
*/
func Sim_SymmsSimulate(p *Sym_Man_t, pPat []uint32, vMatrsNonSym []*extra.Extra_BitMat_t) {
	var clk int64
	// create the simulation matrix
	Sim_SymmsCreateSquare(p, pPat)
	// simulate each node in the DFS order
	clk = util.Abc_Clock()
	for _, pNode := range p.vNodes {
		Sim_UtilSimulateNodeOne(pNode, p.vSim, p.nSimWords, 0)
	}
	p.timeSim += util.Abc_Clock() - clk
	// collect info into the CO matrices
	clk = util.Abc_Clock()
	for i := int32(0); i < p.nOutputs; i++ {
		pNode := baseabc.Abc_ObjFanin0(baseabc.Abc_NtkCo(p.pNtk, i))
		nPairsTotal := p.vPairsTotal[i]
		nPairsSym := p.vPairsSym[i]
		nPairsNonSym := p.vPairsNonSym[i]
		assert(nPairsTotal >= nPairsSym+nPairsNonSym, "nPairsTotal >= nPairsSym + nPairsNonSym")
		if nPairsTotal == nPairsSym+nPairsNonSym {
			continue
		}
		Sim_SymmsDeriveInfo(p, pPat, pNode, vMatrsNonSym, i)
	}
	p.timeMatr += util.Abc_Clock() - clk
}

/*
*Function*************************************************************

	Synopsis    [Creates the square matrix of simulation info.]

**********************************************************************
*/
func Sim_SymmsCreateSquare(p *Sym_Man_t, pPat []uint32) {
	var i, w int32
	// for each PI var copy the pattern
	for i = 0; i < p.nInputs; i++ {
		pSimInfo := p.vSim[baseabc.Abc_NtkCi(p.pNtk, i).ID]
		if symHasBit(pPat, i) {
			for w = 0; w < p.nSimWords; w++ {
				pSimInfo[w] = ^uint32(0) // SIM_MASK_FULL
			}
		} else {
			for w = 0; w < p.nSimWords; w++ {
				pSimInfo[w] = 0
			}
		}
		// flip one bit
		symXorBit(pSimInfo, i)
	}
}

/*
*Function*************************************************************

	Synopsis    [Transfers the info to the POs.]

**********************************************************************
*/
func Sim_SymmsDeriveInfo(p *Sym_Man_t, pPat []uint32, pNode *Abc_Obj_t, vMatrsNonSym []*extra.Extra_BitMat_t, Output int32) {
	var w int32
	// get the matrix, the support, and the simulation info
	pMat := vMatrsNonSym[Output]
	vSupport := p.vSupports[Output]
	pSupport := p.vSuppFun[Output]
	pSimInfo := p.vSim[pNode.ID]
	// generate vectors A1 and A2
	for w = 0; w < p.nSimWords; w++ {
		p.uPatCol[w] = pSupport[w] & pPat[w] & pSimInfo[w]
		p.uPatRow[w] = pSupport[w] & pPat[w] &^ pSimInfo[w]
	}
	// add two dimensions
	for _, i := range vSupport {
		if symHasBit(p.uPatCol, i) {
			extra.Extra_BitMatrixOr(pMat, i, p.uPatRow)
		}
	}
	// add two dimensions
	for _, i := range vSupport {
		if symHasBit(p.uPatRow, i) {
			extra.Extra_BitMatrixOr(pMat, i, p.uPatCol)
		}
	}
	// generate vectors B1 and B2
	for w = 0; w < p.nSimWords; w++ {
		p.uPatCol[w] = pSupport[w] &^ pPat[w] & pSimInfo[w]
		p.uPatRow[w] = pSupport[w] &^ pPat[w] &^ pSimInfo[w]
	}
	// add two dimensions
	for _, i := range vSupport {
		if symHasBit(p.uPatCol, i) {
			extra.Extra_BitMatrixOr(pMat, i, p.uPatRow)
		}
	}
	// add two dimensions
	for _, i := range vSupport {
		if symHasBit(p.uPatRow, i) {
			extra.Extra_BitMatrixOr(pMat, i, p.uPatCol)
		}
	}
}

// ---------------------------------------------------------------- simSymSat.c

/*
*Function*************************************************************

	Synopsis    [Tries to prove the remaining pairs using SAT.]

	Description [Continues to prove as long as it encounters symmetric pairs.
	Returns 1 if a non-symmetric pair is found (which gives a counter-example).
	Returns 0 if it finishes considering all pairs for all outputs.]

**********************************************************************
*/
func Sim_SymmsGetPatternUsingSat(p *Sym_Man_t, pPattern []uint32) int32 {
	// iterate through outputs
	for out := p.iOutput; out < p.nOutputs; out++ {
		pMatSym := p.vMatrSymms[out]
		pMatNonSym := p.vMatrNonSymms[out]

		// go through the remaining variable pairs
		vSupport := p.vSupports[out]
		for Index1, v := range vSupport {
			for _, u := range vSupport[Index1+1:] {
				if extra.Extra_BitMatrixLookup1(pMatSym, v, u) != 0 || extra.Extra_BitMatrixLookup1(pMatNonSym, v, u) != 0 {
					continue
				}
				p.nSatRuns++

				// collect the support variables that are symmetric with u and v
				p.vVarsU = p.vVarsU[:0]
				p.vVarsV = p.vVarsV[:0]
				for _, b := range vSupport {
					if extra.Extra_BitMatrixLookup1(pMatSym, u, b) != 0 {
						p.vVarsU = append(p.vVarsU, b)
					}
					if extra.Extra_BitMatrixLookup1(pMatSym, v, b) != 0 {
						p.vVarsV = append(p.vVarsV, b)
					}
				}

				if Sim_SymmsSatProveOne(p, out, v, u, pPattern) != 0 { // update the symmetric variable info
					p.nSatRunsUnsat++
					for _, i := range p.vVarsU {
						for _, k := range p.vVarsV {
							extra.Extra_BitMatrixInsert1(pMatSym, i, k)  // Theorem 1
							extra.Extra_BitMatrixInsert2(pMatSym, i, k)  // Theorem 1
							extra.Extra_BitMatrixOrTwo(pMatNonSym, i, k) // Theorem 2
						}
					}
				} else { // update the assymmetric variable info
					p.nSatRunsSat++
					for _, i := range p.vVarsU {
						for _, k := range p.vVarsV {
							extra.Extra_BitMatrixInsert1(pMatNonSym, i, k) // Theorem 3
							extra.Extra_BitMatrixInsert2(pMatNonSym, i, k) // Theorem 3
						}
					}

					// remember the out
					p.iOutput = out
					p.iVar1Old = p.iVar1
					p.iVar2Old = p.iVar2
					p.iVar1 = v
					p.iVar2 = u
					return 1
				}
			}
		}
		// make sure that the symmetry matrix contains only cliques
		assert(extra.Extra_BitMatrixIsClique(pMatSym) != 0, "Extra_BitMatrixIsClique( pMatSym )")
	}

	// mark that we finished all outputs
	p.iOutput = p.nOutputs
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the variables are symmetric; 0 otherwise.]

**********************************************************************
*/
func Sim_SymmsSatProveOne(p *Sym_Man_t, Out, Var1, Var2 int32, pPattern []uint32) int32 {
	var Params Fraig_Params_t
	var RetValue, i int32
	var clk int64

	// get the miter for this problem
	pMiter := Abc_NtkMiterForCofactors(p.pNtk, Out, Var1, Var2)
	// transform the miter into a fraig
	fraig.Fraig_ParamsSetDefault(&Params)
	Params.FInternal = 1
	Params.NPatsRand = 512
	Params.NPatsDyna = 512
	Params.NSeconds = util.ABC_INFINITY

	clk = util.Abc_Clock()
	pMan := Abc_NtkToFraig(pMiter, &Params, 0, 0).(*Fraig_Man_t)
	p.timeFraig += util.Abc_Clock() - clk
	clk = util.Abc_Clock()
	fraig.Fraig_ManProveMiter(pMan)
	p.timeSat += util.Abc_Clock() - clk

	// analyze the result
	RetValue = fraig.Fraig_ManCheckMiter(pMan)
	// save the pattern
	if RetValue == 0 {
		// get the pattern
		pModel := fraig.Fraig_ManReadModel(pMan)
		assert(pModel != nil, "pModel != NULL")
		// transfer the model into the pattern
		for i = 0; i < p.nSimWords; i++ {
			pPattern[i] = 0
		}
		for i = 0; i < p.nInputs; i++ {
			if pModel[i] != 0 {
				symSetBit(pPattern, i)
			}
		}
		// make sure these variables have the same value (1)
		symSetBit(pPattern, Var1)
		symSetBit(pPattern, Var2)
	} else if RetValue == -1 {
		// this should never happen; if it happens, such is life
		// we are conservative and assume that there is no symmetry
		for i = 0; i < p.nSimWords; i++ {
			pPattern[i] = 0
		}
		RetValue = 0
	}
	// delete the fraig manager
	fraig.Fraig_ManFree(pMan)
	// delete the miter
	baseabc.Abc_NtkDelete(pMiter)
	return RetValue
}

// ---------------------------------------------------------------- simUtils.c

/*
*Function*************************************************************

	Synopsis    [Simulates one node.]

**********************************************************************
*/
func Sim_UtilSimulateNodeOne(pNode *Abc_Obj_t, vSimInfo [][]uint32, nSimWords int32, nOffset int32) {
	var k int32
	// simulate the internal nodes
	assert(baseabc.Abc_ObjIsNode(pNode) != 0, "Abc_ObjIsNode(pNode)")
	pSimmNode := vSimInfo[pNode.ID][nOffset:]
	pSimmNode1 := vSimInfo[baseabc.Abc_ObjFanin0(pNode).ID][nOffset:]
	pSimmNode2 := vSimInfo[baseabc.Abc_ObjFanin1(pNode).ID][nOffset:]
	fComp1 := baseabc.Abc_ObjFaninC0(pNode) != 0
	fComp2 := baseabc.Abc_ObjFaninC1(pNode) != 0
	if fComp1 && fComp2 {
		for k = 0; k < nSimWords; k++ {
			pSimmNode[k] = ^pSimmNode1[k] & ^pSimmNode2[k]
		}
	} else if fComp1 && !fComp2 {
		for k = 0; k < nSimWords; k++ {
			pSimmNode[k] = ^pSimmNode1[k] & pSimmNode2[k]
		}
	} else if !fComp1 && fComp2 {
		for k = 0; k < nSimWords; k++ {
			pSimmNode[k] = pSimmNode1[k] &^ pSimmNode2[k]
		}
	} else {
		for k = 0; k < nSimWords; k++ {
			pSimmNode[k] = pSimmNode1[k] & pSimmNode2[k]
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns random patterns.]

**********************************************************************
*/
func Sim_UtilSetRandom(pPatRand []uint32, nSimWords int32) {
	for k := int32(0); k < nSimWords; k++ {
		pPatRand[k] = util.Sim_RandomUnsigned() // SIM_RANDOM_UNSIGNED
	}
}

/*
*Function*************************************************************

	Synopsis    [Counts the total number of pairs.]

**********************************************************************
*/
func Sim_UtilCountAllPairs(vSuppFun [][]uint32, nSimWords int32, vCounters []int32) int32 {
	var Counter int32
	for i, pSupp := range vSuppFun {
		nOnes := int32(sim.Sim_UtilCountOnes(pSupp, int(nSimWords)))
		nPairs := nOnes * (nOnes - 1) / 2
		vCounters[i] = nPairs
		Counter += nPairs
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of entries in the array of matrices.]

**********************************************************************
*/
func Sim_UtilCountPairsOne(pMat *extra.Extra_BitMat_t, vSupport []int32) int32 {
	var Counter int32
	for Index1, i := range vSupport {
		for _, k := range vSupport[Index1+1:] {
			Counter += extra.Extra_BitMatrixLookup1(pMat, i, k)
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of entries in the array of matrices.]

**********************************************************************
*/
func Sim_UtilCountPairsAll(p *Sym_Man_t) {
	var nPairsTotal, nPairsSym, nPairsNonSym, i int32
	clk := util.Abc_Clock()
	p.nPairsSymm = 0
	p.nPairsNonSymm = 0
	for i = 0; i < p.nOutputs; i++ {
		nPairsTotal = p.vPairsTotal[i]
		nPairsSym = p.vPairsSym[i]
		nPairsNonSym = p.vPairsNonSym[i]
		assert(nPairsTotal >= nPairsSym+nPairsNonSym, "nPairsTotal >= nPairsSym + nPairsNonSym")
		if nPairsTotal == nPairsSym+nPairsNonSym {
			p.nPairsSymm += nPairsSym
			p.nPairsNonSymm += nPairsNonSym
			continue
		}
		nPairsSym = Sim_UtilCountPairsOne(p.vMatrSymms[i], p.vSupports[i])
		nPairsNonSym = Sim_UtilCountPairsOne(p.vMatrNonSymms[i], p.vSupports[i])
		assert(nPairsTotal >= nPairsSym+nPairsNonSym, "nPairsTotal >= nPairsSym + nPairsNonSym")
		p.vPairsSym[i] = nPairsSym
		p.vPairsNonSym[i] = nPairsNonSym
		p.nPairsSymm += nPairsSym
		p.nPairsNonSymm += nPairsNonSym
	}
	p.nPairsRem = p.nPairsTotal - p.nPairsSymm - p.nPairsNonSymm
	p.timeCount += util.Abc_Clock() - clk
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the symmetric and non-symmetric matrices are disjoint.]

**********************************************************************
*/
func Sim_UtilMatrsAreDisjoint(p *Sym_Man_t) int32 {
	for i := int32(0); i < p.nOutputs; i++ {
		if extra.Extra_BitMatrixIsDisjoint(p.vMatrSymms[i], p.vMatrNonSymms[i]) == 0 {
			return 0
		}
	}
	return 1
}
