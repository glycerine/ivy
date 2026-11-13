package abs

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [absGla2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Scalable gate-level abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absGla2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/main/main.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver2.h"
//#include "bool/kit/kit.h"
//#include "abs.h"
//#include "absRef.h"
//#include "absRef2.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const GA2_BIG_NUM = 0x3FFFFFF0

    // manager
////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes truth table for the marked node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Returns AIG marked for CNF generation.]

  Description [The marking satisfies the following requirements:
  Each marked node has the number of marked fanins no more than N.]
               
  SideEffects [Uses pObj->fPhase to store the markings.]

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ga2_ManDumpStats(pGia *Gia_Man_t, pPars *Abs_Par_t, pSat *sat_solver2, iFrame int32, fUseN int32) {
    var pFile *os.File;
    var pFileName string // was [32]byte in C

    pFileName = fmt.Sprintf("stats_gla%s%s.txt", fUseN ? "n":"", pPars.fUseFullProof ? "p":""); 

    pFile = os.OpenFile(pFileName, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);

    fmt.Fprintf(pFile, "%s pi=%d ff=%d and=%d mem=%d bmc=%d", pGia.pName, Gia_ManPiNum(pGia), Gia_ManRegNum(pGia), Gia_ManAndNum(pGia), int32((1 + sat_solver2_memory_proof(pSat)/(1<<20))), iFrame);

    if pGia.vGateClasses {
    fmt.Fprintf(pFile, " ff=%d and=%d", Gia_GlaCountFlops( pGia, pGia.vGateClasses ), Gia_GlaCountNodes( pGia, pGia.vGateClasses )); }

    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}
func Ga2_ManReportMemory(p *Ga2_Man_t) {
    var memTot float64;
    var memAig float64  = 1.0 * p.pGia.nObjsAlloc * sizeof(Gia_Obj_t) + Vec_IntMemory(p.pGia.vMapping);
    var memSat float64  = sat_solver2_memory( p.pSat, 1 );
    var memPro float64  = sat_solver2_memory_proof( p.pSat );
    var memMap float64  = Vec_VecMemoryInt( (Vec_Vec_t *)p.vId2Lit );
    var memRef float64  = Rnm_ManMemoryUsage( p.pRnm );
    var memHash float64 = sizeof(int) * 6 * p.nTable;
    var memOth float64  = sizeof(Ga2_Man_t);
    memOth += Vec_VecMemoryInt( (Vec_Vec_t *)p.vCnfs );
    memOth += Vec_IntMemory( p.vIds );
    memOth += Vec_IntMemory( p.vProofIds );
    memOth += Vec_IntMemory( p.vAbs );
    memOth += Vec_IntMemory( p.vValues );
    memOth += Vec_IntMemory( p.vLits );
    memOth += Vec_IntMemory( p.vIsopMem );
    memOth += 336450 + (sizeof(char) + sizeof(char*)) * 65536;
    memTot = memAig + memSat + memPro + memMap + memRef + memHash + memOth;
    ABC_PRMP( "Memory: AIG      ", memAig, memTot );
    ABC_PRMP( "Memory: SAT      ", memSat, memTot );
    ABC_PRMP( "Memory: Proof    ", memPro, memTot );
    ABC_PRMP( "Memory: Map      ", memMap, memTot );
    ABC_PRMP( "Memory: Refine   ", memRef, memTot );
    ABC_PRMP( "Memory: Hash     ", memHash,memTot );
    ABC_PRMP( "Memory: Other    ", memOth, memTot );
    ABC_PRMP( "Memory: TOTAL    ", memTot, memTot );
}
/**Function*************************************************************

  Synopsis    [Computes a minimized truth table.]

  Description [Input literals can be 0/1 (const 0/1), non-trivial literals 
  (integers that are more than 1) and unassigned literals (large integers).
  This procedure computes the truth table that essentially depends on input
  variables ordered in the increasing order of their positive literals.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ga2_ObjComputeTruthSpecial_uTruth5 [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
/**Function*************************************************************

  Synopsis    [Returns CNF of the function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derives CNF for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Saig_ManBmcHashKey_s_Primes [5]int32 = [5]int32{ 12582917, 25165843, 50331653, 100663319, 201326611 };
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ga2_ManSetupNode(p *Ga2_Man_t, pObj *Gia_Obj_t, fAbs int32) {
    var uTruth uint32;
    var nLeaves int32;
//    int Id = Gia_ObjId(p->pGia, pObj);
    assert(pObj.fPhase, "pObj.fPhase");
    assert(Vec_PtrSize(p.vCnfs) == 2 * Vec_IntSize(p.vValues), "Vec_PtrSize(p.vCnfs) == 2 * Vec_IntSize(p.vValues)");
    // assign abstraction ID to the node
    if Ga2_ObjId(p,pObj) == -1 {
        Ga2_ObjSetId( p, pObj, Vec_IntSize(p.vValues) );
        Vec_IntPush( p.vValues, Gia_ObjId(p.pGia, pObj) );
        Vec_PtrPush( p.vCnfs, nil );
        Vec_PtrPush( p.vCnfs, nil );
    }
    assert(Ga2_ObjCnf0(p, pObj) == nil, "Ga2_ObjCnf0(p, pObj) == nil");
    if !(fAbs != 0) {
        return; }
    Vec_IntPush( p.vAbs, Gia_ObjId(p.pGia, pObj) );
    assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj)");
    // compute parameters
    nLeaves = Ga2_ObjLeaveNum(p.pGia, pObj);
    uTruth = Ga2_ObjTruth( p.pGia, pObj );
    // create CNF for pos/neg phases
    Vec_PtrWriteEntry( p.vCnfs, 2 * Ga2_ObjId(p,pObj),     Ga2_ManCnfCompute( uTruth, nLeaves, p.vIsopMem) );    
    Vec_PtrWriteEntry( p.vCnfs, 2 * Ga2_ObjId(p,pObj) + 1, Ga2_ManCnfCompute(~uTruth, nLeaves, p.vIsopMem) );
}

func Ga2_ManAddToAbsOneStatic(p *Ga2_Man_t, pObj *Gia_Obj_t, f int32, fUseId int32) {
    var vLeaves *Vec_Int_t;
    var pLeaf *Gia_Obj_t;
    var k int32;
    var Lit int32;
    var iLitOut int32 = Ga2_ObjFindOrAddLit( p, pObj, f );
    assert(iLitOut > 1, "iLitOut > 1");
    assert(Gia_ObjIsConst0(pObj) || Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsConst0(pObj) || Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj)");
    if Gia_ObjIsConst0(pObj) || (f == 0 && Gia_ObjIsRo(p.pGia, pObj)) {
        iLitOut = Abc_LitNot( iLitOut );
        sat_solver2_addclause( p.pSat, &iLitOut, &iLitOut + 1, fUseId ? Gia_ObjId(p.pGia, pObj) : -1 );
    
    } else {
        var fUseStatic int32  = 1;
        Vec_IntClear( p.vLits );
        vLeaves = Ga2_ObjLeaves( p.pGia, pObj );
        Gia_ManForEachObjVec( vLeaves, p.pGia, pLeaf, k )
        {
            Lit = Ga2_ObjFindOrAddLit( p, pLeaf, f - Gia_ObjIsRo(p.pGia, pObj) );
            Vec_IntPush( p.vLits, Lit );
            if Lit < 2 {
                fUseStatic = 0; }
        }
        if (fUseStatic != 0) || Gia_ObjIsRo(p.pGia, pObj) {
            Ga2_ManCnfAddStatic( p.pSat, Ga2_ObjCnf0(p, pObj), Ga2_ObjCnf1(p, pObj), Vec_IntArray(p.vLits), iLitOut, fUseId ? Gia_ObjId(p.pGia, pObj) : -1 );
        } else {
            var uTruth uint32  = Ga2_ObjComputeTruthSpecial( p.pGia, pObj, vLeaves, p.vLits );
            Ga2_ManCnfAddDynamic( p, uTruth & 0xFFFF, Vec_IntArray(p.vLits), iLitOut, Gia_ObjId(p.pGia, pObj) );
        }
    }
}
func Ga2_ManAddToAbsOneDynamic(p *Ga2_Man_t, pObj *Gia_Obj_t, f int32) {
//    int Id = Gia_ObjId(p->pGia, pObj);
    var vLeaves *Vec_Int_t;
    var pLeaf *Gia_Obj_t;
    var uTruth uint32;
    var i, Lit int32;

    assert((Ga2_ObjIsAbs0(p, pObj) != 0), "Ga2_ObjIsAbs0(p, pObj)");
    assert(Gia_ObjIsConst0(pObj) || Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsConst0(pObj) || Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsAnd(pObj)");
    if Gia_ObjIsConst0(pObj) || (f == 0 && Gia_ObjIsRo(p.pGia, pObj)) {
        Ga2_ObjAddLit( p, pObj, f, 0 );
    
    } else if Gia_ObjIsRo(p.pGia, pObj) {
        // if flop is included in the abstraction, but its driver is not
        // flop input driver has no variable assigned -- we assign it here
        pLeaf = Gia_ObjRoToRi( p.pGia, pObj );
        Lit = Ga2_ObjFindOrAddLit( p, Gia_ObjFanin0(pLeaf), f-1 );
        assert(Lit >= 0, "Lit >= 0");
        Lit = Abc_LitNotCond( Lit, Gia_ObjFaninC0(pLeaf) );
        Ga2_ObjAddLit( p, pObj, f, Lit );
    
    } else {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        Vec_IntClear( p.vLits );
        vLeaves = Ga2_ObjLeaves( p.pGia, pObj );
        Gia_ManForEachObjVec( vLeaves, p.pGia, pLeaf, i )
        {
            if Ga2_ObjIsAbs0(p, pLeaf) != 0 {      // belongs to original abstraction
            
                Lit = Ga2_ObjFindLit( p, pLeaf, f );
                assert(Lit >= 0, "Lit >= 0");
            
            } else if Ga2_ObjIsLeaf0(p, pLeaf) != 0 { // belongs to original PPIs
            
                Lit = Ga2_ObjFindLit( p, pLeaf, f );
//                Lit = Ga2_ObjFindOrAddLit( p, pLeaf, f );
                if Lit == -1 {
                    Lit = GA2_BIG_NUM + 2*i;
//                    assert( 0 );
                }
            
            } else { assert(false, "0"); }
            Vec_IntPush( p.vLits, Lit );
        }

        // minimize truth table
        uTruth = Ga2_ObjComputeTruthSpecial( p.pGia, pObj, vLeaves, p.vLits );
        if uTruth == 0 || uTruth == ~0 { // const 0 / 1
        
            Lit = (uTruth > 0);
            Ga2_ObjAddLit( p, pObj, f, Lit );
        
        } else if uTruth == 0xAAAAAAAA || uTruth == 0x55555555 {  // buffer / inverter
        
            Lit = Vec_IntEntry( p.vLits, 0 );
            if Lit >= GA2_BIG_NUM {
                pLeaf = Gia_ManObj( p.pGia, Vec_IntEntry(vLeaves, (Lit-GA2_BIG_NUM)/2) );
                Lit = Ga2_ObjFindLit( p, pLeaf, f );
                assert(Lit == -1, "Lit == -1");
                Lit = Ga2_ObjFindOrAddLit( p, pLeaf, f );
            }
            assert(Lit >= 0, "Lit >= 0");
            Lit = Abc_LitNotCond( Lit, uTruth == 0x55555555 );
            Ga2_ObjAddLit( p, pObj, f, Lit );
            assert(Lit < 10000000, "Lit < 10000000");
        
        } else {
            assert(Vec_IntSize(p.vLits) > 1 && Vec_IntSize(p.vLits) < 6, "Vec_IntSize(p.vLits) > 1 && Vec_IntSize(p.vLits) < 6");
            // replace literals
            Vec_IntForEachEntry( p.vLits, Lit, i )
            {
                if Lit >= GA2_BIG_NUM {
                    pLeaf = Gia_ManObj( p.pGia, Vec_IntEntry(vLeaves, (Lit-GA2_BIG_NUM)/2) );
                    Lit = Ga2_ObjFindLit( p, pLeaf, f );
                    assert(Lit == -1, "Lit == -1");
                    Lit = Ga2_ObjFindOrAddLit( p, pLeaf, f );
                    Vec_IntWriteEntry( p.vLits, i, Lit );
                }
                assert(Lit < 10000000, "Lit < 10000000");
            }

            // add new nodes
            if Vec_IntSize(p.vLits) == 5 {
                Vec_IntClear( p.vLits );
                Gia_ManForEachObjVec( vLeaves, p.pGia, pLeaf, i )
                    Vec_IntPush( p.vLits, Ga2_ObjFindOrAddLit( p, pLeaf, f ) );
                Lit = Ga2_ObjFindOrAddLit(p, pObj, f);
                Ga2_ManCnfAddStatic( p.pSat, Ga2_ObjCnf0(p, pObj), Ga2_ObjCnf1(p, pObj), Vec_IntArray(p.vLits), Lit, -1 );
            
            } else {
//                int fUseHash = 1;
                if !p.pPars.fSkipHash {
                    var pLookup *int32;
                    var nSize int32 = Vec_IntSize(p.vLits);
                    assert(Vec_IntSize(p.vLits) < 5, "Vec_IntSize(p.vLits) < 5");
                    assert(Vec_IntEntry(p.vLits, 0) <= Vec_IntEntryLast(p.vLits), "Vec_IntEntry(p.vLits, 0) <= Vec_IntEntryLast(p.vLits)");
                    assert(Ga2_ObjFindLit(p, pObj, f) == -1, "Ga2_ObjFindLit(p, pObj, f) == -1");
                    for  i = Vec_IntSize(p.vLits); i < 4; i++  {
                        Vec_IntPush( p.vLits, GA2_BIG_NUM );
                    }
                    Vec_IntPush( p.vLits, int32(uTruth) );
                    assert(Vec_IntSize(p.vLits) == 5, "Vec_IntSize(p.vLits) == 5");

                    // perform structural hashing here!!!
                    pLookup = Saig_ManBmcLookup( p, Vec_IntArray(p.vLits) );
                    if *pLookup == 0 {
                        *pLookup = Ga2_ObjFindOrAddLit(p, pObj, f);
                        Vec_IntShrink( p.vLits, nSize );
                        Ga2_ManCnfAddDynamic( p, uTruth & 0xFFFF, Vec_IntArray(p.vLits), *pLookup, -1 );
                    
                    } else {
                        Ga2_ObjAddLit( p, pObj, f, *pLookup ); }

                
                } else {
                    Lit = Ga2_ObjFindOrAddLit(p, pObj, f);
                    Ga2_ManCnfAddDynamic( p, uTruth & 0xFFFF, Vec_IntArray(p.vLits), Lit, -1 );
                }
            }
        }
    }
}

func Ga2_ManAddAbsClauses(p *Ga2_Man_t, f int32) {
    var fSimple int32;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObjVec( p.vValues, p.pGia, pObj, i )
    {
        if i == p.LimAbs {
            break; }
        if fSimple != 0 {
            Ga2_ManAddToAbsOneStatic( p, pObj, f, 0 );
        } else {
            Ga2_ManAddToAbsOneDynamic( p, pObj, f ); }
    }
    Gia_ManForEachObjVec( p.vAbs, p.pGia, pObj, i )
        if ( i >= p.LimAbs )
            Ga2_ManAddToAbsOneStatic( p, pObj, f, 1 );
//    sat_solver2_simplify( p->pSat );
}

func Ga2_ManAddToAbs(p *Ga2_Man_t, vToAdd *Vec_Int_t) {
    var vLeaves *Vec_Int_t;
    var pObj, pFanin *Gia_Obj_t;
    var f, i, k int32;
    // add abstraction objects
    Gia_ManForEachObjVec( vToAdd, p.pGia, pObj, i )
    {
        Ga2_ManSetupNode( p, pObj, 1 );
        if p.pSat.pPrf2 {
            Vec_IntWriteEntry( p.vProofIds, Gia_ObjId(p.pGia, pObj), p.nProofIds++ ); }
    }
    // add PPI objects
    Gia_ManForEachObjVec( vToAdd, p.pGia, pObj, i )
    {
        vLeaves = Ga2_ObjLeaves( p.pGia, pObj );
        Gia_ManForEachObjVec( vLeaves, p.pGia, pFanin, k )
            if ( Ga2_ObjId( p, pFanin ) == -1 )
                Ga2_ManSetupNode( p, pFanin, 0 );
    }
    // add new clauses to the timeframes
    for  f = 0; f <= p.pPars.iFrame; f++  {
        Vec_IntFillExtra( Ga2_MapFrameMap(p, f), Vec_IntSize(p.vValues), -1 );
        Gia_ManForEachObjVec( vToAdd, p.pGia, pObj, i )
            Ga2_ManAddToAbsOneStatic( p, pObj, f, 1 );
    }
//    sat_solver2_simplify( p->pSat );
}

func Ga2_ManShrinkAbs(p *Ga2_Man_t, nAbs int32, nValues int32, nSatVars int32) {
    var vMap *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, k, Entry int32;
    assert(nAbs > 0, "nAbs > 0");
    assert(nValues > 0, "nValues > 0");
    assert(nSatVars > 0, "nSatVars > 0");
    // shrink abstraction
    Gia_ManForEachObjVec( p.vAbs, p.pGia, pObj, i )
    {
        if !(i != 0) { continue; }
        assert(Ga2_ObjCnf0(p, pObj) != nil, "Ga2_ObjCnf0(p, pObj) != nil");
        assert(Ga2_ObjCnf1(p, pObj) != nil, "Ga2_ObjCnf1(p, pObj) != nil");
        if i < nAbs {
            continue; }
        Vec_IntFree( Ga2_ObjCnf0(p, pObj) );
        Vec_IntFree( Ga2_ObjCnf1(p, pObj) );
        Vec_PtrWriteEntry( p.vCnfs, 2 * Ga2_ObjId(p,pObj),     nil );    
        Vec_PtrWriteEntry( p.vCnfs, 2 * Ga2_ObjId(p,pObj) + 1, nil );
    }
    Vec_IntShrink( p.vAbs, nAbs );
    // shrink values
    Gia_ManForEachObjVec( p.vValues, p.pGia, pObj, i )
    {
        assert(Ga2_ObjId(p,pObj) >= 0, "Ga2_ObjId(p,pObj) >= 0");
        if i < nValues {
            continue; }
        Ga2_ObjSetId( p, pObj, -1 );
    }
    Vec_IntShrink( p.vValues, nValues );
    Vec_PtrShrink( p.vCnfs, 2 * nValues );
    // hack to clear constant
    if nValues == 1 {
        nValues = 0; }
    // clean mapping for each timeframe
    Vec_PtrForEachEntry( Vec_Int_t *, p.vId2Lit, vMap, i )
    {
        Vec_IntShrink( vMap, nValues );
        Vec_IntForEachEntryStart( vMap, Entry, k, p.LimAbs )
            if ( Entry >= 2*nSatVars )
                Vec_IntWriteEntry( vMap, k, -1 );
    }
    // shrink SAT variables
    p.nSatVars = nSatVars;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ga2_ManAbsTranslate_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vClasses *Vec_Int_t, fFirst int32) {
    if pObj.fPhase && !(fFirst != 0) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Ga2_ManAbsTranslate_rec( p, Gia_ObjFanin0(pObj), vClasses, 0 );
    Ga2_ManAbsTranslate_rec( p, Gia_ObjFanin1(pObj), vClasses, 0 );
    Vec_IntWriteEntry( vClasses, Gia_ObjId(p, pObj), 1 );
}

func Ga2_ManAbsTranslate(p *Ga2_Man_t) *Vec_Int_t {
    var vGateClasses *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    vGateClasses = Vec_IntStart( Gia_ManObjNum(p.pGia) );
    Vec_IntWriteEntry( vGateClasses, 0, 1 );
    Gia_ManForEachObjVec( p.vAbs, p.pGia, pObj, i )
    {
        if Gia_ObjIsAnd(pObj) {
            Ga2_ManAbsTranslate_rec( p.pGia, pObj, vGateClasses, 1 );
        } else if Gia_ObjIsRo(p.pGia, pObj) {
            Vec_IntWriteEntry( vGateClasses, Gia_ObjId(p.pGia, pObj), 1 );
        } else if !Gia_ObjIsConst0(pObj) {
            assert(false, "0"); }
//        Gia_ObjPrint( p->pGia, pObj );
    }
    return vGateClasses;
}

func Ga2_ManRestart(p *Ga2_Man_t) {
    var vToAdd *Vec_Int_t;
    var Lit int32  = 1;
    assert(p.pGia != nil && p.pGia.vGateClasses != nil, "p.pGia != nil && p.pGia.vGateClasses != nil");
    assert(Gia_ManPi(p.pGia, 0).fPhase, "Gia_ManPi(p.pGia, 0).fPhase"); // marks are set
    // clear SAT variable numbers (begin with 1)
    if p.pSat { sat_solver2_delete( p.pSat ); }
    p.pSat      = sat_solver2_new();
    p.pSat.nLearntStart = p.pPars.nLearnedStart;
    p.pSat.nLearntDelta = p.pPars.nLearnedDelta;
    p.pSat.nLearntRatio = p.pPars.nLearnedPerce;
    p.pSat.nLearntMax   = p.pSat.nLearntStart;
    // add clause x0 = 0  (lit0 = 1; lit1 = 0)
    sat_solver2_addclause( p.pSat, &Lit, &Lit + 1, -1 );
    // remove previous abstraction
    Ga2_ManShrinkAbs( p, 1, 1, 1 );
    // start new abstraction
    vToAdd = Ga2_ManAbsDerive( p.pGia );
    assert(p.pSat.pPrf2 == nil, "p.pSat.pPrf2 == nil");
    assert(p.pPars.iFrame < 0, "p.pPars.iFrame < 0");
    Ga2_ManAddToAbs( p, vToAdd );
    Vec_IntFree( vToAdd );
    p.LimAbs = Vec_IntSize(p.vAbs);
    p.LimPpi = Vec_IntSize(p.vValues);
    // set runtime limit
    if p.pPars.nTimeOut {
        sat_solver2_set_runtime_limit( p.pSat, p.pPars.nTimeOut * int64(time.Second) + p.timeStart ); }
    // clean the hash table
    memset( p.pTable, 0, 6 * sizeof(int) * p.nTable );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ga2_ObjSatValue(p *Ga2_Man_t, pObj *Gia_Obj_t, f int32) int32 {
    var Lit int32  = Ga2_ObjFindLit( p, pObj, f );
    assert(!Gia_ObjIsConst0(pObj), "!Gia_ObjIsConst0(pObj)");
    if Lit == -1 {
        return 0; }
    if Abc_Lit2Var(Lit) >= p.pSat.size {
        return 0; }
    return Abc_LitIsCompl(Lit) ^ sat_solver2_var_value( p.pSat, Abc_Lit2Var(Lit) );
}
func Ga2_ManDeriveCex(p *Ga2_Man_t, vPis *Vec_Int_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    pCex = Abc_CexAlloc( Gia_ManRegNum(p.pGia), Gia_ManPiNum(p.pGia), p.pPars.iFrame+1 );
    pCex.iPo = 0;
    pCex.iFrame = p.pPars.iFrame;
    Gia_ManForEachObjVec( vPis, p.pGia, pObj, i )
    {
        if !Gia_ObjIsPi(p.pGia, pObj) {
            continue; }
        assert(Gia_ObjIsPi(p.pGia, pObj), "Gia_ObjIsPi(p.pGia, pObj)");
        for  f = 0; f <= pCex.iFrame; f++  {
            if Ga2_ObjSatValue( p, pObj, f ) != 0 {
                Abc_InfoSetBit( pCex.pData, pCex.nRegs + f * pCex.nPis + Gia_ObjCioId(pObj) ); }
        }
    }
    return pCex;
} 
func Ga2_ManRefinePrint(p *Ga2_Man_t, vVec *Vec_Int_t) {
    var pObj, pFanin *Gia_Obj_t;
    var i, k int32;
    fmt.Printf("\n         Unsat core: \n");
    Gia_ManForEachObjVec( vVec, p.pGia, pObj, i )
    {
        var vLeaves *Vec_Int_t  = Ga2_ObjLeaves( p.pGia, pObj );
        fmt.Printf("%12d : ", i);
        fmt.Printf("Obj =%6d ", Gia_ObjId(p.pGia, pObj));
        if Gia_ObjIsRo(p.pGia, pObj) {
            fmt.Printf("ff ");
        } else {
            fmt.Printf("   "); }
        if Ga2_ObjIsAbs0(p, pObj) != 0 {
            fmt.Printf("a ");
        } else if Ga2_ObjIsLeaf0(p, pObj) != 0 {
            fmt.Printf("l ");
        } else { 
            fmt.Printf("  "); }
        fmt.Printf("Fanins: ");
        Gia_ManForEachObjVec( vLeaves, p.pGia, pFanin, k )
        {
            fmt.Printf("%6d ", Gia_ObjId(p.pGia, pFanin));
            if Gia_ObjIsRo(p.pGia, pFanin) {
                fmt.Printf("ff ");
            } else {
                fmt.Printf("   "); }
            if Ga2_ObjIsAbs0(p, pFanin) != 0 {
                fmt.Printf("a ");
            } else if Ga2_ObjIsLeaf0(p, pFanin) != 0 {
                fmt.Printf("l ");
            } else {
                fmt.Printf("  "); }
        }
        fmt.Printf("\n");
    }
}
func Ga2_ManRefinePrintPPis(p *Ga2_Man_t) {
    var vVec *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObjVec( p.vValues, p.pGia, pObj, i )
    {
        if !(i != 0) { continue; }
        if Ga2_ObjIsAbs(p, pObj) != 0 {
            continue; }
        assert(pObj.fPhase, "pObj.fPhase");
        assert((Ga2_ObjIsLeaf(p, pObj) != 0), "Ga2_ObjIsLeaf(p, pObj)");
        assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsCi(pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsCi(pObj)");
        Vec_IntPush( vVec, Gia_ObjId(p.pGia, pObj) );
    }
    fmt.Printf("        Current PPIs (%d): ", Vec_IntSize(vVec));
    Vec_IntSort( vVec, 1 );
    Gia_ManForEachObjVec( vVec, p.pGia, pObj, i )
        fmt.Printf("%d ", Gia_ObjId(p.pGia, pObj));
    fmt.Printf("\n");
    Vec_IntFree( vVec );
}


func Ga2_GlaPrepareCexAndMap(p *Ga2_Man_t, ppCex **Abc_Cex_t, pvMaps **Vec_Int_t) {
    var pCex *Abc_Cex_t;
    var vMap *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var f, i, k int32;
/*
    Gia_ManForEachObj( p->pGia, pObj, i )
        if ( Ga2_ObjId(p, pObj) >= 0 )
            assert( Vec_IntEntry(p->vValues, Ga2_ObjId(p, pObj)) == i );
*/
    // find PIs and PPIs
    vMap = Vec_IntAlloc( 1000 );
    Gia_ManForEachObjVec( p.vValues, p.pGia, pObj, i )
    {
        if !(i != 0) { continue; }
        if Ga2_ObjIsAbs(p, pObj) != 0 {
            continue; }
        assert(pObj.fPhase, "pObj.fPhase");
        assert((Ga2_ObjIsLeaf(p, pObj) != 0), "Ga2_ObjIsLeaf(p, pObj)");
        assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsCi(pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsCi(pObj)");
        Vec_IntPush( vMap, Gia_ObjId(p.pGia, pObj) );
    }
    // derive counter-example
    pCex = Abc_CexAlloc( 0, Vec_IntSize(vMap), p.pPars.iFrame+1 );
    pCex.iFrame = p.pPars.iFrame;
    for  f = 0; f <= p.pPars.iFrame; f++  {
        Gia_ManForEachObjVec( vMap, p.pGia, pObj, k )
            if ( Ga2_ObjSatValue( p, pObj, f ) )
                Abc_InfoSetBit( pCex.pData, f * Vec_IntSize(vMap) + k );
    }
    *pvMaps = vMap;
    *ppCex = pCex;
}
func Ga2_ManRefine(p *Ga2_Man_t) *Vec_Int_t {
    var pCex *Abc_Cex_t;
    var vMap, vVec *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, k int32;
    if p.pPars.fAddLayer {
        // use simplified refinement strategy, which adds logic near at PPI without finding important ones
        vVec = Vec_IntAlloc( 100 );
        Gia_ManForEachObjVec( p.vValues, p.pGia, pObj, i )
        {
            if !(i != 0) { continue; }
            if Ga2_ObjIsAbs(p, pObj) != 0 {
                continue; }
            assert(pObj.fPhase, "pObj.fPhase");
            assert((Ga2_ObjIsLeaf(p, pObj) != 0), "Ga2_ObjIsLeaf(p, pObj)");
            assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsCi(pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsCi(pObj)");
            if Gia_ObjIsPi(p.pGia, pObj) {
                continue; }
            Vec_IntPush( vVec, Gia_ObjId(p.pGia, pObj) );
        }
        p.nObjAdded += Vec_IntSize(vVec);
        return vVec;
    }
    Ga2_GlaPrepareCexAndMap( p, &pCex, &vMap );
 //    Rf2_ManRefine( p->pRf2, pCex, vMap, p->pPars->fPropFanout, 1 );
    vVec = Rnm_ManRefine( p.pRnm, pCex, vMap, p.pPars.fPropFanout, p.pPars.fNewRefine, 1 );
//    printf( "Refinement %d\n", Vec_IntSize(vVec) );
    Abc_CexFree( pCex );
    if Vec_IntSize(vVec) == 0 {
        Vec_IntFree( vVec );
        Abc_CexFreeP( &p.pGia.pCexSeq );
        p.pGia.pCexSeq = Ga2_ManDeriveCex( p, vMap );
        Vec_IntFree( vMap );
        return nil;
    }
    Vec_IntFree( vMap );
    // remove those already added
    k = 0;
    Gia_ManForEachObjVec( vVec, p.pGia, pObj, i )
        if ( !Ga2_ObjIsAbs(p, pObj) )
            Vec_IntWriteEntry( vVec, k++, Gia_ObjId(p.pGia, pObj) );
    Vec_IntShrink( vVec, k );

    // these objects should be PPIs that are not abstracted yet
    Gia_ManForEachObjVec( vVec, p.pGia, pObj, i )
        assert(pObj.fPhase, "pObj.fPhase");//&& Ga2_ObjIsLeaf(p, pObj) );
    p.nObjAdded += Vec_IntSize(vVec);
    return vVec;
}

/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ga2_GlaAbsCount(p *Ga2_Man_t, fRo int32, fAnd int32) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    if fRo != 0 {
        Gia_ManForEachObjVec( p.vAbs, p.pGia, pObj, i )
            Counter += Gia_ObjIsRo(p.pGia, pObj);
    } else if fAnd != 0 {
        Gia_ManForEachObjVec( p.vAbs, p.pGia, pObj, i )
            Counter += Gia_ObjIsAnd(pObj);
    } else { assert(false, "0"); }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ga2_ManAbsPrintFrame(p *Ga2_Man_t, nFrames int32, nConfls int32, nCexes int32, Time abctime, fFinal int32) {
    var fUseNewLine int32  = ((fFinal && nCexes) || p.pPars.fVeryVerbose);
    if Abc_FrameIsBatchMode() && !(fUseNewLine != 0) {
        return; }
    p.fUseNewLine = fUseNewLine;
    Abc_Print( 1, "%4d :", nFrames );
    Abc_Print( 1, "%4d", Abc_MinInt(100, 100 * Vec_IntSize(p.vAbs) / p.nMarked) ); 
    Abc_Print( 1, "%6d", Vec_IntSize(p.vAbs) );
    Abc_Print( 1, "%5d", Vec_IntSize(p.vValues)-Vec_IntSize(p.vAbs)-1 );
    Abc_Print( 1, "%5d", Ga2_GlaAbsCount(p, 1, 0) );
    Abc_Print( 1, "%6d", Ga2_GlaAbsCount(p, 0, 1) );
    Abc_Print( 1, "%8d", nConfls );
    if nCexes == 0 {
        Abc_Print( 1, "%5c", '-' ); 
    } else {
        Abc_Print( 1, "%5d", nCexes ); } 
    Abc_PrintInt( sat_solver2_nvars(p.pSat) );
    Abc_PrintInt( sat_solver2_nclauses(p.pSat) );
    Abc_PrintInt( sat_solver2_nlearnts(p.pSat) );
    Abc_Print( 1, "%9.2f sec", 1.0*Time/int64(time.Second) );
    Abc_Print( 1, "%5.0f MB", (sat_solver2_memory_proof(p.pSat) + sat_solver2_memory(p.pSat, 0)) / (1<<20) );
    Abc_Print( 1, "%s", fUseNewLine ? "\n" : "\r" );
    fflush( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ga2_GlaGetFileName_pFileNameDef string = "glabs.aig";
func Ga2_GlaGetFileName(p *Ga2_Man_t, fAbs int32) string {
    
    if p.pPars.pFileVabs {
        return p.pPars.pFileVabs; }
    if p.pGia.pSpec {
        if fAbs != 0 {
            return Extra_FileNameGenericAppend( p.pGia.pSpec, "_abs.aig");
        } else {
            return Extra_FileNameGenericAppend( p.pGia.pSpec, "_gla.aig"); }
    }
    return spStatic_Ga2_GlaGetFileName_pFileNameDef;
}

func Ga2_GlaDumpAbsracted(p *Ga2_Man_t, fVerbose int32) {
    var pFileName string;
    assert(p.pPars.fDumpMabs || p.pPars.fDumpVabs, "p.pPars.fDumpMabs || p.pPars.fDumpVabs");
    if p.pPars.fDumpMabs {
        pFileName = Ga2_GlaGetFileName(p, 0);
        if fVerbose != 0 {
            Abc_Print( 1, "Dumping miter with abstraction map into file \"%s\"...\n", pFileName ); }
        // dump abstraction map
        Vec_IntFreeP( &p.pGia.vGateClasses );
        p.pGia.vGateClasses = Ga2_ManAbsTranslate( p );
        Gia_AigerWrite( p.pGia, pFileName, 0, 0, 0 );
    
    } else if p.pPars.fDumpVabs {
        var vGateClasses *Vec_Int_t;
        var pAbs *Gia_Man_t;
        pFileName = Ga2_GlaGetFileName(p, 1);
        if fVerbose != 0 {
            Abc_Print( 1, "Dumping abstracted model into file \"%s\"...\n", pFileName ); }
        // dump absracted model
        vGateClasses = Ga2_ManAbsTranslate( p );
        pAbs = Gia_ManDupAbsGates( p.pGia, vGateClasses );
        Gia_ManCleanValue( p.pGia );
        Gia_AigerWrite( pAbs, pFileName, 0, 0, 0 );
        Gia_ManStop( pAbs );
        Vec_IntFreeP( &vGateClasses );
    
    } else { assert(false, "0"); }
}

/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Ga2SendAbsracted(p *Ga2_Man_t, fVerbose int32) {
    var pAbs *Gia_Man_t;
    var vGateClasses *Vec_Int_t;
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Sending abstracted model...\n" );
    // create abstraction (value of p->pGia is not used here)
    vGateClasses = Ga2_ManAbsTranslate( p );
    pAbs = Gia_ManDupAbsGates( p.pGia, vGateClasses );
    Vec_IntFreeP( &vGateClasses );
    Gia_ManCleanValue( p.pGia );
    // send it out
    Gia_ManToBridgeAbsNetlist( os.Stdout, pAbs, BRIDGE_ABS_NETLIST );
    Gia_ManStop( pAbs );
}
func Gia_Ga2SendCancel(p *Ga2_Man_t, fVerbose int32) {
    extern int Gia_ManToBridgeBadAbs( FILE * pFile );
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Cancelling previously sent model...\n" );
    Gia_ManToBridgeBadAbs( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Performs gate-level abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformGla(pAig *Gia_Man_t, pPars *Abs_Par_t) int32 {
    var fUseSecondCore int32  = 1;
    var p *Ga2_Man_t;
    var vCore, vPPis *Vec_Int_t;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    var Status int32 = l_Undef;
    var RetValue int32 = -1;
    var iFrameTryToProve int32 = -1;
    var fOneIsSent int32;
    var i, c, f, Lit int32;
    pPars.iFrame = -1;
    // check trivial case 
    assert(Gia_ManPoNum(pAig) == 1, "Gia_ManPoNum(pAig) == 1"); 
    ABC_FREE( pAig.pCexSeq );
    if Gia_ObjIsConst0(Gia_ObjFanin0(Gia_ManPo(pAig,0))) {
        if !Gia_ObjFaninC0(Gia_ManPo(pAig,0)) {
            Abc_Print( 1, "Sequential miter is trivially UNSAT.\n" );
            return 1;
        }
        pAig.pCexSeq = Abc_CexMakeTriv( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), 1, 0 );
        Abc_Print( 1, "Sequential miter is trivially SAT.\n" );
        return 0;
    }
    // create gate classes if not given
    if pAig.vGateClasses == nil {
        pAig.vGateClasses = Vec_IntStart( Gia_ManObjNum(pAig) );
        Vec_IntWriteEntry( pAig.vGateClasses, 0, 1 );
        Vec_IntWriteEntry( pAig.vGateClasses, Gia_ObjFaninId0p(pAig, Gia_ManPo(pAig, 0)), 1 );
    }
    // start the manager
    p = Ga2_ManStart( pAig, pPars );
    p.timeInit = Abc_Clock() - clk;
    // perform initial abstraction
    if p.pPars.fVerbose {
        Abc_Print( 1, "Running gate-level abstraction (GLA) with the following parameters:\n" );
        Abc_Print( 1, "FrameMax = %d  ConfMax = %d  Timeout = %d  Limit = %d %%  Limit2 = %d %%  RatioMax = %d %%\n", 
            pPars.nFramesMax, pPars.nConfLimit, pPars.nTimeOut, pPars.nRatioMin, pPars.nRatioMin2, pPars.nRatioMax );
        Abc_Print( 1, "LrnStart = %d  LrnDelta = %d  LrnRatio = %d %%  Skip = %d  SimpleCNF = %d  Dump = %d\n", 
            pPars.nLearnedStart, pPars.nLearnedDelta, pPars.nLearnedPerce, pPars.fUseSkip, pPars.fUseSimple, pPars.fDumpVabs|pPars.fDumpMabs );
        if pPars.fDumpVabs || pPars.fDumpMabs {
            Abc_Print( 1, "%s will be continously dumped into file \"%s\".\n", 
                pPars.fDumpVabs ? "Abstracted model" : "Miter with abstraction map",
                Ga2_GlaGetFileName(p, pPars.fDumpVabs) ); }
        if pPars.fDumpMabs {
            {
                var Command string // was [1000]byte in C

                Abc_FrameSetStatus( -1 );
                Abc_FrameSetCex( nil );
                Abc_FrameSetNFrames( -1 );
                Command = fmt.Sprintf("write_status %s", Extra_FileNameGenericAppend((char *)(p.pPars.pFileVabs ? p.pPars.pFileVabs : "glabs.aig"), ".status"));
                Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
            }
            {
                // create trivial abstraction map
                var pObj *Gia_Obj_t; 
                var pFileName string  = Ga2_GlaGetFileName(p, 0);
                var vMap *Vec_Int_t  = pAig.vGateClasses; pAig.vGateClasses = nil;
                pAig.vGateClasses = Vec_IntStart( Gia_ManObjNum(pAig) );
                Vec_IntWriteEntry( pAig.vGateClasses, 0, 1 );
                Gia_ManForEachAnd( pAig, pObj, i )
                    Vec_IntWriteEntry( pAig.vGateClasses, i, 1 );
                Gia_ManForEachRo( pAig, pObj, i )
                    Vec_IntWriteEntry( pAig.vGateClasses, Gia_ObjId(pAig, pObj), 1 );
                Gia_AigerWrite( pAig, pFileName, 0, 0, 0 );
                Vec_IntFree( pAig.vGateClasses );
                pAig.vGateClasses = vMap;
                if p.pPars.fVerbose {
                    Abc_Print( 1, "Dumping miter with abstraction map into file \"%s\"...\n", pFileName ); }
            }
        }
        Abc_Print( 1, " Frame   %%   Abs  PPI   FF   LUT   Confl  Cex   Vars   Clas   Lrns     Time        Mem\n" );
    }
    // iterate unrolling
    for  i = f = 0; !pPars.nFramesMax || f < pPars.nFramesMax; i++  {
        var nAbsOld int32;
        // remember the timeframe
        p.pPars.iFrame = -1;
        // create new SAT solver
        Ga2_ManRestart( p );
        // remember abstraction size after the last restart
        nAbsOld = Vec_IntSize(p.vAbs);
        // unroll the circuit
        for  f = 0; !pPars.nFramesMax || f < pPars.nFramesMax; f++  {
            // remember current limits
            var nConflsBeg int32  = sat_solver2_nconflicts(p.pSat);
            var nAbs int32        = Vec_IntSize(p.vAbs);
            var nValues int32     = Vec_IntSize(p.vValues);
            var nVarsOld int32;
            // remember the timeframe
            p.pPars.iFrame = f;
            // extend and clear storage
            if f == Vec_PtrSize(p.vId2Lit) {
                Vec_PtrPush( p.vId2Lit, Vec_IntAlloc(0) ); }
            Vec_IntFillExtra( Ga2_MapFrameMap(p, f), Vec_IntSize(p.vValues), -1 );
            // add static clauses to this timeframe
            Ga2_ManAddAbsClauses( p, f );
            // skip checking if skipcheck is enabled (&gla -s)
            if p.pPars.fUseSkip && f <= p.pPars.iFrameProved {
                continue; }
            // skip checking if we need to skip several starting frames (&gla -S <num>)
            if p.pPars.nFramesStart && f <= p.pPars.nFramesStart {
                continue; }
            // get the output literal
//            Lit = Ga2_ManUnroll_rec( p, Gia_ManPo(pAig,0), f );
            Lit = Ga2_ObjFindLit( p, Gia_ObjFanin0(Gia_ManPo(pAig,0)), f );
            assert(Lit >= 0, "Lit >= 0");
            Lit = Abc_LitNotCond( Lit, Gia_ObjFaninC0(Gia_ManPo(pAig,0)) );
            if Lit == 0 {
                continue; }
            assert(Lit > 1, "Lit > 1");
            // check for counter-examples
            if p.nSatVars > sat_solver2_nvars(p.pSat) {
                sat_solver2_setnvars( p.pSat, p.nSatVars ); }
            nVarsOld = p.nSatVars;
            for  c = 0; ; c++  {
                // consider the special case when the target literal is implied false
                // by implications which happened as a result of previous refinements
                // note that incremental UNSAT core cannot be computed because there is no learned clauses
                // in this case, we will assume that UNSAT core cannot reduce the problem
                if var_is_assigned(p.pSat, Abc_Lit2Var(Lit)) {
                    Prf_ManStopP( &p.pSat.pPrf2 );
                    break;
                }
                // perform SAT solving
                clk2 = Abc_Clock();
                Status = sat_solver2_solve( p.pSat, &Lit, &Lit+1, (ABC_INT64_T)pPars.nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
                if Status == l_True { // perform refinement
                
                    p.nCexes++;
                    p.timeSat += Abc_Clock() - clk2;

                    // cancel old one if it was sent
                    if Abc_FrameIsBridgeMode() && (fOneIsSent != 0) {
                        Gia_Ga2SendCancel( p, pPars.fVerbose );
                        fOneIsSent = 0;
                    }
                    if iFrameTryToProve >= 0 {
                        Gia_GlaProveCancel( pPars.fVerbose );
                        iFrameTryToProve = -1;
                    }

                    // perform refinement
                    clk2 = Abc_Clock();
                    Rnm_ManSetRefId( p.pRnm, c );
                    vPPis = Ga2_ManRefine( p );
                    p.timeCex += Abc_Clock() - clk2;
                    if vPPis == nil {
                        if pPars.fVerbose {
                            Ga2_ManAbsPrintFrame( p, f, sat_solver2_nconflicts(p.pSat)-nConflsBeg, c, Abc_Clock() - clk, 1 ); }
                        goto finish;
                    }

                    if c == 0 {
//                        Ga2_ManRefinePrintPPis( p );
                        // create bookmark to be used for rollback
                        assert(nVarsOld == p.pSat.size, "nVarsOld == p.pSat.size");
                        sat_solver2_bookmark( p.pSat );
                        // start incremental proof manager
                        assert(p.pSat.pPrf2 == nil, "p.pSat.pPrf2 == nil");
                        p.pSat.pPrf2 = Prf_ManAlloc();
                        if p.pSat.pPrf2 {
                            p.nProofIds = 0;
                            Vec_IntFill( p.vProofIds, Gia_ManObjNum(p.pGia), -1 );
                            Prf_ManRestart( p.pSat.pPrf2, p.vProofIds, sat_solver2_nlearnts(p.pSat), Vec_IntSize(vPPis) );
                        }
                    
                    } else {
                        // resize the proof logger
                        if p.pSat.pPrf2 {
                            Prf_ManGrow( p.pSat.pPrf2, p.nProofIds + Vec_IntSize(vPPis) ); }
                    }

                    Ga2_ManAddToAbs( p, vPPis );
                    Vec_IntFree( vPPis );
                    if pPars.fVerbose {
                        Ga2_ManAbsPrintFrame( p, f, sat_solver2_nconflicts(p.pSat)-nConflsBeg, c+1, Abc_Clock() - clk, 0 ); }
                    // check if the number of objects is below limit
                    if pPars.nRatioMin2 && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin2 / 100 {
                        Status = l_Undef;
                        goto finish;
                    }
                    continue;
                }
                p.timeUnsat += Abc_Clock() - clk2;
                if Status == l_Undef { // ran out of resources
                    goto finish; }
                if p.pSat.nRuntimeLimit && Abc_Clock() > p.pSat.nRuntimeLimit { // timeout
                    goto finish; }
                if c == 0 {
                    if f > p.pPars.iFrameProved {
                        p.pPars.nFramesNoChange++; }
                    break;
                }
                if f > p.pPars.iFrameProved {
                    p.pPars.nFramesNoChange = 0; }

                // derive the core
                assert(p.pSat.pPrf2 != nil, "p.pSat.pPrf2 != nil");
                vCore = (Vec_Int_t *)Sat_ProofCore( p.pSat );
                Prf_ManStopP( &p.pSat.pPrf2 );
                // update the SAT solver
                sat_solver2_rollback( p.pSat );
                // reduce abstraction
                Ga2_ManShrinkAbs( p, nAbs, nValues, nVarsOld );

                // purify UNSAT core
                if fUseSecondCore != 0 {
//                    int nOldCore = Vec_IntSize(vCore);
                    // reverse the order of objects in the core
//                    Vec_IntSort( vCore, 0 );
//                    Vec_IntPrint( vCore );
                    // create bookmark to be used for rollback
                    assert(nVarsOld == p.pSat.size, "nVarsOld == p.pSat.size");
                    sat_solver2_bookmark( p.pSat );
                    // start incremental proof manager
                    assert(p.pSat.pPrf2 == nil, "p.pSat.pPrf2 == nil");
                    p.pSat.pPrf2 = Prf_ManAlloc();
                    if p.pSat.pPrf2 {
                        p.nProofIds = 0;
                        Vec_IntFill( p.vProofIds, Gia_ManObjNum(p.pGia), -1 );
                        Prf_ManRestart( p.pSat.pPrf2, p.vProofIds, sat_solver2_nlearnts(p.pSat), Vec_IntSize(vCore) );

                        Ga2_ManAddToAbs( p, vCore );
                        Vec_IntFree( vCore );
                    }
                    // run SAT solver
                    clk2 = Abc_Clock();
                    Status = sat_solver2_solve( p.pSat, &Lit, &Lit+1, (ABC_INT64_T)pPars.nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
                    if Status == l_Undef {
                        goto finish; }
                    assert(Status == l_False, "Status == l_False");
                    p.timeUnsat += Abc_Clock() - clk2;

                    // derive the core
                    assert(p.pSat.pPrf2 != nil, "p.pSat.pPrf2 != nil");
                    vCore = (Vec_Int_t *)Sat_ProofCore( p.pSat );
                    Prf_ManStopP( &p.pSat.pPrf2 );
                    // update the SAT solver
                    sat_solver2_rollback( p.pSat );
                    // reduce abstraction
                    Ga2_ManShrinkAbs( p, nAbs, nValues, nVarsOld );
//                    printf( "\n%4d -> %4d\n", nOldCore, Vec_IntSize(vCore) );
                }

                Ga2_ManAddToAbs( p, vCore );
//                Ga2_ManRefinePrint( p, vCore );
                Vec_IntFree( vCore );
                break;
            }
            // remember the last proved frame
            if p.pPars.iFrameProved < f {
                p.pPars.iFrameProved = f; }
            // print statistics
            if pPars.fVerbose {
                Ga2_ManAbsPrintFrame( p, f, sat_solver2_nconflicts(p.pSat)-nConflsBeg, c, Abc_Clock() - clk, 1 ); }
            // check if abstraction was proved
            if Gia_GlaProveCheck( pPars.fVerbose ) {
                RetValue = 1;
                goto finish;
            }
            if c > 0 {
                if p.pPars.fVeryVerbose {
                    Abc_Print( 1, "\n" ); }
                // recompute the abstraction
                Vec_IntFreeP( &pAig.vGateClasses );
                pAig.vGateClasses = Ga2_ManAbsTranslate( p );
                // check if the number of objects is below limit
                if pPars.nRatioMin && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin / 100 {
                    Status = l_Undef;
                    goto finish;
                }
            }
            // check the number of stable frames
            if p.pPars.nFramesNoChange == p.pPars.nFramesNoChangeLim {
                // dump the model into file
                if p.pPars.fDumpVabs || p.pPars.fDumpMabs {
                    var Command string // was [1000]byte in C

                    Abc_FrameSetStatus( -1 );
                    Abc_FrameSetCex( nil );
                    Abc_FrameSetNFrames( f );
                    Command = fmt.Sprintf("write_status %s", Extra_FileNameGenericAppend((char *)(p.pPars.pFileVabs ? p.pPars.pFileVabs : "glabs.aig"), ".status"));
                    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
                    Ga2_GlaDumpAbsracted( p, pPars.fVerbose );
                }
                // call the prover
                if p.pPars.fCallProver {
                    // cancel old one if it is proving
                    if iFrameTryToProve >= 0 {
                        Gia_GlaProveCancel( pPars.fVerbose ); }
                    // prove new one
                    Gia_GlaProveAbsracted( pAig, pPars.fSimpProver, pPars.fVeryVerbose );
                    iFrameTryToProve = f;
                    p.nPdrCalls++;
                }
                // speak to the bridge
                if Abc_FrameIsBridgeMode() {
                    // cancel old one if it was sent
                    if fOneIsSent != 0 {
                        Gia_Ga2SendCancel( p, pPars.fVerbose ); }
                    // send new one 
                    Gia_Ga2SendAbsracted( p, pPars.fVerbose );
                    fOneIsSent = 1;
                }
            }
            // if abstraction grew more than a certain percentage, force a restart
            if pPars.nRatioMax == 0 {
                continue; }
            if c > 0 && (f > 20 || Vec_IntSize(p.vAbs) > 100) && Vec_IntSize(p.vAbs) - nAbsOld >= nAbsOld * pPars.nRatioMax / 100 {
                if p.pPars.fVerbose {
                Abc_Print( 1, "Forcing restart because abstraction grew from %d to %d (more than %d %%).\n", 
                    nAbsOld, Vec_IntSize(p.vAbs), pPars.nRatioMax ); }
                break;
            }
        }
    }
finish:
    Prf_ManStopP( &p.pSat.pPrf2 );
    // cancel old one if it is proving
    if iFrameTryToProve >= 0 {
        Gia_GlaProveCancel( pPars.fVerbose ); }
    // analize the results
    if !p.fUseNewLine {
        Abc_Print( 1, "\n" ); }
    if RetValue == 1 {
        Abc_Print( 1, "GLA completed %d frames and proved abstraction derived in frame %d  ", p.pPars.iFrameProved+1, iFrameTryToProve );
    } else if pAig.pCexSeq == nil {
        Vec_IntFreeP( &pAig.vGateClasses );
        pAig.vGateClasses = Ga2_ManAbsTranslate( p );
        if p.pPars.nTimeOut && Abc_Clock() >= p.pSat.nRuntimeLimit { 
            Abc_Print( 1, "GLA reached timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.nTimeOut, p.pPars.iFrameProved+1, p.pPars.nFramesNoChange );
        } else if pPars.nConfLimit && sat_solver2_nconflicts(p.pSat) >= pPars.nConfLimit {
            Abc_Print( 1, "GLA exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.nConfLimit, p.pPars.iFrameProved+1, p.pPars.nFramesNoChange );
        } else if pPars.nRatioMin2 && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin2 / 100 {
            Abc_Print( 1, "GLA found that the size of abstraction exceeds %d %% in frame %d during refinement.  ", pPars.nRatioMin2, p.pPars.iFrameProved+1 );
        } else if pPars.nRatioMin && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin / 100 {
            Abc_Print( 1, "GLA found that the size of abstraction exceeds %d %% in frame %d.  ", pPars.nRatioMin, p.pPars.iFrameProved+1 );
        } else {
            Abc_Print( 1, "GLA finished %d frames and produced a %d-stable abstraction.  ", p.pPars.iFrameProved+1, p.pPars.nFramesNoChange ); }
        p.pPars.iFrame = p.pPars.iFrameProved;
    
    } else {
        if p.pPars.fVerbose {
            Abc_Print( 1, "\n" ); }
        if !Gia_ManVerifyCex( pAig, pAig.pCexSeq, 0 ) {
            Abc_Print( 1, "    Gia_ManPerformGlaOld(): CEX verification has failed!\n" ); }
        Abc_Print( 1, "True counter-example detected in frame %d.  ", f );
        p.pPars.iFrame = f - 1;
        Vec_IntFreeP( &pAig.vGateClasses );
        RetValue = 0;
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if p.pPars.fVerbose {
        p.timeOther = (Abc_Clock() - clk) - p.timeUnsat - p.timeSat - p.timeCex - p.timeInit;
        ABC_PRTP( "Runtime: Initializing", p.timeInit,   Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Solver UNSAT", p.timeUnsat,  Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Solver SAT  ", p.timeSat,    Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Refinement  ", p.timeCex,    Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Other       ", p.timeOther,  Abc_Clock() - clk );
        ABC_PRTP( "Runtime: TOTAL       ", Abc_Clock() - clk, Abc_Clock() - clk );
        Ga2_ManReportMemory( p );
    }
//    Ga2_ManDumpStats( p->pGia, p->pPars, p->pSat, p->pPars->iFrameProved, 0 );
    Ga2_ManStop( p );
    fflush( os.Stdout );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

