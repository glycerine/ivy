package fret

import (
	"fmt"
	. "github.com/glycerine/abcgo/aig/hop"
	. "github.com/glycerine/abcgo/base/abc"
	. "github.com/glycerine/abcgo/map/mio"
	. "github.com/glycerine/abcgo/misc/mem"
	. "github.com/glycerine/abcgo/misc/vec"
	"os"
)

/**CFile****************************************************************

  FileName    [fretMain.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Flow-based retiming package.]

  Synopsis    [Main file for retiming package.]

  Author      [Aaron Hurst]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2008.]

  Revision    [$Id: fretMain.c,v 1.00 2008/01/01 00:00:00 ahurst Exp $]

***********************************************************************/

//#include "fretime.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Abc_FlowRetime_AddDummyFanin( Abc_Obj_t * pObj );

// static Abc_Ntk_t* Abc_FlowRetime_MainLoop( );

// static void Abc_FlowRetime_MarkBlocks( Abc_Ntk_t * pNtk );

// static void Abc_FlowRetime_MarkReachable_rec( Abc_Obj_t * pObj, char end );

// static int  Abc_FlowRetime_ImplementCut( Abc_Ntk_t * pNtk );

// static void  Abc_FlowRetime_RemoveLatchBubbles( Abc_Obj_t * pLatch );

// static Abc_Ntk_t* Abc_FlowRetime_NtkDup( Abc_Ntk_t * pNtk );

// static void Abc_FlowRetime_VerifyPathLatencies( Abc_Ntk_t * pNtk );

// static int  Abc_FlowRetime_VerifyPathLatencies_rec( Abc_Obj_t * pObj, int markD );

// static void Abc_FlowRetime_UpdateLags_forw_rec( Abc_Obj_t *pObj );

// static void Abc_FlowRetime_UpdateLags_back_rec( Abc_Obj_t *pObj );

// extern void Abc_NtkMarkCone_rec( Abc_Obj_t * pObj, int fForward );

// void
// print_node3(Abc_Obj_t *pObj);

var pManMR *MinRegMan_t

var fPathError int32

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Performs minimum-register retiming.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_MinReg(ntk *Abc_Ntk_t, verbose, computeInit, guaranteeInit, blockConst, forwardOnly, backwardOnly, maxIterations, maxDelay, fastConservative int32) *Abc_Ntk_t {
	if Abc_NtkHasOnlyLatchBoxes(ntk) == 0 {
		fmt.Fprintln(os.Stdout, "\tERROR: Can not retime with black/white boxes")
		return ntk
	}
	if maxDelay != 0 {
		level := Abc_NtkLevel(ntk)
		if maxDelay < level {
			fmt.Fprintf(os.Stdout, "ERROR: max delay constraint (%d) must be > current max delay (%d)\n", maxDelay, level)
			return ntk
		}
	}

	pManMR = &MinRegMan_t{
		pNtk: ntk, fVerbose: verbose, fComputeInitState: computeInit,
		fGuaranteeInitState: guaranteeInit, fBlockConst: blockConst,
		fForwardOnly: forwardOnly, fBackwardOnly: backwardOnly,
		nMaxIters: maxIterations, maxDelay: maxDelay,
		fConservTimingOnly: fastConservative,
		vNodes:             Vec_PtrAlloc(100), vInitConstraints: Vec_PtrAlloc(2),
	}
	defer func() { pManMR = nil }()

	fretPrintf("Flow-based minimum-register retiming...\n")
	fretPrintf("\tnetlist type = ")
	switch {
	case Abc_NtkIsNetlist(ntk) != 0:
		fretPrintf("netlist/")
	case Abc_NtkIsLogic(ntk) != 0:
		fretPrintf("logic/")
	case Abc_NtkIsStrash(ntk) != 0:
		fretPrintf("strash/")
	default:
		fretPrintf("***unknown***/")
	}
	switch {
	case Abc_NtkHasSop(ntk) != 0:
		fretPrintf("sop\n")
	case Abc_NtkHasBdd(ntk) != 0:
		fretPrintf("bdd\n")
	case Abc_NtkHasAig(ntk) != 0:
		fretPrintf("aig\n")
	case Abc_NtkHasMapping(ntk) != 0:
		fretPrintf("mapped\n")
	default:
		fretPrintf("***unknown***\n")
	}
	fretPrintf("\tinitial reg count = %d\n", Abc_NtkLatchNum(ntk))
	fretPrintf("\tinitial levels = %d\n", Abc_NtkLevel(ntk))

	if verbose != 0 {
		Abc_FlowRetime_PrintInitStateInfo(ntk)
	}
	fretPrintf("\tpushing bubbles out of latch boxes\n")
	for _, latch := range ntk.ObjectsOfKind(ABC_OBJ_LATCH) {
		Abc_FlowRetime_RemoveLatchBubbles(latch)
	}
	if verbose != 0 {
		Abc_FlowRetime_PrintInitStateInfo(ntk)
	}
	for _, latch := range ntk.ObjectsOfKind(ABC_OBJ_LATCH) {
		assert(Abc_ObjFaninNum(latch) == 1, "latch has one fanin")
		assert(Abc_ObjFanoutNum(latch) == 1, "latch has one fanout")
		assert(Abc_ObjFaninC0(latch) == 0, "latch fanin is not complemented")
		bi, bo := Abc_ObjFanin0(latch), Abc_ObjFanout0(latch)
		assert(Abc_ObjIsBi(bi) != 0 && Abc_ObjFaninNum(bi) <= 1, "latch BI has at most one fanin")
		if Abc_ObjFaninNum(bi) == 0 {
			Abc_FlowRetime_AddDummyFanin(bi)
		}
		assert(Abc_ObjIsBo(bo) != 0 && Abc_ObjFaninNum(bo) == 1, "latch BO has one fanin")
		assert(Abc_ObjFaninC0(bo) == 0, "latch BO fanin is not complemented")
	}

	pManMR.nLatches = Abc_NtkLatchNum(ntk)
	pManMR.nNodes = Abc_NtkObjNumMax(ntk) + 1
	pManMR.vSinkDistHist = Vec_IntStart(int(pManMR.nNodes*2 + 10))
	if maxDelay != 0 {
		Abc_FlowRetime_InitTiming(ntk)
	}
	pManMR.vLags = Vec_IntStart(int(pManMR.nNodes))
	pManMR.pDataArray = make([]Flow_Data_t, int(pManMR.nNodes))
	Abc_FlowRetime_ClearFlows(1)

	ntk = Abc_FlowRetime_MainLoop()
	if computeInit == 0 {
		for _, obj := range ntk.Objects() {
			if Abc_ObjIsLatch(obj) != 0 {
				Abc_LatchSetInitDc(obj)
			}
		}
	}
	pManMR.pDataArray = nil
	pManMR.pInitToOrig = nil
	if pManMR.vNodes != nil {
		Vec_PtrFree(pManMR.vNodes)
	}
	if pManMR.vLags != nil {
		Vec_IntFree(pManMR.vLags)
	}
	if pManMR.vSinkDistHist != nil {
		Vec_IntFree(pManMR.vSinkDistHist)
	}
	if maxDelay != 0 {
		Abc_FlowRetime_FreeTiming(ntk)
	}
	for Vec_PtrSize(pManMR.vInitConstraints) > 0 {
		constraint := Vec_PtrPop(pManMR.vInitConstraints)
		if data, ok := constraint.(*InitConstraint_t); ok && data != nil {
			data.vNodes = nil
			data.vLags = nil
		}
	}
	Vec_PtrFree(pManMR.vInitConstraints)
	if Abc_NtkIsStrash(ntk) != 0 {
		Abc_NtkReassignIds(ntk)
		ntk = Abc_FlowRetime_NtkSilentRestrash(ntk, 1)
	}
	fretPrintf("\tfinal reg count = %d\n", Abc_NtkLatchNum(ntk))
	fretPrintf("\tfinal levels = %d\n", Abc_NtkLevel(ntk))
	return ntk
}

func Abc_FlowRetime_MainLoop() *Abc_Ntk_t {
	ntk := pManMR.pNtk
	original := ntk
	pManMR.fIsForward = 1
	pManMR.iteration = 0
	if pManMR.fBackwardOnly == 0 {
		for pManMR.iteration < pManMR.nMaxIters {
			pManMR.subIteration = 0
			fretPrintf("\tforward iteration %d\n", pManMR.iteration)
			last := Abc_NtkLatchNum(ntk)
			Abc_FlowRetime_MarkBlocks(ntk)
			if pManMR.maxDelay != 0 {
				Abc_FlowRetime_ConstrainConserv(ntk)
				for Abc_FlowRetime_RefineConstraints() != 0 {
					pManMR.subIteration++
					Abc_FlowRetime_ClearFlows(0)
				}
			} else {
				Abc_FlowRetime_PushFlows(ntk, 1)
			}
			cut := Abc_FlowRetime_ImplementCut(ntk)
			Abc_FlowRetime_ClearFlows(1)
			pManMR.iteration++
			if cut == last {
				break
			}
		}
	}
	if Abc_NtkIsStrash(ntk) != 0 {
		Abc_NtkReassignIds(ntk)
		ntk = Abc_FlowRetime_NtkSilentRestrash(ntk, 1)
		pManMR.pNtk = ntk
	}
	if pManMR.fComputeInitState != 0 && pManMR.fVerbose != 0 {
		Abc_FlowRetime_PrintInitStateInfo(ntk)
	}

	pManMR.fIsForward = 0
	if pManMR.fForwardOnly == 0 {
		for {
			pManMR.iteration = 0
			if pManMR.fGuaranteeInitState != 0 {
				if ntk != original {
					Abc_NtkDelete(ntk)
				}
				ntk = Abc_FlowRetime_NtkDup(original)
				pManMR.pNtk = ntk
				fretPrintf("\trestoring network. regs = %d\n", Abc_NtkLatchNum(ntk))
			}
			if pManMR.fComputeInitState != 0 {
				Abc_FlowRetime_SetupBackwardInit(ntk)
			}
			for pManMR.iteration < pManMR.nMaxIters {
				pManMR.subIteration = 0
				fretPrintf("\tbackward iteration %d\n", pManMR.iteration)
				last := Abc_NtkLatchNum(ntk)
				Abc_FlowRetime_AddInitBias()
				Abc_FlowRetime_MarkBlocks(ntk)
				if pManMR.maxDelay != 0 {
					Abc_FlowRetime_ConstrainConserv(ntk)
					for Abc_FlowRetime_RefineConstraints() != 0 {
						pManMR.subIteration++
						Abc_FlowRetime_ClearFlows(0)
					}
				} else {
					Abc_FlowRetime_PushFlows(ntk, 1)
				}
				Abc_FlowRetime_RemoveInitBias()
				cut := Abc_FlowRetime_ImplementCut(ntk)
				fretPrintf("\t\tlevels = %d\n", Abc_NtkLevelReverse(ntk))
				Abc_FlowRetime_ClearFlows(1)
				pManMR.iteration++
				if cut == last {
					break
				}
			}
			if pManMR.fComputeInitState == 0 {
				break
			}
			if Abc_FlowRetime_SolveBackwardInit(ntk) != 0 {
				if pManMR.fVerbose != 0 {
					Abc_FlowRetime_PrintInitStateInfo(ntk)
				}
				break
			}
			if pManMR.fGuaranteeInitState == 0 {
				fmt.Fprintln(os.Stdout, "WARNING: no equivalent init state. setting all initial states to don't-cares")
				for _, latch := range ntk.ObjectsOfKind(ABC_OBJ_LATCH) {
					Abc_LatchSetInitDc(latch)
				}
				break
			}
			Abc_FlowRetime_ConstrainInit()
			Abc_NtkDelete(pManMR.pInitNtk)
			pManMR.pInitNtk = nil
		}
	}
	if pManMR.fComputeInitState != 0 && pManMR.pInitNtk != nil {
		Abc_NtkDelete(pManMR.pInitNtk)
	}
	return ntk
}

func Abc_FlowRetime_RemoveLatchBubbles(latch *Abc_Obj_t) {
	ntk := pManMR.pNtk
	bi, bo := Abc_ObjFanin0(latch), Abc_ObjFanout0(latch)
	bubble := int32(0)
	if Abc_ObjFaninC0(bo) != 0 && Abc_ObjFanoutNum(bo) > 0 {
		bubble = 1
		if Abc_LatchIsInit0(latch) != 0 {
			Abc_LatchSetInit1(latch)
		} else if Abc_LatchIsInit1(latch) != 0 {
			Abc_LatchSetInit0(latch)
		}
	}
	wanted := Abc_ObjFaninC0(bi) ^ bubble ^ Abc_ObjFaninC0(latch)
	if (Abc_ObjFaninC0(bi) != 0) != (wanted != 0) {
		Abc_ObjXorFaninC(bi, 0)
	}
	if Abc_NtkIsStrash(ntk) == 0 && Abc_ObjFaninC0(bi) != 0 {
		Abc_ObjXorFaninC(bi, 0)
		inv := Abc_NtkCreateNodeInv(ntk, Abc_ObjFanin0(bi))
		Abc_ObjPatchFanin(bi, Abc_ObjFanin0(bi), inv)
	}
	if Abc_ObjFaninC0(bo) != 0 {
		Abc_ObjXorFaninC(bo, 0)
	}
	if Abc_ObjFaninC0(latch) != 0 {
		Abc_ObjXorFaninC(latch, 0)
	}
}

func Abc_FlowRetime_MarkBlocks(ntk *Abc_Ntk_t) {
	forward := pManMR.fIsForward != 0
	if forward {
		for _, obj := range ntk.ObjectsOfKind(ABC_OBJ_PO) {
			obj.MarkA = true
		}
	} else {
		for _, obj := range ntk.ObjectsOfKind(ABC_OBJ_PI) {
			obj.MarkA = true
		}
	}
	for _, latch := range ntk.ObjectsOfKind(ABC_OBJ_LATCH) {
		latch.MarkA = true
	}
	var visit func(*Abc_Obj_t)
	visit = func(obj *Abc_Obj_t) {
		if obj == nil || obj.MarkA {
			return
		}
		obj.MarkA = true
		if forward {
			for _, next := range obj.Fanouts() {
				visit(next)
			}
		} else {
			for _, edge := range obj.Fanins() {
				visit(edge.Object)
			}
		}
	}
	if forward {
		for _, pi := range ntk.ObjectsOfKind(ABC_OBJ_PI) {
			pi.MarkA = false
			visit(pi)
		}
	} else {
		for _, po := range ntk.ObjectsOfKind(ABC_OBJ_PO) {
			po.MarkA = false
			visit(po)
		}
	}
	if !forward && pManMR.fBlockConst != 0 {
		for _, obj := range ntk.Objects() {
			constant := Abc_AigNodeIsConst(obj) != 0
			if Abc_NtkIsStrash(ntk) == 0 {
				constant = Abc_NodeIsConst(obj) != 0
			}
			if constant {
				FSET(obj, BLOCK)
			}
		}
	}
	for _, obj := range ntk.Objects() {
		if obj.MarkA {
			obj.MarkA = false
			if Abc_ObjIsLatch(obj) == 0 {
				FSET(obj, BLOCK)
			}
		}
	}
}

func Abc_FlowRetime_PushFlows(ntk *Abc_Ntk_t, verbose int32) int32 {
	pManMR.constraintMask |= BLOCK
	pManMR.fSinkDistTerminate = 0
	dfsfast_preorder(ntk)
	var flow int32
	srcDist := int32(0)
	latches := ntk.ObjectsOfKind(ABC_OBJ_LATCH)
	for pManMR.fSinkDistTerminate == 0 && srcDist < MAX_DIST {
		srcDist = MAX_DIST
		for _, latch := range latches {
			if distance := flowData(latch).e_dist; distance > 0 && distance < srcDist {
				srcDist = distance
			}
		}
		for _, latch := range latches {
			if flowData(latch).e_dist == srcDist && dfsfast_e(latch, nil) != 0 {
				flow++
			}
		}
	}
	if verbose != 0 {
		fretPrintf("\t\tmax-flow1 = %d \t", flow)
	}
	for {
		last := flow
		for _, latch := range latches {
			if dfsplain_e(latch, nil) != 0 {
				flow++
				for _, obj := range ntk.Objects() {
					FUNSET(obj, VISITED)
				}
			}
		}
		if flow <= last {
			break
		}
	}
	if verbose != 0 {
		fretPrintf("max-flow2 = %d\n", flow)
	}
	return flow
}

func Abc_FlowRetime_FixLatchBoxes(ntk *Abc_Ntk_t, boxInputs *Vec_Ptr_t) {
	freeBIs := make([]*Abc_Obj_t, 0, Vec_PtrSize(boxInputs))
	freeBOs := make([]*Abc_Obj_t, 0, Vec_PtrSize(boxInputs))
	for Vec_PtrSize(boxInputs) > 0 {
		bi, ok := Vec_PtrPop(boxInputs).(*Abc_Obj_t)
		assert(ok && bi != nil && Abc_ObjIsBi(bi) != 0, "box input list contains BI objects")
		assert(Abc_ObjFanoutNum(bi) == 1, "BI has one fanout")
		bo := Abc_ObjFanout0(bi)
		assert(Abc_ObjFaninC0(bo) == 0, "BO fanin is regular")
		if Abc_ObjIsBo(bo) != 0 {
			Abc_ObjRemoveFanins(bo)
			fanin := Abc_ObjFanin0(bi)
			Abc_ObjBetterTransferFanout(bo, fanin, Abc_ObjFaninC0(bi))
			Abc_ObjRemoveFanins(bi)
			freeBIs = append(freeBIs, bi)
			freeBOs = append(freeBOs, bo)
		} else if Abc_ObjIsLatch(bo) == 0 {
			panic("fret: BI fanout is neither BO nor latch")
		}
	}
	pop := func(items *[]*Abc_Obj_t, kind Abc_ObjType_t) *Abc_Obj_t {
		if len(*items) == 0 {
			panic("fret: no reusable box boundary object")
		}
		last := len(*items) - 1
		obj := (*items)[last]
		*items = (*items)[:last]
		assert(obj.Kind == kind, "reused box boundary has expected kind")
		return obj
	}
	for _, latch := range ntk.ObjectsOfKind(ABC_OBJ_LATCH) {
		assert(Abc_ObjFaninNum(latch) == 1, "latch has one fanin")
		var bo *Abc_Obj_t
		if Abc_ObjFanoutNum(latch) != 0 {
			bo = Abc_ObjFanout0(latch)
		}
		if bo == nil || Abc_ObjIsBo(bo) == 0 {
			newBO := pop(&freeBOs, ABC_OBJ_BO)
			if bo != nil {
				Abc_ObjTransferFanout(latch, newBO)
			}
			if err := ntk.AddFanin(newBO, latch.Signal()); err != nil {
				panic(err)
			}
		}
		bi := Abc_ObjFanin0(latch)
		if Abc_ObjIsBi(bi) == 0 {
			newBI := pop(&freeBIs, ABC_OBJ_BI)
			oldInput := latch.Fanins()[0]
			if err := ntk.AddFanin(newBI, oldInput); err != nil {
				panic(err)
			}
			Abc_ObjRemoveFanins(latch)
			if err := ntk.AddFanin(latch, newBI.Signal()); err != nil {
				panic(err)
			}
		}
	}
	for _, obj := range freeBIs {
		Abc_NtkDeleteObj(obj)
	}
	for _, obj := range freeBOs {
		Abc_NtkDeleteObj(obj)
	}
	for _, obj := range ntk.Objects() {
		if Abc_ObjIsBo(obj) != 0 {
			assert(Abc_ObjFaninNum(obj) == 1 && Abc_ObjIsLatch(Abc_ObjFanin0(obj)) != 0, "BO connects from latch")
		}
		if Abc_ObjIsBi(obj) != 0 {
			assert(Abc_ObjFaninNum(obj) == 1 && Abc_ObjFanoutNum(obj) == 1 && Abc_ObjIsLatch(Abc_ObjFanout0(obj)) != 0, "BI connects to latch")
		}
		if Abc_ObjIsLatch(obj) != 0 {
			assert(Abc_ObjFanoutNum(obj) == 1 && Abc_ObjFaninNum(obj) == 1, "latch has a BI and BO")
		}
	}
}

func Abc_FlowRetime_VerifyPathLatencies(ntk *Abc_Ntk_t) {
	fPathError = 0
	fretPrintf("\t\tVerifying latency along all paths...")
	for _, obj := range ntk.Objects() {
		if Abc_ObjIsBo(obj) != 0 || (pManMR.fIsForward == 0 && Abc_ObjIsPi(obj) != 0) {
			Abc_FlowRetime_VerifyPathLatencies_rec(obj, 0)
		}
		if fPathError != 0 {
			if Abc_ObjFaninNum(obj) > 0 {
				fmt.Fprint(os.Stdout, "fanin ")
				print_node(Abc_ObjFanin0(obj))
			}
			fmt.Fprintln(os.Stdout)
			panic("fret: path latency verification failed")
		}
	}
	fretPrintf(" ok\n")
	for _, obj := range ntk.Objects() {
		obj.MarkA, obj.MarkB, obj.MarkC = false, false, false
	}
}

func Abc_FlowRetime_VerifyPathLatencies_rec(obj *Abc_Obj_t, markD int32) int32 {
	care := false
	markC := obj.MarkC
	if !obj.MarkB {
		obj.MarkB = true
		if Abc_ObjIsLatch(obj) != 0 {
			markC = true
		}
		if pManMR.fIsForward == 0 && Abc_ObjIsPo(obj) == 0 && Abc_ObjFanoutNum(obj) == 0 {
			return -1
		}
		for _, next := range obj.Fanouts() {
			if Abc_ObjIsBo(next) != 0 || (pManMR.fIsForward != 0 && Abc_ObjIsPo(next) != 0) {
				if markD == 0 && Abc_ObjIsLatch(obj) == 0 {
					fmt.Fprintln(os.Stdout, "\nERROR: no-latch path (end)")
					print_node(next)
					fmt.Fprintln(os.Stdout)
					fPathError = 1
				}
			} else if pManMR.fIsForward == 0 && Abc_ObjIsPo(next) != 0 {
				if markD != 0 || Abc_ObjIsLatch(obj) != 0 {
					fmt.Fprintln(os.Stdout, "\nERROR: extra-latch path to outputs")
					print_node(next)
					fmt.Fprintln(os.Stdout)
					fPathError = 1
				}
			} else {
				nextCare := Abc_FlowRetime_VerifyPathLatencies_rec(next, boolInt(markD != 0 || Abc_ObjIsLatch(obj) != 0))
				if nextCare >= 0 {
					markC = markC || nextCare != 0
					care = true
				}
			}
			if fPathError != 0 {
				print_node(obj)
				fmt.Fprintln(os.Stdout)
				return 0
			}
		}
	}
	if !care {
		return -1
	}
	if markC && markD != 0 {
		fmt.Fprintln(os.Stdout, "\nERROR: mult-latch path")
		print_node(obj)
		fmt.Fprintln(os.Stdout)
		fPathError = 1
	}
	if !markC && markD == 0 {
		fmt.Fprintln(os.Stdout, "\nERROR: no-latch path (inter)")
		print_node(obj)
		fmt.Fprintln(os.Stdout)
		fPathError = 1
	}
	obj.MarkC = markC
	return boolInt(markC)
}

func boolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func Abc_FlowRetime_CopyInitState(src, dest *Abc_Obj_t) {
	if pManMR.fComputeInitState == 0 {
		return
	}
	assert(Abc_ObjIsLatch(src) != 0, "init-state source is latch")
	assert(Abc_ObjFanin0(dest) == src && Abc_ObjFaninC0(dest) == 0, "destination BO is driven directly by latch")
	FUNSET(dest, INIT_CARE)
	if Abc_LatchIsInit0(src) != 0 {
		FSET(dest, INIT_0)
	} else if Abc_LatchIsInit1(src) != 0 {
		FSET(dest, INIT_1)
	}
	if pManMR.fIsForward == 0 {
		initPI, ok := Abc_ObjData(src).(*Abc_Obj_t)
		assert(ok && initPI != nil && Abc_ObjIsPi(initPI) != 0, "backward init mapping points to PI")
		flowData(dest).pInitObj = initPI
	}
}

func Abc_FlowRetime_ImplementCut(ntk *Abc_Ntk_t) int32 {
	var cut, unmoved int32
	freeRegisters := make([]*Abc_Obj_t, 0, Abc_NtkLatchNum(ntk))
	boxInputs := Vec_PtrAlloc(int(Abc_NtkLatchNum(ntk)))
	for _, latch := range ntk.ObjectsOfKind(ABC_OBJ_LATCH) {
		bo, bi := Abc_ObjFanout0(latch), Abc_ObjFanin0(latch)
		assert(Abc_ObjIsBo(bo) != 0 && Abc_ObjIsBi(bi) != 0, "latch has BO and BI boundaries")
		Vec_PtrPush(boxInputs, bi)
		Abc_FlowRetime_CopyInitState(latch, bo)
		freeRegisters = append(freeRegisters, latch)
		FSET(bo, CROSS_BOUNDARY)
		Abc_ObjPatchFanin(bo, latch, bi)
		Abc_ObjRemoveFanins(latch)
	}

	for _, obj := range ntk.Objects() {
		if Abc_ObjIsLatch(obj) != 0 || FTEST(obj, BIAS_NODE) != 0 {
			continue
		}
		assert(pManMR.fIsForward == 0 || FTEST(obj, VISITED_E) == 0 || FTEST(obj, VISITED_R) != 0, "forward min-cut has no reverse-only reachable object")
		if FTEST(obj, VISITED_R) == 0 || FTEST(obj, VISITED_E) != 0 {
			continue
		}
		assert(FTEST(obj, FLOW) != 0, "every cut object belongs to a max-flow path")
		cut++
		if (pManMR.fIsForward != 0 && Abc_ObjIsBo(obj) != 0) || (pManMR.fIsForward == 0 && Abc_ObjIsBi(obj) != 0) {
			unmoved++
		}
		toMove := make([]*Abc_Obj_t, 0, Abc_ObjFanoutNum(obj))
		for _, next := range obj.Fanouts() {
			if Abc_FlowRetime_IsAcrossCut(obj, next) != 0 {
				toMove = append(toMove, next)
			}
		}
		assert(len(toMove) > 0, "cut object has a fanout to move")
		assert(len(freeRegisters) > 0, "a latch is available for every cut object")
		last := len(freeRegisters) - 1
		reg := freeRegisters[last]
		freeRegisters = freeRegisters[:last]
		Abc_ObjAddFanin(reg, obj)
		for _, next := range toMove {
			Abc_ObjPatchFanin(next, obj, reg)
			if Abc_ObjIsBi(next) != 0 {
				assert(Abc_ObjFaninNum(next) == 1, "BI retains one fanin")
			}
		}
	}

	Abc_FlowRetime_VerifyPathLatencies(ntk)
	for _, reg := range freeRegisters {
		Abc_NtkDeleteObj(reg)
	}
	Abc_FlowRetime_UpdateLags()
	Abc_FlowRetime_InitState(ntk)
	Abc_FlowRetime_FixLatchBoxes(ntk, boxInputs)
	Vec_PtrFree(boxInputs)
	fretPrintf("\t\tmin-cut = %d (unmoved = %d)\n", cut, unmoved)
	return cut
}

func Abc_FlowRetime_AddDummyFanin(obj *Abc_Obj_t) {
	ntk := Abc_ObjNtk(obj)
	if Abc_NtkIsStrash(ntk) != 0 {
		Abc_ObjAddFanin(obj, Abc_AigConst1(ntk))
	} else {
		Abc_ObjAddFanin(obj, Abc_NtkCreateNodeConst0(ntk))
	}
}

func flowMarks(obj *Abc_Obj_t) string {
	marks := ""
	if obj.MarkA {
		marks += "A"
	}
	if obj.MarkB {
		marks += "B"
	}
	if obj.MarkC {
		marks += "C"
	}
	return marks
}

func print_node(obj *Abc_Obj_t) {
	fmt.Fprintf(os.Stdout, "node %d type=%d lev=%d tedge=%d (%x%s) fanouts {", Abc_ObjId(obj), Abc_ObjType(obj), obj.Level, Vec_PtrSize(FTIMEEDGES(obj)), flowData(obj).mark, flowMarks(obj))
	for _, next := range obj.Fanouts() {
		fmt.Fprintf(os.Stdout, "%d[%d](%d),", Abc_ObjId(next), Abc_ObjType(next), flowData(next).mark)
	}
	fmt.Fprint(os.Stdout, "} fanins {")
	for _, edge := range obj.Fanins() {
		next := edge.Object
		fmt.Fprintf(os.Stdout, "%d[%d](%d),", Abc_ObjId(next), Abc_ObjType(next), flowData(next).mark)
	}
	fmt.Fprintln(os.Stdout, "}")
}

func print_node2(obj *Abc_Obj_t) {
	fmt.Fprintf(os.Stdout, "node %d type=%d %s fanouts {", Abc_ObjId(obj), Abc_ObjType(obj), flowMarks(obj))
	for _, next := range obj.Fanouts() {
		fmt.Fprintf(os.Stdout, "%d ,", Abc_ObjId(next))
	}
	fmt.Fprint(os.Stdout, "} fanins {")
	for _, edge := range obj.Fanins() {
		fmt.Fprintf(os.Stdout, "%d ,", Abc_ObjId(edge.Object))
	}
	fmt.Fprint(os.Stdout, "} ")
}

func print_node3(obj *Abc_Obj_t) {
	fmt.Fprintf(os.Stdout, "\nnode %d type=%d mark=%d %s\nfanouts\n", Abc_ObjId(obj), Abc_ObjType(obj), flowData(obj).mark, flowMarks(obj))
	for _, next := range obj.Fanouts() {
		print_node(next)
		fmt.Fprintln(os.Stdout)
	}
	fmt.Fprintln(os.Stdout, "fanins")
	for _, edge := range obj.Fanins() {
		print_node(edge.Object)
		fmt.Fprintln(os.Stdout)
	}
}

func Abc_ObjBetterTransferFanout(from, to *Abc_Obj_t, complement int32) {
	for Abc_ObjFanoutNum(from) > 0 {
		fanout := Abc_ObjFanout0(from)
		Abc_ObjPatchFanin(fanout, from, Abc_ObjNotCond(to, complement))
	}
}

func Abc_FlowRetime_IsAcrossCut(obj, next *Abc_Obj_t) int32 {
	if FTEST(obj, VISITED_R) == 0 || FTEST(obj, VISITED_E) != 0 {
		return 0
	}
	if pManMR.fIsForward != 0 {
		if FTEST(next, VISITED_R) == 0 || FTEST(next, BLOCK_OR_CONS)&pManMR.constraintMask != 0 || FTEST(next, CROSS_BOUNDARY) != 0 || Abc_ObjIsLatch(next) != 0 {
			return 1
		}
	} else if FTEST(next, VISITED_E) != 0 || FTEST(next, CROSS_BOUNDARY) != 0 {
		return 1
	}
	return 0
}

func Abc_FlowRetime_ClearFlows(fClearAll int32) {
	if pManMR == nil {
		return
	}
	n := int(pManMR.nNodes)
	if n > len(pManMR.pDataArray) {
		n = len(pManMR.pDataArray)
	}
	if fClearAll != 0 {
		clear(pManMR.pDataArray[:n])
		return
	}
	// Keep the init-state mapping and all flags outside the flow search state.
	for i := 0; i < n; i++ {
		data := &pManMR.pDataArray[i]
		data.mark &^= VISITED | FLOW
		data.e_dist = 0
		data.r_dist = 0
		data.pred = nil
	}
}

/*
*Function*************************************************************

	Synopsis    [Duplicates network.]

	Description [Duplicates any type of network. Preserves copy data.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_NtkDup(ntk *Abc_Ntk_t) *Abc_Ntk_t {
	copy := Abc_NtkDup(ntk)
	originalObjects, copiedObjects := ntk.Objects(), copy.Objects()
	assert(len(originalObjects) == len(copiedObjects), "duplicate preserves object count")
	for i, original := range originalObjects {
		flowData(original).pCopy = copiedObjects[i]
		flowData(original).mark = 0
	}
	return copy
}

func Abc_FlowRetime_NtkSilentRestrash(ntk *Abc_Ntk_t, cleanup int32) *Abc_Ntk_t {
	assert(Abc_NtkIsStrash(ntk) != 0, "silent restrash requires AIG network")
	result := Abc_NtkRestrashWithLatches(ntk, Abc_NtkLatchNum(ntk))
	if result == nil || cleanup == 0 {
		return result
	}
	if manager, ok := result.FunctionManager.(*Abc_Aig_t); ok {
		Abc_AigCleanup(manager)
	}
	if Abc_NtkCheck(result) == 0 {
		Abc_NtkDelete(result)
		return nil
	}
	return result
}

func Abc_FlowRetime_UpdateLags_forw_rec(obj *Abc_Obj_t) {
	assert(Abc_ObjIsPi(obj) == 0 && Abc_ObjIsLatch(obj) == 0, "forward lag recursion starts inside combinational cone")
	if Abc_ObjIsBo(obj) != 0 || Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(obj)
	if Abc_ObjIsNode(obj) != 0 {
		Abc_FlowRetime_SetLag(obj, Abc_FlowRetime_GetLag(obj)-1)
	}
	for _, edge := range obj.Fanins() {
		Abc_FlowRetime_UpdateLags_forw_rec(edge.Object)
	}
}

func Abc_FlowRetime_UpdateLags_back_rec(obj *Abc_Obj_t) {
	assert(Abc_ObjIsPo(obj) == 0 && Abc_ObjIsLatch(obj) == 0, "backward lag recursion starts inside combinational cone")
	if Abc_ObjIsBo(obj) != 0 || Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(obj)
	if Abc_ObjIsNode(obj) != 0 {
		Abc_FlowRetime_SetLag(obj, Abc_FlowRetime_GetLag(obj)+1)
	}
	for _, next := range obj.Fanouts() {
		Abc_FlowRetime_UpdateLags_back_rec(next)
	}
}

func Abc_FlowRetime_UpdateLags() {
	Abc_NtkIncrementTravId(pManMR.pNtk)
	for _, latch := range pManMR.pNtk.ObjectsOfKind(ABC_OBJ_LATCH) {
		if pManMR.fIsForward != 0 {
			for _, edge := range latch.Fanins() {
				Abc_FlowRetime_UpdateLags_forw_rec(edge.Object)
			}
		} else {
			for _, next := range latch.Fanouts() {
				Abc_FlowRetime_UpdateLags_back_rec(next)
			}
		}
	}
}

func Abc_FlowRetime_GetLag(pObj *Abc_Obj_t) int32 {
	assert(!Abc_ObjIsLatch(pObj), "!Abc_ObjIsLatch(pObj)")
	assert(int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags), "int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags)")

	return Vec_IntEntry(pManMR.vLags, Abc_ObjId(pObj))
}

/*
*Function*************************************************************

	Synopsis    [Sets lag value of a node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_SetLag(pObj *Abc_Obj_t, lag int32) {
	assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)")
	assert(int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags), "int32(Abc_ObjId(pObj)) < Vec_IntSize(pManMR.vLags)")

	Vec_IntWriteEntry(pManMR.vLags, Abc_ObjId(pObj), lag)
}

func Abc_ObjPrintNeighborhood_rec(pObj *Abc_Obj_t, vNodes *Vec_Ptr_t, depth int32) {
	if pObj == nil || pObj.MarkC || depth < 0 {
		return
	}
	pObj.MarkC = true
	Vec_PtrPush(vNodes, pObj)
	Abc_ObjPrint(os.Stdout, pObj)
	for _, next := range pObj.Fanouts() {
		Abc_ObjPrintNeighborhood_rec(next, vNodes, depth-1)
	}
	for _, edge := range pObj.Fanins() {
		Abc_ObjPrintNeighborhood_rec(edge.Object, vNodes, depth-1)
	}
}

func Abc_ObjPrintNeighborhood(pObj *Abc_Obj_t, depth int32) {
	var vNodes *Vec_Ptr_t = Vec_PtrAlloc(100)
	Abc_ObjPrintNeighborhood_rec(pObj, vNodes, depth)

	for Vec_PtrSize(vNodes) > 0 {
		pObj2 := Vec_PtrPop(vNodes).(*Abc_Obj_t)
		pObj2.MarkC = false
	}

	Vec_PtrFree(vNodes)
}

//ABC_NAMESPACE_IMPL_END
