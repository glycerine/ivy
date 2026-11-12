package live

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [disjunctiveMonotone.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Constraint analysis module for the k-Liveness algorithm
        invented by Koen Classen, Niklas Sorensson.]

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

//ABC_NAMESPACE_IMPL_START

type aigPoIndices struct {
    attrPendingSignalIndex int32;
    attrHintSingalBeginningMarker int32;
    attrHintSingalEndMarker int32;
    attrSafetyInvarIndex int32;
}

extern struct aigPoIndices *allocAigPoIndices();
// extern void deallocAigPoIndices(struct aigPoIndices *toBeDeletedAigPoIndices);

// extern int collectSafetyInvariantPOIndex(Abc_Ntk_t *pNtk);


type antecedentConsequentVectorsStruct struct {
    attrAntecedents *Vec_Int_t;
    attrConsequentCandidates *Vec_Int_t;
}

func allocAntecedentConsequentVectorsStruct() *struct antecedentConsequentVectorsStruct {
    struct antecedentConsequentVectorsStruct *newStructure;

    newStructure = (struct antecedentConsequentVectorsStruct *)malloc(sizeof (struct antecedentConsequentVectorsStruct));

    newStructure.attrAntecedents = nil;
    newStructure.attrConsequentCandidates = nil;
        
    assert(newStructure != nil, "newStructure != nil");
    return newStructure;
}

void deallocAntecedentConsequentVectorsStruct(struct antecedentConsequentVectorsStruct *toBeDeleted)
{
    assert(toBeDeleted != nil, "toBeDeleted != nil");
    if toBeDeleted.attrAntecedents { 
        Vec_IntFree( toBeDeleted.attrAntecedents ); }
    if toBeDeleted.attrConsequentCandidates {
        Vec_IntFree( toBeDeleted.attrConsequentCandidates ); }
    free( toBeDeleted );
}

Aig_Man_t *createDisjunctiveMonotoneTester(Aig_Man_t *pAig, struct aigPoIndices *aigPoIndicesArg, 
                    struct antecedentConsequentVectorsStruct *anteConseVectors, int *startMonotonePropPo)
{
    var pNewAig *Aig_Man_t;
    var iElem, i, nRegCount int32;
    var piCopied, liCopied, liCreated, loCopied, loCreated int32;
    var poCopied, poCreated int32;
    var pObj, pObjPo, pObjDriver, pObjDriverNew, pObjPendingDriverNew, pObjPendingAndNextPending *Aig_Obj_t;
    var pPendingFlop, pObjConseCandFlop, pObjSafetyInvariantPoDriver *Aig_Obj_t;
    //Vec_Ptr_t *vHintMonotoneLocalDriverNew;
    var vConseCandFlopOutput *Vec_Ptr_t;
    //Vec_Ptr_t *vHintMonotoneLocalProp;

    var pObjAnteDisjunction, pObjConsecDriver, pObjConsecDriverNew, pObjCandMonotone, pObjPrevCandMonotone, pObjMonotonePropDriver *Aig_Obj_t;
    var vCandMonotoneProp *Vec_Ptr_t;
    var vCandMonotoneFlopInput *Vec_Ptr_t;

    var pendingSignalIndexLocal int32  = aigPoIndicesArg.attrPendingSignalIndex;

    var vAntecedentsLocal *Vec_Int_t  = anteConseVectors.attrAntecedents;
    var vConsequentCandidatesLocal *Vec_Int_t  = anteConseVectors.attrConsequentCandidates;

    if vConsequentCandidatesLocal == nil {
        return nil; } //no candidates for consequent is provided, hence no need to generate a new AIG

    
    //****************************************************************
    // Step1: create the new manager
    // Note: The new manager is created with "2 * Aig_ManObjNumMax(p)"
    // nodes, but this selection is arbitrary - need to be justified
    //****************************************************************
    pNewAig = Aig_ManStart( Aig_ManObjNumMax(pAig) );
    pNewAig.pName = (char *)malloc( len(pAig.pName) + len("_monotone") + 2 );
    sprintf(pNewAig.pName, "%s_%s", pAig.pName, "monotone");
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
    // Step 6: create "D" register output for PENDING flop 
    //****************************************************************
    loCreated++;
    pPendingFlop = Aig_ObjCreateCi( pNewAig );

    //****************************************************************
    // Step 6.a: create "D" register output for HINT_MONOTONE flop 
    //****************************************************************
    vConseCandFlopOutput = Vec_PtrAlloc(Vec_IntSize(vConsequentCandidatesLocal));
    Vec_IntForEachEntry( vConsequentCandidatesLocal, iElem, i )
    {
        loCreated++;
        pObjConseCandFlop = Aig_ObjCreateCi( pNewAig );
        Vec_PtrPush( vConseCandFlopOutput, pObjConseCandFlop );
    }

    nRegCount = loCreated + loCopied;

    //********************************************************************
    // Step 7: create internal nodes
    //********************************************************************
        Aig_ManForEachNode( pAig, pObj, i )
    {
        pObj.pData = Aig_And( pNewAig, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    }
    
    //********************************************************************
    // Step 8: mapping appropriate new flop drivers 
    //********************************************************************

    if aigPoIndicesArg.attrSafetyInvarIndex != -1 {
        pObjPo = Aig_ManCo( pAig, aigPoIndicesArg.attrSafetyInvarIndex );
        pObjDriver = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
        if !Aig_IsComplement(pObjDriver) {
            pObjDriverNew = (Aig_Obj_t *)(Aig_Regular(pObjDriver).pData);
        } else {
            pObjDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(pObjDriver).pData));
        }
        pObjSafetyInvariantPoDriver = pObjDriverNew;
    
    } else {
        pObjSafetyInvariantPoDriver = Aig_ManConst1(pNewAig); }

    pObjPo = Aig_ManCo( pAig, pendingSignalIndexLocal );
    pObjDriver = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
    if !Aig_IsComplement(pObjDriver) {
        pObjPendingDriverNew = (Aig_Obj_t *)(Aig_Regular(pObjDriver).pData);
    } else {
        pObjPendingDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(pObjDriver).pData));
    }

    pObjPendingAndNextPending = Aig_And( pNewAig, pObjPendingDriverNew, pPendingFlop );

    pObjAnteDisjunction = Aig_Not(Aig_ManConst1( pNewAig ));
    if vAntecedentsLocal != nil {
        Vec_IntForEachEntry( vAntecedentsLocal, iElem, i )
        {
            pObjPo = Aig_ManCo( pAig, iElem );
            pObjDriver = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
            if !Aig_IsComplement(pObjDriver) {
                pObjDriverNew = (Aig_Obj_t *)(Aig_Regular(pObjDriver).pData);
            } else {
                pObjDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(pObjDriver).pData));
            }

            pObjAnteDisjunction = Aig_Or( pNewAig, pObjDriverNew, pObjAnteDisjunction );
        }
    }

    vCandMonotoneProp = Vec_PtrAlloc( Vec_IntSize(vConsequentCandidatesLocal) );
    vCandMonotoneFlopInput = Vec_PtrAlloc( Vec_IntSize(vConsequentCandidatesLocal) );
    Vec_IntForEachEntry( vConsequentCandidatesLocal, iElem, i )
    {
        pObjPo = Aig_ManCo( pAig, iElem );
        pObjConsecDriver = Aig_NotCond((Aig_Obj_t *)Aig_ObjFanin0(pObjPo), Aig_ObjFaninC0(pObjPo));
        if !Aig_IsComplement(pObjConsecDriver) {
            pObjConsecDriverNew = (Aig_Obj_t *)(Aig_Regular(pObjConsecDriver).pData);
        } else {
            pObjConsecDriverNew = Aig_Not((Aig_Obj_t *)(Aig_Regular(pObjConsecDriver).pData));
        }

        pObjCandMonotone = Aig_Or( pNewAig, pObjConsecDriverNew, pObjAnteDisjunction );
        pObjPrevCandMonotone = (Aig_Obj_t *)(Vec_PtrEntry( vConseCandFlopOutput, i ));
        pObjMonotonePropDriver = Aig_Or( pNewAig, Aig_Not( Aig_And( pNewAig, pObjPendingAndNextPending, pObjPrevCandMonotone ) ), 
                        pObjCandMonotone );

        //Conjunting safety invar
        pObjMonotonePropDriver = Aig_And( pNewAig, pObjMonotonePropDriver, pObjSafetyInvariantPoDriver );
                        
        Vec_PtrPush( vCandMonotoneFlopInput, pObjCandMonotone );
        Vec_PtrPush( vCandMonotoneProp, pObjMonotonePropDriver );
    }

    //********************************************************************
    // Step 9: create primary outputs 
    //********************************************************************
    Saig_ManForEachPo( pAig, pObj, i )
    {
        poCopied++;
        pObj.pData = Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) ); 
    }

    *startMonotonePropPo = i;
    Vec_PtrForEachEntry( Aig_Obj_t *, vCandMonotoneProp, pObj, i )
    {
        poCreated++;
        pObjPo = Aig_ObjCreateCo( pNewAig, pObj ); 
    }

    //********************************************************************
    // Step 9: create latch inputs 
    //********************************************************************

    Saig_ManForEachLi( pAig, pObj, i )
    {
        liCopied++;
        Aig_ObjCreateCo( pNewAig, Aig_ObjChild0Copy(pObj) );
    }
    
    //********************************************************************
    // Step 9.a: create latch input for PENDING_FLOP
    //********************************************************************

    liCreated++;
    Aig_ObjCreateCo( pNewAig, pObjPendingDriverNew );

    //********************************************************************
    // Step 9.b: create latch input for MONOTONE_FLOP
    //********************************************************************

    Vec_PtrForEachEntry( Aig_Obj_t *, vCandMonotoneFlopInput, pObj, i )
    {
        liCreated++;
        Aig_ObjCreateCo( pNewAig, pObj );
    }

    Aig_ManSetRegNum( pNewAig, nRegCount );
    Aig_ManCleanup( pNewAig );
    
    assert(Aig_ManCheck( pNewAig ), "Aig_ManCheck( pNewAig )");
    assert(loCopied + loCreated == liCopied + liCreated, "loCopied + loCreated == liCopied + liCreated");

    Vec_PtrFree(vConseCandFlopOutput);
    Vec_PtrFree(vCandMonotoneProp);
    Vec_PtrFree(vCandMonotoneFlopInput);
    return pNewAig;
}

Vec_Int_t *findNewDisjunctiveMonotone( Aig_Man_t *pAig, struct aigPoIndices *aigPoIndicesArg, struct antecedentConsequentVectorsStruct *anteConseVectors )
{
    var pAigNew *Aig_Man_t;
    var pObjTargetPo *Aig_Obj_t;
    var poMarker int32;
    //int i, RetValue, poSerialNum;
    var i, poSerialNum int32;
    var Pars Pdr_Par_t;
    var pPars *Pdr_Par_t = &Pars;
    //Abc_Cex_t * pCex = NULL;
    var vMonotoneIndex *Vec_Int_t;
    //char fileName[20];
    var cexElem *Abc_Cex_t;

    var pendingSignalIndexLocal int32  = aigPoIndicesArg.attrPendingSignalIndex;

    pAigNew = createDisjunctiveMonotoneTester(pAig, aigPoIndicesArg, anteConseVectors, &poMarker );

    //printf("enter an integer : ");
    //waitForInteger = getchar();
    //putchar(waitForInteger);

    vMonotoneIndex = Vec_IntAlloc(0);

    for  i=0; i<Saig_ManPoNum(pAigNew); i++  {
        pObjTargetPo = Aig_ManCo( pAigNew, i );
        Aig_ObjChild0Flip( pObjTargetPo );
    }

    Pdr_ManSetDefaultParams( pPars );
    pPars.fVerbose = 0;
    pPars.fNotVerbose = 1;
    pPars.fSolveAll = 1;
    pAigNew.vSeqModelVec = nil;
    Pdr_ManSolve( pAigNew, pPars );    

    if pAigNew.vSeqModelVec {
        Vec_PtrForEachEntry( Abc_Cex_t *, pAigNew.vSeqModelVec, cexElem, i )
        {
            if cexElem == nil  && i >= pendingSignalIndexLocal + 1 {
                poSerialNum = i - (pendingSignalIndexLocal + 1);
                Vec_IntPush( vMonotoneIndex, Vec_IntEntry( anteConseVectors.attrConsequentCandidates, poSerialNum ));
            }
        }
    }
    for  i=0; i<Saig_ManPoNum(pAigNew); i++  {
        pObjTargetPo = Aig_ManCo( pAigNew, i );
        Aig_ObjChild0Flip( pObjTargetPo );
    }
    
    //if(pAigNew->vSeqModelVec)
    //    Vec_PtrFree(pAigNew->vSeqModelVec);

    Aig_ManStop(pAigNew);
    
    if Vec_IntSize( vMonotoneIndex ) > 0 {
        return vMonotoneIndex;
    
    } else {
        Vec_IntFree(vMonotoneIndex);
        return nil;
    }
}

Vec_Int_t *updateAnteConseVectors(struct antecedentConsequentVectorsStruct *anteConse )
{
    var vCandMonotone *Vec_Int_t;
    var iElem, i int32;

    //if( vKnownMonotone == NULL || Vec_IntSize(vKnownMonotone) <= 0 )
    //    return vHintMonotone;
    if anteConse.attrAntecedents == nil  || Vec_IntSize(anteConse.attrAntecedents) <= 0 {
        return anteConse.attrConsequentCandidates; }
    vCandMonotone = Vec_IntAlloc(0);
    Vec_IntForEachEntry( anteConse.attrConsequentCandidates, iElem, i )
    {
        if Vec_IntFind( anteConse.attrAntecedents, iElem ) == -1 {
            Vec_IntPush( vCandMonotone, iElem ); }
    }
    
    return vCandMonotone;
}

func vectorDifference(A *Vec_Int_t, B *Vec_Int_t) *Vec_Int_t {
    var C *Vec_Int_t;
    var iElem, i int32;

    C = Vec_IntAlloc(0);
    Vec_IntForEachEntry( A, iElem, i )
    {
        if Vec_IntFind( B, iElem ) == -1 {
            Vec_IntPush( C, iElem );
        }
    }

    return C;
}

func createSingletonIntVector(iElem int32) *Vec_Int_t {
    var myVec *Vec_Int_t  = Vec_IntAlloc(1);
    
    Vec_IntPush(myVec, iElem);
    return myVec;
}

     
                                           
 
                        
                             
                                  
     
                                   
                                         
                               
     
 
      

     
                                                             
                                                       
                                                                             
                       
 
                                  
                          
                             
                 

                         
                                                                                                      
                                                                            
                    
    
                                                                  
    
                                                       
     
                                                       
                                                 
     

                   
     
                                                           
         
                                                                       
            
            
         
     

                         
 
      

func Vec_IntPushUniqueLocal(p *Vec_Int_t, Entry int32) int32 {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if p.pArray[i] == Entry {
            return 1; }
    }
    Vec_IntPush( p, Entry );
    return 0;
}

Vec_Ptr_t *findNextLevelDisjunctiveMonotone( 
        Aig_Man_t *pAig, 
        struct aigPoIndices *aigPoIndicesInstance, 
        struct antecedentConsequentVectorsStruct *anteConsecInstance, 
        Vec_Ptr_t *previousMonotoneVectors )
{
    var newLevelPtrVec *Vec_Ptr_t;
    var vElem, vNewDisjunctVector, newDisjunction *Vec_Int_t;
    var i, j, iElem int32;
    struct antecedentConsequentVectorsStruct *anteConsecInstanceLocal;
    var vUnionPrevMonotoneVector, vDiffVector *Vec_Int_t;

    newLevelPtrVec = Vec_PtrAlloc(0);
    vUnionPrevMonotoneVector = Vec_IntAlloc(0);
    Vec_PtrForEachEntry(Vec_Int_t *, previousMonotoneVectors, vElem, i)
        Vec_IntForEachEntry( vElem, iElem, j )
            Vec_IntPushUniqueLocal( vUnionPrevMonotoneVector, iElem );

    Vec_PtrForEachEntry(Vec_Int_t *, previousMonotoneVectors, vElem, i)
    {
        anteConsecInstanceLocal = allocAntecedentConsequentVectorsStruct();

        anteConsecInstanceLocal.attrAntecedents = Vec_IntDup(vElem);    
        vDiffVector = vectorDifference( anteConsecInstance.attrConsequentCandidates, vUnionPrevMonotoneVector);
        anteConsecInstanceLocal.attrConsequentCandidates = vDiffVector;
        assert((vDiffVector != nil), "vDiffVector");

        //printf("Calling target function %d\n", i);
        vNewDisjunctVector = findNewDisjunctiveMonotone( pAig, aigPoIndicesInstance, anteConsecInstanceLocal );

        if vNewDisjunctVector {
            Vec_IntForEachEntry(vNewDisjunctVector, iElem, j)
            {
                newDisjunction = Vec_IntDup(vElem);
                Vec_IntPush( newDisjunction, iElem );
                Vec_PtrPush( newLevelPtrVec, newDisjunction );
            }
            Vec_IntFree(vNewDisjunctVector);
        }
        deallocAntecedentConsequentVectorsStruct( anteConsecInstanceLocal );
    }

    Vec_IntFree(vUnionPrevMonotoneVector);

    return newLevelPtrVec;
}

func printAllIntVectors(vDisjunctions *Vec_Ptr_t, pNtk *Abc_Ntk_t, fileName string) {
    var vElem *Vec_Int_t;
    var i, j, iElem int32;
    var name, hintSubStr string;
    var fp *os.File;

    fp = os.OpenFile(fileName, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666);

    Vec_PtrForEachEntry(Vec_Int_t *, vDisjunctions, vElem, i)
    {    
        fmt.Fprintf(fp, "( ");
        Vec_IntForEachEntry( vElem, iElem, j )
        {
            name = Abc_ObjName( Abc_NtkPo(pNtk, iElem));
            hintSubStr = strstr( name, "hint");
            assert(hintSubStr, "hintSubStr");
            fmt.Fprintf(fp, "%s", hintSubStr);
            if j < Vec_IntSize(vElem) - 1 {
                fmt.Fprintf(fp, " || ");
            
            } else {
                fmt.Fprintf(fp, " )\n");
            }
        }
    }
    _ = (fp).Close();
}

func printAllIntVectorsStabil(vDisjunctions *Vec_Ptr_t, pNtk *Abc_Ntk_t, fileName string) {
    var vElem *Vec_Int_t;
    var i, j, iElem int32;
    var name, hintSubStr string;
    var fp *os.File;

    fp = os.OpenFile(fileName, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666);

    Vec_PtrForEachEntry(Vec_Int_t *, vDisjunctions, vElem, i)
    {    
        fmt.Printf("INT[%d] : ( ", i);
        fmt.Fprintf(fp, "( ");
        Vec_IntForEachEntry( vElem, iElem, j )
        {
            name = Abc_ObjName( Abc_NtkPo(pNtk, iElem));
            hintSubStr = strstr( name, "csLevel1Stabil");
            assert(hintSubStr, "hintSubStr");
            fmt.Printf("%s", hintSubStr);
            fmt.Fprintf(fp, "%s", hintSubStr);
            if j < Vec_IntSize(vElem) - 1 {
                fmt.Printf(" || ");
                fmt.Fprintf(fp, " || ");
            
            } else {
                fmt.Printf(" )\n");
                fmt.Fprintf(fp, " )\n");
            }
        }
        //printf(")\n");
    }
    _ = (fp).Close();
}


func appendVecToMasterVecInt(masterVec *Vec_Ptr_t, candVec *Vec_Ptr_t) {
    var i int32;
    var vCand *Vec_Int_t;
    var vNewIntVec *Vec_Int_t;

    assert(masterVec != nil, "masterVec != nil");
    assert(candVec != nil, "candVec != nil");
    Vec_PtrForEachEntry( Vec_Int_t *, candVec, vCand, i )
    {
        vNewIntVec = Vec_IntDup(vCand);    
        Vec_PtrPush(masterVec, vNewIntVec);
    }
}

func deallocateVecOfIntVec(vecOfIntVec *Vec_Ptr_t) {
    var vInt *Vec_Int_t;
    var i int32;

    if vecOfIntVec != nil {
        Vec_PtrForEachEntry( Vec_Int_t *, vecOfIntVec, vInt, i )
        {
            Vec_IntFree( vInt );
        }
        Vec_PtrFree(vecOfIntVec);
    }    
}

func findDisjunctiveMonotoneSignals(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var pAig *Aig_Man_t;
    var vCandidateMonotoneSignals *Vec_Int_t;
    var vKnownMonotoneSignals *Vec_Int_t;
    //Vec_Int_t *vKnownMonotoneSignalsRoundTwo;
    //Vec_Int_t *vOldConsequentVector;
    //Vec_Int_t *vRemainingConsecVector;
    var i int32;
    var iElem int32;
    var pendingSignalIndex int32;
    var pNtkTemp *Abc_Ntk_t;
    var hintSingalBeginningMarker int32;
    var hintSingalEndMarker int32;
    struct aigPoIndices *aigPoIndicesInstance;
    //struct monotoneVectorsStruct *monotoneVectorsInstance;
    struct antecedentConsequentVectorsStruct *anteConsecInstance;
    //Aig_Obj_t *safetyDriverNew;
    var newIntVec *Vec_Int_t;
    var levelOneMonotne, levelTwoMonotne *Vec_Ptr_t;
    //Vec_Ptr_t *levelThreeMonotne;

    var vMasterDisjunctions *Vec_Ptr_t;

    extern int findPendingSignal(Abc_Ntk_t *pNtk);
    extern Vec_Int_t *findHintOutputs(Abc_Ntk_t *pNtk);    
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

    //system("rm monotone.dat");
    
    /*******************************************/    
    //Finding the PO index of the pending signal 
    /*******************************************/    
    pendingSignalIndex = findPendingSignal(pNtk);
    if pendingSignalIndex == -1 {
        fmt.Printf("\nNo Pending Signal Found\n");
        return nil;
    }
    //else
        //printf("Po[%d] = %s\n", pendingSignalIndex, Abc_ObjName( Abc_NtkPo(pNtk, pendingSignalIndex) ) );

    /*******************************************/    
    //Finding the PO indices of all hint signals
    /*******************************************/    
    vCandidateMonotoneSignals = findHintOutputs(pNtk);
    if vCandidateMonotoneSignals == nil {
        return nil;
    } else {
        //Vec_IntForEachEntry( vCandidateMonotoneSignals, iElem, i )
        //    printf("Po[%d] = %s\n", iElem, Abc_ObjName( Abc_NtkPo(pNtk, iElem) ) );
        hintSingalBeginningMarker = Vec_IntEntry( vCandidateMonotoneSignals, 0 );
        hintSingalEndMarker = Vec_IntEntry( vCandidateMonotoneSignals, Vec_IntSize(vCandidateMonotoneSignals) - 1 );
    }

    /**********************************************/
    //Allocating "struct" with necessary parameters
    /**********************************************/
    aigPoIndicesInstance = allocAigPoIndices();
    aigPoIndicesInstance.attrPendingSignalIndex = pendingSignalIndex;
    aigPoIndicesInstance.attrHintSingalBeginningMarker = hintSingalBeginningMarker;
    aigPoIndicesInstance.attrHintSingalEndMarker = hintSingalEndMarker;
    aigPoIndicesInstance.attrSafetyInvarIndex = collectSafetyInvariantPOIndex(pNtk);

    /****************************************************/
    //Allocating "struct" with necessary monotone vectors
    /****************************************************/
    anteConsecInstance = allocAntecedentConsequentVectorsStruct();
    anteConsecInstance.attrAntecedents = nil;
    anteConsecInstance.attrConsequentCandidates = vCandidateMonotoneSignals;

    /*******************************************/    
    //Generate AIG from Ntk
    /*******************************************/    
    if !Abc_NtkIsStrash( pNtk ) {
        pNtkTemp = Abc_NtkStrash( pNtk, 0, 0, 0 );
        pAig = Abc_NtkToDar( pNtkTemp, 0, 1 );
    
    } else {
        pAig = Abc_NtkToDar( pNtk, 0, 1 );
        pNtkTemp = pNtk;
    }

    /*******************************************/    
    //finding LEVEL 1 monotone signals
    /*******************************************/    
    //printf("Calling target function outside loop\n");
    vKnownMonotoneSignals = findNewDisjunctiveMonotone( pAig, aigPoIndicesInstance, anteConsecInstance );
    levelOneMonotne = Vec_PtrAlloc(0);
    Vec_IntForEachEntry( vKnownMonotoneSignals, iElem, i )
    {
        newIntVec = createSingletonIntVector( iElem );
        Vec_PtrPush( levelOneMonotne, newIntVec );
        //printf("Monotone Po[%d] = %s\n", iElem, Abc_ObjName( Abc_NtkPo(pNtk, iElem) ) );
    }
    //printAllIntVectors( levelOneMonotne, pNtk, "monotone.dat" );

    vMasterDisjunctions = Vec_PtrAlloc( Vec_PtrSize( levelOneMonotne ));
    appendVecToMasterVecInt(vMasterDisjunctions, levelOneMonotne );

    /*******************************************/    
    //finding LEVEL >1 monotone signals
    /*******************************************/    
         
                               
     
                                                              
         
                                                                                         
                                                                                                                   
                                                                                         
                                                                                    
                                                                                
                                                                                
                                                                                        
             
                                                                                      
             
                                                                                                                         
                                                                                
             
                                                                                                                                 
             
                                                       
                                                             
                                                                                    
             
                                                                          
                                                                                    
             
         
     
           

     
    levelTwoMonotne = findNextLevelDisjunctiveMonotone( pAig, aigPoIndicesInstance, anteConsecInstance, levelOneMonotne );
    //printAllIntVectors( levelTwoMonotne, pNtk, "monotone.dat" );
    appendVecToMasterVecInt(vMasterDisjunctions, levelTwoMonotne );
      

    //levelThreeMonotne = findNextLevelDisjunctiveMonotone( pAig, aigPoIndicesInstance, anteConsecInstance, levelTwoMonotne );
    //printAllIntVectors( levelThreeMonotne );
    //printAllIntVectors( levelTwoMonotne, pNtk, "monotone.dat" );
    //appendVecToMasterVecInt(vMasterDisjunctions, levelThreeMonotne );

    deallocAigPoIndices(aigPoIndicesInstance);
    deallocAntecedentConsequentVectorsStruct(anteConsecInstance);
    //deallocPointersToMonotoneVectors(monotoneVectorsInstance);

    deallocateVecOfIntVec( levelOneMonotne );
    deallocateVecOfIntVec( levelTwoMonotne );

    Aig_ManStop(pAig);
    Vec_IntFree(vKnownMonotoneSignals);

    return vMasterDisjunctions;
}

//ABC_NAMESPACE_IMPL_END
