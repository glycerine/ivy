package cnf

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [cnfUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfUtil.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes area, references, and nodes used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManScanMapping_rec(p *Cnf_Man_t, pObj *Aig_Obj_t, vMapped *Vec_Ptr_t) int32 {
    var pLeaf *Aig_Obj_t;
    var pCutBest *Dar_Cut_t;
    var aArea, i int32;
    if pObj.nRefs++ || Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj) {
        return 0; }
    assert(Aig_ObjIsAnd(pObj), "Aig_ObjIsAnd(pObj)");
    // collect the node first to derive pre-order
    if vMapped != nil {
        Vec_PtrPush( vMapped, pObj ); }
    // visit the transitive fanin of the selected cut
    if pObj.fMarkB {
        var vSuper *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
        Aig_ObjCollectSuper( pObj, vSuper );
        aArea = Vec_PtrSize(vSuper) + 1;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pLeaf) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
            aArea += Aig_ManScanMapping_rec( p, Aig_Regular(pLeaf), vMapped );
        }
        Vec_PtrFree( vSuper );
        ////////////////////////////
        pObj.fMarkB = 1;
    
    } else {
        pCutBest = Dar_ObjBestCut( pObj );
        aArea = Cnf_CutSopCost( p, pCutBest );
        // inlining of macro Dar_CutForEachLeaf
        for  i = 0; (i < int32((pCutBest).nLeaves)) && (((pLeaf) = Aig_ManObj(p.pManAig, (pCutBest).pLeaves[i])), 1); i++  {
            aArea += Aig_ManScanMapping_rec( p, pLeaf, vMapped );
        }
    }
    return aArea;
}

/**Function*************************************************************

  Synopsis    [Computes area, references, and nodes used in the mapping.]

  Description [Collects the nodes in reverse topological order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManScanMapping(p *Cnf_Man_t, fCollect int32) *Vec_Ptr_t {
    var vMapped *Vec_Ptr_t  = nil;
    var pObj *Aig_Obj_t;
    var i int32;
    // clean all references
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.nRefs = 0; }
    }
    // allocate the array
    if fCollect != 0 {
        vMapped = Vec_PtrAlloc( 1000 ); }
    // collect nodes reachable from POs in the DFS order through the best cuts
    p.aArea = 0;
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
        p.aArea += Aig_ManScanMapping_rec( p, Aig_ObjFanin0(pObj), vMapped );
    }
//    printf( "Variables = %6d. Clauses = %8d.\n", vMapped? Vec_PtrSize(vMapped) + Aig_ManCiNum(p->pManAig) + 1 : 0, p->aArea + 2 );
    return vMapped;
}

/**Function*************************************************************

  Synopsis    [Computes area, references, and nodes used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManScanMapping_rec(p *Cnf_Man_t, pObj *Aig_Obj_t, vMapped *Vec_Ptr_t, fPreorder int32) int32 {
    var pLeaf *Aig_Obj_t;
    var pCutBest *Cnf_Cut_t;
    var aArea, i int32;
    if pObj.nRefs++ || Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj) {
        return 0; }
    assert(Aig_ObjIsAnd(pObj), "Aig_ObjIsAnd(pObj)");
    assert(pObj.pData != nil, "pObj.pData != nil");
    // add the node to the mapping
    if (vMapped != nil) && (fPreorder != 0) {
         Vec_PtrPush( vMapped, pObj ); }
    // visit the transitive fanin of the selected cut
    if pObj.fMarkB {
        var vSuper *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
        Aig_ObjCollectSuper( pObj, vSuper );
        aArea = Vec_PtrSize(vSuper) + 1;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pLeaf) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
            aArea += Cnf_ManScanMapping_rec( p, Aig_Regular(pLeaf), vMapped, fPreorder );
        }
        Vec_PtrFree( vSuper );
        ////////////////////////////
        pObj.fMarkB = 1;
    
    } else {
        pCutBest = (Cnf_Cut_t *)pObj.pData;
//        assert( pCutBest->nFanins > 0 );
        assert(pCutBest.Cost < 127, "pCutBest.Cost < 127");
        aArea = pCutBest.Cost;
        // inlining of macro Cnf_CutForEachLeaf
        for  i = 0; (i < int32((pCutBest).nFanins)) && ((pLeaf) = Aig_ManObj(p.pManAig, (pCutBest).pFanins[i])); i++  {
            aArea += Cnf_ManScanMapping_rec( p, pLeaf, vMapped, fPreorder );
        }
    }
    // add the node to the mapping
    if (vMapped != nil) && !(fPreorder != 0) {
         Vec_PtrPush( vMapped, pObj ); }
    return aArea;
}

/**Function*************************************************************

  Synopsis    [Computes area, references, and nodes used in the mapping.]

  Description [Collects the nodes in reverse topological order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManScanMapping(p *Cnf_Man_t, fCollect int32, fPreorder int32) *Vec_Ptr_t {
    var vMapped *Vec_Ptr_t  = nil;
    var pObj *Aig_Obj_t;
    var i int32;
    // clean all references
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.nRefs = 0; }
    }
    // allocate the array
    if fCollect != 0 {
        vMapped = Vec_PtrAlloc( 1000 ); }
    // collect nodes reachable from POs in the DFS order through the best cuts
    p.aArea = 0;
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
        p.aArea += Cnf_ManScanMapping_rec( p, Aig_ObjFanin0(pObj), vMapped, fPreorder );
    }
//    printf( "Variables = %6d. Clauses = %8d.\n", vMapped? Vec_PtrSize(vMapped) + Aig_ManCiNum(p->pManAig) + 1 : 0, p->aArea + 2 );
    return vMapped;
}

/**Function*************************************************************

  Synopsis    [Returns the array of CI IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataCollectCiSatNums(pCnf *Cnf_Dat_t, p *Aig_Man_t) *Vec_Int_t {
    var vCiIds *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vCiIds = Vec_IntAlloc( Aig_ManCiNum(p) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Vec_IntPush( vCiIds, pCnf.pVarNums[pObj.Id] );
    }
    return vCiIds;
}

/**Function*************************************************************

  Synopsis    [Returns the array of CI IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataCollectCoSatNums(pCnf *Cnf_Dat_t, p *Aig_Man_t) *Vec_Int_t {
    var vCoIds *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vCoIds = Vec_IntAlloc( Aig_ManCoNum(p) );
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Vec_IntPush( vCoIds, pCnf.pVarNums[pObj.Id] );
    }
    return vCoIds;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataDeriveLitPolarities(p *Cnf_Dat_t) *unsigned char {
    var i int32;
    var c int32;
    var iClaBeg int32;
    var iClaEnd int32;
    var pLit *int32;
    // inlining of macro ABC_CALLOC
    var pPols0 *uint32  = ((unsigned *) calloc((size_t)(Aig_ManObjNumMax(p.pMan)), sizeof(unsigned)));
    // inlining of macro ABC_CALLOC
    var pPols1 *uint32  = ((unsigned *) calloc((size_t)(Aig_ManObjNumMax(p.pMan)), sizeof(unsigned)));
    // inlining of macro ABC_CALLOC
    unsigned char * pPres = ((unsigned char *) calloc((size_t)(p.nClauses), sizeof(unsigned char)));
    for  i = 0; i < Aig_ManObjNumMax(p.pMan); i++  {
        if p.pObj2Count[i] == 0 {
            continue; }
        iClaBeg = p.pObj2Clause[i];
        iClaEnd = p.pObj2Clause[i] + p.pObj2Count[i];
        // go through the negative polarity clauses
        for  c = iClaBeg; c < iClaEnd; c++  {
            for  pLit = p.pClauses[c]+1; pLit < p.pClauses[c+1]; pLit++  {
                if Abc_LitIsCompl(p.pClauses[c][0]) {
                    pPols0[Abc_Lit2Var(*pLit)] |= uint32((2 - Abc_LitIsCompl(*pLit)));  // taking the opposite (!) -- not the case
                } else {
                    pPols1[Abc_Lit2Var(*pLit)] |= uint32((2 - Abc_LitIsCompl(*pLit))); }
            }
        }  // taking the opposite (!) -- not the case
        // record these clauses
        for  c = iClaBeg; c < iClaEnd; c++  {
            for  pLit = p.pClauses[c]+1; pLit < p.pClauses[c+1]; pLit++  {
                if Abc_LitIsCompl(p.pClauses[c][0]) {
                    pPres[c] = (unsigned char)( uint32(pPres[c]) | (pPols0[Abc_Lit2Var(*pLit)] << (2*(pLit-p.pClauses[c]-1))) );
                } else {
                    pPres[c] = (unsigned char)( uint32(pPres[c]) | (pPols1[Abc_Lit2Var(*pLit)] << (2*(pLit-p.pClauses[c]-1))) ); }
            }
        }
        // clean negative polarity
        for  c = iClaBeg; c < iClaEnd; c++  {
            for  pLit = p.pClauses[c]+1; pLit < p.pClauses[c+1]; pLit++  {
                pPols0[Abc_Lit2Var(*pLit)] = pPols1[Abc_Lit2Var(*pLit)] = 0;
            }
        }
    }
    // inlining of macro ABC_FREE
    ((pPols0) ? (free((char *) (pPols0)), (pPols0) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pPols1) ? (free((char *) (pPols1)), (pPols1) = 0) : 0);
/*
//    for ( c = 0; c < p->nClauses; c++ )
    for ( c = 0; c < 100; c++ )
    {
        printf( "Clause %6d : ", c );
        for ( i = 0; i < 4; i++ )
            printf( "%d ", ((unsigned)pPres[c] >> (2*i)) & 3 );
        printf( "  " );
        for ( pLit = p->pClauses[c]; pLit < p->pClauses[c+1]; pLit++ )
            printf( "%6d ", *pLit );
        printf( "\n" );
    }
*/
    return pPres;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataReadFromFile(pFileName string) *Cnf_Dat_t {
    var MaxLine int32  = 1000000;
    var Var int32;
    var Lit int32;
    var nVars int32 = -1;
    var nClas int32 = -1;
    var i int32;
    var Entry int32;
    var iLine int32;
    var pCnf *Cnf_Dat_t  = nil;
    var vClas *Vec_Int_t  = nil;
    var vLits *Vec_Int_t  = nil;
    var pBuffer, pToken string;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return nil;
    }
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(MaxLine)));
    for  fgets(pBuffer, MaxLine, pFile) != nil  {
        iLine++;
        if pBuffer[0] == 'c' {
            continue; }
        if pBuffer[0] == 'p' {
            pToken = strtok( pBuffer+1, " \t" );
            if int32(strings.Compare(pToken, "cnf")) != 0 {
                fmt.Printf("Incorrect input file.\n");
                goto finish;
            }
            pToken = strtok( nil, " \t" );
            nVars = strconv.Atoi(pToken);
            pToken = strtok( nil, " \t" );
            nClas = strconv.Atoi(pToken);
            if nVars <= 0 || nClas <= 0 {
                fmt.Printf("Incorrect parameters.\n");
                goto finish;
            }
            // temp storage
            vClas = Vec_IntAlloc( nClas+1 );
            vLits = Vec_IntAlloc( nClas*8 );
            continue;
        }
        pToken = strtok( pBuffer, " \t\r\n" );
        if pToken == nil {
            continue; }
        Vec_IntPush( vClas, Vec_IntSize(vLits) );
        for  pToken  {
            Var = strconv.Atoi(pToken);
            if Var == 0 {
                break; }
            if Var > 0 {
                Lit = Abc_Var2Lit(Var-1, 0);
            } else {
                Lit = Abc_Var2Lit(-Var-1, 1);
            }
            if Lit >= 2*nVars {
                fmt.Printf("Literal %d is out-of-bound for %d variables.\n", Lit, nVars);
                goto finish;
            }
            Vec_IntPush( vLits, Lit );
            pToken = strtok( nil, " \t\r\n" );
        }
        if Var != 0 {
            fmt.Printf("There is no zero-terminator in line %d.\n", iLine);
            goto finish;
        }
    }
    // finalize
    if Vec_IntSize(vClas) != nClas {
        fmt.Printf("Warning! The number of clauses (%d) is different from declaration (%d).\n", Vec_IntSize(vClas), nClas); }
    Vec_IntPush( vClas, Vec_IntSize(vLits) );
    // create
    // inlining of macro ABC_CALLOC
    pCnf = &Cnf_Dat_t{};
    pCnf.nVars     = nVars;
    pCnf.nClauses  = Vec_IntSize(vClas)-1;
    pCnf.nLiterals = Vec_IntSize(vLits);
    // inlining of macro ABC_ALLOC
    pCnf.pClauses  = ((int * *) malloc(sizeof(int *) * (size_t)(Vec_IntSize(vClas))));
    pCnf.pClauses[0] = Vec_IntReleaseArray(vLits);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClas)) && (((Entry) = Vec_IntEntry(vClas, i)), 1); i++  {
        pCnf.pClauses[i] = pCnf.pClauses[0] + Entry;
    }
finish:
    _ = (pFile).Close();
    Vec_IntFreeP( &vClas );
    Vec_IntFreeP( &vLits );
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
    return pCnf;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataSolveFromFile(pFileName string, nConfLimit int32, nLearnedStart int32, nLearnedDelta int32, nLearnedPerce int32, fVerbose int32, fShowPattern int32, ppModel **int32, nPis int32) int32 {
    var clk abctime  = Abc_Clock();
    var pCnf *Cnf_Dat_t  = Cnf_DataReadFromFile( pFileName );
    var pSat *sat_solver;
    var i int32;
    var status int32;
    var RetValue int32 = -1;
    if pCnf == nil {
        return -1; }
    if fVerbose != 0 {
        fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. ", pCnf.nVars, pCnf.nClauses, pCnf.nLiterals);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    // convert into SAT solver
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    if pSat == nil {
        fmt.Printf("The problem is trivially UNSAT.\n");
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

//sat_solver_start_cardinality( pSat, 100 );

    // solve the miter
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, 0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if status == l_Undef {
        RetValue = -1;
    } else if status == l_True {
        RetValue = 0;
    } else if status == l_False {
        RetValue = 1;
    } else {
        assert(false, "0"); }
    if fVerbose != 0 {
        Sat_SolverPrintStats( os.Stdout, pSat ); }
    if RetValue == -1 {
        Abc_Print( 1, "UNDECIDED      " );
    } else if RetValue == 0 {
        Abc_Print( 1, "SATISFIABLE    " );
    } else {
        Abc_Print( 1, "UNSATISFIABLE  " ); }
    //Abc_Print( -1, "\n" );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // derive SAT assignment
    if RetValue == 0 && nPis > 0 {
        // inlining of macro ABC_ALLOC
        *ppModel = ((int *) malloc(sizeof(int) * (size_t)(nPis)));
        for  i = 0; i < nPis; i++  {
            (*ppModel)[i] = sat_solver_var_value( pSat, pCnf.nVars - nPis + i );
        }
    }
    if RetValue == 0 && (fShowPattern != 0) {
        for  i = 0; i < pCnf.nVars; i++  {
            fmt.Printf("%d", sat_solver_var_value(pSat,i));
        }
        fmt.Printf("\n");
    }
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat );
    return RetValue;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

