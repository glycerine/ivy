package live

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [liveness_sim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Main implementation module.]

  Author      [Sayak Ray]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2009.]

  Revision    [$Id: liveness_sim.c,v 1.00 2009/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include "base/main/main.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include <string.h>

//ABC_NAMESPACE_IMPL_START


#define PROPAGATE_NAMES
//#define DUPLICATE_CKT_DEBUG

// extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern Abc_Ntk_t * Abc_NtkFromAigPhase( Aig_Man_t * pMan );

//char *strdup(const char *string);


/*******************************************************************
LAYOUT OF PI VECTOR:
 
+------------------------------------------------------------------------------------------------------------------------------------+
| TRUE ORIGINAL PI (n) | SAVE(PI) (1) | ORIGINAL LO (k) | SAVED(LO) (1) | SHADOW_ORIGINAL LO (k) | LIVENESS LO (l) | FAIRNESS LO (f) |
+------------------------------------------------------------------------------------------------------------------------------------+
<------------True PI----------------->|<----------------------------LO--------------------------------------------------------------->

LAYOUT OF PO VECTOR:

+-----------------------------------------------------------------------------------------------------------+
| SOLE PO (1) | ORIGINAL LI (k) | SAVED LI (1) | SHADOW_ORIGINAL LI (k) | LIVENESS LI (l) | FAIRNESS LI (f) |
+-----------------------------------------------------------------------------------------------------------+
<--True PO--->|<--------------------------------------LI---------------------------------------------------->

********************************************************************/

func printVecPtrOfString(vec *Vec_Ptr_t) {
    var i int32;

    for  i=0; i< Vec_PtrSize( vec ); i++  {
        fmt.Printf("vec[%d] = %s\n", i, (char *)Vec_PtrEntry(vec, i));
    }
}

func getPoIndex(pAig *Aig_Man_t, pPivot *Aig_Obj_t) int32 {
    var i int32;
    var pObj *Aig_Obj_t;

    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        if pObj == pPivot {
            return i; }
    }
    return -1;
}

func retrieveTruePiName(pNtkOld *Abc_Ntk_t, pAigOld *Aig_Man_t, pAigNew *Aig_Man_t, pObjPivot *Aig_Obj_t) string {
    var pObjOld, pObj *Aig_Obj_t;
    var pNode *Abc_Obj_t;
    var index int32;

    assert(Saig_ObjIsPi( pAigNew, pObjPivot ), "Saig_ObjIsPi( pAigNew, pObjPivot )");
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  index = 0; (index < Vec_PtrSize(pAigNew.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigNew.vCis, index)), 1); index++  {
        if pObj == pObjPivot {
            break; }
    }
    assert(index < Aig_ManCiNum( pAigNew ) - Aig_ManRegNum( pAigNew ), "index < Aig_ManCiNum( pAigNew ) - Aig_ManRegNum( pAigNew )");
    if index == Saig_ManPiNum( pAigNew ) - 1 {
        return "SAVE_BIERE";
    } else {
        pObjOld = Aig_ManCi( pAigOld, index );
        pNode = Abc_NtkPi( pNtkOld, index );
        assert(pObjOld.pData == pObjPivot, "pObjOld.pData == pObjPivot");
        return Abc_ObjName( pNode );
    }
}

func retrieveLOName(pNtkOld *Abc_Ntk_t, pAigOld *Aig_Man_t, pAigNew *Aig_Man_t, pObjPivot *Aig_Obj_t, vLive *Vec_Ptr_t, vFair *Vec_Ptr_t) string {
    var pObjOld, pObj *Aig_Obj_t;
    var pNode *Abc_Obj_t;
    var index int32;
    var oldIndex int32;
    var originalLatchNum int32 = Saig_ManRegNum(pAigOld);
    var strMatch int32;
    var i int32;
    var dummyStr string  = (char *)malloc( sizeof(char) * 50 );

    assert(Saig_ObjIsLo( pAigNew, pObjPivot ), "Saig_ObjIsLo( pAigNew, pObjPivot )");
    // inlining of macro Saig_ManForEachLo
    for  index = 0; (index < Saig_ManRegNum(pAigNew)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigNew.vCis, index+Saig_ManPiNum(pAigNew))), 1); index++  {
        if pObj == pObjPivot {
            break; }
    }
    if index < originalLatchNum {
        oldIndex = Saig_ManPiNum( pAigOld ) + index;
        pObjOld = Aig_ManCi( pAigOld, oldIndex );
        pNode = Abc_NtkCi( pNtkOld, oldIndex );
        assert(pObjOld.pData == pObjPivot, "pObjOld.pData == pObjPivot");
        return Abc_ObjName( pNode );
    
    } else if index == originalLatchNum {
        return "SAVED_LO";
    } else if index > originalLatchNum && index < 2 * originalLatchNum + 1 {
        oldIndex = Saig_ManPiNum( pAigOld ) + index - originalLatchNum - 1;
        pObjOld = Aig_ManCi( pAigOld, oldIndex );
        pNode = Abc_NtkCi( pNtkOld, oldIndex );
        dummyStr = fmt.Sprintf("%s__%s", Abc_ObjName( pNode ), "SHADOW");
        return dummyStr;
    
    } else if index >= 2 * originalLatchNum + 1 && index < 2 * originalLatchNum + 1 + Vec_PtrSize( vLive ) {
        oldIndex = index - 2 * originalLatchNum - 1;
        strMatch = 0;
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAigOld)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigOld.vCos, i)), 1); i++  {
            pNode = Abc_NtkPo( pNtkOld, i );
            if strstr( Abc_ObjName( pNode ), "assert_fair" ) != nil {
                if strMatch == oldIndex {
                    dummyStr = fmt.Sprintf("%s__%s", Abc_ObjName( pNode ), "LIVENESS");
                    return dummyStr;
                
                } else {
                    strMatch++; }
            }
        }
    
    } else if index >= 2 * originalLatchNum + 1 + Vec_PtrSize( vLive ) && index < 2 * originalLatchNum + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) {
        oldIndex = index - 2 * originalLatchNum - 1 - Vec_PtrSize( vLive );
        strMatch = 0;
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAigOld)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAigOld.vCos, i)), 1); i++  {
            pNode = Abc_NtkPo( pNtkOld, i );
            if strstr( Abc_ObjName( pNode ), "assume_fair" ) != nil {
                if strMatch == oldIndex {
                    dummyStr = fmt.Sprintf("%s__%s", Abc_ObjName( pNode ), "FAIRNESS");
                    return dummyStr;
                
                } else {
                    strMatch++; }
            }
        }
    
    } else {
        return "UNKNOWN"; }
    return nil;
}

var vecPis, vecPiNames *Vec_Ptr_t;
var vecLos, vecLoNames *Vec_Ptr_t;


func Aig_ManCiCleanupBiere(p *Aig_Man_t) int32 {
    var nPisOld int32  = Aig_ManCiNum(p);
    
    p.nObjs[AIG_OBJ_CI] = Vec_PtrSize( p.vCis );
    if Aig_ManRegNum(p) {
        p.nTruePis = Aig_ManCiNum(p) - Aig_ManRegNum(p); }
    
    return nPisOld - Aig_ManCiNum(p);
}


func Aig_ManCoCleanupBiere(p *Aig_Man_t) int32 {
    var nPosOld int32  = Aig_ManCoNum(p);

    p.nObjs[AIG_OBJ_CO] = Vec_PtrSize( p.vCos );
    if Aig_ManRegNum(p) {
        p.nTruePos = Aig_ManCoNum(p) - Aig_ManRegNum(p); }
    return nPosOld - Aig_ManCoNum(p);
}

func LivenessToSafetyTransformationSim(pNtk *Abc_Ntk_t, p *Aig_Man_t, vLive *Vec_Ptr_t, vFair *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var i, nRegCount int32;
    var pObjSavePi *Aig_Obj_t;
    var pObjSavedLo, pObjSavedLi *Aig_Obj_t;
    var pObj, pMatch *Aig_Obj_t;
    var pObjSaveOrSaved, pObjSavedLoAndEquality *Aig_Obj_t;
    var pObjShadowLo, pObjShadowLi, pObjShadowLiDriver *Aig_Obj_t;
    var pObjXor, pObjXnor, pObjAndAcc, pObjAndAccDummy *Aig_Obj_t;
    var pObjLive, pObjFair, pObjSafetyGate *Aig_Obj_t;
    var pObjSafetyPropertyOutput *Aig_Obj_t;
    var pDriverImage *Aig_Obj_t;
    var nodeName string;
    var piCopied, liCopied, loCopied, liCreated, loCreated, liveLatch, fairLatch int32;
    
    vecPis = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);
    vecPiNames = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);

    vecLos = Vec_PtrAlloc( Saig_ManRegNum( p )*2 + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );
    vecLoNames = Vec_PtrAlloc( Saig_ManRegNum( p )*2 + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );

#ifdef DUPLICATE_CKT_DEBUG
    printf("\nCode is compiled in DEBUG mode, the input-output behavior will be the same as the original circuit\n");
    printf("Press any key to continue...");
    scanf("%c", &c);
#endif

    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNew = Aig_ManStart( 2 * Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( "live2safe" );
    pNew.pSpec = nil;
    
    //****************************************************************
    // Step 2: map constant nodes
    //****************************************************************
    pObj = Aig_ManConst1( p );
    pObj.pData = Aig_ManConst1( pNew );

    //****************************************************************
    // Step 3: create true PIs
    //****************************************************************
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        piCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecPis, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkPi( pNtk, i ) ));
        Vec_PtrPush( vecPiNames, nodeName );
    }

    //****************************************************************
    // Step 4: create the special Pi corresponding to SAVE
    //****************************************************************
#ifndef DUPLICATE_CKT_DEBUG
    pObjSavePi = Aig_ObjCreateCi( pNew );
    nodeName = Abc_UtilStrsav("SAVE_BIERE"),
    Vec_PtrPush( vecPiNames, nodeName );
#endif
        
    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        loCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecLos, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ));
        Vec_PtrPush( vecLoNames, nodeName );
    }

    //****************************************************************
    // Step 6: create "saved" register output
    //****************************************************************
#ifndef DUPLICATE_CKT_DEBUG
    loCreated++;
    pObjSavedLo = Aig_ObjCreateCi( pNew );
    Vec_PtrPush( vecLos, pObjSavedLo );
    nodeName = Abc_UtilStrsav("SAVED_LO");
    Vec_PtrPush( vecLoNames, nodeName );
#endif

    //****************************************************************
    // Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
    //****************************************************************
#ifndef DUPLICATE_CKT_DEBUG
    pObjSaveOrSaved = Aig_Or( pNew, pObjSavePi, pObjSavedLo );
    //pObjSaveAndNotSaved = Aig_And( pNew, pObjSavePi, Aig_Not(pObjSavedLo) );
#endif

    //********************************************************************
    // Step 8: create internal nodes
    //********************************************************************
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }
    }

    //********************************************************************
    // Step 9: create the safety property output gate
    // create the safety property output gate, this will be the sole true PO 
    // of the whole circuit, discuss with Sat/Alan for an alternative implementation
    //********************************************************************
#ifndef DUPLICATE_CKT_DEBUG
    pObjSafetyPropertyOutput = Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Aig_ObjFanin0(pObj)->pData );
#endif

    //********************************************************************
    // DEBUG: To recreate the same circuit, at least from the input and output
    // behavior, we need to copy the original PO
    //********************************************************************
#ifdef DUPLICATE_CKT_DEBUG
    Saig_ManForEachPo( p, pObj, i )
    {
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
#endif

    // create register inputs for the original registers
    nRegCount = 0;
    
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        pMatch = Saig_ObjLoToLi( p, pObj );
        //Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pMatch) );
        Aig_ObjCreateCo( pNew, Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pMatch).pData, Aig_ObjFaninC0( pMatch ) ) );
        nRegCount++;
        liCopied++;
    }

    // create register input corresponding to the register "saved"
#ifndef DUPLICATE_CKT_DEBUG
    pObjSavedLi = Aig_ObjCreateCo( pNew, pObjSaveOrSaved );
    nRegCount++;
    liCreated++;

    pObjAndAcc = NULL;

    // create the family of shadow registers, then create the cascade of Xnor and And gates for the comparator 
    Saig_ManForEachLo( p, pObj, i )
    {
        pObjShadowLo = Aig_ObjCreateCi( pNew );

#ifdef PROPAGATE_NAMES
        Vec_PtrPush( vecLos, pObjShadowLo );
        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ) ) + 10 );
        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ), "SHADOW" );
        Vec_PtrPush( vecLoNames, nodeName );
#endif

        pObjShadowLiDriver = Aig_Mux( pNew, pObjSavePi, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
        pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
        nRegCount++;
        loCreated++; liCreated++;
        
        pObjXor = Aig_Exor( pNew, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
        pObjXnor = Aig_Not( pObjXor );
        if( pObjAndAcc == NULL )
            pObjAndAcc = pObjXnor;
        else
        {
            pObjAndAccDummy = pObjAndAcc;
            pObjAndAcc = Aig_And( pNew, pObjXnor, pObjAndAccDummy );
        }
    }

    // create the AND gate whose output will be the signal "looped"
    pObjSavedLoAndEquality = Aig_And( pNew, pObjSavedLo, pObjAndAcc );
    
    // create the master AND gate and corresponding AND and OR logic for the liveness properties
    pObjAndAcc = NULL;
    if( vLive == NULL || Vec_PtrSize( vLive ) == 0 )
        printf("\nCircuit without any liveness property\n");
    else
    {
        Vec_PtrForEachEntry( Aig_Obj_t *, vLive, pObj, i )
        {
            //assert( Aig_ObjIsNode( Aig_ObjChild0( pObj ) ) );
            //Aig_ObjPrint( pNew, pObj );
            liveLatch++;
            pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj ))->pData, Aig_ObjFaninC0(pObj));
            pObjShadowLo = Aig_ObjCreateCi( pNew );

#ifdef PROPAGATE_NAMES
        Vec_PtrPush( vecLos, pObjShadowLo );
        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "LIVENESS" );
        Vec_PtrPush( vecLoNames, nodeName );
#endif

            pObjShadowLiDriver = Aig_Or( pNew, Aig_Mux(pNew, pObjSavePi, Aig_Not(Aig_ManConst1(pNew)), pObjShadowLo), 
                                                Aig_And( pNew, pDriverImage, pObjSaveOrSaved ) );
            pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
            nRegCount++;
            loCreated++; liCreated++;
            
            if( pObjAndAcc == NULL )
                pObjAndAcc = pObjShadowLo;
            else
            {
                pObjAndAccDummy = pObjAndAcc;
                pObjAndAcc = Aig_And( pNew, pObjShadowLo, pObjAndAccDummy );
            }
        }
    }

    if( pObjAndAcc != NULL )
        pObjLive = pObjAndAcc;
    else
        pObjLive = Aig_ManConst1( pNew );
    
    // create the master AND gate and corresponding AND and OR logic for the fairness properties
    pObjAndAcc = NULL;
    if( vFair == NULL || Vec_PtrSize( vFair ) == 0 )
        printf("\nCircuit without any fairness property\n");
    else
    {
        Vec_PtrForEachEntry( Aig_Obj_t *, vFair, pObj, i )
        {
            fairLatch++;
            //assert( Aig_ObjIsNode( Aig_ObjChild0( pObj ) ) );
            pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj ))->pData, Aig_ObjFaninC0(pObj));
            pObjShadowLo = Aig_ObjCreateCi( pNew );

#ifdef PROPAGATE_NAMES
        Vec_PtrPush( vecLos, pObjShadowLo );
        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "FAIRNESS" );
        Vec_PtrPush( vecLoNames, nodeName );
#endif

            pObjShadowLiDriver = Aig_Or( pNew, Aig_Mux(pNew, pObjSavePi, Aig_Not(Aig_ManConst1(pNew)), pObjShadowLo), 
                                    Aig_And( pNew, pDriverImage, pObjSaveOrSaved ) );
            pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
            nRegCount++;
            loCreated++; liCreated++;
            
            if( pObjAndAcc == NULL )
                pObjAndAcc = pObjShadowLo;
            else
            {
                pObjAndAccDummy = pObjAndAcc;
                pObjAndAcc = Aig_And( pNew, pObjShadowLo, pObjAndAccDummy );
            }
        }
    }

    if( pObjAndAcc != NULL )
        pObjFair = pObjAndAcc;
    else
        pObjFair = Aig_ManConst1( pNew );
    
    //pObjSafetyGate = Aig_Exor( pNew, Aig_Not(Aig_ManConst1( pNew )), Aig_And( pNew, pObjSavedLoAndEquality, Aig_And( pNew, pObjFair, Aig_Not( pObjLive ) ) ) );
    pObjSafetyGate = Aig_And( pNew, pObjSavedLoAndEquality, Aig_And( pNew, pObjFair, Aig_Not( pObjLive ) ) );
    
    Aig_ObjPatchFanin0( pNew, pObjSafetyPropertyOutput, pObjSafetyGate );
#endif

    Aig_ManSetRegNum( pNew, nRegCount );

    Aig_ManCiCleanupBiere( pNew );
    Aig_ManCoCleanupBiere( pNew );
    
    Aig_ManCleanup( pNew );
    assert(Aig_ManCheck( pNew ), "Aig_ManCheck( pNew )");
    
#ifndef DUPLICATE_CKT_DEBUG
    assert((Aig_Obj_t *)Vec_PtrEntry(pNew->vCos, Saig_ManPoNum(pNew)+Aig_ObjCioId(pObjSavedLo)-Saig_ManPiNum(p)-1) == pObjSavedLi);
    assert( Saig_ManPoNum( pNew ) == 1 );
    assert( Saig_ManPiNum( p ) + 1 == Saig_ManPiNum( pNew ) );
    assert( Saig_ManRegNum( pNew ) == Saig_ManRegNum( p ) * 2 + 1 + liveLatch + fairLatch );
#endif

    return pNew;
}


func LivenessToSafetyTransformationOneStepLoopSim(pNtk *Abc_Ntk_t, p *Aig_Man_t, vLive *Vec_Ptr_t, vFair *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var i, nRegCount int32;
    var pObjSavePi *Aig_Obj_t;
    var pObj, pMatch *Aig_Obj_t;
    var pObjSavedLoAndEquality *Aig_Obj_t;
    var pObjXor, pObjXnor, pObjAndAcc, pObjAndAccDummy *Aig_Obj_t;
    var pObjLive, pObjFair, pObjSafetyGate *Aig_Obj_t;
    var pObjSafetyPropertyOutput *Aig_Obj_t;
    var pDriverImage *Aig_Obj_t;
    var pObjCorrespondingLi *Aig_Obj_t;


    var nodeName string;
    var piCopied, liCopied, loCopied int32;//, liCreated = 0, loCreated = 0, piVecIndex = 0;
    
    vecPis = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);
    vecPiNames = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);

    vecLos = Vec_PtrAlloc( Saig_ManRegNum( p )*2 + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );
    vecLoNames = Vec_PtrAlloc( Saig_ManRegNum( p )*2 + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );

    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNew = Aig_ManStart( 2 * Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( "live2safe" );
    pNew.pSpec = nil;
    
    //****************************************************************
    // Step 2: map constant nodes
    //****************************************************************
    pObj = Aig_ManConst1( p );
    pObj.pData = Aig_ManConst1( pNew );

    //****************************************************************
    // Step 3: create true PIs
    //****************************************************************
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        piCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecPis, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkPi( pNtk, i ) ));
        Vec_PtrPush( vecPiNames, nodeName );
    }

    //****************************************************************
    // Step 4: create the special Pi corresponding to SAVE
    //****************************************************************
    pObjSavePi = Aig_ObjCreateCi( pNew );
    nodeName = "SAVE_BIERE";
Vec_PtrPush( vecPiNames, nodeName );
        
    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        loCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecLos, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ));
        Vec_PtrPush( vecLoNames, nodeName );
    }

    //****************************************************************
    // Step 6: create "saved" register output
    //****************************************************************

     
                
                                          
                                       
                          
                                        
      

    //****************************************************************
    // Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
    //****************************************************************
     
                                                              
                                                                            
      

    //********************************************************************
    // Step 8: create internal nodes
    //********************************************************************
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }
    }

    //********************************************************************
    // Step 9: create the safety property output gate
    // create the safety property output gate, this will be the sole true PO 
    // of the whole circuit, discuss with Sat/Alan for an alternative implementation
    //********************************************************************

    pObjSafetyPropertyOutput = Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData );

    // create register inputs for the original registers
    nRegCount = 0;
    
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        pMatch = Saig_ObjLoToLi( p, pObj );
        //Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pMatch) );
        Aig_ObjCreateCo( pNew, Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pMatch).pData, Aig_ObjFaninC0( pMatch ) ) );
        nRegCount++;
        liCopied++;
    }

     
                                                                  
                                                           
                
                
      

    pObjAndAcc = nil;

    //****************************************************************************************************
    //For detection of loop of length 1 we do not need any shadow register, we only need equality detector
    //between Lo_j and Li_j and then a cascade of AND gates
    //****************************************************************************************************

    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        pObjCorrespondingLi = Saig_ObjLoToLi( p, pObj );
        
        pObjXor = Aig_Exor( pNew, (Aig_Obj_t *)pObj.pData,  Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0( pObjCorrespondingLi ).pData, Aig_ObjFaninC0( pObjCorrespondingLi ) ) );
        pObjXnor = Aig_Not( pObjXor );
        
        if pObjAndAcc == nil {
            pObjAndAcc = pObjXnor;
        } else {
            pObjAndAccDummy = pObjAndAcc;
            pObjAndAcc = Aig_And( pNew, pObjXnor, pObjAndAccDummy );
        }
    }

    // create the AND gate whose output will be the signal "looped"
    pObjSavedLoAndEquality = Aig_And( pNew, pObjSavePi, pObjAndAcc );
    
    // create the master AND gate and corresponding AND and OR logic for the liveness properties
    pObjAndAcc = nil;
    if vLive == nil || Vec_PtrSize( vLive ) == 0 {
        fmt.Printf("\nCircuit without any liveness property\n");
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vLive)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vLive, i)), 1); i++  {
            pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj )).pData, Aig_ObjFaninC0(pObj));
            if pObjAndAcc == nil {
                pObjAndAcc = pDriverImage;
            } else {
                pObjAndAccDummy = pObjAndAcc;
                pObjAndAcc = Aig_And( pNew, pDriverImage, pObjAndAccDummy );
            }
        }
    }

    if pObjAndAcc != nil {
        pObjLive = pObjAndAcc;
    } else {
        pObjLive = Aig_ManConst1( pNew ); }
    
    // create the master AND gate and corresponding AND and OR logic for the fairness properties
    pObjAndAcc = nil;
    if vFair == nil || Vec_PtrSize( vFair ) == 0 {
        fmt.Printf("\nCircuit without any fairness property\n");
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vFair)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vFair, i)), 1); i++  {
            pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj )).pData, Aig_ObjFaninC0(pObj));
            if pObjAndAcc == nil {
                pObjAndAcc = pDriverImage;
            } else {
                pObjAndAccDummy = pObjAndAcc;
                pObjAndAcc = Aig_And( pNew, pDriverImage, pObjAndAccDummy );
            }
        }
    }

    if pObjAndAcc != nil {
        pObjFair = pObjAndAcc;
    } else {
        pObjFair = Aig_ManConst1( pNew ); }
    
    pObjSafetyGate = Aig_And( pNew, pObjSavedLoAndEquality, Aig_And( pNew, pObjFair, Aig_Not( pObjLive ) ) );
    
    Aig_ObjPatchFanin0( pNew, pObjSafetyPropertyOutput, pObjSafetyGate );

    Aig_ManSetRegNum( pNew, nRegCount );

    fmt.Printf("\nSaig_ManPiNum = %d, Reg Num = %d, before everything, before Pi cleanup\n", Vec_PtrSize( pNew.vCis ), pNew.nRegs);

    Aig_ManCiCleanupBiere( pNew );
    Aig_ManCoCleanupBiere( pNew );

    Aig_ManCleanup( pNew );
    
    assert(Aig_ManCheck( pNew ), "Aig_ManCheck( pNew )");
    
    return pNew;
}



func populateLivenessVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) *Vec_Ptr_t {
    var pNode *Abc_Obj_t;
    var i, liveCounter int32;
    var vLive *Vec_Ptr_t;

    vLive = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pNode) = Abc_NtkPo(pNtk, i)), 1); i++  {
        if strstr( Abc_ObjName( pNode ), "assert_fair") != nil {
            Vec_PtrPush( vLive, Aig_ManCo( pAig, i ) );
            liveCounter++;
        }
    }
    fmt.Printf("\nNumber of liveness property found = %d\n", liveCounter);
    return vLive;
}

func populateFairnessVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) *Vec_Ptr_t {
    var pNode *Abc_Obj_t;
    var i, fairCounter int32;
    var vFair *Vec_Ptr_t;

    vFair = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pNode) = Abc_NtkPo(pNtk, i)), 1); i++  {
        if strstr( Abc_ObjName( pNode ), "assume_fair") != nil {
            Vec_PtrPush( vFair, Aig_ManCo( pAig, i ) );
            fairCounter++;
        }
    }
    fmt.Printf("\nNumber of fairness property found = %d\n", fairCounter);
    return vFair;
}

func updateNewNetworkNameManager(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, vPiNames *Vec_Ptr_t, vLoNames *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i, ntkObjId int32;

    pNtk.pManName = Nm_ManCreate( Abc_NtkCiNum( pNtk ) );

    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        ntkObjId = Abc_NtkCi( pNtk, i ).Id;
        //printf("Pi %d, Saved Name = %s, id = %d\n", i, Nm_ManStoreIdName( pNtk->pManName, ntkObjId, Aig_ObjType(pObj), Vec_PtrEntry(vPiNames, i), NULL ), ntkObjId);  
        Nm_ManStoreIdName( pNtk.pManName, ntkObjId, Aig_ObjType(pObj), (char *)Vec_PtrEntry(vPiNames, i), nil );
    }
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        ntkObjId = Abc_NtkCi( pNtk, Saig_ManPiNum( pAig ) + i ).Id;
        //printf("Lo %d, Saved name = %s, id = %d\n", i, Nm_ManStoreIdName( pNtk->pManName, ntkObjId, Aig_ObjType(pObj), Vec_PtrEntry(vLoNames, i), NULL ), ntkObjId);  
        Nm_ManStoreIdName( pNtk.pManName, ntkObjId, Aig_ObjType(pObj), (char *)Vec_PtrEntry(vLoNames, i), nil );
    }
}


func Abc_CommandAbcLivenessToSafetySim(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var pNtk, pNtkTemp, pNtkNew, pNtkOld *Abc_Ntk_t;
    var pAig, pAigNew *Aig_Man_t;
    var c int32;
    var vLive, vFair *Vec_Ptr_t;
                
    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        return 1;
    }

    if !Abc_NtkIsStrash( pNtk ) {
        fmt.Printf("\nThe input network was not strashed, strashing....\n");
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pNtkOld = pNtkTemp;
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkOld = pNtk;
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
    }

     
                              
                                                                         
                                                                          
                                                                           
                                                                                                 
                                                                                                
                                                                 
                                                                          
      

    c = Extra_UtilGetopt( argc, argv, "1" );
    if c == '1' {
        pAigNew = LivenessToSafetyTransformationOneStepLoopSim( pNtk, pAig, vLive, vFair );
    } else {
        pAigNew = LivenessToSafetyTransformationSim( pNtk, pAig, vLive, vFair ); }
    
     
                                 
                                                                         
                                                                             
                                                                              
                                                                                                       
                                                                                                      
                                                                    
                                                                          
      

    pNtkNew = Abc_NtkFromAigPhase( pAigNew );

    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkCreateCone(): Network check has failed.\n"); }
    
    updateNewNetworkNameManager( pNtkNew, pAigNew, vecPiNames,vecLoNames );
    Abc_FrameSetCurrentNetwork( pAbc, pNtkNew );

    //Saig_ManForEachPi( pAigNew, pObj, i )
    //    printf("Name of %d-th Pi = %s\n", i, retrieveTruePiName( pNtk, pAig, pAigNew, pObj ) );

    //Saig_ManForEachLo( pAigNew, pObj, i )
    //    printf("Name of %d-th Lo = %s\n", i, retrieveLOName( pNtk, pAig, pAigNew, pObj, vLive, vFair ) );

    //printVecPtrOfString( vecPiNames );
    //printVecPtrOfString( vecLoNames );

     
                           
                                         
                                                                                                                      
                                                                                                                                          

                                         
                                                                                                                                
          
      
        
    return 0;

}
//ABC_NAMESPACE_IMPL_END

