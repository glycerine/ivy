/**CFile***********************************************************************

  FileName    [cuddApa.c]

  PackageName [cudd]

  Synopsis    [Arbitrary precision arithmetic functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_ApaNumberOfDigits()
                <li> Cudd_NewApaNumber()
                <li> Cudd_ApaCopy()
                <li> Cudd_ApaAdd()
                <li> Cudd_ApaSubtract()
                <li> Cudd_ApaShortDivision()
                <li> Cudd_ApaIntDivision()
                <li> Cudd_ApaShiftRight()
                <li> Cudd_ApaSetToLiteral()
                <li> Cudd_ApaPowerOfTwo()
                <li> Cudd_ApaCompare()
                <li> Cudd_ApaCompareRatios()
                <li> Cudd_ApaPrintHex()
                <li> Cudd_ApaPrintDecimal()
                <li> Cudd_ApaPrintExponential()
                <li> Cudd_ApaCountMinterm()
                <li> Cudd_ApaPrintMinterm()
                <li> Cudd_ApaPrintMintermExp()
                <li> Cudd_ApaPrintDensity()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> cuddApaCountMintermAux()
                <li> cuddApaStCountfree()
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

package core

import (
	"fmt"
	"io"
	"math/big"
	"math/bits"
)

// APA numbers retain CUDD's big-endian base-2^32 digit representation.
func Cudd_ApaNumberOfDigits(binaryDigits int) int {
	if binaryDigits <= 0 {
		return 0
	}
	return (binaryDigits-1)/32 + 1
}
func Cudd_NewApaNumber(digits int) DdApaNumber {
	if digits < 0 {
		return nil
	}
	return make(DdApaNumber, digits)
}
func Cudd_ApaCopy(digits int, source, dest DdApaNumber) { copy(dest[:digits], source[:digits]) }
func Cudd_ApaAdd(digits int, a, b, sum DdApaNumber) uint32 {
	var carry uint32
	for i := digits - 1; i >= 0; i-- {
		sum[i], carry = bits.Add32(a[i], b[i], carry)
	}
	return carry
}
func Cudd_ApaSubtract(digits int, a, b, difference DdApaNumber) uint32 {
	var borrow uint32
	for i := digits - 1; i >= 0; i-- {
		difference[i], borrow = bits.Sub32(a[i], b[i], borrow)
	}
	return 0 - borrow
}
func Cudd_ApaShortDivision(digits int, dividend DdApaNumber, divisor uint32, quotient DdApaNumber) uint32 {
	var remainder uint64
	for i := 0; i < digits; i++ {
		partial := remainder<<32 | uint64(dividend[i])
		quotient[i] = uint32(partial / uint64(divisor))
		remainder = partial % uint64(divisor)
	}
	return uint32(remainder)
}
func Cudd_ApaIntDivision(digits int, dividend DdApaNumber, divisor uint32, quotient DdApaNumber) uint32 {
	return Cudd_ApaShortDivision(digits, dividend, divisor, quotient)
}
func Cudd_ApaShiftRight(digits int, in uint32, a, b DdApaNumber) {
	if digits == 0 {
		return
	}
	for i := digits - 1; i > 0; i-- {
		b[i] = a[i]>>1 | (a[i-1]&1)<<31
	}
	b[0] = a[0]>>1 | in<<31
}
func Cudd_ApaSetToLiteral(digits int, number DdApaNumber, literal uint32) {
	clear(number[:digits])
	if digits > 0 {
		number[digits-1] = literal
	}
}
func Cudd_ApaPowerOfTwo(digits int, number DdApaNumber, power int) {
	clear(number[:digits])
	if power < 0 || power/32 >= digits {
		return
	}
	number[digits-1-power/32] = 1 << uint(power%32)
}
func apaInteger(digits int, number DdApaNumber) *big.Int {
	v := new(big.Int)
	for _, digit := range number[:digits] {
		v.Lsh(v, 32)
		v.Or(v, new(big.Int).SetUint64(uint64(digit)))
	}
	return v
}
func apaDigits(v *big.Int, digits int) DdApaNumber {
	out := make(DdApaNumber, digits)
	work := new(big.Int).Set(v)
	for i := digits - 1; i >= 0; i-- {
		out[i] = uint32(work.Uint64())
		work.Rsh(work, 32)
	}
	return out
}
func Cudd_ApaCompare(da int, a DdApaNumber, db int, b DdApaNumber) int {
	return apaInteger(da, a).Cmp(apaInteger(db, b))
}
func Cudd_ApaCompareRatios(da int, a DdApaNumber, denA uint32, db int, b DdApaNumber, denB uint32) int {
	if denA == 0 || denB == 0 {
		panic("cudd: APA ratio denominator is zero")
	}
	lhs := apaInteger(da, a)
	rhs := apaInteger(db, b)
	lhs.Mul(lhs, new(big.Int).SetUint64(uint64(denB)))
	rhs.Mul(rhs, new(big.Int).SetUint64(uint64(denA)))
	return lhs.Cmp(rhs)
}
func Cudd_ApaPrintHex(w io.Writer, digits int, n DdApaNumber) int {
	for _, digit := range n[:digits] {
		if _, err := fmt.Fprintf(w, "%08x", digit); err != nil {
			return 0
		}
	}
	return 1
}
func Cudd_ApaPrintDecimal(w io.Writer, digits int, n DdApaNumber) int {
	_, err := io.WriteString(w, apaInteger(digits, n).String())
	if err != nil {
		return 0
	}
	return 1
}
func Cudd_ApaPrintExponential(w io.Writer, digits int, n DdApaNumber, precision int) int {
	text := apaInteger(digits, n).String()
	count := min(max(precision, 0), len(text))
	significand := text[:count]
	if len(significand) > 1 {
		significand = significand[:1] + "." + significand[1:]
	}
	_, err := fmt.Fprintf(w, "%se+%d", significand, len(text)-1)
	if err != nil {
		return 0
	}
	return 1
}
func Cudd_ApaCountMinterm(dd *DdManager, node *DdNode, nvars int, digits *int) DdApaNumber {
	if dd == nil || dd.closed || node == nil || node.manager != dd || Cudd_Regular(node).family == 1 || nvars < 0 || digits == nil {
		return nil
	}
	*digits = Cudd_ApaNumberOfDigits(nvars + 1)
	maximum := new(big.Int).Lsh(big.NewInt(1), uint(nvars))
	memo := map[*DdNode]*big.Int{}
	var visit func(*DdNode) *big.Int
	visit = func(n *DdNode) *big.Int {
		if value := memo[n]; value != nil {
			return new(big.Int).Set(value)
		}
		r := Cudd_Regular(n)
		var value *big.Int
		if Cudd_IsConstant(r) {
			if r == dd.Zero || r == dd.Background {
				value = new(big.Int)
			} else {
				value = new(big.Int).Set(maximum)
			}
		} else {
			value = new(big.Int).Add(visit(r.T), visit(r.E))
			value.Rsh(value, 1)
		}
		if Cudd_IsComplement(n) {
			value.Sub(maximum, value)
		}
		memo[n] = new(big.Int).Set(value)
		return value
	}
	return apaDigits(visit(node), *digits)
}
func Cudd_ApaPrintMinterm(w io.Writer, dd *DdManager, node *DdNode, nvars int) int {
	var digits int
	n := Cudd_ApaCountMinterm(dd, node, nvars, &digits)
	if n == nil {
		return 0
	}
	result := Cudd_ApaPrintDecimal(w, digits, n)
	if _, err := io.WriteString(w, "\n"); err != nil {
		return 0
	}
	return result
}
func Cudd_ApaPrintMintermExp(w io.Writer, dd *DdManager, node *DdNode, nvars, precision int) int {
	var digits int
	n := Cudd_ApaCountMinterm(dd, node, nvars, &digits)
	if n == nil {
		return 0
	}
	result := Cudd_ApaPrintExponential(w, digits, n, precision)
	if _, err := io.WriteString(w, "\n"); err != nil {
		return 0
	}
	return result
}
func Cudd_ApaPrintDensity(w io.Writer, dd *DdManager, node *DdNode, nvars int) int {
	var digits int
	n := Cudd_ApaCountMinterm(dd, node, nvars, &digits)
	if n == nil {
		return 0
	}
	size := uint32(Cudd_DagSize(node))
	quotient := make(DdApaNumber, digits)
	remainder := Cudd_ApaIntDivision(digits, n, size, quotient)
	result := Cudd_ApaPrintDecimal(w, digits, quotient)
	if _, err := fmt.Fprintf(w, ".%d\n", uint64(remainder)*1000000/uint64(size)); err != nil {
		return 0
	}
	return result
}
