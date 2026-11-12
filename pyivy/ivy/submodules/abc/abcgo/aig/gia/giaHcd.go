package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaHcd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [New choice computation package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaHcd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "giaAig.h"
//#include "aig/aig/aig.h"
//#include "opt/dar/dar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// choicing parameters
   
type Hcd_Pars_t struct {
    nWords              int32;        // the number of simulation words
    nBTLimit              int32;      // conflict limit at a node
    nSatVarMax              int32;    // the max number of SAT variables
    fSynthesis              int32;    // set to 1 to perform synthesis
    fPolarFlip              int32;    // uses polarity adjustment
    fSimulateTfo              int32;  // uses simulation of TFO classes
    fPower              int32;        // uses power-aware rewriting
    fUseGia              int32;       // uses GIA package 
    fUseCSat              int32;      // uses circuit-based solver
    fVerbose              int32;      // verbose stats
    timeSynth          int64;     // synthesis runtime
    nNodesAhead              int32;   // the lookahead in terms of nodes
    nCallsRecycle              int32; // calls to perform before recycling SAT solver
}

// extern void Gia_ComputeEquivalences( Gia_Man_t * pMiter, int nBTLimit, int fUseMiniSat, int fVerbose );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManSetDefaultParams(p *Hcd_Pars_t) {
    memset( p, 0, sizeof(Hcd_Pars_t) );
    p.nWords         =     8;  // the number of simulation words
    p.nBTLimit       =  1000;  // conflict limit at a node
    p.nSatVarMax     =  5000;  // the max number of SAT variables
    p.fSynthesis     =     1;  // derives three snapshots
    p.fPolarFlip     =     1;  // uses polarity adjustment
    p.fSimulateTfo   =     1;  // simulate TFO
    p.fPower         =     0;  // power-aware rewriting
    p.fVerbose       =     0;  // verbose stats
    p.nNodesAhead    =  1000;  // the lookahead in terms of nodes
    p.nCallsRecycle  =   100;  // calls to perform before recycling SAT solver
}


/**Function*************************************************************

  Synopsis    [Reproduces script "compress".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Aig_Man_t * Hcd_Compress( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
//alias compress2   "b -l; rw -l; rwz -l; b -l; rwz -l; b -l"
{
    var pTemp *Aig_Man_t;

    var ParsRwr Dar_RwrPar_t;
    var pParsRwr *Dar_RwrPar_t = &ParsRwr;
    var ParsRef Dar_RefPar_t;
    var pParsRef *Dar_RefPar_t = &ParsRef;

    Dar_ManDefaultRwrParams( pParsRwr );
    Dar_ManDefaultRefParams( pParsRef );

    pParsRwr.fUpdateLevel = fUpdateLevel;
    pParsRef.fUpdateLevel = fUpdateLevel;

    pParsRwr.fPower = fPower;

    pParsRwr.fVerbose = 0;//fVerbose;
    pParsRef.fVerbose = 0;//fVerbose;

//    pAig = Aig_ManDupDfs( pAig ); 
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    // balance
    if fBalance {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }
    }

    pParsRwr.fUseZeros = 1;
    pParsRef.fUseZeros = 1;
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    return pAig;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Aig_Man_t * Hcd_Compress2( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fFanout, int fPower, int fVerbose )
//alias compress2   "b -l; rw -l; rf -l; b -l; rw -l; rwz -l; b -l; rfz -l; rwz -l; b -l"
{
    var pTemp *Aig_Man_t;

    var ParsRwr Dar_RwrPar_t;
    var pParsRwr *Dar_RwrPar_t = &ParsRwr;
    var ParsRef Dar_RefPar_t;
    var pParsRef *Dar_RefPar_t = &ParsRef;

    Dar_ManDefaultRwrParams( pParsRwr );
    Dar_ManDefaultRefParams( pParsRef );

    pParsRwr.fUpdateLevel = fUpdateLevel;
    pParsRef.fUpdateLevel = fUpdateLevel;
    pParsRwr.fFanout = fFanout;
    pParsRwr.fPower = fPower;

    pParsRwr.fVerbose = 0;//fVerbose;
    pParsRef.fVerbose = 0;//fVerbose;

//    pAig = Aig_ManDupDfs( pAig ); 
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    // balance
//    if ( fBalance )
    {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }
    }
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    pParsRwr.fUseZeros = 1;
    pParsRef.fUseZeros = 1;
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    // balance
    if fBalance {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }
    }
    
    // refactor
    Dar_ManRefactor( pAig, pParsRef );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }
    
    // rewrite
    Dar_ManRewrite( pAig, pParsRwr );
    pAig = Aig_ManDupDfs( pTemp = pAig ); 
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }

    // balance
    if fBalance {
    pAig = Dar_ManBalance( pTemp = pAig, fUpdateLevel );
    Aig_ManStop( pTemp );
    if fVerbose != 0 { Aig_ManPrintStats( pAig ); }
    }
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Vec_Ptr_t * Hcd_ChoiceSynthesis( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
//alias resyn    "b; rw; rwz; b; rwz; b"
//alias resyn2   "b; rw; rf; b; rw; rwz; b; rfz; rwz; b"
{
    var vGias *Vec_Ptr_t;
    var pGia *Gia_Man_t;

    vGias = Vec_PtrAlloc( 3 );
    pGia = Gia_ManFromAig(pAig);
    Vec_PtrPush( vGias, pGia );

    pAig = Hcd_Compress( pAig, fBalance, fUpdateLevel, fPower, fVerbose );
    pGia = Gia_ManFromAig(pAig);
    Vec_PtrPush( vGias, pGia );
//Aig_ManPrintStats( pAig );

    pAig = Hcd_Compress2( pAig, fBalance, fUpdateLevel, 1, fPower, fVerbose );
    pGia = Gia_ManFromAig(pAig);
    Vec_PtrPush( vGias, pGia );
//Aig_ManPrintStats( pAig );

    Aig_ManStop( pAig );
    return vGias;
}


/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManChoiceMiter_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    Hcd_ManChoiceMiter_rec( pNew, p, Gia_ObjFanin0(pObj) );
    if Gia_ObjIsCo(pObj) {
        return pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    Hcd_ManChoiceMiter_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
} 

/**Function*************************************************************

  Synopsis    [Derives the miter of several AIGs for choice computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManChoiceMiter(vGias *Vec_Ptr_t) *Gia_Man_t {
    var pNew, pGia, pGia0 *Gia_Man_t;
    var i, k, iNode, nNodes int32;
    // make sure they have equal parameters
    assert(Vec_PtrSize(vGias) > 0, "Vec_PtrSize(vGias) > 0");
    pGia0 = (Gia_Man_t *)Vec_PtrEntry( vGias, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGias)) && (((pGia) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
        assert(Gia_ManCiNum(pGia)  == Gia_ManCiNum(pGia0), "Gia_ManCiNum(pGia)  == Gia_ManCiNum(pGia0)");
        assert(Gia_ManCoNum(pGia)  == Gia_ManCoNum(pGia0), "Gia_ManCoNum(pGia)  == Gia_ManCoNum(pGia0)");
        assert(Gia_ManRegNum(pGia) == Gia_ManRegNum(pGia0), "Gia_ManRegNum(pGia) == Gia_ManRegNum(pGia0)");
        Gia_ManFillValue( pGia );
        Gia_ManConst0(pGia).Value = 0;
    }
    // start the new manager
    pNew = Gia_ManStart( Vec_PtrSize(vGias) * Gia_ManObjNum(pGia0) );
    pNew.pName = Abc_UtilStrsav( pGia0.pName );
    pNew.pSpec = Abc_UtilStrsav( pGia0.pSpec );
    // create new CIs and assign them to the old manager CIs
    for  k = 0; k < Gia_ManCiNum(pGia0); k++  {
        iNode = Gia_ManAppendCi(pNew);
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vGias)) && (((pGia) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
            Gia_ManCi( pGia, k ).Value = iNode;
        } 
    }
    // create internal nodes
    Gia_ManHashAlloc( pNew );
    for  k = 0; k < Gia_ManCoNum(pGia0); k++  {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vGias)) && (((pGia) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
            Hcd_ManChoiceMiter_rec( pNew, pGia, Gia_ManCo( pGia, k ) );
        }
    }
    Gia_ManHashStop( pNew );
    // check the presence of dangling nodes
    nNodes = Gia_ManHasDangling( pNew );
    assert(nNodes == 0, "nNodes == 0");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNode.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ObjCheckTfi_rec(p *Gia_Man_t, pOld *Gia_Obj_t, pNode *Gia_Obj_t, vVisited *Vec_Ptr_t) int32 {
    // check the trivial cases
    if pNode == nil {
        return 0; }
    if Gia_ObjIsCi(pNode) {
        return 0; }
//    if ( pNode->Id < pOld->Id ) // cannot use because of choices of pNode
//        return 0;
    if pNode == pOld {
        return 1; }
    // skip the visited node
    if pNode.fMark0 {
        return 0; }
    pNode.fMark0 = 1;
    Vec_PtrPush( vVisited, pNode );
    // check the children
    if Hcd_ObjCheckTfi_rec( p, pOld, Gia_ObjFanin0(pNode), vVisited ) != 0 {
        return 1; }
    if Hcd_ObjCheckTfi_rec( p, pOld, Gia_ObjFanin1(pNode), vVisited ) != 0 {
        return 1; }
    // check equivalent nodes
    return Hcd_ObjCheckTfi_rec( p, pOld, Gia_ObjNextObj(p, Gia_ObjId(p, pNode)), vVisited );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNode.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ObjCheckTfi(p *Gia_Man_t, pOld *Gia_Obj_t, pNode *Gia_Obj_t) int32 {
    var vVisited *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var RetValue, i int32;
    assert(!Gia_IsComplement(pOld), "!Gia_IsComplement(pOld)");
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    vVisited = Vec_PtrAlloc( 100 );
    RetValue = Hcd_ObjCheckTfi_rec( p, pOld, pNode, vVisited );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vVisited)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vVisited, i)), 1); i++  {
        pObj.fMark0 = 0;
    }
    Vec_PtrFree( vVisited );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Adds the next entry while making choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManAddNextEntry_rec(p *Gia_Man_t, pOld *Gia_Obj_t, pNode *Gia_Obj_t) {
    if Gia_ObjNext(p, Gia_ObjId(p, pOld)) == 0 {
        Gia_ObjSetNext( p, Gia_ObjId(p, pOld), Gia_ObjId(p, pNode) );
        return;
    }
    Hcd_ManAddNextEntry_rec( p, Gia_ObjNextObj(p, Gia_ObjId(p, pOld)), pNode );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManEquivToChoices_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pRepr, pReprNew, pObjNew *Gia_Obj_t;
    if ~pObj.Value {
        return; }
    if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
        if Gia_ObjIsConst0(pRepr) {
            pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
            return;
        }
        Hcd_ManEquivToChoices_rec( pNew, p, pRepr );
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        Hcd_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
        Hcd_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin1(pObj) );
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        if Abc_LitRegular(pObj.Value) == Abc_LitRegular(pRepr.Value) {
            assert(int32(pObj.Value) == Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) ), "int32(pObj.Value) == Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) )");
            return;
        }
        if pRepr.Value > pObj.Value { // should never happen with high resource limit
            return; }
        assert(pRepr.Value < pObj.Value, "pRepr.Value < pObj.Value");
        pReprNew = Gia_ManObj( pNew, Abc_Lit2Var(pRepr.Value) );
        pObjNew  = Gia_ManObj( pNew, Abc_Lit2Var(pObj.Value) );
        if Gia_ObjReprObj( pNew, Gia_ObjId(pNew, pObjNew) ) {
            assert(Gia_ObjReprObj( pNew, Gia_ObjId(pNew, pObjNew) ) == pReprNew, "Gia_ObjReprObj( pNew, Gia_ObjId(pNew, pObjNew) ) == pReprNew");
            pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
            return;
        }
        if !(Hcd_ObjCheckTfi( pNew, pReprNew, pObjNew ) != 0) {
            assert(Gia_ObjNext(pNew, Gia_ObjId(pNew, pObjNew)) == 0, "Gia_ObjNext(pNew, Gia_ObjId(pNew, pObjNew)) == 0");
            Gia_ObjSetRepr( pNew, Gia_ObjId(pNew, pObjNew), Gia_ObjId(pNew, pReprNew) );
            Hcd_ManAddNextEntry_rec( pNew, pReprNew, pObjNew ); 
        }
        pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Hcd_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Hcd_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Removes choices, which contain fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManRemoveBadChoices(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i, iObj, iPrev, Counter int32;
    // mark nodes with fanout
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark0 = 0;
        if Gia_ObjIsAnd(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
        
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1; }
    }
    // go through the classes and remove 
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        for  iPrev = i, iObj = Gia_ObjNext(p, i); iObj; iObj = Gia_ObjNext(p, iPrev)  {
            if !Gia_ManObj(p, iObj).fMark0 {
                iPrev = iObj; 
                continue;
            }
            Gia_ObjSetRepr( p, iObj, GIA_VOID );
            Gia_ObjSetNext( p, iPrev, Gia_ObjNext(p, iObj) );
            Gia_ObjSetNext( p, iObj, 0 );
            Counter++;
        } 
    }
    }
    // remove the marks
    Gia_ManCleanMark0( p );
//    printf( "Removed %d bad choices.\n", Counter );
}

/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManEquivToChoices(p *Gia_Man_t, nSnapshots int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var i int32;
    assert((Gia_ManCoNum(p) % nSnapshots) == 0, "(Gia_ManCoNum(p) % nSnapshots) == 0");
    Gia_ManSetPhase( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_CALLOC
    pNew.pNexts = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        Gia_ObjSetRepr( pNew, i, GIA_VOID );
    }
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
            assert(Gia_ObjIsConst0(pRepr) || Gia_ObjIsRo(p, pRepr), "Gia_ObjIsConst0(pRepr) || Gia_ObjIsRo(p, pRepr)");
            pObj.Value = pRepr.Value;
        }
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Hcd_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i % nSnapshots == 0 {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Hcd_ManRemoveBadChoices( pNew );
//    Gia_ManEquivPrintClasses( pNew, 0, 0 );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
//    Gia_ManEquivPrintClasses( pNew, 0, 0 );
    return pNew;
} 

/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ComputeChoices(pAig *Aig_Man_t, nBTLimit int32, fSynthesis int32, fUseMiniSat int32, fVerbose int32) *Aig_Man_t {
    var vGias *Vec_Ptr_t;
    var pGia, pMiter *Gia_Man_t;
    var pAigNew *Aig_Man_t;
    var i int32;
    var clk int64  = time.Now().UnixNano();
    // perform synthesis
    if fSynthesis != 0 {
        vGias = Hcd_ChoiceSynthesis( Aig_ManDupDfs(pAig), 1, 1, 0, 0 );
        if fVerbose != 0 {
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Synthesis time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));; }
        // create choices
        clk = time.Now().UnixNano();
        pMiter = Hcd_ManChoiceMiter( vGias );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vGias)) && (((pGia) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
            Gia_ManStop( pGia );
        }

        Gia_AigerWrite( pMiter, "m3.aig", 0, 0, 0 );
    
    } else {
        vGias = Vec_PtrStart( 3 );
        pMiter = Gia_ManFromAig(pAig);
    }
    // perform choicing
    Gia_ComputeEquivalences( pMiter, nBTLimit, fUseMiniSat, fVerbose );
    // derive AIG with choices
    pGia = Hcd_ManEquivToChoices( pMiter, Vec_PtrSize(vGias) );
    Gia_ManSetRegNum( pGia, Aig_ManRegNum(pAig) );
    Gia_ManStop( pMiter );
    Vec_PtrFree( vGias );
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Choicing time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));; }
//    Gia_ManHasChoices_very_old( pGia );
    // transform back
    pAigNew = Gia_ManToAig( pGia, 1 );
    Gia_ManStop( pGia );

    if fVerbose != 0 {
        extern int Dch_DeriveChoiceCountReprs( Aig_Man_t * pAig );
        extern int Dch_DeriveChoiceCountEquivs( Aig_Man_t * pAig );
        fmt.Printf("Choices   : Reprs = %5d. Equivs = %5d. Choices = %5d.\n", Dch_DeriveChoiceCountReprs( pAigNew ), Dch_DeriveChoiceCountEquivs( pAigNew ), Aig_ManChoiceNum( pAigNew ));
    }

    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ComputeChoicesTest(pGia *Gia_Man_t, nBTLimit int32, fSynthesis int32, fUseMiniSat int32, fVerbose int32) {
    var pAig, pAigNew *Aig_Man_t;
    pAig = Gia_ManToAig( pGia, 0 );
    pAigNew = Hcd_ComputeChoices( pAig, nBTLimit, fSynthesis, fUseMiniSat, fVerbose );
    Aig_ManStop( pAigNew );
    Aig_ManStop( pAig );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

