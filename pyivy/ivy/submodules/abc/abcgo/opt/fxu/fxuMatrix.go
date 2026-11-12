package fxu

/**CFile****************************************************************

  FileName    [fxuMatrix.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures to manipulate the sparse matrix.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuMatrix.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fxuInt.h"

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
func Fxu_MatrixAllocate() *Fxu_Matrix {
    var p *Fxu_Matrix;
    // inlining of macro ABC_ALLOC
    p = ((Fxu_Matrix *) malloc(sizeof(Fxu_Matrix) * (size_t)(1)));
    memset( p, 0, sizeof(Fxu_Matrix) );
    p.nTableSize = Abc_PrimeCudd(10000);
    // inlining of macro ABC_ALLOC
    p.pTable = ((Fxu_ListDouble *) malloc(sizeof(Fxu_ListDouble) * (size_t)(p.nTableSize)));
    memset( p.pTable, 0, sizeof(Fxu_ListDouble) * p.nTableSize );
#ifndef USE_SYSTEM_MEMORY_MANAGEMENT
    {
        // get the largest size in bytes for the following structures:
        // Fxu_Cube, Fxu_Var, Fxu_Lit, Fxu_Pair, Fxu_Double, Fxu_Single
        // (currently, Fxu_Var, Fxu_Pair, Fxu_Double take 10 machine words)
        int nSizeMax, nSizeCur;
        nSizeMax = -1;
        nSizeCur = sizeof(Fxu_Cube);
        if ( nSizeMax < nSizeCur )
             nSizeMax = nSizeCur;
        nSizeCur = sizeof(Fxu_Var);
        if ( nSizeMax < nSizeCur )
             nSizeMax = nSizeCur;
        nSizeCur = sizeof(Fxu_Lit);
        if ( nSizeMax < nSizeCur )
             nSizeMax = nSizeCur;
        nSizeCur = sizeof(Fxu_Pair);
        if ( nSizeMax < nSizeCur )
             nSizeMax = nSizeCur;
        nSizeCur = sizeof(Fxu_Double);
        if ( nSizeMax < nSizeCur )
             nSizeMax = nSizeCur;
        nSizeCur = sizeof(Fxu_Single);
        if ( nSizeMax < nSizeCur )
             nSizeMax = nSizeCur;
        p->pMemMan  = Extra_MmFixedStart( nSizeMax ); 
    }
#endif
    p.pHeapDouble = Fxu_HeapDoubleStart();
    p.pHeapSingle = Fxu_HeapSingleStart();
    p.vPairs = Vec_PtrAlloc( 100 );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixDelete(p *Fxu_Matrix) {
    Fxu_HeapDoubleCheck( p.pHeapDouble );
    Fxu_HeapDoubleStop( p.pHeapDouble );
    Fxu_HeapSingleStop( p.pHeapSingle );

    // delete other things
#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    // this code is not needed when the custom memory manager is used
    {
        Fxu_Cube * pCube, * pCube2;
        Fxu_Var * pVar, * pVar2;
        Fxu_Lit * pLit, * pLit2;
        Fxu_Double * pDiv, * pDiv2;
        Fxu_Single * pSingle, * pSingle2;
        Fxu_Pair * pPair, * pPair2;
        int i;
        // delete the divisors
        Fxu_MatrixForEachDoubleSafe( p, pDiv, pDiv2, i )
        {
            Fxu_DoubleForEachPairSafe( pDiv, pPair, pPair2 )
                MEM_FREE_FXU( p, Fxu_Pair, 1, pPair );
               MEM_FREE_FXU( p, Fxu_Double, 1, pDiv );
        }
        Fxu_MatrixForEachSingleSafe( p, pSingle, pSingle2 )
               MEM_FREE_FXU( p, Fxu_Single, 1, pSingle );
        // delete the entries
        Fxu_MatrixForEachCube( p, pCube )
            Fxu_CubeForEachLiteralSafe( pCube, pLit, pLit2 )
                MEM_FREE_FXU( p, Fxu_Lit, 1, pLit );
        // delete the cubes
        Fxu_MatrixForEachCubeSafe( p, pCube, pCube2 )
            MEM_FREE_FXU( p, Fxu_Cube, 1, pCube );
        // delete the vars
        Fxu_MatrixForEachVariableSafe( p, pVar, pVar2 )
            MEM_FREE_FXU( p, Fxu_Var, 1, pVar );
    }
#else
    Extra_MmFixedStop( p->pMemMan );
#endif

    Vec_PtrFree( p.vPairs );
    p.pppPairs = nil;
    p.ppPairs = nil;
//    ABC_FREE( p->pPairsTemp );
    p.pTable = nil;
    p.ppVars = nil;
    p = nil;
}



/**Function*************************************************************

  Synopsis    [Adds a variable to the matrix.]

  Description [This procedure always adds variables at the end of the matrix.
  It assigns the var's node and number. It adds the var to the linked list of
  all variables and to the table of all nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixAddVar(p *Fxu_Matrix) *Fxu_Var {
    var pVar *Fxu_Var;
    pVar = MEM_ALLOC_FXU( p, Fxu_Var, 1 );
    memset( pVar, 0, sizeof(Fxu_Var) );
    pVar.iVar = p.lVars.nItems;
    p.ppVars[pVar.iVar] = pVar;
    Fxu_ListMatrixAddVariable( p, pVar );
    return pVar;
}

/**Function*************************************************************

  Synopsis    [Adds a literal to the matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixAddCube(p *Fxu_Matrix, pVar *Fxu_Var, iCube int32) *Fxu_Cube {
    var pCube *Fxu_Cube;
    pCube = MEM_ALLOC_FXU( p, Fxu_Cube, 1 );
    memset( pCube, 0, sizeof(Fxu_Cube) );
    pCube.pVar  = pVar;
    pCube.iCube = iCube;
    Fxu_ListMatrixAddCube( p, pCube );
    return pCube;
}

/**Function*************************************************************

  Synopsis    [Adds a literal to the matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixAddLiteral(p *Fxu_Matrix, pCube *Fxu_Cube, pVar *Fxu_Var) {
    var pLit *Fxu_Lit;
    pLit = MEM_ALLOC_FXU( p, Fxu_Lit, 1 );
    memset( pLit, 0, sizeof(Fxu_Lit) );
    // insert the literal into two linked lists
    Fxu_ListCubeAddLiteral( pCube, pLit );
    Fxu_ListVarAddLiteral( pVar, pLit );
    // set the back pointers
    pLit.pCube = pCube;
    pLit.pVar  = pVar;
    pLit.iCube = pCube.iCube;
    pLit.iVar  = pVar.iVar;
    // increment the literal counter
    p.nEntries++;
}

/**Function*************************************************************

  Synopsis    [Deletes the divisor from the matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixDelDivisor(p *Fxu_Matrix, pDiv *Fxu_Double) {
    // delete divisor from the table
    Fxu_ListTableDelDivisor( p, pDiv );
    // recycle the divisor
    MEM_FREE_FXU( p, Fxu_Double, 1, pDiv );
}

/**Function*************************************************************

  Synopsis    [Deletes the literal fromthe matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixDelLiteral(p *Fxu_Matrix, pLit *Fxu_Lit) {
    // delete the literal
    Fxu_ListCubeDelLiteral( pLit.pCube, pLit );
    Fxu_ListVarDelLiteral( pLit.pVar, pLit );
    MEM_FREE_FXU( p, Fxu_Lit, 1, pLit );
    // increment the literal counter
    p.nEntries--;
}


/**Function*************************************************************

  Synopsis    [Creates and adds a single cube divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixAddSingle(p *Fxu_Matrix, pVar1 *Fxu_Var, pVar2 *Fxu_Var, Weight int32) {
    var pSingle *Fxu_Single;
    assert(pVar1.iVar < pVar2.iVar, "pVar1.iVar < pVar2.iVar");
    pSingle = MEM_ALLOC_FXU( p, Fxu_Single, 1 );
    memset( pSingle, 0, sizeof(Fxu_Single) );
    pSingle.Num = p.lSingles.nItems;
    pSingle.Weight = Weight;
    pSingle.HNum = 0;
    pSingle.pVar1 = pVar1;
    pSingle.pVar2 = pVar2;
    Fxu_ListMatrixAddSingle( p, pSingle );
    // add to the heap
    Fxu_HeapSingleInsert( p.pHeapSingle, pSingle );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_MatrixAddDivisor(p *Fxu_Matrix, pCube1 *Fxu_Cube, pCube2 *Fxu_Cube) {
    var pPair *Fxu_Pair;
    var pDiv *Fxu_Double;
    var nBase, nLits1, nLits2 int32;
    var fFound int32;
    var Key uint32;

    // canonicize the pair
    Fxu_PairCanonicize( &pCube1, &pCube2 );
    // compute the hash key
    Key = Fxu_PairHashKey( p, pCube1, pCube2, &nBase, &nLits1, &nLits2 );

    // create the cube pair
    pPair = Fxu_PairAlloc( p, pCube1, pCube2 );
    pPair.nBase  = nBase;
    pPair.nLits1 = nLits1;
    pPair.nLits2 = nLits2;

    // check if the divisor for this pair already exists
    fFound = 0;
    Key %= p.nTableSize;
    // inlining of macro Fxu_TableForEachDouble
    for  pDiv = (p).pTable[Key].pHead;          (pDiv != nil);          pDiv = pDiv.pNext  {
        if Fxu_PairCompare( pPair, pDiv.lPairs.pTail ) { // they are equal
        
            fFound = 1;
            break;
        }
    }

    if !(fFound != 0) {   // create the new divisor
        pDiv = MEM_ALLOC_FXU( p, Fxu_Double, 1 );
        memset( pDiv, 0, sizeof(Fxu_Double) );
        pDiv.Key = Key;
        // set the number of this divisor
        pDiv.Num = p.nDivsTotal++; // p->nDivs;
        // insert the divisor in the table
        Fxu_ListTableAddDivisor( p, pDiv );
        // set the initial cost of the divisor
        pDiv.Weight -= pPair.nLits1 + pPair.nLits2;
    }

    // link the pair to the cubes
    Fxu_PairAdd( pPair );
    // connect the pair and the divisor
    pPair.pDiv = pDiv;
    Fxu_ListDoubleAddPairLast( pDiv, pPair );    
    // update the max number of pairs in a divisor
//    if ( p->nPairsMax < pDiv->lPairs.nItems )
//        p->nPairsMax = pDiv->lPairs.nItems;
    // update the divisor's weight
    pDiv.Weight += pPair.nLits1 + pPair.nLits2 - 1 + pPair.nBase;
    if fFound != 0 { // update the divisor in the heap
        Fxu_HeapDoubleUpdate( p.pHeapDouble, pDiv );
    } else {  // add the new divisor to the heap
        Fxu_HeapDoubleInsert( p.pHeapDouble, pDiv ); }
}

/*
    {
        int piVarsC1[100], piVarsC2[100], nVarsC1, nVarsC2;
        Fxu_Double * pDivCur;
        Fxu_MatrixGetDoubleVars( p, pDiv, piVarsC1, piVarsC2, &nVarsC1, &nVarsC2 );
        pDivCur = Fxu_MatrixFindDouble( p, piVarsC1, piVarsC2, nVarsC1, nVarsC2 );
        assert( pDivCur == pDiv );
    }
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

