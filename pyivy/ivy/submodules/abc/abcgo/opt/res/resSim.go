package res

import "github.com/glycerine/abcgo/base/abc"
import "github.com/glycerine/abcgo/misc/extra"

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [resSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Simulation engine.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resSim.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Allocate simulation engine.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimAlloc(nWords int32) *Res_Sim_t {
	p := &Res_Sim_t{}
	// simulation parameters
	p.nWords = nWords
	p.nPats = p.nWords * 32
	p.nWordsIn = p.nPats
	p.nBytesIn = p.nPats * 4
	p.nPatsIn = p.nPats * 32
	p.nWordsOut = p.nPats * p.nWords
	p.nPatsOut = p.nPats * p.nPats
	// simulation info
	p.vPats = Vec_PtrAllocSimInfo(1024, p.nWordsIn)
	p.vPats0 = Vec_PtrAllocSimInfo(128, p.nWords)
	p.vPats1 = Vec_PtrAllocSimInfo(128, p.nWords)
	p.vOuts = Vec_PtrAllocSimInfo(128, p.nWordsOut)
	// resub candidates
	p.vCands = Vec_VecStart(16)
	return p
}

/*
*Function*************************************************************

	Synopsis    [Allocate simulation engine.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimAdjust(p *Res_Sim_t, pAig *Abc_Ntk_t, nTruePis int32) {
	p.rng = &resGlobalRand
	p.rng.seed(0xABC)

	assert(abc.Abc_NtkIsStrash(pAig) != 0, "abc.Abc_NtkIsStrash(pAig)")
	p.pAig = pAig
	p.nTruePis = nTruePis
	if Vec_PtrSize(p.vPats) < abc.Abc_NtkObjNumMax(pAig)+1 {
		Vec_PtrFree(p.vPats)
		p.vPats = Vec_PtrAllocSimInfo(abc.Abc_NtkObjNumMax(pAig)+1, p.nWordsIn)
	}
	if Vec_PtrSize(p.vPats0) < nTruePis {
		Vec_PtrFree(p.vPats0)
		p.vPats0 = Vec_PtrAllocSimInfo(nTruePis, p.nWords)
	}
	if Vec_PtrSize(p.vPats1) < nTruePis {
		Vec_PtrFree(p.vPats1)
		p.vPats1 = Vec_PtrAllocSimInfo(nTruePis, p.nWords)
	}
	if Vec_PtrSize(p.vOuts) < abc.Abc_NtkPoNum(pAig) {
		Vec_PtrFree(p.vOuts)
		p.vOuts = Vec_PtrAllocSimInfo(abc.Abc_NtkPoNum(pAig), p.nWordsOut)
	}
	// clean storage info for patterns
	Abc_InfoClear(resWords(Vec_PtrEntryAny(p.vPats0, 0)), p.nWords*nTruePis)
	Abc_InfoClear(resWords(Vec_PtrEntryAny(p.vPats1, 0)), p.nWords*nTruePis)
	p.nPats0 = 0
	p.nPats1 = 0
	p.fConst0 = 0
	p.fConst1 = 0
}

/*
*Function*************************************************************

	Synopsis    [Free simulation engine.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimFree(p *Res_Sim_t) {
	Vec_PtrFree(p.vPats)
	Vec_PtrFree(p.vPats0)
	Vec_PtrFree(p.vPats1)
	Vec_PtrFree(p.vOuts)
	Vec_VecFree(p.vCands)
	p = nil
}

/*
*Function*************************************************************

	Synopsis    [Sets random PI simulation info.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/

/*
*Function*************************************************************

	Synopsis    [Sets random PI simulation info.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimSetRandomBytes(p *Res_Sim_t) {
	var pObj *Abc_Obj_t
	var pInfo []uint32
	var i int32
	// inlining of macro Abc_NtkForEachPi
	for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
		pObj = abc.Abc_NtkPi(p.pAig, i)
		pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
		if i < p.nTruePis {
			Abc_InfoRandomBytes(pInfo, p.nWordsIn, p.rng)
		} else {
			Abc_InfoRandom(pInfo, p.nWordsIn)
		}
	}
	/*
	   // double-check that all are byte-patterns
	   Abc_NtkForEachPi( p->pAig, pObj, i )

	   	{
	   	    if ( i == p->nTruePis )
	   	        break;
	   	    pInfoC = resWords(Vec_PtrEntryAny( p->vPats, pObj->Id ));
	   	    for ( k = 0; k < p->nBytesIn; k++ )
	   	        assert( pInfoC[k] == 0 || pInfoC[k] == 0xff );
	   	}
	*/
}

/*
*Function*************************************************************

	Synopsis    [Sets random PI simulation info.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimSetDerivedBytes(p *Res_Sim_t, fUseWalk int32) {
	var vPatsSource [2]*Vec_Ptr_t
	var nPatsSource [2]int32
	var pObj *Abc_Obj_t
	var pInfo []uint32
	var i, k, z, s, nPats int32

	// set several random patterns
	assert(p.nBytesIn%32 == 0, "p.nBytesIn % 32 == 0")
	nPats = p.nBytesIn / 8
	// inlining of macro Abc_NtkForEachPi
	for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
		pObj = abc.Abc_NtkPi(p.pAig, i)
		if i == p.nTruePis {
			break
		}
		Abc_InfoRandomBytes(resWords(Vec_PtrEntryAny(p.vPats, pObj.ID)), nPats/4, p.rng)
	}

	// set special patterns
	if fUseWalk != 0 {
		for z = 0; z < 2; z++ {
			// set the zero pattern
			// inlining of macro Abc_NtkForEachPi
			for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
				pObj = abc.Abc_NtkPi(p.pAig, i)
				if i == p.nTruePis {
					break
				}
				pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
				value := byte(0)
				if z != 0 {
					value = 0xff
				}
				setResByte(pInfo, nPats, value)
			}
			nPats++
			if nPats == p.nBytesIn {
				return
			}
			// set the walking zero pattern
			for k = 0; k < p.nTruePis; k++ {
				// inlining of macro Abc_NtkForEachPi
				for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
					pObj = abc.Abc_NtkPi(p.pAig, i)
					if i == p.nTruePis {
						break
					}
					pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
					value := (i == k) != (z != 0)
					if value {
						setResByte(pInfo, nPats, 0xff)
					} else {
						setResByte(pInfo, nPats, 0)
					}
				}
				nPats++
				if nPats == p.nBytesIn {
					return
				}
			}
		}
	}

	// decide what patterns to set first
	if p.nPats0 < p.nPats1 {
		nPatsSource[0] = p.nPats0
		vPatsSource[0] = p.vPats0
		nPatsSource[1] = p.nPats1
		vPatsSource[1] = p.vPats1

	} else {
		nPatsSource[0] = p.nPats1
		vPatsSource[0] = p.vPats1
		nPatsSource[1] = p.nPats0
		vPatsSource[1] = p.vPats0
	}
	for z = 0; z < 2; z++ {
		for s = nPatsSource[z] - 1; s >= 0; s-- {
			//            if ( s == 0 )
			//            printf( "Patterns:\n" );
			// set the given source pattern
			for k = 0; k < p.nTruePis; k++ {
				// inlining of macro Abc_NtkForEachPi
				for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
					pObj = abc.Abc_NtkPi(p.pAig, i)
					if i == p.nTruePis {
						break
					}
					pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
					if (i == k) != Abc_InfoHasBit(resWords(Vec_PtrEntryAny(vPatsSource[z], i)), s) {
						setResByte(pInfo, nPats, 0xff)
						//                        if ( s == 0 )
						//                        printf( "1" );

					} else {
						setResByte(pInfo, nPats, 0)
						//                        if ( s == 0 )
						//                        printf( "0" );
					}
				}
				//                if ( s == 0 )
				//                printf( "\n" );
				nPats++
				if nPats == p.nBytesIn {
					return
				}
			}
		}
	}
	// clean the rest
	for z = nPats; z < p.nBytesIn; z++ {
		// inlining of macro Abc_NtkForEachPi
		for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
			pObj = abc.Abc_NtkPi(p.pAig, i)
			if i == p.nTruePis {
				break
			}
			pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
			clearResBytes(pInfo, nPats, p.nBytesIn)
		}
	}
	/*
	   // double-check that all are byte-patterns
	   Abc_NtkForEachPi( p->pAig, pObj, i )

	   	{
	   	    if ( i == p->nTruePis )
	   	        break;
	   	    pInfo = resWords(Vec_PtrEntryAny( p->vPats, pObj->Id ));
	   	    for ( k = 0; k < p->nBytesIn; k++ )
	   	        assert( pInfo[k] == 0 || pInfo[k] == 0xff );
	   	}
	*/
}

/*
*Function*************************************************************

	Synopsis    [Sets given PI simulation info.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimSetGiven(p *Res_Sim_t, vInfo *Vec_Ptr_t) {
	var pObj *Abc_Obj_t
	var pInfo, pInfo2 []uint32
	var i, w int32
	// inlining of macro Abc_NtkForEachPi
	for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
		pObj = abc.Abc_NtkPi(p.pAig, i)
		if i == p.nTruePis {
			break
		}
		pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
		pInfo2 = resWords(Vec_PtrEntryAny(vInfo, i))
		for w = 0; w < p.nWords; w++ {
			pInfo[w] = pInfo2[w]
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Simulates one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimPerformOne(pNode *Abc_Obj_t, vSimInfo *Vec_Ptr_t, nSimWords int32) {
	var pInfo, pInfo1, pInfo2 []uint32
	var k, fComp1, fComp2 int32
	// simulate the internal nodes
	assert(abc.Abc_ObjIsNode(pNode) != 0, "abc.Abc_ObjIsNode(pNode)")
	pInfo = resWords(Vec_PtrEntryAny(vSimInfo, pNode.ID))
	pInfo1 = resWords(Vec_PtrEntryAny(vSimInfo, abc.Abc_ObjFanin(pNode, 0).ID))
	pInfo2 = resWords(Vec_PtrEntryAny(vSimInfo, abc.Abc_ObjFanin(pNode, 1).ID))
	fComp1 = abc.Abc_ObjFaninC0(pNode)
	fComp2 = abc.Abc_ObjFaninC1(pNode)
	if (fComp1 != 0) && (fComp2 != 0) {
		for k = 0; k < nSimWords; k++ {
			pInfo[k] = ^pInfo1[k] & ^pInfo2[k]
		}
	} else if (fComp1 != 0) && !(fComp2 != 0) {
		for k = 0; k < nSimWords; k++ {
			pInfo[k] = ^pInfo1[k] & pInfo2[k]
		}
	} else if !(fComp1 != 0) && (fComp2 != 0) {
		for k = 0; k < nSimWords; k++ {
			pInfo[k] = pInfo1[k] & ^pInfo2[k]
		}
	} else { // if ( fComp1 && fComp2 )
		for k = 0; k < nSimWords; k++ {
			pInfo[k] = pInfo1[k] & pInfo2[k]
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Simulates one CO node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimTransferOne(pNode *Abc_Obj_t, vSimInfo *Vec_Ptr_t, nSimWords int32) {
	var pInfo, pInfo1 []uint32
	var k, fComp1 int32
	// simulate the internal nodes
	assert(abc.Abc_ObjIsCo(pNode) != 0, "abc.Abc_ObjIsCo(pNode)")
	pInfo = resWords(Vec_PtrEntryAny(vSimInfo, pNode.ID))
	pInfo1 = resWords(Vec_PtrEntryAny(vSimInfo, abc.Abc_ObjFanin(pNode, 0).ID))
	fComp1 = abc.Abc_ObjFaninC0(pNode)
	if fComp1 != 0 {
		for k = 0; k < nSimWords; k++ {
			pInfo[k] = ^pInfo1[k]
		}
	} else {
		for k = 0; k < nSimWords; k++ {
			pInfo[k] = pInfo1[k]
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs one round of simulation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimPerformRound(p *Res_Sim_t, nWords int32) {
	var pObj *Abc_Obj_t
	var i int32
	Abc_InfoFill(resWords(Vec_PtrEntryAny(p.vPats, 0)), nWords)
	// inlining of macro Abc_AigForEachAnd
	for i = 0; i < abc.Abc_NtkObjNumMax(p.pAig); i++ {
		pObj = abc.Abc_NtkObj(p.pAig, i)
		if (pObj) == nil || abc.Abc_AigNodeIsAnd(pObj) == 0 {
		} else {
			Res_SimPerformOne(pObj, p.vPats, nWords)
		}
	}
	// inlining of macro Abc_NtkForEachPo
	for i = 0; i < abc.Abc_NtkPoNum(p.pAig); i++ {
		pObj = abc.Abc_NtkPo(p.pAig, i)
		Res_SimTransferOne(pObj, p.vPats, nWords)
	}
}

/*
*Function*************************************************************

	Synopsis    [Pads the extra space with duplicated simulation info.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimPadSimInfo(vPats *Vec_Ptr_t, nPats int32, nWords int32) {
	var pInfo []uint32
	var i, w, iWords int32
	assert(nPats > 0 && nPats < nWords*32, "nPats > 0 && nPats < nWords * 32")
	// pad the first word
	if nPats < 32 {
		// inlining of macro Vec_PtrForEachEntry
		for i = 0; i < Vec_PtrSize(vPats); i++ {
			pInfo = resWords(Vec_PtrEntryAny(vPats, i))
			if pInfo[0]&1 != 0 {
				pInfo[0] |= (^uint32(0)) << uint(nPats)
			}
		}
		nPats = 32
	}
	// pad the empty words
	iWords = nPats / 32
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < Vec_PtrSize(vPats); i++ {
		pInfo = resWords(Vec_PtrEntryAny(vPats, i))
		for w = iWords; w < nWords; w++ {
			pInfo[w] = pInfo[0]
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Duplicates the simulation info to fill the space.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimDeriveInfoReplicate(p *Res_Sim_t) {
	var pInfo, pInfo2 []uint32
	var pObj *Abc_Obj_t
	var i, j, w int32
	// inlining of macro Abc_NtkForEachPo
	for i = 0; i < abc.Abc_NtkPoNum(p.pAig); i++ {
		pObj = abc.Abc_NtkPo(p.pAig, i)
		pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
		pInfo2 = resWords(Vec_PtrEntryAny(p.vOuts, i))
		for j = 0; j < p.nPats; j++ {
			for w = 0; w < p.nWords; w++ {
				pInfo2[j*p.nWords+w] = pInfo[w]
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Complement the simulation info if necessary.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimDeriveInfoComplement(p *Res_Sim_t) {
	var pInfo, pInfo2 []uint32
	var pObj *Abc_Obj_t
	var i, j, w int32
	// inlining of macro Abc_NtkForEachPo
	for i = 0; i < abc.Abc_NtkPoNum(p.pAig); i++ {
		pObj = abc.Abc_NtkPo(p.pAig, i)
		pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
		pInfo2 = resWords(Vec_PtrEntryAny(p.vOuts, i))
		for j = 0; j < p.nPats; j++ {
			if Abc_InfoHasBit(pInfo, j) {
				for w = 0; w < p.nWords; w++ {
					index := j*p.nWords + w
					pInfo2[index] = ^pInfo2[index]
				}
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints output patterns.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimPrintOutPatterns(p *Res_Sim_t, pAig *Abc_Ntk_t) {
	var pInfo2 []uint32
	var i int32
	// inlining of macro Abc_NtkForEachPo
	for i = 0; i < abc.Abc_NtkPoNum(pAig); i++ {
		pInfo2 = resWords(Vec_PtrEntryAny(p.vOuts, i))
		extra.Extra_PrintBinary(os.Stdout, pInfo2, p.nPatsOut)
		fmt.Printf("\n")
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints output patterns.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimPrintNodePatterns(p *Res_Sim_t, pAig *Abc_Ntk_t) {
	var pInfo []uint32
	pInfo = resWords(Vec_PtrEntryAny(p.vPats, abc.Abc_NtkPo(p.pAig, 1).ID))
	extra.Extra_PrintBinary(os.Stdout, pInfo, p.nPats)
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of patters of different type.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimCountResults(p *Res_Sim_t, pnDcs *int32, pnOnes *int32, pnZeros *int32, fVerbose int32) {
	var pInfoCare, pInfoNode []uint32
	var i, nTotal int32
	pInfoCare = resWords(Vec_PtrEntryAny(p.vPats, abc.Abc_NtkPo(p.pAig, 0).ID))
	pInfoNode = resWords(Vec_PtrEntryAny(p.vPats, abc.Abc_NtkPo(p.pAig, 1).ID))
	for i = 0; i < p.nBytesIn; i++ {
		if resByte(pInfoCare, i) == 0 {
			(*pnDcs)++
		} else if resByte(pInfoNode, i) == 0 {
			(*pnZeros)++
		} else {
			(*pnOnes)++
		}
	}
	nTotal += *pnDcs
	nTotal += *pnZeros
	nTotal += *pnOnes
	if fVerbose != 0 {
		fmt.Printf("Dc = %7.2f %%  ", 100.0*float64(*pnDcs)/float64(nTotal))
		fmt.Printf("On = %7.2f %%  ", 100.0*float64(*pnOnes)/float64(nTotal))
		fmt.Printf("Off = %7.2f %%  ", 100.0*float64(*pnZeros)/float64(nTotal))
	}
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of patters of different type.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimCollectPatterns(p *Res_Sim_t, fVerbose int32) {
	var pObj *Abc_Obj_t
	var pInfoCare, pInfoNode, pInfo []uint32
	var i, j int32
	pInfoCare = resWords(Vec_PtrEntryAny(p.vPats, abc.Abc_NtkPo(p.pAig, 0).ID))
	pInfoNode = resWords(Vec_PtrEntryAny(p.vPats, abc.Abc_NtkPo(p.pAig, 1).ID))
	for i = 0; i < p.nBytesIn; i++ {
		// skip don't-care patterns
		if resByte(pInfoCare, i) == 0 {
			continue
		}
		// separate offset and onset patterns
		assert(resByte(pInfoNode, i) == 0 || resByte(pInfoNode, i) == 0xff, "pInfoNode[i] == 0 || pInfoNode[i] == 0xff")
		if resByte(pInfoNode, i) == 0 {
			if p.nPats0 >= p.nPats {
				continue
			}
			// inlining of macro Abc_NtkForEachPi
			for j = 0; j < abc.Abc_NtkPiNum(p.pAig); j++ {
				pObj = abc.Abc_NtkPi(p.pAig, j)
				if j == p.nTruePis {
					break
				}
				pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
				assert(resByte(pInfo, i) == 0 || resByte(pInfo, i) == 0xff, "pInfo[i] == 0 || pInfo[i] == 0xff")
				if resByte(pInfo, i) != 0 {
					Abc_InfoSetBit(resWords(Vec_PtrEntryAny(p.vPats0, j)), p.nPats0)
				}
			}
			p.nPats0++

		} else {
			if p.nPats1 >= p.nPats {
				continue
			}
			// inlining of macro Abc_NtkForEachPi
			for j = 0; j < abc.Abc_NtkPiNum(p.pAig); j++ {
				pObj = abc.Abc_NtkPi(p.pAig, j)
				if j == p.nTruePis {
					break
				}
				pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
				assert(resByte(pInfo, i) == 0 || resByte(pInfo, i) == 0xff, "pInfo[i] == 0 || pInfo[i] == 0xff")
				if resByte(pInfo, i) != 0 {
					Abc_InfoSetBit(resWords(Vec_PtrEntryAny(p.vPats1, j)), p.nPats1)
				}
			}
			p.nPats1++
		}
		if p.nPats0 >= p.nPats && p.nPats1 >= p.nPats {
			break
		}
	}
	if fVerbose != 0 {
		fmt.Printf("|  ")
		fmt.Printf("On = %3d  ", p.nPats1)
		fmt.Printf("Off = %3d  ", p.nPats0)
		fmt.Printf("\n")
	}
}

/*
*Function*************************************************************

	Synopsis    [Verifies the last pattern.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimVerifyValue(p *Res_Sim_t, fOnSet int32) int32 {
	var pObj *Abc_Obj_t
	var pInfo, pInfo2 []uint32
	var i int32
	var value bool
	// inlining of macro Abc_NtkForEachPi
	for i = 0; i < abc.Abc_NtkPiNum(p.pAig); i++ {
		pObj = abc.Abc_NtkPi(p.pAig, i)
		if i == p.nTruePis {
			break
		}
		if fOnSet != 0 {
			pInfo2 = resWords(Vec_PtrEntryAny(p.vPats1, i))
			value = Abc_InfoHasBit(pInfo2, p.nPats1-1)

		} else {
			pInfo2 = resWords(Vec_PtrEntryAny(p.vPats0, i))
			value = Abc_InfoHasBit(pInfo2, p.nPats0-1)
		}
		pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
		if value {
			pInfo[0] = ^uint32(0)
		} else {
			pInfo[0] = 0
		}
	}
	Res_SimPerformRound(p, 1)
	pObj = abc.Abc_NtkPo(p.pAig, 1)
	pInfo = resWords(Vec_PtrEntryAny(p.vPats, pObj.ID))
	assert(pInfo[0] == 0 || pInfo[0] == ^uint32(0), "pInfo[0] == 0 || pInfo[0] == ^uint32(0)")
	if pInfo[0] > 0 {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Prepares simulation info for candidate filtering.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Res_SimPrepare(p *Res_Sim_t, pAig *Abc_Ntk_t, nTruePis int32, fVerbose int32) int32 {
	var i, nOnes, nZeros, nDcs int32
	if fVerbose != 0 {
		fmt.Printf("\n")
	}
	// prepare the manager
	Res_SimAdjust(p, pAig, nTruePis)
	// estimate the number of patterns
	Res_SimSetRandomBytes(p)
	Res_SimPerformRound(p, p.nWordsIn)
	Res_SimCountResults(p, &nDcs, &nOnes, &nZeros, fVerbose)
	// collect the patterns
	Res_SimCollectPatterns(p, fVerbose)
	// add more patterns using constraint simulation
	if p.nPats0 < 8 {
		if Res_SatSimulate(p, 16, 0) == 0 {
			if p.fConst0 != 0 || p.fConst1 != 0 {
				return 1
			}
			return 0
		}
		//            return 0;
		//        printf( "Value0 = %d\n", Res_SimVerifyValue( p, 0 ) );
	}
	if p.nPats1 < 8 {
		if Res_SatSimulate(p, 16, 1) == 0 {
			if p.fConst0 != 0 || p.fConst1 != 0 {
				return 1
			}
			return 0
		}
		//            return 0;
		//        printf( "Value1 = %d\n", Res_SimVerifyValue( p, 1 ) );
	}
	// generate additional patterns
	for i = 0; i < 2; i++ {
		if p.nPats0 > p.nPats*7/8 && p.nPats1 > p.nPats*7/8 {
			break
		}
		walk := int32(0)
		if i == 0 {
			walk = 1
		}
		Res_SimSetDerivedBytes(p, walk)
		Res_SimPerformRound(p, p.nWordsIn)
		Res_SimCountResults(p, &nDcs, &nOnes, &nZeros, fVerbose)
		Res_SimCollectPatterns(p, fVerbose)
	}
	// create bit-matrix info
	if p.nPats0 < p.nPats {
		Res_SimPadSimInfo(p.vPats0, p.nPats0, p.nWords)
	}
	if p.nPats1 < p.nPats {
		Res_SimPadSimInfo(p.vPats1, p.nPats1, p.nWords)
	}
	// resimulate 0-patterns
	Res_SimSetGiven(p, p.vPats0)
	Res_SimPerformRound(p, p.nWords)
	//Res_SimPrintNodePatterns( p, pAig );
	Res_SimDeriveInfoReplicate(p)
	// resimulate 1-patterns
	Res_SimSetGiven(p, p.vPats1)
	Res_SimPerformRound(p, p.nWords)
	//Res_SimPrintNodePatterns( p, pAig );
	Res_SimDeriveInfoComplement(p)
	// print output patterns
	//    Res_SimPrintOutPatterns( p, pAig );
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
