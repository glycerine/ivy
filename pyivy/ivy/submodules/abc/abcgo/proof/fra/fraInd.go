package fra

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [fraInd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Inductive prover.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraInd.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"
//#include "sat/cnf/cnf.h"
//#include "opt/dar/dar.h"
//#include "aig/saig/saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs AIG rewriting on the constraint manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigInductionRewrite(p *Fra_Man_t) {
    var pTemp *Aig_Man_t;
    var pObj, pObjPo *Aig_Obj_t;
    var nTruePis, k, i int32;
    var clk abctime  = Abc_Clock();
    // perform AIG rewriting on the speculated frames
//    pTemp = Dar_ManRwsat( pTemp, 1, 0 );
    pTemp = Dar_ManRewriteDefault( p.pManFraig );
//    printf( "Before = %6d.  After = %6d.\n", Aig_ManNodeNum(p->pManFraig), Aig_ManNodeNum(pTemp) ); 
//Aig_ManDumpBlif( p->pManFraig, "1.blif", NULL, NULL );
//Aig_ManDumpBlif( pTemp, "2.blif", NULL, NULL );
//    Fra_FramesWriteCone( pTemp );
//    Aig_ManStop( pTemp );
    // transfer PI/register pointers
    assert(p.pManFraig.nRegs == pTemp.nRegs, "p.pManFraig.nRegs == pTemp.nRegs");
    assert(p.pManFraig.nAsserts == pTemp.nAsserts, "p.pManFraig.nAsserts == pTemp.nAsserts");
    nTruePis = Aig_ManCiNum(p.pManAig) - Aig_ManRegNum(p.pManAig);
    memset( p.pMemFraig, 0, sizeof(Aig_Obj_t *) * p.nSizeAlloc * p.nFramesAll );
    Fra_ObjSetFraig( Aig_ManConst1(p.pManAig), p.pPars.nFramesK, Aig_ManConst1(pTemp) );
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p.pManAig)-Aig_ManRegNum(p.pManAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        Fra_ObjSetFraig( pObj, p.pPars.nFramesK, Aig_ManCi(pTemp,nTruePis*p.pPars.nFramesK+i) );
    }
    k = 0;
    assert(Aig_ManRegNum(p.pManAig) == Aig_ManCoNum(pTemp) - pTemp.nAsserts, "Aig_ManRegNum(p.pManAig) == Aig_ManCoNum(pTemp) - pTemp.nAsserts");
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p.pManAig)-Aig_ManRegNum(p.pManAig); (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        pObjPo = Aig_ManCo(pTemp, pTemp.nAsserts + k++);
        Fra_ObjSetFraig( pObj, p.pPars.nFramesK, Aig_ObjChild0(pObjPo) );
    }
    // exchange
    Aig_ManStop( p.pManFraig );
    p.pManFraig = pTemp;
p.timeRwr += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    [Performs speculative reduction for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prepares the inductive case with speculative reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prepares the inductive case with speculative reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FramesAddMore(p *Aig_Man_t, nFrames int32) {
    var pObj *Aig_Obj_t;
    var pLatches **Aig_Obj_t;
    var i, k, f, nNodesOld int32;
    // set copy pointer of each object to point to itself
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = pObj; }
    }
    // iterate and add objects
    nNodesOld = Aig_ManObjNumMax(p);
    // inlining of macro ABC_ALLOC
    pLatches = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManRegNum(p))));
    for  f = 0; f < nFrames; f++  {
        // clean latch inputs and outputs
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            pObj.pData = nil;
        }
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pObj.pData = nil;
        }
        // save the latch input values
        k = 0;
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if Aig_ObjFanin0(pObj).pData {
                pLatches[k++] = Aig_ObjChild0Copy(pObj);
            } else {
                pLatches[k++] = nil; }
        }
        // insert them as the latch output values
        k = 0;
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pObj.pData = pLatches[k]; k++;
        }
        // create the next time frame of nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            if i > nNodesOld {
                break; }
            if Aig_ObjFanin0(pObj).pData && Aig_ObjFanin1(pObj).pData {
                pObj.pData = Aig_And( p, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
            } else {
                pObj.pData = nil; }
        }
        }
    }
    pLatches = nil;
}


/**Function*************************************************************

  Synopsis    [Performs partitioned sequential SAT sweepingG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigInductionPart(pAig *Aig_Man_t, pPars *Fra_Ssw_t) *Aig_Man_t {
    var fPrintParts int32;
    var Buffer string // was [100]byte in C

    var pTemp, pNew *Aig_Man_t;
    var vResult *Vec_Ptr_t;
    var vPart *Vec_Int_t;
    var pMapBack *int32;
    var i, nCountPis, nCountRegs int32;
    var nClasses, nPartSize, fVerbose int32;
    var clk abctime  = Abc_Clock();

    // save parameters
    nPartSize = pPars.nPartSize; pPars.nPartSize = 0;
    fVerbose  = pPars.fVerbose;  pPars.fVerbose = 0;
    // generate partitions
    if pAig.vClockDoms {
        // divide large clock domains into separate partitions
        vResult = Vec_PtrAlloc( 100 );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize((Vec_Ptr_t *)pAig.vClockDoms)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry((Vec_Ptr_t *)pAig.vClockDoms, i)), 1); i++  {
            if nPartSize && Vec_IntSize(vPart) > nPartSize {
                Aig_ManPartDivide( vResult, vPart, nPartSize, pPars.nOverSize );
            } else {
                Vec_PtrPush( vResult, Vec_IntDup(vPart) ); }
        }
    
    } else {
        vResult = Aig_ManRegPartitionSimple( pAig, nPartSize, pPars.nOverSize ); }
//    vResult = Aig_ManPartitionSmartRegisters( pAig, nPartSize, 0 ); 
//    vResult = Aig_ManRegPartitionSmart( pAig, nPartSize );
    if fPrintParts != 0 {
        // print partitions
        fmt.Printf("Simple partitioning. %d partitions are saved:\n", Vec_PtrSize(vResult));
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vResult)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vResult, i)), 1); i++  {
            Buffer = fmt.Sprintf("part%03d.aig", i);
            pTemp = Aig_ManRegCreatePart( pAig, vPart, &nCountPis, &nCountRegs, nil );
            Ioa_WriteAiger( pTemp, Buffer, 0, 0 );
            fmt.Printf("part%03d.aig : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d.\n", i, Vec_IntSize(vPart), Aig_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Aig_ManNodeNum(pTemp));
            Aig_ManStop( pTemp );
        }
    }

    // perform SSW with partitions
    Aig_ManReprStart( pAig, Aig_ManObjNumMax(pAig) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vResult)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vResult, i)), 1); i++  {
        pTemp = Aig_ManRegCreatePart( pAig, vPart, &nCountPis, &nCountRegs, &pMapBack );
        // create the projection of 1-hot registers
        if pAig.vOnehots {
            pTemp.vOnehots = Aig_ManRegProjectOnehots( pAig, pTemp, pAig.vOnehots, fVerbose ); }
        // run SSW
        pNew = Fra_FraigInduction( pTemp, pPars );
        nClasses = Aig_TransferMappedClasses( pAig, pTemp, pMapBack );
        if fVerbose != 0 {
            fmt.Printf("%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. It = %3d. Cl = %5d.\n", i, Vec_IntSize(vPart), Aig_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Aig_ManNodeNum(pTemp), pPars.nIters, nClasses); }
        Aig_ManStop( pNew );
        Aig_ManStop( pTemp );
        pMapBack = nil;
    }
    // remap the AIG
    pNew = Aig_ManDupRepr( pAig, 0 );
    Aig_ManSeqCleanup( pNew );
//    Aig_ManPrintStats( pAig );
//    Aig_ManPrintStats( pNew );
    Vec_VecFree( (Vec_Vec_t *)vResult );
    pPars.nPartSize = nPartSize;
    pPars.fVerbose = fVerbose;
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs sequential SAT sweeping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigInduction(pManAig *Aig_Man_t, pParams *Fra_Ssw_t) *Aig_Man_t {
    var fUseSimpleCnf int32;
    var fUseOldSimulation int32;
    // other paramaters affecting performance
    // - presence of FRAIGing in Abc_NtkDarSeqSweep()
    // - using distance-1 patterns in Fra_SmlAssignDist1()
    // - the number of simulation patterns
    // - the number of BMC frames

    var p *Fra_Man_t;
    var Pars Fra_Par_t;
    var pPars *Fra_Par_t = &Pars; 
    var pObj *Aig_Obj_t;
    var pCnf *Cnf_Dat_t;
    var pManAigNew *Aig_Man_t  = nil;
    var nNodesBeg, nRegsBeg int32;
    var nIter int32  = -1; // Suppress "might be used uninitialized"
    var i int32;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var TimeToStop abctime;
    if pParams.TimeLimit {
        TimeToStop = pParams.TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        TimeToStop = 0;
    }

    if Aig_ManNodeNum(pManAig) == 0 {
        pParams.nIters = 0;
        // Ntl_ManFinalize() needs the following to satisfy an assertion
        Aig_ManReprStart(pManAig,Aig_ManObjNumMax(pManAig));
        return Aig_ManDupOrdered(pManAig);
    }
    assert(Aig_ManRegNum(pManAig) > 0, "Aig_ManRegNum(pManAig) > 0");
    assert(pParams.nFramesK > 0, "pParams.nFramesK > 0");
//Aig_ManShow( pManAig, 0, NULL );

    if pParams.fWriteImps && pParams.nPartSize > 0 {
        pParams.nPartSize = 0;
        fmt.Printf("Partitioning was disabled to allow implication writing.\n");
    }
    // perform partitioning
    if (pParams.nPartSize > 0 && pParams.nPartSize < Aig_ManRegNum(pManAig))
         || (pManAig.vClockDoms && Vec_VecSize(pManAig.vClockDoms) > 0) {
        return Fra_FraigInductionPart( pManAig, pParams ); }
 
    nNodesBeg = Aig_ManNodeNum(pManAig);
    nRegsBeg  = Aig_ManRegNum(pManAig);

    // enhance the AIG by adding timeframes
//    Fra_FramesAddMore( pManAig, 3 );

    // get parameters
    Fra_ParamsDefaultSeq( pPars );
    pPars.nFramesP   = pParams.nFramesP;
    pPars.nFramesK   = pParams.nFramesK;
    pPars.nMaxImps   = pParams.nMaxImps;
    pPars.nMaxLevs   = pParams.nMaxLevs;
    pPars.fVerbose   = pParams.fVerbose;
    pPars.fRewrite   = pParams.fRewrite;
    pPars.fLatchCorr = pParams.fLatchCorr;
    pPars.fUseImps   = pParams.fUseImps;
    pPars.fWriteImps = pParams.fWriteImps;
    pPars.fUse1Hot   = pParams.fUse1Hot;

    assert(!(pPars.nFramesP > 0 && pPars.fUse1Hot), "!(pPars.nFramesP > 0 && pPars.fUse1Hot)");
    assert(!(pPars.nFramesK > 1 && pPars.fUse1Hot), "!(pPars.nFramesK > 1 && pPars.fUse1Hot)");
 
    // start the fraig manager for this run
    p = Fra_ManStart( pManAig, pPars );
    p.pPars.nBTLimitNode = 0;
    // derive and refine e-classes using K initialized frames
    if fUseOldSimulation != 0 {
        if pPars.nFramesP > 0 {
            pPars.nFramesP = 0;
            fmt.Printf("Fra_FraigInduction(): Prefix cannot be used.\n");
        }
        p.pSml = Fra_SmlStart( pManAig, 0, pPars.nFramesK + 1, pPars.nSimWords );
        Fra_SmlSimulate( p, 1 );
    
    } else {
        // bug:  r iscas/blif/s5378.blif    ; st; ssw -v
        // bug:  r iscas/blif/s1238.blif    ; st; ssw -v
        // refine the classes with more simulation rounds
if pPars.fVerbose {
fmt.Printf("Simulating %d AIG nodes for %d cycles ... ", Aig_ManNodeNum(pManAig), pPars.nFramesP + 32); }
        p.pSml = Fra_SmlSimulateSeq( pManAig, pPars.nFramesP, 32, 1, 1  ); //pPars->nFramesK + 1, 1 );  
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
        Fra_ClassesPrepare( p.pCla, p.pPars.fLatchCorr, p.pPars.nMaxLevs );
//        Fra_ClassesPostprocess( p->pCla );
        // compute one-hotness conditions
        if p.pPars.fUse1Hot {
            p.vOneHots = Fra_OneHotCompute( p, p.pSml ); }
        // allocate new simulation manager for simulating counter-examples
        Fra_SmlStop( p.pSml );
        p.pSml = Fra_SmlStart( pManAig, 0, pPars.nFramesK + 1, pPars.nSimWords );
    }

    // select the most expressive implications
    if pPars.fUseImps {
        p.pCla.vImps = Fra_ImpDerive( p, 5000000, pPars.nMaxImps, pPars.fLatchCorr ); }

    if pParams.TimeLimit != 0.0 && Abc_Clock() > TimeToStop {
        if !pParams.fSilent {
            fmt.Printf("Fra_FraigInduction(): Runtime limit exceeded.\n"); }
        goto finish;
    }

    // perform BMC (for the min number of frames)
    Fra_BmcPerform( p, pPars.nFramesP, pPars.nFramesK+1 ); // +1 is needed to prevent non-refinement
//Fra_ClassesPrint( p->pCla, 1 );
//    if ( p->vCex == NULL )
//        p->vCex = Vec_IntAlloc( 1000 );

    p.nLitsBeg  = Fra_ClassesCountLits( p.pCla );
    p.nNodesBeg = nNodesBeg; // Aig_ManNodeNum(pManAig);
    p.nRegsBeg  = nRegsBeg; // Aig_ManRegNum(pManAig);

    // dump AIG of the timeframes
//    pManAigNew = Fra_ClassesDeriveAig( p->pCla, pPars->nFramesK );
//    Aig_ManDumpBlif( pManAigNew, "frame_aig.blif", NULL, NULL );
//    Fra_ManPartitionTest2( pManAigNew );
//    Aig_ManStop( pManAigNew );
 
    // iterate the inductive case
    p.pCla.fRefinement = 1;
    for  nIter = 0; p.pCla.fRefinement; nIter++  {
        var nLitsOld int32  = Fra_ClassesCountLits(p.pCla);
        var nImpsOld int32;
        if p.pCla.vImps {
            nImpsOld = Vec_IntSize(p.pCla.vImps);
        } else {
            nImpsOld = 0;
        }
        var nHotsOld int32;
        if p.vOneHots {
            nHotsOld = Fra_OneHotCount(p, p.vOneHots);
        } else {
            nHotsOld = 0;
        }
        var clk3 abctime  = Abc_Clock();

        if pParams.TimeLimit != 0.0 && Abc_Clock() > TimeToStop {
            if !pParams.fSilent {
                fmt.Printf("Fra_FraigInduction(): Runtime limit exceeded.\n"); }
            goto finish;
        }

        // mark the classes as non-refined
        p.pCla.fRefinement = 0;
        // derive non-init K-timeframes while implementing e-classes
clk2 = Abc_Clock();
        p.pManFraig = Fra_FramesWithClasses( p );
p.timeTrav += Abc_Clock() - clk2;
//Aig_ManDumpBlif( p->pManFraig, "testaig.blif", NULL, NULL );

        // perform AIG rewriting
        if p.pPars.fRewrite {
            Fra_FraigInductionRewrite( p ); }

        // convert the manager to SAT solver (the last nLatches outputs are inputs)
        if (fUseSimpleCnf != 0) || pPars.fUseImps {
            pCnf = Cnf_DeriveSimple( p.pManFraig, Aig_ManRegNum(p.pManFraig) );
        } else {
            pCnf = Cnf_Derive( p.pManFraig, Aig_ManRegNum(p.pManFraig) ); }
//        Cnf_DataTranformPolarity( pCnf, 0 );
//Cnf_DataWriteIntoFile( pCnf, "temp.cnf", 1 );

        p.pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
        p.nSatVars = pCnf.nVars;
        assert(p.pSat != nil, "p.pSat != nil");
        if p.pSat == nil {
            fmt.Printf("Fra_FraigInduction(): Computed CNF is not valid.\n"); }
        if pPars.fUseImps {
            Fra_ImpAddToSolver( p, p.pCla.vImps, pCnf.pVarNums );
            if p.pSat == nil {
                fmt.Printf("Fra_FraigInduction(): Adding implicationsn to CNF led to a conflict.\n"); }
        }

        // set the pointers to the manager
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pManFraig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManFraig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            pObj.pData = p; }
        }

        // prepare solver for fraiging the last timeframe
        Fra_ManClean( p, Aig_ManObjNumMax(p.pManFraig) + Aig_ManNodeNum(p.pManAig) );

        // transfer PI/LO variable numbers
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pManFraig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManFraig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            if pCnf.pVarNums[pObj.Id] == -1 {
                continue; }
            Fra_ObjSetSatNum( pObj, pCnf.pVarNums[pObj.Id] );
            Fra_ObjSetFaninVec( pObj, (Vec_Ptr_t *)1 );
        }
        }
        Cnf_DataFree( pCnf );

        // add one-hotness clauses
        if p.pPars.fUse1Hot {
            Fra_OneHotAssume( p, p.vOneHots ); }
//        if ( p->pManAig->vOnehots )
//            Fra_OneHotAddKnownConstraint( p, p->pManAig->vOnehots );
 
        // report the intermediate results
        if pPars.fVerbose {
            fmt.Printf("%3d : C = %6d. Cl = %6d.  L = %6d. LR = %6d.  ", nIter, Vec_PtrSize(p.pCla.vClasses1), Vec_PtrSize(p.pCla.vClasses), Fra_ClassesCountLits(p.pCla), p.pManFraig.nAsserts);
            if p.pCla.vImps {
                fmt.Printf("I = %6d. ", Vec_IntSize(p.pCla.vImps)); }
            if p.pPars.fUse1Hot {
                fmt.Printf("1h = %6d. ", Fra_OneHotCount(p, p.vOneHots)); }
            fmt.Printf("NR = %6d. ", Aig_ManNodeNum(p.pManFraig));
//            printf( "\n" );
        } 

        // perform sweeping
        p.nSatCallsRecent = 0;
        p.nSatCallsSkipped = 0;
clk2 = Abc_Clock();
        if p.pPars.fUse1Hot {
            Fra_OneHotCheck( p, p.vOneHots ); }
        Fra_FraigSweep( p );
        if pPars.fVerbose {
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("T"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk3)))/(float64(int64(time.Second))));;
        } 

//        Sat_SolverPrintStats( stdout, p->pSat );
        // remove FRAIG and SAT solver
        Aig_ManStop( p.pManFraig );   p.pManFraig = nil;
//        printf( "Vars = %d. Clauses = %d. Learnts = %d.\n", p->pSat->size, p->pSat->clauses.size, p->pSat->learnts.size );
        sat_solver_delete( p.pSat );  p.pSat = nil; 
        memset( p.pMemFraig, 0, sizeof(Aig_Obj_t *) * p.nSizeAlloc * p.nFramesAll );
//        printf( "Recent SAT called = %d. Skipped = %d.\n", p->nSatCallsRecent, p->nSatCallsSkipped );
        assert(p.vTimeouts == nil, "p.vTimeouts == nil");
        if p.vTimeouts {
           fmt.Printf("Fra_FraigInduction(): SAT solver timed out!\n"); }
        // check if refinement has happened
//        p->pCla->fRefinement = (int)(nLitsOld != Fra_ClassesCountLits(p->pCla));
        if p.pCla.fRefinement && 
            nLitsOld == Fra_ClassesCountLits(p.pCla) && 
            nImpsOld == (p.pCla.vImps? Vec_IntSize(p.pCla.vImps) : 0) && 
            nHotsOld == (p.vOneHots? Fra_OneHotCount(p, p.vOneHots) : 0) {
            fmt.Printf("Fra_FraigInduction(): Internal error. The result may not verify.\n");
            break;
        }
    }
/*
    // verify implications using simulation
    if ( p->pCla->vImps && Vec_IntSize(p->pCla->vImps) )
    {
        int Temp;
        abctime clk = Abc_Clock();
        if ( Temp = Fra_ImpVerifyUsingSimulation( p ) )
            printf( "Implications failing the simulation test = %d (out of %d).  ", Temp, Vec_IntSize(p->pCla->vImps) );
        else
            printf( "All %d implications have passed the simulation test.  ", Vec_IntSize(p->pCla->vImps) );
        ABC_PRT( "Time", Abc_Clock() - clk );
    }
*/

    // move the classes into representatives and reduce AIG
clk2 = Abc_Clock();
    if p.pPars.fWriteImps && p.vOneHots && Fra_OneHotCount(p, p.vOneHots) {
        extern void Ioa_WriteAiger( Aig_Man_t * pMan, char * pFileName, int fWriteSymbols, int fCompact );
        var pNew *Aig_Man_t;
        var pFileName string  = Ioa_FileNameGenericAppend( p.pManAig.pName, "_care.aig" );
        fmt.Printf("Care one-hotness clauses will be written into file \"%s\".\n", pFileName);
        pManAigNew = Aig_ManDupOrdered( pManAig );
        pNew = Fra_OneHotCreateExdc( p, p.vOneHots );
        Ioa_WriteAiger( pNew, pFileName, 0, 1 );
        Aig_ManStop( pNew );
    
    } else {
    //    Fra_ClassesPrint( p->pCla, 1 );
        Fra_ClassesSelectRepr( p.pCla );
        Fra_ClassesCopyReprs( p.pCla, p.vTimeouts );
        pManAigNew = Aig_ManDupRepr( pManAig, 0 );
    }
    // add implications to the manager
//    if ( fWriteImps && p->pCla->vImps && Vec_IntSize(p->pCla->vImps) )
//        Fra_ImpRecordInManager( p, pManAigNew );
    // cleanup the new manager
    Aig_ManSeqCleanup( pManAigNew );
    // remove pointers to the dead nodes
//    Aig_ManForEachObj( pManAig, pObj, i )
//        if ( pObj->pData && Aig_ObjIsNone(pObj->pData) )
//            pObj->pData = NULL;
//    Aig_ManCountMergeRegs( pManAigNew );
p.timeTrav += Abc_Clock() - clk2;
p.timeTotal = Abc_Clock() - clk;
    // get the final stats
    p.nLitsEnd  = Fra_ClassesCountLits( p.pCla );
    p.nNodesEnd = Aig_ManNodeNum(pManAigNew);
    p.nRegsEnd  = Aig_ManRegNum(pManAigNew);
    // free the manager
finish:
    Fra_ManStop( p );
    // check the output
//    if ( Aig_ManCoNum(pManAigNew) - Aig_ManRegNum(pManAigNew) == 1 )
//        if ( Aig_ObjChild0( Aig_ManCo(pManAigNew,0) ) == Aig_ManConst0(pManAigNew) )
//            printf( "Proved output constant 0.\n" );
    pParams.nIters = nIter;
    return pManAigNew;
}

/**Function*************************************************************

  Synopsis    [Outputs a set of pairs of equivalent nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigInductionTest(pFileName string, pParams *Fra_Ssw_t) int32 {
    var pFile *os.File;
    var pFilePairs string;
    var pMan, pNew *Aig_Man_t;
    var pObj, pRepr *Aig_Obj_t;
    var pNum2Id *int32;
    var i, Counter int32;
    pMan = Saig_ManReadBlif( pFileName );
    if pMan == nil {
        return 0; }
    // perform seq SAT sweeping
    pNew = Fra_FraigInduction( pMan, pParams );
    if pNew == nil {
        Aig_ManStop( pMan );
        return 0;
    }
    if pParams.fVerbose {
        fmt.Printf("Original AIG: ");
        Aig_ManPrintStats( pMan );
        fmt.Printf("Reduced  AIG: ");
        Aig_ManPrintStats( pNew );
    }
    Aig_ManStop( pNew );
    pNum2Id = (int *)pMan.pData;
    // write the output file
    pFilePairs = Aig_FileNameGenericAppend( pFileName, ".pairs" );
    pFile = os.OpenFile(pFilePairs, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if pRepr = pMan.pReprs[pObj.Id] {
            fmt.Fprintf(pFile, "%d %d %c\n", pNum2Id[pObj.Id], pNum2Id[pRepr.Id], (Aig_ObjPhase(pObj) ^ Aig_ObjPhase(pRepr))? '-' : '+');
            Counter++;
        }
    }
    _ = (pFile).Close();
    if pParams.fVerbose {
        fmt.Printf("Result: %d pairs of seq equiv nodes are written into file \"%s\".\n", Counter, pFilePairs);
    }
    Aig_ManStop( pMan );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

