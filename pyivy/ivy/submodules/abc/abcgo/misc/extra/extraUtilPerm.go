package extra

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [extraUtilPerm.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Permutation computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilPerm.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include "misc/vec/vec.h"
//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

 type Abc_ZddOper int32
const  
(
    ABC_ZDD_OPER_NONE Abc_ZddOper = iota

    ABC_ZDD_OPER_DIFF

    ABC_ZDD_OPER_UNION

    ABC_ZDD_OPER_MIN_UNION

    ABC_ZDD_OPER_INTER

    ABC_ZDD_OPER_PERM

    ABC_ZDD_OPER_PERM_PROD

    ABC_ZDD_OPER_COF0

    ABC_ZDD_OPER_COF1

    ABC_ZDD_OPER_THRESH

    ABC_ZDD_OPER_DOT_PROD

    ABC_ZDD_OPER_DOT_PROD_6

    ABC_ZDD_OPER_INSERT

    ABC_ZDD_OPER_PATHS

    ABC_ZDD_OPER_NODES

    ABC_ZDD_OPER_ITE

)

 

   
type Abc_ZddObj struct {
    Var uint32; /* C bitfield: 31 bits; review packing and truncation */
    Mark uint32; /* C bitfield: 1 bits; review packing and truncation */
    True     uint32;
    False     uint32;
}  

   
type Abc_ZddEnt struct {
    Arg0          int32;
    Arg1          int32;
    Arg2          int32;
    Res          int32;
}

   
type Abc_ZddMan struct {
    nVars          int32;
    nObjs          int32;
    nObjsAlloc          int32;
    nPermSize          int32;
    nUniqueMask     uint32;
    nCacheMask     uint32;
    pUnique         *int32;
    pNexts         *int32;
    pCache  *Abc_ZddEnt;
    pObjs  *Abc_ZddObj;
    nCacheLookups          int32;
    nCacheMisses          int32;
    nMemory         word;
    pV2TI         *int32;
    pV2TJ         *int32;
    pT2V         *int32;
}

func Abc_ZddIthVar(i int32) int32 { return i + 2;                                            }
func Abc_ZddHash(Arg0 int32, Arg1 int32, Arg2 int32) uint32 { return 12582917 * Arg0 + 4256249 * Arg1 + 741457 * Arg2; }

func Abc_ZddNode(p *Abc_ZddMan, i int32) *Abc_ZddObj { return p.pObjs + i;                                     }
func Abc_ZddObjId(p *Abc_ZddMan, pObj *Abc_ZddObj) int32 { return pObj - p.pObjs;                                  }
func Abc_ZddObjVar(p *Abc_ZddMan, i int32) int32 { return Abc_ZddNode(p, i).Var;                           }

func Abc_ZddSetVarIJ(p *Abc_ZddMan, i int32, j int32, v int32) { assert(i < j, "i < j"); p.pT2V[i * p.nPermSize + j] = v;      }
func Abc_ZddVarIJ(p *Abc_ZddMan, i int32, j int32) int32 { assert(i < j, "i < j"); return p.pT2V[i * p.nPermSize + j];   }
func Abc_ZddVarsClash(p *Abc_ZddMan, v0 int32, v1 int32) int32 { return p.pV2TI[v0] == p.pV2TI[v1] || p.pV2TJ[v0] == p.pV2TJ[v1] || p.pV2TI[v0] == p.pV2TJ[v1] || p.pV2TJ[v0] == p.pV2TI[v1];  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddCacheLookup(p *Abc_ZddMan, Arg0 int32, Arg1 int32, Arg2 int32) int32 {
    var pEnt *Abc_ZddEnt  = p.pCache + (Abc_ZddHash(Arg0, Arg1, Arg2) & p.nCacheMask);
    p.nCacheLookups++;
    if pEnt.Arg0 == Arg0 && pEnt.Arg1 == Arg1 && pEnt.Arg2 == Arg2 {
        return pEnt.Res;
    } else {
        return -1;
    }
}
func Abc_ZddCacheInsert(p *Abc_ZddMan, Arg0 int32, Arg1 int32, Arg2 int32, Res int32) int32 {
    var pEnt *Abc_ZddEnt  = p.pCache + (Abc_ZddHash(Arg0, Arg1, Arg2) & p.nCacheMask);
    pEnt.Arg0 = Arg0;  pEnt.Arg1 = Arg1;  pEnt.Arg2 = Arg2;  pEnt.Res = Res;
    p.nCacheMisses++;
    assert(Res >= 0, "Res >= 0");
    return Res;
}
func Abc_ZddUniqueLookup(p *Abc_ZddMan, Var int32, True int32, False int32) int32 {
    var q *int32  = p.pUnique + (Abc_ZddHash(Var, True, False) & p.nUniqueMask);
    for  ; *q; q = p.pNexts + *q  {
        if p.pObjs[*q].Var == uint32(Var) && p.pObjs[*q].True == uint32(True) && p.pObjs[*q].False == uint32(False) {
            return *q; }
    }
    return 0;
}
func Abc_ZddUniqueCreate(p *Abc_ZddMan, Var int32, True int32, False int32) int32 {
    assert(Var >= 0 && Var < p.nVars, "Var >= 0 && Var < p.nVars");
    assert(Var < Abc_ZddObjVar(p, True), "Var < Abc_ZddObjVar(p, True)");
    assert(Var < Abc_ZddObjVar(p, False), "Var < Abc_ZddObjVar(p, False)");
    if True == 0 {
        return False; }
    {
        var q *int32  = p.pUnique + (Abc_ZddHash(Var, True, False) & p.nUniqueMask);
        for  ; *q; q = p.pNexts + *q  {
            if p.pObjs[*q].Var == uint32(Var) && p.pObjs[*q].True == uint32(True) && p.pObjs[*q].False == uint32(False) {
                return *q; }
        }
        if p.nObjs == p.nObjsAlloc {
            fmt.Printf("Aborting because the number of nodes exceeded %d.\n", p.nObjsAlloc);
fflush(os.Stdout); }
        assert(p.nObjs < p.nObjsAlloc, "p.nObjs < p.nObjsAlloc");     
        *q = p.nObjs++;
        p.pObjs[*q].Var = Var;
        p.pObjs[*q].True = True;
        p.pObjs[*q].False = False;
//        printf( "Added node %3d: Var = %3d.  True = %3d.  False = %3d\n", *q, Var, True, False );
        return *q;
    }
}
func Abc_ZddBuildSet(p *Abc_ZddMan, pSet *int32, Size int32) int32 {
    var i int32;
    var Res int32 = 1;
    Vec_IntSelectSort( pSet, Size );
    for  i = Size - 1; i >= 0; i--  {
        Res = Abc_ZddUniqueCreate( p, pSet[i], Res, 0 );
    }
    return Res;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddManAlloc(nVars int32, nObjs int32) *Abc_ZddMan {
     var i int32;
    // inlining of macro ABC_CALLOC
    p := &Abc_ZddMan{};
    p.nVars       = nVars;
    p.nObjsAlloc  = nObjs;
    p.nUniqueMask = (1 << Abc_Base2Log(nObjs)) - 1;
    p.nCacheMask  = (1 << Abc_Base2Log(nObjs)) - 1;
    // inlining of macro ABC_CALLOC
    p.pUnique     = ((int *) calloc((size_t)(p.nUniqueMask + 1), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pNexts      = ((int *) calloc((size_t)(p.nObjsAlloc), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pCache      = ((Abc_ZddEnt *) calloc((size_t)(p.nCacheMask + 1), sizeof(Abc_ZddEnt)));
    // inlining of macro ABC_CALLOC
    p.pObjs       = ((Abc_ZddObj *) calloc((size_t)(p.nObjsAlloc), sizeof(Abc_ZddObj)));
    p.nObjs       = 2;
    memset( p.pObjs, 0xff, sizeof(Abc_ZddObj) * 2 );
    p.pObjs[0].Var = nVars;
    p.pObjs[1].Var = nVars;
    for  i = 0; i < nVars; i++  {
        Abc_ZddUniqueCreate( p, i, 1, 0 );
    }
    assert(p.nObjs == nVars + 2, "p.nObjs == nVars + 2");
    p.nMemory = sizeof(Abc_ZddMan)/4 + 
        p.nUniqueMask + 1 + p.nObjsAlloc + 
        (p.nCacheMask + 1) * sizeof(Abc_ZddEnt)/4 + 
        p.nObjsAlloc * sizeof(Abc_ZddObj)/4;
    return p;
}
func Abc_ZddManCreatePerms(p *Abc_ZddMan, nPermSize int32) {
    var i, j, v int32;
    assert(2 * p.nVars == nPermSize * (nPermSize - 1), "2 * p.nVars == nPermSize * (nPermSize - 1)");
    assert(p.nPermSize == 0, "p.nPermSize == 0");
    p.nPermSize = nPermSize;
    // inlining of macro ABC_FALLOC
    p.pV2TI = ((int *) memset(malloc(sizeof(int) * (size_t)(p.nVars)), 0xff, sizeof(int) * (size_t)(p.nVars)));
    // inlining of macro ABC_FALLOC
    p.pV2TJ = ((int *) memset(malloc(sizeof(int) * (size_t)(p.nVars)), 0xff, sizeof(int) * (size_t)(p.nVars)));
    // inlining of macro ABC_FALLOC
    p.pT2V  = ((int *) memset(malloc(sizeof(int) * (size_t)(p.nPermSize * p.nPermSize)), 0xff, sizeof(int) * (size_t)(p.nPermSize * p.nPermSize)));
    for  i = 0; i < nPermSize; i++  {
        for  j = i + 1; j < nPermSize; j++  {
            p.pV2TI[v] = i;
            p.pV2TJ[v] = j;
            Abc_ZddSetVarIJ( p, i, j, v++ );
        }
    }
    assert(v == p.nVars, "v == p.nVars");
}
func Abc_ZddManFree(p *Abc_ZddMan) {
    fmt.Printf("ZDD stats: Var = %d  Obj = %d  Alloc = %d  Hit = %d  Miss = %d  ", p.nVars, p.nObjs, p.nObjsAlloc, p.nCacheLookups-p.nCacheMisses, p.nCacheMisses);
    fmt.Printf("Mem = %.2f MB\n", 4.0*int32((p.nMemory/(1<<20))));
    // inlining of macro ABC_FREE
    ((p.pT2V) ? (free((char *) (p.pT2V)), (p.pT2V) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pV2TI) ? (free((char *) (p.pV2TI)), (p.pV2TI) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pV2TJ) ? (free((char *) (p.pV2TJ)), (p.pV2TJ) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pUnique) ? (free((char *) (p.pUnique)), (p.pUnique) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pNexts) ? (free((char *) (p.pNexts)), (p.pNexts) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pCache) ? (free((char *) (p.pCache)), (p.pCache) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pObjs) ? (free((char *) (p.pObjs)), (p.pObjs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddDiff(p *Abc_ZddMan, a int32, b int32) int32 {
    var A, B *Abc_ZddObj; 
    var r0, r1, r int32;
    if a == 0 { return 0; }
    if b == 0 { return a; }
    if a == b { return 0; }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_DIFF)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    B = Abc_ZddNode( p, b );
    if A.Var < B.Var {
        r0 = Abc_ZddDiff( p, A.False, b );
r = Abc_ZddUniqueCreate( p, A.Var, A.True, r0 );
    } else if A.Var > B.Var {
        r = Abc_ZddDiff( p, a, B.False );
    } else {
        r0 = Abc_ZddDiff( p, A.False, B.False );
r1 = Abc_ZddDiff( p, A.True, B.True );
r = Abc_ZddUniqueCreate( p, A.Var, r1, r0 ); }
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_DIFF, r );
}
func Abc_ZddUnion(p *Abc_ZddMan, a int32, b int32) int32 {
    var A, B *Abc_ZddObj; 
    var r0, r1, r int32;
    if a == 0 { return b; }
    if b == 0 { return a; }
    if a == b { return a; }
    if a > b {  return Abc_ZddUnion( p, b, a ); }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_UNION)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    B = Abc_ZddNode( p, b );
    if A.Var < B.Var {
        r0 = Abc_ZddUnion( p, A.False, b );
r1 = A.True;
    } else if A.Var > B.Var {
        r0 = Abc_ZddUnion( p, a, B.False );
r1 = B.True;
    } else {
        r0 = Abc_ZddUnion( p, A.False, B.False );
r1 = Abc_ZddUnion( p, A.True, B.True ); }
    r = Abc_ZddUniqueCreate( p, Abc_MinInt(A.Var, B.Var), r1, r0 );
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_UNION, r );
}
func Abc_ZddMinUnion(p *Abc_ZddMan, a int32, b int32) int32 {
    var A, B *Abc_ZddObj; 
    var r0, r1, r int32;
    if a == 0 { return b; }
    if b == 0 { return a; }
    if a == b { return a; }
    if a > b {  return Abc_ZddMinUnion( p, b, a ); }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_MIN_UNION)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    B = Abc_ZddNode( p, b );
    if A.Var < B.Var {
        r0 = Abc_ZddMinUnion( p, A.False, b );
r1 = A.True;
    } else if A.Var > B.Var {
        r0 = Abc_ZddMinUnion( p, a, B.False );
r1 = B.True;
    } else {
        r0 = Abc_ZddMinUnion( p, A.False, B.False );
r1 = Abc_ZddMinUnion( p, A.True, B.True ); }
    r1 = Abc_ZddDiff( p, r1, r0 ); // assume args are minimal
    r = Abc_ZddUniqueCreate( p, Abc_MinInt(A.Var, B.Var), r1, r0 );
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_MIN_UNION, r );
}
func Abc_ZddIntersect(p *Abc_ZddMan, a int32, b int32) int32 {
    var A, B *Abc_ZddObj; 
    var r0, r1, r int32;
    if a == 0 { return 0; }
    if b == 0 { return 0; }
    if a == b { return a; }
    if a > b {  return Abc_ZddIntersect( p, b, a ); }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_INTER)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    B = Abc_ZddNode( p, b );
    if A.Var < B.Var {
        r0 = Abc_ZddIntersect( p, A.False, b );
r1 = A.True;
    } else if A.Var > B.Var {
        r0 = Abc_ZddIntersect( p, a, B.False );
r1 = B.True;
    } else {
        r0 = Abc_ZddIntersect( p, A.False, B.False );
r1 = Abc_ZddIntersect( p, A.True, B.True ); }
    r = Abc_ZddUniqueCreate( p, Abc_MinInt(A.Var, B.Var), r1, r0 );
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_INTER, r );
}
func Abc_ZddCof0(p *Abc_ZddMan, a int32, Var int32) int32 {
    var A *Abc_ZddObj; 
    var r0, r1, r int32;
    if a < 2 { return a; }
    A = Abc_ZddNode( p, a );
    if int32(A.Var) > Var {
        return a; }
    if (r = Abc_ZddCacheLookup(p, a, Var, ABC_ZDD_OPER_COF0)) >= 0 {
        return r; }
    if int32(A.Var) < Var {
        r0 = Abc_ZddCof0( p, A.False, Var );
r1 = Abc_ZddCof0( p, A.True, Var );
r  = Abc_ZddUniqueCreate( p, A.Var, r1, r0 );
    } else {
        r = Abc_ZddCof0( p, A.False, Var ); }
    return Abc_ZddCacheInsert( p, a, Var, ABC_ZDD_OPER_COF0, r );
}
func Abc_ZddCof1(p *Abc_ZddMan, a int32, Var int32) int32 {
    var A *Abc_ZddObj; 
    var r0, r1, r int32;
    if a < 2 { return a; }
    A = Abc_ZddNode( p, a );
    if int32(A.Var) > Var {
        return a; }
    if (r = Abc_ZddCacheLookup(p, a, Var, ABC_ZDD_OPER_COF1)) >= 0 {
        return r; }
    if int32(A.Var) < Var {
        r0 = Abc_ZddCof1( p, A.False, Var );
r1 = Abc_ZddCof1( p, A.True, Var );
    } else {
        r0 = 0;
r1 = Abc_ZddCof1( p, A.True, Var ); }
    r = Abc_ZddUniqueCreate( p, A.Var, r1, r0 );
    return Abc_ZddCacheInsert( p, a, Var, ABC_ZDD_OPER_COF1, r );
}
func Abc_ZddCountPaths(p *Abc_ZddMan, a int32) int32 {
    var A *Abc_ZddObj;  
    var r int32;
    if a < 2 { return a; }
    if (r = Abc_ZddCacheLookup(p, a, 0, ABC_ZDD_OPER_PATHS)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    r = Abc_ZddCountPaths( p, A.False ) + Abc_ZddCountPaths( p, A.True );
    return Abc_ZddCacheInsert( p, a, 0, ABC_ZDD_OPER_PATHS, r );
}
/*
int Abc_ZddCountNodes( Abc_ZddMan * p, int a )
{
    Abc_ZddObj * A;  
    int r;
    if ( a < 2 ) return 0;
    if ( (r = Abc_ZddCacheLookup(p, a, 0, ABC_ZDD_OPER_NODES)) >= 0 )
        return r;
    A = Abc_ZddNode( p, a );
    r = 1 + Abc_ZddCountNodes( p, A->False ) + Abc_ZddCountNodes( p, A->True );
    return Abc_ZddCacheInsert( p, a, 0, ABC_ZDD_OPER_NODES, r );
}
*/
func Abc_ZddCount_rec(p *Abc_ZddMan, i int32) int32 {
    var A *Abc_ZddObj;
    if i < 2 {
        return 0; }
    A = Abc_ZddNode( p, i );
    if A.Mark {
        return 0; }
    A.Mark = 1;
    return 1 + Abc_ZddCount_rec(p, A.False) + Abc_ZddCount_rec(p, A.True);
}
func Abc_ZddUnmark_rec(p *Abc_ZddMan, i int32) {
    var A *Abc_ZddObj;
    if i < 2 {
        return; }
    A = Abc_ZddNode( p, i );
    if !A.Mark {
        return; }
    A.Mark = 0;
    Abc_ZddUnmark_rec( p, A.False );
    Abc_ZddUnmark_rec( p, A.True );
}
func Abc_ZddCountNodes(p *Abc_ZddMan, i int32) int32 {
    var Count int32  = Abc_ZddCount_rec( p, i );
    Abc_ZddUnmark_rec( p, i );
    return Count;
}
func Abc_ZddCountNodesArray(p *Abc_ZddMan, vNodes *Vec_Int_t) int32 {
    var i, Id, Count int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((Id) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Count += Abc_ZddCount_rec( p, Id );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((Id) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Abc_ZddUnmark_rec( p, Id );
    }
    return Count;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddThresh(p *Abc_ZddMan, a int32, b int32) int32 {
    var A *Abc_ZddObj; 
    var r0, r1, r int32;
    if a < 2 {  return a; }
    if b == 0 { return 0; }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_THRESH)) >= 0 {
        return r; }
    A  = Abc_ZddNode( p, a );
    r0 = Abc_ZddThresh( p, A.False, b );
r1 = Abc_ZddThresh( p, A.True, b-1 );
    r  = Abc_ZddUniqueCreate( p, A.Var, r1, r0 );
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_THRESH, r );
}
func Abc_ZddDotProduct(p *Abc_ZddMan, a int32, b int32) int32 {
    var A, B *Abc_ZddObj; 
    var r0, r1, b2, t1, t2, r int32;
    if a == 0 { return 0; }
    if b == 0 { return 0; }
    if a == 1 { return b; }
    if b == 1 { return a; }
    if a > b {  return Abc_ZddDotProduct( p, b, a ); }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_DOT_PROD)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    B = Abc_ZddNode( p, b );
    if A.Var < B.Var {
        r0 = Abc_ZddDotProduct( p, A.False, b );
r1 = Abc_ZddDotProduct( p, A.True, b );
    } else if A.Var > B.Var {
        r0 = Abc_ZddDotProduct( p, a, B.False );
r1 = Abc_ZddDotProduct( p, a, B.True ); 
    } else {
        r0 = Abc_ZddDotProduct( p, A.False, B.False );
b2 = Abc_ZddUnion( p, B.False, B.True );
t1 = Abc_ZddDotProduct( p, A.True, b2 );
t2 = Abc_ZddDotProduct( p, A.False, B.True );
r1 = Abc_ZddUnion( p, t1, t2 ); }
    r = Abc_ZddUniqueCreate( p, Abc_MinInt(A.Var, B.Var), r1, r0 );
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_DOT_PROD, r );
}
func Abc_ZddDotMinProduct6(p *Abc_ZddMan, a int32, b int32) int32 {
    var A, B *Abc_ZddObj; 
    var r0, r1, b2, t1, t2, r int32;
    if a == 0 { return 0; }
    if b == 0 { return 0; }
    if a == 1 { return b; }
    if b == 1 { return a; }
    if a > b {  return Abc_ZddDotMinProduct6( p, b, a ); }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_DOT_PROD_6)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    B = Abc_ZddNode( p, b );
    if A.Var < B.Var {
        r0 = Abc_ZddDotMinProduct6( p, A.False, b );
r1 = Abc_ZddDotMinProduct6( p, A.True, b );
    } else if A.Var > B.Var {
        r0 = Abc_ZddDotMinProduct6( p, a, B.False );
r1 = Abc_ZddDotMinProduct6( p, a, B.True ); 
    } else {
        r0 = Abc_ZddDotMinProduct6( p, A.False, B.False );
b2 = Abc_ZddMinUnion( p, B.False, B.True );
t1 = Abc_ZddDotMinProduct6( p, A.True, b2 );
t2 = Abc_ZddDotMinProduct6( p, A.False, B.True );
r1 = Abc_ZddMinUnion( p, t1, t2 ); }
    r1 = Abc_ZddThresh( p, r1, 5 );
r1 = Abc_ZddDiff( p, r1, r0 ); 
    r = Abc_ZddUniqueCreate( p, Abc_MinInt(A.Var, B.Var), r1, r0 );
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_DOT_PROD_6, r );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddPerm(p *Abc_ZddMan, a int32, Var int32) int32 {
    var A *Abc_ZddObj; 
    var r0, r1, r int32;
    assert(Var < p.nVars, "Var < p.nVars");
    if a == 0 {  return 0; }
    if a == 1 {  return Abc_ZddIthVar(Var); }
    if (r = Abc_ZddCacheLookup(p, a, Var, ABC_ZDD_OPER_PERM)) >= 0 {
        return r; }
    A = Abc_ZddNode( p, a );
    if p.pV2TI[A.Var] > p.pV2TI[Var] { // Ai > Bi
        r = Abc_ZddUniqueCreate( p, Var, a, 0 );
    } else if int32(A.Var) == Var { // Ai == Bi && Aj == Bj
        r0 = Abc_ZddPerm( p, A.False, Var );
r  = Abc_ZddUnion( p, r0, A.True );
    } else {
        var VarPerm, VarTop int32;
        var Ai int32  = p.pV2TI[A.Var];
        var Aj int32  = p.pV2TJ[A.Var];
        var Bi int32  = p.pV2TI[Var];
        var Bj int32  = p.pV2TJ[Var];
        assert(Ai < Aj && Bi < Bj && Ai <= Bi, "Ai < Aj && Bi < Bj && Ai <= Bi");
        if Aj == Bi {
            VarPerm = Var;
VarTop  = Abc_ZddVarIJ(p, Ai, Bj);
        } else if Aj == Bj {
            VarPerm = Var;
VarTop  = Abc_ZddVarIJ(p, Ai, Bi);
        } else if Ai == Bi {
            VarPerm = Abc_ZddVarIJ(p, Abc_MinInt(Aj, Bj), Abc_MaxInt(Aj, Bj));
VarTop  = A.Var;
        } else { // no clash
            VarPerm = Var;
VarTop  = A.Var; }
        assert(p.pV2TI[VarPerm] > p.pV2TI[VarTop], "p.pV2TI[VarPerm] > p.pV2TI[VarTop]");
        r0 = Abc_ZddPerm( p, A.False, Var ); 
        r1 = Abc_ZddPerm( p, A.True, VarPerm );
        assert(Abc_ZddObjVar(p, r1) > VarTop, "Abc_ZddObjVar(p, r1) > VarTop");
        if Abc_ZddObjVar(p, r0) > VarTop {
            r = Abc_ZddUniqueCreate( p, VarTop, r1, r0 );
        } else {
            r1 = Abc_ZddUniqueCreate( p, VarTop, r1, 0 );
r = Abc_ZddUnion( p, r0, r1 ); }
    }
    return Abc_ZddCacheInsert( p, a, Var, ABC_ZDD_OPER_PERM, r );
}
func Abc_ZddPermProduct(p *Abc_ZddMan, a int32, b int32) int32 {
    var B *Abc_ZddObj; 
    var r0, r1, r int32;
    if a == 0 { return 0; }
    if a == 1 { return b; }
    if b == 0 { return 0; }
    if b == 1 { return a; }
    if (r = Abc_ZddCacheLookup(p, a, b, ABC_ZDD_OPER_PERM_PROD)) >= 0 {
        return r; }
    B  = Abc_ZddNode( p, b );
    r0 = Abc_ZddPermProduct( p, a, B.False ); 
    r1 = Abc_ZddPermProduct( p, a, B.True ); 
    r1 = Abc_ZddPerm( p, r1, B.Var );
    r  = Abc_ZddUnion( p, r0, r1 );
    return Abc_ZddCacheInsert( p, a, b, ABC_ZDD_OPER_PERM_PROD, r );
}

/**Function*************************************************************

  Synopsis    [Printing permutations and transpositions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddPermPrint(pPerm *int32, Size int32) {
    var i int32;
    fmt.Printf("{");
    for  i = 0; i < Size; i++  {
        fmt.Printf(" %2d", pPerm[i]);
    }
    fmt.Printf(" }\n");
}
func Abc_ZddCombPrint(pComb *int32, nTrans int32) {
    var i int32;
    if nTrans == 0 {
        fmt.Printf("Empty set"); }
    for  i = 0; i < nTrans; i++  {
        fmt.Printf("(%d %d)", pComb[i] >> 16, pComb[i] & 0xffff);
    }
    fmt.Printf("\n");
}
func Abc_ZddPerm2Comb(pPerm *int32, Size int32, pComb *int32) int32 {
    var i, j, nTrans int32;
    for  i = 0; i < Size; i++  {
        if i != pPerm[i] {
            for  j = i+1; j < Size; j++  {
                if i == pPerm[j] {
                    break; }
            }
            pComb[nTrans++] = (i << 16) | j;
            // inlining of macro ABC_SWAP
            { var t int32  = pPerm[i]; pPerm[i] = pPerm[j]; pPerm[j] = t; };
            assert(i == pPerm[i], "i == pPerm[i]");
        }
    }
    return nTrans;
}
func Abc_ZddComb2Perm(pComb *int32, nTrans int32, pPerm *int32, Size int32) {
    var v int32;
    for  v = 0; v < Size; v++  {
        pPerm[v] = v;
    }
    for  v = nTrans-1; v >= 0; v--  {
        // inlining of macro ABC_SWAP
         var t int32  = pPerm[pComb[v] >> 16]; pPerm[pComb[v] >> 16] = pPerm[pComb[v] & 0xffff]; pPerm[pComb[v] & 0xffff] = t; };
}
func Abc_ZddPermCombTest() {
    var Size int32  = 10;
    var pPerm [10]int32 = [10]int32{ 6, 5, 7, 0, 3, 2, 1, 8, 9, 4 };
    var pComb [10]int32;
    var nTrans int32;
    Abc_ZddPermPrint( pPerm, Size );
    nTrans = Abc_ZddPerm2Comb( pPerm, Size, pComb );
    Abc_ZddCombPrint( pComb, nTrans );
    Abc_ZddComb2Perm( pComb, nTrans, pPerm, Size );
    Abc_ZddPermPrint( pPerm, Size );
    Size = 0;
}

/**Function*************************************************************

  Synopsis    [Printing ZDDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddPrint_rec(p *Abc_ZddMan, a int32, pPath *int32, Size int32) {
    var A *Abc_ZddObj;  
    if a == 0 { return; }
//    if ( a == 1 ) { Abc_ZddPermPrint( pPath, Size ); return; }
    if a == 1 { 
        var pPerm [24]int32;
        var pComb [24]int32;
        var i int32;
        assert(p.nPermSize <= 24, "p.nPermSize <= 24");
        for  i = 0; i < Size; i++  {
            pComb[i] = (p.pV2TI[pPath[i]] << 16) | p.pV2TJ[pPath[i]];
        }
        Abc_ZddCombPrint( pComb, Size ); 
        Abc_ZddComb2Perm( pComb, Size, pPerm, p.nPermSize );
        Abc_ZddPermPrint( pPerm, p.nPermSize );
        return; 
    }
    A = Abc_ZddNode( p, a );
    Abc_ZddPrint_rec( p, A.False, pPath, Size );
    pPath[Size] = A.Var;
    Abc_ZddPrint_rec( p, A.True, pPath, Size + 1 );
}
func Abc_ZddPrint(p *Abc_ZddMan, a int32) {
    // inlining of macro ABC_CALLOC
    var pPath *int32  = ((int *) calloc((size_t)(p.nVars), sizeof(int)));
    Abc_ZddPrint_rec( p, a, pPath, 0 );
    // inlining of macro ABC_FREE
    ((pPath) ? (free((char *) (pPath)), (pPath) = 0) : 0);
}
func Abc_ZddPrintTest(p *Abc_ZddMan) {
//    int nSets = 2;
//    int Size = 2;
//    int pSets[2][2] = { {5, 0}, {3, 11} };
    var nSets int32  = 3;
    var Size int32  = 5;
    var pSets [3][5]int32 = [3][5]int32{ {5, 0, 2, 10, 7}, {3, 11, 10, 7, 2}, {0, 2, 5, 10, 7} };
    var i, Set, Union int32;
    for  i = 0; i < nSets; i++  {
        Abc_ZddPermPrint( pSets[i], Size );
        Set = Abc_ZddBuildSet( p, pSets[i], Size );
        Union = Abc_ZddUnion( p, Union, Set );
    }
    fmt.Printf("Resulting set:\n");
    Abc_ZddPrint( p, Union );
    fmt.Printf("\n");
    fmt.Printf("Nodes = %d.   Path = %d.\n", Abc_ZddCountNodes(p, Union), Abc_ZddCountPaths(p, Union));
    Size = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddGiaTest(pGia *Gia_Man_t) {
    var p *Abc_ZddMan;
    var pObj *Gia_Obj_t;
    var vNodes *Vec_Int_t;
    var i, r, Paths int32;
    p = Abc_ZddManAlloc( Gia_ManObjNum(pGia), 1 << 24 ); // 576 MB  (36 B/node)
    Gia_ManFillValue( pGia );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        pObj.Value = Abc_ZddIthVar( Gia_ObjId(pGia, pObj) );
    }
    vNodes = Vec_IntAlloc( Gia_ManAndNum(pGia) );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        r = Abc_ZddDotMinProduct6( p, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
        r = Abc_ZddUnion( p, r, Abc_ZddIthVar(i) );
        pObj.Value = r;
        Vec_IntPush( vNodes, r );
        // print
//        printf( "Node %d:\n", i );
//        Abc_ZddPrint( p, r );
//        printf( "Node %d   ZddNodes = %d\n", i, Abc_ZddCountNodes(p, r) );
    }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Paths += Abc_ZddCountPaths(p, pObj.Value); }
    }
    fmt.Printf("Paths = %d.  Shared nodes = %d.\n", Paths, Abc_ZddCountNodesArray(p, vNodes));
    Vec_IntFree( vNodes );
    Abc_ZddManFree( p );
}
/*
    abc 01> &r pj1.aig; &ps; &test
    pj1      : i/o =   1769/   1063  and =   16285  lev =  156 (12.91)  mem = 0.23 MB
    Paths = 839934.  Shared nodes = 770999.
    ZDD stats: Var = 19118  Obj = 11578174  All = 16777216  Hits = 25617277  Miss = 40231476  Mem = 576.00 MB
*/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ZddPermTestInt(p *Abc_ZddMan) {
    var nPerms int32  = 3;
    var Size int32  = 5;
    var pPerms [3][5]int32 = [3][5]int32{ {1, 0, 2, 4, 3}, {1, 2, 4, 0, 3}, {0, 3, 2, 1, 4} };
    var pComb [5]int32;
    var nTrans int32;
    var i, k, Set, Union, iPivot int32;
    for  i = 0; i < nPerms; i++  {
        Abc_ZddPermPrint( pPerms[i], Size );
    }
    for  i = 0; i < nPerms; i++  {
        fmt.Printf("Perm %d:\n", i);
        Abc_ZddPermPrint( pPerms[i], Size );
        nTrans = Abc_ZddPerm2Comb( pPerms[i], Size, pComb );
        Abc_ZddCombPrint( pComb, nTrans );
        for  k = 0; k < nTrans; k++  {
            pComb[k] = Abc_ZddVarIJ( p, pComb[k] >> 16, pComb[k] & 0xFFFF );
        }
        Abc_ZddPermPrint( pComb, nTrans );
        // add to ZDD
        Set = Abc_ZddBuildSet( p, pComb, nTrans );
        Union = Abc_ZddUnion( p, Union, Set );
    }
    fmt.Printf("\nResulting set of permutations:\n");
    Abc_ZddPrint( p, Union );
    fmt.Printf("Nodes = %d.   Path = %d.\n", Abc_ZddCountNodes(p, Union), Abc_ZddCountPaths(p, Union));

    iPivot = Abc_ZddVarIJ( p, 3, 4 );
    Union = Abc_ZddPerm( p, Union, iPivot );

    fmt.Printf("\nResulting set of permutations:\n");
    Abc_ZddPrint( p, Union );
    fmt.Printf("Nodes = %d.   Path = %d.\n", Abc_ZddCountNodes(p, Union), Abc_ZddCountPaths(p, Union));
    fmt.Printf("\n");
}

func Abc_ZddPermTest() {
    var p *Abc_ZddMan;
    p = Abc_ZddManAlloc( 10, 1 << 20 );
    Abc_ZddManCreatePerms( p, 5 );
    Abc_ZddPermTestInt( p );
    Abc_ZddManFree( p );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_EnumerateCubeStatesZdd() {
    var pXYZ [3][9][2]int32 = [3][9][2]int32{
        { {3, 5}, {3,17}, {3,15}, {1, 6}, {1,16}, {1,14}, {2, 4}, {2,18}, {2,13} },
        { {2,14}, {2,24}, {2,12}, {3,13}, {3,23}, {3,10}, {1,15}, {1,22}, {1,11} },
        { {1,10}, {1, 7}, {1, 4}, {3,12}, {3, 9}, {3, 6}, {2,11}, {2, 8}, {2, 5} }  };
#ifdef WIN32
    int LogObj = 24;
#else
    int LogObj = 27;
#endif
    var p *Abc_ZddMan;
    var i int32;
    var k int32;
    var pComb [9]int32;
    var pPerm [24]int32;
    var nSize int32;
    var ZddTurn1, ZddTurn2, ZddTurn3, ZddTurns, ZddAll int32;
    var clk abctime  = Abc_Clock();
    fmt.Printf("Enumerating states of 2x2x2 cube.\n");
    p = Abc_ZddManAlloc( 24 * 23 / 2, 1 << LogObj ); // finished with 2^27 (4 GB)
    Abc_ZddManCreatePerms( p, 24 );
    // init state
    fmt.Printf("Iter %2d -> %8d  Nodes = %7d  Used = %10d  ", 0, 1, 0, 2);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // first 9 states
    ZddTurns = 1;
    for  i = 0; i < 3; i++  {
        for  k = 0; k < 24; k++  {
            pPerm[k] = k;
        }
        for  k = 0; k < 9; k++  {
            // inlining of macro ABC_SWAP
             var t int32  = pPerm[pXYZ[i][k][0]-1]; pPerm[pXYZ[i][k][0]-1] = pPerm[pXYZ[i][k][1]-1]; pPerm[pXYZ[i][k][1]-1] = t; };
        nSize = Abc_ZddPerm2Comb( pPerm, 24, pComb );
        assert(nSize == 9, "nSize == 9");
        for  k = 0; k < 9; k++  {
            pComb[k] = Abc_ZddVarIJ( p, pComb[k] >> 16, pComb[k] & 0xffff );
        }
        // add first turn
        ZddTurn1  = Abc_ZddBuildSet( p, pComb, 9 );
        ZddTurns = Abc_ZddUnion( p, ZddTurns, ZddTurn1 );
        //Abc_ZddPrint( p, ZddTurn1 );
        // add second turn
        ZddTurn2 = Abc_ZddPermProduct( p, ZddTurn1, ZddTurn1 );
        ZddTurns = Abc_ZddUnion( p, ZddTurns, ZddTurn2 );
        //Abc_ZddPrint( p, ZddTurn2 );
        // add third turn
        ZddTurn3 = Abc_ZddPermProduct( p, ZddTurn2, ZddTurn1 );
        ZddTurns = Abc_ZddUnion( p, ZddTurns, ZddTurn3 );
        //Abc_ZddPrint( p, ZddTurn3 );
        //printf( "\n" );
    }
    //Abc_ZddPrint( p, ZddTurns );
    fmt.Printf("Iter %2d -> %8d  Nodes = %7d  Used = %10d  ", 1, Abc_ZddCountPaths(p, ZddTurns), Abc_ZddCountNodes(p, ZddTurns), p.nObjs);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // other states
    ZddAll = ZddTurns;
    for  i = 2; i <= 100; i++  {
        var ZddAllPrev int32  = ZddAll;
        ZddAll = Abc_ZddPermProduct( p, ZddAll, ZddTurns );
        fmt.Printf("Iter %2d -> %8d  Nodes = %7d  Used = %10d  ", i, Abc_ZddCountPaths(p, ZddAll), Abc_ZddCountNodes(p, ZddAll), p.nObjs);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        if ZddAllPrev == ZddAll {
            break; }
    }
    Abc_ZddManFree( p );
}

/*
Enumerating states of 2x2x2 cube.
Iter  0 ->        1  Nodes =       0  Used =          2  Time =     0.00 sec
Iter  1 ->       10  Nodes =      63  Used =        577  Time =     0.00 sec
Iter  2 ->       64  Nodes =     443  Used =       4349  Time =     0.03 sec
Iter  3 ->      385  Nodes =    2018  Used =      26654  Time =     0.14 sec
Iter  4 ->     2232  Nodes =    7451  Used =     119442  Time =     0.45 sec
Iter  5 ->    12224  Nodes =   25178  Used =     490038  Time =     1.10 sec
Iter  6 ->    62360  Nodes =   83955  Used =    1919750  Time =     1.79 sec
Iter  7 ->   289896  Nodes =  290863  Used =    7182932  Time =     3.15 sec
Iter  8 ->  1159968  Nodes =  614845  Used =   25301123  Time =     8.03 sec
Iter  9 ->  3047716  Nodes =  585664  Used =   66228369  Time =    20.22 sec
Iter 10 ->  3671516  Nodes =   19430  Used =  102292452  Time =    33.41 sec
Iter 11 ->  3674160  Nodes =     511  Used =  103545878  Time =    33.92 sec
Iter 12 ->  3674160  Nodes =     511  Used =  103566266  Time =    33.93 sec
ZDD stats: Var = 276  Obj = 103566266  Alloc = 134217728  Hit = 63996630  Miss = 141768893  Mem = 4608.00 MB
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

