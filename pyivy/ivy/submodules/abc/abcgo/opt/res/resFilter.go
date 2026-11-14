package res

import "github.com/glycerine/abcgo/base/abc"

/**CFile****************************************************************

  FileName    [resFilter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Filtering resubstitution candidates.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resFilter.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static unsigned * Res_FilterCollectFaninInfo( Res_Win_t * pWin, Res_Sim_t * pSim, unsigned uMask );

// static int        Res_FilterCriticalFanin( Abc_Obj_t * pNode );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Finds sets of feasible candidates.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_FilterCandidates(pWin *Res_Win_t, pAig *Abc_Ntk_t, pSim *Res_Sim_t, vResubs *Vec_Vec_t, vResubsW *Vec_Vec_t, nFaninsMax int32, fArea int32) int32 {
	var pFanin, pFaninTemp *Abc_Obj_t
	var pInfo, pInfoDiv, pInfoDiv2 []uint32
	var Counter, RetValue, i, i2, d, d2, iDiv, iDiv2, k int32

	// check that the info the node is one
	pInfo = resWords(Vec_PtrEntryAny(pSim.vOuts, 1))
	RetValue = Abc_InfoIsOne(pInfo, pSim.nWordsOut)
	if RetValue == 0 {
		//        printf( "Failed 1!\n" );
		return 0
	}

	// collect the fanin info
	pInfo = Res_FilterCollectFaninInfo(pWin, pSim, ^uint32(0))
	RetValue = Abc_InfoIsOne(pInfo, pSim.nWordsOut)
	if RetValue == 0 {
		//        printf( "Failed 2!\n" );
		return 0
	}

	// try removing each fanin
	//    printf( "Fanins: " );
	Counter = 0
	Vec_VecClear(vResubs)
	Vec_VecClear(vResubsW)
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < abc.Abc_ObjFaninNum(pWin.pNode); i++ {
		pFanin = abc.Abc_ObjFanin(pWin.pNode, i)
		if (fArea != 0) && abc.Abc_ObjFanoutNum(pFanin) > 1 {
			continue
		}
		// get simulation info without this fanin
		pInfo = Res_FilterCollectFaninInfo(pWin, pSim, ^(uint32(1) << uint(i)))
		RetValue = Abc_InfoIsOne(pInfo, pSim.nWordsOut)
		if RetValue != 0 {
			//            printf( "Node %4d. Candidate fanin %4d.\n", pWin->pNode->Id, pFanin->Id );
			// collect the nodes
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 0))
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 1))
			// inlining of macro Abc_ObjForEachFanin
			for k = 0; k < abc.Abc_ObjFaninNum(pWin.pNode); k++ {
				pFaninTemp = abc.Abc_ObjFanin(pWin.pNode, k)
				if k != i {
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 2+k))
					Vec_VecPush(vResubsW, Counter, pFaninTemp)
				}
			}
			Counter++
			if Counter == Vec_VecSize(vResubs) {
				return Counter
			}
		}
	}

	// try replacing each critical fanin by a non-critical fanin
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < abc.Abc_ObjFaninNum(pWin.pNode); i++ {
		pFanin = abc.Abc_ObjFanin(pWin.pNode, i)
		if abc.Abc_ObjFanoutNum(pFanin) > 1 {
			continue
		}
		// get simulation info without this fanin
		pInfo = Res_FilterCollectFaninInfo(pWin, pSim, ^(uint32(1) << uint(i)))
		// go over the set of divisors
		for d = abc.Abc_ObjFaninNum(pWin.pNode) + 2; d < abc.Abc_NtkPoNum(pAig); d++ {
			pInfoDiv = resWords(Vec_PtrEntryAny(pSim.vOuts, d))
			iDiv = d - (abc.Abc_ObjFaninNum(pWin.pNode) + 2)
			if !Abc_InfoIsOrOne(pInfo, pInfoDiv, pSim.nWordsOut) {
				continue
			}
			// collect the nodes
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 0))
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 1))
			// collect the remaning fanins and the divisor
			// inlining of macro Abc_ObjForEachFanin
			for k = 0; k < abc.Abc_ObjFaninNum(pWin.pNode); k++ {
				pFaninTemp = abc.Abc_ObjFanin(pWin.pNode, k)
				if k != i {
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 2+k))
					Vec_VecPush(vResubsW, Counter, pFaninTemp)
				}
			}
			// collect the divisor
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, d))
			Vec_VecPush(vResubsW, Counter, Vec_PtrEntry(pWin.vDivs, iDiv))
			Counter++
			if Counter == Vec_VecSize(vResubs) {
				return Counter
			}
		}
	}

	// consider the case when two fanins can be added instead of one
	if abc.Abc_ObjFaninNum(pWin.pNode) < nFaninsMax {
		// try to replace each critical fanin by two non-critical fanins
		// inlining of macro Abc_ObjForEachFanin
		for i = 0; i < abc.Abc_ObjFaninNum(pWin.pNode); i++ {
			pFanin = abc.Abc_ObjFanin(pWin.pNode, i)
			if abc.Abc_ObjFanoutNum(pFanin) > 1 {
				continue
			}
			// get simulation info without this fanin
			pInfo = Res_FilterCollectFaninInfo(pWin, pSim, ^(uint32(1) << uint(i)))
			// go over the set of divisors
			for d = abc.Abc_ObjFaninNum(pWin.pNode) + 2; d < abc.Abc_NtkPoNum(pAig); d++ {
				pInfoDiv = resWords(Vec_PtrEntryAny(pSim.vOuts, d))
				iDiv = d - (abc.Abc_ObjFaninNum(pWin.pNode) + 2)
				// go through the second divisor
				for d2 = d + 1; d2 < abc.Abc_NtkPoNum(pAig); d2++ {
					pInfoDiv2 = resWords(Vec_PtrEntryAny(pSim.vOuts, d2))
					iDiv2 = d2 - (abc.Abc_ObjFaninNum(pWin.pNode) + 2)
					if !Abc_InfoIsOrOne3(pInfo, pInfoDiv, pInfoDiv2, pSim.nWordsOut) {
						continue
					}
					// collect the nodes
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 0))
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 1))
					// collect the remaning fanins and the divisor
					// inlining of macro Abc_ObjForEachFanin
					for k = 0; k < abc.Abc_ObjFaninNum(pWin.pNode); k++ {
						pFaninTemp = abc.Abc_ObjFanin(pWin.pNode, k)
						if k != i {
							Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 2+k))
							Vec_VecPush(vResubsW, Counter, pFaninTemp)
						}
					}
					// collect the divisor
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, d))
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, d2))
					Vec_VecPush(vResubsW, Counter, Vec_PtrEntry(pWin.vDivs, iDiv))
					Vec_VecPush(vResubsW, Counter, Vec_PtrEntry(pWin.vDivs, iDiv2))
					Counter++
					if Counter == Vec_VecSize(vResubs) {
						return Counter
					}
				}
			}
		}
	}

	// try to replace two nets by one
	if !(fArea != 0) {
		// inlining of macro Abc_ObjForEachFanin
		for i = 0; i < abc.Abc_ObjFaninNum(pWin.pNode); i++ {
			pFanin = abc.Abc_ObjFanin(pWin.pNode, i)
			for i2 = i + 1; i2 < abc.Abc_ObjFaninNum(pWin.pNode); i2++ {
				_ = abc.Abc_ObjFanin(pWin.pNode, i2)
				// get simulation info without these fanins
				pInfo = Res_FilterCollectFaninInfo(pWin, pSim, (^(uint32(1) << uint(i)))&(^(uint32(1) << uint(i2))))
				// go over the set of divisors
				for d = abc.Abc_ObjFaninNum(pWin.pNode) + 2; d < abc.Abc_NtkPoNum(pAig); d++ {
					pInfoDiv = resWords(Vec_PtrEntryAny(pSim.vOuts, d))
					iDiv = d - (abc.Abc_ObjFaninNum(pWin.pNode) + 2)
					if !Abc_InfoIsOrOne(pInfo, pInfoDiv, pSim.nWordsOut) {
						continue
					}
					// collect the nodes
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 0))
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 1))
					// collect the remaning fanins and the divisor
					// inlining of macro Abc_ObjForEachFanin
					for k = 0; k < abc.Abc_ObjFaninNum(pWin.pNode); k++ {
						pFaninTemp = abc.Abc_ObjFanin(pWin.pNode, k)
						if k != i && k != i2 {
							Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 2+k))
							Vec_VecPush(vResubsW, Counter, pFaninTemp)
						}
					}
					// collect the divisor
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, d))
					Vec_VecPush(vResubsW, Counter, Vec_PtrEntry(pWin.vDivs, iDiv))
					Counter++
					if Counter == Vec_VecSize(vResubs) {
						return Counter
					}
				}
			}
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Finds sets of feasible candidates.]

	Description [This procedure is a special case of the above.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_FilterCandidatesArea(pWin *Res_Win_t, pAig *Abc_Ntk_t, pSim *Res_Sim_t, vResubs *Vec_Vec_t, vResubsW *Vec_Vec_t, nFaninsMax int32) int32 {
	var pFanin *Abc_Obj_t
	var pInfo, pInfoDiv, pInfoDiv2 []uint32
	var Counter, RetValue, d, d2, k, iDiv, iDiv2, iBest int32

	// check that the info the node is one
	pInfo = resWords(Vec_PtrEntryAny(pSim.vOuts, 1))
	RetValue = Abc_InfoIsOne(pInfo, pSim.nWordsOut)
	if RetValue == 0 {
		//        printf( "Failed 1!\n" );
		return 0
	}

	// collect the fanin info
	pInfo = Res_FilterCollectFaninInfo(pWin, pSim, ^uint32(0))
	RetValue = Abc_InfoIsOne(pInfo, pSim.nWordsOut)
	if RetValue == 0 {
		//        printf( "Failed 2!\n" );
		return 0
	}

	// try removing fanins
	//    printf( "Fanins: " );
	Counter = 0
	Vec_VecClear(vResubs)
	Vec_VecClear(vResubsW)
	// get the best fanins
	iBest = Res_FilterCriticalFanin(pWin.pNode)
	if iBest == -1 {
		return 0
	}

	// get the info without the critical fanin
	pInfo = Res_FilterCollectFaninInfo(pWin, pSim, ^(uint32(1) << uint(iBest)))
	RetValue = Abc_InfoIsOne(pInfo, pSim.nWordsOut)
	if RetValue != 0 {
		//        printf( "Can be done without one!\n" );
		// collect the nodes
		Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 0))
		Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 1))
		// inlining of macro Abc_ObjForEachFanin
		for k = 0; k < abc.Abc_ObjFaninNum(pWin.pNode); k++ {
			pFanin = abc.Abc_ObjFanin(pWin.pNode, k)
			if k != iBest {
				Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 2+k))
				Vec_VecPush(vResubsW, Counter, pFanin)
			}
		}
		Counter++
		//        printf( "*" );
		return Counter
	}

	// go through the divisors
	for d = abc.Abc_ObjFaninNum(pWin.pNode) + 2; d < abc.Abc_NtkPoNum(pAig); d++ {
		pInfoDiv = resWords(Vec_PtrEntryAny(pSim.vOuts, d))
		iDiv = d - (abc.Abc_ObjFaninNum(pWin.pNode) + 2)
		if !Abc_InfoIsOrOne(pInfo, pInfoDiv, pSim.nWordsOut) {
			continue
		}
		//if ( abc.Abc_ObjLevel(pWin->pNode) <= abc.Abc_ObjLevel( Vec_PtrEntry(pWin->vDivs, iDiv) ) )
		//    printf( "Node level = %d. Divisor level = %d.\n", abc.Abc_ObjLevel(pWin->pNode), abc.Abc_ObjLevel( Vec_PtrEntry(pWin->vDivs, iDiv) ) );
		// collect the nodes
		Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 0))
		Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 1))
		// collect the remaning fanins and the divisor
		// inlining of macro Abc_ObjForEachFanin
		for k = 0; k < abc.Abc_ObjFaninNum(pWin.pNode); k++ {
			pFanin = abc.Abc_ObjFanin(pWin.pNode, k)
			if k != iBest {
				Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 2+k))
				Vec_VecPush(vResubsW, Counter, pFanin)
			}
		}
		// collect the divisor
		Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, d))
		Vec_VecPush(vResubsW, Counter, Vec_PtrEntry(pWin.vDivs, iDiv))
		Counter++

		if Counter == Vec_VecSize(vResubs) {
			break
		}
	}

	if Counter > 0 || abc.Abc_ObjFaninNum(pWin.pNode) >= nFaninsMax {
		return Counter
	}

	// try to find the node pairs
	for d = abc.Abc_ObjFaninNum(pWin.pNode) + 2; d < abc.Abc_NtkPoNum(pAig); d++ {
		pInfoDiv = resWords(Vec_PtrEntryAny(pSim.vOuts, d))
		iDiv = d - (abc.Abc_ObjFaninNum(pWin.pNode) + 2)
		// go through the second divisor
		for d2 = d + 1; d2 < abc.Abc_NtkPoNum(pAig); d2++ {
			pInfoDiv2 = resWords(Vec_PtrEntryAny(pSim.vOuts, d2))
			iDiv2 = d2 - (abc.Abc_ObjFaninNum(pWin.pNode) + 2)

			if !Abc_InfoIsOrOne3(pInfo, pInfoDiv, pInfoDiv2, pSim.nWordsOut) {
				continue
			}
			// collect the nodes
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 0))
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 1))
			// collect the remaning fanins and the divisor
			// inlining of macro Abc_ObjForEachFanin
			for k = 0; k < abc.Abc_ObjFaninNum(pWin.pNode); k++ {
				pFanin = abc.Abc_ObjFanin(pWin.pNode, k)
				if k != iBest {
					Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, 2+k))
					Vec_VecPush(vResubsW, Counter, pFanin)
				}
			}
			// collect the divisor
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, d))
			Vec_VecPush(vResubs, Counter, abc.Abc_NtkPo(pAig, d2))
			Vec_VecPush(vResubsW, Counter, Vec_PtrEntry(pWin.vDivs, iDiv))
			Vec_VecPush(vResubsW, Counter, Vec_PtrEntry(pWin.vDivs, iDiv2))
			Counter++

			if Counter == Vec_VecSize(vResubs) {
				break
			}
		}
		if Counter == Vec_VecSize(vResubs) {
			break
		}
	}
	return Counter
}

/*
*Function*************************************************************

	Synopsis    [Finds sets of feasible candidates.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_FilterCollectFaninInfo(pWin *Res_Win_t, pSim *Res_Sim_t, uMask uint32) []uint32 {
	var pInfo []uint32
	var i int32
	pInfo = resWords(Vec_PtrEntryAny(pSim.vOuts, 0))
	Abc_InfoClear(pInfo, pSim.nWordsOut)
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < abc.Abc_ObjFaninNum(pWin.pNode); i++ {
		if uMask&(uint32(1)<<uint(i)) != 0 {
			Abc_InfoOr(pInfo, resWords(Vec_PtrEntryAny(pSim.vOuts, 2+i)), pSim.nWordsOut)
		}
	}
	return pInfo
}

/*
*Function*************************************************************

	Synopsis    [Returns the index of the most critical fanin.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_FilterCriticalFanin(pNode *Abc_Obj_t) int32 {
	var pFanin *Abc_Obj_t
	var i int32
	var iBest int32 = -1
	var CostMax int32
	var CostCur int32
	// inlining of macro Abc_ObjForEachFanin
	for i = 0; i < abc.Abc_ObjFaninNum(pNode); i++ {
		pFanin = abc.Abc_ObjFanin(pNode, i)
		if abc.Abc_ObjIsNode(pFanin) == 0 {
			continue
		}
		if abc.Abc_ObjFanoutNum(pFanin) > 1 {
			continue
		}
		CostCur = Res_WinVisitMffc(pFanin)
		if CostMax < CostCur {
			CostMax = CostCur
			iBest = i
		}
	}
	//    if ( CostMax > 0 )
	//        printf( "<%d>", CostMax );
	return iBest
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
