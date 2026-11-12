package mio

import "strings"

/**CFile****************************************************************

  FileName    [mioApi.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [File reading/writing for technology mapping.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: mioApi.c,v 1.4 2004/06/28 14:20:25 alanmi Exp $]

***********************************************************************/

//#include "mioInt.h"

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
func Mio_LibraryReadName(pLib *Mio_Library_t) string { return pLib.pName;      }
func Mio_LibraryReadGateNum(pLib *Mio_Library_t) int32 { return pLib.nGates;     }
func Mio_LibraryReadGates(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGates;     }
func Mio_LibraryReadGateArray(pLib *Mio_Library_t) **Mio_Gate_t { return pLib.ppGatesName;}
func Mio_LibraryReadGateById(pLib *Mio_Library_t, Id int32) *Mio_Gate_t { assert(pLib.ppGates0[Id].Cell == Id, "pLib.ppGates0[Id].Cell == Id"); return pLib.ppGates0[Id];}
func Mio_LibraryReadBuf(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGateBuf;   }
func Mio_LibraryReadInv(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGateInv;   }
func Mio_LibraryReadConst0(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGate0;     }
func Mio_LibraryReadConst1(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGate1;     }
func Mio_LibraryReadNand2(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGateNand2; }
func Mio_LibraryReadAnd2(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGateAnd2;  }
func Mio_LibraryReadNor2(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGateNor2;  }
func Mio_LibraryReadOr2(pLib *Mio_Library_t) *Mio_Gate_t { return pLib.pGateOr2;   }
func Mio_LibraryReadDelayInvRise(pLib *Mio_Library_t) float32 { return float32((pLib.pGateInv?   pLib.pGateInv.pPins.dDelayBlockRise   : 0.0)); }
func Mio_LibraryReadDelayInvFall(pLib *Mio_Library_t) float32 { return float32((pLib.pGateInv?   pLib.pGateInv.pPins.dDelayBlockFall   : 0.0)); }
func Mio_LibraryReadDelayInvMax(pLib *Mio_Library_t) float32 { return float32((pLib.pGateInv?   pLib.pGateInv.pPins.dDelayBlockMax    : 0.0)); }
func Mio_LibraryReadDelayNand2Rise(pLib *Mio_Library_t) float32 { return float32((pLib.pGateNand2? pLib.pGateNand2.pPins.dDelayBlockRise : 0.0)); }
func Mio_LibraryReadDelayNand2Fall(pLib *Mio_Library_t) float32 { return float32((pLib.pGateNand2? pLib.pGateNand2.pPins.dDelayBlockFall : 0.0)); }
func Mio_LibraryReadDelayNand2Max(pLib *Mio_Library_t) float32 { return float32((pLib.pGateNand2? pLib.pGateNand2.pPins.dDelayBlockMax  : 0.0)); }
func Mio_LibraryReadDelayAnd2Max(pLib *Mio_Library_t) float32 { return float32((pLib.pGateAnd2?  pLib.pGateAnd2.pPins.dDelayBlockMax   : 0.0)); }
func Mio_LibraryReadDelayAigNode(pLib *Mio_Library_t) float32 { if pLib.pGateAnd2 {
    return Mio_LibraryReadDelayAnd2Max(pLib);
} else {
    return Mio_LibraryReadDelayNand2Max(pLib);
} } // approximate delay of the AIG node
func Mio_LibraryReadAreaInv(pLib *Mio_Library_t) float32 { return float32((pLib.pGateInv?   pLib.pGateInv.dArea   : 0.0)); }
func Mio_LibraryReadAreaBuf(pLib *Mio_Library_t) float32 { return float32((pLib.pGateBuf?   pLib.pGateBuf.dArea   : 0.0)); }
func Mio_LibraryReadAreaNand2(pLib *Mio_Library_t) float32 { return float32((pLib.pGateNand2? pLib.pGateNand2.dArea : 0.0)); }
// inlining of macro ABC_FREE
func Mio_LibrarySetName(pLib *Mio_Library_t, pName string) { ((pLib.pName) ? (free((char *) (pLib.pName)), (pLib.pName) = 0) : 0); pLib.pName = pName;         }

/**Function*************************************************************

  Synopsis    [Returns the longest gate name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadGateNameMax(pLib *Mio_Library_t) int32 {
    var pGate *Mio_Gate_t;
    var LenMax, LenCur int32;
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        LenCur = len(Mio_GateReadName(pGate));
        if LenMax < LenCur {
            LenMax = LenCur; }
    }
    return LenMax;
}

/**Function*************************************************************

  Synopsis    [Read Mvc of the gate by name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadGateByName(pLib *Mio_Library_t, pName string, pOutName string) *Mio_Gate_t { 
    var pGate *Mio_Gate_t;
    if ! st__lookup( pLib.tName2Gate, pName, (char **)&pGate ) {
        return nil; }
    if pOutName == nil {
        return pGate; }
    if !(int32(strings.Compare(pGate.pOutName, pOutName)) != 0) {
        return pGate; }
    if pGate.pTwin && !(int32(strings.Compare(pGate.pTwin.pOutName, pOutName)) != 0) {
        return pGate.pTwin; }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Returns the first gate in the library with the given TT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadGateByTruth(pLib *Mio_Library_t, t word) *Mio_Gate_t {
    var pGate *Mio_Gate_t;
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if pGate.nInputs <= 6 && pGate.uTruth == t {
            return pGate; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Read Mvc of the gate by name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadSopByName(pLib *Mio_Library_t, pName string) string { 
    var pGate *Mio_Gate_t;
    if st__lookup( pLib.tName2Gate, pName, (char **)&pGate ) {
        return pGate.pSop; }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_GateReadName(pGate *Mio_Gate_t) string { return pGate.pName;     }
func Mio_GateReadOutName(pGate *Mio_Gate_t) string { return pGate.pOutName;  }
func Mio_GateReadArea(pGate *Mio_Gate_t) float64 { return pGate.dArea;     }
func Mio_GateReadForm(pGate *Mio_Gate_t) string { return pGate.pForm;     }
func Mio_GateReadPins(pGate *Mio_Gate_t) *Mio_Pin_t { return pGate.pPins;     }
func Mio_GateReadLib(pGate *Mio_Gate_t) *Mio_Library_t { return pGate.pLib;      }
func Mio_GateReadNext(pGate *Mio_Gate_t) *Mio_Gate_t { return pGate.pNext;     }
func Mio_GateReadTwin(pGate *Mio_Gate_t) *Mio_Gate_t { return pGate.pTwin;     }
func Mio_GateReadPinNum(pGate *Mio_Gate_t) int32 { return pGate.nInputs;   }
func Mio_GateReadDelayMax(pGate *Mio_Gate_t) float64 { return pGate.dDelayMax; }
func Mio_GateReadSop(pGate *Mio_Gate_t) string { return pGate.pSop;      }
func Mio_GateReadExpr(pGate *Mio_Gate_t) *Vec_Int_t { return pGate.vExpr;     }
func Mio_GateReadTruth(pGate *Mio_Gate_t) word { if pGate.nInputs <= 6 {
    return pGate.uTruth;
} else {
    return 0;
}   }
func Mio_GateReadTruthP(pGate *Mio_Gate_t) *word { if pGate.nInputs <= 6 {
    return &pGate.uTruth;
} else {
    return pGate.pTruth;
} }
func Mio_GateReadValue(pGate *Mio_Gate_t) int32 { return pGate.Value;     }
func Mio_GateReadCell(pGate *Mio_Gate_t) int32 { return pGate.Cell;      }
func Mio_GateReadProfile(pGate *Mio_Gate_t) int32 { return pGate.Profile;   }
func Mio_GateReadProfile2(pGate *Mio_Gate_t) int32 { return pGate.Profile2;  }
func Mio_GateSetValue(pGate *Mio_Gate_t, Value int32) { pGate.Value = Value;    }
func Mio_GateSetCell(pGate *Mio_Gate_t, Cell int32) { pGate.Cell  = Cell;     }
func Mio_GateSetProfile(pGate *Mio_Gate_t, Prof int32) { pGate.Profile = Prof;   }
func Mio_GateSetProfile2(pGate *Mio_Gate_t, Prof int32) { pGate.Profile2 = Prof;  }
func Mio_GateIncProfile2(pGate *Mio_Gate_t) { pGate.Profile2++;       }
func Mio_GateDecProfile2(pGate *Mio_Gate_t) { pGate.Profile2--;       }
func Mio_GateAddToProfile(pGate *Mio_Gate_t, Prof int32) { pGate.Profile  += Prof; }
func Mio_GateAddToProfile2(pGate *Mio_Gate_t, Prof int32) { pGate.Profile2 += Prof; }
func Mio_GateIsInv(pGate *Mio_Gate_t) int32 { return pGate.uTruth == ABC_CONST(0x5555555555555555); }

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_PinReadName(pPin *Mio_Pin_t) string { return pPin.pName;           }
func Mio_PinReadPhase(pPin *Mio_Pin_t) Mio_PinPhase_t { return pPin.Phase;           }
func Mio_PinReadInputLoad(pPin *Mio_Pin_t) float64 { return pPin.dLoadInput;      }
func Mio_PinReadMaxLoad(pPin *Mio_Pin_t) float64 { return pPin.dLoadMax;        }
func Mio_PinReadDelayBlockRise(pPin *Mio_Pin_t) float64 { return pPin.dDelayBlockRise; }
func Mio_PinReadDelayFanoutRise(pPin *Mio_Pin_t) float64 { return pPin.dDelayFanoutRise;}
func Mio_PinReadDelayBlockFall(pPin *Mio_Pin_t) float64 { return pPin.dDelayBlockFall; }
func Mio_PinReadDelayFanoutFall(pPin *Mio_Pin_t) float64 { return pPin.dDelayFanoutFall;}
func Mio_PinReadDelayBlockMax(pPin *Mio_Pin_t) float64 { return pPin.dDelayBlockMax;  }
func Mio_PinReadNext(pPin *Mio_Pin_t) *Mio_Pin_t { return pPin.pNext;           }

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_GateReadPinName(pGate *Mio_Gate_t, iPin int32) string {
    var pPin *Mio_Pin_t;
    var i int32;
    // inlining of macro Mio_GateForEachPin
    for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
        if i++ == iPin {
            return Mio_PinReadName(pPin); }
    }
    return nil;
}
func Mio_GateReadPinDelay(pGate *Mio_Gate_t, iPin int32) float32 {
    var pPin *Mio_Pin_t;
    var i int32;
    // inlining of macro Mio_GateForEachPin
    for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
        if i++ == iPin {
            return 0.5 * pPin.dDelayBlockRise + 0.5 * pPin.dDelayBlockFall; }
    }
    return ABC_INFINITY;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

