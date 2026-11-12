package parse

import "fmt"

/**CFile****************************************************************

  FileName    [parseStack.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Stacks used by the formula parser.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 18, 2003.]

  Revision    [$Id: parseStack.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "parseInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type ParseStackFnStruct struct {
    pData      *any;        // the array of elements
    Top           int32;          // the index
    Size           int32;         // the stack size
}

type ParseStackOpStruct struct {
    pData          *int32;        // the array of elements
    Top           int32;          // the index
    Size           int32;         // the stack size
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackFnStart(nDepth int32) *Parse_StackFn_t {
    var p *Parse_StackFn_t;
    // inlining of macro ABC_ALLOC
    p = ((Parse_StackFn_t *) malloc(sizeof(Parse_StackFn_t) * (size_t)(1)));
    memset( p, 0, sizeof(Parse_StackFn_t) );
    // inlining of macro ABC_ALLOC
    p.pData = ((void * *) malloc(sizeof(void *) * (size_t)(nDepth)));
    p.Size = nDepth;
    return p;
}

/**Function*************************************************************

  Synopsis    [Checks whether the stack is empty.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackFnIsEmpty(p *Parse_StackFn_t) int32 {
    return int32((p.Top == 0));
}

/**Function*************************************************************

  Synopsis    [Pushes an entry into the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackFnPush(p *Parse_StackFn_t, bFunc any) {
    if p.Top >= p.Size {
        fmt.Printf("Parse_StackFnPush(): Stack size is too small!\n");
        return;
    }
    p.pData[ p.Top++ ] = bFunc;
}

/**Function*************************************************************

  Synopsis    [Pops an entry out of the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackFnPop(p *Parse_StackFn_t) any {
    if p.Top == 0 {
        fmt.Printf("Parse_StackFnPush(): Trying to extract data from the empty stack!\n");
        return nil;
    }
    return p.pData[ --p.Top ];
}

/**Function*************************************************************

  Synopsis    [Deletes the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackFnFree(p *Parse_StackFn_t) {
    // inlining of macro ABC_FREE
    ((p.pData) ? (free((char *) (p.pData)), (p.pData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}




/**Function*************************************************************

  Synopsis    [Starts the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackOpStart(nDepth int32) *Parse_StackOp_t {
    var p *Parse_StackOp_t;
    // inlining of macro ABC_ALLOC
    p = ((Parse_StackOp_t *) malloc(sizeof(Parse_StackOp_t) * (size_t)(1)));
    memset( p, 0, sizeof(Parse_StackOp_t) );
    // inlining of macro ABC_ALLOC
    p.pData = ((int *) malloc(sizeof(int) * (size_t)(nDepth)));
    p.Size = nDepth;
    return p;
}

/**Function*************************************************************

  Synopsis    [Checks whether the stack is empty.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackOpIsEmpty(p *Parse_StackOp_t) int32 {
    return int32((p.Top == 0));
}

/**Function*************************************************************

  Synopsis    [Pushes an entry into the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackOpPush(p *Parse_StackOp_t, Oper int32) {
    if p.Top >= p.Size {
        fmt.Printf("Parse_StackOpPush(): Stack size is too small!\n");
        return;
    }
    p.pData[ p.Top++ ] = Oper;
}

/**Function*************************************************************

  Synopsis    [Pops an entry out of the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackOpPop(p *Parse_StackOp_t) int32 {
    if p.Top == 0 {
        fmt.Printf("Parse_StackOpPush(): Trying to extract data from the empty stack!\n");
        return -1;
    }
    return p.pData[ --p.Top ];
}

/**Function*************************************************************

  Synopsis    [Deletes the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Parse_StackOpFree(p *Parse_StackOp_t) {
    // inlining of macro ABC_FREE
    ((p.pData) ? (free((char *) (p.pData)), (p.pData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

