/**CFile***********************************************************************

  FileName    [cuddZddPort.c]

  PackageName [cudd]

  Synopsis    [Functions that translate BDDs to ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddPortFromBdd()
                    <li> Cudd_zddPortToBdd()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> zddPortFromBddStep()
                    <li> zddPortToBddStep()
                    </ul>
              ]

  SeeAlso     []

  Author      [Hyong-kyoon Shin, In-Ho Moon]

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
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START

package core

import (
	"math"
	"reflect"
)

// Cudd_zddPortFromBdd follows the C conversion and requires matching BDD and
// ZDD variable counts and orders.
func Cudd_zddPortFromBdd(dd *DdManager, b *DdNode) *DdNode {
	if !dd.valid(b) {
		return nil
	}
	op := reflect.ValueOf(Cudd_zddPortFromBdd).Pointer()
	var step func(*DdNode, int) *DdNode
	step = func(b *DdNode, expected int) *DdNode {
		if b == dd.LogicZero {
			return dd.Zero
		}
		if b == dd.One {
			if expected >= dd.ZddSize {
				return dd.One
			}
			return dd.Univ[expected]
		}
		n := Cudd_Regular(b)
		level := dd.Perm[n.Index]
		result := Cudd_CacheLookup(dd, op, b, nil, nil, 1)
		if result == nil {
			t := step(Cudd_NotCond(n.T, Cudd_IsComplement(b)), level+1)
			if t == nil {
				return nil
			}
			Cudd_Ref(t)
			e := step(Cudd_NotCond(n.E, Cudd_IsComplement(b)), level+1)
			if e == nil {
				Cudd_RecursiveDeref(dd, t)
				return nil
			}
			Cudd_Ref(e)
			result = dd.zddNode(n.Index, t, e)
			Cudd_Ref(result)
			Cudd_RecursiveDeref(dd, t)
			Cudd_RecursiveDeref(dd, e)
			Cudd_CacheInsert(dd, op, b, nil, nil, result, 1)
		} else {
			Cudd_Ref(result)
		}
		for l := level - 1; l >= expected; l-- {
			prev := result
			result = dd.zddNode(uint32(dd.InvPerm[l]), prev, prev)
			Cudd_Ref(result)
			Cudd_RecursiveDeref(dd, prev)
		}
		Cudd_Deref(result)
		return result
	}
	return step(b, 0)
}

// Cudd_zddPortToBdd converts a ZDD using the source depth and cache rules.
func Cudd_zddPortToBdd(dd *DdManager, f *DdNode) *DdNode {
	if !dd.validZdd(f) {
		return nil
	}
	op := reflect.ValueOf(Cudd_zddPortToBdd).Pointer()
	var step func(*DdNode, int) *DdNode
	step = func(f *DdNode, depth int) *DdNode {
		if f == dd.Zero {
			return dd.LogicZero
		}
		if depth == dd.ZddSize {
			return dd.One
		}
		index := dd.InvPermZ[depth]
		level := math.MaxInt
		if !Cudd_IsConstant(f) {
			level = dd.PermZ[f.Index]
		}
		v := Cudd_bddIthVar(dd, index)
		if v == nil {
			return nil
		}
		Cudd_Ref(v)
		if level > depth {
			e := step(f, depth+1)
			if e == nil {
				Cudd_RecursiveDeref(dd, v)
				return nil
			}
			Cudd_Ref(e)
			res := Cudd_bddIte(dd, v, dd.LogicZero, e)
			if res == nil {
				Cudd_RecursiveDeref(dd, v)
				Cudd_RecursiveDeref(dd, e)
				return nil
			}
			Cudd_Ref(res)
			Cudd_RecursiveDeref(dd, v)
			Cudd_RecursiveDeref(dd, e)
			Cudd_Deref(res)
			return res
		}
		res := Cudd_CacheLookup(dd, op, f, nil, nil, 0)
		if res != nil {
			Cudd_RecursiveDeref(dd, v)
			return res
		}
		t := step(f.T, depth+1)
		if t == nil {
			Cudd_RecursiveDeref(dd, v)
			return nil
		}
		Cudd_Ref(t)
		e := step(f.E, depth+1)
		if e == nil {
			Cudd_RecursiveDeref(dd, v)
			Cudd_RecursiveDeref(dd, t)
			return nil
		}
		Cudd_Ref(e)
		res = Cudd_bddIte(dd, v, t, e)
		if res == nil {
			Cudd_RecursiveDeref(dd, v)
			Cudd_RecursiveDeref(dd, t)
			Cudd_RecursiveDeref(dd, e)
			return nil
		}
		Cudd_Ref(res)
		Cudd_RecursiveDeref(dd, v)
		Cudd_RecursiveDeref(dd, t)
		Cudd_RecursiveDeref(dd, e)
		Cudd_Deref(res)
		Cudd_CacheInsert(dd, op, f, nil, nil, res, 0)
		return res
	}
	return step(f, 0)
}
