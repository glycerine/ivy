package mtr

import "os"

/**CFile***********************************************************************

  FileName    [mtrBasic.c]

  PackageName [mtr]

  Synopsis    [Basic manipulation of multiway branching trees.]

  Description [External procedures included in this module:
            <ul>
            <li> Mtr_AllocNode()
            <li> Mtr_DeallocNode()
            <li> Mtr_InitTree()
            <li> Mtr_FreeTree()
            <li> Mtr_CopyTree()
            <li> Mtr_MakeFirstChild()
            <li> Mtr_MakeLastChild()
            <li> Mtr_CreateFirstChild()
            <li> Mtr_CreateLastChild()
            <li> Mtr_MakeNextSibling()
            <li> Mtr_PrintTree()
            </ul>
            ]

  SeeAlso     [cudd package]

  Author      [Fabio Somenzi]

  Copyright   [Copyright (c) 1995-2004, Regents of the University of Colorado

  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:

  Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright
  notice, this list of conditions and the following disclaimer in the
  documentation and/or other materials provided with the distribution.

  Neither the name of the University of Colorado nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
  FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
  BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
  LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
  ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.]

******************************************************************************/

//#include "misc/util/util_hack.h"
//#include "mtrInt.h"

//ABC_NAMESPACE_IMPL_START

/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] MTR_UNUSED = "$Id: mtrBasic.c,v 1.13 2009/02/20 02:03:47 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Allocates new tree node.]

  Description [Allocates new tree node. Returns pointer to node.]

  SideEffects [None]

  SeeAlso     [Mtr_DeallocNode]

******************************************************************************/
func Mtr_AllocNode() *MtrNode {
    var node *MtrNode;

    // inlining of macro ABC_ALLOC
    node = ((MtrNode *) malloc(sizeof(MtrNode) * (size_t)(1)));
    return node;

} /* Mtr_AllocNode */


/**Function********************************************************************

  Synopsis    [Deallocates tree node.]

  Description []

  SideEffects [None]

  SeeAlso     [Mtr_AllocNode]

******************************************************************************/
void
Mtr_DeallocNode(
  MtrNode * node /* node to be deallocated */)
{
    // inlining of macro ABC_FREE
    ((node) ? (free((char *) (node)), (node) = 0) : 0);
    return;

} /* end of Mtr_DeallocNode */


/**Function********************************************************************

  Synopsis    [Initializes tree with one node.]

  Description [Initializes tree with one node. Returns pointer to node.]

  SideEffects [None]

  SeeAlso     [Mtr_FreeTree Mtr_InitGroupTree]

******************************************************************************/
func Mtr_InitTree() *MtrNode {
    var node *MtrNode;

    node = Mtr_AllocNode();
    if node == nil { return(nil); }

    node.parent = node.child = node.elder = node.younger = nil;
    node.flags = 0;

    return(node);

} /* end of Mtr_InitTree */


/**Function********************************************************************

  Synopsis    [Disposes of tree rooted at node.]

  Description []

  SideEffects [None]

  SeeAlso     [Mtr_InitTree]

******************************************************************************/
func Mtr_FreeTree(node *MtrNode) {
    if node == nil { return; }
    // inlining of macro MTR_TEST
    if ! (node.flags & (MTR_TERMINAL)) { Mtr_FreeTree(node.child); }
    Mtr_FreeTree(node.younger);
    Mtr_DeallocNode(node);
    return;

} /* end of Mtr_FreeTree */


/**Function********************************************************************

  Synopsis    [Makes a copy of tree.]

  Description [Makes a copy of tree. If parameter expansion is greater
  than 1, it will expand the tree by that factor. It is an error for
  expansion to be less than 1. Returns a pointer to the copy if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Mtr_InitTree]

******************************************************************************/
func Mtr_CopyTree(node *MtrNode, expansion int32) *MtrNode {
    var copy *MtrNode;

    if node == nil { return(nil); }
    if expansion < 1 { return(nil); }
    copy = Mtr_AllocNode();
    if copy == nil { return(nil); }
    copy.parent = copy.elder = copy.child = copy.younger = nil;
    if node.child != nil {
        copy.child = Mtr_CopyTree(node.child, expansion);
        if copy.child == nil {
            Mtr_DeallocNode(copy);
            return(nil);
        }
    }
    if node.younger != nil {
        copy.younger = Mtr_CopyTree(node.younger, expansion);
        if copy.younger == nil {
            Mtr_FreeTree(copy);
            return(nil);
        }
    }
    copy.flags = node.flags;
    copy.low = node.low * expansion;
    copy.size = node.size * expansion;
    copy.index = node.index * expansion;
    if copy.younger { copy.younger.elder = copy; }
    if copy.child {
        var auxnode *MtrNode  = copy.child;
        for auxnode != nil {
            auxnode.parent = copy;
            auxnode = auxnode.younger;
        }
    }
    return(copy);

} /* end of Mtr_CopyTree */


/**Function********************************************************************

  Synopsis    [Makes child the first child of parent.]

  Description []

  SideEffects [None]

  SeeAlso     [Mtr_MakeLastChild Mtr_CreateFirstChild]

******************************************************************************/
func Mtr_MakeFirstChild(parent *MtrNode, child *MtrNode) {
    child.parent = parent;
    child.younger = parent.child;
    child.elder = nil;
    if parent.child != nil {
#ifdef MTR_DEBUG
        assert(parent->child->elder == NULL);
#endif
        parent.child.elder = child;
    }
    parent.child = child;
    return;

} /* end of Mtr_MakeFirstChild */


/**Function********************************************************************

  Synopsis    [Makes child the last child of parent.]

  Description []

  SideEffects [None]

  SeeAlso     [Mtr_MakeFirstChild Mtr_CreateLastChild]

******************************************************************************/
func Mtr_MakeLastChild(parent *MtrNode, child *MtrNode) {
    var node *MtrNode;

    child.younger = nil;

    if parent.child == nil {
        parent.child = child;
        child.elder = nil;
    } else {
        for node = parent.child;
             node.younger != nil;
             node = node.younger {}
        node.younger = child;
        child.elder = node;
    }
    child.parent = parent;
    return;

} /* end of Mtr_MakeLastChild */


/**Function********************************************************************

  Synopsis    [Creates a new node and makes it the first child of parent.]

  Description [Creates a new node and makes it the first child of
  parent. Returns pointer to new child.]

  SideEffects [None]

  SeeAlso     [Mtr_MakeFirstChild Mtr_CreateLastChild]

******************************************************************************/
func Mtr_CreateFirstChild(parent *MtrNode) *MtrNode {
    var child *MtrNode;

    child = Mtr_AllocNode();
    if child == nil { return(nil); }

    child.child = child.younger = child. elder = nil;
    child.flags = 0;
    Mtr_MakeFirstChild(parent,child);
    return(child);

} /* end of Mtr_CreateFirstChild */


/**Function********************************************************************

  Synopsis    [Creates a new node and makes it the last child of parent.]

  Description [Creates a new node and makes it the last child of parent.
  Returns pointer to new child.]

  SideEffects [None]

  SeeAlso     [Mtr_MakeLastChild Mtr_CreateFirstChild]

******************************************************************************/
func Mtr_CreateLastChild(parent *MtrNode) *MtrNode {
    var child *MtrNode;

    child = Mtr_AllocNode();
    if child == nil { return(nil); }

    child.child = child.younger = child.elder = nil;
    child.flags = 0;
    Mtr_MakeLastChild(parent,child);
    return(child);

} /* end of Mtr_CreateLastChild */


/**Function********************************************************************

  Synopsis    [Makes second the next sibling of first.]

  Description [Makes second the next sibling of first. Second becomes a
  child of the parent of first.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Mtr_MakeNextSibling(first *MtrNode, second *MtrNode) {
    second.younger = first.younger;
    if first.younger != nil {
        first.younger.elder = second;
    }
    second.parent = first.parent;
    first.younger = second;
    second.elder = first;
    return;

} /* end of Mtr_MakeNextSibling */


/**Function********************************************************************

  Synopsis    [Prints a tree, one node per line.]

  Description []

  SideEffects [None]

  SeeAlso     [Mtr_PrintGroups]

******************************************************************************/
func Mtr_PrintTree(node *MtrNode) {
    if node == nil { return; }
    (void) fprintf(os.Stdout,
    #if SIZEOF_VOID_P == 8
    "N=0x%-8lx C=0x%-8lx Y=0x%-8lx E=0x%-8lx P=0x%-8lx F=%x L=%u S=%u\n",
    (unsigned long) node, (unsigned long) node->child,
    (unsigned long) node->younger, (unsigned long) node->elder,
    (unsigned long) node->parent, node->flags, node->low, node->size);
#else
    "N=0x%-8x C=0x%-8x Y=0x%-8x E=0x%-8x P=0x%-8x F=%x L=%hu S=%hu\n",
    (unsigned) node, (unsigned) node->child,
    (unsigned) node->younger, (unsigned) node->elder,
    (unsigned) node->parent, node->flags, node->low, node->size);
#endif
    // inlining of macro MTR_TEST
    if (!(node.flags & (MTR_TERMINAL))) Mtr_PrintTree(node.child);
    Mtr_PrintTree(node.younger);
    return;

} /* end of Mtr_PrintTree */

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/

//ABC_NAMESPACE_IMPL_END
