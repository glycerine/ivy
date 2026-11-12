package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaScript.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Various hardcoded scripts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaScript.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "giaAig.h"
//#include "base/main/main.h"
//#include "base/cmd/cmd.h"
//#include "proof/dch/dch.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Synthesis script.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Synthesis script.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAigSyn2(pInit *Gia_Man_t, fOldAlgo int32, fCoarsen int32, fCutMin int32, nRelaxRatio int32, fDelayMin int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var p, pNew, pTemp *Gia_Man_t;
    var Pars Jf_Par_t;
    var pPars *Jf_Par_t = &Pars;
    if fOldAlgo != 0 {
        Jf_ManSetDefaultPars( pPars );
        pPars.fCutMin     = fCutMin;
    
    } else {
        Lf_ManSetDefaultPars( pPars );
        pPars.fCutMin     = fCutMin;
        pPars.fCoarsen    = fCoarsen;
        pPars.nRelaxRatio = nRelaxRatio;
        pPars.nAreaTuner  = 1;
        pPars.nCutNum     = 4;
    }
    if fVerbose != 0 {  Gia_ManPrintStats( pInit, nil ); }
    p = Gia_ManDup( pInit );
    Gia_ManTransferTiming( p, pInit );
    if Gia_ManAndNum(p) == 0 {
        pNew = Gia_ManDup(p);
        Gia_ManTransferTiming( pNew, p );
        Gia_ManStop( p );
        return pNew;
    }
    // delay optimization
    if (fDelayMin != 0) && p.pManTime == nil {
        var Area0, Area1, Delay0, Delay1 int32;
        var fCutMin int32  = pPars.fCutMin;
        var fCoarsen int32  = pPars.fCoarsen;
        var nRelaxRatio int32  = pPars.nRelaxRatio;
        pPars.fCutMin = 0;
        pPars.fCoarsen = 0;
        pPars.nRelaxRatio = 0;
        // perform mapping
        if fOldAlgo != 0 {
            Jf_ManPerformMapping( p, pPars );
        } else {
            Lf_ManPerformMapping( p, pPars ); }
        Area0  = int32(pPars.Area);
        Delay0 = int32(pPars.Delay);
        // perform balancing
        pNew = Gia_ManPerformDsdBalance( p, 6, 4, 0, 0 );
        // perform mapping again
        if fOldAlgo != 0 {
            Jf_ManPerformMapping( pNew, pPars );
        } else {
            Lf_ManPerformMapping( pNew, pPars ); }
        Area1  = int32(pPars.Area);
        Delay1 = int32(pPars.Delay);
        // choose the best result
        if Delay1 < Delay0 - 1 || (Delay1 == Delay0 + 1 && 100.0 * (Area1 - Area0) / Area1 < 3.0) {
            Gia_ManStop( p );
            p = pNew;
        
        } else {
            Gia_ManStop( pNew );
            Vec_IntFreeP( &p.vMapping );
        }
        // reset params
        pPars.fCutMin = fCutMin;
        pPars.fCoarsen = fCoarsen;
        pPars.nRelaxRatio = nRelaxRatio;
    }
    // perform balancing
    pNew = Gia_ManAreaBalance( p, 0, ABC_INFINITY, fVeryVerbose, 0 );
    if fVerbose != 0 {     Gia_ManPrintStats( pNew, nil ); }
    Gia_ManStop( p );
    // perform mapping
    if fOldAlgo != 0 {
        pNew = Jf_ManPerformMapping( pTemp = pNew, pPars );
    } else {
        pNew = Lf_ManPerformMapping( pTemp = pNew, pPars ); }
    if fVerbose != 0 {     Gia_ManPrintStats( pNew, nil ); }
    if pTemp != pNew {
        Gia_ManStop( pTemp ); }
    // perform balancing
    pNew = Gia_ManAreaBalance( pTemp = pNew, 0, ABC_INFINITY, fVeryVerbose, 0 );
    if fVerbose != 0 {     Gia_ManPrintStats( pNew, nil ); }
    Gia_ManStop( pTemp );
    return pNew;
}
/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [This duplicator works for AIGs with choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManOrderPios(p *Aig_Man_t, pOrder *Gia_Man_t) *Vec_Ptr_t {
    var vPios *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Aig_ManCiNum(p) == Gia_ManCiNum(pOrder), "Aig_ManCiNum(p) == Gia_ManCiNum(pOrder)");
    assert(Aig_ManCoNum(p) == Gia_ManCoNum(pOrder), "Aig_ManCoNum(p) == Gia_ManCoNum(pOrder)");
    vPios = Vec_PtrAlloc( Aig_ManCiNum(p) + Aig_ManCoNum(p) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pOrder.nObjs) && ((pObj) = Gia_ManObj(pOrder, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            Vec_PtrPush( vPios, Aig_ManCi(p, Gia_ObjCioId(pObj)) );
        } else if Gia_ObjIsCo(pObj) {
            Vec_PtrPush( vPios, Aig_ManCo(p, Gia_ObjCioId(pObj)) ); }
    }
    return vPios;
}

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
func Gia_ManAigSynch2Choices(pGia1 *Gia_Man_t, pGia2 *Gia_Man_t, pGia3 *Gia_Man_t, pPars *Dch_Pars_t) *Gia_Man_t {
    var pMan, pTemp *Aig_Man_t;
    var pGia, pMiter *Gia_Man_t;
    // derive miter
    var vPios *Vec_Ptr_t;
    var vGias *Vec_Ptr_t = Vec_PtrAlloc( 3 );
    if pGia3 != nil { Vec_PtrPush( vGias, pGia3 ); }
    if pGia2 != nil { Vec_PtrPush( vGias, pGia2 ); }
    if pGia1 != nil { Vec_PtrPush( vGias, pGia1 ); }
    pMiter = Gia_ManChoiceMiter( vGias );
    Vec_PtrFree( vGias );
    // transform into an AIG
    pMan = Gia_ManToAigSkip( pMiter, 3 );
    Gia_ManStop( pMiter );
    // compute choices
    pMan = Dch_ComputeChoices( pTemp = pMan, pPars );
    Aig_ManStop( pTemp );
    // reconstruct the network
    vPios = Gia_ManOrderPios( pMan, pGia1 ); 
    pMan = Aig_ManDupDfsGuided( pTemp = pMan, vPios );
    Aig_ManStop( pTemp );
    Vec_PtrFree( vPios );
    // convert to GIA
    pGia = Gia_ManFromAigChoices( pMan );
    Aig_ManStop( pMan );
    return pGia;
}
func Gia_ManAigSynch2(pInit *Gia_Man_t, pPars0 any, nLutSize int32, nRelaxRatio int32) *Gia_Man_t {
    extern Gia_Man_t * Gia_ManLutBalance( Gia_Man_t * p, int nLutSize, int fUseMuxes, int fRecursive, int fOptArea, int fVerbose );
    var pParsDch *Dch_Pars_t  = (Dch_Pars_t *)pPars0;
    var pGia1, pGia2, pGia3, pNew, pTemp *Gia_Man_t;
    var fVerbose int32  = pParsDch.fVerbose;
    var Pars Jf_Par_t;
    var pPars *Jf_Par_t = &Pars;
    Lf_ManSetDefaultPars( pPars );
    pPars.fCutMin     = 1;
    pPars.fCoarsen    = 1;
    pPars.nRelaxRatio = nRelaxRatio;
    pPars.nAreaTuner  = 5;
    pPars.nCutNum     = 12;
    pPars.fVerbose    = fVerbose;
    if fVerbose != 0 {  Gia_ManPrintStats( pInit, nil ); }
    pGia1 = Gia_ManDup( pInit );
    if Gia_ManAndNum(pGia1) == 0 {
        Gia_ManTransferTiming( pGia1, pInit );
        return pGia1;
    }
    if pGia1.pManTime && pGia1.vLevels == nil {
        Gia_ManLevelWithBoxes( pGia1 ); }
    // unmap if mapped
    if Gia_ManHasMapping(pInit) {
        Gia_ManTransferMapping( pGia1, pInit );
        pGia1 = (Gia_Man_t *)Dsm_ManDeriveGia( pTemp = pGia1, 0 );
        Gia_ManStop( pTemp );
    }
    // perform balancing
    if Gia_ManBufNum(pGia1) || true {
        pGia2 = Gia_ManAreaBalance( pGia1, 0, ABC_INFINITY, 0, 0 );
    } else {
        pGia2 = Gia_ManLutBalance( pGia1, nLutSize, 1, 1, 1, 0 );
        pGia2 = Gia_ManAreaBalance( pTemp = pGia2, 0, ABC_INFINITY, 0, 0 );
        Gia_ManStop( pTemp );
    }
    if fVerbose != 0 {     Gia_ManPrintStats( pGia2, nil ); }
    // perform mapping
    pGia2 = Lf_ManPerformMapping( pTemp = pGia2, pPars );
    if fVerbose != 0 {     Gia_ManPrintStats( pGia2, nil ); }
    if pTemp != pGia2 {
        Gia_ManStop( pTemp ); }
    // perform balancing
    if pParsDch.fLightSynth || Gia_ManBufNum(pGia2) {
        pGia3 = Gia_ManAreaBalance( pGia2, 0, ABC_INFINITY, 0, 0 );
    } else {
        assert(Gia_ManBufNum(pGia2) == 0, "Gia_ManBufNum(pGia2) == 0");
        pGia2 = Gia_ManAreaBalance( pTemp = pGia2, 0, ABC_INFINITY, 0, 0 );
        if fVerbose != 0 {     Gia_ManPrintStats( pGia2, nil ); }
        Gia_ManStop( pTemp );
        // perform DSD balancing
        pGia3 = Gia_ManPerformDsdBalance( pGia2, 6, 8, 0, 0 );
    }
    if fVerbose != 0 {     Gia_ManPrintStats( pGia3, nil ); }
    // perform choice computation
    if Gia_ManBufNum(pInit) {
        assert(Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia1), "Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia1)");
        pGia1 = Gia_ManDupFromBarBufs( pTemp = pGia1 );
        Gia_ManStop( pTemp );
        assert(Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia2), "Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia2)");
        pGia2 = Gia_ManDupFromBarBufs( pTemp = pGia2 );
        Gia_ManStop( pTemp );
        assert(Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia3), "Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia3)");
        pGia3 = Gia_ManDupFromBarBufs( pTemp = pGia3 );
        Gia_ManStop( pTemp );
    }
    pNew = Gia_ManAigSynch2Choices( pGia1, pGia2, pGia3, pParsDch );
    Gia_ManStop( pGia1 );
    Gia_ManStop( pGia2 );
    Gia_ManStop( pGia3 );
    if Gia_ManBufNum(pInit) {
        pNew = Gia_ManDupToBarBufs( pTemp = pNew, Gia_ManBufNum(pInit) );
        Gia_ManStop( pTemp );
    }
    // copy names
    // inlining of macro ABC_FREE
    pNew.pName = "";
    // inlining of macro ABC_FREE
    pNew.pSpec = "";
    pNew.pName = Abc_UtilStrsav(pInit.pName);
    pNew.pSpec = Abc_UtilStrsav(pInit.pSpec);
    Gia_ManTransferTiming( pNew, pInit );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformMap(nAnds int32, nLutSize int32, nCutNum int32, fMinAve int32, fUseMfs int32, fVerbose int32) {
    var Command string // was [200]byte in C

    Command = fmt.Sprintf("&unmap; &lf -K %d -C %d -k %s; &save", nLutSize, nCutNum, fMinAve?"-t":"");
//    sprintf( Command, "&unmap; &if -K %d -C %d %s; &save", nLutSize, nCutNum, fMinAve?"-t":"" );
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
    if fVerbose != 0 {
        fmt.Printf("MAPPING:\n");
        fmt.Printf("Mapping with &lf -k:\n");
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" );
    }
    Command = fmt.Sprintf("&unmap; &lf -K %d -C %d %s; &save", nLutSize, nCutNum, fMinAve?"-t":"");
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
    if fVerbose != 0 {
        fmt.Printf("Mapping with &lf:\n");
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" );
    }
    if (nLutSize == 4 && nAnds < 100000) || (nLutSize == 6 && nAnds < 2000) {
        Command = fmt.Sprintf("&unmap; &if -sz -S %d%d -K %d -C %d %s", nLutSize, nLutSize, 2*nLutSize-1, 2*nCutNum, fMinAve?"-t":"");
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
        Vec_IntFreeP( &Abc_FrameReadGia(Abc_FrameGetGlobalFrame()).vPacking );
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&save" );
        if fVerbose != 0 {
            fmt.Printf("Mapping with &if -sz -S %d%d -K %d -C %d %s:\n", nLutSize, nLutSize, 2*nLutSize-1, 2*nCutNum, fMinAve?"-t":"");
            Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" );
        }
    }
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&load" );
    if fUseMfs != 0 {
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&put; mfs2 -W 4 -M 500 -C 7000; &get -m" ); }
    if fVerbose != 0 {
        fmt.Printf("Mapping final:\n");
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" );
    }
}
func Gia_ManPerformRound(fIsMapped int32, nAnds int32, nLevels int32, nLutSize int32, nCutNum int32, fMinAve int32, fUseMfs int32, fVerbose int32) {
    var Command string // was [200]byte in C


    // perform AIG-based synthesis
    if nAnds < 50000 {
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "" );
        Command = fmt.Sprintf("&dsdb; &dch -C 500; &if -K %d -C %d %s; &save", nLutSize, nCutNum, fMinAve?"-t":"");
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
        if fVerbose != 0 {
            fmt.Printf("Mapping with &dch -C 500; &if -K %d -C %d %s:\n", nLutSize, nCutNum, fMinAve?"-t":"");
            Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" );
        }
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );
    }

    // perform AIG-based synthesis
    if nAnds < 20000 {
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "" );
        Command = fmt.Sprintf("&dsdb; &dch -C 500; &if -K %d -C %d %s; &save", nLutSize, nCutNum, fMinAve?"-t":"");
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
        if fVerbose != 0 {
            fmt.Printf("Mapping with &dch -C 500; &if -K %d -C %d %s:\n", nLutSize, nCutNum, fMinAve?"-t":"");
            Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" );
        }
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );
    }

    // perform first round of mapping
    Gia_ManPerformMap( nAnds, nLutSize, nCutNum, fMinAve, fUseMfs, fVerbose );
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );

    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&dsdb" );

    // perform second round of mapping
    Gia_ManPerformMap( nAnds, nLutSize, nCutNum, fMinAve, fUseMfs, fVerbose );
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );

    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&syn2 -m -R 10; &dsdb" );

    // prepare for final mapping
    Command = fmt.Sprintf("&blut -a -K %d", nLutSize);
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );

    // perform third round of mapping
    Gia_ManPerformMap( nAnds, nLutSize, nCutNum, fMinAve, fUseMfs, fVerbose );
}
func Gia_ManPerformFlow(fIsMapped int32, nAnds int32, nLevels int32, nLutSize int32, nCutNum int32, fMinAve int32, fUseMfs int32, fVerbose int32) {
    // remove comb equivs
    if fIsMapped != 0 {
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" ); }
//    if ( Abc_FrameReadGia(Abc_FrameGetGlobalFrame())->pManTime )
//        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&sweep" );
//    else
//        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&fraig -c" );

    // perform first round
    Gia_ManPerformRound( fIsMapped, nAnds, nLevels, nLutSize, nCutNum, fMinAve, fUseMfs, fVerbose );

    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st; &sopb" );

    // perform first round
    Gia_ManPerformRound( fIsMapped, nAnds, nLevels, nLutSize, nCutNum, fMinAve, fUseMfs, fVerbose );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformFlow2(fIsMapped int32, nAnds int32, nLevels int32, nLutSize int32, nCutNum int32, fBalance int32, fMinAve int32, fUseMfs int32, fVerbose int32) {
    var Comm1, Comm2, Comm3, Comm4 [1000]byte;
    if nLutSize == 0 {
        Comm1 = fmt.Sprintf("&synch2 -K 6 -C 500;  &if -m%s       -C %d; %s &save", fMinAve?"t":"", nCutNum, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
        Comm2 = fmt.Sprintf("&dch -C 500;          &if -m%s       -C %d; %s &save", fMinAve?"t":"", nCutNum+4, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
        Comm3 = fmt.Sprintf("&synch2 -K 6 -C 500;  &lf -m%s  -E 5 -C %d; %s &save", fMinAve?"t":"", nCutNum, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
        Comm4 = fmt.Sprintf("&dch -C 500;          &lf -m%sk -E 5 -C %d; %s &save", fMinAve?"t":"", nCutNum+4, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
    
    } else {
        Comm1 = fmt.Sprintf("&synch2 -K %d -C 500; &if -m%s       -K %d -C %d; %s &save", nLutSize, fMinAve?"t":"", nLutSize, nCutNum, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
        Comm2 = fmt.Sprintf("&dch -C 500;          &if -m%s       -K %d -C %d; %s &save", fMinAve?"t":"", nLutSize, nCutNum+4, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
        Comm3 = fmt.Sprintf("&synch2 -K %d -C 500; &lf -m%s  -E 5 -K %d -C %d; %s &save", nLutSize, fMinAve?"t":"", nLutSize, nCutNum, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
        Comm4 = fmt.Sprintf("&dch -C 500;          &lf -m%sk -E 5 -K %d -C %d; %s &save", fMinAve?"t":"", nLutSize, nCutNum+4, fUseMfs ? "&put; mfs2 -W 4 -M 500 -C 7000; &get -m;":"");
    }

    // perform synthesis
    if fVerbose != 0 {
        fmt.Printf("Trying synthesis...\n"); }
    if fIsMapped != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" ); }
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm1 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm2 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // return the result
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&load" );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }


    // perform balancing
    if fBalance != 0 {
        if fVerbose != 0 {
            fmt.Printf("Trying SOP balancing...\n"); }
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st; &sopb -R 10 -C 4" );
    }


    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm3 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm2 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // return the result
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&load" );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }
    if nAnds > 100000 {
        return; }


    // perform balancing
    if fBalance != 0 {
        if fVerbose != 0 {
            fmt.Printf("Trying SOP balancing...\n"); }
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st; &sopb -R 10" );
    }


    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm3 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm2 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // return the result
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&load" );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }
    if nAnds > 50000 {
        return; }


    // perform balancing
    if fBalance != 0 {
        if fVerbose != 0 {
            fmt.Printf("Trying SOP balancing...\n"); }
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st; &sopb -R 10" );
    }


    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm3 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // perform synthesis
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&st" );
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm2 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    // return the result
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&load" );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformFlow3(nLutSize int32, nCutNum int32, fBalance int32, fMinAve int32, fUseMfs int32, fUseLutLib int32, fVerbose int32) {
    var Comm1, Comm2, Comm3 [200]byte;
    if fUseLutLib != 0 {
        Comm1 = fmt.Sprintf("&st; &if -C %d;       &save; &st; &syn2; &if -C %d;       &save; &load", nCutNum, nCutNum);
    } else {
        Comm1 = fmt.Sprintf("&st; &if -C %d -K %d; &save; &st; &syn2; &if -C %d -K %d; &save; &load", nCutNum, nLutSize, nCutNum, nLutSize); }
    if fUseLutLib != 0 {
        Comm2 = fmt.Sprintf("&st; &if -%s -K 6; &dch -f; &if -C %d;       %s&save; &load", Abc_NtkRecIsRunning3() ? "y" : "g", nCutNum, fUseMfs ? "&mfs; ":"");
    } else {
        Comm2 = fmt.Sprintf("&st; &if -%s -K 6; &dch -f; &if -C %d -K %d; %s&save; &load", Abc_NtkRecIsRunning3() ? "y" : "g", nCutNum, nLutSize, fUseMfs ? "&mfs; ":""); }
    if fUseLutLib != 0 {
        Comm3 = fmt.Sprintf("&st; &if -%s -K 6; &synch2; &if -C %d;       %s&save; &load", Abc_NtkRecIsRunning3() ? "y" : "g", nCutNum, fUseMfs ? "&mfs; ":"");
    } else {
        Comm3 = fmt.Sprintf("&st; &if -%s -K 6; &synch2; &if -C %d -K %d; %s&save; &load", Abc_NtkRecIsRunning3() ? "y" : "g", nCutNum, nLutSize, fUseMfs ? "&mfs; ":""); }

    if fVerbose != 0 { fmt.Printf("Trying simple synthesis with %s...\n", Abc_NtkRecIsRunning3() ? "LMS" : "SOP balancing"); }
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm1 );
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }

    if Gia_ManAndNum( Abc_FrameReadGia(Abc_FrameGetGlobalFrame()) ) < 200000 {
        if fVerbose != 0 { fmt.Printf("Trying medium synthesis...\n"); }
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm2 );
        if fVerbose != 0 {
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }
    }

    if Gia_ManAndNum( Abc_FrameReadGia(Abc_FrameGetGlobalFrame()) ) < 10000 {
        if fVerbose != 0 { fmt.Printf("Trying harder synthesis...\n"); }
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Comm3 );
        if fVerbose != 0 {
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }
    }

    if fVerbose != 0 { fmt.Printf("Final result...\n"); }
    if fVerbose != 0 {
    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "&ps" ); }
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

