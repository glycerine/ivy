package if_

import "fmt"

/**CFile****************************************************************

  FileName    [ifCache.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifCache.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "misc/vec/vecHsh.h"

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
func If_ManCacheRecord(p *If_Man_t, iDsd0 int32, iDsd1 int32, nShared int32, iDsd int32) {
    assert(nShared >= 0 && nShared <= p.pPars.nLutSize, "nShared >= 0 && nShared <= p.pPars.nLutSize");
    if p.vCutData == nil {
        p.vCutData = Vec_IntAlloc( 10000 ); }
    if iDsd0 > iDsd1 {
        // inlining of macro ABC_SWAP
         var t int32  = iDsd0; iDsd0 = iDsd1; iDsd1 = t; };
    Vec_IntPush( p.vCutData, iDsd0 );
    Vec_IntPush( p.vCutData, iDsd1 );
    Vec_IntPush( p.vCutData, nShared );
    Vec_IntPush( p.vCutData, iDsd );
//    printf( "%6d %6d %6d %6d\n", iDsd0, iDsd1, nShared, iDsd );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCacheAnalize(p *If_Man_t) {
    var vRes *Vec_Int_t;
    var vTest [32]*Vec_Int_t;
    var i, Entry, uUnique int32;
    vRes = Hsh_IntManHashArray( p.vCutData, 4 );
    for  i = 0; i <= p.pPars.nLutSize; i++  {
        vTest[i] = Vec_IntAlloc( 1000 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRes)) && (((Entry) = Vec_IntEntry(vRes, i)), 1); i++  {
        Vec_IntPush( vTest[Vec_IntEntry(p.vCutData, 4*i+2)], Entry );
    }
    for  i = 0; i <= p.pPars.nLutSize; i++  {
        uUnique = Vec_IntCountUnique(vTest[i]);
        fmt.Printf("%2d-var entries = %8d. (%6.2f %%)  Unique entries = %8d. (%6.2f %%)\n", i, Vec_IntSize(vTest[i]), 100.0*Vec_IntSize(vTest[i])/Abc_MaxInt(1, Vec_IntSize(vRes)), uUnique, 100.0*uUnique/Abc_MaxInt(1, Vec_IntSize(vTest[i])));
    }
    for  i = 0; i <= p.pPars.nLutSize; i++  {
        Vec_IntFree( vTest[i] );
    }
    uUnique = Vec_IntCountUnique(vRes);
    fmt.Printf("Total  entries = %8d. (%6.2f %%)  Unique entries = %8d. (%6.2f %%)\n", Vec_IntSize(p.vCutData)/4, 100.0, uUnique, 100.0*uUnique/Abc_MaxInt(1, Vec_IntSize(p.vCutData)/4));
    Vec_IntFree( vRes );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

