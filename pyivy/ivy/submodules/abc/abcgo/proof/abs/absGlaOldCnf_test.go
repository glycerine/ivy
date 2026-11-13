package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestGlaOldMappedCnfFeedsDuplication(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, c^1)
	gia.Gia_ManAppendCo(g, y)
	data, err := cnf.Cnf_DeriveOtherMapped(gia.Gia_ManToAigSimple(g), 1)
	if err != nil {
		t.Fatal(err)
	}
	mapping := vec.Vec_IntAllocArrayCopy(data.Mapping, len(data.Mapping))
	result := Gia_ManDupMapped(g, mapping)
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := result.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if want := inputs[0] && inputs[1] && !inputs[2]; got[0] != want {
			t.Fatalf("inputs=%v got=%v", inputs, got)
		}
	}
}
