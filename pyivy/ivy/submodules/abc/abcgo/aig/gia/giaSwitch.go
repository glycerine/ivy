package gia

import (
	"fmt"
	"strconv"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSwitch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Computing switching activity.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSwitch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "giaAig.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// switching estimation parameters
   
func Saig_ManComputeSwitchProbs(pAig *Aig_Man_t, nFrames int32, nPref int32, fProbOne int32) *Vec_Int_t {
    var vSwitching, vResult *Vec_Int_t;
    var p *Gia_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // translate AIG into the intermediate form (takes care of choices if present!)
    p = Gia_ManFromAigSwitch( pAig );
    // perform the computation of switching activity
    vSwitching = Gia_ManComputeSwitchProbs( p, nFrames, nPref, fProbOne );
    // transfer the computed result to the original AIG
    vResult = Vec_IntStart( Aig_ManObjNumMax(pAig) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
//        if ( Aig_ObjIsCo(pObj) )
//            printf( "%d=%f\n", i, Abc_Int2Float( Vec_IntEntry(vSwitching, Abc_Lit2Var(pObj->iData)) ) );
        Vec_IntWriteEntry( vResult, i, Vec_IntEntry(vSwitching, Abc_Lit2Var(pObj.iData)) );
    }
    }
    // delete intermediate results
    Vec_IntFree( vSwitching );
    Gia_ManStop( p );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Computes probability of switching (or of being 1).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEvaluateSwitching(p *Gia_Man_t) float32 {
    var pObj *Gia_Obj_t;
    var SwitchTotal float32  = 0.0;
    var i int32;
    assert(p.pSwitching, "p.pSwitching");
    // inlining of macro ABC_FREE
    p.Refs = nil;
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        SwitchTotal += float32(Gia_ObjRefNum(p, pObj)) * p.pSwitching[i] / 255;
    }
    return SwitchTotal;
}

/**Function*************************************************************

  Synopsis    [Computes probability of switching (or of being 1).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
float Gia_ManComputeSwitching( Gia_Man_t * p, int nFrames, int nPref, int fProbOne )
{
    Gia_Man_t * pDfs;
    Gia_Obj_t * pObj, * pObjDfs;
    Vec_Int_t * vSwitching;
    float * pSwitching, Switch, SwitchTotal = 0.0;
    int i;
    // derives the DFS ordered AIG
    if ( Gia_ManHasMapping(p) )
        Gia_ManSetRefsMapped(p);
    else
        Gia_ManCreateRefs( p );
//    pDfs = Gia_ManDupOrderDfs( p );
    pDfs = Gia_ManDup( p );
    assert( Gia_ManObjNum(pDfs) == Gia_ManObjNum(p) );
    // perform the computation of switching activity
    vSwitching = Gia_ManComputeSwitchProbs( pDfs, nFrames, nPref, fProbOne );
    // transfer the computed result to the original AIG
    ABC_FREE( p->pSwitching );
    p->pSwitching = ABC_CALLOC( unsigned char, Gia_ManObjNum(p) );
    pSwitching = (float *)vSwitching->pArray;
    Gia_ManForEachObj( p, pObj, i )
    {
        pObjDfs = Gia_ObjFromLit( pDfs, pObj->Value );
        Switch = pSwitching[ Gia_ObjId(pDfs, pObjDfs) ];
        p->pSwitching[i] = (char)((Switch >= 1.0) ? 255 : (int)((0.002 + Switch) * 255)); // 0.00196 = (1/255)/2
        if ( Gia_ObjIsCi(pObj) || (Gia_ObjIsAnd(pObj) && (!Gia_ManHasMapping(p) || Gia_ObjIsLut(p, i))) )
        {
            SwitchTotal += (float)Gia_ObjRefNum(p, pObj) * p->pSwitching[i] / 255;
//            printf( "%d = %.2f\n", i, (float)Gia_ObjRefNum(p, pObj) * p->pSwitching[i] / 255 );            
        }
    }
    Vec_IntFree( vSwitching );
    Gia_ManStop( pDfs );
    return SwitchTotal;
}
*/
/**Function*************************************************************

  Synopsis    [Determine probability of being 1 at the outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintOutputProb(p *Gia_Man_t) *Vec_Flt_t {
    var vSimData *Vec_Flt_t;
    var pDfs *Gia_Man_t  = Gia_ManDup( p );
    assert(Gia_ManObjNum(pDfs) == Gia_ManObjNum(p), "Gia_ManObjNum(pDfs) == Gia_ManObjNum(p)");
    vSimData = (Vec_Flt_t *)Gia_ManComputeSwitchProbs( pDfs, (Gia_ManRegNum(p) ? 16 : 1), 0, 1 );
    Gia_ManStop( pDfs );
    return vSimData;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

