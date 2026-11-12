package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaDeep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Experiments with synthesis.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaDeep.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "base/main/main.h"
//#include "base/cmd/cmd.h"

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
func Gia_ManDeepSynOne(nNoImpr int32, TimeOut int32, nAnds int32, Seed int32, fUseTwo int32, fVerbose int32) *Gia_Man_t {
    var nTimeToStop abctime;
    if TimeOut {
        nTimeToStop = Abc_Clock() + TimeOut * int64(time.Second);
    } else {
        nTimeToStop = 0;
    }
    var clkStart abctime     = Abc_Clock();
    var s int32;
    var i int32;
    var IterMax int32 = 100000;
    var nAndsMin int32 = -1;
    var iIterLast int32 = -1;
    var pTemp *Gia_Man_t    = Abc_FrameReadGia(Abc_FrameGetGlobalFrame());
    var pNew *Gia_Man_t     = Gia_ManDup( pTemp );
    Abc_Random(1);
    for  s = 0; s < 10+Seed; s++  {
        Abc_Random(0);
    }
    for  i = 0; i < IterMax; i++  {
        var Rand uint32  = Abc_Random(0);
        var fDch int32  = Rand & 1;
        //int fCom = (Rand >> 1) & 3;
        var fCom int32  = (Rand >> 1) & 1;
        var fFx int32   = (Rand >> 2) & 1;
        var KLut int32;
        if fUseTwo != 0 {
            KLut = 2 + (i % 5);
        } else {
            KLut = 3 + (i % 4);
        }
        var fChange int32;
        var Command string // was [1000]byte in C

        var pComp string  = nil;
        if fCom == 3 {
            pComp = "; &put; compress2rs; compress2rs; compress2rs; &get";
        } else if fCom == 2 {
            pComp = "; &put; compress2rs; compress2rs; &get";
        } else if fCom == 1 {
            pComp = "; &put; compress2rs; &get";
        } else if fCom == 0 {
            pComp = "; &dc2"; }
        Command = fmt.Sprintf("&dch%s; &if -a -K %d; &mfs -e -W 20 -L 20%s%s", fDch ? " -f" : "", KLut, fFx ? "; &fx" : "", pComp);
        if Abc_FrameIsBatchMode() {
            if Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), Command) {
                Abc_Print( 1, "Something did not work out with the command \"%s\".\n", Command );
                return nil;
            }
        
        } else {
            Abc_FrameSetBatchMode( 1 );
            if Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), Command) {
                Abc_Print( 1, "Something did not work out with the command \"%s\".\n", Command );
                return nil;
            }
            Abc_FrameSetBatchMode( 0 );
        }
        pTemp = Abc_FrameReadGia(Abc_FrameGetGlobalFrame());
        if Gia_ManAndNum(pNew) > Gia_ManAndNum(pTemp) {
            Gia_ManStop( pNew );
            pNew = Gia_ManDup( pTemp );
            fChange = 1;
            iIterLast = i;
        
        } else if Gia_ManAndNum(pNew) + Gia_ManAndNum(pNew)/10 < Gia_ManAndNum(pTemp) {
            //printf( "Updating\n" );
            //Abc_FrameUpdateGia( Abc_FrameGetGlobalFrame(), Gia_ManDup(pNew) );
        }
        if (fChange != 0) && (fVerbose != 0) {
            fmt.Printf("Iter %6d : ", i);
            fmt.Printf("Time %8.2f sec : ", float32(1.0)*(Abc_Clock() - clkStart)/int64(time.Second));
            fmt.Printf("And = %6d  ", Gia_ManAndNum(pNew));
            fmt.Printf("Lev = %3d  ", Gia_ManLevelNum(pNew));
            if fChange != 0 { 
                fmt.Printf("<== best : ");
            } else if fVerbose != 0 {
                fmt.Printf("           "); }
            printf( "%s", Command );
            fmt.Printf("\n");
        }
        if nTimeToStop && Abc_Clock() > nTimeToStop {
            fmt.Printf("Runtime limit (%d sec) is reached after %d iterations.\n", TimeOut, i);
            break;
        }
        if i - iIterLast > nNoImpr {
            fmt.Printf("Completed %d iterations without improvement in %.2f seconds.\n", nNoImpr, float32(1.0)*(Abc_Clock() - clkStart)/int64(time.Second));
            break;
        }
    }
    if i == IterMax {
        fmt.Printf("Iteration limit (%d iters) is reached after %.2f seconds.\n", IterMax, float32(1.0)*(Abc_Clock() - clkStart)/int64(time.Second));
    } else if (nAnds != 0) && nAndsMin <= nAnds {
        fmt.Printf("Quality goal (%d nodes <= %d nodes) is achieved after %d iterations and %.2f seconds.\n", nAndsMin, nAnds, i, float32(1.0)*(Abc_Clock() - clkStart)/int64(time.Second)); }
    return pNew;
}
func Gia_ManDeepSyn(pGia *Gia_Man_t, nIters int32, nNoImpr int32, TimeOut int32, nAnds int32, Seed int32, fUseTwo int32, fVerbose int32) *Gia_Man_t {
    var pInit *Gia_Man_t  = Gia_ManDup(pGia);
    var pBest *Gia_Man_t  = Gia_ManDup(pGia);
    var pThis *Gia_Man_t;
    var i int32;
    for  i = 0; i < nIters; i++  {
        Abc_FrameUpdateGia( Abc_FrameGetGlobalFrame(), Gia_ManDup(pInit) );
        pThis = Gia_ManDeepSynOne( nNoImpr, TimeOut, nAnds, Seed+i, fUseTwo, fVerbose );
        if Gia_ManAndNum(pBest) > Gia_ManAndNum(pThis) {
            Gia_ManStop( pBest );
            pBest = pThis;
        
        } else { 
            Gia_ManStop( pThis ); }
        
    }
    Gia_ManStop( pInit );
    return pBest;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

