package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestQbfGeneratedMiter(t *testing.T) {
	p := Gia_Gen2CreateMiter(1, 1)
	for mask := 0; mask < 16; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		code := func(x bool) bool {
			if x {
				return input[1]
			}
			return input[0]
		}
		x, y := input[2], input[3]
		want := input[0] != input[1] && code(!(x && y)) == !(code(x) && code(y))
		out, err := p.Evaluate(input)
		if err != nil || len(out) != 1 || out[0] != want {
			t.Fatalf("input %v got %v want %t: %v", input, out, want, err)
		}
	}
}
func TestQbfCodeWords(t *testing.T) {
	code := vec.Vec_IntAlloc(140)
	for i := 0; i < 70; i++ {
		vec.Vec_IntPush(code, 0)
		vec.Vec_IntPush(code, int32(i%2))
	}
	words := Gia_Gen2CodeOneP(1, 70, code, 1)
	if len(words) != 2 || words[0] != 0xaaaaaaaaaaaaaaaa || words[1] != 0x2a {
		t.Fatalf("code %x", words)
	}
}
