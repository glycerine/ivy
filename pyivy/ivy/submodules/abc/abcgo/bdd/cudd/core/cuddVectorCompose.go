/**CFile***********************************************************************

  FileName    [cuddCompose.c]

  PackageName [cudd]

  Synopsis    [Functional composition and variable permutation of DDs.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddCompose()
                <li> Cudd_addCompose()
                <li> Cudd_addPermute()
                <li> Cudd_addSwapVariables()
                <li> Cudd_bddPermute()
                <li> Cudd_bddVarMap()
                <li> Cudd_SetVarMap()
                <li> Cudd_bddSwapVariables()
                <li> Cudd_bddAdjPermuteX()
                <li> Cudd_addVectorCompose()
                <li> Cudd_addGeneralVectorCompose()
                <li> Cudd_addNonSimCompose()
                <li> Cudd_bddVectorCompose()
                </ul>
               Internal procedures included in this module:
                <ul>
                <li> cuddBddComposeRecur()
                <li> cuddAddComposeRecur()
                </ul>
               Static procedures included in this module:
                <ul>
                <li> cuddAddPermuteRecur()
                <li> cuddBddPermuteRecur()
                <li> cuddBddVarMapRecur()
                <li> cuddAddVectorComposeRecur()
                <li> cuddAddGeneralVectorComposeRecur()
                <li> cuddAddNonSimComposeRecur()
                <li> cuddBddVectorComposeRecur()
                <li> ddIsIthAddVar()
                <li> ddIsIthAddVarPair()
               </ul>
  The permutation functions use a local cache because the results to
  be remembered depend on the permutation being applied.  Since the
  permutation is just an array, it cannot be stored in the global
  cache. There are different procedured for BDDs and ADDs. This is
  because bddPermuteRecur uses cuddBddIteRecur. If this were changed,
  the procedures could be merged.]

  Author      [Fabio Somenzi and Kavita Ravi]

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

// Cudd_bddVectorCompose simultaneously replaces each BDD variable with the
// corresponding Boolean function. Inserted functions are not substituted again.
func Cudd_bddVectorCompose(dd *DdManager, f *DdNode, vector []*DdNode) *DdNode {
	if !dd.valid(f) || len(vector) < len(dd.Vars) {
		return nil
	}
	for _, g := range vector[:len(dd.Vars)] {
		if !dd.valid(g) {
			return nil
		}
	}
	memo := make(map[*DdNode]*DdNode)
	var visit func(*DdNode) *DdNode
	visit = func(f *DdNode) *DdNode {
		if Cudd_IsConstant(f) {
			return f
		}
		if r := memo[f]; r != nil {
			return r
		}
		regular := Cudd_Regular(f)
		t, e := regular.T, regular.E
		if Cudd_IsComplement(f) {
			t, e = Cudd_Not(t), Cudd_Not(e)
		}
		r := Cudd_bddIte(dd, vector[regular.Index], visit(t), visit(e))
		memo[f] = r
		return r
	}
	return visit(f)
}
