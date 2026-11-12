package cec

import "fmt"

/**CFile****************************************************************

  FileName    [cecSynth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Partitioned sequential synthesis.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSynth.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Populate sequential synthesis parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SeqSynthesisSetDefaultParams(p *Cec_ParSeq_t) {
    memset( p, 0, sizeof(Cec_ParSeq_t) );
    p.fUseLcorr    =    0;  // enables latch correspondence
    p.fUseScorr    =    0;  // enables signal correspondence
    p.nBTLimit     = 1000;  // (scorr/lcorr) conflict limit at a node
    p.nFrames      =    1;  // (scorr only) the number of timeframes
    p.nLevelMax    =   -1;  // (scorr only) the max number of levels
    p.fConsts      =    1;  // (scl only) merging constants
    p.fEquivs      =    1;  // (scl only) merging equivalences
    p.fUseMiniSat  =    0;  // enables MiniSat in lcorr/scorr
    p.nMinDomSize  =  100;  // the size of minimum clock domain
    p.fVeryVerbose =    0;  // verbose stats
    p.fVerbose     =    0;  // verbose stats
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SeqReadMinDomSize(p *Cec_ParSeq_t) int32 {
    return p.nMinDomSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SeqReadVerbose(p *Cec_ParSeq_t) int32 {
    return p.fVerbose;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRegCreatePart(p *Gia_Man_t, vPart *Vec_Int_t, pnCountPis *int32, pnCountRegs *int32, ppMapBack **int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vNodes, vRoots *Vec_Int_t;
    var i, iOut, nCountPis, nCountRegs int32;
    var pMapBack *int32;
    // collect/mark nodes/PIs in the DFS order from the roots
    Gia_ManIncrementTravId( p );
    vRoots  = Vec_IntAlloc( Vec_IntSize(vPart) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        Vec_IntPush( vRoots, Gia_ObjId(p, Gia_ManCo(p, Gia_ManPoNum(p)+iOut)) );
    }
    vNodes = Gia_ManCollectNodesCis( p, Vec_IntArray(vRoots), Vec_IntSize(vRoots) );
    Vec_IntFree( vRoots );
    // unmark register outputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        Gia_ObjSetTravIdPrevious( p, Gia_ManCi(p, Gia_ManPiNum(p)+iOut) );
    }
    // count pure PIs
    nCountPis = nCountRegs = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        nCountPis += Gia_ObjIsTravIdCurrent(p, pObj);
    }
    // count outputs of other registers
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        nCountRegs += Gia_ObjIsTravIdCurrent(p, pObj);
    } // should be !Gia_... ???
    if pnCountPis != nil {
        *pnCountPis = nCountPis; }
    if pnCountRegs != nil {
        *pnCountRegs = nCountRegs; }
    // clean old manager
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    // create the new manager
    pNew = Gia_ManStart( Vec_IntSize(vNodes) );
    // create the PIs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if Gia_ObjIsTravIdCurrent(p, pObj) {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    // add variables for the register outputs
    // create fake POs to hold the register outputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Gia_ManCi(p, Gia_ManPiNum(p)+iOut);
        pObj.Value = Gia_ManAppendCi(pNew);
        Gia_ManAppendCo( pNew, pObj.Value );
        Gia_ObjSetTravIdCurrent( p, pObj ); // added
    }
    // create the nodes
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // add real POs for the registers
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Gia_ManCo( p, Gia_ManPoNum(p)+iOut );
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Vec_IntSize(vPart) );
    // create map
    if ppMapBack {
        // inlining of macro ABC_FALLOC
        pMapBack = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(pNew))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(pNew))));
        // map constant nodes
        pMapBack[0] = 0;
        // logic cones of register outputs
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
//            pObjNew = Aig_Regular(pObj->pData);
//            pMapBack[pObjNew->Id] = pObj->Id;
            assert(Abc_Lit2Var(Gia_ObjValue(pObj)) >= 0, "Abc_Lit2Var(Gia_ObjValue(pObj)) >= 0");
            assert(Abc_Lit2Var(Gia_ObjValue(pObj)) < Gia_ManObjNum(pNew), "Abc_Lit2Var(Gia_ObjValue(pObj)) < Gia_ManObjNum(pNew)");
            pMapBack[ Abc_Lit2Var(Gia_ObjValue(pObj)) ] = Gia_ObjId(p, pObj);
        }
        // map register outputs
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
            pObj = Gia_ManCi(p, Gia_ManPiNum(p)+iOut);
//            pObjNew = pObj->pData;
//            pMapBack[pObjNew->Id] = pObj->Id;
            assert(Abc_Lit2Var(Gia_ObjValue(pObj)) >= 0, "Abc_Lit2Var(Gia_ObjValue(pObj)) >= 0");
            assert(Abc_Lit2Var(Gia_ObjValue(pObj)) < Gia_ManObjNum(pNew), "Abc_Lit2Var(Gia_ObjValue(pObj)) < Gia_ManObjNum(pNew)");
            pMapBack[ Abc_Lit2Var(Gia_ObjValue(pObj)) ] = Gia_ObjId(p, pObj);
        }
        *ppMapBack = pMapBack;
    }
    Vec_IntFree( vNodes );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Transfers the classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_TransferMappedClasses(pPart *Gia_Man_t, pMapBack *int32, pReprs *int32) int32 {
    var pObj *Gia_Obj_t;
    var i, Id1, Id2, nClasses int32;
    if pPart.pReprs == nil { 
        return 0; }
    nClasses = 0;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pPart.nObjs) && ((pObj) = Gia_ManObj(pPart, i)); i++  {
        if Gia_ObjRepr(pPart, i) == GIA_VOID {
            continue; }
        assert(i                     < Gia_ManObjNum(pPart), "i                     < Gia_ManObjNum(pPart)");
        assert(Gia_ObjRepr(pPart, i) < Gia_ManObjNum(pPart), "Gia_ObjRepr(pPart, i) < Gia_ManObjNum(pPart)");
        Id1 = pMapBack[ i ];
        Id2 = pMapBack[ Gia_ObjRepr(pPart, i) ];
        if Id1 == Id2 {
            continue; }
        if Id1 < Id2 {
            pReprs[Id2] = Id1;
        } else {
            pReprs[Id1] = Id2; }
        nClasses++;
    }
    return nClasses;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindRepr_rec(pReprs *int32, Id int32) int32 {
    if pReprs[Id] == 0 {
        return 0; }
    if pReprs[Id] == ~0 {
        return Id; }
    return Gia_ManFindRepr_rec( pReprs, pReprs[Id] );
}

/**Function*************************************************************

  Synopsis    [Normalizes equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManNormalizeEquivalences(p *Gia_Man_t, pReprs *int32) {
    var i, iRepr int32;
    assert(p.pReprs == nil, "p.pReprs == nil");
    assert(p.pNexts == nil, "p.pNexts == nil");
    // inlining of macro ABC_CALLOC
    p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        Gia_ObjSetRepr( p, i, GIA_VOID );
    }
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        if pReprs[i] == ~0 {
            continue; }
        iRepr = Gia_ManFindRepr_rec( pReprs, i );
        Gia_ObjSetRepr( p, i, iRepr );
    }
    p.pNexts = Gia_ManDeriveNexts( p );
}

/**Function*************************************************************

  Synopsis    [Partitioned sequential synthesis.]

  Description [Returns AIG annotated with equivalence classes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SequentialSynthesisPart(p *Gia_Man_t, pPars *Cec_ParSeq_t) int32 {
    var fPrintParts int32;
    var Buffer string // was [100]byte in C

    var pTemp *Gia_Man_t;
    var vParts *Vec_Ptr_t  = (Vec_Ptr_t *)p.vClockDoms;
    var vPart *Vec_Int_t;
    var pMapBack, pReprs *int32;
    var i, nCountPis, nCountRegs int32;
    var nClasses int32;
    var clk abctime  = Abc_Clock();

    // save parameters
    if fPrintParts != 0 {
        // print partitions
        Abc_Print( 1, "The following clock domains are used:\n" );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
            pTemp = Gia_ManRegCreatePart( p, vPart, &nCountPis, &nCountRegs, nil );
            Buffer = fmt.Sprintf("part%03d.aig", i);
            Gia_AigerWrite( pTemp, Buffer, 0, 0, 0 );
            Abc_Print( 1, "part%03d.aig : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d.\n", 
                i, Vec_IntSize(vPart), Gia_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Gia_ManAndNum(pTemp) );
            Gia_ManStop( pTemp );
        }
    }

    // perform sequential synthesis for clock domains
    // inlining of macro ABC_FALLOC
    pReprs = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        pTemp = Gia_ManRegCreatePart( p, vPart, &nCountPis, &nCountRegs, &pMapBack );
        if nCountPis > 0 {
            if pPars.fUseScorr {
                var CorPars Cec_ParCor_t;
                var pCorPars *Cec_ParCor_t = &CorPars;
                Cec_ManCorSetDefaultParams( pCorPars );
                pCorPars.nBTLimit   = pPars.nBTLimit;
                pCorPars.nLevelMax  = pPars.nLevelMax;
                pCorPars.fVerbose   = pPars.fVeryVerbose;
                pCorPars.fUseCSat   = 1;
                Cec_ManLSCorrespondenceClasses( pTemp, pCorPars );
            
            } else if pPars.fUseLcorr {
                var CorPars Cec_ParCor_t;
                var pCorPars *Cec_ParCor_t = &CorPars;
                Cec_ManCorSetDefaultParams( pCorPars );
                pCorPars.fLatchCorr = 1;
                pCorPars.nBTLimit   = pPars.nBTLimit;
                pCorPars.fVerbose   = pPars.fVeryVerbose;
                pCorPars.fUseCSat   = 1;
                Cec_ManLSCorrespondenceClasses( pTemp, pCorPars );
            
            } else {
//                pNew = Gia_ManSeqStructSweep( pTemp, pPars->fConsts, pPars->fEquivs, pPars->fVerbose );
//                Gia_ManStop( pNew );
                Gia_ManSeqCleanupClasses( pTemp, pPars.fConsts, pPars.fEquivs, pPars.fVerbose );
            }
//Abc_Print( 1, "Part equivalences = %d.\n", Gia_ManEquivCountLitsAll(pTemp) );
            nClasses = Gia_TransferMappedClasses( pTemp, pMapBack, pReprs );
            if pPars.fVerbose {
                Abc_Print( 1, "%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. Cl = %5d.\n", 
                    i, Vec_IntSize(vPart), Gia_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Gia_ManAndNum(pTemp), nClasses );
            }
        }
        Gia_ManStop( pTemp );
        pMapBack = nil;
    }

    // generate resulting equivalences
    Gia_ManNormalizeEquivalences( p, pReprs );
//Abc_Print( 1, "Total equivalences = %d.\n", Gia_ManEquivCountLitsAll(p) );
    pReprs = nil;
    if pPars.fVerbose {
        Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTestOnePair_rec(pSat *sat_solver, p *Gia_Man_t, iObj int32, vSatVar *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t; 
    var iVar, iVar0, iVar1 int32;
    if Vec_IntEntry(vSatVar, iObj) >= 0 {
        return Vec_IntEntry(vSatVar, iObj); }
    iVar = sat_solver_addvar(pSat);
    Vec_IntWriteEntry( vSatVar, iObj, iVar );
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsAnd(pObj) {
        iVar0 = Gia_ManTestOnePair_rec( pSat, p, Gia_ObjFaninId0(pObj, iObj), vSatVar );
        iVar1 = Gia_ManTestOnePair_rec( pSat, p, Gia_ObjFaninId1(pObj, iObj), vSatVar );
        sat_solver_add_and( pSat, iVar, iVar0, iVar1, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 );
    }
    return iVar;
}
func Gia_ManTestOnePair(p *Gia_Man_t, iObj1 int32, iObj2 int32, fPhase int32) int32 {
    var Lits [2]int32 = [2]int32{1};
    var status int32;
    var pSat *sat_solver  = sat_solver_new();
    var vSatVar *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    Vec_IntWriteEntry( vSatVar, 0, sat_solver_addvar(pSat) );
    sat_solver_addclause( pSat, Lits, Lits + 1 );
    Gia_ManTestOnePair_rec( pSat, p, iObj1, vSatVar );
    Gia_ManTestOnePair_rec( pSat, p, iObj2, vSatVar );
    Lits[0] = Abc_Var2Lit( Vec_IntEntry(vSatVar, iObj1), 1 );
    Lits[1] = Abc_Var2Lit( Vec_IntEntry(vSatVar, iObj2), fPhase );
    status = sat_solver_solve( pSat, Lits, Lits + 2, 0, 0, 0, 0 );
    if status == l_False {
        Lits[0] = Abc_LitNot( Lits[0] );
        Lits[1] = Abc_LitNot( Lits[1] );
        status = sat_solver_solve( pSat, Lits, Lits + 2, 0, 0, 0, 0 );
    }
    Vec_IntFree( vSatVar );
    sat_solver_delete( pSat );
    return status;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

