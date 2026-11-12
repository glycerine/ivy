package main

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [mainFrame.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The main package.]

  Synopsis    [The global framework resides in this file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mainFrame.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "mainInt.h"
//#include "bool/dec/dec.h"
//#include "map/if/if.h"
//#include "aig/miniaig/ndr.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var s_GlobalFrame *Abc_Frame_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [APIs to access parameters in the flobal frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadStore() *Vec_Ptr_t { return s_GlobalFrame.vStore;       } 
func Abc_FrameReadStoreSize() int32 { return Vec_PtrSize(s_GlobalFrame.vStore); }
func Abc_FrameReadLibLut() any { return s_GlobalFrame.pLibLut;      } 
func Abc_FrameReadLibBox() any { return s_GlobalFrame.pLibBox;      } 
func Abc_FrameReadLibGen() any { return s_GlobalFrame.pLibGen;      } 
func Abc_FrameReadLibGen2() any { return s_GlobalFrame.pLibGen2;     } 
func Abc_FrameReadLibSuper() any { return s_GlobalFrame.pLibSuper;    } 
func Abc_FrameReadLibScl() any { return s_GlobalFrame.pLibScl;      } 
#ifdef ABC_USE_CUDD
void *      Abc_FrameReadManDd()                             { if ( s_GlobalFrame->dd == NULL )      s_GlobalFrame->dd = Cudd_Init( 0, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );  return s_GlobalFrame->dd;      } 
#endif
func Abc_FrameReadManDec() any { if s_GlobalFrame.pManDec == nil { s_GlobalFrame.pManDec = Dec_ManStart(); }
                                        return s_GlobalFrame.pManDec; } 
func Abc_FrameReadManDsd() any { return s_GlobalFrame.pManDsd;      } 
func Abc_FrameReadManDsd2() any { return s_GlobalFrame.pManDsd2;     }
func Abc_FrameReadFlag(pFlag string) string { return Cmd_FlagReadByName( s_GlobalFrame, pFlag );   }
func Abc_FrameReadSignalNames() *Vec_Ptr_t { return s_GlobalFrame.vSignalNames; }
func Abc_FrameReadSpecName() string { return s_GlobalFrame.pSpecName;    }

func Abc_FrameReadBmcFrames(p *Abc_Frame_t) int32 { return s_GlobalFrame.nFrames;      }               
func Abc_FrameReadProbStatus(p *Abc_Frame_t) int32 { return s_GlobalFrame.Status;       }               
func Abc_FrameReadCex(p *Abc_Frame_t) any { return s_GlobalFrame.pCex;         }        
func Abc_FrameReadCexVec(p *Abc_Frame_t) *Vec_Ptr_t { return s_GlobalFrame.vCexVec;      }        
func Abc_FrameReadStatusVec(p *Abc_Frame_t) *Vec_Int_t { return s_GlobalFrame.vStatuses;    }        
func Abc_FrameReadPoEquivs(p *Abc_Frame_t) *Vec_Ptr_t { return s_GlobalFrame.vPoEquivs;    }        
func Abc_FrameReadPoStatuses(p *Abc_Frame_t) *Vec_Int_t { return s_GlobalFrame.vStatuses;    }        
func Abc_FrameReadObjIds(p *Abc_Frame_t) *Vec_Int_t { return s_GlobalFrame.vAbcObjIds;   }        
func Abc_FrameReadJsonStrs(p *Abc_Frame_t) *Abc_Nam_t { return s_GlobalFrame.pJsonStrs;    }     
func Abc_FrameReadJsonObjs(p *Abc_Frame_t) *Vec_Wec_t { return s_GlobalFrame.vJsonObjs;    }   
       
func Abc_FrameReadCexPiNum(p *Abc_Frame_t) int32 { return s_GlobalFrame.pCex.nPis;   }               
func Abc_FrameReadCexRegNum(p *Abc_Frame_t) int32 { return s_GlobalFrame.pCex.nRegs;  }               
func Abc_FrameReadCexPo(p *Abc_Frame_t) int32 { return s_GlobalFrame.pCex.iPo;    }               
func Abc_FrameReadCexFrame(p *Abc_Frame_t) int32 { return s_GlobalFrame.pCex.iFrame; }               

func Abc_FrameInputNdr(pAbc *Abc_Frame_t, pData any) { Ndr_Delete(s_GlobalFrame.pNdr); s_GlobalFrame.pNdr = pData;                        }
func Abc_FrameOutputNdr(pAbc *Abc_Frame_t) any { var pData any  = s_GlobalFrame.pNdr; s_GlobalFrame.pNdr = nil; return pData;             }  
func Abc_FrameOutputNdrArray(pAbc *Abc_Frame_t) *int32 { var pArray *int32  = s_GlobalFrame.pNdrArray; s_GlobalFrame.pNdrArray = nil; return pArray;  }

func Abc_FrameSetLibLut(pLib any) { s_GlobalFrame.pLibLut   = pLib;    } 
func Abc_FrameSetLibBox(pLib any) { s_GlobalFrame.pLibBox   = pLib;    } 
func Abc_FrameSetLibGen(pLib any) { s_GlobalFrame.pLibGen   = pLib;    } 
func Abc_FrameSetLibGen2(pLib any) { s_GlobalFrame.pLibGen2  = pLib;    } 
func Abc_FrameSetLibSuper(pLib any) { s_GlobalFrame.pLibSuper = pLib;    } 
func Abc_FrameSetFlag(pFlag string, pValue string) { Cmd_FlagUpdateValue( s_GlobalFrame, pFlag, pValue );               } 
func Abc_FrameSetCex(pCex *Abc_Cex_t) { ABC_FREE( s_GlobalFrame.pCex ); s_GlobalFrame.pCex = pCex;       }
func Abc_FrameSetNFrames(nFrames int32) { ABC_FREE( s_GlobalFrame.pCex ); s_GlobalFrame.nFrames = nFrames; }
func Abc_FrameSetStatus(Status int32) { ABC_FREE( s_GlobalFrame.pCex ); s_GlobalFrame.Status = Status;   }
func Abc_FrameSetManDsd(pMan any) { if s_GlobalFrame.pManDsd  && s_GlobalFrame.pManDsd  != pMan { If_DsdManFree((If_DsdMan_t *)s_GlobalFrame.pManDsd,  0); }
 s_GlobalFrame.pManDsd = pMan;  }
func Abc_FrameSetManDsd2(pMan any) { if s_GlobalFrame.pManDsd2 && s_GlobalFrame.pManDsd2 != pMan { If_DsdManFree((If_DsdMan_t *)s_GlobalFrame.pManDsd2, 0); }
 s_GlobalFrame.pManDsd2 = pMan; }
func Abc_FrameSetInv(vInv *Vec_Int_t) { Vec_IntFreeP(&s_GlobalFrame.pAbcWlcInv); s_GlobalFrame.pAbcWlcInv = vInv; }
func Abc_FrameSetJsonStrs(pStrs *Abc_Nam_t) { Abc_NamDeref( s_GlobalFrame.pJsonStrs ); s_GlobalFrame.pJsonStrs = pStrs; }
func Abc_FrameSetJsonObjs(vObjs *Vec_Wec_t) { Vec_WecFreeP(&s_GlobalFrame.vJsonObjs ); s_GlobalFrame.vJsonObjs = vObjs; }
func Abc_FrameSetSignalNames(vNames *Vec_Ptr_t) { if s_GlobalFrame.vSignalNames { Vec_PtrFreeFree( s_GlobalFrame.vSignalNames ); }
 s_GlobalFrame.vSignalNames = vNames; }
func Abc_FrameSetSpecName(pFileName string) { ABC_FREE( s_GlobalFrame.pSpecName ); s_GlobalFrame.pSpecName = pFileName; }

func Abc_FrameIsBatchMode() int32 { if s_GlobalFrame != nil {
    return s_GlobalFrame.fBatchMode;
} else {
    return 0;
}              } 
func Abc_FrameSetBatchMode(Mode int32) { if s_GlobalFrame != nil { s_GlobalFrame.fBatchMode = Mode; }             } 

func Abc_FrameIsBridgeMode() int32 { if s_GlobalFrame != nil {
    return s_GlobalFrame.fBridgeMode;
} else {
    return 0;
}             } 
func Abc_FrameSetBridgeMode() { if s_GlobalFrame != nil { s_GlobalFrame.fBridgeMode = 1; }               } 

func Abc_FrameReadDrivingCell() string { return s_GlobalFrame.pDrivingCell;    }              
func Abc_FrameReadMaxLoad() float32 { return s_GlobalFrame.MaxLoad;         }      
func Abc_FrameSetDrivingCell(pName string) { ABC_FREE(s_GlobalFrame.pDrivingCell); s_GlobalFrame.pDrivingCell   = pName; }      
func Abc_FrameSetMaxLoad(Load float32) { s_GlobalFrame.MaxLoad = Load;         }      

func Abc_FrameReadArrayMapping(pAbc *Abc_Frame_t) *int32 { return pAbc.pArray;                                            }
func Abc_FrameSetArrayMapping(p *int32) { ABC_FREE( s_GlobalFrame.pArray ); s_GlobalFrame.pArray = p;   }      

func Abc_FrameReadBoxes(pAbc *Abc_Frame_t) *int32 { return pAbc.pBoxes;                                            }
func Abc_FrameSetBoxes(p *int32) { ABC_FREE( s_GlobalFrame.pBoxes ); s_GlobalFrame.pBoxes = p;   }      

/**Function*************************************************************

  Synopsis    [Returns 1 if the flag is enabled without value or with value 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameIsFlagEnabled(pFlag string) int32 {
    var pValue string;
    // if flag is not defined, it is not enabled
    pValue = Abc_FrameReadFlag( pFlag );
    if pValue == nil {
        return 0; }
    // if flag is defined but value is not empty (no parameter) or "1", it is not enabled
    if (int32(strings.Compare(pValue, "")) != 0) && (int32(strings.Compare(pValue, "1")) != 0) {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameAllocate() *Abc_Frame_t {
    var p *Abc_Frame_t;
    extern void define_cube_size( int n );
    extern void set_espresso_flags();
    // allocate and clean
    p = new(Abc_Frame_t);
    // get version
    p.sVersion = Abc_UtilsGetVersion( p );
    // set streams
    p.Err = os.Stderr;
    p.Out = os.Stdout;
    p.Hst = nil;
    p.Status     = -1;
    p.nFrames    = -1;
    // set the starting step
    p.nSteps     =  1;
    p.fBatchMode =  0;
    // networks to be used by choice
    p.vStore = Vec_PtrAlloc( 16 );
    p.vAbcObjIds = Vec_IntAlloc( 0 );
    // initialize decomposition manager
//    define_cube_size(20);
//    set_espresso_flags();
    // initialize the trace manager
//    Abc_HManStart();
    p.vPlugInComBinPairs = Vec_PtrAlloc( 100 );
    return p;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameDeallocate(p *Abc_Frame_t) {
    extern void Rwt_ManGlobalStop();
    extern void undefine_cube_size();
//    extern void Ivy_TruthManStop();
//    Abc_HManStop();
//    undefine_cube_size();
    Rwt_ManGlobalStop();
//    Ivy_TruthManStop();
    if p.vAbcObjIds {  Vec_IntFree( p.vAbcObjIds ); }
    if p.vCexVec {  Vec_PtrFreeFree( p.vCexVec ); }
    if p.vPoEquivs {  Vec_VecFree( (Vec_Vec_t *)p.vPoEquivs ); }
    if p.vStatuses {  Vec_IntFree( p.vStatuses ); }
    if p.pManDec {  Dec_ManStop( (Dec_Man_t *)p.pManDec ); }
#ifdef ABC_USE_CUDD
    if ( p->dd        )  Extra_StopManager( p->dd );
#endif
    if p.vStore {  Vec_PtrFree( p.vStore ); }
    if p.pSave1 {  Aig_ManStop( (Aig_Man_t *)p.pSave1 ); }
    if p.pSave2 {  Aig_ManStop( (Aig_Man_t *)p.pSave2 ); }
    if p.pSave3 {  Aig_ManStop( (Aig_Man_t *)p.pSave3 ); }
    if p.pSave4 {  Aig_ManStop( (Aig_Man_t *)p.pSave4 ); }
    if p.pManDsd {  If_DsdManFree( (If_DsdMan_t *)p.pManDsd, 0 ); }
    if p.pManDsd2 {  If_DsdManFree( (If_DsdMan_t *)p.pManDsd2, 0 ); }
    if p.pNtkBackup {  Abc_NtkDelete( p.pNtkBackup ); }
    if p.vPlugInComBinPairs {
        var pTemp string;
        var i int32;
        Vec_PtrForEachEntry( char *, p.vPlugInComBinPairs, pTemp, i )
            ABC_FREE( pTemp );
        Vec_PtrFree( p.vPlugInComBinPairs );
    }
    Vec_IntFreeP( &p.vIndFlops );
    Vec_PtrFreeP( &p.vLTLProperties_global );
    if p.vSignalNames {
    Vec_PtrFreeFree( p.vSignalNames ); }
    ABC_FREE( p.pSpecName );
    Abc_FrameDeleteAllNetworks( p );
    ABC_FREE( p.pDrivingCell );
    ABC_FREE( p.pCex2 );
    ABC_FREE( p.pCex );
    Vec_IntFreeP( &p.pAbcWlcInv );
    Abc_NamDeref( s_GlobalFrame.pJsonStrs );
    Vec_WecFreeP( &s_GlobalFrame.vJsonObjs );  
    Ndr_Delete( s_GlobalFrame.pNdr );
    ABC_FREE( s_GlobalFrame.pNdrArray );

    Gia_ManStopP( &p.pGiaMiniAig );
    Gia_ManStopP( &p.pGiaMiniLut );
    Vec_IntFreeP( &p.vCopyMiniAig );
    Vec_IntFreeP( &p.vCopyMiniLut );
    ABC_FREE( p.pArray );
    ABC_FREE( p.pBoxes );
    

    ABC_FREE( p );
    s_GlobalFrame = nil;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameRestart(p *Abc_Frame_t) {
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameClearVerifStatus(p *Abc_Frame_t) {
    p.nFrames = -1;
    p.Status  = -1;
    ABC_FREE( p.pCex );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameShowProgress(p *Abc_Frame_t) int32 {
    return Abc_FrameIsFlagEnabled( "progressbar" );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadNtk(p *Abc_Frame_t) *Abc_Ntk_t {
    return p.pNtkCur;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadGia(p *Abc_Frame_t) *Gia_Man_t {
    return p.pGia;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadOut(p *Abc_Frame_t) *os.File {
    return p.Out;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadErr(p *Abc_Frame_t) *os.File {
    return p.Err;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadMode(p *Abc_Frame_t) int32 {
    var fShortNames int32;
    var pValue string;
    pValue = Cmd_FlagReadByName( p, "namemode" );
    if pValue == nil {
        fShortNames = 0;
    } else { 
        fShortNames = strconv.Atoi(pValue); }
    return fShortNames;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameSetMode(p *Abc_Frame_t, fNameMode int32) int32 {
    var Buffer [2]byte;
    var fNameModeOld int32;
    fNameModeOld = Abc_FrameReadMode( p );
    Buffer[0] = '0' + fNameMode;
    Buffer[1] = 0;
    Cmd_FlagUpdateValue( p, "namemode", (char *)Buffer );
    return fNameModeOld;
}


/**Function*************************************************************

  Synopsis    [Sets the given network to be the current one.]

  Description [Takes the network and makes it the current network.
  The previous current network is attached to the given network as 
  a backup copy. In the stack of backup networks contains too many
  networks (defined by the paramater "savesteps"), the bottom
  most network is deleted.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameSetCurrentNetwork(p *Abc_Frame_t, pNtkNew *Abc_Ntk_t) {
    var pNtk, pNtk2, pNtk3 *Abc_Ntk_t;
    var nNetsPresent int32;
    var nNetsToSave int32;
    var pValue string;

    if p.pNtkCur == pNtkNew {
        return; }

    // link it to the previous network
    Abc_NtkSetBackup( pNtkNew, p.pNtkCur );
    // set the step of this network
    Abc_NtkSetStep( pNtkNew, ++p.nSteps );
    // set this network to be the current network
    p.pNtkCur = pNtkNew;

    // remove any extra network that may happen to be in the stack
    pValue = Cmd_FlagReadByName( p, "savesteps" );
    // if the value of steps to save is not set, assume 1-level undo
    if pValue == nil {
        nNetsToSave = 1;
    } else { 
        nNetsToSave = strconv.Atoi(pValue); }
    
    // count the network, remember the last one, and the one before the last one
    nNetsPresent = 0;
    pNtk2 = pNtk3 = nil;
    for  pNtk = p.pNtkCur; (pNtk != nil); pNtk = Abc_NtkBackup(pNtk2)  {
        nNetsPresent++;
        pNtk3 = pNtk2;
        pNtk2 = pNtk;
    }

    // remove the earliest backup network if it is more steps away than we store
    if nNetsPresent - 1 > nNetsToSave { // delete the last network
        Abc_NtkDelete( pNtk2 );
        // clean the pointer of the network before the last one
        Abc_NtkSetBackup( pNtk3, nil );
    }
}

/**Function*************************************************************

  Synopsis    [This procedure swaps the current and the backup network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameSwapCurrentAndBackup(p *Abc_Frame_t) {
    var pNtkCur, pNtkBack, pNtkBack2 *Abc_Ntk_t;
    var iStepCur, iStepBack int32;

    pNtkCur  = p.pNtkCur;
    pNtkBack = Abc_NtkBackup( pNtkCur );
    iStepCur = Abc_NtkStep  ( pNtkCur );

    // if there is no backup nothing to reset
    if pNtkBack == nil {
        return; }

    // remember the backup of the backup
    pNtkBack2 = Abc_NtkBackup( pNtkBack );
    iStepBack = Abc_NtkStep  ( pNtkBack );

    // set pNtkCur to be the next after the backup's backup
    Abc_NtkSetBackup( pNtkCur, pNtkBack2 );
    Abc_NtkSetStep  ( pNtkCur, iStepBack );

    // set pNtkCur to be the next after the backup
    Abc_NtkSetBackup( pNtkBack, pNtkCur );
    Abc_NtkSetStep  ( pNtkBack, iStepCur );

    // set the current network
    p.pNtkCur = pNtkBack;
}


/**Function*************************************************************

  Synopsis    [Replaces the current network by the given one.]

  Description [This procedure does not modify the stack of saved
  networks.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReplaceCurrentNetwork(p *Abc_Frame_t, pNtk *Abc_Ntk_t) {
    if pNtk == nil {
        return; }

    if Abc_NtkPoNum(pNtk) == 0 {
        Abc_Print( 0, "The current network has no primary outputs. Some commands may not work correctly.\n" ); }

    // transfer the parameters to the new network
    if p.pNtkCur && (Abc_FrameIsFlagEnabled( "backup" ) != 0) {
        Abc_NtkSetBackup( pNtk, Abc_NtkBackup(p.pNtkCur) );
        Abc_NtkSetStep( pNtk, Abc_NtkStep(p.pNtkCur) );
        // delete the current network
        Abc_NtkDelete( p.pNtkCur );
    
    } else {
        Abc_NtkSetBackup( pNtk, nil );
        Abc_NtkSetStep( pNtk, ++p.nSteps );
        // delete the current network if present but backup is disabled
        if p.pNtkCur {
            Abc_NtkDelete( p.pNtkCur ); }
    }
    // set the new current network
    p.pNtkCur = pNtk;
}

/**Function*************************************************************

  Synopsis    [Removes library binding of all currently stored networks.]

  Description [This procedure is called when the library is freed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameUnmapAllNetworks(p *Abc_Frame_t) {
    var pNtk *Abc_Ntk_t;
    for  pNtk = p.pNtkCur; (pNtk != nil); pNtk = Abc_NtkBackup(pNtk)  {
        if Abc_NtkHasMapping(pNtk) {
            Abc_NtkMapToSop( pNtk ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameDeleteAllNetworks(p *Abc_Frame_t) {
    var pNtk, pNtk2 *Abc_Ntk_t;
    // delete all the currently saved networks
    for  pNtk  = p.pNtkCur, 
          pNtk2 = pNtk? Abc_NtkBackup(pNtk): nil; 
          (pNtk != nil); 
          pNtk  = pNtk2, 
          pNtk2 = pNtk? Abc_NtkBackup(pNtk): nil  {
        Abc_NtkDelete( pNtk );
    }
    // set the current network empty
    p.pNtkCur = nil;
//    fprintf( p->Out, "All networks have been deleted.\n" );
    Gia_ManStopP( &p.pGia );
    Gia_ManStopP( &p.pGia2 );
    Gia_ManStopP( &p.pGiaBest );
    Gia_ManStopP( &p.pGiaBest2 );
    Gia_ManStopP( &p.pGiaSaved );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameSetGlobalFrame(p *Abc_Frame_t) {
    s_GlobalFrame = p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameGetGlobalFrame() *Abc_Frame_t {
    if s_GlobalFrame == 0 {
        // start the framework
        s_GlobalFrame = Abc_FrameAllocate();
        // perform initializations
        Abc_FrameInit( s_GlobalFrame );
    }
    return s_GlobalFrame;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadGlobalFrame() *Abc_Frame_t {
    return s_GlobalFrame;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameSetSave1(pAig any) {
    var pFrame *Abc_Frame_t  = Abc_FrameGetGlobalFrame();
    if pFrame.pSave1 {
        Aig_ManStop( (Aig_Man_t *)pFrame.pSave1 ); }
    pFrame.pSave1 = pAig;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameSetSave2(pAig any) {
    var pFrame *Abc_Frame_t  = Abc_FrameGetGlobalFrame();
    if pFrame.pSave2 {
        Aig_ManStop( (Aig_Man_t *)pFrame.pSave2 ); }
    pFrame.pSave2 = pAig;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadSave1() any { var pAig any  = Abc_FrameGetGlobalFrame().pSave1; Abc_FrameGetGlobalFrame().pSave1 = nil; return pAig; }
func Abc_FrameReadSave2() any { var pAig any  = Abc_FrameGetGlobalFrame().pSave2; Abc_FrameGetGlobalFrame().pSave2 = nil; return pAig; }

/**Function*************************************************************

  Synopsis    [Returns 0/1 if pNtkCur is an AIG and PO is 0/1; -1 otherwise.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameCheckPoConst(p *Abc_Frame_t, iPoNum int32) int32 {
    var pObj *Abc_Obj_t;
    if p.pNtkCur == nil {
        return -1; }
    if !Abc_NtkIsStrash(p.pNtkCur) {
        return -1; }
    if iPoNum < 0 || iPoNum >= Abc_NtkPoNum(p.pNtkCur) {
        return -1; }
    pObj = Abc_NtkPo( p.pNtkCur, iPoNum );
    if !Abc_AigNodeIsConst(Abc_ObjFanin0(pObj)) {
        return -1; }
    return !Abc_ObjFaninC0(pObj);
}
func Abc_FrameCheckPoConstTest(p *Abc_Frame_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_NtkForEachPo( p.pNtkCur, pObj, i )
        fmt.Printf("%d = %d\n", i, Abc_FrameCheckPoConst(p, i));
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

