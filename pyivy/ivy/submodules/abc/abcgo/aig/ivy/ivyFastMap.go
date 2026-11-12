package ivy

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [ivyFastMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Fast FPGA mapping.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyFastMap.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const IVY_INFINITY =   10000

   
type Ivy_SuppMan_t struct {
    nLimit         int32;    // the limit on the number of inputs
    nObjs         int32;     // the number of entries
    nSize         int32;     // size of each entry in bytes
    pMem       string;      // memory allocated
    vLuts  *Vec_Vec_t;     // the array of nodes used in the mapping
}

   
type Ivy_Supp_t struct {
    nSize        byte;     // the number of support nodes
    fMark        byte;     // multipurpose mask
    fMark2        byte;    // multipurpose mask 
    fMark3        byte;    // multipurpose mask 
    nRefs         int32;     // the number of references
    Delay       int16;     // the delay of the node
    DelayR       int16;    // the reverse delay of the node
    pArray [0]int32; // the support nodes
}

func Ivy_ObjSupp(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) *Ivy_Supp_t { 
    return (Ivy_Supp_t *)(((Ivy_SuppMan_t*)pAig.pData).pMem + pObj.Id * ((Ivy_SuppMan_t*)pAig.pData).nSize); 
}
func Ivy_ObjSuppStart(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) *Ivy_Supp_t { 
    var pSupp *Ivy_Supp_t;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    pSupp.fMark = 0;
    pSupp.Delay = 0;
    pSupp.nSize = 1;
    pSupp.pArray[0] = pObj.Id;
    return pSupp;
}

// static void Ivy_FastMapPrint( Ivy_Man_t * pAig, int Delay, int Area, abctime Time, char * pStr );

// static int  Ivy_FastMapDelay( Ivy_Man_t * pAig );

// static int  Ivy_FastMapArea( Ivy_Man_t * pAig );

// static void Ivy_FastMapNode( Ivy_Man_t * pAig, Ivy_Obj_t * pObj, int nLimit );

// static void Ivy_FastMapNodeArea( Ivy_Man_t * pAig, Ivy_Obj_t * pObj, int nLimit );

// static int  Ivy_FastMapMerge( Ivy_Supp_t * pSupp0, Ivy_Supp_t * pSupp1, Ivy_Supp_t * pSupp, int nLimit );

// static void Ivy_FastMapRequired( Ivy_Man_t * pAig, int Delay, int fSetInter );

// static void Ivy_FastMapRecover( Ivy_Man_t * pAig, int nLimit );

// static int  Ivy_FastMapNodeDelay( Ivy_Man_t * pAig, Ivy_Obj_t * pObj );

// static int  Ivy_FastMapNodeAreaRefed( Ivy_Man_t * pAig, Ivy_Obj_t * pObj );

// static int  Ivy_FastMapNodeAreaDerefed( Ivy_Man_t * pAig, Ivy_Obj_t * pObj );

// static void Ivy_FastMapNodeRecover( Ivy_Man_t * pAig, Ivy_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vFrontOld );

// static int  Ivy_FastMapNodeRef( Ivy_Man_t * pAig, Ivy_Obj_t * pObj );

// static int  Ivy_FastMapNodeDeref( Ivy_Man_t * pAig, Ivy_Obj_t * pObj );



var s_MappingTime abctime;
var s_MappingMem int32;


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs fast K-LUT mapping of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapPerform(pAig *Ivy_Man_t, nLimit int32, fRecovery int32, fVerbose int32) {
    var pMan *Ivy_SuppMan_t;
    var pObj *Ivy_Obj_t;
    var i, Delay, Area int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    // start the memory for supports
    // inlining of macro ABC_ALLOC
    pMan = ((Ivy_SuppMan_t *) malloc(sizeof(Ivy_SuppMan_t) * (size_t)(1)));
    *pMan = Ivy_SuppMan_t{};
    pMan.nLimit = nLimit;
    pMan.nObjs  = Ivy_ManObjIdMax(pAig) + 1;
    pMan.nSize  = sizeof(Ivy_Supp_t) + nLimit * sizeof(int);
    // inlining of macro ABC_ALLOC
    pMan.pMem   = (char *)((char *) malloc(sizeof(char) * (size_t)(pMan.nObjs * pMan.nSize)));
    memset( pMan.pMem, 0, pMan.nObjs * pMan.nSize );
    pMan.vLuts  = Vec_VecAlloc( 100 );
    pAig.pData  = pMan;
clk = Abc_Clock();
    // set the PI mapping
    Ivy_ObjSuppStart( pAig, Ivy_ManConst1(pAig) );
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vPis, i)), 1); i++  {
        Ivy_ObjSuppStart( pAig, pObj );
    }
    // iterate through all nodes in the topological order
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        Ivy_FastMapNode( pAig, pObj, nLimit ); }
    }
    // find the best arrival time and area
    Delay = Ivy_FastMapDelay( pAig );
    Area = Ivy_FastMapArea(pAig);
    if fVerbose != 0 {
        Ivy_FastMapPrint( pAig, Delay, Area, Abc_Clock() - clk, "Delay oriented mapping: " ); }

// 2-1-2 (doing 2-1-2-1-2 improves 0.5%)

    if fRecovery != 0 {
clk = Abc_Clock();
    Ivy_FastMapRequired( pAig, Delay, 0 );
    // remap the nodes
    Ivy_FastMapRecover( pAig, nLimit );
    Delay = Ivy_FastMapDelay( pAig );
    Area = Ivy_FastMapArea(pAig);
    if fVerbose != 0 {
        Ivy_FastMapPrint( pAig, Delay, Area, Abc_Clock() - clk, "Area recovery 2       : " ); }

clk = Abc_Clock();
    Ivy_FastMapRequired( pAig, Delay, 0 );
    // iterate through all nodes in the topological order
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        Ivy_FastMapNodeArea( pAig, pObj, nLimit ); }
    }
    Delay = Ivy_FastMapDelay( pAig );
    Area = Ivy_FastMapArea(pAig);
    if fVerbose != 0 {
        Ivy_FastMapPrint( pAig, Delay, Area, Abc_Clock() - clk, "Area recovery 1       : " ); }

clk = Abc_Clock();
    Ivy_FastMapRequired( pAig, Delay, 0 );
    // remap the nodes
    Ivy_FastMapRecover( pAig, nLimit );
    Delay = Ivy_FastMapDelay( pAig );
    Area = Ivy_FastMapArea(pAig);
    if fVerbose != 0 {
        Ivy_FastMapPrint( pAig, Delay, Area, Abc_Clock() - clk, "Area recovery 2       : " ); }
    }


    s_MappingTime = Abc_Clock() - clkTotal;
    s_MappingMem = pMan.nObjs * pMan.nSize;
/*
    {
        Vec_Ptr_t * vNodes;
        vNodes = Vec_PtrAlloc( 100 );
        Vec_VecForEachEntry( Ivy_Obj_t *, pMan->vLuts, pObj, i, k )
            Vec_PtrPush( vNodes, pObj );
        Ivy_ManShow( pAig, 0, vNodes );
        Vec_PtrFree( vNodes );
    }
*/
}

/**Function*************************************************************

  Synopsis    [Cleans memory used for decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapStop(pAig *Ivy_Man_t) {
    var p *Ivy_SuppMan_t  = (Ivy_SuppMan_t *)pAig.pData;
    Vec_VecFree( p.vLuts );
    p.pMem = nil;
    p = nil;
    pAig.pData = nil;
}

/**Function*************************************************************

  Synopsis    [Prints statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapPrint(pAig *Ivy_Man_t, Delay int32, Area int32, Time abctime, pStr string) {
    fmt.Printf("%s : Delay = %3d. Area = %6d. ", pStr, Delay, Area);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Time)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Computes delay after LUT mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapDelay(pAig *Ivy_Man_t) int32 {
    var pSupp *Ivy_Supp_t;
    var pObj *Ivy_Obj_t;
    var i, DelayMax int32;
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vPos, i)), 1); i++  {
        pObj = Ivy_ObjFanin0(pObj);
        if !Ivy_ObjIsNode(pObj) {
            continue; }
        pSupp = Ivy_ObjSupp( pAig, pObj );
        if DelayMax < pSupp.Delay {
            DelayMax = pSupp.Delay; }
    }
    return DelayMax;
}

/**Function*************************************************************

  Synopsis    [Computes area after mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapArea_rec(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, vLuts *Vec_Vec_t) int32 {
    var pSupp *Ivy_Supp_t;
    var i, Counter int32;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    // skip visited nodes and PIs
    if pSupp.fMark || pSupp.nSize == 1 {
        return 0; }
    pSupp.fMark = 1;
    // compute the area of this node
    Counter = 0;
    for  i = 0; i < pSupp.nSize; i++  {
        Counter += Ivy_FastMapArea_rec( pAig, Ivy_ManObj(pAig, pSupp.pArray[i]), vLuts );
    }
    // add the node to the array of LUTs
    Vec_VecPush( vLuts, pSupp.Delay, pObj );
    return 1 + Counter;
}

/**Function*************************************************************

  Synopsis    [Computes area after mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapArea(pAig *Ivy_Man_t) int32 {
    var vLuts *Vec_Vec_t;
    var pObj *Ivy_Obj_t;
    var i, Counter int32;
    // get the array to store the nodes
    vLuts = ((Ivy_SuppMan_t *)pAig.pData).vLuts;
    Vec_VecClear( vLuts );
    // explore starting from each node
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vPos, i)), 1); i++  {
        Counter += Ivy_FastMapArea_rec( pAig, Ivy_ObjFanin0(pObj), vLuts );
    }
    // clean the marks
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        Ivy_ObjSupp( pAig, pObj ).fMark = 0; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Performs fast mapping for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjIsNodeInt1(pObj *Ivy_Obj_t) int32 {
    return Ivy_ObjIsNode(pObj) && Ivy_ObjRefs(pObj) == 1;
}

/**Function*************************************************************

  Synopsis    [Performs fast mapping for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjIsNodeInt2(pObj *Ivy_Obj_t) int32 {
    return Ivy_ObjIsNode(pObj) && Ivy_ObjRefs(pObj) <= 2;
}

/**Function*************************************************************

  Synopsis    [Performs fast mapping for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntRemoveDup(pArray *int32, nSize int32) int32 {
    var i, k int32;
    if nSize < 2 {
        return nSize; }
    for  i = k = 1; i < nSize; i++  {
        if pArray[i] != pArray[i-1] {
            pArray[k++] = pArray[i]; }
    }
    return k;
}

/**Function*************************************************************

  Synopsis    [Performs fast mapping for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeArea2(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32) {
    static int Store[32], StoreSize;
    static char Supp0[16], Supp1[16];
    static Ivy_Supp_t * pTemp0 = (Ivy_Supp_t *)Supp0;
    static Ivy_Supp_t * pTemp1 = (Ivy_Supp_t *)Supp1;
    var pFanin0, pFanin1 *Ivy_Obj_t;
    var pSupp0, pSupp1, pSupp *Ivy_Supp_t;
    var RetValue, DelayOld int32;
    assert(nLimit <= 32, "nLimit <= 32");
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    // get the fanins
    pFanin0 = Ivy_ObjFanin0(pObj);
    pFanin1 = Ivy_ObjFanin1(pObj);
    // get the supports
    pSupp0 = Ivy_ObjSupp( pAig, pFanin0 );
    pSupp1 = Ivy_ObjSupp( pAig, pFanin1 );
    pSupp  = Ivy_ObjSupp( pAig, pObj );
    assert(pSupp.fMark == 0, "pSupp.fMark == 0");
    // get the old delay of the node
    DelayOld = Ivy_FastMapNodeDelay(pAig, pObj);
    assert(DelayOld <= pSupp.DelayR, "DelayOld <= pSupp.DelayR");
    // copy the current cut
    memcpy( Store, pSupp.pArray, sizeof(int) * pSupp.nSize );
    StoreSize = pSupp.nSize;
    // get the fanin support
    if Ivy_ObjRefs(pFanin0) > 1 && pSupp0.Delay < pSupp.DelayR {
        pSupp0 = pTemp0;
        pSupp0.nSize = 1;
        pSupp0.pArray[0] = Ivy_ObjFaninId0(pObj);
    }
    // get the fanin support
    if Ivy_ObjRefs(pFanin1) > 1 && pSupp1.Delay < pSupp.DelayR {
        pSupp1 = pTemp1;
        pSupp1.nSize = 1;
        pSupp1.pArray[0] = Ivy_ObjFaninId1(pObj);
    }
    // merge the cuts
    if pSupp0.nSize < pSupp1.nSize {
        RetValue = Ivy_FastMapMerge( pSupp1, pSupp0, pSupp, nLimit );
    } else {
        RetValue = Ivy_FastMapMerge( pSupp0, pSupp1, pSupp, nLimit ); }
    if !(RetValue != 0) {
        pSupp.nSize = 2;
        pSupp.pArray[0] = Ivy_ObjFaninId0(pObj);
        pSupp.pArray[1] = Ivy_ObjFaninId1(pObj);
    }
    // check the resulting delay
    pSupp.Delay = Ivy_FastMapNodeDelay(pAig, pObj);
    if pSupp.Delay > pSupp.DelayR {
        pSupp.nSize = StoreSize;
        memcpy( pSupp.pArray, Store, sizeof(int) * pSupp.nSize );
        pSupp.Delay = DelayOld;
    }
}

/**Function*************************************************************

  Synopsis    [Performs fast mapping for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeArea(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32) {
    static int Store[32], StoreSize;
    static char Supp0[16], Supp1[16];
    static Ivy_Supp_t * pTemp0 = (Ivy_Supp_t *)Supp0;
    static Ivy_Supp_t * pTemp1 = (Ivy_Supp_t *)Supp1;
    var pFanin0, pFanin1 *Ivy_Obj_t;
    var pSupp0, pSupp1, pSupp *Ivy_Supp_t;
    var RetValue, DelayOld, RefsOld int32;
    var AreaBef, AreaAft int32;
    assert(nLimit <= 32, "nLimit <= 32");
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    // get the fanins
    pFanin0 = Ivy_ObjFanin0(pObj);
    pFanin1 = Ivy_ObjFanin1(pObj);
    // get the supports
    pSupp0 = Ivy_ObjSupp( pAig, pFanin0 );
    pSupp1 = Ivy_ObjSupp( pAig, pFanin1 );
    pSupp  = Ivy_ObjSupp( pAig, pObj );
    assert(pSupp.fMark == 0, "pSupp.fMark == 0");

    // get the area
    if pSupp.nRefs == 0 {
        AreaBef = Ivy_FastMapNodeAreaDerefed( pAig, pObj );
    } else {
        AreaBef = Ivy_FastMapNodeAreaRefed( pAig, pObj ); }
//    if ( AreaBef == 1 )
//        return;

    // deref the cut if the node is refed
    if pSupp.nRefs != 0 {
        Ivy_FastMapNodeDeref( pAig, pObj ); }

    // get the old delay of the node
    DelayOld = Ivy_FastMapNodeDelay(pAig, pObj);
    assert(DelayOld <= pSupp.DelayR, "DelayOld <= pSupp.DelayR");
    // copy the current cut
    memcpy( Store, pSupp.pArray, sizeof(int) * pSupp.nSize );
    StoreSize = pSupp.nSize;
    // get the fanin support
    if Ivy_ObjRefs(pFanin0) > 2 && pSupp0.Delay < pSupp.DelayR {
//    if ( pSupp0->nRefs > 0 && pSupp0->Delay < pSupp->DelayR ) // this leads to 2% worse results
    
        pSupp0 = pTemp0;
        pSupp0.nSize = 1;
        pSupp0.pArray[0] = Ivy_ObjFaninId0(pObj);
    }
    // get the fanin support
    if Ivy_ObjRefs(pFanin1) > 2 && pSupp1.Delay < pSupp.DelayR {
//    if ( pSupp1->nRefs > 0 && pSupp1->Delay < pSupp->DelayR )
    
        pSupp1 = pTemp1;
        pSupp1.nSize = 1;
        pSupp1.pArray[0] = Ivy_ObjFaninId1(pObj);
    }
    // merge the cuts
    if pSupp0.nSize < pSupp1.nSize {
        RetValue = Ivy_FastMapMerge( pSupp1, pSupp0, pSupp, nLimit );
    } else {
        RetValue = Ivy_FastMapMerge( pSupp0, pSupp1, pSupp, nLimit ); }
    if !(RetValue != 0) {
        pSupp.nSize = 2;
        pSupp.pArray[0] = Ivy_ObjFaninId0(pObj);
        pSupp.pArray[1] = Ivy_ObjFaninId1(pObj);
    }

    // check the resulting delay
    pSupp.Delay = Ivy_FastMapNodeDelay(pAig, pObj);

    RefsOld = pSupp.nRefs; pSupp.nRefs = 0;
    AreaAft = Ivy_FastMapNodeAreaDerefed( pAig, pObj );
    pSupp.nRefs = RefsOld;

    if AreaAft > AreaBef || pSupp.Delay > pSupp.DelayR {
//    if ( pSupp->Delay > pSupp->DelayR )
    
        pSupp.nSize = StoreSize;
        memcpy( pSupp.pArray, Store, sizeof(int) * pSupp.nSize );
        pSupp.Delay = DelayOld;
//        printf( "-" );
    }
//    else
//        printf( "+" );

    if pSupp.nRefs != 0 {
        Ivy_FastMapNodeRef( pAig, pObj ); }
}

/**Function*************************************************************

  Synopsis    [Performs fast mapping for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNode(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32) {
    var pSupp0, pSupp1, pSupp *Ivy_Supp_t;
    var fFaninParam int32  = 2;
    var RetValue int32;
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    // get the supports
    pSupp0 = Ivy_ObjSupp( pAig, Ivy_ObjFanin0(pObj) );
    pSupp1 = Ivy_ObjSupp( pAig, Ivy_ObjFanin1(pObj) );
    pSupp  = Ivy_ObjSupp( pAig, pObj );
    pSupp.fMark = 0;
    // get the delays
    if pSupp0.Delay == pSupp1.Delay {
        if pSupp0.Delay == 0 {
            pSupp.Delay = pSupp0.Delay + 1;
        } else {
            pSupp.Delay = pSupp0.Delay;
        }
    } else if pSupp0.Delay > pSupp1.Delay {
        pSupp.Delay = pSupp0.Delay;
        pSupp1 = Ivy_ObjSupp( pAig, Ivy_ManConst1(pAig) );
        pSupp1.pArray[0] = Ivy_ObjFaninId1(pObj);
    
    } else { // if ( pSupp0->Delay < pSupp1->Delay )
    
        pSupp.Delay = pSupp1.Delay;
        pSupp0 = Ivy_ObjSupp( pAig, Ivy_ManConst1(pAig) );
        pSupp0.pArray[0] = Ivy_ObjFaninId0(pObj);
    }
    // merge the cuts
    if pSupp0.nSize < pSupp1.nSize {
        RetValue = Ivy_FastMapMerge( pSupp1, pSupp0, pSupp, nLimit );
    } else {
        RetValue = Ivy_FastMapMerge( pSupp0, pSupp1, pSupp, nLimit ); }
    if !(RetValue != 0) {
        pSupp.Delay++;
        if fFaninParam == 2 {
            pSupp.nSize = 2;
            pSupp.pArray[0] = Ivy_ObjFaninId0(pObj);
            pSupp.pArray[1] = Ivy_ObjFaninId1(pObj);
        
        } else if fFaninParam == 3 {
            var pFanin0, pFanin1, pFaninA, pFaninB *Ivy_Obj_t;
            pFanin0 = Ivy_ObjFanin0(pObj);
            pFanin1 = Ivy_ObjFanin1(pObj);
            pSupp.nSize = 0;
            // process the first fanin
            if Ivy_ObjIsNodeInt1(pFanin0) != 0 {
                pFaninA = Ivy_ObjFanin0(pFanin0);
                pFaninB = Ivy_ObjFanin1(pFanin0);
                if (Ivy_ObjIsNodeInt1(pFaninA) != 0) && (Ivy_ObjIsNodeInt1(pFaninB) != 0) {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin0);
                } else {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninA);
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninB);
                }
            
            } else {
                pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin0); }
            // process the second fanin
            if Ivy_ObjIsNodeInt1(pFanin1) != 0 {
                pFaninA = Ivy_ObjFanin0(pFanin1);
                pFaninB = Ivy_ObjFanin1(pFanin1);
                if (Ivy_ObjIsNodeInt1(pFaninA) != 0) && (Ivy_ObjIsNodeInt1(pFaninB) != 0) {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin1);
                } else {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninA);
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninB);
                }
            
            } else {
                pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin1); }
            // sort the fanins
            Vec_IntSelectSort( pSupp.pArray, pSupp.nSize );
            pSupp.nSize = Vec_IntRemoveDup( pSupp.pArray, pSupp.nSize );
            assert(pSupp.pArray[0] < pSupp.pArray[1], "pSupp.pArray[0] < pSupp.pArray[1]");
        
        } else if fFaninParam == 4 {
            var pFanin0, pFanin1, pFaninA, pFaninB *Ivy_Obj_t;
            pFanin0 = Ivy_ObjFanin0(pObj);
            pFanin1 = Ivy_ObjFanin1(pObj);
            pSupp.nSize = 0;
            // consider the case when exactly one of them is internal
            if Ivy_ObjIsNodeInt1(pFanin0) ^ Ivy_ObjIsNodeInt1(pFanin1) {
                pSupp0 = Ivy_ObjSupp( pAig, Ivy_ObjFanin0(pObj) );
                pSupp1 = Ivy_ObjSupp( pAig, Ivy_ObjFanin1(pObj) );
                if (Ivy_ObjIsNodeInt1(pFanin0) != 0) && pSupp0.nSize < nLimit {
                    // inlining of macro IVY_MAX
                    
                    if (pSupp0.Delay) > (pSupp1.Delay + 1) {
                        pSupp.Delay = (pSupp0.Delay);
                    } else {
                        pSupp.Delay = (pSupp1.Delay + 1);
                    }
                    pSupp1 = Ivy_ObjSupp( pAig, Ivy_ManConst1(pAig) );
                    pSupp1.pArray[0] = Ivy_ObjId(pFanin1);
                    // merge the cuts
                    RetValue = Ivy_FastMapMerge( pSupp0, pSupp1, pSupp, nLimit );
                    assert((RetValue != 0), "RetValue");
                    assert(pSupp.nSize > 1, "pSupp.nSize > 1");
                    return;
                }
                if (Ivy_ObjIsNodeInt1(pFanin1) != 0) && pSupp1.nSize < nLimit {
                    // inlining of macro IVY_MAX
                    
                    if (pSupp1.Delay) > (pSupp0.Delay + 1) {
                        pSupp.Delay = (pSupp1.Delay);
                    } else {
                        pSupp.Delay = (pSupp0.Delay + 1);
                    }
                    pSupp0 = Ivy_ObjSupp( pAig, Ivy_ManConst1(pAig) );
                    pSupp0.pArray[0] = Ivy_ObjId(pFanin0);
                    // merge the cuts
                    RetValue = Ivy_FastMapMerge( pSupp1, pSupp0, pSupp, nLimit );
                    assert((RetValue != 0), "RetValue");
                    assert(pSupp.nSize > 1, "pSupp.nSize > 1");
                    return;
                }
            }
            // process the first fanin
            if Ivy_ObjIsNodeInt1(pFanin0) != 0 {
                pFaninA = Ivy_ObjFanin0(pFanin0);
                pFaninB = Ivy_ObjFanin1(pFanin0);
                if (Ivy_ObjIsNodeInt1(pFaninA) != 0) && (Ivy_ObjIsNodeInt1(pFaninB) != 0) {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin0);
                } else {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninA);
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninB);
                }
            
            } else {
                pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin0); }
            // process the second fanin
            if Ivy_ObjIsNodeInt1(pFanin1) != 0 {
                pFaninA = Ivy_ObjFanin0(pFanin1);
                pFaninB = Ivy_ObjFanin1(pFanin1);
                if (Ivy_ObjIsNodeInt1(pFaninA) != 0) && (Ivy_ObjIsNodeInt1(pFaninB) != 0) {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin1);
                } else {
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninA);
                    pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFaninB);
                }
            
            } else {
                pSupp.pArray[int32((pSupp.nSize++))] = Ivy_ObjId(pFanin1); }
            // sort the fanins
            Vec_IntSelectSort( pSupp.pArray, pSupp.nSize );
            pSupp.nSize = Vec_IntRemoveDup( pSupp.pArray, pSupp.nSize );
            assert(pSupp.pArray[0] < pSupp.pArray[1], "pSupp.pArray[0] < pSupp.pArray[1]");
            assert(pSupp.nSize > 1, "pSupp.nSize > 1");
        }
    }
    assert(pSupp.Delay > 0, "pSupp.Delay > 0");
}

/**Function*************************************************************

  Synopsis    [Merges two supports]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapMerge(pSupp0 *Ivy_Supp_t, pSupp1 *Ivy_Supp_t, pSupp *Ivy_Supp_t, nLimit int32) int32 { 
    var i, k, c int32;
    assert(pSupp0.nSize >= pSupp1.nSize, "pSupp0.nSize >= pSupp1.nSize");
    // the case of the largest cut sizes
    if pSupp0.nSize == nLimit && pSupp1.nSize == nLimit {
        for  i = 0; i < pSupp0.nSize; i++  {
            if pSupp0.pArray[i] != pSupp1.pArray[i] {
                return 0; }
        }
        for  i = 0; i < pSupp0.nSize; i++  {
            pSupp.pArray[i] = pSupp0.pArray[i];
        }
        pSupp.nSize = pSupp0.nSize;
        return 1;
    }
    // the case when one of the cuts is the largest
    if pSupp0.nSize == nLimit {
        for  i = 0; i < pSupp1.nSize; i++  {
            for  k = pSupp0.nSize - 1; k >= 0; k--  {
                if pSupp0.pArray[k] == pSupp1.pArray[i] {
                    break; }
            }
            if k == -1 { // did not find
                return 0; }
        }
        for  i = 0; i < pSupp0.nSize; i++  {
            pSupp.pArray[i] = pSupp0.pArray[i];
        }
        pSupp.nSize = pSupp0.nSize;
        return 1;
    }

    // compare two cuts with different numbers
    i = k = 0;
    for  c = 0; c < nLimit; c++  {
        if k == pSupp1.nSize {
            if i == pSupp0.nSize {
                pSupp.nSize = c;
                return 1;
            }
            pSupp.pArray[c] = pSupp0.pArray[i]; i++;
            continue;
        }
        if i == pSupp0.nSize {
            if k == pSupp1.nSize {
                pSupp.nSize = c;
                return 1;
            }
            pSupp.pArray[c] = pSupp1.pArray[k]; k++;
            continue;
        }
        if pSupp0.pArray[i] < pSupp1.pArray[k] {
            pSupp.pArray[c] = pSupp0.pArray[i]; i++;
            continue;
        }
        if pSupp0.pArray[i] > pSupp1.pArray[k] {
            pSupp.pArray[c] = pSupp1.pArray[k]; k++;
            continue;
        }
        pSupp.pArray[c] = pSupp0.pArray[i]; i++; 
        k++;
    }
    if i < pSupp0.nSize || k < pSupp1.nSize {
        return 0; }
    pSupp.nSize = c;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Creates integer vector with the support of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapReadSupp(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, vLeaves *Vec_Int_t) {
    var pSupp *Ivy_Supp_t;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    vLeaves.nCap   = 8;
    vLeaves.nSize  = pSupp.nSize;
    vLeaves.pArray = pSupp.pArray;
}

/**Function*************************************************************

  Synopsis    [Sets the required times of the intermediate nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapRequired_rec(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, pRoot *Ivy_Obj_t, DelayR int32) {
    var pSupp *Ivy_Supp_t;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    if pObj != pRoot && (pSupp.nRefs > 0 || Ivy_ObjIsCi(pObj)) {
        return; }
    Ivy_FastMapRequired_rec( pAig, Ivy_ObjFanin0(pObj), pRoot, DelayR );
    Ivy_FastMapRequired_rec( pAig, Ivy_ObjFanin1(pObj), pRoot, DelayR );
//    assert( pObj == pRoot || pSupp->DelayR == IVY_INFINITY );
    pSupp.DelayR = DelayR;
}

/**Function*************************************************************

  Synopsis    [Computes the required times for each node.]

  Description [Sets reference counters for each node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapRequired(pAig *Ivy_Man_t, Delay int32, fSetInter int32) {
    var vLuts *Vec_Vec_t;
    var vNodes *Vec_Ptr_t;
    var pObj *Ivy_Obj_t;
    var pSupp, pSuppF *Ivy_Supp_t;
    var i, k, c int32;
    // clean the required times
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vPis, i)), 1); i++  {
        pSupp = Ivy_ObjSupp( pAig, pObj );
        pSupp.DelayR = IVY_INFINITY;
        pSupp.nRefs = 0;
    }
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        pSupp = Ivy_ObjSupp( pAig, pObj );
        pSupp.DelayR = IVY_INFINITY;
        pSupp.nRefs = 0;
    }
    }
    // set the required times of the POs
    // inlining of macro Ivy_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vPos)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vPos, i)), 1); i++  {
        pSupp = Ivy_ObjSupp( pAig, Ivy_ObjFanin0(pObj) );
        pSupp.DelayR = Delay;
        pSupp.nRefs++;
    }
    // get the levelized nodes used in the mapping
    vLuts = ((Ivy_SuppMan_t *)pAig.pData).vLuts;
    // propagate the required times
    // inlining of macro Vec_VecForEachLevelReverse
    for  i = Vec_VecSize(vLuts)-1; (i >= 0) && (((vNodes) = Vec_VecEntry(vLuts, i)), 1); i--  {
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
        pSupp = Ivy_ObjSupp( pAig, pObj );
        assert(pSupp.nRefs > 0, "pSupp.nRefs > 0");
        for  c = 0; c < pSupp.nSize; c++  {
            pSuppF = Ivy_ObjSupp( pAig, Ivy_ManObj(pAig, pSupp.pArray[c]) );
            // inlining of macro IVY_MIN
            
            if (pSuppF.DelayR) < (pSupp.DelayR - 1) {
                pSuppF.DelayR = (pSuppF.DelayR);
            } else {
                pSuppF.DelayR = (pSupp.DelayR - 1);
            }
            pSuppF.nRefs++;
        }
    }
    }
/*
    // print out some of the required times
    Ivy_ManForEachPi( pAig, pObj, i )
    {
        pSupp = Ivy_ObjSupp( pAig, pObj );
        printf( "%d ", pSupp->DelayR );
    }
    printf( "\n" );    
*/

    if fSetInter != 0 {
        // set the required times of the intermediate nodes
        // inlining of macro Vec_VecForEachLevelReverse
        for  i = Vec_VecSize(vLuts)-1; (i >= 0) && (((vNodes) = Vec_VecEntry(vLuts, i)), 1); i--  {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
            pSupp = Ivy_ObjSupp( pAig, pObj );
            Ivy_FastMapRequired_rec( pAig, pObj, pObj, pSupp.DelayR ); 
        }
        }
        // make sure that all required times are assigned
        // inlining of macro Ivy_ManForEachNode
        // inlining of macro Ivy_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
            pSupp = Ivy_ObjSupp( pAig, pObj );
            assert(pSupp.DelayR < IVY_INFINITY, "pSupp.DelayR < IVY_INFINITY");
        }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Performs area recovery for each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapRecover(pAig *Ivy_Man_t, nLimit int32) {
    var vFront, vFrontOld *Vec_Ptr_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    vFront = Vec_PtrAlloc( nLimit );
    vFrontOld = Vec_PtrAlloc( nLimit );
    Ivy_ManCleanTravId( pAig );
    // iterate through all nodes in the topological order
    // inlining of macro Ivy_ManForEachNode
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsNode(pObj) {} else {
        Ivy_FastMapNodeRecover( pAig, pObj, nLimit, vFront, vFrontOld ); }
    }
    Vec_PtrFree( vFrontOld );
    Vec_PtrFree( vFront );
}

/**Function*************************************************************

  Synopsis    [Computes the delay of the cut rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeDelay(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pSupp, pSuppF *Ivy_Supp_t;
    var c, Delay int32;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    for  c = 0; c < pSupp.nSize; c++  {
        pSuppF = Ivy_ObjSupp( pAig, Ivy_ManObj(pAig, pSupp.pArray[c]) );
        // inlining of macro IVY_MAX
        
        if (Delay) > (pSuppF.Delay) {
            Delay = (Delay);
        } else {
            Delay = (pSuppF.Delay);
        }
    }
    return 1 + Delay;
}


/**function*************************************************************

  synopsis    [References the cut.]

  description [This procedure is similar to the procedure NodeReclaim.]
               
  sideeffects []

  seealso     []

***********************************************************************/
func Ivy_FastMapNodeRef(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pSupp, pSuppF *Ivy_Supp_t;
    var pNodeChild *Ivy_Obj_t;
    var aArea, i int32;
    // start the area of this cut
    aArea = 1;
    // go through the children
    pSupp = Ivy_ObjSupp( pAig, pObj );
    assert(pSupp.nSize > 1, "pSupp.nSize > 1");
    for  i = 0; i < pSupp.nSize; i++  {
        pNodeChild = Ivy_ManObj(pAig, pSupp.pArray[i]);
        pSuppF = Ivy_ObjSupp( pAig, pNodeChild );
        assert(pSuppF.nRefs >= 0, "pSuppF.nRefs >= 0");
        if pSuppF.nRefs++ > 0 {  
            continue; }
        if pSuppF.nSize == 1 { 
            continue; }
        aArea += Ivy_FastMapNodeRef( pAig, pNodeChild );
    }
    return aArea;
}

/**function*************************************************************

  synopsis    [Dereferences the cut.]

  description [This procedure is similar to the procedure NodeRecusiveDeref.]
               
  sideeffects []

  seealso     []

***********************************************************************/
func Ivy_FastMapNodeDeref(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pSupp, pSuppF *Ivy_Supp_t;
    var pNodeChild *Ivy_Obj_t;
    var aArea, i int32;
    // start the area of this cut
    aArea = 1;
    // go through the children
    pSupp = Ivy_ObjSupp( pAig, pObj );
    assert(pSupp.nSize > 1, "pSupp.nSize > 1");
    for  i = 0; i < pSupp.nSize; i++  {
        pNodeChild = Ivy_ManObj(pAig, pSupp.pArray[i]);
        pSuppF = Ivy_ObjSupp( pAig, pNodeChild );
        assert(pSuppF.nRefs > 0, "pSuppF.nRefs > 0");
        if --pSuppF.nRefs > 0 {  
            continue; }
        if pSuppF.nSize == 1 { 
            continue; }
        aArea += Ivy_FastMapNodeDeref( pAig, pNodeChild );
    }
    return aArea;
}

/**function*************************************************************

  synopsis    [Computes the exact area associated with the cut.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Ivy_FastMapNodeAreaRefed(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pSupp *Ivy_Supp_t;
    var aResult, aResult2 int32;
    if Ivy_ObjIsCi(pObj) {
        return 0; }
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    pSupp = Ivy_ObjSupp( pAig, pObj );
    assert(pSupp.nRefs > 0, "pSupp.nRefs > 0");
    aResult  = Ivy_FastMapNodeDeref( pAig, pObj );
    aResult2 = Ivy_FastMapNodeRef( pAig, pObj );
    assert(aResult == aResult2, "aResult == aResult2");
    return aResult;
}

/**function*************************************************************

  synopsis    [Computes the exact area associated with the cut.]

  description []
               
  sideeffects []

  seealso     []

***********************************************************************/
func Ivy_FastMapNodeAreaDerefed(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pSupp *Ivy_Supp_t;
    var aResult, aResult2 int32;
    if Ivy_ObjIsCi(pObj) {
        return 0; }
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    pSupp = Ivy_ObjSupp( pAig, pObj );
    assert(pSupp.nRefs == 0, "pSupp.nRefs == 0");
    aResult2 = Ivy_FastMapNodeRef( pAig, pObj );
    aResult  = Ivy_FastMapNodeDeref( pAig, pObj );
    assert(aResult == aResult2, "aResult == aResult2");
    return aResult;
}




/**Function*************************************************************

  Synopsis    [Counts the number of nodes with no external fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapCutCost(pAig *Ivy_Man_t, vFront *Vec_Ptr_t) int32 {
    var pSuppF *Ivy_Supp_t;
    var pFanin *Ivy_Obj_t;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        pSuppF = Ivy_ObjSupp( pAig, pFanin );
        if pSuppF.nRefs == 0 {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Performs area recovery for each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapMark_rec(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) {
    if Ivy_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    Ivy_FastMapMark_rec( pAig, Ivy_ObjFanin0(pObj) );
    Ivy_FastMapMark_rec( pAig, Ivy_ObjFanin1(pObj) );
    Ivy_ObjSetTravIdCurrent(pAig, pObj);
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the number of fanins will grow.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeWillGrow(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pFanin0, pFanin1 *Ivy_Obj_t;
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    pFanin0 = Ivy_ObjFanin0(pObj);
    pFanin1 = Ivy_ObjFanin1(pObj);
    return !Ivy_ObjIsTravIdCurrent(pAig, pFanin0) && !Ivy_ObjIsTravIdCurrent(pAig, pFanin1);
}

/**Function*************************************************************

  Synopsis    [Returns the increase in the number of fanins with no external refs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeFaninCost(pAig *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pSuppF *Ivy_Supp_t;
    var pFanin *Ivy_Obj_t;
    var Counter int32;
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    // check if the node has external refs
    pSuppF = Ivy_ObjSupp( pAig, pObj );
    if pSuppF.nRefs == 0 {
        Counter--; }
    // increment the number of fanins without external refs
    pFanin = Ivy_ObjFanin0(pObj);
    pSuppF = Ivy_ObjSupp( pAig, pFanin );
    if !Ivy_ObjIsTravIdCurrent(pAig, pFanin) && pSuppF.nRefs == 0 {
        Counter++; }
    // increment the number of fanins without external refs
    pFanin = Ivy_ObjFanin1(pObj);
    pSuppF = Ivy_ObjSupp( pAig, pFanin );
    if !Ivy_ObjIsTravIdCurrent(pAig, pFanin) && pSuppF.nRefs == 0 {
        Counter++; }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Updates the frontier.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeFaninUpdate(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, vFront *Vec_Ptr_t) {
    var pFanin *Ivy_Obj_t;
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    Vec_PtrRemove( vFront, pObj );
    pFanin = Ivy_ObjFanin0(pObj);
    if !Ivy_ObjIsTravIdCurrent(pAig, pFanin) {
        Ivy_ObjSetTravIdCurrent(pAig, pFanin);
        Vec_PtrPush( vFront, pFanin );
    }
    pFanin = Ivy_ObjFanin1(pObj);
    if !Ivy_ObjIsTravIdCurrent(pAig, pFanin) {
        Ivy_ObjSetTravIdCurrent(pAig, pFanin);
        Vec_PtrPush( vFront, pFanin );
    }
}

/**Function*************************************************************

  Synopsis    [Compacts the number of external refs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeFaninCompact0(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t) int32 {
    var pFanin *Ivy_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        if Ivy_ObjIsCi(pFanin) {
            continue; }
        if Ivy_FastMapNodeWillGrow(pAig, pFanin) != 0 {
            continue; }
        if Ivy_FastMapNodeFaninCost(pAig, pFanin) <= 0 {
            Ivy_FastMapNodeFaninUpdate( pAig, pFanin, vFront );
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Compacts the number of external refs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeFaninCompact1(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t) int32 {
    var pFanin *Ivy_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        if Ivy_ObjIsCi(pFanin) {
            continue; }
        if Ivy_FastMapNodeFaninCost(pAig, pFanin) < 0 {
            Ivy_FastMapNodeFaninUpdate( pAig, pFanin, vFront );
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Compacts the number of external refs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeFaninCompact2(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t) int32 {
    var pFanin *Ivy_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        if Ivy_ObjIsCi(pFanin) {
            continue; }
        if Ivy_FastMapNodeFaninCost(pAig, pFanin) <= 0 {
            Ivy_FastMapNodeFaninUpdate( pAig, pFanin, vFront );
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Compacts the number of external refs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeFaninCompact_int(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t) int32 {
    if Ivy_FastMapNodeFaninCompact0(pAig, pObj, nLimit, vFront) != 0 {
        return 1; }
    if Vec_PtrSize(vFront) < nLimit && (Ivy_FastMapNodeFaninCompact1(pAig, pObj, nLimit, vFront) != 0) {
        return 1; }
    if Vec_PtrSize(vFront) < nLimit && (Ivy_FastMapNodeFaninCompact2(pAig, pObj, nLimit, vFront) != 0) {
        return 1; }
    assert(Vec_PtrSize(vFront) <= nLimit, "Vec_PtrSize(vFront) <= nLimit");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Compacts the number of external refs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeFaninCompact(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t) {
    for  Ivy_FastMapNodeFaninCompact_int( pAig, pObj, nLimit, vFront ) != 0  {; }
}

/**Function*************************************************************

  Synopsis    [Prepares node mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodePrepare(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vFrontOld *Vec_Ptr_t) {
    var pSupp *Ivy_Supp_t;
    var pFanin *Ivy_Obj_t;
    var i int32;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    // expand the cut downwards from the given place
    Vec_PtrClear( vFront );
    Vec_PtrClear( vFrontOld );
    Ivy_ManIncrementTravId( pAig );
    for  i = 0; i < pSupp.nSize; i++  {
        pFanin = Ivy_ManObj(pAig, pSupp.pArray[i]);
        Vec_PtrPush( vFront, pFanin );
        Vec_PtrPush( vFrontOld, pFanin );
        Ivy_ObjSetTravIdCurrent( pAig, pFanin );
    }
    // mark the nodes in the cone
    Ivy_FastMapMark_rec( pAig, pObj );
}

/**Function*************************************************************

  Synopsis    [Updates the frontier.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeUpdate(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, vFront *Vec_Ptr_t) {
    var pSupp *Ivy_Supp_t;
    var pFanin *Ivy_Obj_t;
    var i int32;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    // deref node's cut
    Ivy_FastMapNodeDeref( pAig, pObj );
    // update the node's cut
    pSupp.nSize = Vec_PtrSize(vFront);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pFanin) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        pSupp.pArray[i] = pFanin.Id;
    }
    // ref the new cut
    Ivy_FastMapNodeRef( pAig, pObj );
}

/**Function*************************************************************

  Synopsis    [Performs area recovery for each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeRecover2(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vFrontOld *Vec_Ptr_t) {
    var pSupp *Ivy_Supp_t;
    var CostBef, CostAft int32;
    var AreaBef, AreaAft int32;
    pSupp = Ivy_ObjSupp( pAig, pObj );
//    if ( pSupp->nRefs == 0 )
//        return;
    if pSupp.nRefs == 0 {
        AreaBef = Ivy_FastMapNodeAreaDerefed( pAig, pObj );
    } else {
        AreaBef = Ivy_FastMapNodeAreaRefed( pAig, pObj ); }
    // get the area
    if AreaBef == 1 {
        return; }

    if pSupp.nRefs == 0 {
        pSupp.nRefs = 1000000;
        Ivy_FastMapNodeRef( pAig, pObj );
    }
    // the cut is non-trivial
    Ivy_FastMapNodePrepare( pAig, pObj, nLimit, vFront, vFrontOld );
    // iteratively modify the cut
    CostBef = Ivy_FastMapCutCost( pAig, vFront );
    Ivy_FastMapNodeFaninCompact( pAig, pObj, nLimit, vFront );
    CostAft = Ivy_FastMapCutCost( pAig, vFront );
    assert(CostBef >= CostAft, "CostBef >= CostAft");
    // update the node
    Ivy_FastMapNodeUpdate( pAig, pObj, vFront );
    // get the new area
    AreaAft = Ivy_FastMapNodeAreaRefed( pAig, pObj );
    if AreaAft > AreaBef {
        Ivy_FastMapNodeUpdate( pAig, pObj, vFrontOld );
        AreaAft = Ivy_FastMapNodeAreaRefed( pAig, pObj );
        assert(AreaAft == AreaBef, "AreaAft == AreaBef");
    }
    if pSupp.nRefs == 1000000 {
        pSupp.nRefs = 0;
        Ivy_FastMapNodeDeref( pAig, pObj );
    }
}

/**Function*************************************************************

  Synopsis    [Performs area recovery for each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeRecover(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vFrontOld *Vec_Ptr_t) {
    var pSupp *Ivy_Supp_t;
    var CostBef, CostAft int32;
    var AreaBef, AreaAft int32;
    var DelayOld int32;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    DelayOld = pSupp.Delay = Ivy_FastMapNodeDelay( pAig, pObj );
    assert(pSupp.Delay <= pSupp.DelayR, "pSupp.Delay <= pSupp.DelayR");
    if pSupp.nRefs == 0 {
        return; }
    // get the area
    AreaBef = Ivy_FastMapNodeAreaRefed( pAig, pObj );
//    if ( AreaBef == 1 )
//        return;
    // the cut is non-trivial
    Ivy_FastMapNodePrepare( pAig, pObj, nLimit, vFront, vFrontOld );
    // iteratively modify the cut
    Ivy_FastMapNodeDeref( pAig, pObj );
    CostBef = Ivy_FastMapCutCost( pAig, vFront );
    Ivy_FastMapNodeFaninCompact( pAig, pObj, nLimit, vFront );
    CostAft = Ivy_FastMapCutCost( pAig, vFront );
    Ivy_FastMapNodeRef( pAig, pObj );
    assert(CostBef >= CostAft, "CostBef >= CostAft");
    // update the node
    Ivy_FastMapNodeUpdate( pAig, pObj, vFront );
    pSupp.Delay = Ivy_FastMapNodeDelay( pAig, pObj );
    // get the new area
    AreaAft = Ivy_FastMapNodeAreaRefed( pAig, pObj );
    if AreaAft > AreaBef || pSupp.Delay > pSupp.DelayR {
        Ivy_FastMapNodeUpdate( pAig, pObj, vFrontOld );
        AreaAft = Ivy_FastMapNodeAreaRefed( pAig, pObj );
        assert(AreaAft == AreaBef, "AreaAft == AreaBef");
        pSupp.Delay = DelayOld;
    }
}

/**Function*************************************************************

  Synopsis    [Performs area recovery for each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_FastMapNodeRecover4(pAig *Ivy_Man_t, pObj *Ivy_Obj_t, nLimit int32, vFront *Vec_Ptr_t, vFrontOld *Vec_Ptr_t) {
    var pSupp *Ivy_Supp_t;
    var CostBef, CostAft int32;
    var AreaBef, AreaAft int32;
    var DelayOld int32;
    pSupp = Ivy_ObjSupp( pAig, pObj );
    DelayOld = pSupp.Delay = Ivy_FastMapNodeDelay( pAig, pObj );
    assert(pSupp.Delay <= pSupp.DelayR, "pSupp.Delay <= pSupp.DelayR");
//    if ( pSupp->nRefs == 0 )
//        return;
//    AreaBef = Ivy_FastMapNodeAreaRefed( pAig, pObj );
    // get the area
    if pSupp.nRefs == 0 {
        AreaBef = Ivy_FastMapNodeAreaDerefed( pAig, pObj );
    } else {
        AreaBef = Ivy_FastMapNodeAreaRefed( pAig, pObj ); }
    if AreaBef == 1 {
        return; }

    if pSupp.nRefs == 0 {
        pSupp.nRefs = 1000000;
        Ivy_FastMapNodeRef( pAig, pObj );
    }
    // the cut is non-trivial
    Ivy_FastMapNodePrepare( pAig, pObj, nLimit, vFront, vFrontOld );
    // iteratively modify the cut
    CostBef = Ivy_FastMapCutCost( pAig, vFront );
    Ivy_FastMapNodeFaninCompact( pAig, pObj, nLimit, vFront );
    CostAft = Ivy_FastMapCutCost( pAig, vFront );
    assert(CostBef >= CostAft, "CostBef >= CostAft");
    // update the node
    Ivy_FastMapNodeUpdate( pAig, pObj, vFront );
    pSupp.Delay = Ivy_FastMapNodeDelay( pAig, pObj );
    // get the new area
    AreaAft = Ivy_FastMapNodeAreaRefed( pAig, pObj );
    if AreaAft > AreaBef || pSupp.Delay > pSupp.DelayR {
        Ivy_FastMapNodeUpdate( pAig, pObj, vFrontOld );
        AreaAft = Ivy_FastMapNodeAreaRefed( pAig, pObj );
        assert(AreaAft == AreaBef, "AreaAft == AreaBef");
        pSupp.Delay = DelayOld;
    }
    if pSupp.nRefs == 1000000 {
        pSupp.nRefs = 0;
        Ivy_FastMapNodeDeref( pAig, pObj );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

