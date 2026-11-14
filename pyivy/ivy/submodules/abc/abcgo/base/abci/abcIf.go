package abci

import aig_gia "github.com/glycerine/abcgo/aig/gia"

import (
	"fmt"
	"math"
	"os"
	"sort"
	"unsafe"

	hoppkg "github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [abcIf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface with the FPGA mapping package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: abcIf.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "map/if/if.h"
//#include "bool/kit/kit.h"
//#include "aig/aig/aig.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern If_Man_t *  Abc_NtkToIf( Abc_Ntk_t * pNtk, If_Par_t * pPars );

// static Abc_Ntk_t * Abc_NtkFromIf( If_Man_t * pIfMan, Abc_Ntk_t * pNtk );

// extern Abc_Obj_t * Abc_NodeFromIf_rec( Abc_Ntk_t * pNtkNew, If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vCover );

// static Hop_Obj_t * Abc_NodeIfToHop( Hop_Man_t * pHopMan, If_Man_t * pIfMan, If_Obj_t * pIfObj );

// static Vec_Ptr_t * Abc_NtkFindGoodOrder( Abc_Ntk_t * pNtk );

// extern void Abc_NtkBddReorder( Abc_Ntk_t * pNtk, int fVerbose );

// extern void Abc_NtkBidecResyn( Abc_Ntk_t * pNtk, int fVerbose );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Interface with the FPGA mapping package.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManComputeSwitching(pIfMan *If_Man_t) {
	clk := Abc_Clock()
	assert(!pIfMan.HasSwitchingVector(), "pIfMan.vSwitching == nil")
	pNew := Gia_ManStart(ifpkg.If_ManObjNum(pIfMan))
	vCopy := Vec_IntAlloc(ifpkg.If_ManObjNum(pIfMan))
	Vec_IntPush(vCopy, 1)
	for i := int32(0); i < ifpkg.If_ManCiNum(pIfMan); i++ {
		Vec_IntPush(vCopy, Gia_ManAppendCi(pNew))
	}
	for i := int32(0); i < ifpkg.If_ManObjNum(pIfMan); i++ {
		pIfObj := ifpkg.If_ManObj(pIfMan, i)
		if ifpkg.If_ObjIsAnd(pIfObj) == 0 {
			continue
		}
		iLit0 := Abc_LitNotCond(Vec_IntEntry(vCopy, ifpkg.If_ObjId(ifpkg.If_ObjFanin0(pIfObj))), ifpkg.If_ObjFaninC0(pIfObj))
		iLit1 := Abc_LitNotCond(Vec_IntEntry(vCopy, ifpkg.If_ObjId(ifpkg.If_ObjFanin1(pIfObj))), ifpkg.If_ObjFaninC1(pIfObj))
		Vec_IntPush(vCopy, Gia_ManAppendAnd(pNew, iLit0, iLit1))
	}
	for i := int32(0); i < ifpkg.If_ManCoNum(pIfMan); i++ {
		pIfObj := ifpkg.If_ManCo(pIfMan, i)
		iLit0 := Abc_LitNotCond(Vec_IntEntry(vCopy, ifpkg.If_ObjId(ifpkg.If_ObjFanin0(pIfObj))), ifpkg.If_ObjFaninC0(pIfObj))
		Vec_IntPush(vCopy, Gia_ManAppendCo(pNew, iLit0))
	}
	assert(Vec_IntSize(vCopy) == ifpkg.If_ManObjNum(pIfMan), "Vec_IntSize(vCopy) == If_ManObjNum(pIfMan)")
	Vec_IntFree(vCopy)
	probabilities := aig_gia.Gia_ManComputeSwitchProbs(pNew, 48, 16, 0)
	values := make([]float32, Vec_IntSize(probabilities))
	for i := range values {
		values[i] = math.Float32frombits(uint32(Vec_IntEntry(probabilities, i)))
	}
	Vec_IntFree(probabilities)
	pIfMan.SetSwitchingProbabilities(values)
	Gia_ManStop(pNew)
	if pIfMan.Parameters().Verbose() {
		Abc_PrintTime(1, "Computing switching activity", Abc_Clock()-clk)
	}
}

/*
*Function*************************************************************

	Synopsis    [Interface with the FPGA mapping package.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIf(pNtk *Abc_Ntk_t, pPars *If_Par_t) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	pPars.SetArrivalTimes(abc.Abc_NtkMappingArrivalTimes(pNtk))
	pPars.SetRequiredTimes(abc.Abc_NtkMappingRequiredTimes(pNtk))
	delaySensitive := pPars.DelayOptimization() || pPars.DsdBalance() || pPars.UserRecLibrary() || pPars.UsesUserSesLib()
	if delaySensitive && abc.Abc_NtkHasMappingTiming(pNtk) {
		if pNtk.AndGateDelay == 0 {
			if library, ok := abcmain.Abc_FrameReadLibGen().(*mio.Mio_Library_t); ok && library != nil {
				pNtk.AndGateDelay = float64(mio.Mio_LibraryReadDelayAigNode(library))
			}
			if pNtk.AndGateDelay == 0 {
				pNtk.AndGateDelay = 1
				fmt.Printf("The AIG-node delay is not set. Assuming unit-delay.\n")
			}
		}
		arrivals, required := pPars.ArrivalTimes(), pPars.RequiredTimes()
		for i := range arrivals {
			arrivals[i] /= float32(pNtk.AndGateDelay)
		}
		for i := range required {
			required[i] /= float32(pNtk.AndGateDelay)
		}
		pPars.SetArrivalTimes(arrivals)
		pPars.SetRequiredTimes(required)
	}
	if pPars.LatchPaths() {
		arrivals := pPars.ArrivalTimes()
		for i := 0; i < int(abc.Abc_NtkPiNum(pNtk)) && i < len(arrivals); i++ {
			arrivals[i] = -float32(ifpkg.IF_INFINITY)
		}
		pPars.SetArrivalTimes(arrivals)
	}
	pIfMan := Abc_NtkToIf(pNtk, pPars)
	if pIfMan == nil {
		return nil
	}
	if pPars.Power() {
		If_ManComputeSwitching(pIfMan)
	}
	if pPars.UsesDsd() {
		p, ok := abcmain.Abc_FrameReadManDsd().(*ifpkg.If_DsdMan_t)
		assert(ok && p != nil, "DSD manager is available")
		assert(pPars.LUTSize() <= ifpkg.If_DsdManVarNum(p), "pPars.nLutSize <= If_DsdManVarNum(p)")
		structure := pPars.LUTStructure()
		assert((structure == "" && ifpkg.If_DsdManLutSize(p) == 0) || (structure != "" && int32(structure[0]-'0') == ifpkg.If_DsdManLutSize(p)), "valid LUT structure")
		pIfMan.SetDsdManager(p)
		pPars.SetDsdManager(p)
		if pPars.DsdBalance() {
			ifpkg.If_DsdManAllocIsops(p, pPars.LUTSize())
		}
	}
	if ifpkg.If_ManPerformMapping(pIfMan) == 0 {
		ifpkg.If_ManStop(pIfMan)
		return nil
	}
	pNtkNew := Abc_NtkFromIf(pIfMan, pNtk)
	if pNtkNew == nil {
		return nil
	}
	ifpkg.If_ManStop(pIfMan)
	if pPars.DelayOptimization() || pPars.DsdBalance() || pPars.UserRecLibrary() {
		pTemp := pNtkNew
		pNtkNew = abc.Abc_NtkStrash(pTemp, 0, 0, 0)
		abc.Abc_NtkDelete(pTemp)
	} else if pPars.Bidec() && pPars.LUTSize() <= 8 {
		Abc_NtkBidecResyn(pNtkNew, 0)
	}
	if pNtk.Exdc != nil {
		pNtkNew.Exdc = abc.Abc_NtkDup(pNtk.Exdc)
	}
	if abc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Abc_NtkIf: The network check has failed.\n")
		abc.Abc_NtkDelete(pNtkNew)
		return nil
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Load the network into FPGA manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_ObjIfCopy(pNode *Abc_Obj_t) *If_Obj_t {
	p, _ := abc.Abc_ObjCopyData(pNode).(*ifpkg.If_Obj_t)
	return p
}
func Abc_NtkToIf(pNtk *Abc_Ntk_t, pPars *If_Par_t) *If_Man_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")

	man := ifpkg.If_ManStart(pPars)
	ifpkg.If_ManSetName(man, abc.Abc_NtkName(pNtk))
	objects := abc.Abc_NtkObjNum(pNtk)
	bytes := ifpkg.If_ManObjBytes(man)
	if float64(objects)*float64(bytes)/(1<<30) > 1 {
		fmt.Printf("Warning: The mapper will allocate %.1f GB for to represent the subject graph with %d AIG nodes.\n", float64(objects)*float64(bytes)/(1<<30), objects)
	}

	abc.Abc_NtkCleanCopy(pNtk)
	constOne := abc.Abc_AigConst1(pNtk).Object
	abc.Abc_ObjSetCopyData(constOne, ifpkg.If_ManConst1(man))
	for i := int32(0); i < abc.Abc_NtkCiNum(pNtk); i++ {
		ci := abc.Abc_NtkCi(pNtk, i)
		mapped := ifpkg.If_ManCreateCi(man)
		ifpkg.If_ObjSetLevel(mapped, abc.Abc_ObjLevel(ci))
		abc.Abc_ObjSetCopyData(ci, mapped)
	}

	progress := extra.Extra_ProgressBarStart(os.Stdout, abc.Abc_NtkObjNumMax(pNtk))
	nodes := abc.Abc_AigDfs(pNtk, 0, 0)
	for i, node := range nodes {
		extra.Extra_ProgressBarUpdate(progress, int32(i), "Initial")
		fanin0 := abc.Abc_ObjCopyData(abc.Abc_ObjFanin0(node)).(*ifpkg.If_Obj_t)
		fanin1 := abc.Abc_ObjCopyData(abc.Abc_ObjFanin1(node)).(*ifpkg.If_Obj_t)
		mapped := ifpkg.If_ManCreateAnd(man,
			ifpkg.If_NotCond(fanin0, abc.Abc_ObjFaninC0(node)),
			ifpkg.If_NotCond(fanin1, abc.Abc_ObjFaninC1(node)))
		abc.Abc_ObjSetCopyData(node, mapped)
		if abc.Abc_AigNodeIsChoice(node) != 0 {
			for prev, equiv := node, abc.Abc_ObjChoice(node); equiv != nil; prev, equiv = equiv, abc.Abc_ObjChoice(equiv) {
				prevMapped := abc.Abc_ObjCopyData(prev).(*ifpkg.If_Obj_t)
				equivMapped := abc.Abc_ObjCopyData(equiv).(*ifpkg.If_Obj_t)
				assert(ifpkg.If_ObjId(prevMapped) > ifpkg.If_ObjId(equivMapped), "choice nodes are topologically ordered")
				ifpkg.If_ObjSetChoice(prevMapped, equivMapped)
			}
			ifpkg.If_ManCreateChoice(man, mapped)
		}
	}
	extra.Extra_ProgressBarStop(progress)

	for i := int32(0); i < abc.Abc_NtkCoNum(pNtk); i++ {
		co := abc.Abc_NtkCo(pNtk, i)
		driver := abc.Abc_ObjCopyData(abc.Abc_ObjFanin0(co)).(*ifpkg.If_Obj_t)
		mapped := ifpkg.If_ManCreateCo(man, ifpkg.If_NotCond(driver, abc.Abc_ObjFaninC0(co)))
		abc.Abc_ObjSetCopyData(co, mapped)
	}
	return man
}

/*
*Function*************************************************************

	Synopsis    [Box mapping procedures.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_MapBoxSetPrevNext(vDrivers *Vec_Ptr_t, vMapIn *Vec_Int_t, vMapOut *Vec_Int_t, Id int32) {
	input := vec.Vec_PtrEntry(vDrivers, int(Id+2)).(*Abc_Obj_t)
	output := vec.Vec_PtrEntry(vDrivers, int(Id+4)).(*Abc_Obj_t)
	vec.Vec_IntWriteEntry(vMapIn, int(abc.Abc_ObjId(abc.Abc_ObjFanin0(input))), Id)
	vec.Vec_IntWriteEntry(vMapOut, int(abc.Abc_ObjId(abc.Abc_ObjFanin0(output))), Id)
}
func Abc_MapBox2Next(vDrivers *Vec_Ptr_t, vMapIn *Vec_Int_t, vMapOut *Vec_Int_t, Id int32) int32 {
	node := vec.Vec_PtrEntry(vDrivers, int(Id+4)).(*Abc_Obj_t)
	return vec.Vec_IntEntry(vMapIn, int(abc.Abc_ObjId(abc.Abc_ObjFanin0(node))))
}
func Abc_MapBox2Prev(vDrivers *Vec_Ptr_t, vMapIn *Vec_Int_t, vMapOut *Vec_Int_t, Id int32) int32 {
	node := vec.Vec_PtrEntry(vDrivers, int(Id+2)).(*Abc_Obj_t)
	return vec.Vec_IntEntry(vMapOut, int(abc.Abc_ObjId(abc.Abc_ObjFanin0(node))))
}

/*
*Function*************************************************************

	Synopsis    [Creates the mapped network.]

	Description [Assuming the copy field of the mapped nodes are NULL.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromIf(pIfMan *If_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	pars := pIfMan.Parameters()
	kind := abc.ABC_FUNC_AIG
	if pars.UsesBdds() || pars.UsesCnfs() || pars.UsesMv() {
		kind = abc.ABC_FUNC_BDD
	} else if pars.UsesSops() || pars.UsesUserSesLib() || pars.GateSize() > 0 {
		kind = abc.ABC_FUNC_SOP
	}
	ntkNew := abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_LOGIC, kind)
	ifpkg.If_ManCleanNodeCopy(pIfMan)
	ifpkg.If_ManCleanCutData(pIfMan)
	ifpkg.If_ObjSetCopy(ifpkg.If_ManConst1(pIfMan), abc.Abc_NtkCreateNodeConst1(ntkNew))
	for i := int32(0); i < abc.Abc_NtkCiNum(pNtk); i++ {
		ifpkg.If_ObjSetCopy(ifpkg.If_ManCi(pIfMan, i), abc.Abc_ObjCopyEdge(abc.Abc_NtkCi(pNtk, i)).Object)
	}

	cover := vec.Vec_IntAlloc(1 << 16)
	progress := extra.Extra_ProgressBarStart(os.Stdout, abc.Abc_NtkCoNum(pNtk))
	for i := int32(0); i < abc.Abc_NtkCoNum(pNtk); i++ {
		extra.Extra_ProgressBarUpdate(progress, i, "Final")
		coIf := ifpkg.If_ManCo(pIfMan, i)
		driver := Abc_NodeFromIf_rec(ntkNew, pIfMan, ifpkg.If_ObjFanin0(coIf), cover)
		abc.Abc_ObjAddFanin(abc.Abc_ObjCopyEdge(abc.Abc_NtkCo(pNtk, i)).Object, abc.Abc_ObjNotCond(driver, ifpkg.If_ObjFaninC0(coIf)))
	}
	extra.Extra_ProgressBarStop(progress)
	vec.Vec_IntFree(cover)

	constNode, _ := ifpkg.If_ObjCopy(ifpkg.If_ManConst1(pIfMan)).(*Abc_Obj_t)
	if constNode != nil && abc.Abc_ObjFanoutNum(constNode) == 0 {
		abc.Abc_NtkDeleteObj(constNode)
	}
	if pars.UsesBdds() || pars.UsesCnfs() || pars.UsesMv() {
		abc.Abc_NtkSweep(ntkNew, 0)
	}
	if pars.UsesBdds() {
		Abc_NtkBddReorder(ntkNew, 0)
	}
	duplicateCOs := int32(1)
	if pars.UsesBuffs() {
		duplicateCOs = 0
	}
	duplicates := abc.Abc_NtkLogicMakeSimpleCos(ntkNew, duplicateCOs)
	if duplicates != 0 && pars.Verbose() && abcmain.Abc_FrameIsFlagEnabled("silentmode") == 0 {
		if pars.UsesBuffs() {
			fmt.Printf("Added %d buffers/inverters to decouple the CO drivers.\n", duplicates)
		} else {
			fmt.Printf("Duplicated %d gates to decouple the CO drivers.\n", duplicates)
		}
	}
	return ntkNew
}

/*
*Function*************************************************************

	Synopsis    [Rebuilds GIA from mini AIG.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeBuildFromMiniInt(pMan *Hop_Man_t, vAig *Vec_Int_t, nLeaves int32) *Hop_Obj_t {
	size := vec.Vec_IntSize(vAig)
	assert(size > 0, "mini AIG is nonempty")
	last := vec.Vec_IntEntry(vAig, size-1)
	assert(last < 2, "mini AIG last literal is a complement bit")
	if size == 1 {
		assert(nLeaves == 0, "constant mini AIG has no leaves")
		return hoppkg.Hop_NotCond(hoppkg.Hop_ManConst0(pMan), last)
	}
	if size == 2 {
		assert(vec.Vec_IntEntry(vAig, 0) == 0 && nLeaves == 1, "single-leaf mini AIG encoding")
		return hoppkg.Hop_NotCond(hoppkg.Hop_IthVar(pMan, 0), last)
	}
	assert(size&1 == 1, "AND mini AIG has an odd number of entries")
	objects := make([]*Hop_Obj_t, nLeaves+(size-1)/2)
	for i := int32(0); i < nLeaves; i++ {
		objects[i] = hoppkg.Hop_IthVar(pMan, i)
	}
	for i := int32(0); i < size-1; i += 2 {
		lit0, lit1 := vec.Vec_IntEntry(vAig, int(i)), vec.Vec_IntEntry(vAig, int(i+1))
		var0, var1 := lit0>>1, lit1>>1
		assert(var0 >= 0 && var0 < int32(len(objects)) && var1 >= 0 && var1 < int32(len(objects)), "mini AIG fanin is already defined")
		fan0 := hoppkg.Hop_NotCond(objects[var0], lit0&1)
		fan1 := hoppkg.Hop_NotCond(objects[var1], lit1&1)
		id := int32(nLeaves) + i/2
		objects[id] = hoppkg.Hop_And(pMan, fan0, fan1)
	}
	return hoppkg.Hop_NotCond(objects[len(objects)-1], last)
}
func Abc_NodeBuildFromMini(pMan *Hop_Man_t, p *If_Man_t, pCut *If_Cut_t, fUseDsd int32) *Hop_Obj_t {
	var delay int32
	if fUseDsd != 0 {
		delay = ifpkg.If_CutDsdBalanceEval(p, pCut, ifpkg.If_ManArray(p))
	} else {
		delay = ifpkg.If_CutSopBalanceEval(p, pCut, ifpkg.If_ManArray(p))
	}
	assert(delay >= 0, "mini AIG balancing succeeded")
	return Abc_NodeBuildFromMiniInt(pMan, ifpkg.If_ManArray(p), ifpkg.If_CutLeaveNum(pCut))
}

/*
*Function*************************************************************

	Synopsis    [Derive one node after FPGA mapping.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeFromIf_rec(pNtkNew *Abc_Ntk_t, pIfMan *If_Man_t, pIfObj *If_Obj_t, vCover *Vec_Int_t) *Abc_Obj_t {
	if cached, ok := ifpkg.If_ObjCopy(pIfObj).(*Abc_Obj_t); ok && cached != nil {
		return cached
	}
	assert(ifpkg.If_ObjIsAnd(pIfObj) != 0, "uncached mapping objects are AND nodes")
	cut := ifpkg.If_ObjCutBest(pIfObj)
	leaves := ifpkg.If_CutLeaves(cut)
	n := ifpkg.If_CutLeaveNum(cut)
	if n < 0 || len(leaves) < int(n) {
		panic("abcIf: invalid best-cut leaf list")
	}
	pars := pIfMan.Parameters()
	if pars.UsesUserSesLib() {
		fanins := make([]*Abc_Obj_t, n)
		for i := int32(0); i < n; i++ {
			fanins[i] = Abc_NodeFromIf_rec(pNtkNew, pIfMan, ifpkg.If_ManObj(pIfMan, leaves[i]), vCover)
		}
		assert(n > 0, "exact mapper has at least one leaf")
		built := Abc_ExactBuildNode((*word)(unsafe.Pointer(ifpkg.If_CutTruth(pIfMan, cut))), n, ifpkg.If_CutArrTimeProfile(pIfMan, cut), (**Abc_Obj_t)(unsafe.Pointer(&fanins[0])), pNtkNew)
		if built != nil {
			ifpkg.If_ObjSetCopy(pIfObj, built)
		}
		return built
	}
	if pars.UserRecLibrary() {
		if pNtkNew.Function != uint8(abc.ABC_FUNC_AIG) {
			panic("abcIf: recursive library requires AIG network")
		}
	}
	if n > 0 && !pars.DelayOptimization() && !pars.DsdBalance() && !pars.UsesTtPerm() && !pars.UsesUserRecLib() && !pars.UsesUserSesLib() && pars.LUTStructure() == "" && pars.GateSize() == 0 {
		ifpkg.If_CutRotatePins(pIfMan, cut)
		leaves = ifpkg.If_CutLeaves(cut)
	}
	// CNF and multi-valued BDD encodings reverse the leaf order in the source.
	reverseLeaves := pars.UsesCnfs() || pars.UsesMv()
	node := abc.Abc_NtkCreateNode(pNtkNew)
	for i := int32(0); i < n; i++ {
		index := i
		if reverseLeaves {
			index = n - i - 1
		}
		leaf := ifpkg.If_ManObj(pIfMan, leaves[index])
		abc.Abc_ObjAddFanin(node, Abc_NodeFromIf_rec(pNtkNew, pIfMan, leaf, vCover))
	}
	abc.Abc_ObjSetLevel(node, abc.Abc_ObjLevelNew(node))
	truthWords := kit.Kit_TruthWordNum(n)
	truth := unsafe.Slice(ifpkg.If_CutTruth(pIfMan, cut), int(truthWords))
	params := pars
	var function any
	if params.TruthMode() {
		if params.DelayOptimization() || params.DsdBalance() {
			mode := int32(0)
			if params.DsdBalance() {
				mode = 1
			}
			function = Abc_NodeBuildFromMini(pNtkNew.FunctionManager.(*hoppkg.Hop_Man_t), pIfMan, cut, mode)
		} else if params.UsesBdds() || params.UsesCnfs() || params.UsesMv() {
			manager := pNtkNew.FunctionManager.(*core.DdManager)
			function = ifTruthToBDD(manager, truth, n, reverseLeaves)
			core.Cudd_Ref(function.(*core.DdNode))
		} else if params.UsesSops() || params.GateSize() > 0 {
			var cubes []uint32
			compl := kit.Kit_TruthIsop(truth, n, &cubes, 1)
			assert(compl == 0 || compl == 1, "ISOP complement is binary")
			cover := vec.Vec_IntAlloc(int32(len(cubes)))
			for _, cube := range cubes {
				vec.Vec_IntPush(cover, int32(cube))
			}
			if len(cubes) == 0 || len(cubes) == 1 && cubes[0] == 0 {
				assert(compl == 0, "constant cover is not complemented")
				if len(cubes) == 0 {
					function = abc.Abc_NtkCreateNodeConst0(pNtkNew)
				} else {
					function = abc.Abc_NtkCreateNodeConst1(pNtkNew)
				}
			} else {
				sop := abc.Abc_SopCreateFromIsop(pNtkNew.FunctionManager.(*mem.Mem_Flex_t), n, cover)
				if compl != 0 {
					sop = abc.Abc_SopComplement(sop)
				}
				function = sop
			}
			vec.Vec_IntFree(cover)
		} else if params.UserRecLibrary() {
			function = Abc_RecToHop3(pNtkNew.FunctionManager.(*hoppkg.Hop_Man_t), pIfMan, cut, pIfObj)
		} else {
			if params.UsesTtPerm() {
				truth = ifTruthPermute(truth, n, cut)
			}
			hopMan := pNtkNew.FunctionManager.(*hoppkg.Hop_Man_t)
			var memory []uint32
			function = kit.Kit_TruthToHop(hopMan, truth, n, &memory)
		}
		if cutComplement := ifpkg.If_CutTruthIsCompl(cut) != 0; cutComplement && !params.DelayOptimization() && !params.DsdBalance() {
			if hop, ok := function.(*hoppkg.Hop_Obj_t); ok {
				function = hoppkg.Hop_Not(hop)
			} else if bdd, ok := function.(*core.DdNode); ok {
				function = core.Cudd_Not(bdd)
			} else if sop, ok := function.(string); ok {
				function = abc.Abc_SopComplement(sop)
			} else if constant, ok := function.(*Abc_Obj_t); ok {
				if abc.Abc_NodeIsConst0(constant) != 0 {
					function = abc.Abc_NtkCreateNodeConst1(pNtkNew)
				} else if abc.Abc_NodeIsConst1(constant) != 0 {
					function = abc.Abc_NtkCreateNodeConst0(pNtkNew)
				}
			}
		}
	} else {
		function = Abc_NodeIfToHop(pNtkNew.FunctionManager.(*hoppkg.Hop_Man_t), pIfMan, pIfObj)
	}
	if obj, ok := function.(*Abc_Obj_t); ok && obj != node {
		ifpkg.If_ObjSetCopy(pIfObj, obj)
		return obj
	}
	abc.Abc_ObjSetData(node, function)
	ifpkg.If_ObjSetCopy(pIfObj, node)
	return node
}

func ifTruthToBDD(dd *core.DdManager, truth []uint32, n int32, reverse bool) *core.DdNode {
	var build func(int32, uint32) *core.DdNode
	build = func(vars int32, base uint32) *core.DdNode {
		if vars == 0 {
			if kit.Kit_TruthHasBit(truth, int32(base)) != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		low := build(vars-1, base)
		high := build(vars-1, base+(1<<uint(vars-1)))
		index := vars - 1
		if reverse {
			index = n - vars
		}
		return core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, int(index)), high, low)
	}
	return build(n, 0)
}

func ifTruthPermute(truth []uint32, n int32, cut *If_Cut_t) []uint32 {
	words := kit.Kit_TruthWordNum(n)
	result := make([]uint32, words)
	for assignment := uint32(0); assignment < uint32(1)<<uint(n); assignment++ {
		source := assignment
		for variable := int32(0); variable < n; variable++ {
			if ifpkg.If_CutLeafBit(cut, variable) != 0 {
				source ^= 1 << uint(variable)
			}
		}
		if kit.Kit_TruthHasBit(truth, int32(source)) != 0 {
			kit.Kit_TruthSetBit(result, int32(assignment))
		}
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Recursively derives the truth table for the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
// Abc_NodeFromIfRec exposes the C helper used by opt/lpk to rebuild one
// mapped node recursively. Keep this as a forwarding API so both callers use
// the same implementation and copy-cache behavior.
func Abc_NodeFromIfRec(pNtkNew *Abc_Ntk_t, pIfMan *If_Man_t, pIfObj *If_Obj_t, vCover *Vec_Int_t) *Abc_Obj_t {
	return Abc_NodeFromIf_rec(pNtkNew, pIfMan, pIfObj, vCover)
}

/*
*Function*************************************************************

	Synopsis    [Recursively derives the truth table for the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeIfToHop_rec(pHopMan *Hop_Man_t, pIfMan *If_Man_t, pIfObj *If_Obj_t, vVisited *Vec_Ptr_t) *Hop_Obj_t {
	cut := ifpkg.If_ObjCutBest(pIfObj)
	if cached, ok := ifpkg.If_CutData(cut).(*hoppkg.Hop_Obj_t); ok && cached != nil {
		return cached
	}
	if ifpkg.If_ObjIsCi(pIfObj) != 0 {
		return nil
	}
	left := Abc_NodeIfToHop_rec(pHopMan, pIfMan, ifpkg.If_ObjFanin0(pIfObj), vVisited)
	right := Abc_NodeIfToHop_rec(pHopMan, pIfMan, ifpkg.If_ObjFanin1(pIfObj), vVisited)
	if left == nil || right == nil {
		return nil
	}
	result := hoppkg.Hop_And(pHopMan, hoppkg.Hop_NotCond(left, ifpkg.If_ObjFaninC0(pIfObj)), hoppkg.Hop_NotCond(right, ifpkg.If_ObjFaninC1(pIfObj)))
	ifpkg.If_CutSetData(cut, result)
	vec.Vec_PtrPush(vVisited, cut)
	return result
}

/*
*Function*************************************************************

	Synopsis    [Recursively derives the truth table for the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeIfToHop2_rec(pHopMan *Hop_Man_t, pIfMan *If_Man_t, pIfObj *If_Obj_t, vVisited *Vec_Ptr_t) *Hop_Obj_t {
	cut := ifpkg.If_ObjCutBest(pIfObj)
	if value := ifpkg.If_CutData(cut); value != nil {
		if result, ok := value.(*hoppkg.Hop_Obj_t); ok {
			return result
		}
		return nil
	}
	vec.Vec_PtrPush(vVisited, cut)
	ifpkg.If_CutSetData(cut, int32(1))
	if ifpkg.If_ObjIsCi(pIfObj) != 0 {
		return nil
	}
	for alternative := pIfObj; alternative != nil; alternative = alternative.Equivalent() {
		left := Abc_NodeIfToHop2_rec(pHopMan, pIfMan, ifpkg.If_ObjFanin0(alternative), vVisited)
		if left == nil {
			continue
		}
		right := Abc_NodeIfToHop2_rec(pHopMan, pIfMan, ifpkg.If_ObjFanin1(alternative), vVisited)
		if right == nil {
			continue
		}
		result := hoppkg.Hop_And(pHopMan,
			hoppkg.Hop_NotCond(left, ifpkg.If_ObjFaninC0(alternative)),
			hoppkg.Hop_NotCond(right, ifpkg.If_ObjFaninC1(alternative)))
		if alternative.Phase() != pIfObj.Phase() {
			result = hoppkg.Hop_Not(result)
		}
		ifpkg.If_CutSetData(cut, result)
		return result
	}
	return nil
}

/*
*Function*************************************************************

	Synopsis    [Derives the truth table for one cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeIfToHop(pHopMan *Hop_Man_t, pIfMan *If_Man_t, pIfObj *If_Obj_t) *Hop_Obj_t {
	cut := ifpkg.If_ObjCutBest(pIfObj)
	leaves := ifpkg.If_CutLeaves(cut)
	n := ifpkg.If_CutLeaveNum(cut)
	if n <= 1 || len(leaves) < int(n) {
		return nil
	}
	for i := int32(0); i < n; i++ {
		leaf := ifpkg.If_ManObj(pIfMan, leaves[i])
		ifpkg.If_CutSetData(ifpkg.If_ObjCutBest(leaf), hoppkg.Hop_IthVar(pHopMan, i))
	}
	visited := vec.Vec_PtrAlloc(16)
	result := Abc_NodeIfToHop2_rec(pHopMan, pIfMan, pIfObj, visited)
	for i := 0; i < vec.Vec_PtrSize(visited); i++ {
		if c, ok := vec.Vec_PtrEntry(visited, i).(*ifpkg.If_Cut_t); ok {
			ifpkg.If_CutSetData(c, nil)
		}
	}
	for i := int32(0); i < n; i++ {
		leaf := ifpkg.If_ManObj(pIfMan, leaves[i])
		ifpkg.If_CutSetData(ifpkg.If_ObjCutBest(leaf), nil)
	}
	vec.Vec_PtrFree(visited)
	if result == nil {
		fmt.Printf("Abc_NodeIfToHop(): Computing local AIG has failed.\n")
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Comparison for two nodes with the flow.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_ObjCompareFlow(a, b *Abc_Obj_t, flows map[*Abc_Obj_t]float32) bool {
	return flows[a] > flows[b]
}

/*
*Function*************************************************************

	Synopsis    [Orders AIG nodes so that nodes from larger cones go first.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFindGoodOrder_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
	if pNode == nil || pNode.Kind != abc.ObjectNode || abc.Abc_AigNodeIsAnd(pNode) == 0 || abc.Abc_NodeIsTravIdCurrent(pNode) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pNode)
	Abc_NtkFindGoodOrder_rec(abc.Abc_ObjFanin0(pNode), vNodes)
	Abc_NtkFindGoodOrder_rec(abc.Abc_ObjFanin1(pNode), vNodes)
	vec.Vec_PtrPush(vNodes, pNode)
}

/*
*Function*************************************************************

	Synopsis    [Orders AIG nodes so that nodes from larger cones go first.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFindGoodOrder(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
	flows := make(map[*Abc_Obj_t]float32, abc.Abc_NtkObjNum(pNtk))
	for _, node := range abc.Abc_AigDfs(pNtk, 1, 0) {
		if abc.Abc_AigNodeIsAnd(node) == 0 {
			continue
		}
		flow := float32(1)
		for i := int32(0); i < abc.Abc_ObjFaninNum(node); i++ {
			fanin := abc.Abc_ObjFanin(node, i)
			fanoutCount := abc.Abc_ObjFanoutNum(fanin)
			if fanoutCount > 0 {
				flow += flows[fanin] / float32(fanoutCount)
			}
		}
		flows[node] = flow
	}
	cos := make([]*Abc_Obj_t, abc.Abc_NtkCoNum(pNtk))
	for i := range cos {
		co := abc.Abc_NtkCo(pNtk, int32(i))
		cos[i] = co
		flows[co] = flows[abc.Abc_ObjFanin0(co)]
	}
	sort.SliceStable(cos, func(i, j int) bool { return Abc_ObjCompareFlow(cos[i], cos[j], flows) })
	abc.Abc_NtkIncrementTravId(pNtk)
	result := vec.Vec_PtrAlloc(100)
	for _, co := range cos {
		Abc_NtkFindGoodOrder_rec(abc.Abc_ObjFanin0(co), result)
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Sets PO drivers.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMarkMux(pDriver *Abc_Obj_t, ppNode1 **Abc_Obj_t, ppNode2 **Abc_Obj_t) {
	*ppNode1, *ppNode2 = nil, nil
	if pDriver == nil || abc.Abc_NodeIsMuxType(pDriver) == 0 {
		return
	}
	var thenEdge, elseEdge abc.Edge
	control := abc.Abc_NodeRecognizeMux(pDriver, &thenEdge, &elseEdge)
	markSkipped := func(node *Abc_Obj_t) {
		mapped, ok := abc.Abc_ObjCopyData(node).(*ifpkg.If_Obj_t)
		if !ok || mapped == nil {
			return
		}
		mapped = ifpkg.If_Regular(mapped)
		if ifpkg.If_ObjIsAnd(mapped) != 0 {
			ifpkg.If_ObjSetSkipCut(mapped, true)
		}
	}
	markSkipped(abc.Abc_ObjFanin0(pDriver))
	markSkipped(abc.Abc_ObjFanin1(pDriver))
	markSkipped(control)
	*ppNode1 = abc.Abc_ObjRegular(control)
	*ppNode2 = abc.Abc_ObjRegular(thenEdge)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
