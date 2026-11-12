package aig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [aigPart.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [AIG partitioning package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigPart.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "misc/tim/tim.h"
//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

       
type Part_Man_t struct {
    nChunkSize              int32;    // the size of one chunk of memory (~1 MB)
    nStepSize              int32;     // the step size in saving memory (~64 bytes)
    pFreeBuf            string;      // the pointer to free memory
    nFreeSize              int32;     // the size of remaining free memory
    vMemory       *Vec_Ptr_t;       // the memory allocated
    vFree       *Vec_Ptr_t;         // the vector of free pieces of memory
}

       
type Part_One_t struct {
    nRefs              int32;         // the number of references
    nOuts              int32;         // the number of outputs
    nOutsAlloc              int32;    // the array size
    pOuts [0]int32;      // the array of outputs
}

func Part_SizeType(nSize int32, nStepSize int32) int32 { return nSize / nStepSize + ((nSize % nStepSize) > 0); }
func Part_OneNext(pPart string) string { return *((char **)pPart);                             }
func Part_OneSetNext(pPart string, pNext string) { *((char **)pPart) = pNext;                            }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start the memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManStart(nChunkSize int32, nStepSize int32) *Part_Man_t {
    var p *Part_Man_t;
    // inlining of macro ABC_ALLOC
    p = ((Part_Man_t *) malloc(sizeof(Part_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Part_Man_t) );
    p.nChunkSize = nChunkSize;
    p.nStepSize  = nStepSize;
    p.vMemory    = Vec_PtrAlloc( 1000 );
    p.vFree      = Vec_PtrAlloc( 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManStop(p *Part_Man_t) {
    var pMemory any;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vMemory)) && (((pMemory) = (void *)Vec_PtrEntry(p.vMemory, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((pMemory) ? (free((char *) (pMemory)), (pMemory) = 0) : 0);
    }
    Vec_PtrFree( p.vMemory );
    Vec_PtrFree( p.vFree );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Fetches the memory entry of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManFetch(p *Part_Man_t, nSize int32) string {
    var Type, nSizeReal int32;
    var pMemory string;
    assert(nSize > 0, "nSize > 0");
    Type = Part_SizeType( nSize, p.nStepSize );
    Vec_PtrFillExtra( p.vFree, Type + 1, nil );
    if pMemory = (char *)Vec_PtrEntry( p.vFree, Type ) {
        Vec_PtrWriteEntry( p.vFree, Type, Part_OneNext(pMemory) );
        return pMemory;
    }
    nSizeReal = p.nStepSize * Type;
    if p.nFreeSize < nSizeReal {
        // inlining of macro ABC_ALLOC
        p.pFreeBuf = ((char *) malloc(sizeof(char) * (size_t)(p.nChunkSize)));
        p.nFreeSize = p.nChunkSize;
        Vec_PtrPush( p.vMemory, p.pFreeBuf );
    }
    assert(p.nFreeSize >= nSizeReal, "p.nFreeSize >= nSizeReal");
    pMemory = p.pFreeBuf;
    p.pFreeBuf  += nSizeReal;
    p.nFreeSize -= nSizeReal;
    return pMemory;
}

/**Function*************************************************************

  Synopsis    [Recycles the memory entry of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManRecycle(p *Part_Man_t, pMemory string, nSize int32) {
    var Type int32;
    Type = Part_SizeType( nSize, p.nStepSize );
    Vec_PtrFillExtra( p.vFree, Type + 1, nil );
    Part_OneSetNext( pMemory, (char *)Vec_PtrEntry(p.vFree, Type) );
    Vec_PtrWriteEntry( p.vFree, Type, pMemory );
}

/**Function*************************************************************

  Synopsis    [Fetches the memory entry of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManFetchEntry(p *Part_Man_t, nWords int32, nRefs int32) *Part_One_t {
    var pPart *Part_One_t;
    pPart = (Part_One_t *)Part_ManFetch( p, sizeof(Part_One_t) + sizeof(int) * nWords );
    pPart.nRefs = nRefs;
    pPart.nOuts = 0;
    pPart.nOutsAlloc = nWords;
    return pPart;
}

/**Function*************************************************************

  Synopsis    [Recycles the memory entry of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManRecycleEntry(p *Part_Man_t, pEntry *Part_One_t) {
    assert(pEntry.nOuts <= pEntry.nOutsAlloc, "pEntry.nOuts <= pEntry.nOutsAlloc");
    assert(pEntry.nOuts >= pEntry.nOutsAlloc/2, "pEntry.nOuts >= pEntry.nOutsAlloc/2");
    Part_ManRecycle( p, (char *)pEntry, sizeof(Part_One_t) + sizeof(int) * pEntry.nOutsAlloc );
}

/**Function*************************************************************

  Synopsis    [Merges two entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManMergeEntry(pMan *Part_Man_t, p1 *Part_One_t, p2 *Part_One_t, nRefs int32) *Part_One_t {
    var p *Part_One_t  = Part_ManFetchEntry( pMan, p1.nOuts + p2.nOuts, nRefs );
    var pBeg1 *int32  = p1.pOuts;
    var pBeg2 *int32  = p2.pOuts;
    var pBeg *int32   = p.pOuts;
    var pEnd1 *int32  = p1.pOuts + p1.nOuts;
    var pEnd2 *int32  = p2.pOuts + p2.nOuts;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            *pBeg++ = *pBeg1++, pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            *pBeg++ = *pBeg1++;
        } else { 
            *pBeg++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pBeg++ = *pBeg2++;
    }
    p.nOuts = pBeg - p.pOuts;
    assert(p.nOuts <= p.nOutsAlloc, "p.nOuts <= p.nOutsAlloc");
    assert(p.nOuts >= p1.nOuts, "p.nOuts >= p1.nOuts");
    assert(p.nOuts >= p2.nOuts, "p.nOuts >= p2.nOuts");
    return p;
}

/**Function*************************************************************

  Synopsis    [Tranfers the entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Part_ManTransferEntry(p *Part_One_t) *Vec_Int_t {
    var vSupp *Vec_Int_t;
    var i int32;
    vSupp = Vec_IntAlloc( p.nOuts );
    for  i = 0; i < p.nOuts; i++  {
        Vec_IntPush( vSupp, p.pOuts[i] );
    }
    return vSupp;
}

/**Function*************************************************************

  Synopsis    [Computes supports of the POs in the multi-output AIG.]

  Description [Returns the array of integer arrays containing indices
  of the primary inputs for each primary output.]
               
  SideEffects [Adds the integer PO number at end of each array.]

  SeeAlso     []

***********************************************************************/
func Aig_ManSupports(pMan *Aig_Man_t) *Vec_Ptr_t {
    var vSupports *Vec_Ptr_t;
    var vSupp *Vec_Int_t;
    var p *Part_Man_t;
    var pPart0, pPart1 *Part_One_t;
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // set the number of PIs/POs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        pObj.pNext = (Aig_Obj_t *)int64(i);
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        pObj.pNext = (Aig_Obj_t *)int64(i);
    }
    // start the support computation manager
    p = Part_ManStart( 1 << 20, 1 << 6 );
    // consider objects in the topological order
    vSupports = Vec_PtrAlloc( Aig_ManCoNum(pMan) );
    Aig_ManCleanData(pMan);
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsNode(pObj) {
            pPart0 = (Part_One_t *)Aig_ObjFanin0(pObj).pData;
            pPart1 = (Part_One_t *)Aig_ObjFanin1(pObj).pData;
            pObj.pData = Part_ManMergeEntry( p, pPart0, pPart1, pObj.nRefs );
            assert(pPart0.nRefs > 0, "pPart0.nRefs > 0");
            if --pPart0.nRefs == 0 {
                Part_ManRecycleEntry( p, pPart0 ); }
            assert(pPart1.nRefs > 0, "pPart1.nRefs > 0");
            if --pPart1.nRefs == 0 {
                Part_ManRecycleEntry( p, pPart1 ); }
            if ((Part_One_t *)pObj.pData).nOuts <= 16 {
                Counter++; }
            continue;
        }
        if Aig_ObjIsCo(pObj) {
            pPart0 = (Part_One_t *)Aig_ObjFanin0(pObj).pData;
            vSupp = Part_ManTransferEntry(pPart0);
            Vec_IntPush( vSupp, int32(int64(pObj.pNext)) );
            Vec_PtrPush( vSupports, vSupp );
            assert(pPart0.nRefs > 0, "pPart0.nRefs > 0");
            if --pPart0.nRefs == 0 {
                Part_ManRecycleEntry( p, pPart0 ); }
            continue;
        }
        if Aig_ObjIsCi(pObj) {
            if pObj.nRefs {
                pPart0 = Part_ManFetchEntry( p, 1, pObj.nRefs );
                pPart0.pOuts[ pPart0.nOuts++ ] = int32(int64(pObj.pNext));
                pObj.pData = pPart0;
            }
            continue;
        }
        if Aig_ObjIsConst1(pObj) {
            if pObj.nRefs {
                pObj.pData = Part_ManFetchEntry( p, 0, pObj.nRefs ); }
            continue;
        }
        assert(false, "0");
    }
    }
//printf( "Memory usage = %d MB.\n", Vec_PtrSize(p->vMemory) * p->nChunkSize / (1<<20) );
    Part_ManStop( p );
    // sort supports by size
    Vec_VecSort( (Vec_Vec_t *)vSupports, 1 );
    // clear the number of PIs/POs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
        pObj.pNext = nil;
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCos, i)), 1); i++  {
        pObj.pNext = nil;
    }
/*
    Aig_ManForEachCo( pMan, pObj, i )
        printf( "%d ", Vec_IntSize( Vec_VecEntryInt(vSupports, i) ) );
    printf( "\n" );
*/
//    printf( "%d \n", Counter );
    return vSupports;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSupportsTest(pMan *Aig_Man_t) {
    var vSupps *Vec_Ptr_t;
    vSupps = Aig_ManSupports( pMan );
    Vec_VecFree( (Vec_Vec_t *)vSupps );
}

/**Function*************************************************************

  Synopsis    [Computes the set of outputs for each input.]

  Description [Returns the array of integer arrays containing indices
  of the primary outputsf for each primary input.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSupportsInverse(p *Aig_Man_t) *Vec_Ptr_t {
    var vSupps, vSuppsInv *Vec_Ptr_t;
    var vSupp *Vec_Int_t;
    var i, k, iIn, iOut int32;
    // get structural supports for each output
    vSupps = Aig_ManSupports( p );
    // start the inverse supports
    vSuppsInv = Vec_PtrAlloc( Aig_ManCiNum(p) );
    for  i = 0; i < Aig_ManCiNum(p); i++  {
        Vec_PtrPush( vSuppsInv, Vec_IntAlloc(8) );
    }
    // transforms the supports into the inverse supports
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupps)) && (((vSupp) = (Vec_Int_t *)Vec_PtrEntry(vSupps, i)), 1); i++  {
        iOut = Vec_IntPop( vSupp );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSupp)) && (((iIn) = Vec_IntEntry(vSupp, k)), 1); k++  {
            Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vSuppsInv, iIn), iOut );
        }
    }
    Vec_VecFree( (Vec_Vec_t *)vSupps );
    return vSuppsInv;
}

/**Function*************************************************************

  Synopsis    [Returns the register dependency matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSupportsRegisters(p *Aig_Man_t) *Vec_Ptr_t {
    var vSupports, vMatrix *Vec_Ptr_t;
    var vSupp *Vec_Int_t;
    var iOut, iIn, k, m, i int32;
    // get structural supports for each output
    vSupports = Aig_ManSupports( p );
    // transforms the supports into the latch dependency matrix
    vMatrix = Vec_PtrStart( Aig_ManRegNum(p) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupports)) && (((vSupp) = (Vec_Int_t *)Vec_PtrEntry(vSupports, i)), 1); i++  {
        // skip true POs
        iOut = Vec_IntPop( vSupp );
        iOut -= Aig_ManCoNum(p) - Aig_ManRegNum(p);
        if iOut < 0 {
            Vec_IntFree( vSupp );
            continue;
        }
        // remove PIs
        m = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSupp)) && (((iIn) = Vec_IntEntry(vSupp, k)), 1); k++  {
            iIn -= Aig_ManCiNum(p) - Aig_ManRegNum(p);
            if iIn < 0 {
                continue; }
            assert(iIn < Aig_ManRegNum(p), "iIn < Aig_ManRegNum(p)");
            Vec_IntWriteEntry( vSupp, m++, iIn );
        }
        Vec_IntShrink( vSupp, m );
        // store support in the matrix
        assert(iOut < Aig_ManRegNum(p), "iOut < Aig_ManRegNum(p)");
        Vec_PtrWriteEntry( vMatrix, iOut, vSupp );
    }
    Vec_PtrFree( vSupports );
    // check that all supports are used
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMatrix)) && (((vSupp) = (Vec_Int_t *)Vec_PtrEntry(vMatrix, i)), 1); i++  {
        assert(vSupp != nil, "vSupp != nil");
    }
    return vMatrix;
}

/**Function*************************************************************

  Synopsis    [Start char-bases support representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSuppCharStart(vOne *Vec_Int_t, nPis int32) *uint32 {
    var pBuffer *uint32;
    var i, Entry int32;
    var nWords int32  = Abc_BitWordNum(nPis);
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nWords)));
    memset( pBuffer, 0, sizeof(unsigned) * nWords );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOne)) && (((Entry) = Vec_IntEntry(vOne, i)), 1); i++  {
        assert(Entry < nPis, "Entry < nPis");
        Abc_InfoSetBit( pBuffer, Entry );
    }
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    [Add to char-bases support representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSuppCharAdd(pBuffer *uint32, vOne *Vec_Int_t, nPis int32) {
    var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOne)) && (((Entry) = Vec_IntEntry(vOne, i)), 1); i++  {
        assert(Entry < nPis, "Entry < nPis");
        Abc_InfoSetBit( pBuffer, Entry );
    }
}

/**Function*************************************************************

  Synopsis    [Find the common variables using char-bases support representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSuppCharCommon(pBuffer *uint32, vOne *Vec_Int_t) int32 {
    var i, Entry, nCommon int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOne)) && (((Entry) = Vec_IntEntry(vOne, i)), 1); i++  {
        nCommon += Abc_InfoHasBit(pBuffer, Entry);
    }
    return nCommon;
}

/**Function*************************************************************

  Synopsis    [Find the best partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionSmartFindPart(vPartSuppsAll *Vec_Ptr_t, vPartsAll *Vec_Ptr_t, vPartSuppsBit *Vec_Ptr_t, nSuppSizeLimit int32, vOne *Vec_Int_t) int32 {
    var vPartSupp *Vec_Int_t;//, * vPart;
    var Attract, Repulse, Value, ValueBest int32;
    var i, nCommon, iBest int32;
    iBest = -1;
    ValueBest = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vPartSupp) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
//        vPart = Vec_PtrEntry( vPartsAll, i );
//        if ( nSuppSizeLimit > 0 && Vec_IntSize(vPart) >= nSuppSizeLimit )
//            continue;
//        nCommon = Vec_IntTwoCountCommon( vPartSupp, vOne );
        nCommon = Aig_ManSuppCharCommon( (unsigned *)Vec_PtrEntry(vPartSuppsBit, i), vOne );
        if nCommon == 0 {
            continue; }
        if nCommon == Vec_IntSize(vOne) {
            return i; }
        // skip partitions whose size exceeds the limit
        if nSuppSizeLimit > 0 && Vec_IntSize(vPartSupp) >= 2 * nSuppSizeLimit {
            continue; }
        Attract = 1000 * nCommon / Vec_IntSize(vOne);
        if Vec_IntSize(vPartSupp) < 100 {
            Repulse = 1;
        } else {
            Repulse = 1+Abc_Base2Log(Vec_IntSize(vPartSupp)-100); }
        Value = Attract/Repulse;
        if ValueBest < Value {
            ValueBest = Value;
            iBest = i;
        }
    }
    if ValueBest < 75 {
        return -1; }
    return iBest;
}

/**Function*************************************************************

  Synopsis    [Perform the smart partitioning.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionPrint(p *Aig_Man_t, vPartsAll *Vec_Ptr_t, vPartSuppsAll *Vec_Ptr_t) {
    var vOne *Vec_Int_t;
    var i, nOutputs, Counter int32;

    Counter = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
        nOutputs = Vec_IntSize((Vec_Int_t *)Vec_PtrEntry(vPartsAll, i));
        fmt.Printf("%d=(%d,%d) ", i, Vec_IntSize(vOne), nOutputs);
        Counter += nOutputs;
        if i == Vec_PtrSize(vPartsAll) - 1 {
            break; }
    }
    assert(Counter == Aig_ManCoNum(p), "Counter == Aig_ManCoNum(p)");
//    printf( "\nTotal = %d. Outputs = %d.\n", Counter, Aig_ManCoNum(p) );
}

/**Function*************************************************************

  Synopsis    [Perform the smart partitioning.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionCompact(vPartsAll *Vec_Ptr_t, vPartSuppsAll *Vec_Ptr_t, nSuppSizeLimit int32) {
    var vOne, vPart, vPartSupp, vTemp *Vec_Int_t;
    var i, iPart int32;

    if nSuppSizeLimit == 0 {
        nSuppSizeLimit = 200; }

    // pack smaller partitions into larger blocks
    iPart = 0;
    vPart = vPartSupp = nil;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
        if Vec_IntSize(vOne) < nSuppSizeLimit {
            if vPartSupp == nil {
                assert(vPart == nil, "vPart == nil");
                vPartSupp = Vec_IntDup(vOne);
                vPart = (Vec_Int_t *)Vec_PtrEntry(vPartsAll, i);
            
            } else {
                vPartSupp = Vec_IntTwoMerge( vTemp = vPartSupp, vOne );
                Vec_IntFree( vTemp );
                vPart = Vec_IntTwoMerge( vTemp = vPart, (Vec_Int_t *)Vec_PtrEntry(vPartsAll, i) );
                Vec_IntFree( vTemp );
                Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vPartsAll, i) );
            }
            if Vec_IntSize(vPartSupp) < nSuppSizeLimit {
                continue; }
        
        } else {
            vPart = (Vec_Int_t *)Vec_PtrEntry(vPartsAll, i); }
        // add the partition 
        Vec_PtrWriteEntry( vPartsAll, iPart, vPart );  
        vPart = nil;
        if vPartSupp != nil {
            Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, iPart) );
            Vec_PtrWriteEntry( vPartSuppsAll, iPart, vPartSupp );  
            vPartSupp = nil;
        }
        iPart++;
    }
    // add the last one
    if vPart != nil {
        Vec_PtrWriteEntry( vPartsAll, iPart, vPart );  
        vPart = nil;

        assert(vPartSupp != nil, "vPartSupp != nil");
        Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, iPart) );
        Vec_PtrWriteEntry( vPartSuppsAll, iPart, vPartSupp );  
        vPartSupp = nil;
        iPart++;
    }
    Vec_PtrShrink( vPartsAll, iPart );
    Vec_PtrShrink( vPartsAll, iPart );
}

/**Function*************************************************************

  Synopsis    [Perform the smart partitioning.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionSmart(p *Aig_Man_t, nSuppSizeLimit int32, fVerbose int32, pvPartSupps **Vec_Ptr_t) *Vec_Ptr_t {
    var vPartSuppsBit *Vec_Ptr_t;
    var vSupports, vPartsAll, vPartsAll2, vPartSuppsAll *Vec_Ptr_t;//, * vPartPtr;
    var vOne, vPart, vPartSupp, vTemp *Vec_Int_t;
    var i, iPart, iOut int32;
    var clk abctime;

    // compute the supports for all outputs
clk = Abc_Clock();
    vSupports = Aig_ManSupports( p );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Supps"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
    // start char-based support representation
    vPartSuppsBit = Vec_PtrAlloc( 1000 );

    // create partitions
clk = Abc_Clock();
    vPartsAll = Vec_PtrAlloc( 256 );
    vPartSuppsAll = Vec_PtrAlloc( 256 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupports)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vSupports, i)), 1); i++  {
        // get the output number
        iOut = Vec_IntPop(vOne);
        // find closely matching part
        iPart = Aig_ManPartitionSmartFindPart( vPartSuppsAll, vPartsAll, vPartSuppsBit, nSuppSizeLimit, vOne );
        if iPart == -1 {
            // create new partition
            vPart = Vec_IntAlloc( 32 );
            Vec_IntPush( vPart, iOut );
            // create new partition support
            vPartSupp = Vec_IntDup( vOne );
            // add this partition and its support
            Vec_PtrPush( vPartsAll, vPart );
            Vec_PtrPush( vPartSuppsAll, vPartSupp );

            Vec_PtrPush( vPartSuppsBit, Aig_ManSuppCharStart(vOne, Aig_ManCiNum(p)) );
        
        } else {
            // add output to this partition
            vPart = (Vec_Int_t *)Vec_PtrEntry( vPartsAll, iPart );
            Vec_IntPush( vPart, iOut );
            // merge supports
            vPartSupp = (Vec_Int_t *)Vec_PtrEntry( vPartSuppsAll, iPart );
            vPartSupp = Vec_IntTwoMerge( vTemp = vPartSupp, vOne );
            Vec_IntFree( vTemp );
            // reinsert new support
            Vec_PtrWriteEntry( vPartSuppsAll, iPart, vPartSupp );

            Aig_ManSuppCharAdd( (unsigned *)Vec_PtrEntry(vPartSuppsBit, iPart), vOne, Aig_ManCiNum(p) );
        }
    }

    // stop char-based support representation
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsBit)) && (((vTemp) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsBit, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((vTemp) ? (free((char *) (vTemp)), (vTemp) = 0) : 0);
    }
    Vec_PtrFree( vPartSuppsBit );

//printf( "\n" );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Parts"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

clk = Abc_Clock();
    // reorder partitions in the decreasing order of support sizes
    // remember partition number in each partition support
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
        Vec_IntPush( vOne, i );
    }
    // sort the supports in the decreasing order
    Vec_VecSort( (Vec_Vec_t *)vPartSuppsAll, 1 );
    // reproduce partitions
    vPartsAll2 = Vec_PtrAlloc( 256 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
        Vec_PtrPush( vPartsAll2, Vec_PtrEntry(vPartsAll, Vec_IntPop(vOne)) );
    }
    Vec_PtrFree( vPartsAll );
    vPartsAll = vPartsAll2;

    // compact small partitions
//    Aig_ManPartitionPrint( p, vPartsAll, vPartSuppsAll );
    Aig_ManPartitionCompact( vPartsAll, vPartSuppsAll, nSuppSizeLimit );
    if fVerbose != 0 {
//    Aig_ManPartitionPrint( p, vPartsAll, vPartSuppsAll );
    fmt.Printf("Created %d partitions.\n", Vec_PtrSize(vPartsAll)); }

if fVerbose != 0 {
//ABC_PRT( "Comps", Abc_Clock() - clk );
}

    // cleanup
    Vec_VecFree( (Vec_Vec_t *)vSupports );
    if pvPartSupps == nil {
        Vec_VecFree( (Vec_Vec_t *)vPartSuppsAll );
    } else {
        *pvPartSupps = vPartSuppsAll; }
/*
    // converts from intergers to nodes
    Vec_PtrForEachEntry( Vec_Int_t *, vPartsAll, vPart, iPart )
    {
        vPartPtr = Vec_PtrAlloc( Vec_IntSize(vPart) );
        Vec_IntForEachEntry( vPart, iOut, i )
            Vec_PtrPush( vPartPtr, Aig_ManCo(p, iOut) );
        Vec_IntFree( vPart );
        Vec_PtrWriteEntry( vPartsAll, iPart, vPartPtr );
    }
*/
    return vPartsAll;
}

/**Function*************************************************************

  Synopsis    [Perform the smart partitioning.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionSmartRegisters(pAig *Aig_Man_t, nSuppSizeLimit int32, fVerbose int32) *Vec_Ptr_t {
    var vPartSuppsBit *Vec_Ptr_t;
    var vSupports, vPartsAll, vPartsAll2, vPartSuppsAll *Vec_Ptr_t;
    var vOne, vPart, vPartSupp, vTemp *Vec_Int_t;
    var i, iPart, iOut int32;
    var clk abctime;

    // add output number to each
clk = Abc_Clock();
    vSupports = Aig_ManSupportsRegisters( pAig );
    assert(Vec_PtrSize(vSupports) == Aig_ManRegNum(pAig), "Vec_PtrSize(vSupports) == Aig_ManRegNum(pAig)");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupports)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vSupports, i)), 1); i++  {
        Vec_IntPush( vOne, i );
    }
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Supps"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

    // start char-based support representation
    vPartSuppsBit = Vec_PtrAlloc( 1000 );

    // create partitions
clk = Abc_Clock();
    vPartsAll = Vec_PtrAlloc( 256 );
    vPartSuppsAll = Vec_PtrAlloc( 256 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupports)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vSupports, i)), 1); i++  {
        // get the output number
        iOut = Vec_IntPop(vOne);
        // find closely matching part
        iPart = Aig_ManPartitionSmartFindPart( vPartSuppsAll, vPartsAll, vPartSuppsBit, nSuppSizeLimit, vOne );
        if iPart == -1 {
            // create new partition
            vPart = Vec_IntAlloc( 32 );
            Vec_IntPush( vPart, iOut );
            // create new partition support
            vPartSupp = Vec_IntDup( vOne );
            // add this partition and its support
            Vec_PtrPush( vPartsAll, vPart );
            Vec_PtrPush( vPartSuppsAll, vPartSupp );

            Vec_PtrPush( vPartSuppsBit, Aig_ManSuppCharStart(vOne, Vec_PtrSize(vSupports)) );
        
        } else {
            // add output to this partition
            vPart = (Vec_Int_t *)Vec_PtrEntry( vPartsAll, iPart );
            Vec_IntPush( vPart, iOut );
            // merge supports
            vPartSupp = (Vec_Int_t *)Vec_PtrEntry( vPartSuppsAll, iPart );
            vPartSupp = Vec_IntTwoMerge( vTemp = vPartSupp, vOne );
            Vec_IntFree( vTemp );
            // reinsert new support
            Vec_PtrWriteEntry( vPartSuppsAll, iPart, vPartSupp );

            Aig_ManSuppCharAdd( (unsigned *)Vec_PtrEntry(vPartSuppsBit, iPart), vOne, Vec_PtrSize(vSupports) );
        }
    }

    // stop char-based support representation
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsBit)) && (((vTemp) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsBit, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((vTemp) ? (free((char *) (vTemp)), (vTemp) = 0) : 0);
    }
    Vec_PtrFree( vPartSuppsBit );

//printf( "\n" );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Parts"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

clk = Abc_Clock();
    // reorder partitions in the decreasing order of support sizes
    // remember partition number in each partition support
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
        Vec_IntPush( vOne, i );
    }
    // sort the supports in the decreasing order
    Vec_VecSort( (Vec_Vec_t *)vPartSuppsAll, 1 );
    // reproduce partitions
    vPartsAll2 = Vec_PtrAlloc( 256 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
        Vec_PtrPush( vPartsAll2, Vec_PtrEntry(vPartsAll, Vec_IntPop(vOne)) );
    }
    Vec_PtrFree( vPartsAll );
    vPartsAll = vPartsAll2;

    // compact small partitions
//    Aig_ManPartitionPrint( p, vPartsAll, vPartSuppsAll );
    Aig_ManPartitionCompact( vPartsAll, vPartSuppsAll, nSuppSizeLimit );
    if fVerbose != 0 {
//    Aig_ManPartitionPrint( p, vPartsAll, vPartSuppsAll );
    fmt.Printf("Created %d partitions.\n", Vec_PtrSize(vPartsAll)); }

if fVerbose != 0 {
//ABC_PRT( "Comps", Abc_Clock() - clk );
}

    // cleanup
    Vec_VecFree( (Vec_Vec_t *)vSupports );
//    if ( pvPartSupps == NULL )
        Vec_VecFree( (Vec_Vec_t *)vPartSuppsAll );
//    else
//        *pvPartSupps = vPartSuppsAll;

/*
    // converts from intergers to nodes
    Vec_PtrForEachEntry( Vec_Int_t *, vPartsAll, vPart, iPart )
    {
        vPartPtr = Vec_PtrAlloc( Vec_IntSize(vPart) );
        Vec_IntForEachEntry( vPart, iOut, i )
            Vec_PtrPush( vPartPtr, Aig_ManCo(p, iOut) );
        Vec_IntFree( vPart );
        Vec_PtrWriteEntry( vPartsAll, iPart, vPartPtr );
    }
*/
    return vPartsAll;
}

/**Function*************************************************************

  Synopsis    [Perform the naive partitioning.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionNaive(p *Aig_Man_t, nPartSize int32) *Vec_Ptr_t {
    var vParts *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var nParts, i int32;
    nParts = (Aig_ManCoNum(p) / nPartSize) + ((Aig_ManCoNum(p) % nPartSize) > 0);
    vParts = (Vec_Ptr_t *)Vec_VecStart( nParts );
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vParts, i / nPartSize), i );
    }
    return vParts;
}



/**Function*************************************************************

  Synopsis    [Adds internal nodes in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupPart_rec(pNew *Aig_Man_t, pOld *Aig_Man_t, pObj *Aig_Obj_t, vSuppMap *Vec_Int_t) *Aig_Obj_t {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsTravIdCurrent(pOld, pObj) {
        return (Aig_Obj_t *)pObj.pData; }
    Aig_ObjSetTravIdCurrent(pOld, pObj);
    if Aig_ObjIsCi(pObj) {
        assert(Vec_IntSize(vSuppMap) == Aig_ManCiNum(pNew), "Vec_IntSize(vSuppMap) == Aig_ManCiNum(pNew)");
        Vec_IntPush( vSuppMap, int32(int64(pObj.pNext)) );
        return (Aig_Obj_t *)(pObj.pData = Aig_ObjCreateCi(pNew));
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Aig_ManDupPart_rec( pNew, pOld, Aig_ObjFanin0(pObj), vSuppMap );
    Aig_ManDupPart_rec( pNew, pOld, Aig_ObjFanin1(pObj), vSuppMap );
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ));
}

/**Function*************************************************************

  Synopsis    [Adds internal nodes in the topological order.]

  Description [Returns the array of new outputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupPart(pNew *Aig_Man_t, pOld *Aig_Man_t, vPart *Vec_Int_t, vSuppMap *Vec_Int_t, fInverse int32) *Vec_Ptr_t {
    var vOutsTotal *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var Entry, i int32;
    // create the PIs
    Aig_ManIncrementTravId( pOld );
    Aig_ManConst1(pOld).pData = Aig_ManConst1(pNew);
    Aig_ObjSetTravIdCurrent( pOld, Aig_ManConst1(pOld) );
    if !(fInverse != 0) {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vSuppMap)) && (((Entry) = Vec_IntEntry(vSuppMap, i)), 1); i++  {
            pObj = Aig_ManCi( pOld, Entry );
            pObj.pData = Aig_ManCi( pNew, i );
            Aig_ObjSetTravIdCurrent( pOld, pObj );
        }
    
    } else {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vSuppMap)) && (((Entry) = Vec_IntEntry(vSuppMap, i)), 1); i++  {
            pObj = Aig_ManCi( pOld, i );
            pObj.pData = Aig_ManCi( pNew, Entry );
            Aig_ObjSetTravIdCurrent( pOld, pObj );
        }
        vSuppMap = nil; // should not be useful
    }
    // create the internal nodes
    vOutsTotal = Vec_PtrAlloc( Vec_IntSize(vPart) );
    if !(fInverse != 0) {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vPart)) && (((Entry) = Vec_IntEntry(vPart, i)), 1); i++  {
            pObj = Aig_ManCo( pOld, Entry );
            Aig_ManDupPart_rec( pNew, pOld, Aig_ObjFanin0(pObj), vSuppMap );
            Vec_PtrPush( vOutsTotal, Aig_ObjChild0Copy(pObj) );
        }
    
    } else {
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pOld.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pOld.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            if Aig_ObjIsCo(pObj) {
                Aig_ManDupPart_rec( pNew, pOld, Aig_ObjFanin0(pObj), vSuppMap );
                Vec_PtrPush( vOutsTotal, Aig_ObjChild0Copy(pObj) );
            
            } else if Aig_ObjIsNode(pObj) && pObj.nRefs == 0 {
                Aig_ManDupPart_rec( pNew, pOld, pObj, vSuppMap ); }
            
        }
        }
    }
    return vOutsTotal; 
}

/**Function*************************************************************

  Synopsis    [Adds internal nodes in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupPartAll_rec(pNew *Aig_Man_t, pOld *Aig_Man_t, pObj *Aig_Obj_t) {
    var pObjNew *Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsTravIdCurrent(pOld, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pOld, pObj);
    if Aig_ObjIsCi(pObj) {
        pObjNew = Aig_ObjCreateCi(pNew);
    } else if Aig_ObjIsCo(pObj) {
        Aig_ManDupPartAll_rec( pNew, pOld, Aig_ObjFanin0(pObj) );
        pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    
    } else {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        Aig_ManDupPartAll_rec( pNew, pOld, Aig_ObjFanin0(pObj) );
        Aig_ManDupPartAll_rec( pNew, pOld, Aig_ObjFanin1(pObj) );
        pObjNew = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }
    pObj.pData = pObjNew;
    pObjNew.pData = pObj;
}

/**Function*************************************************************

  Synopsis    [Adds internal nodes in the topological order.]

  Description [Returns the array of new outputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupPartAll(pOld *Aig_Man_t, vPart *Vec_Int_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var i, Entry int32;
    Aig_ManIncrementTravId( pOld );
    pNew = Aig_ManStart( 5000 );
    // map constant nodes
    pObj = Aig_ManConst1(pOld);
    pObjNew = Aig_ManConst1(pNew);
    pObj.pData = pObjNew;
    pObjNew.pData = pObj;
    Aig_ObjSetTravIdCurrent(pOld, pObj);
    // map all other nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((Entry) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Aig_ManCo( pOld, Entry );
        Aig_ManDupPartAll_rec( pNew, pOld, pObj );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSupportNodes_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vSupport *Vec_Int_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsCi(pObj) {
        Vec_IntPush( vSupport, Aig_ObjCioId(pObj) );
        return;
    }
    Aig_ManSupportNodes_rec( p, Aig_ObjFanin0(pObj), vSupport );
    Aig_ManSupportNodes_rec( p, Aig_ObjFanin1(pObj), vSupport );
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes and PIs in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSupportNodes(p *Aig_Man_t, vParts *Vec_Ptr_t) *Vec_Ptr_t {
    var vPartSupps *Vec_Ptr_t;
    var vPart, vSupport *Vec_Int_t;
    var i, k, iOut int32;
    Aig_ManSetCioIds( p );
    vPartSupps = Vec_PtrAlloc( Vec_PtrSize(vParts) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        vSupport = Vec_IntAlloc( 100 );
        Aig_ManIncrementTravId( p );
        Aig_ObjSetTravIdCurrent( p, Aig_ManConst1(p) );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, k)), 1); k++  {
            Aig_ManSupportNodes_rec( p, Aig_ObjFanin0(Aig_ManCo(p, iOut)), vSupport );
        }
//        Vec_IntSort( vSupport, 0 );
        Vec_PtrPush( vPartSupps, vSupport );
    }
    Aig_ManCleanCioIds( p );
    return vPartSupps;
}

/**Function*************************************************************

  Synopsis    [Create partitioned miter of the two AIGs.]

  Description [Assumes that each output in the second AIG cannot have 
  more supp vars than the same output in the first AIG.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMiterPartitioned(p1 *Aig_Man_t, p2 *Aig_Man_t, nPartSize int32, fSmart int32) *Vec_Ptr_t {
    var pNew *Aig_Man_t;
    var pMiter *Aig_Obj_t;
    var vMiters, vNodes1, vNodes2 *Vec_Ptr_t;
    var vParts, vPartSupps *Vec_Ptr_t;
    var vPart, vPartSupp *Vec_Int_t;
    var i, k int32;
    // partition the first manager
    if fSmart != 0 {
        vParts = Aig_ManPartitionSmart( p1, nPartSize, 0, &vPartSupps );
    } else {
        vParts = Aig_ManPartitionNaive( p1, nPartSize );
        vPartSupps = Aig_ManSupportNodes( p1, vParts );
    }
    // derive miters
    vMiters = Vec_PtrAlloc( Vec_PtrSize(vParts) );
    for  i = 0; i < Vec_PtrSize(vParts); i++  {
        // get partition and its support
        vPart     = (Vec_Int_t *)Vec_PtrEntry( vParts, i );
        vPartSupp = (Vec_Int_t *)Vec_PtrEntry( vPartSupps, i );
        // create the new miter
        pNew = Aig_ManStart( 1000 );
        // create the PIs
        for  k = 0; k < Vec_IntSize(vPartSupp); k++  {
            Aig_ObjCreateCi( pNew );
        }
        // copy the components
        vNodes1 = Aig_ManDupPart( pNew, p1, vPart, vPartSupp, 0 );
        vNodes2 = Aig_ManDupPart( pNew, p2, vPart, vPartSupp, 0 );
        // create the miter
        pMiter = Aig_MiterTwo( pNew, vNodes1, vNodes2 );
        Vec_PtrFree( vNodes1 );
        Vec_PtrFree( vNodes2 );
        // create the output
        Aig_ObjCreateCo( pNew, pMiter );
        // clean up
        Aig_ManCleanup( pNew );
        Vec_PtrPush( vMiters, pNew );
    }
    Vec_VecFree( (Vec_Vec_t *)vParts );
    Vec_VecFree( (Vec_Vec_t *)vPartSupps );
    return vMiters;
}

/**Function*************************************************************

  Synopsis    [Performs partitioned choice computation.]

  Description [Assumes that each output in the second AIG cannot have 
  more supp vars than the same output in the first AIG.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManChoicePartitioned(vAigs *Vec_Ptr_t, nPartSize int32, nConfMax int32, nLevelMax int32, fVerbose int32) *Aig_Man_t {
//    extern int Cmd_CommandExecute( void * pAbc, char * sCommand );
//    extern void * Abc_FrameGetGlobalFrame();
//    extern Aig_Man_t * Fra_FraigChoice( Aig_Man_t * pManAig, int nConfMax, int nLevelMax );

    var vPios *Vec_Ptr_t;
    var vOutsTotal, vOuts *Vec_Ptr_t;
    var pAigTotal, pAigPart, pAig, pTemp *Aig_Man_t;
    var vPart, vPartSupp *Vec_Int_t;
    var vParts *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var ppData *any;
    var i, k, m, nIdMax int32;
    assert(Vec_PtrSize(vAigs) > 1, "Vec_PtrSize(vAigs) > 1");

    // compute the total number of IDs
    nIdMax = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        nIdMax += Aig_ManObjNumMax(pAig);
    }

    // partition the first AIG in the array
    pAig = (Aig_Man_t *)Vec_PtrEntry( vAigs, 0 );
    vParts = Aig_ManPartitionSmart( pAig, nPartSize, 0, nil );

    // start the total fraiged AIG
    pAigTotal = Aig_ManStartFrom( pAig );
    Aig_ManReprStart( pAigTotal, nIdMax );
    vOutsTotal = Vec_PtrStart( Aig_ManCoNum(pAig) );

    // set the PI numbers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            pObj.pNext = (Aig_Obj_t *)int64(k);
        }
    }

//    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "unset progressbar" );

    // create the total fraiged AIG
    vPartSupp = Vec_IntAlloc( 100 ); // maps part PI num into total PI num
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        // derive the partition AIG
        pAigPart = Aig_ManStart( 5000 );
//        pAigPart->pName = Extra_UtilStrsav( pAigPart->pName );
        Vec_IntClear( vPartSupp );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vAigs)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vAigs, k)), 1); k++  {
            vOuts = Aig_ManDupPart( pAigPart, pAig, vPart, vPartSupp, 0 );
            if k == 0 {
                // inlining of macro Vec_PtrForEachEntry
                for  m = 0; (m < Vec_PtrSize(vOuts)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vOuts, m)), 1); m++  {
                    Aig_ObjCreateCo( pAigPart, pObj );
                }
            }
            Vec_PtrFree( vOuts );
        }
        // derive the total AIG from the partitioned AIG
        vOuts = Aig_ManDupPart( pAigTotal, pAigPart, vPart, vPartSupp, 1 );
        // add to the outputs
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vOuts)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vOuts, k)), 1); k++  {
            assert(Vec_PtrEntry( vOutsTotal, Vec_IntEntry(vPart,k) ) == nil, "Vec_PtrEntry( vOutsTotal, Vec_IntEntry(vPart,k) ) == nil");
            Vec_PtrWriteEntry( vOutsTotal, Vec_IntEntry(vPart,k), pObj );
        }
        Vec_PtrFree( vOuts );
        // store contents of pData pointers
        // inlining of macro ABC_ALLOC
        ppData = ((void * *) malloc(sizeof(void *) * (size_t)(Aig_ManObjNumMax(pAigPart))));
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAigPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigPart.vObjs, k)), 1); k++  { if (pObj) == nil {} else {
            ppData[k] = pObj.pData; }
        }
        // report the process
        if fVerbose != 0 {
        fmt.Printf("Part %4d  (out of %4d)  PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, Vec_PtrSize(vParts), Aig_ManCiNum(pAigPart), Aig_ManCoNum(pAigPart), Aig_ManNodeNum(pAigPart), Aig_ManLevelNum(pAigPart)); }
        // compute equivalence classes (to be stored in pNew->pReprs)
        pAig = Fra_FraigChoice( pAigPart, nConfMax, nLevelMax );
        Aig_ManStop( pAig );
        // reset the pData pointers
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAigPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigPart.vObjs, k)), 1); k++  { if (pObj) == nil {} else {
            pObj.pData = ppData[k]; }
        }
        // inlining of macro ABC_FREE
        ((ppData) ? (free((char *) (ppData)), (ppData) = 0) : 0);
        // transfer representatives to the total AIG
        if pAigPart.pReprs {
            Aig_ManTransferRepr( pAigTotal, pAigPart ); }
        Aig_ManStop( pAigPart );
    }
    if fVerbose != 0 {
    fmt.Printf("                                                                                          \r"); }
    Vec_VecFree( (Vec_Vec_t *)vParts );
    Vec_IntFree( vPartSupp );

//    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "set progressbar" );

    // clear the PI numbers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            pObj.pNext = nil;
        }
    }

    // add the outputs in the same order
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOutsTotal)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vOutsTotal, i)), 1); i++  {
        Aig_ObjCreateCo( pAigTotal, pObj );
    }
    Vec_PtrFree( vOutsTotal );

    // derive the result of choicing
    pAig = Aig_ManRehash( pAigTotal );
    // create the equivalent nodes lists
    Aig_ManMarkValidChoices( pAig );
    // reconstruct the network
    vPios = Aig_ManOrderPios( pAig, (Aig_Man_t *)Vec_PtrEntry(vAigs,0) );
    pAig = Aig_ManDupDfsGuided( pTemp = pAig, vPios );
    Aig_ManStop( pTemp );
    Vec_PtrFree( vPios );
    // duplicate the timing manager
    pTemp = (Aig_Man_t *)Vec_PtrEntry( vAigs, 0 );
    if pTemp.pManTime {
        pAig.pManTime = Tim_ManDup( (Tim_Man_t *)pTemp.pManTime, 0 ); }
    // reset levels
    Aig_ManChoiceLevel( pAig );
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Performs partitioned choice computation.]

  Description [Assumes that each output in the second AIG cannot have 
  more supp vars than the same output in the first AIG.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFraigPartitioned(pAig *Aig_Man_t, nPartSize int32, nConfMax int32, nLevelMax int32, fVerbose int32) *Aig_Man_t {
//    extern Aig_Man_t * Fra_FraigChoice( Aig_Man_t * pManAig, int nConfMax, int nLevelMax );

    var pAigPart, pAigTemp *Aig_Man_t;
    var vPart *Vec_Int_t;
    var vParts *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var ppData *any;
    var i, k int32;

    // partition the outputs of the AIG
    vParts = Aig_ManPartitionNaive( pAig, nPartSize );

    // start the equivalence classes
    Aig_ManReprStart( pAig, Aig_ManObjNumMax(pAig) );

    // set the PI numbers
    Aig_ManSetCioIds( pAig );

    // create the total fraiged AIG
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        // derive the partition AIG
        pAigPart = Aig_ManDupPartAll( pAig, vPart );
        // store contents of pData pointers
        // inlining of macro ABC_ALLOC
        ppData = ((void * *) malloc(sizeof(void *) * (size_t)(Aig_ManObjNumMax(pAigPart))));
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAigPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigPart.vObjs, k)), 1); k++  { if (pObj) == nil {} else {
            ppData[k] = pObj.pData; }
        }
        // report the process
        if fVerbose != 0 {
        fmt.Printf("Part %4d  (out of %4d)  PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, Vec_PtrSize(vParts), Aig_ManCiNum(pAigPart), Aig_ManCoNum(pAigPart), Aig_ManNodeNum(pAigPart), Aig_ManLevelNum(pAigPart)); }
        // compute equivalence classes (to be stored in pNew->pReprs)
        pAigTemp = Fra_FraigChoice( pAigPart, nConfMax, nLevelMax );
        Aig_ManStop( pAigTemp );
        // reset the pData pointers
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAigPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigPart.vObjs, k)), 1); k++  { if (pObj) == nil {} else {
            pObj.pData = ppData[k]; }
        }
        // inlining of macro ABC_FREE
        ((ppData) ? (free((char *) (ppData)), (ppData) = 0) : 0);
        // transfer representatives to the total AIG
        if pAigPart.pReprs {
            Aig_ManTransferRepr( pAig, pAigPart ); }
        Aig_ManStop( pAigPart );
    }
    if fVerbose != 0 {
    fmt.Printf("                                                                                          \r"); }
    Vec_VecFree( (Vec_Vec_t *)vParts );

    // clear the PI numbers
    Aig_ManCleanCioIds( pAig );

    // derive the result of choicing
    return Aig_ManDupRepr( pAig, 0 );
}




/**Function*************************************************************

  Synopsis    [Set the representative.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjSetRepr_(p *Aig_Man_t, pNode1 *Aig_Obj_t, pNode2 *Aig_Obj_t) {
    assert(p.pReprs != nil, "p.pReprs != nil");
    assert(!Aig_IsComplement(pNode1), "!Aig_IsComplement(pNode1)");
    assert(!Aig_IsComplement(pNode2), "!Aig_IsComplement(pNode2)");
    assert(pNode1.Id < p.nReprsAlloc, "pNode1.Id < p.nReprsAlloc");
    assert(pNode2.Id < p.nReprsAlloc, "pNode2.Id < p.nReprsAlloc");
    if pNode1 == pNode2 {
        return; }
    if pNode1.Id < pNode2.Id {
        p.pReprs[pNode2.Id] = pNode1;
    } else {
        p.pReprs[pNode1.Id] = pNode2; }
}

/**Function*************************************************************

  Synopsis    [Constructively accumulates choices.]

  Description [pNew is a new AIG with choices under construction.
  pPrev is the AIG preceding pThis in the order of deriving choices.
  pThis is the current AIG to be added to pNew while creating choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManChoiceConstructiveOne(pNew *Aig_Man_t, pPrev *Aig_Man_t, pThis *Aig_Man_t) {
    var pObj, pObjNew *Aig_Obj_t;
    var i int32;
    assert(Aig_ManCiNum(pNew) == Aig_ManCiNum(pPrev), "Aig_ManCiNum(pNew) == Aig_ManCiNum(pPrev)");
    assert(Aig_ManCiNum(pNew) == Aig_ManCiNum(pThis), "Aig_ManCiNum(pNew) == Aig_ManCiNum(pThis)");
    assert(Aig_ManCoNum(pNew) == Aig_ManCoNum(pPrev), "Aig_ManCoNum(pNew) == Aig_ManCoNum(pPrev)");
    assert(Aig_ManCoNum(pNew) == Aig_ManCoNum(pThis), "Aig_ManCoNum(pNew) == Aig_ManCoNum(pThis)");
    // make sure the nodes of pPrev point to pNew
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNew.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pNew.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkB = 1; }
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPrev.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPrev.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(Aig_Regular((Aig_Obj_t *)pObj.pData).fMarkB, "Aig_Regular((Aig_Obj_t *)pObj.pData).fMarkB"); }
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNew.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pNew.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkB = 0; }
    }
    // make sure the nodes of pThis point to pPrev
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPrev.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPrev.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkB = 1; }
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPrev.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPrev.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkB = 0; }
    }
    // remap nodes of pThis on top of pNew using pPrev
    pObj = Aig_ManConst1(pThis);
    pObj.pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pThis.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pThis.vCis, i)), 1); i++  {
        pObj.pData = Aig_ManCi(pNew, i);
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pThis.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pThis.vCos, i)), 1); i++  {
        pObj.pData = Aig_ManCo(pNew, i);
    }
    // go through the nodes in the topological order
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pThis.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pThis.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // set the inputs of POs as equivalent
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pThis.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pThis.vCos, i)), 1); i++  {
        pObjNew = Aig_ObjFanin0( Aig_ManCo(pNew,i) );
        // pObjNew and Aig_ObjFanin0(pObj)->pData are equivalent
        Aig_ObjSetRepr_( pNew, pObjNew, Aig_Regular((Aig_Obj_t *)Aig_ObjFanin0(pObj).pData) );
    }
}

/**Function*************************************************************

  Synopsis    [Constructively accumulates choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManChoiceEval(p *Aig_Man_t) {
    var vSupp *Vec_Ptr_t;
    var pNode, pTemp *Aig_Obj_t;
    var i, Counter int32;

    vSupp = Vec_PtrAlloc( 100 );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pNode) == nil || !Aig_ObjIsNode(pNode) {} else {
        if !Aig_ObjIsChoice(p, pNode) {
            continue; } 
        Counter = 0;
        for  pTemp = pNode; pTemp; pTemp = Aig_ObjEquiv(p, pTemp)  {
            Counter++;
        }
        fmt.Printf("Choice node = %5d. Level = %2d. Choices = %d. { ", pNode.Id, pNode.Level, Counter);
        for  pTemp = pNode; pTemp; pTemp = Aig_ObjEquiv(p, pTemp)  {
            Counter = Aig_NodeMffcSupp( p, pTemp, 0, vSupp );
            fmt.Printf("S=%d N=%d L=%d  ", Vec_PtrSize(vSupp), Counter, pTemp.Level);
        }
        fmt.Printf("}\n");
    }
    }
    Vec_PtrFree( vSupp );
}

/**Function*************************************************************

  Synopsis    [Constructively accumulates choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManChoiceConstructive(vAigs *Vec_Ptr_t, fVerbose int32) *Aig_Man_t {
    var vPios *Vec_Ptr_t;
    var pNew, pThis, pPrev, pTemp *Aig_Man_t;
    var i int32;
    // start AIG with choices
    pPrev = (Aig_Man_t *)Vec_PtrEntry( vAigs, 0 );
    pNew = Aig_ManDupOrdered( pPrev );
    // create room for equivalent nodes and representatives
    assert(pNew.pReprs == nil, "pNew.pReprs == nil");
    pNew.nReprsAlloc = Vec_PtrSize(vAigs) * Aig_ManObjNumMax(pNew);
    // inlining of macro ABC_ALLOC
    pNew.pReprs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(pNew.nReprsAlloc)));
    memset( pNew.pReprs, 0, sizeof(Aig_Obj_t *) * pNew.nReprsAlloc );
    // add other AIGs one by one
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vAigs)) && (((pThis) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        Aig_ManChoiceConstructiveOne( pNew, pPrev, pThis );
        pPrev = pThis;
    }
    // derive the result of choicing
    pNew = Aig_ManRehash( pNew );
    // create the equivalent nodes lists
    Aig_ManMarkValidChoices( pNew );
    // reconstruct the network
    vPios = Aig_ManOrderPios( pNew, (Aig_Man_t *)Vec_PtrEntry( vAigs, 0 ) );
    pNew = Aig_ManDupDfsGuided( pTemp = pNew, vPios );
    Aig_ManStop( pTemp );
    Vec_PtrFree( vPios );
    // duplicate the timing manager
    pTemp = (Aig_Man_t *)Vec_PtrEntry( vAigs, 0 );
    if pTemp.pManTime {
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)pTemp.pManTime, 0 ); }
    // reset levels
    Aig_ManChoiceLevel( pNew );
    return pNew;
}

/*
    Vec_Ptr_t * vPios;
    vPios = Aig_ManOrderPios( pMan, pAig );
    Vec_PtrFree( vPios );
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

