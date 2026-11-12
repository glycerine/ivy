package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigDup.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [AIG duplication (re-strashing).]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigDup.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/saig/saig.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [Orders nodes as follows: PIs, ANDs, POs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupSimple(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i int32;
    assert(p.pManTime == nil, "p.pManTime == nil");
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCi( pNew );
        pObjNew.Level = pObj.Level;
        pObj.pData = pObjNew;
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Aig_ObjIsBuf(pObj) {
            pObjNew = Aig_ObjChild0Copy(pObj);
            pObj.pData = pObjNew;
        
        } else if Aig_ObjIsNode(pObj) {
            pObjNew = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
            pObj.pData = pObjNew;
        }
    }
    // add the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        pObj.pData = pObjNew;
    }
    assert(Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew), "Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew)");
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupSimple(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives AIG with hints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupSimpleWithHints(p *Aig_Man_t, vHints *Vec_Int_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i, Entry int32;
    assert(p.nAsserts == 0 || p.nConstrs == 0, "p.nAsserts == 0 || p.nConstrs == 0");
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
        Entry = Vec_IntEntry( vHints, Aig_ObjId(pObj) );
        if Entry == 0 || Entry == 1 {
            pObj.pData = Aig_NotCond( Aig_ManConst1(pNew), Entry ); } // restrict to the complement of constraint!!!
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
        Entry = Vec_IntEntry( vHints, Aig_ObjId(pObj) );
        if Entry == 0 || Entry == 1 {
            pObj.pData = Aig_NotCond( Aig_ManConst1(pNew), Entry ); } // restrict to the complement of constraint!!!
    }
    }
    // add the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Llb_ManDeriveAigWithHints(): The check has failed.\n"); }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupSimpleDfs_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    Aig_ManDupSimpleDfs_rec( pNew, p, Aig_ObjFanin0(pObj) );
    if Aig_ObjIsBuf(pObj) {
        return (Aig_Obj_t *)(pObj.pData = Aig_ObjChild0Copy(pObj)); }
    Aig_ManDupSimpleDfs_rec( pNew, p, Aig_ObjFanin1(pObj) );
    pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    return (Aig_Obj_t *)pObj.pData;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [Orders nodes as follows: PIs, ANDs, POs.]
               
  SideEffects [This procedure assumes that buffers are not used during
  HAIG recording. This way, each HAIG node is in one-to-one correspondence 
  with old HAIG node. There is no need to create new nodes, just reassign 
  the pointers. If it were not the case, we would need to create HAIG nodes 
  for each new node duplicated. ]

  SeeAlso     []

***********************************************************************/
func Aig_ManDupSimpleDfs(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i int32;
    assert(p.pManTime == nil, "p.pManTime == nil");
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCi( pNew );
        pObjNew.Level = pObj.Level;
        pObj.pData = pObjNew;
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if !Aig_ObjIsCo(pObj) {
            Aig_ManDupSimpleDfs_rec( pNew, p, pObj );        
            assert(pObj.Level == ((Aig_Obj_t*)pObj.pData).Level, "pObj.Level == ((Aig_Obj_t*)pObj.pData).Level");
        }
    }
    // add the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        pObj.pData = pObjNew;
    }
    assert(Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew), "Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew)");
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupSimple(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates part of the AIG manager.]

  Description [Orders nodes as follows: PIs, ANDs, POs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupSimpleDfsPart(p *Aig_Man_t, vPis *Vec_Ptr_t, vPos *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1( pNew );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vPis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // duplicate internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vPos, i)), 1); i++  {
        pObjNew = Aig_ManDupSimpleDfs_rec( pNew, p, Aig_ObjFanin0(pObj) );        
        pObjNew = Aig_NotCond( pObjNew, Aig_ObjFaninC0(pObj) );
        Aig_ObjCreateCo( pNew, pObjNew );
    }
    Aig_ManSetRegNum( pNew, 0 );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupSimple(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupOrdered(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i, nNodes int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // create the PIs
    Aig_ManCleanData( p );
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsBuf(pObj) {
            pObjNew = Aig_ObjChild0Copy(pObj);
        
        } else if Aig_ObjIsNode(pObj) {
            pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
        
        } else if Aig_ObjIsCi(pObj) {
            pObjNew = Aig_ObjCreateCi( pNew );
            pObjNew.Level = pObj.Level;
        
        } else if Aig_ObjIsCo(pObj) {
            pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        
        } else if Aig_ObjIsConst1(pObj) {
            pObjNew = Aig_ManConst1(pNew);
        
        } else {
            assert(false, "0"); }
        pObj.pData = pObjNew;
    }
    }
    assert(Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew), "Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew)");
    if nNodes = Aig_ManCleanup( pNew ) {
        fmt.Printf("Aig_ManDupOrdered(): Cleanup after AIG duplication removed %d nodes.\n", nNodes); }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // duplicate the timing manager
    if p.pManTime {
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)p.pManTime, 0 ); }
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupOrdered(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [Orders nodes as follows: PIs, ANDs, POs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupCof(p *Aig_Man_t, iInput int32, Value int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i int32;
    assert(p.pManTime == nil, "p.pManTime == nil");
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if i == iInput {
            if Value != 0 {
                pObjNew = Aig_ManConst1(pNew);
            } else {
                pObjNew = Aig_ManConst0(pNew);
            }
        } else {
            pObjNew = Aig_ObjCreateCi( pNew );
            pObjNew.Level = pObj.Level;
        }
        pObj.pData = pObjNew;
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Aig_ObjIsBuf(pObj) {
            pObjNew = Aig_ObjChild0Copy(pObj);
            pObj.pData = pObjNew;
        
        } else if Aig_ObjIsNode(pObj) {
            pObjNew = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
            pObj.pData = pObjNew;
        }
    }
    // add the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        pObj.pData = pObjNew;
    }
//    assert( Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew) );
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupSimple(): The check has failed.\n"); }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupTrim(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i, nNodes int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    // create the PIs
    Aig_ManCleanData( p );
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsNode(pObj) {
            pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
        } else if Aig_ObjIsCi(pObj) {
            if Aig_ObjRefs(pObj) > 0 || Saig_ObjIsLo(p, pObj) {
                pObjNew = Aig_ObjCreateCi(pNew);
            } else {
                pObjNew = nil;
            }
        } else if Aig_ObjIsCo(pObj) {
            pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        } else if Aig_ObjIsConst1(pObj) {
            pObjNew = Aig_ManConst1(pNew);
        } else {
            assert(false, "0"); }
        pObj.pData = pObjNew;
    }
    }
    assert(Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew), "Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew)");
    if nNodes = Aig_ManCleanup( pNew ) {
        fmt.Printf("Aig_ManDupTrim(): Cleanup after AIG duplication removed %d nodes.\n", nNodes); }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupTrim(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager to have EXOR gates.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupExor(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.fCatchExor = 1;
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // create the PIs
    Aig_ManCleanData( p );
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsBuf(pObj) {
            pObjNew = Aig_ObjChild0Copy(pObj);
        
        } else if Aig_ObjIsNode(pObj) {
            pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
        
        } else if Aig_ObjIsCi(pObj) {
            pObjNew = Aig_ObjCreateCi( pNew );
            pObjNew.Level = pObj.Level;
        
        } else if Aig_ObjIsCo(pObj) {
            pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        
        } else if Aig_ObjIsConst1(pObj) {
            pObjNew = Aig_ManConst1(pNew);
        
        } else {
            assert(false, "0"); }
        pObj.pData = pObjNew;
    }
    }
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // duplicate the timing manager
    if p.pManTime {
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)p.pManTime, 0 ); }
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupExor(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupDfs_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pObjNew *Aig_Obj_t;
    var pEquivNew *Aig_Obj_t = nil;
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    if p.pEquivs && Aig_ObjEquiv(p, pObj) {
        pEquivNew = Aig_ManDupDfs_rec( pNew, p, Aig_ObjEquiv(p, pObj) ); }
    Aig_ManDupDfs_rec( pNew, p, Aig_ObjFanin0(pObj) );
    if Aig_ObjIsBuf(pObj) {
        return (Aig_Obj_t *)(pObj.pData = Aig_ObjChild0Copy(pObj)); }
    Aig_ManDupDfs_rec( pNew, p, Aig_ObjFanin1(pObj) );
    pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
    if pEquivNew != nil {
        assert(Aig_Regular(pEquivNew).Id < Aig_Regular(pObjNew).Id, "Aig_Regular(pEquivNew).Id < Aig_Regular(pObjNew).Id");
        if pNew.pEquivs {
            pNew.pEquivs[Aig_Regular(pObjNew).Id] = Aig_Regular(pEquivNew); }        
        if pNew.pReprs {
            pNew.pReprs[Aig_Regular(pEquivNew).Id] = Aig_Regular(pObjNew); }
    }
    return (Aig_Obj_t *)(pObj.pData = pObjNew);
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [This duplicator works for AIGs with choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupDfs(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i, nNodes int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // duplicate representation of choice nodes
    if p.pEquivs {
        // inlining of macro ABC_CALLOC
        pNew.pEquivs = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p)), sizeof(Aig_Obj_t *))); }
    if p.pReprs {
        // inlining of macro ABC_CALLOC
        pNew.pReprs = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p)), sizeof(Aig_Obj_t *))); }
    // create the PIs
    Aig_ManCleanData( p );
    // duplicate internal nodes
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsCi(pObj) {
            pObjNew = Aig_ObjCreateCi( pNew );
            pObjNew.Level = pObj.Level;
            pObj.pData = pObjNew;
        
        } else if Aig_ObjIsCo(pObj) {
            Aig_ManDupDfs_rec( pNew, p, Aig_ObjFanin0(pObj) );        
//            assert( pObj->Level == ((Aig_Obj_t*)pObj->pData)->Level );
            pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
            pObj.pData = pObjNew;
        }
    }
    }
    assert(p.pEquivs != nil || Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew), "p.pEquivs != nil || Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew)");
    if p.pEquivs == nil && p.pReprs == nil && (nNodes = Aig_ManCleanup( pNew )) {
        fmt.Printf("Aig_ManDupDfs(): Cleanup after AIG duplication removed %d nodes.\n", nNodes); }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // duplicate the timing manager
    if p.pManTime {
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)p.pManTime, 0 ); }
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupDfs(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [This duplicator works for AIGs with choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManOrderPios(p *Aig_Man_t, pOrder *Aig_Man_t) *Vec_Ptr_t {
    var vPios *Vec_Ptr_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    assert(Aig_ManCiNum(p) == Aig_ManCiNum(pOrder), "Aig_ManCiNum(p) == Aig_ManCiNum(pOrder)");
    assert(Aig_ManCoNum(p) == Aig_ManCoNum(pOrder), "Aig_ManCoNum(p) == Aig_ManCoNum(pOrder)");
    Aig_ManSetCioIds( pOrder );
    vPios = Vec_PtrAlloc( Aig_ManCiNum(p) + Aig_ManCoNum(p) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pOrder.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pOrder.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsCi(pObj) {
            Vec_PtrPush( vPios, Aig_ManCi(p, Aig_ObjCioId(pObj)) );
        } else if Aig_ObjIsCo(pObj) {
            Vec_PtrPush( vPios, Aig_ManCo(p, Aig_ObjCioId(pObj)) ); }
    }
    }
    Aig_ManCleanCioIds( pOrder );
    return vPios;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupDfsGuided_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pObjNew *Aig_Obj_t;
    var pEquivNew *Aig_Obj_t = nil;
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    if Aig_ObjIsCi(pObj) {
        return nil; }
    if p.pEquivs && Aig_ObjEquiv(p, pObj) {
        pEquivNew = Aig_ManDupDfsGuided_rec( pNew, p, Aig_ObjEquiv(p, pObj) ); }
    if !(Aig_ManDupDfsGuided_rec( pNew, p, Aig_ObjFanin0(pObj) ) != nil) {
        return nil; }
    if Aig_ObjIsBuf(pObj) {
        return (Aig_Obj_t *)(pObj.pData = Aig_ObjChild0Copy(pObj)); }
    if !(Aig_ManDupDfsGuided_rec( pNew, p, Aig_ObjFanin1(pObj) ) != nil) {
        return nil; }
    pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
    if pEquivNew != nil {
        if pNew.pEquivs {
            pNew.pEquivs[Aig_Regular(pObjNew).Id] = Aig_Regular(pEquivNew); }        
        if pNew.pReprs {
            pNew.pReprs[Aig_Regular(pEquivNew).Id] = Aig_Regular(pObjNew); }
    }
    return (Aig_Obj_t *)(pObj.pData = pObjNew);
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [This duplicator works for AIGs with choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupDfsGuided(p *Aig_Man_t, vPios *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i, nNodes int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // duplicate representation of choice nodes
    if p.pEquivs {
        // inlining of macro ABC_ALLOC
        pNew.pEquivs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManObjNumMax(p))));
        memset( pNew.pEquivs, 0, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(p) );
    }
    if p.pReprs {
        // inlining of macro ABC_ALLOC
        pNew.pReprs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManObjNumMax(p))));
        memset( pNew.pReprs, 0, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(p) );
    }
    // create the PIs
    Aig_ManCleanData( p );
    // duplicate internal nodes
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPios)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vPios, i)), 1); i++  {
        if Aig_ObjIsCi(pObj) {
            pObjNew = Aig_ObjCreateCi( pNew );
            pObjNew.Level = pObj.Level;
            pObj.pData = pObjNew;
        
        } else if Aig_ObjIsCo(pObj) {
            Aig_ManDupDfsGuided_rec( pNew, p, Aig_ObjFanin0(pObj) );        
//            assert( pObj->Level == ((Aig_Obj_t*)pObj->pData)->Level );
            pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
            pObj.pData = pObjNew;
        }
    }
//    assert( Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew) );
    if p.pEquivs == nil && p.pReprs == nil && (nNodes = Aig_ManCleanup( pNew )) {
        fmt.Printf("Aig_ManDupDfs(): Cleanup after AIG duplication removed %d nodes.\n", nNodes); }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // duplicate the timing manager
    if p.pManTime {
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)p.pManTime, 0 ); }
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupDfs(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [This duplicator works for AIGs with choices.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupLevelized(p *Aig_Man_t) *Aig_Man_t {
    var vLevels *Vec_Vec_t;
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjNew *Aig_Obj_t = nil;
    var i, k int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // duplicate representation of choice nodes
    if p.pEquivs {
        // inlining of macro ABC_ALLOC
        pNew.pEquivs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManObjNumMax(p))));
        memset( pNew.pEquivs, 0, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(p) );
    }
    if p.pReprs {
        // inlining of macro ABC_ALLOC
        pNew.pReprs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManObjNumMax(p))));
        memset( pNew.pReprs, 0, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(p) );
    }
    // create the PIs
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCi( pNew );
        pObjNew.Level = pObj.Level;
        pObj.pData = pObjNew;
    }
    // duplicate internal nodes
    vLevels = Aig_ManLevelize( p );
    // inlining of macro Vec_VecForEachEntry
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; i < Vec_VecSize(vLevels); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vLevels, i))) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vLevels, i), k)), 1); k++  {
        pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
        pObj.pData = pObjNew;
    }
    }
    Vec_VecFree( vLevels );
    // duplicate POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        pObj.pData = pObjNew;
    }
    assert(Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew), "Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew)");
//    if ( (nNodes = Aig_ManCleanup( pNew )) )
//        printf( "Aig_ManDupLevelized(): Cleanup after AIG duplication removed %d nodes.\n", nNodes );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // duplicate the timing manager
    if p.pManTime {
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)p.pManTime, 0 ); }
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupLevelized(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [Assumes topological ordering of nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupWithoutPos(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!Aig_ObjIsBuf(pObj), "!Aig_ObjIsBuf(pObj)");
        if Aig_ObjIsNode(pObj) {
            pObj.pData = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) ); }
    }
    }
    assert(Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew), "Aig_ManBufNum(p) != 0 || Aig_ManNodeNum(p) == Aig_ManNodeNum(pNew)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [Assumes topological ordering of nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupFlopsOnly(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    pNew = Aig_ManDupWithoutPos( p );
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupFlopsOnly(): The check has failed.\n"); }
    return pNew;

}


/**Function*************************************************************

  Synopsis    [Returns representatives of fanin in approapriate polarity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjGetRepres(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pRepr *Aig_Obj_t;
    if pRepr = Aig_ObjRepr(p, pObj) {
        return Aig_NotCond( (Aig_Obj_t *)pRepr.pData, pObj.fPhase ^ pRepr.fPhase ); }
    return (Aig_Obj_t *)pObj.pData;
}
func Aig_ObjChild0Repres(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond( Aig_ObjGetRepres(p, Aig_ObjFanin0(pObj)), Aig_ObjFaninC0(pObj) ); }
func Aig_ObjChild1Repres(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond( Aig_ObjGetRepres(p, Aig_ObjFanin1(pObj)), Aig_ObjFaninC1(pObj) ); }

/**Function*************************************************************

  Synopsis    [Duplicates AIG while substituting representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupRepres(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    // start the HOP package
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // map the const and primary inputs
    Aig_ManCleanData( p );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsNode(pObj) {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Repres(p, pObj), Aig_ObjChild1Repres(p, pObj) );
        } else if Aig_ObjIsCi(pObj) {
            pObj.pData = Aig_ObjCreateCi(pNew);
            pObj.pData = Aig_ObjGetRepres( p, pObj );
        
        } else if Aig_ObjIsCo(pObj) {
            pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Repres(p, pObj) );
        } else if Aig_ObjIsConst1(pObj) {
            pObj.pData = Aig_ManConst1(pNew);
        } else {
            assert(false, "0"); }
    }
    }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the new manager
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupRepres: Check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupRepres_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pRepr *Aig_Obj_t;
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    if pRepr = Aig_ObjRepr(p, pObj) {
        Aig_ManDupRepres_rec( pNew, p, pRepr );
        return (Aig_Obj_t *)(pObj.pData = Aig_NotCond( (Aig_Obj_t *)pRepr.pData, pRepr.fPhase ^ pObj.fPhase ));
    }
    Aig_ManDupRepres_rec( pNew, p, Aig_ObjFanin0(pObj) );
    Aig_ManDupRepres_rec( pNew, p, Aig_ObjFanin1(pObj) );
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, Aig_ObjChild0Repres(p, pObj), Aig_ObjChild1Repres(p, pObj) ));
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while substituting representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupRepresDfs(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    // start the HOP package
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // map the const and primary inputs
    Aig_ManCleanData( p );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsNode(pObj) {
            continue; }
        if Aig_ObjIsCi(pObj) {
            pObj.pData = Aig_ObjCreateCi(pNew);
        } else if Aig_ObjIsCo(pObj) {
            Aig_ManDupRepres_rec( pNew, p, Aig_ObjFanin0(pObj) );
            pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Repres(p, pObj) );
        
        } else if Aig_ObjIsConst1(pObj) {
            pObj.pData = Aig_ManConst1(pNew);
        } else { 
            assert(false, "0"); }
    }
    }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the new manager
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupRepresDfs: Check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Creates the miter of the two AIG managers.]

  Description [Oper is the operation to perform on the outputs of the miter.
  Oper == 0 is XOR
  Oper == 1 is complemented implication (p1 => p2)
  Oper == 2 is OR
  Oper == 3 is AND
  ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCreateMiter(p1 *Aig_Man_t, p2 *Aig_Man_t, Oper int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    assert(Aig_ManRegNum(p1) == 0, "Aig_ManRegNum(p1) == 0");
    assert(Aig_ManRegNum(p2) == 0, "Aig_ManRegNum(p2) == 0");
    assert(Aig_ManCoNum(p1) == 1, "Aig_ManCoNum(p1) == 1");
    assert(Aig_ManCoNum(p2) == 1, "Aig_ManCoNum(p2) == 1");
    assert(Aig_ManCiNum(p1) == Aig_ManCiNum(p2), "Aig_ManCiNum(p1) == Aig_ManCiNum(p2)");
    pNew = Aig_ManStart( Aig_ManObjNumMax(p1) + Aig_ManObjNumMax(p2) );
    // add first AIG
    Aig_ManConst1(p1).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p1.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p1.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p1.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p1.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add second AIG
    Aig_ManConst1(p2).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p2.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p2.vCis, i)), 1); i++  {
        pObj.pData = Aig_ManCi( pNew, i );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p2.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p2.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add the output
    if Oper == 0 { // XOR
        pObj = Aig_Exor( pNew, Aig_ObjChild0Copy(Aig_ManCo(p1,0)), Aig_ObjChild0Copy(Aig_ManCo(p2,0)) );
    } else if Oper == 1 { // implication is PO(p1) -> PO(p2)  ...  complement is PO(p1) & !PO(p2) 
        pObj = Aig_And( pNew, Aig_ObjChild0Copy(Aig_ManCo(p1,0)), Aig_Not(Aig_ObjChild0Copy(Aig_ManCo(p2,0))) );
    } else if Oper == 2 { // OR
        pObj = Aig_Or( pNew, Aig_ObjChild0Copy(Aig_ManCo(p1,0)), Aig_ObjChild0Copy(Aig_ManCo(p2,0)) );
    } else if Oper == 3 { // AND
        pObj = Aig_And( pNew, Aig_ObjChild0Copy(Aig_ManCo(p1,0)), Aig_ObjChild0Copy(Aig_ManCo(p2,0)) );
    } else {
        assert(false, "0"); }
    Aig_ObjCreateCo( pNew, pObj );
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG with only one primary output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupOrpos(p *Aig_Man_t, fAddRegs int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pMiter *Aig_Obj_t;
    var i int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    if p.nConstrs > 0 {
        fmt.Printf("The AIG manager should have no constraints.\n");
        return nil;
    }
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // set registers
    if fAddRegs != 0 {
        pNew.nRegs = p.nRegs;
    } else {
        pNew.nRegs = 0;
    }
    if fAddRegs != 0 {
        pNew.nTruePis = p.nTruePis;
    } else {
        pNew.nTruePis = p.nTruePis + p.nRegs;
    }
    pNew.nTruePos = 1;
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create the PO
    pMiter = Aig_ManConst0(pNew);
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pMiter = Aig_Or( pNew, pMiter, Aig_ObjChild0Copy(pObj) );
    } 
    Aig_ObjCreateCo( pNew, pMiter );
    // create register inputs with MUXes
    if fAddRegs != 0 {
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        }
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG with only one primary output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupOneOutput(p *Aig_Man_t, iPoNum int32, fAddRegs int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    assert(iPoNum < Aig_ManCoNum(p)-Aig_ManRegNum(p), "iPoNum < Aig_ManCoNum(p)-Aig_ManRegNum(p)");
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // set registers
    if fAddRegs != 0 {
        pNew.nRegs = p.nRegs;
    } else {
        pNew.nRegs = 0;
    }
    if fAddRegs != 0 {
        pNew.nTruePis = p.nTruePis;
    } else {
        pNew.nTruePis = p.nTruePis + p.nRegs;
    }
    pNew.nTruePos = 1;
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create the PO
    pObj = Aig_ManCo( p, iPoNum );
    Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    // create register inputs with MUXes
    if fAddRegs != 0 {
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        }
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG with only one primary output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupUnsolvedOutputs(p *Aig_Man_t, fAddRegs int32) *Aig_Man_t { 
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i, nOuts int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    if p.nConstrs > 0 {
        fmt.Printf("The AIG manager should have no constraints.\n");
        return nil;
    }
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // create the POs
    nOuts = 0;
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        nOuts += ( Aig_ObjFanin0(pObj) != Aig_ManConst1(p) );
    }
    // set registers
    if fAddRegs != 0 {
        pNew.nRegs = p.nRegs;
    } else {
        pNew.nRegs = 0;
    }
    if fAddRegs != 0 {
        pNew.nTruePis = p.nTruePis;
    } else {
        pNew.nTruePis = p.nTruePis + p.nRegs;
    }
    pNew.nTruePos = nOuts;
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create the PO
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if Aig_ObjFanin0(pObj) != Aig_ManConst1(p) {
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) ); }
    }
    // create register inputs with MUXes
    if fAddRegs != 0 {
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        }
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG with only one primary output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupArray(vArray *Vec_Ptr_t) *Aig_Man_t {
    var p, pNew *Aig_Man_t;
    var pObj *Aig_Obj_t  = nil;
    var i, k int32;
    if Vec_PtrSize(vArray) == 0 {
        return nil; }
    p = (Aig_Man_t *)Vec_PtrEntry( vArray, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vArray)) && (((pNew) = (Aig_Man_t *)Vec_PtrEntry(vArray, k)), 1); k++  {
        assert(Aig_ManRegNum(pNew) == 0, "Aig_ManRegNum(pNew) == 0");
        assert(Aig_ManCiNum(pNew) == Aig_ManCiNum(p), "Aig_ManCiNum(pNew) == Aig_ManCiNum(p)");
    }
    // create the new manager
    pNew = Aig_ManStart( 10000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Aig_ObjCreateCi(pNew);
    }
    // create the PIs
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vArray)) && (((p) = (Aig_Man_t *)Vec_PtrEntry(vArray, k)), 1); k++  {
        Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pObj.pData = Aig_ManCi( pNew, i );
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        }
    }
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManDupSimple(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG with only one primary output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDupNodes(pMan *Aig_Man_t, vArray *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var vObjs *Vec_Ptr_t;
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    if Vec_PtrSize(vArray) == 0 {
        return nil; }
    vObjs = Aig_ManDfsNodes( pMan, (Aig_Obj_t **)Vec_PtrArray(vArray), Vec_PtrSize(vArray) );
    // create the new manager
    pNew = Aig_ManStart( 10000 );
    pNew.pName = Abc_UtilStrsav( pMan.pName );
    Aig_ManConst1(pMan).pData = Aig_ManConst1(pNew);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        if Aig_ObjIsCi(pObj) {
            pObj.pData = Aig_ObjCreateCi(pNew); }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        if Aig_ObjIsNode(pObj) {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vArray)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vArray, i)), 1); i++  {
            Aig_ObjCreateCo( pNew, (Aig_Obj_t *)pObj.pData );
    }
    Aig_ManSetRegNum( pNew, 0 );
    Vec_PtrFree( vObjs );
    return pNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

