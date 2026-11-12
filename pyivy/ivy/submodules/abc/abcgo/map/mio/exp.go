// Compact AND-expression helpers from exp.h. Expressions are built root first,
// then reversed once for topological evaluation, exactly as in the C API.
package mio

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

const (
	EXP_CONST0 int32 = -1
	EXP_CONST1 int32 = -2
)

func mioBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func expVector(a ...int32) *Vec_Int_t { return vec.Vec_IntAllocArrayCopy(a, len(a)) }
func Exp_Const0() *Vec_Int_t          { return expVector(EXP_CONST0) }
func Exp_Const1() *Vec_Int_t          { return expVector(EXP_CONST1) }
func Exp_IsConst(p *Vec_Int_t) int32  { return mioBool(Exp_IsConst0(p) != 0 || Exp_IsConst1(p) != 0) }
func Exp_IsConst0(p *Vec_Int_t) int32 { return mioBool(vec.Vec_IntEntry(p, 0) == EXP_CONST0) }
func Exp_IsConst1(p *Vec_Int_t) int32 { return mioBool(vec.Vec_IntEntry(p, 0) == EXP_CONST1) }
func Exp_Var(i int32) *Vec_Int_t      { return expVector(2 * i) }
func Exp_LitShift(n, lit, shift int32) int32 {
	if lit < 2*n {
		return lit
	}
	return lit + 2*shift
}
func Exp_IsLit(p *Vec_Int_t) int32   { return mioBool(vec.Vec_IntSize(p) == 1 && Exp_IsConst(p) == 0) }
func Exp_NodeNum(p *Vec_Int_t) int32 { return int32(vec.Vec_IntSize(p) / 2) }
func Exp_Not(p *Vec_Int_t) *Vec_Int_t {
	vec.Vec_IntWriteEntry(p, 0, vec.Vec_IntEntry(p, 0)^1)
	return p
}
func Exp_Reverse(p *Vec_Int_t) *Vec_Int_t {
	a := vec.Vec_IntArray(p)
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return p
}
func Exp_And(man *int32, n int32, a, b *Vec_Int_t, ca, cb int32) *Vec_Int_t {
	x, y := vec.Vec_IntArray(a), vec.Vec_IntArray(b)
	if len(x)%2 != 1 || len(y)%2 != 1 {
		panic("invalid expression length")
	}
	r := expVector(2*(n+int32(len(x)/2+len(y)/2)), Exp_LitShift(n, x[0]^ca, int32(len(y)/2)), y[0]^cb)
	for _, lit := range x[1:] {
		vec.Vec_IntPush(r, Exp_LitShift(n, lit, int32(len(y)/2)))
	}
	for _, lit := range y[1:] {
		vec.Vec_IntPush(r, lit)
	}
	return r
}
func Exp_Or(man *int32, n int32, a, b *Vec_Int_t) *Vec_Int_t {
	return Exp_Not(Exp_And(man, n, a, b, 1, 1))
}
func Exp_Xor(man *int32, n int32, a, b *Vec_Int_t) *Vec_Int_t {
	x, y := vec.Vec_IntArray(a), vec.Vec_IntArray(b)
	if len(x)%2 != 1 || len(y)%2 != 1 {
		panic("invalid expression length")
	}
	base := n + int32(len(x)/2+len(y)/2)
	r := expVector(2*(base+2), 2*(base+1)+1, 2*base+1, Exp_LitShift(n, x[0]^1, int32(len(y)/2)), y[0], Exp_LitShift(n, x[0], int32(len(y)/2)), y[0]^1)
	for _, lit := range x[1:] {
		vec.Vec_IntPush(r, Exp_LitShift(n, lit, int32(len(y)/2)))
	}
	for _, lit := range y[1:] {
		vec.Vec_IntPush(r, lit)
	}
	return Exp_Not(r)
}

var expTruth6 = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func Exp_Truth6Lit(n, lit int32, inputs, nodes []uint64) uint64 {
	if lit == EXP_CONST0 {
		return 0
	}
	if lit == EXP_CONST1 {
		return ^uint64(0)
	}
	var value uint64
	if lit < 2*n {
		value = inputs[lit/2]
	} else {
		value = nodes[lit/2-n]
	}
	if lit&1 != 0 {
		return ^value
	}
	return value
}
func Exp_Truth6(n int32, p *Vec_Int_t, inputs []uint64) uint64 {
	if inputs == nil {
		if n > 6 {
			panic("Exp_Truth6: more than six inputs")
		}
		inputs = expTruth6[:]
	}
	a := vec.Vec_IntArray(p)
	nodes := make([]uint64, len(a)/2)
	for i := range nodes {
		nodes[i] = Exp_Truth6Lit(n, a[2*i], inputs, nodes) & Exp_Truth6Lit(n, a[2*i+1], inputs, nodes)
	}
	return Exp_Truth6Lit(n, a[len(a)-1], inputs, nodes)
}
func Exp_TruthLit(n, lit int32, inputs, nodes [][]uint64, result []uint64, words int32) {
	if int(words) > len(result) {
		panic("short expression truth output")
	}
	for w := int32(0); w < words; w++ {
		var v uint64
		switch {
		case lit == EXP_CONST0:
			v = 0
		case lit == EXP_CONST1:
			v = ^uint64(0)
		default:
			if lit < 2*n {
				v = inputs[lit/2][w]
			} else {
				v = nodes[lit/2-n][w]
			}
			if lit&1 != 0 {
				v = ^v
			}
		}
		result[w] = v
	}
}
func expElementary(n int32, words int) [][]uint64 {
	inputs := make([][]uint64, n)
	for i := int32(0); i < n; i++ {
		inputs[i] = make([]uint64, words)
		for w := range inputs[i] {
			if i < 6 {
				inputs[i][w] = expTruth6[i]
			} else if w&(1<<uint(i-6)) != 0 {
				inputs[i][w] = ^uint64(0)
			}
		}
	}
	return inputs
}
func expEvaluate(n int32, p *Vec_Int_t, inputs [][]uint64, result []uint64, words int) {
	a := vec.Vec_IntArray(p)
	if len(a) == 0 || len(a)%2 == 0 {
		panic("invalid expression")
	}
	nodes := make([][]uint64, len(a)/2)
	left, right := make([]uint64, words), make([]uint64, words)
	for i := range nodes {
		nodes[i] = make([]uint64, words)
		Exp_TruthLit(n, a[2*i], inputs, nodes, left, int32(words))
		Exp_TruthLit(n, a[2*i+1], inputs, nodes, right, int32(words))
		for w := range left {
			nodes[i][w] = left[w] & right[w]
		}
	}
	Exp_TruthLit(n, a[len(a)-1], inputs, nodes, result, int32(words))
}
func Exp_Truth8(n int32, p *Vec_Int_t, inputs [][]uint64, result []uint64) {
	if n < 0 || n > 8 {
		panic("invalid eight-variable expression")
	}
	if inputs == nil {
		inputs = expElementary(n, 4)
	}
	expEvaluate(n, p, inputs, result, 4)
}
func Exp_Truth(n int32, p *Vec_Int_t, result []uint64) {
	if n < 0 || n > 20 {
		panic("expression truth variable count outside 0..20")
	}
	words := 1
	if n > 6 {
		words = 1 << uint(n-6)
	}
	expEvaluate(n, p, expElementary(n, words), result, words)
}
func Exp_PrintLit(n, lit int32) {
	switch {
	case lit == EXP_CONST0:
		fmt.Print("Const0")
	case lit == EXP_CONST1:
		fmt.Print("Const1")
	default:
		if lit&1 != 0 {
			fmt.Print("!")
		} else {
			fmt.Print(" ")
		}
		if lit < 2*n {
			fmt.Printf("%c", 'a'+lit/2)
		} else {
			fmt.Print(lit / 2)
		}
	}
}
func Exp_Print(n int32, p *Vec_Int_t) {
	a := vec.Vec_IntArray(p)
	for i := 0; i < len(a)/2; i++ {
		fmt.Printf("%2d = ", int(n)+i)
		Exp_PrintLit(n, a[2*i])
		fmt.Print(" & ")
		Exp_PrintLit(n, a[2*i+1])
		fmt.Println()
	}
	fmt.Print(" F = ")
	Exp_PrintLit(n, a[len(a)-1])
	fmt.Println()
}
func Exp_PrintReverse(n int32, p *Vec_Int_t) { Exp_Reverse(p); defer Exp_Reverse(p); Exp_Print(n, p) }
