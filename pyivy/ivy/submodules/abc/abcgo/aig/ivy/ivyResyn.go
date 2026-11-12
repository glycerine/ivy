package ivy

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [ivyResyn.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [AIG rewriting script.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyResyn.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs several passes of rewriting on the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManResyn0(pMan *Ivy_Man_t, fUpdateLevel int32, fVerbose int32) *Ivy_Man_t {
    var clk abctime;
    var pTemp *Ivy_Man_t;

if fVerbose != 0 { fmt.Printf("Original:\n"); }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pMan, fUpdateLevel );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Balance"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

//    Ivy_ManRewriteAlg( pMan, fUpdateLevel, 0 );
clk = Abc_Clock();
    Ivy_ManRewritePre( pMan, fUpdateLevel, 0, 0 );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Rewrite"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pTemp = pMan, fUpdateLevel );
    Ivy_ManStop( pTemp );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Balance"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Performs several passes of rewriting on the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManResyn(pMan *Ivy_Man_t, fUpdateLevel int32, fVerbose int32) *Ivy_Man_t {
    var clk abctime;
    var pTemp *Ivy_Man_t;

if fVerbose != 0 { fmt.Printf("Original:\n"); }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pMan, fUpdateLevel );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Balance"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

//    Ivy_ManRewriteAlg( pMan, fUpdateLevel, 0 );
clk = Abc_Clock();
    Ivy_ManRewritePre( pMan, fUpdateLevel, 0, 0 );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Rewrite"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pTemp = pMan, fUpdateLevel );
    Ivy_ManStop( pTemp );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Balance"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

//    Ivy_ManRewriteAlg( pMan, fUpdateLevel, 1 );
clk = Abc_Clock();
    Ivy_ManRewritePre( pMan, fUpdateLevel, 1, 0 );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Rewrite"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pTemp = pMan, fUpdateLevel );
    Ivy_ManStop( pTemp );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Balance"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

//    Ivy_ManRewriteAlg( pMan, fUpdateLevel, 1 );
clk = Abc_Clock();
    Ivy_ManRewritePre( pMan, fUpdateLevel, 1, 0 );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Rewrite"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pTemp = pMan, fUpdateLevel );
    Ivy_ManStop( pTemp );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Balance"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Performs several passes of rewriting on the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManRwsat(pMan *Ivy_Man_t, fVerbose int32) *Ivy_Man_t {
    var clk abctime;
    var pTemp *Ivy_Man_t;

if fVerbose != 0 { fmt.Printf("Original:\n"); }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    Ivy_ManRewritePre( pMan, 0, 0, 0 );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Rewrite"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pTemp = pMan, 0 );
//    pMan = Ivy_ManDup( pTemp = pMan );
    Ivy_ManStop( pTemp );
if fVerbose != 0 { fmt.Printf("\n"); }
// inlining of macro ABC_PRT
if fVerbose != 0 { Abc_Print(1, "%s =", ("Balance"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
if fVerbose != 0 { Ivy_ManPrintStats( pMan ); }

/*
clk = Abc_Clock();
    Ivy_ManRewritePre( pMan, 0, 0, 0 );
if ( fVerbose ) { printf( "\n" ); }
if ( fVerbose ) { ABC_PRT( "Rewrite", Abc_Clock() - clk ); }
if ( fVerbose ) Ivy_ManPrintStats( pMan );

clk = Abc_Clock();
    pMan = Ivy_ManBalance( pTemp = pMan, 0 );
    Ivy_ManStop( pTemp );
if ( fVerbose ) { printf( "\n" ); }
if ( fVerbose ) { ABC_PRT( "Balance", Abc_Clock() - clk ); }
if ( fVerbose ) Ivy_ManPrintStats( pMan );
*/
    return pMan;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

