/**CFile***********************************************************************

  FileName    [cuddAddFind.c]

  PackageName [cudd]

  Synopsis    [Functions to find maximum and minimum in an ADD and to
  extract the i-th bit.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addFindMax()
                <li> Cudd_addFindMin()
                <li> Cudd_addIthBit()
                </ul>
               Static functions included in this module:
                <ul>
                <li> addDoIthBit()
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

func Cudd_addFindMax(dd *DdManager, f *DdNode) *DdNode { return core.Cudd_addFindMax(dd, f) }
func Cudd_addFindMin(dd *DdManager, f *DdNode) *DdNode { return core.Cudd_addFindMin(dd, f) }
func Cudd_addIthBit(dd *DdManager, f *DdNode, bit int32) *DdNode {
	return core.Cudd_addIthBit(dd, f, int(bit))
}
