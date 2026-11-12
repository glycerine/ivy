package cudd

import "fmt"

import "os"

/**CFile***********************************************************************

  FileName    [cuddCheck.c]

  PackageName [cudd]

  Synopsis    [Functions to check consistency of data structures.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_DebugCheck()
                <li> Cudd_CheckKeys()
                </ul>
               Internal procedures included in this module:
                <ul>
                <li> cuddHeapProfile()
                <li> cuddPrintNode()
                <li> cuddPrintVarGroups()
                </ul>
               Static procedures included in this module:
                <ul>
                <li> debugFindParent()
                </ul>
                ]

  SeeAlso     []

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
static char rcsid[] DD_UNUSED = "$Id: cuddCheck.c,v 1.35 2009/03/08 02:49:01 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static void debugFindParent (DdManager *table, DdNode *node);

     
                                                              
      

/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Checks for inconsistencies in the DD heap.]

  Description [Checks for inconsistencies in the DD heap:
  <ul>
  <li> node has illegal index
  <li> live node has dead children
  <li> node has illegal Then or Else pointers
  <li> BDD/ADD node has identical children
  <li> ZDD node has zero then child
  <li> wrong number of total nodes
  <li> wrong number of dead nodes
  <li> ref count error at node
  </ul>
  Returns 0 if no inconsistencies are found; DD_OUT_OF_MEM if there is
  not enough memory; 1 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_CheckKeys]

******************************************************************************/
func Cudd_DebugCheck(table *DdManager) int32 {
    unsigned int i;
    var j, count int32;
    var slots int32;
    var nodelist *DdNodePtr;
    var f *DdNode;
    var sentinel *DdNode  = &(table.sentinel);
    var edgeTable *st__table;     /* stores internal ref count for each node */
    var gen *st__generator;
    var flag int32;
    var totalNode int32;
    var deadNode int32;
    var index int32;


    edgeTable = st__init_table( st__ptrcmp, st__ptrhash);
    if edgeTable == nil { return(CUDD_OUT_OF_MEM); }

    /* Check the BDD/ADD subtables. */
    for i = 0; i < uint32( table.size); i++ {
        index = table.invperm[i];
        if i != uint32( table.perm[index]) {
            (void) fmt.Fprintf(table.err, "Permutation corrupted: invperm[%d] = %d\t perm[%d] = %d\n", i, index, index, table.perm[index]);
        }
        nodelist = table.subtables[i].nodelist;
        slots = table.subtables[i].slots;

        totalNode = 0;
        deadNode = 0;
        for j = 0; j < slots; j++ {   /* for each subtable slot */
            f = nodelist[j];
            for f != sentinel {
                totalNode++;
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((f).type.kids.T) != nil && ((f).type.kids.E) != nil && f.ref != 0 {
                    if int32( f.index) != index {
                        (void) fmt.Fprintf(table.err, "Error: node has illegal index\n");
                        cuddPrintNode(f,table.err);
                        flag = 1;
                    }
                    // inlining of macro cuddI
                    // inlining of macro cuddT
                    if uint32( (((((f).type.kids.T).index)==CUDD_CONST_INDEX)?int32((((f).type.kids.T).index)):(table).perm[(((f).type.kids.T).index)])) <= i ||
                        // inlining of macro cuddI
                        // inlining of macro Cudd_Regular
                        // inlining of macro cuddE
                        uint32( (((((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).index)):(table).perm[(((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).index)]))
                        <= i {
                        (void) fmt.Fprintf(table.err, "Error: node has illegal children\n");
                        cuddPrintNode(f,table.err);
                        flag = 1;
                    }
                    // inlining of macro Cudd_Regular
                    // inlining of macro cuddT
                    if ((DdNode *)((ptruint)(((f).type.kids.T)) & ~01)) != ((f).type.kids.T) {
                        (void) fmt.Fprintf(table.err, "Error: node has illegal form\n");
                        cuddPrintNode(f,table.err);
                        flag = 1;
                    }
                    // inlining of macro cuddT
                    // inlining of macro cuddE
                    if ((f).type.kids.T) == ((f).type.kids.E) {
                        (void) fmt.Fprintf(table.err, "Error: node has identical children\n");
                        cuddPrintNode(f,table.err);
                        flag = 1;
                    }
                    // inlining of macro cuddT
                    // inlining of macro Cudd_Regular
                    // inlining of macro cuddE
                    if ((f).type.kids.T).ref == 0 || ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)).ref == 0 {
                        (void) fmt.Fprintf(table.err, "Error: live node has dead children\n");
                        cuddPrintNode(f,table.err);
                        flag =1;
                    }
                    /* Increment the internal reference count for the
                    ** then child of the current node.
                    */
                    // inlining of macro cuddT
                    if st__lookup_int(edgeTable,(char *)((f).type.kids.T),&count) {
                        count++;
                    } else {
                        count = 1;
                    }
                    // inlining of macro cuddT
                    if st__insert(edgeTable,(char *)((f).type.kids.T),
                    (char *)int64(count)) == st__OUT_OF_MEM {
                        st__free_table(edgeTable);
                        return(CUDD_OUT_OF_MEM);
                    }

                    /* Increment the internal reference count for the
                    ** else child of the current node.
                    */
                    // inlining of macro Cudd_Regular
                    // inlining of macro cuddE
                    if st__lookup_int(edgeTable,(char *)((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)),
                                      &count) {
                        count++;
                    } else {
                        count = 1;
                    }
                    // inlining of macro Cudd_Regular
                    // inlining of macro cuddE
                    if st__insert(edgeTable,(char *)((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)),
                    (char *)int64(count)) == st__OUT_OF_MEM {
                        st__free_table(edgeTable);
                        return(CUDD_OUT_OF_MEM);
                    }
                // inlining of macro cuddT
                // inlining of macro cuddE
                } else if ((f).type.kids.T) != nil && ((f).type.kids.E) != nil && f.ref == 0 {
                    deadNode++;
     
                                              
      
                } else {
                    fmt.Fprintf(table.err, "Error: node has illegal Then or Else pointers\n");
                    cuddPrintNode(f,table.err);
                    flag = 1;
                }

                f = f.next;
            }   /* for each element of the collision list */
        }       /* for each subtable slot */

        if uint32( totalNode) != table.subtables[i].keys {
            fmt.Fprintf(table.err, "Error: wrong number of total nodes\n");
            flag = 1;
        }
        if uint32( deadNode) != table.subtables[i].dead {
            fmt.Fprintf(table.err, "Error: wrong number of dead nodes\n");
            flag = 1;
        }
    }   /* for each BDD/ADD subtable */

    /* Check the ZDD subtables. */
    for i = 0; i < uint32( table.sizeZ); i++ {
        index = table.invpermZ[i];
        if i != uint32( table.permZ[index]) {
            (void) fmt.Fprintf(table.err, "Permutation corrupted: invpermZ[%d] = %d\t permZ[%d] = %d in ZDD\n", i, index, index, table.permZ[index]);
        }
        nodelist = table.subtableZ[i].nodelist;
        slots = table.subtableZ[i].slots;

        totalNode = 0;
        deadNode = 0;
        for j = 0; j < slots; j++ {   /* for each subtable slot */
            f = nodelist[j];
            for f != nil {
                totalNode++;
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((f).type.kids.T) != nil && ((f).type.kids.E) != nil && f.ref != 0 {
                    if int32( f.index) != index {
                        (void) fmt.Fprintf(table.err, "Error: ZDD node has illegal index\n");
                        cuddPrintNode(f,table.err);
                        flag = 1;
                    }
                    // inlining of macro Cudd_IsComplement
                    // inlining of macro cuddT
                    if (int32( ((ptrint) (((f).type.kids.T)) & 01))) ||
                        // inlining of macro Cudd_IsComplement
                        // inlining of macro cuddE
                        (int32( ((ptrint) (((f).type.kids.E)) & 01))) {
                        (void) fmt.Fprintf(table.err, "Error: ZDD node has complemented children\n");
                        cuddPrintNode(f,table.err);
                        flag = 1;
                    }
                    // inlining of macro cuddIZ
                    // inlining of macro cuddT
                    if uint32( (((((f).type.kids.T).index)==CUDD_CONST_INDEX)?int32((((f).type.kids.T).index)):(table).permZ[(((f).type.kids.T).index)])) <= i ||
                    // inlining of macro cuddIZ
                    // inlining of macro cuddE
                    uint32( (((((f).type.kids.E).index)==CUDD_CONST_INDEX)?int32((((f).type.kids.E).index)):(table).permZ[(((f).type.kids.E).index)])) <= i {
                        (void) fmt.Fprintf(table.err, "Error: ZDD node has illegal children\n");
                        cuddPrintNode(f,table.err);
                        // inlining of macro cuddT
                        cuddPrintNode(((f).type.kids.T),table.err);
                        // inlining of macro cuddE
                        cuddPrintNode(((f).type.kids.E),table.err);
                        flag = 1;
                    }
                    // inlining of macro cuddT
                    // inlining of macro DD_ZERO
                    if ((f).type.kids.T) == ((table).zero) {
                        (void) fmt.Fprintf(table.err, "Error: ZDD node has zero then child\n");
                        cuddPrintNode(f,table.err);
                        flag = 1;
                    }
                    // inlining of macro cuddT
                    // inlining of macro cuddE
                    if ((f).type.kids.T).ref == 0 || ((f).type.kids.E).ref == 0 {
                        (void) fmt.Fprintf(table.err, "Error: ZDD live node has dead children\n");
                        cuddPrintNode(f,table.err);
                        flag =1;
                    }
                    /* Increment the internal reference count for the
                    ** then child of the current node.
                    */
                    // inlining of macro cuddT
                    if st__lookup_int(edgeTable,(char *)((f).type.kids.T),&count) {
                        count++;
                    } else {
                        count = 1;
                    }
                    // inlining of macro cuddT
                    if st__insert(edgeTable,(char *)((f).type.kids.T),
                    (char *)int64(count)) == st__OUT_OF_MEM {
                        st__free_table(edgeTable);
                        return(CUDD_OUT_OF_MEM);
                    }

                    /* Increment the internal reference count for the
                    ** else child of the current node.
                    */
                    // inlining of macro cuddE
                    if st__lookup_int(edgeTable,(char *)((f).type.kids.E),&count) {
                        count++;
                    } else {
                        count = 1;
                    }
                    // inlining of macro cuddE
                    if st__insert(edgeTable,(char *)((f).type.kids.E),
                    (char *)int64(count)) == st__OUT_OF_MEM {
                        st__free_table(edgeTable);
                        table.errorCode = CUDD_MEMORY_OUT;
                        return(CUDD_OUT_OF_MEM);
                    }
                // inlining of macro cuddT
                // inlining of macro cuddE
                } else if ((f).type.kids.T) != nil && ((f).type.kids.E) != nil && f.ref == 0 {
                    deadNode++;
     
                                              
      
                } else {
                    fmt.Fprintf(table.err, "Error: ZDD node has illegal Then or Else pointers\n");
                    cuddPrintNode(f,table.err);
                    flag = 1;
                }

                f = f.next;
            }   /* for each element of the collision list */
        }       /* for each subtable slot */

        if uint32( totalNode) != table.subtableZ[i].keys {
            fmt.Fprintf(table.err, "Error: wrong number of total nodes in ZDD\n");
            flag = 1;
        }
        if uint32( deadNode) != table.subtableZ[i].dead {
            fmt.Fprintf(table.err, "Error: wrong number of dead nodes in ZDD\n");
            flag = 1;
        }
    }   /* for each ZDD subtable */

    /* Check the constant table. */
    nodelist = table.constants.nodelist;
    slots = table.constants.slots;

    totalNode = 0;
    deadNode = 0;
    for j = 0; j < slots; j++ {
        f = nodelist[j];
        for f != nil {
            totalNode++;
            if f.ref != 0 {
                if f.index != CUDD_CONST_INDEX {
                    fmt.Fprintf(table.err, "Error: node has illegal index\n");
#if SIZEOF_VOID_P == 8
                    fprintf(table->err,
                            "       node 0x%lx, id = %u, ref = %u, value = %g\n",
                            (ptruint)f,f->index,f->ref,cuddV(f));
#else
                    fprintf(table->err,
                            "       node 0x%x, id = %hu, ref = %hu, value = %g\n",
                            (ptruint)f,f->index,f->ref,cuddV(f));
#endif
                    flag = 1;
                }
            } else {
                deadNode++;
            }
            f = f.next;
        }
    }
    if uint32( totalNode) != table.constants.keys {
        (void) fmt.Fprintf(table.err, "Error: wrong number of total nodes in constants\n");
        flag = 1;
    }
    if uint32( deadNode) != table.constants.dead {
        (void) fmt.Fprintf(table.err, "Error: wrong number of dead nodes in constants\n");
        flag = 1;
    }
    gen = st__init_gen(edgeTable);
    for  st__gen(gen, (char **)&f, (char **)&count) {
        if count > int32((f.ref)) && f.ref != DD_MAXREF {
#if SIZEOF_VOID_P == 8
            fprintf(table->err,"ref count error at node 0x%lx, count = %d, id = %u, ref = %u, then = 0x%lx, else = 0x%lx\n",(ptruint)f,count,f->index,f->ref,(ptruint)cuddT(f),(ptruint)cuddE(f));
#else
            fprintf(table->err,"ref count error at node 0x%x, count = %d, id = %hu, ref = %hu, then = 0x%x, else = 0x%x\n",(ptruint)f,count,f->index,f->ref,(ptruint)cuddT(f),(ptruint)cuddE(f));
#endif
            debugFindParent(table,f);
            flag = 1;
        }
    }
    st__free_gen(gen);
    st__free_table(edgeTable);

    return (flag);

} /* end of Cudd_DebugCheck */


/**Function********************************************************************

  Synopsis    [Checks for several conditions that should not occur.]

  Description [Checks for the following conditions:
  <ul>
  <li>Wrong sizes of subtables.
  <li>Wrong number of keys found in unique subtable.
  <li>Wrong number of dead found in unique subtable.
  <li>Wrong number of keys found in the constant table
  <li>Wrong number of dead found in the constant table
  <li>Wrong number of total slots found
  <li>Wrong number of maximum keys found
  <li>Wrong number of total dead found
  </ul>
  Reports the average length of non-empty lists. Returns the number of
  subtables for which the number of keys is wrong.]

  SideEffects [None]

  SeeAlso     [Cudd_DebugCheck]

******************************************************************************/
func Cudd_CheckKeys(table *DdManager) int32 {
    var size int32;
    var i, j int32;
    var nodelist *DdNodePtr;
    var node *DdNode;
    var sentinel *DdNode  = &(table.sentinel);
    var subtable *DdSubtable;
    var keys int32;
    var dead int32;
    var count int32;
    var totalKeys int32;
    var totalSlots int32;
    var totalDead int32;
    var nonEmpty int32;
    unsigned int slots;
    var logSlots int32;
    var shift int32;

    size = table.size;

    for i = 0; i < size; i++ {
        subtable = &(table.subtables[i]);
        nodelist = subtable.nodelist;
        keys = subtable.keys;
        dead = subtable.dead;
        totalKeys += keys;
        slots = subtable.slots;
        shift = subtable.shift;
        logSlots = sizeof(int) * 8 - shift;
        if ((slots >> logSlots) << logSlots) != slots {
            (void) fmt.Fprintf(table.err, "Unique table %d is not the right power of 2\n", i);
            (void) fmt.Fprintf(table.err, "    slots = %d shift = %d\n", slots, shift);
        }
        totalSlots += slots;
        totalDead += dead;
        for j = 0; uint32( j) < slots; j++ {
            node = nodelist[j];
            if node != sentinel {
                nonEmpty++;
            }
            for node != sentinel {
                keys--;
                if node.ref == 0 {
                    dead--;
                }
                node = node.next;
            }
        }
        if keys != 0 {
            (void) fmt.Fprintf(table.err, "Wrong number of keys found \
            in unique table %d (difference=%d)\n", i, keys);
            count++;
        }
        if dead != 0 {
            (void) fmt.Fprintf(table.err, "Wrong number of dead found \
            in unique table no. %d (difference=%d)\n", i, dead);
        }
    }   /* for each BDD/ADD subtable */

    /* Check the ZDD subtables. */
    size = table.sizeZ;

    for i = 0; i < size; i++ {
        subtable = &(table.subtableZ[i]);
        nodelist = subtable.nodelist;
        keys = subtable.keys;
        dead = subtable.dead;
        totalKeys += keys;
        totalSlots += subtable.slots;
        totalDead += dead;
        for j = 0; uint32( j) < subtable.slots; j++ {
            node = nodelist[j];
            if node != nil {
                nonEmpty++;
            }
            for node != nil {
                keys--;
                if node.ref == 0 {
                    dead--;
                }
                node = node.next;
            }
        }
        if keys != 0 {
            (void) fmt.Fprintf(table.err, "Wrong number of keys found \
            in ZDD unique table no. %d (difference=%d)\n", i, keys);
            count++;
        }
        if dead != 0 {
            (void) fmt.Fprintf(table.err, "Wrong number of dead found \
            in ZDD unique table no. %d (difference=%d)\n", i, dead);
        }
    }   /* for each ZDD subtable */

    /* Check the constant table. */
    subtable = &(table.constants);
    nodelist = subtable.nodelist;
    keys = subtable.keys;
    dead = subtable.dead;
    totalKeys += keys;
    totalSlots += subtable.slots;
    totalDead += dead;
    for j = 0; uint32( j) < subtable.slots; j++ {
        node = nodelist[j];
        if node != nil {
            nonEmpty++;
        }
        for node != nil {
            keys--;
            if node.ref == 0 {
                dead--;
            }
            node = node.next;
        }
    }
    if keys != 0 {
        (void) fmt.Fprintf(table.err, "Wrong number of keys found \
        in the constant table (difference=%d)\n", keys);
        count++;
    }
    if dead != 0 {
        (void) fmt.Fprintf(table.err, "Wrong number of dead found \
        in the constant table (difference=%d)\n", dead);
    }
    if uint32( totalKeys) != table.keys + table.keysZ {
        (void) fmt.Fprintf(table.err, "Wrong number of total keys found \
        (difference=%d)\n", int32( (totalKeys-table.keys)));
    }
    if uint32( totalSlots) != table.slots {
        (void) fmt.Fprintf(table.err, "Wrong number of total slots found \
        (difference=%d)\n", int32( (totalSlots-table.slots)));
    }
    if table.minDead != uint32( (table.gcFrac * table.slots)) {
        (void) fmt.Fprintf(table.err, "Wrong number of minimum dead found \
        (%u vs. %u)\n", table.minDead, uint32( (table.gcFrac * float64( table.slots))));
    }
    if uint32( totalDead) != table.dead + table.deadZ {
        (void) fmt.Fprintf(table.err, "Wrong number of total dead found \
        (difference=%d)\n", int32( (totalDead-table.dead)));
    }
    (void)fmt.Printf("Average length of non-empty lists = %g\n", float64( table.keys) / float64( nonEmpty));

    return(count);

} /* end of Cudd_CheckKeys */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Prints information about the heap.]

  Description [Prints to the manager's stdout the number of live nodes for each
  level of the DD heap that contains at least one live node.  It also
  prints a summary containing:
  <ul>
  <li> total number of tables;
  <li> number of tables with live nodes;
  <li> table with the largest number of live nodes;
  <li> number of nodes in that table.
  </ul>
  If more than one table contains the maximum number of live nodes,
  only the one of lowest index is reported. Returns 1 in case of success
  and 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddHeapProfile(dd *DdManager) int32 {
    var ntables int32  = dd.size;
    var subtables *DdSubtable  = dd.subtables;
    int i,              /* loop index */
        nodes,          /* live nodes in i-th layer */
        retval,         /* return value of fprintf */
        largest = -1,   /* index of the table with most live nodes */
        maxnodes = -1,  /* maximum number of live nodes in a table */
        nonempty = 0;   /* number of tables with live nodes */

    /* Print header. */
#if SIZEOF_VOID_P == 8
    retval = fprintf(dd->out,"*** DD heap profile for 0x%lx ***\n",
                     (ptruint) dd);
#else
    retval = fprintf(dd->out,"*** DD heap profile for 0x%x ***\n",
                     (ptruint) dd);
#endif
    if retval == -1 { return 0; }

    /* Print number of live nodes for each nonempty table. */
    for i=0; i<ntables; i++ {
        nodes = subtables[i].keys - subtables[i].dead;
        if nodes {
            nonempty++;
            retval = fmt.Fprintf(dd.out, "%5d: %5d nodes\n", i, nodes);
            if retval == -1 { return 0; }
            if nodes > maxnodes {
                maxnodes = nodes;
                largest = i;
            }
        }
    }

    nodes = dd.constants.keys - dd.constants.dead;
    if nodes {
        nonempty++;
        retval = fmt.Fprintf(dd.out, "const: %5d nodes\n", nodes);
        if retval == -1 { return 0; }
        if nodes > maxnodes {
            maxnodes = nodes;
            largest = CUDD_CONST_INDEX;
        }
    }

    /* Print summary. */
    retval = fmt.Fprintf(dd.out, "Summary: %d tables, %d non-empty, largest: %d ", ntables+1, nonempty, largest);
    if retval == -1 { return 0; }
    retval = fmt.Fprintf(dd.out, "(with %d nodes)\n", maxnodes);
    if retval == -1 { return 0; }

    return(1);

} /* end of cuddHeapProfile */


/**Function********************************************************************

  Synopsis    [Prints out information on a node.]

  Description [Prints out information on a node.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddPrintNode(f *DdNode, fp *os.File) {
    // inlining of macro Cudd_Regular
    f = ((DdNode *)((ptruint)(f) & ~01));
#if SIZEOF_VOID_P == 8
    (void) fprintf(fp,"       node 0x%lx, id = %u, ref = %u, then = 0x%lx, else = 0x%lx\n",(ptruint)f,f->index,f->ref,(ptruint)cuddT(f),(ptruint)cuddE(f));
#else
    (void) fprintf(fp,"       node 0x%x, id = %hu, ref = %hu, then = 0x%x, else = 0x%x\n",(ptruint)f,f->index,f->ref,(ptruint)cuddT(f),(ptruint)cuddE(f));
#endif

} /* end of cuddPrintNode */



/**Function********************************************************************

  Synopsis    [Prints the variable groups as a parenthesized list.]

  Description [Prints the variable groups as a parenthesized list.
  For each group the level range that it represents is printed. After
  each group, the group's flags are printed, preceded by a `|'.  For
  each flag (except MTR_TERMINAL) a character is printed.
  <ul>
  <li>F: MTR_FIXED
  <li>N: MTR_NEWNODE
  <li>S: MTR_SOFT
  </ul>
  The second argument, silent, if different from 0, causes
  Cudd_PrintVarGroups to only check the syntax of the group tree.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
void
cuddPrintVarGroups(
  DdManager * dd /* manager */,
  MtrNode * root /* root of the group tree */,
  int zdd /* 0: BDD; 1: ZDD */,
  int silent /* flag to check tree syntax only */)
{
    var node *MtrNode;
    var level int32;

    assert(root != nil, "root != nil");
    assert(root.younger == nil || root.younger.elder == root, "root.younger == nil || root.younger.elder == root");
    assert(root.elder == nil || root.elder.younger == root, "root.elder == nil || root.elder.younger == root");
    if zdd {
        level = dd.permZ[root.index];
    } else {
        level = dd.perm[root.index];
    }
    if !silent { (void) fmt.Printf("(%d", level); }
    // inlining of macro MTR_TEST
    if (root.flags & (MTR_TERMINAL)) || root.child == nil {
        if !silent { (void) fmt.Printf(","); }
    } else {
        node = root.child;
        for node != nil {
            assert(node.low >= root.low && int32( (node.low + node.size)) <= int32( (root.low + root.size)), "node.low >= root.low && int32( (node.low + node.size)) <= int32( (root.low + root.size))");
            assert(node.parent == root, "node.parent == root");
            cuddPrintVarGroups(dd,node,zdd,silent);
            node = node.younger;
        }
    }
    if !silent {
        (void) fmt.Printf("%d", int32( (level + root.size - 1)));
        if root.flags != MTR_DEFAULT {
            (void) fmt.Printf("|");
            // inlining of macro MTR_TEST
            if root.flags & (MTR_FIXED) { (void) fmt.Printf("F"); }
            // inlining of macro MTR_TEST
            if root.flags & (MTR_NEWNODE) { (void) fmt.Printf("N"); }
            // inlining of macro MTR_TEST
            if root.flags & (MTR_SOFT) { (void) fmt.Printf("S"); }
        }
        (void) fmt.Printf(")");
        if root.parent == nil { (void) fmt.Printf("\n"); }
    }
    assert((root.flags &~(MTR_TERMINAL | MTR_SOFT | MTR_FIXED | MTR_NEWNODE)) == 0, "(root.flags &~(MTR_TERMINAL | MTR_SOFT | MTR_FIXED | MTR_NEWNODE)) == 0");
    return;

} /* end of cuddPrintVarGroups */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Searches the subtables above node for its parents.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func debugFindParent(table *DdManager, node *DdNode) {
    var i, j int32;
    var slots int32;
    var nodelist *DdNodePtr;
    var f *DdNode;

    // inlining of macro cuddI
    for i = 0; i < (((node.index)==CUDD_CONST_INDEX)?int32((node.index)):(table).perm[(node.index)]); i++ {
        nodelist = table.subtables[i].nodelist;
        slots = table.subtables[i].slots;

        for j=0;j<slots;j++ {
            f = nodelist[j];
            for f != nil {
                // inlining of macro cuddT
                // inlining of macro Cudd_Regular
                // inlining of macro cuddE
                if ((f).type.kids.T) == node || ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)) == node {
#if SIZEOF_VOID_P == 8
                    (void) fprintf(table->out,"parent is at 0x%lx, id = %u, ref = %u, then = 0x%lx, else = 0x%lx\n",
                        (ptruint)f,f->index,f->ref,(ptruint)cuddT(f),(ptruint)cuddE(f));
#else
                    (void) fprintf(table->out,"parent is at 0x%x, id = %hu, ref = %hu, then = 0x%x, else = 0x%x\n",
                        (ptruint)f,f->index,f->ref,(ptruint)cuddT(f),(ptruint)cuddE(f));
#endif
                }
                f = f.next;
            }
        }
    }

} /* end of debugFindParent */


     
                                                                               

                                                                        

                                                                     
                                                                      

                    

                               

                                                                               
           
                 
                    
                
 
                    
                      
                              

                                                    
                                                
                                          

                               
                            
                               
                                                                                          
                                              
                                                                  
                                                                                                                                                   
                                                                                                                                                                  
                 
                            
             
         
     
 
      


//ABC_NAMESPACE_IMPL_END

