package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestNpn3TableCanonicalClasses(t *testing.T) {
	classes := map[uint32]int32{}
	for i, entry := range Npn3Table {
		canonical := extra.Extra_TruthCanonNPN(uint32(i), 3)
		if canonical != uint32(entry[0]) {
			t.Fatalf("truth %02x: canonical %02x table %02x", i, canonical, entry[0])
		}
		class, ok := classes[canonical]
		if !ok {
			class = int32(len(classes))
			classes[canonical] = class
		}
		if class != entry[1] {
			t.Fatalf("truth %02x class=%d want=%d", i, entry[1], class)
		}
	}
	if len(classes) != 14 {
		t.Fatalf("classes %d", len(classes))
	}
}
func TestAcecStatsMatrixMappedAnd(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	size := int(gia.Gia_ManObjNum(g))
	mapping := vec.Vec_IntStart(size + 3)
	vec.Vec_IntWriteEntry(mapping, int(gate>>1), int32(size))
	vec.Vec_IntWriteEntry(mapping, size, 2)
	vec.Vec_IntWriteEntry(mapping, size+1, a>>1)
	vec.Vec_IntWriteEntry(mapping, size+2, b>>1)
	gia.Gia_ManSetMapping(g, mapping)
	matrix := acecStatsMatrix(g)
	if matrix[2] != 2 || matrix[2<<4] != 1 {
		t.Fatalf("unexpected class edges: %v", matrix)
	}
	sum := int32(0)
	for _, v := range matrix {
		sum += v
	}
	if sum != 3 {
		t.Fatalf("expected three edges, got %d", sum)
	}
}
