package mio

/**CFile****************************************************************

  FileName    [mioSop.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Derives SOP from Boolean expression.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: mioSop.c,v 1.4 2004/06/28 14:20:25 alanmi Exp $]

***********************************************************************/

//#include "mioInt.h"
//#include "exp.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Mio_CubeVar0(v int32) uint32 { return (1<< (v<<1)   );                   }
func Mio_CubeVar1(v int32) uint32 { return (1<<((v<<1)+1));                   }
func Mio_CubeHasVar0(x uint32, v int32) int32 { return (x & Mio_CubeVar0(v)) > 0;         }
func Mio_CubeHasVar1(x uint32, v int32) int32 { return (x & Mio_CubeVar1(v)) > 0;         }
func Mio_CubeEmpty(x uint32) int32 { return (x & (x>>1) & 0x55555555) != 0;    }
func Mio_CubeAnd(x uint32, y uint32) uint32 { return x | y;                             }
func Mio_CubeContains(x uint32, y uint32) int32 { return (x | y) == y;                      } // x contains y

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Push while performing SCC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopPushSCC(p *Vec_Int_t, c uint32) {
    var Entry uint32;
    var i, k int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p)) && (((Entry) = Vec_IntEntry(p, i)), 1); i++  {
        if Mio_CubeContains( Entry, c ) != 0 { // Entry contains c
        
            assert(i == k, "i == k");
            return;
        }
        if Mio_CubeContains( c, Entry ) != 0 { // c contains Entry
            continue; }
        Vec_IntWriteEntry( p, k++, Entry );
    }
    Vec_IntShrink( p, k );
    Vec_IntPush( p, c );
}

/**Function*************************************************************

  Synopsis    [Make the OR of two covers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopCoverOr(p *Vec_Int_t, q *Vec_Int_t) *Vec_Int_t {
    var r *Vec_Int_t;
    var Entry uint32;
    var i int32;
    r = Vec_IntAlloc( Vec_IntSize(p) + Vec_IntSize(q) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p)) && (((Entry) = Vec_IntEntry(p, i)), 1); i++  {
        Vec_IntPush( r, Entry );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(q)) && (((Entry) = Vec_IntEntry(q, i)), 1); i++  {
        Mio_SopPushSCC( r, Entry );
    }
    return r;
}

/**Function*************************************************************

  Synopsis    [Make the AND of two covers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopCoverAnd(p *Vec_Int_t, q *Vec_Int_t) *Vec_Int_t {
    var r *Vec_Int_t;
    var EntryP, EntryQ uint32;
    var i, k int32;
    r = Vec_IntAlloc( Vec_IntSize(p) * Vec_IntSize(q) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p)) && (((EntryP) = Vec_IntEntry(p, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(q)) && (((EntryQ) = Vec_IntEntry(q, k)), 1); k++  {
            if !(Mio_CubeEmpty( Mio_CubeAnd(EntryP, EntryQ) ) != 0) {
                Mio_SopPushSCC( r, Mio_CubeAnd(EntryP, EntryQ) ); }
        }
    }
    return r;
}

/**Function*************************************************************

  Synopsis    [Create negative literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopVar0(i int32) *Vec_Int_t {
    var vSop *Vec_Int_t;
    vSop = Vec_IntAlloc( 1 );
    Vec_IntPush( vSop, Mio_CubeVar0(i) );
    return vSop;
}

/**Function*************************************************************

  Synopsis    [Create positive literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopVar1(i int32) *Vec_Int_t {
    var vSop *Vec_Int_t;
    vSop = Vec_IntAlloc( 1 );
    Vec_IntPush( vSop, Mio_CubeVar1(i) );
    return vSop;
}

/**Function*************************************************************

  Synopsis    [Create constant 0 literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopConst0() *Vec_Int_t {
    var vSop *Vec_Int_t;
    vSop = Vec_IntAlloc( 1 );
    return vSop;
}

/**Function*************************************************************

  Synopsis    [Create constant 1 literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopConst1() *Vec_Int_t {
    var vSop *Vec_Int_t;
    vSop = Vec_IntAlloc( 1 );
    Vec_IntPush( vSop, 0 );
    return vSop;
}

/**Function*************************************************************

  Synopsis    [Derives SOP representation as the char string.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopDeriveFromArray(vSop *Vec_Int_t, nVars int32, vStr *Vec_Str_t, fPolarity int32) string {
    var Entry uint32;
    var i, k int32;
    Vec_StrClear( vStr );
    if Vec_IntSize(vSop) == 0 {
        Vec_StrPush( vStr, ' ' );
        Vec_StrPush( vStr, byte(('1'-fPolarity)) );
        Vec_StrPush( vStr, '\n' );
        Vec_StrPush( vStr, '\0' );
        return Vec_StrArray( vStr );
    }
    if Vec_IntSize(vSop) == 1 && Vec_IntEntry(vSop, 0) == 0 {
        Vec_StrPush( vStr, ' ' );
        Vec_StrPush( vStr, byte(('0'+fPolarity)) );
        Vec_StrPush( vStr, '\n' );
        Vec_StrPush( vStr, '\0' );
        return Vec_StrArray( vStr );
    }
    // create cubes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSop)) && (((Entry) = Vec_IntEntry(vSop, i)), 1); i++  {
        for  k = 0; k < nVars; k++  {
            if Mio_CubeHasVar0( Entry, k ) != 0 {
                Vec_StrPush( vStr, '0' );
            } else if Mio_CubeHasVar1( Entry, k ) != 0 {
                Vec_StrPush( vStr, '1' );
            } else { 
                Vec_StrPush( vStr, '-' ); }
        }
        Vec_StrPush( vStr, ' ' );
        Vec_StrPush( vStr, byte(('0'+fPolarity)) );
        Vec_StrPush( vStr, '\n' );
    }
    Vec_StrPush( vStr, '\0' );
    return Vec_StrArray( vStr );
}

/**Function*************************************************************

  Synopsis    [Derives SOP representation.]

  Description [The SOP is guaranteed to be SCC-free but not minimal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibDeriveSop(nVars int32, vExpr *Vec_Int_t, vStr *Vec_Str_t) string {
    var vSop *Vec_Int_t;
    var vSops0, vSops1, vTemp *Vec_Ptr_t;
    var i, Index0, Index1, fCompl0, fCompl1 int32;
    Vec_StrClear( vStr );
    if Exp_IsConst0(vExpr) {
        Vec_StrPrintStr( vStr, " 0\n" );
        Vec_StrPush( vStr, '\0' );
        return Vec_StrArray( vStr );
    }
    if Exp_IsConst1(vExpr) {
        Vec_StrPrintStr( vStr, " 1\n" );
        Vec_StrPush( vStr, '\0' );
        return Vec_StrArray( vStr );
    }
    if Exp_IsLit(vExpr) {
        for  i = 0; i < nVars; i++  {
            Vec_StrPush( vStr, '-' );
        }
        Vec_StrPrintStr( vStr, " 1\n" );
        Vec_StrPush( vStr, '\0' );
        assert((Vec_IntEntry(vExpr,0) >> 1) < nVars, "(Vec_IntEntry(vExpr,0) >> 1) < nVars");
        Vec_StrWriteEntry( vStr, Vec_IntEntry(vExpr,0) >> 1, byte(('1' - (Vec_IntEntry(vExpr,0) & 1))) );
        return Vec_StrArray( vStr );
    }
    vSops0 = Vec_PtrAlloc( nVars + Exp_NodeNum(vExpr) );
    vSops1 = Vec_PtrAlloc( nVars + Exp_NodeNum(vExpr) );
    for  i = 0; i < nVars; i++  {
        Vec_PtrPush( vSops0, Mio_SopVar0(i) );
        Vec_PtrPush( vSops1, Mio_SopVar1(i) );
    }
    for  i = 0; i < Exp_NodeNum(vExpr); i++  {
        Index0  = Vec_IntEntry( vExpr, 2*i+0 ) >> 1;
        Index1  = Vec_IntEntry( vExpr, 2*i+1 ) >> 1;
        fCompl0 = Vec_IntEntry( vExpr, 2*i+0 ) & 1;
        fCompl1 = Vec_IntEntry( vExpr, 2*i+1 ) & 1;
        // positive polarity
        vSop = Mio_SopCoverAnd( fCompl0 ? (Vec_Int_t *)Vec_PtrEntry(vSops0, Index0) : (Vec_Int_t *)Vec_PtrEntry(vSops1, Index0),
                                fCompl1 ? (Vec_Int_t *)Vec_PtrEntry(vSops0, Index1) : (Vec_Int_t *)Vec_PtrEntry(vSops1, Index1) );
        Vec_PtrPush( vSops1, vSop );
        // negative polarity
        vSop = Mio_SopCoverOr( fCompl0 ? (Vec_Int_t *)Vec_PtrEntry(vSops1, Index0) : (Vec_Int_t *)Vec_PtrEntry(vSops0, Index0),
                               fCompl1 ? (Vec_Int_t *)Vec_PtrEntry(vSops1, Index1) : (Vec_Int_t *)Vec_PtrEntry(vSops0, Index1) );
        Vec_PtrPush( vSops0, vSop );
    }
    // complement
    if Vec_IntEntryLast(vExpr) & 1 {
        vTemp  = vSops0;
        vSops0 = vSops1;
        vSops1 = vTemp;
    }
    // select the best polarity
    if Vec_IntSize( (Vec_Int_t *)Vec_PtrEntryLast(vSops0) ) < Vec_IntSize( (Vec_Int_t *)Vec_PtrEntryLast(vSops1) ) {
        vSop = (Vec_Int_t *)Vec_PtrEntryLast(vSops0);
    } else {
        vSop = (Vec_Int_t *)Vec_PtrEntryLast(vSops1); }
    // convert positive polarity into SOP
    Mio_SopDeriveFromArray( vSop, nVars, vStr, (vSop == Vec_PtrEntryLast(vSops1)) );
    Vec_VecFree( (Vec_Vec_t *)vSops0 );
    Vec_VecFree( (Vec_Vec_t *)vSops1 );
    return Vec_StrArray( vStr );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

