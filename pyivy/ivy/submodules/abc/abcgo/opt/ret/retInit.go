package ret

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [retInit.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Retiming package.]

  Synopsis    [Initial state computation for backward retiming.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Oct 31, 2006.]

  Revision    [$Id: retInit.c,v 1.00 2006/10/31 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "retInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Abc_NtkRetimeVerifyModel( Abc_Ntk_t * pNtkCone, Vec_Int_t * vValues, int * pModel );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes initial values of the new latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeInitialValues(pNtkCone *Abc_Ntk_t, vValues *Vec_Int_t, fVerbose int32) *Vec_Int_t {
    var vSolution *Vec_Int_t;
    var pNtkMiter, pNtkLogic *Abc_Ntk_t;
    var RetValue int32;
    var clk abctime;
    if pNtkCone == nil {
        return Vec_IntDup( vValues ); }
    // convert the target network to AIG
    pNtkLogic = Abc_NtkDup( pNtkCone );
    Abc_NtkToAig( pNtkLogic );
    // get the miter
    pNtkMiter = Abc_NtkCreateTarget( pNtkLogic, pNtkLogic.vCos, vValues );
    if fVerbose != 0 {
        fmt.Printf("The miter for initial state computation has %d AIG nodes. ", Abc_NtkNodeNum(pNtkMiter)); }
    // solve the miter
    clk = Abc_Clock();
    RetValue = Abc_NtkMiterSat( pNtkMiter, (ABC_INT64_T)500000, (ABC_INT64_T)50000000, 0, nil, nil );
    // inlining of macro ABC_PRT
    if fVerbose != 0 { Abc_Print(1, "%s =", ("SAT solving time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
    // analyze the result
    if RetValue == 1 {
        fmt.Printf("Abc_NtkRetimeInitialValues(): The problem is unsatisfiable. DC latch values are used.\n");
    } else if RetValue == -1 {
        fmt.Printf("Abc_NtkRetimeInitialValues(): The SAT problem timed out. DC latch values are used.\n");
    } else if !(Abc_NtkRetimeVerifyModel( pNtkCone, vValues, pNtkMiter.pModel ) != 0) {
        fmt.Printf("Abc_NtkRetimeInitialValues(): The computed counter-example is incorrect.\n"); }
    // set the values of the latches
    if RetValue != 0 {
        vSolution = nil;
    } else {
        vSolution = Vec_IntAllocArray( pNtkMiter.pModel, Abc_NtkPiNum(pNtkLogic) );
    }
    pNtkMiter.pModel = nil;
    Abc_NtkDelete( pNtkMiter );
    Abc_NtkDelete( pNtkLogic );
    return vSolution;
}

/**Function*************************************************************

  Synopsis    [Computes the results of simulating one node.]

  Description [Assumes that fanins have pCopy set to the input values.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjSopSimulate(pObj *Abc_Obj_t) int32 {
    var pCube string;
    var pSop string = (char *)pObj.pData;
    var nVars, Value, v, ResOr, ResAnd, ResVar int32;
    assert(pSop && !Abc_SopIsExorType(pSop), "pSop && !Abc_SopIsExorType(pSop)");
    // simulate the SOP of the node
    ResOr = 0;
    nVars = Abc_SopGetVarNum(pSop);
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        ResAnd = 1;
        // inlining of macro Abc_CubeForEachVar
        for  v = 0; (pCube[v] != ' ') && (Value = pCube[v]); v++  {
            if Value == '0' {
                ResVar = 1 ^ (int32((ABC_PTRUINT_T))Abc_ObjFanin(pObj, v).pCopy);
            } else if Value == '1' {
                ResVar = int32((ABC_PTRUINT_T))Abc_ObjFanin(pObj, v).pCopy;
            } else {
                continue; }
            ResAnd &= ResVar;
        }
        ResOr |= ResAnd;
    }
    // complement the result if necessary
    if !Abc_SopGetPhase(pSop) {
        ResOr ^= 1; }
    return ResOr;
}

/**Function*************************************************************

  Synopsis    [Verifies the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeVerifyModel(pNtkCone *Abc_Ntk_t, vValues *Vec_Int_t, pModel *int32) int32 {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, Counter int32;
    assert(Abc_NtkIsSopLogic(pNtkCone), "Abc_NtkIsSopLogic(pNtkCone)");
    // set the PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtkCone)) && (((pObj) = Abc_NtkPi(pNtkCone, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)pModel[i];
    }
    // simulate the internal nodes
    vNodes = Abc_NtkDfs( pNtkCone, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)Abc_ObjSopSimulate( pObj );
    }
    Vec_PtrFree( vNodes );
    // compare the outputs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkCone)) && (((pObj) = Abc_NtkPo(pNtkCone, i)), 1); i++  {
        pObj.pCopy = Abc_ObjFanin0(pObj).pCopy;
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkCone)) && (((pObj) = Abc_NtkPo(pNtkCone, i)), 1); i++  {
        Counter += (Vec_IntEntry(vValues, i) != int32((ABC_PTRUINT_T))pObj.pCopy);
    }
    if Counter > 0 {
        fmt.Printf("%d outputs (out of %d) have a value mismatch.\n", Counter, Abc_NtkPoNum(pNtkCone)); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Transfer latch initial values to pCopy.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeTranferToCopy(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsLatch(pObj) {
            pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)Abc_LatchIsInit1(pObj); }
    }
}

/**Function*************************************************************

  Synopsis    [Transfer latch initial values from pCopy.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeTranferFromCopy(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsLatch(pObj) {
            pObj.pData = (void *)(ABC_PTRUINT_T)(pObj.pCopy? ABC_INIT_ONE : ABC_INIT_ZERO); }
    }
}

/**Function*************************************************************

  Synopsis    [Transfer latch initial values to pCopy.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeCollectLatchValues(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vValues *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i int32;
    vValues = Vec_IntAlloc( Abc_NtkLatchNum(pNtk) );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsLatch(pObj) {
            Vec_IntPush( vValues, Abc_LatchIsInit1(pObj) ); }
    }
    return vValues;
}

/**Function*************************************************************

  Synopsis    [Transfer latch initial values from pCopy.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeInsertLatchValues(pNtk *Abc_Ntk_t, vValues *Vec_Int_t) {
    var pObj *Abc_Obj_t;
    var i, Counter int32;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsLatch(pObj) {
            pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)Counter++; }
    }
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsLatch(pObj) {
            pObj.pData = (Abc_Obj_t *)(ABC_PTRUINT_T)(vValues? (Vec_IntEntry(vValues,int32((ABC_PTRUINT_T))pObj.pCopy)? ABC_INIT_ONE : ABC_INIT_ZERO) : ABC_INIT_DC); }
    }
}

/**Function*************************************************************

  Synopsis    [Transfer latch initial values to pCopy.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeBackwardInitialStart(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // create the network used for the initial state computation
    pNtkNew = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_SOP, 1 );
    // create POs corresponding to the initial values
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsLatch(pObj) {
            pObj.pCopy = Abc_NtkCreatePo(pNtkNew); }
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Transfer latch initial values to pCopy.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeBackwardInitialFinish(pNtk *Abc_Ntk_t, pNtkNew *Abc_Ntk_t, vValuesOld *Vec_Int_t, fVerbose int32) {
    var vValuesNew *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // create PIs corresponding to the initial values
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjIsLatch(pObj) {
            Abc_ObjAddFanin( pObj.pCopy, Abc_NtkCreatePi(pNtkNew) ); }
    }
    // assign dummy node names
    Abc_NtkAddDummyPiNames( pNtkNew );
    Abc_NtkAddDummyPoNames( pNtkNew );
    // check the network
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkRetimeBackwardInitialFinish(): Network check has failed.\n"); }
    // derive new initial values
    vValuesNew = Abc_NtkRetimeInitialValues( pNtkNew, vValuesOld, fVerbose );
    // insert new initial values
    Abc_NtkRetimeInsertLatchValues( pNtk, vValuesNew );
    if vValuesNew != nil { Vec_IntFree( vValuesNew ); }
}


/**Function*************************************************************

  Synopsis    [Cycles the circuit to create a new initial state.]

  Description [Simulates the circuit with random input for the given 
  number of timeframes to get a better initial state.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCycleInitStateSop(pNtk *Abc_Ntk_t, nFrames int32, fVerbose int32) {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, f int32;
    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");
    srand( 0x12341234 );
    // initialize the values
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)(rand() & 1);
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)Abc_LatchIsInit1(pObj); }
    }
    // simulate for the given number of timeframes
    vNodes = Abc_NtkDfs( pNtk, 0 );
    for  f = 0; f < nFrames; f++  {
        // simulate internal nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)Abc_ObjSopSimulate( pObj );
        }
        // bring the results to the COs
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
            pObj.pCopy = Abc_ObjFanin0(pObj).pCopy;
        }
        // assign PI values
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)(rand() & 1);
        }
        // transfer the latch values
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
            Abc_ObjFanout0(pObj).pCopy = Abc_ObjFanin0(pObj).pCopy; }
        }
    }
    Vec_PtrFree( vNodes );
    // set the final values
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        pObj.pData = (void *)(ABC_PTRUINT_T)(Abc_ObjFanout0(pObj).pCopy ? ABC_INIT_ONE : ABC_INIT_ZERO); }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

