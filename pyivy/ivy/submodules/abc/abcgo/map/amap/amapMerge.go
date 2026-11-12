package amap

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [amapMerge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    [Computing cuts for the node.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapMerge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates new cut and adds it to storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManSetupPis(p *Amap_Man_t) *Amap_Cut_t {
    var pObj *Amap_Obj_t;
    var pCut *Amap_Cut_t;
    var i int32;
    var nBytes int32 = sizeof(Amap_Cut_t) + sizeof(int);
    // inlining of macro ABC_ALLOC
    var pBuffer string  = ((char *) malloc(sizeof(char) * (size_t)(Amap_ManPiNum(p) * nBytes)));
    // inlining of macro Amap_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pCut = (Amap_Cut_t *)( pBuffer + i*nBytes );
        pCut.iMat = 0;
        pCut.fInv = 0;
        pCut.nFans = 1;
        pCut.Fans[0] = Abc_Var2Lit( pObj.Id, 0 );
        pObj.pData = pCut;
        pObj.nCuts = 1;
        pObj.EstRefs = float32(1.0);
    }
    return (Amap_Cut_t *)pBuffer;
}

/**Function*************************************************************

  Synopsis    [Creates new cut and adds it to storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManCutStore(p *Amap_Man_t, pCut *Amap_Cut_t, fCompl int32) *Amap_Cut_t {
    var pNew *Amap_Cut_t;
    var iFan int32;
    var nBytes int32 = sizeof(Amap_Cut_t) + sizeof(int) * pCut.nFans + sizeof(Amap_Cut_t *);
    pNew = (Amap_Cut_t *)Aig_MmFlexEntryFetch( p.pMemTemp, nBytes );
    pNew.iMat  = pCut.iMat;
    pNew.fInv  = pCut.fInv ^ fCompl;
    pNew.nFans = pCut.nFans;
    memcpy( pNew.Fans, pCut.Fans, sizeof(int) * pCut.nFans );
    // add it to storage
    iFan = Abc_Var2Lit( pNew.iMat, pNew.fInv );
    if p.ppCutsTemp[ iFan ] == nil {
        Vec_IntPushOrder( p.vTemp, iFan ); }
    *Amap_ManCutNextP( pNew ) = p.ppCutsTemp[ iFan ];
    p.ppCutsTemp[ iFan ] = pNew;
    return pNew;
} 

/**Function*************************************************************

  Synopsis    [Creates new cut and adds it to storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManCutCreate(p *Amap_Man_t, pCut0 *Amap_Cut_t, pCut1 *Amap_Cut_t, iMat int32) *Amap_Cut_t {
    var pCut *Amap_Cut_t;
    var i int32;
    var nSize int32 = pCut0.nFans + pCut1.nFans;
    var nBytes int32  = sizeof(Amap_Cut_t) + sizeof(int) * nSize + sizeof(Amap_Cut_t *);
    assert(pCut0.iMat >= pCut1.iMat, "pCut0.iMat >= pCut1.iMat");
    pCut = (Amap_Cut_t *)Aig_MmFlexEntryFetch( p.pMemTemp, nBytes );
    pCut.iMat  = iMat;
    pCut.fInv  = 0;
    pCut.nFans = nSize;
    for  i = 0; i < int32(pCut0.nFans); i++  {
        pCut.Fans[i] = pCut0.Fans[i];
    }
    for  i = 0; i < int32(pCut1.nFans); i++  {
        pCut.Fans[pCut0.nFans+i] = pCut1.Fans[i];
    }
    // add it to storage
    if p.ppCutsTemp[ pCut.iMat ] == nil {
        Vec_IntPushOrder( p.vTemp, pCut.iMat ); }
    *Amap_ManCutNextP( pCut ) = p.ppCutsTemp[ pCut.iMat ];
    p.ppCutsTemp[ pCut.iMat ] = pCut;
    return pCut;
} 

/**Function*************************************************************

  Synopsis    [Creates new cut and adds it to storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManCutCreate3(p *Amap_Man_t, pCut0 *Amap_Cut_t, pCut1 *Amap_Cut_t, pCut2 *Amap_Cut_t, iMat int32) *Amap_Cut_t {
    var pCut *Amap_Cut_t;
    var i int32;
    var nSize int32 = pCut0.nFans + pCut1.nFans + pCut2.nFans;
    var nBytes int32  = sizeof(Amap_Cut_t) + sizeof(int) * nSize + sizeof(Amap_Cut_t *);
    pCut = (Amap_Cut_t *)Aig_MmFlexEntryFetch( p.pMemTemp, nBytes );
    pCut.iMat  = iMat;
    pCut.fInv  = 0;
    pCut.nFans = nSize;
    for  i = 0; i < int32(pCut0.nFans); i++  {
        pCut.Fans[i] = pCut0.Fans[i];
    }
    for  i = 0; i < int32(pCut1.nFans); i++  {
        pCut.Fans[pCut0.nFans+i] = pCut1.Fans[i];
    }
    for  i = 0; i < int32(pCut2.nFans); i++  {
        pCut.Fans[pCut0.nFans+pCut1.nFans+i] = pCut2.Fans[i];
    }
    // add it to storage
    if p.ppCutsTemp[ pCut.iMat ] == nil {
        Vec_IntPushOrder( p.vTemp, pCut.iMat ); }
    *Amap_ManCutNextP( pCut ) = p.ppCutsTemp[ pCut.iMat ];
    p.ppCutsTemp[ pCut.iMat ] = pCut;
    return pCut;
} 

/**Function*************************************************************

  Synopsis    [Removes cuts from the temporary storage.]

  Description [] 
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Amap_ManCutSaveStored(p *Amap_Man_t, pNode *Amap_Obj_t) {
    var nMaxCuts int32  = p.pPars.nCutsMax;
    var pBuffer *int32;
    var pNext, pCut *Amap_Cut_t;
    var i, nWords, Entry, nCuts, nCuts2 int32;
    assert(pNode.pData == nil, "pNode.pData == nil");
    // count memory needed
    nCuts = 1;
    nWords = 2;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vTemp)) && (((Entry) = Vec_IntEntry(p.vTemp, i)), 1); i++  {
        for  pCut = p.ppCutsTemp[Entry]; (pCut != nil); pCut = *Amap_ManCutNextP(pCut)  {
            nCuts++;
            if nCuts < nMaxCuts {
                nWords += pCut.nFans + 1; }
        }
    }
    p.nBytesUsed += 4*nWords;
    // allocate memory
    pBuffer = (int *)Aig_MmFlexEntryFetch( p.pMemCuts, 4*nWords );
    pNext = (Amap_Cut_t *)pBuffer;
    // add the first cut
    pNext.iMat  = 0; 
    pNext.fInv  = 0;
    pNext.nFans = 1;
    pNext.Fans[0] = Abc_Var2Lit(pNode.Id, 0);
    pNext  = (Amap_Cut_t *)(pBuffer + 2);
    // add other cuts
    nCuts2 = 1;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vTemp)) && (((Entry) = Vec_IntEntry(p.vTemp, i)), 1); i++  {
        for  pCut = p.ppCutsTemp[Entry]; (pCut != nil); pCut = *Amap_ManCutNextP(pCut)  {
            nCuts2++;
            if nCuts2 < nMaxCuts {
                memcpy( pNext, pCut, sizeof(int) * (pCut.nFans + 1) );
                pNext = (Amap_Cut_t *)((int *)pNext + pCut.nFans + 1);
            }
        }
        p.ppCutsTemp[Entry] = nil;
    }
    assert(nCuts == nCuts2, "nCuts == nCuts2");
    assert((int *)pNext - pBuffer == nWords, "(int *)pNext - pBuffer == nWords");
    // restore the storage
    Vec_IntClear( p.vTemp );
    Aig_MmFlexRestart( p.pMemTemp );
    for  i = 0; i < 2*p.pLib.nNodes; i++  {
        if p.ppCutsTemp[i] != nil {
            fmt.Printf("Amap_ManCutSaveStored(): Error!\n"); }
    }
    pNode.pData = (Amap_Cut_t *)pBuffer;
    pNode.nCuts = Abc_MinInt( nCuts, nMaxCuts-1 );
    assert(nCuts < (1<<20), "nCuts < (1<<20)");
//    printf("%d ", nCuts );
    // verify cuts
    pCut = nil;
    // inlining of macro Amap_NodeForEachCut
    for  i = 0, pNext = (Amap_Cut_t *)pNode.pData; i < int32(pNode.nCuts);         i++, pNext = Amap_ManCutNext(pNext)  {
//        for ( i = 0, pNext = (Amap_Cut_t *)pNode->pData; i < (int)pNode->nCuts; 
//        i++, pNext = Amap_ManCutNext(pNext) )
    
        if i == nMaxCuts {
            break; }
        assert(pCut == nil || pCut.iMat <= pNext.iMat, "pCut == nil || pCut.iMat <= pNext.iMat");
        pCut = pNext;
    }
}

/**Function*************************************************************

  Synopsis    [Returns the number of possible new cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMergeCountCuts(p *Amap_Man_t, pNode *Amap_Obj_t) int32 {
    var pFanin0 *Amap_Obj_t  = Amap_ObjFanin0( p, pNode );
    var pFanin1 *Amap_Obj_t  = Amap_ObjFanin1( p, pNode );
    var pCut0, pCut1 *Amap_Cut_t;
    var Entry, c0, c1, iCompl0, iCompl1, iFan0, iFan1 int32;
    var Counter int32  = 1;
    // inlining of macro Amap_NodeForEachCut
    for  c0 = 0, pCut0 = (Amap_Cut_t *)pFanin0.pData; c0 < int32(pFanin0.nCuts);         c0++, pCut0 = Amap_ManCutNext(pCut0)  {
    // inlining of macro Amap_NodeForEachCut
    for  c1 = 0, pCut1 = (Amap_Cut_t *)pFanin1.pData; c1 < int32(pFanin1.nCuts);         c1++, pCut1 = Amap_ManCutNext(pCut1)  {
        iCompl0 = pCut0.fInv ^ Amap_ObjFaninC0(pNode);
        iCompl1 = pCut1.fInv ^ Amap_ObjFaninC1(pNode);
        if !pCut0.iMat {
            iFan0 = 0;
        } else {
            iFan0 = Abc_Var2Lit( pCut0.iMat, iCompl0 );
        }
        if !pCut1.iMat {
            iFan1 = 0;
        } else {
            iFan1 = Abc_Var2Lit( pCut1.iMat, iCompl1 );
        }
        Entry = Amap_LibFindNode( p.pLib, iFan0, iFan1, pNode.Type == AMAP_OBJ_XOR );
        Counter += ( Entry >=0 );
//        if ( Entry >=0 )
//            printf( "Full: %d + %d = %d\n", iFan0, iFan1, Entry );
    }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Print cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManPrintCuts(pNode *Amap_Obj_t) {
    var pCut *Amap_Cut_t;
    var c, i int32;
    fmt.Printf("NODE %5d : Type = ", pNode.Id);
    if pNode.Type == AMAP_OBJ_AND {
        fmt.Printf("AND");
    } else if pNode.Type == AMAP_OBJ_XOR {
        fmt.Printf("XOR");
    } else if pNode.Type == AMAP_OBJ_MUX {
        fmt.Printf("MUX"); }
    fmt.Printf("  Cuts = %d\n", pNode.nCuts);
    // inlining of macro Amap_NodeForEachCut
    for  c = 0, pCut = (Amap_Cut_t *)pNode.pData; c < int32(pNode.nCuts);         c++, pCut = Amap_ManCutNext(pCut)  {
        fmt.Printf("%3d :  Mat= %3d  Inv=%d  ", c, pCut.iMat, pCut.fInv);
        for  i = 0; i < int32(pCut.nFans); i++  {
            fmt.Printf("%d%c ", Abc_Lit2Var(pCut.Fans[i]), Abc_LitIsCompl(pCut.Fans[i])?'-':'+');
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Derives cuts for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMergeNodeChoice(p *Amap_Man_t, pNode *Amap_Obj_t) {
    var pTemp *Amap_Obj_t;
    var pCut *Amap_Cut_t;
    var c int32;
    // go through the nodes of the choice node
    for  pTemp = pNode; (pTemp != nil); pTemp = Amap_ObjChoice(p, pTemp)  {
        // inlining of macro Amap_NodeForEachCut
        for  c = 0, pCut = (Amap_Cut_t *)pTemp.pData; c < int32(pTemp.nCuts);         c++, pCut = Amap_ManCutNext(pCut)  {
            if !(pCut != nil) { break; }   // mikelee added; abort when pCut is NULL
            if pCut.iMat {
                Amap_ManCutStore( p, pCut, pNode.fPhase ^ pTemp.fPhase ); }
        }
        pTemp.pData = nil;
    }
    Amap_ManCutSaveStored( p, pNode );

//    Amap_ManPrintCuts( pNode );
}

/**Function*************************************************************

  Synopsis    [Derives cuts for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManFindCut(pNode *Amap_Obj_t, pFanin *Amap_Obj_t, fComplFanin int32, Val int32, vCuts *Vec_Ptr_t) int32 {
    var pCut *Amap_Cut_t;
    var c, iCompl, iFan int32;
    Vec_PtrClear( vCuts );
    // inlining of macro Amap_NodeForEachCut
    for  c = 0, pCut = (Amap_Cut_t *)pFanin.pData; c < int32(pFanin.nCuts);         c++, pCut = Amap_ManCutNext(pCut)  {
        iCompl = pCut.fInv ^ fComplFanin;
        if !pCut.iMat {
            iFan = 0;
        } else {
            iFan = Abc_Var2Lit( pCut.iMat, iCompl );
        }
        if iFan == Val {
            Vec_PtrPush( vCuts, pCut ); }
    }
    return Vec_PtrSize(vCuts) == 0;
}

/**Function*************************************************************

  Synopsis    [Derives cuts for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMergeNodeCutsMux(p *Amap_Man_t, pNode *Amap_Obj_t) {
    var vRules *Vec_Int_t  = p.pLib.vRules3;
    var pFanin0 *Amap_Obj_t  = Amap_ObjFanin0( p, pNode );
    var pFanin1 *Amap_Obj_t  = Amap_ObjFanin1( p, pNode );
    var pFanin2 *Amap_Obj_t  = Amap_ObjFanin2( p, pNode );
    var fComplFanin0 int32  = Amap_ObjFaninC0( pNode );
    var fComplFanin1 int32  = Amap_ObjFaninC1( pNode );
    var fComplFanin2 int32  = Amap_ObjFaninC2( pNode );
    var pCut0, pCut1, pCut2 *Amap_Cut_t;
    var x, c0, c1, c2 int32;
    assert(pNode.pData == nil, "pNode.pData == nil");
    assert(pNode.Type == AMAP_OBJ_MUX, "pNode.Type == AMAP_OBJ_MUX");
    assert(pNode.fRepr == 0, "pNode.fRepr == 0");
    // go through the rules
    for  x = 0; x < Vec_IntSize(vRules); x += 4  {
        if Amap_ManFindCut( pNode, pFanin0, fComplFanin0, Vec_IntEntry(vRules, x), p.vCuts0 ) != 0 {
            continue; }
        if Amap_ManFindCut( pNode, pFanin1, fComplFanin1, Vec_IntEntry(vRules, x+1), p.vCuts1 ) != 0 {
            continue; }
        if Amap_ManFindCut( pNode, pFanin2, fComplFanin2, Vec_IntEntry(vRules, x+2), p.vCuts2 ) != 0 {
            continue; }
        // inlining of macro Vec_PtrForEachEntry
        for  c0 = 0; (c0 < Vec_PtrSize(p.vCuts0)) && (((pCut0) = (Amap_Cut_t *)Vec_PtrEntry(p.vCuts0, c0)), 1); c0++  {
        // inlining of macro Vec_PtrForEachEntry
        for  c1 = 0; (c1 < Vec_PtrSize(p.vCuts1)) && (((pCut1) = (Amap_Cut_t *)Vec_PtrEntry(p.vCuts1, c1)), 1); c1++  {
        // inlining of macro Vec_PtrForEachEntry
        for  c2 = 0; (c2 < Vec_PtrSize(p.vCuts2)) && (((pCut2) = (Amap_Cut_t *)Vec_PtrEntry(p.vCuts2, c2)), 1); c2++  {
            var pNod *Amap_Nod_t  = Amap_LibNod( p.pLib, Vec_IntEntry(vRules, x+3) );
            if pNod.pSets == nil {
                continue; }
            // complement literals
            if pCut0.nFans == 1 && (pCut0.fInv ^ fComplFanin0) {
                pCut0.Fans[0] = Abc_LitNot(pCut0.Fans[0]); }
            if pCut1.nFans == 1 && (pCut1.fInv ^ fComplFanin1) {
                pCut1.Fans[0] = Abc_LitNot(pCut1.Fans[0]); }
            if pCut2.nFans == 1 && (pCut2.fInv ^ fComplFanin2) {
                pCut2.Fans[0] = Abc_LitNot(pCut2.Fans[0]); }
            // create new cut
            Amap_ManCutCreate3( p, pCut0, pCut1, pCut2, Vec_IntEntry(vRules, x+3) );
            // uncomplement literals
            if pCut0.nFans == 1 && (pCut0.fInv ^ fComplFanin0) {
                pCut0.Fans[0] = Abc_LitNot(pCut0.Fans[0]); }
            if pCut1.nFans == 1 && (pCut1.fInv ^ fComplFanin1) {
                pCut1.Fans[0] = Abc_LitNot(pCut1.Fans[0]); }
            if pCut2.nFans == 1 && (pCut2.fInv ^ fComplFanin2) {
                pCut2.Fans[0] = Abc_LitNot(pCut2.Fans[0]); }
        }
        }
        }
    }
    Amap_ManCutSaveStored( p, pNode );
    p.nCutsUsed += pNode.nCuts;
    p.nCutsTried3 += pFanin0.nCuts * pFanin1.nCuts * pFanin2.nCuts;

//    Amap_ManPrintCuts( pNode );
}

/**Function*************************************************************

  Synopsis    [Derives cuts for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMergeNodeCuts(p *Amap_Man_t, pNode *Amap_Obj_t) {
    var pFanin0 *Amap_Obj_t  = Amap_ObjFanin0( p, pNode );
    var pFanin1 *Amap_Obj_t  = Amap_ObjFanin1( p, pNode );
    var pCut0, pCut1 *Amap_Cut_t;
    var pRules **int32;
    var Entry int32;
    var i int32;
    var k int32;
    var c int32;
    var iCompl0 int32;
    var iCompl1 int32;
    var iFan0 int32;
    var iFan1 int32;
    assert(pNode.pData == nil, "pNode.pData == nil");
    if pNode.Type == AMAP_OBJ_MUX {
        Amap_ManMergeNodeCutsMux( p, pNode );
        return;
    }
    assert(pNode.Type != AMAP_OBJ_MUX, "pNode.Type != AMAP_OBJ_MUX");
    if pNode.Type == AMAP_OBJ_AND {
        pRules = p.pLib.pRules;
    } else {
        pRules = p.pLib.pRulesX;
    }
    // inlining of macro Amap_NodeForEachCut
    for  c = 0, pCut0 = (Amap_Cut_t *)pFanin0.pData; c < int32(pFanin0.nCuts);         c++, pCut0 = Amap_ManCutNext(pCut0)  {
        iCompl0 = pCut0.fInv ^ Amap_ObjFaninC0(pNode);
        if !pCut0.iMat {
            iFan0 = 0;
        } else {
            iFan0 = Abc_Var2Lit( pCut0.iMat, iCompl0 );
        }
        // complement literals
        if pCut0.nFans == 1 && (iCompl0 != 0) {
            pCut0.Fans[0] = Abc_LitNot(pCut0.Fans[0]); }
        // label resulting sets
        for  i = 0; (Entry = pRules[iFan0][i]); i++  {
            p.pMatsTemp[Entry & 0xffff] = (Entry >> 16);
        }
        // iterate through the cuts
        // inlining of macro Amap_NodeForEachCut
        for  k = 0, pCut1 = (Amap_Cut_t *)pFanin1.pData; k < int32(pFanin1.nCuts);         k++, pCut1 = Amap_ManCutNext(pCut1)  {
            iCompl1 = pCut1.fInv ^ Amap_ObjFaninC1(pNode);
            if !pCut1.iMat {
                iFan1 = 0;
            } else {
                iFan1 = Abc_Var2Lit( pCut1.iMat, iCompl1 );
            }
            if p.pMatsTemp[iFan1] == 0 {
                continue; }
            // complement literals
            if pCut1.nFans == 1 && (iCompl1 != 0) {
                pCut1.Fans[0] = Abc_LitNot(pCut1.Fans[0]); }
            // create new cut
            if iFan0 >= iFan1 {
                Amap_ManCutCreate( p, pCut0, pCut1, p.pMatsTemp[iFan1] );
            } else {
                Amap_ManCutCreate( p, pCut1, pCut0, p.pMatsTemp[iFan1] ); }
            // uncomplement literals
            if pCut1.nFans == 1 && (iCompl1 != 0) {
                pCut1.Fans[0] = Abc_LitNot(pCut1.Fans[0]); }
        }
        // uncomplement literals
        if pCut0.nFans == 1 && (iCompl0 != 0) {
            pCut0.Fans[0] = Abc_LitNot(pCut0.Fans[0]); }
        // label resulting sets
        for  i = 0; (Entry = pRules[iFan0][i]); i++  {
            p.pMatsTemp[Entry & 0xffff] = 0;
        }
    }
    Amap_ManCutSaveStored( p, pNode );
    p.nCutsUsed += pNode.nCuts;
    p.nCutsTried += pFanin0.nCuts * pFanin1.nCuts;
//    assert( (int)pNode->nCuts == Amap_ManMergeCountCuts(p, pNode) );
    if pNode.fRepr {
        Amap_ManMergeNodeChoice( p, pNode ); }

//    Amap_ManPrintCuts( pNode );
}

/**Function*************************************************************

  Synopsis    [Derives cuts for all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManMerge(p *Amap_Man_t) {
    var pObj *Amap_Obj_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    p.pCutsPi = Amap_ManSetupPis( p );
    // inlining of macro Amap_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Amap_ObjIsNode(pObj) {} else {
        Amap_ManMergeNodeCuts( p, pObj ); }
    }
    if p.pPars.fVerbose {
        fmt.Printf("AIG object is %d bytes.  ", int32(sizeof(Amap_Obj_t)));
        fmt.Printf("Internal AIG = %5.2f MB.  Cuts = %5.2f MB.  CutsMax = %d.\n", 1.0*Amap_ManObjNum(p)*sizeof(Amap_Obj_t)/(1<<20), 1.0*p.nBytesUsed/(1<<20), p.pPars.nCutsMax);
        fmt.Printf("Node =%6d. Try =%9d. Try3 =%10d. Used =%7d. R =%6.2f.  ", Amap_ManNodeNum(p), p.nCutsTried, p.nCutsTried3, p.nCutsUsed, 1.0*p.nCutsUsed/Amap_ManNodeNum(p));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

