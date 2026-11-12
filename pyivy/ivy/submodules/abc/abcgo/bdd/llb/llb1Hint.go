package llb

/**CFile****************************************************************

  FileName    [llb1Hint.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Cofactors the circuit w.r.t. the high-fanout variables.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb1Hint.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns CI index with the largest number of fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManMaxFanoutCi(pAig *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t; 
    var i int32;
    var WeightMax int32 = -ABC_INFINITY;
    var iInput int32 = -1;
    Aig_ManForEachCi( pAig, pObj, i )
        if WeightMax < Aig_ObjRefs(pObj) {
            WeightMax = Aig_ObjRefs(pObj);
            iInput = i;
        }
    assert(iInput >= 0, "iInput >= 0");
    return iInput;
}

/**Function*************************************************************

  Synopsis    [Derives AIG whose PI is substituted by a constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManPerformHints(pAig *Aig_Man_t, nHintDepth int32) *Aig_Man_t {
    var pNew, pTemp *Aig_Man_t;
    var i, iInput int32;
    pNew = Aig_ManDupDfs( pAig );
    for  i = 0; i < nHintDepth; i++  {
        iInput = Llb_ManMaxFanoutCi( pNew );
        Abc_Print( 1, "%d %3d\n", i, iInput );
        pNew = Aig_ManDupCof( pTemp = pNew, iInput, 1 );
        Aig_ManStop( pTemp );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns CI index with the largest number of fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCollectHighFanoutObjects(pAig *Aig_Man_t, nCandMax int32, fCisOnly int32) *Vec_Int_t {
    var vFanouts, vResult *Vec_Int_t;
    var pObj *Aig_Obj_t; 
    var i, fChanges, PivotValue int32;
//    int Entry;
    // collect fanout counts
    vFanouts = Vec_IntAlloc( 100 );
    Aig_ManForEachObj( pAig, pObj, i )
    {
//        if ( !Aig_ObjIsCi(pObj) && (fCisOnly || !Aig_ObjIsNode(pObj)) )
        if !Saig_ObjIsLo(pAig,pObj) && ((fCisOnly != 0) || !Aig_ObjIsNode(pObj)) {
            continue; }
        Vec_IntPush( vFanouts, Aig_ObjRefs(pObj) );
    }
    Vec_IntSort( vFanouts, 1 );
    // pick the separator
    nCandMax = Abc_MinInt( nCandMax, Vec_IntSize(vFanouts) - 1 );
    PivotValue = Vec_IntEntry( vFanouts, nCandMax );
    Vec_IntFree( vFanouts );
    // collect obj satisfying the constraints
    vResult = Vec_IntAlloc( 100 );
    Aig_ManForEachObj( pAig, pObj, i )
    {
//        if ( !Aig_ObjIsCi(pObj) && (fCisOnly || !Aig_ObjIsNode(pObj)) )
        if !Saig_ObjIsLo(pAig,pObj) && ((fCisOnly != 0) || !Aig_ObjIsNode(pObj)) {
            continue; }
        if Aig_ObjRefs(pObj) < PivotValue {
            continue; }
        Vec_IntPush( vResult, Aig_ObjId(pObj) );
    }
    assert(Vec_IntSize(vResult) >= nCandMax, "Vec_IntSize(vResult) >= nCandMax");
    // order in the decreasing order of fanouts
    for
    {
        fChanges = 0;
        for  i = 0; i < Vec_IntSize(vResult) - 1; i++  {
            if Aig_ObjRefs(Aig_ManObj(pAig, Vec_IntEntry(vResult, i))) < 
                 Aig_ObjRefs(Aig_ManObj(pAig, Vec_IntEntry(vResult, i+1))) {
                var Temp int32  = Vec_IntEntry( vResult, i );
                Vec_IntWriteEntry( vResult, i, Vec_IntEntry(vResult, i+1) );
                Vec_IntWriteEntry( vResult, i+1, Temp );
                fChanges = 1;
            }
        }
    
if !( fChanges ) { break }
}
    
/*
    Vec_IntForEachEntry( vResult, Entry, i )
        printf( "%d ", Aig_ObjRefs(Aig_ManObj(pAig, Entry)) );
printf( "\n" );
*/
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Derives AIG whose PI is substituted by a constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManModelCheckAigWithHints(pAigGlo *Aig_Man_t, pPars *Gia_ParLlb_t) int32 {
    var ddGlo *DdManager  = nil;
    var vHints *Vec_Int_t;
    var vHFCands *Vec_Int_t;
    var i int32;
    var Entry int32;
    var RetValue int32 = -1;
    var clk abctime  = Abc_Clock();
    assert(pPars.nHintDepth > 0, "pPars.nHintDepth > 0");
/*
    // perform reachability without hints
    RetValue = Llb_ManModelCheckAig( pAigGlo, pPars, NULL, NULL );
    if ( RetValue >= 0 )
        return RetValue;
*/
    // create hints representation
    vHFCands = Llb_ManCollectHighFanoutObjects( pAigGlo, pPars.nHintDepth+pPars.HintFirst, 1 );
    vHints   = Vec_IntStartFull( Aig_ManObjNumMax(pAigGlo) );
    // add one hint at a time till the problem is solved
    Vec_IntForEachEntryStart( vHFCands, Entry, i, pPars.HintFirst )
    {
        Vec_IntWriteEntry( vHints, Entry, 1 );   // change to 1 to start from zero cof!!!
        // solve under hints
        RetValue = Llb_ManModelCheckAig( pAigGlo, pPars, vHints, &ddGlo );
        if RetValue == 0 {
            goto Finish; }
        if RetValue == 1 {
            break; }
    }
    if RetValue == -1 {
        goto Finish; }
    // undo the hints one at a time
    for  ; i >= pPars.HintFirst; i--  {
        Entry = Vec_IntEntry( vHFCands, i );
        Vec_IntWriteEntry( vHints, Entry, -1 );        
        // solve under relaxed hints
        RetValue = Llb_ManModelCheckAig( pAigGlo, pPars, vHints, &ddGlo );
        if RetValue == 0 {
            goto Finish; }
        if RetValue == 1 {
            continue; }
        break;
    }
Finish:
    if ddGlo != nil {
        if ddGlo.bFunc {
            Cudd_RecursiveDeref( ddGlo, ddGlo.bFunc ); } 
        Extra_StopManager( ddGlo );
    }
    Vec_IntFreeP( &vHFCands );
    Vec_IntFreeP( &vHints );
    if pPars.fVerbose {
        Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clk ); }
    return RetValue;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

