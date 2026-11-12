package abc

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [abcNames.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures working with net and node names.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcNames.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"
//#include "misc/util/utilNam.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the unique name for the object.]

  Description [If the name previously did not exist, creates a new unique
  name but does not assign this name to the object. The temporary unique
  name is stored in a static buffer inside this procedure. It is important 
  that the name is used before the function is called again!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjName(pObj *Abc_Obj_t) string {
    return Nm_ManCreateUniqueName( pObj.pNtk.pManName, pObj.Id );
}

/**Function*************************************************************

  Synopsis    [Assigns the given name to the object.]

  Description [The object should not have a name assigned. The same
  name may be used for several objects, which they share the same net
  in the original netlist. (For example, latch output and primary output 
  may have the same name.) This procedure returns the pointer to the 
  internally stored representation of the given name.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjAssignName(pObj *Abc_Obj_t, pName string, pSuffix string) string {
    assert(pName != nil, "pName != nil");
    return Nm_ManStoreIdName( pObj.pNtk.pManName, pObj.Id, pObj.Type, pName, pSuffix );
}

/**Function*************************************************************

  Synopsis    [Appends name to the prefix]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_ObjNamePrefix_Buffer string // was [2000]byte in C

func Abc_ObjNamePrefix(pObj *Abc_Obj_t, pPrefix string) string {
    
    spStatic_Abc_ObjNamePrefix_Buffer = fmt.Sprintf("%s%s", pPrefix, Abc_ObjName(pObj));
    return spStatic_Abc_ObjNamePrefix_Buffer;
}

/**Function*************************************************************

  Synopsis    [Appends suffic to the name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_ObjNameSuffix_Buffer string // was [2000]byte in C

func Abc_ObjNameSuffix(pObj *Abc_Obj_t, pSuffix string) string {
    
    spStatic_Abc_ObjNameSuffix_Buffer = fmt.Sprintf("%s%s", Abc_ObjName(pObj), pSuffix);
    return spStatic_Abc_ObjNameSuffix_Buffer;
}

/**Function*************************************************************

  Synopsis    [Returns the dummy PI name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_ObjNameDummy_Buffer string // was [2000]byte in C

func Abc_ObjNameDummy(pPrefix string, Num int32, nDigits int32) string {
    
    spStatic_Abc_ObjNameDummy_Buffer = fmt.Sprintf("%s%0*d", pPrefix, nDigits, Num);
    return spStatic_Abc_ObjNameDummy_Buffer;
}

/**Function*************************************************************

  Synopsis    [Tranfers names to the old network.]

  Description [Assumes that the new nodes are attached using pObj->pCopy.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTrasferNames(pNtk *Abc_Ntk_t, pNtkNew *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkPiNum(pNtk) == Abc_NtkPiNum(pNtkNew), "Abc_NtkPiNum(pNtk) == Abc_NtkPiNum(pNtkNew)");
    assert(Abc_NtkPoNum(pNtk) == Abc_NtkPoNum(pNtkNew), "Abc_NtkPoNum(pNtk) == Abc_NtkPoNum(pNtkNew)");
    assert(Abc_NtkBoxNum(pNtk) == Abc_NtkBoxNum(pNtkNew), "Abc_NtkBoxNum(pNtk) == Abc_NtkBoxNum(pNtkNew)");
    assert(Nm_ManNumEntries(pNtk.pManName) > 0, "Nm_ManNumEntries(pNtk.pManName) > 0");
    assert(Nm_ManNumEntries(pNtkNew.pManName) == 0, "Nm_ManNumEntries(pNtkNew.pManName) == 0");
    // copy the CI/CO/box names
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(Abc_ObjFanout0Ntk(pObj)), nil );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  { 
        Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(Abc_ObjFanin0Ntk(pObj)), nil );
    }
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  { 
        Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(pObj), nil );
    }
}

/**Function*************************************************************

  Synopsis    [Tranfers names to the old network.]

  Description [Assumes that the new nodes are attached using pObj->pCopy.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTrasferNamesNoLatches(pNtk *Abc_Ntk_t, pNtkNew *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkPiNum(pNtk) == Abc_NtkPiNum(pNtkNew), "Abc_NtkPiNum(pNtk) == Abc_NtkPiNum(pNtkNew)");
    assert(Abc_NtkPoNum(pNtk) == Abc_NtkPoNum(pNtkNew), "Abc_NtkPoNum(pNtk) == Abc_NtkPoNum(pNtkNew)");
    assert(Nm_ManNumEntries(pNtk.pManName) > 0, "Nm_ManNumEntries(pNtk.pManName) > 0");
    assert(Nm_ManNumEntries(pNtkNew.pManName) == 0, "Nm_ManNumEntries(pNtkNew.pManName) == 0");
    // copy the CI/CO/box name and skip latches and theirs inputs/outputs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjFaninNum(pObj) == 0 || !Abc_ObjIsLatch(Abc_ObjFanin0(pObj)) {
            Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(Abc_ObjFanout0Ntk(pObj)), nil ); }
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  { 
        if Abc_ObjFanoutNum(pObj) == 0 || !Abc_ObjIsLatch(Abc_ObjFanout0(pObj)) {
            Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(Abc_ObjFanin0Ntk(pObj)), nil ); }
    }
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  { 
        if !Abc_ObjIsLatch(pObj) {
            Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(pObj), nil ); }
    }
}

/**Function*************************************************************

  Synopsis    [Gets fanin node names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeGetFaninNames(pNode *Abc_Obj_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pFanin *Abc_Obj_t;
    var i int32;
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Vec_PtrPush( vNodes, Abc_UtilStrsav(Abc_ObjName(pFanin)) );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Gets fanin node names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeGetFakeNames(nNames int32) *Vec_Ptr_t {
    var vNames *Vec_Ptr_t;
    var Buffer [5]byte;
    var i int32;

    vNames = Vec_PtrAlloc( nNames );
    for  i = 0; i < nNames; i++  {
        if nNames < 26 {
            Buffer[0] = 'a' + i;
            Buffer[1] = 0;
        
        } else {
            Buffer[0] = 'a' + i%26;
            Buffer[1] = '0' + i/26;
            Buffer[2] = 0;
        }
        Vec_PtrPush( vNames, Extra_UtilStrsav(Buffer) );
    }
    return vNames;
}

/**Function*************************************************************

  Synopsis    [Gets fanin node names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeFreeNames(vNames *Vec_Ptr_t) {
    var i int32;
    if vNames == nil {
        return; }
    for  i = 0; i < vNames.nSize; i++  {
        // inlining of macro ABC_FREE
        ((vNames.pArray[i]) ? (free((char *) (vNames.pArray[i])), (vNames.pArray[i]) = 0) : 0);
    }
    Vec_PtrFree( vNames );
}

/**Function*************************************************************

  Synopsis    [Collects the CI or CO names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectCioNames(pNtk *Abc_Ntk_t, fCollectCos int32) *string {
    var pObj *Abc_Obj_t;
    var ppNames *string;
    var i int32;
    if fCollectCos != 0 {
        // inlining of macro ABC_ALLOC
        ppNames = ((char * *) malloc(sizeof(char *) * (size_t)(Abc_NtkCoNum(pNtk))));
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
            ppNames[i] = Abc_ObjName(pObj);
        }
    
    } else {
        // inlining of macro ABC_ALLOC
        ppNames = ((char * *) malloc(sizeof(char *) * (size_t)(Abc_NtkCiNum(pNtk))));
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
            ppNames[i] = Abc_ObjName(pObj);
        }
    }
    return ppNames;
}

/**Function*************************************************************

  Synopsis    [Orders PIs/POs/latches alphabetically.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeCompareNames(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
    var Diff int32  = int32(strings.Compare((char *)(*pp1).pCopy, (char *)(*pp2).pCopy));
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    Diff = (*pp1).Id - (*pp2).Id;
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}
func Abc_NtkOrderObjsByName(pNtk *Abc_Ntk_t, fComb int32) {
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkHasOnlyLatchBoxes(pNtk), "Abc_NtkHasOnlyLatchBoxes(pNtk)");
    // temporarily store the names in the copy field
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)Abc_ObjName(pObj);
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)Abc_ObjName(pObj);
    }
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)Abc_ObjName(Abc_ObjFanout0(pObj));
    }
    // order objects alphabetically
    qsort( (void *)Vec_PtrArray(pNtk.vPis), (size_t)Vec_PtrSize(pNtk.vPis), sizeof(Abc_Obj_t *), 
            (int (*)(void *, void *)) Abc_NodeCompareNames );
    qsort( (void *)Vec_PtrArray(pNtk.vPos), (size_t)Vec_PtrSize(pNtk.vPos), sizeof(Abc_Obj_t *), 
            (int (*)(void *, void *)) Abc_NodeCompareNames );
    // if the comparison if combinational (latches as PIs/POs), order them too
    if fComb != 0 {
        qsort( (void *)Vec_PtrArray(pNtk.vBoxes), (size_t)Vec_PtrSize(pNtk.vBoxes), sizeof(Abc_Obj_t *), 
                    (int (*)(void *, void *)) Abc_NodeCompareNames ); }
    // order CIs/COs first PIs/POs(Asserts) then latches
    Abc_NtkOrderCisCos( pNtk );
    // clean the copy fields
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {
        pObj.pCopy = nil;
    }
}

/**Function*************************************************************

  Synopsis    [Creates name manager storing input/output names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkNameMan(p *Abc_Ntk_t, fOuts int32) *Abc_Nam_t {
    if fOuts != 0 {
        var pObj *Abc_Obj_t;  var i int32;
        var pStrsCo *Abc_Nam_t  = Abc_NamStart( Abc_NtkCoNum(p), 24 );
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(p)) && (((pObj) = Abc_NtkCo(p, i)), 1); i++  {
            Abc_NamStrFindOrAdd( pStrsCo, Abc_ObjName(pObj), nil );
        }
        assert(Abc_NamObjNumMax(pStrsCo) == i + 1, "Abc_NamObjNumMax(pStrsCo) == i + 1");
        return pStrsCo;
    
    } else {
        var pObj *Abc_Obj_t;  var i int32;
        var pStrsCi *Abc_Nam_t  = Abc_NamStart( Abc_NtkCiNum(p), 24 );
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
            Abc_NamStrFindOrAdd( pStrsCi, Abc_ObjName(pObj), nil );
        }
        assert(Abc_NamObjNumMax(pStrsCi) == i + 1, "Abc_NamObjNumMax(pStrsCi) == i + 1");
        return pStrsCi;
   }
}

/**Function*************************************************************

  Synopsis    [Orders PIs/POs/latches alphabetically.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeCompareIndexes(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
    var Diff int32  = (*pp1).iTemp - (*pp2).iTemp;
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}
func Abc_NtkTransferOrder(pNtkOld *Abc_Ntk_t, pNtkNew *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;  var i int32;
    var pStrsCi *Abc_Nam_t  = Abc_NtkNameMan( pNtkOld, 0 );
    var pStrsCo *Abc_Nam_t  = Abc_NtkNameMan( pNtkOld, 1 );
    assert(Abc_NtkPiNum(pNtkOld) == Abc_NtkPiNum(pNtkNew), "Abc_NtkPiNum(pNtkOld) == Abc_NtkPiNum(pNtkNew)");
    assert(Abc_NtkPoNum(pNtkOld) == Abc_NtkPoNum(pNtkNew), "Abc_NtkPoNum(pNtkOld) == Abc_NtkPoNum(pNtkNew)");
    assert(Abc_NtkLatchNum(pNtkOld) == Abc_NtkLatchNum(pNtkNew), "Abc_NtkLatchNum(pNtkOld) == Abc_NtkLatchNum(pNtkNew)");
    // transfer to the new network
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtkNew)) && (((pObj) = Abc_NtkCi(pNtkNew, i)), 1); i++  {
        pObj.iTemp = Abc_NamStrFind(pStrsCi, Abc_ObjName(pObj));
        assert(pObj.iTemp > 0 && pObj.iTemp <= Abc_NtkCiNum(pNtkNew), "pObj.iTemp > 0 && pObj.iTemp <= Abc_NtkCiNum(pNtkNew)");
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtkNew)) && (((pObj) = Abc_NtkCo(pNtkNew, i)), 1); i++  {
        pObj.iTemp = Abc_NamStrFind(pStrsCo, Abc_ObjName(pObj));
        assert(pObj.iTemp > 0 && pObj.iTemp <= Abc_NtkCoNum(pNtkNew), "pObj.iTemp > 0 && pObj.iTemp <= Abc_NtkCoNum(pNtkNew)");
    }
    Abc_NamDeref( pStrsCi );
    Abc_NamDeref( pStrsCo );
    // order PI/PO 
    qsort( (void *)Vec_PtrArray(pNtkNew.vPis), (size_t)Vec_PtrSize(pNtkNew.vPis), sizeof(Abc_Obj_t *), 
            (int (*)(void *, void *)) Abc_NodeCompareIndexes );
    qsort( (void *)Vec_PtrArray(pNtkNew.vPos), (size_t)Vec_PtrSize(pNtkNew.vPos), sizeof(Abc_Obj_t *), 
            (int (*)(void *, void *)) Abc_NodeCompareIndexes );
    // order CI/CO 
    qsort( (void *)Vec_PtrArray(pNtkNew.vCis), (size_t)Vec_PtrSize(pNtkNew.vCis), sizeof(Abc_Obj_t *), 
            (int (*)(void *, void *)) Abc_NodeCompareIndexes );
    qsort( (void *)Vec_PtrArray(pNtkNew.vCos), (size_t)Vec_PtrSize(pNtkNew.vCos), sizeof(Abc_Obj_t *), 
            (int (*)(void *, void *)) Abc_NodeCompareIndexes );
    // order CIs/COs first PIs/POs(Asserts) then latches
    //Abc_NtkOrderCisCos( pNtk );
    // clean the copy fields
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtkNew)) && (((pObj) = Abc_NtkCi(pNtkNew, i)), 1); i++  {
        pObj.iTemp = 0;
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtkNew)) && (((pObj) = Abc_NtkCo(pNtkNew, i)), 1); i++  {
        pObj.iTemp = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Checks that the order and number of CI/CO is the same.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeCompareCiCo(pNtkOld *Abc_Ntk_t, pNtkNew *Abc_Ntk_t) int32 {
    var i int32;
    if Abc_NtkPiNum(pNtkOld) != Abc_NtkPiNum(pNtkNew) {
        return 0; }
    if Abc_NtkPoNum(pNtkOld) != Abc_NtkPoNum(pNtkNew) {
        return 0; }
    if Abc_NtkLatchNum(pNtkOld) != Abc_NtkLatchNum(pNtkNew) {
        return 0; }
    for  i = 0; i < Abc_NtkCiNum(pNtkOld); i++  {
        if int32(strings.Compare(Abc_ObjName(Abc_NtkCi(pNtkOld, i)), Abc_ObjName(Abc_NtkCi(pNtkNew, i)))) != 0 {
            return 0; }
    }
    for  i = 0; i < Abc_NtkCoNum(pNtkOld); i++  {
        if int32(strings.Compare(Abc_ObjName(Abc_NtkCo(pNtkOld, i)), Abc_ObjName(Abc_NtkCo(pNtkNew, i)))) != 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds dummy names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAddDummyPiNames(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var nDigits, i int32;
    nDigits = Abc_Base10Log( Abc_NtkPiNum(pNtk) );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_ObjAssignName( pObj, Abc_ObjNameDummy("pi", i, nDigits), nil );
    }
}

/**Function*************************************************************

  Synopsis    [Adds dummy names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAddDummyPoNames(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var nDigits, i int32;
    nDigits = Abc_Base10Log( Abc_NtkPoNum(pNtk) );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_ObjAssignName( pObj, Abc_ObjNameDummy("po", i, nDigits), nil );
    }
}

/**Function*************************************************************

  Synopsis    [Adds dummy names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAddDummyBoxNames(pNtk *Abc_Ntk_t) {
    var pName string;
    var PrefLi [100]byte;
    var PrefLo [100]byte;
    var pObj *Abc_Obj_t;
    var nDigits, i, k, CountCur, CountMax int32;
    // if PIs/POs already have nodes with what looks like latch names
    // we need to add different prefix for the new latches
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        CountCur = 0;
        pName = Abc_ObjName(pObj);
        for  k = 0; pName[k]; k++  {
            if pName[k] == 'l' {
                CountCur++;
            } else {
                break; }
        }
        CountMax = Abc_MaxInt( CountMax, CountCur );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        CountCur = 0;
        pName = Abc_ObjName(pObj);
        for  k = 0; pName[k]; k++  {
            if pName[k] == 'l' {
                CountCur++;
            } else {
                break; }
        }
        CountMax = Abc_MaxInt( CountMax, CountCur );
    }
//printf( "CountMax = %d\n", CountMax );
    assert(CountMax < 100-2, "CountMax < 100-2");
    for  i = 0; i <= CountMax; i++  {
        PrefLi[i] = PrefLo[i] = 'l';
    }
    PrefLi[i] = 'i';
    PrefLo[i] = 'o';
    PrefLi[i+1] = 0;
    PrefLo[i+1] = 0;
    // create latch names
    assert(!Abc_NtkIsNetlist(pNtk), "!Abc_NtkIsNetlist(pNtk)");
    nDigits = Abc_Base10Log( Abc_NtkLatchNum(pNtk) );
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        Abc_ObjAssignName( pObj, Abc_ObjNameDummy("l", i, nDigits), nil );
        Abc_ObjAssignName( Abc_ObjFanin0(pObj),  Abc_ObjNameDummy(PrefLi, i, nDigits), nil );
        Abc_ObjAssignName( Abc_ObjFanout0(pObj), Abc_ObjNameDummy(PrefLo, i, nDigits), nil );
    }
    }
/*
    nDigits = Abc_Base10Log( Abc_NtkBlackboxNum(pNtk) );
    Abc_NtkForEachBlackbox( pNtk, pObj, i )
    {
        pName = Abc_ObjAssignName( pObj, Abc_ObjNameDummy("B", i, nDigits), NULL );
        nDigitsF = Abc_Base10Log( Abc_ObjFaninNum(pObj) );
        Abc_ObjForEachFanin( pObj, pTerm, k )
            Abc_ObjAssignName( Abc_ObjFanin0(pObj), pName, Abc_ObjNameDummy("i", k, nDigitsF) );
        nDigitsF = Abc_Base10Log( Abc_ObjFanoutNum(pObj) );
        Abc_ObjForEachFanout( pObj, pTerm, k )
            Abc_ObjAssignName( Abc_ObjFanin0(pObj), pName, Abc_ObjNameDummy("o", k, nDigitsF) );
    }
*/
}

/**Function*************************************************************

  Synopsis    [Replaces names by short names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkShortNames(pNtk *Abc_Ntk_t) {
    Nm_ManFree( pNtk.pManName );
    pNtk.pManName = Nm_ManCreate( Abc_NtkCiNum(pNtk) + Abc_NtkCoNum(pNtk) + Abc_NtkBoxNum(pNtk) );
    Abc_NtkAddDummyPiNames( pNtk );
    Abc_NtkAddDummyPoNames( pNtk );
    Abc_NtkAddDummyBoxNames( pNtk );
}
func Abc_NtkCleanNames(pNtk *Abc_Ntk_t) {  
    var pObj *Abc_Obj_t; var i int32;
    var pManName *Nm_Man_t  = Nm_ManCreate( Abc_NtkCiNum(pNtk) + Abc_NtkCoNum(pNtk) + Abc_NtkBoxNum(pNtk) );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Nm_ManStoreIdName( pManName, pObj.Id, pObj.Type, Abc_ObjName(pObj), nil );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Nm_ManStoreIdName( pManName, pObj.Id, pObj.Type, Abc_ObjName(pObj), nil );
    }
    Nm_ManFree( pNtk.pManName );
    pNtk.pManName = pManName;
}

/**Function*************************************************************

  Synopsis    [Moves names from the other network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRedirectCiCo(pNtk *Abc_Ntk_t) {
    var pObj, pObjCi, pFanin *Abc_Obj_t; 
    var i, Count int32;
    // if CO points to CI with the same name, remove buffer between them
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        var nCiId int32  = Nm_ManFindIdByNameTwoTypes( pNtk.pManName, Abc_ObjName(pObj), ABC_OBJ_PI, ABC_OBJ_BO );
        if nCiId == -1 {
            continue; }
        pObjCi = Abc_NtkObj( pNtk, nCiId );
        assert(!(int32(strings.Compare(Abc_ObjName(pObj), Abc_ObjName(pObjCi))) != 0), "!strcmp( Abc_ObjName(pObj), Abc_ObjName(pObjCi) )");
        pFanin = Abc_ObjFanin0(pObj);
        if pFanin == pObjCi {
            continue; }
        assert(Abc_NodeIsBuf(pFanin), "Abc_NodeIsBuf(pFanin)");
        Abc_ObjPatchFanin( pObj, pFanin, pObjCi );
        if Abc_ObjFanoutNum(pFanin) == 0 {
            Abc_NtkDeleteObj( pFanin ); }
        Count++;
    }
    if Count != 0 {
        fmt.Printf("Redirected %d POs from buffers to PIs with the same name.\n", Count); }
}
func Abc_NtkMoveNames(pNtk *Abc_Ntk_t, pOld *Abc_Ntk_t) {
    var pObj *Abc_Obj_t; var i int32;
    Nm_ManFree( pNtk.pManName );
    pNtk.pManName = Nm_ManCreate( Abc_NtkCiNum(pNtk) + Abc_NtkCoNum(pNtk) + Abc_NtkBoxNum(pNtk) );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_ObjAssignName( pObj, Abc_ObjName(Abc_NtkPi(pOld, i)), nil );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_ObjAssignName( pObj, Abc_ObjName(Abc_NtkPo(pOld, i)), nil );
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        Abc_ObjAssignName( Abc_ObjFanin0(pObj),  Abc_ObjName(Abc_ObjFanin0(Abc_NtkBox(pOld, i))),  nil );
        Abc_ObjAssignName( Abc_ObjFanout0(pObj), Abc_ObjName(Abc_ObjFanout0(Abc_NtkBox(pOld, i))), nil );
    }
    }
    Abc_NtkRedirectCiCo( pNtk );
}


/**Function*************************************************************

  Synopsis    [Saves name IDs into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStartNameIds(p *Abc_Ntk_t) {
    var pFileName string // was [1000]byte in C

    var pFile *os.File;
    var pObj, pFanin *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i int32;
    var Counter int32 = 1;
    assert(Abc_NtkIsNetlist(p), "Abc_NtkIsNetlist(p)");
    assert(p.vNameIds == nil, "p.vNameIds == nil");
    assert(len(p.pSpec) < 1000, "strlen(p.pSpec) < 1000");
    pFileName = fmt.Sprintf("%s_%s_names.txt", Extra_FileNameGenericAppend(p.pSpec,""), Extra_FileNameExtension(p.pSpec));
    pFile = fopen( pFileName, "wb" );
    p.vNameIds = Vec_IntStart( Abc_NtkObjNumMax(p) );
    // add inputs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s            \n", Abc_ObjName(Abc_ObjFanout0(pObj)));
Vec_IntWriteEntry(p.vNameIds, Abc_ObjId(pObj), 2*Counter++);
    }
    // add outputs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p)) && (((pObj) = Abc_NtkCo(p, i)), 1); i++  {
        pFanin = Abc_ObjFanin0(Abc_ObjFanin0(pObj));
        if !Vec_IntEntry(p.vNameIds, Abc_ObjId(pFanin)) {
            fmt.Fprintf(pFile, "%s            \n", Abc_ObjName(Abc_ObjFanout0(pFanin)));
Vec_IntWriteEntry(p.vNameIds, Abc_ObjId(pFanin), 2*Counter++); }
    }
    // add nodes in a topo order
    vNodes = Abc_NtkDfs( p, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if !Vec_IntEntry(p.vNameIds, Abc_ObjId(pObj)) {
            fmt.Fprintf(pFile, "%s            \n", Abc_ObjName(Abc_ObjFanout0(pObj)));
Vec_IntWriteEntry(p.vNameIds, Abc_ObjId(pObj), 2*Counter++); }
    }
    Vec_PtrFree( vNodes );
    _ = (pFile).Close();
    // transfer driver node names to COs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p)) && (((pObj) = Abc_NtkCo(p, i)), 1); i++  {
        pFanin = Abc_ObjFanin0(Abc_ObjFanin0(pObj));
        Vec_IntWriteEntry( p.vNameIds, Abc_ObjId(pObj), Vec_IntEntry(p.vNameIds, Abc_ObjId(pFanin)) );
        Vec_IntWriteEntry( p.vNameIds, Abc_ObjId(pFanin), 0 );
    }
}

/**Function*************************************************************

  Synopsis    [Remaps the AIG from the old manager into the new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTransferNameIds(p *Abc_Ntk_t, pNew *Abc_Ntk_t) {
    var pObj, pObjNew *Abc_Obj_t;
    var i int32;
    assert(p.vNameIds != nil, "p.vNameIds != nil");
    assert(pNew.vNameIds == nil, "pNew.vNameIds == nil");
    pNew.vNameIds = Vec_IntStart( Abc_NtkObjNumMax(pNew) );
//    Abc_NtkForEachCi( p, pObj, i )
//        printf( "%d ", Vec_IntEntry(p->vNameIds, Abc_ObjId(pObj)) );
//    printf( "\n" );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {            if (pObj) == nil {} else
        if pObj.pCopy && i < Vec_IntSize(p.vNameIds) && Vec_IntEntry(p.vNameIds, i) {
            pObjNew = Abc_ObjRegular(pObj.pCopy);
            assert(Abc_ObjNtk(pObjNew) == pNew, "Abc_ObjNtk(pObjNew) == pNew");
            if Abc_ObjIsCi(pObjNew) && !Abc_ObjIsCi(pObj) { // do not overwrite CI name by internal node name
                continue; }
            Vec_IntWriteEntry( pNew.vNameIds, Abc_ObjId(pObjNew), Vec_IntEntry(p.vNameIds, i) ^ Abc_ObjIsComplement(pObj.pCopy) );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Updates file with name IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkUpdateNameIds(p *Abc_Ntk_t) {
    var pFileName string // was [1000]byte in C

    var vStarts *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var pFile *os.File;
    var i, c, iVar, fCompl, fSeenSpace, Counter int32;
    assert(!Abc_NtkIsNetlist(p), "!Abc_NtkIsNetlist(p)");
    assert(len(p.pSpec) < 1000, "strlen(p.pSpec) < 1000");
    assert(p.vNameIds != nil, "p.vNameIds != nil");
    pFileName = fmt.Sprintf("%s_%s_names.txt", Extra_FileNameGenericAppend(p.pSpec,""), Extra_FileNameExtension(p.pSpec));
    pFile = fopen( pFileName, "r+" );
    // collect info about lines
    fSeenSpace = 0;
    vStarts = Vec_IntAlloc( 1000 );
    Vec_IntPush( vStarts, -1 );
    for  (c = fgetc(pFile)) != -1 && ++Counter  {
        if c == ' ' && !fSeenSpace {
            Vec_IntPush(vStarts, Counter);
fSeenSpace = 1;
        } else if c == '\n' {
            fSeenSpace = 0; }
    }
    // add info about names
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {            if (pObj) == nil {} else {
        if i == 0 || i >= Vec_IntSize(p.vNameIds) || !Vec_IntEntry(p.vNameIds, i) {
            continue; }
        iVar = Abc_Lit2Var( Vec_IntEntry(p.vNameIds, i) );
        fCompl = Abc_LitIsCompl( Vec_IntEntry(p.vNameIds, i) );
        assert(iVar < Vec_IntSize(vStarts), "iVar < Vec_IntSize(vStarts)");
        fseek( pFile, Vec_IntEntry(vStarts, iVar), SEEK_SET );
        fmt.Fprintf(pFile, "%s%d", fCompl? "-":"", i);
    }
    }
    printf( "Saved %d names into file \"%s\".\n", Vec_IntSize(vStarts)-1, pFileName );
    _ = (pFile).Close();
    Vec_IntFree( vStarts );
    Vec_IntFreeP( &p.vNameIds );
//    Abc_NtkForEachObj( p, pObj, i )
//        Abc_ObjPrint( stdout, pObj );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

