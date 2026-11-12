package aig

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [aigMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [AIG manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigMan.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the AIG manager.]

  Description [The argument of this procedure is a soft limit on the
  the number of nodes, or 0 if the limit is unknown.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManStart(nNodesMax int32) *Aig_Man_t {
    var p *Aig_Man_t;
    if nNodesMax <= 0 {
        nNodesMax = 10007; }
    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Aig_Man_t *) malloc(sizeof(Aig_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Aig_Man_t) );
    // perform initializations
    p.nTravIds = 1;
    p.fCatchExor = 0;
    // allocate arrays for nodes
    p.vCis  = Vec_PtrAlloc( 100 );
    p.vCos  = Vec_PtrAlloc( 100 );
    p.vObjs = Vec_PtrAlloc( 1000 );
    p.vBufs = Vec_PtrAlloc( 100 );
    //--jlong -- begin
       p.unfold2_type_I = Vec_PtrAlloc( 4);
       p.unfold2_type_II = Vec_PtrAlloc( 4);
       //--jlong -- end
    // prepare the internal memory manager
    p.pMemObjs = Aig_MmFixedStart( sizeof(Aig_Obj_t), nNodesMax );
    // create the constant node
    p.pConst1 = Aig_ManFetchMemory( p );
    p.pConst1.Type = AIG_OBJ_CONST1;
    p.pConst1.fPhase = 1;
    p.nObjs[AIG_OBJ_CONST1]++;
    // start the table
    p.nTableSize = Abc_PrimeCudd( nNodesMax );
    // inlining of macro ABC_ALLOC
    p.pTable = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nTableSize)));
    memset( p.pTable, 0, sizeof(Aig_Obj_t *) * p.nTableSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManStartFrom(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var i int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObjNew = Aig_ObjCreateCi( pNew );
        pObjNew.Level = pObj.Level;
        pObj.pData = pObjNew;
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDup_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pObjNew *Aig_Obj_t;
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    Aig_ManDup_rec( pNew, p, Aig_ObjFanin0(pObj) );
    if Aig_ObjIsBuf(pObj) {
        return (Aig_Obj_t *)(pObj.pData = Aig_ObjChild0Copy(pObj)); }
    Aig_ManDup_rec( pNew, p, Aig_ObjFanin1(pObj) );
    pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
    return (Aig_Obj_t *)(pObj.pData = pObjNew);
}

/**Function*************************************************************

  Synopsis    [Extracts the miter composed of XOR of the two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManExtractMiter(p *Aig_Man_t, pNode1 *Aig_Obj_t, pNode2 *Aig_Obj_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
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
        pObj.pData = Aig_ObjCreateCi(pNew);
    }
    // dump the nodes
    Aig_ManDup_rec( pNew, p, pNode1 );   
    Aig_ManDup_rec( pNew, p, pNode2 );   
    // construct the EXOR
    pObj = Aig_Exor( pNew, (Aig_Obj_t *)pNode1.pData, (Aig_Obj_t *)pNode2.pData ); 
    pObj = Aig_NotCond( pObj, Aig_Regular(pObj).fPhase ^ Aig_IsComplement(pObj) );
    // add the PO
    Aig_ObjCreateCo( pNew, pObj );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Aig_ManExtractMiter(): The check has failed.\n"); }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManStop(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro ABC_PRT
    if p.time1 { Abc_Print(1, "%s =", ("time1"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time1)))/(float64(int64(time.Second))));; }
    // inlining of macro ABC_PRT
    if p.time2 { Abc_Print(1, "%s =", ("time2"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time2)))/(float64(int64(time.Second))));; }
    // make sure the nodes have clean marks
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!pObj.fMarkA && !pObj.fMarkB, "!pObj.fMarkA && !pObj.fMarkB"); }
    }
    Tim_ManStopP( (Tim_Man_t **)&p.pManTime );
    if p.pFanData { 
        Aig_ManFanoutStop( p ); }
    if p.pManExdc {  
        Aig_ManStop( p.pManExdc ); }
//    Aig_TableProfile( p );
    Aig_MmFixedStop( p.pMemObjs, 0 );
    Vec_PtrFreeP( &p.vCis );
    Vec_PtrFreeP( &p.vCos );
    Vec_PtrFreeP( &p.vObjs );
    Vec_PtrFreeP( &p.vBufs );
    //--jlong -- begin
    Vec_PtrFreeP( &p.unfold2_type_I );
    Vec_PtrFreeP( &p.unfold2_type_II );
    //--jlong -- end
    Vec_IntFreeP( &p.vLevelR );
    Vec_VecFreeP( &p.vLevels );
    Vec_IntFreeP( &p.vFlopNums );
    Vec_IntFreeP( &p.vFlopReprs );
    Vec_VecFreeP( (Vec_Vec_t **)&p.vOnehots );
    Vec_VecFreeP( &p.vClockDoms );
    Vec_IntFreeP( &p.vProbs );
    Vec_IntFreeP( &p.vCiNumsOrig );
    Vec_PtrFreeP( &p.vMapped );
    if p.vSeqModelVec {
        Vec_PtrFreeFree( p.vSeqModelVec ); }
    // inlining of macro ABC_FREE
    ((p.pTerSimData) ? (free((char *) (p.pTerSimData)), (p.pTerSimData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pFastSim) ? (free((char *) (p.pFastSim)), (p.pFastSim) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pData) ? (free((char *) (p.pData)), (p.pData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSeqModel) ? (free((char *) (p.pSeqModel)), (p.pSeqModel) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pName) ? (free((char *) (p.pName)), (p.pName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSpec) ? (free((char *) (p.pSpec)), (p.pSpec) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pObjCopies) ? (free((char *) (p.pObjCopies)), (p.pObjCopies) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pReprs) ? (free((char *) (p.pReprs)), (p.pReprs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pEquivs) ? (free((char *) (p.pEquivs)), (p.pEquivs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTable) ? (free((char *) (p.pTable)), (p.pTable) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManStopP(p **Aig_Man_t) {
    if *p == nil {
        return; }
    Aig_ManStop( *p );
    *p = nil;
}

/**Function*************************************************************

  Synopsis    [Removes combinational logic that does not feed into POs.]

  Description [Returns the number of dangling nodes removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCleanup(p *Aig_Man_t) int32 {
    var vObjs *Vec_Ptr_t;
    var pNode *Aig_Obj_t;
    var i int32;
    var nNodesOld int32 = Aig_ManNodeNum(p);
    // collect roots of dangling nodes
    vObjs = Vec_PtrAlloc( 100 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pNode) == nil {} else
        if Aig_ObjIsNode(pNode) && Aig_ObjRefs(pNode) == 0 {
            Vec_PtrPush( vObjs, pNode ); }
    }
    // recursively remove dangling nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        Aig_ObjDelete_rec( p, pNode, 1 );
    }
    Vec_PtrFree( vObjs );
    return nNodesOld - Aig_ManNodeNum(p);
}

/**Function*************************************************************

  Synopsis    [Adds POs for the nodes that otherwise would be dangling.]

  Description [Returns the number of POs added.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManAntiCleanup(p *Aig_Man_t) int32 {
    var pNode *Aig_Obj_t;
    var i int32;
    var nNodesOld int32 = Aig_ManCoNum(p);
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pNode) == nil {} else
        if Aig_ObjIsNode(pNode) && Aig_ObjRefs(pNode) == 0 {
            Aig_ObjCreateCo( p, pNode ); }
    }
    return nNodesOld - Aig_ManCoNum(p);
}

/**Function*************************************************************

  Synopsis    [Removes PIs without fanouts.]

  Description [Returns the number of PIs removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCiCleanup(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    var k int32;
    var nPisOld int32 = Aig_ManCiNum(p);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if i >= Aig_ManCiNum(p) - Aig_ManRegNum(p) {
            Vec_PtrWriteEntry( p.vCis, k++, pObj );
        } else if Aig_ObjRefs(pObj) > 0 {
            Vec_PtrWriteEntry( p.vCis, k++, pObj );
        } else {
            Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil ); }
    }
    Vec_PtrShrink( p.vCis, k );
    p.nObjs[AIG_OBJ_CI] = Vec_PtrSize( p.vCis );
    if Aig_ManRegNum(p) {
        p.nTruePis = Aig_ManCiNum(p) - Aig_ManRegNum(p); }
    return nPisOld - Aig_ManCiNum(p);
}

/**Function*************************************************************

  Synopsis    [Removes POs with constant input.]

  Description [Returns the number of POs removed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCoCleanup(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    var k int32;
    var nPosOld int32 = Aig_ManCoNum(p);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if i >= Aig_ManCoNum(p) - Aig_ManRegNum(p) {
            Vec_PtrWriteEntry( p.vCos, k++, pObj );
        } else if !Aig_ObjIsConst1(Aig_ObjFanin0(pObj)) || !Aig_ObjFaninC0(pObj) { // non-const or const1
            Vec_PtrWriteEntry( p.vCos, k++, pObj );
        } else {
            Aig_ObjDisconnect( p, pObj );
            Vec_PtrWriteEntry( p.vObjs, pObj.Id, nil );
        }
    }
    Vec_PtrShrink( p.vCos, k );
    p.nObjs[AIG_OBJ_CO] = Vec_PtrSize( p.vCos );
    if Aig_ManRegNum(p) {
        p.nTruePos = Aig_ManCoNum(p) - Aig_ManRegNum(p); }
    return nPosOld - Aig_ManCoNum(p);
}

/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPrintStats(p *Aig_Man_t) {
    var nChoices int32  = Aig_ManChoiceNum(p);
    fmt.Printf("%-15s : ", p.pName);
    fmt.Printf("pi = %5d  ", Aig_ManCiNum(p)-Aig_ManRegNum(p));
    fmt.Printf("po = %5d  ", Aig_ManCoNum(p)-Aig_ManRegNum(p));
    if Aig_ManRegNum(p) {
    fmt.Printf("lat = %5d  ", Aig_ManRegNum(p)); }
    fmt.Printf("and = %7d  ", Aig_ManAndNum(p));
//    printf( "Eq = %7d  ",     Aig_ManHaigCounter(p) );
    if Aig_ManExorNum(p) {
    fmt.Printf("xor = %5d  ", Aig_ManExorNum(p)); }
    if nChoices != 0 {
    fmt.Printf("ch = %5d  ", nChoices); }
    if Aig_ManBufNum(p) {
    fmt.Printf("buf = %5d  ", Aig_ManBufNum(p)); }
//    printf( "Cre = %6d  ",  p->nCreated );
//    printf( "Del = %6d  ",  p->nDeleted );
//    printf( "Lev = %3d  ",  Aig_ManLevelNum(p) );
//    printf( "Max = %7d  ",  Aig_ManObjNumMax(p) );
    fmt.Printf("lev = %3d", Aig_ManLevels(p));
    fmt.Printf("\n");
    fflush( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Reports the reduction of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManReportImprovement(p *Aig_Man_t, pNew *Aig_Man_t) {
    fmt.Printf("REG: Beg = %5d. End = %5d. (R =%5.1f %%)  ", Aig_ManRegNum(p), Aig_ManRegNum(pNew), Aig_ManRegNum(p)? 100.0*(Aig_ManRegNum(p)-Aig_ManRegNum(pNew))/Aig_ManRegNum(p) : 0.0);
    fmt.Printf("AND: Beg = %6d. End = %6d. (R =%5.1f %%)", Aig_ManNodeNum(p), Aig_ManNodeNum(pNew), Aig_ManNodeNum(p)? 100.0*(Aig_ManNodeNum(p)-Aig_ManNodeNum(pNew))/Aig_ManNodeNum(p) : 0.0);
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Sets the number of registers in the AIG manager.]

  Description [This procedure should be called after the manager is 
  fully constructed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSetRegNum(p *Aig_Man_t, nRegs int32) {
    p.nRegs = nRegs;
    p.nTruePis = Aig_ManCiNum(p) - nRegs;
    p.nTruePos = Aig_ManCoNum(p) - nRegs;
    Aig_ManSetCioIds( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFlipFirstPo(p *Aig_Man_t) {
    Aig_ObjChild0Flip( Aig_ManCo(p, 0) ); 
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManReleaseData(p *Aig_Man_t) any { 
    var pD any  = p.pData; 
    p.pData = nil; 
    return pD;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

