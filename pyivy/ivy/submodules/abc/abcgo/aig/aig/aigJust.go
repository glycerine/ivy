package aig

import "fmt"

/**CFile****************************************************************
 
  FileName    [aigJust.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Justification of node values.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigJust.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const AIG_VAL0 =  1
const AIG_VAL1 =  2
const AIG_VALX =  3

// storing ternary values
func Aig_ObjSetTerValue(pNode *Aig_Obj_t, Value int32) int32 {
    assert((Value != 0), "Value");
    pNode.fMarkA = (Value & 1);
    pNode.fMarkB = ((Value >> 1) & 1);
    return Value;
}
func Aig_ObjGetTerValue(pNode *Aig_Obj_t) int32 {
    return (pNode.fMarkB << 1) | pNode.fMarkA;
}

// working with ternary values
func Aig_ObjNotTerValue(Value int32) int32 {
    if Value == AIG_VAL1 {
        return AIG_VAL0; }
    if Value == AIG_VAL0 {
        return AIG_VAL1; }
    return AIG_VALX;
}
func Aig_ObjAndTerValue(Value0 int32, Value1 int32) int32 {
    if Value0 == AIG_VAL0 || Value1 == AIG_VAL0 {
        return AIG_VAL0; }
    if Value0 == AIG_VAL1 && Value1 == AIG_VAL1 {
        return AIG_VAL1; }
    return AIG_VALX;
}
func Aig_ObjNotCondTerValue(Value int32, Cond int32) int32 {
    if Cond != 0 {
        return Aig_ObjNotTerValue( Value );
    } else {
        return Value;
    }
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns value (0 or 1) or X if unassigned.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjSatValue(pAig *Aig_Man_t, pNode *Aig_Obj_t, fCompl int32) int32 {
    if Aig_ObjIsTravIdCurrent(pAig, pNode) {
        if pNode.fMarkA ^ fCompl {
            return AIG_VAL1;
        } else {
            return AIG_VAL0;
        } }
    return AIG_VALX;
}

/**Function*************************************************************

  Synopsis    [Recursively searched for a satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NtkFindSatAssign_rec(pAig *Aig_Man_t, pNode *Aig_Obj_t, Value int32, vSuppLits *Vec_Int_t, Heur int32) int32 {
    var Value0, Value1 int32;
//    ++Heur;
    if Aig_ObjIsConst1(pNode) {
        return 1; }
    if Aig_ObjIsTravIdCurrent(pAig, pNode) {
        return (int32(pNode.fMarkA) == Value); }
    Aig_ObjSetTravIdCurrent(pAig, pNode);
    pNode.fMarkA = Value;
    if Aig_ObjIsCi(pNode) {
//        if ( Aig_ObjId(pNode) % 1000 == 0 )
//            Value ^= 1;
        if vSuppLits != nil {
            Vec_IntPush( vSuppLits, Abc_Var2Lit( Aig_ObjCioId(pNode), !Value ) ); }
        return 1;
    }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    // propagation
    if Value != 0 {
        if !(Aig_NtkFindSatAssign_rec(pAig, Aig_ObjFanin0(pNode), !Aig_ObjFaninC0(pNode), vSuppLits, Heur) != 0) {
            return 0; }
        return Aig_NtkFindSatAssign_rec(pAig, Aig_ObjFanin1(pNode), !Aig_ObjFaninC1(pNode), vSuppLits, Heur);
    }
    // justification
    Value0 = Aig_ObjSatValue( pAig, Aig_ObjFanin0(pNode), Aig_ObjFaninC0(pNode) );
    if Value0 == AIG_VAL0 {
        return 1; }
    Value1 = Aig_ObjSatValue( pAig, Aig_ObjFanin1(pNode), Aig_ObjFaninC1(pNode) );
    if Value1 == AIG_VAL0 {
        return 1; }
    if Value0 == AIG_VAL1 && Value1 == AIG_VAL1 {
        return 0; }
    if Value0 == AIG_VAL1 {
        return Aig_NtkFindSatAssign_rec(pAig, Aig_ObjFanin1(pNode), Aig_ObjFaninC1(pNode), vSuppLits, Heur); }
    if Value1 == AIG_VAL1 {
        return Aig_NtkFindSatAssign_rec(pAig, Aig_ObjFanin0(pNode), Aig_ObjFaninC0(pNode), vSuppLits, Heur); }
    assert(Value0 == AIG_VALX && Value1 == AIG_VALX, "Value0 == AIG_VALX && Value1 == AIG_VALX");
    // decision making
//    if ( rand() % 10 == Heur )
//    if ( Aig_ObjId(pNode) % 8 == Heur )
    if Heur++; Heur % 8 == 0 {
        return Aig_NtkFindSatAssign_rec(pAig, Aig_ObjFanin1(pNode), Aig_ObjFaninC1(pNode), vSuppLits, Heur);
    } else {
        return Aig_NtkFindSatAssign_rec(pAig, Aig_ObjFanin0(pNode), Aig_ObjFaninC0(pNode), vSuppLits, Heur); }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if SAT assignment is found; 0 otherwise.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjFindSatAssign(pAig *Aig_Man_t, pNode *Aig_Obj_t, Value int32, vSuppLits *Vec_Int_t) int32 {
    var i int32;
    if Aig_ObjIsCo(pNode) {
        return Aig_ObjFindSatAssign( pAig, Aig_ObjFanin0(pNode), Value ^ Aig_ObjFaninC0(pNode), vSuppLits ); }
    for  i = 0; i < 8; i++  {
        Vec_IntClear( vSuppLits );
        Aig_ManIncrementTravId( pAig );
        if Aig_NtkFindSatAssign_rec( pAig, pNode, Value, vSuppLits, i ) != 0 {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjTerSimulate_rec(pAig *Aig_Man_t, pNode *Aig_Obj_t) int32 {
    var Value0, Value1 int32;
    if Aig_ObjIsConst1(pNode) {
        return AIG_VAL1; }
    if Aig_ObjIsTravIdCurrent(pAig, pNode) {
        return Aig_ObjGetTerValue( pNode ); }
    Aig_ObjSetTravIdCurrent( pAig, pNode );
    if Aig_ObjIsCi(pNode) {
        return Aig_ObjSetTerValue( pNode, AIG_VALX ); }
    Value0 = Aig_ObjNotCondTerValue( Aig_ObjTerSimulate_rec(pAig, Aig_ObjFanin0(pNode)), Aig_ObjFaninC0(pNode) );
    if Aig_ObjIsCo(pNode) || Value0 == AIG_VAL0 {
        return Aig_ObjSetTerValue( pNode, Value0 ); }
    assert(Aig_ObjIsNode(pNode), "Aig_ObjIsNode(pNode)");
    Value1 = Aig_ObjNotCondTerValue( Aig_ObjTerSimulate_rec(pAig, Aig_ObjFanin1(pNode)), Aig_ObjFaninC1(pNode) );
    return Aig_ObjSetTerValue( pNode, Aig_ObjAndTerValue(Value0, Value1) );
}

/**Function*************************************************************

  Synopsis    [Returns value of the object under input assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjTerSimulate(pAig *Aig_Man_t, pNode *Aig_Obj_t, vSuppLits *Vec_Int_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Entry int32;
    Aig_ManIncrementTravId( pAig );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuppLits)) && (((Entry) = Vec_IntEntry(vSuppLits, i)), 1); i++  {
        pObj = Aig_ManCi( pAig, Abc_Lit2Var(Entry) );
        Aig_ObjSetTerValue( pObj, Abc_LitIsCompl(Entry) ? AIG_VAL0 : AIG_VAL1 );
        Aig_ObjSetTravIdCurrent( pAig, pObj );
//printf( "%d ", Entry );
    }
//printf( "\n" );
    return Aig_ObjTerSimulate_rec( pAig, pNode );
}


type Aig_ManPack_t = Aig_ManPack_t_;
// extern Aig_ManPack_t *        Aig_ManPackStart( Aig_Man_t * pAig );

// extern void                   Aig_ManPackStop( Aig_ManPack_t * p );

// extern void                   Aig_ManPackAddPattern( Aig_ManPack_t * p, Vec_Int_t * vLits );

// extern Vec_Int_t *            Aig_ManPackConstNodes( Aig_ManPack_t * p );


/**Function*************************************************************

  Synopsis    [Testing of the framework.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManJustExperiment(pAig *Aig_Man_t) {
    var pPack *Aig_ManPack_t;
    var vSuppLits, vNodes *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    var Count0, Count0f, Count1, Count1f int32;
    var nTotalLits int32;
    vSuppLits = Vec_IntAlloc( 100 );
    pPack = Aig_ManPackStart( pAig );
    vNodes = Aig_ManPackConstNodes( pPack );
//    Aig_ManForEachCo( pAig, pObj, i )
    // inlining of macro Aig_ManForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Aig_ManObj(pAig, Vec_IntEntry(vNodes,i))), 1); i++  {
        if pObj.fPhase { // const 1
        
            if Aig_ObjFindSatAssign(pAig, pObj, 0, vSuppLits) != 0 {
//                assert( Aig_ObjTerSimulate(pAig, pObj, vSuppLits) == AIG_VAL0 );
//                if ( Aig_ObjTerSimulate(pAig, pObj, vSuppLits) != AIG_VAL0 )
//                    printf( "Justification error!\n" );
                Count0++;
                nTotalLits += Vec_IntSize(vSuppLits);
                Aig_ManPackAddPattern( pPack, vSuppLits );
            
            } else {
                Count0f++; }
        
        } else {
            if Aig_ObjFindSatAssign(pAig, pObj, 1, vSuppLits) != 0 {
//                assert( Aig_ObjTerSimulate(pAig, pObj, vSuppLits) == AIG_VAL1 );
//                if ( Aig_ObjTerSimulate(pAig, pObj, vSuppLits) != AIG_VAL1 )
//                    printf( "Justification error!\n" );
                Count1++;
                nTotalLits += Vec_IntSize(vSuppLits);
                Aig_ManPackAddPattern( pPack, vSuppLits );
            
            } else {
                Count1f++; }
        }
    }
    Vec_IntFree( vSuppLits );
    fmt.Printf("PO =%6d. C0 =%6d. C0f =%6d. C1 =%6d. C1f =%6d. (%6.2f %%) Ave =%4.1f ", Aig_ManCoNum(pAig), Count0, Count0f, Count1, Count1f, 100.0*(Count0+Count1)/Aig_ManCoNum(pAig), 1.0*nTotalLits/(Count0+Count1));
    Abc_PrintTime( 1, "T", Abc_Clock() - clk );
    Aig_ManCleanMarkAB( pAig );
    Aig_ManPackStop( pPack );
    Vec_IntFree( vNodes );
}







////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

