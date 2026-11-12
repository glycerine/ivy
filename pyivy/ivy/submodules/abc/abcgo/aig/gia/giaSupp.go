package gia



import "fmt"

type Gia_ManMin_t = Gia_ManMin_t_

/**CFile****************************************************************

  FileName    [giaSupp.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Support minimization for AIGs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSupp.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "sat/satoko/satoko.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#ifdef ABC_USE_CUDD

struct Gia_ManMin_t_ 
{
    // problem formulation
    Gia_Man_t *     pGia;
    int             iLits[2];
    // structural information
    Vec_Int_t *     vCis[2];
    Vec_Int_t *     vObjs[2];
    Vec_Int_t *     vCleared;  
    // intermediate functions
    DdManager *     dd;
    Vec_Ptr_t *     vFuncs;
    Vec_Int_t *     vSupp;
};

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create/delete the data representation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Gia_ManMin_t * Gia_ManSuppStart( Gia_Man_t * pGia )
{
    Gia_ManMin_t * p;
    p = ABC_CALLOC( Gia_ManMin_t, 1 );
    p->pGia     = pGia;
    p->vCis[0]  = Vec_IntAlloc( 512 );
    p->vCis[1]  = Vec_IntAlloc( 512 );
    p->vObjs[0] = Vec_IntAlloc( 512 );
    p->vObjs[1] = Vec_IntAlloc( 512 );
    p->vCleared = Vec_IntAlloc( 512 );
    p->dd = Cudd_Init( 0, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
//    Cudd_AutodynEnable( p->dd,  CUDD_REORDER_SYMM_SIFT );
    Cudd_AutodynDisable( p->dd );
    p->vFuncs   = Vec_PtrAlloc( 10000 );
    p->vSupp    = Vec_IntAlloc( 10000 );
    return p;
}
void Gia_ManSuppStop( Gia_ManMin_t * p )
{
    Vec_IntFreeP( &p->vCis[0] );
    Vec_IntFreeP( &p->vCis[1] );
    Vec_IntFreeP( &p->vObjs[0] );
    Vec_IntFreeP( &p->vObjs[1] );
    Vec_IntFreeP( &p->vCleared );
    Vec_PtrFreeP( &p->vFuncs );
    Vec_IntFreeP( &p->vSupp );
    printf( "Refs = %d. \n", Cudd_CheckZeroRef( p->dd ) );
    Cudd_Quit( p->dd );
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Compute variables, which are not in the support.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Gia_ManFindRemoved( Gia_ManMin_t * p ) 
{
    extern void ddSupportStep2( DdNode * f, int * support );
    extern void ddClearFlag2( DdNode * f );

    //int fVerbose = 1;
    int nBddLimit = 100000;
    int nPart0 = Vec_IntSize(p->vCis[0]);
    int n, i, iObj, nVars = 0;
    DdNode * bFunc0, * bFunc1, * bFunc;
    Vec_PtrFillExtra( p->vFuncs, Gia_ManObjNum(p->pGia), NULL );
    // assign variables
    for ( n = 0; n < 2; n++ )
        Vec_IntForEachEntry( p->vCis[n], iObj, i )
            Vec_PtrWriteEntry( p->vFuncs, iObj, Cudd_bddIthVar(p->dd, nVars++) );
    // create nodes
    for ( n = 0; n < 2; n++ )
        Vec_IntForEachEntry( p->vObjs[n], iObj, i )
        {
            Gia_Obj_t * pObj = Gia_ManObj( p->pGia, iObj );
            bFunc0 = Cudd_NotCond( (DdNode *)Vec_PtrEntry(p->vFuncs, Gia_ObjFaninId0(pObj, iObj)), Gia_ObjFaninC0(pObj) );
            bFunc1 = Cudd_NotCond( (DdNode *)Vec_PtrEntry(p->vFuncs, Gia_ObjFaninId1(pObj, iObj)), Gia_ObjFaninC1(pObj) );
            bFunc  = Cudd_bddAndLimit( p->dd, bFunc0, bFunc1, nBddLimit );  
            assert( bFunc != NULL );
            Cudd_Ref( bFunc );
            Vec_PtrWriteEntry( p->vFuncs, iObj, bFunc );
        }
    // create new node
    bFunc0 = Cudd_NotCond( (DdNode *)Vec_PtrEntry(p->vFuncs, Abc_Lit2Var(p->iLits[0])), Abc_LitIsCompl(p->iLits[0]) );
    bFunc1 = Cudd_NotCond( (DdNode *)Vec_PtrEntry(p->vFuncs, Abc_Lit2Var(p->iLits[1])), Abc_LitIsCompl(p->iLits[1]) );
    bFunc  = Cudd_bddAndLimit( p->dd, bFunc0, bFunc1, nBddLimit );  
    assert( bFunc != NULL );
    Cudd_Ref( bFunc );
    //if ( fVerbose ) Extra_bddPrint( p->dd, bFunc ), printf( "\n" );
    // collect support
    Vec_IntFill( p->vSupp, nVars, 0 );
    ddSupportStep2( Cudd_Regular(bFunc), Vec_IntArray(p->vSupp) );
    ddClearFlag2( Cudd_Regular(bFunc) );
    // find variables not present in the support
    Vec_IntClear( p->vCleared );
    for ( i = 0; i < nVars; i++ )
        if ( Vec_IntEntry(p->vSupp, i) == 0 )
            Vec_IntPush( p->vCleared, i < nPart0 ? Vec_IntEntry(p->vCis[0], i) : Vec_IntEntry(p->vCis[1], i-nPart0) );
    //printf( "%d(%d)%d  ", Cudd_SupportSize(p->dd, bFunc), Vec_IntSize(p->vCleared), Cudd_DagSize(bFunc) );
    // deref results
    Cudd_RecursiveDeref( p->dd, bFunc );
    for ( n = 0; n < 2; n++ )
        Vec_IntForEachEntry( p->vObjs[n], iObj, i )
            Cudd_RecursiveDeref( p->dd, (DdNode *)Vec_PtrEntry(p->vFuncs, iObj) );
    //Vec_IntPrint( p->vCleared );
    return Vec_IntSize(p->vCleared);
}

/**Function*************************************************************

  Synopsis    [Compute variables, which are not in the support.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Gia_ManRebuildOne( Gia_ManMin_t * p, int n ) 
{
    int i, iObj, iGiaLitNew = -1;
    Vec_Int_t * vTempIns = p->vCis[n]; 
    Vec_Int_t * vTempNds = p->vObjs[n];
    Vec_Int_t * vCopies  = &p->pGia->vCopies;
    Vec_IntFillExtra( vCopies, Gia_ManObjNum(p->pGia), -1 );
    assert( p->iLits[n] >= 2 );
    // process inputs
    Vec_IntForEachEntry( vTempIns, iObj, i )
        Vec_IntWriteEntry( vCopies, iObj, Abc_Var2Lit(iObj, 0) );
    // process constants
    assert( Vec_IntSize(p->vCleared) > 0 );
    Vec_IntForEachEntry( p->vCleared, iObj, i )
        Vec_IntWriteEntry( vCopies, iObj, 0 );
    if ( Vec_IntSize(vTempNds) == 0 )
        iGiaLitNew = Vec_IntEntry( vCopies, Abc_Lit2Var(p->iLits[n]) );
    else
    {
        Vec_IntForEachEntry( vTempNds, iObj, i )
        {
            Gia_Obj_t * pObj = Gia_ManObj( p->pGia, iObj );
            int iGiaLit0 = Vec_IntEntry( vCopies, Gia_ObjFaninId0p(p->pGia, pObj) );
            int iGiaLit1 = Vec_IntEntry( vCopies, Gia_ObjFaninId1p(p->pGia, pObj) );
            iGiaLit0   = Abc_LitNotCond( iGiaLit0, Gia_ObjFaninC0(pObj) );
            iGiaLit1   = Abc_LitNotCond( iGiaLit1, Gia_ObjFaninC1(pObj) );
            iGiaLitNew = Gia_ManHashAnd( p->pGia, iGiaLit0, iGiaLit1 );
            Vec_IntWriteEntry( vCopies, iObj, iGiaLitNew );
        }
        assert( Abc_Lit2Var(p->iLits[n]) == iObj );
    }
    return Abc_LitNotCond( iGiaLitNew, Abc_LitIsCompl(p->iLits[n]) );
}

/**Function*************************************************************

  Synopsis    [Collect nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static inline int Gia_ManGatherSupp_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vCis, Vec_Int_t * vObjs )
{
    int Val0, Val1;
    Gia_Obj_t * pObj;
    if ( Gia_ObjIsTravIdPreviousId(p, iObj) )
        return 1;
    if ( Gia_ObjIsTravIdCurrentId(p, iObj) )
        return 0;
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if ( Gia_ObjIsCi(pObj) )
    {
        Vec_IntPush( vCis, iObj );
        return 0;
    }
    assert( Gia_ObjIsAnd(pObj) );
    Val0 = Gia_ManGatherSupp_rec( p, Gia_ObjFaninId0(pObj, iObj), vCis, vObjs );
    Val1 = Gia_ManGatherSupp_rec( p, Gia_ObjFaninId1(pObj, iObj), vCis, vObjs );
    Vec_IntPush( vObjs, iObj );
    return Val0 || Val1;
}
int Gia_ManGatherSupp( Gia_ManMin_t * p )
{
    int n, Overlap = 0;
    Gia_ManIncrementTravId( p->pGia );
    for ( n = 0; n < 2; n++ )
    {
        Vec_IntClear( p->vCis[n] );
        Vec_IntClear( p->vObjs[n] );
        Gia_ManIncrementTravId( p->pGia );
        Overlap = Gia_ManGatherSupp_rec( p->pGia, Abc_Lit2Var(p->iLits[n]), p->vCis[n], p->vObjs[n] );
        assert( n || !Overlap );
    }
    return Overlap;
}

/**Function*************************************************************

  Synopsis    [Takes a literal and returns a support-minized literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Gia_ManSupportAnd( Gia_ManMin_t * p, int iLit0, int iLit1 )
{
    int iLitNew0, iLitNew1;
    p->iLits[0] = iLit0;
    p->iLits[1] = iLit1;
    if ( iLit0 < 2 || iLit1 < 2 || !Gia_ManGatherSupp(p) || !Gia_ManFindRemoved(p) )
        return Gia_ManHashAnd( p->pGia, iLit0, iLit1 );
    iLitNew0 = Gia_ManRebuildOne( p, 0 );
    iLitNew1 = Gia_ManRebuildOne( p, 1 );
    return Gia_ManHashAnd( p->pGia, iLitNew0, iLitNew1 );
}


#else

Gia_ManMin_t * Gia_ManSuppStart( Gia_Man_t * pGia )              { return NULL; }
int Gia_ManSupportAnd( Gia_ManMin_t * p, int iLit0, int iLit1 )  { return 0;    }
void Gia_ManSuppStop( Gia_ManMin_t * p )                         {}

#endif


/**Function*************************************************************

  Synopsis    [Testbench.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSupportAndTest(pGia *Gia_Man_t) *Gia_Man_t {
    var pMan *Gia_ManMin_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( pGia );
    pNew = Gia_ManStart( Gia_ManObjNum(pGia) );
    pNew.pName = Abc_UtilStrsav( pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( pGia.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(pGia).Value = 0;
    pMan = Gia_ManSuppStart( pNew );
    Gia_ManForEachObj1( pGia, pObj, i )
    {
        if Gia_ObjIsAnd(pObj) {
//            pObj->Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            pObj.Value = Gia_ManSupportAnd( pMan, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else { assert(false, "0"); }

        if i % 10000 == 0 {
            fmt.Printf("%d\n", i); }
    }
    Gia_ManSuppStop( pMan );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(pGia) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}




/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Min2AddClausesMux(p *Gia_Man_t, pNode *Gia_Obj_t, pSat *satoko_t) {
    var fPolarFlip int32;
    var pNodeI, pNodeT, pNodeE *Gia_Obj_t;
    var pLits [4]int32;
    var RetValue int32;
    var VarF int32;
    var VarI int32;
    var VarT int32;
    var VarE int32;
    var fCompT int32;
    var fCompE int32;

    assert(!Gia_IsComplement( pNode ), "!Gia_IsComplement( pNode )");
    assert(pNode.fMark0, "pNode.fMark0");
    // get nodes (I = if, T = then, E = else)
    pNodeI = Gia_ObjRecognizeMux( pNode, &pNodeT, &pNodeE );
    // get the variable numbers
    VarF = Gia_Min2ObjSatId(p, pNode);
    VarI = Gia_Min2ObjSatId(p, pNodeI);
    VarT = Gia_Min2ObjSatId(p, Gia_Regular(pNodeT));
    VarE = Gia_Min2ObjSatId(p, Gia_Regular(pNodeE));
    // get the complementation flags
    fCompT = Gia_IsComplement(pNodeT);
    fCompE = Gia_IsComplement(pNodeE);

    // f = ITE(i, t, e)

    // i' + t' + f
    // i' + t  + f'
    // i  + e' + f
    // i  + e  + f'

    // create four clauses
    pLits[0] = Abc_Var2Lit(VarI, 1);
    pLits[1] = Abc_Var2Lit(VarT, 1^fCompT);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 1);
    pLits[1] = Abc_Var2Lit(VarT, 0^fCompT);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 0^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");

    // two additional clauses
    // t' & e' -> f'
    // t  & e  -> f 

    // t  + e   + f'
    // t' + e'  + f 

    if VarT == VarE {
//        assert( fCompT == !fCompE );
        return;
    }

    pLits[0] = Abc_Var2Lit(VarT, 0^fCompT);
    pLits[1] = Abc_Var2Lit(VarE, 0^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarT, 1^fCompT);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
}
func Gia_Min2AddClausesSuper(p *Gia_Man_t, pNode *Gia_Obj_t, vSuper *Vec_Ptr_t, pSat *satoko_t) {
    var fPolarFlip int32;
    var pFanin *Gia_Obj_t;
    var pLits *int32;
    var nLits int32;
    var RetValue int32;
    var i int32;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert(Gia_ObjIsAnd( pNode ), "Gia_ObjIsAnd( pNode )");
    // create storage for literals
    nLits = Vec_PtrSize(vSuper) + 1;
    pLits = ABC_ALLOC( int, nLits );
    // suppose AND-gate is A & B = C
    // add !A => !C   or   A + !C
    Vec_PtrForEachEntry( Gia_Obj_t *, vSuper, pFanin, i )
    {
        pLits[0] = Abc_Var2Lit(Gia_Min2ObjSatId(p, Gia_Regular(pFanin)), Gia_IsComplement(pFanin));
        pLits[1] = Abc_Var2Lit(Gia_Min2ObjSatId(p, pNode), 1);
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
            if pNode.fPhase {                pLits[1] = Abc_LitNot( pLits[1] ); }
        }
        RetValue = satoko_add_clause( pSat, pLits, 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    Vec_PtrForEachEntry( Gia_Obj_t *, vSuper, pFanin, i )
    {
        pLits[i] = Abc_Var2Lit(Gia_Min2ObjSatId(p, Gia_Regular(pFanin)), !Gia_IsComplement(pFanin));
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[i] = Abc_LitNot( pLits[i] ); }
        }
    }
    pLits[nLits-1] = Abc_Var2Lit(Gia_Min2ObjSatId(p, pNode), 0);
    if fPolarFlip != 0 {
        if pNode.fPhase {  pLits[nLits-1] = Abc_LitNot( pLits[nLits-1] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, nLits );
    assert((RetValue != 0), "RetValue");
    pLits = nil;
}

/**Function*************************************************************

  Synopsis    [Adds clauses and returns CNF variable of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Min2CollectSuper_rec(pObj *Gia_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) || 
         (!(fFirst != 0) && Gia_ObjValue(pObj) > 1) || 
         ((fUseMuxes != 0) && pObj.fMark0) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Gia_Min2CollectSuper_rec( Gia_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Gia_Min2CollectSuper_rec( Gia_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}
func Gia_Min2CollectSuper(pObj *Gia_Obj_t, fUseMuxes int32, vSuper *Vec_Ptr_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsCi(pObj), "!Gia_ObjIsCi(pObj)");
    Vec_PtrClear( vSuper );
    Gia_Min2CollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
}
func Gia_Min2ObjAddToFrontier(p *Gia_Man_t, pObj *Gia_Obj_t, vFrontier *Vec_Ptr_t, pSat *satoko_t, vSatVars *Vec_Int_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsConst0(pObj), "!Gia_ObjIsConst0(pObj)");
    if Gia_Min2ObjSatId(p, pObj) >= 0 {
        return; }
    assert(Gia_Min2ObjSatId(p, pObj) == -1, "Gia_Min2ObjSatId(p, pObj) == -1");
    Vec_IntPush( vSatVars, Gia_ObjId(p, pObj) );
    Gia_Min2ObjSetSatId( p, pObj, satoko_add_variable(pSat, 0) );
    if Gia_ObjIsAnd(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}
func Gia_Min2ObjGetCnfVar(p *Gia_Man2Min_t, iObj int32) int32 { 
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var pNode, pFanin *Gia_Obj_t;
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    if Gia_Min2ObjSatId(p.pGia, pObj) >= 0 {
        return Gia_Min2ObjSatId(p.pGia, pObj); }
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( p.vSatVars, iObj );
        return Gia_Min2ObjSetSatId( p.pGia, pObj, satoko_add_variable(p.pSat, 0) );
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // start the frontier
    Vec_PtrClear( p.vFrontier );
    Gia_Min2ObjAddToFrontier( p.pGia, pObj, p.vFrontier, p.pSat, p.vSatVars );
    // explore nodes in the frontier
    Vec_PtrForEachEntry( Gia_Obj_t *, p.vFrontier, pNode, i )
    {
        assert(Gia_Min2ObjSatId(p.pGia,pNode) >= 0, "Gia_Min2ObjSatId(p.pGia,pNode) >= 0");
        if (fUseMuxes != 0) && pNode.fMark0 {
            Vec_PtrClear( p.vFanins );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin1(pNode) ) );
            Vec_PtrForEachEntry( Gia_Obj_t *, p.vFanins, pFanin, k )
                Gia_Min2ObjAddToFrontier( p.pGia, Gia_Regular(pFanin), p.vFrontier, p.pSat, p.vSatVars );
            Gia_Min2AddClausesMux( p.pGia, pNode, p.pSat );
        
        } else {
            Gia_Min2CollectSuper( pNode, fUseMuxes, p.vFanins );
            Vec_PtrForEachEntry( Gia_Obj_t *, p.vFanins, pFanin, k )
                Gia_Min2ObjAddToFrontier( p.pGia, Gia_Regular(pFanin), p.vFrontier, p.pSat, p.vSatVars );
            Gia_Min2AddClausesSuper( p.pGia, pNode, p.vFanins, p.pSat );
        }
        assert(Vec_PtrSize(p.vFanins) > 1, "Vec_PtrSize(p.vFanins) > 1");
    }
    return Gia_Min2ObjSatId(p.pGia,pObj);
}

/**Function*************************************************************

  Synopsis    [Returns 0 if the node is not a constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collect nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Takes a literal and returns a support-minized literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

