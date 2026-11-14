package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func extractTestAIG() (*abc.Network, *abc.Object, *abc.Object, *abc.Object) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	return n, a, b, c
}

func TestShareSuperAndFlattensAndFindsContradiction(t *testing.T) {
	n, a, b, c := extractTestAIG()
	m := n.FunctionManager.(*abc.Abc_Aig_t)
	ab := abc.Abc_AigAnd(m, a, b)
	bc := abc.Abc_AigAnd(m, a, c)
	root := abc.Abc_AigAnd(m, ab, bc)
	var count int32
	leaves := Abc_NtkShareSuperAnd(root.Object, &count)
	if got := vec.Vec_WrdSize(leaves); got != 3 {
		t.Fatalf("AND super leaves = %d, want 3", got)
	}
	if count != 3 {
		t.Fatalf("expanded gate count = %d, want 3", count)
	}
	left := abc.Abc_AigAnd(m, a, b)
	right := abc.Abc_AigAnd(m, a, b.Signal().Not())
	contradiction := abc.Abc_AigAnd(m, left, right)
	count = 0
	leaves = Abc_NtkShareSuperAnd(contradiction.Object, &count)
	if vec.Vec_WrdSize(leaves) != 0 {
		t.Fatalf("contradictory AND super has %d leaves, want empty", vec.Vec_WrdSize(leaves))
	}
}

func TestShareSuperXorFlattensParityTree(t *testing.T) {
	n, a, b, c := extractTestAIG()
	m := n.FunctionManager.(*abc.Abc_Aig_t)
	ab := abc.Abc_AigXor(m, a, b)
	root := abc.Abc_AigXor(m, ab, c)
	if abc.Abc_NodeIsExorType(root.Object) == 0 {
		t.Fatal("constructed root is not recognized as XOR")
	}
	var polarity, count int32
	leaves := Abc_NtkShareSuperXor(root.Object, &polarity, &count)
	if got := vec.Vec_WrdSize(leaves); got != 3 {
		t.Fatalf("XOR super leaves = %d, want 3", got)
	}
	if (polarity != 0) != root.Complement || count != 2 {
		t.Fatalf("XOR polarity/count = %d/%d, root complement=%v", polarity, count, root.Complement)
	}
	left := abc.Abc_AigXor(m, a, b)
	right := abc.Abc_AigXor(m, a, c)
	duplicate := abc.Abc_AigXor(m, left, right)
	count = 0
	leaves = Abc_NtkShareSuperXor(duplicate.Object, &polarity, &count)
	if got := vec.Vec_WrdSize(leaves); got != 2 {
		t.Fatalf("XOR duplicate cancellation leaves = %d, want 2", got)
	}
}

func TestShareFindBestMatchUsesLargestCommonInputSet(t *testing.T) {
	buckets := vec.Vec_PtrAlloc(4)
	for i := 0; i < 4; i++ {
		vec.Vec_PtrPush(buckets, vec.Vec_PtrAlloc(2))
	}
	high := vec.Vec_PtrEntry(buckets, 3).(*Vec_Ptr_t)
	first, second := vec.Vec_IntAlloc(5), vec.Vec_IntAlloc(5)
	for _, x := range []int32{1, 0, 0, 1, 2} {
		vec.Vec_IntPush(first, x)
	}
	for _, x := range []int32{3, 0, 0, 1, 3} {
		vec.Vec_IntPush(second, x)
	}
	vec.Vec_PtrPush(high, first)
	vec.Vec_PtrPush(high, second)
	var got1, got2 *Vec_Int_t
	Abc_NtkShareFindBestMatch(buckets, &got1, &got2)
	if got1 == nil || got2 == nil || got1 == got2 {
		t.Fatal("expected a pair sharing two input columns")
	}
	if vec.Vec_PtrSize(high) != 0 {
		t.Fatalf("matched bucket still contains %d entries", vec.Vec_PtrSize(high))
	}
}

func TestShareXorExtractsCommonAndAndXorLogicWithoutChangingOutputs(t *testing.T) {
	for _, andMode := range []int32{1, 0} {
		n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
		pis := make([]*abc.Object, 4)
		for i := range pis {
			pis[i] = n.NewObject(abc.ObjectPI)
		}
		manager := n.FunctionManager.(*abc.Abc_Aig_t)
		var shared, left, right abc.Edge
		if andMode != 0 {
			shared = abc.Abc_AigAnd(manager, pis[0], pis[1])
			left = abc.Abc_AigAnd(manager, shared, pis[2])
			right = abc.Abc_AigAnd(manager, shared, pis[3])
		} else {
			shared = abc.Abc_AigXor(manager, pis[0], pis[1])
			left = abc.Abc_AigXor(manager, shared, pis[2])
			right = abc.Abc_AigXor(manager, shared, pis[3])
		}
		for _, driver := range []abc.Edge{left, right} {
			co := n.NewObject(abc.ObjectPO)
			abc.Abc_ObjAddFanin(co, driver)
		}
		before := make([][2]bool, 16)
		for mask := 0; mask < 16; mask++ {
			values := map[*abc.Object]bool{}
			for i, pi := range pis {
				values[pi] = mask&(1<<i) != 0
			}
			for output, co := range n.Outputs() {
				got, err := n.Evaluate(co.Signal(), values)
				if err != nil {
					t.Fatal(err)
				}
				before[mask][output] = got
			}
		}
		result := Abc_NtkShareXor(n, 3, andMode, 0)
		if result == nil {
			t.Fatal("sharing returned nil network")
		}
		for mask := 0; mask < 16; mask++ {
			values := map[*abc.Object]bool{}
			for i, pi := range result.Inputs() {
				values[pi] = mask&(1<<i) != 0
			}
			for output, co := range result.Outputs() {
				got, err := result.Evaluate(co.Signal(), values)
				if err != nil {
					t.Fatalf("mode %d inputs %04b output %d: %v", andMode, mask, output, err)
				}
				if got != before[mask][output] {
					t.Fatalf("mode %d inputs %04b output %d = %v, want %v", andMode, mask, output, got, before[mask][output])
				}
			}
		}
	}
}
