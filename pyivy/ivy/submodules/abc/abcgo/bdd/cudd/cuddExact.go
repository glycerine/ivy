package cudd

/**CFile***********************************************************************

  FileName    [cuddExact.c]

  PackageName [cudd]

  Synopsis    [Functions for exact variable reordering.]

  Description [External procedures included in this file:
                <ul>
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddExact()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> getMaxBinomial()
                <li> gcd()
                <li> getMatrix()
                <li> freeMatrix()
                <li> getLevelKeys()
                <li> ddShuffle()
                <li> ddSiftUp()
                <li> updateUB()
                <li> ddCountRoots()
                <li> ddClearGlobal()
                <li> computeLB()
                <li> updateEntry()
                <li> pushDown()
                <li> initSymmInfo()
                </ul>]

  Author      [Cheng Hua, Fabio Somenzi]

  Copyright   [Copyright (c) 1995-2004, Regents of the University of Colorado

  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:

  Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright
  notice, this list of conditions and the following disclaimer in the
  documentation and/or other materials provided with the distribution.

  Neither the name of the University of Colorado nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
  FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
  BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
  LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
  ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.]

******************************************************************************/

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

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

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddExact.c,v 1.28 2009/02/19 16:19:19 fabio Exp $";
#endif

#ifdef DD_STATS
static int ddTotalShuffles;
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int getMaxBinomial (int n);

// static DdHalfWord ** getMatrix (int rows, int cols);

// static void freeMatrix (DdHalfWord **matrix);

// static int getLevelKeys (DdManager *table, int l);

// static int ddShuffle (DdManager *table, DdHalfWord *permutation, int lower, int upper);

// static int ddSiftUp (DdManager *table, int x, int xLow);

// static int updateUB (DdManager *table, int oldBound, DdHalfWord *bestOrder, int lower, int upper);

// static int ddCountRoots (DdManager *table, int lower, int upper);

// static void ddClearGlobal (DdManager *table, int lower, int maxlevel);

// static int computeLB (DdManager *table, DdHalfWord *order, int roots, int cost, int lower, int upper, int level);

// static int updateEntry (DdManager *table, DdHalfWord *order, int level, int cost, DdHalfWord **orders, int *costs, int subsets, char *mask, int lower, int upper);

// static void pushDown (DdHalfWord *order, int j, int level);

// static DdHalfWord * initSymmInfo (DdManager *table, int lower, int upper);

// static int checkSymmInfo (DdManager *table, DdHalfWord *symmInfo, int index, int level);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Exact variable ordering algorithm.]

  Description [Exact variable ordering algorithm. Finds an optimum
  order for the variables between lower and upper.  Returns 1 if
  successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddExact(table *DdManager, lower int32, upper int32) int32 {
    var k, i, j int32;
    var maxBinomial, oldSubsets, newSubsets int32;
    var subsetCost int32;
    var size int32;                   /* number of variables to be reordered */
    var unused, nvars, level, result int32;
    var upperBound, lowerBound, cost int32;
    var roots int32;
    var mask string  = nil;
    var symmInfo *DdHalfWord  = nil;
    var newOrder **DdHalfWord  = nil;
    var oldOrder **DdHalfWord  = nil;
    var newCost *int32  = nil;
    var oldCost *int32  = nil;
    var tmpOrder **DdHalfWord;
    var tmpCost *int32;
    var bestOrder *DdHalfWord  = nil;
    var order *DdHalfWord;
#ifdef DD_STATS
    int  ddTotalSubsets;
#endif

    /* Restrict the range to be reordered by excluding unused variables
    ** at the two ends. */
    for table.subtables[lower].keys == 1 &&
           table.vars[table.invperm[lower]].ref == 1 &&
           lower < upper {
        lower++;
    }
    for table.subtables[upper].keys == 1 &&
           table.vars[table.invperm[upper]].ref == 1 &&
           lower < upper {
        upper--;
    }
    if lower == upper { return(1); } /* trivial problem */

    /* Apply symmetric sifting to get a good upper bound and to extract
    ** symmetry information. */
    result = cuddSymmSiftingConv(table,lower,upper);
    if result == 0 { goto cuddExactOutOfMem; }

#ifdef DD_STATS
    (void) fprintf(table->out,"\n");
    ddTotalShuffles = 0;
    ddTotalSubsets = 0;
#endif

    /* Initialization. */
    nvars = table.size;
    size = upper - lower + 1;
    /* Count unused variable among those to be reordered.  This is only
    ** used to compute maxBinomial. */
    unused = 0;
    for i = lower + 1; i < upper; i++ {
        if table.subtables[i].keys == 1 &&
            table.vars[table.invperm[i]].ref == 1 {
            unused++; }
    }

    /* Find the maximum number of subsets we may have to store. */
    maxBinomial = getMaxBinomial(size - unused);
    if maxBinomial == -1 { goto cuddExactOutOfMem; }

    newOrder = getMatrix(maxBinomial, size);
    if newOrder == nil { goto cuddExactOutOfMem; }

    newCost = ABC_ALLOC(int, maxBinomial);
    if newCost == nil { goto cuddExactOutOfMem; }

    oldOrder = getMatrix(maxBinomial, size);
    if oldOrder == nil { goto cuddExactOutOfMem; }

    oldCost = ABC_ALLOC(int, maxBinomial);
    if oldCost == nil { goto cuddExactOutOfMem; }

    bestOrder = ABC_ALLOC(DdHalfWord, size);
    if bestOrder == nil { goto cuddExactOutOfMem; }

    mask = ABC_ALLOC(char, nvars);
    if mask == nil { goto cuddExactOutOfMem; }

    symmInfo = initSymmInfo(table, lower, upper);
    if symmInfo == nil { goto cuddExactOutOfMem; }

    roots = ddCountRoots(table, lower, upper);

    /* Initialize the old order matrix for the empty subset and the best
    ** order to the current order. The cost for the empty subset includes
    ** the cost of the levels between upper and the constants. These levels
    ** are not going to change. Hence, we count them only once.
    */
    oldSubsets = 1;
    for i = 0; i < size; i++ {
        oldOrder[0][i] = bestOrder[i] = (DdHalfWord) table.invperm[i+lower];
    }
    subsetCost = table.constants.keys;
    for i = upper + 1; i < nvars; i++ {
        subsetCost += getLevelKeys(table,i);
    }
    oldCost[0] = subsetCost;
    /* The upper bound is initialized to the current size of the BDDs. */
    upperBound = table.keys - table.isolated;

    /* Now consider subsets of increasing size. */
    for k = 1; k <= size; k++ {
#ifdef DD_STATS
        (void) fprintf(table->out,"Processing subsets of size %d\n", k);
        fflush(table->out);
#endif
        newSubsets = 0;
        level = size - k;               /* offset of first bottom variable */

        for i = 0; i < oldSubsets; i++ { /* for each subset of size k-1 */
            order = oldOrder[i];
            cost = oldCost[i];
            lowerBound = computeLB(table, order, roots, cost, lower, upper,
                                   level);
            if lowerBound >= upperBound {
                continue; }
            /* Impose new order. */
            result = ddShuffle(table, order, lower, upper);
            if result == 0 { goto cuddExactOutOfMem; }
            upperBound = updateUB(table,upperBound,bestOrder,lower,upper);
            /* For each top bottom variable. */
            for j = level; j >= 0; j-- {
                /* Skip unused variables. */
                if table.subtables[j+lower-1].keys == 1 &&
                    table.vars[table.invperm[j+lower-1]].ref == 1 { continue; }
                /* Find cost under this order. */
                subsetCost = cost + getLevelKeys(table, lower + level);
                newSubsets = updateEntry(table, order, level, subsetCost,
                                         newOrder, newCost, newSubsets, mask,
                                         lower, upper);
                if j == 0 {
                    break; }
                if checkSymmInfo(table, symmInfo, order[j-1], level) == 0 {
                    continue; }
                pushDown(order,j-1,level);
                /* Impose new order. */
                result = ddShuffle(table, order, lower, upper);
                if result == 0 { goto cuddExactOutOfMem; }
                upperBound = updateUB(table,upperBound,bestOrder,lower,upper);
            } /* for each bottom variable */
        } /* for each subset of size k */

        /* New orders become old orders in preparation for next iteration. */
        tmpOrder = oldOrder; tmpCost = oldCost;
        oldOrder = newOrder; oldCost = newCost;
        newOrder = tmpOrder; newCost = tmpCost;
#ifdef DD_STATS
        ddTotalSubsets += newSubsets;
#endif
        oldSubsets = newSubsets;
    }
    result = ddShuffle(table, bestOrder, lower, upper);
    if result == 0 { goto cuddExactOutOfMem; }
#ifdef DD_STATS
#ifdef DD_VERBOSE
    (void) fprintf(table->out,"\n");
#endif
    (void) fprintf(table->out,"#:S_EXACT   %8d: total subsets\n",
                   ddTotalSubsets);
    (void) fprintf(table->out,"#:H_EXACT   %8d: total shuffles",
                   ddTotalShuffles);
#endif

    freeMatrix(newOrder);
    freeMatrix(oldOrder);
    ABC_FREE(bestOrder);
    ABC_FREE(oldCost);
    ABC_FREE(newCost);
    ABC_FREE(symmInfo);
    ABC_FREE(mask);
    return(1);

cuddExactOutOfMem:

    if newOrder != nil { freeMatrix(newOrder); }
    if oldOrder != nil { freeMatrix(oldOrder); }
    if bestOrder != nil { ABC_FREE(bestOrder); }
    if oldCost != nil { ABC_FREE(oldCost); }
    if newCost != nil { ABC_FREE(newCost); }
    if symmInfo != nil { ABC_FREE(symmInfo); }
    if mask != nil { ABC_FREE(mask); }
    table.errorCode = CUDD_MEMORY_OUT;
    return(0);

} /* end of cuddExact */


/**Function********************************************************************

  Synopsis    [Returns the maximum value of (n choose k) for a given n.]

  Description [Computes the maximum value of (n choose k) for a given
  n.  The maximum value occurs for k = n/2 when n is even, or k =
  (n-1)/2 when n is odd.  The algorithm used in this procedure avoids
  intermediate overflow problems.  It is based on the identity
  <pre>
    binomial(n,k) = n/k * binomial(n-1,k-1).
  </pre>
  Returns the computed value if successful; -1 if out of range.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func getMaxBinomial(n int32) int32 {
    var i, j, result float64;

    if n < 0 || n > 33 { return(-1); } /* error */
    if n < 2 { return(1); }

    for result = float64(((n+3)/2)), i = result+1, j=2; i <= n; i, j = i+1, j+1 {
        result *= i;
        result /= j;
    }

    return(int32(result));

} /* end of getMaxBinomial */


     
                                                                               

                                                

                                                                   
                                                        

                    

                

                                                                               
          
    
         
         
 
          
          
                

                       
                          
                          

                              

                                                                          
                                                               
      
                    
                          
                                              
                            
                                     
                        
                                     
                 
                                                   
                        
             
                
                                                   
                        
                                               
                              
             
         
     

              

                  
      


/**Function********************************************************************

  Synopsis    [Allocates a two-dimensional matrix of ints.]

  Description [Allocates a two-dimensional matrix of ints.
  Returns the pointer to the matrix if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [freeMatrix]

******************************************************************************/
static DdHalfWord **
getMatrix(
  int  rows /* number of rows */,
  int  cols /* number of columns */)
{
    var matrix **DdHalfWord;
    var i int32;

    if cols*rows == 0 { return(nil); }
    matrix = ABC_ALLOC(DdHalfWord *, rows);
    if matrix == nil { return(nil); }
    matrix[0] = ABC_ALLOC(DdHalfWord, cols*rows);
    if matrix[0] == nil {
        ABC_FREE(matrix);
        return(nil);
    }
    for i = 1; i < rows; i++ {
        matrix[i] = matrix[i-1] + cols;
    }
    return(matrix);

} /* end of getMatrix */


/**Function********************************************************************

  Synopsis    [Frees a two-dimensional matrix allocated by getMatrix.]

  Description []

  SideEffects [None]

  SeeAlso     [getMatrix]

******************************************************************************/
func freeMatrix(matrix **DdHalfWord) {
    ABC_FREE(matrix[0]);
    ABC_FREE(matrix);
    return;

} /* end of freeMatrix */


/**Function********************************************************************

  Synopsis    [Returns the number of nodes at one level of a unique table.]

  Description [Returns the number of nodes at one level of a unique table.
  The projection function, if isolated, is not counted.]

  SideEffects [None]

  SeeAlso []

******************************************************************************/
func getLevelKeys(table *DdManager, l int32) int32 {
    var isolated int32;
    var x int32;        /* x is an index */

    x = table.invperm[l];
    isolated = table.vars[x].ref == 1;

    return(table.subtables[l].keys - isolated);

} /* end of getLevelKeys */


/**Function********************************************************************

  Synopsis    [Reorders variables according to a given permutation.]

  Description [Reorders variables according to a given permutation.
  The i-th permutation array contains the index of the variable that
  should be brought to the i-th level. ddShuffle assumes that no
  dead nodes are present and that the interaction matrix is properly
  initialized.  The reordering is achieved by a series of upward sifts.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso []

******************************************************************************/
func ddShuffle(table *DdManager, permutation *DdHalfWord, lower int32, upper int32) int32 {
    var index DdHalfWord;
    var level int32;
    var position int32;
     
                        
      
    var result int32;
#ifdef DD_STATS
    long        localTime;
    int         initialSize;
#ifdef DD_VERBOSE
    int         finalSize;
#endif
    int         previousSize;
#endif

#ifdef DD_STATS
    localTime = util_cpu_time();
    initialSize = table->keys - table->isolated;
#endif

     
                          

                                                      
                                               
                                                                
     
                                    
      

    for level = 0; level <= upper - lower; level++ {
        index = permutation[level];
        position = table.perm[index];
#ifdef DD_STATS
        previousSize = table->keys - table->isolated;
#endif
        result = ddSiftUp(table,position,level+lower);
        if !(result != 0) { return(0); }
    }

#ifdef DD_STATS
    ddTotalShuffles++;
#ifdef DD_VERBOSE
    finalSize = table->keys - table->isolated;
    if (finalSize < initialSize) {
        (void) fprintf(table->out,"-");
    } else if (finalSize > initialSize) {
        (void) fprintf(table->out,"+");
    } else {
        (void) fprintf(table->out,"=");
    }
    if ((ddTotalShuffles & 63) == 0) (void) fprintf(table->out,"\n");
    fflush(table->out);
#endif
#endif

    return(1);

} /* end of ddShuffle */


/**Function********************************************************************

  Synopsis    [Moves one variable up.]

  Description [Takes a variable from position x and sifts it up to
  position xLow;  xLow should be less than or equal to x.
  Returns 1 if successful; 0 otherwise]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func ddSiftUp(table *DdManager, x int32, xLow int32) int32 {
    var y int32;
    var size int32;

    y = cuddNextLow(table,x);
    for y >= xLow {
        size = cuddSwapInPlace(table,y,x);
        if size == 0 {
            return(0);
        }
        x = y;
        y = cuddNextLow(table,x);
    }
    return(1);

} /* end of ddSiftUp */


/**Function********************************************************************

  Synopsis    [Updates the upper bound and saves the best order seen so far.]

  Description [Updates the upper bound and saves the best order seen so far.
  Returns the current value of the upper bound.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func updateUB(table *DdManager, oldBound int32, bestOrder *DdHalfWord, lower int32, upper int32) int32 {
    var i int32;
    var newBound int32  = table.keys - table.isolated;

    if newBound < oldBound {
#ifdef DD_STATS
        (void) fprintf(table->out,"New upper bound = %d\n", newBound);
        fflush(table->out);
#endif
        for i = lower; i <= upper; i++ {
            bestOrder[i-lower] = (DdHalfWord) table.invperm[i];
        }
        return(newBound);
    } else {
        return(oldBound);
    }

} /* end of updateUB */


/**Function********************************************************************

  Synopsis    [Counts the number of roots.]

  Description [Counts the number of roots at the levels between lower and
  upper.  The computation is based on breadth-first search.
  A node is a root if it is not reachable from any previously visited node.
  (All the nodes at level lower are therefore considered roots.)
  The visited flag uses the LSB of the next pointer.  Returns the root
  count. The roots that are constant nodes are always ignored.]

  SideEffects [None]

  SeeAlso     [ddClearGlobal]

******************************************************************************/
func ddCountRoots(table *DdManager, lower int32, upper int32) int32 {
    var i, j int32;
    var f *DdNode;
    var nodelist *DdNodePtr;
    var sentinel *DdNode  = &(table.sentinel);
    var slots int32;
    var roots int32;
    var maxlevel int32  = lower;

    for i = lower; i <= upper; i++ {
        nodelist = table.subtables[i].nodelist;
        slots = table.subtables[i].slots;
        for j = 0; j < slots; j++ {
            f = nodelist[j];
            for f != sentinel {
                /* A node is a root of the DAG if it cannot be
                ** reached by nodes above it. If a node was never
                ** reached during the previous depth-first searches,
                ** then it is a root, and we start a new depth-first
                ** search from it.
                */
                // inlining of macro Cudd_IsComplement
                if !((int32( ((ptrint) (f.next) & 01)) != 0)) {
                    if f != table.vars[f.index] {
                        roots++;
                    }
                }
                // inlining of macro Cudd_IsConstant
                // inlining of macro Cudd_Regular
                // inlining of macro cuddT
                if !((((DdNode *)((ptruint)(((f).type.kids.T)) & ~01))).index == CUDD_CONST_INDEX) {
                    // inlining of macro cuddT
                    // inlining of macro Cudd_Complement
                    ((f).type.kids.T).next = ((DdNode *)((ptruint)(((f).type.kids.T).next) | 01));
                    // inlining of macro cuddT
                    if table.perm[((f).type.kids.T).index] > maxlevel {
                        // inlining of macro cuddT
                        maxlevel = table.perm[((f).type.kids.T).index]; }
                }
                // inlining of macro Cudd_IsConstant
                // inlining of macro Cudd_Regular
                // inlining of macro cuddE
                if !((((DdNode *)((ptruint)(((f).type.kids.E)) & ~01))).index == CUDD_CONST_INDEX) {
                    // inlining of macro Cudd_Regular
                    // inlining of macro cuddE
                    ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).next =
                        // inlining of macro Cudd_Complement
                        // inlining of macro Cudd_Regular
                        // inlining of macro cuddE
                        ((DdNode *)((ptruint)(((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).next) | 01));
                    // inlining of macro Cudd_Regular
                    // inlining of macro cuddE
                    if table.perm[((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).index] > maxlevel {
                        // inlining of macro Cudd_Regular
                        // inlining of macro cuddE
                        maxlevel = table.perm[((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).index]; }
                }
                // inlining of macro Cudd_Regular
                f = ((DdNode *)((ptruint)(f.next) & ~01));
            }
        }
    }
    ddClearGlobal(table, lower, maxlevel);

    return(roots);

} /* end of ddCountRoots */


/**Function********************************************************************

  Synopsis    [Scans the DD and clears the LSB of the next pointers.]

  Description [Scans the DD and clears the LSB of the next pointers.
  The LSB of the next pointers are used as markers to tell whether a
  node was reached. Once the roots are counted, these flags are
  reset.]

  SideEffects [None]

  SeeAlso     [ddCountRoots]

******************************************************************************/
func ddClearGlobal(table *DdManager, lower int32, maxlevel int32) {
    var i, j int32;
    var f *DdNode;
    var nodelist *DdNodePtr;
    var sentinel *DdNode  = &(table.sentinel);
    var slots int32;

    for i = lower; i <= maxlevel; i++ {
        nodelist = table.subtables[i].nodelist;
        slots = table.subtables[i].slots;
        for j = 0; j < slots; j++ {
            f = nodelist[j];
            for f != sentinel {
                // inlining of macro Cudd_Regular
                f.next = ((DdNode *)((ptruint)(f.next) & ~01));
                f = f.next;
            }
        }
    }

} /* end of ddClearGlobal */


/**Function********************************************************************

  Synopsis    [Computes a lower bound on the size of a BDD.]

  Description [Computes a lower bound on the size of a BDD from the
  following factors:
  <ul>
  <li> size of the lower part of it;
  <li> size of the part of the upper part not subjected to reordering;
  <li> number of roots in the part of the BDD subjected to reordering;
  <li> variable in the support of the roots in the upper part of the
       BDD subjected to reordering.
  <ul/>]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
static int
computeLB(
  DdManager * table             /* manager */,
  DdHalfWord * order            /* optimal order for the subset */,
  int  roots                    /* roots between lower and upper */,
  int  cost                     /* minimum cost for the subset */,
  int  lower                    /* lower level to be reordered */,
  int  upper                    /* upper level to be reordered */,
  int  level                    /* offset for the current top bottom var */
  )
{
    var i int32;
    var lb int32  = cost;
    var lb1 int32;
    var lb2 int32;
    var support int32;
    var ref DdHalfWord;

    /* The levels not involved in reordering are not going to change.
    ** Add their sizes to the lower bound.
    */
    for i = 0; i < lower; i++ {
        lb += getLevelKeys(table,i);
    }
    /* If a variable is in the support, then there is going
    ** to be at least one node labeled by that variable.
    */
    for i = lower; i <= lower+level; i++ {
        support = table.subtables[i].keys > 1 ||
            table.vars[order[i-lower]].ref > 1;
        lb1 += support;
    }

    /* Estimate the number of nodes required to connect the roots to
    ** the nodes in the bottom part. */
    if lower+level+1 < table.size {
        if lower+level < upper {
            ref = table.vars[order[level+1]].ref;
        } else {
            ref = table.vars[table.invperm[upper+1]].ref; }
        lb2 = table.subtables[lower+level+1].keys -
            (ref > (DdHalfWord) 1) - roots;
    } else {
        lb2 = 0;
    }

    lb += lb1 > lb2 ? lb1 : lb2;

    return(lb);

} /* end of computeLB */


/**Function********************************************************************

  Synopsis    [Updates entry for a subset.]

  Description [Updates entry for a subset. Finds the subset, if it exists.
  If the new order for the subset has lower cost, or if the subset did not
  exist, it stores the new order and cost. Returns the number of subsets
  currently in the table.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func updateEntry(table *DdManager, order *DdHalfWord, level int32, cost int32, orders **DdHalfWord, costs *int32, subsets int32, mask string, lower int32, upper int32) int32 {
    var i, j int32;
    var size int32  = upper - lower + 1;

    /* Build a mask that says what variables are in this subset. */
    for i = lower; i <= upper; i++ {
        mask[table.invperm[i]] = 0;
    }
    for i = level; i < size; i++ {
        mask[order[i]] = 1;
    }

    /* Check each subset until a match is found or all subsets are examined. */
    for i = 0; i < subsets; i++ {
        var subset *DdHalfWord  = orders[i];
        for j = level; j < size; j++ {
            if mask[subset[j]] == 0 {
                break; }
        }
        if j == size {          /* no mismatches: success */
            break; }
    }
    if i == subsets || cost < costs[i] {              /* add or replace */
        for j = 0; j < size; j++ {
            orders[i][j] = order[j];
        }
        costs[i] = cost;
        subsets += (i == subsets);
    }
    return(subsets);

} /* end of updateEntry */


/**Function********************************************************************

  Synopsis    [Pushes a variable in the order down to position "level."]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func pushDown(order *DdHalfWord, j int32, level int32) {
    var i int32;
    var tmp DdHalfWord;

    tmp = order[j];
    for i = j; i < level; i++ {
        order[i] = order[i+1];
    }
    order[level] = tmp;
    return;

} /* end of pushDown */


/**Function********************************************************************

  Synopsis    [Gathers symmetry information.]

  Description [Translates the symmetry information stored in the next
  field of each subtable from level to indices. This procedure is called
  immediately after symmetric sifting, so that the next fields are correct.
  By translating this informaton in terms of indices, we make it independent
  of subsequent reorderings. The format used is that of the next fields:
  a circular list where each variable points to the next variable in the
  same symmetry group. Only the entries between lower and upper are
  considered.  The procedure returns a pointer to an array
  holding the symmetry information if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [checkSymmInfo]

******************************************************************************/
func initSymmInfo(table *DdManager, lower int32, upper int32) *DdHalfWord {
    var level, index, next, nextindex int32;
    var symmInfo *DdHalfWord;

    symmInfo =  ABC_ALLOC(DdHalfWord, table.size);
    if symmInfo == nil { return(nil); }

    for level = lower; level <= upper; level++ {
        index = table.invperm[level];
        next =  table.subtables[level].next;
        nextindex = table.invperm[next];
        symmInfo[index] = nextindex;
    }
    return(symmInfo);

} /* end of initSymmInfo */


/**Function********************************************************************

  Synopsis    [Check symmetry condition.]

  Description [Returns 1 if a variable is the one with the highest index
  among those belonging to a symmetry group that are in the top part of
  the BDD.  The top part is given by level.]

  SideEffects [None]

  SeeAlso     [initSymmInfo]

******************************************************************************/
func checkSymmInfo(table *DdManager, symmInfo *DdHalfWord, index int32, level int32) int32 {
    var i int32;

    i = symmInfo[index];
    for i != index {
        if index < i && table.perm[i] <= level {
            return(0); }
        i = symmInfo[i];
    }
    return(1);

} /* end of checkSymmInfo */


//ABC_NAMESPACE_IMPL_END

