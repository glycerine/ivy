package bmc

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [bmcBmc2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Simple BMC package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcBmc2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "sat/satoko/satoko.h"
//#include "proof/ssw/ssw.h"
//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//#define AIG_VISITED       ((Aig_Obj_t *)(ABC_PTRUINT_T)1)

   
type Saig_Bmc_t struct {
    // parameters
    nFramesMax                   int32;     // the max number of timeframes to consider
    nNodesMax                   int32;      // the max number of nodes to add
    nConfMaxOne                   int32;    // the max number of conflicts at a target
    nConfMaxAll                   int32;    // the max number of conflicts for all targets
    fVerbose                   int32;       // enables verbose output
    // AIG managers
    pAig            *Aig_Man_t;           // the user's AIG manager
    pFrm            *Aig_Man_t;           // Frames manager
    vVisited            *Vec_Int_t;       // nodes visited in Frames
    // node mapping
    nObjs                   int32;          // the largest number of an AIG object
    vAig2Frm            *Vec_Ptr_t;       // mapping of AIG nodees into Frames nodes
    // SAT solver
    pSat           *sat_solver;           // SAT solver
    pSat2             *satoko_t;          // SAT solver
    nSatVars                   int32;       // the number of used SAT variables
    vObj2Var            *Vec_Int_t;       // mapping of frames objects in CNF variables
    nStitchVars                   int32;
    // subproblems
    vTargets            *Vec_Ptr_t;       // targets to be solved in this interval
    iFramePrev                   int32;     // previous frame  
    iFrameLast                   int32;     // last frame  
    iOutputLast                   int32;    // last output
    iFrameFail                   int32;     // failed frame 
    iOutputFail                   int32;    // failed output
}

func Saig_BmcObjFrame(p *Saig_Bmc_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { 
//    return (Aig_Obj_t *)Vec_PtrGetEntry( p->vAig2Frm, p->nObjs*i+pObj->Id );     
    var pRes *Aig_Obj_t;
    var vFrame *Vec_Int_t  = (Vec_Int_t *)Vec_PtrEntry( p.vAig2Frm, i );
    var iObjLit int32  = Vec_IntEntry( vFrame, Aig_ObjId(pObj) );
    if iObjLit == -1 {
        return nil; }
    pRes = Aig_ManObj( p.pFrm, Abc_Lit2Var(iObjLit) );
    if pRes == nil {
        Vec_IntWriteEntry( vFrame, Aig_ObjId(pObj), -1 );
    } else {
        pRes = Aig_NotCond( pRes, Abc_LitIsCompl(iObjLit) ); }
    return pRes;
}
func Saig_BmcObjSetFrame(p *Saig_Bmc_t, pObj *Aig_Obj_t, i int32, pNode *Aig_Obj_t) {
//    Vec_PtrSetEntry( p->vAig2Frm, p->nObjs*i+pObj->Id, pNode );     
    var vFrame *Vec_Int_t;
    var iObjLit int32;
    if i == Vec_PtrSize(p.vAig2Frm) {
        Vec_PtrPush( p.vAig2Frm, Vec_IntStartFull(p.nObjs) ); }
    assert(i < Vec_PtrSize(p.vAig2Frm), "i < Vec_PtrSize(p.vAig2Frm)");
    vFrame = (Vec_Int_t *)Vec_PtrEntry( p.vAig2Frm, i );
    if pNode == nil {
        iObjLit = -1;
    } else {
        iObjLit = Abc_Var2Lit( Aig_ObjId(Aig_Regular(pNode)), Aig_IsComplement(pNode) ); }
    Vec_IntWriteEntry( vFrame, Aig_ObjId(pObj), iObjLit );
}

func Saig_BmcObjChild0(p *Saig_Bmc_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); return Aig_NotCond(Saig_BmcObjFrame(p, Aig_ObjFanin0(pObj), i), Aig_ObjFaninC0(pObj));  }
func Saig_BmcObjChild1(p *Saig_Bmc_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); return Aig_NotCond(Saig_BmcObjFrame(p, Aig_ObjFanin1(pObj), i), Aig_ObjFaninC1(pObj));  }

func Saig_BmcSatNum(p *Saig_Bmc_t, pObj *Aig_Obj_t) int32 { return Vec_IntGetEntry( p.vObj2Var, pObj.Id );  }
func Saig_BmcSetSatNum(p *Saig_Bmc_t, pObj *Aig_Obj_t, Num int32) { Vec_IntSetEntry(p.vObj2Var, pObj.Id, Num);      }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

const ABS_ZER = 1
const ABS_ONE = 2
const ABS_UND = 3

func Abs_ManSimInfoNot(Value int32) int32 {
    if Value == ABS_ZER {
        return ABS_ONE; }
    if Value == ABS_ONE {
        return ABS_ZER; }
    return ABS_UND;
}

func Abs_ManSimInfoAnd(Value0 int32, Value1 int32) int32 {
    if Value0 == ABS_ZER || Value1 == ABS_ZER {
        return ABS_ZER; }
    if Value0 == ABS_ONE && Value1 == ABS_ONE {
        return ABS_ONE; }
    return ABS_UND;
}

func Abs_ManSimInfoGet(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry( vSimInfo, iFrame );
    return 3 & (pInfo[Aig_ObjId(pObj) >> 4] >> ((Aig_ObjId(pObj) & 15) << 1));
}

func Abs_ManSimInfoSet(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32, Value int32) {
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry( vSimInfo, iFrame );
    assert(Value >= ABS_ZER && Value <= ABS_UND, "Value >= ABS_ZER && Value <= ABS_UND");
    Value ^= Abs_ManSimInfoGet( vSimInfo, pObj, iFrame );
    pInfo[Aig_ObjId(pObj) >> 4] ^= (Value << ((Aig_ObjId(pObj) & 15) << 1));
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_ManExtendOneEval_rec(vSimInfo *Vec_Ptr_t, p *Aig_Man_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var Value0, Value1, Value int32;
    Value = Abs_ManSimInfoGet( vSimInfo, pObj, iFrame );
    if Value != 0 {
        return Value; }
    if Aig_ObjIsCi(pObj) {
        assert(Saig_ObjIsLo(p, pObj), "Saig_ObjIsLo(p, pObj)");
        Value = Abs_ManExtendOneEval_rec( vSimInfo, p, Saig_ObjLoToLi(p, pObj), iFrame-1 );
        Abs_ManSimInfoSet( vSimInfo, pObj, iFrame, Value );
        return Value;
    }
    Value0 = Abs_ManExtendOneEval_rec( vSimInfo, p, Aig_ObjFanin0(pObj), iFrame );
    if Aig_ObjFaninC0(pObj) {
        Value0 = Abs_ManSimInfoNot( Value0 ); }
    if Aig_ObjIsCo(pObj) {
        Abs_ManSimInfoSet( vSimInfo, pObj, iFrame, Value0 );
        return Value0;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    if Value0 == ABS_ZER {
        Value = ABS_ZER;
    } else {
        Value1 = Abs_ManExtendOneEval_rec( vSimInfo, p, Aig_ObjFanin1(pObj), iFrame );
        if Aig_ObjFaninC1(pObj) {
            Value1 = Abs_ManSimInfoNot( Value1 ); }
        Value = Abs_ManSimInfoAnd( Value0, Value1 );
    }
    Abs_ManSimInfoSet( vSimInfo, pObj, iFrame, Value );
    assert((Value != 0), "Value");
    return Value;
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation for one design.]

  Description [The returned array contains the result of ternary 
  simulation for all the frames where the output could be proved 0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_ManTernarySimulate(p *Aig_Man_t, nFramesMax int32, fVerbose int32) *Vec_Ptr_t {
    var vSimInfo *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i, f, nFramesLimit, nFrameWords int32;
    var clk abctime  = Abc_Clock();
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    // the maximum number of frames will be determined to use at most 200Mb of RAM
    nFramesLimit = 1 + (200000000 * 4)/Aig_ManObjNum(p);
    nFramesLimit = Abc_MinInt( nFramesLimit, nFramesMax );
    nFrameWords  = Abc_BitWordNum( 2 * Aig_ManObjNum(p) );
    // allocate simulation info
    vSimInfo = Vec_PtrAlloc( nFramesLimit );
    for  f = 0; f < nFramesLimit; f++  {
        // inlining of macro ABC_CALLOC
        Vec_PtrPush( vSimInfo, ((unsigned *) calloc((size_t)(nFrameWords), sizeof(unsigned))) );
        if f == 0 {
            // inlining of macro Saig_ManForEachLo
            for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
                Abs_ManSimInfoSet( vSimInfo, pObj, 0, ABS_ZER );
            }
        }
        Abs_ManSimInfoSet( vSimInfo, Aig_ManConst1(p), f, ABS_ONE );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            Abs_ManSimInfoSet( vSimInfo, pObj, f, ABS_UND );
        }
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Abs_ManExtendOneEval_rec( vSimInfo, p, pObj, f );
        }
        // check if simulation has derived at least one fail or unknown
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            if Abs_ManSimInfoGet(vSimInfo, pObj, f) != ABS_ZER {
                if fVerbose != 0 {
                    fmt.Printf("Ternary sim found non-zero output in frame %d.  Used %5.2f MB.  ", f, 0.25 * (f+1) * Aig_ManObjNum(p) / (1<<20));
                    // inlining of macro ABC_PRT
                    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
                }
                return vSimInfo;
            }
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Ternary sim proved all outputs in the first %d frames.  Used %5.2f MB.  ", nFramesLimit, 0.25 * nFramesLimit * Aig_ManObjNum(p) / (1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return vSimInfo;
}

/**Function*************************************************************

  Synopsis    [Free the array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_ManFreeAray(p *Vec_Ptr_t) {
    var pTemp any;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p)) && (((pTemp) = (void *)Vec_PtrEntry(p, i)), 1); i++  {
        // inlining of macro ABC_FREE
        ((pTemp) ? (free((char *) (pTemp)), (pTemp) = 0) : 0);
    }
    Vec_PtrFree( p );
}


/**Function*************************************************************

  Synopsis    [Create manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcManStart(pAig *Aig_Man_t, nFramesMax int32, nNodesMax int32, nConfMaxOne int32, nConfMaxAll int32, fVerbose int32, fUseSatoko int32) *Saig_Bmc_t {
    var p *Saig_Bmc_t;
    var pObj *Aig_Obj_t;
    var i, Lit int32;
//    assert( Aig_ManRegNum(pAig) > 0 );
    // inlining of macro ABC_ALLOC
    p = (Saig_Bmc_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Saig_Bmc_t))));
    memset( p, 0, sizeof(Saig_Bmc_t) );
    // set parameters
    p.nFramesMax   = nFramesMax;
    p.nNodesMax    = nNodesMax;
    p.nConfMaxOne  = nConfMaxOne;
    p.nConfMaxAll  = nConfMaxAll;
    p.fVerbose     = fVerbose;
    p.pAig         = pAig;
    p.nObjs        = Aig_ManObjNumMax(pAig);
    // create node and variable mappings
    p.vAig2Frm     = Vec_PtrAlloc( 100 );
    p.vObj2Var     = Vec_IntAlloc( 0 );
    Vec_IntFill( p.vObj2Var, p.nObjs, 0 );
    // start the AIG manager and map primary inputs
    p.pFrm         = Aig_ManStart( p.nObjs );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        Saig_BmcObjSetFrame( p, pObj, 0, Aig_ManConst0(p.pFrm) );
    } 
    // create SAT solver
    p.nSatVars     = 1;
    Lit = toLit( p.nSatVars );
    if fUseSatoko != 0 {
        var opts satoko_opts_t;
        satoko_default_opts(&opts);
        opts.conf_limit = nConfMaxOne;
        p.pSat2 = satoko_create();  
        satoko_configure(p.pSat2, &opts);
        satoko_setnvars(p.pSat2, 2000);
        satoko_add_clause( p.pSat2, &Lit, 1 );
    
    } else {
        p.pSat = sat_solver_new();
        p.pSat.nLearntStart = 10000;//p->pPars->nLearnedStart;
        p.pSat.nLearntDelta =  5000;//p->pPars->nLearnedDelta;
        p.pSat.nLearntRatio =    75;//p->pPars->nLearnedPerce;
        p.pSat.nLearntMax   = p.pSat.nLearntStart;
        sat_solver_setnvars( p.pSat, 2000 );
        sat_solver_addclause( p.pSat, &Lit, &Lit + 1 );
    }
    Saig_BmcSetSatNum( p, Aig_ManConst1(p.pFrm), p.nSatVars++ );
    // other data structures
    p.vTargets     = Vec_PtrAlloc( 1000 );
    p.vVisited     = Vec_IntAlloc( 1000 );
    p.iOutputFail  = -1;
    p.iFrameFail   = -1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Delete manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcManStop(p *Saig_Bmc_t) {
    Aig_ManStop( p.pFrm  );
    Vec_VecFree( (Vec_Vec_t *)p.vAig2Frm );
    Vec_IntFree( p.vObj2Var );
    if p.pSat {  sat_solver_delete( p.pSat ); }
    if p.pSat2 { satoko_destroy( p.pSat2 ); }
    Vec_PtrFree( p.vTargets );
    Vec_IntFree( p.vVisited );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Explores the possibility of constructing the output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
Aig_Obj_t * Saig_BmcIntervalExplore_rec( Saig_Bmc_t * p, Aig_Obj_t * pObj, int i )
{
    Aig_Obj_t * pRes, * p0, * p1, * pConst1 = Aig_ManConst1(p->pAig);
    pRes = Saig_BmcObjFrame( p, pObj, i );
    if ( pRes != NULL )
        return pRes;
    if ( Saig_ObjIsPi( p->pAig, pObj ) )
        pRes = AIG_VISITED;
    else if ( Saig_ObjIsLo( p->pAig, pObj ) )
        pRes = Saig_BmcIntervalExplore_rec( p, Saig_ObjLoToLi(p->pAig, pObj), i-1 );
    else if ( Aig_ObjIsCo( pObj ) )
    {
        pRes = Saig_BmcIntervalExplore_rec( p, Aig_ObjFanin0(pObj), i );
        if ( pRes != AIG_VISITED )
            pRes = Saig_BmcObjChild0( p, pObj, i );
    }
    else 
    {
        p0 = Saig_BmcIntervalExplore_rec( p, Aig_ObjFanin0(pObj), i );
        if ( p0 != AIG_VISITED )
            p0 = Saig_BmcObjChild0( p, pObj, i );
        p1 = Saig_BmcIntervalExplore_rec( p, Aig_ObjFanin1(pObj), i );
        if ( p1 != AIG_VISITED )
            p1 = Saig_BmcObjChild1( p, pObj, i );

        if ( p0 == Aig_Not(p1) )
            pRes = Aig_ManConst0(p->pFrm);
        else if ( Aig_Regular(p0) == pConst1 )
            pRes = (p0 == pConst1) ? p1 : Aig_ManConst0(p->pFrm);
        else if ( Aig_Regular(p1) == pConst1 )
            pRes = (p1 == pConst1) ? p0 : Aig_ManConst0(p->pFrm);
        else 
            pRes = AIG_VISITED;

        if ( pRes != AIG_VISITED && !Aig_ObjIsConst1(Aig_Regular(pRes)) )
            pRes = AIG_VISITED;
    }
    Saig_BmcObjSetFrame( p, pObj, i, pRes );
    return pRes;
}
*/

/**Function*************************************************************

  Synopsis    [Performs the actual construction of the output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcIntervalConstruct_rec(p *Saig_Bmc_t, pObj *Aig_Obj_t, i int32, vVisited *Vec_Int_t) *Aig_Obj_t {
    var pRes *Aig_Obj_t;
    pRes = Saig_BmcObjFrame( p, pObj, i );
    if pRes != nil {
        return pRes; }
    if Saig_ObjIsPi( p.pAig, pObj ) {
        pRes = Aig_ObjCreateCi(p.pFrm);
    } else if Saig_ObjIsLo( p.pAig, pObj ) {
        pRes = Saig_BmcIntervalConstruct_rec( p, Saig_ObjLoToLi(p.pAig, pObj), i-1, vVisited );
    } else if Aig_ObjIsCo( pObj ) {
        Saig_BmcIntervalConstruct_rec( p, Aig_ObjFanin0(pObj), i, vVisited );
        pRes = Saig_BmcObjChild0( p, pObj, i );
    
    } else {
        Saig_BmcIntervalConstruct_rec( p, Aig_ObjFanin0(pObj), i, vVisited );
        if Saig_BmcObjChild0(p, pObj, i) == Aig_ManConst0(p.pFrm) {
            pRes = Aig_ManConst0(p.pFrm);
        } else {
            Saig_BmcIntervalConstruct_rec( p, Aig_ObjFanin1(pObj), i, vVisited );
            pRes = Aig_And( p.pFrm, Saig_BmcObjChild0(p, pObj, i), Saig_BmcObjChild1(p, pObj, i) );
        }
    }
    assert(pRes != nil, "pRes != nil");
    Saig_BmcObjSetFrame( p, pObj, i, pRes );
    Vec_IntPush( vVisited, Aig_ObjId(pObj) );
    Vec_IntPush( vVisited, i );
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Adds new AIG nodes to the frames.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcInterval(p *Saig_Bmc_t) {
    var pTarget *Aig_Obj_t;
    var i, iObj, iFrame int32;
    var nNodes int32  = Aig_ManObjNum( p.pFrm );
    Vec_PtrClear( p.vTargets );
    p.iFramePrev = p.iFrameLast;
    for  ; p.iFrameLast < p.nFramesMax; p.iFrameLast++, p.iOutputLast = 0  { 
        if p.iOutputLast == 0 {
            Saig_BmcObjSetFrame( p, Aig_ManConst1(p.pAig), p.iFrameLast, Aig_ManConst1(p.pFrm) );
        }
        for  ; p.iOutputLast < Saig_ManPoNum(p.pAig); p.iOutputLast++  {
            if Aig_ManObjNum(p.pFrm) >= nNodes + p.nNodesMax {
                return; }
//            Saig_BmcIntervalExplore_rec( p, Aig_ManCo(p->pAig, p->iOutputLast), p->iFrameLast );
            Vec_IntClear( p.vVisited );
            pTarget = Saig_BmcIntervalConstruct_rec( p, Aig_ManCo(p.pAig, p.iOutputLast), p.iFrameLast, p.vVisited );
            Vec_PtrPush( p.vTargets, pTarget );
            Aig_ObjCreateCo( p.pFrm, pTarget );
            Aig_ManCleanup( p.pFrm ); // it is not efficient to cleanup the whole manager!!!
            // check if the node is gone
            // inlining of macro Vec_IntForEachEntryDouble
            for  i = 0; (i+1 < Vec_IntSize(p.vVisited)) && (((iObj) = Vec_IntEntry(p.vVisited, i)), 1) && (((iFrame) = Vec_IntEntry(p.vVisited, i+1)), 1); i += 2  {
                Saig_BmcObjFrame( p, Aig_ManObj(p.pAig, iObj), iFrame );
            }
            // it is not efficient to remove nodes, which may be used later!!!
        }
    }
}

/**Function*************************************************************

  Synopsis    [Performs the actual construction of the output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcIntervalToAig_rec(p *Saig_Bmc_t, pNew *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    Vec_IntPush( p.vVisited, Aig_ObjId(pObj) );
    if (Saig_BmcSatNum(p, pObj) != 0) || Aig_ObjIsCi(pObj) {
        p.nStitchVars += !Aig_ObjIsCi(pObj);
        return (Aig_Obj_t *)(pObj.pData = Aig_ObjCreateCi(pNew));
    }
    Saig_BmcIntervalToAig_rec( p, pNew, Aig_ObjFanin0(pObj) );
    Saig_BmcIntervalToAig_rec( p, pNew, Aig_ObjFanin1(pObj) );
    assert(pObj.pData == nil, "pObj.pData == nil");
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ));
}

/**Function*************************************************************

  Synopsis    [Creates AIG of the newly added nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcIntervalToAig(p *Saig_Bmc_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrm.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrm.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(pObj.pData == nil, "pObj.pData == nil"); }
    }

    pNew = Aig_ManStart( p.nNodesMax );
    Aig_ManConst1(p.pFrm).pData = Aig_ManConst1(pNew);
    Vec_IntClear( p.vVisited );
    Vec_IntPush( p.vVisited, Aig_ObjId(Aig_ManConst1(p.pFrm)) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTargets)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vTargets, i)), 1); i++  {
//        assert( !Aig_ObjIsConst1(Aig_Regular(pObj)) );
        pObjNew = Saig_BmcIntervalToAig_rec( p, pNew, Aig_Regular(pObj) );
        assert(!Aig_IsComplement(pObjNew), "!Aig_IsComplement(pObjNew)");
        Aig_ObjCreateCo( pNew, pObjNew );
    } 
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives CNF for the newly added nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcLoadCnf(p *Saig_Bmc_t, pCnf *Cnf_Dat_t) {
    var pObj, pObjNew *Aig_Obj_t;
    var i int32;
    var Lits [2]int32;
    var VarNumOld int32;
    var VarNumNew int32;
    // inlining of macro Aig_ManForEachObjVec
    for  i = 0; i < Vec_IntSize(p.vVisited) && (((pObj) = Aig_ManObj(p.pFrm, Vec_IntEntry(p.vVisited,i))), 1); i++  {
        // get the new variable of this node
        pObjNew     = (Aig_Obj_t *)pObj.pData;
        pObj.pData = nil;
        VarNumNew   = pCnf.pVarNums[ pObjNew.Id ];
        if VarNumNew == -1 {
            continue; }
        // get the old variable of this node
        VarNumOld   = Saig_BmcSatNum( p, pObj );
        if VarNumOld == 0 {
            Saig_BmcSetSatNum( p, pObj, VarNumNew );
            continue;
        }
        // add clauses connecting existing variables
        Lits[0] = toLitCond( VarNumOld, 0 );
        Lits[1] = toLitCond( VarNumNew, 1 );
        if p.pSat2 {
            if !satoko_add_clause( p.pSat2, Lits, 2 ) {
                assert(false, "0"); }
        
        } else {
            if !sat_solver_addclause( p.pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
        }
        Lits[0] = toLitCond( VarNumOld, 1 );
        Lits[1] = toLitCond( VarNumNew, 0 );
        if p.pSat2 {
            if !satoko_add_clause( p.pSat2, Lits, 2 ) {
                assert(false, "0"); }
        
        } else {
            if !sat_solver_addclause( p.pSat, Lits, Lits+2 ) {
                assert(false, "0"); }
        }
    }
    // add CNF to the SAT solver
    if p.pSat2 {
        for  i = 0; i < pCnf.nClauses; i++  {
            if !satoko_add_clause( p.pSat2, pCnf.pClauses[i], pCnf.pClauses[i+1]-pCnf.pClauses[i] ) {
                break; }
        }
    
    } else {
        for  i = 0; i < pCnf.nClauses; i++  {
            if !sat_solver_addclause( p.pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
                break; }
        }
    }
    if i < pCnf.nClauses {
        fmt.Printf("SAT solver became UNSAT after adding clauses.\n"); }
}

/**Function*************************************************************

  Synopsis    [Solves targets with the given resource limit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcDeriveFailed(p *Saig_Bmc_t, iTargetFail int32) {
    var k int32;
    p.iOutputFail = p.iOutputLast;
    p.iFrameFail  = p.iFrameLast;
    for  k = Vec_PtrSize(p.vTargets); k > iTargetFail; k--  {
        if p.iOutputFail == 0 {
            p.iOutputFail = Saig_ManPoNum(p.pAig);
            p.iFrameFail--;
        }
        p.iOutputFail--;
    }
}

/**Function*************************************************************

  Synopsis    [Solves targets with the given resource limit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcGenerateCounterExample(p *Saig_Bmc_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj, pObjFrm *Aig_Obj_t;
    var i, f, iVarNum int32;
    // start the counter-example
    pCex = Abc_CexAlloc( Aig_ManRegNum(p.pAig), Saig_ManPiNum(p.pAig), p.iFrameFail+1 );
    pCex.iFrame = p.iFrameFail;
    pCex.iPo    = p.iOutputFail;
    // copy the bit data
    for  f = 0; f <= p.iFrameFail; f++  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pObjFrm = Saig_BmcObjFrame( p, pObj, f );
            if pObjFrm == nil {
                continue; }
            iVarNum = Saig_BmcSatNum( p, pObjFrm );
            if iVarNum == 0 {
                continue; }
            if p.pSat2 ? satoko_read_cex_varvalue(p.pSat2, iVarNum) : sat_solver_var_value(p.pSat, iVarNum) {
                Abc_InfoSetBit( pCex.pData, pCex.nRegs + Saig_ManPiNum(p.pAig) * f + i ); }
        }
    }
    // verify the counter example
    if !Saig_ManVerifyCex( p.pAig, pCex ) {
        fmt.Printf("Saig_BmcGenerateCounterExample(): Counter-example is invalid.\n");
        Abc_CexFree( pCex );
        pCex = nil;
    }
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Solves targets with the given resource limit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcSolveTargets(p *Saig_Bmc_t, nStart int32, pnOutsSolved *int32) int32 {
    var pObj *Aig_Obj_t;
    var i, k, VarNum, Lit, status, RetValue int32;
    assert(Vec_PtrSize(p.vTargets) > 0, "Vec_PtrSize(p.vTargets) > 0");
    if p.pSat && p.pSat.qtail != p.pSat.qhead {
        RetValue = sat_solver_simplify(p.pSat);
        assert(RetValue != 0, "RetValue != 0");
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTargets)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vTargets, i)), 1); i++  {
        if ((*pnOutsSolved)++ / Saig_ManPoNum(p.pAig)) < nStart {
            continue; }
        if p.nConfMaxAll && (p.pSat ? p.pSat.stats.conflicts : satoko_conflictnum(p.pSat2)) > p.nConfMaxAll {
            return l_Undef; }
        VarNum = Saig_BmcSatNum( p, Aig_Regular(pObj) );
        Lit = toLitCond( VarNum, Aig_IsComplement(pObj) );
        if p.pSat2 {
            RetValue = satoko_solve_assumptions_limit( p.pSat2, &Lit, 1, p.nConfMaxOne );
        } else {
            RetValue = sat_solver_solve( p.pSat, &Lit, &Lit + 1, (ABC_INT64_T)p.nConfMaxOne, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 ); }
        if RetValue == l_False { // unsat
        
            // add final unit clause
            Lit = lit_neg( Lit );
            if p.pSat2 {
                status = satoko_add_clause( p.pSat2, &Lit, 1 );
            } else {
                status = sat_solver_addclause( p.pSat, &Lit, &Lit + 1 ); }
            assert(status, "status");
            if p.pSat {
                // add learned units
                for  k = 0; k < veci_size(&p.pSat.unit_lits); k++  {
                    Lit = veci_begin(&p.pSat.unit_lits)[k];
                    status = sat_solver_addclause( p.pSat, &Lit, &Lit + 1 );
                    assert(status, "status");
                }
                veci_resize(&p.pSat.unit_lits, 0);
                // propagate units
                sat_solver_compress( p.pSat );
            }
            continue;
        }
        if RetValue == l_Undef { // undecided
            return l_Undef; }
        // generate counter-example
        Saig_BmcDeriveFailed( p, i );
        p.pAig.pSeqModel = Saig_BmcGenerateCounterExample( p );

        {
//            extern Vec_Int_t * Saig_ManExtendCounterExampleTest( Aig_Man_t * p, int iFirstPi, void * pCex );
//            Saig_ManExtendCounterExampleTest( p->pAig, 0, p->pAig->pSeqModel );
        }
        return l_True;
    }
    return l_False;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcAddTargetsAsPos(p *Saig_Bmc_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTargets)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vTargets, i)), 1); i++  {
        Aig_ObjCreateCo( p.pFrm, pObj );
    }
    Aig_ManPrintStats( p.pFrm );
    Aig_ManCleanup( p.pFrm );
    Aig_ManPrintStats( p.pFrm );
} 

/**Function*************************************************************

  Synopsis    [Performs BMC with the given parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_BmcPerform(pAig *Aig_Man_t, nStart int32, nFramesMax int32, nNodesMax int32, nTimeOut int32, nConfMaxOne int32, nConfMaxAll int32, fVerbose int32, fVerbOverwrite int32, piFrames *int32, fSilent int32, fUseSatoko int32) int32 {
    var p *Saig_Bmc_t;
    var pNew *Aig_Man_t;
    var pCnf *Cnf_Dat_t;
    var nOutsSolved int32;
    var Iter int32;
    var RetValue int32 = -1;
    var nTimeToStop abctime;
    if nTimeOut {
        nTimeToStop = nTimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var clkTotal abctime = Abc_Clock();
    var Status int32  = -1;
/*
    Vec_Ptr_t * vSimInfo;
    vSimInfo = Abs_ManTernarySimulate( pAig, nFramesMax, fVerbose );
    Abs_ManFreeAray( vSimInfo );
*/
    if fVerbose != 0 {
        fmt.Printf("Running \"bmc2\". AIG:  PI/PO/Reg = %d/%d/%d.  Node = %6d. Lev = %5d.\n", Saig_ManPiNum(pAig), Saig_ManPoNum(pAig), Saig_ManRegNum(pAig), Aig_ManNodeNum(pAig), Aig_ManLevelNum(pAig));
        fmt.Printf("Params: FramesMax = %d. NodesDelta = %d. ConfMaxOne = %d. ConfMaxAll = %d.\n", nFramesMax, nNodesMax, nConfMaxOne, nConfMaxAll);
    } 
    if nFramesMax != 0 {
        nFramesMax = nFramesMax;
    } else {
        nFramesMax = ABC_INFINITY;
    }
    p = Saig_BmcManStart( pAig, nFramesMax, nNodesMax, nConfMaxOne, nConfMaxAll, fVerbose, fUseSatoko );
    // set runtime limit
    if nTimeOut {
        if p.pSat2 {
            satoko_set_runtime_limit( p.pSat2, nTimeToStop );
        } else {
            sat_solver_set_runtime_limit( p.pSat, nTimeToStop ); }
    }
    for  Iter = 0; ; Iter++  {
        clk2 = Abc_Clock();
        // add new logic interval to frames
        Saig_BmcInterval( p );
//        Saig_BmcAddTargetsAsPos( p );
        if Vec_PtrSize(p.vTargets) == 0 {
            break; }
        // convert logic slice into new AIG
        pNew = Saig_BmcIntervalToAig( p );
//printf( "StitchVars = %d.\n", p->nStitchVars );
        // derive CNF for the new AIG
        pCnf = Cnf_Derive( pNew, Aig_ManCoNum(pNew) );
        Cnf_DataLift( pCnf, p.nSatVars );
        p.nSatVars += pCnf.nVars;
        // add this CNF to the solver
        Saig_BmcLoadCnf( p, pCnf );
        Cnf_DataFree( pCnf );
        Aig_ManStop( pNew );
        // solve the targets
        RetValue = Saig_BmcSolveTargets( p, nStart, &nOutsSolved );
        if fVerbose != 0 {
            fmt.Printf("%4d : F =%5d. O =%4d.  And =%8d. Var =%8d. Conf =%7d. ", Iter, p.iFrameLast, p.iOutputLast, Aig_ManNodeNum(p.pFrm), p.nSatVars, p.pSat ? int32(p.pSat.stats.conflicts) : satoko_conflictnum(p.pSat2));   
            fmt.Printf("%4.0f MB", 4.0*(p.iFrameLast+1)*p.nObjs/(1<<20));
            fmt.Printf("%9.2f sec", float32((Abc_Clock() - clkTotal))/float32((int64(time.Second))));
            fmt.Printf("\n");
            fflush( os.Stdout );
        }
        if RetValue != l_False {
            break; }
        // check the timeout
        if nTimeOut && Abc_Clock() > nTimeToStop {
            if !(fSilent != 0) {
                fmt.Printf("Reached timeout (%d seconds).\n", nTimeOut); }
            if piFrames != nil {
                *piFrames = p.iFrameLast-1; }
            Saig_BmcManStop( p );
            return Status;
        }
    }
    if RetValue == l_True {
        assert(p.iFrameFail * Saig_ManPoNum(p.pAig) + p.iOutputFail + 1 == nOutsSolved, "p.iFrameFail * Saig_ManPoNum(p.pAig) + p.iOutputFail + 1 == nOutsSolved");
        if !(fSilent != 0) {
            Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d. ", 
                p.iOutputFail, p.pAig.pName, p.iFrameFail ); }
        Status = 0;
        if piFrames != nil {
            *piFrames = p.iFrameFail - 1; }
    
    } else { // if ( RetValue == l_False || RetValue == l_Undef )
    
        if !(fSilent != 0) {
            Abc_Print( 1, "No output failed in %d frames.  ", Abc_MaxInt(p.iFramePrev-1, 0) ); }
        if piFrames != nil {
            if p.iOutputLast > 0 {
                *piFrames = p.iFramePrev - 2;
            } else {
                *piFrames = p.iFramePrev - 1; }
        }
    }
    if !(fSilent != 0) {
        if fVerbOverwrite != 0 {
            // inlining of macro ABC_PRTr
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\r", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        
        } else {
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
        if RetValue != l_True {
            if p.iFrameLast >= p.nFramesMax {
                fmt.Printf("Reached limit on the number of timeframes (%d).\n", p.nFramesMax);
            } else if p.nConfMaxAll && (p.pSat ? int32(p.pSat.stats.conflicts) : satoko_conflictnum(p.pSat2)) > p.nConfMaxAll {
                fmt.Printf("Reached global conflict limit (%d).\n", p.nConfMaxAll);
            } else if nTimeOut && Abc_Clock() > nTimeToStop {
                fmt.Printf("Reached timeout (%d seconds).\n", nTimeOut);
            } else {
                fmt.Printf("Reached local conflict limit (%d).\n", p.nConfMaxOne); }
        }
    }
    Saig_BmcManStop( p );
    fflush( os.Stdout );
    return Status;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

