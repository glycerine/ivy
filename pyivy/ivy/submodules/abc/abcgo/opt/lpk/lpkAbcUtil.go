package lpk

/**CFile****************************************************************

  FileName    [lpkAbcUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    [Procedures working on decomposed functions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkAbcUtil.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_FunAlloc(nVars int32) *Lpk_Fun_t {
    var p *Lpk_Fun_t;
    p = (Lpk_Fun_t *)ABC_ALLOC( char, sizeof(Lpk_Fun_t) + sizeof(unsigned) * Kit_TruthWordNum(nVars) * 3 );
    memset( p, 0, sizeof(Lpk_Fun_t) );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes the function]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_FunFree(p *Lpk_Fun_t) {
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Creates the starting function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_FunCreate(pNtk *Abc_Ntk_t, vLeaves *Vec_Ptr_t, pTruth *uint32, nLutK int32, AreaLim int32, DelayLim int32) *Lpk_Fun_t {
    var p *Lpk_Fun_t;
    var pNode *Abc_Obj_t;
    var i int32;
    p = Lpk_FunAlloc( Vec_PtrSize(vLeaves) );
    p.Id = Vec_PtrSize(vLeaves);
    p.vNodes = vLeaves;
    p.nVars = Vec_PtrSize(vLeaves);
    p.nLutK = nLutK;
    p.nAreaLim = AreaLim;
    p.nDelayLim = DelayLim;
    p.uSupp = Kit_TruthSupport( pTruth, p.nVars );
    Kit_TruthCopy( Lpk_FunTruth(p,0), pTruth, p.nVars );
    Vec_PtrForEachEntry( Abc_Obj_t *, vLeaves, pNode, i )
    {
        p.pFanins[i] = i;
        p.pDelays[i] = pNode.Level;
    }
    Vec_PtrPush( p.vNodes, p );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the new function with the given truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_FunDup(p *Lpk_Fun_t, pTruth *uint32) *Lpk_Fun_t {
    var pNew *Lpk_Fun_t;
    pNew = Lpk_FunAlloc( p.nVars );
    pNew.Id = Vec_PtrSize(p.vNodes);
    pNew.vNodes = p.vNodes;
    pNew.nVars = p.nVars;
    pNew.nLutK = p.nLutK;
    pNew.nAreaLim = p.nAreaLim;
    pNew.nDelayLim = p.nDelayLim;
    pNew.uSupp = Kit_TruthSupport( pTruth, p.nVars );
    Kit_TruthCopy( Lpk_FunTruth(pNew,0), pTruth, p.nVars );
    memcpy( pNew.pFanins, p.pFanins, 16 );
    memcpy( pNew.pDelays, p.pDelays, sizeof(int)*16 );
    Vec_PtrPush( p.vNodes, pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Minimizes support of the function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_FunSuppMinimize(p *Lpk_Fun_t) int32 {
    var i, k, nVarsNew int32;
    // compress the truth table
    if p.uSupp == Kit_BitMask(p.nVars) {
        return 0; }
    // invalidate support info
    p.fSupports = 0;
//Extra_PrintBinary( stdout, &p->uSupp, p->nVars ); printf( "\n" );
    // minimize support
    nVarsNew = Kit_WordCountOnes(p.uSupp);
    Kit_TruthShrink( Lpk_FunTruth(p, 1), Lpk_FunTruth(p, 0), nVarsNew, p.nVars, p.uSupp, 1 );
    k = 0;
    // inlining of macro Lpk_SuppForEachVar
    for  i = 0; i < 16; i++  {        if !(p.uSupp & (1<<i)) {} else {
        p.pFanins[k] = p.pFanins[i];
        p.pDelays[k] = p.pDelays[i];
/*
        if ( p->fSupports )
        {
            p->puSupps[2*k+0] = p->puSupps[2*i+0];
            p->puSupps[2*k+1] = p->puSupps[2*i+1];
        }
*/
        k++;
    }
    }
    assert(k == nVarsNew, "k == nVarsNew");
    p.nVars = k;
    p.uSupp = Kit_BitMask(p.nVars);
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes cofactors w.r.t. each variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_FunComputeCofSupps(p *Lpk_Fun_t) {
    var pTruth *uint32   = Lpk_FunTruth( p, 0 );
    var pTruth0 *uint32  = Lpk_FunTruth( p, 1 );
    var pTruth1 *uint32  = Lpk_FunTruth( p, 2 );
    var Var int32;
    assert(p.fSupports == 0, "p.fSupports == 0");
//    Lpk_SuppForEachVar( p->uSupp, Var )
    for  Var = 0; Var < int32(p.nVars); Var++  {
        Kit_TruthCofactor0New( pTruth0, pTruth, p.nVars, Var );
        Kit_TruthCofactor1New( pTruth1, pTruth, p.nVars, Var );
        p.puSupps[2*Var+0] = Kit_TruthSupport( pTruth0, p.nVars );
        p.puSupps[2*Var+1] = Kit_TruthSupport( pTruth1, p.nVars );
    }
    p.fSupports = 1;
}

/**Function*************************************************************

  Synopsis    [Get the delay of the bound set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_SuppDelay(uSupp uint32, pDelays *int32) int32 {
    var Delay, Var int32;
    Delay = 0;
    // inlining of macro Lpk_SuppForEachVar
    for  Var = 0; Var < 16; Var++  {        if !(uSupp & (1<<Var)) {} else {
        Delay = Abc_MaxInt( Delay, pDelays[Var] ); }
    }
    return Delay + 1;
}

/**Function*************************************************************

  Synopsis    [Converts support into variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lpk_SuppToVars(uBoundSet uint32, pVars string) int32 {
    var i, nVars int32;
    // inlining of macro Lpk_SuppForEachVar
    for  i = 0; i < 16; i++  {        if !(uBoundSet & (1<<i)) {} else {
        pVars[nVars++] = i; }
    }
    return nVars;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

