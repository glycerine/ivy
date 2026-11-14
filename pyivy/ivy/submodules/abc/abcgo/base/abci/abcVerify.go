package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sim"
	"github.com/glycerine/abcgo/proof/fraig"
	"os"
	"unsafe"
)

/**CFile****************************************************************

  FileName    [abcVerify.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Combinational and sequential verification for two networks.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcVerify.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "base/cmd/cmd.h"
//#include "proof/fraig/fraig.h"
//#include "opt/sim/sim.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include "aig/gia/gia.h"
//#include "proof/ssw/ssw.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void  Abc_NtkVerifyReportError( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel );

// extern void  Abc_NtkVerifyReportErrorSeq( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, int nFrames );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Verifies combinational equivalence by brute-force SAT.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCecSat(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConfLimit int32, nInsLimit int32) {
	pMiter := Abc_NtkMiter(pNtk1, pNtk2, 1, 0, 0, 0)
	if pMiter == nil {
		fmt.Printf("Miter computation has failed.\n")
		return
	}
	retValue := Abc_NtkMiterIsConstant(pMiter)
	if retValue == 0 {
		fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n")
		abcVerifySetModel(pMiter, Abc_NtkVerifyGetCleanModel(pMiter, 1), 1)
		Abc_NtkVerifyReportError(pNtk1, pNtk2, abcVerifyModelPtr(pMiter))
		pMiter.Model = nil
		abc.Abc_NtkDelete(pMiter)
		return
	}
	if retValue == 1 {
		abc.Abc_NtkDelete(pMiter)
		fmt.Printf("Networks are equivalent after structural hashing.\n")
		return
	}
	pCnf := Abc_NtkMulti(pMiter, 0, 100, 1, 0, 0, 0)
	abc.Abc_NtkDelete(pMiter)
	if pCnf == nil {
		fmt.Printf("Renoding for CNF has failed.\n")
		return
	}
	retValue = Abc_NtkMiterSat(pCnf, ABC_INT64_T(nConfLimit), ABC_INT64_T(nInsLimit), 0, nil, nil)
	switch retValue {
	case -1:
		fmt.Printf("Networks are undecided (SAT solver timed out).\n")
	case 0:
		fmt.Printf("Networks are NOT EQUIVALENT after SAT.\n")
	default:
		fmt.Printf("Networks are equivalent after SAT.\n")
	}
	if pCnf.Model != nil {
		Abc_NtkVerifyReportError(pNtk1, pNtk2, abcVerifyModelPtr(pCnf))
	}
	pCnf.Model = nil
	abc.Abc_NtkDelete(pCnf)
}

/*
*Function*************************************************************

	Synopsis    [Verifies sequential equivalence by fraiging followed by SAT.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCecFraig(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, fVerbose int32) {
	clk := Abc_Clock()
	var params fraig.Prove_Params_t
	pExdc := (*Abc_Ntk_t)(nil)
	if pNtk1.Exdc != nil || pNtk2.Exdc != nil {
		if pNtk1.Exdc != nil && pNtk2.Exdc != nil {
			fmt.Printf("Comparing EXDC of the two networks:\n")
			Abc_NtkCecFraig(pNtk1.Exdc, pNtk2.Exdc, nSeconds, fVerbose)
			fmt.Printf("Comparing networks under EXDC of the first network.\n")
			pExdc = pNtk1.Exdc
		} else if pNtk1.Exdc != nil {
			fmt.Printf("Second network has no EXDC. Comparing main networks under EXDC of the first network.\n")
			pExdc = pNtk1.Exdc
		} else {
			fmt.Printf("First network has no EXDC. Comparing main networks under EXDC of the second network.\n")
			pExdc = pNtk2.Exdc
		}
	}
	pMiter := Abc_NtkMiter(pNtk1, pNtk2, 1, 0, 0, 0)
	if pMiter == nil {
		fmt.Printf("Miter computation has failed.\n")
		return
	}
	if pExdc != nil {
		assert(abc.Abc_NtkPoNum(pMiter) == 1, "abc.Abc_NtkPoNum(pMiter) == 1")
		assert(abc.Abc_NtkPoNum(pExdc) == 1, "abc.Abc_NtkPoNum(pExdc) == 1")
		pTemp := pMiter
		pMiter = Abc_NtkMiter(pTemp, pExdc, 1, 0, 1, 0)
		abc.Abc_NtkDelete(pTemp)
	}
	retValue := Abc_NtkMiterIsConstant(pMiter)
	if retValue == 0 {
		fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.  ")
		abcVerifySetModel(pMiter, Abc_NtkVerifyGetCleanModel(pMiter, 1), 1)
		Abc_NtkVerifyReportError(pNtk1, pNtk2, abcVerifyModelPtr(pMiter))
		pMiter.Model = nil
		abc.Abc_NtkDelete(pMiter)
		Abc_PrintTime(1, "Time", Abc_Clock()-clk)
		return
	}
	if retValue == 1 {
		fmt.Printf("Networks are equivalent after structural hashing.  ")
		abc.Abc_NtkDelete(pMiter)
		Abc_PrintTime(1, "Time", Abc_Clock()-clk)
		return
	}
	fraig.Prove_ParamsSetDefault(&params)
	fraig.Prove_ParamsSetItersMax(&params, 5)
	retValue = Abc_NtkIvyProve(&pMiter, &params)
	if retValue == -1 {
		fmt.Printf("Networks are undecided (resource limits is reached).  ")
	} else if retValue == 0 {
		values := Abc_NtkVerifySimulatePattern(pMiter, abcVerifyModelPtr(pMiter))
		if values == nil || *values != 1 {
			fmt.Printf("ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.\n")
		} else {
			fmt.Printf("Networks are NOT EQUIVALENT.  ")
		}
	} else {
		fmt.Printf("Networks are equivalent.  ")
	}
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	if pMiter.Model != nil {
		Abc_NtkVerifyReportError(pNtk1, pNtk2, abcVerifyModelPtr(pMiter))
	}
	abc.Abc_NtkDelete(pMiter)
}

/*
*Function*************************************************************

	Synopsis    [Verifies sequential equivalence by fraiging followed by SAT.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCecFraigPart(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, nPartSize int32, fVerbose int32) {
	var params fraig.Prove_Params_t
	fraig.Prove_ParamsSetDefault(&params)
	fraig.Prove_ParamsSetItersMax(&params, 5)
	assert(nPartSize > 0, "nPartSize > 0")
	pMiter := Abc_NtkMiter(pNtk1, pNtk2, 1, nPartSize, 0, 0)
	if pMiter == nil {
		fmt.Printf("Miter computation has failed.\n")
		return
	}
	retValue := Abc_NtkMiterIsConstant(pMiter)
	if retValue == 0 {
		fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n")
		abcVerifySetModel(pMiter, Abc_NtkVerifyGetCleanModel(pMiter, 1), 1)
		Abc_NtkVerifyReportError(pNtk1, pNtk2, abcVerifyModelPtr(pMiter))
		pMiter.Model = nil
		abc.Abc_NtkDelete(pMiter)
		return
	}
	if retValue == 1 {
		fmt.Printf("Networks are equivalent after structural hashing.\n")
		abc.Abc_NtkDelete(pMiter)
		return
	}
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "unset progressbar")
	status, nOutputs := int32(1), int32(0)
	for i := int32(0); i < abc.Abc_NtkPoNum(pMiter); i++ {
		pObj := abc.Abc_NtkPo(pMiter, i)
		var pMiterPart *Abc_Ntk_t
		if abc.Abc_ObjFanin0(pObj) == abc.Abc_AigConst1(pMiter).Object {
			if abc.Abc_ObjFaninC0(pObj) != 0 {
				retValue = 1
			} else {
				retValue = 0
			}
		} else {
			pMiterPart = abc.Abc_NtkCreateCone(pMiter, abc.Abc_ObjFanin0(pObj), abc.Abc_ObjName(pObj), 0)
			if abc.Abc_ObjFaninC0(pObj) != 0 {
				abc.Abc_ObjXorFaninC(abc.Abc_NtkPo(pMiterPart, 0), 0)
			}
			retValue = Abc_NtkIvyProve(&pMiterPart, &params)
		}
		if retValue == -1 {
			fmt.Printf("Networks are undecided (resource limits is reached).\r")
			status = -1
		} else if retValue == 0 {
			if pMiterPart != nil {
				values := Abc_NtkVerifySimulatePattern(pMiterPart, abcVerifyModelPtr(pMiterPart))
				if values == nil || *values != 1 {
					fmt.Printf("ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.\n")
				} else {
					fmt.Printf("Networks are NOT EQUIVALENT.                 \n")
				}
			}
			status = 0
			if pMiterPart != nil {
				abc.Abc_NtkDelete(pMiterPart)
			}
			break
		} else {
			fmt.Printf("Finished part %5d (out of %5d)\r", i+1, abc.Abc_NtkPoNum(pMiter))
			nOutputs += nPartSize
		}
		if pMiterPart != nil {
			abc.Abc_NtkDelete(pMiterPart)
		}
	}
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "set progressbar")
	if status == 1 {
		fmt.Printf("Networks are equivalent.                         \n")
	} else if status == -1 {
		fmt.Printf("Timed out after verifying %d outputs (out of %d).\n", nOutputs, abc.Abc_NtkCoNum(pNtk1))
	}
	abc.Abc_NtkDelete(pMiter)
}

/*
*Function*************************************************************

	Synopsis    [Verifies sequential equivalence by fraiging followed by SAT.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCecFraigPartAuto(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, fVerbose int32) {
	var params fraig.Prove_Params_t
	fraig.Prove_ParamsSetDefault(&params)
	fraig.Prove_ParamsSetItersMax(&params, 5)
	pMiter := Abc_NtkMiter(pNtk1, pNtk2, 1, 1, 0, 0)
	if pMiter == nil {
		fmt.Printf("Miter computation has failed.\n")
		return
	}
	retValue := Abc_NtkMiterIsConstant(pMiter)
	if retValue == 0 || retValue == 1 {
		if retValue == 0 {
			fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n")
			abcVerifySetModel(pMiter, Abc_NtkVerifyGetCleanModel(pMiter, 1), 1)
			Abc_NtkVerifyReportError(pNtk1, pNtk2, abcVerifyModelPtr(pMiter))
			pMiter.Model = nil
		} else {
			fmt.Printf("Networks are equivalent after structural hashing.\n")
		}
		abc.Abc_NtkDelete(pMiter)
		return
	}
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "unset progressbar")
	vParts := Abc_NtkPartitionSmart(pMiter, 300, 0)
	status, nOutputs := int32(1), int32(0)
	vOnePtr := vec.Vec_PtrAlloc(1000)
	for i := int32(0); i < int32(vec.Vec_PtrSize(vParts)); i++ {
		vOne := vec.Vec_PtrEntry(vParts, int(i)).(*Vec_Int_t)
		Abc_NtkConvertCos(pMiter, vOne, vOnePtr)
		roots := make([]*abc.Object, vec.Vec_PtrSize(vOnePtr))
		for j := range roots {
			roots[j] = vec.Vec_PtrEntry(vOnePtr, j).(*abc.Object)
		}
		pMiterPart := abc.Abc_NtkCreateConeArray(pMiter, roots, 0)
		Abc_NtkCombinePos(pMiterPart, 0, 0)
		retValue = Abc_NtkMiterIsConstant(pMiterPart)
		if retValue == 0 {
			fmt.Printf("Networks are NOT EQUIVALENT after partitioning.\n")
			abc.Abc_NtkDelete(pMiterPart)
			break
		}
		if retValue == 1 {
			abc.Abc_NtkDelete(pMiterPart)
			continue
		}
		fmt.Printf("Verifying part %4d  (out of %4d)  PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, vec.Vec_PtrSize(vParts), abc.Abc_NtkPiNum(pMiterPart), abc.Abc_NtkPoNum(pMiterPart), abc.Abc_NtkNodeNum(pMiterPart), abc.Abc_AigLevel(pMiterPart))
		_ = os.Stdout.Sync()
		retValue = Abc_NtkIvyProve(&pMiterPart, &params)
		if retValue == -1 {
			fmt.Printf("Networks are undecided (resource limits is reached).\r")
			status = -1
		} else if retValue == 0 {
			values := Abc_NtkVerifySimulatePattern(pMiterPart, abcVerifyModelPtr(pMiterPart))
			if values == nil || *values != 1 {
				fmt.Printf("ERROR in Abc_NtkMiterProve(): Generated counter-example is invalid.\n")
			} else {
				fmt.Printf("Networks are NOT EQUIVALENT.                 \n")
			}
			status = 0
			abc.Abc_NtkDelete(pMiterPart)
			break
		} else {
			nOutputs += int32(vec.Vec_IntSize(vOne))
		}
		abc.Abc_NtkDelete(pMiterPart)
	}
	fmt.Printf("                                                                                          \r")
	for i := 0; i < vec.Vec_PtrSize(vParts); i++ {
		vec.Vec_IntFree(vec.Vec_PtrEntry(vParts, i).(*vec.Vec_Int_t))
	}
	vec.Vec_PtrFree(vParts)
	vec.Vec_PtrFree(vOnePtr)
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "set progressbar")
	if status == 1 {
		fmt.Printf("Networks are equivalent.                         \n")
	} else if status == -1 {
		fmt.Printf("Timed out after verifying %d outputs (out of %d).\n", nOutputs, abc.Abc_NtkCoNum(pNtk1))
	}
	abc.Abc_NtkDelete(pMiter)
}

/*
*Function*************************************************************

	Synopsis    [Verifies sequential equivalence by brute-force SAT.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSecSat(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConfLimit int32, nInsLimit int32, nFrames int32) {
	pMiter := Abc_NtkMiter(pNtk1, pNtk2, 0, 0, 0, 0)
	if pMiter == nil {
		fmt.Printf("Miter computation has failed.\n")
		return
	}
	retValue := Abc_NtkMiterIsConstant(pMiter)
	if retValue == 0 || retValue == 1 {
		abc.Abc_NtkDelete(pMiter)
		if retValue == 0 {
			fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n")
		} else {
			fmt.Printf("Networks are equivalent after structural hashing.\n")
		}
		return
	}
	pFrames := Abc_NtkFrames(pMiter, nFrames, 1, 0)
	abc.Abc_NtkDelete(pMiter)
	if pFrames == nil {
		fmt.Printf("Frames computation has failed.\n")
		return
	}
	retValue = Abc_NtkMiterIsConstant(pFrames)
	if retValue == 0 || retValue == 1 {
		abc.Abc_NtkDelete(pFrames)
		if retValue == 0 {
			fmt.Printf("Networks are NOT EQUIVALENT after framing.\n")
		} else {
			fmt.Printf("Networks are equivalent after framing.\n")
		}
		return
	}
	pCnf := Abc_NtkMulti(pFrames, 0, 100, 1, 0, 0, 0)
	abc.Abc_NtkDelete(pFrames)
	if pCnf == nil {
		fmt.Printf("Renoding for CNF has failed.\n")
		return
	}
	retValue = Abc_NtkMiterSat(pCnf, ABC_INT64_T(nConfLimit), ABC_INT64_T(nInsLimit), 0, nil, nil)
	switch retValue {
	case -1:
		fmt.Printf("Networks are undecided (SAT solver timed out).\n")
	case 0:
		fmt.Printf("Networks are NOT EQUIVALENT after SAT.\n")
	default:
		fmt.Printf("Networks are equivalent after SAT.\n")
	}
	abc.Abc_NtkDelete(pCnf)
}

/*
*Function*************************************************************

	Synopsis    [Verifies combinational equivalence by fraiging followed by SAT]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func abcVerifyNtkToFraig(pNtk *Abc_Ntk_t, params *fraig.Fraig_Params_t) *fraig.Fraig_Man_t {
	graph := abc.Abc_NtkToDar(pNtk, 0, 0)
	if graph == nil {
		return nil
	}
	manager := fraig.Fraig_ManCreate(params)
	nodes := graph.Nodes()
	mapped := make([]*fraig.Fraig_Node_t, len(nodes))
	mapped[0] = fraig.Fraig_ManReadConst1(manager)
	inputIDs := make(map[int]int, len(graph.Inputs()))
	for i, input := range graph.Inputs() {
		inputIDs[input.NodeID()] = i
		mapped[input.NodeID()] = fraig.Fraig_ManReadIthVar(manager, int32(i))
	}
	var convert func(aig.Edge) *fraig.Fraig_Node_t
	convert = func(edge aig.Edge) *fraig.Fraig_Node_t {
		id := edge.NodeID()
		if id < 0 || id >= len(mapped) {
			panic("abcVerify: FRAIG edge outside AIG")
		}
		base := mapped[id]
		if base == nil {
			node := nodes[id]
			switch node.Kind {
			case aig.And:
				left := convert(node.Fanins[0])
				right := convert(node.Fanins[1])
				base = fraig.Fraig_NodeAnd(manager, left, right)
			case aig.ExclusiveOr:
				left := convert(node.Fanins[0])
				right := convert(node.Fanins[1])
				base = fraig.Fraig_NodeExor(manager, left, right)
			case aig.Buffer, aig.Output:
				base = convert(node.Fanins[0])
			case aig.Input:
				index, ok := inputIDs[id]
				if !ok {
					panic("abcVerify: missing FRAIG input")
				}
				base = fraig.Fraig_ManReadIthVar(manager, int32(index))
			default:
				panic("abcVerify: unsupported AIG node in FRAIG conversion")
			}
			mapped[id] = base
		}
		return fraig.Fraig_NotCond(base, edge.Complemented())
	}
	for _, output := range graph.Outputs() {
		fraig.Fraig_ManSetPo(manager, convert(output))
	}
	aig.Aig_ManStop(graph)
	return manager
}

func Abc_NtkSecFraig(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nSeconds int32, nFrames int32, fVerbose int32) int32 {
	pMiter := Abc_NtkMiter(pNtk1, pNtk2, 0, 0, 0, 0)
	if pMiter == nil {
		fmt.Printf("Miter computation has failed.\n")
		return 0
	}
	retValue := Abc_NtkMiterIsConstant(pMiter)
	if retValue == 0 {
		fmt.Printf("Networks are NOT EQUIVALENT after structural hashing.\n")
		abcVerifySetModel(pMiter, Abc_NtkVerifyGetCleanModel(pMiter, nFrames), nFrames)
		Abc_NtkVerifyReportErrorSeq(pNtk1, pNtk2, abcVerifyModelPtr(pMiter), nFrames)
		pMiter.Model = nil
		abc.Abc_NtkDelete(pMiter)
		return 0
	}
	if retValue == 1 {
		abc.Abc_NtkDelete(pMiter)
		fmt.Printf("Networks are equivalent after structural hashing.\n")
		return 1
	}
	pFrames := Abc_NtkFrames(pMiter, nFrames, 1, 0)
	abc.Abc_NtkDelete(pMiter)
	if pFrames == nil {
		fmt.Printf("Frames computation has failed.\n")
		return 0
	}
	retValue = Abc_NtkMiterIsConstant(pFrames)
	if retValue == 0 || retValue == 1 {
		pFrames.Model = nil
		abc.Abc_NtkDelete(pFrames)
		if retValue == 0 {
			fmt.Printf("Networks are NOT EQUIVALENT after framing.\n")
			return 0
		}
		fmt.Printf("Networks are equivalent after framing.\n")
		return 1
	}
	var params fraig.Fraig_Params_t
	fraig.Fraig_ParamsSetDefault(&params)
	params.FVerbose = fVerbose
	params.NSeconds = nSeconds
	pMan := abcVerifyNtkToFraig(pFrames, &params)
	fraig.Fraig_ManProveMiter(pMan)
	retValue = fraig.Fraig_ManCheckMiter(pMan)
	switch retValue {
	case -1:
		fmt.Printf("Networks are undecided (SAT solver timed out on the final miter).\n")
	case 1:
		fmt.Printf("Networks are equivalent after fraiging.\n")
	case 0:
		fmt.Printf("Networks are NOT EQUIVALENT after fraiging.\n")
	default:
		assert(false, "0")
	}
	fraig.Fraig_ManFree(pMan)
	abc.Abc_NtkDelete(pFrames)
	if retValue == 1 {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Returns a dummy pattern full of zeros.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func abcVerifySetModel(pNtk *Abc_Ntk_t, model *int32, nFrames int32) {
	if model == nil {
		pNtk.Model = nil
		return
	}
	count := int(abc.Abc_NtkCiNum(pNtk) * nFrames)
	if count <= 0 {
		pNtk.Model = nil
		return
	}
	pNtk.Model = unsafe.Slice(model, count)
}

func abcVerifyModelPtr(pNtk *Abc_Ntk_t) *int32 {
	if len(pNtk.Model) == 0 {
		return nil
	}
	return &pNtk.Model[0]
}

func Abc_NtkVerifyGetCleanModel(pNtk *Abc_Ntk_t, nFrames int32) *int32 {
	n := int(abc.Abc_NtkCiNum(pNtk) * nFrames)
	if n <= 0 {
		return nil
	}
	model := make([]int32, n)
	pNtk.Model = model
	return &model[0]
}

/*
*Function*************************************************************

	Synopsis    [Returns the PO values under the given input pattern.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkVerifySimulatePattern(pNtk *Abc_Ntk_t, pModel *int32) *int32 {
	strashed := false
	if abc.Abc_NtkIsStrash(pNtk) == 0 {
		pNtk = Abc_NtkStrash(pNtk, 0, 0, 0)
		if pNtk == nil {
			return nil
		}
		strashed = true
	}
	ciCount := int(abc.Abc_NtkCiNum(pNtk))
	model := unsafe.Slice(pModel, ciCount)
	abc.Abc_AigConst1(pNtk).Object.CopyData = true
	for i := 0; i < ciCount; i++ {
		abc.Abc_NtkCi(pNtk, int32(i)).CopyData = model[i] != 0
	}
	for i := int32(0); i < abc.Abc_NtkObjNum(pNtk); i++ {
		node := abc.Abc_NtkObj(pNtk, i)
		if node == nil || abc.Abc_ObjIsNode(node) == 0 {
			continue
		}
		v0 := abcVerifyCopiedValue(abc.Abc_ObjFanin0(node)) != (abc.Abc_ObjFaninC0(node) != 0)
		v1 := abcVerifyCopiedValue(abc.Abc_ObjFanin1(node)) != (abc.Abc_ObjFaninC1(node) != 0)
		node.CopyData = v0 && v1
	}
	coCount := int(abc.Abc_NtkCoNum(pNtk))
	values := make([]int32, coCount)
	for i := 0; i < coCount; i++ {
		co := abc.Abc_NtkCo(pNtk, int32(i))
		value := abcVerifyCopiedValue(abc.Abc_ObjFanin0(co)) != (abc.Abc_ObjFaninC0(co) != 0)
		if value {
			values[i] = 1
		}
	}
	if strashed {
		abc.Abc_NtkDelete(pNtk)
	}
	if len(values) == 0 {
		return nil
	}
	return &values[0]
}

func abcVerifyCopiedValue(obj *Abc_Obj_t) bool {
	value, ok := obj.CopyData.(bool)
	return ok && value
}

/*
*Function*************************************************************

	Synopsis    [Reports mismatch between the two networks.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkVerifyReportError(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pModel *int32) {
	if abc.Abc_NtkCiNum(pNtk1) != abc.Abc_NtkCiNum(pNtk2) || abc.Abc_NtkCoNum(pNtk1) != abc.Abc_NtkCoNum(pNtk2) {
		return
	}
	coCount := int(abc.Abc_NtkCoNum(pNtk1))
	values1 := unsafe.Slice(Abc_NtkVerifySimulatePattern(pNtk1, pModel), coCount)
	values2 := unsafe.Slice(Abc_NtkVerifySimulatePattern(pNtk2, pModel), coCount)
	nErrors, first := 0, -1
	for i := 0; i < coCount; i++ {
		if values1[i] != values2[i] {
			nErrors++
			if first < 0 {
				first = i
			}
		}
	}
	fmt.Printf("Verification failed for at least %d outputs: ", nErrors)
	printed := 0
	for i := 0; i < coCount && printed < 3; i++ {
		if values1[i] != values2[i] {
			fmt.Printf(" %s", abc.Abc_ObjName(abc.Abc_NtkCo(pNtk1, int32(i))))
			printed++
		}
	}
	if printed != nErrors {
		fmt.Printf(" ...")
	}
	fmt.Printf("\n")
	if first < 0 {
		return
	}
	fmt.Printf("Output %s: Value in Network1 = %d. Value in Network2 = %d.\n", abc.Abc_ObjName(abc.Abc_NtkCo(pNtk1, int32(first))), values1[first], values2[first])
	fmt.Printf("Input pattern: ")
	co := abc.Abc_NtkCo(pNtk1, int32(first))
	support := abc.Abc_NtkNodeSupport(pNtk1, []*abc.Object{co}, 1)
	for i := int32(0); i < abc.Abc_NtkCiNum(pNtk1); i++ {
		abc.Abc_NtkCi(pNtk1, i).CopyData = i
	}
	for _, node := range support {
		if !node.IsCI() {
			continue
		}
		index, ok := node.CopyData.(int32)
		if ok {
			fmt.Printf(" %s=%d", abc.Abc_ObjName(node), unsafe.Slice(pModel, int(abc.Abc_NtkCiNum(pNtk1)))[index])
		}
	}
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Computes the COs in the support of the PO in the given frame.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkGetSeqPoSupp(pNtk *Abc_Ntk_t, iFrame int32, iNumPo int32) {
	frames := Abc_NtkFrames(pNtk, iFrame+1, 0, 0)
	if frames == nil {
		return
	}
	po := abc.Abc_NtkPo(frames, iFrame*abc.Abc_NtkPoNum(pNtk)+iNumPo)
	support := abc.Abc_NtkNodeSupport(frames, []*abc.Object{po}, 1)
	frameSupport := make(map[*Abc_Obj_t]bool, len(support))
	for _, obj := range support {
		frameSupport[obj] = true
	}
	for i := int32(0); i < abc.Abc_NtkCiNum(frames); i++ {
		abc.Abc_NtkCi(frames, i).CopyData = nil
	}
	for _, obj := range support {
		obj.CopyData = true
	}
	for i := int32(0); i < abc.Abc_NtkCiNum(pNtk); i++ {
		abc.Abc_NtkCi(pNtk, i).CopyData = nil
	}
	for i := int32(0); i < abc.Abc_NtkBoxNum(pNtk); i++ {
		latch := abc.Abc_NtkBox(pNtk, i)
		if abc.Abc_ObjIsLatch(latch) == 0 {
			continue
		}
		if frameSupport[abc.Abc_NtkBox(frames, i)] {
			latch.CopyData = true
		}
	}
	for i := int32(0); i < abc.Abc_NtkPiNum(pNtk); i++ {
		for k := int32(0); k <= iFrame; k++ {
			framePi := abc.Abc_NtkPi(frames, k*abc.Abc_NtkPiNum(pNtk)+i)
			if frameSupport[framePi] {
				abc.Abc_NtkPi(pNtk, i).CopyData = true
			}
		}
	}
	abc.Abc_NtkDelete(frames)
}

/*
*Function*************************************************************

	Synopsis    [Reports mismatch between the two sequential networks.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkVerifyReportErrorSeq(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pModel *int32, nFrames int32) {
	values1 := abcVerifySeqValues(pNtk1, pModel, nFrames)
	values2 := abcVerifySeqValues(pNtk2, pModel, nFrames)
	poCount := int(abc.Abc_NtkPoNum(pNtk1))
	firstFrame, firstOutput, nErrors := -1, -1, 0
	for frame := 0; frame < int(nFrames) && firstFrame < 0; frame++ {
		for output := 0; output < poCount; output++ {
			po1 := abc.Abc_NtkPo(pNtk1, int32(output))
			po2 := abc.Abc_NtkPo(pNtk2, int32(output))
			if values1[po1][frame] != values2[po2][frame] {
				nErrors++
				if firstOutput < 0 {
					firstFrame, firstOutput = frame, output
				}
			}
		}
	}
	if firstOutput < 0 {
		fmt.Printf("No output mismatches detected.\n")
		return
	}
	plural := ""
	if nErrors > 1 {
		plural = "s"
	}
	fmt.Printf("Verification failed for at least %d output%s of frame %d: ", nErrors, plural, firstFrame+1)
	printed := 0
	for output := 0; output < poCount && printed < 3; output++ {
		po1, po2 := abc.Abc_NtkPo(pNtk1, int32(output)), abc.Abc_NtkPo(pNtk2, int32(output))
		if values1[po1][firstFrame] != values2[po2][firstFrame] {
			fmt.Printf(" %s", abc.Abc_ObjName(po1))
			printed++
		}
	}
	if printed != nErrors {
		fmt.Printf(" ...")
	}
	fmt.Printf("\n")
	Abc_NtkGetSeqPoSupp(pNtk1, int32(firstFrame), int32(firstOutput))
	Abc_NtkGetSeqPoSupp(pNtk2, int32(firstFrame), int32(firstOutput))
	po1, po2 := abc.Abc_NtkPo(pNtk1, int32(firstOutput)), abc.Abc_NtkPo(pNtk2, int32(firstOutput))
	fmt.Printf("Output %s: Value in Network1 = %d. Value in Network2 = %d.\n", abc.Abc_ObjName(po1), boolToInt32(values1[po1][firstFrame]), boolToInt32(values2[po2][firstFrame]))
	for networkIndex, network := range []*Abc_Ntk_t{pNtk1, pNtk2} {
		fmt.Printf("The cone of influence of output %s in Network%d:\n", abc.Abc_ObjName(po1), networkIndex+1)
		fmt.Printf("PIs: ")
		for i := int32(0); i < abc.Abc_NtkPiNum(network); i++ {
			pi := abc.Abc_NtkPi(network, i)
			if pi.CopyData == true {
				fmt.Printf("%s ", abc.Abc_ObjName(pi))
			}
		}
		fmt.Printf("\nLatches: ")
		for i := int32(0); i < abc.Abc_NtkBoxNum(network); i++ {
			latch := abc.Abc_NtkBox(network, i)
			if abc.Abc_ObjIsLatch(latch) != 0 && latch.CopyData == true {
				fmt.Printf("%s ", abc.Abc_ObjName(latch))
			}
		}
		fmt.Printf("\n")
	}
	for frame := 0; frame <= firstFrame; frame++ {
		fmt.Printf("Frame %d:  PI(1):", frame+1)
		printSeqPattern(pNtk1, values1, frame, true)
		fmt.Printf(" L(1):")
		printSeqLatchPattern(pNtk1, values1, frame, true)
		fmt.Printf(" %s(1):%d  PI(2):", abc.Abc_ObjName(po1), boolToInt32(values1[po1][frame]))
		printSeqPattern(pNtk2, values2, frame, true)
		fmt.Printf(" L(2):")
		printSeqLatchPattern(pNtk2, values2, frame, true)
		fmt.Printf(" %s(2):%d\n", abc.Abc_ObjName(po1), boolToInt32(values2[po2][frame]))
	}
	for _, network := range []*Abc_Ntk_t{pNtk1, pNtk2} {
		for i := int32(0); i < abc.Abc_NtkCiNum(network); i++ {
			abc.Abc_NtkCi(network, i).CopyData = nil
		}
	}
}

func abcVerifySeqValues(pNtk *Abc_Ntk_t, pModel *int32, nFrames int32) map[*Abc_Obj_t][]bool {
	frames := int(nFrames)
	values := make(map[*Abc_Obj_t][]bool, int(abc.Abc_NtkObjNum(pNtk)))
	if frames == 0 {
		return values
	}
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	assert(abc.Abc_NtkBoxNum(pNtk) == abc.Abc_NtkLatchNum(pNtk), "Abc_NtkBoxNum(pNtk) == Abc_NtkLatchNum(pNtk)")
	initial := make([]int8, int(abc.Abc_NtkLatchNum(pNtk)))
	latchIndex := 0
	for i := int32(0); i < abc.Abc_NtkBoxNum(pNtk); i++ {
		latch := abc.Abc_NtkBox(pNtk, i)
		if abc.Abc_ObjIsLatch(latch) == 0 {
			continue
		}
		if abc.Abc_LatchIsInit0(latch) == 0 && abc.Abc_LatchIsInit1(latch) == 0 {
			initial[latchIndex] = -1 // C uses SIM_RANDOM_UNSIGNED for unspecified initial state.
		}
		latchIndex++
	}
	graph := abc.Abc_NtkToDar(pNtk, 0, 1)
	if graph == nil {
		return values
	}
	defer aig.Aig_ManStop(graph)
	piCount := int(abc.Abc_NtkPiNum(pNtk))
	modelInts := unsafe.Slice(pModel, frames*piCount)
	model := make([]bool, len(modelInts))
	for i, bit := range modelInts {
		model[i] = bit != 0
	}
	info := sim.SequentialSimulation(graph, frames, 1, model, initial)
	inputs := graph.Inputs()
	outputs := graph.Outputs()
	readEdge := func(edge aig.Edge) []bool {
		bits := make([]bool, frames)
		row := info[edge.NodeID()]
		for frame := range bits {
			bits[frame] = (row[frame] != 0) != edge.Complemented()
		}
		return bits
	}
	for i := 0; i < piCount; i++ {
		values[abc.Abc_NtkPi(pNtk, int32(i))] = readEdge(inputs[i])
	}
	latchIndex = 0
	for i := int32(0); i < abc.Abc_NtkBoxNum(pNtk); i++ {
		latch := abc.Abc_NtkBox(pNtk, i)
		if abc.Abc_ObjIsLatch(latch) == 0 {
			continue
		}
		state := readEdge(inputs[piCount+latchIndex])
		if abc.Abc_LatchIsInit1(latch) != 0 {
			for frame := range state {
				state[frame] = !state[frame]
			}
		}
		values[abc.Abc_ObjFanout0(latch)] = state
		latchIndex++
	}
	for i := 0; i < int(abc.Abc_NtkPoNum(pNtk)); i++ {
		values[abc.Abc_NtkPo(pNtk, int32(i))] = readEdge(outputs[i])
	}
	return values
}

func printSeqPattern(pNtk *Abc_Ntk_t, values map[*Abc_Obj_t][]bool, frame int, _ bool) {
	for i := int32(0); i < abc.Abc_NtkPiNum(pNtk); i++ {
		pi := abc.Abc_NtkPi(pNtk, i)
		if pi.CopyData == true {
			fmt.Printf("%d", boolToInt32(values[pi][frame]))
		}
	}
}

func printSeqLatchPattern(pNtk *Abc_Ntk_t, values map[*Abc_Obj_t][]bool, frame int, _ bool) {
	for i := int32(0); i < abc.Abc_NtkBoxNum(pNtk); i++ {
		latch := abc.Abc_NtkBox(pNtk, i)
		if abc.Abc_ObjIsLatch(latch) == 0 || latch.CopyData != true {
			continue
		}
		bo := abc.Abc_ObjFanout0(latch)
		if boValues := values[bo]; frame < len(boValues) {
			fmt.Printf("%d", boolToInt32(boValues[frame]))
		}
	}
}

func boolToInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Simulates buggy miter emailed by Mike.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSimulteBuggyMiter(pNtk *Abc_Ntk_t) {
	piValues1 := "01001011100000000011010110101000000"
	piValues2 := "11001101011101011111110100100010001"
	if len(piValues1) != int(abc.Abc_NtkPiNum(pNtk)) || abc.Abc_NtkPoNum(pNtk) != 1 {
		return
	}
	ciCount := int(abc.Abc_NtkCiNum(pNtk))
	piCount := int(abc.Abc_NtkPiNum(pNtk))
	model1 := make([]int32, ciCount)
	model2 := make([]int32, ciCount)
	for i := 0; i < piCount; i++ {
		model1[i] = int32(piValues1[i] - '0')
		model2[i] = int32(piValues2[i] - '0')
	}
	latchIndex := 0
	for i := int32(0); i < abc.Abc_NtkBoxNum(pNtk); i++ {
		latch := abc.Abc_NtkBox(pNtk, i)
		if abc.Abc_ObjIsLatch(latch) == 0 {
			continue
		}
		init, ok := latch.Data.(abc.Abc_InitType_t)
		if ok {
			model1[piCount+latchIndex] = int32(init) - 1
		}
		latchIndex++
	}
	result1 := Abc_NtkVerifySimulatePattern(pNtk, &model1[0])
	if result1 == nil {
		return
	}
	coValues1 := unsafe.Slice(result1, int(abc.Abc_NtkCoNum(pNtk)))
	fmt.Printf("Value = %d\n", coValues1[0])
	latchIndex = 0
	for i := int32(0); i < abc.Abc_NtkBoxNum(pNtk); i++ {
		if abc.Abc_ObjIsLatch(abc.Abc_NtkBox(pNtk, i)) == 0 {
			continue
		}
		model2[piCount+latchIndex] = coValues1[int(abc.Abc_NtkPoNum(pNtk))+latchIndex]
		latchIndex++
	}
	result2 := Abc_NtkVerifySimulatePattern(pNtk, &model2[0])
	if result2 != nil {
		fmt.Printf("Value = %d\n", unsafe.Slice(result2, int(abc.Abc_NtkCoNum(pNtk)))[0])
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns the PO values under the given input pattern.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIsTrueCex(pNtk *Abc_Ntk_t, pCex *Abc_Cex_t) int32 {
	strashed := false
	if abc.Abc_NtkIsStrash(pNtk) == 0 {
		pNtk = Abc_NtkStrash(pNtk, 0, 0, 0)
		if pNtk == nil {
			return 0
		}
		strashed = true
	}
	pMan := abc.Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		if strashed {
			abc.Abc_NtkDelete(pNtk)
		}
		return 0
	}
	status := saig.Saig_ManVerifyCex(pMan, pCex)
	aig.Aig_ManStop(pMan)
	if strashed {
		abc.Abc_NtkDelete(pNtk)
	}
	return status
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the number of PIs matches.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkIsValidCex(pNtk *Abc_Ntk_t, pCex *Abc_Cex_t) int32 {
	if abc.Abc_NtkPiNum(pNtk) == pCex.NPis {
		return 1
	}
	return 0
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
