package cudd

/**CFile***********************************************************************

  FileName    [cuddReorder.c]

  PackageName [cudd]

  Synopsis    [Functions for dynamic variable reordering.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_ReduceHeap()
                <li> Cudd_ShuffleHeap()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddDynamicAllocNode()
                <li> cuddSifting()
                <li> cuddSwapping()
                <li> cuddNextHigh()
                <li> cuddNextLow()
                <li> cuddSwapInPlace()
                <li> cuddBddAlignToZdd()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> ddUniqueCompare()
                <li> ddSwapAny()
                <li> ddSiftingAux()
                <li> ddSiftingUp()
                <li> ddSiftingDown()
                <li> ddSiftingBackward()
                <li> ddReorderPreprocess()
                <li> ddReorderPostprocess()
                <li> ddShuffle()
                <li> ddSiftUp()
                <li> bddFixTree()
                </ul>]

  Author      [Shipra Panda, Bernard Plessier, Fabio Somenzi]

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

// Authoritative pure-Go dynamic reordering; root and core share node handles.
var (
	Cudd_ReduceHeap  = core.Cudd_ReduceHeap
	Cudd_ShuffleHeap = core.Cudd_ShuffleHeap
)
