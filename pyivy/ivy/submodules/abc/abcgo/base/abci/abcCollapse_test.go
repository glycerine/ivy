package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestGlobalBddToLocalNetworkPreservesFamiliesAndOrder(t *testing.T) {
	n, inputs := multiTestAig(t)
	dd := Abc_NtkBuildGlobalBdds(n, 100000, 1, 0, 0, 0)
	if dd == nil {
		t.Fatal("global BDD construction failed")
	}
	defer Abc_NtkFreeGlobalBdds(n, 1)
	order := append([]int(nil), dd.InvPerm...)
	out := Abc_NtkFromGlobalBdds(n, 0)
	if out == nil || abc.Abc_NtkCheck(out) == 0 {
		t.Fatal("global BDD conversion failed")
	}
	defer abc.Abc_NtkDelete(out)
	if len(out.ObjectsOfKind(abc.ObjectPO)) != 2 {
		t.Fatalf("outputs=%d, want 2", len(out.ObjectsOfKind(abc.ObjectPO)))
	}
	for i, po := range out.ObjectsOfKind(abc.ObjectPO) {
		if i < 2 {
			node := po.Fanins()[0].Object
			for level, index := range order {
				if node.Fanins()[level].Object != abc.Abc_NtkCi(out, int32(index)) {
					t.Fatalf("output %d local fanin %d does not follow manager order %v", i, level, order)
				}
			}
		}
	}
	for mask := 0; mask < 1<<len(inputs); mask++ {
		values := make(map[*abc.Object]bool, len(inputs))
		for i, pi := range inputs {
			values[pi] = mask&(1<<i) != 0
		}
		for i, po := range n.ObjectsOfKind(abc.ObjectPO) {
			want := evalMultiNode(t, n, po.Fanins()[0].Object, values, map[*abc.Object]bool{}) != po.Fanins()[0].Complement
			got, err := out.Evaluate(out.ObjectsOfKind(abc.ObjectPO)[i].Signal(), mapLocalInputs(out, values, inputs))
			if err != nil || got != want {
				t.Fatalf("mask=%04b output=%d got=%v want=%v err=%v", mask, i, got, want, err)
			}
		}
	}
}

func mapLocalInputs(out *abc.Network, values map[*abc.Object]bool, source []*abc.Object) map[*abc.Object]bool {
	result := make(map[*abc.Object]bool, len(source))
	for i, pi := range source {
		result[abc.Abc_NtkCi(out, int32(i))] = values[pi]
	}
	return result
}

func TestCollapseReduceUpdatesEveryEquivalentOutputSupport(t *testing.T) {
	sop := vec.Vec_StrAlloc(0)
	for _, c := range []byte("1- 1\n1- 1\n\x00") {
		vec.Vec_StrPush(sop, c)
	}
	support := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(support, 0)
	vec.Vec_IntPush(support, 1)
	class := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(class, 0)
	vec.Vec_IntPush(class, 1)
	supports := vec.Vec_WecStart(2)
	for i := int32(0); i < 2; i++ {
		vec.Vec_IntPush(vec.Vec_WecEntry(supports, int(i)), 0)
		vec.Vec_IntPush(vec.Vec_WecEntry(supports, int(i)), 1)
	}
	defer vec.Vec_StrFree(sop)
	defer vec.Vec_IntFree(support)
	defer vec.Vec_IntFree(class)
	defer vec.Vec_WecFree(supports)
	if Abc_NtkCollapseReduce(sop, support, class, supports) != 1 {
		t.Fatal("expected support reduction")
	}
	if got := string(vec.Vec_StrArray(sop)); got != "1 1\n1 1\n\x00" {
		t.Fatalf("reduced SOP=%q", got)
	}
	for i := 0; i < 2; i++ {
		if vec.Vec_IntSize(vec.Vec_WecEntry(supports, i)) != 1 || vec.Vec_IntEntry(vec.Vec_WecEntry(supports, i), 0) != 0 {
			t.Fatalf("support %d was not reduced", i)
		}
	}
}

func TestSatCollapsePreservesOutputFamiliesWithSharedAndLocalCnf(t *testing.T) {
	n, inputs := multiTestAig(t)
	for _, shared := range []int32{0, 1} {
		out := Abc_NtkCollapseSat(n, 0, 0, 0, 0, 0, shared, 0)
		if out == nil || abc.Abc_NtkCheck(out) == 0 {
			t.Fatalf("SAT collapse failed (shared CNF=%d)", shared)
		}
		defer abc.Abc_NtkDelete(out)
		for mask := 0; mask < 1<<len(inputs); mask++ {
			values := make(map[*abc.Object]bool, len(inputs))
			local := make(map[*abc.Object]bool, len(inputs))
			for i, pi := range inputs {
				value := mask&(1<<i) != 0
				values[pi] = value
				local[abc.Abc_NtkCi(out, int32(i))] = value
			}
			for i, po := range n.ObjectsOfKind(abc.ObjectPO) {
				want := evalMultiNode(t, n, po.Fanins()[0].Object, values, map[*abc.Object]bool{}) != po.Fanins()[0].Complement
				got, err := out.Evaluate(abc.Abc_NtkPo(out, int32(i)).Signal(), local)
				if err != nil || got != want {
					t.Fatalf("shared=%d mask=%04b output=%d got=%v want=%v err=%v", shared, mask, i, got, want, err)
				}
			}
		}
	}
}
