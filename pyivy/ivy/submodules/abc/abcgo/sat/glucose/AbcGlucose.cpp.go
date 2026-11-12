package glucose

import (
	"fmt"
	"strconv"
	"time"
)

/**CFile****************************************************************

  FileName    [AbcGlucose.cpp]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver Glucose 3.0 by Gilles Audemard and Laurent Simon.]

  Synopsis    [Interface to Glucose.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 6, 2017.]

  Revision    [$Id: AbcGlucose.cpp,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sat/glucose/System.h"
//#include "sat/glucose/ParseUtils.h"
//#include "sat/glucose/Options.h"
//#include "sat/glucose/Dimacs.h"
//#include "sat/glucose/SimpSolver.h"

//#include "sat/glucose/AbcGlucose.h"

//#include "base/abc/abc.h"
//#include "aig/gia/gia.h"
//#include "sat/cnf/cnf.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

//using namespace Gluco;


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const USE_SIMP_SOLVER = 1

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

                      
    
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func glucose_solver_start() *SimpSolver {
    var S *SimpSolver  = new SimpSolver;
    S.setIncrementalMode();
    return S;
}

func glucose_solver_stop(S *Gluco.SimpSolver) {
    delete S;
}

func glucose_solver_reset(S *Gluco.SimpSolver) {
    S.reset();
}

func glucose_solver_addclause(S *Gluco.SimpSolver, plits *int32, nlits int32) int32 {
    vec<Lit> lits;
    for  var i int32  = 0; i < nlits; i, plits = i+1, plits+1 {
        // note: Glucose uses the same var->lit conventiaon as ABC
        for (*plits)/2 >= S.nVars() { S.newVar(); }
        assert((*plits)/2 < S.nVars(), "(*plits)/2 < S.nVars()"); // NOTE: since we explicitely use new function bmc_add_var
        Lit p;
        p.x = *plits;
        lits.push(p);
    }
    return S.addClause(lits); // returns 0 if the problem is UNSAT
}

func glucose_solver_setcallback(S *Gluco.SimpSolver, pman any, pfunc func(any, int32, *int32) int32) {
    S.pCnfMan = pman;
    S.pCnfFunc = pfunc;
    S.nCallConfl = 1000;
}

func glucose_solver_solve(S *Gluco.SimpSolver, plits *int32, nlits int32) int32 {
    vec<Lit> lits;
    for var i int32 =0;i<nlits;i, plits = i+1, plits+1 {
        Lit p;
        p.x = *plits;
        lits.push(p);
    }
    Gluco::lbool res = S.solveLimited(lits, 0);
    if res == l_True {
        return 1;
    } else {
        if res == l_False {
            return -1;
        } else {
            return 0;
        }
    }
}

func glucose_solver_addvar(S *Gluco.SimpSolver) int32 {
    S.newVar();
    return S.nVars() - 1;
}

func glucose_solver_read_cex(S *Gluco.SimpSolver) *int32 {
    return S.getCex();
}

func glucose_solver_read_cex_varvalue(S *Gluco.SimpSolver, ivar int32) int32 {
    return S.model[ivar] == l_True;
}

func glucose_solver_setstop(S *Gluco.SimpSolver, pstop *int32) {
    S.pstop = pstop;
}


/**Function*************************************************************

  Synopsis    [Wrapper APIs to calling from ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func bmcg_sat_solver_start() *bmcg_sat_solver {
    return (bmcg_sat_solver *)glucose_solver_start();
}
func bmcg_sat_solver_stop(s *bmcg_sat_solver) {
    glucose_solver_stop((Gluco::SimpSolver*)s);
}
func bmcg_sat_solver_reset(s *bmcg_sat_solver) {
    glucose_solver_reset((Gluco::SimpSolver*)s);
}


func bmcg_sat_solver_addclause(s *bmcg_sat_solver, plits *int32, nlits int32) int32 {
    return glucose_solver_addclause((Gluco::SimpSolver*)s,plits,nlits);
}

func bmcg_sat_solver_setcallback(s *bmcg_sat_solver, pman any, pfunc func(any, int32, *int32) int32) {
    glucose_solver_setcallback((Gluco::SimpSolver*)s,pman,pfunc);
}

func bmcg_sat_solver_solve(s *bmcg_sat_solver, plits *int32, nlits int32) int32 {
    return glucose_solver_solve((Gluco::SimpSolver*)s,plits,nlits);
}

func bmcg_sat_solver_final(s *bmcg_sat_solver, ppArray **int32) int32 {
    *ppArray = (int *)(Lit *)((Gluco::SimpSolver*)s).conflict;
    return ((Gluco::SimpSolver*)s).conflict.size();
}

func bmcg_sat_solver_addvar(s *bmcg_sat_solver) int32 {
    return glucose_solver_addvar((Gluco::SimpSolver*)s);
}

func bmcg_sat_solver_set_nvars(s *bmcg_sat_solver, nvars int32) {
    var i int32;
    for  i = bmcg_sat_solver_varnum(s); i < nvars; i++  {
        bmcg_sat_solver_addvar(s);
    }
}

func bmcg_sat_solver_eliminate(s *bmcg_sat_solver, turn_off_elim int32) int32 {
//    return 1; 
    return ((Gluco::SimpSolver*)s).eliminate(turn_off_elim != 0);
}

func bmcg_sat_solver_var_is_elim(s *bmcg_sat_solver, v int32) int32 {
//    return 0; 
    return ((Gluco::SimpSolver*)s).isEliminated(v);
}

func bmcg_sat_solver_var_set_frozen(s *bmcg_sat_solver, v int32, freeze int32) {
    ((Gluco::SimpSolver*)s).setFrozen(v, freeze != 0);
}

func bmcg_sat_solver_elim_varnum(s *bmcg_sat_solver) int32 {
//    return 0; 
    return ((Gluco::SimpSolver*)s).eliminated_vars;
}

func bmcg_sat_solver_read_cex(s *bmcg_sat_solver) *int32 {
    return glucose_solver_read_cex((Gluco::SimpSolver*)s);
}
func bmcg_sat_solver_read_cex_varvalue(s *bmcg_sat_solver, ivar int32) int32 {
    return glucose_solver_read_cex_varvalue((Gluco::SimpSolver*)s, ivar);
}

func bmcg_sat_solver_set_stop(s *bmcg_sat_solver, pstop *int32) {
    glucose_solver_setstop((Gluco::SimpSolver*)s, pstop);
}

func bmcg_sat_solver_set_runtime_limit(s *bmcg_sat_solver, Limit abctime) abctime {
    var nRuntimeLimit abctime  = ((Gluco::SimpSolver*)s).nRuntimeLimit;
    ((Gluco::SimpSolver*)s).nRuntimeLimit = Limit;
    return nRuntimeLimit;
}

func bmcg_sat_solver_set_conflict_budget(s *bmcg_sat_solver, Limit int32) {
    if Limit > 0 { 
        ((Gluco::SimpSolver*)s).setConfBudget( (int64_t)Limit );
    } else { 
        ((Gluco::SimpSolver*)s).budgetOff(); }
}

func bmcg_sat_solver_varnum(s *bmcg_sat_solver) int32 {
    return ((Gluco::SimpSolver*)s).nVars();
}
func bmcg_sat_solver_clausenum(s *bmcg_sat_solver) int32 {
    return ((Gluco::SimpSolver*)s).nClauses();
}
func bmcg_sat_solver_learntnum(s *bmcg_sat_solver) int32 {
    return ((Gluco::SimpSolver*)s).nLearnts();
}
func bmcg_sat_solver_conflictnum(s *bmcg_sat_solver) int32 {
    return ((Gluco::SimpSolver*)s).conflicts;
}

func bmcg_sat_solver_minimize_assumptions(s *bmcg_sat_solver, plits *int32, nlits int32, pivot int32) int32 {
    vec<int>*array = &((Gluco::SimpSolver*)s).user_vec;
    var i, nlitsL, nlitsR, nresL, nresR, status int32;
    assert(pivot >= 0, "pivot >= 0");
//    assert( nlits - pivot >= 2 );
    assert(nlits - pivot >= 1, "nlits - pivot >= 1");
    if nlits - pivot == 1 {
        // since the problem is UNSAT, we try to solve it without assuming the last literal
        // if the result is UNSAT, the last literal can be dropped; otherwise, it is needed
        status = bmcg_sat_solver_solve( s, plits, pivot );
        return status != GLUCOSE_UNSAT; // return 1 if the problem is not UNSAT
    }
    assert(nlits - pivot >= 2, "nlits - pivot >= 2");
    nlitsL = (nlits - pivot) / 2;
    nlitsR = (nlits - pivot) - nlitsL;
    assert(nlitsL + nlitsR == nlits - pivot, "nlitsL + nlitsR == nlits - pivot");
    // solve with these assumptions
    status = bmcg_sat_solver_solve( s, plits, pivot + nlitsL );
    if status == GLUCOSE_UNSAT { // these are enough
        return bmcg_sat_solver_minimize_assumptions( s, plits, pivot + nlitsL, pivot ); }
    // these are not enough
    // solve for the right lits
//    nResL = nLitsR == 1 ? 1 : sat_solver_minimize_assumptions( s, pLits + nLitsL, nLitsR, nConfLimit );
    if nlitsR == 1 {
        nresL = 1;
    } else {
        nresL = bmcg_sat_solver_minimize_assumptions( s, plits, nlits, pivot + nlitsL );
    }
    // swap literals
    array.clear();
    for  i = 0; i < nlitsL; i++  {
        array.push(plits[pivot + i]);
    }
    for  i = 0; i < nresL; i++  {
        plits[pivot + i] = plits[pivot + nlitsL + i];
    }
    for  i = 0; i < nlitsL; i++  {
        plits[pivot + nresL + i] = (*array)[i];
    }
    // solve with these assumptions
    status = bmcg_sat_solver_solve( s, plits, pivot + nresL );
    if status == GLUCOSE_UNSAT { // these are enough
        return nresL; }
    // solve for the left lits
//    nResR = nLitsL == 1 ? 1 : sat_solver_minimize_assumptions( s, pLits + nResL, nLitsL, nConfLimit );
    if nlitsL == 1 {
        nresR = 1;
    } else {
        nresR = bmcg_sat_solver_minimize_assumptions( s, plits, pivot + nresL + nlitsL, pivot + nresL );
    }
    return nresL + nresR;
}

func bmcg_sat_solver_add_and(s *bmcg_sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
    var Lits [3]int32;

    Lits[0] = Abc_Var2Lit( iVar, !fCompl );
    Lits[1] = Abc_Var2Lit( iVar0, fCompl0 );
    if !(bmcg_sat_solver_addclause( s, Lits, 2 ) != 0) {
        return 0; }

    Lits[0] = Abc_Var2Lit( iVar, !fCompl );
    Lits[1] = Abc_Var2Lit( iVar1, fCompl1 );
    if !(bmcg_sat_solver_addclause( s, Lits, 2 ) != 0) {
        return 0; }

    Lits[0] = Abc_Var2Lit( iVar,   fCompl );
    Lits[1] = Abc_Var2Lit( iVar0, !fCompl0 );
    Lits[2] = Abc_Var2Lit( iVar1, !fCompl1 );
    if !(bmcg_sat_solver_addclause( s, Lits, 3 ) != 0) {
        return 0; }

    return 1;
}

func bmcg_solver_add_xor(pSat *bmcg_sat_solver, iVarA int32, iVarB int32, iVarC int32, fCompl int32) int32 {
    var Lits [3]int32;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarA >= 0 && iVarB >= 0 && iVarC >= 0");

    Lits[0] = Abc_Var2Lit( iVarA, !fCompl );
    Lits[1] = Abc_Var2Lit( iVarB, 1 );
    Lits[2] = Abc_Var2Lit( iVarC, 1 );
    Cid = bmcg_sat_solver_addclause( pSat, Lits, 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = Abc_Var2Lit( iVarA, !fCompl );
    Lits[1] = Abc_Var2Lit( iVarB, 0 );
    Lits[2] = Abc_Var2Lit( iVarC, 0 );
    Cid = bmcg_sat_solver_addclause( pSat, Lits, 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = Abc_Var2Lit( iVarA, fCompl );
    Lits[1] = Abc_Var2Lit( iVarB, 1 );
    Lits[2] = Abc_Var2Lit( iVarC, 0 );
    Cid = bmcg_sat_solver_addclause( pSat, Lits, 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = Abc_Var2Lit( iVarA, fCompl );
    Lits[1] = Abc_Var2Lit( iVarB, 0 );
    Lits[2] = Abc_Var2Lit( iVarC, 1 );
    Cid = bmcg_sat_solver_addclause( pSat, Lits, 3 );
    assert((Cid != 0), "Cid");
    return 4;
}

func bmcg_sat_solver_jftr(s *bmcg_sat_solver) int32 {
    return ((Gluco::SimpSolver*)s).jftr;
}

func bmcg_sat_solver_set_jftr(s *bmcg_sat_solver, jftr int32) {
    ((Gluco::SimpSolver*)s).jftr = jftr;
}

func bmcg_sat_solver_set_var_fanin_lit(s *bmcg_sat_solver, var int32, lit0 int32, lit1 int32) {
    ((Gluco::SimpSolver*)s).sat_solver_set_var_fanin_lit(var, lit0, lit1);
}

func bmcg_sat_solver_start_new_round(s *bmcg_sat_solver) {
    ((Gluco::SimpSolver*)s).sat_solver_start_new_round();
}

func bmcg_sat_solver_mark_cone(s *bmcg_sat_solver, var int32) {
    ((Gluco::SimpSolver*)s).sat_solver_mark_cone(var);
}


     

                                                                        

                

                
               
                

                

                                                                        
                               
 
                            
                            
             
 

                                          
 
             
 

                                           
 
               
 

                                                                      
 
                  
                                            
     
                                                                  
                                                     
                                                                                                   
              
                     
                     
     
                                                                   
 

                                                                                             
 
                      
                        
                         
 

                                                                  
 
                  
                                     
     
              
                     
                     
     
                                             
                                                         
 

                                           
 
                
                          
 

                                                
 
                       
 

                                                                
 
                                    
 

                                                          
 
                     
 


                                                                        

                                                 

                
               
                

                

                                                                        
                                          
 
                                                     
 
                                             
 
                                           
 
                                              
 
                                            
 

                                                                         
 
                                                                   
 

                                                                                                
 
                                                             
 

                                                                     
 
                                                               
 

                                                             
 
                                                           
                                                
 

                                              
 
                                                    
 

                                                               
 
          
                                                         
                                  
 

                                                                      
 
              
                                                                    
 

                                                            
 
              
                                                      
 

                                                                            
 
                                                    
 

                                                   
 
             
                                                      
 

                                                  
 
                                                      
 

                                                                   
 
                                                                     
 

                                                              
 
                                                     
 

                                                                            
 
                                                               
                                               
                         
 

                                                                       
 
                     
                                                             
         
                                         
 

                                              
 
                                        
 
                                                 
 
                                           
 
                                                 
 
                                           
 
                                                   
 
                                          
 

                                                                                                  
 
                                                    
                                                
                         
                                   
                                 
                             
     
                                                                                           
                                                                                           
                                                          
                                                                               
     
                                 
                                 
                                      
                                               
                                   
                                                               
                                                      
                                                                                       
                           
                               
                                                                                                         
                                                                                                      
                    
                   
                                  
                                      
                                 
                                                     
                                  
                                               
                                   
                                                              
                                                      
                     
                              
                                                                                                        
                                                                                                                      
                         
 

                                                                                                                        
 
                

                                           
                                            
                                                   
                 

                                           
                                            
                                                   
                 

                                            
                                             
                                             
                                                   
                 

             
 

                                                                                              
 
                
            
                                                     

                                            
                                      
                                      
                                                     
                  

                                            
                                      
                                      
                                                     
                  

                                           
                                      
                                      
                                                     
                  

                                           
                                      
                                      
                                                     
                  
             
 

                                            
 
                                     
 

                                                           
 
                                     
 

                                                                                       
 
                                                                       
 

                                                        
 
                                                      
 

                                                           
 
                                                   
 

      


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func glucose_print_stats(s *SimpSolver, clk abctime) {
    var cpu_time float64  = float64(uint32(clk)) / int64(time.Second);
    var mem_used float64  = memUsed();
    fmt.Printf("c restarts              : %d (%d conflicts on average)\n", int32(s.starts), s.starts > 0 ? int32((s.conflicts/s.starts)) : 0);
    fmt.Printf("c blocked restarts      : %d (multiple: %d) \n", int32(s.nbstopsrestarts), int32(s.nbstopsrestartssame));
    fmt.Printf("c last block at restart : %d\n", int32(s.lastblockatrestart));
    fmt.Printf("c nb ReduceDB           : %-12d\n", int32(s.nbReduceDB));
    fmt.Printf("c nb removed Clauses    : %-12d\n", int32(s.nbRemovedClauses));
    fmt.Printf("c nb learnts DL2        : %-12d\n", int32(s.nbDL2));
    fmt.Printf("c nb learnts size 2     : %-12d\n", int32(s.nbBin));
    fmt.Printf("c nb learnts size 1     : %-12d\n", int32(s.nbUn));
    fmt.Printf("c conflicts             : %-12d  (%.0f /sec)\n", int32(s.conflicts), s.conflicts   /cpu_time);
    fmt.Printf("c decisions             : %-12d  (%4.2f %% random) (%.0f /sec)\n", int32(s.decisions), float32(s.rnd_decisions)*100 / float32(s.decisions), s.decisions   /cpu_time);
    fmt.Printf("c propagations          : %-12d  (%.0f /sec)\n", int32(s.propagations), s.propagations/cpu_time);
    fmt.Printf("c conflict literals     : %-12d  (%4.2f %% deleted)\n", int32(s.tot_literals), (s.max_literals - s.tot_literals)*100 / float64(s.max_literals));
    fmt.Printf("c nb reduced Clauses    : %-12d\n", int32(s.nbReducedClauses));
    if mem_used != 0 { fmt.Printf("Memory used           : %.2f MB\n", mem_used); }
    //printf("c CPU time              : %.2f sec\n", cpu_time);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Glucose_ReadDimacs(pFileName string, s *SimpSolver) {
    vec<Lit> * lits = &s.user_lits;
    var pBuffer string  = Extra_FileReadContents( pFileName );
    var pTemp string; var fComp, Var, VarMax int32;
    lits.clear();
    for  pTemp = pBuffer; *pTemp; pTemp++  {
        if *pTemp == 'c' || *pTemp == 'p' {
            for  *pTemp != '\n'  {
                pTemp++;
            }
            continue;
        }
        for  *pTemp == ' ' || *pTemp == '\t' || *pTemp == '\r' || *pTemp == '\n'  {
            pTemp++;
        }
        fComp = 0;
        if *pTemp == '-' {
            fComp = 1;
pTemp++; }
        if *pTemp == '+' {
            pTemp++; }
        Var = strconv.Atoi(pTemp);
        if Var == 0 {
            if lits.size() > 0 {
                s.addVar( VarMax );
                s.addClause(*lits);
                lits.clear();
            }
        
        } else {
            Var--;
            VarMax = Abc_MaxInt( VarMax, Var );
            lits.push( toLit(Abc_Var2Lit(Var, fComp)) );
        }
        for  *pTemp >= '0' && *pTemp <= '9'  {
            pTemp++;
        }
    }
    ABC_FREE( pBuffer );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Glucose_SolveCnf(pFileName string, pPars *Glucose_Pars) {
    var clk abctime  = Abc_Clock();

    var S SimpSolver;
    S.verbosity = pPars.verb;
    S.setConfBudget( pPars.nConfls > 0 ? (int64_t)pPars.nConfls : -1 );

//    gzFile in = gzopen(pFilename, "rb");
//    parse_DIMACS(in, S);
//    gzclose(in);
    Glucose_ReadDimacs( pFileName, S );

    if pPars.verb {
        fmt.Printf("c ============================[ Problem Statistics ]=============================\n");
        fmt.Printf("c |                                                                             |\n");
        fmt.Printf("c |  Number of variables:  %12d                                         |\n", S.nVars());
        fmt.Printf("c |  Number of clauses:    %12d                                         |\n", S.nClauses());
    }
    
    if pPars.pre {
        S.eliminate(true);
        fmt.Printf("c Simplication removed %d variables and %d clauses.  ", S.eliminated_vars, S.eliminated_clauses);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }

    vec<Lit> dummy;
    var ret lbool  = S.solveLimited(dummy, 0);
    if pPars.verb { glucose_print_stats(S, Abc_Clock() - clk); }
    fmt.Printf(ret == l_True ? "SATISFIABLE" : ret == l_False ? "UNSATISFIABLE" : "INDETERMINATE");
    Abc_PrintTime( 1, "      Time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Glucose_SolverFromAig(p *Gia_Man_t, s *SimpSolver) *Vec_Int_t {
    var clk abctime  = Abc_Clock();
    vec<Lit> * lits = &s.user_lits;
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8 /*nLutSize*/, 0 /*fCnfObjIds*/, 1/*fAddOrCla*/, 0, 0/*verbose*/ );
    for  var i int32  = 0; i < pCnf.nClauses; i++  {
        lits.clear();
        for  var pLit *int32  = pCnf.pClauses[i]; pLit < pCnf.pClauses[i+1]; pLit++  {
            lits.push( toLit(*pLit) ), s.addVar( *pLit >> 1 );
        }
        s.addClause(*lits);
    }
    var vCnfIds *Vec_Int_t  = Vec_IntAllocArrayCopy(pCnf.pVarNums,pCnf.nVars);
    fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. ", pCnf.nVars, pCnf.nClauses, pCnf.nLiterals);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Cnf_DataFree(pCnf);
    return vCnfIds;
}

// procedure below does not work because glucose_solver_addclause() expects Solver
func Glucose_SolverFromAig2(p *Gia_Man_t, S *SimpSolver) *Vec_Int_t {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8 /*nLutSize*/, 0 /*fCnfObjIds*/, 1/*fAddOrCla*/, 0, 0/*verbose*/ );
    for  var i int32  = 0; i < pCnf.nClauses; i++  {
        if !(glucose_solver_addclause( &S, pCnf.pClauses[i], pCnf.pClauses[i+1]-pCnf.pClauses[i] ) != 0) {
            assert(false, "0"); }
    }
    var vCnfIds *Vec_Int_t  = Vec_IntAllocArrayCopy(pCnf.pVarNums,pCnf.nVars);
    //printf( "CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. ", pCnf->nVars, pCnf->nClauses, pCnf->nLiterals );
    //Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Cnf_DataFree(pCnf);
    return vCnfIds;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Glucose_GenerateCubes(pSat []*bmcg_sat_solver, vCiSatVars *Vec_Int_t, vVar2Index *Vec_Int_t, CubeLimit int32) *Vec_Str_t {
    var fCreatePrime int32  = 1;
    var nCubes int32;
    var nSupp int32 = Vec_IntSize(vCiSatVars);
    var vSop *Vec_Str_t   = Vec_StrAlloc( 1000 );
    var vLits *Vec_Int_t  = Vec_IntAlloc( nSupp );
    var vCube *Vec_Str_t  = Vec_StrAlloc( nSupp + 4 );
    Vec_StrFill( vCube, nSupp, '-' );
    Vec_StrPrintF( vCube, " 1\n\0" );
    for  nCubes = 0; !(CubeLimit != 0) || nCubes < CubeLimit; nCubes++  {
        var pFinal *int32;
        var nFinal int32;
        var iVar int32;
        var i int32;
        var k int32;
        // generate onset minterm
        var status int32  = bmcg_sat_solver_solve( pSat[1], nil, 0 );
        if status == GLUCOSE_UNSAT {
            break; }
        assert(status == GLUCOSE_SAT, "status == GLUCOSE_SAT");
        Vec_IntClear( vLits );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCiSatVars)) && (((iVar) = Vec_IntEntry(vCiSatVars, i)), 1); i++  {
            Vec_IntPush( vLits, Abc_Var2Lit(iVar, !bmcg_sat_solver_read_cex_varvalue(pSat[1], iVar)) );
        }
        // expand against offset
        if fCreatePrime != 0 {
            nFinal = bmcg_sat_solver_minimize_assumptions( pSat[0], Vec_IntArray(vLits), Vec_IntSize(vLits), 0 );
            Vec_IntShrink( vLits, nFinal );
            pFinal = Vec_IntArray( vLits );
            for  i = 0; i < nFinal; i++  {
                pFinal[i] = Abc_LitNot(pFinal[i]);
            }
        
        } else {
            status = bmcg_sat_solver_solve( pSat[0], Vec_IntArray(vLits), Vec_IntSize(vLits) );
            assert(status == GLUCOSE_UNSAT, "status == GLUCOSE_UNSAT");
            nFinal = bmcg_sat_solver_final( pSat[0], &pFinal );
        }
        // print cube
        Vec_StrFill( vCube, nSupp, '-' );
        for  i = 0; i < nFinal; i++  {
            var Index int32  = Vec_IntEntry(vVar2Index, Abc_Lit2Var(pFinal[i]));
            if Index == -1 {
                continue; }
            pFinal[k++] = pFinal[i];
            assert(Index >= 0 && Index < nSupp, "Index >= 0 && Index < nSupp");
            Vec_StrWriteEntry( vCube, Index, byte(('0' + Abc_LitIsCompl(pFinal[i]))) );
        }
        nFinal = k;
        Vec_StrAppend( vSop, Vec_StrArray(vCube) );
        //printf( "%s\n", Vec_StrArray(vCube) );
        // add blocking clause
        if !(bmcg_sat_solver_addclause( pSat[1], pFinal, nFinal ) != 0) {
            break; }
    }
    Vec_IntFree( vLits );
    Vec_StrFree( vCube );
    Vec_StrPush( vSop, '\0' );
    return vSop;
}
func bmcg_sat_solver_sop(p *Gia_Man_t, CubeLimit int32) *Vec_Str_t {
    var pSat [2]*bmcg_sat_solver = [2]*bmcg_sat_solver{ bmcg_sat_solver_start(), bmcg_sat_solver_start() };

    // generate CNF for the on-set and off-set
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8 /*nLutSize*/, 0 /*fCnfObjIds*/, 0/*fAddOrCla*/, 0, 0/*verbose*/ );
    var i int32;
    var n int32;
    var nVars int32 = Gia_ManCiNum(p);
    var Lit int32;//, Count = 0;
    var iFirstVar int32  = pCnf.nVars - nVars;
    assert(Gia_ManCoNum(p) == 1, "Gia_ManCoNum(p) == 1");
    for  n = 0; n < 2; n++  {
        bmcg_sat_solver_set_nvars( pSat[n], pCnf.nVars );
        Lit = Abc_Var2Lit( 1, !n );  // output variable is 1
        for  i = 0; i < pCnf.nClauses; i++  {
            if !(bmcg_sat_solver_addclause( pSat[n], pCnf.pClauses[i], pCnf.pClauses[i+1]-pCnf.pClauses[i] ) != 0) {
                assert(false, "0"); }
        }
        if !(bmcg_sat_solver_addclause( pSat[n], &Lit, 1 ) != 0) {
            var vSop *Vec_Str_t  = Vec_StrAlloc( 10 );
            Vec_StrPrintF( vSop, " %d\n\0", !n );
            Cnf_DataFree( pCnf );
            return vSop;
        }
    }
    Cnf_DataFree( pCnf );

    // collect cube vars and map SAT vars into them
    var vVars *Vec_Int_t  = Vec_IntAlloc( 100 ); 
    var vVarMap *Vec_Int_t  = Vec_IntStartFull( iFirstVar + nVars ); 
    for  i = 0; i < nVars; i++  {
        Vec_IntPush( vVars, iFirstVar+i );
        Vec_IntWriteEntry( vVarMap, iFirstVar+i, i );
    }

    var vSop *Vec_Str_t  = Glucose_GenerateCubes( pSat, vVars, vVarMap, CubeLimit );
    Vec_IntFree( vVarMap );
    Vec_IntFree( vVars );

    bmcg_sat_solver_stop( pSat[0] );
    bmcg_sat_solver_stop( pSat[1] );
    return vSop;
}
func bmcg_sat_solver_print_sop(p *Gia_Man_t) {
    var vSop *Vec_Str_t  = bmcg_sat_solver_sop( p, 0 );
    fmt.Printf("%s", Vec_StrArray(vSop));
    Vec_StrFree( vSop );
}
func bmcg_sat_solver_print_sop_lit(p *Gia_Man_t, Lit int32) {
    var vCisUsed *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32;
    var ObjId int32;
    var iNode int32 = Abc_Lit2Var( Lit );
    Gia_ManCollectCis( p, &iNode, 1, vCisUsed );
    Vec_IntSort( vCisUsed, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCisUsed)) && (((ObjId) = Vec_IntEntry(vCisUsed, i)), 1); i++  {
        Vec_IntWriteEntry( vCisUsed, i, Gia_ManIdToCioId(p, ObjId) );
    }
    Vec_IntPrint( vCisUsed );
    var pNew *Gia_Man_t  = Gia_ManDupConeSupp( p, Lit, vCisUsed );
    Vec_IntFree( vCisUsed );
    bmcg_sat_solver_print_sop( pNew );
    Gia_ManStop( pNew );
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Computing d-literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Original macro:
// #define Gia_CubeForEachVar( pCube, Value, i )                                                      \
//     for ( i = 0; (pCube[i] != ' ') && (Value = pCube[i]); i++ )           

// Original macro:
// #define Gia_SopForEachCube( pSop, nFanins, pCube )                                                 \
//     for ( pCube = (pSop); *pCube; pCube += (nFanins) + 3 )


func bmcg_sat_generate_dvars(vCiVars *Vec_Int_t, vSop *Vec_Str_t, vDLits *Vec_Int_t) {
    var i, Lit, Counter, nCubes int32;
    var Value byte;
    var pCube string;
    var pSop string = Vec_StrArray( vSop );
    var vCounts *Vec_Int_t  = Vec_IntStart( 2*Vec_IntSize(vCiVars) );
    Vec_IntClear( vDLits );
    // inlining of macro Gia_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (Vec_IntSize(vCiVars)) + 3  {
        nCubes++;
        // inlining of macro Gia_CubeForEachVar
        for  i = 0; (pCube[i] != ' ') && (Value = pCube[i]); i++  {
            if Value == '1' {
                Vec_IntAddToEntry( vCounts, 2*i, 1 );
            } else if Value == '0' {
                Vec_IntAddToEntry( vCounts, 2*i+1, 1 ); }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  Lit = 0; (Lit < Vec_IntSize(vCounts)) && (((Counter) = Vec_IntEntry(vCounts, Lit)), 1); Lit++  {
        if Counter == nCubes {
            Vec_IntPush( vDLits, Abc_Var2Lit(Vec_IntEntry(vCiVars, Abc_Lit2Var(Lit)), Abc_LitIsCompl(Lit)) ); }
    }
    Vec_IntSort( vDLits, 0 );
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    [Performs SAT-based quantification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func bmcg_sat_solver_quantify2(p *Gia_Man_t, iLit int32, fHash int32, pFuncCiToKeep func(any, int32) int32, pData any, vDLits *Vec_Int_t) int32 {
    var fSynthesize int32;
    extern Gia_Man_t * Abc_SopSynthesizeOne( char * pSop, int fClp );
    var pMan, pNew, pTemp *Gia_Man_t;  var vSop *Vec_Str_t;
    var i int32;
    var CiId int32;
    var ObjId int32;
    var Res int32;
    var nCubes int32;
    var nNodes int32;
    var Count int32;
    var iNode int32 = Abc_Lit2Var(iLit);
    var vCisUsed *Vec_Int_t  = Vec_IntAlloc( 100 );
    Gia_ManCollectCis( p, &iNode, 1, vCisUsed );
    Vec_IntSort( vCisUsed, 0 );
    if vDLits != nil { Vec_IntClear( vDLits ); }
    if iLit < 2 {
        return iLit; }
    // remap into CI Ids
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCisUsed)) && (((ObjId) = Vec_IntEntry(vCisUsed, i)), 1); i++  {
        Vec_IntWriteEntry( vCisUsed, i, Gia_ManIdToCioId(p, ObjId) );
    }
    // duplicate cone
    pNew = Gia_ManDupConeSupp( p, iLit, vCisUsed );
    assert(Gia_ManCiNum(pNew) == Vec_IntSize(vCisUsed), "Gia_ManCiNum(pNew) == Vec_IntSize(vCisUsed)");
    nNodes = Gia_ManAndNum(pNew);

    // perform quantification one CI at a time
    assert(pFuncCiToKeep, "pFuncCiToKeep");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCisUsed)) && (((CiId) = Vec_IntEntry(vCisUsed, i)), 1); i++  {
        if !pFuncCiToKeep( pData, CiId ) {
            //printf( "Quantifying %d.\n", CiId );
            pNew = Gia_ManDupExist( pTemp = pNew, i );
            Gia_ManStop( pTemp );
            Count++;
        }
    }
    if Gia_ManPoIsConst(pNew, 0) {
        var RetValue int32  = Gia_ManPoIsConst1(pNew, 0);
        Vec_IntFree( vCisUsed );
        Gia_ManStop( pNew );
        return RetValue;
    }

    if fSynthesize != 0 {
        vSop = bmcg_sat_solver_sop( pNew, 0 );
        Gia_ManStop( pNew );
        pMan = Abc_SopSynthesizeOne( Vec_StrArray(vSop), 1 );
        nCubes = Vec_StrCountEntry(vSop, '\n');
        if vDLits != nil {
            // convert into object IDs
            var vCisObjs *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vCisUsed) );
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vCisUsed)) && (((CiId) = Vec_IntEntry(vCisUsed, i)), 1); i++  {
                Vec_IntPush( vCisObjs, CiId + 1 );
            }
            bmcg_sat_generate_dvars( vCisObjs, vSop, vDLits );
            Vec_IntFree( vCisObjs );
        }
        Vec_StrFree( vSop );

        if Gia_ManPoIsConst(pMan, 0) {
            var RetValue int32  = Gia_ManPoIsConst1(pMan, 0);
            Vec_IntFree( vCisUsed );
            Gia_ManStop( pMan );
            return RetValue;
        }
    
    } else {
        pMan = pNew;
    }

    Res = Gia_ManDupConeBack( p, pMan, vCisUsed );

    // report the result
    //printf( "Performed quantification with %5d nodes, %3d keep-vars, %3d quant-vars, resulting in %5d cubes and %5d nodes. ", 
    //    nNodes, Vec_IntSize(vCisUsed) - Count, Count, nCubes, Gia_ManAndNum(pMan) );
    //Abc_PrintTime( 1, "Time", Abc_Clock() - clkAll );

    Vec_IntFree( vCisUsed );
    Gia_ManStop( pMan );
    return Res;
}


/**Function*************************************************************

  Synopsis    [Performs SAT-based quantification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatAndCollect_rec(p *Gia_Man_t, iObj int32, vObjsUsed *Vec_Int_t, vCiVars *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t; var iVar int32;
    if (iVar = Gia_ObjCopyArray(p, iObj)) >= 0 {
        return iVar; }
    pObj = Gia_ManObj( p, iObj );
    assert(Gia_ObjIsCand(pObj), "Gia_ObjIsCand(pObj)");
    if Gia_ObjIsAnd(pObj) {
        Gia_ManSatAndCollect_rec( p, Gia_ObjFaninId0(pObj, iObj), vObjsUsed, vCiVars );
        Gia_ManSatAndCollect_rec( p, Gia_ObjFaninId1(pObj, iObj), vObjsUsed, vCiVars );
    }
    iVar = Vec_IntSize( vObjsUsed );
    Vec_IntPush( vObjsUsed, iObj );
    Gia_ObjSetCopyArray( p, iObj, iVar );
    if (vCiVars != nil) && Gia_ObjIsCi(pObj) {
        Vec_IntPush( vCiVars, iVar ); }
    return iVar;
}                             
func Gia_ManQuantLoadCnf(p *Gia_Man_t, vObjsUsed *Vec_Int_t, pSats []*bmcg_sat_solver) {
    var pObj *Gia_Obj_t; var i int32;
    bmcg_sat_solver_reset( pSats[0] );
    if pSats[1] {
    bmcg_sat_solver_reset( pSats[1] ); }
    bmcg_sat_solver_set_nvars( pSats[0], Vec_IntSize(vObjsUsed) );
    if pSats[1] {
    bmcg_sat_solver_set_nvars( pSats[1], Vec_IntSize(vObjsUsed) ); }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjsUsed)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjsUsed,i))); i++  { 
        if Gia_ObjIsAnd(pObj) {
            var iObj int32   = Gia_ObjId( p, pObj );
            var iVar int32   = Gia_ObjCopyArray(p, iObj);
            var iVar0 int32  = Gia_ObjCopyArray(p, Gia_ObjFaninId0(pObj, iObj));
            var iVar1 int32  = Gia_ObjCopyArray(p, Gia_ObjFaninId1(pObj, iObj));
            bmcg_sat_solver_add_and( pSats[0], iVar, iVar0, iVar1, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 );
            if pSats[1] {
            bmcg_sat_solver_add_and( pSats[1], iVar, iVar0, iVar1, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); }
        
        } else if Gia_ObjIsConst0(pObj) {
            var Lit int32  = Abc_Var2Lit( Gia_ObjCopyArray(p, 0), 1 );
            var RetValue int32  = bmcg_sat_solver_addclause( pSats[0], &Lit, 1 );
            assert((RetValue != 0), "RetValue");
            if pSats[1] {
            bmcg_sat_solver_addclause( pSats[1], &Lit, 1 ); }
            assert(Lit == 1, "Lit == 1");
        }
    }
}
func Gia_ManFactorSop(p *Gia_Man_t, vCiObjIds *Vec_Int_t, vSop *Vec_Str_t, fHash int32) int32 {
    extern Gia_Man_t * Abc_SopSynthesizeOne( char * pSop, int fClp );
    var pMan *Gia_Man_t  = Abc_SopSynthesizeOne( Vec_StrArray(vSop), 1 );
    var pObj *Gia_Obj_t; var i, Result int32;
    assert(Gia_ManPiNum(pMan) == Vec_IntSize(vCiObjIds), "Gia_ManPiNum(pMan) == Vec_IntSize(vCiObjIds)");
    Gia_ManConst0(pMan).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(pMan)) && ((pObj) = Gia_ManCi(pMan, i)); i++  {
        pObj.Value = Abc_Var2Lit( Vec_IntEntry(vCiObjIds, i), 0 );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pMan.nObjs) && ((pObj) = Gia_ManObj(pMan, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if fHash != 0 {
            pObj.Value = Gia_ManHashAnd( p, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else {
            pObj.Value = Gia_ManAppendAnd( p, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    pObj = Gia_ManPo(pMan, 0);
    Result = Gia_ObjFanin0Copy(pObj);
    Gia_ManStop( pMan );
    return Result;
}
func bmcg_sat_solver_quantify(pSats []*bmcg_sat_solver, p *Gia_Man_t, iLit int32, fHash int32, pFuncCiToKeep func(any, int32) int32, pData any, vDLits *Vec_Int_t) int32 {
    var vObjsUsed *Vec_Int_t  = Vec_IntAlloc( 100 ); // GIA objs
    var vCiVars *Vec_Int_t  = Vec_IntAlloc( 100 );   // CI SAT vars
    var vVarMap *Vec_Int_t  = nil; var vSop *Vec_Str_t  = nil; 
    var i int32;
    var iVar int32;
    var iVarLast int32;
    var Lit int32;
    var RetValue int32;
    var Count int32;
    var Result int32 = -1;
    if vDLits != nil { Vec_IntClear( vDLits ); }
    if iLit < 2 {
        return iLit; }
    if Vec_IntSize(&p.vCopies) < Gia_ManObjNum(p) {
        Vec_IntFillExtra( &p.vCopies, Gia_ManObjNum(p), -1 ); }
    // assign variable number 0 to const0 node
    iVar = Vec_IntSize(vObjsUsed); 
    Vec_IntPush( vObjsUsed, 0 );
    Gia_ObjSetCopyArray( p, 0, iVar );
    assert(iVar == 0, "iVar == 0");    

    // collect other variables
    iVarLast = Gia_ManSatAndCollect_rec( p, Abc_Lit2Var(iLit), vObjsUsed, vCiVars );
    Gia_ManQuantLoadCnf( p, vObjsUsed, pSats );

    // check constants
    Lit = Abc_Var2Lit( iVarLast, !Abc_LitIsCompl(iLit) ); 
    RetValue = bmcg_sat_solver_addclause( pSats[0], &Lit, 1 ); // offset
    if !(RetValue != 0) || bmcg_sat_solver_solve(pSats[0], nil, 0) == GLUCOSE_UNSAT {
        Result = 1;
        goto cleanup;
    }
    Lit = Abc_Var2Lit( iVarLast, Abc_LitIsCompl(iLit) );
    RetValue = bmcg_sat_solver_addclause( pSats[1], &Lit, 1 ); // onset
    if !(RetValue != 0) || bmcg_sat_solver_solve(pSats[1], nil, 0) == GLUCOSE_UNSAT {
        Result = 0;
        goto cleanup;
    }
/*
    // reorder CI SAT variables to have keep-vars first
    Vec_Int_t * vCiVars2 = Vec_IntAlloc( 100 );   // CI SAT vars
    Vec_IntForEachEntry( vCiVars, iVar, i )
    {
        Gia_Obj_t * pObj = Gia_ManObj( p, Vec_IntEntry(vObjsUsed, iVar) );
        assert( Gia_ObjIsCi(pObj) );
        if ( pFuncCiToKeep(pData, Gia_ObjCioId(pObj)) )
            Vec_IntPush( vCiVars2, iVar );
    }
    Vec_IntForEachEntry( vCiVars, iVar, i )
    {
        Gia_Obj_t * pObj = Gia_ManObj( p, Vec_IntEntry(vObjsUsed, iVar) );
        assert( Gia_ObjIsCi(pObj) );
        if ( !pFuncCiToKeep(pData, Gia_ObjCioId(pObj)) )
            Vec_IntPush( vCiVars2, iVar );
    }
    ABC_SWAP( Vec_Int_t *, vCiVars2, vCiVars );
    Vec_IntFree( vCiVars2 );
*/
    // map CI SAT variables into their indexes used in the cubes
    vVarMap = Vec_IntStartFull( Vec_IntSize(vObjsUsed) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCiVars)) && (((iVar) = Vec_IntEntry(vCiVars, i)), 1); i++  {
        var pObj *Gia_Obj_t  = Gia_ManObj( p, Vec_IntEntry(vObjsUsed, iVar) );
        assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
        if pFuncCiToKeep(pData, Gia_ObjCioId(pObj)) {
            Vec_IntWriteEntry( vVarMap, iVar, i );
Count++; }
    }
    if Count == 0 || Count == Vec_IntSize(vCiVars) {
        if Count == 0 {
            Result = 1;
        } else {
            Result = iLit;
        }
        goto cleanup;
    }
    // generate cubes
    vSop = Glucose_GenerateCubes( pSats, vCiVars, vVarMap, 0 );
    //printf( "%s", Vec_StrArray(vSop) );
    // convert into object IDs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCiVars)) && (((iVar) = Vec_IntEntry(vCiVars, i)), 1); i++  {
        Vec_IntWriteEntry( vCiVars, i, Vec_IntEntry(vObjsUsed, iVar) );
    }
    // generate unate variables
    if vDLits != nil {
        bmcg_sat_generate_dvars( vCiVars, vSop, vDLits ); }
    // convert into an AIG
    RetValue = Gia_ManAndNum(p);
    Result = Gia_ManFactorSop( p, vCiVars, vSop, fHash );

    // report the result
//    printf( "Performed quantification with %5d nodes, %3d keep-vars, %3d quant-vars, resulting in %5d cubes and %5d nodes. ", 
//        Vec_IntSize(vObjsUsed), Count, Vec_IntSize(vCiVars) - Count, Vec_StrCountEntry(vSop, '\n'), Gia_ManAndNum(p)-RetValue );
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clkAll );

cleanup:
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjsUsed)) && (((iVar) = Vec_IntEntry(vObjsUsed, i)), 1); i++  {
        Gia_ObjSetCopyArray( p, iVar, -1 );
    }
    Vec_IntFree( vObjsUsed );
    Vec_IntFree( vCiVars );
    Vec_IntFreeP( &vVarMap );
    Vec_StrFreeP( &vSop );
    return Result;
}
func Gia_ManCiIsToKeep(pData any, i int32) int32 {
    return i % 5 != 0;
}
func Glucose_QuantifyAigTest(p *Gia_Man_t) {
    var pSats [3]*bmcg_sat_solver = [3]*bmcg_sat_solver{ bmcg_sat_solver_start(), bmcg_sat_solver_start(), bmcg_sat_solver_start() };

    var clk1 abctime  = Abc_Clock();
    var iRes1 int32  = bmcg_sat_solver_quantify( pSats, p, Gia_ObjFaninLit0p(p, Gia_ManPo(p, 0)), 0, Gia_ManCiIsToKeep, nil, nil );
    var clk1d abctime  = Abc_Clock()-clk1;

    var clk2 abctime  = Abc_Clock();
    var iRes2 int32  = bmcg_sat_solver_quantify2( p, Gia_ObjFaninLit0p(p, Gia_ManPo(p, 0)), 0, Gia_ManCiIsToKeep, nil, nil );
    var clk2d abctime  = Abc_Clock()-clk2;

    Abc_PrintTime( 1, "Time1", clk1d );
    Abc_PrintTime( 1, "Time2", clk2d );

    if bmcg_sat_solver_equiv_overlap_check( pSats[2], p, iRes1, iRes2, 1 ) != 0 {
        fmt.Printf("Verification passed.\n");
    } else {
        fmt.Printf("Verification FAILED.\n"); }

    Gia_ManAppendCo( p, iRes1 );
    Gia_ManAppendCo( p, iRes2 );

    bmcg_sat_solver_stop( pSats[0] );
    bmcg_sat_solver_stop( pSats[1] );
    bmcg_sat_solver_stop( pSats[2] );
}
func bmcg_sat_solver_quantify_test(pSats []*bmcg_sat_solver, p *Gia_Man_t, iLit int32, fHash int32, pFuncCiToKeep func(any, int32) int32, pData any, vDLits *Vec_Int_t) int32 {
    extern int Gia_ManQuantExist( Gia_Man_t * p, int iLit, int(*pFuncCiToKeep)(void *, int), void * pData );
    var Res1 int32  = Gia_ManQuantExist( p, iLit, pFuncCiToKeep, pData );
    var Res2 int32  = bmcg_sat_solver_quantify2( p, iLit, 1, pFuncCiToKeep, pData, nil );

    var pSat *bmcg_sat_solver  = bmcg_sat_solver_start();
    if bmcg_sat_solver_equiv_overlap_check( pSat, p, Res1, Res2, 1 ) != 0 {
        fmt.Printf("Verification passed.\n");
    } else {
        fmt.Printf("Verification FAILED.\n");
        bmcg_sat_solver_print_sop_lit( p, Res1 );
        bmcg_sat_solver_print_sop_lit( p, Res2 );
        fmt.Printf("\n");
    }
    return Res1;
}

/**Function*************************************************************

  Synopsis    [Checks equivalence or intersection of two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func bmcg_sat_solver_equiv_overlap_check(pSat *bmcg_sat_solver, p *Gia_Man_t, iLit0 int32, iLit1 int32, fEquiv int32) int32 {
    var pSats [2]*bmcg_sat_solver = [2]*bmcg_sat_solver{ pSat, nil };
    var vObjsUsed *Vec_Int_t  = Vec_IntAlloc( 100 ); 
    var i int32;
    var iVar int32;
    var iSatVar [2]int32;
    var iSatLit [2]int32;
    var Lits [2]int32;
    var status int32;
    if Vec_IntSize(&p.vCopies) < Gia_ManObjNum(p) {
        Vec_IntFillExtra( &p.vCopies, Gia_ManObjNum(p), -1 ); }

    // assign const0 variable number 0
    iVar = Vec_IntSize(vObjsUsed);
    Vec_IntPush( vObjsUsed, 0 );
    Gia_ObjSetCopyArray( p, 0, iVar );
    assert(iVar == 0, "iVar == 0");

    iSatVar[0] = Gia_ManSatAndCollect_rec( p, Abc_Lit2Var(iLit0), vObjsUsed, nil );
    iSatVar[1] = Gia_ManSatAndCollect_rec( p, Abc_Lit2Var(iLit1), vObjsUsed, nil );

    iSatLit[0] = Abc_Var2Lit( iSatVar[0], Abc_LitIsCompl(iLit0) );
    iSatLit[1] = Abc_Var2Lit( iSatVar[1], Abc_LitIsCompl(iLit1) );
    Gia_ManQuantLoadCnf( p, vObjsUsed, pSats );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjsUsed)) && (((iVar) = Vec_IntEntry(vObjsUsed, i)), 1); i++  {
        Gia_ObjSetCopyArray( p, iVar, -1 );
    }
    Vec_IntFree( vObjsUsed );

    if fEquiv != 0 {
        Lits[0] = iSatLit[0];
        Lits[1] = Abc_LitNot(iSatLit[1]);
        status  = bmcg_sat_solver_solve( pSats[0], Lits, 2 );
        if status == GLUCOSE_UNSAT {
            Lits[0] = Abc_LitNot(iSatLit[0]);
            Lits[1] = iSatLit[1];
            status  = bmcg_sat_solver_solve( pSats[0], Lits, 2 );
        }
        return status == GLUCOSE_UNSAT;
    
    } else {
        Lits[0] = iSatLit[0];
        Lits[1] = iSatLit[1];
        status  = bmcg_sat_solver_solve( pSats[0], Lits, 2 );
        return status == GLUCOSE_SAT;
    }
}
func Glucose_CheckTwoNodesTest(p *Gia_Man_t) {
    var n, Res int32;
    var pSat *bmcg_sat_solver  = bmcg_sat_solver_start();
    for  n = 0; n < 2; n++  {
        Res = bmcg_sat_solver_equiv_overlap_check( 
            pSat, p, 
            Gia_ObjFaninLit0p(p, Gia_ManPo(p, 0)), 
            Gia_ObjFaninLit0p(p, Gia_ManPo(p, 1)), 
            n );
        bmcg_sat_solver_reset( pSat );
        fmt.Printf("%s %s.\n", n ? "Equivalence" : "Overlap", Res ? "holds" : "fails");
    }
    bmcg_sat_solver_stop( pSat );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Glucose_SolveAig(p *Gia_Man_t, pPars *Glucose_Pars) int32 {  
    var clk abctime  = Abc_Clock();

    var S SimpSolver;
    S.verbosity = pPars.verb;
    S.verbEveryConflicts = 50000;
    S.showModel = false;
    //S.verbosity = 2;
    S.setConfBudget( pPars.nConfls > 0 ? (int64_t)pPars.nConfls : -1 );

    S.parsing = 1;
    var vCnfIds *Vec_Int_t  = Glucose_SolverFromAig(p,S);
    S.parsing = 0;

    if pPars.verb {
        fmt.Printf("c ============================[ Problem Statistics ]=============================\n");
        fmt.Printf("c |                                                                             |\n");
        fmt.Printf("c |  Number of variables:  %12d                                         |\n", S.nVars());
        fmt.Printf("c |  Number of clauses:    %12d                                         |\n", S.nClauses());
    }

    if pPars.pre {
        S.eliminate(true);
        fmt.Printf("c Simplication removed %d variables and %d clauses.  ", S.eliminated_vars, S.eliminated_clauses);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    
    vec<Lit> dummy;
    var ret lbool  = S.solveLimited(dummy, 0);

    if pPars.verb { glucose_print_stats(S, Abc_Clock() - clk); }
    fmt.Printf(ret == l_True ? "SATISFIABLE" : ret == l_False ? "UNSATISFIABLE" : "INDETERMINATE");
    Abc_PrintTime( 1, "      Time", Abc_Clock() - clk );

    // port counterexample
    if ret == l_True {
        var pObj *Gia_Obj_t;  var i int32;
        p.pCexComb = Abc_CexAlloc(0,Gia_ManCiNum(p),1);
        // inlining of macro Gia_ManForEachCi
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            assert(Vec_IntEntry(vCnfIds,Gia_ObjId(p, pObj))!=-1, "Vec_IntEntry(vCnfIds,Gia_ObjId(p, pObj))!=-1");
            if S.model[Vec_IntEntry(vCnfIds,Gia_ObjId(p, pObj))] == l_True {
                Abc_InfoSetBit( p.pCexComb.pData, i); }
        }
    }
    Vec_IntFree(vCnfIds);
    if ret == l_True {
        return 10;
    } else {
        if ret == l_False {
            return 20;
        } else {
            return 0;
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
