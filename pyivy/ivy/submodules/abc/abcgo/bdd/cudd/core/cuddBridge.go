/**CFile***********************************************************************

  FileName    [cuddBridge.c]

  PackageName [cudd]

  Synopsis    [Translation from BDD to ADD and vice versa and transfer between
  different managers.]

  Description [External procedures included in this file:
            <ul>
            <li> Cudd_addBddThreshold()
            <li> Cudd_addBddStrictThreshold()
            <li> Cudd_addBddInterval()
            <li> Cudd_addBddIthBit()
            <li> Cudd_BddToAdd()
            <li> Cudd_addBddPattern()
            <li> Cudd_bddTransfer()
            </ul>
        Internal procedures included in this file:
            <ul>
            <li> cuddBddTransfer()
            <li> cuddAddBddDoPattern()
            </ul>
        Static procedures included in this file:
            <ul>
            <li> addBddDoThreshold()
            <li> addBddDoStrictThreshold()
            <li> addBddDoInterval()
            <li> addBddDoIthBit()
            <li> ddBddToAddRecur()
            <li> cuddBddTransferRecur()
            </ul>
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

******************************************************************************/

package core

import "math"

// Cudd_BddToAdd converts a Boolean diagram to a numeric 0/1 ADD.
func Cudd_BddToAdd(dd *DdManager, f *DdNode) *DdNode {
	if dd == nil || !dd.valid(f) {
		return nil
	}
	cache := make(map[*DdNode]*DdNode)
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if n == dd.One {
			return dd.One
		}
		if n == dd.LogicZero {
			return dd.Zero
		}
		if r := cache[n]; r != nil {
			return r
		}
		p := Cudd_Regular(n)
		t, e := p.T, p.E
		if Cudd_IsComplement(n) {
			t, e = Cudd_Not(t), Cudd_Not(e)
		}
		r := Cudd_addIte(dd, Cudd_addIthVar(dd, int(p.Index)), visit(t), visit(e))
		cache[n] = r
		return r
	}
	return visit(f)
}

// Cudd_addBddPattern maps each nonzero ADD leaf to Boolean true.
func Cudd_addBddPattern(dd *DdManager, f *DdNode) *DdNode {
	return addToBoolean(dd, f, func(value float64) (bool, bool) { return value != 0, true })
}

// Cudd_bddNewVarAtLevel inserts a fresh projection without changing the
// relative order of existing variables or rebuilding any existing diagrams.
func Cudd_bddNewVarAtLevel(dd *DdManager, level int) *DdNode {
	if dd == nil || level < 0 || level > len(dd.Vars) {
		return nil
	}
	v := Cudd_bddNewVar(dd)
	if v == nil {
		return nil
	}
	order := append([]int(nil), dd.InvPerm...)
	copy(order[level+1:], order[level:len(order)-1])
	order[level] = int(v.Index)
	copy(dd.InvPerm, order)
	for level, index := range order {
		dd.Perm[index] = level
	}
	dd.symmetryGroups = nil
	return v
}

// addToBoolean maps numeric leaves while preserving the ADD's shared graph.
// Returning nil from predicate validation makes invalid integer-bit extraction
// fail as a whole instead of silently constructing a partial BDD.
func addToBoolean(dd *DdManager, f *DdNode, predicate func(float64) (bool, bool)) *DdNode {
	if dd == nil || !dd.validAdd(f) {
		return nil
	}
	memo := make(map[*DdNode]*DdNode)
	var visit func(*DdNode) *DdNode
	visit = func(n *DdNode) *DdNode {
		if r := memo[n]; r != nil {
			return r
		}
		if Cudd_IsConstant(n) {
			value, valid := predicate(n.Value)
			if !valid {
				return nil
			}
			if value {
				return dd.One
			}
			return dd.LogicZero
		}
		t := visit(n.T)
		if t == nil {
			return nil
		}
		e := visit(n.E)
		if e == nil {
			return nil
		}
		r := Cudd_bddIte(dd, Cudd_bddIthVar(dd, int(n.Index)), t, e)
		memo[n] = r
		return r
	}
	return visit(f)
}

// Cudd_addBddThreshold selects leaves at least value. Negated comparisons
// deliberately retain CUDD's behavior for IEEE NaN leaves and thresholds.
func Cudd_addBddThreshold(dd *DdManager, f *DdNode, value float64) *DdNode {
	return addToBoolean(dd, f, func(leaf float64) (bool, bool) { return !(leaf < value), true })
}
func Cudd_addBddStrictThreshold(dd *DdManager, f *DdNode, value float64) *DdNode {
	return addToBoolean(dd, f, func(leaf float64) (bool, bool) { return !(leaf <= value), true })
}
func Cudd_addBddInterval(dd *DdManager, f *DdNode, lower, upper float64) *DdNode {
	return addToBoolean(dd, f, func(leaf float64) (bool, bool) { return !(leaf < lower || leaf > upper), true })
}

// Cudd_addBddIthBit extracts a bit of the truncated signed C int (32 bits).
// Inputs outside C's defined float-to-int domain and shifts outside [0,31]
// fail with nil; the original C operations have undefined behavior there.
func Cudd_addBddIthBit(dd *DdManager, f *DdNode, bit int) *DdNode {
	if bit < 0 || bit > 31 {
		return nil
	}
	return addToBoolean(dd, f, func(leaf float64) (bool, bool) { return addIntegerBit(leaf, bit) })
}

// Shared with ADD-valued bit extraction, so boundary/truncation semantics agree.
func addIntegerBit(leaf float64, bit int) (bool, bool) {
	value := math.Trunc(leaf)
	if math.IsNaN(value) || math.IsInf(value, 0) || value < math.MinInt32 || value > math.MaxInt32 {
		return false, false
	}
	return uint32(int32(value))&(uint32(1)<<uint(bit)) != 0, true
}
