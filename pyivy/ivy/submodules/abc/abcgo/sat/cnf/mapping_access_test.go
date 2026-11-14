package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMappingAccessPreservesGraphCutsAndCover(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	data := Cnf_Derive(g, 1)
	if data == nil {
		t.Fatal("CNF derivation failed")
	}
	m := Cnf_ManStart()
	m.pManAig = g
	defer Cnf_ManStop(m)
	if m.MappingGraph() != g {
		t.Fatal("wrong source graph")
	}
	count := 0
	for _, clause := range Cnf_DataClauses(data) {
		count += len(clause)
	}
	if Cnf_DataLiteralNum(data) != int32(count) {
		t.Fatal("literal count")
	}
	for _, truth := range []uint16{0, 1, 0xAAAA, 0x8888, 0xFFFF} {
		sop, cubes := m.TruthCover(truth)
		v := vec.Vec_IntAlloc(16)
		Cnf_SopConvertToVector(sop, cubes, v)
		if int32(vec.Vec_IntSize(v)) != cubes {
			t.Fatalf("truth %04x: cover size", truth)
		}
	}
	c := Cnf_CutFromTruth([]int32{1, 2, 3, 4, 5}, []uint32{0x80000000})
	if c.PositiveCover() != c.vIsop[1] {
		t.Fatal("cover must be borrowed")
	}
	Cnf_DataFree(data)
	Cnf_ManFree()
}
