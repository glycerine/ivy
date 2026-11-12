package mfs

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [mfsCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Core procedures of this package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfsCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mfsInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Abc_NtkMfsSolveSatResub( Mfs_Man_t * p, Abc_Obj_t * pNode, int iFanin, int fOnlyRemove, int fSkipUpdate );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsParsDefault(pPars *Mfs_Par_t) {
    memset( pPars, 0, sizeof(Mfs_Par_t) );
    pPars.nWinTfoLevs  =    2;
    pPars.nFanoutsMax  =   30;
    pPars.nDepthMax    =   20;
    pPars.nWinMax      =  300;
    pPars.nGrowthLevel =    0;
    pPars.nBTLimit     = 5000;
    pPars.fRrOnly      =    0;
    pPars.fResub       =    1;
    pPars.fArea        =    0;
    pPars.fMoreEffort  =    0;
    pPars.fSwapEdge    =    0;
    pPars.fOneHotness  =    0;
    pPars.fVerbose     =    0;
    pPars.fVeryVerbose =    0;
}
/*
int Abc_NtkMfsEdgePower( Mfs_Man_t * p, Abc_Obj_t * pNode )
{
    Abc_Obj_t * pFanin;
    int i;
    // try replacing area critical fanins
    Abc_ObjForEachFanin( pNode, pFanin, i )
    {
        if ( Abc_MfsObjProb(p, pFanin) >= 0.4 )
        {
            if ( Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 0 ) )
                return 1;
        } else if ( Abc_MfsObjProb(p, pFanin) >= 0.3 )
        {
            if ( Abc_NtkMfsSolveSatResub( p, pNode, i, 1, 0 ) )
                return 1;
        }
    }
    return 0;
}
*/

func Abc_WinNode(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
//    abctime clk;
//    Abc_Obj_t * pFanin;
//    int i;

    p.nNodesTried++;
    // prepare data structure for this node
    Mfs_ManClean( p );
    // compute window roots, window support, and window nodes
    p.vRoots = Abc_MfsComputeRoots( pNode, p.pPars.nWinTfoLevs, p.pPars.nFanoutsMax );
    p.vSupp  = Abc_NtkNodeSupport( p.pNtk, (Abc_Obj_t **)Vec_PtrArray(p.vRoots), Vec_PtrSize(p.vRoots) );
    p.vNodes = Abc_NtkDfsNodes( p.pNtk, (Abc_Obj_t **)Vec_PtrArray(p.vRoots), Vec_PtrSize(p.vRoots) );
    if p.pPars.nWinMax && Vec_PtrSize(p.vNodes) > p.pPars.nWinMax {
        return 1; }
    // compute the divisors of the window
    p.vDivs  = Abc_MfsComputeDivisors( p, pNode, Abc_ObjRequiredLevel(pNode) - 1 );
    p.nTotalDivs += Vec_PtrSize(p.vDivs) - Abc_ObjFaninNum(pNode);
    // construct AIG for the window
    p.pAigWin = Abc_NtkConstructAig( p, pNode );
    // translate it into CNF
    p.pCnf = Cnf_DeriveSimple( p.pAigWin, 1 + Vec_PtrSize(p.vDivs) );
    // create the SAT problem
    p.pSat = Abc_MfsCreateSolverResub( p, nil, 0, 0 );
    if p.pSat == nil {
        p.nNodesBad++;
        return 1;
    }
    return 0;
}

/*
int Abc_NtkMfsPowerResubNode( Mfs_Man_t * p, Abc_Obj_t * pNode )
{
    abctime clk;
    Abc_Obj_t * pFanin;
    int i;

    if (Abc_WinNode(p, pNode)  // something wrong
        return 1;

    // solve the SAT problem
    // Abc_NtkMfsEdgePower( p, pNode );
    // try replacing area critical fanins
    Abc_ObjForEachFanin( pNode, pFanin, i )
        if ( Abc_MfsObjProb(p, pFanin) >= 0.37 && Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 0 ) )
            return 1;

    Abc_ObjForEachFanin( pNode, pFanin, i )
        if ( Abc_MfsObjProb(p, pFanin) >= 0.1 && Abc_NtkMfsSolveSatResub( p, pNode, i, 1, 0 ) )
            return 1;

    if ( Abc_ObjFaninNum(pNode) == p->nFaninMax )
        return 0;

    // try replacing area critical fanins while adding two new fanins
    Abc_ObjForEachFanin( pNode, pFanin, i )
            if ( Abc_MfsObjProb(p, pFanin) >= 0.37 && Abc_NtkMfsSolveSatResub2( p, pNode, i, -1 ) )
                return 1;
        }
    return 0;

    return 1;
}
*/

func Abc_NtkMfsPowerResub(p *Mfs_Man_t, pPars *Mfs_Par_t) {
    var i, k int32;
    var pFanin, pNode *Abc_Obj_t;
    var pNtk *Abc_Ntk_t  = p.pNtk;
    var nFaninMax int32  = Abc_NtkGetFaninMax(p.pNtk);

    // inlining of macro Abc_NtkForEachNode
    for  k = 0; (k < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, k)), 1); k++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if p.pPars.nDepthMax && int32(pNode.Level) > p.pPars.nDepthMax {
            continue; }
        if Abc_ObjFaninNum(pNode) < 2 || Abc_ObjFaninNum(pNode) > nFaninMax {
            continue; }
        if Abc_WinNode(p, pNode) != 0 {  // something wrong
            continue; }

        // solve the SAT problem
        // Abc_NtkMfsEdgePower( p, pNode );
        // try replacing area critical fanins
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
            if Abc_MfsObjProb(p, pFanin) >= 0.35 && Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 0 ) {
                continue; }
        }
    }
    }

    // inlining of macro Abc_NtkForEachNode
    for  k = 0; (k < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, k)), 1); k++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if p.pPars.nDepthMax && int32(pNode.Level) > p.pPars.nDepthMax {
            continue; }
        if Abc_ObjFaninNum(pNode) < 2 || Abc_ObjFaninNum(pNode) > nFaninMax {
            continue; }
        if Abc_WinNode(p, pNode) != 0 {  // something wrong
            continue; }

        // solve the SAT problem
        // Abc_NtkMfsEdgePower( p, pNode );
        // try replacing area critical fanins
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
            if Abc_MfsObjProb(p, pFanin) >= 0.35 && Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 0 ) {
                continue; }
        }
    }
    }

    // inlining of macro Abc_NtkForEachNode
    for  k = 0; (k < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, k)), 1); k++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if p.pPars.nDepthMax && int32(pNode.Level) > p.pPars.nDepthMax {
            continue; }
        if Abc_ObjFaninNum(pNode) < 2 || Abc_ObjFaninNum(pNode) > nFaninMax {
            continue; }
        if Abc_WinNode(p, pNode) != 0 { // something wrong
            continue; }

        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
            if Abc_MfsObjProb(p, pFanin) >= 0.2 && Abc_NtkMfsSolveSatResub( p, pNode, i, 1, 0 ) {
                continue; }
        }
    }
    }
/*
    Abc_NtkForEachNode( pNtk, pNode, k )
    {
        if ( p->pPars->nDepthMax && (int)pNode->Level > p->pPars->nDepthMax )
            continue;
        if ( Abc_ObjFaninNum(pNode) < 2 || Abc_ObjFaninNum(pNode) > nFaninMax - 2)
            continue;
        if (Abc_WinNode(p, pNode) ) // something wrong
            continue;

        Abc_ObjForEachFanin( pNode, pFanin, i )
            if ( Abc_MfsObjProb(p, pFanin) >= 0.37 && Abc_NtkMfsSolveSatResub2( p, pNode, i, -1 ) )
                continue;
    }
*/
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsResub(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
    var clk abctime;
    p.nNodesTried++;
    // prepare data structure for this node
    Mfs_ManClean( p );
    // compute window roots, window support, and window nodes
clk = Abc_Clock();
    p.vRoots = Abc_MfsComputeRoots( pNode, p.pPars.nWinTfoLevs, p.pPars.nFanoutsMax );
    p.vSupp  = Abc_NtkNodeSupport( p.pNtk, (Abc_Obj_t **)Vec_PtrArray(p.vRoots), Vec_PtrSize(p.vRoots) );
    p.vNodes = Abc_NtkDfsNodes( p.pNtk, (Abc_Obj_t **)Vec_PtrArray(p.vRoots), Vec_PtrSize(p.vRoots) );
p.timeWin += Abc_Clock() - clk;
    if p.pPars.nWinMax && Vec_PtrSize(p.vNodes) > p.pPars.nWinMax {
        p.nMaxDivs++;
        return 1;
    }
    // compute the divisors of the window
clk = Abc_Clock();
    p.vDivs  = Abc_MfsComputeDivisors( p, pNode, Abc_ObjRequiredLevel(pNode) - 1 );
    p.nTotalDivs += Vec_PtrSize(p.vDivs) - Abc_ObjFaninNum(pNode);
p.timeDiv += Abc_Clock() - clk;
    // construct AIG for the window
clk = Abc_Clock();
    p.pAigWin = Abc_NtkConstructAig( p, pNode );
p.timeAig += Abc_Clock() - clk;
    // translate it into CNF
clk = Abc_Clock();
    p.pCnf = Cnf_DeriveSimple( p.pAigWin, 1 + Vec_PtrSize(p.vDivs) );
p.timeCnf += Abc_Clock() - clk;
    // create the SAT problem
clk = Abc_Clock();
    p.pSat = Abc_MfsCreateSolverResub( p, nil, 0, 0 );
    if p.pSat == nil {
        p.nNodesBad++;
        return 1;
    }
//clk = Abc_Clock();
//    if ( p->pPars->fGiaSat )
//        Abc_NtkMfsConstructGia( p );
//p->timeGia += Abc_Clock() - clk;
    // solve the SAT problem
    if p.pPars.fPower {
        Abc_NtkMfsEdgePower( p, pNode );
    } else if p.pPars.fSwapEdge {
        Abc_NtkMfsEdgeSwapEval( p, pNode );
    } else {
        Abc_NtkMfsResubNode( p, pNode );
        if p.pPars.fMoreEffort {
            Abc_NtkMfsResubNode2( p, pNode ); }
    }
p.timeSat += Abc_Clock() - clk;
//    if ( p->pPars->fGiaSat )
//        Abc_NtkMfsDeconstructGia( p );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsNode(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
    var pObj *Hop_Obj_t;
    var RetValue int32;
    var dProb float32;
    extern Hop_Obj_t * Abc_NodeIfNodeResyn( Bdc_Man_t * p, Hop_Man_t * pHop, Hop_Obj_t * pRoot, int nVars, Vec_Int_t * vTruth, unsigned * puCare, float dProb );

    var nGain int32;
    var clk abctime;
    p.nNodesTried++;
    // prepare data structure for this node
    Mfs_ManClean( p );
    // compute window roots, window support, and window nodes
clk = Abc_Clock();
    p.vRoots = Abc_MfsComputeRoots( pNode, p.pPars.nWinTfoLevs, p.pPars.nFanoutsMax );
    p.vSupp  = Abc_NtkNodeSupport( p.pNtk, (Abc_Obj_t **)Vec_PtrArray(p.vRoots), Vec_PtrSize(p.vRoots) );
    p.vNodes = Abc_NtkDfsNodes( p.pNtk, (Abc_Obj_t **)Vec_PtrArray(p.vRoots), Vec_PtrSize(p.vRoots) );
p.timeWin += Abc_Clock() - clk;
    // count the number of patterns
//    p->dTotalRatios += Abc_NtkConstraintRatio( p, pNode );
    // construct AIG for the window
clk = Abc_Clock();
    p.pAigWin = Abc_NtkConstructAig( p, pNode );
p.timeAig += Abc_Clock() - clk;
    // translate it into CNF
clk = Abc_Clock();
    p.pCnf = Cnf_DeriveSimple( p.pAigWin, Abc_ObjFaninNum(pNode) );
p.timeCnf += Abc_Clock() - clk;
    // create the SAT problem
clk = Abc_Clock();
    p.pSat = (sat_solver *)Cnf_DataWriteIntoSolver( p.pCnf, 1, 0 );
    if p.pSat && p.pPars.fOneHotness {
        Abc_NtkAddOneHotness( p ); }
    if p.pSat == nil {
        return 0; }
    // solve the SAT problem
    RetValue = Abc_NtkMfsSolveSat( p, pNode );
    p.nTotConfLevel += p.pSat.stats.conflicts;
p.timeSat += Abc_Clock() - clk;
    if RetValue == 0 {
        p.nTimeOutsLevel++;
        p.nTimeOuts++;
        return 0;
    }
    // minimize the local function of the node using bi-decomposition
    assert(p.nFanins == Abc_ObjFaninNum(pNode), "p.nFanins == Abc_ObjFaninNum(pNode)");
    if p.pPars.fPower {
        dProb = ((float *)p.vProbs.pArray)[pNode.Id];
    } else {
        dProb = -1.0;
    }
    pObj = Abc_NodeIfNodeResyn( p.pManDec, (Hop_Man_t *)pNode.pNtk.pManFunc, (Hop_Obj_t *)pNode.pData, p.nFanins, p.vTruth, p.uCare, dProb );
    nGain = Hop_DagSize((Hop_Obj_t *)pNode.pData) - Hop_DagSize(pObj);
    if nGain >= 0 {
        p.nNodesDec++;
        p.nNodesGained += nGain;
        p.nNodesGainedLevel += nGain;
        pNode.pData = pObj;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfs(pNtk *Abc_Ntk_t, pPars *Mfs_Par_t) int32 {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

    var Pars Bdc_Par_t ;
    var pDecPars *Bdc_Par_t = &Pars;
    var pProgress *ProgressBar;
    var p *Mfs_Man_t;
    var pObj *Abc_Obj_t;
    var vLevels *Vec_Vec_t;
    var vNodes *Vec_Ptr_t;
    var i, k, nNodes, nFaninMax int32;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var nTotalNodesBeg int32  = Abc_NtkNodeNum(pNtk);
    var nTotalEdgesBeg int32  = Abc_NtkGetTotalFanins(pNtk);

    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    nFaninMax = Abc_NtkGetFaninMax(pNtk);
    if pPars.fResub {
        if nFaninMax > 8 {
            fmt.Printf("Nodes with more than %d fanins will not be processed.\n", 8);
            nFaninMax = 8;
        }
    
    } else {
        if nFaninMax > MFS_FANIN_MAX {
            fmt.Printf("Nodes with more than %d fanins will not be processed.\n", MFS_FANIN_MAX);
            nFaninMax = MFS_FANIN_MAX;
        }
    }
    // perform the network sweep
//    Abc_NtkSweep( pNtk, 0 );
    // convert into the AIG
    if !Abc_NtkToAig(pNtk) {
        fmt.Fprintf(os.Stdout, "Converting to AIGs has failed.\n");
        return 0;
    }
    assert(Abc_NtkHasAig(pNtk), "Abc_NtkHasAig(pNtk)");

    // start the manager
    p = Mfs_ManAlloc( pPars );
    p.pNtk = pNtk;
    p.nFaninMax = nFaninMax;

    // precomputer power-aware metrics
    if pPars.fPower {
        extern Vec_Int_t * Abc_NtkPowerEstimate( Abc_Ntk_t * pNtk, int fProbOne );
        if pPars.fResub {
            p.vProbs = Abc_NtkPowerEstimate( pNtk, 0 );
        } else {
            p.vProbs = Abc_NtkPowerEstimate( pNtk, 1 ); }
     
                                                                                      
     
        p.TotalSwitchingBeg = Abc_NtkMfsTotalSwitching(pNtk);
      
    }

    if pNtk.pExcare {
        var pTemp *Abc_Ntk_t;
        if Abc_NtkPiNum((Abc_Ntk_t *)pNtk.pExcare) != Abc_NtkCiNum(pNtk) {
            fmt.Printf("The PI count of careset (%d) and logic network (%d) differ. Careset is not used.\n", Abc_NtkPiNum((Abc_Ntk_t *)pNtk.pExcare), Abc_NtkCiNum(pNtk));
        } else {
            pTemp = Abc_NtkStrash( (Abc_Ntk_t *)pNtk.pExcare, 0, 0, 0 );
            p.pCare = Abc_NtkToDar( pTemp, 0, 0 );
            Abc_NtkDelete( pTemp );
            p.vSuppsInv = Aig_ManSupportsInverse( p.pCare );
        }
    }
    if p.pCare != nil {
        fmt.Printf("Performing optimization with %d external care clauses.\n", Aig_ManCoNum(p.pCare)); }
    // prepare the BDC manager
    if !pPars.fResub {
        if nFaninMax < 3 {
            pDecPars.nVarsMax = 3;
        } else {
            pDecPars.nVarsMax = nFaninMax;
        }
        pDecPars.fVerbose = pPars.fVerbose;
        p.vTruth = Vec_IntAlloc( 0 );
        p.pManDec = Bdc_ManAlloc( pDecPars );
    }

    // label the register outputs
    if p.pCare {
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
            pObj.pData = (void *)(ABC_PTRUINT_T)i;
        }
    }

    // compute levels
    Abc_NtkLevel( pNtk );
    Abc_NtkStartReverseLevels( pNtk, pPars.nGrowthLevel );

    // compute don't-cares for each node
    nNodes = 0;
    p.nTotalNodesBeg = nTotalNodesBeg;
    p.nTotalEdgesBeg = nTotalEdgesBeg;
    if pPars.fResub {
     
                                                                               
      
        if pPars.fPower {
            Abc_NtkMfsPowerResub( p, pPars);
        } else {
        pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
            if p.pPars.nDepthMax && int32(pObj.Level) > p.pPars.nDepthMax {
                continue; }
            if Abc_ObjFaninNum(pObj) < 2 || Abc_ObjFaninNum(pObj) > nFaninMax {
                continue; }
            if !p.pPars.fVeryVerbose {
                Extra_ProgressBarUpdate( pProgress, i, nil ); }
            if pPars.fResub {
                Abc_NtkMfsResub( p, pObj );
            } else {
                Abc_NtkMfsNode( p, pObj ); }
        }
        }
        Extra_ProgressBarStop( pProgress );
     
                                                               
      
    }
    } else {
     
                                                                                             
      
        pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkNodeNum(pNtk) );
        vLevels = Abc_NtkLevelize( pNtk );
        // inlining of macro Vec_VecForEachLevelStart
        for  k = 1; (k < Vec_VecSize(vLevels)) && (((vNodes) = Vec_VecEntry(vLevels, k)), 1); k++  {
            if !p.pPars.fVeryVerbose {
                Extra_ProgressBarUpdate( pProgress, nNodes, nil ); }
            p.nNodesGainedLevel = 0;
            p.nTotConfLevel = 0;
            p.nTimeOutsLevel = 0;
            clk2 = Abc_Clock();
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
                if p.pPars.nDepthMax && int32(pObj.Level) > p.pPars.nDepthMax {
                    break; }
                if Abc_ObjFaninNum(pObj) < 2 || Abc_ObjFaninNum(pObj) > nFaninMax {
                    continue; }
                if pPars.fResub {
                    Abc_NtkMfsResub( p, pObj );
                } else {
                    Abc_NtkMfsNode( p, pObj ); }
            }
            nNodes += Vec_PtrSize(vNodes);
            if pPars.fVerbose {
                /*
            printf( "Lev = %2d. Node = %5d. Ave gain = %5.2f. Ave conf = %5.2f. T/o = %6.2f %%  ",
                k, Vec_PtrSize(vNodes),
                1.0*p->nNodesGainedLevel/Vec_PtrSize(vNodes),
                1.0*p->nTotConfLevel/Vec_PtrSize(vNodes),
                100.0*p->nTimeOutsLevel/Vec_PtrSize(vNodes) );
            ABC_PRT( "Time", Abc_Clock() - clk2 );
            */
            }
        }
        Extra_ProgressBarStop( pProgress );
        Vec_VecFree( vLevels );
     
                                                              
      
    }
    Abc_NtkStopReverseLevels( pNtk );

    // perform the sweeping
    if !pPars.fResub {
        extern void Abc_NtkBidecResyn( Abc_Ntk_t * pNtk, int fVerbose );
//        Abc_NtkSweep( pNtk, 0 );
//        Abc_NtkBidecResyn( pNtk, 0 );
    }

    p.nTotalNodesEnd = Abc_NtkNodeNum(pNtk);
    p.nTotalEdgesEnd = Abc_NtkGetTotalFanins(pNtk);

    // undo labesl
    if p.pCare {
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
            pObj.pData = nil;
        }
    }

    if pPars.fPower {
     
        p.TotalSwitchingEnd = Abc_NtkMfsTotalSwitching(pNtk);
//        printf( "Total switching after  = %7.2f.\n", Abc_NtkMfsTotalSwitching(pNtk) );
     
                                                                                      
      
    }

    // free the manager
    p.timeTotal = Abc_Clock() - clk;
    Mfs_ManStop( p );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

