package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaEsop.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [ESOP computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaEsop.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Eso_Man_t struct {
    pGia      *Gia_Man_t;     // user's AIG
    nVars             int32;    // number of variables
    Cube1             int32;    // ID of const1 cube
    vEsops      *Vec_Wec_t;   // ESOP for each node
    pHash   *Hsh_VecMan_t;    // hash table for cubes
    vCubes      *Vec_Wec_t;   // cover during minimization
    // internal
    vCube1      *Vec_Int_t;   // first cube
    vCube2      *Vec_Int_t;   // second cube
    vCube      *Vec_Int_t;    // resulting cube
}

func Eso_ManCube(p *Eso_Man_t, iCube int32) *Vec_Int_t { assert(iCube >= 0, "iCube >= 0"); return Hsh_VecReadEntry(p.pHash, iCube); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
    
/**Function*************************************************************

  Synopsis    [Computation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Eso_ManAlloc(pGia *Gia_Man_t) *Eso_Man_t {
    var i, n, Id int32;
    // inlining of macro ABC_CALLOC
    var p *Eso_Man_t  = &Eso_Man_t{};
    p.pGia   = pGia;
    p.nVars  = Gia_ManCiNum(pGia);
    p.Cube1  = ABC_INFINITY;
    p.vEsops = Vec_WecStart( Gia_ManObjNum(pGia) );
    p.pHash  = Hsh_VecManStart( 1000 );
    p.vCubes = Vec_WecStart(Gia_ManCiNum(pGia)+1);
    p.vCube1 = Vec_IntAlloc(Gia_ManCiNum(pGia));
    p.vCube2 = Vec_IntAlloc(Gia_ManCiNum(pGia));
    p.vCube  = Vec_IntAlloc(Gia_ManCiNum(pGia));
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((Id) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        for  n = 0; n < 2; n++  {
            Vec_IntFill( p.vCube, 1, Abc_Var2Lit(i, n) );
            Hsh_VecManAdd( p.pHash, p.vCube );
        }
        Vec_IntPush( Vec_WecEntry(p.vEsops, Id), Abc_Var2Lit(i, 0) );
    }
    return p;
}
func Eso_ManStop(p *Eso_Man_t) {
    Vec_WecFree( p.vEsops );
    Hsh_VecManStop( p.pHash );
    Vec_WecFree( p.vCubes );
    Vec_IntFree( p.vCube1 );
    Vec_IntFree( p.vCube2 );
    Vec_IntFree( p.vCube );
    // inlining of macro ABC_FREE
    p = nil;
}
    

/**Function*************************************************************

  Synopsis    [Printing/transforming the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Eso_ManCoverPrint(p *Eso_Man_t, vEsop *Vec_Int_t) {
    var vStr *Vec_Str_t;
    var vCube *Vec_Int_t;
    var i, k, Lit, Cube int32;
    if Vec_IntSize(vEsop) == 0 {
        fmt.Printf("Const 0\n");
        return;
    }
    vStr = Vec_StrAlloc( p.nVars + 4 );
    Vec_StrFill( vStr, p.nVars, '-' );
    Vec_StrPush( vStr, ' ' );
    Vec_StrPush( vStr, '1' );
    Vec_StrPush( vStr, '\n' );
    Vec_StrPush( vStr, '\0' );
    assert(Vec_IntSize(vEsop) > 0, "Vec_IntSize(vEsop) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vEsop)) && (((Cube) = Vec_IntEntry(vEsop, i)), 1); i++  {
        if Cube == p.Cube1 {
            fmt.Printf("%s", Vec_StrArray(vStr));
        } else {
            vCube = Eso_ManCube( p, Cube );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
                Vec_StrWriteEntry( vStr, Abc_Lit2Var(Lit), byte((Abc_LitIsCompl(Lit)?'0':'1')) );
            }
            fmt.Printf("%s", Vec_StrArray(vStr));
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
                Vec_StrWriteEntry( vStr, Abc_Lit2Var(Lit), '-' );
            }
        }
    }
    fmt.Printf("\n");
    Vec_StrFree( vStr );
}
func Eso_ManCoverDerive(p *Eso_Man_t, vCover *Vec_Ptr_t) *Vec_Wec_t {
    var vRes *Vec_Wec_t;
    var vEsop, vLevel *Vec_Int_t;  var i int32;
    vRes = Vec_WecAlloc( Vec_VecSizeSize((Vec_Vec_t *)vCover) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCover)) && (((vEsop) = (Vec_Int_t *)Vec_PtrEntry(vCover, i)), 1); i++  {
        if Vec_IntSize(vEsop) > 0 {
            var c, Cube int32;
            // inlining of macro Vec_IntForEachEntry
            for  c = 0; (c < Vec_IntSize(vEsop)) && (((Cube) = Vec_IntEntry(vEsop, c)), 1); c++  {
                vLevel = Vec_WecPushLevel( vRes );
                if Cube != p.Cube1 {
                    var k, Lit int32;
                    var vCube *Vec_Int_t  = Eso_ManCube( p, Cube );
                    // inlining of macro Vec_IntForEachEntry
                    for  k = 0; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
                        Vec_IntPush( vLevel, Lit );
                    }
                }
                Vec_IntPush( vLevel, -i-1 );
            }
        }
    }
    //assert( Abc_MaxInt(Vec_WecSize(vRes), 8) == Vec_WecCap(vRes) );
    return vRes;
}
func Eso_ManCoverConvert(p *Eso_Man_t, vCover *Vec_Ptr_t) *Gia_Man_t {
    var vEsop *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;  var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p.pGia) );
    pNew.pName = Abc_UtilStrsav( p.pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pGia.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p.pGia).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((pObj) = Gia_ManCi(p.pGia, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCover)) && (((vEsop) = (Vec_Int_t *)Vec_PtrEntry(vCover, i)), 1); i++  {
        if Vec_IntSize(vEsop) > 0 {
            var c, Cube, iRoot int32;
            // inlining of macro Vec_IntForEachEntry
            for  c = 0; (c < Vec_IntSize(vEsop)) && (((Cube) = Vec_IntEntry(vEsop, c)), 1); c++  {
                var k int32;
                var Lit int32;
                var iAnd int32 = 1;
                if Cube != p.Cube1 {
                    var vCube *Vec_Int_t  = Eso_ManCube( p, Cube );
                    // inlining of macro Vec_IntForEachEntry
                    for  k = 0; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
                        iAnd = Gia_ManHashAnd( pNew, iAnd, Lit + 2 );
                    }
                }
                iRoot = Gia_ManHashXor( pNew, iRoot, iAnd );
            }
            Gia_ManAppendCo( pNew, iRoot );
        
        } else {
            Gia_ManAppendCo( pNew, 0 ); }
    }
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Minimization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Eso_ManFindDistOneLitEqual( int * pCube1, int * pCube2, int nLits ) // pCube1 and pCube2 both have nLits
{
    var i int32;
    var iDiff int32 = -1;
    for  i = 0; i < nLits; i++  {
        if pCube1[i] != pCube2[i] {
            if iDiff != -1 {
                return -1; }
            if Abc_Lit2Var(pCube1[i]) != Abc_Lit2Var(pCube2[i]) {
                return -1; }
            iDiff = i;
        }
    }
    return iDiff;
}
int Eso_ManFindDistOneLitNotEqual( int * pCube1, int * pCube2, int nLits ) // pCube1 has nLits; pCube2 has nLits + 1
{
    var i int32;
    var k int32;
    var iDiff int32 = -1;
    for  i = k = 0; i < nLits; i, k = i+1, k+1  {
        if pCube1[i] != pCube2[k] {
            if iDiff != -1 {
                return -1; }
            iDiff = i;
            i--;
        }
    }
    if iDiff == -1 {
        iDiff = nLits; }
    return iDiff;
}
func Eso_ManMinimizeAdd(p *Eso_Man_t, Cube int32) {
    var fMimimize int32  = 1;
    var vCube *Vec_Int_t;
    if Cube == p.Cube1 {
        vCube = nil;
    } else {
        vCube = Eso_ManCube(p, Cube);
    } 
    var pCube2 *int32;
    var pCube *int32;
    if Cube == p.Cube1 {
        pCube = nil;
    } else {
        pCube = Vec_IntArray(vCube);
    } 
    var Cube2 int32;
    var nLits int32;
    if Cube == p.Cube1 {
        nLits = 0;
    } else {
        nLits = Vec_IntSize(vCube);
    }
    var vLevel *Vec_Int_t      = Vec_WecEntry( p.vCubes, nLits );
    var c int32;
    var k int32;
    var iLit int32;
    var iPlace int32 = Vec_IntFind( vLevel, Cube );
    if iPlace >= 0 { // identical found
    
        Vec_IntDrop( vLevel, iPlace );
        return;
    }
    if Cube == p.Cube1 { // simple case
    
        assert(Vec_IntSize(vLevel) == 0, "Vec_IntSize(vLevel) == 0");
        Vec_IntPush( vLevel, Cube );
        return;
    }
    // look for distance-1 in next bin
    if (fMimimize != 0) && nLits < p.nVars - 1 {
        var vLevel *Vec_Int_t  = Vec_WecEntry( p.vCubes, nLits+1 );
        // inlining of macro Vec_IntForEachEntry
        for  c = 0; (c < Vec_IntSize(vLevel)) && (((Cube2) = Vec_IntEntry(vLevel, c)), 1); c++  {
            pCube2 = Hsh_VecReadArray( p.pHash, Cube2 ); 
            iLit = Eso_ManFindDistOneLitNotEqual( pCube, pCube2, nLits );
            if iLit == -1 {
                continue; }
            // remove this cube
            Vec_IntDrop( vLevel, c );
            // create new cube
            Vec_IntClear( p.vCube );
            for  k = 0; k <= nLits; k++  {
                Vec_IntPush( p.vCube, Abc_LitNotCond(pCube2[k], k == iLit) );
            }
            Cube = Hsh_VecManAdd( p.pHash, p.vCube );
            // try to add new cube
            Eso_ManMinimizeAdd( p, Cube );
            return;
        }
    }
    // look for distance-1 in the same bin
    if fMimimize != 0 {
        // inlining of macro Vec_IntForEachEntry
        for  c = 0; (c < Vec_IntSize(vLevel)) && (((Cube2) = Vec_IntEntry(vLevel, c)), 1); c++  {
            pCube2 = Hsh_VecReadArray( p.pHash, Cube2 ); 
            iLit = Eso_ManFindDistOneLitEqual( pCube2, pCube, nLits );
            if iLit == -1 {
                continue; }
            // remove this cube
            Vec_IntDrop( vLevel, c );
            // create new cube
            Vec_IntClear( p.vCube );
            for  k = 0; k < nLits; k++  {
                if k != iLit {
                    Vec_IntPush( p.vCube, pCube[k] ); }
            }
            if Vec_IntSize(p.vCube) == 0 {
                Cube = p.Cube1;
            } else {
                Cube = Hsh_VecManAdd( p.pHash, p.vCube ); }
            // try to add new cube
            Eso_ManMinimizeAdd( p, Cube );
            return;
        }
    }
    assert(nLits > 0, "nLits > 0");
    if (fMimimize != 0) && nLits > 0 {
        // look for distance-1 in the previous bin
        var vLevel *Vec_Int_t  = Vec_WecEntry( p.vCubes, nLits-1 );
        // check for the case of one-literal cube
        if nLits == 1 && Vec_IntSize(vLevel) == 1 {
            Vec_IntDrop( vLevel, 0 );
            Cube = Abc_LitNot( Cube );
        
        } else {
        // inlining of macro Vec_IntForEachEntry
        for  c = 0; (c < Vec_IntSize(vLevel)) && (((Cube2) = Vec_IntEntry(vLevel, c)), 1); c++  {
            pCube2 = Hsh_VecReadArray( p.pHash, Cube2 ); 
            iLit = Eso_ManFindDistOneLitNotEqual( pCube2, pCube, nLits-1 );
            if iLit == -1 {
                continue; }
            // remove this cube
            Vec_IntDrop( vLevel, c );
            // create new cube
            Vec_IntClear( p.vCube );
            for  k = 0; k < nLits; k++  {
                Vec_IntPush( p.vCube, Abc_LitNotCond(pCube[k], k == iLit) );
            }
            Cube = Hsh_VecManAdd( p.pHash, p.vCube );
            // try to add new cube
            Eso_ManMinimizeAdd( p, Cube );
            return;
        } }
    }
    // could not find - simply add this cube
    Vec_IntPush( vLevel, Cube );
}

func Eso_ManMinimizeCopy(p *Eso_Man_t, vEsop *Vec_Int_t) {
    var vLevel *Vec_Int_t;
    var i int32;
    Vec_IntClear( vEsop );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vLevel) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        Vec_IntAppend( vEsop, vLevel );
        if i > 0 {
            var k, Cube int32;
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vLevel)) && (((Cube) = Vec_IntEntry(vLevel, k)), 1); k++  {
                assert(Vec_IntSize(Eso_ManCube(p, Cube)) == i, "Vec_IntSize(Eso_ManCube(p, Cube)) == i");
            }
        }
        Vec_IntClear( vLevel );
    }
}

/**Function*************************************************************

  Synopsis    [Compute the produce of two covers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Eso_ManComputeAnd(p *Eso_Man_t, vCube1 *Vec_Int_t, vCube2 *Vec_Int_t, vCube *Vec_Int_t) int32 {
    var pBeg *int32   = vCube.pArray;
    var pBeg1 *int32  = vCube1.pArray;
    var pBeg2 *int32  = vCube2.pArray;
    var pEnd1 *int32  = vCube1.pArray + vCube1.nSize;
    var pEnd2 *int32  = vCube2.pArray + vCube2.nSize;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            *pBeg++ = *pBeg1++, pBeg2++;
        } else if Abc_Lit2Var(*pBeg1) == Abc_Lit2Var(*pBeg2) {
            return -1;
        } else if *pBeg1 < *pBeg2 {
            *pBeg++ = *pBeg1++;
        } else { 
            *pBeg++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pBeg++ = *pBeg2++;
    }
    vCube.nSize = pBeg - vCube.pArray;
    assert(vCube.nSize <= vCube.nCap, "vCube.nSize <= vCube.nCap");
    assert(vCube.nSize >= vCube1.nSize, "vCube.nSize >= vCube1.nSize");
    assert(vCube.nSize >= vCube2.nSize, "vCube.nSize >= vCube2.nSize");
    return Hsh_VecManAdd( p.pHash, vCube );
}
func Eso_ManComputeOne(p *Eso_Man_t, vEsop1 *Vec_Int_t, vEsop2 *Vec_Int_t, vEsop *Vec_Int_t) {
    var vCube1, vCube2 Vec_Int_t;
    var i, k, Cube1, Cube2, Cube int32;
    Vec_IntClear( vEsop );
    if Vec_IntSize(vEsop1) == 0 || Vec_IntSize(vEsop2) == 0 {
        return; }
    Cube1 = Vec_IntEntry(vEsop1, 0);
    Cube2 = Vec_IntEntry(vEsop2, 0);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vEsop1)) && (((Cube1) = Vec_IntEntry(vEsop1, i)), 1); i++  {
        if Cube1 == p.Cube1 {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vEsop2)) && (((Cube2) = Vec_IntEntry(vEsop2, k)), 1); k++  {
                Eso_ManMinimizeAdd( p, Cube2 );
            }
            continue;
        }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vEsop2)) && (((Cube2) = Vec_IntEntry(vEsop2, k)), 1); k++  {
            if Cube2 == p.Cube1 {
                Eso_ManMinimizeAdd( p, Cube1 );
                continue;
            }
            vCube1 = *Hsh_VecReadEntry( p.pHash, Cube1 );
            vCube2 = *Hsh_VecReadEntry( p.pHash, Cube2 );
            Cube = Eso_ManComputeAnd( p, &vCube1, &vCube2, p.vCube );
            if Cube >= 0 {
                Eso_ManMinimizeAdd( p, Cube ); }
        }
    }
    Eso_ManMinimizeCopy( p, vEsop );
}

/**Function*************************************************************

  Synopsis    [Complements the cover if needed, or just copy it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Eso_ManTransformOne(p *Eso_Man_t, vEsop *Vec_Int_t, fCompl int32, vRes *Vec_Int_t) *Vec_Int_t {
    var i, Cube, Start int32;
    Vec_IntClear( vRes );
    if fCompl != 0 {
        if Vec_IntSize(vEsop) == 0 {
            Vec_IntPush( vRes, p.Cube1 );
        } else {
            Cube = Vec_IntEntry(vEsop, 0);
            if Cube == p.Cube1 {
                Start = 1;
            } else if Cube < 2 * p.nVars {
                Vec_IntPush( vRes, Abc_LitNot(Cube) );
Start = 1;
            } else {
                Vec_IntPush( vRes, p.Cube1 ); }
        }
    }
    // inlining of macro Vec_IntForEachEntryStart
    for  i = Start; (i < Vec_IntSize(vEsop)) && (((Cube) = Vec_IntEntry(vEsop, i)), 1); i++  {
        Vec_IntPush( vRes, Cube );
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Computes ESOP from AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Eso_ManCompute(pGia *Gia_Man_t, fVerbose int32, pvRes **Vec_Wec_t) *Gia_Man_t {
    var clk abctime  = Abc_Clock();
    var vCover *Vec_Ptr_t;
    var pNew *Gia_Man_t  = nil;
    var pObj *Gia_Obj_t; 
    var i, nCubes, nCubesUsed int32;
    var vEsop1, vEsop2, vEsop *Vec_Int_t;
    var p *Eso_Man_t  = Eso_ManAlloc( pGia ); 
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        vEsop1 = Vec_WecEntry( p.vEsops, Gia_ObjFaninId0(pObj, i) );
        vEsop2 = Vec_WecEntry( p.vEsops, Gia_ObjFaninId1(pObj, i) );
        vEsop1 = Eso_ManTransformOne( p, vEsop1, Gia_ObjFaninC0(pObj), p.vCube1 );
        vEsop2 = Eso_ManTransformOne( p, vEsop2, Gia_ObjFaninC1(pObj), p.vCube2 );
        vEsop  = Vec_WecEntry( p.vEsops, i );
        Eso_ManComputeOne( p, vEsop1, vEsop2, vEsop );
        nCubes += Vec_IntSize(vEsop);
    }
    }
    vCover = Vec_PtrAlloc( Gia_ManCoNum(pGia) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
        vEsop1 = Vec_WecEntry( p.vEsops, Gia_ObjFaninId0p(pGia, pObj) );
        vEsop1 = Eso_ManTransformOne( p, vEsop1, Gia_ObjFaninC0(pObj), p.vCube1 );
        if fVerbose != 0 {
            fmt.Printf("Output %3d:  ESOP has %5d cubes\n", i, Vec_IntSize(vEsop1)); }
//        if ( fVerbose )
//            Eso_ManCoverPrint( p, vEsop1 );
        Vec_PtrPush( vCover, Vec_IntDup(vEsop1) );
        nCubesUsed += Vec_IntSize(vEsop1);
    }
    if fVerbose != 0 {
        fmt.Printf("Outs = %d.  Cubes = %d.  Used = %d.  Hashed = %d. ", Gia_ManCoNum(pGia), nCubes, nCubesUsed, Hsh_VecSize(p.pHash));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    if pvRes {
        *pvRes = Eso_ManCoverDerive( p, vCover );
    } else {
        pNew = Eso_ManCoverConvert( p, vCover ); }
    Vec_VecFree( (Vec_Vec_t *)vCover );
    Eso_ManStop( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

