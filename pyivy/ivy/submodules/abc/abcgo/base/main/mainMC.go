package main

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [mainMC.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The main package.]

  Synopsis    [The main file for the model checker.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: main.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mainInt.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include "aig/fra/fra.h"
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [The main() procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func main() {
    var fEnableBmcOnly int32;  // enable to make it BMC-only

    var fEnableCounter int32  = 1;  // should be 1 in the final version
    var fEnableComment int32;  // should be 0 in the final version

    var SecPar Fra_Sec_t;
    var pSecPar *Fra_Sec_t = &SecPar;
    var pFile *os.File;
    var pAig *Aig_Man_t;
    var RetValue int32  = -1;
    var Depth int32     = -1;
    // BMC parameters
    var nFrames int32   = 50;
    var nSizeMax int32  = 500000;
    var nBTLimit int32  = 10000;
    var fRewrite int32;
    var fNewAlgo int32  = 1;
    var fVerbose int32;
    var clkTotal int64  = time.Now().UnixNano();

    if len(os.Args) != 2 {
        fmt.Printf("  Expecting one command-line argument (an input file in AIGER format).\n");
        fmt.Printf("  usage: %s <file.aig>\n", os.Args[0]);
        return ;
    }
    pFile = os.Open(os.Args[1]);
    if pFile == nil {
        fmt.Printf("  Cannot open input AIGER file \"%s\".\n", os.Args[1]);
        fmt.Printf("  usage: %s <file.aig>\n", os.Args[0]);
        return ;
    }
    _ = (pFile).Close();
    pAig = Ioa_ReadAiger( os.Args[1], 1 );
    if pAig == nil {
        fmt.Printf("  Parsing the AIGER file \"%s\" has failed.\n", os.Args[1]);
        fmt.Printf("  usage: %s <file.aig>\n", os.Args[0]);
        return ;
    }

    Aig_ManSetRegNum( pAig, pAig.nRegs );
    if !(fEnableBmcOnly != 0) {
        // perform BMC
        if pAig.nRegs != 0 {
            RetValue = Saig_ManBmcSimple( pAig, nFrames, nSizeMax, nBTLimit, fRewrite, fVerbose, nil, 0, 0 ); }

        // perform full-blown SEC
        if RetValue != 0 {
            extern void Dar_LibStart();
            extern void Dar_LibStop();
            extern void Cnf_ManFree();

            Fra_SecSetDefaultParams( pSecPar );
            pSecPar.TimeLimit       = 600;
            pSecPar.nFramesMax      =   4;  // the max number of frames used for induction
            pSecPar.fPhaseAbstract  =   0;  // disable phase-abstraction
            pSecPar.fSilent         =   1;  // disable phase-abstraction

            Dar_LibStart();
            RetValue = Fra_FraigSec( pAig, pSecPar, nil );
            Dar_LibStop();
            Cnf_ManFree();
        }
    }

    // perform BMC again
    if RetValue == -1 && pAig.nRegs != 0 {
        var nFrames int32   = 200;
        var nSizeMax int32  = 500000;
        var nBTLimit int32  = 10000000;
        var fRewrite int32;
        RetValue = Saig_ManBmcSimple( pAig, nFrames, nSizeMax, nBTLimit, fRewrite, fVerbose, &Depth, 0, 0 );
        if RetValue != 0 {
            RetValue = -1; }
    }

    // decide how to report the output
    pFile = os.Stdout;

    // report the result
    if RetValue == 0 {
//        fprintf(stdout, "s SATIFIABLE\n");
        fmt.Fprintf(pFile, "1");
        if fEnableCounter != 0 {
        fmt.Printf("\n");
        if pAig.pSeqModel {
        Fra_SmlWriteCounterExample( pFile, pAig, pAig.pSeqModel ); }
        }

        if fEnableComment != 0 {
        fmt.Printf("  # File %10s.  ", os.Args[1]);
        PRT( "Time", time.Now().UnixNano() - clkTotal );
        }

        if pFile != os.Stdout {
            _ = (pFile).Close(); }
        Aig_ManStop( pAig );
        os.Exit(int(10));
     
    } else if RetValue == 1 {
//    fprintf(stdout, "s UNSATISFIABLE\n");
        fmt.Fprintf(pFile, "0");

        if fEnableComment != 0 {
        fmt.Printf("  # File %10s.  ", os.Args[1]);
        PRT( "Time", time.Now().UnixNano() - clkTotal );
        }
        fmt.Printf("\n");

        if pFile != os.Stdout {
            _ = (pFile).Close(); }
        Aig_ManStop( pAig );
        os.Exit(int(20));
     
    } else { // if ( RetValue == -1 ) 
    
//    fprintf(stdout, "s UNKNOWN\n");
        fmt.Fprintf(pFile, "2");

        if fEnableComment != 0 {
        fmt.Printf("  # File %10s.  ", os.Args[1]);
        PRT( "Time", time.Now().UnixNano() - clkTotal );
        }
        fmt.Printf("\n");

        if pFile != os.Stdout {
            _ = (pFile).Close(); }
        Aig_ManStop( pAig );
        os.Exit(int(0));
    }
    return ;
}




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

