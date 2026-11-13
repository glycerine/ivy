package mpm

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [mpmPre.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [DSD-related precomputations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmPre.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/extra/extra.h"
//#include "bool/kit/kit.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Ifd_Obj_t struct {
    nFreq uint32; /* C bitfield: 18 bits; review packing and truncation */   // frequency
    nAnds uint32; /* C bitfield: 6 bits; review packing and truncation */   // number of AND gates
    nSupp uint32; /* C bitfield: 5 bits; review packing and truncation */   // support size
    Type uint32; /* C bitfield: 2 bits; review packing and truncation */   // type
    fWay uint32; /* C bitfield: 1 bits; review packing and truncation */   // transparent edge
    pFans [3]uint32;     // fanins
}

   
type Ifd_Man_t struct {
    pObjs        *Ifd_Obj_t;
    nObjs               int32;
    nObjsAlloc               int32;
    // hashing operations
    vArgs        *Vec_Int_t;     // iDsd1 op iDsdC
    vRes        *Vec_Int_t;      // result of operation
    vHash     *Hsh_IntMan_t;     // hash table 
    vMarks        *Vec_Int_t;    // marks where given N begins
    vTruths        *Vec_Wrd_t;   // truth tables
    vClauses        *Vec_Int_t;  // truth tables
    // other data
    vSuper        *Vec_Int_t;

}

func Ifd_ObjIsVar(p *Ifd_Obj_t) int32 { return p.Type == 0; }
func Ifd_ObjIsAnd(p *Ifd_Obj_t) int32 { return p.Type == 1; }
func Ifd_ObjIsXor(p *Ifd_Obj_t) int32 { return p.Type == 2; }
func Ifd_ObjIsMux(p *Ifd_Obj_t) int32 { return p.Type == 3; }

func Ifd_ManObj(p *Ifd_Man_t, i int32) *Ifd_Obj_t { assert(i >= 0 && i < p.nObjs, "i >= 0 && i < p.nObjs");                              return p.pObjs + i;    }
func Ifd_ManObjFromLit(p *Ifd_Man_t, iLit int32) *Ifd_Obj_t { return Ifd_ManObj( p, Abc_Lit2Var(iLit) );                                             }
func Ifd_ObjId(p *Ifd_Man_t, pObj *Ifd_Obj_t) int32 { assert(pObj - p.pObjs >= 0 && pObj - p.pObjs < p.nObjs, "pObj - p.pObjs >= 0 && pObj - p.pObjs < p.nObjs");  return pObj - p.pObjs; }
func Ifd_LitSuppSize(p *Ifd_Man_t, iLit int32) int32 { if iLit > 0 {
    return Ifd_ManObjFromLit(p, iLit).nSupp;
} else {
    return 0;
}                               }
func Ifd_LitNumAnds(p *Ifd_Man_t, iLit int32) int32 { if iLit > 0 {
    return Ifd_ManObjFromLit(p, iLit).nAnds;
} else {
    return 0;
}                               }

// Original macro:
// #define Ifd_ManForEachNodeWithSupp( p, nVars, pLeaf, i )                       \
//     for ( i = Vec_IntEntry(p->vMarks, nVars); (i < Vec_IntEntry(p->vMarks, nVars+1)) && (pLeaf = Ifd_ManObj(p, i)); i++ )



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ManStart() *Ifd_Man_t {
    var p *Ifd_Man_t;
    p = new(Ifd_Man_t);
    p.nObjsAlloc = Abc_PrimeCudd( 50000000 );
    p.nObjs      = 2;
    p.pObjs      = ABC_CALLOC( Ifd_Obj_t, p.nObjsAlloc );
    memset( p.pObjs, 0xFF, sizeof(Ifd_Obj_t) ); // const node
    (p.pObjs + 1).nSupp = 1;  // variable
    (p.pObjs + 1).fWay  = 1;  // variable
    // hashing operations
    p.vArgs      = Vec_IntAlloc( 4000 );
    p.vRes       = Vec_IntAlloc( 1000 );
    p.vHash      = Hsh_IntManStart( p.vArgs, 4, 1000 );
    p.vMarks     = Vec_IntAlloc( 100 );
    Vec_IntPush( p.vMarks, 0 );
    Vec_IntPush( p.vMarks, 1 );
    Vec_IntPush( p.vMarks, p.nObjs );
    // other data
    p.vSuper     = Vec_IntAlloc( 1000 );
    p.vTruths    = Vec_WrdAlloc( 1000 );
    p.vClauses   = Vec_IntAlloc( 1000 );
    return p;
}
func Ifd_ManStop(p *Ifd_Man_t) {
    var i, This, Prev int32;
    Vec_IntForEachEntryStart( p.vMarks, This, i, 1 )
    {
        fmt.Printf("%d(%d:%d) ", i-1, This, This - Prev);
        Prev = This;
    }
    fmt.Printf("\n");

    Vec_IntFreeP( &p.vArgs );
    Vec_IntFreeP( &p.vRes );
    Vec_WrdFreeP( &p.vTruths );
    Vec_IntFreeP( &p.vClauses );
    Vec_IntFreeP( &p.vMarks );
    Hsh_IntManStop( p.vHash );
    Vec_IntFreeP( &p.vSuper );
    ABC_FREE( p.pObjs );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Printing structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ObjPrint_rec(p *Ifd_Man_t, iLit int32, pCounter *int32, DiffType int32) {
    var Symb [2][4]byte = [2][4]byte{ {'?','(','[','<'}, {'?',')',']','>'} };
    var pDsd *Ifd_Obj_t;
    if Abc_LitIsCompl(iLit) {
        fmt.Printf("!");
iLit = Abc_LitNot(iLit); }
    if iLit == 2 { fmt.Printf("%c", 'a' + (*pCounter)++); return; }
    pDsd = Ifd_ManObjFromLit( p, iLit );
    if DiffType != 0 {
        fmt.Printf("%c", Symb[0][pDsd.Type]); }
    Ifd_ObjPrint_rec( p, pDsd.pFans[0], pCounter, pDsd.Type == 3 || Abc_LitIsCompl(pDsd.pFans[0]) || pDsd.Type != Ifd_ManObjFromLit(p, pDsd.pFans[0]).Type );
    Ifd_ObjPrint_rec( p, pDsd.pFans[1], pCounter, pDsd.Type == 3 || Abc_LitIsCompl(pDsd.pFans[1]) || pDsd.Type != Ifd_ManObjFromLit(p, pDsd.pFans[1]).Type );
    if pDsd.pFans[2] != -1 {
    Ifd_ObjPrint_rec( p, pDsd.pFans[2], pCounter, pDsd.Type == 3 || Abc_LitIsCompl(pDsd.pFans[2]) || pDsd.Type != Ifd_ManObjFromLit(p, pDsd.pFans[2]).Type ); }
    if DiffType != 0 {
        fmt.Printf("%c", Symb[1][pDsd.Type]); }
}
func Ifd_ObjPrint(p *Ifd_Man_t, iLit int32) {
    var Counter int32;
    if iLit == 0 { fmt.Printf("0"); return; }
    if iLit == 1 { fmt.Printf("1"); return; }
    Ifd_ObjPrint_rec( p, iLit, &Counter, 1 );
}
func Ifd_ManPrint2(p *Ifd_Man_t) {
    var i int32;
    for  i = 0; i < p.nObjs; i++  {
        fmt.Printf("%4d : ", i);
        Ifd_ObjPrint( p, Abc_Var2Lit( i, 0 ) );
        fmt.Printf("\n");
    }
}
func Ifd_ManPrint(p *Ifd_Man_t) {
    var i int32;
    for  i = 0; i < p.nObjs; i++  {
        var Fun word  = Vec_WrdEntry( p.vTruths, i );
        fmt.Printf("    { ");
        fmt.Printf("%d, ", Extra_TruthSupportSize((unsigned *)&Fun, 6));
        fmt.Printf("%2d, ", Ifd_LitNumAnds(p, Abc_Var2Lit(i, 0)));
        fmt.Printf("%2d, ", Vec_IntEntry(p.vClauses, i));
        fmt.Printf("ABC_CONST(");
        Extra_PrintHex( os.Stdout, (unsigned *)&Fun, 6 ); 
        fmt.Printf("), \"");
        Ifd_ObjPrint( p, Abc_Var2Lit( i, 0 ) );
        fmt.Printf("\" },   // %4d \n", i);
    }
}


/**Function*************************************************************

  Synopsis    [Computing truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ObjTruth_rec(p *Ifd_Man_t, iLit int32, pCounter *int32) word {
    static word s_Truths6[6] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00),
        ABC_CONST(0xFFFF0000FFFF0000),
        ABC_CONST(0xFFFFFFFF00000000)
    };
    var pDsd *Ifd_Obj_t;
    var Fun0, Fun1, Fun2 word;
    assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
    if iLit == 2 {
        return s_Truths6[(*pCounter)++]; }
    pDsd = Ifd_ManObjFromLit( p, iLit );

    Fun0 = Ifd_ObjTruth_rec( p, Abc_LitRegular(pDsd.pFans[0]), pCounter );
    Fun1 = Ifd_ObjTruth_rec( p, Abc_LitRegular(pDsd.pFans[1]), pCounter );
    if pDsd.pFans[2] != -1 {
    Fun2 = Ifd_ObjTruth_rec( p, Abc_LitRegular(pDsd.pFans[2]), pCounter ); }

    if Abc_LitIsCompl(pDsd.pFans[0]) {
        Fun0 = ~Fun0;
    } else {
        Fun0 = Fun0;
    }
    if Abc_LitIsCompl(pDsd.pFans[1]) {
        Fun1 = ~Fun1;
    } else {
        Fun1 = Fun1;
    }
    if pDsd.pFans[2] != -1 {
    if Abc_LitIsCompl(pDsd.pFans[2]) {
        Fun2 = ~Fun2;
    } else {
        Fun2 = Fun2;
    } }

    if pDsd.Type == 1 {
        return Fun0 & Fun1; }
    if pDsd.Type == 2 {
        return Fun0 ^ Fun1; }
    if pDsd.Type == 3 {
        return (Fun2 & Fun1) | (~Fun2 & Fun0); }
    assert(false, "0");
    return -1;
}
func Ifd_ObjTruth(p *Ifd_Man_t, iLit int32) word {
    var Fun word;
    var Counter int32;
    if iLit == 0 {
        return 0; }
    if iLit == 1 {
        return ~(word)0; }
    Fun = Ifd_ObjTruth_rec( p, Abc_LitRegular(iLit), &Counter );
    if Abc_LitIsCompl(iLit) {
        return ~Fun;
    } else {
        return Fun;
    }
}
func Ifd_ManTruthAll(p *Ifd_Man_t) {
    var Fun word;
    var i int32;
    assert(Vec_WrdSize(p.vTruths) == 0, "Vec_WrdSize(p.vTruths) == 0");
    for  i = 0; i < p.nObjs; i++  {
        Fun = Ifd_ObjTruth( p, Abc_Var2Lit( i, 0 ) );
        Vec_WrdPush( p.vTruths, Fun );
//        Extra_PrintHex( stdout, (unsigned *)&Fun, 6 ); printf( " " );
//        Kit_DsdPrintFromTruth( (unsigned *)&Fun, 6 ); printf( "\n" );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ComputeCnfSizeOne(Truth word, nVars int32, vCover *Vec_Int_t, vCnf *Vec_Str_t) int32 {
    Vec_StrClear( vCnf );
    if Truth == 0 || ~Truth == 0 {
//        assert( nVars == 0 );
        Vec_StrPush( vCnf, byte((Truth == 0)) );
        Vec_StrPush( vCnf, byte(-1) );
        return 1;
    
    } else {
        var i, k, c, RetValue, Literal, Cube, nCubes int32;
        assert(nVars > 0, "nVars > 0");
        for  c = 0; c < 2; c ++  {
            if c != 0 {
                Truth = ~Truth;
            } else {
                Truth = Truth;
            }
            RetValue = Kit_TruthIsop( (unsigned *)&Truth, nVars, vCover, 0 );
            assert(RetValue == 0, "RetValue == 0");
            nCubes += Vec_IntSize( vCover );
            Vec_IntForEachEntry( vCover, Cube, i )
            {
                for  k = 0; k < nVars; k++  {
                    Literal = 3 & (Cube >> (k << 1));
                    if Literal == 1 {      // '0'  -> pos lit
                        Vec_StrPush( vCnf, byte(Abc_Var2Lit(k, 0)) );
                    } else if Literal == 2 { // '1'  -> neg lit
                        Vec_StrPush( vCnf, byte(Abc_Var2Lit(k, 1)) );
                    } else if Literal != 0 {
                        assert(false, "0"); }
                }
                Vec_StrPush( vCnf, byte(Abc_Var2Lit(nVars, c)) );
                Vec_StrPush( vCnf, byte(-1) );
            }
        }
        return nCubes;
    }
}
func Mpm_ComputeCnfSizeAll(p *Ifd_Man_t) {
    var vCover *Vec_Int_t  = Vec_IntAlloc( 1 << 16 );
    var vCnf *Vec_Str_t  = Vec_StrAlloc( 1000 );
    var Truth word;
    var i int32;
    assert(Vec_IntSize(p.vClauses) == 0, "Vec_IntSize(p.vClauses) == 0");
    Vec_WrdForEachEntry( p.vTruths, Truth, i )
        Vec_IntPush( p.vClauses, Mpm_ComputeCnfSizeOne(Truth, 6, vCover, vCnf) );
    Vec_IntFree( vCover );
    Vec_StrFree( vCnf );
}

/**Function*************************************************************

  Synopsis    [Canonicizing DSD structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ManHashLookup(p *Ifd_Man_t, iDsd0 int32, iDsd1 int32, iDsdC int32, Type int32) int32 {
    var pData [4]int32;
    assert(iDsdC != -1 || iDsd0 >= iDsd1, "iDsdC != -1 || iDsd0 >= iDsd1");
    assert(iDsdC == -1 || !Abc_LitIsCompl(iDsd1), "iDsdC == -1 || !Abc_LitIsCompl(iDsd1)");
    pData[0] = iDsd0;
    pData[1] = iDsd1;
    pData[2] = iDsdC;
    pData[3] = Type;
    return *Hsh_IntManLookup( p.vHash, (unsigned *)pData );
}
func Ifd_ManHashInsert(p *Ifd_Man_t, iDsd0 int32, iDsd1 int32, iDsdC int32, Type int32, Res int32) {
    var iObj int32;
    assert(iDsdC != -1 || iDsd0 >= iDsd1, "iDsdC != -1 || iDsd0 >= iDsd1");
    assert(iDsdC == -1 || !Abc_LitIsCompl(iDsd1), "iDsdC == -1 || !Abc_LitIsCompl(iDsd1)");
    Vec_IntPush( p.vArgs, iDsd0 );
    Vec_IntPush( p.vArgs, iDsd1 );
    Vec_IntPush( p.vArgs, iDsdC );
    Vec_IntPush( p.vArgs, Type );
    iObj = Hsh_IntManAdd( p.vHash, Vec_IntSize(p.vRes) );
    assert(iObj == Vec_IntSize(p.vRes), "iObj == Vec_IntSize(p.vRes)");
    Vec_IntPush( p.vRes, Res );
    assert(4 * Vec_IntSize(p.vRes) == Vec_IntSize(p.vArgs), "4 * Vec_IntSize(p.vRes) == Vec_IntSize(p.vArgs)");
}
func Ifd_ManHashFindOrAdd(p *Ifd_Man_t, iDsd0 int32, iDsd1 int32, iDsdC int32, Type int32) int32 {
    var pObj *Ifd_Obj_t;
    var iObj, Value int32;
    assert(iDsdC != -1 || iDsd0 >= iDsd1, "iDsdC != -1 || iDsd0 >= iDsd1");
    assert(iDsdC == -1 || !Abc_LitIsCompl(iDsd1), "iDsdC == -1 || !Abc_LitIsCompl(iDsd1)");
    Vec_IntPush( p.vArgs, iDsd0 );
    Vec_IntPush( p.vArgs, iDsd1 );
    Vec_IntPush( p.vArgs, iDsdC );
    Vec_IntPush( p.vArgs, Type );
    Value = Hsh_IntManAdd( p.vHash, Vec_IntSize(p.vRes) );
    if Value < Vec_IntSize(p.vRes) {
        iObj = Vec_IntEntry(p.vRes, Value);
        Vec_IntShrink( p.vArgs, Vec_IntSize(p.vArgs) - 4 );
        pObj = Ifd_ManObj( p, iObj );
//        pObj->nFreq++;
        assert(int32(pObj.Type) == Type, "int32(pObj.Type) == Type");
        assert(int32(pObj.nSupp) == Ifd_LitSuppSize(p, iDsd0) + Ifd_LitSuppSize(p, iDsd1) + Ifd_LitSuppSize(p, iDsdC), "int32(pObj.nSupp) == Ifd_LitSuppSize(p, iDsd0) + Ifd_LitSuppSize(p, iDsd1) + Ifd_LitSuppSize(p, iDsdC)");
    
    } else {
        if p.nObjs == p.nObjsAlloc {
            fmt.Printf("The number of nodes is more than %d\n", p.nObjs); }
        assert(p.nObjs < p.nObjsAlloc, "p.nObjs < p.nObjsAlloc");
        iObj = p.nObjs;
        pObj = Ifd_ManObj( p, p.nObjs++ );
//        pObj->nFreq = 1;
        pObj.nSupp = Ifd_LitSuppSize(p, iDsd0) + Ifd_LitSuppSize(p, iDsd1) + Ifd_LitSuppSize(p, iDsdC); 
        pObj.nAnds = Ifd_LitNumAnds(p, iDsd0) + Ifd_LitNumAnds(p, iDsd1) + Ifd_LitNumAnds(p, iDsdC) + ((Type == 1) ? 1 : 3); 
        pObj.Type  = Type;
        if Type == 1 {
            pObj.fWay = 0;
        } else if Type == 2 {
            pObj.fWay = Ifd_ManObjFromLit(p, iDsd0).fWay || Ifd_ManObjFromLit(p, iDsd1).fWay;
        } else if Type == 3 {
//            pObj->fWay = (Ifd_ManObjFromLit(p, iDsd0)->fWay && Ifd_ManObjFromLit(p, iDsd1)->fWay) || (Abc_Lit2Var(iDsd0) == Abc_Lit2Var(iDsd1) && Ifd_ManObjFromLit(p, iDsdC)->fWay);
            pObj.fWay = (Ifd_ManObjFromLit(p, iDsd0).fWay && Ifd_ManObjFromLit(p, iDsd1).fWay) || (iDsd0 == Abc_LitNot(iDsd1) && Ifd_ManObjFromLit(p, iDsdC).fWay);
        } else { assert(false, "0"); }
        pObj.pFans[0] = iDsd0;
        pObj.pFans[1] = iDsd1;
        pObj.pFans[2] = iDsdC;
        Vec_IntPush( p.vRes, iObj );
    }
    assert(4 * Vec_IntSize(p.vRes) == Vec_IntSize(p.vArgs), "4 * Vec_IntSize(p.vRes) == Vec_IntSize(p.vArgs)");
    return iObj;
}
func Ifd_ManOperSuper_rec(p *Ifd_Man_t, iLit int32, Type int32, vObjs *Vec_Int_t) {
    var pDsd *Ifd_Obj_t  = Ifd_ManObjFromLit( p, iLit );
    if Abc_LitIsCompl(iLit) || int32(pDsd.Type) != Type {
        Vec_IntPush( vObjs, iLit );
    } else {
        Ifd_ManOperSuper_rec( p, pDsd.pFans[0], Type, vObjs );
        Ifd_ManOperSuper_rec( p, pDsd.pFans[1], Type, vObjs );
    }
}
func Ifd_ManOper(p *Ifd_Man_t, iDsd0 int32, iDsd1 int32, iDsdC int32, Type int32) int32 {
    var i, iLit0, iLit1, iThis, fCompl int32;
    if Type == 1 { // AND
    
        if iDsd0 == 0 || iDsd1 == 0 {
            return 0; }
        if iDsd0 == 1 || iDsd1 == 1 {
            if iDsd0 == 1 {
                return iDsd1;
            } else {
                return iDsd0;
            } }
    
    } else if Type == 2 { // XOR
    
        if iDsd0 < 2 {
            return Abc_LitNotCond( iDsd1, iDsd0 ); }
        if iDsd1 < 2 {
            return Abc_LitNotCond( iDsd0, iDsd1 ); }
        if Abc_LitIsCompl(iDsd0) {
            fCompl ^= 1, iDsd0 = Abc_LitNot(iDsd0); }
        if Abc_LitIsCompl(iDsd1) {
            fCompl ^= 1, iDsd1 = Abc_LitNot(iDsd1); }
    
    } else if Type == 3 {
        if Abc_LitIsCompl(iDsdC) {
            ABC_SWAP( int, iDsd0, iDsd1 );
            iDsdC = Abc_LitNot(iDsdC);
        }
        if Abc_LitIsCompl(iDsd1) {
            fCompl ^= 1, iDsd0 = Abc_LitNot(iDsd0), iDsd1 = Abc_LitNot(iDsd1); }
    }
    assert(iDsd0 > 1 && iDsd1 > 1 && Type >= 1 && Type <= 3, "iDsd0 > 1 && iDsd1 > 1 && Type >= 1 && Type <= 3");
/*
    // check cache
    iThis = Ifd_ManHashLookup( p, iDsd0, iDsd1, iDsdC, Type );
    if ( iThis != -1 )
        return Abc_Var2Lit( iThis, fCompl );
*/
    // create new entry
    if Type == 3 {
        iThis = Ifd_ManHashFindOrAdd( p, iDsd0, iDsd1, iDsdC, Type );
        return Abc_Var2Lit( iThis, fCompl );
    }
    assert(iDsdC == -1, "iDsdC == -1");
    Vec_IntClear( p.vSuper );
    Ifd_ManOperSuper_rec( p, iDsd0, Type, p.vSuper );
    Ifd_ManOperSuper_rec( p, iDsd1, Type, p.vSuper );
    Vec_IntSort( p.vSuper, 1 );
    iLit0 = Vec_IntEntry( p.vSuper, 0 );
    Vec_IntForEachEntryStart( p.vSuper, iLit1, i, 1 )
        iLit0 = Abc_Var2Lit( Ifd_ManHashFindOrAdd(p, iLit0, iLit1, -1, Type), 0 );
    assert(!Abc_LitIsCompl(iLit0), "!Abc_LitIsCompl(iLit0)");
    // insert into cache
//    if ( Vec_IntSize(p->vSuper) > 2 )
//        Ifd_ManHashInsert( p, iDsd0, iDsd1, iDsdC, Type, iLit0 );
    return Abc_LitNotCond( iLit0, fCompl );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ManFindDsd_rec(pMan *Ifd_Man_t, pStr string, p *string, pMatches *int32) int32 {
    var fCompl int32;
    if **p == '!' {
        (*p)++, fCompl = 1; }
    if **p >= 'a' && **p <= 'f' { // var
    
        assert(**p - 'a' >= 0 && **p - 'a' < 6, "**p - 'a' >= 0 && **p - 'a' < 6");
        return Abc_Var2Lit( 1, fCompl );
    }
    if **p == '(' { // and/or
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var Lit int32;
        var Res int32 = 1;
        assert(**p == '(' && *q == ')', "**p == '(' && *q == ')'");
        for  (*p)++; *p < q; (*p)++  {
            Lit = Ifd_ManFindDsd_rec( pMan, pStr, p, pMatches );
            Res = Ifd_ManOper( pMan, Res, Lit, 0, 1 );
        }
        assert(*p == q, "*p == q");
        return Abc_LitNotCond( Res, fCompl );
    }
    if **p == '[' { // xor
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var Lit, Res int32;
        assert(**p == '[' && *q == ']', "**p == '[' && *q == ']'");
        for  (*p)++; *p < q; (*p)++  {
            Lit = Ifd_ManFindDsd_rec( pMan, pStr, p, pMatches );
            Res = Ifd_ManOper( pMan, Res, Lit, 0, 2 );
        }
        assert(*p == q, "*p == q");
        return Abc_LitNotCond( Res, fCompl );
    }
    if **p == '<' { // mux
    
        var Temp [3]int32;
        var pTemp *int32 = Temp;
        var Res int32;
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(**p == '<' && *q == '>', "**p == '<' && *q == '>'");
        // derive MAX components
        for  (*p)++; *p < q; (*p)++  {
            *pTemp++ = Ifd_ManFindDsd_rec( pMan, pStr, p, pMatches );
        }
        assert(pTemp == Temp + 3, "pTemp == Temp + 3");
        assert(*p == q, "*p == q");
//        Res = (Temp[0] & Temp[1]) | (~Temp[0] & Temp[2]);
        Res = Ifd_ManOper( pMan, Temp[2], Temp[1], Temp[0], 3 );
        return Abc_LitNotCond( Res, fCompl );
    }
    assert(false, "0");
    return 0;
}
const IFM_MAX_STR =  100
const IFM_MAX_VAR =   16
var spStatic_Ifd_ManComputeMatches_pMatches [IFM_MAX_STR]int32;
func Ifd_ManComputeMatches(p string) *int32 {
    
    var pNested [IFM_MAX_VAR]int32;
    var v, nNested int32;
    for  v = 0; p[v]; v++  {
        assert(v < IFM_MAX_STR, "v < IFM_MAX_STR");
        spStatic_Ifd_ManComputeMatches_pMatches[v] = 0;
        if p[v] == '(' || p[v] == '[' || p[v] == '<' || p[v] == '{' {
            pNested[nNested++] = v;
        } else if p[v] == ')' || p[v] == ']' || p[v] == '>' || p[v] == '}' {
            spStatic_Ifd_ManComputeMatches_pMatches[pNested[--nNested]] = v; }
        assert(nNested < IFM_MAX_VAR, "nNested < IFM_MAX_VAR");
    }
    assert(nNested == 0, "nNested == 0");
    return spStatic_Ifd_ManComputeMatches_pMatches;
}
func Ifd_ManFindDsd(pMan *Ifd_Man_t, p string) int32 {
    var Res int32;
    if *p == '0' && *(p+1) == 0 {
        Res = 0;
    } else if *p == '1' && *(p+1) == 0 {
        Res = 1;
    } else {
        Res = Ifd_ManFindDsd_rec( pMan, p, &p, Ifd_ManComputeMatches(p) ); }
    assert(*++p == 0, "*++p == 0");
    return Res;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ManDsdTest2() {
    var p string  = "(abc)";
//    char * q = "(a[bc])";
//    char * r = "[<abc>(def)]";
    var pMan *Ifd_Man_t  = Ifd_ManStart();
    var iLit int32  = Ifd_ManFindDsd( pMan, p );
    Ifd_ObjPrint( pMan, iLit );
    Ifd_ManStop( pMan );
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ManDsdTruths(nVars int32) *Vec_Wrd_t {
    var fUseMux int32  = 1;
    var vTruths *Vec_Wrd_t;
    var pMan *Ifd_Man_t  = Ifd_ManStart();
    var pLeaf0, pLeaf1, pLeaf2 *Ifd_Obj_t;
    var v, i, j, k, c0, c1, c2 int32;
    for  v = 2; v <= nVars; v++  {
        // create ANDs/XORs
        for  i = 1; i < v; i++  {
        for  j = 1; j < v; j++  {
        if i + j == v {
            // inlining of macro Ifd_ManForEachNodeWithSupp
            for  c0 = Vec_IntEntry(pMan.vMarks, i); (c0 < Vec_IntEntry(pMan.vMarks, i+1)) && (pLeaf0 = Ifd_ManObj(pMan, c0)); c0++  {
            // inlining of macro Ifd_ManForEachNodeWithSupp
            for  c1 = Vec_IntEntry(pMan.vMarks, j); (c1 < Vec_IntEntry(pMan.vMarks, j+1)) && (pLeaf1 = Ifd_ManObj(pMan, c1)); c1++  {
                assert(int32(pLeaf0.nSupp) == i, "int32(pLeaf0.nSupp) == i");
                assert(int32(pLeaf1.nSupp) == j, "int32(pLeaf1.nSupp) == j");
                Ifd_ManOper( pMan, Abc_Var2Lit(c0, 0), Abc_Var2Lit(c1, 0), -1, 1 );
                if !pLeaf1.fWay {
                Ifd_ManOper( pMan, Abc_Var2Lit(c0, 0), Abc_Var2Lit(c1, 1), -1, 1 ); }
                if !pLeaf0.fWay {
                Ifd_ManOper( pMan, Abc_Var2Lit(c0, 1), Abc_Var2Lit(c1, 0), -1, 1 ); }
                if !pLeaf0.fWay && !pLeaf1.fWay {
                Ifd_ManOper( pMan, Abc_Var2Lit(c0, 1), Abc_Var2Lit(c1, 1), -1, 1 ); }
                Ifd_ManOper( pMan, Abc_Var2Lit(c0, 0), Abc_Var2Lit(c1, 0), -1, 2 );
            }
            }
        }
        }
        }
        // create MUX
        if fUseMux != 0 {
        for  i = 1; i < v-1; i++  {
        for  j = 1; j < v-1; j++  {
        for  k = 1; k < v-1; k++  {
        if i + j + k == v {
            // inlining of macro Ifd_ManForEachNodeWithSupp
            for  c0 = Vec_IntEntry(pMan.vMarks, i); (c0 < Vec_IntEntry(pMan.vMarks, i+1)) && (pLeaf0 = Ifd_ManObj(pMan, c0)); c0++  {
            // inlining of macro Ifd_ManForEachNodeWithSupp
            for  c1 = Vec_IntEntry(pMan.vMarks, j); (c1 < Vec_IntEntry(pMan.vMarks, j+1)) && (pLeaf1 = Ifd_ManObj(pMan, c1)); c1++  {
            // inlining of macro Ifd_ManForEachNodeWithSupp
            for  c2 = Vec_IntEntry(pMan.vMarks, k); (c2 < Vec_IntEntry(pMan.vMarks, k+1)) && (pLeaf2 = Ifd_ManObj(pMan, c2)); c2++  {
                assert(int32(pLeaf0.nSupp) == i, "int32(pLeaf0.nSupp) == i");
                assert(int32(pLeaf1.nSupp) == j, "int32(pLeaf1.nSupp) == j");
                assert(int32(pLeaf2.nSupp) == k, "int32(pLeaf2.nSupp) == k");
//printf( "%d %d %d   ", i, j, k );
//printf( "%d %d %d\n", Ifd_ObjId(pMan, pLeaf0), Ifd_ObjId(pMan, pLeaf1), Ifd_ObjId(pMan, pLeaf2) );
                if pLeaf2.fWay && c0 < c1 {
                    continue; }
                Ifd_ManOper( pMan, Abc_Var2Lit(c0, 0), Abc_Var2Lit(c1, 0), Abc_Var2Lit(c2, 0), 3 );
                if !pLeaf0.fWay && !pLeaf1.fWay {
                Ifd_ManOper( pMan, Abc_Var2Lit(c0, 1), Abc_Var2Lit(c1, 0), Abc_Var2Lit(c2, 0), 3 ); }
            }
            }
            }
        }
        }
        }
        } }
        // bookmark
        Vec_IntPush( pMan.vMarks, pMan.nObjs );
    }
    Ifd_ManTruthAll( pMan );
    Mpm_ComputeCnfSizeAll( pMan );
//    Ifd_ManPrint( pMan );
    vTruths = pMan.vTruths; pMan.vTruths = nil;
    Ifd_ManStop( pMan );
    return vTruths;
}


/**Function*************************************************************

  Synopsis    [Generating the guided array for minimal permutations.]

  Description [http://icodesnip.com/search/johnson%20trotter/]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ManDsdPermPrint(perm *int32, size int32) {
    var i int32;
    for  i = 0; i < size; i++  {
        fmt.Printf("%d", perm[i]);
    }
    fmt.Printf("\n");
}
func Ifd_ManDsdPermJT(n int32) *Vec_Int_t {
    var vGuide *Vec_Int_t  = Vec_IntAlloc( 100 );
    var array *int32;
    var dir *int32;
    var tmp int32;
    var tmp2 int32;
    var i int32;
    var max int32;
    array = (int*)malloc(sizeof(int) * n);
    dir = (int*)calloc(n, sizeof(int));
    for i = 0; i < n; i++ {
    array[i] = i;
    }
    max = n - 1;
    if n != 1 {
    for 
    {
//        Ifd_ManDsdPermPrint(array, n);
        tmp = array[max];
        tmp2 = dir[max];
        if !dir[max] {
            i = max - 1;
        } else {
            i = max + 1;
        }
        array[max] = array[i];
        array[i] = tmp;
        Vec_IntPush( vGuide, Abc_MinInt(max, i) );
        dir[max] = dir[i];
        dir[i] = tmp2;
        for i = 0; i < n; i++ {
            if array[i] > tmp {
                dir[i] = !dir[i]; }
        }
        max = n;
        for i = 0; i < n; i++ {
            if ((!dir[i] && i != 0 && array[i] > array[i-1]) || (dir[i] && i != n-1 && array[i] > array[i+1])) && (array[i] > array[max] || max == n) {
                max = i; }
        }
    
if !(max < n) { break }
} 
     }
//    Ifd_ManDsdPermPrint(array,n);
    Vec_IntPush( vGuide, 0 );
    free(dir);
    free(array);
    return vGuide;
}
func Ifd_ManDsdTest4() int32 {
    var pPerm [6]int32 = [6]int32{ 0, 1, 2, 3, 4, 5 };
    var vGuide *Vec_Int_t  = Ifd_ManDsdPermJT( 6 );
    var i, Entry int32;
    Vec_IntForEachEntry( vGuide, Entry, i )
    {
        ABC_SWAP( int, pPerm[Entry], pPerm[Entry+1] );
        Ifd_ManDsdPermPrint( pPerm, 6 );
    }
    Vec_IntFree( vGuide );
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_Truth6SwapAdjacent(t word, iVar int32) word {
    masks := [5][3]word{
        {0x9999999999999999, 0x2222222222222222, 0x4444444444444444},
        {0xC3C3C3C3C3C3C3C3, 0x0C0C0C0C0C0C0C0C, 0x3030303030303030},
        {0xF00FF00FF00FF00F, 0x00F000F000F000F0, 0x0F000F000F000F00},
        {0xFF0000FFFF0000FF, 0x0000FF000000FF00, 0x00FF000000FF0000},
        {0xFFFF00000000FFFF, 0x00000000FFFF0000, 0x0000FFFF00000000},
    }
    assert(iVar >= 0 && iVar < 5, "valid adjacent variable")
    shift := uint(1 << iVar)
    return t & masks[iVar][0] | (t & masks[iVar][1]) << shift | (t & masks[iVar][2]) >> shift
}
func Extra_Truth6ChangePhase(t word, iVar int32) word {
    masks := [6]word{0xAAAAAAAAAAAAAAAA, 0xCCCCCCCCCCCCCCCC, 0xF0F0F0F0F0F0F0F0, 0xFF00FF00FF00FF00, 0xFFFF0000FFFF0000, 0xFFFFFFFF00000000}
    assert(iVar >= 0 && iVar < 6, "valid phase variable")
    shift := uint(1 << iVar)
    return (t &^ masks[iVar]) << shift | (t & masks[iVar]) >> shift
}
func Extra_Truth6AllConfigs2(t word, pComp *int32, pPerm *int32, nVars int32) *Vec_Wrd_t {
    var nPerms int32  = Extra_Factorial( nVars );
    var nSwaps int32  = (1 << nVars);
    var vTruths *Vec_Wrd_t  = Vec_WrdStart( nPerms * (1 << (nVars+1)) );
    var tCur, tTemp1, tTemp2 word;
    var i, p, c int32;
    for  i = 0; i < 2; i++  {
        if i != 0 {
            tCur = t;
        } else {
            tCur = ~t;
        }
        tTemp1 = tCur;
        for  p = 0; p < nPerms; p++  {
            tTemp2 = tCur;
            for  c = 0; c < nSwaps; c++  {
                Vec_WrdWriteEntry( vTruths, (p << (nVars+1))|(i << nVars)|c, tCur );
                tCur = Extra_Truth6ChangePhase( tCur, pComp[c] );
            }
            assert(tTemp2 == tCur, "tTemp2 == tCur");
            tCur = Extra_Truth6SwapAdjacent( tCur, pPerm[p] );
        }
        assert(tTemp1 == tCur, "tTemp1 == tCur");
    }
    if t {
        var i int32;
        var Truth word;
        Vec_WrdForEachEntry( vTruths, Truth, i )
            assert(Truth, "Truth");
    }
    return vTruths;
}
func Extra_Truth6AllConfigs(t word, pComp *int32, pPerm *int32, nVars int32) *Vec_Wrd_t {
    var nPerms int32  = Extra_Factorial( nVars );
    var nSwaps int32  = (1 << nVars);
    var vTruths *Vec_Wrd_t  = Vec_WrdStart( nPerms * nSwaps );
    var tCur word = t;
    var tTemp1 word;
    var tTemp2 word;
    var p, c, Config int32;

    tTemp1 = tCur;
    for  p = 0; p < nPerms; p++  {
        tCur = Extra_Truth6SwapAdjacent( tCur, pPerm[p] );
        Config = 0;
        tTemp2 = tCur;
        for  c = 0; c < nSwaps; c++  {
            Vec_WrdWriteEntry( vTruths, (p << nVars)|Config, tCur );
            tCur = Extra_Truth6ChangePhase( tCur, pComp[c] );
            Config ^= (1 << pComp[c]);
        }
        assert(Config == 0, "Config == 0");
        assert(tTemp2 == tCur, "tTemp2 == tCur");
    }
    assert(tTemp1 == tCur, "tTemp1 == tCur");

    if t {
        var i int32;
        var Truth word;
        Vec_WrdForEachEntry( vTruths, Truth, i )
            assert(Truth, "Truth");
    }
    return vTruths;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ComputeSignature(uTruth word, pCounts []int32) {
    var v, Pos, Neg, Xor int32;
    for  v = 0; v < 6; v++  {
        Neg = Abc_TtCountOnes( Abc_Tt6Cofactor0(uTruth, v) ) / 2;
        Pos = Abc_TtCountOnes( Abc_Tt6Cofactor1(uTruth, v) ) / 2;
        Xor = Abc_TtCountOnes( Abc_Tt6Cofactor0(uTruth, v) ^ Abc_Tt6Cofactor1(uTruth, v) ) / 2;
        if Pos <= Neg {
            pCounts[v] = (Pos << 20) | (Neg << 10) | Xor;
        } else {
            pCounts[v] = (Neg << 20) | (Pos << 10) | Xor; }
    }
    Vec_IntSelectSort( pCounts, 6 );
}
func Ifd_ManDsdTest33() int32 {
    var nVars int32  = 6;
    var vTruths *Vec_Wrd_t  = Ifd_ManDsdTruths( nVars );
    var i int32;
    var v int32;
    var pCounts [6]int32;
    var uTruth word;
    Vec_WrdForEachEntry( vTruths, uTruth, i )
    {
        Ifd_ComputeSignature( uTruth, pCounts );
        // print
        fmt.Printf("%5d :  ", i); 
        for  v = 0; v < 6; v++  {
            fmt.Printf("%2d %2d %2d   ", (pCounts[v] >> 20) & 0xFF, (pCounts[v] >> 10) & 0xFF, (pCounts[v] >> 0) & 0xFF);
        }
        fmt.Printf("  ");
        Kit_DsdPrintFromTruth( (unsigned *)&uTruth, nVars );
        fmt.Printf("\n");
    }
    Vec_WrdFree( vTruths );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ifd_ManDsdTest() int32 {
    var nVars int32  = 6;
    var pFile *os.File;
    var pFileName string // was [32]byte in C

    var vTruths *Vec_Wrd_t  = Ifd_ManDsdTruths( nVars );
    var vVariants *Vec_Wrd_t;
    var vUniques *Vec_Int_t;
    var vCompls *Vec_Int_t;
    var vTruthRes *Vec_Wrd_t  = Vec_WrdAlloc( 4000000 );
    var vConfgRes *Vec_Int_t  = Vec_IntAlloc( 4000000 );
    var pComp, pPerm *int32;
    var Truth, Variant word;
    var i, k, Uniq, Runner, Counter int32;
    assert(nVars >= 3 && nVars <= 6, "nVars >= 3 && nVars <= 6");
    assert(Vec_WrdSize(vTruths) < (1<<10), "Vec_WrdSize(vTruths) < (1<<10)");
    vCompls = Vec_IntAlloc( 720 * 64 );
    pComp = Extra_GreyCodeSchedule( nVars );
    pPerm = Extra_PermSchedule( nVars );
    Vec_WrdForEachEntry( vTruths, Truth, i )
    {
        vVariants = Extra_Truth6AllConfigs( Truth, pComp, pPerm, nVars );
        // save compl bits
        Vec_IntClear( vCompls );
        Vec_WrdForEachEntry( vVariants, Variant, k )
        {
            Vec_IntPush( vCompls, int32((Variant & 1)) );
            Vec_WrdWriteEntry( vVariants, k, Variant & 1 ? ~Variant : Variant );
        }
        // uniqify
        vUniques = Hsh_WrdManHashArray( vVariants, 1 );
        Runner = 0;
        Vec_IntForEachEntry( vUniques, Uniq, k )
            if Runner == Uniq {
                Variant = Vec_WrdEntry(vVariants, k);
                assert((Variant & 1) == 0, "(Variant & 1) == 0");
                Vec_WrdPush( vTruthRes, Variant );
                Vec_IntPush( vConfgRes, (i << 17)|(Vec_IntEntry(vCompls, k) << 16)|k );
                Runner++;
            }
        Vec_IntUniqify( vUniques );
        assert(Runner == Vec_IntSize(vUniques), "Runner == Vec_IntSize(vUniques)");
        Counter += Vec_IntSize(vUniques);
//printf( "%5d : ", i ); Kit_DsdPrintFromTruth( &Truth, nVars ), printf( "  " ), Vec_IntPrint( vUniques ), printf( "\n" );
        Vec_IntFree( vUniques );
        Vec_WrdFree( vVariants );
    } 
    Vec_IntFree( vCompls );
    Vec_WrdFree( vTruths );
    ABC_FREE( pPerm );
    ABC_FREE( pComp );
    fmt.Printf("Total = %d.\n", Counter);
    assert(Vec_WrdSize(vTruthRes) == Counter, "Vec_WrdSize(vTruthRes) == Counter");
    // write the data into a file
    pFileName = fmt.Sprintf("dsdfuncs%d.dat", nVars);
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fwrite( Vec_WrdArray(vTruthRes), sizeof(word), Vec_WrdSize(vTruthRes), pFile );
    fwrite( Vec_IntArray(vConfgRes), sizeof(int), Vec_IntSize(vConfgRes), pFile );
    _ = (pFile).Close();
    fmt.Printf("File \"%s\" with %d 6-input functions has been written out.\n", pFileName, Vec_IntSize(vConfgRes));
    Vec_WrdFree( vTruthRes );
    Vec_IntFree( vConfgRes );
    return 1;
}

func Ifd_ManDsdTest55() int32 {
    var clk abctime  = Abc_Clock();
    var pFile *os.File;
    var pFileName string  = "dsdfuncs6.dat";
    var RetValue int32;
    var size int32 = Extra_FileSize( pFileName ) / 12;  // 2866420
    var vTruthRes *Vec_Wrd_t  = Vec_WrdAlloc( size + 1 );
    var vConfgRes *Vec_Int_t  = Vec_IntAlloc( size );
    var pHash *Hsh_IntMan_t;

    pFile = os.Open(pFileName);
    RetValue = fread( Vec_WrdArray(vTruthRes), sizeof(word), size, pFile );
    RetValue = fread( Vec_IntArray(vConfgRes), sizeof(int), size, pFile );
    vTruthRes.nSize = size;
    vConfgRes.nSize = size;
    // create hash table
    pHash = Hsh_WrdManHashArrayStart( vTruthRes, 1 );
    // experiment with functions

    // cleanup
    Hsh_IntManStop( pHash );
    Vec_WrdFree( vTruthRes );
    Vec_IntFree( vConfgRes );
    Abc_PrintTime( 1, "Reading file", Abc_Clock() - clk );
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

