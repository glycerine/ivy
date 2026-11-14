package fret

import (
	"fmt"
	"os"
	"strings"

	hop "github.com/glycerine/abcgo/aig/hop"
	abc "github.com/glycerine/abcgo/base/abc"
	core "github.com/glycerine/abcgo/bdd/cudd/core"
	mio "github.com/glycerine/abcgo/map/mio"
	vec "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [fretInit.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Flow-based retiming package.]

  Synopsis    [Initialization for retiming package.]

  Author      [Aaron Hurst]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2008.]

  Revision    [$Id: fretInit.c,v 1.00 2008/01/01 00:00:00 ahurst Exp $]

***********************************************************************/

//#include "fretime.h"

//#include "base/io/ioAbc.h"
//#include "map/mio/mio.h"
//#include "aig/hop/hop.h"

// Optional ABC_USE_CUDD support is omitted in the portable Go port.

//ABC_NAMESPACE_IMPL_START

// DEBUG_PRINT_INIT_NTK is disabled.

////////////////////////////////////////////////////////////////////////
///                     FUNCTION PROTOTYPES                          ///
////////////////////////////////////////////////////////////////////////

// static void Abc_FlowRetime_UpdateForwardInit_rec( abc.Abc_Obj_t * pObj );

// static void Abc_FlowRetime_VerifyBackwardInit( abc.Abc_Ntk_t * pNtk );

// static void Abc_FlowRetime_VerifyBackwardInit_rec( abc.Abc_Obj_t * pObj );

// static abc.Abc_Obj_t* Abc_FlowRetime_UpdateBackwardInit_rec( abc.Abc_Obj_t *pOrigObj );

// static void Abc_FlowRetime_SimulateNode( abc.Abc_Obj_t * pObj );

// static void Abc_FlowRetime_SimulateSop( abc.Abc_Obj_t * pObj, char *pSop );

// static void Abc_FlowRetime_SetInitToOrig( abc.Abc_Obj_t *pInit, abc.Abc_Obj_t *pOrig );

// static void Abc_FlowRetime_GetInitToOrig( abc.Abc_Obj_t *pInit, abc.Abc_Obj_t **pOrig, int *lag );

// static void Abc_FlowRetime_ClearInitToOrig( abc.Abc_Obj_t *pInit );

// extern void * abc.Abc_FrameReadLibGen();

// extern void abc.Abc_NtkMarkCone_rec( abc.Abc_Obj_t * pObj, int fForward );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Updates initial state information.]

	Description [Assumes latch boxes in original position, latches in
	             new positions.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_InitState(pNtk *abc.Abc_Ntk_t) {

	if pManMR.fComputeInitState == 0 {
		return
	}

	if pManMR.fIsForward != 0 {
		Abc_FlowRetime_UpdateForwardInit(pNtk)
	} else {
		Abc_FlowRetime_UpdateBackwardInit(pNtk)
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints initial state information.]

	Description [Prints distribution of 0,1,and X initial states.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_ObjFirstNonLatchBox(pOrigObj *abc.Abc_Obj_t, pResult **abc.Abc_Obj_t) int32 {
	if pOrigObj == nil || pResult == nil {
		panic("nil object in first non-latch lookup")
	}
	lag := int32(0)
	*pResult = pOrigObj
	ntk := abc.Abc_ObjNtk(pOrigObj)
	abc.Abc_NtkIncrementTravId(ntk)
	for abc.Abc_ObjIsBo(*pResult) != 0 || abc.Abc_ObjIsLatch(*pResult) != 0 || abc.Abc_ObjIsBi(*pResult) != 0 {
		if abc.Abc_ObjFaninNum(*pResult) == 0 {
			panic("box boundary has no fanin")
		}
		*pResult = abc.Abc_ObjFanin0(*pResult)
		if abc.Abc_NodeIsTravIdCurrent(*pResult) != 0 {
			return -1
		}
		abc.Abc_NodeSetTravIdCurrent(*pResult)
		if abc.Abc_ObjIsLatch(*pResult) != 0 {
			lag++
		}
	}
	return lag
}

/*
*Function*************************************************************

	Synopsis    [Prints initial state information.]

	Description [Prints distribution of 0,1,and X initial states.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_PrintInitStateInfo(pNtk *abc.Abc_Ntk_t) {
	var n0, n1, nDC, nOther int
	for _, latch := range pNtk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		switch {
		case abc.Abc_LatchIsInit0(latch) != 0:
			n0++
		case abc.Abc_LatchIsInit1(latch) != 0:
			n1++
		case abc.Abc_LatchIsInitDc(latch) != 0:
			nDC++
		default:
			nOther++
		}
	}
	fmt.Printf("\tinitial states {0,1,x} = {%d, %d, %d}", n0, n1, nDC)
	if nOther != 0 {
		fmt.Printf(" + %d UNKNOWN", nOther)
	}
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Computes initial state after forward retiming.]

	Description [Assumes box outputs in old positions stored w/ init values.
	             Uses three-value simulation to preserve don't cares.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_UpdateForwardInit(pNtk *abc.Abc_Ntk_t) {
	fmt.Fprintln(os.Stdout, "\t\tupdating init state")
	abc.Abc_NtkIncrementTravId(pNtk)
	for _, latch := range pNtk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		fanin := abc.Abc_ObjFanin0(latch)
		Abc_FlowRetime_UpdateForwardInit_rec(fanin)
		switch {
		case FTEST(fanin, INIT_0) != 0:
			abc.Abc_LatchSetInit0(latch)
		case FTEST(fanin, INIT_1) != 0:
			abc.Abc_LatchSetInit1(latch)
		default:
			abc.Abc_LatchSetInitDc(latch)
		}
	}
}

func Abc_FlowRetime_UpdateForwardInit_rec(pObj *abc.Abc_Obj_t) {
	assert(abc.Abc_ObjIsPi(pObj) == 0, "forward init recursion does not reach primary inputs")
	if abc.Abc_ObjIsBo(pObj) != 0 || abc.Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pObj)
	for _, edge := range pObj.Fanins() {
		Abc_FlowRetime_UpdateForwardInit_rec(edge.Object)
	}
	Abc_FlowRetime_SimulateNode(pObj)
}

/*
*Function*************************************************************

	Synopsis    [Recursively evaluates HOP netlist.]

	Description [Exponential.  There aren't enough flags on a HOP node.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func evalHopInit(obj *hop.Hop_Obj_t, primary map[*hop.Hop_Obj_t][2]int32) (int32, int32) {
	regular := hop.Hop_Regular(obj)
	complement := int32(hop.Hop_IsComplement(obj))
	switch {
	case hop.Hop_ObjIsConst1(regular) != 0:
		return 1 ^ complement, 0
	case hop.Hop_ObjIsPi(regular) != 0:
		state, ok := primary[regular]
		if !ok {
			panic("HOP PI has no initial-state value")
		}
		return state[0] ^ complement, 1 - state[1]
	case hop.Hop_ObjIsPo(regular) != 0:
		return evalHopInit(hop.Hop_ObjChild0(regular), primary)
	case hop.Hop_ObjIsAnd(regular) != 0:
		v0, d0 := evalHopInit(hop.Hop_ObjChild0(regular), primary)
		v1, d1 := evalHopInit(hop.Hop_ObjChild1(regular), primary)
		dc := (d0 & v1) | (d1 & v0) | (d0 & d1)
		return (v0 & v1) ^ complement, dc
	default:
		panic("unsupported HOP node in initial-state simulation")
	}
}

/*
*Function*************************************************************

	Synopsis    [Sets initial value flags.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_SetInitValue(pObj *abc.Abc_Obj_t, val int32, dc int32) {

	// store init value
	FUNSET(pObj, INIT_CARE)
	if dc == 0 {
		if val != 0 {
			FSET(pObj, INIT_1)
		} else {
			FSET(pObj, INIT_0)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Propogates initial state through a logic node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_SimulateNode(pObj *abc.Abc_Obj_t) {
	ntk := abc.Abc_ObjNtk(pObj)
	assert(abc.Abc_ObjIsLatch(pObj) == 0, "logic simulation is not applied to latches")
	if abc.Abc_NtkIsStrash(ntk) != 0 && abc.Abc_AigNodeIsConst(pObj) != 0 {
		Abc_FlowRetime_SetInitValue(pObj, 1, 0)
		return
	}
	if abc.Abc_NtkIsStrash(ntk) == 0 && abc.Abc_ObjIsNode(pObj) != 0 {
		if abc.Abc_NodeIsConst0(pObj) != 0 {
			Abc_FlowRetime_SetInitValue(pObj, 0, 0)
			return
		}
		if abc.Abc_NodeIsConst1(pObj) != 0 {
			Abc_FlowRetime_SetInitValue(pObj, 1, 0)
			return
		}
	}
	if abc.Abc_ObjIsNode(pObj) == 0 {
		fanin := abc.Abc_ObjFanin0(pObj)
		value := FTEST(fanin, INIT_1) & 1
		if abc.Abc_ObjFaninC0(pObj) != 0 {
			value ^= 1
		}
		care := int32(0)
		if FTEST(fanin, INIT_CARE) == 0 {
			care = 1
		}
		Abc_FlowRetime_SetInitValue(pObj, value, care)
		return
	}
	if abc.Abc_NtkHasSop(ntk) != 0 {
		sop, ok := abc.Abc_ObjData(pObj).(string)
		assert(ok, "SOP node stores a string")
		Abc_FlowRetime_SimulateSop(pObj, sop)
		return
	}
	if abc.Abc_NtkHasBdd(ntk) != 0 {
		manager, ok := ntk.FunctionManager.(*core.DdManager)
		assert(ok && manager != nil, "BDD network has a manager")
		function, ok := abc.Abc_ObjData(pObj).(*core.DdNode)
		assert(ok && function != nil, "BDD node has a function")
		owned := false
		for i, edge := range pObj.Fanins() {
			if FTEST(edge.Object, INIT_CARE) == 0 {
				continue
			}
			variable := core.Cudd_bddIthVar(manager, i)
			assert(variable != nil, "BDD manager has a variable for each fanin")
			literal := variable
			if FTEST(edge.Object, INIT_0) != 0 {
				literal = core.Cudd_Not(variable)
			}
			next := core.Cudd_Cofactor(manager, function, literal)
			assert(next != nil, "BDD cofactor is valid")
			if next != function {
				core.Cudd_Ref(next)
				if owned {
					core.Cudd_RecursiveDeref(manager, function)
				}
				function = next
				owned = true
			}
		}
		value, dc := int32(0), int32(1)
		if core.Cudd_IsConstant(function) {
			dc = 0
			if function == core.Cudd_ReadOne(manager) {
				value = 1
			}
		}
		if owned {
			core.Cudd_RecursiveDeref(manager, function)
		}
		Abc_FlowRetime_SetInitValue(pObj, value, dc)
		return
	}
	if abc.Abc_NtkIsStrash(ntk) != 0 {
		value, care := int32(1), int32(1)
		for i, edge := range pObj.Fanins() {
			faninValue := int32(1)
			if FTEST(edge.Object, INIT_0) != 0 {
				faninValue = 0
			}
			if edge.Complement {
				faninValue ^= 1
			}
			value &= faninValue
			if FTEST(edge.Object, INIT_CARE) == 0 {
				care = 0
			}
			if i >= 1 {
				break
			}
		}
		if value == 0 {
			care = 1
		}
		Abc_FlowRetime_SetInitValue(pObj, value, 1-care)
		return
	}
	if abc.Abc_NtkHasAig(ntk) != 0 {
		manager, ok := ntk.FunctionManager.(*hop.Hop_Man_t)
		assert(ok, "AIG node has HOP manager")
		hopNode, ok := abc.Abc_ObjData(pObj).(*hop.Hop_Obj_t)
		assert(ok, "AIG node stores HOP node")
		p := map[*hop.Hop_Obj_t][2]int32{}
		for i, edge := range pObj.Fanins() {
			pi := hop.Hop_ManPi(manager, int32(i))
			val := int32(0)
			if FTEST(edge.Object, INIT_1) != 0 {
				val = 1
			}
			care := int32(0)
			if FTEST(edge.Object, INIT_CARE) != 0 {
				care = 1
			}
			p[pi] = [2]int32{val, care}
		}
		val, dc := evalHopInit(hopNode, p)
		Abc_FlowRetime_SetInitValue(pObj, val, dc)
		return
	}
	if abc.Abc_NtkHasMapping(ntk) != 0 {
		gate, ok := abc.Abc_ObjData(pObj).(*mio.Mio_Gate_t)
		assert(ok, "mapped node stores MIO gate")
		Abc_FlowRetime_SimulateSop(pObj, mio.Mio_GateReadSop(gate))
		return
	}
	panic("unsupported logic representation in initial-state simulation")
}

/*
*Function*************************************************************

	Synopsis    [Propogates initial state through a SOP node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_SimulateSop(pObj *abc.Abc_Obj_t, pSop string) {
	assert(pSop != "" && abc.Abc_SopIsExorType(pSop) == 0, "nonempty non-XOR SOP")
	nVars := int(abc.Abc_SopGetVarNum(pSop))
	assert(nVars == int(abc.Abc_ObjFaninNum(pObj)), "SOP variable count matches fanins")
	rOr, dcOr := int32(0), int32(0)
	for _, cube := range strings.Split(strings.TrimSuffix(pSop, "\n"), "\n") {
		if cube == "" {
			continue
		}
		sep := strings.IndexByte(cube, ' ')
		assert(sep == nVars && sep+1 < len(cube), "SOP cube has expected width")
		rAnd, dcAnd := int32(1), int32(0)
		for j, lit := range cube[:sep] {
			fanin := abc.Abc_ObjFanin(pObj, int32(j))
			want := int32(0)
			switch lit {
			case '0':
				if FTEST(fanin, INIT_0) != 0 {
					want = 1
				}
			case '1':
				if FTEST(fanin, INIT_1) != 0 {
					want = 1
				}
			default:
				continue
			}
			if FTEST(fanin, INIT_CARE) == 0 {
				dcAnd = 1
			} else {
				rAnd &= want
			}
		}
		if rAnd == 0 {
			dcAnd = 0
		}
		if dcAnd != 0 {
			dcOr = 1
		} else {
			rOr |= rAnd
		}
	}
	if rOr != 0 {
		dcOr = 0
	}
	if abc.Abc_SopGetPhase(pSop) == 0 {
		rOr ^= 1
	}
	Abc_FlowRetime_SetInitValue(pObj, rOr, dcOr)
}

/*
*Function*************************************************************

	Synopsis    [Sets up backward initial state computation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_SetupBackwardInit(pNtk *abc.Abc_Ntk_t) {
	ntkType, ntkFunc := abc.Abc_NtkType_t(pNtk.Type), abc.Abc_NtkFunc_t(pNtk.Function)
	if abc.Abc_NtkIsStrash(pNtk) != 0 || abc.Abc_NtkHasMapping(pNtk) != 0 {
		ntkFunc = abc.ABC_FUNC_SOP
	}
	if abc.Abc_NtkIsStrash(pNtk) != 0 {
		ntkType = abc.ABC_NTK_LOGIC
	}
	pManMR.pInitNtk = abc.Abc_NtkAlloc(ntkType, ntkFunc, 1)
	constrained := make([]*abc.Abc_Obj_t, 0, int(abc.Abc_NtkLatchNum(pNtk)))
	for _, latch := range pNtk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		pi := abc.Abc_NtkCreatePi(pManMR.pInitNtk)
		switch {
		case abc.Abc_LatchIsInit0(latch) != 0:
			constrained = append(constrained, abc.Abc_NtkCreateNodeInv(pManMR.pInitNtk, pi))
		case abc.Abc_LatchIsInit1(latch) != 0:
			constrained = append(constrained, pi)
		}
		abc.Abc_ObjSetData(latch, pi)
	}
	if len(constrained) == 0 {
		pManMR.fSolutionIsDc = 1
		return
	}
	pManMR.fSolutionIsDc = 0
	and := abc.Abc_NtkCreateNodeAnd(pManMR.pInitNtk, constrained)
	abc.Abc_ObjAddFanin(abc.Abc_NtkCreatePo(pManMR.pInitNtk), and)
}

/*
*Function*************************************************************

	Synopsis    [Solves backward initial state computation.]

	Description []

	SideEffects [Sets object copies in init ntk.]

	SeeAlso     []

**********************************************************************
*/
func Abc_FlowRetime_SolveBackwardInit(ntk *abc.Abc_Ntk_t) int32 {
	assert(pManMR.pInitNtk != nil, "backward init network exists")
	if pManMR.fSolutionIsDc != 0 {
		for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
			abc.Abc_LatchSetInitDc(latch)
		}
		fmt.Fprintln(os.Stdout, "\tno init state computation: all-don't-care solution")
		return 1
	}
	for _, obj := range pManMR.pInitNtk.Objects() {
		assert(abc.Abc_ObjIsLatch(obj) == 0 && abc.Abc_ObjIsBo(obj) == 0 && abc.Abc_ObjIsBi(obj) == 0, "init network is combinational")
	}
	abc.Abc_NtkAddDummyPoNames(pManMR.pInitNtk)
	abc.Abc_NtkAddDummyPiNames(pManMR.pInitNtk)
	if abc.Abc_NtkHasSop(pManMR.pInitNtk) != 0 && abc.Abc_NtkSopToAig(pManMR.pInitNtk) == 0 {
		fmt.Fprintln(os.Stdout, "FAILURE: could not convert init network to AIG")
		return 0
	}
	fmt.Fprintf(os.Stdout, "\tsolving for init state (%d nodes)... ", abc.Abc_NtkObjNum(pManMR.pInitNtk))
	result := abc.Abc_NtkMiterSat(pManMR.pInitNtk, int64(500000), int64(50000000), 0, nil, nil)
	if result != 0 {
		fmt.Fprintln(os.Stdout, "FAILURE")
		return 0
	}
	fmt.Fprintln(os.Stdout, "SUCCESS")

	pis := pManMR.pInitNtk.ObjectsOfKind(abc.ABC_OBJ_PI)
	model := abc.Abc_NtkReadModel(pManMR.pInitNtk)
	assert(len(model) >= len(pis), "SAT model covers init PIs")
	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		pi, ok := abc.Abc_ObjData(latch).(*abc.Abc_Obj_t)
		assert(ok && abc.Abc_ObjIsPi(pi) != 0, "latch maps to init PI")
		abc.Abc_LatchSetInitNone(latch)
		index := -1
		for i, candidate := range pis {
			if candidate == pi {
				index = i
				break
			}
		}
		assert(index >= 0, "mapped init PI belongs to init network")
		if model[index] > 0 {
			abc.Abc_LatchSetInit1(latch)
		} else if model[index] == 0 {
			abc.Abc_LatchSetInit0(latch)
		} else {
			abc.Abc_LatchSetInitDc(latch)
		}
	}
	return 1
}

func Abc_FlowRetime_UpdateBackwardInit(ntk *abc.Abc_Ntk_t) {
	initNtk := pManMR.pInitNtk
	boxOutputs := make([]*abc.Abc_Obj_t, 0)
	freePis := make([]*abc.Abc_Obj_t, 0)
	for _, bo := range ntk.Objects() {
		if abc.Abc_ObjIsBo(bo) == 0 {
			continue
		}
		initPi := flowData(bo).pInitObj
		assert(initPi != nil && abc.Abc_ObjIsPi(initPi) != 0, "BO init mapping is an init PI")
		replacement := abc.Abc_NtkCreateNodeBuf(initNtk, nil)
		Abc_FlowRetime_ClearInitToOrig(replacement)
		Abc_ObjBetterTransferFanout(initPi, replacement, 0)
		flowData(bo).pInitObj = replacement
		bo.MarkA = true
		boxOutputs = append(boxOutputs, bo)
		freePis = append(freePis, initPi)
	}
	for _, pi := range initNtk.ObjectsOfKind(abc.ABC_OBJ_PI) {
		assert(abc.Abc_ObjFanoutNum(pi) == 0, "unconnected init PI before latch remapping")
	}
	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		assert(len(freePis) > 0, "one free init PI per latch")
		pi := freePis[len(freePis)-1]
		freePis = freePis[:len(freePis)-1]
		latch.MarkA, latch.MarkB = true, true
		flowData(latch).pInitObj = pi
		abc.Abc_ObjSetData(latch, pi)
	}
	for _, bo := range boxOutputs {
		Abc_FlowRetime_UpdateBackwardInit_rec(bo)
	}
	for _, obj := range ntk.Objects() {
		obj.MarkA, obj.MarkB = false, false
	}
}

func Abc_FlowRetime_CopyNodeToInitNtk(original *abc.Abc_Obj_t) *abc.Abc_Obj_t {
	assert(original != nil, "original node exists")
	initNtk := pManMR.pInitNtk
	if abc.Abc_ObjIsBi(original) != 0 || abc.Abc_ObjIsBo(original) != 0 || (abc.Abc_ObjIsNode(original) == 0 && abc.Abc_ObjIsLatch(original) == 0) {
		copy := abc.Abc_NtkCreateNodeBuf(initNtk, nil)
		Abc_FlowRetime_ClearInitToOrig(copy)
		return copy
	}
	if abc.Abc_NtkIsStrash(pManMR.pNtk) != 0 {
		if abc.Abc_AigNodeIsConst(original) != 0 {
			constant := abc.Abc_NtkCreateNodeConst1(initNtk)
			Abc_FlowRetime_ClearInitToOrig(constant)
			return constant
		}
		assert(abc.Abc_ObjIsNode(original) != 0 && abc.Abc_ObjFaninNum(original) == 2, "strashed logic node has two inputs")
		copy := abc.Abc_NtkCreateObj(initNtk, abc.ABC_OBJ_NODE)
		zero, one := byte('1'), byte('1')
		if abc.Abc_ObjFaninC0(original) != 0 {
			zero = '0'
		}
		if abc.Abc_ObjFaninC1(original) != 0 {
			one = '0'
		}
		abc.Abc_ObjSetData(copy, string([]byte{zero, one, ' ', '1', '\n'}))
		return copy
	}
	if abc.Abc_NtkHasMapping(pManMR.pNtk) != 0 {
		gate, ok := abc.Abc_ObjData(original).(*mio.Mio_Gate_t)
		if !ok || gate == nil {
			assert(abc.Abc_ObjFaninNum(original) == 1, "mapped terminal is a buffer")
			copy := abc.Abc_NtkCreateNodeBuf(initNtk, nil)
			Abc_FlowRetime_ClearInitToOrig(copy)
			return copy
		}
		copy := abc.Abc_NtkCreateObj(initNtk, abc.ABC_OBJ_NODE)
		abc.Abc_ObjSetData(copy, mio.Mio_GateReadSop(gate))
		return copy
	}
	copy := abc.Abc_NtkDupObj(initNtk, original, 0)
	copy.Phase = original.Phase
	return copy
}

func Abc_FlowRetime_UpdateBackwardInit_rec(original *abc.Abc_Obj_t) *abc.Abc_Obj_t {
	assert(original != nil, "original node exists")
	assert(abc.Abc_ObjIsPi(original) == 0 && abc.Abc_ObjIsPo(original) == 0, "recursion stays inside the logic network")
	if FTEST(original, BIAS_NODE) != 0 {
		return nil
	}
	initObj := flowData(original).pInitObj
	if !original.MarkA {
		initObj = Abc_FlowRetime_CopyNodeToInitNtk(original)
		Abc_FlowRetime_SetInitToOrig(initObj, original)
		flowData(original).pInitObj = initObj
		original.MarkA = true
	}
	assert(initObj != nil, "original node has init-network copy")
	if !original.MarkB {
		for _, edge := range original.Fanins() {
			assert(abc.Abc_ObjIsBo(edge.Object) == 0, "recursion does not cross BO boundary")
			faninCopy := Abc_FlowRetime_UpdateBackwardInit_rec(edge.Object)
			if faninCopy != nil {
				abc.Abc_ObjAddFanin(initObj, faninCopy)
			}
		}
		original.MarkB = true
	}
	return initObj
}

func Abc_FlowRetime_VerifyBackwardInit(ntk *abc.Abc_Ntk_t) {
	fmt.Fprintln(os.Stdout, "\t\tupdating init state")
	abc.Abc_NtkIncrementTravId(ntk)
	for _, obj := range ntk.ObjectsOfKind(abc.ABC_OBJ_BO) {
		fanin := abc.Abc_ObjFanin0(obj)
		Abc_FlowRetime_VerifyBackwardInit_rec(fanin)
		if FTEST(obj, INIT_CARE) != 0 && (FTEST(obj, INIT_CARE) != FTEST(fanin, INIT_CARE) || FTEST(obj, INIT_1) != FTEST(fanin, INIT_1)) {
			fmt.Fprintf(os.Stdout, "ERROR: expected val=%d care=%d and got val=%d care=%d\n", boolInt(FTEST(obj, INIT_1) != 0), boolInt(FTEST(obj, INIT_CARE) != 0), boolInt(FTEST(fanin, INIT_1) != 0), boolInt(FTEST(fanin, INIT_CARE) != 0))
		}
	}
}

func Abc_FlowRetime_VerifyBackwardInit_rec(obj *abc.Abc_Obj_t) {
	assert(obj != nil && abc.Abc_ObjIsBo(obj) == 0 && abc.Abc_ObjIsPi(obj) == 0, "backward init simulation remains inside the logic cone")
	if abc.Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(obj)
	if abc.Abc_ObjIsLatch(obj) != 0 {
		FUNSET(obj, INIT_CARE)
		if abc.Abc_LatchIsInit0(obj) != 0 {
			FSET(obj, INIT_0)
		} else if abc.Abc_LatchIsInit1(obj) != 0 {
			FSET(obj, INIT_1)
		}
		return
	}
	for _, edge := range obj.Fanins() {
		Abc_FlowRetime_VerifyBackwardInit_rec(edge.Object)
	}
	Abc_FlowRetime_SimulateNode(obj)
}

func Abc_FlowRetime_PartialSat(vNodes []*abc.Abc_Obj_t, cut int32) int32 {
	initNtk := pManMR.pInitNtk
	assert(abc.Abc_NtkPoNum(initNtk) == 1, "init network has one output")
	part := abc.Abc_NtkAlloc(abc.Abc_NtkType_t(initNtk.Type), abc.Abc_NtkFunc_t(initNtk.Function), 0)
	defer abc.Abc_NtkDelete(part)
	copies := make(map[*abc.Abc_Obj_t]*abc.Abc_Obj_t, len(vNodes))
	for _, constant := range initNtk.ObjectsOfKind(abc.ABC_OBJ_CONST1) {
		copies[constant] = abc.Abc_NtkCreateNodeConst1(part)
	}
	for i, obj := range vNodes {
		obj.Level = i
		assert(abc.Abc_ObjIsPo(obj) == 0, "DFS node list excludes POs")
		if int32(i) < cut && !obj.MarkA {
			copies[obj] = abc.Abc_NtkCreatePi(part)
			continue
		}
		copy := abc.Abc_NtkDupObj(part, obj, 0)
		copy.Phase = obj.Phase
		for _, edge := range obj.Fanins() {
			fanin := copies[edge.Object]
			assert(fanin != nil, "fanin appears before its consumer in DFS order")
			if err := part.AddFanin(copy, abc.Edge{Object: fanin, Complement: edge.Complement}); err != nil {
				panic(err)
			}
		}
		copies[obj] = copy
	}
	po := abc.Abc_NtkPo(initNtk, 0)
	assert(abc.Abc_ObjFaninNum(po) == 1, "init output has one driver")
	poEdge := po.Fanins()[0]
	output := copies[poEdge.Object]
	assert(output != nil, "init output driver copied")
	if err := part.AddFanin(abc.Abc_NtkCreatePo(part), abc.Edge{Object: output, Complement: poEdge.Complement}); err != nil {
		panic(err)
	}
	result := abc.Abc_NtkMiterSat(part, int64(500000), int64(50000000), 0, nil, nil)
	return boolInt(result == 0)
}

func Abc_FlowRetime_ConstrainInit() {
	assert(pManMR.pInitNtk != nil, "backward init network exists")
	fmt.Fprintln(os.Stdout, "\tsearch for initial state conflict...")
	dfsNodes := abc.Abc_NtkDfs(pManMR.pInitNtk, 0)
	vNodes := initConflictNodeOrder(dfsNodes, pManMR.pInitNtk.ObjectsOfKind(abc.ABC_OBJ_PI))
	constraint := &InitConstraint_t{vNodes: vec.Vec_IntAlloc(8), vLags: vec.Vec_IntAlloc(8)}
	for {
		assert(len(vNodes) > 0, "init network DFS has a conflict node")
		low, high := int32(0), int32(len(vNodes))
		for low < high-1 {
			middle := (low + high) / 2
			if Abc_FlowRetime_PartialSat(vNodes, middle) != 0 {
				high = middle
			} else {
				low = middle
			}
		}
		selected := vNodes[low]
		markTFO := func(root *abc.Abc_Obj_t) {
			seen := make(map[*abc.Abc_Obj_t]bool)
			var visit func(*abc.Abc_Obj_t)
			visit = func(obj *abc.Abc_Obj_t) {
				if obj == nil || seen[obj] {
					return
				}
				seen[obj] = true
				obj.MarkA = true
				for _, next := range obj.Fanouts() {
					visit(next)
				}
			}
			visit(root)
		}
		markTFO(selected)
		var original *abc.Abc_Obj_t
		var lag int32
		Abc_FlowRetime_GetInitToOrig(selected, &original, &lag)
		assert(original != nil, "conflict node maps to original object")
		vec.Vec_IntPush(constraint.vNodes, abc.Abc_ObjId(original))
		vec.Vec_IntPush(constraint.vLags, lag)
		fmt.Fprintf(os.Stdout, "\t\tconflict term %d maps to %d/%d\n", low, abc.Abc_ObjId(original), lag)
		if Abc_FlowRetime_PartialSat(vNodes, int32(len(vNodes))) == 0 {
			break
		}
	}
	constraint.pBiasNode = nil
	vec.Vec_PtrPush(pManMR.vInitConstraints, constraint)
	for _, obj := range pManMR.pInitNtk.Objects() {
		obj.MarkA = false
	}
}

func Abc_FlowRetime_RemoveInitBias() {
	for i := 0; i < vec.Vec_PtrSize(pManMR.vInitConstraints); i++ {
		constraint, ok := vec.Vec_PtrEntry(pManMR.vInitConstraints, i).(*InitConstraint_t)
		assert(ok && constraint != nil, "init constraint has expected type")
		if constraint.pBiasNode != nil {
			abc.Abc_NtkDeleteObj(constraint.pBiasNode)
			constraint.pBiasNode = nil
		}
	}
}

func Abc_FlowRetime_ConnectBiasNode(_ *abc.Abc_Obj_t, start *abc.Abc_Obj_t, biasLag int32) {
	if start == nil {
		return
	}
	type state struct {
		obj         *abc.Abc_Obj_t
		relativeLag int32
	}
	stack := []state{{obj: start}}
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(start))
	for len(stack) != 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if abc.Abc_NodeIsTravIdCurrent(current.obj) != 0 {
			continue
		}
		abc.Abc_NodeSetTravIdCurrent(current.obj)
		obj := current.obj
		if abc.Abc_ObjIsLatch(obj) == 0 && abc.Abc_ObjIsBo(obj) == 0 && Abc_FlowRetime_GetLag(start)+current.relativeLag == biasLag {
			FSET(obj, BLOCK)
		}
		for _, next := range obj.Fanouts() {
			if abc.Abc_ObjIsBi(next) != 0 || abc.Abc_ObjIsLatch(next) != 0 || abc.Abc_ObjIsBo(next) != 0 || abc.Abc_ObjIsBo(obj) != 0 {
				stack = append(stack, state{obj: next, relativeLag: biasNextLag(current.relativeLag, abc.Abc_ObjIsLatch(next) != 0)})
			}
		}
	}
}

func Abc_FlowRetime_AddInitBias() {
	for i := 0; i < vec.Vec_PtrSize(pManMR.vInitConstraints); i++ {
		constraint, ok := vec.Vec_PtrEntry(pManMR.vInitConstraints, i).(*InitConstraint_t)
		assert(ok && constraint != nil, "init constraint has expected type")
		if constraint.pBiasNode != nil {
			continue
		}
		assert(vec.Vec_IntSize(constraint.vNodes) == vec.Vec_IntSize(constraint.vLags), "constraint node and lag vectors align")
		for j := 0; j < vec.Vec_IntSize(constraint.vNodes); j++ {
			obj := abc.Abc_NtkObj(pManMR.pNtk, vec.Vec_IntEntry(constraint.vNodes, j))
			assert(obj != nil, "constraint object still exists")
			Abc_FlowRetime_ConnectBiasNode(nil, obj, vec.Vec_IntEntry(constraint.vLags, j))
		}
	}
}

func Abc_FlowRetime_ClearInitToOrig(pInit *abc.Abc_Obj_t) {
	if pInit == nil {
		panic("cannot clear init mapping for a nil object")
	}
	id := abc.Abc_ObjId(pInit)
	ensureInitToOrig(id)
	pManMR.pInitToOrig[int(id)] = NodeLag_t{id: -1}
}

func Abc_FlowRetime_SetInitToOrig(pInit *abc.Abc_Obj_t, pOrig *abc.Abc_Obj_t) {
	if pInit == nil || pOrig == nil {
		panic("cannot set an init mapping with a nil object")
	}
	id := abc.Abc_ObjId(pInit)
	ensureInitToOrig(id)

	// Box boundaries and latches do not correspond to ordinary logic nodes.
	// Constants do: conflict extraction may select one as the source of an
	// impossible initial-state constraint.
	if abc.Abc_ObjIsBo(pOrig) != 0 || abc.Abc_ObjIsBi(pOrig) != 0 || abc.Abc_ObjIsLatch(pOrig) != 0 {
		Abc_FlowRetime_ClearInitToOrig(pInit)
		return
	}

	lag := Abc_FlowRetime_ObjFirstNonLatchBox(pOrig, &pOrig)
	if lag < 0 || pOrig == nil {
		Abc_FlowRetime_ClearInitToOrig(pInit)
		return
	}
	pManMR.pInitToOrig[int(id)] = NodeLag_t{
		id:  abc.Abc_ObjId(pOrig),
		lag: Abc_FlowRetime_GetLag(pOrig) + lag,
	}
}

func Abc_FlowRetime_GetInitToOrig(pInit *abc.Abc_Obj_t, pOrig **abc.Abc_Obj_t, lag *int32) {
	if pInit == nil || pOrig == nil || lag == nil {
		panic("cannot get an init mapping with nil arguments")
	}
	id := abc.Abc_ObjId(pInit)
	if id < 0 || int(id) >= len(pManMR.pInitToOrig) {
		panic("init object has no original-node mapping")
	}
	mapping := pManMR.pInitToOrig[int(id)]
	if mapping.id < 0 {
		if abc.Abc_ObjFaninNum(pInit) == 0 {
			panic("unmapped init object has no fanin")
		}
		Abc_FlowRetime_GetInitToOrig(abc.Abc_ObjFanin0(pInit), pOrig, lag)
		return
	}

	*pOrig = abc.Abc_NtkObj(pManMR.pNtk, mapping.id)
	*lag = mapping.lag
}

//ABC_NAMESPACE_IMPL_END
