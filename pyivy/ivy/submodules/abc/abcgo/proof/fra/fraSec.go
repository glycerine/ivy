package fra

import "fmt"

/**CFile****************************************************************

  FileName    [fraSec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Performs SEC based on seq sweeping.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraSec.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"
//#include "aig/ioa/ioa.h"
//#include "proof/int/int.h"
//#include "proof/ssw/ssw.h"
//#include "aig/saig/saig.h"
//#include "bdd/bbr/bbr.h"
//#include "proof/pdr/pdr.h"

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
func Fra_SecSetDefaultParams(p *Fra_Sec_t) {
    memset( p, 0, sizeof(Fra_Sec_t) );
    p.fTryComb          =       1;  // try CEC call as a preprocessing step
    p.fTryBmc           =       1;  // try BMC call as a preprocessing step 
    p.nFramesMax        =       4;  // the max number of frames used for induction
    p.nBTLimit          =    1000;  // conflict limit at a node during induction 
    p.nBTLimitGlobal    = 5000000;  // global conflict limit during induction
    p.nBTLimitInter     =   10000;  // conflict limit during interpolation
    p.nBddVarsMax       =     150;  // the limit on the number of registers in BDD reachability
    p.nBddMax           =   50000;  // the limit on the number of BDD nodes 
    p.nBddIterMax       = 1000000;  // the limit on the number of BDD iterations
    p.fPhaseAbstract    =       0;  // enables phase abstraction
    p.fRetimeFirst      =       1;  // enables most-forward retiming at the beginning
    p.fRetimeRegs       =       1;  // enables min-register retiming at the beginning
    p.fFraiging         =       1;  // enables fraiging at the beginning
    p.fInduction        =       1;  // enables the use of induction (signal correspondence)
    p.fInterpolation    =       1;  // enables interpolation
    p.fInterSeparate    =       0;  // enables interpolation for each outputs separately
    p.fReachability     =       1;  // enables BDD based reachability
    p.fReorderImage     =       1;  // enables variable reordering during image computation
    p.fStopOnFirstFail  =       1;  // enables stopping after first output of a miter has failed to prove
    p.fUseNewProver     =       0;  // enables new prover
    p.fUsePdr           =       1;  // enables PDR
    p.nPdrTimeout       =      60;  // enabled PDR timeout
    p.fSilent           =       0;  // disables all output
    p.fVerbose          =       0;  // enables verbose reporting of statistics
    p.fVeryVerbose      =       0;  // enables very verbose reporting  
    p.TimeLimit         =       0;  // enables the timeout
    // internal parameters
    p.fReportSolution   =       0;  // enables specialized format for reporting solution
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Fra_FraigSec_Counter int32 = 1;
func Fra_FraigSec(p *Aig_Man_t, pParSec *Fra_Sec_t, ppResult **Aig_Man_t) int32 {
    var Pars2 Ssw_Pars_t;
    var pPars2 *Ssw_Pars_t = &Pars2;
    var Pars Fra_Ssw_t;
    var pPars *Fra_Ssw_t = &Pars;
    var pSml *Fra_Sml_t;
    var pNew, pTemp *Aig_Man_t;
    var nFrames, RetValue, nIter int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    var TimeOut int32;
    var fLatchCorr int32;
    var TimeLeft float32  = 0.0;
    pParSec.nSMnumber = -1;

    // try the miter before solving
    pNew = Aig_ManDupSimple( p );
    RetValue = Fra_FraigMiterStatus( pNew );
    if RetValue >= 0 {
        goto finish; }

    // prepare parameters
    memset( pPars, 0, sizeof(Fra_Ssw_t) );
    pPars.fLatchCorr  = fLatchCorr;
    pPars.fVerbose = pParSec.fVeryVerbose;
    if pParSec.fVerbose {
        fmt.Printf("Original miter:       Latches = %5d. Nodes = %6d.\n", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
    }
//Aig_ManDumpBlif( pNew, "after.blif", NULL, NULL );

    // perform sequential cleanup
clk = Abc_Clock();
    if pNew.nRegs {
    pNew = Aig_ManReduceLaches( pNew, 0 ); }
    if pNew.nRegs {
    pNew = Aig_ManConstReduce( pNew, 0, -1, -1, 0, 0 ); }
    if pParSec.fVerbose {
        fmt.Printf("Sequential cleanup:   Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
    }
    RetValue = Fra_FraigMiterStatus( pNew );
    if RetValue >= 0 {
        goto finish; }

    // perform phase abstraction
clk = Abc_Clock();
    if pParSec.fPhaseAbstract {
        extern Aig_Man_t * Saig_ManPhaseAbstractAuto( Aig_Man_t * p, int fVerbose );
        pNew.nTruePis = Aig_ManCiNum(pNew) - Aig_ManRegNum(pNew); 
        pNew.nTruePos = Aig_ManCoNum(pNew) - Aig_ManRegNum(pNew); 
        pNew = Saig_ManPhaseAbstractAuto( pTemp = pNew, 0 );
        Aig_ManStop( pTemp );
        if pParSec.fVerbose {
            fmt.Printf("Phase abstraction:    Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
        }
    }

    // perform forward retiming
    if pParSec.fRetimeFirst && pNew.nRegs {
clk = Abc_Clock();
//    pNew = Rtm_ManRetime( pTemp = pNew, 1, 1000, 0 );
    pNew = Saig_ManRetimeForward( pTemp = pNew, 100, 0 );
    Aig_ManStop( pTemp );
    if pParSec.fVerbose {
        fmt.Printf("Forward retiming:     Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
    }
    }
     
    // run latch correspondence
clk = Abc_Clock();
    if pNew.nRegs {
    pNew = Aig_ManDupOrdered( pTemp = pNew );
//    pNew = Aig_ManDupDfs( pTemp = pNew );
    Aig_ManStop( pTemp );
/*
    if ( RetValue == -1 && pParSec->TimeLimit )
    {
        TimeLeft = (float)pParSec->TimeLimit - ((float)(Abc_Clock()-clkTotal)/(float)(CLOCKS_PER_SEC));
        TimeLeft = Abc_MaxInt( TimeLeft, 0.0 );
        if ( TimeLeft == 0.0 )
        {
            if ( !pParSec->fSilent )
                printf( "Runtime limit exceeded.\n" );
            RetValue = -1;
            TimeOut = 1;
            goto finish;
        }
    }
*/

//    pNew = Fra_FraigLatchCorrespondence( pTemp = pNew, 0, 1000, 1, pParSec->fVeryVerbose, &nIter, TimeLeft );
//Aig_ManDumpBlif( pNew, "ex.blif", NULL, NULL );
    Ssw_ManSetDefaultParamsLcorr( pPars2 );
    pNew = Ssw_LatchCorrespondence( pTemp = pNew, pPars2 );
    nIter = pPars2.nIters;

    // prepare parameters for scorr
    Ssw_ManSetDefaultParams( pPars2 );

    if pTemp.pSeqModel {
        if !Saig_ManVerifyCex( pTemp, pTemp.pSeqModel ) {
            fmt.Printf("Fra_FraigSec(): Counter-example verification has FAILED.\n"); }
        if Saig_ManPiNum(p) != Saig_ManPiNum(pTemp) {
            fmt.Printf("The counter-example is invalid because of phase abstraction.\n");
        } else {
        ABC_FREE( p.pSeqModel );
        p.pSeqModel = Abc_CexDup( pTemp.pSeqModel, Aig_ManRegNum(p) );
        ABC_FREE( pTemp.pSeqModel );
        }
    }
    if pNew == nil {
        if p.pSeqModel {
            RetValue = 0;
            if !pParSec.fSilent {
                fmt.Printf("Networks are NOT EQUIVALENT after simulation.   ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
            }
            if pParSec.fReportSolution && !pParSec.fRecursive {
            fmt.Printf("SOLUTION: FAIL       ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
            }
            Aig_ManStop( pTemp );
            return RetValue;
        }
        pNew = pTemp;
        RetValue = -1;
        TimeOut = 1;
        goto finish;
    }
    Aig_ManStop( pTemp );

    if pParSec.fVerbose {
        fmt.Printf("Latch-corr (I=%3d):   Latches = %5d. Nodes = %6d. ", nIter, Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
    }
    }
/*
    if ( RetValue == -1 && pParSec->TimeLimit )
    {
        TimeLeft = (float)pParSec->TimeLimit - ((float)(Abc_Clock()-clkTotal)/(float)(CLOCKS_PER_SEC));
        TimeLeft = Abc_MaxInt( TimeLeft, 0.0 );
        if ( TimeLeft == 0.0 )
        {
            if ( !pParSec->fSilent )
                printf( "Runtime limit exceeded.\n" );
            RetValue = -1;
            TimeOut = 1;
            goto finish;
        }
    }
*/
    // perform fraiging
    if pParSec.fFraiging {
clk = Abc_Clock();
    pNew = Fra_FraigEquivence( pTemp = pNew, 100, 0 );
    Aig_ManStop( pTemp );
    if pParSec.fVerbose {
        fmt.Printf("Fraiging:             Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
    }
    }

    if pNew.nRegs == 0 {
        RetValue = Fra_FraigCec( &pNew, 100000, 0 ); }

    RetValue = Fra_FraigMiterStatus( pNew );
    if RetValue >= 0 {
        goto finish; }
/*
    if ( RetValue == -1 && pParSec->TimeLimit )
    {
        TimeLeft = (float)pParSec->TimeLimit - ((float)(Abc_Clock()-clkTotal)/(float)(CLOCKS_PER_SEC));
        TimeLeft = Abc_MaxInt( TimeLeft, 0.0 );
        if ( TimeLeft == 0.0 )
        {
            if ( !pParSec->fSilent )
                printf( "Runtime limit exceeded.\n" );
            RetValue = -1;
            TimeOut = 1;
            goto finish;
        }
    }
*/
    // perform min-area retiming
    if pParSec.fRetimeRegs && pNew.nRegs {
//    extern Aig_Man_t * Saig_ManRetimeMinArea( Aig_Man_t * p, int nMaxIters, int fForwardOnly, int fBackwardOnly, int fInitial, int fVerbose );
clk = Abc_Clock();
    pNew.nTruePis = Aig_ManCiNum(pNew) - Aig_ManRegNum(pNew); 
    pNew.nTruePos = Aig_ManCoNum(pNew) - Aig_ManRegNum(pNew); 
//        pNew = Rtm_ManRetime( pTemp = pNew, 1, 1000, 0 );
    pNew = Saig_ManRetimeMinArea( pTemp = pNew, 1000, 0, 0, 1, 0 );
    Aig_ManStop( pTemp );
    pNew = Aig_ManDupOrdered( pTemp = pNew );
    Aig_ManStop( pTemp );
    if pParSec.fVerbose {
    fmt.Printf("Min-reg retiming:     Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
    }
    }

    // perform seq sweeping while increasing the number of frames
    RetValue = Fra_FraigMiterStatus( pNew );
    if RetValue == -1 && pParSec.fInduction {
    for  nFrames = 1; nFrames <= pParSec.nFramesMax; nFrames *= 2  {
/*
        if ( RetValue == -1 && pParSec->TimeLimit )
        {
            TimeLeft = (float)pParSec->TimeLimit - ((float)(Abc_Clock()-clkTotal)/(float)(CLOCKS_PER_SEC));
            TimeLeft = Abc_MaxInt( TimeLeft, 0.0 );
            if ( TimeLeft == 0.0 )
            {
                if ( !pParSec->fSilent )
                    printf( "Runtime limit exceeded.\n" );
                RetValue = -1;
                TimeOut = 1;
                goto finish;
            }
        }
*/ 

clk = Abc_Clock();
        pPars.nFramesK = nFrames;
        pPars.TimeLimit = TimeLeft;
        pPars.fSilent = pParSec.fSilent;
//        pNew = Fra_FraigInduction( pTemp = pNew, pPars );

        pPars2.nFramesK = nFrames;
        pPars2.nBTLimit = pParSec.nBTLimit;
        pPars2.nBTLimitGlobal = pParSec.nBTLimitGlobal;
//        pPars2->nBTLimit = 1000 * nFrames;

        if RetValue == -1 && pPars2.nConflicts > pPars2.nBTLimitGlobal {
            if !pParSec.fSilent {
                fmt.Printf("Global conflict limit (%d) exceeded.\n", pPars2.nBTLimitGlobal); }
            RetValue = -1;
            TimeOut = 1;
            goto finish;
        }

        Aig_ManSetRegNum( pNew, pNew.nRegs );
//        pNew = Ssw_SignalCorrespondence( pTemp = pNew, pPars2 );
        if Aig_ManRegNum(pNew) > 0 {
            pNew = Ssw_SignalCorrespondence( pTemp = pNew, pPars2 );
        } else {
            pNew = Aig_ManDupSimpleDfs( pTemp = pNew ); }

        if pNew == nil {
            pNew = pTemp;
            RetValue = -1;
            TimeOut = 1;
            goto finish;
        }

//        printf( "Total conflicts = %d.\n", pPars2->nConflicts );

        Aig_ManStop( pTemp );
        RetValue = Fra_FraigMiterStatus( pNew );
        if pParSec.fVerbose { 
            fmt.Printf("K-step (K=%2d,I=%3d):  Latches = %5d. Nodes = %6d. ", nFrames, pPars2.nIters, Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
        }
        if RetValue != -1 {
            break; }

        // perform retiming
//        if ( pParSec->fRetimeFirst && pNew->nRegs )
        if pNew.nRegs {
//        extern Aig_Man_t * Saig_ManRetimeMinArea( Aig_Man_t * p, int nMaxIters, int fForwardOnly, int fBackwardOnly, int fInitial, int fVerbose );
clk = Abc_Clock();
        pNew.nTruePis = Aig_ManCiNum(pNew) - Aig_ManRegNum(pNew); 
        pNew.nTruePos = Aig_ManCoNum(pNew) - Aig_ManRegNum(pNew); 
//        pNew = Rtm_ManRetime( pTemp = pNew, 1, 1000, 0 );
        pNew = Saig_ManRetimeMinArea( pTemp = pNew, 1000, 0, 0, 1, 0 );
        Aig_ManStop( pTemp );
        pNew = Aig_ManDupOrdered( pTemp = pNew );
        Aig_ManStop( pTemp );
        if pParSec.fVerbose {
            fmt.Printf("Min-reg retiming:     Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
        }
        }  

        if pNew.nRegs {
            pNew = Aig_ManConstReduce( pNew, 0, -1, -1, 0, 0 ); }

        // perform rewriting
clk = Abc_Clock();
        pNew = Aig_ManDupOrdered( pTemp = pNew );
        Aig_ManStop( pTemp );
//        pNew = Dar_ManRewriteDefault( pTemp = pNew );
        pNew = Dar_ManCompress2( pTemp = pNew, 1, 0, 1, 0, 0 ); 
        Aig_ManStop( pTemp );
        if pParSec.fVerbose {
            fmt.Printf("Rewriting:            Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
        } 

        // perform sequential simulation
        if pNew.nRegs {
clk = Abc_Clock();
        pSml = Fra_SmlSimulateSeq( pNew, 0, 128 * nFrames, 1 + 16/(1+Aig_ManNodeNum(pNew)/1000), 1  ); 
        if pParSec.fVerbose {
            fmt.Printf("Seq simulation  :     Latches = %5d. Nodes = %6d. ", Aig_ManRegNum(pNew), Aig_ManNodeNum(pNew));
ABC_PRT( "Time", Abc_Clock() - clk );
        }
        if pSml.fNonConstOut {
            pNew.pSeqModel = Fra_SmlGetCounterExample( pSml );
            // transfer to the original manager
            if Saig_ManPiNum(p) != Saig_ManPiNum(pNew) {
                fmt.Printf("The counter-example is invalid because of phase abstraction.\n");
            } else {
            ABC_FREE( p.pSeqModel );
            p.pSeqModel = Abc_CexDup( pNew.pSeqModel, Aig_ManRegNum(p) );
            ABC_FREE( pNew.pSeqModel );
            }

            Fra_SmlStop( pSml );
            Aig_ManStop( pNew );
            RetValue = 0;
            if !pParSec.fSilent {
                fmt.Printf("Networks are NOT EQUIVALENT after simulation.   ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
            }
            if pParSec.fReportSolution && !pParSec.fRecursive {
            fmt.Printf("SOLUTION: FAIL       ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
            }
            return RetValue;
        }
        Fra_SmlStop( pSml );
        }
    } }

    // get the miter status
    RetValue = Fra_FraigMiterStatus( pNew );

    // try interplation
clk = Abc_Clock();
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(pNew) );
    if pParSec.fInterpolation && RetValue == -1 && Aig_ManRegNum(pNew) > 0 {
        var Pars Inter_ManParams_t;
        var pPars *Inter_ManParams_t = &Pars;
        var Depth int32;
        ABC_FREE( pNew.pSeqModel );
        Inter_ManSetDefaultParams( pPars );
//        pPars->nBTLimit = 100;
        pPars.nBTLimit = pParSec.nBTLimitInter;
        pPars.fVerbose = pParSec.fVeryVerbose;
        if Saig_ManPoNum(pNew) == 1 {
            RetValue = Inter_ManPerformInterpolation( pNew, pPars, &Depth );
        
        } else if pParSec.fInterSeparate {
            var pCex *Abc_Cex_t  = nil;
            var pTemp, pAux *Aig_Man_t;
            var pObjPo *Aig_Obj_t;
            var i, spStatic_Fra_FraigSec_Counter int32;
            Saig_ManForEachPo( pNew, pObjPo, i )
            { 
                if Aig_ObjFanin0(pObjPo) == Aig_ManConst1(pNew) {
                    continue; }
                if pPars.fVerbose {
                    fmt.Printf("Solving output %2d (out of %2d):\n", i, Saig_ManPoNum(pNew)); }
                pTemp = Aig_ManDupOneOutput( pNew, i, 1 );
                pTemp = Aig_ManScl( pAux = pTemp, 1, 1, 0, -1, -1, 0, 0 );
                Aig_ManStop( pAux );
                if Saig_ManRegNum(pTemp) > 0 {
                    RetValue = Inter_ManPerformInterpolation( pTemp, pPars, &Depth );
                    if pTemp.pSeqModel {
                        pCex = p.pSeqModel = Abc_CexDup( pTemp.pSeqModel, Aig_ManRegNum(p) );
                        pCex.iPo = i;
                        Aig_ManStop( pTemp );
                        break;
                    }
                    // if solved, remove the output
                    if RetValue == 1 {
                        Aig_ObjPatchFanin0( pNew, pObjPo, Aig_ManConst0(pNew) );
    //                    printf( "Output %3d : Solved ", i );
                    
                    } else {
                        spStatic_Fra_FraigSec_Counter++;
    //                    printf( "Output %3d : Undec  ", i );
                    }
                
                } else {
                    spStatic_Fra_FraigSec_Counter++; }
//                Aig_ManPrintStats( pTemp );
                Aig_ManStop( pTemp );
                fmt.Printf("Solving output %3d (out of %3d) using interpolation.\r", i, Saig_ManPoNum(pNew));
            }
            Aig_ManCleanup( pNew );
            if pCex == nil {
                fmt.Printf("Interpolation left %d (out of %d) outputs unsolved              \n", spStatic_Fra_FraigSec_Counter, Saig_ManPoNum(pNew));
                if spStatic_Fra_FraigSec_Counter != 0 {
                    RetValue = -1; }
            }
            pNew = Aig_ManDupUnsolvedOutputs( pTemp = pNew, 1 );
            Aig_ManStop( pTemp );
            pNew = Aig_ManScl( pTemp = pNew, 1, 1, 0, -1, -1, 0, 0 );
            Aig_ManStop( pTemp );
        
        } else {
            var pNewOrpos *Aig_Man_t  = Saig_ManDupOrpos( pNew );
            RetValue = Inter_ManPerformInterpolation( pNewOrpos, pPars, &Depth );
            if pNewOrpos.pSeqModel {
                var pCex *Abc_Cex_t;
                pCex = pNew.pSeqModel = pNewOrpos.pSeqModel; pNewOrpos.pSeqModel = nil;
                pCex.iPo = Saig_ManFindFailedPoCex( pNew, pNew.pSeqModel );
            }
            Aig_ManStop( pNewOrpos );
        }

        if pParSec.fVerbose {
        if RetValue == 1 {
            fmt.Printf("Property proved using interpolation.  ");
        } else if RetValue == 0 {
            fmt.Printf("Property DISPROVED in frame %d using interpolation.  ", Depth);
        } else if RetValue == -1 {
            fmt.Printf("Property UNDECIDED after interpolation.  ");
        } else {
            assert(false, "0"); } 
ABC_PRT( "Time", Abc_Clock() - clk );
        }
    }

    // try reachability analysis
    if pParSec.fReachability && RetValue == -1 && Aig_ManRegNum(pNew) > 0 && Aig_ManRegNum(pNew) < pParSec.nBddVarsMax {
        var Pars Saig_ParBbr_t;
        var pPars *Saig_ParBbr_t = &Pars;
        Bbr_ManSetDefaultParams( pPars );
        pPars.TimeLimit     = 0;
        pPars.nBddMax       = pParSec.nBddMax;
        pPars.nIterMax      = pParSec.nBddIterMax;
        pPars.fPartition    = 1;
        pPars.fReorder      = 1;
        pPars.fReorderImage = 1;
        pPars.fVerbose      = 0;
        pPars.fSilent       = pParSec.fSilent;
        pNew.nTruePis = Aig_ManCiNum(pNew) - Aig_ManRegNum(pNew); 
        pNew.nTruePos = Aig_ManCoNum(pNew) - Aig_ManRegNum(pNew); 
        RetValue = Aig_ManVerifyUsingBdds( pNew, pPars );
    }

    // try PDR
    if pParSec.fUsePdr && RetValue == -1 && Aig_ManRegNum(pNew) > 0 {
        var Pars Pdr_Par_t;
        var pPars *Pdr_Par_t = &Pars;
        Pdr_ManSetDefaultParams( pPars );
        pPars.nTimeOut = pParSec.nPdrTimeout;
        pPars.fVerbose = pParSec.fVerbose;
        if pParSec.fVerbose {
            fmt.Printf("Running property directed reachability...\n"); }
        RetValue = Pdr_ManSolve( pNew, pPars );
        if pNew.pSeqModel {
            pNew.pSeqModel.iPo = Saig_ManFindFailedPoCex( pNew, pNew.pSeqModel ); }
    }

finish:
    // report the miter
    if RetValue == 1 {
        if !pParSec.fSilent {
            fmt.Printf("Networks are equivalent.  ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
        }
        if pParSec.fReportSolution && !pParSec.fRecursive {
        fmt.Printf("SOLUTION: PASS       ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
        }
    
    } else if RetValue == 0 {
        if pNew.pSeqModel == nil {
            var i int32;
            // if the CEX is not derives, it is because tricial CEX should be assumed
            pNew.pSeqModel = Abc_CexAlloc( Aig_ManRegNum(pNew), pNew.nTruePis, 1 );
            // if the CEX does not work, we need to change PIs to 1 because 
            // the only way it can happen is when a PO is equal to a PI...
            if Saig_ManFindFailedPoCex( pNew, pNew.pSeqModel ) == -1 {
                for  i = 0; i < pNew.nTruePis; i++  {
                    Abc_InfoSetBit( pNew.pSeqModel.pData, i );
                } }
        }
        if !pParSec.fSilent {
            fmt.Printf("Networks are NOT EQUIVALENT.  ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
        }
        if pParSec.fReportSolution && !pParSec.fRecursive {
        fmt.Printf("SOLUTION: FAIL       ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
        }
    
    } else {
        ///////////////////////////////////
        // save intermediate result
        extern void Abc_FrameSetSave1( void * pAig );
        Abc_FrameSetSave1( Aig_ManDupSimple(pNew) );
        ///////////////////////////////////
        if !pParSec.fSilent {
            fmt.Printf("Networks are UNDECIDED.   ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
        }
        if pParSec.fReportSolution && !pParSec.fRecursive {
        fmt.Printf("SOLUTION: UNDECIDED  ");
ABC_PRT( "Time", Abc_Clock() - clkTotal );
        }
        if !(TimeOut != 0) && !pParSec.fSilent {
            
            var pFileName string // was [1000]byte in C

            pParSec.nSMnumber = spStatic_Fra_FraigSec_Counter;
            pFileName = fmt.Sprintf("sm%02d.aig", spStatic_Fra_FraigSec_Counter++);
            Ioa_WriteAiger( pNew, pFileName, 0, 0 );
            printf( "The unsolved reduced miter is written into file \"%s\".\n", pFileName );
        }
    }
    if pNew.pSeqModel {
        if Saig_ManPiNum(p) != Saig_ManPiNum(pNew) {
            fmt.Printf("The counter-example is invalid because of phase abstraction.\n");
        } else {
        ABC_FREE( p.pSeqModel );
        p.pSeqModel = Abc_CexDup( pNew.pSeqModel, Aig_ManRegNum(p) );
        ABC_FREE( pNew.pSeqModel );
        }
    }
    if ppResult != nil {
        *ppResult = Aig_ManDupSimpleDfs( pNew ); }
    if pNew {
        Aig_ManStop( pNew ); }
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

