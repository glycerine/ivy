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

package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// Public graph types are aliases of the sole implementation in core.
type DdHalfWord = core.DdHalfWord
type CUDD_VALUE_TYPE = core.CUDD_VALUE_TYPE
type DdApaDigit = core.DdApaDigit
type DdApaDoubleDigit = core.DdApaDoubleDigit
type DdApaNumber = core.DdApaNumber
type DdChildren = core.DdChildren
type Cudd_ReorderingType = core.Cudd_ReorderingType
type Cudd_AggregationType = core.Cudd_AggregationType
type Cudd_HookType = core.Cudd_HookType
type Cudd_ErrorType = core.Cudd_ErrorType
type Cudd_LazyGroupType = core.Cudd_LazyGroupType
type Cudd_VariableType = core.Cudd_VariableType
type DD_HFP = core.DD_HFP
type DD_PRFP = core.DD_PRFP
type DD_AOP = core.DD_AOP
type DD_MAOP = core.DD_MAOP
type DD_CTFP = core.DD_CTFP
type DD_CTFP1 = core.DD_CTFP1
type DD_OOMFP = core.DD_OOMFP
type DD_QSFP = core.DD_QSFP

const (
	CUDD_VERSION                 = core.CUDD_VERSION
	CUDD_OUT_OF_MEM              = core.CUDD_OUT_OF_MEM
	CUDD_UNIQUE_SLOTS            = core.CUDD_UNIQUE_SLOTS
	CUDD_CACHE_SLOTS             = core.CUDD_CACHE_SLOTS
	CUDD_RESIDUE_DEFAULT         = core.CUDD_RESIDUE_DEFAULT
	CUDD_RESIDUE_MSB             = core.CUDD_RESIDUE_MSB
	CUDD_RESIDUE_TC              = core.CUDD_RESIDUE_TC
	CUDD_MAXINDEX                = core.CUDD_MAXINDEX
	CUDD_CONST_INDEX             = core.CUDD_CONST_INDEX
	DD_APA_BITS                  = core.DD_APA_BITS
	DD_APA_BASE                  = core.DD_APA_BASE
	DD_APA_MASK                  = core.DD_APA_MASK
	DD_APA_HEXPRINT              = core.DD_APA_HEXPRINT
	CUDD_REORDER_SAME            = core.CUDD_REORDER_SAME
	CUDD_REORDER_NONE            = core.CUDD_REORDER_NONE
	CUDD_REORDER_RANDOM          = core.CUDD_REORDER_RANDOM
	CUDD_REORDER_RANDOM_PIVOT    = core.CUDD_REORDER_RANDOM_PIVOT
	CUDD_REORDER_SIFT            = core.CUDD_REORDER_SIFT
	CUDD_REORDER_SIFT_CONVERGE   = core.CUDD_REORDER_SIFT_CONVERGE
	CUDD_REORDER_SYMM_SIFT       = core.CUDD_REORDER_SYMM_SIFT
	CUDD_REORDER_SYMM_SIFT_CONV  = core.CUDD_REORDER_SYMM_SIFT_CONV
	CUDD_REORDER_WINDOW2         = core.CUDD_REORDER_WINDOW2
	CUDD_REORDER_WINDOW3         = core.CUDD_REORDER_WINDOW3
	CUDD_REORDER_WINDOW4         = core.CUDD_REORDER_WINDOW4
	CUDD_REORDER_WINDOW2_CONV    = core.CUDD_REORDER_WINDOW2_CONV
	CUDD_REORDER_WINDOW3_CONV    = core.CUDD_REORDER_WINDOW3_CONV
	CUDD_REORDER_WINDOW4_CONV    = core.CUDD_REORDER_WINDOW4_CONV
	CUDD_REORDER_GROUP_SIFT      = core.CUDD_REORDER_GROUP_SIFT
	CUDD_REORDER_GROUP_SIFT_CONV = core.CUDD_REORDER_GROUP_SIFT_CONV
	CUDD_REORDER_ANNEALING       = core.CUDD_REORDER_ANNEALING
	CUDD_REORDER_GENETIC         = core.CUDD_REORDER_GENETIC
	CUDD_REORDER_LINEAR          = core.CUDD_REORDER_LINEAR
	CUDD_REORDER_LINEAR_CONVERGE = core.CUDD_REORDER_LINEAR_CONVERGE
	CUDD_REORDER_LAZY_SIFT       = core.CUDD_REORDER_LAZY_SIFT
	CUDD_REORDER_EXACT           = core.CUDD_REORDER_EXACT
	CUDD_NO_CHECK                = core.CUDD_NO_CHECK
	CUDD_GROUP_CHECK             = core.CUDD_GROUP_CHECK
	CUDD_GROUP_CHECK2            = core.CUDD_GROUP_CHECK2
	CUDD_GROUP_CHECK3            = core.CUDD_GROUP_CHECK3
	CUDD_GROUP_CHECK4            = core.CUDD_GROUP_CHECK4
	CUDD_GROUP_CHECK5            = core.CUDD_GROUP_CHECK5
	CUDD_GROUP_CHECK6            = core.CUDD_GROUP_CHECK6
	CUDD_GROUP_CHECK7            = core.CUDD_GROUP_CHECK7
	CUDD_GROUP_CHECK8            = core.CUDD_GROUP_CHECK8
	CUDD_GROUP_CHECK9            = core.CUDD_GROUP_CHECK9
	CUDD_PRE_GC_HOOK             = core.CUDD_PRE_GC_HOOK
	CUDD_POST_GC_HOOK            = core.CUDD_POST_GC_HOOK
	CUDD_PRE_REORDERING_HOOK     = core.CUDD_PRE_REORDERING_HOOK
	CUDD_POST_REORDERING_HOOK    = core.CUDD_POST_REORDERING_HOOK
	CUDD_NO_ERROR                = core.CUDD_NO_ERROR
	CUDD_MEMORY_OUT              = core.CUDD_MEMORY_OUT
	CUDD_TOO_MANY_NODES          = core.CUDD_TOO_MANY_NODES
	CUDD_MAX_MEM_EXCEEDED        = core.CUDD_MAX_MEM_EXCEEDED
	CUDD_INVALID_ARG             = core.CUDD_INVALID_ARG
	CUDD_INTERNAL_ERROR          = core.CUDD_INTERNAL_ERROR
	CUDD_LAZY_NONE               = core.CUDD_LAZY_NONE
	CUDD_LAZY_SOFT_GROUP         = core.CUDD_LAZY_SOFT_GROUP
	CUDD_LAZY_HARD_GROUP         = core.CUDD_LAZY_HARD_GROUP
	CUDD_LAZY_UNGROUP            = core.CUDD_LAZY_UNGROUP
	CUDD_VAR_PRIMARY_INPUT       = core.CUDD_VAR_PRIMARY_INPUT
	CUDD_VAR_PRESENT_STATE       = core.CUDD_VAR_PRESENT_STATE
	CUDD_VAR_NEXT_STATE          = core.CUDD_VAR_NEXT_STATE
)

var (
	Cudd_Complement = core.Cudd_Complement
	Cudd_V          = core.Cudd_V
	Cudd_ReadIndex  = core.Cudd_ReadIndex
)
