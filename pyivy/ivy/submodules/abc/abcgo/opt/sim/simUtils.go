package sim

import "fmt"

/**CFile****************************************************************

  FileName    [simUtils.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Various simulation utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: simUtils.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
var bit_count [256]int32 = [256]int32{
  0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
};

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates simulation information for all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoAlloc(nSize int32, nWords int32, fClean int32) *Vec_Ptr_t {
    var vInfo *Vec_Ptr_t;
    var i int32;
    assert(nSize > 0 && nWords > 0, "nSize > 0 && nWords > 0");
    vInfo = Vec_PtrAlloc( nSize );
    // inlining of macro ABC_ALLOC
    vInfo.pArray[0] = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nSize * nWords)));
    if fClean != 0 {
        memset( vInfo.pArray[0], 0, sizeof(unsigned) * nSize * nWords ); }
    for  i = 1; i < nSize; i++  {
        vInfo.pArray[i] = ((unsigned *)vInfo.pArray[i-1]) + nWords;
    }
    vInfo.nSize = nSize;
    return vInfo;
}

/**Function*************************************************************

  Synopsis    [Allocates simulation information for all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoFree(p *Vec_Ptr_t) {
    p.pArray[0] = nil;
    Vec_PtrFree( p );
}

/**Function*************************************************************

  Synopsis    [Adds the second supp-info the first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoAdd(pInfo1 *uint32, pInfo2 *uint32, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pInfo1[w] |= pInfo2[w];
    }
}

/**Function*************************************************************

  Synopsis    [Returns the positions where pInfo2 is 1 while pInfo1 is 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoDetectDiffs(pInfo1 *uint32, pInfo2 *uint32, nWords int32, vDiffs *Vec_Int_t) {
    var w, b int32;
    var uMask uint32;
    vDiffs.nSize = 0;
    for  w = 0; w < nWords; w++  {
        if uMask = (pInfo2[w] ^ pInfo1[w]) {
            for  b = 0; b < 32; b++  {
                if uMask & (1 << b) {
                    Vec_IntPush( vDiffs, 32*w + b ); }
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Returns the positions where pInfo2 is 1 while pInfo1 is 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoDetectNews(pInfo1 *uint32, pInfo2 *uint32, nWords int32, vDiffs *Vec_Int_t) {
    var w, b int32;
    var uMask uint32;
    vDiffs.nSize = 0;
    for  w = 0; w < nWords; w++  {
        if uMask = (pInfo2[w] & ~pInfo1[w]) {
            for  b = 0; b < 32; b++  {
                if uMask & (1 << b) {
                    Vec_IntPush( vDiffs, 32*w + b ); }
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Flips the simulation info of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoFlip(p *Sim_Man_t, pNode *Abc_Obj_t) {
    var pSimInfo1, pSimInfo2 *uint32;
    var k int32;
    pSimInfo1 = (unsigned *)p.vSim0.pArray[pNode.Id];
    pSimInfo2 = (unsigned *)p.vSim1.pArray[pNode.Id];
    for  k = 0; k < p.nSimWords; k++  {
        pSimInfo2[k] = ~pSimInfo1[k];
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the simulation infos are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoCompare(p *Sim_Man_t, pNode *Abc_Obj_t) int32 {
    var pSimInfo1, pSimInfo2 *uint32;
    var k int32;
    pSimInfo1 = (unsigned *)p.vSim0.pArray[pNode.Id];
    pSimInfo2 = (unsigned *)p.vSim1.pArray[pNode.Id];
    for  k = 0; k < p.nSimWords; k++  {
        if pSimInfo2[k] != pSimInfo1[k] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Simulates the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilSimulate(p *Sim_Man_t, fType int32) {
    var pNode *Abc_Obj_t;
    var i int32;
    // simulate the internal nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pNode) = Abc_NtkObj(p.pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Sim_UtilSimulateNode( p, pNode, fType, fType, fType ); }
    }
    // assign simulation info of the CO nodes
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pNode) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        Sim_UtilSimulateNode( p, pNode, fType, fType, fType );
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilSimulateNode(p *Sim_Man_t, pNode *Abc_Obj_t, fType int32, fType1 int32, fType2 int32) {
    var pSimmNode, pSimmNode1, pSimmNode2 *uint32;
    var k, fComp1, fComp2 int32;
    // simulate the internal nodes
    if Abc_ObjIsNode(pNode) {
        if fType != 0 {
            pSimmNode  = (unsigned *)p.vSim1.pArray[ pNode.Id ];
        } else {
            pSimmNode  = (unsigned *)p.vSim0.pArray[ pNode.Id ]; }

        if fType1 != 0 {
            pSimmNode1 = (unsigned *)p.vSim1.pArray[ Abc_ObjFaninId0(pNode) ];
        } else {
            pSimmNode1 = (unsigned *)p.vSim0.pArray[ Abc_ObjFaninId0(pNode) ]; }

        if fType2 != 0 {
            pSimmNode2 = (unsigned *)p.vSim1.pArray[ Abc_ObjFaninId1(pNode) ];
        } else {
            pSimmNode2 = (unsigned *)p.vSim0.pArray[ Abc_ObjFaninId1(pNode) ]; }

        fComp1 = Abc_ObjFaninC0(pNode);
        fComp2 = Abc_ObjFaninC1(pNode);
        if (fComp1 != 0) && (fComp2 != 0) {
            for  k = 0; k < p.nSimWords; k++  {
                pSimmNode[k] = ~pSimmNode1[k] & ~pSimmNode2[k];
            }
        } else if (fComp1 != 0) && !(fComp2 != 0) {
            for  k = 0; k < p.nSimWords; k++  {
                pSimmNode[k] = ~pSimmNode1[k] &  pSimmNode2[k];
            }
        } else if !(fComp1 != 0) && (fComp2 != 0) {
            for  k = 0; k < p.nSimWords; k++  {
                pSimmNode[k] =  pSimmNode1[k] & ~pSimmNode2[k];
            }
        } else { // if ( fComp1 && fComp2 )
            for  k = 0; k < p.nSimWords; k++  {
                pSimmNode[k] =  pSimmNode1[k] &  pSimmNode2[k];
            } }
    
    } else {
        assert(Abc_ObjFaninNum(pNode) == 1, "Abc_ObjFaninNum(pNode) == 1");
        if fType != 0 {
            pSimmNode  = (unsigned *)p.vSim1.pArray[ pNode.Id ];
        } else {
            pSimmNode  = (unsigned *)p.vSim0.pArray[ pNode.Id ]; }

        if fType1 != 0 {
            pSimmNode1 = (unsigned *)p.vSim1.pArray[ Abc_ObjFaninId0(pNode) ];
        } else {
            pSimmNode1 = (unsigned *)p.vSim0.pArray[ Abc_ObjFaninId0(pNode) ]; }

        fComp1 = Abc_ObjFaninC0(pNode);
        if fComp1 != 0 {
            for  k = 0; k < p.nSimWords; k++  {
                pSimmNode[k] = ~pSimmNode1[k];
            }
        } else { 
            for  k = 0; k < p.nSimWords; k++  {
                pSimmNode[k] =  pSimmNode1[k];
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilSimulateNodeOne(pNode *Abc_Obj_t, vSimInfo *Vec_Ptr_t, nSimWords int32, nOffset int32) {
    var pSimmNode, pSimmNode1, pSimmNode2 *uint32;
    var k, fComp1, fComp2 int32;
    // simulate the internal nodes
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    pSimmNode  = (unsigned *)Vec_PtrEntry(vSimInfo, pNode.Id);
    pSimmNode1 = (unsigned *)Vec_PtrEntry(vSimInfo, Abc_ObjFaninId0(pNode));
    pSimmNode2 = (unsigned *)Vec_PtrEntry(vSimInfo, Abc_ObjFaninId1(pNode));
    pSimmNode  += nOffset;
    pSimmNode1 += nOffset;
    pSimmNode2 += nOffset;
    fComp1 = Abc_ObjFaninC0(pNode);
    fComp2 = Abc_ObjFaninC1(pNode);
    if (fComp1 != 0) && (fComp2 != 0) {
        for  k = 0; k < nSimWords; k++  {
            pSimmNode[k] = ~pSimmNode1[k] & ~pSimmNode2[k];
        }
    } else if (fComp1 != 0) && !(fComp2 != 0) {
        for  k = 0; k < nSimWords; k++  {
            pSimmNode[k] = ~pSimmNode1[k] &  pSimmNode2[k];
        }
    } else if !(fComp1 != 0) && (fComp2 != 0) {
        for  k = 0; k < nSimWords; k++  {
            pSimmNode[k] =  pSimmNode1[k] & ~pSimmNode2[k];
        }
    } else { // if ( fComp1 && fComp2 )
        for  k = 0; k < nSimWords; k++  {
            pSimmNode[k] =  pSimmNode1[k] &  pSimmNode2[k];
        } }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilTransferNodeOne(pNode *Abc_Obj_t, vSimInfo *Vec_Ptr_t, nSimWords int32, nOffset int32, fShift int32) {
    var pSimmNode, pSimmNode1 *uint32;
    var k, fComp1 int32;
    // simulate the internal nodes
    assert(Abc_ObjIsCo(pNode), "Abc_ObjIsCo(pNode)");
    pSimmNode  = (unsigned *)Vec_PtrEntry(vSimInfo, pNode.Id);
    pSimmNode1 = (unsigned *)Vec_PtrEntry(vSimInfo, Abc_ObjFaninId0(pNode));
    pSimmNode  += nOffset + (fShift > 0)*nSimWords;
    pSimmNode1 += nOffset;
    fComp1 = Abc_ObjFaninC0(pNode);
    if fComp1 != 0 {
        for  k = 0; k < nSimWords; k++  {
            pSimmNode[k] = ~pSimmNode1[k];
        }
    } else { 
        for  k = 0; k < nSimWords; k++  {
            pSimmNode[k] =  pSimmNode1[k];
        } }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the simulation infos are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountSuppSizes(p *Sim_Man_t, fStruct int32) int32 {
    var pNode, pNodeCi *Abc_Obj_t;
    var i, v, Counter int32;
    Counter = 0;
    if fStruct != 0 {
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pNode) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
            // inlining of macro Abc_NtkForEachCi
            for  v = 0; (v < Abc_NtkCiNum(p.pNtk)) && (((pNodeCi) = Abc_NtkCi(p.pNtk, v)), 1); v++  {
                // inlining of macro Sim_SuppStrHasVar
                // inlining of macro Sim_HasBit
                Counter += ((((unsigned*)(p.vSuppStr).pArray[(pNode).Id])[((v))>>5]  & (1<<(((v)) & 31))) > 0);
            }
        }
    
    } else {
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pNode) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
            // inlining of macro Abc_NtkForEachCi
            for  v = 0; (v < Abc_NtkCiNum(p.pNtk)) && (((pNodeCi) = Abc_NtkCi(p.pNtk, v)), 1); v++  {
                // inlining of macro Sim_SuppFunHasVar
                // inlining of macro Sim_HasBit
                Counter += ((((unsigned*)(p.vSuppFun).pArray[i])[((v))>>5]  & (1<<(((v)) & 31))) > 0);
            }
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of 1's in the bitstring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountOnes(pSimInfo *uint32, nSimWords int32) int32 {
    unsigned char * pBytes;
    var nOnes, nBytes, i int32;
    pBytes = (unsigned char *)pSimInfo;
    nBytes = 4 * nSimWords;
    nOnes  = 0;
    for  i = 0; i < nBytes; i++  {
        nOnes += bit_count[ pBytes[i] ];
    }
    return nOnes;    
}

/**Function*************************************************************

  Synopsis    [Counts the number of 1's in the bitstring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountOnesArray(vInfo *Vec_Ptr_t, nSimWords int32) *Vec_Int_t {
    var vCounters *Vec_Int_t;
    var pSimInfo *uint32;
    var i int32;
    vCounters = Vec_IntStart( Vec_PtrSize(vInfo) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vInfo)) && (((pSimInfo) = (unsigned *)Vec_PtrEntry(vInfo, i)), 1); i++  {
        Vec_IntWriteEntry( vCounters, i, Sim_UtilCountOnes(pSimInfo, nSimWords) );
    }
    return vCounters;
}

/**Function*************************************************************

  Synopsis    [Returns random patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilSetRandom(pPatRand *uint32, nSimWords int32) {
    var k int32;
    for  k = 0; k < nSimWords; k++  {
        pPatRand[k] = SIM_RANDOM_UNSIGNED;
    }
}

/**Function*************************************************************

  Synopsis    [Returns complemented patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilSetCompl(pPatRand *uint32, nSimWords int32) {
    var k int32;
    for  k = 0; k < nSimWords; k++  {
        pPatRand[k] = ~pPatRand[k];
    }
}

/**Function*************************************************************

  Synopsis    [Returns constant patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilSetConst(pPatRand *uint32, nSimWords int32, fConst1 int32) {
    var k int32;
    for  k = 0; k < nSimWords; k++  {
        pPatRand[k] = 0;
    }
    if fConst1 != 0 {
        Sim_UtilSetCompl( pPatRand, nSimWords ); }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoIsEqual(pPats1 *uint32, pPats2 *uint32, nSimWords int32) int32 {
    var k int32;
    for  k = 0; k < nSimWords; k++  {
        if pPats1[k] != pPats2[k] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if Node1 implies Node2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoIsImp(pPats1 *uint32, pPats2 *uint32, nSimWords int32) int32 {
    var k int32;
    for  k = 0; k < nSimWords; k++  {
        if pPats1[k] & ~pPats2[k] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if Node1 v Node2 is always true.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilInfoIsClause(pPats1 *uint32, pPats2 *uint32, nSimWords int32) int32 {
    var k int32;
    for  k = 0; k < nSimWords; k++  {
        if ~pPats1[k] & ~pPats2[k] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Counts the total number of pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountAllPairs(vSuppFun *Vec_Ptr_t, nSimWords int32, vCounters *Vec_Int_t) int32 {
    var pSupp *uint32;
    var Counter, nOnes, nPairs, i int32;
    Counter = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuppFun)) && (((pSupp) = (unsigned *)Vec_PtrEntry(vSuppFun, i)), 1); i++  {
        nOnes  = Sim_UtilCountOnes( pSupp, nSimWords );
        nPairs = nOnes * (nOnes - 1) / 2;
        Vec_IntWriteEntry( vCounters, i, nPairs );
        Counter += nPairs;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of entries in the array of matrices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountPairsOne(pMat *Extra_BitMat_t, vSupport *Vec_Int_t) int32 {
    var i, k, Index1, Index2 int32;
    var Counter int32;
//    int Counter2;
    // inlining of macro Vec_IntForEachEntry
    for  Index1 = 0; (Index1 < Vec_IntSize(vSupport)) && (((i) = Vec_IntEntry(vSupport, Index1)), 1); Index1++  {
    // inlining of macro Vec_IntForEachEntryStart
    for  Index2 = Index1+1; (Index2 < Vec_IntSize(vSupport)) && (((k) = Vec_IntEntry(vSupport, Index2)), 1); Index2++  {
        Counter += Extra_BitMatrixLookup1( pMat, i, k );
    }
    }
//    Counter2 = Extra_BitMatrixCountOnesUpper(pMat);
//    assert( Counter == Counter2 );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of entries in the array of matrices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountPairsOnePrint(pMat *Extra_BitMat_t, vSupport *Vec_Int_t) int32 {
    var i, k, Index1, Index2 int32;
    // inlining of macro Vec_IntForEachEntry
    for  Index1 = 0; (Index1 < Vec_IntSize(vSupport)) && (((i) = Vec_IntEntry(vSupport, Index1)), 1); Index1++  {
    // inlining of macro Vec_IntForEachEntryStart
    for  Index2 = Index1+1; (Index2 < Vec_IntSize(vSupport)) && (((k) = Vec_IntEntry(vSupport, Index2)), 1); Index2++  {
        if Extra_BitMatrixLookup1( pMat, i, k ) {
            fmt.Printf("(%d,%d) ", i, k); }
    }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Counts the number of entries in the array of matrices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountPairsAllPrint(p *Sym_Man_t) {
    var i int32;
    var clk abctime;
clk = Abc_Clock();
    for  i = 0; i < p.nOutputs; i++  {
        fmt.Printf("Output %2d :", i);
        Sim_UtilCountPairsOnePrint( (Extra_BitMat_t *)Vec_PtrEntry(p.vMatrSymms, i), Vec_VecEntryInt(p.vSupports, i) );
        fmt.Printf("\n");
    }
p.timeCount += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    [Counts the number of entries in the array of matrices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilCountPairsAll(p *Sym_Man_t) {
    var nPairsTotal, nPairsSym, nPairsNonSym, i int32;
    var clk abctime;
clk = Abc_Clock();
    p.nPairsSymm    = 0;
    p.nPairsNonSymm = 0;
    for  i = 0; i < p.nOutputs; i++  {
        nPairsTotal  = Vec_IntEntry(p.vPairsTotal, i);
        nPairsSym    = Vec_IntEntry(p.vPairsSym,   i);
        nPairsNonSym = Vec_IntEntry(p.vPairsNonSym,i);
        assert(nPairsTotal >= nPairsSym + nPairsNonSym, "nPairsTotal >= nPairsSym + nPairsNonSym"); 
        if nPairsTotal == nPairsSym + nPairsNonSym {
            p.nPairsSymm    += nPairsSym;
            p.nPairsNonSymm += nPairsNonSym;
            continue;
        }
        nPairsSym    = Sim_UtilCountPairsOne( (Extra_BitMat_t *)Vec_PtrEntry(p.vMatrSymms,   i), Vec_VecEntryInt(p.vSupports, i) );
        nPairsNonSym = Sim_UtilCountPairsOne( (Extra_BitMat_t *)Vec_PtrEntry(p.vMatrNonSymms,i), Vec_VecEntryInt(p.vSupports, i) );
        assert(nPairsTotal >= nPairsSym + nPairsNonSym, "nPairsTotal >= nPairsSym + nPairsNonSym"); 
        Vec_IntWriteEntry( p.vPairsSym,    i, nPairsSym );
        Vec_IntWriteEntry( p.vPairsNonSym, i, nPairsNonSym );
        p.nPairsSymm    += nPairsSym;
        p.nPairsNonSymm += nPairsNonSym;
//        printf( "%d ", nPairsTotal - nPairsSym - nPairsNonSym );
    }
//printf( "\n" );
    p.nPairsRem = p.nPairsTotal-p.nPairsSymm-p.nPairsNonSymm;
p.timeCount += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilMatrsAreDisjoint(p *Sym_Man_t) int32 {
    var i int32;
    for  i = 0; i < p.nOutputs; i++  {
        if !Extra_BitMatrixIsDisjoint( (Extra_BitMat_t *)Vec_PtrEntry(p.vMatrSymms,i), (Extra_BitMat_t *)Vec_PtrEntry(p.vMatrNonSymms,i) ) {
            return 0; }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

