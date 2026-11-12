package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaBidec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Application of bi-decomposition to AIG minimization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaBidec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "bool/bdc/bdc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes truth table of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManConvertAigToTruth_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vTruth *Vec_Int_t, nWords int32, vVisited *Vec_Int_t) *uint32 {
    var pTruth, pTruth0, pTruth1 *uint32;
    var i int32;
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    if Vec_IntGetEntry(p.vTruths, Gia_ObjId(p, pObj)) != -1 {
        return (unsigned *)Vec_IntEntryP( vTruth, nWords * Vec_IntGetEntry(p.vTruths, Gia_ObjId(p, pObj)) ); }
    // compute the truth tables of the fanins
    pTruth0 = Gia_ManConvertAigToTruth_rec( p, Gia_ObjFanin0(pObj), vTruth, nWords, vVisited );
    pTruth1 = Gia_ManConvertAigToTruth_rec( p, Gia_ObjFanin1(pObj), vTruth, nWords, vVisited );
    // get room for the truth table
    pTruth = Vec_IntFetch( vTruth, nWords );
    // create the truth table of the node
    if !Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] & pTruth1[i];
        }
    } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] & ~pTruth1[i];
        }
    } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = ~pTruth0[i] & pTruth1[i];
        }
    } else { // if ( Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) )
        for  i = 0; i < nWords; i++  {
            pTruth[i] = ~pTruth0[i] & ~pTruth1[i];
        } }
    // save the visited node
    Vec_IntSetEntry( p.vTruths, Gia_ObjId(p, pObj), Vec_IntSize(vVisited) );
    Vec_IntPush( vVisited, Gia_ObjId(p, pObj) );
    return pTruth;
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the node.]

  Description [Assumes that the structural support is no more than 8 inputs.
  Uses array vTruth to store temporary truth tables. The returned pointer should 
  be used immediately.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManConvertAigToTruth(p *Gia_Man_t, pRoot *Gia_Obj_t, vLeaves *Vec_Int_t, vTruth *Vec_Int_t, vVisited *Vec_Int_t) *uint32 {
    static unsigned uTruths[8][8] = { // elementary truth tables
        { 0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA },
        { 0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC },
        { 0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0 },
        { 0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00 },
        { 0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000 }, 
        { 0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF }, 
        { 0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF,0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF }, 
        { 0x00000000,0x00000000,0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF,0xFFFFFFFF,0xFFFFFFFF } 
    };
    var pObj *Gia_Obj_t;
    var vTtElems *Vec_Ptr_t  = nil;
    var pTruth *uint32;//, * pTruth2;
    var i, nWords, nVars int32;
    // get the number of variables and words
    nVars  = Vec_IntSize( vLeaves );
    nWords = Abc_TruthWordNum( nVars );
    // check the case of a constant
    if Gia_ObjIsConst0( Gia_Regular(pRoot) ) {
        Vec_IntClear( vTruth );
        // get room for the truth table
        pTruth = Vec_IntFetch( vTruth, nWords );
        if !Gia_IsComplement(pRoot) {
            Gia_ManTruthClear( pTruth, nVars );
        } else {
            Gia_ManTruthFill( pTruth, nVars ); }
        return pTruth;
    }
    // if the number of variables is more than 8, allocate truth tables
    if nVars > 8 {
        vTtElems = Vec_PtrAllocTruthTables( nVars ); }
    // assign elementary truth tables
    Vec_IntClear( vTruth );
    Vec_IntClear( vVisited );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vLeaves)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLeaves,i))); i++  {
        // get room for the truth table
        pTruth = Vec_IntFetch( vTruth, nWords );
        // assign elementary variable
        if vTtElems != nil {
            Gia_ManTruthCopy( pTruth, (unsigned *)Vec_PtrEntry(vTtElems, i), nVars );
        } else {
            Gia_ManTruthCopy( pTruth, uTruths[i], nVars ); }
        // save the visited node
        Vec_IntSetEntry( p.vTruths, Gia_ObjId(p, pObj), Vec_IntSize(vVisited) );
        Vec_IntPush( vVisited, Gia_ObjId(p, pObj) );
    }
    if vTtElems != nil {
        Vec_PtrFree( vTtElems ); }
    // clear the marks and compute the truth table
//    pTruth2 = Gia_ManConvertAigToTruth_rec( p, Gia_Regular(pRoot), vTruth, nWords, vVisited );
    pTruth = Gia_ManConvertAigToTruth_rec( p, Gia_Regular(pRoot), vTruth, nWords, vVisited );
    // copy the result
//    Gia_ManTruthCopy( pTruth, pTruth2, nVars );
    if Gia_IsComplement(pRoot) {
        Gia_ManTruthNot( pTruth, pTruth, nVars ); }
    // clean truth tables
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vVisited)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vVisited,i))); i++  {
        Vec_IntSetEntry( p.vTruths, Gia_ObjId(p, pObj), -1 );
    }
    return pTruth;
}

/**Function*************************************************************

  Synopsis    [Resynthesizes nodes using bi-decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjPerformBidec(pManDec *Bdc_Man_t, pNew *Gia_Man_t, p *Gia_Man_t, pRoot *Gia_Obj_t, vLeaves *Vec_Int_t, vTruth *Vec_Int_t, vVisited *Vec_Int_t) int32 {
    var pTruth *uint32;
    var pFunc *Bdc_Fun_t;
    var pFanin *Gia_Obj_t;
    var i, iFan, nVars, nNodes int32;
    // collect leaves of this gate
    Vec_IntClear( vLeaves );
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(p,Gia_ObjId(p, pRoot)) && ((iFan = Gia_ObjLutFanins(p,Gia_ObjId(p, pRoot))[i]),1); i++  {
        Vec_IntPush( vLeaves, iFan );
    }
    nVars = Vec_IntSize( vLeaves );
    assert(nVars < 16, "nVars < 16");
    // derive truth table
    pTruth = Gia_ManConvertAigToTruth( p, pRoot, vLeaves, vTruth, vVisited );
//Extra_PrintBinary( stdout, pTruth, (1<<nVars) ); printf( "\n" );
    if Gia_ManTruthIsConst0(pTruth, nVars) {
//printf( "Node %d is const0\n", Gia_ObjId(p, pRoot) );
        return 0;
    }
    if Gia_ManTruthIsConst1(pTruth, nVars) {
//printf( "Node %d is const1\n", Gia_ObjId(p, pRoot) );
        return 1;
    }
    // decompose truth table
    Bdc_ManDecompose( pManDec, pTruth, nil, nVars, nil, 1000 );
/*
if ( Bdc_ManNodeNum(pManDec) == 0 )
    printf( "Node %d has 0 bidec nodes\n", Gia_ObjId(p, pRoot) );
if ( Kit_TruthSupportSize(pTruth, nVars) != nVars )
{
    printf( "Node %d has %d fanins and %d supp size\n", Gia_ObjId(p, pRoot), nVars, Kit_TruthSupportSize(pTruth, nVars) );
    Gia_LutForEachFanin( p, Gia_ObjId(p, pRoot), iFan, i )
    {
        printf( "%d  ", Kit_TruthVarInSupport(pTruth, nVars, i) );
        Gia_ObjPrint( p, Gia_ManObj(p, iFan) );
    }
//    Gia_ManPrintStats( p, 0 );
}
*/
    // convert back into HOP
    Bdc_FuncSetCopy( Bdc_ManFunc( pManDec, 0 ), Gia_ManConst1(pNew) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vLeaves)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(vLeaves,i))); i++  {
        Bdc_FuncSetCopyInt( Bdc_ManFunc( pManDec, i+1 ), Gia_ObjValue(pFanin) );
    }
    nNodes = Bdc_ManNodeNum( pManDec );
    for  i = nVars + 1; i < nNodes; i++  {
        pFunc = Bdc_ManFunc( pManDec, i );
        Bdc_FuncSetCopyInt( pFunc, Gia_ManHashAnd( pNew, Bdc_FunFanin0Copy(pFunc), Bdc_FunFanin1Copy(pFunc) ) );
    }
    return Bdc_FunObjCopy( Bdc_ManRoot(pManDec) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformBidec(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var pManDec *Bdc_Man_t;
    var Pars Bdc_Par_t ;
    var pPars *Bdc_Par_t = &Pars;
    var vLeaves, vTruth, vVisited *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;//, clk = Abc_Clock();
    pPars.nVarsMax = Gia_ManLutSizeMax( p );
    pPars.fVerbose = fVerbose;
    if pPars.nVarsMax < 2 {
        fmt.Printf("Resynthesis is not performed when nodes have less than 2 inputs.\n");
        return nil;
    }
    if pPars.nVarsMax > 15 {
        fmt.Printf("Resynthesis is not performed when nodes have more than 15 inputs.\n");
        return nil;
    }
    vLeaves  = Vec_IntAlloc( 0 );
    vTruth   = Vec_IntAlloc( (1<<16) );
    vVisited = Vec_IntAlloc( 0 );
    // clean the old manager
    Gia_ManCleanTruth( p );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // start the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
//    Gia_ManCleanLevels( pNew, Gia_ManObjNum(p) );
    pManDec = Bdc_ManAlloc( pPars );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {        
        if Gia_ObjIsCi(pObj) { // transfer the CI level (is it needed?)
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsLut(p, i) {
            pObj.Value = Gia_ObjPerformBidec( pManDec, pNew, p, pObj, vLeaves, vTruth, vVisited ); }
    }
    Bdc_ManFree( pManDec );
    // cleanup the AIG
    Gia_ManHashStop( pNew );
    // check the presence of dangling nodes
    if Gia_ManHasDangling(pNew) {
        pNew = Gia_ManCleanup( pTemp = pNew );
        if Gia_ManAndNum(pNew) != Gia_ManAndNum(pTemp) {
            fmt.Printf("Gia_ManPerformBidec() node count before and after: %6d and %6d.\n", Gia_ManAndNum(pNew), Gia_ManAndNum(pTemp)); }
        Gia_ManStop( pTemp );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vTruth );
    Vec_IntFree( vVisited );
    if fVerbose != 0 {
//        printf( "Total gain in AIG nodes = %d.  ", Gia_ManObjNum(p)-Gia_ManObjNum(pNew) );
//        ABC_PRT( "Total runtime", Abc_Clock() - clk );
    }
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

