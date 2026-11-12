package abci

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [abcDetect.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Detect conditions.]

  Author      [Alan Mishchenko, Dao Ai Quoc]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 7, 2016.]

  Revision    [$Id: abcDetect.c,v 1.00 2016/06/07 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/util/utilNam.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "map/mio/mio.h"
//#include "map/mio/exp.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

 type Abc_FinType_t int32
const ( 
    ABC_FIN_NONE Abc_FinType_t = -100
   //  0:  unknown
    ABC_FIN_SA0 Abc_FinType_t = ABC_FIN_NONE + 1
           //  1:
    ABC_FIN_SA1 Abc_FinType_t = ABC_FIN_SA0 + 1
           //  2:
    ABC_FIN_NEG Abc_FinType_t = ABC_FIN_SA1 + 1
           //  3:
    ABC_FIN_RDOB_AND Abc_FinType_t = ABC_FIN_NEG + 1
      //  4:
    ABC_FIN_RDOB_NAND Abc_FinType_t = ABC_FIN_RDOB_AND + 1
     //  5:
    ABC_FIN_RDOB_OR Abc_FinType_t = ABC_FIN_RDOB_NAND + 1
       //  6:
    ABC_FIN_RDOB_NOR Abc_FinType_t = ABC_FIN_RDOB_OR + 1
      //  7:
    ABC_FIN_RDOB_XOR Abc_FinType_t = ABC_FIN_RDOB_NOR + 1
      //  8:
    ABC_FIN_RDOB_NXOR Abc_FinType_t = ABC_FIN_RDOB_XOR + 1
     //  9:
    ABC_FIN_RDOB_NOT Abc_FinType_t = ABC_FIN_RDOB_NXOR + 1
      // 10:
    ABC_FIN_RDOB_BUFF Abc_FinType_t = ABC_FIN_RDOB_NOT + 1
     // 11:
    ABC_FIN_RDOB_LAST Abc_FinType_t = ABC_FIN_RDOB_BUFF + 1      // 12:

)

 

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Generates fault list for the given mapped network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkGenFaultList(pNtk *Abc_Ntk_t, pFileName string, fStuckAt int32) {
    var pLib *Mio_Library_t  = (Mio_Library_t *)pNtk.pManFunc;
    var pGate *Mio_Gate_t;
    var pObj *Abc_Obj_t;
    var i int32;
    var Count int32 = 1;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    assert(Abc_NtkIsMappedLogic(pNtk), "Abc_NtkIsMappedLogic(pNtk)");
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        var pGateObj *Mio_Gate_t  = (Mio_Gate_t *)pObj.pData;
        var nInputs int32  = Mio_GateReadPinNum(pGateObj);
        // add basic faults (SA0, SA1, NEG)
        fmt.Fprintf(pFile, "%d %s %s\n", Count, Abc_ObjName(pObj), "SA0");
Count++;
        fmt.Fprintf(pFile, "%d %s %s\n", Count, Abc_ObjName(pObj), "SA1");
Count++;
        fmt.Fprintf(pFile, "%d %s %s\n", Count, Abc_ObjName(pObj), "NEG");
Count++;
        if fStuckAt != 0 {
            continue; }
        // add other faults, which correspond to changing the given gate
        // by another gate with the same support-size from the same library
        // inlining of macro Mio_LibraryForEachGate
        for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
            if pGate != pGateObj && Mio_GateReadPinNum(pGate) == nInputs {
                fmt.Fprintf(pFile, "%d %s %s\n", Count, Abc_ObjName(pObj), Mio_GateReadName(pGate));
Count++; }
        }
    }
    }
    fmt.Printf("Generated fault list \"%s\" for network \"%s\" with %d nodes and %d %sfaults.\n", pFileName, Abc_NtkName(pNtk), Abc_NtkNodeNum(pNtk), Count-1, fStuckAt ? "stuck-at ":"");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Recognize type.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadFinTypeMapped(pLib *Mio_Library_t, pThis string) int32 {
    var pGate *Mio_Gate_t  = Mio_LibraryReadGateByName( pLib, pThis, nil );
    if pGate == nil {
        fmt.Printf("Cannot find gate \"%s\" in the current library.\n", pThis);
        return ABC_FIN_NONE;
    }
    return Mio_GateReadCell( pGate );
}
func Io_ReadFinType(pThis string) int32 {
    if !(int32(strings.Compare(pThis, "SA0")) != 0) {        return ABC_FIN_SA0; }
    if !(int32(strings.Compare(pThis, "SA1")) != 0) {        return ABC_FIN_SA1; }
    if !(int32(strings.Compare(pThis, "NEG")) != 0) {        return ABC_FIN_NEG; }
    if !(int32(strings.Compare(pThis, "RDOB_AND")) != 0) {   return ABC_FIN_RDOB_AND; }
    if !(int32(strings.Compare(pThis, "RDOB_NAND")) != 0) {  return ABC_FIN_RDOB_NAND; }
    if !(int32(strings.Compare(pThis, "RDOB_OR")) != 0) {    return ABC_FIN_RDOB_OR; }
    if !(int32(strings.Compare(pThis, "RDOB_NOR")) != 0) {   return ABC_FIN_RDOB_NOR; }
    if !(int32(strings.Compare(pThis, "RDOB_XOR")) != 0) {   return ABC_FIN_RDOB_XOR; }
    if !(int32(strings.Compare(pThis, "RDOB_NXOR")) != 0) {  return ABC_FIN_RDOB_NXOR; }
    if !(int32(strings.Compare(pThis, "RDOB_NOT")) != 0) {   return ABC_FIN_RDOB_NOT; }
    if !(int32(strings.Compare(pThis, "RDOB_BUFF")) != 0) {  return ABC_FIN_RDOB_BUFF; }
    return ABC_FIN_NONE;
}
func Io_WriteFinType(Type int32) string {
    if Type == ABC_FIN_SA0 {          return "SA0"; }
    if Type == ABC_FIN_SA1 {          return "SA1"; }
    if Type == ABC_FIN_NEG {          return "NEG"; }
    if Type == ABC_FIN_RDOB_AND {    return "RDOB_AND" ; }
    if Type == ABC_FIN_RDOB_NAND {    return "RDOB_NAND"; }
    if Type == ABC_FIN_RDOB_OR {    return "RDOB_OR"  ; }
    if Type == ABC_FIN_RDOB_NOR {    return "RDOB_NOR" ; }
    if Type == ABC_FIN_RDOB_XOR {    return "RDOB_XOR" ; }
    if Type == ABC_FIN_RDOB_NXOR {    return "RDOB_NXOR"; }
    if Type == ABC_FIN_RDOB_NOT {    return "RDOB_NOT" ; }
    if Type == ABC_FIN_RDOB_BUFF {    return "RDOB_BUFF"; }
    return "Unknown";
}

/**Function*************************************************************

  Synopsis    [Read information from file.]

  Description [Returns information as a set of pairs: (ObjId, TypeId).
  Uses the current network to map ObjName given in the file into ObjId.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadFins(pNtk *Abc_Ntk_t, pFileName string, fVerbose int32) *Vec_Int_t {
    var pLib *Mio_Library_t  = (Mio_Library_t *)pNtk.pManFunc;
    var Buffer [1000]byte;
    var pObj *Abc_Obj_t;
    var pNam *Abc_Nam_t;
    var vMap *Vec_Int_t; 
    var vPairs *Vec_Int_t  = nil;
    var i int32;
    var Type int32;
    var iObj int32;
    var fFound int32;
    var nLines int32 = 1;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open input file \"%s\" for reading.\n", pFileName);
        return nil;
    }
    // map CI/node names into their IDs
    pNam = Abc_NamStart( 1000, 10 );
    vMap = Vec_IntAlloc( 1000 );
    Vec_IntPush( vMap, -1 );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if !Abc_ObjIsCi(pObj) && !Abc_ObjIsNode(pObj) {
            continue; }
        Abc_NamStrFindOrAdd( pNam, Abc_ObjName(pObj), &fFound );
        if fFound != 0 {
            fmt.Printf("The same name \"%s\" appears twice among CIs and internal nodes.\n", Abc_ObjName(pObj));
            goto finish;
        }
        Vec_IntPush( vMap, Abc_ObjId(pObj) );
    }
    }
    assert(Vec_IntSize(vMap) == Abc_NamObjNumMax(pNam), "Vec_IntSize(vMap) == Abc_NamObjNumMax(pNam)");
    // read file lines
    vPairs = Vec_IntAlloc( 1000 );
    Vec_IntPushTwo( vPairs, -1, -1 );
    for  fgets(Buffer, 1000, pFile) != nil  {
        // read line number
        var pToken string  = strtok( Buffer, " \n\r\t" );
        if pToken == nil {
            break; }
        if nLines++ != atoi(pToken) {
            fmt.Printf("Line numbers are not consecutive. Quitting.\n");
            Vec_IntFreeP( &vPairs );
            goto finish;
        }
        // read object name and find its ID
        pToken = strtok( nil, " \n\r\t" );
        iObj = Abc_NamStrFind( pNam, pToken );
        if !(iObj != 0) {
            fmt.Printf("Cannot find object with name \"%s\".\n", pToken);
            continue;
        }
        // read type
        pToken = strtok( nil, " \n\r\t" );
        if Abc_NtkIsMappedLogic(pNtk) {
            if !(int32(strings.Compare(pToken, "SA0")) != 0) || !(int32(strings.Compare(pToken, "SA1")) != 0) || !(int32(strings.Compare(pToken, "NEG")) != 0) {
                Type = Io_ReadFinType( pToken );
            } else {
                Type = Io_ReadFinTypeMapped( pLib, pToken ); }
        
        } else {
            Type = Io_ReadFinType( pToken ); }
        if Type == ABC_FIN_NONE {
            fmt.Printf("Cannot read type \"%s\" of object \"%s\".\n", pToken, Abc_ObjName(Abc_NtkObj(pNtk, iObj)));
            continue;
        }
        Vec_IntPushTwo( vPairs, Vec_IntEntry(vMap, iObj), Type );
    }
    assert(Vec_IntSize(vPairs) == 2 * nLines, "Vec_IntSize(vPairs) == 2 * nLines");
    fmt.Printf("Finished reading %d lines from the fault list file \"%s\".\n", nLines - 1, pFileName);

    // verify the reader by printing the results
    if fVerbose != 0 {
        // inlining of macro Vec_IntForEachEntryDoubleStart
        for  i = 2; (i+1 < Vec_IntSize(vPairs)) && (((iObj) = Vec_IntEntry(vPairs, i)), 1) && (((Type) = Vec_IntEntry(vPairs, i+1)), 1); i += 2  {
            fmt.Printf("%-10d%-10s%-10s\n", i/2, Abc_ObjName(Abc_NtkObj(pNtk, iObj)), Io_WriteFinType(Type));
        } }

finish:
    Vec_IntFree( vMap );
    Abc_NamDeref( pNam );
    _ = (pFile).Close();
    return vPairs;
}


/**Function*************************************************************

  Synopsis    [Extend the network by adding second timeframe.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFrameExtend(pNtk *Abc_Ntk_t) {
    var vFanins, vNodes *Vec_Ptr_t;
    var pObj, pFanin, pReset, pEnable, pSignal *Abc_Obj_t;
    var pResetN, pEnableN, pAnd0, pAnd1, pMux *Abc_Obj_t;
    var i int32;
    var k int32;
    var iStartPo int32;
    var nPisOld int32 = Abc_NtkPiNum(pNtk);
    var nPosOld int32 = Abc_NtkPoNum(pNtk);
    // skip if there are no flops
    if pNtk.nConstrs == 0 {
        return; }
    assert(Abc_NtkPiNum(pNtk) >= pNtk.nConstrs, "Abc_NtkPiNum(pNtk) >= pNtk.nConstrs");
    assert(Abc_NtkPoNum(pNtk) >= pNtk.nConstrs * 4, "Abc_NtkPoNum(pNtk) >= pNtk.nConstrs * 4");
    // collect nodes
    vNodes = Vec_PtrAlloc( Abc_NtkNodeNum(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Vec_PtrPush( vNodes, pObj ); }
    }
    // duplicate PIs
    vFanins = Vec_PtrAlloc( 2 );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        if i == nPisOld {
            break; }
        if i < nPisOld - pNtk.nConstrs {
            Abc_NtkDupObj( pNtk, pObj, 0 );
            Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(pObj), "_frame1" );
            continue;
        }
        // create flop input
        iStartPo = nPosOld + 4 * (i - nPisOld);
        pReset   = Abc_ObjFanin0( Abc_NtkPo( pNtk, iStartPo + 1 ) );
        pEnable  = Abc_ObjFanin0( Abc_NtkPo( pNtk, iStartPo + 2 ) );
        pSignal  = Abc_ObjFanin0( Abc_NtkPo( pNtk, iStartPo + 3 ) );
        pResetN  = Abc_NtkCreateNodeInv( pNtk, pReset );
        pEnableN = Abc_NtkCreateNodeInv( pNtk, pEnable );
        Vec_PtrFillTwo( vFanins, 2, pEnableN, pObj );
        pAnd0    = Abc_NtkCreateNodeAnd( pNtk, vFanins );
        Vec_PtrFillTwo( vFanins, 2, pEnable, pSignal );
        pAnd1    = Abc_NtkCreateNodeAnd( pNtk, vFanins );
        Vec_PtrFillTwo( vFanins, 2, pAnd0, pAnd1 );
        pMux     = Abc_NtkCreateNodeOr( pNtk, vFanins );
        Vec_PtrFillTwo( vFanins, 2, pResetN, pMux );
        pObj.pCopy = Abc_NtkCreateNodeAnd( pNtk, vFanins );
    }
    // duplicate internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_NtkDupObj( pNtk, pObj, 0 );
    }
    // connect objects
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy );
        }
    }
    // create new POs and reconnect flop inputs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        if i == nPosOld {
            break; }
        if i < nPosOld - 4 * pNtk.nConstrs {
            Abc_NtkDupObj( pNtk, pObj, 0 );
            Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(pObj), "_frame1" );
            Abc_ObjAddFanin( pObj.pCopy, Abc_ObjFanin0(pObj).pCopy );
            continue;
        }
        Abc_ObjPatchFanin( pObj, Abc_ObjFanin0(pObj), Abc_ObjFanin0(pObj).pCopy );
    }
    Vec_PtrFree( vFanins );
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Detect equivalence classes of nodes in terms of their TFO.]

  Description [Given is the logic network (pNtk) and the set of objects
  (primary inputs or internal nodes) to be considered (vObjs), this function
  returns a set of equivalence classes of these objects in terms of their 
  transitive fanout (TFO). Two objects belong to the same class if the set 
  of COs they feed into are the same.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDetectObjClasses_rec(pObj *Abc_Obj_t, vMap *Vec_Int_t, pHash *Hsh_VecMan_t, vTemp *Vec_Int_t) int32 {
    var vArray, vSet *Vec_Int_t;
    var pNext *Abc_Obj_t; var i int32;
    // get the CO set for this object
    var Entry int32  = Vec_IntEntry(vMap, Abc_ObjId(pObj));
    if Entry != -1 { // the set is already computed
        return Entry; }
    // compute a new CO set
    assert(Abc_ObjIsCi(pObj) || Abc_ObjIsNode(pObj), "Abc_ObjIsCi(pObj) || Abc_ObjIsNode(pObj)");
    // if there is no fanouts, the set of COs is empty
    if Abc_ObjFanoutNum(pObj) == 0 {
        Vec_IntWriteEntry( vMap, Abc_ObjId(pObj), 0 );
        return 0;
    }
    // compute the set for the first fanout
    Entry = Abc_NtkDetectObjClasses_rec( Abc_ObjFanout0(pObj), vMap, pHash, vTemp );
    if Abc_ObjFanoutNum(pObj) == 1 {
        Vec_IntWriteEntry( vMap, Abc_ObjId(pObj), Entry );
        return Entry;
    }
    vSet = Vec_IntAlloc( 16 );
    // initialize the set with that of first fanout
    vArray = Hsh_VecReadEntry( pHash, Entry );
    Vec_IntClear( vSet );
    Vec_IntAppend( vSet, vArray );
    // iteratively add sets of other fanouts
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if i == 0 { 
            continue; }
        Entry = Abc_NtkDetectObjClasses_rec( pNext, vMap, pHash, vTemp );
        vArray = Hsh_VecReadEntry( pHash, Entry );
        Vec_IntTwoMerge2( vSet, vArray, vTemp );
        // inlining of macro ABC_SWAP
        { var t Vec_Int_t  = *vSet; *vSet = *vTemp; *vTemp = t; };
    }
    // create or find new set and map the object into it
    Entry = Hsh_VecManAdd( pHash, vSet );
    Vec_IntWriteEntry( vMap, Abc_ObjId(pObj), Entry );
    Vec_IntFree( vSet );
    return Entry;
}
func Abc_NtkDetectObjClasses(pNtk *Abc_Ntk_t, vObjs *Vec_Int_t, pvCos **Vec_Wec_t) *Vec_Wec_t {
    var vClasses *Vec_Wec_t;   // classes of equivalence objects from vObjs
    var vClassMap *Vec_Int_t;  // mapping of each CO set into its class in vClasses
    var vClass *Vec_Int_t;     // one equivalence class     
    var pObj *Abc_Obj_t; 
    var i, iObj, SetId, ClassId int32;
    // create hash table to hash sets of CO indexes
    var pHash *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );
    // create elementary sets (each composed of one CO) and map COs into them
    var vMap *Vec_Int_t  = Vec_IntStartFull( Abc_NtkObjNumMax(pNtk) );
    var vSet *Vec_Int_t  = Vec_IntAlloc( 16 );
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    // compute empty set
    SetId = Hsh_VecManAdd( pHash, vSet );
    assert(SetId == 0, "SetId == 0");
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Vec_IntFill( vSet, 1, Abc_ObjId(pObj) );
        SetId = Hsh_VecManAdd( pHash, vSet );
        Vec_IntWriteEntry( vMap, Abc_ObjId(pObj), SetId );
    }
    // make sure the array of objects is sorted
    Vec_IntSort( vObjs, 0 );
    // begin from the objects and map their IDs into sets of COs
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vObjs) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vObjs,i))), 1); i++  {         if (pObj) == nil {} else {
        Abc_NtkDetectObjClasses_rec( pObj, vMap, pHash, vSet ); }
    }
    Vec_IntFree( vSet );
    // create map for mapping CO set its their classes
    vClassMap = Vec_IntStartFull( Hsh_VecSize(pHash) + 1 );
    // collect classes of objects
    vClasses = Vec_WecAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        //char * pName = Abc_ObjName( Abc_NtkObj(pNtk, iObj) );
        // for a given object (iObj), find the ID of its COs set
        SetId = Vec_IntEntry( vMap, iObj );
        assert(SetId >= 0, "SetId >= 0");
        // for the given CO set, finds its equivalence class
        ClassId = Vec_IntEntry( vClassMap, SetId );
        if ClassId == -1 {  // there is no equivalence class
        
            // map this CO set into a new equivalence class
            Vec_IntWriteEntry( vClassMap, SetId, Vec_WecSize(vClasses) );
            vClass = Vec_WecPushLevel( vClasses );
        
        } else { // get hold of the equivalence class
            vClass = Vec_WecEntry( vClasses, ClassId ); }
        // add objects to the class
        Vec_IntPush( vClass, iObj );
        // print the set for this object
        //printf( "Object %5d : ", iObj );
        //Vec_IntPrint( Hsh_VecReadEntry(pHash, SetId) );
    }
    // collect arrays of COs for each class
    *pvCos = Vec_WecStart( Vec_WecSize(vClasses) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vClasses)) && (((vClass) = Vec_WecEntry(vClasses, i)), 1); i++  {
        iObj = Vec_IntEntry( vClass, 0 );
        // for a given object (iObj), find the ID of its COs set
        SetId = Vec_IntEntry( vMap, iObj );
        assert(SetId >= 0, "SetId >= 0");
        // for the given CO set ID, find the set
        vSet = Hsh_VecReadEntry( pHash, SetId );
        Vec_IntAppend( Vec_WecEntry(*pvCos, i), vSet );
    }
    Hsh_VecManStop( pHash );
    Vec_IntFree( vClassMap );
    Vec_IntFree( vMap );
    return vClasses;
}
func Abc_NtkDetectClassesTest2(pNtk *Abc_Ntk_t, fVerbose int32, fVeryVerbose int32) {
    var vObjs *Vec_Int_t;
    var vRes, vCos *Vec_Wec_t;
    // for testing, create the set of object IDs for all combinational inputs (CIs)
    var pObj *Abc_Obj_t; var i int32;
    vObjs = Vec_IntAlloc( Abc_NtkCiNum(pNtk) );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Vec_IntPush( vObjs, Abc_ObjId(pObj) );
    }
    // compute equivalence classes of CIs and print them
    vRes = Abc_NtkDetectObjClasses( pNtk, vObjs, &vCos );
    Vec_WecPrint( vRes, 0 );
    Vec_WecPrint( vCos, 0 );
    // clean up
    Vec_IntFree( vObjs );
    Vec_WecFree( vRes );
    Vec_WecFree( vCos );
}

/**Function*************************************************************

  Synopsis    [Collecting objects.]

  Description [Collects combinational inputs (vCIs) and internal nodes (vNodes)
  reachable from the given set of combinational outputs (vCOs).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFinMiterCollect_rec(pObj *Abc_Obj_t, vCis *Vec_Int_t, vNodes *Vec_Int_t) {
    if Abc_NodeIsTravIdCurrent(pObj) {
        return; }
    Abc_NodeSetTravIdCurrent(pObj);
    if Abc_ObjIsCi(pObj) {
        Vec_IntPush( vCis, Abc_ObjId(pObj) );
    } else {
        var pFanin *Abc_Obj_t; var i int32;
        assert(Abc_ObjIsNode( pObj ), "Abc_ObjIsNode( pObj )");
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
            Abc_NtkFinMiterCollect_rec( pFanin, vCis, vNodes );
        }
        Vec_IntPush( vNodes, Abc_ObjId(pObj) );
    }
}
func Abc_NtkFinMiterCollect(pNtk *Abc_Ntk_t, vCos *Vec_Int_t, vCis *Vec_Int_t, vNodes *Vec_Int_t) {
    var pObj *Abc_Obj_t; var i int32;
    Vec_IntClear( vCis );
    Vec_IntClear( vNodes );
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCos) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vCos,i))), 1); i++  {         if (pObj) == nil {} else {
        Abc_NtkFinMiterCollect_rec( Abc_ObjFanin0(pObj), vCis, vNodes ); }
    }
}

/**Function*************************************************************

  Synopsis    [Simulates expression using given simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibGateSimulate(pGate *Mio_Gate_t, ppFaninSims []*word, nWords int32, pObjSim *word) {
    var i int32;
    var w int32;
    var nVars int32 = Mio_GateReadPinNum(pGate);
    var vExpr *Vec_Int_t  = Mio_GateReadExpr( pGate );
    assert(nVars <= 6, "nVars <= 6");
    for  w = 0; w < nWords; w++  {
        var uFanins [6]word;
        for  i = 0; i < nVars; i++  {
            uFanins[i] = ppFaninSims[i][w];
        }
        pObjSim[w] = Exp_Truth6( nVars, vExpr, uFanins );
    }
}

/**Function*************************************************************

  Synopsis    [Simulates expression for one simulation pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibGateSimulateOne(pGate *Mio_Gate_t, iBits []int32) int32 {
    var nVars int32  = Mio_GateReadPinNum(pGate);
    var i, iMint int32;
    for  i = 0; i < nVars; i++  {
        if iBits[i] {
            iMint |= (1 << i); }
    }
    return Abc_InfoHasBit( (unsigned *)Mio_GateReadTruthP(pGate), iMint );
}

/**Function*************************************************************

  Synopsis    [Simulated expression with one bit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibGateSimulateGia(pGia *Gia_Man_t, pGate *Mio_Gate_t, iLits []int32, vLits *Vec_Int_t) int32 {
    var i int32;
    var nVars int32 = Mio_GateReadPinNum(pGate);
    var vExpr *Vec_Int_t  = Mio_GateReadExpr( pGate );
    if Exp_IsConst0(vExpr) {
        return 0; }
    if Exp_IsConst1(vExpr) {
        return 1; }
    if Exp_IsLit(vExpr) {
        var Index0 int32   = Vec_IntEntry(vExpr,0) >> 1;
        var fCompl0 int32  = Vec_IntEntry(vExpr,0) & 1;
        assert(Index0 < nVars, "Index0 < nVars");
        return Abc_LitNotCond( iLits[Index0], fCompl0 );
    }
    Vec_IntClear( vLits );
    for  i = 0; i < nVars; i++  {
        Vec_IntPush( vLits, iLits[i] );
    }
    for  i = 0; i < Exp_NodeNum(vExpr); i++  {
        var Index0 int32   = Vec_IntEntry( vExpr, 2*i+0 ) >> 1;
        var Index1 int32   = Vec_IntEntry( vExpr, 2*i+1 ) >> 1;
        var fCompl0 int32  = Vec_IntEntry( vExpr, 2*i+0 ) & 1;
        var fCompl1 int32  = Vec_IntEntry( vExpr, 2*i+1 ) & 1;
        Vec_IntPush( vLits, Gia_ManHashAnd( pGia, Abc_LitNotCond(Vec_IntEntry(vLits, Index0), fCompl0), Abc_LitNotCond(Vec_IntEntry(vLits, Index1), fCompl1) ) );
    }
    return Abc_LitNotCond( Vec_IntEntryLast(vLits), Vec_IntEntryLast(vExpr) & 1 );
}

/**Function*************************************************************

  Synopsis    [AIG construction and simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFinSimOneLit(pNew *Gia_Man_t, pObj *Abc_Obj_t, Type int32, vMap *Vec_Int_t, n int32, vTemp *Vec_Int_t) int32 {
    if Abc_NtkIsMappedLogic(pObj.pNtk) && Type >= 0 {
        extern int Mio_LibGateSimulateGia( Gia_Man_t * pGia, Mio_Gate_t * pGate, int iLits[6], Vec_Int_t * vLits );
        var pLib *Mio_Library_t  = (Mio_Library_t *)pObj.pNtk.pManFunc;
        var i int32;
        var Lits [6]int32;
        for  i = 0; i < Abc_ObjFaninNum(pObj); i++  {
            Lits[i] = Vec_IntEntry( vMap, Abc_Var2Lit(Abc_ObjFaninId(pObj, i), n) );
        }
        return Mio_LibGateSimulateGia( pNew, Mio_LibraryReadGateById(pLib, Type), Lits, vTemp );
    
    } else {
        var iLit0 int32;
        if Abc_ObjFaninNum(pObj) > 0 {
            iLit0 = Vec_IntEntry( vMap, Abc_Var2Lit(Abc_ObjFaninId0(pObj), n) );
        } else {
            iLit0 = -1;
        }
        var iLit1 int32;
        if Abc_ObjFaninNum(pObj) > 1 {
            iLit1 = Vec_IntEntry( vMap, Abc_Var2Lit(Abc_ObjFaninId1(pObj), n) );
        } else {
            iLit1 = -1;
        }
        assert(Type != ABC_FIN_NEG, "Type != ABC_FIN_NEG");
        if Type == ABC_FIN_SA0 {            return 0; }
        if Type == ABC_FIN_SA1 {            return 1; }
        if Type == ABC_FIN_RDOB_BUFF {      return iLit0; }
        if Type == ABC_FIN_RDOB_NOT {       return Abc_LitNot( iLit0 ); }
        if Type == ABC_FIN_RDOB_AND {       return Gia_ManHashAnd( pNew, iLit0, iLit1 ); }
        if Type == ABC_FIN_RDOB_OR {        return Gia_ManHashOr( pNew, iLit0, iLit1 ); }
        if Type == ABC_FIN_RDOB_XOR {       return Gia_ManHashXor( pNew, iLit0, iLit1 ); }
        if Type == ABC_FIN_RDOB_NAND {      return Abc_LitNot(Gia_ManHashAnd( pNew, iLit0, iLit1 )); }
        if Type == ABC_FIN_RDOB_NOR {       return Abc_LitNot(Gia_ManHashOr( pNew, iLit0, iLit1 )); }
        if Type == ABC_FIN_RDOB_NXOR {      return Abc_LitNot(Gia_ManHashXor( pNew, iLit0, iLit1 )); }
        assert(false, "0");
        return -1;
    }
}
func Abc_NtkFinSimOneBit(pObj *Abc_Obj_t, Type int32, vSims *Vec_Wrd_t, nWords int32, iBit int32) int32 {
    if Abc_NtkIsMappedLogic(pObj.pNtk) && Type >= 0 {
        extern int Mio_LibGateSimulateOne( Mio_Gate_t * pGate, int iBits[6] );
        var pLib *Mio_Library_t  = (Mio_Library_t *)pObj.pNtk.pManFunc;
        var i int32;
        var iBits [6]int32;
        for  i = 0; i < Abc_ObjFaninNum(pObj); i++  {
            var pSim0 *word  = Vec_WrdEntryP( vSims, nWords * Abc_ObjFaninId(pObj, i) );
            iBits[i] = Abc_InfoHasBit( (unsigned*)pSim0, iBit );
        }
        return Mio_LibGateSimulateOne( Mio_LibraryReadGateById(pLib, Type), iBits );
    
    } else {
        var pSim0 *word;
        if Abc_ObjFaninNum(pObj) > 0 {
            pSim0 = Vec_WrdEntryP( vSims, nWords * Abc_ObjFaninId0(pObj) );
        } else {
            pSim0 = nil;
        }
        var pSim1 *word;
        if Abc_ObjFaninNum(pObj) > 1 {
            pSim1 = Vec_WrdEntryP( vSims, nWords * Abc_ObjFaninId1(pObj) );
        } else {
            pSim1 = nil;
        }
        var iBit0 int32;
        if Abc_ObjFaninNum(pObj) > 0 {
            iBit0 = Abc_InfoHasBit( (unsigned*)pSim0, iBit );
        } else {
            iBit0 = -1;
        }
        var iBit1 int32;
        if Abc_ObjFaninNum(pObj) > 1 {
            iBit1 = Abc_InfoHasBit( (unsigned*)pSim1, iBit );
        } else {
            iBit1 = -1;
        }
        assert(Type != ABC_FIN_NEG, "Type != ABC_FIN_NEG");
        if Type == ABC_FIN_SA0 {            return 0; }
        if Type == ABC_FIN_SA1 {            return 1; }
        if Type == ABC_FIN_RDOB_BUFF {      return iBit0; }
        if Type == ABC_FIN_RDOB_NOT {       return !iBit0; }
        if Type == ABC_FIN_RDOB_AND {       return iBit0 & iBit1; }
        if Type == ABC_FIN_RDOB_OR {        return iBit0 | iBit1; }
        if Type == ABC_FIN_RDOB_XOR {       return iBit0 ^ iBit1; }
        if Type == ABC_FIN_RDOB_NAND {      return !(iBit0 & iBit1); }
        if Type == ABC_FIN_RDOB_NOR {       return !(iBit0 | iBit1); }
        if Type == ABC_FIN_RDOB_NXOR {      return !(iBit0 ^ iBit1); }
        assert(false, "0");
        return -1;
    }
}
func Abc_NtkFinSimOneWord(pObj *Abc_Obj_t, Type int32, vSims *Vec_Wrd_t, nWords int32) {
    if Abc_NtkIsMappedLogic(pObj.pNtk) {
        extern void Mio_LibGateSimulate( Mio_Gate_t * pGate, word * ppFaninSims[6], int nWords, word * pObjSim );
        var ppSims [6]*word; var i int32;
        var pSim *word  = Vec_WrdEntryP( vSims, nWords * Abc_ObjId(pObj) );
        assert(Type == -1, "Type == -1");
        for  i = 0; i < Abc_ObjFaninNum(pObj); i++  {
            ppSims[i] = Vec_WrdEntryP( vSims, nWords * Abc_ObjFaninId(pObj, i) );
        }
        Mio_LibGateSimulate( (Mio_Gate_t *)pObj.pData, ppSims, nWords, pSim );
    
    } else {
        var pSim *word   = Vec_WrdEntryP( vSims, nWords * Abc_ObjId(pObj) );  var w int32;
        var pSim0 *word;
        if Abc_ObjFaninNum(pObj) > 0 {
            pSim0 = Vec_WrdEntryP( vSims, nWords * Abc_ObjFaninId0(pObj) );
        } else {
            pSim0 = nil;
        }
        var pSim1 *word;
        if Abc_ObjFaninNum(pObj) > 1 {
            pSim1 = Vec_WrdEntryP( vSims, nWords * Abc_ObjFaninId1(pObj) );
        } else {
            pSim1 = nil;
        }
        assert(Type != ABC_FIN_NEG, "Type != ABC_FIN_NEG");
        if Type == ABC_FIN_SA0 {            for  w = 0; w < nWords; w++  { pSim[w] = 0; }
        } else if Type == ABC_FIN_SA1 {       for  w = 0; w < nWords; w++  { pSim[w] = ~((word)0); }
        } else if Type == ABC_FIN_RDOB_BUFF { for  w = 0; w < nWords; w++  { pSim[w] = pSim0[w]; }
        } else if Type == ABC_FIN_RDOB_NOT {  for  w = 0; w < nWords; w++  { pSim[w] = ~pSim0[w]; }
        } else if Type == ABC_FIN_RDOB_AND {  for  w = 0; w < nWords; w++  { pSim[w] = pSim0[w] & pSim1[w]; }
        } else if Type == ABC_FIN_RDOB_OR {   for  w = 0; w < nWords; w++  { pSim[w] = pSim0[w] | pSim1[w]; }
        } else if Type == ABC_FIN_RDOB_XOR {  for  w = 0; w < nWords; w++  { pSim[w] = pSim0[w] ^ pSim1[w]; }
        } else if Type == ABC_FIN_RDOB_NAND { for  w = 0; w < nWords; w++  { pSim[w] = ~(pSim0[w] & pSim1[w]); }
        } else if Type == ABC_FIN_RDOB_NOR {  for  w = 0; w < nWords; w++  { pSim[w] = ~(pSim0[w] | pSim1[w]); }
        } else if Type == ABC_FIN_RDOB_NXOR { for  w = 0; w < nWords; w++  { pSim[w] = ~(pSim0[w] ^ pSim1[w]); }
        } else { assert(false, "0"); }
    }
}


// returns 1 if the functionality with indexes i1 and i2 is the same
func Abc_NtkFinCompareSimTwo(pNtk *Abc_Ntk_t, vCos *Vec_Int_t, vSims *Vec_Wrd_t, nWords int32, i1 int32, i2 int32) int32 {
    var pObj *Abc_Obj_t; var i int32;
    assert(i1 != i2, "i1 != i2");
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCos) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vCos,i))), 1); i++  {         if (pObj) == nil {} else {
        var pSim0 *word  = Vec_WrdEntryP( vSims, nWords * Abc_ObjFaninId0(pObj) );
        if Abc_InfoHasBit((unsigned*)pSim0, i1) != Abc_InfoHasBit((unsigned*)pSim0, i2) {
            return 0; }
    }
    }
    return 1;
}

func Abc_NtkFinMiterToGia(pNtk *Abc_Ntk_t, vTypes *Vec_Int_t, vCos *Vec_Int_t, vCis *Vec_Int_t, vNodes *Vec_Int_t, iObjs []int32, Types []int32, vLits *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t = nil;
    var pTemp *Gia_Man_t;
    var pObj *Abc_Obj_t; 
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 100 );
    var n int32;
    var i int32;
    var Type int32;
    var iMiter int32;
    var iLit int32;
    var pLits *int32;
    // create AIG manager
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( pNtk.pName );
    pNew.pSpec = Abc_UtilStrsav( pNtk.pSpec );
    Gia_ManHashStart( pNew );
    // create inputs
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCis) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vCis,i))), 1); i++  {         if (pObj) == nil {} else {
        iLit = Gia_ManAppendCi(pNew);
        for  n = 0; n < 2; n++  {
            if iObjs[n] != int32(Abc_ObjId(pObj)) {
                Vec_IntWriteEntry( vLits, Abc_Var2Lit(Abc_ObjId(pObj), n), iLit );
            } else if Types[n] != ABC_FIN_NEG {
                Vec_IntWriteEntry( vLits, Abc_Var2Lit(Abc_ObjId(pObj), n), Abc_NtkFinSimOneLit(pNew, pObj, Types[n], vLits, n, vTemp) );
            } else { // if ( iObjs[n] == (int)Abc_ObjId(pObj) && Types[n] == ABC_FIN_NEG )
                Vec_IntWriteEntry( vLits, Abc_Var2Lit(Abc_ObjId(pObj), n), Abc_LitNot(iLit) ); }
        }
    }
    }
    // create internal nodes
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        if Abc_NtkIsMappedLogic(pNtk) {
            Type = Mio_GateReadCell((Mio_Gate_t *)pObj.pData);
        } else {
            Type = Vec_IntEntry(vTypes, Abc_ObjId(pObj));
        }
        for  n = 0; n < 2; n++  {
            if iObjs[n] != int32(Abc_ObjId(pObj)) {
                Vec_IntWriteEntry( vLits, Abc_Var2Lit(Abc_ObjId(pObj), n), Abc_NtkFinSimOneLit(pNew, pObj, Type, vLits, n, vTemp) );
            } else if Types[n] != ABC_FIN_NEG {
                Vec_IntWriteEntry( vLits, Abc_Var2Lit(Abc_ObjId(pObj), n), Abc_NtkFinSimOneLit(pNew, pObj, Types[n], vLits, n, vTemp) );
            } else { // if ( iObjs[n] == (int)Abc_ObjId(pObj) && Types[n] == ABC_FIN_NEG )
                Vec_IntWriteEntry( vLits, Abc_Var2Lit(Abc_ObjId(pObj), n), Abc_LitNot(Abc_NtkFinSimOneLit(pNew, pObj, Type, vLits, n, vTemp)) ); }
        }
    }
    }
    // create comparator
    iMiter = 0;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCos) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vCos,i))), 1); i++  {         if (pObj) == nil {} else {
        pLits  = Vec_IntEntryP( vLits, Abc_Var2Lit(Abc_ObjFaninId0(pObj), 0) );
        iLit   = Gia_ManHashXor( pNew, pLits[0], pLits[1] );
        iMiter = Gia_ManHashOr( pNew, iMiter, iLit );
    }
    }
    Gia_ManAppendCo( pNew, iMiter );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    Vec_IntFree( vTemp );
    return pNew;
}
func Abc_NtkFinSimulateOne(pNtk *Abc_Ntk_t, vTypes *Vec_Int_t, vCos *Vec_Int_t, vCis *Vec_Int_t, vNodes *Vec_Int_t, vMap2 *Vec_Wec_t, vPat *Vec_Int_t, vSims *Vec_Wrd_t, nWords int32, vPairs *Vec_Int_t, vRes *Vec_Wec_t, iLevel int32, iItem int32) {
    var pObj *Abc_Obj_t; 
    var vClass, vArray *Vec_Int_t;
    var i, Counter int32;
    var nItems int32  = Vec_WecSizeSize(vRes);
    assert(nItems == Vec_WecSizeSize(vMap2), "nItems == Vec_WecSizeSize(vMap2)");
    assert(nItems <= 128 * nWords, "nItems <= 128 * nWords");
    // assign inputs
    assert(Vec_IntSize(vPat) == Vec_IntSize(vCis), "Vec_IntSize(vPat) == Vec_IntSize(vCis)");
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCis) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vCis,i))), 1); i++  {         if (pObj) == nil {} else {
        var w int32;
        var iObj int32 = Abc_ObjId( pObj );
        var Init word;
        if Vec_IntEntry(vPat, i) {
            Init = ~((word)0);
        } else {
            Init = 0;
        }
        var pSim *word  = Vec_WrdEntryP( vSims, nWords * Abc_ObjId(pObj) );
        for  w = 0; w < nWords; w++  {
            pSim[w] = Init;
        }
        vArray = Vec_WecEntry(vMap2, iObj);
        if Vec_IntSize(vArray) > 0 {
            var k, iFin, Index, iObj, Type int32;
            // inlining of macro Vec_IntForEachEntryDouble
            for  k = 0; (k+1 < Vec_IntSize(vArray)) && (((iFin) = Vec_IntEntry(vArray, k)), 1) && (((Index) = Vec_IntEntry(vArray, k+1)), 1); k += 2  {
                assert(Index < 64, "Index < 64");
                iObj = Vec_IntEntry( vPairs, 2*iFin );
                assert(iObj == int32(Abc_ObjId(pObj)), "iObj == int32(Abc_ObjId(pObj))");
                Type = Vec_IntEntry( vPairs, 2*iFin+1 );
                assert(Type == ABC_FIN_NEG || Type == ABC_FIN_SA0 || Type == ABC_FIN_SA1, "Type == ABC_FIN_NEG || Type == ABC_FIN_SA0 || Type == ABC_FIN_SA1");
                if Type == ABC_FIN_NEG || Abc_InfoHasBit((unsigned *)pSim, Index) != Abc_NtkFinSimOneBit(pObj, Type, vSims, nWords, Index) {
                    Abc_InfoXorBit( (unsigned *)pSim, Index ); }
                Counter++;
            }
        }
    }
    }
    // simulate internal nodes
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        var iObj int32  = Abc_ObjId( pObj );
        var Type int32;
        if Abc_NtkIsMappedLogic(pNtk) {
            Type = -1;
        } else {
            Type = Vec_IntEntry( vTypes, iObj );
        }
        var pSim *word  = Vec_WrdEntryP( vSims, nWords * Abc_ObjId(pObj) );
        Abc_NtkFinSimOneWord( pObj, Type, vSims, nWords );
        vArray = Vec_WecEntry(vMap2, iObj);
        if Vec_IntSize(vArray) > 0 {
            var k, iFin, Index, iObj, Type int32;
            // inlining of macro Vec_IntForEachEntryDouble
            for  k = 0; (k+1 < Vec_IntSize(vArray)) && (((iFin) = Vec_IntEntry(vArray, k)), 1) && (((Index) = Vec_IntEntry(vArray, k+1)), 1); k += 2  {
                assert(Index < 64 * nWords, "Index < 64 * nWords");
                iObj = Vec_IntEntry( vPairs, 2*iFin );
                assert(iObj == int32(Abc_ObjId(pObj)), "iObj == int32(Abc_ObjId(pObj))");
                Type = Vec_IntEntry( vPairs, 2*iFin+1 );
                if Type == ABC_FIN_NEG || Abc_InfoHasBit((unsigned *)pSim, Index) != Abc_NtkFinSimOneBit(pObj, Type, vSims, nWords, Index) {
                    Abc_InfoXorBit( (unsigned *)pSim, Index ); }
                Counter++;
            }
        }
    }
    }
    assert(nItems == 2*Counter, "nItems == 2*Counter");

    // confirm no refinement
    // inlining of macro Vec_WecForEachLevelStop
    for  i = 0; (i < iLevel+1) && (((vClass) = Vec_WecEntry(vRes, i)), 1); i++  {
        var k, iFin, Index, Value int32;
        var Index0 int32  = Vec_IntEntry( vClass, 1 );
        // inlining of macro Vec_IntForEachEntryDoubleStart
        for  k = 2; (k+1 < Vec_IntSize(vClass)) && (((iFin) = Vec_IntEntry(vClass, k)), 1) && (((Index) = Vec_IntEntry(vClass, k+1)), 1); k += 2  {
            if i == iLevel && k/2 >= iItem {
                break; }
            //printf( "Double-checking pair %d and %d\n", iFin0, iFin );
            Value = Abc_NtkFinCompareSimTwo( pNtk, vCos, vSims, nWords, Index0, Index );
            assert((Value != 0), "Value"); // the same value
        }
    }

    // check refinement
    // inlining of macro Vec_WecForEachLevelStart
    for  i = iLevel; (i < Vec_WecSize(vRes)) && (((vClass) = Vec_WecEntry(vRes, i)), 1); i++  {
        var k int32;
        var iFin int32;
        var Index int32;
        var Value int32;
        var Index0 int32 = Vec_IntEntry(vClass, 1);
        var j int32;
        if i == iLevel {
            j = 2*iItem;
        } else {
            j = 2;
        }
        var vNewClass *Vec_Int_t  = nil;
        // inlining of macro Vec_IntForEachEntryDoubleStart
        for  k = j; (k+1 < Vec_IntSize(vClass)) && (((iFin) = Vec_IntEntry(vClass, k)), 1) && (((Index) = Vec_IntEntry(vClass, k+1)), 1); k += 2  {
            Value = Abc_NtkFinCompareSimTwo( pNtk, vCos, vSims, nWords, Index0, Index );
            if Value != 0 {  // the same value
            
                Vec_IntWriteEntry( vClass, j++, iFin );
                Vec_IntWriteEntry( vClass, j++, Index );
                continue;
            }
            // create new class
            if vNewClass != nil {
                vNewClass = vNewClass;
            } else {
                vNewClass = Vec_WecPushLevel( vRes );
            }
            Vec_IntPushTwo( vNewClass, iFin, Index );  // index and first entry
            vClass = Vec_WecEntry( vRes, i );
        }
        Vec_IntShrink( vClass, j );
    }
}

/**Function*************************************************************

  Synopsis    [Check equivalence using SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFinCheckPair(pNtk *Abc_Ntk_t, vTypes *Vec_Int_t, vCos *Vec_Int_t, vCis *Vec_Int_t, vNodes *Vec_Int_t, iObjs []int32, Types []int32, vLits *Vec_Int_t) *Vec_Int_t {
    var pGia *Gia_Man_t  = Abc_NtkFinMiterToGia( pNtk, vTypes, vCos, vCis, vNodes, iObjs, Types, vLits );
    if Gia_ManAndNum(pGia) == 0 && Gia_ObjIsConst0(Gia_ObjFanin0(Gia_ManCo(pGia, 0))) {
        var vPat *Vec_Int_t;
        if Gia_ObjFaninC0(Gia_ManCo(pGia, 0)) {
            vPat = Vec_IntStart(Vec_IntSize(vCis));
        } else {
            vPat = nil;
        }
        Gia_ManStop( pGia );
        return vPat;
    
    } else {
        var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( pGia, 8, 0, 1, 0, 0 );
        var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
        if pSat == nil {
            Gia_ManStop( pGia );
            Cnf_DataFree( pCnf );
            return nil;
        
        } else {
            var i int32;
            var nConfLimit int32 = 10000;
            var vPat *Vec_Int_t  = nil;
            var status int32;
            var iVarBeg int32 = pCnf.nVars - Gia_ManPiNum(pGia);// - 1;
            //Gia_AigerWrite( pGia, "temp_detect.aig", 0, 0, 0 );
            Gia_ManStop( pGia );
            Cnf_DataFree( pCnf );
            status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, 0, 0, 0 );
            if status == l_Undef {
                vPat = Vec_IntAlloc(0);
            } else if status == l_True {
                vPat = Vec_IntAlloc( Vec_IntSize(vCis) );
                for  i = 0; i < Vec_IntSize(vCis); i++  {
                    Vec_IntPush( vPat, sat_solver_var_value(pSat, iVarBeg+i) );
                }
            }
            //printf( "%d ", sat_solver_nconflicts(pSat) );
            sat_solver_delete( pSat );
            return vPat;
        }
    }
}


/**Function*************************************************************

  Synopsis    [Refinement of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFinLocalSetup(vPairs *Vec_Int_t, vList *Vec_Int_t, vMap2 *Vec_Wec_t, vResArray *Vec_Int_t) {
    var i, iFin int32;
    Vec_IntClear( vResArray );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vList)) && (((iFin) = Vec_IntEntry(vList, i)), 1); i++  {
        var iObj int32  = Vec_IntEntry( vPairs, 2*iFin );
        var vArray *Vec_Int_t  = Vec_WecEntry( vMap2, iObj );
        Vec_IntPushTwo( vArray, iFin, i );
        Vec_IntPushTwo( vResArray, iFin, i );
    }
}
func Abc_NtkFinLocalSetdown(vPairs *Vec_Int_t, vList *Vec_Int_t, vMap2 *Vec_Wec_t) {
    var i, iFin int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vList)) && (((iFin) = Vec_IntEntry(vList, i)), 1); i++  {
        var iObj int32  = Vec_IntEntry( vPairs, 2*iFin );
        var vArray *Vec_Int_t  = Vec_WecEntry( vMap2, iObj );
        Vec_IntClear( vArray );
    }
}
func Abc_NtkFinRefinement(pNtk *Abc_Ntk_t, vTypes *Vec_Int_t, vCos *Vec_Int_t, vCis *Vec_Int_t, vNodes *Vec_Int_t, vPairs *Vec_Int_t, vList *Vec_Int_t, vMap2 *Vec_Wec_t, vResult *Vec_Wec_t) int32 {
    var vRes *Vec_Wec_t   = Vec_WecAlloc( 100 );
    var nWords int32  = Abc_Bit6WordNum( Vec_IntSize(vList) );
    var vSims *Vec_Wrd_t  = Vec_WrdStart( nWords * Abc_NtkObjNumMax(pNtk) );  // simulation info for each object
    var vLits *Vec_Int_t  = Vec_IntStart( 2*Abc_NtkObjNumMax(pNtk) );         // two literals for each object
    var vPat, vClass, vArray *Vec_Int_t;  
    var i, k, iFin, Index, nCalls int32;
    // prepare
    vArray = Vec_WecPushLevel( vRes );
    Abc_NtkFinLocalSetup( vPairs, vList, vMap2, vArray );
    // try all-0/all-1 pattern
    for  i = 0; i < 2; i++  {
        vPat = Vec_IntAlloc( Vec_IntSize(vCis) );
        Vec_IntFill( vPat, Vec_IntSize(vCis), i );
        Abc_NtkFinSimulateOne( pNtk, vTypes, vCos, vCis, vNodes, vMap2, vPat, vSims, nWords, vPairs, vRes, 0, 1 );
        Vec_IntFree( vPat );
    }
    // explore the classes
    //Vec_WecPrint( vRes, 0 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vRes)) && (((vClass) = Vec_WecEntry(vRes, i)), 1); i++  {
        var iFin0 int32   = Vec_IntEntry( vClass, 0 );
        // inlining of macro Vec_IntForEachEntryDoubleStart
        for  k = 2; (k+1 < Vec_IntSize(vClass)) && (((iFin) = Vec_IntEntry(vClass, k)), 1) && (((Index) = Vec_IntEntry(vClass, k+1)), 1); k += 2  {
            var Objs [2]int32 = [2]int32{ Vec_IntEntry(vPairs, 2*iFin0),   Vec_IntEntry(vPairs, 2*iFin)   }; 
            var Types [2]int32 = [2]int32{ Vec_IntEntry(vPairs, 2*iFin0+1), Vec_IntEntry(vPairs, 2*iFin+1) }; 
            nCalls++;
            //printf( "Checking pair %d and %d.\n", iFin0, iFin );
            vPat = Abc_NtkFinCheckPair( pNtk, vTypes, vCos, vCis, vNodes, Objs, Types, vLits );
            if vPat == nil { // proved
                continue; }
            assert(Vec_IntEntry(vClass, k) == iFin, "Vec_IntEntry(vClass, k) == iFin");
            if Vec_IntSize(vPat) == 0 {
                var vNewClass *Vec_Int_t  = Vec_WecPushLevel( vRes );
                Vec_IntPushTwo( vNewClass, iFin, Index );  // index and first entry
                vClass = Vec_WecEntry( vRes, i );
                Vec_IntDrop( vClass, k+1 );
                Vec_IntDrop( vClass, k );
            
            } else { // resimulate and refine
                Abc_NtkFinSimulateOne( pNtk, vTypes, vCos, vCis, vNodes, vMap2, vPat, vSims, nWords, vPairs, vRes, i, k/2 ); }
            Vec_IntFree( vPat );
            // make sure refinement happened (k'th entry is now absent or different)
            vClass = Vec_WecEntry( vRes, i );
            assert(Vec_IntSize(vClass) <= k || Vec_IntEntry(vClass, k) != iFin, "Vec_IntSize(vClass) <= k || Vec_IntEntry(vClass, k) != iFin");
            k -= 2;
            //Vec_WecPrint( vRes, 0 );
        }
    }
    // unprepare
    Abc_NtkFinLocalSetdown( vPairs, vList, vMap2 );
    // reload proved equivs into the final array
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vRes)) && (((vArray) = Vec_WecEntry(vRes, i)), 1); i++  {
        assert(Vec_IntSize(vArray) % 2 == 0, "Vec_IntSize(vArray) % 2 == 0");
        if Vec_IntSize(vArray) <= 2 {
            continue; }
        vClass = Vec_WecPushLevel( vResult );
        // inlining of macro Vec_IntForEachEntryDouble
        for  k = 0; (k+1 < Vec_IntSize(vArray)) && (((iFin) = Vec_IntEntry(vArray, k)), 1) && (((Index) = Vec_IntEntry(vArray, k+1)), 1); k += 2  {
            Vec_IntPush( vClass, iFin );
        }
    }
    Vec_WecFree( vRes );
    Vec_WrdFree( vSims );
    Vec_IntFree( vLits );
    return nCalls;
}

/**Function*************************************************************

  Synopsis    [Detecting classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjFinGateType(pNode *Abc_Obj_t) int32 {
    var pSop string  = (char *)pNode.pData;
    if !(int32(strings.Compare(pSop, "1 1\n")) != 0) {         return ABC_FIN_RDOB_BUFF; }
    if !(int32(strings.Compare(pSop, "0 1\n")) != 0) {         return ABC_FIN_RDOB_NOT; }
    if !(int32(strings.Compare(pSop, "11 1\n")) != 0) {        return ABC_FIN_RDOB_AND; }
    if !(int32(strings.Compare(pSop, "11 0\n")) != 0) {        return ABC_FIN_RDOB_NAND; }
    if !(int32(strings.Compare(pSop, "00 0\n")) != 0) {        return ABC_FIN_RDOB_OR; }
    if !(int32(strings.Compare(pSop, "00 1\n")) != 0) {        return ABC_FIN_RDOB_NOR; }
    if !(int32(strings.Compare(pSop, "01 1\n10 1\n")) != 0) {  return ABC_FIN_RDOB_XOR; }
    if !(int32(strings.Compare(pSop, "11 1\n00 1\n")) != 0) {  return ABC_FIN_RDOB_NXOR; }
    return ABC_FIN_NONE;
}
func Abc_NtkFinCheckTypesOk(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t; var i int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
        if Abc_ObjFinGateType(pObj) == ABC_FIN_NONE {
            return i; }
    }
    return 0;
}
func Abc_NtkFinCheckTypesOk2(pNtk *Abc_Ntk_t) int32 {
    var pLib *Mio_Library_t  = (Mio_Library_t *)pNtk.pManFunc;
    var i, iObj, Type int32;
    // inlining of macro Vec_IntForEachEntryDoubleStart
    for  i = 2; (i+1 < Vec_IntSize(pNtk.vFins)) && (((iObj) = Vec_IntEntry(pNtk.vFins, i)), 1) && (((Type) = Vec_IntEntry(pNtk.vFins, i+1)), 1); i += 2  {
        var pObj *Abc_Obj_t  = Abc_NtkObj( pNtk, iObj );
        var pGateFlt *Mio_Gate_t;
        var pGateObj *Mio_Gate_t = (Mio_Gate_t *)pObj.pData;
        if Type < 0 { // SA0, SA1, NEG
            continue; }
        pGateFlt = Mio_LibraryReadGateById( pLib, Type );
        if Mio_GateReadPinNum(pGateFlt) < 1 {
            continue; }
        if Mio_GateReadPinNum(pGateObj) != Mio_GateReadPinNum(pGateFlt) {
            return iObj; }
    }
    return 0;
}
func Abc_NtkFinComputeTypes(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var pObj *Abc_Obj_t; var i int32;
    var vObjs *Vec_Int_t  = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Vec_IntWriteEntry( vObjs, Abc_ObjId(pObj), Abc_ObjFinGateType(pObj) ); }
    }
    return vObjs;
}
func Abc_NtkFinComputeObjects(vPairs *Vec_Int_t, pvMap **Vec_Wec_t, nObjs int32) *Vec_Int_t {
    var i, iObj, Type int32;
    var vObjs *Vec_Int_t  = Vec_IntAlloc( 100 );
    *pvMap = Vec_WecStart( nObjs );
    // inlining of macro Vec_IntForEachEntryDoubleStart
    for  i = 2; (i+1 < Vec_IntSize(vPairs)) && (((iObj) = Vec_IntEntry(vPairs, i)), 1) && (((Type) = Vec_IntEntry(vPairs, i+1)), 1); i += 2  {
        Vec_IntPush( vObjs, iObj );
        Vec_WecPush( *pvMap, iObj, i/2 );
    }
    Vec_IntUniqify( vObjs );
    return vObjs;
}
func Abc_NtkFinCreateList(vMap *Vec_Wec_t, vClass *Vec_Int_t) *Vec_Int_t {
    var i, iObj int32;
    var vList *Vec_Int_t  = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClass)) && (((iObj) = Vec_IntEntry(vClass, i)), 1); i++  {
        Vec_IntAppend( vList, Vec_WecEntry(vMap, iObj) );
    }
    return vList;
}
func Abc_NtkFinCountPairs(vClasses *Vec_Wec_t) int32 {
    var i, Counter int32;
    var vLevel *Vec_Int_t;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vClasses)) && (((vLevel) = Vec_WecEntry(vClasses, i)), 1); i++  {
        Counter += Vec_IntSize(vLevel) - 1;
    }
    return Counter;
}
func Abc_NtkDetectFinClasses(pNtk *Abc_Ntk_t, fVerbose int32) *Vec_Wec_t {
    var vTypes *Vec_Int_t  = nil; // gate types
    var vPairs *Vec_Int_t;        // original info as a set of pairs (ObjId, TypeId)
    var vObjs *Vec_Int_t;         // all those objects that have some fin 
    var vMap *Vec_Wec_t;          // for each object, the set of fins
    var vMap2 *Vec_Wec_t;         // for each local object, the set of pairs (Info, Index)
    var vClasses *Vec_Wec_t;      // classes of objects
    var vCoSets *Vec_Wec_t;       // corresponding CO sets
    var vClass *Vec_Int_t;        // one class
    var vCoSet *Vec_Int_t;        // one set of COs
    var vCiSet *Vec_Int_t;        // one set of CIs
    var vNodeSet *Vec_Int_t;      // one set of nodes
    var vList *Vec_Int_t;         // one info list
    var vResult *Vec_Wec_t;       // resulting equivalences
    var i, iObj, nCalls int32;
    if pNtk.vFins == nil {
        fmt.Printf("Current network does not have the required info.\n");
        return nil;
    }
    assert(Abc_NtkIsSopLogic(pNtk) || Abc_NtkIsMappedLogic(pNtk), "Abc_NtkIsSopLogic(pNtk) || Abc_NtkIsMappedLogic(pNtk)");
    if Abc_NtkIsSopLogic(pNtk) {
        iObj = Abc_NtkFinCheckTypesOk(pNtk);
        if iObj != 0 {
            fmt.Printf("Current network contains unsupported gate types (for example, see node \"%s\").\n", Abc_ObjName(Abc_NtkObj(pNtk, iObj)));
            return nil;
        }
        vTypes   = Abc_NtkFinComputeTypes( pNtk );
    
    } else if Abc_NtkIsMappedLogic(pNtk) {
        iObj = Abc_NtkFinCheckTypesOk2(pNtk);
        if iObj != 0 {
            fmt.Printf("Current network has mismatch between mapped gate size and fault gate size (for example, see node \"%s\").\n", Abc_ObjName(Abc_NtkObj(pNtk, iObj)));
            return nil;
        }
    
    } else { assert(false, "0"); }
    //Abc_NtkFrameExtend( pNtk );
    // collect data
    vPairs   = pNtk.vFins;
    vObjs    = Abc_NtkFinComputeObjects( vPairs, &vMap, Abc_NtkObjNumMax(pNtk) );
    vClasses = Abc_NtkDetectObjClasses( pNtk, vObjs, &vCoSets );
    // refine classes
    vCiSet   = Vec_IntAlloc( 1000 );
    vNodeSet = Vec_IntAlloc( 1000 );
    vMap2    = Vec_WecStart( Abc_NtkObjNumMax(pNtk) );
    vResult  = Vec_WecAlloc( 1000 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vClasses)) && (((vClass) = Vec_WecEntry(vClasses, i)), 1); i++  {
        // extract one window
        vCoSet = Vec_WecEntry( vCoSets, i );
        Abc_NtkFinMiterCollect( pNtk, vCoSet, vCiSet, vNodeSet );
        // refine one class
        vList = Abc_NtkFinCreateList( vMap, vClass );
        nCalls = Abc_NtkFinRefinement( pNtk, vTypes, vCoSet, vCiSet, vNodeSet, vPairs, vList, vMap2, vResult );
        if fVerbose != 0 {
            fmt.Printf("Group %4d :  Obj =%4d. Fins =%4d.  CI =%5d. CO =%5d. Node =%6d.  SAT calls =%5d.\n", i, Vec_IntSize(vClass), Vec_IntSize(vList), Vec_IntSize(vCiSet), Vec_IntSize(vCoSet), Vec_IntSize(vNodeSet), nCalls); }
        Vec_IntFree( vList );
    }
    // sort entries in each array
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vResult)) && (((vClass) = Vec_WecEntry(vResult, i)), 1); i++  {
        Vec_IntSort( vClass, 0 );
    }
    // sort by the index of the first entry
    Vec_WecSortByFirstInt( vResult, 0 ); 
    // cleanup
    Vec_IntFreeP( & vTypes );
    Vec_IntFree( vObjs );
    Vec_WecFree( vClasses );
    Vec_WecFree( vMap );
    Vec_WecFree( vMap2 );
    Vec_WecFree( vCoSets );
    Vec_IntFree( vCiSet );
    Vec_IntFree( vNodeSet );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Print results.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintFinResults(vClasses *Vec_Wec_t) {
    var vClass *Vec_Int_t;
    var i, k, Entry int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vClasses)) && (((vClass) = Vec_WecEntry(vClasses, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vClass)) && (((Entry) = Vec_IntEntry(vClass, k)), 1); k++  {
            fmt.Printf("%d %d\n", Vec_IntEntry(vClass, 0), Entry);
        }
    }
}

/**Function*************************************************************

  Synopsis    [Top-level procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDetectClassesTest(pNtk *Abc_Ntk_t, fSeq int32, fVerbose int32, fVeryVerbose int32) {
    var vResult *Vec_Wec_t;
    var clk abctime  = Abc_Clock();
    if fSeq != 0 { 
        Abc_NtkFrameExtend( pNtk ); }
    vResult = Abc_NtkDetectFinClasses( pNtk, fVerbose );
    fmt.Printf("Computed %d equivalence classes with %d item pairs.  ", Vec_WecSize(vResult), Abc_NtkFinCountPairs(vResult));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if fVeryVerbose != 0 {
        Vec_WecPrint( vResult, 1 ); }
//    if ( fVerbose )
//        Abc_NtkPrintFinResults( vResult );
    Vec_WecFree( vResult );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

