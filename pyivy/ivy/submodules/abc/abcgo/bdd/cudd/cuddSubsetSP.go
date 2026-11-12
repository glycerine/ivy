package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddSubsetSP.c]

  PackageName [cudd]

  Synopsis [Procedure to subset the given BDD choosing the shortest paths
            (largest cubes) in the BDD.]


  Description  [External procedures included in this module:
                <ul>
                <li> Cudd_SubsetShortPaths()
                <li> Cudd_SupersetShortPaths()
                </ul>
                Internal procedures included in this module:
                <ul>
                <li> cuddSubsetShortPaths()
                </ul>
                Static procedures included in this module:
                <ul>
                <li> BuildSubsetBdd()
                <li> CreatePathTable()
                <li> AssessPathLength()
                <li> CreateTopDist()
                <li> CreateBotDist()
                <li> ResizeNodeDistPages()
                <li> ResizeQueuePages()
                <li> stPathTableDdFree()
                </ul>
                ]

  SeeAlso     [cuddSubsetHB.c]

  Author      [Kavita Ravi]

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

const DEFAULT_PAGE_SIZE = 2048 /* page size to store the BFS queue element type */
const DEFAULT_NODE_DIST_PAGE_SIZE = 2048 /*  page sizesto store NodeDist_t type */
#define MAXSHORTINT     ((DdHalfWord) ~0) /* constant defined to store
                                           * maximum distance of a node
                                           * from the root or the
                                           * constant
                                           */
const INITIAL_PAGES = 128 /* number of initial pages for the
                           * queue/NodeDist_t type */

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/* structure created to store subset results for each node and distances with
 * odd and even parity of the node from the root and sink. Main data structure
 * in this procedure.
 */
type NodeDist struct {
    oddTopDist DdHalfWord;
    evenTopDist DdHalfWord;
    oddBotDist DdHalfWord;
    evenBotDist DdHalfWord;
    regResult *DdNode;
    compResult *DdNode;
}

/* assorted information needed by the BuildSubsetBdd procedure. */
type AssortedInfo struct {
    unsigned int maxpath;
    findShortestPath int32;
    thresholdReached int32;
    maxpathTable *st__table;
    threshold int32;
}

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

type NodeDist_t = NodeDist;

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddSubsetSP.c,v 1.34 2009/02/19 16:23:19 fabio Exp $";
#endif

#ifdef DD_DEBUG
static int numCalls;
static int hits;
static int thishit;
#endif


var memOut int32; /* flag to indicate out of memory */
var zero, one *DdNode; /* constant functions */

var nodeDistPages **NodeDist_t; /* pointers to the pages */
var nodeDistPageIndex int32; /* index to next element */
var nodeDistPage int32; /* index to current page */
var nodeDistPageSize int32 = DEFAULT_NODE_DIST_PAGE_SIZE; /* page size */
var maxNodeDistPages int32; /* number of page pointers */
var currentNodeDistPage *NodeDist_t; /* current page */

var queuePages ***DdNode; /* pointers to the pages */
var queuePageIndex int32; /* index to next element */
var queuePage int32; /* index to current page */
var queuePageSize int32 = DEFAULT_PAGE_SIZE; /* page size */
var maxQueuePages int32; /* number of page pointers */
var currentQueuePage **DdNode; /* current page */


/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static void ResizeNodeDistPages (void);

// static void ResizeQueuePages (void);

// static void CreateTopDist ( st__table *pathTable, int parentPage, int parentQueueIndex, int topLen, DdNode **childPage, int childQueueIndex, int numParents, FILE *fp);

// static int CreateBotDist (DdNode *node, st__table *pathTable, unsigned int *pathLengthArray, FILE *fp);

// static st__table * CreatePathTable (DdNode *node, unsigned int *pathLengthArray, FILE *fp);

// static unsigned int AssessPathLength (unsigned int *pathLengthArray, int threshold, int numVars, unsigned int *excess, FILE *fp);

// static DdNode * BuildSubsetBdd (DdManager *dd, st__table *pathTable, DdNode *node, struct AssortedInfo *info, st__table *subsetNodeTable);

// static      st__retval stPathTableDdFree (char *key, char *value, char *arg);


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of Exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Extracts a dense subset from a BDD with the shortest paths
  heuristic.]

  Description [Extracts a dense subset from a BDD.  This procedure
  tries to preserve the shortest paths of the input BDD, because they
  give many minterms and contribute few nodes.  This procedure may
  increase the number of nodes in trying to create the subset or
  reduce the number of nodes due to recombination as compared to the
  original BDD. Hence the threshold may not be strictly adhered to. In
  practice, recombination overshadows the increase in the number of
  nodes and results in small BDDs as compared to the threshold. The
  hardlimit specifies whether threshold needs to be strictly adhered
  to. If it is set to 1, the procedure ensures that result is never
  larger than the specified limit but may be considerably less than
  the threshold.  Returns a pointer to the BDD for the subset if
  successful; NULL otherwise.  The value for numVars should be as
  close as possible to the size of the support of f for better
  efficiency. However, it is safe to pass the value returned by
  Cudd_ReadSize for numVars. If 0 is passed, then the value returned
  by Cudd_ReadSize is used.]

  SideEffects [None]

  SeeAlso     [Cudd_SupersetShortPaths Cudd_SubsetHeavyBranch Cudd_ReadSize]

******************************************************************************/
DdNode *
Cudd_SubsetShortPaths(
  DdManager * dd /* manager */,
  DdNode * f /* function to be subset */,
  int  numVars /* number of variables in the support of f */,
  int  threshold /* maximum number of nodes in the subset */,
  int  hardlimit /* flag: 1 if threshold is a hard limit */)
{
    var subset *DdNode;

    memOut = 0;
    for {
        dd.reordered = 0;
        subset = cuddSubsetShortPaths(dd, f, numVars, threshold, hardlimit);
    
if !((dd.reordered ==1) && (!(memOut != 0))) { break }
} 

    return(subset);

} /* end of Cudd_SubsetShortPaths */


/**Function********************************************************************

  Synopsis    [Extracts a dense superset from a BDD with the shortest paths
  heuristic.]

  Description [Extracts a dense superset from a BDD.  The procedure is
  identical to the subset procedure except for the fact that it
  receives the complement of the given function. Extracting the subset
  of the complement function is equivalent to extracting the superset
  of the function.  This procedure tries to preserve the shortest
  paths of the complement BDD, because they give many minterms and
  contribute few nodes.  This procedure may increase the number of
  nodes in trying to create the superset or reduce the number of nodes
  due to recombination as compared to the original BDD. Hence the
  threshold may not be strictly adhered to. In practice, recombination
  overshadows the increase in the number of nodes and results in small
  BDDs as compared to the threshold.  The hardlimit specifies whether
  threshold needs to be strictly adhered to. If it is set to 1, the
  procedure ensures that result is never larger than the specified
  limit but may be considerably less than the threshold. Returns a
  pointer to the BDD for the superset if successful; NULL
  otherwise. The value for numVars should be as close as possible to
  the size of the support of f for better efficiency.  However, it is
  safe to pass the value returned by Cudd_ReadSize for numVar.  If 0
  is passed, then the value returned by Cudd_ReadSize is used.]

  SideEffects [None]

  SeeAlso     [Cudd_SubsetShortPaths Cudd_SupersetHeavyBranch Cudd_ReadSize]

******************************************************************************/
DdNode *
Cudd_SupersetShortPaths(
  DdManager * dd /* manager */,
  DdNode * f /* function to be superset */,
  int  numVars /* number of variables in the support of f */,
  int  threshold /* maximum number of nodes in the subset */,
  int  hardlimit /* flag: 1 if threshold is a hard limit */)
{
    var subset, g *DdNode;

    // inlining of macro Cudd_Not
    g = ((DdNode *)((ptrint)(f) ^ 01));
    memOut = 0;
    for {
        dd.reordered = 0;
        subset = cuddSubsetShortPaths(dd, g, numVars, threshold, hardlimit);
    
if !((dd.reordered ==1) && (!(memOut != 0))) { break }
} 

    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(subset) ^ ((subset != nil)))));

} /* end of Cudd_SupersetShortPaths */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [The outermost procedure to return a subset of the given BDD
  with the shortest path lengths.]

  Description [The outermost procedure to return a subset of the given
  BDD with the largest cubes. The path lengths are calculated, the maximum
  allowable path length is determined and the number of nodes of this
  path length that can be used to build a subset. If the threshold is
  larger than the size of the original BDD, the original BDD is
  returned. ]

  SideEffects [None]

  SeeAlso     [Cudd_SubsetShortPaths]

******************************************************************************/
DdNode *
cuddSubsetShortPaths(
  DdManager * dd /* DD manager */,
  DdNode * f /* function to be subset */,
  int  numVars /* total number of variables in consideration */,
  int  threshold /* maximum number of nodes allowed in the subset */,
  int  hardlimit /* flag determining whether thershold should be respected strictly */)
{
    var pathTable *st__table;
    var N, subset *DdNode;

    unsigned int  *pathLengthArray;
    unsigned int maxpath, oddLen, evenLen, pathLength, *excess;
    var i int32;
    var nodeStat *NodeDist_t;
    struct AssortedInfo *info;
    var subsetNodeTable *st__table;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    if numVars == 0 {
      /* set default value */
      numVars = Cudd_ReadSize(dd);
    }

    if threshold > numVars {
        threshold = threshold - numVars;
    }
    if f == nil {
        fmt.Fprintf(dd.err, "Cannot partition, nil object\n");
        dd.errorCode = CUDD_INVALID_ARG;
        return(nil);
    }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        return (f); }

    pathLengthArray = ABC_ALLOC(unsigned int, numVars+1);
    for i = 0; i < numVars+1; i++ { pathLengthArray[i] = 0; }


#ifdef DD_DEBUG
    numCalls = 0;
#endif

    pathTable = CreatePathTable(f, pathLengthArray, dd.err);

    if (pathTable == nil) || ((memOut != 0)) {
        if pathTable != nil {
            st__free_table(pathTable); }
        ABC_FREE(pathLengthArray);
        return (NIL(DdNode));
    }

    excess = ABC_ALLOC(unsigned int, 1);
    *excess = 0;
    maxpath = AssessPathLength(pathLengthArray, threshold, numVars, excess,
                               dd.err);

    if maxpath != uint32( (numVars + 1)) {

        info = ABC_ALLOC(struct AssortedInfo, 1);
        info.maxpath = maxpath;
        info.findShortestPath = 0;
        info.thresholdReached = *excess;
        info.maxpathTable = st__init_table( st__ptrcmp, st__ptrhash);
        info.threshold = threshold;

#ifdef DD_DEBUG
        (void) fprintf(dd->out, "Path length array\n");
        for (i = 0; i < (numVars+1); i++) {
            if (pathLengthArray[i])
                (void) fprintf(dd->out, "%d ",i);
        }
        (void) fprintf(dd->out, "\n");
        for (i = 0; i < (numVars+1); i++) {
            if (pathLengthArray[i])
                (void) fprintf(dd->out, "%d ",pathLengthArray[i]);
        }
        (void) fprintf(dd->out, "\n");
        (void) fprintf(dd->out, "Maxpath  = %d, Thresholdreached = %d\n",
                       maxpath, info->thresholdReached);
#endif

        // inlining of macro Cudd_Regular
        N = ((DdNode *)((ptruint)(f) & ~01));
        if ! st__lookup(pathTable, (char *)N, (char **)&nodeStat) {
            fmt.Fprintf(dd.err, "Something wrong, root node must be in table\n");
            dd.errorCode = CUDD_INTERNAL_ERROR;
            ABC_FREE(excess);
            ABC_FREE(info);
            return(nil);
        } else {
            if (nodeStat.oddTopDist != MAXSHORTINT) &&
                (nodeStat.oddBotDist != MAXSHORTINT) {
                oddLen = (nodeStat.oddTopDist + nodeStat.oddBotDist);
            } else {
                oddLen = MAXSHORTINT; }

            if (nodeStat.evenTopDist != MAXSHORTINT) &&
                (nodeStat.evenBotDist != MAXSHORTINT) {
                evenLen = (nodeStat.evenTopDist +nodeStat.evenBotDist);
            } else {
                evenLen = MAXSHORTINT; }

            if oddLen <= evenLen {
                pathLength = oddLen;
            } else {
                pathLength = evenLen;
            }
            if pathLength > maxpath {
                (void) fmt.Fprintf(dd.err, "All computations are bogus, since root has path length greater than max path length within threshold %d, %d\n", maxpath, pathLength);
                dd.errorCode = CUDD_INTERNAL_ERROR;
                return(nil);
            }
        }

#ifdef DD_DEBUG
        numCalls = 0;
        hits = 0;
        thishit = 0;
#endif
        /* initialize a table to store computed nodes */
        if hardlimit {
            subsetNodeTable = st__init_table( st__ptrcmp, st__ptrhash);
        } else {
            subsetNodeTable = NIL( st__table);
        }
        subset = BuildSubsetBdd(dd, pathTable, f, info, subsetNodeTable);
        if subset != nil {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(subset).ref);
        }
        /* record the number of times a computed result for a node is hit */

#ifdef DD_DEBUG
        (void) fprintf(dd->out, "Hits = %d, New==Node = %d, NumCalls = %d\n",
                hits, thishit, numCalls);
#endif

        if subsetNodeTable != NIL( st__table) {
            st__free_table(subsetNodeTable);
        }
        st__free_table(info.maxpathTable);
        st__foreach(pathTable, stPathTableDdFree, (char *)dd);

        ABC_FREE(info);

    } else {/* if threshold larger than size of dd */
        subset = f;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(subset).ref);
    }
    ABC_FREE(excess);
    st__free_table(pathTable);
    ABC_FREE(pathLengthArray);
    for i = 0; i <= nodeDistPage; i++ { ABC_FREE(nodeDistPages[i]); }
    ABC_FREE(nodeDistPages);

#ifdef DD_DEBUG
    /* check containment of subset in f */
    if (subset != NULL) {
        DdNode *check;
        check = Cudd_bddIteConstant(dd, subset, f, one);
        if (check != one) {
            (void) fprintf(dd->err, "Wrong partition\n");
            dd->errorCode = CUDD_INTERNAL_ERROR;
            return(NULL);
        }
    }
#endif

    if subset != nil {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(subset).ref);
        return(subset);
    } else {
        return(nil);
    }

} /* end of cuddSubsetShortPaths */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Resize the number of pages allocated to store the distances
  related to each node.]

  Description [Resize the number of pages allocated to store the distances
  related to each node. The procedure  moves the counter to the
  next page when the end of the page is reached and allocates new
  pages when necessary. ]

  SideEffects [Changes the size of  pages, page, page index, maximum
  number of pages freeing stuff in case of memory out. ]

  SeeAlso     []

******************************************************************************/
func ResizeNodeDistPages() {
    var i int32;
    var newNodeDistPages **NodeDist_t;

    /* move to next page */
    nodeDistPage++;

    /* If the current page index is larger than the number of pages
     * allocated, allocate a new page array. Page numbers are incremented by
     * INITIAL_PAGES
     */
    if nodeDistPage == maxNodeDistPages {
        newNodeDistPages = ABC_ALLOC(NodeDist_t *,maxNodeDistPages + INITIAL_PAGES);
        if newNodeDistPages == nil {
            for i = 0; i < nodeDistPage; i++ { ABC_FREE(nodeDistPages[i]); }
            ABC_FREE(nodeDistPages);
            memOut = 1;
            return;
        } else {
            for i = 0; i < maxNodeDistPages; i++ {
                newNodeDistPages[i] = nodeDistPages[i];
            }
            /* Increase total page count */
            maxNodeDistPages += INITIAL_PAGES;
            ABC_FREE(nodeDistPages);
            nodeDistPages = newNodeDistPages;
        }
    }
    /* Allocate a new page */
    currentNodeDistPage = nodeDistPages[nodeDistPage] = ABC_ALLOC(NodeDist_t,
                                                                  nodeDistPageSize);
    if currentNodeDistPage == nil {
        for i = 0; i < nodeDistPage; i++ { ABC_FREE(nodeDistPages[i]); }
        ABC_FREE(nodeDistPages);
        memOut = 1;
        return;
    }
    /* reset page index */
    nodeDistPageIndex = 0;
    return;

} /* end of ResizeNodeDistPages */


/**Function********************************************************************

  Synopsis    [Resize the number of pages allocated to store nodes in the BFS
  traversal of the Bdd  .]

  Description [Resize the number of pages allocated to store nodes in the BFS
  traversal of the Bdd. The procedure  moves the counter to the
  next page when the end of the page is reached and allocates new
  pages when necessary.]

  SideEffects [Changes the size of pages, page, page index, maximum
  number of pages freeing stuff in case of memory out. ]

  SeeAlso     []

******************************************************************************/
func ResizeQueuePages() {
    var i int32;
    var newQueuePages ***DdNode;

    queuePage++;
    /* If the current page index is larger than the number of pages
     * allocated, allocate a new page array. Page numbers are incremented by
     * INITIAL_PAGES
     */
    if queuePage == maxQueuePages {
        newQueuePages = ABC_ALLOC(DdNode **,maxQueuePages + INITIAL_PAGES);
        if newQueuePages == nil {
            for i = 0; i < queuePage; i++ { ABC_FREE(queuePages[i]); }
            ABC_FREE(queuePages);
            memOut = 1;
            return;
        } else {
            for i = 0; i < maxQueuePages; i++ {
                newQueuePages[i] = queuePages[i];
            }
            /* Increase total page count */
            maxQueuePages += INITIAL_PAGES;
            ABC_FREE(queuePages);
            queuePages = newQueuePages;
        }
    }
    /* Allocate a new page */
    currentQueuePage = queuePages[queuePage] = ABC_ALLOC(DdNode *,queuePageSize);
    if currentQueuePage == nil {
        for i = 0; i < queuePage; i++ { ABC_FREE(queuePages[i]); }
        ABC_FREE(queuePages);
        memOut = 1;
        return;
    }
    /* reset page index */
    queuePageIndex = 0;
    return;

} /* end of ResizeQueuePages */


/**Function********************************************************************

  Synopsis    [ Labels each node with its shortest distance from the root]

  Description [ Labels each node with its shortest distance from the root.
  This is done in a BFS search of the BDD. The nodes are processed
  in a queue implemented as pages(array) to reduce memory fragmentation.
  An entry is created for each node visited. The distance from the root
  to the node with the corresponding  parity is updated. The procedure
  is called recursively each recusion level handling nodes at a given
  level from the root.]


  SideEffects [Creates entries in the pathTable]

  SeeAlso     [CreatePathTable CreateBotDist]

******************************************************************************/
static void
CreateTopDist(
  st__table * pathTable /* hast table to store path lengths */,
  int  parentPage /* the pointer to the page on which the first parent in the queue is to be found. */,
  int  parentQueueIndex /* pointer to the first parent on the page */,
  int  topLen /* current distance from the root */,
  DdNode ** childPage /* pointer to the page on which the first child is to be added. */,
  int  childQueueIndex /* pointer to the first child */,
  int  numParents /* number of parents to process in this recursive call */,
  FILE *fp /* where to write messages */)
{
    var nodeStat *NodeDist_t;
    var N, Nv, Nnv, node, child, regChild *DdNode;
    var i int32;
    var processingDone, childrenCount int32;

#ifdef DD_DEBUG
    numCalls++;

    /* assume this procedure comes in with only the root node*/
    /* set queue index to the next available entry for addition */
    /* set queue page to page of addition */
    if ((queuePages[parentPage] == childPage) && (parentQueueIndex ==
                                                  childQueueIndex)) {
        fprintf(fp, "Should not happen that they are equal\n");
    }
    assert(queuePageIndex == childQueueIndex);
    assert(currentQueuePage == childPage);
#endif
    /* number children added to queue is initialized , needed for
     * numParents in the next call
     */
    childrenCount = 0;
    /* process all the nodes in this level */
    for numParents != 0 {
        numParents--;
        if parentQueueIndex == queuePageSize {
            parentPage++;
            parentQueueIndex = 0;
        }
        /* a parent to process */
        node = *(queuePages[parentPage] + parentQueueIndex);
        parentQueueIndex++;
        /* get its children */
        // inlining of macro Cudd_Regular
        N = ((DdNode *)((ptruint)(node) & ~01));
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        Nv = ((((DdNode *)((ptruint)(N) & ~01))).type.kids.T);
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        Nnv = ((((DdNode *)((ptruint)(N) & ~01))).type.kids.E);

        // inlining of macro Cudd_NotCond
        // inlining of macro Cudd_IsComplement
        Nv = ((DdNode *)((ptrint)(Nv) ^ ((int32( ((ptrint) (node) & 01))))));
        // inlining of macro Cudd_NotCond
        // inlining of macro Cudd_IsComplement
        Nnv = ((DdNode *)((ptrint)(Nnv) ^ ((int32( ((ptrint) (node) & 01))))));

        processingDone = 2;
        for processingDone != 0 {
            /* processing the THEN and the ELSE children, the THEN
             * child first
             */
            if processingDone == 2 {
                child = Nv;
            } else {
                child = Nnv;
            }

            // inlining of macro Cudd_Regular
            regChild = ((DdNode *)((ptruint)(child) & ~01));
            /* dont process if the child is a constant */
            // inlining of macro Cudd_IsConstant
            // inlining of macro Cudd_Regular
            if !((((DdNode *)((ptruint)(child) & ~01))).index == CUDD_CONST_INDEX) {
                /* check is already visited, if not add a new entry in
                 * the path Table
                 */
                if ! st__lookup(pathTable, (char *)regChild, (char **)&nodeStat) {
                    /* if not in table, has never been visited */
                    /* create entry for table */
                    if nodeDistPageIndex == nodeDistPageSize {
                        ResizeNodeDistPages(); }
                    if memOut != 0 {
                        for i = 0; i <= queuePage; i++ { ABC_FREE(queuePages[i]); }
                        ABC_FREE(queuePages);
                        st__free_table(pathTable);
                        return;
                    }
                    /* New entry for child in path Table is created here */
                    nodeStat = currentNodeDistPage + nodeDistPageIndex;
                    nodeDistPageIndex++;

                    /* Initialize fields of the node data */
                    nodeStat.oddTopDist = MAXSHORTINT;
                    nodeStat.evenTopDist = MAXSHORTINT;
                    nodeStat.evenBotDist = MAXSHORTINT;
                    nodeStat.oddBotDist = MAXSHORTINT;
                    nodeStat.regResult = nil;
                    nodeStat.compResult = nil;
                    /* update the table entry element, the distance keeps
                     * track of the parity of the path from the root
                     */
                    // inlining of macro Cudd_IsComplement
                    if int32( ((ptrint) (child) & 01)) != 0 {
                        nodeStat.oddTopDist = (DdHalfWord) topLen + 1;
                    } else {
                        nodeStat.evenTopDist = (DdHalfWord) topLen + 1;
                    }

                    /* insert entry element for  child in the table */
                    if st__insert(pathTable, (char *)regChild,
                                  (char *)nodeStat) == st__OUT_OF_MEM {
                        memOut = 1;
                        for i = 0; i <= nodeDistPage; i++ {
                            ABC_FREE(nodeDistPages[i]);
                        }
                        ABC_FREE(nodeDistPages);
                        for i = 0; i <= queuePage; i++ { ABC_FREE(queuePages[i]); }
                        ABC_FREE(queuePages);
                        st__free_table(pathTable);
                        return;
                    }

                    /* Create list element for this child to process its children.
                     * If this node has been processed already, then it appears
                     * in the path table and hence is never added to the list
                     * again.
                     */

                    if queuePageIndex == queuePageSize { ResizeQueuePages(); }
                    if memOut != 0 {
                        for i = 0; i <= nodeDistPage; i++ {
                            ABC_FREE(nodeDistPages[i]);
                        }
                        ABC_FREE(nodeDistPages);
                        st__free_table(pathTable);
                        return;
                    }
                    *(currentQueuePage + queuePageIndex) = child;
                    queuePageIndex++;

                    childrenCount++;
                } else {
                    /* if not been met in a path with this parity before */
                    /* put in list */
                    // inlining of macro Cudd_IsComplement
                    if ((((int32( ((ptrint) (child) & 01)) != 0))) && (nodeStat.oddTopDist ==
                          // inlining of macro Cudd_IsComplement
                          MAXSHORTINT)) || ((!((int32( ((ptrint) (child) & 01)) != 0))) &&
                                  (nodeStat.evenTopDist == MAXSHORTINT)) {

                        if queuePageIndex == queuePageSize { ResizeQueuePages(); }
                        if memOut != 0 {
                            for i = 0; i <= nodeDistPage; i++ {
                                ABC_FREE(nodeDistPages[i]);
                            }
                            ABC_FREE(nodeDistPages);
                            st__free_table(pathTable);
                            return;

                        }
                        *(currentQueuePage + queuePageIndex) = child;
                        queuePageIndex++;

                        /* update the distance with the appropriate parity */
                        // inlining of macro Cudd_IsComplement
                        if int32( ((ptrint) (child) & 01)) != 0 {
                            nodeStat.oddTopDist = (DdHalfWord) topLen + 1;
                        } else {
                            nodeStat.evenTopDist = (DdHalfWord) topLen + 1;
                        }
                        childrenCount++;
                    }

                } /* end of else (not found in st__table) */
            } /*end of if Not constant child */
            processingDone--;
        } /*end of while processing Nv, Nnv */
    }  /*end of while numParents */

#ifdef DD_DEBUG
    assert(queuePages[parentPage] == childPage);
    assert(parentQueueIndex == childQueueIndex);
#endif

    if childrenCount != 0 {
        topLen++;
        childPage = currentQueuePage;
        childQueueIndex = queuePageIndex;
        CreateTopDist(pathTable, parentPage, parentQueueIndex, topLen,
                      childPage, childQueueIndex, childrenCount, fp);
    }

    return;

} /* end of CreateTopDist */


/**Function********************************************************************

  Synopsis    [ Labels each node with the shortest distance from the constant.]

  Description [Labels each node with the shortest distance from the constant.
  This is done in a DFS search of the BDD. Each node has an odd
  and even parity distance from the sink (since there exists paths to both
  zero and one) which is less than MAXSHORTINT. At each node these distances
  are updated using the minimum distance of its children from the constant.
  SInce now both the length from the root and child is known, the minimum path
  length(length of the shortest path between the root and the constant that
  this node lies on) of this node can be calculated and used to update the
  pathLengthArray]

  SideEffects [Updates Path Table and path length array]

  SeeAlso     [CreatePathTable CreateTopDist AssessPathLength]

******************************************************************************/
static int
CreateBotDist(
  DdNode * node /* current node */,
  st__table * pathTable /* path table with path lengths */,
  unsigned int * pathLengthArray /* array that stores number of nodes belonging to a particular path length. */,
  FILE *fp /* where to write messages */)
{
    var N, Nv, Nnv *DdNode;
    var realChild *DdNode;
    var child, regChild *DdNode;
    var nodeStat, nodeStatChild *NodeDist_t;
    unsigned int  oddLen, evenLen, pathLength;
    var botDist DdHalfWord;
    var processingDone int32;

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(node) & ~01))).index == CUDD_CONST_INDEX {
        return(1); }
    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    /* each node has one table entry */
    /* update as you go down the min dist of each node from
       the root in each (odd and even) parity */
    if ! st__lookup(pathTable, (char *)N, (char **)&nodeStat) {
        fmt.Fprintf(fp, "Something wrong, the entry doesn't exist\n");
        return(0);
    }

    /* compute length of odd parity distances */
    if (nodeStat.oddTopDist != MAXSHORTINT) &&
        (nodeStat.oddBotDist != MAXSHORTINT) {
        oddLen = (nodeStat.oddTopDist + nodeStat.oddBotDist);
    } else {
        oddLen = MAXSHORTINT; }

    /* compute length of even parity distances */
    if !((nodeStat.evenTopDist == MAXSHORTINT) ||
          (nodeStat.evenBotDist == MAXSHORTINT)) {
        evenLen = (nodeStat.evenTopDist +nodeStat.evenBotDist);
    } else {
        evenLen = MAXSHORTINT; }

    /* assign pathlength to minimum of the two */
    if oddLen <= evenLen {
        pathLength = oddLen;
    } else {
        pathLength = evenLen;
    }

    // inlining of macro Cudd_T
    // inlining of macro Cudd_Regular
    Nv = ((((DdNode *)((ptruint)(N) & ~01))).type.kids.T);
    // inlining of macro Cudd_E
    // inlining of macro Cudd_Regular
    Nnv = ((((DdNode *)((ptruint)(N) & ~01))).type.kids.E);

    /* process each child */
    processingDone = 0;
    for processingDone != 2 {
        if !(processingDone != 0) {
            child = Nv;
        } else {
            child = Nnv;
        }

        // inlining of macro Cudd_NotCond
        // inlining of macro Cudd_IsComplement
        realChild = ((DdNode *)((ptrint)(child) ^ ((int32( ((ptrint) (node) & 01))))));
        // inlining of macro Cudd_Regular
        regChild = ((DdNode *)((ptruint)(child) & ~01));
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if (((DdNode *)((ptruint)(realChild) & ~01))).index == CUDD_CONST_INDEX {
            /* Found a minterm; count parity and shortest distance
            ** from the constant.
            */
            // inlining of macro Cudd_IsComplement
            if int32( ((ptrint) (child) & 01)) != 0 {
                nodeStat.oddBotDist = 1;
            } else {
                nodeStat.evenBotDist = 1; }
        } else {
            /* If node not in table, recur. */
            if ! st__lookup(pathTable, (char *)regChild, (char **)&nodeStatChild) {
                fmt.Fprintf(fp, "Something wrong, node in table should have been created in top dist proc.\n");
                return(0);
            }

            if nodeStatChild.oddBotDist == MAXSHORTINT {
                if nodeStatChild.evenBotDist == MAXSHORTINT {
                    if !CreateBotDist(realChild, pathTable, pathLengthArray, fp) {
                        return(0); }
                } else {
                    fmt.Fprintf(fp, "Something wrong, both bot nodeStats should be there\n");
                    return(0);
                }
            }

            /* Update shortest distance from the constant depending on
            **  parity. */

            // inlining of macro Cudd_IsComplement
            if int32( ((ptrint) (child) & 01)) != 0 {
                /* If parity on the edge then add 1 to even distance
                ** of child to get odd parity distance and add 1 to
                ** odd distance of child to get even parity
                ** distance. Change distance of current node only if
                ** the calculated distance is less than existing
                ** distance. */
                if nodeStatChild.oddBotDist != MAXSHORTINT {
                    botDist = nodeStatChild.oddBotDist + 1;
                } else {
                    botDist = MAXSHORTINT; }
                if nodeStat.evenBotDist > botDist {
                    nodeStat.evenBotDist = botDist; }

                if nodeStatChild.evenBotDist != MAXSHORTINT {
                    botDist = nodeStatChild.evenBotDist + 1;
                } else {
                    botDist = MAXSHORTINT; }
                if nodeStat.oddBotDist > botDist {
                    nodeStat.oddBotDist = botDist; }

            } else {
                /* If parity on the edge then add 1 to even distance
                ** of child to get even parity distance and add 1 to
                ** odd distance of child to get odd parity distance.
                ** Change distance of current node only if the
                ** calculated distance is lesser than existing
                ** distance. */
                if nodeStatChild.evenBotDist != MAXSHORTINT {
                    botDist = nodeStatChild.evenBotDist + 1;
                } else {
                    botDist = MAXSHORTINT; }
                if nodeStat.evenBotDist > botDist {
                    nodeStat.evenBotDist = botDist; }

                if nodeStatChild.oddBotDist != MAXSHORTINT {
                    botDist = nodeStatChild.oddBotDist + 1;
                } else {
                    botDist = MAXSHORTINT; }
                if nodeStat.oddBotDist > botDist {
                    nodeStat.oddBotDist = botDist; }
            }
        } /* end of else (if not constant child ) */
        processingDone++;
    } /* end of while processing Nv, Nnv */

    /* Compute shortest path length on the fly. */
    if (nodeStat.oddTopDist != MAXSHORTINT) &&
        (nodeStat.oddBotDist != MAXSHORTINT) {
        oddLen = (nodeStat.oddTopDist + nodeStat.oddBotDist);
    } else {
        oddLen = MAXSHORTINT; }

    if (nodeStat.evenTopDist != MAXSHORTINT) &&
        (nodeStat.evenBotDist != MAXSHORTINT) {
        evenLen = (nodeStat.evenTopDist +nodeStat.evenBotDist);
    } else {
        evenLen = MAXSHORTINT; }

    /* Update path length array that has number of nodes of a particular
    ** path length. */
    if oddLen < pathLength {
        if pathLength != MAXSHORTINT {
            pathLengthArray[pathLength]--; }
        if oddLen != MAXSHORTINT {
            pathLengthArray[oddLen]++; }
        pathLength = oddLen;
    }
    if evenLen < pathLength {
        if pathLength != MAXSHORTINT {
            pathLengthArray[pathLength]--; }
        if evenLen != MAXSHORTINT {
            pathLengthArray[evenLen]++; }
    }

    return(1);

} /*end of CreateBotDist */


/**Function********************************************************************

  Synopsis    [ The outer procedure to label each node with its shortest
  distance from the root and constant]

  Description [ The outer procedure to label each node with its shortest
  distance from the root and constant. Calls CreateTopDist and CreateBotDist.
  The basis for computing the distance between root and constant is that
  the distance may be the sum of even distances from the node to the root
  and constant or the sum of odd distances from the node to the root and
  constant.  Both CreateTopDist and CreateBotDist create the odd and
  even parity distances from the root and constant respectively.]

  SideEffects [None]

  SeeAlso     [CreateTopDist CreateBotDist]

******************************************************************************/
static st__table *
CreatePathTable(
  DdNode * node /* root of function */,
  unsigned int * pathLengthArray /* array of path lengths to store nodes labeled with the various path lengths */,
  FILE *fp /* where to write messages */)
{

    var pathTable *st__table;
    var nodeStat *NodeDist_t;
    var topLen DdHalfWord;
    var N *DdNode;
    var i, numParents int32;
    var insertValue int32;
    var childPage **DdNode;
    var parentPage int32;
    var childQueueIndex, parentQueueIndex int32;

    /* Creating path Table for storing data about nodes */
    pathTable = st__init_table( st__ptrcmp, st__ptrhash);

    /* initializing pages for info about each node */
    maxNodeDistPages = INITIAL_PAGES;
    nodeDistPages = ABC_ALLOC(NodeDist_t *, maxNodeDistPages);
    if nodeDistPages == nil {
        goto OUT_OF_MEM;
    }
    nodeDistPage = 0;
    currentNodeDistPage = nodeDistPages[nodeDistPage] =
        ABC_ALLOC(NodeDist_t, nodeDistPageSize);
    if currentNodeDistPage == nil {
        for i = 0; i <= nodeDistPage; i++ { ABC_FREE(nodeDistPages[i]); }
        ABC_FREE(nodeDistPages);
        goto OUT_OF_MEM;
    }
    nodeDistPageIndex = 0;

    /* Initializing pages for the BFS search queue, implemented as an array. */
    maxQueuePages = INITIAL_PAGES;
    queuePages = ABC_ALLOC(DdNode **, maxQueuePages);
    if queuePages == nil {
        goto OUT_OF_MEM;
    }
    queuePage = 0;
    currentQueuePage  = queuePages[queuePage] = ABC_ALLOC(DdNode *, queuePageSize);
    if currentQueuePage == nil {
        for i = 0; i <= queuePage; i++ { ABC_FREE(queuePages[i]); }
        ABC_FREE(queuePages);
        goto OUT_OF_MEM;
    }
    queuePageIndex = 0;

    /* Enter the root node into the queue to start with. */
    parentPage = queuePage;
    parentQueueIndex = queuePageIndex;
    topLen = 0;
    *(currentQueuePage + queuePageIndex) = node;
    queuePageIndex++;
    childPage = currentQueuePage;
    childQueueIndex = queuePageIndex;

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));

    if nodeDistPageIndex == nodeDistPageSize { ResizeNodeDistPages(); }
    if memOut != 0 {
        for i = 0; i <= nodeDistPage; i++ { ABC_FREE(nodeDistPages[i]); }
        ABC_FREE(nodeDistPages);
        for i = 0; i <= queuePage; i++ { ABC_FREE(queuePages[i]); }
        ABC_FREE(queuePages);
        st__free_table(pathTable);
        goto OUT_OF_MEM;
    }

    nodeStat = currentNodeDistPage + nodeDistPageIndex;
    nodeDistPageIndex++;

    nodeStat.oddTopDist = MAXSHORTINT;
    nodeStat.evenTopDist = MAXSHORTINT;
    nodeStat.evenBotDist = MAXSHORTINT;
    nodeStat.oddBotDist = MAXSHORTINT;
    nodeStat.regResult = nil;
    nodeStat.compResult = nil;

    insertValue = st__insert(pathTable, (char *)N, (char *)nodeStat);
    if insertValue == st__OUT_OF_MEM {
        memOut = 1;
        for i = 0; i <= nodeDistPage; i++ { ABC_FREE(nodeDistPages[i]); }
        ABC_FREE(nodeDistPages);
        for i = 0; i <= queuePage; i++ { ABC_FREE(queuePages[i]); }
        ABC_FREE(queuePages);
        st__free_table(pathTable);
        goto OUT_OF_MEM;
    } else if insertValue == 1 {
        fmt.Fprintf(fp, "Something wrong, the entry exists but didnt show up in st__lookup\n");
        return(nil);
    }

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (node) & 01)) != 0 {
        nodeStat.oddTopDist = 0;
    } else {
        nodeStat.evenTopDist = 0;
    }
    numParents = 1;
    /* call the function that counts the distance of each node from the
     * root
     */
#ifdef DD_DEBUG
    numCalls = 0;
#endif
    CreateTopDist(pathTable, parentPage, parentQueueIndex, int32( topLen),
                  childPage, childQueueIndex, numParents, fp);
    if memOut != 0 {
        fmt.Fprintf(fp, "Out of Memory and cant count path lengths\n");
        goto OUT_OF_MEM;
    }

#ifdef DD_DEBUG
    numCalls = 0;
#endif
    /* call the function that counts the distance of each node from the
     * constant
     */
    if !CreateBotDist(node, pathTable, pathLengthArray, fp) { return(nil); }

    /* free BFS queue pages as no longer required */
    for i = 0; i <= queuePage; i++ { ABC_FREE(queuePages[i]); }
    ABC_FREE(queuePages);
    return(pathTable);

OUT_OF_MEM:
    (void) fmt.Fprintf(fp, "Out of Memory, cannot allocate pages\n");
    memOut = 1;
    return(nil);

} /*end of CreatePathTable */


/**Function********************************************************************

  Synopsis    [Chooses the maximum allowable path length of nodes under the
  threshold.]

  Description [Chooses the maximum allowable path length under each node.
  The corner cases are when the threshold is larger than the number
  of nodes in the BDD iself, in which case 'numVars + 1' is returned.
  If all nodes of a particular path length are needed, then the
  maxpath returned is the next one with excess nodes = 0;]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
static unsigned int
AssessPathLength(
  unsigned int * pathLengthArray /* array determining number of nodes belonging to the different path lengths */,
  int  threshold /* threshold to determine maximum allowable nodes in the subset */,
  int  numVars /* maximum number of variables */,
  unsigned int * excess /* number of nodes labeled maxpath required in the subset */,
  FILE *fp /* where to write messages */)
{
    unsigned int i, maxpath;
    var temp int32;

    temp = threshold;
    i = 0;
    maxpath = 0;
    /* quit loop if i reaches max number of variables or if temp reaches
     * below zero
     */
    for (i < uint32( numVars)+1) && (temp > 0) {
        if pathLengthArray[i] > 0 {
            maxpath = i;
            temp = temp - pathLengthArray[i];
        }
        i++;
    }
    /* if all nodes of max path are needed */
    if temp >= 0 {
        maxpath++; /* now maxpath  becomes the next maxppath or max number
                      of variables */
        *excess = 0;
    } else { /* normal case when subset required is less than size of
                original BDD */
        *excess = temp + pathLengthArray[maxpath];
    }

    if maxpath == 0 {
        fmt.Fprintf(fp, "Path Length array seems to be all zeroes, check\n");
    }
    return(maxpath);

} /* end of AssessPathLength */


/**Function********************************************************************

  Synopsis    [Builds the BDD with nodes labeled with path length less than or equal to maxpath]

  Description [Builds the BDD with nodes labeled with path length
  under maxpath and as many nodes labeled maxpath as determined by the
  threshold. The procedure uses the path table to determine which nodes
  in the original bdd need to be retained. This procedure picks a
  shortest path (tie break decided by taking the child with the shortest
  distance to the constant) and recurs down the path till it reaches the
  constant. the procedure then starts building the subset upward from
  the constant. All nodes labeled by path lengths less than the given
  maxpath are used to build the subset.  However, in the case of nodes
  that have label equal to maxpath, as many are chosen as required by
  the threshold. This number is stored in the info structure in the
  field thresholdReached. This field is decremented whenever a node
  labeled maxpath is encountered and the nodes labeled maxpath are
  aggregated in a maxpath table. As soon as the thresholdReached count
  goes to 0, the shortest path from this node to the constant is found.
  The extraction of nodes with the above labeling is based on the fact
  that each node, labeled with a path length, P, has at least one child
  labeled P or less. So extracting all nodes labeled a given path length
  P ensures complete paths between the root and the constant. Extraction
  of a partial number of nodes with a given path length may result in
  incomplete paths and hence the additional number of nodes are grabbed
  to complete the path. Since the Bdd is built bottom-up, other nodes
  labeled maxpath do lie on complete paths.  The procedure may cause the
  subset to have a larger or smaller number of nodes than the specified
  threshold. The increase in the number of nodes is caused by the
  building of a subset and the reduction by recombination. However in
  most cases, the recombination overshadows the increase and the
  procedure returns a result with lower number of nodes than specified.
  The subsetNodeTable is NIL when there is no hard limit on the number
  of nodes. Further efforts towards keeping the subset closer to the
  threshold number were abandoned in favour of keeping the procedure
  simple and fast.]

  SideEffects [SubsetNodeTable is changed if it is not NIL.]

  SeeAlso     []

******************************************************************************/
static DdNode *
BuildSubsetBdd(
  DdManager * dd /* DD manager */,
  st__table * pathTable /* path table with path lengths and computed results */,
  DdNode * node /* current node */,
  struct AssortedInfo * info /* assorted information structure */,
  st__table * subsetNodeTable /* table storing computed results */)
{
    var N, Nv, Nnv *DdNode;
    var ThenBranch, ElseBranch, childBranch *DdNode;
    var child *DdNode;
    var regChild *DdNode;
    var regNnv *DdNode = nil;
    var regNv *DdNode = nil;
    var nodeStatNv, nodeStat, nodeStatNnv *NodeDist_t;
    var neW, topv, regNew *DdNode;
    var entry string;
    unsigned int topid;
    unsigned int childPathLength, oddLen, evenLen, NnvPathLength = 0, NvPathLength = 0;
    unsigned int NvBotDist, NnvBotDist;
    var tiebreakChild int32;
    var processingDone, thenDone, elseDone int32;


#ifdef DD_DEBUG
    numCalls++;
#endif
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(node) & ~01))).index == CUDD_CONST_INDEX {
        return(node); }

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    /* Find node in table. */
    if ! st__lookup(pathTable, (char *)N, (char **)&nodeStat) {
        (void) fmt.Fprintf(dd.err, "Something wrong, node must be in table \n");
        dd.errorCode = CUDD_INTERNAL_ERROR;
        return(nil);
    }
    /* If the node in the table has been visited, then return the corresponding
    ** Dd. Since a node can become a subset of itself, its
    ** complement (that is te same node reached by a different parity) will
    ** become a superset of the original node and result in some minterms
    ** that were not in the original set. Hence two different results are
    ** maintained, corresponding to the odd and even parities.
    */

    /* If this node is reached with an odd parity, get odd parity results. */
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (node) & 01)) != 0 {
        if  nodeStat.compResult != nil {
#ifdef DD_DEBUG
            hits++;
#endif
            return(nodeStat.compResult);
        }
    } else {
        /* if this node is reached with an even parity, get even parity
         * results
         */
        if nodeStat.regResult != nil {
#ifdef DD_DEBUG
            hits++;
#endif
            return(nodeStat.regResult);
        }
    }


    /* get children */
    // inlining of macro Cudd_T
    // inlining of macro Cudd_Regular
    Nv = ((((DdNode *)((ptruint)(N) & ~01))).type.kids.T);
    // inlining of macro Cudd_E
    // inlining of macro Cudd_Regular
    Nnv = ((((DdNode *)((ptruint)(N) & ~01))).type.kids.E);

    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nv = ((DdNode *)((ptrint)(Nv) ^ ((int32( ((ptrint) (node) & 01))))));
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nnv = ((DdNode *)((ptrint)(Nnv) ^ ((int32( ((ptrint) (node) & 01))))));

    /* no child processed */
    processingDone = 0;
    /* then child not processed */
    thenDone = 0;
    ThenBranch = nil;
    /* else child not processed */
    elseDone = 0;
    ElseBranch = nil;
    /* if then child constant, branch is the child */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(Nv) & ~01))).index == CUDD_CONST_INDEX {
        /*shortest path found */
        // inlining of macro DD_ONE
        if (Nv == ((dd).one)) && (info.findShortestPath) {
            info.findShortestPath = 0;
        }

        ThenBranch = Nv;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(ThenBranch).ref);
        if ThenBranch == nil {
            return(nil);
        }

        thenDone++;
        processingDone++;
        NvBotDist = MAXSHORTINT;
    } else {
        /* Derive regular child for table lookup. */
        // inlining of macro Cudd_Regular
        regNv = ((DdNode *)((ptruint)(Nv) & ~01));
        /* Get node data for shortest path length. */
        if ! st__lookup(pathTable, (char *)regNv, (char **)&nodeStatNv) {
            (void) fmt.Fprintf(dd.err, "Something wrong, node must be in table\n");
            dd.errorCode = CUDD_INTERNAL_ERROR;
            return(nil);
        }
        /* Derive shortest path length for child. */
        if (nodeStatNv.oddTopDist != MAXSHORTINT) &&
            (nodeStatNv.oddBotDist != MAXSHORTINT) {
            oddLen = (nodeStatNv.oddTopDist + nodeStatNv.oddBotDist);
        } else {
            oddLen = MAXSHORTINT;
        }

        if (nodeStatNv.evenTopDist != MAXSHORTINT) &&
            (nodeStatNv.evenBotDist != MAXSHORTINT) {
            evenLen = (nodeStatNv.evenTopDist +nodeStatNv.evenBotDist);
        } else {
            evenLen = MAXSHORTINT;
        }

        if oddLen <= evenLen {
            NvPathLength = oddLen;
        } else {
            NvPathLength = evenLen;
        }
        if oddLen <= evenLen {
            NvBotDist = nodeStatNv.oddBotDist;
        } else {
            NvBotDist = nodeStatNv.evenBotDist;
        }
    }
    /* if else child constant, branch is the child */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(Nnv) & ~01))).index == CUDD_CONST_INDEX {
        /*shortest path found */
        // inlining of macro DD_ONE
        if (Nnv == ((dd).one)) && (info.findShortestPath) {
            info.findShortestPath = 0;
        }

        ElseBranch = Nnv;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(ElseBranch).ref);
        if ElseBranch == nil {
            return(nil);
        }

        elseDone++;
        processingDone++;
        NnvBotDist = MAXSHORTINT;
    } else {
        /* Derive regular child for table lookup. */
        // inlining of macro Cudd_Regular
        regNnv = ((DdNode *)((ptruint)(Nnv) & ~01));
        /* Get node data for shortest path length. */
        if ! st__lookup(pathTable, (char *)regNnv, (char **)&nodeStatNnv) {
            (void) fmt.Fprintf(dd.err, "Something wrong, node must be in table\n");
            dd.errorCode = CUDD_INTERNAL_ERROR;
            return(nil);
        }
        /* Derive shortest path length for child. */
        if (nodeStatNnv.oddTopDist != MAXSHORTINT) &&
            (nodeStatNnv.oddBotDist != MAXSHORTINT) {
            oddLen = (nodeStatNnv.oddTopDist + nodeStatNnv.oddBotDist);
        } else {
            oddLen = MAXSHORTINT;
        }

        if (nodeStatNnv.evenTopDist != MAXSHORTINT) &&
            (nodeStatNnv.evenBotDist != MAXSHORTINT) {
            evenLen = (nodeStatNnv.evenTopDist +nodeStatNnv.evenBotDist);
        } else {
            evenLen = MAXSHORTINT;
        }

        if oddLen <= evenLen {
            NnvPathLength = oddLen;
        } else {
            NnvPathLength = evenLen;
        }
        if oddLen <= evenLen {
            NnvBotDist = nodeStatNnv.oddBotDist;
        } else {
            NnvBotDist = nodeStatNnv.evenBotDist;
        }
    }

    if NvBotDist <= NnvBotDist {
        tiebreakChild = 1;
    } else {
        tiebreakChild = 0;
    }
    /* while both children not processed */
    for processingDone != 2 {
        if !(processingDone != 0) {
            /* if no child processed */
            /* pick the child with shortest path length and record which one
             * picked
             */
            if (NvPathLength < NnvPathLength) ||
                ((NvPathLength == NnvPathLength) && (tiebreakChild == 1)) {
                child = Nv;
                regChild = regNv;
                thenDone = 1;
                childPathLength = NvPathLength;
            } else {
                child = Nnv;
                regChild = regNnv;
                elseDone = 1;
                childPathLength = NnvPathLength;
            } /* then path length less than else path length */
        } else {
            /* if one child processed, process the other */
            if thenDone {
                child = Nnv;
                regChild = regNnv;
                elseDone = 1;
                childPathLength = NnvPathLength;
            } else {
                child = Nv;
                regChild = regNv;
                thenDone = 1;
                childPathLength = NvPathLength;
            } /* end of else pick the Then child if ELSE child processed */
        } /* end of else one child has been processed */

        /* ignore (replace with constant 0) all nodes which lie on paths larger
         * than the maximum length of the path required
         */
        if childPathLength > info.maxpath {
            /* record nodes visited */
            childBranch = zero;
        } else {
            if childPathLength < info.maxpath {
                if info.findShortestPath {
                    info.findShortestPath = 0;
                }
                childBranch = BuildSubsetBdd(dd, pathTable, child, info,
                                             subsetNodeTable);

            } else { /* Case: path length of node = maxpath */
                /* If the node labeled with maxpath is found in the
                ** maxpathTable, use it to build the subset BDD.  */
                if st__lookup(info.maxpathTable, (char *)regChild,
                              (char **)&entry) {
                    /* When a node that is already been chosen is hit,
                    ** the quest for a complete path is over.  */
                    if info.findShortestPath {
                        info.findShortestPath = 0;
                    }
                    childBranch = BuildSubsetBdd(dd, pathTable, child, info,
                                                 subsetNodeTable);
                } else {
                    /* If node is not found in the maxpathTable and
                    ** the threshold has been reached, then if the
                    ** path needs to be completed, continue. Else
                    ** replace the node with a zero.  */
                    if info.thresholdReached <= 0 {
                        if info.findShortestPath {
                            if st__insert(info.maxpathTable, (char *)regChild,
                                          (char *)NIL(char)) == st__OUT_OF_MEM {
                                memOut = 1;
                                (void) fmt.Fprintf(dd.err, "OUT of memory\n");
                                info.thresholdReached = 0;
                                childBranch = zero;
                            } else {
                                info.thresholdReached--;
                                childBranch = BuildSubsetBdd(dd, pathTable,
                                                    child, info,subsetNodeTable);
                            }
                        } else { /* not find shortest path, we dont need this
                                    node */
                            childBranch = zero;
                        }
                    } else { /* Threshold hasn't been reached,
                             ** need the node. */
                        if st__insert(info.maxpathTable, (char *)regChild,
                                      (char *)NIL(char)) == st__OUT_OF_MEM {
                            memOut = 1;
                            (void) fmt.Fprintf(dd.err, "OUT of memory\n");
                            info.thresholdReached = 0;
                            childBranch = zero;
                        } else {
                            info.thresholdReached--;
                            if info.thresholdReached <= 0 {
                                info.findShortestPath = 1;
                            }
                            childBranch = BuildSubsetBdd(dd, pathTable,
                                                 child, info, subsetNodeTable);

                        } /* end of st__insert successful */
                    } /* end of threshold hasnt been reached yet */
                } /* end of else node not found in maxpath table */
            } /* end of if (path length of node = maxpath) */
        } /* end if !(childPathLength > maxpath) */
        if childBranch == nil {
            /* deref other stuff incase reordering has taken place */
            if ThenBranch != nil {
                Cudd_RecursiveDeref(dd, ThenBranch);
                ThenBranch = nil;
            }
            if ElseBranch != nil {
                Cudd_RecursiveDeref(dd, ElseBranch);
                ElseBranch = nil;
            }
            return(nil);
        }

        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(childBranch).ref);

        if child == Nv {
            ThenBranch = childBranch;
        } else {
            ElseBranch = childBranch;
        }
        processingDone++;

    } /*end of while processing Nv, Nnv */

    info.findShortestPath = 0;
    topid = Cudd_NodeReadIndex(N);
    topv = Cudd_ReadVars(dd, topid);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(topv).ref);
    neW = cuddBddIteRecur(dd, topv, ThenBranch, ElseBranch);
    if neW != nil {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(neW).ref);
    }
    Cudd_RecursiveDeref(dd, topv);
    Cudd_RecursiveDeref(dd, ThenBranch);
    Cudd_RecursiveDeref(dd, ElseBranch);


    /* Hard Limit of threshold has been imposed */
    if subsetNodeTable != NIL( st__table) {
        /* check if a new node is created */
        // inlining of macro Cudd_Regular
        regNew = ((DdNode *)((ptruint)(neW) & ~01));
        /* subset node table keeps all new nodes that have been created to keep
         * a running count of how many nodes have been built in the subset.
         */
        if ! st__lookup(subsetNodeTable, (char *)regNew, (char **)&entry) {
            // inlining of macro Cudd_IsConstant
            // inlining of macro Cudd_Regular
            if !((((DdNode *)((ptruint)(regNew) & ~01))).index == CUDD_CONST_INDEX) {
                if st__insert(subsetNodeTable, (char *)regNew,
                              (char *)nil) == st__OUT_OF_MEM {
                    (void) fmt.Fprintf(dd.err, "Out of memory\n");
                    return (nil);
                }
                // inlining of macro st__count
                if  ((subsetNodeTable).num_entries) > info.threshold {
                    info.thresholdReached = 0;
                }
            }
        }
    }


    if neW == nil {
        return(nil);
    } else {
        /*store computed result in regular form*/
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (node) & 01)) != 0 {
            nodeStat.compResult = neW;
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(nodeStat.compResult).ref);
            /* if the new node is the same as the corresponding node in the
             * original bdd then its complement need not be computed as it
             * cannot be larger than the node itself
             */
            if neW == node {
#ifdef DD_DEBUG
                thishit++;
#endif
                /* if a result for the node has already been computed, then
                 * it can only be smaller than teh node itself. hence store
                 * the node result in order not to break recombination
                 */
                if nodeStat.regResult != nil {
                    Cudd_RecursiveDeref(dd, nodeStat.regResult);
                }
                // inlining of macro Cudd_Not
                nodeStat.regResult = ((DdNode *)((ptrint)(neW) ^ 01));
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(nodeStat.regResult).ref);
            }

        } else {
            nodeStat.regResult = neW;
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(nodeStat.regResult).ref);
            if neW == node {
#ifdef DD_DEBUG
                thishit++;
#endif
                if nodeStat.compResult != nil {
                    Cudd_RecursiveDeref(dd, nodeStat.compResult);
                }
                // inlining of macro Cudd_Not
                nodeStat.compResult = ((DdNode *)((ptrint)(neW) ^ 01));
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(nodeStat.compResult).ref);
            }
        }

        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(neW).ref);
        return(neW);
    } /* end of else i.e. Subset != NULL */
} /* end of BuildSubsetBdd */


/**Function********************************************************************

  Synopsis     [Procedure to free te result dds stored in the NodeDist pages.]

  Description [None]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func stPathTableDdFree(key string, value string, arg string) st__retval {
    var nodeStat *NodeDist_t;
    var dd *DdManager;

    nodeStat = (NodeDist_t *)value;
    dd = (DdManager *)arg;
    if nodeStat.regResult != nil {
        Cudd_RecursiveDeref(dd, nodeStat.regResult);
    }
    if nodeStat.compResult != nil {
        Cudd_RecursiveDeref(dd, nodeStat.compResult);
    }
    return( st__CONTINUE);

} /* end of stPathTableFree */


//ABC_NAMESPACE_IMPL_END

