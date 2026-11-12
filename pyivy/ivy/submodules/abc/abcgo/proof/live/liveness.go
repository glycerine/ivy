package live

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [liveness.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Main implementation module.]

  Author      [Sayak Ray]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2009.]

  Revision    [$Id: liveness.c,v 1.00 2009/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include "base/main/main.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include <string.h>
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START

#define PROPAGATE_NAMES
#define MULTIPLE_LTL_FORMULA
#define ALLOW_SAFETY_PROPERTIES

const FULL_BIERE_MODE = 0
const IGNORE_LIVENESS_KEEP_SAFETY_MODE = 1
const IGNORE_SAFETY_KEEP_LIVENESS_MODE = 2
const IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE = 3
const FULL_BIERE_ONE_LOOP_MODE = 4
//#define DUPLICATE_CKT_DEBUG

// extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern Abc_Ntk_t * Abc_NtkFromAigPhase( Aig_Man_t * pMan );

//char *strdup(const char *string);

//******************************************
//external functions defined in ltl_parser.c
//******************************************
type ltlNode = ltlNode_t;
// extern ltlNode *readLtlFormula( char *formula );

// extern void traverseAbstractSyntaxTree( ltlNode *node );

// extern ltlNode *parseFormulaCreateAST( char *inputFormula );

// extern int isWellFormed( ltlNode *topNode );

// extern int checkSignalNameExistence( Abc_Ntk_t *pNtk, ltlNode *topASTNode );

// extern void populateBoolWithAigNodePtr( Abc_Ntk_t *pNtk, Aig_Man_t *pAigOld, Aig_Man_t *pAigNew, ltlNode *topASTNode );

// extern int checkAllBoolHaveAIGPointer( ltlNode *topASTNode );

// extern void populateAigPointerUnitGF( Aig_Man_t *pAigNew, ltlNode *topASTNode, Vec_Ptr_t *vSignal, Vec_Vec_t *vAigGFMap );

// extern void setAIGNodePtrOfGloballyNode( ltlNode *astNode, Aig_Obj_t *pObjLo );

// extern Aig_Obj_t *buildLogicFromLTLNode( Aig_Man_t *pAig, ltlNode *pLtlNode );

// extern Aig_Obj_t *retriveAIGPointerFromLTLNode( ltlNode *astNode );

// extern void traverseAbstractSyntaxTree_postFix( ltlNode *node );

//**********************************
//external function declaration ends
//**********************************

 
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


func nodeName_starts_with(pNode *Abc_Obj_t, prefix string) int32 {
    if strstr( Abc_ObjName( pNode ), prefix ) == Abc_ObjName( pNode ) {
        return 1;
    } else {
        return 0; }
}

func printVecPtrOfString(vec *Vec_Ptr_t) {
    var i int32;

    for  i=0; i< Vec_PtrSize( vec ); i++  {
        fmt.Printf("vec[%d] = %s\n", i, (char *)Vec_PtrEntry(vec, i));
    }
}

func getPoIndex(pAig *Aig_Man_t, pPivot *Aig_Obj_t) int32 {
    var i int32;
    var pObj *Aig_Obj_t;

    Saig_ManForEachPo( pAig, pObj, i )
    {
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
    Aig_ManForEachCi( pAigNew, pObj, index )
        if( pObj == pObjPivot )
            break;
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
    Saig_ManForEachLo( pAigNew, pObj, index )
        if( pObj == pObjPivot )
            break;
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
        dummyStr[0] = '\0';
        Saig_ManForEachPo( pAigOld, pObj, i )
        {
            pNode = Abc_NtkPo( pNtkOld, i );
            //if( strstr( Abc_ObjName( pNode ), "assert_fair" ) != NULL )
            if nodeName_starts_with( pNode, "assert_fair" ) != 0 {
                if strMatch == oldIndex {
                    dummyStr = fmt.Sprintf("%s__%s", Abc_ObjName( pNode ), "LIVENESS");
                    //return dummyStr;
                    break;
                
                } else {
                    strMatch++; }
            }
        }
        assert(dummyStr[0] != '\0', "dummyStr[0] != '\\0'");
        return dummyStr;
    
    } else if index >= 2 * originalLatchNum + 1 + Vec_PtrSize( vLive ) && index < 2 * originalLatchNum + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) {
        oldIndex = index - 2 * originalLatchNum - 1 - Vec_PtrSize( vLive );
        strMatch = 0;
        dummyStr[0] = '\0';
        Saig_ManForEachPo( pAigOld, pObj, i )
        {
            pNode = Abc_NtkPo( pNtkOld, i );
            //if( strstr( Abc_ObjName( pNode ), "assume_fair" ) != NULL )
            if nodeName_starts_with( pNode, "assume_fair" ) != 0 {
                if strMatch == oldIndex {
                    dummyStr = fmt.Sprintf("%s__%s", Abc_ObjName( pNode ), "FAIRNESS");
                    //return dummyStr;
                    break;
                
                } else {
                    strMatch++; }
            }
        }
        assert(dummyStr[0] != '\0', "dummyStr[0] != '\\0'");
        return dummyStr;
    
    } else {
        return "UNKNOWN"; }
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

func LivenessToSafetyTransformation(mode int32, pNtk *Abc_Ntk_t, p *Aig_Man_t, vLive *Vec_Ptr_t, vFair *Vec_Ptr_t, vAssertSafety *Vec_Ptr_t, vAssumeSafety *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var i, nRegCount int32;
    var pObjSavePi *Aig_Obj_t  = nil;
    var pObjSavedLo *Aig_Obj_t = nil;
    var pObjSavedLi *Aig_Obj_t = nil;
    var pObj, pMatch *Aig_Obj_t;
    var pObjSaveOrSaved *Aig_Obj_t = nil;
    var pObjSaveAndNotSaved *Aig_Obj_t = nil;
    var pObjSavedLoAndEquality *Aig_Obj_t;
    var pObjShadowLo, pObjShadowLi, pObjShadowLiDriver *Aig_Obj_t;
    var pObjXor, pObjXnor, pObjAndAcc *Aig_Obj_t;
    var pObjLive, pObjFair, pObjSafetyGate *Aig_Obj_t;
    var pObjSafetyPropertyOutput *Aig_Obj_t  = nil;
    var pObjOriginalSafetyPropertyOutput *Aig_Obj_t;
    var pDriverImage, pArgument, collectiveAssertSafety, collectiveAssumeSafety *Aig_Obj_t;
    var nodeName string;
    var piCopied, liCopied, loCopied, liCreated, loCreated, liveLatch, fairLatch int32;
    
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
    pNew.pName = (char *)malloc( len(pNtk.pName) + len("_l2s") + 1 );
    sprintf(pNew.pName, "%s_%s", pNtk.pName, "l2s");
    pNew.pSpec = nil;
    
    //****************************************************************
    // Step 2: map constant nodes
    //****************************************************************
    pObj = Aig_ManConst1( p );
    pObj.pData = Aig_ManConst1( pNew );

    //****************************************************************
    // Step 3: create true PIs
    //****************************************************************
    Saig_ManForEachPi( p, pObj, i )
    {
        piCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecPis, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkPi( pNtk, i ) ));
        Vec_PtrPush( vecPiNames, nodeName );
    }

    //****************************************************************
    // Step 4: create the special Pi corresponding to SAVE
    //****************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        pObjSavePi = Aig_ObjCreateCi( pNew );
        nodeName = "SAVE_BIERE";
Vec_PtrPush( vecPiNames, nodeName );
    }
        
    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
    Saig_ManForEachLo( p, pObj, i )
    {
        loCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecLos, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ));
        Vec_PtrPush( vecLoNames, nodeName );
    }

    //****************************************************************
    // Step 6: create "saved" register output
    //****************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        loCreated++;
        pObjSavedLo = Aig_ObjCreateCi( pNew );
        Vec_PtrPush( vecLos, pObjSavedLo );
        nodeName = "SAVED_LO";
        Vec_PtrPush( vecLoNames, nodeName );
    }

    //****************************************************************
    // Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
    //****************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        pObjSaveOrSaved = Aig_Or( pNew, pObjSavePi, pObjSavedLo );
        pObjSaveAndNotSaved = Aig_And( pNew, pObjSavePi, Aig_Not(pObjSavedLo) );
    }

    //********************************************************************
    // Step 8: create internal nodes
    //********************************************************************
    Aig_ManForEachNode( p, pObj, i )
    {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }

    
    //********************************************************************
    // Step 8.x : create PO for each safety assertions
    // NOTE : Here the output is purposely inverted as it will be thrown to 
    // dprove
    //********************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
        if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) == 0 {
            pObjAndAcc = Aig_ManConst1( pNew );
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
            {
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
            }
            pObjOriginalSafetyPropertyOutput = Aig_ObjCreateCo( pNew, Aig_Not(pObjAndAcc) );
        
        } else if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) != 0 {
            pObjAndAcc = Aig_ManConst1( pNew );
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
            {
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
            }
            collectiveAssertSafety = pObjAndAcc;

            pObjAndAcc = Aig_ManConst1( pNew );
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssumeSafety, pObj, i )
            {
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
            }
            collectiveAssumeSafety = pObjAndAcc;
            pObjOriginalSafetyPropertyOutput = Aig_ObjCreateCo( pNew, Aig_And( pNew, Aig_Not(collectiveAssertSafety), collectiveAssumeSafety ) );
        
        } else {
            fmt.Printf("WARNING!! No safety property is found, a new (negated) constant 1 output is created\n");
            pObjOriginalSafetyPropertyOutput = Aig_ObjCreateCo( pNew, Aig_Not( Aig_ManConst1(pNew) ) );
        }
    }

    //********************************************************************
    // Step 9: create the safety property output gate for the liveness properties
    // discuss with Sat/Alan for an alternative implementation
    //********************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        pObjSafetyPropertyOutput = Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData );
    }

    // create register inputs for the original registers
    nRegCount = 0;
    
    Saig_ManForEachLo( p, pObj, i )
    {
        pMatch = Saig_ObjLoToLi( p, pObj );
        Aig_ObjCreateCo( pNew, Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pMatch).pData, Aig_ObjFaninC0( pMatch ) ) );
        nRegCount++;
        liCopied++;
    }

    // create register input corresponding to the register "saved"
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        #ifndef DUPLICATE_CKT_DEBUG
            pObjSavedLi = Aig_ObjCreateCo( pNew, pObjSaveOrSaved );
            nRegCount++;
            liCreated++;

            //Changed on October 13, 2009
            //pObjAndAcc = NULL;
            pObjAndAcc = Aig_ManConst1( pNew );

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

                pObjShadowLiDriver = Aig_Mux( pNew, pObjSaveAndNotSaved, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
                pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
                nRegCount++;
                loCreated++; liCreated++;
        
                pObjXor = Aig_Exor( pNew, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
                pObjXnor = Aig_Not( pObjXor );
                
                pObjAndAcc = Aig_And( pNew, pObjXnor, pObjAndAcc );
            }

            // create the AND gate whose output will be the signal "looped"
            pObjSavedLoAndEquality = Aig_And( pNew, pObjSavedLo, pObjAndAcc );

            // create the master AND gate and corresponding AND and OR logic for the liveness properties
            pObjAndAcc = Aig_ManConst1( pNew );
            if( vLive == NULL || Vec_PtrSize( vLive ) == 0 )
            {
                printf("Circuit without any liveness property\n");
            }
            else
            {
                Vec_PtrForEachEntry( Aig_Obj_t *, vLive, pObj, i )
                {
                    liveLatch++;
                    pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj ))->pData, Aig_ObjFaninC0(pObj));
                    pObjShadowLo = Aig_ObjCreateCi( pNew );

                    #ifdef PROPAGATE_NAMES
                        Vec_PtrPush( vecLos, pObjShadowLo );
                        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
                        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "LIVENESS" );
                        Vec_PtrPush( vecLoNames, nodeName );
                    #endif

                    pObjShadowLiDriver = Aig_Or( pNew, pObjShadowLo, Aig_And( pNew, pDriverImage, pObjSaveOrSaved ) );
                    pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
                    nRegCount++;
                    loCreated++; liCreated++;
            
                    pObjAndAcc = Aig_And( pNew, pObjShadowLo, pObjAndAcc );
                }
            }

            pObjLive = pObjAndAcc;
                
            pObjAndAcc = Aig_ManConst1( pNew );
            if( vFair == NULL || Vec_PtrSize( vFair ) == 0 )
                printf("Circuit without any fairness property\n");
            else
            {
                Vec_PtrForEachEntry( Aig_Obj_t *, vFair, pObj, i )
                {
                    fairLatch++;
                    pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj ))->pData, Aig_ObjFaninC0(pObj));
                    pObjShadowLo = Aig_ObjCreateCi( pNew );

                    #ifdef PROPAGATE_NAMES
                        Vec_PtrPush( vecLos, pObjShadowLo );
                        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
                        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "FAIRNESS" );
                        Vec_PtrPush( vecLoNames, nodeName );
                    #endif

                    pObjShadowLiDriver = Aig_Or( pNew, pObjShadowLo, Aig_And( pNew, pDriverImage, pObjSaveOrSaved ) );
                    pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
                    nRegCount++;
                    loCreated++; liCreated++;
            
                    pObjAndAcc = Aig_And( pNew, pObjShadowLo, pObjAndAcc );
                }
            }

            pObjFair = pObjAndAcc;
                
            //pObjSafetyGate = Aig_Exor( pNew, Aig_Not(Aig_ManConst1( pNew )), Aig_And( pNew, pObjSavedLoAndEquality, Aig_And( pNew, pObjFair, Aig_Not( pObjLive ) ) ) );
            //Following is the actual Biere translation
            pObjSafetyGate = Aig_And( pNew, pObjSavedLoAndEquality, Aig_And( pNew, pObjFair, Aig_Not( pObjLive ) ) );

            Aig_ObjPatchFanin0( pNew, pObjSafetyPropertyOutput, pObjSafetyGate );
        #endif
    }

    Aig_ManSetRegNum( pNew, nRegCount );

    Aig_ManCiCleanupBiere( pNew );
    Aig_ManCoCleanupBiere( pNew );
    
    Aig_ManCleanup( pNew );
    
    assert(Aig_ManCheck( pNew ), "Aig_ManCheck( pNew )");
    
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            assert((Aig_Obj_t *)Vec_PtrEntry(pNew.vCos, Saig_ManPoNum(pNew)+Aig_ObjCioId(pObjSavedLo)-Saig_ManPiNum(p)-1) == pObjSavedLi, "(Aig_Obj_t *)Vec_PtrEntry(pNew.vCos, Saig_ManPoNum(pNew)+Aig_ObjCioId(pObjSavedLo)-Saig_ManPiNum(p)-1) == pObjSavedLi");
            assert(Saig_ManPiNum( p ) + 1 == Saig_ManPiNum( pNew ), "Saig_ManPiNum( p ) + 1 == Saig_ManPiNum( pNew )");
            assert(Saig_ManRegNum( pNew ) == Saig_ManRegNum( p ) * 2 + 1 + liveLatch + fairLatch, "Saig_ManRegNum( pNew ) == Saig_ManRegNum( p ) * 2 + 1 + liveLatch + fairLatch");
    }

    return pNew;
}





func LivenessToSafetyTransformationAbs(mode int32, pNtk *Abc_Ntk_t, p *Aig_Man_t, vFlops *Vec_Int_t, vLive *Vec_Ptr_t, vFair *Vec_Ptr_t, vAssertSafety *Vec_Ptr_t, vAssumeSafety *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var i, nRegCount, iEntry int32;
    var pObjSavePi *Aig_Obj_t  = nil;
    var pObjSavedLi *Aig_Obj_t = nil;
    var pObjSavedLo *Aig_Obj_t = nil;
    var pObj, pMatch *Aig_Obj_t;
    var pObjSavedLoAndEquality *Aig_Obj_t;
    var pObjSaveOrSaved *Aig_Obj_t = nil;
    var pObjSaveAndNotSaved *Aig_Obj_t = nil;
    var pObjShadowLo, pObjShadowLi, pObjShadowLiDriver *Aig_Obj_t;
    var pObjXor, pObjXnor, pObjAndAcc *Aig_Obj_t;
    var pObjLive, pObjFair, pObjSafetyGate *Aig_Obj_t;
    var pObjSafetyPropertyOutput *Aig_Obj_t  = nil;
    var pDriverImage, pArgument, collectiveAssertSafety, collectiveAssumeSafety *Aig_Obj_t;
    var nodeName string;
    var piCopied, liCopied, loCopied, liCreated, loCreated, liveLatch, fairLatch int32;//, piVecIndex = 0;
    
    vecPis = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);
    vecPiNames = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);

    vecLos = Vec_PtrAlloc( Saig_ManRegNum( p ) + Vec_IntSize( vFlops ) + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );
    vecLoNames = Vec_PtrAlloc( Saig_ManRegNum( p ) + Vec_IntSize( vFlops ) + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );

    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNew = Aig_ManStart( 2 * Aig_ManObjNumMax(p) );
    pNew.pName = (char *)malloc( len(pNtk.pName) + len("_l2s") + 1 );
    sprintf(pNew.pName, "%s_%s", pNtk.pName, "l2s");
    pNew.pSpec = nil;
    
    //****************************************************************
    // Step 2: map constant nodes
    //****************************************************************
    pObj = Aig_ManConst1( p );
    pObj.pData = Aig_ManConst1( pNew );

    //****************************************************************
    // Step 3: create true PIs
    //****************************************************************
    Saig_ManForEachPi( p, pObj, i )
    {
        piCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecPis, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkPi( pNtk, i ) ));
        Vec_PtrPush( vecPiNames, nodeName );
    }

    //****************************************************************
    // Step 4: create the special Pi corresponding to SAVE
    //****************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        pObjSavePi = Aig_ObjCreateCi( pNew );
        nodeName = "SAVE_BIERE";
Vec_PtrPush( vecPiNames, nodeName );
    }
        
    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
    Saig_ManForEachLo( p, pObj, i )
    {
        loCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecLos, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ));
        Vec_PtrPush( vecLoNames, nodeName );
    }

    //****************************************************************
    // Step 6: create "saved" register output
    //****************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        loCreated++;
        pObjSavedLo = Aig_ObjCreateCi( pNew );
        Vec_PtrPush( vecLos, pObjSavedLo );
        nodeName = "SAVED_LO";
        Vec_PtrPush( vecLoNames, nodeName );
    }

    //****************************************************************
    // Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
    //****************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        pObjSaveOrSaved = Aig_Or( pNew, pObjSavePi, pObjSavedLo );
        pObjSaveAndNotSaved = Aig_And( pNew, pObjSavePi, Aig_Not(pObjSavedLo) );
    }

    //********************************************************************
    // Step 8: create internal nodes
    //********************************************************************
    Aig_ManForEachNode( p, pObj, i )
    {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }

    
    //********************************************************************
    // Step 8.x : create PO for each safety assertions
    // NOTE : Here the output is purposely inverted as it will be thrown to 
    // dprove
    //********************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
        if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) == 0 {
            pObjAndAcc = Aig_ManConst1( pNew );
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
            {
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
            }
            Aig_ObjCreateCo( pNew, Aig_Not(pObjAndAcc) );
        
        } else if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) != 0 {
            pObjAndAcc = Aig_ManConst1( pNew );
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
            {
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
            }
            collectiveAssertSafety = pObjAndAcc;

            pObjAndAcc = Aig_ManConst1( pNew );
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssumeSafety, pObj, i )
            {
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
            }
            collectiveAssumeSafety = pObjAndAcc;
            Aig_ObjCreateCo( pNew, Aig_And( pNew, Aig_Not(collectiveAssertSafety), collectiveAssumeSafety ) );
        
        } else {
            fmt.Printf("WARNING!! No safety property is found, a new (negated) constant 1 output is created\n");
            Aig_ObjCreateCo( pNew, Aig_Not( Aig_ManConst1(pNew) ) );
        }
    }

    //********************************************************************
    // Step 9: create the safety property output gate for the liveness properties
    // discuss with Sat/Alan for an alternative implementation
    //********************************************************************
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        pObjSafetyPropertyOutput = Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData );
    }

    // create register inputs for the original registers
    nRegCount = 0;
    
    Saig_ManForEachLo( p, pObj, i )
    {
        pMatch = Saig_ObjLoToLi( p, pObj );
        Aig_ObjCreateCo( pNew, Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pMatch).pData, Aig_ObjFaninC0( pMatch ) ) );
        nRegCount++;
        liCopied++;
    }

    // create register input corresponding to the register "saved"
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        #ifndef DUPLICATE_CKT_DEBUG
            pObjSavedLi = Aig_ObjCreateCo( pNew, pObjSaveOrSaved );
            nRegCount++;
            liCreated++;

            //Changed on October 13, 2009
            //pObjAndAcc = NULL;
            pObjAndAcc = Aig_ManConst1( pNew );

    // create the family of shadow registers, then create the cascade of Xnor and And gates for the comparator 
            //Saig_ManForEachLo( p, pObj, i )
            Saig_ManForEachLo( p, pObj, i )
            {
                printf("Flop[%d] = %s\n", i, Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ) );
            }
            Vec_IntForEachEntry( vFlops, iEntry, i )
            {
                pObjShadowLo = Aig_ObjCreateCi( pNew );
                pObj = Aig_ManLo( p, iEntry );

                #ifdef PROPAGATE_NAMES
                    Vec_PtrPush( vecLos, pObjShadowLo );
                    nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + iEntry ) ) ) + 10 );
                    sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + iEntry ) ), "SHADOW" );
                    printf("Flop copied [%d] = %s\n", iEntry, nodeName );
                    Vec_PtrPush( vecLoNames, nodeName );
                #endif

                pObjShadowLiDriver = Aig_Mux( pNew, pObjSaveAndNotSaved, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
                pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
                nRegCount++;
                loCreated++; liCreated++;
        
                pObjXor = Aig_Exor( pNew, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
                pObjXnor = Aig_Not( pObjXor );
                
                pObjAndAcc = Aig_And( pNew, pObjXnor, pObjAndAcc );
            }

            // create the AND gate whose output will be the signal "looped"
            pObjSavedLoAndEquality = Aig_And( pNew, pObjSavedLo, pObjAndAcc );

            // create the master AND gate and corresponding AND and OR logic for the liveness properties
            pObjAndAcc = Aig_ManConst1( pNew );
            if( vLive == NULL || Vec_PtrSize( vLive ) == 0 )
            {
                printf("Circuit without any liveness property\n");
            }
            else
            {
                Vec_PtrForEachEntry( Aig_Obj_t *, vLive, pObj, i )
                {
                    liveLatch++;
                    pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj ))->pData, Aig_ObjFaninC0(pObj));
                    pObjShadowLo = Aig_ObjCreateCi( pNew );

                    #ifdef PROPAGATE_NAMES
                        Vec_PtrPush( vecLos, pObjShadowLo );
                        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
                        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "LIVENESS" );
                        Vec_PtrPush( vecLoNames, nodeName );
                    #endif

                    pObjShadowLiDriver = Aig_Or( pNew, pObjShadowLo, Aig_And( pNew, pDriverImage, pObjSaveOrSaved ) );
                    pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
                    nRegCount++;
                    loCreated++; liCreated++;
            
                    pObjAndAcc = Aig_And( pNew, pObjShadowLo, pObjAndAcc );
                }
            }

            pObjLive = pObjAndAcc;
                
            pObjAndAcc = Aig_ManConst1( pNew );
            if( vFair == NULL || Vec_PtrSize( vFair ) == 0 )
                printf("Circuit without any fairness property\n");
            else
            {
                Vec_PtrForEachEntry( Aig_Obj_t *, vFair, pObj, i )
                {
                    fairLatch++;
                    pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj ))->pData, Aig_ObjFaninC0(pObj));
                    pObjShadowLo = Aig_ObjCreateCi( pNew );

                    #ifdef PROPAGATE_NAMES
                        Vec_PtrPush( vecLos, pObjShadowLo );
                        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
                        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "FAIRNESS" );
                        Vec_PtrPush( vecLoNames, nodeName );
                    #endif

                    pObjShadowLiDriver = Aig_Or( pNew, pObjShadowLo, Aig_And( pNew, pDriverImage, pObjSaveOrSaved ) );
                    pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
                    nRegCount++;
                    loCreated++; liCreated++;
            
                    pObjAndAcc = Aig_And( pNew, pObjShadowLo, pObjAndAcc );
                }
            }

            pObjFair = pObjAndAcc;
                
            //pObjSafetyGate = Aig_Exor( pNew, Aig_Not(Aig_ManConst1( pNew )), Aig_And( pNew, pObjSavedLoAndEquality, Aig_And( pNew, pObjFair, Aig_Not( pObjLive ) ) ) );
            //Following is the actual Biere translation
            pObjSafetyGate = Aig_And( pNew, pObjSavedLoAndEquality, Aig_And( pNew, pObjFair, Aig_Not( pObjLive ) ) );

            Aig_ObjPatchFanin0( pNew, pObjSafetyPropertyOutput, pObjSafetyGate );
        #endif
    }

    Aig_ManSetRegNum( pNew, nRegCount );

    Aig_ManCiCleanupBiere( pNew );
    Aig_ManCoCleanupBiere( pNew );
    
    Aig_ManCleanup( pNew );
    
    assert(Aig_ManCheck( pNew ), "Aig_ManCheck( pNew )");
    
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            assert((Aig_Obj_t *)Vec_PtrEntry(pNew.vCos, Saig_ManPoNum(pNew)+Aig_ObjCioId(pObjSavedLo)-Saig_ManPiNum(p)-1) == pObjSavedLi, "(Aig_Obj_t *)Vec_PtrEntry(pNew.vCos, Saig_ManPoNum(pNew)+Aig_ObjCioId(pObjSavedLo)-Saig_ManPiNum(p)-1) == pObjSavedLi");
            assert(Saig_ManPiNum( p ) + 1 == Saig_ManPiNum( pNew ), "Saig_ManPiNum( p ) + 1 == Saig_ManPiNum( pNew )");
            assert(Saig_ManRegNum( pNew ) == Saig_ManRegNum( p ) + Vec_IntSize( vFlops ) + 1 + liveLatch + fairLatch, "Saig_ManRegNum( pNew ) == Saig_ManRegNum( p ) + Vec_IntSize( vFlops ) + 1 + liveLatch + fairLatch");
    }

    return pNew;
}



func LivenessToSafetyTransformationOneStepLoop(mode int32, pNtk *Abc_Ntk_t, p *Aig_Man_t, vLive *Vec_Ptr_t, vFair *Vec_Ptr_t, vAssertSafety *Vec_Ptr_t, vAssumeSafety *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var i, nRegCount int32;
    var pObjSavePi *Aig_Obj_t  = nil;
    var pObj, pMatch *Aig_Obj_t;
    var pObjSavedLoAndEquality *Aig_Obj_t;
    var pObjXor, pObjXnor, pObjAndAcc, pObjAndAccDummy *Aig_Obj_t;
    var pObjLive, pObjFair, pObjSafetyGate *Aig_Obj_t;
    var pObjSafetyPropertyOutput *Aig_Obj_t  = nil;
    var pDriverImage *Aig_Obj_t;
    var pObjCorrespondingLi *Aig_Obj_t;
    var pArgument *Aig_Obj_t;
    var collectiveAssertSafety, collectiveAssumeSafety *Aig_Obj_t;

    var nodeName string;
    var piCopied, liCopied, loCopied int32;//, liCreated = 0, loCreated = 0, piVecIndex = 0;

    if Aig_ManRegNum( p ) == 0 {
        fmt.Printf("The input AIG contains no register, returning the original AIG as it is\n");
        return p;
    }

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
    Saig_ManForEachPi( p, pObj, i )
    {
        piCopied++;
        pObj.pData = Aig_ObjCreateCi(pNew);
        Vec_PtrPush( vecPis, pObj.pData );
        nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkPi( pNtk, i ) ));
        Vec_PtrPush( vecPiNames, nodeName );
    }

    //****************************************************************
    // Step 4: create the special Pi corresponding to SAVE
    //****************************************************************
    if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE {
        pObjSavePi = Aig_ObjCreateCi( pNew );
        nodeName = "SAVE_BIERE";
Vec_PtrPush( vecPiNames, nodeName );
    }
            
    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
    Saig_ManForEachLo( p, pObj, i )
    {
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
    Aig_ManForEachNode( p, pObj, i )
    {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }

     
                                                                          
                                                      
                                                                          
                                                              
     
                                                                                                                
     
      

    if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
        if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) == 0 {
            pObjAndAcc = nil;
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
            {
                //pObj->pData = Aig_ObjCreateCo( pNew, Aig_NotCond(Aig_ObjFanin0(pObj)->pData, Aig_ObjFaninC0( pObj ) ) );
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                if pObjAndAcc == nil {
                    pObjAndAcc = pArgument;
                } else {
                    pObjAndAccDummy = pObjAndAcc;
                    pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAccDummy );
                }
            }
            Aig_ObjCreateCo( pNew, Aig_Not(pObjAndAcc) );
        
        } else if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) != 0 {
            pObjAndAcc = nil;
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
            {
                //pObj->pData = Aig_ObjCreateCo( pNew, Aig_NotCond(Aig_ObjFanin0(pObj)->pData, Aig_ObjFaninC0( pObj ) ) );
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                if pObjAndAcc == nil {
                    pObjAndAcc = pArgument;
                } else {
                    pObjAndAccDummy = pObjAndAcc;
                    pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAccDummy );
                }
            }
            collectiveAssertSafety = pObjAndAcc;
            pObjAndAcc = nil;
            Vec_PtrForEachEntry( Aig_Obj_t *, vAssumeSafety, pObj, i )
            {
                //pObj->pData = Aig_ObjCreateCo( pNew, Aig_NotCond(Aig_ObjFanin0(pObj)->pData, Aig_ObjFaninC0( pObj ) ) );
                pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                if pObjAndAcc == nil {
                    pObjAndAcc = pArgument;
                } else {
                    pObjAndAccDummy = pObjAndAcc;
                    pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAccDummy );
                }
            }
            collectiveAssumeSafety = pObjAndAcc;
            Aig_ObjCreateCo( pNew, Aig_And( pNew, Aig_Not(collectiveAssertSafety), collectiveAssumeSafety ) );
        
        } else {
            fmt.Printf("No safety property is specified, hence no safety gate is created\n"); }
    }

    //********************************************************************
    // Step 9: create the safety property output gate
    // create the safety property output gate, this will be the sole true PO 
    // of the whole circuit, discuss with Sat/Alan for an alternative implementation
    //********************************************************************

    if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE {
        pObjSafetyPropertyOutput = Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData );
    }

    // create register inputs for the original registers
    nRegCount = 0;
    
    Saig_ManForEachLo( p, pObj, i )
    {
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

    if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE {
        Saig_ManForEachLo( p, pObj, i )
        {
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
            fmt.Printf("Circuit without any liveness property\n");
        } else {
            Vec_PtrForEachEntry( Aig_Obj_t *, vLive, pObj, i )
            {
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
            fmt.Printf("Circuit without any fairness property\n");
        } else {
            Vec_PtrForEachEntry( Aig_Obj_t *, vFair, pObj, i )
            {
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
    }

    Aig_ManSetRegNum( pNew, nRegCount );

    //printf("\nSaig_ManPiNum = %d, Reg Num = %d, before everything, before Pi cleanup\n", Vec_PtrSize( pNew->vPis ), pNew->nRegs );

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
    Abc_NtkForEachPo( pNtk, pNode, i )
        //if( strstr( Abc_ObjName( pNode ), "assert_fair") != NULL )
        if nodeName_starts_with( pNode, "assert_fair" ) != 0 {
            Vec_PtrPush( vLive, Aig_ManCo( pAig, i ) );
            liveCounter++;
        }
    fmt.Printf("Number of liveness property found = %d\n", liveCounter);
    return vLive;
}

func populateFairnessVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) *Vec_Ptr_t {
    var pNode *Abc_Obj_t;
    var i, fairCounter int32;
    var vFair *Vec_Ptr_t;

    vFair = Vec_PtrAlloc( 100 );
    Abc_NtkForEachPo( pNtk, pNode, i )
        //if( strstr( Abc_ObjName( pNode ), "assume_fair") != NULL )
        if nodeName_starts_with( pNode, "assume_fair" ) != 0 {
            Vec_PtrPush( vFair, Aig_ManCo( pAig, i ) );
            fairCounter++;
        }
    fmt.Printf("Number of fairness property found = %d\n", fairCounter);
    return vFair;
}

func populateSafetyAssertionVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) *Vec_Ptr_t {
    var pNode *Abc_Obj_t;
    var i, assertSafetyCounter int32;
    var vAssertSafety *Vec_Ptr_t;

    vAssertSafety = Vec_PtrAlloc( 100 );
    Abc_NtkForEachPo( pNtk, pNode, i )
        //if( strstr( Abc_ObjName( pNode ), "Assert") != NULL )
        if (nodeName_starts_with( pNode, "assert_safety" ) != 0) || (nodeName_starts_with( pNode, "Assert" ) != 0) {
            Vec_PtrPush( vAssertSafety, Aig_ManCo( pAig, i ) );
            assertSafetyCounter++;
        }
    fmt.Printf("Number of safety property found = %d\n", assertSafetyCounter);
    return vAssertSafety;
}

func populateSafetyAssumptionVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) *Vec_Ptr_t {
    var pNode *Abc_Obj_t;
    var i, assumeSafetyCounter int32;
    var vAssumeSafety *Vec_Ptr_t;

    vAssumeSafety = Vec_PtrAlloc( 100 );
    Abc_NtkForEachPo( pNtk, pNode, i )
        //if( strstr( Abc_ObjName( pNode ), "Assert") != NULL )
        if (nodeName_starts_with( pNode, "assume_safety" ) != 0) || (nodeName_starts_with( pNode, "Assume" ) != 0) {
            Vec_PtrPush( vAssumeSafety, Aig_ManCo( pAig, i ) );
            assumeSafetyCounter++;
        }
    fmt.Printf("Number of assume_safety property found = %d\n", assumeSafetyCounter);
    return vAssumeSafety;
}

func updateNewNetworkNameManager(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, vPiNames *Vec_Ptr_t, vLoNames *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var pNode *Abc_Obj_t;
    var i, ntkObjId int32;

    pNtk.pManName = Nm_ManCreate( Abc_NtkCiNum( pNtk ) );

    if vPiNames != nil {
        Saig_ManForEachPi( pAig, pObj, i )
        {
            ntkObjId = Abc_NtkCi( pNtk, i ).Id;
            //printf("Pi %d, Saved Name = %s, id = %d\n", i, Nm_ManStoreIdName( pNtk->pManName, ntkObjId, Aig_ObjType(pObj), Vec_PtrEntry(vPiNames, i), NULL ), ntkObjId);  
            Nm_ManStoreIdName( pNtk.pManName, ntkObjId, Aig_ObjType(pObj), (char *)Vec_PtrEntry(vPiNames, i), nil );
        }
    }
    if vLoNames != nil {
        Saig_ManForEachLo( pAig, pObj, i )
        {
            ntkObjId = Abc_NtkCi( pNtk, Saig_ManPiNum( pAig ) + i ).Id;
            //printf("Lo %d, Saved name = %s, id = %d\n", i, Nm_ManStoreIdName( pNtk->pManName, ntkObjId, Aig_ObjType(pObj), Vec_PtrEntry(vLoNames, i), NULL ), ntkObjId);  
            Nm_ManStoreIdName( pNtk.pManName, ntkObjId, Aig_ObjType(pObj), (char *)Vec_PtrEntry(vLoNames, i), nil );
        }
    }

    Abc_NtkForEachPo(pNtk, pNode, i)
    {
        Abc_ObjAssignName(pNode, "assert_safety_", Abc_ObjName(pNode) );
    }

    // assign latch input names
    Abc_NtkForEachLatch(pNtk, pNode, i)
        if ( Nm_ManFindNameById(pNtk.pManName, Abc_ObjFanin0(pNode).Id) == nil )
            Abc_ObjAssignName( Abc_ObjFanin0(pNode), Abc_ObjName(Abc_ObjFanin0(pNode)), nil );
}


func Abc_CommandAbcLivenessToSafety(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var pNtk, pNtkTemp, pNtkNew, pNtkOld *Abc_Ntk_t;
    var pAig *Aig_Man_t;
    var pAigNew *Aig_Man_t = nil;
    var c int32;
    var vLive, vFair, vAssertSafety, vAssumeSafety *Vec_Ptr_t;
    var directive int32  = -1;
                
    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    if argc == 1 {
        assert(directive == -1, "directive == -1");
        directive = FULL_BIERE_MODE;
    
    } else {
        Extra_UtilGetoptReset();
        for  ( c = Extra_UtilGetopt( argc, argv, "1slh" ) ) != -1  {
            switch  c  {
            case '1': 
                if( directive == -1 )
                    directive = FULL_BIERE_ONE_LOOP_MODE;
                else {
                    assert(directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE || directive == IGNORE_SAFETY_KEEP_LIVENESS_MODE, "directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE || directive == IGNORE_SAFETY_KEEP_LIVENESS_MODE");
                    if directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
                        directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                    } else {
                        directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE; }
                }
                break;
            case 's':
                if( directive == -1 )
                    directive = IGNORE_SAFETY_KEEP_LIVENESS_MODE;
                else {
                    if directive != FULL_BIERE_ONE_LOOP_MODE {
                        goto usage; }
                    assert(directive == FULL_BIERE_ONE_LOOP_MODE, "directive == FULL_BIERE_ONE_LOOP_MODE");
                    directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE;
                }
                break;
            case 'l':
                if( directive == -1 )
                    directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                else {
                    if directive != FULL_BIERE_ONE_LOOP_MODE {
                        goto usage; }
                    assert(directive == FULL_BIERE_ONE_LOOP_MODE, "directive == FULL_BIERE_ONE_LOOP_MODE");
                    directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                }
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
            }
        }
    }

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        return 1;
    }
    if !Abc_NtkIsStrash( pNtk ) {
        fmt.Printf("The input network was not strashed, strashing....\n");
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pNtkOld = pNtkTemp;
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
        vAssertSafety = populateSafetyAssertionVector( pNtk, pAig );
        vAssumeSafety = populateSafetyAssumptionVector( pNtk, pAig );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkOld = pNtk;
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
        vAssertSafety = populateSafetyAssertionVector( pNtk, pAig );
        vAssumeSafety = populateSafetyAssumptionVector( pNtk, pAig );
    }

    switch  directive  {
    case FULL_BIERE_MODE:
        //if( Vec_PtrSize(vLive) == 0 && Vec_PtrSize(vAssertSafety) == 0 )
        //{
        //    printf("Input circuit has NO safety and NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformation( FULL_BIERE_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case FULL_BIERE_ONE_LOOP_MODE:
        //if( Vec_PtrSize(vLive) == 0 && Vec_PtrSize(vAssertSafety) == 0 )
        //{
        //    printf("Input circuit has NO safety and NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformationOneStepLoop( FULL_BIERE_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case IGNORE_LIVENESS_KEEP_SAFETY_MODE:
        //if( Vec_PtrSize(vAssertSafety) == 0 )
        //{    
        //    printf("Input circuit has NO safety property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformation( IGNORE_LIVENESS_KEEP_SAFETY_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t1 PO - only for safety property; liveness properties are ignored, if any.\n\tno additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case IGNORE_SAFETY_KEEP_LIVENESS_MODE:
        //if( Vec_PtrSize(vLive) == 0 )
        //{    
        //    printf("Input circuit has NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformation( IGNORE_SAFETY_KEEP_LIVENESS_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t1 PO - only for liveness property; safety properties are ignored, if any.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE:
        //if( Vec_PtrSize(vLive) == 0 )
        //{
        //    printf("Input circuit has NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformationOneStepLoop( IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("New circuit is produced ignoring safety outputs!\nOnly liveness and fairness outputs are considered.\nShadow registers are not created\n"); }
            break;
        //}
    }

     
                   
     
                                                                                                                            
                                         
                                                                                                                                 
     
         
     
                                
                                                   
                      
         
                               
             
                                                    
                         
             
                                          
             
                                                                              
                                                          
                                   
                                                      
                                                             
                                                             
                                                                            
                                                                             
             
                
             
                                                  
                               
                                                             
                                                             
                                                                            
                                                                             
             
                                                                                                                          
         
                           
         
                               
             
                                                    
                         
             
                                          
             
                                                                              
                                                          
                                   
                                                      
                                                             
                                                             
                                                                            
                                                                             
             
                
             
                                                  
                               
                                                             
                                                             
                                                                            
                                                                             
             
                                                                                                                                                 
                                             
                                                                                                                                            
         
                           
         
                               
             
                                                    
                         
             
            
                                          
             
                                                                              
                                                          
                                   
                                                      
                                                             
                                                             
                                                                            
                                                                             
             
                
             
                                                  
                               
                                                             
                                                             
                                                                            
                                                                             
             
                                                                                                                                                 
                                             
                                                                                                                                                           
         
                           
                       
            
                       
     
      
    
     
                                 
                                                                         
                                                                             
                                                                              
                                                                                                       
                                                                                                      
                                                                    
                                                                          
      

    pNtkNew = Abc_NtkFromAigPhase( pAigNew );
    pNtkNew.pName = Abc_UtilStrsav( pAigNew.pName );
    
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkCreateCone(): Network check has failed.\n"); }
    
    updateNewNetworkNameManager( pNtkNew, pAigNew, vecPiNames, vecLoNames );
    Abc_FrameSetCurrentNetwork( pAbc, pNtkNew );

     
                           
                                         
                                                                                                                      
                                                                                                                                          

                                         
                                                                                                                                
          
      
        
    return 0;

usage:
    fmt.Fprintf(os.Stdout, "usage: l2s [-1lsh]\n");
    fmt.Fprintf(os.Stdout, "\t         performs Armin Biere's live-to-safe transformation\n");
    fmt.Fprintf(os.Stdout, "\t-1 : no shadow logic, presume all loops are self loops\n");
    fmt.Fprintf(os.Stdout, "\t-l : ignore liveness and fairness outputs\n");
    fmt.Fprintf(os.Stdout, "\t-s : ignore safety assertions and assumptions\n");
    fmt.Fprintf(os.Stdout, "\t-h : print command usage\n");
    return 1;
}

func prepareFlopVector(pAig *Aig_Man_t, vectorLength int32) *Vec_Int_t {
    var vFlops *Vec_Int_t;
    var i int32;

    vFlops = Vec_IntAlloc( vectorLength );

    for  i=0; i<vectorLength; i++  {
        Vec_IntPush( vFlops, i );
    }

     
                              
                              
                              
                              
                               
                               
                               
                               
                               
                               
                                
                                
                               
                               
      

    return vFlops;
}

func Abc_CommandAbcLivenessToSafetyAbstraction(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var pNtk, pNtkTemp, pNtkNew, pNtkOld *Abc_Ntk_t;
    var pAig *Aig_Man_t;
    var pAigNew *Aig_Man_t = nil;
    var c int32;
    var vLive, vFair, vAssertSafety, vAssumeSafety *Vec_Ptr_t;
    var directive int32  = -1;
    var vFlops *Vec_Int_t;
                
    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    if argc == 1 {
        assert(directive == -1, "directive == -1");
        directive = FULL_BIERE_MODE;
    
    } else {
        Extra_UtilGetoptReset();
        for  ( c = Extra_UtilGetopt( argc, argv, "1slh" ) ) != -1  {
            switch  c  {
            case '1': 
                if( directive == -1 )
                    directive = FULL_BIERE_ONE_LOOP_MODE;
                else {
                    assert(directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE || directive == IGNORE_SAFETY_KEEP_LIVENESS_MODE, "directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE || directive == IGNORE_SAFETY_KEEP_LIVENESS_MODE");
                    if directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
                        directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                    } else {
                        directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE; }
                }
                break;
            case 's':
                if( directive == -1 )
                    directive = IGNORE_SAFETY_KEEP_LIVENESS_MODE;
                else {
                    if directive != FULL_BIERE_ONE_LOOP_MODE {
                        goto usage; }
                    assert(directive == FULL_BIERE_ONE_LOOP_MODE, "directive == FULL_BIERE_ONE_LOOP_MODE");
                    directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE;
                }
                break;
            case 'l':
                if( directive == -1 )
                    directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                else {
                    if directive != FULL_BIERE_ONE_LOOP_MODE {
                        goto usage; }
                    assert(directive == FULL_BIERE_ONE_LOOP_MODE, "directive == FULL_BIERE_ONE_LOOP_MODE");
                    directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                }
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
            }
        }
    }

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        return 1;
    }
    if !Abc_NtkIsStrash( pNtk ) {
        fmt.Printf("The input network was not strashed, strashing....\n");
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pNtkOld = pNtkTemp;
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
        vAssertSafety = populateSafetyAssertionVector( pNtk, pAig );
        vAssumeSafety = populateSafetyAssumptionVector( pNtk, pAig );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkOld = pNtk;
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
        vAssertSafety = populateSafetyAssertionVector( pNtk, pAig );
        vAssumeSafety = populateSafetyAssumptionVector( pNtk, pAig );
    }

    vFlops = prepareFlopVector( pAig, Aig_ManRegNum(pAig)%2 == 0? Aig_ManRegNum(pAig)/2 : (Aig_ManRegNum(pAig)-1)/2);

    //vFlops = prepareFlopVector( pAig, 100 );

    switch  directive  {
    case FULL_BIERE_MODE:
        //if( Vec_PtrSize(vLive) == 0 && Vec_PtrSize(vAssertSafety) == 0 )
        //{
        //    printf("Input circuit has NO safety and NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformationAbs( FULL_BIERE_MODE, pNtk, pAig, vFlops, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case FULL_BIERE_ONE_LOOP_MODE:
        //if( Vec_PtrSize(vLive) == 0 && Vec_PtrSize(vAssertSafety) == 0 )
        //{
        //    printf("Input circuit has NO safety and NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformationOneStepLoop( FULL_BIERE_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case IGNORE_LIVENESS_KEEP_SAFETY_MODE:
        //if( Vec_PtrSize(vAssertSafety) == 0 )
        //{    
        //    printf("Input circuit has NO safety property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformationAbs( IGNORE_LIVENESS_KEEP_SAFETY_MODE, pNtk, pAig, vFlops, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t1 PO - only for safety property; liveness properties are ignored, if any.\n\tno additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case IGNORE_SAFETY_KEEP_LIVENESS_MODE:
        //if( Vec_PtrSize(vLive) == 0 )
        //{    
        //    printf("Input circuit has NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformationAbs( IGNORE_SAFETY_KEEP_LIVENESS_MODE, pNtk, pAig, vFlops, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t1 PO - only for liveness property; safety properties are ignored, if any.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n"); }
            break;
        //}
    case IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE:
        //if( Vec_PtrSize(vLive) == 0 )
        //{
        //    printf("Input circuit has NO liveness property, original network is not disturbed\n");
        //    return 1;
        //}
        //else
        //{
            pAigNew = LivenessToSafetyTransformationOneStepLoop( IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("New circuit is produced ignoring safety outputs!\nOnly liveness and fairness outputs are considered.\nShadow registers are not created\n"); }
            break;
        //}
    }

    pNtkNew = Abc_NtkFromAigPhase( pAigNew );
    pNtkNew.pName = Abc_UtilStrsav( pAigNew.pName );
    
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkCreateCone(): Network check has failed.\n"); }
    
    updateNewNetworkNameManager( pNtkNew, pAigNew, vecPiNames,vecLoNames );
    Abc_FrameSetCurrentNetwork( pAbc, pNtkNew );

     
                           
                                         
                                                                                                                      
                                                                                                                                          

                                         
                                                                                                                                
          
      
        
    return 0;

usage:
    fmt.Fprintf(os.Stdout, "usage: l2s [-1lsh]\n");
    fmt.Fprintf(os.Stdout, "\t         performs Armin Biere's live-to-safe transformation\n");
    fmt.Fprintf(os.Stdout, "\t-1 : no shadow logic, presume all loops are self loops\n");
    fmt.Fprintf(os.Stdout, "\t-l : ignore liveness and fairness outputs\n");
    fmt.Fprintf(os.Stdout, "\t-s : ignore safety assertions and assumptions\n");
    fmt.Fprintf(os.Stdout, "\t-h : print command usage\n");
    return 1;
}

func LivenessToSafetyTransformationWithLTL(mode int32, pNtk *Abc_Ntk_t, p *Aig_Man_t, vLive *Vec_Ptr_t, vFair *Vec_Ptr_t, vAssertSafety *Vec_Ptr_t, vAssumeSafety *Vec_Ptr_t, numLtlProcessed *int32, ltlBuffer *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var i, ii, iii, nRegCount int32;
    var pObjSavePi *Aig_Obj_t  = nil;
    var pObjSavedLo *Aig_Obj_t = nil;
    var pObjSavedLi *Aig_Obj_t = nil;
    var pObj, pMatch *Aig_Obj_t;
    var pObjSaveOrSaved *Aig_Obj_t = nil;
    var pObjSaveAndNotSaved *Aig_Obj_t = nil;
    var pObjSavedLoAndEquality *Aig_Obj_t;
    var pObjShadowLo, pObjShadowLi, pObjShadowLiDriver *Aig_Obj_t;
    var pObjXor, pObjXnor, pObjAndAcc *Aig_Obj_t;
    var pObjLive, pObjSafetyGate *Aig_Obj_t;
    var pObjSafetyPropertyOutput *Aig_Obj_t;
    var pObjOriginalSafetyPropertyOutput *Aig_Obj_t;
    var pDriverImage, pArgument, collectiveAssertSafety, collectiveAssumeSafety *Aig_Obj_t;
    var pNegatedSafetyConjunction *Aig_Obj_t  = nil;
    var pObjSafetyAndLiveToSafety *Aig_Obj_t;
    var nodeName, pFormula string;
    var piCopied, liCopied, loCopied, liCreated, loCreated, liveLatch int32;//, piVecIndex = 0, fairLatch = 0;
    var vSignal *Vec_Ptr_t;
    var vTopASTNodeArray *Vec_Ptr_t = nil;
    var pEnrtyGLOBALLY *ltlNode;
    var topNodeOfAST, tempTopASTNode *ltlNode;
    var vAigGFMap *Vec_Vec_t;
    var vSignalMemory *Vec_Ptr_t;
    var vGFFlopMemory *Vec_Ptr_t;
    var vPoForLtlProps *Vec_Ptr_t = nil;
    var vecInputLtlFormulae *Vec_Ptr_t;
    
    vecPis = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);
    vecPiNames = Vec_PtrAlloc( Saig_ManPiNum( p ) + 1);

    vecLos = Vec_PtrAlloc( Saig_ManRegNum( p )*2 + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );
    vecLoNames = Vec_PtrAlloc( Saig_ManRegNum( p )*2 + 1 + Vec_PtrSize( vLive ) + Vec_PtrSize( vFair ) );

    //****************************************************************
    //step0: Parsing the LTL formula
    //****************************************************************
    //Vec_PtrForEachEntry( char *, pNtk->vLtlProperties, pFormula, i )
    //    printf("\ninput LTL formula [%d] = %s\n", i, pFormula );


                           


    //***************************************************************************
    //Reading input LTL formulae from Ntk data-structure and creating 
    //AST for them, Steps involved: 
    //        parsing -> AST creation -> well-formedness check -> signal name check
    //***************************************************************************

    //resetting numLtlProcessed
    *numLtlProcessed = 0;
    
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
        //if( ltlBuffer )
        vecInputLtlFormulae = ltlBuffer;
        //vecInputLtlFormulae = pNtk->vLtlProperties;
        if vecInputLtlFormulae != nil {
            vTopASTNodeArray = Vec_PtrAlloc( Vec_PtrSize( vecInputLtlFormulae ) );
            fmt.Printf("\n");
            Vec_PtrForEachEntry( char *, vecInputLtlFormulae, pFormula, i )
            {
                tempTopASTNode = parseFormulaCreateAST( pFormula );
                //traverseAbstractSyntaxTree_postFix( tempTopASTNode );
                if tempTopASTNode != nil {
                    fmt.Printf("Formula %d: AST is created, ", i+1);
                    if isWellFormed( tempTopASTNode ) {
                        fmt.Printf("Well-formedness check PASSED, ");
                    } else {
                        fmt.Printf("Well-formedness check FAILED!!\n");
                        fmt.Printf("AST will be ignored for formula %d, no extra logic will be added for this formula\n", i+1);
                        //do memory management to free the created AST
                        continue;
                    }
                    if checkSignalNameExistence( pNtk, tempTopASTNode ) {
                        fmt.Printf("Signal check PASSED\n");
                    } else {
                        fmt.Printf("Signal check FAILED!!");
                        fmt.Printf("AST will be ignored for formula %d, no extra logic will be added for this formula\n", i+1);
                        //do memory management to free the created AST
                        continue;
                    }
                    Vec_PtrPush( vTopASTNodeArray, tempTopASTNode );
                    (*numLtlProcessed)++;
                
                } else {
                    fmt.Printf("\nNo AST has been created for formula %d, no extra logic will be added\n", i+1); }
            }
        }
        fmt.Printf("\n");
        if Vec_PtrSize( vTopASTNodeArray ) == 0 {
            //printf("\nNo AST has been created for any formula; hence the circuit is left untouched\n");
            fmt.Printf("\nCurrently aborting, need to take care when Vec_PtrSize( vTopASTNodeArray ) == 0\n");
            os.Exit(int(0));
        }
    }

        //****************************************************************
        // Step1: create the new manager
        // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
        // nodes, but this selection is arbitrary - need to be justified
        //****************************************************************
        pNew = Aig_ManStart( 2 * Aig_ManObjNumMax(p) );
        pNew.pName = (char *)malloc( len(pNtk.pName) + len("_l3s") + 1 );
        sprintf(pNew.pName, "%s_%s", pNtk.pName, "l3s");
        pNew.pSpec = nil;
    
        //****************************************************************
        // Step 2: map constant nodes
        //****************************************************************
        pObj = Aig_ManConst1( p );
        pObj.pData = Aig_ManConst1( pNew );

        //****************************************************************
        // Step 3: create true PIs
        //****************************************************************
        Saig_ManForEachPi( p, pObj, i )
        {
            piCopied++;
            pObj.pData = Aig_ObjCreateCi(pNew);
            Vec_PtrPush( vecPis, pObj.pData );
            nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkPi( pNtk, i ) ));
            Vec_PtrPush( vecPiNames, nodeName );
        }

        //****************************************************************
        // Step 4: create the special Pi corresponding to SAVE
        //****************************************************************
        if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            pObjSavePi = Aig_ObjCreateCi( pNew );
            nodeName = "SAVE_BIERE";
Vec_PtrPush( vecPiNames, nodeName );
        }
        
        //****************************************************************
        // Step 5: create register outputs
        //****************************************************************
        Saig_ManForEachLo( p, pObj, i )
        {
            loCopied++;
            pObj.pData = Aig_ObjCreateCi(pNew);
            Vec_PtrPush( vecLos, pObj.pData );
            nodeName = Abc_UtilStrsav(Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ));
            Vec_PtrPush( vecLoNames, nodeName );
        }

        //****************************************************************
        // Step 6: create "saved" register output
        //****************************************************************
        if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            loCreated++;
            pObjSavedLo = Aig_ObjCreateCi( pNew );
            Vec_PtrPush( vecLos, pObjSavedLo );
            nodeName = "SAVED_LO";
            Vec_PtrPush( vecLoNames, nodeName );
        }

        //****************************************************************
        // Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
        //****************************************************************
        if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            pObjSaveOrSaved = Aig_Or( pNew, pObjSavePi, pObjSavedLo );
            pObjSaveAndNotSaved = Aig_And( pNew, pObjSavePi, Aig_Not(pObjSavedLo) );
        }

        //********************************************************************
        // Step 8: create internal nodes
        //********************************************************************
        Aig_ManForEachNode( p, pObj, i )
        {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
        }

        
        //********************************************************************
        // Step 8.x : create PO for each safety assertions
        // NOTE : Here the output is purposely inverted as it will be thrown to 
        // dprove
        //********************************************************************
        assert(pNegatedSafetyConjunction == nil, "pNegatedSafetyConjunction == nil");
        if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) == 0 {
                pObjAndAcc = Aig_ManConst1( pNew );
                Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
                {
                    pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                    pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
                }
                pNegatedSafetyConjunction = Aig_Not(pObjAndAcc);
                if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
                    pObjOriginalSafetyPropertyOutput = Aig_ObjCreateCo( pNew, Aig_Not(pObjAndAcc) ); }
            
            } else if Vec_PtrSize( vAssertSafety ) != 0 && Vec_PtrSize( vAssumeSafety ) != 0 {
                pObjAndAcc = Aig_ManConst1( pNew );
                Vec_PtrForEachEntry( Aig_Obj_t *, vAssertSafety, pObj, i )
                {
                    pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                    pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
                }
                collectiveAssertSafety = pObjAndAcc;

                pObjAndAcc = Aig_ManConst1( pNew );
                Vec_PtrForEachEntry( Aig_Obj_t *, vAssumeSafety, pObj, i )
                {
                    pArgument = Aig_NotCond( (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0( pObj ) );
                    pObjAndAcc = Aig_And( pNew, pArgument, pObjAndAcc );
                }
                collectiveAssumeSafety = pObjAndAcc;
                pNegatedSafetyConjunction =  Aig_And( pNew, Aig_Not(collectiveAssertSafety), collectiveAssumeSafety );
                if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
                    pObjOriginalSafetyPropertyOutput = Aig_ObjCreateCo( pNew, Aig_And( pNew, Aig_Not(collectiveAssertSafety), collectiveAssumeSafety ) ); }
            
            } else {
                fmt.Printf("WARNING!! No safety property is found, a new (negated) constant 1 output is created\n");
                pNegatedSafetyConjunction = Aig_Not( Aig_ManConst1(pNew) );
                if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
                    pObjOriginalSafetyPropertyOutput = Aig_ObjCreateCo( pNew, Aig_Not( Aig_ManConst1(pNew) ) ); }
            }
        }
        assert(pNegatedSafetyConjunction != nil, "pNegatedSafetyConjunction != nil");

        //********************************************************************
        // Step 9: create the safety property output gate for the liveness properties
        // discuss with Sat/Alan for an alternative implementation
        //********************************************************************
        if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            vPoForLtlProps = Vec_PtrAlloc( Vec_PtrSize( vTopASTNodeArray ) );
            if Vec_PtrSize( vTopASTNodeArray ) {
                //no effective AST for any input LTL property
                //must do something graceful
            }
            for  i=0; i<Vec_PtrSize( vTopASTNodeArray ); i++  {
                pObjSafetyPropertyOutput = Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Aig_ObjFanin0(pObj).pData );
                Vec_PtrPush( vPoForLtlProps, pObjSafetyPropertyOutput );
            }
        }

        //*************************************************************************************
        // Step 10: Placeholder PO's were created for Liveness property outputs in the
        // last step. FYI, # of new liveness property outputs = # of LTL properties in the circuit
        // It is time for creation of loop LI's and other stuff
        // Now creating register inputs for the original flops
        //*************************************************************************************
        nRegCount = 0;
        
        Saig_ManForEachLo( p, pObj, i )
        {
            pMatch = Saig_ObjLoToLi( p, pObj );
            Aig_ObjCreateCo( pNew, Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pMatch).pData, Aig_ObjFaninC0( pMatch ) ) );
            nRegCount++;
            liCopied++;
        }

        //*************************************************************************************
        // Step 11: create register input corresponding to the register "saved"
        //*************************************************************************************
        if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            #ifndef DUPLICATE_CKT_DEBUG
                pObjSavedLi = Aig_ObjCreateCo( pNew, pObjSaveOrSaved );
                nRegCount++;
                liCreated++;

                pObjAndAcc = Aig_ManConst1( pNew );

                //*************************************************************************************
                // Step 11: create the family of shadow registers, then create the cascade of Xnor 
                // and And gates for the comparator 
                //*************************************************************************************
                Saig_ManForEachLo( p, pObj, i )
                {
                //printf("\nKEMON RENDY = %s", Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i )) );
                //top|route0_target0_queue_with_credit0_queue0
                //top|route0_master0_queue2
                //    if( strcmp(  Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ), "top|route0_queue1_num[0]" ) == 0 
                //            || strcmp(  Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ), "top|route0_queue1_num[1]" ) == 0 || strcmp(  Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ), "top|route0_queue1_num[2]" ) == 0 )
                    {        
                        pObjShadowLo = Aig_ObjCreateCi( pNew );

                    #ifdef PROPAGATE_NAMES
                        Vec_PtrPush( vecLos, pObjShadowLo );
                        nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ) ) + 10 );
                        sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i ) ), "SHADOW" );
                        
                        Vec_PtrPush( vecLoNames, nodeName );
                    #endif

                        pObjShadowLiDriver = Aig_Mux( pNew, pObjSaveAndNotSaved, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
                        pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );
                        nRegCount++;
                        loCreated++; liCreated++;
            
                        pObjXor = Aig_Exor( pNew, (Aig_Obj_t *)pObj->pData, pObjShadowLo );
                        pObjXnor = Aig_Not( pObjXor );
                    
                        pObjAndAcc = Aig_And( pNew, pObjXnor, pObjAndAcc );
                    }
                }

                // create the AND gate whose output will be the signal "looped"
                pObjSavedLoAndEquality = Aig_And( pNew, pObjSavedLo, pObjAndAcc );

                // create the master AND gate and corresponding AND and OR logic for the liveness properties
                
                //*************************************************************************************
                // Step 11: logic for LTL properties:- (looped & ~theta) where theta is the input ltl 
                // property
                // Description of some data-structure:
                //-------------------------------------------------------------------------------------
                // Name          | Type            | Purpose
                //-------------------------------------------------------------------------------------
                // vSignalMemory | Vec_Ptr_t *     | A vector across all ASTs of the LTL properties
                //                 |                   | It remembers if OR+Latch for GF node has already been
                //                 |                   | created for a particular signal.
                //               |                 |
                // vGFFlopMemory | Vec_Ptr_t *     | A vector across all ASTs of the LTL properties
                //                 |                 | remembers if OR+Latch of a GF node has already been created
                //                 |                 |
                // vSignal         | Vec_Ptr_t *     | vector for each AST; contains pointers from GF nodes
                //               |                 | to AIG signals
                //                 |                   |
                // vAigGFMap     | Vec_Vec_t *     | vAigGFMap[ index ] = vector of GF nodes pointing to
                //               |                 | the same AIG node; "index" is the index of that
                //               |                 | AIG node in the vector vSignal
                //*************************************************************************************
                
                vSignalMemory = Vec_PtrAlloc(10);
                vGFFlopMemory = Vec_PtrAlloc(10);

                Vec_PtrForEachEntry( ltlNode *, vTopASTNodeArray, topNodeOfAST, iii )
                {
                                    vSignal = Vec_PtrAlloc( 10 );
                                    vAigGFMap = Vec_VecAlloc( 10 );
                                    
                                    //*************************************************************************************
                                    //Step 11a: for the current AST, find out the leaf level Boolean signal pointers from 
                                    // the NEW aig.
                                    //*************************************************************************************
                                    populateBoolWithAigNodePtr( pNtk, p, pNew, topNodeOfAST );
                                    assert( checkAllBoolHaveAIGPointer( topNodeOfAST ) );

                                    //*************************************************************************************
                                    //Step 11b: for each GF node, compute the pointer in AIG that it should point to
                                    // In particular, if the subtree below GF is some Boolean crown (including the case
                                    // of simple negation, create new logic and populate the AIG pointer in GF node
                                    // accordingly
                                    //*************************************************************************************
                                    populateAigPointerUnitGF( pNew, topNodeOfAST, vSignal, vAigGFMap );
                                    
                                    //*************************************************************************************
                                    //Step 11c: everything below GF are computed. Now, it is time to create logic for individual 
                                    // GF nodes (i.e. the OR gate and the latch and the Boolean crown of the AST
                                    //*************************************************************************************
                                    Vec_PtrForEachEntry( Aig_Obj_t *, vSignal, pObj, i )
                                    {
                                        //*********************************************************
                                        // Step 11c.1: if the OR+Latch of the particular signal is
                                        // not already created, create it. It may have already been 
                                        // created from another property, so check it before creation
                                        //*********************************************************
                                        if( Vec_PtrFind( vSignalMemory, pObj ) == -1 )
                                        {
                                            liveLatch++;

                                            pDriverImage = pObj;
                                            pObjShadowLo = Aig_ObjCreateCi( pNew );
                                            pObjShadowLiDriver = Aig_Or( pNew, pObjShadowLo, Aig_And( pNew, pDriverImage, pObjSaveOrSaved ) );
                                            pObjShadowLi = Aig_ObjCreateCo( pNew, pObjShadowLiDriver );

                                            nRegCount++;
                                            loCreated++; liCreated++;

                                            Vec_PtrPush( vSignalMemory, pObj );
                                            Vec_PtrPush( vGFFlopMemory, pObjShadowLo );

                                            #if 1
                                            #ifdef PROPAGATE_NAMES
                                                Vec_PtrPush( vecLos, pObjShadowLo );
                                                //nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
                                                //sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "LIVENESS" );
                                                nodeName = (char *)malloc( 20 );
                                                sprintf( nodeName, "n%d__%s", Aig_ObjId(pObjShadowLo), "GF_flop" );
                                                Vec_PtrPush( vecLoNames, nodeName );
                                            #endif
                                            #endif
                                        }
                                        else
                                            pObjShadowLo = (Aig_Obj_t *)Vec_PtrEntry( vGFFlopMemory, Vec_PtrFind( vSignalMemory, pObj ) );
                                                                                
                                        Vec_VecForEachEntryLevel( ltlNode *, vAigGFMap, pEnrtyGLOBALLY, ii, i )
                                            setAIGNodePtrOfGloballyNode( pEnrtyGLOBALLY, pObjShadowLo);
                                            

                                        //#ifdef PROPAGATE_NAMES
                                        //    Vec_PtrPush( vecLos, pObjShadowLo );
                                        //    nodeName = (char *)malloc( strlen( Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ) ) + 12 );
                                        //    sprintf( nodeName, "%s__%s", Abc_ObjName( Abc_NtkPo( pNtk, getPoIndex( p, pObj ) ) ), "LIVENESS" );
                                        //    Vec_PtrPush( vecLoNames, nodeName );
                                        //#endif
                                        
                                    }
                                    
                                    //*********************************************************
                                    //Step 11c.2: creating the Boolean crown
                                    //*********************************************************
                                    buildLogicFromLTLNode( pNew, topNodeOfAST );

                                    //*********************************************************
                                    //Step 11c.3: creating logic for (looped & ~theta) and patching
                                    // it with the proper PO
                                    //Note: if ALLOW_SAFETY_PROPERTIES is defined then the final AND
                                    //gate is a conjunction of safety & liveness, i.e. SAFETY & (looped => theta)
                                    //since ABC convention demands a NOT gate at the end, the property logic 
                                    //becomes !( SAFETY & (looped => theta) ) = !SAFETY + (looped & !theta)
                                    //*********************************************************
                                    pObjLive = retriveAIGPointerFromLTLNode( topNodeOfAST );
                                    pObjSafetyGate = Aig_And( pNew, pObjSavedLoAndEquality, Aig_Not(pObjLive) );
                                    #ifdef ALLOW_SAFETY_PROPERTIES
                                        printf("liveness output is conjoined with safety assertions\n");
                                        pObjSafetyAndLiveToSafety = Aig_Or( pNew, pObjSafetyGate, pNegatedSafetyConjunction );
                                        pObjSafetyPropertyOutput = (Aig_Obj_t *)Vec_PtrEntry( vPoForLtlProps, iii );
                                        Aig_ObjPatchFanin0( pNew, pObjSafetyPropertyOutput, pObjSafetyAndLiveToSafety );
                                    #else
                                        pObjSafetyPropertyOutput = Vec_PtrEntry( vPoForLtlProps, iii );
                                        Aig_ObjPatchFanin0( pNew, pObjSafetyPropertyOutput, pObjSafetyGate );
                                    #endif
                                    //refreshing vSignal and vAigGFMap arrays
                                    Vec_PtrFree( vSignal );
                                    Vec_VecFree( vAigGFMap );
                }
                                
            #endif
        }
      

    Aig_ManSetRegNum( pNew, nRegCount );

    Aig_ManCiCleanupBiere( pNew );
    Aig_ManCoCleanupBiere( pNew );
    
    Aig_ManCleanup( pNew );
    
    assert(Aig_ManCheck( pNew ), "Aig_ManCheck( pNew )");
    
    if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
            assert((Aig_Obj_t *)Vec_PtrEntry(pNew.vCos, Saig_ManPoNum(pNew)+Aig_ObjCioId(pObjSavedLo)-Saig_ManPiNum(p)-1) == pObjSavedLi, "(Aig_Obj_t *)Vec_PtrEntry(pNew.vCos, Saig_ManPoNum(pNew)+Aig_ObjCioId(pObjSavedLo)-Saig_ManPiNum(p)-1) == pObjSavedLi");
            assert(Saig_ManPiNum( p ) + 1 == Saig_ManPiNum( pNew ), "Saig_ManPiNum( p ) + 1 == Saig_ManPiNum( pNew )");
            //assert( Saig_ManRegNum( pNew ) == Saig_ManRegNum( p ) * 2 + 1 + liveLatch + fairLatch );
    }


    return pNew;
}

func Abc_CommandAbcLivenessToSafetyWithLTL(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pOut, pErr *os.File;
    var pNtk, pNtkTemp, pNtkNew, pNtkOld *Abc_Ntk_t;
    var pAig *Aig_Man_t;
    var pAigNew *Aig_Man_t = nil;
    var c int32;
    var vLive, vFair, vAssertSafety, vAssumeSafety *Vec_Ptr_t;
    var directive int32  = -1;
//    char *ltfFormulaString = NULL;
    var numOfLtlPropOutput int32;//, LTL_FLAG = 0;
    var ltlBuffer *Vec_Ptr_t;
                
    pNtk = Abc_FrameReadNtk(pAbc);
    pOut = Abc_FrameReadOut(pAbc);
    pErr = Abc_FrameReadErr(pAbc);

    if argc == 1 {
        assert(directive == -1, "directive == -1");
        directive = FULL_BIERE_MODE;
    
    } else {
        Extra_UtilGetoptReset();
        for  ( c = Extra_UtilGetopt( argc, argv, "1slhf" ) ) != -1  {
            switch  c  {
            case '1': 
                if( directive == -1 )
                    directive = FULL_BIERE_ONE_LOOP_MODE;
                else {
                    assert(directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE || directive == IGNORE_SAFETY_KEEP_LIVENESS_MODE, "directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE || directive == IGNORE_SAFETY_KEEP_LIVENESS_MODE");
                    if directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
                        directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                    } else {
                        directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE; }
                }
                break;
            case 's':
                if( directive == -1 )
                    directive = IGNORE_SAFETY_KEEP_LIVENESS_MODE;
                else {
                    if directive != FULL_BIERE_ONE_LOOP_MODE {
                        goto usage; }
                    assert(directive == FULL_BIERE_ONE_LOOP_MODE, "directive == FULL_BIERE_ONE_LOOP_MODE");
                    directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE;
                }
                break;
            case 'l':
                if( directive == -1 )
                    directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                else {
                    if directive != FULL_BIERE_ONE_LOOP_MODE {
                        goto usage; }
                    assert(directive == FULL_BIERE_ONE_LOOP_MODE, "directive == FULL_BIERE_ONE_LOOP_MODE");
                    directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE;
                }
                break;
            case 'f':
                //assert( argc >= 3 );
                //vecLtlFormula = Vec_PtrAlloc( argc - 2 );
                //if( argc >= 3 )
                //{
                //    for( t=3; t<=argc; t++ )
                //    {
                //        printf("argv[%d] = %s\n", t-1, argv[t-1]);
                //        Vec_PtrPush( vecLtlFormula, argv[t-1] );
                //    }
                //}
                //printf("argv[argc] = %s\n", argv[argc-1]);
                //ltfFormulaString = argv[2];
                
                //LTL_FLAG = 1;
                fmt.Printf("\nILLEGAL FLAG: aborting....\n");
                os.Exit(int(0));
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
            }
        }
    }

    if pNtk == nil {
        fmt.Fprintf(pErr, "Empty network.\n");
        return 1;
    }
    if !Abc_NtkIsStrash( pNtk ) {
        fmt.Printf("The input network was not strashed, strashing....\n");
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pNtkOld = pNtkTemp;
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
        vAssertSafety = populateSafetyAssertionVector( pNtk, pAig );
        vAssumeSafety = populateSafetyAssumptionVector( pNtk, pAig );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkOld = pNtk;
        vLive = populateLivenessVector( pNtk, pAig );
        vFair = populateFairnessVector( pNtk, pAig );
        vAssertSafety = populateSafetyAssertionVector( pNtk, pAig );
        vAssumeSafety = populateSafetyAssumptionVector( pNtk, pAig );
    }

    if pAbc.vLTLProperties_global != nil {
        ltlBuffer = pAbc.vLTLProperties_global;
    } else {
        ltlBuffer = nil; }

    switch  directive  {
    case FULL_BIERE_MODE:
            pAigNew = LivenessToSafetyTransformationWithLTL( FULL_BIERE_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety, &numOfLtlPropOutput, ltlBuffer );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t%d POs - one for safety and %d for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n", numOfLtlPropOutput+1, numOfLtlPropOutput); }
            break;
        
    case FULL_BIERE_ONE_LOOP_MODE:
            pAigNew = LivenessToSafetyTransformationOneStepLoop( FULL_BIERE_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n"); }
            break;
        
    case IGNORE_LIVENESS_KEEP_SAFETY_MODE:
            pAigNew = LivenessToSafetyTransformationWithLTL( IGNORE_LIVENESS_KEEP_SAFETY_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety, &numOfLtlPropOutput, ltlBuffer );
            assert(numOfLtlPropOutput == 0, "numOfLtlPropOutput == 0");
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t1 PO - only for safety property; liveness properties are ignored, if any.\n\tno additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n"); }
            break;
        
    case IGNORE_SAFETY_KEEP_LIVENESS_MODE:
            pAigNew = LivenessToSafetyTransformationWithLTL( IGNORE_SAFETY_KEEP_LIVENESS_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety, &numOfLtlPropOutput, ltlBuffer );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("A new circuit is produced with\n\t%d PO - only for liveness property; safety properties are ignored, if any.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n", numOfLtlPropOutput); }
            break;
        
    case IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE:
            pAigNew = LivenessToSafetyTransformationOneStepLoop( IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety );
            if Aig_ManRegNum(pAigNew) != 0 {
                fmt.Printf("New circuit is produced ignoring safety outputs!\nOnly liveness and fairness outputs are considered.\nShadow registers are not created\n"); }
            break;
    }

     
                   
     
                                                                                                                            
                                         
                                                                                                                                 
     
         
     
                                
                                                   
                      
         
                               
             
                                                    
                         
             
                                          
             
                                                                              
                                                          
                                   
                                                      
                                                             
                                                             
                                                                            
                                                                             
             
                
             
                                                  
                               
                                                             
                                                             
                                                                            
                                                                             
             
                                                                                                                          
         
                           
         
                               
             
                                                    
                         
             
                                          
             
                                                                              
                                                          
                                   
                                                      
                                                             
                                                             
                                                                            
                                                                             
             
                
             
                                                  
                               
                                                             
                                                             
                                                                            
                                                                             
             
                                                                                                                                                 
                                             
                                                                                                                                            
         
                           
         
                               
             
                                                    
                         
             
            
                                          
             
                                                                              
                                                          
                                   
                                                      
                                                             
                                                             
                                                                            
                                                                             
             
                
             
                                                  
                               
                                                             
                                                             
                                                                            
                                                                             
             
                                                                                                                                                 
                                             
                                                                                                                                                           
         
                           
                       
            
                       
     
      
    
     
                                 
                                                                         
                                                                             
                                                                              
                                                                                                       
                                                                                                      
                                                                    
                                                                          
      

    pNtkNew = Abc_NtkFromAigPhase( pAigNew );
    pNtkNew.pName = Abc_UtilStrsav( pAigNew.pName );
    
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkCreateCone(): Network check has failed.\n"); }
    
    updateNewNetworkNameManager( pNtkNew, pAigNew, vecPiNames, vecLoNames );
    Abc_FrameSetCurrentNetwork( pAbc, pNtkNew );

     
                           
                                         
                                                                                                                      
                                                                                                                                          

                                         
                                                                                                                                
          
      
        
    return 0;

usage:
    fmt.Fprintf(os.Stdout, "usage: l3s [-1lsh]\n");
    fmt.Fprintf(os.Stdout, "\t         performs Armin Biere's live-to-safe transformation\n");
    fmt.Fprintf(os.Stdout, "\t-1 : no shadow logic, presume all loops are self loops\n");
    fmt.Fprintf(os.Stdout, "\t-l : ignore liveness and fairness outputs\n");
    fmt.Fprintf(os.Stdout, "\t-s : ignore safety assertions and assumptions\n");
    fmt.Fprintf(os.Stdout, "\t-h : print command usage\n");
    return 1;
}


//ABC_NAMESPACE_IMPL_END
