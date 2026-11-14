package abci

import (
	"fmt"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [abcQuant.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [AIG-based variable quantification.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcQuant.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Performs fast synthesis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSynthesize(ppNtk **Abc_Ntk_t, fMoreEffort int32) {
	pNtk := *ppNtk
	var pNtkTemp *Abc_Ntk_t

	Abc_NtkRewrite(pNtk, 0, 0, 0, 0, 0)
	Abc_NtkRefactor(pNtk, 10, 16, 0, 0, 0, 0)
	pNtkTemp = pNtk
	pNtk = Abc_NtkBalance(pNtkTemp, 0, 0, 0)
	Abc_NtkDelete(pNtkTemp)

	if fMoreEffort != 0 {
		Abc_NtkRewrite(pNtk, 0, 0, 0, 0, 0)
		Abc_NtkRefactor(pNtk, 10, 16, 0, 0, 0, 0)
		pNtkTemp = pNtk
		pNtk = Abc_NtkBalance(pNtkTemp, 0, 0, 0)
		Abc_NtkDelete(pNtkTemp)

		pNtkTemp = pNtk
		pNtk = Abc_NtkIvyFraig(pNtkTemp, 100, 1, 0, 0, 0)
		Abc_NtkDelete(pNtkTemp)
	}

	*ppNtk = pNtk
}

/*
*Function*************************************************************

	Synopsis    [Existentially quantifies one variable.]

	Description []

	SideEffects [This procedure creates dangling nodes in the AIG.]

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkQuantify(pNtk *Abc_Ntk_t, fUniv int32, iVar int32, fVerbose int32) int32 {
	assert(Abc_NtkIsStrash(pNtk) != 0, "quantification requires a strashed network")
	assert(iVar >= 0 && iVar < Abc_NtkCiNum(pNtk), "quantified CI is in range")
	pivot := baseabc.Abc_NtkCi(pNtk, iVar)
	levels := baseabc.Abc_NtkDfsReverseNodes(pNtk, []*Abc_Obj_t{pivot}, 1)
	manager := pNtk.FunctionManager.(*baseabc.Abc_Aig_t)
	// C uses pCopy and pData as two cofactor slots. Go's Data also owns
	// the node function, so keep the same two values in local maps.
	zero := map[*Abc_Obj_t]baseabc.Edge{pivot: pNtk.One().Not()}
	one := map[*Abc_Obj_t]baseabc.Edge{pivot: pNtk.One()}
	cofactor := func(values map[*Abc_Obj_t]baseabc.Edge, edge baseabc.Edge) baseabc.Edge {
		result, found := values[edge.Object]
		if !found {
			result = edge.Object.Signal()
		}
		if edge.Complement {
			result = result.Not()
		}
		return result
	}
	for _, head := range levels {
		for node := head; node != nil; node = node.Copy {
			fanins := node.Fanins()
			zero[node] = baseabc.Abc_AigAnd(manager, cofactor(zero, fanins[0]), cofactor(zero, fanins[1]))
			one[node] = baseabc.Abc_AigAnd(manager, cofactor(one, fanins[0]), cofactor(one, fanins[1]))
		}
	}
	for _, output := range pNtk.Outputs() {
		if baseabc.Abc_NodeIsTravIdCurrent(output) == 0 {
			continue
		}
		fanin := Abc_ObjFanin0(output)
		var result baseabc.Edge
		if fUniv != 0 {
			result = baseabc.Abc_AigAnd(manager, cofactor(zero, output.Fanins()[0]), cofactor(one, output.Fanins()[0]))
		} else {
			result = baseabc.Abc_AigOr(manager, cofactor(zero, output.Fanins()[0]), cofactor(one, output.Fanins()[0]))
		}
		result = baseabc.Abc_ObjNotCond(result, baseabc.Abc_ObjFaninC0(output))
		if result.Object != fanin {
			baseabc.Abc_ObjPatchFanin(output, fanin, result)
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Constructs the transition relation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkTransRel(pNtk *Abc_Ntk_t, fInputs int32, fVerbose int32) *Abc_Ntk_t {

	var vPairs []baseabc.Edge
	var pNtkNew *Abc_Ntk_t
	var pObj *Abc_Obj_t
	var pMiter baseabc.Edge
	var i, nLatches int32
	var fSynthesis int32 = 1

	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert((Abc_NtkLatchNum(pNtk) != 0), "Abc_NtkLatchNum(pNtk)")
	nLatches = Abc_NtkLatchNum(pNtk)
	// start the network
	pNtkNew = baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
	// duplicate the name and the spec
	Abc_NtkSetName(pNtkNew, Extra_UtilStrsav(Abc_NtkName(pNtk)))
	//    pNtkNew->pSpec = Extra_UtilStrsav(pNtk->pSpec);
	baseabc.Abc_NtkCleanCopy(pNtk)
	// create current state variables
	// inlining of macro Abc_NtkForEachLatchOutput
	for i = int32(0); i < int32(len(pNtk.Boxes())); i++ {
		box := baseabc.Abc_NtkBox(pNtk, i)
		if baseabc.Abc_ObjIsLatch(box) != 0 {
			pObj = Abc_ObjFanout0(box)
			Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(baseabc.Abc_NtkCreatePi(pNtkNew)))
			baseabc.Abc_ObjAssignName(Abc_ObjReadCopyEdge(pObj).Object, Abc_ObjName(pObj), "")
		}
	}
	// create next state variables
	// inlining of macro Abc_NtkForEachLatchInput
	for i = int32(0); i < int32(len(pNtk.Boxes())); i++ {
		box := baseabc.Abc_NtkBox(pNtk, i)
		if baseabc.Abc_ObjIsLatch(box) != 0 {
			pObj = Abc_ObjFanin0(box)
			baseabc.Abc_ObjAssignName(baseabc.Abc_NtkCreatePi(pNtkNew), Abc_ObjName(pObj), "")
		}
	}
	// create PI variables
	// inlining of macro Abc_NtkForEachPi
	for i = int32(0); i < Abc_NtkPiNum(pNtk); i++ {
		pObj = Abc_NtkPi(pNtk, i)
		baseabc.Abc_NtkDupObj(pNtkNew, pObj, 1)
	}
	// create the PO
	baseabc.Abc_NtkCreatePo(pNtkNew)
	// restrash the nodes (assuming a topological order of the old network)
	Abc_ObjSetCopyEdge(baseabc.Abc_AigConst1(pNtk).Object, abcCompatCopyEdge(baseabc.Abc_AigConst1(pNtkNew)))
	// inlining of macro Abc_NtkForEachNode
	for i = int32(0); i < Abc_NtkObjNumMax(pNtk); i++ {
		pObj = Abc_NtkObj(pNtk, i)
		if pObj != nil && (Abc_ObjIsNode(pObj) != 0) {
			Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(baseabc.Abc_AigAnd(pNtkNew.FunctionManager.(*baseabc.Abc_Aig_t), Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj))))
		}
	}
	// create the function of the primary output
	assert(baseabc.Abc_NtkBoxNum(pNtk) == Abc_NtkLatchNum(pNtk), "Abc_NtkBoxNum(pNtk) == Abc_NtkLatchNum(pNtk)")
	vPairs = make([]baseabc.Edge, 0, 2*nLatches)
	// inlining of macro Abc_NtkForEachLatchInput
	for i = int32(0); i < int32(len(pNtk.Boxes())); i++ {
		box := baseabc.Abc_NtkBox(pNtk, i)
		if baseabc.Abc_ObjIsLatch(box) != 0 {
			pObj = Abc_ObjFanin0(box)
			vPairs = append(vPairs, Abc_ObjChild0Copy(pObj))
			vPairs = append(vPairs, Abc_NtkPi(pNtkNew, i+nLatches).Signal())
		}
	}
	pMiter = baseabc.Abc_AigMiter(pNtkNew.FunctionManager.(*baseabc.Abc_Aig_t), &vPairs, 0)
	// add the primary output
	Abc_ObjAddFanin(Abc_NtkPo(pNtkNew, 0), baseabc.Abc_ObjNot(pMiter))
	baseabc.Abc_ObjAssignName(Abc_NtkPo(pNtkNew, 0), "rel", "")

	// quantify inputs
	if fInputs != 0 {
		assert(Abc_NtkPiNum(pNtkNew) == Abc_NtkPiNum(pNtk)+2*nLatches, "Abc_NtkPiNum(pNtkNew) == Abc_NtkPiNum(pNtk) + 2*nLatches")
		for i = Abc_NtkPiNum(pNtkNew) - 1; i >= 2*nLatches; i-- {
			//        for ( i = 2*nLatches; i < Abc_NtkPiNum(pNtkNew); i++ )

			Abc_NtkQuantify(pNtkNew, 0, i, fVerbose)
			//            if ( fSynthesis && (i % 3 == 2) )
			if fSynthesis != 0 {

				baseabc.Abc_AigCleanup(pNtkNew.FunctionManager.(*baseabc.Abc_Aig_t))
				Abc_NtkSynthesize(&pNtkNew, 1)
			}
			//            printf( "Var = %3d. Nodes = %6d. ", Abc_NtkPiNum(pNtkNew) - 1 - i, Abc_NtkNodeNum(pNtkNew) );
			//            printf( "Var = %3d. Nodes = %6d. ", i - 2*nLatches, Abc_NtkNodeNum(pNtkNew) );
		}
		//        printf( "\n" );

		baseabc.Abc_AigCleanup(pNtkNew.FunctionManager.(*baseabc.Abc_Aig_t))
		for i = Abc_NtkPiNum(pNtkNew) - 1; i >= 2*nLatches; i-- {
			pObj = Abc_NtkPi(pNtkNew, i)
			assert(baseabc.Abc_ObjFanoutNum(pObj) == 0, "Abc_ObjFanoutNum(pObj) == 0")
			baseabc.Abc_NtkDeleteObj(pObj)
		}
	}

	// check consistency of the network
	if baseabc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Abc_NtkTransRel: The network check has failed.\n")
		Abc_NtkDelete(pNtkNew)
		return nil
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Performs one image computation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkInitialState(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	var pNtkNew *Abc_Ntk_t
	var pMiter baseabc.Edge
	var i int32
	var nVars int32 = Abc_NtkPiNum(pNtk) / 2
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	// start the new network
	pNtkNew = baseabc.Abc_NtkStartFrom(pNtk, baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG)
	// compute the all-zero state in terms of the CS variables
	pMiter = baseabc.Abc_AigConst1(pNtkNew)
	for i = 0; i < nVars; i++ {
		pMiter = baseabc.Abc_AigAnd(pNtkNew.FunctionManager.(*baseabc.Abc_Aig_t), pMiter, baseabc.Abc_ObjNot(Abc_NtkPi(pNtkNew, i)))
	}
	// add the PO
	Abc_ObjAddFanin(Abc_NtkPo(pNtkNew, 0), pMiter)
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Swaps current state and next state variables.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSwapVariables(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	var pNtkNew *Abc_Ntk_t
	var pMiter baseabc.Edge
	var pObj, pObj0, pObj1 *Abc_Obj_t
	var i int32
	var nVars int32 = Abc_NtkPiNum(pNtk) / 2
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	// start the new network
	pNtkNew = baseabc.Abc_NtkStartFrom(pNtk, baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG)
	// update the PIs
	for i = 0; i < nVars; i++ {
		pObj0 = Abc_NtkPi(pNtk, i)
		pObj1 = Abc_NtkPi(pNtk, i+nVars)
		pMiter = Abc_ObjReadCopyEdge(pObj0)
		Abc_ObjSetCopyEdge(pObj0, abcCompatCopyEdge(Abc_ObjReadCopyEdge(pObj1)))
		Abc_ObjSetCopyEdge(pObj1, abcCompatCopyEdge(pMiter))
	}
	// restrash
	// inlining of macro Abc_NtkForEachNode
	for i = int32(0); i < Abc_NtkObjNumMax(pNtk); i++ {
		pObj = Abc_NtkObj(pNtk, i)
		if pObj != nil && (Abc_ObjIsNode(pObj) != 0) {
			Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(baseabc.Abc_AigAnd(pNtkNew.FunctionManager.(*baseabc.Abc_Aig_t), Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj))))
		}
	}
	// add the PO
	pMiter = Abc_ObjChild0Copy(Abc_NtkPo(pNtk, 0))
	Abc_ObjAddFanin(Abc_NtkPo(pNtkNew, 0), pMiter)
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Performs reachability analisys.]

	Description [Assumes that the input is the transition relation.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkReachability(pNtkRel *Abc_Ntk_t, nIters int32, fVerbose int32) *Abc_Ntk_t {
	var pObj *Abc_Obj_t
	var pNtkFront, pNtkReached, pNtkNext, pNtkTemp *Abc_Ntk_t
	var i, v, nVars, nNodesOld, nNodesNew, nNodesPrev int32
	var fFixedPoint int32
	var fSynthesis int32 = 1
	var fMoreEffort int32 = 1
	var clk abctime

	assert((Abc_NtkIsStrash(pNtkRel) != 0), "Abc_NtkIsStrash(pNtkRel)")
	assert(Abc_NtkLatchNum(pNtkRel) == 0, "Abc_NtkLatchNum(pNtkRel) == 0")
	assert(Abc_NtkPiNum(pNtkRel)%2 == 0, "Abc_NtkPiNum(pNtkRel) % 2 == 0")

	// compute the network composed of the initial states
	pNtkFront = Abc_NtkInitialState(pNtkRel)
	pNtkReached = Abc_NtkDup(pNtkFront)
	//Abc_NtkShow( pNtkReached, 0, 0, 0 );

	//    if ( fVerbose )
	//        printf( "Transition relation = %6d.\n", Abc_NtkNodeNum(pNtkRel) );

	// perform iterations of reachability analysis
	nNodesPrev = Abc_NtkNodeNum(pNtkFront)
	nVars = Abc_NtkPiNum(pNtkRel) / 2
	for i = 0; i < nIters; i++ {
		clk = Abc_Clock()
		// get the set of next states
		pNtkNext = Abc_NtkMiterAnd(pNtkRel, pNtkFront, 0, 0)
		Abc_NtkDelete(pNtkFront)
		// quantify the current state variables
		for v = 0; v < nVars; v++ {
			Abc_NtkQuantify(pNtkNext, 0, v, fVerbose)
			if fSynthesis != 0 && v%3 == 2 {

				baseabc.Abc_AigCleanup(pNtkNext.FunctionManager.(*baseabc.Abc_Aig_t))
				Abc_NtkSynthesize(&pNtkNext, fMoreEffort)
			}
		}

		baseabc.Abc_AigCleanup(pNtkNext.FunctionManager.(*baseabc.Abc_Aig_t))
		if fSynthesis != 0 {
			Abc_NtkSynthesize(&pNtkNext, 1)
		}
		// map the next states into the current states
		pNtkTemp = pNtkNext
		pNtkNext = Abc_NtkSwapVariables(pNtkTemp)
		Abc_NtkDelete(pNtkTemp)
		// check the termination condition
		if Abc_ObjFanin0(Abc_NtkPo(pNtkNext, 0)) == baseabc.Abc_AigConst1(pNtkNext).Object {
			fFixedPoint = 1
			// A constant-one image reaches every state. The C source stops
			// before merging it, incorrectly returning the old reached set.
			if !Abc_NtkPo(pNtkNext, 0).Fanins()[0].Complement {
				Abc_NtkDelete(pNtkReached)
				pNtkReached = Abc_NtkDup(pNtkNext)
			}
			fmt.Printf("Fixed point is reached!\n")
			Abc_NtkDelete(pNtkNext)
			break
		}
		// compute new front
		pNtkFront = Abc_NtkMiterAnd(pNtkNext, pNtkReached, 0, 1)
		Abc_NtkDelete(pNtkNext)
		// add the reached states
		pNtkTemp = pNtkReached
		pNtkReached = Abc_NtkMiterAnd(pNtkTemp, pNtkFront, 1, 0)
		Abc_NtkDelete(pNtkTemp)
		// compress the size of Front
		nNodesOld = Abc_NtkNodeNum(pNtkFront)
		if fSynthesis != 0 {
			Abc_NtkSynthesize(&pNtkFront, fMoreEffort)
			Abc_NtkSynthesize(&pNtkReached, fMoreEffort)
		}
		nNodesNew = Abc_NtkNodeNum(pNtkFront)
		// print statistics
		if fVerbose != 0 {
			fmt.Printf("I = %3d : Reach = %6d  Fr = %6d  FrM = %6d  %7.2f %%   ", i+1, Abc_NtkNodeNum(pNtkReached), nNodesOld, nNodesNew, 100.0*float64(nNodesNew-nNodesPrev)/float64(nNodesPrev))
			// inlining of macro ABC_PRT
			Abc_Print(1, "%s =", ("T"))
			Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))))
		}
		nNodesPrev = Abc_NtkNodeNum(pNtkFront)
	}
	if !(fFixedPoint != 0) {
		fmt.Fprintf(os.Stdout, "Reachability analysis stopped after %d iterations.\n", nIters)
	}

	// complement the output to represent the set of unreachable states
	baseabc.Abc_ObjXorFaninC(Abc_NtkPo(pNtkReached, 0), 0)

	// remove next state variables
	for i = 2*nVars - 1; i >= nVars; i-- {
		pObj = Abc_NtkPi(pNtkReached, i)
		assert(baseabc.Abc_ObjFanoutNum(pObj) == 0, "Abc_ObjFanoutNum(pObj) == 0")
		baseabc.Abc_NtkDeleteObj(pObj)
	}

	// check consistency of the network
	if baseabc.Abc_NtkCheck(pNtkReached) == 0 {
		fmt.Printf("Abc_NtkReachability: The network check has failed.\n")
		Abc_NtkDelete(pNtkReached)
		return nil
	}
	return pNtkReached
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
