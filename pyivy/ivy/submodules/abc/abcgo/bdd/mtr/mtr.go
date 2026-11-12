/**CHeaderFile*****************************************************************

  FileName    [mtr.h]

  PackageName [mtr]

  Synopsis    [Multiway-branch tree manipulation]

  Description [This package provides two layers of functions. Functions
  of the lower level manipulate multiway-branch trees, implemented
  according to the classical scheme whereby each node points to its
  first child and its previous and next siblings. These functions are
  collected in mtrBasic.c.<p>
  Functions of the upper layer deal with group trees, that is the trees
  used by group sifting to represent the grouping of variables. These
  functions are collected in mtrGroup.c.]

  SeeAlso     [The CUDD package documentation; specifically on group
  sifting.]

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

  Revision    [$Id: mtr.h,v 1.14 2009/02/20 02:03:47 fabio Exp $]

******************************************************************************/

// Package mtr implements the CUDD multiway group-tree algorithms in Go.
// Original algorithms: src/bdd/mtr (University of Colorado BSD license).
package mtr

type MtrHalfWord = uint32

const (
	MTR_DEFAULT  uint32 = 0
	MTR_TERMINAL uint32 = 1
	MTR_SOFT     uint32 = 2
	MTR_FIXED    uint32 = 4
	MTR_NEWNODE  uint32 = 8
	MTR_MAXHIGH  uint32 = 0x7fffffff
)

// MtrNode is an ordered multiway tree node. Low and Size identify a contiguous
// half-open leaf interval. Sibling links are doubly linked.
type MtrNode struct {
	Flags, Low, Size, Index       uint32
	Parent, Child, Elder, Younger *MtrNode
}

func MTR_SET(n *MtrNode, f uint32)         { n.Flags |= f }
func MTR_RESET(n *MtrNode, f uint32)       { n.Flags &^= f }
func MTR_TEST(n *MtrNode, f uint32) uint32 { return n.Flags & f }
