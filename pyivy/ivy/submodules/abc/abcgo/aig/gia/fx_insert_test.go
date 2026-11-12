package gia

import "testing"

func TestFxInsertExtractedNode(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y^1)
	p.pObjs[x>>1].Value = 3
	p.pObjs[y>>1].Value = 4
	// Owner 5 factors the common a&b term, used by owners 3 and 4.
	cubes := Vec_WecAlloc(3)
	for _, row := range [][]int32{{3, 10}, {4, 4, 10}, {5, 0, 2}} {
		cube := Vec_WecPushLevel(cubes)
		for _, value := range row {
			Vec_IntPush(cube, value)
		}
	}
	q := Gia_ManFxInsert(p, cubes, Vec_StrStart(5))
	for assignment := 0; assignment < 8; assignment++ {
		values := make([]uint32, q.nObjs)
		for i := 0; i < 3; i++ {
			values[Gia_ManCi(q, int32(i)).id] = uint32(assignment>>i) & 1
		}
		for _, obj := range q.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				values[obj.id] = (values[Gia_ObjFaninId0(obj, obj.id)] ^ obj.fCompl0) & (values[Gia_ObjFaninId1(obj, obj.id)] ^ obj.fCompl1)
			}
		}
		out := Gia_ManCo(q, 0)
		got := values[Gia_ObjFaninId0(out, out.id)] ^ out.fCompl0
		if got != uint32(giaBool(assignment != 7)) {
			t.Fatalf("assignment %d: got %d", assignment, got)
		}
	}
}

func TestFxDsdTruthTables(t *testing.T) {
	for table := 0; table < 256; table++ {
		p := Gia_ManStart(50)
		leaves := []int32{Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)}
		Gia_ManHashStart(p)
		truth := uint64(table) * 0x0101010101010101
		root := giaFxDsdToGia(p, truth, leaves)
		for assignment := 0; assignment < 8; assignment++ {
			values := make([]uint32, p.nObjs)
			for i, lit := range leaves {
				values[lit>>1] = uint32(assignment>>i) & 1
			}
			for _, obj := range p.pObjs {
				if Gia_ObjIsAnd(obj) != 0 {
					values[obj.id] = (values[Gia_ObjFaninId0(obj, obj.id)] ^ obj.fCompl0) & (values[Gia_ObjFaninId1(obj, obj.id)] ^ obj.fCompl1)
				}
			}
			got := values[root>>1] ^ uint32(root&1)
			if want := uint32(table>>assignment) & 1; got != want {
				t.Fatalf("truth %02x assignment %d: got %d want %d", table, assignment, got, want)
			}
		}
	}
}
