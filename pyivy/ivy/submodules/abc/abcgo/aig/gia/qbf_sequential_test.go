package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestQbfRegisterSelection(t *testing.T) {
	if got := vec.Vec_IntArray(Gia_GenCollectFlopIndexes("", 1, 2, 3)); !reflect.DeepEqual(got, []int32{0, 1}) {
		t.Fatal(got)
	}
	if got := vec.Vec_IntArray(Gia_GenCollectFlopIndexes("2, 0", 1, 2, 3)); !reflect.DeepEqual(got, []int32{2, 0}) {
		t.Fatal(got)
	}
	if Gia_GenCollectFlopIndexes("1,1", 1, 2, 3) != nil || Gia_GenCollectFlopIndexes("-1,0", 1, 2, 3) != nil {
		t.Fatal("invalid selection accepted")
	}
}
func TestQbfSequentialMiter(t *testing.T) {
	g := Gia_ManStart(8)
	pi := Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, pi)
	Gia_ManAppendCo(g, pi)
	Gia_ManSetRegNum(g, 1)
	q := Gia_GenQbfMiter(g, 1, 1, 1, "", 0, 0)
	for bits := 0; bits < 16; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		lut := func(x bool) bool {
			if x {
				return in[1]
			}
			return true
		}
		want := in[0] && !(lut(in[3]) && !lut(in[2]))
		got, err := q.Evaluate(in)
		if err != nil || got[0] != want {
			t.Fatalf("%v got %v want %t: %v", in, got, want, err)
		}
	}
}
