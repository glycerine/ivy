package pdr

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [pdrTsim3.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Improved ternary simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrTsim3.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Txs3_Man_t_ struct {
    pGia  *Gia_Man_t;      // user's AIG
    vPrio  *Vec_Int_t;     // priority of each flop
    vCiObjs  *Vec_Int_t;   // cone leaves (CI obj IDs)
    vFosPre  *Vec_Int_t;   // cone leaves (CI obj IDs)
    vFosAbs  *Vec_Int_t;   // cone leaves (CI obj IDs)
    vCoObjs  *Vec_Int_t;   // cone roots (CO obj IDs)
    vCiVals  *Vec_Int_t;   // cone leaf values (0/1 CI values)
    vCoVals  *Vec_Int_t;   // cone root values (0/1 CO values)
    vNodes  *Vec_Int_t;    // cone nodes (node obj IDs)
    vTemp  *Vec_Int_t;     // cone nodes (node obj IDs)
    vPiLits  *Vec_Int_t;   // resulting array of PI literals
    vFfLits  *Vec_Int_t;   // resulting array of flop literals
    pMan  *Pdr_Man_t;      // calling manager
    nPiLits         int32;   // the number of PI literals
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start and stop the ternary simulation engine.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs3_ManStart(pMan *Pdr_Man_t, pAig *Aig_Man_t, vPrio *Vec_Int_t) *Txs3_Man_t {
    var p *Txs3_Man_t;
//    Aig_Obj_t * pObj;
//    int i;
    assert(Vec_IntSize(vPrio) == Aig_ManRegNum(pAig), "Vec_IntSize(vPrio) == Aig_ManRegNum(pAig)");
    p = new(Txs3_Man_t);
    p.pGia    = Gia_ManFromAigSimple(pAig); // user's AIG
//    Aig_ManForEachObj( pAig, pObj, i )
//        assert( i == 0 || pObj->iData == Abc_Var2Lit(i, 0) );
    p.vPrio   = vPrio;                // priority of each flop
    p.vCiObjs = Vec_IntAlloc( 100 );  // cone leaves (CI obj IDs)
    p.vFosPre = Vec_IntAlloc( 100 );  // present flop outputs
    p.vFosAbs = Vec_IntAlloc( 100 );  // absracted flop outputs
    p.vCoObjs = Vec_IntAlloc( 100 );  // cone roots (CO obj IDs)
    p.vCiVals = Vec_IntAlloc( 100 );  // cone leaf values (0/1 CI values)
    p.vCoVals = Vec_IntAlloc( 100 );  // cone root values (0/1 CO values)
    p.vNodes  = Vec_IntAlloc( 100 );  // cone nodes (node obj IDs)
    p.vTemp   = Vec_IntAlloc( 100 );  // cone nodes (node obj IDs)
    p.vPiLits = Vec_IntAlloc( 100 );  // resulting array of PI literals
    p.vFfLits = Vec_IntAlloc( 100 );  // resulting array of flop literals
    p.pMan    = pMan;                 // calling manager
    return p;
}
func Txs3_ManStop(p *Txs3_Man_t) {
    Gia_ManStop( p.pGia );
    Vec_IntFree( p.vCiObjs );
    Vec_IntFree( p.vFosPre );
    Vec_IntFree( p.vFosAbs );
    Vec_IntFree( p.vCoObjs );
    Vec_IntFree( p.vCiVals );
    Vec_IntFree( p.vCoVals );
    Vec_IntFree( p.vNodes );
    Vec_IntFree( p.vTemp );
    Vec_IntFree( p.vPiLits );
    Vec_IntFree( p.vFfLits );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Marks the TFI cone and collects CIs and nodes.]

  Description [For this procedure to work Value should not be ~0 
  at the beginning.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs3_ManCollectCone_rec(p *Txs3_Man_t, pObj *Gia_Obj_t) {
    if !~pObj.Value {
        return; }
    pObj.Value = ~0;
    if Gia_ObjIsCi(pObj) {
        var Entry int32;
        if Gia_ObjIsPi(p.pGia, pObj) {
            Vec_IntPush( p.vCiObjs, Gia_ObjId(p.pGia, pObj) );
            return;
        }
        Entry = Gia_ObjCioId(pObj) - Gia_ManPiNum(p.pGia);
        if Vec_IntEntry(p.vPrio, Entry) { // present flop
            Vec_IntPush( p.vFosPre, Gia_ObjId(p.pGia, pObj) );
        } else { // asbstracted flop
            Vec_IntPush( p.vFosAbs, Gia_ObjId(p.pGia, pObj) ); }
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Txs3_ManCollectCone_rec( p, Gia_ObjFanin0(pObj) );
    Txs3_ManCollectCone_rec( p, Gia_ObjFanin1(pObj) );
    Vec_IntPush( p.vNodes, Gia_ObjId(p.pGia, pObj) );
}
func Txs3_ManCollectCone(p *Txs3_Man_t, fVerbose int32) {
    var pObj *Gia_Obj_t; var i, Entry int32;
//    printf( "Collecting cones for clause with %d literals.\n", Vec_IntSize(vCoObjs) );
    Vec_IntClear( p.vCiObjs );
    Vec_IntClear( p.vFosPre );
    Vec_IntClear( p.vFosAbs );
    Vec_IntClear( p.vNodes );
    Gia_ManConst0(p.pGia).Value = ~0;
    Gia_ManForEachObjVec( p.vCoObjs, p.pGia, pObj, i )
        Txs3_ManCollectCone_rec( p, Gia_ObjFanin0(pObj) );
if fVerbose != 0 {
fmt.Printf("%d %d %d \n", Vec_IntSize(p.vCiObjs), Vec_IntSize(p.vFosPre), Vec_IntSize(p.vFosAbs)); }
    p.nPiLits = Vec_IntSize(p.vCiObjs);
    // sort primary inputs
    Vec_IntSelectSort( Vec_IntArray(p.vCiObjs), Vec_IntSize(p.vCiObjs) );
    // sort and add present flop variables
    Vec_IntSelectSortReverse( Vec_IntArray(p.vFosPre), Vec_IntSize(p.vFosPre) );
    Vec_IntForEachEntry( p.vFosPre, Entry, i )
        Vec_IntPush( p.vCiObjs, Entry );
    // sort and add absent flop variables
    Vec_IntSelectSortReverse( Vec_IntArray(p.vFosAbs), Vec_IntSize(p.vFosAbs) );
    Vec_IntForEachEntry( p.vFosAbs, Entry, i )
        Vec_IntPush( p.vCiObjs, Entry );
    // cleanup
    Gia_ManForEachObjVec( p.vCiObjs, p.pGia, pObj, i )
        pObj.Value = 0;
    Gia_ManForEachObjVec( p.vNodes, p.pGia, pObj, i )
        pObj.Value = 0;
}

/**Function*************************************************************

  Synopsis    [Shrinks values using ternary simulation.]

  Description [The cube contains the set of flop index literals which,
  when converted into a clause and applied to the combinational outputs, 
  led to a satisfiable SAT run in frame k (values stored in the SAT solver).
  If the cube is NULL, it is assumed that the first property output was
  asserted and failed.
  The resulting array is a set of flop index literals that asserts the COs.
  Priority contains 0 for i-th entry if the i-th FF is desirable to remove.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs3_ManTernarySim(p *Txs3_Man_t, k int32, pCube *Pdr_Set_t) *Pdr_Set_t {
//    int fTryNew = 1;
//    int fUseLit = 1;
    var fVerbose int32;
    var pSat *sat_solver;
    var pRes *Pdr_Set_t;
    var pObj *Gia_Obj_t;
    var vVar2Ids, vLits *Vec_Int_t;
    var i int32;
    var Lit int32;
    var LitAux int32;
    var Var int32;
    var Value int32;
    var RetValue int32;
    var nCoreLits int32;
    var pCoreLits *int32;//, nLits;
//    if ( k == 0 )
//        fVerbose = 1;
    // collect CO objects
    Vec_IntClear( p.vCoObjs );
    if pCube == nil { // the target is the property output
    
        pObj = Gia_ManCo(p.pGia, p.pMan.iOutCur);
        Vec_IntPush( p.vCoObjs, Gia_ObjId(p.pGia, pObj) );
    
    } else { // the target is the cube
    
        var i int32;
        for  i = 0; i < pCube.nLits; i++  {
            if pCube.Lits[i] == -1 {
                continue; }
            pObj = Gia_ManCo(p.pGia, Gia_ManPoNum(p.pGia) + Abc_Lit2Var(pCube.Lits[i]));
            Vec_IntPush( p.vCoObjs, Gia_ObjId(p.pGia, pObj) );
        }
    }
if false {
Abc_Print( 1, "Trying to justify cube " );
if pCube != nil {
    Pdr_SetPrint( os.Stdout, pCube, Gia_ManRegNum(p.pGia), nil );
} else {
    Abc_Print( 1, "<prop=fail>" ); }
Abc_Print( 1, " in frame %d.\n", k );
}

    // collect CI objects
    Txs3_ManCollectCone( p, fVerbose );
    // collect values
    Pdr_ManCollectValues( p.pMan, k, p.vCiObjs, p.vCiVals );
    Pdr_ManCollectValues( p.pMan, k, p.vCoObjs, p.vCoVals );

    // read solver
    pSat = Pdr_ManFetchSolver( p.pMan, k );
    LitAux = Abc_Var2Lit( Pdr_ManFreeVar(p.pMan, k), 0 );
    // add the clause (complemented cube) in terms of next state variables
    if pCube == nil { // the target is the property output
    
        vLits = p.pMan.vLits;
        Lit = Abc_Var2Lit( Pdr_ObjSatVar(p.pMan, k, 2, Aig_ManCo(p.pMan.pAig, p.pMan.iOutCur)), 1 ); // neg literal (property holds)
        Vec_IntFill( vLits, 1, Lit );
    
    } else {
        vLits = Pdr_ManCubeToLits( p.pMan, k, pCube, 1, 1 ); }
    // add activation literal
    Vec_IntPush( vLits, LitAux );
    RetValue = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
    assert(RetValue == 1, "RetValue == 1");
    sat_solver_compress( pSat );

    // collect assumptions 
    Vec_IntClear( p.vTemp );
    Vec_IntPush( p.vTemp, Abc_LitNot(LitAux) );
    // iterate through the values of the CI variables
    Vec_IntForEachEntryTwo( p.vCiObjs, p.vCiVals, Var, Value, i )
    {
        var pObj *Aig_Obj_t  = Aig_ManObj( p.pMan.pAig, Var );
//        iVar = Pdr_ObjSatVar( p->pMan, k, fNext ? 2 - lit_sign(pCube->Lits[i]) : 3, pObj ); assert( iVar >= 0 );
        var iVar int32  = Pdr_ObjSatVar( p.pMan, k, 3, pObj ); assert(iVar >= 0, "iVar >= 0");
        assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
        Vec_IntPush( p.vTemp, Abc_Var2Lit(iVar, !Value) );
    }
    if fVerbose != 0 {
        fmt.Printf("Clause with %d lits on lev %d\n", pCube ? pCube.nLits : 0, k);
        Vec_IntPrint( p.vTemp );
    }

/*
    // solve with assumptions
//printf( "%d -> ", Vec_IntSize(p->vTemp) );
{
abctime clk = Abc_Clock();
// assume all except flops
    Vec_IntForEachEntryStop( p->vTemp, Lit, i, p->nPiLits + 1 )
        if ( !sat_solver_push(pSat, Lit) )
        {
            assert( 0 );
        }
    nLits = sat_solver_minimize_assumptions( pSat, Vec_IntArray(p->vTemp) + p->nPiLits + 1, Vec_IntSize(p->vTemp) - p->nPiLits - 1, p->pMan->pPars->nConfLimit );
    Vec_IntShrink( p->vTemp, p->nPiLits + 1 + nLits );

p->pMan->tAbs += Abc_Clock() - clk;
    for ( i = 0; i <= p->nPiLits; i++ )
        sat_solver_pop(pSat);
}
//printf( "%d    ", nLits );
*/


    //check one last time
    RetValue = sat_solver_solve( pSat, Vec_IntArray(p.vTemp), Vec_IntLimit(p.vTemp), 0, 0, 0, 0 );
    assert(RetValue == l_False, "RetValue == l_False");

    // use analyze final
    nCoreLits = sat_solver_final(pSat, &pCoreLits);
    //assert( Vec_IntSize(p->vTemp) <= nCoreLits );

    Vec_IntClear( p.vTemp );
    for  i = 0; i < nCoreLits; i++  {
        Vec_IntPush( p.vTemp, Abc_LitNot(pCoreLits[i]) );
    }
    Vec_IntSelectSort( Vec_IntArray(p.vTemp), Vec_IntSize(p.vTemp) );

    if fVerbose != 0 {
        Vec_IntPrint( p.vTemp ); }

    // collect the resulting sets
    Vec_IntClear( p.vPiLits );
    Vec_IntClear( p.vFfLits );
    vVar2Ids = (Vec_Int_t *)Vec_PtrGetEntry( &p.pMan.vVar2Ids, k );
    Vec_IntForEachEntry( p.vTemp, Lit, i )
    {
        if Lit != Abc_LitNot(LitAux) {
            var Id int32  = Vec_IntEntry( vVar2Ids, Abc_Lit2Var(Lit) );
            var pObj *Aig_Obj_t  = Aig_ManObj( p.pMan.pAig, Id );
            assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
            if Saig_ObjIsPi(p.pMan.pAig, pObj) {
                Vec_IntPush( p.vPiLits, Abc_Var2Lit(Aig_ObjCioId(pObj), Abc_LitIsCompl(Lit)) );
            } else {
                Vec_IntPush( p.vFfLits, Abc_Var2Lit(Aig_ObjCioId(pObj) - Saig_ManPiNum(p.pMan.pAig), Abc_LitIsCompl(Lit)) ); }
        }
    }
    assert(Vec_IntSize(p.vTemp) == Vec_IntSize(p.vPiLits) + Vec_IntSize(p.vFfLits) + 1, "Vec_IntSize(p.vTemp) == Vec_IntSize(p.vPiLits) + Vec_IntSize(p.vFfLits) + 1");

    // move abstracted literals from flops to inputs
    if p.pMan.pPars.fUseAbs && p.pMan.vAbsFlops {
        var i, iLit, k int32;
        Vec_IntForEachEntry( p.vFfLits, iLit, i )
        {
            if Vec_IntEntry(p.pMan.vAbsFlops, Abc_Lit2Var(iLit)) { // used flop
                Vec_IntWriteEntry( p.vFfLits, k++, iLit );
            } else {
                Vec_IntPush( p.vPiLits, 2*Saig_ManPiNum(p.pMan.pAig) + iLit ); }
        }
        Vec_IntShrink( p.vFfLits, k );
    }

    if fVerbose != 0 {
        Vec_IntPrint( p.vPiLits ); }
    if fVerbose != 0 {
        Vec_IntPrint( p.vFfLits ); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }

    // derive the final set
    pRes = Pdr_SetCreate( p.vFfLits, p.vPiLits );
    //ZH: Disabled assertion because this invariant doesn't hold with down
    //because of the join operation which can bring in initial states
    assert(k == 0 || !Pdr_SetIsInit(pRes, -1), "k == 0 || !Pdr_SetIsInit(pRes, -1)");
    return pRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
