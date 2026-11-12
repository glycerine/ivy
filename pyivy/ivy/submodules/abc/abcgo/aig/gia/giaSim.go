package gia

import (
	"fmt"
	"time"
)

import "os"

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
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derives signal implications.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Finds signals, which reset flops to have constant values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

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

  Synopsis    [Returns index of the first pattern that failed.]

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
func Gia_ManGenerateCounter(pAig *Gia_Man_t, iFrame int32, iOut int32, nWords int32, iPat int32, vCis2Ids *Vec_Int_t) *Abc_Cex_t {
    var p *Abc_Cex_t;
    var pData *uint32;
    var f, i, w, iPioId, Counter int32;
    p = Abc_CexAlloc( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), iFrame+1 );
    p.iFrame = iFrame;
    p.iPo    = iOut;
    // fill in the binary data
    Counter = p.nRegs;
    // inlining of macro ABC_ALLOC
    pData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nWords)));
    for  f = 0; f <= iFrame; f++, Counter += p.nPis  {
    for  i = 0; i < Gia_ManPiNum(pAig); i++  {
        iPioId = Vec_IntEntry( vCis2Ids, i );
        if iPioId >= p.nPis {
            continue; }
        for  w = nWords-1; w >= 0; w--  {
            pData[w] = Gia_ManRandom( 0 );
        }
        if Abc_InfoHasBit( pData, iPat ) {
            Abc_InfoSetBit( p.pData, Counter + iPioId ); }
    }
    }
    // inlining of macro ABC_FREE
    pData = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManResetRandom(pPars *Gia_ParSim_t) {
    var i int32;
    Gia_ManRandom( 1 );
    for  i = 0; i < pPars.RandSeed; i++  {
        Gia_ManRandom( 0 );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimSimulate(pAig *Gia_Man_t, pPars *Gia_ParSim_t) int32 {
    extern int Gia_ManSimSimulateEquiv( Gia_Man_t * pAig, Gia_ParSim_t * pPars );
    var p *Gia_ManSim_t;
    var clkTotal abctime  = Abc_Clock();
    var i, iOut, iPat, RetValue int32;
    var nTimeToStop abctime;
    if pPars.TimeLimit {
        nTimeToStop = pPars.TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    if pAig.pReprs && pAig.pNexts {
        return Gia_ManSimSimulateEquiv( pAig, pPars ); }
    // inlining of macro ABC_FREE
    pAig.pCexSeq = nil;
    p = Gia_ManSimCreate( pAig, pPars );
    Gia_ManResetRandom( pPars );
    Gia_ManSimInfoInit( p );
    for  i = 0; i < pPars.nIters; i++  {
        Gia_ManSimulateRound( p );
        if pPars.fVerbose {
            Abc_Print( 1, "Frame %4d out of %4d and timeout %3d sec. ", i+1, pPars.nIters, pPars.TimeLimit );
            Abc_Print( 1, "Time = %7.2f sec\r", (1.0*Abc_Clock()-clkTotal)/int64(time.Second) );
        }
        if pPars.fCheckMiter && (Gia_ManCheckPos( p, &iOut, &iPat ) != 0) {
            Gia_ManResetRandom( pPars );
            pPars.iOutFail = iOut;
            pAig.pCexSeq = Gia_ManGenerateCounter( pAig, i, iOut, p.nWords, iPat, p.vCis2Ids );
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
        if Abc_Clock() > nTimeToStop {
            i++;
            break;
        }
        if i < pPars.nIters - 1 {
            Gia_ManSimInfoTransfer( p ); }
    }
    Gia_ManSimDelete( p );
    if pAig.pCexSeq == nil {
        Abc_Print( 1, "No bug detected after simulating %d frames with %d words.  ", i, pPars.nWords ); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimReadFile(pFileIn string) *Vec_Int_t {
    var c int32;
    var vPat *Vec_Int_t;
    var pFile *os.File  = os.Open(pFileIn);
    if pFile == nil {
        fmt.Printf("Cannot open input file.\n");
        return nil;
    }
    vPat = Vec_IntAlloc( 1000 );
    for  (c = fgetc(pFile)) != -1  {
        if c == '0' || c == '1' {
            Vec_IntPush( vPat, c - '0' ); }
    }
    _ = (pFile).Close();
    return vPat;
}
func Gia_ManSimWriteFile(pFileOut string, vPat *Vec_Int_t, nOuts int32) int32 {
    var c, i int32;
    var pFile *os.File  = os.OpenFile(pFileOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file.\n");
        return 0;
    }
    assert(Vec_IntSize(vPat) % nOuts == 0, "Vec_IntSize(vPat) % nOuts == 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPat)) && (((c) = Vec_IntEntry(vPat, i)), 1); i++  {
        fputc( '0' + c, pFile );
        if i % nOuts == nOuts - 1 {
            fputc( '\n', pFile ); }
    }
    _ = (pFile).Close();
    return 1;
}
func Gia_ManSimSimulateOne(p *Gia_Man_t, vPat *Vec_Int_t) *Vec_Int_t {
    var vPatOut *Vec_Int_t;
    var pObj, pObjRo *Gia_Obj_t;
    var i, k, f int32;
    assert(Vec_IntSize(vPat) % Gia_ManPiNum(p) == 0, "Vec_IntSize(vPat) % Gia_ManPiNum(p) == 0");
    Gia_ManConst0(p).fMark1 = 0;
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.fMark1 = 0;
    }
    vPatOut = Vec_IntAlloc( 1000 );
    for  k = f = 0; f < Vec_IntSize(vPat) / Gia_ManPiNum(p); f++  {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.fMark1 = Vec_IntEntry( vPat, k++ );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark1 = (Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            pObj.fMark1 = (Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj));
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Vec_IntPush( vPatOut, pObj.fMark1 );
        }
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            pObjRo.fMark1 = pObj.fMark1;
        }
    }
    assert(k == Vec_IntSize(vPat), "k == Vec_IntSize(vPat)");
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark1 = 0;
    }
    return vPatOut;
}
func Gia_ManSimSimulatePattern(p *Gia_Man_t, pFileIn string, pFileOut string) {
    var vPat, vPatOut *Vec_Int_t;
    vPat = Gia_ManSimReadFile( pFileIn );
    if vPat == nil {
        return; }
    if Vec_IntSize(vPat) % Gia_ManPiNum(p) {
        fmt.Printf("The number of 0s and 1s in the input file (%d) does not evenly divide by the number of primary inputs (%d).\n", Vec_IntSize(vPat), Gia_ManPiNum(p));
        Vec_IntFree( vPat );
        return;
    }
    vPatOut = Gia_ManSimSimulateOne( p, vPat );
    if Gia_ManSimWriteFile( pFileOut, vPatOut, Gia_ManPoNum(p) ) != 0 {
        fmt.Printf("Output patterns are written into file \"%s\".\n", pFileOut); }
    Vec_IntFree( vPat );
    Vec_IntFree( vPatOut );
}


/**Function*************************************************************

  Synopsis    [Bit-parallel simulation during AIG construction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManBuiltInDataPi(p *Gia_Man_t, iCiId int32) *word {
    return Vec_WrdEntryP( p.vSimsPi, p.nSimWords * iCiId );
}
func Gia_ManBuiltInData(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Gia_ManBuiltInDataPrint(p *Gia_Man_t, iObj int32) {
//    printf( "Obj %6d : ", iObj ); Extra_PrintBinary( stdout, Gia_ManBuiltInData(p, iObj), p->nSimWords * 64 ); printf( "\n" );
}
func Gia_ManBuiltInSimStart(p *Gia_Man_t, nWords int32, nObjs int32) {
    var i, k int32;
    assert(!p.fBuiltInSim, "!p.fBuiltInSim");
    assert(Gia_ManAndNum(p) == 0, "Gia_ManAndNum(p) == 0");
    p.fBuiltInSim = 1;
    p.iPatsPi = 0;
    p.iPastPiMax = 0;
    p.nSimWords = nWords;
    p.nSimWordsMax = 8;
    Gia_ManRandomW( 1 );
    // init PI care info
    p.vSimsPi = Vec_WrdAlloc( p.nSimWords * Gia_ManCiNum(p) );
    Vec_WrdFill( p.vSimsPi, p.nSimWords * Gia_ManCiNum(p), 0 );
    // init object sim info
    p.vSims = Vec_WrdAlloc( p.nSimWords * nObjs );
    Vec_WrdFill( p.vSims, p.nSimWords, 0 );
    for  i = 0; i < Gia_ManCiNum(p); i++  {
        for  k = 0; k < p.nSimWords; k++  {
            Vec_WrdPush( p.vSims, Gia_ManRandomW(0) );
        }
    }
}



func Gia_ManBuiltInSimResimulateCone_rec(p *Gia_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManBuiltInSimResimulateCone_rec( p, Gia_ObjFaninId0(pObj, iObj) );
    Gia_ManBuiltInSimResimulateCone_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    Gia_ManBuiltInSimPerformInt( p, iObj );
}
func Gia_ManBuiltInSimResimulateCone(p *Gia_Man_t, iLit0 int32, iLit1 int32) {
    Gia_ManIncrementTravId( p );
    Gia_ManBuiltInSimResimulateCone_rec( p, Abc_Lit2Var(iLit0) );
    Gia_ManBuiltInSimResimulateCone_rec( p, Abc_Lit2Var(iLit1) );
}
func Gia_ManBuiltInSimResimulate(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;  var iObj int32;
    // inlining of macro Gia_ManForEachAnd
    for  iObj = 0; (iObj < p.nObjs) && ((pObj) = Gia_ManObj(p, iObj)); iObj++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Gia_ManBuiltInSimPerformInt( p, iObj ); }
    }
}

func Gia_ManBuiltInSimCheckOver(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 {
    var pInfo0 *word  = Gia_ManBuiltInData( p, Abc_Lit2Var(iLit0) );
    var pInfo1 *word  = Gia_ManBuiltInData( p, Abc_Lit2Var(iLit1) ); var w int32;
    assert(p.fBuiltInSim || p.fIncrSim, "p.fBuiltInSim || p.fIncrSim");

//    printf( "%d %d\n", Abc_LitIsCompl(iLit0), Abc_LitIsCompl(iLit1) );
//    Extra_PrintBinary( stdout, pInfo0, 64*p->nSimWords ); printf( "\n" );
//    Extra_PrintBinary( stdout, pInfo1, 64*p->nSimWords ); printf( "\n" );
//    printf( "\n" );

    if Abc_LitIsCompl(iLit0) {
        if Abc_LitIsCompl(iLit1) {
            for  w = 0; w < p.nSimWords; w++  {
                if ~pInfo0[w] & ~pInfo1[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < p.nSimWords; w++  {
                if ~pInfo0[w] & pInfo1[w] {
                    return 1; }
            }
        }
    
    } else {
        if Abc_LitIsCompl(iLit1) {
            for  w = 0; w < p.nSimWords; w++  {
                if pInfo0[w] & ~pInfo1[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < p.nSimWords; w++  {
                if pInfo0[w] & pInfo1[w] {
                    return 1; }
            }
        }
    }
    return 0;
}
func Gia_ManBuiltInSimCheckEqual(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 {
    var pInfo0 *word  = Gia_ManBuiltInData( p, Abc_Lit2Var(iLit0) );
    var pInfo1 *word  = Gia_ManBuiltInData( p, Abc_Lit2Var(iLit1) ); var w int32;
    assert(p.fBuiltInSim || p.fIncrSim, "p.fBuiltInSim || p.fIncrSim");

//    printf( "%d %d\n", Abc_LitIsCompl(iLit0), Abc_LitIsCompl(iLit1) );
//    Extra_PrintBinary( stdout, pInfo0, 64*p->nSimWords ); printf( "\n" );
//    Extra_PrintBinary( stdout, pInfo1, 64*p->nSimWords ); printf( "\n" );
//    printf( "\n" );

    if Abc_LitIsCompl(iLit0) {
        if Abc_LitIsCompl(iLit1) {
            for  w = 0; w < p.nSimWords; w++  {
                if ~pInfo0[w] != ~pInfo1[w] {
                    return 0; }
            }
        
        } else {
            for  w = 0; w < p.nSimWords; w++  {
                if ~pInfo0[w] != pInfo1[w] {
                    return 0; }
            }
        }
    
    } else {
        if Abc_LitIsCompl(iLit1) {
            for  w = 0; w < p.nSimWords; w++  {
                if pInfo0[w] != ~pInfo1[w] {
                    return 0; }
            }
        
        } else {
            for  w = 0; w < p.nSimWords; w++  {
                if pInfo0[w] != pInfo1[w] {
                    return 0; }
            }
        }
    }
    return 1;
}

func Gia_ManBuiltInSimPack(p *Gia_Man_t, vPat *Vec_Int_t) int32 {
    var i, k, iLit int32;
    assert(Vec_IntSize(vPat) > 0, "Vec_IntSize(vPat) > 0");
    //printf( "%d ", Vec_IntSize(vPat) );
    for  i = 0; i < p.iPatsPi; i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vPat)) && (((iLit) = Vec_IntEntry(vPat, k)), 1); k++  {
            if Abc_TtGetBit(Gia_ManBuiltInDataPi(p, Abc_Lit2Var(iLit)), i) && 
                 Abc_TtGetBit(Gia_ManBuiltInData(p, 1+Abc_Lit2Var(iLit)), i) == Abc_LitIsCompl(iLit) {
                break; }
        }
        if k == Vec_IntSize(vPat) {
            return i; } // success
    }
    return -1; // failure
}
// adds PI pattern to storage; the pattern comes in terms of CiIds
func Gia_ManBuiltInSimAddPat(p *Gia_Man_t, vPat *Vec_Int_t) int32 {
    var Period int32  = 0xF;
    var fOverflow int32  = p.iPatsPi == 64 * p.nSimWords && p.nSimWords == p.nSimWordsMax;
    var k int32;
    var iLit int32;
    var iPat int32 = Gia_ManBuiltInSimPack( p, vPat );
    if iPat == -1 {
        if fOverflow != 0 {
            if (p.iPastPiMax & Period) == 0 {
                Gia_ManBuiltInSimResimulate( p ); }
            iPat = p.iPastPiMax;
            //if ( p->iPastPiMax == 64 * p->nSimWordsMax - 1 )
            //    printf( "Completed the cycle.\n" );
            if p.iPastPiMax == 64 * p.nSimWordsMax - 1 {
                p.iPastPiMax = 0;
            } else {
                p.iPastPiMax = p.iPastPiMax + 1;
            }
        
        } else {
            if p.iPatsPi && (p.iPatsPi & Period) == 0 {
                Gia_ManBuiltInSimResimulate( p ); }
            if p.iPatsPi == 64 * p.nSimWords {
                var vTemp *Vec_Wrd_t  = Vec_WrdAlloc( 2 * Vec_WrdSize(p.vSims) ); 
                var Data word; var w, Count, iObj int32;
                // inlining of macro Vec_WrdForEachEntry
                for  w = 0; (w < Vec_WrdSize(p.vSims)) && (((Data) = Vec_WrdEntry(p.vSims, w)), 1); w++  {
                    Vec_WrdPush( vTemp, Data );
                    if Count++; Count == p.nSimWords {
                        var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj++);
                        if Gia_ObjIsCi(pObj) {
                            Vec_WrdPush( vTemp, Gia_ManRandomW(0) ); // Vec_WrdPush( vTemp, 0 );
                        } else if Gia_ObjIsAnd(pObj) {
                            Vec_WrdPush( vTemp, pObj.fPhase ? ~(word)0 : 0 );
                        } else {
                            Vec_WrdPush( vTemp, 0 ); }
                        Count = 0;
                    }
                }
                assert(iObj == Gia_ManObjNum(p), "iObj == Gia_ManObjNum(p)");
                Vec_WrdFree( p.vSims );
                p.vSims = vTemp;

                vTemp = Vec_WrdAlloc( 2 * Vec_WrdSize(p.vSimsPi) ); 
                Count = 0;
                // inlining of macro Vec_WrdForEachEntry
                for  w = 0; (w < Vec_WrdSize(p.vSimsPi)) && (((Data) = Vec_WrdEntry(p.vSimsPi, w)), 1); w++  {
                    Vec_WrdPush( vTemp, Data );
                    if Count++; Count == p.nSimWords {
                        Vec_WrdPush( vTemp, 0 );
                        Count = 0;
                    }
                }
                Vec_WrdFree( p.vSimsPi );
                p.vSimsPi = vTemp;

                // update the word count
                p.nSimWords++;
                assert(Vec_WrdSize(p.vSims)   == p.nSimWords * Gia_ManObjNum(p), "Vec_WrdSize(p.vSims)   == p.nSimWords * Gia_ManObjNum(p)");
                assert(Vec_WrdSize(p.vSimsPi) == p.nSimWords * Gia_ManCiNum(p), "Vec_WrdSize(p.vSimsPi) == p.nSimWords * Gia_ManCiNum(p)");
                //printf( "Resizing to %d words.\n", p->nSimWords );
            }
            iPat = p.iPatsPi++;
        }
    }
    assert(iPat >= 0 && iPat < p.iPatsPi, "iPat >= 0 && iPat < p.iPatsPi");
    // add literals 
    if fOverflow != 0 {
        var iVar int32;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(&p.vSuppVars)) && (((iVar) = Vec_IntEntry(&p.vSuppVars, k)), 1); k++  {
            if Abc_TtGetBit(Gia_ManBuiltInDataPi(p, iVar), iPat) {
                 Abc_TtXorBit(Gia_ManBuiltInDataPi(p, iVar), iPat); }
        }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vPat)) && (((iLit) = Vec_IntEntry(vPat, k)), 1); k++  {
            if Abc_TtGetBit(Gia_ManBuiltInData(p, 1+Abc_Lit2Var(iLit)), iPat) == Abc_LitIsCompl(iLit) {
            Abc_TtXorBit(Gia_ManBuiltInData(p, 1+Abc_Lit2Var(iLit)), iPat); }
            Abc_TtXorBit(Gia_ManBuiltInDataPi(p, Abc_Lit2Var(iLit)), iPat);
        }
    
    } else {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vPat)) && (((iLit) = Vec_IntEntry(vPat, k)), 1); k++  {
            if Abc_TtGetBit(Gia_ManBuiltInDataPi(p, Abc_Lit2Var(iLit)), iPat) {
                assert(Abc_TtGetBit(Gia_ManBuiltInData(p, 1+Abc_Lit2Var(iLit)), iPat) != Abc_LitIsCompl(iLit), "Abc_TtGetBit(Gia_ManBuiltInData(p, 1+Abc_Lit2Var(iLit)), iPat) != Abc_LitIsCompl(iLit)");
            } else {
                if Abc_TtGetBit(Gia_ManBuiltInData(p, 1+Abc_Lit2Var(iLit)), iPat) == Abc_LitIsCompl(iLit) {
                Abc_TtXorBit(Gia_ManBuiltInData(p, 1+Abc_Lit2Var(iLit)), iPat); }
                Abc_TtXorBit(Gia_ManBuiltInDataPi(p, Abc_Lit2Var(iLit)), iPat);
            }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Finds a satisfying assignment.]

  Description [Returns 1 if a sat assignment is found; 0 otherwise.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManObjCheckSat_rec(p *Gia_Man_t, iLit int32, vObjs *Vec_Int_t) int32 {
    var iObj int32  = Abc_Lit2Var(iLit);
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if pObj.fMark0 {
        return pObj.fMark1 == uint32(Abc_LitIsCompl(iLit)); }
    pObj.fMark0 = 1;
    pObj.fMark1 = Abc_LitIsCompl(iLit);
    Vec_IntPush( vObjs, iObj );
    if Gia_ObjIsAnd(pObj) {
        if pObj.fMark1 == 0 { // node value is 1
        
            if !(Gia_ManObjCheckSat_rec( p, Gia_ObjFaninLit0(pObj, iObj), vObjs ) != 0) {   return 0; }
            if !(Gia_ManObjCheckSat_rec( p, Gia_ObjFaninLit1(pObj, iObj), vObjs ) != 0) {   return 0; }
        
        } else {
            if !(Gia_ManObjCheckSat_rec( p, Abc_LitNot(Gia_ObjFaninLit0(pObj, iObj)), vObjs ) != 0) {   return 0; }
        }
    }
    return 1;
}
func Gia_ManObjCheckOverlap1(p *Gia_Man_t, iLit0 int32, iLit1 int32, vObjs *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Res0, Res1 int32;
//    Gia_ManForEachObj( p, pObj, i )
//        assert( pObj->fMark0 == 0 && pObj->fMark1 == 0 );
    Vec_IntClear( vObjs );
    Res0 = Gia_ManObjCheckSat_rec( p, iLit0, vObjs );
    if Res0 {
        Res1 = Gia_ManObjCheckSat_rec( p, iLit1, vObjs ); }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.fMark0 = pObj.fMark1 = 0;
    }
    return Res0 && Res1;
}
func Gia_ManObjCheckOverlap(p *Gia_Man_t, iLit0 int32, iLit1 int32, vObjs *Vec_Int_t) int32 {
    if Gia_ManObjCheckOverlap1( p, iLit0, iLit1, vObjs ) != 0 {
        return 1; }
    return Gia_ManObjCheckOverlap1( p, iLit1, iLit0, vObjs );
}





/**Function*************************************************************

  Synopsis    [Bit-parallel simulation during AIG construction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIncrSimUpdate(p *Gia_Man_t) {
    var i, k int32;
    // extend timestamp info
    assert(Vec_IntSize(p.vTimeStamps) <= Gia_ManObjNum(p), "Vec_IntSize(p.vTimeStamps) <= Gia_ManObjNum(p)");
    Vec_IntFillExtra( p.vTimeStamps, Gia_ManObjNum(p), 0 );
    // extend simulation info
    assert(Vec_WrdSize(p.vSims) <= Gia_ManObjNum(p) * p.nSimWords, "Vec_WrdSize(p.vSims) <= Gia_ManObjNum(p) * p.nSimWords");
    Vec_WrdFillExtra( p.vSims,  Gia_ManObjNum(p) * p.nSimWords,  0 );
    // extend PI info
    assert(p.iNextPi <= Gia_ManCiNum(p), "p.iNextPi <= Gia_ManCiNum(p)");
    for  i = p.iNextPi; i < Gia_ManCiNum(p); i++  {
        var pSims *word  = Gia_ManBuiltInData( p, Gia_ManCiIdToId(p, i) );
        for  k = 0; k < p.nSimWords; k++  {
            pSims[k] = Gia_ManRandomW(0);
        }
    }
    p.iNextPi = Gia_ManCiNum(p);
}
func Gia_ManIncrSimStart(p *Gia_Man_t, nWords int32, nObjs int32) {
    assert(!p.fIncrSim, "!p.fIncrSim");
    p.fIncrSim  = 1;
    p.iPatsPi   = 0;
    p.nSimWords = nWords;
    // init time stamps
    p.iTimeStamp  = 1;
    p.vTimeStamps = Vec_IntAlloc( p.nSimWords );
    // init object sim info
    p.iNextPi = 0;       
    p.vSims   = Vec_WrdAlloc( p.nSimWords * nObjs );
    Gia_ManRandomW( 1 );
}
func Gia_ManIncrSimStop(p *Gia_Man_t) {
    assert(p.fIncrSim, "p.fIncrSim");
    p.fIncrSim   = 0;
    p.iPatsPi    = 0;
    p.nSimWords  = 0;
    p.iTimeStamp = 1;
    Vec_IntFreeP( &p.vTimeStamps );
    Vec_WrdFreeP( &p.vSims );
}
func Gia_ManIncrSimSet(p *Gia_Man_t, vObjLits *Vec_Int_t) {
    var i, iLit int32; 
    assert(Vec_IntSize(vObjLits) > 0, "Vec_IntSize(vObjLits) > 0");
    p.iTimeStamp++;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjLits)) && (((iLit) = Vec_IntEntry(vObjLits, i)), 1); i++  {
        var pSims *word  = Gia_ManBuiltInData( p, Abc_Lit2Var(iLit) );
        if Gia_ObjIsAnd(Gia_ManObj(p, Abc_Lit2Var(iLit))) {
            continue; }
        //assert( Vec_IntEntry(p->vTimeStamps, Abc_Lit2Var(iLit)) == p->iTimeStamp-1 );
        Vec_IntWriteEntry(p.vTimeStamps, Abc_Lit2Var(iLit), p.iTimeStamp);
        if Abc_TtGetBit(pSims, p.iPatsPi) == Abc_LitIsCompl(iLit) {
             Abc_TtXorBit(pSims, p.iPatsPi); }
    }
    if p.iPatsPi == p.nSimWords * 64 - 1 {
        p.iPatsPi = 0;
    } else {
        p.iPatsPi = p.iPatsPi + 1;
    }
}
func Gia_ManIncrSimCone_rec(p *Gia_Man_t, iObj int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return; }
    if Vec_IntEntry(p.vTimeStamps, iObj) == p.iTimeStamp {
        return; }
    assert(Vec_IntEntry(p.vTimeStamps, iObj) < p.iTimeStamp, "Vec_IntEntry(p.vTimeStamps, iObj) < p.iTimeStamp");
    Vec_IntWriteEntry( p.vTimeStamps, iObj, p.iTimeStamp );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManIncrSimCone_rec( p, Gia_ObjFaninId0(pObj, iObj) );
    Gia_ManIncrSimCone_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    Gia_ManBuiltInSimPerformInt( p, iObj );
}
func Gia_ManIncrSimCheckOver(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 {
    assert(iLit0 > 1 && iLit1 > 1, "iLit0 > 1 && iLit1 > 1");
    Gia_ManIncrSimUpdate( p );
    Gia_ManIncrSimCone_rec( p, Abc_Lit2Var(iLit0) );
    Gia_ManIncrSimCone_rec( p, Abc_Lit2Var(iLit1) );
//    return 0; // disable
    return Gia_ManBuiltInSimCheckOver( p, iLit0, iLit1 );
}
func Gia_ManIncrSimCheckEqual(p *Gia_Man_t, iLit0 int32, iLit1 int32) int32 {
    assert(iLit0 > 1 && iLit1 > 1, "iLit0 > 1 && iLit1 > 1");
    Gia_ManIncrSimUpdate( p );
    Gia_ManIncrSimCone_rec( p, Abc_Lit2Var(iLit0) );
    Gia_ManIncrSimCone_rec( p, Abc_Lit2Var(iLit1) );
//    return 1; // disable
    return Gia_ManBuiltInSimCheckEqual( p, iLit0, iLit1 );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

