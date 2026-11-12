package bmc

/**CFile****************************************************************

  FileName    [bmcExpand.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Expanding cubes against the offset.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcExpand.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// iterator thought the cubes
// Original macro:
// #define Bmc_SopForEachCube( pSop, nVars, pCube )  for ( pCube = (pSop); *pCube; pCube += (nVars) + 3 )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Expands cubes against the offset given as an AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjExpandCubesTry(vSop *Vec_Str_t, pSat *sat_solver, vVars *Vec_Int_t) int32 {
    extern int Bmc_CollapseExpandRound( sat_solver * pSat, sat_solver * pSatOn, Vec_Int_t * vLits, Vec_Int_t * vNums, Vec_Int_t * vTemp, int nBTLimit, int fCanon, int fOnOffSetLit );

    var pCube string;
    var pSop string = Vec_StrArray(vSop);
    var nCubes int32  = Abc_SopGetCubeNum(pSop);
    var nVars int32  = Abc_SopGetVarNum(pSop);

    var vLits *Vec_Int_t  = Vec_IntAlloc( nVars );
    var vTemp *Vec_Int_t  = Vec_IntAlloc( nVars );

    assert(nVars == Vec_IntSize(vVars), "nVars == Vec_IntSize(vVars)");
    assert(Vec_StrSize(vSop) == nCubes * (nVars + 3) + 1, "Vec_StrSize(vSop) == nCubes * (nVars + 3) + 1");
    // inlining of macro Bmc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        var k, Entry int32;
        // collect literals and clean cube
        Vec_IntFill( vLits, nVars, -1 );
        for  k = 0; k < nVars; k++  {
            if pCube[k] == '-' {
                continue; }
            Vec_IntWriteEntry( vLits, k, Abc_Var2Lit(Vec_IntEntry(vVars, k), pCube[k] == '0') );
            pCube[k] = '-';
        }
        // expand cube
        Bmc_CollapseExpandRound( pSat, nil, vLits, nil, vTemp, 0, 0, -1 );
        // insert literals
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLits)) && (((Entry) = Vec_IntEntry(vLits, k)), 1); k++  {
            if Entry != -1 {
                pCube[k] = '1' - Abc_LitIsCompl(Entry); }
        }
    }

    Vec_IntFree( vLits );
    Vec_IntFree( vTemp );
    return nCubes;
}
func Abc_ObjExpandCubes(vSop *Vec_Str_t, p *Gia_Man_t, nVars int32) int32 {
    extern int Bmc_CollapseIrredundantFull( Vec_Str_t * vSop, int nCubes, int nVars );

    var fReverse int32;
    var vVars *Vec_Int_t  = Vec_IntAlloc( nVars );
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver(pCnf, 1, 0);
    var v int32;
    var n int32;
    var iLit int32;
    var status int32;
    var nCubesNew int32;
    var iCiVarBeg int32 = sat_solver_nvars(pSat) - nVars;

    // check that on-set/off-set is sat
    var iOutVar int32  = 1;
    for  n = 0; n < 2; n++  {
        iLit  = Abc_Var2Lit( iOutVar, n ); // n=0 => F=1   n=1 => F=0
        status = sat_solver_solve( pSat, &iLit, &iLit + 1, 0, 0, 0, 0 );
        if status == l_False {
            Vec_StrClear( vSop );
            Vec_StrPrintStr( vSop, n ? " 1\n" : " 0\n" );
            Vec_StrPush( vSop, '\0' );
            return 1;
        }
    }

    // add literals to the solver
    iLit  = Abc_Var2Lit( iOutVar, 1 ); 
    status = sat_solver_addclause( pSat, &iLit, &iLit + 1 );
    assert((status != 0), "status");

    // collect variables
    if fReverse != 0 {
        for  v = nVars - 1; v >= 0; v--  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        }
    } else {
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vVars, iCiVarBeg + v );
        } }

    nCubesNew = Abc_ObjExpandCubesTry( vSop, pSat, vVars );

    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Vec_IntFree( vVars );
    if nCubesNew > 1 {
        Bmc_CollapseIrredundantFull( vSop, nCubesNew, nVars ); }
    return 0;
}
func Abc_NtkExpandCubes(pNtk *Abc_Ntk_t, pGia *Gia_Man_t, fVerbose int32) {
    var pNew *Gia_Man_t;
    var pObj *Abc_Obj_t; var i int32;
    var vSop *Vec_Str_t  = Vec_StrAlloc( 1000 );
    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");
    assert(Abc_NtkCiNum(pNtk) == Gia_ManCiNum(pGia), "Abc_NtkCiNum(pNtk) == Gia_ManCiNum(pGia)");
    assert(Abc_NtkCoNum(pNtk) == Gia_ManCoNum(pGia), "Abc_NtkCoNum(pNtk) == Gia_ManCoNum(pGia)");
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObj = Abc_ObjFanin0(pObj);
        if !Abc_ObjIsNode(pObj) || Abc_ObjFaninNum(pObj) == 0 {
            continue; }
        assert(Abc_ObjFaninNum(pObj) == Gia_ManCiNum(pGia), "Abc_ObjFaninNum(pObj) == Gia_ManCiNum(pGia)");

        Vec_StrClear( vSop );
        Vec_StrAppend( vSop, (char *)pObj.pData );
        Vec_StrPush( vSop, '\0' );

        pNew = Gia_ManDupCones( pGia, &i, 1, 0 );
        assert(Gia_ManCiNum(pNew) == Gia_ManCiNum(pGia), "Gia_ManCiNum(pNew) == Gia_ManCiNum(pGia)");
        if Abc_ObjExpandCubes( vSop, pNew, Abc_ObjFaninNum(pObj) ) != 0 {
            Vec_IntClear( &pObj.vFanins ); }
        Gia_ManStop( pNew );

        pObj.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, Vec_StrArray(vSop) );
    }
    Vec_StrFree( vSop );
    Abc_NtkSortSops( pNtk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

