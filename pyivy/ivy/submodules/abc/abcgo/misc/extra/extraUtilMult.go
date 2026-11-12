package extra

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [extraUtilMult.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Simple BDD package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 23, 2018.]

  Revision    [$Id: extraUtilMult.c,v 1.0 2018/05/23 00:00:00 alanmi Exp $]

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

   
type Abc_BddMan struct {
    nVars                int32;         // the number of variables
    nObjs                int32;         // the number of nodes used
    nObjsAlloc                int32;    // the number of nodes allocated
    pUnique               *int32;       // unique table for nodes
    pNexts               *int32;        // next pointer for nodes
    pCache               *int32;        // array of triples <arg0, arg1, AND(arg0, arg1)>
    pObjs               *int32;         // array of pairs <cof0, cof1> for each node
    unsigned char *    pVars;         // array of variables for each node
    unsigned char *    pMark;         // array of marks for each BDD node
    nUniqueMask           uint32;   // selection mask for unique table
    nCacheMask           uint32;    // selection mask for computed table
    nCacheLookups                int32; // the number of computed table lookups
    nCacheMisses                int32;  // the number of computed table misses
    nMemory               word;       // total amount of memory used (in bytes)
}

func Abc_BddIthVar(i int32) int32 { return Abc_Var2Lit(i + 1, 0);                            }
func Abc_BddHash(Arg0 int32, Arg1 int32, Arg2 int32) uint32 { return 12582917 * Arg0 + 4256249 * Arg1 + 741457 * Arg2; }

func Abc_BddVar(p *Abc_BddMan, i int32) int32 { return int32(p.pVars[Abc_Lit2Var(i)]);                    }
func Abc_BddThen(p *Abc_BddMan, i int32) int32 { return Abc_LitNotCond(p.pObjs[Abc_LitRegular(i)], Abc_LitIsCompl(i));    }
func Abc_BddElse(p *Abc_BddMan, i int32) int32 { return Abc_LitNotCond(p.pObjs[Abc_LitRegular(i)+1], Abc_LitIsCompl(i));  }

func Abc_BddMark(p *Abc_BddMan, i int32) int32 { return int32(p.pMark[Abc_Lit2Var(i)]);                    }
func Abc_BddSetMark(p *Abc_BddMan, i int32, m int32) { p.pMark[Abc_Lit2Var(i)] = m;                            }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creating new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BddUniqueCreateInt(p *Abc_BddMan, Var int32, Then int32, Else int32) int32 {
    var q *int32  = p.pUnique + (Abc_BddHash(Var, Then, Else) & p.nUniqueMask);
    for  ; *q; q = p.pNexts + *q  {
        if int32(p.pVars[*q]) == Var && p.pObjs[*q+*q] == Then && p.pObjs[*q+*q+1] == Else {
            return Abc_Var2Lit(*q, 0); }
    }
    if p.nObjs == p.nObjsAlloc {
        fmt.Printf("Aborting because the number of nodes exceeded %d.\n", p.nObjsAlloc);
fflush(os.Stdout); }
    assert(p.nObjs < p.nObjsAlloc, "p.nObjs < p.nObjsAlloc");     
    *q = p.nObjs++;
    p.pVars[*q] = Var;
    p.pObjs[*q+*q] = Then;
    p.pObjs[*q+*q+1] = Else;
//    printf( "Added node %3d: Var = %3d.  Then = %3d.  Else = %3d\n", *q, Var, Then, Else );
    assert(!Abc_LitIsCompl(Else), "!Abc_LitIsCompl(Else)");
    return Abc_Var2Lit(*q, 0);
}
func Abc_BddUniqueCreate(p *Abc_BddMan, Var int32, Then int32, Else int32) int32 {
    assert(Var >= 0 && Var < p.nVars, "Var >= 0 && Var < p.nVars");
    assert(Var < Abc_BddVar(p, Then), "Var < Abc_BddVar(p, Then)");
    assert(Var < Abc_BddVar(p, Else), "Var < Abc_BddVar(p, Else)");
    if Then == Else {
        return Else; }
    if !Abc_LitIsCompl(Else) {
        return Abc_BddUniqueCreateInt( p, Var, Then, Else ); }
    return Abc_LitNot( Abc_BddUniqueCreateInt(p, Var, Abc_LitNot(Then), Abc_LitNot(Else)) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BddCacheLookup(p *Abc_BddMan, Arg1 int32, Arg2 int32) int32 {
    var pEnt *int32  = p.pCache + 3*(Abc_BddHash(0, Arg1, Arg2) & p.nCacheMask);
    p.nCacheLookups++;
    if pEnt[0] == Arg1 && pEnt[1] == Arg2 {
        return pEnt[2];
    } else {
        return -1;
    }
}
func Abc_BddCacheInsert(p *Abc_BddMan, Arg1 int32, Arg2 int32, Res int32) int32 {
    var pEnt *int32  = p.pCache + 3*(Abc_BddHash(0, Arg1, Arg2) & p.nCacheMask);
    pEnt[0] = Arg1;  pEnt[1] = Arg2;  pEnt[2] = Res;
    p.nCacheMisses++;
    assert(Res >= 0, "Res >= 0");
    return Res;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BddManAlloc(nVars int32, nObjs int32) *Abc_BddMan {
     var i int32;
    // inlining of macro ABC_CALLOC
    p := &Abc_BddMan{};
    p.nVars       = nVars;
    p.nObjsAlloc  = nObjs;
    p.nUniqueMask = (1 << Abc_Base2Log(nObjs)) - 1;
    p.nCacheMask  = (1 << Abc_Base2Log(nObjs)) - 1;
    // inlining of macro ABC_CALLOC
    p.pUnique     = ((int *) calloc((size_t)(p.nUniqueMask + 1), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pNexts      = ((int *) calloc((size_t)(p.nObjsAlloc), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pCache      = ((int *) calloc((size_t)(3*(p.nCacheMask + 1)), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pObjs       = ((int *) calloc((size_t)(2*p.nObjsAlloc), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pMark       = ((unsigned char *) calloc((size_t)(p.nObjsAlloc), sizeof(unsigned char)));
    // inlining of macro ABC_CALLOC
    p.pVars       = ((unsigned char *) calloc((size_t)(p.nObjsAlloc), sizeof(unsigned char)));
    p.pVars[0]    = 0xff;
    p.nObjs       = 1;
    for  i = 0; i < nVars; i++  {
        Abc_BddUniqueCreate( p, i, 1, 0 );
    }
    assert(p.nObjs == nVars + 1, "p.nObjs == nVars + 1");
    p.nMemory = sizeof(Abc_BddMan)/4 + 
        p.nUniqueMask + 1 + p.nObjsAlloc + 
        (p.nCacheMask + 1) * 3 * sizeof(int)/4 + 
        p.nObjsAlloc * 2 * sizeof(int)/4;
    return p;
}
func Abc_BddManFree(p *Abc_BddMan) {
    fmt.Printf("BDD stats: Var = %d  Obj = %d  Alloc = %d  Hit = %d  Miss = %d  ", p.nVars, p.nObjs, p.nObjsAlloc, p.nCacheLookups-p.nCacheMisses, p.nCacheMisses);
    fmt.Printf("Mem = %.2f MB\n", 4.0*int32((p.nMemory/(1<<20))));
    p.pUnique = nil;
    p.pNexts = nil;
    p.pCache = nil;
    p.pObjs = nil;
    p.pVars = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Boolean AND.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BddAnd(p *Abc_BddMan, a int32, b int32) int32 {
    var r0, r1, r int32;
    if a == 0 { return 0; }
    if b == 0 { return 0; }
    if a == 1 { return b; }
    if b == 1 { return a; }
    if a == b { return a; }
    if a > b {  return Abc_BddAnd( p, b, a ); }
    if (r = Abc_BddCacheLookup(p, a, b)) >= 0 {
        return r; }
    if Abc_BddVar(p, a) < Abc_BddVar(p, b) {
        r0 = Abc_BddAnd( p, Abc_BddElse(p, a), b );
r1 = Abc_BddAnd( p, Abc_BddThen(p, a), b );
    } else if Abc_BddVar(p, a) > Abc_BddVar(p, b) {
        r0 = Abc_BddAnd( p, a, Abc_BddElse(p, b) );
r1 = Abc_BddAnd( p, a, Abc_BddThen(p, b) );
    } else { // if ( Abc_BddVar(p, a) == Abc_BddVar(p, b) )
        r0 = Abc_BddAnd( p, Abc_BddElse(p, a), Abc_BddElse(p, b) );
r1 = Abc_BddAnd( p, Abc_BddThen(p, a), Abc_BddThen(p, b) ); }
    r = Abc_BddUniqueCreate( p, Abc_MinInt(Abc_BddVar(p, a), Abc_BddVar(p, b)), r1, r0 );
    return Abc_BddCacheInsert( p, a, b, r );
}
func Abc_BddOr(p *Abc_BddMan, a int32, b int32) int32 {
    return Abc_LitNot( Abc_BddAnd(p, Abc_LitNot(a), Abc_LitNot(b)) );
}

/**Function*************************************************************

  Synopsis    [Counting nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BddCount_rec(p *Abc_BddMan, i int32) int32 {
    if i < 2 {
        return 0; }
    if Abc_BddMark(p, i) != 0 {
        return 0; }
    Abc_BddSetMark( p, i, 1 );
    return 1 + Abc_BddCount_rec(p, Abc_BddElse(p, i)) + Abc_BddCount_rec(p, Abc_BddThen(p, i));
}
func Abc_BddUnmark_rec(p *Abc_BddMan, i int32) {
    if i < 2 {
        return; }
    if !(Abc_BddMark(p, i) != 0) {
        return; }
    Abc_BddSetMark( p, i, 0 );
    Abc_BddUnmark_rec( p, Abc_BddElse(p, i) );
    Abc_BddUnmark_rec( p, Abc_BddThen(p, i) );
}
func Abc_BddCountNodes(p *Abc_BddMan, i int32) int32 {
    var Count int32  = Abc_BddCount_rec( p, i );
    Abc_BddUnmark_rec( p, i );
    return Count;
}
func Abc_BddCountNodesArray(p *Abc_BddMan, vNodes *Vec_Int_t) int32 {
    var i, a, Count int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((a) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Count += Abc_BddCount_rec( p, a );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((a) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Abc_BddUnmark_rec( p, a );
    }
    return Count;
}
func Abc_BddCountNodesArray2(p *Abc_BddMan, vNodes *Vec_Int_t) int32 {
    var i, a, Count int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((a) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Count += Abc_BddCount_rec( p, a );
        Abc_BddUnmark_rec( p, a );
    }
    return Count;
}



/**Function*************************************************************

  Synopsis    [Printing BDD.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BddPrint_rec(p *Abc_BddMan, a int32, pPath *int32) {
    if a == 0 { 
        return; }
    if a == 1 { 
        var i int32;
        for  i = 0; i < p.nVars; i++  {
            if pPath[i] == 0 || pPath[i] == 1 {
                fmt.Printf("%c%d", pPath[i] ? '+' : '-', i); }
        }
        fmt.Printf(" ");
        return; 
    }
    pPath[Abc_BddVar(p, a)] =  0;
    Abc_BddPrint_rec( p, Abc_BddElse(p, a), pPath );
    pPath[Abc_BddVar(p, a)] =  1;
    Abc_BddPrint_rec( p, Abc_BddThen(p, a), pPath );
    pPath[Abc_BddVar(p, a)] = -1;
}
func Abc_BddPrint(p *Abc_BddMan, a int32) {
    // inlining of macro ABC_FALLOC
    var pPath *int32  = ((int *) memset(malloc(sizeof(int) * (size_t)(p.nVars)), 0xff, sizeof(int) * (size_t)(p.nVars)));
    fmt.Printf("BDD %d = ", a);
    Abc_BddPrint_rec( p, a, pPath );
    pPath = nil;
    fmt.Printf("\n");
}
func Abc_BddPrintTest(p *Abc_BddMan) {
    var bVarA int32  = Abc_BddIthVar(0);
    var bVarB int32  = Abc_BddIthVar(1);
    var bVarC int32  = Abc_BddIthVar(2);
    var bVarD int32  = Abc_BddIthVar(3);
    var bAndAB int32  = Abc_BddAnd( p, bVarA, bVarB );
    var bAndCD int32  = Abc_BddAnd( p, bVarC, bVarD );
    var bFunc int32   = Abc_BddOr( p, bAndAB, bAndCD );
    Abc_BddPrint( p, bFunc );
    fmt.Printf("Nodes = %d\n", Abc_BddCountNodes(p, bFunc));
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_BddGiaTest2(pGia *Gia_Man_t, fVerbose int32) {
    var p *Abc_BddMan  = Abc_BddManAlloc( 10, 100 );
    Abc_BddPrintTest( p );
    Abc_BddManFree( p );
}

func Abc_BddGiaTest(pGia *Gia_Man_t, fVerbose int32) {
    var p *Abc_BddMan;
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t; var i int32;
    p = Abc_BddManAlloc( Gia_ManCiNum(pGia), 1 << 20 ); // 30 B/node
    Gia_ManFillValue( pGia );
    Gia_ManConst0(pGia).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        pObj.Value = Abc_BddIthVar( i );
    }
    vNodes = Vec_IntAlloc( Gia_ManAndNum(pGia) );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var Cof0 int32  = Abc_LitNotCond(Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj));
        var Cof1 int32  = Abc_LitNotCond(Gia_ObjFanin1(pObj).Value, Gia_ObjFaninC1(pObj));
        pObj.Value = Abc_BddAnd( p, Cof0, Cof1 );
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
        pObj.Value = Abc_LitNotCond(Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj));
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
        if fVerbose != 0 {
            Abc_BddPrint( p, pObj.Value ); }
        Vec_IntPush( vNodes, pObj.Value );
    }
    fmt.Printf("Shared nodes = %d.\n", Abc_BddCountNodesArray2(p, vNodes));
    Vec_IntFree( vNodes );
    Abc_BddManFree( p );
}

/*
    abc 04> c.aig; &get; &test
    Shared nodes = 80.
    BDD stats: Var = 23  Obj = 206  Alloc = 1048576  Hit = 59  Miss = 216  Mem = 28.00 MB

    abc 05> c.aig; clp -r; ps
    c                             : i/o =   23/    2  lat =    0  nd =     2  edge =     46  bdd  =    80  lev = 1
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

