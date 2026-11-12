package ssc

/**CFile****************************************************************

  FileName    [sscSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sweeping under constraints.]

  Synopsis    [Simulation procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: sscSim.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sscInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Ssc_Random() word { return ((word)Gia_ManRandom(0) << 32) | ((word)Gia_ManRandom(0) << 0);              }
func Ssc_Random1(Bit int32) word { return ((word)Gia_ManRandom(0) << 32) | ((word)Gia_ManRandom(0) << 1) | (word)Bit;  }
func Ssc_Random2() word { return ((word)Gia_ManRandom(0) << 32) | ((word)Gia_ManRandom(0) << 2) | (word)2;    }

func Ssc_SimAnd(pSim *word, pSim0 *word, pSim1 *word, nWords int32, fComp0 int32, fComp1 int32) {
    var w int32;
    if (fComp0 != 0) && (fComp1 != 0) { for  w = 0; w < nWords; w++  {  pSim[w] = ~(pSim0[w] | pSim1[w]); }
    } else if fComp0 != 0 {      for  w = 0; w < nWords; w++  {  pSim[w] =  ~pSim0[w] & pSim1[w]; }
    } else if fComp1 != 0 {      for  w = 0; w < nWords; w++  {  pSim[w] =   pSim0[w] &~pSim1[w]; }
    } else {                    for  w = 0; w < nWords; w++  {  pSim[w] =   pSim0[w] & pSim1[w]; } }
}

func Ssc_SimDup(pSim *word, pSim0 *word, nWords int32, fComp0 int32) {
    var w int32;
    if fComp0 != 0 { for  w = 0; w < nWords; w++  {  pSim[w] = ~pSim0[w]; }
    } else {          for  w = 0; w < nWords; w++  {  pSim[w] =  pSim0[w]; } }
}

func Ssc_SimConst(pSim *word, nWords int32, fComp0 int32) {
    var w int32;
    if fComp0 != 0 { for  w = 0; w < nWords; w++  {  pSim[w] = ~(word)0; }
    } else {          for  w = 0; w < nWords; w++  {  pSim[w] = 0; } }
}

func Ssc_SimOr(pSim *word, pSim0 *word, nWords int32, fComp0 int32) {
    var w int32;
    if fComp0 != 0 { for  w = 0; w < nWords; w++  {  pSim[w] |= ~pSim0[w]; }
    } else {          for  w = 0; w < nWords; w++  {  pSim[w] |=  pSim0[w]; } }
}

func Ssc_SimFindBitWord(t word) int32 {
    var n int32;
    if t == 0 { return -1; }
    if (t & 0x00000000FFFFFFFF) == 0 { n += 32; t >>= 32; }
    if (t & 0x000000000000FFFF) == 0 { n += 16; t >>= 16; }
    if (t & 0x00000000000000FF) == 0 { n +=  8; t >>=  8; }
    if (t & 0x000000000000000F) == 0 { n +=  4; t >>=  4; }
    if (t & 0x0000000000000003) == 0 { n +=  2; t >>=  2; }
    if (t & 0x0000000000000001) == 0 { n++; }
    return n;
}
func Ssc_SimFindBit(pSim *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pSim[w] {
            return 64*w + Ssc_SimFindBitWord(pSim[w]); }
    }
    return -1;
}

func Ssc_SimCountBitsWord(x word) int32 {
    x = x - ((x >> 1) & ABC_CONST(0x5555555555555555));   
    x = (x & ABC_CONST(0x3333333333333333)) + ((x >> 2) & ABC_CONST(0x3333333333333333));    
    x = (x + (x >> 4)) & ABC_CONST(0x0F0F0F0F0F0F0F0F);    
    x = x + (x >> 8);
    x = x + (x >> 16);
    x = x + (x >> 32); 
    return int32((x & 0xFF));
}
func Ssc_SimCountBits(pSim *word, nWords int32) int32 {
    var w, Counter int32;
    for  w = 0; w < nWords; w++  {
        Counter += Ssc_SimCountBitsWord(pSim[w]);
    }
    return Counter;
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdDoubleSimInfo(p *Vec_Wrd_t, nObjs int32) {
    // inlining of macro ABC_CALLOC
    var pArray *word  = ((word *) calloc((size_t)(2 * Vec_WrdSize(p)), sizeof(word)));
    var i int32;
    var nWords int32 = Vec_WrdSize(p) / nObjs;
    assert(Vec_WrdSize(p) % nObjs == 0, "Vec_WrdSize(p) % nObjs == 0");
    for  i = 0; i < nObjs; i++  {
        memcpy( pArray + 2*i*nWords, p.pArray + i*nWords, sizeof(word) * nWords );
    }
    // inlining of macro ABC_FREE
    ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0); p.pArray = pArray;
    p.nSize = p.nCap = 2*nWords*nObjs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaResetPiPattern(p *Gia_Man_t, nWords int32) {
    p.iPatsPi = 0;
    if p.vSimsPi == nil {
        p.vSimsPi = Vec_WrdStart(0); }
    Vec_WrdFill( p.vSimsPi, nWords * Gia_ManCiNum(p), 0 );
    assert(nWords == Gia_ObjSimWords( p ), "nWords == Gia_ObjSimWords( p )");
}
func Ssc_GiaSavePiPattern(p *Gia_Man_t, vPat *Vec_Int_t) {
    var pSimPi *word;
    var i int32;
    assert(Vec_IntSize(vPat) == Gia_ManCiNum(p), "Vec_IntSize(vPat) == Gia_ManCiNum(p)");
    if p.iPatsPi == 64 * Gia_ObjSimWords(p) {
        Vec_WrdDoubleSimInfo( p.vSimsPi, Gia_ManCiNum(p) ); }
    assert(p.iPatsPi < 64 * Gia_ObjSimWords(p), "p.iPatsPi < 64 * Gia_ObjSimWords(p)");
    pSimPi = Gia_ObjSimPi( p, 0 );
    for  i = 0; i < Gia_ManCiNum(p); i++, pSimPi += Gia_ObjSimWords(p)  {
        if Vec_IntEntry(vPat, i) {
            Abc_InfoSetBit( (unsigned *)pSimPi, p.iPatsPi ); }
    }
    p.iPatsPi++;
}
func Ssc_GiaRandomPiPattern(p *Gia_Man_t, nWords int32, vPivot *Vec_Int_t) {
    var pSimPi *word;
    var i, w int32;
    Ssc_GiaResetPiPattern( p, nWords );
    pSimPi = Gia_ObjSimPi( p, 0 );
    for  i = 0; i < Gia_ManPiNum(p); i++, pSimPi += nWords  {
        pSimPi[0] = vPivot ? Ssc_Random1(Vec_IntEntry(vPivot, i)) : Ssc_Random2();
        for  w = 1; w < nWords; w++  {
            pSimPi[w] = Ssc_Random();
        }
//        if ( i < 10 )
//            Extra_PrintBinary( stdout, (unsigned *)pSimPi, 64 ), printf( "\n" );
    }
}
func Ssc_GiaPrintPiPatterns(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var pSimAig *word;
    var i int32;//, nWords = Gia_ObjSimWords( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pSimAig = Gia_ObjSimObj( p, pObj );
//        Extra_PrintBinary( stdout, pSimAig, 64 * nWords );
    }
}

/**Function*************************************************************

  Synopsis    [Transfer the simulation pattern from pCare to pAig.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaTransferPiPattern(pAig *Gia_Man_t, pCare *Gia_Man_t, vPivot *Vec_Int_t) int32 {
    extern word * Ssc_GiaGetCareMask( Gia_Man_t * p );
    var pObj *Gia_Obj_t;
    var i int32;
    var w int32;
    var nWords int32 = Gia_ObjSimWords( pCare );
    var pCareMask *word  = Ssc_GiaGetCareMask( pCare );
    var Count int32  = Ssc_SimCountBits( pCareMask, nWords );
    var pSimPiCare, pSimPiAig *word;
    if Count == 0 {
        // inlining of macro ABC_FREE
        ((pCareMask) ? (free((char *) (pCareMask)), (pCareMask) = 0) : 0);
        return 0;
    }
    Ssc_GiaResetPiPattern( pAig, nWords );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pCare.vCis)) && ((pObj) = Gia_ManCi(pCare, i)); i++  {
        pSimPiAig  = Gia_ObjSimPi( pAig, i );
        pSimPiCare = Gia_ObjSimObj( pCare, pObj );
        for  w = 0; w < nWords; w++  {
            if Vec_IntEntry(vPivot, i) {
                pSimPiAig[w] = pSimPiCare[w] | ~pCareMask[w];
            } else {
                pSimPiAig[w] = pSimPiCare[w] & pCareMask[w]; }
        }
    }
    // inlining of macro ABC_FREE
    ((pCareMask) ? (free((char *) (pCareMask)), (pCareMask) = 0) : 0);
    return Count;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaResetSimInfo(p *Gia_Man_t) {
    assert(Vec_WrdSize(p.vSimsPi) % Gia_ManCiNum(p) == 0, "Vec_WrdSize(p.vSimsPi) % Gia_ManCiNum(p) == 0");
    if p.vSims == nil {
        p.vSims = Vec_WrdAlloc(0); }
    Vec_WrdFill( p.vSims, Gia_ObjSimWords(p) * Gia_ManObjNum(p), 0 );
}
func Ssc_GiaSimRound(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var pSim, pSim0, pSim1 *word;
    var i int32;
    var nWords int32 = Gia_ObjSimWords(p);
    Ssc_GiaResetSimInfo( p );
    assert(nWords == Vec_WrdSize(p.vSims) / Gia_ManObjNum(p), "nWords == Vec_WrdSize(p.vSims) / Gia_ManObjNum(p)");
    // constant node
    Ssc_SimConst( Gia_ObjSim(p, 0), nWords, 0 );
    // primary inputs
    pSim = Gia_ObjSim( p, 1 );
    pSim0 = Gia_ObjSimPi( p, 0 );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        assert(pSim == Gia_ObjSimObj( p, pObj ), "pSim == Gia_ObjSimObj( p, pObj )");
        Ssc_SimDup( pSim, pSim0, nWords, 0 );
        pSim += nWords;
        pSim0 += nWords;
    }
    // intermediate nodes
    pSim = Gia_ObjSim( p, 1+Gia_ManCiNum(p) );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(pSim == Gia_ObjSim( p, i ), "pSim == Gia_ObjSim( p, i )");
        pSim0 = pSim - pObj.iDiff0 * nWords;
        pSim1 = pSim - pObj.iDiff1 * nWords;
        Ssc_SimAnd( pSim, pSim0, pSim1, nWords, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj) );
        pSim += nWords;
    }
    }
    // primary outputs
    pSim = Gia_ObjSim( p, Gia_ManObjNum(p) - Gia_ManPoNum(p) );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        assert(pSim == Gia_ObjSimObj( p, pObj ), "pSim == Gia_ObjSimObj( p, pObj )");
        pSim0 = pSim - pObj.iDiff0 * nWords;
        Ssc_SimDup( pSim, pSim0, nWords, Gia_ObjFaninC0(pObj) );
//        Extra_PrintBinary( stdout, pSim, 64 ), printf( "\n" );
        pSim += nWords;
    }
}

/**Function*************************************************************

  Synopsis    [Returns one SAT assignment of the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaGetCareMask(p *Gia_Man_t) *word {
    var pObj *Gia_Obj_t;
    var i int32;
    var nWords int32 = Gia_ObjSimWords( p );
    // inlining of macro ABC_FALLOC
    var pRes *word  = ((word *) memset(malloc(sizeof(word) * (size_t)(nWords)), 0xff, sizeof(word) * (size_t)(nWords)));
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Ssc_SimAnd( pRes, pRes, Gia_ObjSimObj(p, pObj), nWords, 0, 0 );
    }
    return pRes;
}
func Ssc_GiaGetOneSim(p *Gia_Man_t) *Vec_Int_t {
    var vInit *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var iBit int32;
    var nWords int32 = Gia_ObjSimWords( p );
    var pRes *word  = Ssc_GiaGetCareMask( p );
    iBit = Ssc_SimFindBit( pRes, nWords );
    // inlining of macro ABC_FREE
    ((pRes) ? (free((char *) (pRes)), (pRes) = 0) : 0);
    if iBit == -1 {
        return nil; }
    vInit = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_IntPush( vInit, Abc_InfoHasBit((unsigned *)Gia_ObjSimObj(p, pObj), iBit) );
    }
    return vInit;
}
func Ssc_GiaFindPivotSim(p *Gia_Man_t) *Vec_Int_t {
    var vInit *Vec_Int_t;
    Ssc_GiaRandomPiPattern( p, 1, nil );
    Ssc_GiaSimRound( p );
    vInit = Ssc_GiaGetOneSim( p );
    return vInit;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaCountCaresSim(p *Gia_Man_t) int32 {
    var pRes *word  = Ssc_GiaGetCareMask( p );
    var nWords int32  = Gia_ObjSimWords( p );
    var Count int32  = Ssc_SimCountBits( pRes, nWords );
    // inlining of macro ABC_FREE
    ((pRes) ? (free((char *) (pRes)), (pRes) = 0) : 0);
    return Count;
}
func Ssc_GiaEstimateCare(p *Gia_Man_t, nWords int32) int32 {
    Ssc_GiaRandomPiPattern( p, nWords, nil );
    Ssc_GiaSimRound( p );
    return Ssc_GiaCountCaresSim( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

