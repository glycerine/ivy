package gia

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestTasSolveMiterStatusesAndWitnesses(t *testing.T) {
	g := Gia_ManStart(20)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashStart(g)
	positive := Gia_ManHashOr(g, a, b)
	left := Gia_ManHashOr(g, a, b)
	right := Gia_ManHashOr(g, a^1, b)
	contradiction := Gia_ManHashAnd(g, Gia_ManHashAnd(g, left, right), b^1)
	for _, lit := range []int32{0, 1, positive, contradiction} {
		Gia_ManAppendCo(g, lit)
	}
	var statuses *vec.Vec_Str_t
	store := Tas_ManSolveMiterNc(g, 10000, &statuses, 0)
	if got := vec.Vec_StrArray(statuses); !reflect.DeepEqual(got, []byte{1, 0, 0, 1}) {
		t.Fatalf("statuses = %v", got)
	}
	data := vec.Vec_IntArray(store)
	for offset := 0; offset < len(data); {
		output, count := int(data[offset]), int(data[offset+1])
		offset += 2
		if count < 0 || offset+count > len(data) {
			t.Fatal("invalid counterexample encoding")
		}
		inputs := make([]bool, 2)
		for _, lit := range data[offset : offset+count] {
			inputs[lit>>1] = lit&1 == 0
		}
		values, err := g.Evaluate(inputs)
		if err != nil || !values[output] {
			t.Fatalf("output %d witness %v values %v: %v", output, inputs, values, err)
		}
		offset += count
	}
}

func TestTasMiterPacksPatternsAndMarksProved(t *testing.T) {
	g := Gia_ManStart(30)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashStart(g)
	sat := Gia_ManHashAnd(g, a, b)
	left := Gia_ManHashOr(g, a, b)
	right := Gia_ManHashOr(g, a^1, b)
	unsat := Gia_ManHashAnd(g, Gia_ManHashAnd(g, left, right), b^1)
	Gia_ManAppendCo(g, sat)
	Gia_ManAppendCo(g, unsat)
	g.PReprs = make([]Gia_Rpr_t, g.nObjs)
	roots := []*Gia_Obj_t{Gia_ManObj(g, a>>1), Gia_ManObj(g, b>>1)}
	simulation := [][]uint32{{0}, {0}}
	Tas_ManSolveMiterNc2(g, 10000, g, roots, simulation)
	if simulation[0][0]&2 == 0 || simulation[1][0]&2 == 0 {
		t.Fatalf("SAT pattern not packed: %v", simulation)
	}
	if Gia_ObjProved(g, b>>1) == 0 || Gia_ObjProved(g, a>>1) != 0 {
		t.Fatal("incorrect proved-root marking")
	}
}
