package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcGen.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Generating satisfying assignments.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcGen.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMoObj(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, iObj * p.nSimWords );
}
func Gia_ManMoSetCi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSims *word  = Gia_ManMoObj( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSims[w] = Gia_ManRandomW( 0 );
    }
}
func Gia_ManMoSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSims *word   = Gia_ManMoObj( p, iObj );
    var pSims0 *word  = Gia_ManMoObj( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSims1 *word  = Gia_ManMoObj( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        if Gia_ObjFaninC1(pObj) {
            for  w = 0; w < p.nSimWords; w++  {
                pSims[w] = ~(pSims0[w] | pSims1[w]);
            }
        } else { 
            for  w = 0; w < p.nSimWords; w++  {
                pSims[w] = ~pSims0[w] & pSims1[w];
            } }
    
    } else {
        if Gia_ObjFaninC1(pObj) {
            for  w = 0; w < p.nSimWords; w++  {
                pSims[w] = pSims0[w] & ~pSims1[w];
            }
        } else { 
            for  w = 0; w < p.nSimWords; w++  {
                pSims[w] = pSims0[w] & pSims1[w];
            } }
    }
}
func Gia_ManMoSetCo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t   = Gia_ManObj( p, iObj );
    var pSims *word   = Gia_ManMoObj( p, iObj );
    var pSims0 *word  = Gia_ManMoObj( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSims[w] = ~pSims0[w];
        }
    
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSims[w] = pSims0[w];
        }
    }
}
func Gia_ManMoFindSimulate(p *Gia_Man_t, nWords int32) {
    var i, iObj int32;
    Gia_ManRandomW( 1 );
    p.nSimWords = nWords;
    if p.vSims {
        Vec_WrdFill( p.vSims, nWords * Gia_ManObjNum(p), 0 );
    } else {
        p.vSims = Vec_WrdStart( nWords * Gia_ManObjNum(p) ); }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((iObj) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Gia_ManMoSetCi( p, iObj );
    }
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < p.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, iObj)) {} else {
        Gia_ManMoSimAnd( p, iObj ); }
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((iObj) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        Gia_ManMoSetCo( p, iObj );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTestSatEnum(p *Gia_Man_t) int32 {
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var clkTotal abctime;
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0);
    var i int32;
    var v int32;
    var status int32;
    var iLit int32;
    var nWords int32 = 1;
    var iOutVar int32 = 1;
    var Count int32;
    var vVars *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var pSimInfo *word;

    // add literals to the solver
    iLit  = Abc_Var2Lit( iOutVar, 0 ); 
    status = sat_solver_addclause( pSat, &iLit, &iLit + 1 );
    assert((status != 0), "status");

    // simulate the AIG
    Gia_ManMoFindSimulate( p, nWords );

    // print outputs
    pSimInfo = Gia_ManMoObj( p, Gia_ObjId(p, Gia_ManCo(p, 0)) );
    for  i = 0; i < 64*nWords; i++  {
        fmt.Printf("%d", Abc_InfoHasBit( (unsigned *)pSimInfo, i ));
    }
    fmt.Printf("\n");

    // iterate through the assignments
    for  i = 0; i < 64*nWords; i++  {
        Vec_IntClear( vVars );
        for  v = 0; v < Gia_ManObjNum(p); v++  {
            if pCnf.pVarNums[v] == -1 {
                continue; }
            pSimInfo = Gia_ManMoObj( p, v );
            if !Abc_InfoHasBit( (unsigned *)pSimInfo, i ) {
                continue; }
            Vec_IntPush( vVars, pCnf.pVarNums[v] );
        }
        //sat_solver_act_var_clear( pSat );
        //sat_solver_set_polarity( pSat, Vec_IntArray(vVars), Vec_IntSize(vVars) );

        clk2 = Abc_Clock();
        status = sat_solver_solve( pSat, nil, nil, 0, 0, 0, 0 );
        clkTotal += Abc_Clock() - clk2;
        fmt.Printf("%c", status == l_True ? '+' : '-');
        if status == l_True {
            Count++; }
    }
    fmt.Printf("\n");
    fmt.Printf("Finished generating %d assignments.  ", Count);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Abc_PrintTime( 1, "SAT solver time", clkTotal );

    Vec_WrdFreeP( &p.vSims );
    Vec_IntFree( vVars );
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

