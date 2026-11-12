package fra

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [fraCec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [CEC engined based on fraiging.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraCec.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver2.h"

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
func Fra_FraigSat(pMan *Aig_Man_t, nConfLimit ABC_INT64_T, nInsLimit ABC_INT64_T, nLearnedStart int32, nLearnedDelta int32, nLearnedPerce int32, fFlipBits int32, fAndOuts int32, fNewSolver int32, fVerbose int32) int32 {
    if fNewSolver != 0 {
        extern void * Cnf_DataWriteIntoSolver2( Cnf_Dat_t * p, int nFrames, int fInit );
        extern int    Cnf_DataWriteOrClause2( void * pSat, Cnf_Dat_t * pCnf );

        var pSat *sat_solver2;
        var pCnf *Cnf_Dat_t;
        var status, RetValue int32;
        var clk abctime  = Abc_Clock();
        var vCiIds *Vec_Int_t;

        assert(Aig_ManRegNum(pMan) == 0, "Aig_ManRegNum(pMan) == 0");
        pMan.pData = nil;

        // derive CNF
        pCnf = Cnf_Derive( pMan, Aig_ManCoNum(pMan) );
    //    pCnf = Cnf_DeriveSimple( pMan, Aig_ManCoNum(pMan) );

        if fFlipBits != 0 { 
            Cnf_DataTranformPolarity( pCnf, 0 ); }

        if fVerbose != 0 {
            fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. ", pCnf.nVars, pCnf.nClauses, pCnf.nLiterals);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }

        // convert into SAT solver
        pSat = (sat_solver2 *)Cnf_DataWriteIntoSolver2( pCnf, 1, 0 );
        if pSat == nil {
            Cnf_DataFree( pCnf );
            return 1;
        }

        if fAndOuts != 0 {
            // assert each output independently
            if !Cnf_DataWriteAndClauses( pSat, pCnf ) {
                sat_solver2_delete( pSat );
                Cnf_DataFree( pCnf );
                return 1;
            }
        
        } else {
            // add the OR clause for the outputs
            if !Cnf_DataWriteOrClause2( pSat, pCnf ) {
                sat_solver2_delete( pSat );
                Cnf_DataFree( pCnf );
                return 1;
            }
        }
        vCiIds = Cnf_DataCollectPiSatNums( pCnf, pMan );
        Cnf_DataFree( pCnf );


        fmt.Printf("Created SAT problem with %d variable and %d clauses. ", sat_solver2_nvars(pSat), sat_solver2_nclauses(pSat));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;

        // simplify the problem
        clk = Abc_Clock();
        status = sat_solver2_simplify(pSat);
//        printf( "Simplified the problem to %d variables and %d clauses. ", sat_solver2_nvars(pSat), sat_solver2_nclauses(pSat) );
//        ABC_PRT( "Time", Abc_Clock() - clk );
        if status == 0 {
            Vec_IntFree( vCiIds );
            sat_solver2_delete( pSat );
    //        printf( "The problem is UNSATISFIABLE after simplification.\n" );
            return 1;
        }

        // solve the miter
        clk = Abc_Clock();
        if fVerbose != 0 {
            pSat.verbosity = 1; }
        status = sat_solver2_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)nInsLimit, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status == l_Undef {
    //        printf( "The problem timed out.\n" );
            RetValue = -1;
        
        } else if status == l_True {
    //        printf( "The problem is SATISFIABLE.\n" );
            RetValue = 0;
        
        } else if status == l_False {
    //        printf( "The problem is UNSATISFIABLE.\n" );
            RetValue = 1;
        
        } else {
            assert(false, "0"); }

    //    Abc_Print( 1, "The number of conflicts = %6d.  ", (int)pSat->stats.conflicts );
    //    Abc_PrintTime( 1, "Solving time", Abc_Clock() - clk );
 
        // if the problem is SAT, get the counterexample
        if status == l_True {
            pMan.pData = Sat_Solver2GetModel( pSat, vCiIds.pArray, vCiIds.nSize );
        }
        // free the sat_solver2
        if fVerbose != 0 {
            Sat_Solver2PrintStats( os.Stdout, pSat ); }
    //sat_solver2_store_write( pSat, "trace.cnf" );
    //sat_solver2_store_free( pSat );
        sat_solver2_delete( pSat );
        Vec_IntFree( vCiIds );
        return RetValue;
    
    } else {
        var pSat *sat_solver;
        var pCnf *Cnf_Dat_t;
        var status, RetValue int32;
        var clk abctime  = Abc_Clock();
        var vCiIds *Vec_Int_t;

        assert(Aig_ManRegNum(pMan) == 0, "Aig_ManRegNum(pMan) == 0");
        pMan.pData = nil;

        // derive CNF
        pCnf = Cnf_Derive( pMan, Aig_ManCoNum(pMan) );
    //    pCnf = Cnf_DeriveSimple( pMan, Aig_ManCoNum(pMan) );

        if fFlipBits != 0 { 
            Cnf_DataTranformPolarity( pCnf, 0 ); }

        if fVerbose != 0 {
            fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. ", pCnf.nVars, pCnf.nClauses, pCnf.nLiterals);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }

        // convert into SAT solver
        pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
        if pSat == nil {
            Cnf_DataFree( pCnf );
            return 1;
        }

        if nLearnedStart != 0 {
            pSat.nLearntStart = pSat.nLearntMax = nLearnedStart; }
        if nLearnedDelta != 0 {
            pSat.nLearntDelta = nLearnedDelta; }
        if nLearnedPerce != 0 {
            pSat.nLearntRatio = nLearnedPerce; }
        if fVerbose != 0 {
            pSat.fVerbose = fVerbose; }

        if fAndOuts != 0 {
            // assert each output independently
            if !Cnf_DataWriteAndClauses( pSat, pCnf ) {
                sat_solver_delete( pSat );
                Cnf_DataFree( pCnf );
                return 1;
            }
        
        } else {
            // add the OR clause for the outputs
            if !Cnf_DataWriteOrClause( pSat, pCnf ) {
                sat_solver_delete( pSat );
                Cnf_DataFree( pCnf );
                return 1;
            }
        }
        vCiIds = Cnf_DataCollectPiSatNums( pCnf, pMan );
        Cnf_DataFree( pCnf );


    //    printf( "Created SAT problem with %d variable and %d clauses. ", sat_solver_nvars(pSat), sat_solver_nclauses(pSat) );
    //    ABC_PRT( "Time", Abc_Clock() - clk );

        // simplify the problem
        clk = Abc_Clock();
        status = sat_solver_simplify(pSat);
    //    printf( "Simplified the problem to %d variables and %d clauses. ", sat_solver_nvars(pSat), sat_solver_nclauses(pSat) );
    //    ABC_PRT( "Time", Abc_Clock() - clk );
        if status == 0 {
            Vec_IntFree( vCiIds );
            sat_solver_delete( pSat );
    //        printf( "The problem is UNSATISFIABLE after simplification.\n" );
            return 1;
        }

        // solve the miter
        clk = Abc_Clock();
//        if ( fVerbose )
//            pSat->verbosity = 1;
        status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)nInsLimit, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status == l_Undef {
    //        printf( "The problem timed out.\n" );
            RetValue = -1;
        
        } else if status == l_True {
    //        printf( "The problem is SATISFIABLE.\n" );
            RetValue = 0;
        
        } else if status == l_False {
    //        printf( "The problem is UNSATISFIABLE.\n" );
            RetValue = 1;
        
        } else {
            assert(false, "0"); }

    //    Abc_Print( 1, "The number of conflicts = %6d.  ", (int)pSat->stats.conflicts );
    //    Abc_PrintTime( 1, "Solving time", Abc_Clock() - clk );
 
        // if the problem is SAT, get the counterexample
        if status == l_True {
            pMan.pData = Sat_SolverGetModel( pSat, vCiIds.pArray, vCiIds.nSize );
        }
        // free the sat_solver
        if fVerbose != 0 {
            Sat_SolverPrintStats( os.Stdout, pSat ); }
    //sat_solver_store_write( pSat, "trace.cnf" );
    //sat_solver_store_free( pSat );
        sat_solver_delete( pSat );
        Vec_IntFree( vCiIds );
        return RetValue;
    }
}

/**Function*************************************************************

  Synopsis    [Recognizes what nodes are inputs of the EXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCountXors(p *Aig_Man_t) int32 {
    var pObj, pFan0, pFan1 *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjIsMuxType(pObj) && Aig_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            Counter++; }
    }
    return Counter;

}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigCec(ppAig **Aig_Man_t, nConfLimit int32, fVerbose int32) int32 {
    var nBTLimitStart int32  =        300;   // starting SAT run
    var nBTLimitFirst int32  =          2;   // first fraiging iteration
    var nBTLimitLast int32   = nConfLimit;   // the last-gasp SAT run

    var Params Fra_Par_t;
    var pParams *Fra_Par_t = &Params;
    var pAig *Aig_Man_t = *ppAig;
    var pTemp *Aig_Man_t;
    var i, RetValue int32;
    var clk abctime;

    // report the original miter
    if fVerbose != 0 {
        fmt.Printf("Original miter:   Nodes = %6d.\n", Aig_ManNodeNum(pAig));
    }
    RetValue = Fra_FraigMiterStatus( pAig );
//    assert( RetValue == -1 );
    if RetValue == 0 {
        // inlining of macro ABC_ALLOC
        pAig.pData = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManCiNum(pAig))));
        memset( pAig.pData, 0, sizeof(int) * Aig_ManCiNum(pAig) );
        return RetValue;
    }

    // if SAT only, solve without iteration
clk = Abc_Clock();
    RetValue = Fra_FraigSat( pAig, (ABC_INT64_T)2*nBTLimitStart, (ABC_INT64_T)0, 0, 0, 0, 1, 0, 0, 0 );
    if fVerbose != 0 {
        fmt.Printf("Initial SAT:      Nodes = %6d.  ", Aig_ManNodeNum(pAig));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    if RetValue >= 0 {
        return RetValue; }

    // duplicate the AIG
clk = Abc_Clock();
    pAig = Dar_ManRwsat( pTemp = pAig, 1, 0 );
    Aig_ManStop( pTemp );
    if fVerbose != 0 {
        fmt.Printf("Rewriting:        Nodes = %6d.  ", Aig_ManNodeNum(pAig));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

    // perform the loop
    Fra_ParamsDefault( pParams );
    pParams.nBTLimitNode = nBTLimitFirst;
    pParams.nBTLimitMiter = nBTLimitStart;
    pParams.fDontShowBar = 1;
    pParams.fProve = 1;
    for  i = 0; i < 6; i++  {
//printf( "Running fraiging with %d BTnode and %d BTmiter.\n", pParams->nBTLimitNode, pParams->nBTLimitMiter );
        // try XOR balancing
        if Aig_ManCountXors(pAig) * 30 > Aig_ManNodeNum(pAig) + 300 {
clk = Abc_Clock();
            pAig = Dar_ManBalanceXor( pTemp = pAig, 1, 0, 0 );
            Aig_ManStop( pTemp );
            if fVerbose != 0 {
                fmt.Printf("Balance-X:        Nodes = %6d.  ", Aig_ManNodeNum(pAig));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
            } 
        }

        // run fraiging
clk = Abc_Clock();
        pAig = Fra_FraigPerform( pTemp = pAig, pParams );
        Aig_ManStop( pTemp );
        if fVerbose != 0 {
            fmt.Printf("Fraiging (i=%d):   Nodes = %6d.  ", i+1, Aig_ManNodeNum(pAig));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }

        // check the miter status
        RetValue = Fra_FraigMiterStatus( pAig );
        if RetValue >= 0 {
            break; }

        // perform rewriting
clk = Abc_Clock();
        pAig = Dar_ManRewriteDefault( pTemp = pAig );
        Aig_ManStop( pTemp );
        if fVerbose != 0 {
            fmt.Printf("Rewriting:        Nodes = %6d.  ", Aig_ManNodeNum(pAig));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        } 

        // check the miter status
        RetValue = Fra_FraigMiterStatus( pAig );
        if RetValue >= 0 {
            break; }
        // try simulation

        // set the parameters for the next run
        pParams.nBTLimitNode = 8 * pParams.nBTLimitNode;
        pParams.nBTLimitMiter = 2 * pParams.nBTLimitMiter;
    }

    // if still unsolved try last gasp
    if RetValue == -1 {
clk = Abc_Clock();
        RetValue = Fra_FraigSat( pAig, (ABC_INT64_T)nBTLimitLast, (ABC_INT64_T)0, 0, 0, 0, 1, 0, 0, 0 );
        if fVerbose != 0 {
            fmt.Printf("Final SAT:        Nodes = %6d.  ", Aig_ManNodeNum(pAig));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
    }

    *ppAig = pAig;
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigCecPartitioned(pMan1 *Aig_Man_t, pMan2 *Aig_Man_t, nConfLimit int32, nPartSize int32, fSmart int32, fVerbose int32) int32 {
    var pAig *Aig_Man_t;
    var vParts *Vec_Ptr_t;
    var i int32;
    var RetValue int32 = 1;
    var nOutputs int32;
    // create partitions
    vParts = Aig_ManMiterPartitioned( pMan1, pMan2, nPartSize, fSmart );
    // solve the partitions
    nOutputs = -1;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        nOutputs++;
        if fVerbose != 0 {
            fmt.Printf("Verifying part %4d  (out of %4d)  PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, Vec_PtrSize(vParts), Aig_ManCiNum(pAig), Aig_ManCoNum(pAig), Aig_ManNodeNum(pAig), Aig_ManLevelNum(pAig));
            fflush( os.Stdout );
        }
        RetValue = Fra_FraigMiterStatus( pAig );
        if RetValue == 1 {
            continue; }
        if RetValue == 0 {
            break; }
        RetValue = Fra_FraigCec( &pAig, nConfLimit, 0 );
        Vec_PtrWriteEntry( vParts, i, pAig );
        if RetValue == 1 {
            continue; }
        if RetValue == 0 {
            break; }
        break;
    }
    // clear the result
    if fVerbose != 0 {
        fmt.Printf("                                                                                          \r");
        fflush( os.Stdout );
    }
    // report the timeout
    if RetValue == -1 {
        fmt.Printf("Timed out after verifying %d partitions (out of %d).\n", nOutputs, Vec_PtrSize(vParts));
        fflush( os.Stdout );
    }
    // free intermediate results
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        Aig_ManStop( pAig );
    }
    Vec_PtrFree( vParts );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigCecTop(pMan1 *Aig_Man_t, pMan2 *Aig_Man_t, nConfLimit int32, nPartSize int32, fSmart int32, fVerbose int32) int32 {
    var pTemp *Aig_Man_t;
    //Abc_NtkDarCec( pNtk1, pNtk2, fPartition, fVerbose );
    var RetValue int32;
    var clkTotal abctime  = Abc_Clock();

    if Aig_ManCiNum(pMan1) != Aig_ManCiNum(pMan1) {
        fmt.Printf("Abc_CommandAbc8Cec(): Miters have different number of PIs.\n");
        return 0;
    }
    if Aig_ManCoNum(pMan1) != Aig_ManCoNum(pMan1) {
        fmt.Printf("Abc_CommandAbc8Cec(): Miters have different number of POs.\n");
        return 0;
    }
    assert(Aig_ManCiNum(pMan1) == Aig_ManCiNum(pMan1), "Aig_ManCiNum(pMan1) == Aig_ManCiNum(pMan1)");
    assert(Aig_ManCoNum(pMan1) == Aig_ManCoNum(pMan1), "Aig_ManCoNum(pMan1) == Aig_ManCoNum(pMan1)");

    // make sure that the first miter has more nodes
    if Aig_ManNodeNum(pMan1) < Aig_ManNodeNum(pMan2) {
        pTemp = pMan1;
        pMan1 = pMan2;
        pMan2 = pTemp;
    }
    assert(Aig_ManNodeNum(pMan1) >= Aig_ManNodeNum(pMan2), "Aig_ManNodeNum(pMan1) >= Aig_ManNodeNum(pMan2)");

    if nPartSize != 0 {
        RetValue = Fra_FraigCecPartitioned( pMan1, pMan2, nConfLimit, nPartSize, fSmart, fVerbose );
    } else { // no partitioning
        RetValue = Fra_FraigCecPartitioned( pMan1, pMan2, nConfLimit, Aig_ManCoNum(pMan1), 0, fVerbose ); }

    // report the miter
    if RetValue == 1 {
        fmt.Printf("Networks are equivalent.  ");
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    
    } else if RetValue == 0 {
        fmt.Printf("Networks are NOT EQUIVALENT.  ");
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    
    } else {
        fmt.Printf("Networks are UNDECIDED.  ");
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    }
    fflush( os.Stdout );
    return RetValue;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

