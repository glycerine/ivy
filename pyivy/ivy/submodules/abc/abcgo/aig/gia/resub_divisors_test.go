package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubUnateAndXorDivisors(t *testing.T) {
	divs := vec.Vec_PtrAlloc(4)
	a, b := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc)
	off, on := ^(a ^ b), a^b
	for _, v := range []word{off, on, a, b} {
		vec.Vec_PtrPush(divs, []word{v})
	}
	binate := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(binate, 2)
	vec.Vec_IntPush(binate, 3)
	pairs := vec.Vec_IntAlloc(0)
	Gia_ManFindXorInt([]word{off}, []word{on}, binate, divs, 1, pairs)
	if vec.Vec_IntSize(pairs) != 1 {
		t.Fatal("missing XOR pair")
	}
	result := []word{0}
	Gia_ManDeriveDivPair(vec.Vec_IntEntry(pairs, 0), divs, 1, result)
	if result[0] != on {
		t.Fatalf("XOR %x", result[0])
	}
	unate, notUnate := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	Gia_ManFindOneUnateInt([]word{^a}, []word{a}, divs, 1, unate, notUnate)
	if vec.Vec_IntSize(unate) != 1 || vec.Vec_IntEntry(unate, 0) != 4 {
		t.Fatalf("unate %v", vec.Vec_IntArray(unate))
	}
}
func TestResubUnatePairTruth(t *testing.T) {
	a, b := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc)
	on := a &^ b
	divs := vec.Vec_PtrAlloc(4)
	for _, v := range []word{^on, on, a, b} {
		vec.Vec_PtrPush(divs, []word{v})
	}
	vars := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(vars, 2)
	vec.Vec_IntPush(vars, 3)
	pairs := vec.Vec_IntAlloc(0)
	Gia_ManFindUnatePairsInt([]word{^on}, []word{on}, vars, divs, 1, pairs)
	found := false
	for _, pair := range vec.Vec_IntArray(pairs) {
		result := []word{0}
		Gia_ManDeriveDivPair(pair, divs, 1, result)
		if result[0] == on {
			found = true
		}
	}
	if !found {
		t.Fatal("missing complemented AND pair")
	}
}
