package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcIfif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Experiment with technology mapping.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcIfif.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const IFIF_MAX_LEAVES = 6


         
type Abc_IffObj_t struct {
    Delay [IFIF_MAX_LEAVES+1]float32;     // separate delay
//    int              nLeaves;
//    int              pLeaves[IFIF_MAX_LEAVES];
}

         
type Abc_IffMan_t struct {
    pNtk       *Abc_Ntk_t;
    pPars      *Ifif_Par_t;
    // internal data
    nObjs              int32;
    pObjs    *Abc_IffObj_t;
}

func Abc_IffObj(p *Abc_IffMan_t, i int32) *Abc_IffObj_t { assert(i >= 0 && i < p.nObjs, "i >= 0 && i < p.nObjs"); return p.pObjs + i;   }
func Abc_IffDelay(p *Abc_IffMan_t, pObj *Abc_Obj_t, fDelay1 int32) float32 { return Abc_IffObj(p, Abc_ObjId(pObj)).Delay[fDelay1];   }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkIfifStart(pNtk *Abc_Ntk_t, pPars *Ifif_Par_t) *Abc_IffMan_t {
    
    // inlining of macro ABC_CALLOC
    p := &Abc_IffMan_t{};
    p.pNtk       = pNtk;
    p.pPars      = pPars;
    // internal data
    p.nObjs      = Abc_NtkObjNumMax( pNtk );
    // inlining of macro ABC_CALLOC
    p.pObjs      = ((Abc_IffObj_t *) calloc((size_t)(p.nObjs), sizeof(Abc_IffObj_t)));
    return p;
}
func Abc_NtkIfifStop(p *Abc_IffMan_t) {
    // internal data
    // inlining of macro ABC_FREE
    ((p.pObjs) ? (free((char *) (p.pObjs)), (p.pObjs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Collects fanins into ppNodes in decreasing order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjSortByDelay(p *Abc_IffMan_t, pObj *Abc_Obj_t, fDelay1 int32, ppNodes **Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i, a, k int32;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        ppNodes[k++] = pFanin;
        if Abc_ObjIsCi(pFanin) {
            continue; }
        for  a = k-1; a > 0; a--  {
            if Abc_IffDelay(p, ppNodes[a-1], fDelay1) + p.pPars.pLutDelays[a-1] < Abc_IffDelay(p, ppNodes[a], fDelay1) + p.pPars.pLutDelays[a] {
                // inlining of macro ABC_SWAP
                 var t *Abc_Obj_t  = ppNodes[a-1]; ppNodes[a-1] = ppNodes[a]; ppNodes[a] = t; }
        };
    }
/*
    for ( a = 1; a < k; a++ )
    {
        float D1 = Abc_IffDelay(p, ppNodes[a-1], fDelay1);
        float D2 = Abc_IffDelay(p, ppNodes[a], fDelay1);
        if ( Abc_ObjIsCi(ppNodes[a-1]) || Abc_ObjIsCi(ppNodes[a]) )
            continue;
        assert( Abc_IffDelay(p, ppNodes[a-1], fDelay1) + p->pPars->pLutDelays[a-1] >= Abc_IffDelay(p, ppNodes[a], fDelay1) + p->pPars->pLutDelays[a] - 0.01 );
    }
*/
}

/**Function*************************************************************

  Synopsis    [This is the delay which object has all by itself.]

  Description [This delay is stored in Delay0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjDelay0(p *Abc_IffMan_t, pObj *Abc_Obj_t) float32 {
    var i int32;
    var Delay0 float32;
    var ppNodes [6]*Abc_Obj_t;
    Abc_ObjSortByDelay( p, pObj, 1, ppNodes );
    for  i = 0; i < Abc_ObjFaninNum(pObj); i++  {
        Delay0 = Abc_MaxFloat( Delay0, Abc_IffDelay(p, ppNodes[i], 1) + p.pPars.pLutDelays[i] );
    }
    return Delay0;
}

/**Function*************************************************************

  Synopsis    [This is the delay object has in the structure.]

  Description [This delay is stored in Delay1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjDelay1(p *Abc_IffMan_t, pObj *Abc_Obj_t) float32 {
    var i, fVeryVerbose int32;
//    Abc_IffObj_t * pIfif = Abc_IffObj( p, Abc_ObjId(pObj) );
    var ppNodes [6]*Abc_Obj_t;
    var Delay1, DelayNew float32;

    if Abc_ObjFaninNum(pObj) == 0 {
        return 0; }

    // sort fanins by delay1+LutDelay
    Abc_ObjSortByDelay( p, pObj, 1, ppNodes );

    // print verbose results
    if fVeryVerbose != 0 {
        fmt.Printf("Object %d   Level %d\n", Abc_ObjId(pObj), Abc_ObjLevel(pObj));
        for  i = 0; i < Abc_ObjFaninNum(pObj); i++  {
            fmt.Printf("Fanin %d : ", i);
            fmt.Printf("D0 %3.2f  ", Abc_IffDelay(p, ppNodes[i], 0));
            fmt.Printf("D0* %3.2f     ", Abc_IffDelay(p, ppNodes[i], 0) + p.pPars.pLutDelays[i] - p.pPars.DelayWire);
            fmt.Printf("D1 %3.2f", Abc_IffDelay(p, ppNodes[i], 1) + p.pPars.pLutDelays[i]);
            fmt.Printf("\n");
        }
        fmt.Printf("\n");
    }
/*
    // for the first nDegree delays, sort them by the minimum Delay1+LutDelay and Delay0-Wire+LutDelay
    Delay1 = 0;
    pIfif->nLeaves = 0;
    for ( i = 0; i < Abc_ObjFaninNum(pObj); i++ )
    {
        if ( Abc_ObjIsNode(ppNodes[i]) && pIfif->nLeaves < p->pPars->nDegree )
        {
            DelayNew = Abc_MinFloat( Abc_IffDelay(p, ppNodes[i], 1) + p->pPars->pLutDelays[i], 
                                     Abc_IffDelay(p, ppNodes[i], 0) + p->pPars->pLutDelays[i] - p->pPars->DelayWire );
            pIfif->pLeaves[pIfif->nLeaves++] = Abc_ObjId(ppNodes[i]);
        }
        else
            DelayNew = Abc_IffDelay(p, ppNodes[i], 1) + p->pPars->pLutDelays[i];
        Delay1 = Abc_MaxFloat( Delay1, DelayNew );
    }
*/
    // for the first nDegree delays, sort them by the minimum Delay1+LutDelay and Delay0-Wire+LutDelay
    Delay1 = 0;
    for  i = 0; i < Abc_ObjFaninNum(pObj); i++  {
        if i < p.pPars.nDegree {
            DelayNew = Abc_MinFloat( Abc_IffDelay(p, ppNodes[i], 1) + p.pPars.pLutDelays[i], 
                                     Abc_IffDelay(p, ppNodes[i], 0) + p.pPars.pLutDelays[i] - p.pPars.DelayWire );
        } else {
            DelayNew = Abc_IffDelay(p, ppNodes[i], 1) + p.pPars.pLutDelays[i]; }
        Delay1 = Abc_MaxFloat( Delay1, DelayNew );
    }
    assert(Delay1 > 0, "Delay1 > 0");
    return Delay1;
}


/**Function*************************************************************

  Synopsis    [This is the delay which object has all by itself.]

  Description [This delay is stored in Delay0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjDelayDegree(p *Abc_IffMan_t, pObj *Abc_Obj_t, d int32) float32 {
    var i int32;
    var Delay0, DelayNew float32;
    var ppNodes [6]*Abc_Obj_t;
    assert(d >= 0 && d <= p.pPars.nDegree, "d >= 0 && d <= p.pPars.nDegree");
    Abc_ObjSortByDelay( p, pObj, p.pPars.nDegree, ppNodes );
    for  i = 0; i < Abc_ObjFaninNum(pObj); i++  {
        DelayNew = Abc_IffDelay(p, ppNodes[i], p.pPars.nDegree) + p.pPars.pLutDelays[i];
        if i == 0 && d > 0 {
            DelayNew = Abc_MinFloat(DelayNew, Abc_IffDelay(p, ppNodes[i], d-1) + p.pPars.pLutDelays[i] - p.pPars.DelayWire ); }
        Delay0 = Abc_MaxFloat( Delay0, DelayNew );
    }
    return Delay0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPerformIfif(pNtk *Abc_Ntk_t, pPars *Ifif_Par_t) {
    var p *Abc_IffMan_t;
    var pIffObj *Abc_IffObj_t;
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var Delay, Delay10, DegreeFinal float32;
    var i, d, Count10 int32;
    assert(pPars.pLutLib.LutMax >= 0 && pPars.pLutLib.LutMax <= IFIF_MAX_LEAVES, "pPars.pLutLib.LutMax >= 0 && pPars.pLutLib.LutMax <= IFIF_MAX_LEAVES");
    assert(pPars.nLutSize >= 0 && pPars.nLutSize <= IFIF_MAX_LEAVES, "pPars.nLutSize >= 0 && pPars.nLutSize <= IFIF_MAX_LEAVES");
    assert(pPars.nDegree >= 0 && pPars.nDegree <= IFIF_MAX_LEAVES, "pPars.nDegree >= 0 && pPars.nDegree <= IFIF_MAX_LEAVES");
    // convert to AIGs
    Abc_NtkToAig( pNtk );
    Abc_NtkLevel( pNtk );

    // print parameters
    if pPars.fVerbose {
        fmt.Printf("Running mapper into LUT structures with the following parameters:\n");
        fmt.Printf("Pin+Wire: {");
        for  i = 0; i < pPars.pLutLib.LutMax; i++  {
            fmt.Printf(" %3.2f", pPars.pLutDelays[i]);
        }
        fmt.Printf(" }  ");    
        fmt.Printf("Wire %3.2f  Degree %d  Type: %s\n", pPars.DelayWire, pPars.nDegree, pPars.fCascade? "Cascade" : "Cluster");
    }

    // start manager
    p = Abc_NtkIfifStart( pNtk, pPars );
//    printf( "Running experiment with LUT delay %d and degree %d (LUT size is %d).\n", DelayWire, nDegree, nLutSize );

    // compute the delay
    vNodes = Abc_NtkDfs( pNtk, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
        pIffObj = Abc_IffObj( p, Abc_ObjId(pObj) );

        if pPars.fCascade {
            for  d = 0; d <= pPars.nDegree; d++  {
                pIffObj.Delay[d] = Abc_ObjDelayDegree( p, pObj, d );
            }
        
        } else {
            pIffObj.Delay[0] = Abc_ObjDelay0( p, pObj );
            pIffObj.Delay[1] = Abc_ObjDelay1( p, pObj );
        }
    }

    // get final degree number
    if pPars.fCascade {
        DegreeFinal = pPars.nDegree;
    } else {
        DegreeFinal = 1; }

    if p.pPars.fVeryVerbose {
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        fmt.Printf("Node %3d : Lev =%3d   ", Abc_ObjId(pObj), Abc_ObjLevel(pObj));
        for  d = 0; d <= DegreeFinal; d++  {
            fmt.Printf("Del%d =%4.2f  ", d, Abc_IffDelay(p, pObj, d));
        }
        fmt.Printf("\n");
    } }
    Vec_PtrFree( vNodes );


    // consider delay at the outputs
    Delay = 0;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Delay = Abc_MaxFloat( Delay, Abc_IffDelay(p, Abc_ObjFanin0(pObj), DegreeFinal) );
    }
    Delay10 = 0.9 * Delay;

    // consider delay at the outputs
    Count10 = 0;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if Abc_IffDelay(p, Abc_ObjFanin0(pObj), DegreeFinal) >= Delay10 {
            Count10++; }
    }

    fmt.Printf("Critical delay %5.2f. Critical outputs %5.2f %%\n", Delay, 100.0 * Count10 / Abc_NtkCoNum(pNtk));
//    printf( "%.2f %.2f\n", Delay, 100.0 * Count10 / Abc_NtkCoNum(pNtk) );

    // derive a new network

    // stop manager
    Abc_NtkIfifStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

