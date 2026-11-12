package kit

/**CFile****************************************************************

  FileName    [cloud.h]

  PackageName [Fast application-specific BDD package.]

  Synopsis    [Interface of the package.]

  Author      [Alan Mishchenko <alanmi@ece.pdx.edu>]
  
  Affiliation [ECE Department. Portland State University, Portland, Oregon.]

  Date        [Ver. 1.0. Started - June 10, 2002.]

  Revision    [$Id: cloud.h,v 1.0 2002/06/10 03:00:00 alanmi Exp $]

***********************************************************************/





//#include <stdio.h>
//#include <stdlib.h>
//#include <assert.h>

//#include "misc/util/abc_global.h"



//ABC_NAMESPACE_HEADER_START


#ifdef _WIN32
#define inline __inline // compatible with MS VS 6.0
#endif

////////////////////////////////////////////////////////////////////////
//  n |  2^n ||  n |    2^n ||  n |        2^n ||  n |           2^n  //
//====================================================================//
//  1 |    2 ||  9 |    512 || 17 |    131,072 || 25 |    33,554,432  //
//  2 |    4 || 10 |  1,024 || 18 |    262,144 || 26 |    67,108,864  // 
//  3 |    8 || 11 |  2,048 || 19 |    524,288 || 27 |   134,217,728  //
//  4 |   16 || 12 |  4,096 || 20 |  1,048,576 || 28 |   268,435,456  //
//  5 |   32 || 13 |  8,192 || 21 |  2,097,152 || 29 |   536,870,912  //
//  6 |   64 || 14 | 16,384 || 22 |  4,194,304 || 30 | 1,073,741,824  //
//  7 |  128 || 15 | 32,768 || 23 |  8,388,608 || 31 | 2,147,483,648  //
//  8 |  256 || 16 | 65,536 || 24 | 16,777,216 || 32 | 4,294,967,296  //
////////////////////////////////////////////////////////////////////////

// data structure typedefs
type CloudManager =       cloudManager;
typedef unsigned                  CloudVar;
typedef unsigned                  CloudSign;
type CloudNode =          cloudNode;
type CloudCacheEntry1 =   cloudCacheEntry1;
type CloudCacheEntry2 =   cloudCacheEntry2;
type CloudCacheEntry3 =   cloudCacheEntry3;

// operation codes used to set up the cache
 type CloudOper int32
const ( 
    CLOUD_OPER_AND CloudOper = iota
 
    CLOUD_OPER_XOR
 
    CLOUD_OPER_BDIFF
 
    CLOUD_OPER_LEQ 

)

 

/*
// the number of operators using cache
static int CacheOperNum = 4;

// the ratio of cache size to the unique table size for each operator
static int CacheLogRatioDefault[4] = {
    4, // CLOUD_OPER_AND, 
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
*/

// data structure definitions
type cloudManager struct {            // the fast bdd manager     

    // variables
    nVars int32;                 // the number of variables allocated
    // bits
    bitsNode int32;              // the number of bits used for the node
    bitsCache [4]int32;          // default: bitsNode - CacheSizeRatio[i]
    // shifts
    shiftUnique int32;           // 8*sizeof(unsigned) - (bitsNode + 1)
    shiftCache [4]int32;         // 8*sizeof(unsigned) -  bitsCache[i]
    // nodes 
    nNodesAlloc int32;           // 2 ^ (bitsNode + 1)
    nNodesLimit int32;           // 2 ^  bitsNode
    nNodesCur int32;             // the current number of nodes (including const1 and vars)
    // signature
    nSignCur CloudSign;

    // statistics
    nMemUsed int32;              // memory usage in bytes
    // cache stats
    nUniqueHits int32;           // hits in the unique table
    nUniqueMisses int32;         // misses in the unique table
    nCacheHits int32;            // hits in the caches
    nCacheMisses int32;          // misses in the caches
    // the number of steps through the hash table
    nUniqueSteps int32;

    // tables
    tUnique  *CloudNode;       // the unique table to store BDD nodes

    // special nodes
    pNodeStart  *CloudNode;    // the pointer to the first node
    pNodeEnd  *CloudNode;      // the pointer to the first node out of the table

    // constants and variables
    one   *CloudNode;          // the one function
    zero   *CloudNode;         // the zero function
    vars  **CloudNode;         // the elementary variables

    // temporary storage for nodes
    ppNodes  **CloudNode;

    // caches
    tCaches [20]*CloudCacheEntry2;    // caches
}

type cloudNode struct {   // representation of the node in the unique table

    s   CloudSign;         // signature
    v    CloudVar;         // variable
    e  *CloudNode;         // negative cofactor
    t  *CloudNode;         // positive cofactor
}
type cloudCacheEntry1 struct {  // one-argument cache

    s   CloudSign;         // signature
    a  *CloudNode;         // argument 1
    r  *CloudNode;         // result
}
type cloudCacheEntry2 struct {  // the two-argument cache

    s   CloudSign;         // signature
    a  *CloudNode;
    b  *CloudNode;
    r  *CloudNode;
}
type cloudCacheEntry3 struct {  // the three-argument cache

    s   CloudSign;         // signature
    a  *CloudNode;
    b  *CloudNode;
    c  *CloudNode;
    r  *CloudNode;
}


// parameters
const CLOUD_NODE_BITS =              23

#define CLOUD_CONST_INDEX            ((unsigned)0x0fffffff)
#define CLOUD_MARK_ON                ((unsigned)0x10000000)
#define CLOUD_MARK_OFF               ((unsigned)0xefffffff)

// hash functions a la Buddy
// Original macro:
// #define cloudHashBuddy2(x,y,s)       ((((x)+(y))*((x)+(y)+1)/2) & ((1<<(32-(s)))-1))

// Original macro:
// #define cloudHashBuddy3(x,y,z,s)     (cloudHashBuddy2((cloudHashBuddy2((x),(y),(s))),(z),(s)) & ((1<<(32-(s)))-1))

// hash functions a la Cudd
const DD_P1 =                           12582917
const DD_P2 =                         4256249
const DD_P3 =                         741457
const DD_P4 =                         1618033999
// Original macro:
// #define cloudHashCudd2(f,g,s)        ((((unsigned)(ABC_PTRUINT_T)(f) * DD_P1 + (unsigned)(ABC_PTRUINT_T)(g)) * DD_P2) >> (s))

// Original macro:
// #define cloudHashCudd3(f,g,h,s)      (((((unsigned)(ABC_PTRUINT_T)(f) * DD_P1 + (unsigned)(ABC_PTRUINT_T)(g)) * DD_P2 + (unsigned)(ABC_PTRUINT_T)(h)) * DD_P3) >> (s))


// node complementation (using node)
// Original macro:
// #define Cloud_Regular(p)             ((CloudNode*)(((ABC_PTRUINT_T)(p)) & ~01))   // get the regular node (w/o bubble)

// Original macro:
// #define Cloud_Not(p)                 ((CloudNode*)(((ABC_PTRUINT_T)(p)) ^  01))   // complement the node

// Original macro:
// #define Cloud_NotCond(p,c)           ((CloudNode*)(((ABC_PTRUINT_T)(p)) ^ (c)))   // complement the node conditionally

// Original macro:
// #define Cloud_IsComplement(p)        ((int)(((ABC_PTRUINT_T)(p)) & 01))           // check if complemented

// checking constants (using node)
// Original macro:
// #define Cloud_IsConstant(p)          (((Cloud_Regular(p))->v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX)

// Original macro:
// #define cloudIsConstant(p)           (((p)->v & CLOUD_MARK_OFF) == CLOUD_CONST_INDEX)


// retrieving values from the node (using node structure)
// Original macro:
// #define Cloud_V(p)                   ((Cloud_Regular(p))->v)

// Original macro:
// #define Cloud_E(p)                   ((Cloud_Regular(p))->e)

// Original macro:
// #define Cloud_T(p)                   ((Cloud_Regular(p))->t)

// retrieving values from the regular node (using node structure)
// Original macro:
// #define cloudV(p)                    ((p)->v)

// Original macro:
// #define cloudE(p)                    ((p)->e)

// Original macro:
// #define cloudT(p)                    ((p)->t)

// marking/unmarking (using node structure)
// Original macro:
// #define cloudNodeMark(p)             ((p)->v       |= CLOUD_MARK_ON)

// Original macro:
// #define cloudNodeUnmark(p)           ((p)->v       &= CLOUD_MARK_OFF)

// Original macro:
// #define cloudNodeIsMarked(p)         ((int)((p)->v &  CLOUD_MARK_ON))


// cache lookups and inserts (using node)
// Original macro:
// #define cloudCacheLookup1(p,sign,f)     (((p)->s == (sign) && (p)->a == (f))? ((p)->r): (0))

// Original macro:
// #define cloudCacheLookup2(p,sign,f,g)   (((p)->s == (sign) && (p)->a == (f) && (p)->b == (g))? ((p)->r): (0))

// Original macro:
// #define cloudCacheLookup3(p,sign,f,g,h) (((p)->s == (sign) && (p)->a == (f) && (p)->b == (g) && (p)->c == (h))? ((p)->r): (0))

// cache inserts
// Original macro:
// #define cloudCacheInsert1(p,sign,f,r)     (((p)->s = (sign)), ((p)->a = (f)), ((p)->r = (r)))

// Original macro:
// #define cloudCacheInsert2(p,sign,f,g,r)   (((p)->s = (sign)), ((p)->a = (f)), ((p)->b = (g)), ((p)->r = (r)))

// Original macro:
// #define cloudCacheInsert3(p,sign,f,g,h,r) (((p)->s = (sign)), ((p)->a = (f)), ((p)->b = (g)), ((p)->c = (h)), ((p)->r = (r)))


//#define CLOUD_ASSERT(p)              (assert((p) >= (dd->pNodeStart-1) && (p) < dd->pNodeEnd))
// Original macro:
// #define CLOUD_ASSERT(p)            assert((p) >= dd->tUnique && (p) < dd->tUnique+dd->nNodesAlloc)


////////////////////////////////////////////////////////////////////////
///                       FUNCTION DECLARATIONS                      ///
////////////////////////////////////////////////////////////////////////
// starting/stopping 
// extern CloudManager * Cloud_Init( int nVars, int nBits );

// extern void           Cloud_Quit( CloudManager * dd );

// extern void           Cloud_Restart( CloudManager * dd );

// extern void           Cloud_CacheAllocate( CloudManager * dd, CloudOper oper, int size );

// extern CloudNode *    Cloud_MakeNode( CloudManager * dd, CloudVar v, CloudNode * t, CloudNode * e );

// support and node count
// extern CloudNode *    Cloud_Support( CloudManager * dd, CloudNode * n );

// extern int            Cloud_SupportSize( CloudManager * dd, CloudNode * n );

// extern int            Cloud_DagSize( CloudManager * dd, CloudNode * n );

// extern int            Cloud_DagCollect( CloudManager * dd, CloudNode * n );

// extern int            Cloud_SharingSize( CloudManager * dd, CloudNode * * pn, int nn );

// cubes
// extern CloudNode *    Cloud_GetOneCube( CloudManager * dd, CloudNode * n );

// extern void           Cloud_bddPrint( CloudManager * dd, CloudNode * Func );

// extern void           Cloud_bddPrintCube( CloudManager * dd, CloudNode * Cube );

// operations
// extern CloudNode *    Cloud_bddAnd( CloudManager * dd, CloudNode * f, CloudNode * g );

// extern CloudNode *    Cloud_bddOr( CloudManager * dd, CloudNode * f, CloudNode * g );

// stats 
// extern void           Cloud_PrintInfo( CloudManager * dd );

// extern void           Cloud_PrintHashTable( CloudManager * dd );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                            END OF FILE                           ///
////////////////////////////////////////////////////////////////////////
