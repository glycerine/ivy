package gia

import "time"

/**CFile****************************************************************

  FileName    [giaSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Fast sequential simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
// extern void Gia_ManResetRandom( Gia_ParSim_t * pPars );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

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

  Synopsis    [Compares simulation info of two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Compares simulation info of one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Sim2ClassCreate(p *Gia_Man_t, vClass *Vec_Int_t) {
    var Repr int32 = GIA_VOID;
    var EntPrev int32 = -1;
    var Ent int32;
    var i int32;
    assert(Vec_IntSize(vClass) > 0, "Vec_IntSize(vClass) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClass)) && (((Ent) = Vec_IntEntry(vClass, i)), 1); i++  {
        if i == 0 {
            Repr = Ent;
            Gia_ObjSetRepr( p, Ent, GIA_VOID );
            EntPrev = Ent;
        
        } else {
            assert(Repr < Ent, "Repr < Ent");
            Gia_ObjSetRepr( p, Ent, Repr );
            Gia_ObjSetNext( p, EntPrev, Ent );
            EntPrev = Ent;
        }
    }
    Gia_ObjSetNext( p, EntPrev, 0 );
}

/**Function*************************************************************

  Synopsis    [Refines one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Sim2ClassRefineOne(p *Gia_Sim2_t, i int32) int32 {
    var pObj0, pObj1 *Gia_Obj_t;
    var pSim0, pSim1 *uint32;
    var Ent int32;
    Vec_IntClear( p.vClassOld );
    Vec_IntClear( p.vClassNew );
    Vec_IntPush( p.vClassOld, i );
    pObj0 = Gia_ManObj( p.pAig, i );
    pSim0 = Gia_Sim2Data( p, i );
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p.pAig, i), "Gia_ObjIsHead(p.pAig, i)"), Ent = Gia_ObjNext(p.pAig, i); Ent > 0; Ent = Gia_ObjNext(p.pAig, Ent)  {
        pObj1 = Gia_ManObj( p.pAig, Ent );
        pSim1 = Gia_Sim2Data( p, Ent );
        if Gia_Sim2CompareEqual( pSim0, pSim1, p.nWords, Gia_ObjPhase(pObj0) ^ Gia_ObjPhase(pObj1) ) != 0 {
            Vec_IntPush( p.vClassOld, Ent );
        } else {
            Vec_IntPush( p.vClassNew, Ent ); }
    }
    if Vec_IntSize( p.vClassNew ) == 0 {
        return 0; }
    Gia_Sim2ClassCreate( p.pAig, p.vClassOld );
    Gia_Sim2ClassCreate( p.pAig, p.vClassNew );
    if Vec_IntSize(p.vClassNew) > 1 {
        return 1 + Gia_Sim2ClassRefineOne( p, Vec_IntEntry(p.vClassNew,0) ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes hash key of the simuation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_Sim2HashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Gia_Sim2HashKey(pSim *uint32, nWords int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    if pSim[0] & 1 {
        for  i = 0; i < nWords; i++  {
            uHash ^= ~pSim[i] * spStatic_Gia_Sim2HashKey_s_Primes[i & 0xf];
        }
    } else {
        for  i = 0; i < nWords; i++  {
            uHash ^= pSim[i] * spStatic_Gia_Sim2HashKey_s_Primes[i & 0xf];
        } }
    return int32((uHash % nTableSize));

}

/**Function*************************************************************

  Synopsis    [Refines nodes belonging to candidate constant class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Sim2ProcessRefined(p *Gia_Sim2_t, vRefined *Vec_Int_t) {
    var pSim *uint32;
    var pTable *int32;
    var nTableSize int32;
    var i int32;
    var k int32;
    var Key int32;
    if Vec_IntSize(vRefined) == 0 {
        return; }
    nTableSize = Abc_PrimeCudd( 1000 + Vec_IntSize(vRefined) / 3 );
    // inlining of macro ABC_CALLOC
    pTable = ((int *) calloc((size_t)(nTableSize), sizeof(int)));
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        pSim = Gia_Sim2Data( p, i );
        Key = Gia_Sim2HashKey( pSim, p.nWords, nTableSize );
        if pTable[Key] == 0 {
            assert(Gia_ObjRepr(p.pAig, i) == 0, "Gia_ObjRepr(p.pAig, i) == 0");
            assert(Gia_ObjNext(p.pAig, i) == 0, "Gia_ObjNext(p.pAig, i) == 0");
            Gia_ObjSetRepr( p.pAig, i, GIA_VOID );
        
        } else {
            Gia_ObjSetNext( p.pAig, pTable[Key], i );
            Gia_ObjSetRepr( p.pAig, i, Gia_ObjRepr(p.pAig, pTable[Key]) );
            if Gia_ObjRepr(p.pAig, i) == GIA_VOID {
                Gia_ObjSetRepr( p.pAig, i, pTable[Key] ); }
            assert(Gia_ObjRepr(p.pAig, i) > 0, "Gia_ObjRepr(p.pAig, i) > 0");
        }
        pTable[Key] = i;
    }
/*
    Vec_IntForEachEntry( vRefined, i, k )
    {
        if ( Gia_ObjIsHead( p->pAig, i ) )
            Gia_Sim2ClassRefineOne( p, i );
    }
*/
    // inlining of macro ABC_FREE
    pTable = nil;
}
/**Function*************************************************************

  Synopsis    [Refines equivalences after one simulation round.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Sim2InfoRefineEquivs(p *Gia_Sim2_t) {
    var vRefined *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var pSim *uint32;
    var i, Count int32;
    // process constant candidates
    vRefined = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
        if !Gia_ObjIsConst(p.pAig, i) {
            continue; }
        pSim = Gia_Sim2Data( p, i );
//Extra_PrintBinary( stdout, pSim, 32 * p->nWords );  printf( "\n" );
        if !(Gia_Sim2CompareZero( pSim, p.nWords, Gia_ObjPhase(pObj) ) != 0) {
            Vec_IntPush( vRefined, i );
            Count++;
        }
    }
    Gia_Sim2ProcessRefined( p, vRefined );
    Vec_IntFree( vRefined );
    // process other classes
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p.pAig); i++  { if !Gia_ObjIsHead(p.pAig, i) {} else {
        Count += Gia_Sim2ClassRefineOne( p, i ); }
    }
//    if ( Count )
//        printf( "Refined %d times.\n", Count );
}

/**Function*************************************************************

  Synopsis    [Returns index of the first pattern that failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns index of the PO and pattern that failed it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Sim2GenerateCounter(pAig *Gia_Man_t, iFrame int32, iOut int32, nWords int32, iPat int32) *Abc_Cex_t {
    var p *Abc_Cex_t;
    var pData *uint32;
    var f, i, w, Counter int32;
    p = Abc_CexAlloc( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), iFrame+1 );
    p.iFrame = iFrame;
    p.iPo = iOut;
    // fill in the binary data
    Counter = p.nRegs;
    // inlining of macro ABC_ALLOC
    pData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nWords)));
    for  f = 0; f <= iFrame; f++, Counter += p.nPis  {
    for  i = 0; i < Gia_ManPiNum(pAig); i++  {
        for  w = nWords-1; w >= 0; w--  {
            pData[w] = Gia_ManRandom( 0 );
        }
        if Abc_InfoHasBit( pData, iPat ) {
            Abc_InfoSetBit( p.pData, Counter + i ); }
    }
    }
    // inlining of macro ABC_FREE
    pData = nil;
    return p;
}
 
/**Function*************************************************************

  Synopsis    [Performs simulation to refine equivalence classes.]

  Description [Returns 1 if counter-example is detected.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimSimulateEquiv(pAig *Gia_Man_t, pPars *Gia_ParSim_t) int32 {
    var p *Gia_Sim2_t;
    var pObj *Gia_Obj_t;
    var clkTotal abctime  = Abc_Clock();
    var i, RetValue, iOut, iPat int32;
    var nTimeToStop abctime;
    if pPars.TimeLimit {
        nTimeToStop = pPars.TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    assert(pAig.pReprs && pAig.pNexts, "pAig.pReprs && pAig.pNexts");
    // inlining of macro ABC_FREE
    pAig.pCexSeq = nil;
    p = Gia_Sim2Create( pAig, pPars );
    Gia_ManResetRandom( pPars );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCi(p.pAig, Gia_ManPiNum(p.pAig)+i)); i++  {
        Gia_Sim2InfoZero( p, Gia_Sim2Data(p, Gia_ObjValue(pObj)) );
    }
    for  i = 0; i < pPars.nIters; i++  {
        Gia_Sim2SimulateRound( p );
        if pPars.fVerbose {
            Abc_Print( 1, "Frame %4d out of %4d and timeout %3d sec. ", i+1, pPars.nIters, pPars.TimeLimit );
            if pAig.pReprs && pAig.pNexts {
                Abc_Print( 1, "Lits = %4d. ", Gia_ManEquivCountLitsAll(pAig) ); }
            Abc_Print( 1, "Time = %7.2f sec\r", (1.0*Abc_Clock()-clkTotal)/int64(time.Second) );
        }
        if pPars.fCheckMiter && (Gia_Sim2CheckPos( p, &iOut, &iPat ) != 0) {
            Gia_ManResetRandom( pPars );
            pPars.iOutFail = iOut;
            pAig.pCexSeq = Gia_Sim2GenerateCounter( pAig, i, iOut, p.nWords, iPat );
            Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", iOut, pAig.pName, i );
            if !Gia_ManVerifyCex( pAig, pAig.pCexSeq, 0 ) {
//                Abc_Print( 1, "\n" );
                Abc_Print( 1, "\nGenerated counter-example is INVALID.                    " );
//                Abc_Print( 1, "\n" );
            
            } else {
//                Abc_Print( 1, "\n" );
//                if ( pPars->fVerbose )
//                Abc_Print( 1, "\nGenerated counter-example is verified correctly.         " );
//                Abc_Print( 1, "\n" );
            }
            RetValue = 1;
            break;
        }
        if pAig.pReprs && pAig.pNexts {
            Gia_Sim2InfoRefineEquivs( p ); }
        if Abc_Clock() > nTimeToStop {
            i++;
            break;
        }
        if i < pPars.nIters - 1 {
            Gia_Sim2InfoTransfer( p ); }
    }
    Gia_Sim2Delete( p );
    if pAig.pCexSeq == nil {
        Abc_Print( 1, "No bug detected after simulating %d frames with %d words.  ", i, pPars.nWords ); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

