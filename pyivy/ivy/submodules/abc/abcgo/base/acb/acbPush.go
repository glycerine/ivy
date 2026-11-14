package acb

import "fmt"

/**CFile****************************************************************

  FileName    [acbPush.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Implementation of logic pushing.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: acbPush.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acb.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Pushing logic to the fanout.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjPushToFanout(p *Acb_Ntk_t, iObj int32, iFaninIndex int32, iFanout int32) {
	var c0 word
	var uTruthObjNew word
	var uTruthObj word = Acb_ObjTruth(p, iObj)
	var Gate word
	var c1 word
	var uTruthFanNew word
	var uTruthFan word = Acb_ObjTruth(p, iFanout)
	var DecType int32 = Abc_Tt6CheckOutDec(uTruthObj, iFaninIndex, &uTruthObjNew)
	var iFanin int32 = Acb_ObjFanin(p, iObj, iFaninIndex)
	var iFanoutObjIndex int32 = Acb_ObjWhatFanin(p, iFanout, iObj)
	var iFanoutFaninIndex int32 = Acb_ObjWhatFanin(p, iFanout, iFanin)
	if iFanoutFaninIndex == -1 {
		iFanoutFaninIndex = Acb_ObjFaninNum(p, iFanout)
	}
	assert((Acb_ObjIsCio(p, iObj) == 0), "(Acb_ObjIsCio(p, iObj) == 0)")
	assert((Acb_ObjIsCio(p, iFanout) == 0), "(Acb_ObjIsCio(p, iFanout) == 0)")
	assert(iFanoutFaninIndex >= 0, "iFanoutFaninIndex >= 0")
	assert(iFaninIndex < Acb_ObjFaninNum(p, iObj), "iFaninIndex < Acb_ObjFaninNum(p, iObj)")
	assert(Acb_ObjFanoutNum(p, iObj) == 1, "Acb_ObjFanoutNum(p, iObj) == 1")
	// compute new function of the fanout
	c0 = Abc_Tt6Cofactor0(uTruthFan, iFanoutObjIndex)
	c1 = Abc_Tt6Cofactor1(uTruthFan, iFanoutObjIndex)
	if DecType == 0 { //  F = i * G
		Gate = s_Truths6[iFanoutFaninIndex] & s_Truths6[iFanoutObjIndex]
	} else if DecType == 1 { //  F = ^i * G
		Gate = ^s_Truths6[iFanoutFaninIndex] & s_Truths6[iFanoutObjIndex]
	} else if DecType == 2 { //  F = ^i + G
		Gate = ^s_Truths6[iFanoutFaninIndex] | s_Truths6[iFanoutObjIndex]
	} else if DecType == 3 { //  F = i + G
		Gate = s_Truths6[iFanoutFaninIndex] | s_Truths6[iFanoutObjIndex]
	} else if DecType == 4 { //  F = i # G
		Gate = s_Truths6[iFanoutFaninIndex] ^ s_Truths6[iFanoutObjIndex]
	} else {
		assert(false, "0")
	}
	uTruthFanNew = (^Gate & c0) | (Gate & c1)
	// update functions
	Vec_WrdWriteEntry(&p.vObjTruth, iObj, Abc_Tt6RemoveVar(uTruthObjNew, iFaninIndex))
	Vec_WrdWriteEntry(&p.vObjTruth, iFanout, uTruthFanNew)
	// update fanins
	Acb_ObjRemoveFaninFanoutOne(p, iObj, iFanin)
	if iFanoutFaninIndex == Acb_ObjFaninNum(p, iFanout) { // adding new
		Acb_ObjAddFaninFanoutOne(p, iFanout, iFanin)
	}
}

/*
*Function*************************************************************

	Synopsis    [Pushing logic to the fanin.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjPushToFanin(p *Acb_Ntk_t, iObj int32, iFaninIndex2 int32, iFanin int32) {
	var uTruthObjNew word
	var uTruthObj word = Acb_ObjTruth(p, iObj)
	var uTruthFanNew word
	var uTruthFan word = Acb_ObjTruth(p, iFanin)
	var iFaninIndex int32 = Acb_ObjWhatFanin(p, iObj, iFanin)
	var DecType int32 = Abc_TtCheckDsdAnd(uTruthObj, iFaninIndex, iFaninIndex2, &uTruthObjNew)
	var iFanin2 int32 = Acb_ObjFanin(p, iObj, iFaninIndex2)
	var iFaninFaninIndex int32 = Acb_ObjWhatFanin(p, iFanin, iFanin2)
	if iFaninFaninIndex == -1 {
		iFaninFaninIndex = Acb_ObjFaninNum(p, iFanin)
	}
	assert((Acb_ObjIsCio(p, iObj) == 0), "(Acb_ObjIsCio(p, iObj) == 0)")
	assert((Acb_ObjIsCio(p, iFanin) == 0), "(Acb_ObjIsCio(p, iFanin) == 0)")
	assert(iFaninIndex < Acb_ObjFaninNum(p, iObj), "iFaninIndex  < Acb_ObjFaninNum(p, iObj)")
	assert(iFaninIndex2 < Acb_ObjFaninNum(p, iObj), "iFaninIndex2 < Acb_ObjFaninNum(p, iObj)")
	assert(iFaninIndex != iFaninIndex2, "iFaninIndex != iFaninIndex2")
	assert(Acb_ObjFanoutNum(p, iFanin) == 1, "Acb_ObjFanoutNum(p, iFanin) == 1")
	// compute new function of the fanout
	if DecType == 0 { //  i *  j
		uTruthFanNew = uTruthFan & s_Truths6[iFaninFaninIndex]
	} else if DecType == 1 { //  i * !j
		uTruthFanNew = ^uTruthFan & s_Truths6[iFaninFaninIndex]
	} else if DecType == 2 { // !i *  j
		uTruthFanNew = uTruthFan & ^s_Truths6[iFaninFaninIndex]
	} else if DecType == 3 { // !i * !j
		uTruthFanNew = ^uTruthFan & ^s_Truths6[iFaninFaninIndex]
	} else if DecType == 4 { //  i #  j
		uTruthFanNew = uTruthFan ^ s_Truths6[iFaninFaninIndex]
	} else {
		assert(false, "0")
	}
	// update functions
	Vec_WrdWriteEntry(&p.vObjTruth, iObj, Abc_Tt6RemoveVar(uTruthObjNew, iFaninIndex2))
	Vec_WrdWriteEntry(&p.vObjTruth, iFanin, uTruthFanNew)
	// update fanins
	Acb_ObjRemoveFaninFanoutOne(p, iObj, iFanin2)
	if iFaninFaninIndex == Acb_ObjFaninNum(p, iFanin) { // adding new
		Acb_ObjAddFaninFanoutOne(p, iFanin, iFanin2)
	}
}

/*
*Function*************************************************************

	Synopsis    [Removing constants, buffers, duplicated fanins.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjFindNodeFanout(p *Acb_Ntk_t, iObj int32) int32 {
	var i, iFanout int32
	// inlining of macro Acb_ObjForEachFanout
	// inlining of macro Vec_IntForEachEntry
	for i = 0; i < Vec_IntSize(Vec_WecEntry(&p.vFanouts, iObj)); i++ {
		iFanout = Vec_IntEntry(Vec_WecEntry(&p.vFanouts, iObj), i)
		if Acb_ObjType(p, iFanout) != 0 && Acb_ObjIsCio(p, iFanout) == 0 {
			return iFanout
		}
	}
	return -1
}
func Acb_ObjSuppMin_int(p *Acb_Ntk_t, iObj int32) int32 {
	var k int32
	var iFanin int32
	var pFanins []int32
	var uTruth word = Acb_ObjTruth(p, iObj)
	// inlining of macro Acb_ObjForEachFaninFast
	pFanins = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanins[0]; k++ {
		iFanin = pFanins[k+1]
		if Abc_Tt6HasVar(uTruth, k) != 0 {
			continue
		}
		Acb_ObjDeleteFaninIndex(p, iObj, k)
		Vec_IntRemove(Vec_WecEntry(&p.vFanouts, iFanin), iObj)
		Vec_WrdWriteEntry(&p.vObjTruth, iObj, Abc_Tt6RemoveVar(uTruth, k))
		return 1
	}
	return 0
}
func Acb_ObjSuppMin(p *Acb_Ntk_t, iObj int32) {
	for Acb_ObjSuppMin_int(p, iObj) != 0 {
	}
}
func Acb_ObjRemoveDup(p *Acb_Ntk_t, iObj int32, i int32, j int32) {
	var c00 word
	var c11 word
	var uTruthNew word
	var uTruth word = Acb_ObjTruth(p, iObj)
	assert((Acb_ObjIsCio(p, iObj) == 0), "(Acb_ObjIsCio(p, iObj) == 0)")
	assert(Acb_ObjFanin(p, iObj, i) == Acb_ObjFanin(p, iObj, j), "Acb_ObjFanin(p, iObj, i) == Acb_ObjFanin(p, iObj, j)")
	c00 = Abc_Tt6Cofactor0(Abc_Tt6Cofactor0(uTruth, i), j)
	c11 = Abc_Tt6Cofactor1(Abc_Tt6Cofactor1(uTruth, i), j)
	uTruthNew = (^s_Truths6[i] & c00) | (s_Truths6[i] & c11)
	Vec_WrdWriteEntry(&p.vObjTruth, iObj, Abc_Tt6RemoveVar(uTruthNew, j))
	// Save the removed edge before compacting the fanin list. The C source
	// removes from the wrong object's fanouts and reads the shifted index.
	fanin := Acb_ObjFanin(p, iObj, j)
	Acb_ObjDeleteFaninIndex(p, iObj, j)
	Vec_IntRemove(Vec_WecEntry(&p.vFanouts, fanin), iObj)
	Acb_ObjSuppMin(p, iObj)
}
func Acb_ObjRemoveDupFanins_int(p *Acb_Ntk_t, iObj int32) int32 {
	var i int32
	var k int32
	var pFanins []int32 = Acb_ObjFanins(p, iObj)
	for i = 0; i < pFanins[0]; i++ {
		for k = i + 1; k < pFanins[0]; k++ {
			if pFanins[1+i] != pFanins[1+k] {
				continue
			}
			Acb_ObjRemoveDup(p, iObj, i, k)
			return 1
		}
	}
	return 0
}
func Acb_ObjRemoveDupFanins(p *Acb_Ntk_t, iObj int32) {
	assert((Acb_ObjIsCio(p, iObj) == 0), "(Acb_ObjIsCio(p, iObj) == 0)")
	for Acb_ObjRemoveDupFanins_int(p, iObj) != 0 {
	}
}
func Acb_ObjRemoveConst(p *Acb_Ntk_t, iObj int32) {
	var iFanout int32
	var uTruth word = Acb_ObjTruth(p, iObj)
	assert((Acb_ObjIsCio(p, iObj) == 0), "(Acb_ObjIsCio(p, iObj) == 0)")
	assert(Acb_ObjFaninNum(p, iObj) == 0, "Acb_ObjFaninNum(p, iObj) == 0")
	assert(uTruth == 0 || ^uTruth == 0, "uTruth == 0 || ^uTruth == 0")
	for iFanout = Acb_ObjFindNodeFanout(p, iObj); iFanout >= 0; iFanout = Acb_ObjFindNodeFanout(p, iObj) {
		var iObjIndex int32 = Acb_ObjWhatFanin(p, iFanout, iObj)
		var uTruthF word = Acb_ObjTruth(p, iFanout)
		Acb_ObjRemoveFaninFanoutOne(p, iFanout, iObj)
		if uTruth&1 != 0 {
			uTruthF = Abc_Tt6Cofactor1(uTruthF, iObjIndex)
		} else {
			uTruthF = Abc_Tt6Cofactor0(uTruthF, iObjIndex)
		}
		Vec_WrdWriteEntry(&p.vObjTruth, iFanout, Abc_Tt6RemoveVar(uTruthF, iObjIndex))
		Acb_ObjSuppMin(p, iFanout)
	}
	if Acb_ObjFanoutNum(p, iObj) == 0 {
		Acb_ObjCleanType(p, iObj)
	}
}
func Acb_ObjRemoveBufInv(p *Acb_Ntk_t, iObj int32) {
	var iFanout int32
	var uTruth word = Acb_ObjTruth(p, iObj)
	assert((Acb_ObjIsCio(p, iObj) == 0), "(Acb_ObjIsCio(p, iObj) == 0)")
	assert(Acb_ObjFaninNum(p, iObj) == 1, "Acb_ObjFaninNum(p, iObj) == 1")
	assert(uTruth == s_Truths6[0] || ^uTruth == s_Truths6[0], "uTruth == s_Truths6[0] || ^uTruth == s_Truths6[0]")
	for iFanout = Acb_ObjFindNodeFanout(p, iObj); iFanout >= 0; iFanout = Acb_ObjFindNodeFanout(p, iObj) {
		var iFanin int32 = Acb_ObjFanin(p, iObj, 0)
		var iObjIndex int32 = Acb_ObjWhatFanin(p, iFanout, iObj)
		Acb_ObjPatchFanin(p, iFanout, iObj, iFanin)
		if uTruth&1 != 0 { // inv

			var uTruthF word = Acb_ObjTruth(p, iFanout)
			Vec_WrdWriteEntry(&p.vObjTruth, iFanout, Abc_Tt6Flip(uTruthF, iObjIndex))
		}
		Acb_ObjRemoveDupFanins(p, iFanout)
	}
	for (uTruth&1) == 0 && Acb_ObjFanoutNum(p, iObj) > 0 {
		var iFanin int32 = Acb_ObjFanin(p, iObj, 0)
		var iFanout int32 = Acb_ObjFanout(p, iObj, 0)
		assert(Acb_ObjIsCo(p, iFanout) != 0, "Acb_ObjIsCo(p, iFanout)")
		Acb_ObjPatchFanin(p, iFanout, iObj, iFanin)
	}
	if Acb_ObjFanoutNum(p, iObj) == 0 {
		Acb_ObjRemoveFaninFanout(p, iObj)
		Acb_ObjRemoveFanins(p, iObj)
		Acb_ObjCleanType(p, iObj)
	}
}

/*
*Function*************************************************************

	Synopsis    [Check if the node can have its logic pushed.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjFindFaninPushableIndex(p *Acb_Ntk_t, iObj int32, iFanIndex int32) int32 {
	var k int32
	var pFanins []int32
	// inlining of macro Acb_ObjForEachFaninFast
	pFanins = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanins[0]; k++ {
		if k != iFanIndex && Abc_TtCheckDsdAnd(Acb_ObjTruth(p, iObj), k, iFanIndex, nil) >= 0 {
			return k
		}
	}
	return -1
}
func Acb_ObjFindFanoutPushableIndex(p *Acb_Ntk_t, iObj int32) int32 {
	var k int32
	var pFanins []int32
	// inlining of macro Acb_ObjForEachFaninFast
	pFanins = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanins[0]; k++ {
		if Abc_Tt6CheckOutDec(Acb_ObjTruth(p, iObj), k, nil) >= 0 {
			return k
		}
	}
	return -1
}
func Acb_ObjPushToFanins(p *Acb_Ntk_t, iObj int32, nLutSize int32) int32 {
	var k int32
	var k2 int32
	var iFanin int32
	var pFanins []int32
	if Acb_ObjFaninNum(p, iObj) < 2 {
		return 0
	}
	// inlining of macro Acb_ObjForEachFaninFast
	pFanins = Acb_ObjFanins(p, iObj)
	for k = 0; k < pFanins[0]; k++ {
		iFanin = pFanins[k+1]
		if Acb_ObjIsCi(p, iFanin) != 0 {
			continue
		}
		if Acb_ObjFanoutNum(p, iFanin) > 1 {
			continue
		}
		if Acb_ObjFaninNum(p, iFanin) == nLutSize {
			continue
		}
		k2 = Acb_ObjFindFaninPushableIndex(p, iObj, k)
		if k2 == -1 {
			continue
		}
		//printf( "Object %4d : Pushing fanin %d (%d) into fanin %d.\n", iObj, Acb_ObjFanin(p, iObj, k2), k2, iFanin );
		Acb_ObjPushToFanin(p, iObj, k2, iFanin)
		return 1
	}
	if Acb_ObjFaninNum(p, iObj) == 2 && Acb_ObjFanoutNum(p, iObj) == 1 {
		var iFanout int32 = Acb_ObjFanout(p, iObj, 0)
		if (Acb_ObjIsCo(p, iFanout) == 0) && Acb_ObjFaninNum(p, iFanout) < nLutSize {
			k2 = Acb_ObjFindFanoutPushableIndex(p, iObj)
			//printf( "Object %4d : Pushing fanin %d (%d) into fanout %d.\n", iObj, Acb_ObjFanin(p, iObj, k2), k2, iFanout );
			Acb_ObjPushToFanout(p, iObj, k2, iFanout)
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkPushLogic(p *Acb_Ntk_t, nLutSize int32, fVerbose int32) {
	var n int32
	var iObj int32
	var nNodes int32 = Acb_NtkNodeNum(p)
	var nPushes int32
	Acb_NtkCreateFanout(p) // fanout data structure
	// inlining of macro Acb_NtkForEachNodeSupp
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 || Acb_ObjFaninNum(p, iObj) != 0 {
		} else {
			Acb_ObjRemoveConst(p, iObj)
		}
	}
	// inlining of macro Acb_NtkForEachNodeSupp
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 || Acb_ObjFaninNum(p, iObj) != 1 {
		} else {
			Acb_ObjRemoveBufInv(p, iObj)
		}
	}
	for n = 2; n <= nLutSize; n++ {
		// inlining of macro Acb_NtkForEachNodeSupp
		for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
			if Acb_ObjType(p, iObj) == 0 || Acb_ObjIsCio(p, iObj) != 0 || Acb_ObjFaninNum(p, iObj) != n {
			} else {
				for Acb_ObjPushToFanins(p, iObj, nLutSize) != 0 {
					nPushes++
				}
				if Acb_ObjFaninNum(p, iObj) == 1 {
					Acb_ObjRemoveBufInv(p, iObj)
				}
			}
		}
	}
	fmt.Printf("Saved %d nodes after %d pushes.\n", nNodes-Acb_NtkNodeNum(p), nPushes)
}

/*
*Function*************************************************************

	Synopsis    [Pushing logic to the fanin.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkPushLogic2(p *Acb_Ntk_t, nLutSize int32, fVerbose int32) {
	var iObj int32
	Acb_NtkCreateFanout(p) // fanout data structure
	// inlining of macro Acb_NtkForEachObj
	for iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++ {
		if Acb_ObjType(p, iObj) == 0 {
		} else if Acb_ObjIsCio(p, iObj) == 0 {
			break
		}
	}
	Acb_ObjPushToFanout(p, iObj, Acb_ObjFaninNum(p, iObj)-1, Acb_ObjFanout(p, iObj, 0))
	// Acb_ObjPushToFanin( p, Acb_ObjFanout(p, iObj, 0), Acb_ObjFaninNum(p, iObj)-1, iObj );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
