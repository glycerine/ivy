package vec

/**CFile****************************************************************

  FileName    [vecAtt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Array of user-specified attiributes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecAtt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// various attributes
 type Vec_AttrType_t int32
const ( 
    VEC_ATTR_NONE Vec_AttrType_t = 0
     // 0
    VEC_ATTR_COPY Vec_AttrType_t = VEC_ATTR_NONE + 1
         // 1
    VEC_ATTR_LOCAL_AIG Vec_AttrType_t = VEC_ATTR_COPY + 1
    // 2
    VEC_ATTR_LOCAL_SOP Vec_AttrType_t = VEC_ATTR_LOCAL_AIG + 1
    // 3
    VEC_ATTR_LOCAL_BDD Vec_AttrType_t = VEC_ATTR_LOCAL_SOP + 1
    // 4
    VEC_ATTR_GLOBAL_AIG Vec_AttrType_t = VEC_ATTR_LOCAL_BDD + 1
   // 5
    VEC_ATTR_GLOBAL_SOP Vec_AttrType_t = VEC_ATTR_GLOBAL_AIG + 1
   // 6
    VEC_ATTR_GLOBAL_BDD Vec_AttrType_t = VEC_ATTR_GLOBAL_SOP + 1
   // 7
    VEC_ATTR_LEVEL Vec_AttrType_t = VEC_ATTR_GLOBAL_BDD + 1
        // 8
    VEC_ATTR_LEVEL_REV Vec_AttrType_t = VEC_ATTR_LEVEL + 1
    // 9
    VEC_ATTR_RETIME_LAG Vec_AttrType_t = VEC_ATTR_LEVEL_REV + 1
   // 10
    VEC_ATTR_FRAIG Vec_AttrType_t = VEC_ATTR_RETIME_LAG + 1
        // 11
    VEC_ATTR_MVVAR Vec_AttrType_t = VEC_ATTR_FRAIG + 1
        // 12
    VEC_ATTR_DATA1 Vec_AttrType_t = VEC_ATTR_MVVAR + 1
        // 13
    VEC_ATTR_DATA2 Vec_AttrType_t = VEC_ATTR_DATA1 + 1
        // 14
    VEC_ATTR_TOTAL_NUM Vec_AttrType_t = VEC_ATTR_DATA2 + 1     // 15

)

 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

    
type Vec_Att_t struct {
    // storage for attributes
    nCap              int32;                 // the size of array allocated
    // Removed pArrayInt as it's not 64-bit safe, it generates compiler
    // warnings, and it's unused.
    pArrayPtr           *any;            // the pointer attribute array
    // attribute specific info
    pMan            any;                 // the manager for this attribute
    void (*pFuncFreeMan) (void *);         // the procedure to free the manager
    void*(*pFuncStartObj)(void *);         // the procedure to start one attribute
    void (*pFuncFreeObj) (void *, void *); // the procedure to free one attribute
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttAlloc(nSize int32, pMan any, pFuncFreeMan func(any), pFuncStartObj func(any) any, pFuncFreeObj func(any, any)) *Vec_Att_t {
    var p *Vec_Att_t;
    p = new(Vec_Att_t);
    *p = Vec_Att_t{};
    p.pMan          = pMan;
    p.pFuncFreeMan  = pFuncFreeMan;
    p.pFuncStartObj = pFuncStartObj;
    p.pFuncFreeObj  = pFuncFreeObj;
    if nSize != 0 {
        p.nCap = nSize;
    } else {
        p.nCap = 16;
    }
    p.pArrayPtr = ABC_ALLOC( void *, p.nCap );
    memset( p.pArrayPtr, 0, sizeof(void *) * (size_t)p.nCap );
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttFree(p *Vec_Att_t, fFreeMan int32) any {
    var pMan any;
    if p == nil {
        return nil; }
    // free the attributes of objects
    if p.pFuncFreeObj {
        var i int32;
        for  i = 0; i < p.nCap; i++  {
            if p.pArrayPtr[i] {
                p.pFuncFreeObj( p.pMan, p.pArrayPtr[i] ); }
        }
    }
    // free the memory manager
    if fFreeMan != 0 {
        pMan = nil;
    } else {
        pMan = p.pMan;
    }
    if p.pMan && (fFreeMan != 0) {  
        p.pFuncFreeMan( p.pMan ); }
    ABC_FREE( p.pArrayPtr );
    ABC_FREE( p );
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Clears the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttClear(p *Vec_Att_t) {
    // free the attributes of objects
    if p.pFuncFreeObj {
        var i int32;
        if p.pFuncFreeObj {
            for  i = 0; i < p.nCap; i++  {
                if p.pArrayPtr[i] {
                    p.pFuncFreeObj( p.pMan, p.pArrayPtr[i] ); }
            } }
    }
    memset( p.pArrayPtr, 0, sizeof(void *) * (size_t)p.nCap );
}

/**Function*************************************************************

  Synopsis    [Deletes one entry from the attribute manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttFreeEntry(p *Vec_Att_t, i int32) {
    if i >= p.nCap {
        return; }
    if p.pMan {
        if p.pArrayPtr[i] && p.pFuncFreeObj {
            p.pFuncFreeObj( p.pMan, (void *)p.pArrayPtr[i] ); }
    }
    p.pArrayPtr[i] = nil;
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttGrow(p *Vec_Att_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArrayPtr = ABC_REALLOC( void *, p.pArrayPtr, nCapMin );
    memset( p.pArrayPtr + p.nCap, 0, sizeof(void *) * (size_t)(nCapMin - p.nCap) );
    p.nCap = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Writes the entry into its place.]

  Description [Only works if the manager is not defined.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttWriteEntry(p *Vec_Att_t, i int32, pEntry any) {
    assert(p.pArrayPtr, "p.pArrayPtr");
    assert(p.pFuncStartObj == nil, "p.pFuncStartObj == nil");
    if i >= p.nCap {
        Vec_AttGrow( p, (2 * p.nCap > i)? 2 * p.nCap : i + 10 ); }
    p.pArrayPtr[i] = pEntry;
}

/**Function*************************************************************

  Synopsis    [Returns the entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttEntry(p *Vec_Att_t, i int32) any {
    assert(p.pArrayPtr, "p.pArrayPtr");
    if i >= p.nCap {
        Vec_AttGrow( p, (2 * p.nCap > i)? 2 * p.nCap : i + 10 ); }
    if p.pArrayPtr[i] == nil && p.pFuncStartObj {
        p.pArrayPtr[i] = p.pFuncStartObj( p.pMan ); }
    return p.pArrayPtr[i];
}

/**Function*************************************************************

  Synopsis    [Returns the entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttMan(p *Vec_Att_t) any {
    return p.pMan;
}

/**Function*************************************************************

  Synopsis    [Returns the array of attributes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_AttArray(p *Vec_Att_t) *any {
    return p.pArrayPtr;
}



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

