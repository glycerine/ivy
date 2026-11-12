package ivy

import "fmt"

/**CFile****************************************************************

  FileName    [ivyHaig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [HAIG management procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyHaig.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/* 
    HAIGing rules in working AIG:
    - Each node in the working AIG has a pointer to the corresponding node in HAIG 
      (this node is not necessarily the representative of the equivalence class of HAIG nodes)
    - This pointer is complemented if the AIG node and its corresponding HAIG node have different phase

    Choice node rules in HAIG:
    - Equivalent nodes are linked into a ring
    - Exactly one node in the ring has fanouts (this node is called the representative)
    - The pointer going from a node to the next node in the ring is complemented 
      if the first node is complemented, compared to the representative node of the equivalence class
    - (consequence of the above) The representative node always has non-complemented pointer to the next node
    - New nodes are inserted into the ring immediately after the representative node
*/

// returns the representative node of the given HAIG node
// Ivy_HaigObjRepr is implemented in the ported core API group.


// counts the number of nodes in the equivalence class
func Ivy_HaigObjCountClass(pObj *Ivy_Obj_t) int32 {
    var pTemp *Ivy_Obj_t;
    var Counter int32;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(Ivy_ObjRefs(pObj) > 0, "Ivy_ObjRefs(pObj) > 0");
    if pObj.pEquiv == nil {
        return 1; }
    assert(!Ivy_IsComplement(pObj.pEquiv), "!Ivy_IsComplement(pObj.pEquiv)");
    Counter = 1;
    for  pTemp = pObj.pEquiv; pTemp != pObj; pTemp = Ivy_Regular(pTemp.pEquiv)  {
        Counter++;
    }
    return Counter;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts HAIG for the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigStart(p *Ivy_Man_t, fVerbose int32) {
    var vLatches *Vec_Int_t;
    var pObj *Ivy_Obj_t;
    var i int32;
    assert(p.pHaig == nil, "p.pHaig == nil");
    p.pHaig = Ivy_ManDup( p );

    if fVerbose != 0 {
        fmt.Printf("Starting : ");
        Ivy_ManPrintStats( p.pHaig );
    }

    // collect latches of design D and set their values to be DC
    vLatches = Vec_IntAlloc( 100 );
    // inlining of macro Ivy_ManForEachLatch
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pHaig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pHaig.vObjs, i)), 1); i++  { if (pObj) == nil {} else if !Ivy_ObjIsLatch(pObj) {} else {
        pObj.Init = IVY_INIT_DC;
        Vec_IntPush( vLatches, pObj.Id );
    }
    }
    p.pHaig.pData = vLatches;
/*
    {
        int x;
        Ivy_ManShow( p, 0, NULL );
        Ivy_ManShow( p->pHaig, 1, NULL );
        x = 0;
    }
*/
}

/**Function*************************************************************

  Synopsis    [Transfers the HAIG to the newly created manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigTrasfer(p *Ivy_Man_t, pNew *Ivy_Man_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    assert(p.pHaig != nil, "p.pHaig != nil");
    Ivy_ManConst1(pNew).pEquiv = Ivy_ManConst1(p).pEquiv;
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNew.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(pNew.vPis, i)), 1); i++  {
        pObj.pEquiv = Ivy_ManPi( p, i ).pEquiv;
    }
    pNew.pHaig = p.pHaig;
}

/**Function*************************************************************

  Synopsis    [Stops HAIG for the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigStop(p *Ivy_Man_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    assert(p.pHaig != nil, "p.pHaig != nil");
    Vec_IntFree( (Vec_Int_t *)p.pHaig.pData );
    Ivy_ManStop( p.pHaig );
    p.pHaig = nil;
    // remove dangling pointers to the HAIG objects
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pEquiv = nil; }
    }
}

/**Function*************************************************************

  Synopsis    [Creates a new node in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManHaigCreateObj is implemented in the ported core API group.


/**Function*************************************************************

  Synopsis    [Checks if the old node is in the TFI of the new node.]

  Description [] 
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ObjIsInTfi_rec is implemented in the ported object API group.


/**Function*************************************************************

  Synopsis    [Sets the pair of equivalent nodes in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Ivy_ManHaigCreateChoice is implemented in the ported object API group.


/**Function*************************************************************

  Synopsis    [Count the number of choices and choice nodes in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigCountChoices(p *Ivy_Man_t, pnChoices *int32) int32 {
    var pObj *Ivy_Obj_t;
    var nChoices, nChoiceNodes, Counter, i int32;
    assert(p.pHaig != nil, "p.pHaig != nil");
    nChoices = nChoiceNodes = 0;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pHaig.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.pHaig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Ivy_ObjIsTerm(pObj) || i == 0 {
            continue; }
        if Ivy_ObjRefs(pObj) == 0 {
            continue; }
        Counter = Ivy_HaigObjCountClass( pObj );
        nChoiceNodes += int32((Counter > 1));
        nChoices += Counter - 1;
//        if ( Counter > 1 )
//            printf( "Choice node %d %s\n", pObj->Id, Ivy_ObjIsLatch(pObj)? "(latch)": "" );
    }
    }
    *pnChoices = nChoices;
    return nChoiceNodes;
} 

/**Function*************************************************************

  Synopsis    [Prints statistics of the HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigPostprocess(p *Ivy_Man_t, fVerbose int32) {
    var nChoices, nChoiceNodes int32;

    assert(p.pHaig != nil, "p.pHaig != nil");

    if fVerbose != 0 {
        fmt.Printf("Final    : ");
        Ivy_ManPrintStats( p );
        fmt.Printf("HAIG     : ");
        Ivy_ManPrintStats( p.pHaig );

        // print choice node stats
        nChoiceNodes = Ivy_ManHaigCountChoices( p, &nChoices );
        fmt.Printf("Total choice nodes = %d. Total choices = %d. Skipped classes = %d.\n", nChoiceNodes, nChoices, p.pHaig.nClassesSkip); 
    }

    if Ivy_ManIsAcyclic( p.pHaig ) {
        if fVerbose != 0 {
            fmt.Printf("HAIG is acyclic\n"); }
    
    } else {
        fmt.Printf("HAIG contains a cycle\n"); }

//    if ( fVerbose )
//        Ivy_ManHaigSimulate( p );
}


/**Function*************************************************************

  Synopsis    [Applies the simulation rules.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigSimulateAnd(In0 Ivy_Init_t, In1 Ivy_Init_t) Ivy_Init_t {
    assert(In0 != IVY_INIT_NONE && In1 != IVY_INIT_NONE, "In0 != IVY_INIT_NONE && In1 != IVY_INIT_NONE");
    if In0 == IVY_INIT_DC || In1 == IVY_INIT_DC {
        return IVY_INIT_DC; }
    if In0 == IVY_INIT_1 && In1 == IVY_INIT_1 {
        return IVY_INIT_1; }
    return IVY_INIT_0;
}

/**Function*************************************************************

  Synopsis    [Applies the simulation rules.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigSimulateChoice(In0 Ivy_Init_t, In1 Ivy_Init_t) Ivy_Init_t {
    assert(In0 != IVY_INIT_NONE && In1 != IVY_INIT_NONE, "In0 != IVY_INIT_NONE && In1 != IVY_INIT_NONE");
    if (In0 == IVY_INIT_0 && In1 == IVY_INIT_1) || (In0 == IVY_INIT_1 && In1 == IVY_INIT_0) {
         fmt.Printf("Compatibility fails.\n");
         return IVY_INIT_0;
    }
    if In0 == IVY_INIT_DC && In1 == IVY_INIT_DC {
        return IVY_INIT_DC; }
    if In0 != IVY_INIT_DC {
        return In0; }
    return In1;
}

/**Function*************************************************************

  Synopsis    [Simulate HAIG using modified 3-valued simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManHaigSimulate(p *Ivy_Man_t) {
    var vNodes, vLatches, vLatchesD *Vec_Int_t;
    var pObj, pTemp *Ivy_Obj_t;
    var In0, In1 Ivy_Init_t;
    var i, k, Counter int32;
    var fVerbose int32;

    // check choices
    Ivy_ManCheckChoices( p );

    // switch to HAIG
    assert(p.pHaig != nil, "p.pHaig != nil");
    p = p.pHaig;

if fVerbose != 0 {
// inlining of macro Ivy_ManForEachPi
// inlining of macro Vec_PtrForEachEntry
for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
fmt.Printf("Setting PI %d\n", pObj.Id);
} }

    // collect latches and nodes in the DFS order
    vNodes = Ivy_ManDfsSeq( p, &vLatches );

if fVerbose != 0 {
// inlining of macro Ivy_ManForEachNodeVec
for  i = 0; i < Vec_IntSize(vNodes) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
fmt.Printf("Collected node %d with fanins %d and %d\n", pObj.Id, Ivy_ObjFanin0(pObj).Id, Ivy_ObjFanin1(pObj).Id);
} }

    // set the PI values
    Ivy_ManConst1(p).Init = IVY_INIT_1;
    // inlining of macro Ivy_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj.Init = IVY_INIT_0;
    }

    // set the latch values
    // inlining of macro Ivy_ManForEachNodeVec
    for  i = 0; i < Vec_IntSize(vLatches) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatches,i))); i++  {
        pObj.Init = IVY_INIT_DC;
    }
    // set the latches of D to be determinate
    vLatchesD = (Vec_Int_t *)p.pData;
    // inlining of macro Ivy_ManForEachNodeVec
    for  i = 0; i < Vec_IntSize(vLatchesD) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatchesD,i))); i++  {
        pObj.Init = IVY_INIT_0;
    }

    // perform several rounds of simulation
    for  k = 0; k < 10; k++  {
        // count the number of non-determinate values
        Counter = 0;
        // inlining of macro Ivy_ManForEachNodeVec
        for  i = 0; i < Vec_IntSize(vLatches) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatches,i))); i++  {
            Counter += ( pObj.Init == IVY_INIT_DC );
        }
        fmt.Printf("Iter %d : Non-determinate = %d\n", k, Counter);    
        
        // simulate the internal nodes
        // inlining of macro Ivy_ManForEachNodeVec
        for  i = 0; i < Vec_IntSize(vNodes) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
if fVerbose != 0 {
fmt.Printf("Processing node %d with fanins %d and %d\n", pObj.Id, Ivy_ObjFanin0(pObj).Id, Ivy_ObjFanin1(pObj).Id); }
            In0 = Ivy_InitNotCond( (Ivy_Init_t)Ivy_ObjFanin0(pObj).Init, Ivy_ObjFaninC0(pObj) );
            In1 = Ivy_InitNotCond( (Ivy_Init_t)Ivy_ObjFanin1(pObj).Init, Ivy_ObjFaninC1(pObj) );
            pObj.Init = Ivy_ManHaigSimulateAnd( In0, In1 );
            // simulate the equivalence class if the node is a representative
            if pObj.pEquiv && Ivy_ObjRefs(pObj) > 0 {
if fVerbose != 0 {
fmt.Printf("Processing choice node %d\n", pObj.Id); }
                In0 = (Ivy_Init_t)pObj.Init;
                assert(!Ivy_IsComplement(pObj.pEquiv), "!Ivy_IsComplement(pObj.pEquiv)");
                for  pTemp = pObj.pEquiv; pTemp != pObj; pTemp = Ivy_Regular(pTemp.pEquiv)  {
if fVerbose != 0 {
fmt.Printf("Processing secondary node %d\n", pTemp.Id); }
                    In1 = Ivy_InitNotCond( (Ivy_Init_t)pTemp.Init, Ivy_IsComplement(pTemp.pEquiv) );
                    In0 = Ivy_ManHaigSimulateChoice( In0, In1 );
                }
                pObj.Init = In0;
            }
        }

        // simulate the latches
        // inlining of macro Ivy_ManForEachNodeVec
        for  i = 0; i < Vec_IntSize(vLatches) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatches,i))); i++  {
            pObj.Level = Ivy_ObjFanin0(pObj).Init;
if fVerbose != 0 {
fmt.Printf("Using latch %d with fanin %d\n", pObj.Id, Ivy_ObjFanin0(pObj).Id); }
        }
        // inlining of macro Ivy_ManForEachNodeVec
        for  i = 0; i < Vec_IntSize(vLatches) && ((pObj) = Ivy_ManObj(p, Vec_IntEntry(vLatches,i))); i++  {
            pObj.Init = pObj.Level, pObj.Level = 0;
        }
    }
    // free arrays
    Vec_IntFree( vNodes );
    Vec_IntFree( vLatches );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

