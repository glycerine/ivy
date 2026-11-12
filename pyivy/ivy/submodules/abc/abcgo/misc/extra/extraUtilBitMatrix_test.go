package extra

import "testing"

func TestBitMatrixTriangularRelations(t *testing.T) {
	p := Extra_BitMatrixStart(70)
	for _, edge := range [][2]int32{{0, 32}, {1, 69}, {31, 33}} {
		Extra_BitMatrixInsert1(p, edge[1], edge[0])
		if Extra_BitMatrixLookup1(p, edge[0], edge[1]) != 1 {
			t.Fatal(edge)
		}
		if Extra_BitMatrixLookup2(p, edge[0], edge[1]) != 0 {
			t.Fatal("triangles overlap")
		}
	}
	if Extra_BitMatrixCountOnesUpper(p) != 3 {
		t.Fatal("wrong count")
	}
	Extra_BitMatrixInsert2(p, 0, 32)
	Extra_BitMatrixDelete1(p, 0, 32)
	if Extra_BitMatrixLookup2(p, 32, 0) != 1 || Extra_BitMatrixLookup1(p, 0, 32) != 0 {
		t.Fatal("independent triangles")
	}
	Extra_BitMatrixDelete2(p, 32, 0)
	Extra_BitMatrixClean(p)
	if Extra_BitMatrixCountOnesUpper(p) != 0 {
		t.Fatal("clean")
	}
}
func TestBitMatrixCliqueAndDisjoint(t *testing.T) {
	p, q := Extra_BitMatrixStart(4), Extra_BitMatrixStart(4)
	Extra_BitMatrixInsert1(p, 0, 1)
	Extra_BitMatrixInsert1(p, 1, 2)
	if Extra_BitMatrixIsClique(p) != 0 {
		t.Fatal("path is not clique")
	}
	Extra_BitMatrixInsert1(p, 0, 2)
	if Extra_BitMatrixIsClique(p) != 1 {
		t.Fatal("triangle")
	}
	if Extra_BitMatrixIsDisjoint(p, q) != 1 {
		t.Fatal("empty intersection")
	}
	Extra_BitMatrixInsert1(q, 0, 2)
	if Extra_BitMatrixIsDisjoint(p, q) != 0 {
		t.Fatal("common edge")
	}
	Extra_BitMatrixOr(q, 3, []uint32{1})
	Extra_BitMatrixOrTwo(q, 2, 3)
	if q.ppData[2][0] != q.ppData[3][0] || q.ppData[2][0]&1 == 0 {
		t.Fatal("row union")
	}
	Extra_BitMatrixStop(q)
	if q.ppData != nil {
		t.Fatal("stop")
	}
	Extra_BitMatrixClean(Extra_BitMatrixStart(0))
}
