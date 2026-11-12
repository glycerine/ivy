package bmc

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [bmcBmc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Simple BMC package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcBmc.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "proof/fra/fra.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "sat/satoko/satoko.h"
//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create timeframes of the manager for BMC.]

  Description [The resulting manager is combinational. POs correspond to \
  the property outputs in each time-frame.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFramesBmc(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, f int32;
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    pFrames = Aig_ManStart( Aig_ManNodeNum(pAig) * nFrames );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pFrames );
    // create variables for register outputs
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_ManConst0( pFrames );
    }
    // add timeframes
    for  f = 0; f < nFrames; f++  {
        // create PI nodes for this frame
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCi( pFrames );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
        // create POs for this frame
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            Aig_ObjCreateCo( pFrames, Aig_ObjChild0Copy(pObj) );
        }
        if f == nFrames - 1 {
            break; }
        // save register inputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
        }
    }
    Aig_ManCleanup( pFrames );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes that are not counted yet.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFramesCount_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    if !Aig_ObjIsNode(pObj) {
        return 0; }
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    return 1 + Saig_ManFramesCount_rec( p, Aig_ObjFanin0(pObj) ) + 
        Saig_ManFramesCount_rec( p, Aig_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Create timeframes of the manager for BMC.]

  Description [The resulting manager is combinational. POs correspond to 
  the property outputs in each time-frame.
  The unrolling is stopped as soon as the number of nodes in the frames
  exceeds the given maximum size.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFramesBmcLimit(pAig *Aig_Man_t, nFrames int32, nSizeMax int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo, pObjPo *Aig_Obj_t;
    var i, f, Counter int32;
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    pFrames = Aig_ManStart( nSizeMax );
    Aig_ManIncrementTravId( pFrames );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pFrames );
    // create variables for register outputs
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_ManConst0( pFrames );
    }
    // add timeframes
    Counter = 0;
    for  f = 0; f < nFrames; f++  {
        // create PI nodes for this frame
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCi( pFrames );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
        // create POs for this frame
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObjPo = Aig_ObjCreateCo( pFrames, Aig_ObjChild0Copy(pObj) );
            Counter += Saig_ManFramesCount_rec( pFrames, Aig_ObjFanin0(pObjPo) );
        }
        if Counter >= nSizeMax {
            Aig_ManCleanup( pFrames );
            return pFrames;
        }
        if f == nFrames - 1 {
            break; }
        // save register inputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
        }
    }
    Aig_ManCleanup( pFrames );
    return pFrames;
}

//ABC_NAMESPACE_IMPL_END

//#include "misc/util/utilMem.h"

//ABC_NAMESPACE_IMPL_START

/**Function*************************************************************

  Synopsis    [Returns a counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat2_SolverGetModel(p *satoko_t, pVars *int32, nVars int32) *int32 {
    var pModel *int32;
    var i int32;
    // inlining of macro ABC_CALLOC
    pModel = ((int *) calloc((size_t)(nVars+1), sizeof(int)));
    for  i = 0; i < nVars; i++  {
        pModel[i] = satoko_read_cex_varvalue(p, pVars[i]);
    }
    return pModel;    
}
 
/**Function*************************************************************

  Synopsis    [Performs BMC for the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcSimple(pAig *Aig_Man_t, nFrames int32, nSizeMax int32, nConfLimit int32, fRewrite int32, fVerbose int32, piFrame *int32, nCofFanLit int32, fUseSatoko int32) int32 {
    extern Aig_Man_t * Gia_ManCofactorAig( Aig_Man_t * p, int nFrames, int nCofFanLit );
    var pSat *sat_solver  = nil;
    var pSat2 *satoko_t  = nil;
    var pCnf *Cnf_Dat_t;
    var pFrames, pAigTemp *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var status int32;
    var Lit int32;
    var i int32;
    var RetValue int32 = -1;
    var clk abctime;

    // derive the timeframes
    clk = Abc_Clock();
    if nCofFanLit != 0 {
        pFrames = Gia_ManCofactorAig( pAig, nFrames, nCofFanLit );
        if pFrames == nil {
            return -1; }
    
    } else if nSizeMax > 0 {
        pFrames = Saig_ManFramesBmcLimit( pAig, nFrames, nSizeMax );
        nFrames = Aig_ManCoNum(pFrames) / Saig_ManPoNum(pAig) + ((Aig_ManCoNum(pFrames) % Saig_ManPoNum(pAig)) > 0);
    
    } else {
        pFrames = Saig_ManFramesBmc( pAig, nFrames ); }
    if piFrame != nil {
        *piFrame = nFrames; }
    if fVerbose != 0 {
        fmt.Printf("Running \"bmc\". AIG:  PI/PO/Reg = %d/%d/%d.  Node = %6d. Lev = %5d.\n", Saig_ManPiNum(pAig), Saig_ManPoNum(pAig), Saig_ManRegNum(pAig), Aig_ManNodeNum(pAig), Aig_ManLevelNum(pAig));
        fmt.Printf("Time-frames (%d):  PI/PO = %d/%d.  Node = %6d. Lev = %5d.  ", nFrames, Aig_ManCiNum(pFrames), Aig_ManCoNum(pFrames), Aig_ManNodeNum(pFrames), Aig_ManLevelNum(pFrames));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        fflush( os.Stdout );
    }
    // rewrite the timeframes
    if fRewrite != 0 {
        clk = Abc_Clock();
//        pFrames = Dar_ManBalance( pAigTemp = pFrames, 0 );
        pFrames = Dar_ManRwsat( pAigTemp = pFrames, 1, 0 );
        Aig_ManStop( pAigTemp );
        if fVerbose != 0 {
            fmt.Printf("Time-frames after rewriting:  Node = %6d. Lev = %5d.  ", Aig_ManNodeNum(pFrames), Aig_ManLevelNum(pFrames));
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
            fflush( os.Stdout );
        }
    }
    // create the SAT solver
    clk = Abc_Clock();
    pCnf = Cnf_Derive( pFrames, Aig_ManCoNum(pFrames) );  
//if ( s_fInterrupt )
//return -1;
    if fUseSatoko != 0 {
        var opts satoko_opts_t;
        satoko_default_opts(&opts);
        opts.conf_limit = nConfLimit;
        pSat2 = satoko_create();  
        satoko_configure(pSat2, &opts);
        satoko_setnvars(pSat2, pCnf.nVars);
        for  i = 0; i < pCnf.nClauses; i++  {
            if !satoko_add_clause( pSat2, pCnf.pClauses[i], pCnf.pClauses[i+1]-pCnf.pClauses[i] ) {
                assert(false, "0"); }
        }
    
    } else {
        pSat = sat_solver_new();
        sat_solver_setnvars( pSat, pCnf.nVars );
        for  i = 0; i < pCnf.nClauses; i++  {
            if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
                assert(false, "0"); }
        }
    }
    if fVerbose != 0 {
        fmt.Printf("CNF: Variables = %6d. Clauses = %7d. Literals = %8d. ", pCnf.nVars, pCnf.nClauses, pCnf.nLiterals);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        fflush( os.Stdout );
    }
    if pSat != nil {
        status = sat_solver_simplify(pSat);
    } else {
        status = 1;
    }
    if status == 0 {
        if fVerbose != 0 {
            fmt.Printf("The BMC problem is trivially UNSAT\n");
            fflush( os.Stdout );
        }
    
    } else {
        var clkPart abctime  = Abc_Clock();
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pFrames.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pFrames.vCos, i)), 1); i++  {
            Lit = toLitCond( pCnf.pVarNums[pObj.Id], 0 );
            if fVerbose != 0 {
                fmt.Printf("Solving output %2d of frame %3d ... \r", i % Saig_ManPoNum(pAig), i / Saig_ManPoNum(pAig));
            }
            clk = Abc_Clock();
            if pSat2 != nil {
                status = satoko_solve_assumptions_limit( pSat2, &Lit, 1, nConfLimit );
            } else {
                status = sat_solver_solve( pSat, &Lit, &Lit + 1, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 ); }
            if (fVerbose != 0) && (i % Saig_ManPoNum(pAig) == Saig_ManPoNum(pAig) - 1) {
                fmt.Printf("Solved %2d outputs of frame %3d.  ", Saig_ManPoNum(pAig), i / Saig_ManPoNum(pAig));
                fmt.Printf("Conf =%8.0f. Imp =%11.0f. ", float64((pSat ? pSat.stats.conflicts    : satoko_conflictnum(pSat2))), float64((pSat ? pSat.stats.propagations : satoko_stats(pSat2).n_propagations)));
                // inlining of macro ABC_PRT
                Abc_Print(1, "%s =", ("T"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkPart)))/(float64(int64(time.Second))));;
                clkPart = Abc_Clock();
                fflush( os.Stdout );
            }
            if status == l_False {
/*
                Lit = lit_neg( Lit );
                RetValue = sat_solver_addclause( pSat, &Lit, &Lit + 1 );
                assert( RetValue );
                if ( pSat->qtail != pSat->qhead )
                {
                    RetValue = sat_solver_simplify(pSat);
                    assert( RetValue );
                }
*/
            
            } else if status == l_True {
                var vCiIds *Vec_Int_t  = Cnf_DataCollectPiSatNums( pCnf, pFrames );
                var pModel *int32;
                if pSat2 != nil {
                    pModel = Sat2_SolverGetModel(pSat2, vCiIds.pArray, vCiIds.nSize);
                } else {
                    pModel = Sat_SolverGetModel(pSat, vCiIds.pArray, vCiIds.nSize);
                }
                pModel[Aig_ManCiNum(pFrames)] = pObj.Id;
                pAig.pSeqModel = Fra_SmlCopyCounterExample( pAig, pFrames, pModel );
                // inlining of macro ABC_FREE
                ((pModel) ? (free((char *) (pModel)), (pModel) = 0) : 0);
                Vec_IntFree( vCiIds );

                if piFrame != nil {
                    *piFrame = i / Saig_ManPoNum(pAig); }
                RetValue = 0;
                break;
            
            } else {
                if piFrame != nil {
                    *piFrame = i / Saig_ManPoNum(pAig); }
                RetValue = -1;
                break;
            }
        }
    }
    if pSat != nil {  sat_solver_delete( pSat ); }
    if pSat2 != nil { satoko_destroy( pSat2 ); }
    Cnf_DataFree( pCnf );
    Aig_ManStop( pFrames );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

