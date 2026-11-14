package abc

import "testing"

func TestDupObjCreatesDefaultNameAndReplacesCopyEdge(t *testing.T) {
	source := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	input := Abc_NtkCreatePi(source)
	input.SetCopyEdge(source.One().Not())
	target := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	copy := Abc_NtkDupObj(target, input, 1)
	if copy.Name() == "" || copy.Name() != Abc_ObjName(input) {
		t.Fatalf("name lost: %q", copy.Name())
	}
	if got := input.ReadCopyEdge(); got != copy.Signal() {
		t.Fatalf("stale copy edge: %v", got)
	}
}

func TestNetworkMiterSATExhaustiveTwoInputFunctions(t *testing.T) {
	for truth := 0; truth < 16; truth++ {
		n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
		x, y := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
		output := n.One().Not()
		for assignment := 0; assignment < 4; assignment++ {
			if truth&(1<<assignment) == 0 {
				continue
			}
			a, b := x.Signal(), y.Signal()
			if assignment&1 == 0 {
				a = a.Not()
			}
			if assignment&2 == 0 {
				b = b.Not()
			}
			term := n.AndEdges(a, b)
			output = n.AndEdges(output.Not(), term.Not()).Not()
		}
		Abc_ObjAddFanin(Abc_NtkCreatePo(n), output)
		var conflicts, inspections int64 = -1, -1
		status := Abc_NtkMiterSat(n, 0, 0, 0, &conflicts, &inspections)
		want := int32(0)
		if truth == 0 {
			want = 1
		}
		if status != want || conflicts < 0 || inspections < 0 {
			t.Fatalf("truth=%04b status=%d stats=%d,%d", truth, status, conflicts, inspections)
		}
		if status == 0 {
			if len(n.Model) != 2 {
				t.Fatalf("truth=%04b model=%v", truth, n.Model)
			}
			assignment := n.Model[0] | n.Model[1]<<1
			if truth&(1<<assignment) == 0 {
				t.Fatalf("truth=%04b invalid witness %v", truth, n.Model)
			}
		}
		Abc_NtkDelete(n)
	}
}
