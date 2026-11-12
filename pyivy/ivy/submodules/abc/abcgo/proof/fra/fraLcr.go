package fra

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [fraLcorr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Latch correspondence computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraLcorr.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

      
type Fra_Lcr_t struct {
    // original AIG
    pAig       *Aig_Man_t;
    // equivalence class representation
    pCla       *Fra_Cla_t;       
    // partitioning information
    vParts       *Vec_Ptr_t;        // output partitions
    pInToOutPart             *int32;  // mapping of PI num into PO partition num
    pInToOutNum             *int32;   // mapping of PI num into the num of this PO in the partition
    // AIGs for the partitions
    vFraigs       *Vec_Ptr_t;
    // other variables
    fRefining              int32; 
    // parameters
    nFramesP              int32;
    fVerbose              int32;
    // statistics
    nIters              int32;
    nLitsBeg              int32;
    nLitsEnd              int32;
    nNodesBeg              int32;
    nNodesEnd              int32;
    nRegsBeg              int32;
    nRegsEnd              int32;
    // runtime
    timeSim          abctime;
    timePart          abctime;
    timeTrav          abctime;
    timeFraig          abctime;
    timeUpdate          abctime;
    timeTotal          abctime;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the retiming manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lcr_ManAlloc(pAig *Aig_Man_t) *Fra_Lcr_t {
    var p *Fra_Lcr_t;
    // inlining of macro ABC_ALLOC
    p = ((Fra_Lcr_t *) malloc(sizeof(Fra_Lcr_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fra_Lcr_t) );
    p.pAig = pAig;
    // inlining of macro ABC_ALLOC
    p.pInToOutPart = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManCiNum(pAig))));
    memset( p.pInToOutPart, 0, sizeof(int) * Aig_ManCiNum(pAig) );
    // inlining of macro ABC_ALLOC
    p.pInToOutNum = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManCiNum(pAig))));
    memset( p.pInToOutNum, 0, sizeof(int) * Aig_ManCiNum(pAig) );
    p.vFraigs = Vec_PtrAlloc( 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Prints stats for the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lcr_ManPrint(p *Fra_Lcr_t) {
    fmt.Printf("Iterations = %d.  LitBeg = %d.  LitEnd = %d. (%6.2f %%).\n", p.nIters, p.nLitsBeg, p.nLitsEnd, 100.0*p.nLitsEnd/p.nLitsBeg);
    fmt.Printf("NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", p.nNodesBeg, p.nNodesEnd, 100.0*(p.nNodesBeg-p.nNodesEnd)/p.nNodesBeg, p.nRegsBeg, p.nRegsEnd, 100.0*(p.nRegsBeg-p.nRegsEnd)/p.nRegsBeg);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG simulation  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG partitioning"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timePart)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG rebuiding   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTrav)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("FRAIGing        "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeFraig)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("AIG updating    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeUpdate)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL RUNTIME   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Deallocates the retiming manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lcr_ManFree(p *Fra_Lcr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    if p.fVerbose {
        Lcr_ManPrint( p ); }
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pObj.pNext = nil;
    }
    Vec_PtrFree( p.vFraigs );
    if p.pCla {     Fra_ClassesStop( p.pCla ); }
    if p.vParts {   Vec_VecFree( (Vec_Vec_t *)p.vParts ); }
    // inlining of macro ABC_FREE
    ((p.pInToOutPart) ? (free((char *) (p.pInToOutPart)), (p.pInToOutPart) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pInToOutNum) ? (free((char *) (p.pInToOutNum)), (p.pInToOutNum) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prepare the AIG for class computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrAigPrepare(pAig *Aig_Man_t) *Fra_Man_t {
    var p *Fra_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro ABC_ALLOC
    p = ((Fra_Man_t *) malloc(sizeof(Fra_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fra_Man_t) );
//    Aig_ManForEachCi( pAig, pObj, i )
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = p; }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Prepare the AIG for class computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrAigPrepareTwo(pAig *Aig_Man_t, p *Fra_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = p;
    }
}

/**Function*************************************************************

  Synopsis    [Compares two nodes for equivalence after partitioned fraiging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrNodesAreEqual(pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) int32 {
    var pTemp *Fra_Man_t  = (Fra_Man_t *)pObj0.pData;
    var pLcr *Fra_Lcr_t  = (Fra_Lcr_t *)pTemp.pBmc;
    var pFraig *Aig_Man_t;
    var pOut0, pOut1 *Aig_Obj_t;
    var nPart0, nPart1 int32;
    assert(Aig_ObjIsCi(pObj0), "Aig_ObjIsCi(pObj0)");
    assert(Aig_ObjIsCi(pObj1), "Aig_ObjIsCi(pObj1)");
    // find the partition to which these nodes belong
    nPart0 = pLcr.pInToOutPart[int64(pObj0.pNext)];
    nPart1 = pLcr.pInToOutPart[int64(pObj1.pNext)];
    // if this is the result of refinement of the class created const-1 nodes
    // the nodes may end up in different partions - we assume them equivalent
    if nPart0 != nPart1 {
        assert(false, "0");
        return 1;
    }
    assert(nPart0 == nPart1, "nPart0 == nPart1");
    pFraig = (Aig_Man_t *)Vec_PtrEntry( pLcr.vFraigs, nPart0 );
    // get the fraig outputs
    pOut0 = Aig_ManCo( pFraig, pLcr.pInToOutNum[int64(pObj0.pNext)] );
    pOut1 = Aig_ManCo( pFraig, pLcr.pInToOutNum[int64(pObj1.pNext)] );
    return Aig_ObjFanin0(pOut0) == Aig_ObjFanin0(pOut1);
}

/**Function*************************************************************

  Synopsis    [Compares the node with a constant after partioned fraiging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrNodeIsConst(pObj *Aig_Obj_t) int32 {
    var pTemp *Fra_Man_t  = (Fra_Man_t *)pObj.pData;
    var pLcr *Fra_Lcr_t  = (Fra_Lcr_t *)pTemp.pBmc;
    var pFraig *Aig_Man_t;
    var pOut *Aig_Obj_t;
    var nPart int32;
    assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
    // find the partition to which these nodes belong
    nPart = pLcr.pInToOutPart[int64(pObj.pNext)];
    pFraig = (Aig_Man_t *)Vec_PtrEntry( pLcr.vFraigs, nPart );
    // get the fraig outputs
    pOut = Aig_ManCo( pFraig, pLcr.pInToOutNum[int64(pObj.pNext)] );
    return Aig_ObjFanin0(pOut) == Aig_ManConst1(pFraig);
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrManDup_rec(pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    var pObjNew *Aig_Obj_t;
    if pObj.pData {
        return (Aig_Obj_t *)pObj.pData; }
    Fra_LcrManDup_rec( pNew, p, Aig_ObjFanin0(pObj) );
    if Aig_ObjIsBuf(pObj) {
        return (Aig_Obj_t *)(pObj.pData = Aig_ObjChild0Copy(pObj)); }
    Fra_LcrManDup_rec( pNew, p, Aig_ObjFanin1(pObj) );
    pObjNew = Aig_Oper( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj), Aig_ObjType(pObj) );
    return (Aig_Obj_t *)(pObj.pData = pObjNew);
}

/**Function*************************************************************

  Synopsis    [Give the AIG and classes, reduces AIG for partitioning.]

  Description [Ignores registers that are not in the classes. 
  Places candidate equivalent classes of registers into single outputs 
  (for ease of partitioning). The resulting combinational AIG contains 
  outputs in the same order as equivalence classes of registers, 
  followed by constant-1 registers. Preserves the set of all inputs.
  Complemented attributes of the outputs do not matter because we need 
  then only for collecting the structural info.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrDeriveAigForPartitioning(pLcr *Fra_Lcr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pObjPo *Aig_Obj_t;
    var pObjNew *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var pMiter *Aig_Obj_t;
    var i, c, Offset int32;
    // remember the numbers of the inputs of the original AIG
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pLcr.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pLcr.pAig.vCis, i)), 1); i++  {
        pObj.pData = pLcr;
        pObj.pNext = (Aig_Obj_t *)int64(i);
    }
    // compute the LO/LI offset
    Offset = Aig_ManCoNum(pLcr.pAig) - Aig_ManCiNum(pLcr.pAig);
    // create the PIs
    Aig_ManCleanData( pLcr.pAig );
    pNew = Aig_ManStartFrom( pLcr.pAig );
    // go over the equivalence classes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pLcr.pCla.vClasses)) && (((ppClass) = (Aig_Obj_t **)Vec_PtrEntry(pLcr.pCla.vClasses, i)), 1); i++  {
        pMiter = Aig_ManConst0(pNew);
        for  c = 0; ppClass[c]; c++  {
            assert(Aig_ObjIsCi(ppClass[c]), "Aig_ObjIsCi(ppClass[c])");
            pObjPo  = Aig_ManCo( pLcr.pAig, Offset+int64(ppClass[c].pNext) );
            pObjNew = Fra_LcrManDup_rec( pNew, pLcr.pAig, Aig_ObjFanin0(pObjPo) );
            pMiter  = Aig_Exor( pNew, pMiter, pObjNew );
        }
        Aig_ObjCreateCo( pNew, pMiter );
    }
    // go over the constant candidates
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pLcr.pCla.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pLcr.pCla.vClasses1, i)), 1); i++  {
        assert(Aig_ObjIsCi(pObj), "Aig_ObjIsCi(pObj)");
        pObjPo = Aig_ManCo( pLcr.pAig, Offset+int64(pObj.pNext) );
        pMiter = Fra_LcrManDup_rec( pNew, pLcr.pAig, Aig_ObjFanin0(pObjPo) );
        Aig_ObjCreateCo( pNew, pMiter );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Remaps partitions into the inputs of original AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrRemapPartitions(vParts *Vec_Ptr_t, pCla *Fra_Cla_t, pInToOutPart *int32, pInToOutNum *int32) {
    var vOne, vOneNew *Vec_Int_t;
    var ppClass **Aig_Obj_t;
    var pObjPi *Aig_Obj_t;
    var Out, Offset, i, k, c int32;
    // compute the LO/LI offset
    Offset = Aig_ManCoNum(pCla.pAig) - Aig_ManCiNum(pCla.pAig);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vOne) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        vOneNew = Vec_IntAlloc( Vec_IntSize(vOne) );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vOne)) && (((Out) = Vec_IntEntry(vOne, k)), 1); k++  {
            if Out < Vec_PtrSize(pCla.vClasses) {
                ppClass = (Aig_Obj_t **)Vec_PtrEntry( pCla.vClasses, Out );
                for  c = 0; ppClass[c]; c++  {
                    pInToOutPart[int64(ppClass[c].pNext)] = i;
                    pInToOutNum[int64(ppClass[c].pNext)] = Vec_IntSize(vOneNew);
                    Vec_IntPush( vOneNew, Offset+int64(ppClass[c].pNext) );
                }
            
            } else {
                pObjPi = (Aig_Obj_t *)Vec_PtrEntry( pCla.vClasses1, Out - Vec_PtrSize(pCla.vClasses) );
                pInToOutPart[int64(pObjPi.pNext)] = i;
                pInToOutNum[int64(pObjPi.pNext)] = Vec_IntSize(vOneNew);
                Vec_IntPush( vOneNew, Offset+int64(pObjPi.pNext) );
            }
        }
        // replace the class
        Vec_PtrWriteEntry( vParts, i, vOneNew );
        Vec_IntFree( vOne );
    }
}

/**Function*************************************************************

  Synopsis    [Creates AIG of one partition with speculative reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrCreatePart_rec(pCla *Fra_Cla_t, pNew *Aig_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return (Aig_Obj_t *)pObj.pData; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsCi(pObj) {
//        Aig_Obj_t * pRepr = Fra_ClassObjRepr(pObj);
        var pRepr *Aig_Obj_t  = pCla.pMemRepr[pObj.Id];
        if pRepr == nil {
            pObj.pData = Aig_ObjCreateCi( pNew );
        } else {
            pObj.pData = Fra_LcrCreatePart_rec( pCla, pNew, p, pRepr );
            pObj.pData = Aig_NotCond( (Aig_Obj_t *)pObj.pData, pRepr.fPhase ^ pObj.fPhase );
        }
        return (Aig_Obj_t *)pObj.pData;
    }
    Fra_LcrCreatePart_rec( pCla, pNew, p, Aig_ObjFanin0(pObj) );
    Fra_LcrCreatePart_rec( pCla, pNew, p, Aig_ObjFanin1(pObj) );
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ));
}

/**Function*************************************************************

  Synopsis    [Creates AIG of one partition with speculative reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_LcrCreatePart(p *Fra_Lcr_t, vPart *Vec_Int_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var Out, i int32;
    // create new AIG for this partition
    pNew = Aig_ManStartFrom( p.pAig );
    Aig_ManIncrementTravId( p.pAig );
    Aig_ObjSetTravIdCurrent( p.pAig, Aig_ManConst1(p.pAig) );
    Aig_ManConst1(p.pAig).pData = Aig_ManConst1(pNew);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((Out) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Aig_ManCo( p.pAig, Out );
        if pObj.fMarkA {
            pObjNew = Fra_LcrCreatePart_rec( p.pCla, pNew, p.pAig, Aig_ObjFanin0(pObj) );
            pObjNew = Aig_NotCond( pObjNew, Aig_ObjFaninC0(pObj) );
        
        } else {
            pObjNew = Aig_ManConst1( pNew ); }
        Aig_ObjCreateCo( pNew, pObjNew ); 
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Marks the nodes belonging to the equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassNodesMark(p *Fra_Lcr_t) {
    var pObj *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var i, c, Offset int32;
    // compute the LO/LI offset
    Offset = Aig_ManCoNum(p.pCla.pAig) - Aig_ManCiNum(p.pCla.pAig);
    // mark the nodes remaining in the classes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pCla.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pCla.vClasses1, i)), 1); i++  {
        pObj = Aig_ManCo( p.pCla.pAig, Offset+int64(pObj.pNext) );
        pObj.fMarkA = 1;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pCla.vClasses)) && (((ppClass) = (Aig_Obj_t **)Vec_PtrEntry(p.pCla.vClasses, i)), 1); i++  {
        for  c = 0; ppClass[c]; c++  {
            pObj = Aig_ManCo( p.pCla.pAig, Offset+int64(ppClass[c].pNext) );
            pObj.fMarkA = 1;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Unmarks the nodes belonging to the equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassNodesUnmark(p *Fra_Lcr_t) {
    var pObj *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var i, c, Offset int32;
    // compute the LO/LI offset
    Offset = Aig_ManCoNum(p.pCla.pAig) - Aig_ManCiNum(p.pCla.pAig);
    // mark the nodes remaining in the classes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pCla.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pCla.vClasses1, i)), 1); i++  {
        pObj = Aig_ManCo( p.pCla.pAig, Offset+int64(pObj.pNext) );
        pObj.fMarkA = 0;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pCla.vClasses)) && (((ppClass) = (Aig_Obj_t **)Vec_PtrEntry(p.pCla.vClasses, i)), 1); i++  {
        for  c = 0; ppClass[c]; c++  {
            pObj = Aig_ManCo( p.pCla.pAig, Offset+int64(ppClass[c].pNext) );
            pObj.fMarkA = 0;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Performs choicing of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_FraigLatchCorrespondence(pAig *Aig_Man_t, nFramesP int32, nConfMax int32, fProve int32, fVerbose int32, pnIter *int32, TimeLimit float32) *Aig_Man_t {
    var nPartSize int32     = 200;
    var fReprSelect int32;
    var p *Fra_Lcr_t;
    var pSml *Fra_Sml_t;
    var pTemp *Fra_Man_t;
    var pAigPart *Aig_Man_t;
    var pAigTemp *Aig_Man_t;
    var pAigNew *Aig_Man_t = nil;
    var vPart *Vec_Int_t;
    var i, nIter int32;
    var timeSim abctime;
    var clk2 abctime;
    var clk3 abctime;
    var clk abctime = Abc_Clock();
    var TimeToStop abctime;
    if TimeLimit != nil {
        TimeToStop = TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        TimeToStop = 0;
    }
    if Aig_ManNodeNum(pAig) == 0 {
        if pnIter != nil { *pnIter = 0; }
        // Ntl_ManFinalize() requires the following to satisfy an assertion.
        Aig_ManReprStart(pAig,Aig_ManObjNumMax(pAig));
        return Aig_ManDupOrdered(pAig);
    }
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0"); 

    // simulate the AIG 
clk2 = Abc_Clock();
if fVerbose != 0 {
fmt.Printf("Simulating AIG with %d nodes for %d cycles ...  ", Aig_ManNodeNum(pAig), nFramesP + 32); }
    pSml = Fra_SmlSimulateSeq( pAig, nFramesP, 32, 1, 1  ); 
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk2)))/(float64(int64(time.Second))));;
}
timeSim = Abc_Clock() - clk2;

    // check if simulation discovered non-constant-0 POs
    if (fProve != 0) && pSml.fNonConstOut {
        pAig.pSeqModel = Fra_SmlGetCounterExample( pSml );
        Fra_SmlStop( pSml );
        return nil;
    }

    // start the manager
    p = Lcr_ManAlloc( pAig );
    p.nFramesP = nFramesP;
    p.fVerbose = fVerbose;
    p.timeSim += timeSim;

    pTemp = Fra_LcrAigPrepare( pAig );
    pTemp.pBmc = (Fra_Bmc_t *)p;
    pTemp.pSml = pSml;

    // get preliminary info about equivalence classes
    pTemp.pCla = p.pCla = Fra_ClassesStart( p.pAig );
    Fra_ClassesPrepare( p.pCla, 1, 0 );
    p.pCla.pFuncNodeIsConst   = Fra_LcrNodeIsConst;
    p.pCla.pFuncNodesAreEqual = Fra_LcrNodesAreEqual;
    Fra_SmlStop( pTemp.pSml );

    // partition the AIG for latch correspondence computation
clk2 = Abc_Clock();
if fVerbose != 0 {
fmt.Printf("Partitioning AIG ...  "); }
    pAigPart = Fra_LcrDeriveAigForPartitioning( p );
    p.vParts = (Vec_Ptr_t *)Aig_ManPartitionSmart( pAigPart, nPartSize, 0, nil );
    Fra_LcrRemapPartitions( p.vParts, p.pCla, p.pInToOutPart, p.pInToOutNum );
    Aig_ManStop( pAigPart );
if fVerbose != 0 {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk2)))/(float64(int64(time.Second))));;
p.timePart += Abc_Clock() - clk2;
}

    // get the initial stats
    p.nLitsBeg  = Fra_ClassesCountLits( p.pCla );
    p.nNodesBeg = Aig_ManNodeNum(p.pAig);
    p.nRegsBeg  = Aig_ManRegNum(p.pAig);

    // perform iterative reduction of the partitions
    p.fRefining = 1;
    for  nIter = 0; p.fRefining; nIter++  {
        p.fRefining = 0;
        clk3 = Abc_Clock();
        // derive AIGs for each partition
        Fra_ClassNodesMark( p );
        Vec_PtrClear( p.vFraigs );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(p.vParts, i)), 1); i++  {
            if TimeLimit != 0.0 && Abc_Clock() > TimeToStop {
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(p.vFraigs)) && (((pAigPart) = (Aig_Man_t *)Vec_PtrEntry(p.vFraigs, i)), 1); i++  {
                    Aig_ManStop( pAigPart );
                }
                Aig_ManCleanMarkA( pAig );
                Aig_ManCleanMarkB( pAig );
                fmt.Printf("Fra_FraigLatchCorrespondence(): Runtime limit exceeded.\n");
                goto finish;
            }
clk2 = Abc_Clock();
            pAigPart = Fra_LcrCreatePart( p, vPart );
p.timeTrav += Abc_Clock() - clk2;
clk2 = Abc_Clock();
            pAigTemp  = Fra_FraigEquivence( pAigPart, nConfMax, 0 );
p.timeFraig += Abc_Clock() - clk2;
            Vec_PtrPush( p.vFraigs, pAigTemp );
/*
            {
                char Name[1000];
                sprintf( Name, "part%04d.blif", i );
                Aig_ManDumpBlif( pAigPart, Name, NULL, NULL );
            }
printf( "Finished part %4d (out of %4d).  ", i, Vec_PtrSize(p->vParts) );
ABC_PRT( "Time", Abc_Clock() - clk3 );
*/

            Aig_ManStop( pAigPart );
        }
        Fra_ClassNodesUnmark( p );
        // report the intermediate results
        if fVerbose != 0 {
            fmt.Printf("%3d : Const = %6d. Class = %6d.  L = %6d. Part = %3d.  ", nIter, Vec_PtrSize(p.pCla.vClasses1), Vec_PtrSize(p.pCla.vClasses), Fra_ClassesCountLits(p.pCla), Vec_PtrSize(p.vParts));
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("T"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk3)))/(float64(int64(time.Second))));;
        }
        // refine the classes
        Fra_LcrAigPrepareTwo( p.pAig, pTemp );
        if Fra_ClassesRefine( p.pCla ) {
            p.fRefining = 1; }
        if Fra_ClassesRefine1( p.pCla, 0, nil ) {
            p.fRefining = 1; }
        // clean the fraigs
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vFraigs)) && (((pAigPart) = (Aig_Man_t *)Vec_PtrEntry(p.vFraigs, i)), 1); i++  {
            Aig_ManStop( pAigPart );
        }

        // repartition if needed
        if true {
clk2 = Abc_Clock();
            Vec_VecFree( (Vec_Vec_t *)p.vParts );
            pAigPart = Fra_LcrDeriveAigForPartitioning( p );
            p.vParts = (Vec_Ptr_t *)Aig_ManPartitionSmart( pAigPart, nPartSize, 0, nil );
            Fra_LcrRemapPartitions( p.vParts, p.pCla, p.pInToOutPart, p.pInToOutNum );
            Aig_ManStop( pAigPart );
p.timePart += Abc_Clock() - clk2;
        }
    }
    p.nIters = nIter;

    // move the classes into representatives and reduce AIG
clk2 = Abc_Clock();
//    Fra_ClassesPrint( p->pCla, 1 );
    if fReprSelect != 0 {
        Fra_ClassesSelectRepr( p.pCla ); }
    Fra_ClassesCopyReprs( p.pCla, nil );
    pAigNew = Aig_ManDupRepr( p.pAig, 0 );
    Aig_ManSeqCleanup( pAigNew );
//    Aig_ManCountMergeRegs( pAigNew );
p.timeUpdate += Abc_Clock() - clk2;
p.timeTotal = Abc_Clock() - clk;
    // get the final stats
    p.nLitsEnd  = Fra_ClassesCountLits( p.pCla );
    p.nNodesEnd = Aig_ManNodeNum(pAigNew);
    p.nRegsEnd  = Aig_ManRegNum(pAigNew);
finish:
    // inlining of macro ABC_FREE
    ((pTemp) ? (free((char *) (pTemp)), (pTemp) = 0) : 0);
    Lcr_ManFree( p );
    if pnIter != nil { *pnIter = nIter; }
    return pAigNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

