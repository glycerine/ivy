package pdr

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestTxs3ConeOrdersInputsAndClearsMarks(t *testing.T) {
	g := aig.NewGraph()
	x, present, abstract := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(x, present), abstract)
	g.AddOutput(root)
	g.AddOutput(present)
	g.AddOutput(abstract)
	aig.Aig_ManSetRegNum(g, 2)
	priority := vec.Vec_IntAllocArrayCopy([]int32{1, 0}, 2)
	s := Txs3_ManStart(nil, g, priority)
	defer Txs3_ManStop(s)
	output := gia.Gia_ManCo(s.pGia, 0)
	vec.Vec_IntPush(s.vCoObjs, gia.Gia_ObjId(s.pGia, output))
	want := []int32{gia.Gia_ObjId(s.pGia, gia.Gia_ManCi(s.pGia, 0)), gia.Gia_ObjId(s.pGia, gia.Gia_ManCi(s.pGia, 1)), gia.Gia_ObjId(s.pGia, gia.Gia_ManCi(s.pGia, 2))}
	for repeat := 0; repeat < 2; repeat++ {
		Txs3_ManCollectCone(s, 0)
		if got := vec.Vec_IntArray(s.vCiObjs); !reflect.DeepEqual(got, want) || s.nPiLits != 1 {
			t.Fatalf("run %d: inputs %v, primary count %d, want %v and 1", repeat, got, s.nPiLits, want)
		}
		if vec.Vec_IntSize(s.vNodes) != 2 {
			t.Fatalf("run %d: expected two cone nodes", repeat)
		}
		for _, id := range append(append([]int32(nil), vec.Vec_IntArray(s.vCiObjs)...), vec.Vec_IntArray(s.vNodes)...) {
			if gia.Gia_ManObj(s.pGia, id).Value != 0 {
				t.Fatalf("object %d retained visit marker", id)
			}
		}
	}
}

func TestTxs3PriorityCountIsChecked(t *testing.T) {
	g := aig.NewGraph()
	defer func() {
		if recover() == nil {
			t.Error("mismatched priorities accepted")
		}
	}()
	Txs3_ManStart(nil, g, vec.Vec_IntAllocArrayCopy([]int32{1}, 1))
}

func TestTxs3MapsReorderedGiaInterfaceIDs(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	shared := g.AND(x, y)
	flop := g.NewInput() // input appears after a logic node in AIG ID order
	g.AddOutput(g.AND(shared, flop))
	g.AddOutput(flop)
	aig.Aig_ManSetRegNum(g, 1)
	s := Txs3_ManStart(nil, g, vec.Vec_IntAllocArrayCopy([]int32{1}, 1))
	defer Txs3_ManStop(s)
	giaFlop := gia.Gia_ObjId(s.pGia, gia.Gia_ManCi(s.pGia, 2))
	if got := s.giaToAig[giaFlop]; got != int32(flop.NodeID()) {
		t.Fatalf("flop mapped to %d, want %d", got, flop.NodeID())
	}
	output := gia.Gia_ObjId(s.pGia, gia.Gia_ManCo(s.pGia, 0))
	if got := s.giaToAig[output]; got != int32(g.Outputs()[0].NodeID()) {
		t.Fatalf("output mapped to %d", got)
	}
}
