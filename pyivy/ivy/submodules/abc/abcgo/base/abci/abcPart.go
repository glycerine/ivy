package abci

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [abcPart.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Output partitioning package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcPart.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "base/cmd/cmd.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

       
type Supp_Man_t struct {
    nChunkSize              int32;    // the size of one chunk of memory (~1 MB)
    nStepSize              int32;     // the step size in saving memory (~64 bytes)
    pFreeBuf            string;      // the pointer to free memory
    nFreeSize              int32;     // the size of remaining free memory
    vMemory       *Vec_Ptr_t;       // the memory allocated
    vFree       *Vec_Ptr_t;         // the vector of free pieces of memory
}

       
type Supp_One_t struct {
    nRefs              int32;         // the number of references
    nOuts              int32;         // the number of outputs
    nOutsAlloc              int32;    // the array size
    pOuts [0]int32;      // the array of outputs
}

func Supp_SizeType(nSize int32, nStepSize int32) int32 { return nSize / nStepSize + ((nSize % nStepSize) > 0); }
func Supp_OneNext(pPart string) string { return *((char **)pPart);                             }
func Supp_OneSetNext(pPart string, pNext string) { *((char **)pPart) = pNext;                            }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start the memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Supp_ManStart(nChunkSize int32, nStepSize int32) *Supp_Man_t {
    var p *Supp_Man_t;
    // inlining of macro ABC_ALLOC
    p = ((Supp_Man_t *) malloc(sizeof(Supp_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Supp_Man_t) );
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
func Supp_ManStop(p *Supp_Man_t) {
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
func Supp_ManFetch(p *Supp_Man_t, nSize int32) string {
    var Type, nSizeReal int32;
    var pMemory string;
    assert(nSize > 0, "nSize > 0");
    Type = Supp_SizeType( nSize, p.nStepSize );
    Vec_PtrFillExtra( p.vFree, Type + 1, nil );
    if pMemory = (char *)Vec_PtrEntry( p.vFree, Type ) {
        Vec_PtrWriteEntry( p.vFree, Type, Supp_OneNext(pMemory) );
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
func Supp_ManRecycle(p *Supp_Man_t, pMemory string, nSize int32) {
    var Type int32;
    Type = Supp_SizeType( nSize, p.nStepSize );
    Vec_PtrFillExtra( p.vFree, Type + 1, nil );
    Supp_OneSetNext( pMemory, (char *)Vec_PtrEntry(p.vFree, Type) );
    Vec_PtrWriteEntry( p.vFree, Type, pMemory );
}

/**Function*************************************************************

  Synopsis    [Fetches the memory entry of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Supp_ManFetchEntry(p *Supp_Man_t, nWords int32, nRefs int32) *Supp_One_t {
    var pPart *Supp_One_t;
    pPart = (Supp_One_t *)Supp_ManFetch( p, sizeof(Supp_One_t) + sizeof(int) * nWords );
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
func Supp_ManRecycleEntry(p *Supp_Man_t, pEntry *Supp_One_t) {
    assert(pEntry.nOuts <= pEntry.nOutsAlloc, "pEntry.nOuts <= pEntry.nOutsAlloc");
    assert(pEntry.nOuts >= pEntry.nOutsAlloc/2, "pEntry.nOuts >= pEntry.nOutsAlloc/2");
    Supp_ManRecycle( p, (char *)pEntry, sizeof(Supp_One_t) + sizeof(int) * pEntry.nOutsAlloc );
}

/**Function*************************************************************

  Synopsis    [Merges two entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Supp_ManMergeEntry(pMan *Supp_Man_t, p1 *Supp_One_t, p2 *Supp_One_t, nRefs int32) *Supp_One_t {
    var p *Supp_One_t  = Supp_ManFetchEntry( pMan, p1.nOuts + p2.nOuts, nRefs );
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
func Supp_ManTransferEntry(p *Supp_One_t) *Vec_Int_t {
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

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsNatural(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj, pNext *Abc_Obj_t;
    var i, k int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    vNodes = Vec_PtrAlloc( Abc_NtkObjNum(pNtk) );
    Abc_NtkIncrementTravId( pNtk );
    // add the constant-1 nodes
    pObj = Abc_AigConst1(pNtk);
    Abc_NodeSetTravIdCurrent( pObj );
    Vec_PtrPush( vNodes, pObj );
    // add the CIs/nodes/COs in the topological order
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        // check the fanins and add CIs
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if Abc_ObjIsCi(pNext) && !Abc_NodeIsTravIdCurrent(pNext) {
                Abc_NodeSetTravIdCurrent( pNext );
                Vec_PtrPush( vNodes, pNext );
            }
        }
        // add the node
        Vec_PtrPush( vNodes, pObj );
        // check the fanouts and add COs
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
            if Abc_ObjIsCo(pNext) && !Abc_NodeIsTravIdCurrent(pNext) {
                Abc_NodeSetTravIdCurrent( pNext );
                Vec_PtrPush( vNodes, pNext );
            }
        }
    }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes supports of the POs.]

  Description [Returns the ptr-vector of int-vectors.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkComputeSupportsSmart(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vSupports *Vec_Ptr_t;
    var vNodes *Vec_Ptr_t;
    var vSupp *Vec_Int_t;
    var p *Supp_Man_t;
    var pPart0, pPart1 *Supp_One_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // set the number of PIs/POs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pNext = (Abc_Obj_t *)(ABC_PTRINT_T)i;
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObj.pNext = (Abc_Obj_t *)(ABC_PTRINT_T)i;
    }
    // start the support computation manager
    p = Supp_ManStart( 1 << 20, 1 << 6 );
    // consider objects in the topological order
    vSupports = Vec_PtrAlloc( Abc_NtkCoNum(pNtk) );
    Abc_NtkCleanCopy(pNtk);
    // order the nodes so that the PIs and POs follow naturally
    vNodes = Abc_NtkDfsNatural( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjIsNode(pObj) {
            pPart0 = (Supp_One_t *)Abc_ObjFanin0(pObj).pCopy;
            pPart1 = (Supp_One_t *)Abc_ObjFanin1(pObj).pCopy;
            pObj.pCopy = (Abc_Obj_t *)Supp_ManMergeEntry( p, pPart0, pPart1, Abc_ObjFanoutNum(pObj) );
            assert(pPart0.nRefs > 0, "pPart0.nRefs > 0");
            if --pPart0.nRefs == 0 {
                Supp_ManRecycleEntry( p, pPart0 ); }
            assert(pPart1.nRefs > 0, "pPart1.nRefs > 0");
            if --pPart1.nRefs == 0 {
                Supp_ManRecycleEntry( p, pPart1 ); }
            continue;
        }
        if Abc_ObjIsCo(pObj) {
            pPart0 = (Supp_One_t *)Abc_ObjFanin0(pObj).pCopy;
            // only save the CO if it is non-trivial
            if Abc_ObjIsNode(Abc_ObjFanin0(pObj)) {
                vSupp = Supp_ManTransferEntry(pPart0);
                Vec_IntPush( vSupp, int32((ABC_PTRINT_T))pObj.pNext );
                Vec_PtrPush( vSupports, vSupp );
            }
            assert(pPart0.nRefs > 0, "pPart0.nRefs > 0");
            if --pPart0.nRefs == 0 {
                Supp_ManRecycleEntry( p, pPart0 ); }
            continue;
        }
        if Abc_ObjIsCi(pObj) {
            if Abc_ObjFanoutNum(pObj) {
                pPart0 = (Supp_One_t *)Supp_ManFetchEntry( p, 1, Abc_ObjFanoutNum(pObj) );
                pPart0.pOuts[ pPart0.nOuts++ ] = int32((ABC_PTRINT_T))pObj.pNext;
                pObj.pCopy = (Abc_Obj_t *)pPart0;
            }
            continue;
        }
        if pObj == Abc_AigConst1(pNtk) {
            if Abc_ObjFanoutNum(pObj) {
                pObj.pCopy = (Abc_Obj_t *)Supp_ManFetchEntry( p, 0, Abc_ObjFanoutNum(pObj) ); }
            continue;
        }
        assert(false, "0");
    }
    Vec_PtrFree( vNodes );
//printf( "Memory usage = %d MB.\n", Vec_PtrSize(p->vMemory) * p->nChunkSize / (1<<20) );
    Supp_ManStop( p );
    // sort supports by size
    Vec_VecSort( (Vec_Vec_t *)vSupports, 1 );
    // clear the number of PIs/POs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pNext = nil;
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObj.pNext = nil;
    }
/*
    Vec_PtrForEachEntry( Vec_Int_t *, vSupports, vSupp, i )
        printf( "%d ", Vec_IntSize(vSupp) );
    printf( "\n" );
*/
    return vSupports;
}


/**Function*************************************************************

  Synopsis    [Computes supports of the POs using naive method.]

  Description [Returns the ptr-vector of int-vectors.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkComputeSupportsNaive(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vSupp, vSupports *Vec_Ptr_t;
    var vSuppI *Vec_Int_t;
    var pObj, pTemp *Abc_Obj_t;
    var i, k int32;
    // set the PI numbers
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pNext = (Abc_Obj_t *)(ABC_PTRINT_T)i;
    }
    // save the CI numbers
    vSupports = Vec_PtrAlloc( Abc_NtkCoNum(pNtk) );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if !Abc_ObjIsNode(Abc_ObjFanin0(pObj)) {
            continue; }
        vSupp = Abc_NtkNodeSupport( pNtk, &pObj, 1 );
        vSuppI = (Vec_Int_t *)vSupp;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSupp)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, k)), 1); k++  {
            Vec_IntWriteEntry( vSuppI, k, int32((ABC_PTRINT_T))pTemp.pNext );
        }
        Vec_IntSort( vSuppI, 0 );
        // append the number of this output
        Vec_IntPush( vSuppI, i );
        // save the support in the vector
        Vec_PtrPush( vSupports, vSuppI );
    }
    // clean the CI numbers
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pNext = nil;
    }
    // sort supports by size
    Vec_VecSort( (Vec_Vec_t *)vSupports, 1 );
/*
    Vec_PtrForEachEntry( Vec_Int_t *, vSupports, vSuppI, i )
        printf( "%d ", Vec_IntSize(vSuppI) );
    printf( "\n" );
*/
    return vSupports;
}

/**Function*************************************************************

  Synopsis    [Start bitwise support representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSuppCharStart(vOne *Vec_Int_t, nPis int32) *uint32 {
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

  Synopsis    [Add to bitwise support representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSuppCharAdd(pBuffer *uint32, vOne *Vec_Int_t, nPis int32) {
    var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOne)) && (((Entry) = Vec_IntEntry(vOne, i)), 1); i++  {
        assert(Entry < nPis, "Entry < nPis");
        Abc_InfoSetBit( pBuffer, Entry );
    }
}

/**Function*************************************************************

  Synopsis    [Find the common variables using bitwise support representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSuppCharCommon(pBuffer *uint32, vOne *Vec_Int_t) int32 {
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
func Abc_NtkPartitionSmartFindPart(vPartSuppsAll *Vec_Ptr_t, vPartsAll *Vec_Ptr_t, vPartSuppsChar *Vec_Ptr_t, nSuppSizeLimit int32, vOne *Vec_Int_t) int32 {
/*
    Vec_Int_t * vPartSupp, * vPart;
    double Attract, Repulse, Cost, CostBest;
    int i, nCommon, iBest;
    iBest = -1;
    CostBest = 0.0;
    Vec_PtrForEachEntry( Vec_Int_t *, vPartSuppsAll, vPartSupp, i )
    {
        vPart = Vec_PtrEntry( vPartsAll, i );
        if ( nPartSizeLimit > 0 && Vec_IntSize(vPart) >= nPartSizeLimit )
            continue;
        nCommon = Vec_IntTwoCountCommon( vPartSupp, vOne );
        if ( nCommon == 0 )
            continue;
        if ( nCommon == Vec_IntSize(vOne) )
            return i;
        Attract = 1.0 * nCommon / Vec_IntSize(vOne);
        if ( Vec_IntSize(vPartSupp) < 100 )
            Repulse = 1.0;
        else
            Repulse = log10( Vec_IntSize(vPartSupp) / 10.0 );
        Cost = pow( Attract, pow(Repulse, 5.0) );
        if ( CostBest < Cost )
        {
            CostBest = Cost;
            iBest = i;
        }
    }
    if ( CostBest < 0.6 )
        return -1;
    return iBest;
*/

    var vPartSupp *Vec_Int_t;//, * vPart;
    var Attract, Repulse, Value, ValueBest int32;
    var i, nCommon, iBest int32;
//    int nCommon2;
    iBest = -1;
    ValueBest = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsAll)) && (((vPartSupp) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsAll, i)), 1); i++  {
        // skip partitions with too many outputs
//        vPart = Vec_PtrEntry( vPartsAll, i );
//        if ( nSuppSizeLimit > 0 && Vec_IntSize(vPart) >= nSuppSizeLimit )
//            continue;
        // find the number of common variables between this output and the partitions
//        nCommon2 = Vec_IntTwoCountCommon( vPartSupp, vOne );
        nCommon = Abc_NtkSuppCharCommon( (unsigned *)Vec_PtrEntry(vPartSuppsChar, i), vOne );
//        assert( nCommon2 == nCommon );
        // if no common variables, continue searching
        if nCommon == 0 {
            continue; }
        // if all variables are common, the best partition if found
        if nCommon == Vec_IntSize(vOne) {
            return i; }
        // skip partitions whose size exceeds the limit
        if nSuppSizeLimit > 0 && Vec_IntSize(vPartSupp) >= 2 * nSuppSizeLimit {
            continue; }
        // figure out might be the good partition for this one
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
func Abc_NtkPartitionPrint(pNtk *Abc_Ntk_t, vPartsAll *Vec_Ptr_t, vPartSuppsAll *Vec_Ptr_t) {
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
//    assert( Counter == Abc_NtkCoNum(pNtk) );
    fmt.Printf("\nTotal = %d. Outputs = %d.\n", Counter, Abc_NtkCoNum(pNtk));
}

/**Function*************************************************************

  Synopsis    [Perform the smart partitioning.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartitionCompact(vPartsAll *Vec_Ptr_t, vPartSuppsAll *Vec_Ptr_t, nSuppSizeLimit int32) {
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
    if vPart {
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

  Description [Returns the ptr-vector of int-vectors.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartitionSmart(pNtk *Abc_Ntk_t, nSuppSizeLimit int32, fVerbose int32) *Vec_Ptr_t {
    var pProgress *ProgressBar;
    var vPartSuppsChar *Vec_Ptr_t;
    var vSupps, vPartsAll, vPartsAll2, vPartSuppsAll *Vec_Ptr_t;
    var vOne, vPart, vPartSupp, vTemp *Vec_Int_t;
    var i, iPart, iOut, timeFind int32;
    var clk, clk2 abctime;

    // compute the supports for all outputs
clk = Abc_Clock();
//    vSupps = Abc_NtkComputeSupportsNaive( pNtk );
    vSupps = Abc_NtkComputeSupportsSmart( pNtk );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Supps"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
    // start char-based support representation
    vPartSuppsChar = Vec_PtrAlloc( 1000 );

    // create partitions
clk = Abc_Clock();
    vPartsAll = Vec_PtrAlloc( 256 );
    vPartSuppsAll = Vec_PtrAlloc( 256 );
    pProgress = Extra_ProgressBarStart( os.Stdout, Vec_PtrSize(vSupps) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupps)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vSupps, i)), 1); i++  {
        Extra_ProgressBarUpdate( pProgress, i, nil );
//        if ( i % 1000 == 0 )
//            printf( "CIs = %6d. COs = %6d. Processed = %6d (out of %6d). Parts = %6d.\r", 
//                Abc_NtkCiNum(pNtk), Abc_NtkCoNum(pNtk), i, Vec_PtrSize(vSupps), Vec_PtrSize(vPartsAll) ); 
        // get the output number
        iOut = Vec_IntPop(vOne);
        // find closely matching part
clk2 = Abc_Clock();
        iPart = Abc_NtkPartitionSmartFindPart( vPartSuppsAll, vPartsAll, vPartSuppsChar, nSuppSizeLimit, vOne );
timeFind += Abc_Clock() - clk2;
        if iPart == -1 {
            // create new partition
            vPart = Vec_IntAlloc( 32 );
            Vec_IntPush( vPart, iOut );
            // create new partition support
            vPartSupp = Vec_IntDup( vOne );
            // add this partition and its support
            Vec_PtrPush( vPartsAll, vPart );
            Vec_PtrPush( vPartSuppsAll, vPartSupp );

            Vec_PtrPush( vPartSuppsChar, Abc_NtkSuppCharStart(vOne, Abc_NtkCiNum(pNtk)) );
        
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

            Abc_NtkSuppCharAdd( (unsigned *)Vec_PtrEntry(vPartSuppsChar, iPart), vOne, Abc_NtkCiNum(pNtk) );
        }
    }
    Extra_ProgressBarStop( pProgress );

    // stop char-based support representation
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartSuppsChar)) && (((vTemp) = (Vec_Int_t *)Vec_PtrEntry(vPartSuppsChar, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((vTemp) ? (free((char *) (vTemp)), (vTemp) = 0) : 0);
    }
    Vec_PtrFree( vPartSuppsChar );

//printf( "\n" );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Parts"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
//ABC_PRT( "Find ", timeFind );
}

clk = Abc_Clock();
    // remember number of supports
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
//    Abc_NtkPartitionPrint( pNtk, vPartsAll, vPartSuppsAll );
    Abc_NtkPartitionCompact( vPartsAll, vPartSuppsAll, nSuppSizeLimit );

if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Comps"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
    if fVerbose != 0 {
    fmt.Printf("Created %d partitions.\n", Vec_PtrSize(vPartsAll)); }
//    Abc_NtkPartitionPrint( pNtk, vPartsAll, vPartSuppsAll );

    // cleanup
    Vec_VecFree( (Vec_Vec_t *)vSupps );
    Vec_VecFree( (Vec_Vec_t *)vPartSuppsAll );
/*
    // converts from intergers to nodes
    Vec_PtrForEachEntry( Vec_Int_t *, vPartsAll, vPart, iPart )
    {
        vPartPtr = Vec_PtrAlloc( Vec_IntSize(vPart) );
        Vec_IntForEachEntry( vPart, iOut, i )
            Vec_PtrPush( vPartPtr, Abc_NtkCo(pNtk, iOut) );
        Vec_IntFree( vPart );
        Vec_PtrWriteEntry( vPartsAll, iPart, vPartPtr );
    }
*/
    return vPartsAll;
}

/**Function*************************************************************

  Synopsis    [Perform the naive partitioning.]

  Description [Returns the ptr-vector of int-vectors.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartitionNaive(pNtk *Abc_Ntk_t, nPartSize int32) *Vec_Ptr_t {
    var vParts *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var nParts, i int32;
    nParts = (Abc_NtkCoNum(pNtk) / nPartSize) + ((Abc_NtkCoNum(pNtk) % nPartSize) > 0);
    vParts = (Vec_Ptr_t *)Vec_VecStart( nParts );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vParts, i / nPartSize), i );
    }
    return vParts;
}

/**Function*************************************************************

  Synopsis    [Converts from intergers to pointers for the given network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConvertCos(pNtk *Abc_Ntk_t, vOuts *Vec_Int_t, vOutsPtr *Vec_Ptr_t) {
    var Out, i int32;
    Vec_PtrClear( vOutsPtr );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOuts)) && (((Out) = Vec_IntEntry(vOuts, i)), 1); i++  {
        Vec_PtrPush( vOutsPtr, Abc_NtkCo(pNtk, Out) );
    }
}

/**Function*************************************************************

  Synopsis    [Returns representative of the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartStitchFindRepr_rec(vEquiv *Vec_Ptr_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    var pRepr *Abc_Obj_t;
    pRepr = (Abc_Obj_t *)Vec_PtrEntry( vEquiv, pObj.Id );
    if pRepr == nil || pRepr == pObj {
        return pObj; }
    return Abc_NtkPartStitchFindRepr_rec( vEquiv, pRepr );
}

/**Function*************************************************************

  Synopsis    [Returns the representative of the fanin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartStitchCopy0(vEquiv *Vec_Ptr_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    var pFan *Abc_Obj_t  = Abc_ObjFanin0( pObj );
    var pRepr *Abc_Obj_t  = Abc_NtkPartStitchFindRepr_rec( vEquiv, pFan );
    return Abc_ObjNotCond( pRepr.pCopy, pRepr.fPhase ^ pFan.fPhase ^ int32(Abc_ObjFaninC1(pObj)) );
}
func Abc_NtkPartStitchCopy1(vEquiv *Vec_Ptr_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    var pFan *Abc_Obj_t  = Abc_ObjFanin1( pObj );
    var pRepr *Abc_Obj_t  = Abc_NtkPartStitchFindRepr_rec( vEquiv, pFan );
    return Abc_ObjNotCond( pRepr.pCopy, pRepr.fPhase ^ pFan.fPhase ^ int32(Abc_ObjFaninC1(pObj)) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjChild0Next(pObj *Abc_Obj_t) *Hop_Obj_t { return Hop_NotCond( (Hop_Obj_t *)Abc_ObjFanin0(pObj).pNext, Abc_ObjFaninC0(pObj) ); }
func Hop_ObjChild1Next(pObj *Abc_Obj_t) *Hop_Obj_t { return Hop_NotCond( (Hop_Obj_t *)Abc_ObjFanin1(pObj).pNext, Abc_ObjFaninC1(pObj) ); }


/**Function*************************************************************

  Synopsis    [Stitches together several networks with choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartStartHop(pNtk *Abc_Ntk_t) *Hop_Man_t {
    var pMan *Hop_Man_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // start the HOP package
    pMan = Hop_ManStart();
    pMan.vObjs = Vec_PtrAlloc( Abc_NtkObjNumMax(pNtk) + 1  );
    Vec_PtrPush( pMan.vObjs, Hop_ManConst1(pMan) );
    // map constant node and PIs
    Abc_AigConst1(pNtk).pNext = (Abc_Obj_t *)Hop_ManConst1(pMan);
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pNext = (Abc_Obj_t *)Hop_ObjCreatePi(pMan);
    }
    // map the internal nodes
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        pObj.pNext = (Abc_Obj_t *)Hop_And( pMan, Hop_ObjChild0Next(pObj), Hop_ObjChild1Next(pObj) );
        assert(!Abc_ObjIsComplement(pObj.pNext), "!Abc_ObjIsComplement(pObj.pNext)");
    }
    }
    // set the choice nodes
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        if pObj.pCopy {
            ((Hop_Obj_t *)pObj.pNext).pData = pObj.pCopy.pNext; }
    }
    }
    // transfer the POs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Hop_ObjCreatePo( pMan, Hop_ObjChild0Next(pObj) );
    }
    // check the new manager
    if !Hop_ManCheck(pMan) {
        fmt.Printf("Abc_NtkPartStartHop: HOP manager check has failed.\n"); }
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Stitches together several networks with choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartStitchChoices(pNtk *Abc_Ntk_t, vParts *Vec_Ptr_t) *Abc_Ntk_t {
    extern Abc_Ntk_t * Abc_NtkHopRemoveLoops( Abc_Ntk_t * pNtk, Hop_Man_t * pMan );
    var pMan *Hop_Man_t;
    var vNodes *Vec_Ptr_t;
    var pNtkNew, pNtkTemp *Abc_Ntk_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k, iNodeId int32;

    // start a new network similar to the original one
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );

    // annotate parts to point to the new network
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((pNtkTemp) = (Abc_Ntk_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        assert(Abc_NtkIsStrash(pNtkTemp), "Abc_NtkIsStrash(pNtkTemp)");
        Abc_NtkCleanCopy( pNtkTemp );

        // map the CI nodes
        Abc_AigConst1(pNtkTemp).pCopy = Abc_AigConst1(pNtkNew);
        // inlining of macro Abc_NtkForEachCi
        for  k = 0; (k < Abc_NtkCiNum(pNtkTemp)) && (((pObj) = Abc_NtkCi(pNtkTemp, k)), 1); k++  {
            iNodeId = Nm_ManFindIdByNameTwoTypes( pNtkNew.pManName, Abc_ObjName(pObj), ABC_OBJ_PI, ABC_OBJ_BO );
            if iNodeId == -1 {
                fmt.Printf("Cannot find CI node %s in the original network.\n", Abc_ObjName(pObj));
                return nil;
            }
            pObj.pCopy = Abc_NtkObj( pNtkNew, iNodeId );
        }

        // add the internal nodes while saving representatives
        vNodes = Abc_AigDfs( pNtkTemp, 1, 0 );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
            pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
            assert(!Abc_ObjIsComplement(pObj.pCopy), "!Abc_ObjIsComplement(pObj.pCopy)");
            if Abc_AigNodeIsChoice(pObj) {
                for  pFanin = (Abc_Obj_t *)pObj.pData; (pFanin != nil); pFanin = (Abc_Obj_t *)pFanin.pData  {
                    pFanin.pCopy.pCopy = pObj.pCopy;
                } }
        }
        Vec_PtrFree( vNodes );

        // map the CO nodes
        // inlining of macro Abc_NtkForEachCo
        for  k = 0; (k < Abc_NtkCoNum(pNtkTemp)) && (((pObj) = Abc_NtkCo(pNtkTemp, k)), 1); k++  {
            iNodeId = Nm_ManFindIdByNameTwoTypes( pNtkNew.pManName, Abc_ObjName(pObj), ABC_OBJ_PO, ABC_OBJ_BI );
            if iNodeId == -1 {
                fmt.Printf("Cannot find CO node %s in the original network.\n", Abc_ObjName(pObj));
                return nil;
            }
            pObj.pCopy = Abc_NtkObj( pNtkNew, iNodeId );
            Abc_ObjAddFanin( pObj.pCopy, Abc_ObjChild0Copy(pObj) );
        }
    }

    // connect the remaining POs
/*
    Abc_AigConst1(pNtk)->pCopy = Abc_AigConst1(pNtkNew);
    Abc_NtkForEachCi( pNtk, pObj, i )
        pObj->pCopy = Abc_NtkCi( pNtkNew, i );
    Abc_NtkForEachCo( pNtk, pObj, i )
        pObj->pCopy = Abc_NtkCo( pNtkNew, i );
*/
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if Abc_ObjFaninNum(pObj.pCopy) == 0 {
            Abc_ObjAddFanin( pObj.pCopy, Abc_ObjChild0Copy(pObj) ); }
    }

    // transform into the HOP manager
    pMan = Abc_NtkPartStartHop( pNtkNew );
    pNtkNew = Abc_NtkHopRemoveLoops( pNtkTemp = pNtkNew, pMan );
    Abc_NtkDelete( pNtkTemp );

    // check correctness of the new network
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkPartStitchChoices: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Stitches together several networks with choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigPartitioned(vStore *Vec_Ptr_t, pParams any) *Abc_Ntk_t {
    var vParts, vFraigs, vOnePtr *Vec_Ptr_t;
    var vOne *Vec_Int_t;
    var pNtk, pNtk2, pNtkAig, pNtkFraig *Abc_Ntk_t;
    var i, k int32;

    // perform partitioning
    pNtk = (Abc_Ntk_t *)Vec_PtrEntry( vStore, 0 );
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
//    vParts = Abc_NtkPartitionNaive( pNtk, 20 );
    vParts = Abc_NtkPartitionSmart( pNtk, 300, 0 );

    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "unset progressbar" );

    // fraig each partition
    vOnePtr = Vec_PtrAlloc( 1000 );
    vFraigs = Vec_PtrAlloc( Vec_PtrSize(vParts) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        // start the partition 
        Abc_NtkConvertCos( pNtk, vOne, vOnePtr );
        pNtkAig = Abc_NtkCreateConeArray( pNtk, vOnePtr, 0 );
        // add nodes to the partition
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = 1; (k < Vec_PtrSize(vStore)) && (((pNtk2) = (Abc_Ntk_t *)Vec_PtrEntry(vStore, k)), 1); k++  {
            Abc_NtkConvertCos( pNtk2, vOne, vOnePtr );
            Abc_NtkAppendToCone( pNtkAig, pNtk2, vOnePtr );
        }
        fmt.Printf("Fraiging part %4d  (out of %4d)  PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, Vec_PtrSize(vParts), Abc_NtkPiNum(pNtkAig), Abc_NtkPoNum(pNtkAig), Abc_NtkNodeNum(pNtkAig), Abc_AigLevel(pNtkAig));
        // fraig the partition
        pNtkFraig = Abc_NtkFraig( pNtkAig, pParams, 1, 0 );
        Vec_PtrPush( vFraigs, pNtkFraig );
        Abc_NtkDelete( pNtkAig );
    }
    fmt.Printf("                                                                                          \r");
    Vec_VecFree( (Vec_Vec_t *)vParts );

    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "set progressbar" );

    // derive the final network
    pNtkFraig = Abc_NtkPartStitchChoices( pNtk, vFraigs );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFraigs)) && (((pNtkAig) = (Abc_Ntk_t *)Vec_PtrEntry(vFraigs, i)), 1); i++  {
        Abc_NtkDelete( pNtkAig );
    }
    Vec_PtrFree( vFraigs );
    Vec_PtrFree( vOnePtr );
    return pNtkFraig;
}

/**Function*************************************************************

  Synopsis    [Stitches together several networks with choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFraigPartitionedTime(pNtk *Abc_Ntk_t, pParams any) {
    var vParts, vFraigs, vOnePtr *Vec_Ptr_t;
    var vOne *Vec_Int_t;
    var pNtkAig, pNtkFraig *Abc_Ntk_t;
    var i int32;
    var clk abctime  = Abc_Clock();

    // perform partitioning
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
//    vParts = Abc_NtkPartitionNaive( pNtk, 20 );
    vParts = Abc_NtkPartitionSmart( pNtk, 300, 0 );

    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "unset progressbar" );

    // fraig each partition
    vOnePtr = Vec_PtrAlloc( 1000 );
    vFraigs = Vec_PtrAlloc( Vec_PtrSize(vParts) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        Abc_NtkConvertCos( pNtk, vOne, vOnePtr );
        pNtkAig = Abc_NtkCreateConeArray( pNtk, vOnePtr, 0 );
        pNtkFraig = Abc_NtkFraig( pNtkAig, pParams, 0, 0 );
        Vec_PtrPush( vFraigs, pNtkFraig );
        Abc_NtkDelete( pNtkAig );

        fmt.Printf("Finished part %5d (out of %5d)\r", i+1, Vec_PtrSize(vParts));
    }
    Vec_VecFree( (Vec_Vec_t *)vParts );

    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "set progressbar" );

    // derive the final network
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFraigs)) && (((pNtkAig) = (Abc_Ntk_t *)Vec_PtrEntry(vFraigs, i)), 1); i++  {
        Abc_NtkDelete( pNtkAig );
    }
    Vec_PtrFree( vFraigs );
    Vec_PtrFree( vOnePtr );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Partitioned fraiging time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

