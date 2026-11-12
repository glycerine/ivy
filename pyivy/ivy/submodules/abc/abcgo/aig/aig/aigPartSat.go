package aig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [aigPartSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Partitioning for SAT solving.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigPartSat.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/cnf/cnf.h"

//ABC_NAMESPACE_IMPL_START


/* 

The node partitioners defined in this file return array of intergers 
mapping each AND node's ID into the 0-based number of its partition.
The mapping of PIs/POs will be derived automatically in Aig_ManPartSplit().

The partitions can be ordered in any way, but the recommended ordering
is to first include partitions whose nodes are closer to the outputs.

*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Aig_Man_t * Dar_ManRwsat( Aig_Man_t * pAig, int fBalance, int fVerbose );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [No partitioning.]

  Description [The partitioner ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionMonolithic(p *Aig_Man_t) *Vec_Int_t {
    var vId2Part *Vec_Int_t;
    vId2Part = Vec_IntStart( Aig_ManObjNumMax(p) );
    return vId2Part;
}

/**Function*************************************************************

  Synopsis    [Partitioning using levelized order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionLevelized(p *Aig_Man_t, nPartSize int32) *Vec_Int_t {
    var vId2Part *Vec_Int_t;
    var vNodes *Vec_Vec_t;
    var pObj *Aig_Obj_t;
    var i, k, Counter int32;
    vNodes = Aig_ManLevelize( p );
    vId2Part = Vec_IntStart( Aig_ManObjNumMax(p) );
    // inlining of macro Vec_VecForEachEntryReverseReverse
    // inlining of macro Vec_PtrForEachEntryReverse
    for  i = Vec_VecSize(vNodes) - 1; i >= 0; i--  {                                                   for  k = Vec_PtrSize(Vec_VecEntry(vNodes, i)) - 1; (k >= 0) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vNodes, i), k)), 1); k--  { 
        Vec_IntWriteEntry( vId2Part, Aig_ObjId(pObj), Counter++/nPartSize );
    }
    }
    Vec_VecFree( vNodes );
    return vId2Part;
}

/**Function*************************************************************

  Synopsis    [Partitioning using DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionDfs(p *Aig_Man_t, nPartSize int32, fPreorder int32) *Vec_Int_t {
    var vId2Part *Vec_Int_t; 
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    vId2Part = Vec_IntStart( Aig_ManObjNumMax(p) );
    if fPreorder != 0 {
        vNodes = Aig_ManDfsPreorder( p, 1 );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            Vec_IntWriteEntry( vId2Part, Aig_ObjId(pObj), Counter++/nPartSize );
        }
    
    } else {
        vNodes = Aig_ManDfs( p, 1 );
        // inlining of macro Vec_PtrForEachEntryReverse
        for  i = Vec_PtrSize(vNodes) - 1; (i >= 0) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i--  {
            Vec_IntWriteEntry( vId2Part, Aig_ObjId(pObj), Counter++/nPartSize );
        }
    }
    Vec_PtrFree( vNodes );
    return vId2Part;
}

/**Function*************************************************************

  Synopsis    [Recursively constructs the partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartSplitOne_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t, vPio2Id *Vec_Int_t) {
    if !Aig_ObjIsTravIdCurrent( p, pObj ) { // new PI
        Aig_ObjSetTravIdCurrent( p, pObj );
/*
        if ( pObj->fMarkA ) // const0
            pObj->pData = Aig_ManConst0( pNew );
        else if ( pObj->fMarkB ) // const1
            pObj->pData = Aig_ManConst1( pNew );
        else
*/
        {
            pObj.pData = Aig_ObjCreateCi( pNew );
            if pObj.fMarkA { // const0
                ((Aig_Obj_t *)pObj.pData).fMarkA = 1;
            } else if pObj.fMarkB { // const1
                ((Aig_Obj_t *)pObj.pData).fMarkB = 1; }
            Vec_IntPush( vPio2Id, Aig_ObjId(pObj) );
        }
        return;
    }
    if pObj.pData { 
        return; }
    Aig_ManPartSplitOne_rec( pNew, p, Aig_ObjFanin0(pObj), vPio2Id );
    Aig_ManPartSplitOne_rec( pNew, p, Aig_ObjFanin1(pObj), vPio2Id );
    pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Carves out one partition of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartSplitOne(p *Aig_Man_t, vNodes *Vec_Ptr_t, pvPio2Id **Vec_Int_t) *Aig_Man_t {
    var vPio2Id *Vec_Int_t;
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // mark these nodes
    Aig_ManIncrementTravId( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Aig_ObjSetTravIdCurrent( p, pObj );
        pObj.pData = nil;
    }
    // add these nodes in a DFS order
    pNew = Aig_ManStart( Vec_PtrSize(vNodes) );
    vPio2Id = Vec_IntAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Aig_ManPartSplitOne_rec( pNew, p, pObj, vPio2Id );
    }
    // add the POs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Aig_ObjRefs((Aig_Obj_t *)pObj.pData) != Aig_ObjRefs(pObj) {
            assert(Aig_ObjRefs((Aig_Obj_t *)pObj.pData) < Aig_ObjRefs(pObj), "Aig_ObjRefs((Aig_Obj_t *)pObj.pData) < Aig_ObjRefs(pObj)");
            Aig_ObjCreateCo( pNew, (Aig_Obj_t *)pObj.pData );
            Vec_IntPush( vPio2Id, Aig_ObjId(pObj) );
        }
    }
    assert(Aig_ManNodeNum(pNew) == Vec_PtrSize(vNodes), "Aig_ManNodeNum(pNew) == Vec_PtrSize(vNodes)");
    *pvPio2Id = vPio2Id;
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives AIGs for each partition.]

  Description [The first argument is a original AIG. The second argument
  is the array mapping each AND-node's ID into the 0-based number of its
  partition. The last argument is the array of arrays (one for each new AIG) 
  mapping the index of each terminal in the new AIG (the index of each 
  terminal is derived by ordering PIs followed by POs in their natural order) 
  into the ID of the corresponding node in the original AIG. The returned 
  value is the array of AIGs representing the partitions.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartSplit(p *Aig_Man_t, vNode2Part *Vec_Int_t, pvPio2Id **Vec_Ptr_t, pvPart2Pos **Vec_Ptr_t) *Vec_Ptr_t {
    var vGroups, vPart2Pos *Vec_Vec_t;
    var vAigs, vPio2Id, vNodes *Vec_Ptr_t;
    var vPio2IdOne *Vec_Int_t;
    var pAig *Aig_Man_t;
    var pObj, pDriver *Aig_Obj_t;
    var i, nodePart, nParts int32;
    vAigs = Vec_PtrAlloc( 100 );
    vPio2Id = Vec_PtrAlloc( 100 );
    // put all nodes into levels according to their partition
    nParts = Vec_IntFindMax(vNode2Part) + 1;
    assert(nParts > 0, "nParts > 0");
    vGroups = Vec_VecAlloc( nParts );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNode2Part)) && (((nodePart) = Vec_IntEntry(vNode2Part, i)), 1); i++  {
        pObj = Aig_ManObj( p, i );
        if Aig_ObjIsNode(pObj) {
            Vec_VecPush( vGroups, nodePart, pObj ); }
    }

    // label PIs that should be restricted to some values
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        if Aig_ObjIsCi(pDriver) {
            if Aig_ObjFaninC0(pObj) {
                pDriver.fMarkA = 1; // const0 PI
            } else {
                pDriver.fMarkB = 1; } // const1 PI
        }
    }

    // create partitions
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(vGroups)) && (((vNodes) = Vec_VecEntry(vGroups, i)), 1); i++  {
        if Vec_PtrSize(vNodes) == 0 {
            fmt.Printf("Aig_ManPartSplit(): Skipping partition # %d without nodes (warning).\n", i);
            continue;
        }
        pAig = Aig_ManPartSplitOne( p, vNodes, &vPio2IdOne );
        Vec_PtrPush( vPio2Id, vPio2IdOne );
        Vec_PtrPush( vAigs, pAig );
    }
    Vec_VecFree( vGroups );

    // divide POs according to their partitions
    vPart2Pos = Vec_VecStart( Vec_PtrSize(vAigs) );
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pDriver = Aig_ObjFanin0(pObj);
        if Aig_ObjIsCi(pDriver) {
            pDriver.fMarkA = pDriver.fMarkB = 0;
        } else {
            Vec_VecPush( vPart2Pos, Vec_IntEntry(vNode2Part, Aig_ObjFaninId0(pObj)), pObj ); }
    }

    *pvPio2Id = vPio2Id;
    *pvPart2Pos = (Vec_Ptr_t *)vPart2Pos;
    return vAigs;
}

/**Function*************************************************************

  Synopsis    [Resets node polarity to unbias the polarity of CNF variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartResetNodePolarity(pPart *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPart.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fPhase = 0; }
    }
}

/**Function*************************************************************

  Synopsis    [Sets polarity according to the original nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartSetNodePolarity(p *Aig_Man_t, pPart *Aig_Man_t, vPio2Id *Vec_Int_t) {
    var pObj, pObjInit *Aig_Obj_t;
    var i int32;
    Aig_ManConst1(pPart).fPhase = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPart.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPart.vCis, i)), 1); i++  {
        pObjInit = Aig_ManObj( p, Vec_IntEntry(vPio2Id, i) );
        pObj.fPhase = pObjInit.fPhase;
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPart.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPart.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.fPhase = (Aig_ObjFanin0(pObj).fPhase ^ Aig_ObjFaninC0(pObj)) & (Aig_ObjFanin1(pObj).fPhase ^ Aig_ObjFaninC1(pObj)); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPart.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPart.vCos, i)), 1); i++  {
        pObjInit = Aig_ManObj( p, Vec_IntEntry(vPio2Id, Aig_ManCiNum(pPart) + i) );
        pObj.fPhase = (Aig_ObjFanin0(pObj).fPhase ^ Aig_ObjFaninC0(pObj));
        assert(pObj.fPhase == pObjInit.fPhase, "pObj.fPhase == pObjInit.fPhase");
    }
}

/**Function*************************************************************

  Synopsis    [Sets polarity according to the original nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDeriveCounterExample(p *Aig_Man_t, vNode2Var *Vec_Int_t, pSat *sat_solver) { 
    var vPisIds *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // collect IDs of PI variables
    // (fanoutless PIs have SAT var 0, which is an unused in the SAT solver -> has value 0)
    vPisIds = Vec_IntAlloc( Aig_ManCiNum(p) );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Vec_IntPush( vPisIds, Vec_IntEntry(vNode2Var, Aig_ObjId(pObj)) );
    }
    // derive the SAT assignment
    p.pData = Sat_SolverGetModel( pSat, vPisIds.pArray, vPisIds.nSize );
    Vec_IntFree( vPisIds );
}

/**Function*************************************************************

  Synopsis    [Derives CNF for the partition (pAig) and adds it to solver.]

  Description [Array vPio2Id contains mapping of the PI/PO terminal of pAig
  into the IDs of the corresponding original nodes. Array vNode2Var contains
  mapping of the original nodes into their SAT variable numbers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManAddNewCnfToSolver(pSat *sat_solver, pAig *Aig_Man_t, vNode2Var *Vec_Int_t, vPioIds *Vec_Int_t, vPartPos *Vec_Ptr_t, fAlignPol int32) int32 {
    var pCnf *Cnf_Dat_t;
    var pObj *Aig_Obj_t;
    var pBeg, pEnd *int32;
    var i int32;
    var Lits [2]int32;
    var iSatVarOld int32;
    var iNodeIdOld int32;
    // derive CNF and express it using new SAT variables
    pCnf = Cnf_Derive( pAig, Aig_ManCoNum(pAig) );
    Cnf_DataTranformPolarity( pCnf, 1 );
    Cnf_DataLift( pCnf, sat_solver_nvars(pSat) );
    // create new variables in the SAT solver
    sat_solver_setnvars( pSat, sat_solver_nvars(pSat) + pCnf.nVars );
    // add clauses for this CNF
    // inlining of macro Cnf_CnfForClause
    for  i = 0; i < pCnf.nClauses && (pBeg = pCnf.pClauses[i]) && (pEnd = pCnf.pClauses[i+1]); i++  {
        if !sat_solver_addclause( pSat, pBeg, pEnd ) {
            assert(false, "0"); // if it happens, can return 1 (unsatisfiable)
            return 1;
        }
    }
    // derive the connector clauses
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        iNodeIdOld = Vec_IntEntry( vPioIds, i );
        iSatVarOld = Vec_IntEntry( vNode2Var, iNodeIdOld );
        if iSatVarOld == 0 { // iNodeIdOld in the original AIG has no SAT var
         
            // map the corresponding original AIG node into this SAT var
            Vec_IntWriteEntry( vNode2Var, iNodeIdOld, pCnf.pVarNums[Aig_ObjId(pObj)] );
            continue;
        }
        // add connector clauses 
        Lits[0] = toLitCond( iSatVarOld, 0 );
        Lits[1] = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 1 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
        Lits[0] = toLitCond( iSatVarOld, 1 );
        Lits[1] = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 0 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
    }
    // derive the connector clauses
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        iNodeIdOld = Vec_IntEntry( vPioIds, Aig_ManCiNum(pAig) + i );
        iSatVarOld = Vec_IntEntry( vNode2Var, iNodeIdOld );
        if iSatVarOld == 0 { // iNodeIdOld in the original AIG has no SAT var
         
            // map the corresponding original AIG node into this SAT var
            Vec_IntWriteEntry( vNode2Var, iNodeIdOld, pCnf.pVarNums[Aig_ObjId(pObj)] );
            continue;
        }
        // add connector clauses 
        Lits[0] = toLitCond( iSatVarOld, 0 );
        Lits[1] = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 1 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
        Lits[0] = toLitCond( iSatVarOld, 1 );
        Lits[1] = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 0 );
        if !sat_solver_addclause( pSat, Lits, Lits+2 ) {
            assert(false, "0"); }
    }
    // transfer the ID of constant 1 node
    if Vec_IntEntry( vNode2Var, 0 ) == 0 {
        Vec_IntWriteEntry( vNode2Var, 0, pCnf.pVarNums[0] ); }
    // remove the CNF
    Cnf_DataFree( pCnf );
    // constrain the solver with the literals corresponding to the original POs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPartPos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vPartPos, i)), 1); i++  {
        iNodeIdOld = Aig_ObjFaninId0( pObj );
        iSatVarOld = Vec_IntEntry( vNode2Var, iNodeIdOld );
        assert(iSatVarOld != 0, "iSatVarOld != 0");
        // assert the original PO to be 1
        Lits[0] = toLitCond( iSatVarOld, Aig_ObjFaninC0(pObj) );
        // correct the polarity if polarity alignment is enabled
        if (fAlignPol != 0) && Aig_ObjFanin0(pObj).fPhase { 
            Lits[0] = lit_neg( Lits[0] ); }
        if !sat_solver_addclause( pSat, Lits, Lits+1 ) {
            assert(false, "0"); // if it happens, can return 1 (unsatisfiable)
            return 1;
        }
    }
    // constrain some the primary inputs to constant values
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        if !pObj.fMarkA && !pObj.fMarkB {
            continue; }
        iNodeIdOld = Vec_IntEntry( vPioIds, i );
        iSatVarOld = Vec_IntEntry( vNode2Var, iNodeIdOld );
        Lits[0] = toLitCond( iSatVarOld, pObj.fMarkA );
        if !sat_solver_addclause( pSat, Lits, Lits+1 ) {
            assert(false, "0"); // if it happens, can return 1 (unsatisfiable)
            return 1;
        }
        pObj.fMarkA = pObj.fMarkB = 0;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs partitioned SAT solving.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartitionedSat(p *Aig_Man_t, nAlgo int32, nPartSize int32, nConfPart int32, nConfTotal int32, fAlignPol int32, fSynthesize int32, fVerbose int32) int32 {
    var pSat *sat_solver;
    var vAigs *Vec_Ptr_t;
    var vPio2Id, vPart2Pos *Vec_Vec_t;
    var pAig, pTemp *Aig_Man_t;
    var vNode2Part, vNode2Var *Vec_Int_t;
    var nConfRemaining int32 = nConfTotal;
    var nNodes int32;
    var i int32;
    var status int32;
    var RetValue int32 = -1;
    var clk abctime;

    // perform partitioning according to the selected algorithm
    clk = Abc_Clock();
    switch  nAlgo  {
    case 0: 
        vNode2Part = Aig_ManPartitionMonolithic( p );
        break;
    case 1: 
        vNode2Part = Aig_ManPartitionLevelized( p, nPartSize );
        break;
    case 2: 
        vNode2Part = Aig_ManPartitionDfs( p, nPartSize, 0 );
        break;
    case 3: 
        vNode2Part = Aig_ManPartitionDfs( p, nPartSize, 1 );
        break;
    default:
        fmt.Printf("Unknown partitioning algorithm.\n");
        return -1;
    }

    if fVerbose != 0 {
    fmt.Printf("Partitioning derived %d partitions. ", Vec_IntFindMax(vNode2Part) + 1);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

    // split the original AIG into partition AIGs (vAigs)
    // also, derives mapping of PIs/POs of partition AIGs into original nodes
    // also, derives mapping of POs of the original AIG into partitions
    vAigs = Aig_ManPartSplit( p, vNode2Part, (Vec_Ptr_t **)&vPio2Id, (Vec_Ptr_t **)&vPart2Pos );
    Vec_IntFree( vNode2Part );

    if fVerbose != 0 {
    fmt.Printf("Partions were transformed into AIGs. ");
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

    // synthesize partitions
    if fSynthesize != 0 {
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        pAig = Dar_ManRwsat( pTemp = pAig, 0, 0 );
        Vec_PtrWriteEntry( vAigs, i, pAig );
        Aig_ManStop( pTemp );
    } }

    // start the SAT solver
    pSat = sat_solver_new();
//    pSat->verbosity = fVerbose;
    // start mapping of the original AIG IDs into their SAT variable numbers
    vNode2Var = Vec_IntStart( Aig_ManObjNumMax(p) );

    // add partitions, one at a time, and run the SAT solver 
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pAig) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
clk = Abc_Clock();
        // transform polarity of the AIG
        if fAlignPol != 0 {
            Aig_ManPartSetNodePolarity( p, pAig, Vec_VecEntryInt(vPio2Id,i) );
        } else {
            Aig_ManPartResetNodePolarity( pAig ); }
        // add CNF of this partition to the SAT solver
        if Aig_ManAddNewCnfToSolver( pSat, pAig, vNode2Var, 
            Vec_VecEntryInt(vPio2Id,i), Vec_VecEntry(vPart2Pos,i), fAlignPol ) != 0 {
            RetValue = 1;
            break;
        }
        // call the SAT solver
        status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfRemaining, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if fVerbose != 0 {
            fmt.Printf("%4d : Aig = %6d. Vs = %7d. RootCs = %7d. LearnCs = %6d. ", i, nNodes += Aig_ManNodeNum(pAig), sat_solver_nvars(pSat), int32(pSat.stats.clauses), int32(pSat.stats.learnts));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
        // analize the result
        if status == l_False {
            RetValue = 1;
            break;
        
        } else if status == l_True {
            RetValue = 0;
        } else {
            RetValue = -1; }
        nConfRemaining -= pSat.stats.conflicts;
        if nConfRemaining <= 0 {
            fmt.Printf("Exceeded the limit on the total number of conflicts (%d).\n", nConfTotal);
            break;
        }
    }
    if RetValue == 0 {
        Aig_ManDeriveCounterExample( p, vNode2Var, pSat ); }
    // cleanup
    sat_solver_delete( pSat );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAigs)) && (((pTemp) = (Aig_Man_t *)Vec_PtrEntry(vAigs, i)), 1); i++  {
        Aig_ManStop( pTemp );
    }
    Vec_PtrFree( vAigs );
    Vec_VecFree( vPio2Id );
    Vec_VecFree( vPart2Pos );
    Vec_IntFree( vNode2Var );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

