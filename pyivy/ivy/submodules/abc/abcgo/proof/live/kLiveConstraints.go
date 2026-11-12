package live

import "fmt"

/**CFile****************************************************************

  FileName    [kLiveConstraints.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Constraint analysis module for the k-Liveness algorithm
        invented by Koen Classen, Niklas Sorensson.]

  Author      [Sayak Ray]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 31, 2012.]

  Revision    [$Id: liveness.c,v 1.00 2009/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include "base/main/main.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include <string.h>
//#include "base/main/mainInt.h"
//#include "proof/pdr/pdr.h"

//ABC_NAMESPACE_IMPL_START

func createConstrained0LiveCone(pNewAig *Aig_Man_t, signalList *Vec_Ptr_t) *Aig_Obj_t {
    var pConsequent, pConsequentCopy, pAntecedent, p0LiveCone, pObj *Aig_Obj_t;
    var i, numSigAntecedent int32;
    
    numSigAntecedent = Vec_PtrSize( signalList ) - 1;

    pAntecedent = Aig_ManConst1( pNewAig );
    pConsequent = (Aig_Obj_t *)Vec_PtrEntry( signalList, numSigAntecedent );
    pConsequentCopy = Aig_NotCond( (Aig_Obj_t *)(Aig_Regular(pConsequent).pData), Aig_IsComplement( pConsequent ) );

    for i=0; i<numSigAntecedent; i++  {
        pObj = (Aig_Obj_t *)Vec_PtrEntry( signalList, i );
        assert(Aig_Regular(pObj).pData, "Aig_Regular(pObj).pData");
        pAntecedent = Aig_And( pNewAig, pAntecedent, Aig_NotCond((Aig_Obj_t *)(Aig_Regular(pObj).pData), Aig_IsComplement(pObj)) );
    }

    p0LiveCone = Aig_Or( pNewAig, Aig_Not(pAntecedent), pConsequentCopy );

    return p0LiveCone;
}

func collectCSSignals(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) *Vec_Ptr_t {
    var i int32;
    var pObj *Aig_Obj_t;
    var pConsequent *Aig_Obj_t = nil;
    var vNodeArray *Vec_Ptr_t;

    vNodeArray = Vec_PtrAlloc(1);

    Saig_ManForEachPo( pAig, pObj, i )
    {
        if strstr( Abc_ObjName(Abc_NtkPo( pNtk, i )), "csLiveConst_" ) != nil {
            Vec_PtrPush( vNodeArray, Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj)) );    
        } else if strstr( Abc_ObjName(Abc_NtkPo( pNtk, i )), "csLiveTarget_" ) != nil {
            pConsequent = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj)); }
    }
    assert((pConsequent != nil), "pConsequent");
    Vec_PtrPush( vNodeArray, pConsequent );    
    return vNodeArray;
}

func createNewAigWith0LivePo(pAig *Aig_Man_t, signalList *Vec_Ptr_t, index0Live *int32) *Aig_Man_t {
    var pNewAig *Aig_Man_t;
    var pObj, pObjNewPoDriver *Aig_Obj_t;
    var i int32;

    //assert( Vec_PtrSize( signalList ) > 1 );

    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNewAig = Aig_ManStart( Aig_ManObjNumMax(pAig) );
    pNewAig.pName = (char *)malloc( len(pAig.pName) + len("_0Live") + 1 );
    sprintf(pNewAig.pName, "%s_%s", pAig.pName, "0Live");
        pNewAig.pSpec = nil;

    //****************************************************************
    // Step 2: map constant nodes
    //****************************************************************
        pObj = Aig_ManConst1( pAig );
        pObj.pData = Aig_ManConst1( pNewAig );

    //****************************************************************
        // Step 3: create true PIs
    //****************************************************************
        Saig_ManForEachPi( pAig, pObj, i )
    {
        pObj.pData = Aig_ObjCreateCi( pNewAig );
    }

    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
        Saig_ManForEachLo( pAig, pObj, i )
        {
        pObj.pData = Aig_ObjCreateCi( pNewAig );
        }

    //********************************************************************
    // Step 7: create internal nodes
    //********************************************************************
        Aig_ManForEachNode( pAig, pObj, i )
    {
        pObj.pData = Aig_And( pNewAig, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }

    Saig_ManForEachPo( pAig, pObj, i )
    {
        pObj.pData = Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) ); 
    }

    pObjNewPoDriver = createConstrained0LiveCone( pNewAig, signalList );
    Aig_ObjCreateCo( pNewAig, pObjNewPoDriver );
    *index0Live = i;

    Saig_ManForEachLi( pAig, pObj, i )
    {
        pObj.pData = Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) );
    }

    Aig_ManSetRegNum( pNewAig, Aig_ManRegNum(pAig) );
    Aig_ManCleanup( pNewAig );

    assert(Aig_ManCheck( pNewAig ), "Aig_ManCheck( pNewAig )");
    return pNewAig;
}

func checkMonotoneSignal() *Vec_Ptr_t {
    return nil;
}

func gatherMonotoneSignals(pAig *Aig_Man_t) *Vec_Ptr_t {
    var i int32;
    var pObj *Aig_Obj_t;

    Aig_ManForEachNode( pAig, pObj, i )
    {
        Aig_ObjPrint( pAig, pObj );
        fmt.Printf("\n");
    }
    
    return nil;
}

func generateWorkingAig(pAig *Aig_Man_t, pNtk *Abc_Ntk_t, pIndex0Live *int32) *Aig_Man_t {
    var vSignalVector *Vec_Ptr_t;
    var pAigNew *Aig_Man_t;

    vSignalVector = collectCSSignals( pNtk, pAig );    
    assert((vSignalVector != nil), "vSignalVector");
    pAigNew = createNewAigWith0LivePo( pAig, vSignalVector, pIndex0Live );
    Vec_PtrFree(vSignalVector);

    return pAigNew;
}

//ABC_NAMESPACE_IMPL_END

