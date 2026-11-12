package cnf

import "fmt"

/**CFile****************************************************************

  FileName    [cnfFast.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfFast.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Detects multi-input gate rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CollectLeaves_rec(pRoot *Aig_Obj_t, pObj *Aig_Obj_t, vSuper *Vec_Ptr_t, fStopCompl int32) {
    if pRoot != pObj && (pObj.fMarkA || ((fStopCompl != 0) && Aig_IsComplement(pObj))) {
        Vec_PtrPushUnique( vSuper, fStopCompl ? pObj : Aig_Regular(pObj) );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    if fStopCompl != 0 {
        Cnf_CollectLeaves_rec( pRoot, Aig_ObjChild0(pObj), vSuper, 1 );
        Cnf_CollectLeaves_rec( pRoot, Aig_ObjChild1(pObj), vSuper, 1 );
    
    } else {
        Cnf_CollectLeaves_rec( pRoot, Aig_ObjFanin0(pObj), vSuper, 0 );
        Cnf_CollectLeaves_rec( pRoot, Aig_ObjFanin1(pObj), vSuper, 0 );
    }
}

/**Function*************************************************************

  Synopsis    [Detects multi-input gate rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CollectLeaves(pRoot *Aig_Obj_t, vSuper *Vec_Ptr_t, fStopCompl int32) {
    assert(!Aig_IsComplement(pRoot), "!Aig_IsComplement(pRoot)");
    Vec_PtrClear( vSuper );
    Cnf_CollectLeaves_rec( pRoot, pRoot, vSuper, fStopCompl );
}

/**Function*************************************************************

  Synopsis    [Collects nodes inside the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CollectVolume_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vNodes *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent( p, pObj ) {
        return; }
    Aig_ObjSetTravIdCurrent( p, pObj );
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Cnf_CollectVolume_rec( p, Aig_ObjFanin0(pObj), vNodes );
    Cnf_CollectVolume_rec( p, Aig_ObjFanin1(pObj), vNodes );
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Collects nodes inside the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CollectVolume(p *Aig_Man_t, pRoot *Aig_Obj_t, vLeaves *Vec_Ptr_t, vNodes *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManIncrementTravId( p );
    Vec_PtrForEachEntry( Aig_Obj_t *, vLeaves, pObj, i )
        Aig_ObjSetTravIdCurrent( p, pObj );
    Vec_PtrClear( vNodes );
    Cnf_CollectVolume_rec( p, pRoot, vNodes );
}

/**Function*************************************************************

  Synopsis    [Derive truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutDeriveTruth(p *Aig_Man_t, vLeaves *Vec_Ptr_t, vNodes *Vec_Ptr_t) word {
    static word Truth6[6] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00),
        ABC_CONST(0xFFFF0000FFFF0000),
        ABC_CONST(0xFFFFFFFF00000000)
    };
    static word C[2] = { 0, ~(word)0 };
    static word S[256];
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    assert(Vec_PtrSize(vLeaves) <= 6 && Vec_PtrSize(vNodes) > 0, "Vec_PtrSize(vLeaves) <= 6 && Vec_PtrSize(vNodes) > 0");
    assert(Vec_PtrSize(vLeaves) + Vec_PtrSize(vNodes) <= 256, "Vec_PtrSize(vLeaves) + Vec_PtrSize(vNodes) <= 256");
    Vec_PtrForEachEntry( Aig_Obj_t *, vLeaves, pObj, i )
    {
        pObj.iData    = i;
        S[pObj.iData] = Truth6[i];
    }
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        pObj.iData    = Vec_PtrSize(vLeaves) + i;
        S[pObj.iData] = (S[Aig_ObjFanin0(pObj).iData] ^ C[Aig_ObjFaninC0(pObj)]) & 
                         (S[Aig_ObjFanin1(pObj).iData] ^ C[Aig_ObjFaninC1(pObj)]);
    }
    return S[pObj.iData];
}


/**Function*************************************************************

  Synopsis    [Collects nodes inside the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ObjGetLit(vMap *Vec_Int_t, pObj *Aig_Obj_t, fCompl int32) int32 {
    var iSatVar int32;
    if vMap != nil {
        iSatVar = Vec_IntEntry(vMap, Aig_ObjId(pObj));
    } else {
        iSatVar = Aig_ObjId(pObj);
    }
    assert(iSatVar > 0, "iSatVar > 0");
    return iSatVar + iSatVar + fCompl;
}

/**Function*************************************************************

  Synopsis    [Collects nodes inside the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ComputeClauses(p *Aig_Man_t, pRoot *Aig_Obj_t, vLeaves *Vec_Ptr_t, vNodes *Vec_Ptr_t, vMap *Vec_Int_t, vCover *Vec_Int_t, vClauses *Vec_Int_t) {
    var pLeaf *Aig_Obj_t;
    var c, k, Cube, OutLit, RetValue int32;
    var Truth word;
    assert(pRoot.fMarkA, "pRoot.fMarkA");

    Vec_IntClear( vClauses );

    OutLit = Cnf_ObjGetLit( vMap, pRoot, 0 );
    // detect cone
    Cnf_CollectLeaves( pRoot, vLeaves, 0 );
    Cnf_CollectVolume( p, pRoot, vLeaves, vNodes );
    assert(pRoot == Vec_PtrEntryLast(vNodes), "pRoot == Vec_PtrEntryLast(vNodes)");
    // check if this is an AND-gate
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pLeaf, k )
    {
        if Aig_ObjFaninC0(pLeaf) && !Aig_ObjFanin0(pLeaf).fMarkA {
            break; }
        if Aig_ObjFaninC1(pLeaf) && !Aig_ObjFanin1(pLeaf).fMarkA {
            break; }
    }
    if k == Vec_PtrSize(vNodes) {
        Cnf_CollectLeaves( pRoot, vLeaves, 1 );
        // write big clause
        Vec_IntPush( vClauses, 0 );
        Vec_IntPush( vClauses, OutLit );
        Vec_PtrForEachEntry( Aig_Obj_t *, vLeaves, pLeaf, k )
            Vec_IntPush( vClauses, Cnf_ObjGetLit(vMap, Aig_Regular(pLeaf), !Aig_IsComplement(pLeaf)) );
        // write small clauses
        Vec_PtrForEachEntry( Aig_Obj_t *, vLeaves, pLeaf, k )
        {
            Vec_IntPush( vClauses, 0 );
            Vec_IntPush( vClauses, OutLit ^ 1 );
            Vec_IntPush( vClauses, Cnf_ObjGetLit(vMap, Aig_Regular(pLeaf), Aig_IsComplement(pLeaf)) );
        }
        return;
    }
    if Vec_PtrSize(vLeaves) > 6 {
        fmt.Printf("FastCnfGeneration:  Internal error!!!\n"); }
    assert(Vec_PtrSize(vLeaves) <= 6, "Vec_PtrSize(vLeaves) <= 6");

    Truth = Cnf_CutDeriveTruth( p, vLeaves, vNodes );
    if Truth == 0 || Truth == ~(word)0 {
        Vec_IntPush( vClauses, 0 );
        Vec_IntPush( vClauses, (Truth == 0) ? (OutLit ^ 1) : OutLit );
        return;
    }

    RetValue = Kit_TruthIsop( (unsigned *)&Truth, Vec_PtrSize(vLeaves), vCover, 0 );
    assert(RetValue >= 0, "RetValue >= 0");

    Vec_IntForEachEntry( vCover, Cube, c )
    {
        Vec_IntPush( vClauses, 0 );
        Vec_IntPush( vClauses, OutLit );
        for  k = 0; k < Vec_PtrSize(vLeaves); k++, Cube >>= 2  {
            if (Cube & 3) == 0 {
                continue; }
            assert((Cube & 3) != 3, "(Cube & 3) != 3");
            Vec_IntPush( vClauses, Cnf_ObjGetLit(vMap, (Aig_Obj_t *)Vec_PtrEntry(vLeaves,k), (Cube&3)!=1) );
        }
    }

    Truth = ~Truth;

    RetValue = Kit_TruthIsop( (unsigned *)&Truth, Vec_PtrSize(vLeaves), vCover, 0 );
    assert(RetValue >= 0, "RetValue >= 0");
    Vec_IntForEachEntry( vCover, Cube, c )
    {
        Vec_IntPush( vClauses, 0 );
        Vec_IntPush( vClauses, OutLit ^ 1 );
        for  k = 0; k < Vec_PtrSize(vLeaves); k++, Cube >>= 2  {
            if (Cube & 3) == 0 {
                continue; }
            assert((Cube & 3) != 3, "(Cube & 3) != 3");
            Vec_IntPush( vClauses, Cnf_ObjGetLit(vMap, (Aig_Obj_t *)Vec_PtrEntry(vLeaves,k), (Cube&3)!=1) );
        }
    }
}



/**Function*************************************************************

  Synopsis    [Marks AIG for CNF computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveFastMark(p *Aig_Man_t) {
    var vSupps *Vec_Int_t;
    var vLeaves, vNodes *Vec_Ptr_t;
    var pObj, pTemp, pObjC, pObj0, pObj1 *Aig_Obj_t;
    var i, k, nFans, Counter int32;

    vLeaves = Vec_PtrAlloc( 100 );
    vNodes  = Vec_PtrAlloc( 100 );
    vSupps  = Vec_IntStart( Aig_ManObjNumMax(p) );

    // mark CIs
    Aig_ManForEachCi( p, pObj, i )
        pObj.fMarkA = 1;

    // mark CO drivers
    Aig_ManForEachCo( p, pObj, i )
        Aig_ObjFanin0(pObj).fMarkA = 1;

    // mark MUX/XOR nodes
    Aig_ManForEachNode( p, pObj, i )
    {
        assert(!pObj.fMarkB, "!pObj.fMarkB");
        if !Aig_ObjIsMuxType(pObj) {
            continue; }
        pObj0 = Aig_ObjFanin0(pObj);
        if pObj0.fMarkB || Aig_ObjRefs(pObj0) > 1 {
            continue; }
        pObj1 = Aig_ObjFanin1(pObj);
        if pObj1.fMarkB || Aig_ObjRefs(pObj1) > 1 {
            continue; }
        // mark nodes
        pObj.fMarkB = 1;
        pObj0.fMarkB = 1;
        pObj1.fMarkB = 1;
        // mark inputs and outputs
        pObj.fMarkA = 1;
        Aig_ObjFanin0(pObj0).fMarkA = 1;
        Aig_ObjFanin1(pObj0).fMarkA = 1;
        Aig_ObjFanin0(pObj1).fMarkA = 1;
        Aig_ObjFanin1(pObj1).fMarkA = 1;
    }

    // mark nodes with multiple fanouts and pointed to by complemented edges
    Aig_ManForEachNode( p, pObj, i )
    {
        // mark nodes with many fanouts
        if Aig_ObjRefs(pObj) > 1 {
            pObj.fMarkA = 1; }
        // mark nodes pointed to by a complemented edge
        if Aig_ObjFaninC0(pObj) && !Aig_ObjFanin0(pObj).fMarkB {
            Aig_ObjFanin0(pObj).fMarkA = 1; }
        if Aig_ObjFaninC1(pObj) && !Aig_ObjFanin1(pObj).fMarkB {
            Aig_ObjFanin1(pObj).fMarkA = 1; }
    }

    // compute supergate size for internal marked nodes
    Aig_ManForEachNode( p, pObj, i )
    {
        if !pObj.fMarkA {
            continue; }
        if pObj.fMarkB {
            if !Aig_ObjIsMuxType(pObj) {
                continue; }
            pObjC = Aig_ObjRecognizeMux( pObj, &pObj1, &pObj0 );
            pObj0 = Aig_Regular(pObj0);
            pObj1 = Aig_Regular(pObj1);
            assert(pObj0.fMarkA, "pObj0.fMarkA");
            assert(pObj1.fMarkA, "pObj1.fMarkA");
//            if ( pObj0 == pObj1 )
//                continue;
            nFans = 1 + (pObj0 == pObj1);
            if !pObj0.fMarkB && !Aig_ObjIsCi(pObj0) && Aig_ObjRefs(pObj0) == nFans && Vec_IntEntry(vSupps, Aig_ObjId(pObj0)) < 3 {
                pObj0.fMarkA = 0;
                continue;
            }
            if !pObj1.fMarkB && !Aig_ObjIsCi(pObj1) && Aig_ObjRefs(pObj1) == nFans && Vec_IntEntry(vSupps, Aig_ObjId(pObj1)) < 3 {
                pObj1.fMarkA = 0;
                continue;
            }
            continue;
        }

        Cnf_CollectLeaves( pObj, vLeaves, 1 );
        Vec_IntWriteEntry( vSupps, Aig_ObjId(pObj), Vec_PtrSize(vLeaves) );
        if Vec_PtrSize(vLeaves) >= 6 {
            continue; }
        Vec_PtrForEachEntry( Aig_Obj_t *, vLeaves, pTemp, k )
        {
            pTemp = Aig_Regular(pTemp);
            assert(pTemp.fMarkA, "pTemp.fMarkA");
            if pTemp.fMarkB || Aig_ObjIsCi(pTemp) || Aig_ObjRefs(pTemp) > 1 {
                continue; }
            assert(Vec_IntEntry(vSupps, Aig_ObjId(pTemp)) > 0, "Vec_IntEntry(vSupps, Aig_ObjId(pTemp)) > 0");
            if Vec_PtrSize(vLeaves) - 1 + Vec_IntEntry(vSupps, Aig_ObjId(pTemp)) > 6 {
                continue; }
            pTemp.fMarkA = 0;
            Vec_IntWriteEntry( vSupps, Aig_ObjId(pObj), 6 );
//printf( "%d %d   ", Vec_PtrSize(vLeaves), Vec_IntEntry(vSupps, Aig_ObjId(pTemp)) );
            break;
        }
    }
    Aig_ManCleanMarkB( p );

    // check CO drivers
    Counter = 0;
    Aig_ManForEachCo( p, pObj, i )
        Counter += !Aig_ObjFanin0(pObj).fMarkA;
    if Counter != 0 {
    fmt.Printf("PO-driver rule is violated %d times.\n", Counter); }

    // check that the AND-gates are fine
    Counter = 0;
    Aig_ManForEachNode( p, pObj, i )
    {
        assert(pObj.fMarkB == 0, "pObj.fMarkB == 0");
        if !pObj.fMarkA {
            continue; }
        Cnf_CollectLeaves( pObj, vLeaves, 0 );
        if Vec_PtrSize(vLeaves) <= 6 {
            continue; }
        Cnf_CollectVolume( p, pObj, vLeaves, vNodes );
        Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pTemp, k )
        {
            if Aig_ObjFaninC0(pTemp) && !Aig_ObjFanin0(pTemp).fMarkA {
                Counter++; }
            if Aig_ObjFaninC1(pTemp) && !Aig_ObjFanin1(pTemp).fMarkA {
                Counter++; }
        }
    }
    if Counter != 0 {
    fmt.Printf("AND-gate rule is violated %d times.\n", Counter); }

    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vNodes );
    Vec_IntFree( vSupps );
}


/**Function*************************************************************

  Synopsis    [Counts the number of clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutCountClauses(p *Aig_Man_t, vLeaves *Vec_Ptr_t, vNodes *Vec_Ptr_t, vCover *Vec_Int_t) int32 {
    var Truth word;
    var pObj *Aig_Obj_t;
    var i, RetValue, nSize int32;
    if Vec_PtrSize(vLeaves) > 6 {
        // make sure this is an AND gate
        Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        {
            if Aig_ObjFaninC0(pObj) && !Aig_ObjFanin0(pObj).fMarkA {
                fmt.Printf("Unusual 1!\n"); }
            if Aig_ObjFaninC1(pObj) && !Aig_ObjFanin1(pObj).fMarkA {
                fmt.Printf("Unusual 2!\n"); }
            continue;

            assert(!Aig_ObjFaninC0(pObj) || Aig_ObjFanin0(pObj).fMarkA, "!Aig_ObjFaninC0(pObj) || Aig_ObjFanin0(pObj).fMarkA");
            assert(!Aig_ObjFaninC1(pObj) || Aig_ObjFanin1(pObj).fMarkA, "!Aig_ObjFaninC1(pObj) || Aig_ObjFanin1(pObj).fMarkA");
        }
        return Vec_PtrSize(vLeaves) + 1;
    }
    Truth = Cnf_CutDeriveTruth( p, vLeaves, vNodes );

    RetValue = Kit_TruthIsop( (unsigned *)&Truth, Vec_PtrSize(vLeaves), vCover, 0 );
    assert(RetValue >= 0, "RetValue >= 0");
    nSize += Vec_IntSize(vCover);

    Truth = ~Truth;

    RetValue = Kit_TruthIsop( (unsigned *)&Truth, Vec_PtrSize(vLeaves), vCover, 0 );
    assert(RetValue >= 0, "RetValue >= 0");
    nSize += Vec_IntSize(vCover);
    return nSize;
}

/**Function*************************************************************

  Synopsis    [Counts the size of the CNF, assuming marks are set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CountCnfSize(p *Aig_Man_t) int32 {
    var vLeaves, vNodes *Vec_Ptr_t;
    var vCover *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var nVars, nClauses int32;
    var i, nSize int32;

    vLeaves = Vec_PtrAlloc( 100 );
    vNodes  = Vec_PtrAlloc( 100 );
    vCover  = Vec_IntAlloc( 1 << 16 );

    Aig_ManForEachObj( p, pObj, i )
        nVars += pObj.fMarkA;

    Aig_ManForEachNode( p, pObj, i )
    {
        if !pObj.fMarkA {
            continue; }
        Cnf_CollectLeaves( pObj, vLeaves, 0 );
        Cnf_CollectVolume( p, pObj, vLeaves, vNodes );
        assert(pObj == Vec_PtrEntryLast(vNodes), "pObj == Vec_PtrEntryLast(vNodes)");

        nSize = Cnf_CutCountClauses( p, vLeaves, vNodes, vCover );
//        printf( "%d(%d) ", Vec_PtrSize(vLeaves), nSize );

        nClauses += nSize;
    }
//    printf( "\n" );
    fmt.Printf("Vars = %d  Clauses = %d\n", nVars, nClauses);

    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vNodes );
    Vec_IntFree( vCover );
    return nClauses;
}

/**Function*************************************************************

  Synopsis    [Derives CNF from the marked AIG.]

  Description [Assumes that marking is such that when we traverse from each
  marked node, the logic cone has 6 inputs or less, or it is a multi-input AND.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveFastClauses(p *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t;
    var vLits, vClas, vMap, vTemp *Vec_Int_t;
    var vLeaves, vNodes *Vec_Ptr_t;
    var vCover *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, k, nVars, Entry, OutLit, DriLit int32;

    vLits = Vec_IntAlloc( 1 << 16 );
    vClas = Vec_IntAlloc( 1 << 12 );
    vMap  = Vec_IntStartFull( Aig_ManObjNumMax(p) );

    // assign variables for the outputs
    nVars = 1;
    if nOutputs != 0 {
        if Aig_ManRegNum(p) == 0 {
            assert(nOutputs == Aig_ManCoNum(p), "nOutputs == Aig_ManCoNum(p)");
            Aig_ManForEachCo( p, pObj, i )
                Vec_IntWriteEntry( vMap, Aig_ObjId(pObj), nVars++ );
        
        } else {
            assert(nOutputs == Aig_ManRegNum(p), "nOutputs == Aig_ManRegNum(p)");
            Aig_ManForEachLiSeq( p, pObj, i )
                Vec_IntWriteEntry( vMap, Aig_ObjId(pObj), nVars++ );
        }
    }
    // assign variables to the internal nodes
    Aig_ManForEachNodeReverse( p, pObj, i )
        if ( pObj.fMarkA )
            Vec_IntWriteEntry( vMap, Aig_ObjId(pObj), nVars++ );
    // assign variables to the PIs and constant node
    Aig_ManForEachCi( p, pObj, i )
        Vec_IntWriteEntry( vMap, Aig_ObjId(pObj), nVars++ );
    Vec_IntWriteEntry( vMap, Aig_ObjId(Aig_ManConst1(p)), nVars++ );

    // create clauses
    vLeaves = Vec_PtrAlloc( 100 );
    vNodes  = Vec_PtrAlloc( 100 );
    vCover  = Vec_IntAlloc( 1 << 16 );
    vTemp   = Vec_IntAlloc( 100 );
    Aig_ManForEachNodeReverse( p, pObj, i )
    {
        if !pObj.fMarkA {
            continue; }
        Cnf_ComputeClauses( p, pObj, vLeaves, vNodes, vMap, vCover, vTemp );
        Vec_IntForEachEntry( vTemp, Entry, k )
        {
            if Entry == 0 {
                Vec_IntPush( vClas, Vec_IntSize(vLits) );
            } else {
                Vec_IntPush( vLits, Entry ); }
        }       
    }
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vNodes );
    Vec_IntFree( vCover );
    Vec_IntFree( vTemp );

    // create clauses for the outputs
    Aig_ManForEachCo( p, pObj, i )
    {
        DriLit = Cnf_ObjGetLit( vMap, Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj) );
        if i < Aig_ManCoNum(p) - nOutputs {
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            Vec_IntPush( vLits, DriLit );
        
        } else {
            OutLit = Cnf_ObjGetLit( vMap, pObj, 0 );
            // first clause
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            Vec_IntPush( vLits, OutLit );
            Vec_IntPush( vLits, DriLit ^ 1 );
            // second clause
            Vec_IntPush( vClas, Vec_IntSize(vLits) );
            Vec_IntPush( vLits, OutLit ^ 1 );
            Vec_IntPush( vLits, DriLit );
        }
    }
 
    // write the constant literal
    OutLit = Cnf_ObjGetLit( vMap, Aig_ManConst1(p), 0 );
    Vec_IntPush( vClas, Vec_IntSize(vLits) );
    Vec_IntPush( vLits, OutLit );

    // create structure
    pCnf = new(Cnf_Dat_t);
    pCnf.pMan = p;
    pCnf.nVars = nVars;
    pCnf.nLiterals = Vec_IntSize( vLits );
    pCnf.nClauses  = Vec_IntSize( vClas );
    pCnf.pClauses  = ABC_ALLOC( int *, pCnf.nClauses + 1 );
    pCnf.pClauses[0] = Vec_IntReleaseArray( vLits );
    Vec_IntForEachEntry( vClas, Entry, i )
        pCnf.pClauses[i] = pCnf.pClauses[0] + Entry;
    pCnf.pClauses[pCnf.nClauses] = pCnf.pClauses[0] + pCnf.nLiterals;
    pCnf.pVarNums  = Vec_IntReleaseArray( vMap );

    // cleanup
    Vec_IntFree( vLits );
    Vec_IntFree( vClas );
    Vec_IntFree( vMap );
    return pCnf;
}

/**Function*************************************************************

  Synopsis    [Fast CNF computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveFast(p *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t  = nil;
    var clk abctime;//, clkTotal = Abc_Clock();
//    printf( "\n" );
    Aig_ManCleanMarkAB( p );
    // create initial marking
    clk = Abc_Clock();
    Cnf_DeriveFastMark( p );
//    Abc_PrintTime( 1, "Marking", Abc_Clock() - clk );
    // compute CNF size
    clk = Abc_Clock();
    pCnf = Cnf_DeriveFastClauses( p, nOutputs );
//    Abc_PrintTime( 1, "Clauses", Abc_Clock() - clk );
    // derive the resulting CNF
    Aig_ManCleanMarkA( p );
//    Abc_PrintTime( 1, "TOTAL  ", Abc_Clock() - clkTotal );

//    printf( "CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d.   \n", pCnf->nVars, pCnf->nClauses, pCnf->nLiterals );

//    Cnf_DataFree( pCnf );
//    pCnf = NULL;
    return pCnf;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

