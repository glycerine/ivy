package fxu

/**CFile****************************************************************

  FileName    [fxuUpdate.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Updating the sparse matrix when divisors are accepted.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuUpdate.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Fxu_UpdateDoublePairs( Fxu_Matrix * p, Fxu_Double * pDouble, Fxu_Var * pVar );

// static void Fxu_UpdateMatrixDoubleCreateCubes( Fxu_Matrix * p, Fxu_Cube * pCube1, Fxu_Cube * pCube2, Fxu_Double * pDiv );

// static void Fxu_UpdateMatrixDoubleClean( Fxu_Matrix * p, Fxu_Cube * pCubeUse, Fxu_Cube * pCubeRem );

// static void Fxu_UpdateMatrixSingleClean( Fxu_Matrix * p, Fxu_Var * pVar1, Fxu_Var * pVar2, Fxu_Var * pVarNew );


// static void Fxu_UpdateCreateNewVars( Fxu_Matrix * p, Fxu_Var ** ppVarC, Fxu_Var ** ppVarD, int nCubes );

// static int  Fxu_UpdatePairCompare( Fxu_Pair ** ppP1, Fxu_Pair ** ppP2 );

// static void Fxu_UpdatePairsSort( Fxu_Matrix * p, Fxu_Double * pDouble );


// static void Fxu_UpdateCleanOldDoubles( Fxu_Matrix * p, Fxu_Double * pDiv, Fxu_Cube * pCube );

// static void Fxu_UpdateAddNewDoubles( Fxu_Matrix * p, Fxu_Cube * pCube );

// static void Fxu_UpdateCleanOldSingles( Fxu_Matrix * p );

// static void Fxu_UpdateAddNewSingles( Fxu_Matrix * p, Fxu_Var * pVar );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Updates the matrix after selecting two divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_Update(p *Fxu_Matrix, pSingle *Fxu_Single, pDouble *Fxu_Double) {
    var pCube, pCubeNew *Fxu_Cube;
    var pVarC, pVarD *Fxu_Var;
    var pVar1, pVar2 *Fxu_Var;

    // consider trivial cases
    if pSingle == nil {
        assert(pDouble.Weight == Fxu_HeapDoubleReadMaxWeight( p.pHeapDouble ), "pDouble.Weight == Fxu_HeapDoubleReadMaxWeight( p.pHeapDouble )");
        Fxu_UpdateDouble( p );
        return;
    }
    if pDouble == nil {
        assert(pSingle.Weight == Fxu_HeapSingleReadMaxWeight( p.pHeapSingle ), "pSingle.Weight == Fxu_HeapSingleReadMaxWeight( p.pHeapSingle )");
        Fxu_UpdateSingle( p );
        return;
    }

    // get the variables of the single
    pVar1 = pSingle.pVar1;
    pVar2 = pSingle.pVar2;

    // remove the best double from the heap
    Fxu_HeapDoubleDelete( p.pHeapDouble, pDouble );
    // remove the best divisor from the table
    Fxu_ListTableDelDivisor( p, pDouble );

    // create two new columns (vars)
    Fxu_UpdateCreateNewVars( p, &pVarC, &pVarD, 1 );
    // create one new row (cube)
    pCubeNew = Fxu_MatrixAddCube( p, pVarD, 0 );
    pCubeNew.pFirst = pCubeNew;
    // set the first cube of the positive var
    pVarD.pFirst = pCubeNew;

    // start collecting the affected vars and cubes
    // inlining of macro Fxu_MatrixRingCubesStart
    ((p).ppTailCubes = &((p).pOrderCubes));
((p).pOrderCubes = nil);;
    // inlining of macro Fxu_MatrixRingVarsStart
    ((p).ppTailVars  = &((p).pOrderVars));
((p).pOrderVars  = nil);;
    // add the vars
    // inlining of macro Fxu_MatrixRingVarsAdd
    (*((p).ppTailVars ) = pVar1 );
((p).ppTailVars  = &(pVar1).pOrder );
((pVar1).pOrder  = (Fxu_Var *)1);;
    // inlining of macro Fxu_MatrixRingVarsAdd
    (*((p).ppTailVars ) = pVar2 );
((p).ppTailVars  = &(pVar2).pOrder );
((pVar2).pOrder  = (Fxu_Var *)1);;
    // remove the literals and collect the affected cubes
    // remove the divisors associated with this cube
       // add to the affected cube the literal corresponding to the new column
    Fxu_UpdateMatrixSingleClean( p, pVar1, pVar2, pVarD );
    // replace each two cubes of the pair by one new cube
    // the new cube contains the base and the new literal
    Fxu_UpdateDoublePairs( p, pDouble, pVarC );
    // stop collecting the affected vars and cubes
    // inlining of macro Fxu_MatrixRingCubesStop
    ;
    // inlining of macro Fxu_MatrixRingVarsStop
    ;

    // add the literals to the new cube
    assert(pVar1.iVar < pVar2.iVar, "pVar1.iVar < pVar2.iVar");
    assert(Fxu_SingleCountCoincidence( p, pVar1, pVar2 ) == 0, "Fxu_SingleCountCoincidence( p, pVar1, pVar2 ) == 0");
    Fxu_MatrixAddLiteral( p, pCubeNew, pVar1 );
    Fxu_MatrixAddLiteral( p, pCubeNew, pVar2 );

    // create new doubles; we cannot add them in the same loop
    // because we first have to create *all* new cubes for each node
    // inlining of macro Fxu_MatrixForEachCubeInRing
    if (p).pOrderCubes {    for  pCube = (p).pOrderCubes;          pCube != (Fxu_Cube *)1;          pCube = pCube.pOrder  {
           Fxu_UpdateAddNewDoubles( p, pCube );
    } }
    // update the singles after removing some literals
    Fxu_UpdateCleanOldSingles( p );

    // undo the temporary rings with cubes and vars
    Fxu_MatrixRingCubesUnmark( p );
    Fxu_MatrixRingVarsUnmark( p );
    // we should undo the rings before creating new singles

    // create new singles
    Fxu_UpdateAddNewSingles( p, pVarC );
    Fxu_UpdateAddNewSingles( p, pVarD );

    // recycle the divisor
    MEM_FREE_FXU( p, Fxu_Double, 1, pDouble );
    p.nDivs3++;
}

/**Function*************************************************************

  Synopsis    [Updates after accepting single cube divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateSingle(p *Fxu_Matrix) {
    var pSingle *Fxu_Single;
    var pCube, pCubeNew *Fxu_Cube;
    var pVarC, pVarD *Fxu_Var;
    var pVar1, pVar2 *Fxu_Var;

    // read the best divisor from the heap
    pSingle = Fxu_HeapSingleReadMax( p.pHeapSingle );
    // get the variables of this single-cube divisor
    pVar1 = pSingle.pVar1;
    pVar2 = pSingle.pVar2;

    // create two new columns (vars)
    Fxu_UpdateCreateNewVars( p, &pVarC, &pVarD, 1 );
    // create one new row (cube)
    pCubeNew = Fxu_MatrixAddCube( p, pVarD, 0 );
    pCubeNew.pFirst = pCubeNew;
    // set the first cube
    pVarD.pFirst = pCubeNew;

    // start collecting the affected vars and cubes
    // inlining of macro Fxu_MatrixRingCubesStart
    ((p).ppTailCubes = &((p).pOrderCubes));
((p).pOrderCubes = nil);;
    // inlining of macro Fxu_MatrixRingVarsStart
    ((p).ppTailVars  = &((p).pOrderVars));
((p).pOrderVars  = nil);;
    // add the vars
    // inlining of macro Fxu_MatrixRingVarsAdd
    (*((p).ppTailVars ) = pVar1 );
((p).ppTailVars  = &(pVar1).pOrder );
((pVar1).pOrder  = (Fxu_Var *)1);;
    // inlining of macro Fxu_MatrixRingVarsAdd
    (*((p).ppTailVars ) = pVar2 );
((p).ppTailVars  = &(pVar2).pOrder );
((pVar2).pOrder  = (Fxu_Var *)1);;
    // remove the literals and collect the affected cubes
    // remove the divisors associated with this cube
       // add to the affected cube the literal corresponding to the new column
    Fxu_UpdateMatrixSingleClean( p, pVar1, pVar2, pVarD );
    // stop collecting the affected vars and cubes
    // inlining of macro Fxu_MatrixRingCubesStop
    ;
    // inlining of macro Fxu_MatrixRingVarsStop
    ;

    // add the literals to the new cube
    assert(pVar1.iVar < pVar2.iVar, "pVar1.iVar < pVar2.iVar");
    assert(Fxu_SingleCountCoincidence( p, pVar1, pVar2 ) == 0, "Fxu_SingleCountCoincidence( p, pVar1, pVar2 ) == 0");
    Fxu_MatrixAddLiteral( p, pCubeNew, pVar1 );
    Fxu_MatrixAddLiteral( p, pCubeNew, pVar2 );

    // create new doubles; we cannot add them in the same loop
    // because we first have to create *all* new cubes for each node
    // inlining of macro Fxu_MatrixForEachCubeInRing
    if (p).pOrderCubes {    for  pCube = (p).pOrderCubes;          pCube != (Fxu_Cube *)1;          pCube = pCube.pOrder  {
           Fxu_UpdateAddNewDoubles( p, pCube );
    } }
    // update the singles after removing some literals
    Fxu_UpdateCleanOldSingles( p );
    // we should undo the rings before creating new singles

    // unmark the cubes
    Fxu_MatrixRingCubesUnmark( p );
    Fxu_MatrixRingVarsUnmark( p );

    // create new singles
    Fxu_UpdateAddNewSingles( p, pVarC );
    Fxu_UpdateAddNewSingles( p, pVarD );
    p.nDivs1++;
}

/**Function*************************************************************

  Synopsis    [Updates the matrix after accepting a double cube divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateDouble(p *Fxu_Matrix) {
    var pDiv *Fxu_Double;
    var pCube, pCubeNew1, pCubeNew2 *Fxu_Cube;
    var pVarC, pVarD *Fxu_Var;

    // remove the best divisor from the heap
    pDiv = Fxu_HeapDoubleGetMax( p.pHeapDouble );
    // remove the best divisor from the table
    Fxu_ListTableDelDivisor( p, pDiv );

    // create two new columns (vars)
    Fxu_UpdateCreateNewVars( p, &pVarC, &pVarD, 2 );
    // create two new rows (cubes)
    pCubeNew1 = Fxu_MatrixAddCube( p, pVarD, 0 );
    pCubeNew1.pFirst = pCubeNew1;
    pCubeNew2 = Fxu_MatrixAddCube( p, pVarD, 1 );
    pCubeNew2.pFirst = pCubeNew1;
    // set the first cube
    pVarD.pFirst = pCubeNew1;

    // add the literals to the new cubes
    Fxu_UpdateMatrixDoubleCreateCubes( p, pCubeNew1, pCubeNew2, pDiv );

    // start collecting the affected cubes and vars
    // inlining of macro Fxu_MatrixRingCubesStart
    ((p).ppTailCubes = &((p).pOrderCubes));
((p).pOrderCubes = nil);;
    // inlining of macro Fxu_MatrixRingVarsStart
    ((p).ppTailVars  = &((p).pOrderVars));
((p).pOrderVars  = nil);;
    // replace each two cubes of the pair by one new cube
    // the new cube contains the base and the new literal
    Fxu_UpdateDoublePairs( p, pDiv, pVarD );
    // stop collecting the affected cubes and vars
    // inlining of macro Fxu_MatrixRingCubesStop
    ;
    // inlining of macro Fxu_MatrixRingVarsStop
    ;
    
    // create new doubles; we cannot add them in the same loop
    // because we first have to create *all* new cubes for each node
    // inlining of macro Fxu_MatrixForEachCubeInRing
    if (p).pOrderCubes {    for  pCube = (p).pOrderCubes;          pCube != (Fxu_Cube *)1;          pCube = pCube.pOrder  {
           Fxu_UpdateAddNewDoubles( p, pCube );
    } }
    // update the singles after removing some literals
    Fxu_UpdateCleanOldSingles( p );

    // undo the temporary rings with cubes and vars
    Fxu_MatrixRingCubesUnmark( p );
    Fxu_MatrixRingVarsUnmark( p );
    // we should undo the rings before creating new singles

    // create new singles
    Fxu_UpdateAddNewSingles( p, pVarC );
    Fxu_UpdateAddNewSingles( p, pVarD );

    // recycle the divisor
    MEM_FREE_FXU( p, Fxu_Double, 1, pDiv );
    p.nDivs2++;
}

/**Function*************************************************************

  Synopsis    [Update the pairs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateDoublePairs(p *Fxu_Matrix, pDouble *Fxu_Double, pVar *Fxu_Var) {
    var pPair *Fxu_Pair;
    var pCubeUse, pCubeRem *Fxu_Cube;
    var i int32;

    // collect and sort the pairs
    Fxu_UpdatePairsSort( p, pDouble );
//    for ( i = 0; i < p->nPairsTemp; i++ )
    for  i = 0; i < p.vPairs.nSize; i++  {
        // get the pair
//        pPair = p->pPairsTemp[i];
        pPair = (Fxu_Pair *)p.vPairs.pArray[i];
        // out of the two cubes, select the one which comes earlier
        // inlining of macro Fxu_PairMinCube
        
        if (pPair).iCube1 < (pPair).iCube2 {
            pCubeUse = (pPair).pCube1;
        } else {
            pCubeUse = (pPair).pCube2;
        }
        // inlining of macro Fxu_PairMaxCube
        
        if (pPair).iCube1 > (pPair).iCube2 {
            pCubeRem = (pPair).pCube1;
        } else {
            pCubeRem = (pPair).pCube2;
        }
        // collect the affected cube
        assert(pCubeUse.pOrder == nil, "pCubeUse.pOrder == nil");
        // inlining of macro Fxu_MatrixRingCubesAdd
        (*((p).ppTailCubes) = pCubeUse);
((p).ppTailCubes = &(pCubeUse).pOrder);
((pCubeUse).pOrder = (Fxu_Cube *)1);;

        // remove some literals from pCubeUse and all literals from pCubeRem
        Fxu_UpdateMatrixDoubleClean( p, pCubeUse, pCubeRem );
        // add a literal that depends on the new variable
        Fxu_MatrixAddLiteral( p, pCubeUse, pVar );    
        // check the literal count
        assert(pCubeUse.lLits.nItems == pPair.nBase + 1, "pCubeUse.lLits.nItems == pPair.nBase + 1");
        assert(pCubeRem.lLits.nItems == 0, "pCubeRem.lLits.nItems == 0");

        // update the divisors by removing useless pairs
        Fxu_UpdateCleanOldDoubles( p, pDouble, pCubeUse );
        Fxu_UpdateCleanOldDoubles( p, pDouble, pCubeRem );
        // remove the pair
        MEM_FREE_FXU( p, Fxu_Pair, 1, pPair );
    }
    p.vPairs.nSize = 0;
}

/**Function*************************************************************

  Synopsis    [Add two cubes corresponding to the given double-cube divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateMatrixDoubleCreateCubes(p *Fxu_Matrix, pCube1 *Fxu_Cube, pCube2 *Fxu_Cube, pDiv *Fxu_Double) {
    var pLit1, pLit2 *Fxu_Lit;
    var pPair *Fxu_Pair;
    var nBase, nLits1, nLits2 int32;

    // fill in the SOP and copy the fanins
    nBase = nLits1 = nLits2 = 0;
    pPair = pDiv.lPairs.pHead;
    pLit1 = pPair.pCube1.lLits.pHead;
    pLit2 = pPair.pCube2.lLits.pHead;
    for {
        if pLit1 && (pLit2 != nil) {
            if pLit1.iVar == pLit2.iVar { // skip the cube free part
                pLit1 = pLit1.pHNext;
                pLit2 = pLit2.pHNext;
                nBase++;
            
            } else if pLit1.iVar < pLit2.iVar {    // add literal to the first cube
                Fxu_MatrixAddLiteral( p, pCube1, pLit1.pVar );
                // move to the next literal in this cube
                pLit1 = pLit1.pHNext;
                nLits1++;
            
            } else {    // add literal to the second cube
                Fxu_MatrixAddLiteral( p, pCube2, pLit2.pVar );
                // move to the next literal in this cube
                pLit2 = pLit2.pHNext;
                nLits2++;
            }
        
        } else if pLit1 && !(pLit2 != nil) {    // add literal to the first cube
            Fxu_MatrixAddLiteral( p, pCube1, pLit1.pVar );
            // move to the next literal in this cube
            pLit1 = pLit1.pHNext;
            nLits1++;
        
        } else if !pLit1 && (pLit2 != nil) {    // add literal to the second cube
            Fxu_MatrixAddLiteral( p, pCube2, pLit2.pVar );
            // move to the next literal in this cube
            pLit2 = pLit2.pHNext;
            nLits2++;
        
        } else {
            break; }
    }
    assert(pPair.nLits1 == nLits1, "pPair.nLits1 == nLits1");
    assert(pPair.nLits2 == nLits2, "pPair.nLits2 == nLits2");
    assert(pPair.nBase == nBase, "pPair.nBase == nBase");
}


/**Function*************************************************************

  Synopsis    [Create the node equal to double-cube divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateMatrixDoubleClean(p *Fxu_Matrix, pCubeUse *Fxu_Cube, pCubeRem *Fxu_Cube) {
    var pLit1, bLit1Next *Fxu_Lit;
    var pLit2, bLit2Next *Fxu_Lit;

    // initialize the starting literals
    pLit1     = pCubeUse.lLits.pHead;
    pLit2     = pCubeRem.lLits.pHead;
    if pLit1 {
        bLit1Next = pLit1.pHNext;
    } else {
        bLit1Next = nil;
    }
    if pLit2 != nil {
        bLit2Next = pLit2.pHNext;
    } else {
        bLit2Next = nil;
    }
    // go through the pair and remove the literals in the base
    // from the first cube and all literals from the second cube
    for {
        if pLit1 && (pLit2 != nil) {
            if pLit1.iVar == pLit2.iVar {  // this literal is present in both cubes - it belongs to the base
                // mark the affected var
                if pLit1.pVar.pOrder == nil {
                    // inlining of macro Fxu_MatrixRingVarsAdd
                    (*((p).ppTailVars ) = pLit1.pVar );
((p).ppTailVars  = &(pLit1.pVar).pOrder );
((pLit1.pVar).pOrder  = (Fxu_Var *)1);; }
                // leave the base in pCubeUse; delete it from pCubeRem
                Fxu_MatrixDelLiteral( p, pLit2 );
                // step to the next literals
                pLit1     = bLit1Next;
                pLit2     = bLit2Next;
                if pLit1 {
                    bLit1Next = pLit1.pHNext;
                } else {
                    bLit1Next = nil;
                }
                if pLit2 != nil {
                    bLit2Next = pLit2.pHNext;
                } else {
                    bLit2Next = nil;
                }
            
            } else if pLit1.iVar < pLit2.iVar { // this literal is present in pCubeUse - remove it
                // mark the affected var
                if pLit1.pVar.pOrder == nil {
                    // inlining of macro Fxu_MatrixRingVarsAdd
                    (*((p).ppTailVars ) = pLit1.pVar );
((p).ppTailVars  = &(pLit1.pVar).pOrder );
((pLit1.pVar).pOrder  = (Fxu_Var *)1);; }
                // delete this literal
                Fxu_MatrixDelLiteral( p, pLit1 );
                // step to the next literals
                pLit1     = bLit1Next;
                if pLit1 {
                    bLit1Next = pLit1.pHNext;
                } else {
                    bLit1Next = nil;
                }
            
            } else { // this literal is present in pCubeRem - remove it
                // mark the affected var
                if pLit2.pVar.pOrder == nil {
                    // inlining of macro Fxu_MatrixRingVarsAdd
                    (*((p).ppTailVars ) = pLit2.pVar );
((p).ppTailVars  = &(pLit2.pVar).pOrder );
((pLit2.pVar).pOrder  = (Fxu_Var *)1);; }
                // delete this literal
                Fxu_MatrixDelLiteral( p, pLit2 );
                // step to the next literals
                pLit2     = bLit2Next;
                if pLit2 != nil {
                    bLit2Next = pLit2.pHNext;
                } else {
                    bLit2Next = nil;
                }
            }
        
        } else if pLit1 && !(pLit2 != nil) { // this literal is present in pCubeUse - leave it
            // mark the affected var
            if pLit1.pVar.pOrder == nil {
                // inlining of macro Fxu_MatrixRingVarsAdd
                (*((p).ppTailVars ) = pLit1.pVar );
((p).ppTailVars  = &(pLit1.pVar).pOrder );
((pLit1.pVar).pOrder  = (Fxu_Var *)1);; }
            // delete this literal
            Fxu_MatrixDelLiteral( p, pLit1 );
            // step to the next literals
            pLit1     = bLit1Next;
            if pLit1 {
                bLit1Next = pLit1.pHNext;
            } else {
                bLit1Next = nil;
            }
        
        } else if !pLit1 && (pLit2 != nil) { // this literal is present in pCubeRem - remove it
            // mark the affected var
            if pLit2.pVar.pOrder == nil {
                // inlining of macro Fxu_MatrixRingVarsAdd
                (*((p).ppTailVars ) = pLit2.pVar );
((p).ppTailVars  = &(pLit2.pVar).pOrder );
((pLit2.pVar).pOrder  = (Fxu_Var *)1);; }
            // delete this literal
            Fxu_MatrixDelLiteral( p, pLit2 );
            // step to the next literals
            pLit2     = bLit2Next;
            if pLit2 != nil {
                bLit2Next = pLit2.pHNext;
            } else {
                bLit2Next = nil;
            }
        
        } else {
            break; }
    }
}

/**Function*************************************************************

  Synopsis    [Updates the matrix after selecting a single cube divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateMatrixSingleClean(p *Fxu_Matrix, pVar1 *Fxu_Var, pVar2 *Fxu_Var, pVarNew *Fxu_Var) {
    var pLit1, bLit1Next *Fxu_Lit;
    var pLit2, bLit2Next *Fxu_Lit;

    // initialize the starting literals
    pLit1     = pVar1.lLits.pHead;
    pLit2     = pVar2.lLits.pHead;
    if pLit1 {
        bLit1Next = pLit1.pVNext;
    } else {
        bLit1Next = nil;
    }
    if pLit2 != nil {
        bLit2Next = pLit2.pVNext;
    } else {
        bLit2Next = nil;
    }
    for {
        if pLit1 && (pLit2 != nil) {
            if pLit1.pCube.pVar.iVar == pLit2.pCube.pVar.iVar { // these literals coincide 
                if pLit1.iCube == pLit2.iCube { // these literals coincide 
                
                    // collect the affected cube
                    assert(pLit1.pCube.pOrder == nil, "pLit1.pCube.pOrder == nil");
                    // inlining of macro Fxu_MatrixRingCubesAdd
                    (*((p).ppTailCubes) = pLit1.pCube);
((p).ppTailCubes = &(pLit1.pCube).pOrder);
((pLit1.pCube).pOrder = (Fxu_Cube *)1);;

                    // add the literal to this cube corresponding to the new column
                    Fxu_MatrixAddLiteral( p, pLit1.pCube, pVarNew );
                    // clean the old cubes
                    Fxu_UpdateCleanOldDoubles( p, nil, pLit1.pCube );

                    // remove the literals 
                    Fxu_MatrixDelLiteral( p, pLit1 );
                    Fxu_MatrixDelLiteral( p, pLit2 );

                    // go to the next literals
                    pLit1     = bLit1Next;
                    pLit2     = bLit2Next;
                    if pLit1 {
                        bLit1Next = pLit1.pVNext;
                    } else {
                        bLit1Next = nil;
                    }
                    if pLit2 != nil {
                        bLit2Next = pLit2.pVNext;
                    } else {
                        bLit2Next = nil;
                    }
                
                } else if pLit1.iCube < pLit2.iCube {
                    pLit1     = bLit1Next;
                    if pLit1 {
                        bLit1Next = pLit1.pVNext;
                    } else {
                        bLit1Next = nil;
                    }
                
                } else {
                    pLit2     = bLit2Next;
                    if pLit2 != nil {
                        bLit2Next = pLit2.pVNext;
                    } else {
                        bLit2Next = nil;
                    }
                }
            
            } else if pLit1.pCube.pVar.iVar < pLit2.pCube.pVar.iVar {
                pLit1     = bLit1Next;
                if pLit1 {
                    bLit1Next = pLit1.pVNext;
                } else {
                    bLit1Next = nil;
                }
            
            } else {
                pLit2     = bLit2Next;
                if pLit2 != nil {
                    bLit2Next = pLit2.pVNext;
                } else {
                    bLit2Next = nil;
                }
            }
        
        } else if pLit1 && !(pLit2 != nil) {
            pLit1     = bLit1Next;
            if pLit1 {
                bLit1Next = pLit1.pVNext;
            } else {
                bLit1Next = nil;
            }
        
        } else if !pLit1 && (pLit2 != nil) {
            pLit2     = bLit2Next;
            if pLit2 != nil {
                bLit2Next = pLit2.pVNext;
            } else {
                bLit2Next = nil;
            }
        
        } else {
            break; }
    }
}

/**Function*************************************************************

  Synopsis    [Sort the pairs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdatePairsSort(p *Fxu_Matrix, pDouble *Fxu_Double) {
    var pPair *Fxu_Pair;
    // order the pairs by the first cube to ensure that new literals are added 
    // to the matrix from top to bottom - collect pairs into the array
    p.vPairs.nSize = 0;
    // inlining of macro Fxu_DoubleForEachPair
    for  pPair = (pDouble).lPairs.pHead;          (pPair != nil);          pPair = pPair.pDNext  {
        Vec_PtrPush( p.vPairs, pPair );
    }
    if p.vPairs.nSize < 2 {
        return; }
    // sort
    qsort( (void *)p.vPairs.pArray, (size_t)p.vPairs.nSize, sizeof(Fxu_Pair *), 
            (int (*)(void *, void *)) Fxu_UpdatePairCompare );
    assert(Fxu_UpdatePairCompare( (Fxu_Pair**)p.vPairs.pArray, (Fxu_Pair**)p.vPairs.pArray + p.vPairs.nSize - 1 ) < 0, "Fxu_UpdatePairCompare( (Fxu_Pair**)p.vPairs.pArray, (Fxu_Pair**)p.vPairs.pArray + p.vPairs.nSize - 1 ) < 0");
}

/**Function*************************************************************

  Synopsis    [Compares the vars by their number.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdatePairCompare(ppP1 **Fxu_Pair, ppP2 **Fxu_Pair) int32 {
    var pC1 *Fxu_Cube  = (*ppP1).pCube1;
    var pC2 *Fxu_Cube  = (*ppP2).pCube1;
    var iP1CubeMin, iP2CubeMin int32;
    if pC1.pVar.iVar < pC2.pVar.iVar {
        return -1; }
    if pC1.pVar.iVar > pC2.pVar.iVar {
        return 1; }
    // inlining of macro Fxu_PairMinCubeInt
    
    if (*ppP1).iCube1 < (*ppP1).iCube2 {
        iP1CubeMin = (*ppP1).iCube1;
    } else {
        iP1CubeMin = (*ppP1).iCube2;
    }
    // inlining of macro Fxu_PairMinCubeInt
    
    if (*ppP2).iCube1 < (*ppP2).iCube2 {
        iP2CubeMin = (*ppP2).iCube1;
    } else {
        iP2CubeMin = (*ppP2).iCube2;
    }
    if iP1CubeMin < iP2CubeMin {
        return -1; }
    if iP1CubeMin > iP2CubeMin {
        return 1; }
    assert(false, "0");
    return 0;
}


/**Function*************************************************************

  Synopsis    [Create new variables.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateCreateNewVars(p *Fxu_Matrix, ppVarC **Fxu_Var, ppVarD **Fxu_Var, nCubes int32) {
    var pVarC, pVarD *Fxu_Var;

    // add a new column for the complement
    pVarC = Fxu_MatrixAddVar( p );
    pVarC.nCubes = 0;
    // add a new column for the divisor
    pVarD = Fxu_MatrixAddVar( p );
    pVarD.nCubes = nCubes;

    // mark this entry in the Value2Node array
//    assert( p->pValue2Node[pVarC->iVar] > 0 );
//    p->pValue2Node[pVarD->iVar  ] = p->pValue2Node[pVarC->iVar];
//    p->pValue2Node[pVarD->iVar+1] = p->pValue2Node[pVarC->iVar]+1;

    *ppVarC = pVarC;
    *ppVarD = pVarD;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateCleanOldDoubles(p *Fxu_Matrix, pDiv *Fxu_Double, pCube *Fxu_Cube) {
    var pDivCur *Fxu_Double;
    var pPair *Fxu_Pair;
    var i int32;

    // if the cube is a recently introduced one
    // it does not have pairs allocated
    // in this case, there is nothing to update
    if pCube.pVar.ppPairs == nil {
        return; }

    // go through all the pairs of this cube
    // inlining of macro Fxu_CubeForEachPair
    for  i = 0;        i < pCube.pVar.nCubes && (((pPair) = (pCube).pVar.ppPairs[(pCube).iCube][i]), 1);        i++  {        if pPair == nil {} else {
         // get the divisor of this pair
        pDivCur = pPair.pDiv;
        // skip the current divisor
        if pDivCur == pDiv {
            continue; }
        // remove this pair
        Fxu_ListDoubleDelPair( pDivCur, pPair );    
        // the divisor may have become useless by now
        if pDivCur.lPairs.nItems == 0 {
            assert(pDivCur.Weight == pPair.nBase - 1, "pDivCur.Weight == pPair.nBase - 1");
              Fxu_HeapDoubleDelete( p.pHeapDouble, pDivCur );
            Fxu_MatrixDelDivisor( p, pDivCur );
        
        } else {
            // update the divisor's weight
            pDivCur.Weight -= pPair.nLits1 + pPair.nLits2 - 1 + pPair.nBase;
              Fxu_HeapDoubleUpdate( p.pHeapDouble, pDivCur );
        }
        MEM_FREE_FXU( p, Fxu_Pair, 1, pPair );
    }
    }
    // finally erase all the pair info associated with this cube
    Fxu_PairClearStorage( pCube );
}

/**Function*************************************************************

  Synopsis    [Adds the new divisors that depend on the cube.]

  Description [Go through all the non-empty cubes of this cover 
  (except the given cube) and, for each of them, add the new divisor 
  with the given cube.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateAddNewDoubles(p *Fxu_Matrix, pCube *Fxu_Cube) {
    var pTemp *Fxu_Cube;
    assert(pCube.pOrder, "pCube.pOrder");

    // if the cube is a recently introduced one
    // it does not have pairs allocated
    // in this case, there is nothing to update
    if pCube.pVar.ppPairs == nil {
        return; }

    for  pTemp = pCube.pFirst; pTemp.pVar == pCube.pVar; pTemp = pTemp.pNext  {
        // do not add pairs with the empty cubes
        if pTemp.lLits.nItems == 0 {
            continue; }
        // to prevent adding duplicated pairs of the new cubes
        // do not add the pair, if the current cube is marked 
        if pTemp.pOrder && pTemp.iCube >= pCube.iCube {
            continue; }
        Fxu_MatrixAddDivisor( p, pTemp, pCube );
    }
}

/**Function*************************************************************

  Synopsis    [Removes old single cube divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateCleanOldSingles(p *Fxu_Matrix) { 
    var pSingle, pSingle2 *Fxu_Single;
    var WeightNew int32;
    var Counter int32;

    // inlining of macro Fxu_MatrixForEachSingleSafe
    for  pSingle = (p).lSingles.pHead, pSingle2 = (pSingle? pSingle.pNext: nil);          (pSingle != nil);          pSingle = pSingle2, pSingle2 = (pSingle? pSingle.pNext: nil)  {
        // if at least one of the variables is marked, recalculate
        if pSingle.pVar1.pOrder || pSingle.pVar2.pOrder {
            Counter++;
            // get the new weight
            WeightNew = -2 + Fxu_SingleCountCoincidence( p, pSingle.pVar1, pSingle.pVar2 );
            if WeightNew >= 0 {
                pSingle.Weight = WeightNew;
                Fxu_HeapSingleUpdate( p.pHeapSingle, pSingle );
            
            } else {
                Fxu_HeapSingleDelete( p.pHeapSingle, pSingle );
                Fxu_ListMatrixDelSingle( p, pSingle );
                MEM_FREE_FXU( p, Fxu_Single, 1, pSingle );
            }
        }
    }
//    printf( "Called procedure %d times.\n", Counter );
}

/**Function*************************************************************

  Synopsis    [Updates the single cube divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_UpdateAddNewSingles(p *Fxu_Matrix, pVar *Fxu_Var) {
    Fxu_MatrixComputeSinglesOne( p, pVar );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

