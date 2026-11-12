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

package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

func Cudd_addBddThreshold(dd *DdManager, f *DdNode, value CUDD_VALUE_TYPE) *DdNode {
	return core.Cudd_addBddThreshold(dd, f, value)
}
func Cudd_addBddStrictThreshold(dd *DdManager, f *DdNode, value CUDD_VALUE_TYPE) *DdNode {
	return core.Cudd_addBddStrictThreshold(dd, f, value)
}
func Cudd_addBddInterval(dd *DdManager, f *DdNode, lower, upper CUDD_VALUE_TYPE) *DdNode {
	return core.Cudd_addBddInterval(dd, f, lower, upper)
}
func Cudd_addBddIthBit(dd *DdManager, f *DdNode, bit int32) *DdNode {
	return core.Cudd_addBddIthBit(dd, f, int(bit))
}
func Cudd_BddToAdd(dd *DdManager, f *DdNode) *DdNode      { return core.Cudd_BddToAdd(dd, f) }
func Cudd_addBddPattern(dd *DdManager, f *DdNode) *DdNode { return core.Cudd_addBddPattern(dd, f) }
func Cudd_bddTransfer(source, destination *DdManager, f *DdNode) *DdNode {
	return core.Cudd_bddTransfer(source, destination, f)
}
func cuddBddTransfer(source, destination *DdManager, f *DdNode) *DdNode {
	return Cudd_bddTransfer(source, destination, f)
}
func cuddAddBddDoPattern(dd *DdManager, f *DdNode) *DdNode { return Cudd_addBddPattern(dd, f) }
func ddBddToAddRecur(dd *DdManager, f *DdNode) *DdNode     { return Cudd_BddToAdd(dd, f) }
