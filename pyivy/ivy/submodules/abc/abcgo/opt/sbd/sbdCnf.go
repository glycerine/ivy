package sbd

import "fmt"

/**CFile****************************************************************

  FileName    [sbdCnf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [CNF computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbdCnf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"

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
func Sbd_PrintCnf(vCnf *Vec_Str_t) {
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
func Sbd_TruthToCnf(Truth word, nVars int32, vCover *Vec_Int_t, vCnf *Vec_Str_t) int32 {
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
            if c {
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

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_TranslateCnf(vRes *Vec_Wec_t, vCnf *Vec_Str_t, vFaninMap *Vec_Int_t, iPivotVar int32) {
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

