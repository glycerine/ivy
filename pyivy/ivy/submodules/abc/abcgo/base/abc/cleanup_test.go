package abc

import "testing"

func TestCleanupRemovesUnreachableLogic(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	input := n.NewObject(ObjectPI)
	input.SetName("in")
	live := Abc_NtkCreateNodeBuf(n, input)
	dangling := Abc_NtkCreateNodeInv(n, input)
	danglingChild := Abc_NtkCreateNodeBuf(n, dangling)
	output := n.NewObject(ObjectPO)
	output.SetName("out")
	Abc_ObjAddFanin(output, live)
	if got := Abc_NtkCleanup(n, 0); got != 2 {
		t.Fatalf("removed %d nodes, want 2", got)
	}
	if n.contains(dangling) || n.contains(danglingChild) || !n.contains(live) || output.fanins[0].Object != live {
		t.Fatal("cleanup removed the wrong nodes")
	}
	if err := n.Check(); err != nil {
		t.Fatalf("damaged network: %v", err)
	}
}
