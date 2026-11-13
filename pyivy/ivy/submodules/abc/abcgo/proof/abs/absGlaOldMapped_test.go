package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestDuplicateMappedCones(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, c^1)
	gia.Gia_ManAppendCo(g, y)
	n := int(gia.Gia_ManObjNum(g))
	mapping := vec.Vec_IntStart(n + 10)
	for i, entry := range [][3]int32{{x >> 1, a >> 1, b >> 1}, {y >> 1, x >> 1, c >> 1}} {
		off := n + 5*i
		vec.Vec_IntWriteEntry(mapping, int(entry[0]), int32(off))
		vec.Vec_IntWriteEntry(mapping, off+1, entry[1])
		vec.Vec_IntWriteEntry(mapping, off+2, entry[2])
		vec.Vec_IntWriteEntry(mapping, off+3, -1)
		vec.Vec_IntWriteEntry(mapping, off+4, -1)
	}
	result := Gia_ManDupMapped(g, mapping)
	if vec.Vec_IntSize(gia.Gia_ManLutConfigs(result)) != int(gia.Gia_ManObjNum(result)) {
		t.Fatal("missing provenance")
	}
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
