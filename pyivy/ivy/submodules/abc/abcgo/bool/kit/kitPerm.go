package kit

/**CFile****************************************************************

  FileName    [kitPerm.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [Procedures for permuting truth tables.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Oct 26, 2011.]

  Revision    [$Id: kitPerm.c,v 1.00 2011/11/26 00:00:00 alanmi Exp $]

***********************************************************************/

// Truth tables use little-endian words: bit x is the value at assignment x.
// The permutation reverses the order of the truth table's input variables.
import (
	"fmt"
	"math/bits"
	"math/rand"
	"time"
)

type word = uint64
type shot = uint16

var S = func() (table [256]shot) {
	for x := range table {
		for b := 0; b < 8; b++ {
			table[x] |= shot((x>>b)&1) << (2 * b)
		}
	}
	return
}()
var P = func() (table [256]byte) {
	for x := range table {
		a, b := x&0x0a, x&0x50
		table[x] = byte(x ^ a ^ b ^ (a << 3) ^ (b >> 3))
	}
	return
}()

func Kit_PermCreateS() {
	for i, x := range S {
		if i%8 == 0 {
			fmt.Println()
		}
		fmt.Printf("0x%04x, ", x)
	}
}
func Kit_PermCreateP() {
	for i, x := range P {
		if i%16 == 0 {
			fmt.Println()
		}
		fmt.Printf("0x%02x, ", x)
	}
}

func Kit_PermCycleOne(s []shot, b []byte, v int32) {
	if v <= 2 || v >= 16 {
		panic("Kit_PermCycleOne: variable out of range")
	}
	n := 1 << uint(v-3)
	if len(s) < n || len(b) < 2*n {
		panic("Kit_PermCycleOne: short buffer")
	}
	for i := 0; i < n; i++ {
		s[i] = S[b[i]] | (S[b[i+n]] << 1)
	}
}
func Kit_PermCycleMany(s []shot, b []byte, V, v int32) {
	if v <= 2 || v >= V {
		panic("Kit_PermCycleMany: variable out of range")
	}
	n, m := 1<<uint(V-1-v), 1<<uint(v-2)
	for i := 0; i < n; i++ {
		Kit_PermCycleOne(s[i*(m>>1):], b[i*m:], v)
	}
}

// Kit_PermCompute writes the variable-order reversal into out. Slices replace
// the C buffer pointers; input is read before writing so aliased slices work.
func Kit_PermCompute(out, in []word, V int32) {
	if V < 6 || V > 16 {
		panic("Kit_PermCompute: require 6..16 variables")
	}
	n := 1 << uint(V-6)
	if len(out) < n || len(in) < n {
		panic("Kit_PermCompute: short truth table")
	}
	source := append([]word(nil), in[:n]...)
	clear(out[:n])
	for w, x := range source {
		for x != 0 {
			b := bits.TrailingZeros64(x)
			index := uint16(w*64 + b)
			target := int(bits.Reverse16(index) >> uint(16-V))
			out[target>>6] |= uint64(1) << uint(target&63)
			x &= x - 1
		}
	}
}
func Kit_PermComputeNaive(table []word, nVars int32) {
	Kit_PermCompute(table, table, nVars)
}

// M interleaves the low n bits of its arguments. Tf is the original recursive
// single-word transpose, useful as a separate reference for six variables.
func M(f1, f2 word, n int32) word {
	if n < 0 || n > 32 {
		panic("M: bit count out of range")
	}
	var result word
	for i := int32(0); i < n; i++ {
		result |= ((f1>>uint(i))&1)<<uint(2*i) | ((f2>>uint(i))&1)<<uint(2*i+1)
	}
	return result
}
func Tf(f word, n int32) word {
	if n < 1 || n > 6 {
		panic("Tf: require 1..6 variables")
	}
	if n == 1 {
		return f
	}
	x := uint(1) << uint(n-1)
	mask := uint64(1)<<x - 1
	return M(Tf(f&mask, n-1), Tf(f>>x, n-1), int32(x))
}

func Kit_PermComputeTest() {
	const n = 1 << 20
	table := make([]word, n)
	rng := rand.New(rand.NewSource(0))
	start := time.Now()
	for i := range table {
		table[i] = rng.Uint64()
	}
	fmt.Printf("Assign = %.2f sec\n", time.Since(start).Seconds())
	var result uint64
	start = time.Now()
	for _, x := range table {
		result ^= Tf(x, 6)
	}
	fmt.Printf("Perm1 = %.2f sec\n", time.Since(start).Seconds())
	start = time.Now()
	var output [1]word
	for i := range table {
		Kit_PermCompute(output[:], table[i:i+1], 6)
		result ^= output[0]
	}
	fmt.Printf("Perm2 = %.2f sec\n", time.Since(start).Seconds())
	if result != 0 {
		panic("permutation implementations disagree")
	}
}
func Kit_PermComputeTest_() { _ = Tf(0xacaccacaaccaacca, 6) }
