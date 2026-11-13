package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDivisorRecognition(t *testing.T) {
	cases := []struct {
		a, b  [2]int32
		kind  Div_Type_t
		phase int32
	}{
		{[2]int32{0, -1}, [2]int32{2, -1}, DIV_AND, 1},
		{[2]int32{0, 2}, [2]int32{1, 3}, DIV_XOR, 1},
		{[2]int32{0, 2}, [2]int32{1, 4}, DIV_MUX, 0},
		{[2]int32{0, 2}, [2]int32{4, 6}, DIV_NONE, 0},
	}
	for _, tc := range cases {
		var lits [3]int32
		var phase int32
		got := Div_FindType(tc.a[:], tc.b[:], lits[:], &phase)
		if got != tc.kind || phase != tc.phase {
			t.Fatalf("%v %v got %d phase %d", tc.a, tc.b, got, phase)
		}
	}
	a := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 4}, 3)
	b := vec.Vec_IntAllocArrayCopy([]int32{0, 3, 6}, 3)
	la, lb := make([]int32, 2), make([]int32, 2)
	if got := Div_FindDiv(a, b, la, lb); got != 1 || la[0] != 2 || la[1] != 4 || lb[0] != 3 || lb[1] != 6 {
		t.Fatalf("%d %v %v", got, la, lb)
	}
}
