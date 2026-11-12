package cnf

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [cnfMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfMan.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/bsat/satSolver2.h"
//#include "misc/zlib/zlib.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Cnf_Lit2Var(Lit int32) int32 { if Lit & 1 {
    return -(Lit >> 1)-1;
} else {
    return (Lit >> 1)+1;
}  }
func Cnf_Lit2Var2(Lit int32) int32 { if Lit & 1 {
    return -(Lit >> 1);
} else {
    return (Lit >> 1);
}    }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManStart() *Cnf_Man_t {
    var p *Cnf_Man_t;
    var i int32;
    // allocate the manager
    p = new(Cnf_Man_t);
    memset( p, 0, sizeof(Cnf_Man_t) );
    // derive internal data structures
    Cnf_ReadMsops( &p.pSopSizes, &p.pSops );
    // allocate memory manager for cuts
    p.pMemCuts = Aig_MmFlexStart();
    p.nMergeLimit = 10;
    // allocate temporary truth tables
    p.pTruths[0] = ABC_ALLOC( unsigned, 4 * Abc_TruthWordNum(p.nMergeLimit) );
    for  i = 1; i < 4; i++  {
        p.pTruths[i] = p.pTruths[i-1] + Abc_TruthWordNum(p.nMergeLimit);
    }
    p.vMemory = Vec_IntAlloc( 1 << 18 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManStop(p *Cnf_Man_t) {
    Vec_IntFree( p.vMemory );
    ABC_FREE( p.pTruths[0] );
    Aig_MmFlexStop( p.pMemCuts, 0 );
    ABC_FREE( p.pSopSizes );
    ABC_FREE( p.pSops[1] );
    ABC_FREE( p.pSops );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Returns the array of CI IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataCollectPiSatNums(pCnf *Cnf_Dat_t, p *Aig_Man_t) *Vec_Int_t {
    var vCiIds *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vCiIds = Vec_IntAlloc( Aig_ManCiNum(p) );
    Aig_ManForEachCi( p, pObj, i )
        Vec_IntPush( vCiIds, pCnf.pVarNums[pObj.Id] );
    return vCiIds;
}

/**Function*************************************************************

  Synopsis    [Allocates the new CNF.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataAlloc(pAig *Aig_Man_t, nVars int32, nClauses int32, nLiterals int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t;
    var i int32;
    pCnf = new(Cnf_Dat_t);
    memset( pCnf, 0, sizeof(Cnf_Dat_t) );
    pCnf.pMan = pAig;
    pCnf.nVars = nVars;
    pCnf.nClauses = nClauses;
    pCnf.nLiterals = nLiterals;
    pCnf.pClauses = ABC_ALLOC( int *, nClauses + 1 );
    pCnf.pClauses[0] = ABC_ALLOC( int, nLiterals );
    pCnf.pClauses[nClauses] = pCnf.pClauses[0] + nLiterals;
    pCnf.pVarNums = ABC_ALLOC( int, Aig_ManObjNumMax(pAig) );
//    memset( pCnf->pVarNums, 0xff, sizeof(int) * Aig_ManObjNumMax(pAig) );
    for  i = 0; i < Aig_ManObjNumMax(pAig); i++  {
        pCnf.pVarNums[i] = -1;
    }
    return pCnf;
}

/**Function*************************************************************

  Synopsis    [Allocates the new CNF.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataDup(p *Cnf_Dat_t) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t;
    var i int32;
    pCnf = Cnf_DataAlloc( p.pMan, p.nVars, p.nClauses, p.nLiterals );
    memcpy( pCnf.pClauses[0], p.pClauses[0], sizeof(int) * p.nLiterals );
    memcpy( pCnf.pVarNums, p.pVarNums, sizeof(int) * Aig_ManObjNumMax(p.pMan) );
    for  i = 1; i < p.nClauses; i++  {
        pCnf.pClauses[i] = pCnf.pClauses[0] + (p.pClauses[i] - p.pClauses[0]);
    }
    return pCnf;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataFree(p *Cnf_Dat_t) {
    if p == nil {
        return; }
    Vec_IntFreeP( &p.vMapping );
    ABC_FREE( p.pClaPols );
    ABC_FREE( p.pObj2Clause );
    ABC_FREE( p.pObj2Count );
    ABC_FREE( p.pClauses[0] );
    ABC_FREE( p.pClauses );
    ABC_FREE( p.pVarNums );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataLift(p *Cnf_Dat_t, nVarsPlus int32) {
    var pObj *Aig_Obj_t;
    var v int32;
    if p.pMan {
        Aig_ManForEachObj( p.pMan, pObj, v )
            if ( p.pVarNums[pObj.Id] >= 0 )
                p.pVarNums[pObj.Id] += nVarsPlus;
    }
    for  v = 0; v < p.nLiterals; v++  {
        p.pClauses[0][v] += 2*nVarsPlus;
    }
}
func Cnf_DataCollectFlipLits(p *Cnf_Dat_t, iFlipVar int32, vFlips *Vec_Int_t) {
    var v int32;
    assert(p.pMan == nil, "p.pMan == nil");
    Vec_IntClear( vFlips );
    for  v = 0; v < p.nLiterals; v++  {
        if Abc_Lit2Var(p.pClauses[0][v]) == iFlipVar {
            Vec_IntPush( vFlips, v ); }
    }
}
func Cnf_DataLiftAndFlipLits(p *Cnf_Dat_t, nVarsPlus int32, vLits *Vec_Int_t) {
    var i, iLit int32;
    assert(p.pMan == nil, "p.pMan == nil");
    Vec_IntForEachEntry( vLits, iLit, i )
        p.pClauses[0][iLit] = Abc_LitNot(p.pClauses[0][iLit]) + 2*nVarsPlus;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataPrint(p *Cnf_Dat_t, fReadable int32) {
    var pFile *os.File  = os.Stdout;
    var pLit *int32;
    var pStop *int32;
    var i int32;
    fmt.Fprintf(pFile, "p cnf %d %d\n", p.nVars, p.nClauses);
    for  i = 0; i < p.nClauses; i++  {
        for  pLit = p.pClauses[i], pStop = p.pClauses[i+1]; pLit < pStop; pLit++  {
            fmt.Fprintf(pFile, "%s%d ", Abc_LitIsCompl(*pLit) ? "-":"", fReadable? Abc_Lit2Var(*pLit) : Abc_Lit2Var(*pLit)+1);
        }
        fmt.Fprintf(pFile, "\n");
    }
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    [Writes CNF into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteIntoFileGz(p *Cnf_Dat_t, pFileName string, fReadable int32, vForAlls *Vec_Int_t, vExists *Vec_Int_t) {
    var pFile gzFile;
    var pLit *int32;
    var pStop *int32;
    var i int32;
    var VarId int32;
    pFile = gzopen( pFileName, "wb" );
    if pFile == nil {
        fmt.Printf("Cnf_WriteIntoFile(): Output file cannot be opened.\n");
        return;
    }
    gzprintf( pFile, "c Result of efficient AIG-to-CNF conversion using package CNF\n" );
    gzprintf( pFile, "p cnf %d %d\n", p.nVars, p.nClauses );
    if vForAlls != nil {
        gzprintf( pFile, "a " );
        Vec_IntForEachEntry( vForAlls, VarId, i )
            gzprintf( pFile, "%d ", fReadable? VarId : VarId+1 );
        gzprintf( pFile, "0\n" );
    }
    if vExists != nil {
        gzprintf( pFile, "e " );
        Vec_IntForEachEntry( vExists, VarId, i )
            gzprintf( pFile, "%d ", fReadable? VarId : VarId+1 );
        gzprintf( pFile, "0\n" );
    }
    for  i = 0; i < p.nClauses; i++  {
        for  pLit = p.pClauses[i], pStop = p.pClauses[i+1]; pLit < pStop; pLit++  {
            gzprintf( pFile, "%d ", fReadable? Cnf_Lit2Var2(*pLit) : Cnf_Lit2Var(*pLit) );
        }
        gzprintf( pFile, "0\n" );
    }
    gzprintf( pFile, "\n" );
    gzclose( pFile );
}

/**Function*************************************************************

  Synopsis    [Writes CNF into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteIntoFile(p *Cnf_Dat_t, pFileName string, fReadable int32, vForAlls *Vec_Int_t, vExists *Vec_Int_t) {
    var pFile *os.File;
    var pLit *int32;
    var pStop *int32;
    var i int32;
    var VarId int32;
    if !strncmp(pFileName+len(pFileName)-3,".gz",3) {
        Cnf_DataWriteIntoFileGz( p, pFileName, fReadable, vForAlls, vExists );
        return;
    }
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cnf_WriteIntoFile(): Output file cannot be opened.\n");
        return;
    }
    fmt.Fprintf(pFile, "c Result of efficient AIG-to-CNF conversion using package CNF\n");
    fmt.Fprintf(pFile, "p cnf %d %d\n", p.nVars, p.nClauses);
    if vForAlls != nil {
        fmt.Fprintf(pFile, "a ");
        Vec_IntForEachEntry( vForAlls, VarId, i )
            fmt.Fprintf(pFile, "%d ", fReadable? VarId : VarId+1);
        fmt.Fprintf(pFile, "0\n");
    }
    if vExists != nil {
        fmt.Fprintf(pFile, "e ");
        Vec_IntForEachEntry( vExists, VarId, i )
            fmt.Fprintf(pFile, "%d ", fReadable? VarId : VarId+1);
        fmt.Fprintf(pFile, "0\n");
    }
    for  i = 0; i < p.nClauses; i++  {
        for  pLit = p.pClauses[i], pStop = p.pClauses[i+1]; pLit < pStop; pLit++  {
            fmt.Fprintf(pFile, "%d ", fReadable? Cnf_Lit2Var2(*pLit) : Cnf_Lit2Var(*pLit));
        }
        fmt.Fprintf(pFile, "0\n");
    }
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Writes CNF into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteIntoSolverInt(pSolver any, p *Cnf_Dat_t, nFrames int32, fInit int32) any {
    var pSat *sat_solver  = (sat_solver *)pSolver;
    var i, f, status int32;
    assert(nFrames > 0, "nFrames > 0");
    assert(pSat, "pSat");
//    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, p.nVars * nFrames );
    for  i = 0; i < p.nClauses; i++  {
        if !sat_solver_addclause( pSat, p.pClauses[i], p.pClauses[i+1] ) {
            sat_solver_delete( pSat );
            return nil;
        }
    }
    if nFrames > 1 {
        var pObjLo, pObjLi *Aig_Obj_t;
        var nLitsAll int32;
        var pLits *int32;
        var Lits [2]int32;
        nLitsAll = 2 * p.nVars;
        pLits = p.pClauses[0];
        for  f = 1; f < nFrames; f++  {
            // add equality of register inputs/outputs for different timeframes
            Aig_ManForEachLiLoSeq( p.pMan, pObjLi, pObjLo, i )
            {
                Lits[0] = (f-1)*nLitsAll + toLitCond( p.pVarNums[pObjLi.Id], 0 );
                Lits[1] =  f   *nLitsAll + toLitCond( p.pVarNums[pObjLo.Id], 1 );
                if !sat_solver_addclause( pSat, Lits, Lits + 2 ) {
                    sat_solver_delete( pSat );
                    return nil;
                }
                Lits[0]++;
                Lits[1]--;
                if !sat_solver_addclause( pSat, Lits, Lits + 2 ) {
                    sat_solver_delete( pSat );
                    return nil;
                }
            }
            // add clauses for the next timeframe
            for  i = 0; i < p.nLiterals; i++  {
                pLits[i] += nLitsAll;
            }
            for  i = 0; i < p.nClauses; i++  {
                if !sat_solver_addclause( pSat, p.pClauses[i], p.pClauses[i+1] ) {
                    sat_solver_delete( pSat );
                    return nil;
                }
            }
        }
        // return literals to their original state
        nLitsAll = (f-1) * nLitsAll;
        for  i = 0; i < p.nLiterals; i++  {
            pLits[i] -= nLitsAll;
        }
    }
    if fInit != 0 {
        var pObjLo *Aig_Obj_t;
        var Lits [1]int32;
        Aig_ManForEachLoSeq( p.pMan, pObjLo, i )
        {
            Lits[0] = toLitCond( p.pVarNums[pObjLo.Id], 1 );
            if !sat_solver_addclause( pSat, Lits, Lits + 1 ) {
                sat_solver_delete( pSat );
                return nil;
            }
        }
    }
    status = sat_solver_simplify(pSat);
    if status == 0 {
        sat_solver_delete( pSat );
        return nil;
    }
    return pSat;
}

/**Function*************************************************************

  Synopsis    [Writes CNF into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteIntoSolver(p *Cnf_Dat_t, nFrames int32, fInit int32) any {
    return Cnf_DataWriteIntoSolverInt( sat_solver_new(), p, nFrames, fInit );
}

/**Function*************************************************************

  Synopsis    [Writes CNF into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteIntoSolver2(p *Cnf_Dat_t, nFrames int32, fInit int32) any {
    var pSat *sat_solver2;
    var i, f, status int32;
    assert(nFrames > 0, "nFrames > 0");
    pSat = sat_solver2_new();
    sat_solver2_setnvars( pSat, p.nVars * nFrames );
    for  i = 0; i < p.nClauses; i++  {
        if !sat_solver2_addclause( pSat, p.pClauses[i], p.pClauses[i+1], 0 ) {
            sat_solver2_delete( pSat );
            return nil;
        }
    }
    if nFrames > 1 {
        var pObjLo, pObjLi *Aig_Obj_t;
        var nLitsAll int32;
        var pLits *int32;
        var Lits [2]int32;
        nLitsAll = 2 * p.nVars;
        pLits = p.pClauses[0];
        for  f = 1; f < nFrames; f++  {
            // add equality of register inputs/outputs for different timeframes
            Aig_ManForEachLiLoSeq( p.pMan, pObjLi, pObjLo, i )
            {
                Lits[0] = (f-1)*nLitsAll + toLitCond( p.pVarNums[pObjLi.Id], 0 );
                Lits[1] =  f   *nLitsAll + toLitCond( p.pVarNums[pObjLo.Id], 1 );
                if !sat_solver2_addclause( pSat, Lits, Lits + 2, 0 ) {
                    sat_solver2_delete( pSat );
                    return nil;
                }
                Lits[0]++;
                Lits[1]--;
                if !sat_solver2_addclause( pSat, Lits, Lits + 2, 0 ) {
                    sat_solver2_delete( pSat );
                    return nil;
                }
            }
            // add clauses for the next timeframe
            for  i = 0; i < p.nLiterals; i++  {
                pLits[i] += nLitsAll;
            }
            for  i = 0; i < p.nClauses; i++  {
                if !sat_solver2_addclause( pSat, p.pClauses[i], p.pClauses[i+1], 0 ) {
                    sat_solver2_delete( pSat );
                    return nil;
                }
            }
        }
        // return literals to their original state
        nLitsAll = (f-1) * nLitsAll;
        for  i = 0; i < p.nLiterals; i++  {
            pLits[i] -= nLitsAll;
        }
    }
    if fInit != 0 {
        var pObjLo *Aig_Obj_t;
        var Lits [1]int32;
        Aig_ManForEachLoSeq( p.pMan, pObjLo, i )
        {
            Lits[0] = toLitCond( p.pVarNums[pObjLo.Id], 1 );
            if !sat_solver2_addclause( pSat, Lits, Lits + 1, 0 ) {
                sat_solver2_delete( pSat );
                return nil;
            }
        }
    }
    status = sat_solver2_simplify(pSat);
    if status == 0 {
        sat_solver2_delete( pSat );
        return nil;
    }
    return pSat;
}

/**Function*************************************************************

  Synopsis    [Adds the OR-clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteOrClause(p any, pCnf *Cnf_Dat_t) int32 {
    var pSat *sat_solver  = (sat_solver *)p;
    var pObj *Aig_Obj_t;
    var i int32;
    var pLits *int32;
    pLits = ABC_ALLOC( int, Aig_ManCoNum(pCnf.pMan) );
    Aig_ManForEachCo( pCnf.pMan, pObj, i )
        pLits[i] = toLitCond( pCnf.pVarNums[pObj.Id], 0 );
    if !sat_solver_addclause( pSat, pLits, pLits + Aig_ManCoNum(pCnf.pMan) ) {
        ABC_FREE( pLits );
        return 0;
    }
    ABC_FREE( pLits );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds the OR-clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteOrClause2(p any, pCnf *Cnf_Dat_t) int32 {
    var pSat *sat_solver2  = (sat_solver2 *)p;
    var pObj *Aig_Obj_t;
    var i int32;
    var pLits *int32;
    pLits = ABC_ALLOC( int, Aig_ManCoNum(pCnf.pMan) );
    Aig_ManForEachCo( pCnf.pMan, pObj, i )
        pLits[i] = toLitCond( pCnf.pVarNums[pObj.Id], 0 );
    if !sat_solver2_addclause( pSat, pLits, pLits + Aig_ManCoNum(pCnf.pMan), 0 ) {
        ABC_FREE( pLits );
        return 0;
    }
    ABC_FREE( pLits );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds the OR-clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataWriteAndClauses(p any, pCnf *Cnf_Dat_t) int32 {
    var pSat *sat_solver  = (sat_solver *)p;
    var pObj *Aig_Obj_t;
    var i, Lit int32;
    Aig_ManForEachCo( pCnf.pMan, pObj, i )
    {
        Lit = toLitCond( pCnf.pVarNums[pObj.Id], 0 );
        if !sat_solver_addclause( pSat, &Lit, &Lit+1 ) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Transforms polarity of the internal veriables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataTranformPolarity(pCnf *Cnf_Dat_t, fTransformPos int32) {
    var pObj *Aig_Obj_t;
    var pVarToPol *int32;
    var i, iVar int32;
    // create map from the variable number to its polarity
    pVarToPol = ABC_CALLOC( int, pCnf.nVars );
    Aig_ManForEachObj( pCnf.pMan, pObj, i )
    {
        if !(fTransformPos != 0) && Aig_ObjIsCo(pObj) {
            continue; }
        if pCnf.pVarNums[pObj.Id] >= 0 {
            pVarToPol[ pCnf.pVarNums[pObj.Id] ] = pObj.fPhase; }
    }
    // transform literals
    for  i = 0; i < pCnf.nLiterals; i++  {
        iVar = lit_var(pCnf.pClauses[0][i]);
        assert(iVar < pCnf.nVars, "iVar < pCnf.nVars");
        if pVarToPol[iVar] {
            pCnf.pClauses[0][i] = lit_neg( pCnf.pClauses[0][i] ); }
    }
    ABC_FREE( pVarToPol );
}

/**Function*************************************************************

  Synopsis    [Adds constraints for the two-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataAddXorClause(pSat any, iVarA int32, iVarB int32, iVarC int32) int32 {
    var Lits [3]lit;
    assert(iVarA > 0 && iVarB > 0 && iVarC > 0, "iVarA > 0 && iVarB > 0 && iVarC > 0");

    Lits[0] = toLitCond( iVarA, 1 );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 1 );
    if !sat_solver_addclause( (sat_solver *)pSat, Lits, Lits + 3 ) {
        return 0; }

    Lits[0] = toLitCond( iVarA, 1 );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 0 );
    if !sat_solver_addclause( (sat_solver *)pSat, Lits, Lits + 3 ) {
        return 0; }

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 0 );
    if !sat_solver_addclause( (sat_solver *)pSat, Lits, Lits + 3 ) {
        return 0; }

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 1 );
    if !sat_solver_addclause( (sat_solver *)pSat, Lits, Lits + 3 ) {
        return 0; }

    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

