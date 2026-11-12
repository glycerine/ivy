package fxu

/**CFile****************************************************************

  FileName    [fxuInt.h]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Internal declarations of fast extract for unate covers.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuInt.h,v 1.3 2003/04/10 05:42:44 donald Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "base/abc/abc.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// uncomment this macro to switch to standard memory management
//#define USE_SYSTEM_MEMORY_MANAGEMENT 

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*  
    Here is an informal description of the FX data structure.
    (1) The sparse matrix is filled with literals, associated with 
        cubes (row) and variables (columns). The matrix contains 
        all the cubes of all the nodes in the network.
    (2) A cube is associated with 
        (a) its literals in the matrix,
        (b) the output variable of the node, to which this cube belongs,
    (3) A variable is associated with 
        (a) its literals in the matrix and
        (b) the list of cube pairs in the cover, for which it is the output
    (4) A cube pair is associated with two cubes and contains the counters
        of literals in the base and in the cubes without the base
    (5) A double-cube divisor is associated with list of all cube pairs 
        that produce it and its current weight (which is updated automatically 
        each time a new pair is added or an old pair is removed). 
    (6) A single-cube divisor is associated the pair of variables. 
*/

// sparse matrix
type Fxu_Matrix =        FxuMatrix;     // the sparse matrix

// sparse matrix contents: cubes (rows), vars (columns), literals (entries)
type Fxu_Cube =          FxuCube;       // one cube in the sparse matrix
type Fxu_Var =           FxuVar;        // one literal in the sparse matrix
type Fxu_Lit =           FxuLit;        // one entry in the sparse matrix

// double cube divisors
type Fxu_Pair =          FxuPair;       // the pair of cubes
type Fxu_Double =        FxuDouble;     // the double-cube divisor
type Fxu_Single =        FxuSingle;     // the two-literal single-cube divisor

// various lists
type Fxu_ListCube =      FxuListCube;   // the list of cubes
type Fxu_ListVar =       FxuListVar;    // the list of literals
type Fxu_ListLit =       FxuListLit;    // the list of entries
type Fxu_ListPair =      FxuListPair;   // the list of pairs
type Fxu_ListDouble =    FxuListDouble; // the list of divisors
type Fxu_ListSingle =    FxuListSingle; // the list of single-cube divisors

// various heaps
type Fxu_HeapDouble =    FxuHeapDouble; // the heap of divisors
type Fxu_HeapSingle =    FxuHeapSingle; // the heap of variables


// various lists

// the list of cubes in the sparse matrix 
type FxuListCube struct {
    pHead        *Fxu_Cube;
    pTail        *Fxu_Cube;
    nItems              int32;
}

// the list of literals in the sparse matrix 
type FxuListVar struct {
    pHead         *Fxu_Var;
    pTail         *Fxu_Var;
    nItems              int32;
}

// the list of entries in the sparse matrix 
type FxuListLit struct {
    pHead         *Fxu_Lit;
    pTail         *Fxu_Lit;
    nItems              int32;
}

// the list of cube pair in the sparse matrix 
type FxuListPair struct {
    pHead        *Fxu_Pair;
    pTail        *Fxu_Pair;
    nItems              int32;
}

// the list of divisors in the sparse matrix 
type FxuListDouble struct {
    pHead      *Fxu_Double;
    pTail      *Fxu_Double;
    nItems              int32;
}

// the list of divisors in the sparse matrix 
type FxuListSingle struct {
    pHead      *Fxu_Single;
    pTail      *Fxu_Single;
    nItems              int32;
}


// various heaps

// the heap of double cube divisors by weight
type FxuHeapDouble struct {
    pTree     **Fxu_Double;
    nItems              int32;
    nItemsAlloc              int32;
    i              int32;
}

// the heap of variable by their occurrence in the cubes
type FxuHeapSingle struct {
    pTree     **Fxu_Single;
    nItems              int32;
    nItemsAlloc              int32;
    i              int32;
}



// sparse matrix
type FxuMatrix struct { // ~ 30 words

    // the cubes
    lCubes     Fxu_ListCube;      // the double linked list of cubes
    // the values (binary literals)
    lVars      Fxu_ListVar;       // the double linked list of variables
      ppVars        **Fxu_Var;      // the array of variables
    // the double cube divisors
    pTable  *Fxu_ListDouble;      // the hash table of divisors
    nTableSize              int32;  // the hash table size
    nDivs              int32;       // the number of divisors in the table
    nDivsTotal              int32;  // the number of divisors in the table
    pHeapDouble  *Fxu_HeapDouble;    // the heap of divisors by weight
    // the single cube divisors
    lSingles   Fxu_ListSingle;    // the linked list of single cube divisors  
    pHeapSingle  *Fxu_HeapSingle; // the heap of variables by the number of literals in the matrix
    nWeightLimit              int32;// the limit on weight of single cube divisors collected
    nSingleTotal              int32;// the total number of single cube divisors
    // storage for cube pairs
    pppPairs      ***Fxu_Pair;
    ppPairs       **Fxu_Pair;
    // temporary storage for cubes 
    pOrderCubes        *Fxu_Cube;
    ppTailCubes       **Fxu_Cube;
    // temporary storage for variables 
    pOrderVars         *Fxu_Var;
    ppTailVars        **Fxu_Var;
    // temporary storage for pairs
    vPairs       *Vec_Ptr_t;
    // statistics
    nEntries              int32;    // the total number of entries in the sparse matrix
    nDivs1              int32;      // the single cube divisors taken
    nDivs2              int32;      // the double cube divisors taken
    nDivs3              int32;      // the double cube divisors with complement
    // memory manager
    pMemMan  *Extra_MmFixed_t;     // the memory manager for all small sized entries
}

// the cube in the sparse matrix
type FxuCube struct { // 9 words

    iCube              int32;       // the number of this cube in the cover
    pFirst        *Fxu_Cube;      // the pointer to the first cube of this cover
    pVar         *Fxu_Var;        // the variable representing the output of the cover
    lLits      Fxu_ListLit;       // the row in the table 
    pPrev        *Fxu_Cube;       // the previous cube
    pNext        *Fxu_Cube;       // the next cube
    pOrder        *Fxu_Cube;      // the specialized linked list of cubes
}

// the variable in the sparse matrix
type FxuVar struct { // 10 words

    iVar              int32;        // the number of this variable
    nCubes              int32;      // the number of cubes assoc with this var
    pFirst        *Fxu_Cube;      // the first cube assoc with this var
    ppPairs      ***Fxu_Pair;     // the pairs of cubes assoc with this var
    lLits      Fxu_ListLit;       // the column in the table 
    pPrev         *Fxu_Var;       // the previous variable
    pNext         *Fxu_Var;       // the next variable
    pOrder         *Fxu_Var;      // the specialized linked list of variables
}

// the literal entry in the sparse matrix 
type FxuLit struct { // 8 words

    iVar              int32;        // the number of this variable
    iCube              int32;       // the number of this cube
    pCube        *Fxu_Cube;       // the cube of this literal
    pVar         *Fxu_Var;        // the variable of this literal
    pHPrev         *Fxu_Lit;      // prev lit in the cube
    pHNext         *Fxu_Lit;      // next lit in the cube
    pVPrev         *Fxu_Lit;      // prev lit of the var     
    pVNext         *Fxu_Lit;      // next lit of the var   
}

// the cube pair
type FxuPair struct { // 10 words

    nLits1              int32;      // the number of literals in the two cubes
    nLits2              int32;      // the number of literals in the two cubes
    nBase              int32;       // the number of literals in the base
    pDiv      *Fxu_Double;        // the divisor of this pair
    pCube1        *Fxu_Cube;      // the first cube of the pair
    pCube2        *Fxu_Cube;      // the second cube of the pair
    iCube1              int32;      // the first cube of the pair
    iCube2              int32;      // the second cube of the pair
    pDPrev        *Fxu_Pair;      // the previous pair in the divisor
    pDNext        *Fxu_Pair;      // the next pair in the divisor
}

// the double cube divisor
type FxuDouble struct { // 10 words

    Num              int32;         // the unique number of this divisor
    HNum              int32;        // the heap number of this divisor
    Weight              int32;      // the weight of this divisor
    Key         uint32;         // the hash key of this divisor
    lPairs     Fxu_ListPair;      // the pairs of cubes, which produce this divisor
    pPrev      *Fxu_Double;       // the previous divisor in the table
    pNext      *Fxu_Double;       // the next divisor in the table
    pOrder      *Fxu_Double;      // the specialized linked list of divisors
}

// the single cube divisor
type FxuSingle struct { // 7 words

    Num              int32;         // the unique number of this divisor
    HNum              int32;        // the heap number of this divisor
    Weight              int32;      // the weight of this divisor
    pVar1         *Fxu_Var;       // the first variable of the single-cube divisor
    pVar2         *Fxu_Var;       // the second variable of the single-cube divisor
    pPrev      *Fxu_Single;       // the previous divisor in the list
    pNext      *Fxu_Single;       // the next divisor in the list
}

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

// minimum/maximum
// Original macro:
// #define Fxu_Min( a, b ) ( ((a)<(b))? (a):(b) )

// Original macro:
// #define Fxu_Max( a, b ) ( ((a)>(b))? (a):(b) )


// selection of the minimum/maximum cube in the pair
// Original macro:
// #define Fxu_PairMinCube( pPair )    (((pPair)->iCube1 < (pPair)->iCube2)? (pPair)->pCube1: (pPair)->pCube2)

// Original macro:
// #define Fxu_PairMaxCube( pPair )    (((pPair)->iCube1 > (pPair)->iCube2)? (pPair)->pCube1: (pPair)->pCube2)

// Original macro:
// #define Fxu_PairMinCubeInt( pPair ) (((pPair)->iCube1 < (pPair)->iCube2)? (pPair)->iCube1: (pPair)->iCube2)

// Original macro:
// #define Fxu_PairMaxCubeInt( pPair ) (((pPair)->iCube1 > (pPair)->iCube2)? (pPair)->iCube1: (pPair)->iCube2)


// iterators

// Original macro:
// #define Fxu_MatrixForEachCube( Matrix, Cube )\
//     for ( Cube = (Matrix)->lCubes.pHead;\
//           Cube;\
//           Cube = Cube->pNext )

// Original macro:
// #define Fxu_MatrixForEachCubeSafe( Matrix, Cube, Cube2 )\
//     for ( Cube = (Matrix)->lCubes.pHead, Cube2 = (Cube? Cube->pNext: NULL);\
//           Cube;\
//           Cube = Cube2, Cube2 = (Cube? Cube->pNext: NULL) )


// Original macro:
// #define Fxu_MatrixForEachVariable( Matrix, Var )\
//     for ( Var = (Matrix)->lVars.pHead;\
//           Var;\
//           Var = Var->pNext )

// Original macro:
// #define Fxu_MatrixForEachVariableSafe( Matrix, Var, Var2 )\
//     for ( Var = (Matrix)->lVars.pHead, Var2 = (Var? Var->pNext: NULL);\
//           Var;\
//           Var = Var2, Var2 = (Var? Var->pNext: NULL) )


// Original macro:
// #define Fxu_MatrixForEachSingle( Matrix, Single )\
//     for ( Single = (Matrix)->lSingles.pHead;\
//           Single;\
//           Single = Single->pNext )

// Original macro:
// #define Fxu_MatrixForEachSingleSafe( Matrix, Single, Single2 )\
//     for ( Single = (Matrix)->lSingles.pHead, Single2 = (Single? Single->pNext: NULL);\
//           Single;\
//           Single = Single2, Single2 = (Single? Single->pNext: NULL) )


// Original macro:
// #define Fxu_TableForEachDouble( Matrix, Key, Div )\
//     for ( Div = (Matrix)->pTable[Key].pHead;\
//           Div;\
//           Div = Div->pNext )

// Original macro:
// #define Fxu_TableForEachDoubleSafe( Matrix, Key, Div, Div2 )\
//     for ( Div = (Matrix)->pTable[Key].pHead, Div2 = (Div? Div->pNext: NULL);\
//           Div;\
//           Div = Div2, Div2 = (Div? Div->pNext: NULL) )


// Original macro:
// #define Fxu_MatrixForEachDouble( Matrix, Div, Index )\
//     for ( Index = 0; Index < (Matrix)->nTableSize; Index++ )\
//         Fxu_TableForEachDouble( Matrix, Index, Div )

// Original macro:
// #define Fxu_MatrixForEachDoubleSafe( Matrix, Div, Div2, Index )\
//     for ( Index = 0; Index < (Matrix)->nTableSize; Index++ )\
//         Fxu_TableForEachDoubleSafe( Matrix, Index, Div, Div2 )



// Original macro:
// #define Fxu_CubeForEachLiteral( Cube, Lit )\
//     for ( Lit = (Cube)->lLits.pHead;\
//           Lit;\
//           Lit = Lit->pHNext )

// Original macro:
// #define Fxu_CubeForEachLiteralSafe( Cube, Lit, Lit2 )\
//     for ( Lit = (Cube)->lLits.pHead, Lit2 = (Lit? Lit->pHNext: NULL);\
//           Lit;\
//           Lit = Lit2, Lit2 = (Lit? Lit->pHNext: NULL) )


// Original macro:
// #define Fxu_VarForEachLiteral( Var, Lit )\
//     for ( Lit = (Var)->lLits.pHead;\
//           Lit;\
//           Lit = Lit->pVNext )


// Original macro:
// #define Fxu_CubeForEachDivisor( Cube, Div )\
//     for ( Div = (Cube)->lDivs.pHead;\
//           Div;\
//           Div = Div->pCNext )


// Original macro:
// #define Fxu_DoubleForEachPair( Div, Pair )\
//     for ( Pair = (Div)->lPairs.pHead;\
//           Pair;\
//           Pair = Pair->pDNext )

// Original macro:
// #define Fxu_DoubleForEachPairSafe( Div, Pair, Pair2 )\
//     for ( Pair = (Div)->lPairs.pHead, Pair2 = (Pair? Pair->pDNext: NULL);\
//           Pair;\
//           Pair = Pair2, Pair2 = (Pair? Pair->pDNext: NULL) )



// iterator through the cube pairs belonging to the given cube 
// Original macro:
// #define Fxu_CubeForEachPair( pCube, pPair, i )\
//   for ( i = 0;\
//         i < pCube->pVar->nCubes && (((pPair) = (pCube)->pVar->ppPairs[(pCube)->iCube][i]), 1);\
//         i++ )\
//         if ( pPair == NULL ) {} else


// iterator through all the items in the heap
// Original macro:
// #define Fxu_HeapDoubleForEachItem( Heap, Div )\
//     for ( Heap->i = 1;\
//           Heap->i <= Heap->nItems && (Div = Heap->pTree[Heap->i]);\
//           Heap->i++ )

// Original macro:
// #define Fxu_HeapSingleForEachItem( Heap, Single )\
//     for ( Heap->i = 1;\
//           Heap->i <= Heap->nItems && (Single = Heap->pTree[Heap->i]);\
//           Heap->i++ )


// starting the rings
// Original macro:
// #define Fxu_MatrixRingCubesStart( Matrix )    (((Matrix)->ppTailCubes = &((Matrix)->pOrderCubes)), ((Matrix)->pOrderCubes = NULL))

// Original macro:
// #define Fxu_MatrixRingVarsStart(  Matrix )    (((Matrix)->ppTailVars  = &((Matrix)->pOrderVars)),  ((Matrix)->pOrderVars  = NULL))

// stopping the rings
// Original macro:
// #define Fxu_MatrixRingCubesStop(  Matrix )

// Original macro:
// #define Fxu_MatrixRingVarsStop(   Matrix )

// resetting the rings
// Original macro:
// #define Fxu_MatrixRingCubesReset( Matrix )    (((Matrix)->pOrderCubes = NULL), ((Matrix)->ppTailCubes = NULL))

// Original macro:
// #define Fxu_MatrixRingVarsReset(  Matrix )    (((Matrix)->pOrderVars  = NULL), ((Matrix)->ppTailVars  = NULL))

// adding to the rings
// Original macro:
// #define Fxu_MatrixRingCubesAdd( Matrix, Cube) ((*((Matrix)->ppTailCubes) = Cube), ((Matrix)->ppTailCubes = &(Cube)->pOrder), ((Cube)->pOrder = (Fxu_Cube *)1))

// Original macro:
// #define Fxu_MatrixRingVarsAdd(  Matrix, Var ) ((*((Matrix)->ppTailVars ) = Var ), ((Matrix)->ppTailVars  = &(Var)->pOrder ), ((Var)->pOrder  = (Fxu_Var *)1))

// iterating through the rings
// Original macro:
// #define Fxu_MatrixForEachCubeInRing( Matrix, Cube )\
//     if ( (Matrix)->pOrderCubes )\
//     for ( Cube = (Matrix)->pOrderCubes;\
//           Cube != (Fxu_Cube *)1;\
//           Cube = Cube->pOrder )

// Original macro:
// #define Fxu_MatrixForEachCubeInRingSafe( Matrix, Cube, Cube2 )\
//     if ( (Matrix)->pOrderCubes )\
//     for ( Cube = (Matrix)->pOrderCubes, Cube2 = ((Cube != (Fxu_Cube *)1)? Cube->pOrder: (Fxu_Cube *)1);\
//           Cube != (Fxu_Cube *)1;\
//           Cube = Cube2, Cube2 = ((Cube != (Fxu_Cube *)1)? Cube->pOrder: (Fxu_Cube *)1) )

// Original macro:
// #define Fxu_MatrixForEachVarInRing( Matrix, Var )\
//     if ( (Matrix)->pOrderVars )\
//     for ( Var = (Matrix)->pOrderVars;\
//           Var != (Fxu_Var *)1;\
//           Var = Var->pOrder )

// Original macro:
// #define Fxu_MatrixForEachVarInRingSafe( Matrix, Var, Var2 )\
//     if ( (Matrix)->pOrderVars )\
//     for ( Var = (Matrix)->pOrderVars, Var2 = ((Var != (Fxu_Var *)1)? Var->pOrder: (Fxu_Var *)1);\
//           Var != (Fxu_Var *)1;\
//           Var = Var2, Var2 = ((Var != (Fxu_Var *)1)? Var->pOrder: (Fxu_Var *)1) )

// the procedures are related to the above macros
// extern void Fxu_MatrixRingCubesUnmark( Fxu_Matrix * p );

// extern void Fxu_MatrixRingVarsUnmark( Fxu_Matrix * p );



// macros working with memory
// MEM_ALLOC: allocate the given number (Size) of items of type (Type)
// MEM_FREE:  deallocate the pointer (Pointer) to the given number (Size) of items of type (Type)
#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
#define MEM_ALLOC_FXU( Manager, Type, Size )          ((Type *)ABC_ALLOC( char, (Size) * sizeof(Type) ))
#define MEM_FREE_FXU( Manager, Type, Size, Pointer )  if ( Pointer ) { ABC_FREE(Pointer); Pointer = NULL; }
#else
#define MEM_ALLOC_FXU( Manager, Type, Size )\
        ((Type *)Fxu_MemFetch( Manager, (Size) * sizeof(Type) ))
#define MEM_FREE_FXU( Manager, Type, Size, Pointer )\
         if ( Pointer ) { Fxu_MemRecycle( Manager, (char *)(Pointer), (Size) * sizeof(Type) ); Pointer = NULL; }
#endif

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*===== fxu.c ====================================================*/
// extern char *       Fxu_MemFetch( Fxu_Matrix * p, int nBytes );

// extern void         Fxu_MemRecycle( Fxu_Matrix * p, char * pItem, int nBytes );

/*===== fxuCreate.c ====================================================*/
/*===== fxuReduce.c ====================================================*/
/*===== fxuPrint.c ====================================================*/
// extern void         Fxu_MatrixPrint( FILE * pFile, Fxu_Matrix * p );

// extern void         Fxu_MatrixPrintDivisorProfile( FILE * pFile, Fxu_Matrix * p );

/*===== fxuSelect.c ====================================================*/
// extern int          Fxu_Select( Fxu_Matrix * p, Fxu_Single ** ppSingle, Fxu_Double ** ppDouble );

// extern int          Fxu_SelectSCD( Fxu_Matrix * p, int Weight, Fxu_Var ** ppVar1, Fxu_Var ** ppVar2 );

/*===== fxuUpdate.c ====================================================*/
// extern void         Fxu_Update( Fxu_Matrix * p, Fxu_Single * pSingle, Fxu_Double * pDouble );

// extern void         Fxu_UpdateDouble( Fxu_Matrix * p );

// extern void         Fxu_UpdateSingle( Fxu_Matrix * p );

/*===== fxuPair.c ====================================================*/
// extern void         Fxu_PairCanonicize( Fxu_Cube ** ppCube1, Fxu_Cube ** ppCube2 );

// extern unsigned     Fxu_PairHashKeyArray( Fxu_Matrix * p, int piVarsC1[], int piVarsC2[], int nVarsC1, int nVarsC2 );

// extern unsigned     Fxu_PairHashKey( Fxu_Matrix * p, Fxu_Cube * pCube1, Fxu_Cube * pCube2, int * pnBase, int * pnLits1, int * pnLits2 );

// extern unsigned     Fxu_PairHashKeyMv( Fxu_Matrix * p, Fxu_Cube * pCube1, Fxu_Cube * pCube2, int * pnBase, int * pnLits1, int * pnLits2 );

// extern int          Fxu_PairCompare( Fxu_Pair * pPair1, Fxu_Pair * pPair2 );

// extern void         Fxu_PairAllocStorage( Fxu_Var * pVar, int nCubes );

// extern void         Fxu_PairFreeStorage( Fxu_Var * pVar );

// extern void         Fxu_PairClearStorage( Fxu_Cube * pCube );

// extern Fxu_Pair *   Fxu_PairAlloc( Fxu_Matrix * p, Fxu_Cube * pCube1, Fxu_Cube * pCube2 );

// extern void         Fxu_PairAdd( Fxu_Pair * pPair );

/*===== fxuSingle.c ====================================================*/
// extern void         Fxu_MatrixComputeSingles( Fxu_Matrix * p, int fUse0, int nSingleMax );

// extern void         Fxu_MatrixComputeSinglesOne( Fxu_Matrix * p, Fxu_Var * pVar );

// extern int          Fxu_SingleCountCoincidence( Fxu_Matrix * p, Fxu_Var * pVar1, Fxu_Var * pVar2 );

/*===== fxuMatrix.c ====================================================*/
// matrix
// extern Fxu_Matrix * Fxu_MatrixAllocate();

// extern void         Fxu_MatrixDelete( Fxu_Matrix * p );

// double-cube divisor
// extern void         Fxu_MatrixAddDivisor( Fxu_Matrix * p, Fxu_Cube * pCube1, Fxu_Cube * pCube2 );

// extern void         Fxu_MatrixDelDivisor( Fxu_Matrix * p, Fxu_Double * pDiv );

// single-cube divisor
// extern void          Fxu_MatrixAddSingle( Fxu_Matrix * p, Fxu_Var * pVar1, Fxu_Var * pVar2, int Weight );

// variable
// extern Fxu_Var *    Fxu_MatrixAddVar( Fxu_Matrix * p );

// cube
// extern Fxu_Cube *   Fxu_MatrixAddCube( Fxu_Matrix * p, Fxu_Var * pVar, int iCube );

// literal
// extern void         Fxu_MatrixAddLiteral( Fxu_Matrix * p, Fxu_Cube * pCube, Fxu_Var * pVar );

// extern void         Fxu_MatrixDelLiteral( Fxu_Matrix * p, Fxu_Lit * pLit );

/*===== fxuList.c ====================================================*/
// matrix -> variable 
// extern void         Fxu_ListMatrixAddVariable( Fxu_Matrix * p, Fxu_Var * pVar );

// extern void         Fxu_ListMatrixDelVariable( Fxu_Matrix * p, Fxu_Var * pVar );

// matrix -> cube
// extern void         Fxu_ListMatrixAddCube( Fxu_Matrix * p, Fxu_Cube * pCube );

// extern void         Fxu_ListMatrixDelCube( Fxu_Matrix * p, Fxu_Cube * pCube );

// matrix -> single
// extern void         Fxu_ListMatrixAddSingle( Fxu_Matrix * p, Fxu_Single * pSingle );

// extern void         Fxu_ListMatrixDelSingle( Fxu_Matrix * p, Fxu_Single * pSingle );

// table -> divisor
// extern void         Fxu_ListTableAddDivisor( Fxu_Matrix * p, Fxu_Double * pDiv );

// extern void         Fxu_ListTableDelDivisor( Fxu_Matrix * p, Fxu_Double * pDiv );

// cube -> literal 
// extern void         Fxu_ListCubeAddLiteral( Fxu_Cube * pCube, Fxu_Lit * pLit );

// extern void         Fxu_ListCubeDelLiteral( Fxu_Cube * pCube, Fxu_Lit * pLit );

// var -> literal
// extern void         Fxu_ListVarAddLiteral( Fxu_Var * pVar, Fxu_Lit * pLit );

// extern void         Fxu_ListVarDelLiteral( Fxu_Var * pVar, Fxu_Lit * pLit );

// divisor -> pair
// extern void         Fxu_ListDoubleAddPairLast( Fxu_Double * pDiv, Fxu_Pair * pLink );

// extern void         Fxu_ListDoubleAddPairFirst( Fxu_Double * pDiv, Fxu_Pair * pLink );

// extern void         Fxu_ListDoubleAddPairMiddle( Fxu_Double * pDiv, Fxu_Pair * pSpot, Fxu_Pair * pLink );

// extern void         Fxu_ListDoubleDelPair( Fxu_Double * pDiv, Fxu_Pair * pPair );

/*===== fxuHeapDouble.c ====================================================*/
// extern Fxu_HeapDouble * Fxu_HeapDoubleStart();

// extern void         Fxu_HeapDoubleStop( Fxu_HeapDouble * p );

// extern void         Fxu_HeapDoublePrint( FILE * pFile, Fxu_HeapDouble * p );

// extern void         Fxu_HeapDoubleCheck( Fxu_HeapDouble * p );

// extern void         Fxu_HeapDoubleCheckOne( Fxu_HeapDouble * p, Fxu_Double * pDiv );


// extern void         Fxu_HeapDoubleInsert( Fxu_HeapDouble * p, Fxu_Double * pDiv );
  
// extern void         Fxu_HeapDoubleUpdate( Fxu_HeapDouble * p, Fxu_Double * pDiv );
  
// extern void         Fxu_HeapDoubleDelete( Fxu_HeapDouble * p, Fxu_Double * pDiv );
  
// extern int          Fxu_HeapDoubleReadMaxWeight( Fxu_HeapDouble * p );
  
// extern Fxu_Double * Fxu_HeapDoubleReadMax( Fxu_HeapDouble * p );
  
// extern Fxu_Double * Fxu_HeapDoubleGetMax( Fxu_HeapDouble * p );
  
/*===== fxuHeapSingle.c ====================================================*/
// extern Fxu_HeapSingle * Fxu_HeapSingleStart();

// extern void         Fxu_HeapSingleStop( Fxu_HeapSingle * p );

// extern void         Fxu_HeapSinglePrint( FILE * pFile, Fxu_HeapSingle * p );

// extern void         Fxu_HeapSingleCheck( Fxu_HeapSingle * p );

// extern void         Fxu_HeapSingleCheckOne( Fxu_HeapSingle * p, Fxu_Single * pSingle );


// extern void         Fxu_HeapSingleInsert( Fxu_HeapSingle * p, Fxu_Single * pSingle );
  
// extern void         Fxu_HeapSingleUpdate( Fxu_HeapSingle * p, Fxu_Single * pSingle );
  
// extern void         Fxu_HeapSingleDelete( Fxu_HeapSingle * p, Fxu_Single * pSingle );
  
// extern int          Fxu_HeapSingleReadMaxWeight( Fxu_HeapSingle * p );
  
// extern Fxu_Single * Fxu_HeapSingleReadMax( Fxu_HeapSingle * p );

// extern Fxu_Single * Fxu_HeapSingleGetMax( Fxu_HeapSingle * p );
  



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                         END OF FILE                              ///
////////////////////////////////////////////////////////////////////////

