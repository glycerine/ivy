package abci

import aig_gia "github.com/glycerine/abcgo/aig/gia"

import base_abc "github.com/glycerine/abcgo/base/abc"

import misc_extra "github.com/glycerine/abcgo/misc/extra"

import misc_util "github.com/glycerine/abcgo/misc/util"

import sat_bsat "github.com/glycerine/abcgo/sat/bsat"

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [abcMiter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures to derive the miter of two circuits.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMiter.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/io/ioAbc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Abc_Ntk_t * Abc_NtkMiterInt( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int fComb, int nPartSize, int fImplic, int fMulti );

// static void        Abc_NtkMiterPrepare( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Abc_Ntk_t * pNtkMiter, int fComb, int nPartSize, int fMulti );

// static void        Abc_NtkMiterAddOne( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkMiter );

// static void        Abc_NtkMiterFinalize( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Abc_Ntk_t * pNtkMiter, int fComb, int nPartSize, int fImplic, int fMulti );

// static void        Abc_NtkAddFrame( Abc_Ntk_t * pNetNew, Abc_Ntk_t * pNet, int iFrame );

// to be exported
type AddFrameMapping func(*Abc_Obj_t, *Abc_Obj_t, int32, any)

// extern Abc_Ntk_t * Abc_NtkFrames2( Abc_Ntk_t * pNtk, int nFrames, int fInitial, AddFrameMapping addFrameMapping, void* arg );

// static void        Abc_NtkAddFrame2( Abc_Ntk_t * pNtkFrames, Abc_Ntk_t * pNtk, int iFrame, Vec_Ptr_t * vNodes, AddFrameMapping addFrameMapping, void* arg );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Derives the miter of two networks.]

	Description [Preprocesses the networks to make sure that they are strashed.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiter(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, fComb int32, nPartSize int32, fImplic int32, fMulti int32) *Abc_Ntk_t {
	var pTemp *Abc_Ntk_t = nil
	var fRemove1, fRemove2 bool
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk1) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk1)")
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk2) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk2)")
	// check that the networks have the same PIs/POs/latches
	if base_abc.Abc_NtkCompareSignals(pNtk1, pNtk2, fImplic, fComb) == 0 {
		return nil
	}
	// make sure the circuits are strashed
	if (Abc_NtkIsStrash(pNtk1) == 0) || Abc_NtkGetChoiceNum(pNtk1) != 0 {
		pNtk1 = Abc_NtkStrash(pNtk1, 0, 0, 0)
		fRemove1 = pNtk1 != nil
	}
	if (Abc_NtkIsStrash(pNtk2) == 0) || Abc_NtkGetChoiceNum(pNtk2) != 0 {
		pNtk2 = Abc_NtkStrash(pNtk2, 0, 0, 0)
		fRemove2 = pNtk2 != nil
	}
	if pNtk1 != nil && pNtk2 != nil {
		pTemp = Abc_NtkMiterInt(pNtk1, pNtk2, fComb, nPartSize, fImplic, fMulti)
	}
	if fRemove1 {
		Abc_NtkDelete(pNtk1)
	}
	if fRemove2 {
		Abc_NtkDelete(pNtk2)
	}
	return pTemp
}

/*
*Function*************************************************************

	Synopsis    [Derives the miter of two sequential networks.]

	Description [Assumes that the networks are strashed.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterInt(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, fComb int32, nPartSize int32, fImplic int32, fMulti int32) *Abc_Ntk_t {
	var Buffer string // was [1000]byte in C

	var pNtkMiter *Abc_Ntk_t

	assert((Abc_NtkIsStrash(pNtk1) != 0), "Abc_NtkIsStrash(pNtk1)")
	assert((Abc_NtkIsStrash(pNtk2) != 0), "Abc_NtkIsStrash(pNtk2)")

	// start the new network
	pNtkMiter = base_abc.Abc_NtkAlloc(base_abc.ABC_NTK_STRASH, base_abc.ABC_FUNC_AIG, 1)
	Buffer = fmt.Sprintf("%s_%s_miter", Abc_NtkName(pNtk1), Abc_NtkName(pNtk2))
	Abc_NtkSetName(pNtkMiter, Extra_UtilStrsav(Buffer))

	// perform strashing
	Abc_NtkMiterPrepare(pNtk1, pNtk2, pNtkMiter, fComb, nPartSize, fMulti)
	Abc_NtkMiterAddOne(pNtk1, pNtkMiter)
	Abc_NtkMiterAddOne(pNtk2, pNtkMiter)
	Abc_NtkMiterFinalize(pNtk1, pNtk2, pNtkMiter, fComb, nPartSize, fImplic, fMulti)
	base_abc.Abc_AigCleanup(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t))

	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtkMiter) == 0 {
		fmt.Printf("Abc_NtkMiter: The network check has failed.\n")
		Abc_NtkDelete(pNtkMiter)
		return nil
	}
	return pNtkMiter
}

/*
*Function*************************************************************

	Synopsis    [Prepares the network for mitering.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterPrepare(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pNtkMiter *Abc_Ntk_t, fComb int32, nPartSize int32, fMulti int32) {
	var pObj, pObjNew *Abc_Obj_t
	var i int32
	// clean the copy field in all objects
	//    Abc_NtkCleanCopy( pNtk1 );
	//    Abc_NtkCleanCopy( pNtk2 );
	Abc_ObjSetCopyEdge(base_abc.Abc_AigConst1(pNtk1).Object, abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkMiter)))
	Abc_ObjSetCopyEdge(base_abc.Abc_AigConst1(pNtk2).Object, abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkMiter)))

	if fComb != 0 {
		// create new PIs and remember them in the old PIs
		// inlining of macro Abc_NtkForEachCi
		for i = 0; i < Abc_NtkCiNum(pNtk1); i++ {
			pObj = base_abc.Abc_NtkCi(pNtk1, i)
			pObjNew = base_abc.Abc_NtkCreatePi(pNtkMiter)
			// remember this PI in the old PIs
			Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(pObjNew))
			pObj = base_abc.Abc_NtkCi(pNtk2, i)
			Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(pObjNew))
			// add name
			base_abc.Abc_ObjAssignName(pObjNew, Abc_ObjName(pObj), "")
		}
		if nPartSize <= 0 {
			// create POs
			if fMulti != 0 {
				// inlining of macro Abc_NtkForEachCo
				for i = 0; i < Abc_NtkCoNum(pNtk1); i++ {
					pObj = base_abc.Abc_NtkCo(pNtk1, i)
					pObjNew = base_abc.Abc_NtkCreatePo(pNtkMiter)
					base_abc.Abc_ObjAssignName(pObjNew, "miter", Abc_ObjName(pObjNew))
				}

			} else {
				pObjNew = base_abc.Abc_NtkCreatePo(pNtkMiter)
				base_abc.Abc_ObjAssignName(pObjNew, "miter", "")
			}
		}

	} else {
		// create new PIs and remember them in the old PIs
		// inlining of macro Abc_NtkForEachPi
		for i = 0; i < Abc_NtkPiNum(pNtk1); i++ {
			pObj = Abc_NtkPi(pNtk1, i)
			pObjNew = base_abc.Abc_NtkCreatePi(pNtkMiter)
			// remember this PI in the old PIs
			Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(pObjNew))
			pObj = Abc_NtkPi(pNtk2, i)
			Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(pObjNew))
			// add name
			base_abc.Abc_ObjAssignName(pObjNew, Abc_ObjName(pObj), "")
		}
		if nPartSize <= 0 {
			// create POs
			if fMulti != 0 {
				// inlining of macro Abc_NtkForEachPo
				for i = 0; i < Abc_NtkPoNum(pNtk1); i++ {
					pObj = Abc_NtkPo(pNtk1, i)
					pObjNew = base_abc.Abc_NtkCreatePo(pNtkMiter)
					base_abc.Abc_ObjAssignName(pObjNew, "miter", Abc_ObjName(pObjNew))
				}

			} else {
				pObjNew = base_abc.Abc_NtkCreatePo(pNtkMiter)
				base_abc.Abc_ObjAssignName(pObjNew, "miter", "")
			}
		}
		// create the latches
		// inlining of macro Abc_NtkForEachLatch
		for i = 0; i < int32(len(pNtk1.Boxes())); i++ {
			pObj = base_abc.Abc_NtkBox(pNtk1, i)
			if base_abc.Abc_ObjIsLatch(pObj) == 0 {
				continue
			}
			pObjNew = base_abc.Abc_NtkDupBox(pNtkMiter, pObj, 0)
			// add names
			base_abc.Abc_ObjAssignName(pObjNew, Abc_ObjName(pObj), "_1")
			base_abc.Abc_ObjAssignName(Abc_ObjFanin0(pObjNew), Abc_ObjName(Abc_ObjFanin0(pObj)), "_1")
			base_abc.Abc_ObjAssignName(Abc_ObjFanout0(pObjNew), Abc_ObjName(Abc_ObjFanout0(pObj)), "_1")
		}
		// inlining of macro Abc_NtkForEachLatch
		for i = 0; i < int32(len(pNtk2.Boxes())); i++ {
			pObj = base_abc.Abc_NtkBox(pNtk2, i)
			if base_abc.Abc_ObjIsLatch(pObj) == 0 {
				continue
			}
			pObjNew = base_abc.Abc_NtkDupBox(pNtkMiter, pObj, 0)
			// add name
			base_abc.Abc_ObjAssignName(pObjNew, Abc_ObjName(pObj), "_2")
			base_abc.Abc_ObjAssignName(Abc_ObjFanin0(pObjNew), Abc_ObjName(Abc_ObjFanin0(pObj)), "_2")
			base_abc.Abc_ObjAssignName(Abc_ObjFanout0(pObjNew), Abc_ObjName(Abc_ObjFanout0(pObj)), "_2")
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs mitering for one network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterAddOne(pNtk *Abc_Ntk_t, pNtkMiter *Abc_Ntk_t) {
	var pNode *Abc_Obj_t
	var i int32
	assert((base_abc.Abc_NtkIsDfsOrdered(pNtk) != 0), "Abc_NtkIsDfsOrdered(pNtk)")
	// inlining of macro Abc_AigForEachAnd
	for i = 0; i < Abc_NtkObjNumMax(pNtk); i++ {
		pNode = Abc_NtkObj(pNtk, i)
		if pNode == nil || (base_abc.Abc_AigNodeIsAnd(pNode) == 0) {
			continue
		}
		Abc_ObjSetCopyEdge(pNode, abcCompatCopyEdge(base_abc.Abc_AigAnd(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode))))
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs mitering for one network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterAddCone(pNtk *Abc_Ntk_t, pNtkMiter *Abc_Ntk_t, pRoot *Abc_Obj_t) {
	var vNodes []*Abc_Obj_t
	var pNode *Abc_Obj_t
	var i int32
	// map the constant nodes
	Abc_ObjSetCopyEdge(base_abc.Abc_AigConst1(pNtk).Object, abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkMiter)))
	// perform strashing
	vNodes = base_abc.Abc_NtkDfsNodes(pNtk, []*Abc_Obj_t{pRoot}, 1)
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(len(vNodes)); i++ {
		pNode = vNodes[i]
		if base_abc.Abc_AigNodeIsAnd(pNode) != 0 {
			Abc_ObjSetCopyEdge(pNode, abcCompatCopyEdge(base_abc.Abc_AigAnd(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode))))
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Finalizes the miter by adding the output part.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterFinalize(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pNtkMiter *Abc_Ntk_t, fComb int32, nPartSize int32, fImplic int32, fMulti int32) {
	var vPairs []base_abc.Edge
	var pNode *Abc_Obj_t
	var pMiter base_abc.Edge
	var i int32
	assert(nPartSize == 0 || fMulti == 0, "nPartSize == 0 || fMulti == 0")
	// collect the PO pairs from both networks
	vPairs = make([]base_abc.Edge, 0, 100)
	if fComb != 0 {
		// collect the CO nodes for the miter
		// inlining of macro Abc_NtkForEachCo
		for i = 0; i < Abc_NtkCoNum(pNtk1); i++ {
			pNode = base_abc.Abc_NtkCo(pNtk1, i)
			if fMulti != 0 {
				pMiter = base_abc.Abc_AigXor(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), Abc_ObjChild0Copy(pNode), Abc_ObjChild0Copy(base_abc.Abc_NtkCo(pNtk2, i)))
				Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, i), pMiter)

			} else {
				vPairs = append(vPairs, Abc_ObjChild0Copy(pNode))
				pNode = base_abc.Abc_NtkCo(pNtk2, i)
				vPairs = append(vPairs, Abc_ObjChild0Copy(pNode))
			}
		}

	} else {
		// collect the PO nodes for the miter
		// inlining of macro Abc_NtkForEachPo
		for i = 0; i < Abc_NtkPoNum(pNtk1); i++ {
			pNode = Abc_NtkPo(pNtk1, i)
			if fMulti != 0 {
				pMiter = base_abc.Abc_AigXor(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), Abc_ObjChild0Copy(pNode), Abc_ObjChild0Copy(base_abc.Abc_NtkCo(pNtk2, i)))
				Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, i), pMiter)

			} else {
				vPairs = append(vPairs, Abc_ObjChild0Copy(pNode))
				pNode = Abc_NtkPo(pNtk2, i)
				vPairs = append(vPairs, Abc_ObjChild0Copy(pNode))
			}
		}
		// connect new latches
		// inlining of macro Abc_NtkForEachLatch
		for i = 0; i < int32(len(pNtk1.Boxes())); i++ {
			pNode = base_abc.Abc_NtkBox(pNtk1, i)
			if base_abc.Abc_ObjIsLatch(pNode) != 0 {
				Abc_ObjAddFanin(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pNode)).Object, Abc_ObjChild0Copy(Abc_ObjFanin0(pNode)))
			}
		}
		// inlining of macro Abc_NtkForEachLatch
		for i = 0; i < int32(len(pNtk2.Boxes())); i++ {
			pNode = base_abc.Abc_NtkBox(pNtk2, i)
			if base_abc.Abc_ObjIsLatch(pNode) != 0 {
				Abc_ObjAddFanin(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pNode)).Object, Abc_ObjChild0Copy(Abc_ObjFanin0(pNode)))
			}
		}
	}
	// add the miter
	if nPartSize <= 0 {
		if !(fMulti != 0) {
			pMiter = base_abc.Abc_AigMiter(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), &vPairs, fImplic)
			Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, 0), pMiter)
		}

	} else {
		var Buffer string // was [1024]byte in C

		var vPairsPart []base_abc.Edge
		var nParts, i, k, iCur int32
		assert(int32(len(vPairs)) == 2*Abc_NtkCoNum(pNtk1), "int32(len(vPairs)) == 2 * Abc_NtkCoNum(pNtk1)")
		// create partitions
		nParts = Abc_NtkCoNum(pNtk1) / nPartSize
		if Abc_NtkCoNum(pNtk1)%nPartSize > 0 {
			nParts++
		}
		vPairsPart = make([]base_abc.Edge, 0, 2*nPartSize)
		for i = 0; i < nParts; i++ {
			vPairsPart = vPairsPart[:0]
			for k = 0; k < nPartSize; k++ {
				iCur = i*nPartSize + k
				if iCur >= Abc_NtkCoNum(pNtk1) {
					break
				}
				vPairsPart = append(vPairsPart, vPairs[2*iCur])
				vPairsPart = append(vPairsPart, vPairs[2*iCur+1])
			}
			pMiter = base_abc.Abc_AigMiter(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), &vPairsPart, fImplic)
			pNode = base_abc.Abc_NtkCreatePo(pNtkMiter)
			Abc_ObjAddFanin(pNode, pMiter)
			// assign the name to the node
			if nPartSize == 1 {
				Buffer = fmt.Sprintf("%s", Abc_ObjName(base_abc.Abc_NtkCo(pNtk1, i)))
			} else {
				Buffer = fmt.Sprintf("%d", i)
			}
			base_abc.Abc_ObjAssignName(pNode, "miter_", Buffer)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Derives the AND of two miters.]

	Description [The network should have the same names of PIs.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterAnd(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, fOr int32, fCompl2 int32) *Abc_Ntk_t {
	var Buffer string // was [1000]byte in C

	var pNtkMiter *Abc_Ntk_t
	var pOutput1, pOutput2 base_abc.Edge
	var pRoot1, pRoot2 *Abc_Obj_t
	var pMiter base_abc.Edge

	assert((Abc_NtkIsStrash(pNtk1) != 0), "Abc_NtkIsStrash(pNtk1)")
	assert((Abc_NtkIsStrash(pNtk2) != 0), "Abc_NtkIsStrash(pNtk2)")
	assert(1 == Abc_NtkCoNum(pNtk1), "1 == Abc_NtkCoNum(pNtk1)")
	assert(1 == Abc_NtkCoNum(pNtk2), "1 == Abc_NtkCoNum(pNtk2)")
	assert(0 == Abc_NtkLatchNum(pNtk1), "0 == Abc_NtkLatchNum(pNtk1)")
	assert(0 == Abc_NtkLatchNum(pNtk2), "0 == Abc_NtkLatchNum(pNtk2)")
	assert(Abc_NtkCiNum(pNtk1) == Abc_NtkCiNum(pNtk2), "Abc_NtkCiNum(pNtk1) == Abc_NtkCiNum(pNtk2)")
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk1) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk1)")
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk2) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk2)")

	// start the new network
	pNtkMiter = base_abc.Abc_NtkAlloc(base_abc.ABC_NTK_STRASH, base_abc.ABC_FUNC_AIG, 1)
	//    sprintf( Buffer, "%s_%s_miter", pNtk1->pName, pNtk2->pName );
	Buffer = fmt.Sprintf("product")
	Abc_NtkSetName(pNtkMiter, Extra_UtilStrsav(Buffer))

	// perform strashing
	Abc_NtkMiterPrepare(pNtk1, pNtk2, pNtkMiter, 1, -1, 0)
	Abc_NtkMiterAddOne(pNtk1, pNtkMiter)
	Abc_NtkMiterAddOne(pNtk2, pNtkMiter)
	//    Abc_NtkMiterFinalize( pNtk1, pNtk2, pNtkMiter, 1 );
	pRoot1 = Abc_NtkPo(pNtk1, 0)
	pRoot2 = Abc_NtkPo(pNtk2, 0)
	pOutput1 = base_abc.Abc_ObjNotCond(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pRoot1)), base_abc.Abc_ObjFaninC0(pRoot1))
	pOutput2 = base_abc.Abc_ObjNotCond(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pRoot2)), int32(base_abc.Abc_ObjFaninC0(pRoot2))^fCompl2)

	// create the miter of the two outputs
	if fOr != 0 {
		pMiter = base_abc.Abc_AigOr(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), pOutput1, pOutput2)
	} else {
		pMiter = base_abc.Abc_AigAnd(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), pOutput1, pOutput2)
	}
	Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, 0), pMiter)

	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtkMiter) == 0 {
		fmt.Printf("Abc_NtkMiterAnd: The network check has failed.\n")
		Abc_NtkDelete(pNtkMiter)
		return nil
	}
	return pNtkMiter
}

/*
*Function*************************************************************

	Synopsis    [Derives the cofactor of the miter w.r.t. the set of vars.]

	Description [The array of variable values contains -1/0/1 for each PI.
	-1 means this PI remains, 0/1 means this PI is set to 0/1.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterCofactor(pNtk *Abc_Ntk_t, vPiValues *Vec_Int_t) *Abc_Ntk_t {
	var Buffer string // was [1000]byte in C

	var pNtkMiter *Abc_Ntk_t
	var pRoot *Abc_Obj_t
	var pOutput1 base_abc.Edge
	var Value, i int32

	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert(1 == Abc_NtkCoNum(pNtk), "1 == Abc_NtkCoNum(pNtk)")
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk)")

	// start the new network
	pNtkMiter = base_abc.Abc_NtkAlloc(base_abc.ABC_NTK_STRASH, base_abc.ABC_FUNC_AIG, 1)
	Buffer = fmt.Sprintf("%s_miter", Abc_NtkName(pNtk))
	Abc_NtkSetName(pNtkMiter, Extra_UtilStrsav(Buffer))

	// get the root output
	pRoot = base_abc.Abc_NtkCo(pNtk, 0)

	// perform strashing
	Abc_NtkMiterPrepare(pNtk, pNtk, pNtkMiter, 1, -1, 0)
	// set the first cofactor
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(vPiValues); i++ {
		Value = Vec_IntEntry(vPiValues, i)
		if Value == -1 {
			continue
		}
		if Value == 0 {
			Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, i), abcCompatCopyEdge(base_abc.Abc_ObjNot(base_abc.Abc_AigConst1(pNtkMiter))))
			continue
		}
		if Value == 1 {
			Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, i), abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkMiter)))
			continue
		}
		assert(false, "0")
	}
	// add the first cofactor
	Abc_NtkMiterAddCone(pNtk, pNtkMiter, pRoot)

	// save the output
	pOutput1 = base_abc.Abc_ObjNotCond(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pRoot)), base_abc.Abc_ObjFaninC0(pRoot))

	// create the miter of the two outputs
	Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, 0), pOutput1)

	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtkMiter) == 0 {
		fmt.Printf("Abc_NtkMiterCofactor: The network check has failed.\n")
		Abc_NtkDelete(pNtkMiter)
		return nil
	}
	return pNtkMiter
}

/*
*Function*************************************************************

	Synopsis    [Derives the miter of two cofactors of one output.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterForCofactors(pNtk *Abc_Ntk_t, Out int32, In1 int32, In2 int32) *Abc_Ntk_t {
	var Buffer string // was [1000]byte in C

	var pNtkMiter *Abc_Ntk_t
	var pRoot *Abc_Obj_t
	var pOutput1, pOutput2, pMiter base_abc.Edge

	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert(Out < Abc_NtkCoNum(pNtk), "Out < Abc_NtkCoNum(pNtk)")
	assert(In1 < Abc_NtkCiNum(pNtk), "In1 < Abc_NtkCiNum(pNtk)")
	assert(In2 < Abc_NtkCiNum(pNtk), "In2 < Abc_NtkCiNum(pNtk)")
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk)")

	// start the new network
	pNtkMiter = base_abc.Abc_NtkAlloc(base_abc.ABC_NTK_STRASH, base_abc.ABC_FUNC_AIG, 1)
	Buffer = fmt.Sprintf("%s_miter", Abc_ObjName(base_abc.Abc_NtkCo(pNtk, Out)))
	Abc_NtkSetName(pNtkMiter, Extra_UtilStrsav(Buffer))

	// get the root output
	pRoot = base_abc.Abc_NtkCo(pNtk, Out)

	// perform strashing
	Abc_NtkMiterPrepare(pNtk, pNtk, pNtkMiter, 1, -1, 0)
	// set the first cofactor
	Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, In1), abcCompatCopyEdge(base_abc.Abc_ObjNot(base_abc.Abc_AigConst1(pNtkMiter))))
	if In2 >= 0 {
		Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, In2), abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkMiter)))
	}
	// add the first cofactor
	Abc_NtkMiterAddCone(pNtk, pNtkMiter, pRoot)

	// save the output
	pOutput1 = Abc_ObjReadCopyEdge(Abc_ObjFanin0(pRoot))

	// set the second cofactor
	Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, In1), abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkMiter)))
	if In2 >= 0 {
		Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, In2), abcCompatCopyEdge(base_abc.Abc_ObjNot(base_abc.Abc_AigConst1(pNtkMiter))))
	}
	// add the second cofactor
	Abc_NtkMiterAddCone(pNtk, pNtkMiter, pRoot)

	// save the output
	pOutput2 = Abc_ObjReadCopyEdge(Abc_ObjFanin0(pRoot))

	// create the miter of the two outputs
	pMiter = base_abc.Abc_AigXor(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), pOutput1, pOutput2)
	Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, 0), pMiter)

	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtkMiter) == 0 {
		fmt.Printf("Abc_NtkMiter: The network check has failed.\n")
		Abc_NtkDelete(pNtkMiter)
		return nil
	}
	return pNtkMiter
}

/*
*Function*************************************************************

	Synopsis    [Derives the miter of two cofactors of one output.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterQuantify(pNtk *Abc_Ntk_t, In int32, fExist int32) *Abc_Ntk_t {
	var pNtkMiter *Abc_Ntk_t
	var pRoot *Abc_Obj_t
	var pOutput1, pOutput2, pMiter base_abc.Edge

	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert(1 == Abc_NtkCoNum(pNtk), "1 == Abc_NtkCoNum(pNtk)")
	assert(In < Abc_NtkCiNum(pNtk), "In < Abc_NtkCiNum(pNtk)")
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk)")

	// start the new network
	pNtkMiter = base_abc.Abc_NtkAlloc(base_abc.ABC_NTK_STRASH, base_abc.ABC_FUNC_AIG, 1)
	Abc_NtkSetName(pNtkMiter, Extra_UtilStrsav(Abc_ObjName(base_abc.Abc_NtkCo(pNtk, 0))))

	// get the root output
	pRoot = base_abc.Abc_NtkCo(pNtk, 0)

	// perform strashing
	Abc_NtkMiterPrepare(pNtk, pNtk, pNtkMiter, 1, -1, 0)
	// set the first cofactor
	Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, In), abcCompatCopyEdge(base_abc.Abc_ObjNot(base_abc.Abc_AigConst1(pNtkMiter))))
	// add the first cofactor
	Abc_NtkMiterAddCone(pNtk, pNtkMiter, pRoot)
	// save the output
	//    pOutput1 = Abc_ObjFanin0(pRoot)->pCopy;
	pOutput1 = base_abc.Abc_ObjNotCond(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pRoot)), base_abc.Abc_ObjFaninC0(pRoot))

	// set the second cofactor
	Abc_ObjSetCopyEdge(base_abc.Abc_NtkCi(pNtk, In), abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkMiter)))
	// add the second cofactor
	Abc_NtkMiterAddCone(pNtk, pNtkMiter, pRoot)
	// save the output
	//    pOutput2 = Abc_ObjFanin0(pRoot)->pCopy;
	pOutput2 = base_abc.Abc_ObjNotCond(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pRoot)), base_abc.Abc_ObjFaninC0(pRoot))

	// create the miter of the two outputs
	if fExist != 0 {
		pMiter = base_abc.Abc_AigOr(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), pOutput1, pOutput2)
	} else {
		pMiter = base_abc.Abc_AigAnd(pNtkMiter.FunctionManager.(*base_abc.Abc_Aig_t), pOutput1, pOutput2)
	}
	Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, 0), pMiter)

	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtkMiter) == 0 {
		fmt.Printf("Abc_NtkMiter: The network check has failed.\n")
		Abc_NtkDelete(pNtkMiter)
		return nil
	}
	return pNtkMiter
}

/*
*Function*************************************************************

	Synopsis    [Quantifies all the PIs existentially from the only PO of the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterQuantifyPis(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	var pNtkTemp *Abc_Ntk_t
	var pObj *Abc_Obj_t
	var i int32
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk)")

	// inlining of macro Abc_NtkForEachPi
	for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
		pObj = Abc_NtkPi(pNtk, i)
		if base_abc.Abc_ObjFanoutNum(pObj) == 0 {
			continue
		}
		pNtkTemp = pNtk
		pNtk = Abc_NtkMiterQuantify(pNtkTemp, i, 1)
		Abc_NtkDelete(pNtkTemp)
	}

	return pNtk
}

/*
*Function*************************************************************

	Synopsis    [Checks the status of the miter.]

	Description [Return 0 if the miter is sat for at least one output.
	Return 1 if the miter is unsat for all its outputs. Returns -1 if the
	miter is undecided for some outputs.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterIsConstant(pMiter *Abc_Ntk_t) int32 {
	var pNodePo *Abc_Obj_t
	var pChild base_abc.Edge
	var i int32
	assert((Abc_NtkIsStrash(pMiter) != 0), "Abc_NtkIsStrash(pMiter)")
	// inlining of macro Abc_NtkForEachPo
	for i = 0; i < Abc_NtkPoNum(pMiter); i++ {
		pNodePo = Abc_NtkPo(pMiter, i)
		pChild = Abc_ObjChild0(pNodePo)
		// check if the output is constant 1
		if base_abc.Abc_AigNodeIsConst(pChild.Object) != 0 {
			assert(base_abc.Abc_ObjRegular(pChild) == base_abc.Abc_AigConst1(pMiter).Object, "Abc_ObjRegular(pChild) == Abc_AigConst1(pMiter)")
			if base_abc.Abc_ObjIsComplement(pChild) == 0 {
				// if the miter is constant 1, return immediately
				//                printf( "MITER IS CONSTANT 1!\n" );
				return 0
			}

			/*
			   // check if the output is not constant 0
			   else if ( Abc_ObjRegular(pChild)->fPhase != (unsigned)Abc_ObjIsComplement(pChild) )
			   {
			       return 0;
			   }
			*/
			// if the miter is undecided (or satisfiable), return immediately
		} else {
			return -1
		}
	}
	// return 1, meaning all outputs are constant zero
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Reports the status of the miter.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMiterReport(pMiter *Abc_Ntk_t) {
	var pChild base_abc.Edge
	var i int32
	if Abc_NtkPoNum(pMiter) == 1 {
		pChild = Abc_ObjChild0(Abc_NtkPo(pMiter, 0))
		if base_abc.Abc_AigNodeIsConst(pChild.Object) != 0 {
			if base_abc.Abc_ObjIsComplement(pChild) != 0 {
				fmt.Printf("Unsatisfiable.\n")
			} else {
				fmt.Printf("Satisfiable. (Constant 1).\n")
			}

		} else {
			fmt.Printf("Satisfiable.\n")
		}

	} else {
		// inlining of macro Abc_NtkForEachPo
		for i = 0; i < Abc_NtkPoNum(pMiter); i++ {
			pChild = Abc_ObjChild0(Abc_NtkPo(pMiter, i))
			fmt.Printf("Output #%2d : ", i)
			if base_abc.Abc_AigNodeIsConst(pChild.Object) != 0 {
				if base_abc.Abc_ObjIsComplement(pChild) != 0 {
					fmt.Printf("Unsatisfiable.\n")
				} else {
					fmt.Printf("Satisfiable. (Constant 1).\n")
				}

			} else {
				fmt.Printf("Satisfiable.\n")
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Derives the timeframes of the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFrames(pNtk *Abc_Ntk_t, nFrames int32, fInitial int32, fVerbose int32) *Abc_Ntk_t {
	var Buffer string // was [1000]byte in C

	var pProgress *ProgressBar
	var pNtkFrames *Abc_Ntk_t
	var pLatch, pLatchOut *Abc_Obj_t
	var i, Counter int32
	assert(nFrames > 0, "nFrames > 0")
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert((base_abc.Abc_NtkIsDfsOrdered(pNtk) != 0), "Abc_NtkIsDfsOrdered(pNtk)")
	assert((base_abc.Abc_NtkHasOnlyLatchBoxes(pNtk) != 0), "Abc_NtkHasOnlyLatchBoxes(pNtk)")
	// start the new network
	pNtkFrames = base_abc.Abc_NtkAlloc(base_abc.ABC_NTK_STRASH, base_abc.ABC_FUNC_AIG, 1)
	Buffer = fmt.Sprintf("%s_%d_frames", Abc_NtkName(pNtk), nFrames)
	Abc_NtkSetName(pNtkFrames, Extra_UtilStrsav(Buffer))
	// map the constant nodes
	Abc_ObjSetCopyEdge(base_abc.Abc_AigConst1(pNtk).Object, abcCompatCopyEdge(base_abc.Abc_AigConst1(pNtkFrames)))
	// create new latches (or their initial values) and remember them in the new latches
	if !(fInitial != 0) {
		// inlining of macro Abc_NtkForEachLatch
		for i = 0; i < int32(len(pNtk.Boxes())); i++ {
			pLatch = base_abc.Abc_NtkBox(pNtk, i)
			if base_abc.Abc_ObjIsLatch(pLatch) != 0 {
				base_abc.Abc_NtkDupBox(pNtkFrames, pLatch, 1)
			}
		}

	} else {
		Counter = 0
		// inlining of macro Abc_NtkForEachLatch
		for i = 0; i < int32(len(pNtk.Boxes())); i++ {
			pLatch = base_abc.Abc_NtkBox(pNtk, i)
			if base_abc.Abc_ObjIsLatch(pLatch) != 0 {
				pLatchOut = Abc_ObjFanout0(pLatch)
				if (base_abc.Abc_LatchIsInitNone(pLatch) != 0) || (base_abc.Abc_LatchIsInitDc(pLatch) != 0) { // don't-care initial value - create a new PI

					Abc_ObjSetCopyEdge(pLatchOut, abcCompatCopyEdge(base_abc.Abc_NtkCreatePi(pNtkFrames)))
					base_abc.Abc_ObjAssignName(Abc_ObjReadCopyEdge(pLatchOut).Object, Abc_ObjName(pLatchOut), "")
					Counter++

				} else {
					Abc_ObjSetCopyEdge(pLatchOut, abcCompatCopyEdge(base_abc.Abc_ObjNotCond(base_abc.Abc_AigConst1(pNtkFrames), base_abc.Abc_LatchIsInit0(pLatch))))
				}
			}
		}
		if Counter != 0 {
			fmt.Printf("Warning: %d uninitialized latches are replaced by free PI variables.\n", Counter)
		}
	}

	// create the timeframes
	pProgress = misc_extra.Extra_ProgressBarStart(os.Stdout, nFrames)
	for i = 0; i < nFrames; i++ {
		misc_extra.Extra_ProgressBarUpdate(pProgress, i, "")
		Abc_NtkAddFrame(pNtkFrames, pNtk, i)
	}
	misc_extra.Extra_ProgressBarStop(pProgress)

	// connect the new latches to the outputs of the last frame
	if !(fInitial != 0) {
		// we cannot use pLatch->pCopy here because pLatch->pCopy is used for temporary storage of strashed values
		// inlining of macro Abc_NtkForEachLatch
		for i = 0; i < int32(len(pNtk.Boxes())); i++ {
			pLatch = base_abc.Abc_NtkBox(pNtk, i)
			if base_abc.Abc_ObjIsLatch(pLatch) != 0 {
				Abc_ObjAddFanin(Abc_ObjReadCopyEdge(Abc_ObjFanin0(pLatch)).Object, Abc_ObjReadCopyEdge(Abc_ObjFanout0(pLatch)))
			}
		}
	}

	// remove dangling nodes
	base_abc.Abc_AigCleanup(pNtkFrames.FunctionManager.(*base_abc.Abc_Aig_t))
	// reorder the latches
	base_abc.Abc_NtkOrderCisCos(pNtkFrames)
	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtkFrames) == 0 {
		fmt.Printf("Abc_NtkFrames: The network check has failed.\n")
		Abc_NtkDelete(pNtkFrames)
		return nil
	}
	return pNtkFrames
}

/*
*Function*************************************************************

	Synopsis    [Adds one time frame to the new network.]

	Description [Assumes that the latches of the old network point
	to the outputs of the previous frame of the new network (pLatch->pCopy).
	In the end, updates the latches of the old network to point to the
	outputs of the current frame of the new network.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkAddFrame(pNtkFrames *Abc_Ntk_t, pNtk *Abc_Ntk_t, iFrame int32) {
	var fVerbose int32
	var NodeBef int32 = Abc_NtkNodeNum(pNtkFrames)
	var Buffer string // was [10]byte in C

	var pNode, pLatch *Abc_Obj_t
	var i int32
	// create the prefix to be added to the node names
	Buffer = fmt.Sprintf("_%02d", iFrame)
	// add the new PI nodes
	// inlining of macro Abc_NtkForEachPi
	for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
		pNode = Abc_NtkPi(pNtk, i)
		base_abc.Abc_ObjAssignName(base_abc.Abc_NtkDupObj(pNtkFrames, pNode, 0), Abc_ObjName(pNode), Buffer)
	}
	// add the internal nodes
	// inlining of macro Abc_AigForEachAnd
	for i = 0; i < Abc_NtkObjNumMax(pNtk); i++ {
		pNode = Abc_NtkObj(pNtk, i)
		if pNode == nil || (base_abc.Abc_AigNodeIsAnd(pNode) == 0) {
			continue
		}
		Abc_ObjSetCopyEdge(pNode, abcCompatCopyEdge(base_abc.Abc_AigAnd(pNtkFrames.FunctionManager.(*base_abc.Abc_Aig_t), Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode))))
	}
	// add the new POs
	// inlining of macro Abc_NtkForEachPo
	for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
		pNode = Abc_NtkPo(pNtk, i)
		base_abc.Abc_ObjAssignName(base_abc.Abc_NtkDupObj(pNtkFrames, pNode, 0), Abc_ObjName(pNode), Buffer)
		Abc_ObjAddFanin(Abc_ObjReadCopyEdge(pNode).Object, Abc_ObjChild0Copy(pNode))
	}
	// transfer the implementation of the latch inputs to the latch outputs
	// inlining of macro Abc_NtkForEachLatch
	for i = 0; i < int32(len(pNtk.Boxes())); i++ {
		pLatch = base_abc.Abc_NtkBox(pNtk, i)
		if base_abc.Abc_ObjIsLatch(pLatch) != 0 {
			Abc_ObjSetCopyEdge(pLatch, abcCompatCopyEdge(Abc_ObjChild0Copy(Abc_ObjFanin0(pLatch))))
		}
	}
	// inlining of macro Abc_NtkForEachLatch
	for i = 0; i < int32(len(pNtk.Boxes())); i++ {
		pLatch = base_abc.Abc_NtkBox(pNtk, i)
		if base_abc.Abc_ObjIsLatch(pLatch) != 0 {
			Abc_ObjSetCopyEdge(Abc_ObjFanout0(pLatch), Abc_ObjReadCopyEdge(pLatch))
		}
	}
	// nodes after
	if fVerbose != 0 {
		prop := "unknown"
		if base_abc.Abc_AigNodeIsConst(Abc_ObjReadCopyEdge(Abc_ObjFanin0(Abc_NtkPo(pNtk, 0))).Object) != 0 {
			prop = "proof"
		}
		fmt.Printf("F = %4d : Total = %6d. Nodes = %6d. Prop = %s.\n", iFrame, Abc_NtkNodeNum(pNtk), Abc_NtkNodeNum(pNtkFrames)-NodeBef, prop)
	}
}

/*
*Function*************************************************************

	Synopsis    [Derives the timeframes of the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFrames2(pNtk *Abc_Ntk_t, nFrames int32, fInitial int32, addFrameMapping AddFrameMapping, arg any) *Abc_Ntk_t {
	/*
	   char Buffer[1000];
	   ProgressBar * pProgress;
	   Abc_Ntk_t * pNtkFrames;
	   Vec_Ptr_t * vNodes;
	   Abc_Obj_t * pLatch, * pLatchNew;
	   int i, Counter;
	   assert( nFrames > 0 );
	   assert( Abc_NtkIsStrash(pNtk) );
	   // start the new network
	   pNtkFrames = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
	   sprintf( Buffer, "%s_%d_frames", pNtk->pName, nFrames );
	   pNtkFrames->pName = Extra_UtilStrsav(Buffer);
	   // create new latches (or their initial values) and remember them in the new latches
	   if ( !fInitial )
	   {
	       Abc_NtkForEachLatch( pNtk, pLatch, i ) {
	           Abc_NtkDupObj( pNtkFrames, pLatch );
	           if addFrameMapping != nil { addFrameMapping(pLatch->pCopy, pLatch, 0, arg);
	       }
	   }
	   else
	   {
	       Counter = 0;
	       Abc_NtkForEachLatch( pNtk, pLatch, i )
	       {
	           if ( Abc_LatchIsInitDc(pLatch) ) // don't-care initial value - create a new PI
	           {
	               pLatch->pCopy = Abc_NtkCreatePi(pNtkFrames);
	               Abc_ObjAssignName( pLatch->pCopy, Abc_ObjName(pLatch), NULL );
	               Counter++;
	           }
	           else {
	               pLatch->pCopy = Abc_ObjNotCond( Abc_AigConst1(pNtkFrames), Abc_LatchIsInit0(pLatch) );
	           }

	           if addFrameMapping != nil { addFrameMapping(pLatch->pCopy, pLatch, 0, arg);
	       }
	       if ( Counter )
	           printf( "Warning: %d uninitialized latches are replaced by free PI variables.\n", Counter );
	   }

	   // create the timeframes
	   vNodes = Abc_NtkDfs( pNtk, 0 );
	   pProgress = Extra_ProgressBarStart( stdout, nFrames );
	   for ( i = 0; i < nFrames; i++ )
	   {
	       Extra_ProgressBarUpdate( pProgress, i, NULL );
	       Abc_NtkAddFrame2( pNtkFrames, pNtk, i, vNodes, addFrameMapping, arg );
	   }
	   Extra_ProgressBarStop( pProgress );
	   Vec_PtrFree( vNodes );

	   // connect the new latches to the outputs of the last frame
	   if ( !fInitial )
	   {
	       Abc_NtkForEachLatch( pNtk, pLatch, i )
	       {
	           pLatchNew = Abc_NtkBox(pNtkFrames, i);
	           Abc_ObjAddFanin( pLatchNew, pLatch->pCopy );
	           Abc_ObjAssignName( pLatchNew, Abc_ObjName(pLatch), NULL );
	       }
	   }
	   Abc_NtkForEachLatch( pNtk, pLatch, i )
	       pLatch->pNext = NULL;

	   // remove dangling nodes
	   Abc_AigCleanup( pNtkFrames->pManFunc );

	   // reorder the latches
	   Abc_NtkOrderCisCos( pNtkFrames );

	   // make sure that everything is okay
	   if ( !Abc_NtkCheck( pNtkFrames ) )
	   {
	       printf( "Abc_NtkFrames: The network check has failed.\n" );
	       Abc_NtkDelete( pNtkFrames );
	       return NULL;
	   }
	   return pNtkFrames;
	*/
	return nil
}

/*
*Function*************************************************************

	Synopsis    [Adds one time frame to the new network.]

	Description [Assumes that the latches of the old network point
	to the outputs of the previous frame of the new network (pLatch->pCopy).
	In the end, updates the latches of the old network to point to the
	outputs of the current frame of the new network.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkAddFrame2(pNtkFrames *Abc_Ntk_t, pNtk *Abc_Ntk_t, iFrame int32, vNodes *Vec_Ptr_t, addFrameMapping AddFrameMapping, arg any) {
	/*
		char Buffer[10];
		Abc_Obj_t * pNode, * pNodeNew, * pLatch;
		Abc_Obj_t * pConst1, * pConst1New;
		int i;
		// get the constant nodes
		pConst1    = Abc_AigConst1(pNtk);
		pConst1New = Abc_AigConst1(pNtkFrames);
		// create the prefix to be added to the node names
		sprintf( Buffer, "_%02d", iFrame );
		// add the new PI nodes
		Abc_NtkForEachPi( pNtk, pNode, i )

			{
			pNodeNew = Abc_NtkDupObj( pNtkFrames, pNode );
			Abc_ObjAssignName( pNodeNew, Abc_ObjName(pNode), Buffer );
			if addFrameMapping != nil { addFrameMapping(pNodeNew, pNode, iFrame, arg);
			}

		// add the internal nodes
		Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pNode, i )

			{
			if ( pNode == pConst1 )
			pNodeNew = pConst1New;
			else
			pNodeNew = Abc_AigAnd( pNtkFrames->pManFunc, Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode) );
			pNode->pCopy = pNodeNew;
			if addFrameMapping != nil { addFrameMapping(pNodeNew, pNode, iFrame, arg);
			}

		// add the new POs
		Abc_NtkForEachPo( pNtk, pNode, i )

			{
			pNodeNew = Abc_NtkDupObj( pNtkFrames, pNode );
			Abc_ObjAddFanin( pNodeNew, Abc_ObjChild0Copy(pNode) );
			Abc_ObjAssignName( pNodeNew, Abc_ObjName(pNode), Buffer );
			if addFrameMapping != nil { addFrameMapping(pNodeNew, pNode, iFrame, arg);
			}

		// transfer the implementation of the latch drivers to the latches

		// it is important that these two steps are performed it two loops
		// and not in the same loop
		Abc_NtkForEachLatch( pNtk, pLatch, i )

			pLatch->pNext = Abc_ObjChild0Copy(pLatch);

		Abc_NtkForEachLatch( pNtk, pLatch, i )

			pLatch->pCopy = pLatch->pNext;

		Abc_NtkForEachLatch( pNtk, pLatch, i )

			{
			if addFrameMapping != nil {
			// don't give Mike complemented pointers because he doesn't like it
			if (Abc_ObjIsComplement(pLatch->pCopy)) {
			pNodeNew = Abc_NtkCreateNode( pNtkFrames );
			Abc_ObjAddFanin( pNodeNew, pLatch->pCopy );
			assert(Abc_ObjFaninNum(pNodeNew) == 1);
			pNodeNew->Level = 1 + Abc_ObjRegular(pLatch->pCopy)->Level;

			pLatch->pNext = pNodeNew;
			pLatch->pCopy = pNodeNew;
			}
			addFrameMapping(pLatch->pCopy, pLatch, iFrame+1, arg);
			}
			}
	*/
}

/*
*Function*************************************************************

	Synopsis    [Splits the miter into two logic cones combined by an EXOR]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDemiter(pNtk *Abc_Ntk_t) int32 {
	var pNodeC, pNode *Abc_Obj_t
	var pNodeA, pNodeB base_abc.Edge
	var pPoNew *Abc_Obj_t
	var vNodes1, vNodes2 []*Abc_Obj_t
	var nCommon, i int32

	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert(Abc_NtkPoNum(pNtk) == 1, "Abc_NtkPoNum(pNtk) == 1")
	if base_abc.Abc_NodeIsExorType(Abc_ObjFanin0(Abc_NtkPo(pNtk, 0))) == 0 {
		fmt.Printf("The root of the miter is not an EXOR gate.\n")
		return 0
	}
	pNodeC = base_abc.Abc_NodeRecognizeMux(Abc_ObjFanin0(Abc_NtkPo(pNtk, 0)), &pNodeA, &pNodeB)
	assert(base_abc.Abc_ObjRegular(pNodeA) == base_abc.Abc_ObjRegular(pNodeB), "Abc_ObjRegular(pNodeA) == Abc_ObjRegular(pNodeB)")
	if base_abc.Abc_ObjFaninC0(Abc_NtkPo(pNtk, 0)) != 0 {
		pNodeA = base_abc.Abc_ObjNot(pNodeA)
		pNodeB = base_abc.Abc_ObjNot(pNodeB)
	}

	// add the PO corresponding to control input
	pPoNew = base_abc.Abc_NtkCreatePo(pNtk)
	Abc_ObjAddFanin(pPoNew, pNodeC)
	base_abc.Abc_ObjAssignName(pPoNew, "addOut1", "")

	// add the PO corresponding to other input
	pPoNew = base_abc.Abc_NtkCreatePo(pNtk)
	Abc_ObjAddFanin(pPoNew, pNodeB)
	base_abc.Abc_ObjAssignName(pPoNew, "addOut2", "")

	// mark the nodes in the first cone
	pNodeB = pNodeB.Object.Signal()
	vNodes1 = base_abc.Abc_NtkDfsNodes(pNtk, []*Abc_Obj_t{pNodeC}, 1)
	vNodes2 = base_abc.Abc_NtkDfsNodes(pNtk, []*Abc_Obj_t{pNodeB.Object}, 1)

	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(len(vNodes1)); i++ {
		pNode = vNodes1[i]
		pNode.MarkA = true
	}
	nCommon = 0
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(len(vNodes2)); i++ {
		pNode = vNodes2[i]
		nCommon += boolInt(pNode.MarkA)
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(len(vNodes1)); i++ {
		pNode = vNodes1[i]
		pNode.MarkA = false
	}

	fmt.Printf("First cone = %6d.  Second cone = %6d.  Common = %6d.\n", len(vNodes1), len(vNodes2), nCommon)

	// reorder the latches
	base_abc.Abc_NtkOrderCisCos(pNtk)
	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Printf("Abc_NtkDemiter: The network check has failed.\n")
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Computes OR or AND of the POs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCombinePos(pNtk *Abc_Ntk_t, fAnd int32, fXor int32) int32 {
	var pNode *Abc_Obj_t
	var pMiter base_abc.Edge
	var i int32
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	//    assert( Abc_NtkLatchNum(pNtk) == 0 );
	if Abc_NtkPoNum(pNtk) == 1 {
		return 1
	}
	// start the result
	if fAnd != 0 {
		pMiter = base_abc.Abc_AigConst1(pNtk)
	} else {
		pMiter = base_abc.Abc_ObjNot(base_abc.Abc_AigConst1(pNtk))
	}
	// perform operations on the POs
	// inlining of macro Abc_NtkForEachPo
	for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
		pNode = Abc_NtkPo(pNtk, i)
		if fAnd != 0 {
			pMiter = base_abc.Abc_AigAnd(pNtk.FunctionManager.(*base_abc.Abc_Aig_t), pMiter, Abc_ObjChild0(pNode))
		} else if fXor != 0 {
			pMiter = base_abc.Abc_AigXor(pNtk.FunctionManager.(*base_abc.Abc_Aig_t), pMiter, Abc_ObjChild0(pNode))
		} else {
			pMiter = base_abc.Abc_AigOr(pNtk.FunctionManager.(*base_abc.Abc_Aig_t), pMiter, Abc_ObjChild0(pNode))
		}
	}
	// remove the POs and their names
	for i = Abc_NtkPoNum(pNtk) - 1; i >= 0; i-- {
		base_abc.Abc_NtkDeleteObj(Abc_NtkPo(pNtk, i))
	}
	assert(Abc_NtkPoNum(pNtk) == 0, "Abc_NtkPoNum(pNtk) == 0")
	// create the new PO
	pNode = base_abc.Abc_NtkCreatePo(pNtk)
	Abc_ObjAddFanin(pNode, pMiter)
	base_abc.Abc_ObjAssignName(pNode, "miter", "")
	base_abc.Abc_NtkOrderCisCos(pNtk)
	// make sure that everything is okay
	if base_abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Printf("Abc_NtkOrPos: The network check has failed.\n")
		return 0
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Miter construction for two networks.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkTryNewMiter(pFileName0 string, pFileName1 string) *Vec_Ptr_t {
	// Cnf_DataWriteIntoSolver2 is declared in the SAT package.
	var i int32
	var nVars int32
	var pVars []int32
	var iCiVarBeg int32
	var iCoVarBeg int32 = 1
	var nBTLimit int32 = 100000
	var pSat *sat_bsat.Solver
	var pCnf *aig_gia.Cnf_Dat_t = nil
	var vCexes *Vec_Ptr_t = nil
	var pNtk1 *Abc_Ntk_t = Io_Read(pFileName0, IO_FILE_VERILOG, 1, 0)
	var pNtk2 *Abc_Ntk_t = Io_Read(pFileName1, IO_FILE_VERILOG, 1, 0)
	var pNtk1_ *Abc_Ntk_t = Abc_NtkStrash(pNtk1, 1, 1, 0)
	var pNtk2_ *Abc_Ntk_t = Abc_NtkStrash(pNtk2, 1, 1, 0)
	var pMiter *Abc_Ntk_t = Abc_NtkMiter(pNtk1_, pNtk2_, 1, 0, 0, 1)
	var pGia *Gia_Man_t = Abc_NtkClpGia(pMiter)
	assert(Abc_NtkCiNum(pNtk1) == Abc_NtkCiNum(pNtk2), "Abc_NtkCiNum(pNtk1) == Abc_NtkCiNum(pNtk2)")
	assert(Abc_NtkCoNum(pNtk1) == Abc_NtkCoNum(pNtk2), "Abc_NtkCoNum(pNtk1) == Abc_NtkCoNum(pNtk2)")
	Abc_NtkDelete(pNtk1)
	Abc_NtkDelete(pNtk2)
	Abc_NtkDelete(pNtk1_)
	Abc_NtkDelete(pNtk2_)
	Abc_NtkDelete(pMiter)
	vCexes = aig_gia.Vec_PtrStart(int(Gia_ManPoNum(pGia)))
	pCnf = aig_gia.Mf_ManGenerateCnf(pGia, 8, 0, 0, 0, 0)
	nVars = Gia_ManPiNum(pGia)
	iCiVarBeg = int32(pCnf.Variables) - nVars
	// inlining of macro ABC_ALLOC
	pVars = make([]int32, int(nVars))
	for i = 0; i < nVars; i++ {
		pVars[i] = iCiVarBeg + i
	}
	pSat = sat_bsat.NewSolver()
	pSat.SetNumVars(int32(pCnf.Variables))
	for _, clause := range pCnf.Clauses {
		if !pSat.AddClause(clause...) {
			break
		}
	}
	for i = 0; i < Gia_ManPoNum(pGia); i++ {
		var Lit int32 = Abc_Var2Lit(iCoVarBeg+i, 0)
		status := pSat.SolveWithResourceLimits([]int32{Lit}, int64(nBTLimit), 0, 0, 0)
		assert(status != sat_bsat.Unknown, "status != sat_bsat.Unknown")
		if status == sat_bsat.Unsat {
			continue
		}
		Vec_PtrWriteEntry(vCexes, int(i), sat_bsat.Sat_SolverGetModel(pSat, pVars, nVars))
		fmt.Printf("Output %3d (out of %3d) is SAT.\n", i, Gia_ManPoNum(pGia))
	}
	Gia_ManStop(pGia)
	pSat.Delete()
	pVars = nil
	return vCexes
}

/*
*Function*************************************************************

	Synopsis    [Read node names.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkReadNodeNames(pNtk *Abc_Ntk_t, pFileName string) *Vec_Ptr_t {
	pFile, err := os.Open(pFileName)
	if err != nil {
		fmt.Printf("Cannot open node list \"%s\".\n", pFileName)
		return nil
	}
	defer pFile.Close()

	vNodes := Vec_PtrAlloc(100)
	scanner := bufio.NewScanner(pFile)
	for scanner.Scan() {
		for _, token := range strings.Fields(scanner.Text()) {
			pObj := Abc_NtkFindNode(pNtk, token)
			if pObj == nil {
				fmt.Printf("Cannot find node \"%s\".\n", token)
				Vec_PtrFree(vNodes)
				return nil
			}
			Vec_PtrPush(vNodes, pObj)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Cannot read node list \"%s\": %v\n", pFileName, err)
		Vec_PtrFree(vNodes)
		return nil
	}
	return vNodes
}

/*
*Function*************************************************************

	Synopsis    [Deriving specialized miter.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSpecialMuxTree_rec(pNew *Abc_Ntk_t, pCtrl []*Abc_Obj_t, nCtrl int32, pData []*Abc_Obj_t, Shift int32) *Abc_Obj_t {
	var pLit0, pLit1 *Abc_Obj_t
	if nCtrl == 0 {
		return pData[Shift]
	}
	pLit0 = Abc_NtkSpecialMuxTree_rec(pNew, pCtrl, nCtrl-1, pData, Shift)
	pLit1 = Abc_NtkSpecialMuxTree_rec(pNew, pCtrl, nCtrl-1, pData, Shift+(1<<(nCtrl-1)))
	return base_abc.Abc_NtkCreateNodeMux(pNew, pCtrl[nCtrl-1], pLit1, pLit0)
}
func Abc_NtkSpecialMiter(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t) *Abc_Ntk_t {
	var pNtkNew *Abc_Ntk_t
	var pObj, pFanin, pObjNew, pPoNew *Abc_Obj_t
	var pName string
	var vFirsts *Vec_Int_t = Vec_IntAlloc(Vec_PtrSize(vNodes))
	var vNames *Vec_Ptr_t = Vec_PtrAlloc(100)
	var vFanins *Vec_Ptr_t = Vec_PtrAlloc(100)
	var vDatas *Vec_Ptr_t = Vec_PtrAlloc(100)
	var vDfs = base_abc.Abc_NtkDfs(pNtk, 1)
	var vCopies *Vec_Ptr_t = aig_gia.Vec_PtrStart(int(Abc_NtkObjNumMax(pNtk)))
	var i, k, Index, First, nNewPis int32
	// count the number of additional inputs
	base_abc.Abc_NtkCleanCopy(pNtk)
	base_abc.Abc_NtkCleanMarkA(pNtk)
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vNodes); i++ {
		pObj = Vec_PtrEntry(vNodes, i).(*Abc_Obj_t)
		var Buffer string // was [1000]byte in C

		assert((Abc_ObjIsNode(pObj) != 0), "Abc_ObjIsNode(pObj)")
		pObj.MarkA = true
		Vec_IntPush(vFirsts, nNewPis)
		nNewPis += 1 << Abc_ObjFaninNum(pObj)
		for k = 0; k < (1 << Abc_ObjFaninNum(pObj)); k++ {
			Buffer = fmt.Sprintf("pi_%s_%d", Abc_ObjName(pObj), k)
			Vec_PtrPush(vNames, misc_util.Abc_UtilStrsav(Buffer))
		}
	}
	// create new network with the additional PIs
	pNtkNew = base_abc.Abc_NtkAlloc(base_abc.Abc_NtkType_t(pNtk.Type), base_abc.Abc_NtkFunc_t(pNtk.Function), 1)
	Abc_NtkSetName(pNtkNew, Extra_UtilStrsav("miter"))
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vNames); i++ {
		pName = Vec_PtrEntry(vNames, i).(string)
		base_abc.Abc_ObjAssignName(base_abc.Abc_NtkCreatePi(pNtkNew), pName, "")
	}
	// duplicate PIs
	// inlining of macro Abc_NtkForEachCi
	for i = 0; i < Abc_NtkCiNum(pNtk); i++ {
		pObj = base_abc.Abc_NtkCi(pNtk, i)
		Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(base_abc.Abc_NtkDupObj(pNtkNew, pObj, 1)))
	}
	// copy nodes
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(len(vDfs)); i++ {
		pObj = vDfs[i]
		if !pObj.MarkA {
			base_abc.Abc_NtkDupObj(pNtkNew, pObj, 1)
			// inlining of macro Abc_ObjForEachFanin
			for k = 0; k < Abc_ObjFaninNum(pObj); k++ {
				pFanin = base_abc.Abc_ObjFanin(pObj, k)
				Abc_ObjAddFanin(Abc_ObjReadCopyEdge(pObj).Object, Abc_ObjReadCopyEdge(pFanin))
			}
			Vec_PtrWriteEntry(vCopies, pObj.ID, Abc_ObjReadCopyEdge(pObj).Object)
			continue
		}
		Vec_PtrClear(vFanins)
		// inlining of macro Abc_ObjForEachFanin
		for k = 0; k < Abc_ObjFaninNum(pObj); k++ {
			pFanin = base_abc.Abc_ObjFanin(pObj, k)
			Vec_PtrPush(vFanins, Abc_ObjReadCopyEdge(pFanin).Object)
		}
		Index = int32(aig_gia.Vec_PtrFind(vNodes, pObj))
		assert(Index >= 0, "Index >= 0")
		First = Vec_IntEntry(vFirsts, Index)
		Vec_PtrClear(vDatas)
		for k = 0; k < (1 << Abc_ObjFaninNum(pObj)); k++ {
			Vec_PtrPush(vDatas, base_abc.Abc_NtkCi(pNtkNew, First+k))
		}
		Abc_ObjSetCopyEdge(pObj, abcCompatCopyEdge(Abc_NtkSpecialMuxTree_rec(pNtkNew,
			vecObjPtrArray(vFanins), Vec_PtrSize(vFanins),
			vecObjPtrArray(vDatas), 0)))
		Vec_PtrWriteEntry(vCopies, pObj.ID, Abc_ObjReadCopyEdge(pObj).Object)
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < int32(len(vDfs)); i++ {
		pObj = vDfs[i]
		pObj.MarkA = false
		base_abc.Abc_NtkDupObj(pNtkNew, pObj, 0)
		base_abc.Abc_ObjAssignName(Abc_ObjReadCopyEdge(pObj).Object, Abc_ObjName(pObj), "_copy")
		// inlining of macro Abc_ObjForEachFanin
		for k = 0; k < Abc_ObjFaninNum(pObj); k++ {
			pFanin = base_abc.Abc_ObjFanin(pObj, k)
			Abc_ObjAddFanin(Abc_ObjReadCopyEdge(pObj).Object, Abc_ObjReadCopyEdge(pFanin))
		}
	}
	// create miter
	Vec_PtrClear(vDatas)
	// inlining of macro Abc_NtkForEachCo
	for i = 0; i < Abc_NtkCoNum(pNtk); i++ {
		pObj = base_abc.Abc_NtkCo(pNtk, i)
		Vec_PtrClear(vFanins)
		Vec_PtrPush(vFanins, Abc_ObjReadCopyEdge(Abc_ObjFanin0(pObj)).Object)
		Vec_PtrPush(vFanins, Vec_PtrEntry(vCopies, base_abc.Abc_ObjId(Abc_ObjFanin0(pObj))).(*Abc_Obj_t))
		Vec_PtrPush(vDatas, base_abc.Abc_NtkCreateNodeExor(pNtkNew, vecObjPtrArray(vFanins)))
	}
	if Vec_PtrSize(vDatas) > 1 {
		pObjNew = base_abc.Abc_NtkCreateNodeOr(pNtkNew, vecObjPtrArray(vDatas))
	} else {
		pObjNew = Vec_PtrEntry(vDatas, 0).(*Abc_Obj_t)
	}
	pPoNew = base_abc.Abc_NtkCreatePo(pNtkNew)
	Abc_ObjAddFanin(pPoNew, pObjNew)
	base_abc.Abc_ObjAssignName(pPoNew, "miter_output", "")
	// cleanup
	Vec_IntFree(vFirsts)
	Vec_PtrFreeFree(vNames)
	Vec_PtrFree(vFanins)
	Vec_PtrFree(vDatas)
	Vec_PtrFree(vCopies)
	return pNtkNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

// vecObjPtrArray returns the object entries in order and asserts the vector's
// element type at the boundary where C's typed pointer array was used.
func vecObjPtrArray(v *Vec_Ptr_t) []*Abc_Obj_t {
	out := make([]*Abc_Obj_t, Vec_PtrSize(v))
	for i := range out {
		out[i] = Vec_PtrEntry(v, int32(i)).(*Abc_Obj_t)
	}
	return out
}
