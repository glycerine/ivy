package exor

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [exorUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Exclusive sum-of-product minimization.]

  Synopsis    [Utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: exorUtil.c,v 1.0 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                                                                  ///
///                  Implementation of EXORCISM - 4                  ///
///              An Exclusive Sum-of-Product Minimizer               ///
///               Alan Mishchenko  <alanmi@ee.pdx.edu>               ///
///                                                                  ///
////////////////////////////////////////////////////////////////////////
///                                                                  ///
///                        Utility Functions                         ///
///                                                                  ///
///       1) allocating memory for and creating the ESOP cover       ///
///       2) writing the resultant cover into an ESOP PLA file       ///
///                                                                  ///
///  Ver. 1.0. Started - July 15, 2000. Last update - July 20, 2000  ///
///  Ver. 1.4. Started -  Aug 10, 2000. Last update -  Aug 10, 2000  ///
///  Ver. 1.5. Started -  Aug 19, 2000. Last update -  Aug 19, 2000  ///
///  Ver. 1.7. Started -  Sep 20, 2000. Last update -  Sep 23, 2000  ///
///                                                                  ///
////////////////////////////////////////////////////////////////////////
///   This software was tested with the BDD package "CUDD", v.2.3.0  ///
///                          by Fabio Somenzi                        ///
///                  http://vlsi.colorado.edu/~fabio/                ///
////////////////////////////////////////////////////////////////////////

//#include "exor.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                      EXTERNAL VARIABLES                         ////
////////////////////////////////////////////////////////////////////////

// information about the options, the function, and the cover
var g_CoverInfo cinfo;

////////////////////////////////////////////////////////////////////////
///                        EXTERNAL FUNCTIONS                        ///
////////////////////////////////////////////////////////////////////////

// Cube Cover Iterator
// starts an iterator that traverses all the cubes in the ring
// extern Cube* IterCubeSetStart();

// returns the next cube in the ring
// extern Cube* IterCubeSetNext();


// retrieves the variable from the cube
// extern varvalue GetVar( Cube* pC, int Var );


////////////////////////////////////////////////////////////////////////
///                      FUNCTION DECLARATIONS                       ///
////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////
////////////       Cover Service Procedures       /////////////////
///////////////////////////////////////////////////////////////////

func CountLiterals() int32 {
    var p *Cube;
    var LitCounter int32;
    for  p = IterCubeSetStart( ); (p != nil); p = IterCubeSetNext()  {
        LitCounter += p.a;
    }
    return LitCounter;
}

func CountLiteralsCheck() int32 {
    var p *Cube;
    var Value, v int32;
    var LitCounter int32;
    var LitCounterControl int32;

    for  p = IterCubeSetStart( ); (p != nil); p = IterCubeSetNext()  {
        LitCounterControl += p.a;

        assert(p.fMark == 0, "p.fMark == 0");

        // write the input variables
        for  v = 0; v < g_CoverInfo.nVarsIn; v++  {
            Value = GetVar( p, v );
            if Value == VAR_NEG {
                LitCounter++;
            } else if Value == VAR_POS {
                LitCounter++;
            } else if Value != VAR_ABS {
                assert(false, "0");
            }
        }
    }

    if LitCounterControl != LitCounter {
        fmt.Printf("Warning! The recorded number of literals (%d) differs from the actual number (%d)\n", LitCounterControl, LitCounter); }
    return LitCounter;
}

func CountQCost() int32 {
    var p *Cube;
    var QCost int32;
    var QCostControl int32;
    for  p = IterCubeSetStart( ); (p != nil); p = IterCubeSetNext()  {
        QCostControl += p.q;
        QCost += ComputeQCostBits( p );
    }
//    if ( QCostControl != QCost )
//        printf( "Warning! The recorded number of literals (%d) differs from the actual number (%d)\n", QCostControl, QCost );
    return QCost;
}


void WriteTableIntoFile( FILE * pFile )
// nCubesAlloc is the number of allocated cubes 
{
    var v, w int32;
    var p *Cube;
    var cOutputs int32;
    var nOutput int32;
    var WordSize int32;

    for  p = IterCubeSetStart( ); (p != nil); p = IterCubeSetNext()  {
        assert(p.fMark == 0, "p.fMark == 0");

        // write the input variables
        for  v = 0; v < g_CoverInfo.nVarsIn; v++  {
            var Value int32  = GetVar( p, v );
            if Value == VAR_NEG {
                fmt.Fprintf(pFile, "0");
            } else if Value == VAR_POS {
                fmt.Fprintf(pFile, "1");
            } else if Value == VAR_ABS {
                fmt.Fprintf(pFile, "-");
            } else {
                assert(false, "0"); }
        }
        fmt.Fprintf(pFile, " ");

        // write the output variables
        cOutputs = 0;
        nOutput = g_CoverInfo.nVarsOut;
        WordSize = 8*sizeof( unsigned );
        for  w = 0; w < g_CoverInfo.nWordsOut; w++  {
            for  v = 0; v < WordSize; v++  {
                if p.pCubeDataOut[w] & (1<<v) {
                    fmt.Fprintf(pFile, "1");
                } else {
                    fmt.Fprintf(pFile, "0"); }
                if cOutputs++; cOutputs == nOutput {
                    break; }
            }
        }
        fmt.Fprintf(pFile, "\n");
    }
}


int WriteResultIntoFile( char * pFileName )
// write the ESOP cover into the PLA file <NewFileName>
{
    var pFile *os.File;
    var ltime int64;
    var TimeStr string;

    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(pFile, "\n\nCannot open the output file\n");
        return 1;
    }

    // get current time
    time( &ltime );
    TimeStr = asctime( localtime( &ltime ) );
    // get the number of literals
    g_CoverInfo.nLiteralsAfter = CountLiteralsCheck();
    g_CoverInfo.QCostAfter = CountQCost();
    fmt.Fprintf(pFile, "# EXORCISM-4 output for command line arguments: ");
    fmt.Fprintf(pFile, "\"-Q %d -V %d\"\n", g_CoverInfo.Quality, g_CoverInfo.Verbosity);
    fmt.Fprintf(pFile, "# Minimization performed %s", TimeStr);
    fmt.Fprintf(pFile, "# Initial statistics: ");
    fmt.Fprintf(pFile, "Cubes = %d  Literals = %d  QCost = %d\n", g_CoverInfo.nCubesBefore, g_CoverInfo.nLiteralsBefore, g_CoverInfo.QCostBefore);
    fmt.Fprintf(pFile, "# Final   statistics: ");
    fmt.Fprintf(pFile, "Cubes = %d  Literals = %d  QCost = %d\n", g_CoverInfo.nCubesInUse, g_CoverInfo.nLiteralsAfter, g_CoverInfo.QCostAfter);
    // inlining of macro TICKS_TO_SECONDS
    fmt.Fprintf(pFile, "# File reading and reordering time = %.2f sec\n", (float32((g_CoverInfo.TimeRead))/float32((int64(time.Second)))));
    // inlining of macro TICKS_TO_SECONDS
    fmt.Fprintf(pFile, "# Starting cover generation time   = %.2f sec\n", (float32((g_CoverInfo.TimeStart))/float32((int64(time.Second)))));
    // inlining of macro TICKS_TO_SECONDS
    fmt.Fprintf(pFile, "# Pure ESOP minimization time      = %.2f sec\n", (float32((g_CoverInfo.TimeMin))/float32((int64(time.Second)))));
    fmt.Fprintf(pFile, ".i %d\n", g_CoverInfo.nVarsIn);
    fmt.Fprintf(pFile, ".o %d\n", g_CoverInfo.nVarsOut);
    fmt.Fprintf(pFile, ".p %d\n", g_CoverInfo.nCubesInUse);
    fmt.Fprintf(pFile, ".type esop\n");
    WriteTableIntoFile( pFile );
    fmt.Fprintf(pFile, ".e\n");
    _ = (pFile).Close();
    return 0;
}

///////////////////////////////////////////////////////////////////
////////////              End of File             /////////////////
///////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

