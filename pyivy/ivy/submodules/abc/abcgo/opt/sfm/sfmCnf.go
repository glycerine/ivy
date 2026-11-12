package sfm

import "fmt"

/**CFile****************************************************************

  FileName    [sfmCnf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [CNF computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmCnf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_PrintCnf(vCnf *Vec_Str_t) {
    signed char Entry;
    var i, Lit int32;
    Vec_StrForEachEntry( vCnf, Entry, i )
    {
        Lit = int32(Entry);
        if Lit == -1 {
            fmt.Printf("\n");
        } else {
            fmt.Printf("%s%d ", Abc_LitIsCompl(Lit) ? "-":"", Abc_Lit2Var(Lit)); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TruthToCnf(Truth word, pTruth *word, nVars int32, vCover *Vec_Int_t, vCnf *Vec_Str_t) int32 {
    var w int32;
    var nWords int32 = Abc_Truth6WordNum( nVars );
    Vec_StrClear( vCnf );
    if nVars <= 6 {
        if Truth == 0 || ~Truth == 0 {
            //assert( nVars == 0 );
            Vec_StrPush( vCnf, byte((Truth == 0)) );
            Vec_StrPush( vCnf, byte(-1) );
            return 1;
        }
    
    } else {
        // const0
        for  w = 0; w < nWords; w++  {
            if pTruth[w] {
                break; }
        }
        if w == nWords {
            Vec_StrPush( vCnf, byte(1) );
            Vec_StrPush( vCnf, byte(-1) );
            return 1;
        }
        // const1
        for  w = 0; w < nWords; w++  {
            if ~pTruth[w] {
                break; }
        }
        if w == nWords {
            Vec_StrPush( vCnf, byte(0) );
            Vec_StrPush( vCnf, byte(-1) );
            return 1;
        }
    }
    {
        var i, k, c, RetValue, Literal, Cube, nCubes int32;
        assert(nVars > 0, "nVars > 0");
        for  c = 0; c < 2; c ++  {
            if nVars <= 6 {
                if c {
                    Truth = ~Truth;
                } else {
                    Truth = Truth;
                }
                RetValue = Kit_TruthIsop( (unsigned *)&Truth, nVars, vCover, 0 );
            
            } else {
                if c {
                    for  k = 0; k < nWords; k++  {
                        pTruth[k] = ~pTruth[k];
                    } }
                RetValue = Kit_TruthIsop( (unsigned *)pTruth, nVars, vCover, 0 );
                if c {
                    for  k = 0; k < nWords; k++  {
                        pTruth[k] = ~pTruth[k];
                    } }
            }
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

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_CreateCnf(p *Sfm_Ntk_t) *Vec_Wec_t {
    var vCnfs *Vec_Wec_t;
    var vCnf, vCnfBase *Vec_Str_t;
    var uTruth word;
    var i, nCubes int32;
    vCnf = Vec_StrAlloc( 100 );
    vCnfs = Vec_WecStart( p.nObjs );
    Vec_WrdForEachEntryStartStop( p.vTruths, uTruth, i, p.nPis, Vec_WrdSize(p.vTruths)-p.nPos )
    {
        var Offset int32   = Vec_IntEntry( p.vStarts, i );
        var pRes *word;
        if Vec_WrdSize(p.vTruths2) {
            pRes = Vec_WrdEntryP( p.vTruths2, Offset );
        } else {
            pRes = nil;
        }
        nCubes = Sfm_TruthToCnf( uTruth, pRes, Sfm_ObjFaninNum(p, i), p.vCover, vCnf );
        vCnfBase = (Vec_Str_t *)Vec_WecEntry( vCnfs, i );
        Vec_StrGrow( vCnfBase, Vec_StrSize(vCnf) );
        memcpy( Vec_StrArray(vCnfBase), Vec_StrArray(vCnf), (size_t)Vec_StrSize(vCnf) );
        vCnfBase.nSize = Vec_StrSize(vCnf);
    }
    Vec_StrFree( vCnf );
    return vCnfs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TranslateCnf(vRes *Vec_Wec_t, vCnf *Vec_Str_t, vFaninMap *Vec_Int_t, iPivotVar int32) {
    var vClause *Vec_Int_t;
    signed char Entry;
    var i, Lit int32;
    Vec_WecClear( vRes );
    vClause = Vec_WecPushLevel( vRes );
    Vec_StrForEachEntry( vCnf, Entry, i )
    {
        if int32(Entry) == -1 {
            vClause = Vec_WecPushLevel( vRes );
            continue;
        }
        Lit = Abc_Lit2LitV( Vec_IntArray(vFaninMap), int32(Entry) );
        Lit = Abc_LitNotCond( Lit, Abc_Lit2Var(Lit) == iPivotVar );
        Vec_IntPush( vClause, Lit );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

