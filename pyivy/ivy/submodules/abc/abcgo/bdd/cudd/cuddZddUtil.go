package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddZddUtil.c]

  PackageName [cudd]

  Synopsis    [Utility functions for ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddPrintMinterm()
                    <li> Cudd_zddPrintCover()
                    <li> Cudd_zddPrintDebug()
                    <li> Cudd_zddFirstPath()
                    <li> Cudd_zddNextPath()
                    <li> Cudd_zddCoverPathToString()
                    <li> Cudd_zddDumpDot()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddZddP()
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> zp2()
                    <li> zdd_print_minterm_aux()
                    <li> zddPrintCoverAux()
                    </ul>
              ]

  SeeAlso     []

  Author      [Hyong-Kyoon Shin, In-Ho Moon, Fabio Somenzi]

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
static char rcsid[] DD_UNUSED = "$Id: cuddZddUtil.c,v 1.27 2009/03/08 02:49:02 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int zp2 (DdManager *zdd, DdNode *f, st__table *t);

// static void zdd_print_minterm_aux (DdManager *zdd, DdNode *node, int level, int *list);

// static void zddPrintCoverAux (DdManager *zdd, DdNode *node, int level, int *list);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Prints a disjoint sum of product form for a ZDD.]

  Description [Prints a disjoint sum of product form for a ZDD. Returns 1
  if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddPrintDebug Cudd_zddPrintCover]

******************************************************************************/
func Cudd_zddPrintMinterm(zdd *DdManager, node *DdNode) int32 {
    var i, size int32;
    var list *int32;

    size = int32(zdd.sizeZ);
    list = ABC_ALLOC(int, size);
    if list == nil {
        zdd.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    for i = 0; i < size; i++ { list[i] = 3; } /* bogus value should disappear */
    zdd_print_minterm_aux(zdd, node, 0, list);
    ABC_FREE(list);
    return(1);

} /* end of Cudd_zddPrintMinterm */


/**Function********************************************************************

  Synopsis    [Prints a sum of products from a ZDD representing a cover.]

  Description [Prints a sum of products from a ZDD representing a cover.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddPrintMinterm]

******************************************************************************/
func Cudd_zddPrintCover(zdd *DdManager, node *DdNode) int32 {
    var i, size int32;
    var list *int32;

    size = int32(zdd.sizeZ);
    if size % 2 != 0 { return(0); } /* number of variables should be even */
    list = ABC_ALLOC(int, size);
    if list == nil {
        zdd.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    for i = 0; i < size; i++ { list[i] = 3; } /* bogus value should disappear */
    zddPrintCoverAux(zdd, node, 0, list);
    ABC_FREE(list);
    return(1);

} /* end of Cudd_zddPrintCover */


/**Function********************************************************************

  Synopsis [Prints to the standard output a ZDD and its statistics.]

  Description [Prints to the standard output a DD and its statistics.
  The statistics include the number of nodes and the number of minterms.
  (The number of minterms is also the number of combinations in the set.)
  The statistics are printed if pr &gt; 0.  Specifically:
  <ul>
  <li> pr = 0 : prints nothing
  <li> pr = 1 : prints counts of nodes and minterms
  <li> pr = 2 : prints counts + disjoint sum of products
  <li> pr = 3 : prints counts + list of nodes
  <li> pr &gt; 3 : prints counts + disjoint sum of products + list of nodes
  </ul>
  Returns 1 if successful; 0 otherwise.
  ]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_zddPrintDebug(zdd *DdManager, f *DdNode, n int32, pr int32) int32 {
    // inlining of macro DD_ZERO
    var empty *DdNode  = ((zdd).zero);
    var nodes int32;
    var minterms float64;
    var retval int32  = 1;

    if f == empty && pr > 0 {
        (void) fmt.Fprintf(zdd.out, ": is the empty ZDD\n");
        (void) fflush(zdd.out);
        return(1);
    }

    if pr > 0 {
        nodes = Cudd_zddDagSize(f);
        if nodes == CUDD_OUT_OF_MEM { retval = 0; }
        minterms = Cudd_zddCountMinterm(zdd, f, n);
        if minterms == float64(CUDD_OUT_OF_MEM) { retval = 0; }
        (void) fmt.Fprintf(zdd.out, ": %d nodes %g minterms\n", nodes, minterms);
        if pr > 2 {
            if !(cuddZddP(zdd, f) != 0) { retval = 0; } }
        if pr == 2 || pr > 3 {
            if !(Cudd_zddPrintMinterm(zdd, f) != 0) { retval = 0; }
            (void) fmt.Fprintf(zdd.out, "\n");
        }
        (void) fflush(zdd.out);
    }
    return(retval);

} /* end of Cudd_zddPrintDebug */



/**Function********************************************************************

  Synopsis    [Finds the first path of a ZDD.]

  Description [Defines an iterator on the paths of a ZDD
  and finds its first path. Returns a generator that contains the
  information necessary to continue the enumeration if successful; NULL
  otherwise.<p>
  A path is represented as an array of literals, which are integers in
  {0, 1, 2}; 0 represents an else arc out of a node, 1 represents a then arc
  out of a node, and 2 stands for the absence of a node.
  The size of the array equals the number of variables in the manager at
  the time Cudd_zddFirstCube is called.<p>
  The paths that end in the empty terminal are not enumerated.]

  SideEffects [The first path is returned as a side effect.]

  SeeAlso     [Cudd_zddForeachPath Cudd_zddNextPath Cudd_GenFree
  Cudd_IsGenEmpty]

******************************************************************************/
func Cudd_zddFirstPath(zdd *DdManager, f *DdNode, path **int32) *DdGen {
    var gen *DdGen;
    var top, next, prev *DdNode;
    var i int32;
    var nvars int32;

    /* Sanity Check. */
    if zdd == nil || f == nil { return(nil); }

    /* Allocate generator an initialize it. */
    gen = new(DdGen);
    if gen == nil {
        zdd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }

    gen.manager = zdd;
    gen.type = CUDD_GEN_ZDD_PATHS;
    gen.status = CUDD_GEN_EMPTY;
    gen.gen.cubes.cube = nil;
    gen.gen.cubes.value = DD_ZERO_VAL;
    gen.stack.sp = 0;
    gen.stack.stack = nil;
    gen.node = nil;

    nvars = zdd.sizeZ;
    gen.gen.cubes.cube = ABC_ALLOC(int,nvars);
    if gen.gen.cubes.cube == nil {
        zdd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(gen);
        return(nil);
    }
    for i = 0; i < nvars; i++ { gen.gen.cubes.cube[i] = 2; }

    /* The maximum stack depth is one plus the number of variables.
    ** because a path may have nodes at all levels, including the
    ** constant level.
    */
    gen.stack.stack = ABC_ALLOC(DdNodePtr, nvars+1);
    if gen.stack.stack == nil {
        zdd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(gen.gen.cubes.cube);
        ABC_FREE(gen);
        return(nil);
    }
    for i = 0; i <= nvars; i++ { gen.stack.stack[i] = nil; }

    /* Find the first path of the ZDD. */
    gen.stack.stack[gen.stack.sp] = f; gen.stack.sp++;

    for {
        top = gen.stack.stack[gen.stack.sp-1];
        // inlining of macro cuddIsConstant
        // inlining of macro Cudd_Regular
        if !((((DdNode *)((ptruint)(top) & ~01))).index == CUDD_CONST_INDEX) {
            /* Take the else branch first. */
            // inlining of macro Cudd_Regular
            gen.gen.cubes.cube[((DdNode *)((ptruint)(top) & ~01)).index] = 0;
            // inlining of macro cuddE
            // inlining of macro Cudd_Regular
            next = ((((DdNode *)((ptruint)(top) & ~01))).type.kids.E);
            // inlining of macro Cudd_Not
            gen.stack.stack[gen.stack.sp] = ((DdNode *)((ptrint)(next) ^ 01)); gen.stack.sp++;
        // inlining of macro Cudd_Regular
        // inlining of macro DD_ZERO
        } else if ((DdNode *)((ptruint)(top) & ~01)) == ((zdd).zero) {
            /* Backtrack. */
            for {
                if gen.stack.sp == 1 {
                    /* The current node has no predecessor. */
                    gen.status = CUDD_GEN_EMPTY;
                    gen.stack.sp--;
                    goto done;
                }
                // inlining of macro Cudd_Regular
                prev = ((DdNode *)((ptruint)(gen.stack.stack[gen.stack.sp-2]) & ~01));
                // inlining of macro cuddT
                next = ((prev).type.kids.T);
                if next != top { /* follow the then branch next */
                    gen.gen.cubes.cube[prev.index] = 1;
                    gen.stack.stack[gen.stack.sp-1] = next;
                    break;
                }
                /* Pop the stack and try again. */
                gen.gen.cubes.cube[prev.index] = 2;
                gen.stack.sp--;
                top = gen.stack.stack[gen.stack.sp-1];
            }
        } else {
            gen.status = CUDD_GEN_NONEMPTY;
            // inlining of macro cuddV
            // inlining of macro Cudd_Regular
            gen.gen.cubes.value = ((((DdNode *)((ptruint)(top) & ~01))).type.value);
            goto done;
        }
    }

done:
    *path = gen.gen.cubes.cube;
    return(gen);

} /* end of Cudd_zddFirstPath */


/**Function********************************************************************

  Synopsis    [Generates the next path of a ZDD.]

  Description [Generates the next path of a ZDD onset,
  using generator gen. Returns 0 if the enumeration is completed; 1
  otherwise.]

  SideEffects [The path is returned as a side effect. The
  generator is modified.]

  SeeAlso     [Cudd_zddForeachPath Cudd_zddFirstPath Cudd_GenFree
  Cudd_IsGenEmpty]

******************************************************************************/
func Cudd_zddNextPath(gen *DdGen, path **int32) int32 {
    var top, next, prev *DdNode;
    var zdd *DdManager  = gen.manager;

    /* Backtrack from previously reached terminal node. */
    for {
        if gen.stack.sp == 1 {
            /* The current node has no predecessor. */
            gen.status = CUDD_GEN_EMPTY;
            gen.stack.sp--;
            goto done;
        }
        top = gen.stack.stack[gen.stack.sp-1];
        // inlining of macro Cudd_Regular
        prev = ((DdNode *)((ptruint)(gen.stack.stack[gen.stack.sp-2]) & ~01));
        // inlining of macro cuddT
        next = ((prev).type.kids.T);
        if next != top { /* follow the then branch next */
            gen.gen.cubes.cube[prev.index] = 1;
            gen.stack.stack[gen.stack.sp-1] = next;
            break;
        }
        /* Pop the stack and try again. */
        gen.gen.cubes.cube[prev.index] = 2;
        gen.stack.sp--;
    }

    for {
        top = gen.stack.stack[gen.stack.sp-1];
        // inlining of macro cuddIsConstant
        // inlining of macro Cudd_Regular
        if !((((DdNode *)((ptruint)(top) & ~01))).index == CUDD_CONST_INDEX) {
            /* Take the else branch first. */
            // inlining of macro Cudd_Regular
            gen.gen.cubes.cube[((DdNode *)((ptruint)(top) & ~01)).index] = 0;
            // inlining of macro cuddE
            // inlining of macro Cudd_Regular
            next = ((((DdNode *)((ptruint)(top) & ~01))).type.kids.E);
            // inlining of macro Cudd_Not
            gen.stack.stack[gen.stack.sp] = ((DdNode *)((ptrint)(next) ^ 01)); gen.stack.sp++;
        // inlining of macro Cudd_Regular
        // inlining of macro DD_ZERO
        } else if ((DdNode *)((ptruint)(top) & ~01)) == ((zdd).zero) {
            /* Backtrack. */
            for {
                if gen.stack.sp == 1 {
                    /* The current node has no predecessor. */
                    gen.status = CUDD_GEN_EMPTY;
                    gen.stack.sp--;
                    goto done;
                }
                // inlining of macro Cudd_Regular
                prev = ((DdNode *)((ptruint)(gen.stack.stack[gen.stack.sp-2]) & ~01));
                // inlining of macro cuddT
                next = ((prev).type.kids.T);
                if next != top { /* follow the then branch next */
                    gen.gen.cubes.cube[prev.index] = 1;
                    gen.stack.stack[gen.stack.sp-1] = next;
                    break;
                }
                /* Pop the stack and try again. */
                gen.gen.cubes.cube[prev.index] = 2;
                gen.stack.sp--;
                top = gen.stack.stack[gen.stack.sp-1];
            }
        } else {
            gen.status = CUDD_GEN_NONEMPTY;
            // inlining of macro cuddV
            // inlining of macro Cudd_Regular
            gen.gen.cubes.value = ((((DdNode *)((ptruint)(top) & ~01))).type.value);
            goto done;
        }
    }

done:
    if gen.status == CUDD_GEN_EMPTY { return(0); }
    *path = gen.gen.cubes.cube;
    return(1);

} /* end of Cudd_zddNextPath */


/**Function********************************************************************

  Synopsis    [Converts a path of a ZDD representing a cover to a string.]

  Description [Converts a path of a ZDD representing a cover to a
  string.  The string represents an implicant of the cover.  The path
  is typically produced by Cudd_zddForeachPath.  Returns a pointer to
  the string if successful; NULL otherwise.  If the str input is NULL,
  it allocates a new string.  The string passed to this function must
  have enough room for all variables and for the terminator.]

  SideEffects [None]

  SeeAlso     [Cudd_zddForeachPath]

******************************************************************************/
char *
Cudd_zddCoverPathToString(
  DdManager *zdd                /* DD manager */,
  int *path                     /* path of ZDD representing a cover */,
  char *str                     /* pointer to string to use if != NULL */
  )
{
    var nvars int32  = zdd.sizeZ;
    var i int32;
    var res string;

    if nvars & 1 { return(nil); }
    nvars >>= 1;
    if str == nil {
        res = ABC_ALLOC(char, nvars+1);
        if res == nil { return(nil); }
    } else {
        res = str;
    }
    for i = 0; i < nvars; i++ {
        var v int32  = (path[2*i] << 2) | path[2*i+1];
        switch v {
        case 0,
         2,
         8,
         10:
            res[i] = '-';
            break;
        case 1,
         9:
            res[i] = '0';
            break;
        case 4,
         6:
            res[i] = '1';
            break;
        default:
            res[i] = '?';
        }
    }
    res[nvars] = 0;

    return(res);

} /* end of Cudd_zddCoverPathToString */


/**Function********************************************************************

  Synopsis    [Writes a dot file representing the argument ZDDs.]

  Description [Writes a file representing the argument ZDDs in a format
  suitable for the graph drawing program dot.
  It returns 1 in case of success; 0 otherwise (e.g., out-of-memory,
  file system full).
  Cudd_zddDumpDot does not close the file: This is the caller
  responsibility. Cudd_zddDumpDot uses a minimal unique subset of the
  hexadecimal address of a node as name for it.
  If the argument inames is non-null, it is assumed to hold the pointers
  to the names of the inputs. Similarly for onames.
  Cudd_zddDumpDot uses the following convention to draw arcs:
    <ul>
    <li> solid line: THEN arcs;
    <li> dashed line: ELSE arcs.
    </ul>
  The dot options are chosen so that the drawing fits on a letter-size
  sheet.
  ]

  SideEffects [None]

  SeeAlso     [Cudd_DumpDot Cudd_zddPrintDebug]

******************************************************************************/
int
Cudd_zddDumpDot(
  DdManager * dd /* manager */,
  int  n /* number of output nodes to be dumped */,
  DdNode ** f /* array of output nodes to be dumped */,
  char ** inames /* array of input names (or NULL) */,
  char ** onames /* array of output names (or NULL) */,
  FILE * fp /* pointer to the dump file */)
{
    var support *DdNode  = nil;
    var scan *DdNode;
    var sorted *int32  = nil;
    var nvars int32  = dd.sizeZ;
    var visited *st__table  = nil;
    var gen *st__generator;
    var retval int32;
    var i, j int32;
    var slots int32;
    var nodelist *DdNodePtr;
    var refAddr, diff, mask long;

    /* Build a bit array with the support of f. */
    sorted = ABC_ALLOC(int,nvars);
    if sorted == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        goto failure;
    }
    for i = 0; i < nvars; i++ { sorted[i] = 0; }

    /* Take the union of the supports of each output function. */
    for i = 0; i < n; i++ {
        support = Cudd_Support(dd,f[i]);
        if support == nil { goto failure; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(support).ref);
        scan = support;
        // inlining of macro cuddIsConstant
        for !((scan).index == CUDD_CONST_INDEX) {
            sorted[scan.index] = 1;
            // inlining of macro cuddT
            scan = ((scan).type.kids.T);
        }
        Cudd_RecursiveDeref(dd,support);
    }
    support = nil; /* so that we do not try to free it in case of failure */

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);
    if visited == nil { goto failure; }

    /* Collect all the nodes of this DD in the symbol table. */
    for i = 0; i < n; i++ {
        retval = cuddCollectNodes(f[i],visited);
        if retval == 0 { goto failure; }
    }

    /* Find how many most significant hex digits are identical
    ** in the addresses of all the nodes. Build a mask based
    ** on this knowledge, so that digits that carry no information
    ** will not be printed. This is done in two steps.
    **  1. We scan the symbol table to find the bits that differ
    **     in at least 2 addresses.
    **  2. We choose one of the possible masks. There are 8 possible
    **     masks for 32-bit integer, and 16 possible masks for 64-bit
    **     integers.
    */

    /* Find the bits that are different. */
    refAddr = int64( f[0]);
    diff = 0;
    gen = st__init_gen(visited);
    for  st__gen(gen, (char **)&scan, nil) {
        diff |= refAddr ^ int64( scan);
    }
    st__free_gen(gen);

    /* Choose the mask. */
    for i = 0; uint32( i) < 8 * sizeof(long); i += 4 {
        mask = (1 << i) - 1;
        if diff <= mask { break; }
    }

    /* Write the header and the global attributes. */
    retval = fmt.Fprintf(fp, "digraph \"ZDD\" {\n");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "size = \"7.5,10\"\ncenter = true;\nedge [dir = none];\n");
    if retval == -1 { return(0); }

    /* Write the input name subgraph by scanning the support array. */
    retval = fmt.Fprintf(fp, "{ node [shape = plaintext];\n");
    if retval == -1 { goto failure; }
    retval = fmt.Fprintf(fp, "  edge [style = invis];\n");
    if retval == -1 { goto failure; }
    /* We use a name ("CONST NODES") with an embedded blank, because
    ** it is unlikely to appear as an input name.
    */
    retval = fmt.Fprintf(fp, "  \"CONST NODES\" [style = invis];\n");
    if retval == -1 { goto failure; }
    for i = 0; i < nvars; i++ {
        if sorted[dd.invpermZ[i]] {
            if inames == nil {
                retval = fmt.Fprintf(fp, "\" %d \" -> ", dd.invpermZ[i]);
            } else {
                retval = fmt.Fprintf(fp, "\" %s \" -> ", inames[dd.invpermZ[i]]);
            }
            if retval == -1 { goto failure; }
        }
    }
    retval = fmt.Fprintf(fp, "\"CONST NODES\"; \n}\n");
    if retval == -1 { goto failure; }

    /* Write the output node subgraph. */
    retval = fmt.Fprintf(fp, "{ rank = same; node [shape = box]; edge [style = invis];\n");
    if retval == -1 { goto failure; }
    for i = 0; i < n; i++ {
        if onames == nil {
            retval = fmt.Fprintf(fp, "\"F%d\"", i);
        } else {
            retval = fmt.Fprintf(fp, "\"  %s  \"", onames[i]);
        }
        if retval == -1 { goto failure; }
        if i == n - 1 {
            retval = fmt.Fprintf(fp, "; }\n");
        } else {
            retval = fmt.Fprintf(fp, " -> ");
        }
        if retval == -1 { goto failure; }
    }

    /* Write rank info: All nodes with the same index have the same rank. */
    for i = 0; i < nvars; i++ {
        if sorted[dd.invpermZ[i]] {
            retval = fmt.Fprintf(fp, "{ rank = same; ");
            if retval == -1 { goto failure; }
            if inames == nil {
                retval = fmt.Fprintf(fp, "\" %d \";\n", dd.invpermZ[i]);
            } else {
                retval = fmt.Fprintf(fp, "\" %s \";\n", inames[dd.invpermZ[i]]);
            }
            if retval == -1 { goto failure; }
            nodelist = dd.subtableZ[i].nodelist;
            slots = dd.subtableZ[i].slots;
            for j = 0; j < slots; j++ {
                scan = nodelist[j];
                for scan != nil {
                    // inlining of macro st__is_member
                    if  st__lookup(visited,(char *) scan,(char **) 0) {
                        retval = fmt.Fprintf(fp, "\"%p\";\n", (void *)
                                                                 ((mask & (ptrint) scan) /
                                                                  sizeof(DdNode)));
                        if retval == -1 { goto failure; }
                    }
                    scan = scan.next;
                }
            }
            retval = fmt.Fprintf(fp, "}\n");
            if retval == -1 { goto failure; }
        }
    }

    /* All constants have the same rank. */
    retval = fmt.Fprintf(fp, "{ rank = same; \"CONST NODES\";\n{ node [shape = box]; ");
    if retval == -1 { goto failure; }
    nodelist = dd.constants.nodelist;
    slots = dd.constants.slots;
    for j = 0; j < slots; j++ {
        scan = nodelist[j];
        for scan != nil {
            // inlining of macro st__is_member
            if  st__lookup(visited,(char *) scan,(char **) 0) {
                retval = fmt.Fprintf(fp, "\"%p\";\n", (void *)
                                                 ((mask & (ptrint) scan) / sizeof(DdNode)));
                if retval == -1 { goto failure; }
            }
            scan = scan.next;
        }
    }
    retval = fmt.Fprintf(fp, "}\n}\n");
    if retval == -1 { goto failure; }

    /* Write edge info. */
    /* Edges from the output nodes. */
    for i = 0; i < n; i++ {
        if onames == nil {
            retval = fmt.Fprintf(fp, "\"F%d\"", i);
        } else {
            retval = fmt.Fprintf(fp, "\"  %s  \"", onames[i]);
        }
        if retval == -1 { goto failure; }
        retval = fmt.Fprintf(fp, " -> \"%p\" [style = solid];\n", (void *) ((mask & (ptrint) f[i]) /
                                                  sizeof(DdNode)));
        if retval == -1 { goto failure; }
    }

    /* Edges from internal nodes. */
    for i = 0; i < nvars; i++ {
        if sorted[dd.invpermZ[i]] {
            nodelist = dd.subtableZ[i].nodelist;
            slots = dd.subtableZ[i].slots;
            for j = 0; j < slots; j++ {
                scan = nodelist[j];
                for scan != nil {
                    // inlining of macro st__is_member
                    if  st__lookup(visited,(char *) scan,(char **) 0) {
                        // inlining of macro cuddT
                        retval = fmt.Fprintf(fp, "\"%p\" -> \"%p\";\n", (void *) ((mask & (ptrint) scan) / sizeof(DdNode)), (void *) ((mask & (ptrint) ((scan).type.kids.T)) /
                                                              sizeof(DdNode)));
                        if retval == -1 { goto failure; }
                        retval = fmt.Fprintf(fp, "\"%p\" -> \"%p\" [style = dashed];\n", (void *) ((mask & (ptrint) scan)
                                                                           / sizeof(DdNode)), (void *) ((mask & (ptrint)
                                                                            // inlining of macro cuddE
                                                                            ((scan).type.kids.E)) /
                                                                           sizeof(DdNode)));
                        if retval == -1 { goto failure; }
                    }
                    scan = scan.next;
                }
            }
        }
    }

    /* Write constant labels. */
    nodelist = dd.constants.nodelist;
    slots = dd.constants.slots;
    for j = 0; j < slots; j++ {
        scan = nodelist[j];
        for scan != nil {
            // inlining of macro st__is_member
            if  st__lookup(visited,(char *) scan,(char **) 0) {
                retval = fmt.Fprintf(fp, "\"%p\" [label = \"%g\"];\n", (void *) ((mask & (ptrint) scan) /
                                                           // inlining of macro cuddV
                                                           sizeof(DdNode)), ((scan).type.value));
                if retval == -1 { goto failure; }
            }
            scan = scan.next;
        }
    }

    /* Write trailer and return. */
    retval = fmt.Fprintf(fp, "}\n");
    if retval == -1 { goto failure; }

    st__free_table(visited);
    ABC_FREE(sorted);
    return(1);

failure:
    if sorted != nil { ABC_FREE(sorted); }
    if visited != nil { st__free_table(visited); }
    return(0);

} /* end of Cudd_zddDumpBlif */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Prints a ZDD to the standard output. One line per node is
  printed.]

  Description [Prints a ZDD to the standard output. One line per node is
  printed. Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddPrintDebug]

******************************************************************************/
func cuddZddP(zdd *DdManager, f *DdNode) int32 {
    var retval int32;
    var table *st__table  = st__init_table( st__ptrcmp, st__ptrhash);

    if table == nil { return(0); }

    retval = zp2(zdd, f, table);
    st__free_table(table);
    (void) fputc('\n', zdd.out);
    return(retval);

} /* end of cuddZddP */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Performs the recursive step of cuddZddP.]

  Description [Performs the recursive step of cuddZddP. Returns 1 in
  case of success; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func zp2(zdd *DdManager, f *DdNode, t *st__table) int32 {
    var n *DdNode;
    var T, E int32;
    // inlining of macro DD_ONE
    var base *DdNode  = ((zdd).one);

    if f == nil {
        return(0); }

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        (void)fmt.Fprintf(zdd.out, "ID = %d\n", (f == base));
        return(1);
    }
    // inlining of macro st__is_member
    if  st__lookup(t,(char *)f,(char **) 0) == 1 {
        return(1); }

    if st__insert(t, (char *) f, nil) == st__OUT_OF_MEM {
        return(0); }

#if SIZEOF_VOID_P == 8
    (void) fprintf(zdd->out, "ID = 0x%lx\tindex = %u\tr = %u\t",
        (ptruint)f / (ptruint) sizeof(DdNode), f->index, f->ref);
#else
    (void) fprintf(zdd->out, "ID = 0x%x\tindex = %hu\tr = %hu\t",
        (ptruint)f / (ptruint) sizeof(DdNode), f->index, f->ref);
#endif

    // inlining of macro cuddT
    n = ((f).type.kids.T);
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(n) & ~01))).index == CUDD_CONST_INDEX {
        (void) fmt.Fprintf(zdd.out, "T = %d\t\t", (n == base));
        T = 1;
    } else {
#if SIZEOF_VOID_P == 8
        (void) fprintf(zdd->out, "T = 0x%lx\t", (ptruint) n /
                       (ptruint) sizeof(DdNode));
#else
        (void) fprintf(zdd->out, "T = 0x%x\t", (ptruint) n /
                       (ptruint) sizeof(DdNode));
#endif
        T = 0;
    }

    // inlining of macro cuddE
    n = ((f).type.kids.E);
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(n) & ~01))).index == CUDD_CONST_INDEX {
        (void) fmt.Fprintf(zdd.out, "E = %d\n", (n == base));
        E = 1;
    } else {
#if SIZEOF_VOID_P == 8
        (void) fprintf(zdd->out, "E = 0x%lx\n", (ptruint) n /
                      (ptruint) sizeof(DdNode));
#else
        (void) fprintf(zdd->out, "E = 0x%x\n", (ptruint) n /
                       (ptruint) sizeof(DdNode));
#endif
        E = 0;
    }

    if E == 0 {
        // inlining of macro cuddE
        if zp2(zdd, ((f).type.kids.E), t) == 0 { return(0); } }
    if T == 0 {
        // inlining of macro cuddT
        if zp2(zdd, ((f).type.kids.T), t) == 0 { return(0); } }
    return(1);

} /* end of zp2 */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_zddPrintMinterm.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
static void
zdd_print_minterm_aux(
  DdManager * zdd /* manager */,
  DdNode * node /* current node */,
  int  level /* depth in the recursion */,
  int * list /* current recursion path */)
{
    var Nv, Nnv *DdNode;
    var i, v int32;
    // inlining of macro DD_ONE
    var base *DdNode  = ((zdd).one);

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(node) & ~01))).index == CUDD_CONST_INDEX {
        if node == base {
            /* Check for missing variable. */
            if level != zdd.sizeZ {
                list[zdd.invpermZ[level]] = 0;
                zdd_print_minterm_aux(zdd, node, level + 1, list);
                return;
            }
            /* Terminal case: Print one cube based on the current recursion
            ** path.
            */
            for i = 0; i < zdd.sizeZ; i++ {
                v = list[i];
                if v == 0 {
                    (void) fmt.Fprintf(zdd.out, "0");
                } else if v == 1 {
                    (void) fmt.Fprintf(zdd.out, "1");
                } else if v == 3 {
                    (void) fmt.Fprintf(zdd.out, "@");       /* should never happen */
                } else {
                    (void) fmt.Fprintf(zdd.out, "-"); }
            }
            (void) fmt.Fprintf(zdd.out, " 1\n");
        }
    } else {
        /* Check for missing variable. */
        // inlining of macro cuddIZ
        if level != (((node.index)==CUDD_CONST_INDEX)?int32((node.index)):(zdd).permZ[(node.index)]) {
            list[zdd.invpermZ[level]] = 0;
            zdd_print_minterm_aux(zdd, node, level + 1, list);
            return;
        }

        // inlining of macro cuddE
        Nnv = ((node).type.kids.E);
        // inlining of macro cuddT
        Nv = ((node).type.kids.T);
        if Nv == Nnv {
            list[node.index] = 2;
            zdd_print_minterm_aux(zdd, Nnv, level + 1, list);
            return;
        }

        list[node.index] = 1;
        zdd_print_minterm_aux(zdd, Nv, level + 1, list);
        list[node.index] = 0;
        zdd_print_minterm_aux(zdd, Nnv, level + 1, list);
    }
    return;

} /* end of zdd_print_minterm_aux */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_zddPrintCover.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
static void
zddPrintCoverAux(
  DdManager * zdd /* manager */,
  DdNode * node /* current node */,
  int  level /* depth in the recursion */,
  int * list /* current recursion path */)
{
    var Nv, Nnv *DdNode;
    var i, v int32;
    // inlining of macro DD_ONE
    var base *DdNode  = ((zdd).one);

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(node) & ~01))).index == CUDD_CONST_INDEX {
        if node == base {
            /* Check for missing variable. */
            if level != zdd.sizeZ {
                list[zdd.invpermZ[level]] = 0;
                zddPrintCoverAux(zdd, node, level + 1, list);
                return;
            }
            /* Terminal case: Print one cube based on the current recursion
            ** path.
            */
            for i = 0; i < zdd.sizeZ; i += 2 {
                v = list[i] * 4 + list[i+1];
                if v == 0 {
                    (void) putc('-',zdd.out);
                } else if v == 4 {
                    (void) putc('1',zdd.out);
                } else if v == 1 {
                    (void) putc('0',zdd.out);
                } else {
                    (void) putc('@',zdd.out); } /* should never happen */
            }
            (void) fmt.Fprintf(zdd.out, " 1\n");
        }
    } else {
        /* Check for missing variable. */
        // inlining of macro cuddIZ
        if level != (((node.index)==CUDD_CONST_INDEX)?int32((node.index)):(zdd).permZ[(node.index)]) {
            list[zdd.invpermZ[level]] = 0;
            zddPrintCoverAux(zdd, node, level + 1, list);
            return;
        }

        // inlining of macro cuddE
        Nnv = ((node).type.kids.E);
        // inlining of macro cuddT
        Nv = ((node).type.kids.T);
        if Nv == Nnv {
            list[node.index] = 2;
            zddPrintCoverAux(zdd, Nnv, level + 1, list);
            return;
        }

        list[node.index] = 1;
        zddPrintCoverAux(zdd, Nv, level + 1, list);
        list[node.index] = 0;
        zddPrintCoverAux(zdd, Nnv, level + 1, list);
    }
    return;

} /* end of zddPrintCoverAux */


//ABC_NAMESPACE_IMPL_END

