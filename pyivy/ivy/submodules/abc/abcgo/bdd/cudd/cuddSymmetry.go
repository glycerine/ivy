package cudd

/**CFile***********************************************************************

  FileName    [cuddSymmetry.c]

  PackageName [cudd]

  Synopsis    [Functions for symmetry-based variable reordering.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_SymmProfile()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddSymmCheck()
                <li> cuddSymmSifting()
                <li> cuddSymmSiftingConv()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> ddSymmUniqueCompare()
                <li> ddSymmSiftingAux()
                <li> ddSymmSiftingConvAux()
                <li> ddSymmSiftingUp()
                <li> ddSymmSiftingDown()
                <li> ddSymmGroupMove()
                <li> ddSymmGroupMoveBackward()
                <li> ddSymmSiftingBackward()
                <li> ddSymmSummary()
                </ul>]

  Author      [Shipra Panda, Fabio Somenzi]

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

import core "github.com/glycerine/abcgo/bdd/cudd/core"

// Group discovery and movement are implemented in the authoritative core.
func Cudd_SymmProfile(dd *DdManager, lower, upper int32) {
	core.Cudd_SymmProfile(dd, int(lower), int(upper))
}
