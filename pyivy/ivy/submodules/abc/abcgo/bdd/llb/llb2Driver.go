package llb

/**CFile****************************************************************

  FileName    [llb2Driver.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Procedures working with flop drivers.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Driver.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// driver issue:arises when creating
// - driver ref-counter array
// - Ns2Glo maps
// - final partition
// - change-phase cube

// LI variable is used when
// - driver drives more than one LI
// - driver is a PI
// - driver is a constant

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the array of times each flop driver is referenced.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_DriverCountRefs(p *Aig_Man_t) *Vec_Int_t {
    var vCounts *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vCounts = Vec_IntStart( Aig_ManObjNumMax(p) );
    Saig_ManForEachLi( p, pObj, i )
        Vec_IntAddToEntry( vCounts, Aig_ObjFaninId0(pObj), 1 );
    return vCounts;
}

/**Function*************************************************************

  Synopsis    [Returns array of NS variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_DriverCollectNs(pAig *Aig_Man_t, vDriRefs *Vec_Int_t) *Vec_Int_t {
    var vVars *Vec_Int_t;
    var pObj, pDri *Aig_Obj_t;
    var i int32;
    vVars = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    Saig_ManForEachLi( pAig, pObj, i )
    {
        pDri = Aig_ObjFanin0(pObj);
        if Vec_IntEntry( vDriRefs, Aig_ObjId(pDri) ) != 1 || Saig_ObjIsPi(pAig, pDri) || Aig_ObjIsConst1(pDri) {
            Vec_IntPush( vVars, Aig_ObjId(pObj) );
        } else {
            Vec_IntPush( vVars, Aig_ObjId(pDri) ); }
    }
    return vVars;
}

/**Function*************************************************************

  Synopsis    [Returns array of CS variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_DriverCollectCs(pAig *Aig_Man_t) *Vec_Int_t {
    var vVars *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vVars = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    Saig_ManForEachLo( pAig, pObj, i )
        Vec_IntPush( vVars, Aig_ObjId(pObj) );
    return vVars;
}

/**Function*************************************************************

  Synopsis    [Create cube for phase swapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_DriverPhaseCube(pAig *Aig_Man_t, vDriRefs *Vec_Int_t, dd *DdManager) *DdNode {
    var bCube, bVar, bTemp *DdNode;
    var pObj *Aig_Obj_t;
    var i int32;
    var TimeStop abctime;
    TimeStop = dd.TimeStop; dd.TimeStop = 0;
    bCube = Cudd_ReadOne( dd );  Cudd_Ref( bCube );
    Saig_ManForEachLi( pAig, pObj, i )
    {
        assert(Vec_IntEntry( vDriRefs, Aig_ObjFaninId0(pObj) ) >= 1, "Vec_IntEntry( vDriRefs, Aig_ObjFaninId0(pObj) ) >= 1");
        if Vec_IntEntry( vDriRefs, Aig_ObjFaninId0(pObj) ) != 1 {
            continue; }
        if !Aig_ObjFaninC0(pObj) {
            continue; }
        bVar  = Cudd_bddIthVar( dd, Aig_ObjFaninId0(pObj) );
        bCube = Cudd_bddAnd( dd, bTemp = bCube, bVar );  Cudd_Ref( bCube );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bCube );
    dd.TimeStop = TimeStop;
    return bCube;
}

/**Function*************************************************************

  Synopsis    [Compute the last partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_DriverLastPartition(p *Aig_Man_t, vVarsNs *Vec_Int_t, TimeTarget abctime) *DdManager {
//    int fVerbose = 1;
    var dd *DdManager;
    var bVar1, bVar2, bProd, bRes, bTemp *DdNode;
    var pObj *Aig_Obj_t;
    var i int32;
    dd = Cudd_Init( Aig_ManObjNumMax(p), 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    Cudd_AutodynEnable( dd, CUDD_REORDER_SYMM_SIFT );
    dd.TimeStop = TimeTarget;
    bRes = Cudd_ReadOne(dd);                                   Cudd_Ref( bRes );

    // mark the duplicated flop inputs
    Aig_ManForEachObjVec( vVarsNs, p, pObj, i )
    {
        if !Saig_ObjIsLi(p, pObj) {
            continue; }
        bVar1 = Cudd_bddIthVar( dd, Aig_ObjId(pObj) );
        bVar2 = Cudd_bddIthVar( dd, Aig_ObjFaninId0(pObj) );
        if Aig_ObjIsConst1(Aig_ObjFanin0(pObj)) {
            bVar2 = Cudd_ReadOne(dd); }
        // inlining of macro Cudd_NotCond
        bVar2 = ((DdNode *)((ptrint)(bVar2) ^ (Aig_ObjFaninC0(pObj))));
        bProd = Cudd_bddXnor( dd, bVar1, bVar2 );              Cudd_Ref( bProd );
//        bRes  = Cudd_bddAnd( dd, bTemp = bRes, bProd );        Cudd_Ref( bRes );
//        bRes  = Extra_bddAndTime( dd, bTemp = bRes, bProd, TimeTarget );  
        bRes  = Cudd_bddAnd( dd, bTemp = bRes, bProd );  
        if bRes == nil {
            Cudd_RecursiveDeref( dd, bTemp );
            Cudd_RecursiveDeref( dd, bProd );
            return nil;
        }        
        Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bTemp );
        Cudd_RecursiveDeref( dd, bProd );
    }

/*
    Saig_ManForEachLi( p, pObj, i )
        printf( "%d ", Aig_ObjId(pObj) );
    printf( "\n" );
    Saig_ManForEachLi( p, pObj, i )
        printf( "%c%d ", Aig_ObjFaninC0(pObj)? '-':'+', Aig_ObjFaninId0(pObj) );
    printf( "\n" );
*/
    Cudd_AutodynDisable( dd );
//    Cudd_RecursiveDeref( dd, bRes );
//    Extra_StopManager( dd );
    dd.bFunc = bRes;
    dd.TimeStop = 0;
    return dd;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

