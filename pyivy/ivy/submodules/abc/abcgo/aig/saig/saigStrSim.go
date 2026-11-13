package saig

import "time"

/**CFile****************************************************************

  FileName    [saigStrSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Structural matching using simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigStrSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "proof/ssw/ssw.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


func Saig_ObjNext(ppNexts **Aig_Obj_t, pObj *Aig_Obj_t) *Aig_Obj_t { return ppNexts[pObj.Id];  }
func Saig_ObjSetNext(ppNexts **Aig_Obj_t, pObj *Aig_Obj_t, pNext *Aig_Obj_t) { ppNexts[pObj.Id] = pNext; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Saig_StrSimHash_s_SPrimes [128]int32 = [128]int32{
        1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
        1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
        2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
        2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
        3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
        3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
        4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
        4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
        5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
        6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
        6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
        7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
        8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
    };
func Saig_StrSimHash(pObj *Aig_Obj_t) uint32 {
    
    var pSims *uint32;
    var uHash uint32;
    var i int32;
    assert(SAIG_WORDS <= 128, "SAIG_WORDS <= 128");
    pSims = (unsigned *)pObj.pData;
    for  i = 0; i < SAIG_WORDS; i++  {
        uHash ^= pSims[i] * spStatic_Saig_StrSimHash_s_SPrimes[i & 0x7F];
    }
    return uHash;
}

/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimIsEqual(pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) int32 {
    var pSims0 *uint32  = (unsigned *)pObj0.pData;
    var pSims1 *uint32  = (unsigned *)pObj1.pData;
    var i int32;
    for  i = 0; i < SAIG_WORDS; i++  {
        if pSims0[i] != pSims1[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is zero.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimIsZero(pObj *Aig_Obj_t) int32 {
    var pSims *uint32  = (unsigned *)pObj.pData;
    var i int32;
    for  i = 0; i < SAIG_WORDS; i++  {
        if pSims[i] != 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimIsOne(pObj *Aig_Obj_t) int32 {
    var pSims *uint32  = (unsigned *)pObj.pData;
    var i int32;
    for  i = 0; i < SAIG_WORDS; i++  {
        if pSims[i] != ~0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Assigns random simulation info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimAssignRandom(pObj *Aig_Obj_t) {
    var pSims *uint32  = (unsigned *)pObj.pData;
    var i int32;
    for  i = 0; i < SAIG_WORDS; i++  {
        pSims[i] = Aig_ManRandom(0);
    }
}

/**Function*************************************************************

  Synopsis    [Assigns constant 0 simulation info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimAssignOne(pObj *Aig_Obj_t) {
    var pSims *uint32  = (unsigned *)pObj.pData;
    var i int32;
    for  i = 0; i < SAIG_WORDS; i++  {
        pSims[i] = ~0;
    }
}

/**Function*************************************************************

  Synopsis    [Assigns constant 0 simulation info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimAssignZeroInit(pObj *Aig_Obj_t) {
    var pSims *uint32  = (unsigned *)pObj.pData;
    pSims[0] = 0;
}

/**Function*************************************************************

  Synopsis    [Simulated one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimulateNode(pObj *Aig_Obj_t, i int32) {
    var pSims *uint32   = (unsigned *)pObj.pData;
    var pSims0 *uint32  = (unsigned *)Aig_ObjFanin0(pObj).pData;
    var pSims1 *uint32  = (unsigned *)Aig_ObjFanin1(pObj).pData;
    if Aig_ObjFaninC0(pObj) && Aig_ObjFaninC1(pObj) {
        pSims[i] = ~(pSims0[i] | pSims1[i]);
    } else if Aig_ObjFaninC0(pObj) && !Aig_ObjFaninC1(pObj) {
        pSims[i] = (~pSims0[i] & pSims1[i]);
    } else if !Aig_ObjFaninC0(pObj) && Aig_ObjFaninC1(pObj) {
        pSims[i] = (pSims0[i] & ~pSims1[i]);
    } else { // if ( !Aig_ObjFaninC0(pObj) && !Aig_ObjFaninC1(pObj) )
        pSims[i] = (pSims0[i] & pSims1[i]); }
}

/**Function*************************************************************

  Synopsis    [Saves output of one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimSaveOutput(pObj *Aig_Obj_t, i int32) {
    var pSims *uint32  = (unsigned *)pObj.pData;
    var pSims0 *uint32  = (unsigned *)Aig_ObjFanin0(pObj).pData;
    if Aig_ObjFaninC0(pObj) {
        pSims[i] = ~pSims0[i];
    } else {
        pSims[i] = pSims0[i]; }
}

/**Function*************************************************************

  Synopsis    [Transfers simulation output to another node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimTransfer(pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) {
    var pSims0 *uint32  = (unsigned *)pObj0.pData;
    var pSims1 *uint32  = (unsigned *)pObj1.pData;
    var i int32;
    for  i = 0; i < SAIG_WORDS; i++  {
        pSims1[i] = pSims0[i];
    }
}

/**Function*************************************************************

  Synopsis    [Transfers simulation output to another node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimTransferNext(pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t, i int32) {
    var pSims0 *uint32  = (unsigned *)pObj0.pData;
    var pSims1 *uint32  = (unsigned *)pObj1.pData;
    assert(i < SAIG_WORDS - 1, "i < SAIG_WORDS - 1");
    pSims1[i+1] = pSims0[i];
}

/**Function*************************************************************

  Synopsis    [Perform one round of simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimulateRound(p0 *Aig_Man_t, p1 *Aig_Man_t) {
    var pObj0, pObj1 *Aig_Obj_t;
    var f, i int32;
    // simulate the nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        if !Aig_ObjIsCi(pObj0) && !Aig_ObjIsNode(pObj0) {
           continue; }
        pObj1 = Aig_ObjRepr(p0, pObj0);
        if pObj1 == nil {
            continue; }
        assert(Aig_ObjRepr(p1, pObj1) == pObj0, "Aig_ObjRepr(p1, pObj1) == pObj0");
        Saig_StrSimAssignRandom( pObj0 );
        Saig_StrSimTransfer( pObj0, pObj1 );
    }
    }
    // simulate the timeframes
    for  f = 0; f < SAIG_WORDS; f++  {
        // simulate the first AIG
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil || !Aig_ObjIsNode(pObj0) {} else
            if Aig_ObjRepr(p0, pObj0) == nil {
                Saig_StrSimulateNode( pObj0, f ); }
        }
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(p0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCos, i+Saig_ManPoNum(p0))), 1); i++  {
            Saig_StrSimSaveOutput( pObj0, f );
        }
        if f < SAIG_WORDS - 1 {
            // inlining of macro Saig_ManForEachLiLo
            for  i = 0; (i < Saig_ManRegNum(p0)) && (((pObj0) = Saig_ManLi(p0, i)), 1)            && (((pObj1)=Saig_ManLo(p0, i)), 1); i++  {
                Saig_StrSimTransferNext( pObj0, pObj1, f );
            } }
        // simulate the second AIG
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p1.vObjs)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(p1.vObjs, i)), 1); i++  { if (pObj1) == nil || !Aig_ObjIsNode(pObj1) {} else
            if Aig_ObjRepr(p1, pObj1) == nil {
                Saig_StrSimulateNode( pObj1, f ); }
        }
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(p1)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(p1.vCos, i+Saig_ManPoNum(p1))), 1); i++  {
            Saig_StrSimSaveOutput( pObj1, f );
        }
        if f < SAIG_WORDS - 1 {
            // inlining of macro Saig_ManForEachLiLo
            for  i = 0; (i < Saig_ManRegNum(p1)) && (((pObj1) = Saig_ManLi(p1, i)), 1)            && (((pObj0)=Saig_ManLo(p1, i)), 1); i++  {
                Saig_StrSimTransferNext( pObj1, pObj0, f );
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Checks if the entry exists in the table.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimTableLookup(ppTable **Aig_Obj_t, ppNexts **Aig_Obj_t, nTableSize int32, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pEntry *Aig_Obj_t;
    var iEntry int32;
    // find the hash entry
    iEntry = Saig_StrSimHash( pObj ) % nTableSize;
    // check if there are nodes with this signatures
    for  pEntry = ppTable[iEntry]; (pEntry != nil); pEntry = Saig_ObjNext(ppNexts,pEntry)  {
        if Saig_StrSimIsEqual( pEntry, pObj ) != 0 {
            return pEntry; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Inserts the entry into the table.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimTableInsert(ppTable **Aig_Obj_t, ppNexts **Aig_Obj_t, nTableSize int32, pObj *Aig_Obj_t) {
    // find the hash entry
    var iEntry int32  = Saig_StrSimHash( pObj ) % nTableSize;
    // check if there are nodes with this signatures
    if ppTable[iEntry] == nil {
        ppTable[iEntry] = pObj;
    } else {
        Saig_ObjSetNext( ppNexts, pObj, Saig_ObjNext(ppNexts, ppTable[iEntry]) );
        Saig_ObjSetNext( ppNexts, ppTable[iEntry], pObj );
    }
}

/**Function*************************************************************

  Synopsis    [Perform one round of matching.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimDetectUnique(p0 *Aig_Man_t, p1 *Aig_Man_t) int32 {
    var ppTable, ppNexts, ppCands **Aig_Obj_t;
    var pObj, pEntry *Aig_Obj_t;
    var i, nTableSize, Counter int32;

    // allocate the hash table hashing simulation info into nodes
    nTableSize = Abc_PrimeCudd( Aig_ManObjNum(p0)/2 );
    // inlining of macro ABC_CALLOC
    ppTable = ((Aig_Obj_t * *) calloc((size_t)(nTableSize), sizeof(Aig_Obj_t *)));
    // inlining of macro ABC_CALLOC
    ppNexts = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p0)), sizeof(Aig_Obj_t *)));
    // inlining of macro ABC_CALLOC
    ppCands = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p0)), sizeof(Aig_Obj_t *)));

    // hash nodes of the first AIG
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsCi(pObj) && !Aig_ObjIsNode(pObj) {
            continue; }
        if Aig_ObjRepr(p0, pObj) {
            continue; }
        if (Saig_StrSimIsZero(pObj) != 0) || (Saig_StrSimIsOne(pObj) != 0) {
            continue; }
        // check if the entry exists
        pEntry = Saig_StrSimTableLookup( ppTable, ppNexts, nTableSize, pObj );
        if pEntry == nil { // insert
            Saig_StrSimTableInsert( ppTable, ppNexts, nTableSize, pObj );
        } else { // mark the entry as not unique
            pEntry.fMarkA = 1; }
    }
    }

    // hash nodes from the second AIG
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p1.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p1.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsCi(pObj) && !Aig_ObjIsNode(pObj) {
            continue; }
        if Aig_ObjRepr(p1, pObj) {
            continue; }
        if (Saig_StrSimIsZero(pObj) != 0) || (Saig_StrSimIsOne(pObj) != 0) {
            continue; }
        // check if the entry exists
        pEntry = Saig_StrSimTableLookup( ppTable, ppNexts, nTableSize, pObj );
        if pEntry == nil { // skip
            continue; }
        // if there is no candidate, label it
        if Saig_ObjNext( ppCands, pEntry ) == nil {
            Saig_ObjSetNext( ppCands, pEntry, pObj );
        } else { // mark the entry as not unique
            pEntry.fMarkA = 1; }
    }
    }

    // create representatives for the unique entries
    Counter = 0;
    for  i = 0; i < nTableSize; i++  {
        for  pEntry = ppTable[i]; (pEntry != nil); pEntry = Saig_ObjNext(ppNexts,pEntry)  {
            if !pEntry.fMarkA && (pObj = Saig_ObjNext( ppCands, pEntry )) {
//                assert( Aig_ObjIsNode(pEntry) == Aig_ObjIsNode(pObj) );
                if Aig_ObjType(pEntry) != Aig_ObjType(pObj) {
                    continue; }
                Aig_ObjSetRepr( p0, pEntry, pObj );
                Aig_ObjSetRepr( p1, pObj, pEntry );
                Counter++;
            }
        }
    }

    // cleanup
    Aig_ManCleanMarkA( p0 );
    ppTable = nil;
    ppNexts = nil;
    ppCands = nil;
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of matched flops.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimCountMatchedFlops(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        if Aig_ObjRepr(p, pObj) {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of matched nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimCountMatchedNodes(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjRepr(p, pObj) {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Performs structural matching of two AIGs using simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimPrepareAig(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManReprStart( p, Aig_ManObjNumMax(p) );
    // allocate simulation info
    p.pData2 = Vec_PtrAllocSimInfo( Aig_ManObjNumMax(p), SAIG_WORDS );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = Vec_PtrEntry( (Vec_Ptr_t *)p.pData2, i ); }
    }
    // set simulation info for constant1 and register outputs
    Saig_StrSimAssignOne( Aig_ManConst1(p) );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        Saig_StrSimAssignZeroInit( pObj );
    }
}

/**Function*************************************************************

  Synopsis    [Performs structural matching of two AIGs using simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimSetInitMatching(p0 *Aig_Man_t, p1 *Aig_Man_t) {
    var pObj0, pObj1 *Aig_Obj_t;
    var i int32;
    pObj0 = Aig_ManConst1( p0 );
    pObj1 = Aig_ManConst1( p1 );
    Aig_ObjSetRepr( p0, pObj0, pObj1 );
    Aig_ObjSetRepr( p1, pObj1, pObj0 );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCis, i)), 1); i++  {
        pObj1 = Aig_ManCi( p1, i );
        Aig_ObjSetRepr( p0, pObj0, pObj1 );
        Aig_ObjSetRepr( p1, pObj1, pObj0 );
    }
}

/**Function*************************************************************

  Synopsis    [Performs structural matching of two AIGs using simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimSetFinalMatching(p0 *Aig_Man_t, p1 *Aig_Man_t) {
    var pObj0, pObj1 *Aig_Obj_t;
    var pFanin00, pFanin01 *Aig_Obj_t;
    var pFanin10, pFanin11 *Aig_Obj_t;
    var i, CountAll, CountNot int32;
    Aig_ManIncrementTravId( p0 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        pObj1 = Aig_ObjRepr( p0, pObj0 );
        if pObj1 == nil {
            continue; }
        CountAll++;
        assert(pObj0 == Aig_ObjRepr( p1, pObj1 ), "pObj0 == Aig_ObjRepr( p1, pObj1 )");
        if Aig_ObjIsNode(pObj0) {
            assert(Aig_ObjIsNode(pObj1), "Aig_ObjIsNode(pObj1)");
            pFanin00 = Aig_ObjFanin0(pObj0);
            pFanin01 = Aig_ObjFanin1(pObj0);
            pFanin10 = Aig_ObjFanin0(pObj1);
            pFanin11 = Aig_ObjFanin1(pObj1);
            if Aig_ObjRepr(p0, pFanin00) != pFanin10 ||
                 Aig_ObjRepr(p0, pFanin01) != pFanin11 {
                Aig_ObjSetTravIdCurrent(p0, pObj0);
                CountNot++;
            }
        
        } else if Saig_ObjIsLo(p0, pObj0) {
            assert(Saig_ObjIsLo(p1, pObj1), "Saig_ObjIsLo(p1, pObj1)");
            pFanin00 = Aig_ObjFanin0( Saig_ObjLoToLi(p0, pObj0) );
            pFanin10 = Aig_ObjFanin0( Saig_ObjLoToLi(p1, pObj1) );
            if Aig_ObjRepr(p0, pFanin00) != pFanin10 {
                Aig_ObjSetTravIdCurrent(p0, pObj0);
                CountNot++;
            }
        }
    }
    }
    // remove irrelevant matches
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        pObj1 = Aig_ObjRepr( p0, pObj0 );
        if pObj1 == nil {
            continue; }
        assert(pObj0 == Aig_ObjRepr( p1, pObj1 ), "pObj0 == Aig_ObjRepr( p1, pObj1 )");
        if Aig_ObjIsTravIdCurrent( p0, pObj0 ) {
            Aig_ObjSetRepr( p0, pObj0, nil );
            Aig_ObjSetRepr( p1, pObj1, nil );
        }
    }
    }
    Abc_Print( 1, "Total matches = %6d.  Wrong matches = %6d.  Ratio = %5.2f %%\n",
        CountAll, CountNot, 100.0*CountNot/CountAll );
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimSetContiguousMatching_rec(p *Aig_Man_t, pObj *Aig_Obj_t) {
    var pFanout *Aig_Obj_t;
    var i int32;
    var iFanout int32 = -1;
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Saig_ObjIsPo( p, pObj ) {
        return; }
    if Saig_ObjIsLi( p, pObj ) {
        Saig_StrSimSetContiguousMatching_rec( p, Saig_ObjLiToLo(p, pObj) );
        return;
    }
    assert(Aig_ObjIsCi(pObj) || Aig_ObjIsNode(pObj), "Aig_ObjIsCi(pObj) || Aig_ObjIsNode(pObj)");
    if Aig_ObjRepr(p, pObj) == nil {
        return; }
    // go through the fanouts
    // inlining of macro Aig_ObjForEachFanout
    for  assert(p.pFanData, "p.pFanData"), i = 0; (i < int32((pObj).nRefs)) &&                         (((iFanout) = i? Aig_ObjFanoutNext(p, iFanout) : Aig_ObjFanout0Int(p, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(p, iFanout>>1)), 1); i++  {
        Saig_StrSimSetContiguousMatching_rec( p, pFanout );
    }
    // go through the fanins
    if !Aig_ObjIsCi( pObj ) {
        Saig_StrSimSetContiguousMatching_rec( p, Aig_ObjFanin0(pObj) );
        Saig_StrSimSetContiguousMatching_rec( p, Aig_ObjFanin1(pObj) );
    }
}

/**Function*************************************************************

  Synopsis    [Performs structural matching of two AIGs using simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimSetContiguousMatching(p0 *Aig_Man_t, p1 *Aig_Man_t) {
    var pObj0, pObj1 *Aig_Obj_t;
    var i, CountAll, CountNot int32;
    // mark nodes reachable through the PIs
    Aig_ManIncrementTravId( p0 );
    Aig_ObjSetTravIdCurrent( p0, Aig_ManConst1(p0) );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCis, i)), 1); i++  {
        Saig_StrSimSetContiguousMatching_rec( p0, pObj0 );
    }
    // remove irrelevant matches
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        pObj1 = Aig_ObjRepr( p0, pObj0 );
        if pObj1 == nil {
            continue; }
        CountAll++;
        assert(pObj0 == Aig_ObjRepr( p1, pObj1 ), "pObj0 == Aig_ObjRepr( p1, pObj1 )");
        if !Aig_ObjIsTravIdCurrent( p0, pObj0 ) {
            Aig_ObjSetRepr( p0, pObj0, nil );
            Aig_ObjSetRepr( p1, pObj1, nil );
            CountNot++;
        }
    }
    }
    Abc_Print( 1, "Total matches = %6d.  Wrong matches = %6d.  Ratio = %5.2f %%\n",
        CountAll, CountNot, 100.0*CountNot/CountAll );
}



/**Function*************************************************************

  Synopsis    [Establishes relationship between nodes using pairing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_StrSimMatchingExtendOne(p *Aig_Man_t, vNodes *Vec_Ptr_t) {
    var pNext, pObj *Aig_Obj_t;
    var i int32;
    var k int32;
    var iFan int32 = -1;
    Vec_PtrClear( vNodes );
    Aig_ManIncrementTravId( p );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
            continue; }
        if Aig_ObjRepr( p, pObj ) != nil {
            continue; }
        if Saig_ObjIsLo(p, pObj) {
            pNext = Saig_ObjLoToLi(p, pObj);
            pNext = Aig_ObjFanin0(pNext);
            if Aig_ObjRepr( p, pNext ) && !Aig_ObjIsTravIdCurrent(p, pNext) && !Aig_ObjIsConst1(pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
        }
        if Aig_ObjIsNode(pObj) {
            pNext = Aig_ObjFanin0(pObj);
            if Aig_ObjRepr( p, pNext )&& !Aig_ObjIsTravIdCurrent(p, pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
            pNext = Aig_ObjFanin1(pObj);
            if Aig_ObjRepr( p, pNext ) && !Aig_ObjIsTravIdCurrent(p, pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
        }
        // inlining of macro Aig_ObjForEachFanout
        for  assert(p.pFanData, "p.pFanData"), k = 0; (k < int32((pObj).nRefs)) &&                         (((iFan) = k? Aig_ObjFanoutNext(p, iFan) : Aig_ObjFanout0Int(p, pObj.Id)), 1) &&           (((pNext) = Aig_ManObj(p, iFan>>1)), 1); k++  {
            if Saig_ObjIsPo(p, pNext) {
                continue; }
            if Saig_ObjIsLi(p, pNext) {
                pNext = Saig_ObjLiToLo(p, pNext); }
            if Aig_ObjRepr( p, pNext ) && !Aig_ObjIsTravIdCurrent(p, pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Establishes relationship between nodes using pairing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_StrSimMatchingCountUnmached(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
            continue; }
        if Aig_ObjRepr( p, pObj ) != nil {
            continue; }
        Counter++;
    }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Establishes relationship between nodes using pairing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_StrSimMatchingExtend(p0 *Aig_Man_t, p1 *Aig_Man_t, nDist int32, fVerbose int32) {
    var vNodes0, vNodes1 *Vec_Ptr_t;
    var pNext0, pNext1 *Aig_Obj_t;
    var d, k int32;
    vNodes0 = Vec_PtrAlloc( 1000 );
    vNodes1 = Vec_PtrAlloc( 1000 );
    if fVerbose != 0 {
        var nUnmached int32  = Ssw_StrSimMatchingCountUnmached(p0);
        Abc_Print( 1, "Extending islands by %d steps:\n", nDist );
        Abc_Print( 1, "%2d : Total = %6d. Unmatched = %6d.  Ratio = %6.2f %%\n",
            0, Aig_ManCiNum(p0) + Aig_ManNodeNum(p0),
            nUnmached, 100.0 * nUnmached/(Aig_ManCiNum(p0) + Aig_ManNodeNum(p0)) );
    }
    for  d = 0; d < nDist; d++  {
        Ssw_StrSimMatchingExtendOne( p0, vNodes0 );
        Ssw_StrSimMatchingExtendOne( p1, vNodes1 );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes0)) && (((pNext0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes0, k)), 1); k++  {
            pNext1 = Aig_ObjRepr( p0, pNext0 );
            if pNext1 == nil {
                continue; }
            assert(pNext0 == Aig_ObjRepr( p1, pNext1 ), "pNext0 == Aig_ObjRepr( p1, pNext1 )");
            if Saig_ObjIsPi(p1, pNext1) {
                continue; }
            Aig_ObjSetRepr( p0, pNext0, nil );
            Aig_ObjSetRepr( p1, pNext1, nil );
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes1)) && (((pNext1) = (Aig_Obj_t *)Vec_PtrEntry(vNodes1, k)), 1); k++  {
            pNext0 = Aig_ObjRepr( p1, pNext1 );
            if pNext0 == nil {
                continue; }
            assert(pNext1 == Aig_ObjRepr( p0, pNext0 ), "pNext1 == Aig_ObjRepr( p0, pNext0 )");
            if Saig_ObjIsPi(p0, pNext0) {
                continue; }
            Aig_ObjSetRepr( p0, pNext0, nil );
            Aig_ObjSetRepr( p1, pNext1, nil );
        }
        if fVerbose != 0 {
            var nUnmached int32  = Ssw_StrSimMatchingCountUnmached(p0);
            Abc_Print( 1, "%2d : Total = %6d. Unmatched = %6d.  Ratio = %6.2f %%\n",
                d+1, Aig_ManCiNum(p0) + Aig_ManNodeNum(p0),
                nUnmached, 100.0 * nUnmached/(Aig_ManCiNum(p0) + Aig_ManNodeNum(p0)) );
        }
    }
    Vec_PtrFree( vNodes0 );
    Vec_PtrFree( vNodes1 );
}


/**Function*************************************************************

  Synopsis    [Performs structural matching of two AIGs using simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimPerformMatching(p0 *Aig_Man_t, p1 *Aig_Man_t, nDist int32, fVerbose int32, ppMiter **Aig_Man_t) *Vec_Int_t {
    extern Aig_Man_t * Saig_ManWindowExtractMiter( Aig_Man_t * p0, Aig_Man_t * p1 );

    var vPairs *Vec_Int_t;
    var pPart0, pPart1 *Aig_Man_t;
    var pObj0, pObj1 *Aig_Obj_t;
    var i, nMatches int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    Aig_ManRandom( 1 );
    // consider the case when a miter is given
    if p1 == nil {
        if fVerbose != 0 {
            Aig_ManPrintStats( p0 );
        }
        // demiter the miter
        if !Saig_ManDemiterSimpleDiff( p0, &pPart0, &pPart1 ) {
            Abc_Print( 1, "Demitering has failed.\n" );
            return nil;
        }
    
    } else {
        pPart0 = Aig_ManDupSimple( p0 );
        pPart1 = Aig_ManDupSimple( p1 );
    }
    if fVerbose != 0 {
        Aig_ManPrintStats( pPart0 );
        Aig_ManPrintStats( pPart1 );
    }
    // start simulation 
    Saig_StrSimPrepareAig( pPart0 );
    Saig_StrSimPrepareAig( pPart1 );
    Saig_StrSimSetInitMatching( pPart0, pPart1 );
    if fVerbose != 0 {
        Abc_Print( 1, "Allocated %6.2f MB to simulate the first AIG.\n",
            1.0 * Aig_ManObjNumMax(pPart0) * SAIG_WORDS * sizeof(unsigned) / (1<<20) );
        Abc_Print( 1, "Allocated %6.2f MB to simulate the second AIG.\n",
            1.0 * Aig_ManObjNumMax(pPart1) * SAIG_WORDS * sizeof(unsigned) / (1<<20) );
    }
    // iterate matching
    nMatches = 1;
    for  i = 0; nMatches > 0; i++  {
        clk = Abc_Clock();
        Saig_StrSimulateRound( pPart0, pPart1 );
        nMatches = Saig_StrSimDetectUnique( pPart0, pPart1 );
        if fVerbose != 0 {
            var nFlops int32  = Saig_StrSimCountMatchedFlops(pPart0);
            var nNodes int32  = Saig_StrSimCountMatchedNodes(pPart0);
            Abc_Print( 1, "%3d : Match =%6d.  FF =%6d. (%6.2f %%)  Node =%6d. (%6.2f %%)  ",
                i, nMatches,
                nFlops, 100.0*nFlops/Aig_ManRegNum(pPart0),
                nNodes, 100.0*nNodes/Aig_ManNodeNum(pPart0) );
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
        if i == 20 {
            break; }
    }
    // cleanup
    Vec_PtrFree( (Vec_Ptr_t *)pPart0.pData2 ); pPart0.pData2 = nil;
    Vec_PtrFree( (Vec_Ptr_t *)pPart1.pData2 ); pPart1.pData2 = nil;
    // extend the islands
    Aig_ManFanoutStart( pPart0 );
    Aig_ManFanoutStart( pPart1 );
    if nDist != 0 {
        Ssw_StrSimMatchingExtend( pPart0, pPart1, nDist, fVerbose ); }
    Saig_StrSimSetFinalMatching( pPart0, pPart1 );
//    Saig_StrSimSetContiguousMatching( pPart0, pPart1 );
    // copy the results into array
    vPairs = Vec_IntAlloc( 2*Aig_ManObjNumMax(pPart0) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPart0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(pPart0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        pObj1 = Aig_ObjRepr(pPart0, pObj0);
        if pObj1 == nil {
            continue; }
        assert(pObj0 == Aig_ObjRepr(pPart1, pObj1), "pObj0 == Aig_ObjRepr(pPart1, pObj1)");
        Vec_IntPush( vPairs, pObj0.Id );
        Vec_IntPush( vPairs, pObj1.Id );
    }
    }
    // this procedure adds matching of PO and LI
    if ppMiter {
        *ppMiter = Saig_ManWindowExtractMiter( pPart0, pPart1 ); }
    Aig_ManFanoutStop( pPart0 );
    Aig_ManFanoutStop( pPart1 );
    Aig_ManStop( pPart0 );
    Aig_ManStop( pPart1 );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    return vPairs;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
