package abci

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func fxCubes(rows ...[]int32) *vec.Vec_Wec_t {
	cubes := vec.Vec_WecAlloc(len(rows))
	for _, row := range rows {
		level := vec.Vec_WecPushLevel(cubes)
		for _, lit := range row {
			vec.Vec_IntPush(level, lit)
		}
	}
	return cubes
}

func fxCubeValues(cubes *vec.Vec_Wec_t) [][]int32 {
	rows := make([][]int32, vec.Vec_WecSize(cubes))
	for i := range rows {
		rows[i] = append([]int32(nil), vec.Vec_IntArray(vec.Vec_WecEntry(cubes, i))...)
	}
	return rows
}

func TestFxFastExtractSharesRepeatedTerm(t *testing.T) {
	cubes := fxCubes([]int32{3, 0, 2}, []int32{4, 0, 2}, []int32{5, 0, 2})
	if changed := Fx_FastExtract(cubes, 6, 5, 4, 0, 0, 0); changed != 1 {
		t.Fatalf("Fx_FastExtract returned %d", changed)
	}
	want := [][]int32{{3, 12}, {4, 12}, {5, 12}, {6, 0, 2}}
	if got := fxCubeValues(cubes); !reflect.DeepEqual(got, want) {
		t.Fatalf("extracted cubes = %v, want %v", got, want)
	}
}

func TestFxCubeFreeDivisorPreservesCommonAndPhaseEncoding(t *testing.T) {
	first := vec.Vec_IntAllocArrayCopy([]int32{7, 0, 2}, 3)
	second := vec.Vec_IntAllocArrayCopy([]int32{7, 0, 4}, 3)
	free := vec.Vec_IntAlloc(2)
	if common := Fx_ManDivFindCubeFree(first, second, free, new(int32)); common != 1 {
		t.Fatalf("common literal count = %d, want 1", common)
	}
	if got, want := vec.Vec_IntArray(free), []int32{4, 9}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cube-free encoded literals = %v, want %v", got, want)
	}
}

func TestFxMuxDivisorNormalizationCanonicalizesBothPhases(t *testing.T) {
	for _, tc := range []struct {
		input []int32
		phase int32
	}{{[]int32{0, 3, 4, 7}, 0}, {[]int32{0, 3, 5, 6}, 1}} {
		div := vec.Vec_IntAllocArrayCopy(tc.input, len(tc.input))
		if got := Fx_ManDivNormalize(div); got != tc.phase {
			t.Fatalf("normalization phase = %d, want %d for %v", got, tc.phase, tc.input)
		}
		if got, want := vec.Vec_IntArray(div), []int32{0, 3, 4, 7}; !reflect.DeepEqual(got, want) {
			t.Fatalf("normalized divisor = %v, want %v", got, want)
		}
	}
}

func TestAbcNtkFxPerformReconstructsSharedNetworkNode(t *testing.T) {
	n := abc.NewNetwork("fx-network")
	n.Type, n.Function = uint8(abc.ABC_NTK_LOGIC), uint8(abc.ABC_FUNC_SOP)
	a, b := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	var old []*abc.Object
	for i := 0; i < 3; i++ {
		node := n.NewObject(abc.ObjectNode)
		node.Data = "11 1\n"
		if err := n.AddFanin(node, a.Signal()); err != nil {
			t.Fatal(err)
		}
		if err := n.AddFanin(node, b.Signal()); err != nil {
			t.Fatal(err)
		}
		po := n.NewObject(abc.ObjectPO)
		if err := n.AddFanin(po, node.Signal()); err != nil {
			t.Fatal(err)
		}
		old = append(old, node)
	}
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)
	if changed := Abc_NtkFxPerform(n, 4, 4, 0, 0, 0); changed != 1 {
		t.Fatalf("Abc_NtkFxPerform returned %d", changed)
	}
	nodes := n.ObjectsOfKind(abc.ObjectNode)
	if len(nodes) != 4 {
		t.Fatalf("node count = %d, want 4", len(nodes))
	}
	shared := nodes[len(nodes)-1]
	if shared.Data != "11 1\n" || len(shared.Fanins()) != 2 || shared.Fanins()[0].Object != a || shared.Fanins()[1].Object != b {
		t.Fatalf("unexpected extracted node: data=%v fanins=%v", shared.Data, shared.Fanins())
	}
	for _, node := range old {
		fans := node.Fanins()
		if node.Data != "1 1\n" || len(fans) != 1 || fans[0].Object != shared {
			t.Fatalf("old node %d was not rewritten to shared divisor: data=%v fanins=%v", node.ID, node.Data, fans)
		}
	}
}
