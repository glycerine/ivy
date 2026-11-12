package nm

import "fmt"

/**CFile****************************************************************

  FileName    [nmApi.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Name manager.]

  Synopsis    [APIs of the name manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nmApi.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "nmInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the name manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManCreate(nSize int32) *Nm_Man_t {
    var p *Nm_Man_t;
    // allocate the table
    // inlining of macro ABC_ALLOC
    p = ((Nm_Man_t *) malloc(sizeof(Nm_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Nm_Man_t) );
    // set the parameters
    p.nSizeFactor   = 2; // determined the limit on the grow of data before the table resizes
    p.nGrowthFactor = 3; // determined how much the table grows after resizing
    // allocate and clean the bins
    p.nBins = Abc_PrimeCudd(nSize);
    // inlining of macro ABC_ALLOC
    p.pBinsI2N = ((Nm_Entry_t * *) malloc(sizeof(Nm_Entry_t *) * (size_t)(p.nBins)));
    // inlining of macro ABC_ALLOC
    p.pBinsN2I = ((Nm_Entry_t * *) malloc(sizeof(Nm_Entry_t *) * (size_t)(p.nBins)));
    memset( p.pBinsI2N, 0, sizeof(Nm_Entry_t *) * p.nBins );
    memset( p.pBinsN2I, 0, sizeof(Nm_Entry_t *) * p.nBins );
    // start the memory manager
    p.pMem = Extra_MmFlexStart();
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates the name manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManFree(p *Nm_Man_t) {
    Extra_MmFlexStop( p.pMem );
    // inlining of macro ABC_FREE
    ((p.pBinsI2N) ? (free((char *) (p.pBinsI2N)), (p.pBinsI2N) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pBinsN2I) ? (free((char *) (p.pBinsN2I)), (p.pBinsN2I) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Returns the number of objects with names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManNumEntries(p *Nm_Man_t) int32 {
    return p.nEntries;
}

/**Function*************************************************************

  Synopsis    [Creates a new entry in the name manager.]

  Description [Returns 1 if the entry with the given object ID
  already exists in the name manager.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManStoreIdName(p *Nm_Man_t, ObjId int32, Type int32, pName string, pSuffix string) string {
    var pEntry *Nm_Entry_t;
    var RetValue, nEntrySize int32;
    // check if the object with this ID is already stored
    if pEntry = Nm_ManTableLookupId(p, ObjId) {
        fmt.Printf("Nm_ManStoreIdName(): Entry with the same ID already exists.\n");
        return nil;
    }
    // create a new entry
    nEntrySize = sizeof(Nm_Entry_t) + len(pName) + (pSuffix?len(pSuffix):0) + 1;
//    nEntrySize = (nEntrySize / 4 + ((nEntrySize % 4) > 0)) * 4;
    nEntrySize = (nEntrySize / sizeof(char*) + ((nEntrySize % sizeof(char*)) > 0)) * sizeof(char*); // added by Saurabh on Sep 3, 2009
    pEntry = (Nm_Entry_t *)Extra_MmFlexEntryFetch( p.pMem, nEntrySize );
    pEntry.pNextI2N = pEntry.pNextN2I = pEntry.pNameSake = nil;
    pEntry.ObjId = ObjId;
    pEntry.Type = Type;
    sprintf( pEntry.Name, "%s%s", pName, pSuffix? pSuffix : "" );
    // add the entry to the hash table
    RetValue = Nm_ManTableAdd( p, pEntry );
    assert(RetValue == 1, "RetValue == 1");
    return pEntry.Name;
}

/**Function*************************************************************

  Synopsis    [Creates a new entry in the name manager.]

  Description [Returns 1 if the entry with the given object ID
  already exists in the name manager.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManDeleteIdName(p *Nm_Man_t, ObjId int32) {
    var pEntry *Nm_Entry_t;
    pEntry = Nm_ManTableLookupId(p, ObjId);
    if pEntry == nil {
        fmt.Printf("Nm_ManDeleteIdName(): This entry is not in the table.\n");
        return;
    }
    // remove entry from the table
    Nm_ManTableDelete( p, ObjId );
}


/**Function*************************************************************

  Synopsis    [Finds a unique name for the node.]

  Description [If the name exists, tries appending numbers to it until 
  it becomes unique. The name is not added to the table.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Nm_ManCreateUniqueName_NameStr string // was [1000]byte in C

func Nm_ManCreateUniqueName(p *Nm_Man_t, ObjId int32) string {
    
    var pEntry *Nm_Entry_t;
    var i int32;
    if pEntry = Nm_ManTableLookupId(p, ObjId) {
        return pEntry.Name; }
    spStatic_Nm_ManCreateUniqueName_NameStr = fmt.Sprintf("n%d", ObjId);
    for  i = 1; Nm_ManTableLookupName(p, spStatic_Nm_ManCreateUniqueName_NameStr, -1); i++  {
        spStatic_Nm_ManCreateUniqueName_NameStr = fmt.Sprintf("n%d_%d", ObjId, i);
    }
    return spStatic_Nm_ManCreateUniqueName_NameStr;
}

/**Function*************************************************************

  Synopsis    [Returns name of the object if the ID is known.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManFindNameById(p *Nm_Man_t, ObjId int32) string {
    var pEntry *Nm_Entry_t;
    if pEntry = Nm_ManTableLookupId(p, ObjId) {
        return pEntry.Name; }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Returns ID of the object if its name is known.]

  Description [This procedure may return two IDs because POs and latches 
  may have the same name (the only allowed case of name duplication).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManFindIdByName(p *Nm_Man_t, pName string, Type int32) int32 {
    var pEntry *Nm_Entry_t;
    if pEntry = Nm_ManTableLookupName(p, pName, Type) {
        return pEntry.ObjId; }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Returns ID of the object if its name is known.]

  Description [This procedure may return two IDs because POs and latches 
  may have the same name (the only allowed case of name duplication).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManFindIdByNameTwoTypes(p *Nm_Man_t, pName string, Type1 int32, Type2 int32) int32 {
    var iNodeId int32;
    iNodeId = Nm_ManFindIdByName( p, pName, Type1 );
    if iNodeId == -1 {
        iNodeId = Nm_ManFindIdByName( p, pName, Type2 ); }
    if iNodeId == -1 {
        return -1; }
    return iNodeId;
}

/**Function*************************************************************

  Synopsis    [Return the IDs of objects with names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nm_ManReturnNameIds(p *Nm_Man_t) *Vec_Int_t {
    var vNameIds *Vec_Int_t;
    var i int32;
    vNameIds = Vec_IntAlloc( p.nEntries );
    for  i = 0; i < p.nBins; i++  {
        if p.pBinsI2N[i] {
            Vec_IntPush( vNameIds, p.pBinsI2N[i].ObjId ); }
    }
    return vNameIds;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

