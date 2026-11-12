package fxu

/**CFile****************************************************************

  FileName    [fxuSingle.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures to compute the set of single-cube divisors.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuSingle.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"
//#include "misc/vec/vec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Fxu_MatrixComputeSinglesOneCollect( Fxu_Matrix * p, Fxu_Var * pVar, Vec_Ptr_t * vSingles );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes and adds all single-cube divisors to storage.]

  Description [This procedure should be called once when the matrix is
  already contructed before the process of logic extraction begins..]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixComputeSingles(p *Fxu_Matrix, fUse0 int32, nSingleMax int32) {
    var pVar *Fxu_Var;
    var vSingles *Vec_Ptr_t;
    var i, k int32;
    // set the weight limit
    p.nWeightLimit = 1 - fUse0;
    // iterate through columns in the matrix and collect single-cube divisors
    vSingles = Vec_PtrAlloc( 10000 );
    // inlining of macro Fxu_MatrixForEachVariable
    for  pVar = (p).lVars.pHead;          (pVar != nil);          pVar = pVar.pNext  {
        Fxu_MatrixComputeSinglesOneCollect( p, pVar, vSingles );
    }
    p.nSingleTotal = Vec_PtrSize(vSingles) / 3;
    // check if divisors should be filtered
    if Vec_PtrSize(vSingles) > nSingleMax {
        var pWeigtCounts *int32;
        var nDivCount int32;
        var Weight int32;
        var i int32;
        var c int32;;
        assert(Vec_PtrSize(vSingles) % 3 == 0, "Vec_PtrSize(vSingles) % 3 == 0");
        // count how many divisors have the given weight
        // inlining of macro ABC_ALLOC
        pWeigtCounts = ((int *) malloc(sizeof(int) * (size_t)(1000)));
        memset( pWeigtCounts, 0, sizeof(int) * 1000 );
        for  i = 2; i < Vec_PtrSize(vSingles); i += 3  {
            Weight = int32((ABC_PTRUINT_T))Vec_PtrEntry(vSingles, i);
            if Weight >= 999 {
                pWeigtCounts[999]++;
            } else {
                pWeigtCounts[Weight]++; }
        }
        // select the bound on the weight (above this bound, singles will be included)
        nDivCount = 0;
        for  c = 999; c >= 0; c--  {
            nDivCount += pWeigtCounts[c];
            if nDivCount >= nSingleMax {
                break; }
        }
        pWeigtCounts = nil;
        // collect singles with the given costs
        k = 0;
        for  i = 2; i < Vec_PtrSize(vSingles); i += 3  {
            Weight = int32((ABC_PTRUINT_T))Vec_PtrEntry(vSingles, i);
            if Weight < c {
                continue; }
            Vec_PtrWriteEntry( vSingles, k++, Vec_PtrEntry(vSingles, i-2) );
            Vec_PtrWriteEntry( vSingles, k++, Vec_PtrEntry(vSingles, i-1) );
            Vec_PtrWriteEntry( vSingles, k++, Vec_PtrEntry(vSingles, i) );
            if k/3 == nSingleMax {
                break; }
        }
        Vec_PtrShrink( vSingles, k );
        // adjust the weight limit
        p.nWeightLimit = c;
    }
    // collect the selected divisors
    assert(Vec_PtrSize(vSingles) % 3 == 0, "Vec_PtrSize(vSingles) % 3 == 0");
    for  i = 0; i < Vec_PtrSize(vSingles); i += 3  {
        Fxu_MatrixAddSingle( p, 
            (Fxu_Var *)Vec_PtrEntry(vSingles,i), 
            (Fxu_Var *)Vec_PtrEntry(vSingles,i+1), 
            int32((ABC_PTRUINT_T))Vec_PtrEntry(vSingles,i+2) );
    }
    Vec_PtrFree( vSingles );
}

/**Function*************************************************************

  Synopsis    [Adds the single-cube divisors associated with a new column.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixComputeSinglesOneCollect(p *Fxu_Matrix, pVar *Fxu_Var, vSingles *Vec_Ptr_t) {
    var pLitV, pLitH *Fxu_Lit;
    var pVar2 *Fxu_Var;
    var Coin int32;
    var WeightCur int32;

    // start collecting the affected vars
    // inlining of macro Fxu_MatrixRingVarsStart
    ((p).ppTailVars  = &((p).pOrderVars));
((p).pOrderVars  = nil);;
    // go through all the literals of this variable
    for  pLitV = pVar.lLits.pHead; pLitV; pLitV = pLitV.pVNext  {
        // for this literal, go through all the horizontal literals
        for  pLitH = pLitV.pHPrev; (pLitH != nil); pLitH = pLitH.pHPrev  {
            // get another variable
            pVar2 = pLitH.pVar;
            // skip the var if it is already used
            if pVar2.pOrder {
                continue; }
            // skip the var if it belongs to the same node
//            if ( pValue2Node[pVar->iVar] == pValue2Node[pVar2->iVar] )
//                continue;
            // collect the var
            // inlining of macro Fxu_MatrixRingVarsAdd
            (*((p).ppTailVars ) = pVar2 );
((p).ppTailVars  = &(pVar2).pOrder );
((pVar2).pOrder  = (Fxu_Var *)1);;
        }
    }
    // stop collecting the selected vars
    // inlining of macro Fxu_MatrixRingVarsStop
    ;

    // iterate through the selected vars
    // inlining of macro Fxu_MatrixForEachVarInRing
    if (p).pOrderVars {    for  pVar2 = (p).pOrderVars;          pVar2 != (Fxu_Var *)1;          pVar2 = pVar2.pOrder  {
        // count the coincidence
        Coin = Fxu_SingleCountCoincidence( p, pVar2, pVar );
        assert(Coin > 0, "Coin > 0");
        // get the new weight
        WeightCur = Coin - 2;
        // peformance fix (August 24, 2007)
        if WeightCur >= p.nWeightLimit {
            Vec_PtrPush( vSingles, pVar2 );
            Vec_PtrPush( vSingles, pVar );
            Vec_PtrPush( vSingles, (void *)(ABC_PTRUINT_T)WeightCur );
        }
    } }

    // unmark the vars
    Fxu_MatrixRingVarsUnmark( p );
}

/**Function*************************************************************

  Synopsis    [Adds the single-cube divisors associated with a new column.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixComputeSinglesOne(p *Fxu_Matrix, pVar *Fxu_Var) {
    var pLitV, pLitH *Fxu_Lit;
    var pVar2 *Fxu_Var;
    var Coin int32;
    var WeightCur int32;

    // start collecting the affected vars
    // inlining of macro Fxu_MatrixRingVarsStart
    ((p).ppTailVars  = &((p).pOrderVars));
((p).pOrderVars  = nil);;
    // go through all the literals of this variable
    for  pLitV = pVar.lLits.pHead; pLitV; pLitV = pLitV.pVNext  {
        // for this literal, go through all the horizontal literals
        for  pLitH = pLitV.pHPrev; (pLitH != nil); pLitH = pLitH.pHPrev  {
            // get another variable
            pVar2 = pLitH.pVar;
            // skip the var if it is already used
            if pVar2.pOrder {
                continue; }
            // skip the var if it belongs to the same node
//            if ( pValue2Node[pVar->iVar] == pValue2Node[pVar2->iVar] )
//                continue;
            // collect the var
            // inlining of macro Fxu_MatrixRingVarsAdd
            (*((p).ppTailVars ) = pVar2 );
((p).ppTailVars  = &(pVar2).pOrder );
((pVar2).pOrder  = (Fxu_Var *)1);;
        }
    }
    // stop collecting the selected vars
    // inlining of macro Fxu_MatrixRingVarsStop
    ;

    // iterate through the selected vars
    // inlining of macro Fxu_MatrixForEachVarInRing
    if (p).pOrderVars {    for  pVar2 = (p).pOrderVars;          pVar2 != (Fxu_Var *)1;          pVar2 = pVar2.pOrder  {
        // count the coincidence
        Coin = Fxu_SingleCountCoincidence( p, pVar2, pVar );
        assert(Coin > 0, "Coin > 0");
        // get the new weight
        WeightCur = Coin - 2;
        // peformance fix (August 24, 2007)
//        if ( WeightCur >= 0 )
//        Fxu_MatrixAddSingle( p, pVar2, pVar, WeightCur );
        if WeightCur >= p.nWeightLimit {
            Fxu_MatrixAddSingle( p, pVar2, pVar, WeightCur ); }
    } }
    // unmark the vars
    Fxu_MatrixRingVarsUnmark( p );
}

/**Function*************************************************************

  Synopsis    [Computes the coincidence count of two columns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_SingleCountCoincidence(p *Fxu_Matrix, pVar1 *Fxu_Var, pVar2 *Fxu_Var) int32 {
    var pLit1, pLit2 *Fxu_Lit;
    var Result int32;

    // compute the coincidence count
    Result = 0;
    pLit1  = pVar1.lLits.pHead;
    pLit2  = pVar2.lLits.pHead;
    for {
        if pLit1 && (pLit2 != nil) {
            if pLit1.pCube.pVar.iVar == pLit2.pCube.pVar.iVar { // the variables are the same
                if pLit1.iCube == pLit2.iCube { // the literals are the same
                    pLit1 = pLit1.pVNext;
                    pLit2 = pLit2.pVNext;
                    // add this literal to the coincidence
                    Result++;
                
                } else if pLit1.iCube < pLit2.iCube {
                    pLit1 = pLit1.pVNext;
                } else {
                    pLit2 = pLit2.pVNext; }
            
            } else if pLit1.pCube.pVar.iVar < pLit2.pCube.pVar.iVar {
                pLit1 = pLit1.pVNext;
            } else {
                pLit2 = pLit2.pVNext; }
        
        } else if pLit1 && !(pLit2 != nil) {
            pLit1 = pLit1.pVNext;
        } else if !pLit1 && (pLit2 != nil) {
            pLit2 = pLit2.pVNext;
        } else {
            break; }
    }
    return Result;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

