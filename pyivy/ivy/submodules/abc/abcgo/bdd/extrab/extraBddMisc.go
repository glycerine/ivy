package extrab

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [extraBddMisc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [DD-based utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraBddMisc.c,v 1.4 2005/10/04 00:19:54 alanmi Exp $]

***********************************************************************/

//#include "extraBdd.h"

//ABC_NAMESPACE_IMPL_START


/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// file "extraDdTransfer.c"
// static DdNode * extraTransferPermuteRecur( DdManager * ddS, DdManager * ddD, DdNode * f, st__table * table, int * Permute );

// static DdNode * extraTransferPermute( DdManager * ddS, DdManager * ddD, DdNode * f, int * Permute );

// static DdNode * cuddBddPermuteRecur ARGS( ( DdManager * manager, DdHashTable * table, DdNode * node, int *permut ) );


// static DdNode * extraBddAndPermute( DdHashTable * table, DdManager * ddF, DdNode * bF, DdManager * ddG, DdNode * bG, int * pPermute );


// file "cuddUtils.c"
// void ddSupportStep2(DdNode *f, int *support);

// void ddClearFlag2(DdNode *f);


// static DdNode* extraZddPrimes( DdManager *dd, DdNode* F );


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Convert a {A,B}DD from a manager to another with variable remapping.]

  Description [Convert a {A,B}DD from a manager to another one. The orders of the
  variables in the two managers may be different. Returns a
  pointer to the {A,B}DD in the destination manager if successful; NULL
  otherwise. The i-th entry in the array Permute tells what is the index
  of the i-th variable from the old manager in the new manager.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Extra_TransferPermute(ddSource *DdManager, ddDestination *DdManager, f *DdNode, Permute *int32) *DdNode {
    var bRes *DdNode;
    for
    {
        ddDestination.reordered = 0;
        bRes = extraTransferPermute( ddSource, ddDestination, f, Permute );
    
if !( ddDestination.reordered == 1 ) { break }
}
    
    return ( bRes );

}                               /* end of Extra_TransferPermute */

/**Function********************************************************************

  Synopsis    [Transfers the BDD from one manager into another level by level.]

  Description [Transfers the BDD from one manager into another while
  preserving the correspondence between variables level by level.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Extra_TransferLevelByLevel(ddSource *DdManager, ddDestination *DdManager, f *DdNode) *DdNode {
    var bRes *DdNode;
    var pPermute *int32;
    var nMin, nMax, i int32;

    // inlining of macro ddMin
    
    if (ddDestination.size) < (ddSource.size) {
        nMin = (ddDestination.size);
    } else {
        nMin = (ddSource.size);
    }
    // inlining of macro ddMax
    
    if (ddDestination.size) > (ddSource.size) {
        nMax = (ddDestination.size);
    } else {
        nMax = (ddSource.size);
    }
    pPermute = ABC_ALLOC( int, nMax );
    // set up the variable permutation
    for  i = 0; i < nMin; i++  {
        pPermute[ ddSource.invperm[i] ] = ddDestination.invperm[i];
    }
    if ddSource.size > ddDestination.size {
        for       ; i < nMax; i++  {
            pPermute[ ddSource.invperm[i] ] = -1;
        }
    }
    bRes = Extra_TransferPermute( ddSource, ddDestination, f, pPermute );
    ABC_FREE( pPermute );
    return bRes;
}

/**Function********************************************************************

  Synopsis    [Remaps the function to depend on the topmost variables on the manager.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddRemapUp(dd *DdManager, bF *DdNode) *DdNode {
    var pPermute *int32;
    var bSupp, bTemp, bRes *DdNode;
    var Counter int32;

    pPermute = ABC_ALLOC( int, dd.size );

    // get support
    bSupp = Cudd_Support( dd, bF );    Cudd_Ref( bSupp );

    // create the variable map
    // to remap the DD into the upper part of the manager
    Counter = 0;
    // inlining of macro cuddT
    for  bTemp = bSupp; bTemp != dd.one; bTemp = ((bTemp).type.kids.T)  {
        pPermute[bTemp.index] = dd.invperm[Counter++];
    }

    // transfer the BDD and remap it
    bRes = Cudd_bddPermute( dd, bF, pPermute );  Cudd_Ref( bRes );

    // remove support
    Cudd_RecursiveDeref( dd, bSupp );

    // return
    Cudd_Deref( bRes );
    ABC_FREE( pPermute );
    return bRes;
}

/**Function********************************************************************

  Synopsis    [Moves the BDD by the given number of variables up or down.]

  Description []

  SideEffects []

  SeeAlso     [Extra_bddShift]

******************************************************************************/
DdNode * Extra_bddMove( 
  DdManager * dd,   /* the DD manager */
  DdNode * bF,
  int nVars) 
{
    var res *DdNode;
    var bVars *DdNode;
    if nVars == 0 {
        return bF; }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if  (((DdNode *)((ptruint)(bF) & ~01))).index == CUDD_CONST_INDEX {
        return bF; }
    assert(nVars <= dd.size, "nVars <= dd.size");
    if nVars > 0 {
        bVars = dd.vars[nVars];
    } else {
        // inlining of macro Cudd_Not
        bVars = ((DdNode *)((ptrint)(dd.vars[-nVars]) ^ 01)); }

    for {
        dd.reordered = 0;
        res = extraBddMove( dd, bF, bVars );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_bddMove */

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_StopManager(dd *DdManager) {
    var RetValue int32;
    // check for remaining references in the package
    RetValue = Cudd_CheckZeroRef( dd );
    if RetValue > 10 {
//    if ( RetValue )
        fmt.Printf("\nThe number of referenced nodes = %d\n\n", RetValue); }
//  Cudd_PrintInfo( dd, stdout );
    Cudd_Quit( dd );
}

/**Function********************************************************************

  Synopsis    [Outputs the BDD in a readable format.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Extra_bddPrint(dd *DdManager, F *DdNode) {
    var Gen *DdGen;
    var Cube *int32;
    CUDD_VALUE_TYPE Value;
    var nVars int32  = dd.size;
    var fFirstCube int32  = 1;
    var i int32;

    if F == nil {
        fmt.Printf("NULL");
        return;
    }
    if F == b0 {
        fmt.Printf("Constant 0");
        return;
    }
    if F == b1 {
        fmt.Printf("Constant 1");
        return;
    }

    // inlining of macro Cudd_ForeachCube
    for (Gen) = Cudd_FirstCube(dd, F, &Cube, &Value);        Cudd_IsGenEmpty(Gen) ? Cudd_GenFree(Gen) : TRUE;        (void) Cudd_NextCube(Gen, &Cube, &Value) {
        if fFirstCube != 0 {
            fFirstCube = 0;
        } else {
//          Output << " + ";
            fmt.Printf(" + "); }

        for  i = 0; i < nVars; i++  {
            if Cube[i] == 0 {
                fmt.Printf("[%d]'", i);
//              printf( "%c'", (char)('a'+i) );
            } else if Cube[i] == 1 {
                fmt.Printf("[%d]", i); }
        }
//              printf( "%c", (char)('a'+i) );
    }

//  printf("\n");
}

/**Function********************************************************************

  Synopsis    [Outputs the BDD in a readable format.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Extra_bddPrintSupport(dd *DdManager, F *DdNode) {
    var bSupp *DdNode;
    bSupp = Cudd_Support( dd, F );   Cudd_Ref( bSupp );
    Extra_bddPrint( dd, bSupp );
    Cudd_RecursiveDeref( dd, bSupp );
}

/**Function********************************************************************

  Synopsis    [Returns the size of the support.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddSuppSize(dd *DdManager, bSupp *DdNode) int32 {
    var Counter int32;
    for  bSupp != b1  {
        // inlining of macro Cudd_IsComplement
        assert(!((int32( ((ptrint) (bSupp) & 01)) != 0)), "!/*__starting_point_inline_macro:Cudd_IsComplement__*/(int32( ((ptrint) (bSupp) & 01)))");
        // inlining of macro cuddE
        assert(((bSupp).type.kids.E) == b0, "/*__starting_point_inline_macro:cuddE__*/((bSupp).type.kids.E) == b0");

        // inlining of macro cuddT
        bSupp = ((bSupp).type.kids.T);
        Counter++;
    }
    return Counter;
}

/**Function********************************************************************

  Synopsis    [Returns 1 if the support contains the given BDD variable.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddSuppContainVar(dd *DdManager, bS *DdNode, bVar *DdNode) int32 { 
    // inlining of macro cuddT
    for  ; bS != b1; bS = ((bS).type.kids.T)  {
        if bS.index == bVar.index {
            return 1; }
    }
    return 0;
}

/**Function********************************************************************

  Synopsis    [Returns 1 if two supports represented as BDD cubes are overlapping.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddSuppOverlapping(dd *DdManager, S1 *DdNode, S2 *DdNode) int32 {
    for  S1.index != CUDD_CONST_INDEX && S2.index != CUDD_CONST_INDEX  {
        // if the top vars are the same, they intersect
        if S1.index == S2.index {
            return 1; }
        // if the top vars are different, skip the one, which is higher
        if dd.perm[S1.index] < dd.perm[S2.index] {
            // inlining of macro cuddT
            S1 = ((S1).type.kids.T);
        } else {
            // inlining of macro cuddT
            S2 = ((S2).type.kids.T); }
    }
    return 0;
}

/**Function********************************************************************

  Synopsis    [Returns the number of different vars in two supports.]

  Description [Counts the number of variables that appear in one support and 
  does not appear in other support. If the number exceeds DiffMax, returns DiffMax.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddSuppDifferentVars(dd *DdManager, S1 *DdNode, S2 *DdNode, DiffMax int32) int32 {
    var Result int32;
    for  S1.index != CUDD_CONST_INDEX && S2.index != CUDD_CONST_INDEX  {
        // if the top vars are the same, this var is the same
        if S1.index == S2.index {
            // inlining of macro cuddT
            S1 = ((S1).type.kids.T);
            // inlining of macro cuddT
            S2 = ((S2).type.kids.T);
            continue;
        }
        // the top var is different
        Result++;

        if Result >= DiffMax {
            return DiffMax; }

        // if the top vars are different, skip the one, which is higher
        if dd.perm[S1.index] < dd.perm[S2.index] {
            // inlining of macro cuddT
            S1 = ((S1).type.kids.T);
        } else {
            // inlining of macro cuddT
            S2 = ((S2).type.kids.T); }
    }

    // consider the remaining variables
    if S1.index != CUDD_CONST_INDEX {
        Result += Extra_bddSuppSize(dd,S1);
    } else if S2.index != CUDD_CONST_INDEX {
        Result += Extra_bddSuppSize(dd,S2); }

    if Result >= DiffMax {
        return DiffMax; }
    return Result;
}


/**Function********************************************************************

  Synopsis    [Checks the support containment.]

  Description [This function returns 1 if one support is contained in another.
  In this case, bLarge (bSmall) is assigned to point to the larger (smaller) support.
  If the supports are identical, return 0 and does not assign the supports!]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddSuppCheckContainment(dd *DdManager, bL *DdNode, bH *DdNode, bLarge **DdNode, bSmall **DdNode) int32 {
    var bSL *DdNode  = bL;
    var bSH *DdNode  = bH;
    var fLcontainsH int32  = 1;
    var fHcontainsL int32  = 1;
    var TopVar int32;
    
    if bSL == bSH {
        return 0; }

    for  bSL != b1 || bSH != b1  {
        if bSL == b1 { // Low component has no vars; High components has some vars
            fLcontainsH = 0;
            if fHcontainsL == 0 {
                return 0;
            } else {
                break; }
        }

        if bSH == b1 { // similarly
            fHcontainsL = 0;
            if fLcontainsH == 0 {
                return 0;
            } else {
                break; }
        }

        // determine the topmost var of the supports by comparing their levels
        if dd.perm[bSL.index] < dd.perm[bSH.index] {
            TopVar = bSL.index;
        } else {
            TopVar = bSH.index; }

        if TopVar == bSL.index && TopVar == bSH.index { // they are on the same level
            // it does not tell us anything about their containment
            // skip this var
            // inlining of macro cuddT
            bSL = ((bSL).type.kids.T);
            // inlining of macro cuddT
            bSH = ((bSH).type.kids.T);
        
        } else if TopVar == bSL.index { // and TopVar != bSH->index
         // Low components is higher and contains more vars
            // it is not possible that High component contains Low
            fHcontainsL = 0;
            // skip this var
            // inlining of macro cuddT
            bSL = ((bSL).type.kids.T);
        
        } else { // if ( TopVar == bSH->index ) // and TopVar != bSL->index
         // similarly
            fLcontainsH = 0;
            // skip this var
            // inlining of macro cuddT
            bSH = ((bSH).type.kids.T);
        }

        // check the stopping condition
        if !(fHcontainsL != 0) && !(fLcontainsH != 0) {
            return 0; }
    }
    // only one of them can be true at the same time
    assert(!(fHcontainsL != 0) || !(fLcontainsH != 0), "!fHcontainsL || !fLcontainsH");
    if fHcontainsL != 0 {
        *bLarge = bH;
        *bSmall = bL;
    
    } else { // fLcontainsH
    
        *bLarge = bL;
        *bSmall = bH;
    }
    return 1;
}


/**Function********************************************************************

  Synopsis    [Finds variables on which the DD depends and returns them as am array.]

  Description [Finds the variables on which the DD depends. Returns an array 
  with entries set to 1 for those variables that belong to the support; 
  NULL otherwise. The array is allocated by the user and should have at least
  as many entries as the maximum number of variables in BDD and ZDD parts of 
  the manager.]

  SideEffects [None]

  SeeAlso     [Cudd_Support Cudd_VectorSupport Cudd_ClassifySupport]

******************************************************************************/
int * 
Extra_SupportArray(
  DdManager * dd, /* manager */
  DdNode * f,     /* DD whose support is sought */
  int * support ) /* array allocated by the user */
{
    var i, size int32;

    /* Initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    for i = 0; i < size; i++ { 
        support[i] = 0;
    }
    
    /* Compute support and clean up markers. */
    // inlining of macro Cudd_Regular
    ddSupportStep2(((DdNode *)((ptruint)(f) & ~01)),support);
    // inlining of macro Cudd_Regular
    ddClearFlag2(((DdNode *)((ptruint)(f) & ~01)));

    return(support);

} /* end of Extra_SupportArray */

/**Function********************************************************************

  Synopsis    [Finds the variables on which a set of DDs depends.]

  Description [Finds the variables on which a set of DDs depends.
  The set must contain either BDDs and ADDs, or ZDDs.
  Returns a BDD consisting of the product of the variables if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_Support Cudd_ClassifySupport]

******************************************************************************/
int *
Extra_VectorSupportArray( 
  DdManager * dd, /* manager */ 
  DdNode ** F, /* array of DDs whose support is sought */ 
  int n, /* size of the array */  
  int * support ) /* array allocated by the user */
{
    var i, size int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    for  i = 0; i < size; i++  {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    for  i = 0; i < n; i++  {
        // inlining of macro Cudd_Regular
        ddSupportStep2( ((DdNode *)((ptruint)(F[i]) & ~01)), support );
    }
    for  i = 0; i < n; i++  {
        // inlining of macro Cudd_Regular
        ddClearFlag2( ((DdNode *)((ptruint)(F[i]) & ~01)) );
    }

    return support;
}

/**Function********************************************************************

  Synopsis    [Find any cube belonging to the on-set of the function.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddFindOneCube(dd *DdManager, bF *DdNode) *DdNode {
    var s_Temp string;
    var bCube, bTemp *DdNode;
    var v int32;

    // get the vector of variables in the cube
    s_Temp = ABC_ALLOC( char, dd.size );
    Cudd_bddPickOneCube( dd, bF, s_Temp );

    // start the cube
    bCube = b1; Cudd_Ref( bCube );
    for  v = 0; v < dd.size; v++  {
        if s_Temp[v] == 0 {
//          Cube &= !s_XVars[v];
            // inlining of macro Cudd_Not
            bCube = Cudd_bddAnd( dd, bTemp = bCube, ((DdNode *)((ptrint)(dd.vars[v]) ^ 01)) ); Cudd_Ref( bCube );
            Cudd_RecursiveDeref( dd, bTemp );
        
        } else if s_Temp[v] == 1 {
//          Cube &=  s_XVars[v];
            bCube = Cudd_bddAnd( dd, bTemp = bCube,          dd.vars[v]  ); Cudd_Ref( bCube );
            Cudd_RecursiveDeref( dd, bTemp );
        }
    }
    Cudd_Deref(bCube);
    ABC_FREE( s_Temp );
    return bCube;
}

/**Function********************************************************************

  Synopsis    [Returns one cube contained in the given BDD.]

  Description [This function returns the cube with the smallest 
  bits-to-integer value.]

  SideEffects []

******************************************************************************/
func Extra_bddGetOneCube(dd *DdManager, bFunc *DdNode) *DdNode {
    var bFuncR, bFunc0, bFunc1 *DdNode;
    var bRes0, bRes1, bRes *DdNode;

    // inlining of macro Cudd_Regular
    bFuncR = ((DdNode *)((ptruint)(bFunc) & ~01));
    // inlining of macro cuddIsConstant
    if  (bFuncR).index == CUDD_CONST_INDEX {
        return bFunc; }

    // cofactor
    // inlining of macro Cudd_IsComplement
    if  int32( ((ptrint) (bFunc) & 01)) != 0 {
        // inlining of macro Cudd_Not
        // inlining of macro cuddE
        bFunc0 = ((DdNode *)((ptrint)(((bFuncR).type.kids.E)) ^ 01));
        // inlining of macro Cudd_Not
        // inlining of macro cuddT
        bFunc1 = ((DdNode *)((ptrint)(((bFuncR).type.kids.T)) ^ 01));
    
    } else {
        // inlining of macro cuddE
        bFunc0 = ((bFuncR).type.kids.E);
        // inlining of macro cuddT
        bFunc1 = ((bFuncR).type.kids.T);
    }

    // try to find the cube with the negative literal
    bRes0 = Extra_bddGetOneCube( dd, bFunc0 );  Cudd_Ref( bRes0 );

    if bRes0 != b0 {
        // inlining of macro Cudd_Not
        bRes = Cudd_bddAnd( dd, bRes0, ((DdNode *)((ptrint)(dd.vars[bFuncR.index]) ^ 01)) ); Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bRes0 );
    
    } else {
        Cudd_RecursiveDeref( dd, bRes0 );
        // try to find the cube with the positive literal
        bRes1 = Extra_bddGetOneCube( dd, bFunc1 );  Cudd_Ref( bRes1 );
        assert(bRes1 != b0, "bRes1 != b0");
        bRes = Cudd_bddAnd( dd, bRes1, dd.vars[bFuncR.index] ); Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bRes1 );
    }

    Cudd_Deref( bRes );
    return bRes;
}

/**Function********************************************************************

  Synopsis    [Performs the reordering-sensitive step of Extra_bddMove().]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddComputeRangeCube(dd *DdManager, iStart int32, iStop int32) *DdNode {
    var bTemp, bProd *DdNode;
    var i int32;
    assert(iStart <= iStop, "iStart <= iStop");
    assert(iStart >= 0 && iStart <= dd.size, "iStart >= 0 && iStart <= dd.size");
    assert(iStop >= 0  && iStop  <= dd.size, "iStop >= 0  && iStop  <= dd.size");
    bProd = b1;         Cudd_Ref( bProd );
    for  i = iStart; i < iStop; i++  {
        bProd = Cudd_bddAnd( dd, bTemp = bProd, dd.vars[i] );      Cudd_Ref( bProd );
        Cudd_RecursiveDeref( dd, bTemp ); 
    }
    Cudd_Deref( bProd );
    return bProd;
}

/**Function********************************************************************

  Synopsis    [Computes the cube of BDD variables corresponding to bits it the bit-code]

  Description [Returns a bdd composed of elementary bdds found in array BddVars[] such 
  that the bdd vars encode the number Value of bit length CodeWidth (if fMsbFirst is 1, 
  the most significant bit is encoded with the first bdd variable). If the variables 
  BddVars are not specified, takes the first CodeWidth variables of the manager]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddBitsToCube(dd *DdManager, Code int32, CodeWidth int32, pbVars **DdNode, fMsbFirst int32) *DdNode {
    var z int32;
    var bTemp, bVar, bVarBdd, bResult *DdNode;

    bResult = b1;   Cudd_Ref( bResult );
    for  z = 0; z < CodeWidth; z++  {
        if pbVars {
            bVarBdd = pbVars[z];
        } else {
            bVarBdd = dd.vars[z];
        }
        if fMsbFirst != 0 {
            // inlining of macro Cudd_NotCond
            bVar = ((DdNode *)((ptrint)(bVarBdd) ^ ((Code & (1 << (CodeWidth-1-z)))==0)));
        } else {
            // inlining of macro Cudd_NotCond
            bVar = ((DdNode *)((ptrint)(bVarBdd) ^ ((Code & (1 << (z)))==0))); }
        bResult = Cudd_bddAnd( dd, bTemp = bResult, bVar );     Cudd_Ref( bResult );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bResult );

    return bResult;
}  /* end of Extra_bddBitsToCube */

/**Function********************************************************************

  Synopsis    [Finds the support as a negative polarity cube.]

  Description [Finds the variables on which a DD depends. Returns a BDD 
  consisting of the product of the variables in the negative polarity 
  if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_VectorSupport Cudd_Support]

******************************************************************************/
func Extra_bddSupportNegativeCube(dd *DdManager, f *DdNode) *DdNode {
    var support *int32;
    var res, tmp, var *DdNode;
    var i, j int32;
    var size int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    support = ABC_ALLOC( int, size );
    if support == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return ( nil );
    }
    for  i = 0; i < size; i++  {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    // inlining of macro Cudd_Regular
    ddSupportStep2( ((DdNode *)((ptruint)(f) & ~01)), support );
    // inlining of macro Cudd_Regular
    ddClearFlag2( ((DdNode *)((ptruint)(f) & ~01)) );

    /* Transform support from array to cube. */
    for
    {
        dd.reordered = 0;
        // inlining of macro DD_ONE
        res = ((dd).one);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        for  j = size - 1; j >= 0; j--  {                        /* for each level bottom-up */
            if j >= dd.size {
                i = j;
            } else {
                i = dd.invperm[j];
            }
            if support[i] == 1 {
                // inlining of macro Cudd_Not
                var = cuddUniqueInter( dd, i, dd.one, ((DdNode *)((ptrint)(dd.one) ^ 01)) );
                //////////////////////////////////////////////////////////////////
                // inlining of macro Cudd_Not
                var = ((DdNode *)((ptrint)(var) ^ 01));
                //////////////////////////////////////////////////////////////////
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(var).ref);
                tmp = cuddBddAndRecur( dd, res, var );
                if tmp == nil {
                    Cudd_RecursiveDeref( dd, res );
                    Cudd_RecursiveDeref( dd, var );
                    res = nil;
                    break;
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(tmp).ref);
                Cudd_RecursiveDeref( dd, res );
                Cudd_RecursiveDeref( dd, var );
                res = tmp;
            }
        }
    
if !( dd.reordered == 1 ) { break }
}
    

    ABC_FREE( support );
    if res != nil {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref); }
    return ( res );

}                                /* end of Extra_SupportNeg */

/**Function********************************************************************

  Synopsis    [Returns 1 if the BDD is the BDD of elementary variable.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddIsVar(bFunc *DdNode) int32 {
    // inlining of macro Cudd_Regular
    bFunc = ((DdNode *)((ptruint)(bFunc) & ~01));
    // inlining of macro cuddIsConstant
    if  (bFunc).index == CUDD_CONST_INDEX {
        return 0; }
    // inlining of macro cuddIsConstant
    // inlining of macro cuddT
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    return ((((bFunc).type.kids.T)).index == CUDD_CONST_INDEX) && ((((DdNode *)((ptruint)(((bFunc).type.kids.E)) & ~01))).index == CUDD_CONST_INDEX);
}

/**Function********************************************************************

  Synopsis    [Creates AND composed of the first nVars of the manager.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddCreateAnd(dd *DdManager, nVars int32) *DdNode {
    var bFunc, bTemp *DdNode;
    var i int32;
    bFunc = Cudd_ReadOne(dd); Cudd_Ref( bFunc );
    for  i = 0; i < nVars; i++  {
        bFunc = Cudd_bddAnd( dd, bTemp = bFunc, Cudd_bddIthVar(dd,i) );  Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bFunc );
    return bFunc;
}

/**Function********************************************************************

  Synopsis    [Creates OR composed of the first nVars of the manager.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddCreateOr(dd *DdManager, nVars int32) *DdNode {
    var bFunc, bTemp *DdNode;
    var i int32;
    bFunc = Cudd_ReadLogicZero(dd); Cudd_Ref( bFunc );
    for  i = 0; i < nVars; i++  {
        bFunc = Cudd_bddOr( dd, bTemp = bFunc, Cudd_bddIthVar(dd,i) );  Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bFunc );
    return bFunc;
}

/**Function********************************************************************

  Synopsis    [Creates EXOR composed of the first nVars of the manager.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddCreateExor(dd *DdManager, nVars int32) *DdNode {
    var bFunc, bTemp *DdNode;
    var i int32;
    bFunc = Cudd_ReadLogicZero(dd); Cudd_Ref( bFunc );
    for  i = 0; i < nVars; i++  {
        bFunc = Cudd_bddXor( dd, bTemp = bFunc, Cudd_bddIthVar(dd,i) );  Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bFunc );
    return bFunc;
}

/**Function********************************************************************

  Synopsis    [Computes the set of primes as a ZDD.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_zddPrimes(dd *DdManager, F *DdNode) *DdNode {
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraZddPrimes(dd, F);
        if dd.reordered == 1 {
            fmt.Printf("\nReordering in Extra_zddPrimes()\n"); }
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddPrimes */

/**Function********************************************************************

  Synopsis    [Permutes the variables of the array of BDDs.]

  Description [Given a permutation in array permut, creates a new BDD
  with permuted variables. There should be an entry in array permut
  for each variable in the manager. The i-th entry of permut holds the
  index of the variable that is to substitute the i-th variable.
  The DDs in the resulting array are already referenced.]

  SideEffects [None]

  SeeAlso     [Cudd_addPermute Cudd_bddSwapVariables]

******************************************************************************/
func Extra_bddPermuteArray(manager *DdManager, bNodesIn **DdNode, bNodesOut **DdNode, nNodes int32, permut *int32) {
    var table *DdHashTable;
    var i, k int32;
    for
    {
        manager.reordered = 0;
        table = cuddHashTableInit( manager, 1, 2 );

        /* permute the output functions one-by-one */
        for  i = 0; i < nNodes; i++  {
            bNodesOut[i] = cuddBddPermuteRecur( manager, table, bNodesIn[i], permut );
            if bNodesOut[i] == nil {
                /* deref the array of the already computed outputs */
                for  k = 0; k < i; k++  {
                    Cudd_RecursiveDeref( manager, bNodesOut[k] );
                }
                break;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(bNodesOut[i]).ref);
        }
        /* Dispose of local cache. */
        cuddHashTableQuit( table );
    
if !( manager.reordered == 1 ) { break }
}
    
}    /* end of Extra_bddPermuteArray */


/**Function********************************************************************

  Synopsis    [Computes the positive polarty cube composed of the first vars in the array.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_bddComputeCube(dd *DdManager, bXVars **DdNode, nVars int32) *DdNode {
    var bRes *DdNode;
    var bTemp *DdNode;
    var i int32;

    bRes = b1; Cudd_Ref( bRes );
    for  i = 0; i < nVars; i++  {
        bRes = Cudd_bddAnd( dd, bTemp = bRes, bXVars[i] );  Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bTemp );
    }

    Cudd_Deref( bRes );
    return bRes;
}

/**Function********************************************************************

  Synopsis    [Changes the polarity of vars listed in the cube.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_bddChangePolarity( 
  DdManager * dd,   /* the DD manager */
  DdNode * bFunc,
  DdNode * bVars) 
{
    var res *DdNode;
    for {
        dd.reordered = 0;
        res = extraBddChangePolarity( dd, bFunc, bVars );
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_bddChangePolarity */


/**Function*************************************************************

  Synopsis    [Checks if the given variable belongs to the cube.]

  Description [Return -1 if the var does not appear in the cube.
  Otherwise, returns polarity (0 or 1) of the var in the cube.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_bddVarIsInCube(bCube *DdNode, iVar int32) int32 {
    var bCube0, bCube1 *DdNode;
    // inlining of macro Cudd_Regular
    for  ((DdNode *)((ptruint)(bCube) & ~01)).index != CUDD_CONST_INDEX  {
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddE
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        bCube0 = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(bCube) & ~01))).type.kids.E)) ^ ((int32( ((ptrint) (bCube) & 01))))));
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        bCube1 = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(bCube) & ~01))).type.kids.T)) ^ ((int32( ((ptrint) (bCube) & 01))))));
        // make sure it is a cube
        // inlining of macro Cudd_IsComplement
        // inlining of macro Cudd_Regular
        assert(((int32( ((ptrint) (bCube0) & 01))) && ((DdNode *)((ptruint)(bCube0) & ~01)).index == CUDD_CONST_INDEX) || // bCube0 == 0
                        // inlining of macro Cudd_IsComplement
                        // inlining of macro Cudd_Regular
                        ((int32( ((ptrint) (bCube1) & 01))) && ((DdNode *)((ptruint)(bCube1) & ~01)).index == CUDD_CONST_INDEX), "(/*__starting_point_inline_macro:Cudd_IsComplement__*/(int32( ((ptrint) (bCube0) & 01))) && /*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(bCube0) & ~01)).index == CUDD_CONST_INDEX) || // bCube0 == 0\n                (/*__starting_point_inline_macro:Cudd_IsComplement__*/(int32( ((ptrint) (bCube1) & 01))) && /*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(bCube1) & ~01)).index == CUDD_CONST_INDEX)"); // bCube1 == 0
        // quit if it is the last one
        // inlining of macro Cudd_Regular
        if  ((DdNode *)((ptruint)(bCube) & ~01)).index == iVar {
            // inlining of macro Cudd_IsComplement
            // inlining of macro Cudd_Regular
            return int32(((int32( ((ptrint) (bCube0) & 01))) && ((DdNode *)((ptruint)(bCube0) & ~01)).index == CUDD_CONST_INDEX)); }
        // get the next cube
        // inlining of macro Cudd_IsComplement
        // inlining of macro Cudd_Regular
        if (int32( ((ptrint) (bCube0) & 01))) && ((DdNode *)((ptruint)(bCube0) & ~01)).index == CUDD_CONST_INDEX {
            bCube = bCube1;
        } else {
            bCube = bCube0; }
    }
    return -1;
}
 
/**Function*************************************************************

  Synopsis    [Computes the AND of two BDD with different orders.]

  Description [Derives the result of Boolean AND of bF and bG in ddF.
  The array pPermute gives the mapping of variables of ddG into that of ddF.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_bddAndPermute(ddF *DdManager, bF *DdNode, ddG *DdManager, bG *DdNode, pPermute *int32) *DdNode {
    var table *DdHashTable;
    var bRes *DdNode;
    for
    {
        ddF.reordered = 0;
        table = cuddHashTableInit( ddF, 2, 256 );
        if table == nil { return nil; }
        bRes = extraBddAndPermute( table, ddF, bF, ddG, bG, pPermute );  
        // inlining of macro cuddRef
        if bRes != nil { cuddSatInc(Cudd_Regular(bRes).ref); }
        cuddHashTableQuit( table );
        // inlining of macro cuddDeref
        if bRes != nil { cuddSatDec(Cudd_Regular(bRes).ref); }
//if ( ddF->reordered == 1 )
//printf( "Reordering happened\n" );
    
if !( ddF.reordered == 1 ) { break }
}
    
//printf( "|F| =%6d  |G| =%6d  |H| =%6d  |F|*|G| =%9d\n", 
//       Cudd_DagSize(bF), Cudd_DagSize(bG), Cudd_DagSize(bRes), 
//       Cudd_DagSize(bF) * Cudd_DagSize(bG) );
    return ( bRes );
}

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Performs the reordering-sensitive step of Extra_bddMove().]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraBddMove( 
  DdManager * dd,    /* the DD manager */
  DdNode * bF,
  DdNode * bDist) 
{
    var bRes *DdNode;

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if  (((DdNode *)((ptruint)(bF) & ~01))).index == CUDD_CONST_INDEX {
        return bF; }

    if bRes = cuddCacheLookup2(dd, extraBddMove, bF, bDist) {
        return bRes;
    } else {
        var bRes0, bRes1 *DdNode;             
        var bF0, bF1 *DdNode;             
        // inlining of macro Cudd_Regular
        var bFR *DdNode  = ((DdNode *)((ptruint)(bF) & ~01)); 
        var VarNew int32;
        
        // inlining of macro Cudd_IsComplement
        if  int32( ((ptrint) (bDist) & 01)) != 0 {
            // inlining of macro Cudd_Not
            VarNew = bFR.index - ((DdNode *)((ptrint)(bDist) ^ 01)).index;
        } else {
            VarNew = bFR.index + bDist.index; }
        assert(VarNew < dd.size, "VarNew < dd.size");

        // cofactor the functions
        if bFR != bF { // bFunc is complemented 
        
            // inlining of macro Cudd_Not
            // inlining of macro cuddE
            bF0 = ((DdNode *)((ptrint)(((bFR).type.kids.E)) ^ 01));
            // inlining of macro Cudd_Not
            // inlining of macro cuddT
            bF1 = ((DdNode *)((ptrint)(((bFR).type.kids.T)) ^ 01));
        
        } else {
            // inlining of macro cuddE
            bF0 = ((bFR).type.kids.E);
            // inlining of macro cuddT
            bF1 = ((bFR).type.kids.T);
        }

        bRes0 = extraBddMove( dd, bF0, bDist );
        if bRes0 == nil { 
            return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes0).ref);

        bRes1 = extraBddMove( dd, bF1, bDist );
        if bRes1 == nil {
            Cudd_RecursiveDeref( dd, bRes0 );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes1).ref);

        /* only bRes0 and bRes1 are referenced at this point */
        bRes = cuddBddIteRecur( dd, dd.vars[VarNew], bRes1, bRes0 );
        if bRes == nil {
            Cudd_RecursiveDeref( dd, bRes0 );
            Cudd_RecursiveDeref( dd, bRes1 );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes).ref);
        Cudd_RecursiveDeref( dd, bRes0 );
        Cudd_RecursiveDeref( dd, bRes1 );

        /* insert the result into cache */
        cuddCacheInsert2( dd, extraBddMove, bF, bDist, bRes );
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(bRes).ref);
        return bRes;
    }
} /* end of extraBddMove */


/**Function********************************************************************

  Synopsis    [Finds three cofactors of the cover w.r.t. to the topmost variable.]

  Description [Finds three cofactors of the cover w.r.t. to the topmost variable.
  Does not check the cover for being a constant. Assumes that ZDD variables encoding 
  positive and negative polarities are adjacent in the variable order. Is different 
  from cuddZddGetCofactors3() in that it does not compute the cofactors w.r.t. the 
  given variable but takes the cofactors with respent to the topmost variable. 
  This function is more efficient when used in recursive procedures because it does 
  not require referencing of the resulting cofactors (compare cuddZddProduct() 
  and extraZddPrimeProduct()).]

  SideEffects [None]

  SeeAlso     [cuddZddGetCofactors3]

******************************************************************************/
void 
extraDecomposeCover( 
  DdManager* dd,    /* the manager */
  DdNode*  zC,      /* the cover */
  DdNode** zC0,     /* the pointer to the negative var cofactor */ 
  DdNode** zC1,     /* the pointer to the positive var cofactor */ 
  DdNode** zC2 )    /* the pointer to the cofactor without var */ 
{
    if (zC.index & 1) == 0 { /* the top variable is present in positive polarity and maybe in negative */

        // inlining of macro cuddE
        var Temp *DdNode  = ((zC).type.kids.E);
        // inlining of macro cuddT
        *zC1  = ((zC).type.kids.T);
        // inlining of macro cuddIZ
        if  (((Temp.index)==CUDD_CONST_INDEX)?int32((Temp.index)):(dd).permZ[(Temp.index)]) == (((zC.index)==CUDD_CONST_INDEX)?int32((zC.index)):(dd).permZ[(zC.index)]) + 1 {   /* Temp is not a terminal node 
             * top var is present in negative polarity */
            // inlining of macro cuddE
            *zC2 = ((Temp).type.kids.E);
            // inlining of macro cuddT
            *zC0 = ((Temp).type.kids.T);
        
        } else {   /* top var is not present in negative polarity */
            *zC2 = Temp;
            *zC0 = dd.zero;
        }
    
    } else { /* the top variable is present only in negative */
        *zC1 = dd.zero;
        // inlining of macro cuddE
        *zC2 = ((zC).type.kids.E);
        // inlining of macro cuddT
        *zC0 = ((zC).type.kids.T);
    }
} /* extraDecomposeCover */



/**Function********************************************************************

  Synopsis    [Counts the total number of cubes in the ISOPs of the functions.]

  Description [Returns -1 if the number of cubes exceeds the limit.]

  SideEffects [None]

  SeeAlso     [Extra_TransferPermute]

******************************************************************************/
func extraBddCountCubes(dd *DdManager, L *DdNode, U *DdNode, table *st__table, pnCubes *int32, Limit int32) *DdNode {
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro Cudd_Not
    var zero *DdNode  = ((DdNode *)((ptrint)(one) ^ 01));
    var v, top_l, top_u int32;
    var Lsub0, Usub0, Lsub1, Usub1, Ld, Ud *DdNode;
    var Lsuper0, Usuper0, Lsuper1, Usuper1 *DdNode;
    var Isub0, Isub1, Id *DdNode;
    var x *DdNode;
    var term0, term1, sum *DdNode;
    var Lv, Uv, Lnv, Unv *DdNode;
    var r *DdNode;
    var index int32;
    var Count0, Count1, Count2 int32;

    statLine(dd);
    if L == zero {
        *pnCubes = 0;
        return(zero);
    }
    if U == one {
        *pnCubes = 1;
        return(one);
    }

    /* Check cache */
    r = cuddCacheLookup2(dd, cuddBddIsop, L, U);
    if r != nil {
        var nCubes int32;
        if st__lookup_int( table, (char *)r, &nCubes ) {
            *pnCubes = nCubes;
        } else { assert(false, "0"); }
        return r;
    }

    // inlining of macro Cudd_Regular
    top_l = dd.perm[((DdNode *)((ptruint)(L) & ~01)).index];
    // inlining of macro Cudd_Regular
    top_u = dd.perm[((DdNode *)((ptruint)(U) & ~01)).index];
    // inlining of macro ddMin
    
    if (top_u) < (top_l) {
        v = (top_u);
    } else {
        v = (top_l);
    }

    /* Compute cofactors */
    if top_l == v {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(L) & ~01)).index;
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        Lv = ((((DdNode *)((ptruint)(L) & ~01))).type.kids.T);
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        Lnv = ((((DdNode *)((ptruint)(L) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (L) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Lv = ((DdNode *)((ptrint)(Lv) ^ 01));
            // inlining of macro Cudd_Not
            Lnv = ((DdNode *)((ptrint)(Lnv) ^ 01));
        }
    
    } else {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(U) & ~01)).index;
        Lv = Lnv = L;
    }

    if top_u == v {
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        Uv = ((((DdNode *)((ptruint)(U) & ~01))).type.kids.T);
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        Unv = ((((DdNode *)((ptruint)(U) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (U) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Uv = ((DdNode *)((ptrint)(Uv) ^ 01));
            // inlining of macro Cudd_Not
            Unv = ((DdNode *)((ptrint)(Unv) ^ 01));
        }
    
    } else {
        Uv = Unv = U;
    }

    // inlining of macro Cudd_Not
    Lsub0 = cuddBddAndRecur(dd, Lnv, ((DdNode *)((ptrint)(Uv) ^ 01)));
    if Lsub0 == nil {
        return(nil); }
    Cudd_Ref(Lsub0);
    Usub0 = Unv;
    // inlining of macro Cudd_Not
    Lsub1 = cuddBddAndRecur(dd, Lv, ((DdNode *)((ptrint)(Unv) ^ 01)));
    if Lsub1 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        return(nil);
    }
    Cudd_Ref(Lsub1);
    Usub1 = Uv;

    Isub0 = extraBddCountCubes(dd, Lsub0, Usub0, table, &Count0, Limit);
    if Isub0 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        Cudd_RecursiveDeref(dd, Lsub1);
        return(nil);
    }
    Cudd_Ref(Isub0);
    Isub1 = extraBddCountCubes(dd, Lsub1, Usub1, table, &Count1, Limit);
    if Isub1 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        Cudd_RecursiveDeref(dd, Lsub1);
        Cudd_RecursiveDeref(dd, Isub0);
        return(nil);
    }
    Cudd_Ref(Isub1);
    Cudd_RecursiveDeref(dd, Lsub0);
    Cudd_RecursiveDeref(dd, Lsub1);

    // inlining of macro Cudd_Not
    Lsuper0 = cuddBddAndRecur(dd, Lnv, ((DdNode *)((ptrint)(Isub0) ^ 01)));
    if Lsuper0 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        return(nil);
    }
    Cudd_Ref(Lsuper0);
    // inlining of macro Cudd_Not
    Lsuper1 = cuddBddAndRecur(dd, Lv, ((DdNode *)((ptrint)(Isub1) ^ 01)));
    if Lsuper1 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        return(nil);
    }
    Cudd_Ref(Lsuper1);
    Usuper0 = Unv;
    Usuper1 = Uv;

    /* Ld = Lsuper0 + Lsuper1 */
    // inlining of macro Cudd_Not
    Ld = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(Lsuper0) ^ 01)), ((DdNode *)((ptrint)(Lsuper1) ^ 01)));
    // inlining of macro Cudd_NotCond
    Ld = ((DdNode *)((ptrint)(Ld) ^ (Ld != nil)));
    if Ld == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        Cudd_RecursiveDeref(dd, Lsuper1);
        return(nil);
    }
    Cudd_Ref(Ld);
    Ud = cuddBddAndRecur(dd, Usuper0, Usuper1);
    if Ud == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        Cudd_RecursiveDeref(dd, Lsuper1);
        Cudd_RecursiveDeref(dd, Ld);
        return(nil);
    }
    Cudd_Ref(Ud);
    Cudd_RecursiveDeref(dd, Lsuper0);
    Cudd_RecursiveDeref(dd, Lsuper1);

    Id = extraBddCountCubes(dd, Ld, Ud, table, &Count2, Limit);
    if Id == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Ld);
        Cudd_RecursiveDeref(dd, Ud);
        return(nil);
    }
    Cudd_Ref(Id);
    Cudd_RecursiveDeref(dd, Ld);
    Cudd_RecursiveDeref(dd, Ud);

    x = cuddUniqueInter(dd, index, one, zero);
    if x == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Id);
        return(nil);
    }
    Cudd_Ref(x);
    // inlining of macro Cudd_Not
    term0 = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(x) ^ 01)), Isub0);
    if term0 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, x);
        return(nil);
    }
    Cudd_Ref(term0);
    Cudd_RecursiveDeref(dd, Isub0);
    term1 = cuddBddAndRecur(dd, x, Isub1);
    if term1 == nil {
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, x);
        Cudd_RecursiveDeref(dd, term0);
        return(nil);
    }
    Cudd_Ref(term1);
    Cudd_RecursiveDeref(dd, x);
    Cudd_RecursiveDeref(dd, Isub1);
    /* sum = term0 + term1 */
    // inlining of macro Cudd_Not
    sum = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(term0) ^ 01)), ((DdNode *)((ptrint)(term1) ^ 01)));
    // inlining of macro Cudd_NotCond
    sum = ((DdNode *)((ptrint)(sum) ^ (sum != nil)));
    if sum == nil {
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, term0);
        Cudd_RecursiveDeref(dd, term1);
        return(nil);
    }
    Cudd_Ref(sum);
    Cudd_RecursiveDeref(dd, term0);
    Cudd_RecursiveDeref(dd, term1);
    /* r = sum + Id */
    // inlining of macro Cudd_Not
    r = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(sum) ^ 01)), ((DdNode *)((ptrint)(Id) ^ 01)));
    // inlining of macro Cudd_NotCond
    r = ((DdNode *)((ptrint)(r) ^ (r != nil)));
    if r == nil {
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, sum);
        return(nil);
    }
    Cudd_Ref(r);
    Cudd_RecursiveDeref(dd, sum);
    Cudd_RecursiveDeref(dd, Id);

    cuddCacheInsert2(dd, cuddBddIsop, L, U, r);
    *pnCubes = Count0 + Count1 + Count2;
    if st__add_direct( table, (char *)r, (char *)(ABC_PTRINT_T)*pnCubes ) == st__OUT_OF_MEM {
        Cudd_RecursiveDeref( dd, r );
        return nil;
    }
    if *pnCubes > Limit {
        Cudd_RecursiveDeref( dd, r );
        return nil;
    }
    //printf( "%d ", *pnCubes );
    Cudd_Deref(r);
    return r;
} 
func Extra_bddCountCubes(dd *DdManager, pFuncs **DdNode, nFuncs int32, fMode int32, nLimit int32, pGuide *int32) int32 {
    var pF0, pF1 *DdNode; 
    var i, Count, Count0, Count1, CounterAll int32;
    var table *st__table  = st__init_table( st__ptrcmp, st__ptrhash );
    unsigned int saveLimit = dd.maxLive;
    dd.maxLive = dd.keys - dd.dead + 1000000; // limit on intermediate BDD nodes
    for  i = 0; i < nFuncs; i++  {
        if !pFuncs[i] {
            continue; }
        pF1 = pF0 = nil;
        Count0 = Count1 = nLimit;
        if fMode == -1 || fMode == 1 { // both or pos
            pF1 = extraBddCountCubes( dd, pFuncs[i], pFuncs[i], table, &Count1, nLimit ); }
        // inlining of macro Cudd_Not
        pFuncs[i] = ((DdNode *)((ptrint)(pFuncs[i]) ^ 01));
        if fMode == -1 || fMode == 0 { // both or neg
            pF0 = extraBddCountCubes( dd, pFuncs[i], pFuncs[i], table, &Count0, Count1 ); }
        // inlining of macro Cudd_Not
        pFuncs[i] = ((DdNode *)((ptrint)(pFuncs[i]) ^ 01));
        if !(pF1 != nil) && !pF0 {
            break; }
        if !pF0 {                  pGuide[i] = 1, Count = Count1; // use pos
        } else if !(pF1 != nil) {             pGuide[i] = 0, Count = Count0; // use neg
        } else if Count1 <= Count0 { pGuide[i] = 1, Count = Count1; // use pos
        } else {                         pGuide[i] = 0, Count = Count0; } // use neg
        CounterAll += Count;
        //printf( "Output %5d has %5d cubes (%d)  (%5d and %5d)\n", nOuts++, Count, pGuide[i], Count1, Count0 );
    }
    dd.maxLive = saveLimit;
    st__free_table( table );
    if i == nFuncs {
        return CounterAll;
    } else {
        return -1;
    }
}   /* end of Extra_bddCountCubes */

/*---------------------------------------------------------------------------*/
/* Definition of static Functions                                            */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Convert a BDD from a manager to another one.]

  Description [Convert a BDD from a manager to another one. Returns a
  pointer to the BDD in the destination manager if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Extra_TransferPermute]

******************************************************************************/
func extraTransferPermute(ddS *DdManager, ddD *DdManager, f *DdNode, Permute *int32) *DdNode {
    var res *DdNode;
    var table *st__table  = nil;
    var gen *st__generator  = nil;
    var key, value *DdNode;

    table = st__init_table( st__ptrcmp, st__ptrhash );
    if table == nil {
        goto failure; }
    res = extraTransferPermuteRecur( ddS, ddD, f, table, Permute );
    if res != nil {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref); }

    /* Dereference all elements in the table and dispose of the table.
       ** This must be done also if res is NULL to avoid leaks in case of
       ** reordering. */
    gen = st__init_gen( table );
    if gen == nil {
        goto failure; }
    for  st__gen( gen, ( char ** ) &key, ( char ** ) &value )  {
        Cudd_RecursiveDeref( ddD, value );
    }
    st__free_gen( gen );
    gen = nil;
    st__free_table( table );
    table = nil;

    if res != nil {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref); }
    return ( res );

  failure:
    if table != nil {
        st__free_table( table ); }
    if gen != nil {
        st__free_gen( gen ); }
    return ( nil );

}                               /* end of extraTransferPermute */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Extra_TransferPermute.]

  Description [Performs the recursive step of Extra_TransferPermute.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [extraTransferPermute]

******************************************************************************/
func extraTransferPermuteRecur(ddS *DdManager, ddD *DdManager, f *DdNode, table *st__table, Permute *int32) *DdNode {
    var ft, fe, t, e, var, res *DdNode;
    var one, zero *DdNode;
    var index int32;
    var comple int32;

    statLine( ddD );
    // inlining of macro DD_ONE
    one = ((ddD).one);
    // inlining of macro Cudd_IsComplement
    comple = (int32( ((ptrint) (f) & 01)));

    /* Trivial cases. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if  (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_NotCond
        return ( ((DdNode *)((ptrint)(one) ^ (comple))) ); }


    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_NotCond
    f = ((DdNode *)((ptrint)(f) ^ (comple)));
    /* Now f is a regular pointer to a non-constant node. */

    /* Check the cache. */
    if st__lookup( table, ( char * ) f, ( char ** ) &res ) {
        // inlining of macro Cudd_NotCond
        return ( ((DdNode *)((ptrint)(res) ^ (comple))) ); }

    if ddS.TimeStop && Abc_Clock() > ddS.TimeStop {
        return nil; }
    if ddD.TimeStop && Abc_Clock() > ddD.TimeStop {
        return nil; }

    /* Recursive step. */
    if Permute != nil {
        index = Permute[f.index];
    } else {
        index = f.index; }

    // inlining of macro cuddT
    ft = ((f).type.kids.T);
    // inlining of macro cuddE
    fe = ((f).type.kids.E);

    t = extraTransferPermuteRecur( ddS, ddD, ft, table, Permute );
    if t == nil {
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = extraTransferPermuteRecur( ddS, ddD, fe, table, Permute );
    if e == nil {
        Cudd_RecursiveDeref( ddD, t );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(ddD.one) ^ 01));
    var = cuddUniqueInter( ddD, index, one, zero );
    if var == nil {
        Cudd_RecursiveDeref( ddD, t );
        Cudd_RecursiveDeref( ddD, e );
        return ( nil );
    }
    res = cuddBddIteRecur( ddD, var, t, e );

    if res == nil {
        Cudd_RecursiveDeref( ddD, t );
        Cudd_RecursiveDeref( ddD, e );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref( ddD, t );
    Cudd_RecursiveDeref( ddD, e );

    if st__add_direct( table, ( char * ) f, ( char * ) res ) ==
         st__OUT_OF_MEM {
        Cudd_RecursiveDeref( ddD, res );
        return ( nil );
    }
    // inlining of macro Cudd_NotCond
    return ( ((DdNode *)((ptrint)(res) ^ (comple))) );

}                               /* end of extraTransferPermuteRecur */

/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_Support.]

  Description [Performs the recursive step of Cudd_Support. Performs a
  DFS from f. The support is accumulated in supp as a side effect. Uses
  the LSB of the then pointer as visited flag.]

  SideEffects [None]

  SeeAlso     [ddClearFlag]

******************************************************************************/
func ddSupportStep2(f *DdNode, support *int32) {
    // inlining of macro cuddIsConstant
    // inlining of macro Cudd_IsComplement
    if ((f).index == CUDD_CONST_INDEX) || ((int32( ((ptrint) (f.next) & 01)) != 0)) {
    return;
    }

    support[f.index] = 1;
    // inlining of macro cuddT
    ddSupportStep2(((f).type.kids.T),support);
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    ddSupportStep2(((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)),support);
    /* Mark as visited. */
    // inlining of macro Cudd_Not
    f.next = ((DdNode *)((ptrint)(f.next) ^ 01));
    return;

} /* end of ddSupportStep */


/**Function********************************************************************

  Synopsis    [Performs a DFS from f, clearing the LSB of the next
  pointers.]

  Description []

  SideEffects [None]

  SeeAlso     [ddSupportStep ddDagInt]

******************************************************************************/
func ddClearFlag2(f *DdNode) {
    // inlining of macro Cudd_IsComplement
    if !((int32( ((ptrint) (f.next) & 01)) != 0)) {
    return;
    }
    /* Clear visited flag. */
    // inlining of macro Cudd_Regular
    f.next = ((DdNode *)((ptruint)(f.next) & ~01));
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
    return;
    }
    // inlining of macro cuddT
    ddClearFlag2(((f).type.kids.T));
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    ddClearFlag2(((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)));
    return;

} /* end of ddClearFlag */


/**Function********************************************************************

  Synopsis    [Composed three subcovers into one ZDD.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
DdNode *
extraComposeCover( 
  DdManager* dd,    /* the manager */
  DdNode* zC0,     /* the pointer to the negative var cofactor */ 
  DdNode* zC1,     /* the pointer to the positive var cofactor */ 
  DdNode* zC2,     /* the pointer to the cofactor without var */ 
  int TopVar)    /* the index of the positive ZDD var */
{
    var zRes, zTemp *DdNode;
    /* compose with-neg-var and without-var using the neg ZDD var */
    zTemp = cuddZddGetNode( dd, 2*TopVar + 1, zC0, zC2 );
    if zTemp == nil {
        Cudd_RecursiveDerefZdd(dd, zC0);
        Cudd_RecursiveDerefZdd(dd, zC1);
        Cudd_RecursiveDerefZdd(dd, zC2);
        return nil;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(zTemp).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(zC0).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(zC2).ref);

    /* compose with-pos-var and previous result using the pos ZDD var */
    zRes = cuddZddGetNode( dd, 2*TopVar, zC1, zTemp );
    if zRes == nil {
        Cudd_RecursiveDerefZdd(dd, zC1);
        Cudd_RecursiveDerefZdd(dd, zTemp);
        return nil;
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(zC1).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(zTemp).ref);
    return zRes;
} /* extraComposeCover */

/**Function********************************************************************

  Synopsis    [Performs the recursive step of prime computation.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func extraZddPrimes(dd *DdManager, F *DdNode) *DdNode {
    var zRes *DdNode;

    // inlining of macro Cudd_Not
    if F == ((DdNode *)((ptrint)(dd.one) ^ 01)) {
        return dd.zero; }
    if F == dd.one {
        return dd.one; }

    /* check cache */
    zRes = cuddCacheLookup1Zdd(dd, extraZddPrimes, F);
    if zRes != nil {
        return(zRes); }
    {
        /* temporary variables */
        var bF01, zP0, zP1 *DdNode;
        /* three components of the prime set */
        var zResE, zResP, zResN *DdNode;
        // inlining of macro Cudd_IsComplement
        var fIsComp int32  = (int32( ((ptrint) (F) & 01)));

        /* find cofactors of F */
        // inlining of macro Cudd_NotCond
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        var bF0 *DdNode  = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(F) & ~01))).type.kids.E)) ^ (fIsComp)));
        // inlining of macro Cudd_NotCond
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        var bF1 *DdNode  = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(F) & ~01))).type.kids.T)) ^ (fIsComp)));

        /* find the intersection of cofactors */
        bF01 = cuddBddAndRecur( dd, bF0, bF1 );
        if bF01 == nil { return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bF01).ref);

        /* solve the problems for cofactors */
        zP0 = extraZddPrimes( dd, bF0 );
        if zP0 == nil {
            Cudd_RecursiveDeref( dd, bF01 );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zP0).ref);

        zP1 = extraZddPrimes( dd, bF1 );
        if zP1 == nil {
            Cudd_RecursiveDeref( dd, bF01 );
            Cudd_RecursiveDerefZdd( dd, zP0 );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(zP1).ref);

        /* check for local unateness */
        if bF01 == bF0 {       /* unate increasing */
        
            /* intersection is useless */
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(bF01).ref);
            /* the primes of intersection are the primes of F0 */
            zResE = zP0;
            /* there are no primes with negative var */
            zResN = dd.zero;
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zResN).ref);
            /* primes with positive var are primes of F1 that are not primes of F01 */
            zResP = cuddZddDiff( dd, zP1, zP0 );
            if zResP == nil {
                Cudd_RecursiveDerefZdd( dd, zResE );
                Cudd_RecursiveDerefZdd( dd, zResN );
                Cudd_RecursiveDerefZdd( dd, zP1 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zResP).ref);
            Cudd_RecursiveDerefZdd( dd, zP1 );
        
        } else if bF01 == bF1 { /* unate decreasing */
        
            /* intersection is useless */
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(bF01).ref);
            /* the primes of intersection are the primes of F1 */
            zResE = zP1;
            /* there are no primes with positive var */
            zResP = dd.zero;
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zResP).ref);
            /* primes with negative var are primes of F0 that are not primes of F01 */
            zResN = cuddZddDiff( dd, zP0, zP1 );
            if zResN == nil {
                Cudd_RecursiveDerefZdd( dd, zResE );
                Cudd_RecursiveDerefZdd( dd, zResP );
                Cudd_RecursiveDerefZdd( dd, zP0 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zResN).ref);
            Cudd_RecursiveDerefZdd( dd, zP0 );
        
        } else { /* not unate */
        
            /* primes without the top var are primes of F10 */
            zResE = extraZddPrimes( dd, bF01 );
            if zResE == nil {
                Cudd_RecursiveDerefZdd( dd, bF01 );
                Cudd_RecursiveDerefZdd( dd, zP0 );
                Cudd_RecursiveDerefZdd( dd, zP1 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zResE).ref);
            Cudd_RecursiveDeref( dd, bF01 );

            /* primes with the negative top var are those of P0 that are not in F10 */
            zResN = cuddZddDiff( dd, zP0, zResE );
            if zResN == nil {
                Cudd_RecursiveDerefZdd( dd, zResE );
                Cudd_RecursiveDerefZdd( dd, zP0 );
                Cudd_RecursiveDerefZdd( dd, zP1 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zResN).ref);
            Cudd_RecursiveDerefZdd( dd, zP0 );

            /* primes with the positive top var are those of P1 that are not in F10 */
            zResP = cuddZddDiff( dd, zP1, zResE );
            if zResP == nil {
                Cudd_RecursiveDerefZdd( dd, zResE );
                Cudd_RecursiveDerefZdd( dd, zResN );
                Cudd_RecursiveDerefZdd( dd, zP1 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zResP).ref);
            Cudd_RecursiveDerefZdd( dd, zP1 );
        }

        // inlining of macro Cudd_Regular
        zRes = extraComposeCover( dd, zResN, zResP, zResE, ((DdNode *)((ptruint)(F) & ~01)).index );
        if zRes == nil { return nil; }

        /* insert the result into cache */
        cuddCacheInsert1(dd, extraZddPrimes, F, zRes);
        return zRes;
    }
} /* end of extraZddPrimes */

/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_bddPermute.]

  Description [ Recursively puts the BDD in the order given in the array permut.
  Checks for trivial cases to terminate recursion, then splits on the
  children of this node.  Once the solutions for the children are
  obtained, it puts into the current position the node from the rest of
  the BDD that should be here. Then returns this BDD.
  The key here is that the node being visited is NOT put in its proper
  place by this instance, but rather is switched when its proper position
  is reached in the recursion tree.<p>
  The DdNode * that is returned is the same BDD as passed in as node,
  but in the new order.]

  SideEffects [None]

  SeeAlso     [Cudd_bddPermute cuddAddPermuteRecur]

******************************************************************************/
static DdNode *
cuddBddPermuteRecur( DdManager * manager /* DD manager */ ,
                     DdHashTable * table /* computed table */ ,
                     DdNode * node /* BDD to be reordered */ ,
                     int *permut /* permutation array */  )
{
    var N, T, E *DdNode;
    var res *DdNode;
    var index int32;

    statLine( manager );
    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));

    /* Check for terminal case of constant node. */
    // inlining of macro cuddIsConstant
    if  (N).index == CUDD_CONST_INDEX {
        return ( node );
    }

    /* If problem already solved, look up answer and return. */
    if N.ref != 1 && ( res = cuddHashTableLookup1( table, N ) ) != nil {
        // inlining of macro Cudd_NotCond
        return ( ((DdNode *)((ptrint)(res) ^ (N != node))) );
    }

    /* Split and recur on children of this node. */
    // inlining of macro cuddT
    T = cuddBddPermuteRecur( manager, table, ((N).type.kids.T), permut );
    if T == nil {
        return ( nil ); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);
    // inlining of macro cuddE
    E = cuddBddPermuteRecur( manager, table, ((N).type.kids.E), permut );
    if E == nil {
        Cudd_IterDerefBdd( manager, T );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);

    /* Move variable that should be in this position to this position
       ** by retrieving the single var BDD for that variable, and calling
       ** cuddBddIteRecur with the T and E we just created.
     */
    index = permut[N.index];
    res = cuddBddIteRecur( manager, manager.vars[index], T, E );
    if res == nil {
        Cudd_IterDerefBdd( manager, T );
        Cudd_IterDerefBdd( manager, E );
        return ( nil );
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_IterDerefBdd( manager, T );
    Cudd_IterDerefBdd( manager, E );

    /* Do not keep the result if the reference count is only 1, since
       ** it will not be visited again.
     */
    if N.ref != 1 {
        var fanout ptrint  = ( ptrint ) N.ref;
        cuddSatDec( fanout );
        if !cuddHashTableInsert1( table, N, res, fanout ) {
            Cudd_IterDerefBdd( manager, res );
            return ( nil );
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    // inlining of macro Cudd_NotCond
    return ( ((DdNode *)((ptrint)(res) ^ (N != node))) );

}                                /* end of cuddBddPermuteRecur */


/**Function********************************************************************

  Synopsis    [Performs the reordering-sensitive step of Extra_bddChangePolarity().]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraBddChangePolarity( 
  DdManager * dd,    /* the DD manager */
  DdNode * bFunc, 
  DdNode * bVars) 
{
    var bRes *DdNode;

    if bVars == b1 {
        return bFunc; }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if  (((DdNode *)((ptruint)(bFunc) & ~01))).index == CUDD_CONST_INDEX {
        return bFunc; }

    if bRes = cuddCacheLookup2(dd, extraBddChangePolarity, bFunc, bVars) {
        return bRes;
    } else {
        // inlining of macro Cudd_Regular
        var bFR *DdNode  = ((DdNode *)((ptruint)(bFunc) & ~01)); 
        var LevelF int32    = dd.perm[bFR.index];
        var LevelV int32    = dd.perm[bVars.index];

        if LevelV < LevelF {
            // inlining of macro cuddT
            bRes = extraBddChangePolarity( dd, bFunc, ((bVars).type.kids.T) );
        } else { // if ( LevelF <= LevelV )
        
            var bRes0, bRes1 *DdNode;             
            var bF0, bF1 *DdNode;             
            var bVarsNext *DdNode;

            // cofactor the functions
            if bFR != bFunc { // bFunc is complemented 
            
                // inlining of macro Cudd_Not
                // inlining of macro cuddE
                bF0 = ((DdNode *)((ptrint)(((bFR).type.kids.E)) ^ 01));
                // inlining of macro Cudd_Not
                // inlining of macro cuddT
                bF1 = ((DdNode *)((ptrint)(((bFR).type.kids.T)) ^ 01));
            
            } else {
                // inlining of macro cuddE
                bF0 = ((bFR).type.kids.E);
                // inlining of macro cuddT
                bF1 = ((bFR).type.kids.T);
            }

            if LevelF == LevelV {
                // inlining of macro cuddT
                bVarsNext = ((bVars).type.kids.T);
            } else {
                bVarsNext = bVars; }

            bRes0 = extraBddChangePolarity( dd, bF0, bVarsNext );
            if bRes0 == nil { 
                return nil; }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(bRes0).ref);

            bRes1 = extraBddChangePolarity( dd, bF1, bVarsNext );
            if bRes1 == nil {
                Cudd_RecursiveDeref( dd, bRes0 );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(bRes1).ref);

            if LevelF == LevelV { // swap the cofactors
                var bTemp *DdNode;
                bTemp = bRes0;
                bRes0 = bRes1;
                bRes1 = bTemp;
            }

            /* only aRes0 and aRes1 are referenced at this point */

            /* consider the case when Res0 and Res1 are the same node */
            if bRes0 == bRes1 {
                bRes = bRes1;
            /* consider the case when Res1 is complemented */
            // inlining of macro Cudd_IsComplement
            } else if  int32( ((ptrint) (bRes1) & 01)) != 0 {
                // inlining of macro Cudd_Not
                bRes = cuddUniqueInter(dd, bFR.index, ((DdNode *)((ptrint)(bRes1) ^ 01)), ((DdNode *)((ptrint)(bRes0) ^ 01)));
                if bRes == nil {
                    Cudd_RecursiveDeref(dd,bRes0);
                    Cudd_RecursiveDeref(dd,bRes1);
                    return nil;
                }
                // inlining of macro Cudd_Not
                bRes = ((DdNode *)((ptrint)(bRes) ^ 01));
             
            } else {
                bRes = cuddUniqueInter( dd, bFR.index, bRes1, bRes0 );
                if bRes == nil {
                    Cudd_RecursiveDeref(dd,bRes0);
                    Cudd_RecursiveDeref(dd,bRes1);
                    return nil;
                }
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(bRes0).ref);
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(bRes1).ref);
        }
            
        /* insert the result into cache */
        cuddCacheInsert2(dd, extraBddChangePolarity, bFunc, bVars, bRes);
        return bRes;
    }
} /* end of extraBddChangePolarity */



var Counter int32;

/**Function*************************************************************

  Synopsis    [Computes the AND of two BDD with different orders.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func extraBddAndPermute(table *DdHashTable, ddF *DdManager, bF *DdNode, ddG *DdManager, bG *DdNode, pPermute *int32) *DdNode {
    var bF0, bF1, bG0, bG1, bRes0, bRes1, bRes, bVar *DdNode;
    var LevF, LevG, Lev int32;

    // if F == 0, return 0 
    // inlining of macro Cudd_Not
    if bF == ((DdNode *)((ptrint)(ddF.one) ^ 01)) {
        // inlining of macro Cudd_Not
        return ((DdNode *)((ptrint)(ddF.one) ^ 01)); }
    // if G == 0, return 0 
    // inlining of macro Cudd_Not
    if bG == ((DdNode *)((ptrint)(ddG.one) ^ 01)) {
        // inlining of macro Cudd_Not
        return ((DdNode *)((ptrint)(ddF.one) ^ 01)); }
    // if G == 1, return F
    if bG == ddG.one {
        return bF; }
    // cannot use F == 1, because the var order of G has to be changed

    // check cache
    if  //(Cudd_Regular(bF)->ref != 1 || Cudd_Regular(bG)->ref != 1) && 
         bRes = cuddHashTableLookup2(table, bF, bG) {
        return bRes; }
    Counter++;

    if ddF.TimeStop && Abc_Clock() > ddF.TimeStop {
        return nil; }
    if ddG.TimeStop && Abc_Clock() > ddG.TimeStop {
        return nil; }

    // find the topmost variable in F and G using var order of F
    // inlining of macro cuddI
    
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(bF) & ~01)).index)==CUDD_CONST_INDEX {
        // inlining of macro Cudd_Regular
        LevF = int32((((DdNode *)((ptruint)(bF) & ~01)).index));
    } else {
        // inlining of macro Cudd_Regular
        LevF = (ddF).perm[(((DdNode *)((ptruint)(bF) & ~01)).index)];
    }
    // inlining of macro cuddI
    
    // inlining of macro Cudd_Regular
    if (pPermute ? pPermute[((DdNode *)((ptruint)(bG) & ~01)).index] : ((DdNode *)((ptruint)(bG) & ~01)).index)==CUDD_CONST_INDEX {
        // inlining of macro Cudd_Regular
        LevG = int32((pPermute ? pPermute[((DdNode *)((ptruint)(bG) & ~01)).index] : ((DdNode *)((ptruint)(bG) & ~01)).index));
    } else {
        // inlining of macro Cudd_Regular
        LevG = (ddF).perm[(pPermute ? pPermute[((DdNode *)((ptruint)(bG) & ~01)).index] : ((DdNode *)((ptruint)(bG) & ~01)).index)];
    }
    Lev  = Abc_MinInt( LevF, LevG );
    assert(Lev < ddF.size, "Lev < ddF.size");
    bVar = ddF.vars[ddF.invperm[Lev]];

    // cofactor
    if Lev < LevF {
        bF0 = bF;
    } else {
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddE
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        bF0 = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(bF) & ~01))).type.kids.E)) ^ ((int32( ((ptrint) (bF) & 01))))));
    }
    if Lev < LevF {
        bF1 = bF;
    } else {
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        bF1 = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(bF) & ~01))).type.kids.T)) ^ ((int32( ((ptrint) (bF) & 01))))));
    }
    if Lev < LevG {
        bG0 = bG;
    } else {
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddE
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        bG0 = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(bG) & ~01))).type.kids.E)) ^ ((int32( ((ptrint) (bG) & 01))))));
    }
    if Lev < LevG {
        bG1 = bG;
    } else {
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        // inlining of macro Cudd_IsComplement
        bG1 = ((DdNode *)((ptrint)(((((DdNode *)((ptruint)(bG) & ~01))).type.kids.T)) ^ ((int32( ((ptrint) (bG) & 01))))));
    }

    // call for cofactors
    bRes0 = extraBddAndPermute( table, ddF, bF0, ddG, bG0, pPermute );
    if bRes0 == nil { 
        return nil; }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(bRes0).ref);
    // call for cofactors
    bRes1 = extraBddAndPermute( table, ddF, bF1, ddG, bG1, pPermute );
    if bRes1 == nil {
        Cudd_IterDerefBdd( ddF, bRes0 );
        return nil;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(bRes1).ref);

    // compose the result
    bRes = cuddBddIteRecur( ddF, bVar, bRes1, bRes0 );
    if bRes == nil {
        Cudd_IterDerefBdd( ddF, bRes0 );
        Cudd_IterDerefBdd( ddF, bRes1 );
        return nil;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(bRes).ref);
    Cudd_IterDerefBdd( ddF, bRes0 );
    Cudd_IterDerefBdd( ddF, bRes1 );

    // cache the result
//    if ( Cudd_Regular(bF)->ref != 1 || Cudd_Regular(bG)->ref != 1 )
    {
        // inlining of macro Cudd_Regular
        var fanout ptrint  = (ptrint)((DdNode *)((ptruint)(bF) & ~01)).ref * ((DdNode *)((ptruint)(bG) & ~01)).ref;
        cuddSatDec(fanout);
        cuddHashTableInsert2( table, bF, bG, bRes, fanout );
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(bRes).ref);
    return bRes;
} 

/**Function*************************************************************

  Synopsis    [Testbench.]

  Description [This procedure takes BDD manager ddF and two BDDs
  in this manager (bF and bG).  It creates a new manager ddG,
  transfers bG into it and then reorders it, resulting in bG2.
  Then it tries to compute the product of bF and bG2 in ddF.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_TestAndPerm(ddF *DdManager, bF *DdNode, bG *DdNode) {
    var ddG *DdManager;
    var bG2, bRes1, bRes2 *DdNode;
    var clk abctime;
    // disable variable ordering in ddF
    Cudd_AutodynDisable( ddF );

    // create new BDD manager
    ddG = Cudd_Init( ddF.size, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    // transfer BDD into it
    Cudd_ShuffleHeap( ddG, ddF.invperm );
    bG2 = Extra_TransferLevelByLevel( ddF, ddG, bG );   Cudd_Ref( bG2 );
    // reorder the new manager
    Cudd_ReduceHeap( ddG, CUDD_REORDER_SYMM_SIFT, 1 );

    // compute the result
clk = Abc_Clock();
    bRes1 = Cudd_bddAnd( ddF, bF, bG );                 Cudd_Ref( bRes1 );
Abc_PrintTime( 1, "Runtime of Cudd_bddAnd  ", Abc_Clock() - clk );

    // compute the result
Counter = 0;
clk = Abc_Clock();
    bRes2 = Extra_bddAndPermute( ddF, bF, ddG, bG2, nil );      Cudd_Ref( bRes2 );
Abc_PrintTime( 1, "Runtime of new procedure", Abc_Clock() - clk );
fmt.Printf("Recursive calls = %d\n", Counter);
fmt.Printf("|F| =%6d  |G| =%6d  |H| =%6d  |F|*|G| =%9d  ", Cudd_DagSize(bF), Cudd_DagSize(bG), Cudd_DagSize(bRes2), Cudd_DagSize(bF) * Cudd_DagSize(bG));

    if bRes1 == bRes2 {
        fmt.Printf("Result verified.\n\n");
    } else {
        fmt.Printf("Result is incorrect.\n\n"); }

    Cudd_RecursiveDeref( ddF, bRes1 );
    Cudd_RecursiveDeref( ddF, bRes2 );
    // quit the new manager
    Cudd_RecursiveDeref( ddG, bG2 );
    Extra_StopManager( ddG );

    // re-enable variable ordering in ddF
    Cudd_AutodynEnable( ddF, CUDD_REORDER_SYMM_SIFT );
}

/**Function*************************************************************

  Synopsis    [Writes ZDD into a PLA file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_zddDumpPla(dd *DdManager, F *DdNode, nVars int32, pFileName string) {
    var gen *DdGen;
    var pCube string;
    var pPath *int32;
    var i int32;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "# PLA file dumped by Extra_zddDumpPla() in ABC\n");
    fmt.Fprintf(pFile, ".i %d\n", nVars);
    fmt.Fprintf(pFile, ".o 1\n");
    pCube = ABC_CALLOC( char, dd.sizeZ );
    // inlining of macro Cudd_zddForeachPath
    for (gen) = Cudd_zddFirstPath(dd, F, &pPath);        Cudd_IsGenEmpty(gen) ? Cudd_GenFree(gen) : TRUE;        (void) Cudd_zddNextPath(gen, &pPath) {
        for  i = 0; i < nVars; i++  {
            pCube[i] = '-';
        }
        for  i = 0; i < nVars; i++  {
            if pPath[2*i] == 1 || pPath[2*i+1] == 1 {
                pCube[i] = '0' + (pPath[2*i] == 1); }
        }
        fmt.Fprintf(pFile, "%s 1\n", pCube);           
    }
    fmt.Fprintf(pFile, ".e\n\n");
    _ = (pFile).Close();
    ABC_FREE( pCube );
}

/**Function*************************************************************

  Synopsis    [Constructing ZDD of a graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_GraphExperiment() {
    var Edges [5][5]int32 = [5][5]int32{ 
        {1, 3, 4}, 
        {1, 5}, 
        {2, 3, 5}, 
        {2, 4}, 
    };
    var e, n int32;

    var dd *DdManager  = Cudd_Init( 0, 6, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );

    // create the edges
    var zGraph, zEdge, zVar, zTemp *DdNode;
    // inlining of macro DD_ZERO
    zGraph = ((dd).zero);                                                      Cudd_Ref( zGraph );
    for  e = 0; Edges[e][0]; e++  {
        // inlining of macro DD_ONE
        zEdge = ((dd).one);                                                    Cudd_Ref( zEdge );
        for  n = 0; Edges[e][n]; n++  {
            // inlining of macro DD_ONE
            // inlining of macro DD_ZERO
            zVar = cuddZddGetNode( dd, Edges[e][n], ((dd).one), ((dd).zero) ); Cudd_Ref( zVar );
            zEdge = Cudd_zddUnateProduct( dd, zTemp = zEdge, zVar );           Cudd_Ref( zEdge );
            Cudd_RecursiveDerefZdd( dd, zTemp );
            Cudd_RecursiveDerefZdd( dd, zVar );
        }
        zGraph = Cudd_zddUnion( dd, zTemp = zGraph, zEdge );                   Cudd_Ref( zGraph );
        Cudd_RecursiveDerefZdd( dd, zTemp );
        Cudd_RecursiveDerefZdd( dd, zEdge );
    }

    Cudd_zddPrintMinterm( dd, zGraph );

    Cudd_RecursiveDerefZdd( dd, zGraph );
    Cudd_Quit(dd);
}




/**Function********************************************************************

  Synopsis    [Performs the reordering-sensitive step of Extra_zddCombination().]

  Description [Generates in a bottom-up fashion ZDD for one combination 
               whose var values are given in the array VarValues. If necessary,
               creates new variables on the fly.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func extraZddCombination(dd *DdManager, VarValues *int32, nVars int32) *DdNode {
    var lev, index int32;
    var zRes, zTemp *DdNode;

    /* transform the combination from the array VarValues into a ZDD cube. */
    zRes = dd.one;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(zRes).ref);

    /*  go through levels starting bottom-up and create nodes 
     *  if these variables are present in the comb
     */
    for lev = nVars - 1; lev >= 0; lev-- { 
        if lev >= dd.sizeZ {
            index = lev;
        } else {
            index = dd.invpermZ[lev];
        }
        if VarValues[index] == 1 {
            /* compose zRes with ZERO for the given ZDD variable */
            zRes = cuddZddGetNode( dd, index, zTemp = zRes, dd.zero );
            if zRes == nil {
                Cudd_RecursiveDerefZdd( dd, zTemp );
                return nil;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(zRes).ref);
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(zTemp).ref);
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(zRes).ref);
    return zRes;

} /* end of extraZddCombination */

/**Function********************************************************************

  Synopsis    [Creates ZDD of the combination containing given variables.]

  Description [Creates ZDD of the combination containing given variables.
               VarValues contains 1 for a variable that belongs to the 
               combination and 0 for a varible that does not belong. 
               nVars is number of ZDD variables in the array.]

  SideEffects [New ZDD variables are created if indices of the variables
               present in the combination are larger than the currently
               allocated number of ZDD variables.]

  SeeAlso     []

******************************************************************************/
func Extra_zddCombination(dd *DdManager, VarValues *int32, nVars int32) *DdNode {
    var res *DdNode;
    for {
    dd.reordered = 0;
    res = extraZddCombination(dd, VarValues, nVars);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Extra_zddCombination */

/**Function********************************************************************

  Synopsis    [Generates a random set of combinations.]

  Description [Given a set of n elements, each of which is encoded using one
               ZDD variable, this function generates a random set of k subsets 
               (combinations of elements) with density d. Assumes that k and n
               are positive integers. Returns NULL if density is less than 0.0 
               or more than 1.0.]

  SideEffects [Allocates new ZDD variables if their current number is less than n.]

  SeeAlso     []

******************************************************************************/
DdNode* Extra_zddRandomSet( 
  DdManager * dd,   /* the DD manager */
  int n,            /* the number of elements */
  int k,            /* the number of combinations (subsets) */
  double d)         /* average density of elements in combinations */
{
    var Result, TempComb, Aux *DdNode;
    var c int32;
    var v int32;
    var Limit int32;
    var VarValues *int32;

    /* sanity check the parameters */
    if n <= 0 || k <= 0 || d < 0.0 || d > 1.0 {
        return nil; }

    /* allocate temporary storage for variable values */
    VarValues = ABC_ALLOC( int, n );
    if VarValues == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return nil;
    }

    /* start the new set */
    Result = dd.zero;
    Cudd_Ref( Result );

    /* seed random number generator */
    Cudd_Srandom( time.Now().Unix() );
//  Cudd_Srandom( 4 );
    /* determine the limit below which var belongs to the combination */
    Limit = int32((d * 2147483561.0));

    /* add combinations one by one */
    for  c = 0; c < k; c++  {
        for  v = 0; v < n; v++  {
            if Cudd_Random() <= Limit {
                VarValues[v] = 1;
            } else {
                VarValues[v] = 0; }
        }

        TempComb = Extra_zddCombination( dd, VarValues, n );
        Cudd_Ref( TempComb );

        /* make sure that this combination is not already in the set */
        if c != 0 { /* at least one combination is already included */

            Aux = Cudd_zddDiff( dd, Result, TempComb );
            Cudd_Ref( Aux );
            if Aux != Result {
                Cudd_RecursiveDerefZdd( dd, Aux );
                Cudd_RecursiveDerefZdd( dd, TempComb );
                c--;
                continue;
            
            } else { /* Aux is the same node as Result */
                Cudd_Deref( Aux );
            }
        }

        Result = Cudd_zddUnion( dd, Aux = Result, TempComb );
        Cudd_Ref( Result );
        Cudd_RecursiveDerefZdd( dd, Aux );
        Cudd_RecursiveDerefZdd( dd, TempComb );
    }

    ABC_FREE( VarValues );
    Cudd_Deref( Result );
    return Result;

} /* end of Extra_zddRandomSet */

/**Function********************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Extra_ZddTest() {
    var N int32   = 64;
    var K0 int32  = 1000;
    var i, Size int32;
    var dd *DdManager  = Cudd_Init( 0, 32, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    for  i = 1; i <= 10; i++  {
        var K int32  = K0 * i;
        var zRandSet *DdNode  = Extra_zddRandomSet( dd, N, K, 0.5 );  Cudd_Ref(zRandSet);
        Size = Cudd_zddDagSize(zRandSet);
        //Cudd_zddPrintMinterm( dd, zRandSet );
        fmt.Printf("N = %5d  K = %5d  BddSize = %6d   MemBdd = %8.3f MB   MemBit = %8.3f MB   Ratio = %8.3f %%\n", N, K, Size, 20.0*Size/(1<<20), 0.125 * N * K /(1<<20), 100.0*(0.125 * N * K)/(20.0*Size));
        Cudd_RecursiveDerefZdd( dd, zRandSet );
    }
    Cudd_Quit(dd);
}


/**Function********************************************************************

  Synopsis    [Performs the reordering-sensitive step of Extra_bddTuples().]

  Description [Generates in a bottom-up fashion BDD for all combinations
               composed of k variables out of variables belonging to bVarsN.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * extraBddTuples( 
  DdManager * dd,    /* the DD manager */
  DdNode * bVarsK,   /* the number of variables in tuples */
  DdNode * bVarsN)   /* the set of all variables */
{
    var bRes, bRes0, bRes1 *DdNode;
    statLine(dd); 

    /* terminal cases */
/*  if ( k < 0 || k > n )
 *      return dd->zero;
 *  if ( n == 0 )
 *      return dd->one; 
 */
    // inlining of macro cuddI
    if  (((bVarsK.index)==CUDD_CONST_INDEX)?int32((bVarsK.index)):(dd).perm[(bVarsK.index)]) < (((bVarsN.index)==CUDD_CONST_INDEX)?int32((bVarsN.index)):(dd).perm[(bVarsN.index)]) {
        return b0; }
    if bVarsN == b1 {
        return b1; }

    /* check cache */
    bRes = cuddCacheLookup2(dd, extraBddTuples, bVarsK, bVarsN);
    if bRes != nil {
        return(bRes); }

    /* ZDD in which is variable is 0 */
/*  bRes0 = extraBddTuples( dd, k,     n-1 ); */
    // inlining of macro cuddT
    bRes0 = extraBddTuples( dd, bVarsK, ((bVarsN).type.kids.T) );
    if bRes0 == nil { 
        return nil; }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(bRes0).ref);

    /* ZDD in which is variable is 1 */
/*  bRes1 = extraBddTuples( dd, k-1,          n-1 ); */
    if bVarsK == b1 {
        bRes1 = b0;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes1).ref);
    
    } else {
        // inlining of macro cuddT
        bRes1 = extraBddTuples( dd, ((bVarsK).type.kids.T), ((bVarsN).type.kids.T) );
        if bRes1 == nil {
            Cudd_RecursiveDeref( dd, bRes0 );
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes1).ref);
    }

    /* consider the case when Res0 and Res1 are the same node */
    if bRes0 == bRes1 {
        bRes = bRes1;
    /* consider the case when Res1 is complemented */
    // inlining of macro Cudd_IsComplement
    } else if  int32( ((ptrint) (bRes1) & 01)) != 0 {
        // inlining of macro Cudd_Not
        bRes = cuddUniqueInter(dd, bVarsN.index, ((DdNode *)((ptrint)(bRes1) ^ 01)), ((DdNode *)((ptrint)(bRes0) ^ 01)));
        if bRes == nil {
            Cudd_RecursiveDeref(dd,bRes0);
            Cudd_RecursiveDeref(dd,bRes1);
            return nil;
        }
        // inlining of macro Cudd_Not
        bRes = ((DdNode *)((ptrint)(bRes) ^ 01));
     
    } else {
        bRes = cuddUniqueInter( dd, bVarsN.index, bRes1, bRes0 );
        if bRes == nil {
            Cudd_RecursiveDeref(dd,bRes0);
            Cudd_RecursiveDeref(dd,bRes1);
            return nil;
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(bRes0).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(bRes1).ref);

    /* insert the result into cache */
    cuddCacheInsert2(dd, extraBddTuples, bVarsK, bVarsN, bRes);
    return bRes;

} /* end of extraBddTuples */

/**Function********************************************************************

  Synopsis    [Builds BDD representing the set of fixed-size variable tuples.]

  Description [Creates BDD of all combinations of variables in Support that have k vars in them.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
DdNode * Extra_bddTuples( 
  DdManager * dd,   /* the DD manager */
  int K,            /* the number of variables in tuples */
  DdNode * VarsN)   /* the set of all variables represented as a BDD */
{
    var res *DdNode;
    var autoDyn int32;

    /* it is important that reordering does not happen, 
       otherwise, this methods will not work */

    autoDyn = dd.autoDyn;
    dd.autoDyn = 0;

    for {
        /* transform the numeric arguments (K) into a DdNode * argument;
         * this allows us to use the standard internal CUDD cache */
        var VarSet *DdNode = VarsN;
        var VarsK *DdNode = VarsN;
        var nVars, i int32;

        /* determine the number of variables in VarSet */
        for  VarSet != b1  {
            nVars++;
            /* make sure that the VarSet is a cube */
            // inlining of macro cuddE
            if  ((VarSet).type.kids.E) != b0 {
                return nil; }
            // inlining of macro cuddT
            VarSet = ((VarSet).type.kids.T);
        }
        /* make sure that the number of variables in VarSet is less or equal 
           that the number of variables that should be present in the tuples
        */
        if K > nVars {
            return nil; }

        /* the second argument in the recursive call stands for <n>;
         * create the first argument, which stands for <k> 
         * as when we are talking about the tuple of <k> out of <n> */
        for  i = 0; i < nVars-K; i++  {
            // inlining of macro cuddT
            VarsK = ((VarsK).type.kids.T);
        }

        dd.reordered = 0;
        res = extraBddTuples(dd, VarsK, VarsN );

    
if !(dd.reordered == 1) { break }
} 
    dd.autoDyn = autoDyn;
    return(res);

} /* end of Extra_bddTuples */


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

