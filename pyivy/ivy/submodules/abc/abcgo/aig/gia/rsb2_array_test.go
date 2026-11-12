package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRsb2ArrayConversionAndReplacement(t *testing.T) {
	// const, two inputs, XOR, AND of XOR and first input, output.
	raw := []int32{0, 0, 0, 0, 0, 0, 4, 2, 2, 6, 8, 8}
	g := Gia_ManFromResub(raw, 6, 2)
	for mask := 0; mask < 4; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0}
		out, err := g.Evaluate(input)
		if err != nil || out[0] != (input[0] && !input[1]) {
			t.Fatalf("%v: %v %v", input, out, err)
		}
	}
	original := vec.Vec_IntAlloc(len(raw))
	for _, v := range raw {
		vec.Vec_IntPush(original, v)
	}
	divs := vec.Vec_IntAlloc(4)
	for _, v := range []int32{-1, -1, 1, 2} {
		vec.Vec_IntPush(divs, v)
	}
	replacement := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(replacement, 4)
	copies := vec.Vec_IntAlloc(0)
	changed := Gia_Rsb2ManInsert(2, 1, original, 3, replacement, divs, copies)
	q := Gia_ManFromResub(vec.Vec_IntArray(changed), int32(vec.Vec_IntSize(changed)/2), 2)
	out, err := q.Evaluate([]bool{true, false})
	if err != nil || !out[0] {
		t.Fatalf("replacement %v %v", out, err)
	}
}
