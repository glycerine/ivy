package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

// Gia_ManBuild1Hot_rec constructs zero-hot and exactly-one-hot functions.
func Gia_ManBuild1Hot_rec(p *Gia_Man_t, lits []int32, nLits int32, zero, one *int32) {
	if nLits <= 0 || int(nLits) > len(lits) {
		panic("gia: invalid one-hot input count")
	}
	if nLits == 1 {
		*zero, *one = lits[0]^1, lits[0]
		return
	}
	mid := nLits / 2
	var z0, o0, z1, o1 int32
	Gia_ManBuild1Hot_rec(p, lits[:mid], mid, &z0, &o0)
	Gia_ManBuild1Hot_rec(p, lits[mid:], nLits-mid, &z1, &o1)
	*zero = Gia_ManHashAnd(p, z0, z1)
	*one = Gia_ManHashOr(p, Gia_ManHashAnd(p, z0, o1), Gia_ManHashAnd(p, z1, o0))
}
func Gia_ManBuild1Hot(p *Gia_Man_t, lits *Vec_Int_t) int32 {
	var zero, one int32
	Gia_ManBuild1Hot_rec(p, vec.Vec_IntArray(lits), int32(vec.Vec_IntSize(lits)), &zero, &one)
	return one
}
func Gia_SymbSpecial(c byte) int32 {
	if c == '(' || c == ')' || c == '*' || c == '|' {
		return 1
	}
	return 0
}
func Gia_ManRexNumInputs(s string, alphas, states **Vec_Int_t) int32 {
	a, mapping := vec.Vec_IntAlloc(len(s)), vec.Vec_IntStartFull(len(s)+1)
	var n int32
	for i := range s {
		c := s[i]
		if Gia_SymbSpecial(c) != 0 {
			continue
		}
		if vec.Vec_IntFind(a, int32(c)) < 0 {
			vec.Vec_IntPush(a, int32(c))
		}
		vec.Vec_IntWriteEntry(mapping, i, n)
		n++
	}
	vec.Vec_IntWriteEntry(mapping, len(s), n)
	*alphas, *states = a, mapping
	return n
}
func Gia_ManRexPreprocess(s string) string {
	var out strings.Builder
	out.Grow(2*len(s) + 4)
	out.WriteString("((")
	for i := range s {
		c := s[i]
		if c == '(' || c == ')' {
			out.WriteByte(c)
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			out.WriteByte(c)
		}
	}
	out.WriteString("))")
	return out.String()
}
