/**CFile***********************************************************************

  FileName    [cuddZddCount.c]

  PackageName [cudd]

  Synopsis    [Procedures to count the number of minterms of a ZDD.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddCount();
                    <li> Cudd_zddCountDouble();
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> cuddZddCountStep();
                    <li> cuddZddCountDoubleStep();
                    <li> st__zdd_count_dbl_free()
                    <li> st__zdd_countfree()
                    </ul>
              ]

  SeeAlso     []

  Author      [Hyong-Kyoon Shin, In-Ho Moon]

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

import (
	"fmt"
	"math/big"
)

// CountZDD returns the exact number of member sets, without enumerating them.
func (dd *DdManager) CountZDD(f *DdNode) (*big.Int, error) {
	if !dd.validZdd(f) {
		return nil, fmt.Errorf("cudd: invalid ZDD edge")
	}
	memo := map[*DdNode]*big.Int{}
	var visit func(*DdNode) *big.Int
	visit = func(f *DdNode) *big.Int {
		if f == dd.Zero {
			return new(big.Int)
		}
		if f == dd.One {
			return big.NewInt(1)
		}
		if n := memo[f]; n != nil {
			return n
		}
		n := new(big.Int).Add(visit(f.T), visit(f.E))
		memo[f] = n
		return n
	}
	return new(big.Int).Set(visit(f)), nil
}

// Count uses Go's native integer range and returns CUDD_OUT_OF_MEM on overflow
// instead of wrapping into a plausible incorrect count. CountZDD is unbounded.
func Cudd_zddCount(dd *DdManager, f *DdNode) int {
	n, err := dd.CountZDD(f)
	if err != nil || !n.IsInt64() {
		return CUDD_OUT_OF_MEM
	}
	v := n.Int64()
	if uint64(v) > uint64(^uint(0)>>1) {
		return CUDD_OUT_OF_MEM
	}
	return int(v)
}
func Cudd_zddCountDouble(dd *DdManager, f *DdNode) float64 {
	n, err := dd.CountZDD(f)
	if err != nil {
		return CUDD_OUT_OF_MEM
	}
	v, _ := new(big.Float).SetInt(n).Float64()
	return v
}
