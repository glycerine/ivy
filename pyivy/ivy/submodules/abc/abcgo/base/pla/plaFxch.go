package pla

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [plaFxch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaFxch.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pla.h"
//#include "misc/vec/vecHash.h"
//#include "misc/vec/vecQue.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Fxch_Obj_t struct {
    Table uint32; /* C bitfield: 30 bits; review packing and truncation */
    MarkN uint32; /* C bitfield: 1 bits; review packing and truncation */
    MarkP uint32; /* C bitfield: 1 bits; review packing and truncation */
    Next             int32;
    Prev             int32;
    Cube             int32;
}

   
type Fxch_Man_t struct {
    // user's data
    vCubes       Vec_Wec_t;     // cube -> lit
    // internal data
    vLits       Vec_Wec_t;      // lit -> cube
    vRands       Vec_Int_t;     // random numbers for each literal
    vCubeLinks       Vec_Int_t; // first link for each cube
    pBins     *Fxch_Obj_t;      // hash table (lits -> cube + lit)
    vHash  *Hash_IntMan_t;      // divisor hash table
    vPrio      *Vec_Que_t;      // priority queue for divisors by weight
    vWeights       Vec_Flt_t;   // divisor weights
    vPairs       Vec_Wec_t;     // cube pairs for each div
    vDivs       Vec_Wrd_t;      // extracted divisors
    // temporary data 
    vCubesS       Vec_Int_t;    // cube pairs for single
    vCubesD       Vec_Int_t;    // cube pairs for double
    vCube1       Vec_Int_t;     // first cube
    vCube2       Vec_Int_t;     // second cube
    // statistics 
    timeStart         abctime;  // starting time
    SizeMask             int32;   // hash table mask
    nVars             int32;      // original problem variables
    nLits             int32;      // the number of SOP literals
    nCompls             int32;    // the number of complements
    nPairsS             int32;    // number of lit pairs
    nPairsD             int32;    // number of cube pairs
}

// Original macro:
// #define Fxch_ManForEachCubeVec( vVec, vCubes, vCube, i )           \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((vCube) = Vec_WecEntry(vCubes, Vec_IntEntry(vVec, i))); i++ )


func Fxch_ManCube(p *Fxch_Man_t, hCube int32) *Vec_Int_t { return Vec_WecEntry(&p.vCubes, Pla_CubeNum(hCube)); }


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the current state of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_ManWriteBlif(pFileName string, vCubes *Vec_Wec_t, vDivs *Vec_Wrd_t) {
    // find the number of original variables
    var nVarsInit int32  = Vec_WrdCountZero(vDivs);
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
    } else {
        //char * pLits = "-01?";
        var vStr *Vec_Str_t;
        var vCube *Vec_Int_t;
        var i, k, Lit int32;
        var Div word;
        // comment
        fmt.Fprintf(pFile, "# BLIF file written via PLA package in ABC on ");
        fmt.Fprintf(pFile, "%s", Extra_TimeStamp());
        fmt.Fprintf(pFile, "\n\n");
        // header
        fmt.Fprintf(pFile, ".model %s\n", pFileName);
        fmt.Fprintf(pFile, ".inputs");
        for  i = 0; i < nVarsInit; i++  {
            fmt.Fprintf(pFile, " i%d", i);
        }
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, ".outputs o");
        fmt.Fprintf(pFile, "\n");
        // SOP header
        fmt.Fprintf(pFile, ".names");
        for  i = 0; i < Vec_WrdSize(vDivs); i++  {
            fmt.Fprintf(pFile, " i%d", i);
        }
        fmt.Fprintf(pFile, " o\n");
        // SOP cubes
        vStr = Vec_StrStart( Vec_WrdSize(vDivs) + 1 );
        Vec_WecForEachLevel( vCubes, vCube, i )
        {
            if !Vec_IntSize(vCube) {
                continue; }
            for  k = 0; k < Vec_WrdSize(vDivs); k++  {
                Vec_StrWriteEntry( vStr, k, '-' );
            }
            Vec_IntForEachEntry( vCube, Lit, k )
                Vec_StrWriteEntry( vStr, Abc_Lit2Var(Lit), byte((Abc_LitIsCompl(Lit) ? '0' : '1')) );
            fmt.Fprintf(pFile, "%s 1\n", Vec_StrArray(vStr));
        }
        Vec_StrFree( vStr );
        // divisors
        Vec_WrdForEachEntryStart( vDivs, Div, i, nVarsInit )
        {
            var pLits [2]int32 = [2]int32{ int32((Div & 0xFFFFFFFF)), int32((Div >> 32)) };
            fmt.Fprintf(pFile, ".names");
            fmt.Fprintf(pFile, " i%d", Abc_Lit2Var(pLits[0]));
            fmt.Fprintf(pFile, " i%d", Abc_Lit2Var(pLits[1]));
            fmt.Fprintf(pFile, " i%d\n", i);
            fmt.Fprintf(pFile, "%d%d 1\n", !Abc_LitIsCompl(pLits[0]), !Abc_LitIsCompl(pLits[1]));
        }
        fmt.Fprintf(pFile, ".end\n\n");
        _ = (pFile).Close();
        fmt.Printf("Written BLIF file \"%s\".\n", pFileName);
    }
}

/**Function*************************************************************

  Synopsis    [Starting the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_ManStart(vCubes *Vec_Wec_t, vLits *Vec_Wec_t) *Fxch_Man_t {
    var vCube *Vec_Int_t;  var i, LogSize int32;
    var p *Fxch_Man_t  = new(Fxch_Man_t);
    p.vCubes = *vCubes;
    p.vLits  = *vLits;
    p.nVars  = Vec_WecSize(vLits)/2;
    p.nLits  = 0;
    // random numbers
    Gia_ManRandom( 1 );
    Vec_IntGrow( &p.vRands, 2*p.nVars );
    for  i = 0; i < 2*p.nVars; i++  {
        Vec_IntPush( &p.vRands, Gia_ManRandom(0) & 0x3FFFFFF );
    } // assert( LogSize <= 26 );
    // create cube links
    Vec_IntGrow( &p.vCubeLinks, Vec_WecSize(&p.vCubes) );
    Vec_WecForEachLevel( vCubes, vCube, i )
    {
        Vec_IntPush( &p.vCubeLinks, p.nLits+1 );
        p.nLits += Vec_IntSize(vCube);
    }
    assert(Vec_IntSize(&p.vCubeLinks) == Vec_WecSize(&p.vCubes), "Vec_IntSize(&p.vCubeLinks) == Vec_WecSize(&p.vCubes)");
    // create table
    LogSize = Abc_Base2Log( p.nLits+1 );
    assert(LogSize <= 26, "LogSize <= 26");
    p.SizeMask = (1 << LogSize) - 1;
    p.pBins = ABC_CALLOC( Fxch_Obj_t, p.SizeMask + 1 );
    assert(p.nLits+1 < p.SizeMask+1, "p.nLits+1 < p.SizeMask+1");
    // divisor weights and cube pairs
    Vec_FltGrow( &p.vWeights, 1000 );
    Vec_FltPush( &p.vWeights, -1 );
    Vec_WecGrow( &p.vPairs, 1000 );
    Vec_WecPushLevel( &p.vPairs );
    // prepare divisors
    Vec_WrdGrow( &p.vDivs, p.nVars + 1000 );
    Vec_WrdFill( &p.vDivs, p.nVars, 0 );
    return p;
}
func Fxch_ManStop(p *Fxch_Man_t) {
    Vec_WecErase( &p.vCubes );
    Vec_WecErase( &p.vLits );
    Vec_IntErase( &p.vRands );
    Vec_IntErase( &p.vCubeLinks );
    Hash_IntManStop( p.vHash );
    Vec_QueFree( p.vPrio );
    Vec_FltErase( &p.vWeights );
    Vec_WecErase( &p.vPairs );
    Vec_WrdErase( &p.vDivs );
    Vec_IntErase( &p.vCubesS );
    Vec_IntErase( &p.vCubesD );
    Vec_IntErase( &p.vCube1 );
    Vec_IntErase( &p.vCube2 );
    ABC_FREE( p.pBins );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_TabCompare(p *Fxch_Man_t, hCube1 int32, hCube2 int32) int32 {
    var vCube1 *Vec_Int_t  = Fxch_ManCube( p, hCube1 );
    var vCube2 *Vec_Int_t  = Fxch_ManCube( p, hCube2 );
    if !Vec_IntSize(vCube1) || !Vec_IntSize(vCube2) || Vec_IntSize(vCube1) != Vec_IntSize(vCube2) {
        return 0; }
    Vec_IntClear( &p.vCube1 );
    Vec_IntClear( &p.vCube2 );
    Vec_IntAppendSkip( &p.vCube1, vCube1, Pla_CubeLit(hCube1) );
    Vec_IntAppendSkip( &p.vCube2, vCube2, Pla_CubeLit(hCube2) );
    return Vec_IntEqual( &p.vCube1, &p.vCube2 );
}
func Fxch_CompressCubes(p *Fxch_Man_t, vLit2Cube *Vec_Int_t) {
    var i, hCube, k int32;
    Vec_IntForEachEntry( vLit2Cube, hCube, i )
        if ( Vec_IntSize(Vec_WecEntry(&p.vCubes, hCube)) > 0 )
            Vec_IntWriteEntry( vLit2Cube, k++, hCube );
    Vec_IntShrink( vLit2Cube, k );
}
func Fxch_CollectSingles(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vArr *Vec_Int_t) int32 {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    var pBeg1New *int32  = vArr1.pArray;
    var pBeg2New *int32  = vArr2.pArray;
    Vec_IntClear( vArr );
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if Pla_CubeNum(*pBeg1) == Pla_CubeNum(*pBeg2) {
            Vec_IntPushTwo( vArr, *pBeg1, *pBeg2 );
pBeg1++;
pBeg2++;
        } else if *pBeg1 < *pBeg2 {
            *pBeg1New++ = *pBeg1++;
        } else { 
            *pBeg2New++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg1New++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pBeg2New++ = *pBeg2++;
    }
    Vec_IntShrink( vArr1, pBeg1New - vArr1.pArray );
    Vec_IntShrink( vArr2, pBeg2New - vArr2.pArray );
    return Vec_IntSize(vArr);
}
func Fxch_CollectDoubles(p *Fxch_Man_t, vPairs *Vec_Int_t, vRes *Vec_Int_t, Lit0 int32, Lit1 int32) {
    var i, hCube1, hCube2 int32;
    Vec_IntClear( vRes );
    Vec_IntForEachEntryDouble( vPairs, hCube1, hCube2, i )
        if ( Fxch_TabCompare(p, hCube1, hCube2) && 
             Vec_IntEntry(Fxch_ManCube(p, hCube1), Pla_CubeLit(hCube1)) == Lit0 && 
             Vec_IntEntry(Fxch_ManCube(p, hCube2), Pla_CubeLit(hCube2)) == Lit1 )
            Vec_IntPushTwo( vRes, hCube1, hCube2 );
    Vec_IntClear( vPairs );
    // order pairs in the increasing order of the first cube
    //Vec_IntSortPairs( vRes );
}
func Fxch_CompressLiterals2(p *Vec_Int_t, iInd1 int32, iInd2 int32) {
    var i, Lit, k int32;
    assert(iInd1 >= 0 && iInd1 < Vec_IntSize(p), "iInd1 >= 0 && iInd1 < Vec_IntSize(p)");
    if iInd2 != -1 {
    assert(iInd1 >= 0 && iInd1 < Vec_IntSize(p), "iInd1 >= 0 && iInd1 < Vec_IntSize(p)"); }
    Vec_IntForEachEntry( p, Lit, i )
        if ( i != iInd1 && i != iInd2 )
            Vec_IntWriteEntry( p, k++, Lit );
    Vec_IntShrink( p, k );
}
func Fxch_CompressLiterals(p *Vec_Int_t, iLit1 int32, iLit2 int32) {
    var i, Lit, k int32;
    Vec_IntForEachEntry( p, Lit, i )
        if ( Lit != iLit1 && Lit != iLit2 )
            Vec_IntWriteEntry( p, k++, Lit );
    assert(Vec_IntSize(p) == k + 2, "Vec_IntSize(p) == k + 2");
    Vec_IntShrink( p, k );
}
func Fxch_FilterCubes(p *Fxch_Man_t, vCubesS *Vec_Int_t, Lit0 int32, Lit1 int32) {
    var vCube *Vec_Int_t;
    var i, k, Lit, iCube, n int32;
    var fFound0, fFound1 int32;
    Vec_IntForEachEntry( vCubesS, iCube, i )
    {
        vCube = Vec_WecEntry( &p.vCubes, iCube );
        fFound0 = fFound1 = 0;
        Vec_IntForEachEntry( vCube, Lit, k )
        {
            if Lit == Lit0 {
                fFound0 = 1;
            } else if Lit == Lit1 {
                fFound1 = 1; }
        }
        if fFound0 && (fFound1 != 0) {
            Vec_IntWriteEntry( vCubesS, n++, Pla_CubeHandle(iCube, 255) ); }
    }
    Vec_IntShrink( vCubesS, n );
}


/**Function*************************************************************

  Synopsis    [Divisor addition removal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_DivisorAdd(p *Fxch_Man_t, Lit0 int32, Lit1 int32, Weight int32) int32 {
    var iDiv int32;
    assert(Lit0 != Lit1, "Lit0 != Lit1");
    if Lit0 < Lit1 {
        iDiv = Hash_Int2ManInsert( p.vHash, Lit0, Lit1, 0 );
    } else {
        iDiv = Hash_Int2ManInsert( p.vHash, Lit1, Lit0, 0 ); }
    if iDiv == Vec_FltSize(&p.vWeights) {
        Vec_FltPush( &p.vWeights, -2 );
        Vec_WecPushLevel( &p.vPairs );
        assert(Vec_FltSize(&p.vWeights) == Vec_WecSize(&p.vPairs), "Vec_FltSize(&p.vWeights) == Vec_WecSize(&p.vPairs)");
    }
    Vec_FltAddToEntry( &p.vWeights, iDiv, Weight );
    if p.vPrio {
        if Vec_QueIsMember(p.vPrio, iDiv) {
            Vec_QueUpdate( p.vPrio, iDiv );
        } else { 
            Vec_QuePush( p.vPrio, iDiv ); }
        //assert( iDiv < Vec_QueSize(p->vPrio) );
    }
    return iDiv;
}
func Fxch_DivisorRemove(p *Fxch_Man_t, Lit0 int32, Lit1 int32, Weight int32) {
    var iDiv int32;
    assert(Lit0 != Lit1, "Lit0 != Lit1");
    if Lit0 < Lit1 {
        iDiv = *Hash_Int2ManLookup( p.vHash, Lit0, Lit1 );
    } else {
        iDiv = *Hash_Int2ManLookup( p.vHash, Lit1, Lit0 ); }
    assert(iDiv > 0 && iDiv < Vec_FltSize(&p.vWeights), "iDiv > 0 && iDiv < Vec_FltSize(&p.vWeights)");
    Vec_FltAddToEntry( &p.vWeights, iDiv, -Weight );
    if Vec_QueIsMember(p.vPrio, iDiv) {
        Vec_QueUpdate( p.vPrio, iDiv ); }
}

/**Function*************************************************************

  Synopsis    [Starting the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_TabBin(p *Fxch_Man_t, Value int32) *Fxch_Obj_t { return p.pBins + (Value & p.SizeMask);         }
func Fxch_TabEntry(p *Fxch_Man_t, i int32) *Fxch_Obj_t { if i != 0 {
    return p.pBins + i;
} else {
    return nil;
}                  }
func Fxch_TabEntryId(p *Fxch_Man_t, pEnt *Fxch_Obj_t) int32 { assert(pEnt > p.pBins, "pEnt > p.pBins"); return pEnt - p.pBins; }

func Fxch_TabMarkPair(p *Fxch_Man_t, i int32, j int32) {
    var pI *Fxch_Obj_t  = Fxch_TabEntry(p, i);
    var pJ *Fxch_Obj_t  = Fxch_TabEntry(p, j);
    assert(pI.Next == j, "pI.Next == j");
    assert(pJ.Prev == i, "pJ.Prev == i");
    assert(pI.MarkN == 0, "pI.MarkN == 0");
    assert(pI.MarkP == 0, "pI.MarkP == 0");
    assert(pJ.MarkN == 0, "pJ.MarkN == 0");
    assert(pJ.MarkP == 0, "pJ.MarkP == 0");
    pI.MarkN = 1;
    pJ.MarkP = 1;
}
func Fxch_TabUnmarkPair(p *Fxch_Man_t, i int32, j int32) {
    var pI *Fxch_Obj_t  = Fxch_TabEntry(p, i);
    var pJ *Fxch_Obj_t  = Fxch_TabEntry(p, j);
    assert(pI.Next == j, "pI.Next == j");
    assert(pJ.Prev == i, "pJ.Prev == i");
    assert(pI.MarkN == 1, "pI.MarkN == 1");
    assert(pI.MarkP == 0, "pI.MarkP == 0");
    assert(pJ.MarkN == 0, "pJ.MarkN == 0");
    assert(pJ.MarkP == 1, "pJ.MarkP == 1");
    pI.MarkN = 0;
    pJ.MarkP = 0;
}
func Fxch_TabInsertLink(p *Fxch_Man_t, i int32, j int32, fSkipCheck int32) {
    var pI *Fxch_Obj_t  = Fxch_TabEntry(p, i);
    var pN *Fxch_Obj_t  = Fxch_TabEntry(p, pI.Next);
    var pJ *Fxch_Obj_t  = Fxch_TabEntry(p, j);
    //assert( pJ->Cube != 0 );
    assert(pN.Prev == i, "pN.Prev == i");
    assert((fSkipCheck != 0) || pI.MarkN == 0, "fSkipCheck || pI.MarkN == 0");
    assert((fSkipCheck != 0) || pN.MarkP == 0, "fSkipCheck || pN.MarkP == 0");
    assert((fSkipCheck != 0) || pJ.MarkN == 0, "fSkipCheck || pJ.MarkN == 0");
    assert((fSkipCheck != 0) || pJ.MarkP == 0, "fSkipCheck || pJ.MarkP == 0");
    pJ.Next = pI.Next;  pI.Next = j;
    pJ.Prev = i;         pN.Prev = j;
}
func Fxch_TabExtractLink(p *Fxch_Man_t, i int32, j int32) {
    var pI *Fxch_Obj_t  = Fxch_TabEntry(p, i);
    var pJ *Fxch_Obj_t  = Fxch_TabEntry(p, j);
    var pN *Fxch_Obj_t  = Fxch_TabEntry(p, pJ.Next);
    //assert( pJ->Cube != 0 );
    assert(pI.Next == j, "pI.Next == j");
    assert(pJ.Prev == i, "pJ.Prev == i");
    assert(pN.Prev == j, "pN.Prev == j");
    assert(pI.MarkN == 0, "pI.MarkN == 0");
    assert(pJ.MarkP == 0, "pJ.MarkP == 0");
    assert(pJ.MarkN == 0, "pJ.MarkN == 0");
    assert(pN.MarkP == 0, "pN.MarkP == 0");
    pI.Next = pJ.Next;  pJ.Next = 0;
    pN.Prev = pJ.Prev;  pJ.Prev = 0;
}
func Fxch_TabInsert(p *Fxch_Man_t, iLink int32, Value int32, hCube int32) {
    var iEnt int32;
    var iDiv int32;
    var Lit0 int32;
    var Lit1 int32;
    var fStart int32 = 1;
    var pEnt *Fxch_Obj_t;
    var pBin *Fxch_Obj_t   = Fxch_TabBin( p, Value );
    var pCell *Fxch_Obj_t  = Fxch_TabEntry( p, iLink );
    assert(pCell.MarkN == 0, "pCell.MarkN == 0");
    assert(pCell.MarkP == 0, "pCell.MarkP == 0");
    assert(pCell.Cube == 0, "pCell.Cube == 0");
    pCell.Cube = hCube;
    if pBin.Table == 0 {
        pBin.Table = pCell.Next = pCell.Prev = iLink;
        return;
    }
    // find equal cubes
    for  iEnt = pBin.Table; iEnt != int32(pBin.Table) || (fStart != 0); iEnt, fStart = pEnt.Next, 0  {
        pEnt = Fxch_TabBin( p, iEnt );
        if pEnt.MarkN || pEnt.MarkP || !(Fxch_TabCompare(p, pEnt.Cube, hCube) != 0) {
            continue; }
        Fxch_TabInsertLink( p, iEnt, iLink, 0 );
        Fxch_TabMarkPair( p, iEnt, iLink );
        // get literals
        Lit0 = Vec_IntEntry( Fxch_ManCube(p, hCube),      Pla_CubeLit(hCube) );
        Lit1 = Vec_IntEntry( Fxch_ManCube(p, pEnt.Cube), Pla_CubeLit(pEnt.Cube) );
        // increment divisor weight
        iDiv = Fxch_DivisorAdd( p, Abc_LitNot(Lit0), Abc_LitNot(Lit1), Vec_IntSize(Fxch_ManCube(p, hCube)) );
        // add divisor pair
        assert(iDiv < Vec_WecSize(&p.vPairs), "iDiv < Vec_WecSize(&p.vPairs)");
        if Lit0 < Lit1 {
            Vec_WecPush( &p.vPairs, iDiv, hCube );
            Vec_WecPush( &p.vPairs, iDiv, pEnt.Cube );
        
        } else {
            Vec_WecPush( &p.vPairs, iDiv, pEnt.Cube );
            Vec_WecPush( &p.vPairs, iDiv, hCube );
        }
        p.nPairsD++;
        return;
    }
    assert(iEnt == int32(pBin.Table), "iEnt == int32(pBin.Table)");
    pEnt = Fxch_TabBin( p, iEnt );
    Fxch_TabInsertLink( p, pEnt.Prev, iLink, 1 );
}
func Fxch_TabExtract(p *Fxch_Man_t, iLink int32, Value int32, hCube int32) {
    var Lit0, Lit1 int32;
    var pPair *Fxch_Obj_t  = nil;
    var pBin *Fxch_Obj_t   = Fxch_TabBin( p, Value );
    var pLink *Fxch_Obj_t  = Fxch_TabEntry( p, iLink );
    assert(pLink.Cube == hCube, "pLink.Cube == hCube");
    if pLink.MarkN {
        pPair = Fxch_TabEntry( p, pLink.Next );
        Fxch_TabUnmarkPair( p, iLink, pLink.Next );
    
    } else if pLink.MarkP {
        pPair = Fxch_TabEntry( p, pLink.Prev );
        Fxch_TabUnmarkPair( p, pLink.Prev, iLink );
    }
    if int32(pBin.Table) == iLink {
        if pLink.Next != iLink {
            pBin.Table = pLink.Next;
        } else {
            pBin.Table = 0;
        } }
    if pLink.Next == iLink {
        assert(pLink.Prev == iLink, "pLink.Prev == iLink");
        pLink.Next = pLink.Prev = 0;
    
    } else {
        Fxch_TabExtractLink( p, pLink.Prev, iLink ); }
    pLink.Cube = 0;
    if pPair == nil {
        return; }
    assert((Fxch_TabCompare(p, pPair.Cube, hCube) != 0), "Fxch_TabCompare(p, pPair.Cube, hCube)");
    // get literals
    Lit0 = Vec_IntEntry( Fxch_ManCube(p, hCube),       Pla_CubeLit(hCube) );
    Lit1 = Vec_IntEntry( Fxch_ManCube(p, pPair.Cube), Pla_CubeLit(pPair.Cube) );
    // decrement divisor weight
    Fxch_DivisorRemove( p, Abc_LitNot(Lit0), Abc_LitNot(Lit1), Vec_IntSize(Fxch_ManCube(p, hCube)) );
    p.nPairsD--;
}

/**Function*************************************************************

  Synopsis    [Starting the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_TabSingleDivisors(p *Fxch_Man_t, iCube int32, fAdd int32) int32 {
    var vCube *Vec_Int_t  = Vec_WecEntry( &p.vCubes, iCube );
    var i, k, Lit, Lit2 int32;
    if Vec_IntSize(vCube) < 2 {
        return 0; }
    Vec_IntForEachEntry( vCube, Lit, i )
    Vec_IntForEachEntryStart( vCube, Lit2, k, i+1 )
    {
        assert(Lit < Lit2, "Lit < Lit2");
        if fAdd != 0 {
            Fxch_DivisorAdd( p, Lit, Lit2, 1 );
p.nPairsS++;
        } else {
            Fxch_DivisorRemove( p, Lit, Lit2, 1 );
p.nPairsS--; }
    }
    return Vec_IntSize(vCube) * (Vec_IntSize(vCube) - 1) / 2;
}
func Fxch_TabDoubleDivisors(p *Fxch_Man_t, iCube int32, fAdd int32) int32 {
    var vCube *Vec_Int_t  = Vec_WecEntry( &p.vCubes, iCube );
    var iLinkFirst int32  = Vec_IntEntry( &p.vCubeLinks, iCube );
    var k, Lit, Value int32;
    Vec_IntForEachEntry( vCube, Lit, k )
        Value += Vec_IntEntry(&p.vRands, Lit);
    Vec_IntForEachEntry( vCube, Lit, k )
    {
        Value -= Vec_IntEntry(&p.vRands, Lit);
        if fAdd != 0 {
            Fxch_TabInsert( p, iLinkFirst + k, Value, Pla_CubeHandle(iCube, k) );
        } else {
            Fxch_TabExtract( p, iLinkFirst + k, Value, Pla_CubeHandle(iCube, k) ); }
        Value += Vec_IntEntry(&p.vRands, Lit);
    }
    return 1;
}
func Fxch_ManCreateDivisors(p *Fxch_Man_t) {
    var Weight float32;  var i int32;
    // alloc hash table
    assert(p.vHash == nil, "p.vHash == nil");
    p.vHash = Hash_IntManStart( 1000 );
    // create single-cube two-literal divisors
    for  i = 0; i < Vec_WecSize(&p.vCubes); i++  {
        Fxch_TabSingleDivisors( p, i, 1 );
    } // add - no update
    // create two-cube divisors
    for  i = 0; i < Vec_WecSize(&p.vCubes); i++  {
        Fxch_TabDoubleDivisors( p, i, 1 );
    } // add - no update
    // create queue with all divisors
    p.vPrio = Vec_QueAlloc( Vec_FltSize(&p.vWeights) );
    Vec_QueSetPriority( p.vPrio, Vec_FltArrayP(&p.vWeights) );
    Vec_FltForEachEntry( &p.vWeights, Weight, i )
        if ( Weight > 0.0 )
            Vec_QuePush( p.vPrio, i );
}


/**Function*************************************************************

  Synopsis    [Updates the data-structure when one divisor is selected.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_ManUpdate(p *Fxch_Man_t, iDiv int32) {
    var vCube1, vCube2, vLitP, vLitN *Vec_Int_t;
    //int nLitsNew = p->nLits - (int)Vec_FltEntry(&p->vWeights, iDiv);
    var i, Lit0, Lit1, hCube1, hCube2, iVarNew int32;
    //float Diff = Vec_FltEntry(&p->vWeights, iDiv) - (float)((int)Vec_FltEntry(&p->vWeights, iDiv));
    //assert( Diff > 0.0 && Diff < 1.0 );

    // get the divisor and select pivot variables
    Vec_IntPush( &p.vRands, Gia_ManRandom(0) & 0x3FFFFFF ); 
    Vec_IntPush( &p.vRands, Gia_ManRandom(0) & 0x3FFFFFF );
    Lit0 = Hash_IntObjData0( p.vHash, iDiv );
    Lit1 = Hash_IntObjData1( p.vHash, iDiv );
    assert(Lit0 >= 0 && Lit1 >= 0 && Lit0 < Lit1, "Lit0 >= 0 && Lit1 >= 0 && Lit0 < Lit1");
    Vec_WrdPush( &p.vDivs, ((word)Lit1 << 32) | (word)Lit0 );

    // if the input cover is not prime, it may happen that we are extracting divisor (x + !x)
    // although it is not strictly correct, it seems to be fine to just skip such divisors
//    if ( Abc_Lit2Var(Lit0) == Abc_Lit2Var(Lit1) && Vec_IntSize(Hsh_VecReadEntry(p->vHash, iDiv)) == 2 )
//        return;

    // collect single-cube-divisor cubes
    vLitP = Vec_WecEntry(&p.vLits, Lit0);
    vLitN = Vec_WecEntry(&p.vLits, Lit1);
    Fxch_CompressCubes( p, vLitP );
    Fxch_CompressCubes( p, vLitN );
//    Fxch_CollectSingles( vLitP, vLitN, &p->vCubesS );
//    assert( Vec_IntSize(&p->vCubesS) % 2 == 0 );
    Vec_IntTwoRemoveCommon( vLitP, vLitN, &p.vCubesS );
    Fxch_FilterCubes( p, &p.vCubesS, Lit0, Lit1 );

    // collect double-cube-divisor cube pairs
    Fxch_CollectDoubles( p, Vec_WecEntry(&p.vPairs, iDiv), &p.vCubesD, Abc_LitNot(Lit0), Abc_LitNot(Lit1) );
    assert(Vec_IntSize(&p.vCubesD) % 2 == 0, "Vec_IntSize(&p.vCubesD) % 2 == 0");
    Vec_IntUniqifyPairs( &p.vCubesD );
    assert(Vec_IntSize(&p.vCubesD) % 2 == 0, "Vec_IntSize(&p.vCubesD) % 2 == 0");

    // subtract cost of single-cube divisors
//    Vec_IntForEachEntryDouble( &p->vCubesS, hCube1, hCube2, i )
    Vec_IntForEachEntry( &p.vCubesS, hCube1, i )
        Fxch_TabSingleDivisors( p, Pla_CubeNum(hCube1), 0 );  // remove - update
    Vec_IntForEachEntryDouble( &p.vCubesD, hCube1, hCube2, i )
        Fxch_TabSingleDivisors( p, Pla_CubeNum(hCube1), 0 );
// remove - update
        Fxch_TabSingleDivisors( p, Pla_CubeNum(hCube2), 0 );  // remove - update

    // subtract cost of double-cube divisors
//    Vec_IntForEachEntryDouble( &p->vCubesS, hCube1, hCube2, i )
    Vec_IntForEachEntry( &p.vCubesS, hCube1, i )
    {
        //printf( "%d ", Pla_CubeNum(hCube1) );
        Fxch_TabDoubleDivisors( p, Pla_CubeNum(hCube1), 0 );  // remove - update
    }
    //printf( "\n" );

    Vec_IntForEachEntryDouble( &p.vCubesD, hCube1, hCube2, i )
    {
        //printf( "%d ", Pla_CubeNum(hCube1) );
        //printf( "%d  ", Pla_CubeNum(hCube2) );
        Fxch_TabDoubleDivisors( p, Pla_CubeNum(hCube1), 0 );
// remove - update
        Fxch_TabDoubleDivisors( p, Pla_CubeNum(hCube2), 0 );  // remove - update
    }
    //printf( "\n" );

    // create new literals
    p.nLits += 2;
    iVarNew = Vec_WecSize( &p.vLits ) / 2;
    vLitP = Vec_WecPushLevel( &p.vLits );
    vLitN = Vec_WecPushLevel( &p.vLits );
    vLitP = Vec_WecEntry( &p.vLits, Vec_WecSize(&p.vLits) - 2 );
    // update single-cube divisor cubes
//    Vec_IntForEachEntryDouble( &p->vCubesS, hCube1, hCube2, i )
    Vec_IntForEachEntry( &p.vCubesS, hCube1, i )
    {
//        int Lit0s, Lit1s;
        vCube1 = Fxch_ManCube( p, hCube1 );
//        Lit0s = Vec_IntEntry(vCube1, Pla_CubeLit(hCube1));
//        Lit1s = Vec_IntEntry(vCube1, Pla_CubeLit(hCube2));
//        assert( Pla_CubeNum(hCube1) == Pla_CubeNum(hCube2) );
//        assert( Vec_IntEntry(vCube1, Pla_CubeLit(hCube1)) == Lit0 );
//        assert( Vec_IntEntry(vCube1, Pla_CubeLit(hCube2)) == Lit1 );
        Fxch_CompressLiterals( vCube1, Lit0, Lit1 );
//        Vec_IntPush( vLitP, Pla_CubeHandle(Pla_CubeNum(hCube1), Vec_IntSize(vCube1)) );
        Vec_IntPush( vLitP, Pla_CubeNum(hCube1) );
        Vec_IntPush( vCube1, Abc_Var2Lit(iVarNew, 0) );

        //if ( Pla_CubeNum(hCube1) == 3 )
        //    printf( "VecSize = %d\n", Vec_IntSize(vCube1) );

        p.nLits--;
    }
    // update double-cube divisor cube pairs
    Vec_IntForEachEntryDouble( &p.vCubesD, hCube1, hCube2, i )
    {
        vCube1 = Fxch_ManCube( p, hCube1 );
        vCube2 = Fxch_ManCube( p, hCube2 );
        assert(Vec_IntEntry(vCube1, Pla_CubeLit(hCube1)) == Abc_LitNot(Lit0), "Vec_IntEntry(vCube1, Pla_CubeLit(hCube1)) == Abc_LitNot(Lit0)");
        assert(Vec_IntEntry(vCube2, Pla_CubeLit(hCube2)) == Abc_LitNot(Lit1), "Vec_IntEntry(vCube2, Pla_CubeLit(hCube2)) == Abc_LitNot(Lit1)");
        Fxch_CompressLiterals2( vCube1, Pla_CubeLit(hCube1), -1 );
//        Vec_IntPush( vLitN, Pla_CubeHandle(Pla_CubeNum(hCube1), Vec_IntSize(vCube1)) );
        Vec_IntPush( vLitN, Pla_CubeNum(hCube1) );
        Vec_IntPush( vCube1, Abc_Var2Lit(iVarNew, 1) );
        p.nLits -= Vec_IntSize(vCube2);

        //if ( Pla_CubeNum(hCube1) == 3 )
        //    printf( "VecSize = %d\n", Vec_IntSize(vCube1) );

        // remove second cube
        Vec_IntClear( vCube2 ); 
    }
    Vec_IntSort( vLitN, 0 );
    Vec_IntSort( vLitP, 0 );

    // add cost of single-cube divisors
//    Vec_IntForEachEntryDouble( &p->vCubesS, hCube1, hCube2, i )
    Vec_IntForEachEntry( &p.vCubesS, hCube1, i )
        Fxch_TabSingleDivisors( p, Pla_CubeNum(hCube1), 1 );  // add - update
    Vec_IntForEachEntryDouble( &p.vCubesD, hCube1, hCube2, i )
        Fxch_TabSingleDivisors( p, Pla_CubeNum(hCube1), 1 );  // add - update

    // add cost of double-cube divisors
//    Vec_IntForEachEntryDouble( &p->vCubesS, hCube1, hCube2, i )
    Vec_IntForEachEntry( &p.vCubesS, hCube1, i )
        Fxch_TabDoubleDivisors( p, Pla_CubeNum(hCube1), 1 );  // add - update
    Vec_IntForEachEntryDouble( &p.vCubesD, hCube1, hCube2, i )
        Fxch_TabDoubleDivisors( p, Pla_CubeNum(hCube1), 1 );  // add - update
    
    // check predicted improvement: (new SOP lits == old SOP lits - divisor weight)
    //assert( p->nLits == nLitsNew );
}

/**Function*************************************************************

  Synopsis    [Implements the improved fast_extract algorithm.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_PrintStats(p *Fxch_Man_t, clk abctime) {
    fmt.Printf("Num =%6d  ", Vec_WrdSize(&p.vDivs) - p.nVars);
    fmt.Printf("Cubes =%8d  ", Vec_WecSizeUsed(&p.vCubes));
    fmt.Printf("Lits  =%8d  ", p.nLits);
    fmt.Printf("Divs  =%8d  ", Hash_IntManEntryNum(p.vHash));
    fmt.Printf("Divs+ =%8d  ", Vec_QueSize(p.vPrio));
    fmt.Printf("PairS =%6d  ", p.nPairsS);
    fmt.Printf("PairD =%6d  ", p.nPairsD);
    Abc_PrintTime( 1, "Time", clk );
//    printf( "\n" );
}
func Fxch_PrintDivOne(p *Fxch_Man_t, iDiv int32) {
    var i int32;
    var Lit0 int32  = Hash_IntObjData0( p.vHash, iDiv );
    var Lit1 int32  = Hash_IntObjData1( p.vHash, iDiv );
    assert(Lit0 >= 0 && Lit1 >= 0 && Lit0 < Lit1, "Lit0 >= 0 && Lit1 >= 0 && Lit0 < Lit1");
    fmt.Printf("Div %4d : ", iDiv);
    fmt.Printf("Weight %12.5f  ", Vec_FltEntry(&p.vWeights, iDiv));
    fmt.Printf("Pairs = %5d  ", Vec_IntSize(Vec_WecEntry(&p.vPairs, iDiv))/2);
    for  i = 0; i < Vec_WrdSize(&p.vDivs); i++  {
        if i == Abc_Lit2Var(Lit0) {
            fmt.Printf("%d", !Abc_LitIsCompl(Lit0));
        } else if i == Abc_Lit2Var(Lit1) {
            fmt.Printf("%d", !Abc_LitIsCompl(Lit1));
        } else {
            fmt.Printf("-"); }
    }
    fmt.Printf("\n");
}
func Fxch_PrintDivisors(p *Fxch_Man_t) {
    var iDiv int32;
    for  iDiv = 1; iDiv < Vec_FltSize(&p.vWeights); iDiv++  {
        Fxch_PrintDivOne( p, iDiv );
    }
    fmt.Printf("\n");
}

func Fxch_ManFastExtract(p *Fxch_Man_t, fVerbose int32, fVeryVerbose int32) int32 {
    var nNewNodesMax int32  = ABC_INFINITY;
    var clk abctime  = Abc_Clock();
    var i, iDiv int32;
    Fxch_ManCreateDivisors( p );
//    Fxch_PrintDivisors( p );
    if fVerbose != 0 {
        Fxch_PrintStats( p, Abc_Clock() - clk ); }
    p.timeStart = Abc_Clock();
    for  i = 0; i < nNewNodesMax && Vec_QueTopPriority(p.vPrio) > 0.0; i++  {
        iDiv = Vec_QuePop(p.vPrio);
        //if ( fVerbose )
        //    Fxch_PrintStats( p, Abc_Clock() - clk );
        if fVeryVerbose != 0 {
            Fxch_PrintDivOne( p, iDiv ); }
        Fxch_ManUpdate( p, iDiv );
    }
    if fVerbose != 0 {
        Fxch_PrintStats( p, Abc_Clock() - clk ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Implements the improved fast_extract algorithm.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pla_ManPerformFxch(p *Pla_Man_t) int32 {
    var pFileName string // was [1000]byte in C

    var pFxch *Fxch_Man_t;
    Pla_ManConvertFromBits( p );
    pFxch = Fxch_ManStart( &p.vCubeLits, &p.vOccurs );
    Vec_WecZero( &p.vCubeLits );
    Vec_WecZero( &p.vOccurs );
    Fxch_ManFastExtract( pFxch, 1, 0 );
    pFileName = fmt.Sprintf("%s.blif", Pla_ManName(p));
    //Fxch_ManWriteBlif( pFileName, &pFxch->vCubes, &pFxch->vDivs );
    Fxch_ManStop( pFxch );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

