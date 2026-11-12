package int

/**CFile****************************************************************

  FileName    [intCheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Procedures to perform incremental inductive check.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intCheck.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "intInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// checking manager
type Inter_Check_t_ struct {
    nFramesK           int32;     // the number of timeframes (K=1 for simple induction)
    nVars           int32;        // the current number of variables in the solver
    pFrames    *Aig_Man_t;      // unrolled timeframes
    pCnf    *Cnf_Dat_t;         // CNF of unrolled timeframes 
    pSat   *sat_solver;         // SAT solver 
    vOrLits    *Vec_Int_t;      // OR vars in each time frame (total number is the number nFrames)
    vAndLits    *Vec_Int_t;     // AND vars in the last timeframe (total number is the number of interpolants)
    vAssLits    *Vec_Int_t;     // assumptions (the union of the two)
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create timeframes of the manager for interpolation.]

  Description [The resulting manager is combinational. The primary inputs
  corresponding to register outputs are ordered first.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManUnrollFrames(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
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
        pObj.pData = Aig_ObjCreateCi( pFrames );
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
        // save register inputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
            Aig_ObjCreateCo( pFrames, (Aig_Obj_t *)pObjLo.pData );
        }
    }
    Aig_ManCleanup( pFrames );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [This procedure sets default values of interpolation parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_CheckStart(pTrans *Aig_Man_t, nFramesK int32) *Inter_Check_t {
    
    // create solver
    // inlining of macro ABC_CALLOC
    p := &Inter_Check_t{};
    p.vOrLits  = Vec_IntAlloc( 100 );
    p.vAndLits = Vec_IntAlloc( 100 );
    p.vAssLits = Vec_IntAlloc( 100 );
    // generate the timeframes 
    p.pFrames = Inter_ManUnrollFrames( pTrans, nFramesK );
    assert(Aig_ManCiNum(p.pFrames) == nFramesK * Saig_ManPiNum(pTrans) + Saig_ManRegNum(pTrans), "Aig_ManCiNum(p.pFrames) == nFramesK * Saig_ManPiNum(pTrans) + Saig_ManRegNum(pTrans)");
    assert(Aig_ManCoNum(p.pFrames) == nFramesK * Saig_ManRegNum(pTrans), "Aig_ManCoNum(p.pFrames) == nFramesK * Saig_ManRegNum(pTrans)");
    // convert to CNF
    p.pCnf = Cnf_Derive( p.pFrames, Aig_ManCoNum(p.pFrames) ); 
    p.pSat = (sat_solver *)Cnf_DataWriteIntoSolver( p.pCnf, 1, 0 );
    // assign parameters
    p.nFramesK = nFramesK;
    p.nVars    = p.pCnf.nVars;
    return p;
}

/**Function*************************************************************

  Synopsis    [This procedure sets default values of interpolation parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_CheckStop(p *Inter_Check_t) {
    if p == nil {
        return; }
    Vec_IntFree( p.vOrLits );
    Vec_IntFree( p.vAndLits );
    Vec_IntFree( p.vAssLits );
    Cnf_DataFree( p.pCnf );
    Aig_ManStop( p.pFrames );
    sat_solver_delete( p.pSat );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Creates one OR-gate: A + B = C.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_CheckAddOrGate(p *Inter_Check_t, iVarA int32, iVarB int32, iVarC int32) {
    var RetValue int32;
    var pLits [3]int32;
    // add A => C   or   !A + C
    pLits[0] = toLitCond(iVarA, 1);
    pLits[1] = toLitCond(iVarC, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
    assert((RetValue != 0), "RetValue");
    // add B => C   or   !B + C
    pLits[0] = toLitCond(iVarB, 1);
    pLits[1] = toLitCond(iVarC, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
    assert((RetValue != 0), "RetValue");
    // add !A & !B => !C  or A + B + !C
    pLits[0] = toLitCond(iVarA, 0);
    pLits[1] = toLitCond(iVarB, 0);
    pLits[2] = toLitCond(iVarC, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Creates equality: A = B.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_CheckAddEqual(p *Inter_Check_t, iVarA int32, iVarB int32) {
    var RetValue int32;
    var pLits [3]int32;
    // add A => B   or   !A + B
    pLits[0] = toLitCond(iVarA, 1);
    pLits[1] = toLitCond(iVarB, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
    assert((RetValue != 0), "RetValue");
    // add B => A   or   !B + A
    pLits[0] = toLitCond(iVarB, 1);
    pLits[1] = toLitCond(iVarA, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Perform the checking.]

  Description [Returns 1 if the check has passed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_CheckPerform(p *Inter_Check_t, pCnfInt *Cnf_Dat_t, nTimeNewOut abctime) int32 {
    var pObj, pObj2 *Aig_Obj_t;
    var i, f, VarA, VarB, RetValue, Entry, status int32;
    var nRegs int32  = Aig_ManCiNum(pCnfInt.pMan);
    assert(Aig_ManCoNum(p.pCnf.pMan) == p.nFramesK * nRegs, "Aig_ManCoNum(p.pCnf.pMan) == p.nFramesK * nRegs");
    assert(Aig_ManCoNum(pCnfInt.pMan) == 1, "Aig_ManCoNum(pCnfInt.pMan) == 1");

    // set runtime limit
    if nTimeNewOut {
        sat_solver_set_runtime_limit( p.pSat, nTimeNewOut ); }

    // add clauses to the SAT solver
    Cnf_DataLift( pCnfInt, p.nVars );
    for  f = 0; f <= p.nFramesK; f++  {
        // add clauses to the solver
        for  i = 0; i < pCnfInt.nClauses; i++  {
            RetValue = sat_solver_addclause( p.pSat, pCnfInt.pClauses[i], pCnfInt.pClauses[i+1] );
            assert((RetValue != 0), "RetValue");
        }
        // add equality clauses for the flop variables
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pCnfInt.pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pCnfInt.pMan.vCis, i)), 1); i++  {
            if f != 0 {
                pObj2 = Aig_ManCo(p.pFrames, i + (f-1) * nRegs);
            } else {
                pObj2 = Aig_ManCi(p.pFrames, i);
            }
            Inter_CheckAddEqual( p, pCnfInt.pVarNums[pObj.Id], p.pCnf.pVarNums[pObj2.Id] );
        }
        // add final clauses
        if f < p.nFramesK {
            if f == Vec_IntSize(p.vOrLits) { // find time here
            
                // add literal to this frame
                VarB = pCnfInt.pVarNums[ Aig_ManCo(pCnfInt.pMan, 0).Id ];
                Vec_IntPush( p.vOrLits, VarB );
            
            } else {
                // add OR gate for this frame
                VarA = Vec_IntEntry( p.vOrLits, f );
                VarB = pCnfInt.pVarNums[ Aig_ManCo(pCnfInt.pMan, 0).Id ];
                Inter_CheckAddOrGate( p, VarA, VarB, p.nVars + pCnfInt.nVars );
                Vec_IntWriteEntry( p.vOrLits, f, p.nVars + pCnfInt.nVars ); // using var ID!
            }
        
        } else {
            // add AND gate for this frame
            VarB = pCnfInt.pVarNums[ Aig_ManCo(pCnfInt.pMan, 0).Id ];
            Vec_IntPush( p.vAndLits, VarB );
        }
        // update variable IDs
        Cnf_DataLift( pCnfInt, pCnfInt.nVars + 1 );
        p.nVars += pCnfInt.nVars + 1;
    }
    Cnf_DataLift( pCnfInt, -p.nVars );
    assert(Vec_IntSize(p.vOrLits) == p.nFramesK, "Vec_IntSize(p.vOrLits) == p.nFramesK");

    // collect the assumption literals
    Vec_IntClear( p.vAssLits );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vOrLits)) && (((Entry) = Vec_IntEntry(p.vOrLits, i)), 1); i++  {
        Vec_IntPush( p.vAssLits, toLitCond(Entry, 0) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAndLits)) && (((Entry) = Vec_IntEntry(p.vAndLits, i)), 1); i++  {
        Vec_IntPush( p.vAssLits, toLitCond(Entry, 1) );
    }
/*
    if ( pCnfInt->nLiterals == 3635 )
    {
        int s = 0;
    }
*/
    // call the SAT solver
    status = sat_solver_solve( p.pSat, Vec_IntArray(p.vAssLits), 
        Vec_IntArray(p.vAssLits) + Vec_IntSize(p.vAssLits), 
        (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );

    return status == l_False;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

