/**CFile***********************************************************************

  FileName    [cuddZddFuncs.c]

  PackageName [cudd]

  Synopsis    [Functions to manipulate covers represented as ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddProduct();
                    <li> Cudd_zddUnateProduct();
                    <li> Cudd_zddWeakDiv();
                    <li> Cudd_zddWeakDivF();
                    <li> Cudd_zddDivide();
                    <li> Cudd_zddDivideF();
                    <li> Cudd_zddComplement();
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddZddProduct();
                    <li> cuddZddUnateProduct();
                    <li> cuddZddWeakDiv();
                    <li> cuddZddWeakDivF();
                    <li> cuddZddDivide();
                    <li> cuddZddDivideF();
                    <li> cuddZddGetCofactors3()
                    <li> cuddZddGetCofactors2()
                    <li> cuddZddComplement();
                    <li> cuddZddGetPosVarIndex();
                    <li> cuddZddGetNegVarIndex();
                    <li> cuddZddGetPosVarLevel();
                    <li> cuddZddGetNegVarLevel();
                    </ul>
               Static procedures included in this module:
                    <ul>
                    </ul>
              ]

  SeeAlso     []

  Author      [In-Ho Moon]

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

// Completed cover algorithms forward to the sole authoritative implementation.
var (
	Cudd_zddProduct      = core.Cudd_zddProduct
	Cudd_zddUnateProduct = core.Cudd_zddUnateProduct
	Cudd_zddWeakDiv      = core.Cudd_zddWeakDiv
	Cudd_zddDivide       = core.Cudd_zddDivide
	Cudd_zddWeakDivF     = core.Cudd_zddWeakDivF
	Cudd_zddDivideF      = core.Cudd_zddDivideF
	cuddZddProduct       = core.Cudd_zddProduct
	cuddZddUnateProduct  = core.Cudd_zddUnateProduct
	cuddZddWeakDiv       = core.Cudd_zddWeakDiv
	cuddZddDivide        = core.Cudd_zddDivide
	cuddZddWeakDivF      = core.Cudd_zddWeakDivF
	cuddZddDivideF       = core.Cudd_zddDivideF
)

func cuddZddGetCofactors2(dd *DdManager, f *DdNode, index int32, positive, negative **DdNode) int32 {
	p, n, ok := core.ZDDCofactors(dd, f, int(index))
	if !ok || positive == nil || negative == nil {
		return 1
	}
	*positive = p
	*negative = n
	return 0
}
func cuddZddGetCofactors3(dd *DdManager, f *DdNode, index int32, positive, negative, absent **DdNode) int32 {
	p, n, d, ok := core.ZDDCoverCofactors(dd, f, int(index))
	if !ok || positive == nil || negative == nil || absent == nil {
		return 1
	}
	*positive = p
	*negative = n
	*absent = d
	return 0
}
func cuddZddGetPosVarIndex(dd *DdManager, index int32) int32 { return index &^ 1 }
func cuddZddGetNegVarIndex(dd *DdManager, index int32) int32 { return index | 1 }
func cuddZddGetPosVarLevel(dd *DdManager, index int32) int32 { return int32(dd.PermZ[index&^1]) }
func cuddZddGetNegVarLevel(dd *DdManager, index int32) int32 { return int32(dd.PermZ[index|1]) }
// REMAINING ORIGINAL DRAFT: ZDD complement depends on cover-to-BDD conversion
// and ISOP. These bodies deliberately remain visible and unported. Set-family
// difference is not substituted for logical cover complementation.
func Cudd_zddComplement(dd *DdManager, node *DdNode) *DdNode {
    var b, isop, zdd_I *DdNode;

    /* Check cache */
    zdd_I = cuddCacheLookup1Zdd(dd, cuddZddComplement, node);
    if zdd_I != nil {
        return(zdd_I); }

    b = Cudd_MakeBddFromZddCover(dd, node);
    if !b {
        return(nil); }
    Cudd_Ref(b);
    // inlining of macro Cudd_Not
    isop = Cudd_zddIsop(dd, ((DdNode *)((ptrint)(b) ^ 01)), ((DdNode *)((ptrint)(b) ^ 01)), &zdd_I);
    if !isop {
        Cudd_RecursiveDeref(dd, b);
        return(nil);
    }
    Cudd_Ref(isop);
    Cudd_Ref(zdd_I);
    Cudd_RecursiveDeref(dd, b);
    Cudd_RecursiveDeref(dd, isop);

    cuddCacheInsert1(dd, cuddZddComplement, node, zdd_I);
    Cudd_Deref(zdd_I);
    return(zdd_I);
}

func cuddZddComplement(dd *DdManager, node *DdNode) *DdNode {
    var b, isop, zdd_I *DdNode;

    /* Check cache */
    zdd_I = cuddCacheLookup1Zdd(dd, cuddZddComplement, node);
    if zdd_I != nil {
        return(zdd_I); }

    b = cuddMakeBddFromZddCover(dd, node);
    if !b {
        return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(b).ref);
    // inlining of macro Cudd_Not
    isop = cuddZddIsop(dd, ((DdNode *)((ptrint)(b) ^ 01)), ((DdNode *)((ptrint)(b) ^ 01)), &zdd_I);
    if !isop {
        Cudd_RecursiveDeref(dd, b);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(isop).ref);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(zdd_I).ref);
    Cudd_RecursiveDeref(dd, b);
    Cudd_RecursiveDeref(dd, isop);

    cuddCacheInsert1(dd, cuddZddComplement, node, zdd_I);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(zdd_I).ref);
    return(zdd_I);
}
