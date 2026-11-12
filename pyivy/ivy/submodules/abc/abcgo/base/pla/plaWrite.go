package pla

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [plaWrite.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaWrite.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pla.h"

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
func Pla_WritePlaInt(p *Pla_Man_t) *Vec_Str_t {
    var vOut *Vec_Str_t  = Vec_StrAlloc( 10000 );
    var pLits string  = "-01?";
    var pCubeIn, pCubeOut *word; 
    var i, k, Lit int32; 
    // write comments
    Vec_StrPrintStr( vOut, "# SOP \"" );
    Vec_StrPrintStr( vOut, Pla_ManName(p) );
    Vec_StrPrintStr( vOut, "\" written via PLA package in ABC on " );
    Vec_StrPrintStr( vOut, Extra_TimeStamp() );
    Vec_StrPrintStr( vOut, "\n\n" );
    // write header
    if p.Type != PLA_FILE_FD {
        if p.Type == PLA_FILE_F {
            Vec_StrPrintStr( vOut, ".type f\n" );
        } else if p.Type == PLA_FILE_FR {
            Vec_StrPrintStr( vOut, ".type fr\n" );
        } else if p.Type == PLA_FILE_FDR {
            Vec_StrPrintStr( vOut, ".type fdr\n" );
        } else if p.Type == PLA_FILE_NONE {
            Vec_StrPrintStr( vOut, ".type ???\n" ); }
    }
    Vec_StrPrintStr( vOut, ".i " );
    Vec_StrPrintNum( vOut, p.nIns );
    Vec_StrPrintStr( vOut, "\n.o " );
    Vec_StrPrintNum( vOut, p.nOuts );
    Vec_StrPrintStr( vOut, "\n.p " );
    Vec_StrPrintNum( vOut, Pla_ManCubeNum(p) );
    Vec_StrPrintStr( vOut, "\n" );
    // write cube
    // inlining of macro Pla_ForEachCubeInOut
    for  i = 0; (i < Pla_ManCubeNum(p))  && (((pCubeIn) = Pla_CubeIn(p, i)), 1)  && (((pCubeOut) = Pla_CubeOut(p, i)), 1); i++  {
        // inlining of macro Pla_CubeForEachLit
        for  k = 0; (k < p.nIns)  && (((Lit) = Pla_CubeGetLit(pCubeIn, k)), 1); k++  {
            Vec_StrPush( vOut, pLits[Lit] );
        }
        Vec_StrPush( vOut, ' ' );
        // inlining of macro Pla_CubeForEachLit
        for  k = 0; (k < p.nOuts)  && (((Lit) = Pla_CubeGetLit(pCubeOut, k)), 1); k++  {
            Vec_StrPush( vOut, pLits[Lit] );
        }
        Vec_StrPush( vOut, '\n' );
    }
    Vec_StrPrintStr( vOut, ".e\n\n\0" );
    return vOut;
}   
func Pla_WritePla(p *Pla_Man_t, pFileName string) {
    var vOut *Vec_Str_t  = Pla_WritePlaInt( p );
    if Vec_StrSize(vOut) > 0 {
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            fwrite( Vec_StrArray(vOut), 1, Vec_StrSize(vOut), pFile );
            _ = (pFile).Close();
        }
    }
    Vec_StrFreeP( &vOut );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

