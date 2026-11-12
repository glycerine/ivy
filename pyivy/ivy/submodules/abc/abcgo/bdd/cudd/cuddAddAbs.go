/**CFile***********************************************************************

  FileName    [cuddAddAbs.c]

  PackageName [cudd]

  Synopsis    [Quantification functions for ADDs.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addExistAbstract()
                <li> Cudd_addUnivAbstract()
                <li> Cudd_addOrAbstract()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddAddExistAbstractRecur()
                <li> cuddAddUnivAbstractRecur()
                <li> cuddAddOrAbstractRecur()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> addCheckPositiveCube()
                </ul>]

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

func Cudd_addExistAbstract(dd *DdManager, f, cube *DdNode) *DdNode {
	return core.Cudd_addExistAbstract(dd, f, cube)
}
func Cudd_addUnivAbstract(dd *DdManager, f, cube *DdNode) *DdNode {
	return core.Cudd_addUnivAbstract(dd, f, cube)
}
func Cudd_addOrAbstract(dd *DdManager, f, cube *DdNode) *DdNode {
	return core.Cudd_addOrAbstract(dd, f, cube)
}
func cuddAddExistAbstractRecur(dd *DdManager, f, cube *DdNode) *DdNode {
	return Cudd_addExistAbstract(dd, f, cube)
}
func cuddAddUnivAbstractRecur(dd *DdManager, f, cube *DdNode) *DdNode {
	return Cudd_addUnivAbstract(dd, f, cube)
}
func cuddAddOrAbstractRecur(dd *DdManager, f, cube *DdNode) *DdNode {
	return Cudd_addOrAbstract(dd, f, cube)
}
