package cudd

import "fmt"

import "os"

/**CFile***********************************************************************

  FileName    [cuddExport.c]

  PackageName [cudd]

  Synopsis    [Export functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_DumpBlif()
                <li> Cudd_DumpBlifBody()
                <li> Cudd_DumpDot()
                <li> Cudd_DumpDaVinci()
                <li> Cudd_DumpDDcal()
                <li> Cudd_DumpFactoredForm()
                </ul>
        Internal procedures included in this module:
                <ul>
                </ul>
        Static procedures included in this module:
                <ul>
                <li> ddDoDumpBlif()
                <li> ddDoDumpDaVinci()
                <li> ddDoDumpDDcal()
                <li> ddDoDumpFactoredForm()
                </ul>]

  Author      [Fabio Somenzi]

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
static char rcsid[] DD_UNUSED = "$Id: cuddExport.c,v 1.22 2009/03/08 02:49:02 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int ddDoDumpBlif (DdManager *dd, DdNode *f, FILE *fp, st__table *visited, char **names, int mv);

// static int ddDoDumpDaVinci (DdManager *dd, DdNode *f, FILE *fp, st__table *visited, char **names, ptruint mask);

// static int ddDoDumpDDcal (DdManager *dd, DdNode *f, FILE *fp, st__table *visited, char **names, ptruint mask);

// static int ddDoDumpFactoredForm (DdManager *dd, DdNode *f, FILE *fp, char **names);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Writes a blif file representing the argument BDDs.]

  Description [Writes a blif file representing the argument BDDs as a
  network of multiplexers. One multiplexer is written for each BDD
  node. It returns 1 in case of success; 0 otherwise (e.g.,
  out-of-memory, file system full, or an ADD with constants different
  from 0 and 1).  Cudd_DumpBlif does not close the file: This is the
  caller responsibility. Cudd_DumpBlif uses a minimal unique subset of
  the hexadecimal address of a node as name for it.  If the argument
  inames is non-null, it is assumed to hold the pointers to the names
  of the inputs. Similarly for onames.]

  SideEffects [None]

  SeeAlso     [Cudd_DumpBlifBody Cudd_DumpDot Cudd_PrintDebug Cudd_DumpDDcal
  Cudd_DumpDaVinci Cudd_DumpFactoredForm]

******************************************************************************/
int
Cudd_DumpBlif(
  DdManager * dd /* manager */,
  int  n /* number of output nodes to be dumped */,
  DdNode ** f /* array of output nodes to be dumped */,
  char ** inames /* array of input names (or NULL) */,
  char ** onames /* array of output names (or NULL) */,
  char * mname /* model name (or NULL) */,
  FILE * fp /* pointer to the dump file */,
  int mv /* 0: blif, 1: blif-MV */)
{
    var support *DdNode  = nil;
    var scan *DdNode;
    var sorted *int32  = nil;
    var nvars int32  = dd.size;
    var retval int32;
    var i int32;

    /* Build a bit array with the support of f. */
    sorted = ABC_ALLOC(int,nvars);
    if sorted == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        goto failure;
    }
    for i = 0; i < nvars; i++ { sorted[i] = 0; }

    /* Take the union of the supports of each output function. */
    support = Cudd_VectorSupport(dd,f,n);
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
    support = nil; /* so that we do not try to free it in case of failure */

    /* Write the header (.model .inputs .outputs). */
    if mname == nil {
        retval = fmt.Fprintf(fp, ".model DD\n.inputs");
    } else {
        retval = fmt.Fprintf(fp, ".model %s\n.inputs", mname);
    }
    if retval == -1 {
        ABC_FREE(sorted);
        return(0);
    }

    /* Write the input list by scanning the support array. */
    for i = 0; i < nvars; i++ {
        if sorted[i] {
            if inames == nil {
                retval = fmt.Fprintf(fp, " %d", i);
            } else {
                retval = fmt.Fprintf(fp, " %s", inames[i]);
            }
            if retval == -1 { goto failure; }
        }
    }
    ABC_FREE(sorted);
    sorted = nil;

    /* Write the .output line. */
    retval = fmt.Fprintf(fp, "\n.outputs");
    if retval == -1 { goto failure; }
    for i = 0; i < n; i++ {
        if onames == nil {
            retval = fmt.Fprintf(fp, " f%d", i);
        } else {
            retval = fmt.Fprintf(fp, " %s", onames[i]);
        }
        if retval == -1 { goto failure; }
    }
    retval = fmt.Fprintf(fp, "\n");
    if retval == -1 { goto failure; }

    retval = Cudd_DumpBlifBody(dd, n, f, inames, onames, fp, mv);
    if retval == 0 { goto failure; }

    /* Write trailer and return. */
    retval = fmt.Fprintf(fp, ".end\n");
    if retval == -1 { goto failure; }

    return(1);

failure:
    if sorted != nil { ABC_FREE(sorted); }
    if support != nil { Cudd_RecursiveDeref(dd,support); }
    return(0);

} /* end of Cudd_DumpBlif */


/**Function********************************************************************

  Synopsis    [Writes a blif body representing the argument BDDs.]

  Description [Writes a blif body representing the argument BDDs as a
  network of multiplexers.  No header (.model, .inputs, and .outputs) and
  footer (.end) are produced by this function.  One multiplexer is written
  for each BDD node. It returns 1 in case of success; 0 otherwise (e.g.,
  out-of-memory, file system full, or an ADD with constants different
  from 0 and 1).  Cudd_DumpBlifBody does not close the file: This is the
  caller responsibility. Cudd_DumpBlifBody uses a minimal unique subset of
  the hexadecimal address of a node as name for it.  If the argument
  inames is non-null, it is assumed to hold the pointers to the names
  of the inputs. Similarly for onames. This function prints out only
  .names part.]

  SideEffects [None]

  SeeAlso     [Cudd_DumpBlif Cudd_DumpDot Cudd_PrintDebug Cudd_DumpDDcal
  Cudd_DumpDaVinci Cudd_DumpFactoredForm]

******************************************************************************/
int
Cudd_DumpBlifBody(
  DdManager * dd /* manager */,
  int  n /* number of output nodes to be dumped */,
  DdNode ** f /* array of output nodes to be dumped */,
  char ** inames /* array of input names (or NULL) */,
  char ** onames /* array of output names (or NULL) */,
  FILE * fp /* pointer to the dump file */,
  int mv /* 0: blif, 1: blif-MV */)
{
    var visited *st__table  = nil;
    var retval int32;
    var i int32;

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);
    if visited == nil { goto failure; }

    /* Call the function that really gets the job done. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        retval = ddDoDumpBlif(dd,((DdNode *)((ptruint)(f[i]) & ~01)),fp,visited,inames,mv);
        if retval == 0 { goto failure; }
    }

    /* To account for the possible complement on the root,
    ** we put either a buffer or an inverter at the output of
    ** the multiplexer representing the top node.
    */
    for i = 0; i < n; i++ {
        if onames == nil {
            retval = fprintf(fp,
            #if SIZEOF_VOID_P == 8
                ".names %lx f%d\n", (ptruint) f[i] / (ptruint) sizeof(DdNode), i);
#else
                ".names %x f%d\n", (ptruint) f[i] / (ptruint) sizeof(DdNode), i);
#endif
        } else {
            retval = fprintf(fp,
            #if SIZEOF_VOID_P == 8
                ".names %lx %s\n", (ptruint) f[i] / (ptruint) sizeof(DdNode), onames[i]);
#else
                ".names %x %s\n", (ptruint) f[i] / (ptruint) sizeof(DdNode), onames[i]);
#endif
        }
        if retval == -1 { goto failure; }
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f[i]) & 01)) != 0 {
            retval = fmt.Fprintf(fp, "%s0 1\n", mv ? ".def 0\n" : "");
        } else {
            retval = fmt.Fprintf(fp, "%s1 1\n", mv ? ".def 0\n" : "");
        }
        if retval == -1 { goto failure; }
    }

    st__free_table(visited);
    return(1);

failure:
    if visited != nil { st__free_table(visited); }
    return(0);

} /* end of Cudd_DumpBlifBody */


/**Function********************************************************************

  Synopsis    [Writes a dot file representing the argument DDs.]

  Description [Writes a file representing the argument DDs in a format
  suitable for the graph drawing program dot.
  It returns 1 in case of success; 0 otherwise (e.g., out-of-memory,
  file system full).
  Cudd_DumpDot does not close the file: This is the caller
  responsibility. Cudd_DumpDot uses a minimal unique subset of the
  hexadecimal address of a node as name for it.
  If the argument inames is non-null, it is assumed to hold the pointers
  to the names of the inputs. Similarly for onames.
  Cudd_DumpDot uses the following convention to draw arcs:
    <ul>
    <li> solid line: THEN arcs;
    <li> dotted line: complement arcs;
    <li> dashed line: regular ELSE arcs.
    </ul>
  The dot options are chosen so that the drawing fits on a letter-size
  sheet.
  ]

  SideEffects [None]

  SeeAlso     [Cudd_DumpBlif Cudd_PrintDebug Cudd_DumpDDcal
  Cudd_DumpDaVinci Cudd_DumpFactoredForm]

******************************************************************************/
int
Cudd_DumpDot(
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
    var nvars int32  = dd.size;
    var visited *st__table  = nil;
    var gen *st__generator  = nil;
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
    support = Cudd_VectorSupport(dd,f,n);
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
    support = nil; /* so that we do not try to free it in case of failure */

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);
    if visited == nil { goto failure; }

    /* Collect all the nodes of this DD in the symbol table. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        retval = cuddCollectNodes(((DdNode *)((ptruint)(f[i]) & ~01)),visited);
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
    // inlining of macro Cudd_Regular
    refAddr = int64( ((DdNode *)((ptruint)(f[0]) & ~01)));
    diff = 0;
    gen = st__init_gen(visited);
    if gen == nil { goto failure; }
    for  st__gen(gen, (char **)&scan, nil) {
        diff |= refAddr ^ int64( scan);
    }
    st__free_gen(gen); gen = nil;

    /* Choose the mask. */
    for i = 0; uint32( i) < 8 * sizeof(long); i += 4 {
        mask = (1 << i) - 1;
        if diff <= mask { break; }
    }

    /* Write the header and the global attributes. */
    retval = fmt.Fprintf(fp, "digraph \"DD\" {\n");
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
        if sorted[dd.invperm[i]] {
            if inames == nil || inames[dd.invperm[i]] == nil {
                retval = fmt.Fprintf(fp, "\" %d \" -> ", dd.invperm[i]);
            } else {
                retval = fmt.Fprintf(fp, "\" %s \" -> ", inames[dd.invperm[i]]);
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
        if sorted[dd.invperm[i]] {
            retval = fmt.Fprintf(fp, "{ rank = same; ");
            if retval == -1 { goto failure; }
            if inames == nil || inames[dd.invperm[i]] == nil {
                retval = fmt.Fprintf(fp, "\" %d \";\n", dd.invperm[i]);
            } else {
                retval = fmt.Fprintf(fp, "\" %s \";\n", inames[dd.invperm[i]]);
            }
            if retval == -1 { goto failure; }
            nodelist = dd.subtables[i].nodelist;
            slots = dd.subtables[i].slots;
            for j = 0; j < slots; j++ {
                scan = nodelist[j];
                for scan != nil {
                    // inlining of macro st__is_member
                    if  st__lookup(visited,(char *) scan,(char **) 0) {
                        retval = fmt.Fprintf(fp, "\"%x\";\n", ((mask & (ptrint) scan) / sizeof(DdNode)));
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
                retval = fmt.Fprintf(fp, "\"%x\";\n", ((mask & (ptrint) scan) / sizeof(DdNode)));
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
        /* Account for the possible complement on the root. */
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f[i]) & 01)) != 0 {
            retval = fmt.Fprintf(fp, " -> \"%x\" [style = dotted];\n", ((mask & (ptrint) f[i]) / sizeof(DdNode)));
        } else {
            retval = fmt.Fprintf(fp, " -> \"%x\" [style = solid];\n", ((mask & (ptrint) f[i]) / sizeof(DdNode)));
        }
        if retval == -1 { goto failure; }
    }

    /* Edges from internal nodes. */
    for i = 0; i < nvars; i++ {
        if sorted[dd.invperm[i]] {
            nodelist = dd.subtables[i].nodelist;
            slots = dd.subtables[i].slots;
            for j = 0; j < slots; j++ {
                scan = nodelist[j];
                for scan != nil {
                    // inlining of macro st__is_member
                    if  st__lookup(visited,(char *) scan,(char **) 0) {
                        // inlining of macro cuddT
                        retval = fmt.Fprintf(fp, "\"%x\" -> \"%x\";\n", ((mask & (ptrint) scan) / sizeof(DdNode)), ((mask & (ptrint) ((scan).type.kids.T)) / sizeof(DdNode)));
                        if retval == -1 { goto failure; }
                        // inlining of macro Cudd_IsComplement
                        // inlining of macro cuddE
                        if int32( ((ptrint) (((scan).type.kids.E)) & 01)) {
                            // inlining of macro cuddE
                            retval = fmt.Fprintf(fp, "\"%x\" -> \"%x\" [style = dotted];\n", ((mask & (ptrint) scan) / sizeof(DdNode)), ((mask & (ptrint) ((scan).type.kids.E)) / sizeof(DdNode)));
                        } else {
                            // inlining of macro cuddE
                            retval = fmt.Fprintf(fp, "\"%x\" -> \"%x\" [style = dashed];\n", ((mask & (ptrint) scan) / sizeof(DdNode)), ((mask & (ptrint) ((scan).type.kids.E)) / sizeof(DdNode)));
                        }
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
                // inlining of macro cuddV
                retval = fmt.Fprintf(fp, "\"%x\" [label = \"%g\"];\n", ((mask & (ptrint) scan) / sizeof(DdNode)), ((scan).type.value));
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
    if support != nil { Cudd_RecursiveDeref(dd,support); }
    if visited != nil { st__free_table(visited); }
    return(0);

} /* end of Cudd_DumpDot */


/**Function********************************************************************

  Synopsis    [Writes a daVinci file representing the argument BDDs.]

  Description [Writes a daVinci file representing the argument BDDs.
  It returns 1 in case of success; 0 otherwise (e.g., out-of-memory or
  file system full).  Cudd_DumpDaVinci does not close the file: This
  is the caller responsibility. Cudd_DumpDaVinci uses a minimal unique
  subset of the hexadecimal address of a node as name for it.  If the
  argument inames is non-null, it is assumed to hold the pointers to
  the names of the inputs. Similarly for onames.]

  SideEffects [None]

  SeeAlso     [Cudd_DumpDot Cudd_PrintDebug Cudd_DumpBlif Cudd_DumpDDcal
  Cudd_DumpFactoredForm]

******************************************************************************/
int
Cudd_DumpDaVinci(
  DdManager * dd /* manager */,
  int  n /* number of output nodes to be dumped */,
  DdNode ** f /* array of output nodes to be dumped */,
  char ** inames /* array of input names (or NULL) */,
  char ** onames /* array of output names (or NULL) */,
  FILE * fp /* pointer to the dump file */)
{
    var support *DdNode  = nil;
    var scan *DdNode;
    var visited *st__table  = nil;
    var retval int32;
    var i int32;
    var gen *st__generator;
    var refAddr, diff, mask ptruint;

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);
    if visited == nil { goto failure; }

    /* Collect all the nodes of this DD in the symbol table. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        retval = cuddCollectNodes(((DdNode *)((ptruint)(f[i]) & ~01)),visited);
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
    // inlining of macro Cudd_Regular
    refAddr = (ptruint) ((DdNode *)((ptruint)(f[0]) & ~01));
    diff = 0;
    gen = st__init_gen(visited);
    for  st__gen(gen, (char **)&scan, nil) {
        diff |= refAddr ^ (ptruint) scan;
    }
    st__free_gen(gen);

    /* Choose the mask. */
    for i = 0; uint32( i) < 8 * sizeof(ptruint); i += 4 {
        mask = (1 << i) - 1;
        if diff <= mask { break; }
    }
    st__free_table(visited);

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);
    if visited == nil { goto failure; }

    retval = fmt.Fprintf(fp, "[");
    if retval == -1 { goto failure; }
    /* Call the function that really gets the job done. */
    for i = 0; i < n; i++ {
        if onames == nil {
            retval = fmt.Fprintf(fp, "l(\"f%d\",n(\"root\",[a(\"OBJECT\",\"f%d\")],", i, i);
        } else {
            retval = fmt.Fprintf(fp, "l(\"%s\",n(\"root\",[a(\"OBJECT\",\"%s\")],", onames[i], onames[i]);
        }
        if retval == -1 { goto failure; }
        // inlining of macro Cudd_IsComplement
        retval = fmt.Fprintf(fp, "[e(\"edge\",[a(\"EDGECOLOR\",\"%s\"),a(\"_DIR\",\"none\")],", (int32( ((ptrint) (f[i]) & 01))) ? "red" : "blue");
        if retval == -1 { goto failure; }
        // inlining of macro Cudd_Regular
        retval = ddDoDumpDaVinci(dd,((DdNode *)((ptruint)(f[i]) & ~01)),fp,visited,inames,mask);
        if retval == 0 { goto failure; }
        retval = fmt.Fprintf(fp, ")]))%s", i == n-1 ? "" : ",");
        if retval == -1 { goto failure; }
    }

    /* Write trailer and return. */
    retval = fmt.Fprintf(fp, "]\n");
    if retval == -1 { goto failure; }

    st__free_table(visited);
    return(1);

failure:
    if support != nil { Cudd_RecursiveDeref(dd,support); }
    if visited != nil { st__free_table(visited); }
    return(0);

} /* end of Cudd_DumpDaVinci */


/**Function********************************************************************

  Synopsis    [Writes a DDcal file representing the argument BDDs.]

  Description [Writes a DDcal file representing the argument BDDs.
  It returns 1 in case of success; 0 otherwise (e.g., out-of-memory or
  file system full).  Cudd_DumpDDcal does not close the file: This
  is the caller responsibility. Cudd_DumpDDcal uses a minimal unique
  subset of the hexadecimal address of a node as name for it.  If the
  argument inames is non-null, it is assumed to hold the pointers to
  the names of the inputs. Similarly for onames.]

  SideEffects [None]

  SeeAlso     [Cudd_DumpDot Cudd_PrintDebug Cudd_DumpBlif Cudd_DumpDaVinci
  Cudd_DumpFactoredForm]

******************************************************************************/
int
Cudd_DumpDDcal(
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
    var nvars int32  = dd.size;
    var visited *st__table  = nil;
    var retval int32;
    var i int32;
    var gen *st__generator;
    var refAddr, diff, mask ptruint;

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);
    if visited == nil { goto failure; }

    /* Collect all the nodes of this DD in the symbol table. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        retval = cuddCollectNodes(((DdNode *)((ptruint)(f[i]) & ~01)),visited);
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
    // inlining of macro Cudd_Regular
    refAddr = (ptruint) ((DdNode *)((ptruint)(f[0]) & ~01));
    diff = 0;
    gen = st__init_gen(visited);
    for  st__gen(gen, (char **)&scan, nil) {
        diff |= refAddr ^ (ptruint) scan;
    }
    st__free_gen(gen);

    /* Choose the mask. */
    for i = 0; uint32( i) < 8 * sizeof(ptruint); i += 4 {
        mask = (1 << i) - 1;
        if diff <= mask { break; }
    }
    st__free_table(visited);
    visited = nil;

    /* Build a bit array with the support of f. */
    sorted = ABC_ALLOC(int,nvars);
    if sorted == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        goto failure;
    }
    for i = 0; i < nvars; i++ { sorted[i] = 0; }

    /* Take the union of the supports of each output function. */
    support = Cudd_VectorSupport(dd,f,n);
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
    support = nil; /* so that we do not try to free it in case of failure */
    for i = 0; i < nvars; i++ {
        if sorted[dd.invperm[i]] {
            if inames == nil || inames[dd.invperm[i]] == nil {
                retval = fmt.Fprintf(fp, "v%d", dd.invperm[i]);
            } else {
                retval = fmt.Fprintf(fp, "%s", inames[dd.invperm[i]]);
            }
            if retval == -1 { goto failure; }
        }
        retval = fmt.Fprintf(fp, "%s", i == nvars - 1 ? "\n" : " * ");
        if retval == -1 { goto failure; }
    }
    ABC_FREE(sorted);
    sorted = nil;

    /* Initialize symbol table for visited nodes. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);
    if visited == nil { goto failure; }

    /* Call the function that really gets the job done. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        retval = ddDoDumpDDcal(dd,((DdNode *)((ptruint)(f[i]) & ~01)),fp,visited,inames,mask);
        if retval == 0 { goto failure; }
        if onames == nil {
            retval = fmt.Fprintf(fp, "f%d = ", i);
        } else {
            retval = fmt.Fprintf(fp, "%s = ", onames[i]);
        }
        if retval == -1 { goto failure; }
        retval = fmt.Fprintf(fp, "n%p%s\n", (void *) (((ptruint) f[i] & mask) /
                                 // inlining of macro Cudd_IsComplement
                                 (ptruint) sizeof(DdNode)), (int32( ((ptrint) (f[i]) & 01))) ? "'" : "");
        if retval == -1 { goto failure; }
    }

    /* Write trailer and return. */
    retval = fmt.Fprintf(fp, "[");
    if retval == -1 { goto failure; }
    for i = 0; i < n; i++ {
        if onames == nil {
            retval = fmt.Fprintf(fp, "f%d", i);
        } else {
            retval = fmt.Fprintf(fp, "%s", onames[i]);
        }
        retval = fmt.Fprintf(fp, "%s", i == n-1 ? "" : " ");
        if retval == -1 { goto failure; }
    }
    retval = fmt.Fprintf(fp, "]\n");
    if retval == -1 { goto failure; }

    if visited != nil {
        st__free_table(visited); }
    return(1);

failure:
    if sorted != nil { ABC_FREE(sorted); }
    if support != nil { Cudd_RecursiveDeref(dd,support); }
    if visited != nil { st__free_table(visited); }
    return(0);

} /* end of Cudd_DumpDDcal */


/**Function********************************************************************

  Synopsis    [Writes factored forms representing the argument BDDs.]

  Description [Writes factored forms representing the argument BDDs.
  The format of the factored form is the one used in the genlib files
  for technology mapping in sis.  It returns 1 in case of success; 0
  otherwise (e.g., file system full).  Cudd_DumpFactoredForm does not
  close the file: This is the caller responsibility. Caution must be
  exercised because a factored form may be exponentially larger than
  the argument BDD.  If the argument inames is non-null, it is assumed
  to hold the pointers to the names of the inputs. Similarly for
  onames.]

  SideEffects [None]

  SeeAlso     [Cudd_DumpDot Cudd_PrintDebug Cudd_DumpBlif Cudd_DumpDaVinci
  Cudd_DumpDDcal]

******************************************************************************/
int
Cudd_DumpFactoredForm(
  DdManager * dd /* manager */,
  int  n /* number of output nodes to be dumped */,
  DdNode ** f /* array of output nodes to be dumped */,
  char ** inames /* array of input names (or NULL) */,
  char ** onames /* array of output names (or NULL) */,
  FILE * fp /* pointer to the dump file */)
{
    var retval int32;
    var i int32;

    /* Call the function that really gets the job done. */
    for i = 0; i < n; i++ {
        if onames == nil {
            retval = fmt.Fprintf(fp, "f%d = ", i);
        } else {
            retval = fmt.Fprintf(fp, "%s = ", onames[i]);
        }
        if retval == -1 { return(0); }
        // inlining of macro DD_ONE
        if f[i] == ((dd).one) {
            retval = fmt.Fprintf(fp, "CONST1");
            if retval == -1 { return(0); }
        // inlining of macro Cudd_Not
        // inlining of macro DD_ONE
        // inlining of macro DD_ZERO
        } else if f[i] == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) || f[i] == ((dd).zero) {
            retval = fmt.Fprintf(fp, "CONST0");
            if retval == -1 { return(0); }
        } else {
            // inlining of macro Cudd_IsComplement
            retval = fmt.Fprintf(fp, "%s", (int32( ((ptrint) (f[i]) & 01))) ? "!(" : "");
            if retval == -1 { return(0); }
            // inlining of macro Cudd_Regular
            retval = ddDoDumpFactoredForm(dd,((DdNode *)((ptruint)(f[i]) & ~01)),fp,inames);
            if retval == 0 { return(0); }
            // inlining of macro Cudd_IsComplement
            retval = fmt.Fprintf(fp, "%s", (int32( ((ptrint) (f[i]) & 01))) ? ")" : "");
            if retval == -1 { return(0); }
        }
        retval = fmt.Fprintf(fp, "%s", i == n-1 ? "" : "\n");
        if retval == -1 { return(0); }
    }

    return(1);

} /* end of Cudd_DumpFactoredForm */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_DumpBlif.]

  Description [Performs the recursive step of Cudd_DumpBlif. Traverses
  the BDD f and writes a multiplexer-network description to the file
  pointed by fp in blif format. f is assumed to be a regular pointer
  and ddDoDumpBlif guarantees this assumption in the recursive calls.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func ddDoDumpBlif(dd *DdManager, f *DdNode, fp *os.File, visited *st__table, names *string, mv int32) int32 {
    var T, E *DdNode;
    var retval int32;

#ifdef DD_DEBUG
    assert(!Cudd_IsComplement(f));
#endif

    /* If already visited, nothing to do. */
    // inlining of macro st__is_member
    if  st__lookup(visited,(char *) f,(char **) 0) == 1 {
        return(1); }

    /* Check for abnormal condition that should never happen. */
    if f == nil {
        return(0); }

    /* Mark node as visited. */
    if st__insert(visited, (char *) f, nil) == st__OUT_OF_MEM {
        return(0); }

    /* Check for special case: If constant node, generate constant 1. */
    // inlining of macro DD_ONE
    if f == ((dd).one) {
#if SIZEOF_VOID_P == 8
        retval = fprintf(fp, ".names %lx\n1\n",(ptruint) f / (ptruint) sizeof(DdNode));
#else
        retval = fprintf(fp, ".names %x\n1\n",(ptruint) f / (ptruint) sizeof(DdNode));
#endif
        if retval == -1 {
            return(0);
        } else {
            return(1);
        }
    }

    /* Check whether this is an ADD. We deal with 0-1 ADDs, but not
    ** with the general case.
    */
    // inlining of macro DD_ZERO
    if f == ((dd).zero) {
#if SIZEOF_VOID_P == 8
        retval = fprintf(fp, ".names %lx\n%s",
                         (ptruint) f / (ptruint) sizeof(DdNode),
                         mv ? "0\n" : "");
#else
        retval = fprintf(fp, ".names %x\n%s",
                         (ptruint) f / (ptruint) sizeof(DdNode),
                         mv ? "0\n" : "");
#endif
        if retval == -1 {
            return(0);
        } else {
            return(1);
        }
    }
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        return(0); }

    /* Recursive calls. */
    // inlining of macro cuddT
    T = ((f).type.kids.T);
    retval = ddDoDumpBlif(dd,T,fp,visited,names,mv);
    if retval != 1 { return(retval); }
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    E = ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01));
    retval = ddDoDumpBlif(dd,E,fp,visited,names,mv);
    if retval != 1 { return(retval); }

    /* Write multiplexer taking complement arc into account. */
    if names != nil {
        retval = fmt.Fprintf(fp, ".names %s", names[f.index]);
    } else {
#if SIZEOF_VOID_P == 8 && SIZEOF_INT == 4
        retval = fprintf(fp,".names %u", f->index);
#else
        retval = fprintf(fp,".names %hu", f->index);
#endif
    }
    if retval == -1 {
        return(0); }

#if SIZEOF_VOID_P == 8
    if (mv) {
        if (Cudd_IsComplement(cuddE(f))) {
            retval = fprintf(fp," %lx %lx %lx\n.def 0\n1 1 - 1\n0 - 0 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        } else {
            retval = fprintf(fp," %lx %lx %lx\n.def 0\n1 1 - 1\n0 - 1 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        }
    } else {
        if (Cudd_IsComplement(cuddE(f))) {
            retval = fprintf(fp," %lx %lx %lx\n11- 1\n0-0 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        } else {
            retval = fprintf(fp," %lx %lx %lx\n11- 1\n0-1 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        }
    }
#else
    if (mv) {
        if (Cudd_IsComplement(cuddE(f))) {
            retval = fprintf(fp," %x %x %x\n.def 0\n1 1 - 1\n0 - 0 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        } else {
            retval = fprintf(fp," %x %x %x\n.def 0\n1 1 - 1\n0 - 1 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        }
    } else {
        if (Cudd_IsComplement(cuddE(f))) {
            retval = fprintf(fp," %x %x %x\n11- 1\n0-0 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        } else {
            retval = fprintf(fp," %x %x %x\n11- 1\n0-1 1\n",
                (ptruint) T / (ptruint) sizeof(DdNode),
                (ptruint) E / (ptruint) sizeof(DdNode),
                (ptruint) f / (ptruint) sizeof(DdNode));
        }
    }
#endif
    if retval == -1 {
        return(0);
    } else {
        return(1);
    }

} /* end of ddDoDumpBlif */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_DumpDaVinci.]

  Description [Performs the recursive step of Cudd_DumpDaVinci. Traverses
  the BDD f and writes a term expression to the file
  pointed by fp in daVinci format. f is assumed to be a regular pointer
  and ddDoDumpDaVinci guarantees this assumption in the recursive calls.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func ddDoDumpDaVinci(dd *DdManager, f *DdNode, fp *os.File, visited *st__table, names *string, mask ptruint) int32 {
    var T, E *DdNode;
    var retval int32;
    var id ptruint;

#ifdef DD_DEBUG
    assert(!Cudd_IsComplement(f));
#endif

    id = ((ptruint) f & mask) / sizeof(DdNode);

    /* If already visited, insert a reference. */
    // inlining of macro st__is_member
    if  st__lookup(visited,(char *) f,(char **) 0) == 1 {
        retval = fmt.Fprintf(fp, "r(\"%p\")", (void *) id);
        if retval == -1 {
            return(0);
        } else {
            return(1);
        }
    }

    /* Check for abnormal condition that should never happen. */
    if f == nil {
        return(0); }

    /* Mark node as visited. */
    if st__insert(visited, (char *) f, nil) == st__OUT_OF_MEM {
        return(0); }

    /* Check for special case: If constant node, generate constant 1. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        // inlining of macro cuddV
        retval = fmt.Fprintf(fp, "l(\"%p\",n(\"constant\",[a(\"OBJECT\",\"%g\")],[]))", (void *) id, ((f).type.value));
        if retval == -1 {
            return(0);
        } else {
            return(1);
        }
    }

    /* Recursive calls. */
    if names != nil {
        retval = fmt.Fprintf(fp, "l(\"%p\",n(\"internal\",[a(\"OBJECT\",\"%s\"),", (void *) id, names[f.index]);
    } else {
        retval = fmt.Fprintf(fp, #if SIZEOF_VOID_P == 8
                         "l(\"%p\",n(\"internal\",[a(\"OBJECT\",\"%u\"),",
#else
                         "l(\"%p\",n(\"internal\",[a(\"OBJECT\",\"%hu\"),",
#endif
                                 (void *) id, f.index);
    }
    retval = fmt.Fprintf(fp, "a(\"_GO\",\"ellipse\")],[e(\"then\",[a(\"EDGECOLOR\",\"blue\"),a(\"_DIR\",\"none\")],");
    if retval == -1 { return(0); }
    // inlining of macro cuddT
    T = ((f).type.kids.T);
    retval = ddDoDumpDaVinci(dd,T,fp,visited,names,mask);
    if retval != 1 { return(retval); }
    // inlining of macro Cudd_IsComplement
    // inlining of macro cuddE
    retval = fmt.Fprintf(fp, "),e(\"else\",[a(\"EDGECOLOR\",\"%s\"),a(\"_DIR\",\"none\")],", (int32( ((ptrint) (((f).type.kids.E)) & 01))) ? "red" : "green");
    if retval == -1 { return(0); }
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    E = ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01));
    retval = ddDoDumpDaVinci(dd,E,fp,visited,names,mask);
    if retval != 1 { return(retval); }

    retval = fmt.Fprintf(fp, ")]))");
    if retval == -1 {
        return(0);
    } else {
        return(1);
    }

} /* end of ddDoDumpDaVinci */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_DumpDDcal.]

  Description [Performs the recursive step of Cudd_DumpDDcal. Traverses
  the BDD f and writes a line for each node to the file
  pointed by fp in DDcal format. f is assumed to be a regular pointer
  and ddDoDumpDDcal guarantees this assumption in the recursive calls.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func ddDoDumpDDcal(dd *DdManager, f *DdNode, fp *os.File, visited *st__table, names *string, mask ptruint) int32 {
    var T, E *DdNode;
    var retval int32;
    var id, idT, idE ptruint;

#ifdef DD_DEBUG
    assert(!Cudd_IsComplement(f));
#endif

    id = ((ptruint) f & mask) / sizeof(DdNode);

    /* If already visited, do nothing. */
    // inlining of macro st__is_member
    if  st__lookup(visited,(char *) f,(char **) 0) == 1 {
        return(1);
    }

    /* Check for abnormal condition that should never happen. */
    if f == nil {
        return(0); }

    /* Mark node as visited. */
    if st__insert(visited, (char *) f, nil) == st__OUT_OF_MEM {
        return(0); }

    /* Check for special case: If constant node, assign constant. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        // inlining of macro DD_ONE
        // inlining of macro DD_ZERO
        if f != ((dd).one) && f != ((dd).zero) {
            return(0); }
        // inlining of macro cuddV
        retval = fmt.Fprintf(fp, "n%p = %g\n", (void *) id, ((f).type.value));
        if retval == -1 {
            return(0);
        } else {
            return(1);
        }
    }

    /* Recursive calls. */
    // inlining of macro cuddT
    T = ((f).type.kids.T);
    retval = ddDoDumpDDcal(dd,T,fp,visited,names,mask);
    if retval != 1 { return(retval); }
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    E = ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01));
    retval = ddDoDumpDDcal(dd,E,fp,visited,names,mask);
    if retval != 1 { return(retval); }
    idT = ((ptruint) T & mask) / sizeof(DdNode);
    idE = ((ptruint) E & mask) / sizeof(DdNode);
    if names != nil {
        // inlining of macro Cudd_IsComplement
        // inlining of macro cuddE
        retval = fmt.Fprintf(fp, "n%p = %s * n%p + %s' * n%p%s\n", (void *) id, names[f.index], (void *) idT, names[f.index], (void *) idE, (int32( ((ptrint) (((f).type.kids.E)) & 01))) ? "'" : "");
    } else {
#if SIZEOF_VOID_P == 8
        retval = fprintf(fp, "n%p = v%u * n%p + v%u' * n%p%s\n",
#else
        retval = fprintf(fp, "n%p = v%hu * n%p + v%hu' * n%p%s\n",
#endif
                         (void *) id;
f.index;
(void *) idT;
f.index;
(void *) idE;
// inlining of macro Cudd_IsComplement
// inlining of macro cuddE
(int32( ((ptrint) (((f).type.kids.E)) & 01))) ? "'" : "";;
    }
    if retval == -1 {
        return(0);
    } else {
        return(1);
    }

} /* end of ddDoDumpDDcal */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_DumpFactoredForm.]

  Description [Performs the recursive step of
  Cudd_DumpFactoredForm. Traverses the BDD f and writes a factored
  form for each node to the file pointed by fp in terms of the
  factored forms of the children. Constants are propagated, and
  absorption is applied.  f is assumed to be a regular pointer and
  ddDoDumpFActoredForm guarantees this assumption in the recursive
  calls.]

  SideEffects [None]

  SeeAlso     [Cudd_DumpFactoredForm]

******************************************************************************/
func ddDoDumpFactoredForm(dd *DdManager, f *DdNode, fp *os.File, names *string) int32 {
    var T, E *DdNode;
    var retval int32;

#ifdef DD_DEBUG
    assert(!Cudd_IsComplement(f));
    assert(!Cudd_IsConstant(f));
#endif

    /* Check for abnormal condition that should never happen. */
    if f == nil {
        return(0); }

    /* Recursive calls. */
    // inlining of macro cuddT
    T = ((f).type.kids.T);
    // inlining of macro cuddE
    E = ((f).type.kids.E);
    // inlining of macro DD_ZERO
    if T != ((dd).zero) {
        // inlining of macro DD_ONE
        if E != ((dd).one) {
            if names != nil {
                retval = fmt.Fprintf(fp, "%s", names[f.index]);
            } else {
#if SIZEOF_VOID_P == 8 && SIZEOF_INT == 4
                retval = fprintf(fp, "x%u", f->index);
#else
                retval = fprintf(fp, "x%hu", f->index);
#endif
            }
            if retval == -1 { return(0); }
        }
        // inlining of macro DD_ONE
        if T != ((dd).one) {
            // inlining of macro DD_ONE
            retval = fmt.Fprintf(fp, "%s(", E != ((dd).one) ? " * " : "");
            if retval == -1 { return(0); }
            retval = ddDoDumpFactoredForm(dd,T,fp,names);
            if retval != 1 { return(retval); }
            retval = fmt.Fprintf(fp, ")");
            if retval == -1 { return(0); }
        }
        // inlining of macro Cudd_Not
        // inlining of macro DD_ONE
        // inlining of macro DD_ZERO
        if E == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) || E == ((dd).zero) { return(1); }
        retval = fmt.Fprintf(fp, " + ");
        if retval == -1 { return(0); }
    }
    // inlining of macro Cudd_Regular
    E = ((DdNode *)((ptruint)(E) & ~01));
    // inlining of macro DD_ONE
    if T != ((dd).one) {
        if names != nil {
            retval = fmt.Fprintf(fp, "!%s", names[f.index]);
        } else {
#if SIZEOF_VOID_P == 8 && SIZEOF_INT == 4
            retval = fprintf(fp, "!x%u", f->index);
#else
            retval = fprintf(fp, "!x%hu", f->index);
#endif
        }
        if retval == -1 { return(0); }
    }
    // inlining of macro DD_ONE
    if E != ((dd).one) {
        // inlining of macro DD_ONE
        // inlining of macro cuddE
        retval = fmt.Fprintf(fp, "%s%s(", T != ((dd).one) ? " * " : "", E != ((f).type.kids.E) ? "!" : "");
        if retval == -1 { return(0); }
        retval = ddDoDumpFactoredForm(dd,E,fp,names);
        if retval != 1 { return(retval); }
        retval = fmt.Fprintf(fp, ")");
        if retval == -1 { return(0); }
    }
    return(1);

} /* end of ddDoDumpFactoredForm */


//ABC_NAMESPACE_IMPL_END

