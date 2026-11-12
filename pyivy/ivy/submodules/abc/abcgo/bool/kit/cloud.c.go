package kit

import "fmt"

/**CFile****************************************************************

  FileName    [cloudCore.c]

  PackageName [Fast application-specific BDD package.]

  Synopsis    [The package core.]

  Author      [Alan Mishchenko <alanmi@ece.pdx.edu>]
  
  Affiliation [ECE Department. Portland State University, Portland, Oregon.]

  Date        [Ver. 1.0. Started - June 10, 2002.]

  Revision    [$Id: cloudCore.c,v 1.0 2002/06/10 03:00:00 alanmi Exp $]

***********************************************************************/

//#include <string.h>
//#include "cloud.h"

//ABC_NAMESPACE_IMPL_START


// the number of operators using cache
var CacheOperNum int32 = 4;

// the ratio of cache size to the unique table size for each operator
static int CacheLogRatioDefault[4] = {
    2, // CLOUD_OPER_AND, 
    8, // CLOUD_OPER_XOR, 
    8, // CLOUD_OPER_BDIFF, 
    8  // CLOUD_OPER_LEQ 
};

// the ratio of cache size to the unique table size for each operator
static int CacheSize[4] = {
    2, // CLOUD_OPER_AND, 
    2, // CLOUD_OPER_XOR, 
    2, // CLOUD_OPER_BDIFF, 
    2  // CLOUD_OPER_LEQ 
};

////////////////////////////////////////////////////////////////////////
///                       FUNCTION DECLARATIONS                      ///
////////////////////////////////////////////////////////////////////////

// static functions
// static CloudNode * cloudMakeNode( CloudManager * dd, CloudVar v, CloudNode * t, CloudNode * e );

// static void        cloudCacheAllocate( CloudManager * dd, CloudOper oper );


////////////////////////////////////////////////////////////////////////
///                       FUNCTION DEFINITIONS                       ///
////////////////////////////////////////////////////////////////////////

/**Function********************************************************************

  Synopsis    [Starts the cloud manager.]

  Description [The first arguments is the number of elementary variables used.
  The second arguments is the number of bits of the unsigned integer used to 
  represent nodes in the unique table. If the second argument is 0, the package 
  assumes 23 to represent nodes, which is equivalent to 2^23 = 8,388,608 nodes.]
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_Init(nVars int32, nBits int32) *CloudManager {
    var dd *CloudManager;
    var i int32;
    var clk1, clk2 abctime;

    assert(nVars <= 100000, "nVars <= 100000");
    assert(nBits < 32, "nBits < 32");

    // assign the defaults
    if nBits == 0 {
        nBits = CLOUD_NODE_BITS; }

    // start the manager
    // inlining of macro ABC_CALLOC
    dd = &CloudManager{};
    dd.nMemUsed          += sizeof(CloudManager);

    // variables
    dd.nVars             = nVars;              // the number of variables allocated
    // bits
    dd.bitsNode          = nBits;              // the number of bits used for the node
    for  i = 0; i < CacheOperNum; i++  {
        dd.bitsCache[i]  = nBits - CacheLogRatioDefault[i];
    }
    // shifts
    dd.shiftUnique       = 8*sizeof(unsigned) - (nBits + 1); // gets node index in the hash table
    for  i = 0; i < CacheOperNum; i++  {
        dd.shiftCache[i] = 8*sizeof(unsigned) - dd.bitsCache[i];
    }
    // nodes
    dd.nNodesAlloc       = (1 << (nBits + 1)); // 2 ^ (nBits + 1)
    dd.nNodesLimit       = (1 << nBits);       // 2 ^  nBits

    // unique table
clk1 = Abc_Clock();
    // inlining of macro ABC_CALLOC
    dd.tUnique           = ((CloudNode *) calloc((size_t)(dd.nNodesAlloc), sizeof(CloudNode)));
    dd.nMemUsed         += sizeof(CloudNode) * dd.nNodesAlloc;
clk2 = Abc_Clock();
//ABC_PRT( "calloc() time", clk2 - clk1 ); 

    // set up the constant node (the only node that is not in the hash table)
    dd.nSignCur          = 1;
    dd.tUnique[0].s      = dd.nSignCur;
    dd.tUnique[0].v      = CLOUD_CONST_INDEX;
    dd.tUnique[0].e      = nil;
    dd.tUnique[0].t      = nil;
    dd.one               = dd.tUnique;
    // inlining of macro Cloud_Not
    dd.zero              = ((CloudNode*)(((ABC_PTRUINT_T)(dd.one)) ^  01))   // complement the node
;
    dd.nNodesCur         = 1;

    // special nodes
    dd.pNodeStart        = dd.tUnique + 1;
    dd.pNodeEnd          = dd.tUnique + dd.nNodesAlloc;

    // set up the elementary variables
    // inlining of macro ABC_ALLOC
    dd.vars              = ((CloudNode * *) malloc(sizeof(CloudNode *) * (size_t)(dd.nVars)));
    dd.nMemUsed         += sizeof(CloudNode *) * dd.nVars;
    for  i = 0; i < dd.nVars; i++  {
        dd.vars[i]   = cloudMakeNode( dd, i, dd.one, dd.zero );
    }

    return dd;
};

/**Function********************************************************************

  Synopsis    [Stops the cloud manager.]

  Description [The first arguments tells show many elementary variables are used.
  The second arguments tells how many bits of the unsigned integer are used
  to represent regular nodes in the unique table.]
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_Quit(dd *CloudManager) {
    var i int32;
    // inlining of macro ABC_FREE
    ((dd.ppNodes) ? (free((char *) (dd.ppNodes)), (dd.ppNodes) = 0) : 0);
    // inlining of macro ABC_FREE
    ((dd.tUnique) ? (free((char *) (dd.tUnique)), (dd.tUnique) = 0) : 0); 
    // inlining of macro ABC_FREE
    ((dd.vars) ? (free((char *) (dd.vars)), (dd.vars) = 0) : 0); 
    for  i = 0; i < 4; i++  {
        // inlining of macro ABC_FREE
        ((dd.tCaches[i]) ? (free((char *) (dd.tCaches[i])), (dd.tCaches[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((dd) ? (free((char *) (dd)), (dd) = 0) : 0); 
}

/**Function********************************************************************

  Synopsis    [Prepares the manager for another run.]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_Restart(dd *CloudManager) {
    var i int32;
    assert(dd.one.s == dd.nSignCur, "dd.one.s == dd.nSignCur");
    dd.nSignCur++;
    dd.one.s++;
    for  i = 0; i < dd.nVars; i++  {
        dd.vars[i].s++;
    }
    dd.nNodesCur = 1 + dd.nVars;
}

/**Function********************************************************************

  Synopsis    [This optional function allocates operation cache of the given size.]

  Description [Cache for each operation is allocated independently when the first 
  operation of the given type is performed. The user can allocate cache of his/her 
  preferred size by calling Cloud_CacheAllocate before the first operation of the 
  given type is performed, but this call is optional. Argument "logratio" gives
  the binary logarithm of the ratio of the size of the unique table to that of cache.
  For example, if "logratio" is equal to 3, and the unique table will be 2^3=8 times
  larger than cache; so, if unique table is 2^23 = 8,388,608 nodes, the cache size 
  will be 2^3=8 times smaller and equal to 2^20 = 1,048,576 entries.]
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_CacheAllocate(dd *CloudManager, oper CloudOper, logratio int32) {
    assert(logratio > 0, "logratio > 0");            // cache cannot be larger than the unique table 
    assert(logratio < dd.bitsNode, "logratio < dd.bitsNode"); // cache cannot be smaller than 2 entries

    if logratio != 0 {
        dd.bitsCache[oper]  = dd.bitsNode - logratio;
        dd.shiftCache[oper] = 8*sizeof(unsigned) - dd.bitsCache[oper];
    }
    cloudCacheAllocate( dd, oper );
}

/**Function********************************************************************

  Synopsis    [Internal cache allocation.]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func cloudCacheAllocate(dd *CloudManager, oper CloudOper) {
    var nCacheEntries int32  = (1 << dd.bitsCache[oper]);

    if CacheSize[oper] == 1 {
        // inlining of macro ABC_CALLOC
        dd.tCaches[oper] = (CloudCacheEntry2 *)((CloudCacheEntry1 *) calloc((size_t)(nCacheEntries), sizeof(CloudCacheEntry1)));
        dd.nMemUsed     += sizeof(CloudCacheEntry1) * nCacheEntries;
    
    } else if CacheSize[oper] == 2 {
        // inlining of macro ABC_CALLOC
        dd.tCaches[oper] = (CloudCacheEntry2 *)((CloudCacheEntry2 *) calloc((size_t)(nCacheEntries), sizeof(CloudCacheEntry2)));
        dd.nMemUsed     += sizeof(CloudCacheEntry2) * nCacheEntries;
    
    } else if CacheSize[oper] == 3 {
        // inlining of macro ABC_CALLOC
        dd.tCaches[oper] = (CloudCacheEntry2 *)((CloudCacheEntry3 *) calloc((size_t)(nCacheEntries), sizeof(CloudCacheEntry3)));
        dd.nMemUsed     += sizeof(CloudCacheEntry3) * nCacheEntries;
    }
}



/**Function********************************************************************

  Synopsis    [Returns or creates a new node]

  Description [Checks the unique table for the existance of the node. If the node is 
  present, returns the node. If the node is absent, creates a new node.]
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_MakeNode(dd *CloudManager, v CloudVar, t *CloudNode, e *CloudNode) *CloudNode {
    var pRes *CloudNode;
    // inlining of macro CLOUD_ASSERT
    assert((t) >= dd.tUnique && (t) < dd.tUnique+dd.nNodesAlloc, "(t) >= dd.tUnique && (t) < dd.tUnique+dd.nNodesAlloc"); 
    // inlining of macro CLOUD_ASSERT
    assert((e) >= dd.tUnique && (e) < dd.tUnique+dd.nNodesAlloc, "(e) >= dd.tUnique && (e) < dd.tUnique+dd.nNodesAlloc");
    // inlining of macro Cloud_V
    // inlining of macro Cloud_Regular
    assert(v < ((((CloudNode*)(((ABC_PTRUINT_T)(t)) & ~01))   // get the regular node (w/o bubble)
    // inlining of macro Cloud_V
    // inlining of macro Cloud_Regular
    ).v) && v < ((((CloudNode*)(((ABC_PTRUINT_T)(e)) & ~01))   // get the regular node (w/o bubble)
    ).v), "v < /*__starting_point_inline_macro:Cloud_V__*/((/*__starting_point_inline_macro:Cloud_Regular__*/((CloudNode*)(((ABC_PTRUINT_T)(t)) & ~01))   // get the regular node (w/o bubble)\n).v) && v < /*__starting_point_inline_macro:Cloud_V__*/((/*__starting_point_inline_macro:Cloud_Regular__*/((CloudNode*)(((ABC_PTRUINT_T)(e)) & ~01))   // get the regular node (w/o bubble)\n).v)");       // variable should be above in the order
    // inlining of macro Cloud_IsComplement
    if  int32((((ABC_PTRUINT_T)(t)) & 01)) != 0 {           // check if complemented
  
        // inlining of macro Cloud_Not
        pRes = cloudMakeNode( dd, v, ((CloudNode*)(((ABC_PTRUINT_T)(t)) ^  01))   // complement the node
// inlining of macro Cloud_Not
, ((CloudNode*)(((ABC_PTRUINT_T)(e)) ^  01))   // complement the node
 );
        if pRes != nil {
            // inlining of macro Cloud_Not
            pRes = ((CloudNode*)(((ABC_PTRUINT_T)(pRes)) ^  01))   // complement the node
; }            
    
    } else {
        pRes = cloudMakeNode( dd, v, t, e ); }
    return pRes;
}

/**Function********************************************************************

  Synopsis    [Returns or creates a new node]

  Description [Checks the unique table for the existance of the node. If the node is 
  present, returns the node. If the node is absent, creates a new node.]
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func cloudMakeNode(dd *CloudManager, v CloudVar, t *CloudNode, e *CloudNode) *CloudNode {
    var entryUnique *CloudNode;

    // inlining of macro CLOUD_ASSERT
    assert((t) >= dd.tUnique && (t) < dd.tUnique+dd.nNodesAlloc, "(t) >= dd.tUnique && (t) < dd.tUnique+dd.nNodesAlloc"); 
    // inlining of macro CLOUD_ASSERT
    assert((e) >= dd.tUnique && (e) < dd.tUnique+dd.nNodesAlloc, "(e) >= dd.tUnique && (e) < dd.tUnique+dd.nNodesAlloc");

    assert((int32(v)) >= 0 && (int32(v)) < dd.nVars, "(int32(v)) >= 0 && (int32(v)) < dd.nVars");  // the variable must be in the range
    // inlining of macro Cloud_V
    // inlining of macro Cloud_Regular
    assert(v < ((((CloudNode*)(((ABC_PTRUINT_T)(t)) & ~01))   // get the regular node (w/o bubble)
    // inlining of macro Cloud_V
    // inlining of macro Cloud_Regular
    ).v) && v < ((((CloudNode*)(((ABC_PTRUINT_T)(e)) & ~01))   // get the regular node (w/o bubble)
    ).v), "v < /*__starting_point_inline_macro:Cloud_V__*/((/*__starting_point_inline_macro:Cloud_Regular__*/((CloudNode*)(((ABC_PTRUINT_T)(t)) & ~01))   // get the regular node (w/o bubble)\n).v) && v < /*__starting_point_inline_macro:Cloud_V__*/((/*__starting_point_inline_macro:Cloud_Regular__*/((CloudNode*)(((ABC_PTRUINT_T)(e)) & ~01))   // get the regular node (w/o bubble)\n).v)");       // variable should be above in the order
    // inlining of macro Cloud_IsComplement
    assert( !(int32((((ABC_PTRUINT_T)(t)) & 01)))           // check if complemented
     );                 // the THEN edge must not be complemented

    // make sure we are not searching for the constant node
    assert((t != nil) && (e != nil), "t && e");

    // get the unique entry
    // inlining of macro cloudHashCudd3
    entryUnique = dd.tUnique + ((((uint32((ABC_PTRUINT_T)(v)) * DD_P1 + uint32((ABC_PTRUINT_T)(t))) * DD_P2 + uint32((ABC_PTRUINT_T)(e))) * DD_P3) >> (dd.shiftUnique));
    for  entryUnique.s == dd.nSignCur  {
        // compare the node
        if entryUnique.v == v && entryUnique.t == t && entryUnique.e == e { // the node is found
            dd.nUniqueHits++;
            return entryUnique;  // returns the node
        }
        // increment the hash value modulus the hash table size
        if entryUnique++; entryUnique - dd.tUnique == dd.nNodesAlloc {
            entryUnique = dd.tUnique + 1; }
        // increment the number of steps through the table
        dd.nUniqueSteps++;
    }
    dd.nUniqueMisses++;

    // check if the new node can be created
    if ++dd.nNodesCur == dd.nNodesLimit { // initiate the restart
        fmt.Printf("Cloud needs restart!\n");
//        fflush( stdout );
//        exit(1);
        return nil;
    }
    // create the node
    entryUnique.s   = dd.nSignCur;
    entryUnique.v   = v;
    entryUnique.t   = t;
    entryUnique.e   = e;
    return entryUnique;  // returns the node
}


/**Function********************************************************************

  Synopsis    [Performs the AND or two BDDs]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func cloudBddAnd(dd *CloudManager, f *CloudNode, g *CloudNode) *CloudNode {
    var F, G, r *CloudNode;
    var cacheEntry *CloudCacheEntry2;
    var fv, fnv, gv, gnv, t, e *CloudNode;
    var var CloudVar;

    assert(f <= g, "f <= g");

    // terminal cases
    // inlining of macro Cloud_Regular
    F = ((CloudNode*)(((ABC_PTRUINT_T)(f)) & ~01))   // get the regular node (w/o bubble)
;
    // inlining of macro Cloud_Regular
    G = ((CloudNode*)(((ABC_PTRUINT_T)(g)) & ~01))   // get the regular node (w/o bubble)
;
    if F == G {
        if f == g {
            return f;
        } else {
            return dd.zero; }
    }
    if F == dd.one {
        if f == dd.one {
            return g;
        } else {
            return f; }
    }

    // check cache
    // inlining of macro cloudHashCudd2
    cacheEntry = dd.tCaches[CLOUD_OPER_AND] + (((uint32((ABC_PTRUINT_T)(f)) * DD_P1 + uint32((ABC_PTRUINT_T)(g))) * DD_P2) >> (dd.shiftCache[CLOUD_OPER_AND]));
//    cacheEntry = dd->tCaches[CLOUD_OPER_AND] + cloudHashBuddy2(f, g, dd->shiftCache[CLOUD_OPER_AND]);
    // inlining of macro cloudCacheLookup2
    
    if (cacheEntry).s == (dd.nSignCur) && (cacheEntry).a == (f) && (cacheEntry).b == (g) {
        r = ((cacheEntry).r);
    } else {
        r = (0);
    }
    if r != nil {
        dd.nCacheHits++;
        return r;
    }
    dd.nCacheMisses++;


    // compute cofactors
    // inlining of macro cloudV
    if  ((F).v) <= ((G).v) {
        // inlining of macro cloudV
        var = ((F).v);
        // inlining of macro Cloud_IsComplement
        if  int32((((ABC_PTRUINT_T)(f)) & 01)) != 0 {           // check if complemented
  
            // inlining of macro Cloud_Not
            // inlining of macro cloudE
            fnv = ((CloudNode*)(((ABC_PTRUINT_T)(((F).e))) ^  01))   // complement the node
;
            // inlining of macro Cloud_Not
            // inlining of macro cloudT
            fv  = ((CloudNode*)(((ABC_PTRUINT_T)(((F).t))) ^  01))   // complement the node
;
        
        } else {
            // inlining of macro cloudE
            fnv = ((F).e);
            // inlining of macro cloudT
            fv  = ((F).t);
        }
    
    } else {
        // inlining of macro cloudV
        var = ((G).v);
        fv  = fnv = f;
    }

    // inlining of macro cloudV
    if  ((G).v) <= ((F).v) {
        // inlining of macro Cloud_IsComplement
        if  int32((((ABC_PTRUINT_T)(g)) & 01)) != 0 {           // check if complemented
  
            // inlining of macro Cloud_Not
            // inlining of macro cloudE
            gnv = ((CloudNode*)(((ABC_PTRUINT_T)(((G).e))) ^  01))   // complement the node
;
            // inlining of macro Cloud_Not
            // inlining of macro cloudT
            gv  = ((CloudNode*)(((ABC_PTRUINT_T)(((G).t))) ^  01))   // complement the node
;
        
        } else {
            // inlining of macro cloudE
            gnv = ((G).e);
            // inlining of macro cloudT
            gv  = ((G).t);
        }
    
    } else {
        gv = gnv = g;
    }

    if fv <= gv {
        t = cloudBddAnd( dd, fv, gv );
    } else {
        t = cloudBddAnd( dd, gv, fv ); }

    if t == nil {
        return nil; }

    if fnv <= gnv {
        e = cloudBddAnd( dd, fnv, gnv );
    } else {
        e = cloudBddAnd( dd, gnv, fnv ); }

    if e == nil {
        return nil; }

    if t == e {
        r = t;
    } else {
        // inlining of macro Cloud_IsComplement
        if  int32((((ABC_PTRUINT_T)(t)) & 01)) != 0 {           // check if complemented
  
            // inlining of macro Cloud_Not
            r = cloudMakeNode( dd, var, ((CloudNode*)(((ABC_PTRUINT_T)(t)) ^  01))   // complement the node
// inlining of macro Cloud_Not
, ((CloudNode*)(((ABC_PTRUINT_T)(e)) ^  01))   // complement the node
 );
            if r == nil {
                return nil; }
            // inlining of macro Cloud_Not
            r = ((CloudNode*)(((ABC_PTRUINT_T)(r)) ^  01))   // complement the node
;
        
        } else {
            r = cloudMakeNode( dd, var, t, e );
            if r == nil {
                return nil; }
        }
    }
    // inlining of macro cloudCacheInsert2
    ((cacheEntry).s = (dd.nSignCur));
((cacheEntry).a = (f));
((cacheEntry).b = (g));
((cacheEntry).r = (r));;
    return r;
}

/**Function********************************************************************

  Synopsis    [Performs the AND or two BDDs]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func cloudBddAnd_gate(dd *CloudManager, f *CloudNode, g *CloudNode) *CloudNode {
    if f <= g {
        return cloudBddAnd(dd,f,g);
    } else {
        return cloudBddAnd(dd,g,f); }
}

/**Function********************************************************************

  Synopsis    [Performs the AND or two BDDs]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_bddAnd(dd *CloudManager, f *CloudNode, g *CloudNode) *CloudNode {
    // inlining of macro Cloud_Regular
    if  ((CloudNode*)(((ABC_PTRUINT_T)(f)) & ~01))   // get the regular node (w/o bubble)
 // inlining of macro Cloud_Regular
 == nil || ((CloudNode*)(((ABC_PTRUINT_T)(g)) & ~01))   // get the regular node (w/o bubble)
 == nil {
        return nil; }
    // inlining of macro CLOUD_ASSERT
    assert((f) >= dd.tUnique && (f) < dd.tUnique+dd.nNodesAlloc, "(f) >= dd.tUnique && (f) < dd.tUnique+dd.nNodesAlloc");
    // inlining of macro CLOUD_ASSERT
    assert((g) >= dd.tUnique && (g) < dd.tUnique+dd.nNodesAlloc, "(g) >= dd.tUnique && (g) < dd.tUnique+dd.nNodesAlloc");
    if dd.tCaches[CLOUD_OPER_AND] == nil {
        cloudCacheAllocate( dd, CLOUD_OPER_AND ); }
    return cloudBddAnd_gate( dd, f, g );
}

/**Function********************************************************************

  Synopsis    [Performs the OR or two BDDs]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_bddOr(dd *CloudManager, f *CloudNode, g *CloudNode) *CloudNode {
    var res *CloudNode;
    // inlining of macro Cloud_Regular
    if  ((CloudNode*)(((ABC_PTRUINT_T)(f)) & ~01))   // get the regular node (w/o bubble)
 // inlining of macro Cloud_Regular
 == nil || ((CloudNode*)(((ABC_PTRUINT_T)(g)) & ~01))   // get the regular node (w/o bubble)
 == nil {
        return nil; }
    // inlining of macro CLOUD_ASSERT
    assert((f) >= dd.tUnique && (f) < dd.tUnique+dd.nNodesAlloc, "(f) >= dd.tUnique && (f) < dd.tUnique+dd.nNodesAlloc");
    // inlining of macro CLOUD_ASSERT
    assert((g) >= dd.tUnique && (g) < dd.tUnique+dd.nNodesAlloc, "(g) >= dd.tUnique && (g) < dd.tUnique+dd.nNodesAlloc");
    if dd.tCaches[CLOUD_OPER_AND] == nil {
        cloudCacheAllocate( dd, CLOUD_OPER_AND ); }
    // inlining of macro Cloud_Not
    res = cloudBddAnd_gate( dd, ((CloudNode*)(((ABC_PTRUINT_T)(f)) ^  01))   // complement the node
// inlining of macro Cloud_Not
, ((CloudNode*)(((ABC_PTRUINT_T)(g)) ^  01))   // complement the node
 );
    // inlining of macro Cloud_NotCond
    res = ((CloudNode*)(((ABC_PTRUINT_T)(res)) ^ (res != nil)))   // complement the node conditionally
;
    return res;
}

/**Function********************************************************************

  Synopsis    [Performs the XOR or two BDDs]

  Description []
               
  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_bddXor(dd *CloudManager, f *CloudNode, g *CloudNode) *CloudNode {
    var t0, t1, r *CloudNode;
    // inlining of macro Cloud_Regular
    if  ((CloudNode*)(((ABC_PTRUINT_T)(f)) & ~01))   // get the regular node (w/o bubble)
 // inlining of macro Cloud_Regular
 == nil || ((CloudNode*)(((ABC_PTRUINT_T)(g)) & ~01))   // get the regular node (w/o bubble)
 == nil {
        return nil; }
    // inlining of macro CLOUD_ASSERT
    assert((f) >= dd.tUnique && (f) < dd.tUnique+dd.nNodesAlloc, "(f) >= dd.tUnique && (f) < dd.tUnique+dd.nNodesAlloc");
    // inlining of macro CLOUD_ASSERT
    assert((g) >= dd.tUnique && (g) < dd.tUnique+dd.nNodesAlloc, "(g) >= dd.tUnique && (g) < dd.tUnique+dd.nNodesAlloc");
    if dd.tCaches[CLOUD_OPER_AND] == nil {
        cloudCacheAllocate( dd, CLOUD_OPER_AND ); }
    // inlining of macro Cloud_Not
    t0 = cloudBddAnd_gate( dd, f, ((CloudNode*)(((ABC_PTRUINT_T)(g)) ^  01))   // complement the node
 );
    if t0 == nil {
        return nil; }
    // inlining of macro Cloud_Not
    t1 = cloudBddAnd_gate( dd, ((CloudNode*)(((ABC_PTRUINT_T)(f)) ^  01))   // complement the node
, g );
    if t1 == nil {
        return nil; }
    r  = Cloud_bddOr( dd, t0, t1 );
    return r;
}



/**Function********************************************************************

  Synopsis    [Performs a DFS from f, clearing the LSB of the next
  pointers.]

  Description []

  SideEffects [None]

  SeeAlso     [cloudSupport cloudDagSize]

******************************************************************************/
func cloudClearMark(dd *CloudManager, n *CloudNode) {
    // inlining of macro cloudNodeIsMarked
    if !((int32(((n).v &  CLOUD_MARK_ON)) != 0)) {
        return; }
    // clear visited flag
    // inlining of macro cloudNodeUnmark
    ((n).v       &= CLOUD_MARK_OFF);
    // inlining of macro cloudIsConstant
    if  ((n).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX {
        return; }
    // inlining of macro cloudT
    cloudClearMark( dd, ((n).t) );
    // inlining of macro Cloud_Regular
    // inlining of macro cloudE
    cloudClearMark( dd, ((CloudNode*)(((ABC_PTRUINT_T)(((n).e))) & ~01))   // get the regular node (w/o bubble)
 );
}

/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cloud_Support.]

  Description [Performs the recursive step of Cloud_Support. Performs a
  DFS from f. The support is accumulated in supp as a side effect. Uses
  the LSB of the then pointer as visited flag.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cloudSupport(dd *CloudManager, n *CloudNode, support *int32) {
    // inlining of macro cloudIsConstant
    // inlining of macro cloudNodeIsMarked
    if  (((n).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX) || ((int32(((n).v &  CLOUD_MARK_ON)) != 0)) {
        return; }
    // set visited flag
    // inlining of macro cloudNodeMark
    ((n).v       |= CLOUD_MARK_ON);
    // inlining of macro cloudV
    support[((n).v)] = 1;
    // inlining of macro cloudT
    cloudSupport( dd, ((n).t), support );
    // inlining of macro Cloud_Regular
    // inlining of macro cloudE
    cloudSupport( dd, ((CloudNode*)(((ABC_PTRUINT_T)(((n).e))) & ~01))   // get the regular node (w/o bubble)
, support );
}

/**Function********************************************************************

  Synopsis    [Finds the variables on which a DD depends.]

  Description [Finds the variables on which a DD depends.
  Returns a BDD consisting of the product of the variables if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cloud_Support(dd *CloudManager, n *CloudNode) *CloudNode {
    var res *CloudNode;
    var support *int32;
    var i int32;

    // inlining of macro CLOUD_ASSERT
    assert((n) >= dd.tUnique && (n) < dd.tUnique+dd.nNodesAlloc, "(n) >= dd.tUnique && (n) < dd.tUnique+dd.nNodesAlloc");

    // allocate and initialize support array for cloudSupport
    // inlining of macro ABC_CALLOC
    support = ((int *) calloc((size_t)(dd.nVars), sizeof(int)));

    // compute support and clean up markers
    // inlining of macro Cloud_Regular
    cloudSupport( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
, support );
    // inlining of macro Cloud_Regular
    cloudClearMark( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
 );

    // transform support from array to cube
    res = dd.one;
    for  i = dd.nVars - 1; i >= 0; i--  { // for each level bottom-up 
        if support[i] == 1 {
            res = Cloud_bddAnd( dd, res, dd.vars[i] );
            if res == nil {
                break; }
        }
    }
    // inlining of macro ABC_FREE
    ((support) ? (free((char *) (support)), (support) = 0) : 0);
    return res;
}

/**Function********************************************************************

  Synopsis    [Counts the variables on which a DD depends.]

  Description [Counts the variables on which a DD depends.
  Returns the number of the variables if successful; Cloud_OUT_OF_MEM
  otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cloud_SupportSize(dd *CloudManager, n *CloudNode) int32 {
    var support *int32;
    var i int32;
    var count int32;

    // inlining of macro CLOUD_ASSERT
    assert((n) >= dd.tUnique && (n) < dd.tUnique+dd.nNodesAlloc, "(n) >= dd.tUnique && (n) < dd.tUnique+dd.nNodesAlloc");

    // allocate and initialize support array for cloudSupport
    // inlining of macro ABC_CALLOC
    support = ((int *) calloc((size_t)(dd.nVars), sizeof(int)));

    // compute support and clean up markers
    // inlining of macro Cloud_Regular
    cloudSupport( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
, support );
    // inlining of macro Cloud_Regular
    cloudClearMark( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
 );

    // count support variables
    count = 0;
    for  i = 0; i < dd.nVars; i++  {
        if support[i] == 1 {
            count++; }
    }

    // inlining of macro ABC_FREE
    ((support) ? (free((char *) (support)), (support) = 0) : 0);
    return count;
}


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cloud_DagSize.]

  Description [Performs the recursive step of Cloud_DagSize. Returns the
  number of nodes in the graph rooted at n.]

  SideEffects [None]

******************************************************************************/
func cloudDagSize(dd *CloudManager, n *CloudNode) int32 {
    var tval, eval int32;
    // inlining of macro cloudNodeIsMarked
    if  int32(((n).v &  CLOUD_MARK_ON)) != 0 {
        return 0; }
    // set visited flag
    // inlining of macro cloudNodeMark
    ((n).v       |= CLOUD_MARK_ON);
    // inlining of macro cloudIsConstant
    if  ((n).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX {
        return 1; }
    // inlining of macro cloudT
    tval = cloudDagSize( dd, ((n).t) );
    // inlining of macro Cloud_Regular
    // inlining of macro cloudE
    eval = cloudDagSize( dd, ((CloudNode*)(((ABC_PTRUINT_T)(((n).e))) & ~01))   // get the regular node (w/o bubble)
 );
    return tval + eval + 1;

}

/**Function********************************************************************

  Synopsis    [Counts the number of nodes in a DD.]

  Description [Counts the number of nodes in a DD. Returns the number
  of nodes in the graph rooted at node.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_DagSize(dd *CloudManager, n *CloudNode) int32 {
    var res int32;
    // inlining of macro Cloud_Regular
    res = cloudDagSize( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
 );
    // inlining of macro Cloud_Regular
    cloudClearMark( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
 );
    return res;

}


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cloud_DagSize.]

  Description [Performs the recursive step of Cloud_DagSize. Returns the
  number of nodes in the graph rooted at n.]

  SideEffects [None]

******************************************************************************/
func Cloud_DagCollect_rec(dd *CloudManager, n *CloudNode, pCounter *int32) int32 {
    var tval, eval int32;
    // inlining of macro cloudNodeIsMarked
    if  int32(((n).v &  CLOUD_MARK_ON)) != 0 {
        return 0; }
    // set visited flag
    // inlining of macro cloudNodeMark
    ((n).v       |= CLOUD_MARK_ON);
    // inlining of macro cloudIsConstant
    if  ((n).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX {
        dd.ppNodes[(*pCounter)++] = n;
        return 1;
    }
    // inlining of macro cloudT
    tval = Cloud_DagCollect_rec( dd, ((n).t), pCounter );
    // inlining of macro Cloud_Regular
    // inlining of macro cloudE
    eval = Cloud_DagCollect_rec( dd, ((CloudNode*)(((ABC_PTRUINT_T)(((n).e))) & ~01))   // get the regular node (w/o bubble)
, pCounter );
    dd.ppNodes[(*pCounter)++] = n;
    return tval + eval + 1;

}

/**Function********************************************************************

  Synopsis    [Counts the number of nodes in a DD.]

  Description [Counts the number of nodes in a DD. Returns the number
  of nodes in the graph rooted at node.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_DagCollect(dd *CloudManager, n *CloudNode) int32 {
    var res, Counter int32;
    if dd.ppNodes == nil {
        // inlining of macro ABC_ALLOC
        dd.ppNodes = ((CloudNode * *) malloc(sizeof(CloudNode *) * (size_t)(dd.nNodesLimit))); }
    // inlining of macro Cloud_Regular
    res = Cloud_DagCollect_rec( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
, &Counter );
    // inlining of macro Cloud_Regular
    cloudClearMark( dd, ((CloudNode*)(((ABC_PTRUINT_T)(n)) & ~01))   // get the regular node (w/o bubble)
 );
    assert(res == Counter, "res == Counter");
    return res;

}

/**Function********************************************************************

  Synopsis    [Counts the number of nodes in an array of DDs.]

  Description [Counts the number of nodes in a DD. Returns the number
  of nodes in the graph rooted at node.]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_SharingSize(dd *CloudManager, pn **CloudNode, nn int32) int32 {
    var res, i int32;
    res = 0;
    for  i = 0; i < nn; i++  {
        // inlining of macro Cloud_Regular
        res += cloudDagSize( dd, ((CloudNode*)(((ABC_PTRUINT_T)(pn[i])) & ~01))   // get the regular node (w/o bubble)
 );
    }
    for  i = 0; i < nn; i++  {
        // inlining of macro Cloud_Regular
        cloudClearMark( dd, ((CloudNode*)(((ABC_PTRUINT_T)(pn[i])) & ~01))   // get the regular node (w/o bubble)
 );
    }
    return res;
}


/**Function********************************************************************

  Synopsis    [Returns one cube contained in the given BDD.]

  Description []

  SideEffects []

******************************************************************************/
func Cloud_GetOneCube(dd *CloudManager, bFunc *CloudNode) *CloudNode {
    var bFunc0, bFunc1, res *CloudNode;

    // inlining of macro Cloud_IsConstant
    // inlining of macro Cloud_Regular
    if  ((((CloudNode*)(((ABC_PTRUINT_T)(bFunc)) & ~01))   // get the regular node (w/o bubble)
).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX {
        return bFunc; }

    // cofactor
    // inlining of macro Cloud_IsComplement
    if  int32((((ABC_PTRUINT_T)(bFunc)) & 01)) != 0 {           // check if complemented
  
        // inlining of macro Cloud_Not
        // inlining of macro cloudE
        bFunc0 = ((CloudNode*)(((ABC_PTRUINT_T)(((bFunc).e))) ^  01))   // complement the node
;
        // inlining of macro Cloud_Not
        // inlining of macro cloudT
        bFunc1 = ((CloudNode*)(((ABC_PTRUINT_T)(((bFunc).t))) ^  01))   // complement the node
;
    
    } else {
        // inlining of macro cloudE
        bFunc0 = ((bFunc).e);
        // inlining of macro cloudT
        bFunc1 = ((bFunc).t);
    }

    // try to find the cube with the negative literal
    res = Cloud_GetOneCube( dd, bFunc0 );
    if res == nil {
        return nil; }

    if res != dd.zero {
        // inlining of macro Cloud_Not
        // inlining of macro Cloud_V
        // inlining of macro Cloud_Regular
        res = Cloud_bddAnd( dd, res, ((CloudNode*)(((ABC_PTRUINT_T)(dd.vars[((((CloudNode*)(((ABC_PTRUINT_T)(bFunc)) & ~01))   // get the regular node (w/o bubble)
).v)])) ^  01))   // complement the node
 );
    
    } else {
        // try to find the cube with the positive literal
        res = Cloud_GetOneCube( dd, bFunc1 );
        if res == nil {
            return nil; }
        assert(res != dd.zero, "res != dd.zero");
        // inlining of macro Cloud_V
        // inlining of macro Cloud_Regular
        res = Cloud_bddAnd( dd, res, dd.vars[((((CloudNode*)(((ABC_PTRUINT_T)(bFunc)) & ~01))   // get the regular node (w/o bubble)
).v)] );
    }
    return res;
}

/**Function********************************************************************

  Synopsis    [Prints the BDD as a set of disjoint cubes to the standard output.]

  Description []

  SideEffects []

******************************************************************************/
func Cloud_bddPrint(dd *CloudManager, Func *CloudNode) {
    var Cube *CloudNode;
    var fFirst int32  = 1;

    if Func == dd.zero {
        fmt.Printf("Constant 0.");
    } else if Func == dd.one {
        fmt.Printf("Constant 1.");
    } else {
        for {
            Cube = Cloud_GetOneCube( dd, Func );
            if Cube == nil || Cube == dd.zero {
                break; }
            if fFirst != 0 {       fFirst = 0;
            } else {                fmt.Printf(" + "); }
            Cloud_bddPrintCube( dd, Cube );
            // inlining of macro Cloud_Not
            Func = Cloud_bddAnd( dd, Func, ((CloudNode*)(((ABC_PTRUINT_T)(Cube)) ^  01))   // complement the node
 );
        } 
    }
    fmt.Printf("\n");
}

/**Function********************************************************************

  Synopsis    [Prints one cube.]

  Description []

  SideEffects []

******************************************************************************/
func Cloud_bddPrintCube(dd *CloudManager, bCube *CloudNode) {
    var bCube0, bCube1 *CloudNode;

    // inlining of macro Cloud_IsConstant
    // inlining of macro Cloud_Regular
    assert(!(((((CloudNode*)(((ABC_PTRUINT_T)(bCube)) & ~01))   // get the regular node (w/o bubble)
    ).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX), "!/*__starting_point_inline_macro:Cloud_IsConstant__*/(((/*__starting_point_inline_macro:Cloud_Regular__*/((CloudNode*)(((ABC_PTRUINT_T)(bCube)) & ~01))   // get the regular node (w/o bubble)\n).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX)");
    for {
        // get the node structure
        // inlining of macro Cloud_IsConstant
        // inlining of macro Cloud_Regular
        if  ((((CloudNode*)(((ABC_PTRUINT_T)(bCube)) & ~01))   // get the regular node (w/o bubble)
).v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX {
            break; }

        // cofactor the cube
        // inlining of macro Cloud_IsComplement
        if  int32((((ABC_PTRUINT_T)(bCube)) & 01)) != 0 {           // check if complemented
  
            // inlining of macro Cloud_Not
            // inlining of macro cloudE
            bCube0 = ((CloudNode*)(((ABC_PTRUINT_T)(((bCube).e))) ^  01))   // complement the node
;
            // inlining of macro Cloud_Not
            // inlining of macro cloudT
            bCube1 = ((CloudNode*)(((ABC_PTRUINT_T)(((bCube).t))) ^  01))   // complement the node
;
        
        } else {
            // inlining of macro cloudE
            bCube0 = ((bCube).e);
            // inlining of macro cloudT
            bCube1 = ((bCube).t);
        }

        if bCube0 != dd.zero {
            assert(bCube1 == dd.zero, "bCube1 == dd.zero");
            // inlining of macro cloudV
            fmt.Printf("[%d]'", ((bCube).v));
            bCube = bCube0;
        
        } else {
            assert(bCube1 != dd.zero, "bCube1 != dd.zero");
            // inlining of macro cloudV
            fmt.Printf("[%d]", ((bCube).v));
            bCube = bCube1;
        }
    }
}


/**Function********************************************************************

  Synopsis    [Prints info.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_PrintInfo(dd *CloudManager) {
    if dd == nil { return; }
    fmt.Printf("The number of unique table nodes allocated = %12d.\n", dd.nNodesAlloc);
    fmt.Printf("The number of unique table nodes present   = %12d.\n", dd.nNodesCur);
    fmt.Printf("The number of unique table hits            = %12d.\n", dd.nUniqueHits);
    fmt.Printf("The number of unique table misses          = %12d.\n", dd.nUniqueMisses);
    fmt.Printf("The number of unique table steps           = %12d.\n", dd.nUniqueSteps);
    fmt.Printf("The number of cache hits                   = %12d.\n", dd.nCacheHits);
    fmt.Printf("The number of cache misses                 = %12d.\n", dd.nCacheMisses);
    fmt.Printf("The current signature                      = %12d.\n", dd.nSignCur);
    fmt.Printf("The total memory in use                    = %12d.\n", dd.nMemUsed);
}

/**Function********************************************************************

  Synopsis    [Prints the state of the hash table.]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Cloud_PrintHashTable(dd *CloudManager) {
    var i int32;

    for  i = 0; i < dd.nNodesAlloc; i++  {
        if dd.tUnique[i].v == CLOUD_CONST_INDEX {
            fmt.Printf("-");
        } else {
            fmt.Printf("+"); }
    }
    fmt.Printf("\n");
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

