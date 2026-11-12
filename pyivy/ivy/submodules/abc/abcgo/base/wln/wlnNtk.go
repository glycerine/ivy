package wln

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [wlnNtk.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Network construction procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnNtk.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creating networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_NtkAlloc(pName string, nObjsMax int32) *Wln_Ntk_t {
     var i int32;
    // inlining of macro ABC_CALLOC
    p := &Wln_Ntk_t{};
    if pName {
        p.pName = Extra_FileNameGeneric( pName );
    } else {
        p.pName = nil;
    }
    Vec_IntGrow( &p.vCis, 111 );
    Vec_IntGrow( &p.vCos, 111 );
    Vec_IntGrow( &p.vFfs, 111 );
    Vec_IntGrow( &p.vTypes,  nObjsMax+1 );
    Vec_StrGrow( &p.vSigns,  nObjsMax+1 );
    Vec_IntGrow( &p.vRanges, nObjsMax+1 );
    Vec_IntPush( &p.vTypes,  -1 );
    Vec_StrPush( &p.vSigns,  -1 );
    Vec_IntPush( &p.vRanges, -1 );
    // inlining of macro ABC_CALLOC
    p.vFanins = ((Wln_Vec_t *) calloc((size_t)(nObjsMax+1), sizeof(Wln_Vec_t)));
    p.pRanges = Hash_IntManStart( 1000 );
    for  i = 0; i < 65; i++  {
        Hash_Int2ManInsert( p.pRanges, i, i, 0 );
    }
    for  i = 1; i < 64; i++  {
        Hash_Int2ManInsert( p.pRanges, i, 0, 0 );
    }
    assert(Hash_IntManEntryNum(p.pRanges) == 128, "Hash_IntManEntryNum(p.pRanges) == 128");
    return p;
}
func Wln_NtkFree(p *Wln_Ntk_t) {
    var i int32;
    for  i = 0; i < Wln_NtkObjNum(p); i++  {
        if Wln_ObjFaninNum(p, i) > 2 { 
            // inlining of macro ABC_FREE
            ((p.vFanins[i].pArray[0]) ? (free((char *) (p.vFanins[i].pArray[0])), (p.vFanins[i].pArray[0]) = 0) : 0); }
    }
    // inlining of macro ABC_FREE
    ((p.vFanins) ? (free((char *) (p.vFanins)), (p.vFanins) = 0) : 0);

    if p.pRanges {  Hash_IntManStop( p.pRanges ); }
    if p.pManName { Abc_NamStop( p.pManName ); }

    // inlining of macro ABC_FREE
    ((p.vCis.pArray) ? (free((char *) (p.vCis.pArray)), (p.vCis.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCos.pArray) ? (free((char *) (p.vCos.pArray)), (p.vCos.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vFfs.pArray) ? (free((char *) (p.vFfs.pArray)), (p.vFfs.pArray) = 0) : 0);

    // inlining of macro ABC_FREE
    ((p.vTypes.pArray) ? (free((char *) (p.vTypes.pArray)), (p.vTypes.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vSigns.pArray) ? (free((char *) (p.vSigns.pArray)), (p.vSigns.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vRanges.pArray) ? (free((char *) (p.vRanges.pArray)), (p.vRanges.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vNameIds.pArray) ? (free((char *) (p.vNameIds.pArray)), (p.vNameIds.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vInstIds.pArray) ? (free((char *) (p.vInstIds.pArray)), (p.vInstIds.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vTravIds.pArray) ? (free((char *) (p.vTravIds.pArray)), (p.vTravIds.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCopies.pArray) ? (free((char *) (p.vCopies.pArray)), (p.vCopies.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vBits.pArray) ? (free((char *) (p.vBits.pArray)), (p.vBits.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vLevels.pArray) ? (free((char *) (p.vLevels.pArray)), (p.vLevels.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vRefs.pArray) ? (free((char *) (p.vRefs.pArray)), (p.vRefs.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vFanout.pArray) ? (free((char *) (p.vFanout.pArray)), (p.vFanout.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vFaninAttrs.pArray) ? (free((char *) (p.vFaninAttrs.pArray)), (p.vFaninAttrs.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vFaninLists.pArray) ? (free((char *) (p.vFaninLists.pArray)), (p.vFaninLists.pArray) = 0) : 0);

    // inlining of macro ABC_FREE
    ((p.pName) ? (free((char *) (p.pName)), (p.pName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSpec) ? (free((char *) (p.pSpec)), (p.pSpec) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Wln_NtkMemUsage(p *Wln_Ntk_t) int32 {
    var Mem int32  = sizeof(Wln_Ntk_t);
    Mem += 4 * p.vCis.nCap;
    Mem += 4 * p.vCos.nCap;
    Mem += 4 * p.vFfs.nCap;
    Mem += 1 * p.vTypes.nCap;
    Mem += 4 * p.vRanges.nCap;
    Mem += 4 * p.vNameIds.nCap;
    Mem += 4 * p.vInstIds.nCap;
    Mem += 4 * p.vTravIds.nCap;
    Mem += 4 * p.vCopies.nCap;
    Mem += 4 * p.vBits.nCap;
    Mem += 4 * p.vLevels.nCap;
    Mem += 4 * p.vRefs.nCap;
    Mem += 4 * p.vFanout.nCap;
    Mem += 4 * p.vFaninAttrs.nCap;
    Mem += 4 * p.vFaninLists.nCap;
    Mem += 20 * Hash_IntManEntryNum(p.pRanges);
    Mem += Abc_NamMemUsed(p.pManName);
    return Mem;
}
func Wln_NtkPrint(p *Wln_Ntk_t) {
    var iObj int32;
    fmt.Printf("Printing %d objects of network \"%s\":\n", Wln_NtkObjNum(p), p.pName);
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        Wln_ObjPrint( p, iObj );
    }
    fmt.Printf("\n");
}



/**Function*************************************************************

  Synopsis    [Detects combinational loops.]

  Description [This procedure is based on the idea suggested by Donald Chai. 
  As we traverse the network and visit the nodes, we need to distinquish 
  three types of nodes: (1) those that are visited for the first time, 
  (2) those that have been visited in this traversal but are currently not 
  on the traversal path, (3) those that have been visited and are currently 
  on the travesal path. When the node of type (3) is encountered, it means 
  that there is a combinational loop. To mark the three types of nodes, 
  two new values of the traversal IDs are used.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_NtkIsAcyclic_rec(p *Wln_Ntk_t, iObj int32) int32 {
    var i, iFanin int32;
    //printf( "Visiting node %d\n", iObj );
    // skip the node if it is already visited
    if Wln_ObjIsTravIdPrevious(p, iObj) {
        return 1; }
    // check if the node is part of the combinational loop
    if Wln_ObjIsTravIdCurrent(p, iObj) {
        fmt.Fprintf(os.Stdout, "Network contains combinational loop!\n");
        fmt.Fprintf(os.Stdout, "Node %16s is encountered twice on the following path:\n", Wln_ObjName(p, iObj));
        fmt.Fprintf(os.Stdout, "Node %16s (ID %6d) of type %5s (type ID %2d) ->\n", Wln_ObjName(p, iObj), iObj, Abc_OperName(Wln_ObjType(p, iObj)), Wln_ObjType(p, iObj));
        return 0;
    }
    // mark this node as a node on the current path
    Wln_ObjSetTravIdCurrent( p, iObj );
    // quit if it is a CI or constant node
    if Wln_ObjIsCi(p, iObj) || Wln_ObjIsFf(p, iObj) || Wln_ObjFaninNum(p, iObj) == 0 {
        // mark this node as a visited node
        Wln_ObjSetTravIdPrevious( p, iObj );
        return 1;
    }
    // traverse the fanin's cone searching for the loop
    // inlining of macro Wln_ObjForEachFanin
    for  i = 0; (i < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, i)), 1); i++  { if !(iFanin != 0) {} else {
        if !(Wln_NtkIsAcyclic_rec(p, iFanin) != 0) {
            // return as soon as the loop is detected
            fmt.Fprintf(os.Stdout, "Node %16s (ID %6d) of type %5s (type ID %2d) ->\n", Wln_ObjName(p, iObj), iObj, Abc_OperName(Wln_ObjType(p, iObj)), Wln_ObjType(p, iObj));
            return 0;
        }
    }
    }
    // mark this node as a visited node
    Wln_ObjSetTravIdPrevious( p, iObj );
    return 1;
}
func Wln_NtkIsAcyclic(p *Wln_Ntk_t) int32 {
    var fAcyclic, i, iObj, nUnvisited int32;
    // set the traversal ID for this DFS ordering
    Wln_NtkIncrementTravId( p );   
    Wln_NtkIncrementTravId( p );   
    // iObj->TravId == pNet->nTravIds      means "iObj is on the path"
    // iObj->TravId == pNet->nTravIds - 1  means "iObj is visited but is not on the path"
    // iObj->TravId <  pNet->nTravIds - 1  means "iObj is not visited"
    // traverse the network to detect cycles
    fAcyclic = 1;
    // inlining of macro Wln_NtkForEachCo
    for  i = 0; (i < Wln_NtkCoNum(p)) && (((iObj) = Wln_NtkCo(p, i)), 1); i++  {
        // traverse the output logic cone
        if fAcyclic = Wln_NtkIsAcyclic_rec(p, iObj) {
            continue; }
        // stop as soon as the first loop is detected
        fmt.Fprintf(os.Stdout, "Primary output %16s (ID %6d)\n", Wln_ObjName(p, iObj), iObj);
        goto finish;
    }
    // inlining of macro Wln_NtkForEachFf
    for  i = 0; (i < Wln_NtkFfNum(p)) && (((iObj) = Wln_NtkFf(p, i)), 1); i++  {
        // traverse the output logic cone
        if fAcyclic = Wln_NtkIsAcyclic_rec(p, iObj) {
            continue; }
        // stop as soon as the first loop is detected
        fmt.Fprintf(os.Stdout, "Flip-flop %16s (ID %6d)\n", Wln_ObjName(p, iObj), iObj);
        goto finish;
    }
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        nUnvisited += !Wln_ObjIsTravIdPrevious(p, iObj) && !Wln_ObjIsCi(p, iObj);
    }
    if nUnvisited != 0 {
        var nSinks int32;
        Wln_NtkCreateRefs(p);
        fmt.Printf("The network has %d objects and %d (%6.2f %%) of them are not connected to the outputs.\n", Wln_NtkObjNum(p), nUnvisited, 100.0*nUnvisited/Wln_NtkObjNum(p));
        // inlining of macro Wln_NtkForEachObj
        for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
            if !Wln_ObjRefs(p, iObj) && !Wln_ObjIsCi(p, iObj) && !Wln_ObjIsCo(p, iObj) && !Wln_ObjIsFf(p, iObj) {
                nSinks++; }
        }
        if nSinks != 0 {
            var nPrinted int32;
            fmt.Printf("These unconnected objects feed into %d sink objects without fanout:\n", nSinks);
            // inlining of macro Wln_NtkForEachObj
            for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
                if !Wln_ObjRefs(p, iObj) && !Wln_ObjIsCi(p, iObj) && !Wln_ObjIsCo(p, iObj) && !Wln_ObjIsFf(p, iObj) {
                    fmt.Fprintf(os.Stdout, "Node %16s (ID %6d) of type %5s (type ID %2d)\n", Wln_ObjName(p, iObj), iObj, Abc_OperName(Wln_ObjType(p, iObj)), Wln_ObjType(p, iObj));
                    if nPrinted++; nPrinted == 5 {
                        break; }
                }
            }
            if nPrinted < nSinks {
                fmt.Printf("...\n"); }
        }
        // inlining of macro Wln_NtkForEachObj
        for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
            if  /*!Wln_ObjRefs(p, iObj) &&*/ !Wln_ObjIsTravIdPrevious(p, iObj) && !Wln_ObjIsCi(p, iObj) {
                // traverse the output logic cone
                if fAcyclic = Wln_NtkIsAcyclic_rec(p, iObj) {
                    continue; }
                // stop as soon as the first loop is detected
                fmt.Fprintf(os.Stdout, "Unconnected object %s\n", Wln_ObjName(p, iObj));
                goto finish;
            }
        }
    }
finish:
    return fAcyclic;
}

/**Function*************************************************************

  Synopsis    [Duplicating network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_NtkTransferNames(pNew *Wln_Ntk_t, p *Wln_Ntk_t) {
    assert(pNew.pManName  == nil && p.pManName != nil, "pNew.pManName  == nil && p.pManName != nil");
    pNew.pManName = p.pManName; 
    p.pManName = nil;
    assert(!Wln_NtkHasCopy(pNew)   && Wln_NtkHasCopy(p), "!Wln_NtkHasCopy(pNew)   && Wln_NtkHasCopy(p)");
    if Wln_NtkHasNameId(p) {
        var i int32;
        assert(!Wln_NtkHasNameId(pNew) && Wln_NtkHasNameId(p), "!Wln_NtkHasNameId(pNew) && Wln_NtkHasNameId(p)");
        Wln_NtkCleanNameId( pNew );
        // inlining of macro Wln_NtkForEachObj
        for  i = 1; i < Wln_NtkObjNum(p); i++  { 
            if Wln_ObjCopy(p, i) && i < Vec_IntSize(&p.vNameIds) && Wln_ObjNameId(p, i) {
                Wln_ObjSetNameId( pNew, Wln_ObjCopy(p, i), Wln_ObjNameId(p, i) ); }
        }
        Vec_IntErase( &p.vNameIds );
    }
    if Wln_NtkHasInstId(p) {
        var i int32;
        assert(!Wln_NtkHasInstId(pNew) && Wln_NtkHasInstId(p), "!Wln_NtkHasInstId(pNew) && Wln_NtkHasInstId(p)");
        Wln_NtkCleanInstId( pNew );
        // inlining of macro Wln_NtkForEachObj
        for  i = 1; i < Wln_NtkObjNum(p); i++  { 
            if Wln_ObjCopy(p, i) && i < Vec_IntSize(&p.vInstIds) && Wln_ObjInstId(p, i) {
                Wln_ObjSetInstId( pNew, Wln_ObjCopy(p, i), Wln_ObjInstId(p, i) ); }
        }
        Vec_IntErase( &p.vInstIds );
    }
}
func Wln_ObjDup(pNew *Wln_Ntk_t, p *Wln_Ntk_t, iObj int32) int32 {
    var i int32;
    var iFanin int32;
    var iObjNew int32 = Wln_ObjClone( pNew, p, iObj );
    // inlining of macro Wln_ObjForEachFanin
    for  i = 0; (i < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, i)), 1); i++  { if !(iFanin != 0) {} else {
        Wln_ObjAddFanin( pNew, iObjNew, Wln_ObjCopy(p, iFanin) ); }
    }
    if Wln_ObjIsConst(p, iObj) {
        Wln_ObjSetConst( pNew, iObjNew, Wln_ObjFanin0(p, iObj) );
    } else if Wln_ObjIsSlice(p, iObj) || Wln_ObjIsRotate(p, iObj) || Wln_ObjIsTable(p, iObj) {
        Wln_ObjSetFanin( pNew, iObjNew, 1, Wln_ObjFanin1(p, iObj) ); }
    Wln_ObjSetCopy( p, iObj, iObjNew );
    return iObjNew;
}
func Wln_NtkDupDfs_rec(pNew *Wln_Ntk_t, p *Wln_Ntk_t, iObj int32) int32 {
    var i, iFanin int32;
    if iObj == 0 {
        return 0; }
    if Wln_ObjCopy(p, iObj) {
        return Wln_ObjCopy(p, iObj); }
    //printf( "Visiting node %16s (ID %6d) of type %5s (type ID %2d)\n", Wln_ObjName(p, iObj), iObj, Abc_OperName(Wln_ObjType(p, iObj)), Wln_ObjType(p, iObj) );
    assert(!Wln_ObjIsFf(p, iObj), "!Wln_ObjIsFf(p, iObj)");
    // inlining of macro Wln_ObjForEachFanin
    for  i = 0; (i < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, i)), 1); i++  { if !(iFanin != 0) {} else {
        Wln_NtkDupDfs_rec( pNew, p, iFanin ); }
    }
    return Wln_ObjDup( pNew, p, iObj );
}
func Wln_NtkDupDfs(p *Wln_Ntk_t) *Wln_Ntk_t {
    var i, k, iObj, iFanin int32;
    var pNew *Wln_Ntk_t  = Wln_NtkAlloc( p.pName, Wln_NtkObjNum(p) );
    pNew.fSmtLib = p.fSmtLib;
    if p.pSpec { pNew.pSpec = Abc_UtilStrsav( p.pSpec ); }
    Wln_NtkCleanCopy( p );
    // inlining of macro Wln_NtkForEachCi
    for  i = 0; (i < Wln_NtkCiNum(p)) && (((iObj) = Wln_NtkCi(p, i)), 1); i++  {
        Wln_ObjDup( pNew, p, iObj );
    }
    // inlining of macro Wln_NtkForEachFf
    for  i = 0; (i < Wln_NtkFfNum(p)) && (((iObj) = Wln_NtkFf(p, i)), 1); i++  {
        Wln_ObjSetCopy( p, iObj, Wln_ObjClone(pNew, p, iObj) );
    }
    // inlining of macro Wln_NtkForEachCo
    for  i = 0; (i < Wln_NtkCoNum(p)) && (((iObj) = Wln_NtkCo(p, i)), 1); i++  {
        Wln_NtkDupDfs_rec( pNew, p, iObj );
    }
    // inlining of macro Wln_NtkForEachFf
    for  i = 0; (i < Wln_NtkFfNum(p)) && (((iObj) = Wln_NtkFf(p, i)), 1); i++  {
        // inlining of macro Wln_ObjForEachFanin
        for  k = 0; (k < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, k)), 1); k++  { if !(iFanin != 0) {} else {
            Wln_ObjAddFanin( pNew, Wln_ObjCopy(p, iObj), Wln_NtkDupDfs_rec(pNew, p, iFanin) ); }
        }
    }
    if Wln_NtkHasNameId(p) {
        Wln_NtkTransferNames( pNew, p ); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Create fanin/fanout map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_NtkCreateRefs(p *Wln_Ntk_t) {
    var k, iObj, iFanin int32;
    Wln_NtkCleanRefs( p );
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        // inlining of macro Wln_ObjForEachFanin
        for  k = 0; (k < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, k)), 1); k++  { if !(iFanin != 0) {} else {
            Wln_ObjRefsInc( p, iFanin ); }
        }
    }
}
func Wln_NtkFaninNum(p *Wln_Ntk_t) int32 {
    var iObj, nEdges int32;
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        nEdges += Wln_ObjFaninNum(p, iObj);
    }
    return nEdges;
}
func Wln_NtkStartFaninMap(p *Wln_Ntk_t, vFaninMap *Vec_Int_t, nMulti int32) {
    var iObj int32;
    var iOffset int32 = Wln_NtkObjNum(p);
    Vec_IntFill( vFaninMap, iOffset + nMulti * Wln_NtkFaninNum(p), 0 );
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        Vec_IntWriteEntry( vFaninMap, iObj, iOffset );
        iOffset += nMulti * Wln_ObjFaninNum(p, iObj);
    }
    assert(iOffset == Vec_IntSize(vFaninMap), "iOffset == Vec_IntSize(vFaninMap)");
}
func Wln_NtkStartFanoutMap(p *Wln_Ntk_t, vFanoutMap *Vec_Int_t, vFanoutNums *Vec_Int_t, nMulti int32) {
    var iObj int32;
    var iOffset int32 = Wln_NtkObjNum(p);
    Vec_IntFill( vFanoutMap, iOffset + nMulti * Vec_IntSum(vFanoutNums), 0 );
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        Vec_IntWriteEntry( vFanoutMap, iObj, iOffset );
        iOffset += nMulti * Wln_ObjRefs(p, iObj);
    }
    assert(iOffset == Vec_IntSize(vFanoutMap), "iOffset == Vec_IntSize(vFanoutMap)");
}

/**Function*************************************************************

  Synopsis    [Static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_NtkStaticFanoutStart(p *Wln_Ntk_t) {
    var k, iObj, iFanin int32;
    var vRefsCopy *Vec_Int_t  = Vec_IntAlloc(0);
    Wln_NtkCreateRefs( p );
    Wln_NtkStartFanoutMap( p, &p.vFanout, &p.vRefs, 1 );
    // inlining of macro ABC_SWAP
    { var t Vec_Int_t  = *vRefsCopy; *vRefsCopy = p.vRefs; p.vRefs = t; };
    // add fanouts
    Wln_NtkCleanRefs( p );
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        // inlining of macro Wln_ObjForEachFanin
        for  k = 0; (k < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, k)), 1); k++  { if !(iFanin != 0) {} else {
            Wln_ObjSetFanout( p, iFanin, Wln_ObjRefsInc(p, iFanin), iObj ); }
        }
    }
    // double-check the current number of fanouts added
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        assert(Wln_ObjRefs(p, iObj) == Vec_IntEntry(vRefsCopy, iObj), "Wln_ObjRefs(p, iObj) == Vec_IntEntry(vRefsCopy, iObj)");
    }
    Vec_IntFree( vRefsCopy );
}
func Wln_NtkStaticFanoutStop(p *Wln_Ntk_t) {
    Vec_IntErase( &p.vRefs );
    Vec_IntErase( &p.vFanout );
}
func Wln_NtkStaticFanoutTest(p *Wln_Ntk_t) {
    var k, iObj, iFanout int32;
    fmt.Printf("Printing fanouts of %d objects of network \"%s\":\n", Wln_NtkObjNum(p), p.pName);
    Wln_NtkStaticFanoutStart( p );
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  {
        Wln_ObjPrint( p, iObj );
        fmt.Printf("   Fanouts : ");
        // inlining of macro Wln_ObjForEachFanoutStatic
        for  k = 0; (k < Wln_ObjRefs(p, iObj)) && (((iFanout) = Wln_ObjFanout(p, iObj, k)), 1); k++  {
            fmt.Printf("%5d ", iFanout);
        }
        fmt.Printf("\n");
    }
    Wln_NtkStaticFanoutStop( p );
    fmt.Printf("\n");
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

