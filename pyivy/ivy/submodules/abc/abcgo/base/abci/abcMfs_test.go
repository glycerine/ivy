package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/opt/sfm"
)

func TestAbcNtkFillTruthStorePatterns(t *testing.T) {
	var store [16][1 << 10]word
	Abc_NtkFillTruthStore(store[:])
	patterns := spStatic_Abc_NtkFillTruthStore_Truth6
	for variable, pattern := range patterns {
		for wordIndex, got := range store[variable] {
			if got != pattern {
				t.Fatalf("truth store[%d][%d] = %#x, want %#x", variable, wordIndex, got, pattern)
			}
		}
	}
	for variable := 6; variable < len(store); variable++ {
		for index, got := range store[variable] {
			want := word(0)
			if index&(1<<(variable-6)) != 0 {
				want = ^word(0)
			}
			if got != want {
				t.Fatalf("truth store[%d][%d] = %#x, want %#x", variable, index, got, want)
			}
		}
	}
	// The C helper only initializes a zeroed store; a second call must leave a
	// caller-modified nonzero first entry intact.
	store[0][0] = 17
	Abc_NtkFillTruthStore(store[:])
	if store[0][0] != 17 {
		t.Fatalf("nonzero truth-store marker overwritten: %#x", store[0][0])
	}
}

func TestAbcNtkExtractMfsPreservesSevenInputSop(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	inputs := make([]*abc.Abc_Obj_t, 7)
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(network)
	}
	node := abc.Abc_NtkCreateNode(network)
	flags := make([]int32, len(inputs))
	for _, input := range inputs {
		abc.Abc_ObjAddFanin(node, input)
	}
	node.Data = abc.Abc_SopCreateAnd(nil, int32(len(inputs)), flags)
	output := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(output, node)

	model := Abc_NtkExtractMfs(network, 0)
	defer sfm.Sfm_NtkFree(model)
	for mask := 0; mask < 1<<len(inputs); mask++ {
		values := make([]bool, len(inputs))
		for i := range values {
			values[i] = mask&(1<<i) != 0
		}
		got, err := model.Evaluate(values)
		if err != nil {
			t.Fatalf("Evaluate(%07b): %v", mask, err)
		}
		want := mask == (1<<len(inputs))-1
		if len(got) != 1 || got[0] != want {
			t.Fatalf("MFS model(%07b) = %v, want [%v]", mask, got, want)
		}
	}
}

func TestAbcNtkAssignIDs2UsesNetworkObjectOrder(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	first := abc.Abc_NtkCreateNode(network)
	abc.Abc_ObjAddFanin(first, a)
	abc.Abc_ObjAddFanin(first, b)
	first.Data = abc.Abc_SopCreateAnd(nil, 2, []int32{0, 0})
	second := abc.Abc_NtkCreateNode(network)
	abc.Abc_ObjAddFanin(second, first)
	abc.Abc_ObjAddFanin(second, a)
	second.Data = abc.Abc_SopCreateAnd(nil, 2, []int32{0, 0})
	po := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(po, second)

	nodes := Abc_NtkAssignIDs2(network)
	defer Vec_PtrFree(nodes)
	if a.Temp != 0 || b.Temp != 1 || first.Temp != 2 || second.Temp != 3 || po.Temp != 4 {
		t.Fatalf("assigned IDs: CIs=(%d,%d), nodes=(%d,%d), CO=%d", a.Temp, b.Temp, first.Temp, second.Temp, po.Temp)
	}
}
