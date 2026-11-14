/**CHeaderFile*****************************************************************

  FileName    [cudd.h]

  PackageName [cudd]

  Synopsis    [The University of Colorado decision diagram package.]

  Description [External functions and data strucures of the CUDD package.
                <ul>
                <li> To turn on the gathering of statistics, define DD_STATS.
                <li> To link with mis, define DD_MIS.
                </ul>
               Modified by Abelardo Pardo to interface it to VIS.
  ]

  SeeAlso     []

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

  Revision    [$Id: cudd.h,v 1.174 2009/02/21 05:55:18 fabio Exp $]

******************************************************************************/

package core

// Numeric layout uses explicit Go widths instead of host C preprocessor branches.
type DdHalfWord = uint32
type CUDD_VALUE_TYPE = float64
type DdApaDigit = uint32
type DdApaDoubleDigit = uint64
type DdApaNumber = []DdApaDigit
type DdChildren struct{ T, E *DdNode }

const (
	CUDD_VERSION                       = "2.4.2"
	CUDD_OUT_OF_MEM                    = -1
	CUDD_UNIQUE_SLOTS                  = 256
	CUDD_CACHE_SLOTS                   = 262144
	DD_MAX_LOOSE_FRACTION              = 5
	DD_MAX_CACHE_FRACTION              = 3
	DD_MAX_CACHE_TO_SLOTS_RATIO        = 4   // src/bdd/cudd/cuddInt.h
	DD_GC_FRAC_HI                      = 4.0 // src/bdd/cudd/cuddInt.h
	DD_GC_FRAC_LO                      = 1.0 // src/bdd/cudd/cuddInt.h
	DD_GC_FRAC_MIN                     = 0.2 // src/bdd/cudd/cuddInt.h
	CUDD_RESIDUE_DEFAULT               = 0
	CUDD_RESIDUE_MSB                   = 1
	CUDD_RESIDUE_TC                    = 2
	CUDD_MAXINDEX               uint32 = 0x7fffffff
	CUDD_CONST_INDEX                   = CUDD_MAXINDEX
	DD_APA_BITS                        = 32
	DD_APA_BASE                 uint64 = 1 << DD_APA_BITS
	DD_APA_MASK                        = DD_APA_BASE - 1
	DD_APA_HEXPRINT                    = "%08x"
)

type Cudd_ReorderingType int32

const (
	CUDD_REORDER_SAME Cudd_ReorderingType = iota

	CUDD_REORDER_NONE

	CUDD_REORDER_RANDOM

	CUDD_REORDER_RANDOM_PIVOT

	CUDD_REORDER_SIFT

	CUDD_REORDER_SIFT_CONVERGE

	CUDD_REORDER_SYMM_SIFT

	CUDD_REORDER_SYMM_SIFT_CONV

	CUDD_REORDER_WINDOW2

	CUDD_REORDER_WINDOW3

	CUDD_REORDER_WINDOW4

	CUDD_REORDER_WINDOW2_CONV

	CUDD_REORDER_WINDOW3_CONV

	CUDD_REORDER_WINDOW4_CONV

	CUDD_REORDER_GROUP_SIFT

	CUDD_REORDER_GROUP_SIFT_CONV

	CUDD_REORDER_ANNEALING

	CUDD_REORDER_GENETIC

	CUDD_REORDER_LINEAR

	CUDD_REORDER_LINEAR_CONVERGE

	CUDD_REORDER_LAZY_SIFT

	CUDD_REORDER_EXACT
)

type Cudd_AggregationType int32

const (
	CUDD_NO_CHECK Cudd_AggregationType = iota

	CUDD_GROUP_CHECK

	CUDD_GROUP_CHECK2

	CUDD_GROUP_CHECK3

	CUDD_GROUP_CHECK4

	CUDD_GROUP_CHECK5

	CUDD_GROUP_CHECK6

	CUDD_GROUP_CHECK7

	CUDD_GROUP_CHECK8

	CUDD_GROUP_CHECK9
)

type Cudd_HookType int32

const (
	CUDD_PRE_GC_HOOK Cudd_HookType = iota

	CUDD_POST_GC_HOOK

	CUDD_PRE_REORDERING_HOOK

	CUDD_POST_REORDERING_HOOK
)

type Cudd_ErrorType int32

const (
	CUDD_NO_ERROR Cudd_ErrorType = iota

	CUDD_MEMORY_OUT

	CUDD_TOO_MANY_NODES

	CUDD_MAX_MEM_EXCEEDED

	CUDD_INVALID_ARG

	CUDD_INTERNAL_ERROR
)

type Cudd_LazyGroupType int32

const (
	CUDD_LAZY_NONE Cudd_LazyGroupType = iota

	CUDD_LAZY_SOFT_GROUP

	CUDD_LAZY_HARD_GROUP

	CUDD_LAZY_UNGROUP
)

type Cudd_VariableType int32

const (
	CUDD_VAR_PRIMARY_INPUT Cudd_VariableType = iota

	CUDD_VAR_PRESENT_STATE

	CUDD_VAR_NEXT_STATE
)

// Hook and operation callbacks use Go values in place of untyped C pointers.
type DD_HFP func(*DdManager, string, any) int
type DD_PRFP func(*DdManager, int, []*DdNode, []*DdNode, []*DdNode) *DdNode
type DD_AOP func(*DdManager, **DdNode, **DdNode) *DdNode
type DD_MAOP func(*DdManager, *DdNode) *DdNode
type DD_CTFP func(*DdManager, *DdNode, *DdNode) *DdNode
type DD_CTFP1 func(*DdManager, *DdNode) *DdNode
type DD_OOMFP func(int64)
type DD_QSFP func(any, any) int

func Cudd_Complement(n *DdNode) *DdNode           { return Cudd_Not(Cudd_Regular(n)) }
func Cudd_V(n *DdNode) float64                    { return Cudd_Regular(n).Value }
func Cudd_ReadIndex(dd *DdManager, index int) int { return Cudd_ReadPerm(dd, index) }
