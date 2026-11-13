package abci

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcXsim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Using X-valued simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcXsim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#define XVS0   ABC_INIT_ZERO
#define XVS1   ABC_INIT_ONE
#define XVSX   ABC_INIT_DC

func Abc_ObjSetXsim(pObj *Abc_Obj_t, Value int32) { pObj.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)Value;  }
func Abc_ObjGetXsim(pObj *Abc_Obj_t) int32 { return int32((ABC_PTRINT_T))pObj.pCopy;           }
// Three-valued helpers are translated in xsim_value.go.
func Abc_ObjGetXsimFanin0(pObj *Abc_Obj_t) int32 { 
    var RetValue int32;
    RetValue = Abc_ObjGetXsim(Abc_ObjFanin0(pObj));
    if Abc_ObjFaninC0(pObj) {
        return Abc_XsimInv(RetValue);
    } else {
        return RetValue;
    }
}
func Abc_ObjGetXsimFanin1(pObj *Abc_Obj_t) int32 { 
    var RetValue int32;
    RetValue = Abc_ObjGetXsim(Abc_ObjFanin1(pObj));
    if Abc_ObjFaninC1(pObj) {
        return Abc_XsimInv(RetValue);
    } else {
        return RetValue;
    }
}
func Abc_XsimPrint(pFile *os.File, Value int32) { 
    if Value == XVS0 {
        fmt.Fprintf(pFile, "0");
        return;
    }
    if Value == XVS1 {
        fmt.Fprintf(pFile, "1");
        return;
    }
    assert(Value == XVSX, "Value == XVSX");       
    fmt.Fprintf(pFile, "x");
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs X-valued simulation of the sequential network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkXValueSimulate(pNtk *Abc_Ntk_t, nFrames int32, fXInputs int32, fXState int32, fVerbose int32) {
    var pObj *Abc_Obj_t;
    var i, f int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
//    srand( 0x12341234 );
    Gia_ManRandom( 1 );
    // start simulation
    Abc_ObjSetXsim( Abc_AigConst1(pNtk), XVS1 );
    if fXInputs != 0 {
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            Abc_ObjSetXsim( pObj, XVSX );
        }
    
    } else {
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            Abc_ObjSetXsim( pObj, Abc_XsimRand2() );
        }
    }
    if fXState != 0 {
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            Abc_ObjSetXsim( Abc_ObjFanout0(pObj), XVSX ); }
        }
    
    } else {
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            Abc_ObjSetXsim( Abc_ObjFanout0(pObj), Abc_LatchInit(pObj) ); }
        }
    }
    // simulate and print the result
    fmt.Fprintf(os.Stdout, "Frame : Inputs : Latches : Outputs\n");
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Abc_AigForEachAnd
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
            Abc_ObjSetXsim( pObj, Abc_XsimAnd(Abc_ObjGetXsimFanin0(pObj), Abc_ObjGetXsimFanin1(pObj)) ); }
        }
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
            Abc_ObjSetXsim( pObj, Abc_ObjGetXsimFanin0(pObj) );
        }
        // print out
        fmt.Fprintf(os.Stdout, "%2d : ", f);
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            Abc_XsimPrint( os.Stdout, Abc_ObjGetXsim(pObj) );
        }
        fmt.Fprintf(os.Stdout, " : ");
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
//            if ( Abc_ObjGetXsim(Abc_ObjFanout0(pObj)) != XVSX )
//                printf( " %s=", Abc_ObjName(pObj) );
            Abc_XsimPrint( os.Stdout, Abc_ObjGetXsim(Abc_ObjFanout0(pObj)) );
        }
        }
        fmt.Fprintf(os.Stdout, " : ");
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            Abc_XsimPrint( os.Stdout, Abc_ObjGetXsim(pObj) );
        }
        fmt.Fprintf(os.Stdout, "\n");
        // assign input values
        if fXInputs != 0 {
            // inlining of macro Abc_NtkForEachPi
            for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
                Abc_ObjSetXsim( pObj, XVSX );
            }
        
        } else {
            // inlining of macro Abc_NtkForEachPi
            for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
                Abc_ObjSetXsim( pObj, Abc_XsimRand2() );
            }
        }
        // transfer the latch values
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            Abc_ObjSetXsim( Abc_ObjFanout0(pObj), Abc_ObjGetXsim(Abc_ObjFanin0(pObj)) ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Cycles the circuit to create a new initial state.]

  Description [Simulates the circuit with random (or ternary) input 
  for the given number of timeframes to get a better initial state.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCycleInitState(pNtk *Abc_Ntk_t, nFrames int32, fUseXval int32, fVerbose int32) { 
    var pObj *Abc_Obj_t;
    var i, f int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
//    srand( 0x12341234 );
    Gia_ManRandom( 1 );
    // initialize the values
    Abc_ObjSetXsim( Abc_AigConst1(pNtk), XVS1 );
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        Abc_ObjSetXsim( Abc_ObjFanout0(pObj), Abc_LatchInit(pObj) ); }
    }
    // simulate for the given number of timeframes
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            Abc_ObjSetXsim( pObj, fUseXval? ABC_INIT_DC : Abc_XsimRand2() );
        }
//            Abc_ObjSetXsim( pObj, ABC_INIT_ONE );
        // inlining of macro Abc_AigForEachAnd
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
            Abc_ObjSetXsim( pObj, Abc_XsimAnd(Abc_ObjGetXsimFanin0(pObj), Abc_ObjGetXsimFanin1(pObj)) ); }
        }
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
            Abc_ObjSetXsim( pObj, Abc_ObjGetXsimFanin0(pObj) );
        }
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            Abc_ObjSetXsim( Abc_ObjFanout0(pObj), Abc_ObjGetXsim(Abc_ObjFanin0(pObj)) ); }
        }
    }
    // set the final values
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        pObj.pData = (void *)(ABC_PTRINT_T)Abc_ObjGetXsim(Abc_ObjFanout0(pObj));
//        printf( "%d", Abc_LatchIsInit1(pObj) );
    }
    }
//    printf( "\n" );
}

///////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

