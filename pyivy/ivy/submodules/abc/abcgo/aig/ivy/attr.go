package ivy

/**CFile****************************************************************

  FileName    [attr.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network attributes.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: attr.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/extra/extra.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

       
type Attr_Man_t struct {
    // attribute info
    nAttrSize                        int32;     // the size of each attribute in bytes
    pManMem           *Extra_MmFixed_t;       // memory manager for attributes
    nAttrs                        int32;        // the number of attributes allocated
    pAttrs                     *any;        // the array of attributes
    fUseInt                        int32;       // uses integer attributes
    // attribute specific info
    pManAttr                      any;      // the manager for this attribute
    void (*pFuncFreeMan) (void *);            // the procedure to call to free attribute-specific manager
    void (*pFuncFreeObj) (void *, void *);    // the procedure to call to free attribute-specific data
}

// at any time, an attribute of the given ID can be
// - not available (p->nAttrs < Id)
// - available but not allocated (p->nAttrs >= Id && p->pAttrs[Id] == NULL)
// - available and allocated (p->nAttrs >= Id && p->pAttrs[Id] != NULL)

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the attribute manager.]

  Description [The manager is simple if it does not need memory manager.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManAlloc(nAttrSize int32, fManMem int32) *Attr_Man_t {
    var p *Attr_Man_t;
    p = ALLOC( Attr_Man_t, 1 );
    *p = Attr_Man_t{};
    p.nAttrSize = nAttrSize;
    if fManMem != 0 {
        p.pManMem = Extra_MmFixedStart( nAttrSize ); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Start the attribute manager for integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManStartInt(nAttrs int32) *Attr_Man_t {
    var p *Attr_Man_t;
    p = Attr_ManAlloc( sizeof(int), 0 );
    p.nAttrs  = nAttrs;
    p.pAttrs  = (void **)ALLOC( int, nAttrs );
    memset( (int *)p.pAttrs, 0, sizeof(int) * nAttrs );
    p.fUseInt = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Start the attribute manager for pointers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManStartPtr(nAttrs int32) *Attr_Man_t {
    var p *Attr_Man_t;
    p = Attr_ManAlloc( sizeof(void *), 0 );
    p.nAttrs  = nAttrs;
    p.pAttrs  = ALLOC( void *, nAttrs );
    memset( p.pAttrs, 0, sizeof(void *) * nAttrs );
    return p;
}

/**Function*************************************************************

  Synopsis    [Start the attribute manager for the fixed entry size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManStartPtrMem(nAttrs int32, nAttrSize int32) *Attr_Man_t {
    var p *Attr_Man_t;
    var i int32;
    p = Attr_ManAlloc( nAttrSize, 1 );
    p.nAttrs  = nAttrs;
    p.pAttrs  = ALLOC( void *, nAttrs );
    for  i = 0; i < p.nAttrs; i++  {
        p.pAttrs[i] = Extra_MmFixedEntryFetch( p.pManMem );
        memset( p.pAttrs[i], 0, nAttrSize );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Stop the attribute manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManStop(p *Attr_Man_t) {
    // free the attributes of objects
    if p.pFuncFreeObj {
        var i int32;
        if p.fUseInt {
            for  i = 0; i < p.nAttrs; i++  {
                if ((int *)p.pAttrs)[i] {
                    p.pFuncFreeObj( p.pManAttr, (void *)((int *)p.pAttrs)[i] ); }
            }
        
        } else {
            for  i = 0; i < p.nAttrs; i++  {
                if p.pAttrs[i] {
                    p.pFuncFreeObj( p.pManAttr, p.pAttrs[i] ); }
            }
        }
    }
    // free the attribute manager
    if p.pManAttr && p.pFuncFreeMan {
        p.pFuncFreeMan( p.pManAttr ); }
    // free the memory manager
    if p.pManMem {  
        Extra_MmFixedStop( p.pManMem); }
    // free the attribute manager
    FREE( p.pAttrs );
    free( p );
}

/**Function*************************************************************

  Synopsis    [Reads the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManReadAttrInt(p *Attr_Man_t, Id int32) int32 {
    assert(p.fUseInt, "p.fUseInt");
    if Id >= p.nAttrs {
        return 0; }
    return ((int *)p.pAttrs)[Id]; 
}

/**Function*************************************************************

  Synopsis    [Reads the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManReadAttrPtr(p *Attr_Man_t, Id int32) any {
    assert(!p.fUseInt, "!p.fUseInt");
    if Id >= p.nAttrs {
        return nil; }
    return p.pAttrs[Id]; 
}

/**Function*************************************************************

  Synopsis    [Writes the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManWriteAttrInt(p *Attr_Man_t, Id int32, Attr int32) {
    assert(p.fUseInt, "p.fUseInt");
    ((int *)p.pAttrs)[Id] = Attr; 
}

/**Function*************************************************************

  Synopsis    [Writes the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManWriteAttrPtr(p *Attr_Man_t, Id int32, pAttr any) {
    assert(!p.fUseInt, "!p.fUseInt");
    assert(p.pManMem == nil, "p.pManMem == nil");
    p.pAttrs[Id] = pAttr; 
}

/**Function*************************************************************

  Synopsis    [Returns or creates the pointer to the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManFetchSpotInt(p *Attr_Man_t, Id int32) *int32 {
    assert(p.fUseInt, "p.fUseInt");
    if Id >= p.nAttrs {
        // save the old size
        var i int32;
        var nAttrsOld int32 = p.nAttrs;
        // get the new size
        if p.nAttrs {
            p.nAttrs = 2*p.nAttrs;
        } else {
            p.nAttrs = 1024;
        }
        p.pAttrs = realloc( p.pAttrs, sizeof(int) * p.nAttrs ); 
        // fill in the empty spots
        for  i = nAttrsOld; i < p.nAttrs; i++  {
            ((int *)p.pAttrs)[Id] = 0;
        }
    }
    return ((int *)p.pAttrs) + Id;
}

/**Function*************************************************************

  Synopsis    [Returns or creates the pointer to the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManFetchSpotPtr(p *Attr_Man_t, Id int32) *any {
    assert(!p.fUseInt, "!p.fUseInt");
    if Id >= p.nAttrs {
        // save the old size
        var i int32;
        var nAttrsOld int32 = p.nAttrs;
        // get the new size
        if p.nAttrs {
            p.nAttrs = 2*p.nAttrs;
        } else {
            p.nAttrs = 1024;
        }
        p.pAttrs = realloc( p.pAttrs, sizeof(void *) * p.nAttrs ); 
        // fill in the empty spots
        for  i = nAttrsOld; i < p.nAttrs; i++  {
            p.pAttrs[Id] = nil;
        }
    }
    // if memory manager is available but entry is not created, create it
    if p.pManMem && p.pAttrs[Id] != nil {
        p.pAttrs[Id] = Extra_MmFixedEntryFetch( p.pManMem );
        memset( p.pAttrs[Id], 0, p.nAttrSize );
    }
    return p.pAttrs + Id;
}


/**Function*************************************************************

  Synopsis    [Returns or creates the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManFetchAttrInt(p *Attr_Man_t, Id int32) int32 {
    return *Attr_ManFetchSpotInt( p, Id );
}

/**Function*************************************************************

  Synopsis    [Returns or creates the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManFetchAttrPtr(p *Attr_Man_t, Id int32) any {
    return *Attr_ManFetchSpotPtr( p, Id );
}

/**Function*************************************************************

  Synopsis    [Sets the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManSetAttrInt(p *Attr_Man_t, Id int32, Attr int32) {
    *Attr_ManFetchSpotInt( p, Id ) = Attr;
}

/**Function*************************************************************

  Synopsis    [Sets the attribute of the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Attr_ManSetAttrPtr(p *Attr_Man_t, Id int32, pAttr any) {
    assert(p.pManMem == nil, "p.pManMem == nil");
    *Attr_ManFetchSpotPtr( p, Id ) = pAttr;
}





//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

