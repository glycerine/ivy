package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaMf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]`
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecMem.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "sat/cnf/cnf.h"
//#include "opt/dau/dau.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Mf_SetForEachCut( pList, pCut, i )      for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Mf_CutSize(pCut) + 1 )

// Original macro:
// #define Mf_ObjForEachCut( pCuts, i, nCuts )     for ( i = 0, i < nCuts; i++ )


// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computing truth tables of useful DSD classes of 6-functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var s_nCalls int32;
var s_vTtMem *Vec_Mem_t = nil;
func Mf_ManTruthCanonicize(t *word, nVars int32) int32 {
    var Temp word;
    var Best word = *t;
    var r, i, Config int32;
    for  r = 0; r < 1; r++  {
        if Best > (Temp = ~Best) {
            Best = Temp;
Config ^= (1 << nVars); }
        for  i = 0; i < nVars; i++  {
            if Best > (Temp = Abc_Tt6Flip(Best, i)) {
                Best = Temp;
Config ^= (1 << i); }
        }
    }
    *t = Best;
    if s_vTtMem == nil {
        s_vTtMem = Vec_MemAllocForTT( 6, 0 ); }
    Vec_MemHashInsert( s_vTtMem, t );
    s_nCalls++;
    return Config;
}
func Mf_ManTruthQuit() {
    if s_vTtMem == nil {
        return; }
    fmt.Printf("TT = %d (%.2f %%)\n", Vec_MemEntryNum(s_vTtMem), 100.0 * Vec_MemEntryNum(s_vTtMem) / s_nCalls);
    Vec_MemHashFree( s_vTtMem );
    Vec_MemFree( s_vTtMem );
    s_vTtMem = nil;
    s_nCalls = 0;
}

func Mf_ManTruthCollect(Limit int32) *Vec_Wrd_t {
    extern Vec_Wrd_t * Mpm_ManGetTruthWithCnf( int Limit );
    var pPerm *int32  = Extra_PermSchedule( 6 );
    var pComp *int32  = Extra_GreyCodeSchedule( 6 );
    var vTruths *Vec_Wrd_t  = Mpm_ManGetTruthWithCnf( Limit );
    var vResult *Vec_Wrd_t  = Vec_WrdAlloc( 1 << 20 );
    var uTruth, tCur, tTemp1, tTemp2 word;
    var i, p, c, k int32;
    Vec_WrdForEachEntry( vTruths, uTruth, k )
    {
        for  i = 0; i < 2; i++  {
            if i != 0 {
                tCur = ~uTruth;
            } else {
                tCur = uTruth;
            }
            tTemp1 = tCur;
            for  p = 0; p < 720; p++  {
                tTemp2 = tCur;
                for  c = 0; c < 64; c++  {
                    tCur = Abc_Tt6Flip( tCur, pComp[c] );
                    Vec_WrdPush( vResult, tCur );
                }
                assert(tTemp2 == tCur, "tTemp2 == tCur");
                tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[p] );
            }
            assert(tTemp1 == tCur, "tTemp1 == tCur");
        }
    }
    pPerm = nil;
    pComp = nil;
    fmt.Printf("Original = %d.  ", Vec_WrdSize(vTruths));
    Vec_WrdFree( vTruths );
    fmt.Printf("Total = %d.  ", Vec_WrdSize(vResult));
    vTruths = Vec_WrdUniqifyHash( vResult, 1 );
    Vec_WrdFree( vResult );
    fmt.Printf("Unique = %d.  ", Vec_WrdSize(vTruths));
    Vec_WrdForEachEntry( vTruths, uTruth, k )
    {
        Mf_ManTruthCanonicize( &uTruth, 6 );
        Vec_WrdWriteEntry( vTruths, k, uTruth );
    }
    vResult = Vec_WrdUniqifyHash( vTruths, 1 );
    Vec_WrdFree( vTruths );
    fmt.Printf("Unique = %d.  \n", Vec_WrdSize(vResult));
    return vResult;
}
func Mf_ManTruthCount() int32 {
    var vTruths *Vec_Wrd_t  = Mf_ManTruthCollect( 10 );
    var RetValue int32  = Vec_WrdSize( vTruths );
    Vec_WrdFree( vTruths );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Collect truth tables used by the mapper.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManProfileTruths(p *Mf_Man_t) {
    var vCounts *Vec_Int_t;
    var i int32;
    var Entry int32;
    var pCut *int32;
    var Counter int32;
    vCounts = Vec_IntStart( Vec_IntSize(&p.vCnfSizes) );
    Gia_ManForEachAndId( p.pGia, i )
    {
        if !(Mf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, i );
        Vec_IntAddToEntry( vCounts, Abc_Lit2Var(Mf_CutFunc(pCut)), 1 );
    }
    Vec_IntForEachEntry( vCounts, Entry, i )
    {
        if Entry == 0 {
            continue; }
        fmt.Printf("%6d : ", Counter++);
        fmt.Printf("%6d : ", i);
        fmt.Printf("Occur = %4d  ", Entry); 
        fmt.Printf("CNF size = %2d  ", Vec_IntEntry(&p.vCnfSizes, i));
        Dau_DsdPrintFromTruth( Vec_MemReadEntry(p.vTtMem, i), p.pPars.nLutSize );
    }
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    [Derives CNFs for each function used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Derives CNF for the AIG using the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/








/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/











 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    [Flow and area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
















/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [CNF generation]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

