package live

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [kliveness.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis        [Main implementation module of the algorithm k-Liveness    ] 
              [invented by Koen Claessen, Niklas Sorensson. Implements]
              [the code for 'kcs'. 'kcs' pre-processes based on switch]
        [and then runs the (absorber circuit >> pdr) loop  ]

  Author      [Sayak Ray]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 31, 2012.]

***********************************************************************/

//#include <stdio.h>
//#include "base/main/main.h"
//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include <string.h>
//#include "base/main/mainInt.h"
//#include "proof/pdr/pdr.h"
//#include <time.h>

//#define WITHOUT_CONSTRAINTS

//ABC_NAMESPACE_IMPL_START

/***************** Declaration of standard ABC related functions ********************/
// extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern Abc_Ntk_t * Abc_NtkFromAigPhase( Aig_Man_t * pMan );

// extern Abc_Ntk_t * Abc_NtkMakeOnePo( Abc_Ntk_t * pNtk, int Output, int nRange );

// extern void Aig_ManDumpBlif( Aig_Man_t * p, char * pFileName, Vec_Ptr_t * vPiNames, Vec_Ptr_t * vPoNames );

/***********************************************************************************/

/***************** Declaration of kLiveness related functions **********************/
//function defined in kLiveConstraints.c
// extern Aig_Man_t *generateWorkingAig( Aig_Man_t *pAig, Abc_Ntk_t *pNtk, int *pIndex0Live );


//function defined in arenaViolation.c
// extern Aig_Man_t *generateWorkingAigWithDSC( Aig_Man_t *pAig, Abc_Ntk_t *pNtk, int *pIndex0Live, Vec_Ptr_t *vMasterBarriers );


//function defined in disjunctiveMonotone.c
// extern Vec_Ptr_t *findDisjunctiveMonotoneSignals( Abc_Ntk_t *pNtk );

// extern Vec_Int_t *createSingletonIntVector( int i );

/***********************************************************************************/
// extern Aig_Man_t *generateDisjunctiveTester( Abc_Ntk_t *pNtk, Aig_Man_t *pAig, int combN, int combK );

// extern Aig_Man_t *generateGeneralDisjunctiveTester( Abc_Ntk_t *pNtk, Aig_Man_t *pAig, int combK );


//Definition of some macros pertaining to modes/switches
const SIMPLE_kCS = 0
const kCS_WITH_SAFETY_INVARIANTS = 1
const kCS_WITH_DISCOVER_MONOTONE_SIGNALS = 2
const kCS_WITH_SAFETY_AND_DCS_INVARIANTS = 3
const kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS = 4

//unused function
     
                                                                
 
                    
          

                                      
     
                                                                                
         
                                         
                                                                                       
         
         

                
 
      

func readLiveSignal_0(pAig *Aig_Man_t, liveIndex_0 int32) *Aig_Obj_t {
    var pObj *Aig_Obj_t;

    pObj = Aig_ManCo( pAig, liveIndex_0 );
    return Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj));
}

func readLiveSignal_k(pAig *Aig_Man_t, liveIndex_k int32) *Aig_Obj_t {
    var pObj *Aig_Obj_t;

    pObj = Aig_ManCo( pAig, liveIndex_k );
    return Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj));
}

//unused funtion
     
                                                                                        
 
                    
          

                                
                    
        
                                          
         
                                                                                    
             
                                             
                                                                                           
             
             

                
 
      

//unused function
     
                                                                                             
                                                                           
 
                    
                     
                    

                                                          

                  
     
                                          
         
                                                
                                                                                                                      
         
     
                  
     
                                          
         
                                                                        
                                                                                                                      
         
     

                  
     
                                          
         
                                                
                                                                                                                          
         
     

                  
     
                                          
         
                                                                        
                                                                                                                      
         
     

                               
                                       
                                                                                   
                                                                                               
 
      

func introduceAbsorberLogic(pAig *Aig_Man_t, pLiveIndex_0 *int32, pLiveIndex_k *int32, nonFirstIteration int32) *Aig_Man_t {
    var pNewAig *Aig_Man_t;
    var pObj, pObjAbsorberLo, pPInNewArg, pPOutNewArg *Aig_Obj_t;
    var pPIn *Aig_Obj_t = nil;
    var pPOut *Aig_Obj_t = nil;
    var pPOutCo *Aig_Obj_t = nil;
    var pFirstAbsorberOr, pSecondAbsorberOr *Aig_Obj_t;
    var i int32;
    var piCopied, loCreated, loCopied, liCreated, liCopied int32; 
    var nRegCount int32;

    assert(*pLiveIndex_0 != -1, "*pLiveIndex_0 != -1");
    if nonFirstIteration == 0 {
        assert(*pLiveIndex_k == -1, "*pLiveIndex_k == -1");
    } else {
        assert(*pLiveIndex_k != -1, "*pLiveIndex_k != -1"); }

    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNewAig = Aig_ManStart( Aig_ManObjNumMax(pAig) );
    pNewAig.pName = (char *)malloc( len(pAig.pName) + len("_kCS") + 1 );
    sprintf(pNewAig.pName, "%s_%s", pAig.pName, "kCS");
        pNewAig.pSpec = nil;

    //****************************************************************
    // reading the signal pIn, and pOut
    //****************************************************************

    pPIn = readLiveSignal_0( pAig, *pLiveIndex_0 );
    if *pLiveIndex_k == -1 {
        pPOut = nil;
    } else {
        pPOut = readLiveSignal_k( pAig, *pLiveIndex_k ); }
    
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
        piCopied++;
        pObj.pData = Aig_ObjCreateCi(pNewAig);
    }

    //****************************************************************
    // Step 5: create register outputs
    //****************************************************************
        Saig_ManForEachLo( pAig, pObj, i )
        {
        loCopied++;
        pObj.pData = Aig_ObjCreateCi(pNewAig);
        }

    //****************************************************************
    // Step 6: create "D" register output for the ABSORBER logic
    //****************************************************************
    loCreated++;
    pObjAbsorberLo = Aig_ObjCreateCi( pNewAig );

    nRegCount = loCreated + loCopied;

    //********************************************************************
    // Step 7: create internal nodes
    //********************************************************************
        Aig_ManForEachNode( pAig, pObj, i )
    {
        pObj.pData = Aig_And( pNewAig, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }

    //****************************************************************
    // Step 8: create the two OR gates of the OBSERVER logic
    //****************************************************************
    if nonFirstIteration == 0 {
        assert((pPIn != nil), "pPIn");
        
        if !Aig_IsComplement(pPIn) {
            pPInNewArg = (Aig_Obj_t *)((Aig_Regular(pPIn)).pData);
        } else {
            pPInNewArg = Aig_Not((Aig_Obj_t *)((Aig_Regular(pPIn)).pData));
        }

        pFirstAbsorberOr = Aig_Or( pNewAig, Aig_Not(pPInNewArg), pObjAbsorberLo );
        pSecondAbsorberOr = Aig_Or( pNewAig, pPInNewArg, Aig_Not(pObjAbsorberLo) );
    
    } else {
        assert((pPOut != nil), "pPOut");

        if !Aig_IsComplement(pPIn) {
            pPInNewArg = (Aig_Obj_t *)((Aig_Regular(pPIn)).pData);
        } else {
            pPInNewArg = Aig_Not((Aig_Obj_t *)((Aig_Regular(pPIn)).pData));
        }
        if !Aig_IsComplement(pPOut) {
            pPOutNewArg = (Aig_Obj_t *)((Aig_Regular(pPOut)).pData);
        } else {
            pPOutNewArg = Aig_Not((Aig_Obj_t *)((Aig_Regular(pPOut)).pData));
        }
        
        pFirstAbsorberOr = Aig_Or( pNewAig, Aig_Not(pPOutNewArg), pObjAbsorberLo );
        pSecondAbsorberOr = Aig_Or( pNewAig, pPInNewArg, Aig_Not(pObjAbsorberLo) );
    }    
    
    //********************************************************************
    // Step 9: create primary outputs 
    //********************************************************************
    Saig_ManForEachPo( pAig, pObj, i )
    {
        pObj.pData = Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) ); 
        if i == *pLiveIndex_k {
            pPOutCo = (Aig_Obj_t *)(pObj.pData); }
    }

    //create new po
    if nonFirstIteration == 0 {
        assert(pPOutCo == nil, "pPOutCo == nil");
        pPOutCo = Aig_ObjCreateCo( pNewAig, pSecondAbsorberOr );     

        *pLiveIndex_k = i;
        
    } else {
        assert(pPOutCo != nil, "pPOutCo != nil");
        //pPOutCo = Aig_ObjCreateCo( pNewAig, pSecondAbsorberOr );     
        //*pLiveIndex_k = Saig_ManPoNum(pAig);

        Aig_ObjPatchFanin0( pNewAig, pPOutCo, pSecondAbsorberOr );
    }

    Saig_ManForEachLi( pAig, pObj, i )
    {
        liCopied++;
        Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) );
    }

    //create new li
    liCreated++;
    Aig_ObjCreateCo( pNewAig, pFirstAbsorberOr );

    Aig_ManSetRegNum( pNewAig, nRegCount );
    Aig_ManCleanup( pNewAig );
    
    assert(Aig_ManCheck( pNewAig ), "Aig_ManCheck( pNewAig )");

    assert(*pLiveIndex_k != - 1, "*pLiveIndex_k != - 1");
    return pNewAig;
}

func modifyAigToApplySafetyInvar(pAig *Aig_Man_t, csTarget int32, safetyInvarPO int32) {
    var pObjPOSafetyInvar, pObjSafetyInvar *Aig_Obj_t;
    var pObjPOCSTarget, pObjCSTarget *Aig_Obj_t;
    var pObjCSTargetNew *Aig_Obj_t;

    pObjPOSafetyInvar = Aig_ManCo( pAig, safetyInvarPO );
    pObjSafetyInvar =  Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPOSafetyInvar), Aig_ObjFaninC0(pObjPOSafetyInvar));    
    pObjPOCSTarget = Aig_ManCo( pAig, csTarget );
    pObjCSTarget = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPOCSTarget), Aig_ObjFaninC0(pObjPOCSTarget));    

    pObjCSTargetNew = Aig_And( pAig, pObjSafetyInvar, pObjCSTarget );
    Aig_ObjPatchFanin0( pAig, pObjPOCSTarget, pObjCSTargetNew );
}

func flipConePdr(pAig *Aig_Man_t, directive int32, targetCSPropertyIndex int32, safetyInvariantPOIndex int32, absorberCount int32) int32 {
    var RetValue, i int32;
    var pObjTargetPo *Aig_Obj_t;
    var pAigDupl *Aig_Man_t;
    var Pars Pdr_Par_t;
    var pPars *Pdr_Par_t = &Pars;
    var pCex *Abc_Cex_t  = nil;

    var fileName string;
    
    fileName = (char *)malloc(sizeof(char) * 50);
    fileName = fmt.Sprintf("%s_%d.%s", "kLive", absorberCount, "blif");

    if directive == kCS_WITH_SAFETY_INVARIANTS || directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS || directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS {
        assert(safetyInvariantPOIndex != -1, "safetyInvariantPOIndex != -1");
        modifyAigToApplySafetyInvar(pAig, targetCSPropertyIndex, safetyInvariantPOIndex);
    }

    pAigDupl = pAig;
    pAig = Aig_ManDupSimple( pAigDupl );

    for  i=0; i<Saig_ManPoNum(pAig); i++  {
        pObjTargetPo = Aig_ManCo( pAig, i );
        Aig_ObjChild0Flip( pObjTargetPo );
    }

    Pdr_ManSetDefaultParams( pPars );
    pPars.fVerbose = 1;
    pPars.fNotVerbose = 1;
    pPars.fSolveAll = 1;
    pAig.vSeqModelVec = nil;

    Aig_ManCleanup( pAig );
    assert(Aig_ManCheck( pAig ), "Aig_ManCheck( pAig )");

    Pdr_ManSolve( pAig, pPars );    

    if pAig.vSeqModelVec {
        pCex = (Abc_Cex_t *)Vec_PtrEntry( pAig.vSeqModelVec, targetCSPropertyIndex );
        if pCex == nil {
            RetValue = 1;
        
        } else {
            RetValue = 0; }
    
    } else {
        RetValue = -1;
        os.Exit(int(0));
    }

    free(fileName);

    for  i=0; i<Saig_ManPoNum(pAig); i++  {
        pObjTargetPo = Aig_ManCo( pAig, i );
        Aig_ObjChild0Flip( pObjTargetPo );
    }
    
    Aig_ManStop( pAig );
    return RetValue;
}

//unused function
     
                                     
 
                    
          

                                     
     
                                                                 
                     
             

              
 
      

func collectSafetyInvariantPOIndex(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var i int32;

    Abc_NtkForEachPo( pNtk, pObj, i )
    {
        if strstr( Abc_ObjName( pObj ), "csSafetyInvar_" ) != nil {
            return i; }
    }        

    return -1;
}

func collectUserGivenDisjunctiveMonotoneSignals(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var pObj *Abc_Obj_t;
    var i int32;
    var monotoneVector *Vec_Ptr_t;
    var newIntVector *Vec_Int_t;

    monotoneVector = Vec_PtrAlloc(0);    
    Abc_NtkForEachPo( pNtk, pObj, i )
    {
        if strstr( Abc_ObjName( pObj ), "csLevel1Stabil_" ) != nil {
            newIntVector = createSingletonIntVector(i);
            Vec_PtrPush(monotoneVector, newIntVector);
        }
    }

    if Vec_PtrSize(monotoneVector) > 0 {
        return monotoneVector;
    } else {
        return nil; }

}

func deallocateMasterBarrierDisjunctInt(vMasterBarrierDisjunctsArg *Vec_Ptr_t) {
    var vInt *Vec_Int_t;
    var i int32;

    if vMasterBarrierDisjunctsArg != nil {
        Vec_PtrForEachEntry(Vec_Int_t *, vMasterBarrierDisjunctsArg, vInt, i)
        {    
            Vec_IntFree(vInt);
        }
        Vec_PtrFree(vMasterBarrierDisjunctsArg);
    }
}

func deallocateMasterBarrierDisjunctVecPtrVecInt(vMasterBarrierDisjunctsArg *Vec_Ptr_t) {
    var vInt *Vec_Int_t;
    var vPtr *Vec_Ptr_t;
    var i, j, k, iElem int32;

    if vMasterBarrierDisjunctsArg != nil {
        Vec_PtrForEachEntry(Vec_Ptr_t *, vMasterBarrierDisjunctsArg, vPtr, i)
        {    
            assert((vPtr != nil), "vPtr");
            Vec_PtrForEachEntry( Vec_Int_t *, vPtr, vInt, j )
            {
                //Vec_IntFree(vInt);
                Vec_IntForEachEntry( vInt, iElem, k )
                    fmt.Printf("%d - ", iElem);
                //printf("Chung Chang j = %d\n", j);
            }
            Vec_PtrFree(vPtr);
        }
        Vec_PtrFree(vMasterBarrierDisjunctsArg);
    }
}

func getVecOfVecFairness(fp *os.File) *Vec_Ptr_t {
    var masterVector *Vec_Ptr_t  = Vec_PtrAlloc(0);
    //Vec_Ptr_t *currSignalVector;
    var stringBuffer [100]byte;
    //int i;
    
    for fgets(stringBuffer, 50, fp) {
        if strstr(stringBuffer, ":") {

        
        } else {
                
        }
    }

    return masterVector;
}


func Abc_CommandCS_kLiveness(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
        var pNtk, pNtkTemp *Abc_Ntk_t;
    var pAig, pAigCS, pAigCSNew *Aig_Man_t;
    var absorberCount int32;
    var absorberLimit int32  = 500;
    var RetValue int32;
    var liveIndex_0 int32 = -1;
    var liveIndex_k int32 = -1;
    var fVerbose int32  = 1;
    var directive int32  = -1;
    var c int32;
    var safetyInvariantPO int32  = -1;
    var beginTime, endTime abctime;
    var time_spent float64;
    var vMasterBarrierDisjuncts *Vec_Ptr_t  = nil;
    var pWorkingAig *Aig_Man_t;
    //FILE *fp;

    pNtk = Abc_FrameReadNtk(pAbc);

    //fp = fopen("propFile.txt", "r");
    //if( fp )
    //    getVecOfVecFairness(fp);
    //exit(0);

    /*************************************************
    Extraction of Command Line Argument    
    *************************************************/
    if argc == 1 {
        assert(directive == -1, "directive == -1");
        directive = SIMPLE_kCS;
    
    } else {
        Extra_UtilGetoptReset();
        for  ( c = Extra_UtilGetopt( argc, argv, "cmCgh" ) ) != -1  {
            switch  c  {
            case 'c':
                directive = kCS_WITH_SAFETY_INVARIANTS; 
                break;
            case 'm':
                directive = kCS_WITH_DISCOVER_MONOTONE_SIGNALS; 
                break;
            case 'C':
                directive = kCS_WITH_SAFETY_AND_DCS_INVARIANTS;
                break;
            case 'g':
                directive = kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS;
                break;
            case 'h':
                goto usage;
                break;
            default:
                goto usage;
            }
        }
    }
    /*************************************************
    Extraction of Command Line Argument Ends    
    *************************************************/

    if !Abc_NtkIsStrash( pNtk ) {
        fmt.Printf("The input network was not strashed, strashing....\n");
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkTemp = pNtk;
    }

    if directive == kCS_WITH_SAFETY_INVARIANTS {
        safetyInvariantPO = collectSafetyInvariantPOIndex(pNtkTemp);            
        assert(safetyInvariantPO != -1, "safetyInvariantPO != -1");
    }

    if directive == kCS_WITH_DISCOVER_MONOTONE_SIGNALS {
        beginTime = Abc_Clock();
        vMasterBarrierDisjuncts = findDisjunctiveMonotoneSignals( pNtk );
        endTime = Abc_Clock();
        time_spent = float64((endTime - beginTime))/int64(time.Second);
        fmt.Printf("pre-processing time = %f\n", time_spent); 
        return 0;
    }

    if directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS {
        safetyInvariantPO = collectSafetyInvariantPOIndex(pNtkTemp);            
        assert(safetyInvariantPO != -1, "safetyInvariantPO != -1");

        beginTime = Abc_Clock();
        vMasterBarrierDisjuncts = findDisjunctiveMonotoneSignals( pNtk );
        endTime = Abc_Clock();
        time_spent = float64((endTime - beginTime))/int64(time.Second);
        fmt.Printf("pre-processing time = %f\n", time_spent); 

        assert(vMasterBarrierDisjuncts != nil, "vMasterBarrierDisjuncts != nil");
        assert(Vec_PtrSize(vMasterBarrierDisjuncts) > 0, "Vec_PtrSize(vMasterBarrierDisjuncts) > 0");
    }

    if directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS {
        safetyInvariantPO = collectSafetyInvariantPOIndex(pNtkTemp);            
        assert(safetyInvariantPO != -1, "safetyInvariantPO != -1");

        beginTime = Abc_Clock();
        vMasterBarrierDisjuncts = collectUserGivenDisjunctiveMonotoneSignals( pNtk );
        endTime = Abc_Clock();
        time_spent = float64((endTime - beginTime))/int64(time.Second);
        fmt.Printf("pre-processing time = %f\n", time_spent); 

        assert(vMasterBarrierDisjuncts != nil, "vMasterBarrierDisjuncts != nil");
        assert(Vec_PtrSize(vMasterBarrierDisjuncts) > 0, "Vec_PtrSize(vMasterBarrierDisjuncts) > 0");
    }

    if directive == kCS_WITH_SAFETY_AND_DCS_INVARIANTS || directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS {
        assert(vMasterBarrierDisjuncts != nil, "vMasterBarrierDisjuncts != nil");
        pWorkingAig = generateWorkingAigWithDSC( pAig, pNtk, &liveIndex_0, vMasterBarrierDisjuncts );
        pAigCS = introduceAbsorberLogic(pWorkingAig, &liveIndex_0, &liveIndex_k, 0);
    
    } else {
        pWorkingAig = generateWorkingAig( pAig, pNtk, &liveIndex_0 );
        pAigCS = introduceAbsorberLogic(pWorkingAig, &liveIndex_0, &liveIndex_k, 0);
    }

    Aig_ManStop(pWorkingAig);

    for  absorberCount=1; absorberCount<absorberLimit; absorberCount++  {
        //printf( "\nindex of the liveness output = %d\n", liveIndex_k );
        RetValue = flipConePdr( pAigCS, directive, liveIndex_k, safetyInvariantPO, absorberCount );

        if RetValue == 1 {
                Abc_Print( 1, "k = %d, Property proved\n", absorberCount );
            break;
        
            } else if RetValue == 0 {
            if fVerbose != 0 {
                Abc_Print( 1, "k = %d, Property DISPROVED\n", absorberCount );
            }
        
            } else if RetValue == -1 {
                Abc_Print( 1, "Property UNDECIDED with k = %d.\n", absorberCount );
        
            } else {
                assert(false, "0"); }

        pAigCSNew = introduceAbsorberLogic(pAigCS, &liveIndex_0, &liveIndex_k, absorberCount);
        Aig_ManStop(pAigCS);
        pAigCS = pAigCSNew;
    }

    Aig_ManStop(pAigCS);
    Aig_ManStop(pAig);

    if directive == kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS {
        deallocateMasterBarrierDisjunctInt(vMasterBarrierDisjuncts);
    
    } else {
        //if(vMasterBarrierDisjuncts)
        //    Vec_PtrFree(vMasterBarrierDisjuncts);
        //deallocateMasterBarrierDisjunctVecPtrVecInt(vMasterBarrierDisjuncts);
        deallocateMasterBarrierDisjunctInt(vMasterBarrierDisjuncts);
    }
    return 0;

    usage:
        fmt.Fprintf(os.Stdout, "usage: kcs [-cmgCh]\n");
            fmt.Fprintf(os.Stdout, "\timplements Claessen-Sorensson's k-Liveness algorithm\n");
        fmt.Fprintf(os.Stdout, "\t-c : verification with constraints, looks for POs prefixed with csSafetyInvar_\n");
        fmt.Fprintf(os.Stdout, "\t-m : discovers monotone signals\n");
            fmt.Fprintf(os.Stdout, "\t-g : verification with user-supplied barriers, looks for POs prefixed with csLevel1Stabil_\n");
        fmt.Fprintf(os.Stdout, "\t-C : verification with discovered monotone signals\n");
        fmt.Fprintf(os.Stdout, "\t-h : print usage\n");
            return 1;

}

func Abc_CommandNChooseK(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk, pNtkTemp, pNtkCombStabil *Abc_Ntk_t;
    var pAig, pAigCombStabil *Aig_Man_t;
    var directive int32  = -1;
    var c int32;
    var parameterizedCombK int32;

    pNtk = Abc_FrameReadNtk(pAbc);


    /*************************************************
    Extraction of Command Line Argument    
    *************************************************/
    if argc == 1 {
        assert(directive == -1, "directive == -1");
        directive = SIMPLE_kCS;
    
    } else {
        Extra_UtilGetoptReset();
        for  ( c = Extra_UtilGetopt( argc, argv, "cmCgh" ) ) != -1  {
            switch  c  {
            case 'c':
                directive = kCS_WITH_SAFETY_INVARIANTS; 
                break;
            case 'm':
                directive = kCS_WITH_DISCOVER_MONOTONE_SIGNALS; 
                break;
            case 'C':
                directive = kCS_WITH_SAFETY_AND_DCS_INVARIANTS;
                break;
            case 'g':
                directive = kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS;
                break;
            case 'h':
                goto usage;
                break;
            default:
                goto usage;
            }
        }
    }
    /*************************************************
    Extraction of Command Line Argument Ends    
    *************************************************/

    if !Abc_NtkIsStrash( pNtk ) {
        fmt.Printf("The input network was not strashed, strashing....\n");
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkTemp = pNtk;
    }

/**********************Code for generation of nCk outputs**/
    //combCount = countCombination(1000, 3);
    //pAigCombStabil = generateDisjunctiveTester( pNtk, pAig, 7, 2 );
    fmt.Printf("Enter parameterizedCombK = ");
    if scanf("%d", &parameterizedCombK) != 1 {
        fmt.Printf("\nFailed to read integer!\n");
        return 0;
    }
    fmt.Printf("\n");

    pAigCombStabil = generateGeneralDisjunctiveTester( pNtk, pAig, parameterizedCombK );
    Aig_ManPrintStats(pAigCombStabil);
    pNtkCombStabil = Abc_NtkFromAigPhase( pAigCombStabil );
    pNtkCombStabil.pName = Abc_UtilStrsav( pAigCombStabil.pName );
    if !Abc_NtkCheck( pNtkCombStabil ) {
            fmt.Fprintf(os.Stdout, "Abc_NtkCreateCone(): Network check has failed.\n"); }
    Abc_FrameSetCurrentNetwork( pAbc, pNtkCombStabil );

    Aig_ManStop( pAigCombStabil );
    Aig_ManStop( pAig );

    return 1;
    //printf("\ncombCount = %d\n", combCount);
    //exit(0);
/**********************************************************/

    usage:
        fmt.Fprintf(os.Stdout, "usage: nck [-cmgCh]\n");
            fmt.Fprintf(os.Stdout, "\tgenerates combinatorial signals for stabilization\n");
        fmt.Fprintf(os.Stdout, "\t-h : print usage\n");
            return 1;

}


//ABC_NAMESPACE_IMPL_END
