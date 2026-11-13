package acec

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [acecRe.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecRe.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecHash.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Ree_ForEachCut( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 2 )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Detecting FADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ree_TruthPrecompute() {
    var Truths [8]word = [8]word{ 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77 };
    var Truth word;
    var i int32;
    for  i = 0; i < 8; i++  {
        Truth = Truths[i];
        Truth = Abc_Tt6SwapAdjacent( Truth, 1 );
        Abc_TtPrintHexRev( os.Stdout, &Truth, 3 );
        fmt.Printf("\n");
    }
    fmt.Printf("\n");
    for  i = 0; i < 8; i++  {
        Truth = Truths[i];
        Truth = Abc_Tt6SwapAdjacent( Truth, 1 );
        Truth = Abc_Tt6SwapAdjacent( Truth, 0 );
        Abc_TtPrintHexRev( os.Stdout, &Truth, 3 );
        fmt.Printf("\n");
    }
    fmt.Printf("\n");
}
func Ree_TruthPrecompute2() {
    var i, b int32;
    for  i = 0; i < 8; i++  {
        var Truth word  = 0xE8;
        for  b = 0; b < 3; b++  {
            if (i >> b) & 1 {
                Truth = Abc_Tt6Flip( Truth, b ); }
        }
        fmt.Printf("%d = %X\n", i, 0xFF & int32(Truth));
    }
}

/**Function*************************************************************

  Synopsis    [Detecting FADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ree_ManComputeCuts(p *Gia_Man_t, pvXors **Vec_Int_t, fVerbose int32) *Vec_Int_t {
    extern void Ree_ManRemoveTrivial( Gia_Man_t * p, Vec_Int_t * vAdds );
    extern void Ree_ManRemoveContained( Gia_Man_t * p, Vec_Int_t * vAdds );
    var pObj *Gia_Obj_t; 
    var pList0 *int32;
    var pList1 *int32;
    var i int32;
    var nCuts int32;
    var pHash *Hash_IntMan_t  = Hash_IntManStart( 1000 );
    var vAdds *Vec_Int_t;
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vData *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vCuts *Vec_Int_t  = Vec_IntAlloc( 30 * Gia_ManAndNum(p) );
    Vec_IntFill( vCuts, Gia_ManObjNum(p), 0 );
    Gia_ManCleanValue( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_IntWriteEntry( vCuts, Gia_ObjId(p, pObj), Vec_IntSize(vCuts) );
        Vec_IntPush( vCuts, 1 );
        Vec_IntPush( vCuts, 1 );
        Vec_IntPush( vCuts, Gia_ObjId(p, pObj) );
        Vec_IntPush( vCuts, 0xAA );
    }
    if pvXors { *pvXors = Vec_IntAlloc( 1000 ); }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pList0 = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId0(pObj, i)) );
        pList1 = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId1(pObj, i)) );
        Ree_ManCutMerge( p, i, pList0, pList1, vTemp, pHash, vData, pvXors ? *pvXors : nil );
        Vec_IntWriteEntry( vCuts, i, Vec_IntSize(vCuts) );
        Vec_IntAppend( vCuts, vTemp );
        nCuts += Vec_IntEntry( vTemp, 0 );
    }
    }
    if fVerbose != 0 {
        fmt.Printf("AIG nodes = %d.  Cuts = %d.  Cuts/Node = %.2f.  Ints/Node = %.2f.\n", Gia_ManAndNum(p), nCuts, 1.0*nCuts/Gia_ManAndNum(p), 1.0*Vec_IntSize(vCuts)/Gia_ManAndNum(p)); }
    Vec_IntFree( vTemp );
    Vec_IntFree( vCuts );
    vAdds = Ree_ManDeriveAdds( pHash, vData, fVerbose );
    qsort( Vec_IntArray(vAdds), (size_t)(Vec_IntSize(vAdds)/6), 24, (int (*)(void *, void *))Ree_ManCompare );
    if fVerbose != 0 {
        fmt.Printf("Adders = %d.  Total cuts = %d.  Hashed cuts = %d.  Hashed/Adders = %.2f.\n", Vec_IntSize(vAdds)/6, Vec_IntSize(vData)/3, Hash_IntManEntryNum(pHash), 6.0*Hash_IntManEntryNum(pHash)/Vec_IntSize(vAdds)); }
    Vec_IntFree( vData );
    Hash_IntManStop( pHash );
    Ree_ManRemoveTrivial( p, vAdds );
    Ree_ManRemoveContained( p, vAdds );
    //Ree_ManPrintAdders( vAdds, 1 );
    return vAdds;
}

/**Function*************************************************************

  Synopsis    [Highlight nodes inside FAs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
