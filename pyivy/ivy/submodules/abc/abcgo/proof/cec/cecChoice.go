package cec

/**CFile****************************************************************

  FileName    [cecChoice.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Computation of structural choices.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecChoice.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "aig/gia/giaAig.h"
//#include "proof/dch/dch.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Cec_ManCombSpecReduce_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj );


// extern int Cec_ManResimulateCounterExamplesComb( Cec_ManSim_t * pSim, Vec_Int_t * vCexStore );

// extern int Gia_ManCheckRefinements( Gia_Man_t * p, Vec_Str_t * vStatus, Vec_Int_t * vOutputs, Cec_ManSim_t * pSim, int fRings );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes the real value of the literal w/o spec reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCombSpecReal(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Cec_ManCombSpecReduce_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Cec_ManCombSpecReduce_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Recursively performs speculative reduction for the object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCombSpecReduce_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pRepr *Gia_Obj_t;
    if ~pObj.Value {
        return; }
    if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
        Cec_ManCombSpecReduce_rec( pNew, p, pRepr );
        pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
        return;
    }
    pObj.Value = Cec_ManCombSpecReal( pNew, p, pObj );
}

/**Function*************************************************************

  Synopsis    [Derives SRM for signal correspondence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCombSpecReduce(p *Gia_Man_t, pvOutputs **Vec_Int_t, fRings int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var vXorLits *Vec_Int_t;
    var i, iPrev, iObj, iPrevNew, iObjNew int32;
    assert(p.pReprs != nil, "p.pReprs != nil");
    Gia_ManSetPhase( p );
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    *pvOutputs = Vec_IntAlloc( 1000 );
    vXorLits = Vec_IntAlloc( 1000 );
    if fRings != 0 {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if Gia_ObjIsConst( p, i ) {
                iObjNew = Cec_ManCombSpecReal( pNew, p, pObj );
                iObjNew = Abc_LitNotCond( iObjNew, Gia_ObjPhase(pObj) );
                if iObjNew != 0 {
                    Vec_IntPush( *pvOutputs, 0 );
                    Vec_IntPush( *pvOutputs, i );
                    Vec_IntPush( vXorLits, iObjNew );
                }
            
            } else if Gia_ObjIsHead( p, i ) {
                iPrev = i;
                // inlining of macro Gia_ClassForEachObj1
                for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), iObj = Gia_ObjNext(p, i); iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
                    iPrevNew = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iPrev) );
                    iObjNew  = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iObj) );
                    iPrevNew = Abc_LitNotCond( iPrevNew, Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iPrev)) );
                    iObjNew  = Abc_LitNotCond( iObjNew,  Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iObj)) );
                    if iPrevNew != iObjNew && iPrevNew != 0 && iObjNew != 1 {
                        Vec_IntPush( *pvOutputs, iPrev );
                        Vec_IntPush( *pvOutputs, iObj );
                        Vec_IntPush( vXorLits, Gia_ManHashAnd(pNew, iPrevNew, Abc_LitNot(iObjNew)) );
                    }
                    iPrev = iObj;
                }
                iObj = i;
                iPrevNew = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iPrev) );
                iObjNew  = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iObj) );
                iPrevNew = Abc_LitNotCond( iPrevNew, Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iPrev)) );
                iObjNew  = Abc_LitNotCond( iObjNew,  Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iObj)) );
                if iPrevNew != iObjNew && iPrevNew != 0 && iObjNew != 1 {
                    Vec_IntPush( *pvOutputs, iPrev );
                    Vec_IntPush( *pvOutputs, iObj );
                    Vec_IntPush( vXorLits, Gia_ManHashAnd(pNew, iPrevNew, Abc_LitNot(iObjNew)) );
                }
            }
        }
    
    } else {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
            if pRepr == nil {
                continue; }
            if Gia_ObjIsConst(p, i) {
                iPrevNew = 0;
            } else {
                iPrevNew = Cec_ManCombSpecReal( pNew, p, pRepr );
            }
            iObjNew  = Cec_ManCombSpecReal( pNew, p, pObj );
            iObjNew  = Abc_LitNotCond( iObjNew, Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
            if iPrevNew != iObjNew {
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pRepr) );
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pObj) );
                Vec_IntPush( vXorLits, Gia_ManHashXor(pNew, iPrevNew, iObjNew) );
            }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorLits)) && (((iObjNew) = Vec_IntEntry(vXorLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iObjNew );
    }
    Vec_IntFree( vXorLits );
    Gia_ManHashStop( pNew );
//Abc_Print( 1, "Before sweeping = %d\n", Gia_ManAndNum(pNew) );
    pNew = Gia_ManCleanup( pTemp = pNew );
//Abc_Print( 1, "After sweeping = %d\n", Gia_ManAndNum(pNew) );
    Gia_ManStop( pTemp );
    return pNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManChoiceComputation_int(pAig *Gia_Man_t, pPars *Cec_ParChc_t) int32 {
    var nItersMax int32  = 1000;
    var vStatus *Vec_Str_t;
    var vOutputs *Vec_Int_t;
    var vCexStore *Vec_Int_t;
    var ParsSim Cec_ParSim_t;
    var pParsSim *Cec_ParSim_t = &ParsSim;
    var ParsSat Cec_ParSat_t;
    var pParsSat *Cec_ParSat_t = &ParsSat;
    var pSim *Cec_ManSim_t;
    var pSrm *Gia_Man_t;
    var r, RetValue int32;
    var clkSat abctime;
    var clkSim abctime;
    var clkSrm abctime;
    var clkTotal abctime = Abc_Clock();
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    // inlining of macro ABC_FREE
    ((pAig.pReprs) ? (free((char *) (pAig.pReprs)), (pAig.pReprs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pAig.pNexts) ? (free((char *) (pAig.pNexts)), (pAig.pNexts) = 0) : 0);
    Gia_ManRandom( 1 );
    // prepare simulation manager
    Cec_ManSimSetDefaultParams( pParsSim );
    pParsSim.nWords     = pPars.nWords;
    pParsSim.nFrames    = pPars.nRounds;
    pParsSim.fVerbose   = pPars.fVerbose;
    pParsSim.fLatchCorr   = 0;
    pParsSim.fSeqSimulate = 0;
    // create equivalence classes of registers
    pSim = Cec_ManSimStart( pAig, pParsSim );
    Cec_ManSimClassesPrepare( pSim, -1 );
    Cec_ManSimClassesRefine( pSim );
    // prepare SAT solving
    Cec_ManSatSetDefaultParams( pParsSat );
    pParsSat.nBTLimit = pPars.nBTLimit;
    pParsSat.fVerbose = pPars.fVerbose;
    if pPars.fVerbose {
        Abc_Print( 1, "Obj = %7d. And = %7d. Conf = %5d. Ring = %d. CSat = %d.\n",
            Gia_ManObjNum(pAig), Gia_ManAndNum(pAig), pPars.nBTLimit, pPars.fUseRings, pPars.fUseCSat );
        Cec_ManRefinedClassPrintStats( pAig, nil, 0, Abc_Clock() - clk );
    }
    // perform refinement of equivalence classes
    for  r = 0; r < nItersMax; r++  { 
        clk = Abc_Clock();
        // perform speculative reduction
        clk2 = Abc_Clock();
        pSrm = Cec_ManCombSpecReduce( pAig, &vOutputs, pPars.fUseRings );
        assert(Gia_ManRegNum(pSrm) == 0 && Gia_ManCiNum(pSrm) == Gia_ManCiNum(pAig), "Gia_ManRegNum(pSrm) == 0 && Gia_ManCiNum(pSrm) == Gia_ManCiNum(pAig)");
        clkSrm += Abc_Clock() - clk2;
        if Gia_ManCoNum(pSrm) == 0 {
            if pPars.fVerbose {
                Cec_ManRefinedClassPrintStats( pAig, nil, r+1, Abc_Clock() - clk ); }
            Vec_IntFree( vOutputs );
            Gia_ManStop( pSrm );            
            break;
        }
//Gia_DumpAiger( pSrm, "choicesrm", r, 2 );
        // found counter-examples to speculation
        clk2 = Abc_Clock();
        if pPars.fUseCSat {
            vCexStore = Cbs_ManSolveMiterNc( pSrm, pPars.nBTLimit, &vStatus, 0, 0 );
        } else {
            vCexStore = Cec_ManSatSolveMiter( pSrm, pParsSat, &vStatus ); }
        Gia_ManStop( pSrm );
        clkSat += Abc_Clock() - clk2;
        if Vec_IntSize(vCexStore) == 0 {
            if pPars.fVerbose {
                Cec_ManRefinedClassPrintStats( pAig, vStatus, r+1, Abc_Clock() - clk ); }
            Vec_IntFree( vCexStore );
            Vec_StrFree( vStatus );
            Vec_IntFree( vOutputs );
            break;
        }
        // refine classes with these counter-examples
        clk2 = Abc_Clock();
        RetValue = Cec_ManResimulateCounterExamplesComb( pSim, vCexStore );
        Vec_IntFree( vCexStore );
        clkSim += Abc_Clock() - clk2;
        Gia_ManCheckRefinements( pAig, vStatus, vOutputs, pSim, pPars.fUseRings );
        if pPars.fVerbose {
            Cec_ManRefinedClassPrintStats( pAig, vStatus, r+1, Abc_Clock() - clk ); }
        Vec_StrFree( vStatus );
        Vec_IntFree( vOutputs );
//Gia_ManEquivPrintClasses( pAig, 1, 0 );
    }
    // check the overflow
    if r == nItersMax {
        Abc_Print( 1, "The refinement was not finished. The result may be incorrect.\n" ); }
    Cec_ManSimStop( pSim );
    clkTotal = Abc_Clock() - clkTotal;
    // report the results
    if pPars.fVerbose {
        Abc_PrintTimeP( 1, "Srm  ", clkSrm,                        clkTotal );
        Abc_PrintTimeP( 1, "Sat  ", clkSat,                        clkTotal );
        Abc_PrintTimeP( 1, "Sim  ", clkSim,                        clkTotal );
        Abc_PrintTimeP( 1, "Other", clkTotal-clkSat-clkSrm-clkSim, clkTotal );
        Abc_PrintTime( 1, "TOTAL",  clkTotal );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Computes choices for the vector of AIGs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManChoiceComputationVec(pGia *Gia_Man_t, nGias int32, pPars *Cec_ParChc_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var RetValue int32;
    // compute equivalences of the miter
//    pMiter = Gia_ManChoiceMiter( vGias );
//    Gia_ManSetRegNum( pMiter, 0 );
    RetValue = Cec_ManChoiceComputation_int( pGia, pPars );
    // derive AIG with choices
    pNew = Gia_ManEquivToChoices( pGia, nGias );
//    Gia_ManHasChoices_very_old( pNew );
//    Gia_ManStop( pMiter );
    // report the results
    if pPars.fVerbose {
//        Abc_Print( 1, "NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", 
//            Gia_ManAndNum(pAig), Gia_ManAndNum(pNew), 
//            100.0*(Gia_ManAndNum(pAig)-Gia_ManAndNum(pNew))/(Gia_ManAndNum(pAig)?Gia_ManAndNum(pAig):1), 
//            Gia_ManRegNum(pAig), Gia_ManRegNum(pNew), 
//            100.0*(Gia_ManRegNum(pAig)-Gia_ManRegNum(pNew))/(Gia_ManRegNum(pAig)?Gia_ManRegNum(pAig):1) );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Computes choices for one AIGs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManChoiceComputation(pAig *Gia_Man_t, pParsChc *Cec_ParChc_t) *Gia_Man_t {
//    extern Aig_Man_t * Dar_ManChoiceNew( Aig_Man_t * pAig, Dch_Pars_t * pPars );
    var Pars Dch_Pars_t;
    var pPars *Dch_Pars_t = &Pars;
    var pMan, pManNew *Aig_Man_t;
    var pGia *Gia_Man_t;
    if false {
        pGia = Cec_ManChoiceComputationVec( pAig, 3, pParsChc );
    
    } else {
        pMan = Gia_ManToAig( pAig, 0 );
        Dch_ManSetDefaultParams( pPars );
        pPars.fUseGia  = 1;
        pPars.nBTLimit = pParsChc.nBTLimit;
        pPars.fUseCSat = pParsChc.fUseCSat;
        pPars.fVerbose = pParsChc.fVerbose;
        pManNew = Dar_ManChoiceNew( pMan, pPars );
        pGia = Gia_ManFromAig( pManNew );
        Aig_ManStop( pManNew );
//        Aig_ManStop( pMan );
    }
    return pGia;
}

/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description [Takes several AIGs and performs choicing.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ComputeChoices(pGia *Gia_Man_t, pPars *Dch_Pars_t) *Aig_Man_t {
    var ParsChc Cec_ParChc_t;
    var pParsChc *Cec_ParChc_t = &ParsChc;
    var pAig *Aig_Man_t;
    if pPars.fVerbose {
        Abc_PrintTime( 1, "Synthesis time", pPars.timeSynth ); }
    Cec_ManChcSetDefaultParams( pParsChc );
    pParsChc.nBTLimit = pPars.nBTLimit;
    pParsChc.fUseCSat = pPars.fUseCSat;
    if pParsChc.fUseCSat && pParsChc.nBTLimit > 100 {
        pParsChc.nBTLimit = 100; }
    pParsChc.fVerbose = pPars.fVerbose;
    pGia = Cec_ManChoiceComputationVec( pGia, 3, pParsChc );
    Gia_ManSetRegNum( pGia, Gia_ManRegNum(pGia) );
    pAig = Gia_ManToAig( pGia, 1 );
    Gia_ManStop( pGia );
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description [Takes several AIGs and performs choicing.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ComputeChoicesNew(pGia *Gia_Man_t, fVerbose int32) *Aig_Man_t {
    extern void Cec4_ManSimulateTest2( Gia_Man_t * p, int fVerbose );
    var pAig *Aig_Man_t;
    Cec4_ManSimulateTest2( pGia, fVerbose );
    pGia = Gia_ManEquivToChoices( pGia, 3 );
    Gia_ManSetRegNum( pGia, Gia_ManRegNum(pGia) );
    pAig = Gia_ManToAig( pGia, 1 );
    Gia_ManStop( pGia );
    return pAig;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

