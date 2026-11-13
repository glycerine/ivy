package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestFullAdderTimingAndExtraAig(t *testing.T) {
	table := Gia_ManGenerateDelayTableFloat(3, 2)
	if len(table) != 9 || table[1] != 3 || table[2] != 2 || table[6] != -float32(util.ABC_INFINITY) {
		t.Fatalf("table %v", table)
	}
	m := Gia_ManGenerateTim(4, 3, 2, 3, 2)
	if tim.Tim_ManCiNum(m) != 8 || tim.Tim_ManCoNum(m) != 9 || tim.Tim_ManBoxNum(m) != 2 {
		t.Fatal("incorrect timing manager dimensions")
	}
	p := Gia_ManGenerateExtraAig(2, 3, 2)
	for i := 0; i < 8; i++ {
		in := []bool{i&1 != 0, i&2 != 0, i&4 != 0}
		out, err := p.Evaluate(in)
		sum := in[0] != in[1] != in[2]
		carry := (in[0] && in[1]) || (in[0] && in[2]) || (in[1] && in[2])
		if err != nil || len(out) != 4 || out[0] != sum || out[1] != carry || out[2] != sum || out[3] != carry {
			t.Fatalf("%v -> %v error %v", in, out, err)
		}
	}
	if gia.Gia_ManCiNum(p) != 3 {
		t.Fatal("incorrect PI count")
	}
}

func TestExtraAigZeroBoxes(t *testing.T) {
	p := Gia_ManGenerateExtraAig(0, 3, 2)
	if gia.Gia_ManCiNum(p) != 3 || gia.Gia_ManCoNum(p) != 0 {
		t.Fatal("incorrect zero-box graph")
	}
}
