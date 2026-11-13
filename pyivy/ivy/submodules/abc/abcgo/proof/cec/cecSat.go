package cec

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [cecSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Detection of structural isomorphism.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"
//#include "sat/satoko/satoko.h"
//#include "cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// sweeping manager
   
type Cec2_Par_t struct {
    nSimWords              int32;     // simulation words
    nSimRounds              int32;    // simulation rounds
    nItersMax              int32;     // max number of iterations
    nConfLimit              int32;    // SAT solver conflict limit
    fIsMiter              int32;      // this is a miter
    fUseCones              int32;     // use logic cones
    fVeryVerbose              int32;  // verbose stats
    fVerbose              int32;      // verbose stats
}

// SAT solving manager
   
type Cec2_Man_t struct {
    pPars      *Cec2_Par_t;          // parameters
    pAig       *Gia_Man_t;           // user's AIG
    pNew       *Gia_Man_t;           // internal AIG
    // SAT solving
    pSat        *satoko_t;           // SAT solver
    vFrontier       *Vec_Ptr_t;      // CNF construction
    vFanins       *Vec_Ptr_t;        // CNF construction
    vSims       *Vec_Wrd_t;          // CI simulation info
    vNodesNew       *Vec_Int_t;      // nodes
    vSatVars       *Vec_Int_t;       // nodes
    vObjSatPairs       *Vec_Int_t;   // nodes
    vCexTriples       *Vec_Int_t;    // nodes
    // statistics
    nPatterns              int32;
    nSatSat              int32;
    nSatUnsat              int32;
    nSatUndec              int32;
    timeSatSat          abctime;
    timeSatUnsat          abctime;
    timeSatUndec          abctime;
    timeSim          abctime;
    timeRefine          abctime;
    timeExtra          abctime;
    timeStart          abctime;
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets parameter defaults.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_SetDefaultParams(p *Cec2_Par_t) {
    *p = Cec2_Par_t{nSimWords: 12, nSimRounds: 4, nItersMax: 10, nConfLimit: 1000, fUseCones: 1}
}

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ObjGetCnfVar(p *Cec2_Man_t, iObj int32) int32 {
    return Gia_ObjGetCnfVar(p.pNew, iObj, p.vFrontier, p.vFanins, p.pSat)
}

/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ObjSim(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Cec2_ObjSimSetInputBit(p *Gia_Man_t, iObj int32, Bit int32) {
    var pSim *word  = Cec2_ObjSim( p, iObj );
    if Abc_InfoHasBit( (unsigned*)pSim, p.iPatsPi ) != Bit {
        Abc_InfoXorBit( (unsigned*)pSim, p.iPatsPi ); }
}
func Cec2_ObjSimRo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSimRo *word  = Cec2_ObjSim( p, iObj );
    var pSimRi *word  = Cec2_ObjSim( p, Gia_ObjRoToRiId(p, iObj) );
    for  w = 0; w < p.nSimWords; w++  {
        pSimRo[w] = pSimRi[w];
    }
}
func Cec2_ObjSimCo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSimCo *word   = Cec2_ObjSim( p, iObj );
    var pSimDri *word  = Cec2_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] = ~pSimDri[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] =  pSimDri[w];
        } }
}
func Cec2_ObjSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Cec2_ObjSim( p, iObj );
    var pSim0 *word  = Cec2_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Cec2_ObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & ~pSim1[w];
        }
    } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & pSim1[w];
        }
    } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & ~pSim1[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & pSim1[w];
        } }
}
func Cec2_ObjSimEqual(p *Gia_Man_t, iObj0 int32, iObj1 int32) int32 {
    var w int32;
    var pSim0 *word  = Cec2_ObjSim( p, iObj0 );
    var pSim1 *word  = Cec2_ObjSim( p, iObj1 );
    if (pSim0[0] & 1) == (pSim1[0] & 1) {
        for  w = 0; w < p.nSimWords; w++  {
            if pSim0[w] != pSim1[w] {
                return 0; }
        }
        return 1;
    
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            if pSim0[w] != ~pSim1[w] {
                return 0; }
        }
        return 1;
    }
}
func Cec2_ObjSimCi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Cec2_ObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = Gia_ManRandomW( 0 );
    }
    pSim[0] <<= 1;
}
func Cec2_ManSimulateCis(p *Gia_Man_t) {
    var i, Id int32;
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Cec2_ObjSimCi( p, Id );
    }
    p.iPatsPi = 0;
}
func Cec2_ManDeriveCex(p *Gia_Man_t, iOut int32, iPat int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i, Id int32;
    pCex = Abc_CexAlloc( 0, Gia_ManCiNum(p), 1 );
    pCex.iPo = iOut;
    if iPat == -1 {
        return pCex; }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        if Abc_InfoHasBit((unsigned *)Cec2_ObjSim(p, Id), iPat) {
            Abc_InfoSetBit( pCex.pData, i ); }
    }
    return pCex;
}
func Cec2_ManSimulateCos(p *Gia_Man_t) int32 {
    var i, Id int32;
    // check outputs and generate CEX if they fail
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        Cec2_ObjSimCo( p, Id );
        if Cec2_ObjSimEqual(p, Id, 0) != 0 {
            continue; }
        p.pCexSeq = Cec2_ManDeriveCex( p, i, Abc_TtFindFirstBit2(Cec2_ObjSim(p, Id), p.nSimWords) );
        return 0;
    }
    return 1;
}
func Cec2_ManSaveCis(p *Gia_Man_t) {
    var w, i, Id int32;
    assert(p.vSimsPi != nil, "p.vSimsPi != nil");
    for  w = 0; w < p.nSimWords; w++  {
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
            Vec_WrdPush( p.vSimsPi, Cec2_ObjSim(p, Id)[w] );
        }
    }
}
func Cec2_ManSimulate(p *Gia_Man_t, vTriples *Vec_Int_t, pMan *Cec2_Man_t) int32 {
    extern void Cec2_ManSimClassRefineOne( Gia_Man_t * p, int iRepr );
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; 
    var i, iRepr, iObj, Entry, Count int32;
    //Cec2_ManSaveCis( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Cec2_ObjSimAnd( p, i ); }
    }
    pMan.timeSim += Abc_Clock() - clk;
    if p.pReprs == nil {
        return 0; }
    if vTriples != nil {
        // inlining of macro Vec_IntForEachEntryTriple
        for  i = 0; (i+2 < Vec_IntSize(vTriples)) && (((iRepr) = Vec_IntEntry(vTriples, i)), 1) && (((iObj) = Vec_IntEntry(vTriples, i+1)), 1) && (((Entry) = Vec_IntEntry(vTriples, i+2)), 1); i += 3  {
            var pSim0 *word  = Cec2_ObjSim( p, iRepr );
            var pSim1 *word  = Cec2_ObjSim( p, iObj );
            var iPat int32      = Abc_Lit2Var(Entry);
            var fPhase int32    = Abc_LitIsCompl(Entry);
            if (fPhase ^ Abc_InfoHasBit((unsigned *)pSim0, iPat)) == Abc_InfoHasBit((unsigned *)pSim1, iPat) {
                Count++; }
        }
    }
    clk = Abc_Clock();
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Cec2_ManSimClassRefineOne( p, i ); }
    }
    pMan.timeRefine += Abc_Clock() - clk;
    return Count;
}
func Cec2_ManSimAlloc(p *Gia_Man_t, nWords int32) {
    Vec_WrdFreeP( &p.vSims );
    Vec_WrdFreeP( &p.vSimsPi );
    p.vSims   = Vec_WrdStart( Gia_ManObjNum(p) * nWords );
    p.vSimsPi = Vec_WrdAlloc( Gia_ManCiNum(p) * nWords * 4 ); // storage for CI patterns
    p.nSimWords = nWords;
}


/**Function*************************************************************

  Synopsis    [Computes hash key of the simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Cec2_ManSimHashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Cec2_ManSimHashKey(pSim *word, nSims int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var pSimU *uint32 = (unsigned *)pSim;
    var i int32;
    var nSimsU int32 = 2 * nSims;
    if pSimU[0] & 1 {
        for  i = 0; i < nSimsU; i++  {
            uHash ^= ~pSimU[i] * spStatic_Cec2_ManSimHashKey_s_Primes[i & 0xf];
        }
    } else {
        for  i = 0; i < nSimsU; i++  {
            uHash ^= pSimU[i] * spStatic_Cec2_ManSimHashKey_s_Primes[i & 0xf];
        } }
    return int32((uHash % nTableSize));

}

/**Function*************************************************************

  Synopsis    [Creating initial equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ManSimClassRefineOne(p *Gia_Man_t, iRepr int32) {
    var iObj int32;
    var iPrev int32 = iRepr;
    var iPrev2 int32;
    var iRepr2 int32;
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iRepr2 = Gia_ObjNext(p, iRepr); iRepr2 > 0; iRepr2 = Gia_ObjNext(p, iRepr2)  {
        if Cec2_ObjSimEqual(p, iRepr, iRepr2) != 0 {
            iPrev = iRepr2;
        } else {
            break; }
    }
    if iRepr2 <= 0 { // no refinement
        return; }
    // relink remaining nodes of the class
    // nodes that are equal to iRepr, remain in the class of iRepr
    // nodes that are not equal to iRepr, move to the class of iRepr2
    Gia_ObjSetRepr( p, iRepr2, GIA_VOID );
    iPrev2 = iRepr2;
    for  iObj = Gia_ObjNext(p, iRepr2); iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
        if Cec2_ObjSimEqual(p, iRepr, iObj) != 0 { // remains with iRepr
        
            Gia_ObjSetNext( p, iPrev, iObj );
            iPrev = iObj;
        
        } else { // moves to iRepr2
        
            Gia_ObjSetRepr( p, iObj, iRepr2 );
            Gia_ObjSetNext( p, iPrev2, iObj );
            iPrev2 = iObj;
        }
    }
    Gia_ObjSetNext( p, iPrev, -1 );
    Gia_ObjSetNext( p, iPrev2, -1 );
}
func Cec2_ManCreateClasses(p *Gia_Man_t, pMan *Cec2_Man_t) {
    var clk abctime;
    var pObj *Gia_Obj_t;
    var nWords int32  = p.nSimWords;
    var pTable *int32;
    var nTableSize int32;
    var i int32;
    var Key int32;
    // allocate representation
    p.pReprs = nil;
    p.pNexts = nil;
    // inlining of macro ABC_CALLOC
    p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_FALLOC
    p.pNexts = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // hash each node by its simulation info
    nTableSize = Abc_PrimeCudd( Gia_ManObjNum(p) );
    // inlining of macro ABC_FALLOC
    pTable = ((int *) memset(malloc(sizeof(int) * (size_t)(nTableSize)), 0xff, sizeof(int) * (size_t)(nTableSize)));
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        p.pReprs[i].iRepr = GIA_VOID;
        if Gia_ObjIsCo(pObj) {
            continue; }
        Key = Cec2_ManSimHashKey( Cec2_ObjSim(p, i), nWords, nTableSize );
        assert(Key >= 0 && Key < nTableSize, "Key >= 0 && Key < nTableSize");
        if pTable[Key] == -1 {
            pTable[Key] = i;
        } else {
            Gia_ObjSetRepr( p, i, pTable[Key] ); }
    }
    // create classes
    for  i = Gia_ManObjNum(p) - 1; i >= 0; i--  {
        var iRepr int32  = Gia_ObjRepr(p, i);
        if iRepr == GIA_VOID {
            continue; }
        Gia_ObjSetNext( p, i, Gia_ObjNext(p, iRepr) );
        Gia_ObjSetNext( p, iRepr, i );
    }
    pTable = nil;
    clk = Abc_Clock();
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Cec2_ManSimClassRefineOne( p, i ); }
    }
    pMan.timeRefine += Abc_Clock() - clk;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ManCreate(pAig *Gia_Man_t, pPars *Cec2_Par_t) *Cec2_Man_t {
    
    var pObj *Gia_Obj_t; var i int32;
    var Pars satoko_opts_t;
    //assert( Gia_ManRegNum(pAig) == 0 );
    // inlining of macro ABC_CALLOC
    p := &Cec2_Man_t{};
    memset( p, 0, sizeof(Cec2_Man_t) );
    p.timeStart    = Abc_Clock();
    p.pPars        = pPars;
    p.pAig         = pAig;
    // create new manager
    p.pNew         = Gia_ManStart( Gia_ManObjNum(pAig) );
    Gia_ManFillValue( pAig );
    Gia_ManConst0(pAig).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pAig.vCis)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
        pObj.Value = Gia_ManAppendCi( p.pNew );
    }
    Gia_ManHashAlloc( p.pNew );
    Vec_IntFill( &p.pNew.vCopies2, Gia_ManObjNum(p.pNew), -1 );
    // SAT solving
    memset( &Pars, 0, sizeof(satoko_opts_t) );
    p.pSat         = satoko_create();
    p.vFrontier    = Vec_PtrAlloc( 1000 );
    p.vFanins      = Vec_PtrAlloc( 100 );
    p.vNodesNew    = Vec_IntAlloc( 100 );
    p.vSatVars     = Vec_IntAlloc( 100 );
    p.vObjSatPairs = Vec_IntAlloc( 100 );
    p.vCexTriples  = Vec_IntAlloc( 100 );
    Pars.conf_limit = pPars.nConfLimit;
    satoko_configure(p.pSat, &Pars);
    // remember pointer to the solver in the AIG manager
    pAig.pData     = p.pSat;
    return p;
}
func Cec2_ManDestroy(p *Cec2_Man_t) {
    if p.pPars.fVerbose {
        var timeTotal abctime  = Abc_Clock() - p.timeStart;
        var timeSat abctime    = p.timeSatSat + p.timeSatUnsat + p.timeSatUndec;
        var timeOther abctime  = timeTotal - timeSat - p.timeSim - p.timeRefine - p.timeExtra;
//        Abc_Print( 1, "%d\n", p->Num );
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("SAT solving"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  sat      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  unsat    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  fail     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUndec)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUndec)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Simulation "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSim)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Refinement "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeRefine)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeRefine)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Rollback   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeExtra)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeExtra)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeOther)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeOther)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeTotal)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeTotal)))/(float64((timeTotal))) : 0.0);;
        fflush( os.Stdout );
    }

    Vec_WrdFreeP( &p.pAig.vSims );
    //Vec_WrdFreeP( &p->pAig->vSimsPi );
    Gia_ManCleanMark01( p.pAig );
    satoko_destroy( p.pSat );
    Gia_ManStopP( &p.pNew );
    Vec_PtrFreeP( &p.vFrontier );
    Vec_PtrFreeP( &p.vFanins );
    Vec_IntFreeP( &p.vNodesNew );
    Vec_IntFreeP( &p.vSatVars );
    Vec_IntFreeP( &p.vObjSatPairs );
    Vec_IntFreeP( &p.vCexTriples );
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Verify counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ManVerify_rec(p *Gia_Man_t, iObj int32, pSat *satoko_t) int32 {
    var Value0, Value1 int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 { return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return pObj.fMark1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
        return pObj.fMark1 = satoko_var_polarity(pSat, Cec2_ObjSatId(p, pObj)) == SATOKO_LIT_TRUE; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec2_ManVerify_rec( p, Gia_ObjFaninId0(pObj, iObj), pSat ) ^ Gia_ObjFaninC0(pObj);
    Value1 = Cec2_ManVerify_rec( p, Gia_ObjFaninId1(pObj, iObj), pSat ) ^ Gia_ObjFaninC1(pObj);
    return pObj.fMark1 = Value0 & Value1;
}
func Cec2_ManVerify(p *Gia_Man_t, iObj0 int32, iObj1 int32, fPhase int32, pSat *satoko_t) {
    var Value0, Value1 int32;
    Gia_ManIncrementTravId( p );
    Value0 = Cec2_ManVerify_rec( p, iObj0, pSat );
    Value1 = Cec2_ManVerify_rec( p, iObj1, pSat );
    if (Value0 ^ Value1) == fPhase {
        fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", iObj0, iObj1); }
//    else
//        printf( "CEX verification succeeded for obj %d and obj %d.\n", iObj0, iObj1 );;
}

/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ManCollect_rec(p *Cec2_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p.pNew, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p.pNew, iObj);
    pObj = Gia_ManObj( p.pNew, iObj );
    if Cec2_ObjSatId(p.pNew, pObj) >= 0 {
        Vec_IntPush( p.vNodesNew, iObj );
        Vec_IntPush( p.vSatVars, Cec2_ObjSatId(p.pNew, pObj) );
    }
    if !(iObj != 0) {
        return; }
    if Gia_ObjIsAnd(pObj) {
        Cec2_ManCollect_rec( p, Gia_ObjFaninId0(pObj, iObj) );
        Cec2_ManCollect_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    
    } else {
        assert(Cec2_ObjSatId(p.pNew, pObj) >= 0, "Cec2_ObjSatId(p.pNew, pObj) >= 0");
        Vec_IntPushTwo( p.vObjSatPairs, Gia_ManCiIdToId(p.pAig, Gia_ObjCioId(pObj)), Cec2_ObjSatId(p.pNew, pObj) ); // SAT var
    }
}
func Cec2_ManSolveTwo(p *Cec2_Man_t, iObj0 int32, iObj1 int32, fPhase int32) int32 {
    var pObj *Gia_Obj_t;
    var status, i, iVar0, iVar1 int32;
    if iObj1 < iObj0 { 
        iObj1 ^= iObj0, iObj0 ^= iObj1, iObj1 ^= iObj0; }
    assert(iObj0 < iObj1, "iObj0 < iObj1");
    assert(p.pPars.fUseCones || satoko_varnum(p.pSat) == 0, "p.pPars.fUseCones || satoko_varnum(p.pSat) == 0");
    if !(iObj0 != 0) && Cec2_ObjSatId(p.pNew, Gia_ManConst0(p.pNew)) == -1 {
        Cec2_ObjSetSatId( p.pNew, Gia_ManConst0(p.pNew), satoko_add_variable(p.pSat, 0) ); }
    iVar0 = Cec2_ObjGetCnfVar( p, iObj0 );
    iVar1 = Cec2_ObjGetCnfVar( p, iObj1 );
    // collect inputs and internal nodes
    Vec_IntClear( p.vNodesNew );
    Vec_IntClear( p.vSatVars );
    Vec_IntClear( p.vObjSatPairs );
    Gia_ManIncrementTravId( p.pNew );
    Cec2_ManCollect_rec( p, iObj0 );
    Cec2_ManCollect_rec( p, iObj1 );
    // solve direct
    if p.pPars.fUseCones {  satoko_mark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
    satoko_assump_push( p.pSat, Abc_Var2Lit(iVar0, 1) );
    satoko_assump_push( p.pSat, Abc_Var2Lit(iVar1, fPhase) );
    status = satoko_solve( p.pSat );
    satoko_assump_pop( p.pSat );
    satoko_assump_pop( p.pSat );
    if status == SATOKO_UNSAT && iObj0 > 0 {
        // solve reverse
        satoko_assump_push( p.pSat, Abc_Var2Lit(iVar0, 0) );
        satoko_assump_push( p.pSat, Abc_Var2Lit(iVar1, !fPhase) );
        status = satoko_solve( p.pSat );
        satoko_assump_pop( p.pSat );
        satoko_assump_pop( p.pSat );
    }
    if p.pPars.fUseCones {  satoko_unmark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
    //if ( status == SATOKO_SAT )
    //    Cec2_ManVerify( p->pNew, iObj0, iObj1, fPhase, p->pSat );
    if p.pPars.fUseCones {
        return status; }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vNodesNew)) && ((pObj) = Gia_ManObj(p.pNew, Vec_IntEntry(p.vNodesNew,i))); i++  {
        Cec2_ObjCleanSatId( p.pNew, pObj );
    }
    return status;
}

func Cec2_ManSweepNode(p *Cec2_Man_t, iObj int32) int32 {
    var clk abctime  = Abc_Clock();
    var i int32;
    var IdAig int32;
    var IdSat int32;
    var status int32;
    var RetValue int32 = 1;
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pAig, iObj );
    var pRepr *Gia_Obj_t  = Gia_ObjReprObj( p.pAig, iObj );
    var fCompl int32  = Abc_LitIsCompl(pObj.Value) ^ Abc_LitIsCompl(pRepr.Value) ^ pObj.fPhase ^ pRepr.fPhase;
    status = Cec2_ManSolveTwo( p, Abc_Lit2Var(pRepr.Value), Abc_Lit2Var(pObj.Value), fCompl );
    if status == SATOKO_SAT {
        p.nSatSat++;
        p.nPatterns++;
        if p.pAig.iPatsPi == 64 * p.pAig.nSimWords - 1 {
            p.pAig.iPatsPi = 1;
        } else {
            p.pAig.iPatsPi = p.pAig.iPatsPi + 1;
        }
        assert(p.pAig.iPatsPi > 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords, "p.pAig.iPatsPi > 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords");
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(p.vObjSatPairs)) && (((IdAig) = Vec_IntEntry(p.vObjSatPairs, i)), 1) && (((IdSat) = Vec_IntEntry(p.vObjSatPairs, i+1)), 1); i += 2  {
            Cec2_ObjSimSetInputBit( p.pAig, IdAig, satoko_var_polarity(p.pSat, IdSat) == SATOKO_LIT_TRUE );
        }
        p.timeSatSat += Abc_Clock() - clk;
        RetValue = 0;
    
    } else if status == SATOKO_UNSAT {
        p.nSatUnsat++;
        pObj.Value = Abc_LitNotCond( pRepr.Value, fCompl );
        Gia_ObjSetProved( p.pAig, iObj );
        p.timeSatUnsat += Abc_Clock() - clk;
        RetValue = 1;
    
    } else {
        p.nSatUndec++;
        assert(status == SATOKO_UNDEC, "status == SATOKO_UNDEC");
        Gia_ObjSetFailed( p.pAig, iObj );
        p.timeSatUndec += Abc_Clock() - clk;
        RetValue = 2;
    }
    if p.pPars.fUseCones {
        return RetValue; }
    clk = Abc_Clock();
    satoko_rollback( p.pSat );
    p.timeExtra += Abc_Clock() - clk;
    satoko_stats(p.pSat).n_conflicts = 0;
    return RetValue;
}
func Cec2_ManPrintStats(p *Gia_Man_t, pPars *Cec2_Par_t, pMan *Cec2_Man_t) {
    if !pPars.fVerbose {
        return; }
    fmt.Printf("S =%5d ", pMan ? pMan.nSatSat   : 0);
    fmt.Printf("U =%5d ", pMan ? pMan.nSatUnsat : 0);
    fmt.Printf("F =%5d ", pMan ? pMan.nSatUndec : 0);
    Gia_ManEquivPrintClasses( p, pPars.fVeryVerbose, 0 );
}
// Cec2_ManPerformSweeping runs simulation-guided SAT sweeping using the
// source parameters, retaining the source's miter result convention.
func Cec2_ManPerformSweeping(p *Gia_Man_t, pPars *Cec2_Par_t, ppNew **Gia_Man_t) int32 {
    settings := cec2SweepSettings{
        words: int(pPars.nSimWords), rounds: int(pPars.nSimRounds),
        iterations: int(pPars.nItersMax), conflictLimit: int64(pPars.nConfLimit),
        useCones: pPars.fUseCones != 0, isMiter: pPars.fIsMiter != 0, finishGraph: ppNew != nil,
    }
    reduced, _, cex := cec2SweepEquivalences(p, settings)
    if ppNew != nil && reduced != nil { *ppNew = reduced }
    if cex != nil { return 0 }
    return 1
}

func Cec2_ManSimulateTest(p *Gia_Man_t, pPars0 *Cec_ParFra_t) *Gia_Man_t {
    var pars Cec2_Par_t
    Cec2_SetDefaultParams(&pars)
    pars.nConfLimit = pPars0.nBTLimit
    pars.fUseCones = pPars0.fUseCones
    pars.fVerbose = pPars0.fVerbose
    var reduced *Gia_Man_t
    Cec2_ManPerformSweeping(p, &pars, &reduced)
    return reduced
}
